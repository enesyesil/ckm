package balancer

import (
	"sync"
)

// Node represents a backend node
type Node struct {
	ID          string
	Address     string
	Healthy     bool
	Connections int // Current connections
	mu          sync.RWMutex
}

// LoadBalancer distributes requests across multiple nodes
type LoadBalancer struct {
	nodes     []*Node
	algorithm string // "round_robin", "least_connections", "weighted"
	nextIndex int
	mu        sync.Mutex
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(algorithm string) *LoadBalancer {
	return &LoadBalancer{
		nodes:     []*Node{},
		algorithm: algorithm,
	}
}

// AddNode adds a node to the pool
func (lb *LoadBalancer) AddNode(node *Node) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.nodes = append(lb.nodes, node)
}

// SelectNode selects a node based on algorithm
func (lb *LoadBalancer) SelectNode() *Node {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(lb.nodes) == 0 {
		return nil
	}

	// Filter healthy nodes
	healthy := []*Node{}
	for _, node := range lb.nodes {
		if node.Healthy {
			healthy = append(healthy, node)
		}
	}

	if len(healthy) == 0 {
		return nil // No healthy nodes
	}

	switch lb.algorithm {
	case "least_connections":
		return lb.selectLeastConnections(healthy)
	case "weighted":
		return lb.selectWeighted(healthy)
	default: // round_robin
		return lb.selectRoundRobin(healthy)
	}
}

// selectRoundRobin selects next node in rotation
func (lb *LoadBalancer) selectRoundRobin(nodes []*Node) *Node {
	node := nodes[lb.nextIndex%len(nodes)]
	lb.nextIndex++
	return node
}

// selectLeastConnections selects node with fewest connections
func (lb *LoadBalancer) selectLeastConnections(nodes []*Node) *Node {
	best := nodes[0]
	minConn := best.Connections

	for _, node := range nodes[1:] {
		if node.Connections < minConn {
			best = node
			minConn = node.Connections
		}
	}

	return best
}

// selectWeighted selects node based on weight (simplified - uses connections as inverse weight)
func (lb *LoadBalancer) selectWeighted(nodes []*Node) *Node {
	// Simple weighted: prefer nodes with fewer connections
	return lb.selectLeastConnections(nodes)
}

// MarkHealthy marks a node as healthy
func (lb *LoadBalancer) MarkHealthy(nodeID string) {
	for _, node := range lb.nodes {
		if node.ID == nodeID {
			node.mu.Lock()
			node.Healthy = true
			node.mu.Unlock()
		}
	}
}

// MarkUnhealthy marks a node as unhealthy
func (lb *LoadBalancer) MarkUnhealthy(nodeID string) {
	for _, node := range lb.nodes {
		if node.ID == nodeID {
			node.mu.Lock()
			node.Healthy = false
			node.mu.Unlock()
		}
	}
}
