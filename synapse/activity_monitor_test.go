/*
=================================================================================
ACTIVITY MONITOR BASIC TESTS - CORE FUNCTIONALITY VALIDATION
=================================================================================

This file contains basic unit tests for the SynapticActivityMonitor, validating
core functionality including transmission recording, plasticity tracking, and
health assessment under normal operating conditions.

BIOLOGICAL CONTEXT:
These tests model the fundamental monitoring capabilities that astrocytes perform
on synapses in healthy neural tissue. Real astrocytes continuously track:
- Synaptic transmission success rates
- Activity patterns and consistency
- Plasticity events and weight changes
- Metabolic costs and efficiency
- Overall synaptic health metrics

TEST CATEGORIES:
1. Constructor and Initialization
2. Basic Transmission Recording
3. Plasticity Event Tracking
4. Health Score Calculation
5. Activity Information Retrieval
6. Basic Statistics and Metrics

BIOLOGICAL VALIDATION:
Each test validates that the monitor behaves like biological astrocytic surveillance,
maintaining accurate records that would be used for synaptic maintenance decisions.
=================================================================================
*/

package synapse

import (
	"testing"
	"time"
)

// =================================================================================
// CONSTRUCTOR AND INITIALIZATION TESTS
// =================================================================================

// TestActivityMonitorBasicNewSynapticActivityMonitor validates proper initialization
// BIOLOGICAL CONTEXT: Like astrocyte territorial establishment around new synapses
func TestActivityMonitorBasicNewSynapticActivityMonitor(t *testing.T) {
	t.Log("=== TESTING: Activity Monitor Initialization ===")
	t.Log("BIOLOGICAL MODEL: Astrocyte establishing surveillance of new synapse")
	t.Log("EXPECTED: Monitor initialized with perfect health, zero activity")

	synapseID := "test_synapse_001"
	monitor := NewSynapticActivityMonitor(synapseID)

	// Validate basic properties
	if monitor == nil {
		t.Fatal("Monitor should not be nil after creation")
	}

	if monitor.synapseID != synapseID {
		t.Errorf("Expected synapse ID %s, got %s", synapseID, monitor.synapseID)
	}

	// Validate initial health state (like healthy new synapse)
	expectedHealthScore := 1.0 // Perfect health initially
	if monitor.healthScore != expectedHealthScore {
		t.Errorf("Expected initial health score %.2f, got %.2f",
			expectedHealthScore, monitor.healthScore)
	}

	// Validate initial activity state
	expectedActivityLevel := 0.0 // No activity initially
	if monitor.activityLevel != expectedActivityLevel {
		t.Errorf("Expected initial activity level %.2f, got %.2f",
			expectedActivityLevel, monitor.activityLevel)
	}

	// Validate initial reliability
	expectedReliability := 1.0 // Perfect reliability initially
	if monitor.reliabilityScore != expectedReliability {
		t.Errorf("Expected initial reliability %.2f, got %.2f",
			expectedReliability, monitor.reliabilityScore)
	}

	// Validate collections are initialized
	if monitor.recentEvents == nil {
		t.Error("Recent events collection should be initialized")
	}

	if monitor.plasticityEvents == nil {
		t.Error("Plasticity events collection should be initialized")
	}

	if monitor.weightHistory == nil {
		t.Error("Weight history collection should be initialized")
	}

	t.Log("✓ Monitor initialized with biologically realistic defaults")
	t.Log("✓ Perfect initial health reflects new, undamaged synapse")
	t.Log("✓ Zero activity reflects absence of stimulation")
}

// =================================================================================
// BASIC TRANSMISSION RECORDING TESTS
// =================================================================================

// TestActivityMonitorBasicRecordTransmission validates basic transmission event recording
// BIOLOGICAL CONTEXT: Like astrocytes tracking successful vesicle releases
func TestActivityMonitorBasicRecordTransmission(t *testing.T) {
	t.Log("=== TESTING: Basic Transmission Recording ===")
	t.Log("BIOLOGICAL MODEL: Astrocyte monitoring vesicle release events")
	t.Log("EXPECTED: Accurate tracking of transmission success/failure rates")

	monitor := NewSynapticActivityMonitor("test_synapse")

	// Record a successful transmission
	processingTime := 2 * time.Millisecond
	monitor.RecordTransmission(true, true, processingTime)

	// Validate transmission counters
	expectedTotal := int64(1)
	expectedSuccessful := int64(1)
	expectedFailed := int64(0)

	if monitor.transmissionCount != expectedTotal {
		t.Errorf("Expected total transmissions %d, got %d",
			expectedTotal, monitor.transmissionCount)
	}

	if monitor.successfulTransmissions != expectedSuccessful {
		t.Errorf("Expected successful transmissions %d, got %d",
			expectedSuccessful, monitor.successfulTransmissions)
	}

	if monitor.failedTransmissions != expectedFailed {
		t.Errorf("Expected failed transmissions %d, got %d",
			expectedFailed, monitor.failedTransmissions)
	}

	// Validate timing tracking
	if monitor.totalLatency != processingTime {
		t.Errorf("Expected total latency %v, got %v",
			processingTime, monitor.totalLatency)
	}

	// Validate reliability calculation
	expectedReliability := 1.0 // 100% success rate
	if monitor.reliabilityScore != expectedReliability {
		t.Errorf("Expected reliability %.2f, got %.2f",
			expectedReliability, monitor.reliabilityScore)
	}

	t.Log("✓ Successful transmission recorded accurately")
	t.Log("✓ Timing information captured for biological realism")
	t.Log("✓ Reliability score reflects perfect transmission")

	// Record a failed transmission
	monitor.RecordTransmission(false, false, processingTime)

	// Validate updated counters
	expectedTotal = 2
	expectedFailed = 1

	if monitor.transmissionCount != expectedTotal {
		t.Errorf("Expected total transmissions %d, got %d",
			expectedTotal, monitor.transmissionCount)
	}

	if monitor.failedTransmissions != expectedFailed {
		t.Errorf("Expected failed transmissions %d, got %d",
			expectedFailed, monitor.failedTransmissions)
	}

	// Validate reliability update
	expectedReliability = 0.5 // 50% success rate
	if monitor.reliabilityScore != expectedReliability {
		t.Errorf("Expected reliability %.2f, got %.2f",
			expectedReliability, monitor.reliabilityScore)
	}

	t.Log("✓ Failed transmission recorded accurately")
	t.Log("✓ Reliability score updated to reflect transmission failures")
	t.Log("✓ Monitor accurately tracks synaptic performance degradation")
}

// TestActivityMonitorBasicRecordTransmissionWithDetails validates detailed transmission recording
// BIOLOGICAL CONTEXT: Comprehensive monitoring like detailed astrocytic surveillance
func TestActivityMonitorBasicRecordTransmissionWithDetails(t *testing.T) {
	t.Log("=== TESTING: Detailed Transmission Recording ===")
	t.Log("BIOLOGICAL MODEL: Comprehensive astrocytic monitoring with biological context")
	t.Log("EXPECTED: Full biological metadata captured for analysis")

	monitor := NewSynapticActivityMonitor("detailed_synapse")

	// Record transmission with full biological details
	processingTime := 1500 * time.Microsecond
	signalStrength := 0.75
	calciumLevel := 2.5
	errorType := ""

	monitor.RecordTransmissionWithDetails(
		true,           // success
		true,           // vesicle released
		processingTime, // processing time
		signalStrength, // signal strength
		calciumLevel,   // calcium level
		errorType,      // error type
	)

	// Validate basic counters
	if monitor.transmissionCount != 1 {
		t.Errorf("Expected 1 transmission, got %d", monitor.transmissionCount)
	}

	if monitor.successfulTransmissions != 1 {
		t.Errorf("Expected 1 successful transmission, got %d", monitor.successfulTransmissions)
	}

	// Validate detailed event was recorded
	if len(monitor.recentEvents) != 1 {
		t.Fatalf("Expected 1 recent event, got %d", len(monitor.recentEvents))
	}

	event := monitor.recentEvents[0]

	// Validate event details
	if !event.Success {
		t.Error("Event should be marked as successful")
	}

	if !event.VesicleReleased {
		t.Error("Event should indicate vesicle was released")
	}

	if event.ProcessingTime != processingTime {
		t.Errorf("Expected processing time %v, got %v",
			processingTime, event.ProcessingTime)
	}

	if event.SignalStrength != signalStrength {
		t.Errorf("Expected signal strength %.2f, got %.2f",
			signalStrength, event.SignalStrength)
	}

	if event.CalciumLevel != calciumLevel {
		t.Errorf("Expected calcium level %.2f, got %.2f",
			calciumLevel, event.CalciumLevel)
	}

	if event.ErrorType != errorType {
		t.Errorf("Expected error type '%s', got '%s'",
			errorType, event.ErrorType)
	}

	t.Log("✓ Detailed transmission recorded with full biological context")
	t.Log("✓ Signal strength and calcium levels tracked for plasticity analysis")
	t.Log("✓ Processing time captured for temporal precision assessment")
	t.Log("✓ Monitor provides comprehensive biological metadata")
}

// =================================================================================
// PLASTICITY EVENT TRACKING TESTS
// =================================================================================

// TestActivityMonitorBasicRecordPlasticity validates plasticity event recording
// BIOLOGICAL CONTEXT: Tracking synaptic strength changes like LTP/LTD monitoring
func TestActivityMonitorBasicRecordPlasticity(t *testing.T) {
	t.Log("=== TESTING: Plasticity Event Recording ===")
	t.Log("BIOLOGICAL MODEL: Astrocyte monitoring synaptic plasticity (LTP/LTD)")
	t.Log("EXPECTED: Accurate tracking of weight changes and plasticity patterns")

	monitor := NewSynapticActivityMonitor("plastic_synapse")

	// Create a plasticity event (LTP - strengthening)
	event := PlasticityEvent{
		SynapseID:    "plastic_synapse",
		EventType:    PlasticitySTDP,
		Timestamp:    time.Now(),
		WeightBefore: 0.5,
		WeightAfter:  0.6,
		WeightChange: 0.1,
		Context:      map[string]interface{}{"type": "LTP"},
	}

	monitor.RecordPlasticity(event)

	// Validate plasticity tracking
	if len(monitor.plasticityEvents) != 1 {
		t.Fatalf("Expected 1 plasticity event, got %d", len(monitor.plasticityEvents))
	}

	recordedEvent := monitor.plasticityEvents[0]
	if recordedEvent.SynapseID != event.SynapseID {
		t.Errorf("Expected synapse ID %s, got %s",
			event.SynapseID, recordedEvent.SynapseID)
	}

	if recordedEvent.WeightChange != event.WeightChange {
		t.Errorf("Expected weight change %.3f, got %.3f",
			event.WeightChange, recordedEvent.WeightChange)
	}

	// Validate weight history tracking
	if len(monitor.weightHistory) != 1 {
		t.Fatalf("Expected 1 weight history entry, got %d", len(monitor.weightHistory))
	}

	weightSnap := monitor.weightHistory[0]
	if weightSnap.Weight != event.WeightAfter {
		t.Errorf("Expected weight %.3f, got %.3f",
			event.WeightAfter, weightSnap.Weight)
	}

	// Validate total weight change tracking
	expectedTotalChange := 0.1
	if monitor.totalWeightChange != expectedTotalChange {
		t.Errorf("Expected total weight change %.3f, got %.3f",
			expectedTotalChange, monitor.totalWeightChange)
	}

	t.Log("✓ Plasticity event recorded accurately")
	t.Log("✓ Weight history maintained for trend analysis")
	t.Log("✓ Total weight change tracked for learning assessment")

	// Record LTD event (weakening)
	ltdEvent := PlasticityEvent{
		SynapseID:    "plastic_synapse",
		EventType:    PlasticitySTDP,
		Timestamp:    time.Now(),
		WeightBefore: 0.6,
		WeightAfter:  0.4,
		WeightChange: -0.2,
		Context:      map[string]interface{}{"type": "LTD"},
	}

	monitor.RecordPlasticity(ltdEvent)

	// Validate multiple events
	if len(monitor.plasticityEvents) != 2 {
		t.Fatalf("Expected 2 plasticity events, got %d", len(monitor.plasticityEvents))
	}

	// Validate cumulative weight change (absolute values)
	expectedTotalChange = 0.3 // |0.1| + |-0.2| = 0.3
	if monitor.totalWeightChange != expectedTotalChange {
		t.Errorf("Expected total weight change %.3f, got %.3f",
			expectedTotalChange, monitor.totalWeightChange)
	}

	t.Log("✓ Multiple plasticity events tracked correctly")
	t.Log("✓ Bidirectional plasticity (LTP/LTD) recorded")
	t.Log("✓ Cumulative plasticity activity measured for learning analysis")
}

// =================================================================================
// HEALTH ASSESSMENT TESTS
// =================================================================================

// TestActivityMonitorBasicUpdateHealth validates health score calculation
// BIOLOGICAL CONTEXT: Astrocytic assessment of synaptic viability and function
func TestActivityMonitorBasicUpdateHealth(t *testing.T) {
	t.Log("=== TESTING: Health Score Calculation ===")
	t.Log("BIOLOGICAL MODEL: Astrocyte assessing synaptic health and viability")
	t.Log("EXPECTED: Health scores reflect transmission reliability and activity")

	monitor := NewSynapticActivityMonitor("health_test_synapse")

	// Initially perfect health
	monitor.UpdateHealth()
	expectedInitialHealth := 1.0
	if monitor.healthScore != expectedInitialHealth {
		t.Errorf("Expected initial health %.2f, got %.2f",
			expectedInitialHealth, monitor.healthScore)
	}

	t.Log("✓ Initial health score reflects perfect condition")

	// Record some successful transmissions (should maintain good health)
	for i := 0; i < 5; i++ {
		monitor.RecordTransmission(true, true, time.Millisecond)
	}

	monitor.UpdateHealth()
	if monitor.healthScore < 0.9 {
		t.Errorf("Expected health score ≥ 0.9 with all successful transmissions, got %.2f",
			monitor.healthScore)
	}

	t.Log("✓ Successful transmissions maintain high health score")

	// Record some failures (should reduce health)
	for i := 0; i < 5; i++ {
		monitor.RecordTransmission(false, false, time.Millisecond)
	}

	monitor.UpdateHealth()
	if monitor.healthScore >= 0.9 {
		t.Errorf("Expected health score < 0.9 with 50%% failures, got %.2f",
			monitor.healthScore)
	}

	t.Log("✓ Transmission failures appropriately reduce health score")
	t.Log("✓ Health assessment reflects synaptic performance quality")
}

// TestActivityMonitorBasicPerformHealthAssessment validates comprehensive health analysis
// BIOLOGICAL CONTEXT: Detailed astrocytic evaluation for pruning decisions
func TestActivityMonitorBasicPerformHealthAssessment(t *testing.T) {
	t.Log("=== TESTING: Comprehensive Health Assessment ===")
	t.Log("BIOLOGICAL MODEL: Complete astrocytic evaluation for maintenance decisions")
	t.Log("EXPECTED: Detailed health analysis with component scores and recommendations")

	monitor := NewSynapticActivityMonitor("assessment_synapse")

	// Add some activity for realistic assessment
	for i := 0; i < 10; i++ {
		success := i < 8 // 80% success rate
		monitor.RecordTransmission(success, success, time.Millisecond)
	}

	// Add plasticity event
	plasticityEvent := PlasticityEvent{
		SynapseID:    "assessment_synapse",
		EventType:    PlasticitySTDP,
		Timestamp:    time.Now(),
		WeightBefore: 0.5,
		WeightAfter:  0.6,
		WeightChange: 0.1,
	}
	monitor.RecordPlasticity(plasticityEvent)

	// Perform comprehensive assessment
	assessment := monitor.PerformHealthAssessment()

	// Validate assessment structure
	if assessment.OverallScore <= 0.0 || assessment.OverallScore > 1.0 {
		t.Errorf("Overall score should be between 0 and 1, got %.3f",
			assessment.OverallScore)
	}

	// Validate component scores exist
	expectedComponents := []string{
		"transmission_reliability",
		"activity_consistency",
		"plasticity_responsiveness",
		"metabolic_efficiency",
		"temporal_precision",
	}

	for _, component := range expectedComponents {
		if score, exists := assessment.ComponentScores[component]; !exists {
			t.Errorf("Missing component score for %s", component)
		} else if score < 0.0 || score > 1.0 {
			t.Errorf("Invalid component score for %s: %.3f", component, score)
		}
	}

	// Validate issues detection
	if assessment.IssuesDetected == nil {
		t.Error("Issues detected should be initialized")
	}

	// Validate recommendations
	if assessment.Recommendations == nil {
		t.Error("Recommendations should be initialized")
	}

	// Validate trend analysis
	if assessment.TrendAnalysis.ActivityTrend == "" {
		t.Error("Activity trend should be analyzed")
	}

	t.Log("✓ Comprehensive health assessment completed")
	t.Log("✓ Component scores calculated for all biological factors")
	t.Log("✓ Issues and recommendations generated")
	t.Log("✓ Trend analysis provides predictive insights")
}

// =================================================================================
// ACTIVITY INFORMATION RETRIEVAL TESTS
// =================================================================================

// TestActivityMonitorBasicGetActivityInfo validates activity information retrieval
// BIOLOGICAL CONTEXT: Providing comprehensive synaptic status to matrix systems
func TestActivityMonitorBasicGetActivityInfo(t *testing.T) {
	t.Log("=== TESTING: Activity Information Retrieval ===")
	t.Log("BIOLOGICAL MODEL: Astrocyte reporting synaptic status to matrix systems")
	t.Log("EXPECTED: Comprehensive activity data for network coordination")

	synapseID := "info_test_synapse"
	monitor := NewSynapticActivityMonitor(synapseID)

	// Add some activity
	monitor.RecordTransmission(true, true, 2*time.Millisecond)
	monitor.RecordTransmission(true, true, 3*time.Millisecond)
	monitor.RecordTransmission(false, false, 1*time.Millisecond)

	// Add plasticity event
	plasticityEvent := PlasticityEvent{
		SynapseID:    synapseID,
		EventType:    PlasticitySTDP,
		Timestamp:    time.Now(),
		WeightBefore: 0.5,
		WeightAfter:  0.6,
		WeightChange: 0.1,
	}
	monitor.RecordPlasticity(plasticityEvent)

	// Get activity information
	info := monitor.GetActivityInfo()

	// Validate basic information
	if info.SynapseID != synapseID {
		t.Errorf("Expected synapse ID %s, got %s", synapseID, info.SynapseID)
	}

	if info.TotalTransmissions != 3 {
		t.Errorf("Expected 3 total transmissions, got %d", info.TotalTransmissions)
	}

	if info.SuccessfulTransmissions != 2 {
		t.Errorf("Expected 2 successful transmissions, got %d", info.SuccessfulTransmissions)
	}

	if info.TotalPlasticityEvents != 1 {
		t.Errorf("Expected 1 plasticity event, got %d", info.TotalPlasticityEvents)
	}

	// Validate calculated metrics
	expectedAverageDelay := 2 * time.Millisecond // (2+3+1)/3 = 2ms
	if info.AverageDelay != expectedAverageDelay {
		t.Errorf("Expected average delay %v, got %v",
			expectedAverageDelay, info.AverageDelay)
	}

	// Validate health and activity are included
	if info.HealthScore < 0.0 || info.HealthScore > 1.0 {
		t.Errorf("Invalid health score %.3f", info.HealthScore)
	}

	if info.ActivityLevel < 0.0 {
		t.Errorf("Invalid activity level %.3f", info.ActivityLevel)
	}

	t.Log("✓ Activity information retrieved successfully")
	t.Log("✓ Transmission statistics accurately calculated")
	t.Log("✓ Plasticity information included")
	t.Log("✓ Health and activity metrics provided")
}

// =================================================================================
// STATISTICS AND METRICS TESTS
// =================================================================================

// TestActivityMonitorBasicStatistics validates basic statistical calculations
// BIOLOGICAL CONTEXT: Astrocytic data analysis for synaptic performance assessment
func TestActivityMonitorBasicStatistics(t *testing.T) {
	t.Log("=== TESTING: Basic Statistics Calculation ===")
	t.Log("BIOLOGICAL MODEL: Astrocytic analysis of synaptic performance patterns")
	t.Log("EXPECTED: Accurate statistical metrics for biological assessment")

	monitor := NewSynapticActivityMonitor("stats_synapse")

	// Record varied transmission times for statistics
	processingTimes := []time.Duration{
		1 * time.Millisecond,
		2 * time.Millisecond,
		3 * time.Millisecond,
		4 * time.Millisecond,
		5 * time.Millisecond,
	}

	for i, procTime := range processingTimes {
		success := i%2 == 0 // Alternating success/failure
		monitor.RecordTransmission(success, success, procTime)
	}

	// Validate transmission statistics
	expectedTotal := int64(5)
	expectedSuccessful := int64(3) // Indices 0, 2, 4
	expectedFailed := int64(2)     // Indices 1, 3

	if monitor.transmissionCount != expectedTotal {
		t.Errorf("Expected %d total transmissions, got %d",
			expectedTotal, monitor.transmissionCount)
	}

	if monitor.successfulTransmissions != expectedSuccessful {
		t.Errorf("Expected %d successful transmissions, got %d",
			expectedSuccessful, monitor.successfulTransmissions)
	}

	if monitor.failedTransmissions != expectedFailed {
		t.Errorf("Expected %d failed transmissions, got %d",
			expectedFailed, monitor.failedTransmissions)
	}

	// Validate timing statistics
	expectedTotalLatency := 15 * time.Millisecond // 1+2+3+4+5 = 15ms
	if monitor.totalLatency != expectedTotalLatency {
		t.Errorf("Expected total latency %v, got %v",
			expectedTotalLatency, monitor.totalLatency)
	}

	// Validate reliability calculation
	expectedReliability := 0.6 // 3/5 = 60%
	if monitor.reliabilityScore != expectedReliability {
		t.Errorf("Expected reliability %.2f, got %.2f",
			expectedReliability, monitor.reliabilityScore)
	}

	t.Log("✓ Transmission statistics calculated accurately")
	t.Log("✓ Timing metrics properly accumulated")
	t.Log("✓ Reliability scores reflect actual performance")
	t.Log("✓ Statistical foundation for biological assessment established")
}
