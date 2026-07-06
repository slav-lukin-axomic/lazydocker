package domain

// Container is the flat, framework-free model of a Docker container. It replaces
// the pre-migration commands.Container, which embedded the SDK's container.Summary
// and container.InspectResponse and carried a live client. Details is nil until
// the container has been inspected (replacing the old DetailsLoaded() sentinel),
// and Stats holds the latest derived sample (history lives with the stats stream).
type Container struct {
	ID              string
	Name            string
	ServiceName     string
	ContainerNumber string
	ProjectName     string
	// OneOff tells us if the container is just a job container or is actually
	// bound to the service.
	OneOff  bool
	Image   string
	Status  Status
	Ports   []Port
	Labels  map[string]string
	Details *ContainerDetails
	Stats   *DerivedStats
}

// ContainerDetails holds the subset of inspect data lazydocker uses. It is
// populated by inspecting a container and is nil on a Container until then.
type ContainerDetails struct {
	Running  bool
	Paused   bool
	ExitCode int
	Health   Health
	// OpenStdin gates whether the container can be attached to.
	OpenStdin bool
}

// DetailsLoaded reports whether the container has been inspected. It preserves
// the call sites of the pre-migration Container.DetailsLoaded(); the underlying
// check is now simply whether Details is non-nil.
func (c *Container) DetailsLoaded() bool {
	return c.Details != nil
}
