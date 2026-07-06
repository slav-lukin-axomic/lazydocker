package usecase

import (
	"context"
	"io"

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

// StreamLogs streams a container's logs to out via the port. Blocks until the
// stream ends or ctx is cancelled.
func (c *ContainerQueries) StreamLogs(ctx context.Context, id string, opts domain.LogOptions, out io.Writer) error {
	return c.api.StreamLogs(ctx, id, opts, out)
}

// Details returns the lean inspect projection (Running/ExitCode/Health/OpenStdin).
func (c *ContainerQueries) Details(ctx context.Context, id string) (domain.ContainerDetails, error) {
	return c.api.InspectContainer(ctx, id)
}
