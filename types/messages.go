// types/messages.go
package types

import "time"

// =================================================================================
// CORE NEURAL MESSAGING STRUCTURES
// =================================================================================

// NeuralSignal represents the fundamental unit of neural communication
// This is the primary message type passed between neurons and synapses
type NeuralSignal struct {
	Value                float64    `json:"value"`                  // Signal strength/amplitude
	Timestamp            time.Time  `json:"timestamp"`              // When signal was generated
	SourceID             string     `json:"source_id"`              // ID of sending component
	TargetID             string     `json:"target_id"`              // ID of receiving component
	SynapseID            string     `json:"synapse_id,omitempty"`   // ID of transmitting synapse (if applicable)
	NeurotransmitterType LigandType `json:"neurotransmitter_type"`  // Chemical messenger type
	MessageType          string     `json:"message_type,omitempty"` // Optional message classification
}

// SynapseMessage represents a message transmitted through a synapse
// This extends NeuralSignal with synapse-specific information
type SynapseMessage struct {
	Value     float64    `json:"value"`      // Signal strength after synaptic scaling
	Timestamp time.Time  `json:"timestamp"`  // When message was created by synapse
	SourceID  string     `json:"source_id"`  // Pre-synaptic neuron ID
	SynapseID string     `json:"synapse_id"` // Synapse that transmitted message
	TargetID  string     `json:"target_id"`  // Post-synaptic neuron ID
	Ligand    LigandType `json:"ligand"`     // Neurotransmitter type released
}

// =================================================================================
// ACTIVITY AND MONITORING MESSAGES
// =================================================================================

// ActivityInfo provides information about component activity
// Used for health monitoring and network analysis
type ActivityInfo struct {
	ComponentID           string        `json:"component_id"`            // Component identifier
	LastTransmission      time.Time     `json:"last_transmission"`       // Time of last signal transmission
	LastPlasticity        time.Time     `json:"last_plasticity"`         // Time of last plasticity event
	Weight                float64       `json:"weight"`                  // Current synaptic weight (for synapses)
	ActivityLevel         float64       `json:"activity_level"`          // Recent activity rate (0.0-1.0)
	TimeSinceTransmission time.Duration `json:"time_since_transmission"` // Duration since last transmission
	TimeSincePlasticity   time.Duration `json:"time_since_plasticity"`   // Duration since last plasticity
	ConnectionCount       int           `json:"connection_count"`        // Number of connections (for neurons)
}

// SynapticActivity represents detailed synapse activity information
// Used for plasticity monitoring and synaptic analysis
type SynapticActivity struct {
	SynapseID      string    `json:"synapse_id"`      // Synapse identifier
	Timestamp      time.Time `json:"timestamp"`       // When activity occurred
	MessageValue   float64   `json:"message_value"`   // Value of transmitted message
	CurrentWeight  float64   `json:"current_weight"`  // Current synaptic weight
	ActivityType   string    `json:"activity_type"`   // Type: "transmission", "plasticity", etc.
	PresynapticID  string    `json:"presynaptic_id"`  // Source neuron ID
	PostsynapticID string    `json:"postsynaptic_id"` // Target neuron ID
}

// HealthReport represents component health status
// Used by monitoring systems to track component wellness
type HealthReport struct {
	ComponentID     string                 `json:"component_id"`     // Component being reported on
	Timestamp       time.Time              `json:"timestamp"`        // When report was generated
	ActivityLevel   float64                `json:"activity_level"`   // Current activity level
	ConnectionCount int                    `json:"connection_count"` // Number of connections
	HealthScore     float64                `json:"health_score"`     // Overall health score (0.0-1.0)
	Issues          []string               `json:"issues"`           // List of identified issues
	Metadata        map[string]interface{} `json:"metadata"`         // Additional health data
}

// =================================================================================
// CHEMICAL SIGNALING MESSAGES
// =================================================================================

// ChemicalRelease represents neurotransmitter release events
// Used for chemical signaling coordination and monitoring
type ChemicalRelease struct {
	SourceID      string     `json:"source_id"`     // Component releasing chemical
	LigandType    LigandType `json:"ligand_type"`   // Type of neurotransmitter
	Concentration float64    `json:"concentration"` // Concentration released
	Position      Position3D `json:"position"`      // 3D location of release
	Timestamp     time.Time  `json:"timestamp"`     // When release occurred
	Radius        float64    `json:"radius"`        // Effective diffusion radius
}

// ChemicalBinding represents neurotransmitter binding events
// Used for tracking chemical signal reception
type ChemicalBinding struct {
	TargetID      string     `json:"target_id"`     // Component receiving chemical
	SourceID      string     `json:"source_id"`     // Component that released chemical
	LigandType    LigandType `json:"ligand_type"`   // Type of neurotransmitter
	Concentration float64    `json:"concentration"` // Concentration bound
	Position      Position3D `json:"position"`      // 3D location of binding
	Timestamp     time.Time  `json:"timestamp"`     // When binding occurred
	ReceptorType  string     `json:"receptor_type"` // Type of receptor involved
}

// =================================================================================
// ELECTRICAL SIGNALING MESSAGES
// =================================================================================

// ElectricalSignal represents gap junction and network coordination signals
// Used for electrical coupling and network-wide communication
type ElectricalSignal struct {
	SourceID   string      `json:"source_id"`   // Component sending signal
	SignalType SignalType  `json:"signal_type"` // Type of electrical signal
	Data       interface{} `json:"data"`        // Signal payload data
	Timestamp  time.Time   `json:"timestamp"`   // When signal was sent
	Strength   float64     `json:"strength"`    // Signal strength/amplitude
	TargetID   string      `json:"target_id"`   // Specific target (if any)
	Broadcast  bool        `json:"broadcast"`   // Whether signal is broadcast
}

// =================================================================================
// COMPONENT INFORMATION STRUCTURES
// =================================================================================

// SynapseInfo provides read-only information about synapses
// Used by ListSynapses callbacks for synapse analysis
type SynapseInfo struct {
	ID           string                 `json:"id"`            // Unique synapse identifier
	SourceID     string                 `json:"source_id"`     // Pre-synaptic neuron ID
	TargetID     string                 `json:"target_id"`     // Post-synaptic neuron ID
	Weight       float64                `json:"weight"`        // Current synaptic weight/strength
	Delay        time.Duration          `json:"delay"`         // Transmission delay
	LastActivity time.Time              `json:"last_activity"` // Most recent transmission time
	Direction    SynapseDirection       `json:"direction"`     // Direction relative to querying neuron
	SynapseType  string                 `json:"synapse_type"`  // Type: "excitatory", "inhibitory", "modulatory"
	Position     Position3D             `json:"position"`      // 3D spatial location
	LigandType   LigandType             `json:"ligand_type"`   // Neurotransmitter type
	IsActive     bool                   `json:"is_active"`     // Whether synapse is currently active
	Metadata     map[string]interface{} `json:"metadata"`      // Additional synapse data
}

// ComponentInfo provides comprehensive information about any component
// Used for introspection, debugging, and system monitoring
type ComponentInfo struct {
	ID            string                 `json:"id"`             // Unique identifier
	Type          ComponentType          `json:"type"`           // Component type
	Position      Position3D             `json:"position"`       // 3D spatial coordinates
	State         ComponentState         `json:"state"`          // Current operational state
	RegisteredAt  time.Time              `json:"registered_at"`  // When component was created
	LastActivity  time.Time              `json:"last_activity"`  // Most recent activity
	ActivityLevel float64                `json:"activity_level"` // Current activity level
	HealthScore   float64                `json:"health_score"`   // Overall health score
	Connections   []string               `json:"connections"`    // Connected component IDs
	Metadata      map[string]interface{} `json:"metadata"`       // Additional component data
}
