package message

import (
	"time"
)

// SynapseMessage represents a signal transmitted across a synapse from a presynaptic to a postsynaptic neuron.
// This is the fundamental unit of communication between neurons in the simulation.
//
// BIOLOGICAL CONTEXT:
// In biology, a presynaptic neuron releases neurotransmitters into the synaptic cleft,
// which then bind to receptors on the postsynaptic neuron, causing a change in its
// membrane potential (a postsynaptic potential, PSP). This SynapseMessage struct models
// this PSP, carrying information about the signal's strength, origin, and timing.
type SynapseMessage struct {
	// === SIGNAL PROPERTIES ===
	Value           float64 `json:"value"`            // Processed signal strength (post-weight scaling)
	OriginalValue   float64 `json:"original_value"`   // Original input signal strength
	EffectiveWeight float64 `json:"effective_weight"` // Weight applied during transmission

	// === TIMING INFORMATION ===
	Timestamp         time.Time     `json:"timestamp"`          // When signal was generated
	TransmissionDelay time.Duration `json:"transmission_delay"` // Total delay applied
	SynapticDelay     time.Duration `json:"synaptic_delay"`     // Synaptic processing component
	SpatialDelay      time.Duration `json:"spatial_delay"`      // Spatial propagation component

	// === IDENTIFICATION ===
	SourceID  string `json:"source_id"`  // Pre-synaptic neuron ID
	TargetID  string `json:"target_id"`  // Post-synaptic neuron ID
	SynapseID string `json:"synapse_id"` // Transmitting synapse ID

	// === BIOLOGICAL METADATA ===
	NeurotransmitterType LigandType `json:"neurotransmitter_type"` // Type of neurotransmitter released
	VesicleReleased      bool       `json:"vesicle_released"`      // Whether vesicle release occurred
	CalciumLevel         float64    `json:"calcium_level"`         // Calcium level during transmission

	// === PLASTICITY CONTEXT ===
	PreSynapticSpike bool   `json:"presynaptic_spike"` // Whether this represents a pre-synaptic spike
	PlasticityWindow bool   `json:"plasticity_window"` // Whether this is within STDP window
	LearningContext  string `json:"learning_context"`  // Context for learning ("reward", "punishment", etc.)
}

// LigandType represents different types of neurotransmitters or ligands that can exist in the extracellular space.
// It uses an integer type for efficient storage and comparison.
type LigandType int

// Pre-defined constants for common neurotransmitter types.
const (
	LigandNone          LigandType = iota // Default/unspecified ligand type.
	LigandGlutamate                       // Excitatory neurotransmitter.
	LigandGABA                            // Inhibitory neurotransmitter.
	LigandDopamine                        // Modulatory neurotransmitter.
	LigandSerotonin                       // Modulatory neurotransmitter.
	LigandAcetylcholine                   // Excitatory or inhibitory, depending on receptor.
)

// String returns the string representation of a LigandType.
// This is useful for logging, debugging, and displaying information.
func (lt LigandType) String() string {
	switch lt {
	case LigandGlutamate:
		return "Glutamate"
	case LigandGABA:
		return "GABA"
	case LigandDopamine:
		return "Dopamine"
	case LigandSerotonin:
		return "Serotonin"
	case LigandAcetylcholine:
		return "Acetylcholine"
	case LigandNone:
		return "None"
	default:
		return "Unknown"
	}
}

// SignalType represents different types of signals or events that can be propagated within the matrix.
// This allows for a flexible communication system beyond just neurotransmitter release.
type SignalType int

// Pre-defined constants for common signal types.
const (
	SignalNone             SignalType = iota // Default/unspecified signal type.
	SignalFired                              // Indicates a neuron has fired an action potential.
	SignalPlasticityEvent                    // Indicates a synaptic plasticity event has occurred (e.g., STDP, scaling).
	SignalChemicalGradient                   // Represents a change in a chemical gradient in the extracellular space.
	SignalStructuralChange                   // Indicates a structural change (e.g., synapse growth/pruning, neuron death).
)

// String returns the string representation of a SignalType.
func (st SignalType) String() string {
	switch st {
	case SignalFired:
		return "Fired"
	case SignalPlasticityEvent:
		return "PlasticityEvent"
	case SignalChemicalGradient:
		return "ChemicalGradient"
	case SignalStructuralChange:
		return "StructuralChange"
	case SignalNone:
		return "None"
	default:
		return "Unknown"
	}
}
