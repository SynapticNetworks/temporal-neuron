package neuron

import (
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

/*
=================================================================================
FIRING MECHANISM - PURE FIRING LOGIC AND OUTPUT COORDINATION
=================================================================================

This file contains ONLY the firing mechanism and related functionality:
- fireUnsafe() - core firing logic
- Axonal transmission with delays
- Chemical release coordination
- Calcium and firing history management

NO processing logic or homeostatic adjustment - those are in processing.go
NO state management or interfaces - those are in neuron.go

This separation ensures clean responsibilities and eliminates duplication.

=================================================================================
*/

// ============================================================================
// CORE FIRING MECHANISM
// ============================================================================

// fireUnsafe handles the complete firing process including all subsystem coordination
// This method must be called with stateMutex already locked
// Add this debug version to firing.go temporarily
func (n *Neuron) fireUnsafe() {
	now := time.Now()

	// Check refractory period
	if !n.lastFireTime.IsZero() && now.Sub(n.lastFireTime) < n.refractoryPeriod {
		return
	}

	// Update firing state
	n.lastFireTime = now
	n.addCalciumFromFiringUnsafe()
	n.updateFiringHistoryUnsafe(now)

	// Calculate output value
	outputValue := n.accumulator * n.fireFactor

	// === MATRIX COORDINATION VIA CALLBACKS ===
	// Only use callbacks if they are available
	if n.matrixCallbacks != nil {
		// Report health and activity
		activityLevel := n.calculateCurrentFiringRateUnsafe()

		n.outputsMutex.RLock()
		connectionCount := len(n.outputCallbacks)
		n.outputsMutex.RUnlock()

		// Just call the interface methods directly - they handle their own error cases
		n.matrixCallbacks.ReportHealth(activityLevel, connectionCount)

		// Send electrical signal for gap junction coordination
		n.matrixCallbacks.SendElectricalSignal(types.SignalFired, outputValue)

		// Release chemicals into extracellular space
		n.releaseChemicalsViaCallback(outputValue)
	}

	// === SYNAPTIC TRANSMISSION VIA AXONAL DELIVERY ===
	n.transmitToOutputSynapsesWithDelay(outputValue, now)
}

// ============================================================================
// AXONAL TRANSMISSION WITH REALISTIC DELAYS
// ============================================================================

// transmitToOutputSynapsesWithDelay sends signals to all connected synapses with realistic delays
func (n *Neuron) transmitToOutputSynapsesWithDelay(outputValue float64, fireTime time.Time) {
	// Get snapshot of current callbacks to avoid holding lock during transmission
	n.outputsMutex.RLock()
	callbacks := make(map[string]types.OutputCallback, len(n.outputCallbacks))
	for id, callback := range n.outputCallbacks {
		callbacks[id] = callback
	}
	n.outputsMutex.RUnlock()

	// Transmit to all output synapses with axonal delays
	for synapseID, callback := range callbacks {
		msg := types.NeuralSignal{
			Value:                outputValue,
			Timestamp:            fireTime,
			SourceID:             n.ID(),
			SynapseID:            synapseID,
			TargetID:             callback.GetTargetID(),
			NeurotransmitterType: n.getPrimaryNeurotransmitter(),
		}

		// Get delay for this connection
		delay := callback.GetDelay()
		if delay <= 0 {
			delay = AXON_DELAY_DEFAULT_TRANSMISSION
		}

		// OPTION 1: Direct callback transmission (simplest)
		// Just use the callback directly without adapters
		callback.TransmitMessage(msg)
	}
}

// ============================================================================
// CHEMICAL RELEASE COORDINATION
// ============================================================================

// releaseChemicalsViaCallback releases neurotransmitters into extracellular space
func (n *Neuron) releaseChemicalsViaCallback(outputValue float64) {
	for _, ligandType := range n.releasedLigands {
		concentration := n.calculateReleaseConcentration(ligandType, outputValue)
		n.matrixCallbacks.ReleaseChemical(ligandType, concentration)
	}
	// Custom behavior callback
	if n.customBehaviors != nil && n.customBehaviors.CustomChemicalRelease != nil {
		activityRate := n.calculateCurrentFiringRateUnsafe()

		// Pass the release function directly to the callback
		n.customBehaviors.CustomChemicalRelease(activityRate, outputValue, n.matrixCallbacks.ReleaseChemical)
	}
}

// getPrimaryNeurotransmitter returns the main neurotransmitter type for this neuron
func (n *Neuron) getPrimaryNeurotransmitter() types.LigandType {
	if len(n.releasedLigands) > 0 {
		return n.releasedLigands[0]
	}
	return types.LigandGlutamate // Default to glutamate
}

// calculateReleaseConcentration computes neurotransmitter release concentration
func (n *Neuron) calculateReleaseConcentration(ligandType types.LigandType, outputValue float64) float64 {
	baseConcentration := outputValue * DENDRITE_CONCENTRATION_SCALE_BASE

	switch ligandType {
	case types.LigandGlutamate:
		return baseConcentration * DENDRITE_CONCENTRATION_FACTOR_GLUTAMATE
	case types.LigandGABA:
		return baseConcentration * DENDRITE_CONCENTRATION_FACTOR_GABA
	case types.LigandDopamine:
		return baseConcentration * DENDRITE_CONCENTRATION_FACTOR_DOPAMINE
	case types.LigandSerotonin:
		return baseConcentration * DENDRITE_CONCENTRATION_FACTOR_DEFAULT
	case types.LigandAcetylcholine:
		return baseConcentration * DENDRITE_CONCENTRATION_FACTOR_DEFAULT
	default:
		return baseConcentration * DENDRITE_CONCENTRATION_FACTOR_DEFAULT
	}
}

// ============================================================================
// CALCIUM AND FIRING HISTORY MANAGEMENT
// ============================================================================

// addCalciumFromFiringUnsafe adds calcium from action potential
// Must be called with stateMutex already locked
func (n *Neuron) addCalciumFromFiringUnsafe() {
	n.homeostatic.calciumLevel += n.homeostatic.calciumIncrement
}

// updateFiringHistoryUnsafe maintains the firing history buffer
// Must be called with stateMutex already locked
func (n *Neuron) updateFiringHistoryUnsafe(fireTime time.Time) {
	n.homeostatic.firingHistory = append(n.homeostatic.firingHistory, fireTime)

	// Trim old events outside the activity window for efficiency
	cutoff := fireTime.Add(-n.homeostatic.activityWindow)
	for i, t := range n.homeostatic.firingHistory {
		if t.After(cutoff) {
			n.homeostatic.firingHistory = n.homeostatic.firingHistory[i:]
			break
		}
	}

	// Limit total history size to prevent unbounded growth
	maxHistory := 1000
	if len(n.homeostatic.firingHistory) > maxHistory {
		start := len(n.homeostatic.firingHistory) - maxHistory
		n.homeostatic.firingHistory = n.homeostatic.firingHistory[start:]
	}
}

// ============================================================================
// FIRING STATUS AND DIAGNOSTICS
// ============================================================================

// GetFiringStatus returns current firing-related status information
func (n *Neuron) GetFiringStatus() map[string]interface{} {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	n.outputsMutex.RLock()
	outputCount := len(n.outputCallbacks)
	n.outputsMutex.RUnlock()

	status := map[string]interface{}{
		"last_fire_time":      n.lastFireTime,
		"time_since_fire":     time.Since(n.lastFireTime),
		"refractory_period":   n.refractoryPeriod,
		"in_refractory":       !n.lastFireTime.IsZero() && time.Since(n.lastFireTime) < n.refractoryPeriod,
		"current_firing_rate": n.calculateCurrentFiringRateUnsafe(),
		"target_firing_rate":  n.homeostatic.targetFiringRate,
		"calcium_level":       n.homeostatic.calciumLevel,
		"fire_factor":         n.fireFactor,
		"output_connections":  outputCount,
		"firing_history_size": len(n.homeostatic.firingHistory),
		"recent_spike_count":  n.countRecentSpikes(5 * time.Second),
	}

	// Add axonal delivery status
	status["axonal_delivery"] = map[string]interface{}{
		"pending_deliveries":  len(n.pendingDeliveries),
		"delivery_queue_size": len(n.deliveryQueue),
		"delivery_capacity":   cap(n.deliveryQueue),
	}

	return status
}

// countRecentSpikes counts spikes within a given time window
// Must be called with stateMutex already locked
func (n *Neuron) countRecentSpikes(window time.Duration) int {
	cutoff := time.Now().Add(-window)
	count := 0

	for i := len(n.homeostatic.firingHistory) - 1; i >= 0; i-- {
		if n.homeostatic.firingHistory[i].After(cutoff) {
			count++
		} else {
			break
		}
	}

	return count
}

// ============================================================================
// OUTPUT CONNECTION MANAGEMENT
// ============================================================================

// GetOutputConnectionInfo returns information about current output connections
func (n *Neuron) GetOutputConnectionInfo() map[string]interface{} {
	n.outputsMutex.RLock()
	defer n.outputsMutex.RUnlock()

	connections := make(map[string]interface{})
	for synapseID, callback := range n.outputCallbacks {
		connections[synapseID] = map[string]interface{}{
			"target_id": callback.GetTargetID(),
			"weight":    callback.GetWeight(),
			"delay":     callback.GetDelay(),
		}
	}

	return map[string]interface{}{
		"connection_count": len(n.outputCallbacks),
		"connections":      connections,
		"neurotransmitter": n.getPrimaryNeurotransmitter().String(),
		"released_ligands": func() []string {
			var ligands []string
			for _, ligand := range n.releasedLigands {
				ligands = append(ligands, ligand.String())
			}
			return ligands
		}(),
	}
}
