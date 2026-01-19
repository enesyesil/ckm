package runtime

import (
	"context"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

// DockerRuntime manages Docker container lifecycle
type DockerRuntime struct {
	client *client.Client
	logger *zap.Logger
}

// NewDockerRuntime creates a new Docker runtime client
func NewDockerRuntime(logger *zap.Logger) (*DockerRuntime, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerRuntime{client: cli, logger: logger}, nil
}

// CreateContainer creates a new container with resource limits (like cgroups)
func (r *DockerRuntime) CreateContainer(ctx context.Context, image string, cmd []string, memoryMB int, cpuShares int64) (string, error) {
	// Convert MB to bytes for memory limit
	memoryBytes := int64(memoryMB) * 1024 * 1024
	
	// Container configuration
	config := &container.Config{
		Image: image,
		Cmd:   cmd,
	}

	// Resource limits (cgroups-like)
	hostConfig := &container.HostConfig{
		Resources: container.Resources{
			Memory:     memoryBytes,
			MemorySwap: memoryBytes, // Disable swap
			CPUShares:  cpuShares,   // CPU weight (1024 = 1 CPU)
		},
	}

	resp, err := r.client.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return "", err
	}

	r.logger.Info("Container created", zap.String("id", resp.ID[:12]), zap.String("image", image))
	return resp.ID, nil
}

// StartContainer starts a container
func (r *DockerRuntime) StartContainer(ctx context.Context, containerID string) error {
	err := r.client.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}
	r.logger.Info("Container started", zap.String("id", containerID[:12]))
	return nil
}

// StopContainer stops a running container gracefully
func (r *DockerRuntime) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	err := r.client.ContainerStop(ctx, containerID, &timeout)
	if err != nil {
		return err
	}
	r.logger.Info("Container stopped", zap.String("id", containerID[:12]))
	return nil
}

// RemoveContainer removes a container
func (r *DockerRuntime) RemoveContainer(ctx context.Context, containerID string) error {
	err := r.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		return err
	}
	r.logger.Info("Container removed", zap.String("id", containerID[:12]))
	return nil
}

// GetContainerLogs streams container logs
func (r *DockerRuntime) GetContainerLogs(ctx context.Context, containerID string) (io.ReadCloser, error) {
	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	}
	return r.client.ContainerLogs(ctx, containerID, options)
}

// WaitContainer waits for container to finish and returns exit code
func (r *DockerRuntime) WaitContainer(ctx context.Context, containerID string) (int64, error) {
	statusCh, errCh := r.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return -1, err
	case status := <-statusCh:
		return status.StatusCode, nil
	}
}

// InspectContainer gets container status and details
func (r *DockerRuntime) InspectContainer(ctx context.Context, containerID string) (*types.ContainerJSON, error) {
	return r.client.ContainerInspect(ctx, containerID)
}
