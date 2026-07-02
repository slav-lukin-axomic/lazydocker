// Package usecase is the application layer of the hexagon: it orchestrates
// domain operations over the driven ports the core defines. It is framework-free
// and imports only context and pkg/domain — never the Docker SDK, a TUI
// framework, or pkg/gui. A depguard rule enforces this.
package usecase

import (
	"context"

	"github.com/jesseduffield/lazydocker/pkg/domain"
)

// ContainerCommands drives container lifecycle operations over the DockerAPI
// port. It is the seam the GUI depends on for state changes, so the GUI no
// longer reaches for the SDK or commands.Container to start/stop/remove.
type ContainerCommands struct {
	api domain.DockerAPI
}

// NewContainerCommands returns a ContainerCommands backed by the given port.
func NewContainerCommands(api domain.DockerAPI) *ContainerCommands {
	return &ContainerCommands{api: api}
}

// Start starts the container with the given ID.
func (c *ContainerCommands) Start(ctx context.Context, id string) error {
	return c.api.StartContainer(ctx, id)
}

// Stop stops the container with the given ID.
func (c *ContainerCommands) Stop(ctx context.Context, id string) error {
	return c.api.StopContainer(ctx, id)
}

// Restart restarts the container with the given ID.
func (c *ContainerCommands) Restart(ctx context.Context, id string) error {
	return c.api.RestartContainer(ctx, id)
}

// Pause pauses the container with the given ID.
func (c *ContainerCommands) Pause(ctx context.Context, id string) error {
	return c.api.PauseContainer(ctx, id)
}

// Unpause unpauses the container with the given ID.
func (c *ContainerCommands) Unpause(ctx context.Context, id string) error {
	return c.api.UnpauseContainer(ctx, id)
}

// Remove removes the container with the given ID. When Docker refuses to remove
// a running container without force, it surfaces domain.ErrContainerRunning
// unchanged so callers can branch on it with errors.Is.
func (c *ContainerCommands) Remove(ctx context.Context, id string, opts domain.RemoveOptions) error {
	return c.api.RemoveContainer(ctx, id, opts)
}
