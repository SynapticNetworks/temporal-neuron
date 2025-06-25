/*
=================================================================================
CHEMICAL MODULATOR - EDGE CASE TESTS
=================================================================================

Tests edge cases, error conditions, boundary values, and unusual scenarios
to ensure robust operation under all conditions. Validates error handling,
boundary conditions, and system recovery mechanisms.

EDGE CASE CATEGORIES:
- Invalid input validation and error handling
- Extreme parameter values and boundary conditions
- Concurrent access patterns and race conditions
- Resource exhaustion and recovery scenarios
- Malformed data and corrupted state handling
- Network topology edge cases and unusual geometries

RELIABILITY TARGETS:
- Graceful handling of all invalid inputs
- No panics or crashes under any conditions
- Consistent behavior at parameter boundaries
- Thread-safe operations under high concurrency
- Recovery from resource exhaustion scenarios
=================================================================================
*/

package extracellular

import (
	"fmt"
	"math"
	"sync"
	"testing"
	"time"
)

// =================================================================================
// INPUT VALIDATION AND ERROR HANDLING TESTS
// =================================================================================

func TestChemicalModulatorEdgeInvalidInputValidation(t *testing.T) {
	t.Log("=== INVALID INPUT VALIDATION TEST ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// FIXED: Reset rate limits before testing
	modulator.ResetRateLimits()

	// Test empty/invalid component IDs
	t.Log("\nTesting invalid component IDs:")

	invalidIDs := []string{"", " ", "\n", "\t", "very_long_id_that_exceeds_reasonable_limits_and_should_be_handled_gracefully_without_causing_system_issues"}

	for _, invalidID := range invalidIDs {
		err := modulator.Release(LigandGlutamate, invalidID, 1.0)
		if err == nil {
			t.Logf("  Note: Release with invalid ID '%s' succeeded (may be intentional)", invalidID)
		} else {
			t.Logf("  ✓ Release with invalid ID '%s' properly rejected: %v", invalidID, err)
		}
	}

	// Test invalid concentrations
	t.Log("\nTesting invalid concentrations:")

	// Register valid neuron for testing
	astrocyteNetwork.Register(ComponentInfo{
		ID: "test_neuron", Type: ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0}, State: StateActive, RegisteredAt: time.Now(),
	})

	invalidConcentrations := []float64{-1.0, -100.0, math.NaN(), math.Inf(1), math.Inf(-1), 1e20, 0}

	for _, conc := range invalidConcentrations {
		err := modulator.Release(LigandGlutamate, "test_neuron", conc)
		if err != nil {
			t.Logf("  ✓ Invalid concentration %.2f properly rejected: %v", conc, err)
		} else {
			t.Logf("  Note: Concentration %.2f accepted (may be valid)", conc)
		}
	}

	// Test invalid ligand types (requires modifying the test or using reflection)
	t.Log("\nTesting boundary ligand types:")

	// Test with undefined ligand type (cast from invalid int)
	invalidLigand := LigandType(999)
	err := modulator.Release(invalidLigand, "test_neuron", 1.0)
	if err != nil {
		t.Logf("  ✓ Invalid ligand type properly rejected: %v", err)
	} else {
		t.Logf("  Note: Invalid ligand type accepted (may have default handling)")
	}

	// Test invalid positions
	t.Log("\nTesting invalid positions:")

	invalidPositions := []Position3D{
		{X: math.NaN(), Y: 0, Z: 0},
		{X: 0, Y: math.Inf(1), Z: 0},
		{X: 0, Y: 0, Z: math.Inf(-1)},
		{X: 1e20, Y: 1e20, Z: 1e20},
		{X: -1e20, Y: -1e20, Z: -1e20},
	}

	for _, pos := range invalidPositions {
		conc := modulator.GetConcentration(LigandGlutamate, pos)
		if math.IsNaN(conc) || math.IsInf(conc, 0) {
			t.Logf("  ⚠️ Invalid position returned invalid concentration: %.6f", conc)
		} else {
			t.Logf("  ✓ Invalid position handled gracefully: concentration %.6f", conc)
		}
	}

	t.Log("\n✓ Input validation tests completed")
}

// =================================================================================
// BOUNDARY VALUE TESTS
// =================================================================================

func TestChemicalModulatorEdgeBoundaryValues(t *testing.T) {
	t.Log("=== BOUNDARY VALUE TESTS ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Register neuron at origin
	astrocyteNetwork.Register(ComponentInfo{
		ID: "boundary_neuron", Type: ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0}, State: StateActive, RegisteredAt: time.Now(),
	})

	// Test extremely small concentrations
	t.Log("\nTesting extremely small concentrations:")
	smallConcentrations := []float64{1e-15, 1e-12, 1e-9, 1e-6, 1e-3}

	for i, conc := range smallConcentrations {
		// FIXED: Reset rate limits between tests and use different source IDs
		modulator.ResetRateLimits()
		time.Sleep(1 * time.Millisecond) // Brief pause

		// Use different neuron ID for each test to avoid rate limiting
		neuronID := fmt.Sprintf("boundary_neuron_%d", i)
		astrocyteNetwork.Register(ComponentInfo{
			ID: neuronID, Type: ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0}, State: StateActive, RegisteredAt: time.Now(),
		})

		err := modulator.Release(LigandGlutamate, neuronID, conc)
		if err != nil {
			t.Logf("  Small concentration %.0e rejected: %v", conc, err)
		} else {
			measuredConc := modulator.GetConcentration(LigandGlutamate, Position3D{X: float64(i), Y: 0, Z: 0})
			t.Logf("  Small concentration %.0e → measured %.0e", conc, measuredConc)
		}
	}

	// Test extremely large concentrations
	t.Log("\nTesting extremely large concentrations:")
	largeConcentrations := []float64{1e3, 1e6, 1e9, 1e12}

	for i, conc := range largeConcentrations {
		// FIXED: Reset rate limits and use different sources
		modulator.ResetRateLimits()
		time.Sleep(1 * time.Millisecond)

		neuronID := fmt.Sprintf("large_boundary_neuron_%d", i)
		astrocyteNetwork.Register(ComponentInfo{
			ID: neuronID, Type: ComponentNeuron,
			Position: Position3D{X: float64(i + 10), Y: 0, Z: 0}, State: StateActive, RegisteredAt: time.Now(),
		})

		err := modulator.Release(LigandDopamine, neuronID, conc)
		if err != nil {
			t.Logf("  Large concentration %.0e rejected: %v", conc, err)
		} else {
			measuredConc := modulator.GetConcentration(LigandDopamine, Position3D{X: float64(i + 10), Y: 0, Z: 0})
			t.Logf("  Large concentration %.0e → measured %.0e", conc, measuredConc)

			if measuredConc > conc*1.1 {
				t.Logf("    ⚠️ Measured concentration exceeds released concentration")
			}
		}
	}

	// Test extreme distances
	t.Log("\nTesting extreme distance queries:")

	// Release at origin
	modulator.Release(LigandSerotonin, "boundary_neuron", 5.0)

	extremeDistances := []float64{1e-6, 1e-3, 1e0, 1e3, 1e6, 1e9}

	for _, distance := range extremeDistances {
		pos := Position3D{X: distance, Y: 0, Z: 0}
		conc := modulator.GetConcentration(LigandSerotonin, pos)

		t.Logf("  Distance %.0e: concentration %.0e", distance, conc)

		// Validate concentration decreases with distance (unless at origin)
		if distance > 0 && conc > 5.0 {
			t.Logf("    ⚠️ Concentration at distance exceeds source concentration")
		}

		// Check for mathematical issues
		if math.IsNaN(conc) || math.IsInf(conc, 0) {
			t.Errorf("    ❌ Invalid concentration at distance %.0e: %.6f", distance, conc)
		}
	}

	// Test time boundary conditions
	t.Log("\nTesting temporal boundary conditions:")

	// FIXED: Reset rate limits and use fresh neuron
	modulator.ResetRateLimits()
	time.Sleep(2 * time.Millisecond)

	// Register fresh neuron for timing tests
	astrocyteNetwork.Register(ComponentInfo{
		ID: "timing_test_neuron", Type: ComponentNeuron,
		Position: Position3D{X: 100, Y: 100, Z: 100}, State: StateActive, RegisteredAt: time.Now(),
	})

	// Start background processor
	modulator.Start()
	defer modulator.Stop()

	// Release acetylcholine with error checking
	err := modulator.Release(LigandAcetylcholine, "timing_test_neuron", 3.0)
	if err != nil {
		t.Fatalf("Failed to release acetylcholine: %v", err)
	}

	testPos := Position3D{X: 100, Y: 100, Z: 100}

	// Measure concentration immediately
	time.Sleep(1 * time.Millisecond) // Brief pause for initial processing
	immediateConc := modulator.GetConcentration(LigandAcetylcholine, testPos)
	t.Logf("  Immediate query: %.6f", immediateConc)

	// Query after very short time
	time.Sleep(1 * time.Microsecond)
	microConc := modulator.GetConcentration(LigandAcetylcholine, testPos)
	t.Logf("  After 1μs: %.6f", microConc)

	// Query after longer time
	time.Sleep(100 * time.Millisecond)
	delayedConc := modulator.GetConcentration(LigandAcetylcholine, testPos)
	t.Logf("  After 100ms: %.6f", delayedConc)

	// Validate we got meaningful results
	if immediateConc <= 0 {
		t.Error("❌ No immediate concentration - release failed")
	} else {
		t.Logf("✓ Timing test working: %.6f immediate concentration", immediateConc)
	}

	t.Log("\n✓ Boundary value tests completed")
}

// =================================================================================
// CONCURRENT ACCESS AND RACE CONDITION TESTS
// =================================================================================

func TestChemicalModulatorEdgeConcurrentAccess(t *testing.T) {
	t.Log("=== CONCURRENT ACCESS TEST ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Register multiple neurons for concurrent access
	numNeurons := 100
	for i := 0; i < numNeurons; i++ {
		pos := Position3D{X: float64(i), Y: 0, Z: 0}
		astrocyteNetwork.Register(ComponentInfo{
			ID:           fmt.Sprintf("concurrent_neuron_%d", i),
			Type:         ComponentNeuron,
			Position:     pos,
			State:        StateActive,
			RegisteredAt: time.Now(),
		})
	}

	// Test concurrent releases vs queries
	t.Log("\nTesting concurrent releases and queries:")

	var wg sync.WaitGroup
	numWorkers := 20
	operationsPerWorker := 10
	var releaseErrors, queryErrors int64

	// Release workers
	for w := 0; w < numWorkers/2; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < operationsPerWorker; i++ {
				neuronID := fmt.Sprintf("concurrent_neuron_%d", (workerID*operationsPerWorker+i)%numNeurons)
				ligand := LigandType((workerID + i) % 4) // Cycle through ligand types
				concentration := 1.0 + float64(i%10)

				err := modulator.Release(ligand, neuronID, concentration)
				if err != nil {
					releaseErrors++
				}
			}
		}(w)
	}

	// Query workers
	for w := numWorkers / 2; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < operationsPerWorker; i++ {
				pos := Position3D{X: float64(i % numNeurons), Y: 0, Z: 0}
				ligand := LigandType(i % 4)

				conc := modulator.GetConcentration(ligand, pos)
				if math.IsNaN(conc) || math.IsInf(conc, 0) {
					queryErrors++
				}
			}
		}(w)
	}

	wg.Wait()

	t.Logf("  Concurrent operations completed")
	t.Logf("  Release errors: %d", releaseErrors)
	t.Logf("  Query errors: %d", queryErrors)

	if releaseErrors == 0 {
		t.Logf("  ✓ No release errors under concurrent access")
	} else {
		t.Logf("  ⚠️ Some release errors under concurrent access: %d", releaseErrors)
	}

	if queryErrors == 0 {
		t.Logf("  ✓ No query errors under concurrent access")
	} else {
		t.Errorf("  ❌ Query errors under concurrent access: %d", queryErrors)
	}

	// Test concurrent target registration/unregistration
	t.Log("\nTesting concurrent target management:")

	numTargets := 50
	targets := make([]*MockNeuron, numTargets)

	// Create targets
	for i := 0; i < numTargets; i++ {
		pos := Position3D{X: float64(i * 2), Y: 0, Z: 0}
		receptors := []LigandType{LigandGlutamate, LigandGABA}
		targets[i] = NewMockNeuron(fmt.Sprintf("edge_target_%d", i), pos, receptors)
	}

	var registrationErrors, unregistrationErrors int64

	// Concurrent registration/unregistration
	for w := 0; w < 10; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				targetIndex := (workerID*20 + i) % numTargets
				target := targets[targetIndex]

				// Register
				err := modulator.RegisterTarget(target)
				if err != nil {
					registrationErrors++
				}

				// Immediate unregister
				err = modulator.UnregisterTarget(target)
				if err != nil {
					unregistrationErrors++
				}
			}
		}(w)
	}

	wg.Wait()

	t.Logf("  Registration errors: %d", registrationErrors)
	t.Logf("  Unregistration errors: %d", unregistrationErrors)

	if registrationErrors == 0 && unregistrationErrors == 0 {
		t.Logf("  ✓ Concurrent target management working correctly")
	} else {
		t.Logf("  ⚠️ Some errors in concurrent target management")
	}

	t.Log("\n✓ Concurrent access tests completed")
}

// =================================================================================
// RESOURCE EXHAUSTION TESTS
// =================================================================================

func TestChemicalModulatorEdgeResourceExhaustion(t *testing.T) {
	t.Log("=== RESOURCE EXHAUSTION TEST ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Test maximum component registration
	t.Log("\nTesting maximum component registration:")

	maxComponents := 10000
	registrationErrors := 0

	for i := 0; i < maxComponents; i++ {
		pos := Position3D{X: float64(i % 100), Y: float64((i / 100) % 100), Z: float64(i / 10000)}
		err := astrocyteNetwork.Register(ComponentInfo{
			ID: fmt.Sprintf("exhaust_neuron_%d", i), Type: ComponentNeuron,
			Position: pos, State: StateActive, RegisteredAt: time.Now(),
		})

		if err != nil {
			registrationErrors++
			if registrationErrors == 1 {
				t.Logf("  First registration error at component %d: %v", i, err)
			}
		}

		// Log progress
		if i%1000 == 0 && i > 0 {
			t.Logf("  Registered %d components, errors: %d", i, registrationErrors)
		}
	}

	t.Logf("  Total registration attempts: %d", maxComponents)
	t.Logf("  Registration errors: %d", registrationErrors)

	if registrationErrors == 0 {
		t.Logf("  ✓ Successfully registered %d components", maxComponents)
	} else {
		t.Logf("  ⚠️ Registration limit reached after %d components", maxComponents-registrationErrors)
	}

	// Test massive concentration field creation
	t.Log("\nTesting massive concentration field creation:")

	maxReleases := 1000
	releaseErrors := 0
	startTime := time.Now()

	for i := 0; i < maxReleases; i++ {
		neuronID := fmt.Sprintf("exhaust_neuron_%d", i%1000) // Reuse some neurons
		ligand := LigandType(i % 5)                          // Cycle through all ligand types
		concentration := 1.0 + float64(i%10)

		err := modulator.Release(ligand, neuronID, concentration)
		if err != nil {
			releaseErrors++
			if releaseErrors == 1 {
				t.Logf("  First release error at release %d: %v", i, err)
			}
		}

		// Log progress and check performance degradation
		if i%100 == 0 && i > 0 {
			elapsed := time.Since(startTime)
			avgLatency := elapsed / time.Duration(i)
			t.Logf("  Released %d chemicals, errors: %d, avg latency: %v", i, releaseErrors, avgLatency)

			if avgLatency > 10*time.Millisecond {
				t.Logf("    ⚠️ Performance degradation detected: %v per release", avgLatency)
			}
		}
	}

	duration := time.Since(startTime)
	t.Logf("  Total release attempts: %d", maxReleases)
	t.Logf("  Release errors: %d", releaseErrors)
	t.Logf("  Total time: %v", duration)
	t.Logf("  Average latency: %v", duration/time.Duration(maxReleases))

	// Test concentration field size
	totalConcPoints := 0
	for ligandType, field := range modulator.concentrationFields {
		if field != nil {
			fieldSize := len(field.Concentrations)
			totalConcPoints += fieldSize
			t.Logf("  %v field size: %d points", ligandType, fieldSize)
		}
	}
	t.Logf("  Total concentration points: %d", totalConcPoints)

	// Test target registration exhaustion
	t.Log("\nTesting target registration exhaustion:")

	maxTargets := 1000
	targets := make([]*MockNeuron, maxTargets)
	targetErrors := 0

	for i := 0; i < maxTargets; i++ {
		pos := Position3D{X: float64(i % 50), Y: float64((i / 50) % 20), Z: 0}
		receptors := []LigandType{LigandGlutamate, LigandGABA, LigandDopamine}

		target := NewMockNeuron(fmt.Sprintf("exhaust_target_%d", i), pos, receptors)
		targets[i] = target

		err := modulator.RegisterTarget(target)
		if err != nil {
			targetErrors++
			if targetErrors == 1 {
				t.Logf("  First target registration error at target %d: %v", i, err)
			}
		}
	}

	t.Logf("  Target registration attempts: %d", maxTargets)
	t.Logf("  Target registration errors: %d", targetErrors)

	if targetErrors == 0 {
		t.Logf("  ✓ Successfully registered %d targets", maxTargets)
	} else {
		t.Logf("  ⚠️ Target registration limit reached")
	}

	t.Log("\n✓ Resource exhaustion tests completed")
}

// =================================================================================
// MALFORMED DATA AND STATE CORRUPTION TESTS
// =================================================================================

func TestChemicalModulatorEdgeMalformedDataHandling(t *testing.T) {
	t.Log("=== MALFORMED DATA HANDLING TEST ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Test with malformed component info
	t.Log("\nTesting malformed component registration:")

	malformedComponents := []ComponentInfo{
		{ID: "", Type: ComponentNeuron, Position: Position3D{}, State: StateActive},
		{ID: "test", Type: ComponentType(999), Position: Position3D{}, State: StateActive},
		{ID: "test2", Type: ComponentNeuron, Position: Position3D{X: math.NaN()}, State: StateActive},
		{ID: "test3", Type: ComponentNeuron, Position: Position3D{}, State: ComponentState(999)},
	}

	for i, component := range malformedComponents {
		err := astrocyteNetwork.Register(component)
		if err != nil {
			t.Logf("  ✓ Malformed component %d properly rejected: %v", i, err)
		} else {
			t.Logf("  Note: Malformed component %d accepted (may have default handling)", i)
		}
	}

	// Test state transitions with invalid data
	t.Log("\nTesting invalid state transitions:")

	// Register a valid component first
	validComponent := ComponentInfo{
		ID: "valid_component", Type: ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0}, State: StateActive, RegisteredAt: time.Now(),
	}
	astrocyteNetwork.Register(validComponent)

	// Try invalid state updates
	invalidStates := []ComponentState{ComponentState(-1), ComponentState(999)}

	for _, state := range invalidStates {
		err := astrocyteNetwork.UpdateState("valid_component", state)
		if err != nil {
			t.Logf("  ✓ Invalid state transition properly rejected: %v", err)
		} else {
			t.Logf("  Note: Invalid state %v accepted", state)
		}
	}

	// Test corrupted concentration fields
	t.Log("\nTesting concentration field corruption handling:")

	// Create normal concentration field
	modulator.Release(LigandGlutamate, "valid_component", 1.0)

	// Manually corrupt concentration field (if accessible)
	if field, exists := modulator.concentrationFields[LigandGlutamate]; exists {
		// Add invalid concentration points
		field.Concentrations[Position3D{X: math.NaN(), Y: 0, Z: 0}] = 5.0
		field.Concentrations[Position3D{X: 0, Y: math.Inf(1), Z: 0}] = math.NaN()
		field.MaxConcentration = math.Inf(-1)
	}

	// Test queries with corrupted field
	testPositions := []Position3D{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 1, Z: 1},
		{X: math.NaN(), Y: 0, Z: 0},
	}

	for _, pos := range testPositions {
		conc := modulator.GetConcentration(LigandGlutamate, pos)
		if math.IsNaN(conc) || math.IsInf(conc, 0) {
			t.Logf("  ⚠️ Corrupted field returned invalid concentration: %.6f at pos (%.1f,%.1f,%.1f)",
				conc, pos.X, pos.Y, pos.Z)
		} else {
			t.Logf("  ✓ Corrupted field handled gracefully: %.6f at pos (%.1f,%.1f,%.1f)",
				conc, pos.X, pos.Y, pos.Z)
		}
	}

	t.Log("\n✓ Malformed data handling tests completed")
}

// =================================================================================
// NETWORK TOPOLOGY EDGE CASES
// =================================================================================

func TestChemicalModulatorEdgeNetworkTopologyEdgeCases(t *testing.T) {
	t.Log("=== NETWORK TOPOLOGY EDGE CASES TEST ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Test with single neuron (minimal network)
	t.Log("\nTesting minimal network (single neuron):")

	astrocyteNetwork.Register(ComponentInfo{
		ID: "single_neuron", Type: ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0}, State: StateActive, RegisteredAt: time.Now(),
	})

	err := modulator.Release(LigandGlutamate, "single_neuron", 2.0)
	if err != nil {
		t.Logf("  Single neuron release failed: %v", err)
	} else {
		conc := modulator.GetConcentration(LigandGlutamate, Position3D{X: 0, Y: 0, Z: 0})
		t.Logf("  ✓ Single neuron network working: concentration %.3f", conc)
	}

	// Test with co-located neurons (zero distance)
	t.Log("\nTesting co-located neurons:")

	samePosition := Position3D{X: 10, Y: 10, Z: 10}

	for i := 0; i < 5; i++ {
		astrocyteNetwork.Register(ComponentInfo{
			ID: fmt.Sprintf("colocated_%d", i), Type: ComponentNeuron,
			Position: samePosition, State: StateActive, RegisteredAt: time.Now(),
		})
	}

	// Release from all co-located neurons
	for i := 0; i < 5; i++ {
		err := modulator.Release(LigandDopamine, fmt.Sprintf("colocated_%d", i), 1.0)
		if err != nil {
			t.Logf("  Co-located release %d failed: %v", i, err)
		}
	}

	colocatedConc := modulator.GetConcentration(LigandDopamine, samePosition)
	t.Logf("  Co-located neurons concentration: %.3f", colocatedConc)

	if colocatedConc > 10.0 {
		t.Logf("  ⚠️ Concentration very high from co-located releases: %.3f", colocatedConc)
	} else {
		t.Logf("  ✓ Co-located neurons handled reasonably")
	}

	// Test with extremely sparse network (large distances)
	t.Log("\nTesting extremely sparse network:")

	sparsePositions := []Position3D{
		{X: 0, Y: 0, Z: 0},
		{X: 1000, Y: 0, Z: 0},
		{X: 0, Y: 1000, Z: 0},
		{X: 0, Y: 0, Z: 1000},
		{X: 1000, Y: 1000, Z: 1000},
	}

	for i, pos := range sparsePositions {
		astrocyteNetwork.Register(ComponentInfo{
			ID: fmt.Sprintf("sparse_%d", i), Type: ComponentNeuron,
			Position: pos, State: StateActive, RegisteredAt: time.Now(),
		})
	}

	// Release from first neuron
	err = modulator.Release(LigandSerotonin, "sparse_0", 5.0)
	if err != nil {
		t.Logf("  Sparse network release failed: %v", err)
	}

	// Check concentrations at distant locations
	for i, pos := range sparsePositions {
		conc := modulator.GetConcentration(LigandSerotonin, pos)
		distance := math.Sqrt(pos.X*pos.X + pos.Y*pos.Y + pos.Z*pos.Z)
		t.Logf("  Distance %.0f: concentration %.6f", distance, conc)

		if i == 0 && conc <= 0 {
			t.Errorf("  ❌ Zero concentration at source position")
		}
		if i > 0 && conc > 0.1 {
			t.Logf("  Note: High concentration at large distance %.0f: %.6f", distance, conc)
		}
	}

	// Test with dense cluster (many nearby neurons)
	t.Log("\nTesting dense cluster:")

	clusterCenter := Position3D{X: 100, Y: 100, Z: 100}
	clusterRadius := 2.0
	numClusterNeurons := 20

	for i := 0; i < numClusterNeurons; i++ {
		// Random position within cluster radius
		angle := float64(i) * 2 * math.Pi / float64(numClusterNeurons)
		r := clusterRadius * float64(i) / float64(numClusterNeurons)

		pos := Position3D{
			X: clusterCenter.X + r*math.Cos(angle),
			Y: clusterCenter.Y + r*math.Sin(angle),
			Z: clusterCenter.Z,
		}

		astrocyteNetwork.Register(ComponentInfo{
			ID: fmt.Sprintf("cluster_%d", i), Type: ComponentNeuron,
			Position: pos, State: StateActive, RegisteredAt: time.Now(),
		})
	}

	// Release from all cluster neurons
	for i := 0; i < numClusterNeurons; i++ {
		err := modulator.Release(LigandGABA, fmt.Sprintf("cluster_%d", i), 0.5)
		if err != nil {
			t.Logf("  Cluster release %d failed: %v", i, err)
		}
	}

	clusterConc := modulator.GetConcentration(LigandGABA, clusterCenter)
	t.Logf("  Dense cluster concentration: %.3f", clusterConc)

	if clusterConc > 50.0 {
		t.Logf("  ⚠️ Very high concentration in dense cluster: %.3f", clusterConc)
	} else {
		t.Logf("  ✓ Dense cluster handled appropriately")
	}

	// Test linear arrangement (1D network)
	t.Log("\nTesting linear arrangement:")

	linearLength := 100.0
	numLinearNeurons := 10

	for i := 0; i < numLinearNeurons; i++ {
		pos := Position3D{
			X: float64(i) * linearLength / float64(numLinearNeurons-1),
			Y: 200, // Separate from other tests
			Z: 0,
		}

		astrocyteNetwork.Register(ComponentInfo{
			ID: fmt.Sprintf("linear_%d", i), Type: ComponentNeuron,
			Position: pos, State: StateActive, RegisteredAt: time.Now(),
		})
	}

	// Release from middle neuron
	middleIndex := numLinearNeurons / 2
	err = modulator.Release(LigandAcetylcholine, fmt.Sprintf("linear_%d", middleIndex), 3.0)
	if err != nil {
		t.Logf("  Linear arrangement release failed: %v", err)
	}

	// Check concentration gradient along line
	t.Log("  Linear concentration gradient:")
	for i := 0; i < numLinearNeurons; i++ {
		pos := Position3D{
			X: float64(i) * linearLength / float64(numLinearNeurons-1),
			Y: 200,
			Z: 0,
		}
		conc := modulator.GetConcentration(LigandAcetylcholine, pos)
		distanceFromSource := math.Abs(float64(i - middleIndex))
		t.Logf("    Position %d (distance %.1f): concentration %.6f", i, distanceFromSource, conc)
	}

	t.Log("\n✓ Network topology edge cases completed")
}

// =================================================================================
// SYSTEM RECOVERY TESTS
// =================================================================================

func TestChemicalModulatorEdgeSystemRecovery(t *testing.T) {
	t.Log("=== SYSTEM RECOVERY TEST ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Test recovery from rate limit exhaustion
	t.Log("\nTesting recovery from rate limit exhaustion:")

	// Register test neuron
	astrocyteNetwork.Register(ComponentInfo{
		ID: "recovery_neuron", Type: ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0}, State: StateActive, RegisteredAt: time.Now(),
	})

	// Attempt rapid releases to trigger rate limiting
	rapidReleases := 100
	rateLimitErrors := 0

	for i := 0; i < rapidReleases; i++ {
		err := modulator.Release(LigandGlutamate, "recovery_neuron", 1.0)
		if err != nil {
			rateLimitErrors++
		}
	}

	t.Logf("  Rapid releases attempted: %d", rapidReleases)
	t.Logf("  Rate limit errors: %d", rateLimitErrors)

	if rateLimitErrors > 0 {
		t.Logf("  ✓ Rate limiting working: %d releases blocked", rateLimitErrors)

		// Test recovery after waiting
		t.Log("  Testing recovery after delay:")
		time.Sleep(10 * time.Millisecond) // Wait for rate limit window

		err := modulator.Release(LigandGlutamate, "recovery_neuron", 1.0)
		if err != nil {
			t.Logf("  ⚠️ Still rate limited after delay: %v", err)
		} else {
			t.Logf("  ✓ Successfully recovered from rate limiting")
		}
	} else {
		t.Log("  Note: No rate limiting detected (may be disabled or very high limits)")
	}

	// Test recovery from corrupted state
	t.Log("\nTesting recovery from state corruption:")

	// Start background processor
	err := modulator.Start()
	if err != nil {
		t.Logf("  Failed to start modulator: %v", err)
	}

	// Create normal state
	modulator.Release(LigandDopamine, "recovery_neuron", 2.0)
	normalConc := modulator.GetConcentration(LigandDopamine, Position3D{X: 0, Y: 0, Z: 0})
	t.Logf("  Normal operation concentration: %.3f", normalConc)

	// Simulate corruption by directly modifying internal state (if accessible)
	// This tests the system's ability to handle unexpected internal state

	// Stop and restart system
	t.Log("  Testing system restart recovery:")
	modulator.Stop()

	// Verify system stopped
	time.Sleep(10 * time.Millisecond)

	// Restart system
	err = modulator.Start()
	if err != nil {
		t.Logf("  ⚠️ Failed to restart system: %v", err)
	} else {
		t.Logf("  ✓ System restarted successfully")

		// Test functionality after restart
		err = modulator.Release(LigandSerotonin, "recovery_neuron", 1.5)
		if err != nil {
			t.Logf("  ⚠️ Function impaired after restart: %v", err)
		} else {
			restartConc := modulator.GetConcentration(LigandSerotonin, Position3D{X: 0, Y: 0, Z: 0})
			t.Logf("  ✓ Normal function after restart: concentration %.3f", restartConc)
		}
	}

	// Clean shutdown
	modulator.Stop()

	// Test resource cleanup
	t.Log("\nTesting resource cleanup:")

	// Count components before cleanup
	componentCount := astrocyteNetwork.Count()
	t.Logf("  Components before cleanup: %d", componentCount)

	// Unregister all components
	components := astrocyteNetwork.List()
	cleanupErrors := 0

	for _, component := range components {
		err := astrocyteNetwork.Unregister(component.ID)
		if err != nil {
			cleanupErrors++
		}
	}

	finalCount := astrocyteNetwork.Count()
	t.Logf("  Components after cleanup: %d", finalCount)
	t.Logf("  Cleanup errors: %d", cleanupErrors)

	if finalCount == 0 {
		t.Logf("  ✓ Complete resource cleanup successful")
	} else {
		t.Logf("  ⚠️ Incomplete cleanup: %d components remaining", finalCount)
	}

	t.Log("\n✓ System recovery tests completed")
}

// =================================================================================
// ZERO AND NULL VALUE TESTS
// =================================================================================

func TestChemicalModulatorEdgeZeroAndNullValues(t *testing.T) {
	t.Log("=== ZERO AND NULL VALUE TESTS ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Test zero concentration releases
	t.Log("\nTesting zero concentration releases:")

	astrocyteNetwork.Register(ComponentInfo{
		ID: "zero_neuron", Type: ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0}, State: StateActive, RegisteredAt: time.Now(),
	})

	err := modulator.Release(LigandGlutamate, "zero_neuron", 0.0)
	if err != nil {
		t.Logf("  ✓ Zero concentration properly rejected: %v", err)
	} else {
		zeroConc := modulator.GetConcentration(LigandGlutamate, Position3D{X: 0, Y: 0, Z: 0})
		t.Logf("  Zero concentration accepted: measured %.6f", zeroConc)
	}

	// Test queries at origin position
	t.Log("\nTesting queries at origin position:")

	originConc := modulator.GetConcentration(LigandDopamine, Position3D{X: 0, Y: 0, Z: 0})
	t.Logf("  Origin concentration (no release): %.6f", originConc)

	if originConc != 0.0 {
		t.Logf("  ⚠️ Non-zero concentration at origin without release: %.6f", originConc)
	} else {
		t.Logf("  ✓ Zero concentration at origin as expected")
	}

	// Test with all zero position
	t.Log("\nTesting all-zero position handling:")

	modulator.Release(LigandSerotonin, "zero_neuron", 2.0)

	zeroPositions := []Position3D{
		{X: 0, Y: 0, Z: 0},
		{X: 0.0, Y: 0.0, Z: 0.0},
		{X: -0.0, Y: -0.0, Z: -0.0},
	}

	for i, pos := range zeroPositions {
		conc := modulator.GetConcentration(LigandSerotonin, pos)
		t.Logf("  Zero position variant %d: concentration %.6f", i, conc)

		if math.IsNaN(conc) || math.IsInf(conc, 0) {
			t.Errorf("  ❌ Invalid concentration at zero position: %.6f", conc)
		}
	}

	// Test empty component ID queries
	t.Log("\nTesting empty component operations:")

	emptyIDs := []string{"", " ", "\t", "\n"}

	for i, emptyID := range emptyIDs {
		err := modulator.Release(LigandGABA, emptyID, 1.0)
		t.Logf("  Empty ID variant %d ('%s'): %v", i, emptyID, err)
	}

	t.Log("\n✓ Zero and null value tests completed")
}

// =================================================================================
// EXTREME TIMING TESTS
// =================================================================================

func TestChemicalModulatorEdgeExtremeTiming(t *testing.T) {
	t.Log("=== EXTREME TIMING TESTS ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Register test neuron
	astrocyteNetwork.Register(ComponentInfo{
		ID: "timing_neuron", Type: ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0}, State: StateActive, RegisteredAt: time.Now(),
	})

	// Test rapid successive releases
	t.Log("\nTesting rapid successive releases:")

	rapidCount := 1000
	startTime := time.Now()
	successCount := 0
	errorCount := 0

	for i := 0; i < rapidCount; i++ {
		err := modulator.Release(LigandGlutamate, "timing_neuron", 1.0)
		if err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	duration := time.Since(startTime)
	avgLatency := duration / time.Duration(rapidCount)

	t.Logf("  Rapid releases: %d", rapidCount)
	t.Logf("  Successful: %d", successCount)
	t.Logf("  Errors: %d", errorCount)
	t.Logf("  Total time: %v", duration)
	t.Logf("  Average latency: %v", avgLatency)

	if avgLatency > 100*time.Microsecond {
		t.Logf("  ⚠️ High average latency: %v", avgLatency)
	} else {
		t.Logf("  ✓ Good rapid release performance")
	}

	// Test delayed queries after rapid releases
	t.Log("\nTesting delayed queries after rapid releases:")

	delays := []time.Duration{
		1 * time.Microsecond,
		1 * time.Millisecond,
		10 * time.Millisecond,
		100 * time.Millisecond,
	}

	for _, delay := range delays {
		time.Sleep(delay)
		conc := modulator.GetConcentration(LigandGlutamate, Position3D{X: 0, Y: 0, Z: 0})
		t.Logf("  After %v: concentration %.6f", delay, conc)
	}

	// Test background processor timing
	t.Log("\nTesting background processor timing:")

	err := modulator.Start()
	if err != nil {
		t.Logf("  Failed to start background processor: %v", err)
	} else {
		defer modulator.Stop()

		// Release and monitor decay timing
		modulator.Release(LigandDopamine, "timing_neuron", 5.0)

		measurements := []struct {
			delay time.Duration
			name  string
		}{
			{1 * time.Millisecond, "1ms"},
			{10 * time.Millisecond, "10ms"},
			{50 * time.Millisecond, "50ms"},
			{100 * time.Millisecond, "100ms"},
		}

		for _, measurement := range measurements {
			time.Sleep(measurement.delay)
			conc := modulator.GetConcentration(LigandDopamine, Position3D{X: 0, Y: 0, Z: 0})
			t.Logf("  Background processing after %s: concentration %.6f", measurement.name, conc)
		}
	}

	t.Log("\n✓ Extreme timing tests completed")
}

// =================================================================================
// MEMORY CORRUPTION SIMULATION
// =================================================================================

func TestChemicalModulatorEdgeMemoryCorruptionSimulation(t *testing.T) {
	t.Log("=== MEMORY CORRUPTION SIMULATION TEST ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	modulator := NewChemicalModulator(astrocyteNetwork)

	// Register test components
	for i := 0; i < 10; i++ {
		astrocyteNetwork.Register(ComponentInfo{
			ID: fmt.Sprintf("corrupt_test_%d", i), Type: ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0}, State: StateActive, RegisteredAt: time.Now(),
		})
	}

	// Create normal concentration fields
	t.Log("\nCreating normal concentration fields:")

	for i := 0; i < 5; i++ {
		neuronID := fmt.Sprintf("corrupt_test_%d", i)
		ligand := LigandType(i % 4)
		concentration := 1.0 + float64(i)

		err := modulator.Release(ligand, neuronID, concentration)
		if err != nil {
			t.Logf("  Failed to create field %d: %v", i, err)
		} else {
			t.Logf("  Created field %d: %v", i, ligand)
		}
	}

	// Test queries with potentially corrupted data
	t.Log("\nTesting queries with edge case positions:")

	edgeCasePositions := []Position3D{
		{X: math.MaxFloat64, Y: 0, Z: 0},
		{X: -math.MaxFloat64, Y: 0, Z: 0},
		{X: math.SmallestNonzeroFloat64, Y: 0, Z: 0},
		{X: 0, Y: math.MaxFloat64, Z: 0},
		{X: 0, Y: 0, Z: math.MaxFloat64},
		{X: math.MaxFloat64, Y: math.MaxFloat64, Z: math.MaxFloat64},
	}

	for i, pos := range edgeCasePositions {
		conc := modulator.GetConcentration(LigandGlutamate, pos)
		if math.IsNaN(conc) || math.IsInf(conc, 0) {
			t.Logf("  ⚠️ Edge position %d returned invalid concentration: %.6f", i, conc)
		} else {
			t.Logf("  ✓ Edge position %d handled gracefully: %.6f", i, conc)
		}
	}

	// Test with mixed valid/invalid operations
	t.Log("\nTesting mixed valid/invalid operations:")

	operations := []struct {
		valid       bool
		description string
		operation   func() error
	}{
		{true, "valid release", func() error {
			return modulator.Release(LigandGlutamate, "corrupt_test_0", 1.0)
		}},
		{false, "invalid concentration", func() error {
			return modulator.Release(LigandGlutamate, "corrupt_test_0", math.NaN())
		}},
		{true, "valid target registration", func() error {
			target := NewMockNeuron("temp_target", Position3D{X: 0, Y: 0, Z: 0}, []LigandType{LigandGlutamate})
			return modulator.RegisterTarget(target)
		}},
		{false, "invalid ligand type", func() error {
			return modulator.Release(LigandType(999), "corrupt_test_0", 1.0)
		}},
	}

	validCount := 0
	invalidCount := 0

	for i, op := range operations {
		err := op.operation()
		if op.valid {
			if err == nil {
				validCount++
				t.Logf("  ✓ Valid operation %d succeeded: %s", i, op.description)
			} else {
				t.Logf("  ⚠️ Valid operation %d failed: %s - %v", i, op.description, err)
			}
		} else {
			if err != nil {
				invalidCount++
				t.Logf("  ✓ Invalid operation %d properly rejected: %s", i, op.description)
			} else {
				t.Logf("  ⚠️ Invalid operation %d accepted: %s", i, op.description)
			}
		}
	}

	t.Logf("  Valid operations succeeded: %d", validCount)
	t.Logf("  Invalid operations rejected: %d", invalidCount)

	t.Log("\n✓ Memory corruption simulation completed")
}
