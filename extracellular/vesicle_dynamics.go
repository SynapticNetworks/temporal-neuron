/*
=================================================================================
VESICLE DYNAMICS - SYNAPTIC NEUROTRANSMITTER RELEASE REGULATION
=================================================================================

Implements biologically accurate vesicle dynamics that control neurotransmitter
release frequency at synaptic terminals. Models the fundamental biological
constraints that limit how fast neurons can release chemical signals.

BIOLOGICAL FOUNDATION:
Real neurons cannot release neurotransmitters indefinitely - they are limited by
the availability and recycling dynamics of synaptic vesicles. This package models
the key biological processes that regulate chemical release rates:

1. VESICLE POOL DYNAMICS:
   - Ready Releasable Pool (RRP): ~5-20 vesicles immediately available for release
   - Recycling Pool: ~100-200 vesicles that can be mobilized within seconds
   - Reserve Pool: ~1000+ vesicles for sustained high-frequency activity
   - Pool depletion occurs with rapid, repeated stimulation

2. VESICLE RECYCLING MECHANISMS:
   - Fast Endocytosis: ~1-2 seconds (clathrin-independent, kiss-and-run)
   - Slow Endocytosis: ~10-30 seconds (clathrin-mediated, full recycling)
   - Vesicle Refilling: ~5-10 seconds (neurotransmitter loading via transporters)
   - Repriming: ~2-5 seconds (molecular machinery reset for next release)

3. CALCIUM-DEPENDENT RELEASE PROBABILITY:
   - Low calcium: ~0.1-0.3 probability per action potential
   - High calcium: ~0.8-1.0 probability (during high-frequency firing)
   - Calcium buffering limits sustained high-probability release
   - Calcium channel inactivation provides natural rate limiting

4. SYNAPTIC FATIGUE AND DEPRESSION:
   - Short-term depression: Vesicle pool depletion (seconds to minutes)
   - Use-dependent reduction in release probability
   - Recovery follows exponential kinetics with multiple time constants
   - Metabolic limitations affect vesicle recycling efficiency

RESEARCH BASIS:
- Alabi & Tsien (2012): "Synaptic vesicle pools and dynamics"
- Rizzoli & Betz (2005): "Synaptic vesicle pools"
- Wu & Borst (1999): "The reduced release probability of releasable vesicles"
- von Gersdorff & Matthews (1997): "Depletion and replenishment of vesicle pools"

KEY PARAMETERS MODELED:
- Vesicle pool sizes (ready, recycling, reserve)
- Recycling time constants (fast/slow endocytosis)
- Release probability modulation
- Activity-dependent fatigue and recovery
- Calcium-dependent release enhancement

This implementation provides realistic constraints on chemical signaling that
prevent unrealistic "machine-gun" neurotransmitter release while enabling
the full range of biological firing patterns and plasticity.
=================================================================================
*/

package extracellular

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// VesicleDynamics models synaptic vesicle availability and recycling kinetics
// Implements the biological reality that neurotransmitter release is limited by
// vesicle pool dynamics, recycling mechanisms, and calcium-dependent processes
type VesicleDynamics struct {
	// === VESICLE POOL TRACKING ===
	releaseEvents   []VesicleReleaseEvent // Recent release history for pool tracking
	maxReleaseRate  float64               // Maximum sustainable release rate (vesicles/second)
	recyclingWindow time.Duration         // Time window for vesicle recycling (biological constraint)

	// === BIOLOGICAL POOL DYNAMICS ===
	readyPoolSize     int // Ready Releasable Pool (RRP) - immediately available vesicles
	recyclingPoolSize int // Recycling Pool - vesicles available within seconds
	reservePoolSize   int // Reserve Pool - vesicles for sustained activity
	currentReadyPool  int // Current number of ready vesicles available

	// === RELEASE PROBABILITY MODULATION ===
	baseLinesReleaseProbability float64 // Baseline probability of vesicle release per stimulus
	calciumEnhancement          float64 // Current calcium-dependent enhancement (0.0-2.0)
	fatigueLevel                float64 // Accumulated fatigue reducing release probability (0.0-1.0)

	// === RECYCLING KINETICS ===
	fastRecyclingRate float64       // Fast endocytosis rate (vesicles/second)
	slowRecyclingRate float64       // Slow endocytosis rate (vesicles/second)
	refillTime        time.Duration // Time to refill vesicles with neurotransmitter
	reprimingTime     time.Duration // Time to reprime release machinery

	// === THREAD SAFETY ===
	mu sync.RWMutex // Protects all vesicle dynamics state from concurrent access

	// === RANDOM NUMBER GENERATION ===
	rng *rand.Rand // Private random number generator for deterministic behavior
}

// VesicleReleaseEvent records individual vesicle release events for tracking
// Models the biological reality that each vesicle release is a discrete,
// trackable event with specific timing and consequences for future releases
type VesicleReleaseEvent struct {
	Timestamp          time.Time        `json:"timestamp"`           // When vesicle was released
	ReleaseProbability float64          `json:"release_probability"` // Effective release probability at time of release
	PoolState          VesiclePoolState `json:"pool_state"`          // Vesicle pool state at time of release
	CalciumLevel       float64          `json:"calcium_level"`       // Calcium concentration during release
	RecoveryTime       time.Duration    `json:"recovery_time"`       // Expected time for this vesicle to be available again
}

// VesiclePoolState captures the state of all vesicle pools at a given moment
// Provides detailed information about vesicle availability for biological realism
type VesiclePoolState struct {
	ReadyVesicles     int     `json:"ready_vesicles"`     // Immediately available for release
	RecyclingVesicles int     `json:"recycling_vesicles"` // In recycling process
	ReserveVesicles   int     `json:"reserve_vesicles"`   // Long-term reserve
	TotalVesicles     int     `json:"total_vesicles"`     // Total vesicle count
	DepletionLevel    float64 `json:"depletion_level"`    // Pool depletion (0.0-1.0)
	FatigueLevel      float64 `json:"fatigue_level"`      // Synaptic fatigue (0.0-1.0)
}

// Biological constants based on experimental measurements
const (
	// VESICLE POOL SIZES (experimentally measured ranges)
	DEFAULT_READY_POOL_SIZE     = 15   // RRP: 5-20 vesicles typical
	DEFAULT_RECYCLING_POOL_SIZE = 150  // Recycling pool: 100-200 vesicles
	DEFAULT_RESERVE_POOL_SIZE   = 1000 // Reserve pool: 1000+ vesicles

	// RECYCLING TIME CONSTANTS (from patch-clamp studies)
	FAST_RECYCLING_TIME = 2 * time.Second  // Kiss-and-run recycling
	SLOW_RECYCLING_TIME = 20 * time.Second // Full clathrin-mediated recycling
	VESICLE_REFILL_TIME = 8 * time.Second  // Neurotransmitter loading time
	REPRIMING_TIME      = 3 * time.Second  // Release machinery reset time

	// RELEASE PROBABILITY PARAMETERS
	BASELINE_RELEASE_PROBABILITY = 0.25             // Typical release probability per AP
	MAX_CALCIUM_ENHANCEMENT      = 3.0              // Maximum calcium-dependent enhancement
	FATIGUE_RECOVERY_TIME        = 30 * time.Second // Recovery from synaptic depression

	// ACTIVITY-DEPENDENT RATES
	DEFAULT_MAX_RELEASE_RATE  = 50.0 // Conservative maximum sustainable rate (Hz)
	HIGH_FREQUENCY_THRESHOLD  = 20.0 // Frequency above which fatigue becomes significant (Hz)
	FATIGUE_ACCUMULATION_RATE = 0.1  // Rate at which fatigue accumulates during high activity
)

// NewVesicleDynamics creates a biologically realistic vesicle dynamics system
// Initializes vesicle pools, recycling parameters, and release probability based on
// experimental measurements from real synapses
//
// Parameters:
//
//	maxReleasesPerSecond: Maximum sustainable release rate for this synapse type
//	                     Range: 1-100 Hz depending on synapse type
//	                     - Fast GABAergic: 50-100 Hz
//	                     - Glutamatergic: 20-50 Hz
//	                     - Neuromodulatory: 1-10 Hz
func NewVesicleDynamics(maxReleasesPerSecond float64) *VesicleDynamics {
	// Validate biological constraints
	if maxReleasesPerSecond <= 0 {
		maxReleasesPerSecond = DEFAULT_MAX_RELEASE_RATE
	}
	if maxReleasesPerSecond > 200 { // Biological upper limit
		maxReleasesPerSecond = 200
	}

	return &VesicleDynamics{
		// Initialize release tracking
		releaseEvents:   make([]VesicleReleaseEvent, 0),
		maxReleaseRate:  maxReleasesPerSecond,
		recyclingWindow: FAST_RECYCLING_TIME,

		// Initialize vesicle pools at full capacity
		readyPoolSize:     DEFAULT_READY_POOL_SIZE,
		recyclingPoolSize: DEFAULT_RECYCLING_POOL_SIZE,
		reservePoolSize:   DEFAULT_RESERVE_POOL_SIZE,
		currentReadyPool:  DEFAULT_READY_POOL_SIZE, // Start with full ready pool

		// Initialize release probability
		baseLinesReleaseProbability: BASELINE_RELEASE_PROBABILITY,
		calciumEnhancement:          1.0, // No enhancement initially
		fatigueLevel:                0.0, // No fatigue initially

		// Initialize recycling kinetics
		fastRecyclingRate: float64(DEFAULT_READY_POOL_SIZE) / FAST_RECYCLING_TIME.Seconds(),
		slowRecyclingRate: float64(DEFAULT_RECYCLING_POOL_SIZE) / SLOW_RECYCLING_TIME.Seconds(),
		refillTime:        VESICLE_REFILL_TIME,
		reprimingTime:     REPRIMING_TIME,

		// Initialize random number generator with current time for uniqueness
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// HasAvailableVesicles checks if vesicles are available for release
// Models the biological process of vesicle availability checking that occurs
// during synaptic transmission. Considers vesicle pool state, recycling status,
// fatigue level, and calcium-dependent release probability.
//
// BIOLOGICAL PROCESS MODELED:
// 1. Check ready releasable pool for immediately available vesicles
// 2. Calculate effective release probability based on calcium and fatigue
// 3. Consider recent release history and recycling constraints
// 4. Update pools based on ongoing recycling processes
// 5. Record release event if vesicle release occurs
//
// Returns:
//
//	bool: true if vesicle release can occur, false if pools depleted or fatigued
func (vd *VesicleDynamics) HasAvailableVesicles() bool {
	vd.mu.Lock()
	defer vd.mu.Unlock()

	now := time.Now()

	// Update vesicle pools based on recycling processes
	vd.updateVesiclePools(now)

	// Check if ready pool has available vesicles
	if vd.currentReadyPool <= 0 {
		return false // No vesicles immediately available
	}

	// Calculate effective release probability considering all factors
	effectiveReleaseProbability := vd.calculateEffectiveReleaseProbability()

	// Check recent release rate to prevent unrealistic firing
	recentReleaseRate := vd.calculateRecentReleaseRate(now)
	if recentReleaseRate >= vd.maxReleaseRate {
		return false // Rate limiting - too many recent releases
	}

	// Probabilistic release based on biological factors
	// In real synapses, not every action potential causes vesicle release
	releaseOccurs := vd.shouldReleaseVesicle(effectiveReleaseProbability)

	if releaseOccurs {
		// Record the release event and update pools
		vd.executeVesicleRelease(now, effectiveReleaseProbability)
		return true
	}

	return false // No release occurred due to probabilistic factors
}

// GetCurrentReleaseRate returns the current vesicle release rate
// Provides information about recent synaptic activity for monitoring and analysis
//
// Returns:
//
//	float64: Current release rate in vesicles per second over the recycling window
func (vd *VesicleDynamics) GetCurrentReleaseRate() float64 {
	vd.mu.RLock()
	defer vd.mu.RUnlock()
	return vd.calculateRecentReleaseRate(time.Now())
}

// GetVesiclePoolState returns detailed information about current vesicle pool status
// Provides comprehensive information about vesicle availability, depletion, and fatigue
// for monitoring synaptic health and performance
//
// Returns:
//
//	VesiclePoolState: Complete state of all vesicle pools and dynamics
func (vd *VesicleDynamics) GetVesiclePoolState() VesiclePoolState {
	vd.mu.Lock()
	defer vd.mu.Unlock()

	// CRITICAL: Always update pools before returning state to ensure accuracy
	vd.updateVesiclePools(time.Now())

	totalVesicles := vd.readyPoolSize + vd.recyclingPoolSize + vd.reservePoolSize
	depletionLevel := 1.0 - (float64(vd.currentReadyPool) / float64(vd.readyPoolSize))

	return VesiclePoolState{
		ReadyVesicles:     vd.currentReadyPool,
		RecyclingVesicles: vd.recyclingPoolSize - (vd.readyPoolSize - vd.currentReadyPool),
		ReserveVesicles:   vd.reservePoolSize,
		TotalVesicles:     totalVesicles,
		DepletionLevel:    depletionLevel,
		FatigueLevel:      vd.fatigueLevel,
	}
}

// SetCalciumLevel updates calcium-dependent release enhancement
// Models how intracellular calcium concentration affects vesicle release probability
// High calcium (during high-frequency firing) increases release probability
//
// Parameters:
//
//	calciumLevel: Relative calcium concentration (0.0-2.0, where 1.0 = baseline)
func (vd *VesicleDynamics) SetCalciumLevel(calciumLevel float64) {
	vd.mu.Lock()
	defer vd.mu.Unlock()

	// Biological calcium enhancement curve (sigmoidal)
	// Based on measured calcium-release probability relationships
	if calciumLevel < 0.0 {
		calciumLevel = 0.0
	}
	if calciumLevel > 2.0 {
		calciumLevel = 2.0
	}

	// Sigmoidal enhancement: low calcium reduces probability, high calcium enhances
	vd.calciumEnhancement = 0.5 + 1.5*(calciumLevel/(calciumLevel+0.5))
}

// ResetVesiclePools restores vesicle pools to full capacity
// Models the biological recovery process after periods of intense activity
// Useful for testing or simulating recovery after synaptic depression
func (vd *VesicleDynamics) ResetVesiclePools() {
	vd.mu.Lock()
	defer vd.mu.Unlock()

	vd.currentReadyPool = vd.readyPoolSize
	vd.fatigueLevel = 0.0
	vd.releaseEvents = make([]VesicleReleaseEvent, 0)
}

// GetDebugInfo returns internal state information for testing and debugging
// This method should only be used for testing - not part of the public API
func (vd *VesicleDynamics) GetDebugInfo() map[string]interface{} {
	vd.mu.RLock()
	defer vd.mu.RUnlock()

	now := time.Now()

	// Count events that should have recycled by now
	recycleableEvents := 0
	for _, event := range vd.releaseEvents {
		recycleTime := event.Timestamp.Add(event.RecoveryTime)
		if now.After(recycleTime) {
			recycleableEvents++
		}
	}

	return map[string]interface{}{
		"total_events":       len(vd.releaseEvents),
		"recycleable_events": recycleableEvents,
		"current_ready_pool": vd.currentReadyPool,
		"max_ready_pool":     vd.readyPoolSize,
	}
}

// =================================================================================
// INTERNAL BIOLOGICAL PROCESS IMPLEMENTATIONS
// All methods below assume the caller already holds the appropriate mutex lock
// =================================================================================

// updateVesiclePools processes ongoing vesicle recycling and pool replenishment
// Models the continuous biological processes that restore vesicle availability
// NOTE: This method assumes the caller already holds the mutex lock
func (vd *VesicleDynamics) updateVesiclePools(now time.Time) {
	// ALWAYS process vesicle recycling when pools are updated
	vd.processVesicleRecycling(now)

	// Update fatigue recovery
	vd.updateFatigueRecovery(now)

	// Clean up old release events outside recycling window (but keep reasonable history)
	vd.cleanupOldReleaseEvents(now)
}

// calculateEffectiveReleaseProbability computes current release probability
// Integrates multiple biological factors affecting vesicle release
func (vd *VesicleDynamics) calculateEffectiveReleaseProbability() float64 {
	// Start with baseline probability
	probability := vd.baseLinesReleaseProbability

	// Apply calcium enhancement
	probability *= vd.calciumEnhancement

	// Apply fatigue reduction
	probability *= (1.0 - vd.fatigueLevel)

	// Apply pool depletion effects
	poolDepletionFactor := float64(vd.currentReadyPool) / float64(vd.readyPoolSize)
	probability *= poolDepletionFactor

	// Ensure probability stays in valid range
	if probability < 0.0 {
		probability = 0.0
	}
	if probability > 1.0 {
		probability = 1.0
	}

	return probability
}

// calculateRecentReleaseRate determines current release frequency
func (vd *VesicleDynamics) calculateRecentReleaseRate(now time.Time) float64 {
	cutoff := now.Add(-vd.recyclingWindow)
	count := 0

	for _, event := range vd.releaseEvents {
		if event.Timestamp.After(cutoff) {
			count++
		}
	}

	return float64(count) / vd.recyclingWindow.Seconds()
}

// shouldReleaseVesicle determines probabilistic vesicle release
// Models the stochastic nature of biological vesicle release
func (vd *VesicleDynamics) shouldReleaseVesicle(probability float64) bool {
	// Use private RNG for thread safety and deterministic behavior
	return vd.rng.Float64() < probability
}

// executeVesicleRelease performs the vesicle release and updates all relevant state
func (vd *VesicleDynamics) executeVesicleRelease(now time.Time, probability float64) {
	// Consume a vesicle from ready pool
	vd.currentReadyPool--

	// Calculate recovery time for this vesicle
	recoveryTime := vd.calculateVesicleRecoveryTime()

	// Record the release event
	releaseEvent := VesicleReleaseEvent{
		Timestamp:          now,
		ReleaseProbability: probability,
		PoolState: VesiclePoolState{
			ReadyVesicles: vd.currentReadyPool,
			FatigueLevel:  vd.fatigueLevel,
		},
		CalciumLevel: vd.calciumEnhancement,
		RecoveryTime: recoveryTime,
	}

	vd.releaseEvents = append(vd.releaseEvents, releaseEvent)

	// Update fatigue based on release frequency
	vd.updateFatigueLevel()
}

// processVesicleRecycling handles the return of vesicles to available pools
func (vd *VesicleDynamics) processVesicleRecycling(now time.Time) {
	// Process recycling and track which events to remove
	eventsToRemove := make([]int, 0)

	for i, event := range vd.releaseEvents {
		recycleTime := event.Timestamp.Add(event.RecoveryTime)
		if now.After(recycleTime) {
			// Recycle this vesicle if there's room in the ready pool
			if vd.currentReadyPool < vd.readyPoolSize {
				vd.currentReadyPool++
			}
			// Mark this event for removal (vesicle has been recycled)
			eventsToRemove = append(eventsToRemove, i)
		}
	}

	// Remove recycled events (iterate backwards to maintain indices)
	for i := len(eventsToRemove) - 1; i >= 0; i-- {
		idx := eventsToRemove[i]
		// Remove element at idx
		vd.releaseEvents = append(vd.releaseEvents[:idx], vd.releaseEvents[idx+1:]...)
	}
}

// updateFatigueRecovery models the biological recovery from synaptic fatigue
func (vd *VesicleDynamics) updateFatigueRecovery(now time.Time) {
	// Exponential recovery from fatigue
	if vd.fatigueLevel > 0 {
		recoveryRate := 1.0 / FATIGUE_RECOVERY_TIME.Seconds()
		deltaTime := vd.recyclingWindow.Seconds() // Use recycling window as time step
		vd.fatigueLevel *= math.Exp(-recoveryRate * deltaTime)

		if vd.fatigueLevel < 0.01 {
			vd.fatigueLevel = 0.0 // Complete recovery
		}
	}
}

// updateFatigueLevel increases fatigue based on recent release activity
func (vd *VesicleDynamics) updateFatigueLevel() {
	currentRate := vd.calculateRecentReleaseRate(time.Now())
	if currentRate > HIGH_FREQUENCY_THRESHOLD {
		fatigueIncrease := FATIGUE_ACCUMULATION_RATE * (currentRate / vd.maxReleaseRate)
		vd.fatigueLevel += fatigueIncrease

		if vd.fatigueLevel > 1.0 {
			vd.fatigueLevel = 1.0 // Maximum fatigue
		}
	}
}

// calculateVesicleRecoveryTime determines how long until a released vesicle is available again
func (vd *VesicleDynamics) calculateVesicleRecoveryTime() time.Duration {
	// Model both fast and slow recycling pathways
	// Most vesicles (70%) use fast recycling, some (30%) use slow recycling
	if vd.rng.Float64() < 0.7 {
		return FAST_RECYCLING_TIME + time.Duration(vd.rng.Float64()*float64(vd.reprimingTime))
	} else {
		return SLOW_RECYCLING_TIME + vd.refillTime
	}
}

// cleanupOldReleaseEvents removes events outside the recycling window
func (vd *VesicleDynamics) cleanupOldReleaseEvents(now time.Time) {
	cutoff := now.Add(-vd.recyclingWindow * 3) // Keep some history beyond recycling window
	validEvents := make([]VesicleReleaseEvent, 0)

	for _, event := range vd.releaseEvents {
		if event.Timestamp.After(cutoff) {
			validEvents = append(validEvents, event)
		}
	}

	vd.releaseEvents = validEvents
}
