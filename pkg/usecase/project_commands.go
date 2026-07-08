package usecase

import (
	"sort"

	"github.com/jesseduffield/lazydocker/pkg/domain"
)

// ProjectCommands drives the Projects panel. Unlike the other panels, projects are
// not fetched from a port: List derives them (deduped, sorted) from the already-loaded
// container list, so it is a pure function taking containers and returning no error.
// Config is the one project-scoped compose op, routed through the ComposeRunner port.
type ProjectCommands struct {
	compose domain.ComposeRunner
}

func NewProjectCommands(compose domain.ComposeRunner) *ProjectCommands {
	return &ProjectCommands{compose: compose}
}

// List returns the unique docker-compose projects derived from container labels,
// sorted by name. This is the former DockerCommand.GetProjectNames, wrapped into
// domain.Project values.
func (p *ProjectCommands) List(containers []*domain.Container) []*domain.Project {
	seen := make(map[string]bool)
	var names []string
	for _, ctr := range containers {
		if ctr.ProjectName != "" && !seen[ctr.ProjectName] {
			seen[ctr.ProjectName] = true
			names = append(names, ctr.ProjectName)
		}
	}
	sort.Strings(names)

	projects := make([]*domain.Project, len(names))
	for i, name := range names {
		projects[i] = &domain.Project{Name: name}
	}
	return projects
}

// Config returns the docker-compose config output for the given project.
func (p *ProjectCommands) Config(project *domain.Project) (string, error) {
	return p.compose.Config(project)
}
