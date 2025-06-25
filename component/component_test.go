package component

import (
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

// ============================================================================
// BASE COMPONENT TESTS
// ============================================================================

func TestNewBaseComponent(t *testing.T) {
	id := "test-component-1"
	componentType := types.TypeNeuron
	position := types.Position3D{X: 10.0, Y: 20.0, Z: 30.0}

	comp := NewBaseComponent(id, componentType, position)

	if comp.ID() != id {
		t.Errorf("Expected ID %s, got %s", id, comp.ID())
	}

	if comp.Type() != componentType {
		t.Errorf("Expected type %v, got %v", componentType, comp.Type())
	}

	if comp.Position() != position {
		t.Errorf("Expected position %v, got %v", position, comp.Position())
	}

	if comp.State() != types.StateActive {
		t.Errorf("Expected state %v, got %v", types.StateActive, comp.State())
	}

	if !comp.IsActive() {
		t.Error("Expected component to be active")
	}
}

func TestBaseComponentLifecycle(t *testing.T) {
	comp := NewBaseComponent("test", types.TypeNeuron, types.Position3D{})

	// Test Start
	err := comp.Start()
	if err != nil {
		t.Errorf("Start() returned error: %v", err)
	}

	if !comp.IsActive() {
		t.Error("Component should be active after Start()")
	}

	if comp.State() != types.StateActive {
		t.Errorf("Expected state %v after Start(), got %v", types.StateActive, comp.State())
	}

	// Test Stop - UPDATED: Now expects types.StateStopped instead of types.StateInactive
	err = comp.Stop()
	if err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}

	if comp.IsActive() {
		t.Error("Component should be inactive after Stop()")
	}

	if comp.State() != types.StateStopped {
		t.Errorf("Expected state %v after Stop(), got %v", types.StateStopped, comp.State())
	}
}

func TestBaseComponentStateManagement(t *testing.T) {
	comp := NewBaseComponent("test", types.TypeNeuron, types.Position3D{})

	// Test state changes - UPDATED to include types.StateStopped
	states := []types.ComponentState{
		types.StateDeveloping,
		types.StateActive,
		types.StateShuttingDown,
		types.StateStopped, // NEW: Added types.types.StateStopped
		types.StateDying,
		types.StateInactive,
		types.StateDamaged,
		types.StateMaintenance,
		types.StateHibernating,
	}

	for _, state := range states {
		comp.SetState(state)
		if comp.State() != state {
			t.Errorf("Expected state %v, got %v", state, comp.State())
		}

		// Only types.StateActive should make IsActive() return true
		expectedActive := (state == types.StateActive && comp.isActive)
		if comp.IsActive() != expectedActive {
			t.Errorf("For state %v, expected IsActive() = %v, got %v", state, expectedActive, comp.IsActive())
		}
	}
}

func TestBaseComponentMetadata(t *testing.T) {
	comp := NewBaseComponent("test", types.TypeNeuron, types.Position3D{})

	// Test empty metadata
	metadata := comp.GetMetadata()
	if len(metadata) != 0 {
		t.Errorf("Expected empty metadata, got %v", metadata)
	}

	// Test adding metadata
	comp.UpdateMetadata("key1", "value1")
	comp.UpdateMetadata("key2", 42)
	comp.UpdateMetadata("key3", true)

	metadata = comp.GetMetadata()
	if len(metadata) != 3 {
		t.Errorf("Expected 3 metadata entries, got %d", len(metadata))
	}

	if metadata["key1"] != "value1" {
		t.Errorf("Expected key1 = 'value1', got %v", metadata["key1"])
	}

	if metadata["key2"] != 42 {
		t.Errorf("Expected key2 = 42, got %v", metadata["key2"])
	}

	if metadata["key3"] != true {
		t.Errorf("Expected key3 = true, got %v", metadata["key3"])
	}

	// Test metadata isolation (modifications shouldn't affect internal state)
	metadata["key4"] = "external"
	internalMetadata := comp.GetMetadata()
	if _, exists := internalMetadata["key4"]; exists {
		t.Error("External metadata modification should not affect internal state")
	}
}

func TestBaseComponentPositioning(t *testing.T) {
	comp := NewBaseComponent("test", types.TypeNeuron, types.Position3D{X: 1, Y: 2, Z: 3})

	// Test initial position
	pos := comp.Position()
	expected := types.Position3D{X: 1, Y: 2, Z: 3}
	if pos != expected {
		t.Errorf("Expected position %v, got %v", expected, pos)
	}

	// Test position update
	newPos := types.Position3D{X: 10, Y: 20, Z: 30}
	comp.SetPosition(newPos)
	pos = comp.Position()
	if pos != newPos {
		t.Errorf("Expected position %v, got %v", newPos, pos)
	}
}

func TestBaseComponentActivityTracking(t *testing.T) {
	comp := NewBaseComponent("test", types.TypeNeuron, types.Position3D{})

	// Test initial activity
	initialActivity := comp.GetLastActivity()
	if time.Since(initialActivity) > time.Second {
		t.Error("Initial activity should be recent")
	}

	// Test activity level calculation
	activityLevel := comp.GetActivityLevel()
	if activityLevel != 1.0 {
		t.Errorf("Expected activity level 1.0 for recent activity, got %f", activityLevel)
	}

	// Simulate old activity by setting lastActivity to past
	comp.mu.Lock()
	comp.lastActivity = time.Now().Add(-5 * time.Second)
	comp.mu.Unlock()

	activityLevel = comp.GetActivityLevel()
	if activityLevel != 0.5 {
		t.Errorf("Expected activity level 0.5 for medium-old activity, got %f", activityLevel)
	}

	// Simulate very old activity
	comp.mu.Lock()
	comp.lastActivity = time.Now().Add(-15 * time.Second)
	comp.mu.Unlock()

	activityLevel = comp.GetActivityLevel()
	if activityLevel != 0.0 {
		t.Errorf("Expected activity level 0.0 for old activity, got %f", activityLevel)
	}
}

// ============================================================================
// COMPONENT TYPE TESTS
// ============================================================================

func TestComponentTypeString(t *testing.T) {
	tests := []struct {
		componentType types.ComponentType
		expected      string
	}{
		{types.TypeNeuron, "Neuron"},
		{types.TypeSynapse, "Synapse"},
		{types.TypeGlialCell, "GlialCell"},
		{types.TypeMicrogliaCell, "MicrogliaCell"},
		{types.TypeEpendymalCell, "EpendymalCell"},
		{types.ComponentType(999), "Unknown"},
	}

	for _, test := range tests {
		result := test.componentType.String()
		if result != test.expected {
			t.Errorf("ComponentType(%d).String() = %s, expected %s", test.componentType, result, test.expected)
		}
	}
}

func TestComponentStateString(t *testing.T) {
	tests := []struct {
		state    types.ComponentState
		expected string
	}{
		{types.StateActive, "Active"},
		{types.StateInactive, "Inactive"},
		{types.StateShuttingDown, "ShuttingDown"},
		{types.StateStopped, "Stopped"}, // NEW: Added types.StateStopped
		{types.StateDeveloping, "Developing"},
		{types.StateDying, "Dying"},
		{types.StateDamaged, "Damaged"},
		{types.StateMaintenance, "Maintenance"},
		{types.StateHibernating, "Hibernating"},
		{types.ComponentState(999), "Unknown"},
	}

	for _, test := range tests {
		result := test.state.String()
		if result != test.expected {
			t.Errorf("ComponentState(%d).String() = %s, expected %s", test.state, result, test.expected)
		}
	}
}

// ============================================================================
// POSITION 3D TESTS
// ============================================================================

func TestPosition3DBasic(t *testing.T) {
	pos1 := types.Position3D{X: 1.0, Y: 2.0, Z: 3.0}

	if pos1.X != 1.0 {
		t.Errorf("Expected X = 1.0, got %f", pos1.X)
	}

	if pos1.Y != 2.0 {
		t.Errorf("Expected Y = 2.0, got %f", pos1.Y)
	}

	if pos1.Z != 3.0 {
		t.Errorf("Expected Z = 3.0, got %f", pos1.Z)
	}
}

// ============================================================================
// SPATIAL COMPONENT TESTS
// ============================================================================

func TestDefaultSpatialComponent(t *testing.T) {
	position := types.Position3D{X: 0, Y: 0, Z: 0}
	range_ := 10.0
	comp := NewSpatialComponent("spatial-test", types.TypeNeuron, position, range_)

	if comp.GetRange() != range_ {
		t.Errorf("Expected range %f, got %f", range_, comp.GetRange())
	}

	// Test that it's still a valid component
	if comp.ID() != "spatial-test" {
		t.Errorf("Expected ID 'spatial-test', got %s", comp.ID())
	}

	if comp.Type() != types.TypeNeuron {
		t.Errorf("Expected type %v, got %v", types.TypeNeuron, comp.Type())
	}

	// Test position setting
	newPos := types.Position3D{X: 5, Y: 5, Z: 5}
	comp.SetPosition(newPos)

	currentPos := comp.Position()
	if currentPos != newPos {
		t.Errorf("Expected position %v, got %v", newPos, currentPos)
	}
}

// ============================================================================
// MONITORABLE COMPONENT TESTS
// ============================================================================

func TestDefaultMonitorableComponent(t *testing.T) {
	comp := NewMonitorableComponent("monitor-test", types.TypeNeuron, types.Position3D{})

	// Test initial health metrics
	metrics := comp.GetHealthMetrics()
	if metrics.ActivityLevel != 0.0 {
		t.Errorf("Expected initial activity level 0.0, got %f", metrics.ActivityLevel)
	}

	if metrics.HealthScore != 1.0 {
		t.Errorf("Expected initial health score 1.0, got %f", metrics.HealthScore)
	}

	if len(metrics.Issues) != 0 {
		t.Errorf("Expected no initial issues, got %d", len(metrics.Issues))
	}

	// Test updating health metrics
	newMetrics := HealthMetrics{
		ActivityLevel:   0.8,
		ConnectionCount: 5,
		ProcessingLoad:  0.6,
		HealthScore:     0.9,
		Issues:          []string{"minor issue"},
	}

	comp.UpdateHealthMetrics(newMetrics)
	updatedMetrics := comp.GetHealthMetrics()

	if updatedMetrics.ActivityLevel != 0.8 {
		t.Errorf("Expected activity level 0.8, got %f", updatedMetrics.ActivityLevel)
	}

	if updatedMetrics.ConnectionCount != 5 {
		t.Errorf("Expected connection count 5, got %d", updatedMetrics.ConnectionCount)
	}

	if updatedMetrics.ProcessingLoad != 0.6 {
		t.Errorf("Expected processing load 0.6, got %f", updatedMetrics.ProcessingLoad)
	}

	if updatedMetrics.HealthScore != 0.9 {
		t.Errorf("Expected health score 0.9, got %f", updatedMetrics.HealthScore)
	}

	if len(updatedMetrics.Issues) != 1 || updatedMetrics.Issues[0] != "minor issue" {
		t.Errorf("Expected one issue 'minor issue', got %v", updatedMetrics.Issues)
	}

	// Test that LastHealthCheck was updated
	if time.Since(updatedMetrics.LastHealthCheck) > time.Second {
		t.Error("LastHealthCheck should be recent")
	}

	// Test metrics isolation
	updatedMetrics.Issues[0] = "modified"
	isolatedMetrics := comp.GetHealthMetrics()
	if isolatedMetrics.Issues[0] != "minor issue" {
		t.Error("External metrics modification should not affect internal state")
	}
}

// ============================================================================
// UTILITY FUNCTION TESTS
// ============================================================================

func TestCreateComponentInfo(t *testing.T) {
	comp := NewBaseComponent("info-test", types.TypeSynapse, types.Position3D{X: 1, Y: 2, Z: 3})
	comp.UpdateMetadata("test-key", "test-value")
	comp.SetState(types.StateDeveloping)

	info := CreateComponentInfo(comp)

	if info.ID != comp.ID() {
		t.Errorf("Expected ID %s, got %s", comp.ID(), info.ID)
	}

	if info.Type != comp.Type() {
		t.Errorf("Expected type %v, got %v", comp.Type(), info.Type)
	}

	if info.Position != comp.Position() {
		t.Errorf("Expected position %v, got %v", comp.Position(), info.Position)
	}

	if info.State != comp.State() {
		t.Errorf("Expected state %v, got %v", comp.State(), info.State)
	}

	if info.Metadata["test-key"] != "test-value" {
		t.Errorf("Expected metadata value 'test-value', got %v", info.Metadata["test-key"])
	}

	if time.Since(info.RegisteredAt) > time.Second {
		t.Error("RegisteredAt should be recent")
	}
}

func TestFilterComponentsByType(t *testing.T) {
	components := []Component{
		NewBaseComponent("neuron1", types.TypeNeuron, types.Position3D{}),
		NewBaseComponent("synapse1", types.TypeSynapse, types.Position3D{}),
		NewBaseComponent("neuron2", types.TypeNeuron, types.Position3D{}),
		NewBaseComponent("glia1", types.TypeGlialCell, types.Position3D{}),
		NewBaseComponent("microglia1", types.TypeMicrogliaCell, types.Position3D{}),
	}

	neurons := FilterComponentsByType(components, types.TypeNeuron)
	if len(neurons) != 2 {
		t.Errorf("Expected 2 neurons, got %d", len(neurons))
	}

	synapses := FilterComponentsByType(components, types.TypeSynapse)
	if len(synapses) != 1 {
		t.Errorf("Expected 1 synapse, got %d", len(synapses))
	}

	glia := FilterComponentsByType(components, types.TypeGlialCell)
	if len(glia) != 1 {
		t.Errorf("Expected 1 glial cell, got %d", len(glia))
	}

	microglia := FilterComponentsByType(components, types.TypeMicrogliaCell)
	if len(microglia) != 1 {
		t.Errorf("Expected 1 microglia cell, got %d", len(microglia))
	}
}

func TestFilterComponentsByState(t *testing.T) {
	components := []Component{
		NewBaseComponent("comp1", types.TypeNeuron, types.Position3D{}),
		NewBaseComponent("comp2", types.TypeNeuron, types.Position3D{}),
		NewBaseComponent("comp3", types.TypeNeuron, types.Position3D{}),
	}

	// Set different states
	components[1].SetState(types.StateInactive)
	components[2].SetState(types.StateDying)

	activeComponents := FilterComponentsByState(components, types.StateActive)
	if len(activeComponents) != 1 {
		t.Errorf("Expected 1 active component, got %d", len(activeComponents))
	}

	inactiveComponents := FilterComponentsByState(components, types.StateInactive)
	if len(inactiveComponents) != 1 {
		t.Errorf("Expected 1 inactive component, got %d", len(inactiveComponents))
	}

	dyingComponents := FilterComponentsByState(components, types.StateDying)
	if len(dyingComponents) != 1 {
		t.Errorf("Expected 1 dying component, got %d", len(dyingComponents))
	}
}

// ============================================================================
// CONCURRENT ACCESS TESTS
// ============================================================================

func TestConcurrentAccess(t *testing.T) {
	comp := NewBaseComponent("concurrent-test", types.TypeNeuron, types.Position3D{})

	// Test concurrent metadata updates
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			comp.UpdateMetadata("key1", i)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			comp.UpdateMetadata("key2", i*2)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Verify no race conditions occurred
	metadata := comp.GetMetadata()
	if _, exists := metadata["key1"]; !exists {
		t.Error("key1 should exist after concurrent updates")
	}
	if _, exists := metadata["key2"]; !exists {
		t.Error("key2 should exist after concurrent updates")
	}
}

func TestConcurrentStateChanges(t *testing.T) {
	comp := NewBaseComponent("state-test", types.TypeNeuron, types.Position3D{})

	// UPDATED: Added types.StateStopped to concurrent state testing
	states := []types.ComponentState{
		types.StateActive,
		types.StateInactive,
		types.StateShuttingDown,
		types.StateStopped, // NEW: Added types.types.StateStopped
		types.StateDeveloping,
		types.StateDying,
	}

	done := make(chan bool, len(states))

	// Multiple goroutines changing state
	for _, state := range states {
		go func(s types.ComponentState) {
			for i := 0; i < 50; i++ {
				comp.SetState(s)
			}
			done <- true
		}(state)
	}

	// Wait for all goroutines
	for i := 0; i < len(states); i++ {
		<-done
	}

	// Verify state is one of the valid states
	finalState := comp.State()
	validState := false
	for _, state := range states {
		if finalState == state {
			validState = true
			break
		}
	}

	if !validState {
		t.Errorf("Final state %v is not one of the expected states", finalState)
	}
}

func TestBaseComponentStopLifecycle(t *testing.T) {
	comp := NewBaseComponent("test", types.TypeNeuron, types.Position3D{})

	// Verify initial state
	if comp.State() != types.StateActive {
		t.Errorf("Expected initial state %v, got %v", types.StateActive, comp.State())
	}

	if !comp.IsActive() {
		t.Error("Component should be active initially")
	}

	// Test Stop with proper state transition
	err := comp.Stop()
	if err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}

	// After stop, component should be in types.StateStopped (not types.StateInactive)
	if comp.State() != types.StateStopped {
		t.Errorf("Expected state %v after Stop(), got %v", types.StateStopped, comp.State())
	}

	if comp.IsActive() {
		t.Error("Component should not be active after Stop()")
	}
}

func TestBaseComponentCanRestart(t *testing.T) {
	comp := NewBaseComponent("test", types.TypeNeuron, types.Position3D{})

	// Test states that can be restarted
	restartableStates := []types.ComponentState{
		types.StateInactive,
		types.StateStopped,
		types.StateMaintenance,
		types.StateHibernating,
	}

	for _, state := range restartableStates {
		comp.SetState(state)
		comp.isActive = false // Simulate stopped state

		if !comp.CanRestart() {
			t.Errorf("Component in state %v should be able to restart", state)
		}
	}

	// Test states that cannot be restarted
	nonRestartableStates := []types.ComponentState{
		types.StateActive,
		types.StateShuttingDown,
		types.StateDeveloping,
		types.StateDying,
		types.StateDamaged,
	}

	for _, state := range nonRestartableStates {
		comp.SetState(state)

		if comp.CanRestart() {
			t.Errorf("Component in state %v should not be able to restart", state)
		}
	}
}

func TestBaseComponentRestart(t *testing.T) {
	comp := NewBaseComponent("test", types.TypeNeuron, types.Position3D{})

	// Stop the component first
	comp.Stop()
	if comp.State() != types.StateStopped {
		t.Errorf("Expected state %v after Stop(), got %v", types.StateStopped, comp.State())
	}

	// Test successful restart
	err := comp.Restart()
	if err != nil {
		t.Errorf("Restart() returned error: %v", err)
	}

	if comp.State() != types.StateActive {
		t.Errorf("Expected state %v after Restart(), got %v", types.StateActive, comp.State())
	}

	if !comp.IsActive() {
		t.Error("Component should be active after Restart()")
	}

	// Test restart from invalid state
	comp.SetState(types.StateDying)
	err = comp.Restart()
	if err == nil {
		t.Error("Restart() should return error when called from types.StateDying")
	}

	if comp.State() == types.StateActive {
		t.Error("Component should not become active after failed restart")
	}
}

func TestBaseComponentFullLifecycle(t *testing.T) {
	comp := NewBaseComponent("lifecycle-test", types.TypeNeuron, types.Position3D{})

	// Test complete lifecycle: Start -> Stop -> Restart -> Stop

	// Initial state should be active
	if !comp.IsActive() || comp.State() != types.StateActive {
		t.Error("Component should start in active state")
	}

	// Start should work even if already active
	err := comp.Start()
	if err != nil {
		t.Errorf("Start() on active component returned error: %v", err)
	}

	// Stop the component
	err = comp.Stop()
	if err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}

	if comp.IsActive() || comp.State() != types.StateStopped {
		t.Error("Component should be stopped after Stop()")
	}

	// Restart the component
	err = comp.Restart()
	if err != nil {
		t.Errorf("Restart() returned error: %v", err)
	}

	if !comp.IsActive() || comp.State() != types.StateActive {
		t.Error("Component should be active after Restart()")
	}

	// Stop again
	err = comp.Stop()
	if err != nil {
		t.Errorf("Second Stop() returned error: %v", err)
	}

	if comp.IsActive() || comp.State() != types.StateStopped {
		t.Error("Component should be stopped after second Stop()")
	}
}

func TestFilterComponentsByStopped(t *testing.T) {
	components := []Component{
		NewBaseComponent("comp1", types.TypeNeuron, types.Position3D{}),
		NewBaseComponent("comp2", types.TypeNeuron, types.Position3D{}),
		NewBaseComponent("comp3", types.TypeNeuron, types.Position3D{}),
		NewBaseComponent("comp4", types.TypeNeuron, types.Position3D{}),
	}

	// Set different states including types.StateStopped
	components[0].SetState(types.StateActive)
	components[1].SetState(types.StateStopped)
	components[2].SetState(types.StateInactive)
	components[3].SetState(types.StateStopped)

	// Test filtering by types.StateStopped
	stoppedComponents := FilterComponentsByState(components, types.StateStopped)
	if len(stoppedComponents) != 2 {
		t.Errorf("Expected 2 stopped components, got %d", len(stoppedComponents))
	}

	// Verify the correct components are returned
	stoppedIDs := make(map[string]bool)
	for _, comp := range stoppedComponents {
		stoppedIDs[comp.ID()] = true
	}

	if !stoppedIDs["comp2"] || !stoppedIDs["comp4"] {
		t.Error("FilterComponentsByState should return comp2 and comp4 for types.StateStopped")
	}
}

func TestStatePersistenceAfterRestart(t *testing.T) {
	comp := NewBaseComponent("persistence-test", types.TypeNeuron, types.Position3D{})

	// Add some metadata
	comp.UpdateMetadata("test-key", "test-value")
	originalMetadata := comp.GetMetadata()

	// Stop and restart
	comp.Stop()
	comp.Restart()

	// Verify metadata is preserved
	newMetadata := comp.GetMetadata()
	if len(newMetadata) != len(originalMetadata) {
		t.Error("Metadata should be preserved after restart")
	}

	if newMetadata["test-key"] != "test-value" {
		t.Error("Specific metadata values should be preserved after restart")
	}

	// Verify position is preserved
	originalPos := types.Position3D{X: 10, Y: 20, Z: 30}
	comp.SetPosition(originalPos)
	comp.Stop()
	comp.Restart()

	newPos := comp.Position()
	if newPos != originalPos {
		t.Errorf("Position should be preserved after restart: expected %v, got %v", originalPos, newPos)
	}
}

func TestConcurrentStopRestart(t *testing.T) {
	comp := NewBaseComponent("concurrent-lifecycle", types.TypeNeuron, types.Position3D{})

	done := make(chan bool, 2)

	// Goroutine 1: Stop and restart repeatedly
	go func() {
		for i := 0; i < 50; i++ {
			comp.Stop()
			if comp.CanRestart() {
				comp.Restart()
			}
		}
		done <- true
	}()

	// Goroutine 2: Check state and try operations
	go func() {
		for i := 0; i < 50; i++ {
			state := comp.State()
			isActive := comp.IsActive()
			canRestart := comp.CanRestart()

			// Just verify these calls don't panic
			_ = state
			_ = isActive
			_ = canRestart
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Final state should be valid
	finalState := comp.State()
	validStates := []types.ComponentState{types.StateActive, types.StateStopped}
	validFinalState := false
	for _, state := range validStates {
		if finalState == state {
			validFinalState = true
			break
		}
	}

	if !validFinalState {
		t.Errorf("Final state after concurrent operations should be valid, got %v", finalState)
	}
}

// ============================================================================
// INTERFACE TESTS
// ============================================================================

func TestComponentInterface(t *testing.T) {
	// Test that BaseComponent implements Component interface
	var comp Component = NewBaseComponent("test", types.TypeNeuron, types.Position3D{})

	// Test basic interface methods
	if comp.ID() != "test" {
		t.Errorf("Expected ID 'test', got %s", comp.ID())
	}

	if comp.Type() != types.TypeNeuron {
		t.Errorf("Expected type %v, got %v", types.TypeNeuron, comp.Type())
	}

	if !comp.IsActive() {
		t.Error("Component should be active")
	}
}

func TestSpatialComponentInterface(t *testing.T) {
	// Test that DefaultSpatialComponent implements SpatialComponent interface
	var spatial SpatialComponent = NewSpatialComponent("spatial", types.TypeNeuron, types.Position3D{}, 5.0)

	if spatial.GetRange() != 5.0 {
		t.Errorf("Expected range 5.0, got %f", spatial.GetRange())
	}

	// Test component interface is also implemented
	if spatial.ID() != "spatial" {
		t.Errorf("Expected ID 'spatial', got %s", spatial.ID())
	}
}

func TestMonitorableComponentInterface(t *testing.T) {
	// Test that DefaultMonitorableComponent implements MonitorableComponent interface
	var monitorable MonitorableComponent = NewMonitorableComponent("monitor", types.TypeNeuron, types.Position3D{})

	metrics := monitorable.GetHealthMetrics()
	if metrics.HealthScore != 1.0 {
		t.Errorf("Expected health score 1.0, got %f", metrics.HealthScore)
	}

	// Test component interface is also implemented
	if monitorable.ID() != "monitor" {
		t.Errorf("Expected ID 'monitor', got %s", monitorable.ID())
	}
}

// ============================================================================
// MOCK IMPLEMENTATIONS FOR INTERFACE TESTING
// ============================================================================

// MockChemicalReceiver for testing chemical interfaces
type MockChemicalReceiver struct {
	*BaseComponent
	receptors []types.LigandType
	bindings  map[types.LigandType]float64
}

func NewMockChemicalReceiver(id string) *MockChemicalReceiver {
	return &MockChemicalReceiver{
		BaseComponent: NewBaseComponent(id, types.TypeNeuron, types.Position3D{}),
		receptors:     []types.LigandType{types.LigandGlutamate, types.LigandGABA},
		bindings:      make(map[types.LigandType]float64),
	}
}

func (mcr *MockChemicalReceiver) GetReceptors() []types.LigandType {
	return mcr.receptors
}

func (mcr *MockChemicalReceiver) Bind(ligandType types.LigandType, sourceID string, concentration float64) {
	mcr.bindings[ligandType] = concentration
}

func TestChemicalReceiverInterface(t *testing.T) {
	// Test that MockChemicalReceiver implements ChemicalReceiver interface
	var receiver ChemicalReceiver = NewMockChemicalReceiver("chemical-test")

	receptors := receiver.GetReceptors()
	if len(receptors) != 2 {
		t.Errorf("Expected 2 receptors, got %d", len(receptors))
	}

	// Test binding
	receiver.Bind(types.LigandGlutamate, "source1", 0.5)

	// Verify through mock implementation
	mock := receiver.(*MockChemicalReceiver)
	if mock.bindings[types.LigandGlutamate] != 0.5 {
		t.Errorf("Expected binding concentration 0.5, got %f", mock.bindings[types.LigandGlutamate])
	}

	// Test component interface is also implemented
	if receiver.ID() != "chemical-test" {
		t.Errorf("Expected ID 'chemical-test', got %s", receiver.ID())
	}
}

// MockMessageReceiver for testing message interfaces
type MockMessageReceiver struct {
	*BaseComponent
	receivedMessages []types.NeuralSignal
}

func NewMockMessageReceiver(id string) *MockMessageReceiver {
	return &MockMessageReceiver{
		BaseComponent:    NewBaseComponent(id, types.TypeNeuron, types.Position3D{}),
		receivedMessages: make([]types.NeuralSignal, 0),
	}
}

func (mmr *MockMessageReceiver) Receive(msg types.NeuralSignal) {
	mmr.receivedMessages = append(mmr.receivedMessages, msg)
}

func TestMessageReceiverInterface(t *testing.T) {
	// Test that MockMessageReceiver implements MessageReceiver interface
	var receiver MessageReceiver = NewMockMessageReceiver("message-test")

	// Test receiving a message
	testMessage := types.NeuralSignal{
		Value:    1.0,
		SourceID: "source1",
		TargetID: "message-test",
	}

	receiver.Receive(testMessage)

	// Verify through mock implementation
	mock := receiver.(*MockMessageReceiver)
	if len(mock.receivedMessages) != 1 {
		t.Errorf("Expected 1 received message, got %d", len(mock.receivedMessages))
	}

	if mock.receivedMessages[0].Value != 1.0 {
		t.Errorf("Expected message value 1.0, got %f", mock.receivedMessages[0].Value)
	}

	// Test component interface is also implemented
	if receiver.ID() != "message-test" {
		t.Errorf("Expected ID 'message-test', got %s", receiver.ID())
	}
}
