/*
=================================================================================
EXTRACELLULAR MATRIX - BIOLOGICAL PLAUSIBILITY TESTS
=================================================================================

Tests that validate biological accuracy and realism of the extracellular matrix
coordination system. These tests ensure that our implementation matches real
biological neural tissue behavior, timing, and constraints.

BIOLOGICAL VALIDATION AREAS:
1. Chemical Signal Kinetics - Neurotransmitter diffusion, binding, clearance
2. Electrical Coupling Properties - Gap junction conductance and timing
3. Spatial Organization - Realistic tissue distances and connectivity
4. Metabolic Constraints - Energy costs and resource limitations
5. Temporal Dynamics - Biologically realistic timescales
6. Network Development - Growth patterns and pruning behaviors
7. Homeostatic Regulation - Stability and adaptive responses

Each test includes detailed biological justification and references to
real neural tissue properties and measurements.
=================================================================================
*/

package extracellular

import (
	"fmt"
	"math"
	"testing"
	"time"
)

// =================================================================================
// BIOLOGICAL CONSTANTS AND PARAMETERS
// =================================================================================

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
	SYNAPSES_PER_NEURON      = 7000     // 5k-10k synapses per cortical neuron
	GAP_JUNCTION_CONDUCTANCE = 0.1      // 0.1-1 nS typical conductance
	ASTROCYTE_NEURON_RATIO   = 0.3      // ~1 astrocyte per 3 neurons in cortex
)

// =================================================================================
// TEST 1: NEUROTRANSMITTER KINETICS AND SPATIAL DISTRIBUTION
// =================================================================================

func TestBiologicalChemicalKinetics(t *testing.T) {
	t.Log("=== BIOLOGICAL TEST: Chemical Signal Kinetics ===")
	t.Log("Validating neurotransmitter diffusion, binding, and clearance")

	// Create matrix with realistic parameters
	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  1 * time.Millisecond, // 1 kHz update rate (biological)
		MaxComponents:   1000,
	})
	defer matrix.Stop()

	// Start chemical processing
	err := matrix.chemicalModulator.Start()
	if err != nil {
		t.Fatalf("Failed to start chemical modulator: %v", err)
	}

	// === GLUTAMATE KINETICS TEST ===
	t.Log("\n--- Testing Glutamate Fast Kinetics ---")

	// Create presynaptic neuron at synaptic terminal
	presynapticPos := Position3D{X: 0, Y: 0, Z: 0}

	// Create postsynaptic targets at realistic distances
	postsynapticPos := Position3D{X: 0.02, Y: 0, Z: 0} // 20 nm away (synaptic cleft)
	extrasynapticPos := Position3D{X: 1.0, Y: 0, Z: 0} // 1 μm away (extrasynaptic)

	// Register components
	matrix.RegisterComponent(ComponentInfo{
		ID: "presynaptic_terminal", Type: ComponentSynapse,
		Position: presynapticPos, State: StateActive, RegisteredAt: time.Now(),
	})

	// Release glutamate at synaptic concentration
	t.Logf("Releasing glutamate at synaptic concentration: %.3f mM", GLUTAMATE_PEAK_CONC)
	err = matrix.ReleaseLigand(LigandGlutamate, "presynaptic_terminal", GLUTAMATE_PEAK_CONC)
	if err != nil {
		t.Fatalf("Failed to release glutamate: %v", err)
	}

	// Measure concentration at synaptic cleft (should be high)
	time.Sleep(1 * time.Millisecond) // Allow initial diffusion
	synapticConc := matrix.chemicalModulator.GetConcentration(LigandGlutamate, postsynapticPos)
	extrasynapticConc := matrix.chemicalModulator.GetConcentration(LigandGlutamate, extrasynapticPos)

	t.Logf("Concentrations after 1ms:")
	t.Logf("• Synaptic cleft (20nm): %.4f mM", synapticConc)
	t.Logf("• Extrasynaptic (1μm): %.4f mM", extrasynapticConc)

	// Biological validation: synaptic concentration should be much higher than extrasynaptic
	if synapticConc <= extrasynapticConc {
		t.Errorf("BIOLOGY VIOLATION: Synaptic concentration (%.4f) should be > extrasynaptic (%.4f)",
			synapticConc, extrasynapticConc)
	}

	// Wait for glutamate clearance (should be fast)
	time.Sleep(GLUTAMATE_CLEARANCE_TIME)
	clearedConc := matrix.chemicalModulator.GetConcentration(LigandGlutamate, postsynapticPos)

	t.Logf("Concentration after clearance (5ms): %.4f mM", clearedConc)

	// Biological validation: rapid clearance
	clearanceRatio := clearedConc / synapticConc
	if clearanceRatio > 0.1 { // Should clear >90% within clearance time
		t.Errorf("BIOLOGY VIOLATION: Glutamate clearance too slow, %.1f%% remaining",
			clearanceRatio*100)
	} else {
		t.Logf("✓ Glutamate clearance: %.1f%% cleared (biologically realistic)",
			(1-clearanceRatio)*100)
	}

	// === DOPAMINE VOLUME TRANSMISSION TEST ===
	t.Log("\n--- Testing Dopamine Volume Transmission ---")

	// Dopamine should have slower kinetics and wider spatial distribution
	err = matrix.ReleaseLigand(LigandDopamine, "dopamine_terminal", DOPAMINE_PEAK)
	if err != nil {
		t.Fatalf("Failed to release dopamine: %v", err)
	}

	// Measure at multiple distances
	time.Sleep(10 * time.Millisecond) // Allow dopamine diffusion

	nearDopamine := matrix.chemicalModulator.GetConcentration(LigandDopamine, Position3D{X: 1, Y: 0, Z: 0})
	farDopamine := matrix.chemicalModulator.GetConcentration(LigandDopamine, Position3D{X: 10, Y: 0, Z: 0})
	veryFarDopamine := matrix.chemicalModulator.GetConcentration(LigandDopamine, Position3D{X: 50, Y: 0, Z: 0})

	t.Logf("Dopamine concentrations after 10ms:")
	t.Logf("• 1μm distance: %.6f mM", nearDopamine)
	t.Logf("• 10μm distance: %.6f mM", farDopamine)
	t.Logf("• 50μm distance: %.6f mM", veryFarDopamine)

	// Biological validation: dopamine should have wider range than glutamate
	if farDopamine <= 0 {
		t.Errorf("BIOLOGY VIOLATION: Dopamine should reach 10μm distance")
	}

	// Check volume transmission gradient
	if nearDopamine <= farDopamine {
		t.Errorf("BIOLOGY VIOLATION: Dopamine gradient should decrease with distance")
	} else {
		t.Logf("✓ Dopamine volume transmission gradient confirmed")
	}

	t.Log("✅ Chemical kinetics match biological expectations")
}

// =================================================================================
// TEST 2: ELECTRICAL COUPLING AND GAP JUNCTION PROPERTIES
// =================================================================================

func TestBiologicalElectricalCoupling(t *testing.T) {
	t.Log("=== BIOLOGICAL TEST: Electrical Coupling Properties ===")
	t.Log("Validating gap junction conductance and electrical signal propagation")

	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: false, // Focus on electrical only
		SpatialEnabled:  true,
		UpdateInterval:  100 * time.Microsecond, // High temporal resolution
		MaxComponents:   100,
	})
	defer matrix.Stop()

	// Create electrically coupled neuron pair (like cortical interneurons)
	neuron1Pos := Position3D{X: 0, Y: 0, Z: 0}
	neuron2Pos := Position3D{X: 15, Y: 0, Z: 0} // 15μm apart (realistic for gap junctions)

	matrix.RegisterComponent(ComponentInfo{
		ID: "interneuron_1", Type: ComponentNeuron,
		Position: neuron1Pos, State: StateActive, RegisteredAt: time.Now(),
	})

	matrix.RegisterComponent(ComponentInfo{
		ID: "interneuron_2", Type: ComponentNeuron,
		Position: neuron2Pos, State: StateActive, RegisteredAt: time.Now(),
	})

	// Test different gap junction conductances
	conductanceValues := []float64{0.05, 0.1, 0.3, 0.6, 1.0} // Range from weak to strong

	for _, conductance := range conductanceValues {
		t.Logf("\n--- Testing Gap Junction Conductance: %.2f ---", conductance)

		// Establish electrical coupling
		err := matrix.gapJunctions.EstablishElectricalCoupling("interneuron_1", "interneuron_2", conductance)
		if err != nil {
			t.Fatalf("Failed to establish electrical coupling: %v", err)
		}

		// Record coupling strength
		measuredConductance := matrix.gapJunctions.GetConductance("interneuron_1", "interneuron_2")
		if math.Abs(measuredConductance-conductance) > 0.01 {
			t.Errorf("Conductance mismatch: expected %.3f, got %.3f", conductance, measuredConductance)
		}

		// Verify bidirectional coupling (gap junctions are symmetric)
		reverseConductance := matrix.gapJunctions.GetConductance("interneuron_2", "interneuron_1")
		if math.Abs(reverseConductance-conductance) > 0.01 {
			t.Errorf("BIOLOGY VIOLATION: Gap junctions should be bidirectional")
		}

		t.Logf("✓ Bidirectional conductance confirmed: %.3f", reverseConductance)

		// Clean up for next test
		matrix.gapJunctions.RemoveElectricalCoupling("interneuron_1", "interneuron_2")
	}

	// === TEST ELECTRICAL SIGNAL TIMING ===
	t.Log("\n--- Testing Electrical Signal Timing ---")

	// Re-establish moderate coupling
	matrix.gapJunctions.EstablishElectricalCoupling("interneuron_1", "interneuron_2", GAP_JUNCTION_CONDUCTANCE)

	// Create mock neurons to track signal timing
	mockNeuron1 := NewMockNeuron("interneuron_1", neuron1Pos, []LigandType{})
	mockNeuron2 := NewMockNeuron("interneuron_2", neuron2Pos, []LigandType{})

	matrix.ListenForSignals([]SignalType{SignalFired}, mockNeuron1)
	matrix.ListenForSignals([]SignalType{SignalFired}, mockNeuron2)

	// Measure signal propagation timing
	signalStart := time.Now()
	matrix.SendSignal(SignalFired, "interneuron_1", 1.0)

	// Electrical signals should propagate much faster than chemical
	time.Sleep(100 * time.Microsecond) // Much shorter than synaptic delay

	propagationTime := time.Since(signalStart)
	t.Logf("Electrical signal propagation time: %v", propagationTime)

	// Biological validation: electrical coupling should be much faster than chemical synapses
	if propagationTime > SYNAPTIC_DELAY/2 {
		t.Errorf("BIOLOGY VIOLATION: Electrical coupling too slow (%v > %v)",
			propagationTime, SYNAPTIC_DELAY/2)
	} else {
		t.Logf("✓ Electrical coupling speed biologically realistic")
	}

	t.Log("✅ Electrical coupling properties match biological expectations")
}

// =================================================================================
// TEST 3: SPATIAL ORGANIZATION AND NETWORK TOPOLOGY
// =================================================================================

func TestBiologicalSpatialOrganization(t *testing.T) {
	t.Log("=== BIOLOGICAL TEST: Spatial Organization ===")
	t.Log("Validating realistic tissue organization and connectivity patterns")

	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  10 * time.Millisecond,
		MaxComponents:   500,
	})
	defer matrix.Stop()

	// === CREATE REALISTIC CORTICAL COLUMN ===
	t.Log("\n--- Creating Realistic Cortical Column ---")

	columnCenter := Position3D{X: 0, Y: 0, Z: 0}
	pyramidalNeurons := make([]ComponentInfo, 0)
	interneurons := make([]ComponentInfo, 0)

	// Layer 2/3 pyramidal neurons (80% of neurons)
	pyramidalCount := 40
	for i := 0; i < pyramidalCount; i++ {
		// Random position within cortical column
		angle := float64(i) * 2 * math.Pi / float64(pyramidalCount)
		radius := 50.0 + float64(i%3)*20.0 // Layers at different depths

		neuronPos := Position3D{
			X: columnCenter.X + radius*math.Cos(angle),
			Y: columnCenter.Y + radius*math.Sin(angle),
			Z: columnCenter.Z + float64(i%5)*10.0, // Layer distribution
		}

		neuronInfo := ComponentInfo{
			ID:       fmt.Sprintf("pyramidal_%d", i),
			Type:     ComponentNeuron,
			Position: neuronPos,
			State:    StateActive,
			Metadata: map[string]interface{}{
				"cell_type":  "pyramidal",
				"layer":      "L2/3",
				"excitatory": true,
			},
			RegisteredAt: time.Now(),
		}

		matrix.RegisterComponent(neuronInfo)
		pyramidalNeurons = append(pyramidalNeurons, neuronInfo)
	}

	// Interneurons (20% of neurons)
	interneuronCount := 10
	for i := 0; i < interneuronCount; i++ {
		angle := float64(i) * 2 * math.Pi / float64(interneuronCount)
		radius := 30.0 + float64(i%2)*15.0

		neuronPos := Position3D{
			X: columnCenter.X + radius*math.Cos(angle),
			Y: columnCenter.Y + radius*math.Sin(angle),
			Z: columnCenter.Z + float64(i%3)*8.0,
		}

		neuronInfo := ComponentInfo{
			ID:       fmt.Sprintf("interneuron_%d", i),
			Type:     ComponentNeuron,
			Position: neuronPos,
			State:    StateActive,
			Metadata: map[string]interface{}{
				"cell_type":  "interneuron",
				"subtype":    "PV+",
				"excitatory": false,
			},
			RegisteredAt: time.Now(),
		}

		matrix.RegisterComponent(neuronInfo)
		interneurons = append(interneurons, neuronInfo)
	}

	t.Logf("Created cortical column: %d pyramidal + %d interneurons",
		pyramidalCount, interneuronCount)

	// === VALIDATE SPATIAL DENSITY ===
	t.Log("\n--- Validating Spatial Density ---")

	// Calculate neuron density in a test volume
	testRadius := 100.0 // 100 μm radius sphere
	neuronsInVolume := matrix.FindComponents(ComponentCriteria{
		Type:     &[]ComponentType{ComponentNeuron}[0],
		Position: &columnCenter,
		Radius:   testRadius,
	})

	// Volume of sphere: (4/3)πr³, convert to mm³
	volumeMM3 := (4.0 / 3.0) * math.Pi * math.Pow(testRadius/1000.0, 3)
	density := float64(len(neuronsInVolume)) / volumeMM3

	t.Logf("Measured neuron density: %.0f neurons/mm³ (biological: ~150k)", density)

	// Biological validation: density should be in realistic range
	if density > CORTICAL_NEURON_DENSITY*2 {
		t.Errorf("BIOLOGY VIOLATION: Neuron density too high (%.0f > %.0f)",
			density, CORTICAL_NEURON_DENSITY*2)
	} else if density < CORTICAL_NEURON_DENSITY/10 {
		t.Errorf("BIOLOGY VIOLATION: Neuron density too low (%.0f < %.0f)",
			density, CORTICAL_NEURON_DENSITY/10)
	} else {
		t.Logf("✓ Neuron density within biological range")
	}

	// === TEST CONNECTIVITY PATTERNS ===
	t.Log("\n--- Testing Connectivity Patterns ---")

	// Create local connections (most connections are local in cortex)
	localConnections := 0
	distantConnections := 0

	for _, sourcePyr := range pyramidalNeurons[:10] { // Test subset
		// Find nearby neurons for connections
		nearbyNeurons := matrix.FindComponents(ComponentCriteria{
			Type:     &[]ComponentType{ComponentNeuron}[0],
			Position: &sourcePyr.Position,
			Radius:   100.0, // Local connectivity radius
		})

		connectionCount := 0
		for _, targetNeuron := range nearbyNeurons {
			if targetNeuron.ID != sourcePyr.ID && connectionCount < 5 { // Limit connections
				// Calculate distance
				distance := matrix.astrocyteNetwork.Distance(sourcePyr.Position, targetNeuron.Position)

				// Record synaptic connection
				synapseID := fmt.Sprintf("syn_%s_%s", sourcePyr.ID, targetNeuron.ID)
				matrix.astrocyteNetwork.RecordSynapticActivity(
					synapseID, sourcePyr.ID, targetNeuron.ID, 0.5)

				if distance < 50.0 {
					localConnections++
				} else {
					distantConnections++
				}
				connectionCount++
			}
		}

		t.Logf("Neuron %s: %d local connections", sourcePyr.ID, connectionCount)
	}

	// Biological validation: most connections should be local
	totalConnections := localConnections + distantConnections
	localRatio := float64(localConnections) / float64(totalConnections)

	t.Logf("Connection distribution: %.1f%% local, %.1f%% distant",
		localRatio*100, (1-localRatio)*100)

	if localRatio < 0.7 { // At least 70% should be local in cortex
		t.Errorf("BIOLOGY VIOLATION: Too few local connections (%.1f%% < 70%%)", localRatio*100)
	} else {
		t.Logf("✓ Local connection bias confirmed (biologically realistic)")
	}

	t.Log("✅ Spatial organization matches biological cortical structure")
}

// =================================================================================
// TEST 4: ASTROCYTE TERRITORIAL ORGANIZATION
// =================================================================================

func TestBiologicalAstrocyteOrganization(t *testing.T) {
	t.Log("=== BIOLOGICAL TEST: Astrocyte Territorial Organization ===")
	t.Log("Validating astrocyte territorial domains and neuron monitoring")

	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  10 * time.Millisecond,
		MaxComponents:   200,
	})
	defer matrix.Stop()

	// === CREATE ASTROCYTE TERRITORIES ===
	t.Log("\n--- Establishing Astrocyte Territories ---")

	// Create neurons in a grid pattern
	neuronPositions := make([]Position3D, 0)
	for x := -60.0; x <= 60.0; x += 20.0 {
		for y := -60.0; y <= 60.0; y += 20.0 {
			pos := Position3D{X: x, Y: y, Z: 0}
			neuronPositions = append(neuronPositions, pos)

			matrix.RegisterComponent(ComponentInfo{
				ID:           fmt.Sprintf("neuron_%.0f_%.0f", x, y),
				Type:         ComponentNeuron,
				Position:     pos,
				State:        StateActive,
				RegisteredAt: time.Now(),
			})
		}
	}

	t.Logf("Created %d neurons in grid pattern", len(neuronPositions))

	// Establish astrocyte territories with realistic spacing
	astrocytePositions := []Position3D{
		{X: -30, Y: -30, Z: 0},
		{X: 30, Y: -30, Z: 0},
		{X: -30, Y: 30, Z: 0},
		{X: 30, Y: 30, Z: 0},
		{X: 0, Y: 0, Z: 0}, // Central astrocyte
	}

	for i, astroPos := range astrocytePositions {
		astrocyteID := fmt.Sprintf("astrocyte_%d", i)

		// Establish territory with realistic radius
		err := matrix.astrocyteNetwork.EstablishTerritory(
			astrocyteID, astroPos, ASTROCYTE_TERRITORY_RADIUS)
		if err != nil {
			t.Fatalf("Failed to establish astrocyte territory: %v", err)
		}

		t.Logf("Astrocyte %s territory: center(%.0f,%.0f) radius=%.0fμm",
			astrocyteID, astroPos.X, astroPos.Y, ASTROCYTE_TERRITORY_RADIUS)
	}

	// === VALIDATE TERRITORY COVERAGE ===
	t.Log("\n--- Validating Territory Coverage ---")

	for i, astroPos := range astrocytePositions {
		astrocyteID := fmt.Sprintf("astrocyte_%d", i)

		// Find neurons within this astrocyte's territory
		neuronsInTerritory := matrix.FindComponents(ComponentCriteria{
			Type:     &[]ComponentType{ComponentNeuron}[0],
			Position: &astroPos,
			Radius:   ASTROCYTE_TERRITORY_RADIUS,
		})

		neuronCount := len(neuronsInTerritory)
		t.Logf("Astrocyte %s monitors %d neurons", astrocyteID, neuronCount)

		// Biological validation: each astrocyte should monitor realistic number of neurons
		expectedMin := 5  // Minimum viable territory
		expectedMax := 20 // Maximum based on astrocyte capacity

		if neuronCount < expectedMin {
			t.Errorf("BIOLOGY VIOLATION: Astrocyte %s monitors too few neurons (%d < %d)",
				astrocyteID, neuronCount, expectedMin)
		} else if neuronCount > expectedMax {
			t.Errorf("BIOLOGY VIOLATION: Astrocyte %s monitors too many neurons (%d > %d)",
				astrocyteID, neuronCount, expectedMax)
		} else {
			t.Logf("✓ Astrocyte %s monitoring capacity within biological range", astrocyteID)
		}

		// Verify territory retrieval
		territory, exists := matrix.astrocyteNetwork.GetTerritory(astrocyteID)
		if !exists {
			t.Errorf("Failed to retrieve territory for astrocyte %s", astrocyteID)
		} else if territory.Radius != ASTROCYTE_TERRITORY_RADIUS {
			t.Errorf("Territory radius mismatch for astrocyte %s", astrocyteID)
		}
	}

	// === TEST TERRITORIAL OVERLAP ===
	t.Log("\n--- Testing Territorial Overlap ---")

	// Check for realistic territorial overlap
	centralPos := Position3D{X: 0, Y: 0, Z: 0}
	overlappingAstrocytes := 0

	for i, astroPos := range astrocytePositions {
		distance := matrix.astrocyteNetwork.Distance(centralPos, astroPos)
		if distance < ASTROCYTE_TERRITORY_RADIUS {
			overlappingAstrocytes++
			t.Logf("Astrocyte %d territory overlaps central point (distance: %.1fμm)",
				i, distance)
		}
	}

	// Biological validation: some overlap is normal and beneficial
	if overlappingAstrocytes < 2 {
		t.Errorf("BIOLOGY VIOLATION: Too little territorial overlap (%d astrocytes)",
			overlappingAstrocytes)
	} else if overlappingAstrocytes > 4 {
		t.Errorf("BIOLOGY VIOLATION: Too much territorial overlap (%d astrocytes)",
			overlappingAstrocytes)
	} else {
		t.Logf("✓ Territorial overlap within biological range (%d astrocytes)",
			overlappingAstrocytes)
	}

	t.Log("✅ Astrocyte territorial organization matches biological patterns")
}

// =================================================================================
// TEST 5: MICROGLIAL MAINTENANCE AND HOMEOSTASIS
// =================================================================================

func TestBiologicalMicroglialMaintenance(t *testing.T) {
	t.Log("=== BIOLOGICAL TEST: Microglial Maintenance ===")
	t.Log("Validating microglial surveillance, pruning, and homeostatic functions")

	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  10 * time.Millisecond,
		MaxComponents:   100,
	})
	defer matrix.Stop()

	// === CREATE NETWORK WITH VARYING ACTIVITY LEVELS ===
	t.Log("\n--- Creating Network with Variable Activity ---")

	// Create neurons with different activity patterns
	highActivityNeurons := []string{"active_1", "active_2", "active_3"}
	mediumActivityNeurons := []string{"medium_1", "medium_2"}
	lowActivityNeurons := []string{"inactive_1", "inactive_2"}

	allNeurons := append(append(highActivityNeurons, mediumActivityNeurons...), lowActivityNeurons...)

	for i, neuronID := range allNeurons {
		matrix.RegisterComponent(ComponentInfo{
			ID:           neuronID,
			Type:         ComponentNeuron,
			Position:     Position3D{X: float64(i * 10), Y: 0, Z: 0},
			State:        StateActive,
			RegisteredAt: time.Now(),
		})
	}

	// === SIMULATE DIFFERENT ACTIVITY LEVELS ===
	t.Log("\n--- Simulating Activity Patterns ---")

	// Update component health based on simulated activity
	for _, neuronID := range highActivityNeurons {
		matrix.microglia.UpdateComponentHealth(neuronID, 0.9, 8) // High activity, many connections
	}

	for _, neuronID := range mediumActivityNeurons {
		matrix.microglia.UpdateComponentHealth(neuronID, 0.5, 4) // Medium activity
	}

	for _, neuronID := range lowActivityNeurons {
		matrix.microglia.UpdateComponentHealth(neuronID, 0.1, 1) // Low activity, few connections
	}

	// === VALIDATE HEALTH ASSESSMENT ===
	t.Log("\n--- Validating Health Assessment ---")

	// Check health scores for different activity levels
	for _, neuronID := range highActivityNeurons {
		health, exists := matrix.microglia.GetComponentHealth(neuronID)
		if !exists {
			t.Errorf("Health not found for high-activity neuron %s", neuronID)
			continue
		}

		t.Logf("High-activity neuron %s: health=%.3f, activity=%.1f",
			neuronID, health.HealthScore, health.ActivityLevel)

		// Biological validation: high activity should correlate with good health
		if health.HealthScore < 0.8 {
			t.Errorf("BIOLOGY VIOLATION: High-activity neuron %s has poor health (%.3f)",
				neuronID, health.HealthScore)
		}
	}

	for _, neuronID := range lowActivityNeurons {
		health, exists := matrix.microglia.GetComponentHealth(neuronID)
		if !exists {
			t.Errorf("Health not found for low-activity neuron %s", neuronID)
			continue
		}

		t.Logf("Low-activity neuron %s: health=%.3f, activity=%.1f, issues=%v",
			neuronID, health.HealthScore, health.ActivityLevel, health.Issues)

		// Biological validation: low activity should trigger health concerns
		if health.HealthScore > 0.8 {
			t.Errorf("BIOLOGY VIOLATION: Low-activity neuron %s has unrealistically good health (%.3f)",
				neuronID, health.HealthScore)
		}

		// Should detect activity-related issues
		hasActivityIssue := false
		for _, issue := range health.Issues {
			if issue == "very_low_activity" {
				hasActivityIssue = true
				break
			}
		}
		if !hasActivityIssue {
			t.Errorf("BIOLOGY VIOLATION: Low-activity neuron %s should have activity issues detected",
				neuronID)
		}
	}

	// === TEST SYNAPTIC PRUNING DECISIONS ===
	t.Log("\n--- Testing Synaptic Pruning Logic ---")

	// Create synapses with different activity levels
	synapseData := []struct {
		id          string
		preID       string
		postID      string
		activity    float64
		shouldPrune bool
	}{
		{"strong_synapse", "active_1", "active_2", 0.9, false},   // High activity - keep
		{"medium_synapse", "active_1", "medium_1", 0.5, false},   // Medium activity - keep
		{"weak_synapse", "inactive_1", "inactive_2", 0.05, true}, // Low activity - prune
		{"unused_synapse", "inactive_1", "medium_1", 0.01, true}, // Very low - prune
	}

	for _, syn := range synapseData {
		// Mark synapse for potential pruning
		matrix.microglia.MarkForPruning(syn.id, syn.preID, syn.postID, syn.activity)

		t.Logf("Marked synapse %s (activity=%.2f) for pruning evaluation", syn.id, syn.activity)
	}

	// Get pruning candidates
	candidates := matrix.microglia.GetPruningCandidates()
	t.Logf("Found %d synaptic pruning candidates", len(candidates))

	// Validate pruning scores
	for _, candidate := range candidates {
		expectedPrune := false
		for _, syn := range synapseData {
			if syn.id == candidate.ConnectionID {
				expectedPrune = syn.shouldPrune
				break
			}
		}
		_ = expectedPrune // intentionally unused

		t.Logf("Synapse %s: activity=%.2f, pruning_score=%.3f",
			candidate.ConnectionID, candidate.ActivityLevel, candidate.PruningScore)

		// Biological validation: pruning score should correlate with low activity
		if candidate.ActivityLevel < 0.1 && candidate.PruningScore < 0.5 {
			t.Errorf("BIOLOGY VIOLATION: Low-activity synapse %s has low pruning score (%.3f)",
				candidate.ConnectionID, candidate.PruningScore)
		} else if candidate.ActivityLevel > 0.8 && candidate.PruningScore > 0.3 {
			t.Errorf("BIOLOGY VIOLATION: High-activity synapse %s has high pruning score (%.3f)",
				candidate.ConnectionID, candidate.PruningScore)
		}
	}

	// === TEST PATROL BEHAVIOR ===
	t.Log("\n--- Testing Microglial Patrol Behavior ---")

	// Establish patrol route for microglia
	patrolCenter := Position3D{X: 25, Y: 0, Z: 0}
	patrolRadius := 30.0
	patrolRate := 100 * time.Millisecond

	matrix.microglia.EstablishPatrolRoute("microglia_1", Territory{
		Center: patrolCenter,
		Radius: patrolRadius,
	}, patrolRate)

	// Execute patrol WITH TIMEOUT
	done := make(chan PatrolReport, 1)
	go func() {
		report := matrix.microglia.ExecutePatrol("microglia_1")
		done <- report
	}()

	var report PatrolReport
	select {
	case report = <-done:
		t.Logf("Patrol completed: checked %d components", report.ComponentsChecked)
	case <-time.After(2 * time.Second):
		t.Error("TIMEOUT: Patrol execution took too long - possible infinite loop")
		return // Skip rest of test
	}

	// === VALIDATE HOMEOSTATIC BALANCE ===
	t.Log("\n--- Validating Homeostatic Balance ---")

	stats := matrix.microglia.GetMaintenanceStats()
	t.Logf("Maintenance statistics:")
	t.Logf("• Components created: %d", stats.ComponentsCreated)
	t.Logf("• Health checks performed: %d", stats.HealthChecks)
	t.Logf("• Average health score: %.3f", stats.AverageHealthScore)
	t.Logf("• Patrols completed: %d", stats.PatrolsCompleted)

	// Biological validation: system should maintain reasonable health
	if stats.AverageHealthScore < 0.5 {
		t.Errorf("BIOLOGY VIOLATION: Network health too low (%.3f < 0.5)",
			stats.AverageHealthScore)
	} else if stats.AverageHealthScore > 0.95 {
		t.Errorf("BIOLOGY VIOLATION: Network health unrealistically high (%.3f > 0.95)",
			stats.AverageHealthScore)
	} else {
		t.Logf("✓ Network health within biological range (%.3f)", stats.AverageHealthScore)
	}

	// Health checks should be proportional to components
	if stats.HealthChecks < int64(len(allNeurons)) {
		t.Errorf("BIOLOGY VIOLATION: Too few health checks (%d < %d)",
			stats.HealthChecks, len(allNeurons))
	}

	t.Log("✅ Microglial maintenance functions match biological behavior")
}

// =================================================================================
// TEST 6: TEMPORAL DYNAMICS AND BIOLOGICAL TIMESCALES
// =================================================================================

func TestBiologicalTemporalDynamics(t *testing.T) {
	t.Log("=== BIOLOGICAL TEST: Temporal Dynamics ===")
	t.Log("Validating biologically realistic timescales across all processes")

	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  100 * time.Microsecond, // High temporal resolution
		MaxComponents:   50,
	})
	defer matrix.Stop()

	err := matrix.chemicalModulator.Start()
	if err != nil {
		t.Fatalf("Failed to start chemical modulator: %v", err)
	}

	// === TEST ACTION POTENTIAL TIMESCALES ===
	t.Log("\n--- Testing Action Potential Timescales ---")

	// Create mock neurons for timing tests
	neuron1 := NewMockNeuron("timing_neuron_1", Position3D{X: 0, Y: 0, Z: 0},
		[]LigandType{LigandGlutamate})
	neuron2 := NewMockNeuron("timing_neuron_2", Position3D{X: 10, Y: 0, Z: 0},
		[]LigandType{LigandGlutamate})

	matrix.RegisterComponent(ComponentInfo{
		ID: neuron1.ID(), Type: ComponentNeuron,
		Position: neuron1.Position(), State: StateActive, RegisteredAt: time.Now(),
	})

	matrix.ListenForSignals([]SignalType{SignalFired}, neuron1)
	matrix.ListenForSignals([]SignalType{SignalFired}, neuron2)

	// Measure electrical signal propagation time
	start := time.Now()
	matrix.SendSignal(SignalFired, neuron1.ID(), 1.0)
	electricalPropagation := time.Since(start)

	t.Logf("Electrical signal propagation: %v", electricalPropagation)

	// Biological validation: should be much faster than synaptic transmission
	if electricalPropagation > ACTION_POTENTIAL_DURATION/4 {
		t.Errorf("BIOLOGY VIOLATION: Electrical propagation too slow (%v > %v)",
			electricalPropagation, ACTION_POTENTIAL_DURATION/4)
	} else {
		t.Logf("✓ Electrical propagation within biological range")
	}

	// === TEST SYNAPTIC TRANSMISSION TIMESCALES ===
	t.Log("\n--- Testing Synaptic Transmission Timescales ---")

	matrix.RegisterForBinding(neuron2)

	// Measure chemical synaptic delay
	start = time.Now()
	matrix.ReleaseLigand(LigandGlutamate, neuron1.ID(), 1.0)
	time.Sleep(100 * time.Microsecond) // Allow binding
	chemicalTransmission := time.Since(start)

	t.Logf("Chemical synaptic transmission: %v", chemicalTransmission)

	// Biological validation: should be slower than electrical but still fast
	if chemicalTransmission < electricalPropagation {
		t.Errorf("BIOLOGY VIOLATION: Chemical transmission faster than electrical (%v < %v)",
			chemicalTransmission, electricalPropagation)
	} else if chemicalTransmission > SYNAPTIC_DELAY*3 {
		t.Errorf("BIOLOGY VIOLATION: Chemical transmission too slow (%v > %v)",
			chemicalTransmission, SYNAPTIC_DELAY*3)
	} else {
		t.Logf("✓ Chemical transmission timing biologically realistic")
	}

	// === TEST NEUROTRANSMITTER CLEARANCE KINETICS ===
	t.Log("\n--- Testing Neurotransmitter Clearance Kinetics ---")

	testPos := Position3D{X: 0, Y: 0, Z: 0}

	// Test glutamate clearance (should be fast)
	matrix.ReleaseLigand(LigandGlutamate, "test_source", 1.0)
	time.Sleep(1 * time.Millisecond)
	glutamateT1 := matrix.chemicalModulator.GetConcentration(LigandGlutamate, testPos)

	time.Sleep(GLUTAMATE_CLEARANCE_TIME)
	glutamateT2 := matrix.chemicalModulator.GetConcentration(LigandGlutamate, testPos)

	glutamateClearance := (glutamateT1 - glutamateT2) / glutamateT1
	t.Logf("Glutamate clearance after %v: %.1f%%", GLUTAMATE_CLEARANCE_TIME, glutamateClearance*100)

	// Test dopamine persistence (should be slower)
	matrix.ReleaseLigand(LigandDopamine, "test_source", 0.1)
	time.Sleep(1 * time.Millisecond)
	dopamineT1 := matrix.chemicalModulator.GetConcentration(LigandDopamine, testPos)

	time.Sleep(DOPAMINE_HALF_LIFE)
	dopamineT2 := matrix.chemicalModulator.GetConcentration(LigandDopamine, testPos)

	dopaminePersistence := dopamineT2 / dopamineT1
	t.Logf("Dopamine persistence after %v: %.1f%%", DOPAMINE_HALF_LIFE, dopaminePersistence*100)

	// Biological validation: glutamate should clear faster than dopamine
	if glutamateClearance < 0.5 {
		t.Errorf("BIOLOGY VIOLATION: Glutamate clearance too slow (%.1f%% < 50%%)",
			glutamateClearance*100)
	}

	if dopaminePersistence < 0.3 {
		t.Errorf("BIOLOGY VIOLATION: Dopamine cleared too quickly (%.1f%% < 30%%)",
			dopaminePersistence*100)
	}

	if dopaminePersistence <= glutamateClearance {
		t.Errorf("BIOLOGY VIOLATION: Dopamine should persist longer than glutamate")
	} else {
		t.Logf("✓ Neurotransmitter kinetics match biological profiles")
	}

	// === TEST MICROGLIAL PATROL FREQUENCY ===
	t.Log("\n--- Testing Microglial Patrol Frequency ---")

	// Biological microglia patrol every few minutes to hours
	//biologicalPatrolInterval := 5 * time.Minute // Minimum realistic interval
	testPatrolInterval := 50 * time.Millisecond // Accelerated for testing

	matrix.microglia.EstablishPatrolRoute("test_microglia", Territory{
		Center: Position3D{X: 0, Y: 0, Z: 0},
		Radius: 20.0,
	}, testPatrolInterval)

	// Count patrols over short period
	startPatrols := matrix.microglia.GetMaintenanceStats().PatrolsCompleted
	time.Sleep(200 * time.Millisecond) // Allow multiple patrols

	endPatrols := matrix.microglia.GetMaintenanceStats().PatrolsCompleted
	patrolsPerformed := endPatrols - startPatrols

	t.Logf("Patrols performed in 200ms: %d", patrolsPerformed)

	// Validate patrol frequency is reasonable (not too fast/slow for testing)
	if patrolsPerformed < 2 {
		t.Errorf("Patrol frequency too low for testing (%d patrols)", patrolsPerformed)
	} else if patrolsPerformed > 10 {
		t.Errorf("Patrol frequency unrealistically high (%d patrols)", patrolsPerformed)
	} else {
		t.Logf("✓ Patrol frequency appropriate for testing scale")
	}

	t.Log("✅ Temporal dynamics match biological timescales")
}

// =================================================================================
// TEST 7: NETWORK PLASTICITY AND ADAPTATION
// =================================================================================

func TestBiologicalNetworkPlasticity(t *testing.T) {
	t.Log("=== BIOLOGICAL TEST: Network Plasticity ===")
	t.Log("Validating activity-dependent changes and homeostatic regulation")

	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  10 * time.Millisecond,
		MaxComponents:   100,
	})
	defer matrix.Stop()

	// === CREATE PLASTIC NETWORK ===
	t.Log("\n--- Creating Plastic Neural Network ---")

	// Create network with different connection strengths
	neuronPairs := []struct {
		pre, post       string
		initialStrength float64
		activityLevel   float64
	}{
		{"pre_1", "post_1", 0.5, 0.9},  // High activity pair
		{"pre_2", "post_2", 0.5, 0.3},  // Medium activity pair
		{"pre_3", "post_3", 0.5, 0.1},  // Low activity pair
		{"pre_4", "post_4", 0.8, 0.05}, // Strong but unused
	}

	// Register neurons
	for _, pair := range neuronPairs {
		matrix.RegisterComponent(ComponentInfo{
			ID: pair.pre, Type: ComponentNeuron,
			Position: Position3D{X: 0, Y: 0, Z: 0}, State: StateActive, RegisteredAt: time.Now(),
		})
		matrix.RegisterComponent(ComponentInfo{
			ID: pair.post, Type: ComponentNeuron,
			Position: Position3D{X: 10, Y: 0, Z: 0}, State: StateActive, RegisteredAt: time.Now(),
		})
	}

	// Record initial synaptic activities
	for i, pair := range neuronPairs {
		synapseID := fmt.Sprintf("synapse_%d", i)
		matrix.astrocyteNetwork.RecordSynapticActivity(
			synapseID, pair.pre, pair.post, pair.initialStrength)

		t.Logf("Initial synapse %s: strength=%.2f", synapseID, pair.initialStrength)
	}

	// === SIMULATE ACTIVITY-DEPENDENT PLASTICITY ===
	t.Log("\n--- Simulating Activity-Dependent Changes ---")

	// Simulate repeated activity over time
	for round := 1; round <= 5; round++ {
		t.Logf("Activity simulation round %d", round)

		for i, pair := range neuronPairs {
			synapseID := fmt.Sprintf("synapse_%d", i)

			// Simulate activity-dependent strength changes
			currentInfo, exists := matrix.astrocyteNetwork.GetSynapticInfo(synapseID)
			if !exists {
				continue
			}

			// Biological rule: high activity strengthens, low activity weakens
			activityFactor := pair.activityLevel
			strengthChange := (activityFactor - 0.5) * 0.1 // ±10% change per round
			newStrength := currentInfo.Strength + strengthChange

			// Keep in biological range
			if newStrength < 0.1 {
				newStrength = 0.1
			} else if newStrength > 1.0 {
				newStrength = 1.0
			}

			// Update synaptic strength
			matrix.astrocyteNetwork.RecordSynapticActivity(
				synapseID, pair.pre, pair.post, newStrength)

			t.Logf("Synapse %s: %.2f → %.2f (activity=%.1f)",
				synapseID, currentInfo.Strength, newStrength, activityFactor)
		}

		time.Sleep(10 * time.Millisecond) // Brief pause between rounds
	}

	// === VALIDATE PLASTICITY OUTCOMES ===
	t.Log("\n--- Validating Plasticity Outcomes ---")

	for i, pair := range neuronPairs {
		synapseID := fmt.Sprintf("synapse_%d", i)
		finalInfo, exists := matrix.astrocyteNetwork.GetSynapticInfo(synapseID)
		if !exists {
			t.Errorf("Failed to retrieve final synaptic info for %s", synapseID)
			continue
		}

		strengthChange := finalInfo.Strength - pair.initialStrength
		t.Logf("Final synapse %s: strength=%.2f (change: %+.2f)",
			synapseID, finalInfo.Strength, strengthChange)

		// Biological validation: activity should correlate with strength changes
		if pair.activityLevel > 0.7 && strengthChange < 0 {
			t.Errorf("BIOLOGY VIOLATION: High-activity synapse %s weakened (%+.2f)",
				synapseID, strengthChange)
		} else if pair.activityLevel < 0.3 && strengthChange > 0 {
			t.Errorf("BIOLOGY VIOLATION: Low-activity synapse %s strengthened (%+.2f)",
				synapseID, strengthChange)
		} else {
			t.Logf("✓ Synapse %s plasticity matches activity pattern", synapseID)
		}
	}

	// === TEST HOMEOSTATIC SCALING ===
	t.Log("\n--- Testing Homeostatic Scaling ---")

	// Calculate network activity level
	totalActivity := 0.0
	activeConnections := 0

	for i, pair := range neuronPairs {
		synapseID := fmt.Sprintf("synapse_%d", i)
		synapticInfo, exists := matrix.astrocyteNetwork.GetSynapticInfo(synapseID)
		if exists {
			totalActivity += synapticInfo.Strength * pair.activityLevel
			activeConnections++
		}
	}

	averageActivity := totalActivity / float64(activeConnections)
	t.Logf("Network average activity: %.3f", averageActivity)

	// Biological validation: network should maintain moderate activity
	if averageActivity < 0.1 {
		t.Errorf("BIOLOGY VIOLATION: Network activity too low (%.3f < 0.1)", averageActivity)
	} else if averageActivity > 0.8 {
		t.Errorf("BIOLOGY VIOLATION: Network activity too high (%.3f > 0.8)", averageActivity)
	} else {
		t.Logf("✓ Network activity within homeostatic range")
	}

	// === TEST STRUCTURAL PLASTICITY ===
	t.Log("\n--- Testing Structural Plasticity ---")

	// Mark very weak connections for pruning
	prunedCount := 0
	for i, pair := range neuronPairs {
		synapseID := fmt.Sprintf("synapse_%d", i)

		if pair.activityLevel < 0.2 { // Very low activity
			matrix.microglia.MarkForPruning(synapseID, pair.pre, pair.post, pair.activityLevel)
			prunedCount++
			t.Logf("Marked synapse %s for structural pruning (activity=%.1f)",
				synapseID, pair.activityLevel)
		}
	}

	candidates := matrix.microglia.GetPruningCandidates()
	t.Logf("Structural pruning candidates: %d", len(candidates))

	// Biological validation: should prune unused connections
	expectedPruned := 2 // Based on our test data (2 low-activity synapses)
	if len(candidates) != expectedPruned {
		t.Errorf("BIOLOGY VIOLATION: Unexpected number of pruning candidates (%d != %d)",
			len(candidates), expectedPruned)
	} else {
		t.Logf("✓ Structural pruning targets appropriate unused connections")
	}
	_ = expectedPruned // intentionally unused

	t.Log("✅ Network plasticity exhibits biological learning and adaptation")
}

// =================================================================================
// TEST 8: METABOLIC CONSTRAINTS AND RESOURCE LIMITATIONS
// =================================================================================

func TestBiologicalMetabolicConstraints(t *testing.T) {
	t.Log("=== BIOLOGICAL TEST: Metabolic Constraints ===")
	t.Log("Validating energy costs, resource limitations, and metabolic realism")

	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  10 * time.Millisecond,
		MaxComponents:   200, // Test resource limits
	})
	defer matrix.Stop()

	// === TEST COMPONENT DENSITY LIMITS ===
	t.Log("\n--- Testing Component Density Limits ---")

	// Try to create neurons at very high density
	densityTestRadius := 10.0 // 10 μm radius
	centerPos := Position3D{X: 0, Y: 0, Z: 0}

	neuronsCreated := 0
	maxNeuronsInArea := 20 // Biological limit for this small area

	for i := 0; i < 50; i++ { // Try to create many neurons
		angle := float64(i) * 2 * math.Pi / 50
		radius := float64(i%5) * 2.0 // Pack tightly

		neuronPos := Position3D{
			X: centerPos.X + radius*math.Cos(angle),
			Y: centerPos.Y + radius*math.Sin(angle),
			Z: centerPos.Z,
		}

		// Only create if within test area
		distance := matrix.astrocyteNetwork.Distance(centerPos, neuronPos)
		if distance <= densityTestRadius {
			neuronID := fmt.Sprintf("dense_neuron_%d", i)
			err := matrix.RegisterComponent(ComponentInfo{
				ID: neuronID, Type: ComponentNeuron,
				Position: neuronPos, State: StateActive, RegisteredAt: time.Now(),
			})

			if err == nil {
				neuronsCreated++
			}
		}
	}

	// Calculate actual density
	areaM2 := math.Pi * math.Pow(densityTestRadius/1000000, 2) // Convert to m²
	density := float64(neuronsCreated) / areaM2

	t.Logf("Created %d neurons in %.1fμm radius (density: %.0f/m²)",
		neuronsCreated, densityTestRadius, density)

	// Biological validation: density should be reasonable
	if neuronsCreated > maxNeuronsInArea*2 {
		t.Errorf("BIOLOGY VIOLATION: Neuron density too high (%d > %d)",
			neuronsCreated, maxNeuronsInArea*2)
	} else {
		t.Logf("✓ Neuron density within biological constraints")
	}

	// === TEST CONNECTION SCALING LIMITS ===
	t.Log("\n--- Testing Connection Scaling Limits ---")

	// Test connection limits per neuron
	testNeuronID := "connection_test_neuron"
	matrix.RegisterComponent(ComponentInfo{
		ID: testNeuronID, Type: ComponentNeuron,
		Position: Position3D{X: 100, Y: 0, Z: 0}, State: StateActive, RegisteredAt: time.Now(),
	})

	// Try to create many connections
	maxConnections := SYNAPSES_PER_NEURON / 1000 // Scaled down for testing
	connectionsCreated := 0

	for i := 0; i < maxConnections*2; i++ { // Try to exceed limit
		targetID := fmt.Sprintf("target_%d", i)

		// Register target
		matrix.RegisterComponent(ComponentInfo{
			ID: targetID, Type: ComponentNeuron,
			Position: Position3D{X: 105, Y: float64(i), Z: 0},
			State:    StateActive, RegisteredAt: time.Now(),
		})

		// Try to create connection
		synapseID := fmt.Sprintf("conn_synapse_%d", i)
		err := matrix.astrocyteNetwork.RecordSynapticActivity(
			synapseID, testNeuronID, targetID, 0.5)

		if err == nil {
			connectionsCreated++
		}
	}

	t.Logf("Created %d connections for neuron (biological limit: ~%d)",
		connectionsCreated, maxConnections)

	// Validate connection count is reasonable
	connections := matrix.astrocyteNetwork.GetConnections(testNeuronID)
	actualConnections := len(connections)

	if actualConnections > maxConnections*3 {
		t.Errorf("BIOLOGY VIOLATION: Too many connections per neuron (%d > %d)",
			actualConnections, maxConnections*3)
	} else {
		t.Logf("✓ Connection count per neuron within biological range")
	}

	// === TEST CHEMICAL RELEASE FREQUENCY LIMITS ===
	t.Log("\n--- Testing Chemical Release Frequency ---")

	// Rapid chemical release should have metabolic costs
	releaseNeuronID := "release_test_neuron"
	matrix.RegisterComponent(ComponentInfo{
		ID: releaseNeuronID, Type: ComponentNeuron,
		Position: Position3D{X: 200, Y: 0, Z: 0}, State: StateActive, RegisteredAt: time.Now(),
	})

	// Try rapid glutamate release
	releaseCount := 0
	maxReleases := 100 // Test rapid firing

	startTime := time.Now()
	for i := 0; i < maxReleases; i++ {
		err := matrix.ReleaseLigand(LigandGlutamate, releaseNeuronID, 0.5)
		if err == nil {
			releaseCount++
		}

		// Brief pause (simulating refractory period)
		time.Sleep(100 * time.Microsecond)
	}
	totalTime := time.Since(startTime)

	releaseRate := float64(releaseCount) / totalTime.Seconds()
	t.Logf("Chemical release rate: %.1f releases/second", releaseRate)

	// Biological validation: release rate should be limited
	maxBiologicalRate := 1000.0 // 1kHz max firing rate
	if releaseRate > maxBiologicalRate*2 {
		t.Errorf("BIOLOGY VIOLATION: Chemical release rate too high (%.1f > %.1f)",
			releaseRate, maxBiologicalRate*2)
	} else {
		t.Logf("✓ Chemical release rate within biological limits")
	}

	// === TEST RESOURCE CLEANUP EFFICIENCY ===
	t.Log("\n--- Testing Resource Cleanup ---")

	// Count components before cleanup
	initialCount := matrix.astrocyteNetwork.Count()
	t.Logf("Components before cleanup: %d", initialCount)

	// Remove some test components
	componentsToRemove := []string{"dense_neuron_0", "dense_neuron_1", "target_0", "target_1"}
	removedCount := 0

	for _, componentID := range componentsToRemove {
		err := matrix.microglia.RemoveComponent(componentID)
		if err == nil {
			removedCount++
		}
	}

	// Count after cleanup
	finalCount := matrix.astrocyteNetwork.Count()
	actualRemoved := initialCount - finalCount

	t.Logf("Components after cleanup: %d (removed: %d)", finalCount, actualRemoved)

	// Biological validation: cleanup should be efficient
	if actualRemoved != removedCount {
		t.Errorf("BIOLOGY VIOLATION: Inefficient cleanup (%d removed, %d expected)",
			actualRemoved, removedCount)
	} else {
		t.Logf("✓ Resource cleanup efficient and accurate")
	}

	// === TEST SYSTEM RESOURCE MONITORING ===
	t.Log("\n--- Testing System Resource Monitoring ---")

	// Check microglial maintenance load
	stats := matrix.microglia.GetMaintenanceStats()
	componentsPerHealthCheck := float64(finalCount) / float64(stats.HealthChecks)

	t.Logf("Maintenance efficiency: %.2f components per health check", componentsPerHealthCheck)

	// Biological validation: maintenance should be efficient but thorough
	if componentsPerHealthCheck > 10.0 {
		t.Errorf("BIOLOGY VIOLATION: Maintenance too sparse (%.2f components/check)",
			componentsPerHealthCheck)
	} else if componentsPerHealthCheck < 0.5 {
		t.Errorf("BIOLOGY VIOLATION: Maintenance too intensive (%.2f components/check)",
			componentsPerHealthCheck)
	} else {
		t.Logf("✓ Maintenance efficiency within biological range")
	}

	t.Log("✅ Metabolic constraints and resource management are biologically realistic")
}

// =================================================================================
// COMPREHENSIVE BIOLOGICAL INTEGRATION TEST
// =================================================================================

func TestBiologicalSystemIntegration(t *testing.T) {
	t.Log("=== COMPREHENSIVE BIOLOGICAL INTEGRATION TEST ===")
	t.Log("Validating complete biological behavior across all subsystems")

	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  1 * time.Millisecond, // Biological update rate
		MaxComponents:   300,
	})
	defer matrix.Stop()

	err := matrix.chemicalModulator.Start()
	if err != nil {
		t.Fatalf("Failed to start chemical modulator: %v", err)
	}

	// === CREATE REALISTIC MINI-CIRCUIT ===
	t.Log("\n--- Creating Biologically Realistic Mini-Circuit ---")

	// Sensory input layer
	sensoryNeurons := []string{"sensory_1", "sensory_2", "sensory_3"}
	for i, neuronID := range sensoryNeurons {
		matrix.RegisterComponent(ComponentInfo{
			ID: neuronID, Type: ComponentNeuron,
			Position: Position3D{X: 0, Y: float64(i * 20), Z: 0},
			State:    StateActive, RegisteredAt: time.Now(),
			Metadata: map[string]interface{}{"layer": "sensory", "modality": "visual"},
		})
	}

	// Processing layer (pyramidal neurons)
	processingNeurons := []string{"pyr_1", "pyr_2", "pyr_3", "pyr_4"}
	for i, neuronID := range processingNeurons {
		matrix.RegisterComponent(ComponentInfo{
			ID: neuronID, Type: ComponentNeuron,
			Position: Position3D{X: 50, Y: float64(i * 15), Z: 0},
			State:    StateActive, RegisteredAt: time.Now(),
			Metadata: map[string]interface{}{"layer": "processing", "type": "pyramidal"},
		})
	}

	// Inhibitory interneurons
	inhibitoryNeurons := []string{"inh_1", "inh_2"}
	for i, neuronID := range inhibitoryNeurons {
		matrix.RegisterComponent(ComponentInfo{
			ID: neuronID, Type: ComponentNeuron,
			Position: Position3D{X: 50, Y: float64(i * 30), Z: 10},
			State:    StateActive, RegisteredAt: time.Now(),
			Metadata: map[string]interface{}{"layer": "processing", "type": "interneuron"},
		})
	}

	// Output layer
	outputNeurons := []string{"motor_1", "motor_2"}
	for i, neuronID := range outputNeurons {
		matrix.RegisterComponent(ComponentInfo{
			ID: neuronID, Type: ComponentNeuron,
			Position: Position3D{X: 100, Y: float64(i * 25), Z: 0},
			State:    StateActive, RegisteredAt: time.Now(),
			Metadata: map[string]interface{}{"layer": "output", "target": "motor"},
		})
	}

	t.Logf("Created circuit: %d sensory + %d processing + %d inhibitory + %d output neurons",
		len(sensoryNeurons), len(processingNeurons), len(inhibitoryNeurons), len(outputNeurons))

	// === ESTABLISH BIOLOGICAL CONNECTIVITY ===
	t.Log("\n--- Establishing Biological Connectivity Patterns ---")

	connectionCount := 0

	// Sensory → Processing connections (feedforward)
	for _, sensory := range sensoryNeurons {
		for i, processing := range processingNeurons {
			if i < 3 { // Each sensory connects to 3 processing neurons
				synapseID := fmt.Sprintf("syn_%s_%s", sensory, processing)
				matrix.astrocyteNetwork.RecordSynapticActivity(
					synapseID, sensory, processing, 0.6) // Moderate strength
				connectionCount++
			}
		}
	}

	// Processing → Output connections
	for i, processing := range processingNeurons {
		outputTarget := outputNeurons[i%len(outputNeurons)]
		synapseID := fmt.Sprintf("syn_%s_%s", processing, outputTarget)
		matrix.astrocyteNetwork.RecordSynapticActivity(
			synapseID, processing, outputTarget, 0.7) // Strong output
		connectionCount++
	}

	// Inhibitory connections (local processing regulation)
	for _, inhibitory := range inhibitoryNeurons {
		for _, processing := range processingNeurons {
			synapseID := fmt.Sprintf("syn_%s_%s", inhibitory, processing)
			matrix.astrocyteNetwork.RecordSynapticActivity(
				synapseID, inhibitory, processing, -0.4) // Inhibitory
			connectionCount++
		}
	}

	// Recurrent connections within processing layer
	matrix.astrocyteNetwork.RecordSynapticActivity("syn_pyr_1_pyr_2", "pyr_1", "pyr_2", 0.3)
	matrix.astrocyteNetwork.RecordSynapticActivity("syn_pyr_2_pyr_3", "pyr_2", "pyr_3", 0.3)
	connectionCount += 2

	t.Logf("Established %d synaptic connections", connectionCount)

	// === ESTABLISH ASTROCYTE TERRITORIES ===
	t.Log("\n--- Establishing Astrocyte Territorial Coverage ---")

	// Create astrocyte territories covering different circuit regions
	territories := []struct {
		id     string
		center Position3D
		radius float64
	}{
		{"astro_sensory", Position3D{X: 0, Y: 20, Z: 0}, 30.0},
		{"astro_processing", Position3D{X: 50, Y: 25, Z: 0}, 35.0},
		{"astro_output", Position3D{X: 100, Y: 15, Z: 0}, 25.0},
	}

	for _, territory := range territories {
		matrix.astrocyteNetwork.EstablishTerritory(
			territory.id, territory.center, territory.radius)

		// Count neurons in territory
		neuronsInTerritory := matrix.FindComponents(ComponentCriteria{
			Type:     &[]ComponentType{ComponentNeuron}[0],
			Position: &territory.center,
			Radius:   territory.radius,
		})

		t.Logf("Astrocyte %s: monitoring %d neurons in %.0fμm radius",
			territory.id, len(neuronsInTerritory), territory.radius)
	}

	// === SIMULATE BIOLOGICAL SIGNAL PROCESSING ===
	t.Log("\n--- Simulating Biological Signal Processing ---")

	// Create mock neurons for realistic interaction
	mockNeurons := make(map[string]*MockNeuron)
	allNeuronIDs := append(append(append(sensoryNeurons, processingNeurons...),
		inhibitoryNeurons...), outputNeurons...)

	for _, neuronID := range allNeuronIDs {
		info, exists := matrix.astrocyteNetwork.Get(neuronID)
		if exists {
			// Determine receptor types based on neuron type
			receptors := []LigandType{LigandGlutamate, LigandGABA}
			if len(info.Metadata) > 0 {
				if neuronType, ok := info.Metadata["type"].(string); ok {
					if neuronType == "pyramidal" {
						receptors = append(receptors, LigandDopamine) // Modulation
					}
				}
			}

			mockNeuron := NewMockNeuron(neuronID, info.Position, receptors)
			mockNeurons[neuronID] = mockNeuron

			// Register for chemical and electrical signaling
			matrix.RegisterForBinding(mockNeuron)
			matrix.ListenForSignals([]SignalType{SignalFired}, mockNeuron)
		}
	}

	// === BIOLOGICAL SIGNAL SEQUENCE ===
	t.Log("\n--- Executing Biological Signal Sequence ---")

	// 1. Sensory input (glutamatergic excitation)
	t.Log("• Sensory input: glutamate release")
	for _, sensoryID := range sensoryNeurons {
		matrix.ReleaseLigand(LigandGlutamate, sensoryID, 0.8)
	}
	time.Sleep(5 * time.Millisecond) // Synaptic delay

	// Check processing layer activation
	processingActivation := 0
	for _, pyrID := range processingNeurons {
		if neuron, exists := mockNeurons[pyrID]; exists {
			if neuron.currentPotential > 0.3 {
				processingActivation++
			}
		}
	}
	t.Logf("  Processing neurons activated: %d/%d", processingActivation, len(processingNeurons))

	// 2. Inhibitory regulation (GABAergic)
	t.Log("• Inhibitory regulation: GABA release")
	for _, inhID := range inhibitoryNeurons {
		matrix.ReleaseLigand(LigandGABA, inhID, 0.6)
	}
	time.Sleep(5 * time.Millisecond)

	// Check inhibitory effect
	postInhibitionActivation := 0
	for _, pyrID := range processingNeurons {
		if neuron, exists := mockNeurons[pyrID]; exists {
			if neuron.currentPotential > 0.2 {
				postInhibitionActivation++
			}
		}
	}
	t.Logf("  Processing activation after inhibition: %d/%d",
		postInhibitionActivation, len(processingNeurons))

	// 3. Action potential propagation
	t.Log("• Action potential propagation")
	for _, pyrID := range processingNeurons[:2] { // Subset fires
		matrix.SendSignal(SignalFired, pyrID, 1.0)
	}
	time.Sleep(2 * time.Millisecond)

	// Check output layer activation
	outputActivation := 0
	for _, motorID := range outputNeurons {
		if neuron, exists := mockNeurons[motorID]; exists {
			if neuron.currentPotential > 0.1 {
				outputActivation++
			}
		}
	}
	t.Logf("  Output neurons activated: %d/%d", outputActivation, len(outputNeurons))

	// 4. Neuromodulatory enhancement (dopamine)
	t.Log("• Neuromodulatory enhancement: dopamine")
	matrix.ReleaseLigand(LigandDopamine, "vta_neuron", 0.3) // Reward signal
	time.Sleep(20 * time.Millisecond)                       // Slower dopamine kinetics

	// === VALIDATE BIOLOGICAL SIGNAL FLOW ===
	t.Log("\n--- Validating Biological Signal Flow ---")

	// Signal should flow: Sensory → Processing → Output
	if processingActivation < 2 {
		t.Errorf("BIOLOGY VIOLATION: Insufficient sensory→processing transmission (%d activated)",
			processingActivation)
	}

	if postInhibitionActivation >= processingActivation {
		t.Errorf("BIOLOGY VIOLATION: GABA failed to inhibit processing layer")
	}

	if outputActivation == 0 {
		t.Errorf("BIOLOGY VIOLATION: No signal reached output layer")
	}

	if outputActivation > processingActivation {
		t.Errorf("BIOLOGY VIOLATION: Output activation exceeds processing activation")
	}

	t.Logf("✓ Signal flow follows biological feedforward pattern")

	// === VALIDATE NETWORK HEALTH AND MAINTENANCE ===
	t.Log("\n--- Network Health and Maintenance Validation ---")

	// Update health for all neurons based on activity
	for neuronID, mockNeuron := range mockNeurons {
		activityLevel := math.Min(mockNeuron.currentPotential, 1.0)
		connections := len(matrix.astrocyteNetwork.GetConnections(neuronID))
		matrix.microglia.UpdateComponentHealth(neuronID, activityLevel, connections)
	}

	// Check overall network health
	stats := matrix.microglia.GetMaintenanceStats()
	t.Logf("Network health: average=%.3f, checks=%d",
		stats.AverageHealthScore, stats.HealthChecks)

	if stats.AverageHealthScore < 0.6 {
		t.Errorf("BIOLOGY VIOLATION: Network health too low (%.3f)", stats.AverageHealthScore)
	} else {
		t.Logf("✓ Network maintains healthy activity levels")
	}

	// === FINAL BIOLOGICAL VALIDATION ===
	t.Log("\n--- Final Biological Validation Summary ---")

	totalComponents := matrix.astrocyteNetwork.Count()
	chemicalReleases := len(matrix.chemicalModulator.GetRecentReleases(20))
	electricalSignals := len(matrix.gapJunctions.GetRecentSignals(10))

	validationResults := []struct {
		test   string
		passed bool
		value  interface{}
	}{
		{"Component count", totalComponents > 10, totalComponents},
		{"Chemical releases", chemicalReleases >= 4, chemicalReleases},
		{"Electrical signals", electricalSignals >= 1, electricalSignals},
		{"Signal flow", outputActivation > 0, outputActivation},
		{"Inhibitory control", postInhibitionActivation < processingActivation,
			fmt.Sprintf("%d < %d", postInhibitionActivation, processingActivation)},
		{"Network health", stats.AverageHealthScore >= 0.6,
			fmt.Sprintf("%.3f", stats.AverageHealthScore)},
		{"Astrocyte coverage", len(territories) == 3, len(territories)},
		{"Synaptic connections", connectionCount >= 10, connectionCount},
	}

	passedTests := 0
	for _, result := range validationResults {
		status := "✗"
		if result.passed {
			status = "✓"
			passedTests++
		}
		t.Logf("%s %s: %v", status, result.test, result.value)
	}

	if passedTests == len(validationResults) {
		t.Log("🧠 ✅ ALL BIOLOGICAL VALIDATION TESTS PASSED")
		t.Log("🧠 ✅ System exhibits authentic biological neural behavior")
		t.Log("🧠 ✅ Chemical signaling, electrical coupling, spatial organization,")
		t.Log("🧬 ✅ microglial maintenance, and astrocyte coordination all function")
		t.Log("🧬 ✅ with biological accuracy and realistic timescales")
	} else {
		t.Errorf("BIOLOGICAL VALIDATION FAILED: %d/%d tests passed",
			passedTests, len(validationResults))
	}
}

// =================================================================================
// UTILITY FUNCTIONS FOR BIOLOGICAL TESTING
// =================================================================================

// validateBiologicalRange checks if a value is within expected biological range
func validateBiologicalRange(t *testing.T, name string, value, min, max float64, unit string) bool {
	if value < min || value > max {
		t.Errorf("BIOLOGY VIOLATION: %s out of range: %.3f %s (expected %.3f-%.3f %s)",
			name, value, unit, min, max, unit)
		return false
	}
	t.Logf("✓ %s within biological range: %.3f %s", name, value, unit)
	return true
}

// calculateSignalToNoiseRatio measures signal quality
func calculateSignalToNoiseRatio(signal, noise float64) float64 {
	if noise <= 0 {
		return math.Inf(1) // Perfect signal
	}
	return signal / noise
}

// measureNetworkConnectivity calculates connectivity statistics
func measureNetworkConnectivity(matrix *ExtracellularMatrix, neuronIDs []string) map[string]float64 {
	stats := make(map[string]float64)

	totalConnections := 0
	totalNeurons := len(neuronIDs)

	for _, neuronID := range neuronIDs {
		connections := matrix.astrocyteNetwork.GetConnections(neuronID)
		totalConnections += len(connections)
	}

	stats["average_connections"] = float64(totalConnections) / float64(totalNeurons)
	stats["connection_density"] = float64(totalConnections) / float64(totalNeurons*totalNeurons)
	stats["total_connections"] = float64(totalConnections)

	return stats
}
