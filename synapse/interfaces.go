package synapse

import (
	"time"
)

// =================================================================================
// SYNAPSE PROCESSOR INTERFACE
// =================================================================================

// SynapticProcessor defines the universal contract for any component that acts
// as a synapse. It is the key to the pluggable architecture, ensuring that the
// Neuron can work with any synapse type that fulfills these methods.
type SynapticProcessor interface {
	// ID returns the unique identifier for the synapse.
	ID() string

	// Transmit processes an outgoing signal from the pre-synaptic neuron.
	// IMPORTANT: This method completes synchronously - it does NOT block
	// the calling neuron while waiting for delays. Instead, it schedules
	// delayed delivery through the neuron's axonal queue system.
	//
	// Parameters:
	//   signalValue: The strength of the signal from the pre-synaptic neuron
	Transmit(signalValue float64)

	// ApplyPlasticity updates the synapse's internal state based on feedback
	ApplyPlasticity(adjustment PlasticityAdjustment)

	// ShouldPrune evaluates if the synapse should be removed
	ShouldPrune() bool

	// GetWeight returns the current synaptic weight
	GetWeight() float64

	// SetWeight allows direct manipulation of synaptic strength
	SetWeight(weight float64)
}

// =================================================================================
// PLASTICITY AND CONFIGURATION TYPES
// =================================================================================

// PlasticityAdjustment contains feedback for synaptic plasticity
type PlasticityAdjustment struct {
	DeltaT       time.Duration // Spike timing difference for STDP
	PostSynaptic bool          // Whether post-synaptic neuron fired
	PreSynaptic  bool          // Whether pre-synaptic neuron fired recently
	Timestamp    time.Time     // When this adjustment was generated
}

// STDPConfig defines spike-timing dependent plasticity parameters
type STDPConfig struct {
	Enabled        bool          // Whether STDP is active
	LearningRate   float64       // Rate of weight changes
	TimeConstant   time.Duration // STDP time window
	WindowSize     time.Duration // Maximum timing window for plasticity
	MinWeight      float64       // Minimum allowed weight
	MaxWeight      float64       // Maximum allowed weight
	AsymmetryRatio float64       // LTP/LTD asymmetry factor
}

// PruningConfig defines structural plasticity parameters
type PruningConfig struct {
	Enabled             bool          // Whether pruning is active
	WeightThreshold     float64       // Minimum weight to avoid pruning
	InactivityThreshold time.Duration // Maximum inactivity before pruning
}

// ExtracellularMatrix interface for spatial delay enhancement
type ExtracellularMatrix interface {
	EnhanceSynapticDelay(preNeuronID, postNeuronID, synapseID string, baseDelay time.Duration) time.Duration
}
