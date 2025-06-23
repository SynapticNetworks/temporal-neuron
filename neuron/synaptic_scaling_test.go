package neuron

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/message"
)

// ============================================================================
// BASIC FUNCTIONALITY TESTS
// ============================================================================

// TestSynapticScalingCreation validates basic scaling system initialization
//
// BIOLOGICAL CONTEXT:
// Every neuron must initialize its synaptic scaling system with appropriate
// default parameters that reflect biological homeostatic mechanisms.
//
// EXPECTED RESULTS:
// - System initializes with scaling disabled by default
// - Default parameters are within biological ranges
// - All data structures are properly initialized
// - Thread-safe access mechanisms are functional
func TestSynapticScaling_Creation(t *testing.T) {
	t.Log("=== TESTING Synaptic Scaling System Creation ===")

	scaling := NewSynapticScalingState()

	// Verify scaling is disabled by default (backward compatibility)
	if scaling.Config.Enabled {
		t.Error("Synaptic scaling should be disabled by default")
	}

	// Verify default parameters are reasonable
	if scaling.Config.TargetInputStrength != SYNAPTIC_SCALING_TARGET_STRENGTH_DEFAULT {
		t.Errorf("Expected default target strength %.1f, got %.1f",
			SYNAPTIC_SCALING_TARGET_STRENGTH_DEFAULT, scaling.Config.TargetInputStrength)
	}

	if scaling.Config.ScalingRate != SYNAPTIC_SCALING_RATE_DEFAULT {
		t.Errorf("Expected default scaling rate %.4f, got %.4f",
			SYNAPTIC_SCALING_RATE_DEFAULT, scaling.Config.ScalingRate)
	}

	if scaling.Config.ScalingInterval != SYNAPTIC_SCALING_INTERVAL_DEFAULT {
		t.Errorf("Expected default interval %v, got %v",
			SYNAPTIC_SCALING_INTERVAL_DEFAULT, scaling.Config.ScalingInterval)
	}

	// Verify data structures are initialized
	if scaling.InputGains == nil {
		t.Error("InputGains map should be initialized")
	}

	if scaling.InputActivityHistory == nil {
		t.Error("InputActivityHistory map should be initialized")
	}

	// Verify initial state
	gains := scaling.GetInputGains()
	if len(gains) != 0 {
		t.Errorf("Expected no initial gains, got %d", len(gains))
	}

	t.Log("✓ Synaptic scaling system created with correct defaults")
}

// TestSynapticScalingEnableDisable validates enable/disable functionality
//
// BIOLOGICAL CONTEXT:
// Synaptic scaling can be modulated by various factors including development,
// learning phases, and pathological conditions. The system must handle
// enable/disable transitions cleanly.
//
// EXPECTED RESULTS:
// - Enable/disable transitions work correctly
// - Parameters are properly set when enabling
// - Existing gains are preserved when disabling
// - State remains consistent across transitions
func TestSynapticScaling_EnableDisable(t *testing.T) {
	t.Log("=== TESTING Synaptic Scaling Enable/Disable ===")

	scaling := NewSynapticScalingState()

	// Test enabling scaling
	targetStrength := 1.5
	scalingRate := 0.005
	interval := 20 * time.Second

	scaling.EnableScaling(targetStrength, scalingRate, interval)

	if !scaling.Config.Enabled {
		t.Error("Scaling should be enabled after EnableScaling call")
	}

	if scaling.Config.TargetInputStrength != targetStrength {
		t.Errorf("Expected target strength %.1f, got %.1f",
			targetStrength, scaling.Config.TargetInputStrength)
	}

	if scaling.Config.ScalingRate != scalingRate {
		t.Errorf("Expected scaling rate %.4f, got %.4f",
			scalingRate, scaling.Config.ScalingRate)
	}

	if scaling.Config.ScalingInterval != interval {
		t.Errorf("Expected interval %v, got %v", interval, scaling.Config.ScalingInterval)
	}

	// Add some gains to test preservation
	scaling.SetInputGain("source1", 1.5)
	scaling.SetInputGain("source2", 0.8)

	// Test disabling scaling
	scaling.DisableScaling()

	if scaling.Config.Enabled {
		t.Error("Scaling should be disabled after DisableScaling call")
	}

	// Verify gains are preserved
	gains := scaling.GetInputGains()
	if len(gains) != 2 {
		t.Errorf("Expected 2 preserved gains, got %d", len(gains))
	}

	if gains["source1"] != 1.5 {
		t.Errorf("Expected preserved gain 1.5 for source1, got %.1f", gains["source1"])
	}

	if gains["source2"] != 0.8 {
		t.Errorf("Expected preserved gain 0.8 for source2, got %.1f", gains["source2"])
	}

	t.Log("✓ Enable/disable functionality working correctly")
}

// ============================================================================
// RECEPTOR GAIN MANAGEMENT TESTS
// ============================================================================

// TestPostSynapticGainApplication validates gain application to incoming signals
//
// BIOLOGICAL CONTEXT:
// Post-synaptic neurons control their own receptor sensitivity by adjusting
// AMPA/NMDA receptor density. This test validates that incoming signals are
// correctly modulated by these receptor gains.
//
// EXPECTED RESULTS:
// - Signals are multiplied by appropriate receptor gains
// - New sources are automatically registered with default gain
// - Disabled scaling passes signals unchanged
// - Thread-safe operation under concurrent access
func TestSynapticScaling_PostSynapticGainApplication(t *testing.T) {
	t.Log("=== TESTING Post-Synaptic Gain Application ===")

	scaling := NewSynapticScalingState()

	// Test with scaling disabled (should pass signals unchanged)
	t.Run("DisabledScaling", func(t *testing.T) {
		msg := message.NeuralSignal{
			Value:    2.5,
			SourceID: "test_source",
		}

		result := scaling.ApplyPostSynapticGain(msg)
		if result != msg.Value {
			t.Errorf("With scaling disabled, expected unchanged value %.1f, got %.1f",
				msg.Value, result)
		}

		t.Log("✓ Disabled scaling passes signals unchanged")
	})

	// Enable scaling for remaining tests
	scaling.EnableScaling(1.0, 0.001, 30*time.Second)

	// Test automatic registration with default gain
	t.Run("AutomaticRegistration", func(t *testing.T) {
		msg := message.NeuralSignal{
			Value:    1.8,
			SourceID: "new_source",
		}

		result := scaling.ApplyPostSynapticGain(msg)
		expectedResult := msg.Value * SYNAPTIC_SCALING_DEFAULT_GAIN

		if result != expectedResult {
			t.Errorf("Expected default gain application %.1f, got %.1f",
				expectedResult, result)
		}

		// Verify source was registered
		gains := scaling.GetInputGains()
		if gain, exists := gains["new_source"]; !exists {
			t.Error("New source should be automatically registered")
		} else if gain != SYNAPTIC_SCALING_DEFAULT_GAIN {
			t.Errorf("Expected default gain %.1f, got %.1f",
				SYNAPTIC_SCALING_DEFAULT_GAIN, gain)
		}

		t.Log("✓ Automatic source registration working")
	})

	// Test custom gain application
	t.Run("CustomGainApplication", func(t *testing.T) {
		scaling.SetInputGain("custom_source", 2.0)

		msg := message.NeuralSignal{
			Value:    1.5,
			SourceID: "custom_source",
		}

		result := scaling.ApplyPostSynapticGain(msg)
		expectedResult := msg.Value * 2.0

		if result != expectedResult {
			t.Errorf("Expected custom gain application %.1f, got %.1f",
				expectedResult, result)
		}

		t.Log("✓ Custom gain application working")
	})

	// Test gain bounds enforcement
	t.Run("GainBoundsEnforcement", func(t *testing.T) {
		// Test minimum bound
		scaling.SetInputGain("min_test", -0.5) // Below minimum
		gains := scaling.GetInputGains()
		if gains["min_test"] != SYNAPTIC_SCALING_MIN_GAIN {
			t.Errorf("Expected minimum gain %.2f, got %.2f",
				SYNAPTIC_SCALING_MIN_GAIN, gains["min_test"])
		}

		// Test maximum bound
		scaling.SetInputGain("max_test", 50.0) // Above maximum
		gains = scaling.GetInputGains()
		if gains["max_test"] != SYNAPTIC_SCALING_MAX_GAIN {
			t.Errorf("Expected maximum gain %.1f, got %.1f",
				SYNAPTIC_SCALING_MAX_GAIN, gains["max_test"])
		}

		t.Log("✓ Gain bounds enforcement working")
	})

	// Test empty source ID handling
	t.Run("EmptySourceID", func(t *testing.T) {
		msg := message.NeuralSignal{
			Value:    2.0,
			SourceID: "", // Empty source ID
		}

		result := scaling.ApplyPostSynapticGain(msg)
		if result != msg.Value {
			t.Errorf("Empty source ID should pass signal unchanged: expected %.1f, got %.1f",
				msg.Value, result)
		}

		t.Log("✓ Empty source ID handling working")
	})
}

// ============================================================================
// ACTIVITY TRACKING TESTS
// ============================================================================

// TestActivityTracking validates input activity monitoring and history management
//
// BIOLOGICAL CONTEXT:
// Neurons must monitor their synaptic activity patterns over time to make
// appropriate scaling decisions. This test validates the activity tracking
// system that provides the data for scaling calculations.
//
// EXPECTED RESULTS:
// - Activity is correctly recorded with timestamps
// - Old activity data is cleaned up appropriately
// - Average activity calculations are accurate
// - Thread-safe operation under concurrent activity recording
func TestSynapticScaling_ActivityTracking(t *testing.T) {
	t.Log("=== TESTING Activity Tracking System ===")

	scaling := NewSynapticScalingState()
	scaling.EnableScaling(1.0, 0.001, 30*time.Second)

	// Test basic activity recording
	t.Run("BasicActivityRecording", func(t *testing.T) {
		sourceID := "activity_source"
		activityValue := 1.5

		scaling.RecordInputActivity(sourceID, activityValue)

		// Check that activity was recorded
		strength := scaling.GetRecentActivityStrength(sourceID)
		if math.Abs(strength-activityValue) > SYNAPTIC_SCALING_TEST_TOLERANCE {
			t.Errorf("Expected recent activity strength %.1f, got %.1f",
				activityValue, strength)
		}

		t.Log("✓ Basic activity recording working")
	})

	// Test multiple activity records
	t.Run("MultipleActivityRecords", func(t *testing.T) {
		sourceID := "multi_source"
		activities := []float64{1.0, 1.5, 2.0, 1.2, 1.8}

		for _, activity := range activities {
			scaling.RecordInputActivity(sourceID, activity)
			time.Sleep(1 * time.Millisecond) // Small delay for distinct timestamps
		}

		// Calculate expected average
		var sum float64
		for _, activity := range activities {
			sum += activity
		}
		expectedAverage := sum / float64(len(activities))

		actualAverage := scaling.GetRecentActivityStrength(sourceID)
		if math.Abs(actualAverage-expectedAverage) > SYNAPTIC_SCALING_TEST_TOLERANCE {
			t.Errorf("Expected average activity %.2f, got %.2f",
				expectedAverage, actualAverage)
		}

		t.Log("✓ Multiple activity records averaging correctly")
	})

	// Test activity cleanup
	t.Run("ActivityCleanup", func(t *testing.T) {
		// Set very short activity window for testing
		scaling.Config.ActivitySamplingWindow = 10 * time.Millisecond

		sourceID := "cleanup_source"

		// Record old activity
		scaling.RecordInputActivity(sourceID, 1.0)

		// Wait for activity to become old
		time.Sleep(20 * time.Millisecond)

		// Record new activity
		scaling.RecordInputActivity(sourceID, 2.0)

		// Recent activity should only reflect the new value
		recentStrength := scaling.GetRecentActivityStrength(sourceID)
		if math.Abs(recentStrength-2.0) > SYNAPTIC_SCALING_TEST_TOLERANCE {
			t.Errorf("Expected only recent activity 2.0, got %.2f", recentStrength)
		}

		t.Log("✓ Activity cleanup working correctly")
	})

	// Test activity with disabled scaling
	t.Run("DisabledScalingActivity", func(t *testing.T) {
		scaling.DisableScaling()

		sourceID := "disabled_source"
		scaling.RecordInputActivity(sourceID, 1.0)

		// Should not record activity when disabled
		strength := scaling.GetRecentActivityStrength(sourceID)
		if strength != 0.0 {
			t.Errorf("Expected no activity recording when disabled, got %.2f", strength)
		}

		t.Log("✓ Disabled scaling prevents activity recording")
	})

	// Test empty source ID
	t.Run("EmptySourceActivityRecording", func(t *testing.T) {
		scaling.EnableScaling(1.0, 0.001, 30*time.Second)

		// Should not crash or record activity for empty source
		scaling.RecordInputActivity("", 1.0)

		strength := scaling.GetRecentActivityStrength("")
		if strength != 0.0 {
			t.Errorf("Expected no activity for empty source, got %.2f", strength)
		}

		t.Log("✓ Empty source ID handled correctly")
	})
}

// ============================================================================
// CORE SCALING ALGORITHM TESTS
// ============================================================================

// TestScalingDecisionLogic validates the core scaling algorithm
//
// BIOLOGICAL CONTEXT:
// The scaling decision process models the biological mechanisms that determine
// when and how much to adjust receptor sensitivity. This involves activity
// monitoring, calcium-dependent gating, and proportional adjustment calculations.
//
// EXPECTED RESULTS:
// - Scaling occurs only when conditions are met (activity, timing, significance)
// - Scaling factor calculations are biologically accurate
// - Safety bounds are respected
// - Pattern preservation through proportional scaling
func TestSynapticScaling_ScalingDecisionLogic(t *testing.T) {
	t.Log("=== TESTING Core Scaling Decision Logic ===")

	scaling := NewSynapticScalingState()

	// Test scaling disabled case
	t.Run("ScalingDisabled", func(t *testing.T) {
		result := scaling.PerformScaling(1.0, 5.0) // High calcium and firing rate

		if result.ScalingPerformed {
			t.Error("Scaling should not be performed when disabled")
		}

		if result.Reason != "scaling_disabled" {
			t.Errorf("Expected reason 'scaling_disabled', got '%s'", result.Reason)
		}

		t.Log("✓ Disabled scaling correctly prevents operation")
	})

	// Enable scaling for remaining tests
	scaling.EnableScaling(1.0, 0.01, 100*time.Millisecond) // Fast interval for testing

	// Test insufficient activity gating
	t.Run("InsufficientActivity", func(t *testing.T) {
		// Set recent update time to allow interval to pass
		scaling.Config.LastScalingUpdate = time.Now().Add(-200 * time.Millisecond)

		result := scaling.PerformScaling(0.05, 0.05) // Low calcium and firing rate

		if result.ScalingPerformed {
			t.Error("Scaling should not occur with insufficient activity")
		}

		if result.Reason != "insufficient_activity" {
			t.Errorf("Expected reason 'insufficient_activity', got '%s'", result.Reason)
		}

		t.Log("✓ Insufficient activity correctly prevents scaling")
	})

	// Test interval not reached
	t.Run("IntervalNotReached", func(t *testing.T) {
		// Set recent update time
		scaling.Config.LastScalingUpdate = time.Now()

		result := scaling.PerformScaling(1.0, 5.0) // Sufficient activity

		if result.ScalingPerformed {
			t.Error("Scaling should not occur before interval has passed")
		}

		if result.Reason != "interval_not_reached" {
			t.Errorf("Expected reason 'interval_not_reached', got '%s'", result.Reason)
		}

		t.Log("✓ Interval timing correctly enforced")
	})

	// Test no active sources
	t.Run("NoActiveSources", func(t *testing.T) {
		// Reset timing to allow scaling
		scaling.Config.LastScalingUpdate = time.Now().Add(-200 * time.Millisecond)

		result := scaling.PerformScaling(1.0, 5.0)

		if result.ScalingPerformed {
			t.Error("Scaling should not occur with no active sources")
		}

		if result.Reason != "no_active_sources" {
			t.Errorf("Expected reason 'no_active_sources', got '%s'", result.Reason)
		}

		t.Log("✓ No active sources correctly prevents scaling")
	})

	// Test within target range (no scaling needed)
	t.Run("WithinTargetRange", func(t *testing.T) {
		// Set up sources with activity near target
		setupScalingTestData(scaling, 1.0) // Activity at target

		// Reset timing
		scaling.Config.LastScalingUpdate = time.Now().Add(-200 * time.Millisecond)

		result := scaling.PerformScaling(1.0, 5.0)

		if result.ScalingPerformed {
			t.Error("Scaling should not occur when within target range")
		}

		if result.Reason != "within_target_range" {
			t.Errorf("Expected reason 'within_target_range', got '%s'", result.Reason)
		}

		t.Log("✓ Target range check correctly prevents unnecessary scaling")
	})

	// Test successful scaling (input too low)
	t.Run("SuccessfulScalingLow", func(t *testing.T) {
		// Set up sources with low activity
		setupScalingTestData(scaling, 0.5) // Activity below target

		// Reset timing
		scaling.Config.LastScalingUpdate = time.Now().Add(-200 * time.Millisecond)

		result := scaling.PerformScaling(1.0, 5.0)

		if !result.ScalingPerformed {
			t.Errorf("Scaling should be performed with low activity, reason: %s", result.Reason)
		}

		if result.ScalingFactor <= 1.0 {
			t.Errorf("Expected scaling factor > 1.0 for low activity, got %.3f",
				result.ScalingFactor)
		}

		if len(result.SourcesScaled) == 0 {
			t.Error("Expected some sources to be scaled")
		}

		t.Logf("✓ Successful upward scaling: factor=%.3f, sources=%d",
			result.ScalingFactor, len(result.SourcesScaled))
	})

	// Test successful scaling (input too high)
	t.Run("SuccessfulScalingHigh", func(t *testing.T) {
		// Set up sources with high activity
		setupScalingTestData(scaling, 2.0) // Activity above target

		// Reset timing
		scaling.Config.LastScalingUpdate = time.Now().Add(-200 * time.Millisecond)

		result := scaling.PerformScaling(1.0, 5.0)

		if !result.ScalingPerformed {
			t.Errorf("Scaling should be performed with high activity, reason: %s", result.Reason)
		}

		if result.ScalingFactor >= 1.0 {
			t.Errorf("Expected scaling factor < 1.0 for high activity, got %.3f",
				result.ScalingFactor)
		}

		t.Logf("✓ Successful downward scaling: factor=%.3f, sources=%d",
			result.ScalingFactor, len(result.SourcesScaled))
	})
}

// ============================================================================
// PATTERN PRESERVATION TESTS
// ============================================================================

// TestPatternPreservation validates that relative input ratios are maintained
//
// BIOLOGICAL CONTEXT:
// Synaptic scaling must preserve learned patterns from STDP and other plasticity
// mechanisms. This is achieved by scaling all synaptic gains proportionally,
// maintaining the relative strength ratios between different input sources.
//
// EXPECTED RESULTS:
// - All input sources are scaled by the same factor
// - Relative ratios between sources are preserved
// - Strong inputs remain relatively stronger than weak inputs after scaling
// - Pattern preservation works across multiple scaling events
func TestSynapticScaling_PatternPreservation(t *testing.T) {
	t.Log("=== TESTING Pattern Preservation During Scaling ===")

	scaling := NewSynapticScalingState()
	scaling.EnableScaling(1.0, 0.05, 50*time.Millisecond) // Fast scaling for testing

	// Set up different strength patterns
	sources := []struct {
		id       string
		gain     float64
		activity float64
	}{
		{"strong_source", 2.0, 1.5},
		{"medium_source", 1.0, 1.0},
		{"weak_source", 0.5, 0.8},
	}

	// Initialize gains and record activity
	for _, source := range sources {
		scaling.SetInputGain(source.id, source.gain)
		// Record multiple activity samples
		for i := 0; i < 10; i++ {
			scaling.RecordInputActivity(source.id, source.activity)
			time.Sleep(1 * time.Millisecond)
		}
	}

	// Get initial gain ratios
	initialGains := scaling.GetInputGains()
	initialRatioStrongMedium := initialGains["strong_source"] / initialGains["medium_source"]
	initialRatioMediumWeak := initialGains["medium_source"] / initialGains["weak_source"]

	t.Logf("Initial gains: strong=%.2f, medium=%.2f, weak=%.2f",
		initialGains["strong_source"], initialGains["medium_source"], initialGains["weak_source"])
	t.Logf("Initial ratios: strong/medium=%.2f, medium/weak=%.2f",
		initialRatioStrongMedium, initialRatioMediumWeak)

	// Perform scaling (average activity ~1.1, target 1.0, should scale down slightly)
	time.Sleep(60 * time.Millisecond) // Wait for interval
	// Adjust activity to be significantly outside the threshold for scaling to occur
	setupScalingTestData(scaling, 1.3) // Ensure activity is clearly above target

	result := scaling.PerformScaling(1.0, 5.0)

	if !result.ScalingPerformed {
		t.Errorf("Expected scaling to be performed, reason: %s", result.Reason)
	}

	// Get post-scaling gain ratios
	finalGains := scaling.GetInputGains()
	finalRatioStrongMedium := finalGains["strong_source"] / finalGains["medium_source"]
	finalRatioMediumWeak := finalGains["medium_source"] / finalGains["weak_source"]

	t.Logf("Final gains: strong=%.2f, medium=%.2f, weak=%.2f",
		finalGains["strong_source"], finalGains["medium_source"], finalGains["weak_source"])
	t.Logf("Final ratios: strong/medium=%.2f, medium/weak=%.2f",
		finalRatioStrongMedium, finalRatioMediumWeak)

	// Verify ratios are preserved (within tolerance)
	ratioTolerance := 0.01
	if math.Abs(finalRatioStrongMedium-initialRatioStrongMedium) > ratioTolerance {
		t.Errorf("Strong/medium ratio not preserved: initial=%.3f, final=%.3f",
			initialRatioStrongMedium, finalRatioStrongMedium)
	}

	if math.Abs(finalRatioMediumWeak-initialRatioMediumWeak) > ratioTolerance {
		t.Errorf("Medium/weak ratio not preserved: initial=%.3f, final=%.3f",
			initialRatioMediumWeak, finalRatioMediumWeak)
	}

	// Verify all gains changed by approximately the same factor
	strongFactor := finalGains["strong_source"] / initialGains["strong_source"]
	mediumFactor := finalGains["medium_source"] / initialGains["medium_source"]
	weakFactor := finalGains["weak_source"] / initialGains["weak_source"]

	factorTolerance := 0.01
	if math.Abs(strongFactor-result.ScalingFactor) > factorTolerance {
		t.Errorf("Strong source factor %.3f doesn't match expected %.3f",
			strongFactor, result.ScalingFactor)
	}

	if math.Abs(mediumFactor-result.ScalingFactor) > factorTolerance {
		t.Errorf("Medium source factor %.3f doesn't match expected %.3f",
			mediumFactor, result.ScalingFactor)
	}

	if math.Abs(weakFactor-result.ScalingFactor) > factorTolerance {
		t.Errorf("Weak source factor %.3f doesn't match expected %.3f",
			weakFactor, result.ScalingFactor)
	}

	t.Log("✓ Pattern preservation verified - relative ratios maintained")
	t.Logf("✓ Proportional scaling verified - all sources scaled by factor %.3f", result.ScalingFactor)
}

// ============================================================================
// CONCURRENT ACCESS TESTS
// ============================================================================

// TestConcurrentAccess validates thread safety under concurrent operation
//
// BIOLOGICAL CONTEXT:
// Real neurons receive thousands of concurrent synaptic inputs while
// simultaneously performing scaling operations. The system must handle
// this concurrency without data races or corruption.
//
// EXPECTED RESULTS:
// - No data races under concurrent gain application and scaling
// - Activity recording remains consistent under concurrent access
// - Scaling decisions are not corrupted by concurrent modifications
// - All operations complete successfully without deadlocks
func TestSynapticScaling_ConcurrentAccess(t *testing.T) {
	t.Log("=== TESTING Concurrent Access Thread Safety ===")

	scaling := NewSynapticScalingState()
	scaling.EnableScaling(1.0, 0.01, 10*time.Millisecond)

	var wg sync.WaitGroup
	numGoroutines := 20
	operationsPerGoroutine := 100

	// Test concurrent gain application and activity recording
	t.Run("ConcurrentGainApplicationAndActivity", func(t *testing.T) {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < operationsPerGoroutine; j++ {
					sourceID := fmt.Sprintf("source_%d", goroutineID)

					// Apply gain to a signal
					msg := message.NeuralSignal{
						Value:    float64(j) * 0.01,
						SourceID: sourceID,
					}

					result := scaling.ApplyPostSynapticGain(msg)

					// Record activity
					scaling.RecordInputActivity(sourceID, result)

					// Occasionally set custom gain
					if j%10 == 0 {
						scaling.SetInputGain(sourceID, 1.0+float64(j)*0.1)
					}
				}
			}(i)
		}

		wg.Wait()

		// Verify system is still functional
		gains := scaling.GetInputGains()
		if len(gains) == 0 {
			t.Error("Expected some gains to be registered after concurrent operations")
		}

		t.Logf("✓ Concurrent gain application completed - %d sources registered", len(gains))
	})

	// Test concurrent scaling operations
	t.Run("ConcurrentScalingOperations", func(t *testing.T) {
		// Set up some baseline activity
		for i := 0; i < 5; i++ {
			sourceID := fmt.Sprintf("baseline_source_%d", i)
			scaling.SetInputGain(sourceID, 1.0)
			for j := 0; j < 10; j++ {
				scaling.RecordInputActivity(sourceID, 1.0)
			}
		}

		// Run concurrent scaling attempts
		successfulScalings := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				// Wait a bit to spread out the scaling attempts
				time.Sleep(time.Duration(goroutineID) * time.Millisecond)

				result := scaling.PerformScaling(1.0, 5.0)
				successfulScalings <- result.ScalingPerformed
			}(i)
		}

		wg.Wait()
		close(successfulScalings)

		// Count successful scalings (should be limited by interval timing)
		var successCount int
		for success := range successfulScalings {
			if success {
				successCount++
			}
		}

		// Should have some successful scalings but not all (due to interval limits)
		if successCount == 0 {
			t.Error("Expected at least some scaling operations to succeed")
		}

		if successCount == numGoroutines {
			t.Error("Expected interval timing to limit concurrent scalings")
		}

		t.Logf("✓ Concurrent scaling operations: %d/%d succeeded (interval limiting working)",
			successCount, numGoroutines)
	})

	// Test concurrent monitoring operations
	t.Run("ConcurrentMonitoring", func(t *testing.T) {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < operationsPerGoroutine; j++ {
					// Concurrent read operations
					_ = scaling.GetInputGains()
					_ = scaling.GetScalingHistory()
					_ = scaling.GetScalingStatus()
					_ = scaling.GetInputActivitySummary()

					sourceID := fmt.Sprintf("monitor_source_%d", goroutineID)
					_ = scaling.GetRecentActivityStrength(sourceID)
				}
			}(i)
		}

		wg.Wait()

		t.Log("✓ Concurrent monitoring operations completed without deadlock")
	})
}

// ============================================================================
// BIOLOGICAL REALISM TESTS
// ============================================================================

// TestBiologicalRealism validates biological accuracy of scaling parameters
//
// BIOLOGICAL CONTEXT:
// Synaptic scaling parameters must reflect realistic biological constraints
// including receptor synthesis rates, trafficking timescales, and metabolic costs.
//
// EXPECTED RESULTS:
// - Scaling timescales match biological homeostatic mechanisms
// - Scaling factors respect biological receptor trafficking limits
// - Activity thresholds reflect calcium-dependent signaling requirements
// - Safety bounds prevent pathological scaling behaviors
func TestSynapticScaling_BiologicalRealism(t *testing.T) {
	t.Log("=== TESTING Biological Realism of Scaling Parameters ===")

	// Test biologically realistic timescales
	t.Run("BiologicalTimescales", func(t *testing.T) {
		scaling := NewSynapticScalingState()

		// Test developmental timescales (faster)
		scaling.EnableScaling(1.0, SYNAPTIC_SCALING_RATE_DEVELOPMENTAL,
			SYNAPTIC_SCALING_INTERVAL_DEVELOPMENTAL)

		if scaling.Config.ScalingRate != SYNAPTIC_SCALING_RATE_DEVELOPMENTAL {
			t.Error("Developmental scaling rate not set correctly")
		}

		// Test mature timescales (slower)
		scaling.EnableScaling(1.0, SYNAPTIC_SCALING_RATE_CONSERVATIVE,
			SYNAPTIC_SCALING_INTERVAL_SLOW)

		if scaling.Config.ScalingInterval != SYNAPTIC_SCALING_INTERVAL_SLOW {
			t.Error("Mature scaling interval not set correctly")
		}

		t.Log("✓ Biologically realistic timescales working")
	})

	// Test receptor trafficking limits
	t.Run("ReceptorTraffickingLimits", func(t *testing.T) {
		scaling := NewSynapticScalingState()
		scaling.EnableScaling(1.0, 0.1, 10*time.Millisecond) // Aggressive settings

		// Set up extreme activity imbalance
		setupScalingTestData(scaling, 10.0) // Very high activity

		// Attempt scaling
		result := scaling.PerformScaling(1.0, 5.0)

		if !result.ScalingPerformed {
			t.Error("Expected scaling with extreme imbalance")
		}

		// Verify scaling factor is bounded
		if result.ScalingFactor < SYNAPTIC_SCALING_MIN_FACTOR {
			t.Errorf("Scaling factor %.3f below minimum %.3f",
				result.ScalingFactor, SYNAPTIC_SCALING_MIN_FACTOR)
		}

		if result.ScalingFactor > SYNAPTIC_SCALING_MAX_FACTOR {
			t.Errorf("Scaling factor %.3f above maximum %.3f",
				result.ScalingFactor, SYNAPTIC_SCALING_MAX_FACTOR)
		}

		t.Log("✓ Receptor trafficking limits enforced")
	})

	// Test calcium-dependent gating
	t.Run("CalciumDependentGating", func(t *testing.T) {
		scaling := NewSynapticScalingState()
		scaling.EnableScaling(1.0, 0.01, 10*time.Millisecond)

		setupScalingTestData(scaling, 0.5) // Low activity needing upscaling

		// Test with insufficient calcium
		// Ensure interval has passed for the first call not to be "interval_not_reached"
		scaling.Config.LastScalingUpdate = time.Now().Add(-20 * time.Millisecond)
		result := scaling.PerformScaling(0.05, 5.0) // Low calcium, high firing rate
		if result.ScalingPerformed {
			t.Error("Scaling should not occur with insufficient calcium")
		}
		if result.Reason != "insufficient_activity" {
			t.Errorf("Expected reason 'insufficient_activity', got '%s'", result.Reason)
		}

		// Test with sufficient calcium
		// Ensure interval has passed for this call too
		scaling.Config.LastScalingUpdate = time.Now().Add(-20 * time.Millisecond)
		result = scaling.PerformScaling(1.0, 5.0) // High calcium and firing rate
		if !result.ScalingPerformed {
			t.Errorf("Scaling should occur with sufficient calcium, reason: %s", result.Reason)
		}

		t.Log("✓ Calcium-dependent gating working")
	})

	// Test activity significance thresholds
	t.Run("ActivitySignificanceThresholds", func(t *testing.T) {
		scaling := NewSynapticScalingState()
		scaling.EnableScaling(1.0, 0.01, 10*time.Millisecond)

		// Test with activity very close to target (should not scale)
		setupScalingTestData(scaling, 1.05)                                       // 5% above target
		scaling.Config.LastScalingUpdate = time.Now().Add(-20 * time.Millisecond) // Ensure interval passed
		result := scaling.PerformScaling(1.0, 5.0)

		if result.ScalingPerformed {
			t.Error("Should not scale for small deviations from target")
		}

		// Test with significant deviation (should scale)
		setupScalingTestData(scaling, 1.3)                                        // 30% above target
		time.Sleep(15 * time.Millisecond)                                         // Wait for interval
		scaling.Config.LastScalingUpdate = time.Now().Add(-20 * time.Millisecond) // Ensure interval passed for this one too
		result = scaling.PerformScaling(1.0, 5.0)

		if !result.ScalingPerformed {
			t.Errorf("Should scale for significant deviations, reason: %s", result.Reason)
		}

		t.Log("✓ Activity significance thresholds working")
	})
}

// ============================================================================
// INTEGRATION TESTS
// ============================================================================

// TestScalingIntegration validates integration with neuron message processing
//
// BIOLOGICAL CONTEXT:
// Synaptic scaling must work seamlessly with ongoing neural computation,
// integrating with signal processing, activity monitoring, and homeostatic
// regulation without interfering with normal neural function.
//
// EXPECTED RESULTS:
// - Scaling integrates cleanly with message processing workflow
// - Activity tracking works with real message.NeuralSignal processing
// - Scaling decisions are based on actual neural activity patterns
// - System maintains stability under realistic operating conditions
func TestSynapticScaling_ScalingIntegration(t *testing.T) {
	t.Log("=== TESTING Synaptic Scaling Integration ===")

	scaling := NewSynapticScalingState()
	scaling.EnableScaling(1.0, 0.02, 50*time.Millisecond)

	// Simulate realistic message processing workflow
	t.Run("MessageProcessingWorkflow", func(t *testing.T) {
		sources := []string{"cortex_input", "thalamus_input", "feedback_input"}

		// Simulate multiple rounds of message processing
		for round := 0; round < 10; round++ {
			for _, sourceID := range sources {
				// Create realistic neural signal
				msg := message.NeuralSignal{
					Value:                1.5 + 0.5*rand.Float64(), // Variable signal strength
					Timestamp:            time.Now(),
					SourceID:             sourceID,
					NeurotransmitterType: message.LigandGlutamate,
				}

				// Apply gain (as would happen in neuron processing)
				processedValue := scaling.ApplyPostSynapticGain(msg)

				// Record activity (as would happen in neuron processing)
				scaling.RecordInputActivity(sourceID, processedValue)
			}

			// Occasionally perform scaling (as would happen in neuron run loop)
			if round%3 == 0 {
				// Ensure enough time has passed for scaling to occur
				scaling.Config.LastScalingUpdate = time.Now().Add(-60 * time.Millisecond)
				result := scaling.PerformScaling(0.8+0.4*rand.Float64(), 3.0+2.0*rand.Float64())
				if result.ScalingPerformed {
					t.Logf("Round %d: Scaling performed with factor %.3f", round, result.ScalingFactor)
				}
			}

			time.Sleep(10 * time.Millisecond)
		}

		// Verify system is stable and functional
		gains := scaling.GetInputGains()
		if len(gains) != len(sources) {
			t.Errorf("Expected %d sources registered, got %d", len(sources), len(gains))
		}

		status := scaling.GetScalingStatus()
		avgStrength := status["current_avg_strength"].(float64)
		target := status["target_strength"].(float64)

		t.Logf("Final average strength: %.3f, target: %.3f", avgStrength, target)
		t.Log("✓ Message processing workflow integration successful")
	})

	// Test homeostatic convergence
	t.Run("HomeostaticConvergence", func(t *testing.T) {
		scaling := NewSynapticScalingState()
		scaling.EnableScaling(1.2, 0.05, 20*time.Millisecond) // Target above current

		// Set up initial imbalance
		sourceID := "convergence_test"
		scaling.SetInputGain(sourceID, 0.5) // Low initial gain

		// Simulate sustained activity below target
		for i := 0; i < 20; i++ {
			scaling.RecordInputActivity(sourceID, 0.8) // Below target of 1.2

			if i%3 == 0 {
				// Ensure enough time has passed for scaling to occur
				scaling.Config.LastScalingUpdate = time.Now().Add(-25 * time.Millisecond)
				result := scaling.PerformScaling(1.0, 4.0)
				if result.ScalingPerformed {
					t.Logf("Convergence step %d: factor %.3f, gain %.3f",
						i/3, result.ScalingFactor, scaling.GetInputGains()[sourceID])
				}
			}

			time.Sleep(25 * time.Millisecond)
		}

		// Check if system converged toward target
		finalGain := scaling.GetInputGains()[sourceID]
		if finalGain <= 0.5 {
			t.Error("Expected gain to increase toward target, but it remained low")
		}

		t.Logf("✓ Homeostatic convergence: initial gain 0.5 → final gain %.3f", finalGain)
	})
}

// ============================================================================
// HELPER FUNCTIONS FOR TESTS
// ============================================================================

// setupScalingTestData sets up test data with specified average activity level
func setupScalingTestData(scaling *SynapticScalingState, targetActivity float64) {
	sources := []string{"test_source_1", "test_source_2", "test_source_3"}

	for i, sourceID := range sources {
		// Set different initial gains to test proportional scaling
		initialGain := 0.8 + 0.4*float64(i) // 0.8, 1.2, 1.6
		scaling.SetInputGain(sourceID, initialGain)

		// Record activity at the specified level (with some variation)
		for j := 0; j < 15; j++ {
			activity := targetActivity * (0.9 + 0.2*rand.Float64()) // ±10% variation
			scaling.RecordInputActivity(sourceID, activity)
			time.Sleep(time.Millisecond)
		}
	}
}

// ============================================================================
// BENCHMARK TESTS
// ============================================================================

// BenchmarkGainApplication benchmarks the performance of gain application
func BenchmarkGainApplication(b *testing.B) {
	scaling := NewSynapticScalingState()
	scaling.EnableScaling(1.0, 0.001, 30*time.Second)

	msg := message.NeuralSignal{
		Value:    1.5,
		SourceID: "benchmark_source",
	}

	// Prime the system
	scaling.ApplyPostSynapticGain(msg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scaling.ApplyPostSynapticGain(msg)
	}
}

// BenchmarkActivityRecording benchmarks activity recording performance
func BenchmarkActivityRecording(b *testing.B) {
	scaling := NewSynapticScalingState()
	scaling.EnableScaling(1.0, 0.001, 30*time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scaling.RecordInputActivity("benchmark_source", 1.5)
	}
}

// BenchmarkScalingOperation benchmarks the full scaling operation
func BenchmarkScalingOperation(b *testing.B) {
	scaling := NewSynapticScalingState()
	scaling.EnableScaling(1.0, 0.01, 1*time.Millisecond) // Fast for benchmarking

	// Set up test data
	setupScalingTestData(scaling, 1.5) // Above target

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scaling.PerformScaling(1.0, 5.0)
		// Reset timing to allow scaling each iteration
		scaling.Config.LastScalingUpdate = time.Now().Add(-2 * time.Millisecond)
	}
}

/*
=================================================================================
TEST SUITE SUMMARY - SYNAPTIC SCALING VALIDATION
=================================================================================

This comprehensive test suite validates all aspects of the synaptic scaling
system, ensuring biological accuracy, thread safety, and integration with
the broader neural computation framework.

COVERAGE AREAS:
1. **Basic Functionality**: Creation, enable/disable, parameter management
2. **Receptor Gain Management**: Gain application, bounds enforcement, registration
3. **Activity Tracking**: Recording, cleanup, average calculations
4. **Core Algorithm**: Scaling decisions, timing, significance testing
5. **Pattern Preservation**: Proportional scaling, ratio maintenance
6. **Concurrency**: Thread safety under concurrent access
7. **Biological Realism**: Realistic parameters, constraints, timescales
8. **Integration**: Seamless operation with neural message processing

BIOLOGICAL VALIDATION:
- Timescales match biological homeostatic mechanisms (seconds to minutes)
- Scaling factors respect receptor trafficking limits (5-20% per event)
- Activity thresholds reflect calcium-dependent signaling requirements
- Pattern preservation maintains learned connectivity patterns
- Safety bounds prevent pathological scaling behaviors

PERFORMANCE VALIDATION:
- Efficient gain application suitable for high-frequency neural processing
- Fast activity recording that doesn't impede neural computation
- Reasonable scaling operation performance for homeostatic timescales
- Thread-safe operation under realistic concurrent loads

This test suite ensures the synaptic scaling system provides robust,
biologically accurate homeostatic regulation while maintaining the
computational performance required for real-time neural simulation.

=================================================================================
*/
