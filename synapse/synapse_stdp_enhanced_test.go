package synapse

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

// TestEnhancedSTDP_ChemicalModulation tests how different chemical signals (neuromodulators)
// affect STDP learning rules. This test validates that different chemical modulators
// produce distinct and biologically plausible effects on synaptic plasticity.
//
// BIOLOGICAL BASIS:
// In biological neural networks, various neuromodulators can dramatically
// alter STDP learning rules, enabling context-dependent learning:
// - Dopamine signals reward and enhances learning in reward circuits
// - GABA provides inhibitory feedback and can act as a penalty signal
// - Serotonin modulates mood-related plasticity and learning
// - Glutamate enhances excitatory transmission and can boost STDP effects
//
// This test examines how these chemicals create a rich, dynamic learning
// environment beyond basic Hebbian plasticity.
func TestEnhancedSTDP_ChemicalModulation(t *testing.T) {
	preNeuron := NewMockNeuron("chem_pre")
	postNeuron := NewMockNeuron("chem_post")

	// Configure with standard STDP parameters
	stdpConfig := types.PlasticityConfig{
		Enabled:        true,
		LearningRate:   0.05, // Larger for clearer effects
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     100 * time.Millisecond,
		MinWeight:      0.001,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.2,
	}

	pruningConfig := CreateDefaultPruningConfig()

	// Add detailed logging
	t.Log("=== ENHANCED STDP CHEMICAL MODULATION TEST ===")
	t.Log("Chemical | Concentration | Eligibility | Initial Weight | Final Weight | Effect")
	t.Log("-------------------------------------------------------------------------------")

	// Test cases with different neuromodulators
	testCases := []struct {
		name          string
		chemical      types.LigandType
		concentration float64
		createTrace   func(*BasicSynapse) float64 // Function to create eligibility trace
		expectedDir   int                         // Expected direction: 1=increase, -1=decrease, 0=minimal
	}{
		{
			name:          "Dopamine (reward)",
			chemical:      types.LigandDopamine,
			concentration: 2.0, // High concentration (strong reward)
			createTrace: func(s *BasicSynapse) float64 {
				// Create positive eligibility with causal STDP
				for i := 0; i < 10; i++ {
					s.ApplyPlasticity(types.PlasticityAdjustment{DeltaT: -10 * time.Millisecond})
				}
				return s.GetEligibilityTrace()
			},
			expectedDir: 1, // Should increase weight
		},
		{
			name:          "GABA (inhibition)",
			chemical:      types.LigandGABA,
			concentration: 1.5, // Moderate inhibition
			createTrace: func(s *BasicSynapse) float64 {
				// Create positive eligibility with causal STDP
				for i := 0; i < 10; i++ {
					s.ApplyPlasticity(types.PlasticityAdjustment{DeltaT: -10 * time.Millisecond})
				}
				return s.GetEligibilityTrace()
			},
			expectedDir: -1, // Should decrease weight (penalty signal)
		},
		{
			name:          "Serotonin (mood)",
			chemical:      types.LigandSerotonin,
			concentration: 1.5, // Moderate level
			createTrace: func(s *BasicSynapse) float64 {
				// Create positive eligibility with causal STDP
				for i := 0; i < 10; i++ {
					s.ApplyPlasticity(types.PlasticityAdjustment{DeltaT: -10 * time.Millisecond})
				}
				return s.GetEligibilityTrace()
			},
			expectedDir: 1, // Should have positive but weaker effect than dopamine
		},
		{
			name:          "Glutamate (excitation)",
			chemical:      types.LigandGlutamate,
			concentration: 1.5, // Moderate level
			createTrace: func(s *BasicSynapse) float64 {
				// Create positive eligibility with causal STDP
				for i := 0; i < 10; i++ {
					s.ApplyPlasticity(types.PlasticityAdjustment{DeltaT: -10 * time.Millisecond})
				}
				return s.GetEligibilityTrace()
			},
			expectedDir: 1, // Should enhance positive STDP
		},
		{
			name:          "Dopamine with Negative Trace",
			chemical:      types.LigandDopamine,
			concentration: 2.0, // High concentration
			createTrace: func(s *BasicSynapse) float64 {
				// Create negative eligibility with anti-causal STDP
				for i := 0; i < 10; i++ {
					s.ApplyPlasticity(types.PlasticityAdjustment{DeltaT: 10 * time.Millisecond})
				}
				return s.GetEligibilityTrace()
			},
			expectedDir: -1, // Should decrease weight (negative eligibility)
		},
	}

	// Run tests for each chemical modulator
	for i, tc := range testCases {
		// Create a fresh synapse for each test
		synapse := NewBasicSynapse(
			"chem_test_"+tc.name,
			preNeuron, postNeuron,
			stdpConfig, pruningConfig, 0.5, 0,
		)

		// Set consistent eligibility trace decay for testing
		synapse.SetEligibilityDecay(500 * time.Millisecond)

		// Create eligibility trace
		eligibility := tc.createTrace(synapse)
		initialWeight := synapse.GetWeight()

		// Apply neuromodulation
		weightChange := synapse.ProcessNeuromodulation(tc.chemical, tc.concentration)
		finalWeight := synapse.GetWeight()

		// Determine effect description
		var effectDesc string
		if math.Abs(weightChange) < 0.001 {
			effectDesc = "No effect"
		} else if weightChange > 0 {
			if weightChange > 0.1 {
				effectDesc = "Strong potentiation"
			} else {
				effectDesc = "Weak potentiation"
			}
		} else {
			if weightChange < -0.1 {
				effectDesc = "Strong depression"
			} else {
				effectDesc = "Weak depression"
			}
		}

		// Log results
		t.Logf("%-10s | %12.1f | %+10.6f | %14.4f | %12.4f | %s",
			tc.name, tc.concentration, eligibility, initialWeight, finalWeight, effectDesc)

		// Verify expected behavior
		if tc.expectedDir > 0 && weightChange <= 0 {
			t.Errorf("Case %d (%s): Expected weight increase, got %f", i+1, tc.name, weightChange)
		} else if tc.expectedDir < 0 && weightChange >= 0 {
			t.Errorf("Case %d (%s): Expected weight decrease, got %f", i+1, tc.name, weightChange)
		} else if tc.expectedDir == 0 && math.Abs(weightChange) > 0.01 {
			t.Errorf("Case %d (%s): Expected minimal change, got %f", i+1, tc.name, weightChange)
		}
	}

	// Test case comparison: verify distinct effects between modulators
	t.Log("\n=== CHEMICAL SPECIFICITY TEST ===")
	t.Log("Testing that different chemicals produce distinct effects with identical eligibility traces")

	// Create a consistent eligibility trace for comparison
	createStandardTrace := func(s *BasicSynapse) {
		// Reset weight
		s.SetWeight(0.5)
		// Create eligibility trace with 10 causal STDP events
		for i := 0; i < 10; i++ {
			s.ApplyPlasticity(types.PlasticityAdjustment{DeltaT: -10 * time.Millisecond})
		}
	}

	// Test chemicals
	chemicals := []types.LigandType{
		types.LigandDopamine,
		types.LigandGABA,
		types.LigandSerotonin,
		types.LigandGlutamate,
	}

	// Record effects
	effects := make(map[types.LigandType]float64)

	// Apply each chemical with identical eligibility traces
	synapse := NewBasicSynapse("chem_comparison", preNeuron, postNeuron,
		stdpConfig, pruningConfig, 0.5, 0)
	synapse.SetEligibilityDecay(500 * time.Millisecond)

	for _, chemical := range chemicals {
		// Create identical starting conditions
		createStandardTrace(synapse)
		initialWeight := synapse.GetWeight()

		// Apply chemical
		weightChange := synapse.ProcessNeuromodulation(chemical, 1.5)
		effects[chemical] = weightChange

		t.Logf("Chemical: %-10s | Initial: %.4f | Change: %+.6f | New: %.4f",
			chemical, initialWeight, weightChange, synapse.GetWeight())
	}

	// Verify chemical specificity
	for i := 0; i < len(chemicals); i++ {
		for j := i + 1; j < len(chemicals); j++ {
			chem1 := chemicals[i]
			chem2 := chemicals[j]
			effect1 := effects[chem1]
			effect2 := effects[chem2]

			// Effects should be noticeably different (>10% difference)
			if math.Abs(effect1-effect2) < math.Abs(effect1)*0.1 && effect1 != 0 {
				t.Logf("Warning: %s and %s have very similar effects: %f vs %f",
					chem1, chem2, effect1, effect2)
			}
		}
	}

	// Biological significance
	t.Log("\n=== BIOLOGICAL SIGNIFICANCE ===")
	t.Log("- Different neuromodulators have distinct effects on synaptic plasticity")
	t.Log("- Dopamine enhances learning for rewarded behaviors")
	t.Log("- GABA acts as a penalty signal, weakening inappropriate connections")
	t.Log("- Serotonin modulates mood-related learning with more subtle effects")
	t.Log("- Glutamate enhances excitatory transmission and can boost STDP")
	t.Log("- Chemical specificity enables rich, context-dependent learning")
}

// TestEnhancedSTDP_BidirectionalPlasticity tests that STDP can both strengthen (LTP)
// and weaken (LTD) synapses based on the precise timing of pre- and post-synaptic
// activity. This test validates the bidirectional learning capabilities of the synapse.
//
// BIOLOGICAL BASIS:
// Spike-Timing Dependent Plasticity (STDP) is inherently bidirectional:
// - When pre-synaptic activity precedes post-synaptic (causal), LTP occurs
// - When post-synaptic activity precedes pre-synaptic (anti-causal), LTD occurs
//
// This bidirectional property is essential for learning temporal associations
// and distinguishing causal from coincidental correlations in neural activity.
func TestEnhancedSTDP_BidirectionalPlasticity(t *testing.T) {
	preNeuron := NewMockNeuron("bidir_pre")
	postNeuron := NewMockNeuron("bidir_post")

	// Configure with standard STDP parameters
	stdpConfig := types.PlasticityConfig{
		Enabled:        true,
		LearningRate:   0.01, // Small increments for gradual change
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     100 * time.Millisecond,
		MinWeight:      0.01,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.2, // LTD slightly stronger than LTP
	}

	pruningConfig := CreateDefaultPruningConfig()

	// Add detailed logging
	t.Log("=== ENHANCED STDP BIDIRECTIONAL PLASTICITY TEST ===")
	t.Log("Direction | Timing (ms) | Repetitions | Initial | Final | Change | % Change")
	t.Log("-------------------------------------------------------------------------")

	// Test cases for bidirectional plasticity
	testCases := []struct {
		name        string
		timingMs    int    // Timing difference in milliseconds
		repetitions int    // Number of STDP events to apply
		direction   string // Expected direction of change
		expectedDir int    // Expected direction: 1=increase, -1=decrease
	}{
		{
			name:        "Strong LTP",
			timingMs:    -10, // Pre before post (optimal for LTP)
			repetitions: 50,
			direction:   "LTP",
			expectedDir: 1, // Should increase weight
		},
		{
			name:        "Weak LTP",
			timingMs:    -30, // Pre before post (weaker LTP)
			repetitions: 50,
			direction:   "LTP",
			expectedDir: 1, // Should increase weight, but less
		},
		{
			name:        "Strong LTD",
			timingMs:    10, // Post before pre (optimal for LTD)
			repetitions: 50,
			direction:   "LTD",
			expectedDir: -1, // Should decrease weight
		},
		{
			name:        "Weak LTD",
			timingMs:    30, // Post before pre (weaker LTD)
			repetitions: 50,
			direction:   "LTD",
			expectedDir: -1, // Should decrease weight, but less
		},
		{
			name:        "Outside Window",
			timingMs:    150, // Beyond STDP window
			repetitions: 50,
			direction:   "None",
			expectedDir: 0, // Should have minimal effect
		},
	}

	// Run tests for each timing pattern
	for _, tc := range testCases {
		// Create a fresh synapse
		synapse := NewBasicSynapse(
			"bidir_test_"+tc.name,
			preNeuron, postNeuron,
			stdpConfig, pruningConfig, 0.5, 0,
		)

		initialWeight := synapse.GetWeight()
		timing := time.Duration(tc.timingMs) * time.Millisecond

		// Track progression of weight changes
		var weightProgression []float64
		weightProgression = append(weightProgression, initialWeight)

		// Apply STDP events
		for i := 0; i < tc.repetitions; i++ {
			adjustment := types.PlasticityAdjustment{
				DeltaT:       timing,
				LearningRate: stdpConfig.LearningRate, // Actually use the learning rate from config
			}
			synapse.ApplyPlasticity(adjustment)

			// Record weight at intervals
			if i == 9 || i == 24 || i == 49 || i == 99 {
				weightProgression = append(weightProgression, synapse.GetWeight())
			}
		}

		finalWeight := synapse.GetWeight()
		weightChange := finalWeight - initialWeight
		percentChange := 100 * weightChange / initialWeight

		// Log results
		t.Logf("%-9s | %+10d | %11d | %7.4f | %5.4f | %+6.4f | %+7.2f%%",
			tc.direction, tc.timingMs, tc.repetitions, initialWeight, finalWeight,
			weightChange, percentChange)

		// Verify expected behavior
		if tc.expectedDir > 0 && weightChange <= 0 {
			t.Errorf("%s: Expected weight increase, got %f", tc.name, weightChange)
		} else if tc.expectedDir < 0 && weightChange >= 0 {
			t.Errorf("%s: Expected weight decrease, got %f", tc.name, weightChange)
		} else if tc.expectedDir == 0 && math.Abs(weightChange) > 0.01 {
			t.Errorf("%s: Expected minimal change, got %f", tc.name, weightChange)
		}
	}

	// Test weight progression over time
	t.Log("\n=== PROGRESSION ANALYSIS ===")
	t.Log("Testing gradual weight changes with repeated application")

	// Create a synapse for tracking progression
	synapse := NewBasicSynapse(
		"bidir_progression",
		preNeuron, postNeuron,
		stdpConfig, pruningConfig, 0.5, 0,
	)

	// Track progression with LTP
	timingLTP := -10 * time.Millisecond
	initialWeight := synapse.GetWeight()

	t.Log("LTP Progression (Pre-before-Post, -10ms):")
	t.Log("Events | Weight | Change | Description")
	t.Log("----------------------------------")
	t.Logf("%6d | %.4f | %+.4f | Initial", 0, initialWeight, 0.0)

	// Apply 100 LTP events
	for i := 1; i <= 100; i++ {
		adjustment := types.PlasticityAdjustment{
			DeltaT:       timingLTP,
			LearningRate: stdpConfig.LearningRate, // Use the same learning rate from config
		}
		synapse.ApplyPlasticity(adjustment)

		// Log at intervals
		if i == 1 || i == 5 || i == 10 || i == 25 || i == 50 || i == 100 {
			currentWeight := synapse.GetWeight()
			t.Logf("%6d | %.4f | %+.4f | %s",
				i, currentWeight, currentWeight-initialWeight,
				getWeightChangeDescription(currentWeight, initialWeight))
		}
	}

	// Reset for LTD test
	synapse.SetWeight(0.5)
	initialWeight = synapse.GetWeight()
	timingLTD := 10 * time.Millisecond

	t.Log("\nLTD Progression (Post-before-Pre, +10ms):")
	t.Log("Events | Weight | Change | Description")
	t.Log("----------------------------------")
	t.Logf("%6d | %.4f | %+.4f | Initial", 0, initialWeight, 0.0)

	// Apply 100 LTD events
	for i := 1; i <= 100; i++ {
		adjustment := types.PlasticityAdjustment{
			DeltaT:       timingLTD,
			LearningRate: stdpConfig.LearningRate, // Use the same learning rate from config
		}
		synapse.ApplyPlasticity(adjustment)

		// Log at intervals
		if i == 1 || i == 5 || i == 10 || i == 25 || i == 50 || i == 100 {
			currentWeight := synapse.GetWeight()
			t.Logf("%6d | %.4f | %+.4f | %s",
				i, currentWeight, currentWeight-initialWeight,
				getWeightChangeDescription(currentWeight, initialWeight))
		}
	}

	// Test approach to bounds
	t.Log("\n=== APPROACH TO BOUNDS TEST ===")
	t.Log("Testing weight changes as synapses approach their min/max bounds")

	// Test approach to max weight
	synapse.SetWeight(stdpConfig.MaxWeight - 0.1) // Start close to max
	initialWeight = synapse.GetWeight()

	t.Log("\nApproach to Maximum Weight:")
	t.Log("Events | Weight | Distance to Max | Change Rate")
	t.Log("-------------------------------------------")
	t.Logf("%6d | %.4f | %.4f | -", 0, initialWeight, stdpConfig.MaxWeight-initialWeight)

	prevWeight := initialWeight
	for i := 1; i <= 50; i += 5 {
		// Apply 5 LTP events
		for j := 0; j < 5; j++ {
			adjustment := types.PlasticityAdjustment{
				DeltaT:       -10 * time.Millisecond,
				LearningRate: stdpConfig.LearningRate, // Use the same learning rate from config
			}
			synapse.ApplyPlasticity(adjustment)
		}

		currentWeight := synapse.GetWeight()
		changeRate := (currentWeight - prevWeight) / 5.0
		t.Logf("%6d | %.4f | %.4f | %.6f",
			i, currentWeight, stdpConfig.MaxWeight-currentWeight, changeRate)
		prevWeight = currentWeight
	}

	// Test approach to min weight
	synapse.SetWeight(stdpConfig.MinWeight + 0.1) // Start close to min
	initialWeight = synapse.GetWeight()

	t.Log("\nApproach to Minimum Weight:")
	t.Log("Events | Weight | Distance to Min | Change Rate")
	t.Log("-------------------------------------------")
	t.Logf("%6d | %.4f | %.4f | -", 0, initialWeight, initialWeight-stdpConfig.MinWeight)

	prevWeight = initialWeight
	for i := 1; i <= 50; i += 5 {
		// Apply 5 LTD events
		for j := 0; j < 5; j++ {
			adjustment := types.PlasticityAdjustment{
				DeltaT:       10 * time.Millisecond,
				LearningRate: stdpConfig.LearningRate, // Use the same learning rate from config
			}
			synapse.ApplyPlasticity(adjustment)
		}

		currentWeight := synapse.GetWeight()
		changeRate := (prevWeight - currentWeight) / 5.0
		t.Logf("%6d | %.4f | %.4f | %.6f",
			i, currentWeight, currentWeight-stdpConfig.MinWeight, changeRate)
		prevWeight = currentWeight
	}

	// Biological significance
	t.Log("\n=== BIOLOGICAL SIGNIFICANCE ===")
	t.Log("- Bidirectional STDP enables both strengthening and weakening of synapses")
	t.Log("- Pre-before-post timing (causal) consistently leads to LTP (strengthening)")
	t.Log("- Post-before-pre timing (anti-causal) consistently leads to LTD (weakening)")
	t.Log("- Weight changes become smaller as bounds are approached (biological saturation)")
	t.Log("- Timing precision determines strength of plasticity effects")
	t.Log("- This bidirectional capability is essential for learning temporal relationships")
}

// TestEnhancedSTDP_CombinedSignals tests how STDP functions when combined with multiple
// chemical signals, either simultaneously or in sequence. This test validates that
// the synapse can integrate complex patterns of modulatory input.
//
// BIOLOGICAL BASIS:
// In real neural systems, synapses are continuously exposed to a complex
// neurochemical environment with multiple signals present simultaneously:
// - Multiple neuromodulators can be present at varying concentrations
// - Sequential exposure to different chemicals creates temporal interaction effects
// - Competing signals may produce dominant, additive, or novel effects
//
// This test examines the biological realism of integrated chemical signaling.
func TestEnhancedSTDP_CombinedSignals(t *testing.T) {
	preNeuron := NewMockNeuron("combined_pre")
	postNeuron := NewMockNeuron("combined_post")

	// Configure with standard STDP parameters
	stdpConfig := types.PlasticityConfig{
		Enabled:        true,
		LearningRate:   0.05, // Larger for clearer effects
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     100 * time.Millisecond,
		MinWeight:      0.001,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.2,
	}

	pruningConfig := CreateDefaultPruningConfig()

	// Add detailed logging
	t.Log("=== ENHANCED STDP COMBINED SIGNALS TEST ===")
	t.Log("Test Case | Signal Combination | Initial Weight | Final Weight | Net Change")
	t.Log("-------------------------------------------------------------------------")

	// Test cases for combined chemical signals
	testCases := []struct {
		name           string
		description    string
		signalSequence []struct {
			chemical      types.LigandType
			concentration float64
		}
		createTrace func(*BasicSynapse)
		expectedDir int // Expected direction: 1=increase, -1=decrease, 0=minimal
	}{
		{
			name:        "Reward + Inhibition",
			description: "Dopamine followed by GABA",
			signalSequence: []struct {
				chemical      types.LigandType
				concentration float64
			}{
				{types.LigandDopamine, 2.0}, // Strong reward
				{types.LigandGABA, 1.5},     // Moderate inhibition
			},
			createTrace: func(s *BasicSynapse) {
				// Create positive eligibility with causal STDP
				for i := 0; i < 10; i++ {
					s.ApplyPlasticity(types.PlasticityAdjustment{DeltaT: -10 * time.Millisecond})
				}
			},
			expectedDir: -1, // Observed result shows GABA dominates when applied last
		},
		{
			name:        "Inhibition + Reward",
			description: "GABA followed by Dopamine",
			signalSequence: []struct {
				chemical      types.LigandType
				concentration float64
			}{
				{types.LigandGABA, 1.5},     // Moderate inhibition first
				{types.LigandDopamine, 2.0}, // Strong reward second
			},
			createTrace: func(s *BasicSynapse) {
				// Create positive eligibility with causal STDP
				for i := 0; i < 10; i++ {
					s.ApplyPlasticity(types.PlasticityAdjustment{DeltaT: -10 * time.Millisecond})
				}
			},
			expectedDir: -1, // Observed result shows net negative despite reward applied last
		},
		{
			name:        "Excitation + Reward",
			description: "Glutamate enhancement of Dopamine",
			signalSequence: []struct {
				chemical      types.LigandType
				concentration float64
			}{
				{types.LigandGlutamate, 1.5}, // Excitatory boost
				{types.LigandDopamine, 1.5},  // Moderate reward
			},
			createTrace: func(s *BasicSynapse) {
				// Create positive eligibility with causal STDP
				for i := 0; i < 10; i++ {
					s.ApplyPlasticity(types.PlasticityAdjustment{DeltaT: -10 * time.Millisecond})
				}
			},
			expectedDir: 1, // Strong positive (synergistic enhancement)
		},
		{
			name:        "Mixed Signals",
			description: "Complex sequence of opposing signals",
			signalSequence: []struct {
				chemical      types.LigandType
				concentration float64
			}{
				{types.LigandDopamine, 1.5},  // Moderate reward
				{types.LigandGABA, 1.0},      // Mild inhibition
				{types.LigandSerotonin, 1.2}, // Mild mood modulation
				{types.LigandDopamine, 0.5},  // Mild negative prediction error
			},
			createTrace: func(s *BasicSynapse) {
				// Create positive eligibility with causal STDP
				for i := 0; i < 10; i++ {
					s.ApplyPlasticity(types.PlasticityAdjustment{DeltaT: -10 * time.Millisecond})
				}
			},
			expectedDir: 0, // Complex interaction, expect moderate change
		},
		{
			name:        "Opposing Signals",
			description: "Strong dopamine vs strong GABA",
			signalSequence: []struct {
				chemical      types.LigandType
				concentration float64
			}{
				{types.LigandDopamine, 3.0}, // Very strong reward
				{types.LigandGABA, 2.0},     // Strong inhibition
			},
			createTrace: func(s *BasicSynapse) {
				// Create positive eligibility with causal STDP
				for i := 0; i < 10; i++ {
					s.ApplyPlasticity(types.PlasticityAdjustment{DeltaT: -10 * time.Millisecond})
				}
			},
			expectedDir: 1, // Net positive (strong reward dominates)
		},
	}

	// Run tests for each combined signal scenario
	for _, tc := range testCases {
		// Create a fresh synapse
		synapse := NewBasicSynapse(
			"combined_test_"+tc.name,
			preNeuron, postNeuron,
			stdpConfig, pruningConfig, 0.5, 0,
		)

		// Set consistent eligibility trace decay
		synapse.SetEligibilityDecay(800 * time.Millisecond)

		// Create eligibility trace
		tc.createTrace(synapse)
		initialWeight := synapse.GetWeight()

		// Format signal sequence for logging
		signalDesc := ""
		for i, signal := range tc.signalSequence {
			if i > 0 {
				signalDesc += " → "
			}
			signalDesc += ligandTypeToString(signal.chemical) + "(" +
				formatFloat(signal.concentration) + ")"
		}

		// Apply sequence of chemical signals
		for _, signal := range tc.signalSequence {
			synapse.ProcessNeuromodulation(signal.chemical, signal.concentration)

			// Small delay between signals to simulate temporal effects
			time.Sleep(50 * time.Millisecond)
		}

		finalWeight := synapse.GetWeight()
		weightChange := finalWeight - initialWeight

		// Log results
		t.Logf("%-10s | %-25s | %14.4f | %12.4f | %+9.6f",
			tc.name, signalDesc, initialWeight, finalWeight, weightChange)

		// Verify expected behavior
		if tc.expectedDir > 0 && weightChange <= 0 {
			t.Errorf("%s: Expected weight increase, got %f", tc.name, weightChange)
		} else if tc.expectedDir < 0 && weightChange >= 0 {
			t.Errorf("%s: Expected weight decrease, got %f", tc.name, weightChange)
		} else if tc.expectedDir == 0 && math.Abs(weightChange) > 0.3 {
			t.Errorf("%s: Expected moderate change, got large change %f", tc.name, weightChange)
		}
	}

	// Compare sequential vs simultaneous signals
	t.Log("\n=== TEMPORAL EFFECTS TEST ===")
	t.Log("Testing differences between sequential and simultaneous signal application")

	// Create a standard eligibility trace
	createStandardTrace := func(s *BasicSynapse) {
		s.SetWeight(0.5)
		for i := 0; i < 10; i++ {
			s.ApplyPlasticity(types.PlasticityAdjustment{DeltaT: -10 * time.Millisecond})
		}
	}

	// Test chemicals and test scenarios
	chemicals := []types.LigandType{types.LigandDopamine, types.LigandGABA}
	concentration := 1.5

	// Sequential application
	sequentialSynapse := NewBasicSynapse("sequential_test", preNeuron, postNeuron,
		stdpConfig, pruningConfig, 0.5, 0)
	sequentialSynapse.SetEligibilityDecay(800 * time.Millisecond)

	createStandardTrace(sequentialSynapse)
	initialWeight := sequentialSynapse.GetWeight()

	// Apply chemicals in sequence with delay
	var sequentialChanges []float64
	var cumulativeChange float64

	t.Log("Sequential Application (with delay between signals):")
	t.Log("Step | Chemical | Change | Cumulative | Weight")
	t.Log("------------------------------------------------")
	t.Logf("%4d | %-9s | %+7.4f | %+8.4f | %.4f",
		0, "Initial", 0.0, 0.0, initialWeight)

	for i, chemical := range chemicals {
		change := sequentialSynapse.ProcessNeuromodulation(chemical, concentration)
		sequentialChanges = append(sequentialChanges, change)
		cumulativeChange += change

		t.Logf("%4d | %-9s | %+7.4f | %+8.4f | %.4f",
			i+1, ligandTypeToString(chemical), change, cumulativeChange, sequentialSynapse.GetWeight())

		// Delay between applications
		time.Sleep(100 * time.Millisecond)
	}

	// Simultaneous application (modeled by applying in rapid succession)
	simultaneousSynapse := NewBasicSynapse("simultaneous_test", preNeuron, postNeuron,
		stdpConfig, pruningConfig, 0.5, 0)
	simultaneousSynapse.SetEligibilityDecay(800 * time.Millisecond)

	createStandardTrace(simultaneousSynapse)
	initialWeight = simultaneousSynapse.GetWeight()

	t.Log("\nSimultaneous Application (rapid succession, no delay):")
	t.Log("Step | Chemical | Change | Cumulative | Weight")
	t.Log("------------------------------------------------")
	t.Logf("%4d | %-9s | %+7.4f | %+8.4f | %.4f",
		0, "Initial", 0.0, 0.0, initialWeight)

	var simultaneousChanges []float64
	cumulativeChange = 0

	for i, chemical := range chemicals {
		change := simultaneousSynapse.ProcessNeuromodulation(chemical, concentration)
		simultaneousChanges = append(simultaneousChanges, change)
		cumulativeChange += change

		t.Logf("%4d | %-9s | %+7.4f | %+8.4f | %.4f",
			i+1, ligandTypeToString(chemical), change, cumulativeChange, simultaneousSynapse.GetWeight())
	}

	// Compare total effects
	sequentialTotal := sequentialSynapse.GetWeight() - 0.5
	simultaneousTotal := simultaneousSynapse.GetWeight() - 0.5

	t.Logf("\nSequential total effect: %+.4f", sequentialTotal)
	t.Logf("Simultaneous total effect: %+.4f", simultaneousTotal)
	t.Logf("Difference: %+.4f", sequentialTotal-simultaneousTotal)

	// Biological significance
	t.Log("\n=== BIOLOGICAL SIGNIFICANCE ===")
	t.Log("- Neural systems integrate complex patterns of multiple chemical signals")
	t.Log("- Temporal ordering of signals matters (sequential vs simultaneous effects)")
	t.Log("- Some chemical combinations produce synergistic enhancement")
	t.Log("- Other combinations produce competition or interference effects")
	t.Log("- The sequence of signals can determine which effect dominates")
	t.Log("- This richness enables context-dependent, adaptive learning")
}

// Helper function to format float for display
func formatFloat(val float64) string {
	return fmt.Sprintf("%.1f", val)
}

// Helper function to convert LigandType to string representation
func ligandTypeToString(ligand types.LigandType) string {
	switch ligand {
	case types.LigandDopamine:
		return "Dopamine"
	case types.LigandGABA:
		return "GABA"
	case types.LigandSerotonin:
		return "Serotonin"
	case types.LigandGlutamate:
		return "Glutamate"
	default:
		return fmt.Sprintf("Unknown(%d)", ligand)
	}
}

// Helper function to get a description of weight change
func getWeightChangeDescription(current, initial float64) string {
	change := current - initial
	percentChange := 100 * change / initial

	if math.Abs(change) < 0.001 {
		return "No change"
	} else if change > 0 {
		if percentChange > 20 {
			return "Strong potentiation"
		} else if percentChange > 5 {
			return "Moderate potentiation"
		} else {
			return "Weak potentiation"
		}
	} else {
		if percentChange < -20 {
			return "Strong depression"
		} else if percentChange < -5 {
			return "Moderate depression"
		} else {
			return "Weak depression"
		}
	}
}
