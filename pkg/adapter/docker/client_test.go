package docker

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func TestListContainersMapsSummaries(t *testing.T) {
	t.Parallel()

	var gotOpts container.ListOptions
	fake := &fakeAPIClient{
		containerListFn: func(_ context.Context, options container.ListOptions) ([]container.Summary, error) {
			gotOpts = options
			return []container.Summary{composeSummary()}, nil
		},
	}

	got, err := NewAdapter(fake).ListContainers(context.Background())
	assert.NoError(t, err)
	assert.True(t, gotOpts.All, "ListContainers must request all containers")
	assert.Len(t, got, 1)
	assert.Equal(t, "myproj-web-1", got[0].Name)
	assert.Equal(t, domain.StatusRunning, got[0].Status)
	assert.Equal(t, "web", got[0].ServiceName)
	assert.Nil(t, got[0].Details)
}

func TestListContainersPropagatesError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("daemon down")
	fake := &fakeAPIClient{
		containerListFn: func(context.Context, container.ListOptions) ([]container.Summary, error) {
			return nil, sentinel
		},
	}

	got, err := NewAdapter(fake).ListContainers(context.Background())
	assert.Nil(t, got)
	assert.ErrorIs(t, err, sentinel)
}

func TestInspectContainerMapsResponse(t *testing.T) {
	t.Parallel()

	var gotID string
	fake := &fakeAPIClient{
		containerInspectFn: func(_ context.Context, containerID string) (container.InspectResponse, error) {
			gotID = containerID
			return container.InspectResponse{
				ContainerJSONBase: &container.ContainerJSONBase{
					State: &container.State{Running: true, ExitCode: 0, Health: &container.Health{Status: "unhealthy"}},
				},
				Config: &container.Config{OpenStdin: true},
			}, nil
		},
	}

	got, err := NewAdapter(fake).InspectContainer(context.Background(), "abc")
	assert.NoError(t, err)
	assert.Equal(t, "abc", gotID)
	assert.Equal(t, domain.ContainerDetails{
		Running:   true,
		Health:    domain.HealthUnhealthy,
		OpenStdin: true,
	}, got)
}

func TestInspectContainerPropagatesError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("no such container")
	fake := &fakeAPIClient{
		containerInspectFn: func(context.Context, string) (container.InspectResponse, error) {
			return container.InspectResponse{}, sentinel
		},
	}

	_, err := NewAdapter(fake).InspectContainer(context.Background(), "abc")
	assert.ErrorIs(t, err, sentinel)
}

func TestLifecycleMethodsCallCorrectSDKMethod(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		// wire installs a spy on the fake that records the id it was called with,
		// and returns a call that invokes the adapter method under test.
		run func(a *Adapter, id string) error
		// install returns the configured fake plus a pointer whose value the spy
		// sets to the id it received.
		install func() (*fakeAPIClient, *string)
	}{
		{
			name: "start",
			install: func() (*fakeAPIClient, *string) {
				got := new(string)
				return &fakeAPIClient{containerStartFn: func(_ context.Context, id string, _ container.StartOptions) error {
					*got = id
					return nil
				}}, got
			},
			run: func(a *Adapter, id string) error { return a.StartContainer(context.Background(), id) },
		},
		{
			name: "stop",
			install: func() (*fakeAPIClient, *string) {
				got := new(string)
				return &fakeAPIClient{containerStopFn: func(_ context.Context, id string, _ container.StopOptions) error {
					*got = id
					return nil
				}}, got
			},
			run: func(a *Adapter, id string) error { return a.StopContainer(context.Background(), id) },
		},
		{
			name: "restart",
			install: func() (*fakeAPIClient, *string) {
				got := new(string)
				return &fakeAPIClient{containerRestartFn: func(_ context.Context, id string, _ container.StopOptions) error {
					*got = id
					return nil
				}}, got
			},
			run: func(a *Adapter, id string) error { return a.RestartContainer(context.Background(), id) },
		},
		{
			name: "pause",
			install: func() (*fakeAPIClient, *string) {
				got := new(string)
				return &fakeAPIClient{containerPauseFn: func(_ context.Context, id string) error {
					*got = id
					return nil
				}}, got
			},
			run: func(a *Adapter, id string) error { return a.PauseContainer(context.Background(), id) },
		},
		{
			name: "unpause",
			install: func() (*fakeAPIClient, *string) {
				got := new(string)
				return &fakeAPIClient{containerUnpauseFn: func(_ context.Context, id string) error {
					*got = id
					return nil
				}}, got
			},
			run: func(a *Adapter, id string) error { return a.UnpauseContainer(context.Background(), id) },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fake, gotID := tc.install()
			err := tc.run(NewAdapter(fake), "container-xyz")
			assert.NoError(t, err)
			assert.Equal(t, "container-xyz", *gotID)
		})
	}
}

func TestRemoveContainerPassesOptions(t *testing.T) {
	t.Parallel()

	var gotID string
	var gotOpts container.RemoveOptions
	fake := &fakeAPIClient{
		containerRemoveFn: func(_ context.Context, id string, opts container.RemoveOptions) error {
			gotID = id
			gotOpts = opts
			return nil
		},
	}

	err := NewAdapter(fake).RemoveContainer(context.Background(), "abc", domain.RemoveOptions{Force: true, RemoveVolumes: true})
	assert.NoError(t, err)
	assert.Equal(t, "abc", gotID)
	assert.Equal(t, container.RemoveOptions{Force: true, RemoveVolumes: true}, gotOpts)
}

func TestRemoveContainerMapsRunningError(t *testing.T) {
	t.Parallel()

	fake := &fakeAPIClient{
		containerRemoveFn: func(context.Context, string, container.RemoveOptions) error {
			return errors.New("Error response from daemon: You cannot remove a running container abc. Stop the container before attempting removal or force remove")
		},
	}

	err := NewAdapter(fake).RemoveContainer(context.Background(), "abc", domain.RemoveOptions{})
	assert.ErrorIs(t, err, domain.ErrContainerRunning)
}

func TestRemoveContainerPassesThroughOtherErrors(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("no such container: abc")
	fake := &fakeAPIClient{
		containerRemoveFn: func(context.Context, string, container.RemoveOptions) error {
			return sentinel
		},
	}

	err := NewAdapter(fake).RemoveContainer(context.Background(), "abc", domain.RemoveOptions{})
	assert.ErrorIs(t, err, sentinel)
	assert.NotErrorIs(t, err, domain.ErrContainerRunning)
}

func TestContainerTopMapsResult(t *testing.T) {
	t.Parallel()

	var gotID string
	var gotArgs []string
	fake := &fakeAPIClient{
		containerTopFn: func(_ context.Context, id string, arguments []string) (container.TopResponse, error) {
			gotID = id
			gotArgs = arguments
			return container.TopResponse{
				Titles:    []string{"PID", "CMD"},
				Processes: [][]string{{"1", "nginx"}, {"42", "sh"}},
			}, nil
		},
	}

	got, err := NewAdapter(fake).ContainerTop(context.Background(), "abc")
	assert.NoError(t, err)
	assert.Equal(t, "abc", gotID)
	assert.Equal(t, []string{}, gotArgs)
	assert.Equal(t, domain.TopResult{
		Titles:    []string{"PID", "CMD"},
		Processes: [][]string{{"1", "nginx"}, {"42", "sh"}},
	}, got)
}

func TestPruneContainers(t *testing.T) {
	t.Parallel()

	called := false
	fake := &fakeAPIClient{
		containersPruneFn: func(_ context.Context, _ filters.Args) (container.PruneReport, error) {
			called = true
			return container.PruneReport{}, nil
		},
	}

	err := NewAdapter(fake).PruneContainers(context.Background())
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestStreamStatsDecodesSamples(t *testing.T) {
	t.Parallel()

	// Two JSON-lines samples; the second has a CPU delta so the derived CPU
	// percentage is non-zero, proving StreamStats runs the Calculate* derivation.
	body := `{"pids_stats":{"current":7}}` + "\n" +
		`{"cpu_stats":{"cpu_usage":{"total_usage":200},"system_cpu_usage":1000},` +
		`"precpu_stats":{"cpu_usage":{"total_usage":100},"system_cpu_usage":600},` +
		`"memory_stats":{"usage":50,"limit":200}}` + "\n"

	var gotID string
	var gotStream bool
	fake := &fakeAPIClient{
		containerStatsFn: func(_ context.Context, id string, stream bool) (container.StatsResponseReader, error) {
			gotID, gotStream = id, stream
			return container.StatsResponseReader{Body: io.NopCloser(strings.NewReader(body))}, nil
		},
	}

	var samples []*domain.RecordedStats
	err := NewAdapter(fake).StreamStats(context.Background(), "abc", func(rs *domain.RecordedStats) {
		samples = append(samples, rs)
	})

	assert.NoError(t, err)
	assert.Equal(t, "abc", gotID)
	assert.True(t, gotStream, "StreamStats must request a streaming reader")
	assert.Len(t, samples, 2)
	assert.Equal(t, 7, samples[0].ClientStats.PidsStats.Current)
	// (200-100)*100 / (1000-600) = 25
	assert.Equal(t, 25.0, samples[1].DerivedStats.CPUPercentage)
	// 50*100 / 200 = 25
	assert.Equal(t, 25.0, samples[1].DerivedStats.MemoryPercentage)
}

func TestStreamStatsPropagatesOpenError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("stats stream unavailable")
	fake := &fakeAPIClient{
		containerStatsFn: func(context.Context, string, bool) (container.StatsResponseReader, error) {
			return container.StatsResponseReader{}, sentinel
		},
	}

	called := false
	err := NewAdapter(fake).StreamStats(context.Background(), "abc", func(*domain.RecordedStats) { called = true })
	assert.Same(t, sentinel, err)
	assert.False(t, called)
}
