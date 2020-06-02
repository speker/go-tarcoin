package cache

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/content/local"
	"github.com/containerd/containerd/diff/apply"
	"github.com/containerd/containerd/leases"
	ctdmetadata "github.com/containerd/containerd/metadata"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/snapshots"
	"github.com/containerd/containerd/snapshots/native"
	"github.com/moby/buildkit/cache/metadata"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/snapshot"
	containerdsnapshot "github.com/moby/buildkit/snapshot/containerd"
	"github.com/moby/buildkit/util/leaseutil"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

type cmOpt struct {
	snapshotterName string
	snapshotter     snapshots.Snapshotter
	tmpdir          string
}

type cmOut struct {
	manager Manager
	lm      leases.Manager
	cs      content.Store
}

func newCacheManager(ctx context.Context, opt cmOpt) (co *cmOut, cleanup func() error, err error) {
	ns, ok := namespaces.Namespace(ctx)
	if !ok {
		return nil, nil, errors.Errorf("namespace required for test")
	}

	if opt.snapshotterName == "" {
		opt.snapshotterName = "native"
	}

	tmpdir, err := ioutil.TempDir("", "cachemanager")
	if err != nil {
		return nil, nil, err
	}

	defers := make([]func() error, 0)
	cleanup = func() error {
		var err error
		for i := range defers {
			if err1 := defers[len(defers)-1-i](); err1 != nil && err == nil {
				err = err1
			}
		}
		return err
	}
	defer func() {
		if err != nil {
			cleanup()
		}
	}()
	if opt.tmpdir == "" {
		defers = append(defers, func() error {
			return os.RemoveAll(tmpdir)
		})
	} else {
		os.RemoveAll(tmpdir)
		tmpdir = opt.tmpdir
	}

	if opt.snapshotter == nil {
		snapshotter, err := native.NewSnapshotter(filepath.Join(tmpdir, "snapshots"))
		if err != nil {
			return nil, nil, err
		}
		opt.snapshotter = snapshotter
	}

	md, err := metadata.NewStore(filepath.Join(tmpdir, "metadata.db"))
	if err != nil {
		return nil, nil, err
	}

	store, err := local.NewStore(tmpdir)
	if err != nil {
		return nil, nil, err
	}

	db, err := bolt.Open(filepath.Join(tmpdir, "containerdmeta.db"), 0644, nil)
	if err != nil {
		return nil, nil, err
	}
	defers = append(defers, func() error {
		return db.Close()
	})

	mdb := ctdmetadata.NewDB(db, store, map[string]snapshots.Snapshotter{
		opt.snapshotterName: opt.snapshotter,
	})
	if err := mdb.Init(context.TODO()); err != nil {
		return nil, nil, err
	}

	lm := ctdmetadata.NewLeaseManager(mdb)

	cm, err := NewManager(ManagerOpt{
		Snapshotter:    snapshot.FromContainerdSnapshotter(opt.snapshotterName, containerdsnapshot.NSSnapshotter(ns, mdb.Snapshotter(opt.snapshotterName)), nil),
		MetadataStore:  md,
		ContentStore:   mdb.ContentStore(),
		LeaseManager:   leaseutil.WithNamespace(lm, ns),
		GarbageCollect: mdb.GarbageCollect,
		Applier:        apply.NewFileSystemApplier(mdb.ContentStore()),
	})
	if err != nil {
		return nil, nil, err
	}
	return &cmOut{
		manager: cm,
		lm:      lm,
		cs:      mdb.ContentStore(),
	}, cleanup, nil
}

func TestManager(t *testing.T) {
	t.Parallel()

	ctx := namespaces.WithNamespace(context.Background(), "buildkit-test")

	tmpdir, err := ioutil.TempDir("", "cachemanager")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	snapshotter, err := native.NewSnapshotter(filepath.Join(tmpdir, "snapshots"))
	require.NoError(t, err)

	co, cleanup, err := newCacheManager(ctx, cmOpt{
		snapshotter:     snapshotter,
		snapshotterName: "native",
	})
	require.NoError(t, err)

	defer cleanup()
	cm := co.manager

	_, err = cm.Get(ctx, "foobar")
	require.Error(t, err)

	checkDiskUsage(ctx, t, cm, 0, 0)

	active, err := cm.New(ctx, nil, CachePolicyRetain)
	require.NoError(t, err)

	m, err := active.Mount(ctx, false)
	require.NoError(t, err)

	lm := snapshot.LocalMounter(m)
	target, err := lm.Mount()
	require.NoError(t, err)

	fi, err := os.Stat(target)
	require.NoError(t, err)
	require.Equal(t, fi.IsDir(), true)

	err = lm.Unmount()
	require.NoError(t, err)

	_, err = cm.GetMutable(ctx, active.ID())
	require.Error(t, err)
	require.Equal(t, ErrLocked, errors.Cause(err))

	checkDiskUsage(ctx, t, cm, 1, 0)

	snap, err := active.Commit(ctx)
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 1, 0)

	_, err = cm.GetMutable(ctx, active.ID())
	require.Error(t, err)
	require.Equal(t, ErrLocked, errors.Cause(err))

	err = snap.Release(ctx)
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 0, 1)

	active, err = cm.GetMutable(ctx, active.ID())
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 1, 0)

	snap, err = active.Commit(ctx)
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 1, 0)

	err = snap.Finalize(ctx, true)
	require.NoError(t, err)

	err = snap.Release(ctx)
	require.NoError(t, err)

	_, err = cm.GetMutable(ctx, active.ID())
	require.Error(t, err)
	require.Equal(t, errNotFound, errors.Cause(err))

	_, err = cm.GetMutable(ctx, snap.ID())
	require.Error(t, err)
	require.Equal(t, errInvalid, errors.Cause(err))

	snap, err = cm.Get(ctx, snap.ID())
	require.NoError(t, err)

	snap2, err := cm.Get(ctx, snap.ID())
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 1, 0)

	err = snap.Release(ctx)
	require.NoError(t, err)

	active2, err := cm.New(ctx, snap2, CachePolicyRetain)
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 2, 0)

	snap3, err := active2.Commit(ctx)
	require.NoError(t, err)

	err = snap2.Release(ctx)
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 2, 0)

	err = snap3.Release(ctx)
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 0, 2)

	buf := pruneResultBuffer()
	err = cm.Prune(ctx, buf.C, client.PruneInfo{})
	buf.close()
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 0, 0)

	require.Equal(t, len(buf.all), 2)

	err = cm.Close()
	require.NoError(t, err)

	dirs, err := ioutil.ReadDir(filepath.Join(tmpdir, "snapshots/snapshots"))
	require.NoError(t, err)
	require.Equal(t, 0, len(dirs))
}

func TestSnapshotExtract(t *testing.T) {
	t.Parallel()
	ctx := namespaces.WithNamespace(context.Background(), "buildkit-test")

	tmpdir, err := ioutil.TempDir("", "cachemanager")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	snapshotter, err := native.NewSnapshotter(filepath.Join(tmpdir, "snapshots"))
	require.NoError(t, err)

	co, cleanup, err := newCacheManager(ctx, cmOpt{
		snapshotter:     snapshotter,
		snapshotterName: "native",
	})
	require.NoError(t, err)

	defer cleanup()

	cm := co.manager

	b, desc, err := mapToBlob(map[string]string{"foo": "bar"})
	require.NoError(t, err)

	err = content.WriteBlob(ctx, co.cs, "ref1", bytes.NewBuffer(b), desc)
	require.NoError(t, err)

	snap, err := cm.GetByBlob(ctx, desc, nil)
	require.NoError(t, err)

	require.Equal(t, false, snap.Info().Extracted)

	b2, desc2, err := mapToBlob(map[string]string{"foo": "bar123"})
	require.NoError(t, err)

	err = content.WriteBlob(ctx, co.cs, "ref1", bytes.NewBuffer(b2), desc2)
	require.NoError(t, err)

	snap2, err := cm.GetByBlob(ctx, desc2, snap)
	require.NoError(t, err)

	size, err := snap2.Size(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(len(b2)), size)

	require.Equal(t, false, snap2.Info().Extracted)

	dirs, err := ioutil.ReadDir(filepath.Join(tmpdir, "snapshots/snapshots"))
	require.NoError(t, err)
	require.Equal(t, 0, len(dirs))

	checkNumBlobs(ctx, t, co.cs, 2)

	err = snap2.Extract(ctx)
	require.NoError(t, err)

	require.Equal(t, true, snap.Info().Extracted)
	require.Equal(t, true, snap2.Info().Extracted)

	dirs, err = ioutil.ReadDir(filepath.Join(tmpdir, "snapshots/snapshots"))
	require.NoError(t, err)
	require.Equal(t, 2, len(dirs))

	buf := pruneResultBuffer()
	err = cm.Prune(ctx, buf.C, client.PruneInfo{})
	buf.close()
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 2, 0)

	require.Equal(t, len(buf.all), 0)

	dirs, err = ioutil.ReadDir(filepath.Join(tmpdir, "snapshots/snapshots"))
	require.NoError(t, err)
	require.Equal(t, 2, len(dirs))

	checkNumBlobs(ctx, t, co.cs, 2)

	id := snap.ID()

	err = snap.Release(context.TODO())
	require.NoError(t, err)

	buf = pruneResultBuffer()
	err = cm.Prune(ctx, buf.C, client.PruneInfo{})
	buf.close()
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 2, 0)

	dirs, err = ioutil.ReadDir(filepath.Join(tmpdir, "snapshots/snapshots"))
	require.NoError(t, err)
	require.Equal(t, 2, len(dirs))

	snap, err = cm.Get(ctx, id)
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 2, 0)

	err = snap2.Release(context.TODO())
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 1, 1)

	buf = pruneResultBuffer()
	err = cm.Prune(ctx, buf.C, client.PruneInfo{})
	buf.close()
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 1, 0)

	require.Equal(t, len(buf.all), 1)

	dirs, err = ioutil.ReadDir(filepath.Join(tmpdir, "snapshots/snapshots"))
	require.NoError(t, err)
	require.Equal(t, 1, len(dirs))

	checkNumBlobs(ctx, t, co.cs, 1)

	err = snap.Release(context.TODO())
	require.NoError(t, err)

	buf = pruneResultBuffer()
	err = cm.Prune(ctx, buf.C, client.PruneInfo{})
	buf.close()
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 0, 0)

	dirs, err = ioutil.ReadDir(filepath.Join(tmpdir, "snapshots/snapshots"))
	require.NoError(t, err)
	require.Equal(t, 0, len(dirs))

	checkNumBlobs(ctx, t, co.cs, 0)
}

func TestExtractOnMutable(t *testing.T) {
	t.Parallel()
	ctx := namespaces.WithNamespace(context.Background(), "buildkit-test")

	tmpdir, err := ioutil.TempDir("", "cachemanager")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	snapshotter, err := native.NewSnapshotter(filepath.Join(tmpdir, "snapshots"))
	require.NoError(t, err)

	co, cleanup, err := newCacheManager(ctx, cmOpt{
		snapshotter:     snapshotter,
		snapshotterName: "native",
	})
	require.NoError(t, err)

	defer cleanup()

	cm := co.manager

	active, err := cm.New(ctx, nil)
	require.NoError(t, err)

	snap, err := active.Commit(ctx)
	require.NoError(t, err)

	b, desc, err := mapToBlob(map[string]string{"foo": "bar"})
	require.NoError(t, err)

	err = content.WriteBlob(ctx, co.cs, "ref1", bytes.NewBuffer(b), desc)
	require.NoError(t, err)

	b2, desc2, err := mapToBlob(map[string]string{"foo2": "1"})
	require.NoError(t, err)

	err = content.WriteBlob(ctx, co.cs, "ref2", bytes.NewBuffer(b2), desc2)
	require.NoError(t, err)

	snap2, err := cm.GetByBlob(ctx, desc2, snap)
	require.Error(t, err)

	err = snap.SetBlob(ctx, desc)
	require.NoError(t, err)

	snap2, err = cm.GetByBlob(ctx, desc2, snap)
	require.NoError(t, err)

	err = snap.Release(context.TODO())
	require.NoError(t, err)

	require.Equal(t, false, snap2.Info().Extracted)

	size, err := snap2.Size(ctx)
	require.NoError(t, err)
	require.Equal(t, int64(len(b2)), size)

	dirs, err := ioutil.ReadDir(filepath.Join(tmpdir, "snapshots/snapshots"))
	require.NoError(t, err)
	require.Equal(t, 1, len(dirs))

	checkNumBlobs(ctx, t, co.cs, 2)

	err = snap2.Extract(ctx)
	require.NoError(t, err)

	require.Equal(t, true, snap.Info().Extracted)
	require.Equal(t, true, snap2.Info().Extracted)

	buf := pruneResultBuffer()
	err = cm.Prune(ctx, buf.C, client.PruneInfo{})
	buf.close()
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 2, 0)

	require.Equal(t, len(buf.all), 0)

	dirs, err = ioutil.ReadDir(filepath.Join(tmpdir, "snapshots/snapshots"))
	require.NoError(t, err)
	require.Equal(t, 2, len(dirs))

	err = snap2.Release(context.TODO())
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 0, 2)

	buf = pruneResultBuffer()
	err = cm.Prune(ctx, buf.C, client.PruneInfo{})
	buf.close()
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 0, 0)

	require.Equal(t, len(buf.all), 2)

	dirs, err = ioutil.ReadDir(filepath.Join(tmpdir, "snapshots/snapshots"))
	require.NoError(t, err)
	require.Equal(t, 0, len(dirs))

	checkNumBlobs(ctx, t, co.cs, 0)
}

func TestSetBlob(t *testing.T) {
	t.Parallel()
	ctx := namespaces.WithNamespace(context.Background(), "buildkit-test")

	tmpdir, err := ioutil.TempDir("", "cachemanager")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	snapshotter, err := native.NewSnapshotter(filepath.Join(tmpdir, "snapshots"))
	require.NoError(t, err)

	co, cleanup, err := newCacheManager(ctx, cmOpt{
		snapshotter:     snapshotter,
		snapshotterName: "native",
	})
	require.NoError(t, err)

	defer cleanup()

	cm := co.manager

	active, err := cm.New(ctx, nil)
	require.NoError(t, err)

	snap, err := active.Commit(ctx)
	require.NoError(t, err)

	info := snap.Info()
	require.Equal(t, "", string(info.DiffID))
	require.Equal(t, "", string(info.Blob))
	require.Equal(t, "", string(info.ChainID))
	require.Equal(t, "", string(info.BlobChainID))
	require.Equal(t, info.Extracted, true)

	ctx, clean, err := leaseutil.WithLease(ctx, co.lm)
	require.NoError(t, err)
	defer clean(context.TODO())

	b, desc, err := mapToBlob(map[string]string{"foo": "bar"})
	require.NoError(t, err)

	err = content.WriteBlob(ctx, co.cs, "ref1", bytes.NewBuffer(b), desc)
	require.NoError(t, err)

	err = snap.SetBlob(ctx, ocispec.Descriptor{
		Digest: digest.FromBytes([]byte("foobar")),
		Annotations: map[string]string{
			"containerd.io/uncompressed": digest.FromBytes([]byte("foobar2")).String(),
		},
	})
	require.Error(t, err)

	err = snap.SetBlob(ctx, desc)
	require.NoError(t, err)

	info = snap.Info()
	require.Equal(t, desc.Annotations["containerd.io/uncompressed"], string(info.DiffID))
	require.Equal(t, desc.Digest, info.Blob)
	require.Equal(t, desc.MediaType, info.MediaType)
	require.Equal(t, info.DiffID, info.ChainID)
	require.Equal(t, digest.FromBytes([]byte(desc.Digest+" "+info.DiffID)), info.BlobChainID)
	require.Equal(t, snap.ID(), info.SnapshotID)
	require.Equal(t, info.Extracted, true)

	active, err = cm.New(ctx, snap)
	require.NoError(t, err)

	snap2, err := active.Commit(ctx)
	require.NoError(t, err)

	b2, desc2, err := mapToBlob(map[string]string{"foo2": "bar2"})
	require.NoError(t, err)

	err = content.WriteBlob(ctx, co.cs, "ref2", bytes.NewBuffer(b2), desc2)
	require.NoError(t, err)

	err = snap2.SetBlob(ctx, desc2)
	require.NoError(t, err)

	info2 := snap2.Info()
	require.Equal(t, desc2.Annotations["containerd.io/uncompressed"], string(info2.DiffID))
	require.Equal(t, desc2.Digest, info2.Blob)
	require.Equal(t, desc2.MediaType, info2.MediaType)
	require.Equal(t, digest.FromBytes([]byte(info.ChainID+" "+info2.DiffID)), info2.ChainID)
	require.Equal(t, digest.FromBytes([]byte(info.BlobChainID+" "+digest.FromBytes([]byte(desc2.Digest+" "+info2.DiffID)))), info2.BlobChainID)
	require.Equal(t, snap2.ID(), info2.SnapshotID)
	require.Equal(t, info2.Extracted, true)

	b3, desc3, err := mapToBlob(map[string]string{"foo3": "bar3"})
	require.NoError(t, err)

	err = content.WriteBlob(ctx, co.cs, "ref3", bytes.NewBuffer(b3), desc3)
	require.NoError(t, err)

	snap3, err := cm.GetByBlob(ctx, desc3, snap)
	require.NoError(t, err)

	info3 := snap3.Info()
	require.Equal(t, desc3.Annotations["containerd.io/uncompressed"], string(info3.DiffID))
	require.Equal(t, desc3.Digest, info3.Blob)
	require.Equal(t, desc3.MediaType, info3.MediaType)
	require.Equal(t, digest.FromBytes([]byte(info.ChainID+" "+info3.DiffID)), info3.ChainID)
	require.Equal(t, digest.FromBytes([]byte(info.BlobChainID+" "+digest.FromBytes([]byte(desc3.Digest+" "+info3.DiffID)))), info3.BlobChainID)
	require.Equal(t, string(info3.ChainID), info3.SnapshotID)
	require.Equal(t, info3.Extracted, false)

	// snap4 is same as snap2
	snap4, err := cm.GetByBlob(ctx, desc2, snap)
	require.NoError(t, err)

	require.Equal(t, snap2.ID(), snap4.ID())

	// snap5 is same different blob but same diffID as snap2
	b5, desc5, err := mapToBlob(map[string]string{"foo5": "bar5"})
	require.NoError(t, err)

	desc5.Annotations["containerd.io/uncompressed"] = info2.DiffID.String()

	err = content.WriteBlob(ctx, co.cs, "ref5", bytes.NewBuffer(b5), desc5)
	require.NoError(t, err)

	snap5, err := cm.GetByBlob(ctx, desc5, snap)
	require.NoError(t, err)

	require.NotEqual(t, snap2.ID(), snap5.ID())
	require.Equal(t, snap2.Info().SnapshotID, snap5.Info().SnapshotID)
	require.Equal(t, info2.DiffID, snap5.Info().DiffID)
	require.Equal(t, desc5.Digest, snap5.Info().Blob)

	require.Equal(t, snap2.Info().ChainID, snap5.Info().ChainID)
	require.NotEqual(t, snap2.Info().BlobChainID, snap5.Info().BlobChainID)
	require.Equal(t, digest.FromBytes([]byte(info.BlobChainID+" "+digest.FromBytes([]byte(desc5.Digest+" "+info2.DiffID)))), snap5.Info().BlobChainID)

	// snap6 is a child of snap3
	b6, desc6, err := mapToBlob(map[string]string{"foo6": "bar6"})
	require.NoError(t, err)

	err = content.WriteBlob(ctx, co.cs, "ref6", bytes.NewBuffer(b6), desc6)
	require.NoError(t, err)

	snap6, err := cm.GetByBlob(ctx, desc6, snap3)
	require.NoError(t, err)

	info6 := snap6.Info()
	require.Equal(t, desc6.Annotations["containerd.io/uncompressed"], string(info6.DiffID))
	require.Equal(t, desc6.Digest, info6.Blob)
	require.Equal(t, digest.FromBytes([]byte(snap3.Info().ChainID+" "+info6.DiffID)), info6.ChainID)
	require.Equal(t, digest.FromBytes([]byte(info3.BlobChainID+" "+digest.FromBytes([]byte(info6.Blob+" "+info6.DiffID)))), info6.BlobChainID)
	require.Equal(t, string(info6.ChainID), info6.SnapshotID)
	require.Equal(t, info6.Extracted, false)

	_, err = cm.GetByBlob(ctx, ocispec.Descriptor{
		Digest: digest.FromBytes([]byte("notexist")),
		Annotations: map[string]string{
			"containerd.io/uncompressed": digest.FromBytes([]byte("notexist")).String(),
		},
	}, snap3)
	require.Error(t, err)

	clean(context.TODO())

	//snap.SetBlob()
}

func TestPrune(t *testing.T) {
	t.Parallel()
	ctx := namespaces.WithNamespace(context.Background(), "buildkit-test")

	tmpdir, err := ioutil.TempDir("", "cachemanager")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	snapshotter, err := native.NewSnapshotter(filepath.Join(tmpdir, "snapshots"))
	require.NoError(t, err)

	co, cleanup, err := newCacheManager(ctx, cmOpt{
		snapshotter:     snapshotter,
		snapshotterName: "native",
	})
	require.NoError(t, err)

	defer cleanup()
	cm := co.manager

	active, err := cm.New(ctx, nil)
	require.NoError(t, err)

	snap, err := active.Commit(ctx)
	require.NoError(t, err)

	active, err = cm.New(ctx, snap, CachePolicyRetain)
	require.NoError(t, err)

	snap2, err := active.Commit(ctx)
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 2, 0)

	dirs, err := ioutil.ReadDir(filepath.Join(tmpdir, "snapshots/snapshots"))
	require.NoError(t, err)
	require.Equal(t, 2, len(dirs))

	// prune with keeping refs does nothing
	buf := pruneResultBuffer()
	err = cm.Prune(ctx, buf.C, client.PruneInfo{})
	buf.close()
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 2, 0)
	require.Equal(t, len(buf.all), 0)

	dirs, err = ioutil.ReadDir(filepath.Join(tmpdir, "snapshots/snapshots"))
	require.NoError(t, err)
	require.Equal(t, 2, len(dirs))

	err = snap2.Release(ctx)
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 1, 1)

	// prune with keeping single refs deletes one
	buf = pruneResultBuffer()
	err = cm.Prune(ctx, buf.C, client.PruneInfo{})
	buf.close()
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 1, 0)
	require.Equal(t, len(buf.all), 1)

	dirs, err = ioutil.ReadDir(filepath.Join(tmpdir, "snapshots/snapshots"))
	require.NoError(t, err)
	require.Equal(t, 1, len(dirs))

	err = snap.Release(ctx)
	require.NoError(t, err)

	active, err = cm.New(ctx, snap, CachePolicyRetain)
	require.NoError(t, err)

	snap2, err = active.Commit(ctx)
	require.NoError(t, err)

	err = snap.Release(ctx)
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 2, 0)

	// prune with parent released does nothing
	buf = pruneResultBuffer()
	err = cm.Prune(ctx, buf.C, client.PruneInfo{})
	buf.close()
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 2, 0)
	require.Equal(t, len(buf.all), 0)

	// releasing last reference
	err = snap2.Release(ctx)
	require.NoError(t, err)
	checkDiskUsage(ctx, t, cm, 0, 2)

	buf = pruneResultBuffer()
	err = cm.Prune(ctx, buf.C, client.PruneInfo{})
	buf.close()
	require.NoError(t, err)

	checkDiskUsage(ctx, t, cm, 0, 0)
	require.Equal(t, len(buf.all), 2)

	dirs, err = ioutil.ReadDir(filepath.Join(tmpdir, "snapshots/snapshots"))
	require.NoError(t, err)
	require.Equal(t, 0, len(dirs))
}

func TestLazyCommit(t *testing.T) {
	t.Parallel()

	ctx := namespaces.WithNamespace(context.Background(), "buildkit-test")

	tmpdir, err := ioutil.TempDir("", "cachemanager")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	snapshotter, err := native.NewSnapshotter(filepath.Join(tmpdir, "snapshots"))
	require.NoError(t, err)

	co, cleanup, err := newCacheManager(ctx, cmOpt{
		tmpdir:          tmpdir,
		snapshotter:     snapshotter,
		snapshotterName: "native",
	})
	require.NoError(t, err)
	cm := co.manager

	active, err := cm.New(ctx, nil, CachePolicyRetain)
	require.NoError(t, err)

	// after commit mutable is locked
	snap, err := active.Commit(ctx)
	require.NoError(t, err)

	_, err = cm.GetMutable(ctx, active.ID())
	require.Error(t, err)
	require.Equal(t, ErrLocked, errors.Cause(err))

	// immutable refs still work
	snap2, err := cm.Get(ctx, snap.ID())
	require.NoError(t, err)
	require.Equal(t, snap.ID(), snap2.ID())

	err = snap.Release(ctx)
	require.NoError(t, err)

	err = snap2.Release(ctx)
	require.NoError(t, err)

	// immutable work after final release as well
	snap, err = cm.Get(ctx, snap.ID())
	require.NoError(t, err)
	require.Equal(t, snap.ID(), snap2.ID())

	// active can't be get while immutable is held
	_, err = cm.GetMutable(ctx, active.ID())
	require.Error(t, err)
	require.Equal(t, ErrLocked, errors.Cause(err))

	err = snap.Release(ctx)
	require.NoError(t, err)

	// after release mutable becomes available again
	active2, err := cm.GetMutable(ctx, active.ID())
	require.NoError(t, err)
	require.Equal(t, active2.ID(), active.ID())

	// because ref was took mutable old immutable are cleared
	_, err = cm.Get(ctx, snap.ID())
	require.Error(t, err)
	require.Equal(t, errNotFound, errors.Cause(err))

	snap, err = active2.Commit(ctx)
	require.NoError(t, err)

	// this time finalize commit
	err = snap.Finalize(ctx, true)
	require.NoError(t, err)

	err = snap.Release(ctx)
	require.NoError(t, err)

	// mutable is gone after finalize
	_, err = cm.GetMutable(ctx, active2.ID())
	require.Error(t, err)
	require.Equal(t, errNotFound, errors.Cause(err))

	// immutable still works
	snap2, err = cm.Get(ctx, snap.ID())
	require.NoError(t, err)
	require.Equal(t, snap.ID(), snap2.ID())

	err = snap2.Release(ctx)
	require.NoError(t, err)

	// test restarting after commit
	active, err = cm.New(ctx, nil, CachePolicyRetain)
	require.NoError(t, err)

	// after commit mutable is locked
	snap, err = active.Commit(ctx)
	require.NoError(t, err)

	err = cm.Close()
	require.NoError(t, err)

	cleanup()

	// we can't close snapshotter and open it twice (especially, its internal bbolt store)
	co, cleanup, err = newCacheManager(ctx, cmOpt{
		tmpdir:          tmpdir,
		snapshotter:     snapshotter,
		snapshotterName: "native",
	})
	require.NoError(t, err)
	cm = co.manager

	snap2, err = cm.Get(ctx, snap.ID())
	require.NoError(t, err)

	err = snap2.Release(ctx)
	require.NoError(t, err)

	active, err = cm.GetMutable(ctx, active.ID())
	require.NoError(t, err)

	_, err = cm.Get(ctx, snap.ID())
	require.Error(t, err)
	require.Equal(t, errNotFound, errors.Cause(err))

	snap, err = active.Commit(ctx)
	require.NoError(t, err)

	err = cm.Close()
	require.NoError(t, err)

	cleanup()

	co, cleanup, err = newCacheManager(ctx, cmOpt{
		tmpdir:          tmpdir,
		snapshotter:     snapshotter,
		snapshotterName: "native",
	})
	require.NoError(t, err)
	defer cleanup()
	cm = co.manager

	snap2, err = cm.Get(ctx, snap.ID())
	require.NoError(t, err)

	err = snap2.Finalize(ctx, true)
	require.NoError(t, err)

	err = snap2.Release(ctx)
	require.NoError(t, err)

	active, err = cm.GetMutable(ctx, active.ID())
	require.Error(t, err)
	require.Equal(t, errNotFound, errors.Cause(err))
}

func checkDiskUsage(ctx context.Context, t *testing.T, cm Manager, inuse, unused int) {
	du, err := cm.DiskUsage(ctx, client.DiskUsageInfo{})
	require.NoError(t, err)
	var inuseActual, unusedActual int
	for _, r := range du {
		if r.InUse {
			inuseActual++
		} else {
			unusedActual++
		}
	}
	require.Equal(t, inuse, inuseActual)
	require.Equal(t, unused, unusedActual)
}

func checkNumBlobs(ctx context.Context, t *testing.T, cs content.Store, expected int) {
	c := 0
	err := cs.Walk(ctx, func(_ content.Info) error {
		c++
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, expected, c)
}

func pruneResultBuffer() *buf {
	b := &buf{C: make(chan client.UsageInfo), closed: make(chan struct{})}
	go func() {
		for c := range b.C {
			b.all = append(b.all, c)
		}
		close(b.closed)
	}()
	return b
}

type buf struct {
	C      chan client.UsageInfo
	closed chan struct{}
	all    []client.UsageInfo
}

func (b *buf) close() {
	close(b.C)
	<-b.closed
}

func mapToBlob(m map[string]string) ([]byte, ocispec.Descriptor, error) {
	buf := bytes.NewBuffer(nil)
	gz := gzip.NewWriter(buf)
	sha := digest.SHA256.Digester()
	tw := tar.NewWriter(io.MultiWriter(sha.Hash(), gz))

	for k, v := range m {
		if err := tw.WriteHeader(&tar.Header{
			Name: k,
			Size: int64(len(v)),
		}); err != nil {
			return nil, ocispec.Descriptor{}, err
		}
		if _, err := tw.Write([]byte(v)); err != nil {
			return nil, ocispec.Descriptor{}, err
		}
	}
	if err := tw.Close(); err != nil {
		return nil, ocispec.Descriptor{}, err
	}
	if err := gz.Close(); err != nil {
		return nil, ocispec.Descriptor{}, err
	}
	return buf.Bytes(), ocispec.Descriptor{
		Digest:    digest.FromBytes(buf.Bytes()),
		MediaType: ocispec.MediaTypeImageLayerGzip,
		Size:      int64(buf.Len()),
		Annotations: map[string]string{
			"containerd.io/uncompressed": sha.Digest().String(),
		},
	}, nil
}
