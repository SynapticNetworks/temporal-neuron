package neuron

import (
	"math"
	"sync"
	"testing"
	"time"
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

	// Standard STDP configuration matching other tests
	config := STDPConfig{
		Enabled:        true,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.0,
	}

	// Define baseline test cases with expected results
	// These values are the "golden master" - they should never change unless
	// the STDP algorithm is intentionally modified
	testCases := []struct {
		name           string
		timeDifference time.Duration
		expectedChange float64
		tolerance      float64
		description    string
	}{
		{
			name:           "StrongLTP",
			timeDifference: -5 * time.Millisecond, // Pre before post
			expectedChange: 0.007788,              // Known baseline value
			tolerance:      0.000001,              // Very tight tolerance for regression detection
			description:    "Strong LTP at optimal timing",
		},
		{
			name:           "ModerateLTP",
			timeDifference: -15 * time.Millisecond,
			expectedChange: 0.004724,
			tolerance:      0.000001,
			description:    "Moderate LTP at medium timing",
		},
		{
			name:           "WeakLTP",
			timeDifference: -30 * time.Millisecond,
			expectedChange: 0.002231,
			tolerance:      0.000001,
			description:    "Weak LTP at longer timing",
		},
		{
			name:           "StrongLTD",
			timeDifference: 5 * time.Millisecond, // Post before pre
			expectedChange: -0.007788,
			tolerance:      0.000001,
			description:    "Strong LTD at optimal timing",
		},
		{
			name:           "ModerateLTD",
			timeDifference: 15 * time.Millisecond,
			expectedChange: -0.004724,
			tolerance:      0.000001,
			description:    "Moderate LTD at medium timing",
		},
		{
			name:           "WeakLTD",
			timeDifference: 30 * time.Millisecond,
			expectedChange: -0.002231,
			tolerance:      0.000001,
			description:    "Weak LTD at longer timing",
		},
		{
			name:           "NoChange_BoundaryNegative",
			timeDifference: -50 * time.Millisecond, // At window boundary
			expectedChange: 0.000000,
			tolerance:      0.000001,
			description:    "No change at negative window boundary",
		},
		{
			name:           "NoChange_BoundaryPositive",
			timeDifference: 50 * time.Millisecond, // At window boundary
			expectedChange: 0.000000,
			tolerance:      0.000001,
			description:    "No change at positive window boundary",
		},
		{
			name:           "NoChange_OutsideWindow",
			timeDifference: 100 * time.Millisecond, // Well outside window
			expectedChange: 0.000000,
			tolerance:      0.000001,
			description:    "No change well outside window",
		},
	}

	t.Logf("Testing %d baseline patterns with tight tolerances", len(testCases))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.description)
			t.Logf("Timing difference: %v", tc.timeDifference)
			t.Logf("Expected weight change: %.6f", tc.expectedChange)

			// Calculate actual weight change
			actualChange := calculateSTDPWeightChange(tc.timeDifference, config)

			t.Logf("Actual weight change: %.6f", actualChange)
			t.Logf("Difference from baseline: %.9f", actualChange-tc.expectedChange)

			// Very strict comparison for regression detection
			if math.Abs(actualChange-tc.expectedChange) > tc.tolerance {
				t.Errorf("REGRESSION DETECTED in %s:", tc.name)
				t.Errorf("  Expected: %.6f", tc.expectedChange)
				t.Errorf("  Actual:   %.6f", actualChange)
				t.Errorf("  Diff:     %.9f", actualChange-tc.expectedChange)
				t.Errorf("  Tolerance: %.6f", tc.tolerance)
				t.Errorf("This indicates STDP behavior has changed - review carefully!")
			} else {
				t.Logf("✓ Baseline maintained for %s", tc.description)
			}
		})
	}

	t.Log("✓ STDP regression baseline test completed")
	t.Log("All patterns match expected baseline values")
}

// TestSTDPRobustnessParameterBoundaries tests STDP behavior at parameter boundaries
// and with invalid configurations to ensure robust error handling
//
// Biological motivation: Real synapses have physical limits and neurons
// must handle extreme conditions gracefully without crashing
func TestSTDPRobustnessParameterBoundaries(t *testing.T) {
	t.Log("=== STDP PARAMETER BOUNDARIES TEST ===")
	t.Log("Testing edge cases and boundary conditions for STDP parameters")

	// Test zero learning rate
	t.Run("ZeroLearningRate", func(t *testing.T) {
		config := STDPConfig{
			Enabled:        true,
			LearningRate:   0.0, // Zero learning rate
			TimeConstant:   20 * time.Millisecond,
			WindowSize:     50 * time.Millisecond,
			MinWeight:      0.1,
			MaxWeight:      2.0,
			AsymmetryRatio: 1.0,
		}

		change := calculateSTDPWeightChange(10*time.Millisecond, config)
		if change != 0.0 {
			t.Errorf("Expected no weight change with zero learning rate, got %f", change)
		}
		t.Log("✓ Zero learning rate correctly produces no weight change")
	})

	// Test negative learning rate
	t.Run("NegativeLearningRate", func(t *testing.T) {
		config := STDPConfig{
			Enabled:        true,
			LearningRate:   -0.01, // Negative learning rate
			TimeConstant:   20 * time.Millisecond,
			WindowSize:     50 * time.Millisecond,
			MinWeight:      0.1,
			MaxWeight:      2.0,
			AsymmetryRatio: 1.0,
		}

		// Should invert normal STDP behavior
		ltpChange := calculateSTDPWeightChange(-10*time.Millisecond, config) // Should be negative
		ltdChange := calculateSTDPWeightChange(10*time.Millisecond, config)  // Should be positive

		if ltpChange >= 0 {
			t.Errorf("Expected negative LTP change with negative learning rate, got %f", ltpChange)
		}
		if ltdChange <= 0 {
			t.Errorf("Expected positive LTD change with negative learning rate, got %f", ltdChange)
		}
		t.Log("✓ Negative learning rate correctly inverts STDP polarity")
	})

	// Test very small time constant
	t.Run("VerySmallTimeConstant", func(t *testing.T) {
		config := STDPConfig{
			Enabled:        true,
			LearningRate:   0.01,
			TimeConstant:   1 * time.Millisecond, // Very small time constant
			WindowSize:     50 * time.Millisecond,
			MinWeight:      0.1,
			MaxWeight:      2.0,
			AsymmetryRatio: 1.0,
		}

		// Should have very sharp STDP window
		closeChange := calculateSTDPWeightChange(2*time.Millisecond, config)
		farChange := calculateSTDPWeightChange(10*time.Millisecond, config)

		// Far change should be much smaller due to sharp exponential decay
		if math.Abs(farChange) >= math.Abs(closeChange)*0.1 {
			t.Errorf("Expected sharp decay with small time constant")
			t.Errorf("Close change: %f, Far change: %f", closeChange, farChange)
		}
		t.Log("✓ Small time constant produces sharp STDP window")
	})

	// Test zero time constant (edge case)
	t.Run("ZeroTimeConstant", func(t *testing.T) {
		config := STDPConfig{
			Enabled:        true,
			LearningRate:   0.01,
			TimeConstant:   0, // Zero time constant
			WindowSize:     50 * time.Millisecond,
			MinWeight:      0.1,
			MaxWeight:      2.0,
			AsymmetryRatio: 1.0,
		}

		// Should handle division by zero gracefully
		// Note: This tests implementation robustness
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("STDP calculation panicked with zero time constant: %v", r)
			}
		}()

		change := calculateSTDPWeightChange(10*time.Millisecond, config)
		// Result should be well-defined (likely 0 or inf, but no panic)
		t.Logf("Zero time constant result: %f", change)
		t.Log("✓ Zero time constant handled without panic")
	})

	// Test zero window size
	t.Run("ZeroWindowSize", func(t *testing.T) {
		config := STDPConfig{
			Enabled:        true,
			LearningRate:   0.01,
			TimeConstant:   20 * time.Millisecond,
			WindowSize:     0, // Zero window size
			MinWeight:      0.1,
			MaxWeight:      2.0,
			AsymmetryRatio: 1.0,
		}

		// All timings should be outside window
		change := calculateSTDPWeightChange(1*time.Millisecond, config)
		if change != 0.0 {
			t.Errorf("Expected no change with zero window size, got %f", change)
		}
		t.Log("✓ Zero window size correctly blocks all plasticity")
	})

	// Test extreme asymmetry ratios
	t.Run("ExtremeAsymmetryRatios", func(t *testing.T) {
		testRatios := []float64{0.0, 0.001, 1000.0, math.Inf(1)}

		for _, ratio := range testRatios {
			config := STDPConfig{
				Enabled:        true,
				LearningRate:   0.01,
				TimeConstant:   20 * time.Millisecond,
				WindowSize:     50 * time.Millisecond,
				MinWeight:      0.1,
				MaxWeight:      2.0,
				AsymmetryRatio: ratio,
			}

			ltpChange := calculateSTDPWeightChange(-10*time.Millisecond, config)
			ltdChange := calculateSTDPWeightChange(10*time.Millisecond, config)

			t.Logf("Asymmetry ratio %.3f: LTP=%.6f, LTD=%.6f", ratio, ltpChange, ltdChange)

			// Should not produce NaN or panic
			if math.IsNaN(ltpChange) || math.IsNaN(ltdChange) {
				t.Errorf("NaN result with asymmetry ratio %f", ratio)
			}
		}
		t.Log("✓ Extreme asymmetry ratios handled robustly")
	})

	t.Log("✓ Parameter boundary testing completed")
}

// TestSTDPRobustnessConcurrentModification tests thread safety when STDP parameters
// are modified while learning is actively occurring
//
// This is crucial for robustness since real applications may need to
// adjust learning parameters dynamically during network operation
func TestSTDPRobustnessConcurrentModification(t *testing.T) {
	t.Log("=== STDP CONCURRENT MODIFICATION TEST ===")
	t.Log("Testing thread safety of STDP parameter changes during active learning")

	// Create neurons with STDP enabled
	preNeuron := NewNeuronWithLearning("pre", 1.0, 5.0, 0.01)
	postNeuron := NewNeuronWithLearning("post", 1.0, 5.0, 0.01)

	// Connect them
	postNeuron.AddOutput("connection", preNeuron.GetInputChannel(), 1.0, 0)

	// Start both neurons
	go preNeuron.Run()
	go postNeuron.Run()
	defer preNeuron.Close()
	defer postNeuron.Close()

	// Start continuous learning activity
	var wg sync.WaitGroup
	stopLearning := make(chan bool)

	// Goroutine 1: Continuous spike generation for learning
	wg.Add(1)
	go func() {
		defer wg.Done()
		preInput := preNeuron.GetInput()
		postInput := postNeuron.GetInput()

		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-stopLearning:
				return
			case <-ticker.C:
				// Generate causal spike pattern (pre -> post)
				preInput <- Message{Value: 1.5, Timestamp: time.Now(), SourceID: "test"}
				time.Sleep(10 * time.Millisecond)
				postInput <- Message{Value: 1.5, Timestamp: time.Now(), SourceID: "test"}
			}
		}
	}()

	// Goroutine 2: Concurrent parameter modifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		modifications := 0
		maxModifications := 20

		for modifications < maxModifications {
			time.Sleep(25 * time.Millisecond) // Modify parameters frequently

			// Modify STDP configuration while learning is active
			newConfig := STDPConfig{
				Enabled:        true,
				LearningRate:   0.005 + float64(modifications)*0.001, // Gradually increase
				TimeConstant:   time.Duration(15+modifications) * time.Millisecond,
				WindowSize:     time.Duration(40+modifications*2) * time.Millisecond,
				MinWeight:      0.1,
				MaxWeight:      2.0,
				AsymmetryRatio: 1.0 + float64(modifications)*0.1,
			}

			// Apply new configuration (this tests thread safety)
			preNeuron.stateMutex.Lock()
			preNeuron.stdpConfig = newConfig
			preNeuron.stateMutex.Unlock()

			postNeuron.stateMutex.Lock()
			postNeuron.stdpConfig = newConfig
			postNeuron.stateMutex.Unlock()

			modifications++
		}
	}()

	// Let the concurrent modification test run
	testDuration := 2 * time.Second
	time.Sleep(testDuration)

	// Stop learning activity
	close(stopLearning)
	wg.Wait()

	t.Log("✓ Concurrent STDP parameter modification completed without deadlock or panic")
	t.Log("✓ Thread safety validation successful")
}

// TestSTDPRobustnessMemoryManagement tests STDP behavior with large spike histories
// and verifies proper memory cleanup to prevent memory leaks
//
// Biological motivation: Real neurons may receive thousands of spikes
// over short periods, so the system must handle large histories efficiently
func TestSTDPRobustnessMemoryManagement(t *testing.T) {
	t.Log("=== STDP MEMORY MANAGEMENT TEST ===")
	t.Log("Testing memory efficiency and cleanup with large spike histories")

	// Create output with STDP enabled
	output := &Output{
		channel:          make(chan Message, 1000),
		factor:           1.0,
		delay:            0,
		baseWeight:       1.0,
		minWeight:        0.1,
		maxWeight:        2.0,
		learningRate:     0.01,
		preSpikeTimes:    make([]time.Time, 0),
		stdpEnabled:      true,
		stdpTimeConstant: 20 * time.Millisecond,
		stdpWindowSize:   50 * time.Millisecond,
	}

	config := STDPConfig{
		Enabled:        true,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.0,
	}

	// Test 1: Large number of pre-synaptic spikes
	t.Run("LargePreSpikeHistory", func(t *testing.T) {
		t.Log("Adding 10,000 pre-synaptic spikes")

		startTime := time.Now()
		for i := 0; i < 10000; i++ {
			spikeTime := startTime.Add(time.Duration(i) * time.Millisecond)
			output.recordPreSynapticSpike(spikeTime, config)
		}

		t.Logf("Spike history length after 10k spikes: %d", len(output.preSpikeTimes))

		// Should have been cleaned up to reasonable size
		if len(output.preSpikeTimes) > 200 { // Reasonable limit
			t.Errorf("Spike history not properly cleaned: %d spikes retained", len(output.preSpikeTimes))
		}

		// Verify that recent spikes are still present
		if len(output.preSpikeTimes) == 0 {
			t.Error("All spikes were cleaned - should retain recent ones")
		}

		t.Log("✓ Large pre-spike history managed efficiently")
	})

	// Test 2: Memory cleanup with old spikes
	t.Run("OldSpikeCleanup", func(t *testing.T) {
		// Clear previous history
		output.preSpikeTimes = output.preSpikeTimes[:0]

		// Add very old spikes (should be cleaned up)
		oldTime := time.Now().Add(-1 * time.Hour)
		for i := 0; i < 100; i++ {
			output.recordPreSynapticSpike(oldTime.Add(time.Duration(i)*time.Millisecond), config)
		}

		t.Logf("History length after old spikes: %d", len(output.preSpikeTimes))

		// Add one recent spike
		output.recordPreSynapticSpike(time.Now(), config)

		t.Logf("History length after recent spike: %d", len(output.preSpikeTimes))

		// Should have cleaned up old spikes
		if len(output.preSpikeTimes) > 10 {
			t.Errorf("Old spikes not cleaned up: %d spikes retained", len(output.preSpikeTimes))
		}

		t.Log("✓ Old spike cleanup working correctly")
	})

	// Test 3: Stress test with rapid spike generation
	t.Run("RapidSpikeGeneration", func(t *testing.T) {
		output.preSpikeTimes = output.preSpikeTimes[:0]

		startTime := time.Now()

		// Generate spikes very rapidly
		for i := 0; i < 1000; i++ {
			spikeTime := startTime.Add(time.Duration(i) * time.Microsecond) // Microsecond spacing
			output.recordPreSynapticSpike(spikeTime, config)

			// Every 100 spikes, check memory usage
			if i%100 == 0 {
				if len(output.preSpikeTimes) > 500 {
					t.Errorf("Memory usage growing too large: %d spikes at iteration %d", len(output.preSpikeTimes), i)
					break
				}
			}
		}

		t.Logf("Final history length after rapid generation: %d", len(output.preSpikeTimes))
		t.Log("✓ Rapid spike generation handled efficiently")
	})

	t.Log("✓ STDP memory management tests completed")
}

// TestSTDPRobustnessExtremeInputs tests STDP behavior with extreme or unusual inputs
// to ensure the system degrades gracefully under stress conditions
//
// This includes testing with edge case timing values, extreme weights,
// and unusual spike patterns that might occur in pathological conditions
func TestSTDPRobustnessExtremeInputs(t *testing.T) {
	t.Log("=== STDP EXTREME INPUTS TEST ===")
	t.Log("Testing STDP robustness with extreme and unusual inputs")

	config := STDPConfig{
		Enabled:        true,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.0,
	}

	// Test 1: Extreme timing differences
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
			change := calculateSTDPWeightChange(timing, config)

			// Should not produce NaN or infinite values
			if math.IsNaN(change) || math.IsInf(change, 0) {
				t.Errorf("Invalid result for timing %v: %f", timing, change)
			}

			// Should be reasonable magnitude (not astronomically large)
			if math.Abs(change) > 1000 {
				t.Errorf("Unreasonably large weight change for timing %v: %f", timing, change)
			}

			t.Logf("Timing %v: change = %.6f", timing, change)
		}

		t.Log("✓ Extreme timing differences handled robustly")
	})

	// Test 2: Rapid fire spike patterns
	t.Run("RapidFireSpikes", func(t *testing.T) {
		output := &Output{
			channel:          make(chan Message, 100),
			factor:           1.0,
			baseWeight:       1.0,
			minWeight:        0.1,
			maxWeight:        2.0,
			learningRate:     0.01,
			preSpikeTimes:    make([]time.Time, 0),
			stdpEnabled:      true,
			stdpTimeConstant: 20 * time.Millisecond,
			stdpWindowSize:   50 * time.Millisecond,
		}

		// Generate burst of spikes with microsecond spacing
		baseTime := time.Now()
		for i := 0; i < 100; i++ {
			spikeTime := baseTime.Add(time.Duration(i) * time.Microsecond)
			output.recordPreSynapticSpike(spikeTime, config)
		}

		// Apply STDP with post-spike shortly after burst
		postSpikeTime := baseTime.Add(10 * time.Millisecond)
		initialWeight := output.factor

		output.applySTDPToSynapse(postSpikeTime, config)

		weightChange := output.factor - initialWeight
		t.Logf("Weight change from rapid burst: %.6f", weightChange)

		// Should handle burst without causing extreme weight changes
		if math.Abs(weightChange) > 1.0 {
			t.Errorf("Excessive weight change from burst: %.6f", weightChange)
		}

		t.Log("✓ Rapid fire spike patterns handled appropriately")
	})

	// Test 3: Weight at boundaries
	t.Run("WeightAtBoundaries", func(t *testing.T) {
		// Test at minimum weight
		outputMin := &Output{
			factor:           0.1, // At minimum
			baseWeight:       1.0,
			minWeight:        0.1,
			maxWeight:        2.0,
			learningRate:     0.01,
			preSpikeTimes:    []time.Time{time.Now().Add(-10 * time.Millisecond)},
			stdpEnabled:      true,
			stdpTimeConstant: 20 * time.Millisecond,
			stdpWindowSize:   50 * time.Millisecond,
		}

		// Try to weaken further (should be blocked)
		outputMin.applySTDPToSynapse(time.Now(), config) // LTD timing
		if outputMin.factor < 0.1 {
			t.Errorf("Weight went below minimum: %.6f", outputMin.factor)
		}

		// Test at maximum weight
		outputMax := &Output{
			factor:           2.0, // At maximum
			baseWeight:       1.0,
			minWeight:        0.1,
			maxWeight:        2.0,
			learningRate:     0.01,
			preSpikeTimes:    []time.Time{time.Now().Add(-10 * time.Millisecond)},
			stdpEnabled:      true,
			stdpTimeConstant: 20 * time.Millisecond,
			stdpWindowSize:   50 * time.Millisecond,
		}

		// Try to strengthen further (should be blocked)
		outputMax.applySTDPToSynapse(time.Now().Add(-5*time.Millisecond), config) // LTP timing
		if outputMax.factor > 2.0 {
			t.Errorf("Weight went above maximum: %.6f", outputMax.factor)
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

	// Test 1: STDP during rapid homeostatic changes
	t.Run("STDPWithRapidHomeostaticChanges", func(t *testing.T) {
		// Create neuron with both STDP and aggressive homeostasis
		neuron := NewNeuronWithLearning("test", 1.0, 10.0, 0.02) // High target rate, high learning rate
		neuron.homeostatic.homeostasisStrength = 0.5             // Aggressive homeostasis

		output := make(chan Message, 100)
		neuron.AddOutput("test", output, 1.0, 0)

		go neuron.Run()
		defer neuron.Close()

		input := neuron.GetInput()

		// Generate activity that will trigger homeostatic adjustment
		for i := 0; i < 50; i++ {
			input <- Message{
				Value:     1.5,
				Timestamp: time.Now(),
				SourceID:  "test_source",
			}
			time.Sleep(5 * time.Millisecond)
		}

		// Let homeostasis adjust threshold
		time.Sleep(200 * time.Millisecond)

		// Verify neuron still responds (homeostasis + STDP didn't break anything)
		initialFiringRate := neuron.GetCurrentFiringRate()

		// Continue activity
		for i := 0; i < 20; i++ {
			input <- Message{
				Value:     neuron.GetCurrentThreshold() + 0.1, // Slightly above current threshold
				Timestamp: time.Now(),
				SourceID:  "test_source",
			}
			time.Sleep(10 * time.Millisecond)
		}

		finalFiringRate := neuron.GetCurrentFiringRate()
		thresholdChange := neuron.GetCurrentThreshold() - neuron.GetBaseThreshold()

		t.Logf("Initial firing rate: %.2f Hz", initialFiringRate)
		t.Logf("Final firing rate: %.2f Hz", finalFiringRate)
		t.Logf("Threshold change: %.6f", thresholdChange)

		// Should maintain some firing activity despite homeostatic adjustment
		if finalFiringRate == 0 {
			t.Error("Neuron became completely silent - homeostasis + STDP interaction problem")
		}

		t.Log("✓ STDP + homeostasis combination remains functional under stress")
	})

	// Test 2: STDP during network shutdown
	t.Run("STDPDuringNetworkShutdown", func(t *testing.T) {
		// Create small network with STDP
		source := NewNeuronWithLearning("source", 1.0, 5.0, 0.01)
		target := NewNeuronWithLearning("target", 1.0, 5.0, 0.01)

		// Connect with STDP
		source.AddOutput("to_target", target.GetInputChannel(), 1.0, 0)

		go source.Run()
		go target.Run()

		// Generate learning activity
		sourceInput := source.GetInput()
		for i := 0; i < 10; i++ {
			sourceInput <- Message{
				Value:     1.5,
				Timestamp: time.Now(),
				SourceID:  "external",
			}
			time.Sleep(10 * time.Millisecond)
		}

		// Abruptly shutdown target while source might still be sending
		target.Close()

		// Continue sending to source (target channel now closed)
		// This should not panic or deadlock
		for i := 0; i < 5; i++ {
			sourceInput <- Message{
				Value:     1.5,
				Timestamp: time.Now(),
				SourceID:  "external",
			}
			time.Sleep(5 * time.Millisecond)
		}

		// Shutdown source
		source.Close()

		t.Log("✓ Network shutdown handled gracefully with active STDP")
	})

	// Test 3: STDP with refractory period conflicts
	t.Run("STDPWithRefractoryConflicts", func(t *testing.T) {
		// Create neuron with short refractory period
		neuron := NewNeuronWithLearning("test", 1.0, 5.0, 0.02)
		neuron.refractoryPeriod = 5 * time.Millisecond // Short refractory

		output := make(chan Message, 100)
		neuron.AddOutput("test", output, 1.0, 0)

		go neuron.Run()
		defer neuron.Close()

		input := neuron.GetInput()

		// Try to create learning patterns during refractory period
		for i := 0; i < 20; i++ {
			// Fire neuron
			input <- Message{Value: 1.5, Timestamp: time.Now(), SourceID: "test"}

			// Immediately try to fire again (during refractory)
			time.Sleep(1 * time.Millisecond) // Well within refractory period
			input <- Message{Value: 2.0, Timestamp: time.Now(), SourceID: "test"}

			time.Sleep(10 * time.Millisecond) // Let refractory period end
		}

		// Should handle refractory conflicts without breaking STDP
		finalRate := neuron.GetCurrentFiringRate()
		if finalRate == 0 {
			t.Error("Neuron completely stopped firing - refractory + STDP conflict")
		}

		t.Log("✓ STDP + refractory period conflicts handled appropriately")
	})

	// Test 4: Dynamic connection changes during active STDP
	t.Run("DynamicConnectionsDuringSTDP", func(t *testing.T) {
		neuron := NewNeuronWithLearning("test", 1.0, 5.0, 0.01)

		go neuron.Run()
		defer neuron.Close()

		input := neuron.GetInput()

		// Start with some outputs
		output1 := make(chan Message, 100)
		output2 := make(chan Message, 100)
		neuron.AddOutput("out1", output1, 1.0, 0)
		neuron.AddOutput("out2", output2, 1.0, 0)

		// Generate activity while dynamically changing connections
		for i := 0; i < 50; i++ {
			// Send spike
			input <- Message{
				Value:     1.2,
				Timestamp: time.Now(),
				SourceID:  "dynamic_test",
			}

			// Randomly add/remove outputs during learning
			if i%5 == 0 {
				if i%10 == 0 {
					// Remove output
					neuron.RemoveOutput("out1")
				} else {
					// Add new output
					newOutput := make(chan Message, 100)
					neuron.AddOutput("out1", newOutput, 0.8, 0)
				}
			}

			time.Sleep(20 * time.Millisecond)
		}

		// Should complete without panics or deadlocks
		outputCount := neuron.GetOutputCount()
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

	// Standard configuration
	config := STDPConfig{
		Enabled:        true,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.0,
	}

	// Comprehensive timing test points
	timingPoints := []time.Duration{
		-60 * time.Millisecond, // Outside window
		-50 * time.Millisecond, // Window boundary
		-40 * time.Millisecond, // Strong LTP
		-30 * time.Millisecond, // Medium LTP
		-20 * time.Millisecond, // Moderate LTP
		-15 * time.Millisecond, // Good LTP
		-10 * time.Millisecond, // Strong LTP
		-5 * time.Millisecond,  // Peak LTP
		-2 * time.Millisecond,  // Very strong LTP
		-1 * time.Millisecond,  // Optimal LTP
		0,                      // Simultaneous
		1 * time.Millisecond,   // Optimal LTD
		2 * time.Millisecond,   // Very strong LTD
		5 * time.Millisecond,   // Peak LTD
		10 * time.Millisecond,  // Strong LTD
		15 * time.Millisecond,  // Good LTD
		20 * time.Millisecond,  // Moderate LTD
		30 * time.Millisecond,  // Medium LTD
		40 * time.Millisecond,  // Weak LTD
		50 * time.Millisecond,  // Window boundary
		60 * time.Millisecond,  // Outside window
	}

	// Expected results (golden master values)
	// These should NEVER change unless algorithm is intentionally modified
	expectedResults := map[time.Duration]float64{
		-60 * time.Millisecond: 0.000000,
		-50 * time.Millisecond: 0.000000,
		-40 * time.Millisecond: 0.001353,
		-30 * time.Millisecond: 0.002231,
		-20 * time.Millisecond: 0.003679,
		-15 * time.Millisecond: 0.004724,
		-10 * time.Millisecond: 0.006065,
		-5 * time.Millisecond:  0.007788,
		-2 * time.Millisecond:  0.009048,
		-1 * time.Millisecond:  0.009512,
		0:                      -0.010000, // Special case
		1 * time.Millisecond:   -0.009512,
		2 * time.Millisecond:   -0.009048,
		5 * time.Millisecond:   -0.007788,
		10 * time.Millisecond:  -0.006065,
		15 * time.Millisecond:  -0.004724,
		20 * time.Millisecond:  -0.003679,
		30 * time.Millisecond:  -0.002231,
		40 * time.Millisecond:  -0.001353,
		50 * time.Millisecond:  0.000000,
		60 * time.Millisecond:  0.000000,
	}

	t.Logf("Testing %d timing points for golden master validation", len(timingPoints))

	allPassed := true
	tolerance := 0.000001 // Very strict tolerance

	for _, timing := range timingPoints {
		expected, exists := expectedResults[timing]
		if !exists {
			t.Errorf("Missing expected result for timing %v", timing)
			allPassed = false
			continue
		}

		actual := calculateSTDPWeightChange(timing, config)
		diff := math.Abs(actual - expected)

		if diff > tolerance {
			t.Errorf("GOLDEN MASTER VIOLATION at Δt=%v:", timing)
			t.Errorf("  Expected: %.6f", expected)
			t.Errorf("  Actual:   %.6f", actual)
			t.Errorf("  Diff:     %.9f", diff)
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
		// Verify LTP/LTD asymmetry
		ltpPeak := calculateSTDPWeightChange(-1*time.Millisecond, config)
		ltdPeak := calculateSTDPWeightChange(1*time.Millisecond, config)

		if ltpPeak <= 0 {
			t.Errorf("LTP peak should be positive, got %.6f", ltpPeak)
		}
		if ltdPeak >= 0 {
			t.Errorf("LTD peak should be negative, got %.6f", ltdPeak)
		}

		// Verify exponential decay
		close := calculateSTDPWeightChange(-5*time.Millisecond, config)
		far := calculateSTDPWeightChange(-20*time.Millisecond, config)

		if math.Abs(far) >= math.Abs(close) {
			t.Errorf("STDP should decay with distance: close=%.6f, far=%.6f", close, far)
		}

		// Verify symmetry (for asymmetry ratio = 1.0)
		if math.Abs(ltpPeak) != math.Abs(ltdPeak) {
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

	config := STDPConfig{
		Enabled:        true,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.5,
	}

	testTimings := []time.Duration{
		-25 * time.Millisecond,
		-10 * time.Millisecond,
		-1 * time.Millisecond,
		0,
		1 * time.Millisecond,
		10 * time.Millisecond,
		25 * time.Millisecond,
	}

	// Run the same calculations multiple times
	numRuns := 100
	results := make(map[time.Duration][]float64)

	for run := 0; run < numRuns; run++ {
		for _, timing := range testTimings {
			result := calculateSTDPWeightChange(timing, config)
			results[timing] = append(results[timing], result)
		}
	}

	// Verify all runs produced identical results
	allReproducible := true
	for timing, runResults := range results {
		firstResult := runResults[0]

		for i, result := range runResults {
			if result != firstResult {
				t.Errorf("REPRODUCIBILITY FAILURE at Δt=%v:", timing)
				t.Errorf("  Run 0: %.9f", firstResult)
				t.Errorf("  Run %d: %.9f", i, result)
				t.Errorf("  Difference: %.12f", result-firstResult)
				allReproducible = false
				break
			}
		}
	}

	if allReproducible {
		t.Logf("✓ Perfect reproducibility across %d runs", numRuns)
		t.Log("✓ STDP calculations are deterministic")
	} else {
		t.Error("❌ STDP calculations are not reproducible")
		t.Error("❌ This indicates non-deterministic behavior")
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

	config := STDPConfig{
		Enabled:        true,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.0,
	}

	// Test 1: Very small time differences
	t.Run("VerySmallTimeDifferences", func(t *testing.T) {
		smallTimings := []time.Duration{
			1 * time.Nanosecond,
			10 * time.Nanosecond,
			100 * time.Nanosecond,
			1 * time.Microsecond,
			10 * time.Microsecond,
			100 * time.Microsecond,
		}

		for _, timing := range smallTimings {
			positive := calculateSTDPWeightChange(-timing, config) // LTP
			negative := calculateSTDPWeightChange(timing, config)  // LTD

			// Should not be NaN or infinite
			if math.IsNaN(positive) || math.IsInf(positive, 0) {
				t.Errorf("Invalid LTP result for timing %v: %f", timing, positive)
			}
			if math.IsNaN(negative) || math.IsInf(negative, 0) {
				t.Errorf("Invalid LTD result for timing %v: %f", timing, negative)
			}

			// Should maintain sign correctness
			if positive <= 0 {
				t.Errorf("LTP should be positive for timing %v: %f", timing, positive)
			}
			if negative >= 0 {
				t.Errorf("LTD should be negative for timing %v: %f", timing, negative)
			}

			t.Logf("Timing %v: LTP=%.9f, LTD=%.9f", timing, positive, negative)
		}

		t.Log("✓ Small time differences handled with numerical stability")
	})

	// Test 2: Accumulated precision over many operations
	t.Run("AccumulatedPrecision", func(t *testing.T) {
		output := &Output{
			factor:           1.0,
			baseWeight:       1.0,
			minWeight:        0.1,
			maxWeight:        2.0,
			learningRate:     0.001, // Small learning rate
			preSpikeTimes:    make([]time.Time, 0),
			stdpEnabled:      true,
			stdpTimeConstant: 20 * time.Millisecond,
			stdpWindowSize:   50 * time.Millisecond,
		}

		initialWeight := output.factor

		// Apply many small weight changes
		baseTime := time.Now()
		for i := 0; i < 10000; i++ {
			// Add pre-spike
			output.preSpikeTimes = []time.Time{baseTime.Add(-5 * time.Millisecond)}

			// Apply small STDP change
			output.applySTDPToSynapse(baseTime, config)

			// Check for numerical drift
			if math.IsNaN(output.factor) || math.IsInf(output.factor, 0) {
				t.Errorf("Numerical instability after %d iterations: weight=%f", i, output.factor)
				break
			}
		}

		finalWeight := output.factor
		totalChange := finalWeight - initialWeight

		t.Logf("Weight after 10k iterations: %.9f", finalWeight)
		t.Logf("Total change: %.9f", totalChange)

		// Should not have drifted to extreme values
		if finalWeight < 0.01 || finalWeight > 100 {
			t.Errorf("Weight drifted to extreme value: %.9f", finalWeight)
		}

		t.Log("✓ Accumulated operations maintain numerical stability")
	})

	// Test 3: Boundary precision
	t.Run("BoundaryPrecision", func(t *testing.T) {
		// Test at exact window boundaries
		boundaryTimings := []time.Duration{
			-50*time.Millisecond - time.Nanosecond, // Just outside
			-50 * time.Millisecond,                 // Exactly at boundary
			-50*time.Millisecond + time.Nanosecond, // Just inside
			50*time.Millisecond - time.Nanosecond,  // Just inside
			50 * time.Millisecond,                  // Exactly at boundary
			50*time.Millisecond + time.Nanosecond,  // Just outside
		}

		for _, timing := range boundaryTimings {
			result := calculateSTDPWeightChange(timing, config)

			t.Logf("Boundary timing %v: result=%.9f", timing, result)

			// Results should be well-defined
			if math.IsNaN(result) || math.IsInf(result, 0) {
				t.Errorf("Invalid result at boundary %v: %f", timing, result)
			}
		}

		t.Log("✓ Boundary conditions handled with precision")
	})

	t.Log("✓ Numerical stability tests completed")
}
