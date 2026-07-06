package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
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
