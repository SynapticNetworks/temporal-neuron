package neuron

import (
	"sync"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/message"
)

// MockMessageReceiver is a mock implementation of the MessageReceiver interface for testing.
type MockMessageReceiver struct {
	id           string
	receivedMsgs []message.NeuralSignal
	mutex        sync.Mutex // To protect receivedMsgs in concurrent scenarios
}

// NewMockMessageReceiver creates a new MockMessageReceiver.
func NewMockMessageReceiver(id string) *MockMessageReceiver {
	return &MockMessageReceiver{
		id:           id,
		receivedMsgs: make([]message.NeuralSignal, 0),
	}
}

// Receive implements the MessageReceiver interface.
func (m *MockMessageReceiver) Receive(msg message.NeuralSignal) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.receivedMsgs = append(m.receivedMsgs, msg)
}

// ID implements the MessageReceiver interface.
func (m *MockMessageReceiver) ID() string {
	return m.id
}

// GetReceivedMessages returns a copy of the received messages.
func (m *MockMessageReceiver) GetReceivedMessages() []message.NeuralSignal {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	copied := make([]message.NeuralSignal, len(m.receivedMsgs))
	copy(copied, m.receivedMsgs)
	return copied
}

// ClearReceivedMessages clears the list of received messages.
func (m *MockMessageReceiver) ClearReceivedMessages() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.receivedMsgs = m.receivedMsgs[:0]
}

// ============================================================================
// FIXED TESTS
// ============================================================================

// TestAxon_ScheduleDelayedDelivery validates that messages are correctly queued.
func TestAxon_ScheduleDelayedDelivery(t *testing.T) {
	t.Log("=== Testing ScheduleDelayedDelivery: Message Queuing ===")

	deliveryQueue := make(chan delayedMessage, 5) // Small buffer for easy testing
	defer close(deliveryQueue)

	mockTarget := NewMockMessageReceiver("target1")
	testDelay := 10 * time.Millisecond
	testValue := 1.0

	msg := message.NeuralSignal{
		Value:     testValue,
		Timestamp: time.Now(),
		SourceID:  "source1",
		TargetID:  "target1",
	}

	// Schedule a message
	ScheduleDelayedDelivery(deliveryQueue, msg, mockTarget, testDelay)

	// Verify the message is in the queue
	select {
	case qMsg := <-deliveryQueue:
		if qMsg.message.Value != testValue {
			t.Errorf("Expected message value %f, got %f", testValue, qMsg.message.Value)
		}
		if qMsg.target.ID() != mockTarget.ID() {
			t.Errorf("Expected target ID %s, got %s", mockTarget.ID(), qMsg.target.ID())
		}
		if qMsg.deliveryTime.IsZero() {
			t.Error("Delivery time not set")
		}
		t.Logf("✓ Message successfully queued with value %f and delay %v", testValue, testDelay)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Message not queued within timeout")
	}

	// Test queue full fallback
	t.Log("Testing queue full fallback (immediate delivery)...")
	for i := 0; i < cap(deliveryQueue)+1; i++ { // Fill and then one more
		ScheduleDelayedDelivery(deliveryQueue, msg, mockTarget, testDelay)
	}

	// The last message should be delivered immediately if the queue is full
	receivedMsgs := mockTarget.GetReceivedMessages()
	if len(receivedMsgs) == 0 {
		t.Error("Expected immediate delivery for queue full, but no message received by target")
	} else {
		t.Log("✓ Immediate delivery fallback successful when queue is full")
	}
}

// TestAxon_ProcessAxonDeliveries_TimingAndDispatch validates correct dispatching based on time.
func TestAxon_ProcessAxonDeliveries_TimingAndDispatch(t *testing.T) {
	t.Log("=== Testing ProcessAxonDeliveries: Timing and Dispatching ===")

	deliveryQueue := make(chan delayedMessage, 10) // Channel for new incoming messages
	mockTarget := NewMockMessageReceiver("target2")

	// Store initial messages to queue later to precisely control when they arrive in the main loop's processing
	// We'll queue these at specific time points during the test's progression
	messagesToQueueAtSpecificTimes := []struct {
		value       float64
		relativeDue time.Duration // Due relative to the start of the test
	}{
		{4.0, -5 * time.Millisecond},  // Already past due (should dispatch immediately)
		{2.0, 10 * time.Millisecond},  // Due earliest
		{5.0, 20 * time.Millisecond},  // Due next
		{1.0, 50 * time.Millisecond},  // Due after that
		{3.0, 100 * time.Millisecond}, // Due last
	}

	testStartTime := time.Now()
	pendingDeliveries := make([]delayedMessage, 0, 10)

	// Helper to queue a message with a specific relative delivery time
	queueMessage := func(value float64, relativeDue time.Duration) {
		//deliveryTime := testStartTime.Add(relativeDue)
		// ScheduleDelayedDelivery handles the `time.Now()` part, so we pass the relative delay
		ScheduleDelayedDelivery(deliveryQueue, message.NeuralSignal{Value: value, SourceID: "s", TargetID: mockTarget.ID(), Timestamp: testStartTime}, mockTarget, relativeDue)
	}

	// --- Test Step Functions ---
	checkAndClear := func(stepName string, expectedValue float64, totalElapsedTime time.Duration) {
		t.Logf("%s: Processing at approx %v...", stepName, totalElapsedTime)
		mockTarget.ClearReceivedMessages() // Always clear before checking for new messages

		// Simulate the main loop calling ProcessAxonDeliveries
		pendingDeliveries = ProcessAxonDeliveries(pendingDeliveries, deliveryQueue, testStartTime.Add(totalElapsedTime))
		received := mockTarget.GetReceivedMessages()

		if expectedValue != 0 && (len(received) != 1 || received[0].Value != expectedValue) {
			t.Errorf("FAIL %s: Expected 1 message (value %.1f), got %d messages: %v", stepName, expectedValue, len(received), received)
		} else if expectedValue == 0 && len(received) != 0 {
			t.Errorf("FAIL %s: Expected 0 messages, got %d messages: %v", stepName, len(received), received)
		} else if expectedValue != 0 {
			t.Logf("✓ %s: Message %.1f dispatched correctly.", stepName, expectedValue)
		} else {
			t.Logf("✓ %s: No messages dispatched (as expected).", stepName)
		}
	}

	// --- Execution Steps ---

	// Step 1: Queue all initial messages and process immediately.
	// Only 4.0 should be dispatched (due -5ms from start)
	t.Log("Step 1: Queueing all messages and processing immediately...")
	for _, m := range messagesToQueueAtSpecificTimes {
		queueMessage(m.value, m.relativeDue)
	}
	checkAndClear("Step 1 (Immediate)", 4.0, 0*time.Millisecond) // Process at testStartTime

	// Step 2: Advance time to 12ms from test start. Message 2.0 (due at 10ms) should be dispatched.
	checkAndClear("Step 2 (12ms elapsed)", 2.0, 12*time.Millisecond)

	// Step 3: Advance time to 22ms from test start. Message 5.0 (due at 20ms) should be dispatched.
	checkAndClear("Step 3 (22ms elapsed)", 5.0, 22*time.Millisecond)

	// Step 4: Advance time to 52ms from test start. Message 1.0 (due at 50ms) should be dispatched.
	checkAndClear("Step 4 (52ms elapsed)", 1.0, 52*time.Millisecond)

	// Step 5: Advance time to 102ms from test start. Message 3.0 (due at 100ms) should be dispatched.
	checkAndClear("Step 5 (102ms elapsed)", 3.0, 102*time.Millisecond)

	// Final check: All messages should have been dispatched.
	t.Logf("Remaining pending deliveries: %d", len(pendingDeliveries))
	if len(pendingDeliveries) != 0 {
		t.Errorf("FAIL: Expected no pending deliveries, but got %d", len(pendingDeliveries))
	} else {
		t.Log("✓ All messages dispatched successfully.")
	}
}

// TestAxon_ProcessAxonDeliveries_Sorting ensures pending messages are sorted correctly.
func TestAxon_ProcessAxonDeliveries_Sorting(t *testing.T) {
	t.Log("=== Testing ProcessAxonDeliveries: Sorting Logic ===")

	// Create a dummy target for the delayedMessage, as `target` cannot be nil
	dummyTarget := NewMockMessageReceiver("dummy")
	now := time.Now()

	// Simulating direct manipulation of pending slice for sorting test
	pending := []delayedMessage{
		{deliveryTime: now.Add(50 * time.Millisecond), message: message.NeuralSignal{Value: 3.0}, target: dummyTarget},
		{deliveryTime: now.Add(10 * time.Millisecond), message: message.NeuralSignal{Value: 1.0}, target: dummyTarget},
		{deliveryTime: now.Add(100 * time.Millisecond), message: message.NeuralSignal{Value: 5.0}, target: dummyTarget},
		{deliveryTime: now.Add(20 * time.Millisecond), message: message.NeuralSignal{Value: 2.0}, target: dummyTarget},
	}

	// Use a mock delivery queue for ProcessAxonDeliveries signature, it won't be used
	mockQueue := make(chan delayedMessage, 1)
	defer close(mockQueue)

	// FIXED: Pass a time BEFORE all delivery times so no messages are dispatched
	// This tests only the sorting logic
	pastTime := now.Add(-1 * time.Hour) // One hour before the messages were created
	sortedPending := ProcessAxonDeliveries(pending, mockQueue, pastTime)

	if len(sortedPending) != 4 {
		t.Fatalf("Expected 4 messages after sorting, got %d", len(sortedPending))
	}

	// Verify the order by deliveryTime
	for i := 0; i < len(sortedPending)-1; i++ {
		if sortedPending[i].deliveryTime.After(sortedPending[i+1].deliveryTime) {
			t.Errorf("Messages are not sorted correctly: %v is after %v",
				sortedPending[i].message.Value, sortedPending[i+1].message.Value)
		}
	}

	// Verify correct order of values
	expectedOrder := []float64{1.0, 2.0, 3.0, 5.0}
	for i, expected := range expectedOrder {
		if sortedPending[i].message.Value != expected {
			t.Errorf("Message %d: expected value %.1f, got %.1f", i, expected, sortedPending[i].message.Value)
		}
	}

	t.Log("✓ Pending deliveries correctly sorted by delivery time.")
}

// TestAxon_ProcessAxonDeliveries_EmptyQueueAndSlice handles edge cases.
func TestAxon_ProcessAxonDeliveries_EmptyQueueAndSlice(t *testing.T) {
	t.Log("=== Testing ProcessAxonDeliveries: Empty Queue and Slice ===")

	deliveryQueue := make(chan delayedMessage, 10) // Larger buffer to prevent blocking
	defer close(deliveryQueue)

	mockTarget := NewMockMessageReceiver("target3")
	now := time.Now()

	// Case 1: Empty pending slice, empty delivery queue
	t.Log("Case 1: Empty pending slice, empty delivery queue.")
	pending := make([]delayedMessage, 0)
	updatedPending := ProcessAxonDeliveries(pending, deliveryQueue, now)

	if len(updatedPending) != 0 {
		t.Errorf("Expected empty pending slice, got %d", len(updatedPending))
	}
	if len(mockTarget.GetReceivedMessages()) != 0 {
		t.Errorf("Expected no messages received, got %d", len(mockTarget.GetReceivedMessages()))
	}
	t.Log("✓ Handles empty inputs gracefully.")

	// Case 2: Messages in delivery queue, empty pending slice
	t.Log("Case 2: Messages in delivery queue, empty pending slice.")

	// FIXED: Create messages manually with precise delivery times
	baseTime := time.Now()
	earlyDeliveryTime := baseTime.Add(5 * time.Millisecond) // Should be delivered
	lateDeliveryTime := baseTime.Add(50 * time.Millisecond) // Should remain pending

	// Create delayed messages manually with precise timing control
	earlyMessage := delayedMessage{
		message: message.NeuralSignal{
			Value:    10.0,
			SourceID: "s",
			TargetID: mockTarget.ID(),
		},
		target:       mockTarget,
		deliveryTime: earlyDeliveryTime,
	}

	lateMessage := delayedMessage{
		message: message.NeuralSignal{
			Value:    20.0,
			SourceID: "s",
			TargetID: mockTarget.ID(),
		},
		target:       mockTarget,
		deliveryTime: lateDeliveryTime,
	}

	// Add messages directly to queue (bypassing ScheduleDelayedDelivery timing issues)
	deliveryQueue <- earlyMessage
	deliveryQueue <- lateMessage

	// Process at a time when only the early message should be delivered
	processTime := baseTime.Add(10 * time.Millisecond) // 10ms after base, early (5ms) should be delivered, late (50ms) should remain
	updatedPending = ProcessAxonDeliveries(make([]delayedMessage, 0), deliveryQueue, processTime)

	received := mockTarget.GetReceivedMessages()
	if len(received) != 1 || received[0].Value != 10.0 {
		t.Errorf("Expected 1 message (10.0) received, got %d messages: %v", len(received), received)
	}
	if len(updatedPending) != 1 || updatedPending[0].message.Value != 20.0 {
		t.Errorf("Expected 1 message (20.0) in pending, got %d messages with values: %v", len(updatedPending),
			func() []float64 {
				var vals []float64
				for _, msg := range updatedPending {
					vals = append(vals, msg.message.Value)
				}
				return vals
			}())
	}
	t.Log("✓ Correctly drains new messages and manages pending.")
}

// BenchmarkAxon_ScheduleDelayedDelivery benchmarks the message queuing performance.
func BenchmarkAxon_ScheduleDelayedDelivery(b *testing.B) {
	deliveryQueue := make(chan delayedMessage, AXON_QUEUE_CAPACITY_DEFAULT)
	mockTarget := NewMockMessageReceiver("bench_target")
	msg := message.NeuralSignal{
		Value:     1.0,
		Timestamp: time.Now(),
		SourceID:  "bench_source",
		TargetID:  mockTarget.ID(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ScheduleDelayedDelivery(deliveryQueue, msg, mockTarget, 1*time.Millisecond)
	}
	// Drain the queue to prevent blocking subsequent runs or test phases
	for len(deliveryQueue) > 0 {
		<-deliveryQueue
	}
}

// BenchmarkAxon_ProcessAxonDeliveries benchmarks the dispatching performance.
func BenchmarkAxon_ProcessAxonDeliveries(b *testing.B) {
	// Setup a fixed number of pending messages for a consistent benchmark.
	// This simulates a scenario where the axon queue has items to process.
	basePendingCount := 100
	deliveryQueue := make(chan delayedMessage, basePendingCount)
	mockTarget := NewMockMessageReceiver("bench_target")

	// Fill the queue once before starting the timer.
	for i := 0; i < basePendingCount; i++ {
		// Distribute delivery times to ensure sorting and dispatch logic runs.
		delay := time.Duration(i) * time.Microsecond
		msg := message.NeuralSignal{Value: float64(i), Timestamp: time.Now(), SourceID: "s", TargetID: mockTarget.ID()}
		deliveryQueue <- delayedMessage{message: msg, target: mockTarget, deliveryTime: time.Now().Add(delay)}
	}

	// Prepare initial pending slice from the filled queue
	initialPending := make([]delayedMessage, 0, basePendingCount)
	for len(deliveryQueue) > 0 {
		initialPending = append(initialPending, <-deliveryQueue)
	}

	// Reset timer for the core logic.
	b.ResetTimer()
	b.ReportAllocs() // Report memory allocations

	var currentPending []delayedMessage
	// Run the benchmark for N iterations
	for i := 0; i < b.N; i++ {
		// Reset pending to the initial set for each iteration to keep workload consistent.
		// A real scenario would involve new messages arriving and old ones leaving.
		// For benchmarking, we want a repeatable, representative workload.
		// Deep copy to ensure sorting in one iteration doesn't affect the next.
		tempPending := make([]delayedMessage, len(initialPending))
		copy(tempPending, initialPending)

		// Pass `now` as a parameter to avoid `time.Now()` overhead inside the benchmark loop
		// and to make the benchmark deterministic. Assuming most messages are due.
		currentPending = ProcessAxonDeliveries(tempPending, make(chan delayedMessage), time.Now().Add(AXON_DELAY_MAX_BIOLOGICAL_PROPAGATION))
	}
	// To prevent compiler optimizations from removing the call.
	_ = currentPending
}

// TestAxon_DeliveryOrder ensures messages are dispatched in chronological order.
func TestAxon_DeliveryOrder(t *testing.T) {
	t.Log("=== Testing Axon Delivery Order: Chronological Dispatch ===")

	deliveryQueue := make(chan delayedMessage, 5)
	defer close(deliveryQueue)
	mockTarget := NewMockMessageReceiver("target_order")

	// Queue messages out of chronological order of deliveryTime
	ScheduleDelayedDelivery(deliveryQueue, message.NeuralSignal{Value: 30, SourceID: "s1", TargetID: mockTarget.ID()}, mockTarget, 30*time.Millisecond)
	ScheduleDelayedDelivery(deliveryQueue, message.NeuralSignal{Value: 10, SourceID: "s2", TargetID: mockTarget.ID()}, mockTarget, 10*time.Millisecond)
	ScheduleDelayedDelivery(deliveryQueue, message.NeuralSignal{Value: 50, SourceID: "s3", TargetID: mockTarget.ID()}, mockTarget, 50*time.Millisecond)
	ScheduleDelayedDelivery(deliveryQueue, message.NeuralSignal{Value: 20, SourceID: "s4", TargetID: mockTarget.ID()}, mockTarget, 20*time.Millisecond)

	pendingDeliveries := make([]delayedMessage, 0, 5)

	// Simulate neuron's Run loop over time
	currentTime := time.Now()

	// Process multiple times to allow messages to become due and be dispatched
	for i := 0; i < 6; i++ { // Run enough iterations to ensure all are dispatched
		// Advance time for each iteration
		currentTime = currentTime.Add(15 * time.Millisecond)
		t.Logf("Processing at %v...", currentTime.Format("15:04:05.000"))

		pendingDeliveries = ProcessAxonDeliveries(pendingDeliveries, deliveryQueue, currentTime)

		// Log received messages in this tick
		receivedInTick := mockTarget.GetReceivedMessages()
		if len(receivedInTick) > 0 {
			t.Logf("  Received this tick: %v", receivedInTick)
			mockTarget.ClearReceivedMessages()
		}
	}

	//finalReceived := mockTarget.GetReceivedMessages() // Should be empty from previous clear, but here to confirm
	// We need to collect all messages and then check their overall chronological order
	// A better approach for this test is to let all messages collect and then check order.

	// Re-run the test to capture total order, without clearing in between
	t.Log("\n--- Re-running to capture total chronological order ---")
	mockTarget.ClearReceivedMessages()           // Ensure clean slate
	deliveryQueue = make(chan delayedMessage, 5) // Re-initialize queue

	ScheduleDelayedDelivery(deliveryQueue, message.NeuralSignal{Value: 30, SourceID: "s1", TargetID: mockTarget.ID()}, mockTarget, 30*time.Millisecond)
	ScheduleDelayedDelivery(deliveryQueue, message.NeuralSignal{Value: 10, SourceID: "s2", TargetID: mockTarget.ID()}, mockTarget, 10*time.Millisecond)
	ScheduleDelayedDelivery(deliveryQueue, message.NeuralSignal{Value: 50, SourceID: "s3", TargetID: mockTarget.ID()}, mockTarget, 50*time.Millisecond)
	ScheduleDelayedDelivery(deliveryQueue, message.NeuralSignal{Value: 20, SourceID: "s4", TargetID: mockTarget.ID()}, mockTarget, 20*time.Millisecond)

	// Collect all messages into `pendingDeliveries` across multiple calls to ensure all are in the slice
	pendingDeliveries = make([]delayedMessage, 0, 5)
	for len(deliveryQueue) > 0 { // Drain initial queue content
		pendingDeliveries = append(pendingDeliveries, <-deliveryQueue)
	}

	// Now, simulate sufficient time passing and process all messages at once
	allDispatchedTime := time.Now().Add(60 * time.Millisecond) // Time after all messages should be due
	remainingInPending := ProcessAxonDeliveries(pendingDeliveries, deliveryQueue, allDispatchedTime)

	if len(remainingInPending) != 0 {
		t.Fatalf("Expected all messages to be dispatched, but %d remain.", len(remainingInPending))
	}

	finalReceivedMsgs := mockTarget.GetReceivedMessages()
	if len(finalReceivedMsgs) != 4 {
		t.Fatalf("Expected 4 messages received in total, got %d", len(finalReceivedMsgs))
	}

	// Check values and their chronological order
	expectedOrder := []float64{10, 20, 30, 50}
	for i, msg := range finalReceivedMsgs {
		if msg.Value != expectedOrder[i] {
			t.Errorf("Message at index %d has value %f, expected %f. Order: %v", i, msg.Value, expectedOrder[i], finalReceivedMsgs)
		}
	}
	t.Log("✓ Messages dispatched in correct chronological order.")
}
