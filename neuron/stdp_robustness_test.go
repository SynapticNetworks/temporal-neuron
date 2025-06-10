package neuron

import (
	"math"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/synapse"
)

// ============================================================================
// STDP ROBUSTNESS TESTS
// ============================================================================
//
// This file contains tests specifically designed to ensure STDP implementation
// is robust against edge cases, parameter boundaries, and concurrent access,
// while also serving as regression tests to prevent future changes from
// breaking existing STDP behavior.
//
// Test Categories:
// 1. Regression Prevention - Lock in current behavior
// 2. Parameter Boundary Testing - Edge cases and invalid inputs
// 3. Concurrency & Thread Safety - Multi-goroutine robustness
// 4. Memory Management - Large histories and cleanup
// 5. Integration Robustness - STDP with other neuron features
// 6. Extreme Input Handling - Stress testing with unusual inputs
//
// These tests complement the core STDP functionality tests in stdp_test.go
// and integration tests in stdp_integration_test.go
//
// ============================================================================

// mockReceptor is a minimal implementation of the SynapseCompatibleNeuron interface
// used for testing purposes where we only need to intercept received messages.
type mockReceptor struct {
	id        string
	onReceive func(msg synapse.SynapseMessage)
}

func (m *mockReceptor) ID() string {
	return m.id
}

func (m *mockReceptor) Receive(msg synapse.SynapseMessage) {
	if m.onReceive != nil {
		m.onReceive(msg)
	}
}

// TestSTDPRobustnessRegressionBaseline ensures that standard STDP behavior remains
// consistent across code changes. This test locks in the exact weight changes
// for a set of standard timing patterns, serving as a regression detector.
//
// ANY change in these results indicates a modification to STDP behavior that
// needs to be carefully reviewed to ensure it's intentional.
//
// Biological context: These timing patterns represent common spike relationships
// observed in biological neural networks, so maintaining consistent behavior
// is crucial for model validity.
func TestSTDPRobustnessRegressionBaseline(t *testing.T) {
	t.Log("=== STDP REGRESSION BASELINE TEST ===")
	t.Log("This test locks in exact STDP weight changes for standard patterns")
	t.Log("ANY changes in results indicate potential regression - review carefully!")

	// Standard STDP configuration for baseline calculations.
	// The expectedChange values are tied to this specific configuration.
	config := synapse.STDPConfig{
		Enabled:        true,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.0, // Use 1.0 for symmetric LTP/LTD in baseline test
	}

	// Define baseline test cases with expected results.
	// These values are the "golden master" and should not change unless
	// the STDP algorithm is intentionally modified.
	testCases := []struct {
		name           string
		timeDifference time.Duration
		expectedChange float64
		description    string
	}{
		{
			name:           "StrongLTP",
			timeDifference: -5 * time.Millisecond,
			expectedChange: 0.0077880078,
			description:    "Strong LTP at optimal timing",
		},
		{
			name:           "ModerateLTP",
			timeDifference: -15 * time.Millisecond,
			expectedChange: 0.0047236655,
			description:    "Moderate LTP at medium timing",
		},
		{
			name:           "WeakLTP",
			timeDifference: -30 * time.Millisecond,
			expectedChange: 0.0022313016,
			description:    "Weak LTP at longer timing",
		},
		{
			name:           "StrongLTD",
			timeDifference: 5 * time.Millisecond,
			expectedChange: -0.0077880078,
			description:    "Strong LTD at optimal timing",
		},
		{
			name:           "ModerateLTD",
			timeDifference: 15 * time.Millisecond,
			expectedChange: -0.0047236655,
			description:    "Moderate LTD at medium timing",
		},
		{
			name:           "WeakLTD",
			timeDifference: 30 * time.Millisecond,
			expectedChange: -0.0022313016,
			description:    "Weak LTD at longer timing",
		},
		{
			name:           "NoChange_OutsideWindow",
			timeDifference: 100 * time.Millisecond,
			expectedChange: 0.0,
			description:    "No change well outside window",
		},
	}

	// Create mock neurons for the synapses to connect to.
	// Their internal logic doesn't matter for this test.
	mockPre := NewSimpleNeuron("mock_pre", 1, 1, 1, 1)
	mockPost := NewSimpleNeuron("mock_post", 1, 1, 1, 1)

	t.Logf("Testing %d baseline patterns with tight tolerances...", len(testCases))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// For each test case, create a new synapse to ensure isolation.
			initialWeight := 1.0
			syn := synapse.NewBasicSynapse(
				"regression_synapse",
				mockPre,
				mockPost,
				config,
				synapse.CreateDefaultPruningConfig(),
				initialWeight,
				0, // No delay for this test
			)

			// Create the plasticity adjustment based on the test case timing.
			adjustment := synapse.PlasticityAdjustment{DeltaT: tc.timeDifference}

			// Apply the learning rule.
			syn.ApplyPlasticity(adjustment)

			// Calculate the actual change in weight.
			actualChange := syn.GetWeight() - initialWeight

			// Very strict comparison for regression detection.
			tolerance := 1e-9
			if math.Abs(actualChange-tc.expectedChange) > tolerance {
				t.Errorf("REGRESSION DETECTED in %s:", tc.name)
				t.Errorf("  Description: %s", tc.description)
				t.Errorf("  Timing diff: %v", tc.timeDifference)
				t.Errorf("  Expected change: %.8f", tc.expectedChange)
				t.Errorf("  Actual change:   %.8f", actualChange)
				t.Errorf("  Difference:      %.10f", actualChange-tc.expectedChange)
				t.Errorf("This indicates STDP behavior has changed - review changes carefully!")
			} else {
				t.Logf("✓ Baseline maintained for %s (Δt = %v)", tc.name, tc.timeDifference)
			}
		})
	}

	t.Log("✓ STDP regression baseline test completed successfully.")
}

// TestSTDPRobustnessParameterBoundaries tests STDP behavior at parameter boundaries
// and with invalid configurations to ensure robust error handling
//
// Biological motivation: Real synapses have physical limits and neurons
// must handle extreme conditions gracefully without crashing
func TestSTDPRobustnessParameterBoundaries(t *testing.T) {
	t.Log("=== STDP PARAMETER BOUNDARIES TEST ===")
	t.Log("Testing edge cases and boundary conditions for STDP parameters")

	mockPre := NewSimpleNeuron("mock_pre", 1.0, 0.95, 1*time.Millisecond, 1.0)
	mockPost := NewSimpleNeuron("mock_post", 1.0, 0.95, 1*time.Millisecond, 1.0)
	initialWeight := 1.0

	// Test case: Zero Learning Rate
	// Expected Outcome: No change in synaptic weight, as the learning multiplier is zero.
	t.Run("ZeroLearningRate", func(t *testing.T) {
		config := synapse.STDPConfig{
			Enabled:        true,
			LearningRate:   0.0, // The parameter being tested
			TimeConstant:   20 * time.Millisecond,
			WindowSize:     50 * time.Millisecond,
			MinWeight:      0.1,
			MaxWeight:      2.0,
			AsymmetryRatio: 1.0,
		}
		syn := synapse.NewBasicSynapse("syn", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), initialWeight, 0)
		syn.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: 10 * time.Millisecond})

		if math.Abs(syn.GetWeight()-initialWeight) > 1e-9 {
			t.Errorf("Expected no weight change with zero learning rate, got %.6f", syn.GetWeight())
		}
		t.Log("✓ Zero learning rate correctly produces no weight change")
	})

	// Test case: Negative Learning Rate
	// Expected Outcome: Inverted plasticity. Causal pairings (LTP) should weaken the
	// synapse, and anti-causal pairings (LTD) should strengthen it.
	t.Run("NegativeLearningRate", func(t *testing.T) {
		config := synapse.STDPConfig{
			Enabled:        true,
			LearningRate:   -0.01, // The parameter being tested
			TimeConstant:   20 * time.Millisecond,
			WindowSize:     50 * time.Millisecond,
			MinWeight:      0.1,
			MaxWeight:      2.0,
			AsymmetryRatio: 1.0,
		}

		// Test LTP timing (should now be weakening)
		synLTP := synapse.NewBasicSynapse("syn_ltp", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), initialWeight, 0)
		synLTP.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: -10 * time.Millisecond})
		if synLTP.GetWeight() >= initialWeight {
			t.Errorf("Expected weakening for LTP timing with negative learning rate, got %.6f", synLTP.GetWeight())
		}

		// Test LTD timing (should now be strengthening)
		synLTD := synapse.NewBasicSynapse("syn_ltd", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), initialWeight, 0)
		synLTD.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: 10 * time.Millisecond})
		if synLTD.GetWeight() <= initialWeight {
			t.Errorf("Expected strengthening for LTD timing with negative learning rate, got %.6f", synLTD.GetWeight())
		}
		t.Log("✓ Negative learning rate correctly inverts STDP polarity")
	})

	// Test case: Very Small Time Constant
	// Expected Outcome: STDP window should be extremely sharp. Plasticity should
	// decay very rapidly as the spike timing difference increases.
	t.Run("VerySmallTimeConstant", func(t *testing.T) {
		config := synapse.STDPConfig{
			Enabled:        true,
			LearningRate:   0.01,
			TimeConstant:   1 * time.Millisecond, // The parameter being tested
			WindowSize:     50 * time.Millisecond,
			MinWeight:      0.1, // FIX: Added weight boundaries
			MaxWeight:      2.0,
			AsymmetryRatio: 1.0,
		}

		synClose := synapse.NewBasicSynapse("syn_close", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), initialWeight, 0)
		synFar := synapse.NewBasicSynapse("syn_far", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), initialWeight, 0)

		synClose.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: 2 * time.Millisecond})
		synFar.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: 10 * time.Millisecond})

		closeChange := math.Abs(synClose.GetWeight() - initialWeight)
		farChange := math.Abs(synFar.GetWeight() - initialWeight)

		if farChange >= closeChange*0.1 {
			t.Errorf("Expected sharp decay with small time constant. Close: %.6f, Far: %.6f", closeChange, farChange)
		}
		t.Log("✓ Small time constant produces a sharp, narrow STDP window")
	})

	// Test case: Zero Time Constant
	// Expected Outcome: The system should handle division-by-zero gracefully
	// without crashing. The resulting weight change should be a valid number (e.g., 0).
	t.Run("ZeroTimeConstant", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("STDP calculation panicked with zero time constant: %v", r)
			}
		}()

		config := synapse.STDPConfig{
			Enabled:      true,
			LearningRate: 0.01,
			TimeConstant: 0, // The parameter being tested
			WindowSize:   50 * time.Millisecond,
			MinWeight:    0.1, // FIX: Added weight boundaries
			MaxWeight:    2.0,
		}
		syn := synapse.NewBasicSynapse("syn", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), initialWeight, 0)
		syn.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: 10 * time.Millisecond})

		if math.IsNaN(syn.GetWeight()) || math.IsInf(syn.GetWeight(), 0) {
			t.Errorf("Zero time constant produced an invalid weight: %f", syn.GetWeight())
		}
		t.Logf("✓ Zero time constant handled gracefully without panic (weight: %.2f)", syn.GetWeight())
	})

	// Test case: Zero Window Size
	// Expected Outcome: No plasticity should occur, as any timing difference
	// will fall outside a window of zero size.
	t.Run("ZeroWindowSize", func(t *testing.T) {
		config := synapse.STDPConfig{
			Enabled:      true,
			LearningRate: 0.01,
			TimeConstant: 20 * time.Millisecond,
			WindowSize:   0,   // The parameter being tested
			MinWeight:    0.1, // FIX: Added weight boundaries
			MaxWeight:    2.0,
		}
		syn := synapse.NewBasicSynapse("syn", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), initialWeight, 0)
		syn.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: 1 * time.Millisecond})

		if math.Abs(syn.GetWeight()-initialWeight) > 1e-9 {
			t.Errorf("Expected no weight change with zero window size, got %.6f", syn.GetWeight())
		}
		t.Log("✓ Zero window size correctly blocks all plasticity")
	})

	// Test case: Extreme Asymmetry Ratios
	// Expected Outcome: The system should remain numerically stable without producing
	// NaN or infinite values, even with biologically unrealistic ratios.
	t.Run("ExtremeAsymmetryRatios", func(t *testing.T) {
		testRatios := []float64{0.0, 0.001, 1000.0}

		for _, ratio := range testRatios {
			config := synapse.STDPConfig{
				Enabled:        true,
				LearningRate:   0.01,
				TimeConstant:   20 * time.Millisecond,
				WindowSize:     50 * time.Millisecond,
				MinWeight:      0.1, // FIX: Added weight boundaries
				MaxWeight:      2.0,
				AsymmetryRatio: ratio,
			}
			syn := synapse.NewBasicSynapse("syn", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), initialWeight, 0)
			syn.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: 10 * time.Millisecond}) // LTD
			weight := syn.GetWeight()

			if math.IsNaN(weight) || math.IsInf(weight, 0) {
				t.Errorf("NaN or Inf result with asymmetry ratio %.3f", ratio)
			}
		}
		t.Log("✓ Extreme asymmetry ratios handled robustly")
	})

	t.Log("✓ Parameter boundary testing completed successfully.")
}

// TestSTDPRobustnessConcurrentModification tests thread safety of the synapse
// when multiple goroutines are interacting with it simultaneously.
//
// This is crucial for robustness in a concurrent network simulation where
// a synapse might receive a signal to transmit at the same time as it
// receives plasticity feedback from a previously fired post-synaptic neuron.
func TestSTDPRobustnessConcurrentModification(t *testing.T) {
	t.Log("=== STDP CONCURRENT MODIFICATION TEST ===")
	t.Log("Testing thread safety of synapse during concurrent read/write access")

	// Create neurons and a synapse to test
	preNeuron := NewSimpleNeuron("pre", 1, 1, 1, 1)
	postNeuron := NewSimpleNeuron("post", 1, 1, 1, 1)
	config := synapse.CreateDefaultSTDPConfig()
	syn := synapse.NewBasicSynapse("concurrent_syn", preNeuron, postNeuron, config, synapse.CreateDefaultPruningConfig(), 1.0, 0)

	// We don't need the neurons to run for this test, as we will call the synapse methods directly.

	var wg sync.WaitGroup
	const numGoroutines = 200
	const operationsPerGoRoutine = 100

	// Launch many goroutines that will all hammer the same synapse instance
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoRoutine; j++ {
				op := (id + j) % 4
				switch op {
				case 0:
					// Write operation: Apply plasticity
					syn.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: -10 * time.Millisecond})
				case 1:
					// Write operation: Transmit a signal
					syn.Transmit(1.0)
				case 2:
					// Read operation: Get the current weight
					_ = syn.GetWeight()
				case 3:
					// Write operation: Set the weight directly
					syn.SetWeight(1.0 + (float64(id%10) * 0.01))
				}
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// The primary test is that the code completes without panicking due to a race condition.
	// If it completes, the mutexes inside the synapse are working correctly.
	t.Log("✓ Concurrent operations completed without deadlock or panic")
	finalWeight := syn.GetWeight()
	if math.IsNaN(finalWeight) || math.IsInf(finalWeight, 0) {
		t.Errorf("Final weight is not a valid number after concurrent stress test: %f", finalWeight)
	} else {
		t.Logf("✓ Final weight is valid after stress test: %.4f", finalWeight)
	}
	t.Log("✓ Thread safety validation successful")
}

// TestSTDPRobustnessMemoryManagement tests STDP behavior with large spike histories
// and verifies proper memory cleanup to prevent memory leaks
//
// Biological motivation: Real neurons may receive thousands of spikes
// over short periods, so the system must handle large histories efficiently
// TestSTDPRobustnessMemoryManagement tests the resource management of the synapse
// under high-frequency firing conditions.
//
// Biological motivation: Real neurons can fire in high-frequency bursts, sending
// thousands of spikes in short periods. The simulation must handle this load
// gracefully without consuming unbounded memory or leaking resources (like goroutines).
//
// Architectural Context: The current `BasicSynapse` implementation uses `time.AfterFunc`
// for each transmission. This is highly efficient as it avoids creating a persistent
// goroutine for every synapse. However, a very large number of concurrent transmissions
// will create many short-lived timers. This test validates that the Go runtime's
// scheduler and garbage collector can handle this load effectively, ensuring the
// simulation remains stable.
func TestSTDPRobustnessMemoryManagement(t *testing.T) {
	t.Log("=== STDP MEMORY & RESOURCE MANAGEMENT TEST ===")
	t.Log("Testing resource usage under high-frequency burst firing conditions")

	// Create a mock post-synaptic neuron that simply counts received messages.
	// This ensures the test focuses on the synapse's performance without being
	// bottlenecked by the post-synaptic neuron's own processing.
	var messagesReceived int64
	mockPost := &mockReceptor{
		id: "mock_post",
		onReceive: func(msg synapse.SynapseMessage) {
			atomic.AddInt64(&messagesReceived, 1)
		},
	}
	mockPre := NewSimpleNeuron("mock_pre", 1, 1, 1, 1)

	config := synapse.CreateDefaultSTDPConfig()
	// Disable STDP for this test to focus purely on transmission resource management
	config.Enabled = false

	// The synapse that will be stress-tested
	testSynapse := synapse.NewBasicSynapse(
		"stress_test_synapse",
		mockPre,
		mockPost,
		config,
		synapse.CreateDefaultPruningConfig(),
		1.0,                // weight
		5*time.Millisecond, // A small delay to ensure timers are created
	)

	// Test 1: Transmit a large number of spikes in a burst.
	// This simulates a high-frequency input pattern and stress-tests the creation
	// and garbage collection of `time.AfterFunc` timers.
	t.Run("LargeSpikeBurstTransmission", func(t *testing.T) {
		const spikeCount = 10000
		t.Logf("Transmitting %d spikes in a rapid burst...", spikeCount)

		startTime := time.Now()
		for i := 0; i < spikeCount; i++ {
			// Transmit does not block, it schedules a function to run after the delay.
			// This loop will create `spikeCount` pending timers.
			testSynapse.Transmit(1.0)
		}
		duration := time.Since(startTime)
		t.Logf("Finished scheduling %d transmissions in %v", spikeCount, duration)

		// Wait for all messages to be delivered.
		// Add a generous buffer to the expected delay time.
		time.Sleep(testSynapse.GetDelay() + 50*time.Millisecond)

		finalCount := atomic.LoadInt64(&messagesReceived)
		if finalCount != spikeCount {
			t.Errorf("Expected %d messages to be received, but got %d", spikeCount, finalCount)
		} else {
			t.Logf("✓ All %d spikes were successfully transmitted and received", finalCount)
		}
		t.Log("✓ Large spike burst handled without crashing.")
	})

	// Test 2: Sustained high-frequency transmission.
	// This test ensures there are no slow resource leaks over a longer period
	// of sustained high activity.
	t.Run("SustainedHighFrequency", func(t *testing.T) {
		atomic.StoreInt64(&messagesReceived, 0) // Reset counter
		const testDuration = 2 * time.Second
		const frequency = 1000 // Hz (1 spike per millisecond)
		ticker := time.NewTicker(time.Second / frequency)
		defer ticker.Stop()
		stopSignal := time.After(testDuration)

		var totalSent int64
	RunLoop:
		for {
			select {
			case <-ticker.C:
				testSynapse.Transmit(1.0)
				totalSent++
			case <-stopSignal:
				break RunLoop
			}
		}

		// Wait for final messages to arrive
		time.Sleep(testSynapse.GetDelay() + 50*time.Millisecond)

		finalCount := atomic.LoadInt64(&messagesReceived)
		t.Logf("Sustained %d Hz transmission for %v: Sent ~%d, Received %d", frequency, testDuration, totalSent, finalCount)

		if finalCount < totalSent-int64(frequency*0.05) { // Allow for small timing inaccuracies
			t.Errorf("Significant message loss during sustained transmission. Sent ~%d, got %d", totalSent, finalCount)
		} else {
			t.Log("✓ No significant message loss during sustained activity.")
		}
		t.Log("✓ Sustained high-frequency load handled without resource exhaustion.")
	})

	t.Log("✓ STDP memory and resource management tests completed successfully.")
}

// TestSTDPRobustnessExtremeInputs tests STDP behavior with extreme or unusual inputs
// to ensure the system degrades gracefully under stress conditions
//
// This includes testing with edge case timing values, extreme weights,
// and unusual spike patterns that might occur in pathological conditions
func TestSTDPRobustnessExtremeInputs(t *testing.T) {
	t.Log("=== STDP EXTREME INPUTS TEST ===")
	t.Log("Testing STDP robustness with extreme and unusual inputs")

	// Mocks for creating synapses
	mockPre := NewSimpleNeuron("mock_pre", 1, 1, 1, 1)
	mockPost := NewSimpleNeuron("mock_post", 1, 1, 1, 1)

	// Standard config for most tests
	config := synapse.STDPConfig{
		Enabled:        true,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.0,
	}

	// Test 1: Extreme timing differences
	// Expected Outcome: The system should handle very large or small time differences
	// without numerical instability (no NaN or Inf) and correctly apply zero plasticity
	// for timings far outside the STDP window.
	t.Run("ExtremeTimingDifferences", func(t *testing.T) {
		extremeTimings := []time.Duration{
			-24 * time.Hour,  // Very old
			-1 * time.Hour,   // Old
			-time.Nanosecond, // Extremely close
			0,                // Simultaneous
			time.Nanosecond,  // Extremely close
			1 * time.Hour,    // Future
			24 * time.Hour,   // Very future
		}

		for _, timing := range extremeTimings {
			syn := synapse.NewBasicSynapse("syn", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), 1.0, 0)
			syn.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: timing})
			weight := syn.GetWeight()

			// Should not produce NaN or infinite values
			if math.IsNaN(weight) || math.IsInf(weight, 0) {
				t.Errorf("Invalid result for timing %v: %f", timing, weight)
			}
		}
		t.Log("✓ Extreme timing differences handled robustly")
	})

	// Test 2: Rapid fire spike patterns
	// Expected Outcome: A rapid burst of pre-synaptic spikes before a single post-synaptic
	// spike should result in a cumulative, but not pathologically large, weight change.
	// This tests the temporal summation of plasticity.
	t.Run("RapidFireSpikes", func(t *testing.T) {
		syn := synapse.NewBasicSynapse("syn_burst", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), 1.0, 0)
		initialWeight := syn.GetWeight()

		// Generate a burst of 100 pre-synaptic spikes with microsecond spacing
		baseTime := time.Now()
		postSpikeTime := baseTime.Add(10 * time.Millisecond) // Post-spike occurs after the burst

		for i := 0; i < 100; i++ {
			spikeTime := baseTime.Add(time.Duration(i) * time.Microsecond)
			// Apply plasticity for each spike in the burst relative to the single post-spike
			deltaT := spikeTime.Sub(postSpikeTime)
			syn.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: deltaT})
		}

		weightChange := syn.GetWeight() - initialWeight
		t.Logf("Weight change from rapid burst: %.6f", weightChange)

		// Should handle burst without causing extreme weight changes.
		// A single event has a max change of ~0.01. 100 events should not be 100 * 0.01 because of weight boundaries.
		if math.Abs(weightChange) > 1.0 {
			t.Errorf("Excessive weight change from burst: %.6f", weightChange)
		}
		t.Log("✓ Rapid fire spike patterns handled appropriately")
	})

	// Test 3: Weight at boundaries
	// Expected Outcome: When a synapse's weight is already at its minimum or maximum bound,
	// applying plasticity should not push it beyond those limits.
	t.Run("WeightAtBoundaries", func(t *testing.T) {
		// Test at minimum weight
		synMin := synapse.NewBasicSynapse("syn_min", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), config.MinWeight, 0)
		// Try to weaken further (LTD timing) - should be blocked by bounds enforcement in ApplyPlasticity
		synMin.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: 10 * time.Millisecond})
		if synMin.GetWeight() < config.MinWeight {
			t.Errorf("Weight went below minimum: %.6f", synMin.GetWeight())
		}

		// Test at maximum weight
		synMax := synapse.NewBasicSynapse("syn_max", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), config.MaxWeight, 0)
		// Try to strengthen further (LTP timing) - should be blocked by bounds enforcement in ApplyPlasticity
		synMax.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: -10 * time.Millisecond})
		if synMax.GetWeight() > config.MaxWeight {
			t.Errorf("Weight went above maximum: %.6f", synMax.GetWeight())
		}
		t.Log("✓ Weight boundaries properly enforced under extreme conditions")
	})

	t.Log("✓ Extreme inputs testing completed")
}

// TestSTDPRobustnessIntegrationRobustness tests STDP interaction with other neuron
// features under stress conditions to ensure the combined system is robust
//
// This verifies that STDP continues to work correctly when combined with
// homeostatic plasticity, refractory periods, and dynamic network changes
func TestSTDPRobustnessIntegrationRobustness(t *testing.T) {
	t.Log("=== STDP INTEGRATION ROBUSTNESS TEST ===")
	t.Log("Testing STDP robustness in combination with other neuron features")

	// Test 1: STDP during rapid homeostatic changes.
	// Expected Outcome: The system should remain stable. STDP should modify weights while
	// homeostasis adjusts the threshold to maintain the target firing rate, without either
	// process causing pathological oscillations or silencing the neuron.
	t.Run("STDPWithRapidHomeostaticChanges", func(t *testing.T) {
		// Create a neuron with both STDP-enabled synapses and aggressive homeostasis
		source := NewSimpleNeuron("source_homeo", 0.5, 0.95, 4*time.Millisecond, 1.0)
		target := NewNeuron("target_homeo", 1.0, 0.95, 5*time.Millisecond, 1.0, 10.0, 0.5) // High target rate, aggressive homeostasis

		config := synapse.CreateDefaultSTDPConfig()
		syn := synapse.NewBasicSynapse("syn", source, target, config, synapse.CreateDefaultPruningConfig(), 1.0, 0)
		source.AddOutputSynapse("to_target", syn)

		go source.Run()
		defer source.Close()
		go target.Run()
		defer target.Close()

		// Generate activity that will trigger homeostatic adjustment and STDP
		for i := 0; i < 100; i++ {
			preTime := time.Now()
			source.Receive(synapse.SynapseMessage{Value: 1.5, Timestamp: preTime})
			time.Sleep(5 * time.Millisecond) // Causal delay
			postTime := time.Now()
			target.Receive(synapse.SynapseMessage{Value: 1.2, Timestamp: postTime, SourceID: "trigger"})
			syn.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: preTime.Sub(postTime)})
			time.Sleep(15 * time.Millisecond) // High frequency to stress homeostasis
		}

		time.Sleep(200 * time.Millisecond) // Allow homeostasis to adjust

		finalFiringRate := target.GetCurrentFiringRate()
		if finalFiringRate == 0 {
			t.Error("Neuron became completely silent - homeostasis + STDP interaction problem")
		}
		t.Logf("✓ STDP + homeostasis combination remains functional under stress (rate: %.2f Hz)", finalFiringRate)
	})

	// Test 2: STDP during network shutdown.
	// Expected Outcome: The system should handle the shutdown gracefully. Transmissions to a
	// closed neuron channel should be dropped without causing a panic.
	t.Run("STDPDuringNetworkShutdown", func(t *testing.T) {
		source := NewSimpleNeuron("source_shutdown", 0.5, 0.95, 4*time.Millisecond, 1.0)
		target := NewSimpleNeuron("target_shutdown", 0.5, 0.95, 4*time.Millisecond, 1.0)
		syn := synapse.NewBasicSynapse("syn_shutdown", source, target, synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(), 1.0, 0)
		source.AddOutputSynapse("to_target", syn)

		go source.Run()
		go target.Run() // Target must be running to receive signals

		// Generate some initial activity
		source.Receive(synapse.SynapseMessage{Value: 1.5, Timestamp: time.Now()})
		time.Sleep(20 * time.Millisecond)

		// Abruptly shut down the target neuron
		target.Close()
		time.Sleep(10 * time.Millisecond) // Allow channel to close

		// Continue sending from the source. This should not panic.
		source.Receive(synapse.SynapseMessage{Value: 1.5, Timestamp: time.Now()})
		source.Close()
		t.Log("✓ Network shutdown handled gracefully with active STDP")
	})

	// Test 3: STDP with refractory period conflicts.
	// Expected Outcome: The neuron should correctly ignore stimuli arriving during its refractory period.
	// STDP should not be incorrectly applied based on these ignored signals. The neuron should not become silent.
	t.Run("STDPWithRefractoryConflicts", func(t *testing.T) {
		neuron := NewNeuron("refractory_neuron", 1.0, 0.95, 20*time.Millisecond, 1.0, 5.0, 0.1) // Long refractory period
		go neuron.Run()
		defer neuron.Close()

		// Try to create learning patterns during the refractory period
		for i := 0; i < 20; i++ {
			// Fire the neuron to trigger its refractory period
			neuron.Receive(synapse.SynapseMessage{Value: 1.5, Timestamp: time.Now()})
			time.Sleep(2 * time.Millisecond)
			// Send another signal immediately, which should be ignored
			neuron.Receive(synapse.SynapseMessage{Value: 2.0, Timestamp: time.Now()})
			time.Sleep(30 * time.Millisecond) // Let refractory period end
		}

		finalRate := neuron.GetCurrentFiringRate()
		if finalRate == 0 {
			t.Error("Neuron completely stopped firing - refractory + STDP conflict")
		}
		t.Logf("✓ STDP + refractory period conflicts handled appropriately (rate: %.2f Hz)", finalRate)
	})

	// Test 4: Dynamic connection changes during active STDP.
	// Expected Outcome: The neuron should handle the addition and removal of synapses
	// from its output list concurrently with ongoing firing and learning without panicking.
	t.Run("DynamicConnectionsDuringSTDP", func(t *testing.T) {
		neuron := NewNeuron("dynamic_neuron", 1.0, 0.95, 5*time.Millisecond, 1.0, 5.0, 0.1)
		go neuron.Run()
		defer neuron.Close()

		// Start with some outputs
		mockPost1 := &mockReceptor{id: "post1"}
		mockPost2 := &mockReceptor{id: "post2"}
		syn1 := synapse.NewBasicSynapse("syn1", neuron, mockPost1, synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(), 1.0, 0)
		syn2 := synapse.NewBasicSynapse("syn2", neuron, mockPost2, synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(), 1.0, 0)
		neuron.AddOutputSynapse("out1", syn1)
		neuron.AddOutputSynapse("out2", syn2)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Generate activity
			for i := 0; i < 50; i++ {
				neuron.Receive(synapse.SynapseMessage{Value: 1.2, Timestamp: time.Now()})
				time.Sleep(20 * time.Millisecond)
			}
		}()

		// Randomly add/remove outputs during learning
		for i := 0; i < 10; i++ {
			time.Sleep(100 * time.Millisecond)
			if i%2 == 0 {
				neuron.RemoveOutputSynapse("out1")
			} else {
				neuron.AddOutputSynapse("out1", syn1) // Add it back
			}
		}

		wg.Wait()
		outputCount := neuron.GetOutputSynapseCount()
		t.Logf("Final output count: %d", outputCount)
		t.Log("✓ Dynamic connection changes during STDP handled safely")
	})

	t.Log("✓ STDP integration robustness tests completed")
}

// TestSTDPRobustnessGoldenMaster creates a comprehensive golden master test that
// locks in the exact behavior of STDP across multiple scenarios
//
// This is the ultimate regression test - it captures the complete STDP
// behavior profile and will detect ANY changes to the algorithm
func TestSTDPRobustnessGoldenMaster(t *testing.T) {
	t.Log("=== STDP GOLDEN MASTER TEST ===")
	t.Log("Comprehensive STDP behavior capture for regression detection")
	t.Log("This test should NEVER change unless STDP algorithm is intentionally modified")

	// Standard configuration that the golden master values are based on.
	config := synapse.STDPConfig{
		Enabled:        true,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.0,
	}

	// Comprehensive timing test points.
	timingPoints := []time.Duration{
		-60 * time.Millisecond, // Outside window
		-50 * time.Millisecond, // Exactly at window boundary
		-40 * time.Millisecond,
		-30 * time.Millisecond,
		-20 * time.Millisecond,
		-15 * time.Millisecond,
		-10 * time.Millisecond,
		-5 * time.Millisecond,
		-2 * time.Millisecond,
		-1 * time.Millisecond,
		0, // Simultaneous
		1 * time.Millisecond,
		2 * time.Millisecond,
		5 * time.Millisecond,
		10 * time.Millisecond,
		15 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
		40 * time.Millisecond,
		50 * time.Millisecond, // Exactly at window boundary
		60 * time.Millisecond, // Outside window
	}

	// Expected results (golden master values).
	// These values are calibrated to the exact implementation of calculateSTDPWeightChange.
	// They should NEVER change unless the algorithm is intentionally modified.
	expectedResults := map[time.Duration]float64{
		-60 * time.Millisecond: 0.00000000,
		// FIX: Corrected boundary condition. The current implementation uses >=, so at exactly
		// 50ms, the change should be zero.
		-50 * time.Millisecond: 0.00000000,
		-40 * time.Millisecond: 0.00135335,
		-30 * time.Millisecond: 0.00223130,
		-20 * time.Millisecond: 0.00367880,
		-15 * time.Millisecond: 0.00472367,
		-10 * time.Millisecond: 0.00606531,
		-5 * time.Millisecond:  0.00778801,
		-2 * time.Millisecond:  0.00904837,
		-1 * time.Millisecond:  0.00951229,
		0:                      -0.00100000, // Special case for simultaneous
		1 * time.Millisecond:   -0.00951229,
		2 * time.Millisecond:   -0.00904837,
		5 * time.Millisecond:   -0.00778801,
		10 * time.Millisecond:  -0.00606531,
		15 * time.Millisecond:  -0.00472367,
		20 * time.Millisecond:  -0.00367880,
		30 * time.Millisecond:  -0.00223130,
		40 * time.Millisecond:  -0.00135335,
		// FIX: Corrected boundary condition. The current implementation uses >=, so at exactly
		// 50ms, the change should be zero.
		50 * time.Millisecond: 0.00000000,
		60 * time.Millisecond: 0.00000000,
	}

	mockPre := NewSimpleNeuron("mock_pre", 1, 1, 1, 1)
	mockPost := NewSimpleNeuron("mock_post", 1, 1, 1, 1)
	t.Logf("Testing %d timing points for golden master validation", len(timingPoints))

	allPassed := true
	tolerance := 1e-8 // Very strict tolerance

	for _, timing := range timingPoints {
		expected, exists := expectedResults[timing]
		if !exists {
			t.Errorf("Missing expected result for timing %v", timing)
			allPassed = false
			continue
		}

		initialWeight := 1.0
		syn := synapse.NewBasicSynapse("syn_golden", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), initialWeight, 0)
		syn.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: timing})
		actualChange := syn.GetWeight() - initialWeight
		diff := math.Abs(actualChange - expected)

		if diff > tolerance {
			t.Errorf("GOLDEN MASTER VIOLATION at Δt=%v:", timing)
			t.Errorf("  Expected: %.8f", expected)
			t.Errorf("  Actual:   %.8f", actualChange)
			t.Errorf("  Diff:     %.10f", diff)
			allPassed = false
		}
	}

	if allPassed {
		t.Log("✓ ALL GOLDEN MASTER VALUES VERIFIED")
		t.Log("✓ STDP algorithm behavior is consistent with baseline")
	} else {
		t.Error("❌ GOLDEN MASTER VIOLATIONS DETECTED")
		t.Error("❌ STDP algorithm behavior has changed from baseline")
		t.Error("❌ Review changes carefully - this may indicate regression")
	}

	// Additional validation: Check STDP curve properties
	t.Run("STDPCurveProperties", func(t *testing.T) {
		synLTP := synapse.NewBasicSynapse("syn_ltp", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), 1.0, 0)
		synLTD := synapse.NewBasicSynapse("syn_ltd", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), 1.0, 0)

		synLTP.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: -1 * time.Millisecond})
		synLTD.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: 1 * time.Millisecond})

		ltpPeak := synLTP.GetWeight() - 1.0
		ltdPeak := synLTD.GetWeight() - 1.0

		if ltpPeak <= 0 {
			t.Errorf("LTP peak should be positive, got %.6f", ltpPeak)
		}
		if ltdPeak >= 0 {
			t.Errorf("LTD peak should be negative, got %.6f", ltdPeak)
		}

		// Verify symmetry (for asymmetry ratio = 1.0)
		if math.Abs(math.Abs(ltpPeak)-math.Abs(ltdPeak)) > tolerance {
			t.Errorf("LTP/LTD peaks should be symmetric with ratio=1.0: LTP=%.6f, LTD=%.6f", ltpPeak, ltdPeak)
		}

		t.Log("✓ STDP curve properties validated")
	})

	t.Log("✓ Golden master test completed")
}

// TestSTDPRobustnessReproducibility ensures that STDP calculations are deterministic
// and produce identical results across multiple runs with the same inputs
//
// This is crucial for scientific reproducibility and debugging
func TestSTDPRobustnessReproducibility(t *testing.T) {
	t.Log("=== STDP REPRODUCIBILITY TEST ===")
	t.Log("Ensuring deterministic STDP behavior across multiple runs")

	// Standard configuration that will be used for all runs.
	config := synapse.STDPConfig{
		Enabled:        true,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.5, // Use a non-1.0 ratio to test asymmetry determinism
	}

	// A set of standard timing points to test.
	testTimings := []time.Duration{
		-25 * time.Millisecond,
		-10 * time.Millisecond,
		-1 * time.Millisecond,
		0,
		1 * time.Millisecond,
		10 * time.Millisecond,
		25 * time.Millisecond,
	}

	// Create mock neurons for the synapses to connect to.
	mockPre := NewSimpleNeuron("mock_pre", 1, 1, 1, 1)
	mockPost := NewSimpleNeuron("mock_post", 1, 1, 1, 1)

	// Run the same calculations multiple times to check for consistency.
	const numRuns = 100
	results := make(map[time.Duration][]float64)

	for run := 0; run < numRuns; run++ {
		for _, timing := range testTimings {
			initialWeight := 1.0
			syn := synapse.NewBasicSynapse("repro_syn", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), initialWeight, 0)
			syn.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: timing})
			change := syn.GetWeight() - initialWeight
			results[timing] = append(results[timing], change)
		}
	}

	// Verify all runs for each timing point produced the exact same result.
	allReproducible := true
	for timing, runResults := range results {
		firstResult := runResults[0]
		for i, result := range runResults {
			// Use a very small tolerance to account for potential floating point representation nuances,
			// although for this algorithm, they should be identical.
			if math.Abs(result-firstResult) > 1e-12 {
				t.Errorf("REPRODUCIBILITY FAILURE at Δt=%v:", timing)
				t.Errorf("  Run 0:    %.12f", firstResult)
				t.Errorf("  Run %d:    %.12f", i, result)
				t.Errorf("  Difference: %.15f", result-firstResult)
				allReproducible = false
				break // No need to check other runs for this timing
			}
		}
	}

	if allReproducible {
		t.Logf("✓ Perfect reproducibility across %d runs for all tested timings", numRuns)
		t.Log("✓ STDP calculations are deterministic")
	} else {
		t.Error("❌ STDP calculations are not reproducible")
		t.Error("❌ This indicates non-deterministic behavior in the learning algorithm")
	}

	t.Log("✓ Reproducibility test completed")
}

// TestSTDPRobustnessNumericalStability tests STDP calculations for numerical stability
// with edge cases that might cause floating-point precision issues
//
// This ensures the implementation is robust against numerical edge cases
// that could cause gradual drift or instability in long-running simulations
func TestSTDPRobustnessNumericalStability(t *testing.T) {
	t.Log("=== STDP NUMERICAL STABILITY TEST ===")
	t.Log("Testing floating-point precision and numerical stability")

	mockPre := NewSimpleNeuron("mock_pre", 1, 1, 1, 1)
	mockPost := NewSimpleNeuron("mock_post", 1, 1, 1, 1)
	config := synapse.STDPConfig{
		Enabled:        true,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.0,
	}

	// Test 1: Very small time differences
	// Expected Outcome: The system should handle extremely small (nanosecond-scale)
	// time differences without numerical errors, producing valid, non-zero weight changes
	// that maintain the correct sign for LTP and LTD.
	t.Run("VerySmallTimeDifferences", func(t *testing.T) {
		smallTimings := []time.Duration{
			1 * time.Nanosecond,
			10 * time.Nanosecond,
			100 * time.Nanosecond,
			1 * time.Microsecond,
		}

		for _, timing := range smallTimings {
			// Test LTP
			synLTP := synapse.NewBasicSynapse("syn_ltp", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), 1.0, 0)
			synLTP.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: -timing})
			ltpWeight := synLTP.GetWeight()

			if math.IsNaN(ltpWeight) || math.IsInf(ltpWeight, 0) {
				t.Errorf("Invalid LTP result for timing %v: %f", timing, ltpWeight)
			}
			if ltpWeight <= 1.0 {
				t.Errorf("LTP should be positive for timing %v: change=%.12f", timing, ltpWeight-1.0)
			}

			// Test LTD
			synLTD := synapse.NewBasicSynapse("syn_ltd", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), 1.0, 0)
			synLTD.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: timing})
			ltdWeight := synLTD.GetWeight()

			if math.IsNaN(ltdWeight) || math.IsInf(ltdWeight, 0) {
				t.Errorf("Invalid LTD result for timing %v: %f", timing, ltdWeight)
			}
			if ltdWeight >= 1.0 {
				t.Errorf("LTD should be negative for timing %v: change=%.12f", timing, ltdWeight-1.0)
			}
		}
		t.Log("✓ Small time differences handled with numerical stability")
	})

	// Test 2: Accumulated precision over many operations
	// Expected Outcome: Applying thousands of very small weight changes should not
	// lead to significant numerical drift or cause the final weight to become an
	// invalid or extreme number. The total change should be reasonable.
	t.Run("AccumulatedPrecision", func(t *testing.T) {
		config.LearningRate = 0.0001 // Use a very small learning rate
		syn := synapse.NewBasicSynapse("syn_accum", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), 1.0, 0)
		initialWeight := syn.GetWeight()

		// Apply 10,000 small LTP changes
		adjustment := synapse.PlasticityAdjustment{DeltaT: -5 * time.Millisecond}
		for i := 0; i < 10000; i++ {
			syn.ApplyPlasticity(adjustment)
			weight := syn.GetWeight()
			if math.IsNaN(weight) || math.IsInf(weight, 0) {
				t.Errorf("Numerical instability after %d iterations: weight=%f", i, weight)
				break
			}
		}

		finalWeight := syn.GetWeight()
		t.Logf("Weight after 10k iterations: %.9f", finalWeight)

		if finalWeight > config.MaxWeight || finalWeight < config.MinWeight {
			t.Errorf("Weight drifted to extreme value outside of bounds: %.9f", finalWeight)
		}
		if math.Abs(finalWeight-initialWeight) < 0.01 {
			t.Errorf("Expected more significant weight change after 10k operations, got %.9f", finalWeight-initialWeight)
		}
		t.Log("✓ Accumulated operations maintain numerical stability")
	})

	// Test 3: Boundary precision
	// Expected Outcome: The STDP rule should be precise at the boundaries of the time window.
	// Plasticity should occur just inside the window and be zero exactly at or just outside the window.
	t.Run("BoundaryPrecision", func(t *testing.T) {
		boundaryTimings := []struct {
			name     string
			timing   time.Duration
			changeGT float64 // Expected change greater than
			changeLT float64 // Expected change less than
		}{
			{"JustInsideLTP", config.WindowSize - time.Nanosecond, 0, 1e-5},
			{"AtBoundary", config.WindowSize, -1e-9, 1e-9},                    // Should be zero
			{"JustOutside", config.WindowSize + time.Nanosecond, -1e-9, 1e-9}, // Should be zero
		}

		for _, tc := range boundaryTimings {
			// Test LTP boundary
			synLTP := synapse.NewBasicSynapse("syn_b_ltp", mockPre, mockPost, config, synapse.CreateDefaultPruningConfig(), 1.0, 0)
			synLTP.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: -tc.timing})
			changeLTP := synLTP.GetWeight() - 1.0

			if !(changeLTP > tc.changeGT && changeLTP < tc.changeLT) {
				t.Errorf("LTP Boundary fail for %s (Δt=%v): got %.12f, expected between %.9f and %.9f", tc.name, -tc.timing, changeLTP, tc.changeGT, tc.changeLT)
			}
		}
		t.Log("✓ Boundary conditions handled with precision")
	})

	t.Log("✓ Numerical stability tests completed")
}
