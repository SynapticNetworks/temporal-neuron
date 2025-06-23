package neuron

import "time"

/*
=================================================================================
ENHANCED NEURON FACTORY CONSTANTS - BIOLOGICAL PARAMETER DEFINITIONS
=================================================================================

This file defines all constants for the enhanced neuron factory system,
including plasticity, homeostasis, and learning parameters. All values are
based on experimental neuroscience data and biological constraints.

Categories:
- STDP: Spike-timing dependent plasticity parameters
- HOMEOSTASIS: Homeostatic scaling and regulation
- PRUNING: Structural plasticity and synapse elimination
- NEURON_TYPES: Type-specific parameter sets
- TIMING: Biological timing constraints
- LEARNING: Learning rate and adaptation parameters

=================================================================================
*/

// ============================================================================
// STDP (SPIKE-TIMING DEPENDENT PLASTICITY) CONSTANTS
// ============================================================================

const (
	// STDP_FEEDBACK_DELAY_DEFAULT is the standard delay after firing before
	// sending STDP feedback to recently active synapses.
	// Biological basis: Time for calcium signaling to propagate and activate
	// plasticity machinery. Typical range: 1-10ms.
	STDP_FEEDBACK_DELAY_DEFAULT = 5 * time.Millisecond

	// STDP_FEEDBACK_DELAY_FAST for neurons requiring rapid plasticity
	// (e.g., sensory processing, temporal pattern detection)
	STDP_FEEDBACK_DELAY_FAST = 1 * time.Millisecond

	// STDP_FEEDBACK_DELAY_SLOW for neurons requiring stable, conservative learning
	// (e.g., motor control, long-term memory storage)
	STDP_FEEDBACK_DELAY_SLOW = 20 * time.Millisecond

	// STDP_LEARNING_RATE_DEFAULT models typical cortical synapse plasticity
	// Biological range: 0.001-0.1 (0.1%-10% weight change per event)
	STDP_LEARNING_RATE_DEFAULT = 0.01 // 1% weight change per STDP event

	// STDP_LEARNING_RATE_EXCITATORY for glutamatergic synapses
	// Excitatory synapses typically show stronger plasticity
	STDP_LEARNING_RATE_EXCITATORY = 0.015

	// STDP_LEARNING_RATE_INHIBITORY for GABAergic synapses
	// Inhibitory plasticity is typically more conservative
	STDP_LEARNING_RATE_INHIBITORY = 0.005

	// STDP_LEARNING_RATE_AGGRESSIVE for rapid learning scenarios
	STDP_LEARNING_RATE_AGGRESSIVE = 0.02

	// STDP_LEARNING_RATE_CONSERVATIVE for stable, slow learning
	STDP_LEARNING_RATE_CONSERVATIVE = 0.001
)

// ============================================================================
// HOMEOSTATIC SCALING CONSTANTS
// ============================================================================

const (
	// HOMEOSTASIS_CHECK_INTERVAL_DEFAULT is how often neurons evaluate
	// their activity levels for homeostatic adjustments.
	// Biological basis: Homeostatic mechanisms operate on slower timescales
	// than synaptic transmission (seconds to minutes vs milliseconds)
	HOMEOSTASIS_CHECK_INTERVAL_DEFAULT = 10 * time.Second

	// HOMEOSTASIS_CHECK_INTERVAL_FAST for highly dynamic networks
	HOMEOSTASIS_CHECK_INTERVAL_FAST = 5 * time.Second

	// HOMEOSTASIS_CHECK_INTERVAL_SLOW for stable, long-term networks
	HOMEOSTASIS_CHECK_INTERVAL_SLOW = 30 * time.Second

	// HOMEOSTASIS_TARGET_INPUT_STRENGTH_DEFAULT models optimal total
	// synaptic drive for typical cortical neurons
	HOMEOSTASIS_TARGET_INPUT_STRENGTH_DEFAULT = 1.0

	// HOMEOSTASIS_SCALING_RATE_DEFAULT controls how aggressively
	// synaptic weights are adjusted during homeostatic scaling
	HOMEOSTASIS_SCALING_RATE_DEFAULT = 0.001

	// HOMEOSTASIS_SCALING_INTERVAL_DEFAULT is the time between
	// homeostatic scaling operations
	HOMEOSTASIS_SCALING_INTERVAL_DEFAULT = 30 * time.Second
)

// ============================================================================
// STRUCTURAL PLASTICITY (PRUNING) CONSTANTS
// ============================================================================

const (
	// PRUNING_CHECK_INTERVAL_DEFAULT is how often neurons evaluate
	// their synapses for potential elimination.
	// Biological basis: Structural plasticity operates on very slow
	// timescales (minutes to hours) compared to functional plasticity
	PRUNING_CHECK_INTERVAL_DEFAULT = 60 * time.Second

	// PRUNING_CHECK_INTERVAL_CONSERVATIVE for networks where connection
	// stability is critical
	PRUNING_CHECK_INTERVAL_CONSERVATIVE = 300 * time.Second // 5 minutes

	// PRUNING_CHECK_INTERVAL_AGGRESSIVE for dynamic networks that
	// need rapid structural adaptation
	PRUNING_CHECK_INTERVAL_AGGRESSIVE = 30 * time.Second

	// PRUNING_WEIGHT_THRESHOLD_DEFAULT defines when synapses are
	// considered weak enough for potential elimination
	PRUNING_WEIGHT_THRESHOLD_DEFAULT = 0.01

	// PRUNING_INACTIVITY_THRESHOLD_DEFAULT is how long a weak synapse
	// must be inactive before it becomes eligible for pruning
	PRUNING_INACTIVITY_THRESHOLD_DEFAULT = 5 * time.Minute
)

// ============================================================================
// NEURON TYPE-SPECIFIC CONSTANTS
// ============================================================================

const (
	// EXCITATORY_THRESHOLD_DEFAULT for glutamatergic neurons
	// Biological basis: Pyramidal neurons typically have moderate thresholds
	EXCITATORY_THRESHOLD_DEFAULT = 1.0

	// INHIBITORY_THRESHOLD_DEFAULT for GABAergic interneurons
	// Biological basis: Interneurons often have lower thresholds for rapid response
	INHIBITORY_THRESHOLD_DEFAULT = 0.8

	// EXCITATORY_DECAY_RATE_DEFAULT models membrane properties of pyramidal cells
	EXCITATORY_DECAY_RATE_DEFAULT = 0.95

	// INHIBITORY_DECAY_RATE_DEFAULT models faster membrane dynamics of interneurons
	INHIBITORY_DECAY_RATE_DEFAULT = 0.92

	// EXCITATORY_REFRACTORY_PERIOD_DEFAULT for typical cortical pyramidal neurons
	EXCITATORY_REFRACTORY_PERIOD_DEFAULT = 10 * time.Millisecond

	// INHIBITORY_REFRACTORY_PERIOD_DEFAULT for fast-spiking interneurons
	INHIBITORY_REFRACTORY_PERIOD_DEFAULT = 5 * time.Millisecond

	// EXCITATORY_FIRE_FACTOR_DEFAULT for standard pyramidal neuron output
	EXCITATORY_FIRE_FACTOR_DEFAULT = 1.0

	// INHIBITORY_FIRE_FACTOR_DEFAULT for interneuron output scaling
	INHIBITORY_FIRE_FACTOR_DEFAULT = 1.2 // Slightly stronger to provide effective inhibition
)

// ============================================================================
// HOMEOSTATIC TARGET RATES (NEURON TYPE-SPECIFIC)
// ============================================================================

const (
	// EXCITATORY_TARGET_RATE_DEFAULT for cortical pyramidal neurons
	// Biological range: 1-10 Hz for typical cortical activity
	EXCITATORY_TARGET_RATE_DEFAULT = 5.0

	// INHIBITORY_TARGET_RATE_DEFAULT for GABAergic interneurons
	// Biological range: 10-50 Hz for interneurons (higher than pyramidal)
	INHIBITORY_TARGET_RATE_DEFAULT = 15.0

	// MOTOR_NEURON_TARGET_RATE for motor neurons (higher activity)
	MOTOR_NEURON_TARGET_RATE = 25.0

	// SENSORY_NEURON_TARGET_RATE for sensory processing neurons
	SENSORY_NEURON_TARGET_RATE = 8.0

	// HOMEOSTASIS_STRENGTH_DEFAULT controls how aggressively neurons
	// adjust their thresholds to maintain target rates
	HOMEOSTASIS_STRENGTH_DEFAULT = 0.2

	// HOMEOSTASIS_STRENGTH_AGGRESSIVE for rapid homeostatic correction
	HOMEOSTASIS_STRENGTH_AGGRESSIVE = 0.5

	// HOMEOSTASIS_STRENGTH_CONSERVATIVE for gentle homeostatic adjustment
	HOMEOSTASIS_STRENGTH_CONSERVATIVE = 0.1
)

// ============================================================================
// CHEMICAL SIGNALING CONSTANTS
// ============================================================================

const (
	// CHEMICAL_RELEASE_SCALING_FACTOR converts neural firing strength
	// to neurotransmitter concentration for volume transmission
	CHEMICAL_RELEASE_SCALING_FACTOR = 0.1

	// GLUTAMATE_CONCENTRATION_DEFAULT for excitatory volume transmission
	GLUTAMATE_CONCENTRATION_DEFAULT = 1.0

	// GABA_CONCENTRATION_DEFAULT for inhibitory volume transmission
	GABA_CONCENTRATION_DEFAULT = 0.8

	// DOPAMINE_CONCENTRATION_DEFAULT for modulatory signaling
	DOPAMINE_CONCENTRATION_DEFAULT = 0.5
)

// ============================================================================
// NETWORK COORDINATION CONSTANTS
// ============================================================================

const (
	// HEALTH_REPORT_THRESHOLD defines minimum activity level change
	// that triggers a health report to the matrix
	HEALTH_REPORT_THRESHOLD = 0.1

	// CONNECTION_COUNT_REPORT_THRESHOLD defines minimum connection count
	// change that triggers a connectivity report
	CONNECTION_COUNT_REPORT_THRESHOLD = 5

	// SPATIAL_DELAY_DEFAULT for neurons when matrix spatial delays unavailable
	SPATIAL_DELAY_DEFAULT = 1 * time.Millisecond
)
