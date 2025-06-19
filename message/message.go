package message

import (
	"time"
)

// === NEUROTRANSMITTER TYPES ===
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

// === ELECTRICAL SIGNAL TYPES ===
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

// === PLASTICITY TYPES ===
type PlasticityType string

const (
	PlasticitySTDP        PlasticityType = "stdp"        // Spike-timing dependent plasticity
	PlasticityBCM         PlasticityType = "bcm"         // Bienenstock-Cooper-Munro rule
	PlasticityOja         PlasticityType = "oja"         // Oja's learning rule
	PlasticityHomeostatic PlasticityType = "homeostatic" // Homeostatic scaling
	PlasticityHebian      PlasticityType = "hebbian"     // Classic Hebbian learning
	PlasticityStatic      PlasticityType = "static"      // No plasticity
	PlasticityMetaplastic PlasticityType = "metaplastic" // Plasticity of plasticity
)

// === COMPONENT STATES ===
type ComponentState int

const (
	StateActive       ComponentState = iota // Normal operational state
	StateInactive                           // Temporarily disabled
	StateShuttingDown                       // Graceful shutdown in progress
	StateDeveloping                         // Growing/maturing (developmental)
	StateDying                              // Programmed cell death/apoptosis
	StateDamaged                            // Damaged but potentially recoverable
	StateMaintenance                        // Undergoing maintenance/repair
	StateHibernating                        // Low-activity conservation state
)

func (cs ComponentState) String() string {
	switch cs {
	case StateActive:
		return "Active"
	case StateInactive:
		return "Inactive"
	case StateShuttingDown:
		return "ShuttingDown"
	case StateDeveloping:
		return "Developing"
	case StateDying:
		return "Dying"
	case StateDamaged:
		return "Damaged"
	case StateMaintenance:
		return "Maintenance"
	case StateHibernating:
		return "Hibernating"
	default:
		return "Unknown"
	}
}

// === COMPONENT TYPES ===
type ComponentType int

const (
	ComponentNeuron        ComponentType = iota // Excitable neural cell
	ComponentSynapse                            // Synaptic connection
	ComponentGlialCell                          // Support cell (astrocyte, oligodendrocyte, etc.)
	ComponentBloodVessel                        // Vascular component
	ComponentMicrogliaCell                      // Immune cell of the brain
	ComponentEpendymalCell                      // CSF-brain barrier cell
	ComponentAxon                               // Axonal projection
	ComponentDendrite                           // Dendritic branch
)

func (ct ComponentType) String() string {
	switch ct {
	case ComponentNeuron:
		return "Neuron"
	case ComponentSynapse:
		return "Synapse"
	case ComponentGlialCell:
		return "GlialCell"
	case ComponentBloodVessel:
		return "BloodVessel"
	case ComponentMicrogliaCell:
		return "MicrogliaCell"
	case ComponentEpendymalCell:
		return "EpendymalCell"
	case ComponentAxon:
		return "Axon"
	case ComponentDendrite:
		return "Dendrite"
	default:
		return "Unknown"
	}
}

// === TRANSMISSION MODES ===
type TransmissionMode int

const (
	TransmissionChemical   TransmissionMode = iota // Chemical synaptic transmission
	TransmissionElectrical                         // Electrical/gap junction transmission
	TransmissionVolumetric                         // Volume transmission (diffusion)
	TransmissionRetrograde                         // Retrograde signaling
	TransmissionAntidromic                         // Antidromic propagation
	TransmissionEphaptic                           // Ephaptic coupling
)

func (tm TransmissionMode) String() string {
	switch tm {
	case TransmissionChemical:
		return "Chemical"
	case TransmissionElectrical:
		return "Electrical"
	case TransmissionVolumetric:
		return "Volumetric"
	case TransmissionRetrograde:
		return "Retrograde"
	case TransmissionAntidromic:
		return "Antidromic"
	case TransmissionEphaptic:
		return "Ephaptic"
	default:
		return "Unknown"
	}
}

// === SIGNAL RELIABILITY ===
type SignalReliability int

const (
	ReliabilityHigh    SignalReliability = iota // >95% transmission success
	ReliabilityMedium                           // 70-95% transmission success
	ReliabilityLow                              // 30-70% transmission success
	ReliabilityFailing                          // <30% transmission success
)

func (sr SignalReliability) String() string {
	switch sr {
	case ReliabilityHigh:
		return "High"
	case ReliabilityMedium:
		return "Medium"
	case ReliabilityLow:
		return "Low"
	case ReliabilityFailing:
		return "Failing"
	default:
		return "Unknown"
	}
}

// === CORE MESSAGE TYPE ===
// NeuralSignal represents comprehensive neural communication between components
// This is the fundamental unit of neural communication in the simulation
type NeuralSignal struct {
	// === CORE SIGNAL PROPERTIES ===
	Value         float64 `json:"value"`          // Final weighted signal value reaching postsynaptic neuron
	OriginalValue float64 `json:"original_value"` // Pre-synaptic signal strength (before synaptic weighting)

	// === SYNAPTIC TRANSMISSION PROPERTIES ===
	EffectiveWeight   float64       `json:"effective_weight"`   // Synaptic weight applied to signal
	SynapticDelay     time.Duration `json:"synaptic_delay"`     // Base synaptic processing delay
	TransmissionDelay time.Duration `json:"transmission_delay"` // Total delay including spatial
	SpatialDelay      time.Duration `json:"spatial_delay"`      // Axonal conduction + diffusion delay

	// === TIMING INFORMATION ===
	Timestamp            time.Time `json:"timestamp"`              // When signal was initiated
	PropagationStartTime time.Time `json:"propagation_start_time"` // When signal began propagating
	ArrivalTime          time.Time `json:"arrival_time"`           // Expected arrival time at target

	// === COMPONENT IDENTIFICATION ===
	SourceID  string `json:"source_id"`  // Originating neuron ID
	TargetID  string `json:"target_id"`  // Destination neuron ID
	SynapseID string `json:"synapse_id"` // Processing synapse ID

	// === CHEMICAL TRANSMISSION PROPERTIES ===
	NeurotransmitterType LigandType `json:"neurotransmitter_type"` // Released chemical messenger
	VesicleReleased      bool       `json:"vesicle_released"`      // Whether vesicle was consumed
	ReleaseQuantum       float64    `json:"release_quantum"`       // Number of vesicles released
	ReceptorType         string     `json:"receptor_type"`         // Target receptor subtype

	// === BIOLOGICAL STATE INFORMATION ===
	CalciumLevel          float64 `json:"calcium_level"`          // Presynaptic calcium concentration
	PostsynapticPotential float64 `json:"postsynaptic_potential"` // Expected PSP amplitude

	// === TRANSMISSION CHARACTERISTICS ===
	TransmissionMode   TransmissionMode  `json:"transmission_mode"`     // How signal is transmitted
	SignalReliability  SignalReliability `json:"signal_reliability"`    // Expected transmission reliability
	NoiseLevel         float64           `json:"noise_level"`           // Background noise amplitude
	SignalToNoiseRatio float64           `json:"signal_to_noise_ratio"` // Signal quality metric

	// === PLASTICITY AND LEARNING CONTEXT ===
	PlasticityContext map[string]interface{} `json:"plasticity_context"` // Context for learning rules
	LearningPhase     string                 `json:"learning_phase"`     // Current learning state

	// === FAILURE AND ERROR HANDLING ===
	TransmissionSuccess bool   `json:"transmission_success"` // Whether transmission succeeded
	FailureReason       string `json:"failure_reason"`       // Reason for failure (if any)
	RetryCount          int    `json:"retry_count"`          // Number of retry attempts

	// === METADATA AND EXTENSIONS ===
	Metadata map[string]interface{} `json:"metadata"` // Additional custom data
	Tags     []string               `json:"tags"`     // Classification tags
}
