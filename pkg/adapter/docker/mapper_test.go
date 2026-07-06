package docker

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/stretchr/testify/assert"
)

// composeSummary builds a container.Summary with the compose labels lazydocker
// derives service/project/number/oneoff from.
func composeSummary() container.Summary {
	return container.Summary{
		ID:    "runningid",
		Names: []string{"/myproj-web-1"},
		State: "running",
		Image: "sha256:abc123",
		Ports: []container.Port{
			{PrivatePort: 80, PublicPort: 8080, Type: "tcp", IP: "0.0.0.0"},
			{PrivatePort: 443, Type: "tcp"},
		},
		Labels: map[string]string{
			"com.docker.compose.service":   "web",
			"com.docker.compose.project":   "myproj",
			"com.docker.compose.container": "1",
			"com.docker.compose.oneoff":    "False",
		},
	}
}

func TestMapContainerSummary(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		summary container.Summary
		want    domain.Container
	}{
		{
			name:    "compose_container",
			summary: composeSummary(),
			want: domain.Container{
				ID:              "runningid",
				Name:            "myproj-web-1",
				ServiceName:     "web",
				ContainerNumber: "1",
				ProjectName:     "myproj",
				OneOff:          false,
				Image:           "sha256:abc123",
				Status:          domain.StatusRunning,
				Ports: []domain.Port{
					{IP: "0.0.0.0", PublicPort: 8080, PrivatePort: 80, Proto: "tcp"},
					{PrivatePort: 443, Proto: "tcp"},
				},
				Labels: map[string]string{
					"com.docker.compose.service":   "web",
					"com.docker.compose.project":   "myproj",
					"com.docker.compose.container": "1",
					"com.docker.compose.oneoff":    "False",
				},
			},
		},
		{
			name: "name_label_wins_over_names",
			summary: container.Summary{
				ID:     "id1",
				Names:  []string{"/from-names"},
				State:  "exited",
				Labels: map[string]string{"name": "from-label"},
			},
			want: domain.Container{
				ID:     "id1",
				Name:   "from-label",
				Status: domain.StatusExited,
				Labels: map[string]string{"name": "from-label"},
			},
		},
		{
			name: "names_trimmed_when_no_name_label",
			summary: container.Summary{
				ID:    "id2",
				Names: []string{"/solo"},
				State: "created",
			},
			want: domain.Container{
				ID:     "id2",
				Name:   "solo",
				Status: domain.StatusCreated,
			},
		},
		{
			name:    "id_fallback_when_no_names",
			summary: container.Summary{ID: "id3", State: "paused"},
			want: domain.Container{
				ID:     "id3",
				Name:   "id3",
				Status: domain.StatusPaused,
			},
		},
		{
			name: "oneoff_true",
			summary: container.Summary{
				ID:     "id4",
				State:  "running",
				Labels: map[string]string{"com.docker.compose.oneoff": "True"},
			},
			want: domain.Container{
				ID:     "id4",
				Name:   "id4",
				Status: domain.StatusRunning,
				OneOff: true,
				Labels: map[string]string{"com.docker.compose.oneoff": "True"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := mapContainerSummary(tc.summary)
			assert.Equal(t, tc.want, got)
			assert.Nil(t, got.Details, "mapped summary leaves Details nil")
		})
	}
}

func TestMapPortsNil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, mapPorts(nil))
}

func TestMapInspectResponse(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		resp container.InspectResponse
		want domain.ContainerDetails
	}{
		{
			name: "healthy_running_open_stdin",
			resp: container.InspectResponse{
				ContainerJSONBase: &container.ContainerJSONBase{
					State: &container.State{
						Running:  true,
						ExitCode: 0,
						Health:   &container.Health{Status: "healthy"},
					},
				},
				Config: &container.Config{OpenStdin: true},
			},
			want: domain.ContainerDetails{
				Running:   true,
				ExitCode:  0,
				Health:    domain.HealthHealthy,
				OpenStdin: true,
			},
		},
		{
			name: "exited_nonzero_no_health",
			resp: container.InspectResponse{
				ContainerJSONBase: &container.ContainerJSONBase{
					State: &container.State{Running: false, ExitCode: 137},
				},
				Config: &container.Config{OpenStdin: false},
			},
			want: domain.ContainerDetails{
				Running:   false,
				ExitCode:  137,
				Health:    domain.HealthNone,
				OpenStdin: false,
			},
		},
		{
			name: "nil_state_and_config_safe",
			resp: container.InspectResponse{},
			want: domain.ContainerDetails{},
		},
		{
			name: "starting_health",
			resp: container.InspectResponse{
				ContainerJSONBase: &container.ContainerJSONBase{
					State: &container.State{Running: true, Health: &container.Health{Status: "starting"}},
				},
			},
			want: domain.ContainerDetails{Running: true, Health: domain.HealthStarting},
		},
		{
			name: "paused",
			resp: container.InspectResponse{
				ContainerJSONBase: &container.ContainerJSONBase{
					State: &container.State{Running: true, Paused: true},
				},
			},
			want: domain.ContainerDetails{Running: true, Paused: true, Health: domain.HealthNone},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, mapInspectResponse(tc.resp))
		})
	}
}

func TestMapContainerInspect(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		resp container.InspectResponse
		want domain.ContainerInspect
	}{
		{
			name: "rich_response",
			resp: container.InspectResponse{
				ContainerJSONBase: &container.ContainerJSONBase{
					ID:   "abc123",
					Name: "/myproj-web-1",
					Path: "nginx",
					Args: []string{"-g", "daemon off;"},
				},
				Config: &container.Config{
					Image:  "nginx:latest",
					Labels: map[string]string{"com.docker.compose.service": "web"},
					Env:    []string{"PATH=/usr/bin", "FOO=bar"},
				},
				Mounts: []container.MountPoint{
					{Type: mount.TypeVolume, Name: "data", Source: "/var/lib/docker/volumes/data", Destination: "/data"},
					{Type: mount.TypeBind, Source: "/host/path", Destination: "/container/path"},
				},
				NetworkSettings: &container.NetworkSettings{
					NetworkSettingsBase: container.NetworkSettingsBase{
						Ports: nat.PortMap{
							"443/tcp": nil,
							"80/tcp":  []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "8080"}},
						},
					},
				},
			},
			want: domain.ContainerInspect{
				Image:   "nginx:latest",
				Command: []string{"nginx", "-g", "daemon off;"},
				Labels:  map[string]string{"com.docker.compose.service": "web"},
				Env:     []string{"PATH=/usr/bin", "FOO=bar"},
				Mounts: []domain.Mount{
					{Type: "volume", Name: "data", Source: "/var/lib/docker/volumes/data", Destination: "/data"},
					{Type: "bind", Source: "/host/path", Destination: "/container/path"},
				},
				// Sorted by ContainerPort; the empty-bindings key yields an empty (non-nil) HostPorts.
				Ports: []domain.PortBinding{
					{ContainerPort: "443/tcp", HostPorts: []string{}},
					{ContainerPort: "80/tcp", HostPorts: []string{"8080"}},
				},
			},
		},
		{
			name: "nil_base_config_network_safe",
			resp: container.InspectResponse{},
			want: domain.ContainerInspect{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := mapContainerInspect(tc.resp)
			assert.Equal(t, tc.want, got)
			assert.Empty(t, got.ID, "ID left for the GUI to supply from the store")
			assert.Empty(t, got.Name, "Name left for the GUI to supply from the store")
		})
	}
}
