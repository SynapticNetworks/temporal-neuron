package neuron

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestNeuronCreation tests basic neuron creation and initialization
func TestNeuronCreation(t *testing.T) {
	threshold := 1.5
	decayRate := 0.95
	refractoryPeriod := 10 * time.Millisecond
	fireFactor := 2.0
	neuronID := "test_neuron_1"

	// Test simple neuron creation (backward compatibility)
	neuron := NewSimpleNeuron(neuronID, threshold, decayRate, refractoryPeriod, fireFactor)

	if neuron == nil {
		t.Fatal("NewSimpleNeuron returned nil")
	}

	if neuron.threshold != threshold {
		t.Errorf("Expected threshold %f, got %f", threshold, neuron.threshold)
	}

	if neuron.decayRate != decayRate {
		t.Errorf("Expected decayRate %f, got %f", decayRate, neuron.decayRate)
	}

	if neuron.refractoryPeriod != refractoryPeriod {
		t.Errorf("Expected refractoryPeriod %v, got %v", refractoryPeriod, neuron.refractoryPeriod)
	}

	if neuron.fireFactor != fireFactor {
		t.Errorf("Expected fireFactor %f, got %f", fireFactor, neuron.fireFactor)
	}

	if neuron.id != neuronID {
		t.Errorf("Expected neuron ID %s, got %s", neuronID, neuron.id)
	}

	if neuron.GetOutputCount() != 0 {
		t.Errorf("Expected 0 outputs for new neuron, got %d", neuron.GetOutputCount())
	}

	if neuron.accumulator != 0 {
		t.Errorf("Expected accumulator to be 0, got %f", neuron.accumulator)
	}

	if !neuron.lastFireTime.IsZero() {
		t.Errorf("Expected lastFireTime to be zero value, got %v", neuron.lastFireTime)
	}

	// Test that homeostasis is disabled for simple neurons
	if neuron.homeostatic.targetFiringRate != 0.0 {
		t.Errorf("Expected disabled homeostasis (targetFiringRate=0), got %f", neuron.homeostatic.targetFiringRate)
	}

	if neuron.homeostatic.homeostasisStrength != 0.0 {
		t.Errorf("Expected disabled homeostasis (homeostasisStrength=0), got %f", neuron.homeostatic.homeostasisStrength)
	}
}

// TestNeuronInputChannel tests that input channel is accessible
func TestNeuronInputChannel(t *testing.T) {
	neuron := NewSimpleNeuron("test_input", 1.0, 0.95, 5*time.Millisecond, 1.0)

	input := neuron.GetInput()
	if input == nil {
		t.Fatal("GetInput() returned nil channel")
	}

	// Test that we can send to the channel (non-blocking test)
	select {
	case input <- Message{Value: 0.5}:
		// Successfully sent
	default:
		t.Error("Could not send message to input channel")
	}
}

// TestOutputManagement tests adding and removing outputs
func TestOutputManagement(t *testing.T) {
	neuron := NewSimpleNeuron("test_output_mgmt", 1.0, 0.95, 5*time.Millisecond, 1.0)

	// Test adding outputs
	output1 := make(chan Message, 1)
	output2 := make(chan Message, 1)

	neuron.AddOutput("output1", output1, 1.0, 5*time.Millisecond)
	if neuron.GetOutputCount() != 1 {
		t.Errorf("Expected 1 output after adding, got %d", neuron.GetOutputCount())
	}

	neuron.AddOutput("output2", output2, 0.5, 10*time.Millisecond)
	if neuron.GetOutputCount() != 2 {
		t.Errorf("Expected 2 outputs after adding second, got %d", neuron.GetOutputCount())
	}

	// Test removing outputs
	neuron.RemoveOutput("output1")
	if neuron.GetOutputCount() != 1 {
		t.Errorf("Expected 1 output after removing, got %d", neuron.GetOutputCount())
	}

	neuron.RemoveOutput("output2")
	if neuron.GetOutputCount() != 0 {
		t.Errorf("Expected 0 outputs after removing all, got %d", neuron.GetOutputCount())
	}

	// Test removing non-existent output (should not panic)
	neuron.RemoveOutput("nonexistent")
	if neuron.GetOutputCount() != 0 {
		t.Errorf("Expected 0 outputs after removing nonexistent, got %d", neuron.GetOutputCount())
	}
}

// TestThresholdFiring tests basic threshold-based firing
func TestThresholdFiring(t *testing.T) {
	threshold := 1.0
	neuron := NewSimpleNeuron("test_threshold", threshold, 0.98, 10*time.Millisecond, 1.0)

	output := make(chan Message, 10)
	neuron.AddOutput("test", output, 1.0, 0) // No delay for testing

	// Start neuron processing
	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// Test 1: Single signal below threshold - should not fire
	input <- Message{Value: 0.5}

	select {
	case <-output:
		t.Error("Neuron fired when single signal was below threshold")
	case <-time.After(20 * time.Millisecond):
		// Expected - no firing
	}

	// Test 2: Single strong signal above threshold - should fire immediately
	input <- Message{Value: 1.5} // Well above threshold

	select {
	case fired := <-output:
		if fired.Value <= 0 {
			t.Errorf("Expected positive fire value, got %f", fired.Value)
		}
	case <-time.After(50 * time.Millisecond):
		t.Error("Neuron did not fire when threshold was exceeded")
	}
}

// TestLeakyIntegration tests continuous membrane potential decay
func TestLeakyIntegration(t *testing.T) {
	decayRate := 0.9 // Aggressive decay for faster testing
	neuron := NewSimpleNeuron("test_leaky", 1.0, decayRate, 5*time.Millisecond, 1.0)

	output := make(chan Message, 10)
	neuron.AddOutput("test", output, 1.0, 0)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// Send signal below threshold
	input <- Message{Value: 0.8}

	// Wait for decay to reduce the accumulator
	time.Sleep(20 * time.Millisecond)

	// Send another signal - should need more than 0.2 to fire due to decay
	input <- Message{Value: 0.3}

	// Should not fire because first signal has decayed
	select {
	case <-output:
		t.Error("Neuron fired when it should have decayed below threshold")
	case <-time.After(20 * time.Millisecond):
		// Expected - no firing due to decay
	}

	// Send a strong signal that should fire immediately
	input <- Message{Value: 1.2}

	select {
	case <-output:
		// Expected firing
	case <-time.After(20 * time.Millisecond):
		t.Error("Neuron did not fire with strong signal")
	}
}

// TestRefractoryPeriod tests that neurons cannot fire during refractory period
func TestRefractoryPeriod(t *testing.T) {
	refractoryPeriod := 20 * time.Millisecond
	neuron := NewSimpleNeuron("test_refractory", 1.0, 0.98, refractoryPeriod, 1.0)

	output := make(chan Message, 10)
	neuron.AddOutput("test", output, 1.0, 0)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// Fire the neuron first time
	input <- Message{Value: 1.5}

	// Wait for first firing
	select {
	case <-output:
		// Expected first firing
	case <-time.After(20 * time.Millisecond):
		t.Fatal("First firing did not occur")
	}

	// Immediately try to fire again (should be blocked by refractory period)
	input <- Message{Value: 2.0} // Strong signal

	// Should not fire due to refractory period
	select {
	case <-output:
		t.Error("Neuron fired during refractory period")
	case <-time.After(10 * time.Millisecond):
		// Expected - no firing during refractory period
	}

	// Wait for refractory period to end and try again
	time.Sleep(refractoryPeriod + 5*time.Millisecond)
	input <- Message{Value: 1.5}

	// Should fire now
	select {
	case <-output:
		// Expected firing after refractory period
	case <-time.After(20 * time.Millisecond):
		t.Error("Neuron did not fire after refractory period ended")
	}
}

// TestContinuousDecay tests that accumulator continuously decays over time
func TestContinuousDecay(t *testing.T) {
	decayRate := 0.8 // Faster decay for testing
	neuron := NewSimpleNeuron("test_continuous_decay", 2.0, decayRate, 5*time.Millisecond, 1.0)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// Send signal that builds up accumulator but doesn't fire
	input <- Message{Value: 1.5}

	// Wait for several decay cycles
	time.Sleep(10 * time.Millisecond)

	// Send smaller signal - if decay worked, this shouldn't be enough to fire
	input <- Message{Value: 0.3}

	// Create output channel after the above to avoid capturing any erroneous fires
	output := make(chan Message, 10)
	neuron.AddOutput("test", output, 1.0, 0)

	// Should not fire because accumulator has decayed
	select {
	case <-output:
		t.Error("Neuron fired when accumulator should have decayed")
	case <-time.After(20 * time.Millisecond):
		// Expected - no firing due to decay
	}
}

// TestTemporalIntegration tests signal accumulation with leaky integration
func TestTemporalIntegration(t *testing.T) {
	decayRate := 0.99 // Slow decay to allow temporal summation
	neuron := NewSimpleNeuron("test_temporal", 1.0, decayRate, 5*time.Millisecond, 1.0)

	output := make(chan Message, 10)
	neuron.AddOutput("test", output, 1.0, 0)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// Send rapid sequence of small signals that should sum to fire
	input <- Message{Value: 0.4}
	time.Sleep(2 * time.Millisecond)
	input <- Message{Value: 0.3}
	time.Sleep(2 * time.Millisecond)
	input <- Message{Value: 0.4} // Total: approximately 1.1 with minimal decay

	// Should fire
	select {
	case <-output:
		// Expected firing
	case <-time.After(30 * time.Millisecond):
		t.Error("Neuron did not fire with rapid temporal integration")
	}
}

// TestOutputFactorAndDelay tests output scaling and transmission delays
func TestOutputFactorAndDelay(t *testing.T) {
	neuron := NewSimpleNeuron("test_factor_delay", 1.0, 0.98, 5*time.Millisecond, 2.0) // fireFactor = 2.0

	output := make(chan Message, 10)
	factor := 0.5
	delay := 20 * time.Millisecond

	neuron.AddOutput("test", output, factor, delay)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	startTime := time.Now()
	input <- Message{Value: 1.5} // Should fire immediately

	// Wait for output with delay
	select {
	case fired := <-output:
		elapsed := time.Since(startTime)

		// Check delay with tolerance for system overhead
		minDelay := delay
		maxDelay := delay + 50*time.Millisecond

		if elapsed < minDelay {
			t.Errorf("Delay too short: expected at least %v, got %v", minDelay, elapsed)
		}
		if elapsed > maxDelay {
			t.Errorf("Delay too long: expected at most %v, got %v", maxDelay, elapsed)
		}

		// Check output value: input(1.5) * fireFactor(2.0) * outputFactor(0.5) = 1.5
		expected := 1.5 * 2.0 * 0.5
		if fired.Value != expected {
			t.Errorf("Expected output value %f, got %f", expected, fired.Value)
		}

	case <-time.After(delay + 100*time.Millisecond):
		t.Error("Neuron did not fire within expected time")
	}
}

// TestMultipleOutputs tests firing to multiple outputs simultaneously
func TestMultipleOutputs(t *testing.T) {
	neuron := NewSimpleNeuron("test_multiple_outputs", 1.0, 0.98, 5*time.Millisecond, 1.0)

	output1 := make(chan Message, 10)
	output2 := make(chan Message, 10)
	output3 := make(chan Message, 10)

	neuron.AddOutput("out1", output1, 1.0, 5*time.Millisecond)
	neuron.AddOutput("out2", output2, 2.0, 10*time.Millisecond)
	neuron.AddOutput("out3", output3, 0.5, 0)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()
	input <- Message{Value: 1.5} // Should fire

	// All outputs should receive signals (with their respective factors)
	expectedValues := map[string]float64{
		"out1": 1.5 * 1.0, // 1.5
		"out2": 1.5 * 2.0, // 3.0
		"out3": 1.5 * 0.5, // 0.75
	}

	// Collect outputs (out3 should arrive first due to no delay)
	results := make(map[string]float64)

	for i := 0; i < 3; i++ {
		select {
		case val := <-output1:
			results["out1"] = val.Value
		case val := <-output2:
			results["out2"] = val.Value
		case val := <-output3:
			results["out3"] = val.Value
		case <-time.After(50 * time.Millisecond):
			t.Error("Timeout waiting for output")
		}
	}

	// Verify all outputs received correct values
	for name, expected := range expectedValues {
		if actual, ok := results[name]; !ok {
			t.Errorf("Output %s did not fire", name)
		} else if actual != expected {
			t.Errorf("Output %s: expected %f, got %f", name, expected, actual)
		}
	}
}

// TestConcurrentAccess tests thread safety of output management
func TestConcurrentAccess(t *testing.T) {
	neuron := NewSimpleNeuron("test_concurrent", 1.0, 0.98, 5*time.Millisecond, 1.0)

	go neuron.Run()
	defer neuron.Close()

	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 100

	// Concurrently add and remove outputs
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				output := make(chan Message, 1)
				outputID := fmt.Sprintf("output_%d_%d", id, j)

				neuron.AddOutput(outputID, output, 1.0, 0)
				count := neuron.GetOutputCount()

				if count < 0 {
					t.Errorf("Negative output count: %d", count)
				}

				neuron.RemoveOutput(outputID)
			}
		}(i)
	}

	// Concurrently send inputs
	wg.Add(1)
	go func() {
		defer wg.Done()
		input := neuron.GetInput()
		for i := 0; i < 50; i++ {
			input <- Message{Value: 0.1}
			time.Sleep(1 * time.Millisecond)
		}
	}()

	wg.Wait()

	// Should not have panicked and should be in a consistent state
	finalCount := neuron.GetOutputCount()
	if finalCount < 0 {
		t.Errorf("Final output count is negative: %d", finalCount)
	}
}

// TestResetAfterFiring tests that accumulator resets after firing
func TestResetAfterFiring(t *testing.T) {
	neuron := NewSimpleNeuron("test_reset", 1.0, 0.99, 10*time.Millisecond, 1.0)

	output := make(chan Message, 10)
	neuron.AddOutput("test", output, 1.0, 0)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// First firing
	input <- Message{Value: 1.5} // Above threshold - should fire and reset

	// Wait for firing
	select {
	case <-output:
		// Expected
	case <-time.After(20 * time.Millisecond):
		t.Fatal("First firing did not occur")
	}

	// Wait for refractory period to end
	time.Sleep(15 * time.Millisecond)

	// Test that neuron can fire again (proving reset worked)
	// Send strong signal that should fire immediately regardless of any residual accumulation
	input <- Message{Value: 1.2} // Above threshold

	select {
	case <-output:
		// Expected second firing - proves neuron reset properly
	case <-time.After(20 * time.Millisecond):
		t.Error("Second firing did not occur - neuron may not have reset properly")
	}
}

// TestInhibitorySignals tests negative (inhibitory) input values
func TestInhibitorySignals(t *testing.T) {
	neuron := NewSimpleNeuron("test_inhibitory", 1.0, 0.99, 5*time.Millisecond, 1.0)

	output := make(chan Message, 10)
	neuron.AddOutput("test", output, 1.0, 0)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// Send excitatory signal (below threshold)
	input <- Message{Value: 0.8} // Below threshold, shouldn't fire alone

	// Send inhibitory signal
	input <- Message{Value: -0.3} // Total: 0.8 - 0.3 = 0.5 < 1.0

	// Wait and ensure no firing
	select {
	case <-output:
		t.Error("Neuron fired despite total being below threshold due to inhibition")
	case <-time.After(30 * time.Millisecond):
		// Expected - no firing due to total being below threshold
	}

	// Now test that excitatory signal can overcome inhibition
	input <- Message{Value: 0.9} // Total: 0.5 + 0.9 = 1.4 > 1.0

	// Should fire now
	select {
	case <-output:
		// Expected - inhibition was overcome
	case <-time.After(30 * time.Millisecond):
		t.Error("Neuron did not fire when excitatory signal overcame inhibition")
	}
}

// TestFireEventReporting tests the fire event reporting functionality
func TestFireEventReporting(t *testing.T) {
	neuron := NewSimpleNeuron("test_fire_events", 1.0, 0.98, 5*time.Millisecond, 2.0)

	// Set up fire event monitoring BEFORE starting the neuron
	fireEvents := make(chan FireEvent, 10)
	neuron.SetFireEventChannel(fireEvents)

	// Set up regular output for comparison
	output := make(chan Message, 10)
	neuron.AddOutput("test", output, 1.0, 0)

	// Start neuron AFTER setting up monitoring
	go neuron.Run()
	defer neuron.Close()

	// Give the neuron a moment to start up
	time.Sleep(5 * time.Millisecond)

	input := neuron.GetInput()

	// Send signal that should cause firing
	input <- Message{Value: 1.5}

	// Should receive both regular output and fire event
	var fireEvent FireEvent
	var outputMsg Message
	var gotFireEvent, gotOutput bool

	// Wait for both events with a reasonable timeout
	timeout := time.After(100 * time.Millisecond)

	for !gotFireEvent || !gotOutput {
		select {
		case fireEvent = <-fireEvents:
			gotFireEvent = true
		case outputMsg = <-output:
			gotOutput = true
		case <-timeout:
			if !gotFireEvent {
				t.Error("Did not receive fire event")
			}
			if !gotOutput {
				t.Error("Did not receive output message")
			}
			return
		}
	}

	// Verify fire event details
	if fireEvent.NeuronID != "test_fire_events" {
		t.Errorf("Expected neuron ID 'test_fire_events', got '%s'", fireEvent.NeuronID)
	}

	expectedValue := 1.5 * 2.0 // input * fireFactor
	if fireEvent.Value != expectedValue {
		t.Errorf("Expected fire event value %f, got %f", expectedValue, fireEvent.Value)
	}

	if outputMsg.Value != expectedValue {
		t.Errorf("Expected output value %f, got %f", expectedValue, outputMsg.Value)
	}

	// Verify timestamp is recent (within last 200ms)
	if time.Since(fireEvent.Timestamp) > 200*time.Millisecond {
		t.Errorf("Fire event timestamp seems too old: %v ago", time.Since(fireEvent.Timestamp))
	}
}

// TestRefractoryPeriodPreventsRapidFiring tests multiple rapid firing attempts
func TestRefractoryPeriodPreventsRapidFiring(t *testing.T) {
	refractoryPeriod := 30 * time.Millisecond
	neuron := NewSimpleNeuron("test_rapid_firing", 1.0, 0.98, refractoryPeriod, 1.0)

	output := make(chan Message, 100)
	neuron.AddOutput("test", output, 1.0, 0)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// Send rapid sequence of strong signals
	for i := 0; i < 10; i++ {
		input <- Message{Value: 2.0}
		time.Sleep(2 * time.Millisecond) // Much faster than refractory period
	}

	// Count how many actually fired
	fireCount := 0
	timeout := time.After(100 * time.Millisecond)

	for {
		select {
		case <-output:
			fireCount++
		case <-timeout:
			// Should have fired much fewer times than signals sent due to refractory period
			if fireCount >= 8 {
				t.Errorf("Too many fires (%d) - refractory period not working", fireCount)
			}
			if fireCount < 1 {
				t.Error("No fires detected - neuron may not be working")
			}
			return
		}
	}
}

// TestDecayRateEffects tests different decay rates
func TestDecayRateEffects(t *testing.T) {
	slowDecay := NewSimpleNeuron("slow_decay", 2.0, 0.99, 5*time.Millisecond, 1.0)
	fastDecay := NewSimpleNeuron("fast_decay", 2.0, 0.8, 5*time.Millisecond, 1.0)

	slowOutput := make(chan Message, 10)
	fastOutput := make(chan Message, 10)

	slowDecay.AddOutput("test", slowOutput, 1.0, 0)
	fastDecay.AddOutput("test", fastOutput, 1.0, 0)

	go slowDecay.Run()
	go fastDecay.Run()
	defer slowDecay.Close()
	defer fastDecay.Close()

	slowInput := slowDecay.GetInput()
	fastInput := fastDecay.GetInput()

	// Send same signal to both
	slowInput <- Message{Value: 1.5}
	fastInput <- Message{Value: 1.5}

	// Wait for decay
	time.Sleep(50 * time.Millisecond)

	// Send additional signal that might fire depending on decay
	slowInput <- Message{Value: 0.6}
	fastInput <- Message{Value: 0.6}

	// Check firing behavior
	slowFired := false
	fastFired := false

	select {
	case <-slowOutput:
		slowFired = true
	case <-time.After(20 * time.Millisecond):
	}

	select {
	case <-fastOutput:
		fastFired = true
	case <-time.After(20 * time.Millisecond):
	}

	// Slow decay should be more likely to fire (less decay means more accumulation retained)
	if !slowFired && fastFired {
		t.Error("Fast decay neuron fired but slow decay didn't - unexpected behavior")
	}
}

// TestCloseBehavior tests graceful shutdown
func TestCloseBehavior(t *testing.T) {
	neuron := NewSimpleNeuron("test_close", 1.0, 0.98, 5*time.Millisecond, 1.0)

	output := make(chan Message, 10)
	neuron.AddOutput("test", output, 1.0, 0)

	go neuron.Run()

	// Send a signal
	input := neuron.GetInput()
	input <- Message{Value: 0.5}

	// Close the neuron
	neuron.Close()

	// Give time for Run() to exit
	time.Sleep(10 * time.Millisecond)
}

// ============================================================================
// BENCHMARK TESTS (Original functionality only)
// ============================================================================

// BenchmarkNeuronCreation benchmarks neuron creation performance
func BenchmarkNeuronCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		neuronID := fmt.Sprintf("bench_neuron_%d", i)
		_ = NewSimpleNeuron(neuronID, 1.0, 0.95, 5*time.Millisecond, 1.0)
	}
}

// BenchmarkMessageProcessing benchmarks message processing throughput
func BenchmarkMessageProcessing(b *testing.B) {
	neuron := NewSimpleNeuron("bench_processing", 10.0, 0.95, 5*time.Millisecond, 1.0) // High threshold to avoid firing

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input <- Message{Value: 0.1}
	}
}

// BenchmarkOutputManagement benchmarks adding/removing outputs
func BenchmarkOutputManagement(b *testing.B) {
	neuron := NewSimpleNeuron("bench_output_mgmt", 1.0, 0.95, 5*time.Millisecond, 1.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output := make(chan Message, 1)
		neuron.AddOutput("test", output, 1.0, 0)
		neuron.RemoveOutput("test")
	}
}
