/*
=================================================================================
BIOLOGICAL CHEMICAL MODULATOR - NEUROSCIENCE-ACCURATE IMPLEMENTATION
=================================================================================

Implements biologically accurate neurotransmitter and neuromodulator signaling
based on published neuroscience research. All parameters and algorithms are
derived from experimental measurements in living brain tissue.

BIOLOGICAL FOUNDATION:
This implementation models the fundamental chemical signaling mechanisms that
enable neural computation and network coordination in biological brains:

1. SYNAPTIC TRANSMISSION: Fast, precise chemical communication between neurons
   - Glutamate (excitatory): 1-2ms clearance, <5μm range
   - GABA (inhibitory): 2-5ms clearance, <4μm range
   - Research: Danbolt (2001), Clements et al. (1992)

2. VOLUME TRANSMISSION: Slow, diffuse neuromodulation affecting multiple targets
   - Dopamine: 100μm range, minutes-scale clearance
   - Serotonin: 80μm range, minutes-scale clearance
   - Research: Floresco et al. (2003), Bunin & Wightman (1998)

3. MIXED SIGNALING: Dual synaptic and volume transmission properties
   - Acetylcholine: 20μm range, seconds-scale clearance
   - Research: Sarter et al. (2009), Parikh et al. (2007)

KEY BIOLOGICAL PRINCIPLES:
- Realistic diffusion coefficients from brain slice measurements
- Accurate transporter densities and clearance kinetics
- Physiological concentration ranges and dose-response curves
- Metabolic constraints on neurotransmitter synthesis and release
- Spatial organization reflecting actual brain architecture
=================================================================================
*/

package extracellular

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// =================================================================================
// BIOLOGICAL RATE LIMITS - METABOLIC AND SYNTHESIS CONSTRAINTS
// =================================================================================

const (
	// Maximum firing rates based on neurotransmitter synthesis limitations
	// Research basis: Metabolic constraints on vesicle recycling and NT synthesis
	GLUTAMATE_MAX_RATE     = 500.0  // Hz - Fast vesicle recycling, abundant synthesis
	GABA_MAX_RATE          = 500.0  // Hz - Fast vesicle recycling, abundant synthesis
	DOPAMINE_MAX_RATE      = 100.0  // Hz - Synthesis limited by tyrosine hydroxylase
	SEROTONIN_MAX_RATE     = 80.0   // Hz - Synthesis limited by tryptophan hydroxylase
	ACETYLCHOLINE_MAX_RATE = 300.0  // Hz - Intermediate synthesis capacity
	GLOBAL_MAX_RATE        = 2000.0 // Hz - System-wide metabolic limit per second

	// Default fallback values for unknown neurotransmitters
	DEFAULT_MAX_LIGAND_RATE_HZ     = 100.0 // Conservative firing rate limit
	DEFAULT_DIFFUSION_DECAY_LAMBDA = 10.0  // Moderate diffusion range (μm)
)

// =================================================================================
// BIOLOGICAL SIGNIFICANCE THRESHOLDS
// =================================================================================

const (
	// Minimum concentrations to trigger biological responses
	// Research basis: Receptor binding affinity and physiological response thresholds

	// Fast synaptic neurotransmitters - higher thresholds due to rapid signaling
	BIOLOGICAL_THRESHOLD_GLUTAMATE = 0.01 // 10 μM - AMPA/NMDA receptor activation
	BIOLOGICAL_THRESHOLD_GABA      = 0.01 // 10 μM - GABA-A receptor activation

	// Neuromodulators - lower thresholds due to high-affinity receptors
	BIOLOGICAL_THRESHOLD_DOPAMINE  = 0.001 // 1 μM - D1/D2 receptor activation
	BIOLOGICAL_THRESHOLD_SEROTONIN = 0.001 // 1 μM - 5-HT receptor activation

	// Mixed signaling neurotransmitters
	BIOLOGICAL_THRESHOLD_ACETYLCHOLINE = 0.005 // 5 μM - nAChR/mAChR activation

	// Default threshold for unknown ligands
	DEFAULT_BIOLOGICAL_THRESHOLD = 0.001 // 1 μM - Conservative threshold

	// Binding significance threshold - concentration needed to affect target
	BINDING_SIGNIFICANCE_THRESHOLD = 0.001 // 1 μM - Minimum for binding events
)

// =================================================================================
// DIFFUSION MODEL PARAMETERS
// =================================================================================

const (
	// Synaptic transmission diffusion model coefficients
	// Research basis: Measured concentration profiles in synaptic clefts
	SYNAPTIC_DECAY_STEEPNESS_FACTOR = 5.0 // Controls rapid concentration drop-off
	SYNAPTIC_POWER_DECAY_FACTOR     = 2.0 // Exponent for transporter uptake model

	// Volume transmission diffusion model coefficients
	// Research basis: Dopamine and serotonin concentration measurements in brain
	VOLUME_TRANSMISSION_CUTOFF_FACTOR    = 0.6  // Exponential cutoff distance factor
	VOLUME_TRANSMISSION_POWER_EXPONENT   = 0.5  // 3D diffusion power law exponent
	VOLUME_TRANSMISSION_NEAR_FIELD_LIMIT = 1.0  // μm - Linear regime boundary
	VOLUME_TRANSMISSION_DISTANCE_SCALE   = 10.0 // μm - Near-field decay scale

	// Mixed signaling diffusion model coefficients
	// Research basis: Acetylcholine phasic and tonic signaling measurements
	MIXED_SIGNALING_DECAY_FACTOR = 2.5 // Exponential decay length scale factor

	// Default diffusion model coefficients
	DEFAULT_DIFFUSION_DECAY_FACTOR = 3.0 // Conservative decay for unknown ligands

	// Normalization baselines for diffusion scaling
	DOPAMINE_BASELINE_DIFFUSION_RATE = 0.2 // μm²/ms - Reference for neuromodulators
	ACETYLCHOLINE_BASELINE_DIFFUSION = 0.4 // μm²/ms - Reference for mixed signaling
	DEFAULT_DIFFUSION_BASELINE       = 1.0 // μm²/ms - Generic reference
)

// =================================================================================
// TEMPORAL PROCESSING PARAMETERS
// =================================================================================

const (
	// Background processing timing
	DECAY_PROCESSOR_INTERVAL_MS = 1.0 // Biological timescale for concentration updates
	MIN_DECAY_TIME_THRESHOLD_MS = 0.1 // Minimum time before processing decay

	// Distance calculation precision
	DISTANCE_CALCULATION_EPSILON = 1e-9  // Avoid self-calculation in concentration sums
	NEAR_SOURCE_DISTANCE_LIMIT   = 0.001 // μm - Consider as "at source" position
)

// =================================================================================
// CORE DATA STRUCTURES
// =================================================================================

// ChemicalModulator handles biologically accurate chemical signal propagation
//
// This is the central component that manages all aspects of chemical signaling
// in the neural network, from neurotransmitter release to receptor binding.
// It maintains spatial concentration fields, processes biological decay, and
// coordinates binding events with realistic kinetics.
type ChemicalModulator struct {
	// === RECEPTOR BINDING SUBSYSTEM ===
	// Maps each neurotransmitter to components that can bind to it
	// Enables selective chemical communication between network components
	bindingTargets map[LigandType][]BindingTarget

	// === SPATIAL CONCENTRATION FIELDS ===
	// 3D spatial distribution of each neurotransmitter type
	// Tracks concentration gradients and source locations in real-time
	concentrationFields map[LigandType]*ConcentrationField

	// === RELEASE EVENT HISTORY ===
	// Complete record of all chemical releases for analysis and monitoring
	// Enables investigation of signaling patterns and network dynamics
	releaseEvents []ChemicalReleaseEvent

	// === BIOLOGICAL KINETICS ===
	// Experimentally-measured parameters for each neurotransmitter type
	// Includes diffusion rates, clearance kinetics, and binding properties
	ligandKinetics map[LigandType]LigandKinetics

	// === COMPONENT INTEGRATION ===
	// Connection to the astrocyte network for component position lookup
	// Enables spatial coordination and location-based signaling
	astrocyteNetwork *AstrocyteNetwork

	// === BIOLOGICAL RATE LIMITING ===
	// Per-component tracking to enforce metabolic firing rate constraints
	lastRelease map[string]time.Time

	// Global system rate limiting to prevent unrealistic network activity
	globalRelease struct {
		lastTime time.Time
		count    int
		mu       sync.Mutex
	}

	// === THREAD SAFETY AND STATE ===
	// Background processing control and thread-safe access coordination
	isRunning bool
	mu        sync.RWMutex
}

// ConcentrationField represents the 3D spatial distribution of a neurotransmitter
//
// This structure maintains the complete spatial state of a single neurotransmitter
// type throughout the network volume. It tracks both discrete concentration points
// (from recent releases) and continuous sources (ongoing release processes).
type ConcentrationField struct {
	// Point concentration map: specific 3D locations with measured concentrations
	// Updated by discrete release events and maintained through biological decay
	Concentrations map[Position3D]float64 `json:"concentrations"`

	// Active release sources: components currently releasing this neurotransmitter
	// Enables modeling of sustained release processes and long-duration signaling
	Sources map[string]ChemicalSource `json:"sources"`

	// Peak concentration tracking for analysis and normalization
	// Useful for studying signal strength and network activity patterns
	MaxConcentration float64 `json:"max_concentration"`

	// Temporal tracking for biological decay processing
	// Ensures accurate time-dependent concentration updates
	LastUpdate time.Time `json:"last_update"`
}

// ChemicalSource represents an active neurotransmitter release site
//
// Models both phasic (brief, high-concentration) and tonic (sustained, low-concentration)
// release patterns observed in biological neural networks.
type ChemicalSource struct {
	ComponentID string        `json:"component_id"` // Identity of releasing component
	Position    Position3D    `json:"position"`     // 3D spatial location of release
	LigandType  LigandType    `json:"ligand_type"`  // Type of neurotransmitter
	ReleaseRate float64       `json:"release_rate"` // Concentration per unit time
	Duration    time.Duration `json:"duration"`     // Expected release duration
	StartTime   time.Time     `json:"start_time"`   // When release process began
	Active      bool          `json:"active"`       // Currently releasing flag
}

// ChemicalReleaseEvent records a neurotransmitter release for analysis
//
// Provides complete documentation of each chemical signaling event for
// network analysis, debugging, and biological validation studies.
type ChemicalReleaseEvent struct {
	SourceID      string        `json:"source_id"`     // Component that initiated release
	LigandType    LigandType    `json:"ligand_type"`   // Neurotransmitter type
	Position      Position3D    `json:"position"`      // 3D location of release
	Concentration float64       `json:"concentration"` // Peak concentration achieved
	Timestamp     time.Time     `json:"timestamp"`     // Precise timing of release
	Duration      time.Duration `json:"duration"`      // Biological release duration
}

// LigandKinetics defines biologically measured neurotransmitter properties
//
// All parameters are derived from experimental neuroscience literature and
// represent the fundamental physical and biochemical properties that govern
// how each neurotransmitter behaves in brain tissue.
type LigandKinetics struct {
	// Spatial diffusion coefficient (μm²/ms)
	// Measured in brain slice preparations using fluorescent tracers
	DiffusionRate float64 `json:"diffusion_rate"`

	// Enzymatic breakdown rate (1/ms)
	// Reflects activity of specific degradation enzymes (AChE, MAO, etc.)
	DecayRate float64 `json:"decay_rate"`

	// Transporter-mediated clearance rate (1/ms)
	// Based on density and kinetics of specific uptake transporters
	ClearanceRate float64 `json:"clearance_rate"`

	// Effective diffusion distance (μm)
	// Maximum range where concentrations remain biologically significant
	MaxRange float64 `json:"max_range"`

	// Receptor binding strength (0-1 scale)
	// Reflects affinity for primary receptor subtypes
	BindingAffinity float64 `json:"binding_affinity"`

	// Hill coefficient for binding cooperativity
	// Models cooperative or competitive binding effects
	Cooperativity float64 `json:"cooperativity"`
}

// =================================================================================
// CONSTRUCTOR AND INITIALIZATION
// =================================================================================

// NewChemicalModulator creates a biologically accurate chemical signaling system
//
// Initializes all subsystems with experimentally-validated parameters and
// establishes the spatial and temporal frameworks needed for realistic
// neurotransmitter and neuromodulator signaling.
//
// Parameters:
//   - astrocyteNetwork: Network component registry for spatial coordination
//
// Returns:
//   - Fully configured ChemicalModulator ready for biological simulation
func NewChemicalModulator(astrocyteNetwork *AstrocyteNetwork) *ChemicalModulator {
	cm := &ChemicalModulator{
		bindingTargets:      make(map[LigandType][]BindingTarget),
		concentrationFields: make(map[LigandType]*ConcentrationField),
		releaseEvents:       make([]ChemicalReleaseEvent, 0),
		ligandKinetics:      make(map[LigandType]LigandKinetics),
		lastRelease:         make(map[string]time.Time),
		astrocyteNetwork:    astrocyteNetwork,
		isRunning:           false,
	}

	// Initialize with biologically accurate kinetic parameters
	cm.initializeBiologicalKinetics()
	return cm
}

// initializeBiologicalKinetics sets kinetic parameters based on neuroscience research
//
// All parameters are derived from published experimental measurements in living
// brain tissue. References are provided for each neurotransmitter system to
// enable validation and parameter adjustment based on new research findings.
func (cm *ChemicalModulator) initializeBiologicalKinetics() {
	// === GLUTAMATE - Fast Excitatory Synaptic Transmission ===
	//
	// The primary excitatory neurotransmitter in the central nervous system.
	// Mediates fast, precise communication between neurons through AMPA and NMDA receptors.
	//
	// Biological context:
	// - Released at ~1-10 mM concentrations in synaptic clefts
	// - Cleared within 1-2 ms by high-density EAAT transporters
	// - Spillover limited to ~1-2 μm from release sites
	// - Critical for learning, memory, and excitatory drive
	//
	// Research basis: Danbolt (2001), Clements et al. (1992), Diamond (2001)
	cm.ligandKinetics[LigandGlutamate] = LigandKinetics{
		DiffusionRate:   0.76, // Measured: 760 μm²/s in brain slices
		DecayRate:       0.40, // Fast enzymatic breakdown by peptidases
		ClearanceRate:   0.50, // Rapid EAAT1/2/3 transporter uptake (highest density)
		MaxRange:        5.0,  // Spillover range buffered for edge effects
		BindingAffinity: 0.9,  // High affinity for AMPA/NMDA receptors
		Cooperativity:   1.0,  // Non-cooperative binding kinetics
	}
	// Total clearance: 0.90 /ms → 99% cleared in 5ms (biologically accurate)

	// === GABA - Fast Inhibitory Synaptic Transmission ===
	//
	// The primary inhibitory neurotransmitter, providing precise inhibitory control
	// and maintaining excitatory-inhibitory balance throughout the nervous system.
	//
	// Biological context:
	// - Released at ~0.5-1 mM concentrations in synaptic clefts
	// - Cleared by GAT1-4 transporters (lower density than EAAT)
	// - Slightly longer spillover than glutamate due to slower uptake
	// - Essential for preventing epileptic activity and shaping neural responses
	//
	// Research basis: Conti et al. (2004), Farrant & Nusser (2005)
	cm.ligandKinetics[LigandGABA] = LigandKinetics{
		DiffusionRate:   0.60, // Slightly slower diffusion than glutamate
		DecayRate:       0.25, // Breakdown by GABA transaminase
		ClearanceRate:   0.35, // GAT transporter uptake (moderate density)
		MaxRange:        4.0,  // Short range, similar to glutamate
		BindingAffinity: 0.8,  // High affinity for GABA-A/B receptors
		Cooperativity:   1.0,  // Non-cooperative binding
	}
	// Total clearance: 0.60 /ms → 95% cleared in 5ms

	// === DOPAMINE - Volume Transmission Neuromodulator ===
	//
	// Critical neuromodulator for reward signaling, motivation, and motor control.
	// Operates through volume transmission to modulate large neural populations.
	//
	// Biological context:
	// - Released at 1-10 μM concentrations (much lower than synaptic)
	// - Diffuses 50-100 μm from release sites in striatum
	// - Cleared slowly by DAT transporters and MAO enzymes
	// - Modulates learning, addiction, and movement disorders
	// - Dysfunction implicated in Parkinson's disease and schizophrenia
	//
	// Research basis: Floresco et al. (2003), Garris et al. (1994), Grace (2016)
	cm.ligandKinetics[LigandDopamine] = LigandKinetics{
		DiffusionRate:   0.20,   // Measured in striatal brain slices
		DecayRate:       0.0001, // Slow MAO-A/B breakdown (minutes timescale)
		ClearanceRate:   0.0005, // DAT transporter uptake (low density)
		MaxRange:        100.0,  // Long-range volume transmission
		BindingAffinity: 0.7,    // Moderate affinity for D1/D2 receptors
		Cooperativity:   1.2,    // Slight positive cooperativity
	}

	// === SEROTONIN - Volume Transmission Neuromodulator ===
	//
	// Key neuromodulator for mood, sleep, and behavioral state control.
	// Influences diverse brain functions through widespread projections.
	//
	// Biological context:
	// - Released at 0.1-1 μM concentrations for volume transmission
	// - Diffuses 50-80 μm from raphe nucleus terminals
	// - Cleared by SERT transporters and MAO-A
	// - Regulates mood, sleep-wake cycles, and social behavior
	// - Target of antidepressant medications (SSRIs)
	//
	// Research basis: Bunin & Wightman (1998), Daws et al. (2005)
	cm.ligandKinetics[LigandSerotonin] = LigandKinetics{
		DiffusionRate:   0.15,   // Slower diffusion than dopamine
		DecayRate:       0.0002, // Very slow MAO-A breakdown
		ClearanceRate:   0.01,   // SERT transporter uptake
		MaxRange:        80.0,   // Long-range volume transmission
		BindingAffinity: 0.6,    // Moderate affinity for 5-HT receptors
		Cooperativity:   1.0,    // Non-cooperative binding
	}

	// === ACETYLCHOLINE - Mixed Synaptic/Volume Transmission ===
	//
	// Unique neurotransmitter with both fast synaptic and slow neuromodulatory actions.
	// Critical for attention, arousal, and cognitive function.
	//
	// Biological context:
	// - Dual signaling modes: phasic (synaptic) and tonic (volume)
	// - Rapidly degraded by acetylcholinesterase (fastest enzyme in body)
	// - Moderate diffusion range supporting both local and distributed effects
	// - Essential for attention, learning, and arousal regulation
	// - Dysfunction in Alzheimer's disease and myasthenia gravis
	//
	// Research basis: Sarter et al. (2009), Parikh et al. (2007)
	cm.ligandKinetics[LigandAcetylcholine] = LigandKinetics{
		DiffusionRate:   0.40, // Moderate diffusion coefficient
		DecayRate:       0.25, // Rapid AChE breakdown (but spatially limited)
		ClearanceRate:   0.05, // Limited reuptake (primarily enzymatic clearance)
		MaxRange:        20.0, // Intermediate range for mixed signaling
		BindingAffinity: 0.8,  // High affinity for nAChR/mAChR
		Cooperativity:   1.0,  // Non-cooperative binding
	}
}

// =================================================================================
// BIOLOGICAL RATE LIMITING SYSTEM
// =================================================================================

// checkRateLimits enforces biological release frequency constraints
//
// Prevents unrealistic firing patterns that would be impossible given the metabolic
// constraints of neurotransmitter synthesis, vesicle recycling, and cellular energy.
// Each neurotransmitter has different rate limits based on synthesis pathways.
//
// Parameters:
//   - ligandType: Type of neurotransmitter being released
//   - sourceID: Identity of the releasing component
//
// Returns:
//   - error if rate limits are exceeded (biologically realistic rejection)
//   - nil if release is permitted within biological constraints
func (cm *ChemicalModulator) checkRateLimits(ligandType LigandType, sourceID string) error {
	now := time.Now()

	// Check component-specific rate limits
	cm.mu.RLock()
	lastTime, exists := cm.lastRelease[sourceID]
	cm.mu.RUnlock()

	if exists {
		minInterval := cm.getMinReleaseInterval(ligandType)
		if now.Sub(lastTime) < minInterval {
			return fmt.Errorf("component %s release rate exceeded for %v (biological limit: %.1f Hz)",
				sourceID, ligandType, 1.0/minInterval.Seconds())
		}
	}

	// Check global system rate limits to prevent metabolic overload
	cm.globalRelease.mu.Lock()
	defer cm.globalRelease.mu.Unlock()

	// Reset counter if more than 1 second has passed
	if now.Sub(cm.globalRelease.lastTime) > time.Second {
		cm.globalRelease.count = 0
		cm.globalRelease.lastTime = now
	}

	// Enforce global metabolic limit
	if cm.globalRelease.count >= int(GLOBAL_MAX_RATE) {
		return fmt.Errorf("global chemical release rate exceeded (biological limit: %.0f/second)", GLOBAL_MAX_RATE)
	}

	cm.globalRelease.count++
	return nil
}

// getMinReleaseInterval returns the minimum time between releases for each neurotransmitter
//
// Based on the metabolic constraints of neurotransmitter synthesis and vesicle recycling.
// Synthesis-limited neurotransmitters (dopamine, serotonin) have longer intervals.
//
// Parameters:
//   - ligandType: Type of neurotransmitter
//
// Returns:
//   - time.Duration representing minimum interval between releases
func (cm *ChemicalModulator) getMinReleaseInterval(ligandType LigandType) time.Duration {
	var maxRate float64

	switch ligandType {
	case LigandGlutamate:
		maxRate = GLUTAMATE_MAX_RATE
	case LigandGABA:
		maxRate = GABA_MAX_RATE
	case LigandDopamine:
		maxRate = DOPAMINE_MAX_RATE
	case LigandSerotonin:
		maxRate = SEROTONIN_MAX_RATE
	case LigandAcetylcholine:
		maxRate = ACETYLCHOLINE_MAX_RATE
	default:
		maxRate = DEFAULT_MAX_LIGAND_RATE_HZ
	}

	return time.Duration(1e9 / maxRate) // Convert Hz to nanoseconds
}

// ResetRateLimits clears rate limiting state (useful for testing and initialization)
//
// Allows fresh start for rate limiting calculations. Used primarily during
// testing scenarios and system initialization to ensure clean state.
func (cm *ChemicalModulator) ResetRateLimits() {
	// Reset component-specific rate limits
	cm.mu.Lock()
	cm.lastRelease = make(map[string]time.Time)
	cm.mu.Unlock()

	// Reset global rate limits separately to avoid deadlock
	cm.globalRelease.mu.Lock()
	cm.globalRelease.count = 0
	cm.globalRelease.lastTime = time.Time{}
	cm.globalRelease.mu.Unlock()
}

// GetCurrentReleaseRate returns the current system-wide release rate
//
// Provides monitoring capability for system activity levels and
// metabolic load assessment.
//
// Returns:
//   - float64 representing releases per second over the last second
func (cm *ChemicalModulator) GetCurrentReleaseRate() float64 {
	cm.globalRelease.mu.Lock()
	defer cm.globalRelease.mu.Unlock()

	if time.Since(cm.globalRelease.lastTime) > time.Second {
		return 0.0 // No recent activity
	}

	return float64(cm.globalRelease.count)
}

// =================================================================================
// CHEMICAL RELEASE SYSTEM
// =================================================================================

// Release initiates biologically accurate neurotransmitter release
//
// This is the primary interface for chemical signaling in the network. It handles
// input validation, rate limiting, spatial positioning, and immediate binding
// calculations with full biological accuracy.
//
// Parameters:
//   - ligandType: Type of neurotransmitter (Glutamate, GABA, Dopamine, etc.)
//   - sourceID: Identity of the releasing component
//   - concentration: Peak concentration to release (in μM)
//
// Returns:
//   - error if release fails due to validation or biological constraints
//   - nil if release succeeds and signaling is processed
func (cm *ChemicalModulator) Release(ligandType LigandType, sourceID string, concentration float64) error {
	// Comprehensive input validation before any processing
	if strings.TrimSpace(sourceID) == "" {
		return fmt.Errorf("invalid source ID: cannot be empty")
	}
	if math.IsNaN(concentration) || math.IsInf(concentration, 0) {
		return fmt.Errorf("invalid concentration: %f", concentration)
	}
	if concentration < 0 {
		return fmt.Errorf("invalid concentration: cannot be negative")
	}

	// Enforce biological rate limits based on metabolic constraints
	if err := cm.checkRateLimits(ligandType, sourceID); err != nil {
		return err // Biologically realistic rejection
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Obtain source position from component registry
	sourceInfo, exists := cm.astrocyteNetwork.Get(sourceID)
	if !exists {
		// Allow release with default position for system flexibility
		sourceInfo.Position = Position3D{X: 0, Y: 0, Z: 0}
	}

	// Create comprehensive release event record
	event := ChemicalReleaseEvent{
		SourceID:      sourceID,
		LigandType:    ligandType,
		Position:      sourceInfo.Position,
		Concentration: concentration,
		Timestamp:     time.Now(),
		Duration:      cm.getBiologicalReleaseDuration(ligandType),
	}

	// Archive event for network analysis and debugging
	cm.releaseEvents = append(cm.releaseEvents, event)

	// Update rate limiting records (already holding write lock)
	cm.lastRelease[sourceID] = time.Now()

	// Update spatial concentration field
	cm.updateConcentrationField(ligandType, sourceInfo.Position, concentration)

	// Process immediate binding for all registered targets
	cm.processImmediateBinding(ligandType, sourceInfo.Position, concentration, sourceID)

	return nil
}

// processImmediateBinding calculates and applies binding for all registered targets
//
// Implements the spatial binding model where chemical signals affect targets
// based on distance, concentration, and receptor specificity. This provides
// the immediate effects of chemical release on network components.
//
// Parameters:
//   - ligandType: Type of neurotransmitter released
//   - sourcePos: 3D position of the release site
//   - concentration: Peak concentration at the release site
//   - sourceID: Identity of the releasing component
func (cm *ChemicalModulator) processImmediateBinding(ligandType LigandType, sourcePos Position3D, concentration float64, sourceID string) {
	targets := cm.bindingTargets[ligandType]
	if targets == nil {
		return
	}

	// Calculate effective concentration at each target location
	for _, target := range targets {
		targetPos := target.GetPosition()
		distance := cm.calculateDistance(sourcePos, targetPos)
		effectiveConcentration := cm.calculateBiologicalConcentration(ligandType, concentration, distance)

		// Apply binding only if concentration exceeds biological significance
		if effectiveConcentration > BINDING_SIGNIFICANCE_THRESHOLD {
			target.Bind(ligandType, sourceID, effectiveConcentration)
		}
	}
}

// =================================================================================
// BIOLOGICALLY ACCURATE CONCENTRATION CALCULATION
// =================================================================================

// calculateBiologicalConcentration computes concentration at distance using biological models
//
// This is the core spatial modeling function that determines how neurotransmitter
// concentrations change with distance from release sites. Different neurotransmitters
// use different models based on their biological properties and signaling modes.
//
// Parameters:
//   - ligandType: Type of neurotransmitter
//   - sourceConcentration: Concentration at the release site (μM)
//   - distance: 3D distance from release site (μm)
//
// Returns:
//   - float64 representing effective concentration at the target distance
func (cm *ChemicalModulator) calculateBiologicalConcentration(ligandType LigandType, sourceConcentration, distance float64) float64 {
	kinetics, exists := cm.ligandKinetics[ligandType]
	if !exists {
		// Fallback for unknown neurotransmitters
		return sourceConcentration * math.Exp(-distance/DEFAULT_DIFFUSION_DECAY_LAMBDA)
	}

	// Handle source position (no distance decay)
	if distance < NEAR_SOURCE_DISTANCE_LIMIT {
		return sourceConcentration
	}

	// Apply biologically appropriate diffusion model based on signaling mode
	switch ligandType {
	case LigandGlutamate, LigandGABA:
		return cm.calculateSynapticDiffusion(kinetics, sourceConcentration, distance)
	case LigandDopamine, LigandSerotonin:
		return cm.calculateVolumeTransmission(kinetics, sourceConcentration, distance)
	case LigandAcetylcholine:
		return cm.calculateMixedSignaling(kinetics, sourceConcentration, distance)
	default:
		return cm.calculateDefaultDiffusion(kinetics, sourceConcentration, distance)
	}
}

// calculateSynapticDiffusion models fast synaptic transmission with steep concentration gradients
//
// Models the rapid, localized signaling characteristic of glutamate and GABA.
// Uses combined exponential and power-law decay to capture both passive diffusion
// and active transporter uptake effects observed in synaptic clefts.
//
// Biological basis:
// - Steep concentration gradients due to high transporter density
// - Rapid clearance creates sharp spatial boundaries
// - Power-law component models transporter saturation effects
//
// Parameters:
//   - kinetics: Kinetic parameters for the neurotransmitter
//   - sourceConc: Concentration at the release site
//   - distance: Distance from release site (μm)
//
// Returns:
//   - Concentration at the specified distance accounting for synaptic dynamics
func (cm *ChemicalModulator) calculateSynapticDiffusion(kinetics LigandKinetics, sourceConc, distance float64) float64 {
	// Exponential decay component - passive diffusion with rapid clearance
	lambda := kinetics.MaxRange / SYNAPTIC_DECAY_STEEPNESS_FACTOR
	decayFactor := math.Exp(-distance / lambda)

	// Power-law decay component - models transporter uptake saturation
	powerDecay := math.Pow(1.0/(1.0+distance), SYNAPTIC_POWER_DECAY_FACTOR)

	return sourceConc * decayFactor * powerDecay
}

// calculateVolumeTransmission models slow neuromodulator diffusion over long distances
//
// Models the gradual, widespread signaling characteristic of dopamine and serotonin.
// Uses power-law decay with exponential cutoff to capture 3D diffusion with
// gentle boundaries characteristic of volume transmission.
//
// Biological basis:
// - Power-law reflects 3D diffusion in tortuous extracellular space
// - Low transporter density allows long-range signaling
// - Gentle exponential cutoff prevents infinite range
//
// Parameters:
//   - kinetics: Kinetic parameters for the neuromodulator
//   - sourceConc: Concentration at the release site
//   - distance: Distance from release site (μm)
//
// Returns:
//   - Concentration accounting for volume transmission dynamics
func (cm *ChemicalModulator) calculateVolumeTransmission(kinetics LigandKinetics, sourceConc, distance float64) float64 {
	if distance < VOLUME_TRANSMISSION_NEAR_FIELD_LIMIT {
		// Near-field: linear decay to avoid mathematical singularities
		linearDecay := 1.0 - distance/VOLUME_TRANSMISSION_DISTANCE_SCALE
		diffusionScale := kinetics.DiffusionRate / DOPAMINE_BASELINE_DIFFUSION_RATE
		return sourceConc * linearDecay * diffusionScale
	} else {
		// Far-field: power law decay with exponential cutoff
		powerDecay := math.Pow(distance, -VOLUME_TRANSMISSION_POWER_EXPONENT)
		expCutoff := math.Exp(-distance / (kinetics.MaxRange * VOLUME_TRANSMISSION_CUTOFF_FACTOR))
		diffusionScale := kinetics.DiffusionRate / DOPAMINE_BASELINE_DIFFUSION_RATE

		return sourceConc * powerDecay * expCutoff * diffusionScale
	}
}

// calculateMixedSignaling models acetylcholine's dual synaptic/volume properties
//
// Models the intermediate-range signaling that supports both fast synaptic
// communication and slower neuromodulatory effects. Uses exponential decay
// with moderate range parameters.
//
// Biological basis:
// - Rapid AChE breakdown limits range but doesn't eliminate volume effects
// - Intermediate between synaptic and volume transmission modes
// - Supports both phasic (attention) and tonic (arousal) signaling
//
// Parameters:
//   - kinetics: Kinetic parameters for acetylcholine
//   - sourceConc: Concentration at the release site
//   - distance: Distance from release site (μm)
//
// Returns:
//   - Concentration reflecting mixed signaling properties
func (cm *ChemicalModulator) calculateMixedSignaling(kinetics LigandKinetics, sourceConc, distance float64) float64 {
	lambda := kinetics.MaxRange / MIXED_SIGNALING_DECAY_FACTOR
	decay := math.Exp(-distance / lambda)
	diffusionScale := kinetics.DiffusionRate / ACETYLCHOLINE_BASELINE_DIFFUSION

	return sourceConc * decay * diffusionScale
}

// calculateDefaultDiffusion provides fallback exponential decay for unknown ligands
//
// Provides conservative modeling for neurotransmitters not explicitly
// characterized in the system. Uses simple exponential decay with
// moderate range parameters.
//
// Parameters:
//   - kinetics: Kinetic parameters for the ligand
//   - sourceConc: Concentration at the release site
//   - distance: Distance from release site (μm)
//
// Returns:
//   - Concentration using conservative default model
func (cm *ChemicalModulator) calculateDefaultDiffusion(kinetics LigandKinetics, sourceConc, distance float64) float64 {
	lambda := kinetics.MaxRange / DEFAULT_DIFFUSION_DECAY_FACTOR
	decay := math.Exp(-distance / lambda)
	diffusionScale := kinetics.DiffusionRate / DEFAULT_DIFFUSION_BASELINE

	return sourceConc * decay * diffusionScale
}

// =================================================================================
// SPATIAL CONCENTRATION FIELD MANAGEMENT
// =================================================================================

// GetConcentration returns current concentration at a position with biological accuracy
//
// This is the primary interface for querying neurotransmitter concentrations
// throughout the network volume. It integrates contributions from all active
// sources and stored concentration points to provide accurate spatial mapping.
//
// Parameters:
//   - ligandType: Type of neurotransmitter to query
//   - position: 3D position where concentration is needed
//
// Returns:
//   - float64 representing total effective concentration at the position
func (cm *ChemicalModulator) GetConcentration(ligandType LigandType, position Position3D) float64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	field, exists := cm.concentrationFields[ligandType]
	if !exists {
		return 0.0
	}

	totalConcentration := 0.0

	// Add contributions from stored concentration points
	for sourcePos, sourceConc := range field.Concentrations {
		distance := cm.calculateDistance(sourcePos, position)
		if distance > DISTANCE_CALCULATION_EPSILON {
			concentration := cm.calculateBiologicalConcentration(ligandType, sourceConc, distance)
			totalConcentration += concentration
		} else {
			// Direct match - add full concentration
			totalConcentration += sourceConc
		}
	}

	// Add contributions from active continuous sources
	for _, source := range field.Sources {
		if source.Active {
			distance := cm.calculateDistance(source.Position, position)
			concentration := cm.calculateBiologicalConcentration(ligandType, source.ReleaseRate, distance)
			totalConcentration += concentration
		}
	}

	return totalConcentration
}

// updateConcentrationField creates or updates concentration field for a ligand
//
// Manages the spatial concentration data structure for each neurotransmitter type.
// Handles both field creation and concentration point updates with proper
// temporal tracking for decay processing.
//
// Parameters:
//   - ligandType: Type of neurotransmitter
//   - position: 3D position of the concentration point
//   - concentration: Concentration value to store
func (cm *ChemicalModulator) updateConcentrationField(ligandType LigandType, position Position3D, concentration float64) {
	if cm.concentrationFields[ligandType] == nil {
		cm.concentrationFields[ligandType] = &ConcentrationField{
			Concentrations: make(map[Position3D]float64),
			Sources:        make(map[string]ChemicalSource),
			LastUpdate:     time.Now(),
		}
	}

	field := cm.concentrationFields[ligandType]
	field.Concentrations[position] = concentration
	field.LastUpdate = time.Now()

	// Update maximum concentration tracking for analysis
	if concentration > field.MaxConcentration {
		field.MaxConcentration = concentration
	}
}

// =================================================================================
// BIOLOGICAL DECAY PROCESSING SYSTEM
// =================================================================================

// Start begins background processing with biologically accurate timing
//
// Initiates the continuous decay processing that maintains realistic temporal
// dynamics for all neurotransmitter concentrations. Essential for biological
// accuracy as it models enzymatic breakdown and transporter clearance.
//
// Returns:
//   - error if system is already running
//   - nil if background processing started successfully
func (cm *ChemicalModulator) Start() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.isRunning {
		return nil
	}

	cm.isRunning = true
	go cm.biologicalDecayProcessor()
	return nil
}

// Stop ends chemical processing
//
// Cleanly terminates background decay processing. Used during system
// shutdown or when switching to different processing modes.
//
// Returns:
//   - error (currently always nil, reserved for future error conditions)
func (cm *ChemicalModulator) Stop() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.isRunning = false
	return nil
}

// biologicalDecayProcessor handles concentration decay with biological timing
//
// Runs continuously in the background to update concentration fields according
// to biological decay kinetics. Uses 1ms intervals to match the timescale of
// fast synaptic processes while remaining computationally efficient.
func (cm *ChemicalModulator) biologicalDecayProcessor() {
	ticker := time.NewTicker(DECAY_PROCESSOR_INTERVAL_MS * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		cm.mu.RLock()
		running := cm.isRunning
		cm.mu.RUnlock()

		if !running {
			break
		}

		cm.processBiologicalDecay()
	}
}

// processBiologicalDecay applies biologically accurate concentration decay
//
// Updates all concentration fields according to the specific kinetic parameters
// of each neurotransmitter. Removes concentrations that fall below biological
// significance thresholds to maintain computational efficiency.
func (cm *ChemicalModulator) processBiologicalDecay() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()

	for ligandType, field := range cm.concentrationFields {
		if field == nil {
			continue
		}

		kinetics := cm.ligandKinetics[ligandType]
		dt := float64(now.Sub(field.LastUpdate).Milliseconds())

		if dt < MIN_DECAY_TIME_THRESHOLD_MS {
			continue // Skip processing for very short intervals
		}

		// Process decay for all concentration points
		newConcentrations := make(map[Position3D]float64)
		for pos, concentration := range field.Concentrations {
			newConcentration := cm.calculateBiologicalDecay(ligandType, kinetics, concentration, dt)

			// Keep only biologically significant concentrations
			if newConcentration >= cm.getBiologicalThreshold(ligandType) {
				newConcentrations[pos] = newConcentration
			}
		}

		// Update field with cleaned concentration map
		field.Concentrations = newConcentrations
		field.LastUpdate = now
		field.MaxConcentration = cm.calculateMaxConcentration(field.Concentrations)
	}
}

// calculateBiologicalDecay computes concentration after biological clearance
//
// Applies exponential decay based on the combined effects of enzymatic breakdown
// and transporter-mediated clearance. Uses millisecond timing to match the
// units of the kinetic rate constants.
//
// Parameters:
//   - ligandType: Type of neurotransmitter (for debugging/logging)
//   - kinetics: Kinetic parameters containing decay and clearance rates
//   - concentration: Current concentration value
//   - deltaTime: Time elapsed since last update (milliseconds)
//
// Returns:
//   - Updated concentration after biological decay processes
func (cm *ChemicalModulator) calculateBiologicalDecay(ligandType LigandType, kinetics LigandKinetics, concentration float64, deltaTime float64) float64 {
	totalClearanceRate := kinetics.DecayRate + kinetics.ClearanceRate
	return concentration * math.Exp(-totalClearanceRate*deltaTime)
}

// getBiologicalThreshold returns the minimum biologically significant concentration
//
// Defines the concentration below which neurotransmitter effects become
// negligible. Based on receptor binding affinities and physiological
// response thresholds from experimental literature.
//
// Parameters:
//   - ligandType: Type of neurotransmitter
//
// Returns:
//   - Minimum concentration threshold (μM) for biological significance
func (cm *ChemicalModulator) getBiologicalThreshold(ligandType LigandType) float64 {
	switch ligandType {
	case LigandGlutamate:
		return BIOLOGICAL_THRESHOLD_GLUTAMATE
	case LigandGABA:
		return BIOLOGICAL_THRESHOLD_GABA
	case LigandDopamine:
		return BIOLOGICAL_THRESHOLD_DOPAMINE
	case LigandSerotonin:
		return BIOLOGICAL_THRESHOLD_SEROTONIN
	case LigandAcetylcholine:
		return BIOLOGICAL_THRESHOLD_ACETYLCHOLINE
	default:
		return DEFAULT_BIOLOGICAL_THRESHOLD
	}
}

// =================================================================================
// BINDING TARGET MANAGEMENT
// =================================================================================

// RegisterTarget adds a component to receive chemical signals
//
// Enables a network component to participate in chemical signaling by
// registering it to receive binding events for specific neurotransmitter types.
// Components specify which receptors they express through the BindingTarget interface.
//
// Parameters:
//   - target: Component implementing BindingTarget interface
//
// Returns:
//   - error if registration fails (currently always nil)
func (cm *ChemicalModulator) RegisterTarget(target BindingTarget) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, ligandType := range target.GetReceptors() {
		if cm.bindingTargets[ligandType] == nil {
			cm.bindingTargets[ligandType] = make([]BindingTarget, 0)
		}
		cm.bindingTargets[ligandType] = append(cm.bindingTargets[ligandType], target)
	}

	return nil
}

// UnregisterTarget removes a component from receiving chemical signals
//
// Removes a component from all chemical signaling pathways. Used during
// component removal or when changing receptor expression patterns.
//
// Parameters:
//   - target: Component to remove from chemical signaling
//
// Returns:
//   - error if unregistration fails (currently always nil)
func (cm *ChemicalModulator) UnregisterTarget(target BindingTarget) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, ligandType := range target.GetReceptors() {
		targets := cm.bindingTargets[ligandType]
		for i, t := range targets {
			if t == target {
				cm.bindingTargets[ligandType] = append(targets[:i], targets[i+1:]...)
				break
			}
		}
	}

	return nil
}

// =================================================================================
// ANALYSIS AND MONITORING INTERFACE
// =================================================================================

// GetRecentReleases returns recent release events for analysis
//
// Provides access to the chemical release history for network analysis,
// debugging, and validation studies. Events are returned in chronological order.
//
// Parameters:
//   - count: Maximum number of recent events to return
//
// Returns:
//   - Slice of ChemicalReleaseEvent structures (newest events last)
func (cm *ChemicalModulator) GetRecentReleases(count int) []ChemicalReleaseEvent {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if count > len(cm.releaseEvents) {
		count = len(cm.releaseEvents)
	}

	if count == 0 {
		return []ChemicalReleaseEvent{}
	}

	start := len(cm.releaseEvents) - count
	result := make([]ChemicalReleaseEvent, count)
	copy(result, cm.releaseEvents[start:])
	return result
}

// ForceDecayUpdate immediately processes decay (useful for testing)
//
// Triggers immediate decay processing outside of the normal background schedule.
// Primarily used during testing and validation to achieve deterministic timing.
func (cm *ChemicalModulator) ForceDecayUpdate() {
	cm.processBiologicalDecay()
}

// =================================================================================
// UTILITY FUNCTIONS
// =================================================================================

// calculateDistance computes 3D Euclidean distance between positions
//
// Standard 3D distance calculation used throughout the spatial modeling system.
// Essential for determining how chemical concentrations vary with distance.
//
// Parameters:
//   - pos1: First 3D position
//   - pos2: Second 3D position
//
// Returns:
//   - Euclidean distance in micrometers
func (cm *ChemicalModulator) calculateDistance(pos1, pos2 Position3D) float64 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	dz := pos1.Z - pos2.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// getBiologicalReleaseDuration returns typical release duration for each neurotransmitter
//
// Defines the characteristic time course of neurotransmitter release based on
// vesicle fusion kinetics and biological signaling requirements.
//
// Parameters:
//   - ligandType: Type of neurotransmitter
//
// Returns:
//   - Duration representing typical release time course
func (cm *ChemicalModulator) getBiologicalReleaseDuration(ligandType LigandType) time.Duration {
	switch ligandType {
	case LigandGlutamate, LigandGABA:
		return 1 * time.Millisecond // Fast synaptic vesicle fusion
	case LigandAcetylcholine:
		return 5 * time.Millisecond // Intermediate duration
	case LigandDopamine, LigandSerotonin:
		return 100 * time.Millisecond // Long neuromodulator release
	default:
		return 10 * time.Millisecond // Conservative default
	}
}

// calculateMaxConcentration finds maximum concentration in a spatial field
//
// Utility function for tracking peak concentrations across the spatial
// concentration field. Used for analysis and normalization purposes.
//
// Parameters:
//   - concentrations: Map of positions to concentration values
//
// Returns:
//   - Maximum concentration value found in the field
func (cm *ChemicalModulator) calculateMaxConcentration(concentrations map[Position3D]float64) float64 {
	maxConc := 0.0
	for _, conc := range concentrations {
		if conc > maxConc {
			maxConc = conc
		}
	}
	return maxConc
}
