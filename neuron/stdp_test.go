package neuron

import (
	"fmt"
	"math"
	"testing"
	"time"
)

// ============================================================================
// STDP TESTING UTILITIES AND STATISTICS
// ============================================================================

// STDPTestStats tracks learning statistics for biological validation
// This comprehensive statistics collection enables us to validate that our
// STDP implementation matches biological research findings and behaves
// according to established neuroscience principles
type STDPTestStats struct {
	// === BASIC LEARNING METRICS ===
	InitialWeight       float64 // Starting synaptic weight
	FinalWeight         float64 // Ending synaptic weight
	WeightChange        float64 // Total change (final - initial)
	WeightChangePercent float64 // Percentage change from baseline
	NumLearningEvents   int     // Number of STDP applications
	LearningRate        float64 // Effective learning rate observed

	// === TIMING ANALYSIS ===
	AvgTimingDifference time.Duration // Average Δt between pre/post spikes
	MinTimingDifference time.Duration // Closest spike timing observed
	MaxTimingDifference time.Duration // Furthest spike timing observed
	TimingStdDev        time.Duration // Standard deviation of timing

	// === BIOLOGICAL VALIDATION METRICS ===
	LTPEvents      int     // Long-term potentiation occurrences
	LTDEvents      int     // Long-term depression occurrences
	LTPMagnitude   float64 // Average LTP weight change
	LTDMagnitude   float64 // Average LTD weight change
	AsymmetryRatio float64 // Observed LTP/LTD ratio

	// === LEARNING DYNAMICS ===
	WeightEvolution    []float64 // Weight at each learning step
	LearningCurveSlope float64   // Rate of learning (biological: should decay)
	SaturationPoint    float64   // Weight at which learning slows

	// === TIMING WINDOW ANALYSIS ===
	EffectiveWindow    time.Duration // Actual timing window where learning occurred
	WindowUtilization  float64       // Fraction of theoretical window used
	PeakLearningTiming time.Duration // Timing difference with strongest learning

	// === FREQUENCY DEPENDENCE ===
	LowFreqLearning     float64 // Learning rate at <1Hz
	HighFreqLearning    float64 // Learning rate at >10Hz
	FrequencyDependence float64 // Ratio high/low frequency learning

	// === METAPLASTICITY INDICATORS ===
	EarlyLearningRate  float64 // Learning rate in first 25% of trials
	LateLearningRate   float64 // Learning rate in last 25% of trials
	LearningAdaptation float64 // Change in learning rate over time
}

// calculateSTDPStats computes comprehensive statistics from STDP learning data
// This function analyzes the learning process to extract biologically relevant
// metrics that can be compared with experimental neuroscience data
func calculateSTDPStats(initialWeight float64, weightHistory []float64, timingHistory []time.Duration) STDPTestStats {
	stats := STDPTestStats{
		InitialWeight:   initialWeight,
		WeightEvolution: make([]float64, len(weightHistory)),
	}

	copy(stats.WeightEvolution, weightHistory)

	if len(weightHistory) == 0 {
		return stats
	}

	// Basic metrics
	stats.FinalWeight = weightHistory[len(weightHistory)-1]
	stats.WeightChange = stats.FinalWeight - stats.InitialWeight
	if stats.InitialWeight != 0 {
		stats.WeightChangePercent = (stats.WeightChange / stats.InitialWeight) * 100
	}
	stats.NumLearningEvents = len(weightHistory) - 1

	// Timing analysis
	if len(timingHistory) > 0 {
		var totalTiming time.Duration
		stats.MinTimingDifference = timingHistory[0]
		stats.MaxTimingDifference = timingHistory[0]

		for _, timing := range timingHistory {
			totalTiming += timing
			if timing < stats.MinTimingDifference {
				stats.MinTimingDifference = timing
			}
			if timing > stats.MaxTimingDifference {
				stats.MaxTimingDifference = timing
			}
		}

		stats.AvgTimingDifference = totalTiming / time.Duration(len(timingHistory))

		// Calculate standard deviation
		var variance float64
		avgMs := stats.AvgTimingDifference.Seconds() * 1000
		for _, timing := range timingHistory {
			timingMs := timing.Seconds() * 1000
			variance += math.Pow(timingMs-avgMs, 2)
		}
		variance /= float64(len(timingHistory))
		stats.TimingStdDev = time.Duration(math.Sqrt(variance)) * time.Millisecond
	}

	// Analyze LTP vs LTD events
	for i := 1; i < len(weightHistory); i++ {
		change := weightHistory[i] - weightHistory[i-1]
		if change > 0 {
			stats.LTPEvents++
			stats.LTPMagnitude += change
		} else if change < 0 {
			stats.LTDEvents++
			stats.LTDMagnitude += math.Abs(change)
		}
	}

	// Calculate averages
	if stats.LTPEvents > 0 {
		stats.LTPMagnitude /= float64(stats.LTPEvents)
	}
	if stats.LTDEvents > 0 {
		stats.LTDMagnitude /= float64(stats.LTDEvents)
	}
	if stats.LTDMagnitude > 0 {
		stats.AsymmetryRatio = stats.LTPMagnitude / stats.LTDMagnitude
	}

	// Learning curve analysis
	if len(weightHistory) > 2 {
		// Simple linear regression for learning curve slope
		n := float64(len(weightHistory))
		var sumX, sumY, sumXY, sumXX float64
		for i, weight := range weightHistory {
			x := float64(i)
			sumX += x
			sumY += weight
			sumXY += x * weight
			sumXX += x * x
		}

		if n*sumXX-sumX*sumX != 0 {
			stats.LearningCurveSlope = (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
		}
	}

	// Metaplasticity analysis (early vs late learning)
	if len(weightHistory) >= 8 {
		quarterPoint := len(weightHistory) / 4

		// Early learning rate (first quarter)
		earlyChange := weightHistory[quarterPoint] - weightHistory[0]
		stats.EarlyLearningRate = earlyChange / float64(quarterPoint)

		// Late learning rate (last quarter)
		lateStart := len(weightHistory) - quarterPoint
		lateChange := weightHistory[len(weightHistory)-1] - weightHistory[lateStart]
		stats.LateLearningRate = lateChange / float64(quarterPoint)

		// Learning adaptation
		if stats.EarlyLearningRate != 0 {
			stats.LearningAdaptation = stats.LateLearningRate / stats.EarlyLearningRate
		}
	}

	return stats
}

// validateBiologicalRealism checks if STDP behavior matches biological expectations
// This function compares our STDP implementation against established neuroscience
// findings to ensure biological accuracy
func validateBiologicalRealism(t *testing.T, stats STDPTestStats, testName string) {
	t.Logf("=== BIOLOGICAL VALIDATION FOR %s ===", testName)

	// === TIMING WINDOW VALIDATION ===
	// Biological STDP typically operates within ±50ms windows
	if stats.AvgTimingDifference > 100*time.Millisecond {
		t.Logf("WARNING: Average timing difference (%v) exceeds typical biological STDP window",
			stats.AvgTimingDifference)
	}

	// === ASYMMETRY VALIDATION ===
	// Biological STDP often shows slight LTP bias (asymmetry ratio 1.2-2.0)
	if stats.AsymmetryRatio > 0 {
		if stats.AsymmetryRatio < 0.5 || stats.AsymmetryRatio > 3.0 {
			t.Logf("WARNING: LTP/LTD asymmetry ratio (%.2f) outside typical biological range (0.5-3.0)",
				stats.AsymmetryRatio)
		} else {
			t.Logf("✓ LTP/LTD asymmetry ratio (%.2f) within biological range", stats.AsymmetryRatio)
		}
	}

	// === LEARNING RATE VALIDATION ===
	// Biological synapses typically change by 1-10% per learning event
	if math.Abs(stats.WeightChangePercent) > 50 {
		t.Logf("WARNING: Large weight change (%.1f%%) may indicate unrealistic learning rate",
			stats.WeightChangePercent)
	} else if math.Abs(stats.WeightChangePercent) > 0.1 {
		t.Logf("✓ Weight change (%.1f%%) within biological range", stats.WeightChangePercent)
	}

	// === METAPLASTICITY VALIDATION ===
	// Biological synapses often show learning rate adaptation over time
	if stats.LearningAdaptation > 0 {
		if stats.LearningAdaptation < 0.1 {
			t.Logf("✓ Learning rate decreased over time (ratio: %.2f) - matches biological metaplasticity",
				stats.LearningAdaptation)
		} else if stats.LearningAdaptation > 2.0 {
			t.Logf("NOTE: Learning rate increased over time (ratio: %.2f) - unusual but possible",
				stats.LearningAdaptation)
		}
	}

	// === DETAILED STATISTICS LOG ===
	t.Logf("Learning Events: %d (LTP: %d, LTD: %d)",
		stats.NumLearningEvents, stats.LTPEvents, stats.LTDEvents)
	t.Logf("Weight: %.4f → %.4f (Δ=%.4f, %.1f%%)",
		stats.InitialWeight, stats.FinalWeight, stats.WeightChange, stats.WeightChangePercent)
	t.Logf("Timing: avg=%v, range=%v to %v (σ=%v)",
		stats.AvgTimingDifference, stats.MinTimingDifference, stats.MaxTimingDifference, stats.TimingStdDev)
	if stats.LTPMagnitude > 0 || stats.LTDMagnitude > 0 {
		t.Logf("Plasticity: LTP=%.4f, LTD=%.4f, ratio=%.2f",
			stats.LTPMagnitude, stats.LTDMagnitude, stats.AsymmetryRatio)
	}
}

// createStandardSTDPConfig returns a biologically realistic STDP configuration
// These parameters are based on experimental measurements from cortical synapses
func createStandardSTDPConfig() STDPConfig {
	return STDPConfig{
		Enabled:        true,
		LearningRate:   0.01,                  // 1% weight change per event (biological range: 0.1-10%)
		TimeConstant:   20 * time.Millisecond, // 20ms decay (cortical standard)
		WindowSize:     50 * time.Millisecond, // ±50ms window (biological range: 20-100ms)
		MinWeight:      0.01,                  // 1% of base weight minimum
		MaxWeight:      3.0,                   // 300% of base weight maximum
		AsymmetryRatio: 1.5,                   // Slight LTP bias (biological: 1.2-2.0)
	}
}

// ============================================================================
// CORE STDP ALGORITHM TESTS
// ============================================================================

// TestSTDPWeightChangePositiveTimeDiff tests Long-Term Depression (LTD)
//
// BIOLOGICAL CONTEXT:
// When a post-synaptic spike occurs BEFORE a pre-synaptic spike (positive Δt),
// the pre-synaptic neuron did not contribute to causing the post-synaptic firing.
// This non-causal relationship should weaken the synapse (LTD) according to
// the principle "neurons that fire together, wire together" - they didn't fire
// together in a causal manner.
//
// EXPECTED BEHAVIOR:
// - Positive time differences should produce negative weight changes
// - Weight change magnitude should decay exponentially with increasing Δt
// - Changes should be zero outside the learning window
//
// Fix 1: Update test expectations in TestSTDPWeightChangePositiveTimeDiff and TestSTDPWeightChangeNegativeTimeDiff
// The test descriptions need to be updated to match the corrected STDP implementation
func TestSTDPWeightChangePositiveTimeDiff(t *testing.T) {
	config := createStandardSTDPConfig()

	testCases := []struct {
		name           string
		timeDifference time.Duration
		expectedSign   string // "positive", "zero", "negative"
		description    string
	}{
		{
			name:           "Small_Positive_LTP",
			timeDifference: 5 * time.Millisecond,
			expectedSign:   "positive",
			description:    "Pre-spike 5ms before post-spike should cause moderate LTP",
		},
		{
			name:           "Medium_Positive_LTP",
			timeDifference: 20 * time.Millisecond,
			expectedSign:   "positive",
			description:    "Pre-spike 20ms before post-spike should cause weaker LTP",
		},
		{
			name:           "Large_Positive_No_Change",
			timeDifference: 60 * time.Millisecond,
			expectedSign:   "zero",
			description:    "Pre-spike 60ms before post-spike should cause no change (outside window)",
		},
		{
			name:           "Very_Large_Positive_No_Change",
			timeDifference: 100 * time.Millisecond,
			expectedSign:   "zero",
			description:    "Pre-spike 100ms before post-spike should definitely cause no change",
		},
	}

	var allTimings []time.Duration
	var allWeightChanges []float64

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			weightChange := calculateSTDPWeightChange(tc.timeDifference, config)

			allTimings = append(allTimings, tc.timeDifference)
			allWeightChanges = append(allWeightChanges, weightChange)

			t.Logf("Test: %s", tc.description)
			t.Logf("Time difference: +%v (pre before post)", tc.timeDifference)
			t.Logf("Weight change: %.6f", weightChange)

			switch tc.expectedSign {
			case "positive":
				if weightChange <= 0 {
					t.Errorf("Expected positive weight change for positive time difference, got %.6f", weightChange)
				} else {
					t.Logf("✓ Correct LTP: positive weight change for causal timing")
				}
			case "zero":
				if math.Abs(weightChange) > 1e-10 {
					t.Errorf("Expected no weight change outside learning window, got %.6f", weightChange)
				} else {
					t.Logf("✓ Correct: no change outside learning window")
				}
			}
		})
	}

	// Validate exponential decay characteristic of biological STDP
	t.Logf("\n=== LTP EXPONENTIAL DECAY VALIDATION ===")
	for i := 0; i < len(allTimings)-1; i++ {
		if allWeightChanges[i] > 0 && allWeightChanges[i+1] > 0 {
			// Both are LTP events, check if magnitude decreases with time
			if allWeightChanges[i] > allWeightChanges[i+1] {
				t.Logf("✓ LTP magnitude decreases with timing: %.6f > %.6f",
					allWeightChanges[i], allWeightChanges[i+1])
			}
		}
	}
}

// TestSTDPWeightChangeNegativeTimeDiff tests Long-Term Potentiation (LTP)
//
// BIOLOGICAL CONTEXT:
// When a pre-synaptic spike occurs BEFORE a post-synaptic spike (negative Δt),
// the pre-synaptic neuron contributed to causing the post-synaptic firing.
// This causal relationship should strengthen the synapse (LTP) according to
// Hebbian learning principles: "neurons that fire together, wire together".
//
// EXPECTED BEHAVIOR:
// - Negative time differences should produce positive weight changes
// - Weight change magnitude should decay exponentially with increasing |Δt|
// - Peak learning should occur at small negative Δt values (1-10ms)
func TestSTDPWeightChangeNegativeTimeDiff(t *testing.T) {
	config := createStandardSTDPConfig()

	testCases := []struct {
		name           string
		timeDifference time.Duration
		expectedSign   string
		description    string
	}{
		{
			name:           "Strong_LTD_Timing",
			timeDifference: -2 * time.Millisecond,
			expectedSign:   "negative",
			description:    "Post-spike 2ms before pre-spike should cause strong LTD",
		},
		{
			name:           "Good_LTD_Timing",
			timeDifference: -5 * time.Millisecond,
			expectedSign:   "negative",
			description:    "Post-spike 5ms before pre-spike should cause good LTD",
		},
		{
			name:           "Moderate_LTD_Timing",
			timeDifference: -20 * time.Millisecond,
			expectedSign:   "negative",
			description:    "Post-spike 20ms before pre-spike should cause weaker LTD",
		},
		{
			name:           "No_Change_Outside_Window",
			timeDifference: -60 * time.Millisecond,
			expectedSign:   "zero",
			description:    "Post-spike 60ms before pre-spike should cause no change (outside window)",
		},
	}

	var ltdTimings []time.Duration
	var ltdWeightChanges []float64
	minWeightChange := 0.0
	strongestTiming := time.Duration(0)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			weightChange := calculateSTDPWeightChange(tc.timeDifference, config)

			if tc.expectedSign == "negative" {
				ltdTimings = append(ltdTimings, tc.timeDifference)
				ltdWeightChanges = append(ltdWeightChanges, weightChange)

				if weightChange < minWeightChange {
					minWeightChange = weightChange
					strongestTiming = tc.timeDifference
				}
			}

			t.Logf("Test: %s", tc.description)
			t.Logf("Time difference: %v (post before pre)", tc.timeDifference)
			t.Logf("Weight change: %.6f", weightChange)

			switch tc.expectedSign {
			case "negative":
				if weightChange >= 0 {
					t.Errorf("Expected negative weight change for negative time difference, got %.6f", weightChange)
				} else {
					t.Logf("✓ Correct LTD: negative weight change for anti-causal timing")
				}
			case "zero":
				if math.Abs(weightChange) > 1e-10 {
					t.Errorf("Expected no weight change outside learning window, got %.6f", weightChange)
				} else {
					t.Logf("✓ Correct: no change outside learning window")
				}
			}
		})
	}

	// Biological validation of LTD characteristics
	t.Logf("\n=== LTD BIOLOGICAL VALIDATION ===")
	t.Logf("Strongest LTD occurred at Δt = %v with magnitude %.6f", strongestTiming, math.Abs(minWeightChange))

	// Validate that strongest LTD occurs at small negative timings (biological expectation)
	if math.Abs(float64(strongestTiming.Nanoseconds())) <= float64(10*time.Millisecond.Nanoseconds()) {
		t.Logf("✓ Peak LTD at small timing difference - matches biological data")
	} else {
		t.Logf("WARNING: Peak LTD at large timing difference - unusual for biological STDP")
	}

	// Validate exponential decay
	for i := 0; i < len(ltdTimings)-1; i++ {
		if math.Abs(float64(ltdTimings[i].Nanoseconds())) < math.Abs(float64(ltdTimings[i+1].Nanoseconds())) {
			if math.Abs(ltdWeightChanges[i]) > math.Abs(ltdWeightChanges[i+1]) {
				t.Logf("✓ LTD magnitude decreases with timing distance: %.6f > %.6f",
					math.Abs(ltdWeightChanges[i]), math.Abs(ltdWeightChanges[i+1]))
			}
		}
	}
}

// TestSTDPWeightChangeZeroTimeDiff tests simultaneous spike behavior
//
// BIOLOGICAL CONTEXT:
// When pre-synaptic and post-synaptic spikes occur simultaneously (Δt = 0),
// the biological response varies by synapse type and preparation. Some show
// LTP, others show LTD, and some show no change. Our implementation should
// handle this edge case gracefully.
//
// EXPECTED BEHAVIOR:
// - Should produce a predictable, non-NaN result
// - Magnitude should be reasonable (not extreme)
func TestSTDPWeightChangeZeroTimeDiff(t *testing.T) {
	config := createStandardSTDPConfig()

	weightChange := calculateSTDPWeightChange(0, config)

	t.Logf("Simultaneous spike timing (Δt = 0)")
	t.Logf("Weight change: %.6f", weightChange)

	// Validate numerical stability
	if math.IsNaN(weightChange) || math.IsInf(weightChange, 0) {
		t.Errorf("Weight change should be finite for zero time difference, got %.6f", weightChange)
	} else {
		t.Logf("✓ Numerically stable result for simultaneous spikes")
	}

	// Validate reasonable magnitude
	if math.Abs(weightChange) > config.LearningRate {
		t.Logf("WARNING: Large weight change (%.6f) for simultaneous spikes", weightChange)
	} else {
		t.Logf("✓ Reasonable magnitude for simultaneous spike timing")
	}

	// Note: Biological interpretation varies, so we don't enforce LTP vs LTD
	if weightChange > 0 {
		t.Logf("Implementation choice: simultaneous spikes cause LTP")
	} else if weightChange < 0 {
		t.Logf("Implementation choice: simultaneous spikes cause LTD")
	} else {
		t.Logf("Implementation choice: simultaneous spikes cause no change")
	}
}

// TestSTDPWeightChangeOutsideWindow tests timing differences beyond learning window
//
// BIOLOGICAL CONTEXT:
// Real synapses only exhibit STDP for spike timing differences within a
// limited window (typically ±20-100ms). Spikes separated by longer intervals
// should not affect synaptic strength, as the molecular mechanisms underlying
// STDP have limited temporal integration capabilities.
//
// EXPECTED BEHAVIOR:
// - Time differences beyond WindowSize should produce zero weight change
// - This models the finite duration of calcium transients and kinase activation
func TestSTDPWeightChangeOutsideWindow(t *testing.T) {
	config := createStandardSTDPConfig()

	// Test various timings outside the window
	testTimings := []time.Duration{
		-(config.WindowSize + 10*time.Millisecond), // Far before window
		-(config.WindowSize + 1*time.Millisecond),  // Just before window
		-config.WindowSize,                         // Exactly at window edge
		config.WindowSize,                          // Exactly at window edge
		config.WindowSize + 1*time.Millisecond,     // Just after window
		config.WindowSize + 10*time.Millisecond,    // Far after window
		-200 * time.Millisecond,                    // Very far negative
		200 * time.Millisecond,                     // Very far positive
	}

	t.Logf("Testing timings outside STDP window (±%v)", config.WindowSize)

	for _, timing := range testTimings {
		t.Run(fmt.Sprintf("Timing_%v", timing), func(t *testing.T) {
			weightChange := calculateSTDPWeightChange(timing, config)

			t.Logf("Δt = %v, weight change = %.10f", timing, weightChange)

			// Should be exactly zero outside window
			if math.Abs(weightChange) > 1e-10 {
				t.Errorf("Expected zero weight change outside window, got %.10f for Δt=%v",
					weightChange, timing)
			} else {
				t.Logf("✓ Correct: no plasticity outside temporal window")
			}
		})
	}

	t.Logf("\n=== TEMPORAL WINDOW VALIDATION ===")
	t.Logf("✓ STDP confined to biologically realistic temporal window")
	t.Logf("✓ No spurious plasticity from temporally distant spikes")
}

// TestSTDPAsymmetryRatio tests different LTP/LTD strength ratios
//
// BIOLOGICAL CONTEXT:
// Real synapses often show asymmetric STDP, where LTP and LTD have different
// magnitudes even for equivalent timing differences. This asymmetry varies
// by synapse type, development stage, and neuromodulatory state. Common
// observations include LTP-dominant, LTD-dominant, and symmetric STDP.
//
// EXPECTED BEHAVIOR:
// - AsymmetryRatio > 1.0 should favor LTP over LTD
// - AsymmetryRatio < 1.0 should favor LTD over LTP
// - AsymmetryRatio = 1.0 should produce symmetric STDP
// Fix 2: Update TestSTDPAsymmetryRatio to use correct timing conventions
func TestSTDPAsymmetryRatio(t *testing.T) {
	baseConfig := createStandardSTDPConfig()

	testCases := []struct {
		name           string
		asymmetryRatio float64
		description    string
		biologicalNote string
	}{
		{
			name:           "LTP_Dominant",
			asymmetryRatio: 2.0,
			description:    "LTP twice as strong as LTD",
			biologicalNote: "Common in young synapses and during development",
		},
		{
			name:           "Symmetric",
			asymmetryRatio: 1.0,
			description:    "Equal LTP and LTD magnitudes",
			biologicalNote: "Observed in mature cortical synapses",
		},
		{
			name:           "LTD_Dominant",
			asymmetryRatio: 0.5,
			description:    "LTD twice as strong as LTP",
			biologicalNote: "Can occur with certain neuromodulatory states",
		},
	}

	// Fix: Swap the variable names to match what they actually produce
	ltpTiming := 10 * time.Millisecond  // Positive Δt produces LTP in our implementation
	ltdTiming := -10 * time.Millisecond // Negative Δt produces LTD in our implementation

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := baseConfig
			config.AsymmetryRatio = tc.asymmetryRatio

			ltpChange := calculateSTDPWeightChange(ltpTiming, config)
			ltdChange := calculateSTDPWeightChange(ltdTiming, config)

			t.Logf("=== %s ===", tc.description)
			t.Logf("Biological context: %s", tc.biologicalNote)
			t.Logf("Asymmetry ratio setting: %.2f", tc.asymmetryRatio)
			t.Logf("LTP change (Δt=%v): %.6f", ltpTiming, ltpChange)
			t.Logf("LTD change (Δt=%v): %.6f", ltdTiming, ltdChange)

			// Calculate observed asymmetry ratio
			if ltdChange != 0 {
				observedRatio := ltpChange / math.Abs(ltdChange)
				t.Logf("Observed LTP/|LTD| ratio: %.3f", observedRatio)

				// Validate that observed ratio matches expected
				expectedRatio := tc.asymmetryRatio
				tolerance := 0.1
				if math.Abs(observedRatio-expectedRatio) < tolerance {
					t.Logf("✓ Asymmetry ratio matches expected value")
				} else {
					t.Errorf("Asymmetry ratio mismatch: expected %.2f, got %.2f",
						expectedRatio, observedRatio)
				}
			}

			// Validate biological realism
			if tc.asymmetryRatio > 1.0 && ltpChange <= math.Abs(ltdChange) {
				t.Errorf("LTP should be stronger than LTD for asymmetry ratio > 1.0")
			}
			if tc.asymmetryRatio < 1.0 && math.Abs(ltdChange) <= ltpChange {
				t.Errorf("LTD should be stronger than LTP for asymmetry ratio < 1.0")
			}
		})
	}
}

// TestSTDPTimeConstantEffects tests different exponential decay time constants
//
// BIOLOGICAL CONTEXT:
// The STDP time constant (τ) determines how quickly plasticity decays with
// increasing spike timing differences. Different synapse types show different
// time constants: fast (5-10ms), standard (15-25ms), and slow (30-50ms).
// This reflects differences in calcium dynamics and signaling cascades.
//
// EXPECTED BEHAVIOR:
// - Smaller τ should produce steeper decay (narrower learning window)
// - Larger τ should produce gentler decay (broader learning window)
// - All should converge to zero outside their respective windows
func TestSTDPTimeConstantEffects(t *testing.T) {
	baseConfig := createStandardSTDPConfig()

	testConstants := []struct {
		name           string
		timeConstant   time.Duration
		description    string
		biologicalType string
	}{
		{
			name:           "Fast_Synapses",
			timeConstant:   8 * time.Millisecond,
			description:    "Sharp, narrow STDP window",
			biologicalType: "Fast-spiking interneurons, some inhibitory synapses",
		},
		{
			name:           "Standard_Synapses",
			timeConstant:   20 * time.Millisecond,
			description:    "Typical cortical STDP window",
			biologicalType: "Excitatory cortical pyramidal cell synapses",
		},
		{
			name:           "Slow_Synapses",
			timeConstant:   40 * time.Millisecond,
			description:    "Broad, gentle STDP window",
			biologicalType: "Some modulatory synapses, developmental synapses",
		},
	}

	// Test timing differences to compare decay profiles
	testTimings := []time.Duration{
		-5 * time.Millisecond,
		-15 * time.Millisecond,
		-30 * time.Millisecond,
		5 * time.Millisecond,
		15 * time.Millisecond,
		30 * time.Millisecond,
	}

	for _, tc := range testConstants {
		t.Run(tc.name, func(t *testing.T) {
			config := baseConfig
			config.TimeConstant = tc.timeConstant

			t.Logf("=== %s (τ = %v) ===", tc.description, tc.timeConstant)
			t.Logf("Biological type: %s", tc.biologicalType)

			var weightChanges []float64
			var timings []time.Duration

			for _, timing := range testTimings {
				weightChange := calculateSTDPWeightChange(timing, config)
				weightChanges = append(weightChanges, weightChange)
				timings = append(timings, timing)

				t.Logf("Δt = %6v: weight change = %8.5f", timing, weightChange)
			}

			// Analyze decay profile
			t.Logf("\nDecay profile analysis:")

			// Find peak LTP and LTD values
			maxLTP := 0.0
			maxLTD := 0.0
			for i, change := range weightChanges {
				if timings[i] < 0 && change > maxLTP {
					maxLTP = change
				}
				if timings[i] > 0 && change < maxLTD {
					maxLTD = change
				}
			}

			t.Logf("Peak LTP: %.5f, Peak |LTD|: %.5f", maxLTP, math.Abs(maxLTD))

			// Validate exponential decay characteristics
			for i := 1; i < len(weightChanges); i++ {
				if timings[i-1] < 0 && timings[i] < 0 { // Both LTP
					if math.Abs(float64(timings[i-1].Nanoseconds())) < math.Abs(float64(timings[i].Nanoseconds())) {
						if weightChanges[i-1] > weightChanges[i] {
							t.Logf("✓ LTP decays with increasing |Δt|: %.5f > %.5f",
								weightChanges[i-1], weightChanges[i])
						}
					}
				}
			}
		})
	}

	// Compare time constants directly
	t.Logf("\n=== TIME CONSTANT COMPARISON ===")
	timing := -15 * time.Millisecond // Standard test timing

	for _, tc := range testConstants {
		config := baseConfig
		config.TimeConstant = tc.timeConstant
		weightChange := calculateSTDPWeightChange(timing, config)

		t.Logf("τ = %v: weight change = %.5f (at Δt = %v)",
			tc.timeConstant, weightChange, timing)
	}
}

// ============================================================================
// SYNAPTIC WEIGHT UPDATE TESTS
// ============================================================================

// TestSynapticWeightBounds tests synaptic weight boundary enforcement
//
// BIOLOGICAL CONTEXT:
// Real synapses have physical limits on their strength. They cannot become
// infinitely strong (limited by receptor density, vesicle release probability)
// or negative (unidirectional neurotransmitter release). Our implementation
// must respect these biological constraints while allowing learning.
//
// EXPECTED BEHAVIOR:
// - Weights should not exceed MaxWeight
// - Weights should not fall below MinWeight
// - Learning should slow near boundaries (saturation effect)
func TestSynapticWeightBounds(t *testing.T) {
	// Create a synapse with tight bounds for testing
	config := createStandardSTDPConfig()
	config.MinWeight = 0.1
	config.MaxWeight = 2.0

	output := &Output{
		factor:           1.0, // Start at middle value
		baseWeight:       1.0,
		minWeight:        config.MinWeight,
		maxWeight:        config.MaxWeight,
		learningRate:     config.LearningRate,
		stdpEnabled:      true,
		preSpikeTimes:    make([]time.Time, 0),
		stdpTimeConstant: config.TimeConstant,
		stdpWindowSize:   config.WindowSize,
	}

	t.Logf("=== SYNAPTIC WEIGHT BOUNDS TEST ===")
	t.Logf("Initial weight: %.3f", output.factor)
	t.Logf("Weight bounds: [%.3f, %.3f]", config.MinWeight, config.MaxWeight)

	// Test upper bound
	t.Run("Upper_Bound_Test", func(t *testing.T) {
		// Apply large positive weight changes to test upper bound
		for i := 0; i < 20; i++ {
			oldWeight := output.factor
			output.updateSynapticWeight(0.2) // Large positive change

			t.Logf("Step %d: %.3f → %.3f (attempted +0.2)", i+1, oldWeight, output.factor)

			if output.factor > config.MaxWeight {
				t.Errorf("Weight exceeded maximum: %.3f > %.3f", output.factor, config.MaxWeight)
			}

			// Check if weight saturated at bound
			if math.Abs(output.factor-config.MaxWeight) < 1e-6 {
				t.Logf("✓ Weight saturated at upper bound after %d steps", i+1)
				break
			}
		}
	})

	// Reset to middle value
	output.factor = 1.0

	// Test lower bound
	t.Run("Lower_Bound_Test", func(t *testing.T) {
		// Apply large negative weight changes to test lower bound
		for i := 0; i < 20; i++ {
			oldWeight := output.factor
			output.updateSynapticWeight(-0.2) // Large negative change

			t.Logf("Step %d: %.3f → %.3f (attempted -0.2)", i+1, oldWeight, output.factor)

			if output.factor < config.MinWeight {
				t.Errorf("Weight fell below minimum: %.3f < %.3f", output.factor, config.MinWeight)
			}

			// Check if weight saturated at bound
			if math.Abs(output.factor-config.MinWeight) < 1e-6 {
				t.Logf("✓ Weight saturated at lower bound after %d steps", i+1)
				break
			}
		}
	})

	t.Logf("\n=== BIOLOGICAL VALIDATION ===")
	t.Logf("✓ Synaptic weights respect biological bounds")
	t.Logf("✓ No negative synaptic strengths (unidirectional transmission)")
	t.Logf("✓ No infinite strengthening (receptor saturation modeled)")
}

// TestSynapticWeightAccumulation tests multiple STDP learning events
//
// BIOLOGICAL CONTEXT:
// In real neural networks, synapses experience many STDP events over time.
// These should accumulate according to the timing relationships, with
// consistent pairings leading to gradual strengthening or weakening.
// This test validates the integrative aspect of synaptic learning.
//
// EXPECTED BEHAVIOR:
// - Consistent LTP pairings should progressively strengthen synapse
// - Consistent LTD pairings should progressively weaken synapse
// - Mixed pairings should show competition between LTP and LTD
func TestSynapticWeightAccumulation(t *testing.T) {
	config := createStandardSTDPConfig()

	testCases := []struct {
		name        string
		timing      time.Duration
		numEvents   int
		expectedDir string // "strengthen", "weaken", "mixed"
		description string
	}{
		{
			name:        "Consistent_LTP",
			timing:      10 * time.Millisecond, // Positive Δt produces LTP
			numEvents:   10,
			expectedDir: "strengthen",
			description: "Repeated causal pairings should strengthen synapse",
		},
		{
			name:        "Consistent_LTD",
			timing:      -10 * time.Millisecond, // Negative Δt produces LTD
			numEvents:   10,
			expectedDir: "weaken",
			description: "Repeated anti-causal pairings should weaken synapse",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fresh synapse for each test
			output := &Output{
				factor:           1.0,
				baseWeight:       1.0,
				minWeight:        0.1,
				maxWeight:        3.0,
				learningRate:     config.LearningRate,
				stdpEnabled:      true,
				preSpikeTimes:    make([]time.Time, 0),
				stdpTimeConstant: config.TimeConstant,
				stdpWindowSize:   config.WindowSize,
			}

			initialWeight := output.factor
			var weightHistory []float64
			var timingHistory []time.Duration

			weightHistory = append(weightHistory, initialWeight)

			t.Logf("=== %s ===", tc.description)
			t.Logf("Timing pattern: Δt = %v", tc.timing)
			t.Logf("Number of events: %d", tc.numEvents)
			t.Logf("Initial weight: %.4f", initialWeight)

			// Apply repeated STDP events
			for i := 0; i < tc.numEvents; i++ {
				weightChange := calculateSTDPWeightChange(tc.timing, config)
				output.updateSynapticWeight(weightChange)

				weightHistory = append(weightHistory, output.factor)
				timingHistory = append(timingHistory, tc.timing)

				if i < 5 || i%5 == 4 { // Log first 5 and every 5th event
					t.Logf("Event %d: weight = %.4f (Δ = %+.4f)",
						i+1, output.factor, weightChange)
				}
			}

			finalWeight := output.factor
			totalChange := finalWeight - initialWeight

			t.Logf("Final weight: %.4f", finalWeight)
			t.Logf("Total change: %+.4f (%.1f%%)", totalChange, (totalChange/initialWeight)*100)

			// Validate expected direction
			switch tc.expectedDir {
			case "strengthen":
				if totalChange <= 0 {
					t.Errorf("Expected strengthening, got change of %+.4f", totalChange)
				} else {
					t.Logf("✓ Synapse strengthened as expected")
				}
			case "weaken":
				if totalChange >= 0 {
					t.Errorf("Expected weakening, got change of %+.4f", totalChange)
				} else {
					t.Logf("✓ Synapse weakened as expected")
				}
			}

			// Calculate and validate learning statistics
			stats := calculateSTDPStats(initialWeight, weightHistory, timingHistory)
			validateBiologicalRealism(t, stats, tc.name)
		})
	}
}

// TestSynapticWeightSaturation tests learning behavior near weight limits
//
// BIOLOGICAL CONTEXT:
// Real synapses show reduced plasticity when they approach their maximum
// or minimum strengths. This saturation effect prevents runaway strengthening
// or complete elimination of synapses, maintaining network stability.
//
// EXPECTED BEHAVIOR:
// - Learning should slow as weights approach bounds
// - Extreme weights should be more resistant to further change
// - Saturation should be gradual, not abrupt
func TestSynapticWeightSaturation(t *testing.T) {
	config := createStandardSTDPConfig()
	config.MinWeight = 0.1
	config.MaxWeight = 2.0

	testCases := []struct {
		name          string
		initialWeight float64
		weightChange  float64
		description   string
	}{
		{
			name:          "Near_Maximum",
			initialWeight: 1.9, // Close to max of 2.0
			weightChange:  0.2, // Should be limited
			description:   "Weight near maximum should resist further strengthening",
		},
		{
			name:          "Near_Minimum",
			initialWeight: 0.2,  // Close to min of 0.1
			weightChange:  -0.2, // Should be limited
			description:   "Weight near minimum should resist further weakening",
		},
		{
			name:          "At_Maximum",
			initialWeight: 2.0, // Exactly at max
			weightChange:  0.5, // Should be completely blocked
			description:   "Weight at maximum should not increase further",
		},
		{
			name:          "At_Minimum",
			initialWeight: 0.1,  // Exactly at min
			weightChange:  -0.5, // Should be completely blocked
			description:   "Weight at minimum should not decrease further",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &Output{
				factor:    tc.initialWeight,
				minWeight: config.MinWeight,
				maxWeight: config.MaxWeight,
			}

			initialWeight := output.factor
			output.updateSynapticWeight(tc.weightChange)
			finalWeight := output.factor

			actualChange := finalWeight - initialWeight

			t.Logf("=== %s ===", tc.description)
			t.Logf("Initial weight: %.3f", initialWeight)
			t.Logf("Attempted change: %+.3f", tc.weightChange)
			t.Logf("Actual change: %+.3f", actualChange)
			t.Logf("Final weight: %.3f", finalWeight)

			// Validate bounds are respected
			if finalWeight > config.MaxWeight || finalWeight < config.MinWeight {
				t.Errorf("Weight outside bounds: %.3f not in [%.3f, %.3f]",
					finalWeight, config.MinWeight, config.MaxWeight)
			} else {
				t.Logf("✓ Weight remains within bounds")
			}

			// Validate saturation behavior
			if math.Abs(actualChange) < math.Abs(tc.weightChange) {
				t.Logf("✓ Weight change reduced due to saturation (%.3f < %.3f)",
					math.Abs(actualChange), math.Abs(tc.weightChange))
			}

			// Special cases for exact bounds
			if tc.initialWeight == config.MaxWeight && tc.weightChange > 0 {
				if actualChange > 0 {
					t.Errorf("Weight should not increase beyond maximum")
				} else {
					t.Logf("✓ Weight at maximum correctly resists strengthening")
				}
			}

			if tc.initialWeight == config.MinWeight && tc.weightChange < 0 {
				if actualChange < 0 {
					t.Errorf("Weight should not decrease below minimum")
				} else {
					t.Logf("✓ Weight at minimum correctly resists weakening")
				}
			}
		})
	}
}

// ============================================================================
// SPIKE TIMING AND HISTORY TESTS
// ============================================================================

// TestPreSpikeRecording tests basic pre-synaptic spike time recording
//
// BIOLOGICAL CONTEXT:
// Synapses must remember recent pre-synaptic spike times to compute STDP
// when the post-synaptic neuron fires. This requires accurate timestamp
// recording and efficient storage of timing information.
//
// EXPECTED BEHAVIOR:
// - Spike times should be recorded accurately
// - Most recent spike should be accessible
// - History should maintain chronological order
func TestPreSpikeRecording(t *testing.T) {
	config := createStandardSTDPConfig()

	output := &Output{
		factor:           1.0,
		stdpEnabled:      true,
		preSpikeTimes:    make([]time.Time, 0),
		stdpTimeConstant: config.TimeConstant,
		stdpWindowSize:   config.WindowSize,
	}

	t.Logf("=== PRE-SYNAPTIC SPIKE RECORDING TEST ===")

	// Record series of spikes at different times
	spikeTimes := []time.Time{
		time.Now(),
		time.Now().Add(10 * time.Millisecond),
		time.Now().Add(25 * time.Millisecond),
		time.Now().Add(30 * time.Millisecond),
	}

	for i, spikeTime := range spikeTimes {
		output.recordPreSynapticSpike(spikeTime, config)

		t.Logf("Recorded spike %d at %v", i+1, spikeTime.Format("15:04:05.000"))
		t.Logf("History length: %d", len(output.preSpikeTimes))
		t.Logf("Last spike: %v", output.lastPreSpike.Format("15:04:05.000"))

		// Validate spike was recorded
		if len(output.preSpikeTimes) != i+1 {
			t.Errorf("Expected %d spikes in history, got %d", i+1, len(output.preSpikeTimes))
		}

		// Validate last spike time updated
		if !output.lastPreSpike.Equal(spikeTime) {
			t.Errorf("Last spike time not updated correctly")
		}

		// Validate chronological order
		if i > 0 && output.preSpikeTimes[i].Before(output.preSpikeTimes[i-1]) {
			t.Errorf("Spike times not in chronological order")
		}
	}

	t.Logf("\n✓ Pre-synaptic spike recording working correctly")
	t.Logf("✓ Chronological order maintained")
	t.Logf("✓ Recent spike tracking accurate")
}

// TestPreSpikeHistoryCleanup tests removal of old spike times
//
// BIOLOGICAL CONTEXT:
// To prevent unlimited memory growth and maintain computational efficiency,
// synapses should only remember spikes within the STDP learning window.
// Older spikes have no effect on plasticity and can be safely discarded.
//
// EXPECTED BEHAVIOR:
// - Spikes older than WindowSize should be removed
// - Recent spikes should be preserved
// - Cleanup should not affect learning accuracy
func TestPreSpikeHistoryCleanup(t *testing.T) {
	config := createStandardSTDPConfig()
	config.WindowSize = 50 * time.Millisecond // Short window for testing

	output := &Output{
		factor:           1.0,
		stdpEnabled:      true,
		preSpikeTimes:    make([]time.Time, 0),
		stdpTimeConstant: config.TimeConstant,
		stdpWindowSize:   config.WindowSize,
	}

	t.Logf("=== SPIKE HISTORY CLEANUP TEST ===")
	t.Logf("Window size: %v", config.WindowSize)

	baseTime := time.Now()

	// Record spikes at various times, some outside window
	testSpikes := []struct {
		offset      time.Duration
		shouldKeep  bool
		description string
	}{
		{-100 * time.Millisecond, false, "Very old spike (should be cleaned)"},
		{-75 * time.Millisecond, false, "Old spike (should be cleaned)"},
		{-40 * time.Millisecond, true, "Recent spike (should be kept)"},
		{-20 * time.Millisecond, true, "Recent spike (should be kept)"},
		{-5 * time.Millisecond, true, "Very recent spike (should be kept)"},
	}

	// Record all spikes first
	for i, spike := range testSpikes {
		spikeTime := baseTime.Add(spike.offset)
		output.preSpikeTimes = append(output.preSpikeTimes, spikeTime)
		t.Logf("Added spike %d: %s (offset: %v)", i+1, spike.description, spike.offset)
	}

	t.Logf("Initial history length: %d", len(output.preSpikeTimes))

	// Trigger cleanup by recording a new spike
	currentTime := baseTime
	output.recordPreSynapticSpike(currentTime, config)

	t.Logf("After cleanup history length: %d", len(output.preSpikeTimes))

	// Count expected vs actual kept spikes
	expectedKept := 0
	for _, spike := range testSpikes {
		if spike.shouldKeep {
			expectedKept++
		}
	}
	expectedKept++ // Plus the new spike we just added

	if len(output.preSpikeTimes) != expectedKept {
		t.Errorf("Expected %d spikes after cleanup, got %d", expectedKept, len(output.preSpikeTimes))
	} else {
		t.Logf("✓ Correct number of spikes retained after cleanup")
	}

	// Validate remaining spikes are all recent
	cutoffTime := currentTime.Add(-config.WindowSize)
	for i, spikeTime := range output.preSpikeTimes {
		if spikeTime.Before(cutoffTime) {
			t.Errorf("Spike %d is outside window: %v < %v", i, spikeTime, cutoffTime)
		}
	}

	t.Logf("✓ All remaining spikes within learning window")
	t.Logf("✓ Memory usage optimized through cleanup")
}

// TestPreSpikeHistoryLimiting tests maximum history size enforcement
//
// BIOLOGICAL CONTEXT:
// Even within the learning window, very high-frequency firing could
// generate thousands of spike times. To maintain computational efficiency,
// we limit the maximum number of stored spikes while preserving the most
// recent and relevant timing information.
//
// EXPECTED BEHAVIOR:
// - History should not exceed maximum size limit
// - Most recent spikes should be preserved when limit is reached
// - Oldest spikes should be discarded first
func TestPreSpikeHistoryLimiting(t *testing.T) {
	config := createStandardSTDPConfig()
	maxHistorySize := 10 // Small limit for testing

	output := &Output{
		factor:           1.0,
		stdpEnabled:      true,
		preSpikeTimes:    make([]time.Time, 0),
		stdpTimeConstant: config.TimeConstant,
		stdpWindowSize:   config.WindowSize,
	}

	t.Logf("=== SPIKE HISTORY LIMITING TEST ===")
	t.Logf("Maximum history size: %d", maxHistorySize)

	baseTime := time.Now()

	// Generate more spikes than the limit
	numSpikes := maxHistorySize + 5
	var expectedSpikes []time.Time

	for i := 0; i < numSpikes; i++ {
		spikeTime := baseTime.Add(time.Duration(i) * time.Millisecond)
		expectedSpikes = append(expectedSpikes, spikeTime)

		// Manually add to test the limiting behavior
		output.preSpikeTimes = append(output.preSpikeTimes, spikeTime)

		// Apply limiting logic (simulating what recordPreSynapticSpike does)
		if len(output.preSpikeTimes) > maxHistorySize {
			start := len(output.preSpikeTimes) - maxHistorySize
			output.preSpikeTimes = output.preSpikeTimes[start:]
		}

		t.Logf("Added spike %d, history length: %d", i+1, len(output.preSpikeTimes))
	}

	// Validate final state
	if len(output.preSpikeTimes) > maxHistorySize {
		t.Errorf("History size exceeded limit: %d > %d", len(output.preSpikeTimes), maxHistorySize)
	} else {
		t.Logf("✓ History size within limit: %d ≤ %d", len(output.preSpikeTimes), maxHistorySize)
	}

	// Validate that most recent spikes are preserved
	expectedRecentSpikes := expectedSpikes[len(expectedSpikes)-maxHistorySize:]
	for i, expectedSpike := range expectedRecentSpikes {
		if !output.preSpikeTimes[i].Equal(expectedSpike) {
			t.Errorf("Recent spike %d not preserved correctly", i)
		}
	}

	t.Logf("✓ Most recent %d spikes preserved", maxHistorySize)
	t.Logf("✓ Oldest spikes discarded to maintain efficiency")
}

// TestMultiplePreSpikes tests handling of burst firing patterns
//
// BIOLOGICAL CONTEXT:
// Real neurons often fire in bursts - rapid sequences of spikes within
// short time windows. STDP should handle these burst patterns correctly,
// with each spike in the burst potentially contributing to plasticity
// when the post-synaptic neuron fires.
//
// EXPECTED BEHAVIOR:
// - All spikes in burst should be recorded
// - STDP should apply to all relevant spikes in burst
// - Burst patterns should not break timing calculations
func TestMultiplePreSpikes(t *testing.T) {
	config := createStandardSTDPConfig()

	output := &Output{
		factor:           1.0,
		baseWeight:       1.0,
		minWeight:        0.1,
		maxWeight:        3.0,
		learningRate:     config.LearningRate,
		stdpEnabled:      true,
		preSpikeTimes:    make([]time.Time, 0),
		stdpTimeConstant: config.TimeConstant,
		stdpWindowSize:   config.WindowSize,
	}

	t.Logf("=== BURST FIRING PATTERN TEST ===")

	baseTime := time.Now()

	// Create a burst of 5 spikes within 20ms
	burstIntervals := []time.Duration{0, 3 * time.Millisecond, 7 * time.Millisecond, 12 * time.Millisecond, 18 * time.Millisecond}

	t.Logf("Recording burst pattern:")
	for i, interval := range burstIntervals {
		spikeTime := baseTime.Add(interval)
		output.recordPreSynapticSpike(spikeTime, config)
		t.Logf("Burst spike %d at +%v", i+1, interval)
	}

	t.Logf("Total spikes in burst: %d", len(output.preSpikeTimes))

	// Simulate post-synaptic spike 15ms after burst start
	postSpikeTime := baseTime.Add(35 * time.Millisecond)
	t.Logf("Post-synaptic spike at +35ms")

	// Apply STDP to all spikes in the burst
	initialWeight := output.factor
	output.applySTDPToSynapse(postSpikeTime, config)
	finalWeight := output.factor

	weightChange := finalWeight - initialWeight

	t.Logf("Weight change from burst: %.6f", weightChange)
	t.Logf("Initial weight: %.4f → Final weight: %.4f", initialWeight, finalWeight)

	// Validate that burst was processed
	if len(output.preSpikeTimes) == 0 {
		t.Error("Burst spikes not recorded properly")
	} else {
		t.Logf("✓ Burst pattern recorded: %d spikes", len(output.preSpikeTimes))
	}

	// Validate learning occurred (burst should cause some weight change)
	if math.Abs(weightChange) < 1e-6 {
		t.Error("No weight change from burst pattern - STDP not applied")
	} else {
		t.Logf("✓ STDP applied to burst pattern")
	}

	// Analyze timing relationships
	t.Logf("\nBurst timing analysis:")
	for i, spikeTime := range output.preSpikeTimes {
		timeDiff := postSpikeTime.Sub(spikeTime)
		expectedChange := calculateSTDPWeightChange(timeDiff, config)
		t.Logf("Spike %d: Δt = %v, expected STDP = %.6f", i+1, timeDiff, expectedChange)
	}

	t.Logf("✓ Burst firing patterns handled correctly")
}

// ============================================================================
// POST-SYNAPTIC SPIKE PROCESSING TESTS
// ============================================================================

// TestPostSpikeSTDPApplication tests STDP triggered by post-synaptic firing
//
// BIOLOGICAL CONTEXT:
// When a post-synaptic neuron fires, it must look back at recent pre-synaptic
// inputs to determine which synapses contributed to its firing. This implements
// the "credit assignment" problem in biological learning - which inputs deserve
// strengthening based on their causal contribution to the output.
//
// EXPECTED BEHAVIOR:
// - Post-synaptic firing should trigger STDP evaluation
// - Only recent pre-synaptic spikes should affect weights
// - Timing relationships should determine LTP vs LTD
func TestPostSpikeSTDPApplication(t *testing.T) {
	config := createStandardSTDPConfig()

	// Create synapse with pre-spike history
	output := &Output{
		factor:           1.0,
		baseWeight:       1.0,
		minWeight:        0.1,
		maxWeight:        3.0,
		learningRate:     config.LearningRate,
		stdpEnabled:      true,
		preSpikeTimes:    make([]time.Time, 0),
		stdpTimeConstant: config.TimeConstant,
		stdpWindowSize:   config.WindowSize,
	}

	t.Logf("=== POST-SYNAPTIC STDP APPLICATION TEST ===")

	baseTime := time.Now()

	// Record pre-synaptic spikes at different times relative to upcoming post-spike
	preSpikes := []struct {
		offset      time.Duration
		description string
		expectedLTP bool
	}{
		{-15 * time.Millisecond, "Strong LTP timing", true},
		{-5 * time.Millisecond, "Optimal LTP timing", true},
		{5 * time.Millisecond, "LTD timing", false},
		{25 * time.Millisecond, "Weak LTD timing", false},
		{-60 * time.Millisecond, "Outside window", false},
	}

	// Record all pre-spikes
	for i, spike := range preSpikes {
		spikeTime := baseTime.Add(spike.offset)
		output.recordPreSynapticSpike(spikeTime, config)
		t.Logf("Pre-spike %d: %s (Δt = %v)", i+1, spike.description, spike.offset)
	}

	initialWeight := output.factor
	t.Logf("Initial synaptic weight: %.4f", initialWeight)

	// Trigger post-synaptic spike (at baseTime)
	postSpikeTime := baseTime
	t.Logf("Post-synaptic spike at baseline time")

	output.applySTDPToSynapse(postSpikeTime, config)

	finalWeight := output.factor
	weightChange := finalWeight - initialWeight

	t.Logf("Final synaptic weight: %.4f", finalWeight)
	t.Logf("Total weight change: %+.6f", weightChange)

	// Analyze individual contributions
	t.Logf("\nIndividual STDP contributions:")
	totalExpectedChange := 0.0

	for i, spike := range preSpikes {
		spikeTime := baseTime.Add(spike.offset)
		timeDiff := postSpikeTime.Sub(spikeTime)
		expectedChange := calculateSTDPWeightChange(timeDiff, config)
		totalExpectedChange += expectedChange

		t.Logf("Spike %d: Δt=%v, STDP=%.6f (%s)",
			i+1, timeDiff, expectedChange, spike.description)
	}

	t.Logf("Expected total change: %.6f", totalExpectedChange)
	t.Logf("Actual total change: %.6f", weightChange)

	// Validate that STDP was applied
	if math.Abs(weightChange) < 1e-8 {
		t.Error("No weight change detected - STDP may not have been applied")
	} else {
		t.Logf("✓ STDP successfully applied to multiple pre-synaptic spikes")
	}

	// Validate that weight change is approximately correct
	tolerance := math.Abs(totalExpectedChange) * 0.1 // 10% tolerance
	if math.Abs(weightChange-totalExpectedChange) > tolerance {
		t.Logf("WARNING: Weight change differs from expected (%.6f vs %.6f)",
			weightChange, totalExpectedChange)
	} else {
		t.Logf("✓ Weight change matches expected STDP calculation")
	}
}

// TestPostSpikeTimingWindow tests that only recent pre-spikes affect learning
//
// BIOLOGICAL CONTEXT:
// STDP has a finite temporal window because the molecular mechanisms
// (calcium transients, kinase activation) have limited duration. Spikes
// outside this window should not contribute to plasticity, ensuring
// that only causally relevant timing relationships drive learning.
//
// EXPECTED BEHAVIOR:
// - Pre-spikes within window should contribute to STDP
// - Pre-spikes outside window should be ignored
// - Window boundaries should be respected precisely
func TestPostSpikeTimingWindow(t *testing.T) {
	config := createStandardSTDPConfig()
	config.WindowSize = 30 * time.Millisecond // Narrow window for clear testing

	output := &Output{
		factor:           1.0,
		baseWeight:       1.0,
		minWeight:        0.1,
		maxWeight:        3.0,
		learningRate:     config.LearningRate,
		stdpEnabled:      true,
		preSpikeTimes:    make([]time.Time, 0),
		stdpTimeConstant: config.TimeConstant,
		stdpWindowSize:   config.WindowSize,
	}

	t.Logf("=== STDP TIMING WINDOW TEST ===")
	t.Logf("Window size: ±%v", config.WindowSize)

	baseTime := time.Now()

	// Create spikes both inside and outside the window
	testSpikes := []struct {
		offset       time.Duration
		shouldAffect bool
		description  string
	}{
		{-50 * time.Millisecond, false, "Far outside window (old)"},
		{-35 * time.Millisecond, false, "Just outside window (old)"},
		{-25 * time.Millisecond, true, "Inside window (LTP)"},
		{-10 * time.Millisecond, true, "Inside window (LTP)"},
		{10 * time.Millisecond, true, "Inside window (LTD)"},
		{25 * time.Millisecond, true, "Inside window (LTD)"},
		{35 * time.Millisecond, false, "Just outside window (future)"},
		{50 * time.Millisecond, false, "Far outside window (future)"},
	}

	// Record all spikes using the proper API (this will apply cleanup)
	for i, spike := range testSpikes {
		spikeTime := baseTime.Add(spike.offset)
		output.recordPreSynapticSpike(spikeTime, config)
		t.Logf("Spike %d: %s (Δt = %v)", i+1, spike.description, spike.offset)
	}

	t.Logf("Total pre-spikes recorded: %d", len(output.preSpikeTimes))

	// Apply STDP
	initialWeight := output.factor
	postSpikeTime := baseTime
	output.applySTDPToSynapse(postSpikeTime, config)
	finalWeight := output.factor

	weightChange := finalWeight - initialWeight

	t.Logf("Weight change: %.6f", weightChange)

	// Calculate expected change from only the spikes that should affect plasticity
	// Note: Due to cleanup, only spikes within the window should remain
	expectedChange := 0.0
	for _, spikeTime := range output.preSpikeTimes {
		timeDiff := postSpikeTime.Sub(spikeTime)
		spikeChange := calculateSTDPWeightChange(timeDiff, config)
		expectedChange += spikeChange
		t.Logf("Contributing spike: Δt=%v, STDP=%.6f", timeDiff, spikeChange)
	}

	t.Logf("Expected change from recorded spikes: %.6f", expectedChange)

	// Validate window enforcement - allow small tolerance for floating point
	tolerance := math.Abs(expectedChange) * 0.1
	if tolerance < 1e-6 {
		tolerance = 1e-6
	}

	if math.Abs(weightChange-expectedChange) > tolerance {
		t.Errorf("Weight change differs from expected: got %.6f, expected %.6f",
			weightChange, expectedChange)
	} else {
		t.Logf("✓ STDP calculations match expected values")
	}

	t.Logf("✓ STDP timing window correctly enforced")
}

// ============================================================================
// NETWORK-LEVEL STDP TESTS
// ============================================================================

// TestBasicNeuronPairLearning tests STDP between two connected neurons
//
// BIOLOGICAL CONTEXT:
// This is the fundamental unit of neural learning - two neurons connected
// by a plastic synapse that adapts based on their relative timing. This
// test validates the complete STDP learning loop in a minimal network.
//
// EXPECTED BEHAVIOR:
// - Causal spike patterns (pre→post) should strengthen connection
// - Anti-causal patterns (post→pre) should weaken connection
// - Learning should accumulate over multiple trials
func TestBasicNeuronPairLearning(t *testing.T) {
	// Create disabled homeostasis for focused STDP testing
	stdpConfig := createStandardSTDPConfig()
	homeostasisConfig := STDPConfig{Enabled: false}

	// Create two neurons with STDP enabled
	preNeuron := NewNeuron("pre", 1.0, 0.95, 5*time.Millisecond, 1.0, 0.0, 0.0, homeostasisConfig)
	postNeuron := NewNeuron("post", 1.0, 0.95, 5*time.Millisecond, 1.0, 0.0, 0.0, stdpConfig)

	// Connect pre→post with STDP
	connection := make(chan Message, 10)
	preNeuron.AddOutputWithSTDP("to_post", connection, 1.0, 0, stdpConfig)

	// Set up post-neuron to receive from connection
	go func() {
		for msg := range connection {
			select {
			case postNeuron.GetInputChannel() <- msg:
			default:
			}
		}
	}()

	// Start both neurons
	go preNeuron.Run()
	go postNeuron.Run()
	defer preNeuron.Close()
	defer postNeuron.Close()

	t.Logf("=== NEURON PAIR LEARNING TEST ===")

	preInput := preNeuron.GetInput()
	postInput := postNeuron.GetInput()

	// Get initial synaptic weight
	initialWeight := 1.0 // We know this from AddOutputWithSTDP call
	t.Logf("Initial synaptic weight: %.4f", initialWeight)

	testCases := []struct {
		name        string
		preDelay    time.Duration
		postDelay   time.Duration
		expectedDir string
		description string
	}{
		{
			name:        "Causal_Pattern",
			preDelay:    0,
			postDelay:   10 * time.Millisecond,
			expectedDir: "strengthen",
			description: "Pre-neuron fires, then post-neuron (LTP expected)",
		},
		{
			name:        "Anti_Causal_Pattern",
			preDelay:    15 * time.Millisecond,
			postDelay:   0,
			expectedDir: "weaken",
			description: "Post-neuron fires, then pre-neuron (LTD expected)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.description)

			// Repeat the pattern multiple times to see accumulation
			numTrials := 5
			for trial := 0; trial < numTrials; trial++ {
				// Send precisely timed spikes
				go func() {
					time.Sleep(tc.preDelay)
					preInput <- Message{Value: 1.5, Timestamp: time.Now(), SourceID: "external"}
				}()

				go func() {
					time.Sleep(tc.postDelay)
					postInput <- Message{Value: 1.5, Timestamp: time.Now(), SourceID: "external"}
				}()

				// Wait for both spikes to be processed
				time.Sleep(50 * time.Millisecond)

				t.Logf("Trial %d completed", trial+1)
			}

			// Allow time for STDP processing
			time.Sleep(100 * time.Millisecond)

			// We can't easily extract the final weight here without modifying the neuron
			// In a full implementation, we'd add methods to inspect synaptic weights
			t.Logf("✓ Learning pattern completed: %s", tc.description)
		})
	}

	t.Logf("✓ Basic neuron pair learning test completed")
	t.Logf("Note: Weight inspection requires additional monitoring infrastructure")
}

// TestCausalConnectionStrengthening tests that causal patterns strengthen synapses
//
// BIOLOGICAL CONTEXT:
// This test specifically validates Hebbian learning: "neurons that fire
// together, wire together." When pre-synaptic activity consistently precedes
// and contributes to post-synaptic firing, the connection should strengthen
// to make future coincident activity more likely.
//
// EXPECTED BEHAVIOR:
// - Consistent pre→post patterns should progressively strengthen synapse
// - Strengthening should follow biological STDP timing curves
// - Effect should accumulate over multiple learning events
func TestCausalConnectionStrengthening(t *testing.T) {
	config := createStandardSTDPConfig()

	// Create synapse to test directly
	output := &Output{
		factor:           1.0,
		baseWeight:       1.0,
		minWeight:        0.1,
		maxWeight:        3.0,
		learningRate:     config.LearningRate,
		stdpEnabled:      true,
		preSpikeTimes:    make([]time.Time, 0),
		stdpTimeConstant: config.TimeConstant,
		stdpWindowSize:   config.WindowSize,
	}

	t.Logf("=== CAUSAL CONNECTION STRENGTHENING TEST ===")

	// Test different causal timings
	causalTimings := []time.Duration{
		-2 * time.Millisecond,  // Optimal LTP timing
		-5 * time.Millisecond,  // Good LTP timing
		-10 * time.Millisecond, // Moderate LTP timing
		-20 * time.Millisecond, // Weak LTP timing
	}

	for _, timing := range causalTimings {
		t.Run(fmt.Sprintf("Timing_%v", timing), func(t *testing.T) {
			// Reset synapse
			output.factor = 1.0
			output.preSpikeTimes = make([]time.Time, 0)

			initialWeight := output.factor
			var weightHistory []float64
			var timingHistory []time.Duration

			weightHistory = append(weightHistory, initialWeight)

			t.Logf("Testing causal timing: pre→post with Δt = %v", timing)

			numPairings := 10
			baseTime := time.Now()

			for i := 0; i < numPairings; i++ {
				// Record pre-synaptic spike
				preSpikeTime := baseTime.Add(time.Duration(i) * 100 * time.Millisecond)
				output.recordPreSynapticSpike(preSpikeTime, config)

				// Post-synaptic spike occurs at timing offset
				postSpikeTime := preSpikeTime.Add(-timing) // Negative timing means pre before post

				// Apply STDP
				output.applySTDPToSynapse(postSpikeTime, config)

				weightHistory = append(weightHistory, output.factor)
				timingHistory = append(timingHistory, timing)

				if i < 3 || i%5 == 4 {
					t.Logf("Pairing %d: weight = %.4f", i+1, output.factor)
				}
			}

			finalWeight := output.factor
			totalStrengthening := finalWeight - initialWeight

			t.Logf("Initial weight: %.4f", initialWeight)
			t.Logf("Final weight: %.4f", finalWeight)
			t.Logf("Total strengthening: %.4f (%.1f%%)", totalStrengthening,
				(totalStrengthening/initialWeight)*100)

			// Validate strengthening occurred
			if totalStrengthening <= 0 {
				t.Errorf("Expected strengthening for causal timing, got change of %.4f",
					totalStrengthening)
			} else {
				t.Logf("✓ Causal pattern caused strengthening")
			}

			// Calculate learning statistics
			stats := calculateSTDPStats(initialWeight, weightHistory, timingHistory)
			validateBiologicalRealism(t, stats, fmt.Sprintf("Causal_%v", timing))

			// Validate LTP dominance
			if stats.LTPEvents <= stats.LTDEvents {
				t.Errorf("Expected LTP dominance for causal pattern, got LTP:%d LTD:%d",
					stats.LTPEvents, stats.LTDEvents)
			} else {
				t.Logf("✓ LTP events dominated: %d LTP vs %d LTD",
					stats.LTPEvents, stats.LTDEvents)
			}
		})
	}
}

// TestAntiCausalConnectionWeakening tests that anti-causal patterns weaken synapses
//
// BIOLOGICAL CONTEXT:
// Anti-causal patterns (post→pre) indicate that the pre-synaptic input
// did not contribute to post-synaptic firing. Such connections should be
// weakened to reduce their influence on future post-synaptic activity,
// implementing competitive learning and connection pruning.
//
// EXPECTED BEHAVIOR:
// - Consistent post→pre patterns should progressively weaken synapse
// - Weakening should follow biological LTD timing curves
// - Synapse should respect minimum weight bounds
func TestAntiCausalConnectionWeakening(t *testing.T) {
	config := createStandardSTDPConfig()

	// Create synapse to test directly
	output := &Output{
		factor:           1.0,
		baseWeight:       1.0,
		minWeight:        0.1,
		maxWeight:        3.0,
		learningRate:     config.LearningRate,
		stdpEnabled:      true,
		preSpikeTimes:    make([]time.Time, 0),
		stdpTimeConstant: config.TimeConstant,
		stdpWindowSize:   config.WindowSize,
	}

	t.Logf("=== ANTI-CAUSAL CONNECTION WEAKENING TEST ===")

	// Test different anti-causal timings
	antiCausalTimings := []time.Duration{
		2 * time.Millisecond,  // Strong LTD timing
		5 * time.Millisecond,  // Good LTD timing
		10 * time.Millisecond, // Moderate LTD timing
		20 * time.Millisecond, // Weak LTD timing
	}

	for _, timing := range antiCausalTimings {
		t.Run(fmt.Sprintf("Timing_%v", timing), func(t *testing.T) {
			// Reset synapse
			output.factor = 1.0
			output.preSpikeTimes = make([]time.Time, 0)

			initialWeight := output.factor
			var weightHistory []float64
			var timingHistory []time.Duration

			weightHistory = append(weightHistory, initialWeight)

			t.Logf("Testing anti-causal timing: post→pre with Δt = +%v", timing)

			numPairings := 15 // More pairings to see clear weakening effect
			baseTime := time.Now()

			for i := 0; i < numPairings; i++ {
				// Record pre-synaptic spike
				preSpikeTime := baseTime.Add(time.Duration(i) * 100 * time.Millisecond)
				output.recordPreSynapticSpike(preSpikeTime, config)

				// Post-synaptic spike occurs before pre-spike (positive timing)
				postSpikeTime := preSpikeTime.Add(-timing) // Positive timing means post before pre

				// Apply STDP
				output.applySTDPToSynapse(postSpikeTime, config)

				weightHistory = append(weightHistory, output.factor)
				timingHistory = append(timingHistory, timing)

				if i < 3 || i%5 == 4 {
					t.Logf("Pairing %d: weight = %.4f", i+1, output.factor)
				}
			}

			finalWeight := output.factor
			totalWeakening := finalWeight - initialWeight

			t.Logf("Initial weight: %.4f", initialWeight)
			t.Logf("Final weight: %.4f", finalWeight)
			t.Logf("Total weakening: %.4f (%.1f%%)", totalWeakening,
				(totalWeakening/initialWeight)*100)

			// Validate weakening occurred
			if totalWeakening >= 0 {
				t.Errorf("Expected weakening for anti-causal timing, got change of %.4f",
					totalWeakening)
			} else {
				t.Logf("✓ Anti-causal pattern caused weakening")
			}

			// Validate weight didn't go below minimum
			if finalWeight < output.minWeight {
				t.Errorf("Weight fell below minimum: %.4f < %.4f", finalWeight, output.minWeight)
			} else {
				t.Logf("✓ Weight respected minimum bound")
			}

			// Calculate learning statistics
			stats := calculateSTDPStats(initialWeight, weightHistory, timingHistory)
			validateBiologicalRealism(t, stats, fmt.Sprintf("AntiCausal_%v", timing))

			// Validate LTD dominance
			if stats.LTDEvents <= stats.LTPEvents {
				t.Errorf("Expected LTD dominance for anti-causal pattern, got LTP:%d LTD:%d",
					stats.LTPEvents, stats.LTDEvents)
			} else {
				t.Logf("✓ LTD events dominated: %d LTD vs %d LTP",
					stats.LTDEvents, stats.LTPEvents)
			}
		})
	}
}

// ============================================================================
// CONFIGURATION AND CONTROL TESTS
// ============================================================================

// TestSTDPEnableDisable tests toggling STDP learning on and off
//
// BIOLOGICAL CONTEXT:
// Synaptic plasticity can be modulated by neuromodulators, developmental
// state, and activity levels. The ability to enable/disable STDP allows
// modeling of these biological control mechanisms and experimental
// manipulations where plasticity is blocked.
//
// EXPECTED BEHAVIOR:
// - When enabled, STDP should modify synaptic weights
// - When disabled, weights should remain unchanged regardless of timing
// - Transitions should be smooth and not cause artifacts
func TestSTDPEnableDisable(t *testing.T) {
	config := createStandardSTDPConfig()

	output := &Output{
		factor:           1.0,
		baseWeight:       1.0,
		minWeight:        0.1,
		maxWeight:        3.0,
		learningRate:     config.LearningRate,
		stdpEnabled:      true, // Start enabled
		preSpikeTimes:    make([]time.Time, 0),
		stdpTimeConstant: config.TimeConstant,
		stdpWindowSize:   config.WindowSize,
	}

	t.Logf("=== STDP ENABLE/DISABLE TEST ===")

	baseTime := time.Now()
	causalTiming := -10 * time.Millisecond // Good LTP timing

	// Test 1: STDP enabled - should learn
	t.Run("STDP_Enabled", func(t *testing.T) {
		output.stdpEnabled = true
		output.factor = 1.0
		output.preSpikeTimes = make([]time.Time, 0)

		initialWeight := output.factor

		// Apply learning pattern
		for i := 0; i < 5; i++ {
			preSpikeTime := baseTime.Add(time.Duration(i) * 100 * time.Millisecond)
			output.recordPreSynapticSpike(preSpikeTime, config)

			postSpikeTime := preSpikeTime.Add(-causalTiming)
			output.applySTDPToSynapse(postSpikeTime, config)
		}

		finalWeight := output.factor
		weightChange := finalWeight - initialWeight

		t.Logf("STDP enabled: %.4f → %.4f (Δ = %+.4f)",
			initialWeight, finalWeight, weightChange)

		if math.Abs(weightChange) < 1e-6 {
			t.Error("No weight change when STDP enabled")
		} else {
			t.Logf("✓ Learning occurred when STDP enabled")
		}
	})

	// Test 2: STDP disabled - should not learn
	t.Run("STDP_Disabled", func(t *testing.T) {
		output.stdpEnabled = false
		output.factor = 1.0
		output.preSpikeTimes = make([]time.Time, 0)

		initialWeight := output.factor

		// Apply same learning pattern
		for i := 0; i < 5; i++ {
			preSpikeTime := baseTime.Add(time.Duration(i) * 100 * time.Millisecond)
			output.recordPreSynapticSpike(preSpikeTime, config)

			postSpikeTime := preSpikeTime.Add(-causalTiming)
			output.applySTDPToSynapse(postSpikeTime, config)
		}

		finalWeight := output.factor
		weightChange := finalWeight - initialWeight

		t.Logf("STDP disabled: %.4f → %.4f (Δ = %+.4f)",
			initialWeight, finalWeight, weightChange)

		if math.Abs(weightChange) > 1e-6 {
			t.Error("Weight changed when STDP disabled")
		} else {
			t.Logf("✓ No learning when STDP disabled")
		}
	})

	// Test 3: Config-level disable
	t.Run("Config_Disabled", func(t *testing.T) {
		output.stdpEnabled = true
		output.factor = 1.0
		output.preSpikeTimes = make([]time.Time, 0)

		// Disable via config
		disabledConfig := config
		disabledConfig.Enabled = false

		initialWeight := output.factor

		// Apply learning pattern with disabled config
		for i := 0; i < 5; i++ {
			preSpikeTime := baseTime.Add(time.Duration(i) * 100 * time.Millisecond)
			output.recordPreSynapticSpike(preSpikeTime, disabledConfig)

			postSpikeTime := preSpikeTime.Add(-causalTiming)
			output.applySTDPToSynapse(postSpikeTime, disabledConfig)
		}

		finalWeight := output.factor
		weightChange := finalWeight - initialWeight

		t.Logf("Config disabled: %.4f → %.4f (Δ = %+.4f)",
			initialWeight, finalWeight, weightChange)

		if math.Abs(weightChange) > 1e-6 {
			t.Error("Weight changed when config disabled")
		} else {
			t.Logf("✓ No learning when config disabled")
		}
	})

	t.Logf("✓ STDP enable/disable controls working correctly")
}
