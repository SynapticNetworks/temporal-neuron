package neuron

import (
	"fmt"
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/message"
)

// ============================================================================
// Mock Ion Channel for Testing
// ============================================================================

// MockIonChannel provides a simple ion channel implementation for testing.
type MockIonChannel struct {
	name            string
	channelType     string
	ionType         IonType
	maxConductance  float64 // pS
	reversalPot     float64 // mV
	isOpen          bool
	openProbability float64
}

// NewMockIonChannel creates a new mock ion channel for testing.
func NewMockIonChannel(name string, ionType IonType, conductance, reversal float64) *MockIonChannel {
	return &MockIonChannel{
		name:            name,
		channelType:     "mock",
		ionType:         ionType,
		maxConductance:  conductance,
		reversalPot:     reversal,
		isOpen:          true, // Start open for testing
		openProbability: 1.0,
	}
}

// ModulateCurrent implements the IonChannel interface.
func (m *MockIonChannel) ModulateCurrent(msg message.NeuralSignal, voltage, calcium float64) (*message.NeuralSignal, bool, float64) {
	if !m.isOpen {
		return &msg, true, 0.0 // Pass through without channel current
	}

	// Calculate driving force (voltage - reversal potential)
	drivingForce := voltage - m.reversalPot

	// Calculate channel current: I = g * (Vm - Erev)
	channelCurrent := (m.maxConductance * 1e-12) * drivingForce * 1e12 // Convert to pA

	// Modulate the signal slightly
	modifiedMsg := msg
	modifiedMsg.Value *= 1.0 // No modification for testing

	return &modifiedMsg, true, channelCurrent
}

// ShouldOpen implements basic gating for the mock channel.
func (m *MockIonChannel) ShouldOpen(voltage, ligandConc, calcium float64, deltaTime time.Duration) (bool, time.Duration, float64) {
	// Simple voltage-dependent gating
	if voltage > -50.0 {
		return true, 10 * time.Millisecond, 0.9
	}
	return false, 0, 0.1
}

// UpdateKinetics implements state evolution for the mock channel.
func (m *MockIonChannel) UpdateKinetics(feedback *ChannelFeedback, deltaTime time.Duration, voltage float64) {
	// Simple implementation - could be enhanced for more realistic testing
	if feedback != nil && feedback.ContributedToFiring {
		m.openProbability *= 1.01 // Slight facilitation
	}
}

// GetConductance returns the channel's conductance.
func (m *MockIonChannel) GetConductance() float64 {
	return m.maxConductance
}

// GetReversalPotential returns the channel's reversal potential.
func (m *MockIonChannel) GetReversalPotential() float64 {
	return m.reversalPot
}

// GetIonSelectivity returns the channel's ion selectivity.
func (m *MockIonChannel) GetIonSelectivity() IonType {
	return m.ionType
}

// GetState returns the channel's current state.
func (m *MockIonChannel) GetState() ChannelState {
	return ChannelState{
		IsOpen:               m.isOpen,
		Conductance:          m.maxConductance,
		EquilibriumPotential: m.reversalPot,
	}
}

// GetTrigger returns the channel's gating properties.
func (m *MockIonChannel) GetTrigger() ChannelTrigger {
	return ChannelTrigger{
		ActivationVoltage: -50.0,
		VoltageSlope:      10.0,
	}
}

// Name returns the channel's name.
func (m *MockIonChannel) Name() string {
	return m.name
}

// ChannelType returns the channel's type.
func (m *MockIonChannel) ChannelType() string {
	return m.channelType
}

// Close cleans up the mock channel.
func (m *MockIonChannel) Close() {
	// No cleanup needed for mock
}

/*
=================================================================================
NEURON TESTING MOCKS - SIMPLE API CONTRACT DOCUMENTATION
=================================================================================

Minimal mocks for testing neuron functionality in isolation. These consolidate
and clean up the mocks scattered across test files while documenting the
exact API contracts between neuron and its direct dependencies: Matrix and Synapse.

=================================================================================
*/

// ============================================================================
// MATRIX MOCK - NEURON ↔ MATRIX API CONTRACT
// ============================================================================

// MockMatrix simulates matrix coordination for neuron testing
type MockMatrix struct {
	mu sync.Mutex

	// Simple interaction tracking
	healthReports     []HealthReport
	chemicalReleases  []ChemicalRelease
	electricalSignals []ElectricalSignal
	synapseCreations  []string

	// Error injection for testing
	createSynapseError error
}

// API contract types
type HealthReport struct {
	ActivityLevel   float64
	ConnectionCount int
}

type ChemicalRelease struct {
	LigandType    message.LigandType
	Concentration float64
}

type ElectricalSignal struct {
	SignalType message.SignalType
	Data       interface{}
}

func NewMockMatrix() *MockMatrix {
	return &MockMatrix{
		healthReports:     make([]HealthReport, 0),
		chemicalReleases:  make([]ChemicalRelease, 0),
		electricalSignals: make([]ElectricalSignal, 0),
		synapseCreations:  make([]string, 0),
	}
}

// CreateBasicCallbacks returns the standard neuron callbacks
// API CONTRACT: How neurons communicate with matrix
func (mm *MockMatrix) CreateBasicCallbacks() NeuronCallbacks {
	return NeuronCallbacks{
		ReportHealth: func(activityLevel float64, connectionCount int) {
			mm.mu.Lock()
			defer mm.mu.Unlock()
			mm.healthReports = append(mm.healthReports, HealthReport{
				ActivityLevel:   activityLevel,
				ConnectionCount: connectionCount,
			})
		},

		ReleaseChemical: func(ligandType message.LigandType, concentration float64) error {
			mm.mu.Lock()
			defer mm.mu.Unlock()
			mm.chemicalReleases = append(mm.chemicalReleases, ChemicalRelease{
				LigandType:    ligandType,
				Concentration: concentration,
			})
			return nil
		},

		SendElectricalSignal: func(signalType message.SignalType, data interface{}) {
			mm.mu.Lock()
			defer mm.mu.Unlock()
			mm.electricalSignals = append(mm.electricalSignals, ElectricalSignal{
				SignalType: signalType,
				Data:       data,
			})
		},

		CreateSynapse: func(config SynapseCreationConfig) (string, error) {
			mm.mu.Lock()
			defer mm.mu.Unlock()

			if mm.createSynapseError != nil {
				return "", mm.createSynapseError
			}

			synapseID := fmt.Sprintf("synapse-%d", len(mm.synapseCreations))
			mm.synapseCreations = append(mm.synapseCreations, config.TargetNeuronID)
			return synapseID, nil
		},

		// Minimal implementations to prevent panics
		ListSynapses: func(criteria SynapseCriteria) []SynapseInfo {
			return []SynapseInfo{}
		},

		FindNearbyComponents: func(radius float64) []component.ComponentInfo {
			return []component.ComponentInfo{}
		},

		GetMatrix: func() ExtracellularMatrix {
			return &MockExtracellularMatrix{}
		},
	}
}

// Test query methods
func (mm *MockMatrix) GetHealthReportCount() int {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	return len(mm.healthReports)
}

func (mm *MockMatrix) GetChemicalReleaseCount() int {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	return len(mm.chemicalReleases)
}

func (mm *MockMatrix) GetElectricalSignalCount() int {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	return len(mm.electricalSignals)
}

// GetChemicalReleases returns all chemical releases
func (mm *MockMatrix) GetChemicalReleases() []ChemicalRelease {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	return append([]ChemicalRelease{}, mm.chemicalReleases...)
}

func (mm *MockMatrix) GetSynapseCreationCount() int {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	return len(mm.synapseCreations)
}

// Configuration for error testing
func (mm *MockMatrix) SetCreateSynapseError(err error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.createSynapseError = err
}

// ============================================================================
// EXTRACELLULAR MATRIX MOCK
// ============================================================================

type MockExtracellularMatrix struct{}

func (mem *MockExtracellularMatrix) EnhanceSynapticDelay(preNeuronID, postNeuronID, synapseID string, baseDelay time.Duration) time.Duration {
	return baseDelay + 1*time.Millisecond // Simple enhancement
}

// ============================================================================
// SYNAPSE MOCK - NEURON ↔ SYNAPSE API CONTRACT
// ============================================================================

// MockSynapse simulates synaptic connections for output testing
type MockSynapse struct {
	mu sync.Mutex

	// Identity
	id       string
	targetID string
	weight   float64
	delay    time.Duration

	// Interaction tracking
	receivedSignals []message.NeuralSignal

	// Error injection
	transmissionError error
}

func NewMockSynapse(id, targetID string, weight float64, delay time.Duration) *MockSynapse {
	return &MockSynapse{
		id:              id,
		targetID:        targetID,
		weight:          weight,
		delay:           delay,
		receivedSignals: make([]message.NeuralSignal, 0),
	}
}

// CreateOutputCallback returns OutputCallback for this synapse
// API CONTRACT: How neurons transmit signals through synapses
func (ms *MockSynapse) CreateOutputCallback() OutputCallback {
	return OutputCallback{
		TransmitMessage: func(msg message.NeuralSignal) error {
			ms.mu.Lock()
			defer ms.mu.Unlock()

			if ms.transmissionError != nil {
				return ms.transmissionError
			}

			ms.receivedSignals = append(ms.receivedSignals, msg)
			return nil
		},

		GetWeight: func() float64 {
			ms.mu.Lock()
			defer ms.mu.Unlock()
			return ms.weight
		},

		GetDelay: func() time.Duration {
			ms.mu.Lock()
			defer ms.mu.Unlock()
			return ms.delay
		},

		GetTargetID: func() string {
			ms.mu.Lock()
			defer ms.mu.Unlock()
			return ms.targetID
		},
	}
}

// ReceiveSignal - compatibility method for existing tests
func (ms *MockSynapse) ReceiveSignal(signal message.NeuralSignal) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.receivedSignals = append(ms.receivedSignals, signal)
}

// Test query methods
func (ms *MockSynapse) GetReceivedCount() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return len(ms.receivedSignals)
}

func (ms *MockSynapse) GetReceivedSignalCount() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return len(ms.receivedSignals)
}

func (ms *MockSynapse) GetReceivedSignals() []message.NeuralSignal {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return append([]message.NeuralSignal{}, ms.receivedSignals...)
}

// Configuration for error testing
func (ms *MockSynapse) SetTransmissionError(err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.transmissionError = err
}

func (ms *MockSynapse) SetWeight(weight float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.weight = weight
}

// ============================================================================
// HELPER FUNCTIONS FOR HOMEOSTATIC TESTING
// ============================================================================

// CreateMockConnection adds a mock synapse to a neuron for testing
func CreateMockConnection(neuron *Neuron, synapseID, targetID string, weight float64) *MockSynapse {
	mockSynapse := NewMockSynapse(synapseID, targetID, weight, 1*time.Millisecond)
	neuron.AddOutputCallback(synapseID, mockSynapse.CreateOutputCallback())
	return mockSynapse
}

// SendTestSignal sends a signal to neuron for testing
func SendTestSignal(neuron *Neuron, sourceID string, value float64) {
	signal := message.NeuralSignal{
		Value:                value,
		Timestamp:            time.Now(),
		SourceID:             sourceID,
		TargetID:             neuron.ID(),
		NeurotransmitterType: message.LigandGlutamate,
	}
	neuron.Receive(signal)
}
