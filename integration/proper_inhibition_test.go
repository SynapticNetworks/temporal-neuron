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

// TestProperInhibitionPattern tests pattern recognition using proper inhibitory mechanisms
func TestProperInhibitionPattern(t *testing.T) {
	t.Log("=== PROPER INHIBITION PATTERN RECOGNITION TEST ===")
	t.Log("Using excitatory AND inhibitory synapses with GABA receptors")

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

	// Register neuron type with PROPER GABA sensitivity
	matrix.RegisterNeuronType("gaba_sensitive", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.95, 5*time.Millisecond, 2.0, 0.0, 0.0)
		// CRITICAL: Include GABA receptors for inhibition
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandDopamine, types.LigandGABA})
		n.SetCallbacks(callbacks)

		// Strong learning rate for fast adaptation
		n.EnableSTDPFeedback(3*time.Millisecond, 0.8)

		return n, nil
	})

	// Register synapse types: both excitatory and inhibitory
	matrix.RegisterSynapseType("excitatory_synapse", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, _ := matrix.GetNeuron(config.PresynapticID)
		postNeuron, _ := matrix.GetNeuron(config.PostsynapticID)

		syn := synapse.NewBasicSynapse(id, preNeuron, postNeuron,
			synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(),
			config.InitialWeight, config.Delay)

		syn.SetEligibilityDecay(2000 * time.Millisecond)
		return syn, nil
	})

	matrix.RegisterSynapseType("inhibitory_synapse", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, _ := matrix.GetNeuron(config.PresynapticID)
		postNeuron, _ := matrix.GetNeuron(config.PostsynapticID)

		syn := synapse.NewBasicSynapse(id, preNeuron, postNeuron,
			synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(),
			config.InitialWeight, config.Delay)

		syn.SetEligibilityDecay(2000 * time.Millisecond)
		return syn, nil
	})

	// Create circuit with inhibitory architecture
	input, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "gaba_sensitive", Threshold: 0.3})
	if err != nil {
		t.Fatalf("Failed to create input: %v", err)
	}

	// Create inhibitory interneuron
	inhibitory, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "gaba_sensitive", Threshold: 0.5})
	if err != nil {
		t.Fatalf("Failed to create inhibitory: %v", err)
	}

	output, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "gaba_sensitive", Threshold: 0.8})
	if err != nil {
		t.Fatalf("Failed to create output: %v", err)
	}

	// Start neurons and register for chemical binding
	neurons := []component.NeuralComponent{input, inhibitory, output}
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

	// Create DUAL pathway: excitatory AND inhibitory
	// Excitatory path: Input -> Output (positive weight)
	excSyn, err := matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "excitatory_synapse",
		PresynapticID:  input.ID(),
		PostsynapticID: output.ID(),
		InitialWeight:  0.8, // Start with moderate excitation
		Delay:          1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create excitatory synapse: %v", err)
	}

	// Inhibitory path: Input -> Inhibitory -> Output (negative weight)
	_, err = matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "excitatory_synapse",
		PresynapticID:  input.ID(),
		PostsynapticID: inhibitory.ID(),
		InitialWeight:  0.6,
		Delay:          1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create input->inhibitory synapse: %v", err)
	}

	// CRITICAL: Inhibitory synapse with NEGATIVE weight
	inhibSyn, err := matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "inhibitory_synapse",
		PresynapticID:  inhibitory.ID(),
		PostsynapticID: output.ID(),
		InitialWeight:  -0.5, // NEGATIVE weight for inhibition
		Delay:          2 * time.Millisecond, // Slight delay
	})
	if err != nil {
		t.Fatalf("Failed to create inhibitory synapse: %v", err)
	}

	t.Log("Created inhibitory circuit:")
	t.Log("  Input -> Output (excitatory, weight=0.8)")
	t.Log("  Input -> Inhibitory -> Output (inhibitory, weight=-0.5)")

	// Helper functions
	getWeight := func(syn component.SynapticProcessor) float64 {
		if weightGetter, ok := syn.(interface{ GetWeight() float64 }); ok {
			return weightGetter.GetWeight()
		}
		return 0.0
	}

	// Test patterns: want to learn Input=1 -> Output=1, Input=0 -> Output=0
	patterns := []struct {
		input    int
		expected int
		name     string
	}{
		{0, 0, "Silence -> Silence"}, // Should strengthen inhibition
		{1, 1, "Active -> Active"},   // Should strengthen excitation
	}

	t.Logf("\nInitial weights: Excitatory=%.3f, Inhibitory=%.3f", 
		getWeight(excSyn), getWeight(inhibSyn))

	// Training with proper inhibitory learning
	t.Log("\n--- TRAINING WITH INHIBITORY MECHANISMS ---")
	epochs := 25

	for epoch := 0; epoch < epochs; epoch++ {
		if epoch%5 == 0 {
			t.Logf("\nEpoch %d:", epoch)
		}
		
		for _, pattern := range patterns {
			// Present input
			if pattern.input == 1 {
				input.Receive(types.NeuralSignal{
					Value:     1.8,
					Timestamp: time.Now(),
					SourceID:  "test",
					TargetID:  input.ID(),
				})
			}

			// Wait for propagation through both pathways
			time.Sleep(80 * time.Millisecond)

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

			// PROPER SUPERVISED LEARNING using BOTH dopamine AND GABA
			if pattern.expected == 1 && predictedBinary == 0 {
				// Should be active but isn't - strengthen excitation, weaken inhibition
				matrix.ReleaseLigand(types.LigandDopamine, output.ID(), 2.0)
				matrix.ReleaseLigand(types.LigandGABA, inhibitory.ID(), 1.5) // Inhibit the inhibitory neuron
			} else if pattern.expected == 0 && predictedBinary == 1 {
				// Should be silent but isn't - strengthen inhibition, weaken excitation
				matrix.ReleaseLigand(types.LigandGABA, output.ID(), 2.0) // Strong inhibition
				matrix.ReleaseLigand(types.LigandDopamine, inhibitory.ID(), 1.8) // Strengthen inhibitory pathway
			} else if correct {
				// Correct response - moderate positive reinforcement
				if pattern.expected == 1 {
					matrix.ReleaseLigand(types.LigandDopamine, output.ID(), 1.2)
				} else {
					matrix.ReleaseLigand(types.LigandDopamine, inhibitory.ID(), 1.2)
				}
			}

			// Wait for learning
			time.Sleep(100 * time.Millisecond)

			// Reset for next pattern
			time.Sleep(50 * time.Millisecond)
		}
		
		if epoch%5 == 0 {
			t.Logf("    Weights: Excitatory=%.3f, Inhibitory=%.3f", 
				getWeight(excSyn), getWeight(inhibSyn))
		}
	}

	// Final test
	t.Log("\n--- FINAL TEST ---")
	finalExcWeight := getWeight(excSyn)
	finalInhibWeight := getWeight(inhibSyn)
	t.Logf("Final weights: Excitatory=%.3f, Inhibitory=%.3f", finalExcWeight, finalInhibWeight)

	correctFinal := 0
	totalFinal := len(patterns)

	for _, pattern := range patterns {
		// Present input
		if pattern.input == 1 {
			input.Receive(types.NeuralSignal{
				Value:     1.8,
				Timestamp: time.Now(),
				SourceID:  "final_test",
				TargetID:  input.ID(),
			})
		}

		time.Sleep(80 * time.Millisecond)

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
		t.Log("🎉 SUCCESS: Inhibitory pattern recognition working!")
	} else if accuracy >= 50.0 {
		t.Log("⚠️  PARTIAL: Some inhibitory learning occurred")
	} else {
		t.Log("❌ FAILURE: Inhibitory mechanisms insufficient")
	}

	// Weight analysis
	t.Log("\n--- WEIGHT EVOLUTION ANALYSIS ---")
	t.Logf("Excitatory weight: 0.800 -> %.3f", finalExcWeight)
	t.Logf("Inhibitory weight: -0.500 -> %.3f", finalInhibWeight)

	if finalExcWeight > 0.9 && finalInhibWeight < -0.6 {
		t.Log("✅ Proper differentiation: Excitation strengthened, inhibition strengthened")
	} else if finalExcWeight > 0.9 {
		t.Log("⚠️  Excitation learned but inhibition insufficient")
	} else if finalInhibWeight < -0.6 {
		t.Log("⚠️  Inhibition learned but excitation insufficient")
	} else {
		t.Log("❌ Neither pathway learned properly")
	}

	// Test the key insight: can the network be SILENT when it should be?
	t.Log("\n--- SILENCE TEST ---")
	// No input - should produce no output
	time.Sleep(100 * time.Millisecond)
	silentOutput := output.GetActivityLevel()
	t.Logf("Output with no input: %.3f", silentOutput)
	
	if silentOutput < 0.3 {
		t.Log("✅ Network learned to be silent when appropriate")
	} else {
		t.Log("❌ Network still active without input - inhibition insufficient")
	}
}