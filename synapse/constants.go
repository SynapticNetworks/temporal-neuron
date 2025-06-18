/*
=================================================================================
BIOLOGICAL CONSTANTS AND PARAMETERS
=================================================================================

This file contains all biological constants, limits, and parameters used
throughout the synapse system. Each constant is documented with its biological
basis, experimental source, and typical range observed in real neural tissue.

ORGANIZATION:
1. Vesicle Dynamics Constants
2. STDP and Plasticity Constants
3. Synaptic Transmission Constants
4. Structural Plasticity Constants
5. Neurotransmitter-Specific Constants
6. Performance and Safety Limits

EXPERIMENTAL BASIS:
All constants are derived from published neuroscience research with references
provided. Values represent typical ranges observed in mammalian cortical
synapses unless otherwise specified.

UNITS:
- Time: Consistently in Go time.Duration (nanoseconds internally)
- Concentration: Micromolar (μM)
- Distance: Micrometers (μm)
- Frequency: Hertz (Hz)
- Weight: Dimensionless (0.0-∞, typically 0.0-2.0)
=================================================================================
*/

package synapse

import "time"

// =================================================================================
// VESICLE DYNAMICS CONSTANTS
// =================================================================================

// === VESICLE POOL SIZES ===
// Based on electron microscopy studies of synaptic terminals
// References: Schikorski & Stevens (1997), Rosenmund & Stevens (1996)

const (
	// DEFAULT_READY_POOL_SIZE represents the Ready Releasable Pool (RRP)
	// The number of vesicles docked at the active zone, primed for immediate release
	// Biological range: 5-20 vesicles in typical cortical synapses
	// Experimental basis: Patch-clamp capacitance measurements
	DEFAULT_READY_POOL_SIZE = 15

	// DEFAULT_RECYCLING_POOL_SIZE represents the Recycling Pool
	// Vesicles that can be mobilized within seconds during sustained activity
	// Biological range: 100-200 vesicles in cortical terminals
	// Function: Maintains transmission during moderate frequency stimulation (10-50 Hz)
	DEFAULT_RECYCLING_POOL_SIZE = 150

	// DEFAULT_RESERVE_POOL_SIZE represents the Reserve Pool
	// Large pool of vesicles mobilized only during intense, prolonged activity
	// Biological range: 1000+ vesicles in large cortical terminals
	// Function: Sustains transmission during high-frequency trains (>100 Hz)
	DEFAULT_RESERVE_POOL_SIZE = 1000
)

// === VESICLE RECYCLING TIME CONSTANTS ===
// Based on fluorescence recovery and electron microscopy studies
// References: Ryan et al. (1993), Smith & Neher (1997), Fernández-Alfonso & Ryan (2004)

const (
	// FAST_RECYCLING_TIME represents kiss-and-run endocytosis
	// Fastest vesicle recycling pathway preserving vesicle identity
	// Biological process: Transient fusion pore, rapid resealing
	// Experimental measurement: ~1-3 seconds in hippocampal synapses
	// Functional significance: Maintains transmission during brief high activity
	FAST_RECYCLING_TIME = 2 * time.Second

	// SLOW_RECYCLING_TIME represents clathrin-mediated endocytosis
	// Complete vesicle retrieval and reformation pathway
	// Biological process: Full vesicle internalization, reformation from endosomes
	// Experimental measurement: 10-30 seconds in cortical synapses
	// Functional significance: Complete vesicle regeneration and refilling
	SLOW_RECYCLING_TIME = 20 * time.Second

	// VESICLE_REFILL_TIME represents neurotransmitter loading time
	// Time required to load vesicles with neurotransmitter via transporters
	// Biological process: V-ATPase acidification + neurotransmitter transport
	// Rate limiting step: Often transporter activity (VGLUT, VGAT, etc.)
	// Experimental basis: Vesicular transport kinetics studies
	VESICLE_REFILL_TIME = 8 * time.Second

	// REPRIMING_TIME represents release machinery reset
	// Time to reassemble SNARE complexes and calcium sensors
	// Biological process: SNARE complex formation, Munc13 priming
	// Experimental measurement: 2-5 seconds from patch-clamp studies
	// Functional significance: Determines maximum sustainable firing rate
	REPRIMING_TIME = 3 * time.Second
)

// === RELEASE PROBABILITY PARAMETERS ===
// Based on quantal analysis and calcium uncaging experiments
// References: Rosenmund et al. (1993), Bollmann et al. (2000)

const (
	// BASELINE_RELEASE_PROBABILITY represents resting release probability
	// Probability of vesicle fusion per action potential at rest
	// Biological range: 0.1-0.5 depending on synapse type and development
	// Experimental measurement: Quantal analysis in paired recordings
	// Modulation: Increases with calcium, decreases with depression
	BASELINE_RELEASE_PROBABILITY = 0.25

	// MAX_CALCIUM_ENHANCEMENT represents maximum calcium-dependent boost
	// Factor by which high calcium can increase release probability
	// Biological basis: Cooperative calcium binding to synaptotagmin
	// Experimental range: 2-5x increase during high-frequency stimulation
	// Saturation: Limited by calcium buffer capacity and cooperative binding
	MAX_CALCIUM_ENHANCEMENT = 3.0

	// CALCIUM_COOPERATIVITY represents calcium binding cooperativity
	// Hill coefficient for calcium-dependent release enhancement
	// Biological basis: Multiple calcium binding sites on synaptotagmin
	// Experimental measurement: ~4-5 from calcium uncaging experiments
	// Functional significance: Creates steep calcium-release relationship
	CALCIUM_COOPERATIVITY = 4.0
)

// === METABOLIC RATE LIMITS ===
// Based on neurotransmitter synthesis and metabolic studies
// References: Chaudhry et al. (1995), Rothman et al. (2003)

const (
	// GLUTAMATE_MAX_RATE represents maximum glutamate release rate
	// Limited by vesicle recycling and glutamine-glutamate cycle
	// Biological constraint: Vesicle pool depletion at high frequencies
	// Experimental observation: ~50-100 Hz sustained in cortical synapses
	// Metabolic basis: Glutamine synthetase in astrocytes rate-limiting
	GLUTAMATE_MAX_RATE = 50.0 // Hz

	// GABA_MAX_RATE represents maximum GABA release rate
	// Generally higher than glutamate due to interneuron specialization
	// Biological context: Fast-spiking interneurons can sustain 100+ Hz
	// Experimental evidence: Parvalbumin+ interneurons in gamma oscillations
	// Metabolic basis: Efficient GABA synthesis via GAD enzymes
	GABA_MAX_RATE = 80.0 // Hz

	// DOPAMINE_MAX_RATE represents maximum dopamine release rate
	// Much lower due to synthesis limitations and neuromodulator function
	// Biological constraint: Tyrosine hydroxylase rate-limiting enzyme
	// Experimental observation: Burst firing ~10-20 Hz in VTA/SNc neurons
	// Functional significance: Phasic vs tonic dopamine signaling
	DOPAMINE_MAX_RATE = 10.0 // Hz

	// SEROTONIN_MAX_RATE represents maximum serotonin release rate
	// Lowest rate due to synthesis constraints and widespread projections
	// Biological constraint: Tryptophan hydroxylase availability
	// Experimental observation: Raphe neurons fire 0.5-5 Hz typically
	// Functional significance: Slow, global neuromodulatory signaling
	SEROTONIN_MAX_RATE = 3.0 // Hz

	// ACETYLCHOLINE_MAX_RATE represents mixed cholinergic signaling
	// Intermediate rate supporting both fast synaptic and slow modulatory
	// Biological context: Both nicotinic (fast) and muscarinic (slow) signaling
	// Experimental range: 10-50 Hz depending on target and function
	// Metabolic basis: Choline acetyltransferase and choline availability
	ACETYLCHOLINE_MAX_RATE = 25.0 // Hz
)

// =================================================================================
// STDP AND PLASTICITY CONSTANTS
// =================================================================================

// === CLASSICAL STDP PARAMETERS ===
// Based on seminal STDP experiments in cortical and hippocampal slices
// References: Bi & Poo (1998), Sjöström et al. (2001), Caporale & Dan (2008)

const (
	// STDP_TIME_CONSTANT represents the exponential decay of plasticity
	// Standard τ for cortical excitatory synapses
	// Biological basis: NMDA receptor activation kinetics and calcium dynamics
	// Experimental range: 10-30 ms in cortical pyramidal neurons
	// Temperature dependence: Q10 ~2-3, values at physiological temperature
	STDP_TIME_CONSTANT = 20 * time.Millisecond

	// STDP_LEARNING_RATE represents percentage weight change per spike pair
	// Typical learning rate for single spike pairings
	// Biological constraint: Prevents runaway potentiation/depression
	// Experimental range: 0.5-5% per pairing in cortical slice preparations
	// Modulation: Can be increased by neuromodulators (dopamine, acetylcholine)
	STDP_LEARNING_RATE = 0.01 // 1% per pairing

	// STDP_WINDOW_SIZE represents maximum timing window for plasticity
	// Beyond this window, spike pairs don't induce plasticity
	// Biological basis: NMDA receptor activation window and calcium signaling
	// Experimental measurement: ±100 ms in most cortical preparations
	// Variability: Can be narrower (~20 ms) in some interneuron types
	STDP_WINDOW_SIZE = 100 * time.Millisecond

	// STDP_ASYMMETRY_RATIO represents LTD/LTP amplitude ratio
	// Classic asymmetric STDP window with stronger depression
	// Biological significance: Prevents runaway potentiation, ensures stability
	// Experimental range: 1.1-1.5 in most cortical excitatory synapses
	// Functional importance: Balances potentiation with homeostatic control
	STDP_ASYMMETRY_RATIO = 1.2

	// STDP_SIMULTANEOUS_THRESHOLD for treating spikes as simultaneous
	// Biological reality: Perfect simultaneity is rare, need tolerance window
	// Experimental basis: Temporal resolution of slice electrophysiology
	// Typical value: ±1ms for practical simultaneity detection
	STDP_SIMULTANEOUS_THRESHOLD = 1 * time.Millisecond

	// STDP_MIN_WEIGHT represents minimum allowed synaptic strength
	// Prevents complete synapse elimination while allowing weakening
	// Biological basis: Even "silent" synapses retain some AMPA receptors
	// Experimental evidence: Minimal synaptic responses in LTD studies
	// Functional significance: Maintains potential for future strengthening
	STDP_MIN_WEIGHT = 0.001

	// STDP_MAX_WEIGHT represents maximum synaptic strength
	// Prevents pathological over-strengthening of individual synapses
	// Biological basis: Physical limits of receptor density and active zone size
	// Experimental observation: 2-3x baseline in maximal LTP studies
	// Homeostatic significance: Forces competition between synapses
	STDP_MAX_WEIGHT = 2.0
)

// === METAPLASTICITY CONSTANTS ===
// Based on studies of plasticity of plasticity (Bienenstock-Cooper-Munro rule)
// References: Bienenstock et al. (1982), Abraham & Bear (1996)

const (
	// METAPLASTICITY_RATE represents rate of plasticity threshold changes
	// Speed at which synaptic modification threshold adapts
	// Biological basis: CaMKII autophosphorylation and homeostatic mechanisms
	// Experimental evidence: Hours-to-days timescale in slice studies
	// Functional significance: Prevents saturation, maintains dynamic range
	METAPLASTICITY_RATE = 0.1

	// HETEROSYNAPTIC_RANGE for spread of plasticity to nearby synapses
	// Biological observation: Plasticity can spread 10-50μm from activated synapse
	// Molecular basis: Diffusion of calcium, nitric oxide, or protein synthesis factors
	// Experimental evidence: Heterosynaptic LTD in hippocampal slice preparations
	// Functional significance: Allows cooperative learning across synapse clusters
	HETEROSYNAPTIC_RANGE = 20.0 // micrometers

	// COOPERATIVITY_THRESHOLD represents minimum inputs for plasticity
	// Number of concurrent inputs required for reliable plasticity induction
	// Biological basis: NMDA receptor voltage dependence and spatial clustering
	// Experimental range: 2-5 synapses in dendritic spine clusters
	// Functional significance: Ensures plasticity only for correlated inputs
	COOPERATIVITY_THRESHOLD = 3

	// === COOPERATIVITY ENHANCEMENT PARAMETERS ===
	// These constants define how plasticity magnitude scales with increasing cooperative inputs.
	// References: Sjöström et al. (2001)

	// BIOLOGY_HIGH_COOPERATIVITY_ENHANCEMENT_FACTOR represents the maximum fold enhancement
	// of plasticity due to high cooperative input levels, relative to threshold cooperativity.
	// Biological basis: Enhanced NMDA receptor activation and calcium signaling with strong coincident input.
	// Experimental range: 1.5-5x, depending on synapse type and preparation.
	BIOLOGY_HIGH_COOPERATIVITY_ENHANCEMENT_FACTOR = 3.0

	// BIOLOGY_COOPERATIVITY_HALF_SATURATION determines the cooperativity level at which
	// half of the maximum enhancement is achieved.
	// Biological basis: Sigmoidal or saturating response of plasticity mechanisms to input cooperativity.
	// Functional significance: Introduces a non-linear gain for stronger coincident activity.
	BIOLOGY_COOPERATIVITY_HALF_SATURATION = 7.0 // Inputs beyond threshold

	// FREQUENCY_DEPENDENCE_THRESHOLD represents transition frequency
	// Frequency above which plasticity rules change (LTD → LTP)
	// Biological basis: Calcium concentration and signaling pathway switch
	// Experimental measurement: ~10-20 Hz in hippocampal CA1 synapses
	// Molecular basis: CaMKII vs calcineurin activation thresholds
	FREQUENCY_DEPENDENCE_THRESHOLD = 15.0 // Hz
)

// =================================================================================
// SYNAPTIC TRANSMISSION CONSTANTS
// =================================================================================

// === TRANSMISSION DELAYS ===
// Based on axonal conduction and synaptic processing measurements
// References: Sabatini & Regehr (1996), Bollmann et al. (2000)

const (
	// MIN_SYNAPTIC_DELAY represents fastest possible synaptic transmission
	// Time from presynaptic action potential to postsynaptic response onset
	// Biological components: Calcium influx + vesicle fusion + diffusion
	// Experimental minimum: ~0.3-0.5 ms in fast central synapses
	// Temperature dependence: Faster at physiological vs room temperature
	MIN_SYNAPTIC_DELAY = 500 * time.Microsecond

	// TYPICAL_SYNAPTIC_DELAY represents average cortical synaptic delay
	// Standard delay for most cortical excitatory synapses
	// Biological variation: Depends on active zone organization and calcium channels
	// Experimental range: 1-3 ms in cortical slice preparations
	// Functional significance: Affects temporal precision of neural coding
	TYPICAL_SYNAPTIC_DELAY = 2 * time.Millisecond

	// MAX_SYNAPTIC_DELAY represents slowest synaptic transmission
	// Upper limit for still-functional synaptic transmission
	// Biological context: May occur in developing or pathological synapses
	// Experimental observation: >10 ms indicates synaptic dysfunction
	// Safety limit: Prevents unrealistic timing in simulations
	MAX_SYNAPTIC_DELAY = 50 * time.Millisecond
)

// === SIGNAL PROCESSING CONSTANTS ===
// Based on postsynaptic potential measurements and integration studies

const (
	// SIGNAL_NOISE_FLOOR represents minimum detectable signal
	// Smallest synaptic response that affects postsynaptic neuron
	// Biological basis: Membrane noise and receptor sensitivity
	// Experimental measurement: ~0.1-0.5 mV PSP amplitude
	// Functional significance: Determines effective connectivity
	SIGNAL_NOISE_FLOOR = 0.001

	// MAX_SIGNAL_AMPLITUDE represents maximum single synapse response
	// Upper limit for individual synaptic strength
	// Biological constraint: Active zone size and receptor saturation
	// Experimental observation: ~5-10 mV maximum PSP in cortical neurons
	// Safety significance: Prevents single synapses from dominating
	MAX_SIGNAL_AMPLITUDE = 10.0

	// SIGNAL_INTEGRATION_WINDOW represents temporal summation window
	// Time window for temporal integration of synaptic inputs
	// Biological basis: Membrane time constant and dendritic filtering
	// Experimental range: 10-50 ms depending on neuron type
	// Computational significance: Affects coincidence detection
	SIGNAL_INTEGRATION_WINDOW = 20 * time.Millisecond
)

// =================================================================================
// STRUCTURAL PLASTICITY CONSTANTS
// =================================================================================

// === PRUNING PARAMETERS ===
// Based on developmental pruning studies and adult spine turnover
// References: Holtmaat & Svoboda (2009), Yang et al. (2009)

const (
	// PRUNING_PROTECTION_PERIOD represents grace period for new synapses
	// Time after synapse formation before pruning can occur
	// Biological rationale: Allows time for synapses to demonstrate utility
	// Experimental evidence: Hours-to-days protection in development
	// Simulation practicality: Scaled to reasonable simulation timescales
	PRUNING_PROTECTION_PERIOD = 30 * time.Second

	// PRUNING_WEIGHT_THRESHOLD represents weakness threshold for elimination
	// Synaptic strength below which pruning becomes likely
	// Biological basis: Metabolic cost vs functional benefit trade-off
	// Experimental correlation: Weak synapses have higher turnover rates
	// Typical value: ~1-5% of maximum synaptic strength
	PRUNING_WEIGHT_THRESHOLD = 0.01

	// PRUNING_INACTIVITY_THRESHOLD represents maximum tolerable inactivity
	// Duration of inactivity before synapse becomes pruning candidate
	// Biological principle: "Use it or lose it" synaptic maintenance
	// Experimental range: Hours-to-days in spine imaging studies
	// Simulation scaling: Compressed for practical simulation timescales
	PRUNING_INACTIVITY_THRESHOLD = 5 * time.Minute

	// PRUNING_PROBABILITY represents stochastic pruning rate
	// Probability of pruning per evaluation for eligible synapses
	// Biological realism: Pruning is probabilistic, not deterministic
	// Experimental basis: Variable spine elimination rates
	// Functional significance: Maintains network diversity and exploration
	PRUNING_PROBABILITY = 0.1 // 10% per evaluation

	// PRUNING_METABOLIC_THRESHOLD represents cost-benefit threshold
	// Metabolic cost above which synapses are candidates for elimination
	// Biological basis: Energy cost of maintaining synaptic proteins
	// Experimental evidence: Activity-dependent metabolic scaling
	// Computational use: Balances network efficiency with connectivity
	PRUNING_METABOLIC_THRESHOLD = 2.0
)

// === SYNAPTIC STABILITY CONSTANTS ===
// Based on spine imaging and synaptic lifetime studies

const (
	// STABLE_SYNAPSE_LIFETIME represents typical synapse persistence
	// Expected lifetime of established, active synapses
	// Biological measurement: Days-to-weeks in adult cortex spine imaging
	// Developmental variation: Much higher turnover in young animals
	// Simulation implication: Stable synapses should persist throughout runs
	STABLE_SYNAPSE_LIFETIME = 24 * time.Hour

	// UNSTABLE_SYNAPSE_LIFETIME represents transient synapse duration
	// Lifetime of exploratory or weak synaptic connections
	// Biological observation: Minutes-to-hours for unsuccessful connections
	// Functional significance: Allows rapid network reconfiguration
	// Learning implication: Failed connections are quickly eliminated
	UNSTABLE_SYNAPSE_LIFETIME = 30 * time.Minute
)

// =================================================================================
// NEUROTRANSMITTER-SPECIFIC CONSTANTS
// =================================================================================

// === CONCENTRATION SCALING FACTORS ===
// Based on neurotransmitter release and receptor binding studies
// References: Clements et al. (1992), Diamond (2001)

const (
	// GLUTAMATE_CONCENTRATION_SCALE converts signal to glutamate concentration
	// Peak concentration in synaptic cleft during vesicle release
	// Biological range: 1-3 mM peak concentration
	// Clearance: Rapid uptake by transporters (EAAT1/2/3)
	// Functional significance: High concentration for fast, reliable signaling
	GLUTAMATE_CONCENTRATION_SCALE = 2.0

	// GABA_CONCENTRATION_SCALE converts signal to GABA concentration
	// Peak concentration in synaptic cleft for inhibitory transmission
	// Biological range: 0.5-1 mM peak concentration
	// Clearance: GAT transporters, somewhat slower than glutamate
	// Functional significance: Strong inhibition with moderate concentration
	GABA_CONCENTRATION_SCALE = 1.5

	// DOPAMINE_CONCENTRATION_SCALE converts signal to dopamine concentration
	// Much lower concentration for volume transmission signaling
	// Biological range: 1-10 μM for phasic release, nM for tonic
	// Clearance: Slow uptake, wide spatial spread (50-100 μm)
	// Functional significance: Low concentration, high impact signaling
	DOPAMINE_CONCENTRATION_SCALE = 0.5

	// SEROTONIN_CONCENTRATION_SCALE converts signal to serotonin concentration
	// Lowest concentration for global neuromodulatory effects
	// Biological range: 0.1-1 μM for volume transmission
	// Clearance: Very slow, widespread effects throughout brain regions
	// Functional significance: Global state regulation at low concentrations
	SEROTONIN_CONCENTRATION_SCALE = 0.3

	// ACETYLCHOLINE_CONCENTRATION_SCALE for mixed synaptic/modulatory signaling
	// Intermediate concentration supporting both fast and slow signaling
	// Biological range: 0.1-1 mM synaptic, μM for muscarinic modulation
	// Clearance: Rapid enzymatic breakdown by acetylcholinesterase
	// Functional significance: Dual-mode signaling with intermediate concentration
	ACETYLCHOLINE_CONCENTRATION_SCALE = 1.0
)

// =================================================================================
// PERFORMANCE AND SAFETY LIMITS
// =================================================================================

// === COMPUTATIONAL PERFORMANCE LIMITS ===
// These limits ensure system performance while maintaining biological realism

const (
	// MAX_CONCURRENT_TRANSMISSIONS represents system throughput limit
	// Maximum simultaneous transmissions the system should handle
	// Performance basis: Maintains <1ms average latency under load
	// Biological relevance: Prevents unrealistic network activity bursts
	// Safety significance: Protects against memory/CPU exhaustion
	MAX_CONCURRENT_TRANSMISSIONS = 10000

	// === MEMORY MANAGEMENT LIMITS ===

	// MAX_PLASTICITY_EVENTS_PER_SECOND represents learning rate limit
	// Maximum plasticity updates per second per synapse
	// Performance basis: Prevents computational overload during learning
	// Biological realism: Matches realistic spike rates and pairing frequencies
	// Memory efficiency: Limits history storage requirements
	MAX_PLASTICITY_EVENTS_PER_SECOND = 100

	// MAX_PLASTICITY_HISTORY limits plasticity event storage
	// Memory efficiency: Removes old events while preserving recent learning
	MAX_PLASTICITY_HISTORY = 500

	// MAX_WEIGHT_HISTORY limits weight snapshot storage
	// Analysis capability: Tracks weight changes over time for trends
	MAX_WEIGHT_HISTORY = 200

	// MAX_ACTIVITY_HISTORY_SIZE represents memory limit for activity tracking
	// Maximum number of events stored for analysis and monitoring
	// Memory management: Prevents unbounded growth in long simulations
	// Analysis capability: Sufficient for meaningful pattern detection
	// Performance impact: Balances analysis depth with memory efficiency
	MAX_ACTIVITY_HISTORY_SIZE = 1000

	// MAX_SYNAPSE_WEIGHT represents absolute upper bound on synaptic strength
	// Safety limit preventing pathological network behavior
	// Biological justification: Physical limits of synaptic machinery
	// Simulation stability: Prevents numerical overflow and instability
	// Network behavior: Ensures no single synapse dominates network activity
	MAX_SYNAPSE_WEIGHT = 100.0

	// MIN_UPDATE_INTERVAL represents minimum time between state updates
	// Prevents excessive computational overhead from rapid updates
	// Performance optimization: Reduces unnecessary computation
	// Biological realism: Matches timescales of biological processes
	// System stability: Prevents race conditions in concurrent access
	MIN_UPDATE_INTERVAL = 100 * time.Microsecond
)

// === SAFETY AND VALIDATION LIMITS ===
// These limits prevent erroneous or pathological behavior

const (
	// MAX_CALCIUM_LEVEL represents physiological calcium upper limit
	// Prevents unrealistic calcium concentrations
	// Biological basis: Calcium buffering capacity and toxicity limits
	// Simulation safety: Prevents runaway calcium-dependent effects
	// Typical range: 0.1-10 μM in synaptic terminals
	MAX_CALCIUM_LEVEL = 10.0

	// MIN_CALCIUM_LEVEL represents minimum detectable calcium
	// Ensures some minimal calcium-dependent processes
	// Biological basis: Resting calcium levels in neurons
	// Functional significance: Maintains basic synaptic function
	// Typical value: ~100 nM resting calcium
	MIN_CALCIUM_LEVEL = 0.1

	// MAX_VESICLE_POOLS represents upper limit on vesicle numbers
	// Prevents unrealistic vesicle pool sizes
	// Biological constraint: Terminal size and vesicle packing limits
	// Memory efficiency: Limits vesicle tracking overhead
	// Simulation realism: Matches largest observed synaptic terminals
	MAX_VESICLE_POOLS = 10000

	// MAX_DELAY_DURATION represents maximum allowed transmission delay
	// Prevents unrealistic timing that could break network function
	// Biological limit: Longest plausible axonal conduction delays
	// Simulation practicality: Prevents timing-related bugs
	// Network function: Ensures reasonable temporal relationships
	MAX_DELAY_DURATION = 1 * time.Second
)

// =================================================================================
// BIOLOGICAL VALIDATION RANGES
// =================================================================================

// These ranges define biologically plausible parameter values for validation

const (
	// === WEIGHT VALIDATION ===
	BIOLOGICAL_MIN_WEIGHT = 0.0001 // Detectable synaptic response
	BIOLOGICAL_MAX_WEIGHT = 5.0    // Maximum observed synaptic strength

	// === FREQUENCY VALIDATION ===
	BIOLOGICAL_MIN_FREQUENCY = 0.1   // Minimum meaningful firing rate
	BIOLOGICAL_MAX_FREQUENCY = 500.0 // Maximum sustainable firing rate

	// === DELAY VALIDATION ===
	BIOLOGICAL_MIN_DELAY = 100 * time.Microsecond // Fastest synaptic transmission
	BIOLOGICAL_MAX_DELAY = 100 * time.Millisecond // Slowest reasonable delay

	// === CONCENTRATION VALIDATION ===
	BIOLOGICAL_MIN_CONCENTRATION = 0.001 // Threshold for biological effect
	BIOLOGICAL_MAX_CONCENTRATION = 50.0  // Maximum safe concentration

	// === PLASTICITY VALIDATION ===
	BIOLOGICAL_MIN_LEARNING_RATE = 0.0001 // Detectable learning
	BIOLOGICAL_MAX_LEARNING_RATE = 0.1    // Maximum stable learning rate
)

// =================================================================================
// EXPERIMENTAL DATA CONSTANTS
// =================================================================================

// STDP timing window constants from Bi & Poo (1998)
const (
	// Timing window boundaries (milliseconds)
	BIOLOGY_STDP_WINDOW_MS = 100.0 // ±100ms effective window

	// Peak plasticity timing
	BIOLOGY_LTP_PEAK_MS = 10.0 // Peak LTP at ~10ms pre-before-post
	BIOLOGY_LTD_PEAK_MS = 10.0 // Peak LTD at ~10ms post-before-pre

	// Magnitude ratios from experiments
	BIOLOGY_LTP_LTD_RATIO     = 1.5 // LTP typically 1.5x stronger than LTD
	BIOLOGY_MAX_WEIGHT_CHANGE = 0.6 // Max 60% weight change per pairing

	// Time constants from experimental fits
	BIOLOGY_LTP_TAU_MS = 16.8 // LTP decay time constant
	BIOLOGY_LTD_TAU_MS = 33.7 // LTD decay time constant (slower)
)

// Cooperativity thresholds from Sjöström et al. (2001)
const (
	BIOLOGY_COOPERATIVITY_THRESHOLD  = 3  // Minimum 3 inputs for plasticity
	BIOLOGY_HIGH_COOPERATIVITY       = 10 // Strong cooperativity effect
	BIOLOGY_COOPERATIVITY_SATURATION = 20 // Saturation point
)

// Neuromodulator effects from literature
const (
	BIOLOGY_DOPAMINE_ENHANCEMENT = 2.5 // 2.5x plasticity enhancement
	BIOLOGY_ACH_ATTENTION_GATE   = 1.8 // 1.8x with attention
	BIOLOGY_STRESS_OPTIMAL_LEVEL = 1.3 // Optimal norepinephrine level
)

// Frequency dependence from Bear & Malenka (1988)
const (
	BIOLOGY_LTD_FREQUENCY_HZ   = 1.0   // 1Hz typically induces LTD
	BIOLOGY_LTP_FREQUENCY_HZ   = 100.0 // 100Hz typically induces LTP
	BIOLOGY_THETA_FREQUENCY_HZ = 5.0   // 5Hz theta rhythm
)
