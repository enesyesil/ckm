package runtime

import (
	"context"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

// NewDockerClient creates a shared Docker client
func NewDockerClient() (*client.Client, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

// DockerRuntime manages Docker container lifecycle
type DockerRuntime struct {
	client *client.Client
	logger *zap.Logger
}

// NewDockerRuntime creates a new Docker runtime using shared client
func NewDockerRuntime(dockerClient *client.Client, logger *zap.Logger) *DockerRuntime {
	return &DockerRuntime{client: dockerClient, logger: logger}
}

// CreateContainer creates a new container with resource limits (like cgroups)
func (r *DockerRuntime) CreateContainer(ctx context.Context, imageName string, cmd []string, memoryMB int, cpuShares int64) (string, error) {
	// Convert MB to bytes for memory limit
	memoryBytes := int64(memoryMB) * 1024 * 1024

	// Pull image if not present
	_, _, err := r.client.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		r.logger.Info("Pulling image", zap.String("image", imageName))
		reader, pullErr := r.client.ImagePull(ctx, imageName, image.PullOptions{})
		if pullErr != nil {
			return "", pullErr
		}
		defer reader.Close()
		// Read to completion
		io.Copy(io.Discard, reader)
	}

	// Container configuration
	config := &container.Config{
		Image: imageName,
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

	r.logger.Info("Container created", zap.String("id", resp.ID[:12]), zap.String("image", imageName))
	return resp.ID, nil
}

// StartContainer starts a container
func (r *DockerRuntime) StartContainer(ctx context.Context, containerID string) error {
	err := r.client.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		return err
	}
	r.logger.Info("Container started", zap.String("id", containerID[:12]))
	return nil
}

// StopContainer stops a running container gracefully
func (r *DockerRuntime) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	timeoutSecs := int(timeout.Seconds())
	err := r.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeoutSecs})
	if err != nil {
		return err
	}
	r.logger.Info("Container stopped", zap.String("id", containerID[:12]))
	return nil
}

// RemoveContainer removes a container
func (r *DockerRuntime) RemoveContainer(ctx context.Context, containerID string) error {
	err := r.client.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
	if err != nil {
		return err
	}
	r.logger.Info("Container removed", zap.String("id", containerID[:12]))
	return nil
}

// GetContainerLogs streams container logs
func (r *DockerRuntime) GetContainerLogs(ctx context.Context, containerID string) (io.ReadCloser, error) {
	options := container.LogsOptions{
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
func (r *DockerRuntime) InspectContainer(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return r.client.ContainerInspect(ctx, containerID)
}
