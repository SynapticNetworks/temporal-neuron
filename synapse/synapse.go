package synapse

import (
	"math"
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/message" // NEW: Import message package
	// NEW: Import component package (though not directly used by BasicSynapse, it's used by SynapseCompatibleNeuron)
)

// =================================================================================
// PLASTICITY AND CONFIGURATION STRUCTURES
// =================================================================================

// STDPConfig represents configuration parameters for Spike-Timing Dependent Plasticity.
// It encapsulates the complete learning rule for a plastic synapse.
//
// STDP is a fundamental learning mechanism in biological neural networks where
// the precise timing of spikes determines whether synapses get stronger or weaker.
// This configuration controls all aspects of this learning process.
type STDPConfig struct {
	Enabled        bool          // Master switch for STDP learning. If false, the synapse is static.
	LearningRate   float64       // The base rate of synaptic weight changes (biologically ~1-5%).
	TimeConstant   time.Duration // The exponential decay time constant of the STDP window (e.g., 20ms).
	WindowSize     time.Duration // The maximum timing difference for STDP effects (e.g., 50ms).
	MinWeight      float64       // The minimum allowed synaptic weight (prevents synapse elimination).
	MaxWeight      float64       // The maximum allowed synaptic weight (models receptor saturation).
	AsymmetryRatio float64       // The ratio of LTD to LTP strength (LTD_strength / LTP_strength).
}

// PruningConfig defines the parameters for determining when a synapse is
// considered ineffective and should be pruned as part of structural plasticity.
//
// In biological neural networks, weak or unused synapses are naturally eliminated
// through a process called synaptic pruning. This "use it or lose it" mechanism
// helps optimize neural circuits by removing ineffective connections while
// preserving important pathways.
type PruningConfig struct {
	Enabled             bool          // If true, this synapse is subject to pruning.
	WeightThreshold     float64       // The weight below which the synapse is a candidate for pruning.
	InactivityThreshold time.Duration // A weak synapse is only pruned if it has been inactive for this duration.
}

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
	// === IDENTIFICATION ===
	id string // Unique identifier for this synapse

	// === NETWORK CONNECTIONS ===
	// These maintain references to the neurons this synapse connects
	preSynapticNeuron  SynapseCompatibleNeuron // A pointer back to the source neuron
	postSynapticNeuron SynapseCompatibleNeuron // A pointer to the target neuron

	// Optional extracellular matrix for spatial delay enhancement
	extracellularMatrix ExtracellularMatrix // nil means no spatial enhancement

	// === SYNAPTIC PROPERTIES ===
	// These define the core transmission characteristics of the synapse
	weight float64       // Current synaptic weight (the "strength" of the connection)
	delay  time.Duration // Axonal + synaptic transmission delay

	// === PLASTICITY CONFIGURATION ===
	// These control how the synapse learns and adapts over time
	stdpConfig    STDPConfig    // Configuration for spike-timing dependent plasticity
	pruningConfig PruningConfig // Configuration for structural plasticity (pruning)

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
//	pre: Pre-synaptic neuron (source of signals)
//	post: Post-synaptic neuron (destination of signals)
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
func NewBasicSynapse(id string, pre SynapseCompatibleNeuron, post SynapseCompatibleNeuron,
	stdpConfig STDPConfig, pruningConfig PruningConfig, initialWeight float64,
	delay time.Duration) *BasicSynapse {
	return NewBasicSynapseWithMatrix(id, pre, post, stdpConfig, pruningConfig, initialWeight, delay, nil)
}

func NewBasicSynapseWithMatrix(id string, pre SynapseCompatibleNeuron, post SynapseCompatibleNeuron,
	stdpConfig STDPConfig, pruningConfig PruningConfig, initialWeight float64,
	delay time.Duration, extracellular ExtracellularMatrix) *BasicSynapse {
	// Validate and clamp delay to non-negative values
	// Negative delays are non-physical and would cause timing issues
	if delay < 0 {
		delay = 0
	}

	// Ensure initial weight is within the configured bounds
	// This prevents starting with weights that could cause immediate instability
	if initialWeight < stdpConfig.MinWeight {
		initialWeight = stdpConfig.MinWeight
	}
	if initialWeight > stdpConfig.MaxWeight {
		initialWeight = stdpConfig.MaxWeight
	}

	// Create and return the new synapse with all fields properly initialized
	return &BasicSynapse{
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

		// Activity tracking (initialize to current time to prevent immediate pruning)
		lastPlasticityEvent: time.Now(),
		lastTransmission:    time.Now(),

		extracellularMatrix: extracellular, // Can be nil for no spatial enhancement

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
// This method models synaptic transmission including both synaptic properties
// and spatial propagation delays through the extracellular matrix.
//
// BIOLOGICAL PROCESS MODELED:
// 1. Pre-synaptic neuron fires (signalValue represents action potential strength)
// 2. Signal is scaled by synaptic weight (synaptic efficacy)
// 3. Base synaptic delay applied (neurotransmitter release and receptor kinetics)
// 4. Spatial delay added via extracellular matrix (axonal propagation distance)
// 5. Message delivered to post-synaptic neuron after total delay
//
// NEW ARCHITECTURE:
// This version eliminates goroutine explosion by using the pre-synaptic neuron's
// dedicated delivery system instead of creating new goroutines per transmission.
// Spatial delays are calculated via the extracellular matrix, separating synaptic
// properties from spatial network topology.
//
// Parameters:
//
//	signalValue: The strength of the incoming signal from the pre-synaptic neuron
func (s *BasicSynapse) Transmit(signalValue float64) {
	// Thread-safe read of current synapse state
	s.mutex.RLock()
	effectiveSignal := signalValue * s.weight // Apply synaptic weight
	baseSynapticDelay := s.delay              // Base synaptic delay
	s.mutex.RUnlock()

	// Update activity tracking for pruning decisions
	s.mutex.Lock()
	s.lastTransmission = time.Now()
	s.mutex.Unlock()

	// Create the message using message.NeuralSignal
	// Populate relevant fields for neural communication and plasticity tracking.
	msg := message.NeuralSignal{
		Value:     effectiveSignal,
		Timestamp: time.Now(),                // When the signal was generated by the synapse
		SourceID:  s.preSynapticNeuron.ID(),  // The original sender neuron
		SynapseID: s.id,                      // This synapse's ID
		TargetID:  s.postSynapticNeuron.ID(), // The intended recipient neuron
		// Other fields like NeurotransmitterType, VesicleReleased, CalciumLevel
		// can be passed into Transmit or derived if the synapse has that context.
		// For now, we'll keep it minimal as per the previous SynapseMessage structure.
	}

	// Calculate total delay: synaptic properties + spatial propagation
	var totalDelay time.Duration
	if s.extracellularMatrix != nil {
		// ENHANCED DELAY: Combine synaptic and spatial delays
		// Models both neurotransmitter kinetics and axonal propagation
		totalDelay = s.extracellularMatrix.EnhanceSynapticDelay(
			s.preSynapticNeuron.ID(),
			s.postSynapticNeuron.ID(),
			s.id,
			baseSynapticDelay, // Pass base delay for enhancement
		)
	} else {
		// BASIC DELAY: Only synaptic properties (no spatial enhancement)
		totalDelay = baseSynapticDelay
	}

	if totalDelay == 0 {
		// IMMEDIATE DELIVERY: If no delay, deliver directly to post-synaptic neuron
		s.postSynapticNeuron.Receive(msg)
	} else {
		// DELAYED DELIVERY: Ask the pre-synaptic neuron to schedule the delivery
		// The pre-synaptic neuron (implementing SynapseCompatibleNeuron) is
		// responsible for managing its axonal queue and ensuring delivery.
		s.preSynapticNeuron.ScheduleDelayedDelivery(msg, s.postSynapticNeuron, totalDelay)
	}
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
func (s *BasicSynapse) ApplyPlasticity(adjustment PlasticityAdjustment) {
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
func calculateSTDPWeightChange(timeDifference time.Duration, config STDPConfig) float64 {
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
func (s *BasicSynapse) GetActivityInfo() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	now := time.Now()
	info := make(map[string]interface{})

	info["lastTransmission"] = s.lastTransmission
	info["lastPlasticity"] = s.lastPlasticityEvent
	info["weight"] = s.weight
	info["timeSinceTransmission"] = now.Sub(s.lastTransmission)
	info["timeSincePlasticity"] = now.Sub(s.lastPlasticityEvent)
	info["id"] = s.id

	return info
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
func (s *BasicSynapse) IsActive(threshold time.Duration) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return time.Since(s.lastTransmission) <= threshold
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
func CreateDefaultSTDPConfig() STDPConfig {
	return STDPConfig{
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
