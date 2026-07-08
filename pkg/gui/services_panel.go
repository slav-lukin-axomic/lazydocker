package gui

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/fatih/color"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazydocker/pkg/config"
	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/gui/panels"
	"github.com/jesseduffield/lazydocker/pkg/gui/presentation"
	"github.com/jesseduffield/lazydocker/pkg/gui/types"
	"github.com/jesseduffield/lazydocker/pkg/oscommand"
	"github.com/jesseduffield/lazydocker/pkg/tasks"
	"github.com/jesseduffield/lazydocker/pkg/utils"
	"github.com/samber/lo"
)

func (gui *Gui) getServicesPanel() *panels.SideListPanel[*domain.Service] {
	return &panels.SideListPanel[*domain.Service]{
		ContextState: &panels.ContextState[*domain.Service]{
			GetMainTabs: func() []panels.MainTab[*domain.Service] {
				return []panels.MainTab[*domain.Service]{
					{
						Key:    "logs",
						Title:  gui.Tr.LogsTitle,
						Render: gui.renderServiceLogs,
					},
					{
						Key:    "stats",
						Title:  gui.Tr.StatsTitle,
						Render: gui.renderServiceStats,
					},
					{
						Key:    "container-env",
						Title:  gui.Tr.ContainerEnvTitle,
						Render: gui.renderServiceContainerEnv,
					},
					{
						Key:    "container-config",
						Title:  gui.Tr.ContainerConfigTitle,
						Render: gui.renderServiceContainerConfig,
					},
					{
						Key:    "top",
						Title:  gui.Tr.TopTitle,
						Render: gui.renderServiceTop,
					},
				}
			},
			GetItemContextCacheKey: func(service *domain.Service) string {
				if service.Container == nil {
					return "services-" + service.ID
				}
				return "services-" + service.ID + "-" + service.Container.ID + "-" + service.Container.Status.String()
			},
		},
		ListPanel: panels.ListPanel[*domain.Service]{
			List: panels.NewFilteredList[*domain.Service](),
			View: gui.Views.Services,
		},
		NoItemsMessage: gui.Tr.NoServices,
		Gui:            gui.intoInterface(),
		// sort services first by whether they have a linked container, and second by alphabetical order
		Sort: func(a *domain.Service, b *domain.Service) bool {
			if a.Container != nil && b.Container == nil {
				return true
			}

			if a.Container == nil && b.Container != nil {
				return false
			}

			return a.Name < b.Name
		},
		Filter: func(service *domain.Service) bool {
			selectedProject := gui.getSelectedProjectName()
			if selectedProject == "" {
				// Before any project is selected (e.g. startup), default to
				// the local project so we don't briefly flash all services.
				selectedProject = gui.DockerCommand.LocalProjectName
			}
			if selectedProject == "" {
				return true
			}
			return service.ProjectName == selectedProject
		},
		GetTableCells: func(service *domain.Service) []string {
			var stats *domain.DerivedStats
			if service.Container != nil {
				if last, ok := gui.StatsMonitor.LastStats(service.Container.ID); ok {
					stats = &last.DerivedStats
				}
			}
			return presentation.GetServiceDisplayStrings(&gui.Config.UserConfig.Gui, service, stats)
		},
		Hide: func() bool {
			return !gui.DockerCommand.IsProjectScoped()
		},
	}
}

func (gui *Gui) renderServiceContainerConfig(service *domain.Service) tasks.TaskFunc {
	if service.Container == nil {
		return gui.NewSimpleRenderStringTask(func() string { return gui.Tr.NoContainer })
	}

	return gui.renderContainerConfig(service.Container)
}

func (gui *Gui) renderServiceContainerEnv(service *domain.Service) tasks.TaskFunc {
	if service.Container == nil {
		return gui.NewSimpleRenderStringTask(func() string { return gui.Tr.NoContainer })
	}

	return gui.renderContainerEnv(service.Container)
}

func (gui *Gui) renderServiceStats(service *domain.Service) tasks.TaskFunc {
	if service.Container == nil {
		return gui.NewSimpleRenderStringTask(func() string { return gui.Tr.NoContainer })
	}

	return gui.renderContainerStats(service.Container)
}

func (gui *Gui) renderServiceTop(service *domain.Service) tasks.TaskFunc {
	return gui.NewTickerTask(TickerTaskOpts{
		Func: func(ctx context.Context, notifyStopped chan struct{}) {
			contents, err := gui.ServiceCommands.Top(ctx, service)
			if err != nil {
				gui.RenderStringMain(err.Error())
			}

			gui.reRenderStringMain(contents)
		},
		Duration:   time.Second,
		Before:     func(ctx context.Context) { gui.clearMainView() },
		Wrap:       gui.Config.UserConfig.Gui.WrapMainPanel,
		Autoscroll: false,
	})
}

func (gui *Gui) renderServiceLogs(service *domain.Service) tasks.TaskFunc {
	if service.Container == nil {
		return gui.NewSimpleRenderStringTask(func() string { return gui.Tr.NoContainerForService })
	}

	return gui.renderContainerLogsToMain(service.Container)
}

type commandOption struct {
	description string
	command     string
	onPress     func() error
}

func (r *commandOption) getDisplayStrings() []string {
	return []string{r.description, color.New(color.FgCyan).Sprint(r.command)}
}

// isServiceFromLocalProject returns true if the given service belongs to the
// local compose project (the one whose compose file is in the current directory).
// Compose commands like up/stop/restart only work for local project services.
func (gui *Gui) isServiceFromLocalProject(service *domain.Service) bool {
	return service.ProjectName == gui.DockerCommand.LocalProjectName
}

func (gui *Gui) handleServiceRemoveMenu(g *gocui.Gui, v *gocui.View) error {
	service, err := gui.Panels.Services.GetSelectedItem()
	if err != nil {
		return nil
	}

	if !gui.isServiceFromLocalProject(service) {
		return gui.createErrorPanel(gui.Tr.CannotManageNonLocalService)
	}

	composeCommand := gui.DockerCommand.NewCommandObject(oscommand.CommandObject{Service: service}).DockerCompose

	options := []*commandOption{
		{
			description: gui.Tr.Remove,
			command:     fmt.Sprintf("%s rm --stop --force %s", composeCommand, service.Name),
		},
		{
			description: gui.Tr.RemoveWithVolumes,
			command:     fmt.Sprintf("%s rm --stop --force -v %s", composeCommand, service.Name),
		},
	}

	menuItems := lo.Map(options, func(option *commandOption, _ int) *types.MenuItem {
		return &types.MenuItem{
			LabelColumns: option.getDisplayStrings(),
			OnPress: func() error {
				return gui.WithWaitingStatus(gui.Tr.RemovingStatus, func() error {
					if err := gui.OSCommand.RunCommand(option.command); err != nil {
						return gui.createErrorPanel(err.Error())
					}

					return nil
				})
			},
		}
	})

	return gui.Menu(CreateMenuOptions{
		Title: "",
		Items: menuItems,
	})
}

func (gui *Gui) handleServicePause(g *gocui.Gui, v *gocui.View) error {
	service, err := gui.Panels.Services.GetSelectedItem()
	if err != nil {
		return nil
	}
	if service.Container == nil {
		return nil
	}

	return gui.PauseContainer(service.Container)
}

func (gui *Gui) handleServiceStop(g *gocui.Gui, v *gocui.View) error {
	service, err := gui.Panels.Services.GetSelectedItem()
	if err != nil {
		return nil
	}

	return gui.createConfirmationPanel(gui.Tr.Confirm, gui.Tr.StopService, func(g *gocui.Gui, v *gocui.View) error {
		return gui.WithWaitingStatus(gui.Tr.StoppingStatus, func() error {
			if !gui.isServiceFromLocalProject(service) {
				if service.Container == nil {
					return gui.createErrorPanel(gui.Tr.CannotManageNonLocalService)
				}
				return gui.ContainerCommands.Stop(context.Background(), service.Container.ID)
			}
			if err := gui.ServiceCommands.Stop(service); err != nil {
				return gui.createErrorPanel(err.Error())
			}
			return nil
		})
	}, nil)
}

func (gui *Gui) handleServiceUp(g *gocui.Gui, v *gocui.View) error {
	service, err := gui.Panels.Services.GetSelectedItem()
	if err != nil {
		return nil
	}

	if !gui.isServiceFromLocalProject(service) {
		return gui.createErrorPanel(gui.Tr.CannotManageNonLocalService)
	}

	return gui.WithWaitingStatus(gui.Tr.UppingServiceStatus, func() error {
		if err := gui.ServiceCommands.Up(service); err != nil {
			return gui.createErrorPanel(err.Error())
		}

		return nil
	})
}

func (gui *Gui) handleServiceRestart(g *gocui.Gui, v *gocui.View) error {
	service, err := gui.Panels.Services.GetSelectedItem()
	if err != nil {
		return nil
	}

	return gui.WithWaitingStatus(gui.Tr.RestartingStatus, func() error {
		if !gui.isServiceFromLocalProject(service) {
			if service.Container == nil {
				return gui.createErrorPanel(gui.Tr.CannotManageNonLocalService)
			}
			return gui.ContainerCommands.Restart(context.Background(), service.Container.ID)
		}
		if err := gui.ServiceCommands.Restart(service); err != nil {
			return gui.createErrorPanel(err.Error())
		}
		return nil
	})
}

func (gui *Gui) handleServiceStart(g *gocui.Gui, v *gocui.View) error {
	service, err := gui.Panels.Services.GetSelectedItem()
	if err != nil {
		return nil
	}

	if !gui.isServiceFromLocalProject(service) {
		if service.Container == nil {
			return gui.createErrorPanel(gui.Tr.CannotManageNonLocalService)
		}
		return gui.WithWaitingStatus(gui.Tr.StartingStatus, func() error {
			return gui.ContainerCommands.Start(context.Background(), service.Container.ID)
		})
	}

	return gui.WithWaitingStatus(gui.Tr.StartingStatus, func() error {
		if err := gui.ServiceCommands.Start(service); err != nil {
			return gui.createErrorPanel(err.Error())
		}
		return nil
	})
}

func (gui *Gui) handleServiceAttach(g *gocui.Gui, v *gocui.View) error {
	service, err := gui.Panels.Services.GetSelectedItem()
	if err != nil {
		return nil
	}

	if service.Container == nil {
		return gui.createErrorPanel(gui.Tr.NoContainers)
	}

	c, err := gui.attachToContainer(service.Container)
	if err != nil {
		return gui.createErrorPanel(err.Error())
	}

	return gui.runSubprocess(c)
}

func (gui *Gui) handleServiceRenderLogsToMain(g *gocui.Gui, v *gocui.View) error {
	service, err := gui.Panels.Services.GetSelectedItem()
	if err != nil {
		return nil
	}

	c, err := gui.viewServiceLogs(service)
	if err != nil {
		return gui.createErrorPanel(err.Error())
	}

	return gui.runSubprocess(c)
}

// viewServiceLogs builds a subprocess that tails the service's logs. It renders
// the ViewServiceLogs template through the shared command object, matching what
// the pre-migration commands.Service.ViewLogs produced. It lives on the GUI
// (like attachToContainer) rather than the compose adapter because the resulting
// exec.Cmd is driven interactively by the GUI's subprocess machinery.
func (gui *Gui) viewServiceLogs(service *domain.Service) (*exec.Cmd, error) {
	command := utils.ApplyTemplate(
		gui.Config.UserConfig.CommandTemplates.ViewServiceLogs,
		gui.DockerCommand.NewCommandObject(oscommand.CommandObject{Service: service}),
	)

	cmd := gui.OSCommand.ExecutableFromString(command)
	gui.OSCommand.PrepareForChildren(cmd)

	return cmd, nil
}

func (gui *Gui) handleProjectUp(g *gocui.Gui, v *gocui.View) error {
	project, _ := gui.Panels.Projects.GetSelectedItem()
	if project != nil && project.Name != gui.DockerCommand.LocalProjectName {
		return gui.createErrorPanel(gui.Tr.CannotManageNonLocalService)
	}
	return gui.createConfirmationPanel(gui.Tr.Confirm, gui.Tr.ConfirmUpProject, func(g *gocui.Gui, v *gocui.View) error {
		cmdStr := utils.ApplyTemplate(
			gui.Config.UserConfig.CommandTemplates.Up,
			gui.DockerCommand.NewCommandObject(oscommand.CommandObject{Project: project}),
		)

		return gui.WithWaitingStatus(gui.Tr.UppingProjectStatus, func() error {
			if err := gui.OSCommand.RunCommand(cmdStr); err != nil {
				return gui.createErrorPanel(err.Error())
			}
			return nil
		})
	}, nil)
}

func (gui *Gui) handleProjectDown(g *gocui.Gui, v *gocui.View) error {
	project, _ := gui.Panels.Projects.GetSelectedItem()
	if project != nil && project.Name != gui.DockerCommand.LocalProjectName {
		return gui.createErrorPanel(gui.Tr.CannotManageNonLocalService)
	}
	downCommand := utils.ApplyTemplate(
		gui.Config.UserConfig.CommandTemplates.Down,
		gui.DockerCommand.NewCommandObject(oscommand.CommandObject{Project: project}),
	)

	downWithVolumesCommand := utils.ApplyTemplate(
		gui.Config.UserConfig.CommandTemplates.DownWithVolumes,
		gui.DockerCommand.NewCommandObject(oscommand.CommandObject{Project: project}),
	)

	options := []*commandOption{
		{
			description: gui.Tr.Down,
			command:     downCommand,
			onPress: func() error {
				return gui.WithWaitingStatus(gui.Tr.DowningStatus, func() error {
					if err := gui.OSCommand.RunCommand(downCommand); err != nil {
						return gui.createErrorPanel(err.Error())
					}
					return nil
				})
			},
		},
		{
			description: gui.Tr.DownWithVolumes,
			command:     downWithVolumesCommand,
			onPress: func() error {
				return gui.WithWaitingStatus(gui.Tr.DowningStatus, func() error {
					if err := gui.OSCommand.RunCommand(downWithVolumesCommand); err != nil {
						return gui.createErrorPanel(err.Error())
					}
					return nil
				})
			},
		},
	}

	menuItems := lo.Map(options, func(option *commandOption, _ int) *types.MenuItem {
		return &types.MenuItem{
			LabelColumns: option.getDisplayStrings(),
			OnPress:      option.onPress,
		}
	})

	return gui.Menu(CreateMenuOptions{
		Title: "",
		Items: menuItems,
	})
}

func (gui *Gui) handleServiceRestartMenu(g *gocui.Gui, v *gocui.View) error {
	service, err := gui.Panels.Services.GetSelectedItem()
	if err != nil {
		return nil
	}

	if !gui.isServiceFromLocalProject(service) {
		return gui.createErrorPanel(gui.Tr.CannotManageNonLocalService)
	}

	rebuildCommand := utils.ApplyTemplate(
		gui.Config.UserConfig.CommandTemplates.RebuildService,
		gui.DockerCommand.NewCommandObject(oscommand.CommandObject{Service: service}),
	)

	recreateCommand := utils.ApplyTemplate(
		gui.Config.UserConfig.CommandTemplates.RecreateService,
		gui.DockerCommand.NewCommandObject(oscommand.CommandObject{Service: service}),
	)

	options := []*commandOption{
		{
			description: gui.Tr.Restart,
			command: utils.ApplyTemplate(
				gui.Config.UserConfig.CommandTemplates.RestartService,
				gui.DockerCommand.NewCommandObject(oscommand.CommandObject{Service: service}),
			),
			onPress: func() error {
				return gui.WithWaitingStatus(gui.Tr.RestartingStatus, func() error {
					if err := gui.ServiceCommands.Restart(service); err != nil {
						return gui.createErrorPanel(err.Error())
					}
					return nil
				})
			},
		},
		{
			description: gui.Tr.Recreate,
			command: utils.ApplyTemplate(
				gui.Config.UserConfig.CommandTemplates.RecreateService,
				gui.DockerCommand.NewCommandObject(oscommand.CommandObject{Service: service}),
			),
			onPress: func() error {
				return gui.WithWaitingStatus(gui.Tr.RestartingStatus, func() error {
					if err := gui.OSCommand.RunCommand(recreateCommand); err != nil {
						return gui.createErrorPanel(err.Error())
					}
					return nil
				})
			},
		},
		{
			description: gui.Tr.Rebuild,
			command: utils.ApplyTemplate(
				gui.Config.UserConfig.CommandTemplates.RebuildService,
				gui.DockerCommand.NewCommandObject(oscommand.CommandObject{Service: service}),
			),
			onPress: func() error {
				return gui.runSubprocess(gui.OSCommand.RunCustomCommand(rebuildCommand))
			},
		},
	}

	menuItems := lo.Map(options, func(option *commandOption, _ int) *types.MenuItem {
		return &types.MenuItem{
			LabelColumns: option.getDisplayStrings(),
			OnPress:      option.onPress,
		}
	})

	return gui.Menu(CreateMenuOptions{
		Title: "",
		Items: menuItems,
	})
}

func (gui *Gui) handleServicesCustomCommand(g *gocui.Gui, v *gocui.View) error {
	service, err := gui.Panels.Services.GetSelectedItem()
	if err != nil {
		return nil
	}

	commandObject := gui.DockerCommand.NewCommandObject(oscommand.CommandObject{
		Service:   service,
		Container: service.Container,
	})

	var customCommands []config.CustomCommand

	customServiceCommands := gui.Config.UserConfig.CustomCommands.Services
	// we only include service commands if they have no serviceNames defined or if our service happens to be one of the serviceNames defined
L:
	for _, cmd := range customServiceCommands {
		if len(cmd.ServiceNames) == 0 {
			customCommands = append(customCommands, cmd)
			continue L
		}
		for _, serviceName := range cmd.ServiceNames {
			if serviceName == service.Name {
				// appending these to the top given they're more likely to be selected
				customCommands = append([]config.CustomCommand{cmd}, customCommands...)
				continue L
			}
		}
	}

	if service.Container != nil {
		customCommands = append(customCommands, gui.Config.UserConfig.CustomCommands.Containers...)
	}

	return gui.createCustomCommandMenu(customCommands, commandObject)
}

func (gui *Gui) handleServicesBulkCommand(g *gocui.Gui, v *gocui.View) error {
	project, _ := gui.Panels.Projects.GetSelectedItem()
	bulkCommands := gui.Config.UserConfig.BulkCommands.Services
	commandObject := gui.DockerCommand.NewCommandObject(oscommand.CommandObject{Project: project})

	return gui.createBulkCommandMenu(bulkCommands, commandObject)
}

func (gui *Gui) handleServicesExecShell(g *gocui.Gui, v *gocui.View) error {
	service, err := gui.Panels.Services.GetSelectedItem()
	if err != nil {
		return nil
	}

	container := service.Container
	if container == nil {
		return gui.createErrorPanel(gui.Tr.NoContainers)
	}

	return gui.containerExecShell(container)
}

func (gui *Gui) handleServicesOpenInBrowserCommand(g *gocui.Gui, v *gocui.View) error {
	service, err := gui.Panels.Services.GetSelectedItem()
	if err != nil {
		return nil
	}

	container := service.Container
	if container == nil {
		return gui.createErrorPanel(gui.Tr.NoContainers)
	}

	return gui.openContainerInBrowser(container)
}
