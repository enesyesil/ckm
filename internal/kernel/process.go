package kernel

import (
	"sync"
)

// ProcessGroup represents a group of related processes (like Unix process groups)
type ProcessGroup struct {
	ID        int
	LeaderPID int   // PID of the group leader
	PIDs      []int
	mu        sync.RWMutex
}

// ProcessManager manages process hierarchies and groups
type ProcessManager struct {
	processes     map[int]*ProcessInfo
	processGroups map[int]*ProcessGroup
	sessions      map[int][]int // Session ID -> PIDs
	mu            sync.RWMutex
}

// ProcessInfo stores process metadata
type ProcessInfo struct {
	PID      int
	PPID     int    // Parent PID
	PGID     int    // Process group ID
	SID      int    // Session ID
	State    string // "running", "terminated"
	Children []int  // Child PIDs
}

// NewProcessManager creates a new process manager
func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		processes:     make(map[int]*ProcessInfo),
		processGroups: make(map[int]*ProcessGroup),
		sessions:      make(map[int][]int),
	}
}

// CreateProcess creates a new process with parent relationship
func (pm *ProcessManager) CreateProcess(pid, ppid int) *ProcessInfo {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	info := &ProcessInfo{
		PID:   pid,
		PPID:  ppid,
		State: "running",
	}

	// Add to parent's children list
	if ppid > 0 {
		if parent, ok := pm.processes[ppid]; ok {
			parent.Children = append(parent.Children, pid)
		}
	}

	pm.processes[pid] = info
	return info
}

// CreateProcessGroup creates a new process group
func (pm *ProcessManager) CreateProcessGroup(groupID, leaderPID int) *ProcessGroup {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pg := &ProcessGroup{
		ID:        groupID,
		LeaderPID: leaderPID,
		PIDs:      []int{leaderPID},
	}
	pm.processGroups[groupID] = pg

	// Update process info
	if info, ok := pm.processes[leaderPID]; ok {
		info.PGID = groupID
	}

	return pg
}

// AddToProcessGroup adds a PID to a process group
func (pm *ProcessManager) AddToProcessGroup(groupID, pid int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pg, ok := pm.processGroups[groupID]; ok {
		pg.mu.Lock()
		pg.PIDs = append(pg.PIDs, pid)
		pg.mu.Unlock()
	}

	if info, ok := pm.processes[pid]; ok {
		info.PGID = groupID
	}
}

// GetProcessTree returns the process tree starting from a PID
func (pm *ProcessManager) GetProcessTree(pid int) []*ProcessInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var result []*ProcessInfo
	var traverse func(int)
	
	// Recursive traversal
	traverse = func(p int) {
		if info, ok := pm.processes[p]; ok {
			result = append(result, info)
			for _, child := range info.Children {
				traverse(child)
			}
		}
	}

	traverse(pid)
	return result
}

// TerminateProcess marks a process as terminated
func (pm *ProcessManager) TerminateProcess(pid int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if info, ok := pm.processes[pid]; ok {
		info.State = "terminated"
	}
}
