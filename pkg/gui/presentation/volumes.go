package presentation

import "github.com/jesseduffield/lazydocker/pkg/domain"

func GetVolumeDisplayStrings(volume *domain.Volume) []string {
	return []string{volume.Driver, volume.Name}
}
