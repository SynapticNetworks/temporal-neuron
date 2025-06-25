package synapse

import (
	"math"
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component" // NEW: Import message package
	"github.com/SynapticNetworks/temporal-neuron/types"
	// NEW: Import component package (though not directly used by BasicSynapse, it's used by SynapseCompatibleNeuron)
)

// =================================================================================
// DEFAULT SYNAPSE IMPLEMENTATION: BasicSynapse
// =================================================================================

// BasicSynapse is the default, high-performance, non-threaded implementation of
// the SynapticProcessor interface. It encapsulates all the essential properties
// of a biological synapse, including weight, delay, and plasticity configuration.
//
// Key design decisions:
//
//  1. NON-THREADED: Runs in the context of the pre-synaptic neuron's goroutine
//     for efficiency. Does not create additional goroutines.
//
//  2. THREAD-SAFE: All methods are thread-safe to allow plasticity feedback
//     from post-synaptic neurons running in different goroutines.
//
//  4. PRUNING LOGIC: Contains its own logic for determining when it should
//     be eliminated, implementing "use it or lose it" principles.
type BasicSynapse struct {
	*component.BaseComponent
	// === IDENTIFICATION ===
	id string // Unique identifier for this synapse

	// === NETWORK CONNECTIONS ===
	// These maintain references to the neurons this synapse connects
	preSynapticNeuron  component.MessageScheduler // A pointer back to the source neuron
	postSynapticNeuron component.MessageReceiver  // A pointer to the target neuron

	// Optional extracellular matrix for spatial delay enhancement
	extracellularMatrix ExtracellularMatrix // nil means no spatial enhancement

	// === SYNAPTIC PROPERTIES ===
	// These define the core transmission characteristics of the synapse
	weight float64       // Current synaptic weight (the "strength" of the connection)
	delay  time.Duration // Axonal + synaptic transmission delay

	// === PLASTICITY CONFIGURATION ===
	// These control how the synapse learns and adapts over time
	stdpConfig    types.PlasticityConfig // Configuration for spike-timing dependent plasticity
	pruningConfig PruningConfig          // Configuration for structural plasticity (pruning)

	// === ACTIVITY TRACKING ===
	// These track the synapse's recent activity for plasticity and pruning decisions
	lastPlasticityEvent time.Time // Tracks the last time STDP was applied
	lastTransmission    time.Time // Tracks the last time a signal was transmitted

	// === THREAD SAFETY ===
	// A Read-Write mutex ensures thread-safe updates and reads of the synapse's state.
	// This is crucial because a neuron's fire() method (read) and plasticity feedback (write)
	// can be called from different goroutines.
	mutex sync.RWMutex
}

// NewBasicSynapse is the constructor for the default synapse implementation.
// It creates a new synapse with the specified parameters and ensures all
// values are within safe, biological ranges.
//
// Parameters:
//
//	id: Unique identifier for this synapse
//	pre: Pre-synaptic neuron (any neuron-like implementation)
//	post: Post-synaptic neuron (any neuron-like implementation)
//	stdpConfig: Configuration for plasticity learning
//	pruningConfig: Configuration for structural plasticity
//	initialWeight: Starting synaptic weight
//	delay: Transmission delay (axonal + synaptic)
//
// Returns:
//
//	A fully initialized BasicSynapse ready for use
//
// The constructor performs validation and bounds checking to ensure the
// synapse starts in a valid state that won't cause network instabilities.
func NewBasicSynapse(id string, pre component.MessageScheduler, post component.MessageReceiver,
	stdpConfig types.PlasticityConfig, pruningConfig PruningConfig, initialWeight float64,
	delay time.Duration) *BasicSynapse {
	return NewBasicSynapseWithMatrix(id, pre, post, stdpConfig, pruningConfig, initialWeight, delay, nil)
}

func NewBasicSynapseWithMatrix(id string, pre component.MessageScheduler, post component.MessageReceiver,
	stdpConfig types.PlasticityConfig, pruningConfig PruningConfig, initialWeight float64,
	delay time.Duration, extracellular ExtracellularMatrix) *BasicSynapse {

	// Validate and clamp delay to non-negative values
	if delay < 0 {
		delay = 0
	}

	// Ensure initial weight is within the configured bounds
	if initialWeight < stdpConfig.MinWeight {
		initialWeight = stdpConfig.MinWeight
	}
	if initialWeight > stdpConfig.MaxWeight {
		initialWeight = stdpConfig.MaxWeight
	}

	// Calculate synapse position (midpoint between pre and post neurons)
	var synapsePosition types.Position3D
	if pre != nil && post != nil {
		prePos := pre.Position()
		postPos := post.Position()
		synapsePosition = types.Position3D{
			X: (prePos.X + postPos.X) / 2,
			Y: (prePos.Y + postPos.Y) / 2,
			Z: (prePos.Z + postPos.Z) / 2,
		}
	}

	return &BasicSynapse{
		// Initialize the embedded BaseComponent!
		BaseComponent: component.NewBaseComponent(id, types.TypeSynapse, synapsePosition),

		// Identity and connections
		id:                 id,
		preSynapticNeuron:  pre,
		postSynapticNeuron: post,

		// Transmission properties
		weight: initialWeight,
		delay:  delay,

		// Learning and plasticity configurations
		stdpConfig:    stdpConfig,
		pruningConfig: pruningConfig,

		// Activity tracking
		lastPlasticityEvent: time.Now(),
		lastTransmission:    time.Now(),

		extracellularMatrix: extracellular,
	}
}

// =================================================================================
// CORE SYNAPTIC PROCESSOR INTERFACE IMPLEMENTATION
// =================================================================================

// ID returns the unique identifier for the synapse.
// This method is thread-safe as it accesses only immutable data.
func (s *BasicSynapse) ID() string {
	return s.id
}

// Transmit sends a signal through the synapse with proper weight scaling and delay.
// This method completes synchronously and does NOT spawn goroutines. Instead,
// it delegates delay management to the pre-synaptic neuron's delivery system.
//
// BIOLOGICAL PROCESS MODELED:
// 1. Pre-synaptic neuron fires (signalValue represents action potential strength)
// 2. Signal is scaled by synaptic weight (synaptic efficacy)
// 3. Message is created with proper timing and identification metadata
// 4. Delay is calculated (synaptic + spatial components)
// 5. Message is either delivered immediately or scheduled for delayed delivery
//
// NEW ARCHITECTURE BENEFITS:
// - Eliminates goroutine explosion by using neuron's centralized delivery system
// - Maintains synchronous completion of Transmit() call
// - Supports both immediate and delayed delivery through unified interface
// - Allows spatial delays via extracellular matrix without additional complexity
//
// Parameters:
//
//	signalValue: The strength of the incoming signal from the pre-synaptic neuron
func (s *BasicSynapse) Transmit(signalValue float64) {
	// === THREAD-SAFE STATE ACCESS ===
	// Read current synapse state without holding lock during message delivery
	s.mutex.RLock()
	effectiveSignal := signalValue * s.weight // Apply synaptic weight scaling
	baseSynapticDelay := s.delay              // Base synaptic transmission delay
	s.mutex.RUnlock()

	// === ACTIVITY TRACKING FOR PLASTICITY ===
	// Update last transmission time for pruning and plasticity decisions
	s.mutex.Lock()
	s.lastTransmission = time.Now()
	s.mutex.Unlock()

	// === MESSAGE CREATION ===
	// Create neural signal with complete metadata for downstream processing
	msg := types.NeuralSignal{
		Value:     effectiveSignal,           // Signal scaled by synaptic weight
		Timestamp: time.Now(),                // When signal was generated by synapse
		SourceID:  s.preSynapticNeuron.ID(),  // Original sending neuron
		SynapseID: s.id,                      // This synapse's identifier
		TargetID:  s.postSynapticNeuron.ID(), // Intended receiving neuron
		// Additional fields can be populated based on synapse type:
		// NeurotransmitterType, VesicleReleased, CalciumLevel, etc.
	}

	// === DELAY CALCULATION ===
	// Combine synaptic properties with spatial propagation delays
	var totalDelay time.Duration
	if s.extracellularMatrix != nil {
		// ENHANCED DELAY: Synaptic + spatial components
		// Models both neurotransmitter kinetics and axonal propagation distance
		totalDelay = s.extracellularMatrix.SynapticDelay(
			s.preSynapticNeuron.ID(),
			s.postSynapticNeuron.ID(),
			s.id,
			baseSynapticDelay,
		)
	} else {
		// BASIC DELAY: Only synaptic properties
		totalDelay = baseSynapticDelay
	}

	// === MESSAGE DELIVERY STRATEGY ===
	if totalDelay <= 0 {
		// IMMEDIATE DELIVERY: Zero delay, deliver directly to post-synaptic neuron
		// This is the most common case for fast synapses
		s.postSynapticNeuron.Receive(msg)
	} else {
		// Use neuron's centralized delay management
		// No goroutines created here - neuron manages its own delivery queue
		s.preSynapticNeuron.ScheduleDelayedDelivery(msg, s.postSynapticNeuron, totalDelay)
	}

	// === METHOD COMPLETION ===
	// Transmit() returns immediately, allowing the pre-synaptic neuron to continue
	// processing. Delayed messages will be delivered asynchronously by the neuron's
	// own delivery system without blocking this call.
}

// ApplyPlasticity modifies the synapse's weight based on STDP rules.
// This method implements the core learning mechanism that allows synapses
// to strengthen or weaken based on the timing of pre- and post-synaptic activity.
//
// Biological process modeled:
// When a post-synaptic neuron fires, it sends feedback to all synapses that
// recently contributed to its activation. The feedback contains timing information
// that allows each synapse to determine whether it should strengthen (if it
// helped cause the firing) or weaken (if it was not helpful).
//
// Parameters:
//
//	adjustment: Contains the timing difference (Δt) and other plasticity information
//
// The method is thread-safe and respects the STDP configuration parameters
// to ensure biologically plausible learning dynamics.
func (s *BasicSynapse) ApplyPlasticity(adjustment types.PlasticityAdjustment) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Skip plasticity if STDP is disabled for this synapse
	if !s.stdpConfig.Enabled {
		return
	}

	// Calculate the weight change based on spike timing
	change := calculateSTDPWeightChange(adjustment.DeltaT, s.stdpConfig)
	newWeight := s.weight + change

	// Enforce the weight boundaries defined in the configuration.
	// This prevents runaway strengthening or complete elimination of synapses.
	if newWeight < s.stdpConfig.MinWeight {
		newWeight = s.stdpConfig.MinWeight
	} else if newWeight > s.stdpConfig.MaxWeight {
		newWeight = s.stdpConfig.MaxWeight
	}

	// Apply the weight change and update tracking
	s.weight = newWeight
	s.lastPlasticityEvent = time.Now()
}

// ShouldPrune determines if a synapse is a candidate for removal.
// This method implements the "use it or lose it" principle found in biological
// neural networks, where weak or inactive synapses are gradually eliminated.
//
// Biological basis:
// In real brains, synapses that are consistently weak or rarely active are
// gradually eliminated through various molecular mechanisms. This pruning
// process helps optimize neural circuits by removing ineffective connections
// while preserving important pathways.
//
// Returns:
//
//	true if the synapse should be pruned, false if it should be retained
//
// The decision is based on multiple criteria:
// 1. Weight threshold: Is the synapse too weak to be effective?
// 2. Activity threshold: Has the synapse been inactive for too long?
// 3. Plasticity activity: Has the synapse shown any learning recently?
func (s *BasicSynapse) ShouldPrune() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// If pruning is disabled, never prune
	if !s.pruningConfig.Enabled {
		return false
	}

	// Check if the synapse is too weak to be effective
	isWeightWeak := s.weight < s.pruningConfig.WeightThreshold

	// Check if the synapse has been inactive for too long
	timeSinceActivity := time.Since(s.lastPlasticityEvent)
	isInactive := timeSinceActivity > s.pruningConfig.InactivityThreshold

	// Prune only if BOTH conditions are met:
	// - The synapse is weak (low weight)
	// - AND it has been inactive (no recent plasticity)
	//
	// This two-condition requirement prevents premature pruning of synapses
	// that might be temporarily weak but still active, or temporarily inactive
	// but still strong.
	return isWeightWeak && isInactive
}

// GetWeight provides a thread-safe way to read the current synaptic weight.
// This method is essential for monitoring learning progress and network analysis.
//
// Returns:
//
//	The current synaptic weight
//
// This method is frequently called for:
// - Network visualization and monitoring
// - Learning progress analysis
// - Debugging connectivity issues
// - Research data collection
func (s *BasicSynapse) GetWeight() float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.weight
}

// SetWeight provides a thread-safe way to manually set the synaptic weight.
// This method is useful for experimental manipulation and network initialization.
//
// Parameters:
//
//	weight: The new weight to set (will be clamped to configured bounds)
//
// Use cases:
// - Experimental manipulation of network connectivity
// - Initialization of specific weight patterns
// - Testing network behavior with controlled weights
// - Simulating various learning scenarios
//
// The method enforces weight bounds to prevent values that could destabilize
// the network or violate biological constraints.
func (s *BasicSynapse) SetWeight(weight float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Enforce weight boundaries to maintain network stability
	if weight < s.stdpConfig.MinWeight {
		weight = s.stdpConfig.MinWeight
	} else if weight > s.stdpConfig.MaxWeight {
		weight = s.stdpConfig.MaxWeight
	}

	// Update the weight and record this as a plasticity event
	s.weight = weight
	s.lastPlasticityEvent = time.Now() // Reset activity tracking
}

// =================================================================================
// STDP CALCULATION LOGIC
// This section implements the mathematical models for spike-timing dependent plasticity
// =================================================================================

// calculateSTDPWeightChange computes the STDP weight change based on spike timing.
// This function implements the classic asymmetric STDP learning window that is
// fundamental to biological neural learning.
//
// Biological basis:
// STDP is based on the principle that synapses strengthen when they successfully
// contribute to post-synaptic firing (causal relationship) and weaken when they
// fire after the post-synaptic neuron is already committed to firing (non-causal).
//
// The learning window has an asymmetric shape:
// - Negative Δt (pre before post): LTP (Long Term Potentiation) - strengthening
// - Positive Δt (pre after post): LTD (Long Term Depression) - weakening
//
// Parameters:
//
//	timeDifference: Δt = t_pre - t_post (the timing relationship)
//	config: STDP configuration parameters
//
// Returns:
//
//	The weight change to apply (positive = strengthen, negative = weaken)
func calculateSTDPWeightChange(timeDifference time.Duration, config types.PlasticityConfig) float64 {
	// Convert time difference to milliseconds for calculation
	deltaT := timeDifference.Seconds() * 1000.0 // Convert to milliseconds
	windowMs := config.WindowSize.Seconds() * 1000.0

	// Check if the timing difference is within the STDP window
	// Spikes separated by more than the window size have no plasticity effect
	if math.Abs(deltaT) >= windowMs {
		return 0.0 // No plasticity outside the timing window
	}

	// Get the time constant in milliseconds
	tauMs := config.TimeConstant.Seconds() * 1000.0
	if tauMs == 0 {
		return 0.0 // Avoid division by zero
	}

	// Calculate the STDP weight change based on timing
	if deltaT < 0 {
		// CAUSAL (LTP): Pre-synaptic spike before post-synaptic (t_pre - t_post < 0)
		// This represents a causal relationship where the synapse helped cause firing
		// Result: Positive weight change (strengthening)
		return config.LearningRate * math.Exp(deltaT/tauMs)

	} else if deltaT > 0 {
		// ANTI-CAUSAL (LTD): Pre-synaptic spike after post-synaptic (t_pre - t_post > 0)
		// This represents a non-causal relationship where the synapse fired too late
		// Result: Negative weight change (weakening)
		return -config.LearningRate * config.AsymmetryRatio * math.Exp(-deltaT/tauMs)
	}

	// Simultaneous firing (deltaT == 0) - treat as weak LTD
	// In practice, perfectly simultaneous spikes are rare but can occur
	return -config.LearningRate * config.AsymmetryRatio * 0.1
}

// =================================================================================
// UTILITY FUNCTIONS AND HELPERS
// =================================================================================

// GetActivityInfo returns information about the synapse's recent activity.
// This method provides read-only access to activity metrics for monitoring
// and analysis purposes.
//
// Returns:
//
//	A map containing activity information including:
//	- "lastTransmission": Time of last signal transmission
//	- "lastPlasticity": Time of last plasticity event
//	- "weight": Current synaptic weight
//	- "timeSinceTransmission": Duration since last transmission
//	- "timeSincePlasticity": Duration since last plasticity
//
// GetActivityInfo returns information about the synapse's recent activity.
// This method provides read-only access to activity metrics for monitoring
// and analysis purposes using a proper struct instead of a map.
func (s *BasicSynapse) GetActivityInfo() types.ActivityInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	now := time.Now()

	return types.ActivityInfo{
		ComponentID:           s.id,
		LastTransmission:      s.lastTransmission,
		LastPlasticity:        s.lastPlasticityEvent,
		Weight:                s.weight,
		ActivityLevel:         0.0, // TODO: Calculate actual activity level
		TimeSinceTransmission: now.Sub(s.lastTransmission),
		TimeSincePlasticity:   now.Sub(s.lastPlasticityEvent),
		ConnectionCount:       0, // Not applicable for synapses
	}
}

// IsActive returns true if the synapse has been active recently.
// This is a convenience method for quickly checking synapse status.
//
// Parameters:
//
//	threshold: Time duration - synapse is considered active if it transmitted
//	          a signal within this time period
//
// Returns:
//
//	true if the synapse transmitted a signal within the threshold period
//
// IsActiveInWindow checks if the synapse has been active within a specific time threshold.
// This method provides more detailed control over activity checking.
func (s *BasicSynapse) IsActiveInWindow(threshold time.Duration) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return time.Since(s.lastTransmission) <= threshold
}

// IsActive returns true if the synapse is considered generally active.
// This implements the extracellular.SynapseInterface requirement.
// It uses a default activity threshold, or a reasonable heuristic.
func (s *BasicSynapse) IsActive() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	// Use a sensible default threshold, perhaps from constants.go or a configurable field.
	// For example, SYNAPSE_ACTIVITY_THRESHOLD (defined in synapse/constants.go)
	return time.Since(s.lastTransmission) <= SYNAPSE_ACTIVITY_THRESHOLD //
}

// GetDelay returns the current transmission delay for this synapse.
// This method provides read-only access to the delay parameter.
//
// Returns:
//
//	The current transmission delay
func (s *BasicSynapse) GetDelay() time.Duration {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.delay
}

// SetDelay allows modification of the transmission delay.
// This method is useful for experimental manipulation of network timing.
//
// Parameters:
//
//	delay: New transmission delay (will be clamped to non-negative values)
//
// Use cases:
// - Experimental studies of timing effects
// - Simulation of different axon lengths
// - Network optimization studies
func (s *BasicSynapse) SetDelay(delay time.Duration) {
	if delay < 0 {
		delay = 0 // Clamp to non-negative values
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.delay = delay
}

// =================================================================================
// CONFIGURATION HELPERS
// =================================================================================

// CreateDefaultSTDPConfig returns a standard STDP configuration suitable for most applications
// This provides sensible defaults based on biological cortical synapse parameters
func CreateDefaultSTDPConfig() types.PlasticityConfig {
	return types.PlasticityConfig{
		Enabled:        true,
		LearningRate:   STDP_DEFAULT_LEARNING_RATE,
		TimeConstant:   STDP_DEFAULT_TIME_CONSTANT,
		WindowSize:     STDP_DEFAULT_WINDOW_SIZE,
		MinWeight:      STDP_DEFAULT_MIN_WEIGHT,
		MaxWeight:      STDP_DEFAULT_MAX_WEIGHT,
		AsymmetryRatio: STDP_DEFAULT_ASYMMETRY_RATIO,
	}
}

// CreateDefaultPruningConfig returns a standard pruning configuration
// This provides conservative pruning parameters suitable for stable learning
func CreateDefaultPruningConfig() PruningConfig {
	return PruningConfig{
		Enabled:             true,
		WeightThreshold:     PRUNING_DEFAULT_WEIGHT_THRESHOLD,
		InactivityThreshold: PRUNING_DEFAULT_INACTIVITY_THRESHOLD,
	}
}

// CreateConservativePruningConfig returns a more conservative pruning configuration
// Use this when you want to minimize the risk of losing important connections
func CreateConservativePruningConfig() PruningConfig {
	return PruningConfig{
		Enabled:             true,
		WeightThreshold:     PRUNING_CONSERVATIVE_WEIGHT_THRESHOLD,
		InactivityThreshold: PRUNING_CONSERVATIVE_INACTIVITY_THRESHOLD,
	}
}

// GetPlasticityConfig returns the current plasticity configuration
func (s *BasicSynapse) GetPlasticityConfig() types.PlasticityConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Convert internal STDPConfig to types.PlasticityConfig
	return types.PlasticityConfig{
		Enabled:        s.stdpConfig.Enabled,
		LearningRate:   s.stdpConfig.LearningRate,
		TimeConstant:   s.stdpConfig.TimeConstant,
		WindowSize:     s.stdpConfig.WindowSize,
		MinWeight:      s.stdpConfig.MinWeight,
		MaxWeight:      s.stdpConfig.MaxWeight,
		AsymmetryRatio: s.stdpConfig.AsymmetryRatio,
	}
}

// GetPresynapticID returns the ID of the pre-synaptic neuron
func (s *BasicSynapse) GetPresynapticID() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.preSynapticNeuron.ID()
}

// GetPostsynapticID returns the ID of the post-synaptic neuron
func (s *BasicSynapse) GetPostsynapticID() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.postSynapticNeuron.ID()
}

// UpdateWeight applies plasticity events to modify synaptic strength
func (s *BasicSynapse) UpdateWeight(event types.PlasticityEvent) {
	adjustment := types.PlasticityAdjustment{
		DeltaT:       event.DeltaT,
		PostSynaptic: true,
		PreSynaptic:  true,
		Timestamp:    event.Timestamp,
	}
	s.ApplyPlasticity(adjustment)
}

// GetLastActivity returns the timestamp of the most recent transmission
func (s *BasicSynapse) GetLastActivity() time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.GetActivityInfo().LastTransmission
}
