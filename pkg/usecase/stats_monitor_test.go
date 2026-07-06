package usecase

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/stretchr/testify/assert"
)

// sample builds a RecordedStats recorded ago in the past with the given CPU%.
func sample(cpu float64, ago time.Duration) *domain.RecordedStats {
	return &domain.RecordedStats{
		DerivedStats: domain.DerivedStats{CPUPercentage: cpu},
		RecordedAt:   time.Now().Add(-ago),
	}
}

func TestStatsMonitorRecordAndRead(t *testing.T) {
	t.Parallel()

	t.Run("record accumulates and LastStats returns the newest", func(t *testing.T) {
		t.Parallel()
		m := NewStatsMonitor(&fakeDockerAPI{}, 0)

		m.record("c1", sample(10, 0))
		m.record("c1", sample(20, 0))

		last, ok := m.LastStats("c1")
		assert.True(t, ok)
		assert.Equal(t, 20.0, last.DerivedStats.CPUPercentage)
		assert.Len(t, m.History("c1"), 2)
	})

	t.Run("LastStats and History are empty for an unknown id", func(t *testing.T) {
		t.Parallel()
		m := NewStatsMonitor(&fakeDockerAPI{}, 0)

		_, ok := m.LastStats("missing")
		assert.False(t, ok)
		assert.Nil(t, m.History("missing"))
	})

	t.Run("zero maxDuration keeps all history", func(t *testing.T) {
		t.Parallel()
		m := NewStatsMonitor(&fakeDockerAPI{}, 0)

		m.record("c1", sample(1, time.Hour))
		m.record("c1", sample(2, time.Minute))
		m.record("c1", sample(3, 0))

		assert.Len(t, m.History("c1"), 3)
	})

	t.Run("maxDuration trims entries older than the window", func(t *testing.T) {
		t.Parallel()
		m := NewStatsMonitor(&fakeDockerAPI{}, time.Minute)

		// Two stale samples then two fresh ones: after each record the retention
		// logic keeps from the first entry still inside the window.
		m.record("c1", sample(1, 2*time.Hour))
		m.record("c1", sample(2, 90*time.Minute))
		m.record("c1", sample(3, 2*time.Second))
		m.record("c1", sample(4, time.Second))

		history := m.History("c1")
		assert.Len(t, history, 2)
		assert.Equal(t, 3.0, history[0].DerivedStats.CPUPercentage)
		assert.Equal(t, 4.0, history[1].DerivedStats.CPUPercentage)
	})
}

func TestStatsMonitorHistorySnapshot(t *testing.T) {
	t.Parallel()
	m := NewStatsMonitor(&fakeDockerAPI{}, 0)
	m.record("c1", sample(10, 0))

	snapshot := m.History("c1")
	// Recording again must not grow a snapshot the caller already holds.
	m.record("c1", sample(20, 0))

	assert.Len(t, snapshot, 1)
	assert.Len(t, m.History("c1"), 2)
}

func TestStatsMonitorPrune(t *testing.T) {
	t.Parallel()
	m := NewStatsMonitor(&fakeDockerAPI{}, 0)
	m.record("keep", sample(1, 0))
	m.record("drop", sample(2, 0))

	m.Prune([]string{"keep"})

	assert.Len(t, m.History("keep"), 1)
	assert.Nil(t, m.History("drop"))
}

func TestStatsMonitorEnsureMonitoringDedup(t *testing.T) {
	t.Parallel()

	var streams int32
	release := make(chan struct{})
	f := &fakeDockerAPI{
		streamStatsFn: func(ctx context.Context, id string, onSample func(*domain.RecordedStats)) error {
			atomic.AddInt32(&streams, 1)
			// Block until released so the stream is still "live" for the second call.
			select {
			case <-release:
			case <-ctx.Done():
			}
			return nil
		},
	}
	m := NewStatsMonitor(f, 0)

	m.EnsureMonitoring(context.Background(), "c1")
	// Second call while the first stream is still running must not spawn another.
	m.EnsureMonitoring(context.Background(), "c1")

	assert.Eventually(t, func() bool { return atomic.LoadInt32(&streams) == 1 }, time.Second, time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(&streams))

	close(release)

	// After the stream returns, the id is marked not-monitoring so a later call
	// restarts it.
	assert.Eventually(t, func() bool {
		m.mu.Lock()
		defer m.mu.Unlock()
		return !m.monitoring["c1"]
	}, time.Second, time.Millisecond)
}

// TestStatsMonitorConcurrentAccess exercises record/History/LastStats/Prune from
// many goroutines at once; run under -race it is the guard for the monitor's
// locking and for the snapshot copy in History.
func TestStatsMonitorConcurrentAccess(t *testing.T) {
	t.Parallel()
	m := NewStatsMonitor(&fakeDockerAPI{}, time.Minute)

	ids := []string{"a", "b", "c"}
	var wg sync.WaitGroup

	for _, id := range ids {
		id := id
		wg.Add(3)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				m.record(id, sample(float64(i), 0))
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				for _, s := range m.History(id) {
					_ = s.DerivedStats.CPUPercentage
				}
				_, _ = m.LastStats(id)
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				m.Prune(ids)
			}
		}()
	}

	wg.Wait()
}
