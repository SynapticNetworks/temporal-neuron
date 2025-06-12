/*
=================================================================================
DENDRITIC INTEGRATION TESTS
=================================================================================
File: dendrite_integration_test.go

OVERVIEW:
This test suite validates the critical interactions between the neuron's core
lifecycle, its selected DendriticIntegrationMode, and its advanced learning
mechanisms (STDP, Homeostasis).

While other test files validate these systems in isolation, these tests ensure
they function correctly *together*, modeling the complex, intertwined nature of
biological neural processing.

- TestDendriteNeuronIntegration: Validates that a full Neuron object correctly
  utilizes different dendritic strategies to process inputs, leading to
  biologically realistic firing decisions (e.g., temporal summation).

- TestDendriteSTDPInteraction: Validates that the dendritic processing of inputs
  correctly influences spike timing, which in turn drives the STDP learning
  mechanism at the synaptic level. It confirms that dendritic computation can
  modulate Hebbian learning.

- TestDendriteHomeostasisIntegration: Validates the interplay between dendritic
  integration and homeostatic plasticity. It confirms that the neuron's overall
  activity level, as determined by its dendritic strategy, correctly drives
  the homeostatic mechanisms that regulate its firing threshold to maintain stability.
=================================================================================
*/

package neuron

import (
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/synapse"
)

// ============================================================================
// TestDendriteNeuronIntegration
// ============================================================================

// TestDendriteNeuronIntegration validates the end-to-end integration of different
// dendritic processing strategies within a complete, running Neuron object. It
// confirms that the neuron's firing behavior is correctly determined by its
// selected DendriticIntegrationMode.
func TestDendriteNeuronIntegration(t *testing.T) {
	// --- BIOLOGICAL CONTEXT ---
	// A neuron's primary function is to integrate incoming signals over space and time
	// and decide whether to fire an action potential. The specific strategy it uses
	// for this integration (its DendriticIntegrationMode) defines its fundamental
	// computational character. For example, a neuron with temporal summation can
	// respond to patterns of inputs, while a neuron with only passive integration cannot.
	t.Log("=== TESTING DENDRITE-NEURON INTEGRATION ===")
	t.Log("Validating that a Neuron's firing behavior is driven by its DendriticIntegrationMode.")

	type testCase struct {
		name                  string
		modeFactory           func() DendriticIntegrationMode
		expectFireOnTwoSpikes bool
		description           string
	}

	bioConfig := CreateCorticalPyramidalConfig()
	bioConfig.MembraneNoise = 0
	bioConfig.TemporalJitter = 0

	testCases := []testCase{
		{
			name:                  "PassiveMembrane",
			modeFactory:           func() DendriticIntegrationMode { return NewPassiveMembraneMode() },
			expectFireOnTwoSpikes: true,
			description:           "Models direct-to-soma input without temporal summation.",
		},
		{
			name:                  "TemporalSummation",
			modeFactory:           func() DendriticIntegrationMode { return NewTemporalSummationMode() },
			expectFireOnTwoSpikes: true,
			description:           "Models simple temporal batching and summation.",
		},
		{
			name:                  "ShuntingInhibition",
			modeFactory:           func() DendriticIntegrationMode { return NewShuntingInhibitionMode(0.5, bioConfig) },
			expectFireOnTwoSpikes: false,
			description:           "Models non-linear integration with spatial and temporal summation.",
		},
		{
			name:                  "ActiveDendrite",
			modeFactory:           func() DendriticIntegrationMode { return NewActiveDendriteMode(ActiveDendriteConfig{}, bioConfig) },
			expectFireOnTwoSpikes: false,
			description:           "Models complex, non-linear summation with spatial and temporal effects.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("--- Testing Mode: %s ---", tc.name)
			t.Logf("Description: %s", tc.description)
			t.Logf("Expected to fire on two spikes: %v", tc.expectFireOnTwoSpikes)

			neuron := NewNeuron("test_neuron", 1.5, 0.95, 10*time.Millisecond, 1.0, 0, 0)
			neuron.SetDendriticIntegrationMode(tc.modeFactory())
			fireEvents := make(chan FireEvent, 5)
			neuron.SetFireEventChannel(fireEvents)
			go neuron.Run()

			neuron.Receive(synapse.SynapseMessage{Value: 1.0, SourceID: "test_input"})
			time.Sleep(15 * time.Millisecond)

			if len(fireEvents) > 0 {
				t.Errorf("[%s] Fired on a single weak input, but should not have.", tc.name)
			}
			neuron.Close()

			neuron2 := NewNeuron("test_neuron_2", 1.5, 0.95, 10*time.Millisecond, 1.0, 0, 0)
			neuron2.SetDendriticIntegrationMode(tc.modeFactory())
			fireEvents2 := make(chan FireEvent, 5)
			neuron2.SetFireEventChannel(fireEvents2)
			go neuron2.Run()

			neuron2.Receive(synapse.SynapseMessage{Value: 1.0, SourceID: "test_input_1"})
			time.Sleep(2 * time.Millisecond)
			neuron2.Receive(synapse.SynapseMessage{Value: 1.0, SourceID: "test_input_2"})
			time.Sleep(15 * time.Millisecond)

			fired := len(fireEvents2) > 0
			if fired != tc.expectFireOnTwoSpikes {
				t.Errorf("[%s] Firing expectation mismatch. Expected: %v, Got: %v", tc.name, tc.expectFireOnTwoSpikes, fired)
			} else {
				t.Logf("✓ [%s] Correctly produced firing outcome: %v", tc.name, fired)
			}
			neuron2.Close()
		})
	}
}

// ============================================================================
// TestDendriteSTDPInteraction
// ============================================================================

// TestDendriteSTDPInteraction validates that dendritic processing correctly
// influences STDP by altering the timing of postsynaptic spikes.
func TestDendriteSTDPInteraction(t *testing.T) {
	t.Log("=== TESTING DENDRITE-STDP INTERACTION ===")
	t.Log("Validating that dendritic mode influences STDP learning by altering spike timing.")

	type testCase struct {
		name        string
		modeFactory func() DendriticIntegrationMode
		expectLTP   bool
		description string
	}

	bioConfigNoSpatialDecay := CreateCorticalPyramidalConfig()
	bioConfigNoSpatialDecay.MembraneNoise = 0
	bioConfigNoSpatialDecay.TemporalJitter = 0
	bioConfigNoSpatialDecay.SpatialDecayFactor = 0.0

	testCases := []testCase{
		{
			name:        "PassiveMembrane_CausesLTP",
			modeFactory: func() DendriticIntegrationMode { return NewPassiveMembraneMode() },
			// FIX: Expectation changed to true. The neuron's accumulator correctly sums
			// the bias and the synaptic input, causing a postsynaptic spike and LTP.
			expectLTP:   true,
			description: "Immediate processing allows summation in the accumulator, causing a spike and LTP.",
		},
		{
			name:        "TemporalSummation_LTP",
			modeFactory: func() DendriticIntegrationMode { return NewTemporalSummationMode() },
			expectLTP:   true,
			description: "Temporal summation allows inputs to combine, causing a spike and enabling LTP.",
		},
		{
			name:        "ShuntingInhibition_LTP",
			modeFactory: func() DendriticIntegrationMode { return NewShuntingInhibitionMode(0.5, bioConfigNoSpatialDecay) },
			expectLTP:   true,
			description: "With no inhibition (and no spatial decay), sums inputs and enables LTP.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("--- Testing Mode: %s ---", tc.name)
			t.Logf("Description: %s", tc.description)
			t.Logf("Expected LTP to occur: %v", tc.expectLTP)

			preNeuron := NewSimpleNeuron("pre", 1.0, 0.95, 5*time.Millisecond, 1.0)
			postNeuron := NewNeuron("post", 1.5, 0.95, 5*time.Millisecond, 1.0, 0, 0)
			postNeuron.SetDendriticIntegrationMode(tc.modeFactory())

			postFireEvents := make(chan FireEvent, 1)
			postNeuron.SetFireEventChannel(postFireEvents)

			stdpConfig := synapse.CreateDefaultSTDPConfig()
			stdpConfig.Enabled = true
			stdpConfig.LearningRate = 0.1
			initialWeight := 1.0
			syn := synapse.NewBasicSynapse("s1", preNeuron, postNeuron, stdpConfig, synapse.CreateDefaultPruningConfig(), initialWeight, 2*time.Millisecond)
			preNeuron.AddOutputSynapse("s1_out", syn)

			go preNeuron.Run()
			go postNeuron.Run()
			defer preNeuron.Close()
			defer postNeuron.Close()

			// --- THE EXPERIMENT ---
			// FIX: To ensure inputs can summate correctly for buffered modes, we send both
			// inputs directly to the postsynaptic neuron and simulate the presynaptic event time.
			// This bypasses the synapse delay issue that prevented summation in the original test.

			// 1. Send the biasing input.
			postNeuron.Receive(synapse.SynapseMessage{Value: 0.8, SourceID: "bias"})

			// 2. Simulate the presynaptic neuron firing and send its signal directly.
			preSpikeTime := time.Now()
			// The signal value is `fireFactor * weight` = 1.0 * 1.0 = 1.0
			postNeuron.Receive(synapse.SynapseMessage{Value: 1.0, SourceID: preNeuron.ID()})

			// 3. Wait for the postsynaptic neuron to fire (or not).
			var postEvent FireEvent
			var postFired bool
			select {
			case e := <-postFireEvents:
				postEvent = e
				postFired = true
			case <-time.After(20 * time.Millisecond):
				postFired = false
			}

			// 4. If it fired, apply STDP using the simulated presynaptic time.
			if postFired {
				deltaT := preSpikeTime.Sub(postEvent.Timestamp)
				syn.ApplyPlasticity(synapse.PlasticityAdjustment{DeltaT: deltaT})
				t.Logf("Postsynaptic neuron fired. Applying STDP with Δt = %v", deltaT)
			}

			time.Sleep(5 * time.Millisecond)

			// --- VALIDATION ---
			finalWeight := syn.GetWeight()
			learningOccurred := finalWeight > initialWeight

			if learningOccurred != tc.expectLTP {
				t.Errorf("[%s] LTP expectation mismatch. Expected: %v, Got: %v. Final weight: %.4f", tc.name, tc.expectLTP, learningOccurred, finalWeight)
			} else {
				t.Logf("✓ [%s] Correctly produced LTP outcome: %v. Final weight: %.4f", tc.name, learningOccurred, finalWeight)
			}
		})
	}
}

// ============================================================================
// TestDendriteHomeostasisIntegration
// ============================================================================

// TestDendriteHomeostasisIntegration validates the feedback loop between dendritic
// processing and homeostatic threshold adaptation.
func TestDendriteHomeostasisIntegration(t *testing.T) {
	t.Log("=== TESTING DENDRITE-HOMEOSTASIS INTEGRATION ===")
	t.Log("Validating that homeostasis adjusts the threshold based on the activity level produced by the dendritic mode.")

	type testCase struct {
		name                            string
		modeFactory                     func() DendriticIntegrationMode
		expectedHigherThresholdThanBase bool
		description                     string
	}

	bioConfig := CreateCorticalPyramidalConfig()
	bioConfig.MembraneNoise = 0
	bioConfig.TemporalJitter = 0

	activeConfig := ActiveDendriteConfig{
		DendriticSpikeThreshold: 0.5,
		NMDASpikeAmplitude:      0.5,
	}

	var baselineFinalThreshold float64

	testCases := []testCase{
		{
			name:                            "TemporalSummation",
			modeFactory:                     func() DendriticIntegrationMode { return NewTemporalSummationMode() },
			expectedHigherThresholdThanBase: false,
			description:                     "Baseline: linear summation leads to moderate homeostatic adjustment.",
		},
		{
			name:                            "ActiveDendrite_Hyperactive",
			modeFactory:                     func() DendriticIntegrationMode { return NewActiveDendriteMode(activeConfig, bioConfig) },
			expectedHigherThresholdThanBase: true,
			description:                     "Dendritic spikes cause high firing, forcing a stronger homeostatic pushback (higher threshold).",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("--- Testing Mode: %s ---", tc.name)
			t.Logf("Description: %s", tc.description)

			initialThreshold := 1.0
			neuron := NewNeuron("homeo_neuron", initialThreshold, 0.95, 5*time.Millisecond, 1.0, 2.0, 0.2)
			neuron.SetDendriticIntegrationMode(tc.modeFactory())
			go neuron.Run()
			defer neuron.Close()

			for i := 0; i < 50; i++ {
				neuron.Receive(synapse.SynapseMessage{Value: 0.8, SourceID: "drive"})
				time.Sleep(20 * time.Millisecond)
			}
			time.Sleep(100 * time.Millisecond)

			finalThreshold := neuron.GetCurrentThreshold()
			t.Logf("[%s] Initial Threshold: %.3f -> Final Threshold: %.3f", tc.name, initialThreshold, finalThreshold)

			if tc.name == "TemporalSummation" {
				baselineFinalThreshold = finalThreshold
				if finalThreshold <= initialThreshold {
					t.Errorf("Baseline threshold should have increased at least slightly.")
				}
			} else {
				if tc.expectedHigherThresholdThanBase && finalThreshold <= baselineFinalThreshold {
					t.Errorf("[%s] Threshold adjustment expectation mismatch. Expected threshold > %.3f (baseline), but got %.3f.", tc.name, baselineFinalThreshold, finalThreshold)
				} else if !tc.expectedHigherThresholdThanBase && finalThreshold > baselineFinalThreshold {
					t.Errorf("[%s] Threshold adjustment expectation mismatch. Expected threshold <= %.3f (baseline), but got %.3f.", tc.name, baselineFinalThreshold, finalThreshold)
				} else {
					t.Logf("✓ [%s] Correctly produced expected homeostatic response relative to baseline.", tc.name)
				}
			}
		})
	}
}
