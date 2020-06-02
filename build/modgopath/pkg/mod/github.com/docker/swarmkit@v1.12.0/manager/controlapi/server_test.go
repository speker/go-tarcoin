package controlapi

import (
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/docker/swarmkit/api"
	cautils "github.com/docker/swarmkit/ca/testutils"
	"github.com/docker/swarmkit/manager/state/store"
	"github.com/stretchr/testify/assert"
)

type mockProposer struct {
	index uint64
}

func (mp *mockProposer) ProposeValue(ctx context.Context, storeAction []*api.StoreAction, cb func()) error {
	if cb != nil {
		cb()
	}
	return nil
}

func (mp *mockProposer) GetVersion() *api.Version {
	mp.index += 3
	return &api.Version{Index: mp.index}
}

type testServer struct {
	Server *Server
	Client api.ControlClient
	Store  *store.MemoryStore

	grpcServer *grpc.Server
	clientConn *grpc.ClientConn
}

func (ts *testServer) Stop() {
	_ = ts.clientConn.Close()
	ts.grpcServer.Stop()
}

func newTestServer(t *testing.T) *testServer {
	ts := &testServer{}

	// Create a testCA just to get a usable RootCA object
	tc := cautils.NewTestCA(nil)
	tc.Stop()

	ts.Store = store.NewMemoryStore(&mockProposer{})
	assert.NotNil(t, ts.Store)
	ts.Server = NewServer(ts.Store, nil, &tc.RootCA)
	assert.NotNil(t, ts.Server)

	temp, err := ioutil.TempFile("", "test-socket")
	assert.NoError(t, err)
	assert.NoError(t, temp.Close())
	assert.NoError(t, os.Remove(temp.Name()))

	lis, err := net.Listen("unix", temp.Name())
	assert.NoError(t, err)

	ts.grpcServer = grpc.NewServer()
	api.RegisterControlServer(ts.grpcServer, ts.Server)
	go func() {
		// Serve will always return an error (even when properly stopped).
		// Explicitly ignore it.
		_ = ts.grpcServer.Serve(lis)
	}()

	conn, err := grpc.Dial(temp.Name(), grpc.WithInsecure(), grpc.WithTimeout(10*time.Second),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}))
	assert.NoError(t, err)

	ts.Client = api.NewControlClient(conn)

	return ts
}
