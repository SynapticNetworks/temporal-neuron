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
func (n *Neuron) fireUnsafe() {
	now := time.Now()

	// Early return if in refractory period
	if !n.lastFireTime.IsZero() && now.Sub(n.lastFireTime) < n.refractoryPeriod {
		return
	}

	// === STEP 1: Capture all data we need under stateMutex ===
	// Store the current timestamp
	n.lastFireTime = now

	// Calculate output value before releasing lock
	outputValue := n.accumulator * n.fireFactor

	// Update calcium level
	n.homeostatic.calciumLevel += n.homeostatic.calciumIncrement

	// Prepare copies of data we'll need after releasing the lock
	matrixCallbacks := n.matrixCallbacks
	hasSTDPFeedback := n.stdpSystem.IsEnabled()

	// Copy ligand types to avoid holding lock during release
	ligandTypes := make([]types.LigandType, len(n.releasedLigands))
	copy(ligandTypes, n.releasedLigands)

	// Copy custom behaviors
	behaviors := n.customBehaviors

	// === STEP 2: Update firing history with activityMutex ===
	// Release stateMutex and acquire activityMutex to update firing history
	n.stateMutex.Unlock()

	n.activityMutex.Lock()
	n.homeostatic.firingHistory = append(n.homeostatic.firingHistory, now)
	// Trim history if needed
	if len(n.homeostatic.firingHistory) > 1000 {
		start := len(n.homeostatic.firingHistory) - 1000
		n.homeostatic.firingHistory = n.homeostatic.firingHistory[start:]
	}
	n.activityMutex.Unlock()

	// === STEP 3: External callbacks (without any locks) ===
	// Perform matrix callbacks without holding any locks
	if matrixCallbacks != nil {
		// Get connection count
		var connectionCount int
		n.outputsMutex.RLock()
		connectionCount = len(n.outputCallbacks)
		n.outputsMutex.RUnlock()

		// Get activity level safely
		activityLevel := n.GetActivityLevel()

		// DEADLOCK SAFETY: All external callbacks happen without holding any locks
		matrixCallbacks.ReportHealth(activityLevel, connectionCount)
		matrixCallbacks.SendElectricalSignal(types.SignalFired, outputValue)

		// Release neurotransmitters
		for _, ligandType := range ligandTypes {
			concentration := n.calculateReleaseConcentration(ligandType, outputValue)
			matrixCallbacks.ReleaseChemical(ligandType, concentration)
		}

		// Handle custom chemical release
		if behaviors != nil && behaviors.CustomChemicalRelease != nil {
			behaviors.CustomChemicalRelease(activityLevel, outputValue, matrixCallbacks.ReleaseChemical)
		}
	}

	// === STEP 4: Output transmission (requires locks) ===
	// Re-acquire stateMutex for the final steps
	n.stateMutex.Lock()

	// Handle output transmissions
	n.transmitToOutputSynapsesWithDelay(outputValue, now)

	// Schedule STDP feedback if enabled
	if hasSTDPFeedback {
		n.stdpSystem.ScheduleFeedback(now)
	}

	// Update neuron metadata
	n.UpdateMetadata("last_fire", now)
}

// ============================================================================
// AXONAL TRANSMISSION WITH REALISTIC DELAYS
// ============================================================================

// transmitToOutputSynapsesWithDelay sends signals to all connected synapses with realistic delays
func (n *Neuron) transmitToOutputSynapsesWithDelay(outputValue float64, fireTime time.Time) {
	// Take a snapshot of callbacks to minimize lock duration
	var callbacks map[string]types.OutputCallback

	// LOCK OPTIMIZATION: Minimize lock scope to just the copy operation
	n.outputsMutex.RLock()
	callbacks = make(map[string]types.OutputCallback, len(n.outputCallbacks))
	for id, callback := range n.outputCallbacks {
		callbacks[id] = callback
	}
	n.outputsMutex.RUnlock()

	// Safely capture neuron ID without a lock (ID is immutable)
	sourceID := n.ID()

	// Get primary neurotransmitter (avoid locks - this reads immutable data)
	ntType := n.getPrimaryNeurotransmitter()

	// Process each output callback without holding any locks
	for synapseID, callback := range callbacks {
		// Create the message
		msg := types.NeuralSignal{
			Value:                outputValue,
			Timestamp:            fireTime,
			SourceID:             sourceID,
			SynapseID:            synapseID,
			TargetID:             callback.GetTargetID(),
			NeurotransmitterType: ntType,
		}

		// Get delay for this connection
		delay := callback.GetDelay()
		if delay <= 0 {
			delay = AXON_DELAY_DEFAULT_TRANSMISSION
		}

		// Simple direct transmission
		callback.TransmitMessage(msg)
	}
}

// ============================================================================
// CHEMICAL RELEASE COORDINATION
// ============================================================================

// getPrimaryNeurotransmitter returns the main neurotransmitter type for this neuron
// This is safe to call without locks since releasedLigands is only modified during setup
func (n *Neuron) getPrimaryNeurotransmitter() types.LigandType {
	if len(n.releasedLigands) > 0 {
		return n.releasedLigands[0]
	}
	return types.LigandGlutamate // Default to glutamate
}

// calculateReleaseConcentration computes neurotransmitter release concentration
// This is a pure computation function that doesn't need locks
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
// FIRING STATUS AND DIAGNOSTICS
// ============================================================================

// GetFiringStatus returns current firing-related status information
func (n *Neuron) GetFiringStatus() map[string]interface{} {
	// Use separate mutex operations to minimize lock contention

	// Get firing history stats safely
	n.activityMutex.RLock()
	firingHistorySize := len(n.homeostatic.firingHistory)
	// Create a copy to work with after releasing the lock
	firingHistory := make([]time.Time, firingHistorySize)
	copy(firingHistory, n.homeostatic.firingHistory)
	n.activityMutex.RUnlock()

	// Get connection count safely
	n.outputsMutex.RLock()
	outputCount := len(n.outputCallbacks)
	n.outputsMutex.RUnlock()

	// Get state-related information
	n.stateMutex.Lock()
	lastFireTime := n.lastFireTime
	refractoryPeriod := n.refractoryPeriod
	calciumLevel := n.homeostatic.calciumLevel
	targetRate := n.homeostatic.targetFiringRate
	fireFactor := n.fireFactor
	pendingDeliveries := len(n.pendingDeliveries)
	n.stateMutex.Unlock()

	// Calculate if in refractory period
	inRefractory := !lastFireTime.IsZero() && time.Since(lastFireTime) < refractoryPeriod

	// Count recent spikes from the copy we made
	recentSpikeCount := 0
	cutoff := time.Now().Add(-5 * time.Second)
	for i := len(firingHistory) - 1; i >= 0; i-- {
		if firingHistory[i].After(cutoff) {
			recentSpikeCount++
		} else {
			break
		}
	}

	// Get current firing rate without locking
	currentRate := n.GetActivityLevel()

	// Construct the status response
	status := map[string]interface{}{
		"last_fire_time":      lastFireTime,
		"time_since_fire":     time.Since(lastFireTime),
		"refractory_period":   refractoryPeriod,
		"in_refractory":       inRefractory,
		"current_firing_rate": currentRate,
		"target_firing_rate":  targetRate,
		"calcium_level":       calciumLevel,
		"fire_factor":         fireFactor,
		"output_connections":  outputCount,
		"firing_history_size": firingHistorySize,
		"recent_spike_count":  recentSpikeCount,
	}

	// Get queue information
	n.stateMutex.Lock()
	queueLen := len(n.deliveryQueue)
	queueCap := cap(n.deliveryQueue)
	n.stateMutex.Unlock()

	// Add axonal delivery status
	status["axonal_delivery"] = map[string]interface{}{
		"pending_deliveries":  pendingDeliveries,
		"delivery_queue_size": queueLen,
		"delivery_capacity":   queueCap,
	}

	return status
}

// ============================================================================
// OUTPUT CONNECTION MANAGEMENT
// ============================================================================

// GetOutputConnectionInfo returns information about current output connections
func (n *Neuron) GetOutputConnectionInfo() map[string]interface{} {
	// OPTIMIZATION: Get ligands without a lock (immutable after setup)
	ligands := make([]string, 0, len(n.releasedLigands))
	for _, ligand := range n.releasedLigands {
		ligands = append(ligands, ligand.String())
	}

	// Get connection information with minimized lock time
	n.outputsMutex.RLock()
	connectionCount := len(n.outputCallbacks)

	// Build connection map efficiently
	connections := make(map[string]interface{}, connectionCount)
	for synapseID, callback := range n.outputCallbacks {
		connections[synapseID] = map[string]interface{}{
			"target_id": callback.GetTargetID(),
			"weight":    callback.GetWeight(),
			"delay":     callback.GetDelay(),
		}
	}
	n.outputsMutex.RUnlock()

	return map[string]interface{}{
		"connection_count": connectionCount,
		"connections":      connections,
		"neurotransmitter": n.getPrimaryNeurotransmitter().String(),
		"released_ligands": ligands,
	}
}
