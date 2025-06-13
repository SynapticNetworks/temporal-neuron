/*
=================================================================================
SIGNAL COORDINATOR - DISCRETE SIGNAL ROUTING
=================================================================================

Handles discrete signaling between components, like action potentials
or connection events. Simple, direct routing without technical abstractions.
=================================================================================
*/

package extracellular

import (
	"sync"
)

// SignalCoordinator routes discrete signals between components
type SignalCoordinator struct {
	listeners map[SignalType][]SignalListener
	mu        sync.RWMutex
}

// NewSignalCoordinator creates a signal coordinator
func NewSignalCoordinator() *SignalCoordinator {
	return &SignalCoordinator{
		listeners: make(map[SignalType][]SignalListener),
	}
}

// Send delivers a signal to all registered listeners
func (sc *SignalCoordinator) Send(signalType SignalType, sourceID string, data interface{}) {
	sc.mu.RLock()
	listeners := make([]SignalListener, len(sc.listeners[signalType]))
	copy(listeners, sc.listeners[signalType])
	sc.mu.RUnlock()

	// Direct delivery to all listeners
	for _, listener := range listeners {
		listener.OnSignal(signalType, sourceID, data)
	}
}

// AddListener registers a component to receive signals
func (sc *SignalCoordinator) AddListener(signalTypes []SignalType, listener SignalListener) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for _, signalType := range signalTypes {
		sc.listeners[signalType] = append(sc.listeners[signalType], listener)
	}
}

// RemoveListener unregisters a component from receiving signals
func (sc *SignalCoordinator) RemoveListener(signalTypes []SignalType, listener SignalListener) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for _, signalType := range signalTypes {
		listeners := sc.listeners[signalType]
		for i, l := range listeners {
			if l == listener {
				// Remove listener from slice
				sc.listeners[signalType] = append(listeners[:i], listeners[i+1:]...)
				break
			}
		}
	}
}
