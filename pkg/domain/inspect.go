package domain

// ContainerInspect is the framework-free projection of a container's inspect
// data that the Config and Env detail views render. It replaces the direct reads
// of the SDK's container.InspectResponse the pre-migration presentation code did,
// carrying only the fields those views format. The raw full-inspect YAML dump is
// carried separately by the caller, as it is an opaque, already-marshalled blob.
type ContainerInspect struct {
	ID    string
	Name  string
	Image string
	// Command is Details.Path followed by Details.Args, ready to be space-joined.
	Command []string
	Labels  map[string]string
	Mounts  []Mount
	// Ports is sorted by ContainerPort so the Config view renders deterministically
	// (the SDK carries these in a map, which iterates in random order).
	Ports []PortBinding
	// Env holds the raw KEY=VALUE strings for the Env view to split and colorize.
	Env []string
}

// Mount is a container mount point, mirroring the fields the Config view reads
// from the SDK's container.MountPoint (Type is the SDK's mount.Type as a string).
type Mount struct {
	Type        string
	Name        string
	Source      string
	Destination string
}

// PortBinding groups a container port with the host ports it is published to,
// mirroring one entry of the SDK's nat.PortMap. ContainerPort is the map key
// (e.g. "80/tcp"); HostPorts holds the host-side port for each binding, in SDK
// order. A present key with no bindings yields an empty HostPorts slice.
type PortBinding struct {
	ContainerPort string
	HostPorts     []string
}
