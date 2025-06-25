package extracellular

import (
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component" // Import component package
	// Import message package
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// =================================================================================
// TYPE ALIASES AND IMPORTS FROM EXTERNAL PACKAGES
// =================================================================================

// Import types from external packages to create clean separation layer
type LigandType = types.LigandType
type SignalType = types.SignalType
type Position3D = types.Position3D

// Re-export constants for local use without needing message prefix
const (
	LigandNone            = types.LigandNone
	LigandGlutamate       = types.LigandGlutamate
	LigandGABA            = types.LigandGABA
	LigandDopamine        = types.LigandDopamine
	LigandSerotonin       = types.LigandSerotonin
	LigandAcetylcholine   = types.LigandAcetylcholine
	LigandNorepinephrine  = types.LigandNorepinephrine
	LigandHistamine       = types.LigandHistamine
	LigandGlycine         = types.LigandGlycine
	LigandAdenosine       = types.LigandAdenosine
	LigandNitricOxide     = types.LigandNitricOxide
	LigandEndocannabinoid = types.LigandEndocannabinoid
	LigandNeuropeptideY   = types.LigandNeuropeptideY
	LigandSubstanceP      = types.LigandSubstanceP
	LigandVasopressin     = types.LigandVasopressin
	LigandOxytocin        = types.LigandOxytocin
	LigandCalcium         = types.LigandCalcium // Add calcium for plasticity
)

const (
	SignalNone               = types.SignalNone
	SignalFired              = types.SignalFired
	SignalConnected          = types.SignalConnected
	SignalDisconnected       = types.SignalDisconnected
	SignalThresholdChanged   = types.SignalThresholdChanged
	SignalSynchronization    = types.SignalSynchronization
	SignalCalciumWave        = types.SignalCalciumWave
	SignalPlasticityEvent    = types.SignalPlasticityEvent
	SignalChemicalGradient   = types.SignalChemicalGradient
	SignalStructuralChange   = types.SignalStructuralChange
	SignalMetabolicState     = types.SignalMetabolicState
	SignalHealthWarning      = types.SignalHealthWarning
	SignalNetworkOscillation = types.SignalNetworkOscillation
)

// =================================================================================
// FACTORY PATTERN INTERFACES
// =================================================================================

// NeuralComponent defines the minimal contract for any neuron implementation
// This allows the matrix to create and manage neurons without knowing implementation details
type NeuralComponent interface {
	// === CORE IDENTITY ===
	ID() string
	Position() types.Position3D
	Type() types.ComponentType

	// === BIOLOGICAL INTERFACES ===
	SignalListener // Receives electrical signals (gap junctions, etc.)
	BindingTarget  // Receives chemical signals (neurotransmitters)

	// === LIFECYCLE MANAGEMENT ===
	Start() error // Begin neural processing
	Stop() error  // Gracefully shutdown
	IsActive() bool

	// === NEURAL SPECIFIC OPERATIONS ===
	// These methods allow the matrix to configure and monitor neurons
	GetThreshold() float64
	SetThreshold(threshold float64)
	GetActivityLevel() float64 // For health monitoring by microglia
	GetConnectionCount() int   // For connectivity analysis
}

// SynapseInterface defines the contract for synaptic connections
// Enables matrix-managed synapse creation with callback injection
// type SynapseInterface interface {
// 	// === CORE IDENTITY ===
// 	ID() string
// 	Position() types.Position3D
// 	Type() types.ComponentType

// 	// === SYNAPTIC CONNECTIVITY ===
// 	GetPresynapticID() string
// 	GetPostsynapticID() string
// 	GetWeight() float64
// 	SetWeight(weight float64)

// 	// === TRANSMISSION MECHANICS ===
// 	// The matrix injects these callbacks during creation
// 	Transmit(message float64)

// 	// === PLASTICITY ===
// 	UpdateWeight(event types.PlasticityEvent)
// 	GetPlasticityConfig() types.PlasticityConfig
// 	ApplyPlasticity(adjustment types.PlasticityAdjustment)
// 	ShouldPrune() bool

// 	// === LIFECYCLE ===
// 	IsActive() bool
// 	GetActivityInfo() types.ActivityInfo
// }

// =================================================================================
// FACTORY CALLBACK DEFINITIONS
// =================================================================================

// NeuronCallbacks contains function callbacks injected by the matrix during neuron creation
// This implements the Inversion of Control pattern where neurons don't know about the matrix
type NeuronCallbacks = component.NeuronCallbacks

// SynapseCallbacks contains function callbacks injected by the matrix during synapse creation
type SynapseCallbacks struct {
	// Message delivery callback (replaces direct neuron access)
	DeliverMessage func(targetID string, message types.NeuralSignal) error

	// Spatial delay calculation
	GetTransmissionDelay func() time.Duration

	// Chemical release for neurotransmitter signaling
	ReleaseNeurotransmitter func(ligandType types.LigandType, concentration float64) error

	// Activity reporting for monitoring
	ReportActivity func(activity types.SynapticActivity)

	// Plasticity event reporting
	ReportPlasticityEvent func(event types.PlasticityEvent)
}

// =================================================================================
// FACTORY FUNCTION TYPE DEFINITIONS
// =================================================================================

// NeuronFactoryFunc defines the signature for neuron creation functions
// These functions are registered with the matrix for different neuron types
type NeuronFactoryFunc func(id string, config types.NeuronConfig, callbacks NeuronCallbacks) (component.NeuralComponent, error)

// SynapseFactoryFunc defines the signature for synapse creation functions
// These functions are registered with the matrix for different synapse types
type SynapseFactoryFunc func(id string, config types.SynapseConfig, callbacks SynapseCallbacks) (component.SynapticProcessor, error)

// =================================================================================
// OTHER INTERFACES AND TYPES
// =================================================================================

// NeuralComponentType identifies biological component categories
type NeuralComponentType string

const (
	NeuronType          NeuralComponentType = "neuron"
	SynapseType         NeuralComponentType = "synapse"
	AstrocyteType       NeuralComponentType = "astrocyte"
	MicrogliaType       NeuralComponentType = "microglia"
	OligodendrocyteType NeuralComponentType = "oligodendrocyte"
)

// ComponentType categorizes components (use component package types)
type ComponentType = types.ComponentType
type ComponentState = types.ComponentState

// Re-export component constants for convenience
const (
	ComponentNeuron  = types.TypeNeuron
	ComponentSynapse = types.TypeSynapse
	ComponentGate    = types.TypeGlialCell     // Map to glial cell type
	ComponentPlugin  = types.TypeMicrogliaCell // Map to microglia type
)

const (
	StateActive       = types.StateActive
	StateInactive     = types.StateInactive
	StateShuttingDown = types.StateShuttingDown
)

// BindingTarget receives chemical signals (like having receptors)
type BindingTarget interface {
	Bind(ligandType LigandType, sourceID string, concentration float64)
	GetReceptors() []LigandType
	Position() Position3D
}

// SignalListener defines the interface for components that receive discrete signals
type SignalListener interface {
	ID() string
	OnSignal(signalType SignalType, sourceID string, data interface{})
}

// ComponentInfo holds basic component information (use component package type)
type ComponentInfo = component.ComponentInfo

// ComponentCriteria for searching
type ComponentCriteria struct {
	Type     *ComponentType
	State    *ComponentState
	Position *Position3D
	Radius   float64
}
