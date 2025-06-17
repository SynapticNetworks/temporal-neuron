/*
=================================================================================
SYNAPTIC ACTIVITY MONITOR - BIOLOGICAL ACTIVITY TRACKING SYSTEM
=================================================================================

This module implements comprehensive monitoring of synaptic activity patterns,
health metrics, and biological performance indicators. It tracks transmission
events, plasticity changes, and metabolic activity to provide detailed insights
into synaptic function and network dynamics.

BIOLOGICAL FUNCTIONS MONITORED:
1. Transmission Events - Success/failure rates, timing, vesicle usage
2. Plasticity Dynamics - STDP events, weight changes, learning patterns
3. Metabolic Activity - Energy costs, vesicle recycling, resource usage
4. Health Indicators - Activity levels, connectivity health, failure patterns
5. Performance Metrics - Throughput, latency, reliability measures

DESIGN PRINCIPLES:
- Low-overhead monitoring that doesn't impact performance
- Biologically meaningful metrics and thresholds
- Rolling window analysis for real-time health assessment
- Memory-efficient storage with automatic cleanup
- Thread-safe concurrent access for multi-synapse networks

INTEGRATION:
- Used internally by EnhancedSynapse for self-monitoring
- Provides data for matrix-wide network analysis
- Supports debugging and biological validation
- Enables adaptive network optimization based on activity patterns
=================================================================================
*/

package synapse

import (
	"math"
	"sync"
	"time"
)

// =================================================================================
// BIOLOGICAL MONITORING CONSTANTS
// =================================================================================

// === ANALYSIS WINDOW CONFIGURATION ===
const (
	// ANALYSIS_WINDOW_DURATION defines the time window for activity analysis
	// Biological basis: Matches short-term memory and plasticity timescales
	// Experimental range: 5-30 seconds for synaptic activity analysis
	ANALYSIS_WINDOW_DURATION = 10 * time.Second

	// HEALTH_UPDATE_INTERVAL controls frequency of health recalculation
	// Performance optimization: Balances accuracy with computational efficiency
	// Too frequent: wastes CPU, too infrequent: misses rapid changes
	HEALTH_UPDATE_INTERVAL = 1 * time.Second

	// PLASTICITY_ANALYSIS_WINDOW for plasticity-specific trend analysis
	// Biological basis: Learning consolidation occurs over minutes
	// Research: Long-term potentiation stabilization timescale
	PLASTICITY_ANALYSIS_WINDOW = 1 * time.Minute

	// PLASTICITY_REWARD_WINDOW for recent plasticity bonus
	// Biological basis: Synapses with recent plasticity are healthier
	// Experimental evidence: Activity-dependent synaptic maintenance
	PLASTICITY_REWARD_WINDOW = 30 * time.Second
)

// === HEALTH ASSESSMENT THRESHOLDS ===
const (
	// INACTIVITY_PENALTY_THRESHOLD - time after which inactivity reduces health
	// Biological basis: Synapses need regular activity to maintain health
	// Research: Activity-dependent synaptic maintenance mechanisms
	INACTIVITY_PENALTY_THRESHOLD = 2 * time.Hour

	// INACTIVITY_ISSUE_THRESHOLD - time indicating serious inactivity problem
	// Clinical significance: Extended inactivity suggests dysfunction
	// Pruning relevance: Candidate for synaptic elimination
	INACTIVITY_ISSUE_THRESHOLD = 6 * time.Hour

	// RELIABILITY_ISSUE_THRESHOLD - minimum acceptable transmission reliability
	// Biological range: Healthy synapses maintain >70% reliability
	// Pathological: <50% indicates serious transmission problems
	RELIABILITY_ISSUE_THRESHOLD = 0.70

	// CONSISTENCY_ISSUE_THRESHOLD - minimum acceptable activity consistency
	// Statistical basis: Coefficient of variation analysis
	// Healthy synapses: Relatively consistent inter-event intervals
	CONSISTENCY_ISSUE_THRESHOLD = 0.60

	// PLASTICITY_ISSUE_THRESHOLD - minimum plasticity responsiveness
	// Learning capability: Healthy synapses respond to learning signals
	// Range: 0.0 (no plasticity) to 1.0 (highly responsive)
	PLASTICITY_ISSUE_THRESHOLD = 0.30

	// EFFICIENCY_ISSUE_THRESHOLD - minimum acceptable metabolic efficiency
	// Energy considerations: Efficient energy use per successful transmission
	// Biological constraint: Metabolic cost-benefit optimization
	EFFICIENCY_ISSUE_THRESHOLD = 0.50

	// PRECISION_ISSUE_THRESHOLD - minimum acceptable temporal precision
	// Timing consistency: Important for precise neural coding
	// Information processing: Temporal precision affects computation
	PRECISION_ISSUE_THRESHOLD = 0.60
)

// === METABOLIC COST CONSTANTS ===
const (
	// BASE_TRANSMISSION_COST - basic energy cost per transmission attempt
	// Biological basis: ATP cost of action potential propagation
	// Research: Energy requirements of synaptic transmission
	BASE_TRANSMISSION_COST = 1.0

	// SUCCESS_BONUS_COST - additional cost for successful transmission
	// Vesicle release: Energy cost of vesicle fusion and recycling
	// Neurotransmitter synthesis: Cost of replenishing neurotransmitters
	SUCCESS_BONUS_COST = 0.5

	// VESICLE_RELEASE_COST - cost of vesicle release and recycling
	// Biological process: Endocytosis, vesicle reformation, NT loading
	// Experimental measurement: ATP molecules per vesicle cycle
	VESICLE_RELEASE_COST = 2.0
)

// =================================================================================
// ACTIVITY MONITOR IMPLEMENTATION
// =================================================================================

// SynapticActivityMonitor tracks comprehensive activity patterns and health metrics
// for individual synapses, providing biological insights and performance analytics
type SynapticActivityMonitor struct {
	// === IDENTIFICATION ===
	synapseID string // Synapse being monitored

	// === TRANSMISSION METRICS ===
	transmissionCount       int64         // Total transmission attempts
	successfulTransmissions int64         // Successful vesicle releases
	failedTransmissions     int64         // Failed transmission attempts
	totalLatency            time.Duration // Cumulative processing latency
	lastTransmission        time.Time     // Most recent transmission time

	// === PLASTICITY METRICS ===
	plasticityEvents    []PlasticityEvent // Recent plasticity events
	totalWeightChange   float64           // Cumulative absolute weight change
	lastPlasticityEvent time.Time         // Most recent plasticity event
	weightHistory       []weightSnapshot  // Historical weight values

	// === HEALTH AND PERFORMANCE ===
	healthScore      float64   // Overall synaptic health (0.0-1.0)
	lastHealthUpdate time.Time // Last health calculation time
	activityLevel    float64   // Current activity rate (Hz)
	reliabilityScore float64   // Transmission reliability (0.0-1.0)

	// === METABOLIC TRACKING ===
	totalMetabolicCost  float64   // Cumulative energy expenditure
	vesicleUsageRate    float64   // Vesicles consumed per second
	lastMetabolicUpdate time.Time // Last metabolic calculation

	// === ROLLING WINDOW ANALYSIS ===
	recentEvents    []TransmissionEvent // Recent transmission events
	analysisWindow  time.Duration       // Time window for analysis
	maxEventHistory int                 // Maximum events to track

	// === THREAD SAFETY ===
	mu sync.RWMutex // Protects all monitor state
}

// =================================================================================
// SUPPORTING DATA STRUCTURES
// =================================================================================

// TransmissionEvent records details of a single transmission event
type TransmissionEvent struct {
	Timestamp       time.Time     `json:"timestamp"`
	Success         bool          `json:"success"`
	VesicleReleased bool          `json:"vesicle_released"`
	ProcessingTime  time.Duration `json:"processing_time"`
	SignalStrength  float64       `json:"signal_strength"`
	CalciumLevel    float64       `json:"calcium_level"`
	MetabolicCost   float64       `json:"metabolic_cost"`
	ErrorType       string        `json:"error_type,omitempty"`
}

// weightSnapshot stores historical weight values for trend analysis
type weightSnapshot struct {
	Timestamp time.Time `json:"timestamp"`
	Weight    float64   `json:"weight"`
	Event     string    `json:"event"` // "STDP", "manual", "initialization"
}

// HealthAssessment provides detailed health analysis
type HealthAssessment struct {
	OverallScore    float64            `json:"overall_score"`
	ComponentScores map[string]float64 `json:"component_scores"`
	IssuesDetected  []string           `json:"issues_detected"`
	Recommendations []string           `json:"recommendations"`
	LastAssessment  time.Time          `json:"last_assessment"`
	TrendAnalysis   TrendAnalysis      `json:"trend_analysis"`
}

// TrendAnalysis provides statistical analysis of activity trends
type TrendAnalysis struct {
	ActivityTrend     string  `json:"activity_trend"`     // "increasing", "decreasing", "stable"
	WeightTrend       string  `json:"weight_trend"`       // Direction of weight changes
	ReliabilityTrend  string  `json:"reliability_trend"`  // Transmission reliability trend
	PredictedLifetime float64 `json:"predicted_lifetime"` // Expected synapse lifetime (hours)
	ConfidenceLevel   float64 `json:"confidence_level"`   // Prediction confidence (0.0-1.0)
}

// =================================================================================
// CONSTRUCTOR AND INITIALIZATION
// =================================================================================

// NewSynapticActivityMonitor creates a comprehensive activity monitor for a synapse
//
// BIOLOGICAL RATIONALE:
// Real synapses are constantly monitored by astrocytes and other glial cells that
// track their health, activity patterns, and metabolic needs. This monitor models
// that biological surveillance system, providing the data needed for:
// - Activity-dependent synaptic scaling
// - Metabolic resource allocation
// - Synaptic pruning decisions
// - Network optimization and adaptation
//
// PERFORMANCE OPTIMIZATION:
// Uses rolling windows and efficient data structures to minimize memory usage
// while maintaining sufficient history for meaningful biological analysis.
func NewSynapticActivityMonitor(synapseID string) *SynapticActivityMonitor {
	now := time.Now()

	return &SynapticActivityMonitor{
		// Identity
		synapseID: synapseID,

		// Initialize metrics
		healthScore:      1.0, // Perfect health initially
		reliabilityScore: 1.0, // Perfect reliability initially
		activityLevel:    0.0, // No activity initially

		// Time tracking
		lastHealthUpdate:    now,
		lastTransmission:    now,
		lastPlasticityEvent: now,
		lastMetabolicUpdate: now,

		// Analysis configuration
		analysisWindow:  ANALYSIS_WINDOW_DURATION,
		maxEventHistory: MAX_ACTIVITY_HISTORY_SIZE,

		// Initialize collections
		plasticityEvents: make([]PlasticityEvent, 0),
		weightHistory:    make([]weightSnapshot, 0),
		recentEvents:     make([]TransmissionEvent, 0),
	}
}

// =================================================================================
// TRANSMISSION EVENT RECORDING
// =================================================================================

// RecordTransmission logs a transmission event with comprehensive biological details
//
// BIOLOGICAL SIGNIFICANCE:
// Each transmission represents a complex biological process involving:
// - Calcium influx and vesicle fusion
// - Neurotransmitter release and diffusion
// - Metabolic energy consumption
// - Synaptic machinery utilization
//
// This method captures all relevant aspects for biological analysis and health monitoring.
func (sam *SynapticActivityMonitor) RecordTransmission(success bool, vesicleReleased bool, processingTime time.Duration) {
	sam.mu.Lock()
	defer sam.mu.Unlock()

	now := time.Now()

	// Update basic counters
	sam.transmissionCount++
	if success {
		sam.successfulTransmissions++
	} else {
		sam.failedTransmissions++
	}

	// Update timing metrics
	sam.totalLatency += processingTime
	sam.lastTransmission = now

	// Create detailed transmission event
	event := TransmissionEvent{
		Timestamp:       now,
		Success:         success,
		VesicleReleased: vesicleReleased,
		ProcessingTime:  processingTime,
		MetabolicCost:   sam.calculateTransmissionCost(success, vesicleReleased),
	}

	// Add to recent events (with cleanup)
	sam.addRecentEvent(event)

	// Update real-time metrics
	sam.updateActivityLevel()
	sam.updateReliabilityScore()
	sam.updateMetabolicTracking(event.MetabolicCost)

	// Trigger health update if enough time has passed
	if now.Sub(sam.lastHealthUpdate) > HEALTH_UPDATE_INTERVAL {
		sam.updateHealthScore()
	}
}

// RecordTransmissionWithDetails logs transmission with additional biological context
func (sam *SynapticActivityMonitor) RecordTransmissionWithDetails(
	success bool,
	vesicleReleased bool,
	processingTime time.Duration,
	signalStrength float64,
	calciumLevel float64,
	errorType string) {

	sam.mu.Lock()
	defer sam.mu.Unlock()

	now := time.Now()

	// Update basic metrics
	sam.transmissionCount++
	if success {
		sam.successfulTransmissions++
	} else {
		sam.failedTransmissions++
	}

	sam.totalLatency += processingTime
	sam.lastTransmission = now

	// Create detailed event with biological context
	event := TransmissionEvent{
		Timestamp:       now,
		Success:         success,
		VesicleReleased: vesicleReleased,
		ProcessingTime:  processingTime,
		SignalStrength:  signalStrength,
		CalciumLevel:    calciumLevel,
		MetabolicCost:   sam.calculateTransmissionCost(success, vesicleReleased),
		ErrorType:       errorType,
	}

	sam.addRecentEvent(event)
	sam.updateDerivedMetrics()
}

// =================================================================================
// PLASTICITY EVENT RECORDING
// =================================================================================

// RecordPlasticity logs a plasticity event and updates learning-related metrics
//
// BIOLOGICAL CONTEXT:
// Plasticity events represent molecular changes in synaptic strength driven by:
// - STDP (spike-timing dependent plasticity)
// - Homeostatic scaling
// - Neuromodulator-induced changes
// - Metaplastic adaptations
//
// Tracking these events enables analysis of learning patterns and synaptic stability.
func (sam *SynapticActivityMonitor) RecordPlasticity(event PlasticityEvent) {
	sam.mu.Lock()
	defer sam.mu.Unlock()

	// Add to plasticity history
	sam.plasticityEvents = append(sam.plasticityEvents, event)
	sam.lastPlasticityEvent = event.Timestamp

	// Track weight changes for trend analysis
	sam.totalWeightChange += math.Abs(event.WeightChange)

	// Add weight snapshot
	weightSnap := weightSnapshot{
		Timestamp: event.Timestamp,
		Weight:    event.WeightAfter,
		Event:     event.EventType.String(),
	}
	sam.weightHistory = append(sam.weightHistory, weightSnap)

	// Cleanup old events to manage memory
	sam.cleanupPlasticityHistory()
	sam.cleanupWeightHistory()

	// Update health score based on plasticity activity
	sam.updateHealthFromPlasticity(event)
}

// =================================================================================
// HEALTH ASSESSMENT AND ANALYSIS
// =================================================================================

// GetActivityInfo returns comprehensive activity information for external monitoring
func (sam *SynapticActivityMonitor) GetActivityInfo() SynapticActivityInfo {
	sam.mu.RLock()
	defer sam.mu.RUnlock()

	now := time.Now()

	// Calculate transmission rate
	var transmissionRate float64
	if sam.transmissionCount > 0 {
		duration := now.Sub(sam.lastHealthUpdate)
		if duration > 0 {
			transmissionRate = float64(sam.transmissionCount) / duration.Seconds()
		}
	}

	// Calculate average delay
	var averageDelay time.Duration
	if sam.transmissionCount > 0 {
		averageDelay = sam.totalLatency / time.Duration(sam.transmissionCount)
	}

	// Calculate weight change rate
	var weightChangeRate float64
	if len(sam.plasticityEvents) > 0 {
		timeDiff := now.Sub(sam.plasticityEvents[0].Timestamp)
		if timeDiff > 0 {
			weightChangeRate = sam.totalWeightChange / timeDiff.Hours()
		}
	}

	return SynapticActivityInfo{
		SynapseID:               sam.synapseID,
		LastUpdate:              now,
		TotalTransmissions:      sam.transmissionCount,
		SuccessfulTransmissions: sam.successfulTransmissions,
		LastTransmission:        sam.lastTransmission,
		TransmissionRate:        transmissionRate,
		TotalPlasticityEvents:   int64(len(sam.plasticityEvents)),
		LastPlasticityEvent:     sam.lastPlasticityEvent,
		WeightChangeRate:        weightChangeRate,
		AverageDelay:            averageDelay,
		HealthScore:             sam.healthScore,
		ActivityLevel:           sam.activityLevel,
		LastMaintenance:         sam.lastHealthUpdate,
		MetabolicActivity:       sam.totalMetabolicCost,
		EffectiveStrength:       sam.calculateEffectiveStrength(),
	}
}

// PerformHealthAssessment conducts comprehensive health analysis
//
// BIOLOGICAL HEALTH INDICATORS:
// - Transmission reliability (vesicle release success rate)
// - Activity level consistency (stable vs fluctuating firing)
// - Plasticity responsiveness (ability to change when stimulated)
// - Metabolic efficiency (energy cost per successful transmission)
// - Temporal precision (consistency of transmission timing)
func (sam *SynapticActivityMonitor) PerformHealthAssessment() HealthAssessment {
	sam.mu.Lock()
	defer sam.mu.Unlock()

	now := time.Now()

	// Calculate component health scores
	componentScores := map[string]float64{
		"transmission_reliability":  sam.calculateTransmissionReliability(),
		"activity_consistency":      sam.calculateActivityConsistency(),
		"plasticity_responsiveness": sam.calculatePlasticityResponsiveness(),
		"metabolic_efficiency":      sam.calculateMetabolicEfficiency(),
		"temporal_precision":        sam.calculateTemporalPrecision(),
	}

	// Calculate overall health score (weighted average)
	weights := map[string]float64{
		"transmission_reliability":  0.3,
		"activity_consistency":      0.2,
		"plasticity_responsiveness": 0.2,
		"metabolic_efficiency":      0.15,
		"temporal_precision":        0.15,
	}

	overallScore := 0.0
	for component, score := range componentScores {
		overallScore += score * weights[component]
	}

	// Detect issues and generate recommendations
	issues := sam.detectHealthIssues(componentScores)
	recommendations := sam.generateRecommendations(componentScores, issues)

	// Perform trend analysis
	trendAnalysis := sam.analyzeTrends()

	return HealthAssessment{
		OverallScore:    overallScore,
		ComponentScores: componentScores,
		IssuesDetected:  issues,
		Recommendations: recommendations,
		LastAssessment:  now,
		TrendAnalysis:   trendAnalysis,
	}
}

// UpdateHealth recalculates all health metrics based on current activity
func (sam *SynapticActivityMonitor) UpdateHealth() {
	sam.mu.Lock()
	defer sam.mu.Unlock()

	sam.updateHealthScore()
	sam.updateActivityLevel()
	sam.updateReliabilityScore()
	sam.lastHealthUpdate = time.Now()
}

// =================================================================================
// INTERNAL METRIC CALCULATIONS
// =================================================================================

// updateHealthScore recalculates overall synaptic health
func (sam *SynapticActivityMonitor) updateHealthScore() {
	// Base health score starts at 1.0 (perfect health)
	baseHealth := 1.0

	// Penalize for transmission failures
	if sam.transmissionCount > 0 {
		successRate := float64(sam.successfulTransmissions) / float64(sam.transmissionCount)
		baseHealth *= successRate
	}

	// Penalize for inactivity
	timeSinceActivity := time.Since(sam.lastTransmission)
	if timeSinceActivity > INACTIVITY_PENALTY_THRESHOLD {
		inactivityPenalty := math.Min(0.5, timeSinceActivity.Hours()/24.0) // Max 50% penalty
		baseHealth *= (1.0 - inactivityPenalty)
	}

	// Reward for recent plasticity
	timeSincePlasticity := time.Since(sam.lastPlasticityEvent)
	if timeSincePlasticity < PLASTICITY_REWARD_WINDOW {
		plasticityBonus := 0.1 * (1.0 - timeSincePlasticity.Seconds()/PLASTICITY_REWARD_WINDOW.Seconds())
		baseHealth *= (1.0 + plasticityBonus)
	}

	// Clamp to valid range
	sam.healthScore = math.Max(0.0, math.Min(1.0, baseHealth))
}

// updateActivityLevel calculates current activity rate
func (sam *SynapticActivityMonitor) updateActivityLevel() {
	if len(sam.recentEvents) == 0 {
		sam.activityLevel = 0.0
		return
	}

	// Count successful transmissions in analysis window
	now := time.Now()
	cutoff := now.Add(-sam.analysisWindow)

	successCount := 0
	for _, event := range sam.recentEvents {
		if event.Timestamp.After(cutoff) && event.Success {
			successCount++
		}
	}

	// Calculate rate in Hz
	sam.activityLevel = float64(successCount) / sam.analysisWindow.Seconds()
}

// updateReliabilityScore calculates transmission reliability
func (sam *SynapticActivityMonitor) updateReliabilityScore() {
	if sam.transmissionCount == 0 {
		sam.reliabilityScore = 1.0 // No data, assume perfect
		return
	}

	// Recent reliability based on analysis window
	now := time.Now()
	cutoff := now.Add(-sam.analysisWindow)

	recentTotal := 0
	recentSuccess := 0

	for _, event := range sam.recentEvents {
		if event.Timestamp.After(cutoff) {
			recentTotal++
			if event.Success {
				recentSuccess++
			}
		}
	}

	if recentTotal > 0 {
		sam.reliabilityScore = float64(recentSuccess) / float64(recentTotal)
	}
}

// updateMetabolicTracking updates energy cost metrics
func (sam *SynapticActivityMonitor) updateMetabolicTracking(eventCost float64) {
	sam.totalMetabolicCost += eventCost

	// Update vesicle usage rate
	now := time.Now()
	timeDiff := now.Sub(sam.lastMetabolicUpdate)
	if timeDiff > 0 {
		// Estimate vesicle usage from recent successful transmissions
		recentSuccess := sam.countRecentSuccessfulTransmissions()
		sam.vesicleUsageRate = float64(recentSuccess) / sam.analysisWindow.Seconds()
	}

	sam.lastMetabolicUpdate = now
}

// =================================================================================
// HELPER FUNCTIONS
// =================================================================================

// addRecentEvent adds an event to the rolling window with automatic cleanup
func (sam *SynapticActivityMonitor) addRecentEvent(event TransmissionEvent) {
	sam.recentEvents = append(sam.recentEvents, event)

	// Remove events outside analysis window
	cutoff := time.Now().Add(-sam.analysisWindow)
	validEvents := make([]TransmissionEvent, 0)

	for _, e := range sam.recentEvents {
		if e.Timestamp.After(cutoff) {
			validEvents = append(validEvents, e)
		}
	}

	sam.recentEvents = validEvents

	// Enforce maximum history size
	if len(sam.recentEvents) > sam.maxEventHistory {
		sam.recentEvents = sam.recentEvents[len(sam.recentEvents)-sam.maxEventHistory:]
	}
}

// calculateTransmissionCost estimates metabolic cost of transmission
func (sam *SynapticActivityMonitor) calculateTransmissionCost(success bool, vesicleReleased bool) float64 {
	baseCost := BASE_TRANSMISSION_COST

	if success {
		baseCost += SUCCESS_BONUS_COST
	}

	if vesicleReleased {
		baseCost += VESICLE_RELEASE_COST
	}

	return baseCost
}

// countRecentSuccessfulTransmissions counts successes in analysis window
func (sam *SynapticActivityMonitor) countRecentSuccessfulTransmissions() int {
	cutoff := time.Now().Add(-sam.analysisWindow)
	count := 0

	for _, event := range sam.recentEvents {
		if event.Timestamp.After(cutoff) && event.Success {
			count++
		}
	}

	return count
}

// updateDerivedMetrics updates all derived metrics
func (sam *SynapticActivityMonitor) updateDerivedMetrics() {
	sam.updateActivityLevel()
	sam.updateReliabilityScore()

	if time.Since(sam.lastHealthUpdate) > HEALTH_UPDATE_INTERVAL {
		sam.updateHealthScore()
	}
}

// cleanupPlasticityHistory removes old plasticity events
func (sam *SynapticActivityMonitor) cleanupPlasticityHistory() {
	if len(sam.plasticityEvents) > MAX_PLASTICITY_HISTORY {
		sam.plasticityEvents = sam.plasticityEvents[len(sam.plasticityEvents)-MAX_PLASTICITY_HISTORY/2:]
	}
}

// cleanupWeightHistory removes old weight snapshots
func (sam *SynapticActivityMonitor) cleanupWeightHistory() {
	if len(sam.weightHistory) > MAX_WEIGHT_HISTORY {
		sam.weightHistory = sam.weightHistory[len(sam.weightHistory)-MAX_WEIGHT_HISTORY/2:]
	}
}

// =================================================================================
// BIOLOGICAL ANALYSIS FUNCTIONS
// =================================================================================

// calculateTransmissionReliability assesses vesicle release reliability
func (sam *SynapticActivityMonitor) calculateTransmissionReliability() float64 {
	if sam.transmissionCount == 0 {
		return 1.0 // No data, assume perfect
	}

	return float64(sam.successfulTransmissions) / float64(sam.transmissionCount)
}

// calculateActivityConsistency measures firing pattern regularity
func (sam *SynapticActivityMonitor) calculateActivityConsistency() float64 {
	if len(sam.recentEvents) < 3 {
		return 1.0 // Insufficient data
	}

	// Calculate coefficient of variation of inter-event intervals
	intervals := make([]float64, 0)
	for i := 1; i < len(sam.recentEvents); i++ {
		interval := sam.recentEvents[i].Timestamp.Sub(sam.recentEvents[i-1].Timestamp)
		intervals = append(intervals, interval.Seconds())
	}

	if len(intervals) == 0 {
		return 1.0
	}

	// Calculate mean and standard deviation
	mean := 0.0
	for _, interval := range intervals {
		mean += interval
	}
	mean /= float64(len(intervals))

	variance := 0.0
	for _, interval := range intervals {
		diff := interval - mean
		variance += diff * diff
	}
	variance /= float64(len(intervals))
	stddev := math.Sqrt(variance)

	if mean == 0 {
		return 1.0
	}

	// Coefficient of variation (lower = more consistent)
	cv := stddev / mean

	// Convert to consistency score (1.0 = perfectly consistent)
	return math.Max(0.0, 1.0-cv)
}

// calculatePlasticityResponsiveness measures learning capability
func (sam *SynapticActivityMonitor) calculatePlasticityResponsiveness() float64 {
	if len(sam.plasticityEvents) == 0 {
		return 0.5 // No plasticity data, neutral score
	}

	// Recent plasticity activity boosts responsiveness
	recentEvents := 0
	cutoff := time.Now().Add(-PLASTICITY_ANALYSIS_WINDOW)

	for _, event := range sam.plasticityEvents {
		if event.Timestamp.After(cutoff) {
			recentEvents++
		}
	}

	// Score based on recent plasticity events
	maxExpected := 10 // Maximum expected events in analysis window
	return math.Min(1.0, float64(recentEvents)/float64(maxExpected))
}

// calculateMetabolicEfficiency measures energy efficiency
func (sam *SynapticActivityMonitor) calculateMetabolicEfficiency() float64 {
	if sam.successfulTransmissions == 0 {
		return 0.5 // No successful transmissions, neutral score
	}

	// Energy cost per successful transmission
	costPerSuccess := sam.totalMetabolicCost / float64(sam.successfulTransmissions)

	// Score based on efficiency (lower cost = higher efficiency)
	optimalCost := BASE_TRANSMISSION_COST + SUCCESS_BONUS_COST + VESICLE_RELEASE_COST
	efficiency := optimalCost / costPerSuccess

	return math.Min(1.0, efficiency)
}

// calculateTemporalPrecision measures timing consistency
func (sam *SynapticActivityMonitor) calculateTemporalPrecision() float64 {
	if len(sam.recentEvents) < 2 {
		return 1.0 // Insufficient data
	}

	// Calculate processing time variability
	var processingTimes []float64
	for _, event := range sam.recentEvents {
		processingTimes = append(processingTimes, event.ProcessingTime.Seconds())
	}

	if len(processingTimes) == 0 {
		return 1.0
	}

	// Calculate coefficient of variation
	mean := 0.0
	for _, time := range processingTimes {
		mean += time
	}
	mean /= float64(len(processingTimes))

	variance := 0.0
	for _, time := range processingTimes {
		diff := time - mean
		variance += diff * diff
	}
	variance /= float64(len(processingTimes))
	stddev := math.Sqrt(variance)

	if mean == 0 {
		return 1.0
	}

	cv := stddev / mean
	return math.Max(0.0, 1.0-cv) // Lower variability = higher precision
}

// detectHealthIssues identifies potential problems
func (sam *SynapticActivityMonitor) detectHealthIssues(scores map[string]float64) []string {
	var issues []string

	if scores["transmission_reliability"] < RELIABILITY_ISSUE_THRESHOLD {
		issues = append(issues, "poor_transmission_reliability")
	}

	if scores["activity_consistency"] < CONSISTENCY_ISSUE_THRESHOLD {
		issues = append(issues, "inconsistent_activity_pattern")
	}

	if scores["plasticity_responsiveness"] < PLASTICITY_ISSUE_THRESHOLD {
		issues = append(issues, "reduced_plasticity_responsiveness")
	}

	if scores["metabolic_efficiency"] < EFFICIENCY_ISSUE_THRESHOLD {
		issues = append(issues, "poor_metabolic_efficiency")
	}

	if scores["temporal_precision"] < PRECISION_ISSUE_THRESHOLD {
		issues = append(issues, "poor_temporal_precision")
	}

	// Check for inactivity
	if time.Since(sam.lastTransmission) > INACTIVITY_ISSUE_THRESHOLD {
		issues = append(issues, "prolonged_inactivity")
	}

	return issues
}

// generateRecommendations suggests improvements based on health assessment
func (sam *SynapticActivityMonitor) generateRecommendations(scores map[string]float64, issues []string) []string {
	var recommendations []string

	for _, issue := range issues {
		switch issue {
		case "poor_transmission_reliability":
			recommendations = append(recommendations, "check_vesicle_availability_and_calcium_levels")
		case "inconsistent_activity_pattern":
			recommendations = append(recommendations, "investigate_input_signal_variability")
		case "reduced_plasticity_responsiveness":
			recommendations = append(recommendations, "verify_STDP_parameters_and_learning_rate")
		case "poor_metabolic_efficiency":
			recommendations = append(recommendations, "optimize_vesicle_recycling_and_usage_patterns")
		case "poor_temporal_precision":
			recommendations = append(recommendations, "check_delay_calculation_and_processing_latency")
		case "prolonged_inactivity":
			recommendations = append(recommendations, "consider_pruning_or_synaptic_scaling")
		}
	}

	return recommendations
}

// analyzeTrends performs statistical trend analysis
func (sam *SynapticActivityMonitor) analyzeTrends() TrendAnalysis {
	return TrendAnalysis{
		ActivityTrend:     sam.calculateActivityTrend(),
		WeightTrend:       sam.calculateWeightTrend(),
		ReliabilityTrend:  sam.calculateReliabilityTrend(),
		PredictedLifetime: sam.predictSynapticLifetime(),
		ConfidenceLevel:   sam.calculatePredictionConfidence(),
	}
}

// calculateActivityTrend analyzes activity level trends
func (sam *SynapticActivityMonitor) calculateActivityTrend() string {
	if len(sam.recentEvents) < 10 {
		return "insufficient_data"
	}

	// Split recent events into two halves and compare activity rates
	mid := len(sam.recentEvents) / 2

	firstHalf := sam.recentEvents[:mid]
	secondHalf := sam.recentEvents[mid:]

	firstRate := float64(len(firstHalf)) / sam.analysisWindow.Seconds() * 2
	secondRate := float64(len(secondHalf)) / sam.analysisWindow.Seconds() * 2

	rateDiff := (secondRate - firstRate) / firstRate

	if rateDiff > 0.1 {
		return "increasing"
	} else if rateDiff < -0.1 {
		return "decreasing"
	} else {
		return "stable"
	}
}

// calculateWeightTrend analyzes synaptic weight trends
func (sam *SynapticActivityMonitor) calculateWeightTrend() string {
	if len(sam.weightHistory) < 3 {
		return "insufficient_data"
	}

	// Linear regression on recent weight history
	recent := sam.weightHistory
	if len(recent) > 20 {
		recent = recent[len(recent)-20:] // Last 20 weight changes
	}

	// Simple slope calculation
	n := len(recent)
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0

	for i, snap := range recent {
		x := float64(i)
		y := snap.Weight
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumX2 - sumX*sumX)

	if slope > 0.01 {
		return "increasing"
	} else if slope < -0.01 {
		return "decreasing"
	} else {
		return "stable"
	}
}

// calculateReliabilityTrend analyzes transmission reliability trends
func (sam *SynapticActivityMonitor) calculateReliabilityTrend() string {
	if len(sam.recentEvents) < 20 {
		return "insufficient_data"
	}

	// Split events into first and second half
	mid := len(sam.recentEvents) / 2
	firstHalf := sam.recentEvents[:mid]
	secondHalf := sam.recentEvents[mid:]

	// Calculate reliability for each half
	firstSuccess := 0
	for _, event := range firstHalf {
		if event.Success {
			firstSuccess++
		}
	}
	firstReliability := float64(firstSuccess) / float64(len(firstHalf))

	secondSuccess := 0
	for _, event := range secondHalf {
		if event.Success {
			secondSuccess++
		}
	}
	secondReliability := float64(secondSuccess) / float64(len(secondHalf))

	// Compare trends
	reliabilityChange := (secondReliability - firstReliability) / firstReliability

	if reliabilityChange > 0.1 {
		return "improving"
	} else if reliabilityChange < -0.1 {
		return "degrading"
	} else {
		return "stable"
	}
}

// predictSynapticLifetime estimates expected synapse lifetime
func (sam *SynapticActivityMonitor) predictSynapticLifetime() float64 {
	// Based on current health score and trends
	baseLifetime := STABLE_SYNAPSE_LIFETIME.Hours()

	// Adjust based on health score
	healthFactor := sam.healthScore

	// Adjust based on activity level
	activityFactor := 1.0
	if sam.activityLevel > 0 {
		activityFactor = math.Min(2.0, sam.activityLevel/10.0+0.5)
	} else {
		activityFactor = 0.1 // Very low lifetime for inactive synapses
	}

	// Adjust based on reliability
	reliabilityFactor := sam.reliabilityScore

	return baseLifetime * healthFactor * activityFactor * reliabilityFactor
}

// calculatePredictionConfidence estimates confidence in lifetime prediction
func (sam *SynapticActivityMonitor) calculatePredictionConfidence() float64 {
	// Based on amount of data available
	dataPoints := len(sam.recentEvents) + len(sam.plasticityEvents)

	maxConfidence := 100.0 // 100 data points for full confidence
	confidence := math.Min(1.0, float64(dataPoints)/maxConfidence)

	// Reduce confidence if synapse is very new
	age := time.Since(sam.lastHealthUpdate)
	if age < ANALYSIS_WINDOW_DURATION {
		ageFactor := age.Seconds() / ANALYSIS_WINDOW_DURATION.Seconds()
		confidence *= ageFactor
	}

	return confidence
}

// calculateEffectiveStrength estimates current functional synaptic strength
func (sam *SynapticActivityMonitor) calculateEffectiveStrength() float64 {
	// Combine weight with reliability and activity
	baseStrength := 1.0 // Would get actual weight from synapse

	effectiveStrength := baseStrength * sam.reliabilityScore * math.Min(1.0, sam.activityLevel/10.0)

	return math.Min(1.0, effectiveStrength)
}

// updateHealthFromPlasticity updates health based on plasticity event
func (sam *SynapticActivityMonitor) updateHealthFromPlasticity(event PlasticityEvent) {
	// Plasticity events generally indicate healthy, active synapses
	plasticityBonus := 0.05 // 5% health bonus for plasticity activity

	sam.healthScore = math.Min(1.0, sam.healthScore+plasticityBonus)
}
