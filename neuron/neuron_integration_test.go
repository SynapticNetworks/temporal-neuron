package neuron

import (
	"fmt"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

// ============================================================================
// BASIC INTEGRATION TESTS - TESTING REAL FUNCTIONALITY
// ============================================================================

func TestNeuronIntegration_BasicMessageProcessing(t *testing.T) {
	// Test: Can a neuron receive and process messages?

	neuron := NewNeuron(
		"test-neuron",
		0.5, // Low threshold for easy firing
		0.1,
		1*time.Millisecond,
		1.0,
		10.0,
		0.5,
	)

	// Start the neuron's background processing
	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Send a message
	msg := types.NeuralSignal{
		Value:     1.0, // Above threshold
		Timestamp: time.Now(),
		SourceID:  "test-source",
		TargetID:  neuron.ID(),
	}

	neuron.Receive(msg)

	// Allow processing time
	time.Sleep(10 * time.Millisecond)

	// Verify the neuron processed the message
	activity := neuron.GetActivityLevel()
	if activity <= 0 {
		t.Errorf("Expected neuron to show activity after receiving message, got: %f", activity)
	}

	t.Logf("✓ Neuron processed message and shows activity: %f", activity)
}

func TestNeuronIntegration_OutputCallbacks(t *testing.T) {
	// Test: Can a neuron send signals via output callbacks?

	neuron := NewNeuron(
		"sender-neuron",
		0.5, // Low threshold
		0.1,
		1*time.Millisecond,
		1.0,
		10.0,
		0.5,
	)

	// Create a mock synapse to receive signals
	mockSynapse := &MockSynapse{id: "test-synapse"}

	// Set up output callback
	callback := types.OutputCallback{
		TransmitMessage: func(msg types.NeuralSignal) error {
			mockSynapse.ReceiveSignal(msg)
			return nil
		},
		GetWeight:   func() float64 { return 1.0 },
		GetDelay:    func() time.Duration { return 1 * time.Millisecond },
		GetTargetID: func() string { return "target-neuron" },
	}

	neuron.AddOutputCallback("test-synapse", callback)

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Send strong signal to trigger firing
	msg := types.NeuralSignal{
		Value:     2.0, // Well above threshold
		Timestamp: time.Now(),
		SourceID:  "stimulus",
		TargetID:  neuron.ID(),
	}

	neuron.Receive(msg)

	// Allow time for processing and firing
	time.Sleep(20 * time.Millisecond)

	// Check if mock synapse received signals
	receivedCount := mockSynapse.GetReceivedCount()
	if receivedCount == 0 {
		t.Error("Expected mock synapse to receive signals from neuron firing")
	} else {
		t.Logf("✓ Neuron fired and transmitted %d signal(s) via output callback", receivedCount)
	}
}

func TestNeuronIntegration_MatrixCallbacks(t *testing.T) {
	// Test: Can a neuron use matrix callbacks for coordination?

	mockMatrix := &MockMatrix{}

	neuron := NewNeuron(
		"matrix-neuron",
		0.5,
		0.1,
		1*time.Millisecond,
		1.0,
		10.0,
		0.5,
	)

	// Set up matrix callbacks
	callbacks := mockMatrix.CreateBasicCallbacks()
	neuron.SetCallbacks(callbacks)

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Test synapse creation
	err = neuron.ConnectToNeuron("target", 1.0, "excitatory")
	if err != nil {
		t.Errorf("Failed to create synapse via matrix: %v", err)
	}

	// Test firing (should trigger health reporting and chemical release)
	msg := types.NeuralSignal{
		Value:     2.0,
		Timestamp: time.Now(),
		SourceID:  "stimulus",
		TargetID:  neuron.ID(),
	}

	neuron.Receive(msg)
	time.Sleep(20 * time.Millisecond)

	// Verify matrix interactions
	if mockMatrix.GetSynapseCreationCount() == 0 {
		t.Error("Expected synapse creation via matrix callback")
	} else {
		t.Logf("✓ Synapse created via matrix: %d creation(s)", mockMatrix.GetSynapseCreationCount())
	}

	if mockMatrix.GetHealthReportCount() == 0 {
		t.Error("Expected health reports to matrix")
	} else {
		t.Logf("✓ Health reported to matrix: %d report(s)", mockMatrix.GetHealthReportCount())
	}

	if mockMatrix.GetElectricalSignalCount() == 0 {
		t.Error("Expected electrical signals to matrix")
	} else {
		t.Logf("✓ Electrical signals sent to matrix: %d signal(s)", mockMatrix.GetElectricalSignalCount())
	}
}

func TestNeuronIntegration_ChemicalSignaling(t *testing.T) {
	// Test: Can a neuron handle chemical signaling?

	mockMatrix := &MockMatrix{}

	neuron := NewNeuron(
		"chemical-neuron",
		0.5,
		0.1,
		1*time.Millisecond,
		1.0,
		10.0,
		0.5,
	)

	// Set up chemical properties
	neuron.SetReceptors([]types.LigandType{types.LigandGlutamate})
	neuron.SetReleasedLigands([]types.LigandType{types.LigandDopamine})

	callbacks := mockMatrix.CreateBasicCallbacks()
	neuron.SetCallbacks(callbacks)

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Test chemical reception
	neuron.Bind(types.LigandGlutamate, "external-source", 0.3)

	// Test chemical release (via firing)
	msg := types.NeuralSignal{
		Value:     2.0,
		Timestamp: time.Now(),
		SourceID:  "stimulus",
		TargetID:  neuron.ID(),
	}

	neuron.Receive(msg)
	time.Sleep(20 * time.Millisecond)

	// Verify chemical releases
	if mockMatrix.GetChemicalReleaseCount() == 0 {
		t.Error("Expected chemical releases after neuron firing")
	} else {
		t.Logf("✓ Chemical signaling works: %d release(s)", mockMatrix.GetChemicalReleaseCount())
	}
}

func TestNeuronIntegration_ComponentInterfaces(t *testing.T) {
	// Test: Does neuron properly implement component interfaces?

	neuron := NewNeuron(
		"interface-neuron",
		1.0,
		0.1,
		5*time.Millisecond,
		1.0,
		10.0,
		0.5,
	)

	// Test Component interface
	if neuron.ID() != "interface-neuron" {
		t.Errorf("Expected ID 'interface-neuron', got %s", neuron.ID())
	}

	if neuron.Type() != types.TypeNeuron {
		t.Errorf("Expected type TypeNeuron, got %v", neuron.Type())
	}

	if !neuron.IsActive() {
		t.Error("New neuron should be active")
	}

	// Test lifecycle
	err := neuron.Start()
	if err != nil {
		t.Errorf("Failed to start neuron: %v", err)
	}

	err = neuron.Stop()
	if err != nil {
		t.Errorf("Failed to stop neuron: %v", err)
	}

	// Test MessageReceiver interface
	msg := types.NeuralSignal{
		Value:     0.5,
		Timestamp: time.Now(),
		SourceID:  "test",
		TargetID:  neuron.ID(),
	}

	// Should not panic
	neuron.Receive(msg)

	t.Log("✓ Neuron properly implements component interfaces")
}

// ============================================================================
// CLEANED UP ERROR HANDLING TEST - Uses Real Components Instead of Incomplete Callbacks
// ============================================================================

func TestNeuronIntegration_ErrorHandling(t *testing.T) {
	t.Log("=== Testing Neuron Error Handling with Real Components ===")

	// Create neuron with proper mock matrix
	mockMatrix := NewMockMatrix()
	neuron := NewNeuron(
		"error-neuron",
		1.0,
		0.1,
		5*time.Millisecond,
		1.0,
		10.0,
		0.5,
	)

	// Set proper callbacks
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// ============================================================================
	// TEST 1: Error injection for synapse creation
	// ============================================================================
	t.Log("Testing synapse creation error handling...")

	mockMatrix.SetCreateSynapseError(fmt.Errorf("matrix overload"))
	err = neuron.ConnectToNeuron("target", 1.0, "excitatory")
	if err == nil {
		t.Error("Expected error when matrix rejects synapse creation")
	} else {
		t.Logf("✓ Correctly handled synapse creation error: %v", err)
	}

	// Clear error for subsequent tests
	mockMatrix.SetCreateSynapseError(nil)

	// ============================================================================
	// TEST 2: Error injection for plasticity operations
	// ============================================================================
	t.Log("Testing plasticity error handling...")

	// Add a synapse for plasticity testing
	synapseInfo := types.SynapseInfo{
		ID:       "test-synapse",
		SourceID: "source-neuron",
		TargetID: neuron.ID(),
		Weight:   1.5,
	}
	mockMatrix.AddSynapse(synapseInfo)

	// Inject plasticity error
	mockMatrix.SetApplyPlasticityError(fmt.Errorf("plasticity mechanism failure"))

	// These should handle errors gracefully (not panic)
	neuron.SendSTDPFeedback()
	t.Log("✓ SendSTDPFeedback handled plasticity error gracefully")

	// Clear error
	mockMatrix.SetApplyPlasticityError(nil)

	// ============================================================================
	// TEST 3: Error injection for weight operations
	// ============================================================================
	t.Log("Testing weight operation error handling...")

	mockMatrix.SetSetSynapseWeightError(fmt.Errorf("weight adjustment failed"))
	neuron.PerformHomeostasisScaling()
	t.Log("✓ PerformHomeostasisScaling handled weight error gracefully")

	// Clear error
	mockMatrix.SetSetSynapseWeightError(nil)

	// ============================================================================
	// TEST 4: Error injection for synapse access
	// ============================================================================
	t.Log("Testing synapse access error handling...")

	mockMatrix.SetGetSynapseError(fmt.Errorf("synapse not accessible"))
	neuron.PruneDysfunctionalSynapses()
	t.Log("✓ PruneDysfunctionalSynapses handled access error gracefully")

	// Clear error
	mockMatrix.SetGetSynapseError(nil)

	// ============================================================================
	// TEST 5: Basic functionality should still work regardless of errors
	// ============================================================================
	t.Log("Testing basic functionality resilience...")

	msg := types.NeuralSignal{
		Value:     1.0,
		Timestamp: time.Now(),
		SourceID:  "test",
		TargetID:  neuron.ID(),
	}

	// This should always work regardless of callback errors
	neuron.Receive(msg)
	t.Log("✓ Basic message reception works despite callback errors")

	// ============================================================================
	// TEST 6: Recovery after error clearing
	// ============================================================================
	t.Log("Testing recovery after errors are cleared...")

	// Now operations should work normally
	err = neuron.ConnectToNeuron("target", 1.0, "excitatory")
	if err != nil {
		t.Errorf("Expected successful connection after error cleared, got: %v", err)
	} else {
		t.Log("✓ Operations work normally after errors are cleared")
	}

	// Verify the connection was actually created
	creations := mockMatrix.GetSynapseCreations()
	if len(creations) == 0 {
		t.Error("Expected synapse creation to be recorded")
	} else {
		t.Logf("✓ Synapse creation recorded: %s -> %s",
			creations[len(creations)-1].Config.SourceNeuronID,
			creations[len(creations)-1].Config.TargetNeuronID)
	}

	t.Log("✓ Neuron handles errors gracefully and recovers properly")
}

// ============================================================================
// ADDITIONAL: Test with nil callbacks (if you want to test that scenario)
// ============================================================================

func TestNeuronIntegration_NilCallbacks(t *testing.T) {
	t.Log("=== Testing Neuron with Nil Callbacks ===")

	neuron := NewNeuron(
		"nil-callback-neuron",
		1.0,
		0.1,
		5*time.Millisecond,
		1.0,
		10.0,
		0.5,
	)

	// Don't set any callbacks (leave as nil)

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Operations requiring callbacks should handle nil gracefully
	err = neuron.ConnectToNeuron("target", 1.0, "excitatory")
	if err == nil {
		t.Error("Expected error when callbacks are nil")
	} else {
		t.Logf("✓ Correctly handled nil callbacks: %v", err)
	}

	// These should not panic even with nil callbacks
	neuron.SendSTDPFeedback()
	neuron.PerformHomeostasisScaling()
	neuron.PruneDysfunctionalSynapses()

	// Basic functionality should still work
	msg := types.NeuralSignal{
		Value:     1.0,
		Timestamp: time.Now(),
		SourceID:  "test",
		TargetID:  neuron.ID(),
	}

	neuron.Receive(msg)

	t.Log("✓ Neuron handles nil callbacks gracefully")
}

// ============================================================================
// STRESS TEST: Multiple error conditions simultaneously
// ============================================================================

func TestNeuronIntegration_MultipleErrors(t *testing.T) {
	t.Log("=== Testing Multiple Simultaneous Errors ===")

	mockMatrix := NewMockMatrix()
	neuron := NewNeuron(
		"stress-neuron",
		1.0,
		0.1,
		5*time.Millisecond,
		1.0,
		10.0,
		0.5,
	)

	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())
	neuron.Start()
	defer neuron.Stop()

	// Set multiple errors simultaneously
	mockMatrix.SetCreateSynapseError(fmt.Errorf("creation failed"))
	mockMatrix.SetApplyPlasticityError(fmt.Errorf("plasticity failed"))
	mockMatrix.SetSetSynapseWeightError(fmt.Errorf("weight setting failed"))
	mockMatrix.SetGetSynapseError(fmt.Errorf("synapse access failed"))

	// Add some synapses for operations to work on
	for i := 0; i < 3; i++ {
		synapseInfo := types.SynapseInfo{
			ID:       fmt.Sprintf("synapse-%d", i),
			SourceID: "source",
			TargetID: neuron.ID(),
			Weight:   1.0,
		}
		mockMatrix.AddSynapse(synapseInfo)
	}

	// All these operations should handle errors gracefully
	neuron.ConnectToNeuron("target", 1.0, "excitatory")
	neuron.SendSTDPFeedback()
	neuron.PerformHomeostasisScaling()
	neuron.PruneDysfunctionalSynapses()

	// Basic operations should still work
	msg := types.NeuralSignal{
		Value:     1.0,
		Timestamp: time.Now(),
		SourceID:  "test",
		TargetID:  neuron.ID(),
	}
	neuron.Receive(msg)

	t.Log("✓ Neuron handles multiple simultaneous errors gracefully")
}

// ============================================================================
// WHAT TO DELETE FROM YOUR OLD TEST:
// ============================================================================

/*
DELETE THIS ENTIRE SECTION - it's using the old struct-based approach:

incompleteCallbacks := NeuronCallbacks{
    // Only some callbacks provided
    ReportHealth: func(activityLevel float64, connectionCount int) {},
    ListSynapses: func(criteria types.SynapseCriteria) []types.SynapseInfo {
        return []types.SynapseInfo{}
    },
}

This was problematic because:
1. NeuronCallbacks is now an interface, not a struct
2. You can't create incomplete interface implementations like this
3. The mock approach is much more realistic and testable
*/

// ============================================================================
// TEST SUMMARY
// ============================================================================

/*
This integration test suite verifies:

1. ✓ Basic message processing (Receive → Run → process)
2. ✓ Output callback interface (neuron → synapse communication)
3. ✓ Matrix callback interface (neuron → matrix coordination)
4. ✓ Chemical signaling (receptors, releases)
5. ✓ Component interface compliance
6. ✓ Error handling with missing callbacks

These tests confirm that the neuron's callback architecture properly
integrates with external synapse and matrix systems without requiring
actual implementations of those systems.
*/
