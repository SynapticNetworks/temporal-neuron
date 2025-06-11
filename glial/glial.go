/*
=================================================================================
GLIAL CELLS - THE BRAIN'S MONITORING AND SUPPORT SYSTEM
=================================================================================

BIOLOGICAL OVERVIEW:
For over a century, glial cells were dismissed as mere "neural glue" - passive
support cells that simply held neurons in place. This view has been revolutionized
by modern neuroscience, which reveals glial cells as sophisticated, active
participants in neural computation that perform crucial monitoring, regulation,
and maintenance functions.

Glial cells outnumber neurons 10:1 in the human brain and perform functions
essential for neural network operation:

1. ASTROCYTES - The Neural Activity Monitors
   Astrocytes ensheath synapses with their fine processes, creating intimate
   contact with neural communication sites. They monitor neurotransmitter release,
   regulate synaptic transmission strength, and coordinate activity across neural
   regions through calcium wave propagation. A single astrocyte can monitor
   thousands of synapses simultaneously.

2. MICROGLIA - The Neural Health Surveillance System
   Microglia are the brain's resident immune cells, but their role extends far
   beyond immunity. They continuously patrol neural tissue with highly motile
   processes, scanning every synapse multiple times per hour. They monitor neural
   health, prune ineffective synapses during development and learning, and
   coordinate responses to neural dysfunction.

3. OLIGODENDROCYTES - The Transmission Efficiency Optimizers
   Oligodendrocytes wrap neural axons with myelin sheaths, dramatically increasing
   transmission speed and efficiency. They monitor neural activity patterns and
   adaptively adjust myelination based on usage, optimizing frequently-used
   pathways while allowing less-used connections to remain unmyelinated.

COMPUTATIONAL SIGNIFICANCE:
Traditional artificial neural networks completely lack this monitoring layer,
missing the sophisticated state tracking and adaptive regulation that enables
biological brains to:
- Maintain stable operation while continuously learning
- Detect and respond to processing bottlenecks
- Optimize network structure based on usage patterns
- Recover from component failures through adaptive reorganization
- Provide real-time feedback about network health and performance

This glial package brings these biological monitoring capabilities to artificial
neural networks, enabling observability and adaptive regulation.

BIOLOGICAL INSPIRATION VS PRACTICAL UTILITY:
While deeply inspired by biological glial cells, our implementation prioritizes
practical utility for artificial neural networks. We model the essential
monitoring and regulatory functions of glial cells while adapting them for
software neural networks and computational constraints.

=================================================================================
*/

package glial

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/synapse"
)

// GlialType represents the different functional classes of glial cells
// Each type performs specialized monitoring and support functions in biological brains
type GlialType int

const (
	// AstrocyteType models the star-shaped glial cells that monitor synaptic activity
	// Biological function: Neurotransmitter regulation, synaptic strength modulation,
	// calcium wave coordination, metabolic support for neurons
	AstrocyteType GlialType = iota

	// MicrogliaType models the brain's resident immune and maintenance cells
	// Biological function: Continuous tissue surveillance, synaptic pruning,
	// neural health monitoring, damage response coordination
	MicrogliaType

	// OligodendrocyteType models the cells that create myelin sheaths around axons
	// Biological function: Transmission speed optimization, metabolic support,
	// activity-dependent myelination, neural pathway efficiency
	OligodendrocyteType

	// ProcessingMonitorType represents our current implementation focus
	// Models the core monitoring functions shared by all glial cell types:
	// continuous surveillance of neural activity and state tracking
	ProcessingMonitorType
)

func (g GlialType) String() string {
	switch g {
	case AstrocyteType:
		return "Astrocyte"
	case MicrogliaType:
		return "Microglia"
	case OligodendrocyteType:
		return "Oligodendrocyte"
	case ProcessingMonitorType:
		return "ProcessingMonitor"
	default:
		return "Unknown"
	}
}

// ProcessingPhase represents discrete phases of neural computation
// Models the biological reality that neural processing occurs in distinct,
// observable phases rather than as instantaneous mathematical operations
//
// BIOLOGICAL CONTEXT:
// Real neurons don't just "activate" - they progress through distinct phases
// of electrical and chemical activity that can be observed and measured:
// 1. Signal reception at dendrites (synaptic integration)
// 2. Membrane potential integration at the cell body
// 3. Action potential generation at the axon hillock
// 4. Signal propagation down the axon
// 5. Recovery and return to resting state
//
// These phases occur over different timescales (microseconds to milliseconds)
// and can be monitored by glial cells to assess neural function
type ProcessingPhase int

const (
	// PhaseIdle represents the neuron at rest, not actively processing signals
	// Biological state: Resting membrane potential, baseline metabolic activity
	// Duration: Variable (seconds to hours between active processing)
	// Glial monitoring: Background surveillance, metabolic support assessment
	PhaseIdle ProcessingPhase = iota

	// PhaseReceiving represents active signal reception from synaptic inputs
	// Biological state: Postsynaptic potentials arriving at dendrites
	// Duration: Milliseconds (synaptic transmission timescale)
	// Glial monitoring: Synaptic activity detection, neurotransmitter sensing
	PhaseReceiving

	// PhaseIntegrating represents dendritic integration and membrane potential summation
	// Biological state: Electrical signals propagating toward cell body, spatial/temporal summation
	// Duration: Milliseconds (membrane time constant, typically 10-20ms)
	// Glial monitoring: Membrane potential changes, calcium influx detection
	PhaseIntegrating

	// PhaseFiring represents action potential generation and propagation
	// Biological state: Voltage-gated sodium channels open, action potential propagates
	// Duration: 1-2 milliseconds (action potential duration)
	// Glial monitoring: High-frequency electrical activity, metabolic burst detection
	PhaseFiring

	// PhaseRecovery represents the refractory period and return to resting state
	// Biological state: Sodium channel inactivation, potassium efflux, membrane repolarization
	// Duration: 5-15 milliseconds (absolute and relative refractory periods)
	// Glial monitoring: Recovery assessment, metabolic restoration tracking
	PhaseRecovery
)

func (p ProcessingPhase) String() string {
	switch p {
	case PhaseIdle:
		return "Idle"
	case PhaseReceiving:
		return "Receiving"
	case PhaseIntegrating:
		return "Integrating"
	case PhaseFiring:
		return "Firing"
	case PhaseRecovery:
		return "Recovery"
	default:
		return "Unknown"
	}
}

// ProcessingState represents a comprehensive snapshot of neural processing activity
// Models the type of information that biological glial cells continuously gather
// about the neurons they monitor through direct contact and biochemical sensing
//
// BIOLOGICAL CORRESPONDENCE:
// Astrocytes make intimate contact with neural membranes and can detect:
// - Electrical activity through gap junctions and ion channel sensing
// - Chemical activity through neurotransmitter uptake and calcium waves
// - Metabolic activity through glucose consumption and ATP production
// - Timing patterns through repetitive calcium responses
//
// This information enables glial cells to assess neural health, predict failures,
// optimize support, and coordinate network-wide activities
type ProcessingState struct {
	// === NEURAL IDENTITY ===
	NeuronID string `json:"neuron_id"` // Unique identifier of monitored neuron
	// Biological analogy: Spatial location and molecular markers that identify
	// specific neurons for targeted glial support and monitoring

	// === PROCESSING PHASE INFORMATION ===
	Phase ProcessingPhase `json:"phase"` // Current phase of neural processing
	// Biological measurement: Detected through membrane potential changes,
	// ion channel activity, and neurotransmitter release patterns

	IsProcessing bool `json:"is_processing"` // Whether neuron is actively processing
	// Biological indicator: Elevated electrical activity, increased metabolism,
	// calcium influx, and neurotransmitter turnover above baseline levels

	// === TEMPORAL TRACKING ===
	MessageID uint64 `json:"message_id,omitempty"` // Unique identifier for tracked processing event
	// Biological analogy: Glial cells can track individual synaptic events
	// and their consequences through calcium responses and molecular signaling

	StartTime time.Time `json:"start_time,omitempty"` // When current processing began
	// Biological measurement: Timestamp of when electrical/chemical activity
	// deviated from baseline, detectable through real-time calcium imaging

	LastActivity time.Time `json:"last_activity"` // Most recent detected activity
	// Biological tracking: Glial cells maintain continuous surveillance and
	// can detect when neurons become silent or show irregular activity patterns

	ProcessingTime time.Duration `json:"processing_time,omitempty"` // Total time for current processing
	// Biological measurement: Duration from initial synaptic input to completion
	// of response, including action potential generation and propagation
}

// GlialStatus represents the operational state of a glial cell
// Models the biological reality that glial cells have their own activity states,
// territorial boundaries, and functional capacity limits
//
// BIOLOGICAL CONTEXT:
// Glial cells are not passive monitors - they have their own complex biology:
// - Activation states (quiescent, reactive, proliferative)
// - Territorial boundaries (each cell monitors specific neural regions)
// - Resource limitations (finite energy, protein synthesis capacity)
// - Communication networks (gap junctions, chemical signaling)
// - Adaptive responses (increased monitoring during high neural activity)
type GlialStatus struct {
	// === GLIAL IDENTITY ===
	ID   string    `json:"id"`   // Unique identifier for this glial cell
	Type GlialType `json:"type"` // Functional classification of glial cell

	// === OPERATIONAL STATE ===
	IsActive bool `json:"is_active"` // Whether glial cell is actively monitoring
	// Biological state: Glial activation involves increased metabolic activity,
	// process extension, enhanced monitoring capability, and reactive responses

	MonitoredTargets int `json:"monitored_targets"` // Number of neurons under surveillance
	// Biological constraint: Each glial cell has limited monitoring capacity
	// - Single astrocyte: ~100,000 synapses from ~300-600 neurons
	// - Single microglia: ~15-30 neurons in territorial domain
	// - Single oligodendrocyte: ~40-80 axon segments

	// === TEMPORAL INFORMATION ===
	LastUpdate time.Time     `json:"last_update"` // Most recent status update
	Uptime     time.Duration `json:"uptime"`      // How long glial cell has been active
	// Biological tracking: Glial cells maintain persistent activity throughout
	// life, with activity levels varying based on neural demands and health status
}

// GlialCell defines the fundamental interface for all glial cell types
// Captures the essential biological functions shared by all glial cells:
// autonomous operation, continuous monitoring, and adaptive responses
//
// BIOLOGICAL DESIGN PRINCIPLE:
// All glial cell types share core characteristics despite their specializations:
// - Autonomous operation independent of neural activity
// - Continuous surveillance of their territorial domains
// - Ability to detect and respond to neural state changes
// - Communication with other glial cells and neurons
// - Adaptive modification of their monitoring and support functions
type GlialCell interface {
	// === IDENTITY AND CLASSIFICATION ===
	ID() string      // Unique identifier for this glial cell
	Type() GlialType // Functional classification (astrocyte, microglia, etc.)

	// === LIFECYCLE MANAGEMENT ===
	// Models the biological lifecycle of glial cells: activation, sustained
	// operation, and controlled shutdown (apoptosis or quiescence)
	Run() error     // Begin autonomous monitoring and support operations
	Stop() error    // Cease operations and release resources
	IsActive() bool // Current operational status

	// === STATUS MONITORING ===
	GetStatus() GlialStatus // Current operational state and monitoring statistics
	// Biological function: Self-assessment and reporting to coordinate with
	// other glial cells and respond to changing neural demands
}

// NeuronInterface defines the minimal interface required for glial monitoring
// Specifies what information glial cells need to access from neurons they monitor
//
// BIOLOGICAL BASIS:
// Glial cells monitor neurons through multiple channels:
// - Direct membrane contact for electrical activity sensing
// - Neurotransmitter uptake for synaptic activity assessment
// - Ion concentration monitoring for metabolic state evaluation
// - Morphological observation for structural health assessment
//
// This interface captures the essential neural state information that enables
// effective glial monitoring without requiring detailed knowledge of neural
// implementation or creating tight coupling between packages
type NeuronInterface interface {
	// === NEURAL IDENTITY ===
	ID() string // Unique neural identifier for monitoring records

	// === COMMUNICATION ACCESS ===
	GetInputChannel() chan synapse.SynapseMessage // Access to neural input stream
	// Biological analogy: Glial cells can monitor synaptic inputs through
	// direct contact with synaptic terminals and neurotransmitter sensing

	// === ELECTRICAL STATE MONITORING ===
	GetAccumulator() float64      // Current membrane potential state
	GetCurrentThreshold() float64 // Current firing threshold (may be homeostatic)
	// Biological measurement: Glial cells can sense membrane potential through
	// gap junctions and detect threshold changes through calcium imaging

	// === BIOCHEMICAL STATE MONITORING ===
	GetCalciumLevel() float64 // Current intracellular calcium concentration
	// Biological significance: Calcium is the universal neural activity indicator
	// that glial cells use to assess neural function, detect stress, and
	// coordinate their own responses to neural demands
}

// ProcessingMonitor defines the interface for monitoring neural message processing
// Models the core monitoring functions performed by all glial cell types:
// detecting neural activity, tracking processing states, and providing
// real-time information about neural network operation
//
// BIOLOGICAL FOUNDATION:
// This interface captures the essential monitoring capabilities that enable
// biological glial cells to:
// - Detect when neurons receive and process synaptic inputs
// - Track the progression of neural processing through different phases
// - Identify processing completion and return to resting state
// - Provide feedback about neural function to other network components
// - Support experimental observation and network debugging
//
// These capabilities are fundamental to all glial cell types, though each
// specializes in monitoring different aspects of neural function
type ProcessingMonitor interface {
	GlialCell

	// === MONITORING TARGET MANAGEMENT ===
	// Models the biological process of glial cells establishing and maintaining
	// monitoring relationships with specific neurons in their territorial domains

	MonitorNeuron(neuron NeuronInterface) error // Begin monitoring a specific neuron
	// Biological process: Glial cell extends processes to make contact with
	// target neuron, establishes biochemical sensing, begins surveillance

	StopMonitoringNeuron(neuronID string) error // Cease monitoring specific neuron
	// Biological process: Glial cell retracts monitoring processes, reassigns
	// resources to other targets, may occur during development or damage

	GetMonitoredNeurons() []string // List of currently monitored neurons
	// Biological information: The territorial domain of this glial cell,
	// representing its current monitoring capacity and spatial organization

	// === PROCESSING STATE ACCESS ===
	// Models the biological ability of glial cells to assess neural processing
	// state through multiple sensing mechanisms and provide real-time information

	GetProcessingState(neuronID string) (ProcessingState, error) // Current processing state
	// Biological capability: Real-time assessment of neural electrical and
	// chemical activity through direct membrane contact and molecular sensing

	IsNeuronProcessing(neuronID string) bool // Quick processing status check
	// Biological function: Rapid activity detection for immediate response
	// decisions, analogous to fast calcium responses in glial cells

	// === EXPERIMENTAL AND TESTING UTILITIES ===
	// Models the biological reality that glial cells can influence neural activity
	// through neurotransmitter release, electrical coupling, and metabolic support
	// These functions enable controlled testing and network manipulation

	SendTestMessage(neuronID string, msg synapse.SynapseMessage) (uint64, error)
	// Biological analogy: Glial cells can stimulate neurons through gliotransmitter
	// release, electrical coupling, or metabolic support modulation

	WaitForProcessingComplete(neuronID string, messageID uint64, timeout time.Duration) error
	// Biological capability: Glial cells can detect when neural processing events
	// are complete through sustained monitoring of electrical and chemical signals

	WaitForQuiescence(neuronID string, timeout time.Duration) error
	// Biological function: Detection of neural return to resting state, indicating
	// completion of processing and readiness for subsequent inputs
}

// =================================================================================
// PROCESSING MONITOR IMPLEMENTATION
// Core glial monitoring functionality focused on neural message processing
// =================================================================================

// BasicProcessingMonitor implements the ProcessingMonitor interface
// Provides real-time monitoring of neural message processing with minimal overhead
//
// BIOLOGICAL MODEL:
// This implementation models the core surveillance functions performed by all
// glial cell types. Like biological glial cells, it:
// - Operates autonomously and continuously
// - Monitors multiple neurons simultaneously within capacity limits
// - Detects neural activity changes in real-time
// - Tracks processing progression through identifiable phases
// - Provides non-intrusive monitoring without interfering with neural computation
// - Maintains detailed records for analysis and coordination
//
// DESIGN PRINCIPLES:
// 1. NON-INTRUSIVE: Monitoring does not interfere with neural processing
// 2. REAL-TIME: Immediate detection and reporting of neural state changes
// 3. AUTONOMOUS: Operates independently without external control
// 4. SCALABLE: Efficient monitoring of multiple neurons simultaneously
// 5. BIOLOGICALLY-INSPIRED: Processing states and transitions match neural biology
type BasicProcessingMonitor struct {
	// === GLIAL IDENTITY ===
	id        string    // Unique identifier for this monitoring instance
	startTime time.Time // When this glial cell began operation

	// === MONITORING STATE ===
	isActive bool // Current operational status of this glial cell
	// Biological state: Whether glial cell is actively performing surveillance
	// or has entered quiescent state due to resource constraints or damage

	// === NEURAL MONITORING MANAGEMENT ===
	monitoredNeurons map[string]NeuronInterface // Neurons under active surveillance
	neuronsMutex     sync.RWMutex               // Thread-safe access to monitored neurons map
	// Biological constraint: Each glial cell can monitor limited number of neurons
	// based on available resources and territorial boundaries

	// === PROCESSING STATE TRACKING ===
	processingStates map[string]*ProcessingState // Current state of each monitored neuron
	statesMutex      sync.RWMutex                // Thread-safe access to state information
	// Biological function: Continuous tracking of neural electrical and chemical
	// activity enables detection of processing phases and completion events

	// === MESSAGE TRACKING FOR TESTING ===
	messageIDCounter uint64                             // Atomic counter for unique message identification
	pendingMessages  map[uint64]*processingTracker      // Active message processing tracking
	messageWaiters   map[uint64][]chan processingResult // Channels waiting for specific message completion
	messagesMutex    sync.RWMutex                       // Thread-safe access to message tracking

	// === OPERATIONAL CONTROL ===
	ctx    context.Context    // Context for controlled shutdown
	cancel context.CancelFunc // Function to trigger shutdown
	wg     sync.WaitGroup     // Wait group for clean shutdown coordination

	// === CONFIGURATION ===
	config ProcessingMonitorConfig // Operating parameters for this glial cell
}

// ProcessingMonitorConfig contains configuration parameters for glial monitoring
// Models the biological reality that glial cells have adjustable monitoring
// sensitivity, territorial size, and response thresholds based on neural demands
type ProcessingMonitorConfig struct {
	// === MONITORING SENSITIVITY ===
	// Biological parameter: How sensitive glial cells are to neural activity changes
	ActivityThreshold float64 // Minimum activity level to trigger monitoring response
	// Models the calcium threshold for glial cell activation in response to neural activity

	StateUpdateInterval time.Duration // How frequently to update processing state
	// Biological timescale: Glial cells continuously monitor but update their
	// assessment and responses on finite timescales (seconds to minutes)

	// === TERRITORIAL LIMITS ===
	MaxMonitoredNeurons int // Maximum number of neurons this glial cell can monitor
	// Biological constraint: Physical and metabolic limits on glial monitoring capacity
	// Astrocytes: ~300-600 neurons, Microglia: ~15-30 neurons

	// === TEMPORAL PARAMETERS ===
	ProcessingTimeout time.Duration // Maximum time to wait for processing completion
	// Biological significance: Glial cells can detect when neurons become
	// unresponsive and trigger intervention or support responses

	QuiescenceTimeout time.Duration // Maximum time to wait for neural quiescence
	// Biological function: Detection of prolonged neural activity that may
	// indicate pathological states requiring glial intervention
}

// processingTracker tracks individual message processing events
// Models biological glial cell ability to track individual synaptic events
// and their consequences through the neural processing pipeline
type processingTracker struct {
	messageID  uint64                 // Unique identifier for this processing event
	neuronID   string                 // Target neuron being monitored
	message    synapse.SynapseMessage // Original message being processed
	startTime  time.Time              // When processing began
	phase      ProcessingPhase        // Current processing phase
	completion chan processingResult  // Channel for signaling completion
}

// processingResult contains the outcome of message processing monitoring
type processingResult struct {
	success        bool            // Whether processing completed successfully
	finalState     ProcessingState // Final neural state after processing
	processingTime time.Duration   // Total time from start to completion
	error          error           // Error information if processing failed
}

// NewBasicProcessingMonitor creates a new glial processing monitor
// Initializes the monitoring system with biologically-realistic parameters
//
// BIOLOGICAL INITIALIZATION:
// Models the process of glial cell activation and establishment of monitoring
// territory. Like biological glial cells, the monitor begins in an active
// surveillance state ready to establish connections with target neurons.
//
// Parameters:
// id: Unique identifier for this glial cell instance
// config: Operating parameters (if nil, uses biological defaults)
func NewBasicProcessingMonitor(id string, config *ProcessingMonitorConfig) *BasicProcessingMonitor {
	// Use biological defaults if no configuration provided
	if config == nil {
		config = &ProcessingMonitorConfig{
			ActivityThreshold:   0.001,                 // Sensitive detection threshold
			StateUpdateInterval: 10 * time.Millisecond, // Real-time monitoring
			MaxMonitoredNeurons: 100,                   // Reasonable monitoring capacity
			ProcessingTimeout:   1 * time.Second,       // Generous processing allowance
			QuiescenceTimeout:   5 * time.Second,       // Extended quiescence detection
		}
	}

	// Create cancellable context for controlled shutdown
	ctx, cancel := context.WithCancel(context.Background())

	monitor := &BasicProcessingMonitor{
		id:        id,
		startTime: time.Now(),
		isActive:  true,

		// Initialize monitoring data structures
		monitoredNeurons: make(map[string]NeuronInterface),
		processingStates: make(map[string]*ProcessingState),
		pendingMessages:  make(map[uint64]*processingTracker),
		messageWaiters:   make(map[uint64][]chan processingResult),

		// Operational control
		ctx:    ctx,
		cancel: cancel,
		config: *config,
	}

	return monitor
}

// =================================================================================
// GLIAL CELL INTERFACE IMPLEMENTATION
// =================================================================================

// ID returns the unique identifier for this glial cell
func (m *BasicProcessingMonitor) ID() string {
	return m.id
}

// Type returns the functional classification of this glial cell
func (m *BasicProcessingMonitor) Type() GlialType {
	return ProcessingMonitorType
}

// IsActive returns whether this glial cell is currently performing monitoring
func (m *BasicProcessingMonitor) IsActive() bool {
	return m.isActive && m.ctx.Err() == nil
}

// GetStatus returns current operational status and monitoring statistics
// Provides comprehensive information about glial cell function and capacity
func (m *BasicProcessingMonitor) GetStatus() GlialStatus {
	m.neuronsMutex.RLock()
	monitoredCount := len(m.monitoredNeurons)
	m.neuronsMutex.RUnlock()

	return GlialStatus{
		ID:               m.id,
		Type:             ProcessingMonitorType,
		IsActive:         m.IsActive(),
		MonitoredTargets: monitoredCount,
		LastUpdate:       time.Now(),
		Uptime:           time.Since(m.startTime),
	}
}

// Run begins autonomous monitoring operations
// Models the biological activation of glial cells and establishment of
// continuous surveillance of neural activity within territorial boundaries
//
// BIOLOGICAL PROCESS MODELED:
// When glial cells become active, they:
// 1. Extend monitoring processes to establish neural contact
// 2. Begin continuous electrical and chemical surveillance
// 3. Establish baseline activity levels for comparison
// 4. Coordinate with other glial cells in the region
// 5. Maintain autonomous operation until shutdown signals
func (m *BasicProcessingMonitor) Run() error {
	if !m.isActive {
		return fmt.Errorf("glial cell %s is not active", m.id)
	}

	// Start background monitoring routines
	m.wg.Add(1)
	go m.monitoringLoop()

	return nil
}

// Stop ceases all monitoring operations and releases resources
// Models the biological process of glial cell deactivation or apoptosis
func (m *BasicProcessingMonitor) Stop() error {
	if !m.isActive {
		return nil // Already stopped
	}

	m.isActive = false
	m.cancel()  // Signal shutdown to all monitoring routines
	m.wg.Wait() // Wait for clean shutdown

	// Notify any remaining waiters
	m.messagesMutex.Lock()
	for messageID, waiters := range m.messageWaiters {
		for _, waiter := range waiters {
			select {
			case waiter <- processingResult{
				success: false,
				error:   fmt.Errorf("monitoring stopped"),
			}:
			default:
			}
		}
		delete(m.messageWaiters, messageID)
	}
	m.messagesMutex.Unlock()

	return nil
}

// =================================================================================
// PROCESSING MONITOR INTERFACE IMPLEMENTATION
// =================================================================================

// MonitorNeuron begins monitoring a specific neuron
// Models the biological process of glial cells establishing monitoring
// relationships with neurons in their territorial domain
//
// BIOLOGICAL PROCESS:
// When glial cells target a neuron for monitoring:
// 1. Extend processes to make physical contact with neural membrane
// 2. Establish biochemical sensing for activity detection
// 3. Begin baseline activity assessment
// 4. Register neuron in territorial monitoring map
// 5. Begin continuous surveillance operations
func (m *BasicProcessingMonitor) MonitorNeuron(neuron NeuronInterface) error {
	if !m.IsActive() {
		return fmt.Errorf("glial cell %s is not active", m.id)
	}

	neuronID := neuron.ID()

	// Check monitoring capacity limits (biological constraint)
	m.neuronsMutex.Lock()
	defer m.neuronsMutex.Unlock()

	if len(m.monitoredNeurons) >= m.config.MaxMonitoredNeurons {
		return fmt.Errorf("monitoring capacity exceeded: cannot monitor more than %d neurons",
			m.config.MaxMonitoredNeurons)
	}

	// Check if already monitoring this neuron
	if _, exists := m.monitoredNeurons[neuronID]; exists {
		return fmt.Errorf("already monitoring neuron %s", neuronID)
	}

	// Establish monitoring relationship
	m.monitoredNeurons[neuronID] = neuron

	// Initialize processing state tracking
	m.statesMutex.Lock()
	m.processingStates[neuronID] = &ProcessingState{
		NeuronID:     neuronID,
		Phase:        PhaseIdle,
		IsProcessing: false,
		LastActivity: time.Now(),
	}
	m.statesMutex.Unlock()

	return nil
}

// StopMonitoringNeuron ceases monitoring of a specific neuron
// Models glial cell retraction of monitoring processes and resource reallocation
func (m *BasicProcessingMonitor) StopMonitoringNeuron(neuronID string) error {
	m.neuronsMutex.Lock()
	defer m.neuronsMutex.Unlock()

	// Remove from monitoring
	delete(m.monitoredNeurons, neuronID)

	// Clean up state tracking
	m.statesMutex.Lock()
	delete(m.processingStates, neuronID)
	m.statesMutex.Unlock()

	return nil
}

// GetMonitoredNeurons returns list of neurons currently under surveillance
func (m *BasicProcessingMonitor) GetMonitoredNeurons() []string {
	m.neuronsMutex.RLock()
	defer m.neuronsMutex.RUnlock()

	neurons := make([]string, 0, len(m.monitoredNeurons))
	for neuronID := range m.monitoredNeurons {
		neurons = append(neurons, neuronID)
	}

	return neurons
}

// GetProcessingState returns current processing state of specified neuron
// Models glial cell ability to assess neural activity through direct monitoring
func (m *BasicProcessingMonitor) GetProcessingState(neuronID string) (ProcessingState, error) {
	m.statesMutex.RLock()
	defer m.statesMutex.RUnlock()

	state, exists := m.processingStates[neuronID]
	if !exists {
		return ProcessingState{}, fmt.Errorf("neuron %s not monitored", neuronID)
	}

	// Return copy to prevent external modification
	return *state, nil
}

// IsNeuronProcessing provides quick check of neural processing status
func (m *BasicProcessingMonitor) IsNeuronProcessing(neuronID string) bool {
	state, err := m.GetProcessingState(neuronID)
	if err != nil {
		return false
	}
	return state.IsProcessing
}

// SendTestMessage sends a test message to a monitored neuron and tracks processing
// Models glial cell ability to stimulate neurons through gliotransmitter release
//
// BIOLOGICAL ANALOGY:
// Glial cells can influence neural activity through:
// - Release of gliotransmitters (ATP, glutamate, GABA)
// - Modulation of extracellular ion concentrations
// - Electrical coupling through gap junctions
// - Metabolic support modulation
//
// This function provides controlled neural stimulation for testing and research
func (m *BasicProcessingMonitor) SendTestMessage(neuronID string, msg synapse.SynapseMessage) (uint64, error) {
	// Verify neuron is monitored
	m.neuronsMutex.RLock()
	neuron, exists := m.monitoredNeurons[neuronID]
	m.neuronsMutex.RUnlock()

	if !exists {
		return 0, fmt.Errorf("neuron %s not monitored", neuronID)
	}

	// Generate unique message ID for tracking
	messageID := atomic.AddUint64(&m.messageIDCounter, 1)

	// Create processing tracker
	tracker := &processingTracker{
		messageID:  messageID,
		neuronID:   neuronID,
		message:    msg,
		startTime:  time.Now(),
		phase:      PhaseReceiving,
		completion: make(chan processingResult, 1),
	}

	// Register tracker for monitoring
	m.messagesMutex.Lock()
	m.pendingMessages[messageID] = tracker
	m.messagesMutex.Unlock()

	// Send message to neuron
	inputChannel := neuron.GetInputChannel()

	// Update processing state to reflect message sending
	m.updateProcessingState(neuronID, PhaseReceiving, messageID, time.Now())

	select {
	case inputChannel <- msg:
		// Message sent successfully
		return messageID, nil
	case <-time.After(100 * time.Millisecond):
		// Timeout sending message - cleanup tracker
		m.messagesMutex.Lock()
		delete(m.pendingMessages, messageID)
		m.messagesMutex.Unlock()
		return 0, fmt.Errorf("timeout sending message to neuron %s", neuronID)
	}
}

// WaitForProcessingComplete waits for a specific message to complete processing
// Models glial cell ability to detect completion of neural processing events
// through sustained monitoring of electrical and chemical activity
//
// BIOLOGICAL FUNCTION:
// Glial cells can detect when neural processing events complete by monitoring:
// - Return of membrane potential to resting levels
// - Cessation of calcium influx and return to baseline
// - Reduction in metabolic activity to resting rates
// - Stabilization of neurotransmitter concentrations
//
// This capability is essential for coordinating glial support functions
// and detecting when neurons are ready for subsequent inputs
func (m *BasicProcessingMonitor) WaitForProcessingComplete(neuronID string, messageID uint64, timeout time.Duration) error {
	if !m.IsActive() {
		return fmt.Errorf("glial cell %s is not active", m.id)
	}

	// Create completion channel for this waiter
	completionChan := make(chan processingResult, 1)

	// Register waiter for this message
	m.messagesMutex.Lock()
	if m.messageWaiters[messageID] == nil {
		m.messageWaiters[messageID] = make([]chan processingResult, 0)
	}
	m.messageWaiters[messageID] = append(m.messageWaiters[messageID], completionChan)
	m.messagesMutex.Unlock()

	// Wait for completion or timeout
	select {
	case result := <-completionChan:
		if result.error != nil {
			return result.error
		}
		return nil

	case <-time.After(timeout):
		// Remove from waiters list on timeout
		m.messagesMutex.Lock()
		waiters := m.messageWaiters[messageID]
		for i, ch := range waiters {
			if ch == completionChan {
				m.messageWaiters[messageID] = append(waiters[:i], waiters[i+1:]...)
				break
			}
		}
		if len(m.messageWaiters[messageID]) == 0 {
			delete(m.messageWaiters, messageID)
		}
		m.messagesMutex.Unlock()

		return fmt.Errorf("timeout waiting for processing completion of message %d", messageID)

	case <-m.ctx.Done():
		return fmt.Errorf("monitoring stopped")
	}
}

// WaitForQuiescence waits for a neuron to reach stable, non-processing state
// Models glial cell detection of neural return to resting state after activity
//
// BIOLOGICAL SIGNIFICANCE:
// Neural quiescence detection is crucial for glial cells to:
// - Determine when neurons are ready for subsequent stimulation
// - Assess recovery from metabolic stress
// - Detect pathological states (inability to return to rest)
// - Coordinate timing of glial support functions
// - Optimize energy allocation and metabolic support
//
// Biological indicators of neural quiescence:
// - Stable resting membrane potential
// - Baseline calcium levels
// - Normal metabolic rate
// - Cessation of neurotransmitter release
func (m *BasicProcessingMonitor) WaitForQuiescence(neuronID string, timeout time.Duration) error {
	if !m.IsActive() {
		return fmt.Errorf("glial cell %s is not active", m.id)
	}

	// Verify neuron is monitored
	m.neuronsMutex.RLock()
	neuron, exists := m.monitoredNeurons[neuronID]
	m.neuronsMutex.RUnlock()

	if !exists {
		return fmt.Errorf("neuron %s not monitored", neuronID)
	}

	deadline := time.Now().Add(timeout)
	checkInterval := 10 * time.Millisecond

	for time.Now().Before(deadline) {
		// Check if neuron has reached quiescent state
		if m.isNeuronQuiescent(neuron) {
			return nil
		}

		// Wait before next check
		select {
		case <-time.After(checkInterval):
			continue
		case <-m.ctx.Done():
			return fmt.Errorf("monitoring stopped")
		}
	}

	return fmt.Errorf("neuron %s did not reach quiescence within timeout", neuronID)
}

// =================================================================================
// INTERNAL MONITORING FUNCTIONS
// Core biological monitoring and state detection algorithms
// =================================================================================

// monitoringLoop runs the main monitoring routine for all supervised neurons
// Models the continuous surveillance performed by biological glial cells
//
// BIOLOGICAL PROCESS:
// Glial cells maintain continuous monitoring through:
// 1. Periodic scanning of territorial domain
// 2. Real-time detection of neural activity changes
// 3. Assessment of neural health and function
// 4. Coordination of support and regulatory responses
// 5. Communication with other glial cells
//
// This loop implements the core surveillance cycle that enables glial cells
// to detect processing events, track neural states, and provide real-time
// information about network operation
func (m *BasicProcessingMonitor) monitoringLoop() {
	defer m.wg.Done()

	// Monitoring ticker for regular state updates
	ticker := time.NewTicker(m.config.StateUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.updateAllNeuronStates()
			m.checkPendingMessages()

		case <-m.ctx.Done():
			return
		}
	}
}

// updateAllNeuronStates performs periodic state assessment of all monitored neurons
// Models the continuous surveillance scanning performed by glial cells
func (m *BasicProcessingMonitor) updateAllNeuronStates() {
	m.neuronsMutex.RLock()
	neurons := make(map[string]NeuronInterface, len(m.monitoredNeurons))
	for id, neuron := range m.monitoredNeurons {
		neurons[id] = neuron
	}
	m.neuronsMutex.RUnlock()

	// Update state for each monitored neuron
	for neuronID, neuron := range neurons {
		m.assessNeuronState(neuronID, neuron)
	}
}

// assessNeuronState evaluates current neural state and updates monitoring records
// Models the biological process of glial cells assessing neural health and activity
//
// BIOLOGICAL ASSESSMENT PROCESS:
// Glial cells evaluate neural state through multiple indicators:
// 1. Electrical activity (membrane potential, firing patterns)
// 2. Chemical activity (calcium levels, neurotransmitter turnover)
// 3. Metabolic activity (energy consumption, waste production)
// 4. Morphological state (dendritic structure, synaptic integrity)
//
// This assessment enables detection of processing phases, completion events,
// and potential dysfunction requiring intervention
func (m *BasicProcessingMonitor) assessNeuronState(neuronID string, neuron NeuronInterface) {
	// Get current neural parameters
	accumulator := neuron.GetAccumulator()
	threshold := neuron.GetCurrentThreshold()
	calcium := neuron.GetCalciumLevel()

	// Determine processing phase based on biological indicators
	phase := m.determineProcessingPhase(accumulator, threshold, calcium)
	isProcessing := m.isProcessingActive(accumulator, calcium)

	// Update processing state
	m.statesMutex.Lock()
	state, exists := m.processingStates[neuronID]
	if exists {
		// Update existing state
		oldPhase := state.Phase
		state.Phase = phase
		state.IsProcessing = isProcessing
		state.LastActivity = time.Now()

		// If phase changed, update phase timing
		if oldPhase != phase {
			state.StartTime = time.Now()
		}
	}
	m.statesMutex.Unlock()
}

// determineProcessingPhase analyzes neural parameters to identify current processing phase
// Models glial cell ability to recognize distinct phases of neural computation
//
// BIOLOGICAL PHASE DETECTION:
// Glial cells can identify neural processing phases through:
// - Membrane potential changes indicating signal reception/integration
// - Calcium influx patterns indicating action potential generation
// - Metabolic changes indicating increased neural activity
// - Neurotransmitter release indicating synaptic transmission
func (m *BasicProcessingMonitor) determineProcessingPhase(accumulator, threshold, calcium float64) ProcessingPhase {
	// Calculate neural activity indicators
	accumulatorRatio := accumulator / threshold
	calciumActivity := calcium

	// Biological thresholds for phase detection
	const (
		firingThreshold   = 0.95 // Very close to threshold indicates imminent firing
		activeThreshold   = 0.1  // Significant accumulation indicates active processing
		calciumThreshold  = 0.05 // Elevated calcium indicates recent activity
		recoveryThreshold = 0.02 // Low activity indicates recovery phase
	)

	// Phase determination based on biological indicators
	switch {
	case accumulatorRatio >= firingThreshold:
		// High membrane potential indicates imminent or active firing
		return PhaseFiring

	case accumulatorRatio >= activeThreshold:
		// Moderate membrane potential indicates active processing
		if calciumActivity > calciumThreshold {
			return PhaseIntegrating // High calcium suggests integration phase
		}
		return PhaseReceiving // Moderate activity suggests signal reception

	case calciumActivity > recoveryThreshold:
		// Low membrane potential but elevated calcium suggests recovery
		return PhaseRecovery

	case math.Abs(accumulator) > 0.001 || calciumActivity > 0.001:
		// Very low but detectable activity suggests receiving phase
		return PhaseReceiving

	default:
		// Minimal activity indicates resting state
		return PhaseIdle
	}
}

// isProcessingActive determines if neuron is actively processing based on biological indicators
func (m *BasicProcessingMonitor) isProcessingActive(accumulator, calcium float64) bool {
	// Biological activity thresholds - more conservative
	const (
		accumulatorThreshold = 0.01 // Require more significant charge accumulation
		calciumThreshold     = 0.01 // Require more significant calcium elevation
	)

	// Both conditions should be met for true "active processing"
	significantAccumulator := math.Abs(accumulator) > accumulatorThreshold
	significantCalcium := calcium > calciumThreshold

	// Require either significant electrical activity OR significant biochemical activity
	return significantAccumulator || significantCalcium
}

// isNeuronQuiescent determines if neuron has reached stable resting state
// Models glial cell detection of neural quiescence through multiple biological indicators
//
// BIOLOGICAL QUIESCENCE INDICATORS:
// - Membrane potential at or near resting level
// - Calcium concentration at baseline levels
// - Stable electrical activity (no fluctuations)
// - Normal metabolic rate
// - Absence of neurotransmitter release
func (m *BasicProcessingMonitor) isNeuronQuiescent(neuron NeuronInterface) bool {
	accumulator := neuron.GetAccumulator()
	calcium := neuron.GetCalciumLevel()

	// Biological quiescence thresholds
	const (
		maxAccumulator = 0.01 // Very low membrane potential deviation
		maxCalcium     = 0.01 // Baseline calcium levels
	)

	// Check multiple quiescence indicators
	accumulatorQuiet := math.Abs(accumulator) <= maxAccumulator
	calciumQuiet := calcium <= maxCalcium

	// Additional check: verify neuron is not currently processing
	neuronID := neuron.ID()
	isProcessing := m.IsNeuronProcessing(neuronID)

	return accumulatorQuiet && calciumQuiet && !isProcessing
}

// updateProcessingState updates the processing state for a specific neuron
// Models glial cell recording of neural state changes in monitoring records
func (m *BasicProcessingMonitor) updateProcessingState(neuronID string, phase ProcessingPhase, messageID uint64, startTime time.Time) {
	m.statesMutex.Lock()
	defer m.statesMutex.Unlock()

	state, exists := m.processingStates[neuronID]
	if !exists {
		// Create new state if neuron not previously tracked
		state = &ProcessingState{
			NeuronID: neuronID,
		}
		m.processingStates[neuronID] = state
	}

	// Update state information
	state.Phase = phase
	state.IsProcessing = (phase != PhaseIdle)
	state.MessageID = messageID
	state.StartTime = startTime
	state.LastActivity = time.Now()
}

// checkPendingMessages monitors active message processing and detects completion
// Models glial cell tracking of individual synaptic events through neural processing
func (m *BasicProcessingMonitor) checkPendingMessages() {
	m.messagesMutex.RLock()
	pendingCopy := make(map[uint64]*processingTracker, len(m.pendingMessages))
	for id, tracker := range m.pendingMessages {
		pendingCopy[id] = tracker
	}
	m.messagesMutex.RUnlock()

	// Check each pending message for completion
	for messageID, tracker := range pendingCopy {
		if m.isMessageProcessingComplete(tracker) {
			m.completeMessageProcessing(messageID, tracker)
		}
	}
}

// isMessageProcessingComplete determines if a tracked message has finished processing
// Models glial cell detection of processing completion through activity monitoring
func (m *BasicProcessingMonitor) isMessageProcessingComplete(tracker *processingTracker) bool {
	// Get current neuron state
	state, err := m.GetProcessingState(tracker.neuronID)
	if err != nil {
		return true // Neuron no longer monitored - consider complete
	}

	// Get current neuron to check detailed state
	m.neuronsMutex.RLock()
	neuron, exists := m.monitoredNeurons[tracker.neuronID]
	m.neuronsMutex.RUnlock()

	if !exists {
		return true // Neuron no longer monitored
	}

	// Check for completion indicators
	processingTime := time.Since(tracker.startTime)

	// Get current neural activity levels
	accumulator := neuron.GetAccumulator()
	calcium := neuron.GetCalciumLevel()

	// Adaptive completion criteria based on signal strength and time
	const (
		minProcessingTime       = 5 * time.Millisecond   // Minimum time for realistic signal integration
		maxProcessingTime       = 200 * time.Millisecond // Maximum reasonable processing time
		strictActivityThreshold = 0.0001                 // Very strict threshold for strong signals
		weakSignalThreshold     = 0.0002                 // Threshold for detecting weak signal processing
		weakSignalTimeout       = 50 * time.Millisecond  // Timeout for weak signals
		firingCalciumThreshold  = 0.5                    // Calcium level indicating recent firing
		postFireTimeout         = 20 * time.Millisecond  // Time to wait after firing for completion
	)

	// Must have minimum processing time
	if processingTime < minProcessingTime {
		return false
	}

	// If too much time has passed, consider it complete (timeout protection)
	if processingTime > maxProcessingTime {
		return true
	}

	// Check if neuron fired (high calcium indicates recent action potential)
	neuronFired := calcium > firingCalciumThreshold

	if neuronFired {
		// Neuron fired - completion means recovery from firing
		// After firing: accumulator resets to 0, but calcium remains elevated
		// Wait for calcium to start declining and enough time for refractory period
		accumulatorReset := math.Abs(accumulator) < 0.001 // Should be near 0 after firing
		sufficientRecoveryTime := processingTime >= postFireTimeout

		return accumulatorReset && sufficientRecoveryTime
	}

	// Determine if this was a weak signal based on the original message strength
	originalSignal := tracker.message.Value
	isWeakSignal := originalSignal < 0.7 // Signals below 0.7 are considered weak

	// For weak signals that didn't fire, use more lenient criteria
	if isWeakSignal {
		// Check if any neural activity occurred (indicating signal was processed)
		anyActivity := math.Abs(accumulator) > weakSignalThreshold || calcium > weakSignalThreshold

		// For weak signals: either return to very quiet state OR timeout after reasonable time
		if anyActivity {
			// Activity occurred, wait for settling
			isSettled := math.Abs(accumulator) <= strictActivityThreshold && calcium <= strictActivityThreshold
			return isSettled && processingTime >= minProcessingTime
		} else {
			// No significant activity, but enough time passed for weak signal processing
			return processingTime >= weakSignalTimeout
		}
	} else {
		// For stronger signals that didn't fire, use strict criteria
		isReallyQuiescent := math.Abs(accumulator) <= strictActivityThreshold &&
			calcium <= strictActivityThreshold

		hasReturnedToIdle := state.Phase == PhaseIdle
		notActivelyProcessing := !state.IsProcessing

		// All conditions must be met for strong signal completion
		return isReallyQuiescent && hasReturnedToIdle && notActivelyProcessing && processingTime >= minProcessingTime
	}
}

// completeMessageProcessing finalizes tracking for a completed message
// Notifies all waiters and cleans up tracking resources
func (m *BasicProcessingMonitor) completeMessageProcessing(messageID uint64, tracker *processingTracker) {
	// Calculate final processing metrics
	processingTime := time.Since(tracker.startTime)

	// Get final neuron state
	finalState, _ := m.GetProcessingState(tracker.neuronID)

	// Create completion result
	result := processingResult{
		success:        true,
		finalState:     finalState,
		processingTime: processingTime,
		error:          nil,
	}

	// Notify all waiters
	m.messagesMutex.Lock()
	waiters := m.messageWaiters[messageID]
	delete(m.messageWaiters, messageID)
	delete(m.pendingMessages, messageID)
	m.messagesMutex.Unlock()

	// Send completion notification to all waiters
	for _, waiter := range waiters {
		select {
		case waiter <- result:
		default:
			// Channel full or closed - skip this waiter
		}
	}
}

// =================================================================================
// UTILITY FUNCTIONS FOR BIOLOGICAL PARAMETER CONFIGURATION
// =================================================================================

// CreateDefaultProcessingMonitorConfig returns biologically-realistic configuration
// Based on known parameters of glial cell monitoring capabilities and timescales
func CreateDefaultProcessingMonitorConfig() *ProcessingMonitorConfig {
	return &ProcessingMonitorConfig{
		ActivityThreshold:   0.001,                 // Sensitive to minimal neural activity
		StateUpdateInterval: 10 * time.Millisecond, // Real-time monitoring frequency
		MaxMonitoredNeurons: 100,                   // Reasonable glial monitoring capacity
		ProcessingTimeout:   1 * time.Second,       // Generous timeout for processing detection
		QuiescenceTimeout:   5 * time.Second,       // Extended time for quiescence detection
	}
}

// CreateHighSensitivityConfig returns configuration for maximum monitoring sensitivity
// Models highly activated glial cells during periods of intense neural activity
func CreateHighSensitivityConfig() *ProcessingMonitorConfig {
	return &ProcessingMonitorConfig{
		ActivityThreshold:   0.0001,                 // Maximum sensitivity
		StateUpdateInterval: 1 * time.Millisecond,   // Highest monitoring frequency
		MaxMonitoredNeurons: 50,                     // Reduced capacity for detailed monitoring
		ProcessingTimeout:   500 * time.Millisecond, // Faster timeout for rapid detection
		QuiescenceTimeout:   2 * time.Second,        // Shorter quiescence detection
	}
}

// CreateLowOverheadConfig returns configuration optimized for minimal computational impact
// Models quiescent glial cells during periods of low neural activity
func CreateLowOverheadConfig() *ProcessingMonitorConfig {
	return &ProcessingMonitorConfig{
		ActivityThreshold:   0.01,                   // Reduced sensitivity
		StateUpdateInterval: 100 * time.Millisecond, // Lower monitoring frequency
		MaxMonitoredNeurons: 200,                    // Higher capacity with less detailed monitoring
		ProcessingTimeout:   5 * time.Second,        // Extended timeout
		QuiescenceTimeout:   10 * time.Second,       // Extended quiescence detection
	}
}
