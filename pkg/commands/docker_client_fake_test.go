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

// fakeDockerClient is a test double for DockerClient. Each method delegates to
// an optional function field so a test can wire up only the calls it exercises;
// unset fields return zero values. Prefer this over a mock framework per the
// repo's testify-and-fakes convention.
type fakeDockerClient struct {
	containerInspectFn func(ctx context.Context, containerID string) (container.InspectResponse, error)
	containerRemoveFn  func(ctx context.Context, containerID string, options container.RemoveOptions) error
	containerTopFn     func(ctx context.Context, containerID string, arguments []string) (container.TopResponse, error)
}

var _ DockerClient = (*fakeDockerClient)(nil)

func (f *fakeDockerClient) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	if f.containerInspectFn != nil {
		return f.containerInspectFn(ctx, containerID)
	}
	return container.InspectResponse{}, nil
}

func (f *fakeDockerClient) ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
	return nil, nil
}

func (f *fakeDockerClient) ContainerPause(ctx context.Context, containerID string) error {
	return nil
}

func (f *fakeDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	if f.containerRemoveFn != nil {
		return f.containerRemoveFn(ctx, containerID, options)
	}
	return nil
}

func (f *fakeDockerClient) ContainerRestart(ctx context.Context, containerID string, options container.StopOptions) error {
	return nil
}

func (f *fakeDockerClient) ContainersPrune(ctx context.Context, pruneFilters filters.Args) (container.PruneReport, error) {
	return container.PruneReport{}, nil
}

func (f *fakeDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	return nil
}

func (f *fakeDockerClient) ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error) {
	return container.StatsResponseReader{}, nil
}

func (f *fakeDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	return nil
}

func (f *fakeDockerClient) ContainerTop(ctx context.Context, containerID string, arguments []string) (container.TopResponse, error) {
	if f.containerTopFn != nil {
		return f.containerTopFn(ctx, containerID, arguments)
	}
	return container.TopResponse{}, nil
}

func (f *fakeDockerClient) ContainerUnpause(ctx context.Context, containerID string) error {
	return nil
}

func (f *fakeDockerClient) ImageHistory(ctx context.Context, imageID string, historyOpts ...client.ImageHistoryOption) ([]image.HistoryResponseItem, error) {
	return nil, nil
}

func (f *fakeDockerClient) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	return nil, nil
}

func (f *fakeDockerClient) ImageRemove(ctx context.Context, imageID string, options image.RemoveOptions) ([]image.DeleteResponse, error) {
	return nil, nil
}

func (f *fakeDockerClient) ImagesPrune(ctx context.Context, pruneFilters filters.Args) (image.PruneReport, error) {
	return image.PruneReport{}, nil
}

func (f *fakeDockerClient) NetworkList(ctx context.Context, options network.ListOptions) ([]network.Summary, error) {
	return nil, nil
}

func (f *fakeDockerClient) NetworkRemove(ctx context.Context, networkID string) error {
	return nil
}

func (f *fakeDockerClient) NetworksPrune(ctx context.Context, pruneFilters filters.Args) (network.PruneReport, error) {
	return network.PruneReport{}, nil
}

func (f *fakeDockerClient) VolumeList(ctx context.Context, options volume.ListOptions) (volume.ListResponse, error) {
	return volume.ListResponse{}, nil
}

func (f *fakeDockerClient) VolumeRemove(ctx context.Context, volumeID string, force bool) error {
	return nil
}

func (f *fakeDockerClient) VolumesPrune(ctx context.Context, pruneFilters filters.Args) (volume.PruneReport, error) {
	return volume.PruneReport{}, nil
}
