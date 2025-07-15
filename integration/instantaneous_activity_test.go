package integration

import (
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/extracellular"
	"github.com/SynapticNetworks/temporal-neuron/neuron"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// TestInstantaneousVsFiringRate tests the difference between membrane potential and firing rate
func TestInstantaneousVsFiringRate(t *testing.T) {
	t.Log("=== INSTANTANEOUS ACTIVITY vs FIRING RATE TEST ===")
	t.Log("Testing the difference between membrane potential and historical firing rate")

	// Create matrix
	matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  1 * time.Millisecond,
		MaxComponents:   5,
	})

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}
	defer matrix.Stop()

	// Register neuron with very short activity window for testing
	matrix.RegisterNeuronType("test_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		// Create neuron with short activity window (1 second instead of 10)
		n := neuron.NewNeuron(id, config.Threshold, 0.9, 5*time.Millisecond, 1.0, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate})
		n.SetCallbacks(callbacks)
		return n, nil
	})

	// Create a test neuron
	testNeuron, err := matrix.CreateNeuron(types.NeuronConfig{NeuronType: "test_neuron", Threshold: 1.0})
	if err != nil {
		t.Fatalf("Failed to create neuron: %v", err)
	}

	testNeuron.Start()
	if receiver, ok := testNeuron.(component.ChemicalReceiver); ok {
		matrix.RegisterForBinding(receiver)
	}
	defer testNeuron.Stop()

	t.Log("\n--- PHASE 1: BASELINE (No Stimulation) ---")
	
	// Check baseline
	baselineFiringRate := testNeuron.GetActivityLevel()
	t.Logf("Baseline firing rate: %.3f Hz", baselineFiringRate)

	t.Log("\n--- PHASE 2: STIMULATION ---")
	
	// Stimulate the neuron
	testNeuron.Receive(types.NeuralSignal{
		Value:     2.0, // Above threshold
		Timestamp: time.Now(),
		SourceID:  "test",
		TargetID:  testNeuron.ID(),
	})

	// Check immediate response
	time.Sleep(10 * time.Millisecond) // Allow signal processing
	immediateRate := testNeuron.GetActivityLevel()
	t.Logf("Immediate firing rate after stimulation: %.3f Hz", immediateRate)

	t.Log("\n--- PHASE 3: DECAY TRACKING ---")
	
	// Track decay over time
	decayTimes := []time.Duration{
		100 * time.Millisecond,
		500 * time.Millisecond,
		1 * time.Second,
		2 * time.Second,
		5 * time.Second,
		10 * time.Second,
	}

	for _, delay := range decayTimes {
		time.Sleep(delay - (100 * time.Millisecond)) // Subtract previous delays
		currentRate := testNeuron.GetActivityLevel()
		t.Logf("After %v: Firing rate = %.3f Hz", delay, currentRate)
		
		if currentRate < 0.01 {
			t.Logf("✅ Firing rate decayed to near-zero after %v", delay)
			break
		}
	}

	finalRate := testNeuron.GetActivityLevel()
	
	t.Log("\n--- ANALYSIS ---")
	
	t.Logf("Baseline: %.3f Hz", baselineFiringRate)
	t.Logf("After stimulation: %.3f Hz", immediateRate)
	t.Logf("Final: %.3f Hz", finalRate)
	
	if immediateRate > baselineFiringRate {
		t.Log("✅ Neuron responds to stimulation")
	} else {
		t.Log("❌ Neuron fails to respond to stimulation")
	}
	
	if finalRate < immediateRate * 0.1 {
		t.Log("✅ Firing rate decays significantly over time")
	} else {
		t.Log("❌ Firing rate persists (due to historical window)")
	}

	t.Log("\n--- ROOT CAUSE IDENTIFICATION ---")
	
	if finalRate > 0.05 {
		t.Log("🔍 PROBLEM: GetActivityLevel() returns historical firing rate")
		t.Log("   - Once a neuron fires, it shows activity for the entire window duration")
		t.Log("   - This prevents pattern discrimination in rapid sequences")
		t.Log("   - Need instantaneous membrane potential, not historical rate")
		
		t.Log("\n💡 SOLUTION NEEDED:")
		t.Log("   1. Access membrane potential directly (not firing rate)")
		t.Log("   2. Use very short activity windows for pattern discrimination")
		t.Log("   3. Or create separate instantaneous activity measurement")
		
	} else {
		t.Log("✅ Firing rate decays appropriately for pattern discrimination")
	}

	// Check if we can get direct membrane access
	t.Log("\n--- ATTEMPTING DIRECT MEMBRANE ACCESS ---")
	
	// Look for status information that might give us membrane potential
	if statusProvider, ok := testNeuron.(interface{ GetProcessingStatus() map[string]interface{} }); ok {
		status := statusProvider.GetProcessingStatus()
		if neuralState, ok := status["neural_state"].(map[string]interface{}); ok {
			if accumulator, ok := neuralState["accumulator"].(float64); ok {
				if threshold, ok := neuralState["threshold"].(float64); ok {
					t.Logf("Direct membrane access available:")
					t.Logf("  Accumulator: %.3f", accumulator)
					t.Logf("  Threshold: %.3f", threshold)
					t.Logf("  Membrane ratio: %.3f", accumulator/threshold)
					
					if accumulator < threshold * 0.1 {
						t.Log("✅ Membrane potential properly decayed")
					} else {
						t.Log("❌ Membrane potential still elevated")
					}
				}
			}
		}
	} else {
		t.Log("❌ No direct membrane access available")
	}
}