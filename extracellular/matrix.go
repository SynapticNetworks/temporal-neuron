/*
=================================================================================
EXTRACELLULAR MATRIX - BIOLOGICAL COORDINATION LAYER
=================================================================================

A generic coordination system inspired by the brain's extracellular matrix.
Components communicate through chemical signaling and discrete signals,
just like real neural networks. No technical abstractions - pure biology.
=================================================================================
*/

package extracellular

import (
	"context"
	"sync"
	"time"
)

// Matrix provides biological coordination services for autonomous components
type Matrix struct {
	// === CORE BIOLOGICAL SYSTEMS ===
	registry  *ComponentRegistry // Who exists in the network
	modulator *ChemicalModulator // Chemical signaling (neurotransmitters)
	signaling *SignalCoordinator // Discrete signal routing
	lifecycle *LifecycleManager  // Birth/death coordination
	discovery *DiscoveryService  // Finding other components
	plugins   *PluginManager     // Modular functionality

	// === OPERATIONAL STATE ===
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.RWMutex
}

// MatrixConfig provides basic configuration
type MatrixConfig struct {
	ChemicalEnabled bool
	SpatialEnabled  bool
	UpdateInterval  time.Duration
	MaxComponents   int
}

// NewMatrix creates a coordination matrix
func NewMatrix(config MatrixConfig) *Matrix {
	ctx, cancel := context.WithCancel(context.Background())

	registry := NewComponentRegistry()
	modulator := NewChemicalModulator(registry)
	signaling := NewSignalCoordinator()
	lifecycle := NewLifecycleManager(registry)
	discovery := NewDiscoveryService(registry)
	plugins := NewPluginManager()

	return &Matrix{
		registry:  registry,
		modulator: modulator,
		signaling: signaling,
		lifecycle: lifecycle,
		discovery: discovery,
		plugins:   plugins,
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start begins coordination services
func (m *Matrix) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return nil
	}

	m.started = true
	return nil
}

// Stop ends coordination services
func (m *Matrix) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.started {
		return nil
	}

	m.cancel()
	m.started = false
	return nil
}

// =================================================================================
// CHEMICAL SIGNALING (like neurotransmitters)
// =================================================================================

// ReleaseLigand releases a chemical signal
func (m *Matrix) ReleaseLigand(ligandType LigandType, sourceID string, concentration float64) error {
	return m.modulator.Release(ligandType, sourceID, concentration)
}

// RegisterForBinding registers to receive chemical signals
func (m *Matrix) RegisterForBinding(target BindingTarget) error {
	return m.modulator.RegisterTarget(target)
}

// =================================================================================
// DISCRETE SIGNALING (like action potentials)
// =================================================================================

// SendSignal sends a discrete signal to listeners
func (m *Matrix) SendSignal(signalType SignalType, sourceID string, data interface{}) {
	m.signaling.Send(signalType, sourceID, data)
}

// ListenForSignals registers to receive discrete signals
func (m *Matrix) ListenForSignals(signalTypes []SignalType, listener SignalListener) {
	m.signaling.AddListener(signalTypes, listener)
}

// =================================================================================
// COMPONENT MANAGEMENT
// =================================================================================

// RegisterComponent adds a component to the network
func (m *Matrix) RegisterComponent(info ComponentInfo) error {
	return m.registry.Register(info)
}

// FindComponents searches for components
func (m *Matrix) FindComponents(criteria ComponentCriteria) []ComponentInfo {
	return m.registry.Find(criteria)
}

// =================================================================================
// COMMON TYPES AND INTERFACES
// =================================================================================

// Position3D represents spatial coordinates
type Position3D struct {
	X, Y, Z float64
}

// LigandType represents chemical signal types (like neurotransmitters)
type LigandType int

const (
	LigandGlutamate LigandType = iota
	LigandGABA
	LigandDopamine
	LigandSerotonin
	LigandAcetylcholine
)

// SignalType represents discrete signal types (like firing events)
type SignalType int

const (
	SignalFired SignalType = iota
	SignalConnected
	SignalDisconnected
	SignalThresholdChanged
)

// BindingTarget receives chemical signals (like having receptors)
type BindingTarget interface {
	Bind(ligandType LigandType, sourceID string, concentration float64)
	GetReceptors() []LigandType
	GetPosition() Position3D
}

// SignalListener receives discrete signals
type SignalListener interface {
	OnSignal(signalType SignalType, sourceID string, data interface{})
}

// ComponentInfo holds basic component information
type ComponentInfo struct {
	ID           string
	Type         ComponentType
	Position     Position3D
	State        ComponentState
	Metadata     map[string]interface{}
	RegisteredAt time.Time
}

// ComponentType categorizes components
type ComponentType int

const (
	ComponentNeuron ComponentType = iota
	ComponentSynapse
	ComponentGate
	ComponentPlugin
)

// ComponentState tracks lifecycle
type ComponentState int

const (
	StateActive ComponentState = iota
	StateInactive
	StateShuttingDown
)

// ComponentCriteria for searching
type ComponentCriteria struct {
	Type     *ComponentType
	State    *ComponentState
	Position *Position3D
	Radius   float64
}
