package usecase

import (
	"context"
	"strings"

	"github.com/jesseduffield/lazydocker/pkg/domain"
)

// ImageCommands drives image list/history/remove/prune over the DockerAPI port.
// Like networks and volumes it is a single type rather than a Queries/Commands
// split: the image surface is a small CRUD panel with no detail-inspection seam.
// It owns the RepoTags → Name/Tag derivation (including the configured name-prefix
// replacements), which the adapter deliberately leaves to the caller.
type ImageCommands struct {
	docker           domain.DockerAPI
	nameReplacements map[string]string
}

// NewImageCommands returns an ImageCommands backed by the given port. The
// nameReplacements map (config.Replacements.ImageNamePrefixes) rewrites image name
// prefixes during List.
func NewImageCommands(docker domain.DockerAPI, nameReplacements map[string]string) *ImageCommands {
	return &ImageCommands{docker: docker, nameReplacements: nameReplacements}
}

// List returns all images in the order the port reports them (the panel sorts),
// deriving each image's Name and Tag from its first RepoTag and applying the
// configured name-prefix replacements.
func (i *ImageCommands) List(ctx context.Context) ([]*domain.Image, error) {
	images, err := i.docker.ListImages(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Image, len(images))
	for idx := range images {
		img := images[idx]

		firstTag := ""
		tags := img.RepoTags
		if len(tags) > 0 {
			firstTag = tags[0]
		}

		nameParts := strings.Split(firstTag, ":")
		tag := ""
		name := "none"
		if len(nameParts) > 1 {
			tag = nameParts[len(nameParts)-1]
			name = strings.Join(nameParts[:len(nameParts)-1], ":")

			for prefix, replacement := range i.nameReplacements {
				if strings.HasPrefix(name, prefix) {
					name = strings.Replace(name, prefix, replacement, 1)
					break
				}
			}
		}

		img.Name = name
		img.Tag = tag
		result[idx] = &img
	}
	return result, nil
}

// History returns the build-layer history of the image with the given id.
func (i *ImageCommands) History(ctx context.Context, id string) ([]domain.HistoryLayer, error) {
	return i.docker.ImageHistory(ctx, id)
}

// Remove removes the image with the given id; force removes it even when
// referenced, and pruneChildren removes untagged parents.
func (i *ImageCommands) Remove(ctx context.Context, id string, force, pruneChildren bool) error {
	return i.docker.RemoveImage(ctx, id, force, pruneChildren)
}

// Prune removes all dangling images.
func (i *ImageCommands) Prune(ctx context.Context) error {
	return i.docker.PruneImages(ctx)
}
