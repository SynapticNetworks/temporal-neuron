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

// Neuron represents a single processing unit inspired by biological neurons
// Unlike traditional artificial neurons that perform instantaneous calculations,
// this neuron models the temporal dynamics of real neural processing:
// - Accumulates inputs over time (like dendrite integration)
// - Fires when threshold is reached (like action potential generation)
// - Sends outputs with realistic delays (like axon transmission)
// - Supports dynamic connectivity changes (like neuroplasticity)
type Neuron struct {
	// === BIOLOGICAL ACTIVATION PARAMETERS ===
	// These model the electrical properties of real neuron membranes

	threshold float64 // Minimum accumulated charge needed to fire
	// Models the action potential threshold in real neurons
	// Typically around -55mV in biology

	timeWindow time.Duration // Integration time window for input accumulation
	// Models membrane time constant and temporal summation
	// Real neurons: typically 10-20 milliseconds

	fireFactor float64 // Global output multiplier when neuron fires
	// Models the amplitude of the action potential
	// In biology: action potentials have standard amplitude

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

	firstMessage time.Time // Timestamp marking start of current accumulation window
	// Models when the integration period began
	// Used to implement temporal summation windows

	stateMutex sync.Mutex // Protects internal state during message processing
	// Ensures atomic updates to accumulator and timing
}

// NewNeuron creates and initializes a new biologically-inspired neuron
// This factory function sets up all the necessary components for realistic
// neural processing with temporal dynamics and dynamic connectivity
//
// Parameters model key biological properties:
// threshold: electrical threshold for action potential generation
// timeWindow: temporal integration window (membrane time constant)
// fireFactor: action potential amplitude/strength
func NewNeuron(threshold float64, timeWindow time.Duration, fireFactor float64) *Neuron {
	return &Neuron{
		threshold:  threshold,
		timeWindow: timeWindow,
		fireFactor: fireFactor,
		input:      make(chan Message, 100), // Buffered channel prevents blocking
		// Models synaptic vesicle pools
		outputs: make(map[string]*Output), // Empty connection map
	}
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

// Run starts the main neuron processing loop
// This implements the core neural computation cycle that runs continuously:
// 1. Wait for input signals or timeout events
// 2. Integrate incoming signals over time
// 3. Fire when threshold conditions are met
// 4. Reset and repeat
//
// MUST be called as a goroutine: go neuron.Run()
// This allows the neuron to operate independently and concurrently with
// other neurons, modeling the parallel nature of biological neural networks
func (n *Neuron) Run() {
	// Main event loop - the neuron's "life cycle"
	// Continuously processes two types of events:
	for {
		select {
		// Event 1: New input signal received (excitatory or inhibitory)
		// Models: synaptic transmission, neurotransmitter binding,
		//         postsynaptic potential generation
		case msg := <-n.input:
			n.processMessage(msg)

		// Event 2: Time window expired - reset integration period
		// Models: membrane potential decay, end of temporal summation window
		// Prevents indefinite accumulation and maintains realistic timing
		case <-time.After(n.timeWindow):
			n.resetAccumulator()
		}
	}
}

// processMessage handles incoming synaptic signals and manages temporal integration
// This is the core of neural computation - how individual signals are integrated
// over time to determine if the neuron should fire
//
// Biological process modeled:
// 1. Synaptic signal arrives at dendrite
// 2. Creates postsynaptic potential (PSP) - small voltage change
// 3. PSP propagates to cell body (soma)
// 4. Multiple PSPs sum together (spatial and temporal summation)
// 5. If total reaches threshold, action potential is triggered
// 6. Action potential propagates down axon to all output synapses
func (n *Neuron) processMessage(msg Message) {
	n.stateMutex.Lock()         // Protect internal state from concurrent access
	defer n.stateMutex.Unlock() // Ensure lock is always released

	now := time.Now()

	// Initialize new integration window if this is the first signal
	// Models: start of new temporal summation period
	if n.accumulator == 0 {
		n.firstMessage = now
	}

	// Check if we're still within the temporal integration window
	// Models: membrane time constant - how long PSPs last before decaying
	if now.Sub(n.firstMessage) <= n.timeWindow {
		// Add this signal to the running total (temporal summation)
		// Models: algebraic summation of postsynaptic potentials
		// Excitatory signals (positive) depolarize the membrane
		// Inhibitory signals (negative) hyperpolarize the membrane
		n.accumulator += msg.Value

		// Check if accumulated charge has reached the firing threshold
		// Models: action potential initiation at the axon hillock
		if n.accumulator >= n.threshold {
			n.fireUnsafe()             // Generate action potential (already have state lock)
			n.resetAccumulatorUnsafe() // Return to resting state
		}
	} else {
		// Time window has expired - start fresh integration period
		// Models: membrane potential decay back toward resting potential
		// Previous charges have dissipated, start new summation
		n.accumulator = msg.Value
		n.firstMessage = now

		// Check if this single signal alone exceeds threshold
		// Models: strong input causing immediate firing
		if n.accumulator >= n.threshold {
			n.fireUnsafe()
			n.resetAccumulatorUnsafe()
		}
	}
}

// fire triggers action potential propagation to all connected neurons
// Models the all-or-nothing action potential that travels down the axon
// and triggers neurotransmitter release at all synaptic terminals
//
// Biological process:
// 1. Action potential initiated at axon hillock
// 2. Electrical signal propagates down main axon
// 3. Signal reaches all axon terminals simultaneously
// 4. Triggers neurotransmitter release at each synapse
// 5. Each synapse may have different strength/delay characteristics
func (n *Neuron) fire() {
	// Calculate the base signal strength based on current accumulation
	// In biology: action potentials have standard amplitude, but we model
	// signal strength to represent firing frequency or burst patterns
	n.stateMutex.Lock()
	outputValue := n.accumulator * n.fireFactor
	n.stateMutex.Unlock()

	// Get thread-safe snapshot of current connections
	// Prevents connections from changing during signal transmission
	n.outputsMutex.RLock()
	outputsCopy := make(map[string]*Output, len(n.outputs))
	for id, output := range n.outputs {
		outputsCopy[id] = output
	}
	n.outputsMutex.RUnlock()

	// Send signals to all connections in parallel
	// Models: simultaneous action potential propagation to all axon branches
	// Each target receives the signal according to its connection properties
	for _, output := range outputsCopy {
		// Launch separate goroutine for each output connection
		// This models the parallel nature of axonal transmission where
		// one action potential simultaneously affects all connected neurons
		go n.sendToOutput(output, outputValue)
	}
}

// fireUnsafe is the internal firing method called when state lock is already held
// Identical to fire() but assumes caller already has the necessary locks
func (n *Neuron) fireUnsafe() {
	outputValue := n.accumulator * n.fireFactor

	// Get snapshot of outputs (minimal locking since we're already protected)
	n.outputsMutex.RLock()
	outputsCopy := make(map[string]*Output, len(n.outputs))
	for id, output := range n.outputs {
		outputsCopy[id] = output
	}
	n.outputsMutex.RUnlock()

	// Parallel transmission to all outputs
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
