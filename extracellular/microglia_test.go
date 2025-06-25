/*
=================================================================================
MICROGLIA - BASIC TECHNICAL TESTS
=================================================================================

Basic technical tests for the microglia lifecycle management system.
Focuses on core functionality, API correctness, and basic integration.

Test Categories:
1. Constructor and Configuration
2. Component Lifecycle (Create/Remove)
3. Health Monitoring (Basic functionality)
4. Pruning System (Basic marking/execution)
5. Birth Request Processing
6. Patrol System (Basic routes/execution)
7. Statistics Tracking
8. Thread Safety (Basic concurrent access)

For advanced testing see:
- microglia_biology_test.go - Biological realism and parameter validation
- microglia_performance_test.go - Performance, stress, and scalability
- microglia_edge_test.go - Edge cases, error conditions, and boundary testing
=================================================================================
*/

package extracellular

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// =================================================================================
// CONSTRUCTOR AND CONFIGURATION TESTS
// =================================================================================

func TestMicrogliaConstructors(t *testing.T) {
	astrocyteNetwork := NewAstrocyteNetwork()

	// Test compatibility constructor
	mg1 := NewMicroglia(astrocyteNetwork, 500)
	config1 := mg1.GetConfig()
	if config1.ResourceLimits.MaxComponents != 500 {
		t.Errorf("Expected max components 500, got %d", config1.ResourceLimits.MaxComponents)
	}

	// Test default config constructor with zero max components
	mg2 := NewMicroglia(astrocyteNetwork, 0)
	config2 := mg2.GetConfig()
	if config2.ResourceLimits.MaxComponents != 1000 { // Should use default
		t.Errorf("Expected default max components 1000, got %d", config2.ResourceLimits.MaxComponents)
	}

	// Test full config constructor
	customConfig := GetDefaultMicrogliaConfig()
	customConfig.ResourceLimits.MaxComponents = 750
	mg3 := NewMicrogliaWithConfig(astrocyteNetwork, customConfig)
	config3 := mg3.GetConfig()
	if config3.ResourceLimits.MaxComponents != 750 {
		t.Errorf("Expected custom max components 750, got %d", config3.ResourceLimits.MaxComponents)
	}

	t.Log("✓ Constructor variants working correctly")
}

func TestMicrogliaConfigPresets(t *testing.T) {
	// Test preset configurations exist and are different
	defaultConfig := GetDefaultMicrogliaConfig()
	conservativeConfig := GetConservativeMicrogliaConfig()
	aggressiveConfig := GetAggressiveMicrogliaConfig()

	// Verify they're actually different
	if defaultConfig.PruningSettings.AgeThreshold == conservativeConfig.PruningSettings.AgeThreshold {
		t.Error("Conservative config should have different pruning age threshold than default")
	}

	if defaultConfig.HealthThresholds.CriticalActivityThreshold == aggressiveConfig.HealthThresholds.CriticalActivityThreshold {
		t.Error("Aggressive config should have different activity threshold than default")
	}

	// Verify they have reasonable values
	if defaultConfig.ResourceLimits.MaxComponents <= 0 {
		t.Error("Default config should have positive max components")
	}

	if conservativeConfig.PruningSettings.AgeThreshold <= defaultConfig.PruningSettings.AgeThreshold {
		t.Error("Conservative config should have longer pruning age threshold")
	}

	if aggressiveConfig.PruningSettings.AgeThreshold >= defaultConfig.PruningSettings.AgeThreshold {
		t.Error("Aggressive config should have shorter pruning age threshold")
	}

	t.Log("✓ Configuration presets working correctly")
}

func TestMicrogliaConfigUpdate(t *testing.T) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Get initial config
	initialConfig := microglia.GetConfig()
	initialThreshold := initialConfig.HealthThresholds.CriticalActivityThreshold

	// Update config
	newConfig := GetAggressiveMicrogliaConfig()
	microglia.UpdateConfig(newConfig)

	// Verify config was updated
	updatedConfig := microglia.GetConfig()
	if updatedConfig.HealthThresholds.CriticalActivityThreshold == initialThreshold {
		t.Error("Config should have been updated")
	}

	t.Log("✓ Configuration update working correctly")
}

// =================================================================================
// COMPONENT LIFECYCLE TESTS
// =================================================================================

func TestMicrogliaComponentCreation(t *testing.T) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Test basic component creation
	componentInfo := ComponentInfo{
		ID:       "test_neuron_1",
		Type:     ComponentNeuron,
		Position: Position3D{X: 10, Y: 20, Z: 30},
		State:    StateActive,
	}

	err := microglia.CreateComponent(componentInfo)
	if err != nil {
		t.Fatalf("Failed to create component: %v", err)
	}

	// Verify component was registered in astrocyte network
	info, exists := astrocyteNetwork.Get("test_neuron_1")
	if !exists {
		t.Fatal("Component not found in astrocyte network")
	}

	if info.ID != "test_neuron_1" {
		t.Errorf("Expected ID test_neuron_1, got %s", info.ID)
	}

	// Verify health monitoring was initialized
	health, exists := microglia.GetComponentHealth("test_neuron_1")
	if !exists {
		t.Fatal("Health monitoring not initialized")
	}

	if health.HealthScore != 1.0 {
		t.Errorf("Expected initial health score 1.0, got %.3f", health.HealthScore)
	}

	// Verify statistics were updated
	stats := microglia.GetMaintenanceStats()
	if stats.ComponentsCreated != 1 {
		t.Errorf("Expected 1 component created, got %d", stats.ComponentsCreated)
	}

	t.Log("✓ Component creation working correctly")
}

func TestMicrogliaComponentRemoval(t *testing.T) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Create component first
	componentInfo := ComponentInfo{
		ID:       "test_neuron_remove",
		Type:     ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0},
		State:    StateActive,
	}
	microglia.CreateComponent(componentInfo)

	// Verify component exists
	_, exists := astrocyteNetwork.Get("test_neuron_remove")
	if !exists {
		t.Fatal("Component should exist before removal")
	}

	// Remove component
	err := microglia.RemoveComponent("test_neuron_remove")
	if err != nil {
		t.Fatalf("Failed to remove component: %v", err)
	}

	// Verify component was removed from astrocyte network
	_, exists = astrocyteNetwork.Get("test_neuron_remove")
	if exists {
		t.Error("Component should not exist after removal")
	}

	// Verify health monitoring was cleaned up
	_, exists = microglia.GetComponentHealth("test_neuron_remove")
	if exists {
		t.Error("Health monitoring should be cleaned up")
	}

	// Verify statistics were updated
	stats := microglia.GetMaintenanceStats()
	if stats.ComponentsRemoved != 1 {
		t.Errorf("Expected 1 component removed, got %d", stats.ComponentsRemoved)
	}

	t.Log("✓ Component removal working correctly")
}

// =================================================================================
// HEALTH MONITORING TESTS
// =================================================================================

func TestMicrogliaHealthMonitoring(t *testing.T) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Create test component
	componentInfo := ComponentInfo{
		ID:       "health_test_neuron",
		Type:     ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0},
		State:    StateActive,
	}
	microglia.CreateComponent(componentInfo)

	// Test health update with normal activity
	microglia.UpdateComponentHealth("health_test_neuron", 0.8, 10)

	health, exists := microglia.GetComponentHealth("health_test_neuron")
	if !exists {
		t.Fatal("Health record should exist")
	}

	if health.ActivityLevel != 0.8 {
		t.Errorf("Expected activity level 0.8, got %.3f", health.ActivityLevel)
	}

	if health.ConnectionCount != 10 {
		t.Errorf("Expected connection count 10, got %d", health.ConnectionCount)
	}

	if health.HealthScore <= 0 || health.HealthScore > 1 {
		t.Errorf("Health score should be between 0 and 1, got %.3f", health.HealthScore)
	}

	if health.PatrolCount != 1 {
		t.Errorf("Expected patrol count 1, got %d", health.PatrolCount)
	}

	t.Log("✓ Health monitoring working correctly")
}

func TestMicrogliaHealthScoreCalculation(t *testing.T) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Create test component
	componentInfo := ComponentInfo{
		ID:       "score_test_neuron",
		Type:     ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0},
		State:    StateActive,
	}
	microglia.CreateComponent(componentInfo)

	// Test different activity levels and their impact on health score
	testCases := []struct {
		activity    float64
		connections int
		description string
	}{
		{0.8, 10, "High activity, well connected"},
		{0.3, 5, "Moderate activity, moderate connections"},
		{0.05, 2, "Low activity, few connections"},
		{0.01, 1, "Very low activity, poorly connected"},
	}

	for _, tc := range testCases {
		microglia.UpdateComponentHealth("score_test_neuron", tc.activity, tc.connections)

		health, _ := microglia.GetComponentHealth("score_test_neuron")

		// Basic sanity checks
		if health.HealthScore < 0 || health.HealthScore > 1 {
			t.Errorf("%s: Health score should be 0-1, got %.3f", tc.description, health.HealthScore)
		}

		// Higher activity should generally mean higher score
		if tc.activity >= 0.5 && health.HealthScore < 0.5 {
			t.Errorf("%s: High activity should result in reasonable health score, got %.3f", tc.description, health.HealthScore)
		}

		t.Logf("%s: Activity=%.3f, Connections=%d, Health=%.3f",
			tc.description, tc.activity, tc.connections, health.HealthScore)
	}

	t.Log("✓ Health score calculation working correctly")
}

func TestMicrogliaHealthIssueDetection(t *testing.T) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Create test component
	componentInfo := ComponentInfo{
		ID:       "issue_test_neuron",
		Type:     ComponentNeuron,
		Position: Position3D{X: 0, Y: 0, Z: 0},
		State:    StateActive,
	}
	microglia.CreateComponent(componentInfo)

	// Test very low activity detection
	microglia.UpdateComponentHealth("issue_test_neuron", 0.02, 5)
	health, _ := microglia.GetComponentHealth("issue_test_neuron")

	if len(health.Issues) == 0 {
		t.Error("Should detect issues with very low activity")
	}

	foundActivityIssue := false
	for _, issue := range health.Issues {
		if issue == "very_low_activity" || issue == "critically_low_activity" {
			foundActivityIssue = true
			break
		}
	}
	if !foundActivityIssue {
		t.Error("Should detect low activity issues")
	}

	// Test isolation detection
	microglia.UpdateComponentHealth("issue_test_neuron", 0.5, 0)
	health, _ = microglia.GetComponentHealth("issue_test_neuron")

	foundIsolation := false
	for _, issue := range health.Issues {
		if issue == "isolated_component" {
			foundIsolation = true
			break
		}
	}
	if !foundIsolation {
		t.Error("Should detect isolated component")
	}

	t.Log("✓ Health issue detection working correctly")
}

// =================================================================================
// PRUNING SYSTEM TESTS
// =================================================================================

func TestMicrogliaPruningCandidates(t *testing.T) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Mark connections for pruning
	microglia.MarkForPruning("weak_synapse_1", "neuron_A", "neuron_B", 0.1)
	microglia.MarkForPruning("weak_synapse_2", "neuron_C", "neuron_D", 0.05)
	microglia.MarkForPruning("strong_synapse", "neuron_E", "neuron_F", 0.9)

	// Get pruning candidates
	candidates := microglia.GetPruningCandidates()

	if len(candidates) != 3 {
		t.Fatalf("Expected 3 pruning candidates, got %d", len(candidates))
	}

	// Verify all candidates have valid data
	for _, candidate := range candidates {
		if candidate.ConnectionID == "" {
			t.Error("Candidate should have valid connection ID")
		}
		if candidate.SourceID == "" || candidate.TargetID == "" {
			t.Error("Candidate should have valid source and target IDs")
		}
		if candidate.PruningScore < 0 || candidate.PruningScore > 1 {
			t.Errorf("Pruning score should be 0-1, got %.3f", candidate.PruningScore)
		}
		if candidate.MarkedAt.IsZero() {
			t.Error("Candidate should have valid marked timestamp")
		}
	}

	// Basic logic: weak synapses should have higher pruning scores than strong ones
	weakFound := false
	strongFound := false
	for _, candidate := range candidates {
		if candidate.ConnectionID == "weak_synapse_1" {
			weakFound = true
			if candidate.ActivityLevel != 0.1 {
				t.Errorf("Expected activity level 0.1, got %.3f", candidate.ActivityLevel)
			}
		}
		if candidate.ConnectionID == "strong_synapse" {
			strongFound = true
			if candidate.ActivityLevel != 0.9 {
				t.Errorf("Expected activity level 0.9, got %.3f", candidate.ActivityLevel)
			}
		}
	}

	if !weakFound || !strongFound {
		t.Error("Should find both weak and strong synapses in candidates")
	}

	t.Log("✓ Pruning candidate system working correctly")
}

func TestMicrogliaPruningExecution(t *testing.T) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Mark a connection for pruning
	microglia.MarkForPruning("test_synapse", "neuron_X", "neuron_Y", 0.01)

	// Execute pruning (should not prune immediately due to age requirement)
	prunedConnections := microglia.ExecutePruning()

	// Should not prune connections that are too young (24 hour default threshold)
	if len(prunedConnections) > 0 {
		t.Error("Should not prune connections immediately marked")
	}

	// Verify connection is still in pruning candidates
	candidates := microglia.GetPruningCandidates()
	if len(candidates) != 1 {
		t.Error("Connection should still be in pruning candidates")
	}

	// Verify statistics
	stats := microglia.GetMaintenanceStats()
	if stats.ConnectionsPruned != 0 {
		t.Error("No connections should have been pruned yet")
	}

	t.Log("✓ Pruning execution working correctly")
}

// =================================================================================
// BIRTH REQUEST SYSTEM TESTS
// =================================================================================

func TestMicrogliaBirthRequestBasics(t *testing.T) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Submit birth request
	birthRequest := ComponentBirthRequest{
		ComponentType: ComponentNeuron,
		Position:      Position3D{X: 100, Y: 100, Z: 100},
		Justification: "Test component creation",
		Priority:      PriorityHigh,
		RequestedBy:   "test_system",
		Metadata: map[string]interface{}{
			"test_data": "test_value",
		},
	}

	err := microglia.RequestComponentBirth(birthRequest)
	if err != nil {
		t.Fatalf("Failed to submit birth request: %v", err)
	}

	// Process birth requests
	createdComponents := microglia.ProcessBirthRequests()

	if len(createdComponents) != 1 {
		t.Fatalf("Expected 1 component to be created, got %d", len(createdComponents))
	}

	createdComponent := createdComponents[0]
	if createdComponent.Type != ComponentNeuron {
		t.Errorf("Expected neuron type, got %v", createdComponent.Type)
	}

	if createdComponent.Position.X != 100 {
		t.Errorf("Expected position X=100, got %.1f", createdComponent.Position.X)
	}

	// Verify component was actually created in astrocyte network
	_, exists := astrocyteNetwork.Get(createdComponent.ID)
	if !exists {
		t.Error("Created component should exist in astrocyte network")
	}

	// Verify statistics
	stats := microglia.GetMaintenanceStats()
	if stats.ComponentsCreated != 1 {
		t.Errorf("Expected 1 component created in stats, got %d", stats.ComponentsCreated)
	}

	t.Log("✓ Birth request system working correctly")
}

func TestMicrogliaResourceConstraints(t *testing.T) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 5) // Very low limit for testing

	// Create components up to limit
	for i := 0; i < 5; i++ {
		componentInfo := ComponentInfo{
			ID:       fmt.Sprintf("limit_neuron_%d", i),
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0},
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)
	}

	// Submit low priority request when at limit
	lowPriorityRequest := ComponentBirthRequest{
		ComponentType: ComponentNeuron,
		Position:      Position3D{X: 10, Y: 0, Z: 0},
		Justification: "Should be rejected",
		Priority:      PriorityLow,
		RequestedBy:   "test_system",
	}

	microglia.RequestComponentBirth(lowPriorityRequest)
	created := microglia.ProcessBirthRequests()
	if len(created) != 0 {
		t.Error("Low priority request should be rejected when at resource limit")
	}

	// High priority request should bypass limit (default config allows this)
	highPriorityRequest := ComponentBirthRequest{
		ComponentType: ComponentNeuron,
		Position:      Position3D{X: 11, Y: 0, Z: 0},
		Justification: "Should be approved",
		Priority:      PriorityHigh,
		RequestedBy:   "emergency_system",
	}

	microglia.RequestComponentBirth(highPriorityRequest)
	created = microglia.ProcessBirthRequests()
	if len(created) != 1 {
		t.Error("High priority request should bypass resource limits")
	}

	t.Log("✓ Resource constraint system working correctly")
}

// =================================================================================
// PATROL SYSTEM TESTS
// =================================================================================

func TestMicrogliaPatrolBasics(t *testing.T) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Establish patrol route
	territory := Territory{
		Center: Position3D{X: 50, Y: 50, Z: 50},
		Radius: 25.0,
	}

	microglia.EstablishPatrolRoute("microglia_1", territory, 100*time.Millisecond)

	// Create some components in the territory
	for i := 0; i < 3; i++ {
		componentInfo := ComponentInfo{
			ID:       fmt.Sprintf("patrol_neuron_%d", i),
			Type:     ComponentNeuron,
			Position: Position3D{X: 50 + float64(i), Y: 50, Z: 50},
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)
	}

	// Execute patrol
	report := microglia.ExecutePatrol("microglia_1")

	if report.MicrogliaID != "microglia_1" {
		t.Errorf("Expected microglia ID microglia_1, got %s", report.MicrogliaID)
	}

	if report.ComponentsChecked == 0 {
		t.Error("Should have checked some components during patrol")
	}

	if report.PatrolTime.IsZero() {
		t.Error("Patrol time should be set")
	}

	// Verify statistics were updated
	stats := microglia.GetMaintenanceStats()
	if stats.PatrolsCompleted != 1 {
		t.Errorf("Expected 1 patrol completed, got %d", stats.PatrolsCompleted)
	}

	if stats.HealthChecks == 0 {
		t.Error("Should have performed health checks during patrol")
	}

	t.Logf("Patrol report: Checked %d components", report.ComponentsChecked)
	t.Log("✓ Patrol system working correctly")
}

// =================================================================================
// STATISTICS TESTS
// =================================================================================

func TestMicrogliaMaintenanceStats(t *testing.T) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Initial stats should be zero/empty
	initialStats := microglia.GetMaintenanceStats()
	if initialStats.ComponentsCreated != 0 {
		t.Errorf("Expected 0 initial components created, got %d", initialStats.ComponentsCreated)
	}

	if initialStats.LastResetTime.IsZero() {
		t.Error("Last reset time should be set")
	}

	// Create some components
	for i := 0; i < 3; i++ {
		componentInfo := ComponentInfo{
			ID:       fmt.Sprintf("stats_neuron_%d", i),
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i * 10), Y: 0, Z: 0},
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)
	}

	// Update health for components
	for i := 0; i < 3; i++ {
		microglia.UpdateComponentHealth(fmt.Sprintf("stats_neuron_%d", i), 0.7, 5)
	}

	// Remove one component
	microglia.RemoveComponent("stats_neuron_0")

	// Check final stats
	finalStats := microglia.GetMaintenanceStats()

	if finalStats.ComponentsCreated != 3 {
		t.Errorf("Expected 3 components created, got %d", finalStats.ComponentsCreated)
	}

	if finalStats.ComponentsRemoved != 1 {
		t.Errorf("Expected 1 component removed, got %d", finalStats.ComponentsRemoved)
	}

	if finalStats.HealthChecks != 3 {
		t.Errorf("Expected 3 health checks, got %d", finalStats.HealthChecks)
	}

	if finalStats.AverageHealthScore <= 0 {
		t.Error("Average health score should be positive")
	}

	t.Logf("Final stats: Created=%d, Removed=%d, Health checks=%d, Avg health=%.3f",
		finalStats.ComponentsCreated, finalStats.ComponentsRemoved,
		finalStats.HealthChecks, finalStats.AverageHealthScore)

	t.Log("✓ Maintenance statistics working correctly")
}

// =================================================================================
// BASIC THREAD SAFETY TESTS
// =================================================================================

func TestMicrogliaBasicConcurrency(t *testing.T) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Create components concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			componentInfo := ComponentInfo{
				ID:       fmt.Sprintf("concurrent_neuron_%d", id),
				Type:     ComponentNeuron,
				Position: Position3D{X: float64(id), Y: 0, Z: 0},
				State:    StateActive,
			}
			if err := microglia.CreateComponent(componentInfo); err != nil {
				errors <- err
			}
		}(i)
	}

	// Update health concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			microglia.UpdateComponentHealth(fmt.Sprintf("concurrent_neuron_%d", id), 0.5, 3)
		}(i)
	}

	// Mark for pruning concurrently
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			microglia.MarkForPruning(fmt.Sprintf("synapse_%d", id), "src", "dst", 0.1)
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent operation failed: %v", err)
	}

	// Verify final state is consistent
	stats := microglia.GetMaintenanceStats()
	if stats.ComponentsCreated != 5 {
		t.Errorf("Expected 5 components created, got %d", stats.ComponentsCreated)
	}

	candidates := microglia.GetPruningCandidates()
	if len(candidates) != 3 {
		t.Errorf("Expected 3 pruning candidates, got %d", len(candidates))
	}

	t.Log("✓ Basic concurrency working correctly")
}
