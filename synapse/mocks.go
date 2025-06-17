/*
=================================================================================
MOCK NEURON FOR TESTING - REFACTORED
=================================================================================

OVERVIEW:
This file provides a mock implementation of a neuron for testing purposes. It has
been updated to align with the new, decoupled architecture where synapses
interact with their environment via callbacks.

DESIGN PRINCIPLES:
1.  INTERFACE-COMPLIANT: Implements the function signatures required by the
    `SynapseCallbacks` struct, allowing it to be used as a test target.
2.  OBSERVABLE: Stores all received messages for test verification in a thread-safe manner.
3.  DECOUPLED: Contains no references to concrete synapse or matrix implementations.
4.  SIMPLE: Provides the minimal functionality required for testing without the
    complexity of a full neuron simulation.
*/

package synapse

import (
	"sync"
)

// =================================================================================
// MOCK NEURON IMPLEMENTATION FOR TESTING
// =================================================================================

// MockNeuron implements the necessary functions to act as a test endpoint for a synapse.
// It is designed to be thread-safe to support concurrent testing scenarios.
type MockNeuron struct {
	// === IDENTIFICATION ===
	id string // Unique identifier for this mock neuron

	// === MESSAGE STORAGE ===
	receivedMsgs []SynapseMessage // Thread-safe log of all received messages

	// === THREAD-SAFETY ===
	mu sync.Mutex // Protects access to receivedMsgs
}

// NewMockNeuron creates a simple mock neuron for testing synapse functionality.
func NewMockNeuron(id string) *MockNeuron {
	return &MockNeuron{
		id:           id,
		receivedMsgs: make([]SynapseMessage, 0),
	}
}

// ID returns the unique identifier of this mock neuron.
func (m *MockNeuron) ID() string {
	return m.id
}

// Receive is the core method used as a callback for the synapse's DeliverMessage function.
// It accepts a synapse message and stores it for later verification by a test.
// Its signature now matches the `DeliverMessage` callback type precisely.
//
// Parameters:
//   - targetID: The ID of the neuron intended to receive the message (ignored by the mock).
//   - msg: The SynapseMessage delivered by a synapse.
//
// Returns:
//   - nil, to signify successful reception.
func (m *MockNeuron) Receive(targetID string, msg SynapseMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Store the message in our log for test verification.
	m.receivedMsgs = append(m.receivedMsgs, msg)

	return nil
}

// GetReceivedMessages returns a copy of all messages received by this mock neuron.
// It is thread-safe and returns a copy to prevent race conditions during tests
// where the slice could be read by one goroutine while being written to by another.
//
// Returns:
//
//	A slice containing a copy of all SynapseMessage objects received.
func (m *MockNeuron) GetReceivedMessages() []SynapseMessage {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Return a copy to ensure the test has a stable snapshot of the messages
	// and to prevent race conditions.
	msgsCopy := make([]SynapseMessage, len(m.receivedMsgs))
	copy(msgsCopy, m.receivedMsgs)

	return msgsCopy
}

// ClearMessages resets the log of received messages.
// This is useful for resetting state between test cases.
func (m *MockNeuron) ClearMessages() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.receivedMsgs = make([]SynapseMessage, 0)
}

// NOTE: The obsolete `ScheduleDelayedDelivery` method, which referred to the
// undefined `SynapseCompatibleNeuron` type, has been removed as it is no longer
// part of the new architecture. The synapse now handles delays internally before
// invoking the `DeliverMessage` callback.
