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

// TestSimplestPattern tests the absolute simplest pattern recognition
func TestSimplestPattern(t *testing.T) {
	t.Log("=== SIMPLEST PATTERN RECOGNITION TEST ===")
	t.Log("Testing the most basic possible pattern: single bit input -> single bit output")

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

	// Register ultra-simple neuron type
	matrix.RegisterNeuronType("ultra_simple", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.95, 5*time.Millisecond, 2.0, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandDopamine, types.LigandGABA})
		n.SetCallbacks(callbacks)

		// MAXIMUM learning rate
		n.EnableSTDPFeedback(3*time.Millisecond, 1.0) // 100% learning rate!

		return n, nil
	})

	// Register ultra-simple synapse
	matrix.RegisterSynapseType("ultra_simple_synapse", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, _ := matrix.GetNeuron(config.PresynapticID)
		postNeuron, _ := matrix.GetNeuron(config.PostsynapticID)

		syn := synapse.NewBasicSynapse(id, preNeuron, postNeuron,
			synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(),
			config.InitialWeight, config.Delay)

		// Long eligibility trace
		syn.SetEligibilityDecay(3000 * time.Millisecond) // 3 seconds!

		return syn, nil
	})

	// Create minimal circuit: input -> output (direct connection)
	input, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "ultra_simple", Threshold: 0.3})
	if err != nil {
		t.Fatalf("Failed to create input: %v", err)
	}

	output, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "ultra_simple", Threshold: 0.8})
	if err != nil {
		t.Fatalf("Failed to create output: %v", err)
	}

	// Start neurons
	input.Start()
	output.Start()
	defer input.Stop()
	defer output.Stop()

	// Register for chemical binding
	if receiver, ok := input.(component.ChemicalReceiver); ok {
		matrix.RegisterForBinding(receiver)
	}
	if receiver, ok := output.(component.ChemicalReceiver); ok {
		matrix.RegisterForBinding(receiver)
	}

	// Create direct synapse
	syn, err := matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "ultra_simple_synapse",
		PresynapticID:  input.ID(),
		PostsynapticID: output.ID(),
		InitialWeight:  0.5, // Start neutral
		Delay:          1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create synapse: %v", err)
	}

	t.Log("Created minimal circuit: Input -> Output")
	t.Logf("Initial weight: %.3f", getWeight(syn))

	// Test ultra-simple patterns
	patterns := []struct {
		input    int
		expected int
		name     string
	}{
		{0, 0, "Input=0, Expected=0"},
		{1, 1, "Input=1, Expected=1"},
	}

	// Training phase with MAXIMUM supervision
	t.Log("\n--- TRAINING WITH MAXIMUM SUPERVISION ---")
	epochs := 20
	for epoch := 0; epoch < epochs; epoch++ {
		t.Logf("\nEpoch %d:", epoch)
		
		for _, pattern := range patterns {
			// Present input
			if pattern.input == 1 {
				input.Receive(types.NeuralSignal{
					Value:     2.0, // Strong signal
					Timestamp: time.Now(),
					SourceID:  "test",
					TargetID:  input.ID(),
				})
			}

			// Wait for propagation
			time.Sleep(100 * time.Millisecond)

			// Get output
			predicted := output.GetActivityLevel()
			predictedBinary := 0
			if predicted > 0.5 {
				predictedBinary = 1
			}

			correct := predictedBinary == pattern.expected
			status := "❌"
			if correct {
				status = "✅"
			}

			t.Logf("  %s -> Output: %.3f (%d) %s", pattern.name, predicted, predictedBinary, status)

			// MAXIMUM supervision
			if !correct {
				// Strong error signal
				matrix.ReleaseLigand(types.LigandGABA, output.ID(), 2.0)
				
				// Strong target signal if should be active
				if pattern.expected == 1 {
					matrix.ReleaseLigand(types.LigandDopamine, output.ID(), 3.0)
				}
			} else {
				// Positive reinforcement
				matrix.ReleaseLigand(types.LigandDopamine, output.ID(), 1.5)
			}

			// Wait for learning
			time.Sleep(150 * time.Millisecond)

			// Check weight change
			newWeight := getWeight(syn)
			t.Logf("    Weight change: %.3f", newWeight)

			// Reset for next pattern
			time.Sleep(100 * time.Millisecond)
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
				Value:     2.0,
				Timestamp: time.Now(),
				SourceID:  "final_test",
				TargetID:  input.ID(),
			})
		}

		time.Sleep(100 * time.Millisecond)

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

	if accuracy >= 90.0 {
		t.Log("🎉 SUCCESS: Minimal pattern recognition working!")
	} else if accuracy >= 50.0 {
		t.Log("⚠️  PARTIAL: Some learning occurred")
	} else {
		t.Log("❌ FAILURE: No meaningful learning")
	}

	// Diagnostic information
	t.Log("\n--- DIAGNOSTIC INFO ---")
	t.Logf("Weight evolution: 0.500 -> %.3f", finalWeight)
	
	if finalWeight > 0.7 {
		t.Log("Weight increased significantly - network learned to activate")
	} else if finalWeight < 0.3 {
		t.Log("Weight decreased significantly - network learned to suppress")
	} else {
		t.Log("Weight remained neutral - minimal learning occurred")
	}
}

// getWeight extracts weight from synapse
func getWeight(syn component.SynapticProcessor) float64 {
	if weightGetter, ok := syn.(interface{ GetWeight() float64 }); ok {
		return weightGetter.GetWeight()
	}
	return 0.0
}