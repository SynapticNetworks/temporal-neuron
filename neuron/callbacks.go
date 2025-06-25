// neuron/callbacks.go - FIXED TO IMPLEMENT COMPONENT INTERFACE
package neuron

import (
	"fmt"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// === OUTPUT CALLBACK (DIRECT SYNAPSE COMMUNICATION) ===
// type OutputCallback struct {
// 	TransmitMessage func(msg types.NeuralSignal) error
// 	GetWeight       func() float64
// 	GetDelay        func() time.Duration
// 	GetTargetID     func() string
// }

// InputCallback allows synapses to schedule delayed delivery back to neurons
type InputCallback struct {
	ScheduleDelivery func(msg types.NeuralSignal, targetID string, delay time.Duration) error
	GetNeuronID      func() string
}

// === ENHANCED MATRIX SERVICE CALLBACKS (INJECTED COORDINATION) ===
// NeuronCallbacks struct that implements component.NeuronCallbacks interface
type NeuronCallbacks struct {
	// ===  NETWORK-WIDE SIGNALING ===
	ReleaseChemicalFunc      func(ligandType types.LigandType, concentration float64) error
	SendElectricalSignalFunc func(signalType types.SignalType, data interface{})

	// ===  SPATIAL SERVICES ===
	GetSpatialDelayFunc      func(targetID string) time.Duration
	FindNearbyComponentsFunc func(radius float64) []component.ComponentInfo

	// ===  HEALTH & STATE REPORTING ===
	ReportHealthFunc      func(activityLevel float64, connectionCount int)
	ReportStateChangeFunc func(oldState, newState types.ComponentState)

	// ===  BASIC SYNAPSE CREATION ===
	CreateSynapseFunc func(config types.SynapseCreationConfig) (string, error)

	// === ENHANCED SYNAPSE MANAGEMENT ===
	DeleteSynapseFunc func(synapseID string) error

	// === SYNAPSE DISCOVERY & ACCESS ===
	GetSynapseFunc   func(synapseID string) (component.SynapticProcessor, error)
	ListSynapsesFunc func(criteria types.SynapseCriteria) []types.SynapseInfo

	// === PLASTICITY OPERATIONS ===
	ApplyPlasticityFunc  func(synapseID string, adjustment types.PlasticityAdjustment) error
	GetSynapseWeightFunc func(synapseID string) (float64, error)
	SetSynapseWeightFunc func(synapseID string, weight float64) error

	// === MATRIX ACCESS ===
	GetMatrixFunc func() component.ExtracellularMatrix
}

// ============================================================================
// IMPLEMENT component.NeuronCallbacks INTERFACE METHODS
// ============================================================================

// CreateSynapse implements component.NeuronCallbacks interface
func (nc *NeuronCallbacks) CreateSynapse(config types.SynapseCreationConfig) (string, error) {
	if nc.CreateSynapseFunc != nil {
		return nc.CreateSynapseFunc(config)
	}
	return "", fmt.Errorf("CreateSynapse callback not set")
}

// DeleteSynapse implements component.NeuronCallbacks interface
func (nc *NeuronCallbacks) DeleteSynapse(synapseID string) error {
	if nc.DeleteSynapseFunc != nil {
		return nc.DeleteSynapseFunc(synapseID)
	}
	return fmt.Errorf("DeleteSynapse callback not set")
}

// ListSynapses implements component.NeuronCallbacks interface
func (nc *NeuronCallbacks) ListSynapses(criteria types.SynapseCriteria) []types.SynapseInfo {
	if nc.ListSynapsesFunc != nil {
		return nc.ListSynapsesFunc(criteria)
	}
	return []types.SynapseInfo{}
}

// ReleaseChemical implements component.NeuronCallbacks interface
func (nc *NeuronCallbacks) ReleaseChemical(ligandType types.LigandType, concentration float64) error {
	if nc.ReleaseChemicalFunc != nil {
		return nc.ReleaseChemicalFunc(ligandType, concentration)
	}
	return fmt.Errorf("ReleaseChemical callback not set")
}

// SendElectricalSignal implements component.NeuronCallbacks interface
func (nc *NeuronCallbacks) SendElectricalSignal(signalType types.SignalType, data interface{}) {
	if nc.SendElectricalSignalFunc != nil {
		nc.SendElectricalSignalFunc(signalType, data)
	}
}

// ReportHealth implements component.NeuronCallbacks interface
func (nc *NeuronCallbacks) ReportHealth(activityLevel float64, connectionCount int) {
	if nc.ReportHealthFunc != nil {
		nc.ReportHealthFunc(activityLevel, connectionCount)
	}
}

// GetSpatialDelay implements component.NeuronCallbacks interface
func (nc *NeuronCallbacks) GetSpatialDelay(targetID string) time.Duration {
	if nc.GetSpatialDelayFunc != nil {
		return nc.GetSpatialDelayFunc(targetID)
	}
	return 0
}

// ============================================================================
// ADDITIONAL ENHANCED METHODS (NOT IN COMPONENT INTERFACE)
// ============================================================================

// These methods provide extended functionality beyond the basic component interface

// ApplyPlasticity applies plasticity adjustments to synapses
func (nc *NeuronCallbacks) ApplyPlasticity(synapseID string, adjustment types.PlasticityAdjustment) error {
	if nc.ApplyPlasticityFunc != nil {
		return nc.ApplyPlasticityFunc(synapseID, adjustment)
	}
	return fmt.Errorf("ApplyPlasticity callback not set")
}

// GetSynapseWeight retrieves current synapse weight
func (nc *NeuronCallbacks) GetSynapseWeight(synapseID string) (float64, error) {
	if nc.GetSynapseWeightFunc != nil {
		return nc.GetSynapseWeightFunc(synapseID)
	}
	return 0.0, fmt.Errorf("GetSynapseWeight callback not set")
}

// SetSynapseWeight directly sets synapse weight
func (nc *NeuronCallbacks) SetSynapseWeight(synapseID string, weight float64) error {
	if nc.SetSynapseWeightFunc != nil {
		return nc.SetSynapseWeightFunc(synapseID, weight)
	}
	return fmt.Errorf("SetSynapseWeight callback not set")
}

// GetSynapse retrieves synapse processor instance
func (nc *NeuronCallbacks) GetSynapse(synapseID string) (component.SynapticProcessor, error) {
	if nc.GetSynapseFunc != nil {
		return nc.GetSynapseFunc(synapseID)
	}
	return nil, fmt.Errorf("GetSynapse callback not set")
}

// GetMatrix retrieves extracellular matrix instance
func (nc *NeuronCallbacks) GetMatrix() component.ExtracellularMatrix {
	if nc.GetMatrixFunc != nil {
		return nc.GetMatrixFunc()
	}
	return nil
}

// FindNearbyComponents finds components within spatial radius
func (nc *NeuronCallbacks) FindNearbyComponents(radius float64) []component.ComponentInfo {
	if nc.FindNearbyComponentsFunc != nil {
		return nc.FindNearbyComponentsFunc(radius)
	}
	return []component.ComponentInfo{}
}

// ReportStateChange reports component state transitions
func (nc *NeuronCallbacks) ReportStateChange(oldState, newState types.ComponentState) {
	if nc.ReportStateChangeFunc != nil {
		nc.ReportStateChangeFunc(oldState, newState)
	}
}
