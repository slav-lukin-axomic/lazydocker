package usecase

import (
	"context"

	"github.com/jesseduffield/lazydocker/pkg/domain"
)

// VolumeCommands drives volume list/remove/prune over the DockerAPI port. Like
// networks it is a single type rather than a Queries/Commands split: the volume
// surface is a small CRUD panel with no detail-inspection seam to isolate.
type VolumeCommands struct {
	docker domain.DockerAPI
}

// NewVolumeCommands returns a VolumeCommands backed by the given port.
func NewVolumeCommands(docker domain.DockerAPI) *VolumeCommands {
	return &VolumeCommands{docker: docker}
}

// List returns all volumes in the order the port reports them (the panel sorts).
func (v *VolumeCommands) List(ctx context.Context) ([]*domain.Volume, error) {
	volumes, err := v.docker.ListVolumes(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.Volume, len(volumes))
	for i := range volumes {
		vol := volumes[i]
		result[i] = &vol
	}
	return result, nil
}

// Remove removes the volume with the given name; force removes it even when in use.
func (v *VolumeCommands) Remove(ctx context.Context, name string, force bool) error {
	return v.docker.RemoveVolume(ctx, name, force)
}

// Prune removes all unused volumes.
func (v *VolumeCommands) Prune(ctx context.Context) error {
	return v.docker.PruneVolumes(ctx)
}
