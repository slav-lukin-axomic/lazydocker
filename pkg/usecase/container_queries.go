package usecase

import (
	"context"
	"io"
	"sync"

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

// List returns all containers with their inspect details populated. It mirrors
// the legacy DockerCommand.GetContainers+SetContainerDetails: containers whose
// inspect fails are returned with nil Details rather than failing the batch, so
// the details load resiliently (the DetailsLoaded gate handles a nil Details).
func (c *ContainerQueries) List(ctx context.Context) ([]*domain.Container, error) {
	containers, err := c.api.ListContainers(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.Container, len(containers))
	for i := range containers {
		ctr := containers[i]
		result[i] = &ctr
	}
	c.loadDetails(ctx, result)
	return result, nil
}

// RefreshDetails re-inspects the given containers and updates their Details in
// place. Like the pre-migration SetContainerDetails it mutates the shared
// containers concurrently; callers already tolerate that read/write overlap.
func (c *ContainerQueries) RefreshDetails(ctx context.Context, containers []*domain.Container) error {
	c.loadDetails(ctx, containers)
	return nil
}

// loadDetails inspects each container in parallel, setting Details on success
// and leaving it unchanged (nil) on a per-container error.
func (c *ContainerQueries) loadDetails(ctx context.Context, containers []*domain.Container) {
	var wg sync.WaitGroup
	for _, ctr := range containers {
		ctr := ctr
		wg.Add(1)
		go func() {
			defer wg.Done()
			details, err := c.api.InspectContainer(ctx, ctr.ID)
			if err != nil {
				return
			}
			ctr.Details = &details
		}()
	}
	wg.Wait()
}
