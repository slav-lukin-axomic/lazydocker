package gui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazydocker/pkg/commands"
	"github.com/jesseduffield/lazydocker/pkg/config"
	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/gui/panels"
	"github.com/jesseduffield/lazydocker/pkg/gui/presentation"
	"github.com/jesseduffield/lazydocker/pkg/gui/types"
	"github.com/jesseduffield/lazydocker/pkg/tasks"
	"github.com/jesseduffield/lazydocker/pkg/utils"
	"github.com/samber/lo"
)

func (gui *Gui) getImagesPanel() *panels.SideListPanel[*domain.Image] {
	noneLabel := "<none>"

	return &panels.SideListPanel[*domain.Image]{
		ContextState: &panels.ContextState[*domain.Image]{
			GetMainTabs: func() []panels.MainTab[*domain.Image] {
				return []panels.MainTab[*domain.Image]{
					{
						Key:    "config",
						Title:  gui.Tr.ConfigTitle,
						Render: gui.renderImageConfigTask,
					},
				}
			},
			GetItemContextCacheKey: func(image *domain.Image) string {
				return "images-" + image.ID
			},
		},
		ListPanel: panels.ListPanel[*domain.Image]{
			List: panels.NewFilteredList[*domain.Image](),
			View: gui.Views.Images,
		},
		NoItemsMessage: gui.Tr.NoImages,
		Gui:            gui.intoInterface(),
		Sort: func(a *domain.Image, b *domain.Image) bool {
			if a.Name == noneLabel && b.Name != noneLabel {
				return false
			}

			if a.Name != noneLabel && b.Name == noneLabel {
				return true
			}

			if a.Name != b.Name {
				return a.Name < b.Name
			}

			if a.Tag != b.Tag {
				return a.Tag < b.Tag
			}

			return a.ID < b.ID
		},
		GetTableCells: presentation.GetImageDisplayStrings,
	}
}

func (gui *Gui) renderImageConfigTask(image *domain.Image) tasks.TaskFunc {
	return gui.NewRenderStringTask(RenderStringTaskOpts{
		GetStrContent: func() string { return gui.imageConfigStr(image) },
		Autoscroll:    false,
		Wrap:          false, // don't care what your config is this page is ugly without wrapping
	})
}

func (gui *Gui) imageConfigStr(image *domain.Image) string {
	padding := 10
	output := ""
	output += utils.WithPadding("Name: ", padding) + image.Name + "\n"
	output += utils.WithPadding("ID: ", padding) + image.ID + "\n"
	output += utils.WithPadding("Tags: ", padding) + utils.ColoredString(strings.Join(image.RepoTags, ", "), color.FgGreen) + "\n"
	output += utils.WithPadding("Size: ", padding) + utils.FormatDecimalBytes(int(image.Size)) + "\n"
	output += utils.WithPadding("Created: ", padding) + fmt.Sprintf("%v", time.Unix(image.Created, 0).Format(time.RFC1123)) + "\n"

	history, err := gui.ImageCommands.History(context.Background(), image.ID)
	if err != nil {
		gui.Log.Error(err)
	}

	renderedHistory, err := presentation.RenderImageHistory(history)
	if err != nil {
		gui.Log.Error(err)
	}

	output += "\n\n" + renderedHistory

	return output
}

func (gui *Gui) reloadImages() error {
	if err := gui.refreshStateImages(); err != nil {
		return err
	}

	return gui.Panels.Images.RerenderList()
}

func (gui *Gui) refreshStateImages() error {
	images, err := gui.ImageCommands.List(context.Background())
	if err != nil {
		return err
	}

	gui.Panels.Images.SetItems(images)

	return nil
}

func (gui *Gui) FilterString(view *gocui.View) string {
	if gui.State.Filter.panel != nil && gui.State.Filter.panel.GetView() != view {
		return ""
	}

	return gui.State.Filter.needle
}

func (gui *Gui) handleImagesRemoveMenu(g *gocui.Gui, v *gocui.View) error {
	type removeImageOption struct {
		description   string
		command       string
		force         bool
		pruneChildren bool
	}

	img, err := gui.Panels.Images.GetSelectedItem()
	if err != nil {
		return nil
	}

	shortSha := img.ID[7:17]

	// TODO: have a way of toggling in a menu instead of showing each permutation as a separate menu item
	options := []*removeImageOption{
		{
			description:   gui.Tr.Remove,
			command:       "docker image rm " + shortSha,
			force:         false,
			pruneChildren: true,
		},
		{
			description:   gui.Tr.RemoveWithoutPrune,
			command:       "docker image rm --no-prune " + shortSha,
			force:         false,
			pruneChildren: false,
		},
		{
			description:   gui.Tr.RemoveWithForce,
			command:       "docker image rm --force " + shortSha,
			force:         true,
			pruneChildren: true,
		},
		{
			description:   gui.Tr.RemoveWithoutPruneWithForce,
			command:       "docker image rm --no-prune --force " + shortSha,
			force:         true,
			pruneChildren: false,
		},
	}

	menuItems := lo.Map(options, func(option *removeImageOption, _ int) *types.MenuItem {
		return &types.MenuItem{
			LabelColumns: []string{
				option.description,
				color.New(color.FgRed).Sprint(option.command),
			},
			OnPress: func() error {
				if err := gui.ImageCommands.Remove(context.Background(), img.ID, option.force, option.pruneChildren); err != nil {
					return gui.createErrorPanel(err.Error())
				}

				return nil
			},
		}
	})

	return gui.Menu(CreateMenuOptions{
		Title: "",
		Items: menuItems,
	})
}

func (gui *Gui) handlePruneImages() error {
	return gui.createConfirmationPanel(gui.Tr.Confirm, gui.Tr.ConfirmPruneImages, func(g *gocui.Gui, v *gocui.View) error {
		return gui.WithWaitingStatus(gui.Tr.PruningStatus, func() error {
			err := gui.ImageCommands.Prune(context.Background())
			if err != nil {
				return gui.createErrorPanel(err.Error())
			}
			return gui.reloadImages()
		})
	}, nil)
}

func (gui *Gui) handleImagesCustomCommand(g *gocui.Gui, v *gocui.View) error {
	img, err := gui.Panels.Images.GetSelectedItem()
	if err != nil {
		return nil
	}

	commandObject := gui.DockerCommand.NewCommandObject(commands.CommandObject{
		Image: img,
	})

	customCommands := gui.Config.UserConfig.CustomCommands.Images

	return gui.createCustomCommandMenu(customCommands, commandObject)
}

func (gui *Gui) handleImagesBulkCommand(g *gocui.Gui, v *gocui.View) error {
	baseBulkCommands := []config.CustomCommand{
		{
			Name:             gui.Tr.PruneImages,
			InternalFunction: gui.handlePruneImages,
		},
	}

	bulkCommands := append(baseBulkCommands, gui.Config.UserConfig.BulkCommands.Images...)
	commandObject := gui.DockerCommand.NewCommandObject(commands.CommandObject{})

	return gui.createBulkCommandMenu(bulkCommands, commandObject)
}
