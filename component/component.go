package component

import (
	"fmt"
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

// ============================================================================
// METADATA AND MONITORING STRUCTURES
// ============================================================================

// ComponentInfo encapsulates comprehensive information about a registered component.
// This structure is used for introspection, debugging, and overall system monitoring.
type ComponentInfo struct {
	ID           string                 `json:"id"`            // Unique identifier for the component.
	Type         types.ComponentType    `json:"type"`          // The categorical type of the component (e.g., Neuron, Synapse).
	Position     types.Position3D       `json:"position"`      // The 3D spatial coordinates of the component.
	State        types.ComponentState   `json:"state"`         // The current operational state of the component.
	RegisteredAt time.Time              `json:"registered_at"` // Timestamp when the component was first registered or created.
	Metadata     map[string]interface{} `json:"metadata"`      // A flexible map for storing additional, component-specific attributes.
}

// HealthMetrics provides a snapshot of a component's health, performance, and activity.
// This structure is vital for monitoring the dynamic behavior and well-being of
// individual components within a large-scale simulation.
type HealthMetrics struct {
	ActivityLevel   float64   `json:"activity_level"`    // A normalized value indicating the recent activity intensity (e.g., firing rate for neurons).
	ConnectionCount int       `json:"connection_count"`  // The number of active connections (e.g., synapses for a neuron).
	ProcessingLoad  float64   `json:"processing_load"`   // An estimate of the computational resources being consumed by the component.
	LastHealthCheck time.Time `json:"last_health_check"` // Timestamp of the most recent health evaluation.
	HealthScore     float64   `json:"health_score"`      // An aggregate score representing the overall health (e.g., 0.0-1.0).
	Issues          []string  `json:"issues"`            // A list of any identified issues or warnings.
}

// ============================================================================
// CORE COMPONENT INTERFACE DEFINITION
// ============================================================================

// Component is the fundamental interface that all active elements in the neural
// simulation must implement. It defines the contract for basic identification,
// lifecycle management, and metadata access, ensuring a consistent API across
// diverse component types (e.g., neurons, synapses, glial cells).
type Component interface {
	// ID returns the globally unique identifier for this component.
	ID() string
	// Type returns the categorical type of the component (e.g., TypeNeuron, TypeSynapse).
	Type() types.ComponentType
	// Position returns the 3D spatial coordinates of the component.
	Position() types.Position3D
	// types.State returns the current operational state of the component.
	State() types.ComponentState

	// IsActive checks if the component is currently in an active and operational state.
	IsActive() bool
	// Start initiates the component's operation. Returns an error if the component
	// cannot be started.
	Start() error
	// Stop gracefully ceases the component's operation and transitions it to a stopped state.
	// Returns an error if the shutdown process encounters issues.
	Stop() error
	// CanRestart determines if the component can transition back to an active state
	// from its current state (e.g., from Inactive or Stopped).
	CanRestart() bool
	// Restart attempts to reactivate a component that is in a restartable state.
	// Returns an error if the component cannot be restarted.
	Restart() error

	// GetMetadata retrieves a copy of the component's dynamic metadata.
	GetMetadata() map[string]interface{}
	// UpdateMetadata sets or updates a specific key-value pair in the component's metadata.
	UpdateMetadata(key string, value interface{})

	// SetPosition updates the component's 3D spatial coordinates.
	SetPosition(position types.Position3D)
	// SetState manually sets the component's operational state.
	SetState(state types.ComponentState)

	// GetActivityLevel returns a normalized measure of the component's recent activity.
	// The interpretation of 'activity' is component-specific (e.g., firing rate for neurons).
	GetActivityLevel() float64
	// GetLastActivity returns the timestamp of the component's most recent significant activity or state change.
	GetLastActivity() time.Time
}

// ============================================================================
// BASE COMPONENT IMPLEMENTATION
// ============================================================================

// BaseComponent provides a concrete, default implementation of the `Component` interface.
// It is designed to be embedded within more specific neural component structs (e.g., `Neuron`, `Synapse`)
// to provide common functionalities such as ID management, state transitions, position tracking,
// and thread-safe metadata handling. This promotes code reuse and consistency across the simulation.
type BaseComponent struct {
	id            string                 // Unique identifier for this instance.
	componentType types.ComponentType    // The specific type of this component.
	position      types.Position3D       // Current 3D spatial coordinates.
	state         types.ComponentState   // Current operational state (e.g., Active, Stopped).
	metadata      map[string]interface{} // Dynamic, extensible key-value store for component-specific data.
	lastActivity  time.Time              // Timestamp of the last significant activity or state update.
	isActive      bool                   // A boolean flag indicating if the component is considered 'active'.
	mu            sync.RWMutex           // A RWMutex for protecting concurrent access to component state and metadata.
}

// NewBaseComponent is the constructor for BaseComponent. It initializes a new
// base component with mandatory identification and spatial properties, setting
// its initial state to `StateActive`.
func NewBaseComponent(id string, componentType types.ComponentType, position types.Position3D) *BaseComponent {
	return &BaseComponent{
		id:            id,
		componentType: componentType,
		position:      position,
		state:         types.StateActive,            // Components start as active by default.
		metadata:      make(map[string]interface{}), // Initialize metadata map.
		lastActivity:  time.Now(),                   // Record creation time as initial activity.
		isActive:      true,                         // Set active flag.
	}
}

// ============================================================================
// BASE COMPONENT: COMPONENT INTERFACE IMPLEMENTATION
// ============================================================================

// ID returns the unique identifier of the base component. It is safe for concurrent access.
func (bc *BaseComponent) ID() string {
	return bc.id
}

// Type returns the categorical type of the base component. It is safe for concurrent access.
func (bc *BaseComponent) Type() types.ComponentType {
	return bc.componentType
}

// Position returns the current 3D position of the base component.
// It uses a read lock for thread safety.
func (bc *BaseComponent) Position() types.Position3D {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.position
}

// SetPosition updates the 3D position of the base component.
// It uses a write lock for thread safety and updates the last activity timestamp.
func (bc *BaseComponent) SetPosition(position types.Position3D) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.position = position
	bc.lastActivity = time.Now() // Mark activity on position change.
}

// types.State returns the current operational state of the base component.
// It uses a read lock for thread safety.
func (bc *BaseComponent) State() types.ComponentState {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.state
}

// SetState updates the operational state of the base component.
// It uses a write lock for thread safety and updates the last activity timestamp.
func (bc *BaseComponent) SetState(state types.ComponentState) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.state = state
	bc.lastActivity = time.Now() // Mark activity on state change.
}

// CanRestart checks if the component's current state allows it to be restarted.
// It uses a read lock for thread safety.
func (bc *BaseComponent) CanRestart() bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.canRestartUnsafe()
}

// canRestartUnsafe is an internal helper for CanRestart that assumes the lock
// is already held. It checks if the component can transition back to an active state.
func (bc *BaseComponent) canRestartUnsafe() bool {
	switch bc.state {
	case types.StateInactive, types.StateStopped, types.StateMaintenance, types.StateHibernating:
		return true // These states are typically restartable.
	case types.StateDying, types.StateDamaged:
		return false // Requires special recovery or is beyond recovery.
	case types.StateActive, types.StateShuttingDown, types.StateDeveloping:
		return false // Already active or in an ongoing transition.
	default:
		return false // Unknown states are not restartable by default.
	}
}

// Restart attempts to reactivate the component, setting its state to `StateActive`.
// It uses a write lock for thread safety and validates if restarting is possible.
func (bc *BaseComponent) Restart() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Check restartability using the unsafe version to prevent deadlock.
	if !bc.canRestartUnsafe() {
		return fmt.Errorf("component %s cannot be restarted from state %s", bc.id, bc.state)
	}

	bc.isActive = true           // Mark as active.
	bc.state = types.StateActive // Set state to active.
	bc.lastActivity = time.Now() // Update last activity timestamp.

	return nil
}

// Start sets the component to an active state. This is typically called on initialization
// or to bring a stopped/inactive component online. It uses a write lock for thread safety.
func (bc *BaseComponent) Start() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.isActive = true
	bc.state = types.StateActive
	bc.lastActivity = time.Now()
	return nil
}

// Stop gracefully shuts down the component, transitioning its state through `StateShuttingDown`
// to `StateStopped`. This method can be extended to include cleanup operations specific
// to a component (e.g., closing connections, releasing resources).
// It uses a write lock for thread safety.
func (bc *BaseComponent) Stop() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Initiate graceful shutdown.
	bc.state = types.StateShuttingDown
	bc.lastActivity = time.Now()

	// In a more complex simulation, this section would contain:
	// - Logic to close network connections.
	// - Code to release system resources (e.g., memory, file handles).
	// - Mechanisms to notify other dependent components of the shutdown.
	// For this base implementation, we simulate the state transition.

	// Mark as fully stopped and inactive.
	bc.isActive = false
	bc.state = types.StateStopped
	bc.lastActivity = time.Now()

	return nil
}

// IsActive returns true if the component is currently marked as active and its
// state is `StateActive`. It uses a read lock for thread safety.
func (bc *BaseComponent) IsActive() bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.isActive && bc.state == types.StateActive
}

// GetMetadata retrieves a copy of the component's internal metadata map.
// A copy is returned to prevent external modifications from affecting the component's
// internal state. It uses a read lock for thread safety.
func (bc *BaseComponent) GetMetadata() map[string]interface{} {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	metadata := make(map[string]interface{}, len(bc.metadata))
	for k, v := range bc.metadata {
		metadata[k] = v
	}
	return metadata
}

// UpdateMetadata adds or modifies a key-value pair in the component's metadata.
// It uses a write lock for thread safety and updates the last activity timestamp.
func (bc *BaseComponent) UpdateMetadata(key string, value interface{}) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.metadata[key] = value
	bc.lastActivity = time.Now() // Mark activity on metadata change.
}

// GetLastActivity returns the timestamp of the most recent activity or state change
// recorded for this component. It uses a read lock for thread safety.
func (bc *BaseComponent) GetLastActivity() time.Time {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.lastActivity
}

// GetActivityLevel provides a basic, default implementation for assessing the component's
// recent activity. This method can be overridden by more specific component types
// (e.g., Neuron) to provide more meaningful activity metrics (e.g., firing rate).
// It returns 1.0 if active within 1 second, 0.5 within 10 seconds, else 0.0.
// It uses a read lock for thread safety.
func (bc *BaseComponent) GetActivityLevel() float64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	timeSinceActivity := time.Since(bc.lastActivity)
	if timeSinceActivity < time.Second {
		return 1.0 // Highly active recently.
	} else if timeSinceActivity < 10*time.Second {
		return 0.5 // Moderately active.
	}
	return 0.0 // Inactive for a prolonged period.
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// CreateComponentInfo generates a `ComponentInfo` structure from any object
// that implements the `Component` interface. This is a convenient helper for
// collecting and exposing component details.
func CreateComponentInfo(comp Component) ComponentInfo {
	return ComponentInfo{
		ID:           comp.ID(),
		Type:         comp.Type(),
		Position:     comp.Position(),
		State:        comp.State(),
		RegisteredAt: time.Now(), // Captures current time as registration time.
		Metadata:     comp.GetMetadata(),
	}
}

// FilterComponentsByType filters a slice of `Component` interfaces, returning
// a new slice containing only components of the specified `types.ComponentType`.
func FilterComponentsByType(components []Component, componentType types.ComponentType) []Component {
	var filtered []Component
	for _, comp := range components {
		if comp.Type() == componentType {
			filtered = append(filtered, comp)
		}
	}
	return filtered
}

// FilterComponentsByState filters a slice of `Component` interfaces, returning
// a new slice containing only components that are in the specified `types.ComponentState`.
func FilterComponentsByState(components []Component, state types.ComponentState) []Component {
	var filtered []Component
	for _, comp := range components {
		if comp.State() == state {
			filtered = append(filtered, comp)
		}
	}
	return filtered
}
