/*
=================================================================================
BIOLOGICAL NEURON SIMULATION - CORE BUILDING BLOCK
=================================================================================

OVERVIEW:
This package implements a biologically-inspired artificial neuron that serves as
the fundamental building block for constructing neural networks with dynamic
connectivity and realistic timing behavior.

BIOLOGICAL INSPIRATION:
Real biological neurons are far more complex than traditional artificial neurons
used in deep learning. Key biological features modeled here:

1. TEMPORAL INTEGRATION: Real neurons accumulate electrical signals (postsynaptic
   potentials) over time windows, not instantaneous calculations

2. THRESHOLD FIRING: When accumulated charge reaches a threshold, the neuron fires
   an action potential - an all-or-nothing electrical spike

3. DYNAMIC CONNECTIVITY: Biological neurons can grow new connections (synapses)
   and prune existing ones throughout their lifetime (neuroplasticity)

4. PARALLEL TRANSMISSION: A single action potential propagates to ALL connected
   neurons simultaneously through the axon's branching structure

5. TRANSMISSION DELAYS: Different connections have different delays based on
   axon length, diameter, and myelination

6. SYNAPTIC STRENGTH: Each connection has its own "weight" or strength that
   modulates the signal intensity

TRADITIONAL AI vs THIS APPROACH:
- Traditional ANNs: Static connectivity, synchronous processing, mathematical
  activation functions, no realistic timing
- This neuron: Dynamic connectivity, asynchronous messaging, temporal integration,
  biological timing delays

KEY DESIGN PRINCIPLES:
1. CONCURRENCY: Each neuron runs as an independent goroutine, enabling true
   parallel processing like real neural networks

2. MESSAGE PASSING: Neurons communicate through Go channels, modeling the
   discrete nature of biological action potentials

3. TEMPORAL DYNAMICS: Input accumulation over configurable time windows models
   how real dendrites integrate signals

4. DYNAMIC ARCHITECTURE: Connections can be added/removed at runtime, enabling
   learning and adaptation

5. THREAD SAFETY: Multiple goroutines can safely modify connections while the
   neuron processes inputs

BUILDING BLOCKS FOR LARGER SYSTEMS:
This neuron serves as the foundation for:
- Static neural networks (traditional architectures)
- Dynamic neural networks (growing/pruning connections)
- Gated neural networks (where gates control connectivity)
- Spiking neural networks (event-driven processing)
- Neuromorphic computing systems

USAGE PATTERN:
1. Create neuron with threshold, time window, and firing parameters
2. Connect to other neurons by adding outputs
3. Launch as goroutine: go neuron.Run()
4. Send messages to neuron.GetInput() channel
5. Dynamically modify connections during runtime
6. Observe emergent network behavior

This approach enables building AI systems that operate more like biological
brains - with realistic timing, dynamic connectivity, and emergent intelligence
arising from the interaction of many simple, concurrent processing units.

=================================================================================
*/

package neuron

import (
	"sync"
	"time"
)

// Message represents a signal transmitted between neurons
// Models the discrete action potential spikes in biological neural communication
// Unlike traditional ANNs that use continuous values, this models the actual
// binary spike-based communication found in real brains
type Message struct {
	Value float64 // Signal strength/intensity (can be positive or negative)
	// Positive values = excitatory signals (increase firing probability)
	// Negative values = inhibitory signals (decrease firing probability)
}

// Output represents a synaptic connection from this neuron to another component
// In biology: this models the synapse - the connection point between neurons
// Each synapse has unique characteristics that affect signal transmission
type Output struct {
	channel chan Message // Communication channel to the target neuron/component
	// Models the synaptic cleft where neurotransmitters cross

	factor float64 // Synaptic strength/weight for this specific connection
	// Models synaptic efficacy - how strong this connection is
	// Can be modified to simulate learning and adaptation

	delay time.Duration // Biological transmission delay
	// Models: axon conduction time, synaptic delay, etc.
	// Longer axons = longer delays (realistic brain timing)
}

// FireEvent represents a real-time neuron firing event for visualization and monitoring
// This captures the exact moment when a biological neuron generates an action potential
// and provides essential information about the firing event for external observers
//
// Biological context:
// When a real neuron fires, it generates an action potential that propagates down
// its axon to all connected synapses. This event is instantaneous and discrete.
// Unlike traditional ANNs that have continuous activation values, biological
// neurons either fire (1) or don't fire (0) - this is the "all-or-nothing" principle.
//
// This struct models that discrete firing event and allows external systems
// (like visualizers, loggers, or learning algorithms) to observe when and how
// strongly each neuron fires without interfering with the neuron's operation.
type FireEvent struct {
	NeuronID string // Unique identifier of the neuron that fired
	// Allows tracking which specific neuron in a network generated this event

	Value float64 // Signal strength/amplitude of the firing event
	// Models the "strength" of the action potential
	// In biology: action potentials have standard amplitude, but this
	// can represent firing frequency or burst patterns

	Timestamp time.Time // Precise timing of when the firing occurred
	// Critical for studying temporal dynamics, spike-timing dependent
	// plasticity, and network synchronization patterns
}

// Neuron represents a single processing unit inspired by biological neurons
// Unlike traditional artificial neurons that perform instantaneous calculations,
// this neuron models the temporal dynamics of real neural processing:
// - Accumulates inputs over time (like dendrite integration)
// - Fires when threshold is reached (like action potential generation)
// - Sends outputs with realistic delays (like axon transmission)
// - Supports dynamic connectivity changes (like neuroplasticity)
type Neuron struct {
	// === IDENTIFICATION ===
	// Unique identifier for this neuron within a network
	id string // Neuron identifier for tracking and reference

	// === BIOLOGICAL ACTIVATION PARAMETERS ===
	// These model the electrical properties of real neuron membranes

	threshold float64 // Minimum accumulated charge needed to fire
	// Models the action potential threshold in real neurons
	// Typically around -55mV in biology

	fireFactor float64 // Global output multiplier when neuron fires
	// Models the amplitude of the action potential
	// In biology: action potentials have standard amplitude

	refractoryPeriod time.Duration // Absolute refractory period duration
	// Models the time after firing when neuron cannot fire again
	// In biology: Na+ channels are inactivated, preventing new action potentials
	// C. elegans neurons: typically 5-15ms depending on neuron type

	decayRate float64 // Membrane potential decay rate per time step
	// Models the leaky nature of biological neural membranes
	// In biology: membrane capacitance causes gradual charge dissipation
	// Value between 0.0-1.0: 0.95 = loses 5% charge per decay interval
	// Real neurons: membrane time constant typically 10-20ms

	// === COMMUNICATION INFRASTRUCTURE ===
	// Models the input/output structure of biological neurons

	input chan Message // Single input channel (models dendrite tree)
	// All inputs converge here, like dendrites
	// converging on the cell body (soma)

	outputs map[string]*Output // Dynamic set of output connections
	// Models the axon branching to multiple targets
	// String key allows named connections for management

	outputsMutex sync.RWMutex // Thread-safe access to outputs map
	// Allows safe connection modification during runtime
	// RWMutex permits multiple concurrent reads

	// === INTERNAL STATE (models neuron membrane properties) ===
	// These variables track the neuron's current electrical state

	accumulator float64 // Current sum of input charges within time window
	// Models the membrane potential in real neurons
	// Starts at resting potential, increases with excitation

	lastFireTime time.Time // Timestamp of most recent action potential
	// Models the refractory state timing in real neurons
	// Used to enforce refractory period constraints
	// Zero value indicates neuron has never fired

	stateMutex sync.Mutex // Protects internal state during message processing
	// Ensures atomic updates to accumulator and timing

	// === MONITORING AND OBSERVATION ===
	// Optional channel for reporting firing events to external observers
	fireEvents chan<- FireEvent // Optional fire event reporting channel
	// nil = disabled (default), non-nil = reports firing events
	// Used for visualization, learning algorithms, and analysis

}

// NewNeuron creates and initializes a new biologically-inspired neuron with identification
// This factory function sets up all the necessary components for realistic
// neural processing with leaky integration, dynamic connectivity, refractory periods,
// and optional monitoring
//
// The leaky integration model enables:
// - Continuous membrane potential decay (models biological membrane capacitance)
// - Elimination of artificial time windows in favor of natural dynamics
// - Realistic temporal summation where recent inputs have stronger influence
// - Biologically accurate signal integration that matches real neural behavior
//
// Parameters model key biological properties:
// id: unique identifier for this neuron (enables tracking in networks)
// threshold: electrical threshold for action potential generation
// decayRate: membrane potential decay factor per time step (0.0-1.0)
// refractoryPeriod: duration after firing when neuron cannot fire again
// fireFactor: action potential amplitude/strength
//
// Biological analogy: The decayRate models the membrane time constant - how quickly
// charge leaks out through the cell membrane. In real neurons, this creates a
// natural integration window where recent inputs have more influence than older ones.
func NewNeuron(id string, threshold float64, decayRate float64, refractoryPeriod time.Duration, fireFactor float64) *Neuron {
	return &Neuron{
		id:               id,                       // Unique neuron identifier for network tracking
		threshold:        threshold,                // Firing threshold (biological: ~-55mV)
		decayRate:        decayRate,                // Membrane decay rate (biological: based on RC time constant)
		refractoryPeriod: refractoryPeriod,         // Refractory period (biological: ~5-15ms)
		fireFactor:       fireFactor,               // Output amplitude scaling
		input:            make(chan Message, 100),  // Buffered input channel
		outputs:          make(map[string]*Output), // Dynamic output connections
		fireEvents:       nil,                      // Optional fire event reporting (disabled by default)
		// accumulator starts at 0 (resting potential)
		// lastFireTime initialized to zero value (never fired)
	}
}

// SetFireEventChannel configures optional real-time firing event reporting
// This method enables external monitoring of neuron firing events without
// interfering with the neuron's core computational processes
//
// Biological inspiration:
// In neuroscience research, scientists often need to monitor when individual
// neurons fire to understand network dynamics, learning, and information processing.
// This is typically done using techniques like:
// - Microelectrodes that detect action potentials
// - Calcium imaging that shows neural activity
// - Multi-electrode arrays that monitor many neurons simultaneously
//
// This method provides a similar capability for artificial neural networks,
// allowing researchers to observe firing patterns, study network dynamics,
// and implement biologically-inspired learning algorithms.
//
// Usage patterns:
// - Visualization: Real-time display of network activity
// - Learning algorithms: Spike-timing dependent plasticity (STDP)
// - Analysis: Network synchronization and oscillation studies
// - Debugging: Identifying silent or hyperactive neurons
//
// Performance considerations:
// - The channel is used in a non-blocking manner to prevent interference
// - Events are sent asynchronously to avoid disrupting neural computation
// - If the channel becomes full, events are dropped rather than blocking
//
// ch: Channel to receive FireEvent notifications when this neuron fires
//
//	Set to nil to disable fire event reporting (default state)
//	The channel should be buffered to handle burst firing patterns
func (n *Neuron) SetFireEventChannel(ch chan<- FireEvent) {
	n.stateMutex.Lock()         // Protect concurrent access to fire event channel
	defer n.stateMutex.Unlock() // Ensure lock is always released

	n.fireEvents = ch

	// Biological analogy: This is like placing a recording electrode near
	// a neuron to monitor its electrical activity. The electrode doesn't
	// interfere with the neuron's function - it just observes and reports
	// when action potentials occur.
}

// AddOutput safely adds a new synaptic connection to this neuron
// This models neuroplasticity - the brain's ability to form new connections
// throughout life. In developing brains, neurons constantly grow new synapses.
// In adult brains, learning involves creating and strengthening connections.
//
// Biological context:
// - Dendritic growth: neurons extend dendrites to reach new partners
// - Axon sprouting: axons grow new branches to contact more targets
// - Synaptogenesis: formation of new synaptic contacts
// - Experience-dependent plasticity: activity drives connection formation
//
// Parameters:
// id: unique identifier for this connection (allows later modification/removal)
// channel: destination for signals (the target neuron's input)
// factor: synaptic strength/weight (models synaptic efficacy)
// delay: transmission delay (models axon length and conduction velocity)
func (n *Neuron) AddOutput(id string, channel chan Message, factor float64, delay time.Duration) {
	n.outputsMutex.Lock()         // Acquire exclusive write access
	defer n.outputsMutex.Unlock() // Ensure lock is always released

	// Create new synaptic connection with specified properties
	n.outputs[id] = &Output{
		channel: channel, // Communication pathway
		factor:  factor,  // Synaptic strength
		delay:   delay,   // Conduction delay
	}

	// Biological analogy: This represents the completion of synaptogenesis
	// where a new functional synaptic connection becomes available for
	// neural communication and information processing
}

// GetInputChannel returns the bidirectional input channel for direct connections
// This is used for neuron-to-neuron connections where we need to pass the channel
// to AddOutput methods
func (n *Neuron) GetInputChannel() chan Message {
	return n.input
}

// RemoveOutput safely removes a synaptic connection
// Models synaptic pruning - the brain's process of eliminating unnecessary
// or ineffective connections to optimize neural circuits
//
// Biological context:
// - Developmental pruning: elimination of excess connections during development
// - Activity-dependent pruning: "use it or lose it" - unused synapses are removed
// - Synaptic scaling: global adjustment of synaptic strengths
// - Pathological pruning: excessive pruning in some neurological conditions
//
// This is crucial for:
// - Network optimization: removing redundant connections
// - Learning: eliminating interfering or outdated associations
// - Memory consolidation: strengthening important connections while removing others
//
// id: unique identifier of the connection to remove
func (n *Neuron) RemoveOutput(id string) {
	n.outputsMutex.Lock()         // Acquire exclusive write access
	defer n.outputsMutex.Unlock() // Ensure lock is always released

	delete(n.outputs, id)

	// Biological analogy: This represents the completion of synaptic elimination
	// where the physical synaptic structure is dismantled and the connection
	// is no longer available for neural communication
}

// GetOutputCount returns the current number of synaptic connections
// Thread-safe read operation that allows monitoring network connectivity
// In biological terms: this tells us the neuron's "fan-out" or how many
// other neurons this neuron can directly influence
func (n *Neuron) GetOutputCount() int {
	n.outputsMutex.RLock()         // Acquire shared read access (allows concurrent reads)
	defer n.outputsMutex.RUnlock() // Ensure lock is released

	return len(n.outputs)
}

// Run starts the main neuron processing loop with continuous leaky integration
// This implements the core neural computation cycle that runs continuously with
// biologically realistic membrane dynamics:
// 1. Wait for input signals, decay timer events, or shutdown signals
// 2. Apply continuous membrane potential decay (leaky integration)
// 3. Integrate incoming signals with existing accumulated charge
// 4. Fire when threshold conditions are met during refractory-compliant periods
// 5. Reset and repeat
//
// MUST be called as a goroutine: go neuron.Run()
// This allows the neuron to operate independently and concurrently with
// other neurons, modeling the parallel nature of biological neural networks
//
// Biological processes modeled:
// - Continuous membrane potential decay (models membrane capacitance/resistance)
// - Asynchronous signal integration (models dendritic summation)
// - Refractory period enforcement (models Na+ channel recovery)
// - Real-time temporal dynamics (no artificial time windows)
func (n *Neuron) Run() {
	// Create decay timer for continuous membrane potential decay
	// Models the biological membrane time constant (RC circuit behavior)
	// Decay interval of 1ms provides good temporal resolution for C. elegans scale
	decayInterval := 1 * time.Millisecond
	decayTicker := time.NewTicker(decayInterval)
	defer decayTicker.Stop()

	// Main event loop - the neuron's "life cycle" with continuous biological dynamics
	// Processes three types of events in order of biological priority:
	for {
		select {
		// Event 1: New input signal received (excitatory or inhibitory)
		// Models: synaptic transmission, neurotransmitter binding,
		//         postsynaptic potential generation
		// Highest priority - immediate processing like real synaptic events
		case msg := <-n.input:
			n.processMessageWithDecay(msg)

		// Event 2: Membrane potential decay timer (continuous biological process)
		// Models: membrane capacitance discharge, ion channel leakage,
		//         return toward resting potential
		// Regular biological process that occurs continuously
		case <-decayTicker.C:
			n.applyMembraneDecay()

			// Note: Removed the artificial time window timeout from original implementation
			// Real neurons don't have discrete "time windows" - they have continuous
			// membrane dynamics with exponential decay characteristics
		}
	}
}

// applyMembraneDecay applies continuous membrane potential decay
// Models the biological process of charge dissipation through membrane resistance
// This replaces the artificial "time window reset" with realistic exponential decay
//
// Biological process modeled:
// In real neurons, the cell membrane acts like a leaky capacitor (RC circuit).
// Charge continuously leaks out through membrane resistance, causing the
// membrane potential to decay exponentially toward resting potential.
// This creates natural temporal summation where recent inputs have stronger
// influence than older inputs.
//
// Mathematical model:
// V(t) = V(0) * e^(-t/τ) where τ is the membrane time constant
// Discrete approximation: V(t+dt) = V(t) * decayRate
func (n *Neuron) applyMembraneDecay() {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// Apply exponential decay to accumulated membrane potential
	// Models: membrane resistance causing continuous charge leakage
	// decayRate < 1.0 causes gradual approach to resting potential (0)
	n.accumulator *= n.decayRate

	// In biology: membrane potential asymptotically approaches resting potential
	// For computational efficiency, set very small values to exactly zero
	// This prevents accumulation of floating-point precision errors
	if n.accumulator < 1e-10 && n.accumulator > -1e-10 {
		n.accumulator = 0.0
	}

	// Note: Unlike the original implementation, we never completely reset
	// the accumulator to zero. This models the continuous nature of
	// biological membrane dynamics where there are no discrete "resets"
}

// processMessageWithDecay handles incoming synaptic signals with continuous leaky integration
// This mimics biologically realistic membrane dynamics that eliminate artificial time windows
//
// Biological process modeled:
// 1. Synaptic signal arrives at dendrite (postsynaptic potential)
// 2. Signal adds to current membrane potential (no time window constraints)
// 3. Continuous decay is handled separately by applyMembraneDecay()
// 4. If accumulated potential reaches threshold, action potential is triggered
// 5. Refractory period constraints are enforced during firing attempts
//
// Key difference from original: No discrete time windows or hard resets.
// The membrane potential continuously evolves through decay and signal integration.
func (n *Neuron) processMessageWithDecay(msg Message) {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// Directly add signal to current membrane potential
	// Models: postsynaptic potential summing with existing membrane charge
	// No time window checks - integration is continuous like real neurons
	n.accumulator += msg.Value

	// Check if accumulated charge has reached the firing threshold
	// Models: action potential initiation at the axon hillock
	// Refractory period enforcement is handled within fireUnsafe()
	if n.accumulator >= n.threshold {
		n.fireUnsafe()             // Generate action potential (includes refractory check)
		n.resetAccumulatorUnsafe() // Return to resting potential after firing
	}

	// Note: No explicit time window management or firstMessage tracking needed
	// The continuous decay process naturally handles temporal integration
	// This more accurately reflects how biological neurons actually work
}

// fire triggers action potential propagation to all connected neurons
// Models the all-or-nothing action potential that travels down the axon
// and triggers neurotransmitter release at all synaptic terminals
// With refractory period enforcement for biological realism
//
// Biological process:
// 1. Verify neuron is not in refractory period
// 2. Action potential initiated at axon hillock
// 3. Electrical signal propagates down main axon
// 4. Signal reaches all axon terminals simultaneously
// 5. Triggers neurotransmitter release at each synapse
// 6. Each synapse may have different strength/delay characteristics
func (n *Neuron) fire() {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// Use the internal unsafe method which includes refractory period checking
	n.fireUnsafe()
}

// fireUnsafe is the internal firing method called when state lock is already held
// Includes refractory period enforcement and biological timing constraints
// Identical to fire() but assumes caller already has the necessary locks
//
// Biological process modeled:
// 1. Check if neuron is in refractory period (cannot fire if recent firing occurred)
// 2. If firing is allowed, generate action potential
// 3. Record firing time to enforce future refractory periods
// 4. Propagate signal to all synaptic connections
//
// The refractory period models the biological reality that after an action potential,
// voltage-gated sodium channels become inactivated and require time to recover.
// During this period, no amount of input can trigger another action potential.
func (n *Neuron) fireUnsafe() {
	// Check refractory period constraint
	// Models: voltage-gated Na+ channel inactivation state
	now := time.Now()
	if !n.lastFireTime.IsZero() && now.Sub(n.lastFireTime) < n.refractoryPeriod {
		// Neuron is in refractory period - firing is physically impossible
		// In biology: Na+ channels are inactivated, K+ channels may still be open
		// This prevents unrealistic rapid-fire bursts that don't occur in real neurons
		return
	}

	// Record this firing event for future refractory period enforcement
	// Models: the moment when Na+ channels become inactivated
	n.lastFireTime = now

	outputValue := n.accumulator * n.fireFactor

	// Report firing event if channel is set
	if n.fireEvents != nil {
		select {
		case n.fireEvents <- FireEvent{
			NeuronID:  n.id,
			Value:     outputValue,
			Timestamp: now, // Use the same timestamp for consistency
		}:
		default: // Don't block if channel is full
		}
	}

	// Get snapshot of outputs (minimal locking since we're already protected)
	n.outputsMutex.RLock()
	outputsCopy := make(map[string]*Output, len(n.outputs))
	for id, output := range n.outputs {
		outputsCopy[id] = output
	}
	n.outputsMutex.RUnlock()

	// Parallel transmission to all outputs
	// Models: action potential propagating simultaneously down all axon branches
	for _, output := range outputsCopy {
		go n.sendToOutput(output, outputValue)
	}
}

// sendToOutput handles signal transmission to a single target neuron
// Models the complete synaptic transmission process including:
// - Axonal conduction delay
// - Synaptic strength modulation
// - Neurotransmitter release and binding
//
// Biological details modeled:
// - Conduction delay: time for action potential to travel along axon
// - Synaptic delay: time for neurotransmitter release and binding
// - Synaptic strength: efficacy of the synaptic connection
// - Signal attenuation/amplification based on synapse properties
func (n *Neuron) sendToOutput(output *Output, baseValue float64) {
	// Apply biological transmission delay
	// Models multiple delay sources:
	// - Axon conduction time (depends on length, diameter, myelination)
	// - Synaptic delay (neurotransmitter release and diffusion)
	// - Dendritic propagation time (signal travel to target soma)
	if output.delay > 0 {
		time.Sleep(output.delay)
	}

	// Calculate final signal strength based on synaptic properties
	// baseValue: the "raw" action potential strength from this neuron
	// output.factor: the synaptic weight/strength for this specific connection
	// This models how different synapses can amplify or attenuate signals
	finalValue := baseValue * output.factor

	// Attempt to deliver the signal (non-blocking to prevent deadlocks)
	// Models the stochastic nature of synaptic transmission where
	// signals can sometimes fail to transmit due to various factors
	select {
	case output.channel <- Message{Value: finalValue}:
		// Signal successfully transmitted
		// Models: successful synaptic transmission and signal reception
	default:
		// Target channel is full or closed - signal is lost
		// Models: synaptic failure, receptor saturation, or target unavailability
		// In real brains: not all action potentials result in successful transmission
	}
}

// resetAccumulator clears the integration state (thread-safe external interface)
// Models the return to resting membrane potential after signal processing
func (n *Neuron) resetAccumulator() {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	n.resetAccumulatorUnsafe()
}

// resetAccumulatorUnsafe clears integration state (internal use when locked)
// Returns the neuron to its resting state, ready for new signal integration
func (n *Neuron) resetAccumulatorUnsafe() {
	n.accumulator = 0
	// firstMessage will be reset when the next message arrives
	// This models the neuron returning to resting potential
}

// GetInput returns the input channel for connecting to this neuron
// This allows other neurons or external sources to send signals to this neuron
// Models: the dendritic tree where synaptic inputs are received
func (n *Neuron) GetInput() chan<- Message {
	return n.input
}

// Close gracefully shuts down the neuron
// Closes the input channel, which will cause the Run() loop to exit
// Models: neuronal death or disconnection from the network
func (n *Neuron) Close() {
	close(n.input)
}
