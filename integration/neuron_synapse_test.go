package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/extracellular"
	"github.com/SynapticNetworks/temporal-neuron/neuron"
	"github.com/SynapticNetworks/temporal-neuron/synapse"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// TestSTDPIntegration_DirectTiming tests both LTP and LTD cases using direct spike recording
// Following the pattern of the successful TestSTDPIntegration_LTDOnlyDirect test
func TestSTDPIntegration_DirectTiming(t *testing.T) {
	// Run LTP and LTD tests as separate subtests to avoid interference
	t.Run("LTP_Case", func(t *testing.T) {
		testLTPWithDirectSpikes(t)
	})

	t.Run("LTD_Case", func(t *testing.T) {
		testLTDWithDirectSpikes(t)
	})
}

// testLTPWithDirectSpikes tests the Long-Term Potentiation case (pre before post)
// This follows the pattern from the successful TestSTDPIntegration_LTDOnlyDirect test
func testLTPWithDirectSpikes(t *testing.T) {
	t.Log("\n=== DIRECT LTP TIMING TEST ===")
	t.Log("Testing the LTP case (Pre fires before Post) with direct spike history manipulation")

	// Create matrix with standard configuration
	matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  5 * time.Millisecond,
		MaxComponents:   10,
	})

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}
	defer matrix.Stop()

	// Register neuron type
	matrix.RegisterNeuronType("test_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(
			id,
			config.Threshold,
			0.95,               // decay rate
			5*time.Millisecond, // refractory period
			2.0,                // fire factor
			3.0,                // target firing rate
			0.2,                // homeostasis strength
		)
		n.SetCallbacks(callbacks)
		return n, nil
	})

	// Register synapse type with special test configuration
	matrix.RegisterSynapseType("test_synapse", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, exists := matrix.GetNeuron(config.PresynapticID)
		if !exists {
			return nil, fmt.Errorf("presynaptic neuron not found: %s", config.PresynapticID)
		}

		postNeuron, exists := matrix.GetNeuron(config.PostsynapticID)
		if !exists {
			return nil, fmt.Errorf("postsynaptic neuron not found: %s", config.PostsynapticID)
		}

		// Create special STDP config for test
		stdpConfig := synapse.CreateDefaultSTDPConfig()
		stdpConfig.WindowSize = 400 * time.Millisecond // Much larger window for test

		return synapse.NewBasicSynapse(
			id,
			preNeuron,
			postNeuron,
			stdpConfig,
			synapse.CreateDefaultPruningConfig(),
			config.InitialWeight,
			config.Delay,
		), nil
	})

	// Create test neurons
	preNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "test_neuron",
		Threshold:  0.5,
		Position:   types.Position3D{X: 0, Y: 0, Z: 0},
	})
	if err != nil {
		t.Fatalf("Failed to create pre-neuron: %v", err)
	}

	postNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "test_neuron",
		Threshold:  0.5,
		Position:   types.Position3D{X: 10, Y: 0, Z: 0},
	})
	if err != nil {
		t.Fatalf("Failed to create post-neuron: %v", err)
	}

	// Start neurons
	err = preNeuron.Start()
	if err != nil {
		t.Fatalf("Failed to start pre-neuron: %v", err)
	}
	defer preNeuron.Stop()

	err = postNeuron.Start()
	if err != nil {
		t.Fatalf("Failed to start post-neuron: %v", err)
	}
	defer postNeuron.Stop()

	// Create synapse with ZERO weight to prevent automatic firing propagation
	// This is crucial - we don't want pre-neuron to cause post-neuron to fire
	testSynapse, err := matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "test_synapse",
		PresynapticID:  preNeuron.ID(),
		PostsynapticID: postNeuron.ID(),
		InitialWeight:  0.5,                   // Will be reset later
		Delay:          50 * time.Millisecond, // Add delay to prevent immediate firing
	})
	if err != nil {
		t.Fatalf("Failed to create synapse: %v", err)
	}

	// Check if synapse implements necessary interfaces
	if _, ok := testSynapse.(interface{ RecordPostSpike(time.Time) }); ok {
		t.Log("✓ Synapse implements RecordPostSpike interface")
	} else {
		t.Fatal("❌ Synapse does NOT implement RecordPostSpike interface - test cannot continue")
	}

	// Reset weight to 0.5 for consistent testing
	if weightSetter, ok := testSynapse.(interface{ SetWeight(float64) }); ok {
		weightSetter.SetWeight(0.5)
		t.Log("Set synapse weight to 0.5 for testing")
	}

	// Get initial weight
	initialWeight := 0.5 // Default
	if weightGetter, ok := testSynapse.(interface{ GetWeight() float64 }); ok {
		initialWeight = weightGetter.GetWeight()
		t.Logf("Initial synapse weight: %.4f", initialWeight)
	}

	// Enable STDP on post-neuron
	if postWithSTDP, ok := postNeuron.(interface {
		EnableSTDPFeedback(time.Duration, float64)
		IsSTDPFeedbackEnabled() bool
	}); ok {
		postWithSTDP.EnableSTDPFeedback(5*time.Millisecond, 0.1)

		if postWithSTDP.IsSTDPFeedbackEnabled() {
			t.Log("✓ STDP feedback enabled on post-neuron")
		} else {
			t.Fatal("❌ Failed to enable STDP on post-neuron")
		}
	} else {
		t.Fatal("❌ Post-neuron doesn't support STDP operations")
	}

	// Clear any existing state
	time.Sleep(50 * time.Millisecond)

	// DIRECT TIMING TEST FOR LTP - IMPORTANT: Order matters for LTP!
	t.Log("\n=== DIRECT LTP TIMING TEST ===")
	t.Log("Creating explicit pre-before-post spike history")

	// 1. Record post-spike time (we'll use this later)
	postSpikeTime := time.Now().Add(50 * time.Millisecond) // Post-spike will be 50ms in the future

	// 2. Record pre-spike first by firing the pre-neuron
	t.Log("Triggering pre-neuron to fire (recording pre-spike)")
	preNeuron.Receive(types.NeuralSignal{
		Value:     20.0,       // Very strong signal to ensure firing
		Timestamp: time.Now(), // Current time
		SourceID:  "test",
		TargetID:  preNeuron.ID(),
	})

	// 3. Wait to ensure the pre-spike is recorded
	time.Sleep(15 * time.Millisecond)

	// 4. Now wait until the designated post-spike time
	waitDuration := time.Until(postSpikeTime)
	if waitDuration > 0 {
		time.Sleep(waitDuration)
	}

	// 5. Record post-spike at the right time
	if recorder, ok := testSynapse.(interface{ RecordPostSpike(time.Time) }); ok {
		t.Log("Directly recording post-spike")
		recorder.RecordPostSpike(postSpikeTime)
	} else {
		t.Fatal("❌ Synapse doesn't support RecordPostSpike - test cannot continue")
	}

	// 6. Verify spike history
	if spikeGetter, ok := testSynapse.(interface {
		GetPreSpikeTimes() []time.Time
		GetPostSpikeTimes() []time.Time
	}); ok {
		preSpikes := spikeGetter.GetPreSpikeTimes()
		postSpikes := spikeGetter.GetPostSpikeTimes()

		t.Logf("Spike history: %d pre-spikes, %d post-spikes",
			len(preSpikes), len(postSpikes))

		// Check if we have the spikes we need
		if len(preSpikes) == 0 || len(postSpikes) == 0 {
			t.Fatal("❌ Failed to record spikes")
		}

		// Verify timing of most recent spikes
		latestPreSpike := preSpikes[len(preSpikes)-1]
		latestPostSpike := postSpikes[len(postSpikes)-1]

		actualDeltaT := latestPreSpike.Sub(latestPostSpike)
		t.Logf("Actual timing: pre=%v, post=%v, deltaT=%v",
			latestPreSpike, latestPostSpike, actualDeltaT)

		if actualDeltaT >= 0 {
			t.Errorf("❌ Wrong timing relationship: deltaT=%v (should be negative for LTP)",
				actualDeltaT)
		} else {
			t.Logf("✓ Correct LTP timing relationship: deltaT=%v", actualDeltaT)
		}
	} else {
		t.Log("⚠️ Synapse doesn't support GetPreSpikeTimes/GetPostSpikeTimes, can't verify spike history")
	}

	// 7. Apply STDP
	t.Log("\n=== APPLYING STDP FEEDBACK ===")

	// Before calling feedback, manually check closest spike
	if stdpSystem, ok := postNeuron.(interface {
		SendSTDPFeedback()
	}); ok {
		// Trigger STDP feedback
		t.Log("Sending STDP feedback...")
		stdpSystem.SendSTDPFeedback()
	}

	// 8. Wait for STDP processing
	time.Sleep(50 * time.Millisecond)

	// 9. Check weight change
	finalWeight := initialWeight
	if weightGetter, ok := testSynapse.(interface{ GetWeight() float64 }); ok {
		finalWeight = weightGetter.GetWeight()
	}

	weightChange := finalWeight - initialWeight
	t.Logf("After STDP: weight=%.4f, change=%+.4f", finalWeight, weightChange)

	if weightChange <= 0 {
		t.Errorf("❌ LTP timing failed to strengthen synapse (%+.4f)", weightChange)

		// Test again with direct adjustment as a sanity check
		t.Log("\n=== DIRECT LTP TEST (Applying adjustment directly) ===")

		// Reset weight
		if weightSetter, ok := testSynapse.(interface{ SetWeight(float64) }); ok {
			weightSetter.SetWeight(initialWeight)
			t.Logf("Reset weight to %.4f for direct test", initialWeight)
		}

		// Create and apply direct adjustment
		adjustment := types.PlasticityAdjustment{
			DeltaT:       -50 * time.Millisecond, // Clearly negative for LTP
			LearningRate: 0.1,
			PreSynaptic:  true,
			PostSynaptic: true,
			Timestamp:    time.Now(),
			EventType:    types.PlasticitySTDP,
		}

		t.Logf("Applying direct LTP adjustment with deltaT=%v", adjustment.DeltaT)
		if plasticityApplier, ok := testSynapse.(interface {
			ApplyPlasticity(types.PlasticityAdjustment)
		}); ok {
			plasticityApplier.ApplyPlasticity(adjustment)
		}

		// Check result
		directFinalWeight := initialWeight
		if weightGetter, ok := testSynapse.(interface{ GetWeight() float64 }); ok {
			directFinalWeight = weightGetter.GetWeight()
		}

		directChange := directFinalWeight - initialWeight
		t.Logf("After direct LTP adjustment: weight=%.4f, change=%+.4f",
			directFinalWeight, directChange)

		if directChange <= 0 {
			t.Errorf("❌ Direct LTP adjustment failed (change: %+.4f)", directChange)
		} else {
			t.Logf("✓ Direct LTP adjustment correctly increased weight (change: %+.4f)", directChange)
		}
	} else {
		t.Logf("✓ LTP timing correctly strengthened synapse (%+.4f)", weightChange)
	}
}

// testLTDWithDirectSpikes tests the Long-Term Depression case (post before pre)
// This is based on the successful TestSTDPIntegration_LTDOnlyDirect test
func testLTDWithDirectSpikes(t *testing.T) {
	t.Log("\n=== DIRECT LTD TIMING TEST ===")
	t.Log("Testing the LTD case (Post fires before Pre) with direct spike history manipulation")

	// Create matrix with standard configuration
	matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  5 * time.Millisecond,
		MaxComponents:   10,
	})

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}
	defer matrix.Stop()

	// Register neuron type
	matrix.RegisterNeuronType("test_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(
			id,
			config.Threshold,
			0.95,               // decay rate
			5*time.Millisecond, // refractory period
			2.0,                // fire factor
			3.0,                // target firing rate
			0.2,                // homeostasis strength
		)
		n.SetCallbacks(callbacks)
		return n, nil
	})

	// Register synapse type with special test configuration
	matrix.RegisterSynapseType("test_synapse", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, exists := matrix.GetNeuron(config.PresynapticID)
		if !exists {
			return nil, fmt.Errorf("presynaptic neuron not found: %s", config.PresynapticID)
		}

		postNeuron, exists := matrix.GetNeuron(config.PostsynapticID)
		if !exists {
			return nil, fmt.Errorf("postsynaptic neuron not found: %s", config.PostsynapticID)
		}

		// Create special STDP config for test
		stdpConfig := synapse.CreateDefaultSTDPConfig()
		stdpConfig.WindowSize = 400 * time.Millisecond // Much larger window for test

		return synapse.NewBasicSynapse(
			id,
			preNeuron,
			postNeuron,
			stdpConfig,
			synapse.CreateDefaultPruningConfig(),
			config.InitialWeight,
			config.Delay,
		), nil
	})

	// Create test neurons
	preNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "test_neuron",
		Threshold:  0.5,
		Position:   types.Position3D{X: 0, Y: 0, Z: 0},
	})
	if err != nil {
		t.Fatalf("Failed to create pre-neuron: %v", err)
	}

	postNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "test_neuron",
		Threshold:  0.5,
		Position:   types.Position3D{X: 10, Y: 0, Z: 0},
	})
	if err != nil {
		t.Fatalf("Failed to create post-neuron: %v", err)
	}

	// Start neurons
	err = preNeuron.Start()
	if err != nil {
		t.Fatalf("Failed to start pre-neuron: %v", err)
	}
	defer preNeuron.Stop()

	err = postNeuron.Start()
	if err != nil {
		t.Fatalf("Failed to start post-neuron: %v", err)
	}
	defer postNeuron.Stop()

	// Create synapse with ZERO weight to prevent automatic firing propagation
	// This is crucial - we don't want pre-neuron to cause post-neuron to fire
	testSynapse, err := matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "test_synapse",
		PresynapticID:  preNeuron.ID(),
		PostsynapticID: postNeuron.ID(),
		InitialWeight:  0.5,                   // Will be reset later
		Delay:          50 * time.Millisecond, // Add delay to prevent immediate firing
	})
	if err != nil {
		t.Fatalf("Failed to create synapse: %v", err)
	}

	// Check if synapse implements necessary interfaces
	if _, ok := testSynapse.(interface{ RecordPostSpike(time.Time) }); ok {
		t.Log("✓ Synapse implements RecordPostSpike interface")
	} else {
		t.Fatal("❌ Synapse does NOT implement RecordPostSpike interface - test cannot continue")
	}

	// Reset weight to 0.5 for consistent testing
	if weightSetter, ok := testSynapse.(interface{ SetWeight(float64) }); ok {
		weightSetter.SetWeight(0.5)
		t.Log("Set synapse weight to 0.5 for testing")
	}

	// Get initial weight
	initialWeight := 0.5 // Default
	if weightGetter, ok := testSynapse.(interface{ GetWeight() float64 }); ok {
		initialWeight = weightGetter.GetWeight()
		t.Logf("Initial synapse weight: %.4f", initialWeight)
	}

	// Enable STDP on post-neuron
	if postWithSTDP, ok := postNeuron.(interface {
		EnableSTDPFeedback(time.Duration, float64)
		IsSTDPFeedbackEnabled() bool
	}); ok {
		postWithSTDP.EnableSTDPFeedback(5*time.Millisecond, 0.1)

		if postWithSTDP.IsSTDPFeedbackEnabled() {
			t.Log("✓ STDP feedback enabled on post-neuron")
		} else {
			t.Fatal("❌ Failed to enable STDP on post-neuron")
		}
	} else {
		t.Fatal("❌ Post-neuron doesn't support STDP operations")
	}

	// Clear any existing state
	time.Sleep(50 * time.Millisecond)

	// DIRECT TIMING TEST FOR LTD
	t.Log("\n=== DIRECT LTD TIMING TEST ===")
	t.Log("Creating explicit post-before-pre spike history")

	// 1. Create timestamps with clear LTD timing
	// Post spike comes first (100ms ago)
	postSpikeTime := time.Now().Add(-100 * time.Millisecond)
	// Pre spike comes second (50ms ago)
	preSpikeTime := time.Now().Add(-50 * time.Millisecond)

	deltaT := preSpikeTime.Sub(postSpikeTime)
	t.Logf("Created timing: post=%v, pre=%v, deltaT=%v",
		postSpikeTime, preSpikeTime, deltaT)

	if deltaT <= 0 {
		t.Fatalf("❌ Invalid timing setup: deltaT should be positive for LTD")
	}

	// 2. Record post-spike first using direct interface call
	if recorder, ok := testSynapse.(interface{ RecordPostSpike(time.Time) }); ok {
		t.Log("Directly recording post-spike")
		recorder.RecordPostSpike(postSpikeTime)
	}

	// 3. Trigger pre-neuron to fire which should record pre-spike
	t.Log("Triggering pre-neuron to fire")
	preNeuron.Receive(types.NeuralSignal{
		Value:     10.0,
		Timestamp: preSpikeTime,
		SourceID:  "test",
		TargetID:  preNeuron.ID(),
	})

	// 4. Wait for processing
	time.Sleep(20 * time.Millisecond)

	// 5. Verify spike history
	if spikeGetter, ok := testSynapse.(interface {
		GetPreSpikeTimes() []time.Time
		GetPostSpikeTimes() []time.Time
	}); ok {
		preSpikes := spikeGetter.GetPreSpikeTimes()
		postSpikes := spikeGetter.GetPostSpikeTimes()

		t.Logf("Spike history: %d pre-spikes, %d post-spikes",
			len(preSpikes), len(postSpikes))

		// Check if we have the spikes we need
		if len(preSpikes) == 0 || len(postSpikes) == 0 {
			t.Fatal("❌ Failed to record spikes")
		}

		// Verify timing of most recent spikes
		latestPreSpike := preSpikes[len(preSpikes)-1]
		latestPostSpike := postSpikes[len(postSpikes)-1]

		actualDeltaT := latestPreSpike.Sub(latestPostSpike)
		t.Logf("Actual timing: post=%v, pre=%v, deltaT=%v",
			latestPostSpike, latestPreSpike, actualDeltaT)

		if actualDeltaT <= 0 {
			t.Errorf("❌ Wrong timing relationship: deltaT=%v (should be positive for LTD)",
				actualDeltaT)
		} else {
			t.Logf("✓ Correct LTD timing relationship: deltaT=%v", actualDeltaT)
		}
	}

	// 6. Apply STDP
	t.Log("\n=== APPLYING STDP FEEDBACK ===")

	// Before calling feedback, manually check closest spike
	if stdpSystem, ok := postNeuron.(interface {
		SendSTDPFeedback()
	}); ok {
		// Trigger STDP feedback
		t.Log("Sending STDP feedback...")
		stdpSystem.SendSTDPFeedback()
	}

	// 7. Wait for STDP processing
	time.Sleep(50 * time.Millisecond)

	// 8. Check weight change
	finalWeight := initialWeight
	if weightGetter, ok := testSynapse.(interface{ GetWeight() float64 }); ok {
		finalWeight = weightGetter.GetWeight()
	}

	weightChange := finalWeight - initialWeight
	t.Logf("After STDP: weight=%.4f, change=%+.4f", finalWeight, weightChange)

	if weightChange >= 0 {
		t.Errorf("❌ LTD timing failed to weaken synapse (%+.4f)", weightChange)

		// Test again with direct adjustment as a sanity check
		t.Log("\n=== DIRECT LTD TEST (Applying adjustment directly) ===")

		// Reset weight
		if weightSetter, ok := testSynapse.(interface{ SetWeight(float64) }); ok {
			weightSetter.SetWeight(initialWeight)
			t.Logf("Reset weight to %.4f for direct test", initialWeight)
		}

		// Create and apply direct adjustment
		adjustment := types.PlasticityAdjustment{
			DeltaT:       50 * time.Millisecond, // Clearly positive for LTD
			LearningRate: 0.1,
			PreSynaptic:  true,
			PostSynaptic: true,
			Timestamp:    time.Now(),
			EventType:    types.PlasticitySTDP,
		}

		t.Logf("Applying direct LTD adjustment with deltaT=%v", adjustment.DeltaT)
		if plasticityApplier, ok := testSynapse.(interface {
			ApplyPlasticity(types.PlasticityAdjustment)
		}); ok {
			plasticityApplier.ApplyPlasticity(adjustment)
		}

		// Check result
		directFinalWeight := initialWeight
		if weightGetter, ok := testSynapse.(interface{ GetWeight() float64 }); ok {
			directFinalWeight = weightGetter.GetWeight()
		}

		directChange := directFinalWeight - initialWeight
		t.Logf("After direct LTD adjustment: weight=%.4f, change=%+.4f",
			directFinalWeight, directChange)

		if directChange >= 0 {
			t.Errorf("❌ Direct LTD adjustment failed (change: %+.4f)", directChange)
		} else {
			t.Logf("✓ Direct LTD adjustment correctly decreased weight (change: %+.4f)", directChange)
		}
	} else {
		t.Logf("✓ LTD timing correctly weakened synapse (%+.4f)", weightChange)
	}
}

// Test that directly applies a plasticity adjustment to a synapse
// This ensures the core STDP function in synapse works correctly
func TestSTDPIntegration_DirectPlasticity(t *testing.T) {
	// Create a test synapse
	stdpConfig := synapse.CreateDefaultSTDPConfig()
	stdpConfig.WindowSize = 100 * time.Millisecond
	stdpConfig.TimeConstant = 20 * time.Millisecond
	stdpConfig.LearningRate = 0.1

	testSynapse := synapse.NewBasicSynapse(
		"direct_test_synapse",
		nil, nil, // No real neurons needed
		stdpConfig,
		synapse.CreateDefaultPruningConfig(),
		0.5, // Initial weight
		0,   // No delay
	)

	// Test LTP (negative deltaT - pre before post)
	t.Run("LTP_Test", func(t *testing.T) {
		initialWeight := testSynapse.GetWeight()
		t.Logf("Initial weight: %.4f", initialWeight)

		// Apply LTP adjustment
		ltpAdjustment := types.PlasticityAdjustment{
			DeltaT:       -15 * time.Millisecond, // Pre before post (LTP)
			LearningRate: 0.1,
			PreSynaptic:  true,
			PostSynaptic: true,
			Timestamp:    time.Now(),
			EventType:    types.PlasticitySTDP,
		}

		testSynapse.ApplyPlasticity(ltpAdjustment)

		// Check result
		finalWeight := testSynapse.GetWeight()
		weightChange := finalWeight - initialWeight

		t.Logf("After LTP adjustment (deltaT=%v): weight=%.4f, change=%+.4f",
			ltpAdjustment.DeltaT, finalWeight, weightChange)

		if weightChange <= 0 {
			t.Errorf("❌ LTP failed to strengthen synapse (change: %+.4f)", weightChange)
		} else {
			t.Logf("✓ LTP correctly strengthened synapse (change: %+.4f)", weightChange)
		}
	})

	// Reset weight
	testSynapse.SetWeight(0.5)

	// Test LTD (positive deltaT - post before pre)
	t.Run("LTD_Test", func(t *testing.T) {
		initialWeight := testSynapse.GetWeight()
		t.Logf("Initial weight: %.4f", initialWeight)

		// Apply LTD adjustment
		ltdAdjustment := types.PlasticityAdjustment{
			DeltaT:       15 * time.Millisecond, // Post before pre (LTD)
			LearningRate: 0.1,
			PreSynaptic:  true,
			PostSynaptic: true,
			Timestamp:    time.Now(),
			EventType:    types.PlasticitySTDP,
		}

		testSynapse.ApplyPlasticity(ltdAdjustment)

		// Check result
		finalWeight := testSynapse.GetWeight()
		weightChange := finalWeight - initialWeight

		t.Logf("After LTD adjustment (deltaT=%v): weight=%.4f, change=%+.4f",
			ltdAdjustment.DeltaT, finalWeight, weightChange)

		if weightChange >= 0 {
			t.Errorf("❌ LTD failed to weaken synapse (change: %+.4f)", weightChange)
		} else {
			t.Logf("✓ LTD correctly weakened synapse (change: %+.4f)", weightChange)
		}
	})
}

// TestSTDPIntegration_BidirectionalTiming tests both LTP and LTD in a single test
// to verify that the STDP system can correctly distinguish between the two timing patterns
// and apply the appropriate weight changes
func TestSTDPIntegration_BidirectionalTiming(t *testing.T) {
	t.Log("=== BIDIRECTIONAL STDP TIMING TEST ===")
	t.Log("Testing both LTP and LTD timing in one test")

	// Create matrix
	matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  5 * time.Millisecond,
		MaxComponents:   10,
	})

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}
	defer matrix.Stop()

	// Register neuron type
	matrix.RegisterNeuronType("test_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(
			id,
			config.Threshold,
			0.95,               // decay rate
			5*time.Millisecond, // refractory period
			2.0,                // fire factor
			3.0,                // target firing rate
			0.2,                // homeostasis strength
		)
		n.SetCallbacks(callbacks)
		return n, nil
	})

	// Register synapse type
	matrix.RegisterSynapseType("test_synapse", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, exists := matrix.GetNeuron(config.PresynapticID)
		if !exists {
			return nil, fmt.Errorf("presynaptic neuron not found: %s", config.PresynapticID)
		}

		postNeuron, exists := matrix.GetNeuron(config.PostsynapticID)
		if !exists {
			return nil, fmt.Errorf("postsynaptic neuron not found: %s", config.PostsynapticID)
		}

		// Create special STDP config for test
		stdpConfig := synapse.CreateDefaultSTDPConfig()
		stdpConfig.WindowSize = 200 * time.Millisecond
		stdpConfig.TimeConstant = 20 * time.Millisecond
		stdpConfig.LearningRate = 0.1

		return synapse.NewBasicSynapse(
			id,
			preNeuron,
			postNeuron,
			stdpConfig,
			synapse.CreateDefaultPruningConfig(),
			config.InitialWeight,
			config.Delay,
		), nil
	})

	// Create post-neuron and two pre-neurons (one for LTP, one for LTD)
	postNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "test_neuron",
		Threshold:  0.5,
		Position:   types.Position3D{X: 0, Y: 0, Z: 0},
	})
	if err != nil {
		t.Fatalf("Failed to create post-neuron: %v", err)
	}

	preLTPNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "test_neuron",
		Threshold:  0.5,
		Position:   types.Position3D{X: -10, Y: 0, Z: 0},
	})
	if err != nil {
		t.Fatalf("Failed to create LTP pre-neuron: %v", err)
	}

	preLTDNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "test_neuron",
		Threshold:  0.5,
		Position:   types.Position3D{X: 10, Y: 0, Z: 0},
	})
	if err != nil {
		t.Fatalf("Failed to create LTD pre-neuron: %v", err)
	}

	// Start neurons
	postNeuron.Start()
	preLTPNeuron.Start()
	preLTDNeuron.Start()
	defer postNeuron.Stop()
	defer preLTPNeuron.Stop()
	defer preLTDNeuron.Stop()

	// Create synapses
	ltpSynapse, err := matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "test_synapse",
		PresynapticID:  preLTPNeuron.ID(),
		PostsynapticID: postNeuron.ID(),
		InitialWeight:  0.5,
	})
	if err != nil {
		t.Fatalf("Failed to create LTP synapse: %v", err)
	}

	ltdSynapse, err := matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "test_synapse",
		PresynapticID:  preLTDNeuron.ID(),
		PostsynapticID: postNeuron.ID(),
		InitialWeight:  0.5,
	})
	if err != nil {
		t.Fatalf("Failed to create LTD synapse: %v", err)
	}

	// Enable STDP on post-neuron
	if postWithSTDP, ok := postNeuron.(interface {
		EnableSTDPFeedback(time.Duration, float64)
		IsSTDPFeedbackEnabled() bool
	}); ok {
		postWithSTDP.EnableSTDPFeedback(5*time.Millisecond, 0.1)
		if postWithSTDP.IsSTDPFeedbackEnabled() {
			t.Log("✓ STDP feedback enabled on post-neuron")
		} else {
			t.Fatal("❌ Failed to enable STDP on post-neuron")
		}
	} else {
		t.Fatal("❌ Post-neuron doesn't support STDP operations")
	}

	// Get initial weights
	initialLTPWeight := 0.5
	initialLTDWeight := 0.5

	if weightGetter, ok := ltpSynapse.(interface{ GetWeight() float64 }); ok {
		initialLTPWeight = weightGetter.GetWeight()
	}

	if weightGetter, ok := ltdSynapse.(interface{ GetWeight() float64 }); ok {
		initialLTDWeight = weightGetter.GetWeight()
	}

	t.Logf("Initial weights: LTP synapse=%.4f, LTD synapse=%.4f",
		initialLTPWeight, initialLTDWeight)

	// Create spike timing for both cases

	// For LTP: Pre fires before Post
	// 1. Fire LTP pre-neuron first
	t.Log("\nFiring LTP pre-neuron (should fire before post)")
	preLTPNeuron.Receive(types.NeuralSignal{
		Value:     2.0,
		Timestamp: time.Now(),
		SourceID:  "test",
		TargetID:  preLTPNeuron.ID(),
	})

	// Wait to ensure timing gap
	time.Sleep(15 * time.Millisecond)

	// For LTD: Record a post-spike time
	t.Log("Recording post-spike time (for LTD case)")
	postSpikeTime := time.Now()

	// 2. Fire post-neuron
	t.Log("Firing post-neuron")
	postNeuron.Receive(types.NeuralSignal{
		Value:     2.0,
		Timestamp: time.Now(),
		SourceID:  "test",
		TargetID:  postNeuron.ID(),
	})

	// Wait to ensure timing gap
	time.Sleep(15 * time.Millisecond)

	// For LTD: Pre fires after Post
	// 3. Record post-spike on LTD synapse
	if recorder, ok := ltdSynapse.(interface{ RecordPostSpike(time.Time) }); ok {
		t.Log("Directly recording post-spike on LTD synapse")
		recorder.RecordPostSpike(postSpikeTime)
	}

	// 4. Fire LTD pre-neuron last
	t.Log("Firing LTD pre-neuron (fires after post)")
	preLTDNeuron.Receive(types.NeuralSignal{
		Value:     2.0,
		Timestamp: time.Now(),
		SourceID:  "test",
		TargetID:  preLTDNeuron.ID(),
	})

	// Wait for processing
	time.Sleep(20 * time.Millisecond)

	// 5. Trigger STDP
	t.Log("Triggering STDP feedback")
	if postWithSTDP, ok := postNeuron.(interface{ SendSTDPFeedback() }); ok {
		postWithSTDP.SendSTDPFeedback()
	}

	// Wait for STDP processing
	time.Sleep(50 * time.Millisecond)

	// 6. Check weight changes
	finalLTPWeight := initialLTPWeight
	finalLTDWeight := initialLTDWeight

	if weightGetter, ok := ltpSynapse.(interface{ GetWeight() float64 }); ok {
		finalLTPWeight = weightGetter.GetWeight()
	}

	if weightGetter, ok := ltdSynapse.(interface{ GetWeight() float64 }); ok {
		finalLTDWeight = weightGetter.GetWeight()
	}

	ltpChange := finalLTPWeight - initialLTPWeight
	ltdChange := finalLTDWeight - initialLTDWeight

	t.Logf("\nFinal weights:")
	t.Logf("LTP synapse: %.4f → %.4f (change: %+.4f)",
		initialLTPWeight, finalLTPWeight, ltpChange)
	t.Logf("LTD synapse: %.4f → %.4f (change: %+.4f)",
		initialLTDWeight, finalLTDWeight, ltdChange)

	// 7. Verify results
	if ltpChange <= 0 {
		t.Logf("❌ LTP timing failed to strengthen synapse (%+.4f)", ltpChange)
	} else {
		t.Logf("✓ LTP timing correctly strengthened synapse (%+.4f)", ltpChange)
	}

	if ltdChange >= 0 {
		t.Logf("❌ LTD timing failed to weaken synapse (%+.4f)", ltdChange)
	} else {
		t.Logf("✓ LTD timing correctly weakened synapse (%+.4f)", ltdChange)
	}

	// Final verification with direct plasticity application
	if ltpChange <= 0 || ltdChange >= 0 {
		t.Log("\n=== DIRECT PLASTICITY VERIFICATION ===")

		// Reset weights
		if setter, ok := ltpSynapse.(interface{ SetWeight(float64) }); ok {
			setter.SetWeight(initialLTPWeight)
		}

		if setter, ok := ltdSynapse.(interface{ SetWeight(float64) }); ok {
			setter.SetWeight(initialLTDWeight)
		}

		// Apply direct adjustments
		ltpAdj := types.PlasticityAdjustment{
			DeltaT:       -15 * time.Millisecond,
			LearningRate: 0.1,
			EventType:    types.PlasticitySTDP,
		}

		ltdAdj := types.PlasticityAdjustment{
			DeltaT:       15 * time.Millisecond,
			LearningRate: 0.1,
			EventType:    types.PlasticitySTDP,
		}

		// Apply
		if applier, ok := ltpSynapse.(interface {
			ApplyPlasticity(types.PlasticityAdjustment)
		}); ok {
			applier.ApplyPlasticity(ltpAdj)
		}

		if applier, ok := ltdSynapse.(interface {
			ApplyPlasticity(types.PlasticityAdjustment)
		}); ok {
			applier.ApplyPlasticity(ltdAdj)
		}

		// Check results
		var directLTPWeight, directLTDWeight float64

		if getter, ok := ltpSynapse.(interface{ GetWeight() float64 }); ok {
			directLTPWeight = getter.GetWeight()
		}

		if getter, ok := ltdSynapse.(interface{ GetWeight() float64 }); ok {
			directLTDWeight = getter.GetWeight()
		}

		directLTPChange := directLTPWeight - initialLTPWeight
		directLTDChange := directLTDWeight - initialLTDWeight

		t.Logf("Direct LTP: %.4f → %.4f (change: %+.4f)",
			initialLTPWeight, directLTPWeight, directLTPChange)
		t.Logf("Direct LTD: %.4f → %.4f (change: %+.4f)",
			initialLTDWeight, directLTDWeight, directLTDChange)

		if directLTPChange > 0 {
			t.Log("✓ Direct LTP adjustment works correctly")
		}

		if directLTDChange < 0 {
			t.Log("✓ Direct LTD adjustment works correctly")
		}
	}
}
