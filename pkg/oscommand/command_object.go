package oscommand

import (
	"fmt"

	"github.com/imdario/mergo"
	"github.com/jesseduffield/lazydocker/pkg/domain"
)

// CommandObject is what we pass to our template resolvers when we are running a custom command. We do not guarantee that all fields will be populated: just the ones that make sense for the current context
type CommandObject struct {
	DockerCompose string
	Service       *domain.Service
	Container     *domain.Container
	Image         *domain.Image
	Volume        *domain.Volume
	Network       *domain.Network
	Project       *domain.Project
}

// NewCommandObject merges obj onto a default carrying the docker-compose base
// command, appending "-p <name>" when scoped to a specific service or project.
func NewCommandObject(dockerComposeTemplate string, obj CommandObject) CommandObject {
	defaultObj := CommandObject{DockerCompose: dockerComposeTemplate}
	_ = mergo.Merge(&defaultObj, obj)
	if obj.Service != nil && obj.Service.ProjectName != "" {
		defaultObj.DockerCompose = fmt.Sprintf("%s -p %s", defaultObj.DockerCompose, obj.Service.ProjectName)
	} else if obj.Project != nil && obj.Project.Name != "" {
		defaultObj.DockerCompose = fmt.Sprintf("%s -p %s", defaultObj.DockerCompose, obj.Project.Name)
	}
	return defaultObj
}
