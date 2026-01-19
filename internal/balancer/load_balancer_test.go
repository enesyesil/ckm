package balancer

import (
	"testing"
)

// TestLoadBalancerAddNode tests adding nodes
func TestLoadBalancerAddNode(t *testing.T) {
	lb := NewLoadBalancer("round_robin")

	lb.AddNode(&Node{ID: "node-1", Address: "localhost:8080", Healthy: true})
	lb.AddNode(&Node{ID: "node-2", Address: "localhost:8081", Healthy: true})

	if len(lb.nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(lb.nodes))
	}
}

// TestLoadBalancerRoundRobin tests round-robin selection
func TestLoadBalancerRoundRobin(t *testing.T) {
	lb := NewLoadBalancer("round_robin")

	lb.AddNode(&Node{ID: "node-1", Address: "localhost:8080", Healthy: true})
	lb.AddNode(&Node{ID: "node-2", Address: "localhost:8081", Healthy: true})

	// First selection
	node1 := lb.SelectNode()
	if node1.ID != "node-1" {
		t.Errorf("Expected node-1, got %s", node1.ID)
	}

	// Second selection
	node2 := lb.SelectNode()
	if node2.ID != "node-2" {
		t.Errorf("Expected node-2, got %s", node2.ID)
	}

	// Third selection (wraps around)
	node3 := lb.SelectNode()
	if node3.ID != "node-1" {
		t.Errorf("Expected node-1 (wrap), got %s", node3.ID)
	}
}

// TestLoadBalancerLeastConnections tests least-connections selection
func TestLoadBalancerLeastConnections(t *testing.T) {
	lb := NewLoadBalancer("least_connections")

	lb.AddNode(&Node{ID: "node-1", Address: "localhost:8080", Healthy: true, Connections: 10})
	lb.AddNode(&Node{ID: "node-2", Address: "localhost:8081", Healthy: true, Connections: 5})
	lb.AddNode(&Node{ID: "node-3", Address: "localhost:8082", Healthy: true, Connections: 8})

	node := lb.SelectNode()
	if node.ID != "node-2" {
		t.Errorf("Expected node-2 (least connections), got %s", node.ID)
	}
}

// TestLoadBalancerHealthCheck tests health-based filtering
func TestLoadBalancerHealthCheck(t *testing.T) {
	lb := NewLoadBalancer("round_robin")

	lb.AddNode(&Node{ID: "node-1", Address: "localhost:8080", Healthy: false})
	lb.AddNode(&Node{ID: "node-2", Address: "localhost:8081", Healthy: true})

	node := lb.SelectNode()
	if node.ID != "node-2" {
		t.Errorf("Expected node-2 (only healthy), got %s", node.ID)
	}
}

// TestLoadBalancerNoHealthyNodes tests when no healthy nodes
func TestLoadBalancerNoHealthyNodes(t *testing.T) {
	lb := NewLoadBalancer("round_robin")

	lb.AddNode(&Node{ID: "node-1", Address: "localhost:8080", Healthy: false})

	node := lb.SelectNode()
	if node != nil {
		t.Error("Expected nil when no healthy nodes")
	}
}

// TestLoadBalancerMarkHealthy tests marking node healthy
func TestLoadBalancerMarkHealthy(t *testing.T) {
	lb := NewLoadBalancer("round_robin")

	lb.AddNode(&Node{ID: "node-1", Address: "localhost:8080", Healthy: false})
	lb.MarkHealthy("node-1")

	node := lb.SelectNode()
	if node == nil {
		t.Error("Expected node to be selectable after marking healthy")
	}
}

// TestLoadBalancerMarkUnhealthy tests marking node unhealthy
func TestLoadBalancerMarkUnhealthy(t *testing.T) {
	lb := NewLoadBalancer("round_robin")

	lb.AddNode(&Node{ID: "node-1", Address: "localhost:8080", Healthy: true})
	lb.MarkUnhealthy("node-1")

	node := lb.SelectNode()
	if node != nil {
		t.Error("Expected nil after marking unhealthy")
	}
}

// TestLoadBalancerWeighted tests weighted selection
func TestLoadBalancerWeighted(t *testing.T) {
	lb := NewLoadBalancer("weighted")

	// Node-1 has weight 3, node-2 has weight 1
	// Node-1 should be selected ~75% of the time
	lb.AddNode(&Node{ID: "node-1", Address: "localhost:8080", Healthy: true, Weight: 3})
	lb.AddNode(&Node{ID: "node-2", Address: "localhost:8081", Healthy: true, Weight: 1})

	counts := make(map[string]int)
	for i := 0; i < 100; i++ {
		node := lb.SelectNode()
		counts[node.ID]++
	}

	// Node-1 should be selected more often (around 75 times)
	if counts["node-1"] < 50 {
		t.Errorf("Expected node-1 to be selected more often, got %d", counts["node-1"])
	}
}
