package synapse

import (
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

// TestPruningGuidance_InhibitorySignals tests how inhibitory signals like GABA
// can influence pruning decisions. This test verifies that prolonged inhibition
// or inhibitory signals can mark synapses for pruning, mirroring biological
// processes where inactive or inhibited synapses are more likely to be eliminated.
//
// BIOLOGICAL BASIS:
// In biological neural networks, inhibitory signals not only reduce transmission
// efficacy but can also trigger molecular cascades that lead to synapse elimination.
// GABA (gamma-aminobutyric acid), the primary inhibitory neurotransmitter, can
// weaken synapses through various mechanisms including receptor internalization
// and cytoskeletal reorganization when present at high concentrations or for
// extended periods.
//
// TEST COVERAGE:
// - Effect of GABA on pruning eligibility
// - Threshold for inhibition-based pruning
// - Temporal aspects (duration of inhibition)
// - Interaction between inhibition and activity
func TestPruningGuidance_InhibitorySignals(t *testing.T) {
	// Create mock neurons for controlled testing
	preNeuron := NewMockNeuron("inhib_pre")
	postNeuron := NewMockNeuron("inhib_post")

	// Configure STDP parameters
	stdpConfig := CreateDefaultSTDPConfig()

	// Create standard pruning configuration
	pruningConfig := CreateDefaultPruningConfig()
	// Use accelerated inactivity threshold for testing
	pruningConfig.InactivityThreshold = 100 * time.Millisecond

	// Add detailed logging
	t.Log("=== INHIBITORY SIGNAL PRUNING GUIDANCE TEST ===")
	t.Log("Scenario | GABA Level | Duration | Initial Weight | Final Weight | Should Prune")
	t.Log("------------------------------------------------------------------------------")

	// Test scenarios with different levels of inhibition
	testCases := []struct {
		name          string
		gabaLevel     float64
		duration      time.Duration
		initialWeight float64
		expectPrune   bool
	}{
		{
			name:          "No Inhibition",
			gabaLevel:     0.0,
			duration:      200 * time.Millisecond,
			initialWeight: 0.2, // Well above pruning threshold
			expectPrune:   false,
		},
		{
			name:          "Mild Inhibition",
			gabaLevel:     0.5,
			duration:      200 * time.Millisecond,
			initialWeight: 0.2,
			expectPrune:   false, // Mild inhibition shouldn't trigger pruning
		},
		{
			name:          "Strong Inhibition",
			gabaLevel:     2.0,
			duration:      200 * time.Millisecond,
			initialWeight: 0.2,
			expectPrune:   true, // Strong inhibition should trigger pruning
		},
		{
			name:          "Prolonged Inhibition",
			gabaLevel:     1.0,
			duration:      500 * time.Millisecond,
			initialWeight: 0.2,
			expectPrune:   true, // Longer inhibition should trigger pruning
		},
		{
			name:          "Strong Inhibition Near Threshold",
			gabaLevel:     2.0,
			duration:      200 * time.Millisecond,
			initialWeight: 0.11, // Just above pruning threshold
			expectPrune:   true, // Should push it below threshold
		},
	}

	for _, tc := range testCases {
		// Create fresh synapse for each test
		synapse := NewBasicSynapse("inhib_test", preNeuron, postNeuron,
			stdpConfig, pruningConfig, tc.initialWeight, 0)

		// Apply GABA inhibition
		if tc.gabaLevel > 0 {
			// Apply multiple times to simulate continuous inhibition
			iterations := int(tc.duration / (50 * time.Millisecond))
			for i := 0; i < iterations; i++ {
				synapse.ProcessNeuromodulation(types.LigandGABA, tc.gabaLevel)
				time.Sleep(50 * time.Millisecond)
			}
		} else {
			// Just wait for the equivalent duration
			time.Sleep(tc.duration)
		}

		// Check pruning status
		shouldPrune := synapse.ShouldPrune()
		finalWeight := synapse.GetWeight()

		// Log detailed results
		pruneStatus := "No"
		if shouldPrune {
			pruneStatus = "Yes"
		}

		t.Logf("%-18s | %9.1f | %8.0fms | %13.4f | %12.4f | %s",
			tc.name, tc.gabaLevel, float64(tc.duration)/float64(time.Millisecond),
			tc.initialWeight, finalWeight, pruneStatus)

		// Verify expectations
		if shouldPrune != tc.expectPrune {
			t.Errorf("Case %s: Expected prune=%v, got prune=%v (weight: %.4f)",
				tc.name, tc.expectPrune, shouldPrune, finalWeight)
		}
	}

	// Biological significance
	t.Log("\n=== BIOLOGICAL SIGNIFICANCE ===")
	t.Log("- Inhibitory signals like GABA can trigger pruning processes")
	t.Log("- Stronger inhibition is more likely to mark synapses for pruning")
	t.Log("- Prolonged inhibition increases pruning probability")
	t.Log("- This models biological processes where inactive synapses are eliminated")
	t.Log("- Synapses close to the pruning threshold are more vulnerable to inhibition")
}

// TestPruningGuidance_ThresholdModulation tests how pruning thresholds can be
// dynamically adjusted based on network conditions, neuromodulatory signals,
// or specific activity patterns. This test verifies that modulating the
// pruning threshold affects which synapses are marked for elimination.
//
// BIOLOGICAL BASIS:
// In real neural networks, pruning thresholds are not fixed but adjust based on:
// - Developmental stage (extensive pruning during critical periods)
// - Neuromodulatory state (dopamine, acetylcholine levels)
// - Network activity levels (homeostatic regulation)
// - Sleep/wake cycles (sleep promotes pruning of weak connections)
//
// TEST COVERAGE:
// - Baseline pruning with default threshold
// - Effect of lowering pruning threshold (more aggressive pruning)
// - Effect of raising pruning threshold (more conservative pruning)
// - Threshold modulation through neuromodulatory signals
func TestPruningGuidance_ThresholdModulation(t *testing.T) {
	// Create mock neurons for testing
	preNeuron := NewMockNeuron("threshold_pre")
	postNeuron := NewMockNeuron("threshold_post")

	// Configure standard STDP
	stdpConfig := CreateDefaultSTDPConfig()

	// Add detailed logging
	t.Log("=== PRUNING THRESHOLD MODULATION TEST ===")
	t.Log("Scenario | Weight | Threshold | Inactivity | Should Prune")
	t.Log("-------------------------------------------------------")

	// Create a set of synapses with different weights
	weights := []float64{0.02, 0.05, 0.08, 0.12, 0.15}

	// Test different pruning thresholds
	thresholds := []float64{0.05, 0.10, 0.15}

	// Apply inactivity to all synapses
	inactivityPeriod := 200 * time.Millisecond

	for _, threshold := range thresholds {
		// Create pruning config with this threshold
		pruningConfig := PruningConfig{
			Enabled:             true,
			WeightThreshold:     threshold,
			InactivityThreshold: 100 * time.Millisecond, // Short for testing
		}

		t.Logf("\nTesting pruning threshold: %.2f", threshold)

		for _, weight := range weights {
			// Create synapse with this weight
			synapse := NewBasicSynapse(
				"threshold_test", preNeuron, postNeuron,
				stdpConfig, pruningConfig, weight, 0,
			)

			// Simulate inactivity
			time.Sleep(inactivityPeriod)

			// Check pruning status
			shouldPrune := synapse.ShouldPrune()

			// Log results
			pruneStatus := "No"
			if shouldPrune {
				pruneStatus = "Yes"
			}

			scenario := "Below threshold"
			if weight >= threshold {
				scenario = "Above threshold"
			}

			t.Logf("%-15s | %6.3f | %9.2f | %9.0fms | %s",
				scenario, weight, threshold,
				float64(inactivityPeriod)/float64(time.Millisecond), pruneStatus)

			// Verify expectations
			expectedPrune := weight < threshold && inactivityPeriod >= pruningConfig.InactivityThreshold
			if shouldPrune != expectedPrune {
				t.Errorf("Weight %.3f with threshold %.2f: Expected prune=%v, got prune=%v",
					weight, threshold, expectedPrune, shouldPrune)
			}
		}
	}

	// Neuromodulator influence on pruning thresholds
	t.Log("\n=== NEUROMODULATOR INFLUENCE ON PRUNING THRESHOLDS ===")
	t.Log("Modulator | Level | Weight | Result")
	t.Log("-------------------------------------")

	// Create standard config with middle threshold
	standardConfig := PruningConfig{
		Enabled:             true,
		WeightThreshold:     0.10,
		InactivityThreshold: 100 * time.Millisecond,
	}

	// Test weights near threshold
	borderlineWeight := 0.11 // Just above threshold

	// Test how neuromodulators affect pruning decisions for borderline synapses
	modulatorTests := []struct {
		name        string
		modulator   types.LigandType
		level       float64
		expectPrune bool
	}{
		// High dopamine protects synapses from pruning by lowering the effective pruning threshold.
		// This models how reward signals strengthen and protect important neural pathways.
		{"Dopamine (high)", types.LigandDopamine, 2.0, false},

		// Strong GABA promotes pruning by raising the effective pruning threshold.
		// This models how strong inhibitory signals can trigger molecular cascades
		// that lead to synapse elimination, especially during critical developmental periods.
		{"GABA (high)", types.LigandGABA, 2.0, true},

		// Low dopamine does not actively promote pruning, it simply removes protection.
		// Biologically, low dopamine levels don't directly activate pruning machinery
		// but rather remove trophic support that would otherwise protect synapses.
		{"Dopamine (low)", types.LigandDopamine, 0.5, false},

		// Even mild GABA can promote pruning for borderline synapses (those very near threshold).
		// This matches biological observations where weak inhibitory signals can still affect
		// synapses that are already vulnerable due to being near the pruning threshold.
		// The adjustment is scaled according to concentration (GABA_PRUNING_MODIFIER * concentration * GABA_MILD_PRUNING_FACTOR).
		{"GABA (low)", types.LigandGABA, 0.5, true},
	}

	for _, mt := range modulatorTests {
		// Create synapse with borderline weight
		synapse := NewBasicSynapse(
			"modulator_test", preNeuron, postNeuron,
			stdpConfig, standardConfig, borderlineWeight, 0,
		)

		// Apply neuromodulator several times
		for i := 0; i < 3; i++ {
			synapse.ProcessNeuromodulation(mt.modulator, mt.level)
		}

		// Simulate inactivity
		time.Sleep(inactivityPeriod)

		// Check current weight and pruning status
		currentWeight := synapse.GetWeight()
		shouldPrune := synapse.ShouldPrune()

		// Log results
		result := "Protected"
		if shouldPrune {
			result = "Pruned"
		}

		t.Logf("%-13s | %4.1f | %6.3f | %s",
			mt.name, mt.level, currentWeight, result)

		// Verify modulator influence matches expectations
		if shouldPrune != mt.expectPrune {
			t.Errorf("Modulator %s (%.1f): Expected prune=%v, got prune=%v (weight: %.4f)",
				mt.name, mt.level, mt.expectPrune, shouldPrune, currentWeight)
		}
	}

	// Biological significance
	t.Log("\n=== BIOLOGICAL SIGNIFICANCE ===")
	t.Log("- Pruning thresholds can be dynamically modulated based on network state")
	t.Log("- Lower thresholds increase pruning (similar to developmental critical periods)")
	t.Log("- Higher thresholds protect connections (similar to consolidation)")
	t.Log("- Neuromodulators can dynamically influence which synapses get pruned")
	t.Log("- This creates flexible, context-dependent structural plasticity")
}

// TestPruningGuidance_ActivityDependence tests how different activity patterns
// affect pruning decisions. This test builds on the existing activity-dependent
// pruning tests but focuses on more complex activity patterns and their influence
// on pruning eligibility.
//
// BIOLOGICAL BASIS:
// In biological neural networks, synaptic pruning is highly dependent on activity:
// - "Use it or lose it" principle drives elimination of inactive synapses
// - Recent activity protects synapses from pruning
// - Activity pattern (not just total activity) matters
// - Spike timing and correlation with network activity influence protection
//
// TEST COVERAGE:
// - Different activity histories and their effect on pruning
// - Protection through recent vs. historical activity
// - Effect of activity frequency and intensity
// - Recovery from pruning eligibility through activity
func TestPruningGuidance_ActivityDependence(t *testing.T) {
	// Create mock neurons for testing
	preNeuron := NewMockNeuron("activity_pre")
	postNeuron := NewMockNeuron("activity_post")

	// Standard STDP configuration
	stdpConfig := CreateDefaultSTDPConfig()

	// Create pruning config with accelerated timescales for testing
	pruningConfig := PruningConfig{
		Enabled:             true,
		WeightThreshold:     0.10,
		InactivityThreshold: 100 * time.Millisecond, // Short for testing
	}

	// Add detailed logging
	t.Log("=== ACTIVITY-DEPENDENT PRUNING GUIDANCE TEST ===")
	t.Log("Scenario | Activity Pattern | Weight | Time Since Activity | Should Prune")
	t.Log("--------------------------------------------------------------------")

	// Test different activity patterns
	testCases := []struct {
		name             string
		activityPattern  string
		simulateActivity func(*BasicSynapse)
		weight           float64
		expectPrune      bool
	}{
		{
			name:            "No Activity",
			activityPattern: "None",
			simulateActivity: func(s *BasicSynapse) {
				// No activity, just wait
				time.Sleep(200 * time.Millisecond)
			},
			weight:      0.08, // Below threshold
			expectPrune: true,
		},
		{
			name:            "Recent Strong Activity",
			activityPattern: "Recent burst",
			simulateActivity: func(s *BasicSynapse) {
				// Simulate burst of activity
				for i := 0; i < 5; i++ {
					s.Transmit(1.0)
					time.Sleep(10 * time.Millisecond)
				}
			},
			weight:      0.08, // Below threshold but protected by activity
			expectPrune: false,
		},
		{
			name:            "Historical Activity",
			activityPattern: "Old activity",
			simulateActivity: func(s *BasicSynapse) {
				// Activity followed by long silence
				for i := 0; i < 5; i++ {
					s.Transmit(1.0)
				}
				// Long inactivity period
				time.Sleep(200 * time.Millisecond)
			},
			weight:      0.08,
			expectPrune: true, // Historical activity doesn't protect
		},
		{
			name:            "Sparse Regular Activity",
			activityPattern: "Sparse but regular",
			simulateActivity: func(s *BasicSynapse) {
				// Sparse but regular activity
				for i := 0; i < 3; i++ {
					s.Transmit(1.0)
					time.Sleep(50 * time.Millisecond)
				}
			},
			weight:      0.08,
			expectPrune: false, // Regular activity protects
		},
		{
			name:            "Strong Weight No Activity",
			activityPattern: "None",
			simulateActivity: func(s *BasicSynapse) {
				// No activity, just wait
				time.Sleep(200 * time.Millisecond)
			},
			weight:      0.15,  // Above threshold
			expectPrune: false, // Strong weight protects regardless
		},
		{
			name:            "Recovery from Pruning Eligibility",
			activityPattern: "Inactive then active",
			simulateActivity: func(s *BasicSynapse) {
				// First inactive period
				time.Sleep(150 * time.Millisecond)

				// Check if eligible for pruning (should be)
				if !s.ShouldPrune() {
					t.Logf("Warning: Synapse not marked for pruning after inactivity")
				}

				// Then activity burst
				for i := 0; i < 5; i++ {
					s.Transmit(1.0)
					time.Sleep(5 * time.Millisecond)
				}
			},
			weight:      0.08,
			expectPrune: false, // Recent activity should rescue it
		},
	}

	for _, tc := range testCases {
		// Create synapse with specified weight
		synapse := NewBasicSynapse(
			"activity_test", preNeuron, postNeuron,
			stdpConfig, pruningConfig, tc.weight, 0,
		)

		// Simulate the activity pattern
		tc.simulateActivity(synapse)

		// Get activity info and check pruning status
		activityInfo := synapse.GetActivityInfo()
		timeSinceActivity := time.Since(activityInfo.LastTransmission)
		shouldPrune := synapse.ShouldPrune()

		// Log results
		pruneStatus := "No"
		if shouldPrune {
			pruneStatus = "Yes"
		}

		t.Logf("%-25s | %-17s | %6.3f | %15.0fms | %s",
			tc.name, tc.activityPattern, tc.weight,
			float64(timeSinceActivity)/float64(time.Millisecond), pruneStatus)

		// Verify expectations
		if shouldPrune != tc.expectPrune {
			t.Errorf("Case %s: Expected prune=%v, got prune=%v (weight: %.4f, time since activity: %v)",
				tc.name, tc.expectPrune, shouldPrune, tc.weight, timeSinceActivity)
		}
	}

	// Test plasticity activity vs transmission activity
	t.Log("\n=== PLASTICITY vs TRANSMISSION ACTIVITY ===")
	t.Log("Activity Type | Weight | Protects from Pruning")
	t.Log("------------------------------------------")

	// Test different types of activity
	plasticityOnlySynapse := NewBasicSynapse(
		"plasticity_only", preNeuron, postNeuron,
		stdpConfig, pruningConfig, 0.08, 0,
	)

	transmissionOnlySynapse := NewBasicSynapse(
		"transmission_only", preNeuron, postNeuron,
		stdpConfig, pruningConfig, 0.08, 0,
	)

	bothActivitySynapse := NewBasicSynapse(
		"both_activity", preNeuron, postNeuron,
		stdpConfig, pruningConfig, 0.08, 0,
	)

	// Apply plasticity-only activity
	for i := 0; i < 5; i++ {
		plasticityOnlySynapse.ApplyPlasticity(types.PlasticityAdjustment{
			DeltaT: -10 * time.Millisecond,
		})
	}

	// Apply transmission-only activity
	for i := 0; i < 5; i++ {
		transmissionOnlySynapse.Transmit(1.0)
	}

	// Apply both types of activity
	for i := 0; i < 3; i++ {
		bothActivitySynapse.Transmit(1.0)
		bothActivitySynapse.ApplyPlasticity(types.PlasticityAdjustment{
			DeltaT: -10 * time.Millisecond,
		})
	}

	// Wait for potential inactivity threshold
	time.Sleep(50 * time.Millisecond)

	// Check pruning status for each
	plasticityOnlyPrune := plasticityOnlySynapse.ShouldPrune()
	transmissionOnlyPrune := transmissionOnlySynapse.ShouldPrune()
	bothActivityPrune := bothActivitySynapse.ShouldPrune()

	// Log results
	t.Logf("%-15s | %6.3f | %s", "Plasticity only", 0.08,
		protectionStatus(!plasticityOnlyPrune))
	t.Logf("%-15s | %6.3f | %s", "Transmission only", 0.08,
		protectionStatus(!transmissionOnlyPrune))
	t.Logf("%-15s | %6.3f | %s", "Both activity", 0.08,
		protectionStatus(!bothActivityPrune))

	// Biological significance
	t.Log("\n=== BIOLOGICAL SIGNIFICANCE ===")
	t.Log("- Synaptic pruning follows the 'use it or lose it' principle")
	t.Log("- Recent activity protects synapses from elimination")
	t.Log("- Activity pattern matters, not just total activity")
	t.Log("- Both transmission and plasticity events count as 'activity'")
	t.Log("- Synapses can recover from pruning eligibility through activity")
	t.Log("- This creates adaptive networks that preserve useful connections")
}

// Helper function to format protection status
func protectionStatus(protected bool) string {
	if protected {
		return "Yes ✓"
	}
	return "No ✗"
}
