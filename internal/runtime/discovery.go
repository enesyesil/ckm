package runtime

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"ckm/internal/common"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

// ContainerStats holds real-time resource metrics for a container
type ContainerStats struct {
	ContainerID   string
	ContainerName string
	ImageName     string
	CPUPercent    float64
	MemoryUsage   uint64
	MemoryLimit   uint64
	MemoryPercent float64
	NetworkRx     uint64
	NetworkTx     uint64
	BlockRead     uint64
	BlockWrite    uint64
	PIDs          uint64
	Status        string
	StartedAt     time.Time
	UpdatedAt     time.Time
}

// ContainerDiscovery monitors all running Docker containers
type ContainerDiscovery struct {
	client     *client.Client
	logger     *zap.Logger
	interval   time.Duration
	containers map[string]*ContainerStats
	mu         sync.RWMutex
	stopCh     chan struct{}
}

// NewContainerDiscovery creates a new container discovery service using shared Docker client
func NewContainerDiscovery(dockerClient *client.Client, logger *zap.Logger, interval time.Duration) *ContainerDiscovery {
	return &ContainerDiscovery{
		client:     dockerClient,
		logger:     logger,
		interval:   interval,
		containers: make(map[string]*ContainerStats),
		stopCh:     make(chan struct{}),
	}
}

// Start begins the container discovery loop
func (d *ContainerDiscovery) Start(ctx context.Context) {
	d.logger.Info("Starting container discovery", zap.Duration("interval", d.interval))

	// Initial scan
	d.discoverAndCollect(ctx)

	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			d.logger.Info("Container discovery stopped")
			return
		case <-d.stopCh:
			d.logger.Info("Container discovery stopped")
			return
		case <-ticker.C:
			d.discoverAndCollect(ctx)
		}
	}
}

// Stop stops the discovery service
func (d *ContainerDiscovery) Stop() {
	close(d.stopCh)
}

// discoverAndCollect finds all containers and collects their stats
func (d *ContainerDiscovery) discoverAndCollect(ctx context.Context) {
	// List all running containers
	containers, err := d.client.ContainerList(ctx, container.ListOptions{All: false})
	if err != nil {
		d.logger.Error("Failed to list containers", zap.Error(err))
		return
	}

	// Track which containers we've seen
	seen := make(map[string]bool)

	for _, c := range containers {
		seen[c.ID] = true

		// Get container stats
		stats, err := d.getContainerStats(ctx, c.ID)
		if err != nil {
			d.logger.Debug("Failed to get stats", zap.String("id", c.ID[:12]), zap.Error(err))
			continue
		}

		// Get container name (remove leading /)
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}

		// Parse start time
		inspect, _ := d.client.ContainerInspect(ctx, c.ID)
		startedAt, _ := time.Parse(time.RFC3339Nano, inspect.State.StartedAt)

		containerStats := &ContainerStats{
			ContainerID:   c.ID,
			ContainerName: name,
			ImageName:     c.Image,
			CPUPercent:    stats.cpuPercent,
			MemoryUsage:   stats.memoryUsage,
			MemoryLimit:   stats.memoryLimit,
			MemoryPercent: stats.memoryPercent,
			NetworkRx:     stats.networkRx,
			NetworkTx:     stats.networkTx,
			BlockRead:     stats.blockRead,
			BlockWrite:    stats.blockWrite,
			PIDs:          stats.pids,
			Status:        c.State,
			StartedAt:     startedAt,
			UpdatedAt:     time.Now(),
		}

		d.mu.Lock()
		d.containers[c.ID] = containerStats
		d.mu.Unlock()

		// Export metrics to Prometheus
		d.exportMetrics(containerStats)
	}

	// Remove containers that are no longer running
	d.mu.Lock()
	for id := range d.containers {
		if !seen[id] {
			delete(d.containers, id)
		}
	}
	d.mu.Unlock()

	// Update discovered container count
	common.DiscoveredContainers.Set(float64(len(seen)))
}

// statsResult holds parsed stats from Docker
type statsResult struct {
	cpuPercent    float64
	memoryUsage   uint64
	memoryLimit   uint64
	memoryPercent float64
	networkRx     uint64
	networkTx     uint64
	blockRead     uint64
	blockWrite    uint64
	pids          uint64
}

// getContainerStats retrieves real-time stats for a container
func (d *ContainerDiscovery) getContainerStats(ctx context.Context, containerID string) (*statsResult, error) {
	// Get stats stream (one-shot)
	statsResp, err := d.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, err
	}
	defer statsResp.Body.Close()

	// Parse JSON stats
	var stats types.StatsJSON
	if err := json.NewDecoder(statsResp.Body).Decode(&stats); err != nil {
		return nil, err
	}

	// Calculate CPU percentage
	cpuPercent := calculateCPUPercent(&stats)

	// Calculate memory usage
	memoryUsage := stats.MemoryStats.Usage
	memoryLimit := stats.MemoryStats.Limit
	memoryPercent := 0.0
	if memoryLimit > 0 {
		memoryPercent = float64(memoryUsage) / float64(memoryLimit) * 100.0
	}

	// Calculate network I/O
	var networkRx, networkTx uint64
	for _, v := range stats.Networks {
		networkRx += v.RxBytes
		networkTx += v.TxBytes
	}

	// Calculate block I/O
	var blockRead, blockWrite uint64
	for _, v := range stats.BlkioStats.IoServiceBytesRecursive {
		switch v.Op {
		case "read", "Read":
			blockRead += v.Value
		case "write", "Write":
			blockWrite += v.Value
		}
	}

	return &statsResult{
		cpuPercent:    cpuPercent,
		memoryUsage:   memoryUsage,
		memoryLimit:   memoryLimit,
		memoryPercent: memoryPercent,
		networkRx:     networkRx,
		networkTx:     networkTx,
		blockRead:     blockRead,
		blockWrite:    blockWrite,
		pids:          stats.PidsStats.Current,
	}, nil
}

// calculateCPUPercent calculates CPU usage percentage
func calculateCPUPercent(stats *types.StatsJSON) float64 {
	// CPU delta
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)
	// System delta
	systemDelta := float64(stats.CPUStats.SystemUsage - stats.PreCPUStats.SystemUsage)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuCount := float64(stats.CPUStats.OnlineCPUs)
		if cpuCount == 0 {
			cpuCount = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
		}
		if cpuCount == 0 {
			cpuCount = 1
		}
		return (cpuDelta / systemDelta) * cpuCount * 100.0
	}
	return 0.0
}

// exportMetrics exports container stats to Prometheus
func (d *ContainerDiscovery) exportMetrics(stats *ContainerStats) {
	labels := []string{stats.ContainerName, stats.ImageName}

	common.ContainerCPUPercent.WithLabelValues(labels...).Set(stats.CPUPercent)
	common.ContainerMemoryBytes.WithLabelValues(labels...).Set(float64(stats.MemoryUsage))
	common.ContainerMemoryPercent.WithLabelValues(labels...).Set(stats.MemoryPercent)
	common.ContainerNetworkRxBytes.WithLabelValues(labels...).Set(float64(stats.NetworkRx))
	common.ContainerNetworkTxBytes.WithLabelValues(labels...).Set(float64(stats.NetworkTx))
	common.ContainerBlockReadBytes.WithLabelValues(labels...).Set(float64(stats.BlockRead))
	common.ContainerBlockWriteBytes.WithLabelValues(labels...).Set(float64(stats.BlockWrite))
	common.ContainerPIDs.WithLabelValues(labels...).Set(float64(stats.PIDs))
}

// GetAllStats returns stats for all discovered containers
func (d *ContainerDiscovery) GetAllStats() []*ContainerStats {
	d.mu.RLock()
	defer d.mu.RUnlock()

	result := make([]*ContainerStats, 0, len(d.containers))
	for _, stats := range d.containers {
		result = append(result, stats)
	}
	return result
}

// GetStats returns stats for a specific container
func (d *ContainerDiscovery) GetStats(containerID string) *ContainerStats {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.containers[containerID]
}
