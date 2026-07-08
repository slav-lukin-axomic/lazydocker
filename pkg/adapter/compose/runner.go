// Package compose is the driven adapter for Docker Compose lifecycle operations.
// It implements the domain.ComposeRunner port by rendering the user-configured
// command templates against an oscommand.CommandObject and shelling out via the
// shared OSCommand — the exact behaviour the pre-migration commands.Service
// methods had, relocated out of the core so pkg/domain and pkg/usecase stay
// framework- and subprocess-free, and without depending on the legacy
// pkg/commands layer.
package compose

import (
	"context"

	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/oscommand"
	"github.com/jesseduffield/lazydocker/pkg/utils"
)

// Runner implements domain.ComposeRunner by rendering the user's Compose command
// templates against a CommandObject and running them through the OSCommand.
type Runner struct {
	osCommand *oscommand.OSCommand
}

var _ domain.ComposeRunner = &Runner{}

// NewRunner returns a Runner that renders templates via oscommand.NewCommandObject
// and executes them through osCommand.
func NewRunner(osCommand *oscommand.OSCommand) *Runner {
	return &Runner{osCommand: osCommand}
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
		oscommand.NewCommandObject(r.osCommand.Config.UserConfig.CommandTemplates.DockerCompose, oscommand.CommandObject{Service: svc}),
	)
	return r.osCommand.RunCommand(command)
}

// Top renders the process list of the service.
func (r *Runner) Top(ctx context.Context, svc *domain.Service) (string, error) {
	command := utils.ApplyTemplate(
		r.osCommand.Config.UserConfig.CommandTemplates.ServiceTop,
		oscommand.NewCommandObject(r.osCommand.Config.UserConfig.CommandTemplates.DockerCompose, oscommand.CommandObject{Service: svc}),
	)

	return r.osCommand.RunCommandWithOutputContext(ctx, command)
}

// Config renders the docker-compose config of the given project.
func (r *Runner) Config(project *domain.Project) (string, error) {
	command := utils.ApplyTemplate(
		r.osCommand.Config.UserConfig.CommandTemplates.DockerComposeConfig,
		oscommand.NewCommandObject(r.osCommand.Config.UserConfig.CommandTemplates.DockerCompose, oscommand.CommandObject{Project: project}),
	)
	return r.osCommand.RunCommandWithOutput(command)
}
