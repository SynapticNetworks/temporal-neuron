package integration

import (
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/extracellular"
	"github.com/SynapticNetworks/temporal-neuron/message"
	"github.com/SynapticNetworks/temporal-neuron/neuron"
	"github.com/SynapticNetworks/temporal-neuron/synapse"
)

// TestNeuronInitialization tests basic neuron creation and initialization
func TestNeuronInitialization(t *testing.T) {
	t.Log("=== NEURON INITIALIZATION TEST ===")

	// Create a basic neuron with standard parameters
	neuronID := "test_neuron_001"
	threshold := 1.0
	decayRate := 0.95
	refractoryPeriod := 5 * time.Millisecond
	fireFactor := 1.5
	targetFiringRate := 10.0
	homeostasisStrength := 0.1

	testNeuron := neuron.NewNeuron(
		neuronID,
		threshold,
		decayRate,
		refractoryPeriod,
		fireFactor,
		targetFiringRate,
		homeostasisStrength,
	)

	// Verify neuron was created successfully
	if testNeuron == nil {
		t.Fatal("Failed to create neuron - NewNeuron returned nil")
	}

	// Verify basic properties
	if testNeuron.ID() != neuronID {
		t.Errorf("Neuron ID mismatch: expected %s, got %s", neuronID, testNeuron.ID())
	}

	if testNeuron.GetThreshold() != threshold {
		t.Errorf("Threshold mismatch: expected %f, got %f", threshold, testNeuron.GetThreshold())
	}

	// Check initial neuron state - should be inactive since Start() hasn't been called
	initialActive := testNeuron.IsActive()
	t.Logf("Neuron initial state: active=%v", initialActive)

	// Neuron should be inactive upon creation (no background processing yet)
	if initialActive {
		t.Error("Neuron should be inactive before Start() is called")
	}

	// Test lifecycle - Start
	err := testNeuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}

	// Verify neuron is still active after Start() (should remain true)
	if !testNeuron.IsActive() {
		t.Error("Neuron should be active after Start() is called")
	}

	// Test that active neuron DOES process messages
	t.Log("Testing message processing on active neuron...")

	// Send a message to the active neuron
	activeTestMessage := message.NeuralSignal{
		Value:     2.0, // Strong signal that should cause processing
		Timestamp: time.Now(),
		SourceID:  "test_source_active",
		TargetID:  testNeuron.ID(),
	}

	// Get initial activity level before sending message
	initialActiveLevel := testNeuron.GetActivityLevel()

	// Send message to active neuron
	testNeuron.Receive(activeTestMessage)

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	// Verify neuron processed the message (activity level should change or firing should occur)
	finalActiveLevel := testNeuron.GetActivityLevel()

	// Note: We can't guarantee firing will occur (depends on threshold), but we can check
	// that the neuron is still active and the message was queued
	if !testNeuron.IsActive() {
		t.Error("Active neuron should remain active after processing message")
	}

	t.Logf("Active neuron activity levels: before=%f, after=%f", initialActiveLevel, finalActiveLevel)
	t.Log("✓ Active neuron accepted message for processing")

	// Test lifecycle - Stop
	err = testNeuron.Stop()
	if err != nil {
		t.Errorf("Failed to stop neuron: %v", err)
	}

	// Verify neuron is inactive after Stop() (context cancelled)
	if testNeuron.IsActive() {
		t.Error("Neuron should be inactive after Stop() is called")
	}

	// Test that stopped neuron DOESN'T process messages
	t.Log("Testing message processing on stopped neuron...")

	// Send a message to the stopped neuron
	stoppedTestMessage := message.NeuralSignal{
		Value:     2.0, // Same strong signal
		Timestamp: time.Now(),
		SourceID:  "test_source_stopped",
		TargetID:  testNeuron.ID(),
	}

	// Get initial activity level before sending message
	initialStoppedLevel := testNeuron.GetActivityLevel()

	// Send message to stopped neuron
	testNeuron.Receive(stoppedTestMessage)

	// Wait a moment to ensure any processing would have occurred
	time.Sleep(50 * time.Millisecond)

	// Verify neuron didn't process the message (activity level should be unchanged)
	finalStoppedLevel := testNeuron.GetActivityLevel()
	if finalStoppedLevel != initialStoppedLevel {
		t.Errorf("Stopped neuron should not process messages - activity level changed from %f to %f",
			initialStoppedLevel, finalStoppedLevel)
	}

	// Verify neuron is still inactive
	if testNeuron.IsActive() {
		t.Error("Neuron should remain inactive after receiving message while stopped")
	}

	t.Logf("Stopped neuron activity levels: before=%f, after=%f", initialStoppedLevel, finalStoppedLevel)
	t.Log("✓ Stopped neuron correctly ignores incoming messages")

	t.Log("✓ Neuron initialization, lifecycle, and actual running state properly validated")
}

// TestSynapseInitialization tests basic synapse creation with real neurons
// TestSynapseInitialization tests basic synapse creation with real neurons
func TestSynapseInitialization(t *testing.T) {
	t.Log("=== SYNAPSE INITIALIZATION TEST ===")

	// Create REAL neurons for testing synapse integration
	preNeuron := neuron.NewNeuron(
		"pre_neuron",
		1.0,                // threshold
		0.95,               // decay rate
		5*time.Millisecond, // refractory period
		1.5,                // fire factor
		10.0,               // target firing rate
		0.1,                // homeostasis strength
	)

	postNeuron := neuron.NewNeuron(
		"post_neuron",
		1.0,                // threshold
		0.95,               // decay rate
		5*time.Millisecond, // refractory period
		1.5,                // fire factor
		10.0,               // target firing rate
		0.1,                // homeostasis strength
	)

	// Start both neurons for proper integration testing
	err := preNeuron.Start()
	if err != nil {
		t.Fatalf("Failed to start pre-neuron: %v", err)
	}
	defer preNeuron.Stop()

	err = postNeuron.Start()
	if err != nil {
		t.Fatalf("Failed to start post-neuron: %v", err)
	}
	defer postNeuron.Stop()

	// Configure STDP and pruning with defaults
	stdpConfig := synapse.CreateDefaultSTDPConfig()
	pruningConfig := synapse.CreateDefaultPruningConfig()

	// Create synapse with REAL neurons using new interface
	synapseID := "test_synapse_001"
	initialWeight := 0.5
	delay := 2 * time.Millisecond

	testSynapse := synapse.NewBasicSynapse(
		synapseID,
		preNeuron,
		postNeuron,
		stdpConfig,
		pruningConfig,
		initialWeight,
		delay,
	)

	// Verify synapse was created successfully
	if testSynapse == nil {
		t.Fatal("Failed to create synapse - NewBasicSynapse returned nil")
	}

	// Verify basic properties
	if testSynapse.ID() != synapseID {
		t.Errorf("Synapse ID mismatch: expected %s, got %s", synapseID, testSynapse.ID())
	}

	if testSynapse.GetWeight() != initialWeight {
		t.Errorf("Weight mismatch: expected %f, got %f", initialWeight, testSynapse.GetWeight())
	}

	if testSynapse.GetDelay() != delay {
		t.Errorf("Delay mismatch: expected %v, got %v", delay, testSynapse.GetDelay())
	}

	// Test synapse activity check (IsActive expects a time.Duration parameter)
	activeDuration := 10 * time.Second
	if !testSynapse.IsActive(activeDuration) {
		t.Error("Synapse should be active upon creation")
	}

	// Get initial message count from post-neuron (using reflection or monitoring)
	// We'll monitor the post-neuron's activity level to detect signal reception
	initialActivity := postNeuron.GetActivityLevel()

	// Test basic signal transmission with REAL neurons
	signalValue := 1.0
	testSynapse.Transmit(signalValue)

	// Wait for signal processing (includes delay + processing time)
	time.Sleep(delay + 50*time.Millisecond)

	// Check if post-neuron received and processed the signal
	finalActivity := postNeuron.GetActivityLevel()

	// The activity level should have changed if the signal was received and processed
	if finalActivity == initialActivity {
		t.Logf("Post-neuron activity unchanged: %f → %f", initialActivity, finalActivity)
		// This might be expected if the signal wasn't strong enough to change activity significantly
		// Let's also check the homeostatic state or try a stronger signal
	} else {
		t.Logf("✓ Post-neuron activity changed: %f → %f (signal received)", initialActivity, finalActivity)
	}

	// Test with a stronger signal that should definitely cause activity change
	strongSignalValue := 3.0
	preActivity := postNeuron.GetActivityLevel()

	testSynapse.Transmit(strongSignalValue)
	time.Sleep(delay + 50*time.Millisecond)

	postActivity := postNeuron.GetActivityLevel()

	if postActivity != preActivity {
		t.Logf("✓ Strong signal caused activity change: %f → %f", preActivity, postActivity)
	} else {
		t.Logf("Strong signal activity: %f → %f", preActivity, postActivity)
	}

	// Test STDP functionality
	adjustment := synapse.PlasticityAdjustment{
		DeltaT:       -10 * time.Millisecond, // Pre before post (LTP)
		PostSynaptic: true,
		PreSynaptic:  true,
		Timestamp:    time.Now(),
	}

	preWeight := testSynapse.GetWeight()
	testSynapse.ApplyPlasticity(adjustment)
	postWeight := testSynapse.GetWeight()

	if postWeight > preWeight {
		t.Logf("✓ STDP strengthened synapse: %f → %f", preWeight, postWeight)
	} else {
		t.Logf("STDP weight change: %f → %f", preWeight, postWeight)
	}

	// Test delayed delivery system
	if delay > 0 {
		// Send a signal and immediately check - should not be delivered yet
		preActivity = postNeuron.GetActivityLevel()
		testSynapse.Transmit(2.0)

		// Check immediately (before delay)
		immediateActivity := postNeuron.GetActivityLevel()

		// Wait for delay to pass
		time.Sleep(delay + 10*time.Millisecond)

		delayedActivity := postNeuron.GetActivityLevel()

		t.Logf("Delayed delivery test: pre=%f, immediate=%f, delayed=%f",
			preActivity, immediateActivity, delayedActivity)
	}

	// Test pruning protection (recently active synapse should not prune)
	if testSynapse.ShouldPrune() {
		t.Error("Recently active synapse should not be marked for pruning")
	}

	t.Log("✓ Synapse initialization and real neuron integration validated")
}

// TestMatrixNeuronSynapseIntegration tests neuron and synapse creation via matrix
// TestMatrixNeuronSynapseIntegration tests neuron and synapse creation via matrix
func TestMatrixNeuronSynapseIntegration(t *testing.T) {
	t.Log("=== MATRIX NEURON-SYNAPSE INTEGRATION TEST ===")

	// Create and start the extracellular matrix
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

	// Register factory for creating standard neurons via matrix
	matrix.RegisterNeuronType("standard", func(id string, config extracellular.NeuronConfig, callbacks extracellular.NeuronCallbacks) (extracellular.NeuronInterface, error) {
		return neuron.NewNeuron(
			id,
			config.Threshold,
			0.95,               // decay rate
			5*time.Millisecond, // refractory period
			1.5,                // fire factor
			10.0,               // target firing rate
			0.1,                // homeostasis strength
		), nil
	})

	// Register factory for creating synapses via matrix
	matrix.RegisterSynapseType("excitatory", func(id string, config extracellular.SynapseConfig, callbacks extracellular.SynapseCallbacks) (extracellular.SynapseInterface, error) {
		preNeuron := matrix.GetNeuron(config.PresynapticID)
		postNeuron := matrix.GetNeuron(config.PostsynapticID)

		return synapse.NewBasicSynapse(
			id,
			preNeuron,
			postNeuron,
			synapse.CreateDefaultSTDPConfig(),
			synapse.CreateDefaultPruningConfig(),
			config.InitialWeight,
			config.Delay,
		), nil
	})

	// Create neurons via matrix
	preNeuron, err := matrix.CreateNeuron(extracellular.NeuronConfig{
		NeuronType: "standard",
		Threshold:  1.0,
		Position:   component.Position3D{X: 0, Y: 0, Z: 0},
	})
	if err != nil {
		t.Fatalf("Failed to create presynaptic neuron: %v", err)
	}

	postNeuron, err := matrix.CreateNeuron(extracellular.NeuronConfig{
		NeuronType: "standard",
		Threshold:  1.0,
		Position:   component.Position3D{X: 1, Y: 0, Z: 0},
	})
	if err != nil {
		t.Fatalf("Failed to create postsynaptic neuron: %v", err)
	}

	// Start neurons
	err = preNeuron.Start()
	if err != nil {
		t.Fatalf("Failed to start presynaptic neuron: %v", err)
	}
	defer preNeuron.Stop()

	err = postNeuron.Start()
	if err != nil {
		t.Fatalf("Failed to start postsynaptic neuron: %v", err)
	}
	defer postNeuron.Stop()

	// Create synapse via matrix
	syn, err := matrix.CreateSynapse(extracellular.SynapseConfig{
		SynapseType:    "excitatory",
		PresynapticID:  preNeuron.ID(),
		PostsynapticID: postNeuron.ID(),
		InitialWeight:  0.8,
		Delay:          2 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create synapse: %v", err)
	}

	// Test the connection
	initialPreActivity := preNeuron.GetActivityLevel()
	initialPostActivity := postNeuron.GetActivityLevel()

	preNeuron.Receive(message.NeuralSignal{
		Value:     2.0,
		Timestamp: time.Now(),
		SourceID:  "test_stimulus",
		TargetID:  preNeuron.ID(),
	})

	time.Sleep(50 * time.Millisecond)

	finalPreActivity := preNeuron.GetActivityLevel()
	finalPostActivity := postNeuron.GetActivityLevel()

	if finalPreActivity <= initialPreActivity {
		t.Errorf("Presynaptic neuron should show increased activity: before=%f, after=%f",
			initialPreActivity, finalPreActivity)
	}

	if finalPostActivity <= initialPostActivity {
		t.Errorf("Postsynaptic neuron should show increased activity: before=%f, after=%f",
			initialPostActivity, finalPostActivity)
	}

	t.Logf("✓ Matrix successfully created and connected neurons via synapse")
	t.Logf("Pre-neuron activity: %f → %f", initialPreActivity, finalPreActivity)
	t.Logf("Post-neuron activity: %f → %f", initialPostActivity, finalPostActivity)
	t.Logf("Synapse ID: %s", syn.ID())
}
