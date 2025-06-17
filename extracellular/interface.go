package extracellular

import "time"

// =================================================================================
// FACTORY PATTERN INTERFACES
// =================================================================================

// NeuronInterface defines the minimal contract for any neuron implementation
// This allows the matrix to create and manage neurons without knowing implementation details
type NeuronInterface interface {
	// === CORE IDENTITY ===
	ID() string
	Position() Position3D
	ComponentType() ComponentType

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
type SynapseInterface interface {
	// === CORE IDENTITY ===
	ID() string
	Position() Position3D
	ComponentType() ComponentType

	// === SYNAPTIC CONNECTIVITY ===
	GetPresynapticID() string
	GetPostsynapticID() string
	GetWeight() float64
	SetWeight(weight float64)

	// === TRANSMISSION MECHANICS ===
	// The matrix injects these callbacks during creation
	Transmit(message SynapseMessage) error

	// === PLASTICITY ===
	UpdateWeight(event PlasticityEvent)
	GetPlasticityConfig() PlasticityConfig

	// === LIFECYCLE ===
	IsActive() bool
	GetLastActivity() time.Time
}

// =================================================================================
// FACTORY CONFIGURATION STRUCTURES
// =================================================================================

// NeuronConfig defines parameters for neuron creation
type NeuronConfig struct {
	// Biological parameters
	Threshold        float64       `json:"threshold"`
	DecayRate        float64       `json:"decay_rate"`
	RefractoryPeriod time.Duration `json:"refractory_period"`

	// Spatial positioning
	Position Position3D `json:"position"`

	// Chemical receptors this neuron expresses
	Receptors []LigandType `json:"receptors"`

	// Electrical signal types this neuron responds to
	SignalTypes []SignalType `json:"signal_types"`

	// Type-specific configuration
	NeuronType string                 `json:"neuron_type"` // "pyramidal_l5", "fast_spiking_interneuron", etc.
	Metadata   map[string]interface{} `json:"metadata"`
}

// SynapseConfig defines parameters for synapse creation
type SynapseConfig struct {
	// Connection topology
	PresynapticID  string `json:"presynaptic_id"`
	PostsynapticID string `json:"postsynaptic_id"`

	// Synaptic properties
	InitialWeight float64       `json:"initial_weight"`
	Delay         time.Duration `json:"delay"`

	// Neurotransmitter type for this synapse
	LigandType LigandType `json:"ligand_type"`

	// Plasticity configuration
	PlasticityEnabled bool             `json:"plasticity_enabled"`
	PlasticityConfig  PlasticityConfig `json:"plasticity_config"`

	// Spatial positioning
	Position Position3D `json:"position"`

	// Type-specific configuration
	SynapseType string                 `json:"synapse_type"` // "excitatory_plastic", "inhibitory_static", etc.
	Metadata    map[string]interface{} `json:"metadata"`
}

// PlasticityConfig defines synaptic plasticity parameters
type PlasticityConfig struct {
	LearningRate   float64       `json:"learning_rate"`
	STDPWindow     time.Duration `json:"stdp_window"`
	MaxWeight      float64       `json:"max_weight"`
	MinWeight      float64       `json:"min_weight"`
	DecayRate      float64       `json:"decay_rate"`
	PlasticityType string        `json:"plasticity_type"` // "stdp", "bcm", "oja", etc.
}

// PlasticityEvent represents synaptic plasticity trigger events
type PlasticityEvent struct {
	EventType PlasticityEventType `json:"event_type"`
	Timestamp time.Time           `json:"timestamp"`
	PreTime   time.Time           `json:"pre_time"`
	PostTime  time.Time           `json:"post_time"`
	Strength  float64             `json:"strength"`
	SourceID  string              `json:"source_id"`
}

type PlasticityEventType int

const (
	PlasticitySTDP PlasticityEventType = iota
	PlasticityBCM
	PlasticityOja
	PlasticityHomeostatic
)

// SynapseMessage represents a message transmitted through a synapse
type SynapseMessage struct {
	Value     float64    `json:"value"`
	Timestamp time.Time  `json:"timestamp"`
	SourceID  string     `json:"source_id"`
	SynapseID string     `json:"synapse_id"`
	Ligand    LigandType `json:"ligand"` // NEW: Chemical payload
}

// =================================================================================
// FACTORY CALLBACK DEFINITIONS
// =================================================================================

// NeuronCallbacks contains function callbacks injected by the matrix during neuron creation
// This implements the Inversion of Control pattern where neurons don't know about the matrix
type NeuronCallbacks struct {
	// Chemical signaling callback
	ReleaseChemical func(ligandType LigandType, concentration float64) error

	// Electrical signaling callback
	SendElectricalSignal func(signalType SignalType, data interface{})

	// Spatial queries callback
	GetSpatialDelay      func(targetID string) time.Duration
	FindNearbyComponents func(radius float64) []ComponentInfo

	// Health reporting callback (for microglia monitoring)
	ReportHealth func(activityLevel float64, connectionCount int)

	// Lifecycle events callback
	ReportStateChange func(oldState, newState ComponentState)
}

// SynapseCallbacks contains function callbacks injected by the matrix during synapse creation
type SynapseCallbacks struct {
	// Message delivery callback (replaces direct neuron access)
	DeliverMessage func(targetID string, message SynapseMessage) error

	// Spatial delay calculation
	GetTransmissionDelay func() time.Duration

	// Chemical release for neurotransmitter signaling
	ReleaseNeurotransmitter func(ligandType LigandType, concentration float64) error

	// Activity reporting for monitoring
	ReportActivity func(activity SynapticActivity)

	// Plasticity event reporting
	ReportPlasticityEvent func(event PlasticityEvent)
}

// SynapticActivity represents synapse activity information
type SynapticActivity struct {
	SynapseID     string    `json:"synapse_id"`
	Timestamp     time.Time `json:"timestamp"`
	MessageValue  float64   `json:"message_value"`
	CurrentWeight float64   `json:"current_weight"`
	ActivityType  string    `json:"activity_type"` // "transmission", "plasticity", etc.
}

// =================================================================================
// FACTORY FUNCTION TYPE DEFINITIONS
// =================================================================================

// NeuronFactoryFunc defines the signature for neuron creation functions
// These functions are registered with the matrix for different neuron types
type NeuronFactoryFunc func(id string, config NeuronConfig, callbacks NeuronCallbacks) (NeuronInterface, error)

// SynapseFactoryFunc defines the signature for synapse creation functions
// These functions are registered with the matrix for different synapse types
type SynapseFactoryFunc func(id string, config SynapseConfig, callbacks SynapseCallbacks) (SynapseInterface, error)

// =================================================================================
// ENHANCED COMPONENT INTERFACE (Updated)
// =================================================================================

// NeuralComponent represents any component in the neural network
// This is the base interface that all neural components must implement
type NeuralComponent interface {
	ID() string
	Position() Position3D
	ComponentType() ComponentType
	IsActive() bool
	GetMetadata() map[string]interface{}
}

// =================================================================================
// Other INTERFACES
// =================================================================================

// Position3D represents 3D coordinates in neural space
type Position3D struct {
	X, Y, Z float64 // Micrometers in biological space
}

// NeuralComponentType identifies biological component categories
type NeuralComponentType string

const (
	NeuronType          NeuralComponentType = "neuron"
	SynapseType         NeuralComponentType = "synapse"
	AstrocyteType       NeuralComponentType = "astrocyte"
	MicrogliaType       NeuralComponentType = "microglia"
	OligodendrocyteType NeuralComponentType = "oligodendrocyte"
)

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

// SignalListener defines the interface for components that receive discrete signals
type SignalListener interface {
	ID() string
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

// ComponentCriteria for searching
type ComponentCriteria struct {
	Type     *ComponentType
	State    *ComponentState
	Position *Position3D
	Radius   float64
}
