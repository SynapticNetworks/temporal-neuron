/*
=================================================================================
COMPONENT REGISTRY - SIMPLE COMPONENT TRACKING
=================================================================================

Tracks network components without complex abstractions.
Simple registration and lookup for biological coordination.
=================================================================================
*/

package extracellular

import (
	"fmt"
	"sync"
	"time"
)

// ComponentRegistry tracks network components
type ComponentRegistry struct {
	components map[string]ComponentInfo
	mu         sync.RWMutex
}

// NewComponentRegistry creates a component registry
func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{
		components: make(map[string]ComponentInfo),
	}
}

// Register adds a component to the registry
func (cr *ComponentRegistry) Register(info ComponentInfo) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if info.ID == "" {
		return fmt.Errorf("component ID cannot be empty")
	}

	// Set registration time if not provided
	if info.RegisteredAt.IsZero() {
		info.RegisteredAt = time.Now()
	}

	cr.components[info.ID] = info
	return nil
}

// Unregister removes a component from the registry
func (cr *ComponentRegistry) Unregister(id string) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	delete(cr.components, id)
	return nil
}

// Get retrieves a component by ID
func (cr *ComponentRegistry) Get(id string) (ComponentInfo, bool) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	info, exists := cr.components[id]
	return info, exists
}

// Find searches for components matching criteria
func (cr *ComponentRegistry) Find(criteria ComponentCriteria) []ComponentInfo {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	var results []ComponentInfo

	for _, info := range cr.components {
		if cr.matches(info, criteria) {
			results = append(results, info)
		}
	}

	return results
}

// List returns all registered components
func (cr *ComponentRegistry) List() []ComponentInfo {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	results := make([]ComponentInfo, 0, len(cr.components))
	for _, info := range cr.components {
		results = append(results, info)
	}

	return results
}

// Count returns the number of registered components
func (cr *ComponentRegistry) Count() int {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	return len(cr.components)
}

// UpdateState updates a component's state
func (cr *ComponentRegistry) UpdateState(id string, state ComponentState) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if info, exists := cr.components[id]; exists {
		info.State = state
		cr.components[id] = info
		return nil
	}

	return fmt.Errorf("component %s not found", id)
}

// matches checks if a component matches the given criteria
func (cr *ComponentRegistry) matches(info ComponentInfo, criteria ComponentCriteria) bool {
	// Check type filter
	if criteria.Type != nil && info.Type != *criteria.Type {
		return false
	}

	// Check state filter
	if criteria.State != nil && info.State != *criteria.State {
		return false
	}

	// Check spatial filter
	if criteria.Position != nil && criteria.Radius > 0 {
		distance := cr.calculateDistance(info.Position, *criteria.Position)
		if distance > criteria.Radius {
			return false
		}
	}

	return true
}

// calculateDistance computes 3D distance between positions
func (cr *ComponentRegistry) calculateDistance(pos1, pos2 Position3D) float64 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	dz := pos1.Z - pos2.Z
	return dx*dx + dy*dy + dz*dz // Square distance (avoid sqrt for performance)
}
