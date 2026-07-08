package domain

import "context"

// Service is the flat, framework-free model of a Docker Compose service. It
// replaces the pre-migration commands.Service, which carried an OSCommand and a
// logger to shell out compose operations itself. Those operations now live
// behind the ComposeRunner port. Container is nil until a running container is
// matched to the service.
type Service struct {
	Name        string
	ID          string
	ProjectName string
	Container   *Container
}

// ComposeRunner is the driven port for Compose operations: single-service
// lifecycle changes and a project-scoped config read. It is consumer-defined here
// in the core and implemented by the compose adapter, which owns the template
// rendering and subprocess execution.
//
// The blocking ops take no context because the underlying subprocess runner has
// no context-aware variant; Top streams output and so is cancellable via ctx.
type ComposeRunner interface {
	Stop(svc *Service) error
	Up(svc *Service) error
	Start(svc *Service) error
	Restart(svc *Service) error
	Top(ctx context.Context, svc *Service) (string, error)
	Config(project *Project) (string, error)
}
