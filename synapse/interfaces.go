package synapse

import (
	"time"
)

// ExtracellularMatrix interface for spatial delay enhancement
type ExtracellularMatrix interface {
	SynapticDelay(preNeuronID, postNeuronID, synapseID string, baseDelay time.Duration) time.Duration
}
