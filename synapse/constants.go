package synapse

import "time"

// =================================================================================
// SYNAPSE CONSTANTS - BIOLOGICALLY PLAUSIBLE PARAMETER DEFINITIONS
// =================================================================================

// Default values for STDP (Spike-Timing Dependent Plasticity)
const (
	// STDP_DEFAULT_LEARNING_RATE is the standard synaptic weight change rate
	// Typical biological values range from 0.001-0.05 based on synaptic type
	STDP_DEFAULT_LEARNING_RATE float64 = 0.01

	// STDP_DEFAULT_TIME_CONSTANT defines the exponential decay of STDP effects
	// Based on calcium signaling dynamics in dendritic spines (10-30ms)
	STDP_DEFAULT_TIME_CONSTANT time.Duration = 20 * time.Millisecond

	// STDP_DEFAULT_WINDOW_SIZE is the maximum timing difference for plasticity
	// Measured in vitro for cortical pyramidal cells (50-100ms)
	STDP_DEFAULT_WINDOW_SIZE time.Duration = 100 * time.Millisecond

	// STDP_DEFAULT_MIN_WEIGHT prevents complete elimination of synapses
	// Biologically equivalent to silent/minimal synaptic efficacy
	STDP_DEFAULT_MIN_WEIGHT float64 = 0.001

	// STDP_DEFAULT_MAX_WEIGHT prevents runaway potentiation
	// Corresponds to maximum conductance of dendritic spines
	STDP_DEFAULT_MAX_WEIGHT float64 = 2.0

	// STDP_DEFAULT_ASYMMETRY_RATIO controls the relative strength of LTD vs LTP
	// Values >1.0 mean LTD is stronger than LTP (typical in cortical synapses)
	STDP_DEFAULT_ASYMMETRY_RATIO float64 = 1.2
)

// Default values for pruning (structural plasticity)
const (
	// PRUNING_DEFAULT_WEIGHT_THRESHOLD defines when synapses become candidates
	// for elimination due to weakness
	PRUNING_DEFAULT_WEIGHT_THRESHOLD float64 = 0.05

	// PRUNING_DEFAULT_INACTIVITY_THRESHOLD defines how long an inactive synapse
	// must remain inactive before becoming a pruning candidate
	PRUNING_DEFAULT_INACTIVITY_THRESHOLD time.Duration = 30 * time.Second

	// Conservative (slower) pruning values
	PRUNING_CONSERVATIVE_WEIGHT_THRESHOLD     float64       = 0.02
	PRUNING_CONSERVATIVE_INACTIVITY_THRESHOLD time.Duration = 120 * time.Second

	// Default synapse activity threshold for general monitoring
	SYNAPSE_ACTIVITY_THRESHOLD time.Duration = 1 * time.Minute

	// Default transmission delay (axonal + synaptic) for short-range connections
	// This models the minimum delay for local circuit connections.
	// Axonal conduction delay is added by the neuron/matrix.
	SYNAPSE_DEFAULT_TRANSMISSION_DELAY time.Duration = 1 * time.Millisecond
)

// ELIGIBILITY_TRACE_CONSTANTS defines constants for eligibility trace mechanisms
// which are essential for reinforcement learning in biological systems
const (
	// Default decay time for eligibility traces (400-800ms in biological systems)
	ELIGIBILITY_TRACE_DEFAULT_DECAY time.Duration = 500 * time.Millisecond

	// Maximum trace value to prevent runaway traces
	ELIGIBILITY_TRACE_MAX_VALUE float64 = 5.0

	// Minimum trace value to consider significant for learning
	ELIGIBILITY_TRACE_THRESHOLD float64 = 0.01
)

// PruningConfig defines structural plasticity parameters
// Used to configure when and how synapses are eliminated
type PruningConfig struct {
	Enabled             bool          `json:"enabled"`              // Whether pruning is active
	WeightThreshold     float64       `json:"weight_threshold"`     // Minimum weight to avoid pruning
	InactivityThreshold time.Duration `json:"inactivity_threshold"` // Time since last activity to prune
}
