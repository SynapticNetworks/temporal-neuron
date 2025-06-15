/*
=================================================================================
MICROGLIA - UNIT TESTS
=================================================================================

Focused unit tests for the microglia lifecycle management and health monitoring
system. Tests the biological functions of component creation, removal, health
surveillance, pruning, and maintenance in isolation.

These tests complement the integration tests by providing detailed validation
of individual microglia functions and edge cases.
=================================================================================
*/

package extracellular

import (
	"fmt"
	"testing"
	"time"
)

// =================================================================================
// COMPONENT LIFECYCLE TESTS
// =================================================================================

func TestMicrogliaComponentCreation(t *testing.T) {
	t.Log("=== TESTING MICROGLIA COMPONENT CREATION ===")

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
	t.Log("=== TESTING MICROGLIA COMPONENT REMOVAL ===")

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
	t.Log("=== TESTING MICROGLIA HEALTH MONITORING ===")

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

	// Test health update with low activity
	microglia.UpdateComponentHealth("health_test_neuron", 0.05, 2)

	health, _ = microglia.GetComponentHealth("health_test_neuron")

	// Should detect low activity issues
	found_activity_issue := false
	for _, issue := range health.Issues {
		if issue == "very_low_activity" || issue == "low_activity" {
			found_activity_issue = true
			break
		}
	}

	if !found_activity_issue {
		t.Error("Should detect low activity issues")
	}

	t.Log("✓ Health monitoring working correctly")
}

func TestMicrogliaHealthScoreCalculation(t *testing.T) {
	t.Log("=== TESTING HEALTH SCORE CALCULATION ===")

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
		minScore    float64
		description string
	}{
		{0.8, 10, 0.8, "High activity, well connected"},
		{0.3, 5, 0.6, "Moderate activity, moderate connections"},
		{0.05, 2, 0.27, "Low activity, few connections"},
		{0.01, 1, 0.2, "Very low activity, poorly connected"},
	}

	for _, tc := range testCases {
		microglia.UpdateComponentHealth("score_test_neuron", tc.activity, tc.connections)

		health, _ := microglia.GetComponentHealth("score_test_neuron")

		if health.HealthScore < tc.minScore {
			t.Errorf("%s: Expected health score >= %.3f, got %.3f",
				tc.description, tc.minScore, health.HealthScore)
		}

		t.Logf("%s: Activity=%.3f, Connections=%d, Health=%.3f",
			tc.description, tc.activity, tc.connections, health.HealthScore)
	}

	t.Log("✓ Health score calculation working correctly")
}

func TestMicrogliaHealthIssueDetection(t *testing.T) {
	t.Log("=== TESTING HEALTH ISSUE DETECTION ===")

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

	foundLowActivity := false
	for _, issue := range health.Issues {
		if issue == "very_low_activity" || issue == "critically_low_activity" {
			foundLowActivity = true
			break
		}
	}
	if !foundLowActivity {
		t.Error("Should detect very low activity")
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

	// Test poor connectivity detection
	microglia.UpdateComponentHealth("issue_test_neuron", 0.5, 2)
	health, _ = microglia.GetComponentHealth("issue_test_neuron")

	foundPoorConnectivity := false
	for _, issue := range health.Issues {
		if issue == "poorly_connected" {
			foundPoorConnectivity = true
			break
		}
	}
	if !foundPoorConnectivity {
		t.Error("Should detect poor connectivity")
	}

	t.Log("✓ Health issue detection working correctly")
}

// =================================================================================
// PRUNING SYSTEM TESTS
// =================================================================================

func TestMicrogliaPruningCandidates(t *testing.T) {
	t.Log("=== TESTING PRUNING CANDIDATE SYSTEM ===")

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

	// Verify weak synapses have higher pruning scores
	for _, candidate := range candidates {
		if candidate.ConnectionID == "weak_synapse_1" || candidate.ConnectionID == "weak_synapse_2" {
			if candidate.PruningScore < 0.5 {
				t.Errorf("Weak synapse should have high pruning score, got %.3f", candidate.PruningScore)
			}
		}
		if candidate.ConnectionID == "strong_synapse" {
			if candidate.PruningScore > 0.5 {
				t.Errorf("Strong synapse should have low pruning score, got %.3f", candidate.PruningScore)
			}
		}
	}

	t.Log("✓ Pruning candidate system working correctly")
}

func TestMicrogliaPruningExecution(t *testing.T) {
	t.Log("=== TESTING PRUNING EXECUTION ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Mark a connection for pruning with very high score and old age
	microglia.MarkForPruning("old_weak_synapse", "neuron_X", "neuron_Y", 0.01)

	// Execute pruning (should not prune immediately due to age requirement)
	prunedConnections := microglia.ExecutePruning()

	// Should not prune connections that are too young
	if len(prunedConnections) > 0 {
		t.Error("Should not prune connections immediately marked")
	}

	// Verify connection is still in pruning candidates
	candidates := microglia.GetPruningCandidates()
	if len(candidates) != 1 {
		t.Error("Connection should still be in pruning candidates")
	}

	t.Log("✓ Pruning execution respects age requirements")
}

// =================================================================================
// PATROL SYSTEM TESTS
// =================================================================================

func TestMicrogliaPatrolRoutes(t *testing.T) {
	t.Log("=== TESTING PATROL ROUTE ESTABLISHMENT ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Establish patrol route
	territory := Territory{
		Center: Position3D{X: 50, Y: 50, Z: 50},
		Radius: 25.0,
	}

	microglia.EstablishPatrolRoute("microglia_1", territory, 100*time.Millisecond)

	// Create some components in the territory
	for i := 0; i < 5; i++ {
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

	t.Logf("Patrol report: Checked %d components", report.ComponentsChecked)
	t.Log("✓ Patrol route system working correctly")
}

func TestMicrogliaPatrolExecution(t *testing.T) {
	t.Log("=== TESTING PATROL EXECUTION ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Create territory with components
	territory := Territory{
		Center: Position3D{X: 0, Y: 0, Z: 0},
		Radius: 50.0,
	}

	microglia.EstablishPatrolRoute("patrol_microglia", territory, 50*time.Millisecond)

	// Add components in territory
	componentPositions := []Position3D{
		{X: 10, Y: 10, Z: 0},
		{X: 20, Y: 20, Z: 0},
		{X: 30, Y: 30, Z: 0},
	}

	for i, pos := range componentPositions {
		componentInfo := ComponentInfo{
			ID:       fmt.Sprintf("territory_neuron_%d", i),
			Type:     ComponentNeuron,
			Position: pos,
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)
	}

	// Execute multiple patrols
	initialStats := microglia.GetMaintenanceStats()

	for i := 0; i < 3; i++ {
		report := microglia.ExecutePatrol("patrol_microglia")

		if report.ComponentsChecked == 0 {
			t.Error("Patrol should check components in territory")
		}

		t.Logf("Patrol %d: Checked %d components", i+1, report.ComponentsChecked)
	}

	// Verify patrol statistics updated
	finalStats := microglia.GetMaintenanceStats()

	if finalStats.PatrolsCompleted <= initialStats.PatrolsCompleted {
		t.Error("Patrol count should have increased")
	}

	if finalStats.HealthChecks <= initialStats.HealthChecks {
		t.Error("Health check count should have increased")
	}

	t.Log("✓ Patrol execution working correctly")
}

// =================================================================================
// BIRTH REQUEST SYSTEM TESTS
// =================================================================================

func TestMicrogliaBirthRequests(t *testing.T) {
	t.Log("=== TESTING BIRTH REQUEST SYSTEM ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Submit birth request
	birthRequest := ComponentBirthRequest{
		ComponentType: ComponentNeuron,
		Position:      Position3D{X: 100, Y: 100, Z: 100},
		Justification: "High activity region needs additional processing capacity",
		Priority:      PriorityHigh,
		RequestedBy:   "network_analyzer",
		Metadata: map[string]interface{}{
			"target_firing_rate": 10.0,
			"connection_target":  20,
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

	// Verify component was actually created
	_, exists := astrocyteNetwork.Get(createdComponent.ID)
	if !exists {
		t.Error("Created component should exist in astrocyte network")
	}

	t.Log("✓ Birth request system working correctly")
}

// =================================================================================
// MAINTENANCE STATISTICS TESTS
// =================================================================================

func TestMicrogliaMaintenanceStats(t *testing.T) {
	t.Log("=== TESTING MAINTENANCE STATISTICS ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Initial stats should be zero
	initialStats := microglia.GetMaintenanceStats()
	if initialStats.ComponentsCreated != 0 {
		t.Errorf("Expected 0 initial components created, got %d", initialStats.ComponentsCreated)
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
// RESOURCE CONSTRAINT TESTS
// =================================================================================

func TestMicrogliaResourceConstraints(t *testing.T) {
	t.Log("=== TESTING RESOURCE CONSTRAINTS ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 100)

	// Test birth request evaluation with resource constraints
	// Submit low priority request when resources are available
	lowPriorityRequest := ComponentBirthRequest{
		ComponentType: ComponentNeuron,
		Position:      Position3D{X: 0, Y: 0, Z: 0},
		Justification: "Minor optimization",
		Priority:      PriorityLow,
		RequestedBy:   "optimizer",
	}

	err := microglia.RequestComponentBirth(lowPriorityRequest)
	if err != nil {
		t.Fatalf("Failed to submit low priority request: %v", err)
	}

	// Process requests - should be approved when resources available
	created := microglia.ProcessBirthRequests()
	if len(created) != 1 {
		t.Errorf("Low priority request should be approved when resources available")
	}

	// Create many components to simulate resource pressure
	for i := 0; i < 100; i++ {
		componentInfo := ComponentInfo{
			ID:       fmt.Sprintf("resource_neuron_%d", i),
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0},
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)
	}

	// Submit another low priority request - should be rejected due to resource constraints
	anotherLowPriorityRequest := ComponentBirthRequest{
		ComponentType: ComponentNeuron,
		Position:      Position3D{X: 200, Y: 0, Z: 0},
		Justification: "Another minor optimization",
		Priority:      PriorityLow,
		RequestedBy:   "optimizer",
	}

	err = microglia.RequestComponentBirth(anotherLowPriorityRequest)
	if err != nil {
		t.Fatalf("Failed to submit second low priority request: %v", err)
	}

	// Process requests - should be rejected due to resource constraints
	created = microglia.ProcessBirthRequests()
	if len(created) != 0 {
		t.Error("Low priority request should be rejected under resource pressure")
	}

	// High priority request should still be approved
	highPriorityRequest := ComponentBirthRequest{
		ComponentType: ComponentNeuron,
		Position:      Position3D{X: 300, Y: 0, Z: 0},
		Justification: "Critical network failure response",
		Priority:      PriorityHigh,
		RequestedBy:   "emergency_system",
	}

	err = microglia.RequestComponentBirth(highPriorityRequest)
	if err != nil {
		t.Fatalf("Failed to submit high priority request: %v", err)
	}

	created = microglia.ProcessBirthRequests()
	if len(created) != 1 {
		t.Error("High priority request should be approved even under resource pressure")
	}

	t.Log("✓ Resource constraint system working correctly")
}
