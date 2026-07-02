package presentation

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/fatih/color"
	"github.com/jesseduffield/lazydocker/pkg/commands"
	"github.com/stretchr/testify/assert"
)

var update = flag.Bool("update", false, "update golden files")

func TestMain(m *testing.M) {
	// Force colorization on so goldens capture the ANSI escape sequences
	// regardless of TTY. We want the golden to lock color *semantics*: Phase 2
	// replaces ANSI with semantic colors and these tests are the regression
	// guard for that swap.
	color.NoColor = false
	os.Exit(m.Run())
}

// assertGolden compares got against the committed golden file for name, or
// rewrites the golden when -update is passed.
func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("failed to write golden %s: %v", path, err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden %s (run with -update to create it): %v", path, err)
	}
	assert.Equal(t, string(want), got)
}

// --- fixture helpers ---

// makeContainer builds a *commands.Container from a bare container.Summary,
// leaving details unloaded (DetailsLoaded() == false).
func makeContainer(name string, summary container.Summary) *commands.Container {
	return &commands.Container{
		Name:      name,
		ID:        summary.ID,
		Container: summary,
	}
}

// withDetails attaches inspect details so DetailsLoaded() reports true.
func withDetails(c *commands.Container, state container.State) *commands.Container {
	c.Details = container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{State: &state},
	}
	return c
}

// withCPUStats appends a single recorded stat with the given CPU percentage.
func withCPUStats(c *commands.Container, cpuPercentage float64) *commands.Container {
	c.StatHistory = append(c.StatHistory, &commands.RecordedStats{
		DerivedStats: commands.DerivedStats{CPUPercentage: cpuPercentage},
	})
	return c
}

func runningSummary() container.Summary {
	return container.Summary{
		ID:    "runningid",
		State: "running",
		Image: "sha256:abc123def456",
		Ports: []container.Port{
			{PrivatePort: 80, PublicPort: 8080, Type: "tcp", IP: "0.0.0.0"},
			{PrivatePort: 443, Type: "tcp"},
		},
	}
}

func exitedSummary() container.Summary {
	return container.Summary{ID: "exitedid", State: "exited", Image: "myimage:latest"}
}

func historyItems() []image.HistoryResponseItem {
	return []image.HistoryResponseItem{
		{
			ID:        "sha256:0123456789abcdef",
			Tags:      []string{"myimage:latest"},
			Size:      1024 * 1024,
			CreatedBy: "/bin/sh -c #(nop) CMD [\"nginx\"]",
		},
		{
			ID:        "<missing>",
			Size:      0,
			CreatedBy: "RUN apt-get update\t&& apt-get install -y curl",
		},
	}
}
