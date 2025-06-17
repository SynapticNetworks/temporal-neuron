package extracellular

import (
	"sync"
	"sync/atomic"
	"time"
)

// =================================================================================
// MOCK COMPONENTS FOR TESTING
// =================================================================================

// BindingEvent records a chemical binding event for testing analysis
type BindingEvent struct {
	LigandType    LigandType
	SourceID      string
	Concentration float64
	Timestamp     time.Time
}

// MockNeuron represents a simple neuron for testing
type MockNeuron struct {
	id               string
	position         Position3D
	receptors        []LigandType
	firingThreshold  float64
	currentPotential float64
	connections      []string
	isActive         bool

	// Binding event tracking for testing
	bindingEventCount int
	bindingHistory    []BindingEvent
	mu                sync.RWMutex
}

func NewMockNeuron(id string, pos Position3D, receptors []LigandType) *MockNeuron {
	return &MockNeuron{
		id:                id,
		position:          pos,
		receptors:         receptors,
		firingThreshold:   0.7,
		currentPotential:  0.0,
		connections:       make([]string, 0),
		isActive:          true,
		bindingEventCount: 0,
		bindingHistory:    make([]BindingEvent, 0),
	}
}

// Implement Component interface
func (mn *MockNeuron) ID() string                   { return mn.id }
func (mn *MockNeuron) Position() Position3D         { return mn.position }
func (mn *MockNeuron) ComponentType() ComponentType { return ComponentNeuron }

// Implement BindingTarget interface (for chemical signaling)
func (mn *MockNeuron) Bind(ligandType LigandType, sourceID string, concentration float64) {
	mn.mu.Lock()
	defer mn.mu.Unlock()

	// Record binding event for testing
	mn.bindingEventCount++
	mn.bindingHistory = append(mn.bindingHistory, BindingEvent{
		LigandType:    ligandType,
		SourceID:      sourceID,
		Concentration: concentration,
		Timestamp:     time.Now(),
	})

	// Simple binding model: accumulate potential
	switch ligandType {
	case LigandGlutamate:
		mn.currentPotential += concentration * 0.8 // Increased from 0.5
	case LigandGABA:
		mn.currentPotential -= concentration * 0.5 // Increased from 0.3
	case LigandDopamine:
		mn.currentPotential += concentration * 0.4 // Increased from 0.2
	case LigandSerotonin:
		mn.currentPotential += concentration * 0.3
	case LigandAcetylcholine:
		mn.currentPotential += concentration * 0.6
	}

	// Lower threshold for better response
	firingThreshold := 0.3
	if mn.currentPotential > firingThreshold && mn.isActive {
		// Neuron fires - but don't reset to 0, just reduce
		mn.currentPotential *= 0.7 // Partial reset keeps some activation
	}
}

// Override OnSignal for better electrical responsiveness
func (mn *MockNeuron) OnSignal(signalType SignalType, sourceID string, data interface{}) {
	mn.mu.Lock()
	defer mn.mu.Unlock()

	switch signalType {
	case SignalFired:
		// Stronger response to action potentials
		if value, ok := data.(float64); ok && mn.isActive {
			mn.currentPotential += value * 0.2 // Increased from 0.1
		}
	case SignalConnected:
		if connID, ok := data.(string); ok {
			mn.connections = append(mn.connections, connID)
		}
	}
}

func (mn *MockNeuron) GetReceptors() []LigandType { return mn.receptors }
func (mn *MockNeuron) GetPosition() Position3D    { return mn.position }

// Helper methods for testing
func (mn *MockNeuron) GetCurrentPotential() float64 {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	return mn.currentPotential
}

func (mn *MockNeuron) SetPotential(potential float64) {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	mn.currentPotential = potential
}

func (mn *MockNeuron) GetConnections() []string {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	// Return copy to avoid race conditions
	connections := make([]string, len(mn.connections))
	copy(connections, mn.connections)
	return connections
}

func (mn *MockNeuron) IsActive() bool {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	return mn.isActive
}

func (mn *MockNeuron) SetActive(active bool) {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	mn.isActive = active
}

// GetBindingEventCount returns the total number of binding events (REQUIRED FOR TESTS)
func (mn *MockNeuron) GetBindingEventCount() int {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	return mn.bindingEventCount
}

// GetBindingHistory returns a copy of all binding events
func (mn *MockNeuron) GetBindingHistory() []BindingEvent {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	// Return copy to avoid race conditions
	history := make([]BindingEvent, len(mn.bindingHistory))
	copy(history, mn.bindingHistory)
	return history
}

// ResetBindingEvents clears all binding event history (useful for testing)
func (mn *MockNeuron) ResetBindingEvents() {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	mn.bindingEventCount = 0
	mn.bindingHistory = mn.bindingHistory[:0] // Clear slice but keep capacity
}

// MockSynapse represents a simple synapse for testing
type MockSynapse struct {
	id           string
	position     Position3D
	presynaptic  string
	postsynaptic string
	weight       float64
	activity     float64
	mu           sync.RWMutex
}

func NewMockSynapse(id string, pos Position3D, pre, post string, weight float64) *MockSynapse {
	return &MockSynapse{
		id:           id,
		position:     pos,
		presynaptic:  pre,
		postsynaptic: post,
		weight:       weight,
		activity:     0.0,
	}
}

// Implement Component interface
func (ms *MockSynapse) ID() string                   { return ms.id }
func (ms *MockSynapse) Position() Position3D         { return ms.position }
func (ms *MockSynapse) ComponentType() ComponentType { return ComponentSynapse }

// Helper methods for testing
func (ms *MockSynapse) GetWeight() float64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.weight
}

func (ms *MockSynapse) SetWeight(weight float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.weight = weight
}

func (ms *MockSynapse) GetActivity() float64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.activity
}

func (ms *MockSynapse) SetActivity(activity float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.activity = activity
}

func (ms *MockSynapse) GetPresynaptic() string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.presynaptic
}

func (ms *MockSynapse) GetPostsynaptic() string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.postsynaptic
}

// =================================================================================
// MOCK ASTROCYTE LISTENER FOR CALCIUM WAVE TESTING
// =================================================================================

// mockAstrocyteListener implements SignalListener for astrocyte calcium testing
type mockAstrocyteListener struct {
	id            string
	receivedCount int32
	receivedData  []interface{}
	receivedFrom  []string
	mu            sync.Mutex
}

func newMockAstrocyteListener(id string) *mockAstrocyteListener {
	return &mockAstrocyteListener{
		id:           id,
		receivedData: make([]interface{}, 0),
		receivedFrom: make([]string, 0),
	}
}

// ID implements SignalListener interface
func (m *mockAstrocyteListener) ID() string {
	return m.id
}

// OnSignal implements SignalListener interface
func (m *mockAstrocyteListener) OnSignal(signalType SignalType, sourceID string, data interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	atomic.AddInt32(&m.receivedCount, 1)
	m.receivedData = append(m.receivedData, data)
	m.receivedFrom = append(m.receivedFrom, sourceID)
}

// GetReceivedCount safely returns the number of received signals
func (m *mockAstrocyteListener) GetReceivedCount() int {
	return int(atomic.LoadInt32(&m.receivedCount))
}

// GetLastReceivedData returns the most recent signal data
func (m *mockAstrocyteListener) GetLastReceivedData() interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.receivedData) == 0 {
		return nil
	}
	return m.receivedData[len(m.receivedData)-1]
}
