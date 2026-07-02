// Package domain is the framework-free core of the hexagon: domain models,
// value objects, and the driven-port interfaces the use cases depend on. It must
// not import the Docker SDK, any TUI framework (gocui/tcell), fatih/color, or
// pkg/gui — the adapters own those concerns and translate to these types. A
// depguard rule enforces this.
package domain

// Status is a container's lifecycle state. Its String form round-trips the exact
// lowercase strings the Docker SDK reports (see container.ContainerState), so
// presentation code can switch on Status.String() and behave identically to the
// pre-migration code that switched on the raw SDK state string.
type Status int

const (
	// StatusUnknown is the zero value and the fallback for any unrecognised SDK
	// state string; its String form ("unknown") deliberately matches none of the
	// presentation switch cases.
	StatusUnknown Status = iota
	StatusRunning
	StatusExited
	StatusPaused
	StatusCreated
	StatusRestarting
	StatusRemoving
	StatusDead
)

// String returns the lowercase SDK state string for the status, or "unknown".
func (s Status) String() string {
	switch s {
	case StatusRunning:
		return "running"
	case StatusExited:
		return "exited"
	case StatusPaused:
		return "paused"
	case StatusCreated:
		return "created"
	case StatusRestarting:
		return "restarting"
	case StatusRemoving:
		return "removing"
	case StatusDead:
		return "dead"
	case StatusUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// ParseStatus maps a Docker SDK state string to a Status, returning StatusUnknown
// for anything unrecognised.
func ParseStatus(s string) Status {
	switch s {
	case "running":
		return StatusRunning
	case "exited":
		return StatusExited
	case "paused":
		return StatusPaused
	case "created":
		return StatusCreated
	case "restarting":
		return StatusRestarting
	case "removing":
		return StatusRemoving
	case "dead":
		return StatusDead
	default:
		return StatusUnknown
	}
}

// Health is a container's healthcheck state.
type Health int

const (
	// HealthNone means the container has no healthcheck (or none reported yet).
	HealthNone Health = iota
	HealthHealthy
	HealthUnhealthy
	HealthStarting
)

// String returns the lowercase SDK health string for the health, or "" for
// HealthNone (matching the pre-migration behaviour where a nil/absent health had
// no string).
func (h Health) String() string {
	switch h {
	case HealthHealthy:
		return "healthy"
	case HealthUnhealthy:
		return "unhealthy"
	case HealthStarting:
		return "starting"
	case HealthNone:
		return ""
	default:
		return ""
	}
}

// ParseHealth maps a Docker SDK health status string to a Health. The empty
// string (and any unrecognised value, including the SDK's "none") maps to
// HealthNone.
func ParseHealth(s string) Health {
	switch s {
	case "healthy":
		return HealthHealthy
	case "unhealthy":
		return HealthUnhealthy
	case "starting":
		return HealthStarting
	default:
		return HealthNone
	}
}

// Port is a container port mapping. It mirrors the fields presentation reads from
// the SDK's container.Port (Proto is the SDK's Type field).
type Port struct {
	IP          string
	PublicPort  uint16
	PrivatePort uint16
	Proto       string
}

// DerivedStats holds the stats lazydocker computes from raw Docker stats samples.
// It mirrors the fields presentation reads today (see the pre-migration
// commands.DerivedStats).
type DerivedStats struct {
	CPUPercentage    float64
	MemoryPercentage float64
}

// RemoveOptions are the domain-level options for removing a container, decoupled
// from the SDK's container.RemoveOptions.
type RemoveOptions struct {
	Force         bool
	RemoveVolumes bool
}
