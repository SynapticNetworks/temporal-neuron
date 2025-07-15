package integration

import (
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/extracellular"
	"github.com/SynapticNetworks/temporal-neuron/neuron"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// TestNeuronBaselineBehavior diagnoses why neurons show persistent activity
func TestNeuronBaselineBehavior(t *testing.T) {
	t.Log("=== NEURON BASELINE BEHAVIOR DIAGNOSTIC ===")
	t.Log("Investigating why neurons cannot maintain silence")

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

	// Register simple neuron type for diagnosis
	matrix.RegisterNeuronType("diagnostic_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.95, 5*time.Millisecond, 1.5, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandDopamine, types.LigandGABA})
		n.SetCallbacks(callbacks)
		return n, nil
	})

	// Test 1: Isolated neuron behavior
	t.Log("\n--- TEST 1: ISOLATED NEURON (No inputs) ---")
	
	isolated, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "diagnostic_neuron", Threshold: 0.8})
	if err != nil {
		t.Fatalf("Failed to create isolated neuron: %v", err)
	}
	
	isolated.Start()
	defer isolated.Stop()
	
	if receiver, ok := isolated.(component.ChemicalReceiver); ok {
		matrix.RegisterForBinding(receiver)
	}

	// Check baseline activity over time
	for i := 0; i < 10; i++ {
		activity := isolated.GetActivityLevel()
		t.Logf("Time %ds: Isolated neuron activity = %.3f", i, activity)
		time.Sleep(1000 * time.Millisecond)
	}

	// Test 2: Neuron with strong GABA inhibition
	t.Log("\n--- TEST 2: NEURON WITH MAXIMUM GABA INHIBITION ---")
	
	gaba_test, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "diagnostic_neuron", Threshold: 0.8})
	if err != nil {
		t.Fatalf("Failed to create GABA test neuron: %v", err)
	}
	
	gaba_test.Start()
	defer gaba_test.Stop()
	
	if receiver, ok := gaba_test.(component.ChemicalReceiver); ok {
		matrix.RegisterForBinding(receiver)
	}

	// Apply increasing GABA concentrations
	gabaLevels := []float64{0.5, 1.0, 2.0, 3.0, 5.0, 10.0}
	
	for _, gabaLevel := range gabaLevels {
		// Apply GABA
		matrix.ReleaseLigand(types.LigandGABA, gaba_test.ID(), gabaLevel)
		time.Sleep(200 * time.Millisecond)
		
		activity := gaba_test.GetActivityLevel()
		t.Logf("GABA %.1f: Activity = %.3f", gabaLevel, activity)
	}

	// Test 3: Neuron response to stimulation after GABA
	t.Log("\n--- TEST 3: STIMULATION AFTER GABA INHIBITION ---")
	
	// Wait for GABA to clear
	time.Sleep(2000 * time.Millisecond)
	
	beforeStim := gaba_test.GetActivityLevel()
	t.Logf("Before stimulation: %.3f", beforeStim)
	
	// Apply stimulation
	gaba_test.Receive(types.NeuralSignal{
		Value:     2.0,
		Timestamp: time.Now(),
		SourceID:  "test_stim",
		TargetID:  gaba_test.ID(),
	})
	
	time.Sleep(100 * time.Millisecond)
	afterStim := gaba_test.GetActivityLevel()
	t.Logf("After stimulation: %.3f", afterStim)
	
	// Test 4: Threshold effects
	t.Log("\n--- TEST 4: THRESHOLD EFFECTS ---")
	
	thresholds := []float64{0.1, 0.5, 1.0, 2.0, 5.0}
	
	for _, threshold := range thresholds {
		thresh_test, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "diagnostic_neuron", Threshold: threshold})
		if err != nil {
			continue
		}
		
		thresh_test.Start()
		if receiver, ok := thresh_test.(component.ChemicalReceiver); ok {
			matrix.RegisterForBinding(receiver)
		}
		
		time.Sleep(500 * time.Millisecond)
		activity := thresh_test.GetActivityLevel()
		t.Logf("Threshold %.1f: Baseline activity = %.3f", threshold, activity)
		
		thresh_test.Stop()
	}

	// Test 5: Activity decay behavior
	t.Log("\n--- TEST 5: ACTIVITY DECAY BEHAVIOR ---")
	
	decay_test, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "diagnostic_neuron", Threshold: 0.8})
	if err != nil {
		t.Fatalf("Failed to create decay test neuron: %v", err)
	}
	
	decay_test.Start()
	defer decay_test.Stop()
	
	if receiver, ok := decay_test.(component.ChemicalReceiver); ok {
		matrix.RegisterForBinding(receiver)
	}

	// Stimulate strongly
	decay_test.Receive(types.NeuralSignal{
		Value:     5.0,
		Timestamp: time.Now(),
		SourceID:  "decay_test",
		TargetID:  decay_test.ID(),
	})

	// Track decay over time
	t.Log("Activity decay over time:")
	for i := 0; i < 20; i++ {
		activity := decay_test.GetActivityLevel()
		t.Logf("  %dms: %.3f", i*100, activity)
		time.Sleep(100 * time.Millisecond)
	}

	// Analysis
	t.Log("\n--- DIAGNOSTIC ANALYSIS ---")
	
	finalIsolated := isolated.GetActivityLevel()
	finalDecay := decay_test.GetActivityLevel()
	
	if finalIsolated > 0.1 {
		t.Log("❌ PROBLEM: Isolated neurons show persistent baseline activity")
		t.Log("   This suggests intrinsic excitability or accumulation")
	} else {
		t.Log("✅ Isolated neurons can maintain silence")
	}
	
	if finalDecay > 1.0 {
		t.Log("❌ PROBLEM: Activity does not decay to baseline")
		t.Log("   This suggests incomplete decay mechanisms")
	} else {
		t.Log("✅ Activity decays appropriately")
	}
	
	t.Logf("Final states: Isolated=%.3f, Decay=%.3f", finalIsolated, finalDecay)
	
	// Root cause hypothesis
	t.Log("\n--- ROOT CAUSE HYPOTHESIS ---")
	if finalIsolated > 0.1 && finalDecay > 1.0 {
		t.Log("🔍 HYPOTHESIS: Neurons have intrinsic excitability that prevents silence")
		t.Log("   - Baseline membrane potential may be above threshold")
		t.Log("   - Accumulator may not fully reset")
		t.Log("   - Ion channels may provide persistent current")
	} else if finalDecay > 1.0 {
		t.Log("🔍 HYPOTHESIS: Stimulation creates persistent state changes")
		t.Log("   - Synaptic weights or ion channel states persist")
		t.Log("   - Learning mechanisms create permanent excitability")
	} else {
		t.Log("🔍 HYPOTHESIS: Problem may be in circuit architecture or timing")
	}
}