package synapse

import "time"

// SynapticProcessor defines the universal contract for any component that acts
// as a synapse. It is the key to the pluggable architecture, ensuring that the
// Neuron can work with any synapse type that fulfills these methods.
//
// This interface abstracts away the implementation details of different synapse
// types, allowing for:
// - Static synapses (fixed weights)
// - Plastic synapses (learning via STDP)
// - Inhibitory vs excitatory synapses
// - Fast vs slow synapses
// - Complex multi-compartment synapses
//
// All synapse types can be used interchangeably by neurons as long as they
// implement this interface.
type SynapticProcessor interface {
	// ID returns the unique identifier for the synapse.
	// This allows neurons to manage and reference specific synapses by name.
	ID() string

	// Transmit processes an outgoing signal from the pre-synaptic neuron.
	// It is responsible for applying the synapse's weight and handling the
	// axonal transmission delay before delivering the signal to the
	// post-synaptic neuron.
	//
	// Parameters:
	//   signalValue: The strength of the signal from the pre-synaptic neuron
	//
	// The synapse applies its weight and delay, then sends the modified signal
	// to the post-synaptic neuron after the appropriate delay period.
	Transmit(signalValue float64)

	// ApplyPlasticity updates the synapse's internal state (e.g., weight)
	// based on a feedback signal from the post-synaptic neuron, typically
	// containing spike timing information (Δt) for STDP.
	//
	// Parameters:
	//   adjustment: Contains timing and other information needed for plasticity
	//
	// This method implements the learning aspect of synapses, allowing them
	// to strengthen or weaken based on their effectiveness in causing
	// post-synaptic firing.
	ApplyPlasticity(adjustment PlasticityAdjustment)

	// ShouldPrune evaluates the synapse's internal state to determine if it
	// has become ineffective and should be removed by the parent neuron.
	// This method encapsulates the logic for structural plasticity.
	//
	// Returns:
	//   true if the synapse should be eliminated, false otherwise
	//
	// The decision is based on factors like:
	// - Current synaptic weight (too weak?)
	// - Recent activity levels (unused?)
	// - Time since last plasticity event (stagnant?)
	ShouldPrune() bool

	// GetWeight returns the current effective weight of the synapse. This is
	// crucial for monitoring, debugging, and validating learning.
	//
	// Returns:
	//   The current synaptic weight/strength
	GetWeight() float64

	// SetWeight allows for direct experimental manipulation of the synapse's
	// strength. This is a thread-safe method.
	//
	// Parameters:
	//   weight: The new weight to set for this synapse
	//
	// This method is useful for:
	// - Experimental manipulation
	// - Initialization of specific weight patterns
	// - Testing network behavior with controlled weights
	SetWeight(weight float64)
}

// =================================================================================
// NEURON INTERFACE FOR SYNAPSE COMMUNICATION
// This defines what methods a neuron must have to work with synapses
// =================================================================================

// SynapseCompatibleNeuron defines the interface that neurons must implement
// to work with the synapse system. This allows synapses to communicate with
// neurons without depending on specific neuron implementations.
type SynapseCompatibleNeuron interface {
	// ID returns the unique identifier of the neuron
	ID() string

	// Receive accepts a synapse message and processes it
	// This method should be added to existing neuron implementations
	Receive(msg SynapseMessage)
	ScheduleDelayedDelivery(message SynapseMessage, target SynapseCompatibleNeuron, delay time.Duration)
}

// ExtracellularMatrix interface for spatial delay enhancement
type ExtracellularMatrix interface {
	// Enhance existing synaptic delay with spatial factors
	EnhanceSynapticDelay(preNeuronID, postNeuronID, synapseID string, baseDelay time.Duration) time.Duration
}
