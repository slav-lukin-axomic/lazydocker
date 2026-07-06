package presentation

import (
	"testing"

	"github.com/jesseduffield/lazydocker/pkg/domain"
)

func TestRenderImageHistory(t *testing.T) {
	got, err := RenderImageHistory(historyItems())
	if err != nil {
		t.Fatalf("RenderImageHistory returned error: %v", err)
	}
	assertGolden(t, "image_history", got)
}

func TestRenderImageHistoryEmpty(t *testing.T) {
	got, err := RenderImageHistory([]domain.HistoryLayer{})
	if err != nil {
		t.Fatalf("RenderImageHistory returned error: %v", err)
	}
	assertGolden(t, "image_history_empty", got)
}
