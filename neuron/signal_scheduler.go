package neuron

import (
	"container/heap"
	"sync"
	"sync/atomic"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/synapse"
)

// =================================================================================
// SIGNAL SCHEDULING INFRASTRUCTURE
// =================================================================================

// SignalQueue implements a priority queue for efficient signal scheduling
// Uses Go's container/heap interface for O(log n) insertions and O(log n) removals
//
// PERFORMANCE CHARACTERISTICS:
// - Insert: O(log n) time complexity
// - Remove earliest: O(log n) time complexity
// - Peek earliest: O(1) time complexity
// - Memory: O(n) space complexity with minimal overhead
//
// BIOLOGICAL JUSTIFICATION:
// Real axons maintain precise timing of signal propagation. This queue models
// the axon's ability to maintain multiple signals "in flight" simultaneously
// while ensuring they arrive in the correct temporal order.
type SignalQueue []*synapse.ScheduledSignal

// Len returns the number of signals in the queue
// Required by heap.Interface
func (sq SignalQueue) Len() int {
	return len(sq)
}

// Less compares two signals by delivery time for priority ordering
// Earlier delivery times have higher priority (come first in queue)
// Required by heap.Interface
//
// BIOLOGICAL TIMING:
// Signals are delivered in temporal order, just like real axon propagation
// where signals fired earlier arrive earlier (assuming similar path lengths)
func (sq SignalQueue) Less(i, j int) bool {
	// Primary sort: by delivery time (earlier times first)
	if !sq[i].DeliveryTime.Equal(sq[j].DeliveryTime) {
		return sq[i].DeliveryTime.Before(sq[j].DeliveryTime)
	}

	// Secondary sort: by priority for simultaneous delivery times
	// Higher priority values are delivered first
	return sq[i].Priority > sq[j].Priority
}

// Swap exchanges two signals in the queue
// Required by heap.Interface
func (sq SignalQueue) Swap(i, j int) {
	sq[i], sq[j] = sq[j], sq[i]
}

// Push adds a signal to the queue
// Required by heap.Interface - called by heap.Push()
func (sq *SignalQueue) Push(x interface{}) {
	*sq = append(*sq, x.(*synapse.ScheduledSignal))
}

// Pop removes and returns the highest priority signal
// Required by heap.Interface - called by heap.Pop()
func (sq *SignalQueue) Pop() interface{} {
	old := *sq
	n := len(old)
	item := old[n-1]
	*sq = old[0 : n-1]
	return item
}

// =================================================================================
// NEW: SIGNAL SCHEDULER - MANAGES NEURON'S OUTGOING SIGNALS
// =================================================================================

// SignalScheduler manages the outgoing signal delivery queue for a single neuron
// This replaces the distributed time.AfterFunc approach with centralized scheduling
//
// DESIGN PRINCIPLES:
// 1. One scheduler per neuron (matches biological reality of one axon per neuron)
// 2. Uses neuron's existing ticker for timing (no additional goroutines)
// 3. Bounded queue size prevents memory leaks during extreme activity
// 4. Thread-safe for concurrent access from firing and delivery processes
//
// BIOLOGICAL ANALOGY:
// This models the axon's role in managing multiple signals propagating toward
// different target neurons with different delays. Real axons can have hundreds
// of signals "in flight" simultaneously.
type SignalScheduler struct {
	// queue holds all scheduled signals sorted by delivery time
	queue SignalQueue

	// queueMutex protects concurrent access to the queue
	// RWMutex allows multiple readers for statistics while ensuring exclusive writes
	queueMutex sync.RWMutex

	// maxQueueSize prevents memory leaks from pathological firing patterns
	// If queue fills up, oldest signals may be dropped (biological overflow behavior)
	maxQueueSize int

	// Statistics for monitoring and debugging (atomic for lock-free reads)
	totalScheduled  int64 // Total signals ever scheduled
	totalDelivered  int64 // Total signals ever delivered
	totalDropped    int64 // Total signals dropped due to queue overflow
	deliveryLatency int64 // Average delivery latency in nanoseconds
}

// NewSignalScheduler creates a new signal scheduler for a neuron
//
// PARAMETERS:
// maxSize: Maximum number of signals that can be queued simultaneously
//
//	Typical values: 1000-10000 depending on expected firing patterns
//
// BIOLOGICAL REASONING:
// Real axons have physical limits on how many signals can be "in flight"
// simultaneously. This limit prevents pathological states where a neuron
// fires so rapidly that memory is exhausted.
func NewSignalScheduler(maxSize int) *SignalScheduler {
	// Validate maxSize parameter
	if maxSize <= 0 {
		maxSize = 1000 // Conservative default
	}

	scheduler := &SignalScheduler{
		queue:        make(SignalQueue, 0, 64), // Pre-allocate reasonable initial capacity
		maxQueueSize: maxSize,
		// Statistics initialized to zero by Go's zero values
	}

	// Initialize the heap data structure
	heap.Init(&scheduler.queue)

	return scheduler
}

// ScheduleSignal adds a signal to the delivery queue
//
// RETURNS:
// true if signal was successfully queued
// false if queue is full (signal dropped to prevent memory exhaustion)
//
// BIOLOGICAL OVERFLOW BEHAVIOR:
// Real axons can become "saturated" during extreme activity. When this happens,
// some signals may be lost. This models that biological limitation.
//
// THREAD SAFETY:
// This method is safe for concurrent calls from multiple goroutines
func (ss *SignalScheduler) ScheduleSignal(signal *synapse.ScheduledSignal) bool {
	// Validate input signal
	if signal == nil {
		return false
	}

	ss.queueMutex.Lock()
	defer ss.queueMutex.Unlock()

	// Check for queue overflow (prevent memory leaks)
	if len(ss.queue) >= ss.maxQueueSize {
		// Queue is full - drop this signal and record the drop
		atomic.AddInt64(&ss.totalDropped, 1)
		return false
	}

	// Add signal to priority queue
	heap.Push(&ss.queue, signal)

	// Update statistics
	atomic.AddInt64(&ss.totalScheduled, 1)

	return true
}

// ProcessDueSignals delivers all signals that are ready for delivery
// Called by the neuron's Run() loop on each ticker interval (typically 1ms)
//
// PARAMETERS:
// currentTime: The current time for determining which signals are due
//
// RETURNS:
// Number of signals that were delivered during this call
//
// BIOLOGICAL TIMING:
// Real neurons have ~1ms resolution for spike timing, which matches
// perfectly with the neuron's existing decayTicker interval. This provides
// biologically accurate timing without additional computational overhead.
//
// PERFORMANCE:
// This method is designed to be very fast since it's called every millisecond.
// The heap data structure ensures O(log n) performance even with large queues.
func (ss *SignalScheduler) ProcessDueSignals(currentTime time.Time) int {
	ss.queueMutex.Lock()
	defer ss.queueMutex.Unlock()

	delivered := 0

	// Process all signals that are due for delivery
	// The queue is sorted by delivery time, so we can stop at the first non-due signal
	for len(ss.queue) > 0 {
		// Peek at the earliest signal without removing it yet
		nextSignal := ss.queue[0]

		// If this signal is not yet due, stop processing
		// All remaining signals will also not be due (queue is sorted)
		if nextSignal.DeliveryTime.After(currentTime) {
			break
		}

		// Remove the due signal from the queue
		signal := heap.Pop(&ss.queue).(*synapse.ScheduledSignal)

		// Record delivery timing for latency statistics
		deliveryStart := time.Now()

		// Deliver the signal to the target neuron
		// This is the actual biological signal transmission
		if signal.Target != nil {
			signal.Target.Receive(signal.Message)
		}

		// Update delivery statistics
		deliveryDuration := time.Since(deliveryStart)
		atomic.AddInt64(&ss.totalDelivered, 1)

		// Update running average of delivery latency (exponential moving average)
		currentLatency := atomic.LoadInt64(&ss.deliveryLatency)
		newLatency := int64(deliveryDuration.Nanoseconds())
		// Exponential moving average: new_avg = old_avg * 0.9 + new_value * 0.1
		avgLatency := (currentLatency*9 + newLatency) / 10
		atomic.StoreInt64(&ss.deliveryLatency, avgLatency)

		delivered++
	}

	return delivered
}

// GetQueueStats returns current queue statistics for monitoring and debugging
//
// RETURNS:
// queueSize: Number of signals currently queued for delivery
// nextDeliveryTime: Time when the next signal is due (zero if queue empty)
// totalScheduled: Total signals ever scheduled
// totalDelivered: Total signals ever delivered
// totalDropped: Total signals dropped due to queue overflow
// avgLatencyNs: Average signal delivery latency in nanoseconds
//
// THREAD SAFETY:
// Uses read lock for queue access and atomic reads for statistics
func (ss *SignalScheduler) GetQueueStats() (queueSize int, nextDeliveryTime time.Time, totalScheduled, totalDelivered, totalDropped int64, avgLatencyNs int64) {
	// Get queue information (requires read lock)
	ss.queueMutex.RLock()
	queueSize = len(ss.queue)
	if queueSize > 0 {
		nextDeliveryTime = ss.queue[0].DeliveryTime
	}
	ss.queueMutex.RUnlock()

	// Get statistics (atomic reads, no lock needed)
	totalScheduled = atomic.LoadInt64(&ss.totalScheduled)
	totalDelivered = atomic.LoadInt64(&ss.totalDelivered)
	totalDropped = atomic.LoadInt64(&ss.totalDropped)
	avgLatencyNs = atomic.LoadInt64(&ss.deliveryLatency)

	return
}
