// types/events.go
package types

import "time"

// =================================================================================
// PLASTICITY EVENT STRUCTURES
// =================================================================================

// PlasticityEventType categorizes different plasticity mechanisms
type PlasticityEventType string

const (
	PlasticitySTDP        PlasticityEventType = "stdp"        // Spike-timing dependent plasticity
	PlasticityBCM         PlasticityEventType = "bcm"         // Bienenstock-Cooper-Munro rule
	PlasticityOja         PlasticityEventType = "oja"         // Oja's learning rule
	PlasticityHomeostatic PlasticityEventType = "homeostatic" // Homeostatic scaling
	PlasticityMetaplastic PlasticityEventType = "metaplastic" // Metaplasticity
	PlasticityLTD         PlasticityEventType = "ltd"         // Long-term depression
	PlasticityLTP         PlasticityEventType = "ltp"         // Long-term potentiation
	PlasticityISTDP       PlasticityEventType = "istdp"       // Inhibitory STDP
)

// PlasticityEvent represents synaptic plasticity trigger events
// Used to communicate plasticity state changes between components
type PlasticityEvent struct {
	EventType PlasticityEventType `json:"event_type"` // Type of plasticity mechanism
	Timestamp time.Time           `json:"timestamp"`  // When event occurred
	PreTime   time.Time           `json:"pre_time"`   // Pre-synaptic spike time
	PostTime  time.Time           `json:"post_time"`  // Post-synaptic spike time
	Strength  float64             `json:"strength"`   // Event strength/magnitude
	SourceID  string              `json:"source_id"`  // Component that generated event
	SynapseID string              `json:"synapse_id"` // Synapse that experienced plasticity
	DeltaT    time.Duration       `json:"delta_t"`    // Time difference (pre - post)
}

// PlasticityAdjustment for STDP and other learning mechanisms
// Used to communicate plasticity changes to synapses
type PlasticityAdjustment struct {
	DeltaT       time.Duration       `json:"delta_t"`       // Time difference for STDP (t_pre - t_post)
	WeightChange float64             `json:"weight_change"` // Direct weight modification (optional)
	LearningRate float64             `json:"learning_rate"` // Context-specific learning rate (optional)
	PostSynaptic bool                `json:"post_synaptic"` // Whether post-synaptic neuron fired
	PreSynaptic  bool                `json:"pre_synaptic"`  // Whether pre-synaptic neuron fired recently
	Timestamp    time.Time           `json:"timestamp"`     // When adjustment was generated
	EventType    PlasticityEventType `json:"event_type"`    // Type of plasticity adjustment
}

// =================================================================================
// COMPONENT LIFECYCLE EVENTS
// =================================================================================

// LifecycleEventType categorizes component lifecycle events
type LifecycleEventType string

const (
	LifecycleCreated      LifecycleEventType = "created"      // Component was created
	LifecycleStarted      LifecycleEventType = "started"      // Component began operation
	LifecycleStopped      LifecycleEventType = "stopped"      // Component ceased operation
	LifecycleDestroyed    LifecycleEventType = "destroyed"    // Component was destroyed
	LifecycleConnected    LifecycleEventType = "connected"    // Component established connection
	LifecycleDisconnected LifecycleEventType = "disconnected" // Component lost connection
	LifecycleDamaged      LifecycleEventType = "damaged"      // Component suffered damage
	LifecycleRecovered    LifecycleEventType = "recovered"    // Component recovered from damage
)

// LifecycleEvent represents component lifecycle state changes
// Used for tracking component birth, death, and state transitions
type LifecycleEvent struct {
	EventType     LifecycleEventType     `json:"event_type"`     // Type of lifecycle event
	ComponentID   string                 `json:"component_id"`   // Component experiencing event
	ComponentType ComponentType          `json:"component_type"` // Type of component
	Timestamp     time.Time              `json:"timestamp"`      // When event occurred
	OldState      ComponentState         `json:"old_state"`      // Previous component state
	NewState      ComponentState         `json:"new_state"`      // New component state
	Reason        string                 `json:"reason"`         // Reason for state change
	Metadata      map[string]interface{} `json:"metadata"`       // Additional event data
}

// =================================================================================
// NETWORK ACTIVITY EVENTS
// =================================================================================

// NetworkEventType categorizes network-wide events
type NetworkEventType string

const (
	NetworkSynchronization NetworkEventType = "synchronization" // Network synchronization event
	NetworkOscillation     NetworkEventType = "oscillation"     // Network oscillatory activity
	NetworkBurst           NetworkEventType = "burst"           // Population burst activity
	NetworkQuiescence      NetworkEventType = "quiescence"      // Network quiet period
	NetworkReorganization  NetworkEventType = "reorganization"  // Structural reorganization
	NetworkFailure         NetworkEventType = "failure"         // Network failure/dysfunction
	NetworkRecovery        NetworkEventType = "recovery"        // Network recovery event
)

// NetworkEvent represents network-wide activity and state changes
// Used for tracking global network dynamics and emergent behaviors
type NetworkEvent struct {
	EventType        NetworkEventType       `json:"event_type"`        // Type of network event
	Timestamp        time.Time              `json:"timestamp"`         // When event occurred
	Duration         time.Duration          `json:"duration"`          // Event duration
	AffectedRegion   Position3D             `json:"affected_region"`   // Spatial center of event
	AffectedRadius   float64                `json:"affected_radius"`   // Spatial extent of event
	Intensity        float64                `json:"intensity"`         // Event intensity/magnitude
	ParticipantIDs   []string               `json:"participant_ids"`   // Components involved in event
	TriggerID        string                 `json:"trigger_id"`        // Component that triggered event
	PropagationSpeed float64                `json:"propagation_speed"` // Speed of event propagation
	Metadata         map[string]interface{} `json:"metadata"`          // Additional event data
}

// =================================================================================
// CHEMICAL SIGNALING EVENTS - FIXED NAMING CONFLICTS
// =================================================================================

// ChemicalEventType categorizes chemical signaling events
type ChemicalEventType string

const (
	ChemicalReleaseEvent    ChemicalEventType = "release"    // Neurotransmitter release
	ChemicalBindingEvent    ChemicalEventType = "binding"    // Receptor binding
	ChemicalClearanceEvent  ChemicalEventType = "clearance"  // Neurotransmitter clearance
	ChemicalDiffusionEvent  ChemicalEventType = "diffusion"  // Chemical diffusion
	ChemicalGradientEvent   ChemicalEventType = "gradient"   // Concentration gradient formation
	ChemicalSaturationEvent ChemicalEventType = "saturation" // Receptor saturation
	ChemicalDepletionEvent  ChemicalEventType = "depletion"  // Neurotransmitter depletion
)

// ChemicalEvent represents chemical signaling events
type ChemicalEvent struct {
	EventType     ChemicalEventType      `json:"event_type"`    // Type of chemical event
	LigandType    LigandType             `json:"ligand_type"`   // Type of chemical involved
	SourceID      string                 `json:"source_id"`     // Component releasing/affecting chemical
	TargetID      string                 `json:"target_id"`     // Component receiving chemical (if applicable)
	Position      Position3D             `json:"position"`      // 3D location of event
	Concentration float64                `json:"concentration"` // Chemical concentration
	Volume        float64                `json:"volume"`        // Volume affected
	Timestamp     time.Time              `json:"timestamp"`     // When event occurred
	Duration      time.Duration          `json:"duration"`      // Event duration
	ReceptorType  string                 `json:"receptor_type"` // Type of receptor involved
	Metadata      map[string]interface{} `json:"metadata"`      // Additional event data
}

// =================================================================================
// HEALTH AND MONITORING EVENTS
// =================================================================================

// HealthEventType categorizes component health events
type HealthEventType string

const (
	HealthNormal   HealthEventType = "normal"   // Normal health status
	HealthWarning  HealthEventType = "warning"  // Health warning condition
	HealthCritical HealthEventType = "critical" // Critical health condition
	HealthRecovery HealthEventType = "recovery" // Health recovery event
	HealthDegraded HealthEventType = "degraded" // Performance degradation
	HealthOptimal  HealthEventType = "optimal"  // Optimal performance condition
	HealthStressed HealthEventType = "stressed" // Stress condition detected
)

// HealthEvent represents component health status changes
// Used for monitoring component wellness and triggering interventions
type HealthEvent struct {
	EventType       HealthEventType        `json:"event_type"`       // Type of health event
	ComponentID     string                 `json:"component_id"`     // Component being monitored
	ComponentType   ComponentType          `json:"component_type"`   // Type of component
	Timestamp       time.Time              `json:"timestamp"`        // When event occurred
	HealthScore     float64                `json:"health_score"`     // Overall health score (0.0-1.0)
	ActivityLevel   float64                `json:"activity_level"`   // Current activity level
	ConnectionCount int                    `json:"connection_count"` // Number of connections
	Issues          []string               `json:"issues"`           // Identified health issues
	Recommendations []string               `json:"recommendations"`  // Recommended actions
	Severity        int                    `json:"severity"`         // Event severity (1-10)
	Metadata        map[string]interface{} `json:"metadata"`         // Additional health data
}

// =================================================================================
// ELECTRICAL SIGNALING EVENTS
// =================================================================================

// ElectricalEventType categorizes electrical signaling events
type ElectricalEventType string

const (
	ElectricalSpike       ElectricalEventType = "spike"       // Action potential
	ElectricalBurst       ElectricalEventType = "burst"       // Burst of spikes
	ElectricalSynchrony   ElectricalEventType = "synchrony"   // Synchronous activity
	ElectricalOscillation ElectricalEventType = "oscillation" // Oscillatory activity
	ElectricalCoupling    ElectricalEventType = "coupling"    // Gap junction coupling
	ElectricalUncoupling  ElectricalEventType = "uncoupling"  // Gap junction uncoupling
	ElectricalPropagation ElectricalEventType = "propagation" // Signal propagation
)

// ElectricalEvent represents electrical signaling events
// Used for tracking electrical activity and gap junction communication
type ElectricalEvent struct {
	EventType   ElectricalEventType    `json:"event_type"`  // Type of electrical event
	SourceID    string                 `json:"source_id"`   // Component generating signal
	TargetID    string                 `json:"target_id"`   // Component receiving signal (if applicable)
	SignalType  SignalType             `json:"signal_type"` // Type of electrical signal
	Amplitude   float64                `json:"amplitude"`   // Signal amplitude
	Frequency   float64                `json:"frequency"`   // Signal frequency (if applicable)
	Duration    time.Duration          `json:"duration"`    // Signal duration
	Timestamp   time.Time              `json:"timestamp"`   // When event occurred
	Position    Position3D             `json:"position"`    // 3D location of event
	Propagation bool                   `json:"propagation"` // Whether signal propagated
	Metadata    map[string]interface{} `json:"metadata"`    // Additional event data
}

// BindingEvent records a chemical binding event for biological analysis
type BindingEvent struct {
	LigandType    LigandType `json:"ligand_type"`
	SourceID      string     `json:"source_id"`
	Concentration float64    `json:"concentration"`
	Timestamp     time.Time  `json:"timestamp"`
}

// EventType is a string identifier for the type of a biological event.
type EventType string

// Constants for all defined biological event types.
const (
	// --- Microglia Events (Lifecycle & Health) ---
	HealthPenaltyApplied        EventType = "health.penalty.applied"
	HealthReported              EventType = "health.reported" // NEW
	PruningCandidateMarked      EventType = "pruning.candidate.marked"
	ConnectionPruned            EventType = "connection.pruned"
	BirthRequestEvaluated       EventType = "birth.request.evaluated"
	ComponentApoptosisScheduled EventType = "component.apoptosis.scheduled"
	PatrolCompleted             EventType = "patrol.completed"

	// --- Astrocyte Network Events (Structural) ---
	ComponentRegistered   EventType = "component.registered"
	ComponentUnregistered EventType = "component.unregistered"
	TerritoryAdjusted     EventType = "territory.adjusted"

	// --- Chemical Modulator Events (Signaling) ---
	LigandReleased      EventType = "ligand.released"
	LigandBoundToTarget EventType = "ligand.bound.target"

	// --- Gap Junction Events (Electrical) ---
	ElectricalSignalSent          EventType = "electrical.signal.sent"
	ElectricalCouplingEstablished EventType = "electrical.coupling.established"
	ElectricalCouplingRemoved     EventType = "electrical.coupling.removed"

	// --- Neuron Events ---
	NeuronCreated  EventType = "neuron.created"
	NeuronFired    EventType = "neuron.fired"
	NeuronReceived EventType = "neuron.received"

	// --- Synapse Events ---
	SynapseCreated       EventType = "synapse.created"
	SynapseTransmitted   EventType = "synapse.transmitted"
	SynapseWeightChanged EventType = "synapse.weight.changed"
)

// BiologicalEvent represents a single, significant functional occurrence within the matrix.
type BiologicalEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	EventType   EventType `json:"event_type"`
	SourceID    string    `json:"source_id"`
	TargetID    string    `json:"target_id,omitempty"`
	Description string    `json:"description"`

	// === REUSABLE TYPED FIELDS (all from types package) ===
	Position      *Position3D    `json:"position,omitempty"`
	ComponentInfo *ComponentInfo `json:"component_info,omitempty"` // Same package!
	SynapseInfo   *SynapseInfo   `json:"synapse_info,omitempty"`   // Same package!

	LigandType    *LigandType `json:"ligand_type,omitempty"`
	Concentration *float64    `json:"concentration,omitempty"`
	SignalType    *SignalType `json:"signal_type,omitempty"`
	Strength      *float64    `json:"strength,omitempty"`

	Data interface{} `json:"data,omitempty"`
}

// BiologicalObserver defines the interface for an event emission system.
type BiologicalObserver interface {
	Emit(event BiologicalEvent)
}
