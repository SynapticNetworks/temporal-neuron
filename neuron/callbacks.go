package neuron

import (
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/message"
)

// === OUTPUT CALLBACK (DIRECT SYNAPSE COMMUNICATION) ===
type OutputCallback struct {
	TransmitMessage func(msg message.NeuralSignal) error
	GetWeight       func() float64
	GetDelay        func() time.Duration
	GetTargetID     func() string
}

// InputCallback allows synapses to schedule delayed delivery back to neurons
// This avoids circular package dependencies while maintaining clean separation
// unused.
type InputCallback struct {
	ScheduleDelivery func(msg message.NeuralSignal, targetID string, delay time.Duration) error
	GetNeuronID      func() string
}

// === ENHANCED MATRIX SERVICE CALLBACKS (INJECTED COORDINATION) ===
type NeuronCallbacks struct {
	// ===  NETWORK-WIDE SIGNALING ===
	// INTERACTION: Volume transmission, neuromodulation, gap junction coordination
	ReleaseChemical      func(ligandType message.LigandType, concentration float64) error
	SendElectricalSignal func(signalType message.SignalType, data interface{})

	// ===  SPATIAL SERVICES ===
	// INTERACTION: Distance-dependent delays, spatial clustering, anatomical constraints
	GetSpatialDelay      func(targetID string) time.Duration
	FindNearbyComponents func(radius float64) []component.ComponentInfo

	// ===  HEALTH & STATE REPORTING ===
	// INTERACTION: Network monitoring, homeostatic coordination, system health
	ReportHealth      func(activityLevel float64, connectionCount int)
	ReportStateChange func(oldState, newState component.ComponentState)

	// ===  BASIC SYNAPSE CREATION ===
	// INTERACTION: Neuroplasticity, structural plasticity, synaptogenesis
	CreateSynapse func(config SynapseCreationConfig) (string, error)

	// === ENHANCED SYNAPSE MANAGEMENT ===
	// INTERACTION: Structural plasticity, synaptic pruning
	DeleteSynapse func(synapseID string) error

	// === SYNAPSE DISCOVERY & ACCESS ===
	// INTERACTION: STDP feedback, homeostatic scaling, network analysis
	GetSynapse   func(synapseID string) (SynapticProcessor, error)
	ListSynapses func(criteria SynapseCriteria) []SynapseInfo

	// === PLASTICITY OPERATIONS ===
	// INTERACTION: STDP learning, homeostatic plasticity, competitive learning
	ApplyPlasticity  func(synapseID string, adjustment PlasticityAdjustment) error
	GetSynapseWeight func(synapseID string) (float64, error)
	SetSynapseWeight func(synapseID string, weight float64) error

	// === MATRIX ACCESS ===
	// INTERACTION: Spatial delay enhancement, extracellular signaling
	GetMatrix func() ExtracellularMatrix
}

// === SYNAPSE CREATION CONFIGURATION ===
type SynapseCreationConfig struct {
	SourceNeuronID string
	TargetNeuronID string
	InitialWeight  float64
	SynapseType    string
	PlasticityType string
	Delay          time.Duration
	Position       component.Position3D
}

// === SYNAPSE MANAGEMENT TYPES ===

// SynapseCriteria for filtering synapse queries
// USAGE: ListSynapses callback parameter for finding specific synapses
type SynapseCriteria struct {
	Direction     *SynapseDirection // Filter by incoming/outgoing/both
	SourceID      *string           // Filter by source neuron ID
	TargetID      *string           // Filter by target neuron ID
	WeightRange   *WeightRange      // Filter by weight bounds
	ActivitySince *time.Time        // Filter by recent activity
	SynapseType   *string           // Filter by synapse type
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
	Min, Max float64
}

// SynapseInfo provides read-only information about synapses
// USAGE: Returned by ListSynapses callback for synapse analysis
type SynapseInfo struct {
	ID           string           // Unique synapse identifier
	SourceID     string           // Pre-synaptic neuron ID
	TargetID     string           // Post-synaptic neuron ID
	Weight       float64          // Current synaptic weight/strength
	Delay        time.Duration    // Transmission delay
	LastActivity time.Time        // Most recent transmission time
	Direction    SynapseDirection // Direction relative to querying neuron
	SynapseType  string           // Type: "excitatory", "inhibitory", "modulatory"
}

// PlasticityAdjustment for STDP and other learning mechanisms
// USAGE: Parameter for ApplyPlasticity callback
type PlasticityAdjustment struct {
	DeltaT       time.Duration // Time difference for STDP (t_pre - t_post)
	WeightChange float64       // Direct weight modification (optional)
	LearningRate float64       // Context-specific learning rate (optional)
}

// === SYNAPSE PROCESSOR INTERFACE ===
// USAGE: Interface returned by GetSynapse callback for direct synapse operations
type SynapticProcessor interface {
	ID() string                                      // Get synapse identifier
	Transmit(signalValue float64)                    // Send signal through synapse
	ApplyPlasticity(adjustment PlasticityAdjustment) // Apply learning adjustment
	ShouldPrune() bool                               // Check if synapse should be removed
	GetWeight() float64                              // Get current weight
	SetWeight(weight float64)                        // Set weight directly
}

// === EXTRACELLULAR MATRIX INTERFACE ===
// USAGE: Interface returned by GetMatrix callback for spatial operations
type ExtracellularMatrix interface {
	// Enhance synaptic delay with spatial factors (distance, medium properties)
	EnhanceSynapticDelay(preNeuronID, postNeuronID, synapseID string, baseDelay time.Duration) time.Duration
}

// ===  NEURON INTERFACE FOR MATRIX ===
type NeuronInterface interface {
	// Embed component interfaces
	component.Component
	component.ChemicalReceiver
	component.ChemicalReleaser
	component.ElectricalReceiver
	component.ElectricalTransmitter
	component.MessageReceiver
	component.MonitorableComponent

	// Neuron-specific methods
	GetThreshold() float64
	SetThreshold(threshold float64)
	GetConnectionCount() int

	// Configuration
	SetReceptors(receptors []message.LigandType)
	SetReleasedLigands(ligands []message.LigandType)

	// Callback injection
	SetCallbacks(callbacks NeuronCallbacks)
	AddOutputCallback(synapseID string, callback OutputCallback)
	RemoveOutputCallback(synapseID string)

	// Network building
	ConnectToNeuron(targetNeuronID string, weight float64, synapseType string) error

	// === ENHANCED SYNAPSE OPERATIONS ===
	// BIOLOGICAL INTERACTIONS: These methods use the enhanced callbacks
	SendSTDPFeedback()                            // USES: ListSynapses, ApplyPlasticity
	PerformHomeostasisScaling()                   // USES: ListSynapses, SetSynapseWeight
	PruneDysfunctionalSynapses()                  // USES: ListSynapses, GetSynapse, DeleteSynapse
	GetConnectionMetrics() map[string]interface{} // USES: ListSynapses

	// Processing
	Run() // Background processing loop
}

// ===  MESSAGE RECEIVER INTERFACE ===
// This interface allows components (like Neurons or Synapses) to receive incoming neural signals.
// It is used by the axon for dispatching messages to their targets.
type MessageReceiver interface {
	Receive(msg message.NeuralSignal)
	ID() string // Add this method
}

// ============================================================================
// CALLBACK USAGE DOCUMENTATION
// ============================================================================

/*
CALLBACK USAGE BY BIOLOGICAL INTERACTION:

1. STDP LEARNING (Spike-Timing Dependent Plasticity):
   - SendSTDPFeedback() USES:
     * ListSynapses(criteria) -> Find recently active incoming synapses
     * ApplyPlasticity(synapseID, adjustment) -> Send timing-based learning signal

2. HOMEOSTATIC SCALING (Activity-dependent weight scaling):
   - PerformHomeostasisScaling() USES:
     * ListSynapses(criteria) -> Get all incoming connections
     * SetSynapseWeight(synapseID, weight) -> Scale weights proportionally

3. STRUCTURAL PLASTICITY (Synaptic pruning):
   - PruneDysfunctionalSynapses() USES:
     * ListSynapses(criteria) -> Find all connected synapses
     * GetSynapse(synapseID) -> Check pruning criteria
     * DeleteSynapse(synapseID) -> Remove dysfunctional connections

4. NETWORK ANALYSIS (Connectivity monitoring):
   - GetConnectionMetrics() USES:
     * ListSynapses(criteria) -> Analyze connection patterns and weights

5. SPATIAL PROCESSING (Distance-dependent delays):
   - Enhanced transmission USES:
     * GetMatrix() -> Access spatial delay enhancement

6. SYNAPTOGENESIS (New connection formation):
   - ConnectToNeuron() USES:
     * CreateSynapse(config) -> Request new synaptic connection

7. VOLUME TRANSMISSION (Network-wide signaling):
   - fireUnsafe() USES:
     * SendElectricalSignal(type, data) -> Gap junction coordination
     * ReleaseChemical(ligand, concentration) -> Neurotransmitter release

8. HEALTH MONITORING (System coordination):
   - fireUnsafe() USES:
     * ReportHealth(activity, connections) -> Network health updates
*/
