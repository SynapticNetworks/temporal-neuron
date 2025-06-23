package neuron

import (
	"math"
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/message"
)

/*
=================================================================================
SYNAPTIC SCALING - HOMEOSTATIC RECEPTOR SENSITIVITY REGULATION
=================================================================================

BIOLOGICAL OVERVIEW:
Synaptic scaling is a homeostatic mechanism where post-synaptic neurons adjust
their receptor sensitivity to maintain optimal input strength. Unlike STDP which
changes individual synapse weights, synaptic scaling proportionally adjusts ALL
synaptic gains to maintain network stability while preserving learned patterns.

BIOLOGICAL MECHANISMS MODELED:
1. **Activity Sensing**: Post-synaptic neurons monitor total synaptic drive
2. **Receptor Trafficking**: AMPA/NMDA receptor insertion/removal at synapses
3. **Calcium-Dependent Gating**: Scaling only occurs during sufficient activity
4. **Proportional Adjustment**: All gains scaled by same factor (pattern preservation)
5. **Slow Timescales**: Minutes to hours (much slower than STDP milliseconds)

ARCHITECTURAL INTEGRATION:
- Operates independently of pre-synaptic neurons (post-synaptic control)
- Uses callback architecture for matrix coordination
- Integrates with activity tracking from message processing
- Coordinated with homeostatic plasticity but on different timescales

=================================================================================
*/

// ============================================================================
// SYNAPTIC SCALING CONFIGURATION AND STATE
// ============================================================================

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

	// === CORE SCALING PARAMETERS ===
	TargetInputStrength float64       // Desired average effective input strength
	ScalingRate         float64       // Rate of receptor gain adjustment per scaling event
	ScalingInterval     time.Duration // Time between synaptic scaling operations

	// === SAFETY CONSTRAINTS ===
	MinScalingFactor float64 // Minimum multiplier applied to gains per scaling event
	MaxScalingFactor float64 // Maximum multiplier applied to gains per scaling event
	MinReceptorGain  float64 // Absolute minimum receptor sensitivity
	MaxReceptorGain  float64 // Absolute maximum receptor sensitivity

	// === ACTIVITY GATING ===
	MinActivityForScaling  float64       // Minimum activity level required to trigger scaling
	ActivitySamplingWindow time.Duration // Time window for measuring input activity

	// === STATE TRACKING ===
	LastScalingUpdate time.Time // Timestamp of most recent scaling operation
	ScalingHistory    []float64 // Recent scaling factors for monitoring
	MaxHistorySize    int       // Maximum number of scaling events to remember
}

// SynapticScalingState holds the current state of the scaling system
type SynapticScalingState struct {
	mu sync.RWMutex

	// === RECEPTOR GAINS ===
	InputGains     map[string]float64 // Receptor sensitivity per input source
	inputGainMutex sync.RWMutex       // Thread-safe access to input gains

	// === ACTIVITY TRACKING ===
	InputActivityHistory   map[string][]InputActivity // Recent input activities per source
	inputActivityMutex     sync.RWMutex               // Thread-safe access to activity history
	ActivityTrackingWindow time.Duration              // Time window for activity integration
	LastActivityCleanup    time.Time                  // Timestamp of last activity cleanup

	// === CONFIGURATION ===
	Config SynapticScalingConfig // Current scaling configuration
}

// InputActivity represents a single synaptic input event for scaling calculations
type InputActivity struct {
	EffectiveValue float64   // Final signal strength (signal × post-gain)
	Timestamp      time.Time // When the input was received
}

// SynapticScalingResult contains the outcome of a scaling operation
type SynapticScalingResult struct {
	ScalingPerformed     bool      // Whether scaling was actually performed
	ScalingFactor        float64   // Factor applied to all gains
	SourcesScaled        []string  // Which input sources were scaled
	AverageInputStrength float64   // Current average input strength
	Reason               string    // Why scaling was or wasn't performed
	Timestamp            time.Time // When the scaling occurred
}

// ============================================================================
// SYNAPTIC SCALING SYSTEM CREATION AND CONFIGURATION
// ============================================================================

// NewSynapticScalingState creates a new synaptic scaling system with default configuration
func NewSynapticScalingState() *SynapticScalingState {
	return &SynapticScalingState{
		InputGains:             make(map[string]float64),
		InputActivityHistory:   make(map[string][]InputActivity),
		ActivityTrackingWindow: SYNAPTIC_SCALING_ACTIVITY_WINDOW_DEFAULT,
		LastActivityCleanup:    time.Now(),
		Config: SynapticScalingConfig{
			Enabled:                false, // Disabled by default
			TargetInputStrength:    SYNAPTIC_SCALING_TARGET_STRENGTH_DEFAULT,
			ScalingRate:            SYNAPTIC_SCALING_RATE_DEFAULT,
			ScalingInterval:        SYNAPTIC_SCALING_INTERVAL_DEFAULT,
			MinScalingFactor:       SYNAPTIC_SCALING_MIN_FACTOR,
			MaxScalingFactor:       SYNAPTIC_SCALING_MAX_FACTOR,
			MinReceptorGain:        SYNAPTIC_SCALING_MIN_GAIN,
			MaxReceptorGain:        SYNAPTIC_SCALING_MAX_GAIN,
			MinActivityForScaling:  SYNAPTIC_SCALING_MIN_ACTIVITY,
			ActivitySamplingWindow: SYNAPTIC_SCALING_ACTIVITY_WINDOW_DEFAULT,
			LastScalingUpdate:      time.Time{},
			ScalingHistory:         make([]float64, 0, SYNAPTIC_SCALING_HISTORY_SIZE),
			MaxHistorySize:         SYNAPTIC_SCALING_HISTORY_SIZE,
		},
	}
}

// EnableScaling activates synaptic scaling with specified parameters
func (s *SynapticScalingState) EnableScaling(targetStrength, scalingRate float64, interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Config.Enabled = true
	s.Config.TargetInputStrength = targetStrength
	s.Config.ScalingRate = scalingRate
	s.Config.ScalingInterval = interval
	s.Config.LastScalingUpdate = time.Now()
}

// DisableScaling turns off synaptic scaling (preserves existing gains)
func (s *SynapticScalingState) DisableScaling() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Config.Enabled = false
}

// SetScalingParameters updates scaling configuration
func (s *SynapticScalingState) SetScalingParameters(config SynapticScalingConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Config = config
	if s.Config.Enabled && s.Config.LastScalingUpdate.IsZero() {
		s.Config.LastScalingUpdate = time.Now()
	}
}

// ============================================================================
// INPUT GAIN MANAGEMENT
// ============================================================================

// ApplyPostSynapticGain applies receptor sensitivity scaling to incoming signals
// This is the core of biologically accurate synaptic scaling - the post-synaptic
// neuron controls its own sensitivity to different input sources
//
// BIOLOGICAL PROCESS MODELED:
// In real neurons, synaptic scaling occurs through changes in post-synaptic
// receptor density (AMPA, NMDA receptors). The pre-synaptic neuron releases
// the same amount of neurotransmitter, but the post-synaptic response changes
// based on receptor availability.
func (s *SynapticScalingState) ApplyPostSynapticGain(msg message.NeuralSignal) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// If scaling is disabled or no source ID, use original signal
	if !s.Config.Enabled || msg.SourceID == "" {
		return msg.Value
	}

	// Get the receptor gain for this input source
	s.inputGainMutex.RLock()
	gain, exists := s.InputGains[msg.SourceID]
	s.inputGainMutex.RUnlock()

	// If source not yet registered, register it with default gain
	if !exists {
		gain = 1.0 // Default receptor sensitivity
		s.registerInputSource(msg.SourceID)
	}

	// Apply receptor gain to the signal
	// Final signal = synaptic_strength × post-synaptic_receptor_sensitivity
	return msg.Value * gain
}

// SetInputGain manually sets the receptor gain for a specific input source
func (s *SynapticScalingState) SetInputGain(sourceID string, gain float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Clamp gain to biological bounds
	if gain < s.Config.MinReceptorGain {
		gain = s.Config.MinReceptorGain
	}
	if gain > s.Config.MaxReceptorGain {
		gain = s.Config.MaxReceptorGain
	}

	s.inputGainMutex.Lock()
	defer s.inputGainMutex.Unlock()

	if s.InputGains == nil {
		s.InputGains = make(map[string]float64)
	}
	s.InputGains[sourceID] = gain
}

// GetInputGains returns a copy of current input gains for monitoring
func (s *SynapticScalingState) GetInputGains() map[string]float64 {
	s.inputGainMutex.RLock()
	defer s.inputGainMutex.RUnlock()

	gains := make(map[string]float64, len(s.InputGains))
	for sourceID, gain := range s.InputGains {
		gains[sourceID] = gain
	}
	return gains
}

// registerInputSource registers a new input source for synaptic scaling
func (s *SynapticScalingState) registerInputSource(sourceID string) {
	if !s.Config.Enabled {
		return
	}

	s.inputGainMutex.Lock()
	defer s.inputGainMutex.Unlock()

	// Check if already registered (double-check inside lock)
	if _, exists := s.InputGains[sourceID]; !exists {
		if s.InputGains == nil {
			s.InputGains = make(map[string]float64)
		}
		s.InputGains[sourceID] = 1.0 // Default receptor sensitivity
	}
}

// ============================================================================
// ACTIVITY TRACKING FOR SCALING DECISIONS
// ============================================================================

// RecordInputActivity tracks effective input signal strength for scaling decisions
// This models how post-synaptic neurons monitor their actual synaptic input patterns
func (s *SynapticScalingState) RecordInputActivity(sourceID string, effectiveSignalValue float64) {
	if !s.Config.Enabled || sourceID == "" {
		return
	}

	now := time.Now()

	// Create activity record
	activity := InputActivity{
		EffectiveValue: effectiveSignalValue,
		Timestamp:      now,
	}

	// Store activity
	s.inputActivityMutex.Lock()
	if s.InputActivityHistory == nil {
		s.InputActivityHistory = make(map[string][]InputActivity)
	}
	s.InputActivityHistory[sourceID] = append(s.InputActivityHistory[sourceID], activity)
	s.inputActivityMutex.Unlock()

	// Periodic cleanup
	if now.Sub(s.LastActivityCleanup) > s.ActivityTrackingWindow {
		s.cleanOldActivityHistory(now)
		s.LastActivityCleanup = now
	}
}

// cleanOldActivityHistory removes activity data outside the integration window
func (s *SynapticScalingState) cleanOldActivityHistory(currentTime time.Time) {
	s.inputActivityMutex.Lock()
	defer s.inputActivityMutex.Unlock()

	cutoff := currentTime.Add(-s.Config.ActivitySamplingWindow)
	for sourceID, activities := range s.InputActivityHistory {
		var validActivities []InputActivity
		for _, activity := range activities {
			if activity.Timestamp.After(cutoff) {
				validActivities = append(validActivities, activity)
			}
		}
		s.InputActivityHistory[sourceID] = validActivities
	}
}

// GetRecentActivityStrength calculates average recent activity for a source
func (s *SynapticScalingState) GetRecentActivityStrength(sourceID string) float64 {
	s.inputActivityMutex.RLock()
	defer s.inputActivityMutex.RUnlock()

	activities, exists := s.InputActivityHistory[sourceID]
	if !exists || len(activities) == 0 {
		return 0.0
	}

	// Calculate average recent activity
	cutoff := time.Now().Add(-s.Config.ActivitySamplingWindow)
	var sum float64
	var count int

	for _, activity := range activities {
		if activity.Timestamp.After(cutoff) {
			sum += math.Abs(activity.EffectiveValue)
			count++
		}
	}

	if count == 0 {
		return 0.0
	}
	return sum / float64(count)
}

// ============================================================================
// CORE SYNAPTIC SCALING ALGORITHM
// ============================================================================

// PerformScaling executes the main synaptic scaling algorithm
// This is the core biological scaling process that maintains input balance
func (s *SynapticScalingState) PerformScaling(calciumLevel, recentFiringRate float64) SynapticScalingResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := SynapticScalingResult{
		ScalingPerformed: false,
		ScalingFactor:    1.0,
		Timestamp:        time.Now(),
		Reason:           "scaling_disabled",
	}

	// Early exit if scaling is disabled
	if !s.Config.Enabled {
		return result
	}

	// Check if it's time to perform scaling
	if result.Timestamp.Sub(s.Config.LastScalingUpdate) < s.Config.ScalingInterval {
		result.Reason = "interval_not_reached"
		return result
	}

	// Biological activity gating - require minimum activity
	if calciumLevel < s.Config.MinActivityForScaling || recentFiringRate < 0.1 {
		result.Reason = "insufficient_activity"
		s.Config.LastScalingUpdate = result.Timestamp
		return result
	}

	// Calculate current average effective input strength
	averageStrength, activeSourceCount := s.calculateAverageInputStrength()
	result.AverageInputStrength = averageStrength

	if activeSourceCount == 0 {
		result.Reason = "no_active_sources"
		return result
	}

	// Check if scaling is needed (biological significance test)
	strengthDifference := s.Config.TargetInputStrength - averageStrength
	relativeError := math.Abs(strengthDifference) / s.Config.TargetInputStrength

	if relativeError < SYNAPTIC_SCALING_SIGNIFICANCE_THRESHOLD {
		result.Reason = "within_target_range"
		s.Config.LastScalingUpdate = result.Timestamp
		return result
	}

	// Calculate scaling factor
	rawScalingFactor := 1.0 + (strengthDifference * s.Config.ScalingRate)
	scalingFactor := math.Max(s.Config.MinScalingFactor,
		math.Min(s.Config.MaxScalingFactor, rawScalingFactor))

	// Skip if change is too small
	if math.Abs(scalingFactor-1.0) < SYNAPTIC_SCALING_MIN_CHANGE {
		result.Reason = "change_too_small"
		s.Config.LastScalingUpdate = result.Timestamp
		return result
	}

	// Apply scaling to all active sources
	scaledSources := s.applyScalingFactor(scalingFactor)

	// Update result
	result.ScalingPerformed = true
	result.ScalingFactor = scalingFactor
	result.SourcesScaled = scaledSources
	result.Reason = "scaling_applied"

	// Update state
	s.Config.LastScalingUpdate = result.Timestamp
	s.addToScalingHistory(scalingFactor)

	return result
}

// calculateAverageInputStrength computes current average effective input strength
func (s *SynapticScalingState) calculateAverageInputStrength() (float64, int) {
	s.inputGainMutex.RLock()
	s.inputActivityMutex.RLock()
	defer s.inputGainMutex.RUnlock()
	defer s.inputActivityMutex.RUnlock()

	if len(s.InputGains) == 0 {
		return 0.0, 0
	}

	var totalEffectiveStrength float64
	var activeSourceCount int

	for sourceID := range s.InputGains {
		activities, hasActivity := s.InputActivityHistory[sourceID]
		if !hasActivity || len(activities) == 0 {
			continue
		}

		// Calculate average recent activity for this source
		cutoff := time.Now().Add(-s.Config.ActivitySamplingWindow)
		var activitySum float64
		var activityCount int

		for _, activity := range activities {
			if activity.Timestamp.After(cutoff) {
				activitySum += math.Abs(activity.EffectiveValue)
				activityCount++
			}
		}

		if activityCount > 0 {
			averageActivity := activitySum / float64(activityCount)
			totalEffectiveStrength += averageActivity
			activeSourceCount++
		}
	}

	if activeSourceCount == 0 {
		return 0.0, 0
	}

	return totalEffectiveStrength / float64(activeSourceCount), activeSourceCount
}

// applyScalingFactor applies the calculated scaling factor to all active input gains
func (s *SynapticScalingState) applyScalingFactor(scalingFactor float64) []string {
	s.inputGainMutex.Lock()
	defer s.inputGainMutex.Unlock()

	var scaledSources []string

	for sourceID, oldGain := range s.InputGains {
		// Only scale gains for sources with recent activity
		if s.hasRecentActivity(sourceID) {
			newGain := oldGain * scalingFactor

			// Apply biological bounds
			if newGain < s.Config.MinReceptorGain {
				newGain = s.Config.MinReceptorGain
			} else if newGain > s.Config.MaxReceptorGain {
				newGain = s.Config.MaxReceptorGain
			}

			s.InputGains[sourceID] = newGain
			scaledSources = append(scaledSources, sourceID)
		}
	}

	return scaledSources
}

// hasRecentActivity checks if a source has activity within the sampling window
func (s *SynapticScalingState) hasRecentActivity(sourceID string) bool {
	activities, exists := s.InputActivityHistory[sourceID]
	if !exists || len(activities) == 0 {
		return false
	}

	cutoff := time.Now().Add(-s.Config.ActivitySamplingWindow)
	for _, activity := range activities {
		if activity.Timestamp.After(cutoff) {
			return true
		}
	}
	return false
}

// addToScalingHistory adds a scaling factor to the history
func (s *SynapticScalingState) addToScalingHistory(scalingFactor float64) {
	s.Config.ScalingHistory = append(s.Config.ScalingHistory, scalingFactor)

	// Limit history size
	if len(s.Config.ScalingHistory) > s.Config.MaxHistorySize {
		start := len(s.Config.ScalingHistory) - s.Config.MaxHistorySize
		s.Config.ScalingHistory = s.Config.ScalingHistory[start:]
	}
}

// ============================================================================
// MONITORING AND ANALYSIS
// ============================================================================

// GetScalingHistory returns a copy of recent scaling factors
func (s *SynapticScalingState) GetScalingHistory() []float64 {
	history := make([]float64, len(s.Config.ScalingHistory))
	copy(history, s.Config.ScalingHistory)
	return history
}

// GetScalingStatus returns current scaling system status
func (s *SynapticScalingState) GetScalingStatus() map[string]interface{} {
	s.inputGainMutex.RLock()
	s.inputActivityMutex.RLock()
	defer s.inputGainMutex.RUnlock()
	defer s.inputActivityMutex.RUnlock()

	avgStrength, activeCount := s.calculateAverageInputStrength()

	return map[string]interface{}{
		"enabled":               s.Config.Enabled,
		"target_strength":       s.Config.TargetInputStrength,
		"current_avg_strength":  avgStrength,
		"active_source_count":   activeCount,
		"total_source_count":    len(s.InputGains),
		"last_scaling_update":   s.Config.LastScalingUpdate,
		"scaling_interval":      s.Config.ScalingInterval,
		"time_until_next":       s.Config.ScalingInterval - time.Since(s.Config.LastScalingUpdate),
		"scaling_history_count": len(s.Config.ScalingHistory),
	}
}

// GetInputActivitySummary returns summary of recent input activity
func (s *SynapticScalingState) GetInputActivitySummary() map[string]interface{} {
	s.inputActivityMutex.RLock()
	defer s.inputActivityMutex.RUnlock()

	summary := make(map[string]interface{})
	cutoff := time.Now().Add(-s.Config.ActivitySamplingWindow)

	for sourceID, activities := range s.InputActivityHistory {
		var recentCount int
		var recentSum float64

		for _, activity := range activities {
			if activity.Timestamp.After(cutoff) {
				recentCount++
				recentSum += math.Abs(activity.EffectiveValue)
			}
		}

		var avgActivity float64
		if recentCount > 0 {
			avgActivity = recentSum / float64(recentCount)
		}

		summary[sourceID] = map[string]interface{}{
			"recent_count":   recentCount,
			"avg_activity":   avgActivity,
			"total_recorded": len(activities),
		}
	}

	return summary
}
