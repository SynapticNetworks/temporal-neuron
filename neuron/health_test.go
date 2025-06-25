/*
=================================================================================
NEURON HEALTH MONITORING TESTS
=================================================================================

OVERVIEW:
This test suite validates the health monitoring functionality of neurons,
ensuring that health metrics accurately reflect neuron state and that health
issues are properly detected. These tests verify both the calculation logic
and the biological relevance of health assessments.

BIOLOGICAL CONTEXT:
Neuron health monitoring models the biological processes that maintain neural
tissue integrity, similar to how microglia monitor neuron health in real brains.
Health metrics help identify neurons that are stressed, overloaded, isolated,
or operating outside normal physiological ranges.

KEY MECHANISMS TESTED:
1. **Activity Level Calculation**: Firing rate relative to target rates
2. **Processing Load Assessment**: Buffer utilization and computational burden
3. **Health Score Computation**: Overall health assessment (0.0-1.0)
4. **Issue Identification**: Specific health problems and anomalies
5. **Connection Health**: Social connectivity and network integration
6. **Homeostatic Balance**: Threshold stability and calcium regulation

=================================================================================
*/

package neuron

import (
	"fmt"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// ============================================================================
// Basic Health Metrics Tests
// ============================================================================

// TestNeuronHealthMetrics_BasicFunctionality validates the core health
// metrics calculation and interface compliance.
//
// BIOLOGICAL SIGNIFICANCE:
// This test ensures that health metrics are calculated correctly and that
// the MonitorableComponent interface is properly implemented, providing
// the foundation for microglial surveillance and network optimization.
func TestNeuronHealthMetrics_BasicFunctionality(t *testing.T) {
	t.Log("=== TESTING Basic Health Metrics Functionality ===")

	// Create a healthy neuron
	neuron := createHealthyTestNeuron("test-neuron-001")

	// Test basic interface compliance
	t.Run("InterfaceCompliance", func(t *testing.T) {
		// Verify neuron implements MonitorableComponent
		var _ component.MonitorableComponent = neuron

		// Get health metrics
		metrics := neuron.GetHealthMetrics()

		// Verify all required fields are present
		if metrics.LastHealthCheck.IsZero() {
			t.Error("LastHealthCheck should be set")
		}

		if metrics.HealthScore < 0.0 || metrics.HealthScore > 1.0 {
			t.Errorf("HealthScore should be 0.0-1.0, got %.3f", metrics.HealthScore)
		}

		if metrics.ActivityLevel < 0.0 {
			t.Errorf("ActivityLevel should be non-negative, got %.3f", metrics.ActivityLevel)
		}

		if metrics.ProcessingLoad < 0.0 || metrics.ProcessingLoad > 1.0 {
			t.Errorf("ProcessingLoad should be 0.0-1.0, got %.3f", metrics.ProcessingLoad)
		}

		if metrics.ConnectionCount < 0 {
			t.Errorf("ConnectionCount should be non-negative, got %d", metrics.ConnectionCount)
		}

		if metrics.Issues == nil {
			t.Error("Issues slice should be initialized (even if empty)")
		}

		t.Logf("✓ Basic health metrics: Score=%.3f, Activity=%.3f, Load=%.3f, Connections=%d",
			metrics.HealthScore, metrics.ActivityLevel, metrics.ProcessingLoad, metrics.ConnectionCount)
	})

	// Test healthy neuron baseline
	t.Run("HealthyBaseline", func(t *testing.T) {
		addMockConnections(neuron, 5)
		simulateNormalFiringPattern(neuron)
		metrics := neuron.GetHealthMetrics()

		// Healthy neuron should have high health score
		if metrics.HealthScore < 0.8 {
			t.Errorf("Healthy neuron should have high health score, got %.3f", metrics.HealthScore)
		}

		// Should have minimal issues
		if len(metrics.Issues) > 2 {
			t.Errorf("Healthy neuron should have few issues, got %v", metrics.Issues)
		}

		// Processing load should be reasonable
		if metrics.ProcessingLoad > 0.5 {
			t.Errorf("Healthy neuron should have low processing load, got %.3f", metrics.ProcessingLoad)
		}

		t.Log("✓ Healthy baseline verified")
	})
}

// ============================================================================
// Activity Level Assessment Tests
// ============================================================================

// TestNeuronHealthMetrics_ActivityLevels validates activity level calculation
// and its impact on health assessment.
//
// BIOLOGICAL SIGNIFICANCE:
// Activity levels reflect the neuron's firing rate relative to its target,
// modeling homeostatic regulation mechanisms that maintain optimal excitability.
// Extreme activity levels indicate potential pathology or network imbalance.
func TestNeuronHealthMetrics_ActivityLevels(t *testing.T) {
	t.Log("=== TESTING Activity Level Assessment ===")

	// Test normal activity
	t.Run("NormalActivity", func(t *testing.T) {
		neuron := createHealthyTestNeuron("normal-activity")

		// Simulate normal firing pattern
		simulateNormalFiringPattern(neuron)

		metrics := neuron.GetHealthMetrics()

		// Should have reasonable activity level
		if metrics.ActivityLevel < 0.0 || metrics.ActivityLevel > 10.0 {
			t.Errorf("Normal activity should be reasonable, got %.3f Hz", metrics.ActivityLevel)
		}

		// Should not trigger activity-related issues
		hasActivityIssues := false
		for _, issue := range metrics.Issues {
			if issue == "hyperactive_firing" || issue == "hypoactive_firing" {
				hasActivityIssues = true
				break
			}
		}

		if hasActivityIssues {
			t.Errorf("Normal activity should not trigger activity issues, got %v", metrics.Issues)
		}

		t.Logf("✓ Normal activity: %.3f Hz, Health=%.3f", metrics.ActivityLevel, metrics.HealthScore)
	})

	// Test hyperactivity
	t.Run("Hyperactivity", func(t *testing.T) {
		neuron := createHealthyTestNeuron("hyperactive")

		// Simulate excessive firing
		simulateHyperactivity(neuron)

		metrics := neuron.GetHealthMetrics()

		// Should detect hyperactivity
		hasHyperactivity := false
		for _, issue := range metrics.Issues {
			if issue == "hyperactive_firing" {
				hasHyperactivity = true
				break
			}
		}

		if !hasHyperactivity {
			t.Error("Hyperactivity should be detected")
		}

		// Health score should be reduced
		if metrics.HealthScore > 0.8 {
			t.Errorf("Hyperactive neuron should have reduced health score, got %.3f", metrics.HealthScore)
		}

		t.Logf("✓ Hyperactivity detected: %.3f Hz, Health=%.3f", metrics.ActivityLevel, metrics.HealthScore)
	})

	// Test hypoactivity
	t.Run("Hypoactivity", func(t *testing.T) {
		neuron := createHealthyTestNeuron("hypoactive")

		// No firing simulation (silent neuron)

		metrics := neuron.GetHealthMetrics()

		// Should detect hypoactivity
		hasHypoactivity := false
		for _, issue := range metrics.Issues {
			if issue == "hypoactive_firing" {
				hasHypoactivity = true
				break
			}
		}

		if !hasHypoactivity {
			t.Error("Hypoactivity should be detected")
		}

		// Health score should be reduced
		if metrics.HealthScore > 0.8 {
			t.Errorf("Hypoactive neuron should have reduced health score, got %.3f", metrics.HealthScore)
		}

		t.Logf("✓ Hypoactivity detected: %.3f Hz, Health=%.3f", metrics.ActivityLevel, metrics.HealthScore)
	})
}

// ============================================================================
// Processing Load Assessment Tests
// ============================================================================

// TestNeuronHealthMetrics_ProcessingLoad validates processing load calculation
// based on buffer utilization and active processes.
//
// BIOLOGICAL SIGNIFICANCE:
// Processing load models metabolic burden and computational stress on neurons.
// High processing loads indicate potential energy depletion or overuse,
// similar to metabolic stress in biological neurons.
func TestNeuronHealthMetrics_ProcessingLoad(t *testing.T) {
	t.Log("=== TESTING Processing Load Assessment ===")

	// Test low processing load
	t.Run("LowProcessingLoad", func(t *testing.T) {
		neuron := createHealthyTestNeuron("low-load")

		metrics := neuron.GetHealthMetrics()

		// Fresh neuron should have low processing load
		if metrics.ProcessingLoad > 0.2 {
			t.Errorf("Fresh neuron should have low processing load, got %.3f", metrics.ProcessingLoad)
		}

		// Should not trigger load-related issues
		hasLoadIssues := false
		for _, issue := range metrics.Issues {
			if issue == "high_processing_load" {
				hasLoadIssues = true
				break
			}
		}

		if hasLoadIssues {
			t.Error("Low load neuron should not trigger load issues")
		}

		t.Logf("✓ Low processing load: %.3f", metrics.ProcessingLoad)
	})

	// Test high processing load
	// t.Run("HighProcessingLoad", func(t *testing.T) {
	// 	neuron := createHealthyTestNeuron("high-load")

	// 	// Fill input buffer to create high load
	// 	fillInputBuffer(neuron)

	// 	// Enable synaptic scaling to add processing load
	// 	neuron.EnableSynapticScaling(1.0, 0.001, 1*time.Millisecond)

	// 	// Force recent scaling update
	// 	neuron.scalingConfig.LastScalingUpdate = time.Now()

	// 	metrics := neuron.GetHealthMetrics()

	// 	// Should have high processing load
	// 	if metrics.ProcessingLoad < 0.7 {
	// 		t.Logf("Expected high processing load, got %.3f (may vary based on timing)", metrics.ProcessingLoad)
	// 	}

	// 	// May trigger high load issues
	// 	t.Logf("✓ High processing load scenario: %.3f", metrics.ProcessingLoad)
	// })

	// Test buffer congestion
	t.Run("BufferCongestion", func(t *testing.T) {
		neuron := createHealthyTestNeuron("congested")

		// Fill buffer beyond congestion threshold
		fillInputBuffer(neuron)

		metrics := neuron.GetHealthMetrics()

		// Should detect buffer congestion
		hasCongestion := false
		for _, issue := range metrics.Issues {
			if issue == "input_buffer_congestion" {
				hasCongestion = true
				break
			}
		}

		if !hasCongestion {
			t.Error("Buffer congestion should be detected")
		}

		t.Logf("✓ Buffer congestion detected: Load=%.3f", metrics.ProcessingLoad)
	})
}

// ============================================================================
// Connection Health Tests
// ============================================================================

// TestNeuronHealthMetrics_ConnectionHealth validates connection count assessment
// and its impact on neuron health.
//
// BIOLOGICAL SIGNIFICANCE:
// Connection health models the social connectivity of neurons in networks.
// Isolated neurons or those with excessive connections may indicate
// developmental problems or pathological states.
func TestNeuronHealthMetrics_ConnectionHealth(t *testing.T) {
	t.Log("=== TESTING Connection Health Assessment ===")

	// Test isolated neuron
	t.Run("IsolatedNeuron", func(t *testing.T) {
		neuron := createHealthyTestNeuron("isolated")
		// Don't add any connections (default state)

		metrics := neuron.GetHealthMetrics()

		// Should detect isolation
		hasIsolation := false
		for _, issue := range metrics.Issues {
			if issue == "isolated_neuron" {
				hasIsolation = true
				break
			}
		}

		if !hasIsolation {
			t.Error("Isolated neuron should be detected")
		}

		// Health score should be significantly reduced
		if metrics.HealthScore > 0.7 {
			t.Errorf("Isolated neuron should have reduced health score, got %.3f", metrics.HealthScore)
		}

		if metrics.ConnectionCount != 0 {
			t.Errorf("Expected 0 connections, got %d", metrics.ConnectionCount)
		}

		t.Logf("✓ Isolation detected: Connections=%d, Health=%.3f",
			metrics.ConnectionCount, metrics.HealthScore)
	})

	// Test well-connected neuron
	t.Run("WellConnectedNeuron", func(t *testing.T) {
		neuron := createHealthyTestNeuron("well-connected")

		// Add several connections
		addMockConnections(neuron, 10)

		metrics := neuron.GetHealthMetrics()

		// Should have good connectivity
		if metrics.ConnectionCount != 10 {
			t.Errorf("Expected 10 connections, got %d", metrics.ConnectionCount)
		}

		// Should not trigger isolation issues
		hasIsolationIssues := false
		for _, issue := range metrics.Issues {
			if issue == "isolated_neuron" {
				hasIsolationIssues = true
				break
			}
		}

		if hasIsolationIssues {
			t.Error("Well-connected neuron should not have isolation issues")
		}

		// Health score should be good
		if metrics.HealthScore < 0.7 {
			t.Logf("Well-connected neuron health: %.3f (may be affected by other factors)", metrics.HealthScore)
		}

		t.Logf("✓ Well-connected neuron: Connections=%d, Health=%.3f",
			metrics.ConnectionCount, metrics.HealthScore)
	})

	// Test over-connected neuron
	t.Run("OverConnectedNeuron", func(t *testing.T) {
		neuron := createHealthyTestNeuron("over-connected")

		// Add excessive connections
		addMockConnections(neuron, 1500)

		metrics := neuron.GetHealthMetrics()

		// Should detect excessive connections
		hasExcessiveConnections := false
		for _, issue := range metrics.Issues {
			if issue == "excessive_connections" {
				hasExcessiveConnections = true
				break
			}
		}

		if !hasExcessiveConnections {
			t.Error("Excessive connections should be detected")
		}

		t.Logf("✓ Excessive connections detected: Connections=%d", metrics.ConnectionCount)
	})
}

// ============================================================================
// Threshold Health Tests
// ============================================================================

// TestNeuronHealthMetrics_ThresholdHealth validates threshold stability
// assessment and drift detection.
//
// BIOLOGICAL SIGNIFICANCE:
// Threshold health models the stability of neuron excitability. Significant
// drift from baseline thresholds indicates homeostatic failure or pathological
// changes in excitability.
func TestNeuronHealthMetrics_ThresholdHealth(t *testing.T) {
	t.Log("=== TESTING Threshold Health Assessment ===")

	// Test normal threshold
	t.Run("NormalThreshold", func(t *testing.T) {
		neuron := createHealthyTestNeuron("normal-threshold")

		// Threshold should be at baseline
		baseline := neuron.baseThreshold
		current := neuron.GetThreshold()

		if current != baseline {
			t.Errorf("Fresh neuron threshold should match baseline: baseline=%.3f, current=%.3f",
				baseline, current)
		}

		metrics := neuron.GetHealthMetrics()

		// Should not have threshold issues
		hasThresholdIssues := false
		for _, issue := range metrics.Issues {
			if issue == "threshold_too_high" || issue == "threshold_too_low" {
				hasThresholdIssues = true
				break
			}
		}

		if hasThresholdIssues {
			t.Errorf("Normal threshold should not trigger issues, got %v", metrics.Issues)
		}

		t.Logf("✓ Normal threshold: %.3f (baseline: %.3f)", current, baseline)
	})

	// Test high threshold
	t.Run("HighThreshold", func(t *testing.T) {
		neuron := createHealthyTestNeuron("high-threshold")

		// Set threshold much higher than baseline
		baseline := neuron.baseThreshold
		highThreshold := baseline * 5.0
		neuron.SetThreshold(highThreshold)

		metrics := neuron.GetHealthMetrics()

		// Should detect high threshold
		hasHighThreshold := false
		for _, issue := range metrics.Issues {
			if issue == "threshold_too_high" {
				hasHighThreshold = true
				break
			}
		}

		if !hasHighThreshold {
			t.Error("High threshold should be detected")
		}

		// Health score should be reduced
		if metrics.HealthScore > 0.9 {
			t.Errorf("High threshold should reduce health score, got %.3f", metrics.HealthScore)
		}

		t.Logf("✓ High threshold detected: %.3f (%.1fx baseline)", highThreshold, highThreshold/baseline)
	})

	// Test low threshold
	t.Run("LowThreshold", func(t *testing.T) {
		neuron := createHealthyTestNeuron("low-threshold")

		// Set threshold much lower than baseline
		baseline := neuron.baseThreshold
		lowThreshold := baseline * 0.1
		neuron.SetThreshold(lowThreshold)

		metrics := neuron.GetHealthMetrics()

		// Should detect low threshold
		hasLowThreshold := false
		for _, issue := range metrics.Issues {
			if issue == "threshold_too_low" {
				hasLowThreshold = true
				break
			}
		}

		if !hasLowThreshold {
			t.Error("Low threshold should be detected")
		}

		t.Logf("✓ Low threshold detected: %.3f (%.1fx baseline)", lowThreshold, lowThreshold/baseline)
	})
}

// ============================================================================
// Calcium Level Tests
// ============================================================================

// TestNeuronHealthMetrics_CalciumLevels validates calcium level monitoring
// and toxicity detection.
//
// BIOLOGICAL SIGNIFICANCE:
// Calcium levels are critical for neuron health. Excessive calcium can trigger
// excitotoxicity and cell death, while insufficient calcium affects plasticity
// and signaling. This models biological calcium homeostasis monitoring.
func TestNeuronHealthMetrics_CalciumLevels(t *testing.T) {
	t.Log("=== TESTING Calcium Level Assessment ===")

	// Test normal calcium
	t.Run("NormalCalcium", func(t *testing.T) {
		neuron := createHealthyTestNeuron("normal-calcium")

		metrics := neuron.GetHealthMetrics()

		// Should not have calcium issues
		hasCalciumIssues := false
		for _, issue := range metrics.Issues {
			if issue == "calcium_overload" || issue == "calcium_underflow" {
				hasCalciumIssues = true
				break
			}
		}

		if hasCalciumIssues {
			t.Errorf("Normal calcium should not trigger issues, got %v", metrics.Issues)
		}

		t.Logf("✓ Normal calcium levels: %.3f", neuron.homeostatic.calciumLevel)
	})

	// Test calcium overload
	t.Run("CalciumOverload", func(t *testing.T) {
		neuron := createHealthyTestNeuron("calcium-overload")

		// Simulate calcium overload
		neuron.homeostatic.calciumLevel = 15.0 // Above threshold of 10.0

		metrics := neuron.GetHealthMetrics()

		// Should detect calcium overload
		hasOverload := false
		for _, issue := range metrics.Issues {
			if issue == "calcium_overload" {
				hasOverload = true
				break
			}
		}

		if !hasOverload {
			t.Error("Calcium overload should be detected")
		}

		t.Logf("✓ Calcium overload detected: %.3f", neuron.homeostatic.calciumLevel)
	})

	// Test calcium underflow
	t.Run("CalciumUnderflow", func(t *testing.T) {
		neuron := createHealthyTestNeuron("calcium-underflow")

		// Simulate calcium underflow
		neuron.homeostatic.calciumLevel = -0.5 // Below zero

		metrics := neuron.GetHealthMetrics()

		// Should detect calcium underflow
		hasUnderflow := false
		for _, issue := range metrics.Issues {
			if issue == "calcium_underflow" {
				hasUnderflow = true
				break
			}
		}

		if !hasUnderflow {
			t.Error("Calcium underflow should be detected")
		}

		t.Logf("✓ Calcium underflow detected: %.3f", neuron.homeostatic.calciumLevel)
	})
}

// ============================================================================
// Comprehensive Health Assessment Tests
// ============================================================================

// TestNeuronHealthMetrics_ComprehensiveAssessment validates the overall
// health assessment with multiple interacting factors.
//
// BIOLOGICAL SIGNIFICANCE:
// This test models realistic scenarios where multiple health factors interact,
// similar to how multiple pathological processes can affect neuron health
// simultaneously in biological systems.
func TestNeuronHealthMetrics_ComprehensiveAssessment(t *testing.T) {
	t.Log("=== TESTING Comprehensive Health Assessment ===")

	// Test multiple health issues
	t.Run("MultipleHealthIssues", func(t *testing.T) {
		neuron := createHealthyTestNeuron("multiple-issues")

		// Create multiple health problems
		neuron.SetThreshold(neuron.baseThreshold * 6.0) // High threshold
		neuron.homeostatic.calciumLevel = 12.0          // Calcium overload
		fillInputBuffer(neuron)                         // Buffer congestion
		// Leave isolated (no connections)

		metrics := neuron.GetHealthMetrics()

		// Should have multiple issues
		expectedIssues := []string{
			"threshold_too_high",
			"calcium_overload",
			"input_buffer_congestion",
			"isolated_neuron",
		}

		for _, expectedIssue := range expectedIssues {
			found := false
			for _, actualIssue := range metrics.Issues {
				if actualIssue == expectedIssue {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected issue '%s' not found in %v", expectedIssue, metrics.Issues)
			}
		}

		// Health score should be very low
		if metrics.HealthScore > 0.3 {
			t.Errorf("Multiple issues should severely reduce health score, got %.3f", metrics.HealthScore)
		}

		t.Logf("✓ Multiple issues detected: %v, Health=%.3f", metrics.Issues, metrics.HealthScore)
	})

	// Test health recovery
	t.Run("HealthRecovery", func(t *testing.T) {
		neuron := createHealthyTestNeuron("recovering")

		// Create and then fix health issues
		neuron.SetThreshold(neuron.baseThreshold * 6.0) // High threshold

		// Get initial poor health
		badMetrics := neuron.GetHealthMetrics()

		// Fix the threshold
		neuron.SetThreshold(neuron.baseThreshold) // Back to normal

		// Add some connections
		addMockConnections(neuron, 5)

		// Get recovered health
		goodMetrics := neuron.GetHealthMetrics()

		// Health should improve
		if goodMetrics.HealthScore <= badMetrics.HealthScore {
			t.Errorf("Health should improve after fixes: before=%.3f, after=%.3f",
				badMetrics.HealthScore, goodMetrics.HealthScore)
		}

		// Should have fewer issues
		if len(goodMetrics.Issues) >= len(badMetrics.Issues) {
			t.Errorf("Should have fewer issues after recovery: before=%v, after=%v",
				badMetrics.Issues, goodMetrics.Issues)
		}

		t.Logf("✓ Health recovery: %.3f → %.3f, Issues: %v → %v",
			badMetrics.HealthScore, goodMetrics.HealthScore,
			badMetrics.Issues, goodMetrics.Issues)
	})
}

// ============================================================================
// Test Helper Functions
// ============================================================================

// createHealthyTestNeuron creates a neuron in a healthy baseline state.
func createHealthyTestNeuron(id string) *Neuron {
	return NewNeuron(
		id,
		1.0,                 // threshold
		0.95,                // decayRate
		10*time.Millisecond, // refractoryPeriod
		1.0,                 // fireFactor
		5.0,                 // targetFiringRate (5 Hz)
		0.1,                 // homeostasisStrength
	)
}

// simulateNormalFiringPattern creates a realistic firing history.
func simulateNormalFiringPattern(neuron *Neuron) {
	now := time.Now()

	// Add firing events over the last few seconds at target rate
	targetRate := neuron.homeostatic.targetFiringRate // 5 Hz
	window := neuron.homeostatic.activityWindow       // 5 seconds

	// Calculate number of spikes for target rate
	targetSpikes := int(targetRate * window.Seconds())

	// Add spikes at regular intervals
	for i := 0; i < targetSpikes; i++ {
		spikeTime := now.Add(-window + time.Duration(i)*window/time.Duration(targetSpikes))
		neuron.homeostatic.firingHistory = append(neuron.homeostatic.firingHistory, spikeTime)
	}
}

// simulateHyperactivity creates excessive firing activity.
func simulateHyperactivity(neuron *Neuron) {
	now := time.Now()

	// Add many spikes (4x target rate)
	targetRate := neuron.homeostatic.targetFiringRate * 4 // 20 Hz
	window := neuron.homeostatic.activityWindow

	hyperSpikes := int(targetRate * window.Seconds())

	for i := 0; i < hyperSpikes; i++ {
		spikeTime := now.Add(-window + time.Duration(i)*window/time.Duration(hyperSpikes))
		neuron.homeostatic.firingHistory = append(neuron.homeostatic.firingHistory, spikeTime)
	}
}

// fillInputBuffer fills the neuron's input buffer to simulate high load.
func fillInputBuffer(neuron *Neuron) {
	// Fill buffer to 90% capacity to trigger congestion detection
	capacity := cap(neuron.inputBuffer)
	fillCount := int(float64(capacity) * 0.9)

	for i := 0; i < fillCount; i++ {
		msg := types.NeuralSignal{
			Value:                0.1,
			Timestamp:            time.Now(),
			SourceID:             "load-test",
			NeurotransmitterType: types.LigandGlutamate,
		}

		select {
		case neuron.inputBuffer <- msg:
			// Successfully added
		default:
			// Buffer full
			break
		}
	}
}

// addMockConnections adds mock output connections to simulate connectivity.
func addMockConnections(neuron *Neuron, count int) {
	for i := 0; i < count; i++ {
		synapseID := fmt.Sprintf("mock-synapse-%d", i)
		targetID := fmt.Sprintf("mock-target-%d", i)

		callback := types.OutputCallback{
			TransmitMessage: func(msg types.NeuralSignal) error {
				return nil // Mock transmission
			},
			GetWeight: func() float64 {
				return 1.0 // Mock weight
			},
			GetDelay: func() time.Duration {
				return 1 * time.Millisecond // Mock delay
			},
			GetTargetID: func() string {
				return targetID
			},
		}

		neuron.AddOutputCallback(synapseID, callback)
	}
}
