package neuron

import (
	"math"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

/*
=================================================================================
INTEGRATED NEURAL PROCESSING PIPELINE
=================================================================================

This file contains the consolidated processing logic that integrates all
neuron subsystems in a clean, non-duplicated way:

- Main processing loop with proper timing
- Message processing through dendritic integration
- Synaptic scaling integration
- Homeostatic plasticity processing
- Axonal delivery coordination

All firing logic is delegated to firing.go to maintain separation of concerns.

=================================================================================
*/

// ============================================================================
// MAIN PROCESSING LOOP - INTEGRATED ARCHITECTURE
// ============================================================================

// Run is the main background processing loop that coordinates all neuron subsystems
func (n *Neuron) Run() {
	// Setup timing for different processing phases
	decayTicker := time.NewTicker(1 * time.Millisecond) // Fast membrane decay
	axonTicker := time.NewTicker(AXON_TICK_INTERVAL)    // Axonal delivery processing

	defer decayTicker.Stop()
	defer axonTicker.Stop()

	for {
		select {
		case msg := <-n.inputBuffer:
			n.processIncomingMessage(msg)

		case <-decayTicker.C:
			// Process regular decay and homeostasis
			n.processDecayAndHomeostasis()

			// Check STDP feedback separately
			n.processScheduledSTDPFeedback()

		case <-axonTicker.C:
			n.processAxonalDeliveries()

		case <-n.ctx.Done():
			return
		}
	}
}

func (n *Neuron) processScheduledSTDPFeedback() {

	// Get neuron ID and callbacks
	neuronID := n.ID()
	callbacks := n.matrixCallbacks

	// Skip if no callbacks available
	if callbacks == nil {
		return
	}

	// Check and deliver feedback if it's time
	feedbackDelivered := n.stdpSystem.CheckAndDeliverFeedback(neuronID, callbacks)

	// Update metadata if feedback was delivered
	if feedbackDelivered {
		n.UpdateMetadata("scheduled_stdp_feedback_delivered", time.Now())
	}
}

// ============================================================================
// MESSAGE PROCESSING - INTEGRATED DENDRITIC AND SYNAPTIC SCALING
// ============================================================================

// processIncomingMessage handles incoming synaptic messages through the full processing pipeline
func (n *Neuron) processIncomingMessage(msg types.NeuralSignal) {
	// Check if the neuron has a dendrite system that needs to be accessed
	hasDendrite := n.dendrite != nil

	// Get a copy of the synapse scaling system outside the lock
	var synapticScaling *SynapticScalingState
	n.stateMutex.Lock()
	synapticScaling = n.synapticScaling
	n.stateMutex.Unlock()

	// Start processing the message
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// === STEP 1: DENDRITIC INTEGRATION ===
	var finalValue float64

	if hasDendrite {
		// Process through dendritic integration system
		dendriticResult := n.dendrite.Handle(msg)

		if dendriticResult != nil {
			// Immediate dendritic processing result
			finalValue = dendriticResult.NetCurrent

			// Update metadata with dendritic computation details
			if dendriticResult.DendriticSpike {
				n.UpdateMetadata("last_dendritic_spike", time.Now())
			}
			if dendriticResult.NonlinearAmplification != 0 {
				n.UpdateMetadata("nonlinear_amplification", dendriticResult.NonlinearAmplification)
			}
		} else {
			// Message buffered for temporal integration
			finalValue = n.applySynapticScalingToMessageWithSystem(msg, synapticScaling)
		}
	} else {
		// Fallback: direct synaptic scaling
		finalValue = n.applySynapticScalingToMessageWithSystem(msg, synapticScaling)
	}

	// === STEP 2: ACCUMULATOR INTEGRATION ===
	n.accumulator += finalValue

	// === STEP 3: FIRING DECISION ===
	if n.accumulator >= n.threshold {
		n.fireUnsafe() // Implemented in firing.go
		n.resetAccumulatorUnsafe()
	}
}

// applySynapticScalingToMessageWithSystem applies scaling with an explicit system reference
// This helps avoid deadlocks by clearly separating lock domains
func (n *Neuron) applySynapticScalingToMessageWithSystem(msg types.NeuralSignal, synapticScaling *SynapticScalingState) float64 {
	// Default to the original value
	scaledValue := msg.Value

	// Skip if no scaling system
	if synapticScaling == nil {
		return scaledValue
	}

	// Check if scaling is enabled (using explicit RLock/RUnlock)
	synapticScaling.mu.RLock()
	scalingEnabled := synapticScaling.Config.Enabled
	synapticScaling.mu.RUnlock()

	if scalingEnabled {
		// Apply receptor sensitivity scaling
		scaledValue = synapticScaling.ApplyPostSynapticGain(msg)

		// Record activity for scaling decisions
		synapticScaling.RecordInputActivity(msg.SourceID, scaledValue)
	}

	return scaledValue
}

// ============================================================================
// DECAY AND HOMEOSTASIS - INTEGRATED PROCESSING
// ============================================================================

// processDecayAndHomeostasis handles all slow processes with better mutex management
func (n *Neuron) processDecayAndHomeostasis() {
	// Get references to subsystems outside the main lock
	hasDendrite := n.dendrite != nil

	var synapticScaling *SynapticScalingState
	n.stateMutex.Lock()
	synapticScaling = n.synapticScaling
	n.stateMutex.Unlock()

	// Now acquire the main lock for state updates
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// === STEP 1: BASIC MEMBRANE DECAY ===
	n.accumulator *= n.decayRate

	// === STEP 2: CALCIUM DYNAMICS ===
	n.homeostatic.calciumLevel *= n.homeostatic.calciumDecayRate

	// === STEP 3: DENDRITIC TEMPORAL PROCESSING ===
	if hasDendrite {
		// Create a snapshot of the current membrane state
		state := MembraneSnapshot{
			Accumulator:          n.accumulator,
			CurrentThreshold:     n.threshold,
			RestingPotential:     n.baseThreshold,
			IntracellularCalcium: n.homeostatic.calciumLevel,
			LastSpikeTime:        n.lastFireTime,
			RecentSpikeCount:     len(n.homeostatic.firingHistory),
			BackPropagatingSpike: !n.lastFireTime.IsZero() && time.Since(n.lastFireTime) < 5*time.Millisecond,
		}

		// Process any buffered dendritic inputs
		dendriticResult := n.dendrite.Process(state)
		if dendriticResult != nil {
			n.accumulator += dendriticResult.NetCurrent

			// Track dendritic computation metadata
			if dendriticResult.DendriticSpike {
				n.UpdateMetadata("last_dendritic_spike", time.Now())
			}

			// Update calcium from dendritic activity
			if dendriticResult.CalciumCurrent > 0 {
				n.homeostatic.calciumLevel += dendriticResult.CalciumCurrent * 0.1
			}
		}
	}

	// === STEP 4: SYNAPTIC SCALING OPERATIONS ===
	// Only process if we have a scaling system
	if synapticScaling != nil {
		// Check if scaling is enabled
		synapticScaling.mu.RLock()
		scalingEnabled := synapticScaling.Config.Enabled
		synapticScaling.mu.RUnlock()

		if scalingEnabled {
			// Get current firing rate (uses our unsafe version since we already hold stateMutex)
			currentRate := n.calculateCurrentFiringRateUnsafe()

			// Perform scaling
			scalingResult := synapticScaling.PerformScaling(n.homeostatic.calciumLevel, currentRate)

			// Update metadata with scaling information
			if scalingResult.ScalingPerformed {
				n.UpdateMetadata("last_scaling_event", scalingResult.Timestamp)
				n.UpdateMetadata("scaling_factor", scalingResult.ScalingFactor)
				n.UpdateMetadata("avg_input_strength", scalingResult.AverageInputStrength)
			}
		}
	}

	// === STEP 5: CHECK FIRING AFTER ALL PROCESSING ===
	if n.accumulator >= n.threshold {
		n.fireUnsafe() // Implemented in firing.go
		n.resetAccumulatorUnsafe()
	}

	// === STEP 6: HOMEOSTATIC THRESHOLD ADJUSTMENT ===
	if n.shouldPerformHomeostaticUpdateUnsafe() {
		n.performHomeostaticAdjustmentUnsafe()
	}
}

// ============================================================================
// AXONAL DELIVERY PROCESSING
// ============================================================================

// processAxonalDeliveries handles delayed message delivery through axons
func (n *Neuron) processAxonalDeliveries() {
	now := time.Now()

	// Minimize lock duration by copying what we need
	var pendingDeliveries []delayedMessage
	var deliveryQueue chan delayedMessage

	n.stateMutex.Lock()
	pendingDeliveries = n.pendingDeliveries
	deliveryQueue = n.deliveryQueue
	n.stateMutex.Unlock()

	// Process deliveries without holding the lock
	updatedDeliveries := ProcessAxonDeliveries(pendingDeliveries, deliveryQueue, now)

	// Update the state with the processed result
	n.stateMutex.Lock()
	n.pendingDeliveries = updatedDeliveries
	n.stateMutex.Unlock()
}

// ============================================================================
// HOMEOSTATIC ADJUSTMENT LOGIC
// ============================================================================

// shouldPerformHomeostaticUpdateUnsafe checks if it's time for homeostatic adjustment
// Must be called with stateMutex already locked
func (n *Neuron) shouldPerformHomeostaticUpdateUnsafe() bool {
	return time.Since(n.homeostatic.lastHomeostaticUpdate) >= n.homeostatic.homeostaticInterval
}

// performHomeostaticAdjustmentUnsafe adjusts firing threshold based on activity
// Must be called with stateMutex already locked
func (n *Neuron) performHomeostaticAdjustmentUnsafe() {
	// Calculate current and target firing rates
	currentRate := n.calculateCurrentFiringRateUnsafe()
	targetRate := n.homeostatic.targetFiringRate

	// Skip if target rate is zero (homeostasis disabled)
	if targetRate <= 0 {
		return
	}

	// Calculate threshold adjustment
	rateDifference := currentRate - targetRate
	adjustment := rateDifference * n.homeostatic.homeostasisStrength
	newThreshold := n.threshold + adjustment

	// Apply biological bounds
	if newThreshold < n.homeostatic.minThreshold {
		newThreshold = n.homeostatic.minThreshold
	} else if newThreshold > n.homeostatic.maxThreshold {
		newThreshold = n.homeostatic.maxThreshold
	}

	// Only update if change is significant
	if math.Abs(newThreshold-n.threshold) > 0.001 {
		oldThreshold := n.threshold
		n.threshold = newThreshold
		n.homeostatic.lastHomeostaticUpdate = time.Now()

		// Update metadata for monitoring
		n.UpdateMetadata("homeostatic_adjustment", map[string]interface{}{
			"old_threshold": oldThreshold,
			"new_threshold": newThreshold,
			"current_rate":  currentRate,
			"target_rate":   targetRate,
			"adjustment":    adjustment,
		})
	}
}

// ============================================================================
// STATUS AND MONITORING - OPTIMIZED FOR MINIMAL LOCK CONTENTION
// ============================================================================

// GetProcessingStatus returns comprehensive status with minimal lock contention
func (n *Neuron) GetProcessingStatus() map[string]interface{} {
	// Build the result incrementally with minimal lock durations
	status := make(map[string]interface{})

	// Get neural state with a single lock
	n.stateMutex.Lock()
	neuralState := map[string]interface{}{
		"accumulator":    n.accumulator,
		"threshold":      n.threshold,
		"last_fire_time": n.lastFireTime,
		"firing_rate":    n.calculateCurrentFiringRateUnsafe(),
	}

	// Get homeostatic data with the same lock
	homeostaticState := map[string]interface{}{
		"calcium_level":        n.homeostatic.calciumLevel,
		"target_firing_rate":   n.homeostatic.targetFiringRate,
		"last_update":          n.homeostatic.lastHomeostaticUpdate,
		"homeostasis_strength": n.homeostatic.homeostasisStrength,
	}

	// Get buffer and delivery data
	bufferStatus := map[string]interface{}{
		"input_buffer_length":   len(n.inputBuffer),
		"input_buffer_capacity": cap(n.inputBuffer),
		"buffer_utilization":    float64(len(n.inputBuffer)) / float64(cap(n.inputBuffer)),
	}

	axonalStatus := map[string]interface{}{
		"pending_deliveries": len(n.pendingDeliveries),
		"delivery_queue_len": len(n.deliveryQueue),
	}

	// Get references to subsystems
	synapticScaling := n.synapticScaling
	dendrite := n.dendrite
	n.stateMutex.Unlock()

	// Add the data to the result
	status["neural_state"] = neuralState
	status["homeostatic"] = homeostaticState
	status["buffer_status"] = bufferStatus
	status["axonal_delivery"] = axonalStatus
	status["stdp_system"] = n.stdpSystem.GetStatus()

	// Add synaptic scaling status if available (outside main lock)
	if synapticScaling != nil {
		status["synaptic_scaling"] = synapticScaling.GetScalingStatus()
	}

	// Add dendritic integration status if available
	if dendrite != nil {
		status["dendritic_integration"] = map[string]interface{}{
			"mode": dendrite.Name(),
		}
	}

	// Add connection information with separate lock
	n.outputsMutex.RLock()
	status["connections"] = map[string]interface{}{
		"output_count": len(n.outputCallbacks),
	}
	n.outputsMutex.RUnlock()

	return status
}

// GetSubsystemHealth returns health status with minimal lock contention
func (n *Neuron) GetSubsystemHealth() map[string]interface{} {
	health := make(map[string]interface{})

	// Get basic neural health metrics with minimal lock time
	var firingRate float64
	var thresholdRatio float64
	var calciumLevel float64
	var targetRate float64
	var synapticScaling *SynapticScalingState
	var dendrite DendriticIntegrationMode
	var pendingDeliveries int
	var deliveryQueueCap int

	// Get metrics from state mutex
	n.stateMutex.Lock()
	firingRate = n.calculateCurrentFiringRateUnsafe()
	thresholdRatio = n.threshold / n.baseThreshold
	calciumLevel = n.homeostatic.calciumLevel
	targetRate = n.homeostatic.targetFiringRate
	synapticScaling = n.synapticScaling
	dendrite = n.dendrite
	pendingDeliveries = len(n.pendingDeliveries)
	deliveryQueueCap = cap(n.deliveryQueue)
	n.stateMutex.Unlock()

	// Get buffer metrics
	bufferSize := len(n.inputBuffer)
	bufferCap := cap(n.inputBuffer)
	bufferUtilization := float64(bufferSize) / float64(bufferCap)

	// Calculate neural core health
	health["neural_core"] = map[string]interface{}{
		"firing_rate_ok":   firingRate > 0.1 && firingRate < targetRate*3,
		"threshold_stable": thresholdRatio > 0.5 && thresholdRatio < 2.0,
		"calcium_ok":       calciumLevel >= 0 && calciumLevel < 10.0,
		"buffer_ok":        bufferUtilization < 0.8,
	}

	// Synaptic scaling health
	if synapticScaling != nil {
		synapticScaling.mu.RLock()
		scalingEnabled := synapticScaling.Config.Enabled
		synapticScaling.mu.RUnlock()

		if scalingEnabled {
			status := synapticScaling.GetScalingStatus()
			avgStrength, ok1 := status["current_avg_strength"].(float64)
			targetStrength, ok2 := status["target_strength"].(float64)

			scalingHealthy := true
			if ok1 && ok2 && targetStrength > 0 {
				deviation := math.Abs(avgStrength-targetStrength) / targetStrength
				scalingHealthy = deviation < 0.5 // Within 50% of target
			}

			health["synaptic_scaling"] = map[string]interface{}{
				"enabled":         true,
				"healthy":         scalingHealthy,
				"avg_strength":    avgStrength,
				"target_strength": targetStrength,
			}
		} else {
			health["synaptic_scaling"] = map[string]interface{}{
				"enabled": false,
				"healthy": true, // Not enabled = not unhealthy
			}
		}
	}

	// Dendritic integration health
	if dendrite != nil {
		health["dendritic_integration"] = map[string]interface{}{
			"mode":        dendrite.Name(),
			"operational": true,
		}
	} else {
		health["dendritic_integration"] = map[string]interface{}{
			"mode":        "none",
			"operational": false,
		}
	}

	// Axonal delivery health
	deliveryQueueLen := len(n.deliveryQueue)
	threshold := int(float64(deliveryQueueCap) * 0.8)

	health["axonal_delivery"] = map[string]interface{}{
		"backlog_ok":      pendingDeliveries < 50,
		"pending_count":   pendingDeliveries,
		"deliveryQueueOk": deliveryQueueLen < threshold,
	}

	// Connection health with separate lock
	n.outputsMutex.RLock()
	connectionCount := len(n.outputCallbacks)
	n.outputsMutex.RUnlock()

	health["connectivity"] = map[string]interface{}{
		"connected":        connectionCount > 0,
		"connection_count": connectionCount,
		"well_connected":   connectionCount >= 3 && connectionCount <= 100,
	}

	return health
}

// GetPerformanceMetrics returns performance metrics with minimal lock contention
func (n *Neuron) GetPerformanceMetrics() map[string]interface{} {
	// Get metrics that require stateMutex with minimal lock duration
	var firingRate float64
	var pendingDeliveries int
	var targetRate float64
	var synapticScaling *SynapticScalingState

	n.stateMutex.Lock()
	firingRate = n.calculateCurrentFiringRateUnsafe()
	pendingDeliveries = len(n.pendingDeliveries)
	targetRate = n.homeostatic.targetFiringRate
	synapticScaling = n.synapticScaling
	n.stateMutex.Unlock()

	// Calculate buffer utilization
	bufferSize := len(n.inputBuffer)
	bufferCap := cap(n.inputBuffer)
	bufferUtilization := float64(bufferSize) / float64(bufferCap)

	// Estimate message processing rate (messages per second)
	var messageRate float64
	if synapticScaling != nil {
		// Use activity tracking if available
		status := synapticScaling.GetInputActivitySummary()
		totalMessages := 0

		for _, sourceActivity := range status {
			if sourceData, ok := sourceActivity.(map[string]interface{}); ok {
				if count, ok := sourceData["recent_count"].(int); ok {
					totalMessages += count
				}
			}
		}

		// Get sampling window
		synapticScaling.mu.RLock()
		samplingWindow := synapticScaling.Config.ActivitySamplingWindow
		synapticScaling.mu.RUnlock()

		// Estimate rate over the activity sampling window
		if samplingWindow > 0 {
			messageRate = float64(totalMessages) / samplingWindow.Seconds()
		}
	}

	// Calculate efficiency score
	efficiency := 1.0

	// Penalize extreme firing rates
	if targetRate > 0 {
		rateRatio := firingRate / targetRate
		if rateRatio > 2.0 || rateRatio < 0.5 {
			efficiency *= 0.8 // 20% penalty for rate deviation
		}
	}

	// Penalize high buffer utilization
	if bufferUtilization > 0.8 {
		efficiency *= 0.7 // 30% penalty for high buffer utilization
	} else if bufferUtilization > 0.6 {
		efficiency *= 0.9 // 10% penalty for moderate buffer utilization
	}

	// Penalize large axonal backlog
	if pendingDeliveries > 20 {
		efficiency *= 0.8 // 20% penalty for delivery backlog
	}

	return map[string]interface{}{
		"firing_rate_hz":          firingRate,
		"message_processing_rate": messageRate,
		"buffer_utilization":      bufferUtilization,
		"axonal_backlog":          pendingDeliveries,
		"efficiency_score":        efficiency,
		"timestamp":               time.Now(),
	}
}
