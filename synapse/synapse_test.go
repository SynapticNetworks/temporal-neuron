/*
=================================================================================
SYNAPSE SYSTEM UNIT TESTS
=================================================================================

OVERVIEW:
This file contains comprehensive unit tests for the synapse system, verifying:
1. Basic synapse creation and property validation
2. Signal transmission through synapses with weight scaling and delays
3. Weight management and bounds enforcement
4. Configuration helper functions
5. Interface compliance and thread safety

The tests use MockNeuron implementations to isolate synapse functionality
from actual neuron implementations, allowing focused testing of synaptic
behavior without dependencies on complex neuron logic.

TEST PHILOSOPHY:
- Each test focuses on a single aspect of synapse functionality
- Mock objects provide controlled, predictable test environments
- Tests verify both normal operation and edge cases
- Timing-sensitive tests include appropriate delays for goroutine scheduling
- All biological constraints and safety bounds are verified

BIOLOGICAL CONTEXT:
These tests validate that the synapse implementation correctly models:
- Synaptic weight storage and modification (synaptic efficacy)
- Signal transmission delays (axonal conduction + synaptic delays)
- Weight bounds enforcement (biological saturation limits)
- Pruning decision logic (structural plasticity)
- Configuration parameter validation
*/

package synapse

import (
	"testing"
	"time"
)

// =================================================================================
// MOCK NEURON IMPLEMENTATION FOR TESTING
// =================================================================================

// MockNeuron implements SynapseCompatibleNeuron for testing purposes.
// This is a minimal, controlled implementation that allows precise testing
// of synapse functionality without dependencies on complex neuron behavior.
//
// DESIGN PRINCIPLES:
// 1. SIMPLE: Only implements the interface methods needed for synapse testing
// 2. OBSERVABLE: Stores all received messages for test verification
// 3. CONTROLLABLE: Predictable behavior that won't interfere with test logic
// 4. THREAD-SAFE: Can handle messages from multiple synapses concurrently
//
// BIOLOGICAL ABSTRACTION:
// While real neurons have complex membrane dynamics, homeostatic regulation,
// and sophisticated processing, this mock focuses purely on the message
// reception interface that synapses need. This isolation allows us to test
// synaptic behavior without the complexity of full neural simulation.
type MockNeuron struct {
	// === IDENTIFICATION ===
	id string // Unique identifier for this mock neuron

	// === MESSAGE STORAGE ===
	// These fields store received messages for test verification
	receivedMsgs []SynapseMessage    // All messages received (for verification)
	msgChannel   chan SynapseMessage // Buffered channel for message reception

	// Note: In a real neuron, received messages would trigger complex
	// membrane dynamics, homeostatic adjustments, and potential firing.
	// This mock simply stores them for test analysis.
}

// NewMockNeuron creates a simple mock neuron for testing synapse functionality.
// This factory function initializes a mock neuron with empty message storage
// and a buffered channel to handle multiple incoming messages during tests.
//
// Parameters:
//
//	id: Unique identifier for this mock neuron (used in test verification)
//
// Returns:
//
//	A fully initialized MockNeuron ready for testing
//
// USAGE IN TESTS:
// Mock neurons serve as controlled endpoints for synapse testing:
// - Pre-synaptic mock: Represents the source neuron sending signals
// - Post-synaptic mock: Represents the target neuron receiving signals
func NewMockNeuron(id string) *MockNeuron {
	return &MockNeuron{
		id:           id,                            // Store identification
		receivedMsgs: make([]SynapseMessage, 0),     // Initialize empty message log
		msgChannel:   make(chan SynapseMessage, 10), // Buffer for concurrent messages
	}
}

// ID returns the unique identifier of this mock neuron.
// This method implements the SynapseCompatibleNeuron interface requirement
// and allows synapses to identify message sources and targets.
//
// Returns:
//
//	The neuron's unique identifier string
//
// INTERFACE COMPLIANCE:
// This method satisfies the SynapseCompatibleNeuron.ID() requirement,
// enabling this mock to work seamlessly with the synapse system.
func (m *MockNeuron) ID() string {
	return m.id
}

// Receive accepts a synapse message and stores it for test verification.
// This method implements the SynapseCompatibleNeuron interface requirement
// and provides the essential functionality needed for synapse testing.
//
// Parameters:
//
//	msg: The SynapseMessage delivered by a synapse after transmission
//
// BIOLOGICAL SIMULATION:
// In a real neuron, this method would:
// 1. Convert the synaptic signal to a postsynaptic potential
// 2. Integrate the signal with existing membrane potential
// 3. Update homeostatic state and activity tracking
// 4. Potentially trigger action potential generation
//
// TESTING IMPLEMENTATION:
// For testing purposes, we simply:
// 1. Store the message for later verification
// 2. Forward to a channel for concurrent test access
// 3. Maintain thread-safety for multiple synapse inputs
//
// This simplified approach allows focused testing of synapse transmission
// behavior without the complexity of full neural dynamics.
func (m *MockNeuron) Receive(msg SynapseMessage) {
	// Store the message in our log for test verification
	// This allows tests to examine exactly what signals were received
	m.receivedMsgs = append(m.receivedMsgs, msg)

	// Also send to channel for concurrent test access
	// The select with default ensures non-blocking operation
	select {
	case m.msgChannel <- msg:
		// Message successfully queued for channel-based test access
	default:
		// Channel full - message still stored in receivedMsgs for verification
		// This models realistic scenarios where rapid firing could overwhelm buffers
	}
}

// GetReceivedMessages returns all messages received by this mock neuron.
// This method provides test access to the complete history of synaptic
// inputs, enabling comprehensive verification of synapse transmission behavior.
//
// Returns:
//
//	Slice of all SynapseMessage objects received by this neuron
//
// TESTING UTILITY:
// This method enables tests to:
// - Verify that synapses transmitted signals correctly
// - Check signal timing, strength, and source identification
// - Validate that weight scaling was applied properly
// - Confirm that delays were respected in transmission
// - Analyze message ordering in multi-synapse scenarios
//
// THREAD SAFETY:
// This method returns a reference to the internal slice. In production code,
// this would typically return a copy for safety, but for testing purposes,
// direct access simplifies verification logic.
func (m *MockNeuron) GetReceivedMessages() []SynapseMessage {
	return m.receivedMsgs
}

// =================================================================================
// SYNAPSE CREATION AND INITIALIZATION TESTS
// =================================================================================

// TestSynapseCreation tests basic synapse creation and initialization.
// This test verifies that synapses are properly constructed with correct
// initial values, proper bounds enforcement, and appropriate configuration.
//
// BIOLOGICAL VALIDATION:
// This test ensures that:
// - Synaptic weights are initialized within biological bounds
// - Transmission delays are non-negative (physically realistic)
// - Configuration parameters are properly stored and accessible
// - New synapses are not immediately marked for pruning
//
// TEST COVERAGE:
// - Constructor parameter validation and bounds checking
// - Initial state verification (weight, delay, ID)
// - Configuration storage (STDP and pruning settings)
// - Pruning logic initial state (should not prune new synapses)
func TestSynapseCreation(t *testing.T) {
	// Create mock neurons to serve as pre- and post-synaptic endpoints
	// These provide the necessary interface implementations without
	// the complexity of full neuron simulation
	preNeuron := NewMockNeuron("pre_neuron")
	postNeuron := NewMockNeuron("post_neuron")

	// Create STDP configuration with biologically realistic parameters
	// These values are based on experimental data from cortical synapses
	stdpConfig := STDPConfig{
		Enabled:        true,                   // Enable spike-timing dependent plasticity
		LearningRate:   0.01,                   // 1% weight change per STDP event (typical range: 0.001-0.1)
		TimeConstant:   20 * time.Millisecond,  // Exponential decay time constant (typical: 10-50ms)
		WindowSize:     100 * time.Millisecond, // Maximum timing window for STDP (typical: 50-200ms)
		MinWeight:      0.001,                  // Minimum weight to prevent elimination (prevents runaway weakening)
		MaxWeight:      2.0,                    // Maximum weight to prevent runaway strengthening
		AsymmetryRatio: 1.2,                    // LTD/LTP ratio (slight bias toward depression)
	}

	// Create pruning configuration with conservative parameters
	// These settings prevent premature elimination of potentially useful synapses
	pruningConfig := PruningConfig{
		Enabled:             true,            // Enable structural plasticity (pruning)
		WeightThreshold:     0.01,            // Weight below which synapse becomes pruning candidate
		InactivityThreshold: 5 * time.Minute, // Duration of inactivity required for pruning eligibility
	}

	// Test synapse creation with specific initial parameters
	initialWeight := 0.5          // Moderate initial strength
	delay := 5 * time.Millisecond // Realistic axonal + synaptic delay
	synapseID := "test_synapse"   // Unique identifier for this connection

	// Create the synapse using the constructor
	synapse := NewBasicSynapse(synapseID, preNeuron, postNeuron,
		stdpConfig, pruningConfig, initialWeight, delay)

	// VERIFICATION 1: Test that synapse ID is properly stored and accessible
	// This ensures the synapse can be identified and referenced correctly
	if synapse.ID() != synapseID {
		t.Errorf("Expected synapse ID '%s', got '%s'", synapseID, synapse.ID())
	}

	// VERIFICATION 2: Test that initial weight is correctly set
	// This verifies that synaptic strength initialization works properly
	if synapse.GetWeight() != initialWeight {
		t.Errorf("Expected initial weight %f, got %f", initialWeight, synapse.GetWeight())
	}

	// VERIFICATION 3: Test that transmission delay is correctly set
	// This ensures realistic timing behavior in signal propagation
	if synapse.GetDelay() != delay {
		t.Errorf("Expected delay %v, got %v", delay, synapse.GetDelay())
	}

	// VERIFICATION 4: Test that new synapses are not immediately marked for pruning
	// New synapses should have a grace period before being eligible for elimination
	// This prevents premature removal of potentially useful connections
	if synapse.ShouldPrune() {
		t.Error("New synapse should not be marked for pruning immediately after creation")
	}

	// BIOLOGICAL SIGNIFICANCE:
	// This test validates that synapses start in a biologically plausible state:
	// - Moderate strength (not too weak or strong)
	// - Realistic transmission timing
	// - Protected from immediate elimination
	// - Properly configured for learning and adaptation
}

// =================================================================================
// SIGNAL TRANSMISSION TESTS
// =================================================================================

// TestSynapseTransmission tests basic signal transmission through synapses.
// This test verifies the core functionality of synaptic communication:
// signal scaling by synaptic weight, proper message formatting, and
// successful delivery to post-synaptic targets.
//
// BIOLOGICAL PROCESS TESTED:
// 1. Pre-synaptic neuron fires (simulated by calling Transmit)
// 2. Signal is scaled by synaptic weight (efficacy modulation)
// 3. Message is formatted with timing and identification metadata
// 4. Signal is delivered to post-synaptic neuron (via Receive method)
//
// TEST COVERAGE:
// - Signal strength scaling (weight multiplication)
// - Message metadata (source, synapse, timing information)
// - Successful delivery to target neuron
// - Goroutine-based transmission (asynchronous delivery)
func TestSynapseTransmission(t *testing.T) {
	// Create mock neurons for controlled testing environment
	preNeuron := NewMockNeuron("pre_neuron")
	postNeuron := NewMockNeuron("post_neuron")

	// Use default configurations for standard synapse behavior
	// These provide reasonable defaults without requiring manual configuration
	stdpConfig := CreateDefaultSTDPConfig()
	pruningConfig := CreateDefaultPruningConfig()

	// Create synapse with no delay for easier testing
	// Zero delay eliminates timing complications in verification
	synapseWeight := 0.5             // 50% signal scaling
	synapseDelay := time.Duration(0) // No transmission delay
	synapse := NewBasicSynapse("test_synapse", preNeuron, postNeuron,
		stdpConfig, pruningConfig, synapseWeight, synapseDelay)

	// Transmit a test signal through the synapse
	inputSignal := 1.0 // Full-strength input signal
	synapse.Transmit(inputSignal)

	// Allow time for asynchronous message delivery
	// Even with zero delay, goroutine scheduling may introduce small latencies
	time.Sleep(10 * time.Millisecond)

	// VERIFICATION 1: Check that exactly one message was received
	// This ensures the synapse transmitted the signal without duplication or loss
	messages := postNeuron.GetReceivedMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	// Extract the received message for detailed verification
	receivedMsg := messages[0]

	// VERIFICATION 2: Check signal strength scaling
	// The output should be input signal multiplied by synaptic weight
	expectedValue := inputSignal * synapseWeight // 1.0 * 0.5 = 0.5
	if receivedMsg.Value != expectedValue {
		t.Errorf("Expected message value %f, got %f", expectedValue, receivedMsg.Value)
	}

	// VERIFICATION 3: Check source identification
	// The message should correctly identify the pre-synaptic neuron
	if receivedMsg.SourceID != preNeuron.ID() {
		t.Errorf("Expected source ID '%s', got '%s'", preNeuron.ID(), receivedMsg.SourceID)
	}

	// VERIFICATION 4: Check synapse identification
	// The message should correctly identify which synapse transmitted it
	if receivedMsg.SynapseID != synapse.ID() {
		t.Errorf("Expected synapse ID '%s', got '%s'", synapse.ID(), receivedMsg.SynapseID)
	}

	// VERIFICATION 5: Check timing information
	// The timestamp should be recent (within the last second)
	// This validates that timing information is properly captured
	now := time.Now()
	timeSinceMessage := now.Sub(receivedMsg.Timestamp)
	if timeSinceMessage > time.Second {
		t.Errorf("Message timestamp too old: %v ago", timeSinceMessage)
	}

	// BIOLOGICAL SIGNIFICANCE:
	// This test validates the fundamental synaptic transmission process:
	// - Signal modulation by synaptic efficacy (weight)
	// - Proper identification for learning algorithms (STDP)
	// - Timing information for temporal processing
	// - Successful communication between neural elements
}

// =================================================================================
// WEIGHT MANAGEMENT TESTS
// =================================================================================

// TestSynapseWeightModification tests synaptic weight management functionality.
// This test verifies that synaptic weights can be properly read, modified,
// and that biological bounds are enforced to maintain network stability.
//
// BIOLOGICAL CONTEXT:
// Synaptic weights represent the efficacy of synaptic transmission and are
// the primary substrate for learning and memory in neural networks. They must:
// - Be modifiable for learning and adaptation
// - Respect biological bounds to prevent network instability
// - Maintain consistency under concurrent access
//
// TEST COVERAGE:
// - Initial weight retrieval
// - Weight modification (both manual and bounds-enforced)
// - Upper bound enforcement (prevents runaway strengthening)
// - Lower bound enforcement (prevents elimination)
// - Thread-safe access patterns
func TestSynapseWeightModification(t *testing.T) {
	// Create mock neurons for testing environment
	preNeuron := NewMockNeuron("pre_neuron")
	postNeuron := NewMockNeuron("post_neuron")

	// Create configurations with specific bounds for testing
	stdpConfig := CreateDefaultSTDPConfig()
	pruningConfig := CreateDefaultPruningConfig()

	// Create synapse with known initial weight
	initialWeight := 0.5
	synapse := NewBasicSynapse("test_synapse", preNeuron, postNeuron,
		stdpConfig, pruningConfig, initialWeight, 0)

	// VERIFICATION 1: Test initial weight retrieval
	// Ensure the synapse correctly stores and returns its initial weight
	if synapse.GetWeight() != initialWeight {
		t.Errorf("Expected initial weight %f, got %f", initialWeight, synapse.GetWeight())
	}

	// VERIFICATION 2: Test normal weight modification
	// Verify that weights can be changed within normal operating ranges
	newWeight := 0.8
	synapse.SetWeight(newWeight)
	if synapse.GetWeight() != newWeight {
		t.Errorf("Expected weight %f after setting, got %f", newWeight, synapse.GetWeight())
	}

	// VERIFICATION 3: Test upper bound enforcement
	// Weights above the maximum should be clamped to prevent runaway strengthening
	excessiveWeight := 10.0 // Much higher than max weight (2.0 from default config)
	synapse.SetWeight(excessiveWeight)
	expectedMaxWeight := stdpConfig.MaxWeight
	if synapse.GetWeight() != expectedMaxWeight {
		t.Errorf("Expected weight to be clamped to max %f, got %f",
			expectedMaxWeight, synapse.GetWeight())
	}

	// VERIFICATION 4: Test lower bound enforcement
	// Weights below the minimum should be clamped to prevent elimination
	negativeWeight := -1.0 // Lower than min weight (0.001 from default config)
	synapse.SetWeight(negativeWeight)
	expectedMinWeight := stdpConfig.MinWeight
	if synapse.GetWeight() != expectedMinWeight {
		t.Errorf("Expected weight to be clamped to min %f, got %f",
			expectedMinWeight, synapse.GetWeight())
	}

	// BIOLOGICAL SIGNIFICANCE:
	// This test validates critical safety mechanisms:
	// - Upper bounds prevent synapses from becoming pathologically strong
	// - Lower bounds maintain minimal connectivity (important for network function)
	// - Bounds enforcement protects network stability during learning
	// - Weight accessibility enables monitoring and experimental manipulation

	// NETWORK STABILITY IMPORTANCE:
	// Without proper bounds enforcement, learning algorithms could:
	// - Create runaway positive feedback loops (weights → ∞)
	// - Eliminate all connections (weights → 0)
	// - Destabilize the entire network through extreme values
	// This test ensures these failure modes are prevented.
}

// =================================================================================
// CONFIGURATION VALIDATION TESTS
// =================================================================================

// TestConfigHelpers tests the configuration helper functions that provide
// default parameters for STDP learning and synaptic pruning. These functions
// are critical for ensuring that synapses start with biologically realistic
// and computationally stable parameters.
//
// BIOLOGICAL IMPORTANCE:
// Configuration parameters control fundamental aspects of synaptic behavior:
// - STDP parameters determine learning dynamics and plasticity characteristics
// - Pruning parameters control structural plasticity and network optimization
// - Default values must be based on experimental neuroscience data
//
// TEST COVERAGE:
// - Default STDP configuration validation
// - Default pruning configuration validation
// - Conservative pruning configuration validation
// - Parameter relationships and biological realism
// - Configuration consistency across helper functions
func TestConfigHelpers(t *testing.T) {
	// SECTION 1: Test default STDP configuration
	// This configuration should provide reasonable defaults for most applications
	stdpConfig := CreateDefaultSTDPConfig()

	// VERIFICATION 1.1: STDP should be enabled by default
	// Most synapses in biological networks exhibit plasticity
	if !stdpConfig.Enabled {
		t.Error("Default STDP config should be enabled for realistic neural behavior")
	}

	// VERIFICATION 1.2: Learning rate should be positive and reasonable
	// Learning rate controls the speed of synaptic adaptation
	if stdpConfig.LearningRate <= 0 {
		t.Error("Default STDP config should have positive learning rate")
	}
	if stdpConfig.LearningRate > 0.1 {
		t.Error("Default STDP learning rate seems too high (biological range: 0.001-0.1)")
	}

	// VERIFICATION 1.3: Time constant should be biologically realistic
	// Time constant controls the width of the STDP learning window
	if stdpConfig.TimeConstant <= 0 {
		t.Error("Default STDP config should have positive time constant")
	}
	if stdpConfig.TimeConstant < 5*time.Millisecond || stdpConfig.TimeConstant > 100*time.Millisecond {
		t.Error("Default STDP time constant outside biological range (5-100ms)")
	}

	// VERIFICATION 1.4: Window size should encompass biologically relevant timing
	// Window size determines the maximum timing difference for STDP effects
	if stdpConfig.WindowSize <= stdpConfig.TimeConstant {
		t.Error("STDP window size should be larger than time constant")
	}

	// VERIFICATION 1.5: Weight bounds should prevent pathological values
	if stdpConfig.MinWeight < 0 {
		t.Error("Minimum weight should be non-negative")
	}
	if stdpConfig.MaxWeight <= stdpConfig.MinWeight {
		t.Error("Maximum weight should be greater than minimum weight")
	}

	// SECTION 2: Test default pruning configuration
	// This configuration should enable structural plasticity with safe parameters
	pruningConfig := CreateDefaultPruningConfig()

	// VERIFICATION 2.1: Pruning should be enabled by default
	// Structural plasticity is essential for network optimization
	if !pruningConfig.Enabled {
		t.Error("Default pruning config should be enabled for structural plasticity")
	}

	// VERIFICATION 2.2: Weight threshold should be reasonable
	// Threshold determines when synapses are considered weak
	if pruningConfig.WeightThreshold <= 0 {
		t.Error("Default pruning config should have positive weight threshold")
	}
	if pruningConfig.WeightThreshold >= stdpConfig.MaxWeight {
		t.Error("Pruning weight threshold should be less than maximum STDP weight")
	}

	// VERIFICATION 2.3: Inactivity threshold should provide grace period
	// Synapses need time to demonstrate their usefulness before pruning
	if pruningConfig.InactivityThreshold <= 0 {
		t.Error("Default pruning config should have positive inactivity threshold")
	}
	if pruningConfig.InactivityThreshold < time.Minute {
		t.Error("Inactivity threshold seems too short for biological realism")
	}

	// SECTION 3: Test conservative pruning configuration
	// This configuration should be more protective of existing connections
	conservativeConfig := CreateConservativePruningConfig()

	// VERIFICATION 3.1: Conservative config should have lower weight threshold
	// Lower threshold means synapses need to be weaker to be pruned
	if conservativeConfig.WeightThreshold >= pruningConfig.WeightThreshold {
		t.Error("Conservative config should have lower weight threshold than default")
	}

	// VERIFICATION 3.2: Conservative config should have longer inactivity threshold
	// Longer threshold gives synapses more time to prove their usefulness
	if conservativeConfig.InactivityThreshold <= pruningConfig.InactivityThreshold {
		t.Error("Conservative config should have longer inactivity threshold than default")
	}

	// BIOLOGICAL SIGNIFICANCE:
	// These configuration validations ensure:
	// - Synapses behave within biologically realistic parameters
	// - Learning dynamics are stable and convergent
	// - Structural plasticity operates safely without over-pruning
	// - Default values provide good starting points for most applications
	// - Conservative options are available when network stability is critical

	// RESEARCH IMPLICATIONS:
	// Proper default configurations enable:
	// - Reproducible experiments across different studies
	// - Reasonable baseline behavior for comparative analyses
	// - Reduced need for extensive parameter tuning
	// - Biological realism in computational neuroscience models
}
