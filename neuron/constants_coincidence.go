package neuron

import "time"

/*
=================================================================================
COINCIDENCE DETECTION CONSTANTS - BIOLOGICAL PARAMETER DEFINITIONS
=================================================================================

This file centralizes all biological constants used in coincidence detection
mechanisms within the dendrite. This ensures consistency between implementation
and tests, and provides clear biological documentation for each parameter.

All constants follow the naming convention: COINCIDENCE_[CATEGORY]_[PARAMETER]

Categories:
- TEMPORAL: Time windows and delays for input simultaneity
- THRESHOLD: Criteria for detecting coincident events (e.g., voltage, current)
- AMPLIFICATION: Factors for non-linear boosting of coincident signals
- MECHANISM: Parameters specific to certain biophysical detection mechanisms

=================================================================================
*/

// ============================================================================
// TEMPORAL COINCIDENCE WINDOWS (time.Duration)
// ============================================================================

const (
	// COINCIDENCE_TEMPORAL_WINDOW_DEFAULT defines the typical time window within
	// which inputs are considered "coincident" for summation. This reflects
	// the neuron's integration time constant and the duration of EPSP/IPSPs.
	// Biological Range: 1-10 ms for rapid coincidence detection.
	COINCIDENCE_TEMPORAL_WINDOW_DEFAULT = 5 * time.Millisecond

	// COINCIDENCE_TEMPORAL_WINDOW_SHORT for very precise, fast coincidence.
	// E.g., for detecting highly synchronized inputs.
	COINCIDENCE_TEMPORAL_WINDOW_SHORT = 2 * time.Millisecond

	// COINCIDENCE_TEMPORAL_WINDOW_LONG for broader, slower coincidence.
	// E.g., for integrating inputs over a slightly longer period.
	COINCIDENCE_TEMPORAL_WINDOW_LONG = 10 * time.Millisecond

	// COINCIDENCE_TEMPORAL_JITTER_TOLERANCE defines the acceptable timing jitter
	// between inputs for them to still be considered coincident.
	COINCIDENCE_TEMPORAL_JITTER_TOLERANCE = 1 * time.Millisecond
)

// ============================================================================
// DETECTION THRESHOLDS (mV, pA, dimensionless)
// ============================================================================

const (
	// COINCIDENCE_CURRENT_THRESHOLD_DEFAULT is the default summed input current
	// required to trigger a non-linear coincident event (e.g., a dendritic spike).
	COINCIDENCE_CURRENT_THRESHOLD_DEFAULT = 1.8 // pA (similar to DENDRITE_CURRENT_SPIKE_THRESHOLD_DEFAULT)

	// COINCIDENCE_VOLTAGE_THRESHOLD_DEFAULT is the default membrane voltage level
	// that must be reached for certain voltage-gated coincidence mechanisms to activate.
	COINCIDENCE_VOLTAGE_THRESHOLD_DEFAULT = -45.0 // mV (similar to DENDRITE_VOLTAGE_SPIKE_THRESHOLD_STRICT)

	// COINCIDENCE_MIN_INPUTS_REQUIRED specifies the minimum number of distinct
	// coincident inputs needed to trigger a detection event.
	COINCIDENCE_MIN_INPUTS_REQUIRED = 2 // At least two inputs for a coincidence

	// COINCIDENCE_TRIGGER_SENSITIVITY controls how sensitive the detector is
	// to the combined strength of coincident inputs. Higher values mean
	// stronger inputs are needed.
	COINCIDENCE_TRIGGER_SENSITIVITY = 0.5 // Dimensionless (e.g., for a sigmoid activation)
)

// ============================================================================
// AMPLIFICATION FACTORS (dimensionless, pA)
// ============================================================================

const (
	// COINCIDENCE_AMPLIFICATION_FACTOR_DEFAULT is the multiplicative factor
	// applied to the membrane potential or current when a coincidence is detected,
	// modeling regenerative dendritic events (e.g., NMDA spikes).
	COINCIDENCE_AMPLIFICATION_FACTOR_DEFAULT = 1.2 // Multiplier (similar to DENDRITE_CURRENT_SPIKE_AMPLITUDE_DEFAULT * factor)

	// COINCIDENCE_AMPLIFICATION_CURRENT_BOOST_DEFAULT is the direct current
	// (pA) added to the membrane potential when a coincidence leads to a
	// dendritic spike or similar regenerative event.
	COINCIDENCE_AMPLIFICATION_CURRENT_BOOST_DEFAULT = 1.0 // pA (directly added)

	// COINCIDENCE_AMPLIFICATION_CALCIUM_BOOST_DEFAULT is the additional calcium
	// influx (unitless, will be scaled) associated with coincidence detection,
	// crucial for plasticity mechanisms.
	COINCIDENCE_AMPLIFICATION_CALCIUM_BOOST_DEFAULT = 0.5 // Unitless (scaled by a calcium increment)
)

// ============================================================================
// MECHANISM SPECIFIC PARAMETERS
// ============================================================================

const (
	// COINCIDENCE_NMDA_MAGNIFICATION_FACTOR represents the specific non-linear
	// amplification provided by NMDA receptors when both ligand (glutamate)
	// and voltage thresholds are met.
	COINCIDENCE_NMDA_MAGNIFICATION_FACTOR = 1.5 // Multiplier

	// COINCIDENCE_BACKPROPAGATION_AFFECT_WINDOW specifies how long a
	// back-propagating action potential (bAP) can influence dendritic
	// coincidence detection.
	COINCIDENCE_BACKPROPAGATION_AFFECT_WINDOW = 5 * time.Millisecond // Time for bAP to influence
)

// ============================================================================
// TESTING AND DEBUGGING CONSTANTS
// ============================================================================

const (
	// COINCIDENCE_TEST_TOLERANCE for floating point comparisons in tests.
	COINCIDENCE_TEST_TOLERANCE = 0.0001

	// COINCIDENCE_DEBUG_LOGGING enables verbose logging for coincidence detection.
	COINCIDENCE_DEBUG_LOGGING = false
)
