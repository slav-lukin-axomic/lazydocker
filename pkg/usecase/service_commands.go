package usecase

import (
	"context"

	"github.com/jesseduffield/lazydocker/pkg/domain"
)

// ServiceCommands drives Compose service lifecycle operations over the
// ComposeRunner port. It is the seam the GUI depends on for service state
// changes, so the GUI no longer reaches for commands.Service to stop/up/restart.
type ServiceCommands struct {
	runner domain.ComposeRunner
}

// NewServiceCommands returns a ServiceCommands backed by the given port.
func NewServiceCommands(runner domain.ComposeRunner) *ServiceCommands {
	return &ServiceCommands{runner: runner}
}

// Stop stops the service's containers.
func (s *ServiceCommands) Stop(svc *domain.Service) error {
	return s.runner.Stop(svc)
}

// Up up's the service.
func (s *ServiceCommands) Up(svc *domain.Service) error {
	return s.runner.Up(svc)
}

// Start starts the service.
func (s *ServiceCommands) Start(svc *domain.Service) error {
	return s.runner.Start(svc)
}

// Restart restarts the service.
func (s *ServiceCommands) Restart(svc *domain.Service) error {
	return s.runner.Restart(svc)
}

// Top renders the process list of the service.
func (s *ServiceCommands) Top(ctx context.Context, svc *domain.Service) (string, error) {
	return s.runner.Top(ctx, svc)
}
