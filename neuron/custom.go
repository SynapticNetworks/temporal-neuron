package neuron

import "github.com/SynapticNetworks/temporal-neuron/types"

type CustomBehaviors struct {
	// Custom chemical release behavior
	CustomChemicalRelease func(activityRate, outputValue float64) []types.ChemicalRelease

	// Custom activity thresholds
	ActivityThresholdOverride *float64

	// Force specific ligand releases
	ForceReleaseList []types.LigandType
}

type ChemicalRelease struct {
	LigandType    types.LigandType
	Concentration float64
}
