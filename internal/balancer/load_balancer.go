package balancer

import (
	"sync"
)

// Node represents a backend node
type Node struct {
	ID          string
	Address     string
	Healthy     bool
	Weight      int // Node weight for weighted selection (higher = more traffic)
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

// selectWeighted selects node based on weight (higher weight = more likely to be selected)
func (lb *LoadBalancer) selectWeighted(nodes []*Node) *Node {
	// Calculate total weight
	totalWeight := 0
	for _, node := range nodes {
		weight := node.Weight
		if weight <= 0 {
			weight = 1 // Default weight
		}
		totalWeight += weight
	}

	// Use nextIndex as a simple counter to distribute based on weight
	lb.nextIndex++
	position := lb.nextIndex % totalWeight

	// Find the node at this position
	cumulative := 0
	for _, node := range nodes {
		weight := node.Weight
		if weight <= 0 {
			weight = 1
		}
		cumulative += weight
		if position < cumulative {
			return node
		}
	}

	return nodes[0] // Fallback
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
