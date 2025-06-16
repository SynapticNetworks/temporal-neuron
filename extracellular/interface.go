package extracellular

import "time"

// NeuralComponent represents any component in the neural network
// (biological: neurons, synapses, glial cells are all neural tissue components)
type NeuralComponent interface {
	ID() string
	Position() Position3D
	ComponentType() NeuralComponentType
}

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

// NeuralComponentInfo holds registration data for any neural component
type NeuralComponentInfo struct {
	ID             string              `json:"id"`
	Type           NeuralComponentType `json:"type"`
	Position       Position3D          `json:"position"`
	RegisteredAt   time.Time           `json:"registered_at"`
	LastActivity   time.Time           `json:"last_activity"`
	MetabolicState MetabolicState      `json:"metabolic_state"`
}

// MetabolicState represents energy and chemical state
type MetabolicState struct {
	EnergyLevel  float64 `json:"energy_level"`  // ATP availability (0.0-1.0)
	CalciumLevel float64 `json:"calcium_level"` // Intracellular calcium
	Active       bool    `json:"active"`        // Currently processing
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

// SignalListener defines the interface for any component that can receive
// discrete signals from the SignalMediator.
type SignalListener interface {
	// ID returns the unique identifier of the listener component.
	// This is crucial for the SignalMediator to prevent a component
	// from receiving its own broadcasted signals.
	ID() string

	// OnSignal is the callback method invoked when a subscribed signal is received.
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
