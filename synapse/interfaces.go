package synapse

import (
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component" // NEW
	"github.com/SynapticNetworks/temporal-neuron/message"   // NEW
)

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
	// Type returns the categorical type of the component.
	Type() component.ComponentType // Added for consistency with component.Component

	// Receive accepts a neural signal and processes it
	Receive(msg message.NeuralSignal) // CHANGED: From SynapseMessage to message.NeuralSignal

	// ScheduleDelayedDelivery requests the neuron to schedule a message for later delivery.
	// The neuron is responsible for its own axonal queue.
	ScheduleDelayedDelivery(message message.NeuralSignal, target SynapseCompatibleNeuron, delay time.Duration) // CHANGED: From SynapseMessage to message.NeuralSignal
}

// ExtracellularMatrix interface for spatial delay enhancement
type ExtracellularMatrix interface {
	// Enhance existing synaptic delay with spatial factors
	EnhanceSynapticDelay(preNeuronID, postNeuronID, synapseID string, baseDelay time.Duration) time.Duration
}

// PlasticityAdjustment is a feedback message sent from a post-synaptic neuron
// back to a pre-synaptic synapse to trigger a plasticity event (e.g., STDP).
// This models the retrograde signaling mechanisms found in biological systems.
//
// In biology, when a post-synaptic neuron fires, it can send feedback signals
// back to the synapses that contributed to its firing. This feedback contains
// information about the timing relationship between pre- and post-synaptic
// activity, which is used to strengthen or weaken the synaptic connection.
type PlasticityAdjustment struct {
	// DeltaT is the time difference between the pre-synaptic and post-synaptic spikes.
	// Its sign and magnitude determine the direction and strength of the synaptic
	// weight change according to the STDP rule.
	//
	// Convention: Δt = t_pre - t_post
	//   - Δt < 0 (causal): pre-synaptic spike occurred BEFORE post-synaptic spike -> LTP
	//   - Δt > 0 (anti-causal): pre-synaptic spike occurred AFTER post-synaptic spike -> LTD
	//
	// Biological basis: This timing relationship determines whether synapses are
	// strengthened (if they helped cause the post-synaptic firing) or weakened
	// (if they fired after the neuron was already committed to firing).
	DeltaT time.Duration
}
