package presentation

import (
	"regexp"
	"testing"
	"time"

	"github.com/jesseduffield/lazydocker/pkg/config"
	"github.com/jesseduffield/lazydocker/pkg/domain"
)

// elapsedCaption matches the "(1s)" style elapsed-duration suffix that plotGraph
// derives from time.Since(RecordedAt). recordedAt matches the RFC3339 timestamp
// the stats YAML dump prints for each RecordedAt. Both are wall-clock dependent,
// so we normalize them before golden comparison; every other byte (the plotted
// graph, axis labels, traffic/PID lines, colored YAML) is locked exactly.
var (
	elapsedCaption = regexp.MustCompile(` \(\d+[a-z]+\)`)
	recordedAt     = regexp.MustCompile(`\d{4}-\d{2}-\d{2}T[\d:.]+[+-]\d{2}:\d{2}`)
)

func TestRenderStats(t *testing.T) {
	userConfig := config.GetDefaultConfig()

	base := time.Now().Add(-30 * time.Second)
	// Fixed-shape history so the plotted graph is deterministic.
	var history []*domain.RecordedStats
	for i, cpu := range []float64{10, 40, 25} {
		stats := domain.ContainerStats{}
		stats.PidsStats.Current = 7
		stats.Networks.Eth0.RxBytes = 2048
		stats.Networks.Eth0.TxBytes = 4096
		history = append(history, &domain.RecordedStats{
			ClientStats: stats,
			DerivedStats: domain.DerivedStats{
				CPUPercentage:    cpu,
				MemoryPercentage: cpu / 2,
			},
			RecordedAt: base.Add(time.Duration(i) * time.Second),
		})
	}

	got, err := RenderStats(&userConfig, history, 80)
	if err != nil {
		t.Fatalf("RenderStats returned error: %v", err)
	}

	got = elapsedCaption.ReplaceAllString(got, " (ELAPSED)")
	got = recordedAt.ReplaceAllString(got, "RECORDED_AT")
	assertGolden(t, "container_stats", got)
}

func TestRenderStatsNoHistory(t *testing.T) {
	userConfig := config.GetDefaultConfig()

	got, err := RenderStats(&userConfig, nil, 80)
	if err != nil {
		t.Fatalf("RenderStats returned error: %v", err)
	}
	assertGolden(t, "container_stats_empty", got)
}
