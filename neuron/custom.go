package neuron

import "github.com/SynapticNetworks/temporal-neuron/types"

// ============================================================================
// CUSTOM BEHAVIORS DOCUMENTATION
// ============================================================================

/*
CUSTOM BEHAVIORS SYSTEM - EXTENSIBILITY FOR TESTING AND RESEARCH

The CustomBehaviors system allows extending neuron functionality without
modifying core code. This is particularly useful for:

1. TESTING SCENARIOS
   - Simulating activity-dependent chemical release (BDNF, growth factors)
   - Testing neuron-matrix communication patterns
   - Validating chemical signaling pathways

2. RESEARCH APPLICATIONS
   - Modeling novel neurotransmitter systems
   - Simulating pharmacological interventions
   - Implementing experimental chemical release patterns

3. SPECIALIZED NEURAL MODELS
   - Custom neuromodulator release
   - Activity-dependent gene expression simulation
   - Non-standard neurotransmitter combinations

USAGE EXAMPLES:

// Basic activity-dependent release:
neuron.SetCustomChemicalRelease(func(activityRate, outputValue float64, release func(types.LigandType, float64) error) {
    if activityRate > 5.0 {  // 5 Hz threshold
        release(types.LigandBDNF, activityRate * 0.02)
    }
})

// Multiple chemical release based on different conditions:
neuron.SetCustomChemicalRelease(func(activityRate, outputValue float64, release func(types.LigandType, float64) error) {
    // Activity-dependent BDNF
    if activityRate > 5.0 {
        release(types.LigandBDNF, activityRate * 0.02)
    }

    // Output-dependent dopamine
    if outputValue > 2.0 {
        release(types.LigandDopamine, 0.5)
    }

    // Complex logic
    if activityRate > 10.0 && outputValue > 1.5 {
        release(types.LigandSerotonin, 0.3)
    }
})

// Disable custom behaviors:
neuron.DisableCustomBehaviors()

INTEGRATION WITH CORE SYSTEMS:
- Custom behaviors execute AFTER normal chemical release
- Errors in custom behaviors don't affect normal neuron operation
- Custom release uses the same matrix callbacks as normal release
- Activity rate and output value are calculated from normal firing process

THREAD SAFETY:
- Custom behavior functions are called from the main neuron firing thread
- No additional synchronization needed in custom behavior functions
- Matrix callbacks handle their own thread safety

BIOLOGICAL ACCURACY:
- Custom behaviors should respect biological timing constraints
- Chemical concentrations should use realistic ranges (0.001-10.0 μM)
- Activity thresholds should match biological firing rates (0.1-100 Hz)
*/

type CustomBehaviors struct {
	// Direct callback for custom chemical release
	CustomChemicalRelease func(activityRate, outputValue float64, releaseFunc func(types.LigandType, float64) error)

	// Other custom behaviors...
}

type ChemicalRelease struct {
	LigandType    types.LigandType
	Concentration float64
}

// EnableCustomBehaviors allows custom behavior configuration
func (n *Neuron) EnableCustomBehaviors() {
	n.customBehaviors = &CustomBehaviors{}
}

// DisableCustomBehaviors removes custom behavior
func (n *Neuron) DisableCustomBehaviors() {
	n.customBehaviors = nil
}

// SetCustomChemicalRelease sets custom chemical release behavior
func (n *Neuron) SetCustomChemicalRelease(fn func(activityRate, outputValue float64, releaseFunc func(types.LigandType, float64) error)) {
	if n.customBehaviors == nil {
		n.EnableCustomBehaviors()
	}
	n.customBehaviors.CustomChemicalRelease = fn
}
