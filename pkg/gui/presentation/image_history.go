package presentation

import (
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/fatih/color"
	"github.com/jesseduffield/lazydocker/pkg/utils"
	"github.com/samber/lo"
)

// RenderImageHistory renders the layer history of an image as a colored table.
func RenderImageHistory(history []image.HistoryResponseItem) (string, error) {
	tableBody := lo.Map(history, func(layer image.HistoryResponseItem, _ int) []string {
		return getHistoryResponseItemDisplayStrings(layer)
	})

	headers := [][]string{{"ID", "TAG", "SIZE", "COMMAND"}}
	table := append(headers, tableBody...)

	return utils.RenderTable(table)
}

func getHistoryResponseItemDisplayStrings(layer image.HistoryResponseItem) []string {
	tag := ""
	if len(layer.Tags) > 0 {
		tag = layer.Tags[0]
	}

	id := strings.TrimPrefix(layer.ID, "sha256:")
	if len(id) > 10 {
		id = id[0:10]
	}
	idColor := color.FgWhite
	if id == "<missing>" {
		idColor = color.FgBlue
	}

	dockerFileCommandPrefix := "/bin/sh -c #(nop) "
	createdBy := layer.CreatedBy
	if strings.Contains(layer.CreatedBy, dockerFileCommandPrefix) {
		createdBy = strings.Trim(strings.TrimPrefix(layer.CreatedBy, dockerFileCommandPrefix), " ")
		split := strings.Split(createdBy, " ")
		createdBy = utils.ColoredString(split[0], color.FgYellow) + " " + strings.Join(split[1:], " ")
	}

	createdBy = strings.Replace(createdBy, "\t", " ", -1)

	size := utils.FormatBinaryBytes(int(layer.Size))
	sizeColor := color.FgWhite
	if size == "0B" {
		sizeColor = color.FgBlue
	}

	return []string{
		utils.ColoredString(id, idColor),
		utils.ColoredString(tag, color.FgGreen),
		utils.ColoredString(size, sizeColor),
		createdBy,
	}
}
