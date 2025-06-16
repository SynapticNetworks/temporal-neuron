/*
=================================================================================
BIOLOGICAL OBSERVER - FUNCTIONAL EVENT EMISSION
=================================================================================

This file defines the interface for a "Biological Observer," a system for
emitting structured, functional events from the Extracellular Matrix. It acts
as a decoupled logging and telemetry layer, allowing external tools to monitor,
analyze, or visualize the simulation's behavior in real-time.

BIOLOGICAL ANALOGY:
This system is analogous to using metabolic or fluorescent tracers in a real
biological experiment. We "tag" specific actions (like pruning, neurotransmitter
release, or cell creation) to observe their dynamics without interfering with
the core processes. The observer is the virtual microscope, and the events are
the signals it captures.

DESIGN:
- The `BiologicalObserver` interface is optional. If not provided to the
  ExtracellularMatrix, the system runs with zero event-handling overhead.
- The `Emit` method must be non-blocking to ensure it doesn't slow down the
  simulation's critical path.
- `BiologicalEvent` provides a rich, structured format for all events,
  containing what happened, who was involved, when it happened, and any
  relevant contextual data.
=================================================================================
*/

package extracellular

import "time"

// EventType is a string identifier for the type of a biological event.
// Using a dedicated type provides clarity and enables easier filtering.
type EventType string

// Constants for all defined biological event types.
const (
	// --- Microglia Events (Lifecycle & Health) ---
	HealthPenaltyApplied        EventType = "health.penalty.applied"
	PruningCandidateMarked      EventType = "pruning.candidate.marked"
	ConnectionPruned            EventType = "connection.pruned"
	BirthRequestEvaluated       EventType = "birth.request.evaluated"
	ComponentApoptosisScheduled EventType = "component.apoptosis.scheduled"
	PatrolCompleted             EventType = "patrol.completed"

	// --- Astrocyte Network Events (Structural) ---
	ComponentRegistered   EventType = "component.registered"
	ComponentUnregistered EventType = "component.unregistered"
	TerritoryAdjusted     EventType = "territory.adjusted"

	// --- Chemical Modulator Events (Signaling) ---
	LigandReleased      EventType = "ligand.released"
	LigandBoundToTarget EventType = "ligand.bound.target" // Potentially high-traffic

	// --- Gap Junction Events (Electrical) ---
	ElectricalSignalSent          EventType = "electrical.signal.sent"
	ElectricalCouplingEstablished EventType = "electrical.coupling.established"
	ElectricalCouplingRemoved     EventType = "electrical.coupling.removed"
)

// BiologicalEvent represents a single, significant functional occurrence within the matrix.
// It is designed to be a rich, structured data object for easy consumption.
type BiologicalEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	EventType   EventType              `json:"event_type"`
	SourceID    string                 `json:"source_id"`           // Component that initiated the event (e.g., a specific Microglia).
	TargetID    string                 `json:"target_id,omitempty"` // Component affected by the event (e.g., a Neuron being pruned).
	Description string                 `json:"description"`         // Human-readable summary of the event.
	Metadata    map[string]interface{} `json:"metadata,omitempty"`  // Rich, structured data specific to the event type.
}

// BiologicalObserver defines the interface for an event emission system.
// Any system that implements this interface can be registered with the
// ExtracellularMatrix to receive a real-time stream of biological events.
type BiologicalObserver interface {
	// Emit records a significant biological or functional event.
	// Implementations of this method MUST be non-blocking and thread-safe
	// to avoid impacting simulation performance. A common pattern is to
	// send the event to a buffered channel for asynchronous processing.
	Emit(event BiologicalEvent)
}
