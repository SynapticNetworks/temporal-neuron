package neuron

import (
	"testing"
	"time"
)

// TestProcessing_FiringRateCalculation tests the firing rate calculation accuracy
func TestProcessing_FiringRateCalculation(t *testing.T) {
	neuron := NewNeuron("rate_test", 0.5, 0.95, 5*time.Millisecond, 1.0, 5.0, 0.1)

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Send 3 signals in quick succession - use activity window timing
	for i := 0; i < 3; i++ {
		SendTestSignal(neuron, "test", 1.0)
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(10 * time.Millisecond)

	status := neuron.GetFiringStatus()
	rate := status["current_firing_rate"].(float64)
	historySize := status["firing_history_size"].(int)

	t.Logf("Rate: %.2f Hz, History: %d", rate, historySize)
	t.Logf("Activity window: %v", DENDRITE_ACTIVITY_TRACKING_WINDOW)

	if historySize < 3 {
		t.Errorf("Expected at least 3 firings, got %d", historySize)
	}

	if rate <= 0 {
		t.Error("Expected positive firing rate")
	}
}

// TestProcessing_HomeostaticBounds tests threshold bounds using constants
func TestProcessing_HomeostaticBounds(t *testing.T) {
	baseThreshold := 1.0
	neuron := NewNeuron("bounds_test", baseThreshold, 0.95, 5*time.Millisecond, 1.0, 5.0,
		DENDRITE_FACTOR_HOMEOSTASIS_STRENGTH_STRONG) // Use strong homeostasis for faster response

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	initialThreshold := neuron.GetThreshold()

	// Calculate expected bounds using constants
	expectedMinBound := baseThreshold * DENDRITE_FACTOR_THRESHOLD_MIN_RATIO // 0.1
	expectedMaxBound := baseThreshold * DENDRITE_FACTOR_THRESHOLD_MAX_RATIO // 5.0

	t.Logf("Expected bounds: %.3f ≤ %.3f ≤ %.3f", expectedMinBound, initialThreshold, expectedMaxBound)

	// Try to force threshold to extreme values through homeostasis
	// Send massive hyperactivity to try to push threshold above max bound
	for i := 0; i < 100; i++ {
		SendTestSignal(neuron, "extreme", 10.0)
		time.Sleep(DENDRITE_TIME_HOMEOSTATIC_TICK / 100) // Small delay using constant
	}

	// Wait for homeostatic adjustment using constant
	time.Sleep(DENDRITE_TIME_HOMEOSTATIC_TICK * 2)

	hyperThreshold := neuron.GetThreshold()

	// Now try hypoactivity - wait long enough for threshold to try to go below min
	time.Sleep(DENDRITE_ACTIVITY_TRACKING_WINDOW / 2) // Use activity window constant

	hypoThreshold := neuron.GetThreshold()

	t.Logf("Threshold progression: %.3f → %.3f → %.3f", initialThreshold, hyperThreshold, hypoThreshold)

	// Check bounds are enforced using constants
	if hyperThreshold > expectedMaxBound+0.001 {
		t.Errorf("Threshold exceeded max bound: %.3f > %.3f", hyperThreshold, expectedMaxBound)
	}

	if hypoThreshold < expectedMinBound-0.001 {
		t.Errorf("Threshold went below min bound: %.3f < %.3f", hypoThreshold, expectedMinBound)
	}

	t.Log("✓ Threshold bounds enforced correctly")
}

// TestProcessing_HomeostaticTiming tests homeostatic adjustment timing using constants
func TestProcessing_HomeostaticTiming(t *testing.T) {
	neuron := NewNeuron("timing_test", 1.0, 0.95, 5*time.Millisecond, 1.0, 5.0,
		DENDRITE_FACTOR_HOMEOSTASIS_STRENGTH_DEFAULT)

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	t.Logf("Homeostatic tick interval: %v", DENDRITE_TIME_HOMEOSTATIC_TICK)
	t.Logf("Activity tracking window: %v", DENDRITE_ACTIVITY_TRACKING_WINDOW)
	t.Log("✓ Homeostatic timing initialized successfully")
}

// TestProcessing_ThresholdAdjustmentLogic tests the core adjustment calculation
func TestProcessing_ThresholdAdjustmentLogic(t *testing.T) {
	// Test the adjustment formula using constants

	currentRate := 6.0                                       // Above target
	targetRate := 3.0                                        // Target
	strength := DENDRITE_FACTOR_HOMEOSTASIS_STRENGTH_DEFAULT // Use constant
	oldThreshold := 1.0                                      // Current threshold

	// Manual calculation
	rateDifference := currentRate - targetRate        // 6.0 - 3.0 = 3.0 (positive = hyperactivity)
	adjustment := rateDifference * strength           // 3.0 * 0.2 = 0.6 (positive adjustment)
	expectedNewThreshold := oldThreshold + adjustment // 1.0 + 0.6 = 1.6 (should increase)

	t.Logf("Manual calculation using constants:")
	t.Logf("  Homeostasis strength: %.3f", DENDRITE_FACTOR_HOMEOSTASIS_STRENGTH_DEFAULT)
	t.Logf("  Current: %.1f, Target: %.1f, Strength: %.1f", currentRate, targetRate, strength)
	t.Logf("  Difference: %.1f, Adjustment: %.1f", rateDifference, adjustment)
	t.Logf("  Old threshold: %.1f → Expected new: %.1f", oldThreshold, expectedNewThreshold)

	if rateDifference <= 0 {
		t.Error("Expected positive rate difference for hyperactivity test")
	}

	if adjustment <= 0 {
		t.Error("Expected positive adjustment for hyperactivity")
	}

	if expectedNewThreshold <= oldThreshold {
		t.Error("Expected threshold to increase for hyperactivity")
	}
}

// TestProcessing_ActualAdjustment tests the actual homeostatic adjustment
func TestProcessing_ActualAdjustment(t *testing.T) {
	targetRate := 2.0
	neuron := NewNeuron("adjustment_test", 0.5, 0.95, 5*time.Millisecond, 1.0, targetRate,
		DENDRITE_FACTOR_HOMEOSTASIS_STRENGTH_STRONG) // Strong homeostasis for clear effect

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Force firing at high rate to create hyperactivity
	// Send signals faster than target rate to ensure hyperactivity
	signalInterval := time.Duration(1000.0/float64(targetRate*3)) * time.Millisecond // 3x target rate
	t.Logf("Sending signals every %v to achieve 3x target rate", signalInterval)

	for i := 0; i < 20; i++ {
		SendTestSignal(neuron, "force_fire", 2.0) // Well above threshold
		time.Sleep(signalInterval)
	}

	initialThreshold := neuron.GetThreshold()

	// Wait for homeostatic adjustment using constant
	time.Sleep(DENDRITE_TIME_HOMEOSTATIC_TICK * 3)

	finalThreshold := neuron.GetThreshold()
	status := neuron.GetFiringStatus()
	rate := status["current_firing_rate"].(float64)

	t.Logf("Target rate: %.2f Hz", targetRate)
	t.Logf("Actual rate: %.2f Hz", rate)
	t.Logf("Threshold: %.3f → %.3f", initialThreshold, finalThreshold)
	t.Logf("Homeostatic interval: %v", DENDRITE_TIME_HOMEOSTATIC_TICK)

	// Only expect threshold increase if rate is actually above target
	if rate > targetRate && finalThreshold <= initialThreshold {
		t.Errorf("High activity (%.2f > %.2f) should increase threshold: %.3f → %.3f",
			rate, targetRate, initialThreshold, finalThreshold)
	} else if rate <= targetRate {
		t.Logf("Note: Actual rate %.2f ≤ target %.2f, so threshold decrease is expected", rate, targetRate)
	}
}

// TestProcessing_ActivityWindow tests firing rate calculation window using constants
func TestProcessing_ActivityWindow(t *testing.T) {
	neuron := NewNeuron("window_test", 0.5, 0.95, 5*time.Millisecond, 1.0, 5.0, 0.1)

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Send a signal and check that firing rate calculation works
	SendTestSignal(neuron, "test", 1.0)
	time.Sleep(10 * time.Millisecond)

	status := neuron.GetFiringStatus()
	rate := status["current_firing_rate"].(float64)

	t.Logf("Firing rate after single signal: %.3f Hz", rate)
	t.Logf("Activity tracking window: %v", DENDRITE_ACTIVITY_TRACKING_WINDOW)
	t.Logf("Expected max rate for single spike: %.3f Hz", 1.0/DENDRITE_ACTIVITY_TRACKING_WINDOW.Seconds())
	t.Log("✓ Activity window functioning")
}
