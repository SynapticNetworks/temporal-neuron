package synapse

import (
	"sort"
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/message"
)

// =================================================================================
// MOCK NEURON IMPLEMENTATION FOR TESTING
// =================================================================================

// MockNeuron implements SynapseNeuronInterface and component.MessageReceiver for testing purposes.
type MockNeuron struct {
	// Embed BaseComponent to satisfy component.Component and thus component.MessageReceiver requirements.
	*component.BaseComponent

	// === MESSAGE STORAGE ===
	receivedMsgs []message.NeuralSignal    // All messages received by this neuron
	msgChannel   chan message.NeuralSignal // Channel for concurrent test access

	// === DELAY SCHEDULING ===
	delayQueue  []delayedMessage // Internal queue for delayed messages
	currentTime time.Time        // Simulated current time for testing

	// === Mock-specific concurrency control ===
	mockMutex sync.RWMutex
}

// delayedMessage represents a message awaiting delivery in the mock system
type delayedMessage struct {
	message      message.NeuralSignal
	target       component.MessageReceiver // Target must be component.MessageReceiver
	deliveryTime time.Time
}

// NewMockNeuron creates a simple mock neuron for testing synapse functionality.
func NewMockNeuron(id string) *MockNeuron {
	base := component.NewBaseComponent(id, component.TypeNeuron, component.Position3D{})
	base.SetState(component.StateInactive)

	return &MockNeuron{
		BaseComponent: base,
		receivedMsgs:  make([]message.NeuralSignal, 0),
		msgChannel:    make(chan message.NeuralSignal, 10),
		delayQueue:    make([]delayedMessage, 0),
		currentTime:   time.Now(),
		mockMutex:     sync.RWMutex{},
	}
}

// =================================================================================
// COMPONENT.MESSAGERECEIVER AND SYNAPSENEURONINTERFACE IMPLEMENTATION
// =================================================================================

// Receive accepts a neural signal and stores it for test verification.
func (m *MockNeuron) Receive(msg message.NeuralSignal) {
	m.mockMutex.Lock()
	defer m.mockMutex.Unlock()

	m.receivedMsgs = append(m.receivedMsgs, msg)

	select {
	case m.msgChannel <- msg:
	default:
	}
}

// ScheduleDelayedDelivery implements the SynapseNeuronInterface.ScheduleDelayedDelivery() requirement.
func (m *MockNeuron) ScheduleDelayedDelivery(msg message.NeuralSignal, target component.MessageReceiver, delay time.Duration) {
	m.mockMutex.Lock()
	defer m.mockMutex.Unlock()

	if delay <= 0 {
		target.Receive(msg)
		return
	}

	delayedMsg := delayedMessage{
		message:      msg,
		target:       target,
		deliveryTime: m.currentTime.Add(delay),
	}
	m.delayQueue = append(m.delayQueue, delayedMsg)
}

// =================================================================================
// TEST UTILITY METHODS FOR MOCKNEURON
// =================================================================================

// GetReceivedMessages returns all messages received by this mock neuron.
func (m *MockNeuron) GetReceivedMessages() []message.NeuralSignal {
	m.mockMutex.RLock()
	defer m.mockMutex.RUnlock()

	copied := make([]message.NeuralSignal, len(m.receivedMsgs))
	copy(copied, m.receivedMsgs)
	return copied
}

// ClearReceivedMessages clears the message history for clean test states.
func (m *MockNeuron) ClearReceivedMessages() {
	m.mockMutex.Lock()
	defer m.mockMutex.Unlock()

	m.receivedMsgs = m.receivedMsgs[:0]

	for len(m.msgChannel) > 0 {
		<-m.msgChannel
	}
}

// ProcessDelayedMessages simulates the axonal delivery system by processing
// all queued delayed messages that are due for delivery based on the current
// simulated time.
func (m *MockNeuron) ProcessDelayedMessages(currentTime time.Time) int {
	m.mockMutex.Lock()
	defer m.mockMutex.Unlock()

	m.currentTime = currentTime
	deliveredCount := 0

	sort.Slice(m.delayQueue, func(i, j int) bool {
		return m.delayQueue[i].deliveryTime.Before(m.delayQueue[j].deliveryTime)
	})

	remainingMessages := make([]delayedMessage, 0, len(m.delayQueue))
	for _, delayedMsg := range m.delayQueue {
		if currentTime.After(delayedMsg.deliveryTime) || currentTime.Equal(delayedMsg.deliveryTime) {
			delayedMsg.target.Receive(delayedMsg.message)
			deliveredCount++
		} else {
			remainingMessages = append(remainingMessages, delayedMsg)
		}
	}

	m.delayQueue = remainingMessages
	return deliveredCount
}

// GetQueuedMessageCount returns the number of messages currently waiting
// for delayed delivery.
func (m *MockNeuron) GetQueuedMessageCount() int {
	m.mockMutex.RLock()
	defer m.mockMutex.RUnlock()
	return len(m.delayQueue)
}

// SetCurrentTime updates the mock's internal time for testing time-dependent behavior.
func (m *MockNeuron) SetCurrentTime(t time.Time) {
	m.mockMutex.Lock()
	defer m.mockMutex.Unlock()
	m.currentTime = t
}

// =================================================================================
// MOCK SYNAPSE IMPLEMENTATION FOR TESTING
// =================================================================================

type MockSynapse struct {
	id string

	preSynapticID  string
	postSynapticID string

	weight     float64
	delay      time.Duration
	ligandType message.LigandType

	plasticityEnabled bool
	stdpConfig        STDPConfig
	pruningConfig     PruningConfig

	activity          float64
	lastActivity      time.Time
	transmissionCount int64

	// callbacks *extracellular.SynapseCallbacks  <-- REMOVED THIS PROBLEMATIC FIELD
	// MockSynapse does not need to store these, as it mimics BasicSynapse
	// which also does not store them as fields.

	isActive bool

	mutex sync.RWMutex

	receivedSignals []message.NeuralSignal
}

// NewMockSynapse creates a mock synapse.
func NewMockSynapse(id string, preID, postID string, weight float64) *MockSynapse {
	return &MockSynapse{
		id:                id,
		preSynapticID:     preID,
		postSynapticID:    postID,
		weight:            weight,
		delay:             time.Millisecond,
		ligandType:        message.LigandGlutamate,
		plasticityEnabled: true,
		stdpConfig:        CreateDefaultSTDPConfig(),
		pruningConfig:     CreateDefaultPruningConfig(),
		activity:          0.0,
		lastActivity:      time.Now(),
		transmissionCount: 0,
		// callbacks:         nil, // This line is also removed as the field is gone
		isActive:        true,
		receivedSignals: make([]message.NeuralSignal, 0),
		mutex:           sync.RWMutex{},
	}
}

// GetReceivedMessages for MockSynapse, for testing purposes.
func (m *MockSynapse) GetReceivedMessages() []message.NeuralSignal {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	copied := make([]message.NeuralSignal, len(m.receivedSignals))
	copy(copied, m.receivedSignals)
	return copied
}

// ClearReceivedMessages for MockSynapse.
func (m *MockSynapse) ClearReceivedMessages() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.receivedSignals = m.receivedSignals[:0]
}

// Transmit simulates signal transmission for testing the synapse itself.
func (ms *MockSynapse) Transmit(signalValue float64) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	msg := message.NeuralSignal{
		Value:                signalValue * ms.weight,
		Timestamp:            time.Now(),
		SourceID:             ms.preSynapticID,
		SynapseID:            ms.id,
		TargetID:             ms.postSynapticID,
		NeurotransmitterType: ms.ligandType,
	}
	ms.receivedSignals = append(ms.receivedSignals, msg)

	ms.transmissionCount++
	ms.lastActivity = time.Now()
}

// ApplyPlasticity updates the synapse's internal state based on feedback.
func (ms *MockSynapse) ApplyPlasticity(adjustment PlasticityAdjustment) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.lastActivity = time.Now()
}

// ShouldPrune evaluates if the synapse should be removed.
func (ms *MockSynapse) ShouldPrune() bool {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	return false
}

// GetWeight returns the current synaptic weight.
func (ms *MockSynapse) GetWeight() float64 {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()
	return ms.weight
}

// SetWeight allows direct manipulation of synaptic strength.
func (ms *MockSynapse) SetWeight(weight float64) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.weight = weight
}

// GetPresynapticID returns the ID of the presynaptic neuron.
func (ms *MockSynapse) GetPresynapticID() string {
	return ms.preSynapticID
}

// GetPostsynapticID returns the ID of the postsynaptic neuron.
func (ms *MockSynapse) GetPostsynapticID() string {
	return ms.postSynapticID
}

// GetDelay returns the transmission delay.
func (ms *MockSynapse) GetDelay() time.Duration {
	return ms.delay
}

// GetPlasticityConfig returns the STDP configuration.
func (ms *MockSynapse) GetPlasticityConfig() STDPConfig {
	return ms.stdpConfig
}

// GetLastActivity returns the time of the last activity.
func (ms *MockSynapse) GetLastActivity() time.Time {
	return ms.lastActivity
}

// ID (for component.MessageReceiver compatibility and SynapticProcessor.ID())
// This is included to ensure MockSynapse can potentially act as a MessageReceiver
// if needed in some test scenarios, even though its primary role is transmitting.
func (ms *MockSynapse) ID() string {
	return ms.id
}
