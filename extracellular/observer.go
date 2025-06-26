package extracellular

// Add these implementations to extracellular/observer.go

import (
	"log"
	"sync"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

// =================================================================================
// CONCRETE OBSERVER IMPLEMENTATIONS
// =================================================================================

// LoggingObserver logs all biological events to standard output
// Useful for debugging and monitoring neural activity
type LoggingObserver struct {
	prefix string
	mu     sync.Mutex
}

// NewLoggingObserver creates a new logging observer with optional prefix
func NewLoggingObserver(prefix string) *LoggingObserver {
	if prefix == "" {
		prefix = "[NEURAL]"
	}
	return &LoggingObserver{prefix: prefix}
}

// Emit logs the biological event (non-blocking, thread-safe)
func (lo *LoggingObserver) Emit(event types.BiologicalEvent) {
	// Use goroutine to ensure non-blocking as per interface requirements
	go func() {
		lo.mu.Lock()
		defer lo.mu.Unlock()

		log.Printf("%s %s: %s->%s at %v",
			lo.prefix,
			event.EventType,
			event.SourceID,
			event.TargetID,
			event.Timestamp.Format("15:04:05.000"))

		if event.Description != "" {
			log.Printf("  Description: %s", event.Description)
		}

	}()
}

// BufferedObserver collects events in a buffer for batch processing
// Optimized for high-frequency neural activity with minimal overhead
type BufferedObserver struct {
	buffer   []types.BiologicalEvent
	capacity int
	mu       sync.RWMutex
	onFlush  func([]types.BiologicalEvent)
}

// NewBufferedObserver creates a buffered observer with specified capacity
func NewBufferedObserver(capacity int, onFlush func([]types.BiologicalEvent)) *BufferedObserver {
	return &BufferedObserver{
		buffer:   make([]types.BiologicalEvent, 0, capacity),
		capacity: capacity,
		onFlush:  onFlush,
	}
}

// Emit adds event to buffer (non-blocking, thread-safe)
func (bo *BufferedObserver) Emit(event types.BiologicalEvent) {
	bo.mu.Lock()
	defer bo.mu.Unlock()

	bo.buffer = append(bo.buffer, event)

	// Auto-flush when buffer is full
	if len(bo.buffer) >= bo.capacity {
		if bo.onFlush != nil {
			// Copy buffer for async processing
			events := make([]types.BiologicalEvent, len(bo.buffer))
			copy(events, bo.buffer)

			// Async flush to maintain non-blocking guarantee
			go bo.onFlush(events)
		}

		// Reset buffer
		bo.buffer = bo.buffer[:0]
	}
}

// Flush manually flushes the current buffer
func (bo *BufferedObserver) Flush() {
	bo.mu.Lock()
	defer bo.mu.Unlock()

	if len(bo.buffer) > 0 && bo.onFlush != nil {
		events := make([]types.BiologicalEvent, len(bo.buffer))
		copy(events, bo.buffer)

		go bo.onFlush(events)
		bo.buffer = bo.buffer[:0]
	}
}

// GetBufferSize returns current buffer size (thread-safe)
func (bo *BufferedObserver) GetBufferSize() int {
	bo.mu.RLock()
	defer bo.mu.RUnlock()
	return len(bo.buffer)
}

// FilteredObserver filters events by type before forwarding to target observer
// Useful for creating specialized monitoring for specific activities
type FilteredObserver struct {
	target       types.BiologicalObserver
	allowedTypes map[types.EventType]bool
	mu           sync.RWMutex
}

// NewFilteredObserver creates a filtered observer that only passes specified event types
func NewFilteredObserver(target types.BiologicalObserver, allowedTypes []types.EventType) *FilteredObserver {
	typeMap := make(map[types.EventType]bool)
	for _, eventType := range allowedTypes {
		typeMap[eventType] = true
	}

	return &FilteredObserver{
		target:       target,
		allowedTypes: typeMap,
	}
}

// Emit forwards event to target observer if event type is allowed (non-blocking)
func (fo *FilteredObserver) Emit(event types.BiologicalEvent) {
	fo.mu.RLock()
	allowed := fo.allowedTypes[event.EventType]
	fo.mu.RUnlock()

	if allowed && fo.target != nil {
		fo.target.Emit(event)
	}
}

// AddEventType adds an event type to the filter (thread-safe)
func (fo *FilteredObserver) AddEventType(eventType types.EventType) {
	fo.mu.Lock()
	defer fo.mu.Unlock()
	fo.allowedTypes[eventType] = true
}

// RemoveEventType removes an event type from the filter (thread-safe)
func (fo *FilteredObserver) RemoveEventType(eventType types.EventType) {
	fo.mu.Lock()
	defer fo.mu.Unlock()
	delete(fo.allowedTypes, eventType)
}

// MultiObserver broadcasts events to multiple observers
// Allows combining different observation strategies
type MultiObserver struct {
	observers []types.BiologicalObserver
	mu        sync.RWMutex
}

// NewMultiObserver creates an observer that forwards events to multiple targets
func NewMultiObserver(observers ...types.BiologicalObserver) *MultiObserver {
	return &MultiObserver{
		observers: observers,
	}
}

// Emit forwards event to all registered observers (non-blocking, concurrent)
func (mo *MultiObserver) Emit(event types.BiologicalEvent) {
	mo.mu.RLock()
	observers := make([]types.BiologicalObserver, len(mo.observers))
	copy(observers, mo.observers)
	mo.mu.RUnlock()

	// Forward to all observers concurrently
	for _, observer := range observers {
		if observer != nil {
			// Each observer is called concurrently to maintain non-blocking guarantee
			go observer.Emit(event)
		}
	}
}

// AddObserver adds a new observer to the broadcast list (thread-safe)
func (mo *MultiObserver) AddObserver(observer types.BiologicalObserver) {
	if observer == nil {
		return
	}

	mo.mu.Lock()
	defer mo.mu.Unlock()
	mo.observers = append(mo.observers, observer)
}

// RemoveObserver removes an observer from the broadcast list (thread-safe)
func (mo *MultiObserver) RemoveObserver(target types.BiologicalObserver) {
	mo.mu.Lock()
	defer mo.mu.Unlock()

	for i, observer := range mo.observers {
		if observer == target {
			mo.observers = append(mo.observers[:i], mo.observers[i+1:]...)
			return
		}
	}
}

// CountObserver tracks event counts by type
// Useful for performance monitoring and activity analysis
type CountObserver struct {
	counts map[types.EventType]int64
	total  int64
	mu     sync.RWMutex
}

// NewCountObserver creates a new counting observer
func NewCountObserver() *CountObserver {
	return &CountObserver{
		counts: make(map[types.EventType]int64),
	}
}

// Emit increments counters for the event type (non-blocking, thread-safe)
func (co *CountObserver) Emit(event types.BiologicalEvent) {
	go func() {
		co.mu.Lock()
		defer co.mu.Unlock()

		co.counts[event.EventType]++
		co.total++
	}()
}

// GetCounts returns a copy of current event counts (thread-safe)
func (co *CountObserver) GetCounts() map[types.EventType]int64 {
	co.mu.RLock()
	defer co.mu.RUnlock()

	counts := make(map[types.EventType]int64)
	for eventType, count := range co.counts {
		counts[eventType] = count
	}
	return counts
}

// GetTotal returns total number of events observed (thread-safe)
func (co *CountObserver) GetTotal() int64 {
	co.mu.RLock()
	defer co.mu.RUnlock()
	return co.total
}

// Reset clears all counters (thread-safe)
func (co *CountObserver) Reset() {
	co.mu.Lock()
	defer co.mu.Unlock()

	co.counts = make(map[types.EventType]int64)
	co.total = 0
}

// NullObserver discards all events (useful for testing overhead)
// Provides absolute minimum overhead implementation
type NullObserver struct{}

// NewNullObserver creates a null observer that discards all events
func NewNullObserver() *NullObserver {
	return &NullObserver{}
}

// Emit does nothing (maximum performance, zero overhead)
func (no *NullObserver) Emit(event types.BiologicalEvent) {
	// Intentionally empty - discard all events
}

// =================================================================================
// OBSERVER UTILITIES
// =================================================================================

// ObserverChain creates a chain of observers for event processing pipeline
func ObserverChain(observers ...types.BiologicalObserver) types.BiologicalObserver {
	if len(observers) == 0 {
		return NewNullObserver()
	}

	if len(observers) == 1 {
		return observers[0]
	}

	return NewMultiObserver(observers...)
}

// NewActivityMonitor creates a specialized observer for neural activity monitoring
func NewActivityMonitor() *FilteredObserver {
	activityTypes := []types.EventType{
		types.LigandReleased,
		types.ElectricalSignalSent,
		// Add neuron firing events when available
	}

	logger := NewLoggingObserver("[ACTIVITY]")
	return NewFilteredObserver(logger, activityTypes)
}

// NewPerformanceMonitor creates an observer optimized for performance testing
func NewPerformanceMonitor(capacity int) *BufferedObserver {
	return NewBufferedObserver(capacity, func(events []types.BiologicalEvent) {
		// Minimal processing for performance testing
		log.Printf("[PERF] Processed %d events", len(events))
	})
}
