// types/configs.go
package types

import "time"

// =================================================================================
// PLASTICITY CONFIGURATION STRUCTURES
// =================================================================================

// PlasticityConfig defines spike-timing dependent plasticity parameters
// Used to configure how synapses learn and adapt over time
type PlasticityConfig struct {
	Enabled        bool          `json:"enabled"`         // Whether STDP is active
	LearningRate   float64       `json:"learning_rate"`   // Rate of weight changes (0.001-0.1)
	TimeConstant   time.Duration `json:"time_constant"`   // STDP time window decay (10-50ms)
	WindowSize     time.Duration `json:"window_size"`     // Maximum timing window for plasticity (50-200ms)
	MinWeight      float64       `json:"min_weight"`      // Minimum allowed weight (prevents elimination)
	MaxWeight      float64       `json:"max_weight"`      // Maximum allowed weight (prevents saturation)
	AsymmetryRatio float64       `json:"asymmetry_ratio"` // LTP/LTD asymmetry factor (typically 1.0-1.5)
}

// PruningConfig defines structural plasticity parameters
// Used to configure when and how synapses are eliminated
type PruningConfig struct {
	Enabled             bool          `json:"enabled"`              // Whether pruning is active
	WeightThreshold     float64       `json:"weight_threshold"`     // Minimum weight to avoid pruning
	InactivityThreshold time.Duration `json:"inactivity_threshold"` // Maximum inactivity before pruning
}

// HomeostasisConfig defines homeostatic plasticity parameters
// Used to configure activity-dependent scaling mechanisms
type HomeostasisConfig struct {
	Enabled          bool          `json:"enabled"`            // Whether homeostasis is active
	TargetActivity   float64       `json:"target_activity"`    // Desired activity level
	ScalingRate      float64       `json:"scaling_rate"`       // Rate of homeostatic adjustment
	TimeWindow       time.Duration `json:"time_window"`        // Window for activity measurement
	MinScalingFactor float64       `json:"min_scaling_factor"` // Minimum scaling multiplier
	MaxScalingFactor float64       `json:"max_scaling_factor"` // Maximum scaling multiplier
}

// =================================================================================
// NEURON CONFIGURATION STRUCTURES
// =================================================================================

// NeuronConfig defines parameters for neuron creation
// Used by factory functions to create neurons with specific properties
type NeuronConfig struct {
	// Core neural parameters
	Threshold        float64       `json:"threshold"`         // Action potential threshold
	DecayRate        float64       `json:"decay_rate"`        // Membrane potential decay rate
	RefractoryPeriod time.Duration `json:"refractory_period"` // Refractory period duration
	FireFactor       float64       `json:"fire_factor"`       // Firing strength multiplier

	// Homeostatic parameters
	TargetFiringRate    float64 `json:"target_firing_rate"`   // Desired firing rate (Hz)
	HomeostasisStrength float64 `json:"homeostasis_strength"` // Homeostatic adjustment strength

	// Spatial positioning
	Position Position3D `json:"position"` // 3D spatial location

	// Chemical signaling
	Receptors       []LigandType `json:"receptors"`        // Neurotransmitter receptors expressed
	ReleasedLigands []LigandType `json:"released_ligands"` // Neurotransmitters this neuron releases

	// Electrical signaling
	SignalTypes []SignalType `json:"signal_types"` // Electrical signal types processed

	// Classification and metadata
	NeuronType string                 `json:"neuron_type"` // Neuron type classification
	Metadata   map[string]interface{} `json:"metadata"`    // Additional properties
}

// =================================================================================
// SYNAPSE CONFIGURATION STRUCTURES
// =================================================================================

// SynapseConfig defines parameters for synapse creation
// Used by factory functions to create synapses with specific properties
type SynapseConfig struct {
	// Connection topology
	PresynapticID  string `json:"presynaptic_id"`  // Source neuron ID
	PostsynapticID string `json:"postsynaptic_id"` // Target neuron ID

	// Synaptic properties
	InitialWeight float64       `json:"initial_weight"` // Starting synaptic weight
	Delay         time.Duration `json:"delay"`          // Transmission delay

	// Chemical signaling
	LigandType LigandType `json:"ligand_type"` // Neurotransmitter type

	// Plasticity configuration
	PlasticityEnabled bool             `json:"plasticity_enabled"` // Whether plasticity is active
	PlasticityConfig  PlasticityConfig `json:"plasticity_config"`  // Plasticity parameters
	PruningConfig     PruningConfig    `json:"pruning_config"`     // Pruning parameters

	// Spatial positioning
	Position Position3D `json:"position"` // 3D spatial location

	// Classification and metadata
	SynapseType string                 `json:"synapse_type"` // Synapse type classification
	Metadata    map[string]interface{} `json:"metadata"`     // Additional properties
}

// SynapseCreationConfig for neuron-initiated synapse creation
// Used when neurons request new synaptic connections
type SynapseCreationConfig struct {
	SourceNeuronID string        `json:"source_neuron_id"` // Pre-synaptic neuron ID
	TargetNeuronID string        `json:"target_neuron_id"` // Post-synaptic neuron ID
	InitialWeight  float64       `json:"initial_weight"`   // Starting weight
	SynapseType    string        `json:"synapse_type"`     // Type of synapse
	PlasticityType string        `json:"plasticity_type"`  // Type of plasticity
	Delay          time.Duration `json:"delay"`            // Transmission delay
	Position       Position3D    `json:"position"`         // 3D spatial location
}

// =================================================================================
// NETWORK CONFIGURATION STRUCTURES
// =================================================================================

// NetworkConfig defines parameters for entire network creation
// Used for configuring large-scale neural network properties
type NetworkConfig struct {
	// Network topology
	NeuronCount       int     `json:"neuron_count"`       // Total number of neurons
	ConnectionDensity float64 `json:"connection_density"` // Fraction of possible connections

	// Spatial organization
	Dimensions   Position3D `json:"dimensions"`    // Network spatial extent
	LayerCount   int        `json:"layer_count"`   // Number of distinct layers
	LayerSpacing float64    `json:"layer_spacing"` // Distance between layers

	// Default component parameters
	DefaultNeuronConfig  NeuronConfig  `json:"default_neuron_config"`  // Default neuron settings
	DefaultSynapseConfig SynapseConfig `json:"default_synapse_config"` // Default synapse settings

	// Global properties
	TimeStep       time.Duration `json:"time_step"`       // Simulation time step
	UpdateInterval time.Duration `json:"update_interval"` // Component update frequency
	MaxComponents  int           `json:"max_components"`  // Maximum total components

	// Feature flags
	ChemicalEnabled   bool `json:"chemical_enabled"`   // Enable chemical signaling
	SpatialEnabled    bool `json:"spatial_enabled"`    // Enable spatial processing
	PlasticityEnabled bool `json:"plasticity_enabled"` // Enable plasticity

	// Metadata
	NetworkType string                 `json:"network_type"` // Network classification
	Metadata    map[string]interface{} `json:"metadata"`     // Additional properties
}

// =================================================================================
// COMPONENT FILTERING AND QUERY STRUCTURES
// =================================================================================

// SynapseCriteria for filtering synapse queries
// Used by ListSynapses callbacks to find specific synapses
type SynapseCriteria struct {
	Direction     *SynapseDirection `json:"direction,omitempty"`      // Filter by incoming/outgoing/both
	SourceID      *string           `json:"source_id,omitempty"`      // Filter by source neuron ID
	TargetID      *string           `json:"target_id,omitempty"`      // Filter by target neuron ID
	WeightRange   *WeightRange      `json:"weight_range,omitempty"`   // Filter by weight bounds
	ActivitySince *time.Time        `json:"activity_since,omitempty"` // Filter by recent activity
	SynapseType   *string           `json:"synapse_type,omitempty"`   // Filter by synapse type
}

// SynapseDirection specifies synapse directionality relative to neuron
type SynapseDirection int

const (
	SynapseIncoming SynapseDirection = iota // Synapses targeting this neuron
	SynapseOutgoing                         // Synapses originating from this neuron
	SynapseBoth                             // All synapses connected to this neuron
)

// WeightRange defines bounds for synapse weight filtering
type WeightRange struct {
	Min float64 `json:"min"` // Minimum weight (inclusive)
	Max float64 `json:"max"` // Maximum weight (inclusive)
}

// ComponentCriteria for filtering component queries
// Used for spatial and property-based component searches
type ComponentCriteria struct {
	ComponentType *ComponentType  `json:"component_type,omitempty"` // Filter by component type
	State         *ComponentState `json:"state,omitempty"`          // Filter by component state
	Position      *Position3D     `json:"position,omitempty"`       // Center point for spatial search
	Radius        float64         `json:"radius,omitempty"`         // Search radius from position
	ActivityLevel *float64        `json:"activity_level,omitempty"` // Minimum activity level
}

// =================================================================================
// ANALYSIS AND MONITORING CONFIGURATIONS
// =================================================================================

// MonitoringConfig defines parameters for component monitoring
// Used to configure health monitoring and analysis systems
type MonitoringConfig struct {
	Enabled          bool             `json:"enabled"`           // Whether monitoring is active
	UpdateInterval   time.Duration    `json:"update_interval"`   // How often to collect metrics
	HistorySize      int              `json:"history_size"`      // Number of historical samples to keep
	HealthThresholds HealthThresholds `json:"health_thresholds"` // Thresholds for health assessment
	AlertEnabled     bool             `json:"alert_enabled"`     // Whether to generate alerts
	ReportingEnabled bool             `json:"reporting_enabled"` // Whether to generate reports
}

// HealthThresholds defines thresholds for component health assessment
type HealthThresholds struct {
	MinActivityLevel    float64       `json:"min_activity_level"`    // Minimum healthy activity
	MaxActivityLevel    float64       `json:"max_activity_level"`    // Maximum healthy activity
	MaxInactivityPeriod time.Duration `json:"max_inactivity_period"` // Maximum time without activity
	MinConnectionCount  int           `json:"min_connection_count"`  // Minimum healthy connections
	MaxConnectionCount  int           `json:"max_connection_count"`  // Maximum healthy connections
}

// =================================================================================
// COMPONENT TYPE AND STATE DEFINITIONS
// =================================================================================

// ComponentType represents the distinct categories of neural components
type ComponentType int

const (
	TypeNeuron        ComponentType = iota // Neural cell (processes and transmits signals)
	TypeSynapse                            // Synaptic connection (chemical/electrical junction)
	TypeGlialCell                          // Non-neuronal support cell (astrocyte, oligodendrocyte)
	TypeMicrogliaCell                      // Immune cell of the brain (surveillance, cleanup)
	TypeEpendymalCell                      // Cell lining brain ventricles (CSF barrier)
)

// String provides a human-readable representation for Componenttypes.
func (ct ComponentType) String() string {
	switch ct {
	case TypeNeuron:
		return "Neuron"
	case TypeSynapse:
		return "Synapse"
	case TypeGlialCell:
		return "GlialCell"
	case TypeMicrogliaCell:
		return "MicrogliaCell"
	case TypeEpendymalCell:
		return "EpendymalCell"
	default:
		return "Unknown"
	}
}

// ComponentState represents the operational states a component can be in
type ComponentState int

const (
	StateActive       ComponentState = iota // Fully operational and participating
	StateInactive                           // Temporarily disabled but can be reactivated
	StateShuttingDown                       // Gracefully terminating operations
	StateStopped                            // Fully ceased operations
	StateDeveloping                         // Undergoing developmental/maturation phase
	StateDying                              // Process of decay or programmed cell death
	StateDamaged                            // Incurred damage, may be impaired
	StateMaintenance                        // Undergoing internal maintenance/repair
	StateHibernating                        // Low-activity state to conserve resources
)

// String provides a human-readable representation for ComponentState.
func (cs ComponentState) String() string {
	switch cs {
	case StateActive:
		return "Active"
	case StateInactive:
		return "Inactive"
	case StateShuttingDown:
		return "ShuttingDown"
	case StateStopped:
		return "Stopped"
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

type OutputCallback struct {
	TransmitMessage func(msg NeuralSignal) error
	GetWeight       func() float64
	GetDelay        func() time.Duration
	GetTargetID     func() string
}
