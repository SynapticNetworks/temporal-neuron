/*
=================================================================================
SYNAPSE SYSTEM UNIT TESTS - REFACTORED
=================================================================================

OVERVIEW:
This file contains comprehensive unit tests for the new, refactored synapse
system. It verifies the functionality of the `EnhancedSynapse` and its
integration with its biological sub-components.

TEST PHILOSOPHY:
- Tests use the new factory pattern (`CreateSynapse`) for construction.
- Mock callbacks are used to simulate the matrix environment and observe outputs.
- Tests cover successful operation, failure modes (e.g., vesicle depletion),
  and biological rules (plasticity, pruning).

BIOLOGICAL CONTEXT VALIDATED:
- Correct initialization of all sub-components.
- The full transmission sequence: vesicle check -> delay calculation -> delivery.
- Failure of transmission when vesicles are depleted.
- STDP logic delegating to the PlasticityCalculator.
- Pruning logic based on activity and weight thresholds.
*/

package synapse

import (
	"sync"
	"testing"
	"time"
)

// mockCallbacks provides a controllable implementation of SynapseCallbacks for testing.
// It allows tests to inspect what functions were called and what data was passed.
type mockCallbacks struct {
	mu                      sync.Mutex
	deliveredMessages       []SynapseMessage
	getCalciumLevel         func() float64
	getTransmissionDelay    func() time.Duration
	releaseNeurotransmitter func(ligandType LigandType, concentration float64) error
	reportPlasticityEvent   func(event PlasticityEvent)
}

// newMockCallbacks creates a set of default mock callbacks.
func newMockCallbacks() *mockCallbacks {
	return &mockCallbacks{
		deliveredMessages: make([]SynapseMessage, 0),
		getCalciumLevel: func() float64 {
			return 1.0 // Default baseline calcium
		},
		getTransmissionDelay: func() time.Duration {
			return 2 * time.Millisecond // Default test delay
		},
	}
}

// DeliverMessage captures the message for later inspection by the test.
func (mc *mockCallbacks) DeliverMessage(targetID string, message SynapseMessage) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.deliveredMessages = append(mc.deliveredMessages, message)
	return nil
}

// setupTestSynapse is a helper function to reduce boilerplate in tests.
// It creates a standard synapse with mock callbacks for testing.
func setupTestSynapse(t *testing.T) (*EnhancedSynapse, SynapseConfig, *mockCallbacks) {
	// Use a preset to get a valid base configuration
	config := CreateExcitatoryGlutamatergicConfig("syn-001", "neuron-pre", "neuron-post")
	callbacks := newMockCallbacks()

	// Use the actual factory to create the synapse
	processor, err := CreateSynapse("syn-001", config, SynapseCallbacks{
		GetCalciumLevel:      callbacks.getCalciumLevel,
		GetTransmissionDelay: callbacks.getTransmissionDelay,
		DeliverMessage:       callbacks.DeliverMessage,
	})

	if err != nil {
		t.Fatalf("Failed to create synapse for testing: %v", err)
	}

	synapse, ok := processor.(*EnhancedSynapse)
	if !ok {
		t.Fatalf("Created processor is not of type *EnhancedSynapse")
	}

	return synapse, config, callbacks
}

// =================================================================================
// SYNAPSE CREATION AND INITIALIZATION TESTS
// =================================================================================

// TestSynapseCreationAndInitialization verifies that the factory correctly
// constructs an EnhancedSynapse with all its sub-components.
func TestSynapseCreationAndInitialization(t *testing.T) {
	synapse, config, _ := setupTestSynapse(t)

	if synapse.ID() != config.SynapseID {
		t.Errorf("Expected synapse ID '%s', got '%s'", config.SynapseID, synapse.ID())
	}

	if synapse.GetWeight() != config.InitialWeight {
		t.Errorf("Expected initial weight %f, got %f", config.InitialWeight, synapse.GetWeight())
	}

	// Verify that sub-components were initialized
	if synapse.activityMonitor == nil {
		t.Error("Activity monitor was not initialized")
	}
	if synapse.plasticityCalculator == nil {
		t.Error("Plasticity calculator was not initialized")
	}
	if config.VesicleConfig.Enabled && synapse.vesicleSystem == nil {
		t.Error("Vesicle system was not initialized despite being enabled in config")
	}

	// Check that the initial weight was recorded by the activity monitor
	activityInfo := synapse.GetActivityInfo()
	if activityInfo.TotalPlasticityEvents == 0 {
		t.Error("Expected initial weight to be recorded as a plasticity event, but none found")
	}
}

// =================================================================================
// SIGNAL TRANSMISSION TESTS
// =================================================================================

// TestSuccessfulTransmission verifies the complete, successful transmission pathway.
func TestSuccessfulTransmission(t *testing.T) {
	synapse, config, callbacks := setupTestSynapse(t)

	inputSignal := 1.0
	err := synapse.Transmit(inputSignal)
	if err != nil {
		t.Fatalf("Transmit failed unexpectedly: %v", err)
	}

	// VERIFICATION 1: Check that the message was delivered via the callback
	if len(callbacks.deliveredMessages) != 1 {
		t.Fatalf("Expected 1 message to be delivered, got %d", len(callbacks.deliveredMessages))
	}
	msg := callbacks.deliveredMessages[0]

	// VERIFICATION 2: Check signal scaling
	expectedValue := inputSignal * config.InitialWeight
	if msg.Value != expectedValue {
		t.Errorf("Expected message value %.2f, got %.2f", expectedValue, msg.Value)
	}

	// VERIFICATION 3: Check that the delay was applied via the callback
	if msg.TransmissionDelay != callbacks.getTransmissionDelay() {
		t.Errorf("Expected transmission delay %v, got %v", callbacks.getTransmissionDelay(), msg.TransmissionDelay)
	}

	// VERIFICATION 4: Check activity monitor for a successful event
	activityInfo := synapse.GetActivityInfo()
	if activityInfo.TotalTransmissions != 1 || activityInfo.SuccessfulTransmissions != 1 {
		t.Errorf("Expected 1 successful transmission to be logged, got %d total and %d successful",
			activityInfo.TotalTransmissions, activityInfo.SuccessfulTransmissions)
	}
}

// TestTransmissionFailureOnVesicleDepletion verifies that transmission fails
// when the vesicle pool is depleted.
func TestTransmissionFailureOnVesicleDepletion(t *testing.T) {
	synapse, config, callbacks := setupTestSynapse(t)

	// Deplete the ready releasable pool of vesicles
	readyPoolSize := config.VesicleConfig.ReadyPoolSize
	for i := 0; i < readyPoolSize; i++ {
		err := synapse.Transmit(1.0)
		if err != nil {
			t.Fatalf("Transmission failed prematurely on attempt %d: %v", i+1, err)
		}
	}

	// The next transmission attempt should fail
	err := synapse.Transmit(1.0)
	if err == nil {
		t.Fatal("Expected transmission to fail due to vesicle depletion, but it succeeded")
	}

	if err != ErrVesicleDepleted {
		t.Errorf("Expected error ErrVesicleDepleted, got %v", err)
	}

	// Verify no new message was delivered on the failed attempt
	if len(callbacks.deliveredMessages) != readyPoolSize {
		t.Errorf("Expected %d delivered messages, but got %d after depletion",
			readyPoolSize, len(callbacks.deliveredMessages))
	}

	// Verify the activity monitor logged the failure
	info := synapse.GetActivityInfo()
	expectedTotal := int64(readyPoolSize + 1)
	expectedSuccess := int64(readyPoolSize)
	if info.TotalTransmissions != expectedTotal || info.SuccessfulTransmissions != expectedSuccess {
		t.Errorf("Activity monitor log is incorrect. Expected %d total and %d successful, got %d and %d.",
			expectedTotal, expectedSuccess, info.TotalTransmissions, info.SuccessfulTransmissions)
	}
}

// =================================================================================
// PLASTICITY AND WEIGHT MANAGEMENT TESTS
// =================================================================================

// TestPlasticityApplication verifies that STDP correctly modifies the synapse weight.
func TestPlasticityApplication(t *testing.T) {
	synapse, _, _ := setupTestSynapse(t)
	initialWeight := synapse.GetWeight()

	// 1. Test LTP (causal: pre-before-post)
	ltpAdjustment := PlasticityAdjustment{DeltaT: -15 * time.Millisecond}
	err := synapse.ApplyPlasticity(ltpAdjustment)
	if err != nil {
		t.Fatalf("ApplyPlasticity failed for LTP: %v", err)
	}

	weightAfterLTP := synapse.GetWeight()
	if weightAfterLTP <= initialWeight {
		t.Errorf("Expected weight to increase for LTP. Initial: %.4f, After LTP: %.4f", initialWeight, weightAfterLTP)
	}

	// 2. Test LTD (anti-causal: post-before-pre)
	ltdAdjustment := PlasticityAdjustment{DeltaT: 15 * time.Millisecond}
	err = synapse.ApplyPlasticity(ltdAdjustment)
	if err != nil {
		t.Fatalf("ApplyPlasticity failed for LTD: %v", err)
	}

	weightAfterLTD := synapse.GetWeight()
	if weightAfterLTD >= weightAfterLTP {
		t.Errorf("Expected weight to decrease for LTD. After LTP: %.4f, After LTD: %.4f", weightAfterLTP, weightAfterLTD)
	}
}

// TestWeightBounds verifies that weights are clamped to the min/max values from the config.
func TestWeightBounds(t *testing.T) {
	synapse, config, _ := setupTestSynapse(t)

	// Test upper bound
	synapse.SetWeight(config.STDPConfig.MaxWeight + 10.0)
	if synapse.GetWeight() != config.STDPConfig.MaxWeight {
		t.Errorf("Expected weight to be clamped to max %.4f, got %.4f",
			config.STDPConfig.MaxWeight, synapse.GetWeight())
	}

	// Test lower bound
	synapse.SetWeight(config.STDPConfig.MinWeight - 10.0)
	if synapse.GetWeight() != config.STDPConfig.MinWeight {
		t.Errorf("Expected weight to be clamped to min %.4f, got %.4f",
			config.STDPConfig.MinWeight, synapse.GetWeight())
	}
}

// =================================================================================
// STRUCTURAL PLASTICITY (PRUNING) TESTS
// =================================================================================

// TestPruningLogic verifies the "use it or lose it" rule.
func TestPruningLogic(t *testing.T) {
	config := CreateExcitatoryGlutamatergicConfig("syn-prune", "pre", "post")
	// Make pruning easy to trigger for the test
	config.PruningConfig.Enabled = true
	config.PruningConfig.WeightThreshold = 0.1
	config.PruningConfig.InactivityThreshold = 10 * time.Millisecond
	config.PruningConfig.ProtectionPeriod = 1 * time.Millisecond

	// Create the synapse
	processor, _ := CreateSynapse(config.SynapseID, config, SynapseCallbacks{})
	synapse := processor.(*EnhancedSynapse)
	time.Sleep(2 * time.Millisecond) // Wait for protection period to pass

	// Condition 1: Synapse is active and strong, should NOT be pruned
	synapse.SetWeight(0.5)
	_ = synapse.Transmit(1.0)
	if synapse.ShouldPrune() {
		t.Error("Strong, active synapse should not be pruned.")
	}

	// Condition 2: Synapse becomes weak but is still active, should NOT be pruned yet.
	synapse.SetWeight(config.PruningConfig.WeightThreshold / 2)
	_ = synapse.Transmit(1.0)
	if synapse.ShouldPrune() {
		t.Error("Weak but active synapse should not be pruned immediately.")
	}

	// Condition 3: Synapse is weak AND becomes inactive, should BE pruned.
	time.Sleep(config.PruningConfig.InactivityThreshold * 2) // Wait for inactivity
	if !synapse.ShouldPrune() {
		t.Error("Weak and inactive synapse should be marked for pruning.")
	}
}
