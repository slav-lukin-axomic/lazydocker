package commands

import (
	"time"

	"github.com/jesseduffield/lazydocker/pkg/domain"
)

func (c *Container) appendStats(stats *domain.RecordedStats, maxDuration time.Duration) {
	c.StatsMutex.Lock()
	defer c.StatsMutex.Unlock()

	c.StatHistory = append(c.StatHistory, stats)
	c.eraseOldHistory(maxDuration)
}

// eraseOldHistory removes any history before the user-specified max duration
func (c *Container) eraseOldHistory(maxDuration time.Duration) {
	if maxDuration == 0 {
		return
	}

	for i, stat := range c.StatHistory {
		if time.Since(stat.RecordedAt) < maxDuration {
			c.StatHistory = c.StatHistory[i:]
			return
		}
	}
}

func (c *Container) GetLastStats() (*domain.RecordedStats, bool) {
	c.StatsMutex.Lock()
	defer c.StatsMutex.Unlock()
	history := c.StatHistory
	if len(history) == 0 {
		return nil, false
	}
	return history[len(history)-1], true
}
