package extracellular

import (
	"fmt"
	"math"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

// =================================================================================
// PLASTICITY ADJUSTMENT TYPE DEFINITION
// =================================================================================

// PlasticityAdjustment represents synaptic plasticity feedback from post-synaptic neurons
// This models retrograde signaling mechanisms in biological neural networks
type PlasticityAdjustment struct {
	// DeltaT is the time difference for STDP calculations (t_pre - t_post)
	// Convention:
	//   - Negative values (pre before post) → LTP (strengthening)
	//   - Positive values (pre after post) → LTD (weakening)
	DeltaT time.Duration `json:"delta_t"`

	// WeightChange provides direct weight modification (optional)
	// Use this for non-STDP plasticity mechanisms
	WeightChange float64 `json:"weight_change"`

	// LearningRate allows context-specific learning rate override
	// If 0, synapse uses its default learning rate
	LearningRate float64 `json:"learning_rate"`

	// PlasticityType specifies the learning mechanism (using existing enum)
	PlasticityType types.PlasticityEventType `json:"plasticity_type"`

	// Source identifies the neuron initiating the plasticity
	SourceNeuronID string `json:"source_neuron_id"`

	// Timestamp when the plasticity event occurred
	Timestamp time.Time `json:"timestamp"`
}

// Note: Using existing PlasticityEventType enum:
// const (
//     PlasticitySTDP PlasticityEventType = iota
//     PlasticityBCM
//     PlasticityOja
//     PlasticityHomeostatic
// )

// =================================================================================
// MATRIX ENHANCED CALLBACK IMPLEMENTATIONS
// =================================================================================

// SetSynapseWeight directly modifies a synapse's weight with biological constraints
//
// BIOLOGICAL FUNCTION:
// This models direct synaptic modification mechanisms found in biology:
// - Homeostatic scaling during sleep/development
// - Neuromodulator-induced weight changes
// - Experimental optogenetic manipulation
// - Metaplasticity (plasticity of plasticity)
//
// The function ensures biological realism by enforcing weight bounds and
// triggering appropriate biological responses to weight changes.
func (ecm *ExtracellularMatrix) SetSynapseWeight(synapseID string, weight float64) error {
	ecm.mu.RLock()
	synapse, exists := ecm.synapses[synapseID]
	ecm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("synapse not found: %s", synapseID)
	}

	// Validate weight bounds to maintain biological realism
	if weight < 0 {
		return fmt.Errorf("synaptic weight cannot be negative: %f", weight)
	}

	// Get current weight for change tracking
	oldWeight := synapse.GetWeight()

	// Apply weight change through synapse's native method
	// This ensures the synapse can apply its own biological constraints
	synapse.SetWeight(weight)

	// Verify the weight was actually changed (synapse may have clamped it)
	newWeight := synapse.GetWeight()
	weightChange := newWeight - oldWeight

	// Report the weight change to biological monitoring systems
	if math.Abs(weightChange) > 0.001 { // Only report significant changes
		// Use existing astrocyte network method (RecordSynapticActivity)
		preID := synapse.GetPresynapticID()
		postID := synapse.GetPostsynapticID()

		// Record the synaptic activity with new weight
		ecm.astrocyteNetwork.RecordSynapticActivity(synapseID, preID, postID, newWeight)

		// Use existing microglia method (UpdateComponentHealth)
		// Report to microglia for network health monitoring
		ecm.microglia.UpdateComponentHealth(synapseID, math.Abs(weightChange), 1)

		// Create plasticity event for biological consistency
		plasticityEvent := types.PlasticityEvent{
			EventType: types.PlasticityHomeostatic, // Use existing enum for manual changes
			Timestamp: time.Now(),
			PreTime:   time.Now(), // Dummy times for manual changes
			PostTime:  time.Now(),
			Strength:  weightChange,
			SourceID:  "matrix_admin", // Administrative change
		}

		// Update synapse with plasticity event
		synapse.UpdateWeight(plasticityEvent)
	}

	return nil
}

// ApplyPlasticity triggers synaptic plasticity based on neural activity patterns
//
// BIOLOGICAL FUNCTION:
// This models the complete biological plasticity cascade:
// 1. Post-synaptic neuron fires and sends retrograde signals
// 2. Retrograde signals carry timing and activity information
// 3. Pre-synaptic terminals receive and decode the signals
// 4. Synaptic strength is modified based on biological learning rules
// 5. Changes are integrated into network-wide homeostatic mechanisms
//
// This function coordinates between multiple biological systems to ensure
// plasticity events are properly processed and integrated.
func (ecm *ExtracellularMatrix) ApplyPlasticity(synapseID string, adjustment PlasticityAdjustment) error {
	ecm.mu.RLock()
	synapse, exists := ecm.synapses[synapseID]
	ecm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("synapse not found: %s", synapseID)
	}

	// Record pre-change state for monitoring
	oldWeight := synapse.GetWeight()

	// Create comprehensive plasticity event
	plasticityEvent := types.PlasticityEvent{
		EventType: adjustment.PlasticityType,
		Timestamp: adjustment.Timestamp,
		SourceID:  adjustment.SourceNeuronID,
		Strength:  0, // Will be calculated based on plasticity type
	}

	// Handle different plasticity mechanisms using existing enum
	switch adjustment.PlasticityType {
	case types.PlasticitySTDP:
		// STDP requires precise timing information
		if adjustment.DeltaT == 0 {
			return fmt.Errorf("STDP plasticity requires non-zero DeltaT")
		}

		// Calculate pre/post spike times from DeltaT
		// Convention: DeltaT = t_pre - t_post
		postTime := adjustment.Timestamp
		preTime := postTime.Add(adjustment.DeltaT)

		plasticityEvent.PreTime = preTime
		plasticityEvent.PostTime = postTime

		// Calculate STDP strength based on timing
		plasticityEvent.Strength = ecm.calculateSTDPStrength(adjustment.DeltaT, adjustment.LearningRate)

	case types.PlasticityHomeostatic:
		// Homeostatic scaling uses direct weight modification
		plasticityEvent.PreTime = adjustment.Timestamp
		plasticityEvent.PostTime = adjustment.Timestamp
		plasticityEvent.Strength = adjustment.WeightChange

	case types.PlasticityBCM:
		// BCM plasticity for future implementation
		plasticityEvent.PreTime = adjustment.Timestamp
		plasticityEvent.PostTime = adjustment.Timestamp
		plasticityEvent.Strength = adjustment.WeightChange

	case types.PlasticityOja:
		// Oja's learning rule for future implementation
		plasticityEvent.PreTime = adjustment.Timestamp
		plasticityEvent.PostTime = adjustment.Timestamp
		plasticityEvent.Strength = adjustment.WeightChange

	default:
		return fmt.Errorf("unsupported plasticity type: %v", adjustment.PlasticityType)
	}

	// Apply plasticity through synapse's biological mechanisms
	synapse.UpdateWeight(plasticityEvent)

	// Monitor biological network effects
	newWeight := synapse.GetWeight()
	weightChange := newWeight - oldWeight

	if math.Abs(weightChange) > 0.001 {
		// Report plasticity event to biological monitoring systems
		preID := synapse.GetPresynapticID()
		postID := synapse.GetPostsynapticID()

		// Use existing astrocyte network method to record synaptic activity
		ecm.astrocyteNetwork.RecordSynapticActivity(
			synapseID, preID, postID, newWeight)

		// Use existing microglia method for health monitoring
		// Report plasticity event as a health update
		ecm.microglia.UpdateComponentHealth(synapseID, math.Abs(weightChange), 1)

		// Trigger chemical signaling if significant change
		if math.Abs(weightChange) > 0.01 {
			// Model calcium/protein synthesis signaling using glutamate as proxy for plasticity signals
			concentration := math.Abs(weightChange) * 10.0 // Scale to biological range
			ecm.chemicalModulator.Release(LigandCalcium, synapseID, concentration)
		}
	}

	return nil
}

// =================================================================================
// BIOLOGICAL HELPER FUNCTIONS
// =================================================================================

// calculateSTDPStrength computes STDP weight change based on spike timing
// This implements the canonical asymmetric STDP learning window
func (ecm *ExtracellularMatrix) calculateSTDPStrength(deltaT time.Duration, learningRate float64) float64 {
	// Use default learning rate if not specified
	if learningRate == 0 {
		learningRate = STDP_DEFAULT_LEARNING_RATE
	}

	// Convert to milliseconds for calculation
	deltaTMs := deltaT.Seconds() * 1000.0

	// STDP time constant (typical biological value: 20ms)
	tauMs := STDP_DEFAULT_TIME_CONSTANT.Seconds() * 1000.0

	// Check if within STDP window
	windowMs := STDP_DEFAULT_WINDOW_SIZE.Seconds() * 1000.0
	if math.Abs(deltaTMs) >= windowMs {
		return 0.0 // No plasticity outside window
	}

	if deltaTMs < 0 {
		// Causal (LTP): pre before post
		return learningRate * math.Exp(deltaTMs/tauMs)
	} else if deltaTMs > 0 {
		// Anti-causal (LTD): pre after post
		return -learningRate * STDP_DEFAULT_ASYMMETRY_RATIO * math.Exp(-deltaTMs/tauMs)
	}

	// Simultaneous spikes
	return -learningRate * STDP_DEFAULT_ASYMMETRY_RATIO * 0.1
}

// =================================================================================
// BIOLOGICAL CONSTANTS FOR PLASTICITY
// =================================================================================

const (
	// STDP learning parameters (from biological measurements)
	STDP_DEFAULT_LEARNING_RATE   = 0.01                   // 1% weight change per event
	STDP_DEFAULT_TIME_CONSTANT   = 20 * time.Millisecond  // Exponential decay constant
	STDP_DEFAULT_WINDOW_SIZE     = 100 * time.Millisecond // Maximum timing window
	STDP_DEFAULT_ASYMMETRY_RATIO = 1.05                   // LTD slightly stronger than LTP

	// Weight bounds
	SYNAPSE_MIN_WEIGHT = 0.001 // Prevent complete elimination
	SYNAPSE_MAX_WEIGHT = 2.0   // Prevent runaway strengthening
)

// =================================================================================
// EXTENDED BIOLOGICAL FUNCTIONALITY
// =================================================================================

// Additional methods that extend biological realism for the matrix

// getBiologicalWeightBounds returns appropriate weight limits for synapse type
func (ecm *ExtracellularMatrix) getBiologicalWeightBounds(synapseID string) (min, max float64) {
	synapse, exists := ecm.synapses[synapseID]
	if !exists {
		return SYNAPSE_MIN_WEIGHT, SYNAPSE_MAX_WEIGHT
	}

	// Get plasticity configuration from synapse
	config := synapse.GetPlasticityConfig()
	return config.MinWeight, config.MaxWeight
}

// validatePlasticityAdjustment ensures biological realism of plasticity parameters
func (ecm *ExtracellularMatrix) validatePlasticityAdjustment(adjustment PlasticityAdjustment) error {
	// Validate timing constraints for STDP
	if adjustment.PlasticityType == types.PlasticitySTDP {
		if math.Abs(adjustment.DeltaT.Seconds()) > 0.2 { // 200ms biological limit
			return fmt.Errorf("STDP DeltaT outside biological range: %v", adjustment.DeltaT)
		}
	}

	// Validate learning rate bounds
	if adjustment.LearningRate < 0 || adjustment.LearningRate > 0.1 {
		return fmt.Errorf("learning rate outside biological range: %f", adjustment.LearningRate)
	}

	// Validate weight change bounds
	if math.Abs(adjustment.WeightChange) > 1.0 {
		return fmt.Errorf("weight change too large for single event: %f", adjustment.WeightChange)
	}

	return nil
}
