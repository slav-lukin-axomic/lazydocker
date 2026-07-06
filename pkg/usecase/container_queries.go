package usecase

import (
	"context"

	"github.com/jesseduffield/lazydocker/pkg/domain"
)

// ContainerQueries drives read-only container inspection over the DockerAPI
// port. It is kept separate from ContainerCommands so the GUI's detail views
// depend only on the query seam, not the lifecycle mutations.
type ContainerQueries struct {
	api domain.DockerAPI
}

// NewContainerQueries returns a ContainerQueries backed by the given port.
func NewContainerQueries(api domain.DockerAPI) *ContainerQueries {
	return &ContainerQueries{api: api}
}

// Top returns the process listing for the container with the given ID.
func (c *ContainerQueries) Top(ctx context.Context, id string) (domain.TopResult, error) {
	return c.api.ContainerTop(ctx, id)
}

// Inspect returns the inspect projection and raw YAML dump for the Config/Env
// detail views. Identity fields on the projection are left for the caller to fill.
func (c *ContainerQueries) Inspect(ctx context.Context, id string) (domain.ContainerInspect, string, error) {
	return c.api.InspectContainerVerbose(ctx, id)
}
