package presentation

import (
	"testing"

	"github.com/jesseduffield/lazydocker/pkg/domain"
)

func topResult() domain.TopResult {
	return domain.TopResult{
		Titles: []string{"PID", "USER", "TIME", "COMMAND"},
		Processes: [][]string{
			{"1", "root", "0:00", "nginx: master process nginx -g daemon off;"},
			{"31", "nginx", "0:00", "nginx: worker process"},
		},
	}
}

func TestRenderContainerTop(t *testing.T) {
	got, err := RenderContainerTop(topResult())
	if err != nil {
		t.Fatalf("RenderContainerTop returned error: %v", err)
	}
	assertGolden(t, "container_top", got)
}

// TestRenderContainerTopNoProcesses locks the "empty" case that actually reaches
// the formatter: titles present, zero process rows. A truly empty TopResult (nil
// Titles) is never produced for a running container, and feeding one to
// utils.RenderTable panics — a pre-existing quirk of RenderTable, out of scope here.
func TestRenderContainerTopNoProcesses(t *testing.T) {
	got, err := RenderContainerTop(domain.TopResult{
		Titles: []string{"PID", "USER", "TIME", "COMMAND"},
	})
	if err != nil {
		t.Fatalf("RenderContainerTop returned error: %v", err)
	}
	assertGolden(t, "container_top_no_processes", got)
}
