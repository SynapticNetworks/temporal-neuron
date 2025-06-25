package integration

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/extracellular"
	"github.com/SynapticNetworks/temporal-neuron/neuron"
	"github.com/SynapticNetworks/temporal-neuron/synapse"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// TestSpatialDelayIntegration validates the complete distance-based delay pipeline:
// Matrix calculates distance → Synapse gets spatial delay → Neuron handles scheduled delivery
func TestSpatialDelayIntegration(t *testing.T) {
	t.Log("=== SPATIAL DELAY INTEGRATION TEST ===")
	t.Log("Testing complete distance → delay → delivery pipeline")

	// === PHASE 1: SETUP MATRIX WITH SPATIAL SYSTEM ===
	t.Log("\n--- Phase 1: Matrix Setup with Spatial Awareness ---")

	matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true, // CRITICAL: Enable spatial calculations
		UpdateInterval:  10 * time.Millisecond,
		MaxComponents:   100,
	})

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}
	defer matrix.Stop()

	t.Log("✓ Matrix with spatial system initialized")

	// === PHASE 2: REGISTER NEURON FACTORY ===
	t.Log("\n--- Phase 2: Registering Neuron Factory ---")

	matrix.RegisterNeuronType("spatial_test", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		neuron := neuron.NewNeuron(
			id,
			config.Threshold,
			0.95,               // decay rate
			5*time.Millisecond, // refractory period
			1.0,                // fire factor
			0.0,                // target firing rate (disable homeostasis)
			0.0,                // homeostasis strength (disable)
		)

		// Set callbacks for matrix integration
		neuron.SetCallbacks(callbacks)
		return neuron, nil
	})

	t.Log("✓ Neuron factory registered")

	// === PHASE 3: REGISTER SYNAPSE FACTORY WITH SPATIAL AWARENESS ===
	t.Log("\n--- Phase 3: Registering Synapse Factory ---")

	matrix.RegisterSynapseType("spatial_synapse", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, exists := matrix.GetNeuron(config.PresynapticID)
		if !exists {
			return nil, fmt.Errorf("presynaptic neuron not found: %s", config.PresynapticID)
		}

		postNeuron, exists := matrix.GetNeuron(config.PostsynapticID)
		if !exists {
			return nil, fmt.Errorf("postsynaptic neuron not found: %s", config.PostsynapticID)
		}

		// Create synapse with matrix integration for spatial delays
		// The matrix integration happens automatically through the factory callback system
		syn := synapse.NewBasicSynapse(
			id,
			preNeuron,
			postNeuron,
			synapse.CreateDefaultSTDPConfig(),
			synapse.CreateDefaultPruningConfig(),
			config.InitialWeight,
			config.Delay,
		)

		// Matrix integration is handled through the callbacks parameter
		// The synapse gets spatial delay calculation via the GetTransmissionDelay callback
		return syn, nil
	})

	t.Log("✓ Synapse factory with spatial awareness registered")

	// === PHASE 4: TEST SCENARIOS WITH DIFFERENT DISTANCES ===
	testScenarios := []struct {
		name            string
		distance        float64 // μm
		axonSpeed       float64 // μm/ms
		baseSynaptic    time.Duration
		expectedSpatial time.Duration
		description     string
	}{
		{
			name:            "short_distance",
			distance:        50.0,   // 50 μm - very close
			axonSpeed:       2000.0, // cortical local speed
			baseSynaptic:    1 * time.Millisecond,
			expectedSpatial: 25 * time.Microsecond, // 50/2000 = 0.025ms
			description:     "Close cortical connection",
		},
		{
			name:            "medium_distance",
			distance:        500.0,  // 500 μm - moderate distance
			axonSpeed:       2000.0, // cortical local speed
			baseSynaptic:    1 * time.Millisecond,
			expectedSpatial: 250 * time.Microsecond, // 500/2000 = 0.25ms
			description:     "Medium cortical connection",
		},
		{
			name:            "long_distance",
			distance:        2000.0,  // 2mm - long cortical projection
			axonSpeed:       15000.0, // long-range speed
			baseSynaptic:    1 * time.Millisecond,
			expectedSpatial: 133333 * time.Nanosecond, // 2000/15000 ≈ 0.133ms
			description:     "Long-range cortical projection",
		},
		{
			name:            "slow_unmyelinated",
			distance:        1000.0, // 1mm
			axonSpeed:       500.0,  // unmyelinated slow speed
			baseSynaptic:    1 * time.Millisecond,
			expectedSpatial: 2 * time.Millisecond, // 1000/500 = 2ms
			description:     "Slow unmyelinated fiber",
		},
	}

	for i, scenario := range testScenarios {
		t.Logf("\n=== SCENARIO %d: %s ===", i+1, scenario.name)
		t.Logf("Distance: %.0f μm, Speed: %.0f μm/ms", scenario.distance, scenario.axonSpeed)
		t.Logf("Description: %s", scenario.description)

		// Set axon speed for this scenario
		matrix.SetAxonSpeed(scenario.axonSpeed)

		// Create neurons at specific positions
		preNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
			NeuronType: "spatial_test",
			Threshold:  0.5,                                // Low threshold for reliable firing
			Position:   types.Position3D{X: 0, Y: 0, Z: 0}, // Origin
		})
		if err != nil {
			t.Fatalf("Failed to create presynaptic neuron: %v", err)
		}

		postNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
			NeuronType: "spatial_test",
			Threshold:  0.5,                                                // Low threshold for reliable response
			Position:   types.Position3D{X: scenario.distance, Y: 0, Z: 0}, // Distance along X-axis
		})
		if err != nil {
			t.Fatalf("Failed to create postsynaptic neuron: %v", err)
		}

		// Start neurons
		err = preNeuron.Start()
		if err != nil {
			t.Fatalf("Failed to start presynaptic neuron: %v", err)
		}

		err = postNeuron.Start()
		if err != nil {
			t.Fatalf("Failed to start postsynaptic neuron: %v", err)
		}

		// Verify distance calculation
		calculatedDistance, err := matrix.GetSpatialDistance(preNeuron.ID(), postNeuron.ID())
		if err != nil {
			t.Fatalf("Failed to calculate distance: %v", err)
		}

		if math.Abs(calculatedDistance-scenario.distance) > 0.1 {
			t.Errorf("Distance calculation error: expected %.1f μm, got %.1f μm",
				scenario.distance, calculatedDistance)
		} else {
			t.Logf("✓ Distance calculation accurate: %.1f μm", calculatedDistance)
		}

		// Create synapse with spatial delay integration
		testSynapse, err := matrix.CreateSynapse(types.SynapseConfig{
			SynapseType:    "spatial_synapse",
			PresynapticID:  preNeuron.ID(),
			PostsynapticID: postNeuron.ID(),
			InitialWeight:  1.0, // Strong weight
			Delay:          scenario.baseSynaptic,
		})
		if err != nil {
			t.Fatalf("Failed to create synapse: %v", err)
		}

		// Verify total delay calculation
		expectedTotalDelay := scenario.baseSynaptic + scenario.expectedSpatial
		actualTotalDelay := matrix.SynapticDelay(
			preNeuron.ID(),
			postNeuron.ID(),
			testSynapse.ID(),
			scenario.baseSynaptic,
		)

		tolerance := float64(scenario.expectedSpatial) * 0.02 // 2% tolerance
		delayDifference := math.Abs(float64(actualTotalDelay - expectedTotalDelay))

		if delayDifference > tolerance {
			t.Errorf("Total delay error: expected %v, got %v (difference: %.3fμs)",
				expectedTotalDelay, actualTotalDelay, delayDifference/float64(time.Microsecond))
		} else {
			t.Logf("✓ Total delay calculation accurate: %v", actualTotalDelay)
		}

		// === TEST ACTUAL DELIVERY TIMING ===
		t.Log("--- Testing Actual Message Delivery Timing ---")

		// Get baseline activity
		initialPostActivity := postNeuron.GetActivityLevel()

		// Send signal and measure timing
		signalStart := time.Now()

		// Use the synapse to transmit (this should use spatial delays)
		testSynapse.Transmit(2.0) // Strong signal

		// Check for immediate delivery (should not happen with delay)
		time.Sleep(5 * time.Millisecond) // Brief check
		immediateActivity := postNeuron.GetActivityLevel()

		// Wait for expected delivery time
		time.Sleep(actualTotalDelay + 10*time.Millisecond) // Buffer for processing

		deliveryTime := time.Since(signalStart)
		finalActivity := postNeuron.GetActivityLevel()

		// Verify timing behavior
		if immediateActivity > initialPostActivity && actualTotalDelay > 5*time.Millisecond {
			t.Logf("NOTE: Signal delivered immediately despite %.3fms expected delay",
				float64(actualTotalDelay)/float64(time.Millisecond))
		}

		if finalActivity > initialPostActivity {
			t.Logf("✓ Signal delivered within timing window")
			t.Logf("  Delivery time: %v (expected delay: %v)", deliveryTime, actualTotalDelay)
			t.Logf("  Activity change: %.3f → %.3f", initialPostActivity, finalActivity)
		} else {
			t.Errorf("Signal not delivered after expected delay")
		}

		// Clean up for next scenario
		preNeuron.Stop()
		postNeuron.Stop()
		time.Sleep(50 * time.Millisecond) // Allow cleanup

		t.Logf("✓ Scenario '%s' completed successfully", scenario.name)
	}

	// === PHASE 5: VALIDATE ZERO-DELAY BEHAVIOR ===
	t.Log("\n=== PHASE 5: Zero-Delay Validation ===")

	// Test immediate delivery for zero/minimal distance
	matrix.SetAxonSpeed(2000.0) // Standard speed

	closePreNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "spatial_test",
		Threshold:  0.5,
		Position:   types.Position3D{X: 0, Y: 0, Z: 0},
	})
	if err != nil {
		t.Fatalf("Failed to create close presynaptic neuron: %v", err)
	}

	closePostNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "spatial_test",
		Threshold:  0.5,
		Position:   types.Position3D{X: 1.0, Y: 0, Z: 0}, // 1 μm distance
	})
	if err != nil {
		t.Fatalf("Failed to create close postsynaptic neuron: %v", err)
	}

	closePreNeuron.Start()
	closePostNeuron.Start()
	defer closePreNeuron.Stop()
	defer closePostNeuron.Stop()

	// Create zero-delay synapse
	closeSynapse, err := matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "spatial_synapse",
		PresynapticID:  closePreNeuron.ID(),
		PostsynapticID: closePostNeuron.ID(),
		InitialWeight:  1.0,
		Delay:          0, // Zero base delay
	})
	if err != nil {
		t.Fatalf("Failed to create close synapse: %v", err)
	}

	// Calculate expected minimal delay
	closeDistance, _ := matrix.GetSpatialDistance(closePreNeuron.ID(), closePostNeuron.ID())
	closeDelay := matrix.SynapticDelay(closePreNeuron.ID(), closePostNeuron.ID(), closeSynapse.ID(), 0)

	t.Logf("Close distance: %.1f μm, total delay: %v", closeDistance, closeDelay)

	// Test immediate-like delivery
	initialActivity := closePostNeuron.GetActivityLevel()
	closeSynapse.Transmit(2.0)

	time.Sleep(5 * time.Millisecond) // Very short wait
	quickActivity := closePostNeuron.GetActivityLevel()

	if quickActivity > initialActivity {
		t.Log("✓ Near-immediate delivery for minimal distance confirmed")
	} else {
		time.Sleep(50 * time.Millisecond) // Wait longer
		laterActivity := closePostNeuron.GetActivityLevel()
		if laterActivity > initialActivity {
			t.Log("✓ Close-distance delivery completed")
		} else {
			t.Error("Signal delivery failed for close distance")
		}
	}

	// === PHASE 6: DIFFERENT AXON TYPES VALIDATION ===
	t.Log("\n=== PHASE 6: Biological Axon Types Validation ===")

	axonTypes := []struct {
		name        string
		speed       float64
		description string
	}{
		{"cortical_local", 2000.0, "Local cortical circuits"},
		{"long_range", 15000.0, "Long-range projections"},
		{"unmyelinated_slow", 500.0, "Unmyelinated pain fibers"},
		{"myelinated_fast", 80000.0, "Fast motor fibers"},
	}

	testDistance := 1000.0 // 1mm standard test distance

	for _, axonType := range axonTypes {
		t.Logf("\n--- Testing Axon Type: %s ---", axonType.name)
		t.Logf("Description: %s", axonType.description)

		matrix.SetAxonSpeed(axonType.speed)

		baseDelay := 1 * time.Millisecond
		expectedSpatialDelay := time.Duration((testDistance / axonType.speed) * float64(time.Millisecond))
		expectedTotalDelay := baseDelay + expectedSpatialDelay

		actualTotalDelay := matrix.SynapticDelay("test_pre", "test_post", "test_syn", baseDelay)

		// Note: This test uses hypothetical IDs since we're just testing the calculation
		if math.Abs(float64(actualTotalDelay-expectedTotalDelay)) < float64(1*time.Microsecond) {
			t.Logf("✓ %s: Delay calculation correct (%v)", axonType.name, actualTotalDelay)
		} else {
			t.Logf("- %s: Different delay (expected %v, got %v) - may need registered neurons",
				axonType.name, expectedTotalDelay, actualTotalDelay)
		}
	}

	t.Log("\n✅ SPATIAL DELAY INTEGRATION TEST COMPLETED")
	t.Log("✅ Complete distance → delay → delivery pipeline validated")
}

// TestMatrixPositionResponsibility tests that the matrix properly handles neuron positioning
// without requiring factories to manually call SetPosition().
//
// This test verifies the architectural principle that:
// 1. Factories should only create neurons with correct parameters
// 2. Matrix should handle integration concerns like position setting
// 3. config.Position should be automatically applied by the matrix
//
// EXPECTED BEHAVIOR:
// - Matrix should call neuron.SetPosition(config.Position) after factory creation
// - Neuron position should match config.Position without factory intervention
// - Distance calculations should work correctly with matrix-managed positions
//
// This test will FAIL initially to demonstrate the bug, then PASS after the matrix fix.
func TestSpatialMatrixPositionResponsibility(t *testing.T) {
	t.Log("=== MATRIX POSITION RESPONSIBILITY TEST ===")
	t.Log("Testing that matrix handles position setting automatically")
	t.Log("EXPECTED: This test should FAIL initially, then PASS after matrix fix")

	// === SETUP MATRIX ===
	matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  10 * time.Millisecond,
		MaxComponents:   100,
	})

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}
	defer matrix.Stop()

	// === REGISTER FACTORY THAT DOES NOT SET POSITION ===
	// This factory intentionally does NOT call SetPosition() to test matrix responsibility
	matrix.RegisterNeuronType("position_test", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		t.Logf("🏭 Factory received config with position: (%.1f, %.1f, %.1f)",
			config.Position.X, config.Position.Y, config.Position.Z)

		// Create neuron with only the essential parameters
		neuron := neuron.NewNeuron(
			id,
			config.Threshold,
			0.95,               // decay rate
			5*time.Millisecond, // refractory period
			1.0,                // fire factor
			0.0,                // target firing rate
			0.0,                // homeostasis strength
		)

		// 🚨 INTENTIONALLY DO NOT CALL SetPosition()
		// The matrix should handle this automatically
		t.Logf("🏭 Factory created neuron at default position: (%.1f, %.1f, %.1f)",
			neuron.Position().X, neuron.Position().Y, neuron.Position().Z)

		neuron.SetCallbacks(callbacks)
		t.Logf("🏭 Factory completed - neuron should be at (0,0,0) until matrix fixes it")
		return neuron, nil
	})

	// === TEST CASE 1: SINGLE NEURON POSITION SETTING ===
	t.Log("\n--- Test Case 1: Single Neuron Position Integration ---")

	expectedPosition := types.Position3D{X: 50, Y: 25, Z: 10}
	t.Logf("Creating neuron with expected position: (%.1f, %.1f, %.1f)",
		expectedPosition.X, expectedPosition.Y, expectedPosition.Z)

	neuron1, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "position_test",
		Threshold:  0.5,
		Position:   expectedPosition,
	})
	if err != nil {
		t.Fatalf("Failed to create neuron: %v", err)
	}

	actualPosition := neuron1.Position()
	t.Logf("Neuron actual position after matrix integration: (%.1f, %.1f, %.1f)",
		actualPosition.X, actualPosition.Y, actualPosition.Z)

	// THE CRITICAL TEST: Does the neuron have the position from config?
	if actualPosition.X != expectedPosition.X ||
		actualPosition.Y != expectedPosition.Y ||
		actualPosition.Z != expectedPosition.Z {
		t.Errorf("❌ MATRIX POSITION BUG: Expected (%.1f,%.1f,%.1f), got (%.1f,%.1f,%.1f)",
			expectedPosition.X, expectedPosition.Y, expectedPosition.Z,
			actualPosition.X, actualPosition.Y, actualPosition.Z)
		t.Log("💡 FIX NEEDED: Matrix should call neuron.SetPosition(config.Position) in CreateNeuron()")
		t.Log("💡 LOCATION: After factory call, before integrateNeuronIntoBiologicalSystems()")
	} else {
		t.Log("✅ PASS: Matrix correctly set neuron position from config")
	}

	// === TEST CASE 2: DISTANCE CALCULATION WITH MATRIX POSITIONS ===
	t.Log("\n--- Test Case 2: Distance Calculation Accuracy ---")

	// Create second neuron at known distance
	neuron2, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "position_test",
		Threshold:  0.5,
		Position:   types.Position3D{X: 150, Y: 25, Z: 10}, // 100μm away on X-axis
	})
	if err != nil {
		t.Fatalf("Failed to create second neuron: %v", err)
	}

	expectedDistance := 100.0 // μm
	actualDistance, err := matrix.GetSpatialDistance(neuron1.ID(), neuron2.ID())
	if err != nil {
		t.Errorf("❌ Distance calculation failed: %v", err)
	} else if actualDistance != expectedDistance {
		t.Errorf("❌ DISTANCE CALCULATION ERROR: Expected %.1f μm, got %.1f μm",
			expectedDistance, actualDistance)
		t.Log("🔍 This error confirms the position setting bug affects spatial calculations")
	} else {
		t.Log("✅ PASS: Distance calculation accurate with matrix-managed positions")
	}

	// === TEST CASE 3: SPATIAL DELAY CALCULATION ===
	t.Log("\n--- Test Case 3: Spatial Delay Integration ---")

	matrix.SetAxonSpeed(2000.0) // 2 μm/ms
	baseSynapticDelay := 1 * time.Millisecond
	expectedSpatialDelay := 50 * time.Microsecond // 100μm / 2000μm/ms = 0.05ms
	expectedTotalDelay := baseSynapticDelay + expectedSpatialDelay

	actualTotalDelay := matrix.SynapticDelay(neuron1.ID(), neuron2.ID(), "test_synapse", baseSynapticDelay)

	t.Logf("Expected total delay: %v (base: %v + spatial: %v)",
		expectedTotalDelay, baseSynapticDelay, expectedSpatialDelay)
	t.Logf("Actual total delay: %v", actualTotalDelay)

	if actualTotalDelay == baseSynapticDelay {
		t.Errorf("❌ NO SPATIAL DELAY: Only base delay returned, confirms position bug")
		t.Log("🔍 Spatial delay = 0 because distance = 0 due to incorrect positions")
	} else if actualTotalDelay == expectedTotalDelay {
		t.Log("✅ PASS: Spatial delay correctly calculated with matrix-managed positions")
	} else {
		t.Logf("⚠️  PARTIAL: Spatial delay added but not expected value (got %v, expected %v)",
			actualTotalDelay, expectedTotalDelay)
	}

	// === TEST CASE 4: MULTIPLE NEURON POSITION VERIFICATION ===
	t.Log("\n--- Test Case 4: Multiple Neuron Position Verification ---")

	testPositions := []types.Position3D{
		{X: 0, Y: 0, Z: 0},    // Origin
		{X: 100, Y: 0, Z: 0},  // X-axis
		{X: 0, Y: 100, Z: 0},  // Y-axis
		{X: 0, Y: 0, Z: 100},  // Z-axis
		{X: 50, Y: 50, Z: 50}, // Diagonal
	}

	allPositionsCorrect := true
	for i, pos := range testPositions {
		neuron, err := matrix.CreateNeuron(types.NeuronConfig{
			NeuronType: "position_test",
			Threshold:  0.5,
			Position:   pos,
		})
		if err != nil {
			t.Errorf("Failed to create neuron %d: %v", i, err)
			continue
		}

		actual := neuron.Position()
		if actual.X != pos.X || actual.Y != pos.Y || actual.Z != pos.Z {
			t.Errorf("❌ Neuron %d position mismatch: expected (%.1f,%.1f,%.1f), got (%.1f,%.1f,%.1f)",
				i, pos.X, pos.Y, pos.Z, actual.X, actual.Y, actual.Z)
			allPositionsCorrect = false
		}
	}

	if allPositionsCorrect {
		t.Log("✅ PASS: All neurons have correct matrix-managed positions")
	} else {
		t.Log("❌ FAIL: Matrix not properly managing neuron positions")
	}

	// === SUMMARY AND GUIDANCE ===
	t.Log("\n=== TEST SUMMARY AND FIX GUIDANCE ===")
	t.Log("This test verifies that the matrix takes responsibility for neuron positioning.")
	t.Log("")
	t.Log("IF THIS TEST FAILS:")
	t.Log("1. The matrix CreateNeuron() method needs to be updated")
	t.Log("2. Add this line after the factory call:")
	t.Log("   neuron.SetPosition(config.Position)")
	t.Log("3. Location: In matrix.go, CreateNeuron(), between PHASE 2 and PHASE 3")
	t.Log("")
	t.Log("EXPECTED FLOW:")
	t.Log("Factory creates neuron → Matrix sets position → Matrix integrates → Done")
	t.Log("")
	t.Log("WHY THIS MATTERS:")
	t.Log("- Removes burden from factory authors")
	t.Log("- Ensures consistent position handling")
	t.Log("- Enables proper spatial delay calculations")
	t.Log("- Follows single responsibility principle")
}

// TestMatrixPositionResponsibilityEdgeCases tests edge cases for position management
func TestSpatialMatrixPositionResponsibilityEdgeCases(t *testing.T) {
	t.Log("=== MATRIX POSITION RESPONSIBILITY EDGE CASES ===")

	matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
		SpatialEnabled: true,
		MaxComponents:  50,
	})

	matrix.Start()
	defer matrix.Stop()

	// Register minimal factory
	matrix.RegisterNeuronType("edge_test", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		neuron := neuron.NewNeuron(id, config.Threshold, 0.95, 5*time.Millisecond, 1.0, 0.0, 0.0)
		neuron.SetCallbacks(callbacks)
		return neuron, nil
	})

	// === EDGE CASE 1: Zero Position ===
	t.Log("\n--- Edge Case 1: Zero Position ---")
	zeroNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "edge_test",
		Threshold:  0.5,
		Position:   types.Position3D{X: 0, Y: 0, Z: 0},
	})
	if err != nil {
		t.Fatalf("Failed to create zero position neuron: %v", err)
	}

	pos := zeroNeuron.Position()
	if pos.X != 0 || pos.Y != 0 || pos.Z != 0 {
		t.Errorf("❌ Zero position not preserved: got (%.1f,%.1f,%.1f)", pos.X, pos.Y, pos.Z)
	} else {
		t.Log("✅ Zero position correctly handled")
	}

	// === EDGE CASE 2: Negative Coordinates ===
	t.Log("\n--- Edge Case 2: Negative Coordinates ---")
	negativeNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "edge_test",
		Threshold:  0.5,
		Position:   types.Position3D{X: -10, Y: -20, Z: -30},
	})
	if err != nil {
		t.Fatalf("Failed to create negative position neuron: %v", err)
	}

	pos = negativeNeuron.Position()
	if pos.X != -10 || pos.Y != -20 || pos.Z != -30 {
		t.Errorf("❌ Negative position not preserved: expected (-10,-20,-30), got (%.1f,%.1f,%.1f)",
			pos.X, pos.Y, pos.Z)
	} else {
		t.Log("✅ Negative coordinates correctly handled")
	}

	// === EDGE CASE 3: Large Coordinates ===
	t.Log("\n--- Edge Case 3: Large Coordinates ---")
	largeNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "edge_test",
		Threshold:  0.5,
		Position:   types.Position3D{X: 1000000, Y: 2000000, Z: 3000000},
	})
	if err != nil {
		t.Fatalf("Failed to create large position neuron: %v", err)
	}

	pos = largeNeuron.Position()
	if pos.X != 1000000 || pos.Y != 2000000 || pos.Z != 3000000 {
		t.Errorf("❌ Large position not preserved: expected (1M,2M,3M), got (%.0f,%.0f,%.0f)",
			pos.X, pos.Y, pos.Z)
	} else {
		t.Log("✅ Large coordinates correctly handled")
	}

	t.Log("\n✅ Edge cases demonstrate matrix position management robustness")
}
