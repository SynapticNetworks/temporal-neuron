package neuron

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

/*
=================================================================================
NEURON CORE - STATE AND INTERFACE IMPLEMENTATION
=================================================================================

This file contains ONLY:
- Neuron struct definition and core state
- Component interface implementations
- Configuration methods
- Constructor
- Basic state access helpers
- Enhanced synapse management methods

ALL PROCESSING LOGIC is in processing.go
ALL FIRING LOGIC is in firing.go

This separation ensures clear responsibilities and eliminates duplication.

DEADLOCK FIX:
- Added separate activityMutex to prevent re-entrant lock deadlock
- GetActivityLevel() now uses activityMutex instead of stateMutex
- SendSTDPFeedback() releases stateMutex before calling matrix callbacks

=================================================================================
*/

// ============================================================================
// CORE NEURON STRUCTURE
// ============================================================================

type Neuron struct {
	// === EMBED BASE COMPONENT ===
	*component.BaseComponent

	// === NEURAL PROPERTIES ===
	threshold        float64
	baseThreshold    float64
	decayRate        float64
	refractoryPeriod time.Duration
	fireFactor       float64

	// === BIOLOGICAL PROPERTIES ===
	receptors       []types.LigandType // ChemicalReceiver
	releasedLigands []types.LigandType // ChemicalReleaser
	signalTypes     []types.SignalType // ElectricalReceiver/Transmitter

	// === NEURAL PROCESSING STATE ===
	accumulator  float64
	lastFireTime time.Time
	inputBuffer  chan types.NeuralSignal

	// === HOMEOSTATIC SYSTEM ===
	homeostatic HomeostaticMetrics

	// === MODULAR SYNAPTIC SCALING SYSTEM ===
	synapticScaling *SynapticScalingState

	// === ENHANCED PLASTICITY CONFIGURATION ===
	scalingCheckInterval time.Duration        // 0 = disabled, >0 = enabled with interval
	pruningCheckInterval time.Duration        // 0 = disabled, >0 = enabled with interval
	stdpSystem           *STDPSignalingSystem // ADD: New STDP system

	// === DENDRITIC INTEGRATION ===
	dendrite DendriticIntegrationMode

	// === AXONAL DELIVERY SYSTEM ===
	pendingDeliveries []delayedMessage
	deliveryQueue     chan delayedMessage

	// === CALLBACK-BASED OUTPUTS (NO SYNAPSE DEPENDENCY) ===
	outputCallbacks map[string]types.OutputCallback

	// === INJECTED MATRIX CALLBACKS ===
	matrixCallbacks component.NeuronCallbacks

	// === LIFECYCLE MANAGEMENT ===
	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once

	// === CUSTOM BEHAVIORS (OPTIONAL) ===
	customBehaviors *CustomBehaviors

	// === THREAD SAFETY ===
	stateMutex    sync.Mutex   // Protects neuron state (accumulator, threshold, etc.)
	activityMutex sync.RWMutex // DEADLOCK FIX: Separate mutex for activity calculations
	outputsMutex  sync.RWMutex
}

// ============================================================================
// HOMEOSTATIC METRICS STRUCTURE
// ============================================================================

type HomeostaticMetrics struct {
	firingHistory         []time.Time
	activityWindow        time.Duration
	targetFiringRate      float64
	calciumLevel          float64
	calciumIncrement      float64
	calciumDecayRate      float64
	homeostasisStrength   float64
	minThreshold          float64
	maxThreshold          float64
	lastHomeostaticUpdate time.Time
	homeostaticInterval   time.Duration
}

// ============================================================================
// CONSTRUCTOR
// ============================================================================

func NewNeuron(id string, threshold float64, decayRate float64, refractoryPeriod time.Duration, fireFactor float64, targetFiringRate float64, homeostasisStrength float64) *Neuron {
	ctx, cancel := context.WithCancel(context.Background())

	// Create base component
	baseComponent := component.NewBaseComponent(id, types.TypeNeuron, types.Position3D{})

	// Calculate homeostatic bounds
	minThreshold := threshold * DENDRITE_FACTOR_THRESHOLD_MIN_RATIO // Using new constant
	maxThreshold := threshold * DENDRITE_FACTOR_THRESHOLD_MAX_RATIO // Using new constant

	stdpSystem := NewSTDPSignalingSystem(false, STDP_FEEDBACK_DELAY_DEFAULT, STDP_LEARNING_RATE_DEFAULT)

	neuron := &Neuron{
		BaseComponent:    baseComponent,
		threshold:        threshold,
		baseThreshold:    threshold,
		decayRate:        decayRate,
		refractoryPeriod: refractoryPeriod,
		fireFactor:       fireFactor,

		// Initialize arrays
		receptors:       make([]types.LigandType, 0),
		releasedLigands: make([]types.LigandType, 0),
		signalTypes:     []types.SignalType{types.SignalFired},

		// Initialize processing
		inputBuffer:     make(chan types.NeuralSignal, 100),
		outputCallbacks: make(map[string]types.OutputCallback),

		// Initialize homeostatic system
		homeostatic: HomeostaticMetrics{
			firingHistory:         make([]time.Time, 0, DENDRITE_BUFFER_HISTORY_CAPACITY), // Using new constant
			activityWindow:        DENDRITE_ACTIVITY_TRACKING_WINDOW,                      // Using new constant
			targetFiringRate:      targetFiringRate,
			calciumLevel:          DENDRITE_CALCIUM_BASELINE_INTRACELLULAR, // Using new constant
			calciumIncrement:      DENDRITE_FACTOR_CALCIUM_INCREMENT,       // Using new constant
			calciumDecayRate:      DENDRITE_FACTOR_CALCIUM_DECAY,           // Using new constant
			homeostasisStrength:   homeostasisStrength,
			minThreshold:          minThreshold,
			maxThreshold:          maxThreshold,
			lastHomeostaticUpdate: time.Now(),
			homeostaticInterval:   DENDRITE_TIME_HOMEOSTATIC_TICK, // Using new constant
		},

		// Initialize modular synaptic scaling
		synapticScaling: NewSynapticScalingState(),

		// === INITIALIZE ENHANCED PLASTICITY SETTINGS ===
		scalingCheckInterval: 0,          // 0 means disabled
		pruningCheckInterval: 0,          // 0 means disabled
		stdpSystem:           stdpSystem, // Initialize STDP system

		// Initialize dendritic integration (default to passive)
		dendrite: NewPassiveMembraneMode(),

		// Initialize axonal delivery system
		pendingDeliveries: make([]delayedMessage, 0),
		deliveryQueue:     make(chan delayedMessage, AXON_QUEUE_CAPACITY_DEFAULT),

		// Lifecycle
		ctx:    ctx,
		cancel: cancel,
	}

	neuron.SetState(types.StateInactive) // Start inactive, not active

	return neuron
}

// ============================================================================
// COMPONENT INTERFACE IMPLEMENTATIONS
// ============================================================================

// ChemicalReceiver interface
func (n *Neuron) GetReceptors() []types.LigandType {
	return n.receptors
}

func (n *Neuron) Bind(ligandType types.LigandType, sourceID string, concentration float64) {
	if !n.hasReceptor(ligandType) {
		return
	}

	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// Apply chemical effect
	effect := n.calculateChemicalEffect(ligandType, concentration)
	n.accumulator += effect

	// Update activity
	n.UpdateMetadata("last_chemical_input", time.Now())

	// Check firing (delegated to processing pipeline for consistency)
	if n.accumulator >= n.threshold {
		n.fireUnsafe() // Implemented in firing.go
		n.resetAccumulatorUnsafe()
	}
}

// ChemicalReleaser interface
func (n *Neuron) GetReleasedLigands() []types.LigandType {
	return n.releasedLigands
}

// OnSignal handles electrical signals from gap junctions and network coordination
func (n *Neuron) OnSignal(signalType types.SignalType, sourceID string, data interface{}) {
	switch signalType {
	case types.SignalFired:
		// Gap junction synchronization
		if value, ok := data.(float64); ok {
			n.stateMutex.Lock()
			n.accumulator += value * 0.1 // Small sync effect
			// Check firing after gap junction input
			if n.accumulator >= n.threshold {
				n.fireUnsafe() // Implemented in firing.go
				n.resetAccumulatorUnsafe()
			}
			n.stateMutex.Unlock()
		}
	case types.SignalThresholdChanged:
		// Network-wide threshold adjustment
		if adjustment, ok := data.(float64); ok {
			n.stateMutex.Lock()
			n.threshold += adjustment
			if n.threshold < 0.1 {
				n.threshold = 0.1
			} else if n.threshold > 2.0 {
				n.threshold = 2.0
			}
			n.stateMutex.Unlock()
		}
	}
}

// ElectricalTransmitter interface
func (n *Neuron) GetSignalTypes() []types.SignalType {
	return n.signalTypes
}

// Fix for the data race in neuron.go
// The Receive method needs to protect the lastFireTime read with mutex
// MessageReceiver interface
func (n *Neuron) Receive(msg types.NeuralSignal) {
	// Check refractory period with proper synchronization
	n.stateMutex.Lock()
	inRefractory := !n.lastFireTime.IsZero() && time.Since(n.lastFireTime) < n.refractoryPeriod
	n.stateMutex.Unlock()

	if inRefractory {
		return
	}

	// Update component activity
	n.UpdateMetadata("last_message", time.Now())

	// Queue for processing (actual processing happens in processing.go)
	select {
	case n.inputBuffer <- msg:
		// Successfully queued
	default:
		// Buffer full - message lost (biologically realistic)
	}
}

// DEADLOCK FIX: GetActivityLevel now uses separate activityMutex
func (n *Neuron) GetActivityLevel() float64 {
	n.activityMutex.RLock()
	defer n.activityMutex.RUnlock()
	return n.calculateCurrentFiringRateSafe()
}

// DEADLOCK FIX: Safe firing rate calculation that doesn't hold stateMutex
func (n *Neuron) calculateCurrentFiringRateSafe() float64 {
	// Get a snapshot of firing history without holding stateMutex
	n.stateMutex.Lock()
	firingHistory := make([]time.Time, len(n.homeostatic.firingHistory))
	copy(firingHistory, n.homeostatic.firingHistory)
	targetRate := n.homeostatic.targetFiringRate
	activityWindow := n.homeostatic.activityWindow
	n.stateMutex.Unlock()

	if len(firingHistory) == 0 {
		return 0.0
	}

	// Calculate rate from copied history (safe from mutations)
	now := time.Now()
	windowSize := activityWindow
	if windowSize <= 0 {
		windowSize = 10 * time.Second // Default fallback
	}

	// Count recent firings
	recentCount := 0
	for _, fireTime := range firingHistory {
		if now.Sub(fireTime) <= windowSize {
			recentCount++
		}
	}

	// Convert to Hz
	rate := float64(recentCount) / windowSize.Seconds()

	// Apply reasonable bounds
	if targetRate > 0 && rate > targetRate*2 {
		rate = targetRate * 2
	}

	return rate
}

// ============================================================================
// CONFIGURATION METHODS
// ============================================================================

func (n *Neuron) SetReceptors(receptors []types.LigandType) {
	n.receptors = make([]types.LigandType, len(receptors))
	copy(n.receptors, receptors)
	n.UpdateMetadata("receptors", receptors)
}

func (n *Neuron) SetReleasedLigands(ligands []types.LigandType) {
	n.releasedLigands = make([]types.LigandType, len(ligands))
	copy(n.releasedLigands, ligands)
	n.UpdateMetadata("released_ligands", ligands)
}

func (n *Neuron) GetThreshold() float64 {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	return n.threshold
}

func (n *Neuron) SetThreshold(threshold float64) {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	n.threshold = threshold
}

func (n *Neuron) GetConnectionCount() int {
	n.outputsMutex.RLock()
	defer n.outputsMutex.RUnlock()
	return len(n.outputCallbacks)
}

// ============================================================================
// MODULAR SUBSYSTEM CONFIGURATION
// ============================================================================

// === SYNAPTIC SCALING INTEGRATION ===
func (n *Neuron) EnableSynapticScaling(targetStrength, scalingRate float64, interval time.Duration) error {
	if n.synapticScaling == nil {
		return fmt.Errorf("synaptic scaling system not initialized for neuron %s", n.ID())
	}

	if targetStrength <= 0 {
		return fmt.Errorf("target strength must be positive: %f", targetStrength)
	}

	if scalingRate <= 0 || scalingRate > 1 {
		return fmt.Errorf("scaling rate must be 0 < rate <= 1: %f", scalingRate)
	}

	if interval <= 0 {
		return fmt.Errorf("scaling interval must be positive: %v", interval)
	}

	n.synapticScaling.EnableScaling(targetStrength, scalingRate, interval)

	n.UpdateMetadata("synaptic_scaling_enabled", map[string]interface{}{
		"target_strength": targetStrength,
		"scaling_rate":    scalingRate,
		"interval":        interval,
		"timestamp":       time.Now(),
	})

	return nil
}

func (n *Neuron) DisableSynapticScaling() error {
	if n.synapticScaling == nil {
		return fmt.Errorf("synaptic scaling system not initialized for neuron %s", n.ID())
	}

	n.synapticScaling.DisableScaling()
	n.UpdateMetadata("synaptic_scaling_disabled", time.Now())

	return nil
}

func (n *Neuron) GetSynapticScalingStatus() map[string]interface{} {
	if n.synapticScaling != nil {
		return n.synapticScaling.GetScalingStatus()
	}
	return map[string]interface{}{"enabled": false, "error": "synaptic scaling not initialized"}
}

// === ENHANCED PLASTICITY CONFIGURATION ===
func (n *Neuron) EnableSTDPFeedback(feedbackDelay time.Duration, learningRate float64) {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	n.stdpSystem.Enable(feedbackDelay, learningRate)

	n.UpdateMetadata("stdp_feedback_enabled", map[string]interface{}{
		"feedback_delay": feedbackDelay,
		"learning_rate":  learningRate,
		"timestamp":      time.Now(),
	})
}

func (n *Neuron) DisableSTDPFeedback() {
	// Simply forward to the STDP system
	n.stdpSystem.Disable()

	n.UpdateMetadata("stdp_feedback_disabled", time.Now())
}

func (n *Neuron) EnableAutoHomeostasis(checkInterval time.Duration) {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	n.scalingCheckInterval = checkInterval

	n.UpdateMetadata("auto_homeostasis_enabled", map[string]interface{}{
		"check_interval": checkInterval,
		"timestamp":      time.Now(),
	})
}

func (n *Neuron) DisableAutoHomeostasis() {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	n.scalingCheckInterval = 0 // 0 means disabled
	n.UpdateMetadata("auto_homeostasis_disabled", time.Now())
}

func (n *Neuron) EnableAutoPruning(checkInterval time.Duration) {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	n.pruningCheckInterval = checkInterval

	n.UpdateMetadata("auto_pruning_enabled", map[string]interface{}{
		"check_interval": checkInterval,
		"timestamp":      time.Now(),
	})
}

func (n *Neuron) DisableAutoPruning() {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	n.pruningCheckInterval = 0 // 0 means disabled
	n.UpdateMetadata("auto_pruning_disabled", time.Now())
}

// IsSTDPFeedbackEnabled returns whether STDP feedback is enabled
func (n *Neuron) IsSTDPFeedbackEnabled() bool {
	// Simply forward to the STDP system
	return n.stdpSystem.IsEnabled()
}

func (n *Neuron) IsAutoScalingEnabled() bool {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	return n.scalingCheckInterval > 0
}

func (n *Neuron) IsAutoPruningEnabled() bool {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	return n.pruningCheckInterval > 0
}

// === DENDRITIC INTEGRATION ===
// SetDendriticMode configures the dendritic integration strategy for this neuron
func (n *Neuron) SetDendriticMode(mode DendriticIntegrationMode) error {
	if mode == nil {
		return fmt.Errorf("dendritic mode cannot be nil")
	}

	// Close existing dendrite safely
	if n.dendrite != nil {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Log but don't fail - we want to continue with the new mode
					n.UpdateMetadata("dendrite_close_panic", r)
				}
			}()
			n.dendrite.Close()
		}()
	}

	n.dendrite = mode

	n.UpdateMetadata("dendritic_mode_changed", map[string]interface{}{
		"new_mode":  mode.Name(),
		"timestamp": time.Now(),
	})

	return nil
}

func (n *Neuron) GetDendriticMode() DendriticIntegrationMode {
	return n.dendrite
}

// ============================================================================
// CALLBACK MANAGEMENT
// ============================================================================

func (n *Neuron) SetCallbacks(callbacks component.NeuronCallbacks) {
	n.matrixCallbacks = callbacks
}

func (n *Neuron) AddOutputCallback(synapseID string, callback types.OutputCallback) {
	n.outputsMutex.Lock()
	defer n.outputsMutex.Unlock()
	n.outputCallbacks[synapseID] = callback
}

func (n *Neuron) RemoveOutputCallback(synapseID string) {
	n.outputsMutex.Lock()
	defer n.outputsMutex.Unlock()
	delete(n.outputCallbacks, synapseID)
}

// ConnectToNeuron creates a synapse connection to another neuron via matrix callbacks
func (n *Neuron) ConnectToNeuron(targetNeuronID string, weight float64, synapseType string) error {
	if n.matrixCallbacks == nil {
		return fmt.Errorf("matrix callbacks not available for neuron %s", n.ID())
	}

	if targetNeuronID == "" {
		return fmt.Errorf("target neuron ID cannot be empty")
	}

	if targetNeuronID == n.ID() {
		return fmt.Errorf("neuron %s cannot connect to itself", n.ID())
	}

	if weight < 0 {
		return fmt.Errorf("connection weight cannot be negative: %f", weight)
	}

	config := types.SynapseCreationConfig{
		SourceNeuronID: n.ID(),
		TargetNeuronID: targetNeuronID,
		InitialWeight:  weight,
		SynapseType:    synapseType,
		PlasticityType: "stdp",
		Delay:          AXON_DELAY_DEFAULT_TRANSMISSION,
		Position:       n.Position(), // Inherited from BaseComponent
	}

	synapseID, err := n.matrixCallbacks.CreateSynapse(config)
	if err != nil {
		return fmt.Errorf("failed to create synapse from %s to %s: %w", n.ID(), targetNeuronID, err)
	}

	// Log successful connection for debugging
	n.UpdateMetadata("last_connection_created", map[string]interface{}{
		"target_id":  targetNeuronID,
		"synapse_id": synapseID,
		"weight":     weight,
		"type":       synapseType,
		"timestamp":  time.Now(),
	})

	return nil
}

// ============================================================================
// ENHANCED SYNAPSE MANAGEMENT METHODS
// ============================================================================

// SendSTDPFeedback triggers STDP feedback to update synaptic weights
// Forward this to the new DeliverFeedbackNow method
func (n *Neuron) SendSTDPFeedback() {
	// Get my ID and the matrix callbacks
	myID := n.ID()
	callbacks := n.matrixCallbacks

	if callbacks == nil {
		return
	}

	// Forward to the STDP system
	feedbackCount := n.stdpSystem.DeliverFeedbackNow(myID, callbacks)

	// Update metadata if feedback was delivered
	if feedbackCount > 0 {
		n.UpdateMetadata("stdp_feedback", map[string]interface{}{
			"synapse_count": feedbackCount,
			"timestamp":     time.Now(),
		})
	}
}

// PerformHomeostasisScaling adjusts synaptic weights for stability
// CALLBACKS USED: ListSynapses, SetSynapseWeight
// BIOLOGICAL INTERACTION: Homeostatic plasticity, synaptic scaling
func (n *Neuron) PerformHomeostasisScaling() {
	// Early exit if no callbacks available
	if n.matrixCallbacks == nil {
		return
	}

	// Calculate homeostatic scaling factor based on recent activity
	// Use the activityMutex instead of stateMutex to avoid deadlocks
	n.activityMutex.RLock()
	currentRate := n.calculateCurrentFiringRateSafe() // Use the safe version that doesn't lock stateMutex
	n.activityMutex.RUnlock()

	// Get target rate without holding any locks
	var targetRate float64
	func() {
		n.stateMutex.Lock()
		defer n.stateMutex.Unlock()
		targetRate = n.homeostatic.targetFiringRate
	}()

	if targetRate == 0 {
		return // Homeostasis disabled
	}

	// Calculate scaling factor outside any locks
	scalingFactor := n.calculateScalingFactor(currentRate, targetRate)

	// Get all incoming synapses without holding any neuron locks
	incomingDirection := types.SynapseIncoming
	myID := n.ID()

	incomingSynapses := n.matrixCallbacks.ListSynapses(types.SynapseCriteria{
		Direction: &incomingDirection,
		TargetID:  &myID,
	})

	// Scale all incoming synaptic weights proportionally
	for _, synapseInfo := range incomingSynapses {
		newWeight := synapseInfo.Weight * scalingFactor

		// Clamp to biological bounds
		if newWeight < SYNAPTIC_SCALING_MIN_GAIN {
			newWeight = SYNAPTIC_SCALING_MIN_GAIN
		}
		if newWeight > SYNAPTIC_SCALING_MAX_GAIN {
			newWeight = SYNAPTIC_SCALING_MAX_GAIN
		}

		// Set weight without holding any neuron locks
		err := n.matrixCallbacks.SetSynapseWeight(synapseInfo.ID, newWeight)
		if err != nil {
			n.UpdateMetadata("scaling_error", err.Error())
		}
	}

	// Update metadata when done
	n.UpdateMetadata("homeostasis_scaling_performed", map[string]interface{}{
		"current_rate":    currentRate,
		"target_rate":     targetRate,
		"scaling_factor":  scalingFactor,
		"synapses_scaled": len(incomingSynapses),
		"timestamp":       time.Now(),
	})
}

// PruneDysfunctionalSynapses removes weak or inactive connections
// CALLBACKS USED: ListSynapses, GetSynapse, DeleteSynapse
// BIOLOGICAL INTERACTION: Structural plasticity, synaptic pruning
func (n *Neuron) PruneDysfunctionalSynapses() {
	if n.matrixCallbacks == nil {
		return
	}

	// Get all synapses (both incoming and outgoing)
	bothDirections := types.SynapseBoth
	//myID := n.ID()

	allSynapses := n.matrixCallbacks.ListSynapses(types.SynapseCriteria{
		Direction: &bothDirections,
		//SourceID:  &myID,
		//TargetID:  &myID,
	})

	prunedCount := 0
	for _, synapseInfo := range allSynapses {
		// Get full synapse object to check pruning criteria
		synapse, err := n.matrixCallbacks.GetSynapse(synapseInfo.ID)
		if err != nil {
			continue
		}

		// Check if synapse should be pruned
		if synapse.ShouldPrune() {
			err := n.matrixCallbacks.DeleteSynapse(synapseInfo.ID)
			if err == nil {
				prunedCount++
			}
		}
	}

	if prunedCount > 0 {
		n.UpdateMetadata("synapses_pruned", prunedCount)
	}
}

// GetConnectionMetrics provides network connectivity information
// CALLBACKS USED: ListSynapses
// BIOLOGICAL INTERACTION: Network analysis, connectivity monitoring
func (n *Neuron) GetConnectionMetrics() map[string]interface{} {
	if n.matrixCallbacks == nil {
		return map[string]interface{}{"error": "callbacks not available"}
	}

	myID := n.ID()

	// Count incoming synapses
	incomingDirection := types.SynapseIncoming
	incoming := n.matrixCallbacks.ListSynapses(types.SynapseCriteria{
		Direction: &incomingDirection,
		TargetID:  &myID,
	})

	// Count outgoing synapses
	outgoingDirection := types.SynapseOutgoing
	outgoing := n.matrixCallbacks.ListSynapses(types.SynapseCriteria{
		Direction: &outgoingDirection,
		SourceID:  &myID,
	})

	// Calculate weight statistics
	incomingWeights := make([]float64, len(incoming))
	outgoingWeights := make([]float64, len(outgoing))

	for i, syn := range incoming {
		incomingWeights[i] = syn.Weight
	}
	for i, syn := range outgoing {
		outgoingWeights[i] = syn.Weight
	}

	return map[string]interface{}{
		"incoming_count":      len(incoming),
		"outgoing_count":      len(outgoing),
		"total_count":         len(incoming) + len(outgoing),
		"incoming_weights":    incomingWeights,
		"outgoing_weights":    outgoingWeights,
		"avg_incoming_weight": calculateAverage(incomingWeights),
		"avg_outgoing_weight": calculateAverage(outgoingWeights),
	}
}

// ============================================================================
// ENHANCED HELPER METHODS
// ============================================================================

func (n *Neuron) calculateSTDPTiming(synapseInfo types.SynapseInfo) time.Duration {
	// Simplified STDP timing calculation
	// In real implementation, would track precise spike times
	timeSinceActivity := time.Since(synapseInfo.LastActivity)

	// Causal: synapse fired before neuron (negative deltaT = LTP)
	// Anti-causal: synapse fired after neuron (positive deltaT = LTD)
	if timeSinceActivity < STDP_FEEDBACK_DELAY_DEFAULT {
		return -timeSinceActivity // Causal - strengthen
	}
	return timeSinceActivity // Anti-causal - weaken
}

func (n *Neuron) calculateScalingFactor(currentRate, targetRate float64) float64 {
	// Simple homeostatic scaling
	if currentRate <= 0 {
		return 1.5 // Increase if no activity
	}

	rateRatio := targetRate / currentRate

	// Clamp scaling to reasonable bounds
	if rateRatio > 1.5 {
		return 1.5 // Max 50% increase
	}
	if rateRatio < 0.5 {
		return 0.5 // Max 50% decrease
	}

	return rateRatio
}

func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// ============================================================================
// HELPER METHODS - BASIC STATE ACCESS ONLY
// ============================================================================

// hasReceptor checks if neuron has a specific ligand receptor
func (n *Neuron) hasReceptor(ligandType types.LigandType) bool {
	for _, receptor := range n.receptors {
		if receptor == ligandType {
			return true
		}
	}
	return false
}

// calculateChemicalEffect computes the effect of chemical binding
func (n *Neuron) calculateChemicalEffect(ligandType types.LigandType, concentration float64) float64 {
	switch ligandType {
	case types.LigandGlutamate:
		return concentration * DENDRITE_FACTOR_EFFECT_GLUTAMATE
	case types.LigandGABA:
		return concentration * DENDRITE_FACTOR_EFFECT_GABA
	case types.LigandDopamine:
		return concentration * DENDRITE_FACTOR_EFFECT_DOPAMINE
	case types.LigandSerotonin:
		return concentration * DENDRITE_FACTOR_EFFECT_SEROTONIN
	case types.LigandAcetylcholine:
		return concentration * DENDRITE_FACTOR_EFFECT_ACETYLCHOLINE
	default:
		return 0.0
	}
}

// ============================================================================
// LIFECYCLE MANAGEMENT
// ============================================================================

// Override IsActive from BaseComponent to check actual running state
func (n *Neuron) IsActive() bool {
	// Only active if Start() was called and Stop() hasn't cancelled the context
	select {
	case <-n.ctx.Done():
		return false // Context cancelled = not running
	default:
		// Context is active, but we need to check if Start() was actually called
		// We could track this with a boolean flag or check base component state
		return n.BaseComponent.IsActive()
	}
}

func (n *Neuron) Start() error {
	// Validate neuron state before starting
	if err := n.validateNeuronState(); err != nil {
		return fmt.Errorf("cannot start neuron %s: %w", n.ID(), err)
	}

	n.SetState(types.StateActive)
	go n.Run() // Run() method is in processing.go
	return nil
}

func (n *Neuron) Stop() error {
	var lastErr error

	n.closeOnce.Do(func() {
		n.SetState(types.StateStopped)

		// Signal cancellation first
		if n.cancel != nil {
			n.cancel()
		}

		// Wait a moment for the goroutine to notice cancellation
		time.Sleep(10 * time.Millisecond)

		// Clear callbacks to break circular references
		n.matrixCallbacks = nil

		// Clear output callbacks
		n.outputsMutex.Lock()
		n.outputCallbacks = make(map[string]types.OutputCallback)
		n.outputsMutex.Unlock()

		// Close synaptic scaling with error handling
		if n.synapticScaling != nil {
			func() {
				defer func() {
					if r := recover(); r != nil {
						if lastErr == nil {
							lastErr = fmt.Errorf("panic disabling synaptic scaling: %v", r)
						}
					}
				}()
				n.synapticScaling.DisableScaling() // Now thread-safe
			}()
		}

		// Close dendritic integration with error handling
		if n.dendrite != nil {
			func() {
				defer func() {
					if r := recover(); r != nil {
						if lastErr == nil {
							lastErr = fmt.Errorf("panic closing dendrite: %v", r)
						}
					}
				}()
				n.dendrite.Close()
			}()
		}

		// Close delivery queue with error handling
		func() {
			defer func() {
				if r := recover(); r != nil {
					if lastErr == nil {
						lastErr = fmt.Errorf("panic closing delivery queue: %v", r)
					}
				}
			}()
			close(n.deliveryQueue)
		}()
	})

	return lastErr
}

// validateNeuronState checks if neuron is in a valid state for operation
func (n *Neuron) validateNeuronState() error {
	if n.threshold <= 0 {
		return fmt.Errorf("invalid threshold: %f (must be > 0)", n.threshold)
	}

	if n.decayRate <= 0 || n.decayRate > 1 {
		return fmt.Errorf("invalid decay rate: %f (must be 0 < rate <= 1)", n.decayRate)
	}

	if n.refractoryPeriod < 0 {
		return fmt.Errorf("invalid refractory period: %v (must be >= 0)", n.refractoryPeriod)
	}

	if n.inputBuffer == nil {
		return fmt.Errorf("input buffer not initialized")
	}

	if n.deliveryQueue == nil {
		return fmt.Errorf("delivery queue not initialized")
	}

	if n.synapticScaling == nil {
		return fmt.Errorf("synaptic scaling system not initialized")
	}

	if n.dendrite == nil {
		return fmt.Errorf("dendritic integration system not initialized")
	}

	return nil
}

func (n *Neuron) SetSynapseWeight(synapseID string, weight float64) error {
	if n.matrixCallbacks == nil {
		return fmt.Errorf("SetSynapseWeight callback not available")
	}
	return n.matrixCallbacks.SetSynapseWeight(synapseID, weight)
}

// ScheduleDelayedDelivery implements the SynapseNeuronInterface requirement.
// This method queues messages for delayed delivery without spawning goroutines.
// ScheduleDelayedDelivery implements the SynapseNeuronInterface requirement
func (n *Neuron) ScheduleDelayedDelivery(msg types.NeuralSignal, target component.MessageReceiver, delay time.Duration) {
	// Use your existing axon delivery mechanism
	ScheduleDelayedDelivery(n.deliveryQueue, msg, target, delay)
}
