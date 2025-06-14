/*
=================================================================================
CHEMICAL MODULATOR - FOCUSED TESTS
=================================================================================

Focused tests for the chemical signaling system to debug and validate
neurotransmitter release, diffusion, binding, and concentration calculations.

These tests isolate chemical system functionality from other matrix components
to identify and fix issues with concentration field management.
=================================================================================
*/

package extracellular

import (
	"math"
	"testing"
	"time"
)

// =================================================================================
// TEST 1: BASIC CHEMICAL MODULATOR FUNCTIONALITY
// =================================================================================

func TestChemicalModulatorBasic(t *testing.T) {
	t.Log("=== BASIC CHEMICAL MODULATOR TEST ===")

	// Create minimal astrocyte network for chemical modulator
	astrocyteNetwork := NewAstrocyteNetwork()

	// Create chemical modulator
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Test that modulator is created properly
	if modulator == nil {
		t.Fatal("Failed to create chemical modulator")
	}

	// Test ligand kinetics initialization
	if len(modulator.ligandKinetics) == 0 {
		t.Error("Ligand kinetics not initialized")
	}

	// Check specific ligand kinetics
	glutamateKinetics, exists := modulator.ligandKinetics[LigandGlutamate]
	if !exists {
		t.Error("Glutamate kinetics not found")
	} else {
		t.Logf("Glutamate kinetics: diffusion=%.3f, decay=%.3f, range=%.1f",
			glutamateKinetics.DiffusionRate, glutamateKinetics.DecayRate, glutamateKinetics.MaxRange)

		// Validate kinetics are reasonable
		if glutamateKinetics.DiffusionRate <= 0 {
			t.Error("Invalid glutamate diffusion rate")
		}
		if glutamateKinetics.DecayRate <= 0 {
			t.Error("Invalid glutamate decay rate")
		}
	}

	dopamineKinetics, exists := modulator.ligandKinetics[LigandDopamine]
	if !exists {
		t.Error("Dopamine kinetics not found")
	} else {
		t.Logf("Dopamine kinetics: diffusion=%.3f, decay=%.3f, range=%.1f",
			dopamineKinetics.DiffusionRate, dopamineKinetics.DecayRate, dopamineKinetics.MaxRange)

		// Biological validation: dopamine should be slower and longer-range than glutamate
		if dopamineKinetics.DiffusionRate >= glutamateKinetics.DiffusionRate {
			t.Error("Dopamine should diffuse slower than glutamate")
		}
		if dopamineKinetics.DecayRate >= glutamateKinetics.DecayRate {
			t.Error("Dopamine should decay slower than glutamate")
		}
		if dopamineKinetics.MaxRange <= glutamateKinetics.MaxRange {
			t.Error("Dopamine should have longer range than glutamate")
		}
	}

	t.Log("✓ Chemical modulator basic functionality working")
}

// =================================================================================
// TEST 2: CHEMICAL RELEASE AND EVENT TRACKING
// =================================================================================

func TestChemicalReleaseAndTracking(t *testing.T) {
	t.Log("=== CHEMICAL RELEASE AND TRACKING TEST ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Register a source neuron
	sourcePos := Position3D{X: 0, Y: 0, Z: 0}
	astrocyteNetwork.Register(ComponentInfo{
		ID: "test_neuron", Type: ComponentNeuron,
		Position: sourcePos, State: StateActive, RegisteredAt: time.Now(),
	})

	// Test chemical release
	err := modulator.Release(LigandGlutamate, "test_neuron", 1.0)
	if err != nil {
		t.Fatalf("Failed to release glutamate: %v", err)
	}

	// Check that release event was recorded
	recentReleases := modulator.GetRecentReleases(5)
	if len(recentReleases) == 0 {
		t.Error("No release events recorded")
	} else {
		release := recentReleases[0]
		t.Logf("Release event: ligand=%v, concentration=%.3f, source=%s",
			release.LigandType, release.Concentration, release.SourceID)

		if release.LigandType != LigandGlutamate {
			t.Error("Wrong ligand type recorded")
		}
		if release.Concentration != 1.0 {
			t.Error("Wrong concentration recorded")
		}
		if release.SourceID != "test_neuron" {
			t.Error("Wrong source ID recorded")
		}
	}

	// Test release from unknown source (should work with default position)
	err = modulator.Release(LigandDopamine, "unknown_source", 0.5)
	if err != nil {
		t.Fatalf("Failed to release from unknown source: %v", err)
	}

	recentReleases = modulator.GetRecentReleases(5)
	if len(recentReleases) < 2 {
		t.Error("Second release not recorded")
	}

	t.Log("✓ Chemical release and tracking working")
}

// =================================================================================
// TEST 3: CONCENTRATION FIELD MANAGEMENT
// =================================================================================

func TestConcentrationFieldManagement(t *testing.T) {
	t.Log("=== CONCENTRATION FIELD MANAGEMENT TEST ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Register source neuron
	sourcePos := Position3D{X: 10, Y: 20, Z: 30}
	astrocyteNetwork.Register(ComponentInfo{
		ID: "source_neuron", Type: ComponentNeuron,
		Position: sourcePos, State: StateActive, RegisteredAt: time.Now(),
	})

	// Release chemical and check concentration field creation
	err := modulator.Release(LigandGlutamate, "source_neuron", 0.8)
	if err != nil {
		t.Fatalf("Failed to release glutamate: %v", err)
	}

	// Debug: Check if concentration fields exist
	t.Logf("DEBUG: Number of concentration fields: %d", len(modulator.concentrationFields))

	if field, exists := modulator.concentrationFields[LigandGlutamate]; exists {
		t.Logf("DEBUG: Glutamate field has %d concentration points", len(field.Concentrations))
		t.Logf("DEBUG: Glutamate field max concentration: %.4f", field.MaxConcentration)

		// Check if concentration was recorded at source position
		if conc, exists := field.Concentrations[sourcePos]; exists {
			t.Logf("DEBUG: Concentration at source position: %.4f", conc)
		} else {
			t.Error("No concentration recorded at source position")
		}
	} else {
		t.Error("No concentration field created for glutamate")
	}

	// Test concentration retrieval at source position
	sourceConc := modulator.GetConcentration(LigandGlutamate, sourcePos)
	t.Logf("GetConcentration at source: %.4f", sourceConc)

	if sourceConc <= 0 {
		t.Error("No concentration detected at source position")
	}

	// Test concentration at nearby position
	nearbyPos := Position3D{X: 10.1, Y: 20.1, Z: 30.1}
	nearbyConc := modulator.GetConcentration(LigandGlutamate, nearbyPos)
	t.Logf("GetConcentration nearby: %.4f", nearbyConc)

	// Test concentration at distant position
	distantPos := Position3D{X: 50, Y: 50, Z: 50}
	distantConc := modulator.GetConcentration(LigandGlutamate, distantPos)
	t.Logf("GetConcentration distant: %.4f", distantConc)

	t.Log("✓ Concentration field management test completed")
}

// =================================================================================
// TEST 4: CONCENTRATION CALCULATION ALGORITHM
// =================================================================================

func TestConcentrationCalculationAlgorithm(t *testing.T) {
	t.Log("=== CONCENTRATION CALCULATION ALGORITHM TEST ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Test the distance-based concentration calculation directly
	// testPos := Position3D{X: 0, Y: 0, Z: 0}

	// Test calculation at different distances for glutamate
	distances := []float64{0.0, 0.1, 1.0, 5.0, 10.0, 50.0}
	sourceConcentration := 1.0

	t.Log("Distance-based concentration calculation for glutamate:")
	for _, distance := range distances {
		conc := modulator.calculateConcentrationAtDistance(LigandGlutamate, sourceConcentration, distance)
		t.Logf("Distance %.1fμm: concentration %.6f", distance, conc)

		// Validate that concentration decreases with distance
		if distance > 0 && conc >= sourceConcentration {
			t.Errorf("Concentration should decrease with distance (%.6f at %.1fμm)", conc, distance)
		}

		// Check that concentration is zero beyond max range
		glutamateKinetics := modulator.ligandKinetics[LigandGlutamate]
		if distance > glutamateKinetics.MaxRange && conc > 0 {
			t.Errorf("Concentration should be zero beyond max range (%.6f at %.1fμm > %.1fμm)",
				conc, distance, glutamateKinetics.MaxRange)
		}
	}

	// Compare glutamate vs dopamine diffusion
	t.Log("\nComparing glutamate vs dopamine diffusion:")
	testDistance := 10.0
	glutamateConc := modulator.calculateConcentrationAtDistance(LigandGlutamate, sourceConcentration, testDistance)
	dopamineConc := modulator.calculateConcentrationAtDistance(LigandDopamine, sourceConcentration, testDistance)

	t.Logf("At %.1fμm: glutamate=%.6f, dopamine=%.6f", testDistance, glutamateConc, dopamineConc)

	// Dopamine should have higher concentration at distance due to slower decay and larger range
	if dopamineConc <= glutamateConc {
		t.Errorf("Dopamine should have higher concentration at distance (%.6f vs %.6f)",
			dopamineConc, glutamateConc)
	}

	t.Log("✓ Concentration calculation algorithm working correctly")
}

// =================================================================================
// TEST 5: BINDING TARGET REGISTRATION AND SIGNALING
// =================================================================================

func TestBindingTargetSystem(t *testing.T) {
	t.Log("=== BINDING TARGET SYSTEM TEST ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Create mock binding targets
	target1 := NewMockNeuron("target_1", Position3D{X: 0, Y: 0, Z: 0},
		[]LigandType{LigandGlutamate, LigandGABA})
	target2 := NewMockNeuron("target_2", Position3D{X: 5, Y: 0, Z: 0},
		[]LigandType{LigandDopamine})

	// Register binding targets
	err := modulator.RegisterTarget(target1)
	if err != nil {
		t.Fatalf("Failed to register target1: %v", err)
	}

	err = modulator.RegisterTarget(target2)
	if err != nil {
		t.Fatalf("Failed to register target2: %v", err)
	}

	// Check binding target registration
	glutamateTargets := modulator.bindingTargets[LigandGlutamate]
	if len(glutamateTargets) != 1 {
		t.Errorf("Expected 1 glutamate target, got %d", len(glutamateTargets))
	}

	dopamineTargets := modulator.bindingTargets[LigandDopamine]
	if len(dopamineTargets) != 1 {
		t.Errorf("Expected 1 dopamine target, got %d", len(dopamineTargets))
	}

	// Register source neuron
	astrocyteNetwork.Register(ComponentInfo{
		ID: "source", Type: ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0}, State: StateActive, RegisteredAt: time.Now(),
	})

	// Test chemical release and binding
	initialPotential1 := target1.currentPotential
	initialPotential2 := target2.currentPotential

	t.Logf("Initial potentials: target1=%.3f, target2=%.3f", initialPotential1, initialPotential2)

	// Release glutamate (should affect target1 only)
	err = modulator.Release(LigandGlutamate, "source", 1.0)
	if err != nil {
		t.Fatalf("Failed to release glutamate: %v", err)
	}

	finalPotential1 := target1.currentPotential
	finalPotential2 := target2.currentPotential

	t.Logf("After glutamate: target1=%.3f, target2=%.3f", finalPotential1, finalPotential2)

	// Validate binding selectivity
	if finalPotential1 <= initialPotential1 {
		t.Error("Target1 should respond to glutamate")
	}

	if finalPotential2 != initialPotential2 {
		t.Error("Target2 should not respond to glutamate")
	}

	// Test unregistration
	err = modulator.UnregisterTarget(target1)
	if err != nil {
		t.Fatalf("Failed to unregister target1: %v", err)
	}

	// Check that target was removed
	glutamateTargetsAfter := modulator.bindingTargets[LigandGlutamate]
	if len(glutamateTargetsAfter) != 0 {
		t.Errorf("Expected 0 glutamate targets after unregistration, got %d", len(glutamateTargetsAfter))
	}

	t.Log("✓ Binding target system working correctly")
}

// =================================================================================
// TEST 6: BACKGROUND PROCESSOR AND TEMPORAL DYNAMICS
// =================================================================================

func TestBackgroundProcessorAndDecay(t *testing.T) {
	t.Log("=== BACKGROUND PROCESSOR AND DECAY TEST ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Register source
	astrocyteNetwork.Register(ComponentInfo{
		ID: "decay_source", Type: ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0}, State: StateActive, RegisteredAt: time.Now(),
	})

	// Start background processor
	err := modulator.Start()
	if err != nil {
		t.Fatalf("Failed to start modulator: %v", err)
	}
	defer modulator.Stop()

	// Release chemical
	err = modulator.Release(LigandGlutamate, "decay_source", 1.0)
	if err != nil {
		t.Fatalf("Failed to release glutamate: %v", err)
	}

	testPos := Position3D{X: 0, Y: 0, Z: 0}

	// Measure concentration immediately
	time.Sleep(1 * time.Millisecond) // Brief pause for initial processing
	initialConc := modulator.GetConcentration(LigandGlutamate, testPos)
	t.Logf("Initial concentration: %.6f", initialConc)

	// Wait for decay
	time.Sleep(50 * time.Millisecond) // Allow significant decay
	decayedConc := modulator.GetConcentration(LigandGlutamate, testPos)
	t.Logf("Concentration after 50ms: %.6f", decayedConc)

	// Validate decay occurred
	if decayedConc >= initialConc {
		t.Errorf("Concentration should decay over time (%.6f >= %.6f)", decayedConc, initialConc)
	}

	// Calculate decay rate
	if initialConc > 0 {
		decayRatio := decayedConc / initialConc
		t.Logf("Decay ratio after 50ms: %.3f (%.1f%% remaining)", decayRatio, decayRatio*100)

		// Validate that decay is reasonable (not too fast, not too slow)
		if decayRatio > 0.9 {
			t.Error("Decay too slow - concentration barely changed")
		}
		if decayRatio < 0.01 {
			t.Error("Decay too fast - concentration almost gone")
		}
	}

	// Test that concentration eventually goes to near zero
	time.Sleep(200 * time.Millisecond) // Wait longer
	finalConc := modulator.GetConcentration(LigandGlutamate, testPos)
	t.Logf("Concentration after 250ms total: %.6f", finalConc)

	if initialConc > 0 && finalConc > initialConc*0.1 {
		t.Errorf("Concentration should be mostly cleared after 250ms (%.6f > %.6f)",
			finalConc, initialConc*0.1)
	}

	t.Log("✓ Background processor and decay working correctly")
}

// =================================================================================
// TEST 7: SPATIAL CONCENTRATION GRADIENTS
// =================================================================================

func TestSpatialConcentrationGradients(t *testing.T) {
	t.Log("=== SPATIAL CONCENTRATION GRADIENTS TEST ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Register source at origin
	sourcePos := Position3D{X: 0, Y: 0, Z: 0}
	astrocyteNetwork.Register(ComponentInfo{
		ID: "gradient_source", Type: ComponentNeuron,
		Position: sourcePos, State: StateActive, RegisteredAt: time.Now(),
	})

	// Release chemical
	err := modulator.Release(LigandGlutamate, "gradient_source", 1.0)
	if err != nil {
		t.Fatalf("Failed to release glutamate: %v", err)
	}

	// Test concentration at various distances and directions
	testPositions := []struct {
		name             string
		pos              Position3D
		expectedDistance float64
	}{
		{"origin", Position3D{X: 0, Y: 0, Z: 0}, 0.0},
		{"1μm_x", Position3D{X: 1, Y: 0, Z: 0}, 1.0},
		{"1μm_y", Position3D{X: 0, Y: 1, Z: 0}, 1.0},
		{"1μm_z", Position3D{X: 0, Y: 0, Z: 1}, 1.0},
		{"1μm_diagonal", Position3D{X: 0.577, Y: 0.577, Z: 0.577}, 1.0}, // √3/3 for unit distance
		{"3μm_x", Position3D{X: 3, Y: 0, Z: 0}, 3.0},
		{"5μm_diagonal", Position3D{X: 3, Y: 4, Z: 0}, 5.0}, // 3-4-5 triangle
	}

	t.Log("Concentration gradient measurements:")
	var previousConc float64 = -1

	for i, test := range testPositions {
		conc := modulator.GetConcentration(LigandGlutamate, test.pos)
		actualDistance := modulator.calculateDistance(sourcePos, test.pos)

		t.Logf("%s (%.1fμm): concentration=%.6f", test.name, actualDistance, conc)

		// Validate distance calculation
		if abs(actualDistance-test.expectedDistance) > 0.1 {
			t.Errorf("Distance calculation error for %s: expected %.1f, got %.1f",
				test.name, test.expectedDistance, actualDistance)
		}

		// Validate gradient (concentration should generally decrease with distance)
		if i > 0 && test.expectedDistance > 0 {
			if previousConc > 0 && conc > previousConc && test.expectedDistance > testPositions[i-1].expectedDistance {
				t.Errorf("Concentration gradient violation: %s (%.6f) > previous (%.6f) at greater distance",
					test.name, conc, previousConc)
			}
		}

		if test.expectedDistance == 0 && conc <= 0 {
			t.Error("Should have non-zero concentration at source position")
		}

		previousConc = conc
	}

	// Test gradient calculation
	stepSize := 0.1
	gradX, gradY, gradZ := modulator.GetConcentrationGradient(LigandGlutamate,
		Position3D{X: 1, Y: 0, Z: 0}, stepSize)

	t.Logf("Concentration gradient at (1,0,0): x=%.6f, y=%.6f, z=%.6f", gradX, gradY, gradZ)

	// For a point source, gradient should point toward source (negative x direction)
	if gradX >= 0 {
		t.Error("Gradient X component should be negative (pointing toward source)")
	}

	// Y and Z gradients should be near zero for point on X axis
	if abs(gradY) > abs(gradX)*0.1 || abs(gradZ) > abs(gradX)*0.1 {
		t.Error("Gradient Y and Z components should be small compared to X")
	}

	t.Log("✓ Spatial concentration gradients working correctly")
}

// =================================================================================
// TEST 8: GLUTAMATE DECAY ISSUE - FOCUSED DEBUGGING
// =================================================================================

func TestGlutamateDecayIssue(t *testing.T) {
	t.Log("=== GLUTAMATE DECAY ISSUE TEST ===")
	t.Log("Debugging why glutamate concentration doesn't decrease over time")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Register source neuron
	sourcePos := Position3D{X: 0, Y: 0, Z: 0}
	astrocyteNetwork.Register(ComponentInfo{
		ID: "decay_test_neuron", Type: ComponentNeuron,
		Position: sourcePos, State: StateActive, RegisteredAt: time.Now(),
	})

	// Start background processor
	err := modulator.Start()
	if err != nil {
		t.Fatalf("Failed to start modulator: %v", err)
	}
	defer modulator.Stop()

	// Release glutamate
	err = modulator.Release(LigandGlutamate, "decay_test_neuron", 1.0)
	if err != nil {
		t.Fatalf("Failed to release glutamate: %v", err)
	}

	// DEBUG: Check concentration field details
	t.Log("DEBUG: Checking concentration field after release")
	if field, exists := modulator.concentrationFields[LigandGlutamate]; exists {
		t.Logf("Field has %d stored concentrations", len(field.Concentrations))
		t.Logf("Field last update: %v", field.LastUpdate)
		t.Logf("Field max concentration: %.6f", field.MaxConcentration)

		for pos, conc := range field.Concentrations {
			t.Logf("Position (%.1f,%.1f,%.1f): concentration %.6f", pos.X, pos.Y, pos.Z, conc)
		}
	} else {
		t.Error("No concentration field found for glutamate")
	}

	// Test concentration immediately
	initialConc := modulator.GetConcentration(LigandGlutamate, sourcePos)
	t.Logf("Initial concentration: %.6f", initialConc)

	// Wait for background decay and check multiple times
	intervals := []time.Duration{1 * time.Millisecond, 5 * time.Millisecond, 10 * time.Millisecond, 50 * time.Millisecond}

	for _, interval := range intervals {
		time.Sleep(interval)
		currentConc := modulator.GetConcentration(LigandGlutamate, sourcePos)

		// DEBUG: Check field state
		if field, exists := modulator.concentrationFields[LigandGlutamate]; exists {
			t.Logf("After %v: GetConcentration=%.6f, field_last_update=%v",
				interval, currentConc, field.LastUpdate)

			// Check stored concentration vs calculated
			if storedConc, exists := field.Concentrations[sourcePos]; exists {
				t.Logf("  Stored concentration: %.6f", storedConc)
			} else {
				t.Log("  No stored concentration at source position")
			}
		}
	}

	// Final concentration after significant time
	time.Sleep(100 * time.Millisecond)
	finalConc := modulator.GetConcentration(LigandGlutamate, sourcePos)
	t.Logf("Final concentration after 100ms: %.6f", finalConc)

	// Calculate expected decay
	glutamateKinetics := modulator.ligandKinetics[LigandGlutamate]
	totalDecayRate := glutamateKinetics.DecayRate + glutamateKinetics.ClearanceRate
	expectedConc := initialConc * math.Exp(-totalDecayRate*0.1) // 100ms = 0.1s
	t.Logf("Expected concentration after 100ms: %.6f (decay rate: %.3f)", expectedConc, totalDecayRate)

	// Test if background processor is actually running
	if finalConc == initialConc {
		t.Error("ISSUE IDENTIFIED: Concentration unchanged - background processor not updating stored values")
	} else if finalConc > expectedConc*2 {
		t.Error("ISSUE IDENTIFIED: Decay too slow - background processor running but not effective")
	} else {
		t.Log("✓ Decay working correctly")
	}
}

// =================================================================================
// TEST 9: DOPAMINE RANGE ISSUE - FOCUSED DEBUGGING
// =================================================================================

func TestDopamineRangeIssue(t *testing.T) {
	t.Log("=== DOPAMINE RANGE ISSUE TEST ===")
	t.Log("Debugging why dopamine shows 0.0 concentration at all distances")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Register source neuron
	sourcePos := Position3D{X: 0, Y: 0, Z: 0}
	astrocyteNetwork.Register(ComponentInfo{
		ID: "dopamine_source", Type: ComponentNeuron,
		Position: sourcePos, State: StateActive, RegisteredAt: time.Now(),
	})

	// Release dopamine
	err := modulator.Release(LigandDopamine, "dopamine_source", 1.0)
	if err != nil {
		t.Fatalf("Failed to release dopamine: %v", err)
	}

	// Check dopamine kinetics
	dopamineKinetics := modulator.ligandKinetics[LigandDopamine]
	t.Logf("Dopamine kinetics: diffusion=%.3f, decay=%.3f, clearance=%.3f, range=%.1f",
		dopamineKinetics.DiffusionRate, dopamineKinetics.DecayRate,
		dopamineKinetics.ClearanceRate, dopamineKinetics.MaxRange)

	// Test concentration calculation directly
	testDistances := []float64{0.0, 1.0, 5.0, 10.0, 20.0, 30.0, 50.0}

	t.Log("Testing distance-based concentration calculation:")
	for _, distance := range testDistances {
		calcConc := modulator.calculateConcentrationAtDistance(LigandDopamine, 1.0, distance)
		t.Logf("Distance %.1fμm: calculated concentration %.6f", distance, calcConc)

		// Check if distance exceeds max range
		if distance > dopamineKinetics.MaxRange && calcConc > 0 {
			t.Errorf("Concentration should be 0 beyond max range (%.1f > %.1f)",
				distance, dopamineKinetics.MaxRange)
		}

		// Check if calculation is reasonable
		if distance == 0.0 && calcConc != 1.0 {
			t.Errorf("At distance 0, concentration should equal source (%.6f != 1.0)", calcConc)
		}
	}

	// Test GetConcentration at various positions
	t.Log("Testing GetConcentration at various positions:")
	testPositions := []Position3D{
		{X: 0, Y: 0, Z: 0},  // Source
		{X: 1, Y: 0, Z: 0},  // 1μm away
		{X: 10, Y: 0, Z: 0}, // 10μm away
		{X: 20, Y: 0, Z: 0}, // 20μm away
	}

	for _, pos := range testPositions {
		conc := modulator.GetConcentration(LigandDopamine, pos)
		distance := modulator.calculateDistance(sourcePos, pos)
		t.Logf("Position (%.0f,%.0f,%.0f) [%.1fμm]: concentration %.6f",
			pos.X, pos.Y, pos.Z, distance, conc)
	}

	// Check if concentration field exists and has data
	if field, exists := modulator.concentrationFields[LigandDopamine]; exists {
		t.Logf("Dopamine field: %d stored concentrations, max=%.6f",
			len(field.Concentrations), field.MaxConcentration)

		if len(field.Concentrations) == 0 {
			t.Error("ISSUE IDENTIFIED: No concentrations stored in dopamine field")
		}
	} else {
		t.Error("ISSUE IDENTIFIED: No concentration field created for dopamine")
	}

	// Test if issue is in the calculation algorithm
	sourceConc := modulator.GetConcentration(LigandDopamine, sourcePos)
	nearbyConc := modulator.GetConcentration(LigandDopamine, Position3D{X: 1, Y: 0, Z: 0})

	if sourceConc == 0.0 {
		t.Error("ISSUE IDENTIFIED: No concentration at source position")
	} else if nearbyConc == 0.0 {
		t.Error("ISSUE IDENTIFIED: GetConcentration not calculating distance-based values")
	} else {
		t.Log("✓ Dopamine concentration calculation working")
	}
}

// =================================================================================
// TEST 10: BACKGROUND PROCESSOR EFFECTIVENESS
// =================================================================================

func TestBackgroundProcessorEffectiveness(t *testing.T) {
	t.Log("=== BACKGROUND PROCESSOR EFFECTIVENESS TEST ===")
	t.Log("Testing if background processor actually updates concentration fields")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Register source
	sourcePos := Position3D{X: 0, Y: 0, Z: 0}
	astrocyteNetwork.Register(ComponentInfo{
		ID: "bg_test_source", Type: ComponentNeuron,
		Position: sourcePos, State: StateActive, RegisteredAt: time.Now(),
	})

	// Test WITHOUT background processor first
	t.Log("Testing without background processor:")
	err := modulator.Release(LigandGlutamate, "bg_test_source", 1.0)
	if err != nil {
		t.Fatalf("Failed to release glutamate: %v", err)
	}

	conc1 := modulator.GetConcentration(LigandGlutamate, sourcePos)
	time.Sleep(50 * time.Millisecond)
	conc2 := modulator.GetConcentration(LigandGlutamate, sourcePos)

	t.Logf("Without background processor: %.6f → %.6f", conc1, conc2)

	// Now start background processor
	t.Log("Starting background processor:")
	err = modulator.Start()
	if err != nil {
		t.Fatalf("Failed to start background processor: %v", err)
	}
	defer modulator.Stop()

	// Release new chemical after starting processor
	err = modulator.Release(LigandGlutamate, "bg_test_source", 1.0)
	if err != nil {
		t.Fatalf("Failed to release glutamate with processor: %v", err)
	}

	conc3 := modulator.GetConcentration(LigandGlutamate, sourcePos)
	time.Sleep(50 * time.Millisecond)
	conc4 := modulator.GetConcentration(LigandGlutamate, sourcePos)

	t.Logf("With background processor: %.6f → %.6f", conc3, conc4)

	// Test if processor made a difference
	if conc2 == conc1 && conc4 == conc3 {
		t.Error("ISSUE: Background processor has no effect on concentrations")
	} else if conc4 >= conc3 {
		t.Error("ISSUE: Background processor not causing decay")
	} else {
		decayRate := (conc3 - conc4) / conc3
		t.Logf("✓ Background processor working: %.1f%% decay in 50ms", decayRate*100)
	}

	// Check if the processor updates stored values vs just calculation
	if field, exists := modulator.concentrationFields[LigandGlutamate]; exists {
		storedConc, hasStored := field.Concentrations[sourcePos]
		calculatedConc := modulator.GetConcentration(LigandGlutamate, sourcePos)

		t.Logf("Stored concentration: %.6f, Calculated: %.6f", storedConc, calculatedConc)

		if hasStored && math.Abs(storedConc-calculatedConc) > 0.001 {
			t.Error("ISSUE: Stored vs calculated concentration mismatch")
		}
	}
}

// =================================================================================
// TEST 11: CHEMICAL PARAMETERS VALIDATION
// =================================================================================

func TestChemicalParametersValidation(t *testing.T) {
	t.Log("=== CHEMICAL PARAMETERS VALIDATION TEST ===")
	t.Log("Validating that kinetic parameters produce biologically realistic behavior")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Test all ligand types
	ligandTypes := []LigandType{LigandGlutamate, LigandGABA, LigandDopamine, LigandSerotonin, LigandAcetylcholine}

	for _, ligandType := range ligandTypes {
		kinetics := modulator.ligandKinetics[ligandType]
		t.Logf("\n%v kinetics:", ligandType)
		t.Logf("  Diffusion: %.3f, Decay: %.3f, Clearance: %.3f",
			kinetics.DiffusionRate, kinetics.DecayRate, kinetics.ClearanceRate)
		t.Logf("  Range: %.1fμm, Affinity: %.3f", kinetics.MaxRange, kinetics.BindingAffinity)

		// Test range effectiveness at different distances
		sourceConc := 1.0
		distances := []float64{1.0, 5.0, 10.0, 20.0}

		for _, distance := range distances {
			conc := modulator.calculateConcentrationAtDistance(ligandType, sourceConc, distance)
			if distance <= kinetics.MaxRange && conc == 0.0 {
				t.Errorf("ISSUE: %v has zero concentration at %.1fμm (within range %.1fμm)",
					ligandType, distance, kinetics.MaxRange)
			}
		}

		// Test that fast neurotransmitters have shorter range than slow ones
		if ligandType == LigandGlutamate || ligandType == LigandGABA {
			if kinetics.MaxRange > 15.0 {
				t.Errorf("BIOLOGY ISSUE: Fast neurotransmitter %v has too long range (%.1f > 15.0)",
					ligandType, kinetics.MaxRange)
			}
		}

		if ligandType == LigandDopamine || ligandType == LigandSerotonin {
			if kinetics.MaxRange < 15.0 {
				t.Errorf("BIOLOGY ISSUE: Slow neuromodulator %v has too short range (%.1f < 15.0)",
					ligandType, kinetics.MaxRange)
			}
		}
	}

	// Validate relative parameters
	glutamate := modulator.ligandKinetics[LigandGlutamate]
	dopamine := modulator.ligandKinetics[LigandDopamine]

	if dopamine.DecayRate >= glutamate.DecayRate {
		t.Error("BIOLOGY ISSUE: Dopamine should decay slower than glutamate")
	}

	if dopamine.MaxRange <= glutamate.MaxRange {
		t.Error("BIOLOGY ISSUE: Dopamine should have longer range than glutamate")
	}

	t.Log("✓ Chemical parameters validation completed")
}

/*
=================================================================================
TEST FUNCTION FOR FIXED CHEMICAL KINETICS
=================================================================================

Add this test to your matrix_chemical_test.go file to verify the fixes work.
This test specifically validates that all neurotransmitters maintain reasonable
concentrations throughout their specified ranges.
=================================================================================
*/

func TestFixedKinetics(t *testing.T) {
	t.Log("=== TESTING FIXED CHEMICAL KINETICS ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// First, validate that all kinetics work properly
	issues := modulator.ValidateKinetics()

	if len(issues) == 0 {
		t.Log("✓ All neurotransmitter kinetics validated successfully")
	} else {
		t.Log("❌ Issues found:")
		for _, issue := range issues {
			t.Logf("  - %s", issue)
		}
	}

	// Test each neurotransmitter at various distances
	ligandTypes := []LigandType{LigandGlutamate, LigandGABA, LigandDopamine, LigandSerotonin, LigandAcetylcholine}
	ligandNames := []string{"Glutamate", "GABA", "Dopamine", "Serotonin", "Acetylcholine"}

	t.Log("\nDetailed concentration profiles:")

	allPassed := true

	for i, ligandType := range ligandTypes {
		kinetics := modulator.ligandKinetics[ligandType]
		t.Logf("\n%s (range: %.1fμm):", ligandNames[i], kinetics.MaxRange)

		// Test at key distances within the range
		distances := []float64{
			0.0,                      // Source
			kinetics.MaxRange * 0.1,  // 10% of range
			kinetics.MaxRange * 0.25, // Quarter range
			kinetics.MaxRange * 0.5,  // Half range
			kinetics.MaxRange * 0.75, // Three-quarters range
			kinetics.MaxRange * 0.9,  // Near max range
		}

		previousConc := 1.0
		for j, distance := range distances {
			conc := modulator.calculateConcentrationAtDistance(ligandType, 1.0, distance)
			percentage := int(distance / kinetics.MaxRange * 100)

			t.Logf("  %3d%% range (%.1fμm): %.6f", percentage, distance, conc)

			// Validation checks
			if j > 0 && conc > previousConc {
				t.Errorf("❌ %s: concentration increased with distance (%.6f > %.6f)", ligandNames[i], conc, previousConc)
				allPassed = false
			}

			// Key validation: should have detectable concentration within 90% of range
			// Use more lenient thresholds based on neurotransmitter type
			var minThreshold float64
			if ligandType == LigandGlutamate || ligandType == LigandGABA {
				minThreshold = 0.002 // More lenient for fast neurotransmitters
			} else if ligandType == LigandDopamine || ligandType == LigandSerotonin {
				minThreshold = 0.005 // Higher standard for neuromodulators
			} else {
				minThreshold = 0.003 // Acetylcholine intermediate
			}

			if distance <= kinetics.MaxRange*0.9 && conc < minThreshold {
				t.Errorf("❌ %s: concentration too low (%.6f < %.6f) at %.1fμm (within 90%% of range)", ligandNames[i], conc, minThreshold, distance)
				allPassed = false
			}

			// For neuromodulators, ensure they maintain reasonable concentrations for volume transmission
			if ligandType == LigandDopamine || ligandType == LigandSerotonin {
				if distance <= kinetics.MaxRange*0.5 && conc < 0.02 {
					t.Errorf("❌ %s (neuromodulator): concentration too low (%.6f) at half range", ligandNames[i], conc)
					allPassed = false
				}
			}

			previousConc = conc
		}
	}

	// Test biological realism constraints
	t.Log("\nTesting biological realism:")

	// Fast neurotransmitters should have shorter ranges
	glutamateRange := modulator.ligandKinetics[LigandGlutamate].MaxRange
	gabaRange := modulator.ligandKinetics[LigandGABA].MaxRange
	dopamineRange := modulator.ligandKinetics[LigandDopamine].MaxRange
	serotoninRange := modulator.ligandKinetics[LigandSerotonin].MaxRange

	if glutamateRange > dopamineRange || gabaRange > serotoninRange {
		t.Error("❌ Fast neurotransmitters should have shorter range than neuromodulators")
		allPassed = false
	} else {
		t.Log("✓ Range hierarchy correct: fast neurotransmitters < neuromodulators")
	}

	// Test concentration differences between fast and slow
	fastConc := modulator.calculateConcentrationAtDistance(LigandGlutamate, 1.0, 10.0)
	slowConc := modulator.calculateConcentrationAtDistance(LigandDopamine, 1.0, 10.0)

	if slowConc <= fastConc {
		t.Errorf("❌ Neuromodulators should maintain higher concentration at distance (dopamine: %.6f vs glutamate: %.6f at 10μm)", slowConc, fastConc)
		allPassed = false
	} else {
		t.Logf("✓ Distance concentration correct: dopamine %.6f > glutamate %.6f at 10μm", slowConc, fastConc)
	}

	// Final validation
	if allPassed && len(issues) == 0 {
		t.Log("\n🎯 ALL KINETIC PARAMETERS FIXED AND VALIDATED!")
		t.Log("✓ All neurotransmitters maintain effective concentrations within their ranges")
		t.Log("✓ Biological realism constraints satisfied")
		t.Log("✓ Fast vs slow neurotransmitter distinctions preserved")
	} else {
		t.Error("\n❌ Kinetic parameters still need adjustment")
	}
}

// BONUS: Test actual release and measurement integration
func TestFixedKineticsIntegration(t *testing.T) {
	t.Log("=== TESTING FIXED KINETICS WITH RELEASE INTEGRATION ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Register a source neuron
	sourcePos := Position3D{X: 0, Y: 0, Z: 0}
	astrocyteNetwork.Register(ComponentInfo{
		ID: "test_source", Type: ComponentNeuron,
		Position: sourcePos, State: StateActive, RegisteredAt: time.Now(),
	})

	// Test each neurotransmitter with actual release
	ligandTypes := []LigandType{LigandGlutamate, LigandGABA, LigandDopamine, LigandSerotonin, LigandAcetylcholine}
	ligandNames := []string{"Glutamate", "GABA", "Dopamine", "Serotonin", "Acetylcholine"}

	for i, ligandType := range ligandTypes {
		t.Logf("\nTesting %s release and measurement:", ligandNames[i])

		// Release neurotransmitter
		err := modulator.Release(ligandType, "test_source", 1.0)
		if err != nil {
			t.Fatalf("Failed to release %s: %v", ligandNames[i], err)
		}

		// Test concentrations at various positions
		kinetics := modulator.ligandKinetics[ligandType]
		testPositions := []Position3D{
			{X: 0, Y: 0, Z: 0},                        // Source
			{X: kinetics.MaxRange * 0.25, Y: 0, Z: 0}, // Quarter range
			{X: kinetics.MaxRange * 0.5, Y: 0, Z: 0},  // Half range
			{X: kinetics.MaxRange * 0.75, Y: 0, Z: 0}, // Three-quarter range
		}

		for _, pos := range testPositions {
			conc := modulator.GetConcentration(ligandType, pos)
			distance := modulator.calculateDistance(sourcePos, pos)

			t.Logf("  Distance %.1fμm: concentration %.6f", distance, conc)

			// Validate that we get reasonable concentrations
			if pos.X == 0 && conc <= 0 {
				t.Errorf("❌ %s: No concentration at source position", ligandNames[i])
			}

			// More lenient threshold based on neurotransmitter type
			var minThreshold float64
			if ligandType == LigandGlutamate || ligandType == LigandGABA {
				minThreshold = 0.002 // Lenient for fast neurotransmitters
			} else {
				minThreshold = 0.005 // Higher for others
			}

			if distance < kinetics.MaxRange*0.9 && conc < minThreshold {
				t.Errorf("❌ %s: Concentration too low (%.6f < %.6f) at %.1fμm", ligandNames[i], conc, minThreshold, distance)
			}
		}
	}

	t.Log("\n✓ Release and measurement integration working correctly")
}

// =================================================================================
// UTILITY FUNCTIONS
// =================================================================================

// abs returns absolute value of float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
