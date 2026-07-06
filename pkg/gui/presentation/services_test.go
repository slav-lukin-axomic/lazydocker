package presentation

import (
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/jesseduffield/lazydocker/pkg/commands"
	"github.com/jesseduffield/lazydocker/pkg/config"
	"github.com/jesseduffield/lazydocker/pkg/domain"
)

func TestGetServiceDisplayStrings(t *testing.T) {
	cases := []struct {
		name    string
		service *commands.Service
		stats   *domain.DerivedStats
	}{
		{
			name:    "no_container",
			service: &commands.Service{Name: "worker"},
		},
		{
			name: "running_container",
			service: &commands.Service{
				Name: "web",
				Container: withDetails(
					makeContainer("web", runningSummary()),
					container.State{Health: &container.Health{Status: "healthy"}},
				),
			},
		},
		{
			name: "running_container_with_cpu",
			service: &commands.Service{
				Name: "web",
				Container: withDetails(
					makeContainer("web", runningSummary()),
					container.State{Health: &container.Health{Status: "healthy"}},
				),
			},
			stats: &domain.DerivedStats{CPUPercentage: 42.0},
		},
		{
			name: "exited_container",
			service: &commands.Service{
				Name:      "job",
				Container: withDetails(makeContainer("job", exitedSummary()), container.State{ExitCode: 1}),
			},
		},
	}

	for _, tc := range cases {
		for _, style := range healthStyles() {
			t.Run(tc.name+"_"+style, func(t *testing.T) {
				guiConfig := &config.GuiConfig{ContainerStatusHealthStyle: style}
				got := strings.Join(GetServiceDisplayStrings(guiConfig, tc.service, tc.stats), colSep)
				assertGolden(t, "services_"+tc.name+"_"+style, got)
			})
		}
	}
}
