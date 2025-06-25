// component/callbacks.go - ENHANCED TO INCLUDE ALL NEURON CALLBACK METHODS
package component

import (
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

// =================================================================================
// CALLBACK INTERFACES - ENHANCED TO INCLUDE ALL REQUIRED METHODS
// =================================================================================

// SynapseCallbacks defines the interface for synapse-to-matrix communication
type SynapseCallbacks interface {
	DeliverMessage(targetID string, msg types.NeuralSignal) error
	GetTransmissionDelay() time.Duration
	ReleaseNeurotransmitter(ligandType types.LigandType, concentration float64) error
	ReportActivity(info types.ActivityInfo)
	ReportPlasticity(event types.PlasticityEvent)
}

// NeuronCallbacks defines the COMPLETE interface for neuron-to-matrix communication
// This includes both basic and enhanced methods that neurons need for full functionality
type NeuronCallbacks interface {
	// === BASIC SYNAPSE MANAGEMENT ===
	CreateSynapse(config types.SynapseCreationConfig) (string, error)
	DeleteSynapse(synapseID string) error
	ListSynapses(criteria types.SynapseCriteria) []types.SynapseInfo

	// === CHEMICAL AND ELECTRICAL SIGNALING ===
	ReleaseChemical(ligandType types.LigandType, concentration float64) error
	SendElectricalSignal(signalType types.SignalType, data interface{})

	// === HEALTH AND SPATIAL SERVICES ===
	ReportHealth(activityLevel float64, connectionCount int)
	GetSpatialDelay(targetID string) time.Duration

	// === ENHANCED PLASTICITY OPERATIONS ===
	// These methods are required for STDP, homeostatic scaling, and pruning
	ApplyPlasticity(synapseID string, adjustment types.PlasticityAdjustment) error
	GetSynapseWeight(synapseID string) (float64, error)
	SetSynapseWeight(synapseID string, weight float64) error

	// === SYNAPSE ACCESS AND DISCOVERY ===
	// Required for advanced synaptic operations and network analysis
	GetSynapse(synapseID string) (SynapticProcessor, error)

	// === MATRIX AND SPATIAL SERVICES ===
	// Required for spatial awareness and extracellular matrix interactions
	GetMatrix() ExtracellularMatrix
	FindNearbyComponents(radius float64) []ComponentInfo
	ReportStateChange(oldState, newState types.ComponentState)
}

// =================================================================================
// SUPPORTING INTERFACES FOR ENHANCED CALLBACKS
// =================================================================================

// ExtracellularMatrix defines the interface for matrix interactions
type ExtracellularMatrix interface {
	// SynapticDelay calculates enhanced delay based on spatial properties
	SynapticDelay(preNeuronID, postNeuronID, synapseID string, baseDelay time.Duration) time.Duration

	// Additional matrix methods can be added here as needed
	// GetChemicalConcentration(ligandType types.LigandType, position types.Position3D) float64
	// PropagateChemicalSignal(signal types.ChemicalSignal)
}

// =================================================================================
// COMPONENT INTERFACES - UPDATED TO USE TYPES PACKAGE
// =================================================================================

// NeuralComponent interface for neurons that need callback setup
type NeuralComponent interface {
	Component // Base component interface
	MessageReceiver
	MessageScheduler
	SetCallbacks(callbacks NeuronCallbacks)
	AddOutputCallback(synapseID string, callback types.OutputCallback)
}

// SynapticComponent interface for synapses that need callback setup
type SynapticComponent interface {
	Component // Base component interface
	SetCallbacks(callbacks SynapseCallbacks)
	GetPlasticityConfig() types.PlasticityConfig
	GetActivityInfo() types.ActivityInfo
}
