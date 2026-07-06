package presentation

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/jesseduffield/lazydocker/pkg/commands"
	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func TestContainerToDomain(t *testing.T) {
	t.Run("nil input maps to nil", func(t *testing.T) {
		assert.Nil(t, ContainerToDomain(nil))
	})

	t.Run("details and ports map field-by-field", func(t *testing.T) {
		c := &commands.Container{
			ID:              "abc123",
			Name:            "web",
			ServiceName:     "web-svc",
			ContainerNumber: "1",
			ProjectName:     "proj",
			OneOff:          true,
			Container: container.Summary{
				Image: "sha256:deadbeef",
				State: "running",
				Labels: map[string]string{
					"com.docker.compose.service": "web-svc",
				},
				Ports: []container.Port{
					{IP: "0.0.0.0", PublicPort: 8080, PrivatePort: 80, Type: "tcp"},
					{PrivatePort: 443, Type: "tcp"},
				},
			},
			Details: container.InspectResponse{
				ContainerJSONBase: &container.ContainerJSONBase{
					State: &container.State{
						Running:  true,
						ExitCode: 0,
						Health:   &container.Health{Status: "healthy"},
					},
				},
				Config: &container.Config{OpenStdin: true},
			},
		}

		got := ContainerToDomain(c)

		// The bridge no longer populates Stats — history lives in the
		// StatsMonitor and callers set domain.Container.Stats from it — so Stats
		// stays nil here.
		want := &domain.Container{
			ID:              "abc123",
			Name:            "web",
			ServiceName:     "web-svc",
			ContainerNumber: "1",
			ProjectName:     "proj",
			OneOff:          true,
			Image:           "sha256:deadbeef",
			Status:          domain.StatusRunning,
			Ports: []domain.Port{
				{IP: "0.0.0.0", PublicPort: 8080, PrivatePort: 80, Proto: "tcp"},
				{PrivatePort: 443, Proto: "tcp"},
			},
			Labels: map[string]string{
				"com.docker.compose.service": "web-svc",
			},
			Details: &domain.ContainerDetails{
				Running:   true,
				ExitCode:  0,
				Health:    domain.HealthHealthy,
				OpenStdin: true,
			},
		}

		assert.Equal(t, want, got)
	})

	t.Run("no details leaves Details nil", func(t *testing.T) {
		c := &commands.Container{
			ID:   "noinspect",
			Name: "db",
			Container: container.Summary{
				Image: "postgres:15",
				State: "created",
			},
		}

		got := ContainerToDomain(c)

		assert.Nil(t, got.Details)
		assert.Nil(t, got.Stats)
		assert.Equal(t, domain.StatusCreated, got.Status)
	})

	t.Run("details without healthcheck maps to HealthNone", func(t *testing.T) {
		c := &commands.Container{
			ID: "nohealth",
			Container: container.Summary{
				State: "running",
			},
			Details: container.InspectResponse{
				ContainerJSONBase: &container.ContainerJSONBase{
					State: &container.State{Running: true},
				},
			},
		}

		got := ContainerToDomain(c)

		assert.NotNil(t, got.Details)
		assert.Equal(t, domain.HealthNone, got.Details.Health)
	})
}
