/*
=================================================================================
SIGNAL MEDIATOR - BIOLOGICAL ELECTRICAL SIGNALING
=================================================================================

              --- ARCHITECTURAL ROLE AND BIOLOGICAL CONTEXT ---

This module models ELECTRICAL SYNAPSES (Gap Junctions) and is complementary to,
not redundant with, the CHEMICAL SYNAPSES handled by the 'synapse' package.
They represent two distinct, coexisting modes of communication in the brain.

KEY DISTINCTIONS:

  - CHEMICAL SYNAPSES (direct neuron->synapse->neuron model):
    - One-way, from presynaptic to postsynaptic.
    - Slower, with a significant delay from vesicles, diffusion, and receptors.
    - The primary mechanism for complex computation and learning (plasticity).
    - Implemented by 'synapse' package using VesicleDynamics.

  - ELECTRICAL SYNAPSES (this module):
    - Two-way (bidirectional), allowing ions to flow directly between neurons.
    - Extremely fast, with virtually no delay.
    - Primarily used for synchronizing the firing of entire neuron populations.
    - Less plastic and more structural in nature.

WHEN TO USE THIS MODULE:
You do not need this module for standard information processing via chemical
synapses. However, it is essential for modeling more advanced phenomena:

  - Network Synchronization: To make groups of inhibitory interneurons fire in
    unison, creating brain wave oscillations (e.g., gamma waves).

  - Rapid Reflex Pathways: For circuits where speed is more critical than
    complex computation.

  - Developmental Modeling: To coordinate the activity of developing neurons.

INTEGRATION:
A single neuron can have both types of connections. It would receive chemical
inputs on its main input channel (from the 'synapse' package) and electrical
inputs by implementing the `SignalListener` interface from this package.

=================================================================================
*/

package extracellular

import (
	"sync"
	"time"
)

// SignalMediator routes electrical signals between components
type SignalMediator struct {
	// === SIGNAL ROUTING ===
	listeners map[SignalType][]SignalListener // Signal type -> Listeners

	// === ELECTRICAL COUPLING ===
	connections map[string][]string // Component ID -> Electrically coupled IDs
	conductance map[string]float64  // Connection ID -> Electrical conductance

	// === SIGNAL HISTORY ===
	signalHistory []ElectricalSignalEvent // Recent signal events for analysis
	maxHistory    int                     // Maximum history to maintain

	// === CONCURRENCY CONTROL ===
	mu sync.RWMutex
}

// ElectricalSignalEvent records electrical signal transmission
type ElectricalSignalEvent struct {
	SignalType  SignalType  `json:"signal_type"`
	SourceID    string      `json:"source_id"`
	TargetIDs   []string    `json:"target_ids"`
	Data        interface{} `json:"data"`
	Timestamp   time.Time   `json:"timestamp"`
	Conductance float64     `json:"conductance"`
}

// ElectricalCoupling represents a gap junction connection
type ElectricalCoupling struct {
	ComponentA  string    `json:"component_a"`
	ComponentB  string    `json:"component_b"`
	Conductance float64   `json:"conductance"` // Electrical conductance (0.0-1.0)
	Established time.Time `json:"established"`
	LastUsed    time.Time `json:"last_used"`
	SignalCount int64     `json:"signal_count"`
}

// NewSignalMediator creates a biological electrical signaling system
func NewSignalMediator() *SignalMediator {
	return &SignalMediator{
		listeners:     make(map[SignalType][]SignalListener),
		connections:   make(map[string][]string),
		conductance:   make(map[string]float64),
		signalHistory: make([]ElectricalSignalEvent, 0),
		maxHistory:    1000, // Keep last 1000 signals
	}
}

// =================================================================================
// ELECTRICAL SIGNAL ROUTING (was SignalCoordinator functions)
// =================================================================================

// Send delivers an electrical signal to all registered listeners
func (gj *SignalMediator) Send(signalType SignalType, sourceID string, data interface{}) {
	gj.mu.RLock()
	listeners := make([]SignalListener, len(gj.listeners[signalType]))
	copy(listeners, gj.listeners[signalType])

	// Get electrically coupled components for this source
	coupledIDs := make([]string, 0)
	if coupled, exists := gj.connections[sourceID]; exists {
		coupledIDs = make([]string, len(coupled))
		copy(coupledIDs, coupled)
	}
	gj.mu.RUnlock()

	// Record signal event
	event := ElectricalSignalEvent{
		SignalType: signalType,
		SourceID:   sourceID,
		TargetIDs:  coupledIDs,
		Data:       data,
		Timestamp:  time.Now(),
	}

	gj.recordSignalEvent(event)

	// Direct delivery to all listeners (broadcast signaling)
	for _, listener := range listeners {
		// --- FIX: A component should not receive its own broadcast signal. ---
		// This check prevents a component from reacting to its own events.
		if listener.ID() != sourceID {
			listener.OnSignal(signalType, sourceID, data)
		}
	}

	// Also send to electrically coupled components if they're listeners
	gj.sendToCoupledComponents(signalType, sourceID, data, coupledIDs)
}

// AddListener registers a component to receive electrical signals
func (gj *SignalMediator) AddListener(signalTypes []SignalType, listener SignalListener) {
	gj.mu.Lock()
	defer gj.mu.Unlock()

	for _, signalType := range signalTypes {
		if gj.listeners[signalType] == nil {
			gj.listeners[signalType] = make([]SignalListener, 0)
		}
		// Avoid adding duplicate listeners
		isAlreadyAdded := false
		for _, l := range gj.listeners[signalType] {
			if l.ID() == listener.ID() {
				isAlreadyAdded = true
				break
			}
		}
		if !isAlreadyAdded {
			gj.listeners[signalType] = append(gj.listeners[signalType], listener)
		}
	}
}

// RemoveListener unregisters a component from receiving electrical signals
func (gj *SignalMediator) RemoveListener(signalTypes []SignalType, listener SignalListener) {
	gj.mu.Lock()
	defer gj.mu.Unlock()

	for _, signalType := range signalTypes {
		listeners := gj.listeners[signalType]
		for i, l := range listeners {
			if l.ID() == listener.ID() {
				// Remove listener from slice
				gj.listeners[signalType] = append(listeners[:i], listeners[i+1:]...)
				break
			}
		}
	}
}

// =================================================================================
// ELECTRICAL COUPLING MANAGEMENT (NEW - Gap Junction Biology)
// =================================================================================

// EstablishElectricalCoupling creates a gap junction between two components
func (gj *SignalMediator) EstablishElectricalCoupling(componentA, componentB string, conductance float64) error {
	gj.mu.Lock()
	defer gj.mu.Unlock()

	// Validate conductance (0.0 = no coupling, 1.0 = perfect coupling)
	if conductance < 0.0 || conductance > 1.0 {
		conductance = 0.5 // Default moderate coupling
	}

	// Create bidirectional coupling
	if gj.connections[componentA] == nil {
		gj.connections[componentA] = make([]string, 0)
	}
	if gj.connections[componentB] == nil {
		gj.connections[componentB] = make([]string, 0)
	}

	// Add connections (avoid duplicates)
	if !gj.contains(gj.connections[componentA], componentB) {
		gj.connections[componentA] = append(gj.connections[componentA], componentB)
	}
	if !gj.contains(gj.connections[componentB], componentA) {
		gj.connections[componentB] = append(gj.connections[componentB], componentA)
	}

	// Store conductance for both directions
	connectionID_AB := componentA + "->" + componentB
	connectionID_BA := componentB + "->" + componentA
	gj.conductance[connectionID_AB] = conductance
	gj.conductance[connectionID_BA] = conductance

	return nil
}

// RemoveElectricalCoupling removes a gap junction between components
func (gj *SignalMediator) RemoveElectricalCoupling(componentA, componentB string) error {
	gj.mu.Lock()
	defer gj.mu.Unlock()

	// Remove bidirectional connections
	gj.connections[componentA] = gj.removeFromSlice(gj.connections[componentA], componentB)
	gj.connections[componentB] = gj.removeFromSlice(gj.connections[componentB], componentA)

	// Remove conductance records
	connectionID_AB := componentA + "->" + componentB
	connectionID_BA := componentB + "->" + componentA
	delete(gj.conductance, connectionID_AB)
	delete(gj.conductance, connectionID_BA)

	return nil
}

// GetElectricalCouplings returns all components electrically coupled to the given component
func (gj *SignalMediator) GetElectricalCouplings(componentID string) []string {
	gj.mu.RLock()
	defer gj.mu.RUnlock()

	couplings := gj.connections[componentID]
	if couplings == nil {
		return []string{}
	}

	// Return copy to avoid concurrent modification
	result := make([]string, len(couplings))
	copy(result, couplings)
	return result
}

// GetConductance returns the electrical conductance between two components
func (gj *SignalMediator) GetConductance(componentA, componentB string) float64 {
	gj.mu.RLock()
	defer gj.mu.RUnlock()

	connectionID := componentA + "->" + componentB
	if conductance, exists := gj.conductance[connectionID]; exists {
		return conductance
	}
	return 0.0 // No electrical coupling
}

// =================================================================================
// SIGNAL HISTORY AND ANALYSIS (NEW - Network monitoring)
// =================================================================================

// GetRecentSignals returns recent electrical signal events
func (gj *SignalMediator) GetRecentSignals(count int) []ElectricalSignalEvent {
	gj.mu.RLock()
	defer gj.mu.RUnlock()

	if count > len(gj.signalHistory) {
		count = len(gj.signalHistory)
	}

	if count == 0 {
		return []ElectricalSignalEvent{}
	}

	// Return most recent signals
	start := len(gj.signalHistory) - count
	result := make([]ElectricalSignalEvent, count)
	copy(result, gj.signalHistory[start:])
	return result
}

// GetSignalCount returns total number of signals processed
func (gj *SignalMediator) GetSignalCount() int {
	gj.mu.RLock()
	defer gj.mu.RUnlock()

	return len(gj.signalHistory)
}

// ClearSignalHistory clears the signal history (for memory management)
func (gj *SignalMediator) ClearSignalHistory() {
	gj.mu.Lock()
	defer gj.mu.Unlock()

	gj.signalHistory = make([]ElectricalSignalEvent, 0)
}

// =================================================================================
// INTERNAL UTILITY FUNCTIONS
// =================================================================================

// sendToCoupledComponents sends signals to electrically coupled components
func (gj *SignalMediator) sendToCoupledComponents(signalType SignalType, sourceID string, data interface{}, coupledIDs []string) {
	// This could be extended to implement conductance-based signal attenuation
	// For now, we rely on the broadcast mechanism via listeners
	// Future enhancement: filter signals based on electrical conductance
}

// recordSignalEvent adds an event to the signal history
func (gj *SignalMediator) recordSignalEvent(event ElectricalSignalEvent) {
	gj.mu.Lock()
	defer gj.mu.Unlock()

	gj.signalHistory = append(gj.signalHistory, event)

	// Maintain history size limit
	if len(gj.signalHistory) > gj.maxHistory {
		// Remove oldest events
		copy(gj.signalHistory, gj.signalHistory[len(gj.signalHistory)-gj.maxHistory:])
		gj.signalHistory = gj.signalHistory[:gj.maxHistory]
	}
}

// contains checks if a string slice contains an element
func (gj *SignalMediator) contains(slice []string, element string) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}

// removeFromSlice removes an element from a string slice
func (gj *SignalMediator) removeFromSlice(slice []string, element string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item != element {
			result = append(result, item)
		}
	}
	return result
}
