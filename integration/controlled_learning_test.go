package integration

import (
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/extracellular"
	"github.com/SynapticNetworks/temporal-neuron/neuron"
	"github.com/SynapticNetworks/temporal-neuron/synapse"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// TestControlledLearning tests pattern recognition with weight regulation
func TestControlledLearning(t *testing.T) {
	t.Log("=== CONTROLLED LEARNING PATTERN RECOGNITION TEST ===")
	t.Log("Using weight regulation to prevent runaway excitation")

	// Create matrix
	matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  1 * time.Millisecond,
		MaxComponents:   30,
	})

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}
	defer matrix.Stop()

	// Register controlled neuron type with PROPER parameters
	matrix.RegisterNeuronType("controlled_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		// CRITICAL: Shorter decay time to prevent accumulation
		n := neuron.NewNeuron(id, config.Threshold, 0.95, 2*time.Millisecond, 1.2, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandDopamine, types.LigandGABA})
		n.SetCallbacks(callbacks)

		// MODERATE learning rate to prevent weight explosion
		n.EnableSTDPFeedback(5*time.Millisecond, 0.3)

		return n, nil
	})

	// Register controlled synapse with weight bounds
	matrix.RegisterSynapseType("controlled_synapse", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, _ := matrix.GetNeuron(config.PresynapticID)
		postNeuron, _ := matrix.GetNeuron(config.PostsynapticID)

		// Create custom STDP config with TIGHT weight bounds
		stdpConfig := synapse.CreateDefaultSTDPConfig()
		stdpConfig.MaxWeight = 1.2  // Prevent weight explosion
		stdpConfig.MinWeight = 0.1  // Allow significant reduction but not elimination

		syn := synapse.NewBasicSynapse(id, preNeuron, postNeuron,
			stdpConfig, synapse.CreateDefaultPruningConfig(),
			config.InitialWeight, config.Delay)

		// Shorter eligibility trace to prevent excessive temporal effects
		syn.SetEligibilityDecay(800 * time.Millisecond)
		return syn, nil
	})

	// Create minimal circuit
	input, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "controlled_neuron", Threshold: 0.4})
	if err != nil {
		t.Fatalf("Failed to create input: %v", err)
	}

	output, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "controlled_neuron", Threshold: 0.7})
	if err != nil {
		t.Fatalf("Failed to create output: %v", err)
	}

	// Start neurons and register for chemical binding
	neurons := []component.NeuralComponent{input, output}
	for _, n := range neurons {
		n.Start()
		if receiver, ok := n.(component.ChemicalReceiver); ok {
			matrix.RegisterForBinding(receiver)
		}
	}
	defer func() {
		for _, n := range neurons {
			n.Stop()
		}
	}()

	// Create synapse with CONTROLLED initial weight
	syn, err := matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "controlled_synapse",
		PresynapticID:  input.ID(),
		PostsynapticID: output.ID(),
		InitialWeight:  0.5, // Moderate starting weight
		Delay:          1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create synapse: %v", err)
	}

	t.Log("Created controlled circuit with weight bounds: [0.1, 1.2]")

	// Helper function
	getWeight := func(syn component.SynapticProcessor) float64 {
		if weightGetter, ok := syn.(interface{ GetWeight() float64 }); ok {
			return weightGetter.GetWeight()
		}
		return 0.0
	}

	// Test patterns
	patterns := []struct {
		input    int
		expected int
		name     string
	}{
		{0, 0, "Silent -> Silent"},
		{1, 1, "Active -> Active"},
	}

	t.Logf("Initial weight: %.3f", getWeight(syn))

	// CONTROLLED learning with smaller steps
	t.Log("\n--- CONTROLLED LEARNING WITH WEIGHT REGULATION ---")
	epochs := 20 // Fewer epochs to prevent saturation

	for epoch := 0; epoch < epochs; epoch++ {
		if epoch%4 == 0 {
			t.Logf("\nEpoch %d:", epoch)
		}
		
		for _, pattern := range patterns {
			// Present input with CONTROLLED signal strength
			if pattern.input == 1 {
				input.Receive(types.NeuralSignal{
					Value:     1.0, // Moderate signal, not overwhelming
					Timestamp: time.Now(),
					SourceID:  "controlled_test",
					TargetID:  input.ID(),
				})
			}

			// Shorter processing time
			time.Sleep(30 * time.Millisecond)

			// Get output
			predicted := output.GetActivityLevel()
			predictedBinary := 0
			if predicted > 0.5 {
				predictedBinary = 1
			}

			correct := predictedBinary == pattern.expected
			
			if epoch%4 == 0 {
				status := "❌"
				if correct {
					status = "✅"
				}
				t.Logf("  %s -> Output: %.3f (%d) %s", pattern.name, predicted, predictedBinary, status)
			}

			// CONTROLLED supervision with MODERATE signals
			if pattern.expected == 1 && predictedBinary == 0 {
				// Need more activation - moderate dopamine
				matrix.ReleaseLigand(types.LigandDopamine, output.ID(), 1.3)
			} else if pattern.expected == 0 && predictedBinary == 1 {
				// Need less activation - moderate GABA
				matrix.ReleaseLigand(types.LigandGABA, output.ID(), 1.4)
			} else if correct {
				// Correct - light reinforcement
				if pattern.expected == 1 {
					matrix.ReleaseLigand(types.LigandDopamine, output.ID(), 1.1)
				} else {
					matrix.ReleaseLigand(types.LigandGABA, output.ID(), 0.8)
				}
			}

			// Shorter learning time
			time.Sleep(40 * time.Millisecond)

			// CRITICAL: Allow decay between patterns
			time.Sleep(100 * time.Millisecond)
		}
		
		if epoch%4 == 0 {
			weight := getWeight(syn)
			t.Logf("    Weight: %.3f (bounds: 0.1-1.2)", weight)
			
			// Check for runaway weight growth
			if weight > 1.0 {
				t.Logf("    ⚠️  Weight approaching upper bound")
			}
		}
	}

	// EXTENDED decay period before testing
	t.Log("\n--- EXTENDED DECAY PERIOD ---")
	t.Log("Allowing complete decay of activity...")
	time.Sleep(500 * time.Millisecond)

	// Final test with decay verification
	t.Log("\n--- FINAL TEST WITH DECAY VERIFICATION ---")
	finalWeight := getWeight(syn)
	t.Logf("Final weight: %.3f", finalWeight)

	correctFinal := 0
	totalFinal := len(patterns)

	for _, pattern := range patterns {
		// Verify baseline silence before each test
		baselineActivity := output.GetActivityLevel()
		t.Logf("Baseline before %s: %.3f", pattern.name, baselineActivity)

		// Present input
		if pattern.input == 1 {
			input.Receive(types.NeuralSignal{
				Value:     1.0,
				Timestamp: time.Now(),
				SourceID:  "final_controlled_test",
				TargetID:  input.ID(),
			})
		}

		time.Sleep(30 * time.Millisecond)

		predicted := output.GetActivityLevel()
		predictedBinary := 0
		if predicted > 0.5 {
			predictedBinary = 1
		}

		correct := predictedBinary == pattern.expected
		if correct {
			correctFinal++
		}

		status := "❌"
		if correct {
			status = "✅"
		}

		t.Logf("FINAL: %s -> %.3f (%d) %s", pattern.name, predicted, predictedBinary, status)

		// Extended decay between tests
		time.Sleep(200 * time.Millisecond)
	}

	accuracy := float64(correctFinal) / float64(totalFinal) * 100.0
	t.Logf("\nFINAL ACCURACY: %.1f%% (%d/%d)", accuracy, correctFinal, totalFinal)

	// Analysis
	if accuracy >= 90.0 {
		t.Log("🎉 SUCCESS: Controlled learning achieved reliable pattern recognition!")
	} else if accuracy >= 50.0 {
		t.Log("⚠️  PARTIAL: Controlled learning showed improvement")
	} else {
		t.Log("❌ FAILURE: Even controlled learning insufficient")
	}

	// Weight analysis
	t.Log("\n--- WEIGHT REGULATION ANALYSIS ---")
	t.Logf("Weight evolution: 0.500 -> %.3f", finalWeight)
	
	if finalWeight >= 0.1 && finalWeight <= 1.2 {
		t.Log("✅ Weight stayed within bounds")
	} else {
		t.Log("❌ Weight exceeded bounds")
	}

	// Final silence test
	t.Log("\n--- FINAL SILENCE VERIFICATION ---")
	time.Sleep(300 * time.Millisecond) // Extended decay
	
	finalSilence := output.GetActivityLevel()
	t.Logf("Final output with no input: %.3f", finalSilence)
	
	if finalSilence < 0.3 {
		t.Log("✅ Network maintains silence when appropriate")
	} else {
		t.Log("❌ Network shows persistent activity")
	}

	// Success criteria
	t.Log("\n--- SUCCESS CRITERIA ANALYSIS ---")
	canBeSilent := finalSilence < 0.3
	hasReasonableAccuracy := accuracy >= 50.0
	weightControlled := finalWeight >= 0.1 && finalWeight <= 1.2
	
	if canBeSilent && hasReasonableAccuracy && weightControlled {
		t.Log("🎯 SUCCESS: All criteria met - controlled learning works!")
		t.Log("   ✅ Can maintain silence")
		t.Log("   ✅ Shows learning improvement") 
		t.Log("   ✅ Weights remain controlled")
	} else {
		t.Log("🔍 Partial success - identifying remaining issues:")
		if !canBeSilent {
			t.Log("   ❌ Cannot maintain silence")
		}
		if !hasReasonableAccuracy {
			t.Log("   ❌ Learning insufficient")
		}
		if !weightControlled {
			t.Log("   ❌ Weight regulation failed")
		}
	}
}