package neuron

import (
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/message"
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
	msg := message.NeuralSignal{
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
	callback := OutputCallback{
		TransmitMessage: func(msg message.NeuralSignal) error {
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
	msg := message.NeuralSignal{
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
	msg := message.NeuralSignal{
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
	neuron.SetReceptors([]message.LigandType{message.LigandGlutamate})
	neuron.SetReleasedLigands([]message.LigandType{message.LigandDopamine})

	callbacks := mockMatrix.CreateBasicCallbacks()
	neuron.SetCallbacks(callbacks)

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Test chemical reception
	neuron.Bind(message.LigandGlutamate, "external-source", 0.3)

	// Test chemical release (via firing)
	msg := message.NeuralSignal{
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

	if neuron.Type() != component.TypeNeuron {
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
	msg := message.NeuralSignal{
		Value:     0.5,
		Timestamp: time.Now(),
		SourceID:  "test",
		TargetID:  neuron.ID(),
	}

	// Should not panic
	neuron.Receive(msg)

	t.Log("✓ Neuron properly implements component interfaces")
}

func TestNeuronIntegration_ErrorHandling(t *testing.T) {
	// Test: Does neuron handle missing callbacks gracefully?

	neuron := NewNeuron(
		"error-neuron",
		1.0,
		0.1,
		5*time.Millisecond,
		1.0,
		10.0,
		0.5,
	)

	// Set incomplete callbacks
	incompleteCallbacks := NeuronCallbacks{
		// Only some callbacks provided
		ReportHealth: func(activityLevel float64, connectionCount int) {},
		ListSynapses: func(criteria SynapseCriteria) []SynapseInfo {
			return []SynapseInfo{}
		},
	}

	neuron.SetCallbacks(incompleteCallbacks)

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Test operations that require missing callbacks
	err = neuron.ConnectToNeuron("target", 1.0, "excitatory")
	if err == nil {
		t.Error("Expected error when CreateSynapse callback missing")
	}

	// These should not panic
	neuron.SendSTDPFeedback()
	neuron.PerformHomeostasisScaling()
	neuron.PruneDysfunctionalSynapses()

	// Basic functionality should still work
	msg := message.NeuralSignal{
		Value:     1.0,
		Timestamp: time.Now(),
		SourceID:  "test",
		TargetID:  neuron.ID(),
	}

	neuron.Receive(msg)

	t.Log("✓ Neuron handles missing callbacks gracefully")
}

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
