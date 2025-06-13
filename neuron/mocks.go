package neuron

import (
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/synapse"
)

// MockNeuron implements the synapse.SynapseCompatibleNeuron interface
// for testing purposes. This allows us to create controlled test networks
// where we can precisely verify signal transmission and timing.
//
// BIOLOGICAL CONTEXT:
// In real neural networks, neurons must communicate through synaptic connections.
// This mock neuron models a simplified post-synaptic neuron that can receive
// and record synaptic inputs for verification during testing.
type MockNeuron struct {
	id           string
	receivedMsgs []synapse.SynapseMessage
	mutex        sync.Mutex
}

// NewMockNeuron creates a mock neuron for testing synapse communication
func NewMockNeuron(id string) *MockNeuron {
	return &MockNeuron{
		id:           id,
		receivedMsgs: make([]synapse.SynapseMessage, 0),
	}
}

// ID returns the neuron's unique identifier
func (m *MockNeuron) ID() string {
	return m.id
}

// Receive implements the synapse.SynapseCompatibleNeuron interface
// Records all received messages for test verification
func (m *MockNeuron) Receive(msg synapse.SynapseMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.receivedMsgs = append(m.receivedMsgs, msg)
}

// GetReceivedMessages returns a copy of all received messages for testing
func (m *MockNeuron) GetReceivedMessages() []synapse.SynapseMessage {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Return a copy to prevent external modification
	messages := make([]synapse.SynapseMessage, len(m.receivedMsgs))
	copy(messages, m.receivedMsgs)
	return messages
}

// ClearReceivedMessages clears the message history for fresh testing
func (m *MockNeuron) ClearReceivedMessages() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.receivedMsgs = m.receivedMsgs[:0]
}

func (m *MockNeuron) ScheduleDelayedDelivery(message synapse.SynapseMessage, target synapse.SynapseCompatibleNeuron, delay time.Duration) {
	// Simple mock implementation - immediate delivery
	target.Receive(message)
}
