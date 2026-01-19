package kernel

import (
	"testing"
	"time"
)

// TestNextPID tests PID generation
func TestNextPID(t *testing.T) {
	pid1 := NextPID()
	pid2 := NextPID()

	if pid2 <= pid1 {
		t.Errorf("Expected pid2 > pid1, got pid1=%d, pid2=%d", pid1, pid2)
	}
}

// TestClassifyWorkload tests workload classification by file extension
func TestClassifyWorkload(t *testing.T) {
	tests := []struct {
		file         string
		expectedType string
		expectedPrio int
	}{
		{"notebook.ipynb", "notebook", 0},
		{"script.py", "task", 1},
		{"deploy.sh", "task", 1},
		{"ubuntu.iso", "vm", 2},
		{"disk.qcow2", "vm", 2},
		{"unknown.txt", "task", 2},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			typ, prio := ClassifyWorkload(tt.file)
			if typ != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, typ)
			}
			if prio != tt.expectedPrio {
				t.Errorf("Expected priority %d, got %d", tt.expectedPrio, prio)
			}
		})
	}
}

// TestFIFOSchedulerAdd tests FIFO scheduler add operation
func TestFIFOSchedulerAdd(t *testing.T) {
	s := NewFIFOScheduler()
	
	w := Workload{ID: "test-1", Type: "task", CPUTime: time.Second}
	s.Add(w)

	if len(s.queue) != 1 {
		t.Errorf("Expected queue length 1, got %d", len(s.queue))
	}
}

// TestRoundRobinSchedulerAdd tests round-robin scheduler add operation
func TestRoundRobinSchedulerAdd(t *testing.T) {
	s := NewRoundRobinScheduler(100 * time.Millisecond)
	
	w := Workload{ID: "test-1", Type: "task", CPUTime: time.Second}
	s.Add(w)

	if len(s.queue) != 1 {
		t.Errorf("Expected queue length 1, got %d", len(s.queue))
	}
}

// TestPrioritySchedulerAdd tests priority scheduler add operation
func TestPrioritySchedulerAdd(t *testing.T) {
	s := NewPriorityScheduler()
	
	w := Workload{ID: "test-1", Type: "task", Priority: 1}
	s.Add(w)

	if len(s.queue) != 1 {
		t.Errorf("Expected queue length 1, got %d", len(s.queue))
	}
}

// TestFairSchedulerAdd tests fair scheduler add operation
func TestFairSchedulerAdd(t *testing.T) {
	s := NewFairScheduler(100 * time.Millisecond)
	
	w := Workload{ID: "test-1", Type: "task", CPUTime: time.Second}
	s.Add(w)

	if len(s.queue) != 1 {
		t.Errorf("Expected queue length 1, got %d", len(s.queue))
	}
}

// TestMultilevelSchedulerRouting tests multilevel scheduler routing
func TestMultilevelSchedulerRouting(t *testing.T) {
	vmSched := NewFIFOScheduler()
	taskSched := NewFIFOScheduler()
	m := NewMultilevelScheduler(vmSched, taskSched)

	// Add VM workload
	vmWork := Workload{ID: "vm-1", Type: "vm"}
	m.Add(vmWork)

	// Add task workload
	taskWork := Workload{ID: "task-1", Type: "task"}
	m.Add(taskWork)

	if len(vmSched.queue) != 1 {
		t.Errorf("Expected VM queue length 1, got %d", len(vmSched.queue))
	}
	if len(taskSched.queue) != 1 {
		t.Errorf("Expected task queue length 1, got %d", len(taskSched.queue))
	}
}
