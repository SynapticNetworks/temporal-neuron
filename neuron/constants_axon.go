package neuron

import "time"

/*
=================================================================================
AXONAL PROPAGATION CONSTANTS - BIOLOGICAL PARAMETER DEFINITIONS
=================================================================================

This file centralizes all biological constants related to axonal propagation
and signal delivery, ensuring consistency between implementation and tests,
and providing clear biological documentation for each parameter.

All constants follow the naming convention: AXON_[CATEGORY]_[PARAMETER]

Categories:
- DELAY: Temporal constants for signal propagation (time.Duration)
- CAPACITY: Limits for internal queuing mechanisms
- PERFORMANCE: Timing for internal processing loops

=================================================================================
*/

// ============================================================================
// AXONAL DELAY CONSTANTS (time.Duration)
// ============================================================================

const (
	// AXON_DELAY_MIN_TRANSMISSION models the theoretical minimum time for
	// action potential propagation and synaptic transmission. This can represent
	// very short, myelinated axons or direct electrical synapses (gap junctions).
	// Biological Range: ~0ms (gap junctions) to <1ms for highly myelinated axons.
	AXON_DELAY_MIN_TRANSMISSION = 0 * time.Millisecond // Represents immediate or very fast transmission

	// AXON_DELAY_DEFAULT_TRANSMISSION models a typical synaptic transmission delay,
	// encompassing axonal propagation, synaptic cleft diffusion, and postsynaptic
	// receptor activation. This is a common delay for standard chemical synapses.
	// Biological Range: ~1-5ms.
	AXON_DELAY_DEFAULT_TRANSMISSION = 1 * time.Millisecond // Default transmission delay

	// AXON_DELAY_LONG_TRANSMISSION models longer axonal propagation delays,
	// found in unmyelinated axons or over longer distances in the nervous system.
	// Biological Range: ~5-20ms.
	AXON_DELAY_LONG_TRANSMISSION = 5 * time.Millisecond // Example of a longer transmission delay

	// AXON_DELAY_MAX_BIOLOGICAL_PROPAGATION represents the upper bound for
	// biologically realistic signal propagation delays within a single neuron's axon.
	// Beyond this, signals would typically be considered too slow for direct causal
	// interactions in most neural circuits.
	// Biological Range: Up to ~20-50ms in some longer pathways.
	AXON_DELAY_MAX_BIOLOGICAL_PROPAGATION = 20 * time.Millisecond // Maximum expected biological propagation delay
)

// ============================================================================
// AXONAL CAPACITY CONSTANTS
// ============================================================================

const (
	// AXON_QUEUE_CAPACITY_DEFAULT defines the default buffer size for outgoing
	// messages awaiting axonal delivery. A larger capacity can handle bursts
	// but consumes more memory.
	// Biological Analogy: Limited resources for rapid vesicle fusion or action
	// potential generation frequency.
	AXON_QUEUE_CAPACITY_DEFAULT = 100 // Default size for the axonal delivery queue
)

// ============================================================================
// AXONAL PERFORMANCE / INTERNAL PROCESSING CONSTANTS
// ============================================================================

const (
	// AXON_TICK_INTERVAL defines the frequency at which the axonal delivery
	// worker checks for messages ready for transmission. A smaller interval
	// increases temporal precision but also computational overhead.
	// Biological Analogy: The discrete nature of action potential generation
	// and transmission, though much slower computationally than real biology.
	AXON_TICK_INTERVAL = 100 * time.Microsecond // Frequency of axonal delivery worker checks
)
