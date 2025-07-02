// neuron/neuron_stdp_deadlock_test.go
package neuron

import (
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

/*
=================================================================================
STDP DEADLOCK TEST - REPRODUCES AND VALIDATES DEADLOCK FIX
=================================================================================

This test reproduces the exact deadlock scenario that occurs when:
1. Neuron fires and calls SendSTDPFeedback()
2. SendSTDPFeedback() calls matrix callbacks
3. Matrix callbacks call back to GetActivityLevel()
4. GetActivityLevel() tries to acquire the same stateMutex → DEADLOCK

The test will:
- FAIL (timeout/hang) with the current implementation
- PASS after implementing separate mutexes or proper lock ordering

=================================================================================
*/

// DeadlockTriggerCallbacks extends your existing MockNeuronCallbacks to trigger the deadlock
type DeadlockTriggerCallbacks struct {
	*MockNeuronCallbacks
	neuron *Neuron // Reference to create re-entrant calls
}

func NewDeadlockTriggerCallbacks(matrix *MockMatrix, neuron *Neuron) *DeadlockTriggerCallbacks {
	return &DeadlockTriggerCallbacks{
		MockNeuronCallbacks: NewMockNeuronCallbacks(matrix),
		neuron:              neuron,
	}
}

// Override ListSynapses to trigger the deadlock
func (dtc *DeadlockTriggerCallbacks) ListSynapses(criteria types.SynapseCriteria) []types.SynapseInfo {
	// *** THIS IS THE DEADLOCK TRIGGER ***
	// When SendSTDPFeedback() calls this, and we call back to GetActivityLevel(),
	// it creates a re-entrant lock attempt on the same stateMutex

	// Simulate matrix callback that tries to read neuron activity
	_ = dtc.neuron.GetActivityLevel() // This will deadlock with current implementation

	// Return dummy synapse info
	return []types.SynapseInfo{
		{
			ID:           "test_synapse",
			SourceID:     "source",
			TargetID:     dtc.neuron.ID(),
			Weight:       1.0,
			LastActivity: time.Now(),
		},
	}
}

// TestSTDPDeadlockReproduction reproduces the deadlock scenario
func TestNeuronSTDP_DeadlockReproduction(t *testing.T) {
	t.Log("=== STDP DEADLOCK REPRODUCTION TEST ===")
	t.Log("This test will hang/timeout with current implementation, pass after fix")

	// Create a neuron with STDP enabled
	neuron := NewNeuron(
		"test_neuron_deadlock",
		1.0,                // threshold
		0.95,               // decay rate
		5*time.Millisecond, // refractory period
		2.0,                // fire factor
		3.0,                // target firing rate
		0.2,                // homeostasis strength
	)

	// Enable STDP feedback (this is what causes the deadlock)
	neuron.EnableSTDPFeedback(
		5*time.Millisecond, // feedback delay
		0.01,               // learning rate
	)

	// Create mock callbacks that will cause re-entrant lock
	mockMatrix := NewMockMatrix()
	deadlockCallbacks := NewDeadlockTriggerCallbacks(mockMatrix, neuron)
	neuron.SetCallbacks(deadlockCallbacks)

	// Start the neuron
	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	t.Log("Neuron started, attempting to trigger STDP deadlock...")

	// Create a test channel to detect if the test completes
	done := make(chan bool, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
			done <- true
		}()

		// *** THIS IS THE DEADLOCK TRIGGER ***
		// Step 1: This will acquire stateMutex and call ListSynapses
		// Step 2: ListSynapses calls GetActivityLevel()
		// Step 3: GetActivityLevel() tries to acquire the same stateMutex → DEADLOCK
		t.Log("Calling SendSTDPFeedback() - this should deadlock with current implementation")
		neuron.SendSTDPFeedback()
		t.Log("SendSTDPFeedback() completed successfully")
	}()

	// Wait for either completion or timeout
	select {
	case <-done:
		t.Log("✅ SUCCESS: No deadlock detected - the fix is working!")
	case <-time.After(2 * time.Second):
		t.Fatal("❌ DEADLOCK: Test timed out - SendSTDPFeedback() is hanging due to re-entrant lock")
	}
}

// TestSTDPDeadlockWithActivityRead specifically tests the GetActivityLevel() deadlock
func TestSTDPDeadlockWithActivityRead(t *testing.T) {
	t.Log("=== STDP + ACTIVITY READ DEADLOCK TEST ===")
	t.Log("Tests concurrent GetActivityLevel() calls during STDP feedback")

	neuron := NewNeuron(
		"test_neuron_activity",
		1.0,
		0.95,
		5*time.Millisecond,
		2.0,
		3.0,
		0.2,
	)

	// Enable STDP
	neuron.EnableSTDPFeedback(10*time.Millisecond, 0.01)

	// Mock callbacks that read activity
	mockMatrix := NewMockMatrix()
	deadlockCallbacks := NewDeadlockTriggerCallbacks(mockMatrix, neuron)
	neuron.SetCallbacks(deadlockCallbacks)

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Test concurrent access pattern that causes deadlock
	results := make(chan error, 2)

	// Goroutine 1: Continuously read activity level
	go func() {
		defer func() {
			if r := recover(); r != nil {
				results <- nil // Panic is okay, deadlock is not
			} else {
				results <- nil // Success
			}
		}()

		for i := 0; i < 10; i++ {
			_ = neuron.GetActivityLevel()
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Goroutine 2: Trigger STDP feedback
	go func() {
		defer func() {
			if r := recover(); r != nil {
				results <- nil // Panic is okay, deadlock is not
			} else {
				results <- nil // Success
			}
		}()

		time.Sleep(5 * time.Millisecond) // Let activity reading start
		neuron.SendSTDPFeedback()
	}()

	// Wait for both goroutines or timeout
	timeout := time.After(3 * time.Second)
	completed := 0

	for completed < 2 {
		select {
		case <-results:
			completed++
		case <-timeout:
			t.Fatal("❌ DEADLOCK: Concurrent activity read and STDP feedback deadlocked")
		}
	}

	t.Log("✅ SUCCESS: Concurrent activity reads and STDP feedback completed without deadlock")
}

// TestSTDPDeadlockFixValidation validates the fix is working correctly
func TestSTDPDeadlockFixValidation(t *testing.T) {
	t.Log("=== STDP DEADLOCK FIX VALIDATION ===")
	t.Log("Validates that STDP feedback works correctly after deadlock fix")

	neuron := NewNeuron(
		"test_neuron_validation",
		1.0,
		0.95,
		5*time.Millisecond,
		2.0,
		3.0,
		0.2,
	)

	// Enable STDP
	neuron.EnableSTDPFeedback(5*time.Millisecond, 0.01)

	// Use real callbacks that don't cause re-entrant locks
	mockMatrix := NewMockMatrix()
	simpleCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(simpleCallbacks)

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// This should work smoothly after the fix
	initialActivity := neuron.GetActivityLevel()

	// Call STDP multiple times - should not hang
	for i := 0; i < 5; i++ {
		neuron.SendSTDPFeedback()

		// Should be able to read activity level without issues
		currentActivity := neuron.GetActivityLevel()
		t.Logf("Iteration %d: Activity level = %.6f", i+1, currentActivity)

		time.Sleep(10 * time.Millisecond)
	}

	finalActivity := neuron.GetActivityLevel()
	t.Logf("Activity progression: %.6f → %.6f", initialActivity, finalActivity)

	t.Log("✅ SUCCESS: STDP feedback and activity reading work correctly after fix")
}
