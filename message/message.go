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

const (
	// Fast synaptic neurotransmitters
	LigandGlutamate LigandType = iota // Primary excitatory neurotransmitter
	LigandGABA                        // Primary inhibitory neurotransmitter
	LigandGlycine                     // Inhibitory neurotransmitter (spinal cord, brainstem)

	// Slow neuromodulators
	LigandDopamine       // Reward, motivation, motor control
	LigandSerotonin      // Mood, arousal, behavioral state
	LigandNorepinephrine // Attention, arousal, stress response
	LigandAcetylcholine  // Attention, learning, autonomic function

	// Neuropeptides and others
	LigandEndorphin   // Pain modulation, reward
	LigandOxytocin    // Social bonding, trust
	LigandVasopressin // Social behavior, memory
)

// String returns human-readable name for ligand type
func (lt LigandType) String() string {
	switch lt {
	case LigandGlutamate:
		return "Glutamate"
	case LigandGABA:
		return "GABA"
	case LigandGlycine:
		return "Glycine"
	case LigandDopamine:
		return "Dopamine"
	case LigandSerotonin:
		return "Serotonin"
	case LigandNorepinephrine:
		return "Norepinephrine"
	case LigandAcetylcholine:
		return "Acetylcholine"
	case LigandEndorphin:
		return "Endorphin"
	case LigandOxytocin:
		return "Oxytocin"
	case LigandVasopressin:
		return "Vasopressin"
	default:
		return "Unknown"
	}
}

// IsExcitatory returns true for excitatory neurotransmitters
func (lt LigandType) IsExcitatory() bool {
	switch lt {
	case LigandGlutamate, LigandAcetylcholine:
		return true
	default:
		return false
	}
}

// IsInhibitory returns true for inhibitory neurotransmitters
func (lt LigandType) IsInhibitory() bool {
	switch lt {
	case LigandGABA, LigandGlycine:
		return true
	default:
		return false
	}
}

// IsModulatory returns true for neuromodulators
func (lt LigandType) IsModulatory() bool {
	switch lt {
	case LigandDopamine, LigandSerotonin, LigandNorepinephrine,
		LigandEndorphin, LigandOxytocin, LigandVasopressin:
		return true
	default:
		return false
	}
}
