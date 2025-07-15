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

// TestDecayVerification tests if neurons can properly decay after stimulation
func TestDecayVerification(t *testing.T) {
	t.Log("=== DECAY VERIFICATION TEST ===")
	t.Log("Testing if neurons with synaptic connections can decay properly")

	// Create matrix
	matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  1 * time.Millisecond,
		MaxComponents:   20,
	})

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}
	defer matrix.Stop()

	// Register neuron type WITHOUT STDP to isolate the issue
	matrix.RegisterNeuronType("decay_test_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		// Very short decay time for rapid testing
		n := neuron.NewNeuron(id, config.Threshold, 0.95, 1*time.Millisecond, 1.2, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandDopamine, types.LigandGABA})
		n.SetCallbacks(callbacks)

		// NO STDP to eliminate learning effects
		// n.EnableSTDPFeedback() - COMMENTED OUT

		return n, nil
	})

	// Register simple synapse without STDP
	matrix.RegisterSynapseType("simple_synapse", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, _ := matrix.GetNeuron(config.PresynapticID)
		postNeuron, _ := matrix.GetNeuron(config.PostsynapticID)

		syn := synapse.NewBasicSynapse(id, preNeuron, postNeuron,
			synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(),
			config.InitialWeight, config.Delay)

		return syn, nil
	})

	// Test 1: Single neuron decay (control)
	t.Log("\n--- TEST 1: SINGLE NEURON DECAY (Control) ---")
	
	single, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "decay_test_neuron", Threshold: 0.5})
	if err != nil {
		t.Fatalf("Failed to create single neuron: %v", err)
	}
	
	single.Start()
	defer single.Stop()
	
	if receiver, ok := single.(component.ChemicalReceiver); ok {
		matrix.RegisterForBinding(receiver)
	}

	// Stimulate
	single.Receive(types.NeuralSignal{
		Value:     2.0,
		Timestamp: time.Now(),
		SourceID:  "decay_test",
		TargetID:  single.ID(),
	})

	// Track decay
	t.Log("Single neuron decay:")
	for i := 0; i < 10; i++ {
		activity := single.GetActivityLevel()
		t.Logf("  %dms: %.3f", i*50, activity)
		time.Sleep(50 * time.Millisecond)
	}

	// Test 2: Two neurons with synapse
	t.Log("\n--- TEST 2: TWO NEURONS WITH SYNAPSE ---")
	
	input, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "decay_test_neuron", Threshold: 0.4})
	if err != nil {
		t.Fatalf("Failed to create input: %v", err)
	}

	output, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "decay_test_neuron", Threshold: 0.6})
	if err != nil {
		t.Fatalf("Failed to create output: %v", err)
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

	// Create synapse
	syn, err := matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "simple_synapse",
		PresynapticID:  input.ID(),
		PostsynapticID: output.ID(),
		InitialWeight:  0.8,
		Delay:          1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create synapse: %v", err)
	}

	// Helper
	getWeight := func() float64 {
		if weightGetter, ok := syn.(interface{ GetWeight() float64 }); ok {
			return weightGetter.GetWeight()
		}
		return 0.0
	}

	t.Logf("Initial synapse weight: %.3f", getWeight())

	// Stimulate input
	input.Receive(types.NeuralSignal{
		Value:     1.5,
		Timestamp: time.Now(),
		SourceID:  "synapse_test",
		TargetID:  input.ID(),
	})

	// Track both neurons
	t.Log("Two-neuron system decay:")
	for i := 0; i < 15; i++ {
		inputActivity := input.GetActivityLevel()
		outputActivity := output.GetActivityLevel()
		weight := getWeight()
		t.Logf("  %dms: Input=%.3f, Output=%.3f, Weight=%.3f", i*50, inputActivity, outputActivity, weight)
		time.Sleep(50 * time.Millisecond)
	}

	// Test 3: With STDP enabled
	t.Log("\n--- TEST 3: WITH STDP ENABLED ---")

	// Register neuron WITH STDP
	matrix.RegisterNeuronType("stdp_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.95, 1*time.Millisecond, 1.2, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandDopamine, types.LigandGABA})
		n.SetCallbacks(callbacks)

		// ENABLE STDP
		n.EnableSTDPFeedback(5*time.Millisecond, 0.3)

		return n, nil
	})

	input_stdp, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "stdp_neuron", Threshold: 0.4})
	if err != nil {
		t.Fatalf("Failed to create STDP input: %v", err)
	}

	output_stdp, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "stdp_neuron", Threshold: 0.6})
	if err != nil {
		t.Fatalf("Failed to create STDP output: %v", err)
	}

	// Start STDP neurons
	stdp_neurons := []component.NeuralComponent{input_stdp, output_stdp}
	for _, n := range stdp_neurons {
		n.Start()
		if receiver, ok := n.(component.ChemicalReceiver); ok {
			matrix.RegisterForBinding(receiver)
		}
	}
	defer func() {
		for _, n := range stdp_neurons {
			n.Stop()
		}
	}()

	// Create STDP synapse
	syn_stdp, err := matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "simple_synapse",
		PresynapticID:  input_stdp.ID(),
		PostsynapticID: output_stdp.ID(),
		InitialWeight:  0.8,
		Delay:          1 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create STDP synapse: %v", err)
	}

	getSTDPWeight := func() float64 {
		if weightGetter, ok := syn_stdp.(interface{ GetWeight() float64 }); ok {
			return weightGetter.GetWeight()
		}
		return 0.0
	}

	t.Logf("Initial STDP synapse weight: %.3f", getSTDPWeight())

	// Stimulate STDP input
	input_stdp.Receive(types.NeuralSignal{
		Value:     1.5,
		Timestamp: time.Now(),
		SourceID:  "stdp_test",
		TargetID:  input_stdp.ID(),
	})

	// Track STDP system
	t.Log("STDP system decay:")
	for i := 0; i < 15; i++ {
		inputActivity := input_stdp.GetActivityLevel()
		outputActivity := output_stdp.GetActivityLevel()
		weight := getSTDPWeight()
		t.Logf("  %dms: Input=%.3f, Output=%.3f, Weight=%.3f", i*50, inputActivity, outputActivity, weight)
		time.Sleep(50 * time.Millisecond)
	}

	// Analysis
	t.Log("\n--- DECAY ANALYSIS ---")
	
	singleFinal := single.GetActivityLevel()
	outputFinal := output.GetActivityLevel()
	outputSTDPFinal := output_stdp.GetActivityLevel()
	
	t.Logf("Final activities:")
	t.Logf("  Single neuron: %.3f", singleFinal)
	t.Logf("  Output (no STDP): %.3f", outputFinal)
	t.Logf("  Output (with STDP): %.3f", outputSTDPFinal)
	
	if singleFinal < 0.1 {
		t.Log("✅ Single neurons decay properly")
	} else {
		t.Log("❌ Single neuron decay issue")
	}
	
	if outputFinal < 0.1 {
		t.Log("✅ Synaptic connections don't prevent decay")
	} else {
		t.Log("❌ Synaptic connections cause persistent activity")
	}
	
	if outputSTDPFinal > outputFinal + 0.2 {
		t.Log("❌ STDP causes additional persistent activity")
	} else {
		t.Log("✅ STDP doesn't significantly increase persistent activity")
	}

	// Weight analysis
	finalWeight := getWeight()
	finalSTDPWeight := getSTDPWeight()
	
	t.Logf("Final weights:")
	t.Logf("  No STDP: %.3f (initial: 0.8)", finalWeight)
	t.Logf("  With STDP: %.3f (initial: 0.8)", finalSTDPWeight)
	
	if finalSTDPWeight != 0.8 {
		t.Log("✅ STDP modified weights as expected")
	} else {
		t.Log("⚠️  STDP didn't modify weights")
	}

	// Root cause identification
	t.Log("\n--- ROOT CAUSE IDENTIFICATION ---")
	
	if singleFinal < 0.1 && outputFinal > 0.5 {
		t.Log("🔍 ROOT CAUSE: Synaptic connections prevent proper decay")
		t.Log("   - Isolated neurons decay fine")
		t.Log("   - Connected neurons show persistent activity")
		t.Log("   - Suggests recurrent activity or feedback loops")
	} else if outputSTDPFinal > outputFinal + 0.5 {
		t.Log("🔍 ROOT CAUSE: STDP mechanisms prevent decay")
		t.Log("   - Basic connections decay fine")
		t.Log("   - STDP adds persistent activity")
		t.Log("   - Suggests STDP feedback creates recurrence")
	} else if singleFinal > 0.5 {
		t.Log("🔍 ROOT CAUSE: Fundamental neuron decay issue")
		t.Log("   - Even isolated neurons don't decay")
		t.Log("   - Problem in basic neuron implementation")
	} else {
		t.Log("🔍 ROOT CAUSE: Complex interaction between components")
		t.Log("   - Multiple factors contributing to persistent activity")
	}
}