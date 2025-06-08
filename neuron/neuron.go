/*
=================================================================================
BIOLOGICAL NEURON SIMULATION - CORE BUILDING BLOCK WITH HOMEOSTATIC PLASTICITY
=================================================================================

OVERVIEW:
This package implements a biologically-inspired artificial neuron that serves as
the fundamental building block for constructing neural networks with dynamic
connectivity, realistic timing behavior, homeostatic self-regulation, and
synaptic scaling for long-term stability.

BIOLOGICAL INSPIRATION:
Real biological neurons are far more complex than traditional artificial neurons
used in deep learning. Key biological features modeled here:

1. TEMPORAL INTEGRATION: Real neurons accumulate electrical signals (postsynaptic
   potentials) over time windows, not instantaneous calculations

2. THRESHOLD FIRING: When accumulated charge reaches a threshold, the neuron fires
   an action potential - an all-or-nothing electrical spike

3. HOMEOSTATIC PLASTICITY: Neurons automatically adjust their firing thresholds
   and sensitivity to maintain stable activity levels, preventing runaway
   excitation or neural silence

4. CALCIUM-BASED ACTIVITY SENSING: Action potentials cause calcium influx which
   serves as a biological activity sensor for homeostatic regulation

5. SYNAPTIC SCALING: Neurons monitor total input strength and proportionally
   adjust their receptor sensitivity to maintain stable responsiveness while
   preserving learned patterns from STDP

6. SPIKE-TIMING DEPENDENT PLASTICITY (STDP): Individual synapses strengthen or
   weaken based on precise timing relationships between pre- and post-synaptic
   spikes, implementing biological learning rules

7. DYNAMIC CONNECTIVITY: Biological neurons can grow new connections (synapses)
   and prune existing ones throughout their lifetime (neuroplasticity)

8. PARALLEL TRANSMISSION: A single action potential propagates to ALL connected
   neurons simultaneously through the axon's branching structure

9. TRANSMISSION DELAYS: Different connections have different delays based on
   axon length, diameter, and myelination

10. SYNAPTIC STRENGTH: Each connection has its own "weight" or strength that
    modulates the signal intensity and adapts through learning

11. REFRACTORY PERIODS: Cannot fire immediately after firing (recovery time)

12. LEAKY INTEGRATION: Membrane potential naturally decays over time

MULTI-TIMESCALE PLASTICITY:
This implementation models the multiple timescales of biological plasticity:

- STDP (milliseconds to seconds): Fast synaptic learning based on spike timing
- Homeostatic Plasticity (seconds to minutes): Intrinsic excitability adjustment
- Synaptic Scaling (minutes to hours): Proportional adjustment of receptor sensitivity

These mechanisms work together to create stable yet adaptive learning:
- STDP learns specific patterns and associations
- Homeostatic plasticity maintains appropriate firing rates
- Synaptic scaling prevents runaway strengthening while preserving learned patterns

HOMEOSTATIC PLASTICITY DETAILS:
Homeostatic plasticity is the neuron's ability to maintain stable activity levels
by automatically adjusting its intrinsic properties. This biological mechanism:

- Prevents network instability (runaway excitation or silence)
- Maintains optimal firing rates without manual intervention
- Uses calcium accumulation as an activity sensor
- Adjusts firing threshold based on recent activity history
- Operates on slower timescales than synaptic plasticity (seconds vs milliseconds)

In real neurons:
- Hyperactive neurons (high calcium) increase their firing threshold
- Silent neurons (low calcium) decrease their firing threshold
- This creates a negative feedback loop that stabilizes network activity
- The process occurs over seconds to minutes, much slower than synaptic events

SYNAPTIC SCALING DETAILS (BIOLOGICALLY ACCURATE):
Synaptic scaling is a homeostatic mechanism that maintains stable total synaptic
input strength while preserving the relative patterns learned through STDP:

- POST-SYNAPTIC CONTROL: The receiving neuron controls its own receptor sensitivity
- RECEPTOR SCALING: Adjusts AMPA/NMDA receptor density at each synapse
- PATTERN PRESERVATION: Maintains relative ratios between different inputs
- PRE-SYNAPTIC INDEPENDENCE: Source neurons remain unaware of scaling
- BIOLOGICAL SEPARATION: Pre-synaptic strength vs post-synaptic sensitivity

Example: If all synapses strengthen 2x due to STDP, post-synaptic scaling reduces
receptor sensitivity by 50% to restore target total strength while maintaining
learned relative patterns.

TRADITIONAL AI vs THIS APPROACH:
- Traditional ANNs: Static connectivity, synchronous processing, mathematical
  activation functions, no realistic timing, no homeostatic regulation
- This neuron: Dynamic connectivity, asynchronous messaging, temporal integration,
  biological timing delays, self-regulating activity levels, multi-timescale
  learning, and automatic synaptic balance maintenance

KEY DESIGN PRINCIPLES:
1. CONCURRENCY: Each neuron runs as an independent goroutine, enabling true
   parallel processing like real neural networks

2. MESSAGE PASSING: Neurons communicate through Go channels, modeling the
   discrete nature of biological action potentials

3. TEMPORAL DYNAMICS: Input accumulation over configurable time windows models
   how real dendrites integrate signals

4. DYNAMIC ARCHITECTURE: Connections can be added/removed at runtime, enabling
   learning and adaptation

5. HOMEOSTATIC REGULATION: Neurons automatically maintain stable activity levels
   through threshold adjustment and activity monitoring

6. SYNAPTIC BALANCE: Automatic receptor scaling maintains optimal input strength
   while preserving learned patterns

7. THREAD SAFETY: Multiple goroutines can safely modify connections while the
   neuron processes inputs

BUILDING BLOCKS FOR LARGER SYSTEMS:
This neuron serves as the foundation for:
- Static neural networks (traditional architectures)
- Dynamic neural networks (growing/pruning connections)
- Gated neural networks (where gates control connectivity)
- Spiking neural networks (event-driven processing)
- Neuromorphic computing systems
- Self-regulating biological brain simulations
- Long-term stable learning networks

USAGE PATTERN:
1. Create neuron with threshold, decay rate, refractory period, and plasticity parameters
2. Connect to other neurons by adding outputs (automatically registers for scaling)
3. Launch as goroutine: go neuron.Run()
4. Send messages to neuron.GetInput() channel
5. Observe automatic homeostatic regulation maintaining stable activity
6. Monitor STDP learning strengthening/weakening specific connections
7. Watch synaptic scaling maintain overall network stability
8. Dynamically modify connections during runtime
9. Monitor emergent network behavior with multi-timescale self-regulation

This approach enables building AI systems that operate more like biological
brains - with realistic timing, dynamic connectivity, self-regulation at multiple
timescales, stable long-term learning, and emergent intelligence arising from
the interaction of many simple, concurrent processing units with sophisticated
internal regulatory mechanisms.

=================================================================================
*/

package neuron

import (
	"math"
	"sync"
	"time"
)

// Message represents a signal transmitted between neurons with spike timing information
// Models the discrete action potential spikes in biological neural communication
// Unlike traditional ANNs that use continuous values, this models the actual
// binary spike-based communication found in real brains
//
// Enhanced with timing information for Spike-Timing Dependent Plasticity (STDP):
// STDP requires precise timing information to determine whether synaptic connections
// should be strengthened or weakened based on the temporal relationship between
// pre-synaptic and post-synaptic spikes
type Message struct {
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
}

// SynapticScalingConfig contains all parameters controlling synaptic scaling behavior
// This structure encapsulates the homeostatic mechanism that maintains synaptic balance
//
// BIOLOGICAL BACKGROUND:
// Synaptic scaling is a homeostatic mechanism observed in real neurons that prevents
// runaway strengthening or weakening of synaptic connections. When total synaptic
// input becomes too strong or weak, neurons proportionally scale their receptor
// sensitivity to maintain optimal responsiveness while preserving learned patterns.
//
// TIMESCALES:
// - STDP: milliseconds to seconds (fast learning)
// - Synaptic Scaling: minutes to hours (slow homeostasis)
// - Homeostatic Plasticity: seconds to minutes (medium regulation)
//
// INTERACTION WITH OTHER MECHANISMS:
// Synaptic scaling works alongside STDP and homeostatic plasticity to create
// a multi-layered regulatory system that enables stable yet adaptive learning.
type SynapticScalingConfig struct {
	Enabled bool // Master switch for synaptic scaling functionality
	// When false, synaptic scaling is completely disabled
	// When true, scaling occurs according to the parameters below
	// Useful for experimental comparisons and debugging

	// === CORE SCALING PARAMETERS ===
	// These control the target behavior and speed of synaptic scaling

	TargetInputStrength float64 // Desired average effective input strength
	// Biological interpretation: optimal total synaptic drive for this neuron
	// This is the target for (synaptic_weight × receptor_gain) averaged across inputs
	// Typical values: 0.5-2.0 depending on neuron type and network role
	// Higher values = neuron expects stronger overall input
	// Lower values = neuron operates with weaker overall input

	ScalingRate float64 // Rate of receptor gain adjustment per scaling event
	// Controls how aggressively gains are adjusted toward target
	// Range: 0.0001 (very conservative) to 0.01 (aggressive)
	// Higher values = faster correction but potentially less stable
	// Lower values = slower correction but more stable learning

	ScalingInterval time.Duration // Time between synaptic scaling operations
	// Biological timescale: much slower than STDP (which operates in milliseconds)
	// Typical range: 10 seconds to 10 minutes
	// Shorter intervals = more responsive but higher computational cost
	// Longer intervals = more efficient but slower adaptation

	// === SAFETY CONSTRAINTS ===
	// These prevent extreme scaling that could destabilize the network

	MinScalingFactor float64 // Minimum multiplier applied to gains per scaling event
	// Prevents excessive reduction in a single scaling operation
	// Typical values: 0.8-0.95 (don't reduce gains by more than 5-20% per event)
	// Protects against sudden loss of important learned connections

	MaxScalingFactor float64 // Maximum multiplier applied to gains per scaling event
	// Prevents excessive increase in a single scaling operation
	// Typical values: 1.05-1.2 (don't increase gains by more than 5-20% per event)
	// Prevents runaway excitation from overly aggressive scaling

	// === STATE TRACKING ===
	// These fields track scaling history and timing for proper operation

	LastScalingUpdate time.Time // Timestamp of most recent scaling operation
	// Used to determine when next scaling should occur based on ScalingInterval
	// Updated automatically each time scaling is performed

	ScalingHistory []float64 // Recent scaling factors for monitoring and analysis
	// Stores the actual scaling factors applied in recent operations
	// Useful for debugging, visualization, and detecting scaling oscillations
	// Limited to recent history to prevent unlimited memory growth
}

// calculateSTDPWeightChange computes the STDP weight change based on spike timing.
// This implements the core STDP learning rule with exponential timing windows.
//
// Time difference convention: Δt = post_spike_time - pre_spike_time
//   - Δt < 0: pre-synaptic spike before post-synaptic spike (causal) -> LTP (strengthen) -> POSITIVE change
//   - Δt > 0: post-synaptic spike before pre-synaptic spike (anti-causal) -> LTD (weaken) -> NEGATIVE change
//
// Mathematical Model:
// For Δt < 0 (causal):   ΔW = +LearningRate * AsymmetryRatio * exp(Δt / τ)
// For Δt > 0 (anti-causal): ΔW = -LearningRate * exp(-Δt / τ)
//
// Parameters:
// timeDifference: The result of postSpikeTime.Sub(preSpikeTime).
// config: STDP configuration containing learning rates and time constants.
//
// Returns: The calculated weight change (positive for LTP, negative for LTD).
func calculateSTDPWeightChange(timeDifference time.Duration, config STDPConfig) float64 {
	deltaT := timeDifference.Seconds() * 1000.0 // Convert to milliseconds
	windowMs := config.WindowSize.Seconds() * 1000.0

	// No plasticity outside the timing window
	if math.Abs(deltaT) >= windowMs {
		return 0.0
	}

	tauMs := config.TimeConstant.Seconds() * 1000.0
	if tauMs == 0 {
		return 0.0 // Avoid division by zero
	}

	// CORRECTED STDP logic
	if deltaT < 0 {
		// CAUSAL (LTP): Pre-synaptic spike occurred *before* post-synaptic (deltaT < 0).
		// Positive change to strengthen the synapse.
		// deltaT is negative, so exp(deltaT/tau) naturally decays as |deltaT| increases
		return config.LearningRate * config.AsymmetryRatio * math.Exp(deltaT/tauMs)
	} else if deltaT > 0 {
		// ANTI-CAUSAL (LTD): Post-synaptic spike occurred *before* pre-synaptic (deltaT > 0).
		// Negative change to weaken the synapse.
		// deltaT is positive, so we use -deltaT/tau for the decay
		return -config.LearningRate * math.Exp(-deltaT/tauMs)
	} else {
		// Simultaneous firing (deltaT == 0). Can be treated as LTD or a special case.
		// Returning a small negative value is a common choice.
		return -config.LearningRate
	}
}

// cleanOldSpikeHistory removes spike events outside the STDP timing window
// This maintains computational efficiency by preventing unlimited memory growth
// and ensures only biologically relevant timing relationships are considered
//
// Biological rationale:
// Real synapses have limited temporal integration windows. Spikes separated
// by more than ~50ms have no effect on synaptic plasticity. This function
// models this biological constraint while optimizing memory usage.
//
// Parameters:
// spikes: slice of spike events to clean
// currentTime: reference time for determining what constitutes "old"
// windowSize: maximum age of spikes to retain
//
// Returns: cleaned slice with only recent spikes
func cleanOldSpikeHistory(spikes []SpikeEvent, currentTime time.Time, windowSize time.Duration) []SpikeEvent {
	if len(spikes) == 0 {
		return spikes
	}

	// Calculate cutoff time - spikes older than this are removed
	cutoffTime := currentTime.Add(-windowSize)

	// Find the first spike that is still within the window
	keepFromIndex := -1
	for i, spike := range spikes {
		if !spike.Timestamp.Before(cutoffTime) {
			keepFromIndex = i
			break
		}
	}

	// If all spikes are too old, return empty slice
	if keepFromIndex == -1 {
		return []SpikeEvent{}
	}

	// Return slice containing only recent spikes
	// Create new slice to allow garbage collection of old events
	recentSpikes := make([]SpikeEvent, len(spikes)-keepFromIndex)
	copy(recentSpikes, spikes[keepFromIndex:])
	return recentSpikes
}

// cleanOldPreSpikeHistory removes old pre-synaptic spike times from synapse history
// Similar to cleanOldSpikeHistory but operates on time.Time slices for efficiency
//
// This function is called on individual synapses to maintain their pre-spike
// timing history within biologically relevant windows
func cleanOldPreSpikeHistory(spikeTimes []time.Time, currentTime time.Time, windowSize time.Duration) []time.Time {
	if len(spikeTimes) == 0 {
		return spikeTimes
	}

	cutoffTime := currentTime.Add(-windowSize)

	// Find first spike time within window
	keepFromIndex := -1
	for i, spikeTime := range spikeTimes {
		if !spikeTime.Before(cutoffTime) {
			keepFromIndex = i
			break
		}
	}

	// If all spikes are too old, return empty slice
	if keepFromIndex == -1 {
		return []time.Time{}
	}

	recentTimes := make([]time.Time, len(spikeTimes)-keepFromIndex)
	copy(recentTimes, spikeTimes[keepFromIndex:])
	return recentTimes
}

// recordPreSynapticSpike records a new pre-synaptic spike for STDP processing
// This method is called on the output synapse when the pre-synaptic neuron fires
//
// Biological context:
// When a neuron fires, all of its output synapses need to record this event
// for future STDP calculations. If any of the target neurons fire later,
// they will look back at this timing information to determine how to modify
// the synaptic weights.
//
// Parameters:
// spikeTime: precise timestamp when the pre-synaptic neuron fired
// config: STDP configuration for cleanup timing
func (o *Output) recordPreSynapticSpike(spikeTime time.Time, config STDPConfig) {
	// Skip recording if STDP is disabled for this synapse
	if !o.stdpEnabled || !config.Enabled {
		return
	}

	// Add this spike time to the history
	o.preSpikeTimes = append(o.preSpikeTimes, spikeTime)
	o.lastPreSpike = spikeTime

	// Clean old spike times to prevent unlimited memory growth
	// Only keep spikes within the STDP timing window
	o.preSpikeTimes = cleanOldPreSpikeHistory(o.preSpikeTimes, spikeTime, config.WindowSize)

	// Limit maximum history size for computational efficiency
	maxHistorySize := 100 // Reasonable limit for most applications
	if len(o.preSpikeTimes) > maxHistorySize {
		// Keep only the most recent spikes
		start := len(o.preSpikeTimes) - maxHistorySize
		o.preSpikeTimes = o.preSpikeTimes[start:]
	}
}

// applySTDPToSynapse modifies synaptic weight based on post-synaptic firing
// This is the core STDP implementation that strengthens or weakens synapses
// based on spike timing relationships
//
// Biological process:
// When the post-synaptic neuron fires, it examines all recent pre-synaptic
// spikes. For each pre-spike, it calculates the timing difference and applies
// the appropriate weight change. This implements the "neurons that fire together,
// wire together" principle with precise temporal requirements.
//
// Parameters:
// postSpikeTime: when the post-synaptic neuron fired
// config: STDP configuration parameters
func (o *Output) applySTDPToSynapse(postSpikeTime time.Time, config STDPConfig) {
	// Skip if STDP is disabled
	if !o.stdpEnabled || !config.Enabled {
		return
	}

	// Process all recent pre-synaptic spikes
	totalWeightChange := 0.0
	for _, preSpikeTime := range o.preSpikeTimes {
		// Calculate timing difference (post - pre)
		timeDifference := postSpikeTime.Sub(preSpikeTime)

		// Calculate STDP weight change for this spike pair
		weightChange := calculateSTDPWeightChange(timeDifference, config)
		totalWeightChange += weightChange
	}

	// Apply the total weight change with bounds checking
	if totalWeightChange != 0.0 {
		o.updateSynapticWeight(totalWeightChange)
	}
}

// updateSynapticWeight safely modifies the synaptic weight within biological bounds
// This function ensures that synaptic weights remain within realistic ranges
// and implements saturation effects observed in biological synapses
//
// Biological constraints:
// - Synapses cannot become infinitely strong (physical saturation)
// - Synapses cannot become negative (unidirectional transmission)
// - Weight changes may be subject to homeostatic scaling
//
// Parameters:
// deltaWeight: proposed change in synaptic weight (can be positive or negative)
func (o *Output) updateSynapticWeight(deltaWeight float64) {
	// Calculate new proposed weight
	newWeight := o.factor + deltaWeight

	// Apply hard bounds to prevent runaway strengthening or elimination
	if newWeight < o.minWeight {
		newWeight = o.minWeight
	} else if newWeight > o.maxWeight {
		newWeight = o.maxWeight
	}

	// Update the synaptic weight
	o.factor = newWeight
}

// getSynapticStrength returns the current synaptic weight for external monitoring
// This provides read-only access to the current synaptic strength for analysis
func (o *Output) getSynapticStrength() float64 {
	return o.factor
}

// getSynapticLearningStats returns statistics about this synapse's learning
// Useful for monitoring and debugging STDP behavior in large networks
func (o *Output) getSynapticLearningStats() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["currentWeight"] = o.factor
	stats["baseWeight"] = o.baseWeight
	stats["minWeight"] = o.minWeight
	stats["maxWeight"] = o.maxWeight
	stats["learningRate"] = o.learningRate
	stats["recentPreSpikes"] = len(o.preSpikeTimes)
	stats["stdpEnabled"] = o.stdpEnabled

	// Calculate weight change percentage from baseline
	if o.baseWeight != 0 {
		stats["weightChangePercent"] = (o.factor - o.baseWeight) / o.baseWeight * 100
	}

	return stats
}

// addRecentInputSpike records an incoming spike for STDP processing
// Called when this neuron receives a spike from another neuron
//
// Biological context:
// When a post-synaptic neuron receives a spike, it needs to record the timing
// and source information. If this neuron fires later, it will use this information
// to apply STDP to the appropriate synapses.
//
// mustBeLocked: true (stateMutex must be held by caller)
func (n *Neuron) addRecentInputSpikeUnsafe(spike SpikeEvent) {
	// Skip if STDP is disabled
	if !n.stdpConfig.Enabled {
		return
	}

	// Add the spike to recent input history
	n.recentInputSpikes = append(n.recentInputSpikes, spike)

	// Clean old spikes to prevent memory growth
	n.recentInputSpikes = cleanOldSpikeHistory(n.recentInputSpikes, spike.Timestamp, n.stdpConfig.WindowSize)

	// Limit maximum history size for computational efficiency
	maxInputHistory := 200 // Allow more input history than pre-spike history
	if len(n.recentInputSpikes) > maxInputHistory {
		start := len(n.recentInputSpikes) - maxInputHistory
		n.recentInputSpikes = n.recentInputSpikes[start:]
	}
}

// applySTDPToAllRecentInputsUnsafe applies STDP learning to all recent input synapses
// Called when this neuron fires to strengthen/weaken incoming connections
//
// Biological process:
// When a neuron fires, it examines all recent input spikes and applies STDP
// to strengthen synapses that contributed to its firing (causal relationships)
// and weaken synapses that fired after it was already committed to firing
//
// mustBeLocked: true (stateMutex must be held by caller)
func (n *Neuron) applySTDPToAllRecentInputsUnsafe(postSpikeTime time.Time) {
	if !n.stdpConfig.Enabled {
		return
	}

	// Track STDP changes to apply to input gains
	stdpChanges := make(map[string]float64)

	// Process all recent input spikes for STDP
	for _, inputSpike := range n.recentInputSpikes {
		sourceID := inputSpike.SourceID
		if sourceID == "" {
			continue // Do not learn from anonymous triggers
		}

		// CORRECTED: Calculate time difference (pre - post) for STDP function
		// This ensures that causal timing (pre before post) gives negative values
		// which the STDP function correctly interprets as LTP
		timeDiff := inputSpike.Timestamp.Sub(postSpikeTime) // pre - post

		// NOTE: This is the opposite of the previous calculation:
		// OLD (wrong): timeDiff := postSpikeTime.Sub(inputSpike.Timestamp) // post - pre
		// NEW (correct): timeDiff := inputSpike.Timestamp.Sub(postSpikeTime) // pre - post
		//
		// With this correction:
		// - If input spike was 8ms BEFORE neuron firing: timeDiff = -8ms → LTP (positive)
		// - If input spike was 8ms AFTER neuron firing: timeDiff = +8ms → LTD (negative)

		// Calculate STDP change for this spike pair
		stdpChange := calculateSTDPWeightChange(timeDiff, n.stdpConfig)

		// Accumulate changes for this source
		stdpChanges[sourceID] += stdpChange
	}

	// Apply STDP changes to input gains
	if len(stdpChanges) > 0 {
		n.inputGainsMutex.Lock()
		for sourceID, change := range stdpChanges {
			if change != 0.0 {
				if _, exists := n.inputGains[sourceID]; !exists {
					n.inputGains[sourceID] = 1.0 // Initialize if not present
				}

				// Apply STDP change to the gain
				newGain := n.inputGains[sourceID] + change

				// Apply bounds (could be part of STDPConfig)
				minGain := 0.1
				maxGain := 5.0
				if newGain < minGain {
					newGain = minGain
				} else if newGain > maxGain {
					newGain = maxGain
				}

				n.inputGains[sourceID] = newGain
			}
		}
		n.inputGainsMutex.Unlock()
	}

	// Clean up old spikes
	n.recentInputSpikes = cleanOldSpikeHistory(n.recentInputSpikes, postSpikeTime, n.stdpConfig.WindowSize)
	n.lastSTDPUpdate = postSpikeTime
}

// recordOutputSpikeForSTDP records a spike on all output synapses for STDP
// Called when this neuron fires to update all its outgoing synapses
//
// mustBeLocked: true (stateMutex must be held by caller)
func (n *Neuron) recordOutputSpikeForSTDPUnsafe(spikeTime time.Time) {
	// Skip if STDP is disabled
	if !n.stdpConfig.Enabled {
		return
	}

	// Record this spike on all output synapses
	n.outputsMutex.RLock()
	for _, output := range n.outputs {
		output.recordPreSynapticSpike(spikeTime, n.stdpConfig)
	}
	n.outputsMutex.RUnlock()
}

// processIncomingSpikeForSTDP handles the STDP aspects of an incoming message
// This function extracts timing information and prepares for STDP learning
//
// Enhanced for synaptic scaling: Also registers input sources for scaling
//
// mustBeLocked: true (stateMutex must be held by caller)
func (n *Neuron) processIncomingSpikeForSTDPUnsafe(msg Message) {
	// Skip if STDP is disabled or message lacks a SourceID
	if !n.stdpConfig.Enabled || msg.SourceID == "" {
		return
	}

	// Create spike event for STDP tracking
	spikeEvent := SpikeEvent{
		SourceID:  msg.SourceID,
		Timestamp: msg.Timestamp,
		Value:     msg.Value,
		SynapseID: msg.SourceID, // For now, assume one synapse per source
	}

	// Add to recent input spikes for future STDP processing
	n.addRecentInputSpikeUnsafe(spikeEvent)

	// Register input source for synaptic scaling if not already tracked
	// This ensures all active input sources are considered during scaling
	n.registerInputSourceForScaling(msg.SourceID)
}

// registerInputSourceForScaling registers a new input source for synaptic scaling
// This method ensures that all active input sources have corresponding synaptic gains
//
// BIOLOGICAL CONTEXT:
// When a post-synaptic neuron receives input from a new source, it needs to
// establish receptor sensitivity (gain) for that synapse. Initially, the gain
// is set to 1.0 (normal sensitivity), but it will be adjusted by synaptic scaling
// to maintain optimal total input strength.
//
// This replaces the complex pointer sharing system with a simpler approach where
// each neuron manages its own synaptic gains independently.
func (n *Neuron) registerInputSourceForScaling(sourceID string) {
	// Check if scaling is enabled for this neuron
	if !n.scalingConfig.Enabled {
		return
	}

	// Check if this source is already registered
	n.inputGainsMutex.RLock()
	_, exists := n.inputGains[sourceID]
	n.inputGainsMutex.RUnlock()

	// If not registered, add with default gain of 1.0
	if !exists {
		n.inputGainsMutex.Lock()
		if n.inputGains == nil {
			n.inputGains = make(map[string]float64)
		}
		// Check again inside the lock to prevent race conditions
		if _, exists := n.inputGains[sourceID]; !exists {
			n.inputGains[sourceID] = 1.0 // Default receptor sensitivity
		}
		n.inputGainsMutex.Unlock()
	}
}

// Output represents a synaptic connection from this neuron to another component
// In biology: this models the synapse - the connection point between neurons
// Each synapse has unique characteristics that affect signal transmission
//
// Enhanced with Spike-Timing Dependent Plasticity (STDP) capabilities:
// STDP is a biological learning mechanism where synaptic strength changes based
// on the precise timing relationship between pre-synaptic and post-synaptic spikes.
// This enables synapses to learn temporal patterns and causal relationships.
type Output struct {
	channel chan Message // Communication channel to the target neuron/component
	// Models the synaptic cleft where neurotransmitters cross

	factor float64 // Current synaptic strength/weight (modified by STDP)
	// Models synaptic efficacy - how strong this connection is
	// This value changes over time based on spike timing relationships
	// Range typically bounded between minWeight and maxWeight

	delay time.Duration // Biological transmission delay
	// Models: axon conduction time, synaptic delay, etc.
	// Longer axons = longer delays (realistic brain timing)

	// === STDP LEARNING PARAMETERS ===
	// These control how the synapse learns from spike timing relationships

	baseWeight float64 // Original/reference synaptic weight
	// The initial weight value when synapse was created
	// Used as reference point for learning and weight bounds calculation

	minWeight float64 // Minimum allowed synaptic weight
	// Prevents synaptic weights from becoming too negative or being eliminated entirely
	// Biological basis: even "weak" synapses maintain some minimal efficacy
	// Typical value: 0.0 or small positive value

	maxWeight float64 // Maximum allowed synaptic weight
	// Prevents runaway synaptic strengthening that could destabilize the network
	// Biological basis: synapses have physical limits on their maximum strength
	// Typical value: 2-5 times the baseWeight

	learningRate float64 // STDP learning rate coefficient
	// Controls how quickly synaptic weights change in response to spike timing
	// Higher values = faster learning but potentially less stable
	// Lower values = slower learning but more stable
	// Typical biological range: 0.001 - 0.1

	// === STDP TIMING TRACKING ===
	// These fields track spike timing history needed for STDP calculations

	preSpikeTimes []time.Time // Recent pre-synaptic spike timestamps
	// Sliding window of recent spikes from the pre-synaptic neuron (this neuron)
	// Used to calculate timing differences when post-synaptic neuron fires
	// Window size typically covers ±50ms to capture all biologically relevant timing

	lastPreSpike time.Time // Timestamp of most recent pre-synaptic spike
	// Optimized access to the most recent pre-synaptic spike time
	// Used for fast STDP calculations without searching through preSpikeTimes array

	// === STDP CONFIGURATION ===
	// Parameters that control STDP behavior and timing windows

	stdpEnabled bool // Whether STDP learning is active for this synapse
	// Allows selective enabling/disabling of plasticity on individual synapses
	// Useful for creating fixed connections or temporary learning freezes

	stdpTimeConstant time.Duration // STDP exponential decay time constant
	// Controls the width of the STDP learning window
	// Biological value: typically 20ms for cortical synapses
	// Determines how quickly STDP effects decay with increasing time differences

	stdpWindowSize time.Duration // Maximum STDP timing window (±window)
	// Spikes separated by more than this time have no STDP effect
	// Computational optimization: avoids processing very old spikes
	// Biological range: typically ±20ms to ±50ms depending on synapse type
}

// STDPConfig represents configuration parameters for Spike-Timing Dependent Plasticity
// This structure encapsulates all the parameters needed to control how synapses
// learn from spike timing relationships in biologically realistic ways
//
// STDP Biological Background:
// In real brains, synaptic strength changes based on the precise timing of spikes.
// If a pre-synaptic spike arrives shortly before a post-synaptic spike (causally
// related), the synapse strengthens (Long-Term Potentiation, LTP). If the order
// is reversed, the synapse weakens (Long-Term Depression, LTD). This implements
// the principle "neurons that fire together, wire together" with temporal precision.
type STDPConfig struct {
	Enabled bool // Master switch for STDP learning
	// When false, synaptic weights remain fixed at their initial values
	// When true, weights adapt based on spike timing relationships
	// Useful for comparing learned vs. fixed network behavior

	LearningRate float64 // Rate of synaptic weight changes
	// Controls how quickly synapses adapt to spike timing patterns
	// Typical biological range: 0.001 - 0.1
	// Higher values = faster learning but potentially less stable
	// Lower values = slower learning but more robust to noise

	TimeConstant time.Duration // STDP exponential decay time constant
	// Controls the temporal precision of STDP learning windows
	// Biological typical value: 20ms for cortical synapses
	// Shorter constants = more precise timing requirements
	// Longer constants = more forgiving timing windows

	WindowSize time.Duration // Maximum timing difference for STDP effects
	// Spike pairs separated by more than this time have no effect
	// Computational optimization to avoid processing very old spikes
	// Biological range: ±20ms to ±50ms depending on synapse type
	// Should be 2-3 times the TimeConstant for proper decay

	MinWeight float64 // Minimum allowed synaptic weight
	// Prevents weights from becoming negative or zero (synaptic elimination)
	// Typical values: 0.0 (hard lower bound) or 0.1*baseWeight (soft bound)
	// Biological basis: even weak synapses maintain some minimal efficacy

	MaxWeight float64 // Maximum allowed synaptic weight
	// Prevents runaway synaptic strengthening that could destabilize networks
	// Typical values: 2-5 times the initial weight
	// Biological basis: synapses have physical saturation limits

	AsymmetryRatio float64 // Ratio of LTP to LTD strength (typically 1.0-2.0)
	// Controls the relative strength of potentiation vs. depression
	// Values > 1.0 favor strengthening over weakening
	// Values < 1.0 favor weakening over strengthening
	// Biological variation: different synapse types have different ratios
}

// SpikeEvent represents an incoming spike event for STDP processing
// This structure captures the essential information needed to apply
// spike-timing dependent plasticity when the post-synaptic neuron fires
//
// Biological context:
// When a post-synaptic neuron fires, it needs to "look back" at recent
// pre-synaptic spikes to determine how to modify synaptic weights.
// This structure stores the timing and source information for each
// recent input spike that could be relevant for STDP calculations.
type SpikeEvent struct {
	SourceID string // Identifier of the pre-synaptic neuron that sent this spike
	// Allows the post-synaptic neuron to identify which specific synapse
	// should have its weight modified based on the timing relationship
	// Essential for networks where neurons receive inputs from multiple sources

	Timestamp time.Time // When the pre-synaptic spike occurred
	// Precise timing is critical for STDP calculations
	// The time difference between this timestamp and the post-synaptic
	// firing time determines the magnitude and direction of weight changes

	Value float64 // Strength of the original spike signal
	// May be used to scale STDP effects based on signal strength
	// Some STDP models incorporate signal amplitude into learning rules
	// Allows for more nuanced learning beyond pure timing relationships

	SynapseID string // Identifier of the specific synapse that delivered this spike
	// Enables precise targeting of weight modifications to the correct synapse
	// A single pre-synaptic neuron may have multiple synapses to the same
	// post-synaptic neuron (though rare in our current model)
}

// FireEvent represents a real-time neuron firing event for visualization and monitoring
// This captures the exact moment when a biological neuron generates an action potential
// and provides essential information about the firing event for external observers
//
// Biological context:
// When a real neuron fires, it generates an action potential that propagates down
// its axon to all connected synapses. This event is instantaneous and discrete.
// Unlike traditional ANNs that have continuous activation values, biological
// neurons either fire (1) or don't fire (0) - this is the "all-or-nothing" principle.
//
// This struct models that discrete firing event and allows external systems
// (like visualizers, loggers, or learning algorithms) to observe when and how
// strongly each neuron fires without interfering with the neuron's operation.
type FireEvent struct {
	NeuronID string // Unique identifier of the neuron that fired
	// Allows tracking which specific neuron in a network generated this event

	Value float64 // Signal strength/amplitude of the firing event
	// Models the "strength" of the action potential
	// In biology: action potentials have standard amplitude, but this
	// can represent firing frequency or burst patterns

	Timestamp time.Time // Precise timing of when the firing occurred
	// Critical for studying temporal dynamics, spike-timing dependent
	// plasticity, and network synchronization patterns
}

// HomeostaticMetrics represents the homeostatic state and activity tracking for a neuron
// This models the biological mechanisms neurons use to monitor and regulate their own
// activity levels to maintain stable operation and prevent pathological states
//
// Biological context:
// Real neurons continuously monitor their activity levels through various mechanisms:
// - Calcium influx during action potentials serves as an activity sensor
// - Gene expression changes occur in response to chronic activity changes
// - Intrinsic excitability adjustments maintain optimal firing rates
// - Multiple timescales of regulation from seconds to hours/days
//
// This struct captures the key computational aspects of these biological processes
// while remaining computationally efficient for large-scale network simulation
type HomeostaticMetrics struct {
	// === ACTIVITY TRACKING ===
	// Models the biological mechanisms for monitoring recent neural activity

	firingHistory []time.Time // Recent firing timestamps within activity window
	// Models: calcium-dependent gene expression, activity-dependent signaling
	// Used to calculate recent firing rates for homeostatic adjustments

	activityWindow time.Duration // Time window for calculating firing rates
	// Models: the biological timescale over which neurons assess their activity
	// Typical biological values: seconds to minutes for homeostatic regulation
	// Shorter than structural plasticity (hours/days) but longer than STDP (ms)

	targetFiringRate float64 // Desired firing rate in Hz for optimal function
	// Models: the genetically determined "set point" for neural activity
	// Real neurons have intrinsic target activity levels that vary by cell type
	// Motor neurons: ~50-100 Hz, cortical neurons: ~1-10 Hz, etc.

	// === CALCIUM DYNAMICS ===
	// Models intracellular calcium as a biological activity sensor

	calciumLevel float64 // Current intracellular calcium concentration (arbitrary units)
	// Models: calcium influx through voltage-gated calcium channels during APs
	// Biological basis: calcium binds to calmodulin, activating kinases that
	// trigger gene expression changes and homeostatic adjustments

	calciumIncrement float64 // Amount of calcium added per action potential
	// Models: the calcium influx per spike, depends on channel density and AP amplitude
	// Larger increments = more sensitive homeostatic responses

	calciumDecayRate float64 // Rate of calcium removal/buffering per time step
	// Models: calcium pumps, buffers, and diffusion that clear calcium
	// Biological timescale: calcium clears over seconds to minutes

	// === HOMEOSTATIC ADJUSTMENT PARAMETERS ===
	// Controls for how the neuron adjusts its properties based on activity

	homeostasisStrength float64 // How aggressively to adjust threshold (0.0-1.0)
	// Models: the magnitude of homeostatic responses
	// Higher values = stronger regulation but potentially less stable
	// Lower values = gentler regulation but slower correction

	minThreshold float64 // Minimum allowed firing threshold
	// Prevents homeostatic adjustments from making neurons pathologically excitable
	// Models: biophysical limits on how low threshold can go

	maxThreshold float64 // Maximum allowed firing threshold
	// Prevents homeostatic adjustments from completely silencing neurons
	// Models: biophysical limits on how high threshold can be raised

	lastHomeostaticUpdate time.Time // Timestamp of last homeostatic adjustment
	// Used to control the timing of homeostatic updates
	// Models: the biological timescale of homeostatic processes (slower than synaptic)

	homeostaticInterval time.Duration // How often to perform homeostatic updates
	// Models: the characteristic timescale of homeostatic plasticity
	// Biological range: seconds to minutes for acute adjustments
	// Hours to days for gene expression-mediated changes (not modeled here)
}

// Neuron represents a single processing unit inspired by biological neurons
// Unlike traditional artificial neurons that perform instantaneous calculations,
// this neuron models the temporal dynamics of real neural processing:
// - Accumulates inputs over time (like dendrite integration)
// - Fires when threshold is reached (like action potential generation)
// - Sends outputs with realistic delays (like axon transmission)
// - Supports dynamic connectivity changes (like neuroplasticity)
// - Maintains stable activity through homeostatic regulation (like real neurons)
// - Scales synaptic sensitivity to maintain input balance (like receptor scaling)
type Neuron struct {
	// === IDENTIFICATION ===
	// Unique identifier for this neuron within a network
	id string // Neuron identifier for tracking and reference

	// === BIOLOGICAL ACTIVATION PARAMETERS ===
	// These model the electrical properties of real neuron membranes

	threshold float64 // Minimum accumulated charge needed to fire
	// Models the action potential threshold in real neurons
	// Typically around -55mV in biology
	// NOTE: This value can be adjusted by homeostatic plasticity

	baseThreshold float64 // Original threshold value before homeostatic adjustments
	// Stores the initial threshold for reference and bounds checking
	// Homeostatic adjustments modify 'threshold' but 'baseThreshold' remains constant

	fireFactor float64 // Global output multiplier when neuron fires
	// Models the amplitude of the action potential
	// In biology: action potentials have standard amplitude

	refractoryPeriod time.Duration // Absolute refractory period duration
	// Models the time after firing when neuron cannot fire again
	// In biology: Na+ channels are inactivated, preventing new action potentials
	// C. elegans neurons: typically 5-15ms depending on neuron type

	decayRate float64 // Membrane potential decay rate per time step
	// Models the leaky nature of biological neural membranes
	// In biology: membrane capacitance causes gradual charge dissipation
	// Value between 0.0-1.0: 0.95 = loses 5% charge per decay interval
	// Real neurons: membrane time constant typically 10-20ms

	// === HOMEOSTATIC PLASTICITY STATE ===
	// Models the biological mechanisms for activity monitoring and self-regulation

	homeostatic HomeostaticMetrics // All homeostatic plasticity state and parameters
	// Encapsulates the complex biological machinery for activity sensing,
	// threshold adjustment, and activity regulation that maintains network stability

	// === SPIKE-TIMING DEPENDENT PLASTICITY (STDP) STATE ===
	// Models synaptic learning based on precise spike timing relationships

	stdpConfig STDPConfig // Configuration parameters for STDP learning
	// Contains all the parameters that control how synapses learn from timing
	// Can be modified at runtime to enable/disable learning or adjust rates

	recentInputSpikes []SpikeEvent // Sliding window of recent input spikes
	// Tracks incoming spikes with their timing and source information
	// Used when this neuron fires to apply STDP to relevant synapses
	// Window size matches stdpConfig.WindowSize for computational efficiency

	lastSTDPUpdate time.Time // Timestamp of last STDP processing
	// Used to optimize STDP calculations by avoiding redundant processing
	// Also useful for debugging and monitoring learning dynamics

	// === SYNAPTIC SCALING STATE ===
	// Post-synaptic receptor sensitivity control for homeostatic balance

	inputGains map[string]float64 // Receptor sensitivity for each input source
	// Maps source neuron ID to synaptic gain (receptor sensitivity)
	// Models AMPA/NMDA receptor density scaling at post-synaptic sites
	// Key: source neuron ID, Value: gain multiplier (1.0 = normal sensitivity)
	// BIOLOGICAL: Post-synaptic neuron controls its own receptor sensitivity

	inputGainsMutex sync.RWMutex // Thread-safe access to input gains map
	// Protects concurrent access to inputGains during scaling and message processing
	// Read-write mutex allows multiple concurrent reads but exclusive writes
	// Essential for thread safety during synaptic scaling operations

	// === SYNAPTIC SCALING CONFIGURATION ===
	// Controls the homeostatic mechanism that maintains synaptic balance

	scalingConfig SynapticScalingConfig // All synaptic scaling parameters and state
	// Encapsulates the complete synaptic scaling system
	// Includes target strengths, rates, timing, and safety constraints
	// Modified through dedicated methods to ensure consistency

	// === ACTIVITY-BASED SCALING TRACKING ===
	// Models the biological activity sensing that drives synaptic scaling decisions

	inputActivityHistory map[string][]float64 // Recent input signal strengths per source
	// Maps source neuron ID to sliding window of recent effective signal strengths
	// Used to calculate actual average input strength for scaling decisions
	// Biological basis: neurons integrate recent synaptic activity over time windows

	inputActivityMutex sync.RWMutex // Thread-safe access to activity history
	// Protects concurrent access to inputActivityHistory during message processing
	// and scaling calculations

	activityTrackingWindow time.Duration // Time window for activity integration
	// How far back to look when calculating average input strengths
	// Biological timescale: 5-10 seconds (shorter than scaling interval)
	// Models the temporal integration window of calcium-dependent signaling

	minActivityForScaling float64 // Minimum activity level required to trigger scaling
	// Prevents scaling during periods of very low input activity
	// Biological basis: scaling only occurs when neurons are actively processing signals
	// Typical value: 10-20% of normal activity levels

	lastActivityCleanup time.Time // Timestamp of last activity history cleanup
	// Used to periodically remove old activity data to prevent memory growth
	// Cleanup occurs less frequently than activity tracking (every few minutes)

	// === COMMUNICATION INFRASTRUCTURE ===
	// Models the input/output structure of biological neurons

	input chan Message // Single input channel (models dendrite tree)
	// All inputs converge here, like dendrites
	// converging on the cell body (soma)

	outputs map[string]*Output // Dynamic set of output connections
	// Models the axon branching to multiple targets
	// String key allows named connections for management

	outputsMutex sync.RWMutex // Thread-safe access to outputs map
	// Allows safe connection modification during runtime
	// RWMutex permits multiple concurrent reads

	// === INTERNAL STATE (models neuron membrane properties) ===
	// These variables track the neuron's current electrical state

	accumulator float64 // Current sum of input charges within time window
	// Models the membrane potential in real neurons
	// Starts at resting potential, increases with excitation

	lastFireTime time.Time // Timestamp of most recent action potential
	// Models the refractory state timing in real neurons
	// Used to enforce refractory period constraints
	// Zero value indicates neuron has never fired

	stateMutex sync.Mutex // Protects internal state during message processing
	// Ensures atomic updates to accumulator, timing, and homeostatic state

	// === MONITORING AND OBSERVATION ===
	// Optional channel for reporting firing events to external observers
	fireEvents chan<- FireEvent // Optional fire event reporting channel
	// nil = disabled (default), non-nil = reports firing events
	// Used for visualization, learning algorithms, and analysis
}

// NewNeuron creates and initializes a new biologically-inspired neuron with homeostatic plasticity and STDP
// This factory function sets up all the necessary components for realistic neural processing
// with leaky integration, dynamic connectivity, refractory periods, homeostatic regulation,
// spike-timing dependent plasticity, and biologically accurate synaptic scaling
//
// The complete biological learning system enables:
// - Automatic activity monitoring through calcium-based sensing (homeostatic plasticity)
// - Self-regulation of firing threshold to maintain target activity levels
// - Synaptic learning based on precise spike timing relationships (STDP)
// - Post-synaptic receptor scaling for input balance (synaptic scaling)
// - Prevention of runaway excitation or neural silence
// - Temporal pattern recognition and causal relationship learning
// - Network stability without manual parameter tuning
//
// Parameters model key biological properties:
// id: unique identifier for this neuron (enables tracking in networks)
// threshold: electrical threshold for action potential generation (will be homeostatic base)
// decayRate: membrane potential decay factor per time step (0.0-1.0)
// refractoryPeriod: duration after firing when neuron cannot fire again
// fireFactor: action potential amplitude/strength
// targetFiringRate: desired firing rate in Hz for homeostatic regulation
// homeostasisStrength: how aggressively to adjust threshold (0.0-1.0)
// stdpConfig: configuration for spike-timing dependent plasticity learning
//
// Biological learning mechanisms:
// - Homeostatic: tracks firing history and adjusts threshold to maintain target rate
// - STDP: modifies synaptic weights based on pre/post-synaptic spike timing
// - Synaptic Scaling: adjusts post-synaptic receptor sensitivity to maintain input balance
// - Combined: creates stable yet adaptive networks that learn temporal patterns
func NewNeuron(id string, threshold float64, decayRate float64, refractoryPeriod time.Duration, fireFactor float64, targetFiringRate float64, homeostasisStrength float64, stdpConfig STDPConfig) *Neuron {
	// Calculate homeostatic bounds based on base threshold
	// Biological rationale: neurons can't adjust indefinitely - there are biophysical limits
	minThreshold := threshold * 0.1 // Can reduce to 10% of original (very excitable)
	maxThreshold := threshold * 5.0 // Can increase to 5x original (very quiet)

	// Set up homeostatic parameters with biologically reasonable defaults
	activityWindow := 5 * time.Second             // Track activity over 5 seconds
	calciumIncrement := 1.0                       // Arbitrary units of calcium per spike
	calciumDecayRate := 0.98                      // 2% calcium decay per millisecond
	homeostaticInterval := 100 * time.Millisecond // Check homeostasis every 100ms

	return &Neuron{
		id:               id,                       // Unique neuron identifier for network tracking
		threshold:        threshold,                // Current firing threshold (homeostatic)
		baseThreshold:    threshold,                // Original threshold (reference)
		decayRate:        decayRate,                // Membrane decay rate (biological: based on RC time constant)
		refractoryPeriod: refractoryPeriod,         // Refractory period (biological: ~5-15ms)
		fireFactor:       fireFactor,               // Output amplitude scaling
		input:            make(chan Message, 100),  // Buffered input channel
		outputs:          make(map[string]*Output), // Dynamic output connections
		fireEvents:       nil,                      // Optional fire event reporting (disabled by default)

		// Initialize homeostatic plasticity system
		homeostatic: HomeostaticMetrics{
			firingHistory:         make([]time.Time, 0, 100), // Pre-allocate for efficiency
			activityWindow:        activityWindow,
			targetFiringRate:      targetFiringRate,
			calciumLevel:          0.0, // Start with no calcium
			calciumIncrement:      calciumIncrement,
			calciumDecayRate:      calciumDecayRate,
			homeostasisStrength:   homeostasisStrength,
			minThreshold:          minThreshold,
			maxThreshold:          maxThreshold,
			lastHomeostaticUpdate: time.Now(),
			homeostaticInterval:   homeostaticInterval,
		},

		// Initialize STDP learning system
		stdpConfig:        stdpConfig,
		recentInputSpikes: make([]SpikeEvent, 0, 200), // Track recent input spikes

		// === NEW BIOLOGICALLY ACCURATE SYNAPTIC SCALING ===
		inputGains: make(map[string]float64), // Post-synaptic receptor sensitivity map
		scalingConfig: SynapticScalingConfig{
			Enabled:             false,                  // Disabled by default for backward compatibility
			TargetInputStrength: 1.0,                    // Moderate target strength
			ScalingRate:         0.001,                  // Conservative scaling rate
			ScalingInterval:     30 * time.Second,       // Scale every 30 seconds
			MinScalingFactor:    0.9,                    // Don't reduce more than 10% per step
			MaxScalingFactor:    1.1,                    // Don't increase more than 10% per step
			LastScalingUpdate:   time.Time{},            // Will be set when scaling starts
			ScalingHistory:      make([]float64, 0, 10), // Track recent scaling factors
		},

		lastSTDPUpdate: time.Now(), // Initialize STDP timing

		// accumulator starts at 0 (resting potential)
		// lastFireTime initialized to zero value (never fired)
	}
}

// NewSimpleNeuron creates a neuron with homeostatic plasticity and STDP disabled for backward compatibility
// This convenience function creates a neuron that behaves like the original implementation
// but with the learning infrastructure in place (just not active)
//
// Use this when you want the original temporal neuron behavior without self-regulation or learning,
// or when building networks that will implement learning through other mechanisms
//
// Parameters are the same as the original NewNeuron function:
// id: unique identifier for this neuron
// threshold: firing threshold (fixed, no homeostatic adjustment)
// decayRate: membrane potential decay factor per time step (0.0-1.0)
// refractoryPeriod: duration after firing when neuron cannot fire again
// fireFactor: action potential amplitude/strength
func NewSimpleNeuron(id string, threshold float64, decayRate float64, refractoryPeriod time.Duration, fireFactor float64) *Neuron {
	// Create disabled STDP configuration
	disabledSTDP := STDPConfig{
		Enabled:        false,
		LearningRate:   0.0,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.0,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.0,
	}

	// Create neuron with both homeostatic plasticity and STDP disabled
	return NewNeuron(
		id,
		threshold,
		decayRate,
		refractoryPeriod,
		fireFactor,
		0.0,          // targetFiringRate = 0 disables homeostatic regulation
		0.0,          // homeostasisStrength = 0 disables threshold adjustments
		disabledSTDP, // STDP disabled
	)
}

// NewNeuronWithLearning creates a neuron with both homeostatic and STDP learning enabled
// This convenience constructor sets up a neuron with biologically realistic learning
// parameters suitable for most applications
//
// Parameters:
// id: unique identifier for this neuron
// threshold: base firing threshold
// targetFiringRate: desired firing rate in Hz for homeostatic regulation
// stdpLearningRate: how quickly synapses adapt (typical: 0.01)
//
// Returns a neuron with:
// - Moderate homeostatic regulation (20% strength)
// - STDP learning with 20ms time constants
// - Reasonable weight bounds (0.1x to 3x base weight)
// - Biological timing windows (±50ms)
func NewNeuronWithLearning(id string, threshold float64, targetFiringRate float64, stdpLearningRate float64) *Neuron {
	// Standard biological parameters
	decayRate := 0.95
	refractoryPeriod := 10 * time.Millisecond
	fireFactor := 1.0
	homeostasisStrength := 0.2

	// STDP configuration with biological parameters
	stdpConfig := STDPConfig{
		Enabled:        true,
		LearningRate:   stdpLearningRate,
		TimeConstant:   20 * time.Millisecond, // Standard cortical value
		WindowSize:     50 * time.Millisecond, // ±50ms window
		MinWeight:      threshold * 0.1,       // 10% of base threshold
		MaxWeight:      threshold * 3.0,       // 3x base threshold
		AsymmetryRatio: 1.5,                   // Slight LTP bias
	}

	return NewNeuron(id, threshold, decayRate, refractoryPeriod, fireFactor,
		targetFiringRate, homeostasisStrength, stdpConfig)
}

// updateCalciumLevel applies calcium dynamics based on firing activity
// Models the biological process of calcium accumulation during action potentials
// and subsequent removal through pumps, buffers, and diffusion
//
// Biological process modeled:
// When a neuron fires an action potential, voltage-gated calcium channels open,
// allowing calcium influx. This calcium serves as an activity sensor that
// accumulates with repeated firing and slowly decays over time. The calcium
// level provides a running average of recent activity that drives homeostatic
// adjustments.
//
// mustBeLocked: true (stateMutex must be held by caller)
func (n *Neuron) updateCalciumLevelUnsafe() {
	// Skip calcium tracking if homeostasis is disabled
	if n.homeostatic.homeostasisStrength == 0.0 || n.homeostatic.targetFiringRate == 0.0 {
		return
	}

	// Apply exponential decay to calcium level
	// Models: calcium pumps, buffers, and diffusion removing calcium
	// Biological timescale: calcium clears over seconds to minutes
	n.homeostatic.calciumLevel *= n.homeostatic.calciumDecayRate

	// Set very small values to zero for computational efficiency
	// Prevents accumulation of floating-point precision errors
	if n.homeostatic.calciumLevel < 1e-10 {
		n.homeostatic.calciumLevel = 0.0
	}
}

// addCalciumFromFiring increases calcium level due to action potential firing
// Models the calcium influx that occurs during action potential generation
// This calcium accumulation serves as the activity sensor for homeostatic regulation
//
// Biological context:
// Action potentials cause voltage-gated calcium channels to open, leading to
// calcium influx. The amount of calcium depends on:
// - Channel density and distribution
// - Action potential amplitude and duration
// - Extracellular calcium concentration
// - Cell volume and buffering capacity
//
// mustBeLocked: true (stateMutex must be held by caller)
func (n *Neuron) addCalciumFromFiringUnsafe() {
	// Skip calcium tracking if homeostasis is disabled
	if n.homeostatic.homeostasisStrength == 0.0 || n.homeostatic.targetFiringRate == 0.0 {
		return
	}

	// Add calcium increment for this firing event
	// Models: calcium influx through voltage-gated calcium channels
	n.homeostatic.calciumLevel += n.homeostatic.calciumIncrement

	// Note: In real neurons, excessive calcium can be toxic, but for our
	// model we rely on the decay process to prevent unlimited accumulation
}

// updateFiringHistory maintains a sliding window of recent firing times
// This provides the temporal data needed to calculate firing rates for
// homeostatic regulation
//
// Biological context:
// Neurons don't explicitly track firing times, but the calcium-dependent
// signaling cascades effectively integrate recent activity. This sliding
// window approach approximates that biological integration process.
//
// mustBeLocked: true (stateMutex must be held by caller)
func (n *Neuron) updateFiringHistoryUnsafe(firingTime time.Time) {
	// Skip firing history tracking if homeostasis is disabled
	if n.homeostatic.homeostasisStrength == 0.0 || n.homeostatic.targetFiringRate == 0.0 {
		return
	}

	// Add this firing time to history
	n.homeostatic.firingHistory = append(n.homeostatic.firingHistory, firingTime)

	// Remove old firing times outside the activity window
	// This maintains a sliding window of recent activity
	cutoffTime := firingTime.Add(-n.homeostatic.activityWindow)

	// Find the first firing time within the window
	validStart := 0
	for i, t := range n.homeostatic.firingHistory {
		if t.After(cutoffTime) {
			validStart = i
			break
		}
	}

	// Keep only recent firing times (efficient slice operation)
	if validStart > 0 {
		// Create new slice with only valid times to prevent memory leaks
		newHistory := make([]time.Time, len(n.homeostatic.firingHistory)-validStart)
		copy(newHistory, n.homeostatic.firingHistory[validStart:])
		n.homeostatic.firingHistory = newHistory
	}
}

// calculateCurrentFiringRate computes the current firing rate from recent history
// Returns the firing rate in Hz based on the sliding window of recent spikes
//
// Biological context:
// This approximates how biological calcium-dependent signaling cascades
// effectively compute a running average of recent neural activity. The
// firing rate serves as the signal for homeostatic regulation.
//
// mustBeLocked: true (stateMutex must be held by caller)
func (n *Neuron) calculateCurrentFiringRateUnsafe() float64 {
	now := time.Now()
	cutoffTime := now.Add(-n.homeostatic.activityWindow)

	// Count spikes within the activity window
	spikesInWindow := 0
	for _, t := range n.homeostatic.firingHistory {
		if t.After(cutoffTime) {
			spikesInWindow++
		}
	}

	// Calculate firing rate in Hz
	// Convert activity window to seconds for rate calculation
	windowSeconds := n.homeostatic.activityWindow.Seconds()
	if windowSeconds > 0 {
		return float64(spikesInWindow) / windowSeconds
	}
	return 0.0
}

// performHomeostaticAdjustment adjusts the firing threshold based on recent activity
// This is the core homeostatic mechanism that maintains stable neural activity
//
// Biological process modeled:
// Neurons use calcium-dependent signaling to detect when their activity deviates
// from optimal levels and adjust their intrinsic excitability accordingly:
// - High activity (high calcium) → increase threshold → reduce excitability
// - Low activity (low calcium) → decrease threshold → increase excitability
//
// # This creates a negative feedback loop that stabilizes network activity
//
// mustBeLocked: true (stateMutex must be held by caller)
func (n *Neuron) performHomeostaticAdjustmentUnsafe() {
	// Skip if homeostatic plasticity is disabled
	if n.homeostatic.homeostasisStrength == 0.0 || n.homeostatic.targetFiringRate == 0.0 {
		return
	}

	// Calculate current firing rate
	currentRate := n.calculateCurrentFiringRateUnsafe()

	// Calculate the error between current and target firing rates
	rateError := currentRate - n.homeostatic.targetFiringRate

	// Calculate threshold adjustment based on activity error
	// Positive error (too active) → increase threshold (reduce excitability)
	// Negative error (too quiet) → decrease threshold (increase excitability)
	thresholdAdjustment := rateError * n.homeostatic.homeostasisStrength * n.baseThreshold * 0.01

	// Apply the adjustment
	newThreshold := n.threshold + thresholdAdjustment

	// Enforce biological bounds on threshold adjustment
	// Neurons can't adjust their threshold indefinitely
	if newThreshold < n.homeostatic.minThreshold {
		newThreshold = n.homeostatic.minThreshold
	} else if newThreshold > n.homeostatic.maxThreshold {
		newThreshold = n.homeostatic.maxThreshold
	}

	// Update the threshold
	n.threshold = newThreshold

	// Update timestamp of last homeostatic adjustment
	n.homeostatic.lastHomeostaticUpdate = time.Now()
}

// shouldPerformHomeostaticUpdate checks if it's time for homeostatic regulation
// Homeostatic processes operate on slower timescales than synaptic transmission
//
// Biological rationale:
// Homeostatic adjustments occur over seconds to minutes, much slower than
// the millisecond timescales of synaptic transmission and action potentials.
// This separation of timescales is crucial for stability.
//
// mustBeLocked: true (stateMutex must be held by caller)
func (n *Neuron) shouldPerformHomeostaticUpdateUnsafe() bool {
	return time.Since(n.homeostatic.lastHomeostaticUpdate) >= n.homeostatic.homeostaticInterval
}

// SetFireEventChannel configures optional real-time firing event reporting
// This method enables external monitoring of neuron firing events without
// interfering with the neuron's core computational processes
//
// Biological inspiration:
// In neuroscience research, scientists often need to monitor when individual
// neurons fire to understand network dynamics, learning, and information processing.
// This is typically done using techniques like:
// - Microelectrodes that detect action potentials
// - Calcium imaging that shows neural activity
// - Multi-electrode arrays that monitor many neurons simultaneously
//
// This method provides a similar capability for artificial neural networks,
// allowing researchers to observe firing patterns, study network dynamics,
// and implement biologically-inspired learning algorithms.
//
// Usage patterns:
// - Visualization: Real-time display of network activity
// - Learning algorithms: Spike-timing dependent plasticity (STDP)
// - Analysis: Network synchronization and oscillation studies
// - Debugging: Identifying silent or hyperactive neurons
// - Homeostatic monitoring: Observing self-regulation in action
//
// Performance considerations:
// - The channel is used in a non-blocking manner to prevent interference
// - Events are sent asynchronously to avoid disrupting neural computation
// - If the channel becomes full, events are dropped rather than blocking
//
// ch: Channel to receive FireEvent notifications when this neuron fires
//
//	Set to nil to disable fire event reporting (default state)
//	The channel should be buffered to handle burst firing patterns
func (n *Neuron) SetFireEventChannel(ch chan<- FireEvent) {
	n.stateMutex.Lock()         // Protect concurrent access to fire event channel
	defer n.stateMutex.Unlock() // Ensure lock is always released

	n.fireEvents = ch

	// Biological analogy: This is like placing a recording electrode near
	// a neuron to monitor its electrical activity. The electrode doesn't
	// interfere with the neuron's function - it just observes and reports
	// when action potentials occur.
}

// AddOutput safely adds a new synaptic connection to this neuron with STDP learning capabilities
// This models neuroplasticity - the brain's ability to form new connections
// throughout life. In developing brains, neurons constantly grow new synapses.
// In adult brains, learning involves creating and strengthening connections.
//
// Enhanced with STDP: The synapse will learn from spike timing relationships
// if STDP is enabled for this neuron, allowing the connection to strengthen
// or weaken based on causal relationships between pre and post-synaptic firing.
//
// Enhanced with Synaptic Scaling: If the target neuron has scaling enabled,
// it will automatically register this connection and adjust its receptor
// sensitivity to maintain optimal input balance.
//
// Biological context:
// - Dendritic growth: neurons extend dendrites to reach new partners
// - Axon sprouting: axons grow new branches to contact more targets
// - Synaptogenesis: formation of new synaptic contacts
// - Experience-dependent plasticity: activity drives connection formation
// - STDP learning: synapses adapt based on timing relationships
// - Receptor scaling: post-synaptic sensitivity adjustments
//
// Parameters:
// id: unique identifier for this connection (allows later modification/removal)
// channel: destination for signals (the target neuron's input)
// factor: initial synaptic strength/weight (will be modified by STDP if enabled)
// delay: transmission delay (models axon length and conduction velocity)
// targetNeuron: optional - if provided, enables synaptic scaling registration (can be nil)
func (n *Neuron) AddOutput(id string, channel chan Message, factor float64, delay time.Duration, targetNeuron ...*Neuron) {
	n.outputsMutex.Lock()         // Acquire exclusive write access
	defer n.outputsMutex.Unlock() // Ensure lock is always released

	// Create new synaptic connection with STDP capabilities
	output := &Output{
		channel: channel, // Communication pathway
		factor:  factor,  // Current synaptic strength
		delay:   delay,   // Conduction delay

		// STDP learning parameters
		baseWeight:       factor,                    // Reference weight for learning
		minWeight:        n.stdpConfig.MinWeight,    // Use neuron's STDP config
		maxWeight:        n.stdpConfig.MaxWeight,    // Use neuron's STDP config
		learningRate:     n.stdpConfig.LearningRate, // Use neuron's STDP config
		preSpikeTimes:    make([]time.Time, 0, 50),  // Pre-allocate spike history
		lastPreSpike:     time.Time{},               // Initialize to zero
		stdpEnabled:      n.stdpConfig.Enabled,      // Inherit from neuron
		stdpTimeConstant: n.stdpConfig.TimeConstant, // Copy timing parameters
		stdpWindowSize:   n.stdpConfig.WindowSize,   // Copy window size
	}

	// If STDP config specifies relative weight bounds, calculate them
	if n.stdpConfig.MinWeight == 0 && n.stdpConfig.MaxWeight == 0 {
		// Auto-calculate bounds based on base weight
		output.minWeight = factor * 0.1 // 10% of base weight
		output.maxWeight = factor * 3.0 // 300% of base weight
	}

	n.outputs[id] = output

	// === BIOLOGICALLY ACCURATE SCALING REGISTRATION ===
	// If target neuron provided, it will automatically register this input source
	// when it receives the first message from this neuron. This models how
	// post-synaptic neurons detect new input sources and establish receptor sensitivity.

	// Biological analogy: This represents the completion of synaptogenesis
	// where a new functional synaptic connection becomes available for
	// neural communication, information processing, and learning
}

// AddOutputWithSTDP safely adds a new synaptic connection with custom STDP parameters
// This allows fine-grained control over individual synapse learning properties
// while maintaining the neuron's overall STDP configuration
//
// Use this method when you need synapses with different learning characteristics
// within the same neuron (e.g., different learning rates for different input types)
//
// Parameters:
// id: unique identifier for this connection
// channel: destination for signals
// factor: initial synaptic strength/weight
// delay: transmission delay
// customSTDP: custom STDP configuration for this specific synapse
// targetNeuron: optional - if provided, enables synaptic scaling (can be nil)
func (n *Neuron) AddOutputWithSTDP(id string, channel chan Message, factor float64, delay time.Duration, customSTDP STDPConfig, targetNeuron ...*Neuron) {
	n.outputsMutex.Lock()
	defer n.outputsMutex.Unlock()

	// Create synapse with custom STDP parameters
	output := &Output{
		channel:          channel,
		factor:           factor,
		delay:            delay,
		baseWeight:       factor,
		minWeight:        customSTDP.MinWeight,
		maxWeight:        customSTDP.MaxWeight,
		learningRate:     customSTDP.LearningRate,
		preSpikeTimes:    make([]time.Time, 0, 50),
		lastPreSpike:     time.Time{},
		stdpEnabled:      customSTDP.Enabled,
		stdpTimeConstant: customSTDP.TimeConstant,
		stdpWindowSize:   customSTDP.WindowSize,
	}

	// Auto-calculate bounds if not specified
	if customSTDP.MinWeight == 0 && customSTDP.MaxWeight == 0 {
		output.minWeight = factor * 0.1
		output.maxWeight = factor * 3.0
	}

	n.outputs[id] = output

	// Target neuron will automatically register this input source when it receives
	// the first message from this neuron (if scaling is enabled)
}

// GetInputChannel returns the bidirectional input channel for direct connections
// This is used for neuron-to-neuron connections where we need to pass the channel
// to AddOutput methods
func (n *Neuron) GetInputChannel() chan Message {
	return n.input
}

// RemoveOutput safely removes a synaptic connection
// Models synaptic pruning - the brain's process of eliminating unnecessary
// or ineffective connections to optimize neural circuits
//
// Biological context:
// - Developmental pruning: elimination of excess connections during development
// - Activity-dependent pruning: "use it or lose it" - unused synapses are removed
// - Synaptic scaling: global adjustment of synaptic strengths
// - Pathological pruning: excessive pruning in some neurological conditions
//
// This is crucial for:
// - Network optimization: removing redundant connections
// - Learning: eliminating interfering or outdated associations
// - Memory consolidation: strengthening important connections while removing others
//
// id: unique identifier of the connection to remove
func (n *Neuron) RemoveOutput(id string) {
	n.outputsMutex.Lock()         // Acquire exclusive write access
	defer n.outputsMutex.Unlock() // Ensure lock is always released

	delete(n.outputs, id)

	// Biological analogy: This represents the completion of synaptic elimination
	// where the physical synaptic structure is dismantled and the connection
	// is no longer available for neural communication
}

// GetOutputCount returns the current number of synaptic connections
// Thread-safe read operation that allows monitoring network connectivity
// In biological terms: this tells us the neuron's "fan-out" or how many
// other neurons this neuron can directly influence
func (n *Neuron) GetOutputCount() int {
	n.outputsMutex.RLock()         // Acquire shared read access (allows concurrent reads)
	defer n.outputsMutex.RUnlock() // Ensure lock is released

	return len(n.outputs)
}

// Run starts the main neuron processing loop with continuous leaky integration, homeostatic regulation, and synaptic scaling
// This implements the core neural computation cycle that runs continuously with
// biologically realistic membrane dynamics, self-regulation, and synaptic balance maintenance:
// 1. Wait for input signals, decay timer events, scaling timer events, or shutdown signals
// 2. Apply continuous membrane potential decay (leaky integration)
// 3. Apply calcium decay (homeostatic activity sensing)
// 4. Integrate incoming signals with existing accumulated charge
// 5. Fire when threshold conditions are met during refractory-compliant periods
// 6. Update homeostatic state (calcium, firing history, threshold adjustment)
// 7. Apply synaptic scaling to maintain input balance (slowest timescale)
// 8. Reset and repeat
//
// MUST be called as a goroutine: go neuron.Run()
// This allows the neuron to operate independently and concurrently with
// other neurons, modeling the parallel nature of biological neural networks
//
// Biological processes modeled with multi-timescale regulation:
// - Continuous membrane potential decay (models membrane capacitance/resistance)
// - Calcium dynamics for activity sensing (models intracellular calcium signaling)
// - Asynchronous signal integration (models dendritic summation)
// - Refractory period enforcement (models Na+ channel recovery)
// - Homeostatic threshold adjustment (models intrinsic plasticity - seconds to minutes)
// - Synaptic scaling for input balance (models synaptic homeostasis - minutes to hours)
// - Real-time temporal dynamics (no artificial time windows)
//
// MULTI-TIMESCALE BIOLOGICAL REALISM:
// - Membrane dynamics: 1ms (fastest - electrical properties)
// - STDP learning: milliseconds to seconds (synaptic plasticity)
// - Homeostatic plasticity: seconds to minutes (intrinsic regulation)
// - Synaptic scaling: minutes to hours (synaptic homeostasis - slowest)
func (n *Neuron) Run() {
	// Create decay timer for continuous membrane potential decay
	// Models the biological membrane time constant (RC circuit behavior)
	// Decay interval of 1ms provides good temporal resolution for C. elegans scale
	decayInterval := 1 * time.Millisecond
	decayTicker := time.NewTicker(decayInterval)
	defer decayTicker.Stop()

	// Create scaling timer for synaptic homeostasis (much slower than membrane dynamics)
	// Models the biological timescale of synaptic scaling (minutes to hours)
	// Check interval of 1 second allows responsive scaling without excessive computation
	scalingCheckInterval := 1 * time.Second
	scalingTicker := time.NewTicker(scalingCheckInterval)
	defer scalingTicker.Stop()

	// Main event loop - the neuron's "life cycle" with multi-timescale biological dynamics
	// Processes events in order of biological priority and timescale:
	for {
		select {
		// Event 1: New input signal received (excitatory or inhibitory)
		// Models: synaptic transmission, neurotransmitter binding,
		//         postsynaptic potential generation
		// Highest priority - immediate processing like real synaptic events
		// Timescale: sub-millisecond (fastest biological process)
		case msg, ok := <-n.input:
			if !ok {
				return // Channel closed, exit goroutine
			}
			n.processMessageWithDecay(msg)

		// Event 2: Membrane potential and calcium decay timer (continuous biological processes)
		// Models: membrane capacitance discharge, ion channel leakage,
		//         calcium removal, return toward resting potential
		// Regular biological processes that occur continuously
		// Timescale: 1ms intervals (membrane electrical dynamics)
		case <-decayTicker.C:
			n.stateMutex.Lock()

			// Apply membrane potential decay (fastest process)
			n.applyMembraneDecayUnsafe()

			// Apply calcium decay for homeostatic sensing (medium timescale)
			n.updateCalciumLevelUnsafe()

			// Check if it's time for homeostatic adjustment (medium timescale)
			// Operates on seconds to minutes - much slower than membrane dynamics
			if n.shouldPerformHomeostaticUpdateUnsafe() {
				n.performHomeostaticAdjustmentUnsafe()
			}

			n.stateMutex.Unlock()

		// Event 3: Synaptic scaling timer (slowest biological process)
		// Models: synaptic homeostasis, input strength balance maintenance
		// Operates on minutes to hours - slowest regulatory mechanism
		// Timescale: 1s check interval, actual scaling every 30s-10min depending on configuration
		case <-scalingTicker.C:
			// Apply synaptic scaling to maintain stable input strength
			// This is the slowest homeostatic mechanism, operating on the longest timescale
			// Preserves learned STDP patterns while maintaining overall synaptic balance
			n.applySynapticScaling()

			// Biological analogy: This represents the ongoing synaptic homeostasis
			// that prevents runaway strengthening or weakening while preserving
			// the relative patterns learned through experience and STDP
		}
	}
}

// applyMembraneDecay applies continuous membrane potential decay
// Models the biological process of charge dissipation through membrane resistance
// This replaces the artificial "time window reset" with realistic exponential decay
//
// Biological process modeled:
// In real neurons, the cell membrane acts like a leaky capacitor (RC circuit).
// Charge continuously leaks out through membrane resistance, causing the
// membrane potential to decay exponentially toward resting potential.
// This creates natural temporal summation where recent inputs have stronger
// influence than older inputs.
//
// Mathematical model:
// V(t) = V(0) * e^(-t/τ) where τ is the membrane time constant
// Discrete approximation: V(t+dt) = V(t) * decayRate
//
// mustBeLocked: true (stateMutex must be held by caller)
func (n *Neuron) applyMembraneDecayUnsafe() {
	// Apply exponential decay to accumulated membrane potential
	// Models: membrane resistance causing continuous charge leakage
	// decayRate < 1.0 causes gradual approach to resting potential (0)
	n.accumulator *= n.decayRate

	// In biology: membrane potential asymptotically approaches resting potential
	// For computational efficiency, set very small values to exactly zero
	// This prevents accumulation of floating-point precision errors
	if n.accumulator < 1e-10 && n.accumulator > -1e-10 {
		n.accumulator = 0.0
	}

	// Note: Unlike the original implementation, we never completely reset
	// the accumulator to zero. This models the continuous nature of
	// biological membrane dynamics where there are no discrete "resets"
}

// Legacy method for backward compatibility - calls the unsafe version with proper locking
func (n *Neuron) applyMembraneDecay() {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	n.applyMembraneDecayUnsafe()
}

// processMessageWithDecay handles incoming synaptic signals with continuous leaky integration
// Enhanced with homeostatic state tracking and STDP spike timing recording
// Now includes biologically accurate synaptic scaling through post-synaptic receptor sensitivity
// **ENHANCED: Now tracks input activity for biologically accurate scaling decisions**
//
// Biological process modeled:
// 1. Synaptic signal arrives at dendrite (postsynaptic potential)
// 2. Record spike timing and source for STDP learning
// 3. Apply post-synaptic receptor gain (synaptic scaling)
// 4. **NEW: Track effective input strength for scaling algorithm**
// 5. Signal adds to current membrane potential (no time window constraints)
// 6. Continuous decay is handled separately by applyMembraneDecay()
// 7. If accumulated potential reaches threshold, action potential is triggered
// 8. Apply STDP learning to recent input spikes
// 9. Update homeostatic state (calcium, firing history)
// 10. Refractory period constraints are enforced during firing attempts
func (n *Neuron) processMessageWithDecay(msg Message) {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// Process incoming spike for STDP learning
	// Record timing and source information for future weight updates
	n.processIncomingSpikeForSTDPUnsafe(msg)

	// === BIOLOGICALLY ACCURATE SYNAPTIC SCALING ===
	// Apply post-synaptic receptor gain to incoming signal
	// This models how the post-synaptic neuron controls its own sensitivity
	// to different input sources through receptor density regulation
	finalSignalValue := n.applyPostSynapticGainUnsafe(msg)

	// === BIOLOGICAL ACTIVITY TRACKING FOR SCALING ===
	// Record the effective input strength for scaling algorithm
	// This models how neurons monitor their actual synaptic input patterns
	// over time to detect when scaling should occur
	if n.scalingConfig.Enabled && msg.SourceID != "" {
		n.recordInputActivityUnsafe(msg.SourceID, finalSignalValue)
	}

	// Directly add scaled signal to current membrane potential
	// Models: postsynaptic potential summing with existing membrane charge
	// No time window checks - integration is continuous like real neurons
	n.accumulator += finalSignalValue

	// Check if accumulated charge has reached the firing threshold
	// Models: action potential initiation at the axon hillock
	// Refractory period enforcement is handled within fireUnsafe()
	if n.accumulator >= n.threshold {
		n.fireUnsafe()             // Generate action potential (includes refractory check)
		n.resetAccumulatorUnsafe() // Return to resting potential after firing
	}
}

// recordInputActivityUnsafe tracks effective input signal strength for biological scaling
// This models how post-synaptic neurons monitor their actual synaptic input patterns
// over time to detect activity imbalances that should trigger receptor scaling
//
// BIOLOGICAL PROCESS MODELED:
// In real neurons, synaptic scaling is triggered by detecting changes in overall
// synaptic drive. Neurons integrate synaptic activity over time windows (seconds to
// minutes) to assess whether their total input strength has shifted away from
// optimal levels. This function captures that biological activity monitoring.
//
// CALCIUM-DEPENDENT SIGNALING:
// Each effective synaptic input contributes to intracellular calcium signaling
// cascades that ultimately drive scaling decisions. By tracking the actual
// effective signal strengths, we model the biological activity sensor that
// determines when receptor density should be adjusted.
//
// Parameters:
// sourceID: identifier of the input source neuron
// effectiveSignalValue: final signal strength (pre-weight × post-gain)
//
// mustBeLocked: true (stateMutex must be held by caller)
func (n *Neuron) recordInputActivityUnsafe(sourceID string, effectiveSignalValue float64) {
	// Initialize activity tracking structures if needed
	if n.inputActivityHistory == nil {
		n.inputActivityHistory = make(map[string][]float64)
	}

	// Get current time for activity timestamping
	now := time.Now()

	// === RECORD NEW ACTIVITY ===
	// Add this signal to the activity history for this source
	// Models: accumulation of synaptic activity over biological time windows
	n.inputActivityMutex.Lock()
	n.inputActivityHistory[sourceID] = append(n.inputActivityHistory[sourceID],
		math.Abs(effectiveSignalValue)) // Use absolute value for activity strength
	n.inputActivityMutex.Unlock()

	// === PERIODIC CLEANUP (BIOLOGICAL FORGETTING) ===
	// Clean old activity data periodically to model biological forgetting
	// and prevent unlimited memory growth
	if now.Sub(n.lastActivityCleanup) > n.activityTrackingWindow {
		n.cleanOldActivityHistoryUnsafe(now)
		n.lastActivityCleanup = now
	}
}

// cleanOldActivityHistoryUnsafe removes activity data outside the biological integration window
// This models the natural decay of activity-dependent signaling in real neurons
//
// BIOLOGICAL RATIONALE:
// Neurons don't maintain indefinite memory of past activity. The calcium-dependent
// signaling cascades that drive scaling decisions integrate activity over finite
// time windows (typically 5-10 seconds for immediate scaling decisions, longer
// for developmental scaling). This cleanup models that biological forgetting.
//
// mustBeLocked: true (stateMutex must be held by caller)
func (n *Neuron) cleanOldActivityHistoryUnsafe(currentTime time.Time) {
	n.inputActivityMutex.Lock()
	defer n.inputActivityMutex.Unlock()

	// For each input source, limit activity history size
	// This is a simplified cleanup - in full biological accuracy, we'd
	// timestamp each activity entry and remove based on actual time
	maxHistorySize := 100 // Reasonable limit for biological integration window

	for sourceID, activities := range n.inputActivityHistory {
		if len(activities) > maxHistorySize {
			// Keep only the most recent activities (biological recency bias)
			start := len(activities) - maxHistorySize
			n.inputActivityHistory[sourceID] = activities[start:]
		}
	}
}

// applyPostSynapticGainUnsafe applies receptor sensitivity scaling to incoming signals
// This is the core of biologically accurate synaptic scaling - the post-synaptic
// neuron controls its own sensitivity to different input sources
//
// BIOLOGICAL PROCESS MODELED:
// In real neurons, synaptic scaling occurs through changes in post-synaptic
// receptor density (AMPA, NMDA receptors). The pre-synaptic neuron releases
// the same amount of neurotransmitter, but the post-synaptic response changes
// based on receptor availability. This allows independent scaling control.
//
// ADVANTAGES OF THIS APPROACH:
// - Post-synaptic control (biologically accurate)
// - No cross-neuron communication required
// - Independent receptor sensitivity per input source
// - Preserves pre-synaptic weight learning (STDP)
// - Thread-safe (each neuron manages own data)
//
// mustBeLocked: true (stateMutex must be held by caller)
func (n *Neuron) applyPostSynapticGainUnsafe(msg Message) float64 {
	// If scaling is disabled or no source ID, use original signal
	if !n.scalingConfig.Enabled || msg.SourceID == "" {
		return msg.Value
	}

	// Get the receptor gain for this input source
	n.inputGainsMutex.RLock()
	gain, exists := n.inputGains[msg.SourceID]
	n.inputGainsMutex.RUnlock()

	// If source not yet registered, register it with default gain
	if !exists {
		gain = 1.0 // Default receptor sensitivity
		n.registerInputSourceForScaling(msg.SourceID)
	}

	// Apply receptor gain to the signal
	// Final signal = pre-synaptic_strength × post-synaptic_receptor_sensitivity
	return msg.Value * gain
}

// fire triggers action potential propagation to all connected neurons
// Models the all-or-nothing action potential that travels down the axon
// and triggers neurotransmitter release at all synaptic terminals
// With refractory period enforcement for biological realism
//
// Biological process:
// 1. Verify neuron is not in refractory period
// 2. Action potential initiated at axon hillock
// 3. Electrical signal propagates down main axon
// 4. Signal reaches all axon terminals simultaneously
// 5. Triggers neurotransmitter release at each synapse
// 6. Each synapse may have different strength/delay characteristics
func (n *Neuron) fire() {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// Use the internal unsafe method which includes refractory period checking
	n.fireUnsafe()
}

// fireUnsafe is the internal firing method called when state lock is already held
// Enhanced with homeostatic state updates and STDP spike timing recording
// Includes refractory period enforcement and biological timing constraints
//
// Biological process modeled:
// 1. Check if neuron is in refractory period (cannot fire if recent firing occurred)
// 2. If firing is allowed, generate action potential
// 3. Record firing time to enforce future refractory periods
// 4. Update homeostatic state (calcium accumulation, firing history)
// 5. Record spike timing on all output synapses for STDP learning
// 6. Propagate signal to all synaptic connections with timing information
//
// The refractory period models the biological reality that after an action potential,
// voltage-gated sodium channels become inactivated and require time to recover.
// During this period, no amount of input can trigger another action potential.
//
// Homeostatic updates model the calcium influx and activity tracking that real
// neurons use to monitor and regulate their own excitability levels.
//
// STDP integration ensures that all output synapses record this spike timing
// for future learning when their target neurons fire.
func (n *Neuron) fireUnsafe() {
	// Check refractory period constraint
	// Models: voltage-gated Na+ channel inactivation state
	now := time.Now()
	if !n.lastFireTime.IsZero() && now.Sub(n.lastFireTime) < n.refractoryPeriod {
		// Neuron is in refractory period - firing is physically impossible
		// In biology: Na+ channels are inactivated, K+ channels may still be open
		// This prevents unrealistic rapid-fire bursts that don't occur in real neurons
		return
	}

	// Record this firing event for future refractory period enforcement
	// Models: the moment when Na+ channels become inactivated
	n.lastFireTime = now

	// === HOMEOSTATIC UPDATES ===
	// Update homeostatic state to track this firing event

	// Add calcium from this action potential
	// Models: calcium influx through voltage-gated calcium channels
	n.addCalciumFromFiringUnsafe()

	// Update firing history for rate calculation
	// Models: activity-dependent signaling cascades
	n.updateFiringHistoryUnsafe(now)

	// === STDP UPDATES ===
	// Record this spike on all output synapses for STDP learning
	n.recordOutputSpikeForSTDPUnsafe(now)

	// Apply STDP learning to all recent input synapses
	n.applySTDPToAllRecentInputsUnsafe(now)

	// Calculate output signal strength
	outputValue := n.accumulator * n.fireFactor

	// Report firing event if channel is set
	if n.fireEvents != nil {
		select {
		case n.fireEvents <- FireEvent{
			NeuronID:  n.id,
			Value:     outputValue,
			Timestamp: now, // Use the same timestamp for consistency
		}:
		default: // Don't block if channel is full
		}
	}

	// Get snapshot of outputs (minimal locking since we're already protected)
	n.outputsMutex.RLock()
	outputsCopy := make(map[string]*Output, len(n.outputs))
	for id, output := range n.outputs {
		outputsCopy[id] = output
	}
	n.outputsMutex.RUnlock()

	// Parallel transmission to all outputs with STDP timing information
	// Models: action potential propagating simultaneously down all axon branches
	// Enhanced: includes precise timing and source identification for STDP
	for _, output := range outputsCopy {
		go n.sendToOutputWithSTDP(output, outputValue, now)
	}
}

// sendToOutputWithSTDP handles signal transmission to a single target neuron with STDP timing
// Enhanced version that includes precise timing and source identification for STDP learning
// Models the complete synaptic transmission process including:
// - Axonal conduction delay
// - Synaptic strength modulation
// - Neurotransmitter release and binding
// - Spike timing information for STDP learning
//
// Biological details modeled:
// - Conduction delay: time for action potential to travel along axon
// - Synaptic delay: time for neurotransmitter release and binding
// - Synaptic strength: efficacy of the synaptic connection (plastic via STDP)
// - Timing precision: exact spike timing for learning algorithms
// - Source identification: which neuron sent the spike
func (n *Neuron) sendToOutputWithSTDP(output *Output, baseValue float64, spikeTime time.Time) {
	// Add this defer function at the top of the method
	defer func() {
		if r := recover(); r != nil {
			// You can log this to see how often it happens and for which neuron/output
			// This is particularly useful during testing.
			// For example:
			// fmt.Printf("Recovered in sendToOutputWithSTDP (neuron: %s): %v. Target channel likely closed.\n", n.id, r)
		}
	}()

	if output.delay > 0 {
		time.Sleep(output.delay)
	}

	finalValue := baseValue * output.factor
	message := Message{
		Value:     finalValue,
		Timestamp: spikeTime,
		SourceID:  n.id,
	}

	// The select with default handles full channels, but not sends on already closed channels.
	// The defer/recover above handles the "send on closed channel" panic.
	select {
	case output.channel <- message:
		// Signal successfully transmitted
	default:
		// Target channel is full.
		// With the recover, if the channel was closed, the panic is caught,
		// and this default case might not even be reached for a closed channel scenario.
		// fmt.Printf("Neuron %s: Signal to output channel (type %T) lost or channel full/closed.\n", n.id, output.channel)
	}
}

// sendToOutput handles signal transmission without STDP timing (legacy method)
// Maintained for backward compatibility with existing code
// This method creates messages without timing information for simple networks
func (n *Neuron) sendToOutput(output *Output, baseValue float64) {
	// Apply biological transmission delay
	if output.delay > 0 {
		time.Sleep(output.delay)
	}

	// Calculate final signal strength
	finalValue := baseValue * output.factor

	// Create simple message without STDP timing information
	message := Message{
		Value:     finalValue,
		Timestamp: time.Time{}, // No timing information
		SourceID:  "",          // No source identification
	}

	// Attempt to deliver the signal
	select {
	case output.channel <- message:
		// Signal successfully transmitted
	default:
		// Signal lost due to full channel
	}
}

// resetAccumulator clears the integration state (thread-safe external interface)
// Models the return to resting membrane potential after signal processing
func (n *Neuron) resetAccumulator() {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	n.resetAccumulatorUnsafe()
}

// resetAccumulatorUnsafe clears integration state (internal use when locked)
// Returns the neuron to its resting state, ready for new signal integration
func (n *Neuron) resetAccumulatorUnsafe() {
	n.accumulator = 0
	// firstMessage will be reset when the next message arrives
	// This models the neuron returning to resting potential
}

// GetInput returns the input channel for connecting to this neuron
// This allows other neurons or external sources to send signals to this neuron
// Models: the dendritic tree where synaptic inputs are received
func (n *Neuron) GetInput() chan<- Message {
	return n.input
}

// Close gracefully shuts down the neuron
// Closes the input channel, which will cause the Run() loop to exit
// Models: neuronal death or disconnection from the network
func (n *Neuron) Close() {
	// Safely close the input channel to stop the Run() goroutine.
	// This prevents panics from sending to a closed channel.
	defer func() {
		if r := recover(); r != nil {
			// A panic might occur if the channel is already closed, which is fine.
			// We can ignore it.
		}
	}()
	close(n.input)
}

// GetHomeostaticInfo returns current homeostatic state information for monitoring
// This provides read-only access to the neuron's self-regulation metrics
// Thread-safe operation that allows external monitoring of homeostatic behavior
//
// Biological context:
// In neuroscience research, monitoring homeostatic state is crucial for
// understanding how neurons maintain stable activity levels and adapt to
// changing conditions. This method provides access to key metrics:
// - Current firing rate vs target rate
// - Calcium levels (activity sensor)
// - Threshold adjustments over time
// - Homeostatic regulation strength
//
// Usage patterns:
// - Network analysis: Understanding population dynamics
// - Debugging: Identifying neurons with homeostatic problems
// - Visualization: Real-time display of self-regulation
// - Research: Studying emergent network stability
//
// Returns a copy of homeostatic metrics to prevent external modification
func (n *Neuron) GetHomeostaticInfo() HomeostaticMetrics {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// Create a safe copy of homeostatic state
	// Note: firingHistory is copied to prevent external modification
	historyCopy := make([]time.Time, len(n.homeostatic.firingHistory))
	copy(historyCopy, n.homeostatic.firingHistory)

	return HomeostaticMetrics{
		firingHistory:         historyCopy,
		activityWindow:        n.homeostatic.activityWindow,
		targetFiringRate:      n.homeostatic.targetFiringRate,
		calciumLevel:          n.homeostatic.calciumLevel,
		calciumIncrement:      n.homeostatic.calciumIncrement,
		calciumDecayRate:      n.homeostatic.calciumDecayRate,
		homeostasisStrength:   n.homeostatic.homeostasisStrength,
		minThreshold:          n.homeostatic.minThreshold,
		maxThreshold:          n.homeostatic.maxThreshold,
		lastHomeostaticUpdate: n.homeostatic.lastHomeostaticUpdate,
		homeostaticInterval:   n.homeostatic.homeostaticInterval,
	}
}

// GetCurrentFiringRate returns the neuron's current firing rate in Hz
// Thread-safe method for monitoring neural activity levels
//
// This is useful for:
// - Real-time monitoring of network activity
// - Detecting silent or hyperactive neurons
// - Validating homeostatic regulation effectiveness
// - Network analysis and visualization
func (n *Neuron) GetCurrentFiringRate() float64 {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	return n.calculateCurrentFiringRateUnsafe()
}

// GetCurrentThreshold returns the neuron's current firing threshold
// This may differ from the original threshold due to homeostatic adjustments
// Thread-safe method for monitoring homeostatic regulation
//
// Biological context:
// Real neurons continuously adjust their firing thresholds based on activity.
// Monitoring these changes helps understand how the network self-regulates
// and maintains stable operation.
//
// Use cases:
// - Tracking homeostatic adjustments over time
// - Debugging network stability issues
// - Understanding adaptation mechanisms
// - Validating biological realism
func (n *Neuron) GetCurrentThreshold() float64 {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	return n.threshold
}

// GetBaseThreshold returns the original firing threshold before homeostatic adjustments
// This provides a reference point for understanding how much the neuron has adapted
// Thread-safe read-only access to the original threshold value
func (n *Neuron) GetBaseThreshold() float64 {
	// baseThreshold never changes, so no lock needed
	return n.baseThreshold
}

// GetCalciumLevel returns the current intracellular calcium level
// Thread-safe access to the biological activity sensor used for homeostatic regulation
//
// Biological context:
// Calcium serves as the primary activity sensor in real neurons. High calcium
// indicates recent high activity, while low calcium indicates the neuron has
// been quiet. This drives homeostatic threshold adjustments.
//
// Monitoring calcium levels helps understand:
// - How activity translates to homeostatic signals
// - Whether the calcium dynamics are working correctly
// - The relationship between firing and regulation
func (n *Neuron) GetCalciumLevel() float64 {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	return n.homeostatic.calciumLevel
}

// SetHomeostaticParameters allows dynamic adjustment of homeostatic regulation
// This enables runtime modification of self-regulation behavior for experimentation
// Thread-safe method for updating homeostatic settings
//
// Parameters:
// targetFiringRate: desired firing rate in Hz (0 disables homeostasis)
// homeostasisStrength: how aggressively to adjust threshold (0.0-1.0)
//
// Use cases:
// - Experimental manipulation of network dynamics
// - Adaptive adjustment of regulation strength
// - Disabling/enabling homeostasis during runtime
// - Testing different regulatory parameters
//
// Biological context:
// While real neurons have genetically determined homeostatic parameters,
// they can be modulated by neuromodulators, development, and experience.
// This method allows simulation of such modulation.
func (n *Neuron) SetHomeostaticParameters(targetFiringRate float64, homeostasisStrength float64) {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	n.homeostatic.targetFiringRate = targetFiringRate
	n.homeostatic.homeostasisStrength = homeostasisStrength

	// If homeostasis is being disabled, reset threshold to base value
	if targetFiringRate == 0.0 || homeostasisStrength == 0.0 {
		n.threshold = n.baseThreshold
	}
}

// ===================================================================================================
// BIOLOGICALLY ACCURATE SYNAPTIC SCALING IMPLEMENTATION
// ===================================================================================================
//
// This section implements post-synaptic receptor scaling, which is how synaptic scaling
// actually works in biology. The post-synaptic neuron controls its own receptor sensitivity
// to different input sources, allowing for homeostatic balance without requiring complex
// coordination between neurons.
//
// KEY BIOLOGICAL PRINCIPLES:
// 1. POST-SYNAPTIC CONTROL: The receiving neuron controls receptor sensitivity
// 2. INDEPENDENCE: No cross-neuron communication required
// 3. RECEPTOR SCALING: Models AMPA/NMDA receptor density changes
// 4. PATTERN PRESERVATION: Maintains relative input ratios
// 5. HOMEOSTATIC BALANCE: Prevents runaway excitation/inhibition
//
// ===================================================================================================

// applySynapticScaling performs biologically accurate receptor sensitivity scaling
// This is the core synaptic scaling algorithm that maintains stable total effective
// input strength while preserving the relative patterns learned through STDP
//
// # Uses real tracked activity for biological accuracy
//
// BIOLOGICAL PROCESS MODELED:
// 1. Post-synaptic neuron monitors actual effective input activity over time
// 2. Compares current activity patterns to target activity levels
// 3. Only scales when sufficient activity AND significant imbalance detected
// 4. Uses calcium-dependent gating (only scale during active periods)
// 5. Proportionally adjusts ALL receptor gains by the same factor
// 6. Preserves relative input ratios (learned patterns intact)
// 7. Operates on slower timescale than STDP (minutes vs milliseconds)
//
// BIOLOGICAL IMPROVEMENTS:
// - Activity-dependent scaling (only when neuron is active)
// - Real signal tracking (not estimated weights)
// - Calcium-gated scaling (biological activity sensor)
// - Minimum activity threshold (prevents scaling during silence)
func (n *Neuron) applySynapticScaling() {
	// Early exit if scaling is disabled
	if !n.scalingConfig.Enabled {
		return
	}

	// Check if it's time to perform scaling (respects biological timescales)
	now := time.Now()
	if now.Sub(n.scalingConfig.LastScalingUpdate) < n.scalingConfig.ScalingInterval {
		return
	}

	// === BIOLOGICAL ACTIVITY GATING ===
	// Only scale when there's sufficient neural activity
	// Models: calcium-dependent gene expression requires minimum activity levels
	n.stateMutex.Lock()
	calciumLevel := n.homeostatic.calciumLevel
	recentFiringRate := n.calculateCurrentFiringRateUnsafe()
	n.stateMutex.Unlock()

	// Biological gate: require minimum activity to trigger scaling
	if calciumLevel < n.minActivityForScaling || recentFiringRate < 0.1 {
		n.scalingConfig.LastScalingUpdate = now // Update timing but don't scale
		return                                  // Not enough activity for biological scaling
	}

	// === STEP 1: CALCULATE REAL EFFECTIVE INPUT STRENGTH ===
	n.inputGainsMutex.RLock()
	n.inputActivityMutex.RLock()

	// Skip scaling if no input sources registered
	if len(n.inputGains) == 0 {
		n.inputActivityMutex.RUnlock()
		n.inputGainsMutex.RUnlock()
		return
	}

	// Calculate current average effective input strength using REAL tracked activity
	totalEffectiveStrength := 0.0
	activeInputCount := 0

	for sourceID, _ := range n.inputGains {
		// Get actual recent activity for this source
		activities, hasActivity := n.inputActivityHistory[sourceID]
		if !hasActivity || len(activities) == 0 {
			continue // Skip sources with no recent activity
		}

		// Calculate average recent activity (biological integration)
		activitySum := 0.0
		for _, activity := range activities {
			activitySum += activity
		}
		averageActivity := activitySum / float64(len(activities))

		// This IS the effective strength (activity already includes gain effect)
		totalEffectiveStrength += averageActivity
		activeInputCount++
	}

	n.inputActivityMutex.RUnlock()
	n.inputGainsMutex.RUnlock()

	// Need minimum number of active inputs for meaningful scaling
	if activeInputCount == 0 {
		return
	}

	// Calculate current average effective input strength
	currentAverageStrength := totalEffectiveStrength / float64(activeInputCount)

	// === STEP 2: BIOLOGICAL SIGNIFICANCE TEST ===
	// Only scale if there's a significant deviation from target
	targetStrength := n.scalingConfig.TargetInputStrength
	strengthDifference := targetStrength - currentAverageStrength
	relativeError := math.Abs(strengthDifference) / targetStrength

	// Biological threshold: only scale for significant imbalances (>10%)
	if relativeError < 0.1 {
		n.scalingConfig.LastScalingUpdate = now
		return // Activity is close enough to target
	}

	// === STEP 3: CALCULATE SCALING FACTOR ===
	// Calculate scaling factor with gradual adjustment
	rawScalingFactor := 1.0 + (strengthDifference * n.scalingConfig.ScalingRate)

	// Apply safety bounds (prevent extreme scaling)
	scalingFactor := rawScalingFactor
	if scalingFactor < n.scalingConfig.MinScalingFactor {
		scalingFactor = n.scalingConfig.MinScalingFactor
	}
	if scalingFactor > n.scalingConfig.MaxScalingFactor {
		scalingFactor = n.scalingConfig.MaxScalingFactor
	}

	// Skip scaling if factor is very close to 1.0 (no significant change needed)
	if math.Abs(scalingFactor-1.0) < 0.0001 {
		n.scalingConfig.LastScalingUpdate = now
		return
	}

	// === STEP 4: APPLY BIOLOGICAL RECEPTOR SCALING ===
	n.inputGainsMutex.Lock()
	scaledGainCount := 0

	for sourceID, oldGain := range n.inputGains {
		// Only scale gains for sources with recent activity
		if activities, hasActivity := n.inputActivityHistory[sourceID]; hasActivity && len(activities) > 0 {
			// Calculate new receptor gain
			newGain := oldGain * scalingFactor

			// Apply biological bounds to receptor sensitivity
			minGain := 0.01 // Minimum receptor sensitivity
			maxGain := 10.0 // Maximum receptor sensitivity

			if newGain < minGain {
				newGain = minGain
			} else if newGain > maxGain {
				newGain = maxGain
			}

			// Apply the scaling (immediately affects signal processing)
			n.inputGains[sourceID] = newGain
			scaledGainCount++
		}
	}
	n.inputGainsMutex.Unlock()

	// === STEP 5: UPDATE SCALING STATE ===
	n.scalingConfig.LastScalingUpdate = now
	n.scalingConfig.ScalingHistory = append(n.scalingConfig.ScalingHistory, scalingFactor)

	// Limit history size
	maxHistorySize := 100
	if len(n.scalingConfig.ScalingHistory) > maxHistorySize {
		start := len(n.scalingConfig.ScalingHistory) - maxHistorySize
		n.scalingConfig.ScalingHistory = n.scalingConfig.ScalingHistory[start:]
	}

	// Biological analogy: This represents the completion of calcium-dependent
	// gene expression changes that adjust AMPA/NMDA receptor density at synapses
}

// EnableSynapticScaling enables synaptic scaling with specified parameters
// This method allows runtime activation of synaptic scaling for experimental control
//
// Parameters:
// targetStrength: desired average effective input strength
// scalingRate: how aggressively to adjust gains (0.0001 to 0.01)
// scalingInterval: time between scaling operations (e.g., 30 * time.Second)
//
// Biological context:
// In real neurons, synaptic scaling can be modulated by activity, development,
// and various signaling pathways. This method allows simulation of enabling
// these homeostatic mechanisms during network operation.
func (n *Neuron) EnableSynapticScaling(targetStrength float64, scalingRate float64, scalingInterval time.Duration) {
	n.scalingConfig.Enabled = true
	n.scalingConfig.TargetInputStrength = targetStrength
	n.scalingConfig.ScalingRate = scalingRate
	n.scalingConfig.ScalingInterval = scalingInterval
	n.scalingConfig.LastScalingUpdate = time.Now() // Reset timing
}

// DisableSynapticScaling disables synaptic scaling
// This method allows runtime deactivation of synaptic scaling for experimental control
//
// When disabled, receptor gains remain at their current values but no further
// scaling adjustments will be made. This is useful for studying the effects
// of scaling vs. non-scaling conditions.
func (n *Neuron) DisableSynapticScaling() {
	n.scalingConfig.Enabled = false
}

// GetSynapticScalingInfo returns current synaptic scaling state for monitoring
// Thread-safe method for observing scaling behavior and parameters
//
// Returns information about:
// - Current scaling configuration
// - Recent scaling history
// - Current receptor gains for all input sources
// - Effective input strengths
//
// Useful for:
// - Debugging scaling behavior
// - Monitoring network stability
// - Research and analysis
// - Visualization of homeostatic processes
func (n *Neuron) GetSynapticScalingInfo() map[string]interface{} {
	info := make(map[string]interface{})

	// Copy scaling configuration
	info["enabled"] = n.scalingConfig.Enabled
	info["targetInputStrength"] = n.scalingConfig.TargetInputStrength
	info["scalingRate"] = n.scalingConfig.ScalingRate
	info["scalingInterval"] = n.scalingConfig.ScalingInterval
	info["minScalingFactor"] = n.scalingConfig.MinScalingFactor
	info["maxScalingFactor"] = n.scalingConfig.MaxScalingFactor
	info["lastScalingUpdate"] = n.scalingConfig.LastScalingUpdate

	// Copy scaling history (safe copy)
	historyCopy := make([]float64, len(n.scalingConfig.ScalingHistory))
	copy(historyCopy, n.scalingConfig.ScalingHistory)
	info["scalingHistory"] = historyCopy

	// Copy current receptor gains (thread-safe)
	n.inputGainsMutex.RLock()
	gainsCopy := make(map[string]float64)
	for sourceID, gain := range n.inputGains {
		gainsCopy[sourceID] = gain
	}
	n.inputGainsMutex.RUnlock()
	info["receptorGains"] = gainsCopy

	// Calculate current statistics
	if len(gainsCopy) > 0 {
		totalGain := 0.0
		minGain := math.Inf(1)
		maxGain := math.Inf(-1)

		for _, gain := range gainsCopy {
			totalGain += gain
			if gain < minGain {
				minGain = gain
			}
			if gain > maxGain {
				maxGain = gain
			}
		}

		info["averageGain"] = totalGain / float64(len(gainsCopy))
		info["minGain"] = minGain
		info["maxGain"] = maxGain
		info["numInputSources"] = len(gainsCopy)
	} else {
		info["averageGain"] = 0.0
		info["minGain"] = 0.0
		info["maxGain"] = 0.0
		info["numInputSources"] = 0
	}

	return info
}

// GetInputGains returns a copy of current receptor gains for all input sources
// Thread-safe method for accessing current synaptic scaling state
//
// Returns:
// Map of source neuron ID to current receptor gain (sensitivity)
// Empty map if no input sources are registered
//
// Use cases:
// - Monitoring scaling effects on individual synapses
// - Analyzing input balance across sources
// - Research on homeostatic mechanisms
// - Network visualization and debugging
func (n *Neuron) GetInputGains() map[string]float64 {
	n.inputGainsMutex.RLock()
	defer n.inputGainsMutex.RUnlock()

	// Create a safe copy to prevent external modification
	gainsCopy := make(map[string]float64)
	for sourceID, gain := range n.inputGains {
		gainsCopy[sourceID] = gain
	}

	return gainsCopy
}

// SetInputGain manually sets the receptor gain for a specific input source
// This method allows experimental manipulation of individual receptor sensitivities
//
// Parameters:
// sourceID: identifier of the input source neuron
// gain: new receptor sensitivity (typically 0.1 to 10.0)
//
// Use cases:
// - Experimental manipulation of specific synapses
// - Testing effects of receptor density changes
// - Simulating pharmacological interventions
// - Research on synaptic scaling mechanisms
//
// Note: This manually set gain will be overwritten by automatic scaling
// if synaptic scaling is enabled. Disable scaling first if you want to
// maintain manual control.
func (n *Neuron) SetInputGain(sourceID string, gain float64) {
	n.inputGainsMutex.Lock()
	defer n.inputGainsMutex.Unlock()

	// Ensure gains map is initialized
	if n.inputGains == nil {
		n.inputGains = make(map[string]float64)
	}

	// Apply reasonable bounds
	if gain < 0.001 {
		gain = 0.001 // Minimum sensitivity
	} else if gain > 100.0 {
		gain = 100.0 // Maximum sensitivity
	}

	n.inputGains[sourceID] = gain
}

// GetOutputWeight returns the current synaptic weight of a specific output connection.
// This is a thread-safe method for monitoring and validating learning.
//
// Parameters:
// id: The unique identifier of the output connection.
//
// Returns:
// The current weight (factor) and a boolean indicating if the output was found.
func (n *Neuron) GetOutputWeight(id string) (float64, bool) {
	n.outputsMutex.RLock()
	defer n.outputsMutex.RUnlock()

	output, exists := n.outputs[id]
	if !exists {
		return 0, false
	}
	return output.factor, true
}

// Receive accepts a synapse message and integrates it into the existing neuron processing pipeline
// This method serves as a bridge between the new synapse system and the existing neuron architecture,
// allowing synapses to deliver signals to neurons without breaking existing functionality.
//
// BRIDGE PATTERN IMPLEMENTATION:
// This method implements the adapter/bridge pattern to allow the new synapse system to work
// seamlessly with existing neuron implementations. It translates between the synapse message
// format and the neuron's existing message format, preserving all timing and source information
// needed for STDP learning while maintaining backward compatibility.
//
// BIOLOGICAL CONTEXT:
// In real neurons, synaptic inputs arrive at dendrites and are integrated at the cell body (soma).
// This method models the dendritic integration process where:
// 1. Synaptic signals arrive with precise timing information
// 2. Signals are converted to postsynaptic potentials
// 3. Potentials are integrated with existing membrane dynamics
// 4. The neuron's existing firing logic determines the response
//
// CONCURRENCY SAFETY:
// This method is thread-safe and designed to be called from multiple synapse goroutines
// simultaneously. The select statement with default case ensures non-blocking operation,
// preventing synapses from being blocked if the neuron's input buffer is full.
//
// INTEGRATION STRATEGY:
// By converting SynapseMessage to the existing Message format, this method allows:
// - New synapse system to coexist with existing Output system
// - Gradual migration from Output to synapse-based connectivity
// - Preservation of all existing neuron processing logic (homeostasis, STDP, etc.)
// - No modifications needed to existing neuron internal processing methods
//
// Parameters:
//
//	msg: SynapseMessage containing the synaptic signal with timing and source information
//
// The method preserves all essential information for neural processing:
// - Signal strength (Value): Determines the magnitude of postsynaptic potential
// - Timing (Timestamp): Critical for STDP learning and temporal dynamics
// - Source identification (SourceID): Enables synapse-specific learning and tracking
// func (n *Neuron) Receive(msg SynapseMessage) {
// 	// Convert SynapseMessage to your existing Message format
// 	// This translation preserves all critical information while maintaining compatibility
// 	// with existing neuron processing logic that expects the original Message structure
// 	existingMsg := Message{
// 		Value:     msg.Value,     // Signal strength - models postsynaptic potential amplitude
// 		Timestamp: msg.Timestamp, // Precise spike timing - essential for STDP calculations
// 		SourceID:  msg.SourceID,  // Source neuron ID - enables input-specific learning
// 	}

// 	// Forward to your existing input processing pipeline
// 	// This integrates the synaptic signal into the neuron's standard processing workflow,
// 	// ensuring that synaptic inputs are handled identically to existing input sources
// 	// while maintaining the neuron's existing behavior for homeostasis, threshold dynamics,
// 	// calcium tracking, and all other biological features
// 	select {
// 	case n.input <- existingMsg:
// 		// Successfully forwarded to existing system
// 		// The message will now be processed by the neuron's Run() loop using all
// 		// existing biological mechanisms:
// 		// - Leaky integration (applyMembraneDecay)
// 		// - Homeostatic regulation (calcium dynamics, threshold adjustment)
// 		// - STDP learning (if enabled in the neuron)
// 		// - Refractory period enforcement
// 		// - Fire event generation and monitoring

// 	default:
// 		// Drop if input buffer is full (same as existing behavior)
// 		// This non-blocking approach models biological synaptic failure that can occur
// 		// when the postsynaptic neuron is overwhelmed with inputs. In real neurons:
// 		// - Synaptic vesicles can be depleted during high activity
// 		// - Postsynaptic receptors can become saturated
// 		// - Dendritic processing can reach capacity limits
// 		//
// 		// By dropping the message rather than blocking, we:
// 		// - Prevent deadlocks in the synapse system
// 		// - Model realistic biological saturation effects
// 		// - Maintain the existing neuron's performance characteristics
// 		// - Preserve the behavior that existing unit tests expect
// 	}
// }

// Helper function for absolute value (not available in older Go versions)
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
