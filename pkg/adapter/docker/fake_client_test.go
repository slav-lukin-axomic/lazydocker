package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

// fakeAPIClient is a test double for apiClient. Each method delegates to an
// optional function field so a test wires up only the calls it exercises; unset
// fields return zero values. Mirrors the repo's testify-and-fakes convention
// (see pkg/commands/docker_client_fake_test.go).
type fakeAPIClient struct {
	containerListFn    func(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	containerInspectFn func(ctx context.Context, containerID string) (container.InspectResponse, error)
	containerStartFn   func(ctx context.Context, containerID string, options container.StartOptions) error
	containerStopFn    func(ctx context.Context, containerID string, options container.StopOptions) error
	containerRestartFn func(ctx context.Context, containerID string, options container.StopOptions) error
	containerPauseFn   func(ctx context.Context, containerID string) error
	containerUnpauseFn func(ctx context.Context, containerID string) error
	containerRemoveFn  func(ctx context.Context, containerID string, options container.RemoveOptions) error
	containerTopFn     func(ctx context.Context, containerID string, arguments []string) (container.TopResponse, error)
	containersPruneFn  func(ctx context.Context, pruneFilters filters.Args) (container.PruneReport, error)
	containerStatsFn   func(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error)
	containerLogsFn    func(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error)
	networkListFn      func(ctx context.Context, options network.ListOptions) ([]network.Summary, error)
	networkRemoveFn    func(ctx context.Context, networkID string) error
	networksPruneFn    func(ctx context.Context, pruneFilters filters.Args) (network.PruneReport, error)
	volumeListFn       func(ctx context.Context, options volume.ListOptions) (volume.ListResponse, error)
	volumeRemoveFn     func(ctx context.Context, volumeID string, force bool) error
	volumesPruneFn     func(ctx context.Context, pruneFilters filters.Args) (volume.PruneReport, error)
	imageListFn        func(ctx context.Context, options image.ListOptions) ([]image.Summary, error)
	imageHistoryFn     func(ctx context.Context, imageID string, historyOpts ...client.ImageHistoryOption) ([]image.HistoryResponseItem, error)
	imageRemoveFn      func(ctx context.Context, imageID string, options image.RemoveOptions) ([]image.DeleteResponse, error)
	imagesPruneFn      func(ctx context.Context, pruneFilters filters.Args) (image.PruneReport, error)
}

var _ apiClient = (*fakeAPIClient)(nil)

func (f *fakeAPIClient) ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
	if f.containerListFn != nil {
		return f.containerListFn(ctx, options)
	}
	return nil, nil
}

func (f *fakeAPIClient) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	if f.containerInspectFn != nil {
		return f.containerInspectFn(ctx, containerID)
	}
	return container.InspectResponse{}, nil
}

func (f *fakeAPIClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	if f.containerStartFn != nil {
		return f.containerStartFn(ctx, containerID, options)
	}
	return nil
}

func (f *fakeAPIClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	if f.containerStopFn != nil {
		return f.containerStopFn(ctx, containerID, options)
	}
	return nil
}

func (f *fakeAPIClient) ContainerRestart(ctx context.Context, containerID string, options container.StopOptions) error {
	if f.containerRestartFn != nil {
		return f.containerRestartFn(ctx, containerID, options)
	}
	return nil
}

func (f *fakeAPIClient) ContainerPause(ctx context.Context, containerID string) error {
	if f.containerPauseFn != nil {
		return f.containerPauseFn(ctx, containerID)
	}
	return nil
}

func (f *fakeAPIClient) ContainerUnpause(ctx context.Context, containerID string) error {
	if f.containerUnpauseFn != nil {
		return f.containerUnpauseFn(ctx, containerID)
	}
	return nil
}

func (f *fakeAPIClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	if f.containerRemoveFn != nil {
		return f.containerRemoveFn(ctx, containerID, options)
	}
	return nil
}

func (f *fakeAPIClient) ContainerTop(ctx context.Context, containerID string, arguments []string) (container.TopResponse, error) {
	if f.containerTopFn != nil {
		return f.containerTopFn(ctx, containerID, arguments)
	}
	return container.TopResponse{}, nil
}

func (f *fakeAPIClient) ContainersPrune(ctx context.Context, pruneFilters filters.Args) (container.PruneReport, error) {
	if f.containersPruneFn != nil {
		return f.containersPruneFn(ctx, pruneFilters)
	}
	return container.PruneReport{}, nil
}

func (f *fakeAPIClient) ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error) {
	if f.containerStatsFn != nil {
		return f.containerStatsFn(ctx, containerID, stream)
	}
	return container.StatsResponseReader{}, nil
}

func (f *fakeAPIClient) ContainerLogs(ctx context.Context, containerID string, options container.LogsOptions) (io.ReadCloser, error) {
	if f.containerLogsFn != nil {
		return f.containerLogsFn(ctx, containerID, options)
	}
	return nil, nil
}

func (f *fakeAPIClient) NetworkList(ctx context.Context, options network.ListOptions) ([]network.Summary, error) {
	if f.networkListFn != nil {
		return f.networkListFn(ctx, options)
	}
	return nil, nil
}

func (f *fakeAPIClient) NetworkRemove(ctx context.Context, networkID string) error {
	if f.networkRemoveFn != nil {
		return f.networkRemoveFn(ctx, networkID)
	}
	return nil
}

func (f *fakeAPIClient) NetworksPrune(ctx context.Context, pruneFilters filters.Args) (network.PruneReport, error) {
	if f.networksPruneFn != nil {
		return f.networksPruneFn(ctx, pruneFilters)
	}
	return network.PruneReport{}, nil
}

func (f *fakeAPIClient) VolumeList(ctx context.Context, options volume.ListOptions) (volume.ListResponse, error) {
	if f.volumeListFn != nil {
		return f.volumeListFn(ctx, options)
	}
	return volume.ListResponse{}, nil
}

func (f *fakeAPIClient) VolumeRemove(ctx context.Context, volumeID string, force bool) error {
	if f.volumeRemoveFn != nil {
		return f.volumeRemoveFn(ctx, volumeID, force)
	}
	return nil
}

func (f *fakeAPIClient) VolumesPrune(ctx context.Context, pruneFilters filters.Args) (volume.PruneReport, error) {
	if f.volumesPruneFn != nil {
		return f.volumesPruneFn(ctx, pruneFilters)
	}
	return volume.PruneReport{}, nil
}

func (f *fakeAPIClient) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	if f.imageListFn != nil {
		return f.imageListFn(ctx, options)
	}
	return nil, nil
}

func (f *fakeAPIClient) ImageHistory(ctx context.Context, imageID string, historyOpts ...client.ImageHistoryOption) ([]image.HistoryResponseItem, error) {
	if f.imageHistoryFn != nil {
		return f.imageHistoryFn(ctx, imageID, historyOpts...)
	}
	return nil, nil
}

func (f *fakeAPIClient) ImageRemove(ctx context.Context, imageID string, options image.RemoveOptions) ([]image.DeleteResponse, error) {
	if f.imageRemoveFn != nil {
		return f.imageRemoveFn(ctx, imageID, options)
	}
	return nil, nil
}

func (f *fakeAPIClient) ImagesPrune(ctx context.Context, pruneFilters filters.Args) (image.PruneReport, error) {
	if f.imagesPruneFn != nil {
		return f.imagesPruneFn(ctx, pruneFilters)
	}
	return image.PruneReport{}, nil
}
