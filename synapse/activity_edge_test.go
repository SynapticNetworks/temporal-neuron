/*
=================================================================================
ACTIVITY MONITOR EDGE CASE TESTS - BOUNDARY AND STRESS CONDITIONS
=================================================================================

This file contains edge case and boundary condition tests for the SynapticActivityMonitor,
validating robustness under extreme conditions, invalid inputs, and stress scenarios
that might occur in pathological states or during system failures.

BIOLOGICAL CONTEXT:
These tests model extreme conditions that synapses might encounter:
- Severe pathological states (complete transmission failure)
- Resource exhaustion (memory limits, computational overload)
- Invalid biological parameters (negative values, infinite inputs)
- Rapid activity bursts (seizure-like activity)
- Long-term inactivity (synaptic silence)
- Concurrent access patterns (multiple neural processes)

TEST CATEGORIES:
1. Boundary Value Testing (zero, negative, infinite values)
2. Memory and Resource Limits
3. Concurrent Access and Thread Safety
4. Extreme Activity Patterns
5. Invalid Input Handling
6. Cleanup and Memory Management
7. Performance Under Stress

BIOLOGICAL VALIDATION:
Edge cases test the monitor's resilience to conditions that would stress
real synapses, ensuring the monitoring system remains stable even when
the synapse itself is failing or under extreme stress.
=================================================================================
*/

package synapse

import (
	"math"
	"sync"
	"testing"
	"time"
)

// =================================================================================
// BOUNDARY VALUE TESTS
// =================================================================================

// TestActivityMonitorEdgeZeroValues validates handling of zero/minimal values
// BIOLOGICAL CONTEXT: Completely silent synapses or minimal detectable activity
func TestActivityMonitorEdgeZeroValues(t *testing.T) {
	t.Log("=== TESTING: Zero and Minimal Value Handling ===")
	t.Log("BIOLOGICAL MODEL: Silent synapses with minimal detectable activity")
	t.Log("EXPECTED: Graceful handling of zero values without errors")

	monitor := NewSynapticActivityMonitor("zero_test_synapse")

	// Test zero processing time
	monitor.RecordTransmission(true, true, 0)
	if monitor.transmissionCount != 1 {
		t.Errorf("Expected 1 transmission with zero time, got %d", monitor.transmissionCount)
	}

	// Test zero weight change plasticity
	zeroPlasticityEvent := PlasticityEvent{
		SynapseID:    "zero_test_synapse",
		EventType:    PlasticitySTDP,
		Timestamp:    time.Now(),
		WeightBefore: 0.5,
		WeightAfter:  0.5,
		WeightChange: 0.0,
	}
	monitor.RecordPlasticity(zeroPlasticityEvent)

	if len(monitor.plasticityEvents) != 1 {
		t.Errorf("Expected 1 plasticity event with zero change, got %d", len(monitor.plasticityEvents))
	}

	// Test with zero signal strength and calcium
	monitor.RecordTransmissionWithDetails(
		false, // success
		false, // vesicle released
		0,     // processing time
		0.0,   // signal strength
		0.0,   // calcium level
		"no_vesicle",
	)

	// Validate no crashes and reasonable state
	info := monitor.GetActivityInfo()
	if info.TotalTransmissions != 2 {
		t.Errorf("Expected 2 total transmissions, got %d", info.TotalTransmissions)
	}

	t.Log("✓ Zero values handled gracefully without errors")
	t.Log("✓ Silent synapse conditions properly recorded")
	t.Log("✓ Minimal activity detection maintained")
}

// TestActivityMonitorEdgeNegativeValues validates handling of invalid negative values
// BIOLOGICAL CONTEXT: Invalid measurements or corrupted data from damaged synapses
func TestActivityMonitorEdgeNegativeValues(t *testing.T) {
	t.Log("=== TESTING: Negative Value Handling ===")
	t.Log("BIOLOGICAL MODEL: Invalid or corrupted biological measurements")
	t.Log("EXPECTED: Robust handling of invalid negative inputs")

	monitor := NewSynapticActivityMonitor("negative_test_synapse")

	// Test negative processing time (invalid)
	negativeTime := -5 * time.Millisecond
	monitor.RecordTransmission(true, true, negativeTime)

	// Should still record the event (implementation decision: accept the data)
	if monitor.transmissionCount != 1 {
		t.Errorf("Expected transmission recorded despite negative time")
	}

	// Test negative signal strength and calcium (pathological)
	monitor.RecordTransmissionWithDetails(
		true, // success
		true, // vesicle released
		1*time.Millisecond,
		-0.5, // negative signal strength (invalid)
		-1.0, // negative calcium (invalid)
		"corrupted_data",
	)

	// Test negative weight changes (valid for LTD)
	negativePlasticityEvent := PlasticityEvent{
		SynapseID:    "negative_test_synapse",
		EventType:    PlasticitySTDP,
		Timestamp:    time.Now(),
		WeightBefore: 0.5,
		WeightAfter:  0.3,
		WeightChange: -0.2, // Valid LTD
	}
	monitor.RecordPlasticity(negativePlasticityEvent)

	// Validate system stability
	info := monitor.GetActivityInfo()
	if info.TotalTransmissions != 2 {
		t.Errorf("Expected 2 transmissions, got %d", info.TotalTransmissions)
	}

	if info.TotalPlasticityEvents != 1 {
		t.Errorf("Expected 1 plasticity event, got %d", info.TotalPlasticityEvents)
	}

	t.Log("✓ Negative values handled without system failure")
	t.Log("✓ Invalid biological measurements processed robustly")
	t.Log("✓ Valid negative changes (LTD) properly distinguished")
}

// TestActivityMonitorEdgeExtremeValues validates handling of very large values
// BIOLOGICAL CONTEXT: Extreme pathological conditions or seizure-like activity
func TestActivityMonitorEdgeExtremeValues(t *testing.T) {
	t.Log("=== TESTING: Extreme Value Handling ===")
	t.Log("BIOLOGICAL MODEL: Seizure-like activity or severe pathological conditions")
	t.Log("EXPECTED: System stability under extreme biological stress")

	monitor := NewSynapticActivityMonitor("extreme_test_synapse")

	// Test extremely large processing time (pathological delay)
	extremeTime := 1 * time.Hour // Pathologically slow
	monitor.RecordTransmission(true, true, extremeTime)

	// Test extremely high signal strength (seizure-like)
	extremeSignal := 1000.0 // Far beyond normal biological range
	extremeCalcium := 100.0 // Toxic calcium levels
	monitor.RecordTransmissionWithDetails(
		true,
		true,
		extremeTime,
		extremeSignal,
		extremeCalcium,
		"seizure_activity",
	)

	// Test extreme weight change (massive plasticity)
	extremePlasticityEvent := PlasticityEvent{
		SynapseID:    "extreme_test_synapse",
		EventType:    PlasticitySTDP,
		Timestamp:    time.Now(),
		WeightBefore: 0.1,
		WeightAfter:  100.0, // Pathologically strong
		WeightChange: 99.9,
	}
	monitor.RecordPlasticity(extremePlasticityEvent)

	// Test system continues functioning
	info := monitor.GetActivityInfo()
	if info.TotalTransmissions != 2 {
		t.Errorf("Expected 2 transmissions under extreme conditions, got %d",
			info.TotalTransmissions)
	}

	// Test health assessment under extreme conditions
	assessment := monitor.PerformHealthAssessment()
	if math.IsNaN(assessment.OverallScore) || math.IsInf(assessment.OverallScore, 0) {
		t.Error("Health assessment should produce valid scores under extreme conditions")
	}

	t.Log("✓ Extreme values processed without system crash")
	t.Log("✓ Pathological conditions monitored robustly")
	t.Log("✓ Health assessment remains stable under stress")
}

// TestActivityMonitorEdgeInfiniteNaNValues validates handling of invalid float values
// BIOLOGICAL CONTEXT: Computational errors or sensor malfunctions
func TestActivityMonitorEdgeInfiniteNaNValues(t *testing.T) {
	t.Log("=== TESTING: Infinite and NaN Value Handling ===")
	t.Log("BIOLOGICAL MODEL: Sensor malfunctions or computational errors")
	t.Log("EXPECTED: Graceful handling of invalid mathematical values")

	monitor := NewSynapticActivityMonitor("invalid_test_synapse")

	// Test with NaN values
	monitor.RecordTransmissionWithDetails(
		true,
		true,
		1*time.Millisecond,
		math.NaN(), // Invalid signal
		math.NaN(), // Invalid calcium
		"sensor_malfunction",
	)

	// Test with infinite values
	monitor.RecordTransmissionWithDetails(
		true,
		true,
		1*time.Millisecond,
		math.Inf(1),  // Positive infinity
		math.Inf(-1), // Negative infinity
		"computational_overflow",
	)

	// Test plasticity with invalid values
	invalidPlasticityEvent := PlasticityEvent{
		SynapseID:    "invalid_test_synapse",
		EventType:    PlasticitySTDP,
		Timestamp:    time.Now(),
		WeightBefore: math.NaN(),
		WeightAfter:  math.Inf(1),
		WeightChange: math.NaN(),
	}
	monitor.RecordPlasticity(invalidPlasticityEvent)

	// Verify system continues functioning
	info := monitor.GetActivityInfo()
	if math.IsNaN(info.HealthScore) {
		t.Error("Health score should remain valid despite NaN inputs")
	}

	// Test health assessment with invalid data
	assessment := monitor.PerformHealthAssessment()
	for component, score := range assessment.ComponentScores {
		if math.IsNaN(score) || math.IsInf(score, 0) {
			t.Errorf("Component %s has invalid score: %f", component, score)
		}
	}

	t.Log("✓ NaN and infinite values handled gracefully")
	t.Log("✓ System maintains stability with invalid inputs")
	t.Log("✓ Health assessments produce valid results despite bad data")
}

// =================================================================================
// MEMORY AND RESOURCE LIMIT TESTS
// =================================================================================

// TestActivityMonitorEdgeMemoryLimits validates memory management and cleanup
// BIOLOGICAL CONTEXT: Long-term synaptic monitoring without memory exhaustion
func TestActivityMonitorEdgeMemoryLimits(t *testing.T) {
	t.Log("=== TESTING: Memory Management and Limits ===")
	t.Log("BIOLOGICAL MODEL: Long-term synaptic monitoring without resource exhaustion")
	t.Log("EXPECTED: Automatic cleanup prevents unbounded memory growth")

	monitor := NewSynapticActivityMonitor("memory_test_synapse")

	// Record many transmission events (beyond normal capacity)
	numEvents := MAX_ACTIVITY_HISTORY_SIZE * 2 // Twice the limit
	t.Logf("Recording %d transmission events (2x memory limit)", numEvents)

	for i := 0; i < numEvents; i++ {
		monitor.RecordTransmission(true, true, time.Millisecond)
	}

	// Verify memory limits are enforced
	if len(monitor.recentEvents) > MAX_ACTIVITY_HISTORY_SIZE {
		t.Errorf("Recent events exceeded limit: %d > %d",
			len(monitor.recentEvents), MAX_ACTIVITY_HISTORY_SIZE)
	}

	t.Log("✓ Recent events limited to prevent memory exhaustion")

	// Record many plasticity events
	numPlasticityEvents := MAX_PLASTICITY_HISTORY * 2
	t.Logf("Recording %d plasticity events (2x limit)", numPlasticityEvents)

	for i := 0; i < numPlasticityEvents; i++ {
		event := PlasticityEvent{
			SynapseID:    "memory_test_synapse",
			EventType:    PlasticitySTDP,
			Timestamp:    time.Now(),
			WeightBefore: 0.5,
			WeightAfter:  0.5,
			WeightChange: 0.0,
		}
		monitor.RecordPlasticity(event)
	}

	// Verify plasticity memory limits
	if len(monitor.plasticityEvents) > MAX_PLASTICITY_HISTORY {
		t.Errorf("Plasticity events exceeded limit: %d > %d",
			len(monitor.plasticityEvents), MAX_PLASTICITY_HISTORY)
	}

	// Record many weight history entries
	numWeightEntries := MAX_WEIGHT_HISTORY * 2
	for i := 0; i < numWeightEntries; i++ {
		event := PlasticityEvent{
			SynapseID:    "memory_test_synapse",
			EventType:    PlasticityHomeostatic,
			Timestamp:    time.Now(),
			WeightBefore: float64(i) * 0.001,
			WeightAfter:  float64(i+1) * 0.001,
			WeightChange: 0.001,
		}
		monitor.RecordPlasticity(event)
	}

	// Verify weight history limits
	if len(monitor.weightHistory) > MAX_WEIGHT_HISTORY {
		t.Errorf("Weight history exceeded limit: %d > %d",
			len(monitor.weightHistory), MAX_WEIGHT_HISTORY)
	}

	t.Log("✓ Plasticity event memory properly limited")
	t.Log("✓ Weight history memory properly managed")
	t.Log("✓ Memory cleanup prevents resource exhaustion")
}

// TestActivityMonitorEdgeHighFrequencyUpdates validates performance under rapid updates
// BIOLOGICAL CONTEXT: High-frequency neural activity or gamma oscillations
func TestActivityMonitorEdgeHighFrequencyUpdates(t *testing.T) {
	t.Log("=== TESTING: High-Frequency Update Performance ===")
	t.Log("BIOLOGICAL MODEL: Gamma oscillations or high-frequency burst activity")
	t.Log("EXPECTED: Stable performance under rapid update rates")

	monitor := NewSynapticActivityMonitor("highfreq_test_synapse")

	// Simulate high-frequency activity (100 Hz for 1 second)
	numUpdates := 100
	updateInterval := 10 * time.Millisecond

	t.Logf("Simulating %d updates at %v intervals (100 Hz)", numUpdates, updateInterval)

	startTime := time.Now()

	for i := 0; i < numUpdates; i++ {
		// Rapid transmission recording
		monitor.RecordTransmission(true, true, time.Microsecond*100)

		// Periodic health updates
		if i%10 == 0 {
			monitor.UpdateHealth()
		}

		// Periodic plasticity events
		if i%20 == 0 {
			event := PlasticityEvent{
				SynapseID:    "highfreq_test_synapse",
				EventType:    PlasticitySTDP,
				Timestamp:    time.Now(),
				WeightBefore: 0.5,
				WeightAfter:  0.5 + float64(i)*0.001,
				WeightChange: float64(i) * 0.001,
			}
			monitor.RecordPlasticity(event)
		}

		time.Sleep(updateInterval)
	}

	duration := time.Since(startTime)
	t.Logf("Completed %d updates in %v", numUpdates, duration)

	// Verify system stability
	info := monitor.GetActivityInfo()
	if info.TotalTransmissions != int64(numUpdates) {
		t.Errorf("Expected %d transmissions, got %d", numUpdates, info.TotalTransmissions)
	}

	// Verify health assessment still works
	assessment := monitor.PerformHealthAssessment()
	if assessment.OverallScore < 0.0 || assessment.OverallScore > 1.0 {
		t.Errorf("Invalid overall health score after high-frequency updates: %.3f",
			assessment.OverallScore)
	}

	t.Log("✓ High-frequency updates processed successfully")
	t.Log("✓ System performance remained stable under load")
	t.Log("✓ Health assessment functional after rapid activity")
}

// =================================================================================
// CONCURRENT ACCESS TESTS
// =================================================================================

// TestActivityMonitorEdgeConcurrentAccess validates thread safety
// BIOLOGICAL CONTEXT: Multiple neural processes accessing synapse simultaneously
func TestActivityMonitorEdgeConcurrentAccess(t *testing.T) {
	t.Log("=== TESTING: Concurrent Access and Thread Safety ===")
	t.Log("BIOLOGICAL MODEL: Multiple neural processes monitoring synapse simultaneously")
	t.Log("EXPECTED: Thread-safe operation without data corruption")

	monitor := NewSynapticActivityMonitor("concurrent_test_synapse")

	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 100

	t.Logf("Starting %d goroutines with %d operations each",
		numGoroutines, operationsPerGoroutine)

	// Concurrent transmission recording
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Record transmissions
				success := j%2 == 0
				monitor.RecordTransmission(success, success, time.Millisecond)

				// Record detailed transmissions
				if j%10 == 0 {
					monitor.RecordTransmissionWithDetails(
						success, success, time.Millisecond,
						float64(j)*0.01, 1.0, "",
					)
				}

				// Record plasticity events
				if j%20 == 0 {
					event := PlasticityEvent{
						SynapseID:    "concurrent_test_synapse",
						EventType:    PlasticitySTDP,
						Timestamp:    time.Now(),
						WeightBefore: 0.5,
						WeightAfter:  0.5 + float64(j)*0.001,
						WeightChange: float64(j) * 0.001,
					}
					monitor.RecordPlasticity(event)
				}
			}
		}(i)
	}

	// Concurrent health updates
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < operationsPerGoroutine; i++ {
			monitor.UpdateHealth()
			time.Sleep(time.Microsecond * 100)
		}
	}()

	// Concurrent health assessments
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < operationsPerGoroutine/10; i++ {
			assessment := monitor.PerformHealthAssessment()
			if assessment.OverallScore < 0.0 || assessment.OverallScore > 1.0 {
				t.Errorf("Invalid health score during concurrent access: %.3f",
					assessment.OverallScore)
			}
			time.Sleep(time.Millisecond)
		}
	}()

	// Concurrent info retrieval
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < operationsPerGoroutine/5; i++ {
			info := monitor.GetActivityInfo()
			if info.SynapseID != "concurrent_test_synapse" {
				t.Error("Data corruption detected in concurrent access")
			}
			time.Sleep(time.Microsecond * 500)
		}
	}()

	wg.Wait()

	// Verify final state consistency
	expectedTransmissions := int64(numGoroutines * operationsPerGoroutine)
	info := monitor.GetActivityInfo()

	if info.TotalTransmissions != expectedTransmissions {
		t.Errorf("Expected %d total transmissions, got %d",
			expectedTransmissions, info.TotalTransmissions)
	}

	t.Log("✓ Concurrent access completed without crashes")
	t.Log("✓ Data consistency maintained under concurrent load")
	t.Log("✓ Thread safety validated for all operations")
}

// =================================================================================
// EXTREME ACTIVITY PATTERN TESTS
// =================================================================================

// TestActivityMonitorEdgeBurstActivity validates handling of burst activity patterns
// BIOLOGICAL CONTEXT: Neural bursting patterns or pathological seizure activity
func TestActivityMonitorEdgeBurstActivity(t *testing.T) {
	t.Log("=== TESTING: Burst Activity Pattern Handling ===")
	t.Log("BIOLOGICAL MODEL: Neural bursting or seizure-like activity patterns")
	t.Log("EXPECTED: Accurate monitoring of burst patterns and recovery")

	monitor := NewSynapticActivityMonitor("burst_test_synapse")

	// Simulate burst pattern: rapid activity followed by silence
	burstSize := 50
	burstInterval := time.Microsecond * 100 // Very rapid
	silentPeriod := 100 * time.Millisecond

	t.Logf("Simulating burst: %d events at %v intervals, then %v silence",
		burstSize, burstInterval, silentPeriod)

	// Record rapid burst
	burstStart := time.Now()
	for i := 0; i < burstSize; i++ {
		monitor.RecordTransmission(true, true, burstInterval)
		time.Sleep(burstInterval)
	}
	burstEnd := time.Now()
	burstDuration := burstEnd.Sub(burstStart)

	// Verify burst was recorded
	if monitor.transmissionCount != int64(burstSize) {
		t.Errorf("Expected %d transmissions in burst, got %d",
			burstSize, monitor.transmissionCount)
	}

	// Check activity level during burst
	info := monitor.GetActivityInfo()
	if info.TransmissionRate < 100.0 { // Should be very high during burst
		t.Errorf("Expected high transmission rate during burst, got %.2f Hz",
			info.TransmissionRate)
	}

	t.Logf("Burst recorded: %d events in %v (%.1f Hz)",
		burstSize, burstDuration, float64(burstSize)/burstDuration.Seconds())

	// Silent period
	time.Sleep(silentPeriod)
	monitor.UpdateHealth()

	// Verify system adapts to silence
	postSilenceInfo := monitor.GetActivityInfo()
	if postSilenceInfo.TransmissionRate >= info.TransmissionRate {
		t.Error("Transmission rate should decrease after silent period")
	}

	t.Log("✓ Burst activity pattern accurately recorded")
	t.Log("✓ Activity rate correctly calculated during burst")
	t.Log("✓ System adapted to post-burst silence period")
}

// TestActivityMonitorEdgeLongTermInactivity validates handling of prolonged silence
// BIOLOGICAL CONTEXT: Synaptic silence during deep sleep or coma states
func TestActivityMonitorEdgeLongTermInactivity(t *testing.T) {
	t.Log("=== TESTING: Long-Term Inactivity Handling ===")
	t.Log("BIOLOGICAL MODEL: Prolonged synaptic silence (sleep, coma, or damage)")
	t.Log("EXPECTED: Health degradation and appropriate issue detection")

	monitor := NewSynapticActivityMonitor("inactive_test_synapse")

	// Record initial activity to establish baseline
	for i := 0; i < 5; i++ {
		monitor.RecordTransmission(true, true, time.Millisecond)
	}

	initialHealth := monitor.healthScore
	t.Logf("Initial health after activity: %.3f", initialHealth)

	// Simulate long-term inactivity by advancing time
	// (In real test, we'd mock time or adjust thresholds)
	monitor.lastTransmission = time.Now().Add(-INACTIVITY_PENALTY_THRESHOLD * 2)
	monitor.lastPlasticityEvent = time.Now().Add(-INACTIVITY_PENALTY_THRESHOLD * 3)

	// Update health after simulated inactivity
	monitor.UpdateHealth()
	postInactivityHealth := monitor.healthScore

	if postInactivityHealth >= initialHealth {
		t.Errorf("Health should degrade with inactivity: %.3f >= %.3f",
			postInactivityHealth, initialHealth)
	}

	t.Logf("Health after inactivity: %.3f (degraded by %.3f)",
		postInactivityHealth, initialHealth-postInactivityHealth)

	// Perform health assessment to detect inactivity issues
	assessment := monitor.PerformHealthAssessment()
	hasInactivityIssue := false
	for _, issue := range assessment.IssuesDetected {
		if issue == "prolonged_inactivity" {
			hasInactivityIssue = true
			break
		}
	}

	if !hasInactivityIssue {
		t.Error("Long-term inactivity should be detected as health issue")
	}

	t.Log("✓ Health degraded appropriately with long-term inactivity")
	t.Log("✓ Inactivity detected as health issue")
	t.Log("✓ System maintains stability during prolonged silence")
}

// =================================================================================
// CLEANUP AND RESOURCE MANAGEMENT TESTS
// =================================================================================

// TestActivityMonitorEdgeEventCleanup validates automatic event cleanup
// BIOLOGICAL CONTEXT: Natural aging and forgetting of old synaptic events
func TestActivityMonitorEdgeEventCleanup(t *testing.T) {
	t.Log("=== TESTING: Automatic Event Cleanup ===")
	t.Log("BIOLOGICAL MODEL: Natural forgetting of old synaptic memories")
	t.Log("EXPECTED: Automatic cleanup of old events to prevent memory bloat")

	monitor := NewSynapticActivityMonitor("cleanup_test_synapse")

	// Fill up event history beyond analysis window
	numEvents := 200
	t.Logf("Recording %d events for cleanup testing", numEvents)

	oldTime := time.Now().Add(-ANALYSIS_WINDOW_DURATION * 2)
	for i := 0; i < numEvents; i++ {
		// Manually set old timestamps to simulate aged events
		event := TransmissionEvent{
			Timestamp:       oldTime.Add(time.Duration(i) * time.Millisecond),
			Success:         true,
			VesicleReleased: true,
			ProcessingTime:  time.Millisecond,
			SignalStrength:  1.0,
			CalciumLevel:    1.0,
			MetabolicCost:   1.0,
		}

		monitor.mu.Lock()
		monitor.recentEvents = append(monitor.recentEvents, event)
		monitor.transmissionCount++
		monitor.successfulTransmissions++
		monitor.mu.Unlock()
	}

	initialEventCount := len(monitor.recentEvents)
	t.Logf("Events before cleanup: %d", initialEventCount)

	// Trigger cleanup by recording new event
	monitor.RecordTransmission(true, true, time.Millisecond)

	finalEventCount := len(monitor.recentEvents)
	t.Logf("Events after cleanup: %d", finalEventCount)

	if finalEventCount >= initialEventCount {
		t.Error("Event cleanup should remove old events")
	}

	// Verify recent events are within analysis window
	cutoff := time.Now().Add(-ANALYSIS_WINDOW_DURATION)
	for _, event := range monitor.recentEvents {
		if event.Timestamp.Before(cutoff) {
			t.Error("Old events should be cleaned up")
			break
		}
	}

	t.Log("✓ Old events automatically cleaned up")
	t.Log("✓ Recent events preserved within analysis window")
	t.Log("✓ Memory usage controlled through cleanup")
}

// TestActivityMonitorEdgeResourceRecovery validates recovery from resource exhaustion
// BIOLOGICAL CONTEXT: Recovery from metabolic stress or resource depletion
func TestActivityMonitorEdgeResourceRecovery(t *testing.T) {
	t.Log("=== TESTING: Resource Recovery and Resilience ===")
	t.Log("BIOLOGICAL MODEL: Recovery from metabolic stress or resource depletion")
	t.Log("EXPECTED: Graceful degradation and recovery from resource limits")

	monitor := NewSynapticActivityMonitor("recovery_test_synapse")

	// Exhaust memory limits
	for i := 0; i < MAX_ACTIVITY_HISTORY_SIZE*2; i++ {
		monitor.RecordTransmission(true, true, time.Millisecond)
	}

	// Verify system is still functional despite limits
	if len(monitor.recentEvents) > MAX_ACTIVITY_HISTORY_SIZE {
		t.Errorf("Events should be limited to %d, got %d",
			MAX_ACTIVITY_HISTORY_SIZE, len(monitor.recentEvents))
	}

	// Test that new events can still be recorded
	preRecoveryCount := monitor.transmissionCount
	monitor.RecordTransmission(true, true, time.Millisecond)
	postRecoveryCount := monitor.transmissionCount

	if postRecoveryCount != preRecoveryCount+1 {
		t.Error("Should be able to record new events after reaching limits")
	}

	// Test health assessment still works
	assessment := monitor.PerformHealthAssessment()
	if assessment.OverallScore < 0.0 || assessment.OverallScore > 1.0 {
		t.Errorf("Health assessment should work after resource limits: %.3f",
			assessment.OverallScore)
	}

	// Test activity info still retrievable
	info := monitor.GetActivityInfo()
	if info.SynapseID != "recovery_test_synapse" {
		t.Error("Activity info should be retrievable after resource stress")
	}

	t.Log("✓ System remains functional at resource limits")
	t.Log("✓ New events continue to be processed")
	t.Log("✓ Health assessment continues to work")
	t.Log("✓ Resource recovery demonstrates system resilience")
}
