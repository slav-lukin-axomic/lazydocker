package domain

// Image is the flat, framework-free model of a Docker image. It replaces the
// pre-migration commands.Image, which embedded the SDK's image.Summary and carried
// a live client. Name and Tag are not populated by the adapter: they are derived
// from RepoTags by usecase.ImageCommands (applying the configured name-prefix
// replacements), the same pattern as InspectContainerVerbose leaving identity
// fields for the caller to supply.
type Image struct {
	ID       string
	Name     string
	Tag      string
	RepoTags []string
	Size     int64
	Created  int64
}

// HistoryLayer is a single layer in an image's build history, holding the subset
// of the SDK's image.HistoryResponseItem the history view renders.
type HistoryLayer struct {
	ID        string
	Tags      []string
	CreatedBy string
	Size      int64
}
