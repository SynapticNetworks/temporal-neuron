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

// TestGABAInhibitionPattern tests pattern recognition using GABA-based inhibition (not negative weights)
func TestGABAInhibitionPattern(t *testing.T) {
	t.Log("=== GABA-BASED INHIBITION PATTERN RECOGNITION TEST ===")
	t.Log("Using GABA neurotransmitter inhibition instead of negative weights")

	// Create matrix
	matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  1 * time.Millisecond,
		MaxComponents:   50,
	})

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}
	defer matrix.Stop()

	// Register GABA-sensitive neuron type
	matrix.RegisterNeuronType("gaba_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.95, 8*time.Millisecond, 1.5, 0.0, 0.0)
		// CRITICAL: GABA receptors for inhibition
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandDopamine, types.LigandGABA})
		n.SetCallbacks(callbacks)

		// Strong learning rate
		n.EnableSTDPFeedback(5*time.Millisecond, 0.6)

		return n, nil
	})

	// Standard positive-weight synapses (work WITH the framework)
	matrix.RegisterSynapseType("learning_synapse", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, _ := matrix.GetNeuron(config.PresynapticID)
		postNeuron, _ := matrix.GetNeuron(config.PostsynapticID)

		syn := synapse.NewBasicSynapse(id, preNeuron, postNeuron,
			synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(),
			config.InitialWeight, config.Delay)

		syn.SetEligibilityDecay(1500 * time.Millisecond)
		return syn, nil
	})

	// Create simple 2-neuron circuit
	input, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "gaba_neuron", Threshold: 0.3})
	if err != nil {
		t.Fatalf("Failed to create input: %v", err)
	}

	output, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "gaba_neuron", Threshold: 0.8})
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

	// Single positive-weight synapse (working WITH the framework)
	syn, err := matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "learning_synapse",
		PresynapticID:  input.ID(),
		PostsynapticID: output.ID(),
		InitialWeight:  0.6, // Moderate positive weight
		Delay:          1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create synapse: %v", err)
	}

	t.Log("Created GABA-based circuit: Input -> Output (positive weight, GABA inhibition)")

	// Helper function
	getWeight := func(syn component.SynapticProcessor) float64 {
		if weightGetter, ok := syn.(interface{ GetWeight() float64 }); ok {
			return weightGetter.GetWeight()
		}
		return 0.0
	}

	// Test patterns: learn when to be active vs silent
	patterns := []struct {
		input    int
		expected int
		name     string
	}{
		{0, 0, "Silent Input -> Silent Output"},
		{1, 1, "Active Input -> Active Output"},
	}

	t.Logf("Initial weight: %.3f", getWeight(syn))

	// Training using GABA-based learning
	t.Log("\n--- TRAINING WITH GABA-BASED INHIBITION ---")
	epochs := 30

	for epoch := 0; epoch < epochs; epoch++ {
		if epoch%5 == 0 {
			t.Logf("\nEpoch %d:", epoch)
		}
		
		for _, pattern := range patterns {
			// Present input
			if pattern.input == 1 {
				input.Receive(types.NeuralSignal{
					Value:     1.2,
					Timestamp: time.Now(),
					SourceID:  "gaba_test",
					TargetID:  input.ID(),
				})
			}

			// Wait for propagation
			time.Sleep(60 * time.Millisecond)

			// Get output
			predicted := output.GetActivityLevel()
			predictedBinary := 0
			if predicted > 0.5 {
				predictedBinary = 1
			}

			correct := predictedBinary == pattern.expected
			
			if epoch%5 == 0 {
				status := "❌"
				if correct {
					status = "✅"
				}
				t.Logf("  %s -> Output: %.3f (%d) %s", pattern.name, predicted, predictedBinary, status)
			}

			// GABA-BASED SUPERVISED LEARNING
			if pattern.expected == 1 && predictedBinary == 0 {
				// Should be active but isn't - strengthen with dopamine
				matrix.ReleaseLigand(types.LigandDopamine, output.ID(), 1.8)
				matrix.ReleaseLigand(types.LigandDopamine, input.ID(), 1.5)
			} else if pattern.expected == 0 && predictedBinary == 1 {
				// Should be silent but isn't - CRITICAL: Use GABA to create inhibitory learning
				matrix.ReleaseLigand(types.LigandGABA, output.ID(), 2.5) // Strong GABA inhibition
				matrix.ReleaseLigand(types.LigandGABA, input.ID(), 1.8)  // Reduce input sensitivity
			} else if correct {
				// Correct response - reinforce appropriately
				if pattern.expected == 1 {
					matrix.ReleaseLigand(types.LigandDopamine, output.ID(), 1.2)
				} else {
					// For correct silence, mild GABA to maintain inhibitory state
					matrix.ReleaseLigand(types.LigandGABA, output.ID(), 1.0)
				}
			}

			// Wait for GABA/dopamine effects
			time.Sleep(80 * time.Millisecond)

			// Reset between patterns
			time.Sleep(40 * time.Millisecond)
		}
		
		if epoch%5 == 0 {
			t.Logf("    Weight: %.3f", getWeight(syn))
		}
	}

	// Final test
	t.Log("\n--- FINAL TEST ---")
	finalWeight := getWeight(syn)
	t.Logf("Final weight: %.3f", finalWeight)

	correctFinal := 0
	totalFinal := len(patterns)

	for _, pattern := range patterns {
		// Present input
		if pattern.input == 1 {
			input.Receive(types.NeuralSignal{
				Value:     1.2,
				Timestamp: time.Now(),
				SourceID:  "final_gaba_test",
				TargetID:  input.ID(),
			})
		}

		time.Sleep(60 * time.Millisecond)

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

		time.Sleep(50 * time.Millisecond)
	}

	accuracy := float64(correctFinal) / float64(totalFinal) * 100.0
	t.Logf("\nFINAL ACCURACY: %.1f%% (%d/%d)", accuracy, correctFinal, totalFinal)

	// Analysis
	if accuracy >= 90.0 {
		t.Log("🎉 SUCCESS: GABA-based inhibition working!")
	} else if accuracy >= 50.0 {
		t.Log("⚠️  PARTIAL: Some GABA-based learning occurred")
	} else {
		t.Log("❌ FAILURE: GABA inhibition insufficient")
	}

	// Test the critical capability: silence
	t.Log("\n--- SILENCE CAPABILITY TEST ---")
	time.Sleep(150 * time.Millisecond) // Let all activity decay
	
	silentOutput := output.GetActivityLevel()
	t.Logf("Output with no input (after decay): %.3f", silentOutput)
	
	if silentOutput < 0.2 {
		t.Log("✅ Network can maintain silence")
	} else {
		t.Log("❌ Network still shows baseline activity")
	}

	// Test responsiveness to input
	t.Log("\n--- RESPONSIVENESS TEST ---")
	input.Receive(types.NeuralSignal{
		Value:     1.2,
		Timestamp: time.Now(),
		SourceID:  "responsiveness_test",
		TargetID:  input.ID(),
	})
	
	time.Sleep(60 * time.Millisecond)
	activeOutput := output.GetActivityLevel()
	t.Logf("Output with input: %.3f", activeOutput)
	
	if activeOutput > 0.8 {
		t.Log("✅ Network can activate when stimulated")
	} else {
		t.Log("❌ Network shows reduced responsiveness")
	}

	// Summary
	t.Log("\n--- GABA INHIBITION ANALYSIS ---")
	t.Logf("Weight change: 0.600 -> %.3f", finalWeight)
	
	canBeSilent := silentOutput < 0.2
	canBeActive := activeOutput > 0.8
	
	if canBeSilent && canBeActive {
		t.Log("🎯 SUCCESS: Network learned selective activation through GABA inhibition!")
	} else if canBeActive && !canBeSilent {
		t.Log("⚠️  Network can activate but cannot maintain silence")
	} else if canBeSilent && !canBeActive {
		t.Log("⚠️  Network can be silent but cannot activate properly")
	} else {
		t.Log("❌ Network failed to learn either state")
	}
}