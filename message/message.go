package message

import (
	"time"
)

// ============================================================================
// NEUROTRANSMITTER AND CHEMICAL SIGNALING TYPES
// ============================================================================

// LigandType represents different types of neurotransmitters and signaling molecules
// These are the actual chemical messengers that carry information between neurons
type LigandType int

const (
	LigandNone            LigandType = iota // Default/unspecified ligand type
	LigandGlutamate                         // Primary excitatory neurotransmitter
	LigandGABA                              // Primary inhibitory neurotransmitter
	LigandDopamine                          // Modulatory neurotransmitter (reward, motor control)
	LigandSerotonin                         // Modulatory neurotransmitter (mood, arousal)
	LigandAcetylcholine                     // Excitatory or inhibitory depending on receptor
	LigandNorepinephrine                    // Modulatory neurotransmitter (attention, arousal)
	LigandHistamine                         // Modulatory neurotransmitter (arousal, inflammation)
	LigandGlycine                           // Inhibitory neurotransmitter (spinal cord)
	LigandAdenosine                         // Neuromodulator (sleep, neuroprotection)
	LigandNitricOxide                       // Gaseous signaling molecule (retrograde signaling)
	LigandEndocannabinoid                   // Lipid-based retrograde messenger
	LigandNeuropeptideY                     // Peptide neurotransmitter (feeding, anxiety)
	LigandSubstanceP                        // Peptide neurotransmitter (pain, inflammation)
	LigandVasopressin                       // Peptide hormone/neurotransmitter
	LigandOxytocin                          // Peptide hormone/neurotransmitter (social bonding)
)

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
	case LigandNorepinephrine:
		return "Norepinephrine"
	case LigandHistamine:
		return "Histamine"
	case LigandGlycine:
		return "Glycine"
	case LigandAdenosine:
		return "Adenosine"
	case LigandNitricOxide:
		return "NitricOxide"
	case LigandEndocannabinoid:
		return "Endocannabinoid"
	case LigandNeuropeptideY:
		return "NeuropeptideY"
	case LigandSubstanceP:
		return "SubstanceP"
	case LigandVasopressin:
		return "Vasopressin"
	case LigandOxytocin:
		return "Oxytocin"
	case LigandNone:
		return "None"
	default:
		return "Unknown"
	}
}

// GetPolarityEffect returns the typical effect of this ligand (excitatory: +1, inhibitory: -1, modulatory: 0)
func (lt LigandType) GetPolarityEffect() float64 {
	switch lt {
	case LigandGlutamate:
		return 1.0 // Excitatory
	case LigandGABA, LigandGlycine:
		return -1.0 // Inhibitory
	case LigandAcetylcholine:
		return 1.0 // Usually excitatory in CNS
	case LigandDopamine, LigandSerotonin, LigandNorepinephrine, LigandHistamine:
		return 0.0 // Modulatory (depends on receptor subtype)
	case LigandAdenosine:
		return -0.5 // Generally inhibitory/protective
	case LigandNitricOxide, LigandEndocannabinoid:
		return 0.0 // Complex modulatory effects
	case LigandNeuropeptideY, LigandSubstanceP, LigandVasopressin, LigandOxytocin:
		return 0.0 // Peptide modulators
	default:
		return 0.0
	}
}

// ============================================================================
// ELECTRICAL SIGNAL TYPES
// ============================================================================

// SignalType represents different types of electrical or coordination signals
// These are for gap junction communication and network-wide coordination
type SignalType int

const (
	SignalNone               SignalType = iota // Default/unspecified signal type
	SignalFired                                // Neuron has fired an action potential
	SignalConnected                            // New synaptic connection established
	SignalDisconnected                         // Synaptic connection removed
	SignalThresholdChanged                     // Neuron firing threshold adjusted
	SignalSynchronization                      // Network synchronization pulse
	SignalCalciumWave                          // Calcium wave propagation (glial)
	SignalPlasticityEvent                      // Synaptic plasticity occurred
	SignalChemicalGradient                     // Chemical concentration gradient change
	SignalStructuralChange                     // Physical structure modification
	SignalMetabolicState                       // Energy/metabolic state change
	SignalHealthWarning                        // Component health issue detected
	SignalNetworkOscillation                   // Network-wide oscillatory activity
)

func (st SignalType) String() string {
	switch st {
	case SignalFired:
		return "Fired"
	case SignalConnected:
		return "Connected"
	case SignalDisconnected:
		return "Disconnected"
	case SignalThresholdChanged:
		return "ThresholdChanged"
	case SignalSynchronization:
		return "Synchronization"
	case SignalCalciumWave:
		return "CalciumWave"
	case SignalPlasticityEvent:
		return "PlasticityEvent"
	case SignalChemicalGradient:
		return "ChemicalGradient"
	case SignalStructuralChange:
		return "StructuralChange"
	case SignalMetabolicState:
		return "MetabolicState"
	case SignalHealthWarning:
		return "HealthWarning"
	case SignalNetworkOscillation:
		return "NetworkOscillation"
	case SignalNone:
		return "None"
	default:
		return "Unknown"
	}
}

// ============================================================================
// CORE NEURAL SIGNAL TYPE - PURE SIGNAL CONTENT
// ============================================================================

// NeuralSignal represents the content of neural communication between components
// This contains ONLY the signal data, timing, and biological properties
// NO architectural or configuration information
type NeuralSignal struct {
	// === CORE SIGNAL PROPERTIES ===
	Value         float64 `json:"value"`          // Final signal strength reaching target
	OriginalValue float64 `json:"original_value"` // Pre-processing signal strength

	// === TIMING INFORMATION ===
	Timestamp     time.Time     `json:"timestamp"`      // When signal was initiated
	SynapticDelay time.Duration `json:"synaptic_delay"` // Synaptic processing delay
	SpatialDelay  time.Duration `json:"spatial_delay"`  // Axonal conduction delay
	TotalDelay    time.Duration `json:"total_delay"`    // Combined transmission delay

	// === ROUTING INFORMATION ===
	SourceID  string `json:"source_id"`  // Originating component ID
	TargetID  string `json:"target_id"`  // Destination component ID
	SynapseID string `json:"synapse_id"` // Processing synapse ID (if applicable)

	// === CHEMICAL SIGNAL CONTENT ===
	NeurotransmitterType LigandType `json:"neurotransmitter_type"` // Chemical messenger type
	VesicleReleased      bool       `json:"vesicle_released"`      // Whether vesicle was consumed
	CalciumLevel         float64    `json:"calcium_level"`         // Presynaptic calcium level

	// === SIGNAL QUALITY AND RELIABILITY ===
	TransmissionSuccess bool    `json:"transmission_success"` // Whether transmission succeeded
	FailureReason       string  `json:"failure_reason"`       // Reason for failure (if any)
	NoiseLevel          float64 `json:"noise_level"`          // Background noise amplitude

	// === METADATA ===
	Metadata map[string]interface{} `json:"metadata"` // Additional signal-specific data
}
