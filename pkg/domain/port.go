package domain

import (
	"context"
	"errors"
)

// ErrContainerRunning is returned by RemoveContainer when Docker refuses to
// remove a running container without force. Callers branch on it with errors.Is
// to surface the "stop the container first" affordance (the pre-migration
// MustStopContainer error code).
var ErrContainerRunning = errors.New("container must be stopped before removal")

// TopResult is the process listing for a container: column titles and one row of
// values per process, matching what the pre-migration Container.RenderTop/Top
// produced from the SDK's container.TopResponse.
type TopResult struct {
	Titles    []string
	Processes [][]string
}

// DockerAPI is the driven port for request/response Docker Engine operations. It
// is consumer-defined here in the core and implemented by the docker adapter,
// which owns the SDK↔domain mapping.
//
// This is the container slice of the port. It intentionally covers containers
// only for now; image, volume, and network methods are added in later migration
// slices (see docs/tui-migration-phase1-design.md §4 and §7).
type DockerAPI interface {
	// ListContainers returns all containers with Details left nil (inspect
	// populates details separately).
	ListContainers(ctx context.Context) ([]Container, error)
	InspectContainer(ctx context.Context, id string) (ContainerDetails, error)
	// InspectContainerVerbose inspects a container and returns the framework-free
	// projection the Config/Env detail views render, plus the raw full-inspect
	// data marshalled to a YAML string (an opaque display blob). Identity fields
	// (ID, Name) are intentionally left zero on the projection: they are the
	// summary-derived display identity, which the caller supplies from the store.
	InspectContainerVerbose(ctx context.Context, id string) (ContainerInspect, string, error)
	StartContainer(ctx context.Context, id string) error
	StopContainer(ctx context.Context, id string) error
	RestartContainer(ctx context.Context, id string) error
	PauseContainer(ctx context.Context, id string) error
	UnpauseContainer(ctx context.Context, id string) error
	// RemoveContainer returns ErrContainerRunning (wrapped) when Docker refuses to
	// remove a running container without force.
	RemoveContainer(ctx context.Context, id string, opts RemoveOptions) error
	ContainerTop(ctx context.Context, id string) (TopResult, error)
	PruneContainers(ctx context.Context) error
	// StreamStats streams recorded stats samples for a container until the stream
	// ends or ctx is cancelled, invoking onSample for each. It blocks.
	StreamStats(ctx context.Context, id string, onSample func(*RecordedStats)) error
}
