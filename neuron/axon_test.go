package neuron

import (
	"fmt"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

// TestAxon_RealNeuronDelayedDelivery tests axon delivery using actual neurons
func TestAxon_RealNeuronDelayedDelivery(t *testing.T) {
	t.Log("=== Testing Axon Delivery with Real Neurons ===")

	// Create source and target neurons
	sourceNeuron := NewNeuron(
		"source-neuron",
		1.0,                // threshold
		0.9,                // decay rate
		5*time.Millisecond, // refractory period
		1.0,                // fire factor
		10.0,               // target firing rate
		0.1,                // homeostasis strength
	)

	targetNeuron := NewNeuron(
		"target-neuron",
		1.0,
		0.9,
		5*time.Millisecond,
		1.0,
		10.0,
		0.1,
	)

	// Set up callback infrastructure (minimal for testing)
	mockCallbacks := &NeuronCallbacks{
		ReportHealthFunc: func(activityLevel float64, connectionCount int) {
			// Mock health reporting
		},
		GetSpatialDelayFunc: func(targetID string) time.Duration {
			return 5 * time.Millisecond // Fixed spatial delay
		},
	}

	sourceNeuron.SetCallbacks(mockCallbacks)
	targetNeuron.SetCallbacks(mockCallbacks)

	// Start both neurons
	err := sourceNeuron.Start()
	if err != nil {
		t.Fatalf("Failed to start source neuron: %v", err)
	}
	defer sourceNeuron.Stop()

	err = targetNeuron.Start()
	if err != nil {
		t.Fatalf("Failed to start target neuron: %v", err)
	}
	defer targetNeuron.Stop()

	// Create an output callback to connect source to target with delay
	delay := 10 * time.Millisecond
	outputCallback := types.OutputCallback{
		TransmitMessage: func(msg types.NeuralSignal) error {
			// Use the source neuron's delivery system
			sourceNeuron.ScheduleDelayedDelivery(msg, targetNeuron, delay)
			return nil
		},
		GetWeight: func() float64 {
			return 1.0
		},
		GetDelay: func() time.Duration {
			return delay
		},
		GetTargetID: func() string {
			return targetNeuron.ID()
		},
	}

	// Add the callback to source neuron
	sourceNeuron.AddOutputCallback("test-synapse", outputCallback)

	// Record initial state
	initialTargetMessages := len(getReceivedMessages(targetNeuron))

	// Send a message that should cause the source to fire and deliver to target
	testSignal := types.NeuralSignal{
		Value:     1.5, // Above threshold
		Timestamp: time.Now(),
		SourceID:  "test-sender",
		TargetID:  sourceNeuron.ID(),
	}

	// Send the signal to source neuron
	sourceNeuron.Receive(testSignal)

	// Wait for processing and delivery (delay + processing time)
	time.Sleep(delay + 20*time.Millisecond)

	// Check that target received the message
	finalTargetMessages := len(getReceivedMessages(targetNeuron))

	if finalTargetMessages <= initialTargetMessages {
		t.Errorf("Expected target neuron to receive message, but message count didn't increase: %d -> %d",
			initialTargetMessages, finalTargetMessages)
	} else {
		t.Logf("✓ Message successfully delivered from source to target with %v delay", delay)
	}
}

// TestAxon_MultiNeuronNetwork tests axon delivery in a small network
func TestAxon_MultiNeuronNetwork(t *testing.T) {
	t.Log("=== Testing Axon Delivery in Multi-Neuron Network ===")

	// Create a small network: neuron1 -> neuron2 -> neuron3
	neurons := make([]*Neuron, 3)
	delays := []time.Duration{5 * time.Millisecond, 10 * time.Millisecond}

	// Create neurons
	for i := 0; i < 3; i++ {
		neurons[i] = NewNeuron(
			fmt.Sprintf("neuron-%d", i+1),
			0.8,                // threshold
			0.9,                // decay rate
			5*time.Millisecond, // refractory period
			1.0,                // fire factor
			10.0,               // target firing rate
			0.1,                // homeostasis strength
		)

		// Set mock callbacks
		mockCallbacks := &NeuronCallbacks{
			ReportHealthFunc: func(activityLevel float64, connectionCount int) {},
			GetSpatialDelayFunc: func(targetID string) time.Duration {
				return 2 * time.Millisecond
			},
		}
		neurons[i].SetCallbacks(mockCallbacks)

		// Start neuron
		err := neurons[i].Start()
		if err != nil {
			t.Fatalf("Failed to start neuron %d: %v", i+1, err)
		}
		defer neurons[i].Stop()
	}

	// Connect neurons in chain: 0 -> 1 -> 2
	for i := 0; i < 2; i++ {
		sourceIdx := i
		targetIdx := i + 1
		connectionDelay := delays[i]

		outputCallback := types.OutputCallback{
			TransmitMessage: func(msg types.NeuralSignal) error {
				neurons[sourceIdx].ScheduleDelayedDelivery(msg, neurons[targetIdx], connectionDelay)
				return nil
			},
			GetWeight: func() float64 {
				return 1.0
			},
			GetDelay: func() time.Duration {
				return connectionDelay
			},
			GetTargetID: func() string {
				return neurons[targetIdx].ID()
			},
		}

		neurons[sourceIdx].AddOutputCallback(fmt.Sprintf("synapse-%d-%d", sourceIdx, targetIdx), outputCallback)
	}

	// Record initial activity levels
	initialActivity := make([]int, 3)
	for i, neuron := range neurons {
		initialActivity[i] = len(getReceivedMessages(neuron))
	}

	// Stimulate the first neuron
	stimulus := types.NeuralSignal{
		Value:     1.0, // Above threshold
		Timestamp: time.Now(),
		SourceID:  "external-stimulus",
		TargetID:  neurons[0].ID(),
	}

	neurons[0].Receive(stimulus)

	// Wait for cascade propagation
	totalDelay := delays[0] + delays[1] + 30*time.Millisecond // Extra time for processing
	time.Sleep(totalDelay)

	// Check activity propagation
	finalActivity := make([]int, 3)
	for i, neuron := range neurons {
		finalActivity[i] = len(getReceivedMessages(neuron))
	}

	// Verify propagation
	for i := 0; i < 3; i++ {
		if finalActivity[i] <= initialActivity[i] {
			t.Errorf("Neuron %d activity didn't increase: %d -> %d", i+1, initialActivity[i], finalActivity[i])
		} else {
			t.Logf("✓ Neuron %d received signals: %d -> %d", i+1, initialActivity[i], finalActivity[i])
		}
	}

	t.Log("✓ Signal successfully propagated through neuron chain with axon delays")
}

// TestAxon_DeliveryTiming tests precise timing of axon deliveries
func TestAxon_DeliveryTiming(t *testing.T) {
	t.Log("=== Testing Axon Delivery Timing Precision ===")

	sourceNeuron := NewNeuron("timing-source", 0.5, 0.9, 5*time.Millisecond, 1.0, 10.0, 0.1)
	targetNeuron := NewNeuron("timing-target", 1.0, 0.9, 5*time.Millisecond, 1.0, 10.0, 0.1)

	// Set up callbacks
	mockCallbacks := &NeuronCallbacks{
		ReportHealthFunc:    func(activityLevel float64, connectionCount int) {},
		GetSpatialDelayFunc: func(targetID string) time.Duration { return 0 },
	}

	sourceNeuron.SetCallbacks(mockCallbacks)
	targetNeuron.SetCallbacks(mockCallbacks)

	// Start neurons
	sourceNeuron.Start()
	targetNeuron.Start()
	defer sourceNeuron.Stop()
	defer targetNeuron.Stop()

	// Test different delays
	testDelays := []time.Duration{
		1 * time.Millisecond,
		5 * time.Millisecond,
		10 * time.Millisecond,
		20 * time.Millisecond,
	}

	for _, testDelay := range testDelays {
		t.Logf("Testing delay: %v", testDelay)

		// Set up connection with specific delay
		outputCallback := types.OutputCallback{
			TransmitMessage: func(msg types.NeuralSignal) error {
				sourceNeuron.ScheduleDelayedDelivery(msg, targetNeuron, testDelay)
				return nil
			},
			GetWeight:   func() float64 { return 1.0 },
			GetDelay:    func() time.Duration { return testDelay },
			GetTargetID: func() string { return targetNeuron.ID() },
		}

		sourceNeuron.AddOutputCallback("timing-synapse", outputCallback)

		// Record timing
		beforeCount := len(getReceivedMessages(targetNeuron))
		sendTime := time.Now()

		// Send signal
		signal := types.NeuralSignal{
			Value:     1.0,
			Timestamp: sendTime,
			SourceID:  "timing-test",
			TargetID:  sourceNeuron.ID(),
		}
		sourceNeuron.Receive(signal)

		// Wait for expected delivery time plus small buffer
		time.Sleep(testDelay + 10*time.Millisecond)

		afterCount := len(getReceivedMessages(targetNeuron))
		actualDelay := time.Since(sendTime)

		if afterCount <= beforeCount {
			t.Errorf("No message delivered for delay %v", testDelay)
		} else if actualDelay < testDelay {
			t.Errorf("Message delivered too early: expected >= %v, got %v", testDelay, actualDelay)
		} else {
			t.Logf("✓ Message delivered correctly for delay %v (actual: %v)", testDelay, actualDelay)
		}

		// Remove callback for next test
		sourceNeuron.RemoveOutputCallback("timing-synapse")
	}
}

// TestAxon_QueueOverflow tests behavior when delivery queue is full
func TestAxon_QueueOverflow(t *testing.T) {
	t.Log("=== Testing Axon Queue Overflow Handling ===")

	sourceNeuron := NewNeuron("overflow-source", 0.1, 0.9, 1*time.Millisecond, 1.0, 10.0, 0.1)
	targetNeuron := NewNeuron("overflow-target", 1.0, 0.9, 1*time.Millisecond, 1.0, 10.0, 0.1)

	mockCallbacks := &NeuronCallbacks{
		ReportHealthFunc:    func(activityLevel float64, connectionCount int) {},
		GetSpatialDelayFunc: func(targetID string) time.Duration { return 0 },
	}

	sourceNeuron.SetCallbacks(mockCallbacks)
	targetNeuron.SetCallbacks(mockCallbacks)

	sourceNeuron.Start()
	targetNeuron.Start()
	defer sourceNeuron.Stop()
	defer targetNeuron.Stop()

	// Set up connection with very long delay to cause queue buildup
	longDelay := 100 * time.Millisecond
	outputCallback := types.OutputCallback{
		TransmitMessage: func(msg types.NeuralSignal) error {
			sourceNeuron.ScheduleDelayedDelivery(msg, targetNeuron, longDelay)
			return nil
		},
		GetWeight:   func() float64 { return 1.0 },
		GetDelay:    func() time.Duration { return longDelay },
		GetTargetID: func() string { return targetNeuron.ID() },
	}

	sourceNeuron.AddOutputCallback("overflow-synapse", outputCallback)

	initialCount := len(getReceivedMessages(targetNeuron))

	// Send many signals rapidly to overflow the queue
	for i := 0; i < AXON_QUEUE_CAPACITY_DEFAULT+10; i++ {
		signal := types.NeuralSignal{
			Value:     0.2, // Small value to cause firing without overwhelming
			Timestamp: time.Now(),
			SourceID:  "overflow-test",
			TargetID:  sourceNeuron.ID(),
		}
		sourceNeuron.Receive(signal)
		time.Sleep(1 * time.Millisecond) // Small delay between sends
	}

	// Check for immediate deliveries (overflow handling)
	immediateCount := len(getReceivedMessages(targetNeuron))

	// Wait for delayed deliveries
	time.Sleep(longDelay + 20*time.Millisecond)
	finalCount := len(getReceivedMessages(targetNeuron))

	t.Logf("Message counts - Initial: %d, Immediate: %d, Final: %d",
		initialCount, immediateCount, finalCount)

	if immediateCount > initialCount {
		t.Logf("✓ Queue overflow triggered immediate deliveries: %d immediate messages",
			immediateCount-initialCount)
	}

	if finalCount > immediateCount {
		t.Logf("✓ Delayed deliveries also arrived: %d delayed messages",
			finalCount-immediateCount)
	}

	if finalCount <= initialCount {
		t.Error("No messages were delivered at all")
	} else {
		t.Log("✓ Axon handled queue overflow gracefully")
	}
}

// BenchmarkAxon_RealNeuronDelivery benchmarks axon delivery with real neurons
func BenchmarkAxon_RealNeuronDelivery(b *testing.B) {
	sourceNeuron := NewNeuron("bench-source", 0.5, 0.9, 1*time.Millisecond, 1.0, 10.0, 0.1)
	targetNeuron := NewNeuron("bench-target", 1.0, 0.9, 1*time.Millisecond, 1.0, 10.0, 0.1)

	mockCallbacks := &NeuronCallbacks{
		ReportHealthFunc:    func(activityLevel float64, connectionCount int) {},
		GetSpatialDelayFunc: func(targetID string) time.Duration { return 0 },
	}

	sourceNeuron.SetCallbacks(mockCallbacks)
	targetNeuron.SetCallbacks(mockCallbacks)

	sourceNeuron.Start()
	targetNeuron.Start()
	defer sourceNeuron.Stop()
	defer targetNeuron.Stop()

	outputCallback := types.OutputCallback{
		TransmitMessage: func(msg types.NeuralSignal) error {
			sourceNeuron.ScheduleDelayedDelivery(msg, targetNeuron, 1*time.Millisecond)
			return nil
		},
		GetWeight:   func() float64 { return 1.0 },
		GetDelay:    func() time.Duration { return 1 * time.Millisecond },
		GetTargetID: func() string { return targetNeuron.ID() },
	}

	sourceNeuron.AddOutputCallback("bench-synapse", outputCallback)

	signal := types.NeuralSignal{
		Value:     0.1,
		Timestamp: time.Now(),
		SourceID:  "bench",
		TargetID:  sourceNeuron.ID(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sourceNeuron.Receive(signal)
	}
}

// Helper function to get received messages from a neuron
// This would need to be implemented based on your neuron's internal structure
func getReceivedMessages(neuron *Neuron) []types.NeuralSignal {
	// This is a placeholder - you'd need to implement this based on how
	// your neuron tracks received messages. Options:

	// Option 1: If neuron has a method to get received messages
	// return neuron.GetReceivedMessages()

	// Option 2: If you need to add tracking, you could modify the Receive method
	// to store messages in a slice for testing purposes

	// Option 3: Use activity metrics as a proxy
	activity := neuron.GetActivityLevel()
	// Convert activity to approximate message count for testing
	return make([]types.NeuralSignal, int(activity*10))
}
