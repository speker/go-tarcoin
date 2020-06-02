package tsdb_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/influxdata/influxdb/logger"
	"github.com/influxdata/influxdb/models"
	"github.com/influxdata/influxdb/tsdb"
	"golang.org/x/sync/errgroup"
)

func TestParseSeriesKeyInto(t *testing.T) {
	name := []byte("cpu")
	tags := models.NewTags(map[string]string{"region": "east", "server": "a"})
	key := tsdb.AppendSeriesKey(nil, name, tags)

	dst := make(models.Tags, 0)
	gotName, gotTags := tsdb.ParseSeriesKeyInto(key, dst)

	if !bytes.Equal(gotName, name) {
		t.Fatalf("got %q, expected %q", gotName, name)
	}

	if got, exp := len(gotTags), 2; got != exp {
		t.Fatalf("got tags length %d, expected %d", got, exp)
	} else if got, exp := gotTags, tags; !got.Equal(exp) {
		t.Fatalf("got tags %v, expected %v", got, exp)
	}

	dst = make(models.Tags, 0, 5)
	_, gotTags = tsdb.ParseSeriesKeyInto(key, dst)
	if got, exp := len(gotTags), 2; got != exp {
		t.Fatalf("got tags length %d, expected %d", got, exp)
	} else if got, exp := cap(gotTags), 5; got != exp {
		t.Fatalf("got tags capacity %d, expected %d", got, exp)
	} else if got, exp := gotTags, tags; !got.Equal(exp) {
		t.Fatalf("got tags %v, expected %v", got, exp)
	}

	dst = make(models.Tags, 1)
	_, gotTags = tsdb.ParseSeriesKeyInto(key, dst)
	if got, exp := len(gotTags), 2; got != exp {
		t.Fatalf("got tags length %d, expected %d", got, exp)
	} else if got, exp := gotTags, tags; !got.Equal(exp) {
		t.Fatalf("got tags %v, expected %v", got, exp)
	}
}

// Ensure that broken series files are closed
func TestSeriesFile_Open_WhenFileCorrupt_ShouldReturnErr(t *testing.T) {
	f := NewBrokenSeriesFile([]byte{0, 0, 0, 0, 0})
	defer f.Close()
	f.Logger = logger.New(os.Stdout)

	err := f.Open()

	if err == nil {
		t.Fatalf("should report error")
	}
}

// Ensure series file contains the correct set of series.
func TestSeriesFile_Series(t *testing.T) {
	sfile := MustOpenSeriesFile()
	defer sfile.Close()

	series := []Series{
		{Name: []byte("cpu"), Tags: models.NewTags(map[string]string{"region": "east"})},
		{Name: []byte("cpu"), Tags: models.NewTags(map[string]string{"region": "west"})},
		{Name: []byte("mem"), Tags: models.NewTags(map[string]string{"region": "east"})},
	}
	for _, s := range series {
		if _, err := sfile.CreateSeriesListIfNotExists([][]byte{[]byte(s.Name)}, []models.Tags{s.Tags}); err != nil {
			t.Fatal(err)
		}
	}

	// Verify total number of series is correct.
	if n := sfile.SeriesCount(); n != 3 {
		t.Fatalf("unexpected series count: %d", n)
	}

	// Verify all series exist.
	for i, s := range series {
		if seriesID := sfile.SeriesID(s.Name, s.Tags, nil); seriesID == 0 {
			t.Fatalf("series does not exist: i=%d", i)
		}
	}

	// Verify non-existent series doesn't exist.
	if sfile.HasSeries([]byte("foo"), models.NewTags(map[string]string{"region": "north"}), nil) {
		t.Fatal("series should not exist")
	}
}

// Ensure series file can be compacted.
func TestSeriesFileCompactor(t *testing.T) {
	sfile := MustOpenSeriesFile()
	defer sfile.Close()

	// Disable automatic compactions.
	for _, p := range sfile.Partitions() {
		p.CompactThreshold = 0
	}

	var names [][]byte
	var tagsSlice []models.Tags
	for i := 0; i < 10000; i++ {
		names = append(names, []byte(fmt.Sprintf("m%d", i)))
		tagsSlice = append(tagsSlice, models.NewTags(map[string]string{"foo": "bar"}))
	}
	if _, err := sfile.CreateSeriesListIfNotExists(names, tagsSlice); err != nil {
		t.Fatal(err)
	}

	// Verify total number of series is correct.
	if n := sfile.SeriesCount(); n != uint64(len(names)) {
		t.Fatalf("unexpected series count: %d", n)
	}

	// Compact in-place for each partition.
	for _, p := range sfile.Partitions() {
		compactor := tsdb.NewSeriesPartitionCompactor()
		if err := compactor.Compact(p); err != nil {
			t.Fatal(err)
		}
	}

	// Verify all series exist.
	for i := range names {
		if seriesID := sfile.SeriesID(names[i], tagsSlice[i], nil); seriesID == 0 {
			t.Fatalf("series does not exist: %s,%s", names[i], tagsSlice[i].String())
		}
	}
}

// Ensure series file deletions persist across compactions.
func TestSeriesFile_DeleteSeriesID(t *testing.T) {
	sfile := MustOpenSeriesFile()
	defer sfile.Close()

	ids0, err := sfile.CreateSeriesListIfNotExists([][]byte{[]byte("m1")}, []models.Tags{nil})
	if err != nil {
		t.Fatal(err)
	} else if _, err := sfile.CreateSeriesListIfNotExists([][]byte{[]byte("m2")}, []models.Tags{nil}); err != nil {
		t.Fatal(err)
	} else if err := sfile.ForceCompact(); err != nil {
		t.Fatal(err)
	}

	// Delete and ensure deletion.
	if err := sfile.DeleteSeriesID(ids0[0]); err != nil {
		t.Fatal(err)
	} else if _, err := sfile.CreateSeriesListIfNotExists([][]byte{[]byte("m1")}, []models.Tags{nil}); err != nil {
		t.Fatal(err)
	} else if !sfile.IsDeleted(ids0[0]) {
		t.Fatal("expected deletion before compaction")
	}

	if err := sfile.ForceCompact(); err != nil {
		t.Fatal(err)
	} else if !sfile.IsDeleted(ids0[0]) {
		t.Fatal("expected deletion after compaction")
	}

	if err := sfile.Reopen(); err != nil {
		t.Fatal(err)
	} else if !sfile.IsDeleted(ids0[0]) {
		t.Fatal("expected deletion after reopen")
	}
}

func TestSeriesFile_Compaction(t *testing.T) {
	sfile := MustOpenSeriesFile()
	defer sfile.Close()

	// Generate a bunch of keys.
	var mms [][]byte
	var tagSets []models.Tags
	for i := 0; i < 1000; i++ {
		mms = append(mms, []byte("cpu"))
		tagSets = append(tagSets, models.NewTags(map[string]string{"region": fmt.Sprintf("r%d", i)}))
	}

	// Add all to the series file.
	ids, err := sfile.CreateSeriesListIfNotExists(mms, tagSets)
	if err != nil {
		t.Fatal(err)
	}

	// Delete a subset of keys.
	for i, id := range ids {
		if i%10 == 0 {
			if err := sfile.DeleteSeriesID(id); err != nil {
				t.Fatal(err)
			}
		}
	}

	// Compute total size of all series data.
	origSize, err := sfile.FileSize()
	if err != nil {
		t.Fatal(err)
	}

	// Compact all segments.
	var paths []string
	for _, p := range sfile.Partitions() {
		for _, ss := range p.Segments() {
			if err := ss.CompactToPath(ss.Path()+".tmp", p.Index()); err != nil {
				t.Fatal(err)
			}
			paths = append(paths, ss.Path())
		}
	}

	// Close index.
	if err := sfile.SeriesFile.Close(); err != nil {
		t.Fatal(err)
	}

	// Overwrite files.
	for _, path := range paths {
		if err := os.Rename(path+".tmp", path); err != nil {
			t.Fatal(err)
		}
	}

	// Reopen index.
	sfile.SeriesFile = tsdb.NewSeriesFile(sfile.SeriesFile.Path())
	if err := sfile.SeriesFile.Open(); err != nil {
		t.Fatal(err)
	}

	// Ensure series status is correct.
	for i, id := range ids {
		if got, want := sfile.IsDeleted(id), (i%10) == 0; got != want {
			t.Fatalf("IsDeleted(%d)=%v, want %v", id, got, want)
		}
	}

	// Verify new size is smaller.
	newSize, err := sfile.FileSize()
	if err != nil {
		t.Fatal(err)
	} else if newSize >= origSize {
		t.Fatalf("expected new size (%d) to be smaller than original size (%d)", newSize, origSize)
	}

	t.Logf("original size: %d, new size: %d", origSize, newSize)
}

var cachedCompactionSeriesFile *SeriesFile

func BenchmarkSeriesFile_Compaction(b *testing.B) {
	const n = 1000000

	if cachedCompactionSeriesFile == nil {
		sfile := MustOpenSeriesFile()

		// Generate a bunch of keys.
		var ids []uint64
		for i := 0; i < n; i++ {
			tmp, err := sfile.CreateSeriesListIfNotExists([][]byte{[]byte("cpu")}, []models.Tags{models.NewTags(map[string]string{"region": fmt.Sprintf("r%d", i)})})
			if err != nil {
				b.Fatal(err)
			}
			ids = append(ids, tmp...)
		}

		// Delete a subset of keys.
		for i := 0; i < len(ids); i += 10 {
			if err := sfile.DeleteSeriesID(ids[i]); err != nil {
				b.Fatal(err)
			}
		}

		cachedCompactionSeriesFile = sfile
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Compact all segments in parallel.
		var g errgroup.Group
		for _, p := range cachedCompactionSeriesFile.Partitions() {
			for _, segment := range p.Segments() {
				p, segment := p, segment
				g.Go(func() error {
					return segment.CompactToPath(segment.Path()+".tmp", p.Index())
				})
			}
		}

		if err := g.Wait(); err != nil {
			b.Fatal(err)
		}
	}
}

// Series represents name/tagset pairs that are used in testing.
type Series struct {
	Name    []byte
	Tags    models.Tags
	Deleted bool
}

// SeriesFile is a test wrapper for tsdb.SeriesFile.
type SeriesFile struct {
	*tsdb.SeriesFile
}

// NewSeriesFile returns a new instance of SeriesFile with a temporary file path.
func NewSeriesFile() *SeriesFile {
	dir, err := ioutil.TempDir("", "tsdb-series-file-")
	if err != nil {
		panic(err)
	}
	return &SeriesFile{SeriesFile: tsdb.NewSeriesFile(dir)}
}

func NewBrokenSeriesFile(content []byte) *SeriesFile {
	sFile := NewSeriesFile()
	fPath := sFile.Path()
	sFile.Open()
	sFile.SeriesFile.Close()

	segPath := path.Join(fPath, "00", "0000")
	if _, err := os.Stat(segPath); os.IsNotExist(err) {
		panic(err)
	}
	err := ioutil.WriteFile(segPath, content, 0777)
	if err != nil {
		panic(err)
	}
	return sFile
}

// MustOpenSeriesFile returns a new, open instance of SeriesFile. Panic on error.
func MustOpenSeriesFile() *SeriesFile {
	f := NewSeriesFile()
	f.Logger = logger.New(os.Stdout)
	if err := f.Open(); err != nil {
		panic(err)
	}
	return f
}

// Close closes the log file and removes it from disk.
func (f *SeriesFile) Close() error {
	defer os.RemoveAll(f.Path())
	return f.SeriesFile.Close()
}

// Reopen close & reopens the series file.
func (f *SeriesFile) Reopen() error {
	if err := f.SeriesFile.Close(); err != nil {
		return err
	}
	f.SeriesFile = tsdb.NewSeriesFile(f.SeriesFile.Path())
	return f.SeriesFile.Open()
}

// ForceCompact executes an immediate compaction across all partitions.
func (f *SeriesFile) ForceCompact() error {
	for _, p := range f.Partitions() {
		if err := tsdb.NewSeriesPartitionCompactor().Compact(p); err != nil {
			return err
		}
	}
	return nil
}
