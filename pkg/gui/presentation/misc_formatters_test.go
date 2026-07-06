package presentation

import (
	"strings"
	"testing"

	"github.com/jesseduffield/lazydocker/pkg/commands"
	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/gui/types"
)

func TestGetImageDisplayStrings(t *testing.T) {
	img := &domain.Image{
		Name: "nginx",
		Tag:  "1.25",
		Size: 142 * 1000 * 1000,
	}
	got := strings.Join(GetImageDisplayStrings(img), colSep)
	assertGolden(t, "image", got)
}

func TestGetVolumeDisplayStrings(t *testing.T) {
	vol := &domain.Volume{
		Name:   "app-data",
		Driver: "local",
	}
	got := strings.Join(GetVolumeDisplayStrings(vol), colSep)
	assertGolden(t, "volume", got)
}

func TestGetNetworkDisplayStrings(t *testing.T) {
	nw := &domain.Network{
		Name:   "bridge",
		Driver: "bridge",
	}
	got := strings.Join(GetNetworkDisplayStrings(nw), colSep)
	assertGolden(t, "network", got)
}

func TestGetProjectDisplayStrings(t *testing.T) {
	project := &commands.Project{Name: "my-compose-project"}
	got := strings.Join(GetProjectDisplayStrings(project), colSep)
	assertGolden(t, "project", got)
}

func TestGetMenuItemDisplayStrings(t *testing.T) {
	item := &types.MenuItem{LabelColumns: []string{"Restart", "restart the container"}}
	got := strings.Join(GetMenuItemDisplayStrings(item), colSep)
	assertGolden(t, "menu_item", got)
}
