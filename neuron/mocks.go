package neuron

import (
	"fmt"
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/types"
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
func (m *MockIonChannel) ModulateCurrent(msg types.NeuralSignal, voltage, calcium float64) (*types.NeuralSignal, bool, float64) {
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
	synapseCreations  []SynapseCreation

	// Enhanced callback tracking
	synapseList      []types.SynapseInfo
	spatialDelays    map[string]time.Duration
	nearbyComponents []component.ComponentInfo

	plasticityAdjustments []types.PlasticityAdjustment

	// Error injection for testing
	createSynapseError    error
	deleteSynapseError    error
	applePlasticityError  error
	setSynapseWeightError error
	getSynapseError       error
}

// API contract types
type HealthReport struct {
	ActivityLevel   float64
	ConnectionCount int
}

type ElectricalSignal struct {
	SignalType types.SignalType
	Data       interface{}
}

type SynapseCreation struct {
	Config types.SynapseCreationConfig
	ID     string
}

// SetSynapseList allows tests to configure what synapses ListSynapses will return
func (mm *MockMatrix) SetSynapseList(synapses []types.SynapseInfo) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Make a deep copy to avoid external modifications
	mm.synapseList = make([]types.SynapseInfo, len(synapses))
	copy(mm.synapseList, synapses)
}

// GetPlasticityAdjustments returns recorded STDP adjustments for testing
func (mm *MockMatrix) GetPlasticityAdjustments() []types.PlasticityAdjustment {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Return a copy to prevent external modification
	result := make([]types.PlasticityAdjustment, len(mm.plasticityAdjustments))
	copy(result, mm.plasticityAdjustments)
	return result
}

// ClearPlasticityAdjustments resets the record of plasticity adjustments
func (mm *MockMatrix) ClearPlasticityAdjustments() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.plasticityAdjustments = mm.plasticityAdjustments[:0]
}

// Modify the ApplyPlasticity method to record adjustments for testing
// This is in the MockNeuronCallbacks implementation
func (mnc *MockNeuronCallbacks) ApplyPlasticity(synapseID string, adjustment types.PlasticityAdjustment) error {
	mnc.matrix.mu.Lock()
	defer mnc.matrix.mu.Unlock()

	if mnc.matrix.applePlasticityError != nil {
		return mnc.matrix.applePlasticityError
	}

	// Record the adjustment for testing
	mnc.matrix.plasticityAdjustments = append(mnc.matrix.plasticityAdjustments, adjustment)

	// Find and update synapse
	for i, synapse := range mnc.matrix.synapseList {
		if synapse.ID == synapseID {
			// Apply simple plasticity adjustment
			mnc.matrix.synapseList[i].Weight += adjustment.WeightChange
			if mnc.matrix.synapseList[i].Weight < 0 {
				mnc.matrix.synapseList[i].Weight = 0
			}
			break
		}
	}
	return nil
}

func NewMockMatrix() *MockMatrix {
	m := &MockMatrix{
		healthReports:         make([]HealthReport, 0),
		chemicalReleases:      make([]ChemicalRelease, 0),
		electricalSignals:     make([]ElectricalSignal, 0),
		synapseCreations:      make([]SynapseCreation, 0),
		synapseList:           make([]types.SynapseInfo, 0),
		spatialDelays:         make(map[string]time.Duration),
		nearbyComponents:      make([]component.ComponentInfo, 0),
		plasticityAdjustments: make([]types.PlasticityAdjustment, 0),
	}

	// Make sure synapseList is initialized
	if m.synapseList == nil {
		m.synapseList = make([]types.SynapseInfo, 0)
	}

	return m
}

// ============================================================================
// FIXED: MockNeuronCallbacks implementing component.NeuronCallbacks interface
// ============================================================================

// MockNeuronCallbacks implements the component.NeuronCallbacks interface for testing
type MockNeuronCallbacks struct {
	matrix *MockMatrix
}

// NewMockNeuronCallbacks creates a new mock callbacks instance
func NewMockNeuronCallbacks(matrix *MockMatrix) *MockNeuronCallbacks {
	return &MockNeuronCallbacks{matrix: matrix}
}

// INTERFACE METHODS - These implement component.NeuronCallbacks

func (mnc *MockNeuronCallbacks) CreateSynapse(config types.SynapseCreationConfig) (string, error) {
	mnc.matrix.mu.Lock()
	defer mnc.matrix.mu.Unlock()

	if mnc.matrix.createSynapseError != nil {
		return "", mnc.matrix.createSynapseError
	}

	synapseID := fmt.Sprintf("synapse-%d", len(mnc.matrix.synapseCreations))
	creation := SynapseCreation{
		Config: config,
		ID:     synapseID,
	}
	mnc.matrix.synapseCreations = append(mnc.matrix.synapseCreations, creation)
	return synapseID, nil
}

func (mnc *MockNeuronCallbacks) DeleteSynapse(synapseID string) error {
	mnc.matrix.mu.Lock()
	defer mnc.matrix.mu.Unlock()

	if mnc.matrix.deleteSynapseError != nil {
		return mnc.matrix.deleteSynapseError
	}

	// Remove from synapse list
	for i, synapse := range mnc.matrix.synapseList {
		if synapse.ID == synapseID {
			mnc.matrix.synapseList = append(mnc.matrix.synapseList[:i], mnc.matrix.synapseList[i+1:]...)
			break
		}
	}
	return nil
}

func (mnc *MockNeuronCallbacks) ListSynapses(criteria types.SynapseCriteria) []types.SynapseInfo {
	mnc.matrix.mu.Lock()
	defer mnc.matrix.mu.Unlock()

	// Simple filtering - in real implementation would be more sophisticated
	result := make([]types.SynapseInfo, 0)
	for _, synapse := range mnc.matrix.synapseList {
		// Simple criteria matching
		if criteria.SourceID != nil && synapse.SourceID != *criteria.SourceID {
			continue
		}
		if criteria.TargetID != nil && synapse.TargetID != *criteria.TargetID {
			continue
		}
		result = append(result, synapse)
	}
	return result
}

func (mnc *MockNeuronCallbacks) ReleaseChemical(ligandType types.LigandType, concentration float64) error {
	mnc.matrix.mu.Lock()
	defer mnc.matrix.mu.Unlock()

	mnc.matrix.chemicalReleases = append(mnc.matrix.chemicalReleases, ChemicalRelease{
		LigandType:    ligandType,
		Concentration: concentration,
	})
	return nil
}

func (mnc *MockNeuronCallbacks) SendElectricalSignal(signalType types.SignalType, data interface{}) {
	mnc.matrix.mu.Lock()
	defer mnc.matrix.mu.Unlock()

	mnc.matrix.electricalSignals = append(mnc.matrix.electricalSignals, ElectricalSignal{
		SignalType: signalType,
		Data:       data,
	})
}

func (mnc *MockNeuronCallbacks) ReportHealth(activityLevel float64, connectionCount int) {
	mnc.matrix.mu.Lock()
	defer mnc.matrix.mu.Unlock()

	mnc.matrix.healthReports = append(mnc.matrix.healthReports, HealthReport{
		ActivityLevel:   activityLevel,
		ConnectionCount: connectionCount,
	})
}

func (mnc *MockNeuronCallbacks) GetSpatialDelay(targetID string) time.Duration {
	mnc.matrix.mu.Lock()
	defer mnc.matrix.mu.Unlock()

	if delay, exists := mnc.matrix.spatialDelays[targetID]; exists {
		return delay
	}
	return 1 * time.Millisecond // Default delay
}

// ============================================================================
// ENHANCED METHODS (beyond basic interface)
// ============================================================================

func (mnc *MockNeuronCallbacks) GetSynapseWeight(synapseID string) (float64, error) {
	mnc.matrix.mu.Lock()
	defer mnc.matrix.mu.Unlock()

	for _, synapse := range mnc.matrix.synapseList {
		if synapse.ID == synapseID {
			return synapse.Weight, nil
		}
	}
	return 0.0, fmt.Errorf("synapse %s not found", synapseID)
}

func (mnc *MockNeuronCallbacks) SetSynapseWeight(synapseID string, weight float64) error {
	mnc.matrix.mu.Lock()
	defer mnc.matrix.mu.Unlock()

	if mnc.matrix.setSynapseWeightError != nil {
		return mnc.matrix.setSynapseWeightError
	}

	for i, synapse := range mnc.matrix.synapseList {
		if synapse.ID == synapseID {
			mnc.matrix.synapseList[i].Weight = weight
			return nil
		}
	}
	return fmt.Errorf("synapse %s not found", synapseID)
}

func (mnc *MockNeuronCallbacks) GetSynapse(synapseID string) (component.SynapticProcessor, error) {
	mnc.matrix.mu.Lock()
	defer mnc.matrix.mu.Unlock()

	if mnc.matrix.getSynapseError != nil {
		return nil, mnc.matrix.getSynapseError
	}

	// Return a mock synaptic processor
	return &MockSynapticProcessor{synapseID: synapseID}, nil
}

func (mnc *MockNeuronCallbacks) GetMatrix() component.ExtracellularMatrix {
	return &MockExtracellularMatrix{}
}

func (mnc *MockNeuronCallbacks) FindNearbyComponents(radius float64) []component.ComponentInfo {
	mnc.matrix.mu.Lock()
	defer mnc.matrix.mu.Unlock()
	return append([]component.ComponentInfo{}, mnc.matrix.nearbyComponents...)
}

func (mnc *MockNeuronCallbacks) ReportStateChange(oldState, newState types.ComponentState) {
	// Simple implementation for testing
}

// ============================================================================
// MOCK SYNAPTIC PROCESSOR
// ============================================================================
// ============================================================================
// COMPLETE AND CORRECT MOCK SYNAPTIC PROCESSOR
// ============================================================================

type MockSynapticProcessor struct {
	synapseID   string
	weight      float64
	shouldPrune bool

	// Tracking for testing
	plasticityApplications []types.PlasticityAdjustment
	transmissions          []float64
	lastTransmission       time.Time
	mutex                  sync.Mutex
}

// NewMockSynapticProcessor creates a properly initialized mock
func NewMockSynapticProcessor(synapseID string) *MockSynapticProcessor {
	return &MockSynapticProcessor{
		synapseID:              synapseID,
		weight:                 1.0, // Default weight
		shouldPrune:            false,
		plasticityApplications: make([]types.PlasticityAdjustment, 0),
		transmissions:          make([]float64, 0),
		lastTransmission:       time.Now(),
	}
}

// ============================================================================
// IMPLEMENT component.SynapticProcessor INTERFACE - ALL REQUIRED METHODS
// ============================================================================

// ID implements component.SynapticProcessor
func (msp *MockSynapticProcessor) ID() string {
	return msp.synapseID
}

func (msp *MockSynapticProcessor) Type() types.ComponentType {
	return types.TypeSynapse
}

func (ms *MockSynapticProcessor) UpdateWeight(event types.PlasticityEvent) {
}

// ID implements component.SynapticProcessor
func (msp *MockSynapticProcessor) IsActive() bool {
	return true
}

// GetPlasticityConfig returns plasticity parameters
func (msp *MockSynapticProcessor) GetPlasticityConfig() types.PlasticityConfig {
	return types.PlasticityConfig{}
}

// GetLastActivity returns the timestamp of the most recent transmission
func (msp *MockSynapticProcessor) GetLastActivity() time.Time {
	return time.Now()
}

func (msp *MockSynapticProcessor) GetDelay() time.Duration {
	return 0 * time.Second
}

func (msp *MockSynapticProcessor) GetPostsynapticID() string {
	return ""
}

func (msp *MockSynapticProcessor) GetPresynapticID() string {
	return ""
}

func (msp *MockSynapticProcessor) Position() types.Position3D {
	return types.Position3D{}
}

// Transmit implements component.SynapticProcessor - THIS WAS MISSING!
func (msp *MockSynapticProcessor) Transmit(signalValue float64) {
	msp.mutex.Lock()
	defer msp.mutex.Unlock()

	// Store the transmission for testing verification
	msp.transmissions = append(msp.transmissions, signalValue)
	msp.lastTransmission = time.Now()

	// In a real implementation, this would schedule delayed delivery
	// For testing, we just record the signal
}

// ApplyPlasticity implements component.SynapticProcessor
func (msp *MockSynapticProcessor) ApplyPlasticity(adjustment types.PlasticityAdjustment) {
	msp.mutex.Lock()
	defer msp.mutex.Unlock()

	// Store the plasticity adjustment for testing verification
	msp.plasticityApplications = append(msp.plasticityApplications, adjustment)

	// Apply the weight change
	msp.weight += adjustment.WeightChange

	// Clamp weight to reasonable bounds
	if msp.weight < 0.0 {
		msp.weight = 0.0
	}
	if msp.weight > 10.0 {
		msp.weight = 10.0
	}

}

// ShouldPrune implements component.SynapticProcessor
func (msp *MockSynapticProcessor) ShouldPrune() bool {
	msp.mutex.Lock()
	defer msp.mutex.Unlock()
	return msp.shouldPrune
}

// GetWeight implements component.SynapticProcessor
func (msp *MockSynapticProcessor) GetWeight() float64 {
	msp.mutex.Lock()
	defer msp.mutex.Unlock()
	return msp.weight
}

// SetWeight implements component.SynapticProcessor
func (msp *MockSynapticProcessor) SetWeight(weight float64) {
	msp.mutex.Lock()
	defer msp.mutex.Unlock()
	msp.weight = weight
}

// ============================================================================
// ADDITIONAL METHODS - Check if these are required by your interface
// ============================================================================

// GetActivityInfo returns activity information using CORRECT types.ActivityInfo struct
func (msp *MockSynapticProcessor) GetActivityInfo() types.ActivityInfo {
	msp.mutex.Lock()
	defer msp.mutex.Unlock()

	now := time.Now()

	return types.ActivityInfo{
		ComponentID:           msp.synapseID,
		LastTransmission:      msp.lastTransmission,
		LastPlasticity:        now, // Mock doesn't track plasticity events precisely
		Weight:                msp.weight,
		ActivityLevel:         0.5, // Mock activity level
		TimeSinceTransmission: now.Sub(msp.lastTransmission),
		TimeSincePlasticity:   time.Duration(0), // Mock doesn't track this
		ConnectionCount:       0,                // Not applicable for synapses
	}
}

// ============================================================================
// TESTING HELPER METHODS
// ============================================================================

// SetShouldPrune configures whether this synapse should be pruned
func (msp *MockSynapticProcessor) SetShouldPrune(shouldPrune bool) {
	msp.mutex.Lock()
	defer msp.mutex.Unlock()
	msp.shouldPrune = shouldPrune
}

// GetPlasticityApplications returns all plasticity adjustments applied to this synapse
func (msp *MockSynapticProcessor) GetPlasticityApplications() []types.PlasticityAdjustment {
	msp.mutex.Lock()
	defer msp.mutex.Unlock()
	return append([]types.PlasticityAdjustment{}, msp.plasticityApplications...)
}

// GetPlasticityApplicationCount returns the number of plasticity applications
func (msp *MockSynapticProcessor) GetPlasticityApplicationCount() int {
	msp.mutex.Lock()
	defer msp.mutex.Unlock()
	return len(msp.plasticityApplications)
}

// GetTransmissions returns all transmitted signals
func (msp *MockSynapticProcessor) GetTransmissions() []float64 {
	msp.mutex.Lock()
	defer msp.mutex.Unlock()
	return append([]float64{}, msp.transmissions...)
}

// GetTransmissionCount returns the number of transmissions
func (msp *MockSynapticProcessor) GetTransmissionCount() int {
	msp.mutex.Lock()
	defer msp.mutex.Unlock()
	return len(msp.transmissions)
}

// ClearHistory clears all recorded history for testing
func (msp *MockSynapticProcessor) ClearHistory() {
	msp.mutex.Lock()
	defer msp.mutex.Unlock()
	msp.plasticityApplications = msp.plasticityApplications[:0]
	msp.transmissions = msp.transmissions[:0]
}

// ============================================================================
// MATRIX CONFIGURATION AND TESTING METHODS
// ============================================================================

// CreateBasicCallbacks returns the standard neuron callbacks
// API CONTRACT: How neurons communicate with matrix
func (mm *MockMatrix) CreateBasicCallbacks() component.NeuronCallbacks {
	return NewMockNeuronCallbacks(mm)
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

func (mm *MockMatrix) GetSynapseCreations() []SynapseCreation {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	return append([]SynapseCreation{}, mm.synapseCreations...)
}

// Configuration methods for testing scenarios
func (mm *MockMatrix) SetCreateSynapseError(err error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.createSynapseError = err
}

func (mm *MockMatrix) SetDeleteSynapseError(err error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.deleteSynapseError = err
}

func (mm *MockMatrix) SetApplyPlasticityError(err error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.applePlasticityError = err
}

func (mm *MockMatrix) SetSetSynapseWeightError(err error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.setSynapseWeightError = err
}

func (mm *MockMatrix) SetGetSynapseError(err error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.getSynapseError = err
}

func (mm *MockMatrix) AddSynapse(synapse types.SynapseInfo) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.synapseList = append(mm.synapseList, synapse)
}

func (mm *MockMatrix) SetSpatialDelay(targetID string, delay time.Duration) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.spatialDelays[targetID] = delay
}

func (mm *MockMatrix) AddNearbyComponent(info component.ComponentInfo) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.nearbyComponents = append(mm.nearbyComponents, info)
}

// ============================================================================
// EXTRACELLULAR MATRIX MOCK
// ============================================================================

type MockExtracellularMatrix struct{}

func (mem *MockExtracellularMatrix) SynapticDelay(preNeuronID, postNeuronID, synapseID string, baseDelay time.Duration) time.Duration {
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
	receivedSignals []types.NeuralSignal

	// Error injection
	transmissionError error
}

func NewMockSynapse(id, targetID string, weight float64, delay time.Duration) *MockSynapse {
	return &MockSynapse{
		id:              id,
		targetID:        targetID,
		weight:          weight,
		delay:           delay,
		receivedSignals: make([]types.NeuralSignal, 0),
	}
}

// CreateOutputCallback returns OutputCallback for this synapse
// API CONTRACT: How neurons transmit signals through synapses
func (ms *MockSynapse) CreateOutputCallback() types.OutputCallback {
	return types.OutputCallback{
		TransmitMessage: func(msg types.NeuralSignal) error {
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
func (ms *MockSynapse) ReceiveSignal(signal types.NeuralSignal) {
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

func (ms *MockSynapse) GetReceivedSignals() []types.NeuralSignal {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return append([]types.NeuralSignal{}, ms.receivedSignals...)
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
	signal := types.NeuralSignal{
		Value:                value,
		Timestamp:            time.Now(),
		SourceID:             sourceID,
		TargetID:             neuron.ID(),
		NeurotransmitterType: types.LigandGlutamate,
	}
	neuron.Receive(signal)
}
