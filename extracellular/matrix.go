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
	"fmt"
	"math"
	"sync"
	"time"
)

// Biological axon speed constants (μm/ms)
const (
	UNMYELINATED_SLOW = 500.0   // 0.5 m/s - C fibers
	UNMYELINATED_FAST = 2000.0  // 2 m/s - cortical axons
	MYELINATED_MEDIUM = 10000.0 // 10 m/s - A-delta fibers
	MYELINATED_FAST   = 80000.0 // 80 m/s - A-alpha fibers

	// Typical cortical ranges
	LOCAL_CIRCUIT = 2000.0  // Local cortical circuits
	INTER_LAMINAR = 5000.0  // Between cortical layers
	LONG_RANGE    = 15000.0 // Long-distance projections
)

// Global axon speed configuration with mutex for thread safety
var (
	axonSpeedMutex sync.RWMutex
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
// SPATIAL DELAY ENHANCEMENT
// =================================================================================

// EnhanceSynapticDelay calculates total delay including spatial propagation
// This is called by synapses to get the complete transmission delay
func (ecm *ExtracellularMatrix) EnhanceSynapticDelay(
	preNeuronID, postNeuronID, synapseID string,
	baseSynapticDelay time.Duration) time.Duration {

	// Get positions of pre and post-synaptic neurons
	preInfo, preExists := ecm.astrocyteNetwork.Get(preNeuronID)
	postInfo, postExists := ecm.astrocyteNetwork.Get(postNeuronID)

	if !preExists || !postExists {
		// If we can't find the neurons, return just the base delay
		return baseSynapticDelay
	}

	// Calculate 3D distance between neurons
	distance := ecm.calculateSpatialDistance(preInfo.Position, postInfo.Position)

	// Convert distance to propagation delay
	spatialDelay := ecm.calculatePropagationDelay(distance)

	// Return total delay: synaptic + spatial
	return baseSynapticDelay + spatialDelay
}

// GetSpatialDistance returns the 3D distance between two components
func (ecm *ExtracellularMatrix) GetSpatialDistance(componentID1, componentID2 string) (float64, error) {
	info1, exists1 := ecm.astrocyteNetwork.Get(componentID1)
	info2, exists2 := ecm.astrocyteNetwork.Get(componentID2)

	if !exists1 {
		return 0, fmt.Errorf("component %s not found", componentID1)
	}
	if !exists2 {
		return 0, fmt.Errorf("component %s not found", componentID2)
	}

	return ecm.calculateSpatialDistance(info1.Position, info2.Position), nil
}

// SetAxonSpeed allows customization of axon propagation speed
func (ecm *ExtracellularMatrix) SetAxonSpeed(speedUmPerMs float64) {
	axonSpeedMutex.Lock()
	defer axonSpeedMutex.Unlock()
	globalAxonSpeed = speedUmPerMs
}

// GetAxonSpeed returns the current axon speed
func (ecm *ExtracellularMatrix) GetAxonSpeed() float64 {
	axonSpeedMutex.RLock()
	defer axonSpeedMutex.RUnlock()
	return globalAxonSpeed
}

// calculateSpatialDistance computes 3D Euclidean distance between neurons
func (ecm *ExtracellularMatrix) calculateSpatialDistance(pos1, pos2 Position3D) float64 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	dz := pos1.Z - pos2.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// calculatePropagationDelay converts distance to time delay based on axon properties
func (ecm *ExtracellularMatrix) calculatePropagationDelay(distance float64) time.Duration {
	if distance <= 0 {
		return 0
	}

	// Get current axon speed safely
	axonSpeedMutex.RLock()
	speed := globalAxonSpeed
	axonSpeedMutex.RUnlock()

	// Calculate delay: distance / speed = time
	delayMs := distance / speed

	// Convert to time.Duration
	return time.Duration(delayMs * float64(time.Millisecond))
}

// SetBiologicalAxonType configures realistic axon speeds
func (ecm *ExtracellularMatrix) SetBiologicalAxonType(axonType string) {
	switch axonType {
	case "unmyelinated_slow":
		ecm.SetAxonSpeed(UNMYELINATED_SLOW)
	case "unmyelinated_fast":
		ecm.SetAxonSpeed(UNMYELINATED_FAST)
	case "cortical_local":
		ecm.SetAxonSpeed(LOCAL_CIRCUIT)
	case "cortical_inter":
		ecm.SetAxonSpeed(INTER_LAMINAR)
	case "long_range":
		ecm.SetAxonSpeed(LONG_RANGE)
	default:
		ecm.SetAxonSpeed(LOCAL_CIRCUIT) // Default to cortical local
	}
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

// Configurable axon speed for different network types
var globalAxonSpeed = 2000.0 // Default: 2000 μm/ms
