package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/jesseduffield/lazydocker/pkg/utils"
)

// mustStopSubstring is the fragment Docker includes in the error when it refuses
// to remove a running container without force. Matched to map onto
// domain.ErrContainerRunning, preserving the pre-migration Container.Remove
// behaviour.
const mustStopSubstring = "Stop the container before attempting removal or force remove"

// apiClient is the consumer-defined subset of the Docker SDK client this adapter
// slice calls. Depending on this interface (rather than *client.Client directly)
// keeps the adapter testable with a local fake; the concrete *client.Client
// satisfies it (see the assertion below).
type apiClient interface {
	ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRestart(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerPause(ctx context.Context, containerID string) error
	ContainerUnpause(ctx context.Context, containerID string) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerTop(ctx context.Context, containerID string, arguments []string) (container.TopResponse, error)
	ContainersPrune(ctx context.Context, pruneFilters filters.Args) (container.PruneReport, error)
	ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error)
	ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error)
	NetworkList(ctx context.Context, options network.ListOptions) ([]network.Summary, error)
	NetworkRemove(ctx context.Context, networkID string) error
	NetworksPrune(ctx context.Context, pruneFilters filters.Args) (network.PruneReport, error)
}

var _ apiClient = (*client.Client)(nil)

// Adapter implements the domain.DockerAPI port over the Docker Engine SDK,
// translating SDK types to and from domain types via the mapper.
type Adapter struct {
	client apiClient
}

var _ domain.DockerAPI = (*Adapter)(nil)

// NewAdapter returns an Adapter backed by the given Docker SDK client.
func NewAdapter(c apiClient) *Adapter {
	return &Adapter{client: c}
}

// ListContainers returns all containers (details left nil), mapped to domain
// types.
func (a *Adapter) ListContainers(ctx context.Context) ([]domain.Container, error) {
	summaries, err := a.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	containers := make([]domain.Container, len(summaries))
	for i, summary := range summaries {
		containers[i] = mapContainerSummary(summary)
	}
	return containers, nil
}

// InspectContainer inspects a container and maps the result to domain details.
func (a *Adapter) InspectContainer(ctx context.Context, id string) (domain.ContainerDetails, error) {
	resp, err := a.client.ContainerInspect(ctx, id)
	if err != nil {
		return domain.ContainerDetails{}, err
	}
	return mapInspectResponse(resp), nil
}

// InspectContainerVerbose inspects a container and returns the Config/Env
// projection plus the raw full-inspect data marshalled to YAML. The YAML is
// produced from the SDK's container.InspectResponse via utils.MarshalIntoYaml,
// the same type and function the pre-migration GUI marshalled, so the bytes are
// unchanged.
func (a *Adapter) InspectContainerVerbose(ctx context.Context, id string) (domain.ContainerInspect, string, error) {
	resp, err := a.client.ContainerInspect(ctx, id)
	if err != nil {
		return domain.ContainerInspect{}, "", err
	}
	yamlBytes, err := utils.MarshalIntoYaml(&resp)
	if err != nil {
		return domain.ContainerInspect{}, "", err
	}
	return mapContainerInspect(resp), string(yamlBytes), nil
}

// StartContainer starts a container.
func (a *Adapter) StartContainer(ctx context.Context, id string) error {
	return a.client.ContainerStart(ctx, id, container.StartOptions{})
}

// StopContainer stops a container.
func (a *Adapter) StopContainer(ctx context.Context, id string) error {
	return a.client.ContainerStop(ctx, id, container.StopOptions{})
}

// RestartContainer restarts a container.
func (a *Adapter) RestartContainer(ctx context.Context, id string) error {
	return a.client.ContainerRestart(ctx, id, container.StopOptions{})
}

// PauseContainer pauses a container.
func (a *Adapter) PauseContainer(ctx context.Context, id string) error {
	return a.client.ContainerPause(ctx, id)
}

// UnpauseContainer unpauses a container.
func (a *Adapter) UnpauseContainer(ctx context.Context, id string) error {
	return a.client.ContainerUnpause(ctx, id)
}

// RemoveContainer removes a container. When Docker refuses to remove a running
// container without force, the error is wrapped as domain.ErrContainerRunning so
// callers can branch on it with errors.Is; other errors pass through.
func (a *Adapter) RemoveContainer(ctx context.Context, id string, opts domain.RemoveOptions) error {
	sdkOpts := container.RemoveOptions{
		Force:         opts.Force,
		RemoveVolumes: opts.RemoveVolumes,
	}
	if err := a.client.ContainerRemove(ctx, id, sdkOpts); err != nil {
		if strings.Contains(err.Error(), mustStopSubstring) {
			return fmt.Errorf("%w: %s", domain.ErrContainerRunning, err.Error())
		}
		return err
	}
	return nil
}

// ContainerTop returns the process listing for a container, mapped to a
// domain.TopResult.
func (a *Adapter) ContainerTop(ctx context.Context, id string) (domain.TopResult, error) {
	resp, err := a.client.ContainerTop(ctx, id, []string{})
	if err != nil {
		return domain.TopResult{}, err
	}
	return domain.TopResult{Titles: resp.Titles, Processes: resp.Processes}, nil
}

// PruneContainers removes stopped containers.
func (a *Adapter) PruneContainers(ctx context.Context) error {
	_, err := a.client.ContainersPrune(ctx, filters.Args{})
	return err
}

// ListNetworks returns all networks mapped to domain types, preserving the order
// the Engine reports them (the panel owns sorting).
func (a *Adapter) ListNetworks(ctx context.Context) ([]domain.Network, error) {
	summaries, err := a.client.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, err
	}

	networks := make([]domain.Network, len(summaries))
	for i, summary := range summaries {
		networks[i] = mapNetwork(summary)
	}
	return networks, nil
}

// RemoveNetwork removes the network with the given name (the Engine accepts the
// name in place of the id).
func (a *Adapter) RemoveNetwork(ctx context.Context, name string) error {
	return a.client.NetworkRemove(ctx, name)
}

// PruneNetworks removes all unused networks.
func (a *Adapter) PruneNetworks(ctx context.Context) error {
	_, err := a.client.NetworksPrune(ctx, filters.Args{})
	return err
}

// StreamStats opens the Docker stats stream for a container and invokes onSample
// for each decoded sample until the stream ends or ctx is cancelled. It derives
// CPU/memory percentages the same way the pre-migration
// DockerCommand.CreateClientStatMonitor did, including ignoring per-sample decode
// errors (a malformed line yields a zero-valued sample rather than aborting the
// stream). It returns the stream-open error or the scanner's terminal error.
func (a *Adapter) StreamStats(ctx context.Context, id string, onSample func(*domain.RecordedStats)) error {
	resp, err := a.client.ContainerStats(ctx, id, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var stats domain.ContainerStats
		_ = json.Unmarshal(scanner.Bytes(), &stats)

		onSample(&domain.RecordedStats{
			ClientStats: stats,
			DerivedStats: domain.DerivedStats{
				CPUPercentage:    stats.CalculateContainerCPUPercentage(),
				MemoryPercentage: stats.CalculateContainerMemoryUsage(),
			},
			RecordedAt: time.Now(),
		})
	}

	return scanner.Err()
}

// StreamLogs streams a container's logs to out until the stream ends or ctx is
// cancelled. It inspects the container up front to learn whether a TTY is
// allocated (replacing the pre-migration GUI's wait-for-DetailsLoaded poll), then
// copies the follow stream, keeping all SDK/TTY knowledge inside the adapter.
func (a *Adapter) StreamLogs(ctx context.Context, id string, opts domain.LogOptions, out io.Writer) error {
	resp, err := a.client.ContainerInspect(ctx, id)
	if err != nil {
		return err
	}

	rc, err := a.client.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: opts.Timestamps,
		Since:      opts.Since,
		Tail:       opts.Tail,
		Follow:     true,
	})
	if err != nil {
		return err
	}
	defer rc.Close()

	// A TTY-allocated container multiplexes nothing; a non-TTY stream is the
	// Docker stdout/stderr multiplex that stdcopy must de-mux (both sinks are the
	// same writer, matching the pre-migration GUI behaviour).
	if resp.Config != nil && resp.Config.Tty {
		_, err = io.Copy(out, rc)
	} else {
		_, err = stdcopy.StdCopy(out, out, rc)
	}
	return err
}
