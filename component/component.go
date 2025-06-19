package component

import (
	"sync"
	"time"
)

// ============================================================================
// COMPONENT ARCHITECTURE TYPES
// ============================================================================

// ComponentType represents different types of neural components
type ComponentType int

const (
	TypeNeuron        ComponentType = iota // Excitable neural cell
	TypeSynapse                            // Synaptic connection
	TypeGlialCell                          // Support cell (astrocyte, oligodendrocyte, etc.)
	TypeMicrogliaCell                      // Immune cell of the brain
	TypeEpendymalCell                      // CSF-brain barrier cell
)

func (ct ComponentType) String() string {
	switch ct {
	case TypeNeuron:
		return "Neuron"
	case TypeSynapse:
		return "Synapse"
	case TypeGlialCell:
		return "GlialCell"
	case TypeMicrogliaCell:
		return "MicrogliaCell"
	case TypeEpendymalCell:
		return "EpendymalCell"
	default:
		return "Unknown"
	}
}

// ComponentState represents the operational state of neural components
type ComponentState int

const (
	StateActive       ComponentState = iota // Normal operational state
	StateInactive                           // Temporarily disabled
	StateShuttingDown                       // Graceful shutdown in progress
	StateDeveloping                         // Growing/maturing (developmental)
	StateDying                              // Programmed cell death/apoptosis
	StateDamaged                            // Damaged but potentially recoverable
	StateMaintenance                        // Undergoing maintenance/repair
	StateHibernating                        // Low-activity conservation state
)

func (cs ComponentState) String() string {
	switch cs {
	case StateActive:
		return "Active"
	case StateInactive:
		return "Inactive"
	case StateShuttingDown:
		return "ShuttingDown"
	case StateDeveloping:
		return "Developing"
	case StateDying:
		return "Dying"
	case StateDamaged:
		return "Damaged"
	case StateMaintenance:
		return "Maintenance"
	case StateHibernating:
		return "Hibernating"
	default:
		return "Unknown"
	}
}

// ============================================================================
// 3D SPATIAL POSITIONING
// ============================================================================

// Position3D represents spatial coordinates in 3D space
type Position3D struct {
	X, Y, Z float64
}

// ============================================================================
// COMPONENT METADATA AND INFORMATION
// ============================================================================

// ComponentInfo contains complete information about a registered component
type ComponentInfo struct {
	ID           string                 `json:"id"`
	Type         ComponentType          `json:"type"`
	Position     Position3D             `json:"position"`
	State        ComponentState         `json:"state"`
	RegisteredAt time.Time              `json:"registered_at"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// HealthMetrics represents health and performance monitoring data
type HealthMetrics struct {
	ActivityLevel   float64   `json:"activity_level"`
	ConnectionCount int       `json:"connection_count"`
	ProcessingLoad  float64   `json:"processing_load"`
	LastHealthCheck time.Time `json:"last_health_check"`
	HealthScore     float64   `json:"health_score"`
	Issues          []string  `json:"issues"`
}

// ============================================================================
// CORE COMPONENT INTERFACE
// ============================================================================

// Component is the base interface that all neural components must implement
type Component interface {
	// Core identification
	ID() string
	Type() ComponentType
	Position() Position3D
	State() ComponentState

	// Lifecycle management
	IsActive() bool
	Start() error
	Stop() error

	// Metadata and monitoring
	GetMetadata() map[string]interface{}
	UpdateMetadata(key string, value interface{})

	// State management
	SetState(state ComponentState)

	// Activity monitoring
	GetActivityLevel() float64
	GetLastActivity() time.Time
}

// ============================================================================
// BASE COMPONENT IMPLEMENTATION
// ============================================================================

// BaseComponent provides default implementation of the Component interface
// All neural components should embed this to get core functionality
type BaseComponent struct {
	id            string
	componentType ComponentType
	position      Position3D
	state         ComponentState
	metadata      map[string]interface{}
	lastActivity  time.Time
	isActive      bool
	mu            sync.RWMutex
}

// NewBaseComponent creates a new base component with the specified properties
func NewBaseComponent(id string, componentType ComponentType, position Position3D) *BaseComponent {
	return &BaseComponent{
		id:            id,
		componentType: componentType,
		position:      position,
		state:         StateActive,
		metadata:      make(map[string]interface{}),
		lastActivity:  time.Now(),
		isActive:      true,
	}
}

// ============================================================================
// CORE COMPONENT INTERFACE IMPLEMENTATION
// ============================================================================

func (bc *BaseComponent) ID() string {
	return bc.id
}

func (bc *BaseComponent) Type() ComponentType {
	return bc.componentType
}

func (bc *BaseComponent) Position() Position3D {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.position
}

func (bc *BaseComponent) SetPosition(position Position3D) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.position = position
	bc.lastActivity = time.Now()
}

func (bc *BaseComponent) State() ComponentState {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.state
}

func (bc *BaseComponent) SetState(state ComponentState) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.state = state
	bc.lastActivity = time.Now()
}

func (bc *BaseComponent) IsActive() bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.isActive && bc.state == StateActive
}

func (bc *BaseComponent) GetMetadata() map[string]interface{} {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	// Return copy to prevent external modification
	metadata := make(map[string]interface{}, len(bc.metadata))
	for k, v := range bc.metadata {
		metadata[k] = v
	}
	return metadata
}

func (bc *BaseComponent) UpdateMetadata(key string, value interface{}) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.metadata[key] = value
	bc.lastActivity = time.Now()
}

func (bc *BaseComponent) GetLastActivity() time.Time {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.lastActivity
}

func (bc *BaseComponent) GetActivityLevel() float64 {
	// Default implementation - components can override
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	timeSinceActivity := time.Since(bc.lastActivity)
	if timeSinceActivity < time.Second {
		return 1.0
	} else if timeSinceActivity < 10*time.Second {
		return 0.5
	}
	return 0.0
}

// ============================================================================
// LIFECYCLE MANAGEMENT
// ============================================================================

func (bc *BaseComponent) Start() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.isActive = true
	bc.state = StateActive
	bc.lastActivity = time.Now()
	return nil
}

func (bc *BaseComponent) Stop() error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.isActive = false
	bc.state = StateInactive
	bc.lastActivity = time.Now()
	return nil
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// CreateComponentInfo creates a ComponentInfo from a Component instance
func CreateComponentInfo(comp Component) ComponentInfo {
	return ComponentInfo{
		ID:           comp.ID(),
		Type:         comp.Type(),
		Position:     comp.Position(),
		State:        comp.State(),
		RegisteredAt: time.Now(),
		Metadata:     comp.GetMetadata(),
	}
}

// FilterComponentsByType filters components by their type
func FilterComponentsByType(components []Component, componentType ComponentType) []Component {
	var filtered []Component
	for _, comp := range components {
		if comp.Type() == componentType {
			filtered = append(filtered, comp)
		}
	}
	return filtered
}

// FilterComponentsByState filters components by their state
func FilterComponentsByState(components []Component, state ComponentState) []Component {
	var filtered []Component
	for _, comp := range components {
		if comp.State() == state {
			filtered = append(filtered, comp)
		}
	}
	return filtered
}
