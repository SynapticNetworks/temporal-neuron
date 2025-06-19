package component

import (
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/message"
)

// ============================================================================
// BASE COMPONENT TESTS
// ============================================================================

func TestNewBaseComponent(t *testing.T) {
	id := "test-component-1"
	componentType := TypeNeuron
	position := Position3D{X: 10.0, Y: 20.0, Z: 30.0}

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

	if comp.State() != StateActive {
		t.Errorf("Expected state %v, got %v", StateActive, comp.State())
	}

	if !comp.IsActive() {
		t.Error("Expected component to be active")
	}
}

func TestBaseComponentLifecycle(t *testing.T) {
	comp := NewBaseComponent("test", TypeNeuron, Position3D{})

	// Test Start
	err := comp.Start()
	if err != nil {
		t.Errorf("Start() returned error: %v", err)
	}

	if !comp.IsActive() {
		t.Error("Component should be active after Start()")
	}

	if comp.State() != StateActive {
		t.Errorf("Expected state %v after Start(), got %v", StateActive, comp.State())
	}

	// Test Stop
	err = comp.Stop()
	if err != nil {
		t.Errorf("Stop() returned error: %v", err)
	}

	if comp.IsActive() {
		t.Error("Component should be inactive after Stop()")
	}

	if comp.State() != StateInactive {
		t.Errorf("Expected state %v after Stop(), got %v", StateInactive, comp.State())
	}
}

func TestBaseComponentStateManagement(t *testing.T) {
	comp := NewBaseComponent("test", TypeNeuron, Position3D{})

	// Test state changes
	states := []ComponentState{
		StateDeveloping,
		StateActive,
		StateShuttingDown,
		StateDying,
		StateInactive,
		StateDamaged,
		StateMaintenance,
		StateHibernating,
	}

	for _, state := range states {
		comp.SetState(state)
		if comp.State() != state {
			t.Errorf("Expected state %v, got %v", state, comp.State())
		}

		// Only StateActive should make IsActive() return true
		expectedActive := (state == StateActive && comp.isActive)
		if comp.IsActive() != expectedActive {
			t.Errorf("For state %v, expected IsActive() = %v, got %v", state, expectedActive, comp.IsActive())
		}
	}
}

func TestBaseComponentMetadata(t *testing.T) {
	comp := NewBaseComponent("test", TypeNeuron, Position3D{})

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
	comp := NewBaseComponent("test", TypeNeuron, Position3D{X: 1, Y: 2, Z: 3})

	// Test initial position
	pos := comp.Position()
	expected := Position3D{X: 1, Y: 2, Z: 3}
	if pos != expected {
		t.Errorf("Expected position %v, got %v", expected, pos)
	}

	// Test position update
	newPos := Position3D{X: 10, Y: 20, Z: 30}
	comp.SetPosition(newPos)
	pos = comp.Position()
	if pos != newPos {
		t.Errorf("Expected position %v, got %v", newPos, pos)
	}
}

func TestBaseComponentActivityTracking(t *testing.T) {
	comp := NewBaseComponent("test", TypeNeuron, Position3D{})

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
		componentType ComponentType
		expected      string
	}{
		{TypeNeuron, "Neuron"},
		{TypeSynapse, "Synapse"},
		{TypeGlialCell, "GlialCell"},
		{TypeMicrogliaCell, "MicrogliaCell"},
		{TypeEpendymalCell, "EpendymalCell"},
		{ComponentType(999), "Unknown"},
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
		state    ComponentState
		expected string
	}{
		{StateActive, "Active"},
		{StateInactive, "Inactive"},
		{StateShuttingDown, "ShuttingDown"},
		{StateDeveloping, "Developing"},
		{StateDying, "Dying"},
		{StateDamaged, "Damaged"},
		{StateMaintenance, "Maintenance"},
		{StateHibernating, "Hibernating"},
		{ComponentState(999), "Unknown"},
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
	pos1 := Position3D{X: 1.0, Y: 2.0, Z: 3.0}

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
	position := Position3D{X: 0, Y: 0, Z: 0}
	range_ := 10.0
	comp := NewSpatialComponent("spatial-test", TypeNeuron, position, range_)

	if comp.GetRange() != range_ {
		t.Errorf("Expected range %f, got %f", range_, comp.GetRange())
	}

	// Test that it's still a valid component
	if comp.ID() != "spatial-test" {
		t.Errorf("Expected ID 'spatial-test', got %s", comp.ID())
	}

	if comp.Type() != TypeNeuron {
		t.Errorf("Expected type %v, got %v", TypeNeuron, comp.Type())
	}

	// Test position setting
	newPos := Position3D{X: 5, Y: 5, Z: 5}
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
	comp := NewMonitorableComponent("monitor-test", TypeNeuron, Position3D{})

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
	comp := NewBaseComponent("info-test", TypeSynapse, Position3D{X: 1, Y: 2, Z: 3})
	comp.UpdateMetadata("test-key", "test-value")
	comp.SetState(StateDeveloping)

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

func TestComponentDistance(t *testing.T) {
	// This test is removed because distance calculation is now handled by the matrix
	// Components only store position data, they don't calculate distances
	t.Skip("Distance calculation is handled by the matrix, not by components")
}

func TestFilterComponentsByType(t *testing.T) {
	components := []Component{
		NewBaseComponent("neuron1", TypeNeuron, Position3D{}),
		NewBaseComponent("synapse1", TypeSynapse, Position3D{}),
		NewBaseComponent("neuron2", TypeNeuron, Position3D{}),
		NewBaseComponent("glia1", TypeGlialCell, Position3D{}),
		NewBaseComponent("microglia1", TypeMicrogliaCell, Position3D{}),
	}

	neurons := FilterComponentsByType(components, TypeNeuron)
	if len(neurons) != 2 {
		t.Errorf("Expected 2 neurons, got %d", len(neurons))
	}

	synapses := FilterComponentsByType(components, TypeSynapse)
	if len(synapses) != 1 {
		t.Errorf("Expected 1 synapse, got %d", len(synapses))
	}

	glia := FilterComponentsByType(components, TypeGlialCell)
	if len(glia) != 1 {
		t.Errorf("Expected 1 glial cell, got %d", len(glia))
	}

	microglia := FilterComponentsByType(components, TypeMicrogliaCell)
	if len(microglia) != 1 {
		t.Errorf("Expected 1 microglia cell, got %d", len(microglia))
	}
}

func TestFilterComponentsByState(t *testing.T) {
	components := []Component{
		NewBaseComponent("comp1", TypeNeuron, Position3D{}),
		NewBaseComponent("comp2", TypeNeuron, Position3D{}),
		NewBaseComponent("comp3", TypeNeuron, Position3D{}),
	}

	// Set different states
	components[1].SetState(StateInactive)
	components[2].SetState(StateDying)

	activeComponents := FilterComponentsByState(components, StateActive)
	if len(activeComponents) != 1 {
		t.Errorf("Expected 1 active component, got %d", len(activeComponents))
	}

	inactiveComponents := FilterComponentsByState(components, StateInactive)
	if len(inactiveComponents) != 1 {
		t.Errorf("Expected 1 inactive component, got %d", len(inactiveComponents))
	}

	dyingComponents := FilterComponentsByState(components, StateDying)
	if len(dyingComponents) != 1 {
		t.Errorf("Expected 1 dying component, got %d", len(dyingComponents))
	}
}

// ============================================================================
// CONCURRENT ACCESS TESTS
// ============================================================================

func TestConcurrentAccess(t *testing.T) {
	comp := NewBaseComponent("concurrent-test", TypeNeuron, Position3D{})

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
	comp := NewBaseComponent("state-test", TypeNeuron, Position3D{})

	states := []ComponentState{
		StateActive,
		StateInactive,
		StateDeveloping,
		StateDying,
	}

	done := make(chan bool, len(states))

	// Multiple goroutines changing state
	for _, state := range states {
		go func(s ComponentState) {
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

// ============================================================================
// INTERFACE TESTS
// ============================================================================

func TestComponentInterface(t *testing.T) {
	// Test that BaseComponent implements Component interface
	var comp Component = NewBaseComponent("test", TypeNeuron, Position3D{})

	// Test basic interface methods
	if comp.ID() != "test" {
		t.Errorf("Expected ID 'test', got %s", comp.ID())
	}

	if comp.Type() != TypeNeuron {
		t.Errorf("Expected type %v, got %v", TypeNeuron, comp.Type())
	}

	if !comp.IsActive() {
		t.Error("Component should be active")
	}
}

func TestSpatialComponentInterface(t *testing.T) {
	// Test that DefaultSpatialComponent implements SpatialComponent interface
	var spatial SpatialComponent = NewSpatialComponent("spatial", TypeNeuron, Position3D{}, 5.0)

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
	var monitorable MonitorableComponent = NewMonitorableComponent("monitor", TypeNeuron, Position3D{})

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
	receptors []message.LigandType
	bindings  map[message.LigandType]float64
}

func NewMockChemicalReceiver(id string) *MockChemicalReceiver {
	return &MockChemicalReceiver{
		BaseComponent: NewBaseComponent(id, TypeNeuron, Position3D{}),
		receptors:     []message.LigandType{message.LigandGlutamate, message.LigandGABA},
		bindings:      make(map[message.LigandType]float64),
	}
}

func (mcr *MockChemicalReceiver) GetReceptors() []message.LigandType {
	return mcr.receptors
}

func (mcr *MockChemicalReceiver) Bind(ligandType message.LigandType, sourceID string, concentration float64) {
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
	receiver.Bind(message.LigandGlutamate, "source1", 0.5)

	// Verify through mock implementation
	mock := receiver.(*MockChemicalReceiver)
	if mock.bindings[message.LigandGlutamate] != 0.5 {
		t.Errorf("Expected binding concentration 0.5, got %f", mock.bindings[message.LigandGlutamate])
	}

	// Test component interface is also implemented
	if receiver.ID() != "chemical-test" {
		t.Errorf("Expected ID 'chemical-test', got %s", receiver.ID())
	}
}

// MockMessageReceiver for testing message interfaces
type MockMessageReceiver struct {
	*BaseComponent
	receivedMessages []message.NeuralSignal
}

func NewMockMessageReceiver(id string) *MockMessageReceiver {
	return &MockMessageReceiver{
		BaseComponent:    NewBaseComponent(id, TypeNeuron, Position3D{}),
		receivedMessages: make([]message.NeuralSignal, 0),
	}
}

func (mmr *MockMessageReceiver) Receive(msg message.NeuralSignal) {
	mmr.receivedMessages = append(mmr.receivedMessages, msg)
}

func TestMessageReceiverInterface(t *testing.T) {
	// Test that MockMessageReceiver implements MessageReceiver interface
	var receiver MessageReceiver = NewMockMessageReceiver("message-test")

	// Test receiving a message
	testMessage := message.NeuralSignal{
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
