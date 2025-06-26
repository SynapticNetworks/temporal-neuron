// types/signals.go
package types

// =================================================================================
// CHEMICAL SIGNALING TYPES
// =================================================================================

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
	LigandCalcium                           // Intracellular calcium signaling (essential for plasticity)
	LigandBDNF                              // Brain-Derived Neurotrophic Factor
	LigandNGF                               // Nerve Growth Factor
)

func (lt LigandType) String() string {
	switch lt {
	case LigandBDNF:
		return "BDNF"
	case LigandNGF:
		return "NGF"
	case LigandCalcium:
		return "Calcium"
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
	case LigandCalcium:
		return 0.0 // Modulatory signaling cascade
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

// =================================================================================
// ELECTRICAL SIGNALING TYPES
// =================================================================================

// SignalType represents different types of electrical or coordination signals
// These are for gap junction communication and network-wide coordination
type SignalType int

const (
	SignalNone               SignalType = iota // Default/unspecified signal type
	SignalFired                                // Neuron fired an action potential
	SignalConnected                            // New connection established
	SignalDisconnected                         // Connection removed
	SignalThresholdChanged                     // Firing threshold adjustment
	SignalSynchronization                      // Network synchronization pulse
	SignalCalciumWave                          // Calcium wave propagation (astrocytes)
	SignalPlasticityEvent                      // Plasticity-related signaling
	SignalChemicalGradient                     // Chemical concentration gradient
	SignalStructuralChange                     // Network topology change
	SignalMetabolicState                       // Metabolic/energy state change
	SignalHealthWarning                        // Component health alert
	SignalNetworkOscillation                   // Network-wide oscillatory activity
)

// =================================================================================
// SIGNAL POLARITY AND EFFECTS
// =================================================================================

// SignalPolarity represents the effect direction of a signal
type SignalPolarity int

const (
	PolarityNeutral    SignalPolarity = iota // No net effect (modulatory)
	PolarityExcitatory                       // Increases activity/probability
	PolarityInhibitory                       // Decreases activity/probability
)

// SignalStrength represents signal intensity categories
type SignalStrength int

const (
	StrengthWeak    SignalStrength = iota // Low intensity signal
	StrengthMedium                        // Moderate intensity signal
	StrengthStrong                        // High intensity signal
	StrengthMaximal                       // Maximum intensity signal
)
