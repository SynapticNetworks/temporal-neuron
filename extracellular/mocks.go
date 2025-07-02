/*
=================================================================================
ENHANCED MOCK IMPLEMENTATIONS FOR FACTORY PATTERN TESTING
=================================================================================

Updated mock neurons and synapses that implement the new factory interfaces
while maintaining compatibility with existing tests. These mocks demonstrate
how to properly implement the biological interfaces for complete decoupling.

BIOLOGICAL MODELING:
These mocks implement realistic biological behaviors including:
- Chemical receptor expression and neurotransmitter binding
- Electrical signal processing through gap junctions
- Spatial positioning and activity monitoring
- Lifecycle management and health reporting

=================================================================================
*/

package extracellular

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// =================================================================================
// ENHANCED MOCK NEURON WITH FACTORY PATTERN SUPPORT
// =================================================================================

// MockNeuron represents a biologically-inspired test neuron that implements all
// required interfaces for factory pattern creation and biological coordination.
//
// BIOLOGICAL PROPERTIES MODELED:
// - Membrane potential and threshold dynamics
// - Neurotransmitter receptor expression
// - Electrical coupling through gap junctions
// - Metabolic activity and health status
// - Spatial positioning in 3D neural tissue
type MockNeuron struct {
	// === CORE NEURAL IDENTITY ===
	id            string              // Unique biological identifier
	position      types.Position3D    // 3D spatial location in neural tissue
	componentType types.ComponentType // Component classification

	// === RECEPTOR EXPRESSION PROFILE ===
	receptors []types.LigandType // Neurotransmitter receptors expressed

	// === ELECTRICAL PROPERTIES ===
	threshold        float64 // Action potential threshold
	currentPotential float64 // Current membrane potential
	isActive         bool    // Neural activity state

	// === CONNECTIVITY AND SIGNALING ===
	connections []string           // Connected component IDs
	signalTypes []types.SignalType // Electrical signal types processed

	// === BIOLOGICAL ACTIVITY MONITORING ===
	activityLevel   float64 // Recent activity rate (0.0-1.0)
	connectionCount int     // Number of synaptic connections

	// === CHEMICAL SIGNALING TRACKING ===
	bindingEventCount int                  // Total chemical binding events
	bindingHistory    []types.BindingEvent // Detailed binding event log

	// === BIOLOGICAL CALLBACKS (INJECTED BY MATRIX) ===
	callbacks component.NeuronCallbacks // Matrix-provided biological functions

	// === LIFECYCLE STATE ===
	isStarted bool // Whether neuron is actively processing

	// === THREAD SAFETY ===
	mu sync.RWMutex // Protects concurrent access to neuron state

	outputCallbacks map[string]types.OutputCallback
}

// =================================================================================
// FACTORY CONSTRUCTOR FOR MOCK NEURONS
// =================================================================================

// NewMockNeuron creates a mock neuron with biological properties.
//
// BIOLOGICAL INITIALIZATION:
// This constructor sets up a neuron with realistic biological parameters:
// - Spatial positioning in 3D neural tissue
// - Receptor expression profile for chemical signaling
// - Electrical properties for action potential generation
// - Activity monitoring for health assessment
func NewMockNeuron(id string, pos types.Position3D, receptors []types.LigandType) *MockNeuron {
	return &MockNeuron{
		id:                id,
		position:          pos,
		componentType:     types.TypeNeuron,
		receptors:         receptors,
		threshold:         0.7,  // Biological action potential threshold
		currentPotential:  0.0,  // Resting potential
		isActive:          true, // Start in active state
		connections:       make([]string, 0),
		signalTypes:       []types.SignalType{types.SignalFired, types.SignalConnected}, // Default signal responsiveness
		activityLevel:     0.0,                                                          // No initial activity
		connectionCount:   0,                                                            // No initial connections
		bindingEventCount: 0,
		bindingHistory:    make([]types.BindingEvent, 0),
		callbacks:         nil, // Injected during factory creation
		isStarted:         false,
		outputCallbacks:   make(map[string]types.OutputCallback),
	}
}

// =================================================================================
// CORE COMPONENT INTERFACE IMPLEMENTATION
// =================================================================================

// ID returns the unique biological identifier of this neuron
func (mn *MockNeuron) ID() string {
	return mn.id
}

// Position returns the 3D spatial location of this neuron in neural tissue
func (mn *MockNeuron) Position() types.Position3D {
	return mn.position
}

// IsActive returns whether the neuron is currently processing signals
func (mn *MockNeuron) IsActive() bool {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	return mn.isActive
}

// GetMetadata returns component metadata for biological analysis
func (mn *MockNeuron) GetMetadata() map[string]interface{} {
	mn.mu.RLock()
	defer mn.mu.RUnlock()

	return map[string]interface{}{
		"neuron_type":      "mock_neuron",
		"receptor_count":   len(mn.receptors),
		"connection_count": len(mn.connections),
		"activity_level":   mn.activityLevel,
		"threshold":        mn.threshold,
	}
}

func (mn *MockNeuron) Receive(msg types.NeuralSignal) {
	// NO NEEDED - just to fullfill interface
}

func (mn *MockNeuron) ScheduleDelayedDelivery(msg types.NeuralSignal, target component.MessageReceiver, delay time.Duration) {
	// NO NEEDED - just to fullfill interface
}

// =================================================================================
// NEURON INTERFACE IMPLEMENTATION
// =================================================================================

// GetThreshold returns the current action potential threshold
func (mn *MockNeuron) GetThreshold() float64 {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	return mn.threshold
}

// SetThreshold modifies the action potential threshold (homeostatic plasticity)
func (mn *MockNeuron) SetThreshold(threshold float64) {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	mn.threshold = threshold
}

// GetActivityLevel returns recent neural activity rate for health monitoring
func (mn *MockNeuron) GetActivityLevel() float64 {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	return mn.activityLevel
}

// GetConnectionCount returns number of synaptic connections
func (mn *MockNeuron) GetConnectionCount() int {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	return mn.connectionCount
}

// =================================================================================
// LIFECYCLE MANAGEMENT
// =================================================================================

// Start begins neural processing and biological activity
func (mn *MockNeuron) Start() error {
	mn.mu.Lock()
	defer mn.mu.Unlock()

	mn.isStarted = true
	mn.isActive = true
	return nil
}

// Stop gracefully shuts down neural processing
func (mn *MockNeuron) Stop() error {
	mn.mu.Lock()
	defer mn.mu.Unlock()

	mn.isStarted = false
	mn.isActive = false
	return nil
}

// Start begins neural processing and biological activity
func (mn *MockNeuron) CanRestart() bool {
	return true
}

// Add these missing methods to MockNeuron to implement component.NeuralComponent

// GetLastActivity returns the timestamp of the most recent activity
func (mn *MockNeuron) GetLastActivity() time.Time {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	// Return a recent timestamp based on activity
	return time.Now().Add(-time.Duration(int64((1.0 - mn.activityLevel) * float64(time.Hour))))
}

// State returns the current operational state
func (mn *MockNeuron) State() types.ComponentState {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	if mn.isActive {
		return types.StateActive
	}
	return types.StateInactive
}

// SetState sets the operational state
func (mn *MockNeuron) SetState(state types.ComponentState) {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	mn.isActive = (state == types.StateActive)
}

// SetPosition updates the 3D spatial coordinates
func (mn *MockNeuron) SetPosition(position types.Position3D) {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	mn.position = position
}

// UpdateMetadata sets or updates a metadata key-value pair
func (mn *MockNeuron) UpdateMetadata(key string, value interface{}) {
	// MockNeuron doesn't store dynamic metadata, so this is a no-op
	// In a real implementation, you'd store this in a metadata map
}

// Restart attempts to reactivate the component
func (mn *MockNeuron) Restart() error {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	mn.isActive = true
	mn.isStarted = true
	return nil
}

// =================================================================================
// CHEMICAL SIGNALING (BINDING TARGET INTERFACE)
// =================================================================================

// Bind processes incoming chemical signals (neurotransmitter binding)
//
// BIOLOGICAL RECEPTOR BINDING:
// This models the complex process of neurotransmitter binding to membrane receptors:
// 1. Receptor specificity check (only bind if neuron expresses this receptor type)
// 2. Concentration-dependent response (stronger signals produce larger effects)
// 3. Integration with membrane potential (affecting firing probability)
// 4. Event logging for biological analysis and plasticity mechanisms
//
// NEUROTRANSMITTER EFFECTS MODELED:
// - Glutamate: Excitatory (increases membrane potential)
// - GABA: Inhibitory (decreases membrane potential)
// - Dopamine: Neuromodulatory (affects plasticity and excitability)
// - Serotonin: Neuromodulatory (influences mood and arousal)
// - Acetylcholine: Mixed effects (attention and arousal)
func (mn *MockNeuron) Bind(ligandType types.LigandType, sourceID string, concentration float64) {
	mn.mu.Lock()
	defer mn.mu.Unlock()

	// Check receptor expression - neurons only respond to neurotransmitters
	// for which they express appropriate receptors
	hasReceptor := false
	for _, receptor := range mn.receptors {
		if receptor == ligandType {
			hasReceptor = true
			break
		}
	}

	if !hasReceptor {
		return // No receptor for this neurotransmitter - no response
	}

	// Record binding event for biological analysis
	mn.bindingEventCount++
	mn.bindingHistory = append(mn.bindingHistory, types.BindingEvent{
		LigandType:    ligandType,
		SourceID:      sourceID,
		Concentration: concentration,
		Timestamp:     time.Now(),
	})

	// Apply neurotransmitter-specific effects on membrane potential
	switch ligandType {
	case types.LigandGlutamate:
		// Excitatory effect: AMPA/NMDA receptor activation
		mn.currentPotential += concentration * 1.2 // Stronger excitatory response

	case types.LigandGABA:
		// Inhibitory effect: GABA-A receptor activation (chloride influx)
		mn.currentPotential -= concentration * 1.0 // Stronger inhibitory response

	case types.LigandDopamine:
		// Neuromodulatory effect: D1/D2 receptor activation
		mn.currentPotential += concentration * 0.8 // Enhanced excitatory modulation

	case types.LigandSerotonin:
		// Neuromodulatory effect: 5-HT receptor activation
		mn.currentPotential += concentration * 0.5 // Enhanced modulation

	case types.LigandAcetylcholine:
		// Mixed nicotinic/muscarinic effects
		mn.currentPotential += concentration * 0.9 // Enhanced excitatory effect
	}

	// Update activity level based on membrane potential changes
	if mn.currentPotential > mn.threshold && mn.isActive {
		// Neuron fires - reset potential but maintain some activation
		mn.activityLevel = 1.0     // Maximum activity during firing
		mn.currentPotential *= 0.7 // Partial reset (refractory period)

		// Report firing through electrical signaling if callbacks are available
		if mn.callbacks != nil {
			mn.callbacks.SendElectricalSignal(types.SignalFired, 1.0)
		}
	} else {
		// Update activity level based on subthreshold activity
		potentialRatio := mn.currentPotential / mn.threshold
		if potentialRatio > 0 {
			mn.activityLevel = potentialRatio * 0.5 // Subthreshold activity
		}
	}

	// Natural decay of membrane potential (leaky integration)
	mn.currentPotential *= 0.98 // Reduced decay to maintain response for validation

	// Report health to microglia if callbacks are available
	if mn.callbacks != nil {
		mn.callbacks.ReportHealth(mn.activityLevel, mn.connectionCount)
	}
}

// GetReceptors returns the neurotransmitter receptors expressed by this neuron
func (mn *MockNeuron) GetReceptors() []types.LigandType {
	mn.mu.RLock()
	defer mn.mu.RUnlock()

	// Return defensive copy to prevent external modification
	receptors := make([]types.LigandType, len(mn.receptors))
	copy(receptors, mn.receptors)
	return receptors
}

// GetPosition returns the 3D spatial position for chemical signaling calculations
func (mn *MockNeuron) GetPosition() types.Position3D {
	return mn.position
}

// =================================================================================
// ELECTRICAL SIGNALING (SIGNAL LISTENER INTERFACE)
// =================================================================================

// OnSignal processes incoming electrical signals through gap junctions
//
// BIOLOGICAL ELECTRICAL COUPLING:
// This models direct electrical communication between neurons through gap junctions:
// 1. Current injection from coupled neurons (gap junction conductance)
// 2. Synchronization signals for network coordination
// 3. State change notifications from network components
// 4. Connection establishment signals for development
//
// ELECTRICAL SIGNAL PROCESSING:
// - Immediate membrane potential effects (current injection)
// - Network synchronization responses
// - Connectivity updates for circuit formation
func (mn *MockNeuron) OnSignal(signalType types.SignalType, sourceID string, data interface{}) {
	mn.mu.Lock()
	defer mn.mu.Unlock()

	// Only process signals if neuron is active
	if !mn.isActive {
		return
	}

	switch signalType {
	case types.SignalFired:
		// Another neuron fired - gap junction current injection
		if value, ok := data.(float64); ok {
			// Model gap junction conductance - smaller effect than chemical synapses
			mn.currentPotential += value * 0.2 // Gap junction coupling strength

			// Update activity level
			if mn.currentPotential > mn.threshold*0.8 { // Close to threshold
				mn.activityLevel = 0.8
			}
		}

	case types.SignalConnected:
		// New connection established
		if connID, ok := data.(string); ok {
			// Add to connection list if not already present
			for _, existing := range mn.connections {
				if existing == connID {
					return // Already connected
				}
			}
			mn.connections = append(mn.connections, connID)
			mn.connectionCount = len(mn.connections)
		}

	case types.SignalDisconnected:
		// Connection removed
		if connID, ok := data.(string); ok {
			for i, existing := range mn.connections {
				if existing == connID {
					// Remove connection
					mn.connections = append(mn.connections[:i], mn.connections[i+1:]...)
					mn.connectionCount = len(mn.connections)
					break
				}
			}
		}

	case types.SignalThresholdChanged:
		// Network-wide threshold adjustment (homeostatic plasticity)
		if adjustment, ok := data.(float64); ok {
			mn.threshold += adjustment
			// Ensure threshold stays in biological range
			if mn.threshold < 0.1 {
				mn.threshold = 0.1
			} else if mn.threshold > 2.0 {
				mn.threshold = 2.0
			}
		}
	}
}

// =================================================================================
// CALLBACK INJECTION (FACTORY PATTERN SUPPORT)
// =================================================================================

// SetCallbacks injects biological functions provided by the matrix
// This is called during factory creation to wire the neuron into matrix systems
func (mn *MockNeuron) SetCallbacks(callbacks component.NeuronCallbacks) {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	mn.callbacks = callbacks
}

// =================================================================================
// TESTING AND ANALYSIS UTILITIES
// =================================================================================

// GetCurrentPotential returns the current membrane potential for testing
func (mn *MockNeuron) GetCurrentPotential() float64 {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	return mn.currentPotential
}

// SetPotential directly sets membrane potential (for testing)
func (mn *MockNeuron) SetPotential(potential float64) {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	mn.currentPotential = potential
}

// GetConnections returns a copy of current connections
func (mn *MockNeuron) GetConnections() []string {
	mn.mu.RLock()
	defer mn.mu.RUnlock()

	connections := make([]string, len(mn.connections))
	copy(connections, mn.connections)
	return connections
}

// SetActive controls neuron active state (for testing)
func (mn *MockNeuron) SetActive(active bool) {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	mn.isActive = active
}

// GetBindingEventCount returns total chemical binding events
func (mn *MockNeuron) GetBindingEventCount() int {
	mn.mu.RLock()
	defer mn.mu.RUnlock()
	return mn.bindingEventCount
}

// GetBindingHistory returns a copy of all binding events
func (mn *MockNeuron) GetBindingHistory() []types.BindingEvent {
	mn.mu.RLock()
	defer mn.mu.RUnlock()

	history := make([]types.BindingEvent, len(mn.bindingHistory))
	copy(history, mn.bindingHistory)
	return history
}

// ResetBindingEvents clears binding event history (for testing)
func (mn *MockNeuron) ResetBindingEvents() {
	mn.mu.Lock()
	defer mn.mu.Unlock()
	mn.bindingEventCount = 0
	mn.bindingHistory = mn.bindingHistory[:0]
}

// ComponentType returns the biological classification
func (mn *MockNeuron) Type() types.ComponentType {
	return mn.componentType
}

func (mn *MockNeuron) AddOutputCallback(synapseID string, callback types.OutputCallback) {
	mn.mu.Lock()
	defer mn.mu.Unlock()

	if mn.outputCallbacks == nil {
		mn.outputCallbacks = make(map[string]types.OutputCallback)
	}

	mn.outputCallbacks[synapseID] = callback

	// For testing, track connections using existing connections slice
	for _, existing := range mn.connections {
		if existing == synapseID {
			return // Already tracked
		}
	}
	mn.connections = append(mn.connections, synapseID)
	mn.connectionCount = len(mn.connections)
}

func (mn *MockNeuron) RemoveOutputCallback(synapseID string) {
	mn.mu.Lock()
	defer mn.mu.Unlock()

	delete(mn.outputCallbacks, synapseID)

	// Remove from connections tracking
	for i, conn := range mn.connections {
		if conn == synapseID {
			mn.connections = append(mn.connections[:i], mn.connections[i+1:]...)
			mn.connectionCount = len(mn.connections)
			break
		}
	}
}

// ADD: Method to trigger output callbacks (simulate neuron firing)
func (mn *MockNeuron) FireAndTransmit(signal float64) {
	mn.mu.RLock()
	callbacks := make(map[string]types.OutputCallback)
	for id, callback := range mn.outputCallbacks {
		callbacks[id] = callback
	}
	mn.mu.RUnlock()

	// Fire to all connected synapses
	message := types.NeuralSignal{
		Value:     signal,
		Timestamp: time.Now(),
		SourceID:  mn.id,
	}

	for synapseID, callback := range callbacks {
		if err := callback.TransmitMessage(message); err != nil {
			// Log error but continue with other synapses
			fmt.Printf("Failed to transmit to synapse %s: %v\n", synapseID, err)
		}
	}

	// Track firing activity using existing fields
	mn.mu.Lock()
	mn.activityLevel += 0.1 // Increase activity level
	if mn.activityLevel > 1.0 {
		mn.activityLevel = 1.0
	}
	mn.mu.Unlock()
}

// =================================================================================
// ENHANCED MOCK SYNAPSE WITH FACTORY PATTERN SUPPORT
// =================================================================================

// MockSynapse represents a biologically-inspired test synapse that implements all
// required interfaces for factory pattern creation and synaptic coordination.
//
// BIOLOGICAL PROPERTIES MODELED:
// - Presynaptic and postsynaptic neuron connectivity
// - Synaptic weight and activity-dependent plasticity
// - Neurotransmitter release and spatial positioning
// - Transmission delays and biological timing
type MockSynapse struct {
	// === CORE SYNAPTIC IDENTITY ===
	id            string              // Unique biological identifier
	position      types.Position3D    // 3D spatial location in neural tissue
	componentType types.ComponentType // Component classification

	// === SYNAPTIC CONNECTIVITY ===
	presynapticID  string // Source neuron identifier
	postsynapticID string // Target neuron identifier

	// === SYNAPTIC PROPERTIES ===
	weight     float64          // Current synaptic strength
	delay      time.Duration    // Transmission delay
	ligandType types.LigandType // Neurotransmitter type

	// === PLASTICITY CONFIGURATION ===
	plasticityEnabled bool                   // Whether plasticity is active
	plasticityConfig  types.PlasticityConfig // Plasticity parameters

	// === ACTIVITY MONITORING ===
	activity          float64   // Recent synaptic activity
	lastActivity      time.Time // Most recent activity
	lastTransmission  time.Time // Most recent transmission
	transmissionCount int64     // Total transmissions

	// === BIOLOGICAL CALLBACKS ===
	callbacks *SynapseCallbacks // Matrix-provided biological functions

	// === LIFECYCLE STATE ===
	isActive bool // Whether synapse is functional

	// === THREAD SAFETY ===
	mu sync.RWMutex // Protects concurrent access
}

// =================================================================================
// FACTORY CONSTRUCTOR FOR MOCK SYNAPSES
// =================================================================================

// NewMockSynapse creates a mock synapse with biological properties
func NewMockSynapse(id string, pos types.Position3D, pre, post string, weight float64) *MockSynapse {
	return &MockSynapse{
		id:                id,
		position:          pos,
		componentType:     types.TypeSynapse,
		presynapticID:     pre,
		postsynapticID:    post,
		weight:            weight,
		delay:             time.Millisecond,      // Default 1ms synaptic delay
		ligandType:        types.LigandGlutamate, // Default to excitatory
		plasticityEnabled: true,
		plasticityConfig: types.PlasticityConfig{
			Enabled:        true,
			LearningRate:   0.01,
			TimeConstant:   10 * time.Millisecond,
			WindowSize:     20 * time.Millisecond,
			MinWeight:      0.0,
			MaxWeight:      2.0,
			AsymmetryRatio: 1.0,
		},
		activity:          0.0,
		lastActivity:      time.Now(),
		lastTransmission:  time.Now(),
		transmissionCount: 0,
		callbacks:         nil,
		isActive:          true,
	}
}

// =================================================================================
// CORE COMPONENT INTERFACE IMPLEMENTATION
// =================================================================================

// ID returns the unique biological identifier
func (ms *MockSynapse) ID() string {
	return ms.id
}

// Position returns the 3D spatial location
func (ms *MockSynapse) Position() types.Position3D {
	return ms.position
}

// Type returns the biological classification
func (ms *MockSynapse) Type() types.ComponentType {
	return ms.componentType
}

// IsActive returns whether the synapse is functional
func (ms *MockSynapse) IsActive() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.isActive
}

// GetMetadata returns synapse metadata
func (ms *MockSynapse) GetMetadata() map[string]interface{} {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return map[string]interface{}{
		"synapse_type":       "mock_synapse",
		"presynaptic_id":     ms.presynapticID,
		"postsynaptic_id":    ms.postsynapticID,
		"weight":             ms.weight,
		"ligand_type":        ms.ligandType,
		"plasticity_enabled": ms.plasticityEnabled,
		"transmission_count": ms.transmissionCount,
	}
}

// =================================================================================
// SYNAPSE INTERFACE IMPLEMENTATION
// =================================================================================

// GetPresynapticID returns the source neuron identifier
func (ms *MockSynapse) GetPresynapticID() string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.presynapticID
}

// GetPostsynapticID returns the target neuron identifier
func (ms *MockSynapse) GetPostsynapticID() string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.postsynapticID
}

// GetWeight returns the current synaptic strength
func (ms *MockSynapse) GetWeight() float64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.weight
}

// SetWeight modifies the synaptic strength
func (ms *MockSynapse) SetWeight(weight float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Apply biological bounds
	if weight < ms.plasticityConfig.MinWeight {
		weight = ms.plasticityConfig.MinWeight
	} else if weight > ms.plasticityConfig.MaxWeight {
		weight = ms.plasticityConfig.MaxWeight
	}

	ms.weight = weight
}

// GetPlasticityConfig returns plasticity parameters
func (ms *MockSynapse) GetPlasticityConfig() types.PlasticityConfig {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.plasticityConfig
}

// GetActivityInfo returns activity information using types.ActivityInfo struct
func (ms *MockSynapse) GetActivityInfo() types.ActivityInfo {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	now := time.Now()

	return types.ActivityInfo{
		ComponentID:           ms.id,
		LastTransmission:      ms.lastTransmission,
		LastPlasticity:        now, // Mock doesn't track plasticity events yet
		Weight:                ms.weight,
		ActivityLevel:         0.0, // Mock default
		TimeSinceTransmission: now.Sub(ms.lastTransmission),
		TimeSincePlasticity:   time.Duration(0), // Mock doesn't track this
		ConnectionCount:       0,                // Not applicable for synapses
	}
}

// GetLastActivity returns the timestamp of the most recent transmission
func (ms *MockSynapse) GetLastActivity() time.Time {
	return ms.GetActivityInfo().LastTransmission
}

// =================================================================================
// SYNAPTIC TRANSMISSION
// =================================================================================

// Transmit sends a synaptic message with biological realism
func (ms *MockSynapse) Transmit(msg float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Update activity tracking
	ms.activity = msg * ms.weight
	ms.lastActivity = time.Now()
	ms.lastTransmission = time.Now()
	ms.transmissionCount++

	// Calculate transmission delay if callbacks available
	var totalDelay time.Duration = ms.delay
	if ms.callbacks != nil {
		totalDelay = ms.callbacks.GetTransmissionDelay()
	}

	// Create enhanced message with synaptic processing
	signal := types.NeuralSignal{
		Value:                ms.activity,
		Timestamp:            time.Now().Add(totalDelay),
		SourceID:             ms.presynapticID,
		SynapseID:            ms.id,
		TargetID:             ms.postsynapticID,
		NeurotransmitterType: ms.ligandType,
	}

	// Deliver message with delay if callbacks available
	if ms.callbacks != nil {
		// Schedule delivery after delay
		go func() {
			time.Sleep(totalDelay)
			ms.callbacks.DeliverMessage(ms.postsynapticID, signal)

			// Release neurotransmitter into synaptic cleft
			concentration := ms.activity * 0.1 // Convert activity to concentration
			ms.callbacks.ReleaseNeurotransmitter(ms.ligandType, concentration)

			// Report activity
			ms.callbacks.ReportActivity(types.SynapticActivity{
				SynapseID:      ms.id,
				Timestamp:      time.Now(),
				MessageValue:   msg,
				CurrentWeight:  ms.weight,
				ActivityType:   "transmission",
				PresynapticID:  ms.presynapticID,
				PostsynapticID: ms.postsynapticID,
			})
		}()
	}

}

func (ms *MockSynapse) ApplyPlasticity(adjustment types.PlasticityAdjustment) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Convert PlasticityAdjustment to PlasticityEvent and use existing UpdateWeight
	plasticityEvent := types.PlasticityEvent{
		EventType: types.PlasticitySTDP, // Default to STDP
		Timestamp: adjustment.Timestamp,
		Strength:  adjustment.WeightChange, // Assuming this field exists in PlasticityAdjustment
		SourceID:  ms.id,
		// Set PreTime and PostTime based on adjustment if available
		PreTime:  adjustment.Timestamp,
		PostTime: adjustment.Timestamp,
	}

	// Use existing UpdateWeight method
	ms.UpdateWeight(plasticityEvent)

}

func (msp *MockSynapse) GetDelay() time.Duration {
	return 0 * time.Second
}

// ShouldPrune implements component.SynapticProcessor interface
func (ms *MockSynapse) ShouldPrune() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Simple pruning logic: prune if weight is too low or inactive for too long
	if ms.weight < ms.plasticityConfig.MinWeight*0.1 {
		return true // Weight too low
	}

	// Check for inactivity (mock implementation)
	timeSinceActivity := time.Since(ms.lastActivity)
	if timeSinceActivity > 30*time.Second { // 30 seconds without activity
		return true
	}

	return false // Don't prune
}

// UpdateWeight modifies synaptic strength based on plasticity events
func (ms *MockSynapse) UpdateWeight(event types.PlasticityEvent) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if !ms.plasticityEnabled {
		return
	}

	// Apply plasticity rule based on event type
	switch event.EventType {
	case types.PlasticitySTDP:
		// Spike-Timing Dependent Plasticity
		timeDiff := event.PostTime.Sub(event.PreTime)
		if timeDiff > 0 && timeDiff < ms.plasticityConfig.WindowSize {
			// Potentiation: post after pre
			weightChange := ms.plasticityConfig.LearningRate * event.Strength
			ms.weight += weightChange
		} else if timeDiff < 0 && timeDiff > -ms.plasticityConfig.WindowSize {
			// Depression: pre after post
			weightChange := ms.plasticityConfig.LearningRate * event.Strength * 0.5
			ms.weight -= weightChange
		}

	case types.PlasticityBCM:
		// BCM (Bienenstock-Cooper-Munro) plasticity
		// Implementation would depend on postsynaptic activity history

	case types.PlasticityOja:
		// Oja's learning rule
		// Implementation would normalize weights

	case types.PlasticityHomeostatic:
		// Homeostatic plasticity
		ms.weight *= event.Strength // Scaling factor
	}

	// Apply weight bounds
	if ms.weight < ms.plasticityConfig.MinWeight {
		ms.weight = ms.plasticityConfig.MinWeight
	} else if ms.weight > ms.plasticityConfig.MaxWeight {
		ms.weight = ms.plasticityConfig.MaxWeight
	}

	// Report plasticity event if callbacks available
	if ms.callbacks != nil {
		ms.callbacks.ReportPlasticityEvent(event)
	}
}

// =================================================================================
// CALLBACK INJECTION
// =================================================================================

// SetCallbacks injects biological functions provided by the matrix
func (ms *MockSynapse) SetCallbacks(callbacks SynapseCallbacks) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.callbacks = &callbacks
}

// =================================================================================
// TESTING UTILITIES
// =================================================================================

// GetActivity returns current synaptic activity
func (ms *MockSynapse) GetActivity() float64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.activity
}

// SetActivity directly sets activity level (for testing)
func (ms *MockSynapse) SetActivity(activity float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.activity = activity
}

// GetTransmissionCount returns total number of transmissions
func (ms *MockSynapse) GetTransmissionCount() int64 {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.transmissionCount
}

// SetLigandType sets the neurotransmitter type (for testing)
func (ms *MockSynapse) SetLigandType(ligandType types.LigandType) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.ligandType = ligandType
}

// =================================================================================
// MOCK ASTROCYTE LISTENER FOR CALCIUM WAVE TESTING
// =================================================================================

// mockAstrocyteListener implements SignalListener for astrocyte calcium testing
type mockAstrocyteListener struct {
	id            string
	receivedCount int32
	receivedData  []interface{}
	receivedFrom  []string
	mu            sync.Mutex
}

func newMockAstrocyteListener(id string) *mockAstrocyteListener {
	return &mockAstrocyteListener{
		id:           id,
		receivedData: make([]interface{}, 0),
		receivedFrom: make([]string, 0),
	}
}

// ID implements SignalListener interface
func (m *mockAstrocyteListener) ID() string {
	return m.id
}

// OnSignal implements SignalListener interface
func (m *mockAstrocyteListener) OnSignal(signalType types.SignalType, sourceID string, data interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	atomic.AddInt32(&m.receivedCount, 1)
	m.receivedData = append(m.receivedData, data)
	m.receivedFrom = append(m.receivedFrom, sourceID)
}

// GetReceivedCount safely returns the number of received signals
func (m *mockAstrocyteListener) GetReceivedCount() int {
	return int(atomic.LoadInt32(&m.receivedCount))
}

// GetLastReceivedData returns the most recent signal data
func (m *mockAstrocyteListener) GetLastReceivedData() interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.receivedData) == 0 {
		return nil
	}
	return m.receivedData[len(m.receivedData)-1]
}
