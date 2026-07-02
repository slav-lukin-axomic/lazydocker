package presentation

import (
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/jesseduffield/lazydocker/pkg/commands"
	"github.com/jesseduffield/lazydocker/pkg/config"
)

// colSep joins the display columns for golden capture. Using the ASCII unit
// separator keeps the golden readable while never colliding with rendered text.
const colSep = "\x1f"

func healthStyles() []string {
	return []string{"long", "short", "icon"}
}

func TestGetContainerDisplayStrings(t *testing.T) {
	healthy := func(status string) container.State {
		return container.State{Health: &container.Health{Status: status}}
	}

	cases := []struct {
		name      string
		container *commands.Container
	}{
		{
			name:      "running_no_details",
			container: makeContainer("web", runningSummary()),
		},
		{
			name:      "running_healthy",
			container: withDetails(makeContainer("web", runningSummary()), healthy("healthy")),
		},
		{
			name:      "running_unhealthy",
			container: withDetails(makeContainer("web", runningSummary()), healthy("unhealthy")),
		},
		{
			name:      "running_starting",
			container: withDetails(makeContainer("web", runningSummary()), healthy("starting")),
		},
		{
			name:      "running_no_healthcheck",
			container: withDetails(makeContainer("web", runningSummary()), container.State{}),
		},
		{
			name:      "running_with_cpu",
			container: withCPUStats(withDetails(makeContainer("web", runningSummary()), healthy("healthy")), 12.5),
		},
		{
			name:      "running_high_cpu",
			container: withCPUStats(withDetails(makeContainer("web", runningSummary()), healthy("healthy")), 95.0),
		},
		{
			name:      "running_mid_cpu",
			container: withCPUStats(withDetails(makeContainer("web", runningSummary()), healthy("healthy")), 70.0),
		},
		{
			name:      "exited_zero_no_details",
			container: makeContainer("job", exitedSummary()),
		},
		{
			name:      "exited_zero",
			container: withDetails(makeContainer("job", exitedSummary()), container.State{ExitCode: 0}),
		},
		{
			name:      "exited_nonzero",
			container: withDetails(makeContainer("job", exitedSummary()), container.State{ExitCode: 137}),
		},
		{
			name:      "paused",
			container: makeContainer("db", container.Summary{ID: "pausedid", State: "paused", Image: "postgres:15"}),
		},
		{
			name:      "created",
			container: makeContainer("new", container.Summary{ID: "createdid", State: "created", Image: "alpine"}),
		},
	}

	for _, tc := range cases {
		for _, style := range healthStyles() {
			t.Run(tc.name+"_"+style, func(t *testing.T) {
				guiConfig := &config.GuiConfig{ContainerStatusHealthStyle: style}
				got := strings.Join(GetContainerDisplayStrings(guiConfig, ContainerToDomain(tc.container)), colSep)
				assertGolden(t, "containers_"+tc.name+"_"+style, got)
			})
		}
	}
}
