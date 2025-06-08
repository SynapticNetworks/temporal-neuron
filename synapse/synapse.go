/*
=================================================================================
SYNAPTIC PROCESSOR - A MODULAR, PLUGGABLE SYNAPSE ARCHITECTURE
=================================================================================

OVERVIEW:
This file defines the architecture for synaptic connections, separating the concept
of a synapse from the neuron. This modular approach is central to building a
biologically realistic and extensible simulation. It allows different types of
synapses (e.g., fast, slow, plastic, static) to be developed and "plugged into"
a neuron without changing the neuron's core logic.

The design is centered around the `SynapticProcessor` interface, which defines a
standard contract for what any synapse can do. The default, high-performance
implementation provided here is the `BasicSynapse`, a non-threaded component
managed by the pre-synaptic neuron's goroutine.

ARCHITECTURAL PRINCIPLES:
1. INTERFACE-BASED DESIGN: The Neuron interacts with the `SynapticProcessor`
   interface, not a concrete implementation. This decouples the neuron from the
   synapse, enabling true modularity and future expansion.

2. ENCAPSULATION: All logic related to a specific connection—its weight, delay,
   and plasticity rules (including pruning)—is encapsulated within the synapse
   component. The neuron's role is simplified to managing its portfolio of
   synaptic connections.

3. PERFORMANCE: The default `BasicSynapse` is designed for high performance
   in large-scale networks. It avoids the massive overhead of creating one
   goroutine per synapse by using a non-blocking timer (`time.AfterFunc`) for
   handling signal transmission delays efficiently.

4. STRUCTURAL PLASTICITY: The synapse itself contains the logic to determine
   if it has become ineffective. It signals this to the parent neuron via the
   `ShouldPrune` method, allowing the neuron to implement "use-it-or-lose-it"
   rules without needing to know the internal details of the synapse.

STANDALONE DESIGN:
This synapse system is designed to work alongside existing neuron implementations
without breaking existing functionality. It defines its own Message type and
communication protocols, allowing for gradual migration from existing systems.
*/

package synapse

import (
	"math"
	"sync"
	"time"
)

// =================================================================================
// SYNAPSE MESSAGE SYSTEM
// Separate from existing neuron Message to avoid conflicts during transition
// =================================================================================

// SynapseMessage represents a signal transmitted between neurons through synapses
// This is separate from any existing Message types to ensure no conflicts
// Models the discrete action potential spikes in biological neural communication
//
// Enhanced with timing information for Spike-Timing Dependent Plasticity (STDP):
// STDP requires precise timing information to determine whether synaptic connections
// should be strengthened or weakened based on the temporal relationship between
// pre-synaptic and post-synaptic spikes
type SynapseMessage struct {
	Value float64 // Signal strength/intensity (can be positive or negative)
	// Positive values = excitatory signals (increase firing probability)
	// Negative values = inhibitory signals (decrease firing probability)

	Timestamp time.Time // Precise timestamp when the spike occurred at the source
	// Critical for STDP: the relative timing between pre and post-synaptic spikes
	// determines whether synaptic strength increases (LTP) or decreases (LTD)
	// Biological timescale: STDP effects occur within ±20-50ms windows

	SourceID string // Identifier of the neuron that generated this spike
	// Enables post-synaptic neurons to track which specific synapses contributed
	// to their firing and apply appropriate STDP weight modifications
	// Essential for learning in networks with multiple inputs per neuron

	SynapseID string // Identifier of the specific synapse that transmitted this signal
	// Allows precise tracking of which synapse delivered the signal
	// Useful for debugging, analysis, and synapse-specific feedback
}

// =================================================================================
// PLASTICITY AND CONFIGURATION STRUCTURES
// =================================================================================

// PlasticityAdjustment is a feedback message sent from a post-synaptic neuron
// back to a pre-synaptic synapse to trigger a plasticity event (e.g., STDP).
// This models the retrograde signaling mechanisms found in biological systems.
//
// In biology, when a post-synaptic neuron fires, it can send feedback signals
// back to the synapses that contributed to its firing. This feedback contains
// information about the timing relationship between pre- and post-synaptic
// activity, which is used to strengthen or weaken the synaptic connection.
type PlasticityAdjustment struct {
	// DeltaT is the time difference between the pre-synaptic and post-synaptic spikes.
	// Its sign and magnitude determine the direction and strength of the synaptic
	// weight change according to the STDP rule.
	//
	// Convention: Δt = t_pre - t_post
	//   - Δt < 0 (causal): pre-synaptic spike occurred BEFORE post-synaptic spike -> LTP
	//   - Δt > 0 (anti-causal): pre-synaptic spike occurred AFTER post-synaptic spike -> LTD
	//
	// Biological basis: This timing relationship determines whether synapses are
	// strengthened (if they helped cause the post-synaptic firing) or weakened
	// (if they fired after the neuron was already committed to firing).
	DeltaT time.Duration
}

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
// SYNAPTIC PROCESSOR INTERFACE
// This is the core contract for the pluggable synapse architecture.
// =================================================================================

// SynapticProcessor defines the universal contract for any component that acts
// as a synapse. It is the key to the pluggable architecture, ensuring that the
// Neuron can work with any synapse type that fulfills these methods.
//
// This interface abstracts away the implementation details of different synapse
// types, allowing for:
// - Static synapses (fixed weights)
// - Plastic synapses (learning via STDP)
// - Inhibitory vs excitatory synapses
// - Fast vs slow synapses
// - Complex multi-compartment synapses
//
// All synapse types can be used interchangeably by neurons as long as they
// implement this interface.
type SynapticProcessor interface {
	// ID returns the unique identifier for the synapse.
	// This allows neurons to manage and reference specific synapses by name.
	ID() string

	// Transmit processes an outgoing signal from the pre-synaptic neuron.
	// It is responsible for applying the synapse's weight and handling the
	// axonal transmission delay before delivering the signal to the
	// post-synaptic neuron.
	//
	// Parameters:
	//   signalValue: The strength of the signal from the pre-synaptic neuron
	//
	// The synapse applies its weight and delay, then sends the modified signal
	// to the post-synaptic neuron after the appropriate delay period.
	Transmit(signalValue float64)

	// ApplyPlasticity updates the synapse's internal state (e.g., weight)
	// based on a feedback signal from the post-synaptic neuron, typically
	// containing spike timing information (Δt) for STDP.
	//
	// Parameters:
	//   adjustment: Contains timing and other information needed for plasticity
	//
	// This method implements the learning aspect of synapses, allowing them
	// to strengthen or weaken based on their effectiveness in causing
	// post-synaptic firing.
	ApplyPlasticity(adjustment PlasticityAdjustment)

	// ShouldPrune evaluates the synapse's internal state to determine if it
	// has become ineffective and should be removed by the parent neuron.
	// This method encapsulates the logic for structural plasticity.
	//
	// Returns:
	//   true if the synapse should be eliminated, false otherwise
	//
	// The decision is based on factors like:
	// - Current synaptic weight (too weak?)
	// - Recent activity levels (unused?)
	// - Time since last plasticity event (stagnant?)
	ShouldPrune() bool

	// GetWeight returns the current effective weight of the synapse. This is
	// crucial for monitoring, debugging, and validating learning.
	//
	// Returns:
	//   The current synaptic weight/strength
	GetWeight() float64

	// SetWeight allows for direct experimental manipulation of the synapse's
	// strength. This is a thread-safe method.
	//
	// Parameters:
	//   weight: The new weight to set for this synapse
	//
	// This method is useful for:
	// - Experimental manipulation
	// - Initialization of specific weight patterns
	// - Testing network behavior with controlled weights
	SetWeight(weight float64)
}

// =================================================================================
// NEURON INTERFACE FOR SYNAPSE COMMUNICATION
// This defines what methods a neuron must have to work with synapses
// =================================================================================

// SynapseCompatibleNeuron defines the interface that neurons must implement
// to work with the synapse system. This allows synapses to communicate with
// neurons without depending on specific neuron implementations.
type SynapseCompatibleNeuron interface {
	// ID returns the unique identifier of the neuron
	ID() string

	// Receive accepts a synapse message and processes it
	// This method should be added to existing neuron implementations
	Receive(msg SynapseMessage)
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
//  2. EFFICIENT DELAYS: Uses time.AfterFunc() for handling transmission delays
//     without blocking the neuron's processing.
//
//  3. THREAD-SAFE: All methods are thread-safe to allow plasticity feedback
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
func NewBasicSynapse(id string, pre SynapseCompatibleNeuron, post SynapseCompatibleNeuron, stdpConfig STDPConfig, pruningConfig PruningConfig, initialWeight float64, delay time.Duration) *BasicSynapse {
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
// This is the core method that models synaptic transmission in biological networks.
//
// Biological process modeled:
// 1. Pre-synaptic neuron fires (signalValue represents action potential strength)
// 2. Signal travels down axon (axonal delay component)
// 3. Neurotransmitter is released at synapse (synaptic delay component)
// 4. Signal strength is modulated by synaptic efficacy (weight multiplication)
// 5. Post-synaptic neuron receives the scaled, delayed signal
//
// Parameters:
//
//	signalValue: The strength of the incoming signal from the pre-synaptic neuron
//
// The method is thread-safe and non-blocking, using time.AfterFunc to handle
// delays efficiently without creating additional goroutines or blocking the
// pre-synaptic neuron's processing.
func (s *BasicSynapse) Transmit(signalValue float64) {
	// Thread-safe read of current synapse state
	s.mutex.RLock()
	effectiveSignal := signalValue * s.weight // Apply synaptic weight
	currentDelay := s.delay                   // Get current delay
	s.mutex.RUnlock()

	// Update activity tracking for pruning decisions
	s.mutex.Lock()
	s.lastTransmission = time.Now()
	s.mutex.Unlock()

	// Create the message to be delivered to the post-synaptic neuron
	// Uses the precise timing information required for STDP learning
	message := SynapseMessage{
		Value:     effectiveSignal,
		Timestamp: time.Now(),               // When the signal was generated
		SourceID:  s.preSynapticNeuron.ID(), // Which neuron sent the signal
		SynapseID: s.id,                     // Which synapse transmitted it
	}

	// Use time.AfterFunc for efficient, non-blocking delay handling.
	// This avoids creating goroutines per transmission while still
	// providing accurate timing for biological realism.
	time.AfterFunc(currentDelay, func() {
		// Deliver the message to the post-synaptic neuron after the delay
		// The timestamp in the message reflects when it was originally sent,
		// allowing the receiving neuron to calculate actual transmission timing
		s.postSynapticNeuron.Receive(message)
	})
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
// neural networks, where weak or inactive synapses are naturally eliminated.
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
		LearningRate:   0.01,                   // 1% weight change per STDP event
		TimeConstant:   20 * time.Millisecond,  // Standard cortical time constant
		WindowSize:     100 * time.Millisecond, // ±100ms learning window
		MinWeight:      0.001,                  // Prevent complete elimination
		MaxWeight:      2.0,                    // Prevent runaway strengthening
		AsymmetryRatio: 1.2,                    // Slight LTD bias (biologically typical)
	}
}

// CreateDefaultPruningConfig returns a standard pruning configuration
// This provides conservative pruning parameters suitable for stable learning
func CreateDefaultPruningConfig() PruningConfig {
	return PruningConfig{
		Enabled:             true,
		WeightThreshold:     0.01,            // Consider weak if weight < 1% of max
		InactivityThreshold: 5 * time.Minute, // Must be inactive for 5 minutes
	}
}

// CreateConservativePruningConfig returns a more conservative pruning configuration
// Use this when you want to minimize the risk of losing important connections
func CreateConservativePruningConfig() PruningConfig {
	return PruningConfig{
		Enabled:             true,
		WeightThreshold:     0.001,            // Very low threshold
		InactivityThreshold: 30 * time.Minute, // Much longer grace period
	}
}
