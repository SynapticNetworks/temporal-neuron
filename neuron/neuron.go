/*
=================================================================================
BIOLOGICAL NEURON SIMULATION - SYNAPSE-INTEGRATED ARCHITECTURE
=================================================================================

OVERVIEW:
This package implements a biologically-inspired artificial neuron that serves as
the fundamental building block for constructing neural networks with dynamic
connectivity, realistic timing behavior, homeostatic self-regulation, and
synaptic scaling for long-term stability.

The neuron integrates seamlessly with the synapse package, which provides
biologically accurate synaptic connections with STDP learning, structural
plasticity, and realistic transmission dynamics.

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

6. DYNAMIC CONNECTIVITY: Biological neurons can grow new connections (synapses)
   and prune existing ones throughout their lifetime (neuroplasticity)

7. PARALLEL TRANSMISSION: A single action potential propagates to ALL connected
   neurons simultaneously through the axon's branching structure

8. TRANSMISSION DELAYS: Different connections have different delays based on
   axon length, diameter, and myelination

9. REFRACTORY PERIODS: Cannot fire immediately after firing (recovery time)

10. LEAKY INTEGRATION: Membrane potential naturally decays over time

SYNAPSE INTEGRATION:
This neuron implementation is designed to work with the synapse package, which
provides biologically accurate synaptic connections with:
- STDP learning capabilities
- Realistic transmission delays
- Structural plasticity (pruning)
- Thread-safe concurrent operation
- Message-based communication

MULTI-TIMESCALE PLASTICITY:
This implementation models the multiple timescales of biological plasticity:

- STDP (milliseconds to seconds): Fast synaptic learning based on spike timing
- Homeostatic Plasticity (seconds to minutes): Intrinsic excitability adjustment
- Synaptic Scaling (minutes to hours): Proportional adjustment of receptor sensitivity

These mechanisms work together to create stable yet adaptive learning.

=================================================================================
*/

package neuron

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/synapse"
)

// FireEvent represents a real-time neuron firing event for visualization and monitoring
// This captures the exact moment when a biological neuron generates an action potential
// and provides essential information about the firing event for external observers
//
// Biological context:
// When a real neuron fires, it generates an action potential that propagates down
// its axon to all connected synapses. This event is instantaneous and discrete.
// Unlike traditional ANNs that have continuous activation values, biological
// neurons either fire (1) or don't fire (0) - this is the "all-or-nothing" principle.
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

// SynapticScalingConfig contains all parameters controlling synaptic scaling behavior
// This structure encapsulates the homeostatic mechanism that maintains synaptic balance
//
// BIOLOGICAL BACKGROUND:
// Synaptic scaling is a homeostatic mechanism observed in real neurons that prevents
// runaway strengthening or weakening of synaptic connections. When total synaptic
// input becomes too strong or weak, neurons proportionally scale their receptor
// sensitivity to maintain optimal responsiveness while preserving learned patterns.
type SynapticScalingConfig struct {
	Enabled bool // Master switch for synaptic scaling functionality
	// When false, synaptic scaling is completely disabled
	// When true, scaling occurs according to the parameters below

	// === CORE SCALING PARAMETERS ===
	// These control the target behavior and speed of synaptic scaling

	TargetInputStrength float64 // Desired average effective input strength
	// Biological interpretation: optimal total synaptic drive for this neuron
	// This is the target for (synaptic_weight × receptor_gain) averaged across inputs
	// Typical values: 0.5-2.0 depending on neuron type and network role

	ScalingRate float64 // Rate of receptor gain adjustment per scaling event
	// Controls how aggressively gains are adjusted toward target
	// Range: 0.0001 (very conservative) to 0.01 (aggressive)
	// Higher values = faster correction but potentially less stable

	ScalingInterval time.Duration // Time between synaptic scaling operations
	// Biological timescale: much slower than STDP (which operates in milliseconds)
	// Typical range: 10 seconds to 10 minutes
	// Shorter intervals = more responsive but higher computational cost

	// === SAFETY CONSTRAINTS ===
	// These prevent extreme scaling that could destabilize the network

	MinScalingFactor float64 // Minimum multiplier applied to gains per scaling event
	// Prevents excessive reduction in a single scaling operation
	// Typical values: 0.8-0.95 (don't reduce gains by more than 5-20% per event)

	MaxScalingFactor float64 // Maximum multiplier applied to gains per scaling event
	// Prevents excessive increase in a single scaling operation
	// Typical values: 1.05-1.2 (don't increase gains by more than 5-20% per event)

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

// HomeostaticInfo contains read-only homeostatic state information
// Used for monitoring and analysis of neural self-regulation
type HomeostaticInfo struct {
	targetFiringRate      float64       // Target firing rate for regulation
	homeostasisStrength   float64       // Strength of homeostatic adjustments
	calciumLevel          float64       // Current calcium concentration
	firingHistory         []time.Time   // Recent firing times (copy)
	minThreshold          float64       // Minimum allowed threshold
	maxThreshold          float64       // Maximum allowed threshold
	activityWindow        time.Duration // Time window for rate calculation
	lastHomeostaticUpdate time.Time     // When homeostasis last ran
}

// delayedMessage represents a synaptic message awaiting axonal delivery.
// Models action potential propagation down the axon with timing delays
// before reaching target synapses. Includes both synaptic and spatial delays.
type delayedMessage struct {
	message      synapse.SynapseMessage          // The synaptic message to deliver
	target       synapse.SynapseCompatibleNeuron // Target post-synaptic neuron
	deliveryTime time.Time                       // When the message should be delivered
}

// InputActivity represents a single, discrete synaptic input event and its effect on the postsynaptic neuron.
// This structure serves as a fundamental unit of "synaptic trace memory," a short-lived record of
// recent activity that is essential for the neuron's internal homeostatic and computational mechanisms.
//
// BIOLOGICAL CONTEXT:
// When a presynaptic neuron fires, it causes a transient change in the postsynaptic neuron's membrane
// potential, known as a postsynaptic potential (PSP). The InputActivity struct is a digital model of a single PSP.
//
// Within the neuron, this history of PSPs is used for:
//   - SYNAPTIC SCALING: The neuron monitors the average strength of these inputs over seconds to minutes to
//     proportionally scale its synaptic gains and maintain homeostatic balance.
//   - TEMPORAL SUMMATION & COINCIDENCE DETECTION: The neuron uses the timing and strength of recent events
//     to integrate signals and detect correlated activity within a millisecond-scale window.
type InputActivity struct {
	// EffectiveValue represents the final strength and polarity of the postsynaptic potential (PSP).
	// This value is the result of the entire synaptic transmission process, including the presynaptic
	// signal strength, the synapse's current weight (efficacy), and any postsynaptic scaling factors.
	//
	// BIOLOGICAL BASIS:
	// - A positive value models an Excitatory Postsynaptic Potential (EPSP), which depolarizes the
	//   membrane (e.g., by opening Na+ channels) and pushes the neuron closer to its firing threshold.
	// - A negative value models an Inhibitory Postsynaptic Potential (IPSP), which hyperpolarizes
	//   the membrane (e.g., by opening Cl- channels) and makes the neuron less likely to fire.
	EffectiveValue float64

	// Timestamp marks the precise time the synaptic input was received by the postsynaptic neuron.
	//
	// BIOLOGICAL BASIS:
	// This precise timing is essential for the neuron's internal computations. While learning rules like
	// Spike-Timing-Dependent Plasticity (STDP) also rely on this timing, in this architecture the
	// STDP calculation itself is handled by the `synapse` package. The neuron uses this timestamp for its
	// own purposes, such as:
	// - Coincidence Detection: Identifying inputs that arrive within the configured CoincidenceWindow.
	// - Activity Tracking: Calculating recent average input rates for homeostatic scaling.
	Timestamp time.Time
}

// Neuron represents a single processing unit inspired by biological neurons
// Unlike traditional artificial neurons that perform instantaneous calculations,
// this neuron models the temporal dynamics of real neural processing:
// - Accumulates inputs over time (like dendrite integration)
// - Fires when threshold is reached (like action potential generation)
// - Sends outputs through synapses with realistic delays (like axon transmission)
// - Supports dynamic connectivity changes (like neuroplasticity)
// - Maintains stable activity through homeostatic regulation (like real neurons)
// - Scales synaptic sensitivity to maintain input balance (like receptor scaling)
//
// The neuron is designed to work seamlessly with the synapse package for
// biologically accurate connections with STDP learning and structural plasticity.
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

	// === NEW: INPUT BUFFERING SYSTEM ===
	// Models biological dendritic capacity constraints for simultaneous inputs.
	// Real dendrites can only handle limited concurrent postsynaptic potentials
	// before saturation occurs, leading to input loss under overload conditions.
	inputBuffer chan synapse.SynapseMessage // Bounded buffer (size 10) for incoming messages

	// === NEW: AXONAL DELIVERY SYSTEM ===
	// Models the biological axon: each neuron has one axon that handles all
	// outgoing signal delivery with appropriate propagation delays. This replaces
	// the goroutine-per-message approach with a single delivery worker per neuron,
	// eliminating goroutine explosion while maintaining biological realism.
	deliveryQueue  chan delayedMessage // Queue for messages awaiting axonal delivery
	deliveryWorker sync.Once           // Ensures exactly one delivery goroutine per neuron

	decayRate float64 // Membrane potential decay rate per time step
	// Models the leaky nature of biological neural membranes
	// In biology: membrane capacitance causes gradual charge dissipation
	// Value between 0.0-1.0: 0.95 = loses 5% charge per decay interval
	// Real neurons: membrane time constant typically 10-20ms

	// === COINCIDENCE DETECTION PARAMETERS ===
	// These parameters control the neuron's ability to act as a coincidence detector,
	// a fundamental computational role where a neuron fires preferentially in response
	// to multiple, near-simultaneous excitatory inputs.
	//
	// Biological basis:
	// This models the function of specialized synaptic receptors like NMDA, which act
	// as molecular "and-gates." They require both neurotransmitter binding (an input)
	// and significant membrane depolarization (often from other coincident inputs)
	// to activate. This mechanism is crucial for learning, memory formation, and
	// processing correlated signals in the brain.

	EnableCoincidenceDetection bool // Master switch to enable or disable coincidence detection logic.
	// Models the biological diversity of neurons; some neuron types are specialized
	// as powerful coincidence detectors (e.g., with high densities of NMDA receptors),
	// while others act as more general integrators.

	CoincidenceWindow time.Duration // The time window within which inputs are considered simultaneous.
	// Models the temporal integration window of the postsynaptic membrane. This is
	// determined by biophysical properties like the membrane time constant (how fast
	// charge leaks) and the kinetics of synaptic receptors.
	// Typical biological values range from 5ms to 20ms.

	// === HOMEOSTATIC PLASTICITY STATE ===
	// Models the biological mechanisms for activity monitoring and self-regulation

	homeostatic HomeostaticMetrics // All homeostatic plasticity state and parameters
	// Encapsulates the complex biological machinery for activity sensing,
	// threshold adjustment, and activity regulation that maintains network stability

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

	inputActivityHistory map[string][]InputActivity // Recent input signal strengths per source
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

	outputSynapses map[string]synapse.SynapticProcessor // Dynamic set of output synapses
	// Models the axon branching to multiple targets with sophisticated synapses
	// String key allows named connections for management
	// Each synapse handles its own learning, delays, and plasticity

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

	// === DENDRITIC INTEGRATION STRATEGY ===
	// This is the new field that holds the current strategy for processing inputs.
	dendriticIntegrationMode DendriticIntegrationMode

	// === LIFECYCLE MANAGEMENT ===
	// Use a context for managing the lifecycle of the neuron's goroutine.
	// This is a standard and robust pattern in Go for managing cancellation.
	ctx       context.Context
	cancel    context.CancelFunc // Function to signal shutdown
	wg        sync.WaitGroup     // WaitGroup to ensure background tasks finish before Close() returns
	closeOnce sync.Once

	// === MONITORING AND OBSERVATION ===
	// Optional channel for reporting firing events to external observers
	fireEvents chan<- FireEvent // Optional fire event reporting channel
	// nil = disabled (default), non-nil = reports firing events
	// Used for visualization, learning algorithms, and analysis
}

// NewNeuron creates and initializes a new biologically-inspired neuron with homeostatic plasticity
// This factory function sets up all the necessary components for realistic neural processing
// with leaky integration, dynamic connectivity, refractory periods, homeostatic regulation,
// and biologically accurate synaptic scaling
//
// The complete biological learning system enables:
// - Automatic activity monitoring through calcium-based sensing (homeostatic plasticity)
// - Self-regulation of firing threshold to maintain target activity levels
// - Post-synaptic receptor scaling for input balance (synaptic scaling)
// - Prevention of runaway excitation or neural silence
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
//
// Biological learning mechanisms:
// - Homeostatic: tracks firing history and adjusts threshold to maintain target rate
// - Synaptic Scaling: adjusts post-synaptic receptor sensitivity to maintain input balance
// - Combined: creates stable yet adaptive networks
func NewNeuron(id string, threshold float64, decayRate float64, refractoryPeriod time.Duration, fireFactor float64, targetFiringRate float64, homeostasisStrength float64) *Neuron {
	// Create a cancellable context to manage the neuron's lifecycle.
	ctx, cancel := context.WithCancel(context.Background())
	// Calculate homeostatic bounds based on base threshold
	// Biological rationale: neurons can't adjust indefinitely - there are biophysical limits
	minThreshold := threshold * 0.1 // Can reduce to 10% of original (very excitable)
	maxThreshold := threshold * 5.0 // Can increase to 5x original (very quiet)

	// Set up homeostatic parameters with biologically reasonable defaults
	activityWindow := 5 * time.Second // Track activity over 5 seconds
	calciumIncrement := 1.0           // Arbitrary units of calcium per spike
	calciumDecayRate := 0.9995        // The original calcium decay rate of 0.98 per millisecond was far too aggressive.
	// This new value provides a half-life of ~1.4 seconds, allowing calcium to integrate
	// activity over a biologically plausible timescale.
	homeostaticInterval := 100 * time.Millisecond // Check homeostasis every 100ms

	return &Neuron{
		id:                         id,                                         // Unique neuron identifier for network tracking
		threshold:                  threshold,                                  // Current firing threshold (homeostatic)
		baseThreshold:              threshold,                                  // Original threshold (reference)
		decayRate:                  decayRate,                                  // Membrane decay rate (biological: based on RC time constant)
		refractoryPeriod:           refractoryPeriod,                           // Refractory period (biological: ~5-15ms)
		fireFactor:                 fireFactor,                                 // Output amplitude scaling
		outputSynapses:             make(map[string]synapse.SynapticProcessor), // Dynamic synapse connections
		fireEvents:                 nil,                                        // Optional fire event reporting (disabled by default)
		EnableCoincidenceDetection: false,                                      // deactivated by default - TODO check if we can activate
		CoincidenceWindow:          50 * time.Millisecond,

		// NEW: Initialize input buffering system
		// Buffer size of 10 reflects biological dendritic capacity constraints
		// Larger buffers reduce message loss but increase memory usage
		inputBuffer: make(chan synapse.SynapseMessage, 10),

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

		// Initialize synaptic scaling system
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
		dendriticIntegrationMode: NewPassiveMembraneMode(),

		// Initialize activity tracking
		inputActivityHistory:   make(map[string][]InputActivity),
		activityTrackingWindow: 10 * time.Second, // Track activity over 10 seconds
		minActivityForScaling:  0.1,              // Minimum activity for scaling
		lastActivityCleanup:    time.Now(),
		ctx:                    ctx,    // Lifecycle context
		cancel:                 cancel, // Function to stop the neuron
	}
}

// NewSimpleNeuron creates a neuron with homeostatic plasticity disabled for backward compatibility
// This convenience function creates a neuron that behaves like the original implementation
// but with the learning infrastructure in place (just not active)
//
// Use this when you want the original temporal neuron behavior without self-regulation,
// or when building networks that will implement learning through other mechanisms
//
// Parameters are the same as the original NewNeuron function:
// id: unique identifier for this neuron
// threshold: firing threshold (fixed, no homeostatic adjustment)
// decayRate: membrane potential decay factor per time step (0.0-1.0)
// refractoryPeriod: duration after firing when neuron cannot fire again
// fireFactor: action potential amplitude/strength
func NewSimpleNeuron(id string, threshold float64, decayRate float64, refractoryPeriod time.Duration, fireFactor float64) *Neuron {
	// Create neuron with homeostatic plasticity disabled
	return NewNeuron(
		id,
		threshold,
		decayRate,
		refractoryPeriod,
		fireFactor,
		0.0, // targetFiringRate = 0 disables homeostatic regulation
		0.0, // homeostasisStrength = 0 disables threshold adjustments
	)
}

// NewNeuronWithLearning creates a neuron with homeostatic plasticity enabled
// This convenience constructor sets up a neuron with biologically realistic learning
// parameters suitable for most applications
//
// Parameters:
// id: unique identifier for this neuron
// threshold: base firing threshold
// targetFiringRate: desired firing rate in Hz for homeostatic regulation
//
// Returns a neuron with:
// - Moderate homeostatic regulation (20% strength)
// - Reasonable threshold bounds (0.1x to 5x base threshold)
// - Standard biological timing parameters
func NewNeuronWithLearning(id string, threshold float64, targetFiringRate float64) *Neuron {
	// Standard biological parameters
	decayRate := 0.95
	refractoryPeriod := 10 * time.Millisecond
	fireFactor := 1.0
	homeostasisStrength := 0.2

	return NewNeuron(id, threshold, decayRate, refractoryPeriod, fireFactor,
		targetFiringRate, homeostasisStrength)
}

// GetInputStrengths returns the strengths of recent input activities for a given source ID.
// It provides compatibility for tests expecting float64 slices from inputActivityHistory.
// This method is safe for concurrent use.
func (n *Neuron) GetInputStrengths(sourceID string) []float64 {
	n.inputActivityMutex.RLock()
	defer n.inputActivityMutex.RUnlock()
	var strengths []float64
	for _, activity := range n.inputActivityHistory[sourceID] {
		strengths = append(strengths, math.Abs(activity.EffectiveValue))
	}
	return strengths
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
	if n.homeostatic.homeostasisStrength == 0.0 || n.homeostatic.targetFiringRate == 0.0 {
		return
	}
	currentRate := n.calculateCurrentFiringRateUnsafe()
	rateError := currentRate - n.homeostatic.targetFiringRate
	// Allow adjustment if calcium is sufficient OR rate is critically low
	if n.homeostatic.calciumLevel < n.minActivityForScaling && currentRate > 0 {
		return // Skip if low calcium and some activity
	}
	scalingFactor := 1.0 + (rateError * n.homeostatic.homeostasisStrength * 0.005)
	if scalingFactor > 1.05 {
		scalingFactor = 1.05
	} else if scalingFactor < 0.95 {
		scalingFactor = 0.95
	}
	newThreshold := n.threshold * scalingFactor
	if newThreshold < n.homeostatic.minThreshold {
		newThreshold = n.homeostatic.minThreshold
	} else if newThreshold > n.homeostatic.maxThreshold {
		newThreshold = n.homeostatic.maxThreshold
	}
	n.threshold = newThreshold
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
// - Learning algorithms: External learning systems
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
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	n.fireEvents = ch
}

// AddOutputSynapse safely adds a new synaptic connection to this neuron
// This models neuroplasticity - the brain's ability to form new connections
// throughout life. In developing brains, neurons constantly grow new synapses.
// In adult brains, learning involves creating and strengthening connections.
//
// The synapse handles all aspects of synaptic transmission including:
// - Signal transmission with realistic delays
// - STDP learning based on spike timing
// - Structural plasticity (pruning decisions)
// - Thread-safe concurrent operation
//
// Biological context:
// - Dendritic growth: neurons extend dendrites to reach new partners
// - Axon sprouting: axons grow new branches to contact more targets
// - Synaptogenesis: formation of new synaptic contacts
// - Experience-dependent plasticity: activity drives connection formation
//
// Parameters:
// id: unique identifier for this connection (allows later modification/removal)
// synapseProcessor: the synapse object that handles transmission and learning
func (n *Neuron) AddOutputSynapse(id string, synapseProcessor synapse.SynapticProcessor) {
	n.outputsMutex.Lock()
	defer n.outputsMutex.Unlock()

	n.outputSynapses[id] = synapseProcessor
}

// RemoveOutputSynapse safely removes a synaptic connection
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
func (n *Neuron) RemoveOutputSynapse(id string) {
	n.outputsMutex.Lock()
	defer n.outputsMutex.Unlock()

	delete(n.outputSynapses, id)
}

// GetOutputSynapseCount returns the current number of synaptic connections
// Thread-safe read operation that allows monitoring network connectivity
// In biological terms: this tells us the neuron's "fan-out" or how many
// other neurons this neuron can directly influence
func (n *Neuron) GetOutputSynapseCount() int {
	n.outputsMutex.RLock()
	defer n.outputsMutex.RUnlock()

	return len(n.outputSynapses)
}

// GetOutputSynapseWeight returns the current synaptic weight of a specific output connection
// This is a thread-safe method for monitoring and validating learning
//
// Parameters:
// id: The unique identifier of the output synapse
//
// Returns:
// The current weight and a boolean indicating if the synapse was found
func (n *Neuron) GetOutputSynapseWeight(id string) (float64, bool) {
	n.outputsMutex.RLock()
	defer n.outputsMutex.RUnlock()

	synapseProcessor, exists := n.outputSynapses[id]
	if !exists {
		return 0, false
	}
	return synapseProcessor.GetWeight(), true
}

// Run starts the main neuron processing loop with continuous leaky integration,
// homeostatic regulation, and synaptic scaling
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
// - Homeostatic plasticity: seconds to minutes (intrinsic regulation)
// - Synaptic scaling: minutes to hours (synaptic homeostasis - slowest)
//
// BUFFERED ARCHITECTURE:
// Inputs are processed individually from the buffer, maintaining temporal order
// and allowing immediate firing checks after each input. This models the
// continuous nature of biological neural processing while using discrete
// computational cycles for efficiency.
//
// MUST be called as a goroutine: go neuron.Run()
// Each neuron operates independently, modeling biological neural autonomy.
func (n *Neuron) Run() {
	// Create decay timer for continuous biological processes
	// 1ms interval provides good resolution for neural dynamics while
	// remaining computationally efficient
	decayTicker := time.NewTicker(1 * time.Millisecond)
	defer decayTicker.Stop()

	// Main neural processing loop - handles inputs and biological maintenance
	for {
		select {
		// EVENT 1: Process individual synaptic input
		// Models arrival of postsynaptic potential (PSP) at dendrites
		// Each input processed individually to maintain temporal precision
		case msg := <-n.inputBuffer:
			n.stateMutex.Lock()

			// Apply post-synaptic receptor scaling (synaptic homeostasis)
			// Models AMPA/NMDA receptor density changes at post-synaptic sites
			finalSignalValue := n.applyPostSynapticGainUnsafe(msg)
			modifiedMsg := msg
			modifiedMsg.Value = finalSignalValue

			// Record input activity for homeostatic scaling and coincidence detection
			// Models calcium-dependent activity sensing for regulatory mechanisms
			if (n.scalingConfig.Enabled || n.EnableCoincidenceDetection) && modifiedMsg.SourceID != "" {
				n.recordInputActivityUnsafe(modifiedMsg.SourceID, finalSignalValue)
			}

			// Process through dendritic integration strategies
			// Models sophisticated dendritic computation (summation, gating, etc.)
			result := n.dendriticIntegrationMode.Handle(modifiedMsg)
			if result != nil {
				n.accumulator += result.NetInput
			}

			// Immediate firing check after each input
			// Models continuous threshold monitoring at axon hillock
			if n.accumulator >= n.threshold {
				n.fireUnsafe()
				n.resetAccumulatorUnsafe()
			}

			n.stateMutex.Unlock()

		// EVENT 2: Regular biological maintenance cycle
		// Models continuous cellular processes independent of synaptic input
		case <-decayTicker.C:
			n.stateMutex.Lock()

			// Process any buffered dendritic computations
			// Some integration modes batch inputs for temporal summation
			stateSnapshot := MembraneSnapshot{
				Accumulator:      n.accumulator,
				CurrentThreshold: n.threshold,
			}
			result := n.dendriticIntegrationMode.Process(stateSnapshot)
			if result != nil {
				n.accumulator += result.NetInput
			}

			// Apply continuous membrane potential decay
			// Models charge dissipation through membrane resistance (RC circuit)
			n.applyMembraneDecayUnsafe()

			// Update calcium dynamics for homeostatic sensing
			// Models intracellular calcium as biological activity sensor
			n.updateCalciumLevelUnsafe()

			// Check firing after decay (may have reached threshold)
			if n.accumulator >= n.threshold {
				n.fireUnsafe()
				n.resetAccumulatorUnsafe()
			}

			// Homeostatic threshold adjustment (seconds to minutes timescale)
			// Models activity-dependent intrinsic excitability regulation
			if n.shouldPerformHomeostaticUpdateUnsafe() {
				n.performHomeostaticAdjustmentUnsafe()
			}

			n.stateMutex.Unlock()

		// EVENT 3: Controlled shutdown
		case <-n.ctx.Done():
			return
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
// Parameters:
// sourceID: identifier of the input source neuron
// effectiveSignalValue: final signal strength (signal × post-gain)
//
// mustBeLocked: true (stateMutex must be held by caller)
func (n *Neuron) recordInputActivityUnsafe(sourceID string, effectiveSignalValue float64) {
	// Initialize activity tracking structures if needed
	if n.inputActivityHistory == nil {
		n.inputActivityHistory = make(map[string][]InputActivity)
	}

	// Get current time for activity timestamp
	now := time.Now()

	// === RECORD NEW ACTIVITY ===
	// Add this signal to the activity history for this source
	// Models: accumulation of synaptic activity over biological time windows
	n.inputActivityMutex.Lock()
	n.inputActivityHistory[sourceID] = append(n.inputActivityHistory[sourceID], InputActivity{
		EffectiveValue: effectiveSignalValue,
		Timestamp:      now,
	})
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
	cutoff := currentTime.Add(-n.activityTrackingWindow)
	for sourceID, activities := range n.inputActivityHistory {
		var valid []InputActivity
		for _, activity := range activities {
			if activity.Timestamp.After(cutoff) {
				valid = append(valid, activity)
			}
		}
		n.inputActivityHistory[sourceID] = valid
	}
}

// SetCoincidenceDetection enables or disables coincidence detection and sets the window
// This allows dynamic configuration of temporal synaptic coincidence behavior
// Parameters:
// enabled: true to enable coincidence detection, false to disable
// window: time window for detecting coincident inputs (e.g., 50ms)
func (n *Neuron) SetCoincidenceDetection(enabled bool, window time.Duration) {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	n.EnableCoincidenceDetection = enabled
	if window > 0 {
		n.CoincidenceWindow = window
	}
}

// DetectCoincidentInputs returns the number of unique excitatory input sources active within the configured coincidence window.
// It supports temporal synaptic coincidence by identifying near-simultaneous inputs, excluding inhibitory feedback.
// This method is safe for concurrent use.
//
// BIOLOGICAL CONTEXT:
// Coincidence detection is a fundamental computational capability of biological neurons.
// Many neurons act as "coincidence detectors," firing preferentially when they receive
// multiple excitatory inputs within a narrow time frame. This is crucial for:
//
//  1. TEMPORAL SUMMATION: Near-simultaneous excitatory postsynaptic potentials (EPSPs)
//     sum together more effectively to depolarize the membrane and reach the firing
//     threshold. Inputs that are too far apart in time decay before they can summate.
//
//  2. NMDA RECEPTOR ACTIVATION: This is a key molecular mechanism for coincidence
//     detection. NMDA receptors require two conditions to be met simultaneously:
//     - The binding of glutamate (the signal from a presynaptic neuron).
//     - Sufficient postsynaptic membrane depolarization (often from other coincident inputs)
//     to expel a magnesium ion (Mg2+) that blocks the receptor's channel.
//     This function models the outcome of this process: detecting correlated inputs.
//
//  3. FEATURE BINDING: In sensory systems, coincidence detection allows neurons to
//     bind together different features of a stimulus. For example, a neuron might
//     only fire when it receives simultaneous inputs representing a vertical edge
//     and a specific color, thus detecting a "vertical red line."
//
//  4. SYNAPTIC PLASTICITY: The Hebbian principle ("cells that fire together, wire
//     together") relies on detecting coincident pre- and post-synaptic activity.
//     Detecting coincident inputs is the first step in this process.
//
// The exclusion of inhibitory inputs (like 'feedback') is also biologically realistic,
// as the goal of this mechanism is to detect a convergence of *excitatory* drive.
func (n *Neuron) DetectCoincidentInputs() int {
	// Coincidence detection must be explicitly enabled in the neuron's configuration.
	if !n.EnableCoincidenceDetection {
		return 0
	}

	n.inputActivityMutex.RLock()
	defer n.inputActivityMutex.RUnlock()

	now := time.Now()
	// Define the temporal window for coincidence. This models the integration
	// timescale of the neuron's membrane and the kinetics of its synaptic receptors.
	// Biologically, this is often in the range of 5-20 milliseconds.
	cutoff := now.Add(-n.CoincidenceWindow)

	uniqueSources := make(map[string]bool)

	for sourceID, activities := range n.inputActivityHistory {
		// Models the specific exclusion of inhibitory circuits from the summation
		// process for triggering a coincident-driven spike. For example, feedback
		// inhibition serves to regulate activity, not contribute to excitatory summation.

		for _, activity := range activities {
			// Check for two conditions modeling an effective excitatory input:
			// 1. activity.Timestamp.After(cutoff): The input must be recent enough
			//    to contribute to temporal summation (i.e., within the coincidence window).
			// 2. activity.EffectiveValue > 0: The input must be excitatory (a positive EPSP).
			if activity.Timestamp.After(cutoff) && activity.EffectiveValue > 0 {
				// By adding the source ID to a map, we count each upstream neuron
				// only once, even if it fired a burst of spikes within the window.
				// This detects how many *different* sources are active together.
				uniqueSources[sourceID] = true
				break // Only count each source once within the window
			}
		}
	}

	// The number of unique sources represents the degree of correlated input.
	// A higher number indicates a stronger temporal correlation, which would
	// significantly increase the firing probability of a real neuron.
	return len(uniqueSources)
}

func (n *Neuron) DetectCoincidentInputsRelativeToMostRecent() int {
	if !n.EnableCoincidenceDetection {
		return 0
	}

	n.inputActivityMutex.RLock()
	defer n.inputActivityMutex.RUnlock()

	if n.inputActivityHistory == nil || len(n.inputActivityHistory) == 0 {
		return 0
	}

	// Find the most recent input across all sources
	var mostRecentTime time.Time
	for _, activities := range n.inputActivityHistory {
		for _, activity := range activities {
			if activity.Timestamp.After(mostRecentTime) {
				mostRecentTime = activity.Timestamp
			}
		}
	}

	if mostRecentTime.IsZero() {
		return 0
	}

	// Count inputs within the window relative to the most recent input
	cutoff := mostRecentTime.Add(-n.CoincidenceWindow)
	uniqueSources := make(map[string]bool)

	for sourceID, activities := range n.inputActivityHistory {
		for _, activity := range activities {
			if activity.Timestamp.After(cutoff) && activity.EffectiveValue > 0 {
				uniqueSources[sourceID] = true
				break
			}
		}
	}

	return len(uniqueSources)
}

// scheduleDelayedDelivery queues a message for delivery after total propagation delay.
// This method models the biological axon with both synaptic and spatial delay components.
//
// BIOLOGICAL PROCESS MODELED:
// When a neuron fires, the action potential must propagate down the axon to reach
// target synapses. The total delay includes both synaptic processing time
// (neurotransmitter release, receptor binding) and spatial propagation time
// (axonal conduction velocity, distance, myelination).
//
// ARCHITECTURE:
// Each neuron has exactly one delivery worker goroutine (created on first use)
// that handles all delayed deliveries. This models the biological reality that
// each neuron has one axon, eliminating the goroutine explosion that occurred
// with per-message goroutines.
//
// Parameters:
//
//	message: The synaptic message to deliver (includes timing and source info)
//	target: The post-synaptic neuron to receive the message
//	delay: Total delay including synaptic and spatial components
func (n *Neuron) ScheduleDelayedDelivery(message synapse.SynapseMessage, target synapse.SynapseCompatibleNeuron, delay time.Duration) {
	// Initialize delivery system on first use (exactly once per neuron)
	// Models biological axon development and maturation
	n.deliveryWorker.Do(func() {
		n.deliveryQueue = make(chan delayedMessage, 100)
		go n.axonDeliveryWorker()
	})

	// Create delivery entry with precise timing
	// Models action potential propagation with calculated arrival time
	delayedMsg := delayedMessage{
		message:      message,
		target:       target,
		deliveryTime: time.Now().Add(delay),
	}

	// Queue for axonal delivery (non-blocking)
	select {
	case n.deliveryQueue <- delayedMsg:
		// Successfully queued for delivery via axonal propagation
	default:
		// Delivery queue full - immediate delivery fallback
		// Models graceful degradation under extreme network load
		target.Receive(message)
	}
}

// axonDeliveryWorker handles all delayed message delivery for this neuron.
// This goroutine models the biological axon: it manages precise timing of
// action potential arrival at target synapses after propagation delays.
//
// BIOLOGICAL CORRESPONDENCE:
// Real neurons generate action potentials at the axon hillock, which then
// propagate down all axon branches to reach target synapses. Different targets
// receive the signal at different times based on distance, axon diameter,
// and myelination. This worker models that propagation and delivery process.
//
// PERFORMANCE BENEFITS:
// - Fixed resource usage: N neurons = N delivery goroutines (not millions)
// - Precise timing: 100μs resolution for biological realism
// - Efficient batching: Multiple messages delivered per cycle
// - Graceful scaling: Performance predictable regardless of activity level
func (n *Neuron) axonDeliveryWorker() {
	// Pending deliveries sorted by delivery time
	// Pre-allocate to avoid memory allocation overhead
	pending := make([]delayedMessage, 0, 100)

	// High-frequency timer for precise delivery timing
	// 100μs provides excellent temporal resolution for biological realism
	// while remaining computationally efficient
	ticker := time.NewTicker(100 * time.Microsecond)
	defer ticker.Stop()

	for {
		select {
		// New message queued for delayed delivery
		case msg := <-n.deliveryQueue:
			pending = append(pending, msg)

		// Check for messages ready for delivery
		case <-ticker.C:
			if len(pending) == 0 {
				continue // No pending deliveries
			}

			now := time.Now()
			remaining := pending[:0] // Reuse slice capacity efficiently

			// Deliver all messages whose delivery time has arrived
			for _, msg := range pending {
				if now.After(msg.deliveryTime) || now.Equal(msg.deliveryTime) {
					// Delivery time reached - transmit to target neuron
					// Models action potential arrival at target synapse
					msg.target.Receive(msg.message)
				} else {
					// Not yet time - keep in pending list for future delivery
					remaining = append(remaining, msg)
				}
			}

			// Update pending list (removes delivered messages)
			pending = remaining

		// Shutdown signal - clean termination of delivery worker
		case <-n.ctx.Done():
			return
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
// mustBeLocked: true (stateMutex must be held by caller)
func (n *Neuron) applyPostSynapticGainUnsafe(msg synapse.SynapseMessage) float64 {
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
	// Final signal = synaptic_strength × post-synaptic_receptor_sensitivity
	return msg.Value * gain
}

// registerInputSourceForScaling registers a new input source for synaptic scaling
// This method ensures that all active input sources have corresponding synaptic gains
//
// BIOLOGICAL CONTEXT:
// When a post-synaptic neuron receives input from a new source, it needs to
// establish receptor sensitivity (gain) for that synapse. Initially, the gain
// is set to 1.0 (normal sensitivity), but it will be adjusted by synaptic scaling
// to maintain optimal total input strength.
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

// fireUnsafe is the internal firing method called when state lock is already held
// Includes refractory period enforcement and biological timing constraints
//
// Biological process modeled:
// 1. Check if neuron is in refractory period (cannot fire if recent firing occurred)
// 2. If firing is allowed, generate action potential
// 3. Record firing time to enforce future refractory periods
// 4. Update homeostatic state (calcium accumulation, firing history)
// 5. Propagate signal to all synaptic connections with timing information
//
// The refractory period models the biological reality that after an action potential,
// voltage-gated sodium channels become inactivated and require time to recover.
// During this period, no amount of input can trigger another action potential.
//
// mustBeLocked: true (stateMutex must be held by caller)
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

	// Get snapshot of output synapses (minimal locking since we're already protected)
	n.outputsMutex.RLock()
	synapsesCopy := make(map[string]synapse.SynapticProcessor, len(n.outputSynapses))
	for id, synapseProcessor := range n.outputSynapses {
		synapsesCopy[id] = synapseProcessor
	}
	n.outputsMutex.RUnlock()

	// Parallel transmission to all synapses
	// Models: action potential propagating simultaneously down all axon branches
	// Transmit() is already non-blocking because it uses time.AfterFunc.
	for _, synapseProcessor := range synapsesCopy {
		synapseProcessor.Transmit(outputValue)
	}
}

// resetAccumulatorUnsafe clears integration state (internal use when locked)
// Returns the neuron to its resting state, ready for new signal integration
//
// mustBeLocked: true (stateMutex must be held by caller)
func (n *Neuron) resetAccumulatorUnsafe() {
	n.accumulator = 0
	// Models the neuron returning to resting potential
}

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
// 7. Operates on slower timescale than synaptic learning (minutes vs milliseconds)
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

	for sourceID := range n.inputGains {
		// Get actual recent activity for this source
		activities, hasActivity := n.inputActivityHistory[sourceID]
		if !hasActivity || len(activities) == 0 {
			continue // Skip sources with no recent activity
		}

		// Calculate average recent activity (biological integration)
		activitySum := 0.0
		for _, activity := range activities {
			activitySum += math.Abs(activity.EffectiveValue) // Use absolute value
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
	scalingFactor := math.Max(n.scalingConfig.MinScalingFactor, math.Min(n.scalingConfig.MaxScalingFactor, rawScalingFactor))

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
}

// ID returns the unique identifier of this neuron
// This implements the synapse.SynapseCompatibleNeuron interface
func (n *Neuron) ID() string {
	return n.id
}

// Close gracefully shuts down the neuron by closing its input channel
// This signals the Run() goroutine to exit cleanly
// Models: neural death or experimental disconnection
func (n *Neuron) Close() {
	n.closeOnce.Do(func() {
		// Signal the Run() goroutine to exit.
		n.cancel()
		// Wait for the Run() goroutine to complete its shutdown.
		n.wg.Wait()
	})
}

// GetOutputCount returns the number of output connections (backward compatibility)
func (n *Neuron) GetOutputCount() int {
	return n.GetOutputSynapseCount()
}

// ============================================================================
// HOMEOSTATIC AND THRESHOLD MANAGEMENT METHODS
// ============================================================================

// GetCurrentThreshold returns the current firing threshold (thread-safe)
// The threshold may be different from the base threshold due to homeostatic adjustments
func (n *Neuron) GetCurrentThreshold() float64 {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	return n.threshold
}

// GetBaseThreshold returns the original threshold before homeostatic adjustments
// This value never changes and represents the neuron's initial excitability
func (n *Neuron) GetBaseThreshold() float64 {
	return n.baseThreshold // Immutable, no lock needed
}

// GetCalciumLevel returns the current calcium concentration (thread-safe)
// Calcium level indicates recent firing activity and drives homeostatic regulation
func (n *Neuron) GetCalciumLevel() float64 {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	return n.homeostatic.calciumLevel
}

// GetCurrentFiringRate calculates the current firing rate based on recent history
// Returns the firing rate in Hz based on the configured activity window
func (n *Neuron) GetCurrentFiringRate() float64 {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	return n.calculateCurrentFiringRateUnsafe()
}

// GetHomeostaticInfo returns a snapshot of homeostatic state (thread-safe)
// Returns a copy of internal data to prevent external modification
func (n *Neuron) GetHomeostaticInfo() HomeostaticInfo {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	// Create a copy of firing history to prevent external modification
	historyCopy := make([]time.Time, len(n.homeostatic.firingHistory))
	copy(historyCopy, n.homeostatic.firingHistory)

	return HomeostaticInfo{
		targetFiringRate:      n.homeostatic.targetFiringRate,
		homeostasisStrength:   n.homeostatic.homeostasisStrength,
		calciumLevel:          n.homeostatic.calciumLevel,
		firingHistory:         historyCopy,
		minThreshold:          n.homeostatic.minThreshold,
		maxThreshold:          n.homeostatic.maxThreshold,
		activityWindow:        n.homeostatic.activityWindow,
		lastHomeostaticUpdate: n.homeostatic.lastHomeostaticUpdate,
	}
}

// SetHomeostaticParameters updates homeostatic regulation parameters (thread-safe)
// This allows dynamic adjustment of self-regulation behavior
//
// Parameters:
// targetFiringRate: desired firing rate in Hz (0 disables homeostasis)
// homeostasisStrength: adjustment strength 0.0-1.0 (0 disables)
//
// Setting either parameter to 0 disables homeostatic regulation
// When disabled, threshold resets to base value
func (n *Neuron) SetHomeostaticParameters(targetFiringRate, homeostasisStrength float64) {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	n.homeostatic.targetFiringRate = targetFiringRate
	n.homeostatic.homeostasisStrength = homeostasisStrength

	// If homeostasis is being disabled, reset threshold to base value
	if targetFiringRate == 0.0 || homeostasisStrength == 0.0 {
		n.threshold = n.baseThreshold
		n.homeostatic.calciumLevel = 0.0
		n.homeostatic.firingHistory = n.homeostatic.firingHistory[:0] // Clear history
	}
}

// ============================================================================
// SYNAPTIC SCALING METHODS
// ============================================================================

// EnableSynapticScaling enables the synaptic scaling mechanism with specified parameters
// This should be called before starting the neuron's Run() method
//
// Parameters:
// targetStrength: desired average effective input strength
// scalingRate: rate of adjustment per scaling event (0.001-0.01 typical)
// interval: time between scaling operations (30s-10min typical)
func (n *Neuron) EnableSynapticScaling(targetStrength, scalingRate float64, interval time.Duration) {
	n.scalingConfig.Enabled = true
	n.scalingConfig.TargetInputStrength = targetStrength
	n.scalingConfig.ScalingRate = scalingRate
	n.scalingConfig.ScalingInterval = interval
	n.scalingConfig.LastScalingUpdate = time.Now()
}

// DisableSynapticScaling turns off synaptic scaling
// Existing input gains are preserved but no further scaling occurs
func (n *Neuron) DisableSynapticScaling() {
	n.scalingConfig.Enabled = false
}

// GetInputGains returns a copy of current input gains for monitoring (thread-safe)
// Returns map[sourceID]gain where gain is the receptor sensitivity multiplier
func (n *Neuron) GetInputGains() map[string]float64 {
	n.inputGainsMutex.RLock()
	defer n.inputGainsMutex.RUnlock()

	// Return a copy to prevent external modification
	gains := make(map[string]float64, len(n.inputGains))
	for sourceID, gain := range n.inputGains {
		gains[sourceID] = gain
	}
	return gains
}

// GetScalingHistory returns recent scaling factors for analysis (thread-safe)
// Returns a copy of the scaling history to prevent external modification
func (n *Neuron) GetScalingHistory() []float64 {
	history := make([]float64, len(n.scalingConfig.ScalingHistory))
	copy(history, n.scalingConfig.ScalingHistory)
	return history
}

// SetDendriticIntegrationMode allows swapping the neuron's input processing strategy.
// This enables dynamic changes to a neuron's computational behavior.
func (n *Neuron) SetDendriticIntegrationMode(mode DendriticIntegrationMode) {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	n.dendriticIntegrationMode = mode
}

// SetInputGain manually sets the receptor gain for a specific input source
// This allows external control of synaptic scaling for experimental purposes
//
// Parameters:
// sourceID: identifier of the input source
// gain: receptor sensitivity multiplier (typically 0.1-10.0)
func (n *Neuron) SetInputGain(sourceID string, gain float64) {
	// Clamp gain to reasonable biological bounds
	if gain < 0.01 {
		gain = 0.01
	}
	if gain > 10.0 {
		gain = 10.0
	}

	n.inputGainsMutex.Lock()
	defer n.inputGainsMutex.Unlock()

	if n.inputGains == nil {
		n.inputGains = make(map[string]float64)
	}
	n.inputGains[sourceID] = gain
}

// ============================================================================
// RECEIVE METHOD
// ============================================================================

// Receive accepts a synapse message and queues it for processing.
// This method implements biological dendritic reception with realistic
// capacity constraints and refractory period enforcement.
//
// BIOLOGICAL PROCESS MODELED:
// In real neurons, synaptic inputs arrive at dendrites and create postsynaptic
// potentials (PSPs). This method models dendritic reception with finite capacity:
// when dendrites are saturated or the neuron is in refractory period, additional
// inputs are lost, which is biologically realistic.
//
// BUFFERED ARCHITECTURE:
// All inputs are queued in a bounded buffer for processing by the main Run() loop.
// This maintains single-threaded processing within each neuron while allowing
// concurrent input delivery from multiple source neurons. The small buffer size
// (10 messages) reflects biological dendritic capacity limitations.
//
// Parameters:
//
//	msg: SynapseMessage containing synaptic signal with timing and source information
//
// The method is thread-safe and non-blocking. Refractory period and buffer overflow
// result in message loss, modeling realistic biological constraints.
func (n *Neuron) Receive(msg synapse.SynapseMessage) {
	// BIOLOGICAL CONSTRAINT: Refractory period enforcement
	// During refractory period, voltage-gated sodium channels are inactivated
	// and the neuron cannot respond to inputs regardless of strength.
	// This models the absolute refractory period in biological neurons.
	if !n.lastFireTime.IsZero() && time.Since(n.lastFireTime) < n.refractoryPeriod {
		// Message lost during refractory period - biologically accurate
		return
	}

	// SIMPLIFIED INPUT HANDLING: Always use buffer approach
	// This eliminates complexity while maintaining biological realism.
	// Maximum 1ms delay from buffering is negligible compared to biological
	// neural processing timescales (10-50ms membrane time constants).
	select {
	case n.inputBuffer <- msg:
		// Successfully queued for processing in next computational cycle
	default:
		// Buffer full - input lost due to dendritic saturation
		// This models realistic biological behavior under high input load
	}
}

// ============================================================================
// METHODS FOR TEST, MONITORING AND OBSERVATION
// ============================================================================

// ProcessTestMessage is a test helper that processes a message and ensures
// activity tracking is properly updated for testing coincidence detection
// This method is specifically designed for testing scenarios
func (n *Neuron) ProcessTestMessage(msg synapse.SynapseMessage) {
	// Process the message through normal pathways
	n.Receive(msg)

	// Additional processing time to ensure message is fully integrated
	time.Sleep(1 * time.Millisecond)
}

// GetAccumulator returns the current accumulator value for testing/debugging
func (n *Neuron) GetAccumulator() float64 {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	return n.accumulator
}

// ResetAccumulator clears the accumulator for testing purposes
func (n *Neuron) ResetAccumulator() {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()
	n.accumulator = 0.0
}

// GetNeuronState returns comprehensive neuron state for debugging
func (n *Neuron) GetNeuronState() map[string]interface{} {
	n.stateMutex.Lock()
	defer n.stateMutex.Unlock()

	return map[string]interface{}{
		"id":                n.id,
		"accumulator":       n.accumulator,
		"threshold":         n.threshold,
		"baseThreshold":     n.baseThreshold,
		"lastFireTime":      n.lastFireTime,
		"calciumLevel":      n.homeostatic.calciumLevel,
		"currentFiringRate": n.calculateCurrentFiringRateUnsafe(),
		"refractoryPeriod":  n.refractoryPeriod,
		"decayRate":         n.decayRate,
	}
}

// WaitForQuiescence waits for the neuron to reach a stable state
func (n *Neuron) WaitForQuiescence(timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		acc := n.GetAccumulator()
		if acc < 0.001 { // Close to zero
			return true
		}
		time.Sleep(1 * time.Millisecond)
	}
	return false
}

// GetInputActivityHistory returns a copy of the input activity history for testing
// This allows tests to verify that activity tracking is working correctly
func (n *Neuron) GetInputActivityHistory() map[string][]InputActivity {
	n.inputActivityMutex.RLock()
	defer n.inputActivityMutex.RUnlock()

	// Return a copy to prevent external modification
	history := make(map[string][]InputActivity)
	for sourceID, activities := range n.inputActivityHistory {
		historyCopy := make([]InputActivity, len(activities))
		copy(historyCopy, activities)
		history[sourceID] = historyCopy
	}
	return history
}
