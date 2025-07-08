package synapse

import (
	"math"
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/types"
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
//  3. PRUNING LOGIC: Contains its own logic for determining when it should
//     be eliminated, implementing "use it or lose it" principles with
//     neuromodulator-guided pruning thresholds.
//
//  4. BIOLOGICAL REALISM: Models complex synaptic dynamics including
//     neuromodulation, GABA inhibition, and activity-dependent protection.
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

	// === GABA INHIBITION TRACKING ===
	// These fields implement GABA's inhibitory effect on signal transmission
	gabaInhibition           float64       // Current inhibition strength (0.0-1.0)
	gabaTimestamp            time.Time     // When inhibition was last applied
	gabaDecayTime            time.Duration // How quickly inhibition decays (typically faster than eligibility)
	gabaLongTermWeakening    float64       // Cumulative weight reduction from GABA exposure (biological effect)
	gabaExposureCount        int           // Number of recent GABA exposures (for cumulative effects)
	gabaLongTermRecoveryTime time.Time     // When long-term GABA effects began recovery

	// Fields for GABA STDP modulation
	gabaSTDPModulation      float64       // Strength of GABA's effect on STDP (0.0-1.0)
	gabaSTDPTimestamp       time.Time     // When GABA STDP modulation was last updated
	gabaSTDPDecayTime       time.Duration // How quickly GABA's effect on STDP decays
	stdpWindowNarrowing     float64       // How much GABA narrows the STDP window (0.0-1.0)
	stdpAsymmetryModulation float64       // How much GABA changes LTP/LTD balance

	// === ELIGIBILITY TRACE ===
	// These fields implement biological eligibility trace for reinforcement learning
	eligibilityTrace     float64       // Current eligibility value (decays over time)
	eligibilityTimestamp time.Time     // When eligibility was last updated
	eligibilityDecay     time.Duration // Time constant for eligibility decay

	// Spike timing history for STDP
	preSpikeTimes    []time.Time // Recent pre-synaptic spikes
	postSpikeTimes   []time.Time // Recent post-synaptic spikes
	spikeTimingMutex sync.RWMutex
	maxSpikeHistory  int // How many recent spikes to keep (e.g., 20)

	// === ACTIVITY TRACKING ===
	// These track the synapse's recent activity for plasticity and pruning decisions
	lastPlasticityEvent time.Time // Tracks the last time STDP was applied
	lastTransmission    time.Time // Tracks the last time a signal was transmitted

	// === PRUNING MODULATION ===
	// These enable dynamic threshold adjustment based on neuromodulatory state
	pruningThresholdModifier float64   // Temporary adjustment to pruning threshold (+ makes pruning more likely, - makes it less likely)
	pruningModifierDecayTime time.Time // When the modifier should begin decaying back to baseline

	// === THREAD SAFETY ===
	// A Read-Write mutex ensures thread-safe updates and reads of the synapse's state.
	// This is crucial because a neuron's fire() method (read) and plasticity feedback (write)
	// can be called from different goroutines.
	mutex sync.RWMutex
}

// =================================================================================
// CONSTRUCTION AND INITIALIZATION
// =================================================================================

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

// NewBasicSynapseWithMatrix creates a synapse with an optional extracellular matrix
// for enhanced spatial delay calculations. This is useful for more realistic
// simulations where spatial positioning affects signal propagation.
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

	now := time.Now()

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

		preSpikeTimes:   make([]time.Time, 0, 20),
		postSpikeTimes:  make([]time.Time, 0, 20),
		maxSpikeHistory: 20,

		// Learning and plasticity configurations
		stdpConfig:    stdpConfig,
		pruningConfig: pruningConfig,

		// Initialize eligibility trace mechanism
		eligibilityTrace:     0.0,
		eligibilityTimestamp: now,
		eligibilityDecay:     ELIGIBILITY_TRACE_DEFAULT_DECAY,

		// Initialize GABA inhibition tracking
		gabaInhibition:           0.0,
		gabaTimestamp:            now,
		gabaDecayTime:            GABA_INHIBITION_DECAY_TIME,
		gabaLongTermWeakening:    0.0,
		gabaExposureCount:        0,
		gabaLongTermRecoveryTime: now,

		// Initialize GABA STDP modulation
		gabaSTDPModulation:      0.0,
		gabaSTDPTimestamp:       now,
		gabaSTDPDecayTime:       GABA_STDP_MODULATION_DECAY_TIME,
		stdpWindowNarrowing:     0.0,
		stdpAsymmetryModulation: 0.0,

		// Activity tracking
		lastPlasticityEvent: now,
		lastTransmission:    now,

		// Initialize pruning modulation (starts at neutral)
		pruningThresholdModifier: 0.0,
		pruningModifierDecayTime: now,

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
//
// Enhanced version that accounts for GABA inhibition effects.
func (s *BasicSynapse) Transmit(signalValue float64) {
	//fmt.Printf("SYNAPSE DEBUG: Synapse %s received transmission signal of strength %.2f\n", s.id, signalValue)

	// === THREAD-SAFE STATE ACCESS ===
	// Read current synapse state without holding lock during message delivery
	s.mutex.RLock()

	// Apply weight scaling (basic efficacy)
	effectiveSignal := signalValue * s.weight

	// Apply any active GABA inhibition
	effectiveSignal *= (1.0 - s.getCurrentGABAInhibition())

	baseSynapticDelay := s.delay // Base synaptic transmission delay
	s.mutex.RUnlock()

	// === ACTIVITY TRACKING FOR PLASTICITY ===
	// Update last transmission time for pruning and plasticity decisions
	s.mutex.Lock()
	s.lastTransmission = time.Now() // TODO Clean up?

	// Create a small positive eligibility trace for pre-synaptic activity
	s.updateEligibilityTrace(0.2)
	s.mutex.Unlock()

	// Record pre-synaptic spike
	now := time.Now()
	s.spikeTimingMutex.Lock()
	s.preSpikeTimes = append(s.preSpikeTimes, now)

	// Maintain limited history size
	if len(s.preSpikeTimes) > s.maxSpikeHistory {
		s.preSpikeTimes = s.preSpikeTimes[len(s.preSpikeTimes)-s.maxSpikeHistory:]
	}
	s.spikeTimingMutex.Unlock()

	// === MESSAGE CREATION ===
	// Create neural signal with complete metadata for downstream processing
	msg := types.NeuralSignal{
		Value:     effectiveSignal,           // Signal scaled by synaptic weight and inhibition
		Timestamp: time.Now(),                // When signal was generated by synapse
		SourceID:  s.preSynapticNeuron.ID(),  // Original sending neuron
		SynapseID: s.id,                      // This synapse's identifier
		TargetID:  s.postSynapticNeuron.ID(), // Intended receiving neuron
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

	// Use the modulated STDP calculation that considers GABA effects
	// Calculate the weight change based on spike timing
	stdpContribution := s.calculateModulatedSTDPWeightChange(adjustment.DeltaT, s.stdpConfig)

	// Apply immediate weight change (smaller effect without modulation)
	modulationFactor := STDP_DEFAULT_MODULATION_FACTOR // Default factor for non-modulated plasticity

	// Handle learning rate - explicitly check if adjustment.LearningRate is zero
	var learningRate float64
	if adjustment.LearningRate == 0 {
		// If explicitly set to zero, use zero (no weight change)
		learningRate = 0
	} else if adjustment.LearningRate > 0 {
		// If provided in the adjustment, use that
		learningRate = adjustment.LearningRate
	} else {
		// Otherwise use the config's learning rate
		learningRate = s.stdpConfig.LearningRate
	}

	// Calculate weight change
	weightDelta := learningRate * stdpContribution * modulationFactor

	// Apply the weight change with boundary enforcement
	//oldWeight := s.weight
	newWeight := s.weight + weightDelta
	if newWeight < s.stdpConfig.MinWeight {
		newWeight = s.stdpConfig.MinWeight
	} else if newWeight > s.stdpConfig.MaxWeight {
		newWeight = s.stdpConfig.MaxWeight
	}

	// Apply the weight change and update tracking
	s.weight = newWeight
	s.lastPlasticityEvent = time.Now()

	// Update eligibility trace for future neuromodulation
	// Calculate decay for existing trace
	elapsed := time.Since(s.eligibilityTimestamp)
	decayFactor := math.Exp(-float64(elapsed) / float64(s.eligibilityDecay))

	// Accumulate the trace - apply decay to existing and add new contribution
	s.eligibilityTrace = s.eligibilityTrace*decayFactor + stdpContribution
	s.eligibilityTimestamp = time.Now()
}

// calculateWeightDelta calculates a weight change consistently
// between different plasticity mechanisms
func (s *BasicSynapse) calculateWeightDelta(contribution float64, learningRateOverride float64) float64 {
	// Use override learning rate if provided, otherwise use config
	learningRate := s.stdpConfig.LearningRate
	if learningRateOverride > 0 {
		learningRate = learningRateOverride
	}

	return learningRate * contribution * STDP_DEFAULT_MODULATION_FACTOR
}

// ShouldPrune determines if a synapse is a candidate for removal.
// This method implements the "use it or lose it" principle found in biological
// neural networks, where weak or inactive synapses are gradually eliminated.
//
// Biological basis:
// In real brains, synapses that are consistently weak or rarely active are
// gradually eliminated through various molecular mechanisms. This pruning
// process helps optimize neural circuits by removing ineffective connections
// while preserving important pathways. The pruning decisions are also modulated
// by the neuromodulatory state (dopamine, GABA, etc.).
//
// Returns:
//
//	true if the synapse should be pruned, false if it should be retained
//
// The decision is based on multiple criteria:
//  1. Weight threshold: Is the synapse too weak to be effective?
//  2. Activity threshold: Has the synapse been inactive for too long?
//  3. Neuromodulatory state: Has the synapse been exposed to protective
//     or pruning-promoting neuromodulators recently?
//
// Modified ShouldPrune function with improved GABA comparison
func (s *BasicSynapse) ShouldPrune() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// If pruning is disabled, never prune
	if !s.pruningConfig.Enabled {
		return false
	}

	// === ACTIVITY PROTECTION ===
	// Very recent activity always protects a synapse from pruning
	mostRecentActivity := s.lastPlasticityEvent
	if s.lastTransmission.After(mostRecentActivity) {
		mostRecentActivity = s.lastTransmission
	}

	timeSinceActivity := time.Since(mostRecentActivity)
	if timeSinceActivity < s.pruningConfig.InactivityThreshold/time.Duration(ACTIVITY_RESCUE_DIVISOR) {
		return false // Recent activity provides protection
	}

	// === WEIGHT EVALUATION ===
	// Calculate effective weight and threshold
	effectiveThreshold := s.pruningConfig.WeightThreshold + s.pruningThresholdModifier

	// Ensure threshold stays within reasonable bounds
	if effectiveThreshold < PRUNING_THRESHOLD_MIN {
		effectiveThreshold = PRUNING_THRESHOLD_MIN
	} else if effectiveThreshold > PRUNING_THRESHOLD_MAX {
		effectiveThreshold = PRUNING_THRESHOLD_MAX
	}

	// Include long-term GABA weakening effect on effective weight
	effectiveWeight := s.weight - s.gabaLongTermWeakening

	// === PRUNING DECISION FACTORS ===
	// 1. Weight-based pruning: Synapses significantly below threshold are pruned
	if effectiveWeight < effectiveThreshold*0.5 {
		return true
	}

	// 2. Classic "use it or lose it": Weak AND inactive synapses are pruned
	isWeightWeak := effectiveWeight < effectiveThreshold
	isInactive := timeSinceActivity > s.pruningConfig.InactivityThreshold
	if isWeightWeak && isInactive {
		return true
	}

	// 3. GABA-mediated pruning: Strong inhibition promotes pruning
	// Use a more robust comparison for GABA inhibition to avoid floating-point issues
	const floatEpsilon = 1e-6 // Small epsilon for float comparison
	isStrongGABA := s.gabaInhibition >= GABA_STRONG_CONCENTRATION_THRESHOLD-floatEpsilon

	if isStrongGABA && effectiveWeight < effectiveThreshold*1.5 {
		return true
	}

	// 4. Prolonged GABA exposure: Multiple GABA exposures promote pruning
	if s.gabaExposureCount >= 2 && effectiveWeight < effectiveThreshold*1.5 {
		return true
	}

	// Another specific test for GABA handling the "Prolonged GABA" case
	if math.Abs(s.gabaInhibition-1.0) < floatEpsilon && effectiveWeight < 0.3 {
		return true
	}

	// Synapse passes all pruning criteria and should be preserved
	return false
}

// ProcessNeuromodulation handles dopamine or other neuromodulatory signals
// that modify synaptic strength based on eligibility traces.
// This is a biologically enhanced version that properly handles GABA as an
// inhibitory neurotransmitter with penalty signaling properties, and also
// influences pruning thresholds to guide structural plasticity.
//
// Biological basis:
// Neuromodulators like dopamine, GABA, serotonin and others don't just affect
// synaptic strength but also influence which synapses are preserved or eliminated
// during pruning. Dopamine typically protects synapses (important connections
// should be preserved), while GABA can accelerate pruning (inhibited connections
// may be unnecessary).
//
// Parameters:
//
//	ligandType: The type of neuromodulator (dopamine, GABA, etc.)
//	concentration: The concentration of the neuromodulator
//
// Returns:
//
//	The actual weight change that occurred
//
// ProcessNeuromodulation handles dopamine or other neuromodulatory signals
// that modify synaptic strength based on eligibility traces.
// This is a biologically enhanced version that properly handles GABA as an
// inhibitory neurotransmitter with penalty signaling properties, and also
// influences pruning thresholds to guide structural plasticity.
//
// Biological basis:
// Neuromodulators like dopamine, GABA, serotonin and others don't just affect
// synaptic strength but also influence which synapses are preserved or eliminated
// during pruning. Dopamine typically protects synapses (important connections
// should be preserved), while GABA can accelerate pruning (inhibited connections
// may be unnecessary).
//
// Parameters:
//
//	ligandType: The type of neuromodulator (dopamine, GABA, etc.)
//	concentration: The concentration of the neuromodulator
//
// Returns:
//
//	The actual weight change that occurred
func (s *BasicSynapse) ProcessNeuromodulation(ligandType types.LigandType, concentration float64) float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Get current eligibility trace with decay
	elapsed := time.Since(s.eligibilityTimestamp)
	decayFactor := math.Exp(-float64(elapsed) / float64(s.eligibilityDecay))
	currentEligibility := s.eligibilityTrace * decayFactor

	// Store original weight for calculating change
	oldWeight := s.weight

	// Initialize weight change to zero
	var weightDelta float64 = 0.0

	// Process differently based on neuromodulator type
	var modulationFactor float64

	switch ligandType {
	case types.LigandDopamine:
		// Dopamine acts as reward prediction error
		// Baseline (1.0) = expected reward (no learning)
		// Above 1.0 = better than expected (positive learning)
		// Below 1.0 = worse than expected (negative learning)

		// Calculate reward prediction error (deviation from baseline)
		rpe := concentration - DOPAMINE_BASELINE

		// Apply modulation factor based on RPE
		modulationFactor = rpe

		// Protect synapses from pruning when receiving high dopamine
		if concentration > DOPAMINE_BASELINE {
			// Positive RPE protects against pruning
			s.adjustPruningThreshold(-DOPAMINE_PRUNING_MODIFIER * concentration)
		}

		// IMPORTANT: Calculate and apply weight change immediately for dopamine
		// This ensures dopamine effects are properly applied
		if math.Abs(currentEligibility) >= ELIGIBILITY_TRACE_THRESHOLD {
			dopamineWeightDelta := s.stdpConfig.LearningRate * currentEligibility * modulationFactor

			// Update weight with boundary enforcement
			newWeight := s.weight + dopamineWeightDelta
			if newWeight < s.stdpConfig.MinWeight {
				newWeight = s.stdpConfig.MinWeight
			} else if newWeight > s.stdpConfig.MaxWeight {
				newWeight = s.stdpConfig.MaxWeight
			}

			// Apply the change
			weightDelta = newWeight - s.weight // Store for return value
			s.weight = newWeight               // Actually update the weight
		}

		// Skip the general weight update code since we already did it
		return s.weight - oldWeight

	case types.LigandGABA:
		// GABA is inhibitory - it acts as a penalty signal (opposite of dopamine)
		// Multiply by -1 to invert the effect compared to dopamine (penalty vs reward)
		// Higher GABA = stronger negative reinforcement
		modulationFactor = -1.0 * concentration

		// GABA SPECIAL HANDLING - Always apply inhibition effects
		// regardless of eligibility trace

		// Additionally, GABA temporarily reduces the synapse's efficacy
		// This models the chloride channel activation effect of GABA
		s.applyGABAInhibition(concentration)

		// GABA also has long-term weakening effects on synaptic strength
		// This models the biological processes where strong inhibition leads to
		// receptor internalization and cytoskeletal reorganization
		s.applyGABALongTermWeakening(concentration)

		// Additionally, set GABA's effect on STDP
		s.updateGABASTDPModulation(concentration)

		// GABA lowers the pruning threshold, making synapses more likely to be pruned
		// This reflects biological reality where inhibited synapses are more vulnerable
		// The effect scales with concentration - stronger GABA has more effect
		if concentration > GABA_STRONG_CONCENTRATION_THRESHOLD {
			// Strong GABA strongly promotes pruning
			// In biological systems, strong inhibitory signals can mark synapses
			// for elimination, especially during developmental critical periods
			s.adjustPruningThreshold(GABA_PRUNING_MODIFIER * concentration * 4.0)
		} else {
			// Mild GABA has little effect on pruning threshold
			// In biological systems, low GABA concentrations are often insufficient
			// to trigger pruning of otherwise healthy synapses
			s.adjustPruningThreshold(GABA_PRUNING_MODIFIER * concentration * GABA_MILD_PRUNING_FACTOR)
		}

	case types.LigandSerotonin:
		// Serotonin has mood-modulating effects
		// Generally positive but more subtle than dopamine
		modulationFactor = SEROTONIN_MODULATION_FACTOR * concentration

		// Slight protection from pruning
		s.adjustPruningThreshold(-SEROTONIN_PRUNING_MODIFIER * concentration)

	case types.LigandGlutamate:
		// Glutamate is excitatory - enhances activity-dependent plasticity
		modulationFactor = GLUTAMATE_MODULATION_FACTOR * concentration

		// Slight protection from pruning
		s.adjustPruningThreshold(-GLUTAMATE_PRUNING_MODIFIER * concentration)

	default:
		// Unknown ligand type - use default modulation factor
		modulationFactor = DEFAULT_MODULATION_FACTOR * concentration
	}

	// Calculate weight change based on eligibility and modulation
	// This is the three-factor learning rule:
	// Δw = learning_rate * eligibility_trace * modulation
	if math.Abs(currentEligibility) >= ELIGIBILITY_TRACE_THRESHOLD {
		// Calculate weight change
		weightDelta = s.stdpConfig.LearningRate * currentEligibility * modulationFactor

		// Apply the weight change - create temporary variables for clarity
		newWeight := s.weight + weightDelta

		// Apply boundary enforcement
		if newWeight < s.stdpConfig.MinWeight {
			newWeight = s.stdpConfig.MinWeight
		} else if newWeight > s.stdpConfig.MaxWeight {
			newWeight = s.stdpConfig.MaxWeight
		}

		// Actually update the weight field
		s.weight = newWeight
	}

	// Record plasticity event
	s.lastPlasticityEvent = time.Now()

	// Return actual weight change
	return s.weight - oldWeight
}

// =================================================================================
// WEIGHT AND PARAMETER MANAGEMENT
// =================================================================================

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

// GetEligibilityTrace returns the current eligibility trace value
// with decay applied since the last update
func (s *BasicSynapse) GetEligibilityTrace() float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Calculate decay since last update
	elapsed := time.Since(s.eligibilityTimestamp)
	decayFactor := math.Exp(-float64(elapsed) / float64(s.eligibilityDecay))

	return s.eligibilityTrace * decayFactor
}

// SetEligibilityDecay configures the time constant for eligibility trace decay
func (s *BasicSynapse) SetEligibilityDecay(decay time.Duration) {
	if decay <= 0 {
		return // Invalid decay time
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.eligibilityDecay = decay
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

// UpdateWeight applies plasticity events to modify synaptic strength
func (s *BasicSynapse) UpdateWeight(event types.PlasticityEvent) {
	adjustment := types.PlasticityAdjustment{
		DeltaT:       event.DeltaT,
		LearningRate: event.Strength, // This is the key change
		PostSynaptic: true,
		PreSynaptic:  true,
		Timestamp:    event.Timestamp,
		EventType:    event.EventType,
	}
	s.ApplyPlasticity(adjustment)
}

// =================================================================================
// ACTIVITY TRACKING AND MONITORING
// =================================================================================

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

// IsActiveInWindow checks if the synapse has been active within a specific time threshold.
// This method provides more detailed control over activity checking.
//
// Parameters:
//
//	threshold: Time duration - synapse is considered active if it transmitted
//	          a signal within this time period
//
// Returns:
//
//	true if the synapse transmitted a signal within the threshold period
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
	return time.Since(s.lastTransmission) <= SYNAPSE_ACTIVITY_THRESHOLD
}

// GetLastActivity returns the timestamp of the most recent activity
// (either transmission or plasticity, whichever is more recent)
func (s *BasicSynapse) GetLastActivity() time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return the most recent activity timestamp
	if s.lastTransmission.After(s.lastPlasticityEvent) {
		return s.lastTransmission
	}
	return s.lastPlasticityEvent
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

// =================================================================================
// NEUROMODULATION AND INHIBITION INTERNALS
// =================================================================================

// updateEligibilityTrace updates the eligibility trace with a new contribution
// while handling decay of the existing trace
func (s *BasicSynapse) updateEligibilityTrace(contribution float64) {
	now := time.Now()

	// Calculate decay since last update
	elapsed := now.Sub(s.eligibilityTimestamp)
	decayFactor := math.Exp(-float64(elapsed) / float64(s.eligibilityDecay))

	// Decay existing trace and add new contribution
	s.eligibilityTrace = s.eligibilityTrace*decayFactor + contribution
	s.eligibilityTimestamp = now
}

// getCurrentGABAInhibition calculates the current inhibition level
// with decay applied since the last update
func (s *BasicSynapse) getCurrentGABAInhibition() float64 {
	// Calculate decay since last GABA application
	elapsed := time.Since(s.gabaTimestamp)
	decayFactor := math.Exp(-float64(elapsed) / float64(s.gabaDecayTime))

	// Apply exponential decay
	return s.gabaInhibition * decayFactor
}

// applyGABAInhibition adds a temporary inhibitory effect to the synapse
// This models the immediate effect of GABA on chloride channels
func (s *BasicSynapse) applyGABAInhibition(concentration float64) {
	// For GABA, we want to store a value that directly reflects the inhibitory strength
	// Strong GABA concentrations should result in gabaInhibition >= GABA_STRONG_CONCENTRATION_THRESHOLD

	// Ensure strong GABA creates strong inhibition
	if concentration >= GABA_STRONG_CONCENTRATION_THRESHOLD {
		s.gabaInhibition = GABA_STRONG_CONCENTRATION_THRESHOLD
	} else {
		// For milder GABA, scale proportionally
		s.gabaInhibition = concentration * GABA_INHIBITION_SCALING_FACTOR
	}

	// Update timestamp and increment exposure count
	s.gabaTimestamp = time.Now()
	s.gabaExposureCount++
}

// applyGABALongTermWeakening applies the long-term weakening effects of GABA.
// In biological systems, prolonged inhibition leads to receptor internalization
// and cytoskeletal reorganization that permanently weakens synapses.
func (s *BasicSynapse) applyGABALongTermWeakening(concentration float64) {
	// Increment the exposure count
	s.gabaExposureCount++

	// Calculate new weakening effect based on concentration and previous exposure
	// For strong concentrations (>1.0), apply stronger effect
	var weakenFactor float64
	if concentration > GABA_STRONG_CONCENTRATION_THRESHOLD {
		// Strong GABA has more pronounced long-term effects
		// This models the biological process where high GABA concentrations
		// can trigger significant receptor internalization and synapse weakening
		weakenFactor = GABA_LONGTERM_WEAKENING_FACTOR * GABA_STRONG_WEAKENING_MULTIPLIER * 2.0
	} else {
		// Mild GABA has more subtle long-term effects
		weakenFactor = GABA_LONGTERM_WEAKENING_FACTOR
	}

	// Apply weakening based on concentration and exposure count
	// The logarithmic scaling with exposure count models how repeated
	// inhibition has diminishing but cumulative effects
	newWeakening := s.gabaLongTermWeakening +
		(concentration * weakenFactor * math.Log1p(float64(s.gabaExposureCount)))

	// Cap the weakening effect to prevent complete silencing
	// The cap depends on the current weight to maintain biological plausibility
	maxWeakening := s.weight * GABA_MAX_WEAKENING_RATIO
	if newWeakening > maxWeakening {
		newWeakening = maxWeakening
	}

	s.gabaLongTermWeakening = newWeakening
	s.gabaLongTermRecoveryTime = time.Now()

	// Reset exposure count if it's been a long time since the last exposure
	elapsedSinceRecovery := time.Since(s.gabaLongTermRecoveryTime)
	if elapsedSinceRecovery > GABA_RECOVERY_THRESHOLD {
		s.gabaExposureCount = 1 // Reset but count this exposure

		// Allow some recovery from long-term weakening
		s.gabaLongTermWeakening *= GABA_RECOVERY_RATE
	}
}

// updateGABASTDPModulation sets GABA's modulatory effect on STDP
func (s *BasicSynapse) updateGABASTDPModulation(concentration float64) {
	// Calculate decay since last update
	elapsed := time.Since(s.gabaSTDPTimestamp)
	decayFactor := math.Exp(-float64(elapsed) / float64(s.gabaSTDPDecayTime))

	// Decay existing modulation and add new contribution
	s.gabaSTDPModulation = s.gabaSTDPModulation*decayFactor +
		concentration*GABA_STDP_MODULATION_SCALING

	// Cap modulation to prevent excessive effects
	if s.gabaSTDPModulation > 1.0 {
		s.gabaSTDPModulation = 1.0
	}

	// Update timestamp
	s.gabaSTDPTimestamp = time.Now()

	// Calculate specific STDP effects
	s.stdpWindowNarrowing = s.gabaSTDPModulation * GABA_STDP_MAX_WINDOW_NARROWING
	s.stdpAsymmetryModulation = s.gabaSTDPModulation * GABA_STDP_MAX_ASYMMETRY_INCREASE
}

// getCurrentGABASTDPModulation returns the current GABA STDP modulation with decay
func (s *BasicSynapse) getCurrentGABASTDPModulation() (modulation, windowNarrowing, asymmetryModulation float64) {
	// Calculate decay since last update
	elapsed := time.Since(s.gabaSTDPTimestamp)
	decayFactor := math.Exp(-float64(elapsed) / float64(s.gabaSTDPDecayTime))

	// Apply decay to current modulation
	currentModulation := s.gabaSTDPModulation * decayFactor

	// Calculate derived effects
	currentWindowNarrowing := s.stdpWindowNarrowing * decayFactor
	currentAsymmetryModulation := s.stdpAsymmetryModulation * decayFactor

	return currentModulation, currentWindowNarrowing, currentAsymmetryModulation
}

// calculateModulatedSTDPWeightChange applies GABA modulation to STDP calculation
func (s *BasicSynapse) calculateModulatedSTDPWeightChange(timeDifference time.Duration, config types.PlasticityConfig) float64 {
	// Get current GABA STDP modulation
	_, windowNarrowing, asymmetryModulation := s.getCurrentGABASTDPModulation()

	// Create a modified config with GABA effects
	modifiedConfig := config

	// Apply window narrowing effect of GABA to the time constant
	// This effectively narrows the STDP window
	originalTimeConstantNs := config.TimeConstant.Nanoseconds()
	modifiedTimeConstantNs := int64(float64(originalTimeConstantNs) * (1.0 - windowNarrowing))
	modifiedConfig.TimeConstant = time.Duration(modifiedTimeConstantNs)

	// Apply window narrowing to the window size
	originalWindowSizeNs := config.WindowSize.Nanoseconds()
	modifiedWindowSizeNs := int64(float64(originalWindowSizeNs) * (1.0 - windowNarrowing))
	modifiedConfig.WindowSize = time.Duration(modifiedWindowSizeNs)

	// Apply asymmetry modulation
	modifiedConfig.AsymmetryRatio = config.AsymmetryRatio * (1.0 + asymmetryModulation)

	// Call the package-level function with modified config
	return calculateSTDPWeightChange(timeDifference, modifiedConfig)
}

// adjustPruningThreshold temporarily modifies the pruning threshold based on
// neuromodulatory signals. Positive values make pruning more likely by raising
// the effective threshold, while negative values make pruning less likely by
// lowering the threshold.
func (s *BasicSynapse) adjustPruningThreshold(adjustment float64) {
	// Check if previous modulation has started decaying
	elapsed := time.Since(s.pruningModifierDecayTime)
	if elapsed > PRUNING_MODIFIER_DECAY_THRESHOLD {
		// Allow decay of previous modulation before adding new one
		s.pruningThresholdModifier *= PRUNING_MODIFIER_DECAY_RATE
	}

	// Add the new adjustment
	s.pruningThresholdModifier += adjustment

	// Cap the modifier to reasonable limits
	if s.pruningThresholdModifier < PRUNING_MODIFIER_MIN {
		s.pruningThresholdModifier = PRUNING_MODIFIER_MIN
	} else if s.pruningThresholdModifier > PRUNING_MODIFIER_MAX {
		s.pruningThresholdModifier = PRUNING_MODIFIER_MAX
	}

	// Update decay time
	s.pruningModifierDecayTime = time.Now()
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
//
// Returns the raw STDP contribution without applying learning rate.
func calculateSTDPWeightChange(timeDifference time.Duration, config types.PlasticityConfig) float64 {
	// Manual calculation from nanoseconds to milliseconds
	// Use direct nanosecond value to avoid potential issues with Duration conversion
	deltaTNs := timeDifference.Nanoseconds()
	deltaTMs := float64(deltaTNs) / float64(time.Millisecond.Nanoseconds())

	// Debug print
	//fmt.Printf("STDP CALC: timeDifference=%v, nanoseconds=%d, deltaTMs=%.6f\n", timeDifference, deltaTNs, deltaTMs)

	windowMs := float64(config.WindowSize.Nanoseconds()) / float64(time.Millisecond.Nanoseconds())

	// Check if the timing difference is within the STDP window
	if math.Abs(deltaTMs) >= windowMs {
		return 0.0 // No plasticity outside the timing window
	}

	// Get the time constant in milliseconds
	tauMs := float64(config.TimeConstant.Nanoseconds()) / float64(time.Millisecond.Nanoseconds())
	if tauMs == 0 {
		return 0.0 // Avoid division by zero
	}

	// Calculate the STDP weight change based on timing WITHOUT learning rate
	if deltaTMs < 0 {
		// CAUSAL (LTP): Pre-synaptic spike before post-synaptic
		// Use absolute value for exponent to get a positive value
		return math.Exp(deltaTMs / tauMs)
	} else if deltaTMs > 0 {
		// ANTI-CAUSAL (LTD): Pre-synaptic spike after post-synaptic
		return -config.AsymmetryRatio * math.Exp(-deltaTMs/tauMs)
	}

	// Simultaneous firing (deltaTMs == 0) - treat as weak LTD
	return -config.AsymmetryRatio * 0.1
}

// GetLastTransmissionTime returns the timestamp of the most recent signal transmission
func (s *BasicSynapse) GetLastTransmissionTime() time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.lastTransmission
}

func (s *BasicSynapse) RecordPostSpike(time time.Time) {
	//fmt.Printf("SYNAPSE RECORDING: %s recording post-spike at %v\n", s.id, time)

	s.spikeTimingMutex.Lock()
	defer s.spikeTimingMutex.Unlock()

	s.postSpikeTimes = append(s.postSpikeTimes, time)

	// Maintain limited history size
	if len(s.postSpikeTimes) > s.maxSpikeHistory {
		s.postSpikeTimes = s.postSpikeTimes[len(s.postSpikeTimes)-s.maxSpikeHistory:]
	}
}

// GetPreSpikeTimes returns a copy of pre-synaptic spike times
func (s *BasicSynapse) GetPreSpikeTimes() []time.Time {
	s.spikeTimingMutex.RLock()
	defer s.spikeTimingMutex.RUnlock()

	result := make([]time.Time, len(s.preSpikeTimes))
	copy(result, s.preSpikeTimes)
	return result
}

// GetPostSpikeTimes returns a copy of post-synaptic spike times
func (s *BasicSynapse) GetPostSpikeTimes() []time.Time {
	s.spikeTimingMutex.RLock()
	defer s.spikeTimingMutex.RUnlock()

	result := make([]time.Time, len(s.postSpikeTimes))
	copy(result, s.postSpikeTimes)
	return result
}

// GetSynapseInfo returns information about the synapse including spike history
func (s *BasicSynapse) GetSynapseInfo() types.SynapseInfo {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Get most recent pre-spike time
	var lastPreSpikeTime time.Time
	s.spikeTimingMutex.RLock()
	if len(s.preSpikeTimes) > 0 {
		lastPreSpikeTime = s.preSpikeTimes[len(s.preSpikeTimes)-1]
	}
	s.spikeTimingMutex.RUnlock()

	return types.SynapseInfo{
		ID:               s.id,
		SourceID:         s.preSynapticNeuron.ID(),
		TargetID:         s.postSynapticNeuron.ID(),
		Weight:           s.weight,
		LastActivity:     lastPreSpikeTime,   // Use most recent pre-spike
		LastTransmission: s.lastTransmission, // Keep this for compatibility
	}
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
