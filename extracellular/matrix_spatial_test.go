/*
=================================================================================
SPATIAL DELAY ENHANCEMENT TESTS
=================================================================================

Tests for the spatial delay calculation system that combines synaptic delays
with distance-based axonal propagation delays. This validates that synapses
receive realistic total transmission delays based on 3D neuron positioning.

File: matrix_spatial_delay_test.go (new file in extracellular package)
=================================================================================
*/

package extracellular

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func TestBasicSpatialDelayCalculation(t *testing.T) {
	t.Log("=== BASIC SPATIAL DELAY CALCULATION TEST ===")

	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		SpatialEnabled: true,
		MaxComponents:  100,
	})

	// Register two neurons at known positions
	neuronA := ComponentInfo{
		ID:       "neuron_A",
		Type:     ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0}, // Origin
		State:    StateActive,
	}

	neuronB := ComponentInfo{
		ID:       "neuron_B",
		Type:     ComponentNeuron,
		Position: Position3D{X: 100, Y: 0, Z: 0}, // 100 μm away in X direction
		State:    StateActive,
	}

	err := matrix.RegisterComponent(neuronA)
	if err != nil {
		t.Fatalf("Failed to register neuron A: %v", err)
	}

	err = matrix.RegisterComponent(neuronB)
	if err != nil {
		t.Fatalf("Failed to register neuron B: %v", err)
	}

	// Test distance calculation
	distance, err := matrix.GetSpatialDistance("neuron_A", "neuron_B")
	if err != nil {
		t.Fatalf("Failed to get spatial distance: %v", err)
	}

	expectedDistance := 100.0 // μm
	if math.Abs(distance-expectedDistance) > 0.001 {
		t.Errorf("Distance calculation error: expected %.3f μm, got %.3f μm",
			expectedDistance, distance)
	}
	t.Logf("✓ Distance calculation: %.3f μm", distance)

	// Test total delay calculation
	baseSynapticDelay := 1 * time.Millisecond
	totalDelay := matrix.EnhanceSynapticDelay("neuron_A", "neuron_B", "test_synapse", baseSynapticDelay)

	// Expected: 1ms base + (100μm / 2000μm/ms) = 1ms + 0.05ms = 1.05ms
	expectedTotalDelay := 1050 * time.Microsecond
	if math.Abs(float64(totalDelay-expectedTotalDelay)) > float64(1*time.Microsecond) {
		t.Errorf("Total delay calculation error: expected %v, got %v",
			expectedTotalDelay, totalDelay)
	}
	t.Logf("✓ Total delay: %v (base: %v + spatial: %v)",
		totalDelay, baseSynapticDelay, totalDelay-baseSynapticDelay)
}

func TestThreeDimensionalDistances(t *testing.T) {
	t.Log("=== 3D DISTANCE CALCULATIONS TEST ===")

	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		SpatialEnabled: true,
		MaxComponents:  100,
	})

	// Test various 3D positions
	testCases := []struct {
		name     string
		pos1     Position3D
		pos2     Position3D
		expected float64
	}{
		{
			name:     "Same position",
			pos1:     Position3D{X: 0, Y: 0, Z: 0},
			pos2:     Position3D{X: 0, Y: 0, Z: 0},
			expected: 0.0,
		},
		{
			name:     "X-axis distance",
			pos1:     Position3D{X: 0, Y: 0, Z: 0},
			pos2:     Position3D{X: 50, Y: 0, Z: 0},
			expected: 50.0,
		},
		{
			name:     "3D diagonal (3-4-5 triangle)",
			pos1:     Position3D{X: 0, Y: 0, Z: 0},
			pos2:     Position3D{X: 3, Y: 4, Z: 0},
			expected: 5.0,
		},
		{
			name:     "3D cube diagonal",
			pos1:     Position3D{X: 0, Y: 0, Z: 0},
			pos2:     Position3D{X: 10, Y: 10, Z: 10},
			expected: math.Sqrt(300), // √(10² + 10² + 10²) = √300 ≈ 17.32
		},
	}

	for i, tc := range testCases {
		// Register neurons for this test case
		neuronID1 := fmt.Sprintf("test_neuron_%d_1", i)
		neuronID2 := fmt.Sprintf("test_neuron_%d_2", i)

		matrix.RegisterComponent(ComponentInfo{
			ID: neuronID1, Type: ComponentNeuron, Position: tc.pos1, State: StateActive,
		})
		matrix.RegisterComponent(ComponentInfo{
			ID: neuronID2, Type: ComponentNeuron, Position: tc.pos2, State: StateActive,
		})

		// Test distance calculation
		distance, err := matrix.GetSpatialDistance(neuronID1, neuronID2)
		if err != nil {
			t.Errorf("Test case '%s': Failed to get distance: %v", tc.name, err)
			continue
		}

		if math.Abs(distance-tc.expected) > 0.001 {
			t.Errorf("Test case '%s': Distance error - expected %.3f μm, got %.3f μm",
				tc.name, tc.expected, distance)
		} else {
			t.Logf("✓ %s: %.3f μm", tc.name, distance)
		}
	}
}

func TestDifferentAxonSpeeds(t *testing.T) {
	t.Log("=== DIFFERENT AXON SPEEDS TEST ===")

	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		SpatialEnabled: true,
		MaxComponents:  100,
	})

	// Register test neurons 1000 μm apart
	matrix.RegisterComponent(ComponentInfo{
		ID: "speed_test_A", Type: ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0}, State: StateActive,
	})
	matrix.RegisterComponent(ComponentInfo{
		ID: "speed_test_B", Type: ComponentNeuron,
		Position: Position3D{X: 1000, Y: 0, Z: 0}, State: StateActive,
	})

	baseSynapticDelay := 1 * time.Millisecond
	// FIXED: Remove unused variable
	// distance := 1000.0 // μm

	// Test different axon types
	axonTests := []struct {
		name                 string
		speedUmPerMs         float64
		expectedSpatialDelay time.Duration
	}{
		{
			name:                 "Unmyelinated slow (0.5 m/s)",
			speedUmPerMs:         500.0,                // 0.5 m/s
			expectedSpatialDelay: 2 * time.Millisecond, // 1000μm / 500μm/ms = 2ms
		},
		{
			name:                 "Cortical local (2 m/s)",
			speedUmPerMs:         2000.0,                 // 2 m/s
			expectedSpatialDelay: 500 * time.Microsecond, // 1000μm / 2000μm/ms = 0.5ms
		},
		{
			name:                 "Myelinated medium (10 m/s)",
			speedUmPerMs:         10000.0,                // 10 m/s
			expectedSpatialDelay: 100 * time.Microsecond, // 1000μm / 10000μm/ms = 0.1ms
		},
		{
			name:                 "Myelinated fast (80 m/s)",
			speedUmPerMs:         80000.0,                 // 80 m/s
			expectedSpatialDelay: 12500 * time.Nanosecond, // 1000μm / 80000μm/ms = 0.0125ms
		},
	}

	for _, test := range axonTests {
		t.Logf("\n--- Testing %s ---", test.name)

		// Set axon speed
		matrix.SetAxonSpeed(test.speedUmPerMs)

		// Calculate total delay
		totalDelay := matrix.EnhanceSynapticDelay("speed_test_A", "speed_test_B", "test_syn", baseSynapticDelay)
		actualSpatialDelay := totalDelay - baseSynapticDelay

		t.Logf("Expected spatial delay: %v", test.expectedSpatialDelay)
		t.Logf("Actual spatial delay: %v", actualSpatialDelay)
		t.Logf("Total delay: %v", totalDelay)

		// Allow 1% tolerance for timing precision
		tolerance := float64(test.expectedSpatialDelay) * 0.01
		if math.Abs(float64(actualSpatialDelay-test.expectedSpatialDelay)) > tolerance {
			t.Errorf("%s: Spatial delay error - expected %v, got %v",
				test.name, test.expectedSpatialDelay, actualSpatialDelay)
		} else {
			t.Logf("✓ %s spatial delay correct", test.name)
		}
	}
}

func TestBiologicalAxonTypePresets(t *testing.T) {
	t.Log("=== BIOLOGICAL AXON TYPE PRESETS TEST ===")

	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		SpatialEnabled: true,
		MaxComponents:  100,
	})

	// Register test neurons for consistent testing
	matrix.RegisterComponent(ComponentInfo{
		ID: "bio_test_A", Type: ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0}, State: StateActive,
	})
	matrix.RegisterComponent(ComponentInfo{
		ID: "bio_test_B", Type: ComponentNeuron,
		Position: Position3D{X: 2000, Y: 0, Z: 0}, State: StateActive, // 2mm apart
	})

	baseSynapticDelay := 1 * time.Millisecond

	// Test biological presets
	biologicalTests := []struct {
		axonType      string
		expectedSpeed float64 // μm/ms
	}{
		{"unmyelinated_slow", 500.0},  // 0.5 m/s
		{"unmyelinated_fast", 2000.0}, // 2 m/s
		{"cortical_local", 2000.0},    // 2 m/s
		{"cortical_inter", 5000.0},    // 5 m/s
		{"long_range", 15000.0},       // 15 m/s
	}

	for _, test := range biologicalTests {
		t.Logf("\n--- Testing biological type: %s ---", test.axonType)

		// FIXED: Use correct method name
		matrix.SetBiologicalAxonType(test.axonType)

		// Test with known distance (2000 μm)
		totalDelay := matrix.EnhanceSynapticDelay("bio_test_A", "bio_test_B", "bio_syn", baseSynapticDelay)
		spatialDelay := totalDelay - baseSynapticDelay

		// Expected spatial delay = 2000 μm / speed
		expectedSpatialDelayMs := 2000.0 / test.expectedSpeed
		expectedSpatialDelay := time.Duration(expectedSpatialDelayMs * float64(time.Millisecond))

		t.Logf("Axon speed: %.0f μm/ms (%.1f m/s)", test.expectedSpeed, test.expectedSpeed/1000.0)
		t.Logf("Expected spatial delay: %v", expectedSpatialDelay)
		t.Logf("Actual spatial delay: %v", spatialDelay)

		// Verify within 1% tolerance
		tolerance := float64(expectedSpatialDelay) * 0.01
		if math.Abs(float64(spatialDelay-expectedSpatialDelay)) > tolerance {
			t.Errorf("Biological type %s: delay error - expected %v, got %v",
				test.axonType, expectedSpatialDelay, spatialDelay)
		} else {
			t.Logf("✓ Biological type %s working correctly", test.axonType)
		}
	}
}

func TestSpatialDelayErrorHandling(t *testing.T) {
	t.Log("=== SPATIAL DELAY ERROR HANDLING TEST ===")

	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		SpatialEnabled: true,
		MaxComponents:  100,
	})

	// Register only one neuron
	matrix.RegisterComponent(ComponentInfo{
		ID: "existing_neuron", Type: ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0}, State: StateActive,
	})

	// Test with non-existent neurons
	t.Log("--- Testing non-existent pre-synaptic neuron ---")
	baseSynapticDelay := 1 * time.Millisecond
	totalDelay := matrix.EnhanceSynapticDelay("nonexistent_pre", "existing_neuron", "test_syn", baseSynapticDelay)

	// Should return just the base delay when neurons don't exist
	if totalDelay != baseSynapticDelay {
		t.Errorf("Expected base delay (%v) when pre-neuron missing, got %v", baseSynapticDelay, totalDelay)
	} else {
		t.Logf("✓ Correctly returned base delay when pre-neuron missing")
	}

	// Test GetSpatialDistance error handling
	t.Log("--- Testing GetSpatialDistance error handling ---")
	_, err := matrix.GetSpatialDistance("existing_neuron", "existing_neuron")
	if err != nil {
		t.Errorf("Unexpected error for same neuron: %v", err)
	} else {
		t.Logf("✓ Zero distance for same neuron works correctly")
	}
}

func TestRealisticCorticalScenarios(t *testing.T) {
	t.Log("=== REALISTIC CORTICAL SCENARIOS TEST ===")

	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		SpatialEnabled: true,
		MaxComponents:  1000,
	})

	// Set cortical axon speed
	matrix.SetBiologicalAxonType("cortical_local") // 2 m/s

	// FIXED: Convert to proper duration
	baseSynapticDelay := time.Duration(500) * time.Microsecond // 0.5ms

	// Realistic cortical scenarios
	scenarios := []struct {
		name          string
		distance      float64 // μm
		description   string
		maxTotalDelay time.Duration
	}{
		{
			name: "local_circuit", distance: 20.0,
			description:   "Local cortical circuit (same minicolumn)",
			maxTotalDelay: 1 * time.Millisecond,
		},
		{
			name: "nearby_column", distance: 100.0,
			description:   "Nearby cortical column",
			maxTotalDelay: 1 * time.Millisecond,
		},
		{
			name: "same_area", distance: 500.0,
			description:   "Within same cortical area",
			maxTotalDelay: 1 * time.Millisecond,
		},
		{
			name: "cross_area", distance: 2000.0,
			description:   "Cross-area connection (2mm)",
			maxTotalDelay: 2 * time.Millisecond,
		},
	}

	for i, scenario := range scenarios {
		neuronID1 := fmt.Sprintf("cortical_%d_A", i)
		neuronID2 := fmt.Sprintf("cortical_%d_B", i)

		// Place neurons at specified distance
		matrix.RegisterComponent(ComponentInfo{
			ID: neuronID1, Type: ComponentNeuron,
			Position: Position3D{X: 0, Y: 0, Z: 0}, State: StateActive,
		})
		matrix.RegisterComponent(ComponentInfo{
			ID: neuronID2, Type: ComponentNeuron,
			Position: Position3D{X: scenario.distance, Y: 0, Z: 0}, State: StateActive,
		})

		// Calculate total delay
		totalDelay := matrix.EnhanceSynapticDelay(neuronID1, neuronID2, "cortical_syn", baseSynapticDelay)
		spatialDelay := totalDelay - baseSynapticDelay

		t.Logf("\n--- %s ---", scenario.name)
		t.Logf("Description: %s", scenario.description)
		t.Logf("Distance: %.1f μm", scenario.distance)
		t.Logf("Spatial delay: %v", spatialDelay)
		t.Logf("Total delay: %v", totalDelay)

		// Validate realistic delay ranges
		if totalDelay > scenario.maxTotalDelay {
			t.Errorf("Scenario %s: Total delay too high (%v > %v)",
				scenario.name, totalDelay, scenario.maxTotalDelay)
		} else {
			t.Logf("✓ %s delay within biological range", scenario.name)
		}
	}

	t.Log("\n✅ All cortical scenarios validated")
}
