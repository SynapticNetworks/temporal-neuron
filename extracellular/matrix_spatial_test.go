/*
=================================================================================
EXTRACELLULAR MATRIX - SPATIAL DELAY AND DISTANCE VALIDATION SUITE
=================================================================================

This test suite validates the spatial delay calculation system that combines
synaptic processing delays with distance-based axonal propagation delays.
These tests ensure that synaptic transmission timing reflects realistic
3D neural tissue properties and axonal conduction velocities.

SPATIAL SYSTEMS TESTED:
1. Basic Distance Calculation - 3D Euclidean distance between neural components
2. Axonal Conduction Delays - Speed-dependent propagation timing
3. Biological Axon Types - Realistic conduction velocity presets
4. Cortical Circuit Scenarios - Real-world neural pathway timing
5. Error Handling - Robust behavior with missing components

BIOLOGICAL BASIS:
- Axonal conduction velocities: 0.5-120 m/s depending on myelination
- Spatial delays: Distance/velocity calculations based on cable theory
- Synaptic delays: 0.5-2ms chemical transmission time
- Total transmission delay: Spatial + synaptic components

VALIDATION APPROACH:
- Tests use experimentally-measured conduction velocities
- Distance calculations validated against geometric expectations
- Delay computations checked against neurophysiology data
- Error conditions ensure system robustness

USAGE:
Run all spatial tests: go test -run TestMatrixSpatial
Run specific test: go test -run TestMatrixSpatialBasicCalculation

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
// SPATIAL DELAY CONSTANTS - Based on Neurophysiology Data
// =================================================================================

const (
	// Axonal Conduction Velocities (μm/ms)
	UNMYELINATED_SLOW_SPEED = 500.0   // 0.5 m/s - C fibers (pain, temperature)
	UNMYELINATED_FAST_SPEED = 2000.0  // 2 m/s - cortical local circuits
	MYELINATED_MEDIUM_SPEED = 10000.0 // 10 m/s - A-delta fibers (fast pain)
	MYELINATED_FAST_SPEED   = 80000.0 // 80 m/s - A-alpha fibers (motor, proprioception)

	// Cortical Circuit Speeds
	CORTICAL_LOCAL_SPEED = 2000.0  // Local cortical circuits (within columns)
	CORTICAL_INTER_SPEED = 5000.0  // Inter-laminar connections
	CORTICAL_LONG_SPEED  = 15000.0 // Long-range cortical projections

	// Default synaptic processing delay
	DEFAULT_SYNAPTIC_DELAY = 500 * time.Microsecond // 0.5ms
)

// =================================================================================
// TEST 1: BASIC SPATIAL DISTANCE AND DELAY CALCULATION
// =================================================================================

// TestMatrixSpatialBasicCalculation validates fundamental 3D distance measurement
// and spatial delay computation for neural components.
//
// BIOLOGICAL PROCESSES TESTED:
// - 3D Euclidean distance calculation between neurons
// - Axonal propagation delay: distance/velocity
// - Total transmission delay: synaptic + spatial components
// - Coordinate system accuracy in neural tissue space
//
// EXPERIMENTAL BASIS:
// - Cable theory: Conduction velocity = f(axon diameter, myelination)
// - Spatial propagation: delay = distance / conduction_velocity
// - Typical cortical speeds: 2 m/s local, 15 m/s long-range
func TestMatrixSpatialBasicCalculation(t *testing.T) {
	t.Log("=== SPATIAL TEST: Basic Distance and Delay Calculation ===")
	t.Log("Validating 3D spatial measurements and propagation timing")

	// Initialize matrix with spatial systems enabled
	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		SpatialEnabled: true,
		MaxComponents:  100,
	})

	// === REGISTER TEST NEURONS AT KNOWN POSITIONS ===
	t.Log("\n--- Registering Test Neurons ---")

	neuronA := ComponentInfo{
		ID:           "neuron_A",
		Type:         ComponentNeuron,
		Position:     Position3D{X: 0, Y: 0, Z: 0}, // Origin reference point
		State:        StateActive,
		RegisteredAt: time.Now(),
	}

	neuronB := ComponentInfo{
		ID:           "neuron_B",
		Type:         ComponentNeuron,
		Position:     Position3D{X: 100, Y: 0, Z: 0}, // 100 μm along X-axis
		State:        StateActive,
		RegisteredAt: time.Now(),
	}

	err := matrix.RegisterComponent(neuronA)
	if err != nil {
		t.Fatalf("Failed to register neuron A: %v", err)
	}

	err = matrix.RegisterComponent(neuronB)
	if err != nil {
		t.Fatalf("Failed to register neuron B: %v", err)
	}

	t.Logf("Registered neurons: A at origin, B at (100, 0, 0) μm")

	// === TEST 3D DISTANCE CALCULATION ===
	t.Log("\n--- Testing 3D Distance Calculation ---")

	distance, err := matrix.GetSpatialDistance("neuron_A", "neuron_B")
	if err != nil {
		t.Fatalf("Failed to calculate spatial distance: %v", err)
	}

	expectedDistance := 100.0 // μm
	tolerance := 0.001        // μm precision

	if math.Abs(distance-expectedDistance) > tolerance {
		t.Errorf("Distance calculation error: expected %.3f μm, measured %.3f μm",
			expectedDistance, distance)
	} else {
		t.Logf("✓ Distance calculation accurate: %.3f μm", distance)
	}

	// === TEST SPATIAL DELAY COMPUTATION ===
	t.Log("\n--- Testing Spatial Delay Computation ---")

	baseSynapticDelay := 1 * time.Millisecond
	totalDelay := matrix.SynapticDelay("neuron_A", "neuron_B", "test_synapse", baseSynapticDelay)

	// Expected calculation with default cortical speed (2000 μm/ms):
	// Spatial delay = 100 μm / 2000 μm/ms = 0.05ms
	// Total delay = 1ms (synaptic) + 0.05ms (spatial) = 1.05ms
	expectedSpatialDelay := 50 * time.Microsecond
	expectedTotalDelay := baseSynapticDelay + expectedSpatialDelay

	actualSpatialDelay := totalDelay - baseSynapticDelay

	if math.Abs(float64(totalDelay-expectedTotalDelay)) > float64(1*time.Microsecond) {
		t.Errorf("Total delay calculation error: expected %v, calculated %v",
			expectedTotalDelay, totalDelay)
	} else {
		t.Logf("✓ Delay calculation accurate:")
		t.Logf("  Base synaptic delay: %v", baseSynapticDelay)
		t.Logf("  Spatial propagation delay: %v", actualSpatialDelay)
		t.Logf("  Total transmission delay: %v", totalDelay)
	}

	t.Log("✅ Basic spatial calculations validated")
}

// =================================================================================
// TEST 2: THREE-DIMENSIONAL DISTANCE CALCULATIONS
// =================================================================================

// TestMatrixSpatialThreeDimensional validates 3D distance calculations across
// various geometric configurations in neural tissue space.
//
// BIOLOGICAL PROCESSES TESTED:
// - Zero distance (same position)
// - Single-axis distances (X, Y, Z directions)
// - Planar distances (2D diagonals)
// - Full 3D distances (space diagonals)
// - Mathematical accuracy of Euclidean distance formula
//
// EXPERIMENTAL BASIS:
// - Neural tissue is 3D: neurons positioned in cortical layers (Z), columns (X,Y)
// - Axonal pathways follow 3D trajectories through tissue
// - Distance measurement critical for realistic conduction delays
func TestMatrixSpatialThreeDimensional(t *testing.T) {
	t.Log("=== SPATIAL TEST: Three-Dimensional Distance Calculations ===")
	t.Log("Validating 3D geometry in neural tissue coordinate system")

	// Initialize spatial matrix
	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		SpatialEnabled: true,
		MaxComponents:  100,
	})

	// === DEFINE TEST CASES FOR 3D GEOMETRY ===
	t.Log("\n--- Testing Various 3D Geometric Configurations ---")

	testCases := []struct {
		name        string
		pos1        Position3D
		pos2        Position3D
		expected    float64
		description string
	}{
		{
			name:        "same_position",
			pos1:        Position3D{X: 0, Y: 0, Z: 0},
			pos2:        Position3D{X: 0, Y: 0, Z: 0},
			expected:    0.0,
			description: "Identical neuron positions (zero distance)",
		},
		{
			name:        "x_axis_distance",
			pos1:        Position3D{X: 0, Y: 0, Z: 0},
			pos2:        Position3D{X: 50, Y: 0, Z: 0},
			expected:    50.0,
			description: "X-axis separation (horizontal cortical distance)",
		},
		{
			name:        "y_axis_distance",
			pos1:        Position3D{X: 0, Y: 0, Z: 0},
			pos2:        Position3D{X: 0, Y: 30, Z: 0},
			expected:    30.0,
			description: "Y-axis separation (anterior-posterior distance)",
		},
		{
			name:        "z_axis_distance",
			pos1:        Position3D{X: 0, Y: 0, Z: 0},
			pos2:        Position3D{X: 0, Y: 0, Z: 25},
			expected:    25.0,
			description: "Z-axis separation (cortical layer distance)",
		},
		{
			name:        "planar_diagonal",
			pos1:        Position3D{X: 0, Y: 0, Z: 0},
			pos2:        Position3D{X: 3, Y: 4, Z: 0},
			expected:    5.0, // 3-4-5 right triangle
			description: "2D diagonal in cortical plane (3-4-5 triangle)",
		},
		{
			name:        "space_diagonal",
			pos1:        Position3D{X: 0, Y: 0, Z: 0},
			pos2:        Position3D{X: 10, Y: 10, Z: 10},
			expected:    math.Sqrt(300), // √(10² + 10² + 10²) ≈ 17.32
			description: "3D space diagonal (cube corner to corner)",
		},
		{
			name:        "realistic_cortical",
			pos1:        Position3D{X: 0, Y: 0, Z: 0},
			pos2:        Position3D{X: 200, Y: 150, Z: 50},
			expected:    math.Sqrt(200*200 + 150*150 + 50*50), // √67500 ≈ 259.8
			description: "Realistic cortical connection distance",
		},
	}

	// === EXECUTE TEST CASES ===
	for i, tc := range testCases {
		t.Logf("\n--- Test Case: %s ---", tc.name)
		t.Logf("Description: %s", tc.description)

		// Create unique neuron IDs for this test
		neuronID1 := fmt.Sprintf("test_neuron_%d_1", i)
		neuronID2 := fmt.Sprintf("test_neuron_%d_2", i)

		// Register neurons at test positions
		matrix.RegisterComponent(ComponentInfo{
			ID:           neuronID1,
			Type:         ComponentNeuron,
			Position:     tc.pos1,
			State:        StateActive,
			RegisteredAt: time.Now(),
		})
		matrix.RegisterComponent(ComponentInfo{
			ID:           neuronID2,
			Type:         ComponentNeuron,
			Position:     tc.pos2,
			State:        StateActive,
			RegisteredAt: time.Now(),
		})

		// Calculate distance
		distance, err := matrix.GetSpatialDistance(neuronID1, neuronID2)
		if err != nil {
			t.Errorf("Test case '%s': Failed to calculate distance: %v", tc.name, err)
			continue
		}

		// Validate accuracy
		tolerance := 0.001 // μm precision
		if math.Abs(distance-tc.expected) > tolerance {
			t.Errorf("Test case '%s': Distance error - expected %.3f μm, calculated %.3f μm",
				tc.name, tc.expected, distance)
		} else {
			t.Logf("✓ %s: %.3f μm (accurate)", tc.name, distance)
		}
	}

	t.Log("✅ 3D distance calculations validated across all geometries")
}

// =================================================================================
// TEST 3: AXONAL CONDUCTION VELOCITY TESTING
// =================================================================================

// TestMatrixSpatialAxonalConduction validates axonal conduction velocity effects
// on spatial delay calculations using biologically realistic speeds.
//
// BIOLOGICAL PROCESSES TESTED:
// - Unmyelinated axons: 0.5-2 m/s (pain fibers, local circuits)
// - Myelinated axons: 10-80 m/s (motor fibers, fast sensory)
// - Speed-distance-delay relationships
// - Biological velocity ranges and fiber type classification
//
// EXPERIMENTAL BASIS:
// - Conduction velocity measurements from electrophysiology
// - Fiber diameter and myelination determine speed
// - C fibers: 0.5-2 m/s, A-delta: 5-30 m/s, A-alpha: 30-120 m/s
func TestMatrixSpatialAxonalConduction(t *testing.T) {
	t.Log("=== SPATIAL TEST: Axonal Conduction Velocity Effects ===")
	t.Log("Validating conduction speed impact on spatial delays")

	// Initialize spatial matrix
	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		SpatialEnabled: true,
		MaxComponents:  100,
	})

	// === REGISTER TEST NEURONS AT STANDARD DISTANCE ===
	t.Log("\n--- Setting Up Standard Test Distance ---")

	testDistance := 1000.0 // μm (1mm - typical cortical connection)

	matrix.RegisterComponent(ComponentInfo{
		ID:           "speed_test_A",
		Type:         ComponentNeuron,
		Position:     Position3D{X: 0, Y: 0, Z: 0},
		State:        StateActive,
		RegisteredAt: time.Now(),
	})
	matrix.RegisterComponent(ComponentInfo{
		ID:           "speed_test_B",
		Type:         ComponentNeuron,
		Position:     Position3D{X: testDistance, Y: 0, Z: 0},
		State:        StateActive,
		RegisteredAt: time.Now(),
	})

	t.Logf("Test neurons separated by %.0f μm", testDistance)

	baseSynapticDelay := 1 * time.Millisecond

	// === TEST DIFFERENT AXONAL CONDUCTION SPEEDS ===
	t.Log("\n--- Testing Biological Axon Conduction Velocities ---")

	axonTests := []struct {
		name                 string
		speedUmPerMs         float64
		expectedSpatialDelay time.Duration
		fiberType            string
	}{
		{
			name:                 "Unmyelinated slow (C fibers)",
			speedUmPerMs:         UNMYELINATED_SLOW_SPEED,
			expectedSpatialDelay: 2 * time.Millisecond, // 1000μm / 500μm/ms = 2ms
			fiberType:            "Pain and temperature fibers",
		},
		{
			name:                 "Cortical local circuits",
			speedUmPerMs:         CORTICAL_LOCAL_SPEED,
			expectedSpatialDelay: 500 * time.Microsecond, // 1000μm / 2000μm/ms = 0.5ms
			fiberType:            "Local cortical connections",
		},
		{
			name:                 "Myelinated medium (A-delta)",
			speedUmPerMs:         MYELINATED_MEDIUM_SPEED,
			expectedSpatialDelay: 100 * time.Microsecond, // 1000μm / 10000μm/ms = 0.1ms
			fiberType:            "Fast pain and temperature",
		},
		{
			name:                 "Myelinated fast (A-alpha)",
			speedUmPerMs:         MYELINATED_FAST_SPEED,
			expectedSpatialDelay: 12500 * time.Nanosecond, // 1000μm / 80000μm/ms = 0.0125ms
			fiberType:            "Motor and proprioceptive fibers",
		},
	}

	for _, test := range axonTests {
		t.Logf("\n--- Testing: %s ---", test.name)
		t.Logf("Fiber type: %s", test.fiberType)
		t.Logf("Conduction speed: %.0f μm/ms (%.1f m/s)", test.speedUmPerMs, test.speedUmPerMs/1000.0)

		// Set axon speed for this test
		matrix.SetAxonSpeed(test.speedUmPerMs)

		// Calculate total delay
		totalDelay := matrix.SynapticDelay("speed_test_A", "speed_test_B", "test_synapse", baseSynapticDelay)
		actualSpatialDelay := totalDelay - baseSynapticDelay

		t.Logf("Expected spatial delay: %v", test.expectedSpatialDelay)
		t.Logf("Actual spatial delay: %v", actualSpatialDelay)
		t.Logf("Total transmission delay: %v", totalDelay)

		// Validate accuracy (1% tolerance for timing precision)
		tolerance := float64(test.expectedSpatialDelay) * 0.01
		delayDifference := math.Abs(float64(actualSpatialDelay - test.expectedSpatialDelay))

		if delayDifference > tolerance {
			t.Errorf("%s: Spatial delay error - expected %v, calculated %v (difference: %.3fμs)",
				test.name, test.expectedSpatialDelay, actualSpatialDelay, delayDifference/float64(time.Microsecond))
		} else {
			t.Logf("✓ %s: Spatial delay calculation accurate", test.name)
		}
	}

	t.Log("✅ Axonal conduction velocity effects validated")
}

// =================================================================================
// TEST 4: BIOLOGICAL AXON TYPE PRESETS
// =================================================================================

// TestMatrixSpatialBiologicalPresets validates predefined biological axon types
// with experimentally-measured conduction velocities.
//
// BIOLOGICAL PROCESSES TESTED:
// - Biological axon type classification system
// - Preset conduction velocities from literature
// - Functional pathway speed differences
// - Easy selection of realistic parameters
//
// EXPERIMENTAL BASIS:
// - Axon type classification from neurophysiology literature
// - Preset speeds match published experimental measurements
// - Different pathways have characteristic conduction properties
func TestMatrixSpatialBiologicalPresets(t *testing.T) {
	t.Log("=== SPATIAL TEST: Biological Axon Type Presets ===")
	t.Log("Validating predefined biological conduction velocities")

	// Initialize spatial matrix
	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		SpatialEnabled: true,
		MaxComponents:  100,
	})

	// === REGISTER TEST NEURONS FOR CONSISTENT TESTING ===
	t.Log("\n--- Setting Up Test Neurons ---")

	testDistance := 2000.0 // μm (2mm - longer cortical projection)

	matrix.RegisterComponent(ComponentInfo{
		ID:           "bio_test_A",
		Type:         ComponentNeuron,
		Position:     Position3D{X: 0, Y: 0, Z: 0},
		State:        StateActive,
		RegisteredAt: time.Now(),
	})
	matrix.RegisterComponent(ComponentInfo{
		ID:           "bio_test_B",
		Type:         ComponentNeuron,
		Position:     Position3D{X: testDistance, Y: 0, Z: 0},
		State:        StateActive,
		RegisteredAt: time.Now(),
	})

	t.Logf("Test separation: %.0f μm (%.1f mm)", testDistance, testDistance/1000.0)

	baseSynapticDelay := 1 * time.Millisecond

	// === TEST BIOLOGICAL AXON TYPE PRESETS ===
	t.Log("\n--- Testing Biological Axon Type Presets ---")

	biologicalPresets := []struct {
		axonType       string
		expectedSpeed  float64 // μm/ms
		pathwayType    string
		functionalRole string
	}{
		{
			axonType:       "unmyelinated_slow",
			expectedSpeed:  UNMYELINATED_SLOW_SPEED,
			pathwayType:    "C fiber pathways",
			functionalRole: "Pain and temperature sensation",
		},
		{
			axonType:       "unmyelinated_fast",
			expectedSpeed:  UNMYELINATED_FAST_SPEED,
			pathwayType:    "Local cortical circuits",
			functionalRole: "Intracortical processing",
		},
		{
			axonType:       "cortical_local",
			expectedSpeed:  CORTICAL_LOCAL_SPEED,
			pathwayType:    "Within-column connections",
			functionalRole: "Local cortical computation",
		},
		{
			axonType:       "cortical_inter",
			expectedSpeed:  CORTICAL_INTER_SPEED,
			pathwayType:    "Inter-laminar connections",
			functionalRole: "Between cortical layers",
		},
		{
			axonType:       "long_range",
			expectedSpeed:  CORTICAL_LONG_SPEED,
			pathwayType:    "Cortico-cortical projections",
			functionalRole: "Long-distance cortical communication",
		},
	}

	for _, preset := range biologicalPresets {
		t.Logf("\n--- Testing Preset: %s ---", preset.axonType)
		t.Logf("Pathway: %s", preset.pathwayType)
		t.Logf("Function: %s", preset.functionalRole)

		// Apply biological preset
		matrix.SetBiologicalAxonType(preset.axonType)

		// Calculate delays
		totalDelay := matrix.SynapticDelay("bio_test_A", "bio_test_B", "bio_synapse", baseSynapticDelay)
		spatialDelay := totalDelay - baseSynapticDelay

		// Expected spatial delay = distance / speed
		expectedSpatialDelayMs := testDistance / preset.expectedSpeed
		expectedSpatialDelay := time.Duration(expectedSpatialDelayMs * float64(time.Millisecond))

		t.Logf("Conduction speed: %.0f μm/ms (%.1f m/s)", preset.expectedSpeed, preset.expectedSpeed/1000.0)
		t.Logf("Expected spatial delay: %v", expectedSpatialDelay)
		t.Logf("Actual spatial delay: %v", spatialDelay)

		// Validate accuracy (1% tolerance)
		tolerance := float64(expectedSpatialDelay) * 0.01
		delayDifference := math.Abs(float64(spatialDelay - expectedSpatialDelay))

		if delayDifference > tolerance {
			t.Errorf("Preset %s: Delay error - expected %v, calculated %v",
				preset.axonType, expectedSpatialDelay, spatialDelay)
		} else {
			t.Logf("✓ Preset %s: Accurate delay calculation", preset.axonType)
		}
	}

	t.Log("✅ Biological axon type presets validated")
}

// =================================================================================
// TEST 5: ERROR HANDLING AND EDGE CASES
// =================================================================================

// TestMatrixSpatialErrorHandling validates robust error handling for spatial
// calculations with missing components and edge cases.
//
// ERROR CONDITIONS TESTED:
// - Non-existent neurons in distance calculations
// - Missing presynaptic neurons in delay calculations
// - Missing postsynaptic neurons in delay calculations
// - Same neuron distance calculations (zero distance)
// - Graceful fallback to base delays when spatial data unavailable
//
// EXPECTED BEHAVIORS:
// - Functions return appropriate errors for missing components
// - Delay calculations fallback to base synaptic delay when spatial data missing
// - Zero distance calculations handled correctly
// - System remains stable under error conditions
func TestMatrixSpatialErrorHandling(t *testing.T) {
	t.Log("=== SPATIAL TEST: Error Handling and Edge Cases ===")
	t.Log("Validating robust behavior with missing components and edge cases")

	// Initialize minimal spatial matrix
	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		SpatialEnabled: true,
		MaxComponents:  100,
	})

	// === REGISTER SINGLE TEST NEURON ===
	t.Log("\n--- Setting Up Limited Test Environment ---")

	matrix.RegisterComponent(ComponentInfo{
		ID:           "existing_neuron",
		Type:         ComponentNeuron,
		Position:     Position3D{X: 0, Y: 0, Z: 0},
		State:        StateActive,
		RegisteredAt: time.Now(),
	})

	t.Logf("Registered single test neuron: existing_neuron")

	// === TEST NON-EXISTENT NEURON HANDLING ===
	t.Log("\n--- Testing Non-Existent Neuron Handling ---")

	baseSynapticDelay := 1 * time.Millisecond

	// Test with non-existent presynaptic neuron
	t.Log("Testing non-existent presynaptic neuron...")
	totalDelay := matrix.SynapticDelay("nonexistent_pre", "existing_neuron", "test_synapse", baseSynapticDelay)

	if totalDelay != baseSynapticDelay {
		t.Errorf("Expected fallback to base delay (%v) when presynaptic neuron missing, got %v",
			baseSynapticDelay, totalDelay)
	} else {
		t.Logf("✓ Correctly returned base delay when presynaptic neuron missing")
	}

	// Test with non-existent postsynaptic neuron
	t.Log("Testing non-existent postsynaptic neuron...")
	totalDelay = matrix.SynapticDelay("existing_neuron", "nonexistent_post", "test_synapse", baseSynapticDelay)

	if totalDelay != baseSynapticDelay {
		t.Errorf("Expected fallback to base delay (%v) when postsynaptic neuron missing, got %v",
			baseSynapticDelay, totalDelay)
	} else {
		t.Logf("✓ Correctly returned base delay when postsynaptic neuron missing")
	}

	// Test with both neurons non-existent
	t.Log("Testing both neurons non-existent...")
	totalDelay = matrix.SynapticDelay("nonexistent_pre", "nonexistent_post", "test_synapse", baseSynapticDelay)

	if totalDelay != baseSynapticDelay {
		t.Errorf("Expected fallback to base delay (%v) when both neurons missing, got %v",
			baseSynapticDelay, totalDelay)
	} else {
		t.Logf("✓ Correctly returned base delay when both neurons missing")
	}

	// === TEST DISTANCE CALCULATION ERROR HANDLING ===
	t.Log("\n--- Testing Distance Calculation Error Handling ---")

	// Test distance between same neuron (should be zero)
	distance, err := matrix.GetSpatialDistance("existing_neuron", "existing_neuron")
	if err != nil {
		t.Errorf("Unexpected error for same neuron distance calculation: %v", err)
	} else if distance != 0.0 {
		t.Errorf("Expected zero distance for same neuron, got %.3f μm", distance)
	} else {
		t.Logf("✓ Zero distance correctly calculated for same neuron")
	}

	// Test distance with non-existent neurons
	_, err = matrix.GetSpatialDistance("existing_neuron", "nonexistent_neuron")
	if err == nil {
		t.Errorf("Expected error for distance calculation with non-existent neuron")
	} else {
		t.Logf("✓ Correctly returned error for non-existent neuron: %v", err)
	}

	_, err = matrix.GetSpatialDistance("nonexistent_neuron", "existing_neuron")
	if err == nil {
		t.Errorf("Expected error for distance calculation with non-existent neuron")
	} else {
		t.Logf("✓ Correctly returned error for non-existent neuron: %v", err)
	}

	t.Log("✅ Error handling validated - system robust under all tested conditions")
}

// =================================================================================
// TEST 6: REALISTIC CORTICAL CIRCUIT SCENARIOS
// =================================================================================

// TestMatrixSpatialCorticalScenarios validates spatial delay calculations for
// realistic cortical circuit pathways and connection distances.
//
// BIOLOGICAL PROCESSES TESTED:
// - Local cortical circuits: 20-100 μm (same minicolumn)
// - Nearby column connections: 100-500 μm (adjacent columns)
// - Same area connections: 500-2000 μm (within cortical area)
// - Cross-area projections: 2-20 mm (between cortical areas)
//
// EXPERIMENTAL BASIS:
// - Cortical column organization: 50-100 μm diameter minicolumns
// - Local circuit distances from anatomical studies
// - Cortico-cortical projection lengths from tract tracing
// - Realistic delay ranges for cortical computation
func TestMatrixSpatialCorticalScenarios(t *testing.T) {
	t.Log("=== SPATIAL TEST: Realistic Cortical Circuit Scenarios ===")
	t.Log("Validating spatial delays for authentic cortical pathways")

	// Initialize matrix with cortical parameters
	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		SpatialEnabled: true,
		MaxComponents:  1000,
	})

	// Set cortical local circuit speed (2 m/s typical)
	matrix.SetBiologicalAxonType("cortical_local")

	t.Logf("Using cortical local circuit speed: %.1f m/s", CORTICAL_LOCAL_SPEED/1000.0)

	baseSynapticDelay := DEFAULT_SYNAPTIC_DELAY

	// === DEFINE REALISTIC CORTICAL SCENARIOS ===
	t.Log("\n--- Testing Realistic Cortical Connection Scenarios ---")

	corticalScenarios := []struct {
		name          string
		distance      float64 // μm
		description   string
		maxTotalDelay time.Duration
		circuitType   string
	}{
		{
			name:          "local_circuit",
			distance:      20.0,
			description:   "Local cortical circuit (same minicolumn)",
			maxTotalDelay: 1 * time.Millisecond,
			circuitType:   "Intracolumnar processing",
		},
		{
			name:          "nearby_column",
			distance:      100.0,
			description:   "Nearby cortical column connection",
			maxTotalDelay: 1 * time.Millisecond,
			circuitType:   "Adjacent column interaction",
		},
		{
			name:          "same_area",
			distance:      500.0,
			description:   "Within same cortical area",
			maxTotalDelay: 1 * time.Millisecond,
			circuitType:   "Intra-areal processing",
		},
		{
			name:          "cross_area_short",
			distance:      2000.0,
			description:   "Cross-area connection (2mm)",
			maxTotalDelay: 2 * time.Millisecond,
			circuitType:   "Inter-areal communication",
		},
		{
			name:          "long_projection",
			distance:      5000.0,
			description:   "Long-range cortical projection (5mm)",
			maxTotalDelay: 4 * time.Millisecond,
			circuitType:   "Long-distance coordination",
		},
	}

	// === EXECUTE CORTICAL SCENARIO TESTS ===
	for i, scenario := range corticalScenarios {
		t.Logf("\n--- Scenario: %s ---", scenario.name)
		t.Logf("Description: %s", scenario.description)
		t.Logf("Circuit type: %s", scenario.circuitType)
		t.Logf("Connection distance: %.1f μm", scenario.distance)

		// Create neuron pair for this scenario
		neuronID1 := fmt.Sprintf("cortical_%d_source", i)
		neuronID2 := fmt.Sprintf("cortical_%d_target", i)

		// Position neurons at specified distance
		matrix.RegisterComponent(ComponentInfo{
			ID:           neuronID1,
			Type:         ComponentNeuron,
			Position:     Position3D{X: 0, Y: 0, Z: 0},
			State:        StateActive,
			RegisteredAt: time.Now(),
		})
		matrix.RegisterComponent(ComponentInfo{
			ID:           neuronID2,
			Type:         ComponentNeuron,
			Position:     Position3D{X: scenario.distance, Y: 0, Z: 0},
			State:        StateActive,
			RegisteredAt: time.Now(),
		})

		// Calculate transmission delays
		totalDelay := matrix.SynapticDelay(neuronID1, neuronID2, "cortical_synapse", baseSynapticDelay)
		spatialDelay := totalDelay - baseSynapticDelay

		// Expected spatial delay calculation
		expectedSpatialDelayMs := scenario.distance / CORTICAL_LOCAL_SPEED
		expectedSpatialDelay := time.Duration(expectedSpatialDelayMs * float64(time.Millisecond))

		t.Logf("Expected spatial delay: %v", expectedSpatialDelay)
		t.Logf("Actual spatial delay: %v", spatialDelay)
		t.Logf("Total transmission delay: %v", totalDelay)

		// Validate delay is within biological range
		if totalDelay > scenario.maxTotalDelay {
			t.Errorf("Scenario %s: Total delay too high (%v > %v)",
				scenario.name, totalDelay, scenario.maxTotalDelay)
		} else {
			t.Logf("✓ %s: Delay within biological range", scenario.name)
		}

		// Validate spatial delay accuracy
		tolerance := float64(expectedSpatialDelay) * 0.01 // 1% tolerance
		delayDifference := math.Abs(float64(spatialDelay - expectedSpatialDelay))

		if delayDifference > tolerance {
			t.Errorf("Scenario %s: Spatial delay calculation error - expected %v, got %v",
				scenario.name, expectedSpatialDelay, spatialDelay)
		} else {
			t.Logf("✓ %s: Spatial delay calculation accurate", scenario.name)
		}
	}

	// === CORTICAL TIMING SUMMARY ===
	t.Log("\n--- Cortical Circuit Timing Summary ---")
	t.Log("Circuit Type           | Distance | Spatial Delay | Total Delay(")
	t.Log("----------------------|----------|---------------|-------------")

	for i, scenario := range corticalScenarios {
		neuronID1 := fmt.Sprintf("cortical_%d_source", i)
		neuronID2 := fmt.Sprintf("cortical_%d_target", i)

		totalDelay := matrix.SynapticDelay(neuronID1, neuronID2, "summary(", baseSynapticDelay)
		spatialDelay := totalDelay - baseSynapticDelay

		t.Logf("%-20s | %6.0f μm | %10v | %10v",
			scenario.circuitType[:20], scenario.distance, spatialDelay, totalDelay)
	}

	t.Log("✅ All cortical scenarios validated successfully(")
}

// =================================================================================
// UTILITY FUNCTIONS FOR SPATIAL TESTING
// =================================================================================

// validateSpatialAccuracy checks if calculated spatial delay matches expected value
func validateSpatialAccuracy(t *testing.T, name string, calculated, expected time.Duration, tolerance float64) bool {
	difference := math.Abs(float64(calculated - expected))
	maxDifference := float64(expected) * tolerance

	if difference > maxDifference {
		t.Errorf("%s: Spatial timing error - expected %v, calculated %v (difference: %.1fμs)",
			name, expected, calculated, difference/float64(time.Microsecond))
		return false
	}

	t.Logf("✓ %s: Spatial timing accurate (difference: %.1fμs)",
		name, difference/float64(time.Microsecond))
	return true
}

// calculateExpectedDelay computes expected spatial delay from distance and speed
func calculateExpectedDelay(distance float64, speedUmPerMs float64) time.Duration {
	if speedUmPerMs <= 0 {
		return 0
	}
	delayMs := distance / speedUmPerMs
	return time.Duration(delayMs * float64(time.Millisecond))
}

// logSpatialTest provides consistent logging format for spatial tests
func logSpatialTest(t *testing.T, testName string, distance float64, speed float64,
	spatialDelay time.Duration, totalDelay time.Duration) {
	t.Logf("--- %s ---", testName)
	t.Logf("  Distance: %.1f μm", distance)
	t.Logf("  Speed: %.1f μm/ms (%.2f m/s)", speed, speed/1000.0)
	t.Logf("  Spatial delay: %v", spatialDelay)
	t.Logf("  Total delay: %v", totalDelay)
}
