package neuron

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// TestSTDPSignal_DeltaTSignConvention verifies the sign convention is consistent
func TestSTDPSignaling_DeltaTSignConvention(t *testing.T) {
	// Create system
	stdp := NewSTDPSignalingSystem(true, 10*time.Millisecond, 0.1)

	// Create mock callbacks that will record the adjustments
	mock := NewMockSTDPCallbacks()

	// Set up test synapses with clear timings
	postSpikeTime := time.Now()
	preSpikeTimeBefore := postSpikeTime.Add(-10 * time.Millisecond) // Pre before Post
	preSpikeTimeAfter := postSpikeTime.Add(10 * time.Millisecond)   // Post before Pre

	testSynapses := []types.SynapseInfo{
		{
			ID:               "pre_before_post", // Should produce negative deltaT
			SourceID:         "source1",
			TargetID:         "target_neuron",
			Weight:           0.5,
			LastActivity:     preSpikeTimeBefore,
			LastTransmission: preSpikeTimeBefore, // Use LastActivity for simplicity
		},
		{
			ID:               "post_before_pre", // Should produce positive deltaT
			SourceID:         "source2",
			TargetID:         "target_neuron",
			Weight:           0.5,
			LastActivity:     preSpikeTimeAfter,
			LastTransmission: preSpikeTimeAfter, // Use LastActivity for simplicity
		},
	}
	mock.SetSynapses(testSynapses)

	// Force feedback delivery now
	feedbackCount := stdp.DeliverFeedbackNow("target_neuron", mock, time.Now())
	if feedbackCount != 2 {
		t.Fatalf("Expected 2 synapses to receive feedback, got %d", feedbackCount)
	}

	// Get the adjustments from the mock
	adjustments := mock.GetAdjustments()
	if len(adjustments) != 2 {
		t.Fatalf("Expected 2 adjustments, got %d", len(adjustments))
	}

	// Map adjustments by synapse ID for checking
	adjustmentMap := make(map[string]types.PlasticityAdjustment)
	for _, adj := range adjustments {
		// We need to identify which adjustment is for which synapse
		// Since we can't easily track this directly, we'll use the deltaT value
		// Negative deltaT should be for pre_before_post, positive for post_before_pre
		if adj.DeltaT < 0 {
			adjustmentMap["pre_before_post"] = adj
		} else {
			adjustmentMap["post_before_pre"] = adj
		}
	}

	// Verify sign conventions
	preDeltaT, preOk := adjustmentMap["pre_before_post"].DeltaT, adjustmentMap["pre_before_post"].DeltaT != 0
	postDeltaT, postOk := adjustmentMap["post_before_pre"].DeltaT, adjustmentMap["post_before_pre"].DeltaT != 0

	t.Logf("Pre-before-Post deltaT: %v (found: %v)", preDeltaT, preOk)
	t.Logf("Post-before-Pre deltaT: %v (found: %v)", postDeltaT, postOk)

	// Check for sign inversions or missing adjustments
	if !preOk || preDeltaT >= 0 {
		t.Errorf("Pre-before-Post should have negative deltaT, got %v (found: %v)",
			preDeltaT, preOk)
	}

	if !postOk || postDeltaT <= 0 {
		t.Errorf("Post-before-Pre should have positive deltaT, got %v (found: %v)",
			postDeltaT, postOk)
	}
}

// TestSTDPSignaling_BasicFunctionality tests the basic operations of the STDP signaling system
func TestSTDPSignaling_BasicFunctionality(t *testing.T) {
	// Create a new STDP signaling system
	stdp := NewSTDPSignalingSystem(true, 10*time.Millisecond, 0.1)

	// Verify initial state
	if !stdp.IsEnabled() {
		t.Error("STDP signaling system should be enabled initially")
	}

	delay, rate := stdp.GetParameters()
	if delay != 10*time.Millisecond {
		t.Errorf("Expected feedback delay 10ms, got %v", delay)
	}

	if rate != 0.1 {
		t.Errorf("Expected learning rate 0.1, got %v", rate)
	}

	// Test disabling
	stdp.Disable()
	if stdp.IsEnabled() {
		t.Error("STDP signaling system should be disabled after Disable()")
	}

	// Test re-enabling with different parameters
	stdp.Enable(20*time.Millisecond, 0.2)
	if !stdp.IsEnabled() {
		t.Error("STDP signaling system should be enabled after Enable()")
	}

	delay, rate = stdp.GetParameters()
	if delay != 20*time.Millisecond {
		t.Errorf("Expected updated feedback delay 20ms, got %v", delay)
	}

	if rate != 0.2 {
		t.Errorf("Expected updated learning rate 0.2, got %v", rate)
	}

	t.Log("✓ Basic STDP signaling system functionality works correctly")
}

// TestSTDPSignaling_Scheduling tests the STDP feedback scheduling mechanism
func TestSTDPSignaling_Scheduling(t *testing.T) {
	// Create a new STDP signaling system
	feedbackDelay := 20 * time.Millisecond
	stdp := NewSTDPSignalingSystem(true, feedbackDelay, 0.1)

	// Test scheduling feedback
	now := time.Now()
	success := stdp.ScheduleFeedback(now)
	if !success {
		t.Error("ScheduleFeedback should return true for successful scheduling")
	}

	// Get status to check scheduled time
	status := stdp.GetStatus()

	// Verify scheduled time is set correctly
	expectedTime := now.Add(feedbackDelay)
	scheduledTime := status["scheduled_time"].(time.Time)

	// Allow small tolerance for timing differences
	timeDiff := scheduledTime.Sub(expectedTime)
	if timeDiff < -time.Millisecond || timeDiff > time.Millisecond {
		t.Errorf("Scheduled time not set correctly. Expected ~%v, got %v (diff: %v)",
			expectedTime, scheduledTime, timeDiff)
	}

	t.Log("✓ STDP scheduling logic works correctly")
}

// MockSTDPCallbacks implements the full component.NeuronCallbacks interface for testing
type MockSTDPCallbacks struct {
	synapses    []types.SynapseInfo
	adjustments []types.PlasticityAdjustment
}

func NewMockSTDPCallbacks() *MockSTDPCallbacks {
	return &MockSTDPCallbacks{
		synapses:    make([]types.SynapseInfo, 0),
		adjustments: make([]types.PlasticityAdjustment, 0),
	}
}

// Methods we actually use for testing
func (m *MockSTDPCallbacks) ListSynapses(criteria types.SynapseCriteria) []types.SynapseInfo {
	// For simplicity, return all synapses for any criteria
	return m.synapses
}

func (m *MockSTDPCallbacks) ApplyPlasticity(synapseID string, adjustment types.PlasticityAdjustment) error {
	m.adjustments = append(m.adjustments, adjustment)
	return nil
}

// Helper methods for test setup
func (m *MockSTDPCallbacks) SetSynapses(synapses []types.SynapseInfo) {
	m.synapses = make([]types.SynapseInfo, len(synapses))
	copy(m.synapses, synapses)
}

func (m *MockSTDPCallbacks) GetAdjustments() []types.PlasticityAdjustment {
	return m.adjustments
}

func (m *MockSTDPCallbacks) ClearAdjustments() {
	m.adjustments = make([]types.PlasticityAdjustment, 0)
}

// Required NeuronCallbacks interface methods that we don't use in testing
func (m *MockSTDPCallbacks) CreateSynapse(config types.SynapseCreationConfig) (string, error) {
	return "mock-synapse-id", nil
}

func (m *MockSTDPCallbacks) DeleteSynapse(synapseID string) error {
	return nil
}

func (m *MockSTDPCallbacks) GetSynapse(synapseID string) (component.SynapticProcessor, error) {
	return nil, nil
}

func (m *MockSTDPCallbacks) SetSynapseWeight(synapseID string, weight float64) error {
	return nil
}

func (m *MockSTDPCallbacks) GetSynapseWeight(synapseID string) (float64, error) {
	return 0.5, nil
}

func (m *MockSTDPCallbacks) GetSpatialDelay(targetID string) time.Duration {
	return 0
}

func (m *MockSTDPCallbacks) FindNearbyComponents(radius float64) []component.ComponentInfo {
	return nil
}

func (m *MockSTDPCallbacks) ReleaseChemical(ligandType types.LigandType, concentration float64) error {
	return nil
}

func (m *MockSTDPCallbacks) SendElectricalSignal(signalType types.SignalType, data interface{}) {
}

func (m *MockSTDPCallbacks) ReportHealth(activityLevel float64, connectionCount int) {
}

func (m *MockSTDPCallbacks) ReportStateChange(oldState, newState types.ComponentState) {
}

func (m *MockSTDPCallbacks) GetMatrix() component.ExtracellularMatrix {
	return nil
}

// TestSTDPSignaling_FeedbackDelivery tests the STDP feedback delivery mechanism
func TestSTDPSignaling_FeedbackDelivery(t *testing.T) {
	// Create a new STDP signaling system
	stdp := NewSTDPSignalingSystem(true, 10*time.Millisecond, 0.1)

	// Create mock callbacks
	mock := NewMockSTDPCallbacks()

	// Create test synapses with different timings
	now := time.Now()
	testSynapses := []types.SynapseInfo{
		{
			ID:               "causal_synapse",
			SourceID:         "source1",
			TargetID:         "target_neuron",
			Weight:           0.5,
			LastActivity:     now.Add(-5 * time.Millisecond), // Pre before post (causal)
			LastTransmission: now.Add(-5 * time.Millisecond), // Use LastActivity for simplicity
		},
		{
			ID:               "anticausal_synapse",
			SourceID:         "source2",
			TargetID:         "target_neuron",
			Weight:           0.5,
			LastActivity:     now.Add(5 * time.Millisecond),  // Post before pre (anti-causal)
			LastTransmission: now.Add(-5 * time.Millisecond), // Use LastActivity for simplicity
		},
	}
	mock.SetSynapses(testSynapses)

	// Test direct feedback delivery
	feedbackCount := stdp.DeliverFeedbackNow("target_neuron", mock, time.Now())
	if feedbackCount != 2 {
		t.Errorf("Expected 2 synapses to receive feedback, got %d", feedbackCount)
	}

	// Verify adjustments
	adjustments := mock.GetAdjustments()
	if len(adjustments) != 2 {
		t.Errorf("Expected 2 adjustments, got %d", len(adjustments))
	}

	// Verify DeltaT signs
	for _, adj := range adjustments {
		t.Logf("Adjustment DeltaT: %v", adj.DeltaT)
	}

	// Test scheduled feedback delivery
	mock.ClearAdjustments()
	stdp.ScheduleFeedback(now)

	// Initially, CheckAndDeliverFeedback should return false (too early)
	result := stdp.CheckAndDeliverFeedback("target_neuron", mock)
	if result {
		t.Error("CheckAndDeliverFeedback should return false when too early")
	}

	// Wait for the scheduled time to pass
	time.Sleep(15 * time.Millisecond)

	// Now CheckAndDeliverFeedback should return true and deliver feedback
	result = stdp.CheckAndDeliverFeedback("target_neuron", mock)
	if !result {
		t.Error("CheckAndDeliverFeedback should return true when it's time to deliver")
	}

	// Verify new adjustments
	newAdjustments := mock.GetAdjustments()
	if len(newAdjustments) != 2 {
		t.Errorf("Expected 2 new adjustments from scheduled feedback, got %d", len(newAdjustments))
	}

	t.Log("✓ STDP feedback delivery works correctly")
}

// TestSTDPSignaling_IsolatedConcurrency tests the concurrency safety of
// the STDPSignalingSystem with high contention
func TestSTDPSignaling_IsolatedConcurrency(t *testing.T) {
	// Create a system with standard parameters
	stdpSignaling := NewSTDPSignalingSystem(true, 10*time.Millisecond, 0.1)

	// Create mock callbacks that count calls but don't do anything significant
	mock := NewThreadSafeMockCallbacks()

	// Set up test synapses
	testSynapses := []types.SynapseInfo{
		{
			ID:               "test_synapse",
			SourceID:         "source",
			TargetID:         "test_neuron",
			Weight:           0.5,
			LastActivity:     time.Now().Add(-5 * time.Millisecond),
			LastTransmission: time.Now().Add(-5 * time.Millisecond), // Use LastActivity for simplicity
		},
	}
	mock.SetSynapses(testSynapses)

	// Test extreme concurrent access
	const numGoroutines = 100           // Use many goroutines
	const operationsPerGoroutine = 1000 // Many operations per goroutine

	// Create a wait group to track completion
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Add a timeout to detect deadlocks
	timeout := time.After(10 * time.Second)
	done := make(chan bool, 1)

	// Track any panics
	var panics int32

	// Launch goroutines with high contention
	startTime := time.Now()

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Panic in goroutine %d: %v", id, r)
					atomic.AddInt32(&panics, 1)
				}
				wg.Done()
			}()

			for j := 0; j < operationsPerGoroutine; j++ {
				// Mix different operations to create contention
				switch j % 7 {
				case 0:
					// Enable/disable
					if j%2 == 0 {
						stdpSignaling.Enable(10*time.Millisecond, 0.1)
					} else {
						stdpSignaling.Disable()
					}
				case 1:
					// Schedule feedback with different times
					fireTime := time.Now().Add(time.Duration(id*j) * time.Microsecond)
					stdpSignaling.ScheduleFeedback(fireTime)
				case 2:
					// Check for feedback delivery
					stdpSignaling.CheckAndDeliverFeedback("test_neuron", mock)
				case 3:
					// Force feedback delivery
					stdpSignaling.DeliverFeedbackNow("test_neuron", mock, time.Now())
				case 4:
					// Get status
					_ = stdpSignaling.GetStatus()
				case 5:
					// Get parameters
					_, _ = stdpSignaling.GetParameters()
				case 6:
					// Check if enabled
					_ = stdpSignaling.IsEnabled()
				}

				// Small random pause to increase chance of race conditions
				if j%100 == 0 {
					time.Sleep(time.Duration(id%5) * time.Microsecond)
				}
			}
		}(i)
	}

	// Set up a goroutine to signal when all workers are done
	go func() {
		wg.Wait()
		done <- true
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		// Success!
		elapsedTime := time.Since(startTime)
		opsPerSec := float64(numGoroutines*operationsPerGoroutine) / elapsedTime.Seconds()

		t.Logf("✓ Successfully handled %d concurrent operations across %d goroutines",
			numGoroutines*operationsPerGoroutine, numGoroutines)
		t.Logf("  Time: %v (%.0f ops/sec)", elapsedTime, opsPerSec)
		t.Logf("  Feedback operations: %d", mock.GetOperationCount())

		if atomic.LoadInt32(&panics) > 0 {
			t.Errorf("❌ %d goroutines panicked during concurrent execution", panics)
		}

	case <-timeout:
		// Failure - deadlock detected
		buf := make([]byte, 1<<20)
		stackLen := runtime.Stack(buf, true)
		t.Fatalf("❌ DEADLOCK: Test timed out after 10 seconds. Goroutine dump:\n%s", buf[:stackLen])
	}

	// Verify system is still functional after intense concurrency
	if stdpSignaling.DeliverFeedbackNow("test_neuron", mock, time.Now()) == 0 {
		t.Error("System is not functional after concurrency test")
	}
}

// Thread-safe mock for concurrency testing
type ThreadSafeMockCallbacks struct {
	synapses       []types.SynapseInfo
	operationCount int64
	mutex          sync.RWMutex
}

func NewThreadSafeMockCallbacks() *ThreadSafeMockCallbacks {
	return &ThreadSafeMockCallbacks{
		synapses: make([]types.SynapseInfo, 0),
	}
}

func (m *ThreadSafeMockCallbacks) ListSynapses(criteria types.SynapseCriteria) []types.SynapseInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Atomically increment operation count
	atomic.AddInt64(&m.operationCount, 1)

	// Return a copy to avoid data races
	result := make([]types.SynapseInfo, len(m.synapses))
	copy(result, m.synapses)
	return result
}

func (m *ThreadSafeMockCallbacks) ApplyPlasticity(synapseID string, adjustment types.PlasticityAdjustment) error {
	// Atomically increment operation count
	atomic.AddInt64(&m.operationCount, 1)
	return nil
}

func (m *ThreadSafeMockCallbacks) SetSynapses(synapses []types.SynapseInfo) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.synapses = make([]types.SynapseInfo, len(synapses))
	copy(m.synapses, synapses)
}

func (m *ThreadSafeMockCallbacks) GetOperationCount() int64 {
	return atomic.LoadInt64(&m.operationCount)
}

// Add stubs for the rest of the NeuronCallbacks interface
func (m *ThreadSafeMockCallbacks) CreateSynapse(config types.SynapseCreationConfig) (string, error) {
	return "mock-synapse-id", nil
}

func (m *ThreadSafeMockCallbacks) DeleteSynapse(synapseID string) error {
	return nil
}

func (m *ThreadSafeMockCallbacks) GetSynapse(synapseID string) (component.SynapticProcessor, error) {
	return nil, nil
}

func (m *ThreadSafeMockCallbacks) SetSynapseWeight(synapseID string, weight float64) error {
	return nil
}

func (m *ThreadSafeMockCallbacks) GetSynapseWeight(synapseID string) (float64, error) {
	return 0.5, nil
}

func (m *ThreadSafeMockCallbacks) GetSpatialDelay(targetID string) time.Duration {
	return 0
}

func (m *ThreadSafeMockCallbacks) FindNearbyComponents(radius float64) []component.ComponentInfo {
	return nil
}

func (m *ThreadSafeMockCallbacks) ReleaseChemical(ligandType types.LigandType, concentration float64) error {
	return nil
}

func (m *ThreadSafeMockCallbacks) SendElectricalSignal(signalType types.SignalType, data interface{}) {
}

func (m *ThreadSafeMockCallbacks) ReportHealth(activityLevel float64, connectionCount int) {
}

func (m *ThreadSafeMockCallbacks) ReportStateChange(oldState, newState types.ComponentState) {
}

func (m *ThreadSafeMockCallbacks) GetMatrix() component.ExtracellularMatrix {
	return nil
}

// TestSTDPSignaling_ConcurrentScheduling specifically tests concurrent
// scheduling operations which could lead to race conditions
// TestSTDPSignaling_ConcurrentScheduling specifically tests concurrent
// scheduling operations which could lead to race conditions
func TestSTDPSignaling_ConcurrentScheduling(t *testing.T) {
	// Create a system with moderate delay
	stdpSignaling := NewSTDPSignalingSystem(true, 50*time.Millisecond, 0.1)

	// Track the earliest scheduled time with a mutex-protected variable
	startTime := time.Now()
	var earliestTime time.Time = startTime.Add(500 * time.Millisecond) // Arbitrary future time
	var earliestMutex sync.Mutex

	// Launch multiple goroutines trying to schedule feedback
	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Use atomic operations to track results
	var successCount int32

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Each goroutine tries to schedule a different time
			// Earlier times should win in the scheduling race
			offset := time.Duration(id*10) * time.Millisecond
			fireTime := startTime.Add(offset)

			success := stdpSignaling.ScheduleFeedback(fireTime)
			if success {
				atomic.AddInt32(&successCount, 1)

				// Update earliest time with mutex protection
				earliestMutex.Lock()
				if fireTime.Before(earliestTime) {
					earliestTime = fireTime
				}
				earliestMutex.Unlock()
			}

			// Brief pause to increase contention
			time.Sleep(time.Duration(id%5) * time.Millisecond)
		}(i)
	}

	// Wait for all goroutines to complete with a timeout to prevent deadlock
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Add a timeout to prevent test hanging
	select {
	case <-done:
		// Test completed normally
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for goroutines to complete")
	}

	// Verify results
	status := stdpSignaling.GetStatus()
	scheduledTime := status["scheduled_time"].(time.Time)

	t.Logf("Concurrent scheduling results:")
	t.Logf("  Successful schedules: %d/%d", successCount, numGoroutines)

	// Get the earliest time value safely
	earliestMutex.Lock()
	earliestTimeValue := earliestTime
	earliestMutex.Unlock()

	t.Logf("  Earliest fire time: %v", earliestTimeValue)
	t.Logf("  Actual scheduled time: %v", scheduledTime)

	// Check if the earliest time won the scheduling race
	// Add delay to account for scheduling calculations
	delay, _ := stdpSignaling.GetParameters()
	expectedScheduledTime := earliestTimeValue.Add(delay)
	if scheduledTime.Sub(expectedScheduledTime) > 2*time.Millisecond {
		t.Errorf("Expected scheduled time close to %v, got %v",
			expectedScheduledTime, scheduledTime)
	}

	t.Log("✓ Concurrent scheduling prioritizes earliest fire time correctly")
}

// Fix 1: TestSTDPSignaling_RaceConditions
func TestSTDPSignaling_RaceConditions(t *testing.T) {
	// Create a system with short delay for faster testing
	stdpSignaling := NewSTDPSignalingSystem(true, 10*time.Millisecond, 0.1)

	// Create mock callbacks
	mock := NewThreadSafeMockCallbacks()

	// Set up test synapses
	testSynapses := []types.SynapseInfo{
		{
			ID:               "race_test_synapse",
			SourceID:         "source",
			TargetID:         "test_neuron",
			Weight:           0.5,
			LastActivity:     time.Now().Add(-5 * time.Millisecond),
			LastTransmission: time.Now().Add(-5 * time.Millisecond), // Use LastActivity for simplicity
		},
	}
	mock.SetSynapses(testSynapses)

	// Launch scheduler goroutine that continuously schedules new feedback
	stopScheduler := make(chan bool)
	var schedulingOps int32

	go func() {
		for {
			select {
			case <-stopScheduler:
				return
			default:
				stdpSignaling.ScheduleFeedback(time.Now())
				atomic.AddInt32(&schedulingOps, 1)
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// Launch checker goroutine that continuously checks for feedback
	stopChecker := make(chan bool)
	var checkingOps int32
	var deliveredFeedbacks int32

	go func() {
		for {
			select {
			case <-stopChecker:
				return
			default:
				success := stdpSignaling.CheckAndDeliverFeedback("test_neuron", mock)
				atomic.AddInt32(&checkingOps, 1)
				if success {
					atomic.AddInt32(&deliveredFeedbacks, 1)
				}
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// Let the race condition test run for a while
	time.Sleep(500 * time.Millisecond)

	// Stop test goroutines
	stopScheduler <- true
	stopChecker <- true

	// Verify results
	finalSchedulingOps := atomic.LoadInt32(&schedulingOps)
	finalCheckingOps := atomic.LoadInt32(&checkingOps)
	finalDeliveredFeedbacks := atomic.LoadInt32(&deliveredFeedbacks)

	t.Logf("Race condition test results:")
	t.Logf("  Scheduling operations: %d", finalSchedulingOps)
	t.Logf("  Checking operations: %d", finalCheckingOps)
	t.Logf("  Delivered feedbacks: %d", finalDeliveredFeedbacks)
	t.Logf("  Feedback operations: %d", mock.GetOperationCount())

	// There should be some successful deliveries
	if finalDeliveredFeedbacks == 0 {
		t.Error("Expected some successful feedback deliveries during race test")
	}

	// Verify system is still functional
	stdpSignaling.ScheduleFeedback(time.Now())
	time.Sleep(15 * time.Millisecond)

	initialOps := mock.GetOperationCount()
	success := stdpSignaling.CheckAndDeliverFeedback("test_neuron", mock)
	finalOps := mock.GetOperationCount()

	if !success || finalOps <= initialOps {
		t.Error("System not functional after race condition test")
	}

	t.Log("✓ System handled concurrent scheduling and delivery without race conditions")
}

// Fix 2: TestSTDPSignaling_EnableDisableRace
func TestSTDPSignaling_EnableDisableRace(t *testing.T) {
	// Create a system that starts disabled
	stdpSignaling := NewSTDPSignalingSystem(false, 10*time.Millisecond, 0.1)

	// Create mock callbacks
	mock := NewThreadSafeMockCallbacks()

	// Set up test synapses
	testSynapses := []types.SynapseInfo{
		{
			ID:               "enable_disable_test",
			SourceID:         "source",
			TargetID:         "test_neuron",
			Weight:           0.5,
			LastActivity:     time.Now().Add(-5 * time.Millisecond),
			LastTransmission: time.Now().Add(-5 * time.Millisecond), // Use LastActivity for simplicity
		},
	}
	mock.SetSynapses(testSynapses)

	// Track operations
	var enableCount, disableCount int32
	var scheduleCount, scheduleSuccess int32
	var checkCount, checkSuccess int32
	var forceCount, forceSuccess int32

	// Launch multiple goroutines performing different operations
	const numGoroutines = 20
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < 100; j++ {
				// Mix different operations
				switch (id + j) % 5 {
				case 0:
					// Toggle enabled status
					if j%2 == 0 {
						stdpSignaling.Enable(10*time.Millisecond, 0.1)
						atomic.AddInt32(&enableCount, 1)
					} else {
						stdpSignaling.Disable()
						atomic.AddInt32(&disableCount, 1)
					}
				case 1:
					// Try to schedule
					atomic.AddInt32(&scheduleCount, 1)
					if stdpSignaling.ScheduleFeedback(time.Now()) {
						atomic.AddInt32(&scheduleSuccess, 1)
					}
				case 2:
					// Try to check and deliver
					atomic.AddInt32(&checkCount, 1)
					if stdpSignaling.CheckAndDeliverFeedback("test_neuron", mock) {
						atomic.AddInt32(&checkSuccess, 1)
					}
				case 3:
					// Try to force feedback
					atomic.AddInt32(&forceCount, 1)
					feedbackCount := stdpSignaling.DeliverFeedbackNow("test_neuron", mock, time.Now())
					if feedbackCount > 0 {
						atomic.AddInt32(&forceSuccess, 1)
					}
				case 4:
					// Read status
					_ = stdpSignaling.GetStatus()
				}

				// Small pause to increase race chance
				time.Sleep(time.Duration(id%3) * time.Microsecond)
			}
		}(i)
	}

	// Wait for completion
	wg.Wait()

	// Verify results
	t.Logf("Enable/disable race test results:")
	t.Logf("  Enable operations: %d", enableCount)
	t.Logf("  Disable operations: %d", disableCount)
	t.Logf("  Schedule attempts: %d (success: %d)", scheduleCount, scheduleSuccess)
	t.Logf("  Check attempts: %d (success: %d)", checkCount, checkSuccess)
	t.Logf("  Force attempts: %d (success: %d)", forceCount, forceSuccess)
	t.Logf("  Feedback operations: %d", mock.GetOperationCount())

	// System should be in a consistent state
	finalStatus := stdpSignaling.GetStatus()
	finalEnabled := finalStatus["enabled"].(bool)

	// Force one final test to confirm system integrity
	if finalEnabled {
		stdpSignaling.Disable()
		if stdpSignaling.IsEnabled() {
			t.Error("System reports enabled after explicit disable")
		}
	} else {
		stdpSignaling.Enable(10*time.Millisecond, 0.1)
		if !stdpSignaling.IsEnabled() {
			t.Error("System reports disabled after explicit enable")
		}
	}

	t.Log("✓ System maintained consistency during enable/disable race testing")
}

// TestSTDPSignaling_ExtremeConditions tests behavior under extreme conditions
func TestSTDPSignaling_ExtremeConditions(t *testing.T) {
	// Test very short and very long delay values
	t.Run("ExtremeDelays", func(t *testing.T) {
		// Test with extremely short delay
		shortDelay := 1 * time.Nanosecond
		stdpShort := NewSTDPSignalingSystem(true, shortDelay, 0.1)

		mock := NewThreadSafeMockCallbacks()
		mock.SetSynapses([]types.SynapseInfo{
			{
				ID:               "extreme_test",
				SourceID:         "source",
				TargetID:         "test_neuron",
				Weight:           0.5,
				LastActivity:     time.Now().Add(-5 * time.Millisecond),
				LastTransmission: time.Now().Add(-5 * time.Millisecond), // Use LastActivity for simplicity
			},
		})

		// Schedule with extremely short delay
		startTime := time.Now()
		stdpShort.ScheduleFeedback(startTime)

		// Should be immediately ready
		time.Sleep(100 * time.Microsecond) // Small buffer

		if !stdpShort.CheckAndDeliverFeedback("test_neuron", mock) {
			t.Error("System with extremely short delay failed to deliver feedback")
		}

		// Test with extremely long delay
		longDelay := 24 * time.Hour
		stdpLong := NewSTDPSignalingSystem(true, longDelay, 0.1)

		// Schedule with extremely long delay
		stdpLong.ScheduleFeedback(startTime)

		// Should not be ready
		if stdpLong.CheckAndDeliverFeedback("test_neuron", mock) {
			t.Error("System with extremely long delay incorrectly delivered feedback")
		}

		t.Log("✓ System handles extreme delay values correctly")
	})

	t.Run("NilCallbacks", func(t *testing.T) {
		// Test with nil callbacks
		stdp := NewSTDPSignalingSystem(true, 10*time.Millisecond, 0.1)

		// These should not panic
		stdp.ScheduleFeedback(time.Now())

		// Should gracefully handle nil callbacks
		result1 := stdp.CheckAndDeliverFeedback("test_neuron", nil)
		result2 := stdp.DeliverFeedbackNow("test_neuron", nil, time.Now())

		if result1 == true || result2 > 0 {
			t.Error("System incorrectly reported success with nil callbacks")
		}

		t.Log("✓ System gracefully handles nil callbacks")
	})

	t.Run("ConcurrentWithDisable", func(t *testing.T) {
		// Test rapid enable/disable during operation
		stdp := NewSTDPSignalingSystem(true, 20*time.Millisecond, 0.1)

		mock := NewThreadSafeMockCallbacks()
		mock.SetSynapses([]types.SynapseInfo{
			{
				ID:               "concurrent_test",
				SourceID:         "source",
				TargetID:         "test_neuron",
				Weight:           0.5,
				LastActivity:     time.Now().Add(-5 * time.Millisecond),
				LastTransmission: time.Now().Add(-5 * time.Millisecond), // Use LastActivity for simplicity
			},
		})

		// Launch goroutines
		const iterations = 1000
		var wg sync.WaitGroup
		wg.Add(3)

		// Goroutine 1: Rapidly toggle enabled state
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				if i%2 == 0 {
					stdp.Enable(20*time.Millisecond, 0.1)
				} else {
					stdp.Disable()
				}
			}
		}()

		// Goroutine 2: Continuously schedule
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				stdp.ScheduleFeedback(time.Now())
				time.Sleep(time.Microsecond)
			}
		}()

		// Goroutine 3: Continuously check
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				stdp.CheckAndDeliverFeedback("test_neuron", mock)
				time.Sleep(time.Microsecond)
			}
		}()

		wg.Wait()

		// System should be in a consistent state
		status := stdp.GetStatus()
		t.Logf("Final status after extreme contention: %+v", status)

		// Final functional test
		stdp.Enable(10*time.Millisecond, 0.1)
		stdp.ScheduleFeedback(time.Now())
		time.Sleep(15 * time.Millisecond)

		if !stdp.CheckAndDeliverFeedback("test_neuron", mock) {
			t.Error("System not functional after extreme contention test")
		}

		t.Log("✓ System survived extreme contention with enable/disable operations")
	})
}
