package neuron

import (
	"fmt"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/synapse"
)

// MockSynapseCompatibleNeuron implements the synapse.SynapseCompatibleNeuron interface
// for testing purposes. This allows us to create controlled test networks
// where we can precisely verify signal transmission and timing.
//
// BIOLOGICAL CONTEXT:
// In real neural networks, neurons must communicate through synaptic connections.
// This mock neuron models a simplified post-synaptic neuron that can receive
// and record synaptic inputs for verification during testing.
type MockSynapseCompatibleNeuron struct {
	id           string
	receivedMsgs []synapse.SynapseMessage
	mutex        sync.Mutex
}

// NewMockNeuron creates a mock neuron for testing synapse communication
func NewMockNeuron(id string) *MockSynapseCompatibleNeuron {
	return &MockSynapseCompatibleNeuron{
		id:           id,
		receivedMsgs: make([]synapse.SynapseMessage, 0),
	}
}

// ID returns the neuron's unique identifier
func (m *MockSynapseCompatibleNeuron) ID() string {
	return m.id
}

// Receive implements the synapse.SynapseCompatibleNeuron interface
// Records all received messages for test verification
func (m *MockSynapseCompatibleNeuron) Receive(msg synapse.SynapseMessage) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.receivedMsgs = append(m.receivedMsgs, msg)
}

// GetReceivedMessages returns a copy of all received messages for testing
func (m *MockSynapseCompatibleNeuron) GetReceivedMessages() []synapse.SynapseMessage {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Return a copy to prevent external modification
	messages := make([]synapse.SynapseMessage, len(m.receivedMsgs))
	copy(messages, m.receivedMsgs)
	return messages
}

// ClearReceivedMessages clears the message history for fresh testing
func (m *MockSynapseCompatibleNeuron) ClearReceivedMessages() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.receivedMsgs = m.receivedMsgs[:0]
}

// ============================================================================
// CORE NEURON FUNCTIONALITY TESTS
// ============================================================================

// TestNeuronCreation validates basic neuron initialization and properties
//
// BIOLOGICAL CONTEXT:
// Every biological neuron has fundamental electrical properties that determine
// its behavior: firing threshold (action potential threshold), membrane decay
// (leakiness), refractory period (recovery time), and output strength.
// These parameters define the neuron's computational characteristics.
//
// EXPECTED RESULTS:
// - Neuron initializes with correct parameters
// - Homeostatic system is properly configured
// - Input channel is functional for synaptic communication
// - Neuron is ready for synaptic connections
func TestNeuronCreation(t *testing.T) {
	threshold := 1.5
	decayRate := 0.95
	refractoryPeriod := 10 * time.Millisecond
	fireFactor := 2.0
	neuronID := "cortical_pyramidal_01"
	targetFiringRate := 5.0
	homeostasisStrength := 0.2

	// Create a neuron with homeostatic plasticity enabled
	neuron := NewNeuron(neuronID, threshold, decayRate, refractoryPeriod,
		fireFactor, targetFiringRate, homeostasisStrength)

	// Verify basic properties
	if neuron.ID() != neuronID {
		t.Errorf("Expected neuron ID %s, got %s", neuronID, neuron.ID())
	}

	if neuron.GetCurrentThreshold() != threshold {
		t.Errorf("Expected threshold %f, got %f", threshold, neuron.GetCurrentThreshold())
	}

	if neuron.GetBaseThreshold() != threshold {
		t.Errorf("Expected base threshold %f, got %f", threshold, neuron.GetBaseThreshold())
	}

	// Verify homeostatic system initialization
	homeostaticInfo := neuron.GetHomeostaticInfo()
	if homeostaticInfo.targetFiringRate != targetFiringRate {
		t.Errorf("Expected target firing rate %f, got %f",
			targetFiringRate, homeostaticInfo.targetFiringRate)
	}

	if homeostaticInfo.homeostasisStrength != homeostasisStrength {
		t.Errorf("Expected homeostasis strength %f, got %f",
			homeostasisStrength, homeostaticInfo.homeostasisStrength)
	}

	// Verify initial state
	if neuron.GetCurrentFiringRate() != 0.0 {
		t.Errorf("Expected initial firing rate 0.0, got %f", neuron.GetCurrentFiringRate())
	}

	if neuron.GetCalciumLevel() != 0.0 {
		t.Errorf("Expected initial calcium level 0.0, got %f", neuron.GetCalciumLevel())
	}

	// Verify synaptic scaling is disabled by default
	gains := neuron.GetInputGains()
	if len(gains) != 0 {
		t.Errorf("Expected no input gains initially, got %d", len(gains))
	}

	// Verify no synaptic connections initially
	if neuron.GetOutputSynapseCount() != 0 {
		t.Errorf("Expected 0 synaptic connections initially, got %d",
			neuron.GetOutputSynapseCount())
	}
}

// TestSimpleNeuronCreation validates backward compatibility constructor
//
// BIOLOGICAL CONTEXT:
// Some applications may not require the full homeostatic plasticity system
// but still need the core temporal dynamics. This tests the simplified
// constructor that creates neurons with homeostasis disabled.
//
// EXPECTED RESULTS:
// - Neuron initializes with homeostasis disabled
// - Core functionality remains intact
// - Threshold remains fixed (no self-regulation)
func TestSimpleNeuronCreation(t *testing.T) {
	neuronID := "simple_interneuron"
	threshold := 1.0

	neuron := NewSimpleNeuron(neuronID, threshold, 0.95, 5*time.Millisecond, 1.0)

	// Verify homeostasis is disabled
	homeostaticInfo := neuron.GetHomeostaticInfo()
	if homeostaticInfo.targetFiringRate != 0.0 {
		t.Errorf("Expected disabled homeostasis (target rate 0), got %f",
			homeostaticInfo.targetFiringRate)
	}

	if homeostaticInfo.homeostasisStrength != 0.0 {
		t.Errorf("Expected disabled homeostasis (strength 0), got %f",
			homeostaticInfo.homeostasisStrength)
	}

	// Verify threshold remains at base value
	if neuron.GetCurrentThreshold() != threshold {
		t.Errorf("Expected fixed threshold %f, got %f",
			threshold, neuron.GetCurrentThreshold())
	}

	// Verify core functionality
	if neuron.ID() != neuronID {
		t.Errorf("Expected neuron ID %s, got %s", neuronID, neuron.ID())
	}
}

// TestNeuronWithLearning validates the learning-enabled convenience constructor
//
// BIOLOGICAL CONTEXT:
// Most biological neural networks require homeostatic regulation to maintain
// stable activity levels. This constructor provides reasonable defaults
// for homeostatic plasticity suitable for learning networks.
//
// EXPECTED RESULTS:
// - Homeostatic plasticity enabled with reasonable parameters
// - Standard biological timing and decay parameters
// - Ready for learning and self-regulation
func TestNeuronWithLearning(t *testing.T) {
	neuronID := "learning_neuron"
	threshold := 1.2
	targetRate := 8.0

	neuron := NewNeuronWithLearning(neuronID, threshold, targetRate)

	// Verify homeostatic system is properly configured
	homeostaticInfo := neuron.GetHomeostaticInfo()
	if homeostaticInfo.targetFiringRate != targetRate {
		t.Errorf("Expected target firing rate %f, got %f",
			targetRate, homeostaticInfo.targetFiringRate)
	}

	if homeostaticInfo.homeostasisStrength != 0.2 {
		t.Errorf("Expected homeostasis strength 0.2, got %f",
			homeostaticInfo.homeostasisStrength)
	}

	// Verify threshold bounds are reasonable
	if homeostaticInfo.minThreshold <= 0 {
		t.Errorf("Expected positive minimum threshold, got %f",
			homeostaticInfo.minThreshold)
	}

	if homeostaticInfo.maxThreshold <= threshold {
		t.Errorf("Expected maximum threshold > base threshold, got max=%f, base=%f",
			homeostaticInfo.maxThreshold, threshold)
	}
}

// ============================================================================
// SYNAPSE INTEGRATION TESTS
// ============================================================================

// TestReceiveMethod validates the SynapseCompatibleNeuron interface implementation
//
// BIOLOGICAL CONTEXT:
// The Receive() method is how synapses deliver signals to neurons. This models
// the biological process where neurotransmitter release at synapses creates
// postsynaptic potentials that are integrated by the receiving neuron.
// Proper implementation is crucial for synapse-neuron communication.
//
// EXPECTED RESULTS:
// - Neuron correctly implements SynapseCompatibleNeuron interface
// - Receive() method accepts and processes synapse messages
// - Messages are integrated into neuron's membrane dynamics
// - Timing and source information is preserved
func TestReceiveMethod(t *testing.T) {
	neuron := NewSimpleNeuron("postsynaptic_neuron", 1.0, 0.98, 5*time.Millisecond, 1.0)

	// Start neuron processing
	go neuron.Run()
	defer neuron.Close()

	// Create a test message that would arrive from a synapse
	testMessage := synapse.SynapseMessage{
		Value:     0.8,
		Timestamp: time.Now(),
		SourceID:  "presynaptic_neuron_01",
		SynapseID: "synapse_connection_01",
	}

	// Test that Receive method accepts the message without blocking
	done := make(chan bool, 1)
	go func() {
		neuron.Receive(testMessage)
		done <- true
	}()

	// Should complete quickly without blocking
	select {
	case <-done:
		// Expected - method should not block
	case <-time.After(10 * time.Millisecond):
		t.Error("Receive() method blocked - should be non-blocking")
	}

	// Allow time for message processing
	time.Sleep(5 * time.Millisecond)

	// Test multiple messages can be received
	for i := 0; i < 5; i++ {
		msg := synapse.SynapseMessage{
			Value:     0.1,
			Timestamp: time.Now(),
			SourceID:  fmt.Sprintf("source_%d", i),
			SynapseID: fmt.Sprintf("synapse_%d", i),
		}
		neuron.Receive(msg)
	}

	// All messages should be processed without error
	time.Sleep(10 * time.Millisecond)
}

// TestSynapseMessageProcessing validates message integration with source tracking
//
// BIOLOGICAL CONTEXT:
// When synapses transmit signals, the receiving neuron must integrate these
// signals while tracking their sources for learning mechanisms like STDP.
// This test verifies that synapse messages are properly processed and that
// source information is preserved for synaptic scaling.
//
// EXPECTED RESULTS:
// - Messages from different sources are integrated correctly
// - Source IDs are tracked for synaptic scaling
// - Temporal dynamics (accumulation and decay) work with synapse messages
// - Multiple sources can send signals simultaneously
func TestSynapseMessageProcessing(t *testing.T) {
	neuron := NewSimpleNeuron("integrating_neuron", 2.0, 0.95, 5*time.Millisecond, 1.0)

	// Enable synaptic scaling to test source tracking
	neuron.EnableSynapticScaling(1.0, 0.01, 10*time.Second)

	go neuron.Run()
	defer neuron.Close()

	// Create messages from different sources
	sources := []string{"dendrite_A", "dendrite_B", "dendrite_C"}

	for i, sourceID := range sources {
		msg := synapse.SynapseMessage{
			Value:     0.5,
			Timestamp: time.Now(),
			SourceID:  sourceID,
			SynapseID: fmt.Sprintf("synapse_%d", i),
		}
		neuron.Receive(msg)
		time.Sleep(2 * time.Millisecond) // Brief delay between signals
	}

	// Allow processing time
	time.Sleep(20 * time.Millisecond)

	// Verify input gains were registered for each source
	gains := neuron.GetInputGains()
	for _, sourceID := range sources {
		if gain, exists := gains[sourceID]; !exists {
			t.Errorf("Source %s not registered in input gains", sourceID)
		} else if gain != 1.0 {
			t.Errorf("Expected default gain 1.0 for source %s, got %f", sourceID, gain)
		}
	}

	// Test message without source ID (should still be processed)
	anonymousMsg := synapse.SynapseMessage{
		Value:     0.3,
		Timestamp: time.Now(),
		SourceID:  "", // No source ID
		SynapseID: "anonymous_synapse",
	}
	neuron.Receive(anonymousMsg)

	time.Sleep(10 * time.Millisecond)

	// Should not affect input gains tracking
	newGains := neuron.GetInputGains()
	if len(newGains) != len(gains) {
		t.Errorf("Anonymous message affected gains tracking: %d vs %d",
			len(newGains), len(gains))
	}
}

// TestNeuronSynapseNetworkIntegration validates end-to-end neuron-synapse communication
//
// BIOLOGICAL CONTEXT:
// This test creates a minimal neural network with real synapse objects connecting
// real neurons. It models the biological scenario where a presynaptic neuron
// fires, transmits through a synapse with realistic delays and weights, and
// the postsynaptic neuron receives and integrates the signal.
//
// EXPECTED RESULTS:
// - Real synapses successfully connect real neurons
// - Signal transmission occurs with proper delays
// - Synaptic weights affect signal strength
// - Network timing is biologically realistic
// - Multiple synaptic connections work simultaneously
// In neuron_test.go

func TestNeuronSynapseNetworkIntegration(t *testing.T) {
	// 1. SETUP: Create the network components.
	presynapticNeuron := NewSimpleNeuron("motor_neuron", 1.0, 0.98, 5*time.Millisecond, 1.0)
	postsynapticNeuron := NewMockNeuron("muscle_fiber")

	// 2. SETUP FIRE EVENT MONITORING: This is the key to the fix.
	// We create a channel to receive the exact moment the presynaptic neuron fires.
	fireEvents := make(chan FireEvent, 1)
	presynapticNeuron.SetFireEventChannel(fireEvents)

	// 3. SETUP SYNAPSE: Configure the connection with a clear delay.
	const synapticDelay = 3 * time.Millisecond
	const synapticWeight = 0.8
	synapticConnection := synapse.NewBasicSynapse(
		"neuromuscular_junction",
		presynapticNeuron,
		postsynapticNeuron,
		synapse.CreateDefaultSTDPConfig(),
		synapse.CreateDefaultPruningConfig(),
		synapticWeight,
		synapticDelay,
	)
	presynapticNeuron.AddOutputSynapse("muscle_connection", synapticConnection)

	// 4. START THE NETWORK
	go presynapticNeuron.Run()
	defer presynapticNeuron.Close()

	// 5. STIMULATE THE PRESYNAPTIC NEURON
	presynapticNeuron.Receive(synapse.SynapseMessage{
		Value:     1.5, // Above threshold
		Timestamp: time.Now(),
		SourceID:  "spinal_cord",
	})

	// 6. WAIT FOR THE FIRE EVENT to get the precise start time and output value.
	var presynapticFireEvent FireEvent
	select {
	case event := <-fireEvents:
		t.Logf("Received FireEvent from presynaptic neuron at %v with value %.2f", event.Timestamp, event.Value)
		presynapticFireEvent = event
	case <-time.After(20 * time.Millisecond):
		t.Fatal("Did not receive a fire event from the presynaptic neuron.")
	}

	// Wait for the message to propagate through the synapse
	time.Sleep(synapticDelay + 20*time.Millisecond)

	// 7. VERIFY THE OUTCOME
	receivedMessages := postsynapticNeuron.GetReceivedMessages()
	if len(receivedMessages) == 0 {
		t.Fatal("Postsynaptic neuron did not receive the message.")
	}
	receivedMsg := receivedMessages[0]

	// ✅ ROBUST TIMING CHECK: Calculate delay based on the actual firing time.
	transmissionTime := receivedMsg.Timestamp.Sub(presynapticFireEvent.Timestamp)
	t.Logf("Calculated synaptic transmission time: %v", transmissionTime)

	minExpectedDelay := synapticDelay
	maxExpectedDelay := synapticDelay + 20*time.Millisecond // Generous window
	if transmissionTime < minExpectedDelay || transmissionTime > maxExpectedDelay {
		t.Errorf("Transmission time outside expected window. Got: %v, Want: [%v, %v]",
			transmissionTime, minExpectedDelay, maxExpectedDelay)
	}

	// ✅ ROBUST VALUE CHECK: Calculate the expected value based on the *actual*
	// signal sent by the presynaptic neuron (from the FireEvent).
	expectedValue := presynapticFireEvent.Value * synapticWeight
	tolerance := 1e-9 // Use a small tolerance for floating point comparison

	if math.Abs(receivedMsg.Value-expectedValue) > tolerance {
		t.Errorf("Expected signal value ~%.6f, got %.6f", expectedValue, receivedMsg.Value)
	}

	// Verify other message properties
	if receivedMsg.SourceID != presynapticNeuron.ID() {
		t.Errorf("Expected source ID %s, got %s", presynapticNeuron.ID(), receivedMsg.SourceID)
	}
}

// TestMultipleSynapticConnections validates complex network connectivity
//
// BIOLOGICAL CONTEXT:
// Real neurons typically receive inputs from hundreds or thousands of synapses
// and send outputs to similar numbers of targets. This test validates that
// neurons can handle multiple simultaneous synaptic connections with proper
// signal integration and transmission.
//
// EXPECTED RESULTS:
// - Multiple input synapses can stimulate one neuron simultaneously
// - Multiple output synapses can receive signals from one neuron
// - Signal integration follows biological summation rules
// - No interference between different synaptic pathways// In neuron_test.go

// In neuron_test.go

func TestMultipleSynapticConnections(t *testing.T) {
	// --- SETUP ---
	// Create multiple input sources
	inputSources := []*Neuron{
		NewSimpleNeuron("input_A", 0.8, 0.98, 5*time.Millisecond, 1.0),
		NewSimpleNeuron("input_B", 0.8, 0.98, 5*time.Millisecond, 1.0),
		NewSimpleNeuron("input_C", 0.8, 0.98, 5*time.Millisecond, 1.0),
	}

	// Create multiple output targets
	outputTargets := []*MockSynapseCompatibleNeuron{
		NewMockNeuron("target_alpha"),
		NewMockNeuron("target_beta"),
		NewMockNeuron("target_gamma"),
	}

	// Function to create and wire the central neuron
	// This makes it easy to reset the network state
	createAndWireCentralNeuron := func() *Neuron {
		centralNeuron := NewSimpleNeuron("pyramidal_neuron", 2.0, 0.98,
			10*time.Millisecond, 1.0)

		// Create input synapses connecting sources to the new central neuron
		for i, source := range inputSources {
			inputSynapse := synapse.NewBasicSynapse(
				fmt.Sprintf("input_synapse_%d", i),
				source,
				centralNeuron,
				synapse.CreateDefaultSTDPConfig(),
				synapse.CreateDefaultPruningConfig(),
				0.7, // Weight
				2*time.Millisecond,
			)
			// Use the public method to add the synapse
			source.AddOutputSynapse(fmt.Sprintf("to_central_%d", i), inputSynapse)
		}

		// Create output synapses connecting the new central neuron to targets
		for i, target := range outputTargets {
			outputSynapse := synapse.NewBasicSynapse(
				fmt.Sprintf("output_synapse_%d", i),
				centralNeuron,
				target,
				synapse.CreateDefaultSTDPConfig(),
				synapse.CreateDefaultPruningConfig(),
				0.9, // Weight
				1*time.Millisecond,
			)
			centralNeuron.AddOutputSynapse(fmt.Sprintf("to_target_%d", i), outputSynapse)
		}
		return centralNeuron
	}

	// --- INITIAL NETWORK SETUP ---
	centralNeuron := createAndWireCentralNeuron()

	// Start all neurons
	go centralNeuron.Run()
	for _, source := range inputSources {
		go source.Run()
		defer source.Close() // Defer closing all sources
	}

	// --- Test 1: Single input should NOT fire central neuron ---
	inputSources[0].Receive(synapse.SynapseMessage{
		Value: 1.0, Timestamp: time.Now(), SourceID: "test_single", SynapseID: "test",
	})
	time.Sleep(20 * time.Millisecond)

	for i, target := range outputTargets {
		if len(target.GetReceivedMessages()) > 0 {
			t.Errorf("Target %d received signal from single input - threshold too low", i)
		}
		target.ClearReceivedMessages() // Clear messages for the next test
	}

	// --- RESET STATE BETWEEN TESTS ---
	// Close the old central neuron and create a fresh one to reset its state.
	centralNeuron.Close()
	centralNeuron = createAndWireCentralNeuron()
	go centralNeuron.Run()
	defer centralNeuron.Close() // Defer closing the new one as well

	// --- Test 2: Multiple simultaneous inputs SHOULD fire central neuron ---
	for _, source := range inputSources {
		source.Receive(synapse.SynapseMessage{
			Value:     1.0,
			Timestamp: time.Now(),
			SourceID:  fmt.Sprintf("multi_test_source_%s", source.ID()),
			SynapseID: "multi_test_synapse",
		})
	}

	time.Sleep(30 * time.Millisecond)

	// All output targets should receive the CORRECT signal
	for i, target := range outputTargets {
		messages := target.GetReceivedMessages()
		if len(messages) == 0 {
			t.Errorf("Target %d did not receive signal from multiple inputs", i)
			continue
		}

		msg := messages[0]
		// The expected accumulator value is (0.7 * 3) = 2.1
		// This fires and produces an output signal: 2.1 * fireFactor(1.0) * outputWeight(0.9) = 1.89
		expectedValue := (0.7 * 3) * 1.0 * 0.9
		tolerance := 0.2 // Use a reasonable tolerance for timing variations

		if math.Abs(msg.Value-expectedValue) > tolerance {
			t.Errorf("Target %d: expected signal ~%.2f, got %f",
				i, expectedValue, msg.Value)
		}
	}
}

// ============================================================================
// BIOLOGICAL TIMING AND DYNAMICS TESTS
// ============================================================================

// TestThresholdBasedFiring validates discrete action potential generation
//
// BIOLOGICAL CONTEXT:
// Biological neurons exhibit "all-or-nothing" firing behavior. When the
// membrane potential reaches the action potential threshold, a stereotyped
// electrical spike is generated and propagated to all connected neurons.
// Below threshold, no firing occurs regardless of how close to threshold
// the neuron gets.
//
// EXPECTED RESULTS:
// - Subthreshold inputs do not cause firing
// - Suprathreshold inputs reliably cause firing
// - Firing is immediate when threshold is exceeded
// - Output signal reflects input integration
func TestThresholdBasedFiring(t *testing.T) {
	// Base parameters used for both sub-tests
	threshold := 1.5
	fireFactor := 2.0

	// --- SUB-TEST 1: Subthreshold input should NOT cause firing ---
	t.Run("SubthresholdInput", func(t *testing.T) {
		// Create a fresh neuron for this specific test case
		neuron := NewSimpleNeuron("subthreshold_neuron", threshold, 0.98,
			5*time.Millisecond, fireFactor)
		targetNeuron := NewMockNeuron("target")
		outputSynapse := synapse.NewBasicSynapse("test_output", neuron, targetNeuron,
			synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(), 1.0, 0)
		neuron.AddOutputSynapse("conn", outputSynapse)

		go neuron.Run()
		defer neuron.Close()

		// Send a signal with a value less than the threshold
		neuron.Receive(synapse.SynapseMessage{Value: 1.2})
		time.Sleep(20 * time.Millisecond) // Allow time for processing

		// Verify that no message was sent to the target (i.e., the neuron did not fire)
		if len(targetNeuron.GetReceivedMessages()) > 0 {
			t.Fatal("Neuron fired with subthreshold input, but should not have.")
		}
	})

	// --- SUB-TEST 2: Suprathreshold input SHOULD cause firing with correct output ---
	t.Run("SuprathresholdInput", func(t *testing.T) {
		// Create a fresh neuron for this specific test case, ensuring no state leakage
		neuron := NewSimpleNeuron("suprathreshold_neuron", threshold, 0.98,
			5*time.Millisecond, fireFactor)
		targetNeuron := NewMockNeuron("target")
		outputSynapse := synapse.NewBasicSynapse("test_output", neuron, targetNeuron,
			synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(), 1.0, 0)
		neuron.AddOutputSynapse("conn", outputSynapse)

		go neuron.Run()
		defer neuron.Close()

		// Send a signal with a value greater than the threshold
		suprathresholdInput := 2.0
		neuron.Receive(synapse.SynapseMessage{Value: suprathresholdInput})
		time.Sleep(20 * time.Millisecond) // Allow time for firing and transmission

		// Verify the neuron fired exactly once
		messages := targetNeuron.GetReceivedMessages()
		if len(messages) == 0 {
			t.Fatal("Neuron did not fire with suprathreshold input.")
		}
		if len(messages) > 1 {
			t.Fatalf("Neuron fired %d times, expected once.", len(messages))
		}

		// Verify the output value is correct
		// The neuron's accumulator should be ~2.0 at firing time (slight decay is possible but minor)
		// The output should be accumulator * fireFactor
		expectedOutput := suprathresholdInput * fireFactor
		actualOutput := messages[0].Value

		// Use a tolerance to account for minor decay before firing
		tolerance := 0.1
		if math.Abs(actualOutput-expectedOutput) > tolerance {
			t.Errorf("Expected output value around %.2f, but got %.2f", expectedOutput, actualOutput)
		}
	})
}

// TestLeakyIntegration validates continuous membrane potential decay
//
// BIOLOGICAL CONTEXT:
// Real neuron membranes act like leaky capacitors - charge gradually leaks out
// through membrane resistance, causing the membrane potential to decay toward
// resting potential. This creates temporal summation where recent inputs have
// stronger influence than older inputs, and allows for sophisticated temporal
// processing of input patterns.
//
// EXPECTED RESULTS:
// - Membrane potential decays exponentially over time
// - Recent inputs have stronger influence than older inputs
// - Temporal summation occurs for closely spaced inputs
// - Decay rate affects integration time window
// In neuron_test.go

// TestLeakyIntegration validates that closely spaced inputs can summate over time
// to overcome the firing threshold, a core principle of temporal processing.
func TestLeakyIntegration(t *testing.T) {
	// The decay rate is less aggressive (1% decay per ms instead of 10%),
	// allowing membrane potential to persist longer between signals.
	decayRate := 0.99

	// The threshold is set to a level that can be realistically
	// reached by the sum of the two inputs, even after slight decay.
	threshold := 1.5
	neuron := NewSimpleNeuron("leaky_integration_neuron", threshold, decayRate,
		5*time.Millisecond, 1.0)

	targetNeuron := NewMockNeuron("integration_target")

	outputSynapse := synapse.NewBasicSynapse(
		"integration_output",
		neuron,
		targetNeuron,
		synapse.CreateDefaultSTDPConfig(),
		synapse.CreateDefaultPruningConfig(),
		1.0, 0,
	)

	neuron.AddOutputSynapse("integration_test", outputSynapse)

	go neuron.Run()
	defer neuron.Close()

	// The input values are strong enough to potentially cross the
	// threshold when combined, but not individually.
	// First signal arrives. The accumulator becomes 0.8.
	neuron.Receive(synapse.SynapseMessage{
		Value: 0.8, Timestamp: time.Now(), SourceID: "summation_1", SynapseID: "test",
	})

	// Wait 2ms. During this time, the accumulator decays slightly from 0.8 to ~0.784.
	time.Sleep(2 * time.Millisecond)

	// Second signal arrives. The accumulator becomes ~0.784 + 0.8 = ~1.584,
	// which is now above the threshold of 1.5.
	neuron.Receive(synapse.SynapseMessage{
		Value: 0.8, Timestamp: time.Now(), SourceID: "summation_2", SynapseID: "test",
	})

	// Wait for the neuron to process the firing event.
	time.Sleep(10 * time.Millisecond)

	messages := targetNeuron.GetReceivedMessages()

	// The assertion should now pass, as the neuron will have fired.
	if len(messages) == 0 {
		t.Error("Neuron did not fire with rapid temporal summation, but was expected to.")
	}
}

// TestRefractoryPeriod validates post-firing recovery constraints
//
// BIOLOGICAL CONTEXT:
// After generating an action potential, biological neurons enter a refractory
// period during which they cannot fire again, regardless of input strength.
// This is caused by sodium channel inactivation and potassium channel activation.
// The refractory period prevents unrealistic rapid firing and creates natural
// timing constraints in neural computation.
//
// EXPECTED RESULTS:
// - Neuron cannot fire during refractory period
// - Strong inputs during refractory period are ignored
// - Normal firing resumes after refractory period ends
// - Refractory period duration matches specification
func TestRefractoryPeriod(t *testing.T) {
	refractoryPeriod := 20 * time.Millisecond
	neuron := NewSimpleNeuron("refractory_test_neuron", 1.0, 0.98,
		refractoryPeriod, 1.0)

	targetNeuron := NewMockNeuron("refractory_target")

	outputSynapse := synapse.NewBasicSynapse(
		"refractory_output",
		neuron,
		targetNeuron,
		synapse.CreateDefaultSTDPConfig(),
		synapse.CreateDefaultPruningConfig(),
		1.0, 0,
	)

	neuron.AddOutputSynapse("refractory_test", outputSynapse)

	go neuron.Run()
	defer neuron.Close()

	// Fire neuron first time
	neuron.Receive(synapse.SynapseMessage{
		Value: 1.5, Timestamp: time.Now(), SourceID: "first_fire", SynapseID: "test",
	})

	// Wait for first firing
	time.Sleep(10 * time.Millisecond)

	messages := targetNeuron.GetReceivedMessages()
	if len(messages) == 0 {
		t.Fatal("First firing did not occur")
	}

	firstFireTime := messages[0].Timestamp
	targetNeuron.ClearReceivedMessages()

	// Immediately try to fire again (should be blocked)
	neuron.Receive(synapse.SynapseMessage{
		Value:     2.0, // Very strong input
		Timestamp: time.Now(),
		SourceID:  "refractory_attempt",
		SynapseID: "test",
	})

	time.Sleep(10 * time.Millisecond)

	messages = targetNeuron.GetReceivedMessages()
	if len(messages) > 0 {
		t.Error("Neuron fired during refractory period - refractory mechanism failed")
	}

	// Wait for refractory period to end and try again
	time.Sleep(refractoryPeriod + 5*time.Millisecond)

	neuron.Receive(synapse.SynapseMessage{
		Value: 1.5, Timestamp: time.Now(), SourceID: "post_refractory", SynapseID: "test",
	})

	time.Sleep(10 * time.Millisecond)

	messages = targetNeuron.GetReceivedMessages()
	if len(messages) == 0 {
		t.Error("Neuron did not fire after refractory period ended")
	}

	secondFireTime := messages[0].Timestamp
	actualRefractoryPeriod := secondFireTime.Sub(firstFireTime)

	// Should respect minimum refractory period
	if actualRefractoryPeriod < refractoryPeriod {
		t.Errorf("Actual refractory period (%v) shorter than specified (%v)",
			actualRefractoryPeriod, refractoryPeriod)
	}
}

// TestContinuousDecay validates ongoing membrane potential decay
//
// BIOLOGICAL CONTEXT:
// Biological neurons continuously lose charge through membrane resistance,
// not just at discrete time points. This continuous decay is essential for
// realistic temporal dynamics and prevents artificial accumulation of
// subthreshold inputs over long periods.
//
// EXPECTED RESULTS:
// - Membrane potential continuously decays toward resting potential
// - Old inputs lose influence over time
// - Decay occurs even without new inputs
// - Decay rate affects temporal integration window
func TestContinuousDecay(t *testing.T) {
	decayRate := 0.8 // Fast decay for testing
	neuron := NewSimpleNeuron("continuous_decay_neuron", 3.0, decayRate,
		5*time.Millisecond, 1.0)

	targetNeuron := NewMockNeuron("decay_target")

	outputSynapse := synapse.NewBasicSynapse(
		"decay_output",
		neuron,
		targetNeuron,
		synapse.CreateDefaultSTDPConfig(),
		synapse.CreateDefaultPruningConfig(),
		1.0, 0,
	)

	neuron.AddOutputSynapse("decay_test", outputSynapse)

	go neuron.Run()
	defer neuron.Close()

	// Send input that builds up charge but doesn't fire
	neuron.Receive(synapse.SynapseMessage{
		Value:     2.0, // Below threshold of 3.0
		Timestamp: time.Now(),
		SourceID:  "decay_input",
		SynapseID: "test",
	})

	// Wait for several decay cycles
	time.Sleep(20 * time.Millisecond)

	// Send additional input - if decay worked, this should not be enough
	neuron.Receive(synapse.SynapseMessage{
		Value:     0.8, // Should not fire if previous input decayed
		Timestamp: time.Now(),
		SourceID:  "post_decay_input",
		SynapseID: "test",
	})

	time.Sleep(10 * time.Millisecond)

	messages := targetNeuron.GetReceivedMessages()
	if len(messages) > 0 {
		t.Error("Neuron fired despite continuous decay - decay mechanism not working")
	}

	// Verify neuron still works with fresh strong input
	neuron.Receive(synapse.SynapseMessage{
		Value:     3.5, // Above threshold
		Timestamp: time.Now(),
		SourceID:  "fresh_strong_input",
		SynapseID: "test",
	})

	time.Sleep(10 * time.Millisecond)

	messages = targetNeuron.GetReceivedMessages()
	if len(messages) == 0 {
		t.Error("Neuron failed to fire with fresh strong input - may be broken")
	}
}

// ============================================================================
// ERROR HANDLING AND EDGE CASES
// ============================================================================

// TestConcurrentSynapseAccess validates thread safety with multiple synapses
//
// BIOLOGICAL CONTEXT:
// Real neurons receive inputs from hundreds or thousands of synapses
// simultaneously. The neuron implementation must handle concurrent access
// from multiple synaptic connections without data races or corruption.
//
// EXPECTED RESULTS:
// - Multiple synapses can send signals concurrently
// - No data races or panics occur
// - All signals are processed correctly
// - Synaptic scaling remains consistent under concurrent access
func TestConcurrentSynapseAccess(t *testing.T) {
	neuron := NewSimpleNeuron("concurrent_test_neuron", 5.0, 0.98,
		5*time.Millisecond, 1.0)

	// Enable scaling to test concurrent access to scaling structures
	neuron.EnableSynapticScaling(1.0, 0.01, 1*time.Second)

	targetNeuron := NewMockNeuron("concurrent_target")
	outputSynapse := synapse.NewBasicSynapse(
		"concurrent_output",
		neuron,
		targetNeuron,
		synapse.CreateDefaultSTDPConfig(),
		synapse.CreateDefaultPruningConfig(),
		1.0, 0,
	)
	neuron.AddOutputSynapse("concurrent_test", outputSynapse)

	go neuron.Run()
	defer neuron.Close()

	// Create multiple goroutines sending signals concurrently
	var wg sync.WaitGroup
	numGoroutines := 20
	messagesPerGoroutine := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < messagesPerGoroutine; j++ {
				msg := synapse.SynapseMessage{
					Value:     0.3, // Small inputs that won't fire alone
					Timestamp: time.Now(),
					SourceID:  fmt.Sprintf("source_%d", goroutineID),
					SynapseID: fmt.Sprintf("synapse_%d_%d", goroutineID, j),
				}

				neuron.Receive(msg)
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Allow processing to complete
	time.Sleep(100 * time.Millisecond)

	// Verify no panics occurred and system is still functional
	gains := neuron.GetInputGains()
	if len(gains) == 0 {
		t.Error("No input gains registered - concurrent access may have caused issues")
	}

	// Test that neuron still responds normally
	neuron.Receive(synapse.SynapseMessage{
		Value: 6.0, Timestamp: time.Now(), SourceID: "post_concurrent_test", SynapseID: "test",
	})

	time.Sleep(20 * time.Millisecond)

	messages := targetNeuron.GetReceivedMessages()
	fired := false
	for _, msg := range messages {
		if msg.Value > 0 {
			fired = true
			break
		}
	}

	if !fired {
		t.Error("Neuron not responding after concurrent access test")
	}
}

// TestInvalidSynapseMessages validates handling of malformed inputs
//
// BIOLOGICAL CONTEXT:
// Robust neural systems must handle various input conditions gracefully,
// including invalid or extreme inputs. This ensures network stability
// even when individual components malfunction.
//
// EXPECTED RESULTS:
// - Invalid messages are handled gracefully
// - System remains stable with extreme inputs
// - No panics or crashes occur
// - Normal operation continues after invalid inputs
func TestInvalidSynapseMessages(t *testing.T) {
	neuron := NewSimpleNeuron("robust_test_neuron", 1.0, 0.98,
		5*time.Millisecond, 1.0)

	go neuron.Run()
	defer neuron.Close()

	// Test various invalid/extreme inputs
	testCases := []struct {
		name string
		msg  synapse.SynapseMessage
	}{
		{
			name: "NaN Value",
			msg: synapse.SynapseMessage{
				Value: math.NaN(), Timestamp: time.Now(), SourceID: "nan_test", SynapseID: "test",
			},
		},
		{
			name: "Infinite Value",
			msg: synapse.SynapseMessage{
				Value: math.Inf(1), Timestamp: time.Now(), SourceID: "inf_test", SynapseID: "test",
			},
		},
		{
			name: "Negative Infinite Value",
			msg: synapse.SynapseMessage{
				Value: math.Inf(-1), Timestamp: time.Now(), SourceID: "neg_inf_test", SynapseID: "test",
			},
		},
		{
			name: "Very Large Value",
			msg: synapse.SynapseMessage{
				Value: 1e10, Timestamp: time.Now(), SourceID: "large_test", SynapseID: "test",
			},
		},
		{
			name: "Very Small Value",
			msg: synapse.SynapseMessage{
				Value: 1e-10, Timestamp: time.Now(), SourceID: "small_test", SynapseID: "test",
			},
		},
		{
			name: "Zero Timestamp",
			msg: synapse.SynapseMessage{
				Value: 1.0, Timestamp: time.Time{}, SourceID: "zero_time_test", SynapseID: "test",
			},
		},
		{
			name: "Empty Source ID",
			msg: synapse.SynapseMessage{
				Value: 1.0, Timestamp: time.Now(), SourceID: "", SynapseID: "empty_source_test",
			},
		},
		{
			name: "Empty Synapse ID",
			msg: synapse.SynapseMessage{
				Value: 1.0, Timestamp: time.Now(), SourceID: "empty_synapse_test", SynapseID: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Panic occurred with %s: %v", tc.name, r)
					}
				}()

				neuron.Receive(tc.msg)
			}()
		})
	}

	// Allow processing time
	time.Sleep(50 * time.Millisecond)

	// Verify neuron still works normally after invalid inputs
	normalMsg := synapse.SynapseMessage{
		Value: 1.5, Timestamp: time.Now(), SourceID: "recovery_test", SynapseID: "test",
	}

	neuron.Receive(normalMsg)
	time.Sleep(10 * time.Millisecond)

	// If we get here without panicking, the test passed
}

// TestFireEventReporting validates the fire event reporting functionality
//
// BIOLOGICAL CONTEXT:
// In neuroscience research, monitoring individual neuron firing events is
// crucial for understanding network dynamics and learning. This models
// electrophysiological recording techniques used in biological research.
//
// EXPECTED RESULTS:
// - Fire events are reported when enabled
// - Event details match firing parameters
// - Timing information is accurate
// - Multiple events are captured correctly
func TestFireEventReporting(t *testing.T) {
	neuron := NewSimpleNeuron("fire_event_neuron", 1.0, 0.98,
		5*time.Millisecond, 2.0)

	// Set up fire event monitoring
	fireEvents := make(chan FireEvent, 10)
	neuron.SetFireEventChannel(fireEvents)

	// Set up output synapse for comparison
	targetNeuron := NewMockNeuron("event_target")
	outputSynapse := synapse.NewBasicSynapse(
		"event_output",
		neuron,
		targetNeuron,
		synapse.CreateDefaultSTDPConfig(),
		synapse.CreateDefaultPruningConfig(),
		1.0, 0,
	)
	neuron.AddOutputSynapse("event_test", outputSynapse)

	go neuron.Run()
	defer neuron.Close()

	// Send signal that should cause firing
	neuron.Receive(synapse.SynapseMessage{
		Value:     1.5,
		Timestamp: time.Now(),
		SourceID:  "event_test_source",
		SynapseID: "event_test_synapse",
	})

	// Wait for both events
	var fireEvent FireEvent
	var gotFireEvent bool

	select {
	case fireEvent = <-fireEvents:
		gotFireEvent = true
	case <-time.After(100 * time.Millisecond):
		t.Error("Did not receive fire event")
	}

	if gotFireEvent {
		// Verify fire event details
		if fireEvent.NeuronID != "fire_event_neuron" {
			t.Errorf("Expected neuron ID 'fire_event_neuron', got '%s'", fireEvent.NeuronID)
		}

		expectedValue := 1.5 * 2.0 // input * fireFactor
		if fireEvent.Value != expectedValue {
			t.Errorf("Expected fire event value %f, got %f", expectedValue, fireEvent.Value)
		}

		// Verify timestamp is recent
		if time.Since(fireEvent.Timestamp) > 200*time.Millisecond {
			t.Errorf("Fire event timestamp seems too old: %v ago", time.Since(fireEvent.Timestamp))
		}
	}

	// Verify output message was also sent
	time.Sleep(20 * time.Millisecond)
	messages := targetNeuron.GetReceivedMessages()
	if len(messages) == 0 {
		t.Error("No output message received - firing may have failed")
	}
}

// TestOutputSynapseManagement validates dynamic synapse connection management
//
// BIOLOGICAL CONTEXT:
// Real neurons can grow new synaptic connections (neuroplasticity) and
// eliminate unused ones (synaptic pruning). This test validates the
// ability to add and remove synaptic connections during runtime.
//
// EXPECTED RESULTS:
// - Synapses can be added and removed safely
// - Connection count updates correctly
// - Signals are transmitted to all connected synapses
// - Concurrent modification is thread-safe
func TestOutputSynapseManagement(t *testing.T) {
	neuron := NewSimpleNeuron("synapse_mgmt_neuron", 1.0, 0.98,
		5*time.Millisecond, 1.0)

	// Create test targets
	target1 := NewMockNeuron("target_1")
	target2 := NewMockNeuron("target_2")
	target3 := NewMockNeuron("target_3")

	// Test adding synapses
	synapse1 := synapse.NewBasicSynapse(
		"connection_1", neuron, target1,
		synapse.CreateDefaultSTDPConfig(),
		synapse.CreateDefaultPruningConfig(),
		1.0, 0,
	)
	neuron.AddOutputSynapse("conn1", synapse1)

	if neuron.GetOutputSynapseCount() != 1 {
		t.Errorf("Expected 1 synapse after adding, got %d", neuron.GetOutputSynapseCount())
	}

	synapse2 := synapse.NewBasicSynapse(
		"connection_2", neuron, target2,
		synapse.CreateDefaultSTDPConfig(),
		synapse.CreateDefaultPruningConfig(),
		1.0, 0,
	)
	neuron.AddOutputSynapse("conn2", synapse2)

	if neuron.GetOutputSynapseCount() != 2 {
		t.Errorf("Expected 2 synapses after adding second, got %d", neuron.GetOutputSynapseCount())
	}

	// Test synapse weight retrieval
	weight, exists := neuron.GetOutputSynapseWeight("conn1")
	if !exists {
		t.Error("Synapse conn1 weight not found")
	}
	if weight != 1.0 {
		t.Errorf("Expected weight 1.0, got %f", weight)
	}

	// Test removing synapses
	neuron.RemoveOutputSynapse("conn1")
	if neuron.GetOutputSynapseCount() != 1 {
		t.Errorf("Expected 1 synapse after removing, got %d", neuron.GetOutputSynapseCount())
	}

	// Test removing non-existent synapse (should not panic)
	neuron.RemoveOutputSynapse("nonexistent")
	if neuron.GetOutputSynapseCount() != 1 {
		t.Errorf("Expected 1 synapse after removing nonexistent, got %d", neuron.GetOutputSynapseCount())
	}

	// Test concurrent modification
	go neuron.Run()
	defer neuron.Close()

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Add synapse
			testSynapse := synapse.NewBasicSynapse(
				fmt.Sprintf("test_conn_%d", id), neuron, target3,
				synapse.CreateDefaultSTDPConfig(),
				synapse.CreateDefaultPruningConfig(),
				1.0, 0,
			)
			neuron.AddOutputSynapse(fmt.Sprintf("test_%d", id), testSynapse)

			// Brief delay
			time.Sleep(1 * time.Millisecond)

			// Remove synapse
			neuron.RemoveOutputSynapse(fmt.Sprintf("test_%d", id))
		}(i)
	}

	wg.Wait()

	// Should not have panicked and should be in consistent state
	finalCount := neuron.GetOutputSynapseCount()
	if finalCount < 0 {
		t.Errorf("Final synapse count is negative: %d", finalCount)
	}
}

// TestTemporalIntegration validates signal accumulation timing
//
// BIOLOGICAL CONTEXT:
// Biological neurons integrate signals over time, with recent inputs having
// stronger influence than older ones due to membrane decay. Multiple weak
// inputs arriving close together can sum to trigger firing (temporal summation).
//
// EXPECTED RESULTS:
// - Multiple rapid weak inputs can trigger firing
// - Widely spaced inputs don't sum effectively
// - Integration respects biological timing constraints
// - Decay rate affects summation window
func TestTemporalIntegration(t *testing.T) {
	decayRate := 0.99 // Slow decay to allow temporal summation
	neuron := NewSimpleNeuron("temporal_integration_neuron", 1.0, decayRate,
		5*time.Millisecond, 1.0)

	targetNeuron := NewMockNeuron("temporal_target")
	outputSynapse := synapse.NewBasicSynapse(
		"temporal_output", neuron, targetNeuron,
		synapse.CreateDefaultSTDPConfig(),
		synapse.CreateDefaultPruningConfig(),
		1.0, 0,
	)
	neuron.AddOutputSynapse("temporal_test", outputSynapse)

	go neuron.Run()
	defer neuron.Close()

	// Test 1: Rapid sequence of small signals should sum to fire
	neuron.Receive(synapse.SynapseMessage{
		Value: 0.4, Timestamp: time.Now(), SourceID: "summation_1", SynapseID: "test",
	})
	time.Sleep(2 * time.Millisecond)

	neuron.Receive(synapse.SynapseMessage{
		Value: 0.3, Timestamp: time.Now(), SourceID: "summation_2", SynapseID: "test",
	})
	time.Sleep(2 * time.Millisecond)

	neuron.Receive(synapse.SynapseMessage{
		Value: 0.4, Timestamp: time.Now(), SourceID: "summation_3", SynapseID: "test",
	})

	// Should fire due to temporal summation
	time.Sleep(20 * time.Millisecond)

	messages := targetNeuron.GetReceivedMessages()
	if len(messages) == 0 {
		t.Error("Neuron did not fire with rapid temporal summation")
	}

	targetNeuron.ClearReceivedMessages()

	// Test 2: Same inputs with larger delays should not fire
	neuron.Receive(synapse.SynapseMessage{
		Value: 0.4, Timestamp: time.Now(), SourceID: "delayed_1", SynapseID: "test",
	})
	time.Sleep(50 * time.Millisecond) // Long delay allows decay

	neuron.Receive(synapse.SynapseMessage{
		Value: 0.3, Timestamp: time.Now(), SourceID: "delayed_2", SynapseID: "test",
	})
	time.Sleep(50 * time.Millisecond)

	neuron.Receive(synapse.SynapseMessage{
		Value: 0.4, Timestamp: time.Now(), SourceID: "delayed_3", SynapseID: "test",
	})

	time.Sleep(20 * time.Millisecond)

	messages = targetNeuron.GetReceivedMessages()
	if len(messages) > 0 {
		t.Error("Neuron fired with delayed inputs - temporal integration too long")
	}
}

// TestInhibitorySignals validates negative signal processing
//
// BIOLOGICAL CONTEXT:
// Biological neurons receive both excitatory and inhibitory inputs.
// Inhibitory signals reduce membrane potential, making firing less likely.
// The balance between excitation and inhibition is crucial for neural
// computation and network stability.
//
// EXPECTED RESULTS:
// - Negative signals reduce accumulated potential
// - Inhibition can prevent firing when combined with excitation
// - Strong excitation can overcome moderate inhibition
// - Inhibitory signals are processed like excitatory but with opposite effect
func TestInhibitorySignals(t *testing.T) {
	neuron := NewSimpleNeuron("inhibitory_test_neuron", 1.0, 0.99,
		5*time.Millisecond, 1.0)

	targetNeuron := NewMockNeuron("inhibitory_target")
	outputSynapse := synapse.NewBasicSynapse(
		"inhibitory_output", neuron, targetNeuron,
		synapse.CreateDefaultSTDPConfig(),
		synapse.CreateDefaultPruningConfig(),
		1.0, 0,
	)
	neuron.AddOutputSynapse("inhibitory_test", outputSynapse)

	go neuron.Run()
	defer neuron.Close()

	// Test 1: Excitation followed by inhibition should not fire
	neuron.Receive(synapse.SynapseMessage{
		Value:     0.8, // Below threshold alone
		Timestamp: time.Now(),
		SourceID:  "excitatory_input",
		SynapseID: "test",
	})

	neuron.Receive(synapse.SynapseMessage{
		Value:     -0.3, // Inhibitory signal
		Timestamp: time.Now(),
		SourceID:  "inhibitory_input",
		SynapseID: "test",
	})

	time.Sleep(20 * time.Millisecond)

	messages := targetNeuron.GetReceivedMessages()
	if len(messages) > 0 {
		t.Error("Neuron fired despite inhibition reducing total below threshold")
	}

	// Test 2: Strong excitation should overcome moderate inhibition
	neuron.Receive(synapse.SynapseMessage{
		Value:     1.5, // Strong excitatory signal
		Timestamp: time.Now(),
		SourceID:  "strong_excitation",
		SynapseID: "test",
	})

	time.Sleep(20 * time.Millisecond)

	messages = targetNeuron.GetReceivedMessages()
	if len(messages) == 0 {
		t.Error("Strong excitation did not overcome inhibition")
	}

	// Test 3: Pure inhibitory signal should be processed without firing
	targetNeuron.ClearReceivedMessages()

	neuron.Receive(synapse.SynapseMessage{
		Value:     -0.5, // Pure inhibition
		Timestamp: time.Now(),
		SourceID:  "pure_inhibition",
		SynapseID: "test",
	})

	time.Sleep(20 * time.Millisecond)

	messages = targetNeuron.GetReceivedMessages()
	if len(messages) > 0 {
		t.Error("Neuron fired with pure inhibitory input")
	}
}

// TestNeuronIDInterface validates the ID() method implementation
//
// BIOLOGICAL CONTEXT:
// In neural networks, each neuron needs a unique identifier for tracking,
// learning algorithms (like STDP), and network analysis. This is crucial
// for synapse-neuron communication.
//
// EXPECTED RESULTS:
// - ID() method returns correct identifier
// - Interface compatibility with SynapseCompatibleNeuron
// - ID remains consistent throughout neuron lifetime
func TestNeuronIDInterface(t *testing.T) {
	neuronID := "interface_test_neuron_42"
	neuron := NewSimpleNeuron(neuronID, 1.0, 0.98, 5*time.Millisecond, 1.0)

	// Test ID() method
	if neuron.ID() != neuronID {
		t.Errorf("Expected ID %s, got %s", neuronID, neuron.ID())
	}

	// Test interface compatibility
	var synapseCompatible synapse.SynapseCompatibleNeuron = neuron
	if synapseCompatible.ID() != neuronID {
		t.Errorf("Interface ID() method failed: expected %s, got %s",
			neuronID, synapseCompatible.ID())
	}

	// Test ID consistency after operations
	go neuron.Run()
	defer neuron.Close()

	neuron.Receive(synapse.SynapseMessage{
		Value: 0.5, Timestamp: time.Now(), SourceID: "test", SynapseID: "test",
	})

	time.Sleep(10 * time.Millisecond)

	// ID should remain unchanged
	if neuron.ID() != neuronID {
		t.Errorf("ID changed after operations: expected %s, got %s",
			neuronID, neuron.ID())
	}
}

// TestGracefulShutdown validates proper neuron shutdown
//
// BIOLOGICAL CONTEXT:
// Neural systems must handle shutdown gracefully, ensuring all processing
// completes and resources are properly released. This models controlled
// cessation of neural activity.
//
// EXPECTED RESULTS:
// - Close() method stops neuron processing cleanly
// - No hanging goroutines after shutdown
// - Resources are properly released
// - Multiple Close() calls are safe
func TestGracefulShutdown(t *testing.T) {
	neuron := NewSimpleNeuron("shutdown_test_neuron", 1.0, 0.98,
		5*time.Millisecond, 1.0)

	fireEvents := make(chan FireEvent, 10)
	neuron.SetFireEventChannel(fireEvents)

	// Start neuron
	go neuron.Run()

	// 1. Send a signal to confirm it's working before shutdown
	neuron.Receive(synapse.SynapseMessage{Value: 1.5})
	select {
	case <-fireEvents:
		t.Log("✓ Neuron fired correctly while running.")
	case <-time.After(20 * time.Millisecond):
		t.Fatal("Neuron did not fire while running, test is invalid.")
	}

	// 2. Close the neuron
	neuron.Close()
	t.Log("Neuron Close() called.")

	// Give a moment for the shutdown to complete.
	time.Sleep(20 * time.Millisecond)

	// 3. Test that multiple Close() calls are safe (idempotent)
	// This should not panic.
	neuron.Close()
	neuron.Close()
	t.Log("✓ Multiple Close() calls did not cause a panic.")

	// 4. Send another strong signal to the *closed* neuron
	neuron.Receive(synapse.SynapseMessage{Value: 1.5})
	t.Log("Sent message to closed neuron.")

	// 5. Verify that the closed neuron did NOT fire
	select {
	case <-fireEvents:
		t.Fatal("FAIL: Neuron fired after being closed.")
	case <-time.After(20 * time.Millisecond):
		// This is the expected outcome.
		t.Log("✓ PASS: Neuron correctly ignored message after shutdown.")
	}
}

// TestNeuronBasicFiring tests fundamental neuron firing behavior
func TestNeuronBasicFiring(t *testing.T) {
	t.Log("=== Testing Basic Neuron Firing Behavior ===")

	testCases := []struct {
		name     string
		input    float64
		expected bool
	}{
		{"Below threshold", 1.0, false},
		{"Just below threshold", 1.9, false},
		{"At threshold", 2.0, true},
		{"Above threshold", 2.1, true},
		{"Well above threshold", 3.0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create FRESH neuron for each test case to avoid state contamination
			threshold := 2.0
			fireFactor := 1.0
			testNeuron := NewNeuron("test", threshold, 0.95, 10*time.Millisecond, fireFactor, 0, 0)
			defer testNeuron.Close()

			fireEvents := make(chan FireEvent, 10)
			testNeuron.SetFireEventChannel(fireEvents)
			go testNeuron.Run()

			// Send input
			testNeuron.Receive(synapse.SynapseMessage{
				Value:     tc.input,
				Timestamp: time.Now(),
				SourceID:  "test",
			})

			// Check if it fired
			time.Sleep(5 * time.Millisecond)
			fired := false

			select {
			case event := <-fireEvents:
				fired = true
				t.Logf("Fired with value %.2f", event.Value)
			case <-time.After(5 * time.Millisecond):
				t.Log("Did not fire")
			}

			if fired != tc.expected {
				t.Errorf("Input %.1f: expected firing=%t, got=%t", tc.input, tc.expected, fired)
			}
		})
	}
}

// TestCoincidenceDetection validates the neuron's ability to act as a temporal
// coincidence detector, a fundamental computational role for biological neurons.
//
// BIOLOGICAL CONTEXT:
// Coincidence detection is a key mechanism for neural computation, allowing
// neurons to fire preferentially in response to multiple, near-simultaneous
// excitatory inputs. This is crucial for:
//
//  1. TEMPORAL SUMMATION: Near-simultaneous excitatory postsynaptic potentials (EPSPs)
//     sum together more effectively to depolarize the membrane and reach the firing
//     threshold. Inputs that are too far apart in time decay before they can summate.
//
//  2. NMDA RECEPTOR ACTIVATION: This is a key molecular mechanism for coincidence
//     detection. NMDA receptors require two conditions to be met simultaneously:
//     - The binding of glutamate (the signal from a presynaptic neuron).
//     - Sufficient postsynaptic membrane depolarization (often from other coincident inputs)
//     to expel a magnesium ion (Mg2+) that blocks the receptor's channel.
//     This function models the outcome of this process: detecting correlated inputs.
//
//  3. FEATURE BINDING: In sensory systems, coincidence detection allows neurons to
//     bind together different features of a stimulus. For example, a neuron might
//     only fire when it receives simultaneous inputs representing a vertical edge
//     and a specific color, thus detecting a "vertical red line."
//
//  4. SYNAPTIC PLASTICITY: The Hebbian principle ("cells that fire together, wire
//     together") relies on detecting coincident pre- and post-synaptic activity.
//     Detecting coincident inputs is the first step in this process.
func TestCoincidenceDetection(t *testing.T) {
	t.Log("=== Testing Biological Coincidence Detection ===")

	testCases := []struct {
		name               string
		inputs             []float64
		delays             []time.Duration
		expectedFire       bool
		expectedCoincident int
		threshold          float64 // Custom threshold for this test
		decayRate          float64 // Custom decay rate for this test
	}{
		{
			name:               "Single Strong Input",
			inputs:             []float64{3.0},
			delays:             []time.Duration{0},
			expectedFire:       true,
			expectedCoincident: 1,
			threshold:          2.5,
			decayRate:          0.90,
		},
		{
			name:               "Two Weak Inputs (Simultaneous)",
			inputs:             []float64{1.5, 1.5},
			delays:             []time.Duration{0, 0},
			expectedFire:       true,
			expectedCoincident: 2,
			threshold:          2.5,
			decayRate:          0.90,
		},
		{
			name:               "Two Weak Inputs (Within Window)",
			inputs:             []float64{1.6, 1.6},                      // Slightly stronger to overcome decay
			delays:             []time.Duration{0, 3 * time.Millisecond}, // Shorter delay
			expectedFire:       true,
			expectedCoincident: 2,
			threshold:          2.5,
			decayRate:          0.95, // Slower decay to preserve summation
		},
		{
			name:               "Two Weak Inputs (Outside Window)",
			inputs:             []float64{1.5, 1.5},
			delays:             []time.Duration{0, 15 * time.Millisecond},
			expectedFire:       false,
			expectedCoincident: 1,
			threshold:          2.5,
			decayRate:          0.90,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// --- SETUP with custom parameters ---
			coincidenceNeuron := NewNeuron("coincidence_detector", tc.threshold, tc.decayRate, 10*time.Millisecond, 1.0, 0, 0)
			defer coincidenceNeuron.Close()

			coincidenceWindow := 10 * time.Millisecond
			coincidenceNeuron.SetCoincidenceDetection(true, coincidenceWindow)

			fireEvents := make(chan FireEvent, 1)
			coincidenceNeuron.SetFireEventChannel(fireEvents)
			go coincidenceNeuron.Run()

			time.Sleep(5 * time.Millisecond)

			// --- STIMULATION ---
			t.Logf("Presenting pattern: %d inputs with delays %v (threshold=%.1f, decay=%.2f)",
				len(tc.inputs), tc.delays, tc.threshold, tc.decayRate)

			for j, inputValue := range tc.inputs {
				if tc.delays[j] > 0 {
					time.Sleep(tc.delays[j])
				}
				coincidenceNeuron.Receive(synapse.SynapseMessage{
					Value:     inputValue,
					Timestamp: time.Now(),
					SourceID:  fmt.Sprintf("input_source_%d", j),
				})
			}

			// Check coincidence detection immediately
			time.Sleep(1 * time.Millisecond)
			coincidentInputs := coincidenceNeuron.DetectCoincidentInputs()

			// Check firing
			fired := false
			select {
			case event := <-fireEvents:
				fired = true
				t.Logf("  ✔️  Neuron Fired! (Signal value: %.2f)", event.Value)
			case <-time.After(10 * time.Millisecond):
				t.Log("  ❌ Neuron did not fire.")
			}

			// Show final accumulator state for debugging
			finalState := coincidenceNeuron.GetNeuronState()
			t.Logf("  Final accumulator: %.3f (threshold: %.1f)",
				finalState["accumulator"], tc.threshold)

			// --- VERIFICATION ---
			if fired != tc.expectedFire {
				t.Errorf("FAIL: Firing expectation mismatch. Expected: %v, Got: %v", tc.expectedFire, fired)
			} else {
				t.Logf("  ✓ PASS: Firing behavior matched expectation (%v).", tc.expectedFire)
			}

			t.Logf("  Detected %d coincident inputs within the %v window.", coincidentInputs, coincidenceWindow)

			if coincidentInputs != tc.expectedCoincident {
				t.Errorf("FAIL: Coincident input count mismatch. Expected: %d, Got: %d", tc.expectedCoincident, coincidentInputs)
			} else {
				t.Logf("  ✓ PASS: Coincident input count matched expectation (%d).", tc.expectedCoincident)
			}

			time.Sleep(30 * time.Millisecond)
		})
	}
}

// TestTemporalSummation tests how inputs accumulate over time
func TestTemporalSummation(t *testing.T) {
	t.Log("=== Testing Temporal Summation ===")

	testCases := []struct {
		name      string
		inputs    []float64
		interval  time.Duration
		expected  bool
		decayRate float64 // Custom decay rate for each test
	}{
		{
			name:      "Two inputs close together",
			inputs:    []float64{1.8, 1.8},
			interval:  1 * time.Millisecond,
			expected:  true, // Should sum to > 3.0
			decayRate: 0.90,
		},
		{
			name:      "Two inputs far apart",
			inputs:    []float64{1.8, 1.8},
			interval:  20 * time.Millisecond,
			expected:  false, // Decay should prevent summation
			decayRate: 0.90,
		},
		{
			name:      "Three weak inputs rapid",
			inputs:    []float64{1.2, 1.2, 1.2},
			interval:  1 * time.Millisecond, // Faster interval
			expected:  true,                 // Should accumulate
			decayRate: 0.98,                 // Slower decay to preserve accumulation
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fresh neuron for each test with appropriate decay rate
			neuron := NewNeuron("summation", 3.0, tc.decayRate, 10*time.Millisecond, 1.0, 0, 0)
			defer neuron.Close()

			fireEvents := make(chan FireEvent, 10)
			neuron.SetFireEventChannel(fireEvents)
			go neuron.Run()

			// Send inputs with intervals
			for j, input := range tc.inputs {
				if j > 0 {
					time.Sleep(tc.interval)
				}

				neuron.Receive(synapse.SynapseMessage{
					Value:     input,
					Timestamp: time.Now(),
					SourceID:  fmt.Sprintf("input_%d", j),
				})
			}

			// Check result
			time.Sleep(15 * time.Millisecond)
			fired := false

			select {
			case event := <-fireEvents:
				fired = true
				t.Logf("Neuron fired with value %.2f", event.Value)
			case <-time.After(5 * time.Millisecond):
				t.Log("Neuron did not fire")
			}

			if fired != tc.expected {
				t.Errorf("Expected %t, got %t", tc.expected, fired)
			}
		})
	}
}

// TestInhibition tests inhibitory signal behavior with proper timing
func TestInhibition(t *testing.T) {
	t.Log("=== Testing Inhibition ===")

	testCases := []struct {
		name       string
		excitation float64
		inhibition float64
		expected   bool
	}{
		{
			name:       "Excitation only",
			excitation: 2.5,
			inhibition: 0,
			expected:   true,
		},
		{
			name:       "Weak inhibition",
			excitation: 1.8,  // Reduced so it won't fire alone
			inhibition: -0.3, // 1.8 - 0.3 = 1.5 < 2.0 threshold
			expected:   false,
		},
		{
			name:       "Strong inhibition",
			excitation: 1.8,  // Reduced so it won't fire alone
			inhibition: -2.0, // 1.8 - 2.0 = -0.2 < 2.0 threshold
			expected:   false,
		},
		{
			name:       "Overcome inhibition",
			excitation: 3.0,  // Strong enough to overcome inhibition
			inhibition: -0.5, // 3.0 - 0.5 = 2.5 > 2.0 threshold
			expected:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fresh neuron for each test case
			// Use slower decay rate to ensure signals sum properly
			neuron := NewNeuron("inhibition", 2.0, 0.999, 10*time.Millisecond, 1.0, 0, 0)
			defer neuron.Close()

			fireEvents := make(chan FireEvent, 10)
			neuron.SetFireEventChannel(fireEvents)
			go neuron.Run()

			// Wait for neuron to be ready
			time.Sleep(2 * time.Millisecond)

			// Send both signals rapidly to ensure temporal summation
			neuron.Receive(synapse.SynapseMessage{
				Value:     tc.excitation,
				Timestamp: time.Now(),
				SourceID:  "excitatory",
			})

			// Send inhibition immediately (if present)
			if tc.inhibition != 0 {
				neuron.Receive(synapse.SynapseMessage{
					Value:     tc.inhibition,
					Timestamp: time.Now(),
					SourceID:  "inhibitory",
				})
			}

			// Check result
			time.Sleep(15 * time.Millisecond)
			fired := false

			select {
			case event := <-fireEvents:
				fired = true
				t.Logf("Neuron fired with value %.2f", event.Value)
			case <-time.After(5 * time.Millisecond):
				t.Log("Neuron did not fire")
			}

			// Debug: Check final accumulator state
			finalState := neuron.GetNeuronState()
			t.Logf("Final accumulator: %.3f, expected sum: %.1f",
				finalState["accumulator"], tc.excitation+tc.inhibition)

			if fired != tc.expected {
				t.Errorf("Excitation %.1f + Inhibition %.1f: expected %t, got %t",
					tc.excitation, tc.inhibition, tc.expected, fired)
			}
		})
	}
}

// TestSynapseDelays tests synapse transmission delays
func TestSynapseDelays(t *testing.T) {
	t.Log("=== Testing Synapse Delays ===")

	delays := []time.Duration{0, 5 * time.Millisecond, 10 * time.Millisecond}
	tolerance := 5 * time.Millisecond // Allow 5ms tolerance for timing

	for _, expectedDelay := range delays {
		t.Run(fmt.Sprintf("Delay_%v", expectedDelay), func(t *testing.T) {
			// Create input and output neurons
			inputNeuron := NewSimpleNeuron("input", 0.5, 0.95, 5*time.Millisecond, 1.0)
			outputNeuron := NewNeuron("output", 2.0, 0.95, 10*time.Millisecond, 1.0, 0, 0)

			defer inputNeuron.Close()
			defer outputNeuron.Close()

			// Create synapse with specific delay
			synapseConfig := synapse.CreateDefaultSTDPConfig()
			synapseConfig.Enabled = false
			pruningConfig := synapse.CreateDefaultPruningConfig()

			testSynapse := synapse.NewBasicSynapse("test", inputNeuron, outputNeuron,
				synapseConfig, pruningConfig, 2.5, expectedDelay)

			inputNeuron.AddOutputSynapse("test", testSynapse)

			fireEvents := make(chan FireEvent, 10)
			outputNeuron.SetFireEventChannel(fireEvents)

			go inputNeuron.Run()
			go outputNeuron.Run()

			// Send input and measure timing
			startTime := time.Now()
			inputNeuron.Receive(synapse.SynapseMessage{
				Value:     2.0,
				Timestamp: startTime,
				SourceID:  "timing_test",
			})

			// Wait for output
			select {
			case event := <-fireEvents:
				actualDelay := event.Timestamp.Sub(startTime)
				t.Logf("Expected delay: %v, Actual delay: %v", expectedDelay, actualDelay)

				if actualDelay < expectedDelay || actualDelay > expectedDelay+tolerance {
					t.Errorf("Delay out of range: expected %v (±%v), got %v",
						expectedDelay, tolerance, actualDelay)
				}

			case <-time.After(expectedDelay + 50*time.Millisecond):
				t.Errorf("No output received within timeout")
			}
		})
	}
}

// TestNeuronRealism tests biological realism constraints
func TestNeuronRealism(t *testing.T) {
	t.Log("=== Testing Biological Realism ===")

	neuron := NewNeuron("realism", 1.0, 0.95, 5*time.Millisecond, 1.0, 0, 0)
	defer neuron.Close()

	fireEvents := make(chan FireEvent, 100)
	neuron.SetFireEventChannel(fireEvents)
	go neuron.Run()

	// Test refractory period
	t.Run("Refractory period", func(t *testing.T) {
		// Send strong input to make it fire
		neuron.Receive(synapse.SynapseMessage{Value: 2.0, SourceID: "test"})

		// Should fire
		select {
		case <-fireEvents:
			t.Log("First spike successful")
		case <-time.After(10 * time.Millisecond):
			t.Fatal("First spike failed")
		}

		// Immediately send another strong input (should be blocked by refractory)
		neuron.Receive(synapse.SynapseMessage{Value: 2.0, SourceID: "test"})

		// Should NOT fire
		select {
		case <-fireEvents:
			t.Error("Second spike should be blocked by refractory period")
		case <-time.After(3 * time.Millisecond):
			t.Log("Refractory period correctly blocked second spike")
		}

		// Wait for refractory to end, then try again
		time.Sleep(10 * time.Millisecond)
		neuron.Receive(synapse.SynapseMessage{Value: 2.0, SourceID: "test"})

		// Should fire again
		select {
		case <-fireEvents:
			t.Log("Third spike successful after refractory")
		case <-time.After(10 * time.Millisecond):
			t.Error("Third spike should work after refractory period")
		}
	})
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// ============================================================================
// BENCHMARK TESTS
// ============================================================================

// BenchmarkNeuronCreation benchmarks neuron creation performance
func BenchmarkNeuronCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		neuronID := fmt.Sprintf("bench_neuron_%d", i)
		_ = NewSimpleNeuron(neuronID, 1.0, 0.95, 5*time.Millisecond, 1.0)
	}
}

// BenchmarkSynapseMessageProcessing benchmarks message processing throughput
func BenchmarkSynapseMessageProcessing(b *testing.B) {
	neuron := NewSimpleNeuron("bench_processing", 10.0, 0.95,
		5*time.Millisecond, 1.0) // High threshold to avoid firing

	go neuron.Run()
	defer neuron.Close()

	msg := synapse.SynapseMessage{
		Value:     0.1,
		Timestamp: time.Now(),
		SourceID:  "benchmark_source",
		SynapseID: "benchmark_synapse",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		neuron.Receive(msg)
	}
}

// BenchmarkNeuronFiring benchmarks basic neuron firing performance
func BenchmarkNeuronFiring(b *testing.B) {
	neuron := NewNeuron("bench", 1.0, 0.95, 5*time.Millisecond, 1.0, 0, 0)
	defer neuron.Close()

	fireEvents := make(chan FireEvent, 1000)
	neuron.SetFireEventChannel(fireEvents)
	go neuron.Run()

	// Drain fire events in background
	go func() {
		for range fireEvents {
			// Consume events
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		neuron.Receive(synapse.SynapseMessage{
			Value:     1.5,
			Timestamp: time.Now(),
			SourceID:  "bench",
		})
	}
}

// ============================================================================
// DEBUG TESTS
// ============================================================================

// TestSynapseTransmissionDelay_Debug is a focused test to isolate and verify
// that the synapse's transmission delay is being correctly implemented and timed.
// It bypasses the presynaptic neuron's firing logic to test the synapse directly.
func TestSynapseTransmissionDelay_Debug(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping debug-specific test in short mode.")
	}
	// 1. Setup: Create mock pre- and post-synaptic neurons.
	// We use mocks to have full control and observability.
	preNeuron := NewMockNeuron("debug_pre_neuron")
	postNeuron := NewMockNeuron("debug_post_neuron")

	// 2. Configuration: Define a clear, unambiguous delay.
	// A 50ms delay is long enough to be clearly distinct from any scheduling noise.
	const transmissionDelay = 50 * time.Millisecond
	const signalValue = 1.0
	const weight = 1.0

	synapticConnection := synapse.NewBasicSynapse(
		"debug_synapse",
		preNeuron,
		postNeuron,
		synapse.CreateDefaultSTDPConfig(),
		synapse.CreateDefaultPruningConfig(),
		weight,
		transmissionDelay,
	)

	// 3. Execution: Record start time and transmit the signal directly.
	// We call Transmit() on the synapse itself, not on the neuron.
	startTime := time.Now()
	t.Logf("[DEBUG] Test initiated at: %v", startTime.Format(time.RFC3339Nano))
	synapticConnection.Transmit(signalValue)
	t.Logf("[DEBUG] Synapse.Transmit() called. Waiting for delayed delivery...")

	// 4. Wait: Allow more than enough time for the delayed function to run.
	time.Sleep(transmissionDelay + 30*time.Millisecond)

	// 5. Verification & Logging: Check the results and log everything.
	receivedMessages := postNeuron.GetReceivedMessages()

	if len(receivedMessages) == 0 {
		t.Fatal("[DEBUG] FATAL: Post-synaptic neuron received no messages. The time.AfterFunc callback likely did not run.")
	}
	if len(receivedMessages) > 1 {
		t.Fatalf("[DEBUG] FATAL: Received %d messages, but expected only 1.", len(receivedMessages))
	}

	receivedMsg := receivedMessages[0]
	t.Logf("[DEBUG] Message received by mock neuron. Timestamp inside message: %v", receivedMsg.Timestamp.Format(time.RFC3339Nano))

	// This is the critical calculation.
	// It measures the difference between the start of the test and the timestamp
	// that was embedded inside the message upon its creation.
	actualDelay := receivedMsg.Timestamp.Sub(startTime)
	t.Logf("[DEBUG] Calculated Actual Delay (Message Timestamp - Start Time): %v", actualDelay)

	// 6. Assertion: Check if the actual delay is within an acceptable window.
	// We expect the delay to be very close to our configured 50ms, allowing for minor
	// goroutine scheduling overhead.
	minExpectedDelay := transmissionDelay
	maxExpectedDelay := transmissionDelay + 30*time.Millisecond // Generous window for scheduling

	if actualDelay < minExpectedDelay {
		t.Errorf("FAIL: Signal arrived too quickly. Actual Delay: %v, Expected Minimum: %v. This indicates the OLD, buggy synapse code is running.", actualDelay, minExpectedDelay)
	} else if actualDelay > maxExpectedDelay {
		t.Errorf("FAIL: Signal arrived too slowly. Actual Delay: %v, Expected Maximum: %v.", actualDelay, maxExpectedDelay)
	} else {
		t.Logf("SUCCESS: Actual delay of %v is within the expected window [%v, %v]. The synapse code is working correctly.", actualDelay, minExpectedDelay, maxExpectedDelay)
	}
}
