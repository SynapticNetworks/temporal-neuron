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

// TestMinimalPatternDiscrimination tests the absolute minimum case for pattern discrimination
func TestMinimalPatternDiscrimination(t *testing.T) {
	t.Log("=== MINIMAL PATTERN DISCRIMINATION TEST ===")
	t.Log("Testing if network can learn to distinguish Pattern A from silence")

	// Create matrix
	matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  1 * time.Millisecond,
		MaxComponents:   10,
	})

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}
	defer matrix.Stop()

	// Register very simple neuron with no STDP, pure threshold
	matrix.RegisterNeuronType("simple_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		// High threshold, fast decay
		n := neuron.NewNeuron(id, 1.5, 0.8, 5*time.Millisecond, 1.0, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandGABA})
		n.SetCallbacks(callbacks)
		return n, nil
	})

	// Register basic synapse with fixed weights
	matrix.RegisterSynapseType("fixed_synapse", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, _ := matrix.GetNeuron(config.PresynapticID)
		postNeuron, _ := matrix.GetNeuron(config.PostsynapticID)

		syn := synapse.NewBasicSynapse(id, preNeuron, postNeuron,
			synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(),
			config.InitialWeight, config.Delay)

		return syn, nil
	})

	// Create single input and single output neuron
	input, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "simple_neuron", Threshold: 1.0})
	if err != nil {
		t.Fatalf("Failed to create input neuron: %v", err)
	}

	output, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "simple_neuron", Threshold: 1.5})
	if err != nil {
		t.Fatalf("Failed to create output neuron: %v", err)
	}

	// Start neurons
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

	// Create synapse with controlled weight
	_, err = matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "fixed_synapse",
		PresynapticID:  input.ID(),
		PostsynapticID: output.ID(),
		InitialWeight:  1.0, // Exactly enough to make output fire when input fires
		Delay:          1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create synapse: %v", err)
	}

	t.Log("\n--- TEST 1: SILENCE (No Input) ---")
	
	// Wait for any initial activity to decay
	time.Sleep(100 * time.Millisecond)
	
	// Check baseline activity - should be zero
	inputActivity := input.GetActivityLevel()
	outputActivity := output.GetActivityLevel()
	
	t.Logf("Baseline: Input=%.3f, Output=%.3f", inputActivity, outputActivity)
	
	if outputActivity > 0.1 {
		t.Log("❌ PROBLEM: Output shows activity during silence")
	} else {
		t.Log("✅ Output correctly silent during baseline")
	}

	t.Log("\n--- TEST 2: PATTERN A (Input Signal) ---")
	
	// Present Pattern A: stimulate input
	input.Receive(types.NeuralSignal{
		Value:     2.0, // Above input threshold
		Timestamp: time.Now(),
		SourceID:  "pattern_test",
		TargetID:  input.ID(),
	})

	// Wait for signal propagation
	time.Sleep(10 * time.Millisecond)

	// Check if output responds
	inputActivityA := input.GetActivityLevel()
	outputActivityA := output.GetActivityLevel()
	
	t.Logf("Pattern A: Input=%.3f, Output=%.3f", inputActivityA, outputActivityA)
	
	if outputActivityA > 0.5 {
		t.Log("✅ Output correctly responds to Pattern A")
	} else {
		t.Log("❌ PROBLEM: Output fails to respond to Pattern A")
	}

	t.Log("\n--- TEST 3: RETURN TO SILENCE ---")
	
	// Wait for decay
	time.Sleep(200 * time.Millisecond)
	
	// Check if system returns to silence
	inputActivityFinal := input.GetActivityLevel()
	outputActivityFinal := output.GetActivityLevel()
	
	t.Logf("Return to silence: Input=%.3f, Output=%.3f", inputActivityFinal, outputActivityFinal)
	
	if outputActivityFinal < 0.1 {
		t.Log("✅ System correctly returns to silence")
	} else {
		t.Log("❌ PROBLEM: System maintains persistent activity")
	}

	// --- ANALYSIS ---
	t.Log("\n--- DISCRIMINATION ANALYSIS ---")
	
	silenceResponse := outputActivity
	patternResponse := outputActivityA
	
	if patternResponse > silenceResponse * 3 {
		t.Log("✅ SUCCESSFUL DISCRIMINATION: Pattern A clearly distinguishable from silence")
		t.Logf("   Discrimination ratio: %.2f", patternResponse/silenceResponse)
	} else {
		t.Log("❌ FAILED DISCRIMINATION: Cannot distinguish Pattern A from silence")
		t.Logf("   Pattern response: %.3f", patternResponse)
		t.Logf("   Silence response: %.3f", silenceResponse)
		t.Logf("   Ratio: %.2f (need > 3.0)", patternResponse/(silenceResponse+0.001))
	}

	// Final assessment
	canDiscriminate := patternResponse > silenceResponse*3 && outputActivityFinal < 0.2
	
	if canDiscriminate {
		t.Log("\n🎯 CONCLUSION: Network CAN perform basic pattern discrimination")
		t.Log("   Ready to proceed with more complex learning tasks")
	} else {
		t.Log("\n💥 CONCLUSION: Network CANNOT perform basic pattern discrimination")
		t.Log("   Must fix fundamental architecture before attempting learning")
	}
}