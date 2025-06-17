/*
=================================================================================
SIGNAL MEDIATOR - BIOLOGICAL ELECTRICAL SIGNALING SYSTEM
=================================================================================

This module implements ELECTRICAL SYNAPSES (Gap Junctions) - the bidirectional,
high-speed electrical connections between neurons that enable rapid synchronization
and coordination across neural networks.

BIOLOGICAL CONTEXT AND NEURAL COMMUNICATION MODES:

The brain uses two fundamentally different communication systems:

1. CHEMICAL SYNAPSES (handled by the 'synapse' package):
   - Unidirectional signal transmission via neurotransmitter release
   - Transmission delays of 0.5-5ms due to vesicle dynamics and diffusion
   - Primary mechanism for complex computation, learning, and plasticity
   - Supports sophisticated signal modulation and memory formation

2. ELECTRICAL SYNAPSES (this module):
   - Bidirectional ionic current flow through gap junction channels
   - Virtually instantaneous transmission (<0.1ms delay)
   - Essential for network synchronization and coordinated oscillations
   - More structurally stable, less plastic than chemical synapses

FUNCTIONAL ROLES IN NEURAL NETWORKS:

Gap junctions are crucial for several biological phenomena:
- Gamma Oscillations (30-100Hz): Interneuron synchronization for attention/cognition
- Theta Rhythms (4-8Hz): Hippocampal network coordination for memory
- Motor Control: Rapid reflex pathways requiring minimal delay
- Development: Coordinating neural activity during circuit formation
- Homeostasis: Maintaining consistent network activity levels

WHEN TO USE THIS MODULE:

Use electrical coupling for:
✓ Network synchronization and oscillatory behavior
✓ Rapid signal propagation where speed > computational complexity
✓ Population-level coordination (e.g., inhibitory interneuron networks)
✓ Modeling brain rhythms and network-wide activity patterns

Don't use for:
✗ Individual learning and plasticity (use chemical synapses)
✗ Complex signal transformation (use computational neurons)
✗ Memory formation and storage (use synaptic plasticity)

INTEGRATION WITH CHEMICAL SYNAPSES:

A single neuron typically has both connection types:
- Receives chemical inputs for computation and learning
- Participates in electrical networks for synchronization
- Implements SignalListener interface to receive electrical signals
- Maintains separate processing for electrical and chemical inputs

=================================================================================
*/

package extracellular

import (
	"sync"
	"time"
)

// =================================================================================
// BIOLOGICAL CONSTANTS AND CONFIGURATION
// =================================================================================

const (
	// === GAP JUNCTION CONDUCTANCE PARAMETERS ===

	// DEFAULT_CONDUCTANCE represents moderate electrical coupling strength
	// Biological basis: Typical gap junction conductance in neural networks
	// Range: 0.1-1.0 normalized units (actual: 10-100 pS per channel)
	DEFAULT_CONDUCTANCE float64 = 0.5

	// MIN_CONDUCTANCE represents the minimum detectable electrical coupling
	// Below this threshold, gap junctions have negligible functional impact
	MIN_CONDUCTANCE float64 = 0.0

	// MAX_CONDUCTANCE represents perfect electrical coupling
	// Rarely achieved in biology but useful for synchronization studies
	MAX_CONDUCTANCE float64 = 1.0

	// === SIGNAL HISTORY MANAGEMENT ===

	// DEFAULT_HISTORY_SIZE limits memory usage for signal event tracking
	// Biological rationale: Balances analysis capability with resource efficiency
	// Sufficient for analyzing ~1 second of high-frequency (1kHz) activity
	DEFAULT_HISTORY_SIZE int = 1000

	// === BIOLOGICAL TIMING CONSTRAINTS ===

	// GAP_JUNCTION_DELAY represents the minimal transmission delay through electrical synapses
	// Biological basis: Gap junctions have near-instantaneous transmission (~0.01-0.1ms)
	// Much faster than chemical synapses (0.5-5ms) due to direct ionic coupling
	GAP_JUNCTION_DELAY = 50 * time.Microsecond

	// SYNCHRONIZATION_WINDOW defines the time window for considering signals as synchronized
	// Biological basis: Interneurons can synchronize within 1-2ms for gamma oscillations
	SYNCHRONIZATION_WINDOW = 2 * time.Millisecond
)

// =================================================================================
// CORE DATA STRUCTURES
// =================================================================================

// SignalMediator coordinates electrical signal transmission between neural components
// via gap junction connections, modeling the bidirectional, high-speed electrical
// communication pathways found in biological neural networks.
//
// BIOLOGICAL INSPIRATION:
// Gap junctions are specialized membrane channels that directly connect the cytoplasm
// of adjacent cells, allowing ions and small molecules to flow freely between neurons.
// This creates electrical continuity that enables rapid signal propagation and
// network synchronization without the delays inherent in chemical transmission.
type SignalMediator struct {
	// === SIGNAL ROUTING INFRASTRUCTURE ===

	// listeners maps signal types to components that should receive them
	// Implements the broadcast signaling mechanism where electrical events
	// are delivered to all registered components simultaneously
	listeners map[SignalType][]SignalListener

	// === GAP JUNCTION CONNECTIVITY MATRIX ===

	// connections stores the bidirectional electrical coupling topology
	// Maps each component ID to all its electrically coupled neighbors
	// Biological basis: Gap junction distribution in neural tissue
	connections map[string][]string

	// conductance stores the electrical coupling strength for each connection
	// Key format: "componentA->componentB" for directed lookup efficiency
	// Biological basis: Variable gap junction channel density and permeability
	conductance map[string]float64

	// === ACTIVITY MONITORING AND ANALYSIS ===

	// signalHistory maintains a chronological record of electrical signal events
	// Enables analysis of network synchronization patterns and oscillatory behavior
	// Essential for studying gamma rhythms, theta oscillations, and population dynamics
	signalHistory []ElectricalSignalEvent

	// maxHistory limits memory usage while preserving recent activity patterns
	// Configurable to balance analysis needs with computational resources
	maxHistory int

	// === THREAD SAFETY AND CONCURRENCY CONTROL ===

	// mu protects all shared data structures from concurrent access
	// Read-write mutex allows multiple concurrent reads but exclusive writes
	// Essential for thread-safe operation in multi-neuron simulations
	mu sync.RWMutex
}

// ElectricalSignalEvent records a single electrical signal transmission event
// through the gap junction network, capturing all relevant biological and
// computational parameters for analysis and replay.
//
// BIOLOGICAL SIGNIFICANCE:
// Each event represents the propagation of ionic current through gap junctions,
// similar to how electrical activity spreads through interconnected neurons
// in biological networks. The timing and connectivity information enables
// analysis of synchronization patterns and network dynamics.
type ElectricalSignalEvent struct {
	// Signal classification and routing information
	SignalType SignalType  `json:"signal_type"` // Type of electrical signal (spike, subthreshold, etc.)
	SourceID   string      `json:"source_id"`   // Component that initiated the signal
	TargetIDs  []string    `json:"target_ids"`  // All components that received the signal
	Data       interface{} `json:"data"`        // Signal payload (voltage, current, etc.)

	// Temporal and biophysical parameters
	Timestamp   time.Time `json:"timestamp"`   // Precise timing for synchronization analysis
	Conductance float64   `json:"conductance"` // Effective coupling strength for this transmission
}

// ElectricalCoupling represents a bidirectional gap junction connection between
// two neural components, including all relevant biophysical properties and
// activity statistics for analysis and optimization.
//
// BIOLOGICAL CONTEXT:
// Gap junctions are dynamic structures that can be modulated by calcium levels,
// pH, voltage, and various signaling molecules. This structure captures both
// the static connectivity and dynamic usage patterns.
type ElectricalCoupling struct {
	// Connection topology
	ComponentA string `json:"component_a"` // First connected component
	ComponentB string `json:"component_b"` // Second connected component

	// Biophysical properties
	Conductance float64 `json:"conductance"` // Electrical coupling strength (0.0-1.0)

	// Temporal tracking
	Established time.Time `json:"established"`  // When this coupling was created
	LastUsed    time.Time `json:"last_used"`    // Most recent signal transmission
	SignalCount int64     `json:"signal_count"` // Total signals transmitted through this junction
}

// =================================================================================
// CONSTRUCTOR AND INITIALIZATION
// =================================================================================

// NewSignalMediator creates a new electrical signaling coordinator for managing
// gap junction connections and signal routing in neural networks.
//
// BIOLOGICAL MODELING:
// Initializes the gap junction network infrastructure with empty connectivity
// matrices, ready to establish electrical couplings as neurons are added and
// connections are formed based on spatial proximity and functional requirements.
//
// Returns a fully initialized SignalMediator with default biological parameters.
func NewSignalMediator() *SignalMediator {
	return &SignalMediator{
		listeners:     make(map[SignalType][]SignalListener),
		connections:   make(map[string][]string),
		conductance:   make(map[string]float64),
		signalHistory: make([]ElectricalSignalEvent, 0, DEFAULT_HISTORY_SIZE),
		maxHistory:    DEFAULT_HISTORY_SIZE,
	}
}

// =================================================================================
// ELECTRICAL SIGNAL TRANSMISSION
// =================================================================================

// Send propagates an electrical signal through the gap junction network to all
// connected components, modeling the rapid bidirectional current flow that
// characterizes electrical synaptic transmission in biological neural networks.
//
// BIOLOGICAL PROCESS:
// When a neuron's membrane potential changes (due to action potentials or
// subthreshold activity), ionic current flows through gap junction channels
// to all electrically coupled neighbors. This creates near-instantaneous
// signal propagation that enables network synchronization and coordinated
// oscillatory activity.
//
// SIGNAL ROUTING:
// 1. Delivers signal to all registered listeners (broadcast mechanism)
// 2. Propagates through direct electrical couplings (gap junctions)
// 3. Records signal event for network analysis and monitoring
// 4. Prevents self-signaling to avoid feedback loops
//
// Parameters:
//
//	signalType: Classification of the electrical signal (spike, subthreshold, etc.)
//	sourceID: Component ID that initiated the signal
//	data: Signal payload containing voltage/current information
//
// BIOLOGICAL CONSTRAINTS:
//   - Self-signaling prevention models the biological reality that neurons
//     don't stimulate themselves through their own gap junctions
//   - Instantaneous delivery models the <0.1ms transmission delay of gap junctions
func (sm *SignalMediator) Send(signalType SignalType, sourceID string, data interface{}) {
	// Acquire read lock to safely access listener and connection data
	sm.mu.RLock()

	// Create local copies to avoid holding locks during signal delivery
	listeners := make([]SignalListener, len(sm.listeners[signalType]))
	copy(listeners, sm.listeners[signalType])

	// Identify all electrically coupled components for this source
	coupledIDs := make([]string, 0)
	if coupled, exists := sm.connections[sourceID]; exists {
		coupledIDs = make([]string, len(coupled))
		copy(coupledIDs, coupled)
	}

	sm.mu.RUnlock()

	// Record signal transmission event for network analysis
	event := ElectricalSignalEvent{
		SignalType:  signalType,
		SourceID:    sourceID,
		TargetIDs:   coupledIDs,
		Data:        data,
		Timestamp:   time.Now(),
		Conductance: sm.getAverageConductance(sourceID, coupledIDs),
	}
	sm.recordSignalEvent(event)

	// Deliver signal to all registered listeners (broadcast transmission)
	// Models the volume conduction aspect of electrical signaling
	for _, listener := range listeners {
		// BIOLOGICAL CONSTRAINT: Prevent self-signaling
		// Gap junctions don't create feedback loops within the same neuron
		if listener.ID() != sourceID {
			listener.OnSignal(signalType, sourceID, data)
		}
	}

	// Signal propagation through direct electrical couplings is handled
	// via the listener mechanism - coupled components register as listeners
}

// AddListener registers a neural component to receive electrical signals of
// specified types, enabling participation in the gap junction signaling network.
//
// BIOLOGICAL CONTEXT:
// This models the process of gap junction formation during neural development
// or the activation of existing electrical synapses. In biology, gap junctions
// can be dynamically regulated by various factors including calcium levels,
// membrane voltage, and neuromodulatory signals.
//
// DUPLICATE PREVENTION:
// Prevents the same component from being registered multiple times for the
// same signal type, modeling the biological constraint that a single gap
// junction connection provides a fixed level of electrical coupling.
//
// Parameters:
//
//	signalTypes: Array of signal types this component should receive
//	listener: Component implementing the SignalListener interface
//
// Thread-safe operation with write locking to prevent concurrent modification.
func (sm *SignalMediator) AddListener(signalTypes []SignalType, listener SignalListener) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, signalType := range signalTypes {
		// Initialize listener slice if needed
		if sm.listeners[signalType] == nil {
			sm.listeners[signalType] = make([]SignalListener, 0)
		}

		// Check for existing registration to prevent duplicates
		// Models biological constraint of fixed gap junction connectivity
		isAlreadyAdded := false
		for _, existingListener := range sm.listeners[signalType] {
			if existingListener.ID() == listener.ID() {
				isAlreadyAdded = true
				break
			}
		}

		// Add listener only if not already registered
		if !isAlreadyAdded {
			sm.listeners[signalType] = append(sm.listeners[signalType], listener)
		}
	}
}

// RemoveListener unregisters a neural component from receiving electrical signals,
// modeling gap junction closure or electrical synapse elimination during neural
// development, injury, or homeostatic regulation.
//
// BIOLOGICAL CONTEXT:
// Gap junctions can be dynamically closed through various mechanisms including
// changes in intracellular calcium, membrane voltage, or specific signaling
// pathways. This function models the removal of electrical connectivity while
// preserving other network connections.
//
// Parameters:
//
//	signalTypes: Array of signal types to stop receiving
//	listener: Component to remove from signal distribution
//
// Thread-safe operation with write locking for atomic listener removal.
func (sm *SignalMediator) RemoveListener(signalTypes []SignalType, listener SignalListener) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, signalType := range signalTypes {
		listeners := sm.listeners[signalType]

		// Find and remove the specified listener
		for i, existingListener := range listeners {
			if existingListener.ID() == listener.ID() {
				// Remove listener using slice manipulation
				sm.listeners[signalType] = append(listeners[:i], listeners[i+1:]...)
				break // Exit loop after removal
			}
		}
	}
}

// =================================================================================
// GAP JUNCTION MANAGEMENT
// =================================================================================

// EstablishElectricalCoupling creates a bidirectional gap junction connection
// between two neural components with specified electrical conductance, modeling
// the formation of direct electrical synapses in biological neural networks.
//
// BIOLOGICAL PROCESS:
// Gap junction formation involves the alignment of connexin hemichannels from
// adjacent cell membranes to create aqueous pores allowing direct ionic flow.
// The conductance depends on the number of channels, their open probability,
// and the electrochemical driving forces.
//
// BIDIRECTIONAL CONNECTIVITY:
// Unlike chemical synapses, gap junctions provide symmetric, bidirectional
// coupling. Current flows in both directions based on the voltage gradient,
// enabling true electrical continuity between coupled neurons.
//
// CONDUCTANCE VALIDATION:
// Ensures conductance values remain within biological limits (0.0-1.0).
// Invalid values are set to DEFAULT_CONDUCTANCE to maintain network stability.
//
// Parameters:
//
//	componentA: First component in the electrical coupling
//	componentB: Second component in the electrical coupling
//	conductance: Electrical coupling strength (0.0 = no coupling, 1.0 = perfect coupling)
//
// Returns error if coupling establishment fails (currently always nil for compatibility).
func (sm *SignalMediator) EstablishElectricalCoupling(componentA, componentB string, conductance float64) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Validate and normalize conductance to biological limits
	if conductance < MIN_CONDUCTANCE || conductance > MAX_CONDUCTANCE {
		conductance = DEFAULT_CONDUCTANCE
	}

	// Initialize connection slices if needed
	if sm.connections[componentA] == nil {
		sm.connections[componentA] = make([]string, 0)
	}
	if sm.connections[componentB] == nil {
		sm.connections[componentB] = make([]string, 0)
	}

	// Establish bidirectional connectivity (avoiding duplicates)
	if !sm.contains(sm.connections[componentA], componentB) {
		sm.connections[componentA] = append(sm.connections[componentA], componentB)
	}
	if !sm.contains(sm.connections[componentB], componentA) {
		sm.connections[componentB] = append(sm.connections[componentB], componentA)
	}

	// Store symmetric conductance values for both directions
	// Format: "sourceID->targetID" for efficient bidirectional lookup
	connectionKeyAB := componentA + "->" + componentB
	connectionKeyBA := componentB + "->" + componentA
	sm.conductance[connectionKeyAB] = conductance
	sm.conductance[connectionKeyBA] = conductance

	return nil
}

// RemoveElectricalCoupling eliminates a gap junction connection between two
// neural components, modeling electrical synapse elimination during development,
// injury recovery, or homeostatic network reorganization.
//
// BIOLOGICAL CONTEXT:
// Gap junction closure can occur through various mechanisms including
// phosphorylation of connexin proteins, changes in intracellular pH or calcium,
// or developmental pruning of electrical connections during circuit refinement.
//
// BIDIRECTIONAL REMOVAL:
// Removes both directions of the electrical coupling and cleans up all
// associated conductance records to prevent memory leaks and ensure
// network consistency.
//
// Parameters:
//
//	componentA: First component in the coupling to remove
//	componentB: Second component in the coupling to remove
//
// Returns error if removal fails (currently always nil for compatibility).
func (sm *SignalMediator) RemoveElectricalCoupling(componentA, componentB string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Remove bidirectional connections from adjacency lists
	sm.connections[componentA] = sm.removeFromSlice(sm.connections[componentA], componentB)
	sm.connections[componentB] = sm.removeFromSlice(sm.connections[componentB], componentA)

	// Clean up conductance records for both directions
	connectionKeyAB := componentA + "->" + componentB
	connectionKeyBA := componentB + "->" + componentA
	delete(sm.conductance, connectionKeyAB)
	delete(sm.conductance, connectionKeyBA)

	return nil
}

// GetElectricalCouplings returns all neural components electrically coupled to
// the specified component via gap junctions, providing network topology information
// for analysis and visualization of electrical connectivity patterns.
//
// BIOLOGICAL APPLICATION:
// This function enables analysis of gap junction networks, which are crucial
// for understanding synchronization patterns, oscillatory behavior, and
// population dynamics in neural circuits.
//
// Parameters:
//
//	componentID: Component to query for electrical connections
//
// Returns a copy of the coupling list to prevent concurrent modification issues.
// Empty slice returned for components with no electrical connections.
func (sm *SignalMediator) GetElectricalCouplings(componentID string) []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	couplings := sm.connections[componentID]
	if couplings == nil {
		return []string{}
	}

	// Return defensive copy to prevent external modification
	result := make([]string, len(couplings))
	copy(result, couplings)
	return result
}

// GetConductance returns the electrical coupling strength between two neural
// components, representing the gap junction conductance that determines the
// efficiency of electrical signal transmission.
//
// BIOLOGICAL SIGNIFICANCE:
// Gap junction conductance varies based on the number of connexin channels,
// their open probability, and regulatory factors. Higher conductance enables
// stronger electrical coupling and more effective synchronization.
//
// Parameters:
//
//	componentA: Source component for conductance measurement
//	componentB: Target component for conductance measurement
//
// Returns 0.0 if no electrical coupling exists between the components.
func (sm *SignalMediator) GetConductance(componentA, componentB string) float64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	connectionKey := componentA + "->" + componentB
	if conductance, exists := sm.conductance[connectionKey]; exists {
		return conductance
	}
	return 0.0 // No electrical coupling present
}

// =================================================================================
// NETWORK ANALYSIS AND MONITORING
// =================================================================================

// GetRecentSignals returns the most recent electrical signal events for analysis
// of network activity patterns, synchronization behavior, and oscillatory dynamics
// in the gap junction network.
//
// ANALYTICAL APPLICATIONS:
// - Gamma oscillation analysis (30-100Hz interneuron synchronization)
// - Theta rhythm tracking (4-8Hz hippocampal coordination)
// - Population burst detection and characterization
// - Network synchronization measurement and optimization
//
// Parameters:
//
//	count: Number of recent signal events to retrieve
//
// Returns chronologically ordered events (oldest first) up to the requested count.
// Empty slice returned if no signals have been recorded.
func (sm *SignalMediator) GetRecentSignals(count int) []ElectricalSignalEvent {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Validate and limit count to available history
	if count > len(sm.signalHistory) {
		count = len(sm.signalHistory)
	}

	if count == 0 {
		return []ElectricalSignalEvent{}
	}

	// Extract most recent signals (chronologically ordered)
	start := len(sm.signalHistory) - count
	result := make([]ElectricalSignalEvent, count)
	copy(result, sm.signalHistory[start:])
	return result
}

// GetSignalCount returns the total number of electrical signals processed by
// this mediator, providing a measure of overall network electrical activity
// and gap junction utilization.
//
// BIOLOGICAL RELEVANCE:
// Signal count reflects the level of electrical communication in the network,
// which correlates with synchronization strength and oscillatory activity.
// High signal counts indicate active electrical coupling networks.
func (sm *SignalMediator) GetSignalCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.signalHistory)
}

// ClearSignalHistory removes all recorded signal events to free memory and
// reset analysis state, useful for long-running simulations or when transitioning
// between different experimental phases.
//
// MEMORY MANAGEMENT:
// Signal history can accumulate significantly during long simulations.
// Periodic clearing prevents excessive memory usage while maintaining
// the ability to analyze recent network activity patterns.
func (sm *SignalMediator) ClearSignalHistory() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.signalHistory = make([]ElectricalSignalEvent, 0, sm.maxHistory)
}

// =================================================================================
// INTERNAL UTILITY FUNCTIONS
// =================================================================================

// recordSignalEvent adds a new electrical signal event to the history buffer
// with automatic memory management to prevent unbounded growth during long
// simulations while preserving recent activity for analysis.
//
// MEMORY MANAGEMENT STRATEGY:
// Maintains a rolling buffer of recent signals by removing oldest events
// when the history size exceeds maxHistory. This balances analysis capability
// with memory efficiency for long-running neural network simulations.
func (sm *SignalMediator) recordSignalEvent(event ElectricalSignalEvent) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.signalHistory = append(sm.signalHistory, event)

	// Enforce history size limit using efficient slice operations
	if len(sm.signalHistory) > sm.maxHistory {
		// Shift buffer to maintain most recent events
		excessCount := len(sm.signalHistory) - sm.maxHistory
		copy(sm.signalHistory, sm.signalHistory[excessCount:])
		sm.signalHistory = sm.signalHistory[:sm.maxHistory]
	}
}

// getAverageConductance calculates the average electrical coupling strength
// for signal transmission from a source to multiple coupled targets, used
// for signal event recording and network analysis.
//
// BIOLOGICAL MODELING:
// When a signal propagates through multiple gap junctions simultaneously,
// the average conductance provides a measure of overall coupling effectiveness
// for that transmission event.
func (sm *SignalMediator) getAverageConductance(sourceID string, targetIDs []string) float64 {
	if len(targetIDs) == 0 {
		return 0.0
	}

	totalConductance := 0.0
	for _, targetID := range targetIDs {
		totalConductance += sm.GetConductance(sourceID, targetID)
	}

	return totalConductance / float64(len(targetIDs))
}

// contains checks if a string slice contains a specific element, used for
// duplicate prevention in connection management and listener registration.
//
// EFFICIENCY NOTE:
// Linear search is acceptable for typical gap junction networks where
// each neuron connects to a small number (5-20) of electrical neighbors.
func (sm *SignalMediator) contains(slice []string, element string) bool {
	for _, item := range slice {
		if item == element {
			return true
		}
	}
	return false
}

// removeFromSlice removes a specific element from a string slice and returns
// the modified slice, used for cleaning up electrical connections during
// gap junction removal operations.
//
// MEMORY EFFICIENCY:
// Creates a new slice with appropriate capacity to avoid memory waste
// while maintaining efficient removal operations for connection management.
func (sm *SignalMediator) removeFromSlice(slice []string, element string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item != element {
			result = append(result, item)
		}
	}
	return result
}
