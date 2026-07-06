package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/stretchr/testify/assert"
)

// fakeDockerAPI is a test double for domain.DockerAPI. Each method delegates to
// an optional function field so a test wires up only the calls it exercises;
// unset fields return zero values. Mirrors the repo's function-field fakes (see
// pkg/adapter/docker/fake_client_test.go).
type fakeDockerAPI struct {
	listContainersFn   func(ctx context.Context) ([]domain.Container, error)
	inspectContainerFn func(ctx context.Context, id string) (domain.ContainerDetails, error)
	startContainerFn   func(ctx context.Context, id string) error
	stopContainerFn    func(ctx context.Context, id string) error
	restartContainerFn func(ctx context.Context, id string) error
	pauseContainerFn   func(ctx context.Context, id string) error
	unpauseContainerFn func(ctx context.Context, id string) error
	removeContainerFn  func(ctx context.Context, id string, opts domain.RemoveOptions) error
	containerTopFn     func(ctx context.Context, id string) (domain.TopResult, error)
	pruneContainersFn  func(ctx context.Context) error
	streamStatsFn      func(ctx context.Context, id string, onSample func(*domain.RecordedStats)) error
}

var _ domain.DockerAPI = (*fakeDockerAPI)(nil)

func (f *fakeDockerAPI) ListContainers(ctx context.Context) ([]domain.Container, error) {
	if f.listContainersFn != nil {
		return f.listContainersFn(ctx)
	}
	return nil, nil
}

func (f *fakeDockerAPI) InspectContainer(ctx context.Context, id string) (domain.ContainerDetails, error) {
	if f.inspectContainerFn != nil {
		return f.inspectContainerFn(ctx, id)
	}
	return domain.ContainerDetails{}, nil
}

func (f *fakeDockerAPI) StartContainer(ctx context.Context, id string) error {
	if f.startContainerFn != nil {
		return f.startContainerFn(ctx, id)
	}
	return nil
}

func (f *fakeDockerAPI) StopContainer(ctx context.Context, id string) error {
	if f.stopContainerFn != nil {
		return f.stopContainerFn(ctx, id)
	}
	return nil
}

func (f *fakeDockerAPI) RestartContainer(ctx context.Context, id string) error {
	if f.restartContainerFn != nil {
		return f.restartContainerFn(ctx, id)
	}
	return nil
}

func (f *fakeDockerAPI) PauseContainer(ctx context.Context, id string) error {
	if f.pauseContainerFn != nil {
		return f.pauseContainerFn(ctx, id)
	}
	return nil
}

func (f *fakeDockerAPI) UnpauseContainer(ctx context.Context, id string) error {
	if f.unpauseContainerFn != nil {
		return f.unpauseContainerFn(ctx, id)
	}
	return nil
}

func (f *fakeDockerAPI) RemoveContainer(ctx context.Context, id string, opts domain.RemoveOptions) error {
	if f.removeContainerFn != nil {
		return f.removeContainerFn(ctx, id, opts)
	}
	return nil
}

func (f *fakeDockerAPI) ContainerTop(ctx context.Context, id string) (domain.TopResult, error) {
	if f.containerTopFn != nil {
		return f.containerTopFn(ctx, id)
	}
	return domain.TopResult{}, nil
}

func (f *fakeDockerAPI) PruneContainers(ctx context.Context) error {
	if f.pruneContainersFn != nil {
		return f.pruneContainersFn(ctx)
	}
	return nil
}

func (f *fakeDockerAPI) StreamStats(ctx context.Context, id string, onSample func(*domain.RecordedStats)) error {
	if f.streamStatsFn != nil {
		return f.streamStatsFn(ctx, id, onSample)
	}
	return nil
}

const testID = "abc123"

// TestContainerCommandsDelegation verifies each single-ID lifecycle method
// delegates to the matching port call with the container ID it was given, and
// surfaces the port's error unchanged.
func TestContainerCommandsDelegation(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("boom")

	cases := []struct {
		name string
		// wire installs a recorder on the fake and returns a pointer to the id
		// the fake observed; call invokes the ContainerCommands method under test.
		wire func(f *fakeDockerAPI, err *error) *string
		call func(c *ContainerCommands) error
	}{
		{
			name: "Start",
			wire: func(f *fakeDockerAPI, err *error) *string {
				var got string
				f.startContainerFn = func(_ context.Context, id string) error { got = id; return *err }
				return &got
			},
			call: func(c *ContainerCommands) error { return c.Start(context.Background(), testID) },
		},
		{
			name: "Stop",
			wire: func(f *fakeDockerAPI, err *error) *string {
				var got string
				f.stopContainerFn = func(_ context.Context, id string) error { got = id; return *err }
				return &got
			},
			call: func(c *ContainerCommands) error { return c.Stop(context.Background(), testID) },
		},
		{
			name: "Restart",
			wire: func(f *fakeDockerAPI, err *error) *string {
				var got string
				f.restartContainerFn = func(_ context.Context, id string) error { got = id; return *err }
				return &got
			},
			call: func(c *ContainerCommands) error { return c.Restart(context.Background(), testID) },
		},
		{
			name: "Pause",
			wire: func(f *fakeDockerAPI, err *error) *string {
				var got string
				f.pauseContainerFn = func(_ context.Context, id string) error { got = id; return *err }
				return &got
			},
			call: func(c *ContainerCommands) error { return c.Pause(context.Background(), testID) },
		},
		{
			name: "Unpause",
			wire: func(f *fakeDockerAPI, err *error) *string {
				var got string
				f.unpauseContainerFn = func(_ context.Context, id string) error { got = id; return *err }
				return &got
			},
			call: func(c *ContainerCommands) error { return c.Unpause(context.Background(), testID) },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("delegates with id and passes error through", func(t *testing.T) {
				t.Parallel()
				f := &fakeDockerAPI{}
				err := wantErr
				got := tc.wire(f, &err)
				c := NewContainerCommands(f)

				result := tc.call(c)

				assert.Equal(t, testID, *got)
				assert.Same(t, wantErr, result)
			})

			t.Run("success returns nil", func(t *testing.T) {
				t.Parallel()
				f := &fakeDockerAPI{}
				var noErr error
				tc.wire(f, &noErr)
				c := NewContainerCommands(f)

				assert.NoError(t, tc.call(c))
			})
		})
	}
}

// TestContainerCommandsRemove verifies Remove forwards its options to the port
// and surfaces domain.ErrContainerRunning unchanged (branchable with errors.Is),
// as the remove-menu UX depends on that seam.
func TestContainerCommandsRemove(t *testing.T) {
	t.Parallel()

	t.Run("forwards id and options", func(t *testing.T) {
		t.Parallel()
		var gotID string
		var gotOpts domain.RemoveOptions
		f := &fakeDockerAPI{
			removeContainerFn: func(_ context.Context, id string, opts domain.RemoveOptions) error {
				gotID, gotOpts = id, opts
				return nil
			},
		}
		c := NewContainerCommands(f)

		wantOpts := domain.RemoveOptions{Force: true, RemoveVolumes: true}
		err := c.Remove(context.Background(), testID, wantOpts)

		assert.NoError(t, err)
		assert.Equal(t, testID, gotID)
		assert.Equal(t, wantOpts, gotOpts)
	})

	t.Run("surfaces ErrContainerRunning unchanged", func(t *testing.T) {
		t.Parallel()
		// Wrap the sentinel the way the adapter does, to prove errors.Is still
		// matches through the use case.
		f := &fakeDockerAPI{
			removeContainerFn: func(_ context.Context, _ string, _ domain.RemoveOptions) error {
				return fmt.Errorf("%w: docker says stop it first", domain.ErrContainerRunning)
			},
		}
		c := NewContainerCommands(f)

		err := c.Remove(context.Background(), testID, domain.RemoveOptions{})

		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrContainerRunning))
	})

	t.Run("passes unrelated error through", func(t *testing.T) {
		t.Parallel()
		wantErr := errors.New("some other docker failure")
		f := &fakeDockerAPI{
			removeContainerFn: func(_ context.Context, _ string, _ domain.RemoveOptions) error {
				return wantErr
			},
		}
		c := NewContainerCommands(f)

		err := c.Remove(context.Background(), testID, domain.RemoveOptions{})

		assert.Same(t, wantErr, err)
		assert.False(t, errors.Is(err, domain.ErrContainerRunning))
	})
}
