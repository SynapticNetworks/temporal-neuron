package neuron

import (
	"math"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/message"
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
			n.processDecayAndHomeostasis()

		case <-axonTicker.C:
			n.processAxonalDeliveries()

		case <-n.ctx.Done():
			return
		}
	}
}

// ============================================================================
// MESSAGE PROCESSING - INTEGRATED DENDRITIC AND SYNAPTIC SCALING
// ============================================================================

// processIncomingMessage handles incoming synaptic messages through the full processing pipeline:
// 1. Dendritic integration (handles temporal summation, active dendrites, etc.)
// 2. Synaptic scaling (post-synaptic receptor sensitivity)
// 3. Accumulator integration
// 4. Firing decision (delegated to firing.go)
func (n *Neuron) processIncomingMessage(msg message.NeuralSignal) {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// === STEP 1: DENDRITIC INTEGRATION ===
	var finalValue float64
	var dendriticResult *IntegratedPotential

	if n.dendrite != nil {
		// Process through dendritic integration system
		dendriticResult = n.dendrite.Handle(msg)

		if dendriticResult != nil {
			// Immediate dendritic processing result (e.g., PassiveMembraneMode)
			finalValue = dendriticResult.NetCurrent

			// Update metadata with dendritic computation details
			if dendriticResult.DendriticSpike {
				n.UpdateMetadata("last_dendritic_spike", time.Now())
			}
			if dendriticResult.NonlinearAmplification != 0 {
				n.UpdateMetadata("nonlinear_amplification", dendriticResult.NonlinearAmplification)
			}
		} else {
			// Message buffered for temporal integration - apply synaptic scaling to original
			finalValue = n.applySynapticScalingToMessage(msg)
		}
	} else {
		// Fallback: direct synaptic scaling without dendritic processing
		finalValue = n.applySynapticScalingToMessage(msg)
	}

	// === STEP 2: ACCUMULATOR INTEGRATION ===
	n.accumulator += finalValue

	// === STEP 3: FIRING DECISION (DELEGATED TO FIRING.GO) ===
	if n.accumulator >= n.threshold {
		n.fireUnsafe() // Implemented in firing.go
		n.resetAccumulatorUnsafe()
	}
}

// applySynapticScalingToMessage applies post-synaptic receptor scaling and records activity
func (n *Neuron) applySynapticScalingToMessage(msg message.NeuralSignal) float64 {
	var scaledValue float64

	// BEFORE:
	// if n.synapticScaling != nil && n.synapticScaling.Config.Enabled {

	// AFTER:
	var scalingEnabled bool
	if n.synapticScaling != nil {
		n.synapticScaling.mu.RLock()
		scalingEnabled = n.synapticScaling.Config.Enabled
		n.synapticScaling.mu.RUnlock()
	}

	if scalingEnabled {
		// Apply receptor sensitivity scaling
		scaledValue = n.synapticScaling.ApplyPostSynapticGain(msg)

		// Record activity for scaling decisions
		n.synapticScaling.RecordInputActivity(msg.SourceID, scaledValue)
	} else {
		// No scaling - use original value
		scaledValue = msg.Value
	}

	return scaledValue
}

// ============================================================================
// DECAY AND HOMEOSTASIS - INTEGRATED PROCESSING
// ============================================================================

// processDecayAndHomeostasis handles all slow processes:
// 1. Membrane potential decay
// 2. Calcium dynamics
// 3. Dendritic temporal processing
// 4. Synaptic scaling operations
// 5. Homeostatic threshold adjustment
func (n *Neuron) processDecayAndHomeostasis() {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// === STEP 1: BASIC MEMBRANE DECAY ===
	n.accumulator *= n.decayRate

	// === STEP 2: CALCIUM DYNAMICS ===
	n.homeostatic.calciumLevel *= n.homeostatic.calciumDecayRate

	// === STEP 3: DENDRITIC TEMPORAL PROCESSING ===
	if n.dendrite != nil {
		// Create membrane state snapshot for dendritic processing
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
	// BEFORE:
	// if n.synapticScaling != nil && n.synapticScaling.Config.Enabled {

	// AFTER:
	var scalingEnabled bool
	if n.synapticScaling != nil {
		n.synapticScaling.mu.RLock()
		scalingEnabled = n.synapticScaling.Config.Enabled
		n.synapticScaling.mu.RUnlock()
	}

	if scalingEnabled {
		currentRate := n.calculateCurrentFiringRateUnsafe()
		scalingResult := n.synapticScaling.PerformScaling(n.homeostatic.calciumLevel, currentRate)

		// Update metadata with scaling information
		if scalingResult.ScalingPerformed {
			n.UpdateMetadata("last_scaling_event", scalingResult.Timestamp)
			n.UpdateMetadata("scaling_factor", scalingResult.ScalingFactor)
			n.UpdateMetadata("avg_input_strength", scalingResult.AverageInputStrength)
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
	// Protect pendingDeliveries modification
	n.stateMutex.Lock()
	n.pendingDeliveries = ProcessAxonDeliveries(n.pendingDeliveries, n.deliveryQueue, now)
	n.stateMutex.Unlock()
}

// ============================================================================
// HOMEOSTATIC ADJUSTMENT LOGIC
// ============================================================================

// shouldPerformHomeostaticUpdateUnsafe checks if it's time for homeostatic adjustment
func (n *Neuron) shouldPerformHomeostaticUpdateUnsafe() bool {
	return time.Since(n.homeostatic.lastHomeostaticUpdate) >= n.homeostatic.homeostaticInterval
}

// performHomeostaticAdjustmentUnsafe adjusts firing threshold based on activity
func (n *Neuron) performHomeostaticAdjustmentUnsafe() {
	currentRate := n.calculateCurrentFiringRateUnsafe()
	targetRate := n.homeostatic.targetFiringRate

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
	if newThreshold != n.threshold {
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
// INTEGRATION STATUS AND MONITORING
// ============================================================================

// GetProcessingStatus returns comprehensive status of all processing subsystems
func (n *Neuron) GetProcessingStatus() map[string]interface{} {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	status := map[string]interface{}{
		"neural_state": map[string]interface{}{
			"accumulator":    n.accumulator,
			"threshold":      n.threshold,
			"last_fire_time": n.lastFireTime,
			"firing_rate":    n.calculateCurrentFiringRateUnsafe(),
		},
		"homeostatic": map[string]interface{}{
			"calcium_level":        n.homeostatic.calciumLevel,
			"target_firing_rate":   n.homeostatic.targetFiringRate,
			"last_update":          n.homeostatic.lastHomeostaticUpdate,
			"homeostasis_strength": n.homeostatic.homeostasisStrength,
		},
		"buffer_status": map[string]interface{}{
			"input_buffer_length":   len(n.inputBuffer),
			"input_buffer_capacity": cap(n.inputBuffer),
			"buffer_utilization":    float64(len(n.inputBuffer)) / float64(cap(n.inputBuffer)),
		},
		"axonal_delivery": map[string]interface{}{
			"pending_deliveries": len(n.pendingDeliveries),
			"delivery_queue_len": len(n.deliveryQueue),
		},
	}

	// Add synaptic scaling status if available
	if n.synapticScaling != nil {
		status["synaptic_scaling"] = n.synapticScaling.GetScalingStatus()
	}

	// Add dendritic integration status if available
	if n.dendrite != nil {
		status["dendritic_integration"] = map[string]interface{}{
			"mode": n.dendrite.Name(),
		}
	}

	// Add connection information
	n.outputsMutex.RLock()
	status["connections"] = map[string]interface{}{
		"output_count": len(n.outputCallbacks),
	}
	n.outputsMutex.RUnlock()

	return status
}

// GetSubsystemHealth returns health status of all integrated subsystems
func (n *Neuron) GetSubsystemHealth() map[string]interface{} {
	health := make(map[string]interface{})

	// Basic neural health
	n.stateMutex.Lock()
	firingRate := n.calculateCurrentFiringRateUnsafe()
	thresholdRatio := n.threshold / n.baseThreshold
	n.stateMutex.Unlock()

	health["neural_core"] = map[string]interface{}{
		"firing_rate_ok":   firingRate > 0.1 && firingRate < n.homeostatic.targetFiringRate*3,
		"threshold_stable": thresholdRatio > 0.5 && thresholdRatio < 2.0,
		"calcium_ok":       n.homeostatic.calciumLevel >= 0 && n.homeostatic.calciumLevel < 10.0,
		"buffer_ok":        float64(len(n.inputBuffer))/float64(cap(n.inputBuffer)) < 0.8,
	}

	// Synaptic scaling health
	// BEFORE:
	// if n.synapticScaling != nil && n.synapticScaling.Config.Enabled {

	// AFTER:
	var scalingEnabled bool
	if n.synapticScaling != nil {
		n.synapticScaling.mu.RLock()
		scalingEnabled = n.synapticScaling.Config.Enabled
		n.synapticScaling.mu.RUnlock()
	}

	if scalingEnabled {
		status := n.synapticScaling.GetScalingStatus()
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

	// Dendritic integration health
	if n.dendrite != nil {
		health["dendritic_integration"] = map[string]interface{}{
			"mode":        n.dendrite.Name(),
			"operational": true,
		}
	} else {
		health["dendritic_integration"] = map[string]interface{}{
			"mode":        "none",
			"operational": false,
		}
	}

	// Axonal delivery health
	deliveryBacklog := len(n.pendingDeliveries)
	threshold := int(float64(cap(n.deliveryQueue)) * 0.8)
	health["axonal_delivery"] = map[string]interface{}{
		"backlog_ok":      deliveryBacklog < 50,
		"pending_count":   deliveryBacklog,
		"deliveryQueueOk": len(n.deliveryQueue) < threshold,
	}

	// Connection health
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

// ============================================================================
// PERFORMANCE MONITORING
// ============================================================================

// GetPerformanceMetrics returns performance metrics for the integrated neuron
func (n *Neuron) GetPerformanceMetrics() map[string]interface{} {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// Calculate processing rates
	firingRate := n.calculateCurrentFiringRateUnsafe()

	// Buffer throughput
	bufferUtilization := float64(len(n.inputBuffer)) / float64(cap(n.inputBuffer))

	// Estimate message processing rate (messages per second)
	var messageRate float64
	if n.synapticScaling != nil {
		// Use activity tracking if available
		status := n.synapticScaling.GetInputActivitySummary()
		totalMessages := 0
		for _, sourceActivity := range status {
			if sourceData, ok := sourceActivity.(map[string]interface{}); ok {
				if count, ok := sourceData["recent_count"].(int); ok {
					totalMessages += count
				}
			}
		}
		// Estimate rate over the activity sampling window
		if samplingWindow := n.synapticScaling.Config.ActivitySamplingWindow; samplingWindow > 0 {
			messageRate = float64(totalMessages) / samplingWindow.Seconds()
		}
	}

	return map[string]interface{}{
		"firing_rate_hz":          firingRate,
		"message_processing_rate": messageRate,
		"buffer_utilization":      bufferUtilization,
		"axonal_backlog":          len(n.pendingDeliveries),
		"efficiency_score":        n.calculateEfficiencyScore(firingRate, bufferUtilization),
		"timestamp":               time.Now(),
	}
}

// calculateEfficiencyScore provides an overall efficiency metric (0.0 - 1.0)
func (n *Neuron) calculateEfficiencyScore(firingRate, bufferUtilization float64) float64 {
	efficiency := 1.0

	// Penalize extreme firing rates
	targetRate := n.homeostatic.targetFiringRate
	if targetRate > 0 {
		rateRatio := firingRate / targetRate
		if rateRatio > 2.0 || rateRatio < 0.5 {
			efficiency *= 0.8 // 20% penalty for rate deviation
		}
	}

	// Penalize high buffer utilization (processing bottleneck)
	if bufferUtilization > 0.8 {
		efficiency *= 0.7 // 30% penalty for high buffer utilization
	} else if bufferUtilization > 0.6 {
		efficiency *= 0.9 // 10% penalty for moderate buffer utilization
	}

	// Penalize large axonal backlog
	if len(n.pendingDeliveries) > 20 {
		efficiency *= 0.8 // 20% penalty for delivery backlog
	}

	return efficiency
}
