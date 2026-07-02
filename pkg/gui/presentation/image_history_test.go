package presentation

import (
	"testing"

	"github.com/docker/docker/api/types/image"
)

func TestRenderImageHistory(t *testing.T) {
	got, err := RenderImageHistory(historyItems())
	if err != nil {
		t.Fatalf("RenderImageHistory returned error: %v", err)
	}
	assertGolden(t, "image_history", got)
}

func TestRenderImageHistoryEmpty(t *testing.T) {
	got, err := RenderImageHistory([]image.HistoryResponseItem{})
	if err != nil {
		t.Fatalf("RenderImageHistory returned error: %v", err)
	}
	assertGolden(t, "image_history_empty", got)
}
