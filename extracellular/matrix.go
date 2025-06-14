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

// ExtracellularMatrix provides biological coordination services for autonomous components
type ExtracellularMatrix struct {
	// === CORE BIOLOGICAL SYSTEMS ===
	astrocyteNetwork  *AstrocyteNetwork  // Who exists + connectivity (was registry + discovery)
	chemicalModulator *ChemicalModulator // Chemical signaling (neurotransmitters)
	gapJunctions      *GapJunctions      // Discrete signal routing (was signaling)
	microglia         *Microglia         // Birth/death coordination (was lifecycle)
	plugins           *PluginManager     // Modular functionality

	// === OPERATIONAL STATE ===
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.RWMutex
}

// ExtracellularMatrixConfig provides basic configuration
type ExtracellularMatrixConfig struct {
	ChemicalEnabled bool
	SpatialEnabled  bool
	UpdateInterval  time.Duration
	MaxComponents   int
}

// NewExtracellularMatrix creates a coordination matrix
func NewExtracellularMatrix(config ExtracellularMatrixConfig) *ExtracellularMatrix {
	ctx, cancel := context.WithCancel(context.Background())

	astrocyteNetwork := NewAstrocyteNetwork()           // was NewComponentRegistry()
	modulator := NewChemicalModulator(astrocyteNetwork) // was NewChemicalModulator(registry)
	gapJunctions := NewGapJunctions()                   // was NewSignalCoordinator()
	microglia := NewMicroglia(astrocyteNetwork)         // was NewLifecycleManager(registry)
	plugins := NewPluginManager()

	return &ExtracellularMatrix{
		astrocyteNetwork:  astrocyteNetwork,
		chemicalModulator: modulator,
		gapJunctions:      gapJunctions,
		microglia:         microglia,
		plugins:           plugins,
		ctx:               ctx,
		cancel:            cancel,
	}
}

// Start begins coordination services
func (ecm *ExtracellularMatrix) Start() error {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	if ecm.started {
		return nil
	}

	ecm.started = true
	return nil
}

// Stop ends coordination services
func (ecm *ExtracellularMatrix) Stop() error {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	if !ecm.started {
		return nil
	}

	ecm.cancel()
	ecm.started = false
	return nil
}

// =================================================================================
// CHEMICAL SIGNALING (like neurotransmitters) - KEEP YOUR EXCELLENT API
// =================================================================================

// ReleaseLigand releases a chemical signal
func (ecm *ExtracellularMatrix) ReleaseLigand(ligandType LigandType, sourceID string, concentration float64) error {
	return ecm.chemicalModulator.Release(ligandType, sourceID, concentration)
}

// RegisterForBinding registers to receive chemical signals
func (ecm *ExtracellularMatrix) RegisterForBinding(target BindingTarget) error {
	return ecm.chemicalModulator.RegisterTarget(target)
}

// =================================================================================
// DISCRETE SIGNALING (like action potentials) - KEEP YOUR EXCELLENT API
// =================================================================================

// SendSignal sends a discrete signal to listeners
func (ecm *ExtracellularMatrix) SendSignal(signalType SignalType, sourceID string, data interface{}) {
	ecm.gapJunctions.Send(signalType, sourceID, data)
}

// ListenForSignals registers to receive discrete signals
func (ecm *ExtracellularMatrix) ListenForSignals(signalTypes []SignalType, listener SignalListener) {
	ecm.gapJunctions.AddListener(signalTypes, listener)
}

// =================================================================================
// COMPONENT MANAGEMENT - USING ASTROCYTE NETWORK
// =================================================================================

// RegisterComponent adds a component to the network
func (ecm *ExtracellularMatrix) RegisterComponent(info ComponentInfo) error {
	return ecm.astrocyteNetwork.Register(info)
}

// FindComponents searches for components
func (ecm *ExtracellularMatrix) FindComponents(criteria ComponentCriteria) []ComponentInfo {
	return ecm.astrocyteNetwork.Find(criteria)
}
