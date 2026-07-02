package presentation

import (
	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/utils"
)

// RenderContainerTop renders a container's process listing as a table, with the
// titles as the header row followed by one row per process. The formatting is
// carried over verbatim from the pre-migration commands.Container.RenderTop.
func RenderContainerTop(result domain.TopResult) (string, error) {
	return utils.RenderTable(append([][]string{result.Titles}, result.Processes...))
}
