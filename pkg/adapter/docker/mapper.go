// Package docker is the driven adapter over the Docker Engine SDK. It is the only
// package outside vendored code allowed to import github.com/docker/docker, and it
// translates SDK types to and from the framework-free pkg/domain types (the
// anti-corruption layer).
package docker

import (
	"sort"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/jesseduffield/lazydocker/pkg/domain"
)

// mapContainerSummary maps an SDK container.Summary to a domain.Container. It
// replicates the per-container field derivation the pre-migration
// DockerCommand.GetContainers performed (name/label parsing), so a mapped
// Container is byte-identical to what that code produced. Details is left nil;
// inspecting the container populates it separately.
func mapContainerSummary(summary container.Summary) domain.Container {
	c := domain.Container{
		ID:     summary.ID,
		Image:  summary.Image,
		Status: domain.ParseStatus(summary.State),
		Ports:  mapPorts(summary.Ports),
		Labels: summary.Labels,
	}

	// Name resolution mirrors GetContainers: a "name" label wins, else the first
	// entry in Names with the leading slash trimmed, else the container ID.
	if name, ok := summary.Labels["name"]; ok {
		c.Name = name
	} else if len(summary.Names) > 0 {
		c.Name = strings.TrimLeft(summary.Names[0], "/")
	} else {
		c.Name = summary.ID
	}

	c.ServiceName = summary.Labels["com.docker.compose.service"]
	c.ProjectName = summary.Labels["com.docker.compose.project"]
	c.ContainerNumber = summary.Labels["com.docker.compose.container"]
	c.OneOff = summary.Labels["com.docker.compose.oneoff"] == "True"

	return c
}

// mapPorts maps a slice of SDK container.Port to domain.Port (Proto is the SDK's
// Type field). A nil input yields a nil slice.
func mapPorts(ports []container.Port) []domain.Port {
	if ports == nil {
		return nil
	}
	out := make([]domain.Port, len(ports))
	for i, p := range ports {
		out[i] = domain.Port{
			IP:          p.IP,
			PublicPort:  p.PublicPort,
			PrivatePort: p.PrivatePort,
			Proto:       p.Type,
		}
	}
	return out
}

// mapInspectResponse maps an SDK container.InspectResponse to domain
// ContainerDetails. Health is derived nil-safely: an absent State/Health maps to
// HealthNone.
func mapInspectResponse(resp container.InspectResponse) domain.ContainerDetails {
	details := domain.ContainerDetails{}

	if resp.ContainerJSONBase != nil && resp.State != nil {
		details.Running = resp.State.Running
		details.Paused = resp.State.Paused
		details.ExitCode = resp.State.ExitCode
		if resp.State.Health != nil {
			details.Health = domain.ParseHealth(resp.State.Health.Status)
		}
	}

	if resp.Config != nil {
		details.OpenStdin = resp.Config.OpenStdin
	}

	return details
}

// mapContainerInspect maps an SDK container.InspectResponse to the
// domain.ContainerInspect projection the Config/Env views render, nil-safely
// (mirroring mapInspectResponse). ID and Name are left zero: they are the
// summary-derived display identity the GUI supplies from the store.
func mapContainerInspect(resp container.InspectResponse) domain.ContainerInspect {
	inspect := domain.ContainerInspect{}

	if resp.ContainerJSONBase != nil {
		inspect.Command = append([]string{resp.Path}, resp.Args...)
	}

	if resp.Config != nil {
		inspect.Image = resp.Config.Image
		inspect.Labels = resp.Config.Labels
		inspect.Env = resp.Config.Env
	}

	for _, m := range resp.Mounts {
		inspect.Mounts = append(inspect.Mounts, domain.Mount{
			Type:        string(m.Type),
			Name:        m.Name,
			Source:      m.Source,
			Destination: m.Destination,
		})
	}

	if resp.NetworkSettings != nil {
		for containerPort, bindings := range resp.NetworkSettings.Ports {
			hostPorts := make([]string, 0, len(bindings))
			for _, binding := range bindings {
				hostPorts = append(hostPorts, binding.HostPort)
			}
			inspect.Ports = append(inspect.Ports, domain.PortBinding{
				ContainerPort: string(containerPort),
				HostPorts:     hostPorts,
			})
		}
		// The SDK carries ports in a map, which iterates in random order; sorting by
		// container port makes the Config view goldenable.
		sort.Slice(inspect.Ports, func(i, j int) bool {
			return inspect.Ports[i].ContainerPort < inspect.Ports[j].ContainerPort
		})
	}

	return inspect
}

// mapNetwork maps an SDK network.Inspect (the type NetworkList also returns, via
// the network.Summary alias) to a domain.Network. Containers is rebuilt as a
// domain-typed map keyed by the same container id; Labels and Options pass through.
func mapNetwork(n network.Inspect) domain.Network {
	nw := domain.Network{
		ID:         n.ID,
		Name:       n.Name,
		Driver:     n.Driver,
		Scope:      n.Scope,
		EnableIPv6: n.EnableIPv6,
		Internal:   n.Internal,
		Attachable: n.Attachable,
		Ingress:    n.Ingress,
		Labels:     n.Labels,
		Options:    n.Options,
	}

	if len(n.Containers) > 0 {
		nw.Containers = make(map[string]domain.NetworkEndpoint, len(n.Containers))
		for id, endpoint := range n.Containers {
			nw.Containers[id] = domain.NetworkEndpoint{
				Name:       endpoint.Name,
				EndpointID: endpoint.EndpointID,
			}
		}
	}

	return nw
}

// mapVolume maps an SDK *volume.Volume (VolumeList returns non-nil elements, the
// pre-migration assumption) to a domain.Volume. Status passes straight through as
// a map[string]any so a nil Status is preserved (the config view renders "n/a");
// UsageData is copied only when the Engine reported it, else left nil.
func mapVolume(v *volume.Volume) domain.Volume {
	vol := domain.Volume{
		Name:       v.Name,
		Driver:     v.Driver,
		Scope:      v.Scope,
		Mountpoint: v.Mountpoint,
		Labels:     v.Labels,
		Options:    v.Options,
		Status:     v.Status,
	}

	if v.UsageData != nil {
		vol.UsageData = &domain.VolumeUsageData{
			RefCount: v.UsageData.RefCount,
			Size:     v.UsageData.Size,
		}
	}

	return vol
}
