package presentation

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/utils"
	"github.com/samber/lo"
)

// RenderContainerEnv renders a container's environment as a two-column table of
// KEY: / value rows. The formatting is carried over verbatim from the
// pre-migration Gui.containerEnv.
func RenderContainerEnv(env []string) (string, error) {
	envVarsList := lo.Map(env, func(envVar string, _ int) []string {
		splitEnv := strings.SplitN(envVar, "=", 2)
		key := splitEnv[0]
		value := ""
		if len(splitEnv) > 1 {
			value = splitEnv[1]
		}
		return []string{
			utils.ColoredString(key+":", color.FgGreen),
			utils.ColoredString(value, color.FgYellow),
		}
	})

	return utils.RenderTable(envVarsList)
}

// RenderContainerConfig renders a container's config summary followed by the full
// inspect YAML dump. The assembly is carried over verbatim from the pre-migration
// Gui.containerConfigStr, reading from the framework-free inspect projection and
// the already-marshalled fullDetailsYAML rather than the SDK types.
func RenderContainerConfig(inspect domain.ContainerInspect, fullDetailsYAML string) string {
	padding := 10
	output := ""
	output += utils.WithPadding("ID: ", padding) + inspect.ID + "\n"
	output += utils.WithPadding("Name: ", padding) + inspect.Name + "\n"
	output += utils.WithPadding("Image: ", padding) + inspect.Image + "\n"
	output += utils.WithPadding("Command: ", padding) + strings.Join(inspect.Command, " ") + "\n"
	output += utils.WithPadding("Labels: ", padding) + utils.FormatMap(padding, inspect.Labels)
	output += "\n"

	output += utils.WithPadding("Mounts: ", padding)
	if len(inspect.Mounts) > 0 {
		output += "\n"
		for _, mount := range inspect.Mounts {
			if mount.Type == "volume" {
				output += fmt.Sprintf("%s%s %s\n", strings.Repeat(" ", padding), utils.ColoredString(mount.Type+":", color.FgYellow), mount.Name)
			} else {
				output += fmt.Sprintf("%s%s %s:%s\n", strings.Repeat(" ", padding), utils.ColoredString(mount.Type+":", color.FgYellow), mount.Source, mount.Destination)
			}
		}
	} else {
		output += "none\n"
	}

	output += utils.WithPadding("Ports: ", padding)
	if len(inspect.Ports) > 0 {
		output += "\n"
		for _, portBinding := range inspect.Ports {
			for _, hostPort := range portBinding.HostPorts {
				output += fmt.Sprintf("%s%s %s\n", strings.Repeat(" ", padding), utils.ColoredString(hostPort+":", color.FgYellow), portBinding.ContainerPort)
			}
		}
	} else {
		output += "none\n"
	}

	output += fmt.Sprintf("\nFull details:\n\n%s", utils.ColoredYamlString(fullDetailsYAML))

	return output
}
