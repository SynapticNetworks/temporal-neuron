package neuron

import "time"

/*
=================================================================================
DENDRITIC INTEGRATION CONSTANTS - BIOLOGICAL PARAMETER DEFINITIONS
=================================================================================

This file centralizes all biological constants used in dendritic integration
modes to ensure consistency between implementation and tests, and to provide
clear biological documentation for each parameter.

All constants follow the naming convention: DENDRITE_[CATEGORY]_[PARAMETER]

Categories:
- VOLTAGE: Membrane potentials and thresholds (mV)
- TIME: Temporal constants and delays (time.Duration)
- CURRENT: Current amplitudes and thresholds (pA)
- FACTOR: Dimensionless scaling factors and ratios
- NOISE: Noise levels and variability parameters
- SPATIAL: Distance-dependent attenuation factors

=================================================================================
*/

// ============================================================================
// MEMBRANE VOLTAGE CONSTANTS (mV)
// ============================================================================

const (
	// Resting and baseline potentials
	DENDRITE_VOLTAGE_RESTING_CORTICAL    = -70.0 // Typical cortical resting potential
	DENDRITE_VOLTAGE_RESTING_HIPPOCAMPAL = -65.0 // Hippocampal CA1 resting potential
	DENDRITE_VOLTAGE_RESTING_INTERNEURON = -75.0 // Fast-spiking interneuron resting

	// Dendritic spike thresholds
	DENDRITE_VOLTAGE_SPIKE_THRESHOLD_DEFAULT = -40.0 // Default dendritic spike threshold
	DENDRITE_VOLTAGE_SPIKE_THRESHOLD_STRICT  = -45.0 // More restrictive threshold
	DENDRITE_VOLTAGE_SPIKE_THRESHOLD_LENIENT = -35.0 // More permissive threshold

	// Ion channel reversal potentials
	DENDRITE_VOLTAGE_REVERSAL_SODIUM    = 60.0  // ENa (sodium reversal)
	DENDRITE_VOLTAGE_REVERSAL_POTASSIUM = -90.0 // EK (potassium reversal)
	DENDRITE_VOLTAGE_REVERSAL_CALCIUM   = 120.0 // ECa (calcium reversal)
	DENDRITE_VOLTAGE_REVERSAL_CHLORIDE  = -70.0 // ECl (chloride reversal)
	DENDRITE_VOLTAGE_REVERSAL_MIXED     = 0.0   // Mixed cation reversal
)

// ============================================================================
// TEMPORAL CONSTANTS (time.Duration)
// ============================================================================

const (
	// Membrane time constants (τ = Rm × Cm)
	DENDRITE_TIME_CONSTANT_CORTICAL    = 20 * time.Millisecond // Cortical pyramidal τ
	DENDRITE_TIME_CONSTANT_HIPPOCAMPAL = 35 * time.Millisecond // Hippocampal CA1 τ
	DENDRITE_TIME_CONSTANT_INTERNEURON = 8 * time.Millisecond  // Fast interneuron τ

	// Branch-specific time constants
	DENDRITE_TIME_CONSTANT_APICAL   = 25 * time.Millisecond // Apical dendrite τ
	DENDRITE_TIME_CONSTANT_BASAL    = 15 * time.Millisecond // Basal dendrite τ
	DENDRITE_TIME_CONSTANT_DISTAL   = 30 * time.Millisecond // Distal branch τ
	DENDRITE_TIME_CONSTANT_PROXIMAL = 10 * time.Millisecond // Proximal branch τ

	// Processing and integration intervals
	DENDRITE_TIME_DECAY_TICK       = 1 * time.Millisecond   // Membrane decay update
	DENDRITE_TIME_HOMEOSTATIC_TICK = 100 * time.Millisecond // Homeostatic update
	DENDRITE_TIME_SCALING_INTERVAL = 30 * time.Second       // Synaptic scaling

	// Biological timing variability
	DENDRITE_TIME_JITTER_CORTICAL    = 500 * time.Microsecond // Cortical timing jitter
	DENDRITE_TIME_JITTER_HIPPOCAMPAL = 300 * time.Microsecond // Hippocampal jitter
	DENDRITE_TIME_JITTER_INTERNEURON = 1 * time.Millisecond   // Interneuron jitter

	// Ion channel kinetics
	DENDRITE_TIME_CHANNEL_ACTIVATION   = 1 * time.Millisecond  // Channel activation τ
	DENDRITE_TIME_CHANNEL_DEACTIVATION = 2 * time.Millisecond  // Channel deactivation τ
	DENDRITE_TIME_CHANNEL_INACTIVATION = 10 * time.Millisecond // Channel inactivation τ
	DENDRITE_TIME_CHANNEL_RECOVERY     = 5 * time.Millisecond  // Channel recovery τ
)

// ============================================================================
// CURRENT CONSTANTS (pA - picoAmperes)
// ============================================================================

const (
	// Synaptic current limits and thresholds
	DENDRITE_CURRENT_SATURATION_DEFAULT = 2.0 // Default maximum synaptic effect
	DENDRITE_CURRENT_SATURATION_STRONG  = 3.0 // Strong saturation limit
	DENDRITE_CURRENT_SATURATION_WEAK    = 1.5 // Weak saturation limit

	// Dendritic spike parameters
	DENDRITE_CURRENT_SPIKE_THRESHOLD_DEFAULT = 1.5 // Default spike threshold
	DENDRITE_CURRENT_SPIKE_THRESHOLD_HIGH    = 2.0 // High spike threshold
	DENDRITE_CURRENT_SPIKE_THRESHOLD_LOW     = 1.0 // Low spike threshold

	DENDRITE_CURRENT_SPIKE_AMPLITUDE_DEFAULT = 1.0 // Default NMDA spike amplitude
	DENDRITE_CURRENT_SPIKE_AMPLITUDE_STRONG  = 1.5 // Strong spike amplitude
	DENDRITE_CURRENT_SPIKE_AMPLITUDE_WEAK    = 0.5 // Weak spike amplitude

	// Biological current limits
	DENDRITE_CURRENT_MAX_BIOLOGICAL = 100.0  // Maximum realistic current
	DENDRITE_CURRENT_MIN_BIOLOGICAL = -100.0 // Maximum hyperpolarizing current
	DENDRITE_CURRENT_NOISE_FLOOR    = 0.001  // Minimum significant current

	// Ion channel conductances (pS - picoSiemens)
	DENDRITE_CONDUCTANCE_SODIUM_DEFAULT    = 20.0 // Default Na+ channel conductance
	DENDRITE_CONDUCTANCE_POTASSIUM_DEFAULT = 10.0 // Default K+ channel conductance
	DENDRITE_CONDUCTANCE_CALCIUM_DEFAULT   = 5.0  // Default Ca2+ channel conductance
)

// ============================================================================
// SCALING FACTORS AND RATIOS (dimensionless)
// ============================================================================

const (
	// Spatial attenuation factors
	DENDRITE_FACTOR_SPATIAL_DECAY_DEFAULT = 0.1  // Default spatial decay per unit
	DENDRITE_FACTOR_SPATIAL_DECAY_STRONG  = 0.2  // Strong spatial decay
	DENDRITE_FACTOR_SPATIAL_DECAY_WEAK    = 0.05 // Weak spatial decay

	// Spatial weights by dendritic location
	DENDRITE_FACTOR_WEIGHT_PROXIMAL = 1.0 // No attenuation for proximal
	DENDRITE_FACTOR_WEIGHT_BASAL    = 0.8 // Slight attenuation for basal
	DENDRITE_FACTOR_WEIGHT_APICAL   = 0.7 // Moderate attenuation for apical
	DENDRITE_FACTOR_WEIGHT_DISTAL   = 0.5 // Strong attenuation for distal

	// Shunting inhibition parameters
	DENDRITE_FACTOR_SHUNTING_DEFAULT = 0.5 // Default shunting strength
	DENDRITE_FACTOR_SHUNTING_STRONG  = 0.8 // Strong shunting
	DENDRITE_FACTOR_SHUNTING_WEAK    = 0.3 // Weak shunting
	DENDRITE_FACTOR_SHUNTING_FLOOR   = 0.1 // Minimum shunting factor

	// Synaptic scaling parameters
	DENDRITE_FACTOR_SCALING_RATE_DEFAULT   = 0.001 // Default scaling rate
	DENDRITE_FACTOR_SCALING_TARGET_DEFAULT = 1.0   // Target input strength
	DENDRITE_FACTOR_SCALING_MIN            = 0.9   // Minimum scaling factor
	DENDRITE_FACTOR_SCALING_MAX            = 1.1   // Maximum scaling factor

	// Chemical effect multipliers by neurotransmitter
	DENDRITE_FACTOR_EFFECT_GLUTAMATE     = 1.0  // Glutamate effect multiplier
	DENDRITE_FACTOR_EFFECT_GABA          = -0.8 // GABA effect multiplier
	DENDRITE_FACTOR_EFFECT_DOPAMINE      = 0.5  // Dopamine effect multiplier
	DENDRITE_FACTOR_EFFECT_SEROTONIN     = 0.3  // Serotonin effect multiplier
	DENDRITE_FACTOR_EFFECT_ACETYLCHOLINE = 0.7  // Acetylcholine effect multiplier

	// Homeostatic parameters
	DENDRITE_FACTOR_HOMEOSTASIS_STRENGTH_DEFAULT = 0.2 // Default homeostatic strength
	DENDRITE_FACTOR_HOMEOSTASIS_STRENGTH_STRONG  = 0.5 // Strong homeostasis
	DENDRITE_FACTOR_HOMEOSTASIS_STRENGTH_WEAK    = 0.1 // Weak homeostasis
	DENDRITE_FACTOR_THRESHOLD_MIN_RATIO          = 0.1 // Min threshold as fraction of base
	DENDRITE_FACTOR_THRESHOLD_MAX_RATIO          = 5.0 // Max threshold as fraction of base

	// Ion channel activation parameters
	DENDRITE_FACTOR_CALCIUM_INCREMENT  = 1.0    // Calcium increment per spike
	DENDRITE_FACTOR_CALCIUM_DECAY      = 0.9995 // Calcium decay rate per tick
	DENDRITE_FACTOR_CHANNEL_ACTIVATION = 0.5    // Channel activation threshold
)

// ============================================================================
// NOISE AND VARIABILITY PARAMETERS
// ============================================================================

const (
	// Membrane noise levels (as fraction of signal)
	DENDRITE_NOISE_MEMBRANE_CORTICAL    = 0.01  // 1% noise for cortical
	DENDRITE_NOISE_MEMBRANE_HIPPOCAMPAL = 0.005 // 0.5% noise for hippocampal
	DENDRITE_NOISE_MEMBRANE_INTERNEURON = 0.02  // 2% noise for interneuron
	DENDRITE_NOISE_MEMBRANE_DISABLED    = 0.0   // No noise (for testing)

	// Activity tracking and cleanup
	DENDRITE_ACTIVITY_TRACKING_WINDOW  = 10 * time.Second // Activity tracking window
	DENDRITE_ACTIVITY_MIN_SCALING      = 0.1              // Minimum activity for scaling
	DENDRITE_ACTIVITY_CLEANUP_INTERVAL = 1 * time.Minute  // Activity cleanup interval

	// Buffer and processing limits
	DENDRITE_BUFFER_DEFAULT_CAPACITY = 100  // Default message buffer size
	DENDRITE_BUFFER_LARGE_CAPACITY   = 1000 // Large buffer for high activity
	DENDRITE_BUFFER_HISTORY_CAPACITY = 100  // Firing history buffer size
)

// ============================================================================
// CALCIUM AND IONIC CONCENTRATIONS
// ============================================================================

const (
	// Baseline ionic concentrations
	DENDRITE_CALCIUM_BASELINE_INTRACELLULAR   = 0.1   // Baseline [Ca2+]i (μM)
	DENDRITE_CALCIUM_BASELINE_EXTRACELLULAR   = 2.0   // Baseline [Ca2+]o (mM)
	DENDRITE_SODIUM_BASELINE_INTRACELLULAR    = 10.0  // Baseline [Na+]i (mM)
	DENDRITE_POTASSIUM_BASELINE_INTRACELLULAR = 140.0 // Baseline [K+]i (mM)

	// Calcium-dependent thresholds
	DENDRITE_CALCIUM_THRESHOLD_CHANNEL    = 0.5 // [Ca2+] threshold for Ca2+-activated channels
	DENDRITE_CALCIUM_THRESHOLD_PLASTICITY = 1.0 // [Ca2+] threshold for plasticity

	// ATP and metabolic parameters
	DENDRITE_ATP_BASELINE              = 1.0 // Baseline ATP level
	DENDRITE_METABOLIC_STRESS_BASELINE = 0.0 // Baseline metabolic stress
)

// ============================================================================
// NEUROTRANSMITTER CONCENTRATION SCALING
// ============================================================================

const (
	// Base concentration scaling factors
	DENDRITE_CONCENTRATION_SCALE_BASE = 0.1 // Base scaling for concentration calculation

	// Neurotransmitter-specific concentration factors
	DENDRITE_CONCENTRATION_FACTOR_GLUTAMATE = 1.0 // Glutamate concentration factor
	DENDRITE_CONCENTRATION_FACTOR_GABA      = 0.8 // GABA concentration factor
	DENDRITE_CONCENTRATION_FACTOR_DOPAMINE  = 0.3 // Dopamine concentration factor
	DENDRITE_CONCENTRATION_FACTOR_DEFAULT   = 0.5 // Default concentration factor
)

// ============================================================================
// TEST-SPECIFIC CONSTANTS
// ============================================================================

const (
	// Test tolerance values
	DENDRITE_TEST_TOLERANCE_CURRENT = 0.001 // Current comparison tolerance (pA)
	DENDRITE_TEST_TOLERANCE_VOLTAGE = 0.1   // Voltage comparison tolerance (mV)
	DENDRITE_TEST_TOLERANCE_FACTOR  = 0.01  // Factor comparison tolerance

	// Test timing parameters
	DENDRITE_TEST_DECAY_WAIT    = 10 * time.Millisecond // Wait time for decay testing
	DENDRITE_TEST_PROCESS_DELAY = 1 * time.Millisecond  // Processing delay for tests

	// Test input values
	DENDRITE_TEST_INPUT_SMALL  = 0.01 // Small test input
	DENDRITE_TEST_INPUT_MEDIUM = 1.0  // Medium test input
	DENDRITE_TEST_INPUT_LARGE  = 10.0 // Large test input

	// Concurrency test parameters
	DENDRITE_TEST_GOROUTINES           = 50 // Number of test goroutines
	DENDRITE_TEST_INPUTS_PER_GOROUTINE = 20 // Inputs per goroutine
)
