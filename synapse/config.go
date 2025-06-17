/*
=================================================================================
SYNAPSE CONFIGURATION - BIOLOGICAL PRESETS AND CONFIGURATION HELPERS
=================================================================================

This module provides comprehensive configuration management for the synapse
system, including biologically accurate presets for different synapse types,
validation functions, and helper utilities for creating realistic neural
connections.

BIOLOGICAL SYNAPSE TYPES MODELED:
1. Excitatory Glutamatergic Synapses - Fast cortical excitation
2. Inhibitory GABAergic Synapses - Fast cortical inhibition
3. Neuromodulatory Synapses - Dopamine, serotonin, acetylcholine
4. Developmental Synapses - Enhanced plasticity for learning
5. Aged Synapses - Reduced plasticity modeling aging
6. Pathological Synapses - Disease-state configurations

DESIGN PRINCIPLES:
- Biologically grounded default parameters
- Easy preset selection for common synapse types
- Comprehensive validation with helpful error messages
- Extensible configuration system for research applications
- Matrix integration support through callback configuration

EXPERIMENTAL BASIS:
All configurations are derived from published neuroscience research with
references provided in constants.go. Parameters represent typical ranges
observed in mammalian cortical preparations unless otherwise specified.
=================================================================================
*/

package synapse

import (
	"fmt"
	"strings"
	"time"
)

// =================================================================================
// CORE CONFIGURATION STRUCTURES
// =================================================================================

// SynapseConfig defines the complete configuration for a biologically realistic synapse.
// This is the main configuration structure used by factory systems to create new synapses.
type SynapseConfig struct {
	// === CORE IDENTIFICATION ===
	SynapseID      string `json:"synapse_id"`      // Unique identifier for the synapse.
	SynapseType    string `json:"synapse_type"`    // Biological type classification (e.g., "excitatory_glutamatergic").
	PresynapticID  string `json:"presynaptic_id"`  // Identifier for the source neuron.
	PostsynapticID string `json:"postsynaptic_id"` // Identifier for the target neuron.

	// === TRANSMISSION PROPERTIES ===
	InitialWeight     float64       `json:"initial_weight"`      // Starting synaptic strength, must be within STDPConfig bounds.
	BaseSynapticDelay time.Duration `json:"base_synaptic_delay"` // Base processing delay, excluding spatial/axonal components.

	// === BIOLOGICAL PROPERTIES ===
	NeurotransmitterType LigandType `json:"neurotransmitter_type"` // The primary chemical signaling molecule for this synapse.
	Position             Position3D `json:"position"`              // The 3D spatial location of the synapse in micrometers.

	// === VESICLE CONFIGURATION ===
	VesicleConfig VesicleConfig `json:"vesicle_config"` // Parameters governing neurotransmitter release dynamics.

	// === PLASTICITY CONFIGURATION ===
	STDPConfig    STDPConfig    `json:"stdp_config"`    // Configuration for Spike-Timing Dependent Plasticity.
	PruningConfig PruningConfig `json:"pruning_config"` // Configuration for structural plasticity (synapse elimination).

	// === INTEGRATION METADATA ===
	Metadata map[string]interface{} `json:"metadata"` // Flexible key-value store for additional properties.

	// === MATRIX INTEGRATION ===
	MatrixIntegration bool `json:"matrix_integration"` // Flag to enable matrix-level callbacks and coordination.

	// === VALIDATION FLAGS ===
	SkipValidation bool `json:"skip_validation"` // If true, skips validation; intended for testing purposes only.
}

// VesicleConfig controls neurotransmitter release dynamics and constraints,
// modeling the biological machinery of synaptic vesicles.
type VesicleConfig struct {
	// === VESICLE SYSTEM CONTROL ===
	Enabled bool `json:"enabled"` // Master switch to enable or disable vesicle-based rate limiting.

	// === RELEASE KINETICS ===
	MaxReleaseRate      float64 `json:"max_release_rate"`     // Maximum sustainable release rate in Hertz (Hz).
	BaselineProbability float64 `json:"baseline_probability"` // Baseline probability of vesicle release per stimulus (0.0 to 1.0).
	CalciumSensitivity  float64 `json:"calcium_sensitivity"`  // Factor modulating release probability based on calcium levels.

	// === VESICLE POOL CONFIGURATION ===
	ReadyPoolSize     int `json:"ready_pool_size"`     // Ready Releasable Pool (RRP) size.
	RecyclingPoolSize int `json:"recycling_pool_size"` // Recycling Pool size.
	ReservePoolSize   int `json:"reserve_pool_size"`   // Reserve Pool size.

	// === RECYCLING KINETICS ===
	FastRecyclingTime time.Duration `json:"fast_recycling_time"` // Time constant for fast, kiss-and-run endocytosis.
	SlowRecyclingTime time.Duration `json:"slow_recycling_time"` // Time constant for slow, clathrin-mediated endocytosis.
	RefillTime        time.Duration `json:"refill_time"`         // Time required to load vesicles with neurotransmitter.

	// === FATIGUE AND DEPRESSION ===
	FatigueThreshold float64       `json:"fatigue_threshold"` // Activity level (Hz) above which fatigue accumulates.
	RecoveryTime     time.Duration `json:"recovery_time"`     // Time constant for recovery from synaptic depression.
	DepletionFactor  float64       `json:"depletion_factor"`  // Sensitivity to vesicle pool depletion.
}

// =================================================================================
// BIOLOGICAL SYNAPSE TYPE PRESETS
// =================================================================================

// CreateExcitatoryGlutamatergicConfig creates a standard configuration for a fast
// excitatory cortical synapse.
// Biological basis: Typical glutamatergic synapse found on cortical pyramidal neurons,
// characterized by fast transmission, moderate plasticity, and AMPA/NMDA receptors.
func CreateExcitatoryGlutamatergicConfig(synapseID, preID, postID string) SynapseConfig {
	return SynapseConfig{
		SynapseID:            synapseID,
		SynapseType:          "excitatory_glutamatergic",
		PresynapticID:        preID,
		PostsynapticID:       postID,
		InitialWeight:        0.5,
		BaseSynapticDelay:    TYPICAL_SYNAPTIC_DELAY,
		NeurotransmitterType: LigandGlutamate,
		Position:             Position3D{X: 0, Y: 0, Z: 0},
		VesicleConfig: VesicleConfig{
			Enabled:             true,
			MaxReleaseRate:      GLUTAMATE_MAX_RATE,
			BaselineProbability: BASELINE_RELEASE_PROBABILITY,
			CalciumSensitivity:  MAX_CALCIUM_ENHANCEMENT / 2.0,
			ReadyPoolSize:       DEFAULT_READY_POOL_SIZE,
			RecyclingPoolSize:   DEFAULT_RECYCLING_POOL_SIZE,
			ReservePoolSize:     DEFAULT_RESERVE_POOL_SIZE,
			FastRecyclingTime:   FAST_RECYCLING_TIME,
			SlowRecyclingTime:   SLOW_RECYCLING_TIME,
			RefillTime:          VESICLE_REFILL_TIME,
			FatigueThreshold:    HIGH_FREQUENCY_THRESHOLD,
			RecoveryTime:        FATIGUE_RECOVERY_TIME,
			DepletionFactor:     0.8,
		},
		STDPConfig:        CreateDefaultSTDPConfig(),
		PruningConfig:     CreateDefaultPruningConfig(),
		MatrixIntegration: true,
		Metadata: map[string]interface{}{
			"synapse_class":    "excitatory",
			"receptor_types":   []string{"AMPA", "NMDA"},
			"plasticity_type":  "stdp",
			"biological_model": "cortical_pyramidal",
		},
	}
}

func CreateDefaultPruningConfig() PruningConfig {
	return PruningConfig{
		Enabled:             true,
		WeightThreshold:     0.01,
		InactivityThreshold: 5 * time.Minute,
	}
}

// CreateInhibitoryGABAergicConfig creates a configuration for a fast inhibitory
// cortical synapse.
// Biological basis: Typical GABAergic synapse from a fast-spiking interneuron onto
// a pyramidal cell, characterized by rapid transmission and often limited plasticity to ensure stability.
func CreateInhibitoryGABAergicConfig(synapseID, preID, postID string) SynapseConfig {
	stdpConfig := CreateDefaultSTDPConfig()
	stdpConfig.Enabled = false // Inhibitory plasticity is often disabled for stability
	stdpConfig.MinWeight = 0.1 // Ensure inhibition is not completely eliminated
	stdpConfig.MaxWeight = 2.0

	pruningConfig := CreateDefaultPruningConfig()
	pruningConfig.Enabled = false // Often protected from pruning to maintain balance

	return SynapseConfig{
		SynapseID:            synapseID,
		SynapseType:          "inhibitory_gabaergic",
		PresynapticID:        preID,
		PostsynapticID:       postID,
		InitialWeight:        0.8,
		BaseSynapticDelay:    MIN_SYNAPTIC_DELAY * 2, // Faster than excitatory
		NeurotransmitterType: LigandGABA,
		Position:             Position3D{X: 0, Y: 0, Z: 0},
		VesicleConfig: VesicleConfig{
			Enabled:             true,
			MaxReleaseRate:      GABA_MAX_RATE,
			BaselineProbability: BASELINE_RELEASE_PROBABILITY * 1.2, // Slightly higher prob
			CalciumSensitivity:  MAX_CALCIUM_ENHANCEMENT / 1.5,
			ReadyPoolSize:       DEFAULT_READY_POOL_SIZE,
			RecyclingPoolSize:   DEFAULT_RECYCLING_POOL_SIZE,
			ReservePoolSize:     DEFAULT_RESERVE_POOL_SIZE,
			FastRecyclingTime:   time.Duration(float64(FAST_RECYCLING_TIME) * 0.8), // Faster recycling
			SlowRecyclingTime:   time.Duration(float64(SLOW_RECYCLING_TIME) * 0.8),
			RefillTime:          time.Duration(float64(VESICLE_REFILL_TIME) * 0.7),   // Faster NT loading
			FatigueThreshold:    HIGH_FREQUENCY_THRESHOLD * 2.5,                      // More resistant to fatigue
			RecoveryTime:        time.Duration(float64(FATIGUE_RECOVERY_TIME) * 0.7), // Faster recovery
			DepletionFactor:     0.6,
		},
		STDPConfig:        stdpConfig,
		PruningConfig:     pruningConfig,
		MatrixIntegration: true,
		Metadata: map[string]interface{}{
			"synapse_class":    "inhibitory",
			"receptor_types":   []string{"GABA-A", "GABA-B"},
			"plasticity_type":  "static",
			"biological_model": "fast_spiking_interneuron",
		},
	}
}

// CreateDopaminergicNeuromodulatoryConfig creates a configuration for a neuromodulatory dopamine synapse.
// Biological basis: Dopaminergic projections from the VTA or SNc, characterized by
// slow, volume-based transmission that modulates learning and motivation.
func CreateDopaminergicNeuromodulatoryConfig(synapseID, preID, postID string) SynapseConfig {
	stdpConfig := CreateDefaultSTDPConfig()
	stdpConfig.LearningRate *= DOPAMINE_LEARNING_MULTIPLIER
	stdpConfig.WindowSize *= 2
	stdpConfig.AsymmetryRatio = 0.8 // Bias towards LTP for reward signaling

	pruningConfig := CreateDefaultPruningConfig()
	pruningConfig.Enabled = false // Neuromodulatory synapses are critical infrastructure

	return SynapseConfig{
		SynapseID:            synapseID,
		SynapseType:          "neuromodulatory_dopaminergic",
		PresynapticID:        preID,
		PostsynapticID:       postID,
		InitialWeight:        0.3,
		BaseSynapticDelay:    MAX_SYNAPTIC_DELAY, // Very slow transmission
		NeurotransmitterType: LigandDopamine,
		Position:             Position3D{X: 0, Y: 0, Z: 0},
		VesicleConfig: VesicleConfig{
			Enabled:             true,
			MaxReleaseRate:      DOPAMINE_MAX_RATE,
			BaselineProbability: BASELINE_RELEASE_PROBABILITY * 0.5, // Lower baseline release
			CalciumSensitivity:  MAX_CALCIUM_ENHANCEMENT,            // Highly sensitive to bursts
			ReadyPoolSize:       DEFAULT_READY_POOL_SIZE / 2,
			RecyclingPoolSize:   DEFAULT_RECYCLING_POOL_SIZE / 2,
			ReservePoolSize:     DEFAULT_RESERVE_POOL_SIZE / 2,
			FastRecyclingTime:   FAST_RECYCLING_TIME * 2,
			SlowRecyclingTime:   SLOW_RECYCLING_TIME * 2,
			RefillTime:          VESICLE_REFILL_TIME * 3, // Very slow NT synthesis
			FatigueThreshold:    HIGH_FREQUENCY_THRESHOLD / 2.0,
			RecoveryTime:        FATIGUE_RECOVERY_TIME * 4,
			DepletionFactor:     1.2,
		},
		STDPConfig:        stdpConfig,
		PruningConfig:     pruningConfig,
		MatrixIntegration: true,
		Metadata: map[string]interface{}{
			"synapse_class":    "neuromodulatory",
			"receptor_types":   []string{"D1", "D2"},
			"plasticity_type":  "reward_modulated_stdp",
			"biological_model": "vta_projection",
			"signaling_mode":   "volume_transmission",
		},
	}
}

// CreateDevelopmentalConfig adapts a base configuration to model a synapse
// during a critical developmental period.
// Biological basis: Juvenile synapses exhibit enhanced plasticity (higher learning rates,
// wider STDP windows) to facilitate rapid learning and circuit formation.
func CreateDevelopmentalConfig(synapseID, preID, postID string) SynapseConfig {
	config := CreateExcitatoryGlutamatergicConfig(synapseID, preID, postID)
	config.SynapseType = "developmental_plastic"

	// Enhance vesicle dynamics for higher activity levels
	config.VesicleConfig.MaxReleaseRate *= 1.5
	config.VesicleConfig.BaselineProbability *= 1.3
	config.VesicleConfig.RecoveryTime /= 2

	// Apply developmental plasticity rules
	config.STDPConfig = CreateDevelopmentalSTDPConfig()

	// Protect new synapses from being pruned too early
	config.PruningConfig.ProtectionPeriod = PRUNING_PROTECTION_PERIOD * 10

	config.Metadata["developmental_stage"] = "critical_period"
	config.Metadata["plasticity_enhancement"] = CRITICAL_PERIOD_MULTIPLIER
	return config
}

// CreateAgedConfig adapts a base configuration to model an aged synapse.
// Biological basis: Aged synapses often show reduced plasticity, slower kinetics,
// and are more susceptible to dysfunction.
func CreateAgedConfig(synapseID, preID, postID string) SynapseConfig {
	config := CreateExcitatoryGlutamatergicConfig(synapseID, preID, postID)
	config.SynapseType = "aged_reduced_plasticity"

	// Impair vesicle dynamics
	config.VesicleConfig.MaxReleaseRate *= 0.7
	config.VesicleConfig.BaselineProbability *= 0.8
	config.VesicleConfig.RecoveryTime *= 2
	config.VesicleConfig.RefillTime = time.Duration(float64(config.VesicleConfig.RefillTime) * 1.5)

	// Apply age-related plasticity reduction
	config.STDPConfig = CreateAgedSTDPConfig()

	// Make pruning more likely for inefficient synapses
	config.PruningConfig.Enabled = true
	config.PruningConfig.WeightThreshold *= 1.5
	config.PruningConfig.InactivityThreshold /= 2
	config.PruningConfig.PruningProbability *= 2.0

	config.Metadata["developmental_stage"] = "aged"
	config.Metadata["plasticity_reduction"] = AGING_PLASTICITY_REDUCTION
	return config
}

// =================================================================================
// VALIDATION FUNCTIONS
// =================================================================================

// ValidateConfig performs a comprehensive validation of the synapse configuration
// against biologically plausible ranges defined in constants.go.
// It returns a ValidationResult containing any errors or warnings.
func ValidateConfig(config SynapseConfig) ValidationResult {
	result := NewValidationResult()

	if config.SkipValidation {
		return result
	}

	// Validate core identifiers
	if config.SynapseID == "" {
		result.AddError("synapse ID cannot be empty")
	}
	if config.PresynapticID == "" {
		result.AddError("presynaptic neuron ID cannot be empty")
	}
	if config.PostsynapticID == "" {
		result.AddError("postsynaptic neuron ID cannot be empty")
	}
	if config.PresynapticID == config.PostsynapticID {
		result.AddWarning("presynaptic and postsynaptic IDs are the same (self-connection)")
	}

	// Validate transmission properties
	if config.InitialWeight < config.STDPConfig.MinWeight || config.InitialWeight > config.STDPConfig.MaxWeight {
		result.AddError(fmt.Sprintf("initial weight %.4f is outside the configured STDP bounds [%.4f, %.4f]",
			config.InitialWeight, config.STDPConfig.MinWeight, config.STDPConfig.MaxWeight))
	}
	if config.BaseSynapticDelay < BIOLOGICAL_MIN_DELAY {
		result.AddError(fmt.Sprintf("base synaptic delay %v is below the biological minimum %v",
			config.BaseSynapticDelay, BIOLOGICAL_MIN_DELAY))
	}
	if config.BaseSynapticDelay > BIOLOGICAL_MAX_DELAY {
		result.AddError(fmt.Sprintf("base synaptic delay %v exceeds the biological maximum %v",
			config.BaseSynapticDelay, BIOLOGICAL_MAX_DELAY))
	}

	// Validate spatial position
	if !config.Position.IsValid() {
		result.AddError("spatial position contains invalid coordinates (NaN or Inf)")
	}

	// Validate sub-configurations
	validateVesicleConfig(config.VesicleConfig, &result)

	if !config.STDPConfig.IsValid() {
		result.AddError("STDP configuration is invalid")
	} else {
		for _, warning := range ValidateSTDPParameters(config.STDPConfig) {
			result.AddWarning(warning)
		}
	}

	if !config.PruningConfig.IsValid() {
		result.AddError("pruning configuration is invalid")
	}

	// Validate for logical consistency
	validateNeurotransmitterCompatibility(config, &result)

	return result
}

// validateVesicleConfig validates the vesicle dynamics sub-configuration.
func validateVesicleConfig(vconfig VesicleConfig, result *ValidationResult) {
	if !vconfig.Enabled {
		return // No validation needed if vesicle dynamics are disabled.
	}

	if vconfig.MaxReleaseRate < BIOLOGICAL_MIN_FREQUENCY {
		result.AddError(fmt.Sprintf("max release rate %.2f Hz is below the biological minimum %.2f Hz",
			vconfig.MaxReleaseRate, BIOLOGICAL_MIN_FREQUENCY))
	}
	if vconfig.MaxReleaseRate > BIOLOGICAL_MAX_FREQUENCY {
		result.AddWarning(fmt.Sprintf("max release rate %.2f Hz is extremely high (max observed: %.2f Hz)",
			vconfig.MaxReleaseRate, BIOLOGICAL_MAX_FREQUENCY))
	}

	if vconfig.BaselineProbability < 0.0 || vconfig.BaselineProbability > 1.0 {
		result.AddError(fmt.Sprintf("baseline release probability %.3f must be between 0.0 and 1.0",
			vconfig.BaselineProbability))
	}

	if vconfig.ReadyPoolSize <= 0 {
		result.AddError("ready pool size must be a positive integer")
	}
	if vconfig.RecyclingPoolSize <= 0 {
		result.AddError("recycling pool size must be a positive integer")
	}
	if vconfig.ReservePoolSize < 0 {
		result.AddError("reserve pool size cannot be negative")
	}

	if vconfig.FastRecyclingTime <= 0 || vconfig.SlowRecyclingTime <= 0 || vconfig.RefillTime <= 0 {
		result.AddError("all recycling and refill times must be positive durations")
	}
	if vconfig.SlowRecyclingTime <= vconfig.FastRecyclingTime {
		result.AddError("slow recycling time must be greater than fast recycling time")
	}

	if vconfig.ReadyPoolSize > 100 {
		result.AddWarning(fmt.Sprintf("ready pool size %d is unusually large for a single synapse (typical range: 5-20)",
			vconfig.ReadyPoolSize))
	}
}

// validateNeurotransmitterCompatibility checks for logical consistency between the
// assigned neurotransmitter and other synaptic properties like weight and delay.
func validateNeurotransmitterCompatibility(config SynapseConfig, result *ValidationResult) {
	nt := config.NeurotransmitterType
	weight := config.InitialWeight

	if nt.IsExcitatory() && weight < 0 {
		result.AddWarning(fmt.Sprintf("negative weight (%.2f) used with an excitatory neurotransmitter (%s)", weight, nt.String()))
	}

	if nt.IsInhibitory() && weight > 0 {
		result.AddWarning(fmt.Sprintf("positive weight (%.2f) used with an inhibitory neurotransmitter (%s); consider using negative weights for inhibition", weight, nt.String()))
	}

	if nt.IsModulatory() && config.BaseSynapticDelay < 10*time.Millisecond {
		result.AddWarning(fmt.Sprintf("unusually fast synaptic delay (%v) for a neuromodulatory synapse (%s)", config.BaseSynapticDelay, nt.String()))
	}

	synapseTypeLower := strings.ToLower(config.SynapseType)
	ntNameLower := strings.ToLower(nt.String())

	if !strings.Contains(synapseTypeLower, ntNameLower) {
		result.AddWarning(fmt.Sprintf("synapse type '%s' may not match the specified neurotransmitter '%s'", config.SynapseType, nt.String()))
	}
}

// =================================================================================
// CONFIGURATION HELPERS AND UTILITIES
// =================================================================================

// CloneConfig creates a deep copy of a synapse configuration, ensuring that
// modifications to the new config do not affect the original.
func CloneConfig(config SynapseConfig) SynapseConfig {
	cloned := config

	// Deep copy the metadata map to prevent shared state.
	if config.Metadata != nil {
		cloned.Metadata = make(map[string]interface{}, len(config.Metadata))
		for k, v := range config.Metadata {
			cloned.Metadata[k] = v // Note: This is a shallow copy of map values.
		}
	}

	return cloned
}

// ApplyMetadata allows for overlaying new or updated metadata onto an existing configuration.
// It returns a new config with the merged metadata.
func (c SynapseConfig) ApplyMetadata(meta map[string]interface{}) SynapseConfig {
	newConfig := CloneConfig(c)
	if newConfig.Metadata == nil {
		newConfig.Metadata = make(map[string]interface{})
	}

	for key, value := range meta {
		newConfig.Metadata[key] = value
	}
	return newConfig
}
