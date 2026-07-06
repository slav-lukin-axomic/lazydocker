package commands

import (
	"context"
	"errors"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/jesseduffield/lazydocker/pkg/i18n"
	"github.com/stretchr/testify/assert"
)

func newTestContainer(t *testing.T, cli DockerClient) *Container {
	t.Helper()
	return &Container{
		ID:     "abc123",
		Name:   "test-container",
		Client: cli,
		Log:    NewDummyLog(),
		Tr:     i18n.NewTranslationSet(NewDummyLog(), "en"),
	}
}

// TestContainerRemove covers the mapping of the daemon's "stop the container
// first" error onto a ComplexError with the MustStopContainer code, which the
// GUI branches on to prompt for a force-remove.
func TestContainerRemove(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name              string
		removeErr         error
		wantMustStopCode  bool
		wantErr           bool
		wantPassthroughEq bool
	}{
		{
			name:      "success",
			removeErr: nil,
			wantErr:   false,
		},
		{
			name:             "stop-first error maps to MustStopContainer",
			removeErr:        errors.New("cannot remove: Stop the container before attempting removal or force remove"),
			wantErr:          true,
			wantMustStopCode: true,
		},
		{
			name:              "unrelated error passes through unchanged",
			removeErr:         errors.New("some other docker failure"),
			wantErr:           true,
			wantPassthroughEq: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cli := &fakeDockerClient{
				containerRemoveFn: func(_ context.Context, _ string, _ container.RemoveOptions) error {
					return tc.removeErr
				},
			}
			ctr := newTestContainer(t, cli)

			err := ctr.Remove(container.RemoveOptions{})

			if !tc.wantErr {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
			if tc.wantMustStopCode {
				assert.True(t, HasErrorCode(err, MustStopContainer))
			}
			if tc.wantPassthroughEq {
				assert.False(t, HasErrorCode(err, MustStopContainer))
				assert.Equal(t, tc.removeErr, err)
			}
		})
	}
}

// TestContainerTop covers the "container is not running" guard, which relies on
// the inspect result rather than the cached summary.
func TestContainerTop(t *testing.T) {
	t.Parallel()

	t.Run("not running returns guard error", func(t *testing.T) {
		t.Parallel()
		cli := &fakeDockerClient{
			containerInspectFn: func(_ context.Context, _ string) (container.InspectResponse, error) {
				return container.InspectResponse{
					ContainerJSONBase: &container.ContainerJSONBase{
						State: &container.State{Running: false},
					},
				}, nil
			},
		}
		ctr := newTestContainer(t, cli)

		_, err := ctr.Top(context.Background())

		assert.Error(t, err)
		assert.Equal(t, "container is not running", err.Error())
	})

	t.Run("running delegates to ContainerTop", func(t *testing.T) {
		t.Parallel()
		wantTitles := []string{"PID", "CMD"}
		cli := &fakeDockerClient{
			containerInspectFn: func(_ context.Context, _ string) (container.InspectResponse, error) {
				return container.InspectResponse{
					ContainerJSONBase: &container.ContainerJSONBase{
						State: &container.State{Running: true},
					},
				}, nil
			},
			containerTopFn: func(_ context.Context, _ string, _ []string) (container.TopResponse, error) {
				return container.TopResponse{Titles: wantTitles}, nil
			},
		}
		ctr := newTestContainer(t, cli)

		result, err := ctr.Top(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, wantTitles, result.Titles)
	})

	t.Run("inspect error propagates", func(t *testing.T) {
		t.Parallel()
		inspectErr := errors.New("inspect boom")
		cli := &fakeDockerClient{
			containerInspectFn: func(_ context.Context, _ string) (container.InspectResponse, error) {
				return container.InspectResponse{}, inspectErr
			},
		}
		ctr := newTestContainer(t, cli)

		_, err := ctr.Top(context.Background())

		assert.Equal(t, inspectErr, err)
	})
}
