package neuron

import "time"

/*
=================================================================================
SYNAPTIC SCALING CONSTANTS - BIOLOGICAL PARAMETER DEFINITIONS
=================================================================================

This file centralizes all biological constants related to synaptic scaling
and homeostatic receptor regulation, ensuring consistency between implementation
and tests, and providing clear biological documentation for each parameter.

All constants follow the naming convention: SYNAPTIC_SCALING_[CATEGORY]_[PARAMETER]

Categories:
- TARGET: Desired activity levels and thresholds
- RATE: Speed of scaling adjustments
- INTERVAL: Timing parameters for scaling operations
- FACTOR: Multiplicative bounds for scaling
- GAIN: Receptor sensitivity limits
- ACTIVITY: Activity monitoring and gating
- THRESHOLD: Decision thresholds for scaling
- HISTORY: Data retention parameters

=================================================================================
*/

// ============================================================================
// TARGET STRENGTH CONSTANTS
// ============================================================================

const (
	// SYNAPTIC_SCALING_TARGET_STRENGTH_DEFAULT models the optimal total synaptic
	// drive for maintaining stable neural computation. This represents the
	// "homeostatic set point" that neurons try to maintain.
	// Biological Range: 0.5-2.0 depending on neuron type and network role
	SYNAPTIC_SCALING_TARGET_STRENGTH_DEFAULT = 1.0

	// SYNAPTIC_SCALING_TARGET_STRENGTH_LOW for neurons requiring lower drive
	// (e.g., inhibitory interneurons with high intrinsic excitability)
	SYNAPTIC_SCALING_TARGET_STRENGTH_LOW = 0.5

	// SYNAPTIC_SCALING_TARGET_STRENGTH_HIGH for neurons requiring higher drive
	// (e.g., motor neurons that need strong inputs to drive muscle contraction)
	SYNAPTIC_SCALING_TARGET_STRENGTH_HIGH = 2.0
)

// ============================================================================
// SCALING RATE CONSTANTS
// ============================================================================

const (
	// SYNAPTIC_SCALING_RATE_DEFAULT controls how aggressively gains are adjusted
	// toward target. Conservative rate prevents oscillations and maintains stability.
	// Biological Range: 0.0001 (very conservative) to 0.01 (aggressive)
	SYNAPTIC_SCALING_RATE_DEFAULT = 0.001

	// SYNAPTIC_SCALING_RATE_CONSERVATIVE for stable, slow adjustments
	// Suitable for mature, stable networks
	SYNAPTIC_SCALING_RATE_CONSERVATIVE = 0.0005

	// SYNAPTIC_SCALING_RATE_AGGRESSIVE for faster adaptation
	// Suitable for development or rapid learning phases
	SYNAPTIC_SCALING_RATE_AGGRESSIVE = 0.005

	// SYNAPTIC_SCALING_RATE_DEVELOPMENTAL for very rapid early adjustments
	// Models the high plasticity of developing neural circuits
	SYNAPTIC_SCALING_RATE_DEVELOPMENTAL = 0.01
)

// ============================================================================
// TIMING INTERVAL CONSTANTS
// ============================================================================

const (
	// SYNAPTIC_SCALING_INTERVAL_DEFAULT defines the standard time between
	// scaling operations. Biological timescale is much slower than STDP.
	// Biological Range: 10 seconds to 10 minutes for acute regulation
	SYNAPTIC_SCALING_INTERVAL_DEFAULT = 30 * time.Second

	// SYNAPTIC_SCALING_INTERVAL_FAST for rapid homeostatic responses
	// Models acute regulation during high activity periods
	SYNAPTIC_SCALING_INTERVAL_FAST = 10 * time.Second

	// SYNAPTIC_SCALING_INTERVAL_SLOW for long-term stability
	// Models chronic regulation for established networks
	SYNAPTIC_SCALING_INTERVAL_SLOW = 5 * time.Minute

	// SYNAPTIC_SCALING_INTERVAL_DEVELOPMENTAL for very frequent adjustments
	// Models the rapid scaling during neural development
	SYNAPTIC_SCALING_INTERVAL_DEVELOPMENTAL = 5 * time.Second
)

// ============================================================================
// SCALING FACTOR BOUNDS
// ============================================================================

const (
	// SYNAPTIC_SCALING_MIN_FACTOR prevents excessive reduction in single scaling event
	// Ensures stability by limiting how much gains can decrease per operation
	// Biological rationale: receptor trafficking has physical limits per time period
	SYNAPTIC_SCALING_MIN_FACTOR = 0.9 // Don't reduce more than 10% per event

	// SYNAPTIC_SCALING_MAX_FACTOR prevents excessive increase in single scaling event
	// Ensures stability by limiting how much gains can increase per operation
	// Biological rationale: finite receptor synthesis and trafficking capacity
	SYNAPTIC_SCALING_MAX_FACTOR = 1.1 // Don't increase more than 10% per event

	// SYNAPTIC_SCALING_MIN_FACTOR_CONSERVATIVE for very stable networks
	SYNAPTIC_SCALING_MIN_FACTOR_CONSERVATIVE = 0.95 // Max 5% reduction

	// SYNAPTIC_SCALING_MAX_FACTOR_CONSERVATIVE for very stable networks
	SYNAPTIC_SCALING_MAX_FACTOR_CONSERVATIVE = 1.05 // Max 5% increase

	// SYNAPTIC_SCALING_MIN_FACTOR_AGGRESSIVE for rapid adaptation
	SYNAPTIC_SCALING_MIN_FACTOR_AGGRESSIVE = 0.8 // Up to 20% reduction

	// SYNAPTIC_SCALING_MAX_FACTOR_AGGRESSIVE for rapid adaptation
	SYNAPTIC_SCALING_MAX_FACTOR_AGGRESSIVE = 1.2 // Up to 20% increase
)

// ============================================================================
// RECEPTOR GAIN LIMITS
// ============================================================================

const (
	// SYNAPTIC_SCALING_MIN_GAIN represents the minimum receptor sensitivity
	// Prevents complete silencing of synaptic inputs due to scaling
	// Biological basis: minimum functional receptor density
	SYNAPTIC_SCALING_MIN_GAIN = 0.01

	// SYNAPTIC_SCALING_MAX_GAIN represents the maximum receptor sensitivity
	// Prevents runaway amplification of synaptic inputs
	// Biological basis: physical limits of receptor density at synapses
	SYNAPTIC_SCALING_MAX_GAIN = 10.0

	// SYNAPTIC_SCALING_DEFAULT_GAIN for newly registered input sources
	// Represents normal receptor sensitivity before scaling adjustments
	SYNAPTIC_SCALING_DEFAULT_GAIN = 1.0

	// SYNAPTIC_SCALING_MIN_GAIN_CONSERVATIVE for safer minimum bounds
	SYNAPTIC_SCALING_MIN_GAIN_CONSERVATIVE = 0.1

	// SYNAPTIC_SCALING_MAX_GAIN_CONSERVATIVE for safer maximum bounds
	SYNAPTIC_SCALING_MAX_GAIN_CONSERVATIVE = 5.0
)

// ============================================================================
// ACTIVITY MONITORING CONSTANTS
// ============================================================================

const (
	// SYNAPTIC_SCALING_MIN_ACTIVITY minimum activity level required for scaling
	// Prevents scaling during periods of very low input activity
	// Biological basis: scaling requires sufficient calcium-dependent signaling
	SYNAPTIC_SCALING_MIN_ACTIVITY = 0.1

	// SYNAPTIC_SCALING_ACTIVITY_WINDOW_DEFAULT time window for activity sampling
	// Determines how far back to look when calculating average input strengths
	// Biological timescale: shorter than scaling interval for responsiveness
	SYNAPTIC_SCALING_ACTIVITY_WINDOW_DEFAULT = 10 * time.Second

	// SYNAPTIC_SCALING_ACTIVITY_WINDOW_SHORT for rapid activity assessment
	SYNAPTIC_SCALING_ACTIVITY_WINDOW_SHORT = 5 * time.Second

	// SYNAPTIC_SCALING_ACTIVITY_WINDOW_LONG for stable activity assessment
	SYNAPTIC_SCALING_ACTIVITY_WINDOW_LONG = 30 * time.Second

	// SYNAPTIC_SCALING_MIN_ACTIVITY_CONSERVATIVE higher threshold for scaling
	SYNAPTIC_SCALING_MIN_ACTIVITY_CONSERVATIVE = 0.2

	// SYNAPTIC_SCALING_MIN_ACTIVITY_PERMISSIVE lower threshold for scaling
	SYNAPTIC_SCALING_MIN_ACTIVITY_PERMISSIVE = 0.05
)

// ============================================================================
// DECISION THRESHOLD CONSTANTS
// ============================================================================

const (
	// SYNAPTIC_SCALING_SIGNIFICANCE_THRESHOLD minimum relative error to trigger scaling
	// Only scale for significant deviations from target (prevents noise-driven scaling)
	// Biological rationale: metabolic cost of receptor trafficking requires clear need
	SYNAPTIC_SCALING_SIGNIFICANCE_THRESHOLD = 0.1 // 10% deviation required

	// SYNAPTIC_SCALING_MIN_CHANGE minimum scaling factor change to apply
	// Prevents tiny adjustments that have no biological significance
	SYNAPTIC_SCALING_MIN_CHANGE = 0.0001

	// SYNAPTIC_SCALING_SIGNIFICANCE_THRESHOLD_STRICT for more selective scaling
	SYNAPTIC_SCALING_SIGNIFICANCE_THRESHOLD_STRICT = 0.15 // 15% deviation required

	// SYNAPTIC_SCALING_SIGNIFICANCE_THRESHOLD_PERMISSIVE for more responsive scaling
	SYNAPTIC_SCALING_SIGNIFICANCE_THRESHOLD_PERMISSIVE = 0.05 // 5% deviation triggers scaling

	// SYNAPTIC_SCALING_MIN_SOURCES minimum number of active sources for scaling
	// Ensures scaling decisions are based on sufficient input diversity
	SYNAPTIC_SCALING_MIN_SOURCES = 2
)

// ============================================================================
// HISTORY AND MONITORING CONSTANTS
// ============================================================================

const (
	// SYNAPTIC_SCALING_HISTORY_SIZE maximum number of scaling events to remember
	// Used for analysis, debugging, and detecting scaling oscillations
	SYNAPTIC_SCALING_HISTORY_SIZE = 100

	// SYNAPTIC_SCALING_CLEANUP_INTERVAL how often to clean old activity data
	// Prevents unlimited memory growth while maintaining useful history
	SYNAPTIC_SCALING_CLEANUP_INTERVAL = 1 * time.Minute

	// SYNAPTIC_SCALING_HISTORY_SIZE_COMPACT for memory-constrained environments
	SYNAPTIC_SCALING_HISTORY_SIZE_COMPACT = 20

	// SYNAPTIC_SCALING_HISTORY_SIZE_EXTENDED for detailed analysis
	SYNAPTIC_SCALING_HISTORY_SIZE_EXTENDED = 500
)

// ============================================================================
// BIOLOGICAL REALISM CONSTANTS
// ============================================================================

const (
	// SYNAPTIC_SCALING_CALCIUM_THRESHOLD minimum calcium level for scaling
	// Models the calcium-dependent gene expression required for receptor trafficking
	// Biological basis: calcium-calmodulin-dependent kinase activity thresholds
	SYNAPTIC_SCALING_CALCIUM_THRESHOLD = 0.5

	// SYNAPTIC_SCALING_METABOLIC_COST relative metabolic cost of receptor trafficking
	// Models the ATP cost of synthesizing and trafficking receptors
	// Higher values make scaling more selective (occurs only when really needed)
	SYNAPTIC_SCALING_METABOLIC_COST = 0.1

	// SYNAPTIC_SCALING_RECEPTOR_SYNTHESIS_RATE maximum rate of new receptor production
	// Models the biological limits of protein synthesis for receptor scaling
	// Constrains how rapidly receptor density can change
	SYNAPTIC_SCALING_RECEPTOR_SYNTHESIS_RATE = 0.05 // 5% of total receptors per scaling event

	// SYNAPTIC_SCALING_RECEPTOR_DEGRADATION_RATE maximum rate of receptor removal
	// Models the biological limits of receptor endocytosis and degradation
	SYNAPTIC_SCALING_RECEPTOR_DEGRADATION_RATE = 0.05 // 5% of total receptors per scaling event
)

// ============================================================================
// DEVELOPMENTAL CONSTANTS
// ============================================================================

const (
	// SYNAPTIC_SCALING_DEVELOPMENTAL_PERIOD duration of high plasticity phase
	// During development, scaling parameters are more aggressive
	// Biological basis: critical periods of high synaptic plasticity
	SYNAPTIC_SCALING_DEVELOPMENTAL_PERIOD = 24 * time.Hour

	// SYNAPTIC_SCALING_MATURATION_FACTOR how much scaling slows after development
	// Mature neurons have reduced scaling rates for stability
	SYNAPTIC_SCALING_MATURATION_FACTOR = 0.5

	// SYNAPTIC_SCALING_CRITICAL_PERIOD_FACTOR scaling enhancement during critical periods
	// Some developmental phases have enhanced scaling for circuit refinement
	SYNAPTIC_SCALING_CRITICAL_PERIOD_FACTOR = 2.0
)

// ============================================================================
// NETWORK COORDINATION CONSTANTS
// ============================================================================

const (
	// SYNAPTIC_SCALING_NETWORK_COORDINATION_WINDOW time window for network-wide scaling
	// When coordinating scaling across multiple neurons in a network
	SYNAPTIC_SCALING_NETWORK_COORDINATION_WINDOW = 5 * time.Second

	// SYNAPTIC_SCALING_GLOBAL_INHIBITION_FACTOR how much to reduce scaling during high network activity
	// Prevents all neurons from scaling simultaneously during network bursts
	SYNAPTIC_SCALING_GLOBAL_INHIBITION_FACTOR = 0.5

	// SYNAPTIC_SCALING_LOCAL_COMPETITION_RADIUS spatial range for competitive scaling
	// Neurons within this range compete for limited scaling resources
	SYNAPTIC_SCALING_LOCAL_COMPETITION_RADIUS = 100.0 // Micrometers in biological space
)

// ============================================================================
// ERROR HANDLING AND SAFETY CONSTANTS
// ============================================================================

const (
	// SYNAPTIC_SCALING_MAX_CONSECUTIVE_INCREASES maximum scaling increases in a row
	// Prevents runaway scaling by limiting consecutive increases
	SYNAPTIC_SCALING_MAX_CONSECUTIVE_INCREASES = 5

	// SYNAPTIC_SCALING_MAX_CONSECUTIVE_DECREASES maximum scaling decreases in a row
	// Prevents complete silencing by limiting consecutive decreases
	SYNAPTIC_SCALING_MAX_CONSECUTIVE_DECREASES = 5

	// SYNAPTIC_SCALING_OSCILLATION_DETECTION_WINDOW time window for detecting oscillations
	// Used to identify and prevent scaling oscillations
	SYNAPTIC_SCALING_OSCILLATION_DETECTION_WINDOW = 10 * time.Minute

	// SYNAPTIC_SCALING_EMERGENCY_STOP_THRESHOLD factor change that triggers emergency stop
	// If scaling tries to change gains by this much, something is wrong
	SYNAPTIC_SCALING_EMERGENCY_STOP_THRESHOLD = 0.5 // 50% change triggers safety stop
)

// ============================================================================
// TESTING AND DEBUGGING CONSTANTS
// ============================================================================

const (
	// SYNAPTIC_SCALING_TEST_TOLERANCE tolerance for testing scaling calculations
	SYNAPTIC_SCALING_TEST_TOLERANCE = 0.001

	// SYNAPTIC_SCALING_TEST_ACTIVITY_SAMPLES number of activity samples for testing
	SYNAPTIC_SCALING_TEST_ACTIVITY_SAMPLES = 50

	// SYNAPTIC_SCALING_TEST_SOURCES number of input sources for testing
	SYNAPTIC_SCALING_TEST_SOURCES = 10

	// SYNAPTIC_SCALING_TEST_INTERVAL short interval for accelerated testing
	SYNAPTIC_SCALING_TEST_INTERVAL = 100 * time.Millisecond

	// SYNAPTIC_SCALING_DEBUG_VERBOSE_LOGGING whether to enable detailed logging
	SYNAPTIC_SCALING_DEBUG_VERBOSE_LOGGING = false
)
