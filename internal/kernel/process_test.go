package kernel

import (
	"testing"
)

// TestProcessManagerCreateProcess tests process creation
func TestProcessManagerCreateProcess(t *testing.T) {
	pm := NewProcessManager()

	info := pm.CreateProcess(1001, 0)

	if info.PID != 1001 {
		t.Errorf("Expected PID 1001, got %d", info.PID)
	}
	if info.State != "running" {
		t.Errorf("Expected state running, got %s", info.State)
	}
}

// TestProcessManagerParentChild tests parent-child relationships
func TestProcessManagerParentChild(t *testing.T) {
	pm := NewProcessManager()

	// Create parent
	parent := pm.CreateProcess(1001, 0)
	
	// Create child
	pm.CreateProcess(1002, 1001)

	if len(parent.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(parent.Children))
	}
	if parent.Children[0] != 1002 {
		t.Errorf("Expected child PID 1002, got %d", parent.Children[0])
	}
}

// TestProcessManagerProcessGroup tests process group creation
func TestProcessManagerProcessGroup(t *testing.T) {
	pm := NewProcessManager()

	pm.CreateProcess(1001, 0)
	pg := pm.CreateProcessGroup(100, 1001)

	if pg.LeaderPID != 1001 {
		t.Errorf("Expected leader PID 1001, got %d", pg.LeaderPID)
	}
	if len(pg.PIDs) != 1 {
		t.Errorf("Expected 1 PID in group, got %d", len(pg.PIDs))
	}
}

// TestProcessManagerAddToProcessGroup tests adding to process group
func TestProcessManagerAddToProcessGroup(t *testing.T) {
	pm := NewProcessManager()

	pm.CreateProcess(1001, 0)
	pm.CreateProcess(1002, 0)
	pm.CreateProcessGroup(100, 1001)
	pm.AddToProcessGroup(100, 1002)

	pg, _ := pm.processGroups[100]
	if len(pg.PIDs) != 2 {
		t.Errorf("Expected 2 PIDs in group, got %d", len(pg.PIDs))
	}
}

// TestProcessManagerGetProcessTree tests process tree traversal
func TestProcessManagerGetProcessTree(t *testing.T) {
	pm := NewProcessManager()

	// Create process tree: 1001 -> 1002 -> 1003
	pm.CreateProcess(1001, 0)
	pm.CreateProcess(1002, 1001)
	pm.CreateProcess(1003, 1002)

	tree := pm.GetProcessTree(1001)

	if len(tree) != 3 {
		t.Errorf("Expected 3 processes in tree, got %d", len(tree))
	}
}

// TestProcessManagerTerminateProcess tests process termination
func TestProcessManagerTerminateProcess(t *testing.T) {
	pm := NewProcessManager()

	pm.CreateProcess(1001, 0)
	pm.TerminateProcess(1001)

	info := pm.processes[1001]
	if info.State != "terminated" {
		t.Errorf("Expected state terminated, got %s", info.State)
	}
}
