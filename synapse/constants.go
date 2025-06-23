// synapse/constants.go
package synapse

import "time"

// ===========================================================================
// STDP (Spike-Timing Dependent Plasticity) Constants
// ===========================================================================

const (
	// STDP_DEFAULT_LEARNING_RATE is the default base rate of synaptic weight changes.
	// Biologically, this is often a small percentage (e.g., 1-5% per STDP event).
	STDP_DEFAULT_LEARNING_RATE float64 = 0.01

	// STDP_DEFAULT_TIME_CONSTANT is the default exponential decay time constant
	// of the STDP window. This determines how quickly the plasticity effect
	// diminishes with increasing time difference between pre- and post-synaptic spikes.
	// Typical biological values are around 10-50 milliseconds.
	STDP_DEFAULT_TIME_CONSTANT time.Duration = 20 * time.Millisecond

	// STDP_DEFAULT_WINDOW_SIZE is the default maximum timing difference for
	// STDP effects. Spikes separated by more than this duration generally
	// exhibit no plasticity. Typical biological windows are around ±50-200 milliseconds.
	STDP_DEFAULT_WINDOW_SIZE time.Duration = 100 * time.Millisecond

	// STDP_DEFAULT_MIN_WEIGHT is the default minimum allowed synaptic weight.
	// This prevents the complete elimination of a synapse due to prolonged LTD,
	// maintaining a residual connection. Biological synapses are rarely truly zero.
	STDP_DEFAULT_MIN_WEIGHT float64 = 0.001

	// STDP_DEFAULT_MAX_WEIGHT is the default maximum allowed synaptic weight.
	// This models receptor saturation and prevents runaway strengthening due to
	// pathological LTP, maintaining network stability.
	STDP_DEFAULT_MAX_WEIGHT float64 = 2.0

	// STDP_DEFAULT_ASYMMETRY_RATIO is the default ratio of LTD to LTP strength.
	// In many biological systems, Long-Term Depression (LTD) is slightly stronger
	// or has a different curve than Long-Term Potentiation (LTP).
	// A value > 1.0 means LTD is stronger.
	STDP_DEFAULT_ASYMMETRY_RATIO float64 = 1.2
)

// ===========================================================================
// Pruning (Structural Plasticity) Constants
// ===========================================================================

const (
	// PRUNING_DEFAULT_WEIGHT_THRESHOLD is the default weight below which a synapse
	// is considered a candidate for pruning. Very weak synapses are often
	// functionally irrelevant and can be removed.
	PRUNING_DEFAULT_WEIGHT_THRESHOLD float64 = 0.01

	// PRUNING_DEFAULT_INACTIVITY_THRESHOLD is the default duration of inactivity
	// required for a weak synapse to be eligible for pruning. This provides a
	// "grace period" for synapses to demonstrate their usefulness.
	// Biological timescales for pruning are typically hours to days.
	PRUNING_DEFAULT_INACTIVITY_THRESHOLD time.Duration = 5 * time.Minute

	// PRUNING_CONSERVATIVE_WEIGHT_THRESHOLD is a more lenient weight threshold
	// for pruning, meaning synapses need to be extremely weak to be considered.
	PRUNING_CONSERVATIVE_WEIGHT_THRESHOLD float64 = 0.001

	// PRUNING_CONSERVATIVE_INACTIVITY_THRESHOLD is a longer inactivity period
	// for pruning, providing a more conservative approach to structural plasticity.
	PRUNING_CONSERVATIVE_INACTIVITY_THRESHOLD time.Duration = 30 * time.Minute
)

// ===========================================================================
// Synaptic Transmission Constants
// ===========================================================================

const (
	// SYNAPSE_DEFAULT_TRANSMISSION_DELAY is a typical default base delay for
	// synaptic transmission, representing neurotransmitter release, diffusion,
	// receptor binding, and initial postsynaptic response.
	// Axonal conduction delay is added by the neuron/matrix.
	SYNAPSE_DEFAULT_TRANSMISSION_DELAY time.Duration = 1 * time.Millisecond

	// SYNAPSE_ACTIVITY_THRESHOLD is a common duration used to determine if a
	// synapse is "active" for general monitoring purposes.
	SYNAPSE_ACTIVITY_THRESHOLD time.Duration = 1 * time.Minute
)
