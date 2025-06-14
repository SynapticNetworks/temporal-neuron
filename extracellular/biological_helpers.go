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
