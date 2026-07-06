package domain

import (
	"context"
	"errors"
	"io"
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
// This slice covers containers, networks, volumes, and images — the port is
// complete for the migrated simple-resource panels (see
// docs/tui-migration-phase1-design.md §4 and §7).
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
	// StreamLogs streams a container's logs to out until the stream ends or ctx is
	// cancelled. The adapter owns TTY detection and stdout/stderr de-multiplexing,
	// so callers receive already-demuxed bytes and need no SDK knowledge. It blocks.
	StreamLogs(ctx context.Context, id string, opts LogOptions, out io.Writer) error

	// ListNetworks returns all networks in the order the Engine reports them (the
	// panel applies its own sort).
	ListNetworks(ctx context.Context) ([]Network, error)
	// RemoveNetwork removes the network with the given name (the Engine accepts the
	// name as the id).
	RemoveNetwork(ctx context.Context, name string) error
	// PruneNetworks removes all unused networks.
	PruneNetworks(ctx context.Context) error

	// ListVolumes returns all volumes in the order the Engine reports them (the
	// panel applies its own sort).
	ListVolumes(ctx context.Context) ([]Volume, error)
	// RemoveVolume removes the volume with the given name; force removes it even
	// when in use.
	RemoveVolume(ctx context.Context, name string, force bool) error
	// PruneVolumes removes all unused volumes.
	PruneVolumes(ctx context.Context) error

	// ListImages returns all images in the order the Engine reports them (the
	// panel applies its own sort). Name/Tag are left zero for the caller to derive
	// from RepoTags.
	ListImages(ctx context.Context) ([]Image, error)
	// ImageHistory returns the build-layer history of the image with the given id.
	ImageHistory(ctx context.Context, id string) ([]HistoryLayer, error)
	// RemoveImage removes the image with the given id; force removes it even when
	// referenced, and pruneChildren removes untagged parents.
	RemoveImage(ctx context.Context, id string, force, pruneChildren bool) error
	// PruneImages removes all dangling images.
	PruneImages(ctx context.Context) error
}
