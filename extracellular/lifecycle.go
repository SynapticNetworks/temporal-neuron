// =================================================================================
// LIFECYCLE MANAGER
// =================================================================================

package extracellular

import "sync"

// LifecycleManager handles component birth/death coordination
type LifecycleManager struct {
	registry *ComponentRegistry
	mu       sync.RWMutex
}

// NewLifecycleManager creates a lifecycle manager
func NewLifecycleManager(registry *ComponentRegistry) *LifecycleManager {
	return &LifecycleManager{
		registry: registry,
	}
}

// CreateComponent handles component creation
func (lm *LifecycleManager) CreateComponent(info ComponentInfo) error {
	return lm.registry.Register(info)
}

// RemoveComponent handles component removal
func (lm *LifecycleManager) RemoveComponent(id string) error {
	return lm.registry.Unregister(id)
}
