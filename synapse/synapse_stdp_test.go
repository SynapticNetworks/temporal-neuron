package synapse

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types" // Assuming types package is imported correctly
)

// / TestSTDPLearningCurve tests the classic STDP learning curve across a full
// range of timing differences, validating the shape matches biological expectations.
func TestSynapseSTDP_LearningCurve(t *testing.T) {
	preNeuron := NewMockNeuron("curve_pre")
	postNeuron := NewMockNeuron("curve_post")

	// Standard STDP parameters
	stdpConfig := types.PlasticityConfig{
		Enabled:        true,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     100 * time.Millisecond,
		MinWeight:      0.001,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.2,
	}

	pruningConfig := CreateDefaultPruningConfig()

	// Sample the STDP curve at multiple points
	timingPoints := []int{-100, -80, -60, -40, -30, -20, -10, -5, 0, 5, 10, 20, 30, 40, 60, 80, 100}
	weightChanges := make([]float64, len(timingPoints))

	t.Log("=== STDP LEARNING CURVE TEST ===")
	t.Log("Timing (ms) | Weight Change | Expected Shape")
	t.Log("-----------------------------------------")

	for i, timingMs := range timingPoints {
		// Create fresh synapse for each measurement
		synapse := NewBasicSynapse(
			fmt.Sprintf("curve_test_%d", i),
			preNeuron, postNeuron,
			stdpConfig, pruningConfig, 1.0, 0,
		)

		// Apply plasticity with this timing
		deltaT := time.Duration(timingMs) * time.Millisecond
		adjustment := types.PlasticityAdjustment{DeltaT: deltaT}

		// Measure weight change
		weightBefore := synapse.GetWeight()
		synapse.ApplyPlasticity(adjustment)
		weightAfter := synapse.GetWeight()

		weightChanges[i] = weightAfter - weightBefore

		// Determine expected behavior
		expectedShape := "Neutral"
		if timingMs < 0 && timingMs > -80 {
			expectedShape = "LTP (strengthening)"
		} else if timingMs > 0 && timingMs < 80 {
			expectedShape = "LTD (weakening)"
		}

		t.Logf("%+6d ms | %+10.6f | %s", timingMs, weightChanges[i], expectedShape)
	}

	// Verify the curve has the right shape

	// 1. Verify LTP region (negative timing)
	hasLTP := false
	for i, timingMs := range timingPoints {
		if timingMs < 0 && timingMs > -80 && weightChanges[i] > 0 {
			hasLTP = true
			break
		}
	}

	// 2. Verify LTD region (positive timing)
	hasLTD := false
	for i, timingMs := range timingPoints {
		if timingMs > 0 && timingMs < 80 && weightChanges[i] < 0 {
			hasLTD = true
			break
		}
	}

	// 3. Verify strongest effects near zero
	maxLTP := 0.0
	maxLTPindex := 0
	for i, timingMs := range timingPoints {
		if timingMs < 0 && weightChanges[i] > maxLTP {
			maxLTP = weightChanges[i]
			maxLTPindex = i
		}
	}

	maxLTD := 0.0
	maxLTDindex := 0
	for i, timingMs := range timingPoints {
		if timingMs > 0 && weightChanges[i] < maxLTD {
			maxLTD = weightChanges[i]
			maxLTDindex = i
		}
	}

	// Verify strongest effects are near zero timing
	if timingPoints[maxLTPindex] < -50 {
		t.Errorf("Maximum LTP effect should be near zero, was at %d ms", timingPoints[maxLTPindex])
	}

	if timingPoints[maxLTDindex] > 50 {
		t.Errorf("Maximum LTD effect should be near zero, was at %d ms", timingPoints[maxLTDindex])
	}

	// Log results
	t.Log("\n=== STDP CURVE VALIDATION ===")
	if hasLTP {
		t.Log("✓ LTP region verified (pre-before-post produces strengthening)")
	} else {
		t.Error("✗ No LTP detected in pre-before-post region")
	}

	if hasLTD {
		t.Log("✓ LTD region verified (post-before-pre produces weakening)")
	} else {
		t.Error("✗ No LTD detected in post-before-pre region")
	}

	t.Logf("✓ Maximum LTP at %d ms: %+.6f", timingPoints[maxLTPindex], maxLTP)
	t.Logf("✓ Maximum LTD at %d ms: %+.6f", timingPoints[maxLTDindex], maxLTD)

	// Verify asymmetry ratio
	observedRatio := math.Abs(maxLTD) / maxLTP
	t.Logf("Observed LTD/LTP ratio: %.2f (expected ~%.1f)",
		observedRatio, stdpConfig.AsymmetryRatio)

	if math.Abs(observedRatio-stdpConfig.AsymmetryRatio) > 0.5 {
		t.Errorf("Asymmetry ratio incorrect: expected ~%.1f, got %.2f",
			stdpConfig.AsymmetryRatio, observedRatio)
	}
}

// TestSTDPLearningAccumulation tests how repeated STDP events accumulate
// over time to produce gradual, stable learning.
func TestSynapseSTDP_LearningAccumulation(t *testing.T) {
	preNeuron := NewMockNeuron("accum_pre")
	postNeuron := NewMockNeuron("accum_post")

	stdpConfig := types.PlasticityConfig{
		Enabled:        true,
		LearningRate:   0.005, // Smaller learning rate for stability
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     100 * time.Millisecond,
		MinWeight:      0.001,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.2,
	}

	pruningConfig := CreateDefaultPruningConfig()

	// Create synapse for repeated STDP application
	synapse := NewBasicSynapse("accum_test", preNeuron, postNeuron,
		stdpConfig, pruningConfig, 1.0, 0)

	// Define test parameters
	repetitions := []int{1, 5, 10, 20, 50, 100}
	deltaT := -10 * time.Millisecond // Strong LTP timing

	t.Log("=== STDP LEARNING ACCUMULATION TEST ===")
	t.Log("Repetitions | Weight | Total Change | Avg Change per Event")
	t.Log("--------------------------------------------------")

	initialWeight := synapse.GetWeight()
	t.Logf("%11d | %.4f | %+.6f | %+.6f",
		0, initialWeight, 0.0, 0.0)

	// Apply repeated STDP events and measure accumulated changes
	cumulativeChange := 0.0

	for _, reps := range repetitions {
		// Reset synapse weight to initial value
		synapse.SetWeight(initialWeight)

		// Apply STDP multiple times
		for i := 0; i < reps; i++ {
			adjustment := types.PlasticityAdjustment{DeltaT: deltaT}
			synapse.ApplyPlasticity(adjustment)
		}

		// Measure accumulated weight change
		finalWeight := synapse.GetWeight()
		totalChange := finalWeight - initialWeight
		avgChangePerEvent := totalChange / float64(reps)

		t.Logf("%11d | %.4f | %+.6f | %+.6f",
			reps, finalWeight, totalChange, avgChangePerEvent)

		// Save for later validation
		if reps == repetitions[len(repetitions)-1] {
			cumulativeChange = totalChange
		}
	}

	// Validation
	t.Log("\n=== ACCUMULATION VALIDATION ===")

	// 1. Verify learning occurred
	if cumulativeChange <= 0 {
		t.Error("❌ No cumulative learning detected after repeated STDP events")
	} else {
		t.Logf("✓ Cumulative learning detected: %+.6f weight change", cumulativeChange)
	}

	// 2. Verify learning is bounded (doesn't grow indefinitely)
	maxReps := repetitions[len(repetitions)-1]
	theoreticalUnboundedChange := float64(maxReps) * stdpConfig.LearningRate

	if cumulativeChange >= theoreticalUnboundedChange {
		t.Errorf("❌ Unbounded growth detected: %+.6f ≥ %+.6f",
			cumulativeChange, theoreticalUnboundedChange)
	} else {
		boundingRatio := cumulativeChange / theoreticalUnboundedChange
		t.Logf("✓ Learning is bounded: %.1f%% of theoretical maximum", boundingRatio*100)
	}

	// 3. Check for reasonable saturation behavior
	// Learning should slow down as weight approaches bounds
	if maxReps >= 50 && cumulativeChange < 0.1 {
		t.Error("❌ Learning saturated too quickly (too little change)")
	} else if cumulativeChange > stdpConfig.MaxWeight-initialWeight {
		t.Error("❌ Learning exceeded maximum weight bound")
	} else {
		t.Log("✓ Learning shows appropriate saturation behavior")
	}
}

// TestSTDPWithTimingJitter tests how STDP behaves with variable/noisy
// timing patterns, which is more realistic in biological systems.
func TestSynapseSTDP_WithTimingJitter(t *testing.T) {
	preNeuron := NewMockNeuron("jitter_pre")
	postNeuron := NewMockNeuron("jitter_post")

	stdpConfig := types.PlasticityConfig{
		Enabled:        true,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     100 * time.Millisecond,
		MinWeight:      0.001,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.2,
	}

	pruningConfig := CreateDefaultPruningConfig()

	t.Log("=== STDP WITH TIMING JITTER TEST ===")
	t.Log("Scenario | Mean Timing | Jitter | Mean Weight Change | Std Dev")
	t.Log("-----------------------------------------------------------")

	// Test scenarios
	scenarios := []struct {
		name        string
		meanTiming  time.Duration
		jitter      time.Duration
		repetitions int
	}{
		{"Precise LTP", -10 * time.Millisecond, 0, 20},
		{"Jittery LTP (small)", -10 * time.Millisecond, 2 * time.Millisecond, 20},
		{"Jittery LTP (medium)", -10 * time.Millisecond, 5 * time.Millisecond, 20},
		{"Jittery LTP (large)", -10 * time.Millisecond, 15 * time.Millisecond, 20},
		{"Precise LTD", 10 * time.Millisecond, 0, 20},
		{"Jittery LTD (small)", 10 * time.Millisecond, 2 * time.Millisecond, 20},
		{"Jittery LTD (medium)", 10 * time.Millisecond, 5 * time.Millisecond, 20},
		{"Jittery LTD (large)", 10 * time.Millisecond, 15 * time.Millisecond, 20},
		{"Around Zero", 0 * time.Millisecond, 5 * time.Millisecond, 20},
	}

	// Helper for random timing with jitter
	randTiming := func(mean time.Duration, jitter time.Duration) time.Duration {
		if jitter == 0 {
			return mean
		}
		// Random value in range [-jitter, +jitter]
		jitterValue := time.Duration(rand.Int63n(int64(2*jitter)) - int64(jitter))
		return mean + jitterValue
	}

	// Calculate mean and standard deviation
	calcStats := func(values []float64) (mean, stdDev float64) {
		sum := 0.0
		for _, v := range values {
			sum += v
		}
		mean = sum / float64(len(values))

		sumSquaredDiff := 0.0
		for _, v := range values {
			diff := v - mean
			sumSquaredDiff += diff * diff
		}
		stdDev = math.Sqrt(sumSquaredDiff / float64(len(values)))
		return mean, stdDev
	}

	for _, scenario := range scenarios {
		weightChanges := make([]float64, scenario.repetitions)

		for i := 0; i < scenario.repetitions; i++ {
			// Create fresh synapse for each trial
			synapse := NewBasicSynapse(
				fmt.Sprintf("jitter_test_%s_%d", scenario.name, i),
				preNeuron, postNeuron,
				stdpConfig, pruningConfig, 1.0, 0,
			)

			// Apply jittered timing
			deltaT := randTiming(scenario.meanTiming, scenario.jitter)
			adjustment := types.PlasticityAdjustment{DeltaT: deltaT}

			// Measure weight change
			weightBefore := synapse.GetWeight()
			synapse.ApplyPlasticity(adjustment)
			weightAfter := synapse.GetWeight()

			weightChanges[i] = weightAfter - weightBefore
		}

		// Calculate statistics
		meanChange, stdDev := calcStats(weightChanges)

		// Log results
		t.Logf("%-14s | %+6.1f ms | %4.1f ms | %+12.6f | %9.6f",
			scenario.name,
			float64(scenario.meanTiming)/float64(time.Millisecond),
			float64(scenario.jitter)/float64(time.Millisecond),
			meanChange, stdDev)
	}

	// No explicit validation needed - this is an observational test
	t.Log("\nNote: Higher jitter should result in higher standard deviation")
	t.Log("and potentially reduced mean effectiveness of learning.")
}

// TestSTDPBoundaryInteractions tests how STDP behaves when weights
// approach minimum and maximum boundaries.
func TestSynapseSTDP_BoundaryInteractions(t *testing.T) {
	preNeuron := NewMockNeuron("boundary_pre")
	postNeuron := NewMockNeuron("boundary_post")

	stdpConfig := types.PlasticityConfig{
		Enabled:        true,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     100 * time.Millisecond,
		MinWeight:      0.01,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.2,
	}

	pruningConfig := CreateDefaultPruningConfig()

	t.Log("=== STDP BOUNDARY INTERACTIONS TEST ===")
	t.Log("Scenario | Initial Weight | Final Weight | Weight Change | Bounded")
	t.Log("--------------------------------------------------------------")

	// Test scenarios
	scenarios := []struct {
		name          string
		initialWeight float64
		deltaT        time.Duration
		repetitions   int
		expectBounded bool
	}{
		{"Near Min + LTD", stdpConfig.MinWeight + 0.005, 10 * time.Millisecond, 5, true},
		{"Near Max + LTP", stdpConfig.MaxWeight - 0.005, -10 * time.Millisecond, 5, true},
		{"Mid-range + LTP", 1.0, -10 * time.Millisecond, 5, false},
		{"Mid-range + LTD", 1.0, 10 * time.Millisecond, 5, false},
		{"Very Low + LTP", 0.05, -10 * time.Millisecond, 5, false},
		{"Very High + LTD", 1.9, 10 * time.Millisecond, 5, false},
	}

	for _, scenario := range scenarios {
		// Create synapse with specific initial weight
		synapse := NewBasicSynapse(
			fmt.Sprintf("boundary_test_%s", scenario.name),
			preNeuron, postNeuron,
			stdpConfig, pruningConfig, scenario.initialWeight, 0,
		)

		initialWeight := synapse.GetWeight()

		// Apply STDP multiple times
		for i := 0; i < scenario.repetitions; i++ {
			adjustment := types.PlasticityAdjustment{DeltaT: scenario.deltaT}
			synapse.ApplyPlasticity(adjustment)
		}

		finalWeight := synapse.GetWeight()
		weightChange := finalWeight - initialWeight

		// Determine if bounded
		bounded := false
		if (scenario.deltaT < 0 && finalWeight >= stdpConfig.MaxWeight-0.0001) || // LTP at max
			(scenario.deltaT > 0 && finalWeight <= stdpConfig.MinWeight+0.0001) { // LTD at min
			bounded = true
		}

		boundedStr := "No"
		if bounded {
			boundedStr = "Yes ✓"
		}

		// Log results
		t.Logf("%-12s | %13.4f | %12.4f | %+12.6f | %s",
			scenario.name, initialWeight, finalWeight, weightChange, boundedStr)

		// Validate expectations
		if scenario.expectBounded && !bounded {
			t.Errorf("❌ Expected %s to be bounded, but it wasn't", scenario.name)
		} else if !scenario.expectBounded && bounded {
			t.Errorf("❌ Did not expect %s to be bounded, but it was", scenario.name)
		}
	}

	// Boundary behavior validation
	t.Log("\n=== BOUNDARY BEHAVIOR VALIDATION ===")
	t.Log("Weights should be properly bounded at minimum and maximum values.")
	t.Log("Learning should slow down as weights approach boundaries.")
}

// TestSTDPDirectionReversal tests advanced STDP behavior where the
// direction of weight change can reverse based on other factors.
// This mimics homeostatic plasticity effects in biological systems.
func TestSynapseSTDP_DirectionReversal(t *testing.T) {
	preNeuron := NewMockNeuron("reversal_pre")
	postNeuron := NewMockNeuron("reversal_post")

	// Standard configuration
	stdpConfig := types.PlasticityConfig{
		Enabled:        true,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     100 * time.Millisecond,
		MinWeight:      0.001,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.2,
	}

	pruningConfig := CreateDefaultPruningConfig()

	// This test requires modification to the ApplyPlasticity method to support
	// context-dependent learning. For testing purposes, we can modify the BasicSynapse
	// implementation or create a special test implementation.

	// Here we'll focus on testing the standard implementation's robustness
	// to alternating LTP and LTD patterns, which is a simpler way to test
	// direction reversal without modifying the implementation.

	t.Log("=== STDP DIRECTION REVERSAL TEST ===")
	t.Log("Scenario | Pattern | Initial Weight | Final Weight | Net Change")
	t.Log("------------------------------------------------------------")

	// Test scenarios
	scenarios := []struct {
		name     string
		pattern  string
		sequence []time.Duration
	}{
		{
			name:     "Alternating LTP-LTD",
			pattern:  "LTP→LTD→LTP→LTD→LTP",
			sequence: []time.Duration{-10 * time.Millisecond, 10 * time.Millisecond, -10 * time.Millisecond, 10 * time.Millisecond, -10 * time.Millisecond},
		},
		{
			name:     "LTD Dominated",
			pattern:  "LTD→LTD→LTP→LTD→LTD",
			sequence: []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, -10 * time.Millisecond, 10 * time.Millisecond, 10 * time.Millisecond},
		},
		{
			name:     "LTP Dominated",
			pattern:  "LTP→LTP→LTD→LTP→LTP",
			sequence: []time.Duration{-10 * time.Millisecond, -10 * time.Millisecond, 10 * time.Millisecond, -10 * time.Millisecond, -10 * time.Millisecond},
		},
		{
			name:     "Varying Intensity",
			pattern:  "Strong LTP→Weak LTD→Medium LTP",
			sequence: []time.Duration{-5 * time.Millisecond, 20 * time.Millisecond, -10 * time.Millisecond},
		},
	}

	for _, scenario := range scenarios {
		// Create synapse for this scenario
		synapse := NewBasicSynapse(
			fmt.Sprintf("reversal_test_%s", scenario.name),
			preNeuron, postNeuron,
			stdpConfig, pruningConfig, 1.0, 0,
		)

		initialWeight := synapse.GetWeight()

		// Apply the sequence of STDP events
		for _, deltaT := range scenario.sequence {
			adjustment := types.PlasticityAdjustment{DeltaT: deltaT}
			synapse.ApplyPlasticity(adjustment)
		}

		finalWeight := synapse.GetWeight()
		netChange := finalWeight - initialWeight

		// Log results
		t.Logf("%-15s | %-20s | %13.4f | %12.4f | %+9.6f",
			scenario.name, scenario.pattern, initialWeight, finalWeight, netChange)
	}

	// No explicit validation beyond observing behavior
	t.Log("\nNote: This test demonstrates the synapse's ability to integrate")
	t.Log("multiple, potentially conflicting plasticity signals over time.")
}
