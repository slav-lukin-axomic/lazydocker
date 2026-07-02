package commands

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/sirupsen/logrus"
)

// Image : A docker Image
type Image struct {
	Name          string
	Tag           string
	ID            string
	Image         image.Summary
	Client        DockerClient
	OSCommand     *OSCommand
	Log           *logrus.Entry
	DockerCommand LimitedDockerCommand
}

// Remove removes the image
func (i *Image) Remove(options image.RemoveOptions) error {
	if _, err := i.Client.ImageRemove(context.Background(), i.ID, options); err != nil {
		return err
	}

	return nil
}

// History returns the raw layer history of the image. Formatting and coloring
// live in the presentation layer (see presentation.RenderImageHistory).
func (i *Image) History() ([]image.HistoryResponseItem, error) {
	return i.Client.ImageHistory(context.Background(), i.ID)
}

// RefreshImages returns a slice of docker images
func (c *DockerCommand) RefreshImages() ([]*Image, error) {
	images, err := c.Client.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return nil, err
	}

	ownImages := make([]*Image, len(images))

	for i, img := range images {
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

			for prefix, replacement := range c.Config.UserConfig.Replacements.ImageNamePrefixes {
				if strings.HasPrefix(name, prefix) {
					name = strings.Replace(name, prefix, replacement, 1)
					break
				}
			}
		}

		ownImages[i] = &Image{
			ID:            img.ID,
			Name:          name,
			Tag:           tag,
			Image:         img,
			Client:        c.Client,
			OSCommand:     c.OSCommand,
			Log:           c.Log,
			DockerCommand: c,
		}
	}

	return ownImages, nil
}

// PruneImages prunes images
func (c *DockerCommand) PruneImages() error {
	_, err := c.Client.ImagesPrune(context.Background(), filters.Args{})
	return err
}
