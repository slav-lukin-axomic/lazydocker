package gui

import (
	"bytes"
	"context"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/gui/panels"
	"github.com/jesseduffield/lazydocker/pkg/gui/presentation"
	"github.com/jesseduffield/lazydocker/pkg/oscommand"
	"github.com/jesseduffield/lazydocker/pkg/tasks"
	"github.com/jesseduffield/lazydocker/pkg/utils"
	"github.com/jesseduffield/yaml"
)

func (gui *Gui) getProjectPanel() *panels.SideListPanel[*domain.Project] {
	return &panels.SideListPanel[*domain.Project]{
		ContextState: &panels.ContextState[*domain.Project]{
			GetMainTabs: func() []panels.MainTab[*domain.Project] {
				return []panels.MainTab[*domain.Project]{
					{
						Key:    "logs",
						Title:  gui.Tr.LogsTitle,
						Render: gui.renderAllLogs,
					},
					{
						Key:    "config",
						Title:  gui.Tr.DockerComposeConfigTitle,
						Render: gui.renderDockerComposeConfig,
					},
					{
						Key:    "credits",
						Title:  gui.Tr.CreditsTitle,
						Render: gui.renderCredits,
					},
				}
			},
			GetItemContextCacheKey: func(project *domain.Project) string {
				return "projects-" + project.Name
			},
		},

		ListPanel: panels.ListPanel[*domain.Project]{
			List: panels.NewFilteredList[*domain.Project](),
			View: gui.Views.Project,
		},
		NoItemsMessage: "",
		Gui:            gui.intoInterface(),

		Sort: func(a *domain.Project, b *domain.Project) bool {
			return a.Name < b.Name
		},
		GetTableCells: presentation.GetProjectDisplayStrings,
		OnSelect: func(project *domain.Project) error {
			// When a different project is selected, re-filter services and
			// containers to show only those belonging to the selected project.
			return gui.renderContainersAndServices()
		},
		Hide: func() bool {
			return !gui.DockerCommand.IsProjectScoped()
		},
	}
}

func (gui *Gui) refreshProject() error {
	projects := gui.getDiscoveredProjects()

	// Preserve the current selection across refreshes. On the first refresh,
	// select the project specified via -p flag, or fall back to the local project.
	selectedName := gui.getSelectedProjectName()
	if selectedName == "" {
		if gui.Config.ProjectName != "" {
			selectedName = gui.Config.ProjectName
		} else {
			selectedName = gui.DockerCommand.LocalProjectName
		}
	}

	gui.Panels.Projects.SetItems(projects)

	if selectedName != "" {
		for i, p := range gui.Panels.Projects.List.GetItems() {
			if p.Name == selectedName {
				gui.Panels.Projects.SetSelectedLineIdx(i)
				gui.Panels.Projects.Refocus()
				break
			}
		}
	}

	return gui.Panels.Projects.RerenderList()
}

// getDiscoveredProjects returns all docker compose projects by examining container labels.
// The local project (from docker-compose.yml in the current directory, or from -p) is
// included even when it has no running containers, so the user always sees the project
// they explicitly scoped to.
func (gui *Gui) getDiscoveredProjects() []*domain.Project {
	containers := gui.Panels.Containers.List.GetAllItems()
	projects := gui.ProjectCommands.List(containers)

	// If we're scoped to a project but it has no running containers, still
	// include it. We don't fall back to the directory name here to avoid
	// briefly flashing the wrong project name on startup.
	localName := gui.DockerCommand.LocalProjectName

	if gui.DockerCommand.IsProjectScoped() && localName != "" {
		found := false
		for _, p := range projects {
			if p.Name == localName {
				found = true
				break
			}
		}
		if !found {
			projects = append([]*domain.Project{{Name: localName}}, projects...)
		}
	}

	return projects
}

// getSelectedProjectName returns the name of the currently selected project,
// or empty string if none is selected.
func (gui *Gui) getSelectedProjectName() string {
	project, err := gui.Panels.Projects.GetSelectedItem()
	if err != nil {
		return ""
	}
	return project.Name
}

func (gui *Gui) renderCredits(_project *domain.Project) tasks.TaskFunc {
	return gui.NewSimpleRenderStringTask(func() string { return gui.creditsStr() })
}

func (gui *Gui) creditsStr() string {
	var configBuf bytes.Buffer
	_ = yaml.NewEncoder(&configBuf, yaml.IncludeOmitted).Encode(gui.Config.UserConfig)

	return strings.Join(
		[]string{
			lazydockerTitle(),
			"Copyright (c) 2019 Jesse Duffield",
			"Keybindings: https://github.com/jesseduffield/lazydocker/blob/master/docs/keybindings",
			"Config Options: https://github.com/jesseduffield/lazydocker/blob/master/docs/Config.md",
			"Raise an Issue: https://github.com/jesseduffield/lazydocker/issues",
			utils.ColoredString("Buy Jesse a coffee: https://github.com/sponsors/jesseduffield", color.FgMagenta), // caffeine ain't free
			"Here's your lazydocker config when merged in with the defaults (you can open your config by pressing 'o'):",
			utils.ColoredYamlString(configBuf.String()),
		}, "\n\n")
}

func (gui *Gui) renderAllLogs(project *domain.Project) tasks.TaskFunc {
	return gui.NewTask(TaskOpts{
		Autoscroll: true,
		Wrap:       gui.Config.UserConfig.Gui.WrapMainPanel,
		Func: func(ctx context.Context) {
			gui.clearMainView()

			cmd := gui.OSCommand.RunCustomCommand(
				utils.ApplyTemplate(
					gui.Config.UserConfig.CommandTemplates.AllLogs,
					gui.DockerCommand.NewCommandObject(oscommand.CommandObject{Project: project}),
				),
			)

			cmd.Stdout = gui.Views.Main
			cmd.Stderr = gui.Views.Main

			gui.OSCommand.PrepareForChildren(cmd)
			_ = cmd.Start()

			go func() {
				<-ctx.Done()
				if err := gui.OSCommand.Kill(cmd); err != nil {
					gui.Log.Error(err)
				}
			}()

			_ = cmd.Wait()
		},
	})
}

func (gui *Gui) renderDockerComposeConfig(project *domain.Project) tasks.TaskFunc {
	if !gui.DockerCommand.InDockerComposeProject {
		return gui.NewSimpleRenderStringTask(func() string {
			return "Compose config is only available when launched from a docker-compose project directory"
		})
	}
	if project != nil && project.Name != gui.DockerCommand.LocalProjectName {
		return gui.NewSimpleRenderStringTask(func() string {
			return "Compose config is not available for non-local projects"
		})
	}
	return gui.NewSimpleRenderStringTask(func() string {
		out, err := gui.ProjectCommands.Config(project)
		if err != nil {
			out = err.Error()
		}
		return utils.ColoredYamlString(out)
	})
}

func (gui *Gui) handleOpenConfig(g *gocui.Gui, v *gocui.View) error {
	return gui.openFile(gui.Config.ConfigFilename())
}

func (gui *Gui) handleEditConfig(g *gocui.Gui, v *gocui.View) error {
	return gui.editFile(gui.Config.ConfigFilename())
}

func lazydockerTitle() string {
	return `
   _                     _            _
  | |                   | |          | |
  | | __ _ _____   _  __| | ___   ___| | _____ _ __
  | |/ _` + "`" + ` |_  / | | |/ _` + "`" + ` |/ _ \ / __| |/ / _ \ '__|
  | | (_| |/ /| |_| | (_| | (_) | (__|   <  __/ |
  |_|\__,_/___|\__, |\__,_|\___/ \___|_|\_\___|_|
                __/ |
               |___/
`
}

// handleViewAllLogs switches to a subprocess viewing all the logs from docker-compose
func (gui *Gui) handleViewAllLogs(g *gocui.Gui, v *gocui.View) error {
	project, _ := gui.Panels.Projects.GetSelectedItem()
	c, err := gui.viewAllLogs(project)
	if err != nil {
		return gui.createErrorPanel(err.Error())
	}

	return gui.runSubprocess(c)
}

// viewAllLogs builds the subprocess command that streams all docker-compose logs
// for the given project. It renders the ViewAllLogs template against a
// project-scoped CommandObject, mirroring the service-panel viewServiceLogs helper.
func (gui *Gui) viewAllLogs(project *domain.Project) (*exec.Cmd, error) {
	command := utils.ApplyTemplate(
		gui.Config.UserConfig.CommandTemplates.ViewAllLogs,
		gui.DockerCommand.NewCommandObject(oscommand.CommandObject{Project: project}),
	)
	cmd := gui.OSCommand.ExecutableFromString(command)
	gui.OSCommand.PrepareForChildren(cmd)
	return cmd, nil
}
