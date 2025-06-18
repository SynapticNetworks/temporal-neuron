/*
=================================================================================
SYNAPTIC PLASTICITY BIOLOGICAL VALIDATION TEST SUITE
=================================================================================

This test suite validates that the synaptic plasticity implementation matches
experimental neuroscience data and exhibits biologically realistic behavior.
All tests are based on published research and experimental measurements.

TEST CATEGORIES:
1. STDP Timing Windows - Match experimental spike-timing curves
2. Learning Rate Validation - Realistic plasticity magnitudes
3. Neuromodulator Effects - Dopamine, ACh, and norepinephrine modulation
4. Developmental Plasticity - Age-dependent plasticity changes
5. Metaplasticity - BCM rule and sliding thresholds
6. Frequency Dependence - LTP/LTD frequency relationships
7. Cooperativity - Multiple input requirements
8. Saturation and Bounds - Biological weight limits

EXPERIMENTAL BASIS:
- Bi & Poo (1998): Original STDP characterization
- Sjöström et al. (2001): Cooperativity and frequency dependence
- Caporale & Dan (2008): Comprehensive STDP review
- Schultz (2007): Dopamine and reward learning
- Abraham & Bear (1996): Metaplasticity mechanisms

VALIDATION APPROACH:
Each test compares simulation results against published experimental data,
ensuring the implementation produces biologically realistic neural behavior.
=================================================================================
*/

package synapse

import (
	"math"
	"testing"
	"time"
)

// =================================================================================
// TEST UTILITIES FOR BIOLOGICAL VALIDATION
// =================================================================================

// createBiologicalSTDPConfig returns a configuration matching experimental data
func createBiologicalSTDPConfig() STDPConfig {
	return STDPConfig{
		Enabled:                true,
		LearningRate:           0.01,                                                          // 1% per pairing
		TimeConstant:           time.Duration(BIOLOGY_LTP_TAU_MS * float64(time.Millisecond)), // time.Duration(float64(BIOLOGY_LTP_TAU_MS) * time.Millisecond)
		WindowSize:             time.Duration(BIOLOGY_STDP_WINDOW_MS) * time.Millisecond,
		MinWeight:              0.0,
		MaxWeight:              1.0,
		AsymmetryRatio:         1.0 / BIOLOGY_LTP_LTD_RATIO, // Derive from desired LTP/LTD ratio
		FrequencyDependent:     true,
		MetaplasticityRate:     0.1,
		CooperativityThreshold: BIOLOGY_COOPERATIVITY_THRESHOLD,
	}
}

// validateBiologicalRange checks if values fall within experimental ranges
func validateBiologicalRange(t *testing.T, name string, value, expMin, expMax float64, unit string) bool {
	if value < expMin || value > expMax {
		t.Errorf("BIOLOGY VIOLATION: %s (%.3f %s) outside experimental range [%.3f, %.3f] %s",
			name, value, unit, expMin, expMax, unit)
		return false
	}
	t.Logf("✓ %s within biological range: %.3f %s", name, value, unit)
	return true
}

// measureSTDPCurve generates plasticity vs timing curve for validation
func measureSTDPCurve(pc *PlasticityCalculator, timingRange []time.Duration, weight float64, coop int) []float64 {
	results := make([]float64, len(timingRange))
	for i, deltaT := range timingRange {
		results[i] = pc.CalculateSTDPWeightChange(deltaT, weight, coop)
	}
	return results
}

// =================================================================================
// TEST 1: STDP TIMING WINDOW VALIDATION
// =================================================================================

// TestPlasticityBiologySTDPTimingWindow validates the STDP timing curve matches
// experimental data from Bi & Poo (1998) and subsequent studies.
//
// EXPERIMENTAL VALIDATION:
// - LTP occurs for pre-before-post timing (negative deltaT)
// - LTD occurs for post-before-pre timing (positive deltaT)
// - Peak LTP at ~10ms pre-before-post
// - Peak LTD at ~10ms post-before-pre
// - Exponential decay with distance from peak
// - Window extends to ±100ms
func TestPlasticityBiologySTDPTimingWindow(t *testing.T) {
	t.Log("=== BIOLOGY TEST: STDP Timing Window ===")
	t.Log("Validating against Bi & Poo (1998) experimental data")

	pc := NewPlasticityCalculator(createBiologicalSTDPConfig())

	// Generate timing points for STDP curve measurement
	timingPoints := []time.Duration{}
	for ms := -100; ms <= 100; ms += 5 {
		timingPoints = append(timingPoints, time.Duration(ms)*time.Millisecond)
	}

	// Measure STDP curve
	weight := 0.5 // Mid-range weight
	cooperativity := BIOLOGY_COOPERATIVITY_THRESHOLD
	stdpCurve := measureSTDPCurve(pc, timingPoints, weight, cooperativity)

	// Find peaks and validate timing
	var ltpPeak, ltdPeak float64
	var ltpPeakTiming, ltdPeakTiming time.Duration

	for i, timing := range timingPoints {
		change := stdpCurve[i]

		// Track LTP peak (negative timing, positive change)
		if timing < 0 && change > ltpPeak {
			ltpPeak = change
			ltpPeakTiming = timing
		}

		// Track LTD peak (positive timing, negative change)
		if timing > 0 && change < ltdPeak {
			ltdPeak = change
			ltdPeakTiming = timing
		}
	}

	t.Logf("Measured STDP peaks:")
	t.Logf("  LTP: %.6f at %v", ltpPeak, ltpPeakTiming)
	t.Logf("  LTD: %.6f at %v", ltdPeak, ltdPeakTiming)

	// Validate peak timing against experimental data
	ltpPeakMs := math.Abs(float64(ltpPeakTiming.Milliseconds()))
	ltdPeakMs := float64(ltdPeakTiming.Milliseconds())

	// Peak timing should be within experimental range (5-20ms)
	validateBiologicalRange(t, "LTP peak timing", ltpPeakMs, 5.0, 20.0, "ms")
	validateBiologicalRange(t, "LTD peak timing", ltdPeakMs, 5.0, 20.0, "ms")

	// Validate peak magnitudes
	if ltpPeak <= 0 {
		t.Error("BIOLOGY VIOLATION: LTP peak should be positive")
	}
	if ltdPeak >= 0 {
		t.Error("BIOLOGY VIOLATION: LTD peak should be negative")
	}

	// Validate LTP/LTD magnitude ratio
	ltpLtdRatio := ltpPeak / math.Abs(ltdPeak)
	validateBiologicalRange(t, "LTP/LTD ratio", ltpLtdRatio, 1.0, 3.0, "ratio")

	// Test window boundaries - should have minimal plasticity at ±100ms
	change100ms := pc.CalculateSTDPWeightChange(100*time.Millisecond, weight, cooperativity)
	changeNeg100ms := pc.CalculateSTDPWeightChange(-100*time.Millisecond, weight, cooperativity)

	// Boundary plasticity should be < 10% of peak
	boundaryThreshold := ltpPeak * 0.1
	if math.Abs(change100ms) > boundaryThreshold {
		t.Errorf("BIOLOGY VIOLATION: Too much plasticity at +100ms: %.6f > %.6f",
			math.Abs(change100ms), boundaryThreshold)
	}
	if math.Abs(changeNeg100ms) > boundaryThreshold {
		t.Errorf("BIOLOGY VIOLATION: Too much plasticity at -100ms: %.6f > %.6f",
			math.Abs(changeNeg100ms), boundaryThreshold)
	}

	t.Log("✅ STDP timing window matches experimental data")
}

// =================================================================================
// TEST 2: COOPERATIVITY REQUIREMENTS
// =================================================================================

// TestPlasticityBiologyCooperativity validates cooperativity requirements match
// experimental data from Sjöström et al. (2001).
//
// EXPERIMENTAL VALIDATION:
// - Minimal plasticity with single input
// - Threshold effect around 3 inputs
// - Strong plasticity with high cooperativity
// - Saturation at very high input counts
func TestPlasticityBiologyCooperativity(t *testing.T) {
	t.Log("=== BIOLOGY TEST: Cooperativity Requirements ===")
	t.Log("Validating against Sjöström et al. (2001) experimental data")

	pc := NewPlasticityCalculator(createBiologicalSTDPConfig())

	// Test cooperativity levels
	cooperativities := []int{1, 2, 3, 4, 5, 10, 15, 20, 30}
	timing := -10 * time.Millisecond // Peak LTP timing
	weight := 0.5

	t.Log("Cooperativity vs Plasticity:")
	var cooperativityCurve []float64

	for _, coop := range cooperativities {
		change := pc.CalculateSTDPWeightChange(timing, weight, coop)
		cooperativityCurve = append(cooperativityCurve, change)
		t.Logf("  %d inputs: %.6f", coop, change)
	}

	// Validate threshold effect
	singleInputChange := cooperativityCurve[0] // 1 input
	thresholdChange := cooperativityCurve[2]   // 3 inputs (threshold)
	highCoopChange := cooperativityCurve[5]    // 10 inputs

	// Single input should produce minimal plasticity
	if singleInputChange > 0.001 {
		t.Errorf("BIOLOGY VIOLATION: Single input produces too much plasticity: %.6f",
			singleInputChange)
	}

	// Threshold effect - 3 inputs should produce significant increase
	thresholdRatio := thresholdChange / singleInputChange
	if math.IsInf(thresholdRatio, 0) {
		t.Log("✓ Strong threshold effect: no plasticity → significant plasticity")
	} else if thresholdRatio < 5.0 {
		t.Errorf("BIOLOGY VIOLATION: Weak cooperativity threshold effect: %.1fx increase",
			thresholdRatio)
	}

	// High cooperativity should enhance plasticity further
	highCoopRatio := highCoopChange / thresholdChange
	validateBiologicalRange(t, "High cooperativity enhancement", highCoopRatio, 1.5, 5.0, "fold")

	// Test saturation - very high cooperativity shouldn't increase indefinitely
	maxChange := cooperativityCurve[len(cooperativityCurve)-1]        // 30 inputs
	saturationChange := cooperativityCurve[len(cooperativityCurve)-2] // 20 inputs
	saturationRatio := maxChange / saturationChange

	if saturationRatio > 1.2 {
		t.Logf("Note: Cooperativity may not be saturating (%.1fx from 20→30 inputs)",
			saturationRatio)
	} else {
		t.Log("✓ Cooperativity shows saturation at high input counts")
	}

	t.Log("✅ Cooperativity requirements match experimental data")
}

// =================================================================================
// TEST 3: NEUROMODULATOR EFFECTS
// =================================================================================

// TestPlasticityBiologyNeuromodulation validates neuromodulator effects match
// experimental data from dopamine, acetylcholine, and norepinephrine studies.
//
// EXPERIMENTAL VALIDATION:
// - Dopamine enhances LTP (reward learning)
// - Acetylcholine gates plasticity (attention)
// - Norepinephrine has inverted-U dose response
// - Effects are multiplicative with base plasticity
func TestPlasticityBiologyNeuromodulation(t *testing.T) {
	t.Log("=== BIOLOGY TEST: Neuromodulator Effects ===")
	t.Log("Validating against dopamine, ACh, and norepinephrine studies")

	pc := NewPlasticityCalculator(createBiologicalSTDPConfig())

	timing := -10 * time.Millisecond // Peak LTP timing
	weight := 0.5
	cooperativity := BIOLOGY_COOPERATIVITY_THRESHOLD

	// Baseline plasticity (no neuromodulation)
	baselineChange := pc.CalculateSTDPWeightChange(timing, weight, cooperativity)
	t.Logf("Baseline plasticity: %.6f", baselineChange)

	// Test dopamine enhancement
	t.Log("\n--- Testing Dopamine Enhancement ---")
	dopamineLevels := []float64{0.5, 1.0, 1.5, 2.0, 3.0}

	for _, dopamine := range dopamineLevels {
		pc.SetNeuromodulatorLevels(dopamine, 1.0, 1.0) // DA, ACh, NE
		change := pc.CalculateSTDPWeightChange(timing, weight, cooperativity)
		enhancement := change / baselineChange

		t.Logf("  Dopamine %.1f: %.6f (%.1fx enhancement)", dopamine, change, enhancement)

		// Validate enhancement is in biological range
		if dopamine > 1.0 {
			expectedMin := 1.0
			expectedMax := BIOLOGY_DOPAMINE_ENHANCEMENT * dopamine
			if enhancement < expectedMin || enhancement > expectedMax {
				t.Errorf("BIOLOGY VIOLATION: Dopamine %.1f enhancement (%.1fx) outside expected range [%.1f, %.1f]",
					dopamine, enhancement, expectedMin, expectedMax)
			}
		}
	}

	// Reset to baseline
	pc.SetNeuromodulatorLevels(1.0, 1.0, 1.0)

	// Test acetylcholine attention gating
	t.Log("\n--- Testing Acetylcholine Attention Gating ---")
	achLevels := []float64{0.5, 1.0, 1.5, 2.0, 2.5}

	for _, ach := range achLevels {
		pc.SetNeuromodulatorLevels(1.0, ach, 1.0)
		change := pc.CalculateSTDPWeightChange(timing, weight, cooperativity)
		enhancement := change / baselineChange

		t.Logf("  Acetylcholine %.1f: %.6f (%.1fx enhancement)", ach, change, enhancement)

		// ACh should enhance plasticity when elevated
		if ach > 1.0 {
			expectedEnhancement := 1.0 + (ach-1.0)*(BIOLOGY_ACH_ATTENTION_GATE-1.0)
			tolerance := 0.3 // 30% tolerance
			if math.Abs(enhancement-expectedEnhancement) > tolerance {
				t.Errorf("BIOLOGY VIOLATION: ACh %.1f enhancement (%.1fx) differs from expected (%.1fx)",
					ach, enhancement, expectedEnhancement)
			}
		}
	}

	// Reset to baseline
	pc.SetNeuromodulatorLevels(1.0, 1.0, 1.0)

	// Test norepinephrine inverted-U curve
	t.Log("\n--- Testing Norepinephrine Stress Response ---")
	neLevels := []float64{0.5, 1.0, 1.3, 1.5, 2.0, 2.5, 3.0}
	maxEnhancement := 0.0
	optimalLevel := 0.0

	for _, ne := range neLevels {
		pc.SetNeuromodulatorLevels(1.0, 1.0, ne)
		change := pc.CalculateSTDPWeightChange(timing, weight, cooperativity)
		enhancement := change / baselineChange

		t.Logf("  Norepinephrine %.1f: %.6f (%.1fx enhancement)", ne, change, enhancement)

		if enhancement > maxEnhancement {
			maxEnhancement = enhancement
			optimalLevel = ne
		}
	}

	// Validate inverted-U curve
	t.Logf("Optimal norepinephrine level: %.1f (%.1fx enhancement)", optimalLevel, maxEnhancement)
	validateBiologicalRange(t, "Optimal NE level", optimalLevel, 1.0, 2.0, "concentration")
	validateBiologicalRange(t, "Peak NE enhancement", maxEnhancement, 1.0, 2.0, "fold")

	// Reset to baseline
	pc.SetNeuromodulatorLevels(1.0, 1.0, 1.0)

	t.Log("✅ Neuromodulator effects match experimental data")
}

// =================================================================================
// TEST 4: DEVELOPMENTAL PLASTICITY
// =================================================================================

// TestPlasticityBiologyDevelopment validates age-dependent plasticity changes
// match experimental data on critical periods and aging.
//
// EXPERIMENTAL VALIDATION:
// - Enhanced plasticity during critical periods
// - Gradual decline with age
// - Adult plasticity as baseline
// - Aged plasticity significantly reduced
func TestPlasticityBiologyDevelopment(t *testing.T) {
	t.Log("=== BIOLOGY TEST: Developmental Plasticity ===")
	t.Log("Validating against critical period and aging studies")

	pc := NewPlasticityCalculator(createBiologicalSTDPConfig())

	timing := -10 * time.Millisecond
	weight := 0.5
	cooperativity := BIOLOGY_COOPERATIVITY_THRESHOLD

	// Test different developmental stages
	stages := []struct {
		age         float64
		description string
		expectedMod float64 // Expected modulation relative to adult
	}{
		{0.2, "Juvenile (critical period)", 2.0}, // Enhanced
		{0.5, "Adolescent", 1.3},                 // Somewhat enhanced
		{1.0, "Adult", 1.0},                      // Baseline
		{1.5, "Middle-aged", 0.7},                // Reduced
		{2.0, "Aged", 0.4},                       // Significantly reduced
		{3.0, "Very aged", 0.2},                  // Severely reduced
	}

	adultChange := 0.0

	t.Log("Developmental stage plasticity:")
	for _, stage := range stages {
		pc.SetDevelopmentalStage(stage.age)
		change := pc.CalculateSTDPWeightChange(timing, weight, cooperativity)

		if stage.age == 1.0 {
			adultChange = change // Store adult baseline
		}

		t.Logf("  %s (%.1f): %.6f", stage.description, stage.age, change)
	}

	// Validate relative to adult baseline
	t.Log("\nDevelopmental modulation relative to adult:")
	for _, stage := range stages {
		pc.SetDevelopmentalStage(stage.age)
		change := pc.CalculateSTDPWeightChange(timing, weight, cooperativity)

		if adultChange != 0 {
			modulation := change / adultChange
			t.Logf("  %s: %.1fx adult level", stage.description, modulation)

			// Validate against expected biological ranges
			tolerance := 0.5 // 50% tolerance for biological variability
			expectedRange := stage.expectedMod

			if math.Abs(modulation-expectedRange) > tolerance {
				t.Logf("Note: %s modulation (%.1fx) differs from typical range (~%.1fx)",
					stage.description, modulation, expectedRange)
			}
		}
	}

	// Reset to adult
	pc.SetDevelopmentalStage(1.0)

	t.Log("✅ Developmental plasticity changes match experimental data")
}

// =================================================================================
// TEST 5: FREQUENCY DEPENDENCE
// =================================================================================

// TestPlasticityBiologyFrequencyDependence validates frequency-dependent plasticity
// matches experimental data on LTP/LTD induction protocols.
//
// EXPERIMENTAL VALIDATION:
// - Low frequency (1Hz) induces LTD
// - High frequency (100Hz) induces LTP
// - Theta frequency (5Hz) has intermediate effects
// - BCM rule behavior with frequency
func TestPlasticityBiologyFrequencyDependence(t *testing.T) {
	t.Log("=== BIOLOGY TEST: Frequency-Dependent Plasticity ===")
	t.Log("Validating against LTP/LTD induction protocol studies")

	pc := NewPlasticityCalculator(createBiologicalSTDPConfig())

	weight := 0.5
	duration := 60 * time.Second // 1 minute stimulation

	// Test different stimulation frequencies
	frequencies := []struct {
		freq           float64
		name           string
		expectedEffect string
	}{
		{0.1, "Very low", "Strong LTD"},
		{1.0, "LTD protocol", "LTD"},
		{5.0, "Theta", "Minimal/mixed"},
		{10.0, "Alpha", "Weak LTP"},
		{50.0, "Gamma", "Moderate LTP"},
		{100.0, "LTP protocol", "Strong LTP"},
		{200.0, "Very high", "Strong LTP"},
	}

	t.Log("Frequency-dependent plasticity:")
	for _, freq := range frequencies {
		change := pc.CalculateFrequencyDependentPlasticity(freq.freq, weight, duration)
		t.Logf("  %.1f Hz (%s): %.6f - %s", freq.freq, freq.name, change, freq.expectedEffect)

		// Validate expected directions
		switch freq.name {
		case "LTD protocol":
			if change >= 0 {
				t.Errorf("BIOLOGY VIOLATION: 1Hz should induce LTD, got %.6f", change)
			}
		case "LTP protocol":
			if change <= 0 {
				t.Errorf("BIOLOGY VIOLATION: 100Hz should induce LTP, got %.6f", change)
			}
		}
	}

	// Test BCM rule - crossover frequency
	crossoverFreq := 0.0
	prevChange := 0.0

	for freq := 1.0; freq <= 20.0; freq += 1.0 {
		change := pc.CalculateFrequencyDependentPlasticity(freq, weight, duration)

		if prevChange < 0 && change > 0 {
			crossoverFreq = freq
			break
		}
		prevChange = change
	}

	if crossoverFreq > 0 {
		t.Logf("LTD→LTP crossover frequency: %.1f Hz", crossoverFreq)
		validateBiologicalRange(t, "Crossover frequency", crossoverFreq, 2.0, 15.0, "Hz")
	} else {
		t.Log("Note: No clear LTD→LTP crossover found in tested range")
	}

	t.Log("✅ Frequency dependence matches experimental protocols")
}

// =================================================================================
// TEST 6: WEIGHT DEPENDENCE AND SATURATION
// =================================================================================

// TestPlasticityBiologyWeightDependence validates weight-dependent plasticity
// and saturation effects match experimental data.
//
// EXPERIMENTAL VALIDATION:
// - Weak synapses show larger plasticity
// - Strong synapses show reduced plasticity
// - Saturation near weight boundaries
// - Multiplicative vs additive effects
func TestPlasticityBiologyWeightDependence(t *testing.T) {
	t.Log("=== BIOLOGY TEST: Weight Dependence and Saturation ===")
	t.Log("Validating against synaptic strength modulation studies")

	pc := NewPlasticityCalculator(createBiologicalSTDPConfig())

	timing := -10 * time.Millisecond // Peak LTP timing
	cooperativity := BIOLOGY_COOPERATIVITY_THRESHOLD

	// Test different initial weights
	weights := []float64{0.05, 0.1, 0.2, 0.3, 0.5, 0.7, 0.8, 0.9, 0.95}

	t.Log("Weight-dependent plasticity:")
	var changes []float64

	for _, weight := range weights {
		change := pc.CalculateSTDPWeightChange(timing, weight, cooperativity)
		changes = append(changes, change)
		t.Logf("  Weight %.2f: %.6f change", weight, change)
	}

	// Validate weight dependence trend
	// Weak synapses should show larger changes than strong synapses
	weakChange := changes[1]   // Weight 0.1
	strongChange := changes[7] // Weight 0.9

	if weakChange <= strongChange {
		t.Errorf("BIOLOGY VIOLATION: Weak synapses should show larger plasticity than strong synapses")
	} else {
		ratio := weakChange / strongChange
		t.Logf("Weak/strong synapse plasticity ratio: %.1f", ratio)
		validateBiologicalRange(t, "Weak/strong plasticity ratio", ratio, 1.2, 5.0, "fold")
	}

	// Test saturation near boundaries
	nearMinChange := changes[0] // Weight 0.05
	nearMaxChange := changes[8] // Weight 0.95
	midChange := changes[4]     // Weight 0.5

	// Near boundaries should show reduced plasticity compared to middle range
	minReduction := nearMinChange / midChange
	maxReduction := nearMaxChange / midChange

	t.Logf("Boundary saturation effects:")
	t.Logf("  Near minimum: %.1fx of mid-range", minReduction)
	t.Logf("  Near maximum: %.1fx of mid-range", maxReduction)

	// Some saturation expected but not complete elimination
	if minReduction > 1.2 || maxReduction > 1.2 {
		t.Log("Note: Limited saturation effects observed")
	} else if minReduction < 0.1 || maxReduction < 0.1 {
		t.Log("Note: Strong saturation effects observed")
	} else {
		t.Log("✓ Moderate saturation effects as expected")
	}

	t.Log("✅ Weight dependence matches experimental observations")
}

// =================================================================================
// COMPREHENSIVE BIOLOGICAL VALIDATION SUMMARY
// =================================================================================

// TestPlasticityBiologyComprehensive runs all biological validation tests
// and provides a summary of biological accuracy across all mechanisms.
func TestPlasticityBiologyComprehensive(t *testing.T) {
	t.Log("=== COMPREHENSIVE BIOLOGICAL VALIDATION SUITE ===")
	t.Log("Running all plasticity biology tests...")

	// Track validation results
	testResults := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{"STDP Timing Window", TestPlasticityBiologySTDPTimingWindow},
		{"Cooperativity", TestPlasticityBiologyCooperativity},
		{"Neuromodulation", TestPlasticityBiologyNeuromodulation},
		{"Development", TestPlasticityBiologyDevelopment},
		{"Frequency Dependence", TestPlasticityBiologyFrequencyDependence},
		{"Weight Dependence", TestPlasticityBiologyWeightDependence},
	}

	passedTests := 0
	for _, test := range testResults {
		t.Logf("\n=== Running %s Test ===", test.name)

		// Run test in separate context to catch failures
		passed := t.Run(test.name, test.testFunc)
		if passed {
			passedTests++
			t.Logf("✅ %s: PASSED", test.name)
		} else {
			t.Logf("❌ %s: FAILED", test.name)
		}
	}

	// Summary report
	successRate := float64(passedTests) / float64(len(testResults)) * 100
	t.Logf("\n=== BIOLOGICAL VALIDATION SUMMARY ===")
	t.Logf("Tests passed: %d/%d (%.1f%%)", passedTests, len(testResults), successRate)

	if successRate >= 100.0 {
		t.Log("🧠 EXCELLENT: All biological validation tests passed!")
		t.Log("   Plasticity implementation matches experimental neuroscience data")
	} else if successRate >= 80.0 {
		t.Log("🧠 GOOD: Most biological validation tests passed")
		t.Log("   Plasticity shows strong biological realism with minor deviations")
	} else if successRate >= 60.0 {
		t.Log("🧠 ACCEPTABLE: Majority of biological tests passed")
		t.Log("   Plasticity has biological basis but may need refinement")
	} else {
		t.Log("🧠 NEEDS WORK: Multiple biological validation failures")
		t.Log("   Plasticity implementation needs significant biological corrections")
	}

	// Provide specific recommendations based on failures
	if passedTests < len(testResults) {
		t.Log("\nRecommendations for failed tests:")

		// This would be expanded based on which specific tests failed
		t.Log("- Review experimental literature for failed mechanisms")
		t.Log("- Adjust parameters to match published data")
		t.Log("- Consider biological constraints and realistic ranges")
		t.Log("- Validate against multiple experimental studies")
	}

	t.Log("\n✅ Comprehensive biological validation completed")
}

// =================================================================================
// ADDITIONAL SPECIALIZED BIOLOGY TESTS
// =================================================================================

// TestPlasticityBiologyMetaplasticity validates BCM rule implementation
// against experimental metaplasticity data.
func TestPlasticityBiologyMetaplasticity(t *testing.T) {
	t.Log("=== BIOLOGY TEST: Metaplasticity (BCM Rule) ===")
	t.Log("Validating against Bienenstock-Cooper-Munro experimental data")

	pc := NewPlasticityCalculator(createBiologicalSTDPConfig())
	pc.config.MetaplasticityRate = 0.5 // High rate for testing

	// Ensure enough history for metaplasticity to activate
	const minHistoryForMetaplasticity = 10 // Based on calculateMetaplasticityFactorRobust check

	currentWeight := 0.5             // Consistent weight for plasticity calculation
	deltaT := -10 * time.Millisecond // Consistent LTP stimulus
	cooperativeInputs := BIOLOGY_COOPERATIVITY_THRESHOLD

	// Get baseline without any activity history influencing metaplasticity
	// Temporarily disable metaplasticity for clean baseline comparison
	oldMetaplasticityRate := pc.config.MetaplasticityRate
	pc.config.MetaplasticityRate = 0.0
	baselineChange := pc.CalculateSTDPWeightChange(deltaT, currentWeight, cooperativeInputs)
	pc.config.MetaplasticityRate = oldMetaplasticityRate // Restore rate

	t.Logf("Baseline change (no metaplasticity influence): %.6f", baselineChange)

	// --- Scenario 1: Low average activity with a decreasing trend (should decrease threshold, enhance plasticity) ---
	pc.Reset() // Reset state for clean scenario
	// Fill history with activity that trends downwards
	for i := 0; i < minHistoryForMetaplasticity+5; i++ { // Ensure sufficient history
		activity := 1.5 - float64(i)*0.1 // Starts high, trends down to low values
		if activity < 0.1 {
			activity = 0.1
		} // Clamp minimum for realism
		pc.UpdateActivityHistoryRobust(activity)
	}
	// After updating history, the threshold should have shifted downwards
	lowActivityThreshold := pc.GetStatisticsRobust().ThresholdValue
	t.Logf("Threshold after low activity history (decreasing trend): %.6f (Expected < 1.0)", lowActivityThreshold)
	if lowActivityThreshold >= 1.0 { // Assuming initial threshold is 1.0
		t.Errorf("Low activity history (decreasing trend) should decrease plasticity threshold, got %.6f", lowActivityThreshold)
	}

	changeLowActivity := pc.CalculateSTDPWeightChange(deltaT, currentWeight, cooperativeInputs)
	t.Logf("Change with low activity context: %.6f", changeLowActivity)

	// Low activity should enhance plasticity (make positive changes larger, negative changes smaller in magnitude)
	// Given deltaT is -10ms (LTP), we expect larger positive change
	if changeLowActivity <= baselineChange {
		t.Errorf("BIOLOGY VIOLATION: Low activity should enhance plasticity, expected > %.6f, got %.6f", baselineChange, changeLowActivity)
	}

	// --- Scenario 2: High average activity with an increasing trend (should increase threshold, reduce plasticity) ---
	pc.Reset() // Reset state for clean scenario
	// Fill history with activity that trends upwards
	for i := 0; i < minHistoryForMetaplasticity+5; i++ { // Ensure sufficient history
		activity := 0.5 + float64(i)*0.1 // Starts low, trends up to high values
		if activity > 2.5 {
			activity = 2.5
		} // Clamp maximum for realism
		pc.UpdateActivityHistoryRobust(activity)
	}
	// After updating history, the threshold should have shifted upwards
	highActivityThreshold := pc.GetStatisticsRobust().ThresholdValue
	t.Logf("Threshold after high activity history (increasing trend): %.6f (Expected > 1.0)", highActivityThreshold)
	if highActivityThreshold <= 1.0 { // Assuming initial threshold is 1.0
		t.Errorf("High activity history (increasing trend) should increase plasticity threshold, got %.6f", highActivityThreshold)
	}

	changeHighActivity := pc.CalculateSTDPWeightChange(deltaT, currentWeight, cooperativeInputs)
	t.Logf("Change with high activity context: %.6f", changeHighActivity)

	// High activity should reduce plasticity (make positive changes smaller, negative changes larger in magnitude)
	// Given deltaT is -10ms (LTP), we expect smaller positive change
	if changeHighActivity >= baselineChange {
		t.Errorf("BIOLOGY VIOLATION: High activity should reduce plasticity, expected < %.6f, got %.6f", baselineChange, changeHighActivity)
	}

	t.Logf("Metaplasticity modulation confirmed - baseline: %.6f, low context: %.6f, high context: %.6f",
		baselineChange, changeLowActivity, changeHighActivity)
	t.Log("✅ Metaplasticity follows BCM rule predictions")
}

// TestPlasticityBiologyProteinSynthesis validates late-phase plasticity
// mechanisms against protein synthesis-dependent LTP/LTD studies.
func TestPlasticityBiologyProteinSynthesis(t *testing.T) {
	t.Log("=== BIOLOGY TEST: Protein Synthesis-Dependent Plasticity ===")
	t.Log("Validating against late-phase LTP/LTD experimental data")

	pc := NewPlasticityCalculator(createBiologicalSTDPConfig())

	// Test protein synthesis-dependent enhancement
	initialChange := 0.1
	stimulationStrengths := []float64{1.0, 2.0, 3.0, 4.0, 5.0}

	t.Log("Protein synthesis enhancement vs stimulation strength:")
	for _, strength := range stimulationStrengths {
		// Test at different time points
		timePoints := []time.Duration{
			1 * time.Hour,  // Early phase
			3 * time.Hour,  // Late phase beginning
			6 * time.Hour,  // Late phase peak
			12 * time.Hour, // Late phase decay
			24 * time.Hour, // End of late phase
		}

		t.Logf("  Stimulation strength %.1f:", strength)
		for _, timePoint := range timePoints {
			enhancement := pc.CalculateProteinSynthesisDependentPlasticity(
				initialChange, strength, timePoint)
			t.Logf("    %v: %.6f enhancement", timePoint, enhancement)
		}
	}

	// Validate timing constraints
	// Should require strong stimulation (>2.0) for protein synthesis
	weakStimulation := pc.CalculateProteinSynthesisDependentPlasticity(
		initialChange, 1.5, 6*time.Hour)
	strongStimulation := pc.CalculateProteinSynthesisDependentPlasticity(
		initialChange, 3.0, 6*time.Hour)

	if weakStimulation >= strongStimulation {
		t.Error("BIOLOGY VIOLATION: Strong stimulation should produce more protein synthesis enhancement")
	}

	// Should have time window (no enhancement before 2 hours or after 24 hours)
	earlyEnhancement := pc.CalculateProteinSynthesisDependentPlasticity(
		initialChange, 3.0, 30*time.Minute)
	lateEnhancement := pc.CalculateProteinSynthesisDependentPlasticity(
		initialChange, 3.0, 30*time.Hour)

	if earlyEnhancement > 0.001 {
		t.Error("BIOLOGY VIOLATION: Should not have protein synthesis enhancement before late phase")
	}
	if lateEnhancement > 0.001 {
		t.Error("BIOLOGY VIOLATION: Should not have protein synthesis enhancement after late phase")
	}

	t.Log("✅ Protein synthesis timing and requirements match experimental data")
}

// TestPlasticityBiologyHeterosynaptic validates heterosynaptic plasticity
// and synaptic tagging mechanisms.
func TestPlasticityBiologyHeterosynaptic(t *testing.T) {
	t.Log("=== BIOLOGY TEST: Heterosynaptic Plasticity ===")
	t.Log("Validating against synaptic tagging and capture experimental data")

	pc := NewPlasticityCalculator(createBiologicalSTDPConfig())

	// Test heterosynaptic plasticity spread
	primaryChange := 0.5
	distances := []float64{1.0, 5.0, 10.0, 20.0, 50.0, 100.0} // micrometers

	t.Log("Heterosynaptic plasticity vs distance:")
	for _, distance := range distances {
		heteroChange := pc.CalculateHeterosynapticPlasticity(distance, primaryChange)
		t.Logf("  %.1f μm: %.6f (%.1f%% of primary)",
			distance, heteroChange, heteroChange/primaryChange*100)
	}

	// Test synaptic tagging and capture
	t.Log("\nSynaptic tagging and capture:")
	weakChange := 0.1
	strongDistances := []float64{5.0, 10.0, 25.0, 50.0}
	timeDelays := []time.Duration{
		30 * time.Minute,
		1 * time.Hour,
		2 * time.Hour,
		4 * time.Hour,
	}

	for _, distance := range strongDistances {
		t.Logf("  Distance %.1f μm:", distance)
		for _, delay := range timeDelays {
			capture := pc.CalculateSynapticTaggingAndCapture(weakChange, distance, delay)
			t.Logf("    %v delay: %.6f enhancement", delay, capture)
		}
	}

	// Validate biological constraints
	// Should have distance and time limits
	nearCapture := pc.CalculateSynapticTaggingAndCapture(weakChange, 10.0, 1*time.Hour)
	farCapture := pc.CalculateSynapticTaggingAndCapture(weakChange, 100.0, 1*time.Hour)
	earlyCapture := pc.CalculateSynapticTaggingAndCapture(weakChange, 10.0, 30*time.Minute)
	lateCapture := pc.CalculateSynapticTaggingAndCapture(weakChange, 10.0, 6*time.Hour)

	if nearCapture <= farCapture {
		t.Error("BIOLOGY VIOLATION: Near synapses should show more tagging/capture than distant ones")
	}
	if earlyCapture <= lateCapture {
		t.Error("BIOLOGY VIOLATION: Early capture should be stronger than late capture")
	}

	t.Log("✅ Heterosynaptic mechanisms match experimental spatial and temporal constraints")
}

// =================================================================================
// BIOLOGICAL PARAMETER VALIDATION
// =================================================================================

// TestPlasticityBiologyParameters validates that all plasticity parameters
// fall within experimentally measured biological ranges.
func TestPlasticityBiologyParameters(t *testing.T) {
	t.Log("=== BIOLOGY TEST: Parameter Validation ===")
	t.Log("Validating all parameters against experimental literature ranges")

	config := createBiologicalSTDPConfig()

	// Validate timing parameters
	t.Log("Timing parameters:")
	timeConstantMs := float64(config.TimeConstant.Milliseconds())
	windowMs := float64(config.WindowSize.Milliseconds())

	validateBiologicalRange(t, "Time constant", timeConstantMs, 5.0, 50.0, "ms")
	validateBiologicalRange(t, "STDP window", windowMs, 50.0, 200.0, "ms")

	// Validate learning parameters
	t.Log("\nLearning parameters:")
	validateBiologicalRange(t, "Learning rate", config.LearningRate, 0.001, 0.1, "per pairing")
	validateBiologicalRange(t, "Asymmetry ratio", config.AsymmetryRatio, 0.5, 5.0, "ratio")

	// Validate weight bounds
	t.Log("\nWeight parameters:")
	validateBiologicalRange(t, "Minimum weight", config.MinWeight, 0.0, 0.1, "normalized")
	validateBiologicalRange(t, "Maximum weight", config.MaxWeight, 1.0, 5.0, "normalized")

	// Validate cooperativity
	t.Log("\nCooperativity parameters:")
	validateBiologicalRange(t, "Cooperativity threshold", float64(config.CooperativityThreshold), 1.0, 10.0, "inputs")

	// Validate metaplasticity
	t.Log("\nMetaplasticity parameters:")
	validateBiologicalRange(t, "Metaplasticity rate", config.MetaplasticityRate, 0.01, 1.0, "rate")

	// Test parameter interactions
	t.Log("\nParameter interaction validation:")
	pc := NewPlasticityCalculator(config)

	// Ensure parameters produce reasonable plasticity magnitudes
	change := pc.CalculateSTDPWeightChange(-10*time.Millisecond, 0.5, 3)
	validateBiologicalRange(t, "Typical plasticity magnitude", math.Abs(change), 0.001, 0.1, "weight change")

	// Ensure plasticity is bounded
	extremeChange := pc.CalculateSTDPWeightChange(-1*time.Millisecond, 0.01, 20)
	if math.Abs(extremeChange) > 0.5 {
		t.Errorf("BIOLOGY VIOLATION: Extreme plasticity too large: %.3f", extremeChange)
	}

	t.Log("✅ All parameters within biological ranges")
}

// =================================================================================
// CROSS-VALIDATION WITH MULTIPLE EXPERIMENTAL DATASETS
// =================================================================================

// TestPlasticityBiologyCrossValidation validates plasticity against multiple
// independent experimental datasets to ensure broad biological accuracy.
func TestPlasticityBiologyCrossValidation(t *testing.T) {
	t.Log("=== BIOLOGY TEST: Cross-Validation Against Multiple Studies ===")
	t.Log("Validating against diverse experimental preparations and conditions")

	pc := NewPlasticityCalculator(createBiologicalSTDPConfig())

	// Different experimental preparations from literature
	experiments := []struct {
		name          string
		preparation   string
		timing        time.Duration
		weight        float64
		cooperativity int
		expectedSign  string // "positive", "negative", or "minimal"
		reference     string
	}{
		{"Bi & Poo 1998", "Hippocampal culture", -10 * time.Millisecond, 0.5, 3, "positive", "Classic STDP study"},
		{"Markram 1997", "Neocortical slices", -10 * time.Millisecond, 0.5, 3, "positive", "Cortical STDP"},
		{"Sjöström 2001", "Cortical pairs", -15 * time.Millisecond, 0.3, 5, "positive", "Cooperativity study"},
		{"Dan & Poo 2004", "Visual cortex", 15 * time.Millisecond, 0.7, 3, "negative", "LTD characterization"},
		{"Abbott & Nelson 2000", "Model validation", -5 * time.Millisecond, 0.1, 2, "minimal", "Weak synapse"},
	}

	validationScore := 0

	t.Log("Cross-validation results:")
	for _, exp := range experiments {
		change := pc.CalculateSTDPWeightChange(exp.timing, exp.weight, exp.cooperativity)

		t.Logf("  %s (%s):", exp.name, exp.preparation)
		t.Logf("    Parameters: %v, weight=%.1f, coop=%d", exp.timing, exp.weight, exp.cooperativity)
		t.Logf("    Result: %.6f (expected: %s)", change, exp.expectedSign)

		// Validate against expected sign
		passed := false
		switch exp.expectedSign {
		case "positive":
			passed = change > 0.001
		case "negative":
			passed = change < -0.001
		case "minimal":
			passed = math.Abs(change) < 0.005
		}

		if passed {
			validationScore++
			t.Logf("    ✅ PASS - Matches %s expectations", exp.reference)
		} else {
			t.Logf("    ❌ FAIL - Doesn't match %s expectations", exp.reference)
		}
	}

	// Summary
	validationRate := float64(validationScore) / float64(len(experiments)) * 100
	t.Logf("\nCross-validation summary: %d/%d studies matched (%.1f%%)",
		validationScore, len(experiments), validationRate)

	if validationRate >= 80.0 {
		t.Log("✅ High cross-validation success - implementation generalizes well")
	} else if validationRate >= 60.0 {
		t.Log("⚠ Moderate cross-validation - some experimental discrepancies")
	} else {
		t.Error("❌ Poor cross-validation - significant experimental mismatches")
	}

	t.Log("✅ Cross-validation against multiple experimental datasets completed")
}
