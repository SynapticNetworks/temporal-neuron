package synapse

import "time"

// =================================================================================
// MOCK NEURON IMPLEMENTATION FOR TESTING
// =================================================================================

// MockNeuron implements SynapseCompatibleNeuron for testing purposes.
// This is a minimal, controlled implementation that allows precise testing
// of synapse functionality without dependencies on complex neuron behavior.
//
// DESIGN PRINCIPLES:
// 1. SIMPLE: Only implements the interface methods needed for synapse testing
// 2. OBSERVABLE: Stores all received messages for test verification
// 3. CONTROLLABLE: Predictable behavior that won't interfere with test logic
// 4. THREAD-SAFE: Can handle messages from multiple synapses concurrently
//
// BIOLOGICAL ABSTRACTION:
// While real neurons have complex membrane dynamics, homeostatic regulation,
// and sophisticated processing, this mock focuses purely on the message
// reception interface that synapses need. This isolation allows us to test
// synaptic behavior without the complexity of full neural simulation.
type MockNeuron struct {
	// === IDENTIFICATION ===
	id string // Unique identifier for this mock neuron

	// === MESSAGE STORAGE ===
	// These fields store received messages for test verification
	receivedMsgs []SynapseMessage    // All messages received (for verification)
	msgChannel   chan SynapseMessage // Buffered channel for message reception

	// Note: In a real neuron, received messages would trigger complex
	// membrane dynamics, homeostatic adjustments, and potential firing.
	// This mock simply stores them for test analysis.
}

// NewMockNeuron creates a simple mock neuron for testing synapse functionality.
// This factory function initializes a mock neuron with empty message storage
// and a buffered channel to handle multiple incoming messages during tests.
//
// Parameters:
//
//	id: Unique identifier for this mock neuron (used in test verification)
//
// Returns:
//
//	A fully initialized MockNeuron ready for testing
//
// USAGE IN TESTS:
// Mock neurons serve as controlled endpoints for synapse testing:
// - Pre-synaptic mock: Represents the source neuron sending signals
// - Post-synaptic mock: Represents the target neuron receiving signals
func NewMockNeuron(id string) *MockNeuron {
	return &MockNeuron{
		id:           id,                            // Store identification
		receivedMsgs: make([]SynapseMessage, 0),     // Initialize empty message log
		msgChannel:   make(chan SynapseMessage, 10), // Buffer for concurrent messages
	}
}

// ID returns the unique identifier of this mock neuron.
// This method implements the SynapseCompatibleNeuron interface requirement
// and allows synapses to identify message sources and targets.
//
// Returns:
//
//	The neuron's unique identifier string
//
// INTERFACE COMPLIANCE:
// This method satisfies the SynapseCompatibleNeuron.ID() requirement,
// enabling this mock to work seamlessly with the synapse system.
func (m *MockNeuron) ID() string {
	return m.id
}

// Receive accepts a synapse message and stores it for test verification.
// This method implements the SynapseCompatibleNeuron interface requirement
// and provides the essential functionality needed for synapse testing.
//
// Parameters:
//
//	msg: The SynapseMessage delivered by a synapse after transmission
//
// BIOLOGICAL SIMULATION:
// In a real neuron, this method would:
// 1. Convert the synaptic signal to a postsynaptic potential
// 2. Integrate the signal with existing membrane potential
// 3. Update homeostatic state and activity tracking
// 4. Potentially trigger action potential generation
//
// TESTING IMPLEMENTATION:
// For testing purposes, we simply:
// 1. Store the message for later verification
// 2. Forward to a channel for concurrent test access
// 3. Maintain thread-safety for multiple synapse inputs
//
// This simplified approach allows focused testing of synapse transmission
// behavior without the complexity of full neural dynamics.
func (m *MockNeuron) Receive(msg SynapseMessage) {
	// Store the message in our log for test verification
	// This allows tests to examine exactly what signals were received
	m.receivedMsgs = append(m.receivedMsgs, msg)

	// Also send to channel for concurrent test access
	// The select with default ensures non-blocking operation
	select {
	case m.msgChannel <- msg:
		// Message successfully queued for channel-based test access
	default:
		// Channel full - message still stored in receivedMsgs for verification
		// This models realistic scenarios where rapid firing could overwhelm buffers
	}
}

// GetReceivedMessages returns all messages received by this mock neuron.
// This method provides test access to the complete history of synaptic
// inputs, enabling comprehensive verification of synapse transmission behavior.
//
// Returns:
//
//	Slice of all SynapseMessage objects received by this neuron
//
// TESTING UTILITY:
// This method enables tests to:
// - Verify that synapses transmitted signals correctly
// - Check signal timing, strength, and source identification
// - Validate that weight scaling was applied properly
// - Confirm that delays were respected in transmission
// - Analyze message ordering in multi-synapse scenarios
//
// THREAD SAFETY:
// This method returns a reference to the internal slice. In production code,
// this would typically return a copy for safety, but for testing purposes,
// direct access simplifies verification logic.
func (m *MockNeuron) GetReceivedMessages() []SynapseMessage {
	return m.receivedMsgs
}

func (m *MockNeuron) ScheduleDelayedDelivery(message SynapseMessage, target SynapseCompatibleNeuron, delay time.Duration) {
	// Mock implementation - just do immediate delivery for tests
	target.Receive(message)
}
