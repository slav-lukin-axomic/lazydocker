package presentation

import (
	"strings"
	"testing"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/jesseduffield/lazydocker/pkg/commands"
	"github.com/jesseduffield/lazydocker/pkg/gui/types"
)

func TestGetImageDisplayStrings(t *testing.T) {
	img := &commands.Image{
		Name:  "nginx",
		Tag:   "1.25",
		Image: image.Summary{Size: 142 * 1000 * 1000},
	}
	got := strings.Join(GetImageDisplayStrings(img), colSep)
	assertGolden(t, "image", got)
}

func TestGetVolumeDisplayStrings(t *testing.T) {
	vol := &commands.Volume{
		Name:   "app-data",
		Volume: &volume.Volume{Driver: "local"},
	}
	got := strings.Join(GetVolumeDisplayStrings(vol), colSep)
	assertGolden(t, "volume", got)
}

func TestGetNetworkDisplayStrings(t *testing.T) {
	nw := &commands.Network{
		Name:    "bridge",
		Network: network.Inspect{Driver: "bridge"},
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
