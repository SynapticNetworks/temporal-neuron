package neuron

import (
	"fmt"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// === NEURON CONFIGURATION ===
type NeuronConfig struct {
	// Spatial properties
	Position types.Position3D

	// Basic neural properties
	Threshold        float64
	DecayRate        float64
	RefractoryPeriod time.Duration
	FireFactor       float64

	// Homeostatic properties
	TargetFiringRate    float64
	HomeostasisStrength float64

	// Chemical properties (using message types)
	Receptors       []types.LigandType
	ReleasedLigands []types.LigandType

	// Synaptic scaling
	EnableSynapticScaling bool
	TargetInputStrength   float64
	ScalingRate           float64
	ScalingInterval       time.Duration

	// Type-safe dendritic integration configuration
	DendriticMode DendriticIntegrationMode // Direct mode instance instead of string-based

	// === ENHANCED PLASTICITY CONFIGURATION ===
	// Controls automatic STDP feedback behavior
	EnableSTDPFeedback bool          // Automatically send STDP feedback on firing
	STDPFeedbackDelay  time.Duration // Delay after firing before sending STDP feedback
	STDPLearningRate   float64       // Learning rate for STDP adjustments

	// Controls automatic homeostatic scaling
	EnableAutoScaling    bool          // Automatically perform homeostatic scaling
	ScalingCheckInterval time.Duration // How often to check for scaling needs

	// Controls automatic synaptic pruning
	EnableAutoPruning    bool          // Automatically prune dysfunctional synapses
	PruningCheckInterval time.Duration // How often to check for pruning candidates

	// Metadata
	Metadata map[string]interface{}
}

// === CONFIGURATION HELPERS ===

// DefaultExcitatoryConfig returns standard configuration for excitatory neurons
func DefaultExcitatoryConfig() NeuronConfig {
	return NeuronConfig{
		Threshold:             EXCITATORY_THRESHOLD_DEFAULT,
		DecayRate:             EXCITATORY_DECAY_RATE_DEFAULT,
		RefractoryPeriod:      EXCITATORY_REFRACTORY_PERIOD_DEFAULT,
		FireFactor:            EXCITATORY_FIRE_FACTOR_DEFAULT,
		TargetFiringRate:      EXCITATORY_TARGET_RATE_DEFAULT,
		HomeostasisStrength:   HOMEOSTASIS_STRENGTH_DEFAULT,
		EnableSynapticScaling: true,
		TargetInputStrength:   HOMEOSTASIS_TARGET_INPUT_STRENGTH_DEFAULT,
		ScalingRate:           HOMEOSTASIS_SCALING_RATE_DEFAULT,
		ScalingInterval:       HOMEOSTASIS_SCALING_INTERVAL_DEFAULT,
		EnableSTDPFeedback:    true,
		STDPFeedbackDelay:     STDP_FEEDBACK_DELAY_DEFAULT,
		STDPLearningRate:      STDP_LEARNING_RATE_EXCITATORY,
		EnableAutoScaling:     true,
		ScalingCheckInterval:  HOMEOSTASIS_CHECK_INTERVAL_DEFAULT,
		EnableAutoPruning:     false, // Conservative by default
		PruningCheckInterval:  PRUNING_CHECK_INTERVAL_DEFAULT,
	}
}

// DefaultInhibitoryConfig returns standard configuration for inhibitory neurons
func DefaultInhibitoryConfig() NeuronConfig {
	return NeuronConfig{
		Threshold:             INHIBITORY_THRESHOLD_DEFAULT,
		DecayRate:             INHIBITORY_DECAY_RATE_DEFAULT,
		RefractoryPeriod:      INHIBITORY_REFRACTORY_PERIOD_DEFAULT,
		FireFactor:            INHIBITORY_FIRE_FACTOR_DEFAULT,
		TargetFiringRate:      INHIBITORY_TARGET_RATE_DEFAULT,
		HomeostasisStrength:   HOMEOSTASIS_STRENGTH_DEFAULT,
		EnableSynapticScaling: true,
		TargetInputStrength:   HOMEOSTASIS_TARGET_INPUT_STRENGTH_DEFAULT,
		ScalingRate:           HOMEOSTASIS_SCALING_RATE_DEFAULT,
		ScalingInterval:       HOMEOSTASIS_SCALING_INTERVAL_DEFAULT,
		EnableSTDPFeedback:    true,
		STDPFeedbackDelay:     STDP_FEEDBACK_DELAY_SLOW, // More conservative
		STDPLearningRate:      STDP_LEARNING_RATE_INHIBITORY,
		EnableAutoScaling:     true,
		ScalingCheckInterval:  HOMEOSTASIS_CHECK_INTERVAL_DEFAULT,
		EnableAutoPruning:     false, // Conservative by default
		PruningCheckInterval:  PRUNING_CHECK_INTERVAL_CONSERVATIVE,
	}
}

// DefaultLearningConfig returns configuration optimized for learning scenarios
func DefaultLearningConfig() NeuronConfig {
	return NeuronConfig{
		Threshold:             EXCITATORY_THRESHOLD_DEFAULT,
		DecayRate:             EXCITATORY_DECAY_RATE_DEFAULT,
		RefractoryPeriod:      EXCITATORY_REFRACTORY_PERIOD_DEFAULT,
		FireFactor:            EXCITATORY_FIRE_FACTOR_DEFAULT,
		TargetFiringRate:      EXCITATORY_TARGET_RATE_DEFAULT,
		HomeostasisStrength:   HOMEOSTASIS_STRENGTH_AGGRESSIVE,
		EnableSynapticScaling: true,
		TargetInputStrength:   HOMEOSTASIS_TARGET_INPUT_STRENGTH_DEFAULT,
		ScalingRate:           HOMEOSTASIS_SCALING_RATE_DEFAULT,
		ScalingInterval:       HOMEOSTASIS_SCALING_INTERVAL_DEFAULT,
		EnableSTDPFeedback:    true,
		STDPFeedbackDelay:     STDP_FEEDBACK_DELAY_FAST,
		STDPLearningRate:      STDP_LEARNING_RATE_AGGRESSIVE,
		EnableAutoScaling:     true,
		ScalingCheckInterval:  HOMEOSTASIS_CHECK_INTERVAL_FAST,
		EnableAutoPruning:     true,
		PruningCheckInterval:  PRUNING_CHECK_INTERVAL_AGGRESSIVE,
	}
}

// DefaultConservativeConfig returns configuration for stable, conservative networks
func DefaultConservativeConfig() NeuronConfig {
	return NeuronConfig{
		Threshold:             EXCITATORY_THRESHOLD_DEFAULT,
		DecayRate:             EXCITATORY_DECAY_RATE_DEFAULT,
		RefractoryPeriod:      EXCITATORY_REFRACTORY_PERIOD_DEFAULT,
		FireFactor:            EXCITATORY_FIRE_FACTOR_DEFAULT,
		TargetFiringRate:      EXCITATORY_TARGET_RATE_DEFAULT,
		HomeostasisStrength:   HOMEOSTASIS_STRENGTH_CONSERVATIVE,
		EnableSynapticScaling: true,
		TargetInputStrength:   HOMEOSTASIS_TARGET_INPUT_STRENGTH_DEFAULT,
		ScalingRate:           HOMEOSTASIS_SCALING_RATE_DEFAULT,
		ScalingInterval:       HOMEOSTASIS_SCALING_INTERVAL_DEFAULT,
		EnableSTDPFeedback:    true,
		STDPFeedbackDelay:     STDP_FEEDBACK_DELAY_SLOW,
		STDPLearningRate:      STDP_LEARNING_RATE_CONSERVATIVE,
		EnableAutoScaling:     false, // Manual scaling only
		ScalingCheckInterval:  HOMEOSTASIS_CHECK_INTERVAL_SLOW,
		EnableAutoPruning:     false, // Manual pruning only
		PruningCheckInterval:  PRUNING_CHECK_INTERVAL_CONSERVATIVE,
	}
}

// === FACTORY FUNCTION SIGNATURE ===
// FIXED: Updated to use component.NeuronCallbacks and return component.NeuralComponent
type NeuronFactoryFunc func(id string, config NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error)

// === ENHANCED CALLBACK NEURON FACTORY ===
// FIXED: Updated signature to match the corrected factory function type
func CallbackNeuronFactory(id string, config NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
	// Validate that essential callbacks are provided for enabled features
	if err := validateCallbacks(config, callbacks); err != nil {
		return nil, fmt.Errorf("callback validation failed: %w", err)
	}

	neuron := NewNeuron(
		id,
		config.Threshold,
		config.DecayRate,
		config.RefractoryPeriod,
		config.FireFactor,
		config.TargetFiringRate,
		config.HomeostasisStrength,
	)

	// Set position
	neuron.SetPosition(config.Position)

	// Set chemical properties
	neuron.SetReceptors(config.Receptors)
	neuron.SetReleasedLigands(config.ReleasedLigands)

	// === INJECT ENHANCED MATRIX CALLBACKS ===
	neuron.SetCallbacks(callbacks)

	// Configure synaptic scaling
	if config.EnableSynapticScaling {
		neuron.EnableSynapticScaling(
			config.TargetInputStrength,
			config.ScalingRate,
			config.ScalingInterval,
		)
	}

	// Configure dendritic integration mode
	if config.DendriticMode != nil {
		err := neuron.SetDendriticMode(config.DendriticMode)
		if err != nil {
			return nil, fmt.Errorf("failed to set dendritic mode: %w", err)
		}
	}

	// === CONFIGURE ENHANCED PLASTICITY FEATURES ===
	if config.EnableSTDPFeedback {
		neuron.EnableSTDPFeedback(config.STDPFeedbackDelay, config.STDPLearningRate)
	}

	if config.EnableAutoScaling {
		neuron.EnableAutoHomeostasis(config.ScalingCheckInterval)
	}

	if config.EnableAutoPruning {
		neuron.EnableAutoPruning(config.PruningCheckInterval)
	}

	// Set metadata
	for key, value := range config.Metadata {
		neuron.UpdateMetadata(key, value)
	}

	// FIXED: Return as component.NeuralComponent interface
	return neuron, nil
}

// === VALIDATION FUNCTIONS ===

// FIXED: Updated to use component.NeuronCallbacks interface
func validateCallbacks(config NeuronConfig, callbacks component.NeuronCallbacks) error {
	// For interface-based callbacks, we can't check if function fields are nil
	// since they're methods. Instead, we just validate the interface is not nil
	if callbacks == nil {
		return fmt.Errorf("callbacks cannot be nil")
	}

	// The actual validation will happen at runtime when methods are called
	// This is the nature of interface-based dependency injection

	return nil
}

// === SPECIALIZED FACTORY VARIANTS ===

// FIXED: Updated all factory signatures to use component.NeuronCallbacks
func HomeostaticNeuronFactory(id string, config NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
	// Start with default excitatory configuration
	if isEmptyConfig(config) {
		config = DefaultExcitatoryConfig()
	}

	// Enhanced homeostatic configuration with automatic features
	config.EnableSynapticScaling = true
	config.EnableAutoScaling = true
	config.EnableSTDPFeedback = true

	// Override with aggressive homeostatic parameters
	config.HomeostasisStrength = HOMEOSTASIS_STRENGTH_AGGRESSIVE
	config.ScalingCheckInterval = HOMEOSTASIS_CHECK_INTERVAL_FAST
	config.STDPFeedbackDelay = STDP_FEEDBACK_DELAY_DEFAULT
	config.STDPLearningRate = STDP_LEARNING_RATE_DEFAULT

	return CallbackNeuronFactory(id, config, callbacks)
}

func ExcitatoryNeuronFactory(id string, config NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
	// Start with default excitatory configuration
	if isEmptyConfig(config) {
		config = DefaultExcitatoryConfig()
	}

	// Excitatory neuron-specific configuration
	config.ReleasedLigands = []types.LigandType{types.LigandGlutamate}
	config.Receptors = []types.LigandType{
		types.LigandGlutamate,
		types.LigandGABA,
		types.LigandDopamine,
	}

	// Enable STDP for excitatory neurons (they drive learning)
	config.EnableSTDPFeedback = true
	config.STDPFeedbackDelay = STDP_FEEDBACK_DELAY_DEFAULT
	config.STDPLearningRate = STDP_LEARNING_RATE_EXCITATORY

	return CallbackNeuronFactory(id, config, callbacks)
}

func InhibitoryNeuronFactory(id string, config NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
	// Start with default inhibitory configuration
	if isEmptyConfig(config) {
		config = DefaultInhibitoryConfig()
	}

	// Inhibitory neuron-specific configuration
	config.ReleasedLigands = []types.LigandType{types.LigandGABA}
	config.Receptors = []types.LigandType{
		types.LigandGlutamate,
		types.LigandGABA,
		types.LigandSerotonin,
	}

	// Inhibitory neurons typically have different plasticity characteristics
	config.EnableSTDPFeedback = true
	config.STDPFeedbackDelay = STDP_FEEDBACK_DELAY_SLOW // More conservative
	config.STDPLearningRate = STDP_LEARNING_RATE_INHIBITORY

	// Enable homeostatic scaling for stability
	config.EnableAutoScaling = true
	config.ScalingCheckInterval = HOMEOSTASIS_CHECK_INTERVAL_DEFAULT

	return CallbackNeuronFactory(id, config, callbacks)
}

// === LEARNING-FOCUSED FACTORY VARIANTS ===

func LearningNeuronFactory(id string, config NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
	// Start with learning-optimized configuration
	if isEmptyConfig(config) {
		config = DefaultLearningConfig()
	}

	// Ensure all plasticity features are enabled
	config.EnableSTDPFeedback = true
	config.EnableAutoScaling = true
	config.EnableAutoPruning = true
	config.EnableSynapticScaling = true

	// Use aggressive learning parameters
	config.STDPFeedbackDelay = STDP_FEEDBACK_DELAY_FAST
	config.STDPLearningRate = STDP_LEARNING_RATE_AGGRESSIVE
	config.ScalingCheckInterval = HOMEOSTASIS_CHECK_INTERVAL_FAST
	config.PruningCheckInterval = PRUNING_CHECK_INTERVAL_AGGRESSIVE

	return CallbackNeuronFactory(id, config, callbacks)
}

func ConservativeNeuronFactory(id string, config NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
	// Start with conservative configuration
	if isEmptyConfig(config) {
		config = DefaultConservativeConfig()
	}

	// Conservative plasticity settings
	config.EnableSTDPFeedback = true
	config.EnableAutoScaling = false // Manual scaling only
	config.EnableAutoPruning = false // Manual pruning only

	// Use conservative learning parameters
	config.STDPFeedbackDelay = STDP_FEEDBACK_DELAY_SLOW
	config.STDPLearningRate = STDP_LEARNING_RATE_CONSERVATIVE
	config.HomeostasisStrength = HOMEOSTASIS_STRENGTH_CONSERVATIVE

	return CallbackNeuronFactory(id, config, callbacks)
}

// === DENDRITIC MODE FACTORY FUNCTIONS ===

// FIXED: Updated all dendritic factory signatures
func CreatePassiveDendriticNeuron(id string, config NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
	if isEmptyConfig(config) {
		config = DefaultExcitatoryConfig()
	}
	config.DendriticMode = NewPassiveMembraneMode()
	return CallbackNeuronFactory(id, config, callbacks)
}

func CreateBiologicalDendriticNeuron(id string, config NeuronConfig, callbacks component.NeuronCallbacks, bioConfig BiologicalConfig) (component.NeuralComponent, error) {
	if isEmptyConfig(config) {
		config = DefaultExcitatoryConfig()
	}
	config.DendriticMode = NewBiologicalTemporalSummationMode(bioConfig)
	return CallbackNeuronFactory(id, config, callbacks)
}

func CreateActiveDendriticNeuron(id string, config NeuronConfig, callbacks component.NeuronCallbacks, activeConfig ActiveDendriteConfig, bioConfig BiologicalConfig) (component.NeuralComponent, error) {
	if isEmptyConfig(config) {
		config = DefaultExcitatoryConfig()
	}
	config.DendriticMode = NewActiveDendriteMode(activeConfig, bioConfig)
	return CallbackNeuronFactory(id, config, callbacks)
}

func CreateCorticalPyramidalNeuron(id string, config NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
	if isEmptyConfig(config) {
		config = DefaultExcitatoryConfig()
	}

	// Use cortical biological parameters
	bioConfig := CreateCorticalPyramidalConfig()

	// Create active dendrite configuration with NMDA coincidence detection
	activeConfig := CreateActiveDendriteConfig()

	config.DendriticMode = NewActiveDendriteMode(activeConfig, bioConfig)
	return CallbackNeuronFactory(id, config, callbacks)
}

func CreateCustomCoincidenceNeuron(id string, config NeuronConfig, callbacks component.NeuronCallbacks, detectorConfig CoincidenceDetectorConfig) (component.NeuralComponent, error) {
	if isEmptyConfig(config) {
		config = DefaultExcitatoryConfig()
	}

	// Create biological configuration
	bioConfig := CreateCorticalPyramidalConfig()

	// Create active dendrite configuration with custom detector
	activeConfig := ActiveDendriteConfig{
		MaxSynapticEffect:       DENDRITE_CURRENT_SATURATION_DEFAULT,
		ShuntingStrength:        DENDRITE_FACTOR_SHUNTING_DEFAULT,
		DendriticSpikeThreshold: DENDRITE_CURRENT_SPIKE_THRESHOLD_DEFAULT,
		NMDASpikeAmplitude:      DENDRITE_CURRENT_SPIKE_AMPLITUDE_DEFAULT,
		VoltageThreshold:        DENDRITE_VOLTAGE_SPIKE_THRESHOLD_DEFAULT,
		CoincidenceDetector:     detectorConfig, // Type-safe custom detector
	}

	config.DendriticMode = NewActiveDendriteMode(activeConfig, bioConfig)
	return CallbackNeuronFactory(id, config, callbacks)
}

// === UTILITY FUNCTIONS ===

// isEmptyConfig checks if a config struct has default/zero values
func isEmptyConfig(config NeuronConfig) bool {
	return config.Threshold == 0 && config.DecayRate == 0 && config.RefractoryPeriod == 0
}

// === FACTORY REGISTRATION ===
func RegisterFactories(registerFunc func(string, NeuronFactoryFunc)) {
	registerFunc("basic", CallbackNeuronFactory)
	registerFunc("homeostatic", HomeostaticNeuronFactory)
	registerFunc("excitatory", ExcitatoryNeuronFactory)
	registerFunc("inhibitory", InhibitoryNeuronFactory)
	registerFunc("learning", LearningNeuronFactory)
	registerFunc("conservative", ConservativeNeuronFactory)

	// Dendritic mode factories (require additional parameters, so register as callable creators)
	// These would typically be used through the specific Create* functions above
}
