package container

import (
	"testing"

	"github.com/docker/swarmkit/api"
)

func TestVolumesAndBinds(t *testing.T) {
	c := containerConfig{
		task: &api.Task{
			Spec: api.TaskSpec{Runtime: &api.TaskSpec_Container{
				Container: &api.ContainerSpec{
					Mounts: []api.Mount{
						{Type: api.MountTypeBind, Source: "/banana", Target: "/kerfluffle"},
						{Type: api.MountTypeBind, Source: "/banana", Target: "/kerfluffle", BindOptions: &api.Mount_BindOptions{Propagation: api.MountPropagationRPrivate}},
						{Type: api.MountTypeVolume, Source: "banana", Target: "/kerfluffle"},
						{Type: api.MountTypeVolume, Source: "banana", Target: "/kerfluffle", VolumeOptions: &api.Mount_VolumeOptions{NoCopy: true}},
						{Type: api.MountTypeVolume, Target: "/kerfluffle"},
					},
				},
			}},
		},
	}

	config := c.config()
	if len(config.Volumes) != 1 {
		t.Fatalf("expected only 1 anonymous volume: %v", config.Volumes)
	}
	if _, exists := config.Volumes["/kerfluffle"]; !exists {
		t.Fatal("missing anonymous volume entry for target `/kerfluffle`")
	}

	hostConfig := c.hostConfig()
	if len(hostConfig.Binds) != 4 {
		t.Fatalf("exepcted 4 binds: %v", hostConfig.Binds)
	}

	expected := "/banana:/kerfluffle"
	actual := hostConfig.Binds[0]
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}

	expected = "/banana:/kerfluffle:rprivate"
	actual = hostConfig.Binds[1]
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}

	expected = "banana:/kerfluffle"
	actual = hostConfig.Binds[2]
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}

	expected = "banana:/kerfluffle:nocopy"
	actual = hostConfig.Binds[3]
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}
