package extracellular

import (
	"fmt"
	"sync"
)

// =================================================================================
// PLUGIN MANAGER
// =================================================================================

// PluginManager handles modular functionality
type PluginManager struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

// Plugin defines basic plugin interface
type Plugin interface {
	Name() string
	Start() error
	Stop() error
}

// NewPluginManager creates a plugin manager
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
	}
}

// Register adds a plugin
func (pm *PluginManager) Register(plugin Plugin) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	name := plugin.Name()
	if _, exists := pm.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}

	pm.plugins[name] = plugin
	return nil
}

// Get retrieves a plugin
func (pm *PluginManager) Get(name string) (Plugin, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if plugin, exists := pm.plugins[name]; exists {
		return plugin, nil
	}

	return nil, fmt.Errorf("plugin %s not found", name)
}

// Start starts all plugins
func (pm *PluginManager) Start() error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, plugin := range pm.plugins {
		if err := plugin.Start(); err != nil {
			return err
		}
	}

	return nil
}

// Stop stops all plugins
func (pm *PluginManager) Stop() error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	for _, plugin := range pm.plugins {
		plugin.Stop() // Ignore errors during shutdown
	}

	return nil
}
