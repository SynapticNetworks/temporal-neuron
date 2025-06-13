/*
=================================================================================
CHEMICAL MODULATOR - BIOLOGICAL CHEMICAL SIGNALING
=================================================================================

Implements chemical signaling like neurotransmitters and neuromodulators.
Components can release chemicals and bind to them, just like real biology.
=================================================================================
*/

package extracellular

import (
	"sync"
)

// ChemicalModulator handles chemical signal propagation and binding
type ChemicalModulator struct {
	// Targets that can bind to each ligand type
	bindingTargets map[LigandType][]BindingTarget

	// Component registry for validation
	registry *ComponentRegistry

	// State management
	mu sync.RWMutex
}

// NewChemicalModulator creates a chemical modulator
func NewChemicalModulator(registry *ComponentRegistry) *ChemicalModulator {
	return &ChemicalModulator{
		bindingTargets: make(map[LigandType][]BindingTarget),
		registry:       registry,
	}
}

// Release sends a chemical signal to all targets that can bind to it
func (cm *ChemicalModulator) Release(ligandType LigandType, sourceID string, concentration float64) error {
	cm.mu.RLock()
	targets := make([]BindingTarget, len(cm.bindingTargets[ligandType]))
	copy(targets, cm.bindingTargets[ligandType])
	cm.mu.RUnlock()

	// Send to all targets that have receptors for this ligand
	for _, target := range targets {
		target.Bind(ligandType, sourceID, concentration)
	}

	return nil
}

// RegisterTarget adds a component that can receive chemical signals
func (cm *ChemicalModulator) RegisterTarget(target BindingTarget) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Register target for each ligand type it can bind to
	for _, ligandType := range target.GetReceptors() {
		cm.bindingTargets[ligandType] = append(cm.bindingTargets[ligandType], target)
	}

	return nil
}

// UnregisterTarget removes a component from receiving chemical signals
func (cm *ChemicalModulator) UnregisterTarget(target BindingTarget) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Remove from all ligand type lists
	for _, ligandType := range target.GetReceptors() {
		targets := cm.bindingTargets[ligandType]
		for i, t := range targets {
			if t == target {
				// Remove target from slice
				cm.bindingTargets[ligandType] = append(targets[:i], targets[i+1:]...)
				break
			}
		}
	}

	return nil
}

// GetConcentration returns current concentration at a position (placeholder)
func (cm *ChemicalModulator) GetConcentration(ligandType LigandType, position Position3D) float64 {
	// TODO: Implement spatial concentration calculation
	return 0.0
}

// Start begins chemical processing
func (cm *ChemicalModulator) Start() error {
	// TODO: Start any background processing if needed
	return nil
}

// Stop ends chemical processing
func (cm *ChemicalModulator) Stop() error {
	// TODO: Clean up any background processing
	return nil
}
