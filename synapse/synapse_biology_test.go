/*
=================================================================================
BIOLOGICAL SYNAPSE TESTS - REFACTORED
=================================================================================

OVERVIEW:
This file contains tests that validate the biological realism of the refactored
`EnhancedSynapse`. These tests ensure that the synaptic behavior matches what is
observed in real biological neural networks.

EXPERIMENTAL VALIDATION:
The tests are updated to use the new factory-based creation and callback-driven
architecture, while still verifying the same core biological principles:
- The classic STDP learning window.
- Activity-dependent structural plasticity ("use it or lose it").
- Biologically plausible transmission delays and weight scaling.
*/

package synapse

import (
	"math"
	"testing"
	"time"
)

// setupBiologyTestSynapse is a helper for creating synapses with custom configs for biology tests.
// IT NOW RETURNS THE MOCK NEURON so tests can inspect it.
func setupBiologyTestSynapse(t *testing.T, stdpConfig STDPConfig, pruningConfig PruningConfig, initialWeight float64, delay time.Duration) (*EnhancedSynapse, *MockNeuron) {
	if delay < BIOLOGICAL_MIN_DELAY {
		delay = BIOLOGICAL_MIN_DELAY
	}

	// Create a mock neuron to act as the target.
	postNeuron := NewMockNeuron("neuron-post")

	config := SynapseConfig{
		SynapseID:         "bio-test-synapse",
		SynapseType:       "excitatory_glutamatergic",
		PresynapticID:     "neuron-pre",
		PostsynapticID:    postNeuron.ID(),
		InitialWeight:     initialWeight,
		BaseSynapticDelay: delay,
		VesicleConfig:     CreateExcitatoryGlutamatergicConfig("", "", "").VesicleConfig,
		STDPConfig:        stdpConfig,
		PruningConfig:     pruningConfig,
	}
	config.VesicleConfig.Enabled = false

	// The mock neuron's Receive method is used as the callback.
	processor, err := CreateSynapse(config.SynapseID, config, SynapseCallbacks{
		DeliverMessage: postNeuron.Receive,
	})
	if err != nil {
		t.Fatalf("setupBiologyTestSynapse failed: %v", err)
	}

	synapse, ok := processor.(*EnhancedSynapse)
	if !ok {
		t.Fatalf("Created processor is not of type *EnhancedSynapse")
	}
	// Return the synapse and the mock neuron.
	return synapse, postNeuron
}

// =================================================================================
// SPIKE-TIMING DEPENDENT PLASTICITY (STDP) BIOLOGICAL TESTS
// =================================================================================

// TestSTDPClassicTimingWindow verifies the classic STDP learning window using the new synapse model.
func TestSTDPClassicTimingWindow(t *testing.T) {
	stdpConfig := STDPConfig{
		Enabled:        true,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     100 * time.Millisecond,
		MinWeight:      0.001,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.2,
	}
	pruningConfig := CreateDefaultPruningConfig()

	testCases := []struct {
		name           string
		timeDifference time.Duration
		expectedSign   float64
		description    string
	}{
		{"StrongLTP", -10 * time.Millisecond, 1.0, "Pre-synaptic spike 10ms before post-synaptic"},
		{"WeakLTP", -50 * time.Millisecond, 1.0, "Pre-synaptic spike 50ms before post-synaptic"},
		{"WeakLTD", 10 * time.Millisecond, -1.0, "Pre-synaptic spike 10ms after post-synaptic"},
		{"StrongLTD", 30 * time.Millisecond, -1.0, "Pre-synaptic spike 30ms after post-synaptic"},
		{"NoPlasticity", 150 * time.Millisecond, 0.0, "Timing difference outside STDP window"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			synapse, _ := setupBiologyTestSynapse(t, stdpConfig, pruningConfig, 1.0, BIOLOGICAL_MIN_DELAY)
			weightBefore := synapse.GetWeight()

			adjustment := PlasticityAdjustment{DeltaT: tc.timeDifference}
			synapse.ApplyPlasticity(adjustment)
			weightChange := synapse.GetWeight() - weightBefore

			var sign float64
			if weightChange > 1e-9 {
				sign = 1.0
			} else if weightChange < -1e-9 {
				sign = -1.0
			}

			if sign != tc.expectedSign {
				t.Errorf("Expected sign of change to be %.1f, but got %.1f (change: %f) for %s",
					tc.expectedSign, sign, weightChange, tc.description)
			}
		})
	}
}

// TestSTDPExponentialDecay verifies that plasticity effects decay exponentially with time.
func TestSTDPExponentialDecay(t *testing.T) {
	stdpConfig := STDPConfig{
		Enabled:        true,
		LearningRate:   0.02,
		TimeConstant:   15 * time.Millisecond,
		WindowSize:     80 * time.Millisecond,
		MinWeight:      0.001,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.0,
	}
	pruningConfig := CreateDefaultPruningConfig()

	timeDifferences := []time.Duration{
		-5 * time.Millisecond,
		-15 * time.Millisecond,
		-30 * time.Millisecond,
	}
	var lastChange float64 = math.MaxFloat64

	for _, deltaT := range timeDifferences {
		synapse, _ := setupBiologyTestSynapse(t, stdpConfig, pruningConfig, 1.0, BIOLOGICAL_MIN_DELAY)
		adjustment := PlasticityAdjustment{DeltaT: deltaT}
		synapse.ApplyPlasticity(adjustment)

		change := synapse.GetWeight() - 1.0
		if change >= lastChange {
			t.Errorf("STDP effect did not decay. Previous change: %f, current change: %f for DeltaT %v", lastChange, change, deltaT)
		}
		lastChange = change
	}
}

// =================================================================================
// STRUCTURAL PLASTICITY BIOLOGICAL TESTS
// =================================================================================

// TestActivityDependentPruning validates the "use it or lose it" principle.
func TestActivityDependentPruning(t *testing.T) {
	stdpConfig := CreateDefaultSTDPConfig()
	pruningConfig := PruningConfig{
		Enabled:             true,
		WeightThreshold:     0.1,
		InactivityThreshold: 50 * time.Millisecond,
		ProtectionPeriod:    10 * time.Millisecond,
	}

	t.Run("WeakInactiveSynapsePruning", func(t *testing.T) {
		synapse, _ := setupBiologyTestSynapse(t, stdpConfig, pruningConfig, 0.05, BIOLOGICAL_MIN_DELAY)
		time.Sleep(pruningConfig.ProtectionPeriod * 2)

		synapse.mu.Lock()
		synapse.lastTransmission = time.Now().Add(-pruningConfig.InactivityThreshold * 2)
		synapse.mu.Unlock()

		if !synapse.ShouldPrune() {
			t.Error("Weak and inactive synapse should be marked for pruning")
		}
	})

	t.Run("WeakButActiveSynapseProtection", func(t *testing.T) {
		synapse, _ := setupBiologyTestSynapse(t, stdpConfig, pruningConfig, 0.05, BIOLOGICAL_MIN_DELAY)
		time.Sleep(pruningConfig.ProtectionPeriod * 2)

		synapse.Transmit(1.0)

		if synapse.ShouldPrune() {
			t.Error("Weak but recently active synapse should NOT be pruned")
		}
	})
}

// =================================================================================
// SYNAPTIC TRANSMISSION BIOLOGICAL TESTS
// =================================================================================

// TestTransmissionDelayAccuracy now uses a mock callback that simulates the delay.
func TestTransmissionDelayAccuracy(t *testing.T) {
	stdpConfig := CreateDefaultSTDPConfig()
	delay := 20 * time.Millisecond
	tolerance := 15 * time.Millisecond

	config := SynapseConfig{
		SynapseID:         "delay-test",
		SynapseType:       "excitatory_glutamatergic",
		InitialWeight:     1.0,
		BaseSynapticDelay: delay,
		STDPConfig:        stdpConfig,
		PruningConfig:     CreateDefaultPruningConfig(),
	}

	deliveredChan := make(chan SynapseMessage, 1)
	mockDelivery := func(targetID string, message SynapseMessage) error {
		time.AfterFunc(message.TransmissionDelay, func() {
			deliveredChan <- message
		})
		return nil
	}

	processor, err := CreateSynapse(config.SynapseID, config, SynapseCallbacks{DeliverMessage: mockDelivery})
	if err != nil {
		t.Fatalf("CreateSynapse failed unexpectedly: %v", err)
	}
	synapse := processor.(*EnhancedSynapse)

	startTime := time.Now()
	synapse.Transmit(1.0)

	select {
	case <-deliveredChan:
		actualDelay := time.Since(startTime)
		if actualDelay < delay {
			t.Errorf("Message arrived too early. Expected >=%v, got %v", delay, actualDelay)
		}
	case <-time.After(delay + tolerance):
		t.Fatalf("Timed out waiting for message. Expected delivery around %v", delay)
	}
}

// TestSynapticWeightScaling validates signal scaling by synaptic weight.
func TestSynapticWeightScaling(t *testing.T) {
	weight := 0.5
	synapse, postNeuron := setupBiologyTestSynapse(t, CreateDefaultSTDPConfig(), CreateDefaultPruningConfig(), weight, BIOLOGICAL_MIN_DELAY)

	inputSignal := 1.0
	synapse.Transmit(inputSignal)

	receivedMessages := postNeuron.GetReceivedMessages()
	if len(receivedMessages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(receivedMessages))
	}
	expectedOutput := inputSignal * weight
	if math.Abs(receivedMessages[0].Value-expectedOutput) > 1e-9 {
		t.Errorf("Incorrect weight scaling. Expected %.2f, got %.2f", expectedOutput, receivedMessages[0].Value)
	}
}

// =================================================================================
// INTEGRATED BIOLOGICAL BEHAVIOR TESTS
// =================================================================================

// TestRealisticSynapticDynamics validates integrated behavior under a learning scenario.
func TestRealisticSynapticDynamics(t *testing.T) {
	stdpConfig := STDPConfig{
		Enabled: true, LearningRate: 0.005, TimeConstant: 15 * time.Millisecond,
		WindowSize: 60 * time.Millisecond, MinWeight: 0.01, MaxWeight: 3.0, AsymmetryRatio: 1.3,
	}
	pruningConfig := PruningConfig{Enabled: true, WeightThreshold: 0.05, InactivityThreshold: 30 * time.Second}
	initialWeight := 0.5

	synapse, postNeuron := setupBiologyTestSynapse(t, stdpConfig, pruningConfig, initialWeight, 2*time.Millisecond)

	numPairings := 50
	for i := 0; i < numPairings; i++ {
		synapse.Transmit(1.0)
		synapse.ApplyPlasticity(PlasticityAdjustment{DeltaT: -8 * time.Millisecond})
	}

	finalWeight := synapse.GetWeight()
	if finalWeight <= initialWeight {
		t.Errorf("Expected weight to increase from repeated LTP. Start: %.3f, End: %.3f", initialWeight, finalWeight)
	}

	// FIX: Use the postNeuron object to clear and get messages.
	postNeuron.ClearMessages()
	synapse.Transmit(1.0)

	receivedMessages := postNeuron.GetReceivedMessages()
	if len(receivedMessages) != 1 {
		t.Fatalf("Synapse should remain functional after learning")
	}
	expectedSignal := 1.0 * finalWeight
	if math.Abs(receivedMessages[0].Value-expectedSignal) > 1e-9 {
		t.Errorf("Signal strength should reflect learned weight. Expected %.3f, got %.3f", expectedSignal, receivedMessages[0].Value)
	}

	if synapse.ShouldPrune() {
		t.Error("Active, strengthened synapse should not be pruned.")
	}
}
