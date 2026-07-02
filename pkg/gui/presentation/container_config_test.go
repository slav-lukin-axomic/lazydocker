package presentation

import (
	"testing"

	"github.com/jesseduffield/lazydocker/pkg/domain"
)

func TestRenderContainerEnv(t *testing.T) {
	// PATH exercises a value containing "=" (SplitN N=2 keeps it intact) and
	// EMPTY exercises the empty-value branch.
	env := []string{"PATH=/usr/local/bin:/usr/bin", "HOME=/root", "EMPTY="}

	got, err := RenderContainerEnv(env)
	if err != nil {
		t.Fatalf("RenderContainerEnv returned error: %v", err)
	}
	assertGolden(t, "container_env", got)
}

func richInspect() domain.ContainerInspect {
	return domain.ContainerInspect{
		ID:      "abc123",
		Name:    "web",
		Image:   "nginx:latest",
		Command: []string{"nginx", "-g", "daemon off;"},
		Labels: map[string]string{
			"com.docker.compose.project": "myproject",
			"com.docker.compose.service": "web",
			"maintainer":                 "team@example.com",
		},
		Mounts: []domain.Mount{
			{Type: "volume", Name: "web-data"},
			{Type: "bind", Source: "/host/conf", Destination: "/etc/nginx/conf.d"},
		},
		Ports: []domain.PortBinding{
			// Multiple host bindings for one container port.
			{ContainerPort: "443/tcp", HostPorts: []string{"8443", "9443"}},
			// A present key with no host bindings proves the "has ports but no
			// binding lines" path (decided by len(Ports) > 0, not host-port count).
			{ContainerPort: "80/tcp", HostPorts: []string{}},
		},
	}
}

func fixedDetailsYAML() string {
	return "State:\n  Running: true\n  ExitCode: 0\nName: web\n"
}

func TestRenderContainerConfig(t *testing.T) {
	got := RenderContainerConfig(richInspect(), fixedDetailsYAML())
	assertGolden(t, "container_config", got)
}

func TestRenderContainerConfigNoMountsNoPorts(t *testing.T) {
	inspect := domain.ContainerInspect{
		ID:      "def456",
		Name:    "db",
		Image:   "postgres:16",
		Command: []string{"postgres"},
	}

	got := RenderContainerConfig(inspect, fixedDetailsYAML())
	assertGolden(t, "container_config_no_mounts_ports", got)
}
