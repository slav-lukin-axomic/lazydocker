package commands

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// DockerClient is the subset of the Docker SDK client that lazydocker's command
// layer actually uses. Depending on this interface rather than *client.Client
// lets command logic be exercised in tests with a fake, without a live daemon.
// The concrete *client.Client satisfies it (see the assertion below), so wiring
// is unchanged.
type DockerClient interface {
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
	ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	ContainerPause(ctx context.Context, containerID string) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerRestart(ctx context.Context, containerID string, options container.StopOptions) error
	ContainersPrune(ctx context.Context, pruneFilters filters.Args) (container.PruneReport, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error)
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerTop(ctx context.Context, containerID string, arguments []string) (container.TopResponse, error)
	ContainerUnpause(ctx context.Context, containerID string) error
	ImageHistory(ctx context.Context, imageID string, historyOpts ...client.ImageHistoryOption) ([]image.HistoryResponseItem, error)
	ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error)
	ImageRemove(ctx context.Context, imageID string, options image.RemoveOptions) ([]image.DeleteResponse, error)
	ImagesPrune(ctx context.Context, pruneFilters filters.Args) (image.PruneReport, error)
	NetworkList(ctx context.Context, options network.ListOptions) ([]network.Summary, error)
	NetworkRemove(ctx context.Context, networkID string) error
	NetworksPrune(ctx context.Context, pruneFilters filters.Args) (network.PruneReport, error)
	VolumeList(ctx context.Context, options volume.ListOptions) (volume.ListResponse, error)
	VolumeRemove(ctx context.Context, volumeID string, force bool) error
	VolumesPrune(ctx context.Context, pruneFilters filters.Args) (volume.PruneReport, error)
}

var _ DockerClient = (*client.Client)(nil)
