package presentation

import "github.com/jesseduffield/lazydocker/pkg/domain"

func GetProjectDisplayStrings(project *domain.Project) []string {
	return []string{project.Name}
}
