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
	timeWindow := 100 * time.Millisecond
	fireFactor := 2.0
	neuronID := "test_neuron_1"

	neuron := NewNeuron(neuronID, threshold, timeWindow, fireFactor)

	if neuron == nil {
		t.Fatal("NewNeuron returned nil")
	}

	if neuron.threshold != threshold {
		t.Errorf("Expected threshold %f, got %f", threshold, neuron.threshold)
	}

	if neuron.timeWindow != timeWindow {
		t.Errorf("Expected timeWindow %v, got %v", timeWindow, neuron.timeWindow)
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
}

// TestNeuronInputChannel tests that input channel is accessible
func TestNeuronInputChannel(t *testing.T) {
	neuron := NewNeuron("test_input", 1.0, 50*time.Millisecond, 1.0)

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
	neuron := NewNeuron("test_output_mgmt", 1.0, 50*time.Millisecond, 1.0)

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
	neuron := NewNeuron("test_threshold", threshold, 100*time.Millisecond, 1.0)

	output := make(chan Message, 10)
	neuron.AddOutput("test", output, 1.0, 0) // No delay for testing

	// Start neuron processing
	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// Send signal below threshold - should not fire
	input <- Message{Value: 0.5}

	// Wait a bit and check no output
	select {
	case <-output:
		t.Error("Neuron fired when signal was below threshold")
	case <-time.After(20 * time.Millisecond):
		// Expected - no firing
	}

	// Send signal that brings total above threshold
	input <- Message{Value: 0.6} // Total: 1.1 > 1.0

	// Should fire now
	select {
	case fired := <-output:
		if fired.Value <= 0 {
			t.Errorf("Expected positive fire value, got %f", fired.Value)
		}
	case <-time.After(50 * time.Millisecond):
		t.Error("Neuron did not fire when threshold was exceeded")
	}
}

// TestTemporalIntegration tests signal accumulation over time windows
func TestTemporalIntegration(t *testing.T) {
	timeWindow := 50 * time.Millisecond
	neuron := NewNeuron("test_temporal", 1.0, timeWindow, 1.0)

	output := make(chan Message, 10)
	neuron.AddOutput("test", output, 1.0, 0)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// Send signals within time window
	input <- Message{Value: 0.4}
	time.Sleep(10 * time.Millisecond)
	input <- Message{Value: 0.3}
	time.Sleep(10 * time.Millisecond)
	input <- Message{Value: 0.4} // Total: 1.1 > 1.0

	// Should fire
	select {
	case <-output:
		// Expected
	case <-time.After(30 * time.Millisecond):
		t.Error("Neuron did not fire with temporal integration")
	}
}

// TestTimeWindowExpiry tests that accumulator resets after time window
func TestTimeWindowExpiry(t *testing.T) {
	timeWindow := 30 * time.Millisecond
	neuron := NewNeuron("test_window_expiry", 1.0, timeWindow, 1.0)

	output := make(chan Message, 10)
	neuron.AddOutput("test", output, 1.0, 0)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// Send signal below threshold
	input <- Message{Value: 0.7}

	// Wait for time window to expire
	time.Sleep(timeWindow + 10*time.Millisecond)

	// Send another signal - should start fresh accumulation
	input <- Message{Value: 0.8} // Below threshold alone

	// Should not fire (previous 0.7 should be forgotten)
	select {
	case <-output:
		t.Error("Neuron fired when it should have reset accumulator")
	case <-time.After(20 * time.Millisecond):
		// Expected - no firing
	}
}

// TestOutputFactorAndDelay tests output scaling and transmission delays
func TestOutputFactorAndDelay(t *testing.T) {
	neuron := NewNeuron("test_factor_delay", 1.0, 50*time.Millisecond, 2.0) // fireFactor = 2.0

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

		// Check delay with more generous tolerance for system overhead
		// Real systems have goroutine scheduling, channel operations, etc.
		minDelay := delay
		maxDelay := delay + 50*time.Millisecond // More realistic tolerance

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

	case <-time.After(delay + 100*time.Millisecond): // Generous timeout
		t.Error("Neuron did not fire within expected time")
	}
}

// TestMultipleOutputs tests firing to multiple outputs simultaneously
func TestMultipleOutputs(t *testing.T) {
	neuron := NewNeuron("test_multiple_outputs", 1.0, 50*time.Millisecond, 1.0)

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
	neuron := NewNeuron("test_concurrent", 1.0, 50*time.Millisecond, 1.0)

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
	neuron := NewNeuron("test_reset", 1.0, 100*time.Millisecond, 1.0)

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

	// Send another signal - should start from 0, not 1.5
	input <- Message{Value: 0.8} // Below threshold

	// Should not fire (proves reset occurred)
	select {
	case <-output:
		t.Error("Neuron fired when it should have reset after previous firing")
	case <-time.After(30 * time.Millisecond):
		// Expected - no firing
	}

	// Now send enough to fire again
	input <- Message{Value: 0.3} // Total: 0.8 + 0.3 = 1.1 > 1.0

	select {
	case <-output:
		// Expected second firing
	case <-time.After(20 * time.Millisecond):
		t.Error("Second firing did not occur")
	}
}

// TestInhibitorySignals tests negative (inhibitory) input values
func TestInhibitorySignals(t *testing.T) {
	neuron := NewNeuron("test_inhibitory", 1.0, 100*time.Millisecond, 1.0)

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
	neuron := NewNeuron("test_fire_events", 1.0, 50*time.Millisecond, 2.0)

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

// TestCloseBehavior tests graceful shutdown
func TestCloseBehavior(t *testing.T) {
	neuron := NewNeuron("test_close", 1.0, 50*time.Millisecond, 1.0)

	output := make(chan Message, 10)
	neuron.AddOutput("test", output, 1.0, 0)

	go neuron.Run()

	// Send a signal
	input := neuron.GetInput()
	input <- Message{Value: 0.5}

	// Close the neuron
	neuron.Close()

	// Try to send another signal (should not panic, but channel is closed)
	// This tests that the Run() loop exits gracefully
	time.Sleep(10 * time.Millisecond) // Give time for Run() to exit
}

// BenchmarkNeuronCreation benchmarks neuron creation performance
func BenchmarkNeuronCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		neuronID := fmt.Sprintf("bench_neuron_%d", i)
		_ = NewNeuron(neuronID, 1.0, 50*time.Millisecond, 1.0)
	}
}

// BenchmarkMessageProcessing benchmarks message processing throughput
func BenchmarkMessageProcessing(b *testing.B) {
	neuron := NewNeuron("bench_processing", 10.0, 100*time.Millisecond, 1.0) // High threshold to avoid firing

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
	neuron := NewNeuron("bench_output_mgmt", 1.0, 50*time.Millisecond, 1.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		output := make(chan Message, 1)
		neuron.AddOutput("test", output, 1.0, 0)
		neuron.RemoveOutput("test")
	}
}
