/*
=================================================================================
ENHANCED SYNAPSE INTERFACES - BIOLOGICAL INTEGRATION ARCHITECTURE
=================================================================================

This file defines the core interfaces and types for the enhanced synapse system
that integrates vesicle dynamics, matrix coordination, and biological realism
while maintaining complete decoupling through callback patterns.

DESIGN PRINCIPLES:
1. NO COMPILE-TIME DEPENDENCIES: All matrix/neuron integration via callbacks
2. BIOLOGICAL REALISM: Full vesicle dynamics and spatial coordination
3. HIGH PERFORMANCE: Optimized for 1M+ operations/second
4. CLEAN ARCHITECTURE: Clear separation of concerns

INTEGRATION STRATEGY:
- Vesicle dynamics embedded for neurotransmitter release constraints
- Matrix callbacks for spatial delays and chemical coordination
- Activity monitoring for network-wide plasticity coordination
- Flexible configuration supporting diverse synapse types
=================================================================================
*/

package synapse

import (
	"time"
)

// =================================================================================
// CORE SYNAPSE INTERFACE
// =================================================================================

// SynapticProcessor defines the complete contract for biologically realistic synapses
// with vesicle dynamics, spatial coordination, and matrix integration.
type SynapticProcessor interface {
	// === CORE IDENTIFICATION ===
	ID() string

	// === SIGNAL TRANSMISSION ===
	// Transmit processes signals with vesicle availability, calcium modulation,
	// spatial delays, and neurotransmitter release coordination
	Transmit(signalValue float64) error

	// === PLASTICITY SYSTEM ===
	ApplyPlasticity(adjustment PlasticityAdjustment) error
	GetWeight() float64
	SetWeight(weight float64)

	// === STRUCTURAL PLASTICITY ===
	ShouldPrune() bool

	// === VESICLE SYSTEM ===
	GetVesicleState() VesiclePoolState
	SetCalciumLevel(level float64)

	// === BIOLOGICAL COORDINATION ===
	SetCallbacks(callbacks SynapseCallbacks)
	GetActivityInfo() SynapticActivityInfo

	// === LIFECYCLE ===
	Start() error
	Stop() error
	IsActive() bool
}

// =================================================================================
// BIOLOGICAL CALLBACK SYSTEM
// =================================================================================

// SynapseCallbacks provides all biological functions injected by the matrix
// These callbacks wire the synapse into the broader neural environment
type SynapseCallbacks struct {
	// === VESICLE SYSTEM COORDINATION ===
	// Called to check calcium levels for release probability modulation
	GetCalciumLevel func() float64

	// === SPATIAL DELAY CALCULATION ===
	// Called to compute total transmission delay (synaptic + spatial)
	CalculateTransmissionDelay func() time.Duration

	// === SPATIAL DELAY CALCULATION ===
	// Called to compute total transmission delay (synaptic + spatial)
	GetTransmissionDelay func() time.Duration

	// === CHEMICAL SIGNALING ===
	// Called to release neurotransmitters into extracellular space
	ReleaseNeurotransmitter func(ligandType LigandType, concentration float64) error

	// === MESSAGE DELIVERY ===
	// Called to deliver processed signals to target neurons
	DeliverMessage func(targetID string, message SynapseMessage) error

	// === ACTIVITY MONITORING ===
	// Called to report synaptic activity for network coordination
	ReportActivity func(activity SynapticActivity)

	// === PLASTICITY COORDINATION ===
	// Called to report plasticity events for matrix-wide learning coordination
	ReportPlasticityEvent func(event PlasticityEvent)
}

// =================================================================================
// UTILITY INTERFACES
// =================================================================================

// VesicleSystem provides vesicle dynamics functionality
type VesicleSystem interface {
	HasAvailableVesicles() bool
	GetVesicleState() VesiclePoolState
	SetCalciumLevel(level float64)
	GetCurrentReleaseRate() float64
}

// ActivityMonitor tracks synaptic activity patterns
type ActivityMonitor interface {
	RecordTransmission(success bool, vesicleReleased bool, delay time.Duration)
	// RecordTransmissionWithDetails logs a transmission with additional biological context.
	RecordTransmissionWithDetails(
		success bool,
		vesicleReleased bool,
		processingTime time.Duration,
		signalStrength float64,
		calciumLevel float64,
		errorType string,
	)
	RecordPlasticity(event PlasticityEvent)
	GetActivityInfo() SynapticActivityInfo
	UpdateHealth()
}
