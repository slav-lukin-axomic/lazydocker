package usecase

import (
	"context"

	"github.com/jesseduffield/lazydocker/pkg/domain"
)

// NetworkCommands drives network list/remove/prune over the DockerAPI port. Unlike
// containers it is a single type rather than a Queries/Commands split: the network
// surface is a small CRUD panel with no detail-inspection seam to isolate.
type NetworkCommands struct {
	docker domain.DockerAPI
}

// NewNetworkCommands returns a NetworkCommands backed by the given port.
func NewNetworkCommands(docker domain.DockerAPI) *NetworkCommands {
	return &NetworkCommands{docker: docker}
}

// List returns all networks in the order the port reports them (the panel sorts).
func (n *NetworkCommands) List(ctx context.Context) ([]*domain.Network, error) {
	networks, err := n.docker.ListNetworks(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.Network, len(networks))
	for i := range networks {
		nw := networks[i]
		result[i] = &nw
	}
	return result, nil
}

// Remove removes the network with the given name.
func (n *NetworkCommands) Remove(ctx context.Context, name string) error {
	return n.docker.RemoveNetwork(ctx, name)
}

// Prune removes all unused networks.
func (n *NetworkCommands) Prune(ctx context.Context) error {
	return n.docker.PruneNetworks(ctx)
}
