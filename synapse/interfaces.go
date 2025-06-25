package synapse

import (
	"time"
)

// =================================================================================
// PLASTICITY AND CONFIGURATION TYPES
// =================================================================================

// PruningConfig defines structural plasticity parameters
type PruningConfig struct {
	Enabled             bool          // Whether pruning is active
	WeightThreshold     float64       // Minimum weight to avoid pruning
	InactivityThreshold time.Duration // Maximum inactivity before pruning
}

// ExtracellularMatrix interface for spatial delay enhancement
type ExtracellularMatrix interface {
	SynapticDelay(preNeuronID, postNeuronID, synapseID string, baseDelay time.Duration) time.Duration
}
