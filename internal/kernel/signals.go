package kernel

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// SignalHandler manages OS signal handling
type SignalHandler struct {
	signals  chan os.Signal
	handlers map[os.Signal][]func()
	mu       sync.RWMutex
}

// NewSignalHandler creates a new signal handler
func NewSignalHandler() *SignalHandler {
	sh := &SignalHandler{
		signals:  make(chan os.Signal, 1),
		handlers: make(map[os.Signal][]func()),
	}
	
	// Register common signals
	signal.Notify(sh.signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	return sh
}

// RegisterHandler registers a handler for a specific signal
func (sh *SignalHandler) RegisterHandler(sig os.Signal, handler func()) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	sh.handlers[sig] = append(sh.handlers[sig], handler)
}

// Start starts listening for signals
func (sh *SignalHandler) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case sig := <-sh.signals:
				sh.mu.RLock()
				handlers := sh.handlers[sig]
				sh.mu.RUnlock()
				
				// Execute all handlers for this signal
				for _, handler := range handlers {
					handler()
				}
			}
		}
	}()
}

// SendSignal sends a signal to handlers (for testing)
func (sh *SignalHandler) SendSignal(sig os.Signal) {
	sh.signals <- sig
}
