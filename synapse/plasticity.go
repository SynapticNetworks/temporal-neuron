/*
=================================================================================
SYNAPTIC PLASTICITY - BIOLOGICAL LEARNING MECHANISMS
=================================================================================

This module implements biologically accurate synaptic plasticity mechanisms,
primarily Spike-Timing Dependent Plasticity (STDP), that enable synapses to
strengthen or weaken based on the temporal relationship between pre- and
post-synaptic activity.

BIOLOGICAL FOUNDATION:
Synaptic plasticity is the biological basis of learning and memory. The most
well-studied form is STDP, discovered by Bi & Poo (1998) and extensively
characterized in subsequent research. This implementation models the key
biological mechanisms that govern synaptic strength changes.

STDP PRINCIPLES:
1. CAUSAL RELATIONSHIPS: When pre-synaptic spikes consistently precede
   post-synaptic spikes, the synapse strengthens (LTP - Long Term Potentiation)
2. ANTI-CAUSAL RELATIONSHIPS: When pre-synaptic spikes follow post-synaptic
   spikes, the synapse weakens (LTD - Long Term Depression)
3. TIMING WINDOW: Only spikes within a critical timing window (~±100ms)
   produce plasticity effects
4. EXPONENTIAL DECAY: Plasticity strength decays exponentially with spike
   timing difference

MOLECULAR MECHANISMS MODELED:
- NMDA receptor activation and calcium dynamics
- CaMKII autophosphorylation (LTP induction)
- Calcineurin activation (LTD induction)
- AMPA receptor trafficking and phosphorylation
- Protein synthesis and structural changes

EXPERIMENTAL BASIS:
- Bi & Poo (1998): "Synaptic modifications in cultured hippocampal neurons"
- Sjöström et al. (2001): "Rate, timing, and cooperativity jointly determine cortical synaptic plasticity"
- Caporale & Dan (2008): "Spike timing-dependent plasticity: a Hebbian learning rule"
- Markram et al. (1997): "Regulation of synaptic efficacy by coincidence of postsynaptic APs and EPSPs"

METAPLASTICITY:
Advanced implementation includes metaplasticity - the plasticity of plasticity
itself, where the threshold and magnitude of plasticity changes based on the
history of synaptic activity (Bienenstock-Cooper-Munro rule).
=================================================================================
*/

package synapse

import (
	"math"
	"time"
)

// =================================================================================
// BIOLOGICAL PLASTICITY CONSTANTS
// =================================================================================

// === WEIGHT BOUNDARY CONSTANTS ===
// Biologically realistic limits on synaptic strength to prevent pathological behavior

const (
	// STDP_DEFAULT_MIN_WEIGHT prevents complete synapse elimination
	// Biological basis: Even "silent" synapses retain minimal AMPA receptors
	// Experimental evidence: LTD saturates at ~10-20% of baseline response
	// Functional significance: Allows future re-strengthening (synaptic tagging)
	// Network stability: Prevents irreversible connection loss
	STDP_DEFAULT_MIN_WEIGHT = 0.001

	// STDP_DEFAULT_MAX_WEIGHT prevents pathological over-strengthening
	// Biological basis: Physical limits of postsynaptic density and receptor density
	// Experimental observation: LTP saturates at 2-3x baseline in most preparations
	// Molecular constraint: Limited by available AMPA receptor pool
	// Network function: Forces competition between synapses for strength
	STDP_DEFAULT_MAX_WEIGHT = 2.0

	// STDP_SATURATION_FACTOR controls approach to weight boundaries
	// Mathematical implementation: Prevents abrupt cutoffs at boundaries
	// Biological realism: Weight changes slow as boundaries are approached
	// Stability benefit: Smoother dynamics near saturation points
	STDP_SATURATION_FACTOR = 0.1
)

// === NEUROMODULATOR INFLUENCE CONSTANTS ===
// How different neuromodulators affect plasticity induction and expression

const (
	// DOPAMINE_LEARNING_MULTIPLIER for reward-based learning enhancement
	// Biological mechanism: D1 receptor activation enhances LTP, reduces LTD threshold
	// Experimental evidence: VTA stimulation enhances hippocampal and cortical LTP
	// Behavioral relevance: Links rewards to synaptic strengthening
	// Typical enhancement: 2-5x normal plasticity in presence of dopamine
	DOPAMINE_LEARNING_MULTIPLIER = 3.0

	// ACETYLCHOLINE_ATTENTION_MULTIPLIER for attention-gated learning
	// Biological mechanism: Muscarinic and nicotinic receptor modulation of plasticity
	// Experimental evidence: ACh enhances LTP in cortical and hippocampal preparations
	// Behavioral function: Gates plasticity during attention and arousal states
	// Cholinergic enhancement: 1.5-2x normal plasticity rates
	ACETYLCHOLINE_ATTENTION_MULTIPLIER = 1.8

	// NOREPINEPHRINE_STRESS_MULTIPLIER for stress-related learning changes
	// Biological mechanism: β-adrenergic receptor effects on cAMP and plasticity
	// Stress effects: Low stress enhances learning, high stress can impair it
	// Experimental range: 0.5-2.0x depending on concentration and timing
	// Inverted-U relationship: Optimal at moderate norepinephrine levels
	NOREPINEPHRINE_STRESS_MULTIPLIER = 1.4

	// GABA_INHIBITION_THRESHOLD for inhibitory modulation of plasticity
	// Biological mechanism: GABA-B receptors reduce calcium influx, impair LTP
	// Disinhibition effects: Reducing inhibition enhances plasticity induction
	// Experimental evidence: GABA-B antagonists facilitate LTP induction
	// Typical suppression: 0.3-0.7x normal plasticity under strong inhibition
	GABA_INHIBITION_THRESHOLD = 0.5
)

// === DEVELOPMENTAL AND AGE-RELATED CONSTANTS ===
// How plasticity changes across the lifespan

const (
	// CRITICAL_PERIOD_MULTIPLIER for enhanced juvenile plasticity
	// Biological basis: Higher NMDA/AMPA ratios and calcium permeability in young animals
	// Experimental evidence: LTP magnitude and duration greater in juvenile tissue
	// Developmental decline: Gradual reduction in plasticity with age
	// Critical period: Enhanced plasticity during specific developmental windows
	CRITICAL_PERIOD_MULTIPLIER = 2.5

	// AGING_PLASTICITY_REDUCTION for age-related plasticity decline
	// Biological mechanisms: Reduced NMDA receptor function, calcium dysregulation
	// Experimental evidence: LTP harder to induce and maintain in aged animals
	// Cognitive correlation: Reduced plasticity correlates with memory impairments
	// Typical reduction: 0.3-0.6x youthful plasticity in aged preparations
	AGING_PLASTICITY_REDUCTION = 0.4

	// HOMEOSTATIC_SCALING_RATE for activity-dependent scaling
	// Biological mechanism: Cell-wide scaling of synaptic strengths to maintain stability
	// Experimental timescale: Hours to days for significant scaling changes
	// Molecular basis: Changes in postsynaptic receptor density
	// Functional role: Prevents saturation while preserving relative weight differences
	HOMEOSTATIC_SCALING_RATE = 0.01 // 1% per hour of extreme activity
)

// === MOLECULAR PATHWAY CONSTANTS ===
// Time constants for different molecular processes underlying plasticity

const (
	// EARLY_PHASE_DURATION for protein kinase-dependent early LTP/LTD
	// Biological process: CaMKII autophosphorylation, PKA/PKC activation
	// Experimental measurement: 1-3 hours duration without protein synthesis
	// Molecular requirement: Existing proteins and post-translational modifications
	// Functional significance: Immediate learning without gene expression
	EARLY_PHASE_DURATION = 2 * time.Hour

	// LATE_PHASE_DURATION for protein synthesis-dependent late LTP/LTD
	// Biological process: CREB-mediated gene expression and protein synthesis
	// Experimental measurement: >4 hours, can last days to weeks
	// Molecular requirement: New protein synthesis and structural changes
	// Memory correlation: Late phase required for long-term memory formation
	LATE_PHASE_DURATION = 24 * time.Hour

	// CONSOLIDATION_WINDOW for synaptic tag and capture
	// Biological mechanism: Period when synapses can capture plasticity-related proteins
	// Experimental evidence: ~3 hour window for tag-based protein capture
	// Functional significance: Allows weak inputs to benefit from strong stimulation
	// Molecular basis: Local protein synthesis and trafficking
	CONSOLIDATION_WINDOW = 3 * time.Hour
)

// === COMPUTATIONAL PERFORMANCE CONSTANTS ===
// Limits to ensure efficient computation while maintaining biological accuracy

const (
	// MAX_SPIKE_HISTORY_SIZE limits memory usage for spike timing analysis
	// Performance consideration: Prevents unbounded memory growth
	// Biological relevance: Matches working memory timescales (~10-30 seconds)
	// Plasticity window: Only recent spikes are relevant for STDP
	MAX_SPIKE_HISTORY_SIZE = 1000

	// PLASTICITY_UPDATE_INTERVAL controls frequency of plasticity calculations
	// Performance optimization: Batch processing of plasticity events
	// Biological timescale: Millisecond precision sufficient for STDP
	// Computational efficiency: Reduces overhead in large networks
	PLASTICITY_UPDATE_INTERVAL = 1 * time.Millisecond

	// MIN_WEIGHT_CHANGE_THRESHOLD ignores tiny weight changes for efficiency
	// Computational optimization: Avoids processing insignificant changes
	// Biological justification: Molecular noise threshold for detectable changes
	// Network efficiency: Reduces unnecessary computation and memory updates
	MIN_WEIGHT_CHANGE_THRESHOLD = 0.0001 // 0.01% of typical weight
)

// =================================================================================
// PLASTICITY CALCULATOR - CORE STDP IMPLEMENTATION
// =================================================================================

// PlasticityCalculator implements biologically accurate STDP and related mechanisms
// Encapsulates all plasticity algorithms with configurable parameters
type PlasticityCalculator struct {
	// Configuration parameters
	config STDPConfig

	// Spike timing history for STDP calculation
	preSpikes  []time.Time // Recent pre-synaptic spike times
	postSpikes []time.Time // Recent post-synaptic spike times

	// Metaplasticity state
	activityHistory     []float64 // Recent activity levels for metaplasticity
	plasticityThreshold float64   // Current threshold for plasticity induction

	// Neuromodulator levels
	dopamineLevel       float64 // Current dopamine concentration
	acetylcholineLevel  float64 // Current acetylcholine concentration
	norepinephrineLevel float64 // Current norepinephrine concentration

	// Developmental factors
	developmentalStage float64 // 0.0 = newborn, 1.0 = adult, >1.0 = aged

	// Statistics and monitoring
	totalEvents   int64     // Total plasticity events processed
	lastUpdate    time.Time // Last plasticity calculation
	averageChange float64   // Running average of weight changes
}

// NewPlasticityCalculator creates a new plasticity calculator with biological defaults
func NewPlasticityCalculator(config STDPConfig) *PlasticityCalculator {
	// Validate configuration parameters
	if !config.IsValid() {
		// Use biological defaults if invalid config provided
		config = CreateDefaultSTDPConfig()
	}

	return &PlasticityCalculator{
		config: config,

		// Initialize spike history
		preSpikes:  make([]time.Time, 0),
		postSpikes: make([]time.Time, 0),

		// Initialize metaplasticity
		activityHistory:     make([]float64, 0),
		plasticityThreshold: 1.0, // Baseline threshold

		// Initialize neuromodulator levels (baseline)
		dopamineLevel:       1.0,
		acetylcholineLevel:  1.0,
		norepinephrineLevel: 1.0,

		// Initialize as adult brain (can be adjusted)
		developmentalStage: 1.0,

		// Initialize statistics
		totalEvents:   0,
		lastUpdate:    time.Now(),
		averageChange: 0.0,
	}
}

// =================================================================================
// CORE STDP CALCULATION METHODS
// =================================================================================

// CalculateSTDPWeightChange computes weight change based on spike timing
// This is the core STDP algorithm implementing the biological learning rule
//
// BIOLOGICAL ALGORITHM:
// 1. Determine spike timing relationship (Δt = t_pre - t_post)
// 2. Check if timing falls within plasticity window
// 3. Calculate exponential decay based on time constant
// 4. Apply asymmetry for LTD vs LTP
// 5. Modulate by neuromodulators and developmental factors
// 6. Apply cooperativity and metaplasticity constraints
//
// Parameters:
//
//	deltaT: Spike timing difference (t_pre - t_post)
//	currentWeight: Current synaptic weight for normalization
//	cooperativeInputs: Number of concurrent inputs (for cooperativity)
//
// Returns:
//
//	Weight change to apply (positive = LTP, negative = LTD)
func (pc *PlasticityCalculator) CalculateSTDPWeightChange(deltaT time.Duration, currentWeight float64, cooperativeInputs int) float64 {
	// Skip calculation if STDP disabled
	if !pc.config.Enabled {
		return 0.0
	}

	// Convert timing to milliseconds for calculation
	deltaTMs := deltaT.Seconds() * 1000.0
	windowMs := pc.config.WindowSize.Seconds() * 1000.0

	// Check if timing is within plasticity window
	if math.Abs(deltaTMs) >= windowMs {
		return 0.0 // No plasticity outside window
	}

	// Check cooperativity requirement
	if cooperativeInputs < COOPERATIVITY_THRESHOLD {
		return 0.0 // Insufficient cooperative inputs
	}

	// Get time constant in milliseconds
	tauMs := pc.config.TimeConstant.Seconds() * 1000.0
	if tauMs <= 0 {
		return 0.0 // Invalid time constant
	}

	// Calculate base STDP change
	var baseChange float64

	if math.Abs(deltaTMs) < STDP_SIMULTANEOUS_THRESHOLD.Seconds()*1000.0 {
		// Simultaneous spikes - small LTP
		baseChange = pc.config.LearningRate * 0.1

	} else if deltaTMs < 0 {
		// CAUSAL: Pre before post → LTP (strengthening)
		baseChange = pc.config.LearningRate * math.Exp(deltaTMs/tauMs)

	} else {
		// ANTI-CAUSAL: Pre after post → LTD (weakening)
		baseChange = -pc.config.LearningRate * pc.config.AsymmetryRatio * math.Exp(-deltaTMs/tauMs)
	}

	// Apply weight-dependent scaling (weak synapses change more)
	weightFactor := pc.calculateWeightDependence(currentWeight)
	baseChange *= weightFactor

	// Apply neuromodulator influences
	neuromodulatorFactor := pc.calculateNeuromodulatorInfluence()
	baseChange *= neuromodulatorFactor

	// Apply developmental factors
	developmentalFactor := pc.calculateDevelopmentalFactor()
	baseChange *= developmentalFactor

	// Apply metaplasticity (sliding threshold)
	metaplasticityFactor := pc.calculateMetaplasticityFactor(currentWeight)
	baseChange *= metaplasticityFactor

	// Update statistics
	pc.updatePlasticityStatistics(baseChange)

	return baseChange
}

// CalculateFrequencyDependentPlasticity implements frequency-dependent learning rules
// Biological basis: Low frequency → LTD, high frequency → LTP (Bienenstock-Cooper-Munro)
//
// Parameters:
//
//	frequency: Stimulation frequency in Hz
//	currentWeight: Current synaptic weight
//	duration: Duration of stimulation
//
// Returns:
//
//	Weight change based on frequency-dependent rules
func (pc *PlasticityCalculator) CalculateFrequencyDependentPlasticity(frequency float64, currentWeight float64, duration time.Duration) float64 {
	if !pc.config.FrequencyDependent {
		return 0.0
	}

	// Determine plasticity direction based on frequency
	var baseChange float64

	if frequency < FREQUENCY_DEPENDENCE_THRESHOLD {
		// Low frequency → LTD
		intensity := 1.0 - (frequency / FREQUENCY_DEPENDENCE_THRESHOLD)
		baseChange = -pc.config.LearningRate * intensity

	} else {
		// High frequency → LTP
		intensity := (frequency - FREQUENCY_DEPENDENCE_THRESHOLD) / FREQUENCY_DEPENDENCE_THRESHOLD
		intensity = math.Min(intensity, 2.0) // Cap at 2x threshold
		baseChange = pc.config.LearningRate * intensity
	}

	// Scale by stimulation duration
	durationFactor := math.Min(duration.Seconds()/60.0, 1.0) // Max effect at 1 minute
	baseChange *= durationFactor

	// Apply weight dependence and modulation
	weightFactor := pc.calculateWeightDependence(currentWeight)
	neuromodulatorFactor := pc.calculateNeuromodulatorInfluence()

	return baseChange * weightFactor * neuromodulatorFactor
}

// CalculateHomeostatic Scaling implements synaptic scaling for network stability
// Biological mechanism: Global scaling of all synapses to maintain total input strength
//
// Parameters:
//
//	targetActivity: Desired activity level
//	currentActivity: Current activity level
//	currentWeight: Current synaptic weight
//
// Returns:
//
//	Scaling factor to apply to weight (multiplicative)
func (pc *PlasticityCalculator) CalculateHomeostaticScaling(targetActivity, currentActivity, currentWeight float64) float64 {
	// Calculate activity ratio
	if currentActivity <= 0 {
		return 1.0 // No scaling if no activity
	}

	activityRatio := targetActivity / currentActivity

	// Gradual scaling to prevent instability
	scalingStrength := HOMEOSTATIC_SCALING_RATE
	scalingFactor := 1.0 + scalingStrength*(activityRatio-1.0)

	// Constrain scaling to reasonable bounds
	scalingFactor = math.Max(0.5, math.Min(2.0, scalingFactor))

	return scalingFactor
}

// =================================================================================
// SPIKE TIMING MANAGEMENT
// =================================================================================

// AddPreSynapticSpike records a pre-synaptic spike for STDP calculation
func (pc *PlasticityCalculator) AddPreSynapticSpike(spikeTime time.Time) {
	pc.preSpikes = append(pc.preSpikes, spikeTime)
	pc.cleanupOldSpikes()
}

// AddPostSynapticSpike records a post-synaptic spike for STDP calculation
func (pc *PlasticityCalculator) AddPostSynapticSpike(spikeTime time.Time) {
	pc.postSpikes = append(pc.postSpikes, spikeTime)
	pc.cleanupOldSpikes()
}

// GetRecentSpikePairs finds all spike pairs within the STDP window
// Returns pairs of (preTime, postTime, deltaT) for STDP calculation
func (pc *PlasticityCalculator) GetRecentSpikePairs() []SpikePair {
	pairs := make([]SpikePair, 0)

	// Find all combinations within STDP window
	for _, preTime := range pc.preSpikes {
		for _, postTime := range pc.postSpikes {
			deltaT := preTime.Sub(postTime)

			if math.Abs(float64(deltaT)) <= float64(pc.config.WindowSize) {
				pairs = append(pairs, SpikePair{
					PreTime:  preTime,
					PostTime: postTime,
					DeltaT:   deltaT,
				})
			}
		}
	}

	return pairs
}

// cleanupOldSpikes removes spikes outside the STDP window to manage memory
func (pc *PlasticityCalculator) cleanupOldSpikes() {
	now := time.Now()
	cutoff := now.Add(-pc.config.WindowSize)

	// Clean pre-synaptic spikes
	validPreSpikes := make([]time.Time, 0)
	for _, spike := range pc.preSpikes {
		if spike.After(cutoff) {
			validPreSpikes = append(validPreSpikes, spike)
		}
	}
	pc.preSpikes = validPreSpikes

	// Clean post-synaptic spikes
	validPostSpikes := make([]time.Time, 0)
	for _, spike := range pc.postSpikes {
		if spike.After(cutoff) {
			validPostSpikes = append(validPostSpikes, spike)
		}
	}
	pc.postSpikes = validPostSpikes

	// Enforce maximum history size
	if len(pc.preSpikes) > MAX_SPIKE_HISTORY_SIZE {
		excess := len(pc.preSpikes) - MAX_SPIKE_HISTORY_SIZE
		pc.preSpikes = pc.preSpikes[excess:]
	}

	if len(pc.postSpikes) > MAX_SPIKE_HISTORY_SIZE {
		excess := len(pc.postSpikes) - MAX_SPIKE_HISTORY_SIZE
		pc.postSpikes = pc.postSpikes[excess:]
	}
}

// =================================================================================
// NEUROMODULATOR AND CONTEXT MANAGEMENT
// =================================================================================

// SetNeuromodulatorLevels updates neuromodulator concentrations
func (pc *PlasticityCalculator) SetNeuromodulatorLevels(dopamine, acetylcholine, norepinephrine float64) {
	// Clamp to reasonable biological ranges
	pc.dopamineLevel = math.Max(0.0, math.Min(5.0, dopamine))
	pc.acetylcholineLevel = math.Max(0.0, math.Min(3.0, acetylcholine))
	pc.norepinephrineLevel = math.Max(0.0, math.Min(3.0, norepinephrine))
}

// SetDevelopmentalStage sets the developmental stage for age-dependent plasticity
func (pc *PlasticityCalculator) SetDevelopmentalStage(stage float64) {
	// 0.0 = newborn, 1.0 = adult, >1.0 = aged
	pc.developmentalStage = math.Max(0.0, stage)
}

// UpdateActivityHistory adds recent activity level for metaplasticity
func (pc *PlasticityCalculator) UpdateActivityHistory(activityLevel float64) {
	pc.activityHistory = append(pc.activityHistory, activityLevel)

	// Keep only recent history for metaplasticity calculation
	maxHistory := 100 // Last 100 activity measurements
	if len(pc.activityHistory) > maxHistory {
		pc.activityHistory = pc.activityHistory[len(pc.activityHistory)-maxHistory:]
	}

	// Update metaplasticity threshold based on activity history
	pc.updateMetaplasticityThreshold()
}

// =================================================================================
// INTERNAL CALCULATION HELPERS
// =================================================================================

// calculateWeightDependence implements weight-dependent plasticity scaling
// Biological observation: Weak synapses show larger plasticity than strong ones
func (pc *PlasticityCalculator) calculateWeightDependence(currentWeight float64) float64 {
	// Normalize weight to [0,1] range
	normalizedWeight := (currentWeight - pc.config.MinWeight) / (pc.config.MaxWeight - pc.config.MinWeight)
	normalizedWeight = math.Max(0.0, math.Min(1.0, normalizedWeight))

	// Weak synapses (low weight) have higher plasticity
	// Strong synapses (high weight) have lower plasticity
	weightFactor := 2.0 - normalizedWeight // Range: [1.0, 2.0]

	return weightFactor
}

// calculateNeuromodulatorInfluence combines effects of multiple neuromodulators
func (pc *PlasticityCalculator) calculateNeuromodulatorInfluence() float64 {
	influence := 1.0 // Baseline (no modulation)

	// Dopamine enhances learning (especially LTP)
	if pc.dopamineLevel > 1.0 {
		dopamineEffect := 1.0 + (pc.dopamineLevel-1.0)*(DOPAMINE_LEARNING_MULTIPLIER-1.0)
		influence *= dopamineEffect
	}

	// Acetylcholine enhances attention-gated learning
	if pc.acetylcholineLevel > 1.0 {
		acetylfecholineEffect := 1.0 + (pc.acetylcholineLevel-1.0)*(ACETYLCHOLINE_ATTENTION_MULTIPLIER-1.0)
		influence *= acetylfecholineEffect
	}

	// Norepinephrine has complex effects (inverted U-curve)
	if pc.norepinephrineLevel != 1.0 {
		// Optimal at moderate levels, reduced at very high or low levels
		optimal := 1.5 // Optimal norepinephrine level
		deviation := math.Abs(pc.norepinephrineLevel-optimal) / optimal
		norepinephrineEffect := NOREPINEPHRINE_STRESS_MULTIPLIER * (1.0 - 0.5*deviation)
		norepinephrineEffect = math.Max(0.2, norepinephrineEffect)
		influence *= norepinephrineEffect
	}

	return influence
}

// calculateDevelopmentalFactor adjusts plasticity based on age/development
func (pc *PlasticityCalculator) calculateDevelopmentalFactor() float64 {
	if pc.developmentalStage < 0.5 {
		// Juvenile: Enhanced plasticity
		return CRITICAL_PERIOD_MULTIPLIER
	} else if pc.developmentalStage <= 1.0 {
		// Adult: Normal plasticity
		return 1.0
	} else {
		// Aged: Reduced plasticity
		agingFactor := 1.0 / pc.developmentalStage
		return AGING_PLASTICITY_REDUCTION * agingFactor
	}
}

// calculateMetaplasticityFactor implements sliding threshold metaplasticity
func (pc *PlasticityCalculator) calculateMetaplasticityFactor(currentWeight float64) float64 {
	// Metaplasticity: plasticity threshold slides with activity history
	// High activity → higher threshold (harder to potentiate)
	// Low activity → lower threshold (easier to potentiate)

	if len(pc.activityHistory) < 10 {
		return 1.0 // Not enough history for metaplasticity
	}

	// Calculate average recent activity
	recentActivity := 0.0
	for _, activity := range pc.activityHistory {
		recentActivity += activity
	}
	recentActivity /= float64(len(pc.activityHistory))

	// Threshold slides with activity (BCM rule)
	thresholdShift := (recentActivity - 1.0) * pc.config.MetaplasticityRate
	adjustedThreshold := pc.plasticityThreshold + thresholdShift
	adjustedThreshold = math.Max(0.1, math.Min(3.0, adjustedThreshold))

	// Factor based on current weight relative to threshold
	if currentWeight < adjustedThreshold {
		return 1.0 + 0.5*(adjustedThreshold-currentWeight)/adjustedThreshold
	} else {
		return 1.0 - 0.3*(currentWeight-adjustedThreshold)/adjustedThreshold
	}
}

// updateMetaplasticityThreshold adjusts the plasticity threshold over time
func (pc *PlasticityCalculator) updateMetaplasticityThreshold() {
	if len(pc.activityHistory) < 5 {
		return
	}

	// Calculate trend in recent activity
	recent := pc.activityHistory[len(pc.activityHistory)-5:]
	trend := 0.0
	for i := 1; i < len(recent); i++ {
		trend += recent[i] - recent[i-1]
	}
	trend /= float64(len(recent) - 1)

	// Adjust threshold based on activity trend
	thresholdChange := trend * METAPLASTICITY_RATE * 0.1
	pc.plasticityThreshold += thresholdChange
	pc.plasticityThreshold = math.Max(0.1, math.Min(5.0, pc.plasticityThreshold))
}

// updatePlasticityStatistics tracks plasticity events for analysis
func (pc *PlasticityCalculator) updatePlasticityStatistics(weightChange float64) {
	pc.totalEvents++
	pc.lastUpdate = time.Now()

	// Update running average of weight changes
	alpha := 0.1 // Learning rate for running average
	pc.averageChange = (1-alpha)*pc.averageChange + alpha*math.Abs(weightChange)
}

// =================================================================================
// ADVANCED PLASTICITY MECHANISMS
// =================================================================================

// CalculateHeterosynapticPlasticity implements spreading plasticity to nearby synapses
// Biological basis: Plasticity can spread to neighboring synapses via diffusible factors
//
// Parameters:
//
//	distance: Distance from activated synapse (micrometers)
//	primaryChange: Weight change at the primary synapse
//
// Returns:
//
//	Weight change for synapse at given distance
func (pc *PlasticityCalculator) CalculateHeterosynapticPlasticity(distance float64, primaryChange float64) float64 {
	// No heterosynaptic effects beyond range
	if distance > HETEROSYNAPTIC_RANGE {
		return 0.0
	}

	// Exponential decay with distance
	decayFactor := math.Exp(-distance / (HETEROSYNAPTIC_RANGE / 3.0))

	// Heterosynaptic changes are typically smaller and opposite sign
	heterosynapticChange := -0.1 * primaryChange * decayFactor

	return heterosynapticChange
}

// CalculateProteinSynthesisDependentPlasticity models late-phase LTP/LTD
// Biological basis: Long-lasting plasticity requires new protein synthesis
//
// Parameters:
//
//	initialChange: Early-phase weight change
//	stimulationStrength: Strength of inducing stimulation
//	timeSinceInduction: Time since plasticity induction
//
// Returns:
//
//	Additional weight change from protein synthesis
func (pc *PlasticityCalculator) CalculateProteinSynthesisDependentPlasticity(initialChange float64, stimulationStrength float64, timeSinceInduction time.Duration) float64 {
	// Only strong stimulation triggers protein synthesis
	if math.Abs(stimulationStrength) < 2.0 {
		return 0.0
	}

	// Late phase begins after early phase
	if timeSinceInduction < EARLY_PHASE_DURATION {
		return 0.0
	}

	// Late phase decays over time
	if timeSinceInduction > LATE_PHASE_DURATION {
		return 0.0
	}

	// Protein synthesis enhances initial change
	enhancementFactor := 2.0 * stimulationStrength / 3.0 // Proportional to stimulation
	latePhaseChange := initialChange * enhancementFactor

	// Apply temporal profile (ramp up then decay)
	timeProgress := (timeSinceInduction - EARLY_PHASE_DURATION).Seconds() / LATE_PHASE_DURATION.Seconds()
	temporalProfile := 4 * timeProgress * (1 - timeProgress) // Peaked at 0.5

	return latePhaseChange * temporalProfile
}

// CalculateSynapticTaggingAndCapture models capture of plasticity-related proteins
// Biological mechanism: Weak inputs can capture proteins triggered by strong inputs
//
// Parameters:
//
//	weakSynapseChange: Weight change at weakly stimulated synapse
//	strongSynapseDistance: Distance to strongly stimulated synapse
//	timeSinceStrongStimulation: Time since strong stimulation
//
// Returns:
//
//	Enhanced weight change due to protein capture
func (pc *PlasticityCalculator) CalculateSynapticTaggingAndCapture(weakSynapseChange float64, strongSynapseDistance float64, timeSinceStrongStimulation time.Duration) float64 {
	// Only within consolidation window
	if timeSinceStrongStimulation > CONSOLIDATION_WINDOW {
		return 0.0
	}

	// Only for nearby synapses
	if strongSynapseDistance > HETEROSYNAPTIC_RANGE*2 {
		return 0.0
	}

	// Calculate capture efficiency
	distanceFactor := math.Exp(-strongSynapseDistance / HETEROSYNAPTIC_RANGE)
	timeFactor := 1.0 - (timeSinceStrongStimulation.Seconds() / CONSOLIDATION_WINDOW.Seconds())

	captureEfficiency := distanceFactor * timeFactor

	// Captured proteins enhance weak changes
	enhancement := weakSynapseChange * 3.0 * captureEfficiency

	return enhancement
}

// =================================================================================
// CONFIGURATION AND FACTORY FUNCTIONS
// =================================================================================

// CreateDefaultSTDPConfig returns biologically realistic default STDP parameters
func CreateDefaultSTDPConfig() STDPConfig {
	return STDPConfig{
		Enabled:        true,
		LearningRate:   STDP_LEARNING_RATE,
		TimeConstant:   STDP_TIME_CONSTANT,
		WindowSize:     STDP_WINDOW_SIZE,
		MinWeight:      STDP_DEFAULT_MIN_WEIGHT,
		MaxWeight:      STDP_DEFAULT_MAX_WEIGHT,
		AsymmetryRatio: STDP_ASYMMETRY_RATIO,

		// Advanced features
		FrequencyDependent:     true,
		MetaplasticityRate:     METAPLASTICITY_RATE,
		CooperativityThreshold: COOPERATIVITY_THRESHOLD,
	}
}

// CreateConservativeSTDPConfig returns conservative plasticity parameters
func CreateConservativeSTDPConfig() STDPConfig {
	config := CreateDefaultSTDPConfig()

	// Reduce learning rate for stability
	config.LearningRate = STDP_LEARNING_RATE * 0.5

	// Narrow timing window
	config.WindowSize = time.Duration(float64(STDP_WINDOW_SIZE) * 0.7)

	// Higher cooperativity requirement
	config.CooperativityThreshold = COOPERATIVITY_THRESHOLD + 2

	return config
}

// CreateDevelopmentalSTDPConfig returns enhanced plasticity for development
func CreateDevelopmentalSTDPConfig() STDPConfig {
	config := CreateDefaultSTDPConfig()

	// Enhanced learning for development
	config.LearningRate = STDP_LEARNING_RATE * CRITICAL_PERIOD_MULTIPLIER

	// Wider timing window
	config.WindowSize = time.Duration(float64(STDP_WINDOW_SIZE) * 1.5)
	// Lower cooperativity requirement
	config.CooperativityThreshold = max(1, COOPERATIVITY_THRESHOLD-1)

	return config
}

// CreateAgedSTDPConfig returns reduced plasticity for aging
func CreateAgedSTDPConfig() STDPConfig {
	config := CreateDefaultSTDPConfig()

	// Reduced learning with age
	config.LearningRate = STDP_LEARNING_RATE * AGING_PLASTICITY_REDUCTION

	// Narrower timing window
	config.WindowSize = time.Duration(float64(STDP_WINDOW_SIZE) * 0.8)

	// Higher cooperativity requirement
	config.CooperativityThreshold = COOPERATIVITY_THRESHOLD + 1

	return config
}

// =================================================================================
// UTILITY TYPES AND FUNCTIONS
// =================================================================================

// SpikePair represents a pair of spikes for STDP calculation
type SpikePair struct {
	PreTime  time.Time     `json:"pre_time"`  // Pre-synaptic spike time
	PostTime time.Time     `json:"post_time"` // Post-synaptic spike time
	DeltaT   time.Duration `json:"delta_t"`   // Timing difference (t_pre - t_post)
}

// GetDirection returns whether timing is causal, anti-causal, or simultaneous
func (sp SpikePair) GetDirection() string {
	if math.Abs(float64(sp.DeltaT)) < float64(STDP_SIMULTANEOUS_THRESHOLD) {
		return "simultaneous"
	} else if sp.DeltaT < 0 {
		return "causal" // Pre before post
	} else {
		return "anti_causal" // Pre after post
	}
}

// PlasticityStats provides statistics about plasticity calculator performance
type PlasticityStats struct {
	TotalEvents    int64     `json:"total_events"`     // Total plasticity events
	AverageChange  float64   `json:"average_change"`   // Average weight change magnitude
	LastUpdate     time.Time `json:"last_update"`      // Last plasticity calculation
	PreSpikeCount  int       `json:"pre_spike_count"`  // Number of stored pre-spikes
	PostSpikeCount int       `json:"post_spike_count"` // Number of stored post-spikes
	ThresholdValue float64   `json:"threshold_value"`  // Current metaplasticity threshold
}

// GetStatistics returns current plasticity calculator statistics
func (pc *PlasticityCalculator) GetStatistics() PlasticityStats {
	return PlasticityStats{
		TotalEvents:    pc.totalEvents,
		AverageChange:  pc.averageChange,
		LastUpdate:     pc.lastUpdate,
		PreSpikeCount:  len(pc.preSpikes),
		PostSpikeCount: len(pc.postSpikes),
		ThresholdValue: pc.plasticityThreshold,
	}
}

// Reset clears all plasticity history and resets to initial state
func (pc *PlasticityCalculator) Reset() {
	pc.preSpikes = make([]time.Time, 0)
	pc.postSpikes = make([]time.Time, 0)
	pc.activityHistory = make([]float64, 0)
	pc.plasticityThreshold = 1.0
	pc.totalEvents = 0
	pc.averageChange = 0.0
	pc.lastUpdate = time.Now()
}

// ValidateSTDPParameters checks if STDP parameters are biologically reasonable
func ValidateSTDPParameters(config STDPConfig) []string {
	var warnings []string

	// Check learning rate
	if config.LearningRate > 0.1 {
		warnings = append(warnings, "Learning rate > 10% may cause instability")
	}
	if config.LearningRate < 0.001 {
		warnings = append(warnings, "Learning rate < 0.1% may be too slow for learning")
	}

	// Check time constant
	if config.TimeConstant > 100*time.Millisecond {
		warnings = append(warnings, "Time constant > 100ms is unusually large")
	}
	if config.TimeConstant < 5*time.Millisecond {
		warnings = append(warnings, "Time constant < 5ms is unusually small")
	}

	// Check window size
	if config.WindowSize > 500*time.Millisecond {
		warnings = append(warnings, "STDP window > 500ms is extremely large")
	}
	if config.WindowSize < 10*time.Millisecond {
		warnings = append(warnings, "STDP window < 10ms may miss relevant spike pairs")
	}

	// Check weight bounds
	if config.MaxWeight > 10.0 {
		warnings = append(warnings, "Maximum weight > 10.0 may cause network instability")
	}
	if config.MinWeight < 0.0 {
		warnings = append(warnings, "Negative minimum weight is non-biological")
	}

	return warnings
}
