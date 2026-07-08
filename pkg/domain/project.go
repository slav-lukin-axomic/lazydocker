package domain

// Project is the flat, framework-free model of a Docker Compose project. It
// replaces the pre-migration commands.Project. Projects are derived from
// container labels (see usecase.ProjectCommands.List), not fetched from a port.
type Project struct {
	Name string
}
