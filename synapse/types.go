/*
=================================================================================
SYNAPSE TYPES - CORE DATA STRUCTURES AND BIOLOGICAL EVENTS
=================================================================================

This file defines all the core types, data structures, and event types used
throughout the synapse system. These types support biological integration,
matrix coordination, and vesicle dynamics while maintaining clean separation
of concerns.

DESIGN PRINCIPLES:
1. BIOLOGICAL ACCURACY: All types model real neural structures and processes
2. ZERO DEPENDENCIES: No imports from matrix or neuron packages
3. EXTENSIBILITY: Easy to add new neurotransmitter types and events
4. PERFORMANCE: Efficient serialization and memory usage
5. INTEGRATION: Support for callback patterns and factory creation

ORGANIZATION:
- Position and Spatial Types
- Neurotransmitter and Chemical Types
- Message and Communication Types
- Event and Activity Types
- Configuration and State Types
- Biological Process Types
=================================================================================
*/

package synapse

import (
	"fmt"
	"math"
	"time"
)

// =================================================================================
// SPATIAL AND POSITIONING TYPES
// =================================================================================

// Position3D represents 3D spatial coordinates in neural tissue
// Units: micrometers (μm) - standard scale for neural structures
// Biological context: Used for calculating axonal delays and chemical diffusion
type Position3D struct {
	X float64 `json:"x"` // X coordinate in micrometers
	Y float64 `json:"y"` // Y coordinate in micrometers
	Z float64 `json:"z"` // Z coordinate in micrometers
}

// Distance calculates Euclidean distance between two 3D positions
func (p Position3D) Distance(other Position3D) float64 {
	dx := p.X - other.X
	dy := p.Y - other.Y
	dz := p.Z - other.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// IsValid checks if position coordinates are within reasonable biological bounds
func (p Position3D) IsValid() bool {
	const maxCoord = 1e6 // 1 meter - larger than any reasonable neural structure
	return !math.IsNaN(p.X) && !math.IsInf(p.X, 0) && math.Abs(p.X) < maxCoord &&
		!math.IsNaN(p.Y) && !math.IsInf(p.Y, 0) && math.Abs(p.Y) < maxCoord &&
		!math.IsNaN(p.Z) && !math.IsInf(p.Z, 0) && math.Abs(p.Z) < maxCoord
}

// =================================================================================
// NEUROTRANSMITTER AND CHEMICAL TYPES
// =================================================================================

// LigandType represents different neurotransmitter types
// Biological basis: Major neurotransmitter systems in the mammalian brain
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

// =================================================================================
// MESSAGE AND COMMUNICATION TYPES
// =================================================================================

// SynapseMessage represents a processed synaptic signal with full biological metadata
// This is the primary communication structure between synapses and neurons
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

// IsValid performs basic validation on the message
func (sm SynapseMessage) IsValid() bool {
	return !math.IsNaN(sm.Value) && !math.IsInf(sm.Value, 0) &&
		!math.IsNaN(sm.OriginalValue) && !math.IsInf(sm.OriginalValue, 0) &&
		!math.IsNaN(sm.EffectiveWeight) && !math.IsInf(sm.EffectiveWeight, 0) &&
		sm.SourceID != "" && sm.TargetID != "" && sm.SynapseID != ""
}

// GetTotalDelay returns the sum of synaptic and spatial delays
func (sm SynapseMessage) GetTotalDelay() time.Duration {
	return sm.SynapticDelay + sm.SpatialDelay
}

// =================================================================================
// EVENT AND ACTIVITY TYPES
// =================================================================================

// PlasticityEventType categorizes different types of plasticity events
type PlasticityEventType int

const (
	PlasticitySTDP        PlasticityEventType = iota // Spike-timing dependent plasticity
	PlasticityBCM                                    // Bienenstock-Cooper-Munro rule
	PlasticityOja                                    // Oja's learning rule
	PlasticityHomeostatic                            // Homeostatic scaling
	PlasticityMetaplastic                            // Metaplasticity (plasticity of plasticity)
	PlasticityReward                                 // Reward-based plasticity
	PlasticityPunishment                             // Punishment-based plasticity
	PlasticityHebbian                                // Basic Hebbian learning
	PlasticityAntiHebbian                            // Anti-Hebbian learning
)

// String returns human-readable name for plasticity type
func (pet PlasticityEventType) String() string {
	switch pet {
	case PlasticitySTDP:
		return "STDP"
	case PlasticityBCM:
		return "BCM"
	case PlasticityOja:
		return "Oja"
	case PlasticityHomeostatic:
		return "Homeostatic"
	case PlasticityMetaplastic:
		return "Metaplastic"
	case PlasticityReward:
		return "Reward"
	case PlasticityPunishment:
		return "Punishment"
	case PlasticityHebbian:
		return "Hebbian"
	case PlasticityAntiHebbian:
		return "AntiHebbian"
	default:
		return "Unknown"
	}
}

// PlasticityEvent represents STDP and other plasticity mechanisms
type PlasticityEvent struct {
	SynapseID    string                 `json:"synapse_id"`    // Synapse undergoing plasticity
	EventType    PlasticityEventType    `json:"event_type"`    // Type of plasticity event
	Timestamp    time.Time              `json:"timestamp"`     // When event occurred
	PreTime      time.Time              `json:"pre_time"`      // Pre-synaptic spike time
	PostTime     time.Time              `json:"post_time"`     // Post-synaptic spike time
	DeltaT       time.Duration          `json:"delta_t"`       // Spike timing difference
	WeightBefore float64                `json:"weight_before"` // Weight before change
	WeightAfter  float64                `json:"weight_after"`  // Weight after change
	WeightChange float64                `json:"weight_change"` // Magnitude of change
	Strength     float64                `json:"strength"`      // Event strength/magnitude
	LearningRate float64                `json:"learning_rate"` // Learning rate used
	Context      map[string]interface{} `json:"context"`       // Additional context
}

// GetTimingDirection returns the direction of spike timing (causal vs anti-causal)
func (pe PlasticityEvent) GetTimingDirection() string {
	if pe.DeltaT < 0 {
		return "causal" // Pre before post - strengthening
	} else if pe.DeltaT > 0 {
		return "anti_causal" // Pre after post - weakening
	} else {
		return "simultaneous"
	}
}

// IsValid performs basic validation on the plasticity event
func (pe PlasticityEvent) IsValid() bool {
	return pe.SynapseID != "" &&
		!math.IsNaN(pe.WeightBefore) && !math.IsInf(pe.WeightBefore, 0) &&
		!math.IsNaN(pe.WeightAfter) && !math.IsInf(pe.WeightAfter, 0) &&
		!math.IsNaN(pe.WeightChange) && !math.IsInf(pe.WeightChange, 0)
}

// SynapticActivity reports synaptic transmission activity for monitoring
type SynapticActivity struct {
	SynapseID         string           `json:"synapse_id"`         // Synapse identifier
	Timestamp         time.Time        `json:"timestamp"`          // When activity occurred
	MessageValue      float64          `json:"message_value"`      // Signal strength
	CurrentWeight     float64          `json:"current_weight"`     // Current synaptic weight
	ActivityType      string           `json:"activity_type"`      // "transmission", "plasticity", "pruning"
	VesicleState      VesiclePoolState `json:"vesicle_state"`      // Current vesicle pool state
	MetabolicCost     float64          `json:"metabolic_cost"`     // Energy cost of activity
	SuccessfulRelease bool             `json:"successful_release"` // Whether release succeeded
	CalciumLevel      float64          `json:"calcium_level"`      // Calcium concentration
	ErrorMessage      string           `json:"error_message"`      // Error description if failed
}

// IsValid performs basic validation on synaptic activity
func (sa SynapticActivity) IsValid() bool {
	return sa.SynapseID != "" &&
		!math.IsNaN(sa.MessageValue) && !math.IsInf(sa.MessageValue, 0) &&
		!math.IsNaN(sa.CurrentWeight) && !math.IsInf(sa.CurrentWeight, 0) &&
		!math.IsNaN(sa.MetabolicCost) && !math.IsInf(sa.MetabolicCost, 0)
}

// =================================================================================
// VESICLE SYSTEM TYPES
// =================================================================================

// VesiclePoolState captures the state of all vesicle pools at a given moment
// Provides detailed information about vesicle availability for biological realism
type VesiclePoolState struct {
	ReadyVesicles     int       `json:"ready_vesicles"`     // Immediately available
	RecyclingVesicles int       `json:"recycling_vesicles"` // In recycling process
	ReserveVesicles   int       `json:"reserve_vesicles"`   // Long-term reserve
	TotalVesicles     int       `json:"total_vesicles"`     // Total pool size
	DepletionLevel    float64   `json:"depletion_level"`    // Pool depletion (0.0-1.0)
	FatigueLevel      float64   `json:"fatigue_level"`      // Synaptic fatigue (0.0-1.0)
	CalciumLevel      float64   `json:"calcium_level"`      // Current calcium concentration
	ReleaseRate       float64   `json:"release_rate"`       // Current release rate (Hz)
	LastUpdate        time.Time `json:"last_update"`        // Last state update
}

// IsHealthy returns true if vesicle pools are in good condition
func (vps VesiclePoolState) IsHealthy() bool {
	return vps.DepletionLevel < 0.8 && // Less than 80% depleted
		vps.FatigueLevel < 0.6 && // Less than 60% fatigued
		vps.ReadyVesicles > 0 // Some vesicles available
}

// GetEffectiveAvailability returns effective vesicle availability (0.0-1.0)
func (vps VesiclePoolState) GetEffectiveAvailability() float64 {
	if vps.TotalVesicles == 0 {
		return 0.0
	}

	availability := float64(vps.ReadyVesicles) / float64(vps.TotalVesicles)
	availability *= (1.0 - vps.FatigueLevel) // Reduce by fatigue

	return math.Max(0.0, math.Min(1.0, availability))
}

// =================================================================================
// CONFIGURATION AND STATE TYPES
// =================================================================================

// ComponentState represents the functional state of a neural component
type ComponentState int

const (
	StateActive       ComponentState = iota // Fully functional and active
	StateInactive                           // Present but not functioning
	StateDormant                            // Temporarily inactive
	StateShuttingDown                       // In process of deactivation
	StateDamaged                            // Damaged but potentially recoverable
	StateFailed                             // Permanently non-functional
)

// String returns human-readable state name
func (cs ComponentState) String() string {
	switch cs {
	case StateActive:
		return "Active"
	case StateInactive:
		return "Inactive"
	case StateDormant:
		return "Dormant"
	case StateShuttingDown:
		return "ShuttingDown"
	case StateDamaged:
		return "Damaged"
	case StateFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// SynapticActivityInfo provides comprehensive synapse state information
type SynapticActivityInfo struct {
	// === BASIC STATE ===
	SynapseID     string    `json:"synapse_id"`     // Synapse identifier
	CurrentWeight float64   `json:"current_weight"` // Current synaptic strength
	IsActive      bool      `json:"is_active"`      // Whether synapse is functional
	LastUpdate    time.Time `json:"last_update"`    // Last information update

	// === TRANSMISSION METRICS ===
	TotalTransmissions      int64     `json:"total_transmissions"`      // Total attempts
	SuccessfulTransmissions int64     `json:"successful_transmissions"` // Successful releases
	LastTransmission        time.Time `json:"last_transmission"`        // Most recent transmission
	TransmissionRate        float64   `json:"transmission_rate"`        // Current rate (Hz)

	// === PLASTICITY METRICS ===
	TotalPlasticityEvents int64     `json:"total_plasticity_events"` // Total plasticity events
	LastPlasticityEvent   time.Time `json:"last_plasticity_event"`   // Most recent plasticity
	WeightChangeRate      float64   `json:"weight_change_rate"`      // Rate of weight change

	// === VESICLE STATE ===
	VesicleState VesiclePoolState `json:"vesicle_state"` // Current vesicle pool state

	// === BIOLOGICAL METRICS ===
	MetabolicActivity    float64    `json:"metabolic_activity"`    // Energy consumption
	NeurotransmitterType LigandType `json:"neurotransmitter_type"` // NT type
	EffectiveStrength    float64    `json:"effective_strength"`    // Functional strength

	// === SPATIAL PROPERTIES ===
	Position     Position3D    `json:"position"`      // 3D location
	AverageDelay time.Duration `json:"average_delay"` // Average transmission delay

	// === HEALTH INDICATORS ===
	HealthScore     float64   `json:"health_score"`     // Overall health (0.0-1.0)
	PruningRisk     float64   `json:"pruning_risk"`     // Risk of elimination (0.0-1.0)
	ActivityLevel   float64   `json:"activity_level"`   // Recent activity level
	LastMaintenance time.Time `json:"last_maintenance"` // Last health check
}

// IsHealthy returns true if synapse is in good health
func (sai SynapticActivityInfo) IsHealthy() bool {
	return sai.HealthScore > 0.7 && // Good health score
		sai.PruningRisk < 0.3 && // Low pruning risk
		sai.IsActive && // Currently active
		sai.VesicleState.IsHealthy() // Vesicles healthy
}

// GetReliability calculates transmission reliability
func (sai SynapticActivityInfo) GetReliability() float64 {
	if sai.TotalTransmissions == 0 {
		return 1.0 // No data, assume perfect
	}
	return float64(sai.SuccessfulTransmissions) / float64(sai.TotalTransmissions)
}

// =================================================================================
// PLASTICITY CONFIGURATION TYPES
// =================================================================================

// PlasticityAdjustment contains information for synaptic plasticity updates
type PlasticityAdjustment struct {
	DeltaT         time.Duration          `json:"delta_t"`          // Spike timing difference (t_pre - t_post)
	PreSpikeTrain  []time.Time            `json:"pre_spike_train"`  // Recent pre-synaptic spike times
	PostSpikeTrain []time.Time            `json:"post_spike_train"` // Recent post-synaptic spike times
	Context        map[string]interface{} `json:"context"`          // Additional context (reward, attention, etc.)
	LearningRate   float64                `json:"learning_rate"`    // Override learning rate (0 = use default)
	ForceUpdate    bool                   `json:"force_update"`     // Force update regardless of timing window
	Neuromodulator LigandType             `json:"neuromodulator"`   // Neuromodulator present during learning
}

// IsValid performs basic validation on plasticity adjustment
func (pa PlasticityAdjustment) IsValid() bool {
	return !math.IsNaN(pa.LearningRate) && !math.IsInf(pa.LearningRate, 0) &&
		pa.LearningRate >= 0.0 && pa.LearningRate <= 1.0
}

// STDPConfig defines spike-timing dependent plasticity parameters
type STDPConfig struct {
	Enabled        bool          `json:"enabled"`         // Master switch for STDP
	LearningRate   float64       `json:"learning_rate"`   // Base learning rate
	TimeConstant   time.Duration `json:"time_constant"`   // Exponential decay τ
	WindowSize     time.Duration `json:"window_size"`     // Max timing window
	MinWeight      float64       `json:"min_weight"`      // Lower bound
	MaxWeight      float64       `json:"max_weight"`      // Upper bound
	AsymmetryRatio float64       `json:"asymmetry_ratio"` // LTD/LTP ratio

	// Advanced STDP features
	FrequencyDependent     bool    `json:"frequency_dependent"`     // Enable frequency-dependent STDP
	MetaplasticityRate     float64 `json:"metaplasticity_rate"`     // Rate of metaplastic changes
	CooperativityThreshold int     `json:"cooperativity_threshold"` // Threshold for cooperative plasticity
}

// IsValid performs validation on STDP configuration
func (sc STDPConfig) IsValid() bool {
	return sc.LearningRate >= 0.0 && sc.LearningRate <= 1.0 &&
		sc.TimeConstant > 0 &&
		sc.WindowSize > 0 &&
		sc.MinWeight >= 0.0 &&
		sc.MaxWeight > sc.MinWeight &&
		sc.AsymmetryRatio > 0.0 &&
		sc.MetaplasticityRate >= 0.0 && sc.MetaplasticityRate <= 1.0 &&
		sc.CooperativityThreshold >= 0
}

// PruningConfig defines structural plasticity parameters
type PruningConfig struct {
	Enabled             bool          `json:"enabled"`              // Enable pruning
	WeightThreshold     float64       `json:"weight_threshold"`     // Weight below which pruning considered
	InactivityThreshold time.Duration `json:"inactivity_threshold"` // Time without activity
	MetabolicThreshold  float64       `json:"metabolic_threshold"`  // Metabolic cost threshold
	ProtectionPeriod    time.Duration `json:"protection_period"`    // Grace period after creation
	PruningProbability  float64       `json:"pruning_probability"`  // Stochastic pruning probability
}

// IsValid performs validation on pruning configuration
func (pc PruningConfig) IsValid() bool {
	return pc.WeightThreshold >= 0.0 &&
		pc.InactivityThreshold > 0 &&
		pc.MetabolicThreshold >= 0.0 &&
		pc.ProtectionPeriod >= 0 &&
		pc.PruningProbability >= 0.0 && pc.PruningProbability <= 1.0
}

// =================================================================================
// BIOLOGICAL PROCESS TYPES
// =================================================================================

// BiologicalTimescale represents different temporal scales in biology
type BiologicalTimescale int

const (
	TimescaleMicrosecond BiologicalTimescale = iota // Ion channel kinetics
	TimescaleMillisecond                            // Action potentials, fast synaptic transmission
	TimescaleSecond                                 // Slow synaptic transmission, vesicle recycling
	TimescaleMinute                                 // Short-term plasticity, protein synthesis
	TimescaleHour                                   // Long-term potentiation, structural changes
	TimescaleDay                                    // Memory consolidation, growth
)

// GetDuration returns typical duration for biological timescale
func (bt BiologicalTimescale) GetDuration() time.Duration {
	switch bt {
	case TimescaleMicrosecond:
		return time.Microsecond
	case TimescaleMillisecond:
		return time.Millisecond
	case TimescaleSecond:
		return time.Second
	case TimescaleMinute:
		return time.Minute
	case TimescaleHour:
		return time.Hour
	case TimescaleDay:
		return 24 * time.Hour
	default:
		return time.Millisecond
	}
}

// NeuralEventType categorizes different types of neural events
type NeuralEventType int

const (
	EventTransmission NeuralEventType = iota // Synaptic transmission
	EventPlasticity                          // Plasticity change
	EventPruning                             // Structural pruning
	EventGrowth                              // Synapse formation
	EventDamage                              // Damage or dysfunction
	EventRecovery                            // Recovery from damage
	EventMaintenance                         // Routine maintenance
)

// String returns human-readable event type name
func (net NeuralEventType) String() string {
	switch net {
	case EventTransmission:
		return "Transmission"
	case EventPlasticity:
		return "Plasticity"
	case EventPruning:
		return "Pruning"
	case EventGrowth:
		return "Growth"
	case EventDamage:
		return "Damage"
	case EventRecovery:
		return "Recovery"
	case EventMaintenance:
		return "Maintenance"
	default:
		return "Unknown"
	}
}

// =================================================================================
// ERROR AND VALIDATION TYPES
// =================================================================================

// SynapseError represents errors specific to synaptic operations
type SynapseError struct {
	Type      string                 `json:"type"`       // Error category
	Message   string                 `json:"message"`    // Human-readable message
	SynapseID string                 `json:"synapse_id"` // Synapse where error occurred
	Timestamp time.Time              `json:"timestamp"`  // When error occurred
	Context   map[string]interface{} `json:"context"`    // Additional context
}

// Error implements the error interface
func (se SynapseError) Error() string {
	return fmt.Sprintf("[%s] %s (synapse: %s)", se.Type, se.Message, se.SynapseID)
}

// ValidationResult contains validation information
type ValidationResult struct {
	IsValid  bool     `json:"is_valid"` // Whether validation passed
	Errors   []string `json:"errors"`   // List of validation errors
	Warnings []string `json:"warnings"` // List of warnings
}

// AddError adds an error to the validation result
func (vr *ValidationResult) AddError(message string) {
	vr.IsValid = false
	vr.Errors = append(vr.Errors, message)
}

// AddWarning adds a warning to the validation result
func (vr *ValidationResult) AddWarning(message string) {
	vr.Warnings = append(vr.Warnings, message)
}

// =================================================================================
// FACTORY AND CALLBACK TYPES
// =================================================================================

// SynapseFactory creates synapses with specific biological properties
// Used by the matrix factory system for dependency injection
type SynapseFactory func(id string, config SynapseConfig, callbacks SynapseCallbacks) (SynapticProcessor, error)

// CallbackFunction represents a generic callback function signature
type CallbackFunction func(args ...interface{}) error

// CallbackRegistry manages callback function registration
type CallbackRegistry map[string]CallbackFunction

// Register adds a callback function to the registry
func (cr CallbackRegistry) Register(name string, fn CallbackFunction) {
	cr[name] = fn
}

// Call executes a registered callback function
func (cr CallbackRegistry) Call(name string, args ...interface{}) error {
	if fn, exists := cr[name]; exists {
		return fn(args...)
	}
	return fmt.Errorf("callback %s not found", name)
}

// =================================================================================
// UTILITY FUNCTIONS FOR TYPES
// =================================================================================

// NewPosition3D creates a new 3D position with validation
func NewPosition3D(x, y, z float64) Position3D {
	pos := Position3D{X: x, Y: y, Z: z}
	if !pos.IsValid() {
		// Return origin if invalid coordinates provided
		return Position3D{X: 0, Y: 0, Z: 0}
	}
	return pos
}

// NewValidationResult creates a new validation result
func NewValidationResult() ValidationResult {
	return ValidationResult{
		IsValid:  true,
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}
}

// NewSynapseError creates a new synapse-specific error
func NewSynapseError(errorType, message, synapseID string) SynapseError {
	return SynapseError{
		Type:      errorType,
		Message:   message,
		SynapseID: synapseID,
		Timestamp: time.Now(),
		Context:   make(map[string]interface{}),
	}
}
