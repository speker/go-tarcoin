package controlapi

import (
	"strings"
	"testing"
	"time"

	"github.com/docker/swarmkit/api"
	"github.com/docker/swarmkit/api/duration"
	"github.com/docker/swarmkit/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func createSpec(name, image string, instances uint64) *api.ServiceSpec {
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
	}
}

func createService(t *testing.T, ts *testServer, name, image string, instances uint64) *api.Service {
	spec := createSpec(name, image, instances)
	r, err := ts.Client.CreateService(context.Background(), &api.CreateServiceRequest{Spec: spec})
	assert.NoError(t, err)
	return r.Service
}

func TestValidateResources(t *testing.T) {
	bad := []*api.Resources{
		{MemoryBytes: 1},
		{NanoCPUs: 42},
	}

	good := []*api.Resources{
		{MemoryBytes: 4096 * 1024 * 1024},
		{NanoCPUs: 1e9},
	}

	for _, b := range bad {
		err := validateResources(b)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, grpc.Code(err))
	}

	for _, g := range good {
		assert.NoError(t, validateResources(g))
	}
}

func TestValidateResourceRequirements(t *testing.T) {
	bad := []*api.ResourceRequirements{
		{Limits: &api.Resources{MemoryBytes: 1}},
		{Reservations: &api.Resources{MemoryBytes: 1}},
	}
	good := []*api.ResourceRequirements{
		{Limits: &api.Resources{NanoCPUs: 1e9}},
		{Reservations: &api.Resources{NanoCPUs: 1e9}},
	}
	for _, b := range bad {
		err := validateResourceRequirements(b)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, grpc.Code(err))
	}

	for _, g := range good {
		assert.NoError(t, validateResourceRequirements(g))
	}
}

func TestValidateTask(t *testing.T) {
	type badSource struct {
		s api.TaskSpec
		c codes.Code
	}

	for _, bad := range []badSource{
		{
			s: api.TaskSpec{
				Runtime: &api.TaskSpec_Container{
					Container: nil,
				},
			},
			c: codes.InvalidArgument,
		},
		// NOTE(stevvooe): can't actually test this case because we don't have
		// another runtime defined.
		// {
		//	s: &api.ServiceSpec{
		//		Template: &api.TaskSpec{
		//			Runtime:
		//		},
		//	},
		//	c: codes.Unimplemented,
		// },
		{
			s: createSpec("", "", 0).Task,
			c: codes.InvalidArgument,
		},
		{
			s: createSpec("", "busybox###", 0).Task,
			c: codes.InvalidArgument,
		},
	} {
		err := validateTask(bad.s)
		assert.Error(t, err)
		assert.Equal(t, bad.c, grpc.Code(err))
	}

	for _, good := range []api.TaskSpec{
		createSpec("", "image", 0).Task,
	} {
		err := validateTask(good)
		assert.NoError(t, err)
	}
}

func TestValidateServiceSpec(t *testing.T) {
	type BadServiceSpec struct {
		spec *api.ServiceSpec
		c    codes.Code
	}

	for _, bad := range []BadServiceSpec{
		{
			spec: nil,
			c:    codes.InvalidArgument,
		},
		{
			spec: &api.ServiceSpec{Annotations: api.Annotations{Name: "name"}},
			c:    codes.InvalidArgument,
		},
		{
			spec: createSpec("", "", 1),
			c:    codes.InvalidArgument,
		},
		{
			spec: createSpec("name", "", 1),
			c:    codes.InvalidArgument,
		},
		{
			spec: createSpec("", "image", 1),
			c:    codes.InvalidArgument,
		},
	} {
		err := validateServiceSpec(bad.spec)
		assert.Error(t, err)
		assert.Equal(t, bad.c, grpc.Code(err), grpc.ErrorDesc(err))
	}

	for _, good := range []*api.ServiceSpec{
		createSpec("name", "image", 1),
	} {
		err := validateServiceSpec(good)
		assert.NoError(t, err)
	}
}

func TestValidateRestartPolicy(t *testing.T) {
	bad := []*api.RestartPolicy{
		{
			Delay:  ptypes.DurationProto(time.Duration(-1 * time.Second)),
			Window: ptypes.DurationProto(time.Duration(-1 * time.Second)),
		},
		{
			Delay:  ptypes.DurationProto(time.Duration(20 * time.Second)),
			Window: ptypes.DurationProto(time.Duration(-4 * time.Second)),
		},
	}

	good := []*api.RestartPolicy{
		{
			Delay:  ptypes.DurationProto(time.Duration(10 * time.Second)),
			Window: ptypes.DurationProto(time.Duration(1 * time.Second)),
		},
	}

	for _, b := range bad {
		err := validateRestartPolicy(b)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, grpc.Code(err))
	}

	for _, g := range good {
		assert.NoError(t, validateRestartPolicy(g))
	}
}

func TestValidateUpdate(t *testing.T) {
	bad := []*api.UpdateConfig{
		{Delay: duration.Duration{Seconds: -1, Nanos: 0}},
		{Delay: duration.Duration{Seconds: -1000, Nanos: 0}},
	}

	good := []*api.UpdateConfig{
		{Delay: duration.Duration{Seconds: 1, Nanos: 0}},
	}

	for _, b := range bad {
		err := validateUpdate(b)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, grpc.Code(err))
	}

	for _, g := range good {
		assert.NoError(t, validateUpdate(g))
	}
}

func TestCreateService(t *testing.T) {
	ts := newTestServer(t)
	_, err := ts.Client.CreateService(context.Background(), &api.CreateServiceRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))

	spec := createSpec("name", "image", 1)
	r, err := ts.Client.CreateService(context.Background(), &api.CreateServiceRequest{Spec: spec})
	assert.NoError(t, err)
	assert.NotEmpty(t, r.Service.ID)

	// test port conflicts
	spec = createSpec("name2", "image", 1)
	spec.Endpoint = &api.EndpointSpec{Ports: []*api.PortConfig{
		{PublishedPort: uint32(9000), TargetPort: uint32(9000), Protocol: api.PortConfig_Protocol(api.ProtocolTCP)},
	}}
	r, err = ts.Client.CreateService(context.Background(), &api.CreateServiceRequest{Spec: spec})
	assert.NoError(t, err)
	assert.NotEmpty(t, r.Service.ID)

	spec2 := createSpec("name3", "image", 1)
	spec2.Endpoint = &api.EndpointSpec{Ports: []*api.PortConfig{
		{PublishedPort: uint32(9000), TargetPort: uint32(9000), Protocol: api.PortConfig_Protocol(api.ProtocolTCP)},
	}}
	_, err = ts.Client.CreateService(context.Background(), &api.CreateServiceRequest{Spec: spec2})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))

	// test no port conflicts when no publish port is specified
	spec3 := createSpec("name4", "image", 1)
	spec3.Endpoint = &api.EndpointSpec{Ports: []*api.PortConfig{
		{TargetPort: uint32(9000), Protocol: api.PortConfig_Protocol(api.ProtocolTCP)},
	}}
	r, err = ts.Client.CreateService(context.Background(), &api.CreateServiceRequest{Spec: spec3})
	assert.NoError(t, err)
	assert.NotEmpty(t, r.Service.ID)
	spec4 := createSpec("name5", "image", 1)
	spec4.Endpoint = &api.EndpointSpec{Ports: []*api.PortConfig{
		{TargetPort: uint32(9001), Protocol: api.PortConfig_Protocol(api.ProtocolTCP)},
	}}
	_, err = ts.Client.CreateService(context.Background(), &api.CreateServiceRequest{Spec: spec4})
	assert.NoError(t, err)
}

func TestGetService(t *testing.T) {
	ts := newTestServer(t)
	_, err := ts.Client.GetService(context.Background(), &api.GetServiceRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))

	_, err = ts.Client.GetService(context.Background(), &api.GetServiceRequest{ServiceID: "invalid"})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, grpc.Code(err))

	service := createService(t, ts, "name", "image", 1)
	r, err := ts.Client.GetService(context.Background(), &api.GetServiceRequest{ServiceID: service.ID})
	assert.NoError(t, err)
	service.Meta.Version = r.Service.Meta.Version
	assert.Equal(t, service, r.Service)
}

func TestUpdateService(t *testing.T) {
	ts := newTestServer(t)
	service := createService(t, ts, "name", "image", 1)

	_, err := ts.Client.UpdateService(context.Background(), &api.UpdateServiceRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))

	_, err = ts.Client.UpdateService(context.Background(), &api.UpdateServiceRequest{ServiceID: "invalid", Spec: &service.Spec, ServiceVersion: &api.Version{}})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, grpc.Code(err))

	// No update options.
	_, err = ts.Client.UpdateService(context.Background(), &api.UpdateServiceRequest{ServiceID: service.ID, Spec: &service.Spec})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))

	_, err = ts.Client.UpdateService(context.Background(), &api.UpdateServiceRequest{ServiceID: service.ID, Spec: &service.Spec, ServiceVersion: &service.Meta.Version})
	assert.NoError(t, err)

	r, err := ts.Client.GetService(context.Background(), &api.GetServiceRequest{ServiceID: service.ID})
	assert.NoError(t, err)
	assert.Equal(t, service.Spec.Annotations.Name, r.Service.Spec.Annotations.Name)
	mode, ok := r.Service.Spec.GetMode().(*api.ServiceSpec_Replicated)
	assert.Equal(t, ok, true)
	assert.True(t, mode.Replicated.Replicas == 1)

	mode.Replicated.Replicas = 42
	_, err = ts.Client.UpdateService(context.Background(), &api.UpdateServiceRequest{
		ServiceID:      service.ID,
		Spec:           &r.Service.Spec,
		ServiceVersion: &r.Service.Meta.Version,
	})
	assert.NoError(t, err)

	r, err = ts.Client.GetService(context.Background(), &api.GetServiceRequest{ServiceID: service.ID})
	assert.NoError(t, err)
	assert.Equal(t, service.Spec.Annotations.Name, r.Service.Spec.Annotations.Name)
	mode, ok = r.Service.Spec.GetMode().(*api.ServiceSpec_Replicated)
	assert.Equal(t, ok, true)
	assert.True(t, mode.Replicated.Replicas == 42)

	// mode change not allowed
	r, err = ts.Client.GetService(context.Background(), &api.GetServiceRequest{ServiceID: service.ID})
	assert.NoError(t, err)
	r.Service.Spec.Mode = &api.ServiceSpec_Global{
		Global: &api.GlobalService{},
	}
	_, err = ts.Client.UpdateService(context.Background(), &api.UpdateServiceRequest{
		ServiceID:      service.ID,
		Spec:           &r.Service.Spec,
		ServiceVersion: &r.Service.Meta.Version,
	})
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), errModeChangeNotAllowed.Error()))

	// Versioning.
	r, err = ts.Client.GetService(context.Background(), &api.GetServiceRequest{ServiceID: service.ID})
	assert.NoError(t, err)
	version := &r.Service.Meta.Version

	_, err = ts.Client.UpdateService(context.Background(), &api.UpdateServiceRequest{
		ServiceID:      service.ID,
		Spec:           &r.Service.Spec,
		ServiceVersion: version,
	})
	assert.NoError(t, err)

	// Perform an update with the "old" version.
	_, err = ts.Client.UpdateService(context.Background(), &api.UpdateServiceRequest{
		ServiceID:      service.ID,
		Spec:           &r.Service.Spec,
		ServiceVersion: version,
	})
	assert.Error(t, err)

	// test port conflicts
	spec2 := createSpec("name2", "image", 1)
	spec2.Endpoint = &api.EndpointSpec{Ports: []*api.PortConfig{
		{PublishedPort: uint32(9000), TargetPort: uint32(9000), Protocol: api.PortConfig_Protocol(api.ProtocolTCP)},
	}}
	_, err = ts.Client.CreateService(context.Background(), &api.CreateServiceRequest{Spec: spec2})
	assert.NoError(t, err)

	spec3 := createSpec("name3", "image", 1)
	rs, err := ts.Client.CreateService(context.Background(), &api.CreateServiceRequest{Spec: spec3})
	assert.NoError(t, err)

	spec3.Endpoint = &api.EndpointSpec{Ports: []*api.PortConfig{
		{PublishedPort: uint32(9000), TargetPort: uint32(9000), Protocol: api.PortConfig_Protocol(api.ProtocolTCP)},
	}}
	_, err = ts.Client.UpdateService(context.Background(), &api.UpdateServiceRequest{
		ServiceID:      rs.Service.ID,
		Spec:           spec3,
		ServiceVersion: &rs.Service.Meta.Version,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))
	spec3.Endpoint = &api.EndpointSpec{Ports: []*api.PortConfig{
		{PublishedPort: uint32(9001), TargetPort: uint32(9000), Protocol: api.PortConfig_Protocol(api.ProtocolTCP)},
	}}
	_, err = ts.Client.UpdateService(context.Background(), &api.UpdateServiceRequest{
		ServiceID:      rs.Service.ID,
		Spec:           spec3,
		ServiceVersion: &rs.Service.Meta.Version,
	})
	assert.NoError(t, err)
}

// TODO(dongluochen): Network update is not supported yet and it's blocked
// from controlapi. This test should be removed once network update is supported.
func TestServiceUpdateRejectNetworkChange(t *testing.T) {
	ts := newTestServer(t)
	spec := createSpec("name", "image", 1)
	spec.Networks = []*api.ServiceSpec_NetworkAttachmentConfig{
		{
			Target: "net20",
		},
	}
	cr, err := ts.Client.CreateService(context.Background(), &api.CreateServiceRequest{Spec: spec})
	assert.NoError(t, err)

	ur, err := ts.Client.GetService(context.Background(), &api.GetServiceRequest{ServiceID: cr.Service.ID})
	assert.NoError(t, err)
	service := ur.Service

	service.Spec.Networks[0].Target = "net30"

	_, err = ts.Client.UpdateService(context.Background(), &api.UpdateServiceRequest{
		ServiceID:      service.ID,
		Spec:           &service.Spec,
		ServiceVersion: &service.Meta.Version,
	})
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), errNetworkUpdateNotSupported.Error()))
}

func TestRemoveService(t *testing.T) {
	ts := newTestServer(t)
	_, err := ts.Client.RemoveService(context.Background(), &api.RemoveServiceRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))

	service := createService(t, ts, "name", "image", 1)
	r, err := ts.Client.RemoveService(context.Background(), &api.RemoveServiceRequest{ServiceID: service.ID})
	assert.NoError(t, err)
	assert.NotNil(t, r)
}

func TestValidateEndpointSpec(t *testing.T) {
	err := validateEndpointSpec(&api.EndpointSpec{
		Mode: api.ResolutionModeDNSRoundRobin,
		Ports: []*api.PortConfig{
			{
				Name: "http", TargetPort: 80,
			},
		},
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))
}

func TestServiceEndpointSpecUpdate(t *testing.T) {
	ts := newTestServer(t)
	spec := &api.ServiceSpec{
		Annotations: api.Annotations{
			Name: "name",
		},
		Task: api.TaskSpec{
			Runtime: &api.TaskSpec_Container{
				Container: &api.ContainerSpec{
					Image: "image",
				},
			},
		},
		Mode: &api.ServiceSpec_Replicated{
			Replicated: &api.ReplicatedService{
				Replicas: 1,
			},
		},
		Endpoint: &api.EndpointSpec{
			Ports: []*api.PortConfig{
				{
					Name:       "http",
					TargetPort: 80,
				},
			},
		},
	}

	r, err := ts.Client.CreateService(context.Background(),
		&api.CreateServiceRequest{Spec: spec})
	assert.NoError(t, err)
	assert.NotNil(t, r)

	// Update the service with duplicate ports
	spec.Endpoint.Ports = append(spec.Endpoint.Ports, &api.PortConfig{
		Name:       "fakehttp",
		TargetPort: 80,
	})
	_, err = ts.Client.UpdateService(context.Background(),
		&api.UpdateServiceRequest{Spec: spec})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, grpc.Code(err))
}

func TestListServices(t *testing.T) {
	ts := newTestServer(t)
	r, err := ts.Client.ListServices(context.Background(), &api.ListServicesRequest{})
	assert.NoError(t, err)
	assert.Empty(t, r.Services)

	s1 := createService(t, ts, "name1", "image", 1)
	r, err = ts.Client.ListServices(context.Background(), &api.ListServicesRequest{})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(r.Services))

	createService(t, ts, "name2", "image", 1)
	createService(t, ts, "name3", "image", 1)

	// List all.
	r, err = ts.Client.ListServices(context.Background(), &api.ListServicesRequest{})
	assert.NoError(t, err)
	assert.Equal(t, 3, len(r.Services))

	// List with an ID prefix.
	r, err = ts.Client.ListServices(context.Background(), &api.ListServicesRequest{
		Filters: &api.ListServicesRequest_Filters{
			IDPrefixes: []string{s1.ID[0:4]},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(r.Services))
	assert.Equal(t, s1.ID, r.Services[0].ID)

	// List with simple filter.
	r, err = ts.Client.ListServices(context.Background(), &api.ListServicesRequest{
		Filters: &api.ListServicesRequest_Filters{
			NamePrefixes: []string{"name1"},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(r.Services))

	// List with union filter.
	r, err = ts.Client.ListServices(context.Background(), &api.ListServicesRequest{
		Filters: &api.ListServicesRequest_Filters{
			NamePrefixes: []string{"name1", "name2"},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(r.Services))

	r, err = ts.Client.ListServices(context.Background(), &api.ListServicesRequest{
		Filters: &api.ListServicesRequest_Filters{
			NamePrefixes: []string{"name1", "name2", "name4"},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(r.Services))

	r, err = ts.Client.ListServices(context.Background(), &api.ListServicesRequest{
		Filters: &api.ListServicesRequest_Filters{
			NamePrefixes: []string{"name4"},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 0, len(r.Services))

	// List with filter intersection.
	r, err = ts.Client.ListServices(context.Background(),
		&api.ListServicesRequest{
			Filters: &api.ListServicesRequest_Filters{
				NamePrefixes: []string{"name1"},
				IDPrefixes:   []string{s1.ID},
			},
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(r.Services))

	r, err = ts.Client.ListServices(context.Background(),
		&api.ListServicesRequest{
			Filters: &api.ListServicesRequest_Filters{
				NamePrefixes: []string{"name2"},
				IDPrefixes:   []string{s1.ID},
			},
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(r.Services))

	// List filter by label.
	r, err = ts.Client.ListServices(context.Background(),
		&api.ListServicesRequest{
			Filters: &api.ListServicesRequest_Filters{
				Labels: map[string]string{
					"common": "yes",
				},
			},
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(r.Services))

	// Value-less label.
	r, err = ts.Client.ListServices(context.Background(),
		&api.ListServicesRequest{
			Filters: &api.ListServicesRequest_Filters{
				Labels: map[string]string{
					"common": "",
				},
			},
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(r.Services))

	// Label intersection.
	r, err = ts.Client.ListServices(context.Background(),
		&api.ListServicesRequest{
			Filters: &api.ListServicesRequest_Filters{
				Labels: map[string]string{
					"common": "",
					"unique": "name1",
				},
			},
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(r.Services))

	r, err = ts.Client.ListServices(context.Background(),
		&api.ListServicesRequest{
			Filters: &api.ListServicesRequest_Filters{
				Labels: map[string]string{
					"common": "",
					"unique": "error",
				},
			},
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(r.Services))
}
