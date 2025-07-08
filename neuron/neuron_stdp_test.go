package neuron

import (
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

/*
=================================================================================
COMPREHENSIVE STDP NEURON TEST SUITE
=================================================================================

This file contains tests for Spike-Timing-Dependent Plasticity (STDP):
1. Basic STDP functionality tests
2. Concurrent STDP access tests
3. Edge case and error handling tests
4. Deadlock prevention tests
5. Performance tests

All tests use the prefix TestSTDPNeuronBasic_ for easy filtering.
Tests are organized into logical categories.

=================================================================================
*/

// ============================================================================
// BASIC FUNCTIONALITY TESTS
// ============================================================================

// TestSTDPNeuronBasic_BasicFunctionality verifies that STDP causes synaptic weights to change
// in the expected direction based on the timing difference between pre and post-synaptic spikes
func TestSTDPNeuronBasic_BasicFunctionality(t *testing.T) {
	// Create a neuron with STDP enabled
	neuron := NewNeuron(
		"stdp_basic_test",
		1.0,                // threshold
		0.95,               // decay rate
		5*time.Millisecond, // refractory period
		2.0,                // fire factor
		3.0,                // target firing rate
		0.2,                // homeostasis strength
	)

	// Enable STDP with a moderate learning rate
	neuron.EnableSTDPFeedback(
		10*time.Millisecond, // feedback delay
		0.1,                 // learning rate
	)

	// Create a mock matrix to track STDP calls
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Start the neuron
	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	t.Log("Testing causal spike timing (pre before post)")

	// Create a test synapse that fired recently (causal - should strengthen)
	causalSynapse := types.SynapseInfo{
		ID:               "causal_synapse",
		SourceID:         "source",
		TargetID:         neuron.ID(),
		Weight:           0.5,
		LastActivity:     time.Now().Add(-5 * time.Millisecond), // Pre-synaptic spike 5ms ago
		LastTransmission: time.Now().Add(-5 * time.Millisecond), // Same as LastActivity for simplicity
	}

	// Setup our mock to return this synapse when ListSynapses is called
	mockMatrix.SetSynapseList([]types.SynapseInfo{causalSynapse})

	// Simulate a post-synaptic spike now (by calling SendSTDPFeedback)
	// This should strengthen the synapse because pre fired before post
	neuron.SendSTDPFeedback()

	// Check if plasticity was applied in the right direction
	adjustments := mockMatrix.GetPlasticityAdjustments()
	if len(adjustments) == 0 {
		t.Fatal("No plasticity adjustments were made")
	}

	// For causal timing, DeltaT should be negative and weight should increase
	adjustment := adjustments[0]
	t.Logf("Causal timing: DeltaT = %v, LearningRate = %v", adjustment.DeltaT, adjustment.LearningRate)

	if adjustment.DeltaT >= 0 {
		t.Errorf("Expected negative DeltaT for causal timing, got %v", adjustment.DeltaT)
	}

	// Clear adjustments for next test
	mockMatrix.ClearPlasticityAdjustments()

	t.Log("Testing anti-causal spike timing (post before pre)")

	// Create a test synapse that will fire in the future (anti-causal - should weaken)
	antiCausalSynapse := types.SynapseInfo{
		ID:               "anticausal_synapse",
		SourceID:         "source",
		TargetID:         neuron.ID(),
		Weight:           0.5,
		LastActivity:     time.Now().Add(15 * time.Millisecond), // Pre-synaptic spike will happen 15ms later
		LastTransmission: time.Now().Add(15 * time.Millisecond), // Same as LastActivity for simplicity
	}

	// Update our mock to return this synapse
	mockMatrix.SetSynapseList([]types.SynapseInfo{antiCausalSynapse})

	// Simulate a post-synaptic spike now
	// This should weaken the synapse because post fired before pre
	neuron.SendSTDPFeedback()

	// Check if plasticity was applied in the right direction
	adjustments = mockMatrix.GetPlasticityAdjustments()
	if len(adjustments) == 0 {
		t.Fatal("No plasticity adjustments were made for anti-causal timing")
	}

	// For anti-causal timing, DeltaT should be positive
	adjustment = adjustments[0]
	t.Logf("Anti-causal timing: DeltaT = %v, LearningRate = %v", adjustment.DeltaT, adjustment.LearningRate)

	if adjustment.DeltaT <= 0 {
		t.Errorf("Expected positive DeltaT for anti-causal timing, got %v", adjustment.DeltaT)
	}

	t.Log("✓ STDP basic functionality test completed successfully")
}

// TestSTDPNeuronBasic_LearningRateEffects verifies that different learning rates properly affect weight changes
func TestSTDPNeuronBasic_LearningRateEffects(t *testing.T) {
	// Test with multiple learning rates
	learningRates := []float64{0.01, 0.1, 0.5}

	for _, rate := range learningRates {
		t.Logf("Testing with learning rate: %.2f", rate)

		// Create a neuron with this learning rate
		neuron := NewNeuron(
			"stdp_rate_test",
			1.0,
			0.95,
			5*time.Millisecond,
			2.0,
			3.0,
			0.2,
		)

		// Enable STDP with this learning rate
		neuron.EnableSTDPFeedback(10*time.Millisecond, rate)

		// Create mock matrix and track adjustments
		mockMatrix := NewMockMatrix()
		mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
		neuron.SetCallbacks(mockCallbacks)

		// Start the neuron
		neuron.Start()

		// Create a test synapse
		testSynapse := types.SynapseInfo{
			ID:               "test_synapse",
			SourceID:         "source",
			TargetID:         neuron.ID(),
			Weight:           0.5,
			LastActivity:     time.Now().Add(-5 * time.Millisecond),
			LastTransmission: time.Now().Add(-5 * time.Millisecond), // Use LastActivity for simplicity
		}

		// Setup our mock
		mockMatrix.SetSynapseList([]types.SynapseInfo{testSynapse})

		// Trigger STDP
		neuron.SendSTDPFeedback()

		// Check the learning rate in the adjustment
		adjustments := mockMatrix.GetPlasticityAdjustments()
		if len(adjustments) == 0 {
			t.Fatalf("No plasticity adjustments made with learning rate %.2f", rate)
		}

		adjustment := adjustments[0]
		if math.Abs(adjustment.LearningRate-rate) > 0.0001 {
			t.Errorf("Expected learning rate %.2f, got %.2f", rate, adjustment.LearningRate)
		}

		// Clean up
		neuron.Stop()
		mockMatrix.ClearPlasticityAdjustments()
	}

	t.Log("✓ STDP learning rate effects test completed successfully")
}

// TestSTDPNeuronBasic_TimingCurve verifies that the STDP timing curve behaves correctly
// for different pre-post spike timing differences
func TestSTDPNeuronBasic_TimingCurve(t *testing.T) {
	// Create a neuron with STDP
	neuron := NewNeuron(
		"stdp_curve_test",
		1.0,
		0.95,
		5*time.Millisecond,
		2.0,
		3.0,
		0.2,
	)

	neuron.EnableSTDPFeedback(20*time.Millisecond, 0.1)

	// Create mock matrix
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Start the neuron
	neuron.Start()
	defer neuron.Stop()

	// Test various spike timing differences
	timingDifferences := []time.Duration{
		-30 * time.Millisecond, // Pre well before post - potentiation
		-20 * time.Millisecond, // Pre before post - potentiation
		-10 * time.Millisecond, // Pre just before post - max potentiation
		-5 * time.Millisecond,  // Pre just before post - strong potentiation
		0 * time.Millisecond,   // Simultaneous - border case
		5 * time.Millisecond,   // Post just before pre - depression
		10 * time.Millisecond,  // Post before pre - depression
		20 * time.Millisecond,  // Post well before pre - depression
		30 * time.Millisecond,  // Post well before pre - weak depression
	}

	// Map to store results
	results := make(map[time.Duration]float64)

	for _, deltaT := range timingDifferences {
		// Calculate LastActivity time based on deltaT
		// Negative deltaT: pre fired before post (LastActivity in past)
		// Positive deltaT: post fired before pre (LastActivity in future)
		lastActivity := time.Now().Add(deltaT)

		// Create a test synapse with this timing
		testSynapse := types.SynapseInfo{
			ID:               "timing_test_synapse",
			SourceID:         "source",
			TargetID:         neuron.ID(),
			Weight:           0.5,
			LastActivity:     lastActivity,
			LastTransmission: lastActivity, // Use LastActivity for simplicity
		}

		// Setup mock and clear previous adjustments
		mockMatrix.ClearPlasticityAdjustments()
		mockMatrix.SetSynapseList([]types.SynapseInfo{testSynapse})

		// Trigger STDP
		neuron.SendSTDPFeedback()

		// Check adjustment
		adjustments := mockMatrix.GetPlasticityAdjustments()
		if len(adjustments) == 0 {
			t.Fatalf("No plasticity adjustment for deltaT = %v", deltaT)
		}

		// Store the adjustment
		results[deltaT] = float64(adjustments[0].DeltaT.Milliseconds())
	}

	// Verify the STDP curve shape
	// For negative deltaT (pre before post), should be negative values
	// For positive deltaT (post before pre), should be positive values
	for deltaT, resultDeltaT := range results {
		t.Logf("DeltaT = %v ms, Measured = %.2f ms", deltaT.Milliseconds(), resultDeltaT)

		// Check if sign is correct
		if deltaT < 0 && resultDeltaT > 0 {
			t.Errorf("Expected negative DeltaT for pre-before-post timing, got %.2f", resultDeltaT)
		}
		if deltaT > 0 && resultDeltaT < 0 {
			t.Errorf("Expected positive DeltaT for post-before-pre timing, got %.2f", resultDeltaT)
		}
	}

	t.Log("✓ STDP timing curve test completed successfully")
}

// TestSTDPNeuronBasic_DisableEnable tests enabling and disabling STDP
func TestSTDPNeuronBasic_DisableEnable(t *testing.T) {
	// Create a neuron
	neuron := NewNeuron(
		"stdp_toggle_test",
		1.0,
		0.95,
		5*time.Millisecond,
		2.0,
		3.0,
		0.2,
	)

	// Create mock matrix
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Start the neuron
	neuron.Start()
	defer neuron.Stop()

	// Create a test synapse
	testSynapse := types.SynapseInfo{
		ID:               "toggle_test_synapse",
		SourceID:         "source",
		TargetID:         neuron.ID(),
		Weight:           0.5,
		LastActivity:     time.Now().Add(-5 * time.Millisecond),
		LastTransmission: time.Now().Add(-5 * time.Millisecond), // Use LastActivity for simplicity
	}

	mockMatrix.SetSynapseList([]types.SynapseInfo{testSynapse})

	// Test 1: STDP disabled by default
	t.Log("Testing with STDP disabled (default)")
	neuron.SendSTDPFeedback()

	// Should have no adjustments
	adjustments := mockMatrix.GetPlasticityAdjustments()
	if len(adjustments) > 0 {
		t.Errorf("Expected no adjustments with STDP disabled, got %d", len(adjustments))
	}

	// Test 2: Enable STDP
	t.Log("Enabling STDP")
	neuron.EnableSTDPFeedback(5*time.Millisecond, 0.1)

	// Verify STDP is enabled
	if !neuron.IsSTDPFeedbackEnabled() {
		t.Error("STDP should be enabled but IsSTDPFeedbackEnabled() returned false")
	}

	// Clear previous adjustments
	mockMatrix.ClearPlasticityAdjustments()

	// Trigger STDP
	neuron.SendSTDPFeedback()

	// Should have adjustments now
	adjustments = mockMatrix.GetPlasticityAdjustments()
	if len(adjustments) == 0 {
		t.Error("Expected adjustments with STDP enabled, got none")
	}

	// Test 3: Disable STDP
	t.Log("Disabling STDP")
	neuron.DisableSTDPFeedback()

	// Verify STDP is disabled
	if neuron.IsSTDPFeedbackEnabled() {
		t.Error("STDP should be disabled but IsSTDPFeedbackEnabled() returned true")
	}

	// Clear previous adjustments
	mockMatrix.ClearPlasticityAdjustments()

	// Trigger STDP
	neuron.SendSTDPFeedback()

	// Should have no adjustments again
	adjustments = mockMatrix.GetPlasticityAdjustments()
	if len(adjustments) > 0 {
		t.Errorf("Expected no adjustments after disabling STDP, got %d", len(adjustments))
	}

	t.Log("✓ STDP disable/enable test completed successfully")
}

// TestSTDPNeuronBasic_ScheduledFeedback tests the automatic STDP feedback scheduling
func TestSTDPNeuronBasic_ScheduledFeedback(t *testing.T) {
	// Create a neuron with STDP
	neuron := NewNeuron(
		"stdp_schedule_test",
		0.5, // low threshold to ensure firing
		0.95,
		5*time.Millisecond,
		2.0,
		3.0,
		0.2,
	)

	// Enable STDP with short delay
	feedbackDelay := 20 * time.Millisecond
	neuron.EnableSTDPFeedback(feedbackDelay, 0.1)

	// Create mock matrix
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Setup a test synapse
	testSynapse := types.SynapseInfo{
		ID:               "schedule_test_synapse",
		SourceID:         "source",
		TargetID:         neuron.ID(),
		Weight:           0.5,
		LastActivity:     time.Now().Add(-5 * time.Millisecond),
		LastTransmission: time.Now().Add(-5 * time.Millisecond), // Pre-synaptic spike 5ms ago
	}

	mockMatrix.SetSynapseList([]types.SynapseInfo{testSynapse})

	// Start the neuron
	neuron.Start()
	defer neuron.Stop()

	// Trigger firing with strong signal
	testSignal := types.NeuralSignal{
		Value:     2.0, // Strong signal to ensure firing
		Timestamp: time.Now(),
		SourceID:  "test",
		TargetID:  neuron.ID(),
	}

	t.Log("Sending signal to trigger firing")
	neuron.Receive(testSignal)

	// Wait a bit to allow for firing processing
	time.Sleep(10 * time.Millisecond)

	// No STDP adjustments yet (before the delay expires)
	initialAdjustments := mockMatrix.GetPlasticityAdjustments()
	initialCount := len(initialAdjustments)

	// Wait for scheduled STDP to occur
	t.Logf("Waiting for scheduled STDP (delay: %v)", feedbackDelay)
	time.Sleep(feedbackDelay + 50*time.Millisecond) // Add extra time for processing

	// Should have new adjustments now
	finalAdjustments := mockMatrix.GetPlasticityAdjustments()
	finalCount := len(finalAdjustments)

	t.Logf("Adjustments before delay: %d, after delay: %d", initialCount, finalCount)

	if finalCount <= initialCount {
		t.Error("Expected additional STDP adjustments from scheduled feedback")
	}

	t.Log("✓ STDP scheduled feedback test completed successfully")
}

// ============================================================================
// CONCURRENCY AND DEADLOCK PREVENTION TESTS
// ============================================================================

// TestSTDPNeuronBasic_ConcurrentAccess tests that STDP can be triggered concurrently without deadlocks
func TestSTDPNeuronBasic_ConcurrentAccess(t *testing.T) {
	// Create a neuron with STDP
	neuron := NewNeuron(
		"stdp_concurrent_test",
		1.0,
		0.95,
		5*time.Millisecond,
		2.0,
		3.0,
		0.2,
	)

	neuron.EnableSTDPFeedback(5*time.Millisecond, 0.1)

	// Create mock matrix
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Start the neuron
	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Setup some test synapses
	testSynapses := []types.SynapseInfo{
		{
			ID:               "synapse1",
			SourceID:         "source1",
			TargetID:         neuron.ID(),
			Weight:           0.5,
			LastActivity:     time.Now().Add(-5 * time.Millisecond),
			LastTransmission: time.Now().Add(-5 * time.Millisecond), // Use LastActivity for simplicity
		},
		{
			ID:               "synapse2",
			SourceID:         "source2",
			TargetID:         neuron.ID(),
			Weight:           0.5,
			LastActivity:     time.Now().Add(-10 * time.Millisecond),
			LastTransmission: time.Now().Add(-10 * time.Millisecond), // Use LastActivity for simplicity
		},
	}
	mockMatrix.SetSynapseList(testSynapses)

	// Test concurrent STDP calls
	const numGoroutines = 5
	const callsPerGoroutine = 3

	var wg sync.WaitGroup

	// Create a done channel to detect deadlocks
	done := make(chan bool, 1)

	go func() {
		wg.Wait()
		done <- true
	}()

	// Launch multiple goroutines that call SendSTDPFeedback
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Add small delay to stagger goroutines
			time.Sleep(time.Duration(id) * time.Millisecond)

			for j := 0; j < callsPerGoroutine; j++ {
				// Trigger STDP
				neuron.SendSTDPFeedback()

				// Also read activity level (would cause deadlock without fixes)
				_ = neuron.GetActivityLevel()

				time.Sleep(2 * time.Millisecond)
			}
		}(i)
	}

	// Wait with timeout
	select {
	case <-done:
		// Success!
		adjustments := mockMatrix.GetPlasticityAdjustments()
		t.Logf("Concurrent STDP completed successfully with %d adjustments", len(adjustments))
	case <-time.After(5 * time.Second):
		t.Fatal("❌ DEADLOCK: Concurrent STDP test timed out")
	}

	t.Log("✓ STDP concurrent access test completed successfully")
}

// TestSTDPNeuronBasic_STDPandActivity tests STDP feedback combined with activity level reading
func TestSTDPNeuronBasic_STDPandActivity(t *testing.T) {
	// Create a neuron with STDP
	neuron := NewNeuron(
		"stdp_activity_test",
		1.0,
		0.95,
		5*time.Millisecond,
		2.0,
		3.0,
		0.2,
	)

	neuron.EnableSTDPFeedback(5*time.Millisecond, 0.1)

	// Create mock matrix that reads activity level during callbacks
	mockMatrix := NewMockMatrix()

	// Create special callbacks that read activity level
	activityReadingCallbacks := &ActivityReadingCallbacks{
		MockNeuronCallbacks: NewMockNeuronCallbacks(mockMatrix),
		neuron:              neuron,
	}

	neuron.SetCallbacks(activityReadingCallbacks)

	// Start the neuron
	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Setup a test synapse
	testSynapse := types.SynapseInfo{
		ID:               "activity_test_synapse",
		SourceID:         "source",
		TargetID:         neuron.ID(),
		Weight:           0.5,
		LastActivity:     time.Now().Add(-5 * time.Millisecond),
		LastTransmission: time.Now().Add(-5 * time.Millisecond), // Use LastActivity for simplicity
	}

	mockMatrix.SetSynapseList([]types.SynapseInfo{testSynapse})

	// Read activity before STDP
	beforeActivity := neuron.GetActivityLevel()

	// Simulate firing
	testSignal := types.NeuralSignal{
		Value:     2.0,
		Timestamp: time.Now(),
		SourceID:  "test",
		TargetID:  neuron.ID(),
	}

	// Send signal to make neuron fire
	neuron.Receive(testSignal)

	// Wait for processing
	time.Sleep(20 * time.Millisecond)

	// Manually trigger STDP (would normally happen via scheduledSTDPFeedback)
	t.Log("Triggering STDP feedback (should cause activity reads)")
	neuron.SendSTDPFeedback()

	// Read activity after STDP
	afterActivity := neuron.GetActivityLevel()

	t.Logf("Activity before: %.6f, Activity after: %.6f", beforeActivity, afterActivity)
	t.Logf("Activity reads during STDP: %d", activityReadingCallbacks.activityReads)

	// Ensure activity is non-zero after firing
	if afterActivity <= 0 {
		t.Errorf("Expected non-zero activity after firing, got %.6f", afterActivity)
	}

	// Check that activity was read during STDP
	if activityReadingCallbacks.activityReads == 0 {
		t.Error("No activity reads occurred during STDP feedback")
	}

	t.Log("✓ STDP and activity test completed successfully")
}

// TestSTDPNeuronBasic_HighContention tests STDP under high thread contention
func TestSTDPNeuronBasic_HighContention(t *testing.T) {
	// Create a neuron with STDP enabled
	neuron := NewNeuron(
		"high_contention_test",
		1.0,
		0.95,
		5*time.Millisecond,
		2.0,
		3.0,
		0.2,
	)

	neuron.EnableSTDPFeedback(5*time.Millisecond, 0.1)

	// Create mock matrix
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Start the neuron
	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Setup test synapses
	testSynapses := []types.SynapseInfo{
		{
			ID:           "synapse1",
			SourceID:     "source1",
			TargetID:     neuron.ID(),
			Weight:       0.5,
			LastActivity: time.Now().Add(-5 * time.Millisecond),
		},
	}
	mockMatrix.SetSynapseList(testSynapses)

	// Create operations to run concurrently
	operations := []func(){
		// Op 1: Send STDP feedback
		func() { neuron.SendSTDPFeedback() },
		// Op 2: Get activity level
		func() { _ = neuron.GetActivityLevel() },
		// Op 3: Toggle STDP enable/disable
		func() {
			if neuron.IsSTDPFeedbackEnabled() {
				neuron.DisableSTDPFeedback()
			} else {
				neuron.EnableSTDPFeedback(5*time.Millisecond, 0.1)
			}
		},
		// Op 4: Send signal to neuron
		func() {
			neuron.Receive(types.NeuralSignal{
				Value:     1.0,
				Timestamp: time.Now(),
				SourceID:  "test",
				TargetID:  neuron.ID(),
			})
		},
		// Op 5: Get firing status
		func() { _ = neuron.GetFiringStatus() },
	}

	// Run all operations concurrently with high contention
	const numGoroutines = 10
	const operationsPerGoroutine = 20

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Add a timeout to detect deadlocks
	timeout := time.After(5 * time.Second)
	done := make(chan bool, 1)

	// Track any panics
	var panics int32

	// Launch goroutines with high contention
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic in goroutine %d: %v", id, r)
					atomic.AddInt32(&panics, 1)
				}
				wg.Done()
			}()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Choose a random operation
				opIndex := (id + j) % len(operations)
				operations[opIndex]()

				// Small random pause to increase chance of race conditions
				if j%5 == 0 {
					time.Sleep(time.Duration(id%3) * time.Microsecond)
				}
			}
		}(i)
	}

	// Set up a goroutine to signal when all workers are done
	go func() {
		wg.Wait()
		done <- true
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		// Success!
		t.Logf("✓ Successfully handled %d concurrent operations across %d goroutines",
			numGoroutines*operationsPerGoroutine, numGoroutines)

		if atomic.LoadInt32(&panics) > 0 {
			t.Errorf("❌ %d goroutines panicked during concurrent execution", panics)
		}

	case <-timeout:
		// Dump goroutine stacks for debugging
		buf := make([]byte, 1<<20)
		stackLen := runtime.Stack(buf, true)
		t.Fatalf("❌ DEADLOCK: Test timed out after 5 seconds. Goroutine dump:\n%s", buf[:stackLen])
	}

	// Verify system is still functional after intense concurrency
	neuron.EnableSTDPFeedback(5*time.Millisecond, 0.1)
	neuron.SendSTDPFeedback()

	// If we get here without deadlock, the test passes
	t.Log("✓ STDP high contention test completed successfully")
}

// TestSTDPNeuronBasic_DeadlockScenario specifically tests the previous deadlock scenario
func TestSTDPNeuronBasic_DeadlockScenario(t *testing.T) {
	// Create a neuron with STDP
	neuron := NewNeuron(
		"deadlock_test",
		1.0,
		0.95,
		5*time.Millisecond,
		2.0,
		3.0,
		0.2,
	)

	neuron.EnableSTDPFeedback(5*time.Millisecond, 0.1)

	// Create mock matrix
	mockMatrix := NewMockMatrix()

	// Create a custom callback that will try to read activity level
	// during both ListSynapses and ApplyPlasticity
	customCallbacks := &DeadlockTestCallbacks{
		neuron:              neuron,
		originalCallbacks:   NewMockNeuronCallbacks(mockMatrix),
		readActivityInList:  true,                  // Known deadlock trigger
		readActivityInApply: true,                  // Known deadlock trigger
		delay:               10 * time.Millisecond, // Add delay to increase chance of deadlock
	}

	neuron.SetCallbacks(customCallbacks)

	// Start the neuron
	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Setup a test synapse
	testSynapse := types.SynapseInfo{
		ID:               "deadlock_test_synapse",
		SourceID:         "source",
		TargetID:         neuron.ID(),
		Weight:           0.5,
		LastActivity:     time.Now().Add(-5 * time.Millisecond),
		LastTransmission: time.Now().Add(-5 * time.Millisecond), // Use LastActivity for simplicity
	}

	mockMatrix.SetSynapseList([]types.SynapseInfo{testSynapse})

	// Add a timeout channel to detect deadlock
	done := make(chan bool, 1)

	go func() {
		// Trigger STDP with the problematic callbacks
		t.Log("Triggering STDP feedback with callbacks that read activity level - testing for deadlock...")
		neuron.SendSTDPFeedback()
		done <- true
	}()

	// Wait with timeout
	select {
	case <-done:
		t.Log("✓ No deadlock detected - STDP feedback completed successfully")
	case <-time.After(3 * time.Second):
		// Dump goroutine stacks for debugging
		buf := make([]byte, 1<<20)
		stackLen := runtime.Stack(buf, true)
		t.Fatalf("❌ DEADLOCK DETECTED: STDP feedback did not complete within timeout. Goroutine dump:\n%s", buf[:stackLen])
	}
}

// TestSTDPNeuronBasic_MultipleThreadsReadingActivity tests multiple threads reading activity during STDP
func TestSTDPNeuronBasic_MultipleThreadsReadingActivity(t *testing.T) {
	// Create a neuron with STDP
	neuron := NewNeuron(
		"multi_thread_activity_test",
		1.0,
		0.95,
		5*time.Millisecond,
		2.0,
		3.0,
		0.2,
	)

	neuron.EnableSTDPFeedback(5*time.Millisecond, 0.1)

	// Create mock matrix
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Start the neuron
	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Setup test synapses
	testSynapses := []types.SynapseInfo{
		{
			ID:               "synapse1",
			SourceID:         "source1",
			TargetID:         neuron.ID(),
			Weight:           0.5,
			LastActivity:     time.Now().Add(-5 * time.Millisecond),
			LastTransmission: time.Now().Add(-5 * time.Millisecond), // Use LastActivity for simplicity
		},
	}
	mockMatrix.SetSynapseList(testSynapses)

	// Track successful activity reads
	var activityReads int32

	// Launch multiple goroutines that read activity level
	const numReadThreads = 20
	var wg sync.WaitGroup
	wg.Add(numReadThreads + 1) // +1 for STDP thread

	// Start activity reading threads
	for i := 0; i < numReadThreads; i++ {
		go func(id int) {
			defer wg.Done()

			// Read activity level repeatedly while STDP is happening
			for j := 0; j < 10; j++ {
				_ = neuron.GetActivityLevel()
				atomic.AddInt32(&activityReads, 1)
				time.Sleep(time.Duration(id%5) * time.Millisecond)
			}
		}(i)
	}

	// Start STDP thread
	go func() {
		defer wg.Done()
		// Short delay to let reading threads start
		time.Sleep(5 * time.Millisecond)

		// Trigger STDP feedback
		neuron.SendSTDPFeedback()
	}()

	// Wait with timeout
	done := make(chan bool, 1)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Success!
		t.Logf("✓ Multiple threads read activity %d times during STDP without deadlock", atomic.LoadInt32(&activityReads))
	case <-time.After(5 * time.Second):
		t.Fatal("❌ DEADLOCK: Multiple threads reading activity during STDP test timed out")
	}
}

// ============================================================================
// EDGE CASE TESTS
// ============================================================================

// TestSTDPNeuronBasic_EmptySynapseList tests STDP behavior with no synapses
func TestSTDPNeuronBasic_EmptySynapseList(t *testing.T) {
	// Create a neuron with STDP
	neuron := NewNeuron(
		"empty_synapse_test",
		1.0,
		0.95,
		5*time.Millisecond,
		2.0,
		3.0,
		0.2,
	)

	neuron.EnableSTDPFeedback(5*time.Millisecond, 0.1)

	// Create mock matrix
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Start the neuron
	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Set an empty synapse list
	mockMatrix.SetSynapseList([]types.SynapseInfo{})

	// Trigger STDP - should not panic or error
	neuron.SendSTDPFeedback()

	// Should have no adjustments (no synapses)
	adjustments := mockMatrix.GetPlasticityAdjustments()
	if len(adjustments) > 0 {
		t.Errorf("Expected no adjustments with empty synapse list, got %d", len(adjustments))
	}

	t.Log("✓ STDP with empty synapse list completed successfully")
}

// TestSTDPNeuronBasic_NilCallbacks tests STDP behavior with nil callbacks
func TestSTDPNeuronBasic_NilCallbacks(t *testing.T) {
	// Create a neuron with STDP
	neuron := NewNeuron(
		"nil_callbacks_test",
		1.0,
		0.95,
		5*time.Millisecond,
		2.0,
		3.0,
		0.2,
	)

	neuron.EnableSTDPFeedback(5*time.Millisecond, 0.1)

	// Do NOT set any callbacks (nil callbacks)

	// Start the neuron
	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Add a timeout channel to detect deadlock or hang
	done := make(chan bool, 1)

	go func() {
		// Trigger STDP with nil callbacks - should not panic
		neuron.SendSTDPFeedback()
		done <- true
	}()

	// Wait with timeout
	select {
	case <-done:
		t.Log("✓ STDP with nil callbacks completed successfully without panic")
	case <-time.After(1 * time.Second):
		t.Fatal("❌ STDP with nil callbacks test timed out")
	}
}

// TestSTDPNeuronBasic_ExtremeTimings tests STDP with extreme timing differences
func TestSTDPNeuronBasic_ExtremeTimings(t *testing.T) {
	// Create a neuron with STDP
	neuron := NewNeuron(
		"extreme_timing_test",
		1.0,
		0.95,
		5*time.Millisecond,
		2.0,
		3.0,
		0.2,
	)

	neuron.EnableSTDPFeedback(20*time.Millisecond, 0.1)

	// Create mock matrix
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Start the neuron
	neuron.Start()
	defer neuron.Stop()

	// Test extreme timing differences
	extremeTimings := []time.Duration{
		-500 * time.Millisecond, // Very old pre-synaptic spike
		-1 * time.Microsecond,   // Almost simultaneous (pre before post)
		0 * time.Millisecond,    // Exactly simultaneous
		1 * time.Microsecond,    // Almost simultaneous (post before pre)
		500 * time.Millisecond,  // Far future pre-synaptic spike
	}

	for _, deltaT := range extremeTimings {
		t.Logf("Testing extreme timing: %v", deltaT)

		// Calculate LastActivity time based on deltaT
		lastActivity := time.Now().Add(deltaT)

		// Create a test synapse with this timing
		testSynapse := types.SynapseInfo{
			ID:               "extreme_timing_synapse",
			SourceID:         "source",
			TargetID:         neuron.ID(),
			Weight:           0.5,
			LastActivity:     lastActivity,
			LastTransmission: lastActivity, // Use LastActivity for simplicity
		}

		// Setup mock and clear previous adjustments
		mockMatrix.ClearPlasticityAdjustments()
		mockMatrix.SetSynapseList([]types.SynapseInfo{testSynapse})

		// Trigger STDP - should not panic with extreme timings
		neuron.SendSTDPFeedback()

		// Verify adjustment happened
		adjustments := mockMatrix.GetPlasticityAdjustments()
		if len(adjustments) == 0 {
			t.Errorf("No plasticity adjustment for deltaT = %v", deltaT)
			continue
		}

		// Print the adjustment for verification
		adjustment := adjustments[0]
		t.Logf("  DeltaT = %v resulted in plasticity adjustment with DeltaT = %v",
			deltaT, adjustment.DeltaT)
	}

	t.Log("✓ STDP with extreme timings completed successfully")
}

// TestSTDPNeuronBasic_ZeroLearningRate tests STDP with a zero learning rate
func TestSTDPNeuronBasic_ZeroLearningRate(t *testing.T) {
	// Create a neuron with STDP
	neuron := NewNeuron(
		"zero_rate_test",
		1.0,
		0.95,
		5*time.Millisecond,
		2.0,
		3.0,
		0.2,
	)

	// Enable STDP with zero learning rate
	neuron.EnableSTDPFeedback(10*time.Millisecond, 0.0)

	// Create mock matrix
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Start the neuron
	neuron.Start()
	defer neuron.Stop()

	// Create a test synapse
	testSynapse := types.SynapseInfo{
		ID:               "zero_rate_synapse",
		SourceID:         "source",
		TargetID:         neuron.ID(),
		Weight:           0.5,
		LastActivity:     time.Now().Add(-5 * time.Millisecond),
		LastTransmission: time.Now().Add(-5 * time.Millisecond), // Use LastActivity for simplicity
	}

	mockMatrix.SetSynapseList([]types.SynapseInfo{testSynapse})

	// Trigger STDP - should not error with zero learning rate
	neuron.SendSTDPFeedback()

	// Check the learning rate in the adjustment
	adjustments := mockMatrix.GetPlasticityAdjustments()
	if len(adjustments) == 0 {
		t.Fatal("No plasticity adjustments made with zero learning rate")
	}

	adjustment := adjustments[0]
	if adjustment.LearningRate != 0.0 {
		t.Errorf("Expected learning rate 0.0, got %.4f", adjustment.LearningRate)
	}

	t.Log("✓ STDP with zero learning rate completed successfully")
}

// TestSTDPNeuronBasic_ExtremeLearningRate tests STDP with very large learning rate
func TestSTDPNeuronBasic_ExtremeLearningRate(t *testing.T) {
	// Create a neuron with STDP
	neuron := NewNeuron(
		"extreme_rate_test",
		1.0,
		0.95,
		5*time.Millisecond,
		2.0,
		3.0,
		0.2,
	)

	// Enable STDP with extremely large learning rate
	extremeRate := 10.0 // Very large learning rate
	neuron.EnableSTDPFeedback(10*time.Millisecond, extremeRate)

	// Create mock matrix
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Start the neuron
	neuron.Start()
	defer neuron.Stop()

	// Create a test synapse
	testSynapse := types.SynapseInfo{
		ID:               "extreme_rate_synapse",
		SourceID:         "source",
		TargetID:         neuron.ID(),
		Weight:           0.5,
		LastActivity:     time.Now().Add(-5 * time.Millisecond),
		LastTransmission: time.Now().Add(-5 * time.Millisecond), // Use LastActivity for simplicity
	}

	mockMatrix.SetSynapseList([]types.SynapseInfo{testSynapse})

	// Trigger STDP - should not error with extreme learning rate
	neuron.SendSTDPFeedback()

	// Check the learning rate in the adjustment
	adjustments := mockMatrix.GetPlasticityAdjustments()
	if len(adjustments) == 0 {
		t.Fatal("No plasticity adjustments made with extreme learning rate")
	}

	adjustment := adjustments[0]
	if math.Abs(adjustment.LearningRate-extremeRate) > 0.0001 {
		t.Errorf("Expected learning rate %.2f, got %.2f", extremeRate, adjustment.LearningRate)
	}

	t.Log("✓ STDP with extreme learning rate completed successfully")
}

// ============================================================================
// HELPER TYPES FOR TESTING
// ============================================================================

// ActivityReadingCallbacks reads activity level during STDP operations
type ActivityReadingCallbacks struct {
	*MockNeuronCallbacks
	neuron         *Neuron
	activityReads  int
	activityReadMu sync.Mutex
}

// Override ListSynapses to read activity level
func (arc *ActivityReadingCallbacks) ListSynapses(criteria types.SynapseCriteria) []types.SynapseInfo {
	// Read activity level during callback
	_ = arc.neuron.GetActivityLevel()

	// Count the read
	arc.activityReadMu.Lock()
	arc.activityReads++
	arc.activityReadMu.Unlock()

	// Call original implementation
	return arc.MockNeuronCallbacks.ListSynapses(criteria)
}

// Override ApplyPlasticity to also read activity
func (arc *ActivityReadingCallbacks) ApplyPlasticity(synapseID string, adjustment types.PlasticityAdjustment) error {
	// Read activity level during callback
	_ = arc.neuron.GetActivityLevel()

	// Count the read
	arc.activityReadMu.Lock()
	arc.activityReads++
	arc.activityReadMu.Unlock()

	// Call original implementation
	return arc.MockNeuronCallbacks.ApplyPlasticity(synapseID, adjustment)
}

// DeadlockTestCallbacks is a custom implementation of NeuronCallbacks
// that can be configured to read the neuron's activity level during callbacks
type DeadlockTestCallbacks struct {
	neuron              *Neuron
	originalCallbacks   *MockNeuronCallbacks
	readActivityInList  bool
	readActivityInApply bool
	delay               time.Duration
}

// ListSynapses retrieves synapses matching specific criteria
func (dtc *DeadlockTestCallbacks) ListSynapses(criteria types.SynapseCriteria) []types.SynapseInfo {
	if dtc.readActivityInList {
		// Read activity level before or during ListSynapses
		_ = dtc.neuron.GetActivityLevel()

		// Optional delay to simulate work
		if dtc.delay > 0 {
			time.Sleep(dtc.delay)
		}
	}

	// Call original implementation
	return dtc.originalCallbacks.ListSynapses(criteria)
}

// ApplyPlasticity applies a plasticity adjustment to a synapse
func (dtc *DeadlockTestCallbacks) ApplyPlasticity(synapseID string, adjustment types.PlasticityAdjustment) error {
	if dtc.readActivityInApply {
		// Read activity level before or during ApplyPlasticity
		_ = dtc.neuron.GetActivityLevel()

		// Optional delay to simulate work
		if dtc.delay > 0 {
			time.Sleep(dtc.delay)
		}
	}

	// Call original implementation
	return dtc.originalCallbacks.ApplyPlasticity(synapseID, adjustment)
}

// Forward all other methods to the original callbacks
func (dtc *DeadlockTestCallbacks) CreateSynapse(config types.SynapseCreationConfig) (string, error) {
	return dtc.originalCallbacks.CreateSynapse(config)
}

func (dtc *DeadlockTestCallbacks) DeleteSynapse(synapseID string) error {
	return dtc.originalCallbacks.DeleteSynapse(synapseID)
}

func (dtc *DeadlockTestCallbacks) GetSynapse(synapseID string) (component.SynapticProcessor, error) {
	return dtc.originalCallbacks.GetSynapse(synapseID)
}

func (dtc *DeadlockTestCallbacks) SetSynapseWeight(synapseID string, weight float64) error {
	return dtc.originalCallbacks.SetSynapseWeight(synapseID, weight)
}

func (dtc *DeadlockTestCallbacks) GetSynapseWeight(synapseID string) (float64, error) {
	return dtc.originalCallbacks.GetSynapseWeight(synapseID)
}

func (dtc *DeadlockTestCallbacks) GetSpatialDelay(targetID string) time.Duration {
	return dtc.originalCallbacks.GetSpatialDelay(targetID)
}

func (dtc *DeadlockTestCallbacks) FindNearbyComponents(radius float64) []component.ComponentInfo {
	return dtc.originalCallbacks.FindNearbyComponents(radius)
}

func (dtc *DeadlockTestCallbacks) ReleaseChemical(ligandType types.LigandType, concentration float64) error {
	return dtc.originalCallbacks.ReleaseChemical(ligandType, concentration)
}

func (dtc *DeadlockTestCallbacks) SendElectricalSignal(signalType types.SignalType, data interface{}) {
	dtc.originalCallbacks.SendElectricalSignal(signalType, data)
}

func (dtc *DeadlockTestCallbacks) ReportHealth(activityLevel float64, connectionCount int) {
	dtc.originalCallbacks.ReportHealth(activityLevel, connectionCount)
}

func (dtc *DeadlockTestCallbacks) ReportStateChange(oldState, newState types.ComponentState) {
	dtc.originalCallbacks.ReportStateChange(oldState, newState)
}

func (dtc *DeadlockTestCallbacks) GetMatrix() component.ExtracellularMatrix {
	return dtc.originalCallbacks.GetMatrix()
}
