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
	Value float64 // The numerical strength or amplitude of the signal.
	// In biological terms, this relates to the magnitude of the postsynaptic potential (PSP).
	// Positive values typically represent excitatory signals (EPSPs), negative for inhibitory (IPSPs).

	OriginalValue float64 // The original, unmodified signal value from the presynaptic neuron.
	// Useful for debugging and understanding the transformation across the synapse.

	EffectiveWeight float64 // The weight of the synapse that transmitted this message.
	// This is the synaptic efficacy (strength) applied to the original signal.

	Timestamp time.Time // The precise time at which the signal arrived at the postsynaptic neuron.
	// Critical for spike-timing dependent plasticity (STDP) and temporal summation.

	TransmissionDelay time.Duration // The total time taken for the signal to travel from presynaptic axon hillock to postsynaptic dendrite.
	// This includes both the intrinsic synaptic delay and any spatial propagation delay across the extracellular matrix.

	SynapticDelay time.Duration // The intrinsic delay within the synapse itself (e.g., neurotransmitter diffusion, receptor binding).
	// This is part of the total TransmissionDelay.

	SpatialDelay time.Duration // The delay caused by the physical distance and propagation speed in the extracellular matrix.
	// This is the other part of the total TransmissionDelay.

	SourceID string // The unique identifier of the presynaptic neuron that sent this message.
	// Important for tracking input sources for activity-dependent plasticity and scaling.

	TargetID string // The unique identifier of the postsynaptic neuron intended to receive this message.
	// Ensures messages are routed to the correct destination.

	SynapseID string // The unique identifier of the synapse through which this message was transmitted.
	// Enables synapse-specific tracking and plasticity updates.

	NeurotransmitterType LigandType // The type of neurotransmitter released (e.g., Glutamate, GABA).
	// Influences how the postsynaptic neuron responds (excitatory vs. inhibitory).

	VesicleReleased bool // Indicates if a vesicle was successfully released to transmit this message.
	// Relevant for vesicle pool dynamics and probabilistic release models.

	CalciumLevel float64 // The presynaptic calcium level at the time of transmission.
	// Important for calcium-dependent plasticity rules.
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
