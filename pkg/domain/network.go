package domain

// Network is the flat, framework-free model of a Docker network. It replaces the
// pre-migration commands.Network, which embedded the SDK's network.Inspect and
// carried a live client. Containers is a map (keyed by container ID) rather than a
// slice to preserve the config view's iteration semantics.
type Network struct {
	ID         string
	Name       string
	Driver     string
	Scope      string
	EnableIPv6 bool
	Internal   bool
	Attachable bool
	Ingress    bool
	Containers map[string]NetworkEndpoint
	Labels     map[string]string
	Options    map[string]string
}

// NetworkEndpoint is a container endpoint attached to a network, holding the
// subset of the SDK's network.EndpointResource the config view renders.
type NetworkEndpoint struct {
	Name       string
	EndpointID string
}
