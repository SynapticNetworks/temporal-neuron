/*
=================================================================================
BIOLOGICAL HELPERS
=================================================================================

Helper functions for creating biologically realistic network configurations
that pass validation tests.
=================================================================================
*/

package extracellular

import (
	"math"
	"math/rand"
	"time"
)

const (
	// Spatial scales (micrometers)
	NEURON_SOMA_DIAMETER       = 15.0  // Typical cortical neuron soma: 10-20 μm
	SYNAPTIC_CLEFT_WIDTH       = 0.02  // Synaptic cleft: 20 nanometers
	CORTICAL_COLUMN_DIAMETER   = 500.0 // Cortical column: ~500 μm diameter
	ASTROCYTE_TERRITORY_RADIUS = 50.0  // Astrocyte domain: ~50-100 μm radius

	// Temporal scales
	ACTION_POTENTIAL_DURATION = 2 * time.Millisecond   // 1-2 ms
	SYNAPTIC_DELAY            = 1 * time.Millisecond   // 0.5-1 ms
	GLUTAMATE_CLEARANCE_TIME  = 5 * time.Millisecond   // 1-10 ms
	GABA_CLEARANCE_TIME       = 10 * time.Millisecond  // 5-20 ms
	DOPAMINE_HALF_LIFE        = 100 * time.Millisecond // 50-200 ms

	// Concentration ranges (molar)
	GLUTAMATE_PEAK_CONC = 1.0   // 1 mM peak in synaptic cleft
	GABA_PEAK_CONC      = 0.5   // 0.5 mM peak concentration
	DOPAMINE_BASELINE   = 0.001 // 1 μM baseline in striatum
	DOPAMINE_PEAK       = 0.01  // 10 μM peak during reward

	// Network properties
	CORTICAL_NEURON_DENSITY  = 150000.0 // ~150k neurons/mm³ in cortex
	GAP_JUNCTION_CONDUCTANCE = 0.1      // 0.1-1 nS typical conductance
	ASTROCYTE_NEURON_RATIO   = 0.3      // ~1 astrocyte per 3 neurons in cortex

	SYNAPSES_PER_NEURON = 7000 // 5k-10k synapses per cortical neuron (average)
)

// CreateBiologicalNeuronDensity calculates appropriate neuron count for biological density
func CreateBiologicalNeuronDensity(volumeRadius float64, targetDensityPerMM3 int) int {
	// Calculate volume in mm³ (convert from μm)
	radiusInMM := volumeRadius / 1000.0
	volumeInMM3 := (4.0 / 3.0) * math.Pi * math.Pow(radiusInMM, 3)

	// Calculate appropriate neuron count for biological density
	targetCount := int(float64(targetDensityPerMM3) * volumeInMM3)

	// Ensure minimum biological realism
	if targetCount < 10 {
		targetCount = 10
	}

	return targetCount
}

// EstablishBiologicalConnectivity creates realistic local/distant connection ratios
func EstablishBiologicalConnectivity(neurons []ComponentInfo, localRadius float64,
	astrocyteNetwork *AstrocyteNetwork) error {

	for _, neuron := range neurons {
		localConnections := 0
		totalConnections := 0

		// Find nearby neurons for local connections
		nearbyNeurons := astrocyteNetwork.FindNearby(neuron.Position, localRadius)

		for _, nearby := range nearbyNeurons {
			if nearby.ID != neuron.ID {
				// Establish local connection with high probability
				if rand.Float64() < 0.8 { // 80% local connection probability
					err := astrocyteNetwork.MapConnection(neuron.ID, nearby.ID)
					if err == nil {
						localConnections++
						totalConnections++
					}
				}
			}
		}

		// Add some distant connections to reach biological ratio
		// Aim for 70% local, 30% distant connections
		targetTotal := localConnections * 10 / 7 // Scale to achieve 70% local
		distantNeeded := targetTotal - localConnections

		// Find distant neurons
		allNeurons := astrocyteNetwork.FindByType(ComponentNeuron)
		for i := 0; i < distantNeeded && i < len(allNeurons); i++ {
			distant := allNeurons[i]
			if distant.ID != neuron.ID {
				distance := astrocyteNetwork.Distance(neuron.Position, distant.Position)
				if distance > localRadius {
					err := astrocyteNetwork.MapConnection(neuron.ID, distant.ID)
					if err == nil {
						totalConnections++
					}
				}
			}
		}
	}

	return nil
}

// BiologicalMetrics for validation
type BiologicalMetrics struct {
	GlutamateClearanceRate float64 // Should be >50% in 5ms
	NeuronDensity          int     // Should be >15,000/mm³
	LocalConnectionRatio   float64 // Should be >70%
	AstrocyteLoadAverage   float64 // Should be <20 neurons/astrocyte
	ChemicalReleaseRate    float64 // Should be <2000/second
	NetworkHealthScore     float64 // Should be >0.8
}
