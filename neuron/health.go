package neuron

import (
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
)

// ============================================================================
// HEALTH MONITORING IMPLEMENTATION
// ============================================================================

// GetHealthMetrics implements the MonitorableComponent interface
// Provides comprehensive health and performance metrics for monitoring
func (n *Neuron) GetHealthMetrics() component.HealthMetrics {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// Calculate current activity level
	activityLevel := n.calculateCurrentFiringRateUnsafe()

	// Get connection count
	n.outputsMutex.RLock()
	connectionCount := len(n.outputCallbacks)
	n.outputsMutex.RUnlock()

	// Calculate processing load based on recent activity and system state
	processingLoad := n.calculateProcessingLoad()

	// Calculate overall health score
	healthScore := n.calculateHealthScore(activityLevel, processingLoad, connectionCount)

	// Identify any health issues
	issues := n.identifyHealthIssues(activityLevel, processingLoad, connectionCount)

	return component.HealthMetrics{
		ActivityLevel:   activityLevel,
		ConnectionCount: connectionCount,
		ProcessingLoad:  processingLoad,
		LastHealthCheck: time.Now(),
		HealthScore:     healthScore,
		Issues:          issues,
	}
}

// ============================================================================
// PROCESSING LOAD CALCULATION
// ============================================================================

// calculateProcessingLoad estimates the computational load on the neuron
// Updated to work with the new modular synaptic scaling system
func (n *Neuron) calculateProcessingLoad() float64 {
	// Base load from buffer utilization
	bufferLoad := float64(len(n.inputBuffer)) / float64(cap(n.inputBuffer))

	// Load from homeostatic processing
	homeostaticLoad := 0.0
	if time.Since(n.homeostatic.lastHomeostaticUpdate) < n.homeostatic.homeostaticInterval {
		homeostaticLoad = 0.1 // Active homeostatic processing
	}

	// Load from synaptic scaling (using new modular system)
	scalingLoad := 0.0
	if n.synapticScaling != nil && n.synapticScaling.Config.Enabled {
		// Check if scaling is actively running
		if time.Since(n.synapticScaling.Config.LastScalingUpdate) < n.synapticScaling.Config.ScalingInterval {
			scalingLoad = 0.05 // Active synaptic scaling
		}
	}

	// Load from dendritic integration processing
	dendriticLoad := 0.0
	if n.dendrite != nil {
		// Active dendritic modes create processing load
		switch n.dendrite.Name() {
		case "BiologicalTemporalSummation", "ActiveDendrite":
			dendriticLoad = 0.03 // Complex dendritic computation
		case "ShuntingInhibition":
			dendriticLoad = 0.02 // Moderate dendritic computation
		case "TemporalSummation":
			dendriticLoad = 0.01 // Simple temporal integration
		default:
			dendriticLoad = 0.0 // Passive dendrite
		}
	}

	// Load from axonal processing
	axonalLoad := 0.0
	if len(n.pendingDeliveries) > 0 {
		axonalLoad = float64(len(n.pendingDeliveries)) / 100.0 // Normalize to expected range
		if axonalLoad > 0.05 {
			axonalLoad = 0.05 // Cap axonal load contribution
		}
	}

	// Calculate total processing load (0.0 - 1.0)
	totalLoad := bufferLoad + homeostaticLoad + scalingLoad + dendriticLoad + axonalLoad
	if totalLoad > 1.0 {
		totalLoad = 1.0
	}

	return totalLoad
}

// ============================================================================
// HEALTH SCORE CALCULATION
// ============================================================================

// calculateHealthScore provides an overall health assessment (0.0 - 1.0)
// Enhanced to consider all neuron subsystems
func (n *Neuron) calculateHealthScore(activityLevel, processingLoad float64, connectionCount int) float64 {
	healthScore := 1.0

	// === ACTIVITY LEVEL ASSESSMENT ===
	targetActivity := n.homeostatic.targetFiringRate
	if targetActivity > 0 {
		activityRatio := activityLevel / targetActivity
		if activityRatio > 2.0 || activityRatio < 0.1 {
			healthScore -= 0.3 // Significant penalty for extreme activity
		} else if activityRatio > 1.5 || activityRatio < 0.5 {
			healthScore -= 0.1 // Minor penalty for moderate deviation
		}
	}

	// === PROCESSING LOAD ASSESSMENT ===
	if processingLoad > 0.8 {
		healthScore -= 0.2 // High load penalty
	} else if processingLoad > 0.6 {
		healthScore -= 0.1 // Moderate load penalty
	}

	// === CONNECTION HEALTH ASSESSMENT ===
	if connectionCount == 0 {
		healthScore -= 0.4 // Isolated neuron
	} else if connectionCount < 3 {
		healthScore -= 0.1 // Under-connected
	} else if connectionCount > 1000 {
		healthScore -= 0.15 // Over-connected (metabolic burden)
	}

	// === THRESHOLD STABILITY ASSESSMENT ===
	thresholdRatio := n.threshold / n.baseThreshold
	if thresholdRatio > 3.0 || thresholdRatio < 0.3 {
		healthScore -= 0.2 // Significant threshold drift
	} else if thresholdRatio > 2.0 || thresholdRatio < 0.5 {
		healthScore -= 0.1 // Moderate threshold drift
	}

	// === CALCIUM LEVEL ASSESSMENT ===
	if n.homeostatic.calciumLevel > 10.0 {
		healthScore -= 0.25 // Calcium toxicity risk
	} else if n.homeostatic.calciumLevel < -1.0 {
		healthScore -= 0.15 // Abnormally low calcium
	}

	// === SYNAPTIC SCALING HEALTH ===
	if n.synapticScaling != nil && n.synapticScaling.Config.Enabled {
		// Check for scaling oscillations
		history := n.synapticScaling.GetScalingHistory()
		if len(history) > 10 && n.detectScalingOscillations(history) {
			healthScore -= 0.1 // Scaling instability
		}

		// Check for extreme receptor gains
		gains := n.synapticScaling.GetInputGains()
		extremeGains := 0
		for _, gain := range gains {
			if gain < 0.1 || gain > 5.0 {
				extremeGains++
			}
		}
		if extremeGains > len(gains)/3 { // More than 1/3 of gains are extreme
			healthScore -= 0.15
		}
	}

	// === DENDRITIC INTEGRATION HEALTH ===
	if n.dendrite != nil {
		// Complex dendritic modes are more prone to issues
		switch n.dendrite.Name() {
		case "ActiveDendrite":
			// Active dendrites can have integration issues
			if processingLoad > 0.7 {
				healthScore -= 0.05 // Additional penalty for overloaded active dendrites
			}
		case "BiologicalTemporalSummation":
			// Biological modes can have noise-related issues
			if activityLevel < 0.1 {
				healthScore -= 0.05 // Biological dendrites need some activity
			}
		}
	}

	// === ENSURE BOUNDS ===
	if healthScore < 0.0 {
		healthScore = 0.0
	}

	return healthScore
}

// ============================================================================
// HEALTH ISSUE IDENTIFICATION
// ============================================================================

// identifyHealthIssues returns a list of specific health problems
// Enhanced to cover all neuron subsystems
func (n *Neuron) identifyHealthIssues(activityLevel, processingLoad float64, connectionCount int) []string {
	var issues []string

	// === ACTIVITY LEVEL ISSUES ===
	targetActivity := n.homeostatic.targetFiringRate
	if targetActivity > 0 {
		activityRatio := activityLevel / targetActivity
		if activityRatio > 3.0 {
			issues = append(issues, "hyperactive_firing")
		} else if activityRatio < 0.1 {
			issues = append(issues, "hypoactive_firing")
		}
	}

	// === PROCESSING LOAD ISSUES ===
	if processingLoad > 0.9 {
		issues = append(issues, "high_processing_load")
	}

	// === BUFFER ISSUES ===
	bufferUtilization := float64(len(n.inputBuffer)) / float64(cap(n.inputBuffer))
	if bufferUtilization > 0.8 {
		issues = append(issues, "input_buffer_congestion")
	}

	// === CONNECTION ISSUES ===
	if connectionCount == 0 {
		issues = append(issues, "isolated_neuron")
	} else if connectionCount > 1000 {
		issues = append(issues, "excessive_connections")
	}

	// === THRESHOLD ISSUES ===
	thresholdRatio := n.threshold / n.baseThreshold
	if thresholdRatio > 4.0 {
		issues = append(issues, "threshold_too_high")
	} else if thresholdRatio < 0.2 {
		issues = append(issues, "threshold_too_low")
	}

	// === CALCIUM LEVEL ISSUES ===
	if n.homeostatic.calciumLevel > 10.0 {
		issues = append(issues, "calcium_overload")
	} else if n.homeostatic.calciumLevel < 0.0 {
		issues = append(issues, "calcium_underflow")
	}

	// === TEMPORAL ISSUES ===
	if !n.lastFireTime.IsZero() && time.Since(n.lastFireTime) > 10*n.refractoryPeriod {
		issues = append(issues, "prolonged_silence")
	}

	// === SYNAPTIC SCALING ISSUES ===
	if n.synapticScaling != nil && n.synapticScaling.Config.Enabled {
		// Check for scaling problems
		gains := n.synapticScaling.GetInputGains()
		extremeGains := 0
		for _, gain := range gains {
			if gain < 0.05 {
				extremeGains++
			} else if gain > 10.0 {
				extremeGains++
			}
		}

		if extremeGains > 0 {
			issues = append(issues, "extreme_receptor_gains")
		}

		// Check for scaling oscillations
		history := n.synapticScaling.GetScalingHistory()
		if len(history) > 10 && n.detectScalingOscillations(history) {
			issues = append(issues, "scaling_oscillations")
		}

		// Check for scaling convergence failure
		status := n.synapticScaling.GetScalingStatus()
		if avgStrength, ok := status["current_avg_strength"].(float64); ok {
			targetStrength := n.synapticScaling.Config.TargetInputStrength
			if targetStrength > 0 {
				relativeError := (avgStrength - targetStrength) / targetStrength
				if relativeError > 0.5 || relativeError < -0.5 {
					issues = append(issues, "scaling_convergence_failure")
				}
			}
		}
	}

	// === DENDRITIC INTEGRATION ISSUES ===
	if n.dendrite != nil {
		// Check for dendritic integration problems
		switch n.dendrite.Name() {
		case "ActiveDendrite":
			// Active dendrites can have saturation issues
			if processingLoad > 0.8 && activityLevel > targetActivity*2 {
				issues = append(issues, "dendritic_saturation")
			}
		case "BiologicalTemporalSummation":
			// Biological modes need some activity to function properly
			if activityLevel < 0.05 {
				issues = append(issues, "dendritic_underutilization")
			}
		case "ShuntingInhibition":
			// Shunting modes can have inhibition balance issues
			if activityLevel < targetActivity*0.1 {
				issues = append(issues, "excessive_shunting_inhibition")
			}
		}
	}

	// === AXONAL DELIVERY ISSUES ===
	if len(n.pendingDeliveries) > 50 {
		issues = append(issues, "axonal_delivery_backlog")
	}

	// === CALLBACK SYSTEM ISSUES ===
	if n.matrixCallbacks == nil {
		issues = append(issues, "missing_matrix_callbacks")
	}

	return issues
}

// ============================================================================
// FIRING RATE CALCULATION
// ============================================================================

// calculateCurrentFiringRateUnsafe calculates the current firing rate
// This method must be called with stateMutex already locked
func (n *Neuron) calculateCurrentFiringRateUnsafe() float64 {
	now := time.Now()
	recentFires := 0

	// Count spikes within the activity window
	for i := len(n.homeostatic.firingHistory) - 1; i >= 0; i-- {
		if now.Sub(n.homeostatic.firingHistory[i]) <= n.homeostatic.activityWindow {
			recentFires++
		} else {
			break // History is ordered, so we can break here
		}
	}

	// Calculate rate in Hz
	return float64(recentFires) / n.homeostatic.activityWindow.Seconds()
}

// ============================================================================
// HELPER METHODS FOR HEALTH ASSESSMENT
// ============================================================================

// detectScalingOscillations detects if synaptic scaling is oscillating
func (n *Neuron) detectScalingOscillations(history []float64) bool {
	if len(history) < 6 {
		return false
	}

	// Look for alternating pattern in recent history
	recent := history[len(history)-6:]
	oscillations := 0

	for i := 1; i < len(recent)-1; i++ {
		// Check if this point is a local extremum
		if (recent[i] > recent[i-1] && recent[i] > recent[i+1]) ||
			(recent[i] < recent[i-1] && recent[i] < recent[i+1]) {
			oscillations++
		}
	}

	// If more than half the points are extrema, likely oscillating
	return oscillations > len(recent)/2
}

// resetAccumulatorUnsafe resets the membrane potential accumulator
// This method must be called with stateMutex already locked
func (n *Neuron) resetAccumulatorUnsafe() {
	n.accumulator = 0.0
}

// ============================================================================
// HEALTH MONITORING EXTENSIONS
// ============================================================================

// GetDetailedHealthReport returns comprehensive health information
func (n *Neuron) GetDetailedHealthReport() map[string]interface{} {
	metrics := n.GetHealthMetrics()

	n.stateMutex.Lock()
	threshold := n.threshold
	baseThreshold := n.baseThreshold
	accumulator := n.accumulator
	calciumLevel := n.homeostatic.calciumLevel
	n.stateMutex.Unlock()

	report := map[string]interface{}{
		"basic_metrics": map[string]interface{}{
			"health_score":     metrics.HealthScore,
			"activity_level":   metrics.ActivityLevel,
			"processing_load":  metrics.ProcessingLoad,
			"connection_count": metrics.ConnectionCount,
			"issues":           metrics.Issues,
		},
		"neural_state": map[string]interface{}{
			"threshold":       threshold,
			"base_threshold":  baseThreshold,
			"threshold_ratio": threshold / baseThreshold,
			"accumulator":     accumulator,
			"calcium_level":   calciumLevel,
		},
		"subsystem_health": make(map[string]interface{}),
	}

	// Add synaptic scaling health if available
	if n.synapticScaling != nil {
		scalingStatus := n.synapticScaling.GetScalingStatus()
		activitySummary := n.synapticScaling.GetInputActivitySummary()

		report["subsystem_health"].(map[string]interface{})["synaptic_scaling"] = map[string]interface{}{
			"status":   scalingStatus,
			"activity": activitySummary,
			"gains":    n.synapticScaling.GetInputGains(),
			"history":  n.synapticScaling.GetScalingHistory(),
		}
	}

	// Add dendritic integration health if available
	if n.dendrite != nil {
		report["subsystem_health"].(map[string]interface{})["dendritic_integration"] = map[string]interface{}{
			"mode": n.dendrite.Name(),
		}
	}

	return report
}
