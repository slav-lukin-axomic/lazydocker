package presentation

import "github.com/jesseduffield/lazydocker/pkg/domain"

func GetNetworkDisplayStrings(network *domain.Network) []string {
	return []string{network.Driver, network.Name}
}
