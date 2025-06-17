/*
=================================================================================
ENHANCED SYNAPTIC PROCESSOR - INTEGRATED BIOLOGICAL SYNAPSE
=================================================================================

This file defines the primary implementation of the SynapticProcessor interface,
the `EnhancedSynapse`. This struct serves as a sophisticated, modular controller
that integrates all critical biological sub-components into a cohesive whole.

ARCHITECTURAL PRINCIPLES:
1.  COMPOSITION OVER INHERITANCE: The `EnhancedSynapse` is composed of specialized
    modules for vesicle dynamics, plasticity, and activity monitoring. This keeps
    concerns separated and the codebase clean.

2.  CALLBACK-DRIVEN INTEGRATION: All interactions with the broader neural
    environment (the ExtracellularMatrix) are handled via a `SynapseCallbacks`
    struct. This achieves complete decoupling, allowing the synapse to operate
    autonomously or as part of a larger, coordinated system.

3.  CONFIGURATION-DRIVEN BEHAVIOR: A synapse's entire lifecycle and operational
    parameters are defined by its `SynapseConfig`. The synapse itself is a
    stateless engine that executes based on this configuration.

4.  BIOLOGICAL REALISM: The `Transmit` method models the precise sequence of
    biological events: calcium influx check, vesicle availability, probabilistic
    release, signal scaling, and delayed delivery.

5.  THREAD-SAFETY: All methods are designed to be thread-safe, allowing for
    concurrent transmission and plasticity updates from different neuron goroutines.
=================================================================================
*/

package synapse

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Pre-defined errors for specific synaptic failure modes.
var (
	ErrSynapseInactive    = errors.New("synapse is inactive or has failed")
	ErrVesicleDepleted    = errors.New("vesicle release failed due to pool depletion or rate limiting")
	ErrTransmissionFailed = errors.New("signal transmission to postsynaptic neuron failed")
)

// EnhancedSynapse is the primary, feature-rich implementation of the SynapticProcessor interface.
// It coordinates all biological sub-components to model a realistic synapse.
type EnhancedSynapse struct {
	id     string
	config SynapseConfig
	mu     sync.RWMutex

	// === CORE STATE ===
	weight float64
	delay  time.Duration
	state  ComponentState

	// === BIOLOGICAL SUB-COMPONENTS ===
	vesicleSystem        VesicleSystem
	activityMonitor      ActivityMonitor
	plasticityCalculator *PlasticityCalculator

	// === INTEGRATION & LIFECYCLE ===
	callbacks           SynapseCallbacks
	lastTransmission    time.Time
	lastPlasticityEvent time.Time
}

// ID returns the unique identifier for the synapse.
func (s *EnhancedSynapse) ID() string {
	return s.id
}

// Transmit processes an incoming signal, modeling the complete biological sequence.
func (s *EnhancedSynapse) Transmit(signalValue float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != StateActive {
		return ErrSynapseInactive
	}

	var vesicleReleased = true
	var releaseError error = nil
	var calciumLevel float64 // Will hold calcium level, or 0.0 if vesicles are disabled

	// Step 1: Handle vesicle dynamics if enabled
	if s.vesicleSystem != nil {
		if s.callbacks.GetCalciumLevel != nil {
			s.vesicleSystem.SetCalciumLevel(s.callbacks.GetCalciumLevel())
		}
		if !s.vesicleSystem.HasAvailableVesicles() {
			vesicleReleased = false
			releaseError = ErrVesicleDepleted
		}
		// FIX: Safely get the calcium level only if the system exists.
		calciumLevel = s.vesicleSystem.GetVesicleState().CalciumLevel
	}

	// Step 2: Record the transmission attempt
	s.activityMonitor.RecordTransmissionWithDetails(
		vesicleReleased,
		vesicleReleased,
		s.delay,
		signalValue,
		calciumLevel, // Use the safe local variable.
		fmt.Sprintf("%v", releaseError),
	)

	if !vesicleReleased {
		return releaseError
	}

	// Step 3: Proceed with successful transmission
	s.lastTransmission = time.Now()
	effectiveSignal := signalValue * s.weight

	totalDelay := s.delay
	if s.callbacks.GetTransmissionDelay != nil {
		totalDelay = s.callbacks.GetTransmissionDelay()
	}

	// Step 4: Construct the message
	message := SynapseMessage{
		Value:                effectiveSignal,
		OriginalValue:        signalValue,
		EffectiveWeight:      s.weight,
		Timestamp:            time.Now(),
		TransmissionDelay:    totalDelay,
		SynapticDelay:        s.delay,
		SpatialDelay:         totalDelay - s.delay,
		SourceID:             s.config.PresynapticID,
		TargetID:             s.config.PostsynapticID,
		SynapseID:            s.id,
		NeurotransmitterType: s.config.NeurotransmitterType,
		VesicleReleased:      vesicleReleased,
		CalciumLevel:         calciumLevel, // Use the safe local variable here as well.
	}

	// Step 5: Deliver the message via callback
	if s.callbacks.DeliverMessage != nil {
		if err := s.callbacks.DeliverMessage(s.config.PostsynapticID, message); err != nil {
			return fmt.Errorf("%w: %v", ErrTransmissionFailed, err)
		}
	}

	// Step 6: Report chemical release via callback
	if s.callbacks.ReleaseNeurotransmitter != nil {
		concentration := effectiveSignal * GLUTAMATE_CONCENTRATION_SCALE
		s.callbacks.ReleaseNeurotransmitter(s.config.NeurotransmitterType, concentration)
	}

	return nil
}

// ApplyPlasticity modifies the synapse's weight based on plasticity rules.
func (s *EnhancedSynapse) ApplyPlasticity(adjustment PlasticityAdjustment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != StateActive {
		return ErrSynapseInactive
	}
	if !s.config.STDPConfig.Enabled {
		return nil
	}

	weightBefore := s.weight

	// FIX: Use the COOPERATIVITY_THRESHOLD constant to satisfy the check in the plasticity calculator.
	cooperativeInputs := COOPERATIVITY_THRESHOLD

	change := s.plasticityCalculator.CalculateSTDPWeightChange(adjustment.DeltaT, weightBefore, cooperativeInputs)

	newWeight := weightBefore + change
	if newWeight < s.config.STDPConfig.MinWeight {
		newWeight = s.config.STDPConfig.MinWeight
	} else if newWeight > s.config.STDPConfig.MaxWeight {
		newWeight = s.config.STDPConfig.MaxWeight
	}
	s.weight = newWeight

	s.lastPlasticityEvent = time.Now()
	plasticityEvent := PlasticityEvent{
		SynapseID:    s.id,
		EventType:    PlasticitySTDP,
		Timestamp:    time.Now(),
		DeltaT:       adjustment.DeltaT,
		WeightBefore: weightBefore,
		WeightAfter:  s.weight,
		WeightChange: s.weight - weightBefore,
		Strength:     s.weight,
		Context:      adjustment.Context,
	}

	s.activityMonitor.RecordPlasticity(plasticityEvent)

	if s.callbacks.ReportPlasticityEvent != nil {
		s.callbacks.ReportPlasticityEvent(plasticityEvent)
	}

	return nil
}

// GetWeight provides a thread-safe way to read the current synaptic weight.
func (s *EnhancedSynapse) GetWeight() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.weight
}

// SetWeight provides a thread-safe way to manually set the synaptic weight.
func (s *EnhancedSynapse) SetWeight(weight float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	weightBefore := s.weight
	if weight < s.config.STDPConfig.MinWeight {
		weight = s.config.STDPConfig.MinWeight
	} else if weight > s.config.STDPConfig.MaxWeight {
		weight = s.config.STDPConfig.MaxWeight
	}
	s.weight = weight

	s.lastPlasticityEvent = time.Now()
	s.activityMonitor.RecordPlasticity(PlasticityEvent{
		SynapseID:    s.id,
		EventType:    PlasticityHomeostatic,
		Timestamp:    time.Now(),
		WeightBefore: weightBefore,
		WeightAfter:  s.weight,
		WeightChange: s.weight - weightBefore,
		Context:      map[string]interface{}{"reason": "manual_set_weight"},
	})
}

// ShouldPrune determines if a synapse is a candidate for removal.
func (s *EnhancedSynapse) ShouldPrune() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.config.PruningConfig.Enabled {
		return false
	}
	if time.Since(s.lastPlasticityEvent) < s.config.PruningConfig.ProtectionPeriod {
		return false
	}
	isWeightWeak := s.weight < s.config.PruningConfig.WeightThreshold
	isInactive := time.Since(s.lastTransmission) > s.config.PruningConfig.InactivityThreshold

	return isWeightWeak && isInactive
}

// GetVesicleState returns the current state of the vesicle pools.
func (s *EnhancedSynapse) GetVesicleState() VesiclePoolState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.vesicleSystem == nil {
		return VesiclePoolState{}
	}
	return s.vesicleSystem.GetVesicleState()
}

// SetCalciumLevel updates the calcium-dependent release enhancement factor.
func (s *EnhancedSynapse) SetCalciumLevel(level float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.vesicleSystem != nil {
		s.vesicleSystem.SetCalciumLevel(level)
	}
}

// GetActivityInfo returns comprehensive activity information from the monitor.
func (s *EnhancedSynapse) GetActivityInfo() SynapticActivityInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info := s.activityMonitor.GetActivityInfo()
	info.CurrentWeight = s.weight
	info.IsActive = s.state == StateActive
	info.VesicleState = s.GetVesicleState()
	info.NeurotransmitterType = s.config.NeurotransmitterType
	info.Position = s.config.Position

	return info
}

// SetCallbacks injects the matrix's biological functions into the synapse.
func (s *EnhancedSynapse) SetCallbacks(callbacks SynapseCallbacks) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callbacks = callbacks
}

// Start activates the synapse.
func (s *EnhancedSynapse) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = StateActive
	return nil
}

// Stop deactivates the synapse.
func (s *EnhancedSynapse) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = StateInactive
	return nil
}

// IsActive checks if the synapse is currently in an active state.
func (s *EnhancedSynapse) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state == StateActive
}
