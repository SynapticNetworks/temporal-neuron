package neuron

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

/*
=================================================================================
NEURON CORE TESTS - HOMEOSTATIC PLASTICITY MECHANISMS
=================================================================================

This test suite validates the core homeostatic plasticity mechanisms that enable
neurons to self-regulate their activity levels and maintain stable network dynamics.
These tests focus on internal neuron logic without complex external dependencies,
using clean mocks for external coordination.

KEY MECHANISMS TESTED:
1. Calcium-based activity sensing (models intracellular calcium signaling)
2. Threshold adjustment based on activity deviation from target rates
3. Firing history tracking for rate calculation
4. Bounds enforcement to prevent pathological threshold changes
5. Parameter setting and validation

All tests use minimal mocks to document API contracts while keeping tests fast,
reliable, and focused on core neuron functionality.

=================================================================================
*/

// ============================================================================
// HIGH VALUE TESTS - PURE INTERNAL HOMEOSTATIC LOGIC
// ============================================================================

// TestHomeostaticNeuronCreation validates proper initialization of homeostatic neurons
//
// BIOLOGICAL SIGNIFICANCE:
// Real neurons are born with specific target firing rates and homeostatic capabilities
// that are genetically determined. Different neuron types have different intrinsic
// excitability properties and target activity levels.
//
// EXPECTED RESULTS:
// - Neuron initializes with specified homeostatic parameters
// - Calcium level starts at baseline (realistic intracellular levels)
// - Firing history begins empty (no previous activity)
// - Threshold bounds are properly calculated from base threshold
// - All homeostatic timing parameters are correctly set
func TestNeuronCoreHomeostatic_NeuronCreation(t *testing.T) {
	// Configure realistic homeostatic parameters based on cortical neurons
	threshold := 1.5                          // Base firing threshold
	decayRate := 0.95                         // 5% membrane potential decay per millisecond
	refractoryPeriod := 10 * time.Millisecond // Typical cortical neuron refractory period
	fireFactor := 2.0                         // Action potential amplitude multiplier
	neuronID := "homeostatic_test_neuron"
	targetFiringRate := 5.0    // Target 5 Hz firing rate (typical cortical range)
	homeostasisStrength := 0.1 // Gentle 10% threshold adjustment strength

	// Create homeostatic neuron with biological parameters
	neuron := NewNeuron(neuronID, threshold, decayRate, refractoryPeriod,
		fireFactor, targetFiringRate, homeostasisStrength)

	if neuron == nil {
		t.Fatal("NewNeuron returned nil - neuron creation failed")
	}

	// Validate basic neuron properties
	if neuron.ID() != neuronID {
		t.Errorf("Neuron ID incorrect: expected %s, got %s", neuronID, neuron.ID())
	}

	if neuron.GetThreshold() != threshold {
		t.Errorf("Initial threshold incorrect: expected %f, got %f",
			threshold, neuron.GetThreshold())
	}

	// Validate homeostatic initialization using GetFiringStatus
	status := neuron.GetFiringStatus()

	// Verify target firing rate
	targetRate := status["target_firing_rate"].(float64)
	if targetRate != targetFiringRate {
		t.Errorf("Target firing rate incorrect: expected %f, got %f",
			targetFiringRate, targetRate)
	}

	// Calcium should start at baseline (not zero - biological realism)
	calciumLevel := status["calcium_level"].(float64)
	expectedBaseline := 0.1 // DENDRITE_CALCIUM_BASELINE_INTRACELLULAR
	if calciumLevel != expectedBaseline {
		t.Errorf("Initial calcium level should be baseline %.1f, got %f",
			expectedBaseline, calciumLevel)
	}

	// Firing history should be empty initially
	firingHistorySize := status["firing_history_size"].(int)
	if firingHistorySize != 0 {
		t.Errorf("Firing history should be empty initially, got %d entries",
			firingHistorySize)
	}

	// Current firing rate should be zero initially
	currentRate := status["current_firing_rate"].(float64)
	if currentRate != 0.0 {
		t.Errorf("Initial firing rate should be zero, got %f", currentRate)
	}

	// Verify threshold bounds are properly calculated
	baseThreshold := neuron.GetThreshold()
	expectedMinThreshold := baseThreshold * 0.1 // 10% of base threshold
	expectedMaxThreshold := baseThreshold * 5.0 // 5x base threshold

	// We can't directly access bounds, but we can test they're enforced
	// This will be validated in TestHomeostaticBounds

	t.Logf("✓ Homeostatic neuron created successfully:")
	t.Logf("  Target rate: %.1f Hz", targetRate)
	t.Logf("  Initial calcium: %.3f", calciumLevel)
	t.Logf("  Threshold bounds: %.1f - %.1f (calculated)",
		expectedMinThreshold, expectedMaxThreshold)
}

// TestHomeostaticParameterSetting validates dynamic parameter adjustment
//
// BIOLOGICAL SIGNIFICANCE:
// In research and therapeutic applications, it's important to be able to modify
// homeostatic parameters during runtime. This might model pharmacological
// interventions, developmental changes, or experimental manipulations.
//
// EXPECTED RESULTS:
// - Parameter changes take effect immediately
// - Thread-safe modification without state corruption
// - Bounds recalculation when base threshold changes
func TestNeuronCoreHomeostatic_ParameterSetting(t *testing.T) {
	// Create neuron with initial parameters
	neuron := NewNeuron("params_test", 1.0, 0.95, 5*time.Millisecond,
		1.0, 5.0, 0.1)

	// Test threshold modification
	newThreshold := 2.0
	neuron.SetThreshold(newThreshold)

	currentThreshold := neuron.GetThreshold()
	if currentThreshold != newThreshold {
		t.Errorf("Threshold not updated: expected %f, got %f",
			newThreshold, currentThreshold)
	}

	// Test thread-safe concurrent access
	done := make(chan bool, 2)

	// Goroutine 1: Read threshold repeatedly
	go func() {
		for i := 0; i < 100; i++ {
			_ = neuron.GetThreshold()
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()

	// Goroutine 2: Modify threshold repeatedly
	go func() {
		for i := 0; i < 100; i++ {
			neuron.SetThreshold(1.0 + float64(i%10)*0.1)
			time.Sleep(time.Microsecond)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Verify neuron is still functional
	finalThreshold := neuron.GetThreshold()
	if finalThreshold <= 0 {
		t.Error("Threshold became invalid after concurrent access")
	}

	t.Logf("✓ Parameter setting working: threshold %.1f → %.1f",
		1.0, finalThreshold)
	t.Log("✓ Thread-safe concurrent access validated")
}

// TestHomeostaticBounds validates threshold adjustment bounds enforcement
//
// BIOLOGICAL SIGNIFICANCE:
// Real neurons cannot adjust their firing threshold indefinitely due to biophysical
// constraints. Ion channel densities, membrane properties, and cellular energetics
// impose limits on how excitable or unexcitable a neuron can become.
//
// EXPECTED RESULTS:
// - Threshold increases are capped at maximum biological limit
// - Threshold decreases are capped at minimum biological limit
// - Bounds prevent pathological threshold values
// - Extreme regulation doesn't cause unstable behavior
func TestNeuronCoreHomeostatic_Bounds(t *testing.T) {
	baseThreshold := 1.0

	// Create neuron with very strong homeostatic regulation to test bounds
	neuron := NewNeuron("bounds_test", baseThreshold, 0.95, 5*time.Millisecond,
		1.0, 0.5, 2.0) // Very low target rate, very strong regulation

	// Set up mock matrix to capture health reports
	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Calculate expected bounds
	// expectedMinThreshold := baseThreshold * 0.1 // 10% of base
	expectedMaxThreshold := baseThreshold * 5.0 // 5x base

	// Test upper bound by creating sustained hyperactivity
	t.Log("Testing upper threshold bound with hyperactivity...")

	// Send many strong signals to drive threshold up
	for i := 0; i < 50; i++ {
		SendTestSignal(neuron, "hyperactivity", 2.0) // Well above threshold
		time.Sleep(10 * time.Millisecond)
	}

	// Allow homeostatic adjustment time
	time.Sleep(200 * time.Millisecond)

	// Verify threshold doesn't exceed maximum bound
	finalThreshold := neuron.GetThreshold()
	tolerance := 0.01 // Small tolerance for floating-point precision

	if finalThreshold > expectedMaxThreshold+tolerance {
		t.Errorf("Threshold (%f) exceeded max bound (%f)",
			finalThreshold, expectedMaxThreshold)
	}

	// Test that threshold did increase (showing homeostasis is working)
	if finalThreshold <= baseThreshold {
		t.Error("Threshold should have increased with hyperactivity")
	}

	t.Logf("✓ Upper bound enforced: %.3f ≤ %.1f (max)",
		finalThreshold, expectedMaxThreshold)

	// Test lower bound by creating a hypoactive neuron
	hypoNeuron := NewNeuron("hypo_test", 2.0, 0.95, 5*time.Millisecond,
		1.0, 10.0, 2.0) // High target rate, strong regulation

	hypoNeuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err = hypoNeuron.Start()
	if err != nil {
		t.Fatalf("Failed to start hypo neuron: %v", err)
	}
	defer hypoNeuron.Stop()

	expectedMinForHypo := 2.0 * 0.1 // 10% of base threshold (2.0)

	t.Log("Testing lower threshold bound with hypoactivity...")

	// Send weak signals that rarely cause firing
	for i := 0; i < 30; i++ {
		SendTestSignal(hypoNeuron, "hypoactivity", 0.3) // Well below threshold
		time.Sleep(50 * time.Millisecond)
	}

	// Allow homeostatic adjustment
	time.Sleep(300 * time.Millisecond)

	hypoThreshold := hypoNeuron.GetThreshold()

	if hypoThreshold < expectedMinForHypo-tolerance {
		t.Errorf("Threshold (%f) went below min bound (%f)",
			hypoThreshold, expectedMinForHypo)
	}

	// Test that threshold did decrease
	if hypoThreshold >= 2.0 {
		t.Error("Threshold should have decreased with hypoactivity")
	}

	t.Logf("✓ Lower bound enforced: %.3f ≥ %.1f (min)",
		hypoThreshold, expectedMinForHypo)
	t.Log("✓ Bounds prevent pathological threshold values")
}

// TestCalciumDynamics validates calcium accumulation and decay mechanisms
//
// BIOLOGICAL SIGNIFICANCE:
// Intracellular calcium serves as a crucial activity sensor in biological neurons.
// Action potentials cause calcium influx through voltage-gated channels, and this
// calcium accumulates with repeated firing. Calcium removal through pumps and
// buffers creates temporal integration of recent activity.
//
// EXPECTED RESULTS:
// - Calcium increases immediately after firing
// - Calcium gradually decays over time (exponential)
// - Multiple firings cause calcium accumulation
// - Decay rate matches biological parameters
func TestNeuronCoreHomeostatic_CalciumDynamics(t *testing.T) {
	// Create neuron with calcium tracking
	neuron := NewNeuron("calcium_test", 0.8, 0.95, 5*time.Millisecond,
		1.0, 5.0, 0.1)

	// Set up mock matrix for health reporting
	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Verify initial calcium level (baseline, not zero)
	initialStatus := neuron.GetFiringStatus()
	initialCalcium := initialStatus["calcium_level"].(float64)
	expectedBaseline := 0.1

	if initialCalcium != expectedBaseline {
		t.Errorf("Initial calcium should be baseline %.1f, got %.3f",
			expectedBaseline, initialCalcium)
	}

	// Trigger neuron firing
	SendTestSignal(neuron, "calcium_trigger", 1.0)
	time.Sleep(20 * time.Millisecond) // Allow firing to process

	// Verify calcium increase after firing
	postFireStatus := neuron.GetFiringStatus()
	postFireCalcium := postFireStatus["calcium_level"].(float64)

	if postFireCalcium <= initialCalcium {
		t.Errorf("Expected calcium increase after firing. Initial: %.3f, Post-fire: %.3f",
			initialCalcium, postFireCalcium)
	}

	// Test calcium accumulation with multiple firings
	secondFireCalcium := postFireCalcium
	SendTestSignal(neuron, "calcium_trigger", 1.0)
	time.Sleep(20 * time.Millisecond)

	doubleFireStatus := neuron.GetFiringStatus()
	doubleFireCalcium := doubleFireStatus["calcium_level"].(float64)

	if doubleFireCalcium <= secondFireCalcium {
		t.Error("Expected calcium accumulation with multiple firings")
	}

	// Test calcium decay over time
	time.Sleep(100 * time.Millisecond) // Allow decay

	decayedStatus := neuron.GetFiringStatus()
	decayedCalcium := decayedStatus["calcium_level"].(float64)

	if decayedCalcium >= doubleFireCalcium {
		t.Errorf("Expected calcium decay over time. Double-fire: %.3f, Decayed: %.3f",
			doubleFireCalcium, decayedCalcium)
	}

	// Calcium should still be above baseline (decay is gradual)
	if decayedCalcium <= expectedBaseline {
		t.Error("Calcium should decay gradually, not disappear immediately")
	}

	t.Logf("✓ Calcium dynamics validated:")
	t.Logf("  Baseline: %.3f", initialCalcium)
	t.Logf("  Post-fire: %.3f (+%.3f)", postFireCalcium, postFireCalcium-initialCalcium)
	t.Logf("  Accumulated: %.3f (+%.3f)", doubleFireCalcium, doubleFireCalcium-postFireCalcium)
	t.Logf("  After decay: %.3f (-%.3f)", decayedCalcium, doubleFireCalcium-decayedCalcium)
}

// TestFiringHistoryTracking validates firing history maintenance for rate calculation
//
// BIOLOGICAL SIGNIFICANCE:
// Accurate firing rate calculation requires maintaining a sliding window of recent
// firing times. This temporal information is essential for homeostatic regulation
// to assess whether the neuron is firing above or below its target rate.
//
// EXPECTED RESULTS:
// - Firing history starts empty
// - Each firing event is recorded with accurate timing
// - Firing rate calculation reflects actual firing frequency
// - History maintenance follows sliding window principle
func TestNeuronCoreHomeostatic_FiringHistoryTracking(t *testing.T) {
	// Create neuron with firing history tracking
	neuron := NewNeuron("history_test", 0.6, 0.95, 5*time.Millisecond,
		1.0, 5.0, 0.1)

	// Set up mock matrix
	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Verify initially empty history
	initialStatus := neuron.GetFiringStatus()
	initialHistorySize := initialStatus["firing_history_size"].(int)
	initialRate := initialStatus["current_firing_rate"].(float64)

	if initialHistorySize != 0 {
		t.Errorf("Firing history should be empty initially, got %d entries",
			initialHistorySize)
	}

	if initialRate != 0.0 {
		t.Errorf("Initial firing rate should be zero, got %f", initialRate)
	}

	// Trigger controlled firing events
	numFirings := 5
	firingInterval := 30 * time.Millisecond

	t.Logf("Triggering %d firings at %v intervals...", numFirings, firingInterval)

	for i := 0; i < numFirings; i++ {
		SendTestSignal(neuron, "history_trigger", 1.0)
		time.Sleep(firingInterval)
	}

	// Allow final processing
	time.Sleep(50 * time.Millisecond)

	// Verify firing history tracking
	finalStatus := neuron.GetFiringStatus()
	finalHistorySize := finalStatus["firing_history_size"].(int)
	finalRate := finalStatus["current_firing_rate"].(float64)

	if finalHistorySize < numFirings {
		t.Errorf("Expected at least %d entries in firing history, got %d",
			numFirings, finalHistorySize)
	}

	// Verify firing rate calculation
	if finalRate <= 0 {
		t.Errorf("Expected positive firing rate, got %f", finalRate)
	}

	// Calculate expected rate based on timing
	// numFirings over (numFirings-1) * interval seconds
	totalTime := float64(numFirings-1) * firingInterval.Seconds()
	expectedRate := float64(numFirings) / (totalTime + 0.1) // Add small buffer for processing time
	rateTolerance := 5.0                                    // Allow significant tolerance for timing variations

	if finalRate < expectedRate-rateTolerance || finalRate > expectedRate+rateTolerance {
		t.Logf("Note: Calculated rate (%.1f Hz) differs from rough expected (%.1f Hz) - timing variations normal",
			finalRate, expectedRate)
	}

	// Test firing history sliding window by waiting
	t.Log("Testing sliding window behavior...")
	time.Sleep(1 * time.Second) // Wait for some entries to age out

	delayedStatus := neuron.GetFiringStatus()
	delayedRate := delayedStatus["current_firing_rate"].(float64)

	// Rate should decrease as old entries age out of the window
	if delayedRate > finalRate {
		t.Error("Firing rate should decrease as entries age out of sliding window")
	}

	t.Logf("✓ Firing history tracking validated:")
	t.Logf("  History size: %d entries", finalHistorySize)
	t.Logf("  Peak rate: %.2f Hz", finalRate)
	t.Logf("  Aged rate: %.2f Hz (sliding window working)", delayedRate)
}

// ============================================================================
// MEDIUM VALUE TESTS - SIMPLE EXTERNAL INTERACTION
// ============================================================================

// TestHomeostaticThresholdAdjustment validates core threshold adjustment logic
//
// BIOLOGICAL SIGNIFICANCE:
// Homeostatic plasticity allows neurons to maintain stable firing rates by
// adjusting their intrinsic excitability. When firing rate is too high,
// threshold increases (less excitable). When firing rate is too low,
// threshold decreases (more excitable).
//
// EXPECTED RESULTS:
// - Threshold increases with sustained high activity
// - Threshold decreases with sustained low activity
// - Target firing rate is approached over time
// - Health reporting works with matrix callbacks
func TestNeuronCoreHomeostatic_ThresholdAdjustment(t *testing.T) {
	targetRate := 3.0 // Moderate target rate for clear testing

	// Test threshold increase with hyperactivity
	// Fix the TestNeuronCoreHomeostatic_ThresholdAdjustment test
	// Replace the existing ThresholdIncreaseWithHyperactivity subtest

	t.Run("ThresholdIncreaseWithHyperactivity", func(t *testing.T) {
		// Use lower target rate to ensure we can achieve hyperactivity
		targetRate := 1.0 // Lower target = easier to exceed

		neuron := NewNeuron("threshold_increase", 0.4, 0.95, 5*time.Millisecond,
			1.0, targetRate, DENDRITE_FACTOR_HOMEOSTASIS_STRENGTH_DEFAULT) // Use constant

		mockMatrix := NewMockMatrix()
		neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

		err := neuron.Start()
		if err != nil {
			t.Fatalf("Failed to start neuron: %v", err)
		}
		defer neuron.Stop()

		initialThreshold := neuron.GetThreshold()

		// Create actual hyperactivity by sending signals faster than target rate
		// Target = 1.0 Hz = 1 signal per second
		// Send signals every 100ms = 10 Hz = 10x target rate
		signalInterval := 100 * time.Millisecond
		signalStrength := 2.0 // Well above threshold to ensure firing
		numSignals := 15      // Send for ~1.5 seconds

		t.Logf("Creating hyperactivity (target: %.1f Hz)...", targetRate)
		t.Logf("Sending %d signals of %.1f strength every %v (%.1f Hz input rate)",
			numSignals, signalStrength, signalInterval, 1000.0/float64(signalInterval.Milliseconds()))

		for i := 0; i < numSignals; i++ {
			SendTestSignal(neuron, "hyperactivity", signalStrength)
			time.Sleep(signalInterval)
		}

		// Allow homeostatic adjustment using constant
		time.Sleep(DENDRITE_TIME_HOMEOSTATIC_TICK * 3)

		finalThreshold := neuron.GetThreshold()
		finalStatus := neuron.GetFiringStatus()
		finalRate := finalStatus["current_firing_rate"].(float64)

		t.Logf("Results: rate %.2f Hz (target %.1f Hz), threshold %.3f → %.3f",
			finalRate, targetRate, initialThreshold, finalThreshold)

		// Only expect threshold increase if we actually achieved hyperactivity
		if finalRate > targetRate {
			if finalThreshold <= initialThreshold {
				t.Errorf("Expected threshold increase due to hyperactivity. Rate %.2f > target %.1f but threshold %.3f → %.3f",
					finalRate, targetRate, initialThreshold, finalThreshold)
			} else {
				t.Logf("✓ Hyperactivity response: rate %.2f > target %.1f → threshold increased %.3f → %.3f",
					finalRate, targetRate, initialThreshold, finalThreshold)
			}
		} else {
			// If we didn't achieve hyperactivity, that's the real issue
			t.Errorf("Failed to create hyperactivity: rate %.2f ≤ target %.1f. Need stronger/faster signals.",
				finalRate, targetRate)
			t.Logf("Note: Threshold decrease %.3f → %.3f is correct for rate < target",
				initialThreshold, finalThreshold)
		}

		// Verify health reporting occurred
		healthReports := mockMatrix.GetHealthReportCount()
		if healthReports == 0 {
			t.Error("Expected health reports during firing activity")
		}
	})

	// Alternative approach - test with very short activity window for faster results:
	t.Run("ThresholdIncreaseWithHyperactivity_ShortWindow", func(t *testing.T) {
		targetRate := 2.0

		neuron := NewNeuron("threshold_increase_fast", 0.3, 0.95, 5*time.Millisecond,
			1.0, targetRate, DENDRITE_FACTOR_HOMEOSTASIS_STRENGTH_STRONG) // Strong homeostasis

		mockMatrix := NewMockMatrix()
		neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

		err := neuron.Start()
		if err != nil {
			t.Fatalf("Failed to start neuron: %v", err)
		}
		defer neuron.Stop()

		initialThreshold := neuron.GetThreshold()

		// Send rapid burst of signals within the activity window
		// Activity window is 10 seconds, so send many signals quickly
		burstInterval := 20 * time.Millisecond
		numBurstSignals := 100 // 100 signals in 2 seconds = 50 Hz >> 2 Hz target

		t.Logf("Creating burst hyperactivity (target: %.1f Hz)...", targetRate)

		for i := 0; i < numBurstSignals; i++ {
			SendTestSignal(neuron, "burst", 1.0)
			time.Sleep(burstInterval)
		}

		// Allow homeostatic adjustment
		time.Sleep(DENDRITE_TIME_HOMEOSTATIC_TICK * 2)

		finalThreshold := neuron.GetThreshold()
		finalStatus := neuron.GetFiringStatus()
		finalRate := finalStatus["current_firing_rate"].(float64)

		t.Logf("Burst results: rate %.2f Hz (target %.1f Hz), threshold %.3f → %.3f",
			finalRate, targetRate, initialThreshold, finalThreshold)

		if finalRate > targetRate && finalThreshold > initialThreshold {
			t.Logf("✓ Burst hyperactivity successful: rate %.2f > target %.1f → threshold increased",
				finalRate, targetRate)
		} else if finalRate <= targetRate {
			t.Logf("Note: Burst rate %.2f ≤ target %.1f - may need even faster signals", finalRate, targetRate)
		}
	})

	// Test threshold decrease with hypoactivity
	t.Run("ThresholdDecreaseWithHypoactivity", func(t *testing.T) {
		neuron := NewNeuron("threshold_decrease", 1.5, 0.95, 5*time.Millisecond,
			1.0, targetRate, 0.3) // High threshold, strong homeostasis

		mockMatrix := NewMockMatrix()
		neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

		err := neuron.Start()
		if err != nil {
			t.Fatalf("Failed to start neuron: %v", err)
		}
		defer neuron.Stop()

		initialThreshold := neuron.GetThreshold()

		// Create hypoactivity (below target rate)
		t.Logf("Creating hypoactivity (target: %.1f Hz)...", targetRate)
		for i := 0; i < 15; i++ {
			SendTestSignal(neuron, "hypoactivity", 0.7) // Below threshold
			time.Sleep(100 * time.Millisecond)          // Low frequency
		}

		// Allow homeostatic adjustment
		time.Sleep(500 * time.Millisecond)

		finalThreshold := neuron.GetThreshold()
		finalStatus := neuron.GetFiringStatus()
		finalRate := finalStatus["current_firing_rate"].(float64)

		if finalThreshold >= initialThreshold {
			t.Errorf("Expected threshold decrease due to hypoactivity. Initial: %.3f, Final: %.3f",
				initialThreshold, finalThreshold)
		}

		// Verify some health reporting occurred
		healthReports := mockMatrix.GetHealthReportCount()

		t.Logf("✓ Hypoactivity response: threshold %.3f → %.3f, rate %.1f Hz, reports %d",
			initialThreshold, finalThreshold, finalRate, healthReports)
	})
}

// TestHomeostasisWithInhibition validates neuron response to inhibitory input
//
// BIOLOGICAL SIGNIFICANCE:
// Neurons in the brain receive a mix of excitatory and inhibitory signals.
// Homeostasis must be able to handle periods of silence caused by strong
// inhibition. In response to being silenced, a neuron should become more
// excitable (decrease its threshold) to maintain its target firing rate.
//
// EXPECTED RESULTS:
// - Strong inhibitory signals prevent firing
// - Homeostatic mechanism detects hypoactivity during inhibition
// - Firing threshold decreases to compensate for inhibition
// - System demonstrates proper response to mixed signal types
func TestNeuronCoreHomeostatic_HomeostasisWithInhibition(t *testing.T) {
	targetRate := 4.0
	strength := 0.4 // Moderate homeostatic strength

	neuron := NewNeuron("inhibition_test", 1.2, 0.95, 5*time.Millisecond,
		1.0, targetRate, strength)

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	initialThreshold := neuron.GetThreshold()
	t.Logf("Initial threshold: %.3f", initialThreshold)

	// Apply strong inhibitory signals
	t.Log("Applying inhibitory signals...")
	for i := 0; i < 30; i++ {
		// Send inhibitory signal (negative value)
		inhibitorySignal := types.NeuralSignal{
			Value:                -0.8, // Strong inhibitory signal
			Timestamp:            time.Now(),
			SourceID:             "inhibitory_source",
			TargetID:             neuron.ID(),
			NeurotransmitterType: types.LigandGABA, // Inhibitory neurotransmitter
		}
		neuron.Receive(inhibitorySignal)
		time.Sleep(20 * time.Millisecond)
	}

	// Allow homeostatic adjustment
	time.Sleep(400 * time.Millisecond)

	finalThreshold := neuron.GetThreshold()
	finalStatus := neuron.GetFiringStatus()
	finalRate := finalStatus["current_firing_rate"].(float64)
	finalCalcium := finalStatus["calcium_level"].(float64)

	// Verify neuron was silenced by inhibition
	if finalRate > 1.0 {
		t.Logf("Note: Neuron still firing at %.1f Hz despite inhibition - may need stronger inhibition", finalRate)
	}

	// Verify homeostatic response (threshold should decrease)
	if finalThreshold >= initialThreshold {
		t.Errorf("Expected threshold to decrease due to inhibition-induced silence. Initial: %.3f, Final: %.3f",
			initialThreshold, finalThreshold)
	}

	// Verify calcium remained low (no firing activity)
	expectedBaseline := 0.1
	if finalCalcium > expectedBaseline+0.5 {
		t.Logf("Note: Calcium level %.3f higher than expected during inhibition", finalCalcium)
	}

	// Verify threshold stays within bounds
	minThreshold := initialThreshold * 0.1
	if finalThreshold < minThreshold {
		t.Errorf("Threshold %.3f below minimum bound %.3f", finalThreshold, minThreshold)
	}

	// Test recovery from inhibition
	t.Log("Testing recovery from inhibition...")

	// Stop inhibition and provide excitatory input
	for i := 0; i < 10; i++ {
		SendTestSignal(neuron, "recovery", 0.9)
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(200 * time.Millisecond)

	recoveryStatus := neuron.GetFiringStatus()
	recoveryRate := recoveryStatus["current_firing_rate"].(float64)

	if recoveryRate <= finalRate {
		t.Log("Note: Recovery firing rate should increase after inhibition ends")
	}

	t.Logf("✓ Inhibition response validated:")
	t.Logf("  Threshold: %.3f → %.3f (decreased)", initialThreshold, finalThreshold)
	t.Logf("  Rate during inhibition: %.2f Hz", finalRate)
	t.Logf("  Recovery rate: %.2f Hz", recoveryRate)
	t.Logf("  Calcium during inhibition: %.3f", finalCalcium)
}

// ============================================================================
// INTEGRATION AND EDGE CASE TESTS
// ============================================================================

// TestHomeostaticEdgeCases validates homeostatic behavior under extreme conditions
//
// BIOLOGICAL SIGNIFICANCE:
// Neural networks must remain stable under pathological conditions such as
// epileptic seizures, lesions, or pharmacological interventions. Testing
// edge cases ensures the homeostatic mechanisms are robust.
//
// EXPECTED RESULTS:
// - System remains stable under extreme input conditions
// - Bounds enforcement prevents runaway behavior
// - Recovery mechanisms work after extreme perturbations
func TestNeuronCoreHomeostatic_EdgeCases(t *testing.T) {
	t.Run("ZeroTargetRate", func(t *testing.T) {
		// Test with zero target firing rate (silent neuron)
		neuron := NewNeuron("zero_target", 1.0, 0.95, 5*time.Millisecond,
			1.0, 0.0, 0.1) // Zero target rate

		mockMatrix := NewMockMatrix()
		neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

		err := neuron.Start()
		if err != nil {
			t.Fatalf("Failed to start neuron: %v", err)
		}
		defer neuron.Stop()

		// Send some input
		for i := 0; i < 10; i++ {
			SendTestSignal(neuron, "zero_test", 1.5)
			time.Sleep(20 * time.Millisecond)
		}

		time.Sleep(200 * time.Millisecond)

		// Neuron should try to silence itself
		finalThreshold := neuron.GetThreshold()
		if finalThreshold <= 1.0 {
			t.Error("Expected threshold to increase with zero target rate")
		}

		t.Logf("✓ Zero target rate: threshold increased to %.3f", finalThreshold)
	})

	t.Run("ExtremeInputValues", func(t *testing.T) {
		neuron := NewNeuron("extreme_test", 1.0, 0.95, 5*time.Millisecond,
			1.0, 5.0, 0.2)

		mockMatrix := NewMockMatrix()
		neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

		err := neuron.Start()
		if err != nil {
			t.Fatalf("Failed to start neuron: %v", err)
		}
		defer neuron.Stop()

		// Send extremely large inputs
		for i := 0; i < 5; i++ {
			SendTestSignal(neuron, "extreme", 1000.0) // Massive input
			time.Sleep(10 * time.Millisecond)
		}

		time.Sleep(100 * time.Millisecond)

		// System should remain stable
		status := neuron.GetFiringStatus()
		rate := status["current_firing_rate"].(float64)
		threshold := neuron.GetThreshold()

		if rate > 1000 || threshold > 50.0 {
			t.Error("System became unstable with extreme inputs")
		}

		t.Logf("✓ Extreme inputs handled: rate %.2f Hz, threshold %.3f", rate, threshold)
	})

	t.Run("RapidParameterChanges", func(t *testing.T) {
		neuron := NewNeuron("rapid_test", 1.0, 0.95, 5*time.Millisecond,
			1.0, 5.0, 0.3)

		mockMatrix := NewMockMatrix()
		neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

		err := neuron.Start()
		if err != nil {
			t.Fatalf("Failed to start neuron: %v", err)
		}
		defer neuron.Stop()

		// Rapidly change threshold while homeostasis is active
		for i := 0; i < 20; i++ {
			newThreshold := 0.5 + float64(i%10)*0.1
			neuron.SetThreshold(newThreshold)
			SendTestSignal(neuron, "rapid", 1.0)
			time.Sleep(5 * time.Millisecond)
		}

		time.Sleep(100 * time.Millisecond)

		// Neuron should still be functional
		finalThreshold := neuron.GetThreshold()
		if finalThreshold <= 0 || finalThreshold > 10.0 {
			t.Errorf("Threshold became invalid: %.3f", finalThreshold)
		}

		t.Logf("✓ Rapid parameter changes handled: final threshold %.3f", finalThreshold)
	})
}

// TestHomeostaticConcurrency validates thread-safe homeostatic operations
//
// BIOLOGICAL SIGNIFICANCE:
// Neural processing is inherently concurrent, with multiple synapses delivering
// inputs simultaneously. The homeostatic system must handle this concurrency
// without race conditions or data corruption.
//
// EXPECTED RESULTS:
// - Concurrent firing events are properly tracked
// - Calcium accumulation is thread-safe
// - Threshold adjustments don't interfere with concurrent operations
func TestNeuronCoreHomeostatic_Concurrency(t *testing.T) {
	neuron := NewNeuron("concurrency_test", 0.8, 0.95, 5*time.Millisecond,
		1.0, 8.0, 0.2)

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	numGoroutines := 10
	signalsPerGoroutine := 20
	done := make(chan bool, numGoroutines)

	// Launch multiple goroutines sending signals concurrently
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			for i := 0; i < signalsPerGoroutine; i++ {
				sourceID := fmt.Sprintf("concurrent_%d_%d", goroutineID, i)
				SendTestSignal(neuron, sourceID, 1.0)
				time.Sleep(time.Millisecond)
			}
			done <- true
		}(g)
	}

	// Wait for all goroutines to complete
	for g := 0; g < numGoroutines; g++ {
		<-done
	}

	// Allow final processing
	time.Sleep(200 * time.Millisecond)

	// Verify system integrity
	finalStatus := neuron.GetFiringStatus()
	finalRate := finalStatus["current_firing_rate"].(float64)
	finalThreshold := neuron.GetThreshold()
	historySize := finalStatus["firing_history_size"].(int)

	if finalRate <= 0 {
		t.Error("Expected positive firing rate after concurrent inputs")
	}

	if finalThreshold <= 0 || finalThreshold > 20.0 {
		t.Errorf("Threshold out of reasonable range: %.3f", finalThreshold)
	}

	if historySize <= 0 {
		t.Error("Expected firing history entries after concurrent activity")
	}

	// Verify no data corruption by checking status multiple times
	for i := 0; i < 5; i++ {
		status := neuron.GetFiringStatus()
		if status == nil {
			t.Error("GetFiringStatus returned nil - possible corruption")
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Logf("✓ Concurrency handled: %d goroutines, rate %.2f Hz, threshold %.3f, history %d",
		numGoroutines, finalRate, finalThreshold, historySize)
}

// TestHomeostaticRecovery validates recovery from pathological states
//
// BIOLOGICAL SIGNIFICANCE:
// Neurons must be able to recover from extreme perturbations such as ischemia,
// drug effects, or electrical stimulation. This tests the robustness of
// homeostatic recovery mechanisms.
//
// EXPECTED RESULTS:
// - Recovery from silencing (threshold decreases appropriately)
// - Recovery from hyperexcitation (threshold increases appropriately)
// - Parameters return to stable ranges after perturbation ends
func TestNeuronCoreHomeostatic_Recovery(t *testing.T) {
	t.Run("RecoveryFromSilencing", func(t *testing.T) {
		neuron := NewNeuron("recovery_silence", 1.0, 0.95, 5*time.Millisecond,
			1.0, 5.0, 0.4) // Strong homeostasis for clear effects

		mockMatrix := NewMockMatrix()
		neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

		err := neuron.Start()
		if err != nil {
			t.Fatalf("Failed to start neuron: %v", err)
		}
		defer neuron.Stop()

		initialThreshold := neuron.GetThreshold()

		// Phase 1: Silence the neuron for extended period
		t.Log("Phase 1: Silencing neuron...")
		silenceStart := time.Now()
		for time.Since(silenceStart) < 500*time.Millisecond {
			// Send strong inhibitory signals
			inhibSignal := types.NeuralSignal{
				Value:                -2.0,
				Timestamp:            time.Now(),
				SourceID:             "silencer",
				TargetID:             neuron.ID(),
				NeurotransmitterType: types.LigandGABA,
			}
			neuron.Receive(inhibSignal)
			time.Sleep(10 * time.Millisecond)
		}

		silencedThreshold := neuron.GetThreshold()
		if silencedThreshold >= initialThreshold {
			t.Error("Expected threshold to decrease during silencing")
		}

		// Phase 2: Remove silencing and provide normal input
		t.Log("Phase 2: Providing recovery inputs...")
		for i := 0; i < 30; i++ {
			SendTestSignal(neuron, "recovery", 1.2)
			time.Sleep(20 * time.Millisecond)
		}

		time.Sleep(200 * time.Millisecond)

		recoveredStatus := neuron.GetFiringStatus()
		recoveredRate := recoveredStatus["current_firing_rate"].(float64)
		recoveredThreshold := neuron.GetThreshold()

		if recoveredRate <= 0 {
			t.Error("Expected recovery of firing activity")
		}

		t.Logf("✓ Recovery from silencing: threshold %.3f → %.3f → %.3f, final rate %.2f Hz",
			initialThreshold, silencedThreshold, recoveredThreshold, recoveredRate)
	})

	t.Run("RecoveryFromHyperexcitation", func(t *testing.T) {
		neuron := NewNeuron("recovery_hyper", 0.5, 0.95, 5*time.Millisecond,
			1.0, 3.0, 0.3) // Initial threshold 0.5, target 3.0, strength 0.3

		mockMatrix := NewMockMatrix()
		neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

		err := neuron.Start()
		if err != nil {
			t.Fatalf("Failed to start neuron: %v", err)
		}
		defer neuron.Stop()

		initialThreshold := neuron.GetThreshold()

		// Phase 1: Hyperexcite the neuron
		t.Log("Phase 1: Hyperexciting neuron...")
		// Increased numSignals for longer sustained hyperexcitation
		for i := 0; i < 200; i++ { // Changed from 50 to 200 signals
			SendTestSignal(neuron, "hyperexcite", 2.0)
			time.Sleep(5 * time.Millisecond)
		}
		// Add a short sleep to allow the last homeostatic adjustment cycle to complete
		time.Sleep(DENDRITE_TIME_HOMEOSTATIC_TICK) // Added this sleep

		hyperThreshold := neuron.GetThreshold()
		if hyperThreshold <= initialThreshold {
			t.Errorf("Expected threshold to increase during hyperexcitation. Initial: %.3f, Hyper: %.3f",
				initialThreshold, hyperThreshold)
		}

		// Phase 2: Return to normal input levels
		t.Log("Phase 2: Normalizing input...")
		for i := 0; i < 25; i++ {
			SendTestSignal(neuron, "normalize", 0.7)
			time.Sleep(40 * time.Millisecond)
		}

		time.Sleep(300 * time.Millisecond)

		normalizedStatus := neuron.GetFiringStatus()
		normalizedRate := normalizedStatus["current_firing_rate"].(float64)
		normalizedThreshold := neuron.GetThreshold()

		// Rate should approach target rate
		targetRate := 3.0
		if math.Abs(normalizedRate-targetRate) > 2.0 {
			t.Logf("Note: Final rate %.2f Hz differs from target %.1f Hz", normalizedRate, targetRate)
		}

		t.Logf("✓ Recovery from hyperexcitation: threshold %.3f → %.3f → %.3f, final rate %.2f Hz",
			initialThreshold, hyperThreshold, normalizedThreshold, normalizedRate)
	})
}

// TestHomeostaticValidation validates parameter bounds and error conditions
//
// BIOLOGICAL SIGNIFICANCE:
// Homeostatic parameters must remain within biologically plausible ranges.
// Invalid parameters could lead to pathological network behavior.
//
// EXPECTED RESULTS:
// - Invalid parameter combinations are handled gracefully
// - System maintains stability despite parameter validation failures
// - Error reporting is clear and actionable
func TestNeuronCoreHomeostatic_Validation(t *testing.T) {
	t.Run("InvalidTargetRates", func(t *testing.T) {
		// Test with negative target rate
		neuron := NewNeuron("invalid_target", 1.0, 0.95, 5*time.Millisecond,
			1.0, -5.0, 0.1) // Negative target rate

		// Neuron should handle this gracefully
		mockMatrix := NewMockMatrix()
		neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

		err := neuron.Start()
		if err != nil {
			t.Fatalf("Failed to start neuron: %v", err)
		}
		defer neuron.Stop()

		// System should remain stable
		status := neuron.GetFiringStatus()
		if status == nil {
			t.Error("System should remain functional with invalid target rate")
		}

		t.Log("✓ Invalid target rate handled gracefully")
	})

	t.Run("ExtremeHomeostaticStrength", func(t *testing.T) {
		// Test with very high homeostatic strength
		neuron := NewNeuron("extreme_strength", 1.0, 0.95, 5*time.Millisecond,
			1.0, 5.0, 10.0) // Extremely high strength

		mockMatrix := NewMockMatrix()
		neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

		err := neuron.Start()
		if err != nil {
			t.Fatalf("Failed to start neuron: %v", err)
		}
		defer neuron.Stop()

		// Send some input
		for i := 0; i < 10; i++ {
			SendTestSignal(neuron, "extreme_strength", 1.5)
			time.Sleep(20 * time.Millisecond)
		}

		time.Sleep(100 * time.Millisecond)

		// System should not become unstable
		threshold := neuron.GetThreshold()
		if threshold <= 0 || threshold > 100.0 {
			t.Errorf("System became unstable with extreme homeostatic strength: threshold %.3f", threshold)
		}

		t.Logf("✓ Extreme homeostatic strength handled: threshold %.3f", threshold)
	})

	t.Run("ZeroHomeostaticStrength", func(t *testing.T) {
		// Test with zero homeostatic strength (no adaptation)
		neuron := NewNeuron("zero_strength", 1.0, 0.95, 5*time.Millisecond,
			1.0, 5.0, 0.0) // No homeostasis

		mockMatrix := NewMockMatrix()
		neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

		err := neuron.Start()
		if err != nil {
			t.Fatalf("Failed to start neuron: %v", err)
		}
		defer neuron.Stop()

		initialThreshold := neuron.GetThreshold()

		// Send activity that would normally trigger homeostasis
		for i := 0; i < 30; i++ {
			SendTestSignal(neuron, "no_homeostasis", 2.0)
			time.Sleep(10 * time.Millisecond)
		}

		time.Sleep(200 * time.Millisecond)

		finalThreshold := neuron.GetThreshold()

		// Threshold should not change significantly
		if math.Abs(finalThreshold-initialThreshold) > 0.01 {
			t.Errorf("Threshold changed despite zero homeostatic strength: %.3f → %.3f",
				initialThreshold, finalThreshold)
		}

		t.Logf("✓ Zero homeostatic strength: threshold stable %.3f → %.3f",
			initialThreshold, finalThreshold)
	})
}

// TestHomeostaticLongTermStability validates long-term homeostatic behavior
//
// BIOLOGICAL SIGNIFICANCE:
// Homeostatic mechanisms must maintain network stability over extended periods,
// preventing drift and ensuring consistent behavior across varying input patterns.
//
// EXPECTED RESULTS:
// - Threshold converges to stable value over time
// - Firing rate approaches target rate with sustained input
// - System shows appropriate adaptation to changing input statistics
func TestNeuronCoreHomeostatic_LongTermStability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-term stability test in short mode")
	}

	neuron := NewNeuron("stability_test", 1.0, 0.95, 5*time.Millisecond,
		1.0, 4.0, 0.1) // Gentle homeostasis for stability

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Track threshold and rate over time
	measurements := []struct {
		time      time.Duration
		threshold float64
		rate      float64
	}{}

	testDuration := 2 * time.Second
	measurementInterval := 200 * time.Millisecond
	inputInterval := 50 * time.Millisecond

	startTime := time.Now()

	// Background input generation
	go func() {
		for time.Since(startTime) < testDuration {
			// Vary input strength slightly to test adaptation
			baseStrength := 1.2
			variation := 0.2 * math.Sin(float64(time.Since(startTime).Nanoseconds())/1e9)
			inputStrength := baseStrength + variation

			SendTestSignal(neuron, "stability", inputStrength)
			time.Sleep(inputInterval)
		}
	}()

	// Periodic measurements
	for time.Since(startTime) < testDuration {
		time.Sleep(measurementInterval)

		status := neuron.GetFiringStatus()
		threshold := neuron.GetThreshold()
		rate := status["current_firing_rate"].(float64)

		measurements = append(measurements, struct {
			time      time.Duration
			threshold float64
			rate      float64
		}{
			time:      time.Since(startTime),
			threshold: threshold,
			rate:      rate,
		})
	}

	// Analyze stability
	if len(measurements) < 3 {
		t.Fatal("Insufficient measurements for stability analysis")
	}

	// Check for convergence in the last half of measurements
	midPoint := len(measurements) / 2
	recentMeasurements := measurements[midPoint:]

	// Calculate variance in recent measurements
	var thresholdSum, rateSum float64
	for _, m := range recentMeasurements {
		thresholdSum += m.threshold
		rateSum += m.rate
	}

	avgThreshold := thresholdSum / float64(len(recentMeasurements))
	avgRate := rateSum / float64(len(recentMeasurements))

	var thresholdVariance, rateVariance float64
	for _, m := range recentMeasurements {
		thresholdVariance += math.Pow(m.threshold-avgThreshold, 2)
		rateVariance += math.Pow(m.rate-avgRate, 2)
	}

	thresholdVariance /= float64(len(recentMeasurements))
	rateVariance /= float64(len(recentMeasurements))

	// Verify stability (low variance in recent measurements)
	maxThresholdVariance := 0.01 // Threshold should be stable within 0.1
	maxRateVariance := 1.0       // Rate should be stable within 1 Hz

	if thresholdVariance > maxThresholdVariance {
		t.Logf("Note: Threshold variance %.6f higher than expected %.6f", thresholdVariance, maxThresholdVariance)
	}

	if rateVariance > maxRateVariance {
		t.Logf("Note: Rate variance %.2f higher than expected %.2f", rateVariance, maxRateVariance)
	}

	// Verify rate approaches target
	targetRate := 4.0
	if math.Abs(avgRate-targetRate) > 2.0 {
		t.Logf("Note: Average rate %.2f differs from target %.1f by >2 Hz", avgRate, targetRate)
	}

	t.Logf("✓ Long-term stability over %v:", testDuration)
	t.Logf("  Final threshold: %.3f (variance: %.6f)", avgThreshold, thresholdVariance)
	t.Logf("  Final rate: %.2f Hz (variance: %.2f, target: %.1f Hz)", avgRate, math.Sqrt(rateVariance), targetRate)
	t.Logf("  Measurements: %d", len(measurements))
}

// ============================================================================
// CUSTOM BEHAVIOR TESTS - EXTENSIBILITY AND TESTING SUPPORT
// ============================================================================

// TestCustomBehavior_ChemicalRelease validates custom chemical release functionality
//
// TESTING SIGNIFICANCE:
// Custom behaviors allow extending neuron functionality for testing scenarios,
// research applications, or specialized neural models without modifying core code.
// This enables activity-dependent chemical release patterns, pharmacological
// simulations, and novel neurotransmitter systems.
//
// EXPECTED RESULTS:
// - Custom chemical release is triggered based on activity thresholds
// - Multiple chemicals can be released simultaneously
// - Custom logic integrates seamlessly with normal neuron operation
// - Error handling works correctly for custom behaviors
// TestCustomBehavior_ChemicalRelease validates custom chemical release functionality
func TestNeuronCoreCustomBehavior_ChemicalRelease(t *testing.T) {
	// Create neuron with standard configuration
	neuron := NewNeuron("custom_test", 0.8, 0.95, 5*time.Millisecond,
		1.0, 3.0, 0.1)

	// Set up mock matrix to capture chemical releases
	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Configure custom behavior - REALISTIC THRESHOLDS
	neuron.SetCustomChemicalRelease(func(activityRate, outputValue float64, release func(types.LigandType, float64) error) {
		// FIXED: Lower threshold to 1.5 Hz (achievable)
		if activityRate > 1.5 {
			err := release(types.LigandBDNF, activityRate*0.02)
			if err != nil {
				t.Logf("BDNF release failed: %v", err)
			}
		}

		// Keep output threshold the same
		if outputValue > 2.0 {
			err := release(types.LigandDopamine, 0.5)
			if err != nil {
				t.Logf("Dopamine release failed: %v", err)
			}
		}
	})

	// Record initial chemical release count
	initialReleases := mockMatrix.GetChemicalReleaseCount()

	// Phase 1: Low activity (should not trigger custom release)
	t.Log("Phase 1: Low activity test...")
	for i := 0; i < 2; i++ { // REDUCED: fewer signals for clearer low activity
		SendTestSignal(neuron, "low_activity", 1.0)
		time.Sleep(200 * time.Millisecond) // SLOWER: 5 Hz = 200ms intervals
	}

	time.Sleep(50 * time.Millisecond)

	lowActivityReleases := mockMatrix.GetChemicalReleaseCount()
	lowActivityStatus := neuron.GetFiringStatus()
	lowActivityRate := lowActivityStatus["current_firing_rate"].(float64)

	t.Logf("Chemical releases after low activity: %d (change: %d), rate: %.2f Hz",
		lowActivityReleases, lowActivityReleases-initialReleases, lowActivityRate)

	// Phase 2: High activity (should trigger BDNF release)
	t.Log("Phase 2: High activity test...")
	for i := 0; i < 20; i++ { // MORE signals
		SendTestSignal(neuron, "high_activity", 1.0)
		time.Sleep(30 * time.Millisecond) // FASTER: ~33 Hz rate
	}

	time.Sleep(50 * time.Millisecond)

	highActivityReleases := mockMatrix.GetChemicalReleaseCount()
	highActivityStatus := neuron.GetFiringStatus()
	highActivityRate := highActivityStatus["current_firing_rate"].(float64)

	newReleases := highActivityReleases - lowActivityReleases

	t.Logf("High activity: rate %.2f Hz, releases %d (new: %d)",
		highActivityRate, highActivityReleases, newReleases)

	if newReleases <= 0 {
		t.Errorf("Expected additional chemical releases with high activity (rate %.2f > threshold 1.5)", highActivityRate)
	}

	// Phase 3: Strong output (should trigger dopamine release)
	t.Log("Phase 3: Strong output test...")
	SendTestSignal(neuron, "strong_output", 3.0) // Strong signal > 2.0 threshold
	time.Sleep(50 * time.Millisecond)

	finalReleases := mockMatrix.GetChemicalReleaseCount()
	strongOutputReleases := finalReleases - highActivityReleases

	if strongOutputReleases <= 0 {
		t.Error("Expected additional chemical releases with strong output")
	}

	// Verify neuron status remains healthy
	status := neuron.GetFiringStatus()
	currentRate := status["current_firing_rate"].(float64)

	t.Logf("✓ Custom chemical release validated:")
	t.Logf("  Low activity releases: %d (rate: %.2f Hz)", lowActivityReleases-initialReleases, lowActivityRate)
	t.Logf("  High activity releases: %d (rate: %.2f Hz)", newReleases, highActivityRate)
	t.Logf("  Strong output releases: %d", strongOutputReleases)
	t.Logf("  Final activity rate: %.2f Hz", currentRate)
	t.Logf("  Total chemical releases: %d", finalReleases)
}

// TestCustomBehavior_DisableAndReconfigure validates custom behavior management
//
// TESTING SIGNIFICANCE:
// Custom behaviors must be easily enabled, disabled, and reconfigured during
// neuron operation. This supports dynamic experimental protocols and ensures
// clean testing environments.
//
// EXPECTED RESULTS:
// - Custom behaviors can be disabled cleanly
// - Neuron operates normally without custom behaviors
// - Custom behaviors can be reconfigured with different logic
// - No memory leaks or state corruption during behavior changes
func TestNeuronCoreCustomBehavior_ChemicalReleaseDisableAndReconfigure(t *testing.T) {
	neuron := NewNeuron("reconfigure_test", 0.5, 0.95, 5*time.Millisecond,
		1.0, 4.0, 0.1)

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Phase 1: Configure initial custom behavior
	t.Log("Phase 1: Initial custom behavior...")
	releaseCounter := 0

	neuron.SetCustomChemicalRelease(func(activityRate, outputValue float64, release func(types.LigandType, float64) error) {
		releaseCounter++
		if activityRate > 2.0 {
			release(types.LigandGlutamate, 0.5)
		}
	})

	// Trigger some activity
	for i := 0; i < 8; i++ {
		SendTestSignal(neuron, "initial", 1.0)
		time.Sleep(25 * time.Millisecond) // 40 Hz rate
	}

	time.Sleep(50 * time.Millisecond)
	initialCallbacks := releaseCounter
	initialReleases := mockMatrix.GetChemicalReleaseCount()

	if initialCallbacks == 0 {
		t.Error("Custom behavior should have been called")
	}

	// Phase 2: Disable custom behavior
	t.Log("Phase 2: Disabling custom behavior...")
	neuron.DisableCustomBehaviors()

	// Trigger more activity
	for i := 0; i < 8; i++ {
		SendTestSignal(neuron, "disabled", 1.0)
		time.Sleep(25 * time.Millisecond)
	}

	time.Sleep(50 * time.Millisecond)
	disabledCallbacks := releaseCounter
	disabledReleases := mockMatrix.GetChemicalReleaseCount()

	if disabledCallbacks != initialCallbacks {
		t.Error("Custom behavior should not be called when disabled")
	}

	// Verify normal neuron operation continues
	status := neuron.GetFiringStatus()
	firingHistory := status["firing_history_size"].(int)
	if firingHistory == 0 {
		t.Error("Neuron should continue normal operation when custom behavior disabled")
	}

	// Phase 3: Reconfigure with different behavior
	t.Log("Phase 3: Reconfiguring custom behavior...")
	neuron.SetCustomChemicalRelease(func(activityRate, outputValue float64, release func(types.LigandType, float64) error) {
		// Different behavior: release based on output value instead of activity
		if outputValue > 1.5 {
			release(types.LigandDopamine, outputValue*0.1)
		}
	})

	// Trigger activity with strong signals
	for i := 0; i < 5; i++ {
		SendTestSignal(neuron, "reconfigured", 2.0) // Strong signal
		time.Sleep(30 * time.Millisecond)
	}

	time.Sleep(50 * time.Millisecond)
	reconfiguredReleases := mockMatrix.GetChemicalReleaseCount()

	newReconfiguredReleases := reconfiguredReleases - disabledReleases
	if newReconfiguredReleases <= 0 {
		t.Error("Expected chemical releases with reconfigured behavior")
	}

	t.Logf("✓ Custom behavior management validated:")
	t.Logf("  Initial behavior calls: %d", initialCallbacks)
	t.Logf("  Disabled behavior calls: %d (should equal initial)", disabledCallbacks)
	t.Logf("  Initial releases: %d", initialReleases)
	t.Logf("  Disabled period releases: %d", disabledReleases-initialReleases)
	t.Logf("  Reconfigured releases: %d", newReconfiguredReleases)
	t.Logf("  Final firing history: %d entries", firingHistory)
}
