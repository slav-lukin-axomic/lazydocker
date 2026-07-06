// Package compose is the driven adapter for Docker Compose lifecycle operations.
// It implements the domain.ComposeRunner port by rendering the user-configured
// command templates and shelling out via the shared OSCommand — the exact
// behaviour the pre-migration commands.Service methods had, relocated out of the
// core so pkg/domain and pkg/usecase stay framework- and subprocess-free.
package compose

import (
	"context"

	"github.com/jesseduffield/lazydocker/pkg/commands"
	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/utils"
)

// Runner implements domain.ComposeRunner by rendering the user's Compose command
// templates against a CommandObject and running them through the OSCommand.
type Runner struct {
	osCommand *commands.OSCommand
	builder   commands.LimitedDockerCommand
}

var _ domain.ComposeRunner = &Runner{}

// NewRunner returns a Runner that renders templates via builder.NewCommandObject
// and executes them through osCommand.
func NewRunner(osCommand *commands.OSCommand, builder commands.LimitedDockerCommand) *Runner {
	return &Runner{osCommand: osCommand, builder: builder}
}

// Stop stops the service's containers.
func (r *Runner) Stop(svc *domain.Service) error {
	return r.run(svc, r.osCommand.Config.UserConfig.CommandTemplates.StopService)
}

// Up up's the service.
func (r *Runner) Up(svc *domain.Service) error {
	return r.run(svc, r.osCommand.Config.UserConfig.CommandTemplates.UpService)
}

// Start starts the service.
func (r *Runner) Start(svc *domain.Service) error {
	return r.run(svc, r.osCommand.Config.UserConfig.CommandTemplates.StartService)
}

// Restart restarts the service.
func (r *Runner) Restart(svc *domain.Service) error {
	return r.run(svc, r.osCommand.Config.UserConfig.CommandTemplates.RestartService)
}

func (r *Runner) run(svc *domain.Service, templateCmdStr string) error {
	command := utils.ApplyTemplate(
		templateCmdStr,
		r.builder.NewCommandObject(commands.CommandObject{Service: svc}),
	)
	return r.osCommand.RunCommand(command)
}

// Top renders the process list of the service.
func (r *Runner) Top(ctx context.Context, svc *domain.Service) (string, error) {
	command := utils.ApplyTemplate(
		r.osCommand.Config.UserConfig.CommandTemplates.ServiceTop,
		r.builder.NewCommandObject(commands.CommandObject{Service: svc}),
	)

	return r.osCommand.RunCommandWithOutputContext(ctx, command)
}
