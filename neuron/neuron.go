package neuron

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/message"
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
	receptors       []message.LigandType // ChemicalReceiver
	releasedLigands []message.LigandType // ChemicalReleaser
	signalTypes     []message.SignalType // ElectricalReceiver/Transmitter

	// === NEURAL PROCESSING STATE ===
	accumulator  float64
	lastFireTime time.Time
	inputBuffer  chan message.NeuralSignal

	// === HOMEOSTATIC SYSTEM ===
	homeostatic HomeostaticMetrics

	// === MODULAR SYNAPTIC SCALING SYSTEM ===
	synapticScaling *SynapticScalingState

	// === ENHANCED PLASTICITY CONFIGURATION ===
	stdpFeedbackDelay    time.Duration // 0 = disabled, >0 = enabled with delay
	stdpLearningRate     float64       // Learning rate for STDP adjustments
	scalingCheckInterval time.Duration // 0 = disabled, >0 = enabled with interval
	pruningCheckInterval time.Duration // 0 = disabled, >0 = enabled with interval

	// === DENDRITIC INTEGRATION ===
	dendrite DendriticIntegrationMode

	// === AXONAL DELIVERY SYSTEM ===
	pendingDeliveries []delayedMessage
	deliveryQueue     chan delayedMessage

	// === CALLBACK-BASED OUTPUTS (NO SYNAPSE DEPENDENCY) ===
	outputCallbacks map[string]OutputCallback

	// === INJECTED MATRIX CALLBACKS ===
	matrixCallbacks *NeuronCallbacks

	// === LIFECYCLE MANAGEMENT ===
	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once

	// === THREAD SAFETY ===
	stateMutex   sync.Mutex
	outputsMutex sync.RWMutex
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
	baseComponent := component.NewBaseComponent(id, component.TypeNeuron, component.Position3D{})

	// Calculate homeostatic bounds
	minThreshold := threshold * DENDRITE_FACTOR_THRESHOLD_MIN_RATIO // Using new constant
	maxThreshold := threshold * DENDRITE_FACTOR_THRESHOLD_MAX_RATIO // Using new constant

	neuron := &Neuron{
		BaseComponent:    baseComponent,
		threshold:        threshold,
		baseThreshold:    threshold,
		decayRate:        decayRate,
		refractoryPeriod: refractoryPeriod,
		fireFactor:       fireFactor,

		// Initialize arrays
		receptors:       make([]message.LigandType, 0),
		releasedLigands: make([]message.LigandType, 0),
		signalTypes:     []message.SignalType{message.SignalFired},

		// Initialize processing
		inputBuffer:     make(chan message.NeuralSignal, 100),
		outputCallbacks: make(map[string]OutputCallback),

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
		stdpFeedbackDelay:    0, // 0 means disabled
		stdpLearningRate:     STDP_LEARNING_RATE_DEFAULT,
		scalingCheckInterval: 0, // 0 means disabled
		pruningCheckInterval: 0, // 0 means disabled

		// Initialize dendritic integration (default to passive)
		dendrite: NewPassiveMembraneMode(),

		// Initialize axonal delivery system
		pendingDeliveries: make([]delayedMessage, 0),
		deliveryQueue:     make(chan delayedMessage, AXON_QUEUE_CAPACITY_DEFAULT),

		// Lifecycle
		ctx:    ctx,
		cancel: cancel,
	}

	return neuron
}

// ============================================================================
// COMPONENT INTERFACE IMPLEMENTATIONS
// ============================================================================

// ChemicalReceiver interface
func (n *Neuron) GetReceptors() []message.LigandType {
	return n.receptors
}

func (n *Neuron) Bind(ligandType message.LigandType, sourceID string, concentration float64) {
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
func (n *Neuron) GetReleasedLigands() []message.LigandType {
	return n.releasedLigands
}

// OnSignal handles electrical signals from gap junctions and network coordination
func (n *Neuron) OnSignal(signalType message.SignalType, sourceID string, data interface{}) {
	switch signalType {
	case message.SignalFired:
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
	case message.SignalThresholdChanged:
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
func (n *Neuron) GetSignalTypes() []message.SignalType {
	return n.signalTypes
}

// Fix for the data race in neuron.go
// The Receive method needs to protect the lastFireTime read with mutex
// MessageReceiver interface
func (n *Neuron) Receive(msg message.NeuralSignal) {
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

// Override GetActivityLevel from BaseComponent
func (n *Neuron) GetActivityLevel() float64 {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	return n.calculateCurrentFiringRateUnsafe()
}

// ============================================================================
// CONFIGURATION METHODS
// ============================================================================

func (n *Neuron) SetReceptors(receptors []message.LigandType) {
	n.receptors = make([]message.LigandType, len(receptors))
	copy(n.receptors, receptors)
	n.UpdateMetadata("receptors", receptors)
}

func (n *Neuron) SetReleasedLigands(ligands []message.LigandType) {
	n.releasedLigands = make([]message.LigandType, len(ligands))
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

	n.stdpFeedbackDelay = feedbackDelay
	n.stdpLearningRate = learningRate

	n.UpdateMetadata("stdp_feedback_enabled", map[string]interface{}{
		"feedback_delay": feedbackDelay,
		"learning_rate":  learningRate,
		"timestamp":      time.Now(),
	})
}

func (n *Neuron) DisableSTDPFeedback() {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	n.stdpFeedbackDelay = 0 // 0 means disabled
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

// Getter methods for checking current settings
func (n *Neuron) IsSTDPFeedbackEnabled() bool {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	return n.stdpFeedbackDelay > 0
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

func (n *Neuron) SetCallbacks(callbacks NeuronCallbacks) {
	n.matrixCallbacks = &callbacks
}

func (n *Neuron) AddOutputCallback(synapseID string, callback OutputCallback) {
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

	if n.matrixCallbacks.CreateSynapse == nil {
		return fmt.Errorf("CreateSynapse callback not available for neuron %s", n.ID())
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

	config := SynapseCreationConfig{
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

// SendSTDPFeedback implements spike-timing dependent plasticity
// CALLBACKS USED: ListSynapses, ApplyPlasticity
// BIOLOGICAL INTERACTION: Hebbian learning, synaptic strengthening/weakening
func (n *Neuron) SendSTDPFeedback() {
	// Check if STDP feedback is enabled by checking if delay > 0
	n.stateMutex.Lock()
	feedbackDelay := n.stdpFeedbackDelay
	learningRate := n.stdpLearningRate
	n.stateMutex.Unlock()

	if feedbackDelay <= 0 {
		return // STDP feedback disabled
	}

	if n.matrixCallbacks == nil || n.matrixCallbacks.ListSynapses == nil || n.matrixCallbacks.ApplyPlasticity == nil {
		return // Callbacks not available
	}

	// Get all incoming synapses that recently contributed to firing
	incomingDirection := SynapseIncoming
	myID := n.ID()
	recentActivity := time.Now().Add(-feedbackDelay * 10) // STDP window

	incomingSynapses := n.matrixCallbacks.ListSynapses(SynapseCriteria{
		Direction:     &incomingDirection,
		TargetID:      &myID,
		ActivitySince: &recentActivity,
	})

	// Apply STDP to each synapse based on timing
	for _, synapseInfo := range incomingSynapses {
		// Calculate timing difference (simplified - would need actual spike times)
		deltaT := n.calculateSTDPTiming(synapseInfo)

		adjustment := PlasticityAdjustment{
			DeltaT:       deltaT,
			LearningRate: learningRate, // Use configured learning rate
		}

		// Send plasticity feedback to synapse
		err := n.matrixCallbacks.ApplyPlasticity(synapseInfo.ID, adjustment)
		if err != nil {
			n.UpdateMetadata("stdp_error", err.Error())
		}
	}
}

// PerformHomeostasisScaling adjusts synaptic weights for stability
// CALLBACKS USED: ListSynapses, SetSynapseWeight
// BIOLOGICAL INTERACTION: Homeostatic plasticity, synaptic scaling
func (n *Neuron) PerformHomeostasisScaling() {
	if n.matrixCallbacks == nil || n.matrixCallbacks.ListSynapses == nil || n.matrixCallbacks.SetSynapseWeight == nil {
		return
	}

	// Calculate homeostatic scaling factor based on recent activity
	n.stateMutex.Lock()
	currentRate := n.calculateCurrentFiringRateUnsafe()
	targetRate := n.homeostatic.targetFiringRate
	n.stateMutex.Unlock()

	if targetRate == 0 {
		return // Homeostasis disabled
	}

	scalingFactor := n.calculateScalingFactor(currentRate, targetRate)

	// Get all incoming synapses
	incomingDirection := SynapseIncoming
	myID := n.ID()

	incomingSynapses := n.matrixCallbacks.ListSynapses(SynapseCriteria{
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

		err := n.matrixCallbacks.SetSynapseWeight(synapseInfo.ID, newWeight)
		if err != nil {
			n.UpdateMetadata("scaling_error", err.Error())
		}
	}
}

// PruneDysfunctionalSynapses removes weak or inactive connections
// CALLBACKS USED: ListSynapses, GetSynapse, DeleteSynapse
// BIOLOGICAL INTERACTION: Structural plasticity, synaptic pruning
func (n *Neuron) PruneDysfunctionalSynapses() {
	if n.matrixCallbacks == nil {
		return
	}

	// Get all synapses (both incoming and outgoing)
	bothDirections := SynapseBoth
	myID := n.ID()

	allSynapses := n.matrixCallbacks.ListSynapses(SynapseCriteria{
		Direction: &bothDirections,
		SourceID:  &myID,
		TargetID:  &myID,
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
	if n.matrixCallbacks == nil || n.matrixCallbacks.ListSynapses == nil {
		return map[string]interface{}{"error": "callbacks not available"}
	}

	myID := n.ID()

	// Count incoming synapses
	incomingDirection := SynapseIncoming
	incoming := n.matrixCallbacks.ListSynapses(SynapseCriteria{
		Direction: &incomingDirection,
		TargetID:  &myID,
	})

	// Count outgoing synapses
	outgoingDirection := SynapseOutgoing
	outgoing := n.matrixCallbacks.ListSynapses(SynapseCriteria{
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

func (n *Neuron) calculateSTDPTiming(synapseInfo SynapseInfo) time.Duration {
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
func (n *Neuron) hasReceptor(ligandType message.LigandType) bool {
	for _, receptor := range n.receptors {
		if receptor == ligandType {
			return true
		}
	}
	return false
}

// calculateChemicalEffect computes the effect of chemical binding
func (n *Neuron) calculateChemicalEffect(ligandType message.LigandType, concentration float64) float64 {
	switch ligandType {
	case message.LigandGlutamate:
		return concentration * DENDRITE_FACTOR_EFFECT_GLUTAMATE
	case message.LigandGABA:
		return concentration * DENDRITE_FACTOR_EFFECT_GABA
	case message.LigandDopamine:
		return concentration * DENDRITE_FACTOR_EFFECT_DOPAMINE
	case message.LigandSerotonin:
		return concentration * DENDRITE_FACTOR_EFFECT_SEROTONIN
	case message.LigandAcetylcholine:
		return concentration * DENDRITE_FACTOR_EFFECT_ACETYLCHOLINE
	default:
		return 0.0
	}
}

// ============================================================================
// LIFECYCLE MANAGEMENT
// ============================================================================

func (n *Neuron) Start() error {
	// Validate neuron state before starting
	if err := n.validateNeuronState(); err != nil {
		return fmt.Errorf("cannot start neuron %s: %w", n.ID(), err)
	}

	n.SetState(component.StateActive)
	go n.Run() // Run() method is in processing.go
	return nil
}

func (n *Neuron) Stop() error {
	var lastErr error

	n.closeOnce.Do(func() {
		n.SetState(component.StateStopped)

		// Signal cancellation first
		n.cancel()

		// Wait a moment for the goroutine to notice cancellation
		time.Sleep(10 * time.Millisecond)

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
	if n.matrixCallbacks == nil || n.matrixCallbacks.SetSynapseWeight == nil {
		return fmt.Errorf("SetSynapseWeight callback not available")
	}
	return n.matrixCallbacks.SetSynapseWeight(synapseID, weight)
}
