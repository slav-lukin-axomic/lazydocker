package presentation

import (
	"sort"

	"github.com/docker/docker/api/types/container"
	"github.com/jesseduffield/lazydocker/pkg/commands"
	"github.com/jesseduffield/lazydocker/pkg/domain"
)

// ContainerToDomain is throwaway strangler glue: it converts a legacy
// *commands.Container (which embeds the Docker SDK types and carries a live
// client) into the framework-free *domain.Container that presentation now
// renders. It is deliberately confined to presentation and inlines the mapping
// rather than importing pkg/adapter/docker, which is a driven adapter that
// presentation must not depend on. Delete this once the container store itself
// holds domain.Container in a later strangler slice.
func ContainerToDomain(c *commands.Container) *domain.Container {
	if c == nil {
		return nil
	}

	out := &domain.Container{
		ID:              c.ID,
		Name:            c.Name,
		ServiceName:     c.ServiceName,
		ContainerNumber: c.ContainerNumber,
		ProjectName:     c.ProjectName,
		OneOff:          c.OneOff,
		Image:           c.Container.Image,
		Status:          domain.ParseStatus(c.Container.State),
		Ports:           portsToDomain(c.Container.Ports),
		Labels:          c.Container.Labels,
	}

	if c.DetailsLoaded() {
		details := &domain.ContainerDetails{}
		if c.Details.State != nil {
			details.Running = c.Details.State.Running
			details.ExitCode = c.Details.State.ExitCode
			if c.Details.State.Health != nil {
				details.Health = domain.ParseHealth(c.Details.State.Health.Status)
			}
		}
		if c.Details.Config != nil {
			details.OpenStdin = c.Details.Config.OpenStdin
		}
		out.Details = details
	}

	if stats, ok := c.GetLastStats(); ok {
		out.Stats = &stats.DerivedStats
	}

	return out
}

// ContainerInspectToDomain is throwaway strangler glue (same lifetime as
// ContainerToDomain above): it projects the SDK inspect fields the Config and Env
// views read off a legacy *commands.Container into a framework-free
// domain.ContainerInspect. Callers guard on DetailsLoaded() before calling, so
// c.Details is assumed populated. Delete this at the store-swap slice.
func ContainerInspectToDomain(c *commands.Container) domain.ContainerInspect {
	inspect := domain.ContainerInspect{
		ID:      c.ID,
		Name:    c.Name,
		Image:   c.Details.Config.Image,
		Command: append([]string{c.Details.Path}, c.Details.Args...),
		Labels:  c.Details.Config.Labels,
		Env:     c.Details.Config.Env,
	}

	for _, m := range c.Details.Mounts {
		inspect.Mounts = append(inspect.Mounts, domain.Mount{
			Type:        string(m.Type),
			Name:        m.Name,
			Source:      m.Source,
			Destination: m.Destination,
		})
	}

	for containerPort, bindings := range c.Details.NetworkSettings.Ports {
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
	// container port is what makes the Config view goldenable and is a deliberate
	// improvement over the pre-migration random ordering.
	sort.Slice(inspect.Ports, func(i, j int) bool {
		return inspect.Ports[i].ContainerPort < inspect.Ports[j].ContainerPort
	})

	return inspect
}

// portsToDomain maps SDK container ports to domain ports (Proto is the SDK's
// Type field). A nil input yields a nil slice.
func portsToDomain(ports []container.Port) []domain.Port {
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
