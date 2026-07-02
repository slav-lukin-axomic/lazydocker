package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusRoundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		status Status
		str    string
	}{
		{"running", StatusRunning, "running"},
		{"exited", StatusExited, "exited"},
		{"paused", StatusPaused, "paused"},
		{"created", StatusCreated, "created"},
		{"restarting", StatusRestarting, "restarting"},
		{"removing", StatusRemoving, "removing"},
		{"dead", StatusDead, "dead"},
		{"unknown", StatusUnknown, "unknown"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.str, tc.status.String())
			if tc.status != StatusUnknown {
				assert.Equal(t, tc.status, ParseStatus(tc.str))
			}
		})
	}
}

func TestParseStatusUnrecognised(t *testing.T) {
	t.Parallel()

	for _, s := range []string{"", "bogus", "Running", "UP"} {
		assert.Equal(t, StatusUnknown, ParseStatus(s), "input %q should be StatusUnknown", s)
	}
}

func TestHealthRoundTrip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		health Health
		str    string
	}{
		{"healthy", HealthHealthy, "healthy"},
		{"unhealthy", HealthUnhealthy, "unhealthy"},
		{"starting", HealthStarting, "starting"},
		{"none", HealthNone, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.str, tc.health.String())
			if tc.health != HealthNone {
				assert.Equal(t, tc.health, ParseHealth(tc.str))
			}
		})
	}
}

func TestParseHealthNoneAndUnrecognised(t *testing.T) {
	t.Parallel()

	// The empty string, the SDK's "none", and anything unrecognised all map to
	// HealthNone.
	for _, s := range []string{"", "none", "bogus", "Healthy"} {
		assert.Equal(t, HealthNone, ParseHealth(s), "input %q should be HealthNone", s)
	}
}
