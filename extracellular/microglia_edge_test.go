/*
=================================================================================
MICROGLIA - EDGE CASES AND ERROR CONDITIONS TESTS
=================================================================================

Comprehensive tests for edge cases, error conditions, and boundary testing of
the microglia lifecycle management system. These tests ensure robustness and
graceful handling of unusual or pathological conditions.

Test Categories:
1. Configuration Edge Cases - Extreme and invalid configurations
2. Resource Exhaustion - Memory, component, and processing limits
3. Network Topology Edge Cases - Empty networks, isolated components
4. Concurrent Access Stress - Race conditions and deadlock prevention
5. Data Corruption Scenarios - Invalid states and malformed data
6. Performance Boundaries - Large-scale operations and timeouts
7. Error Recovery - Graceful degradation and self-healing
8. New Function Edge Cases - Redundancy and metabolic cost calculations

Research Foundation:
- Stress testing methodologies for distributed systems
- Fault injection and chaos engineering principles
- Biological system resilience and adaptation mechanisms
- Network topology analysis and graph theory applications

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
// CONFIGURATION EDGE CASES
// =================================================================================

func TestMicrogliaEdgeExtremeConfigurations(t *testing.T) {
	t.Log("=== TESTING EXTREME CONFIGURATION EDGE CASES ===")

	astrocyteNetwork := NewAstrocyteNetwork()

	// Test zero/negative values
	extremeConfig := GetDefaultMicrogliaConfig()
	extremeConfig.HealthThresholds.CriticalActivityThreshold = -1.0 // Invalid negative
	extremeConfig.HealthThresholds.VeryLowActivityThreshold = 2.0   // Invalid > 1.0
	extremeConfig.PruningSettings.ActivityWeight = -0.5             // Invalid negative
	extremeConfig.ResourceLimits.MaxComponents = -100               // Invalid negative

	// Should handle gracefully without crashing
	microglia := NewMicrogliaWithConfig(astrocyteNetwork, extremeConfig)

	// Test basic operations still work
	componentInfo := ComponentInfo{
		ID:       "extreme_test_neuron",
		Type:     ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0},
		State:    StateActive,
	}

	// Should not crash even with extreme config
	err := microglia.CreateComponent(componentInfo)
	if err != nil {
		t.Logf("INFO: Component creation failed with extreme config (expected): %v", err)
	}

	// Test health calculation with extreme values
	microglia.UpdateComponentHealth("extreme_test_neuron", 1.5, -5) // Out of range values
	health, exists := microglia.GetComponentHealth("extreme_test_neuron")
	if exists {
		if health.HealthScore < 0 || health.HealthScore > 1 {
			t.Errorf("Health score should be clamped to 0-1, got %.3f", health.HealthScore)
		}
	}

	t.Log("✓ Extreme configurations handled gracefully")
}

func TestMicrogliaEdgeInfiniteAndNaNValues(t *testing.T) {
	t.Log("=== TESTING INFINITE AND NaN VALUE HANDLING ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Create test component
	componentInfo := ComponentInfo{
		ID:       "nan_test_neuron",
		Type:     ComponentNeuron,
		Position: Position3D{X: math.Inf(1), Y: math.NaN(), Z: -math.Inf(1)},
		State:    StateActive,
	}

	err := microglia.CreateComponent(componentInfo)
	if err != nil {
		t.Logf("INFO: Component creation with infinite position failed (expected): %v", err)
	}

	// Test health update with NaN/infinite values
	microglia.UpdateComponentHealth("nan_test_neuron", math.NaN(), 1000000)
	health, exists := microglia.GetComponentHealth("nan_test_neuron")
	if exists {
		if math.IsNaN(health.HealthScore) || math.IsInf(health.HealthScore, 0) {
			t.Error("Health score should not be NaN or infinite")
		}
	}

	// Test pruning with extreme activity levels
	microglia.MarkForPruning("inf_synapse", "source", "target", math.Inf(1))
	candidates := microglia.GetPruningCandidates()
	for _, candidate := range candidates {
		if math.IsNaN(candidate.PruningScore) || math.IsInf(candidate.PruningScore, 0) {
			t.Error("Pruning score should not be NaN or infinite")
		}
	}

	t.Log("✓ NaN and infinite values handled gracefully")
}

// =================================================================================
// RESOURCE EXHAUSTION EDGE CASES
// =================================================================================

func TestMicrogliaEdgeResourceExhaustion(t *testing.T) {
	t.Log("=== TESTING RESOURCE EXHAUSTION SCENARIOS ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 3) // Very small limit

	// Fill to capacity
	for i := 0; i < 3; i++ {
		componentInfo := ComponentInfo{
			ID:       fmt.Sprintf("resource_test_%d", i),
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0},
			State:    StateActive,
		}
		err := microglia.CreateComponent(componentInfo)
		if err != nil {
			t.Fatalf("Failed to create component %d: %v", i, err)
		}
	}

	// Try to create beyond capacity
	overflowComponent := ComponentInfo{
		ID:       "overflow_neuron",
		Type:     ComponentNeuron,
		Position: Position3D{X: 100, Y: 0, Z: 0},
		State:    StateActive,
	}

	err := microglia.CreateComponent(overflowComponent)
	if err == nil {
		t.Error("Should fail to create component when at capacity")
	}

	// Test massive birth request queue
	for i := 0; i < 1000; i++ {
		request := ComponentBirthRequest{
			ComponentType: ComponentNeuron,
			Position:      Position3D{X: float64(i), Y: 0, Z: 0},
			Justification: fmt.Sprintf("Mass request %d", i),
			Priority:      PriorityLow,
			RequestedBy:   "stress_test",
		}
		microglia.RequestComponentBirth(request)
	}

	// Process requests - should handle gracefully
	created := microglia.ProcessBirthRequests()
	if len(created) > 0 {
		t.Logf("INFO: Created %d components despite being at capacity", len(created))
	}

	t.Log("✓ Resource exhaustion handled gracefully")
}

func TestMicrogliaEdgeMemoryStress(t *testing.T) {
	t.Log("=== TESTING MEMORY STRESS CONDITIONS ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 10000) // Large capacity

	// Create many health records
	for i := 0; i < 5000; i++ {
		componentID := fmt.Sprintf("stress_neuron_%d", i)
		microglia.UpdateComponentHealth(componentID, 0.5, 5)
	}

	// Create many pruning targets
	for i := 0; i < 5000; i++ {
		connectionID := fmt.Sprintf("stress_synapse_%d", i)
		microglia.MarkForPruning(connectionID, "src", "dst", 0.1)
	}

	// Test that system remains stable
	stats := microglia.GetMaintenanceStats()
	if stats.HealthChecks != 5000 {
		t.Errorf("Expected 5000 health checks, got %d", stats.HealthChecks)
	}

	candidates := microglia.GetPruningCandidates()
	if len(candidates) != 5000 {
		t.Errorf("Expected 5000 pruning candidates, got %d", len(candidates))
	}

	t.Log("✓ Memory stress conditions handled correctly")
}

// =================================================================================
// NETWORK TOPOLOGY EDGE CASES
// =================================================================================

func TestMicrogliaEdgeEmptyNetwork(t *testing.T) {
	t.Log("=== TESTING EMPTY NETWORK EDGE CASES ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Test operations on empty network
	_, exists := microglia.GetComponentHealth("nonexistent")
	if exists {
		t.Error("Should not find health for nonexistent component")
	}

	// Test pruning with no components
	microglia.MarkForPruning("phantom_synapse", "ghost_src", "ghost_dst", 0.5)
	candidates := microglia.GetPruningCandidates()
	if len(candidates) != 1 {
		t.Error("Should still track pruning candidates even for nonexistent components")
	}

	// Test patrol on empty territory
	territory := Territory{
		Center: Position3D{X: 0, Y: 0, Z: 0},
		Radius: 100.0,
	}
	microglia.EstablishPatrolRoute("ghost_patrol", territory, 100*time.Millisecond)
	report := microglia.ExecutePatrol("ghost_patrol")

	if report.ComponentsChecked != 0 {
		t.Errorf("Should check 0 components in empty territory, got %d", report.ComponentsChecked)
	}

	t.Log("✓ Empty network edge cases handled correctly")
}

func TestMicrogliaEdgeIsolatedComponents(t *testing.T) {
	t.Log("=== TESTING ISOLATED COMPONENT SCENARIOS ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Create isolated components (no connections)
	for i := 0; i < 10; i++ {
		componentInfo := ComponentInfo{
			ID:       fmt.Sprintf("isolated_%d", i),
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i * 1000), Y: 0, Z: 0}, // Far apart
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)
	}

	// Update health - should detect isolation
	for i := 0; i < 10; i++ {
		microglia.UpdateComponentHealth(fmt.Sprintf("isolated_%d", i), 0.8, 0) // No connections
	}

	// Verify isolation detection
	isolatedCount := 0
	for i := 0; i < 10; i++ {
		health, exists := microglia.GetComponentHealth(fmt.Sprintf("isolated_%d", i))
		if exists {
			for _, issue := range health.Issues {
				if issue == "isolated_component" {
					isolatedCount++
					break
				}
			}
		}
	}

	if isolatedCount != 10 {
		t.Errorf("Expected 10 isolated components, detected %d", isolatedCount)
	}

	// Test pruning calculations on isolated components
	microglia.MarkForPruning("isolated_synapse", "isolated_0", "isolated_1", 0.5)
	candidates := microglia.GetPruningCandidates()

	for _, candidate := range candidates {
		if candidate.ConnectionID == "isolated_synapse" {
			// Should handle isolated components gracefully
			if math.IsNaN(candidate.PruningScore) {
				t.Error("Pruning score should not be NaN for isolated components")
			}
		}
	}

	t.Log("✓ Isolated component scenarios handled correctly")
}

// =================================================================================
// NEW FUNCTION EDGE CASES
// =================================================================================

func TestMicrogliaEdgeRedundancyCalculation(t *testing.T) {
	t.Log("=== TESTING REDUNDANCY CALCULATION EDGE CASES ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Test with nonexistent components
	redundancy := microglia.calculateRedundancyScore("ghost_source", "ghost_target")
	// FIXED: Implementation correctly returns 0.5 as default moderate redundancy for nonexistent components
	if redundancy != 0.5 {
		t.Errorf("Expected 0.5 redundancy for nonexistent components (moderate default), got %.3f", redundancy)
	}

	// Create components with extreme positions
	extremeComponents := []ComponentInfo{
		{ID: "extreme_near", Type: ComponentNeuron, Position: Position3D{X: 0, Y: 0, Z: 0}, State: StateActive},
		{ID: "extreme_far", Type: ComponentNeuron, Position: Position3D{X: 1e6, Y: 1e6, Z: 1e6}, State: StateActive},
		{ID: "extreme_nan", Type: ComponentNeuron, Position: Position3D{X: math.NaN(), Y: 0, Z: 0}, State: StateActive},
	}

	for _, comp := range extremeComponents {
		microglia.CreateComponent(comp)
	}

	// Test redundancy with extreme distances
	redundancy = microglia.calculateRedundancyScore("extreme_near", "extreme_far")
	if math.IsNaN(redundancy) || redundancy < 0 || redundancy > 1 {
		t.Errorf("Redundancy score should be valid 0-1 range, got %.3f", redundancy)
	}

	// Test with NaN positions
	redundancy = microglia.calculateRedundancyScore("extreme_near", "extreme_nan")
	if math.IsNaN(redundancy) {
		t.Error("Redundancy calculation should handle NaN positions gracefully")
	}

	// Test spatial redundancy edge cases
	spatialRedundancy := microglia.calculateSpatialRedundancy(
		Position3D{X: 0, Y: 0, Z: 0},
		Position3D{X: math.Inf(1), Y: 0, Z: 0},
	)
	if math.IsNaN(spatialRedundancy) || math.IsInf(spatialRedundancy, 0) {
		t.Error("Spatial redundancy should handle infinite distances gracefully")
	}

	t.Log("✓ Redundancy calculation edge cases handled correctly")
}

func TestMicrogliaEdgeMetabolicCostCalculation(t *testing.T) {
	t.Log("=== TESTING METABOLIC COST CALCULATION EDGE CASES ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Test with nonexistent components
	cost := microglia.calculateMetabolicCost("ghost_src", "ghost_dst", 0.5)
	// FIXED: Implementation correctly returns 0.5 as default moderate cost for nonexistent components
	if cost != 0.5 {
		t.Errorf("Expected 0.5 cost for nonexistent components (moderate default), got %.3f", cost)
	}

	// Create components for testing
	microglia.CreateComponent(ComponentInfo{
		ID: "cost_test_a", Type: ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0}, State: StateActive,
	})
	microglia.CreateComponent(ComponentInfo{
		ID: "cost_test_b", Type: ComponentNeuron,
		Position: Position3D{X: math.Inf(1), Y: 0, Z: 0}, State: StateActive,
	})

	// Test with infinite distance
	cost = microglia.calculateMetabolicCost("cost_test_a", "cost_test_b", 0.5)
	if math.IsNaN(cost) || math.IsInf(cost, 0) || cost < 0 || cost > 1 {
		t.Errorf("Metabolic cost should be valid 0-1 range, got %.3f", cost)
	}

	// Test with extreme activity levels
	extremeActivities := []float64{-1.0, 0.0, math.NaN(), math.Inf(1), 1e6}
	for _, activity := range extremeActivities {
		cost = microglia.calculateMetabolicCost("cost_test_a", "cost_test_a", activity)
		if math.IsNaN(cost) || math.IsInf(cost, 0) {
			t.Errorf("Metabolic cost should handle extreme activity %.3f gracefully", activity)
		}
		if cost < 0 || cost > 1 {
			t.Errorf("Metabolic cost should be 0-1 range for activity %.3f, got %.3f", activity, cost)
		}
	}

	t.Log("✓ Metabolic cost calculation edge cases handled correctly")
}

func TestMicrogliaEdgePruningScoreIntegration(t *testing.T) {
	t.Log("=== TESTING INTEGRATED PRUNING SCORE EDGE CASES ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Test with completely invalid inputs
	microglia.MarkForPruning("edge_synapse", "", "", math.NaN())
	candidates := microglia.GetPruningCandidates()

	var edgeCandidate *PruningInfo
	for _, candidate := range candidates {
		if candidate.ConnectionID == "edge_synapse" {
			edgeCandidate = &candidate
			break
		}
	}

	if edgeCandidate == nil {
		t.Fatal("Should create pruning candidate even with invalid inputs")
	}

	// Verify score is valid despite invalid inputs
	if math.IsNaN(edgeCandidate.PruningScore) || math.IsInf(edgeCandidate.PruningScore, 0) {
		t.Error("Pruning score should be valid despite invalid inputs")
	}

	if edgeCandidate.PruningScore < 0 || edgeCandidate.PruningScore > 1 {
		t.Errorf("Pruning score should be 0-1 range, got %.3f", edgeCandidate.PruningScore)
	}

	// Test with massive numbers of connections
	microglia.CreateComponent(ComponentInfo{
		ID: "hub_neuron", Type: ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0}, State: StateActive,
	})

	// Simulate health with extreme connection count
	microglia.UpdateComponentHealth("hub_neuron", 0.5, 1000000)

	microglia.MarkForPruning("hub_synapse", "hub_neuron", "hub_neuron", 0.5)
	candidates = microglia.GetPruningCandidates()

	for _, candidate := range candidates {
		if candidate.ConnectionID == "hub_synapse" {
			if math.IsNaN(candidate.PruningScore) || candidate.PruningScore < 0 || candidate.PruningScore > 1 {
				t.Errorf("Hub pruning score should be valid, got %.3f", candidate.PruningScore)
			}
		}
	}

	t.Log("✓ Integrated pruning score edge cases handled correctly")
}

// =================================================================================
// CONCURRENT ACCESS STRESS TESTS
// =================================================================================

func TestMicrogliaEdgeConcurrentStress(t *testing.T) {
	t.Log("=== TESTING CONCURRENT ACCESS STRESS ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent component creation/removal
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func(id int) {
			defer wg.Done()
			componentInfo := ComponentInfo{
				ID:       fmt.Sprintf("stress_neuron_%d", id),
				Type:     ComponentNeuron,
				Position: Position3D{X: float64(id), Y: 0, Z: 0},
				State:    StateActive,
			}
			if err := microglia.CreateComponent(componentInfo); err != nil {
				errors <- err
			}
		}(i)

		go func(id int) {
			defer wg.Done()
			// Try to remove component (might not exist yet)
			microglia.RemoveComponent(fmt.Sprintf("stress_neuron_%d", id))
		}(i)
	}

	// Concurrent health updates
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			microglia.UpdateComponentHealth(fmt.Sprintf("stress_neuron_%d", id%50), 0.5, 5)
		}(i)
	}

	// Concurrent pruning operations
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			microglia.MarkForPruning(fmt.Sprintf("stress_synapse_%d", id), "src", "dst", 0.1)
		}(i)
	}

	// Concurrent configuration updates
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			config := GetDefaultMicrogliaConfig()
			microglia.UpdateConfig(config)
		}()
	}

	wg.Wait()
	close(errors)

	// Check for deadlocks or race conditions
	errorCount := 0
	for err := range errors {
		t.Logf("Concurrent operation error: %v", err)
		errorCount++
	}

	// Some errors are expected due to rapid creation/removal
	if errorCount > 25 { // Allow some tolerance
		t.Errorf("Too many concurrent errors: %d", errorCount)
	}

	// Verify system is still functional
	stats := microglia.GetMaintenanceStats()
	if stats.HealthChecks == 0 {
		t.Error("No health checks recorded - system may be deadlocked")
	}

	t.Log("✓ Concurrent access stress handled correctly")
}

// =================================================================================
// PERFORMANCE BOUNDARY TESTS
// =================================================================================

func TestMicrogliaEdgePerformanceBoundaries(t *testing.T) {
	t.Log("=== TESTING PERFORMANCE BOUNDARY CONDITIONS ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 100000) // Large capacity

	// Test large-scale operations with timing
	startTime := time.Now()

	// Create many components rapidly
	for i := 0; i < 1000; i++ {
		componentInfo := ComponentInfo{
			ID:       fmt.Sprintf("perf_neuron_%d", i),
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0},
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)
	}

	creationTime := time.Since(startTime)
	if creationTime > 5*time.Second {
		t.Errorf("Component creation too slow: %v", creationTime)
	}

	// Test health updates at scale
	startTime = time.Now()
	for i := 0; i < 1000; i++ {
		microglia.UpdateComponentHealth(fmt.Sprintf("perf_neuron_%d", i), 0.5, 10)
	}

	healthUpdateTime := time.Since(startTime)
	if healthUpdateTime > 2*time.Second {
		t.Errorf("Health updates too slow: %v", healthUpdateTime)
	}

	// Test pruning candidate generation at scale
	startTime = time.Now()
	for i := 0; i < 1000; i++ {
		microglia.MarkForPruning(fmt.Sprintf("perf_synapse_%d", i),
			fmt.Sprintf("perf_neuron_%d", i%100),
			fmt.Sprintf("perf_neuron_%d", (i+1)%100), 0.3)
	}

	candidates := microglia.GetPruningCandidates()
	pruningTime := time.Since(startTime)

	if pruningTime > 3*time.Second {
		t.Errorf("Pruning operations too slow: %v", pruningTime)
	}

	if len(candidates) != 1000 {
		t.Errorf("Expected 1000 pruning candidates, got %d", len(candidates))
	}

	t.Logf("Performance: Creation=%v, Health=%v, Pruning=%v",
		creationTime, healthUpdateTime, pruningTime)
	t.Log("✓ Performance boundaries within acceptable limits")
}

// =================================================================================
// ERROR RECOVERY AND SELF-HEALING
// =================================================================================

func TestMicrogliaEdgeErrorRecovery(t *testing.T) {
	t.Log("=== TESTING ERROR RECOVERY AND SELF-HEALING ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Create some components
	for i := 0; i < 10; i++ {
		componentInfo := ComponentInfo{
			ID:       fmt.Sprintf("recovery_neuron_%d", i),
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0},
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)
	}

	// Simulate corrupt health data by directly accessing internals
	// Note: This is normally not possible due to encapsulation
	microglia.UpdateComponentHealth("recovery_neuron_0", math.NaN(), -1000)

	// System should still function
	health, exists := microglia.GetComponentHealth("recovery_neuron_0")
	if exists {
		if math.IsNaN(health.HealthScore) {
			t.Error("System should recover from NaN health scores")
		}
	}

	// Test recovery from invalid pruning data
	microglia.MarkForPruning("corrupt_synapse", "recovery_neuron_0", "nonexistent", math.Inf(1))

	// Should still generate valid candidates
	candidates := microglia.GetPruningCandidates()
	for _, candidate := range candidates {
		if candidate.ConnectionID == "corrupt_synapse" {
			if math.IsNaN(candidate.PruningScore) || math.IsInf(candidate.PruningScore, 0) {
				t.Error("System should recover from infinite activity levels")
			}
		}
	}

	// Test statistics consistency after errors
	stats := microglia.GetMaintenanceStats()
	if math.IsNaN(stats.AverageHealthScore) {
		t.Error("Statistics should remain valid despite data corruption")
	}

	t.Log("✓ Error recovery and self-healing working correctly")
}

// =================================================================================
// DATA CORRUPTION AND MALFORMED INPUT TESTS
// =================================================================================

func TestMicrogliaEdgeDataCorruption(t *testing.T) {
	t.Log("=== TESTING DATA CORRUPTION SCENARIOS ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Test with malformed component IDs
	malformedIDs := []string{
		"",             // Empty string
		"\x00\x01\x02", // Binary data
		"extremely_long_id_that_exceeds_reasonable_length_limits_and_might_cause_buffer_overflows_in_poorly_designed_systems",
		"unicode_test_🧠🔬⚡",              // Unicode characters
		"injection'attempt--DROP TABLE", // SQL injection style
		"\n\r\t",                        // Whitespace characters
	}

	for i, badID := range malformedIDs {
		componentInfo := ComponentInfo{
			ID:       badID,
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0},
			State:    StateActive,
		}

		err := microglia.CreateComponent(componentInfo)
		if err != nil {
			t.Logf("INFO: Rejected malformed ID '%s': %v", badID, err)
		} else {
			// If accepted, ensure it doesn't break the system
			microglia.UpdateComponentHealth(badID, 0.5, 5)
			health, exists := microglia.GetComponentHealth(badID)
			if exists && (math.IsNaN(health.HealthScore) || health.HealthScore < 0) {
				t.Errorf("Malformed ID caused invalid health state: %s", badID)
			}
		}
	}

	// Test with malformed metadata
	corruptMetadata := map[string]interface{}{
		"circular_ref": nil, // Will be set to self-reference
		"deep_nesting": map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": "deep",
				},
			},
		},
		"large_data": make([]byte, 1000000), // 1MB of data
		"nil_value":  nil,
		"func_ptr":   func() {}, // Function pointer
	}
	corruptMetadata["circular_ref"] = corruptMetadata

	corruptRequest := ComponentBirthRequest{
		ComponentType: ComponentNeuron,
		Position:      Position3D{X: 0, Y: 0, Z: 0},
		Justification: "Corruption test",
		Priority:      PriorityLow,
		RequestedBy:   "corruption_test",
		Metadata:      corruptMetadata,
	}

	err := microglia.RequestComponentBirth(corruptRequest)
	if err != nil {
		t.Logf("INFO: Rejected corrupt metadata: %v", err)
	} else {
		// Should handle gracefully
		created := microglia.ProcessBirthRequests()
		t.Logf("INFO: System handled corrupt metadata, created %d components", len(created))
	}

	t.Log("✓ Data corruption scenarios handled gracefully")
}

func TestMicrogliaEdgeTimeManipulation(t *testing.T) {
	t.Log("=== TESTING TIME MANIPULATION EDGE CASES ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Create component with normal time
	componentInfo := ComponentInfo{
		ID:       "time_test_neuron",
		Type:     ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0},
		State:    StateActive,
	}
	microglia.CreateComponent(componentInfo)

	// Update health normally
	microglia.UpdateComponentHealth("time_test_neuron", 0.5, 5)

	// Test with very old timestamps (simulate clock going backwards)
	// Note: In real implementation, we can't easily manipulate internal timestamps
	// but we can test the system's response to unusual timing patterns

	// Mark for pruning with immediate execution attempt
	microglia.MarkForPruning("time_synapse", "time_test_neuron", "time_test_neuron", 0.1)

	// Try to execute pruning immediately (should fail due to age requirement)
	pruned := microglia.ExecutePruning()
	if len(pruned) > 0 {
		t.Error("Should not prune immediately marked connections")
	}

	// Test rapid health updates (stress timing logic)
	for i := 0; i < 1000; i++ {
		microglia.UpdateComponentHealth("time_test_neuron", 0.5, 5)
	}

	health, _ := microglia.GetComponentHealth("time_test_neuron")
	if health.PatrolCount != 1001 { // 1 initial + 1000 updates
		t.Errorf("Expected 1001 patrol count, got %d", health.PatrolCount)
	}

	// Test patrol timing with very short intervals
	territory := Territory{
		Center: Position3D{X: 0, Y: 0, Z: 0},
		Radius: 50.0,
	}
	microglia.EstablishPatrolRoute("rapid_patrol", territory, 1*time.Nanosecond)

	// Execute multiple rapid patrols
	for i := 0; i < 10; i++ {
		report := microglia.ExecutePatrol("rapid_patrol")
		if report.PatrolTime.IsZero() {
			t.Error("Patrol time should always be set")
		}
	}

	t.Log("✓ Time manipulation edge cases handled correctly")
}

// =================================================================================
// SPATIAL CALCULATION EDGE CASES
// =================================================================================

func TestMicrogliaEdgeSpatialCalculations(t *testing.T) {
	t.Log("=== TESTING SPATIAL CALCULATION EDGE CASES ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Test with extreme spatial coordinates
	extremePositions := []Position3D{
		{X: 0, Y: 0, Z: 0},                               // Origin
		{X: math.MaxFloat64, Y: 0, Z: 0},                 // Maximum value
		{X: -math.MaxFloat64, Y: 0, Z: 0},                // Minimum value
		{X: math.SmallestNonzeroFloat64, Y: 0, Z: 0},     // Smallest positive
		{X: math.Inf(1), Y: math.Inf(-1), Z: math.NaN()}, // Invalid values
		{X: 1e100, Y: 1e100, Z: 1e100},                   // Very large
		{X: 1e-100, Y: 1e-100, Z: 1e-100},                // Very small
	}

	for i, pos := range extremePositions {
		componentID := fmt.Sprintf("spatial_test_%d", i)
		componentInfo := ComponentInfo{
			ID:       componentID,
			Type:     ComponentNeuron,
			Position: pos,
			State:    StateActive,
		}

		err := microglia.CreateComponent(componentInfo)
		if err != nil {
			t.Logf("INFO: Rejected extreme position %v: %v", pos, err)
			continue
		}

		// Test spatial redundancy calculation
		for j, otherPos := range extremePositions {
			if i != j {
				redundancy := microglia.calculateSpatialRedundancy(pos, otherPos)
				if math.IsNaN(redundancy) {
					t.Errorf("Spatial redundancy should not be NaN for positions %v and %v", pos, otherPos)
				}
				if redundancy < 0 || redundancy > 1 {
					t.Errorf("Spatial redundancy should be 0-1, got %.3f for positions %v and %v",
						redundancy, pos, otherPos)
				}
			}
		}
	}

	// Test patrol territory with extreme values
	extremeTerritories := []Territory{
		{Center: Position3D{X: 0, Y: 0, Z: 0}, Radius: 0},           // Zero radius
		{Center: Position3D{X: 0, Y: 0, Z: 0}, Radius: math.Inf(1)}, // Infinite radius
		{Center: Position3D{X: math.NaN(), Y: 0, Z: 0}, Radius: 50}, // NaN center
		{Center: Position3D{X: 0, Y: 0, Z: 0}, Radius: -50},         // Negative radius
	}

	for i, territory := range extremeTerritories {
		patrolID := fmt.Sprintf("extreme_patrol_%d", i)
		microglia.EstablishPatrolRoute(patrolID, territory, 100*time.Millisecond)

		// Should not crash when executing patrol
		report := microglia.ExecutePatrol(patrolID)
		if report.MicrogliaID != patrolID {
			t.Errorf("Patrol ID mismatch: expected %s, got %s", patrolID, report.MicrogliaID)
		}
	}

	t.Log("✓ Spatial calculation edge cases handled correctly")
}

// =================================================================================
// CONFIGURATION BOUNDARY TESTS
// =================================================================================

func TestMicrogliaEdgeConfigurationBoundaries(t *testing.T) {
	t.Log("=== TESTING CONFIGURATION BOUNDARY CONDITIONS ===")

	astrocyteNetwork := NewAstrocyteNetwork()

	// Test configuration with all zeros
	zeroConfig := MicrogliaConfig{
		HealthThresholds: HealthScoringConfig{},
		PruningSettings:  PruningConfig{},
		PatrolSettings:   PatrolConfig{},
		ResourceLimits:   ResourceConfig{},
	}

	microglia := NewMicrogliaWithConfig(astrocyteNetwork, zeroConfig)

	// Should still function with zero config
	componentInfo := ComponentInfo{
		ID:       "zero_config_test",
		Type:     ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0},
		State:    StateActive,
	}

	err := microglia.CreateComponent(componentInfo)
	if err == nil {
		microglia.UpdateComponentHealth("zero_config_test", 0.5, 5)
		health, exists := microglia.GetComponentHealth("zero_config_test")
		if exists && (math.IsNaN(health.HealthScore) || health.HealthScore < 0) {
			t.Error("Zero config should not produce invalid health scores")
		}
	}

	// Test configuration with maximum values
	maxConfig := MicrogliaConfig{
		HealthThresholds: HealthScoringConfig{
			CriticalActivityThreshold: math.MaxFloat64,
			VeryLowActivityThreshold:  math.MaxFloat64,
			LowActivityThreshold:      math.MaxFloat64,
			ModerateActivityThreshold: math.MaxFloat64,
			CriticalActivityPenalty:   math.MaxFloat64,
			LowActivityPenalty:        math.MaxFloat64,
			ModerateActivityPenalty:   math.MaxFloat64,
		},
		PruningSettings: PruningConfig{
			ActivityWeight:   math.MaxFloat64,
			RedundancyWeight: math.MaxFloat64,
			MetabolicWeight:  math.MaxFloat64,
			MaxPruningScore:  math.MaxFloat64,
		},
		ResourceLimits: ResourceConfig{
			MaxComponents: math.MaxInt32,
		},
	}

	maxMicroglia := NewMicrogliaWithConfig(astrocyteNetwork, maxConfig)

	// Test that extreme config doesn't break basic operations
	maxMicroglia.CreateComponent(componentInfo)
	maxMicroglia.UpdateComponentHealth("zero_config_test", 0.5, 5)
	maxMicroglia.MarkForPruning("max_synapse", "zero_config_test", "zero_config_test", 0.5)

	candidates := maxMicroglia.GetPruningCandidates()
	for _, candidate := range candidates {
		if math.IsNaN(candidate.PruningScore) || math.IsInf(candidate.PruningScore, 0) {
			t.Error("Max config should not produce NaN or infinite pruning scores")
		}
	}

	t.Log("✓ Configuration boundary conditions handled correctly")
}

// =================================================================================
// SYSTEM INTEGRATION EDGE CASES
// =================================================================================

func TestMicrogliaEdgeSystemIntegration(t *testing.T) {
	t.Log("=== TESTING SYSTEM INTEGRATION EDGE CASES ===")

	// Test with nil astrocyte network (should be caught in constructor)
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("INFO: Correctly panicked with nil astrocyte network: %v", r)
			}
		}()

		// This might panic, which is acceptable behavior
		NewMicroglia(nil, 1000)
	}()

	// Test normal operation
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Test cross-system consistency
	componentInfo := ComponentInfo{
		ID:       "integration_test",
		Type:     ComponentNeuron,
		Position: Position3D{X: 50, Y: 50, Z: 50},
		State:    StateActive,
	}

	err := microglia.CreateComponent(componentInfo)
	if err != nil {
		t.Fatalf("Failed to create component: %v", err)
	}

	// Verify component exists in astrocyte network
	info, exists := astrocyteNetwork.Get("integration_test")
	if !exists {
		t.Error("Component should exist in astrocyte network")
	}

	if info.Position.X != 50 {
		t.Errorf("Position mismatch: expected X=50, got X=%.1f", info.Position.X)
	}

	// Test removal consistency
	err = microglia.RemoveComponent("integration_test")
	if err != nil {
		t.Errorf("Failed to remove component: %v", err)
	}

	// Should be removed from astrocyte network
	_, exists = astrocyteNetwork.Get("integration_test")
	if exists {
		t.Error("Component should be removed from astrocyte network")
	}

	// Should be removed from health monitoring
	_, exists = microglia.GetComponentHealth("integration_test")
	if exists {
		t.Error("Component health should be cleaned up")
	}

	// Test statistics consistency
	stats := microglia.GetMaintenanceStats()
	if stats.ComponentsCreated != 1 || stats.ComponentsRemoved != 1 {
		t.Errorf("Statistics inconsistent: created=%d, removed=%d",
			stats.ComponentsCreated, stats.ComponentsRemoved)
	}

	t.Log("✓ System integration edge cases handled correctly")
}

// =================================================================================
// MEMORY LEAK AND RESOURCE CLEANUP TESTS
// =================================================================================

func TestMicrogliaEdgeResourceCleanup(t *testing.T) {
	t.Log("=== TESTING RESOURCE CLEANUP AND MEMORY MANAGEMENT ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Create and remove many components to test cleanup
	for cycle := 0; cycle < 10; cycle++ {
		// Create components
		for i := 0; i < 100; i++ {
			componentID := fmt.Sprintf("cleanup_test_%d_%d", cycle, i)
			componentInfo := ComponentInfo{
				ID:       componentID,
				Type:     ComponentNeuron,
				Position: Position3D{X: float64(i), Y: float64(cycle), Z: 0},
				State:    StateActive,
			}
			microglia.CreateComponent(componentInfo)
			microglia.UpdateComponentHealth(componentID, 0.5, 5)
			microglia.MarkForPruning(fmt.Sprintf("synapse_%s", componentID), componentID, componentID, 0.1)
		}

		// Remove all components
		for i := 0; i < 100; i++ {
			componentID := fmt.Sprintf("cleanup_test_%d_%d", cycle, i)
			microglia.RemoveComponent(componentID)
		}

		// Verify cleanup
		candidates := microglia.GetPruningCandidates()
		healthRecords := 0
		for i := 0; i < 100; i++ {
			componentID := fmt.Sprintf("cleanup_test_%d_%d", cycle, i)
			_, exists := microglia.GetComponentHealth(componentID)
			if exists {
				healthRecords++
			}
		}

		if healthRecords > 0 {
			t.Errorf("Cycle %d: %d health records not cleaned up", cycle, healthRecords)
		}

		// Pruning targets for removed components should be cleaned up
		orphanedPruning := 0
		for _, candidate := range candidates {
			if candidate.SourceID == fmt.Sprintf("cleanup_test_%d_0", cycle) {
				orphanedPruning++
			}
		}

		if orphanedPruning > 0 {
			t.Errorf("Cycle %d: %d orphaned pruning targets", cycle, orphanedPruning)
		}
	}

	// Final verification
	stats := microglia.GetMaintenanceStats()
	if stats.ComponentsCreated != 1000 || stats.ComponentsRemoved != 1000 {
		t.Errorf("Final stats incorrect: created=%d, removed=%d",
			stats.ComponentsCreated, stats.ComponentsRemoved)
	}

	t.Log("✓ Resource cleanup and memory management working correctly")
}

// =================================================================================
// EDGE CASE SUMMARY AND VALIDATION
// =================================================================================

func TestMicrogliaEdgeSummaryValidation(t *testing.T) {
	t.Log("=== EDGE CASE TESTING SUMMARY AND VALIDATION ===")

	// Create a system and run through various edge cases quickly
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Test a combination of edge cases
	edgeCases := []struct {
		name string
		test func() error
	}{
		{"Empty operations", func() error {
			microglia.ExecutePruning()
			microglia.ProcessBirthRequests()
			return nil
		}},
		{"Invalid health update", func() error {
			microglia.UpdateComponentHealth("nonexistent", math.NaN(), -1)
			return nil
		}},
		{"Invalid pruning", func() error {
			microglia.MarkForPruning("", "", "", math.Inf(1))
			return nil
		}},
		{"Extreme config update", func() error {
			extremeConfig := GetDefaultMicrogliaConfig()
			extremeConfig.PruningSettings.ActivityWeight = math.MaxFloat64
			microglia.UpdateConfig(extremeConfig)
			return nil
		}},
		{"Rapid operations", func() error {
			for i := 0; i < 100; i++ {
				microglia.UpdateComponentHealth(fmt.Sprintf("rapid_%d", i), 0.5, 5)
			}
			return nil
		}},
	}

	failureCount := 0
	for _, edgeCase := range edgeCases {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Edge case '%s' caused panic: %v", edgeCase.name, r)
					failureCount++
				}
			}()

			if err := edgeCase.test(); err != nil {
				t.Errorf("Edge case '%s' failed: %v", edgeCase.name, err)
				failureCount++
			}
		}()
	}

	if failureCount == 0 {
		t.Log("✓ All edge cases handled gracefully")
	} else {
		t.Errorf("Failed %d out of %d edge cases", failureCount, len(edgeCases))
	}

	// Final system state validation
	stats := microglia.GetMaintenanceStats()
	if math.IsNaN(stats.AverageHealthScore) || math.IsInf(stats.AverageHealthScore, 0) {
		t.Error("System statistics corrupted by edge cases")
	}

	// Ensure system is still functional after all edge cases
	finalComponent := ComponentInfo{
		ID:       "final_validation",
		Type:     ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0},
		State:    StateActive,
	}

	if err := microglia.CreateComponent(finalComponent); err != nil {
		t.Errorf("System not functional after edge case testing: %v", err)
	}

	t.Log("✓ System remains functional after comprehensive edge case testing")
	t.Logf("Final statistics: Created=%d, Removed=%d, Health Checks=%d",
		stats.ComponentsCreated, stats.ComponentsRemoved, stats.HealthChecks)
}
