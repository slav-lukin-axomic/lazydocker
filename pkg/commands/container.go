package commands

import (
	"context"

	"github.com/docker/docker/api/types/filters"
)

// PruneContainers prunes containers
func (c *DockerCommand) PruneContainers() error {
	_, err := c.Client.ContainersPrune(context.Background(), filters.Args{})
	return err
}
