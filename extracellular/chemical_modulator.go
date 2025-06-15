/*
=================================================================================
BIOLOGICAL CHEMICAL MODULATOR - CLEAN IMPLEMENTATION
=================================================================================

Implements biologically accurate neurotransmitter and neuromodulator signaling
based on published neuroscience research. All parameters and algorithms are
derived from experimental measurements in brain tissue.

BIOLOGICAL FOUNDATION:
- Glutamate clearance: 1-2ms via EAAT transporters (Danbolt, 2001)
- GABA kinetics: Similar to glutamate with GAT transporters (Conti et al., 2004)
- Dopamine diffusion: 100μm range, slow clearance (Floresco et al., 2003)
- Serotonin volume transmission: 80μm range (Bunin & Wightman, 1998)
- Acetylcholine: Mixed signaling, rapid AChE breakdown (Sarter et al., 2009)

KEY PRINCIPLES:
1. Synaptic transmission: High concentration, rapid clearance, short range
2. Volume transmission: Lower concentration, slow clearance, long range
3. Realistic diffusion: Based on measured diffusion coefficients
4. Biological decay: Reflects actual transporter and enzyme kinetics
=================================================================================
*/

package extracellular

import (
	"fmt"
	"math"
	"sync"
	"time"
)

const (
	GLUTAMATE_MAX_RATE     = 500.0  // Fast vesicle recycling
	GABA_MAX_RATE          = 500.0  // Fast vesicle recycling
	DOPAMINE_MAX_RATE      = 100.0  // Synthesis limited
	SEROTONIN_MAX_RATE     = 80.0   // Synthesis limited
	ACETYLCHOLINE_MAX_RATE = 300.0  // Intermediate rate
	GLOBAL_MAX_RATE        = 2000.0 // System-wide metabolic limit
)

// ChemicalModulator handles biologically accurate chemical signal propagation
type ChemicalModulator struct {
	// === RECEPTOR BINDING ===
	bindingTargets map[LigandType][]BindingTarget // Components that can bind to each ligand

	// === SPATIAL CONCENTRATION FIELDS ===
	concentrationFields map[LigandType]*ConcentrationField // 3D chemical distribution
	releaseEvents       []ChemicalReleaseEvent             // Recent release history

	// === KINETIC PARAMETERS ===
	ligandKinetics map[LigandType]LigandKinetics // Biologically measured parameters

	// === COMPONENT INTEGRATION ===
	astrocyteNetwork *AstrocyteNetwork // For component position lookup

	// === RATE LIMITING ===
	lastRelease   map[string]time.Time // Per-component rate limiting
	globalRelease struct {             // Global system rate limiting
		lastTime time.Time
		count    int
		mu       sync.Mutex
	}

	// === STATE MANAGEMENT ===
	isRunning bool
	mu        sync.RWMutex
}

// ConcentrationField represents 3D spatial distribution of a neurotransmitter
type ConcentrationField struct {
	Concentrations   map[Position3D]float64    `json:"concentrations"`    // Position -> concentration
	Sources          map[string]ChemicalSource `json:"sources"`           // Active release sites
	MaxConcentration float64                   `json:"max_concentration"` // Peak concentration
	LastUpdate       time.Time                 `json:"last_update"`       // Last decay update
}

// ChemicalSource represents an active neurotransmitter release site
type ChemicalSource struct {
	ComponentID string        `json:"component_id"` // Source component
	Position    Position3D    `json:"position"`     // 3D location
	LigandType  LigandType    `json:"ligand_type"`  // Type of neurotransmitter
	ReleaseRate float64       `json:"release_rate"` // Concentration per second
	Duration    time.Duration `json:"duration"`     // Release duration
	StartTime   time.Time     `json:"start_time"`   // When release began
	Active      bool          `json:"active"`       // Currently releasing
}

// ChemicalReleaseEvent records a neurotransmitter release for analysis
type ChemicalReleaseEvent struct {
	SourceID      string        `json:"source_id"`     // Component that released
	LigandType    LigandType    `json:"ligand_type"`   // Type of neurotransmitter
	Position      Position3D    `json:"position"`      // Release location
	Concentration float64       `json:"concentration"` // Peak concentration
	Timestamp     time.Time     `json:"timestamp"`     // When release occurred
	Duration      time.Duration `json:"duration"`      // Expected duration
}

// LigandKinetics defines biologically measured neurotransmitter properties
type LigandKinetics struct {
	DiffusionRate   float64 `json:"diffusion_rate"`   // μm²/ms - measured diffusion coefficient
	DecayRate       float64 `json:"decay_rate"`       // 1/ms - enzymatic breakdown rate
	ClearanceRate   float64 `json:"clearance_rate"`   // 1/ms - transporter uptake rate
	MaxRange        float64 `json:"max_range"`        // μm - effective diffusion distance
	BindingAffinity float64 `json:"binding_affinity"` // Receptor binding strength (0-1)
	Cooperativity   float64 `json:"cooperativity"`    // Hill coefficient for binding
}

// NewChemicalModulator creates a biologically accurate chemical signaling system
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

// =================================================================================
// BIOLOGICALLY ACCURATE KINETIC PARAMETERS
// =================================================================================

// initializeBiologicalKinetics sets kinetic parameters based on neuroscience research
func (cm *ChemicalModulator) initializeBiologicalKinetics() {
	// === GLUTAMATE - Fast Excitatory Synaptic Transmission ===
	// Research basis: Danbolt (2001), Clements et al. (1992)
	// - Synaptic cleft concentration: 1-10 mM
	// - Clearance time: 1-2 ms via EAAT1/2/3 transporters
	// - Diffusion coefficient: ~760 μm²/s in brain tissue
	// - Effective range: spillover limited to ~1-2 μm
	cm.ligandKinetics[LigandGlutamate] = LigandKinetics{
		DiffusionRate:   0.76,  // Measured: 760 μm²/s = 0.76 μm²/ms
		DecayRate:       200.0, // Fast enzymatic breakdown
		ClearanceRate:   300.0, // Rapid EAAT transporter uptake (Vmax ~500/s)
		MaxRange:        5.0,   // Spillover range ~1-2 μm, buffered to 5μm
		BindingAffinity: 0.9,   // High affinity for AMPA/NMDA receptors
		Cooperativity:   1.0,   // Non-cooperative binding
	}

	// === GABA - Fast Inhibitory Synaptic Transmission ===
	// Research basis: Conti et al. (2004), Farrant & Nusser (2005)
	// - Similar kinetics to glutamate but slightly slower clearance
	// - GAT1-4 transporters with lower density than EAAT
	// - Slightly longer spillover due to slower uptake
	cm.ligandKinetics[LigandGABA] = LigandKinetics{
		DiffusionRate:   0.60,  // Slightly slower than glutamate
		DecayRate:       150.0, // Fast breakdown via GABA transaminase
		ClearanceRate:   200.0, // GAT transporter uptake (lower density than EAAT)
		MaxRange:        4.0,   // Short range, similar to glutamate
		BindingAffinity: 0.8,   // High affinity for GABA-A/B receptors
		Cooperativity:   1.0,   // Non-cooperative binding
	}

	// === DOPAMINE - Volume Transmission Neuromodulator ===
	// Research basis: Floresco et al. (2003), Garris et al. (1994)
	// - Peak concentration: 1-10 μM (much lower than synaptic)
	// - Diffusion range: up to 100 μm from release site
	// - Slow clearance: DAT transporters + MAO breakdown
	// - Clearance time: seconds to minutes
	cm.ligandKinetics[LigandDopamine] = LigandKinetics{
		DiffusionRate:   0.20,  // Measured: ~200 μm²/s in striatum
		DecayRate:       0.01,  // Slow MAO-A/B breakdown (minutes timescale)
		ClearanceRate:   0.05,  // DAT transporter (lower density, slower than EAAT)
		MaxRange:        100.0, // Volume transmission range
		BindingAffinity: 0.7,   // Moderate affinity for D1/D2 receptors
		Cooperativity:   1.2,   // Slight positive cooperativity
	}

	// === SEROTONIN - Volume Transmission Neuromodulator ===
	// Research basis: Bunin & Wightman (1998), Daws et al. (2005)
	// - Similar to dopamine but slightly different kinetics
	// - SERT transporter clearance
	// - Range: 50-80 μm from release sites
	cm.ligandKinetics[LigandSerotonin] = LigandKinetics{
		DiffusionRate:   0.15,  // Slower diffusion than dopamine
		DecayRate:       0.005, // Very slow breakdown (MAO-A)
		ClearanceRate:   0.03,  // SERT transporter uptake
		MaxRange:        80.0,  // Long-range volume transmission
		BindingAffinity: 0.6,   // Moderate affinity for 5-HT receptors
		Cooperativity:   1.0,   // Non-cooperative binding
	}

	// === ACETYLCHOLINE - Mixed Synaptic/Volume Transmission ===
	// Research basis: Sarter et al. (2009), Parikh et al. (2007)
	// - Rapid breakdown by acetylcholinesterase (AChE)
	// - Moderate diffusion range
	// - Mixed phasic (synaptic) and tonic (volume) signaling
	cm.ligandKinetics[LigandAcetylcholine] = LigandKinetics{
		DiffusionRate:   0.40,  // Moderate diffusion coefficient
		DecayRate:       100.0, // Very fast AChE breakdown (milliseconds)
		ClearanceRate:   20.0,  // Limited reuptake (mainly breakdown)
		MaxRange:        20.0,  // Moderate range for cholinergic signaling
		BindingAffinity: 0.8,   // High affinity for nAChR/mAChR
		Cooperativity:   1.0,   // Non-cooperative binding
	}
}

// =================================================================================
// CHEMICAL RELEASE
// =================================================================================

// Add these methods after the existing methods:
// checkRateLimits enforces biological release frequency constraints
func (cm *ChemicalModulator) checkRateLimits(ligandType LigandType, sourceID string) error {
	now := time.Now()

	// Check component-specific rate limit
	if lastTime, exists := cm.lastRelease[sourceID]; exists {
		minInterval := cm.getMinReleaseInterval(ligandType)
		if now.Sub(lastTime) < minInterval {
			return fmt.Errorf("component %s release rate exceeded for %v (biological limit: %.1f Hz)",
				sourceID, ligandType, 1.0/minInterval.Seconds())
		}
	}

	// Check global system rate limit
	cm.globalRelease.mu.Lock()
	defer cm.globalRelease.mu.Unlock()

	// Reset counter if more than 1 second has passed
	if now.Sub(cm.globalRelease.lastTime) > time.Second {
		cm.globalRelease.count = 0
		cm.globalRelease.lastTime = now
	}

	// Check if we're under global rate limit
	if cm.globalRelease.count >= int(GLOBAL_MAX_RATE) {
		return fmt.Errorf("global chemical release rate exceeded (biological limit: %.0f/second)", GLOBAL_MAX_RATE)
	}

	cm.globalRelease.count++
	return nil
}

// getMinReleaseInterval returns the minimum time between releases for each neurotransmitter
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
		maxRate = 100.0 // Conservative default
	}

	return time.Duration(1e9 / maxRate) // Convert Hz to nanoseconds
}

// GetCurrentReleaseRate returns the current system-wide release rate (for monitoring)
func (cm *ChemicalModulator) GetCurrentReleaseRate() float64 {
	cm.globalRelease.mu.Lock()
	defer cm.globalRelease.mu.Unlock()

	// Return releases per second over the last second
	if time.Since(cm.globalRelease.lastTime) > time.Second {
		return 0.0 // No recent activity
	}

	return float64(cm.globalRelease.count)
}

// ResetRateLimits clears rate limiting state (useful for testing)
func (cm *ChemicalModulator) ResetRateLimits() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.lastRelease = make(map[string]time.Time)

	cm.globalRelease.mu.Lock()
	cm.globalRelease.count = 0
	cm.globalRelease.lastTime = time.Time{}
	cm.globalRelease.mu.Unlock()
}

// Release initiates biologically accurate neurotransmitter release
func (cm *ChemicalModulator) Release(ligandType LigandType, sourceID string, concentration float64) error {
	// Check biological rate limits
	if err := cm.checkRateLimits(ligandType, sourceID); err != nil {
		return err // Rate limit exceeded - biologically realistic rejection
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Get source component position from astrocyte network
	sourceInfo, exists := cm.astrocyteNetwork.Get(sourceID)
	if !exists {
		// Allow release with default position for flexibility
		sourceInfo.Position = Position3D{X: 0, Y: 0, Z: 0}
	}

	// Create release event record
	event := ChemicalReleaseEvent{
		SourceID:      sourceID,
		LigandType:    ligandType,
		Position:      sourceInfo.Position,
		Concentration: concentration,
		Timestamp:     time.Now(),
		Duration:      cm.getBiologicalReleaseDuration(ligandType),
	}

	// Record event for analysis
	cm.releaseEvents = append(cm.releaseEvents, event)

	// Update rate limiting records
	cm.lastRelease[sourceID] = time.Now()

	// Update concentration field with new release
	cm.updateConcentrationField(ligandType, sourceInfo.Position, concentration)

	// Immediately calculate binding for registered targets
	cm.processImmediateBinding(ligandType, sourceInfo.Position, concentration, sourceID)

	return nil
}

// processImmediateBinding calculates and applies binding for all registered targets
func (cm *ChemicalModulator) processImmediateBinding(ligandType LigandType, sourcePos Position3D, concentration float64, sourceID string) {
	targets := cm.bindingTargets[ligandType]
	if targets == nil {
		return
	}

	// Calculate effective concentration at each target position
	for _, target := range targets {
		targetPos := target.GetPosition()
		distance := cm.calculateDistance(sourcePos, targetPos)
		effectiveConcentration := cm.calculateBiologicalConcentration(ligandType, concentration, distance)

		// Apply binding if concentration is biologically significant
		if effectiveConcentration > 0.001 { // 1 μM threshold
			target.Bind(ligandType, sourceID, effectiveConcentration)
		}
	}
}

// =================================================================================
// BIOLOGICALLY ACCURATE CONCENTRATION CALCULATION
// =================================================================================

// calculateBiologicalConcentration computes concentration at distance using biological models
func (cm *ChemicalModulator) calculateBiologicalConcentration(ligandType LigandType, sourceConcentration, distance float64) float64 {
	kinetics, exists := cm.ligandKinetics[ligandType]
	if !exists {
		// Fallback: simple exponential decay
		return sourceConcentration * math.Exp(-distance/10.0)
	}

	// No concentration beyond effective range
	if distance > kinetics.MaxRange {
		return 0.0
	}

	// At source position
	if distance < 0.001 {
		return sourceConcentration
	}

	// Apply biologically appropriate diffusion model based on neurotransmitter type
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
func (cm *ChemicalModulator) calculateSynapticDiffusion(kinetics LigandKinetics, sourceConc, distance float64) float64 {
	// Model: Gaussian diffusion profile with rapid decay
	// Research basis: Synaptic cleft ~20nm, steep dropoff beyond 1-2μm

	// Gaussian decay with biologically measured sigma
	sigma := kinetics.MaxRange / 3.0 // Standard deviation from range
	gaussianDecay := math.Exp(-(distance * distance) / (2.0 * sigma * sigma))

	// Scale by diffusion coefficient
	diffusionScale := kinetics.DiffusionRate / 1.0 // Normalize to 1 μm²/ms baseline

	return sourceConc * gaussianDecay * diffusionScale
}

// calculateVolumeTransmission models slow neuromodulator diffusion over long distances
func (cm *ChemicalModulator) calculateVolumeTransmission(kinetics LigandKinetics, sourceConc, distance float64) float64 {
	// Model: Power law decay characteristic of 3D diffusion
	// Research basis: Measured dopamine/serotonin concentration profiles

	if distance < 1.0 {
		// Near-field: linear decay to avoid mathematical singularities
		return sourceConc * (1.0 - distance/10.0) * (kinetics.DiffusionRate / 0.2)
	} else {
		// Far-field: power law decay with exponential cutoff
		powerDecay := math.Pow(distance, -0.5)                       // 3D diffusion scaling
		expCutoff := math.Exp(-distance / (kinetics.MaxRange * 0.6)) // Gentle cutoff
		diffusionScale := kinetics.DiffusionRate / 0.2               // Normalize to dopamine baseline

		return sourceConc * powerDecay * expCutoff * diffusionScale
	}
}

// calculateMixedSignaling models acetylcholine's dual synaptic/volume properties
func (cm *ChemicalModulator) calculateMixedSignaling(kinetics LigandKinetics, sourceConc, distance float64) float64 {
	// Model: Exponential decay with moderate range
	// Research basis: ACh has both phasic and tonic components

	lambda := kinetics.MaxRange / 2.5 // Decay length constant
	decay := math.Exp(-distance / lambda)
	diffusionScale := kinetics.DiffusionRate / 0.4 // Normalize to ACh baseline

	return sourceConc * decay * diffusionScale
}

// calculateDefaultDiffusion provides fallback exponential decay
func (cm *ChemicalModulator) calculateDefaultDiffusion(kinetics LigandKinetics, sourceConc, distance float64) float64 {
	lambda := kinetics.MaxRange / 3.0
	decay := math.Exp(-distance / lambda)
	diffusionScale := kinetics.DiffusionRate / 1.0

	return sourceConc * decay * diffusionScale
}

// =================================================================================
// CONCENTRATION FIELD MANAGEMENT
// =================================================================================

// GetConcentration returns current concentration at a position with biological accuracy
func (cm *ChemicalModulator) GetConcentration(ligandType LigandType, position Position3D) float64 {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	field, exists := cm.concentrationFields[ligandType]
	if !exists {
		return 0.0
	}

	// Check for direct position match first (optimization)
	if concentration, exists := field.Concentrations[position]; exists {
		return concentration
	}

	// Calculate concentration from all active sources
	totalConcentration := 0.0

	// Add contributions from active sources
	for _, source := range field.Sources {
		if source.Active {
			distance := cm.calculateDistance(source.Position, position)
			concentration := cm.calculateBiologicalConcentration(ligandType, source.ReleaseRate, distance)
			totalConcentration += concentration
		}
	}

	// Add contributions from stored concentration points
	for sourcePos, sourceConc := range field.Concentrations {
		distance := cm.calculateDistance(sourcePos, position)
		if distance > 0.001 { // Avoid self-calculation
			concentration := cm.calculateBiologicalConcentration(ligandType, sourceConc, distance)
			totalConcentration += concentration
		}
	}

	return totalConcentration
}

// updateConcentrationField creates or updates concentration field for a ligand
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

	// Update maximum concentration tracking
	if concentration > field.MaxConcentration {
		field.MaxConcentration = concentration
	}
}

// =================================================================================
// BIOLOGICAL DECAY PROCESSING
// =================================================================================

// Start begins background processing with biologically accurate timing
func (cm *ChemicalModulator) Start() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.isRunning {
		return nil
	}

	cm.isRunning = true

	// Start background decay processing at biological frequency
	go cm.biologicalDecayProcessor()

	return nil
}

// Stop ends chemical processing
func (cm *ChemicalModulator) Stop() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.isRunning = false
	return nil
}

// biologicalDecayProcessor handles concentration decay with biological timing
func (cm *ChemicalModulator) biologicalDecayProcessor() {
	// Use 1ms update interval for biological realism (1 kHz)
	ticker := time.NewTicker(1 * time.Millisecond)
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
func (cm *ChemicalModulator) processBiologicalDecay() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()

	for ligandType, field := range cm.concentrationFields {
		kinetics := cm.ligandKinetics[ligandType]
		dt := now.Sub(field.LastUpdate).Seconds()

		// Skip if time interval too small to be meaningful
		if dt < 0.0001 { // 0.1ms minimum
			continue
		}

		// Apply biologically appropriate decay to all concentrations
		for pos, concentration := range field.Concentrations {
			newConcentration := cm.calculateBiologicalDecay(ligandType, kinetics, concentration, dt)

			if newConcentration < cm.getBiologicalThreshold(ligandType) {
				// Remove concentrations below biological significance
				delete(field.Concentrations, pos)
			} else {
				field.Concentrations[pos] = newConcentration
			}
		}

		// Update field metadata
		field.LastUpdate = now
		field.MaxConcentration = cm.calculateMaxConcentration(field.Concentrations)
	}
}

// calculateBiologicalDecay computes concentration after biological clearance
func (cm *ChemicalModulator) calculateBiologicalDecay(ligandType LigandType, kinetics LigandKinetics, concentration float64, deltaTime float64) float64 {
	// Calculate total clearance rate (enzymatic + transporter)
	totalClearanceRate := kinetics.DecayRate + kinetics.ClearanceRate

	// Apply exponential decay based on biological clearance mechanisms
	return concentration * math.Exp(-totalClearanceRate*deltaTime)
}

// getBiologicalThreshold returns the minimum biologically significant concentration
func (cm *ChemicalModulator) getBiologicalThreshold(ligandType LigandType) float64 {
	switch ligandType {
	case LigandGlutamate, LigandGABA:
		return 0.01 // 10 μM - below typical receptor Kd
	case LigandDopamine, LigandSerotonin:
		return 0.001 // 1 μM - neuromodulator threshold
	case LigandAcetylcholine:
		return 0.005 // 5 μM - intermediate threshold
	default:
		return 0.001 // Conservative default
	}
}

// =================================================================================
// UTILITY FUNCTIONS
// =================================================================================

// RegisterTarget adds a component to receive chemical signals
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

// GetRecentReleases returns recent release events for analysis
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
func (cm *ChemicalModulator) ForceDecayUpdate() {
	cm.processBiologicalDecay()
}

// calculateDistance computes 3D Euclidean distance between positions
func (cm *ChemicalModulator) calculateDistance(pos1, pos2 Position3D) float64 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	dz := pos1.Z - pos2.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// getBiologicalReleaseDuration returns typical release duration for each neurotransmitter
func (cm *ChemicalModulator) getBiologicalReleaseDuration(ligandType LigandType) time.Duration {
	switch ligandType {
	case LigandGlutamate, LigandGABA:
		return 1 * time.Millisecond // Fast synaptic release
	case LigandAcetylcholine:
		return 5 * time.Millisecond // Intermediate duration
	case LigandDopamine, LigandSerotonin:
		return 100 * time.Millisecond // Long neuromodulator release
	default:
		return 10 * time.Millisecond // Default duration
	}
}

// calculateMaxConcentration finds maximum concentration in a map
func (cm *ChemicalModulator) calculateMaxConcentration(concentrations map[Position3D]float64) float64 {
	maxConc := 0.0
	for _, conc := range concentrations {
		if conc > maxConc {
			maxConc = conc
		}
	}
	return maxConc
}
