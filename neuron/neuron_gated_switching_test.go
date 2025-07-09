package neuron

import (
	"context"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

/*
=================================================================================
NEURON ION CHANNEL GATING TESTS - LIGAND-BASED PATHWAY SWITCHING
=================================================================================

This test validates the core "gating" paradigm by using a specific, realistic
ion channel (`RealisticGabaAChannel`) to dynamically switch a neural pathway off.
A control neuron releases a ligand (GABA), which activates the channel on a
target neuron's dendrite, thereby inhibiting it and rerouting the signal flow.

=================================================================================
*/

// TestNeuron_IonChannelGatedSwitching demonstrates a control neuron using a ligand
// to gate a pathway off by activating a realistic ion channel on the target's dendrite.
func TestNeuron_IonChannelGatedSwitching(t *testing.T) {
	t.Log("=== TESTING Ion Channel-Based Gated Switching ===")

	// --- Test Setup ---
	mockMatrix := NewMockMatrix()

	// Create the neurons
	inputNeuron := NewNeuron("input_neuron", 0.5, 0.9, 2*time.Millisecond, 1.0, 5.0, 0)
	controlNeuron := NewNeuron("control_neuron", 0.5, 0.9, 2*time.Millisecond, 1.0, 5.0, 0)
	targetA := NewNeuron("target_A", 0.8, 0.9, 2*time.Millisecond, 1.0, 5.0, 0)
	targetB := NewNeuron("target_B", 0.8, 0.9, 2*time.Millisecond, 1.0, 5.0, 0)

	allNeurons := []*Neuron{inputNeuron, controlNeuron, targetA, targetB}
	for _, n := range allNeurons {
		n.SetCallbacks(mockMatrix.CreateBasicCallbacks())
	}

	// --- CRITICAL: Configure Dendrites and Ion Channels ---
	dendriteA := NewBiologicalTemporalSummationMode(CreateCorticalPyramidalConfig())
	gabaChannel := NewRealisticGabaAChannel("gaba_gate_on_A")
	dendriteA.AddChannel(gabaChannel)
	targetA.SetDendriticMode(dendriteA)

	observerA := NewMockSynapse("observer_A", "output_A", 1.0, 1*time.Millisecond)
	observerB := NewMockSynapse("observer_B", "output_B", 1.0, 1*time.Millisecond)
	targetA.AddOutputCallback(observerA.id, observerA.CreateOutputCallback())
	targetB.AddOutputCallback(observerB.id, observerB.CreateOutputCallback())

	// --- Connect the network using fully initialized OutputCallbacks ---
	inputCallbackA := types.OutputCallback{
		TransmitMessage: func(msg types.NeuralSignal) error { targetA.Receive(msg); return nil },
		GetWeight:       func() float64 { return 1.0 }, GetDelay: func() time.Duration { return 1 * time.Millisecond }, GetTargetID: func() string { return targetA.ID() },
	}
	inputCallbackB := types.OutputCallback{
		TransmitMessage: func(msg types.NeuralSignal) error { targetB.Receive(msg); return nil },
		GetWeight:       func() float64 { return 1.0 }, GetDelay: func() time.Duration { return 1 * time.Millisecond }, GetTargetID: func() string { return targetB.ID() },
	}
	controlCallbackA := types.OutputCallback{
		TransmitMessage: func(msg types.NeuralSignal) error {
			gabaSignal := types.NeuralSignal{
				Value: -2.0, NeurotransmitterType: types.LigandGABA, SourceID: controlNeuron.ID(), TargetID: targetA.ID(), Timestamp: time.Now(),
			}
			targetA.Receive(gabaSignal)
			return nil
		},
		GetWeight: func() float64 { return 1.0 }, GetDelay: func() time.Duration { return 1 * time.Millisecond }, GetTargetID: func() string { return targetA.ID() },
	}

	inputNeuron.AddOutputCallback("input_to_A", inputCallbackA)
	inputNeuron.AddOutputCallback("input_to_B", inputCallbackB)
	controlNeuron.AddOutputCallback("control_to_A", controlCallbackA)

	for _, n := range allNeurons {
		if err := n.Start(); err != nil {
			t.Fatalf("Failed to start neuron %s: %v", n.ID(), err)
		}
		defer n.Stop()
	}

	// --- Phase 1: Default State (Both Paths ON) ---
	t.Log("--- Phase 1: Firing input neuron. Both A and B should fire. ---")
	SendTestSignal(inputNeuron, "input_stimulus_1", 1.0)
	time.Sleep(50 * time.Millisecond)

	if observerA.GetReceivedSignalCount() != 1 || observerB.GetReceivedSignalCount() != 1 {
		t.Fatalf("Phase 1 FAILED: Expected both targets to fire once. Got A: %d, B: %d", observerA.GetReceivedSignalCount(), observerB.GetReceivedSignalCount())
	}
	t.Log("✓ Default state confirmed: Both pathways are active.")

	observerA.receivedSignals = nil
	observerB.receivedSignals = nil

	// --- Phase 2: Gating Path A "OFF" with a Sustained Ligand Tone ---
	t.Log("--- Phase 2: Applying sustained GABA inhibition to gate OFF path A. ---")

	ctx, cancel := context.WithCancel(context.Background())
	// Start a goroutine to provide a continuous inhibitory tone to Target A
	go func() {
		ticker := time.NewTicker(5 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				SendTestSignal(controlNeuron, "control_stimulus", 1.0)
			case <-ctx.Done():
				return
			}
		}
	}()

	time.Sleep(20 * time.Millisecond) // Allow the inhibitory tone to establish

	// Now, send the main input signal while the inhibition is active.
	SendTestSignal(inputNeuron, "input_stimulus_2", 1.0)
	time.Sleep(50 * time.Millisecond)

	cancel() // Stop the inhibitory tone

	// CORRECTED: Assert the switch with proper if/else logic for clear reporting
	if observerA.GetReceivedSignalCount() > 0 {
		t.Errorf("Phase 2 FAILED: Target A fired, but its GABA gate should have inhibited it. Fired %d times.", observerA.GetReceivedSignalCount())
	} else if observerB.GetReceivedSignalCount() != 1 {
		t.Errorf("Phase 2 FAILED: Expected Target B to fire normally, but it fired %d times.", observerB.GetReceivedSignalCount())
	} else {
		t.Log("✓ Gating successful! Pathway to Target A was switched OFF via its ion channel, while Target B remained ON.")
	}
}
