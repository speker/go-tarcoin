package controlapi

import (
	"testing"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/docker/swarmkit/api"
	"github.com/docker/swarmkit/identity"
	"github.com/docker/swarmkit/manager/state/store"
	"github.com/stretchr/testify/assert"
)

func createNetworkSpec(name string) *api.NetworkSpec {
	return &api.NetworkSpec{
		Annotations: api.Annotations{
			Name: name,
		},
	}
}

// createInternalNetwork creates an internal network for testing. it is the same
// as Server.CreateNetwork except without the label check.
func (s *Server) createInternalNetwork(ctx context.Context, request *api.CreateNetworkRequest) (*api.CreateNetworkResponse, error) {
	if err := validateNetworkSpec(request.Spec); err != nil {
		return nil, err
	}

	// TODO(mrjana): Consider using `Name` as a primary key to handle
	// duplicate creations. See #65
	n := &api.Network{
		ID:   identity.NewID(),
		Spec: *request.Spec,
	}

	err := s.store.Update(func(tx store.Tx) error {
		return store.CreateNetwork(tx, n)
	})
	if err != nil {
		return nil, err
	}

	return &api.CreateNetworkResponse{
		Network: n,
	}, nil
}

func createServiceInNetworkSpec(name, image string, nwid string, instances uint64) *api.ServiceSpec {
	return &api.ServiceSpec{
		Annotations: api.Annotations{
			Name: name,
			Labels: map[string]string{
				"common": "yes",
				"unique": name,
			},
		},
		Task: api.TaskSpec{
			Runtime: &api.TaskSpec_Container{
				Container: &api.ContainerSpec{
					Image: image,
				},
			},
		},
		Mode: &api.ServiceSpec_Replicated{
			Replicated: &api.ReplicatedService{
				Replicas: instances,
			},
		},
		Networks: []*api.ServiceSpec_NetworkAttachmentConfig{
			{
				Target: nwid,
			},
		},
	}
}

func createServiceInNetwork(t *testing.T, ts *testServer, name, image string, nwid string, instances uint64) *api.Service {
	spec := createServiceInNetworkSpec(name, image, nwid, instances)
	r, err := ts.Client.CreateService(context.Background(), &api.CreateServiceRequest{Spec: spec})
	assert.NoError(t, err)
	return r.Service
}

func TestValidateDriver(t *testing.T) {
	assert.NoError(t, validateDriver(nil))

	err := validateDriver(&api.Driver{Name: ""})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))
}

func TestValidateIPAMConfiguration(t *testing.T) {
	err := validateIPAMConfiguration(nil)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))

	IPAMConf := &api.IPAMConfig{
		Subnet: "",
	}

	err = validateIPAMConfiguration(IPAMConf)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))

	IPAMConf.Subnet = "bad"
	err = validateIPAMConfiguration(IPAMConf)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))

	IPAMConf.Subnet = "192.168.0.0/16"
	err = validateIPAMConfiguration(IPAMConf)
	assert.NoError(t, err)

	IPAMConf.Range = "bad"
	err = validateIPAMConfiguration(IPAMConf)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))

	IPAMConf.Range = "192.169.1.0/24"
	err = validateIPAMConfiguration(IPAMConf)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))

	IPAMConf.Range = "192.168.1.0/24"
	err = validateIPAMConfiguration(IPAMConf)
	assert.NoError(t, err)

	IPAMConf.Gateway = "bad"
	err = validateIPAMConfiguration(IPAMConf)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))

	IPAMConf.Gateway = "192.169.1.1"
	err = validateIPAMConfiguration(IPAMConf)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))

	IPAMConf.Gateway = "192.168.1.1"
	err = validateIPAMConfiguration(IPAMConf)
	assert.NoError(t, err)
}

func TestValidateIPAM(t *testing.T) {
	assert.NoError(t, validateIPAM(nil))
	ipam := &api.IPAMOptions{
		Driver: &api.Driver{
			Name: "external",
		},
	}

	err := validateIPAM(ipam)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))
}

func TestValidateNetworkSpec(t *testing.T) {
	err := validateNetworkSpec(nil)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))

	spec := createNetworkSpec("invalid_driver")
	spec.DriverConfig = &api.Driver{
		Name: "external",
	}

	err = validateNetworkSpec(spec)
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))
}

func TestCreateNetwork(t *testing.T) {
	ts := newTestServer(t)
	nr, err := ts.Client.CreateNetwork(context.Background(), &api.CreateNetworkRequest{
		Spec: createNetworkSpec("testnet1"),
	})
	assert.NoError(t, err)
	assert.NotEqual(t, nr.Network, nil)
	assert.NotEqual(t, nr.Network.ID, "")
}

func TestCreateInternalNetwork(t *testing.T) {
	ts := newTestServer(t)
	spec := createNetworkSpec("testnetint")
	spec.Annotations.Labels = map[string]string{"com.docker.swarm.internal": "true"}
	nr, err := ts.Client.CreateNetwork(context.Background(), &api.CreateNetworkRequest{
		Spec: spec,
	})
	assert.Error(t, err)
	assert.Equal(t, grpc.Code(err), codes.PermissionDenied)
	assert.Nil(t, nr)
}

func TestGetNetwork(t *testing.T) {
	ts := newTestServer(t)
	nr, err := ts.Client.CreateNetwork(context.Background(), &api.CreateNetworkRequest{
		Spec: createNetworkSpec("testnet2"),
	})
	assert.NoError(t, err)
	assert.NotEqual(t, nr.Network, nil)
	assert.NotEqual(t, nr.Network.ID, "")

	_, err = ts.Client.GetNetwork(context.Background(), &api.GetNetworkRequest{NetworkID: nr.Network.ID})
	assert.NoError(t, err)
}

func TestRemoveNetwork(t *testing.T) {
	ts := newTestServer(t)
	nr, err := ts.Client.CreateNetwork(context.Background(), &api.CreateNetworkRequest{
		Spec: createNetworkSpec("testnet2"),
	})
	assert.NoError(t, err)
	assert.NotEqual(t, nr.Network, nil)
	assert.NotEqual(t, nr.Network.ID, "")

	_, err = ts.Client.RemoveNetwork(context.Background(), &api.RemoveNetworkRequest{NetworkID: nr.Network.ID})
	assert.NoError(t, err)
}

func TestRemoveNetworkWithAttachedService(t *testing.T) {
	ts := newTestServer(t)
	nr, err := ts.Client.CreateNetwork(context.Background(), &api.CreateNetworkRequest{
		Spec: createNetworkSpec("testnet4"),
	})
	assert.NoError(t, err)
	assert.NotEqual(t, nr.Network, nil)
	assert.NotEqual(t, nr.Network.ID, "")
	createServiceInNetwork(t, ts, "name", "image", nr.Network.ID, 1)
	_, err = ts.Client.RemoveNetwork(context.Background(), &api.RemoveNetworkRequest{NetworkID: nr.Network.ID})
	assert.Error(t, err)
}

func TestRemoveInternalNetwork(t *testing.T) {
	ts := newTestServer(t)
	spec := createNetworkSpec("testnet3")
	// add label denoting internal network
	spec.Annotations.Labels = map[string]string{"com.docker.swarm.internal": "true"}
	nr, err := ts.Server.createInternalNetwork(context.Background(), &api.CreateNetworkRequest{
		Spec: spec,
	})
	assert.NoError(t, err)
	assert.NotEqual(t, nr.Network, nil)
	assert.NotEqual(t, nr.Network.ID, "")

	_, err = ts.Client.RemoveNetwork(context.Background(), &api.RemoveNetworkRequest{NetworkID: nr.Network.ID})
	// this SHOULD fail, because the internal network cannot be removed
	assert.Error(t, err)
	assert.Equal(t, grpc.Code(err), codes.PermissionDenied)

	// then, check to make sure network is still there
	ng, err := ts.Client.GetNetwork(context.Background(), &api.GetNetworkRequest{NetworkID: nr.Network.ID})
	assert.NoError(t, err)
	assert.NotEqual(t, ng.Network, nil)
	assert.NotEqual(t, ng.Network.ID, "")
}

func TestListNetworks(t *testing.T) {
	ts := newTestServer(t)

	nr1, err := ts.Client.CreateNetwork(context.Background(), &api.CreateNetworkRequest{
		Spec: createNetworkSpec("listtestnet1"),
	})
	assert.NoError(t, err)
	assert.NotEqual(t, nr1.Network, nil)
	assert.NotEqual(t, nr1.Network.ID, "")

	nr2, err := ts.Client.CreateNetwork(context.Background(), &api.CreateNetworkRequest{
		Spec: createNetworkSpec("listtestnet2"),
	})
	assert.NoError(t, err)
	assert.NotEqual(t, nr2.Network, nil)
	assert.NotEqual(t, nr2.Network.ID, "")

	r, err := ts.Client.ListNetworks(context.Background(), &api.ListNetworksRequest{})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(r.Networks))
	assert.True(t, r.Networks[0].ID == nr1.Network.ID || r.Networks[0].ID == nr2.Network.ID)
	assert.True(t, r.Networks[1].ID == nr1.Network.ID || r.Networks[1].ID == nr2.Network.ID)
}
