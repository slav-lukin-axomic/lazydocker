package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/jesseduffield/lazydocker/pkg/domain"
)

// StatsMonitor owns per-container stats history and monitoring dedup. It replaces
// the stats state that used to live on commands.Container (StatHistory,
// MonitoringStats, StatsMutex), keying everything by container ID so the store can
// later hold framework-free domain containers without carrying live stats state.
type StatsMonitor struct {
	api         domain.DockerAPI
	maxDuration time.Duration

	mu         sync.Mutex
	history    map[string][]*domain.RecordedStats
	monitoring map[string]bool
}

// NewStatsMonitor returns a StatsMonitor that streams from the given port and
// trims history older than maxDuration (a zero maxDuration keeps all history).
func NewStatsMonitor(api domain.DockerAPI, maxDuration time.Duration) *StatsMonitor {
	return &StatsMonitor{
		api:         api,
		maxDuration: maxDuration,
		history:     make(map[string][]*domain.RecordedStats),
		monitoring:  make(map[string]bool),
	}
}

// EnsureMonitoring starts a single streaming goroutine for id if one is not
// already running. When the stream returns (ends or ctx cancelled), id is marked
// not-monitoring so a later call can restart it — preserving the pre-migration
// "one goroutine per container, restart when the stream dies" behaviour.
func (m *StatsMonitor) EnsureMonitoring(ctx context.Context, id string) {
	m.mu.Lock()
	if m.monitoring[id] {
		m.mu.Unlock()
		return
	}
	m.monitoring[id] = true
	m.mu.Unlock()

	go func() {
		// The stream error is intentionally dropped: a disconnected daemon already
		// surfaces an error panel elsewhere, matching the pre-migration monitor.
		_ = m.api.StreamStats(ctx, id, func(rs *domain.RecordedStats) {
			m.record(id, rs)
		})

		m.mu.Lock()
		m.monitoring[id] = false
		m.mu.Unlock()
	}()
}

// record appends a sample for id and trims history older than maxDuration, using
// the same retention logic as the pre-migration eraseOldHistory.
func (m *StatsMonitor) record(id string, rs *domain.RecordedStats) {
	m.mu.Lock()
	defer m.mu.Unlock()

	history := append(m.history[id], rs)

	if m.maxDuration != 0 {
		for i, stat := range history {
			if time.Since(stat.RecordedAt) < m.maxDuration {
				history = history[i:]
				break
			}
		}
	}

	m.history[id] = history
}

// History returns a snapshot copy of id's recorded stats, so callers can render
// without holding the lock and without racing concurrent appends.
func (m *StatsMonitor) History(id string) []*domain.RecordedStats {
	m.mu.Lock()
	defer m.mu.Unlock()

	history := m.history[id]
	if len(history) == 0 {
		return nil
	}
	snapshot := make([]*domain.RecordedStats, len(history))
	copy(snapshot, history)
	return snapshot
}

// LastStats returns id's most recent recorded sample, matching the pre-migration
// Container.GetLastStats.
func (m *StatsMonitor) LastStats(id string) (*domain.RecordedStats, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	history := m.history[id]
	if len(history) == 0 {
		return nil, false
	}
	return history[len(history)-1], true
}

// Prune drops history and monitoring flags for any id not in activeIDs, bounding
// map growth as containers come and go (the pre-migration design GC'd this state
// with the container object).
func (m *StatsMonitor) Prune(activeIDs []string) {
	active := make(map[string]bool, len(activeIDs))
	for _, id := range activeIDs {
		active[id] = true
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for id := range m.history {
		if !active[id] {
			delete(m.history, id)
		}
	}
	for id := range m.monitoring {
		if !active[id] {
			delete(m.monitoring, id)
		}
	}
}
