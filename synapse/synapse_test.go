/*
=================================================================================
SYNAPSE SYSTEM UNIT TESTS - REFACTORED
=================================================================================

OVERVIEW:
This file contains comprehensive unit tests for the new, refactored synapse
system. It verifies the functionality of the `Synapse` and its
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
func setupTestSynapse(t *testing.T) (*Synapse, SynapseConfig, *mockCallbacks) {
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

	synapse, ok := processor.(*Synapse)
	if !ok {
		t.Fatalf("Created processor is not of type *Synapse")
	}

	return synapse, config, callbacks
}

// =================================================================================
// SYNAPSE CREATION AND INITIALIZATION TESTS
// =================================================================================

// TestSynapseCreationAndInitialization verifies that the factory correctly
// constructs an Synapse with all its sub-components.
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
// TestSuccessfulTransmission tests that transmission can succeed when vesicles are available
// Updated to account for biological vesicle dynamics and probabilistic release
func TestSuccessfulTransmission(t *testing.T) {
	// Create synapse with mock neuron
	mockNeuron := NewMockNeuron("test_neuron")
	config := CreateExcitatoryGlutamatergicConfig("test-synapse", "source-neuron", mockNeuron.ID())

	processor, err := CreateSynapse(config.SynapseID, config, SynapseCallbacks{
		DeliverMessage: mockNeuron.Receive,
	})
	if err != nil {
		t.Fatalf("Failed to create synapse: %v", err)
	}

	synapse := processor.(*Synapse)

	t.Log("Testing biological transmission with vesicle dynamics...")

	// Try multiple transmission attempts to account for probabilistic release
	// In biological synapses, not every attempt succeeds due to vesicle availability
	successCount := 0
	attemptCount := 10

	for i := 0; i < attemptCount; i++ {
		err := synapse.Transmit(1.0)
		if err == nil {
			successCount++
			t.Logf("✅ Transmission %d succeeded", i+1)
		} else if err == ErrVesicleDepleted {
			t.Logf("🧬 Transmission %d: vesicle depletion (normal biology)", i+1)
		} else {
			t.Errorf("❌ Unexpected error on transmission %d: %v", i+1, err)
		}
	}

	t.Logf("📊 Results: %d successes out of %d attempts (%.1f%%)",
		successCount, attemptCount, float64(successCount)/float64(attemptCount)*100)

	// Biological validation: Should get some successes (not zero, not all)
	if successCount == 0 {
		t.Error("No successful transmissions - vesicle system may be too restrictive")
	} else if successCount == attemptCount {
		t.Error("All transmissions succeeded - vesicle dynamics may not be working")
	} else {
		t.Logf("✅ Biological behavior confirmed: partial success rate as expected")
	}

	// Verify that successful transmissions actually delivered messages
	messages := mockNeuron.GetReceivedMessages()
	if len(messages) != successCount {
		t.Errorf("Message count mismatch: expected %d, got %d", successCount, len(messages))
	}

	// Verify vesicle depletion occurred
	vesicleState := synapse.GetVesicleState()
	if vesicleState.DepletionLevel == 0.0 {
		t.Error("Expected some vesicle depletion after multiple transmissions")
	} else {
		t.Logf("✅ Vesicle depletion confirmed: %.1f%% depleted", vesicleState.DepletionLevel*100)
	}

	t.Log("✅ Biological transmission test completed successfully")
}

// TestTransmissionFailureOnVesicleDepletion tests vesicle depletion behavior
// Updated to properly test biological vesicle pool exhaustion
func TestTransmissionFailureOnVesicleDepletion(t *testing.T) {
	mockNeuron := NewMockNeuron("test_neuron")
	config := CreateExcitatoryGlutamatergicConfig("test-synapse", "source-neuron", mockNeuron.ID())

	processor, err := CreateSynapse(config.SynapseID, config, SynapseCallbacks{
		DeliverMessage: mockNeuron.Receive,
	})
	if err != nil {
		t.Fatalf("Failed to create synapse: %v", err)
	}

	synapse := processor.(*Synapse)

	t.Log("Testing vesicle depletion behavior...")

	// Get initial vesicle state
	initialState := synapse.GetVesicleState()
	t.Logf("Initial ready pool: %d vesicles", initialState.ReadyVesicles)

	// Attempt many transmissions to exhaust vesicle pools
	successCount := 0
	vesicleFailureCount := 0
	maxAttempts := 100 // Try many more than the ready pool size

	for i := 0; i < maxAttempts; i++ {
		err := synapse.Transmit(1.0)
		if err == nil {
			successCount++
		} else if err == ErrVesicleDepleted {
			vesicleFailureCount++
			if i < 5 { // Log first few failures
				t.Logf("🧬 Vesicle depletion on attempt %d (normal biology)", i+1)
			}
		} else {
			t.Errorf("❌ Unexpected error on attempt %d: %v", i+1, err)
		}

		// Stop if we've had many consecutive failures (pool is exhausted)
		if vesicleFailureCount > 20 && successCount == 0 {
			t.Logf("Stopping after %d attempts - vesicle pool appears exhausted", i+1)
			break
		}
	}

	finalState := synapse.GetVesicleState()
	t.Logf("📊 Depletion Test Results:")
	t.Logf("  • Successful transmissions: %d", successCount)
	t.Logf("  • Vesicle depletion failures: %d", vesicleFailureCount)
	t.Logf("  • Final ready pool: %d vesicles (started with %d)",
		finalState.ReadyVesicles, initialState.ReadyVesicles)
	t.Logf("  • Final depletion level: %.1f%%", finalState.DepletionLevel*100)

	// Biological validation
	if successCount == 0 {
		t.Error("No successful transmissions - system may be too restrictive")
	}

	if vesicleFailureCount == 0 {
		t.Error("No vesicle depletion failures - vesicle dynamics not working")
	}

	if successCount > int(initialState.ReadyVesicles)*2 {
		t.Errorf("Too many successes (%d) for initial pool size (%d) - biological limit exceeded",
			successCount, initialState.ReadyVesicles)
	}

	// Should see significant depletion
	if finalState.DepletionLevel < 0.3 {
		t.Errorf("Expected significant depletion (>30%%), got %.1f%%", finalState.DepletionLevel*100)
	}

	// Verify that successful transmissions delivered messages
	messages := mockNeuron.GetReceivedMessages()
	if len(messages) != successCount {
		t.Errorf("Message delivery mismatch: expected %d messages, got %d", successCount, len(messages))
	}

	// Verify that high failure rate is due to vesicle depletion, not other errors
	totalAttempts := successCount + vesicleFailureCount
	vesicleFailureRate := float64(vesicleFailureCount) / float64(totalAttempts) * 100

	if vesicleFailureRate < 50.0 {
		t.Errorf("Expected high vesicle failure rate (>50%%), got %.1f%%", vesicleFailureRate)
	} else {
		t.Logf("✅ High vesicle failure rate confirmed: %.1f%% (biologically realistic)", vesicleFailureRate)
	}

	t.Log("✅ Vesicle depletion test completed successfully")
}

// TestTransmissionFailureOnVesicleDepletion verifies that transmission fails
// when the vesicle pool is depleted
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
	synapse := processor.(*Synapse)
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
