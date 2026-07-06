package gui

import (
	"context"
	"strconv"

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

func (gui *Gui) getNetworksPanel() *panels.SideListPanel[*domain.Network] {
	return &panels.SideListPanel[*domain.Network]{
		ContextState: &panels.ContextState[*domain.Network]{
			GetMainTabs: func() []panels.MainTab[*domain.Network] {
				return []panels.MainTab[*domain.Network]{
					{
						Key:    "config",
						Title:  gui.Tr.ConfigTitle,
						Render: gui.renderNetworkConfig,
					},
				}
			},
			GetItemContextCacheKey: func(network *domain.Network) string {
				return "networks-" + network.Name
			},
		},
		ListPanel: panels.ListPanel[*domain.Network]{
			List: panels.NewFilteredList[*domain.Network](),
			View: gui.Views.Networks,
		},
		NoItemsMessage: gui.Tr.NoNetworks,
		Gui:            gui.intoInterface(),
		// we're sorting these networks based on whether they have labels defined,
		// because those are the ones you typically care about.
		// Within that, we also sort them alphabetically
		Sort: func(a *domain.Network, b *domain.Network) bool {
			return a.Name < b.Name
		},
		GetTableCells: presentation.GetNetworkDisplayStrings,
	}
}

func (gui *Gui) renderNetworkConfig(network *domain.Network) tasks.TaskFunc {
	return gui.NewSimpleRenderStringTask(func() string { return gui.networkConfigStr(network) })
}

func (gui *Gui) networkConfigStr(network *domain.Network) string {
	padding := 15
	output := ""
	output += utils.WithPadding("ID: ", padding) + network.ID + "\n"
	output += utils.WithPadding("Name: ", padding) + network.Name + "\n"
	output += utils.WithPadding("Driver: ", padding) + network.Driver + "\n"
	output += utils.WithPadding("Scope: ", padding) + network.Scope + "\n"
	output += utils.WithPadding("EnabledIPV6: ", padding) + strconv.FormatBool(network.EnableIPv6) + "\n"
	output += utils.WithPadding("Internal: ", padding) + strconv.FormatBool(network.Internal) + "\n"
	output += utils.WithPadding("Attachable: ", padding) + strconv.FormatBool(network.Attachable) + "\n"
	output += utils.WithPadding("Ingress: ", padding) + strconv.FormatBool(network.Ingress) + "\n"

	output += utils.WithPadding("Containers: ", padding)
	if len(network.Containers) > 0 {
		output += "\n"
		for _, v := range network.Containers {
			output += utils.FormatMapItem(padding, v.Name, v.EndpointID)
		}
	} else {
		output += "none\n"
	}

	output += "\n"
	output += utils.WithPadding("Labels: ", padding) + utils.FormatMap(padding, network.Labels) + "\n"
	output += utils.WithPadding("Options: ", padding) + utils.FormatMap(padding, network.Options)

	return output
}

func (gui *Gui) reloadNetworks() error {
	if err := gui.refreshStateNetworks(); err != nil {
		return err
	}

	return gui.Panels.Networks.RerenderList()
}

func (gui *Gui) refreshStateNetworks() error {
	networks, err := gui.NetworkCommands.List(context.Background())
	if err != nil {
		return err
	}

	gui.Panels.Networks.SetItems(networks)

	return nil
}

func (gui *Gui) handleNetworksRemoveMenu(g *gocui.Gui, v *gocui.View) error {
	network, err := gui.Panels.Networks.GetSelectedItem()
	if err != nil {
		return nil
	}

	type removeNetworkOption struct {
		description string
		command     string
	}

	options := []*removeNetworkOption{
		{
			description: gui.Tr.Remove,
			command:     utils.WithShortSha("docker network rm " + network.Name),
		},
	}

	menuItems := lo.Map(options, func(option *removeNetworkOption, _ int) *types.MenuItem {
		return &types.MenuItem{
			LabelColumns: []string{option.description, color.New(color.FgRed).Sprint(option.command)},
			OnPress: func() error {
				return gui.WithWaitingStatus(gui.Tr.RemovingStatus, func() error {
					if err := gui.NetworkCommands.Remove(context.Background(), network.Name); err != nil {
						return gui.createErrorPanel(err.Error())
					}
					return nil
				})
			},
		}
	})

	return gui.Menu(CreateMenuOptions{
		Title: "",
		Items: menuItems,
	})
}

func (gui *Gui) handlePruneNetworks() error {
	return gui.createConfirmationPanel(gui.Tr.Confirm, gui.Tr.ConfirmPruneNetworks, func(g *gocui.Gui, v *gocui.View) error {
		return gui.WithWaitingStatus(gui.Tr.PruningStatus, func() error {
			err := gui.NetworkCommands.Prune(context.Background())
			if err != nil {
				return gui.createErrorPanel(err.Error())
			}
			return nil
		})
	}, nil)
}

func (gui *Gui) handleNetworksCustomCommand(g *gocui.Gui, v *gocui.View) error {
	network, err := gui.Panels.Networks.GetSelectedItem()
	if err != nil {
		return nil
	}

	commandObject := gui.DockerCommand.NewCommandObject(commands.CommandObject{
		Network: network,
	})

	customCommands := gui.Config.UserConfig.CustomCommands.Networks

	return gui.createCustomCommandMenu(customCommands, commandObject)
}

func (gui *Gui) handleNetworksBulkCommand(g *gocui.Gui, v *gocui.View) error {
	baseBulkCommands := []config.CustomCommand{
		{
			Name:             gui.Tr.PruneNetworks,
			InternalFunction: gui.handlePruneNetworks,
		},
	}

	bulkCommands := append(baseBulkCommands, gui.Config.UserConfig.BulkCommands.Networks...)
	commandObject := gui.DockerCommand.NewCommandObject(commands.CommandObject{})

	return gui.createBulkCommandMenu(bulkCommands, commandObject)
}
