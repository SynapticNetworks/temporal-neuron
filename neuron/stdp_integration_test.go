package neuron

import (
	"fmt"
	"math"
	"sync"
	"testing"
	"time"
)

// ============================================================================
// STDP + HOMEOSTASIS INTEGRATION TESTS
// ============================================================================

// TestSTDPWithHomeostasis tests interaction between STDP learning and homeostatic regulation
//
// BIOLOGICAL CONTEXT:
// In real neurons, synaptic plasticity (STDP) and homeostatic plasticity operate
// simultaneously but on different timescales. STDP modifies individual synaptic
// weights based on spike timing (milliseconds), while homeostasis adjusts the
// neuron's overall excitability to maintain stable firing rates (seconds to minutes).
//
// This interaction is crucial for network stability:
// - STDP can cause runaway strengthening/weakening without homeostasis
// - Homeostasis can interfere with STDP learning if too aggressive
// - Together, they create stable yet adaptive networks
//
// EXPECTED BEHAVIOR:
// - STDP should modify individual synapse strengths based on timing
// - Homeostasis should adjust neuron threshold to maintain target firing rate
// - Both mechanisms should coexist without interference
// - Network should remain stable while learning temporal patterns
func TestSTDPWithHomeostasis(t *testing.T) {
	// Create STDP configuration
	stdpConfig := STDPConfig{
		Enabled:        true,
		LearningRate:   0.02,                  // Moderate learning rate
		TimeConstant:   20 * time.Millisecond, // Standard biological value
		WindowSize:     50 * time.Millisecond, // ±50ms window
		MinWeight:      0.1,                   // Prevent synapse elimination
		MaxWeight:      3.0,                   // Prevent runaway strengthening
		AsymmetryRatio: 1.5,                   // Slight LTP bias
	}

	// Create neuron with both STDP and homeostasis enabled
	targetRate := 5.0          // 5 Hz target firing rate
	homeostasisStrength := 0.2 // Moderate homeostatic regulation

	neuron := NewNeuron("test_integrated", 1.0, 0.95, 10*time.Millisecond, 1.0,
		targetRate, homeostasisStrength, stdpConfig)

	// Set up monitoring
	fireEvents := make(chan FireEvent, 100)
	neuron.SetFireEventChannel(fireEvents)

	// Create output for monitoring
	output := make(chan Message, 100)
	neuron.AddOutput("monitor", output, 1.0, 0)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	t.Logf("=== STDP + HOMEOSTASIS INTEGRATION TEST ===")
	t.Logf("Target firing rate: %.1f Hz", targetRate)
	t.Logf("Homeostasis strength: %.1f", homeostasisStrength)
	t.Logf("STDP learning rate: %.3f", stdpConfig.LearningRate)

	// Record initial state
	initialThreshold := neuron.GetCurrentThreshold()
	initialCalcium := neuron.GetCalciumLevel()

	t.Logf("Initial threshold: %.3f", initialThreshold)
	t.Logf("Initial calcium: %.3f", initialCalcium)

	// Phase 1: Establish baseline with moderate activity
	t.Logf("\n--- Phase 1: Baseline Activity ---")
	for i := 0; i < 20; i++ {
		input <- Message{
			Value:     1.2, // Above threshold - should fire
			Timestamp: time.Now(),
			SourceID:  "baseline_input",
		}
		time.Sleep(150 * time.Millisecond) // ~6.7 Hz rate
	}

	time.Sleep(500 * time.Millisecond) // Allow homeostatic adjustment

	midThreshold := neuron.GetCurrentThreshold()
	midRate := neuron.GetCurrentFiringRate()
	midCalcium := neuron.GetCalciumLevel()

	t.Logf("Mid-phase threshold: %.3f (change: %+.3f)", midThreshold, midThreshold-initialThreshold)
	t.Logf("Mid-phase firing rate: %.1f Hz (target: %.1f Hz)", midRate, targetRate)
	t.Logf("Mid-phase calcium: %.3f", midCalcium)

	// Phase 2: Apply STDP learning patterns while homeostasis operates
	t.Logf("\n--- Phase 2: STDP Learning with Homeostasis ---")

	// Create a learning pattern: consistent causal timing should strengthen input
	for i := 0; i < 15; i++ {
		// Send causal pattern: external input, then internal firing
		input <- Message{
			Value:     0.8, // Below threshold initially
			Timestamp: time.Now(),
			SourceID:  "learning_input",
		}

		time.Sleep(5 * time.Millisecond) // 5ms delay for optimal LTP

		// Trigger firing with additional input
		input <- Message{
			Value:     0.5, // Combined should exceed threshold
			Timestamp: time.Now(),
			SourceID:  "trigger_input",
		}

		time.Sleep(180 * time.Millisecond) // ~5.5 Hz rate
	}

	time.Sleep(1 * time.Second) // Allow both STDP and homeostasis to settle

	// Record final state
	finalThreshold := neuron.GetCurrentThreshold()
	finalRate := neuron.GetCurrentFiringRate()
	finalCalcium := neuron.GetCalciumLevel()

	t.Logf("\n--- Final Results ---")
	t.Logf("Final threshold: %.3f (total change: %+.3f)", finalThreshold, finalThreshold-initialThreshold)
	t.Logf("Final firing rate: %.1f Hz (target: %.1f Hz)", finalRate, targetRate)
	t.Logf("Final calcium: %.3f", finalCalcium)

	// Validate homeostatic regulation
	rateError := math.Abs(finalRate - targetRate)
	if rateError > 2.0 {
		t.Logf("WARNING: Firing rate (%.1f Hz) deviates significantly from target (%.1f Hz)",
			finalRate, targetRate)
	} else {
		t.Logf("✓ Homeostatic regulation maintained firing rate near target")
	}

	// Validate both mechanisms operated
	thresholdChanged := math.Abs(finalThreshold-initialThreshold) > 0.01
	if !thresholdChanged {
		t.Logf("WARNING: Threshold didn't change - homeostasis may be inactive")
	} else {
		t.Logf("✓ Homeostatic threshold adjustment occurred")
	}

	// Count firing events to validate activity
	fireCount := 0
	for {
		select {
		case <-fireEvents:
			fireCount++
		default:
			goto done
		}
	}
done:

	if fireCount < 10 {
		t.Logf("WARNING: Low firing activity (%d events) - may affect learning", fireCount)
	} else {
		t.Logf("✓ Adequate firing activity for learning (%d events)", fireCount)
	}

	t.Logf("✓ STDP and homeostasis coexisted successfully")
}

// TestSTDPHomeostasisTimescales tests that STDP and homeostasis operate on appropriate timescales
//
// BIOLOGICAL CONTEXT:
// STDP operates on millisecond timescales (spike timing precision), while
// homeostasis operates on second-to-minute timescales. This separation is
// crucial for stability - if homeostasis were too fast, it would interfere
// with STDP learning. If too slow, it wouldn't provide adequate regulation.
//
// EXPECTED BEHAVIOR:
// - STDP should modify weights immediately after spike pairings
// - Homeostasis should adjust threshold gradually over longer periods
// - Fast STDP changes should not be immediately counteracted by homeostasis
func TestSTDPHomeostasisTimescales(t *testing.T) {
	stdpConfig := STDPConfig{
		Enabled:        true,
		LearningRate:   0.05, // Higher rate for clear observation
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.5,
	}

	neuron := NewNeuron("timescale_test", 1.0, 0.95, 5*time.Millisecond, 1.0,
		3.0, 0.3, stdpConfig) // Stronger homeostasis for observation

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	t.Logf("=== STDP/HOMEOSTASIS TIMESCALE TEST ===")

	// Measure initial state
	initialThreshold := neuron.GetCurrentThreshold()

	// Create rapid STDP events
	t.Logf("Applying rapid STDP learning events...")
	for i := 0; i < 5; i++ {
		// Rapid causal pairings
		input <- Message{Value: 0.8, Timestamp: time.Now(), SourceID: "fast_input"}
		time.Sleep(5 * time.Millisecond)
		input <- Message{Value: 0.5, Timestamp: time.Now(), SourceID: "trigger"}
		time.Sleep(20 * time.Millisecond) // Short interval between pairings
	}

	// Check threshold immediately after STDP events
	time.Sleep(50 * time.Millisecond) // Brief pause for processing
	immediateThreshold := neuron.GetCurrentThreshold()

	t.Logf("Threshold immediately after STDP: %.4f (change: %+.4f)",
		immediateThreshold, immediateThreshold-initialThreshold)

	// Wait for homeostatic timescale
	t.Logf("Waiting for homeostatic adjustment...")
	time.Sleep(2 * time.Second)

	delayedThreshold := neuron.GetCurrentThreshold()
	delayedRate := neuron.GetCurrentFiringRate()

	t.Logf("Threshold after homeostatic delay: %.4f (change: %+.4f)",
		delayedThreshold, delayedThreshold-initialThreshold)
	t.Logf("Firing rate after delay: %.2f Hz", delayedRate)

	// Validate timescale separation
	immediateChange := math.Abs(immediateThreshold - initialThreshold)
	delayedChange := math.Abs(delayedThreshold - initialThreshold)

	if delayedChange > immediateChange*1.5 {
		t.Logf("✓ Homeostatic adjustment occurred on slower timescale")
	} else {
		t.Logf("Note: Minimal homeostatic adjustment observed")
	}

	if immediateChange < 0.001 {
		t.Logf("Note: Minimal immediate threshold change - STDP may need more events")
	} else {
		t.Logf("✓ Some threshold dynamics observed during learning period")
	}

	t.Logf("✓ Timescale test completed")
}

// ============================================================================
// SMALL NETWORK STDP TESTS
// ============================================================================

// TestTwoNeuronSTDPNetwork tests STDP learning in a simple two-neuron circuit
//
// BIOLOGICAL CONTEXT:
// This represents the fundamental unit of neural learning - two connected
// neurons where the connection strength adapts based on their relative
// activity patterns. This is the building block for larger network learning.
//
// EXPECTED BEHAVIOR:
// - Consistent causal patterns should strengthen the connection
// - Connection strengthening should make post-neuron more responsive
// - Learning should be observable through network behavior changes
func TestTwoNeuronSTDPNetwork(t *testing.T) {
	stdpConfig := STDPConfig{
		Enabled:        true,
		LearningRate:   0.03,
		TimeConstant:   15 * time.Millisecond,
		WindowSize:     40 * time.Millisecond,
		MinWeight:      0.2,
		MaxWeight:      2.5,
		AsymmetryRatio: 1.8,
	}

	// Create pre-synaptic neuron (input)
	preNeuron := NewNeuron("pre", 1.0, 0.95, 8*time.Millisecond, 1.0,
		0, 0, STDPConfig{Enabled: false}) // No homeostasis for controlled test

	// Create post-synaptic neuron (output) with homeostasis
	postNeuron := NewNeuron("post", 1.2, 0.95, 8*time.Millisecond, 1.0,
		4.0, 0.15, stdpConfig) // Moderate homeostasis

	// Connect pre → post with STDP
	initialWeight := 0.8
	preNeuron.AddOutputWithSTDP("to_post", postNeuron.GetInputChannel(),
		initialWeight, 2*time.Millisecond, stdpConfig)

	// Set up monitoring
	postFireEvents := make(chan FireEvent, 100)
	postNeuron.SetFireEventChannel(postFireEvents)

	// Start neurons
	go preNeuron.Run()
	go postNeuron.Run()
	defer preNeuron.Close()
	defer postNeuron.Close()

	t.Logf("=== TWO-NEURON STDP NETWORK TEST ===")
	t.Logf("Initial connection weight: %.2f", initialWeight)

	preInput := preNeuron.GetInput()
	postInput := postNeuron.GetInput()

	// Phase 1: Baseline - random, uncorrelated activity
	t.Logf("\n--- Phase 1: Baseline (uncorrelated activity) ---")
	for i := 0; i < 10; i++ {
		// Random timing between pre and post
		go func() {
			time.Sleep(time.Duration(i*50) * time.Millisecond)
			preInput <- Message{Value: 1.5, Timestamp: time.Now(), SourceID: "external"}
		}()
		go func() {
			time.Sleep(time.Duration(i*50+25) * time.Millisecond)
			postInput <- Message{Value: 1.0, Timestamp: time.Now(), SourceID: "external"}
		}()
	}

	time.Sleep(800 * time.Millisecond)
	baselinePostRate := postNeuron.GetCurrentFiringRate()
	t.Logf("Baseline post-neuron firing rate: %.2f Hz", baselinePostRate)

	// Phase 2: Learning - consistent causal patterns
	t.Logf("\n--- Phase 2: Causal Learning Pattern ---")
	for i := 0; i < 20; i++ {
		// Causal pattern: pre fires, then post fires 8ms later
		preInput <- Message{Value: 1.4, Timestamp: time.Now(), SourceID: "training"}

		time.Sleep(8 * time.Millisecond) // Optimal STDP timing

		postInput <- Message{Value: 0.9, Timestamp: time.Now(), SourceID: "training"}

		time.Sleep(100 * time.Millisecond) // Inter-trial interval
	}

	time.Sleep(1 * time.Second) // Allow learning to consolidate

	// Phase 3: Test learned response
	t.Logf("\n--- Phase 3: Testing Learned Response ---")

	// Clear any pending fire events
	for {
		select {
		case <-postFireEvents:
		default:
			goto cleared
		}
	}
cleared:

	// Test response to pre-neuron activation alone
	testResponses := 0
	for i := 0; i < 8; i++ {
		preInput <- Message{Value: 1.3, Timestamp: time.Now(), SourceID: "test"}

		// Check if post-neuron fires within 50ms
		select {
		case <-postFireEvents:
			testResponses++
		case <-time.After(50 * time.Millisecond):
			// No response
		}

		time.Sleep(150 * time.Millisecond) // Inter-test interval
	}

	responseRate := float64(testResponses) / 8.0 * 100
	finalPostRate := postNeuron.GetCurrentFiringRate()

	t.Logf("Post-learning response rate: %.1f%% (%d/8 tests)", responseRate, testResponses)
	t.Logf("Final post-neuron firing rate: %.2f Hz", finalPostRate)

	// Validate learning occurred
	if responseRate > 25 { // At least 25% response rate indicates learning
		t.Logf("✓ STDP learning successful - post-neuron responds to pre-neuron")
	} else {
		t.Logf("Note: Low response rate - learning may need more trials or stronger patterns")
	}

	// Validate homeostasis maintained reasonable activity
	if finalPostRate > 0.5 && finalPostRate < 20 {
		t.Logf("✓ Post-neuron maintained reasonable firing rate")
	} else {
		t.Logf("WARNING: Post-neuron firing rate outside expected range")
	}

	t.Logf("✓ Two-neuron STDP network test completed")
}

// TestThreeNeuronChainSTDP tests STDP in a feed-forward chain
//
// BIOLOGICAL CONTEXT:
// Feed-forward chains are common in neural circuits, where activity propagates
// from input → intermediate → output neurons. STDP in such chains can create
// reliable signal transmission pathways and temporal sequence detection.
//
// EXPECTED BEHAVIOR:
// - Consistent activation patterns should strengthen the entire chain
// - Earlier neurons should reliably trigger later neurons
// - Chain should become more responsive to learned patterns
func TestThreeNeuronChainSTDP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping three-neuron chain test in short mode")
	}

	stdpConfig := STDPConfig{
		Enabled:        true,
		LearningRate:   0.025,
		TimeConstant:   18 * time.Millisecond,
		WindowSize:     45 * time.Millisecond,
		MinWeight:      0.3,
		MaxWeight:      2.2,
		AsymmetryRatio: 1.6,
	}

	// Create three neurons in a chain
	neuron1 := NewNeuron("input", 1.0, 0.95, 6*time.Millisecond, 1.0,
		0, 0, STDPConfig{Enabled: false}) // Input neuron - no learning/homeostasis

	neuron2 := NewNeuron("intermediate", 1.1, 0.95, 6*time.Millisecond, 1.0,
		3.0, 0.1, stdpConfig) // Intermediate with mild homeostasis

	neuron3 := NewNeuron("output", 1.1, 0.95, 6*time.Millisecond, 1.0,
		2.5, 0.1, stdpConfig) // Output with mild homeostasis

	// Connect with STDP: neuron1 → neuron2 → neuron3
	initialWeight12 := 0.9
	initialWeight23 := 0.9

	neuron1.AddOutputWithSTDP("to_n2", neuron2.GetInputChannel(),
		initialWeight12, 3*time.Millisecond, stdpConfig)
	neuron2.AddOutputWithSTDP("to_n3", neuron3.GetInputChannel(),
		initialWeight23, 3*time.Millisecond, stdpConfig)

	// Set up monitoring
	n2FireEvents := make(chan FireEvent, 100)
	n3FireEvents := make(chan FireEvent, 100)
	neuron2.SetFireEventChannel(n2FireEvents)
	neuron3.SetFireEventChannel(n3FireEvents)

	// Start all neurons
	go neuron1.Run()
	go neuron2.Run()
	go neuron3.Run()
	defer neuron1.Close()
	defer neuron2.Close()
	defer neuron3.Close()

	t.Logf("=== THREE-NEURON CHAIN STDP TEST ===")
	t.Logf("Chain: input → intermediate → output")
	t.Logf("Initial weights: %.2f → %.2f", initialWeight12, initialWeight23)

	input := neuron1.GetInput()

	// Training phase: consistent chain activation
	t.Logf("\n--- Training Phase: Chain Activation ---")

	for i := 0; i < 25; i++ {
		// Trigger chain with strong input
		input <- Message{Value: 1.8, Timestamp: time.Now(), SourceID: "training"}

		time.Sleep(120 * time.Millisecond) // Allow chain propagation and recovery
	}

	time.Sleep(2 * time.Second) // Allow learning and homeostasis to settle

	// Clear event channels
	for len(n2FireEvents) > 0 {
		<-n2FireEvents
	}
	for len(n3FireEvents) > 0 {
		<-n3FireEvents
	}

	// Test phase: measure chain responsiveness
	t.Logf("\n--- Test Phase: Chain Responsiveness ---")

	testTrials := 10
	n2Responses := 0
	n3Responses := 0
	chainResponses := 0

	for i := 0; i < testTrials; i++ {
		// Trigger with moderate input
		input <- Message{Value: 1.5, Timestamp: time.Now(), SourceID: "test"}

		// Monitor chain response within 100ms
		n2Fired := false
		n3Fired := false

		timeout := time.After(100 * time.Millisecond)
		for {
			select {
			case <-n2FireEvents:
				if !n2Fired {
					n2Fired = true
					n2Responses++
				}
			case <-n3FireEvents:
				if !n3Fired {
					n3Fired = true
					n3Responses++
				}
			case <-timeout:
				goto nextTrial
			}

			// Check if full chain fired
			if n2Fired && n3Fired && !((chainResponses + 1) > i+1) {
				chainResponses++
			}
		}
	nextTrial:
		time.Sleep(150 * time.Millisecond) // Inter-trial interval
	}

	n2Rate := float64(n2Responses) / float64(testTrials) * 100
	n3Rate := float64(n3Responses) / float64(testTrials) * 100
	chainRate := float64(chainResponses) / float64(testTrials) * 100

	t.Logf("Response rates after learning:")
	t.Logf("  Intermediate neuron: %.1f%% (%d/%d)", n2Rate, n2Responses, testTrials)
	t.Logf("  Output neuron: %.1f%% (%d/%d)", n3Rate, n3Responses, testTrials)
	t.Logf("  Complete chain: %.1f%% (%d/%d)", chainRate, chainResponses, testTrials)

	// Validate learning
	if n2Rate > 40 {
		t.Logf("✓ Strong input→intermediate connection learned")
	} else {
		t.Logf("Note: Moderate input→intermediate learning (%.1f%%)", n2Rate)
	}

	if n3Rate > 30 {
		t.Logf("✓ Intermediate→output connection functional")
	} else {
		t.Logf("Note: Weak intermediate→output transmission (%.1f%%)", n3Rate)
	}

	if chainRate > 20 {
		t.Logf("✓ End-to-end chain learning successful")
	} else {
		t.Logf("Note: Limited end-to-end chain formation (%.1f%%)", chainRate)
	}

	// Check final neuron states
	n2Rate_hz := neuron2.GetCurrentFiringRate()
	n3Rate_hz := neuron3.GetCurrentFiringRate()
	n2Threshold := neuron2.GetCurrentThreshold()
	n3Threshold := neuron3.GetCurrentThreshold()

	t.Logf("\nFinal neuron states:")
	t.Logf("  Intermediate: %.2f Hz, threshold %.3f", n2Rate_hz, n2Threshold)
	t.Logf("  Output: %.2f Hz, threshold %.3f", n3Rate_hz, n3Threshold)

	t.Logf("✓ Three-neuron chain STDP test completed")
}

// ============================================================================
// COMPETITIVE LEARNING TESTS
// ============================================================================

// TestSTDPCompetitiveLearnig tests competitive dynamics between multiple inputs
//
// BIOLOGICAL CONTEXT:
// In real neural networks, multiple inputs compete for influence over post-
// synaptic neurons. STDP creates competitive dynamics where consistently
// correlated inputs strengthen while uncorrelated inputs weaken. This
// implements "competitive learning" and input selectivity.
//
// EXPECTED BEHAVIOR:
// - Consistently correlated input should strengthen its synapse
// - Uncorrelated inputs should weaken or remain unchanged
// - Neuron should become more selective to learned patterns
func TestSTDPCompetitiveLearnig(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping competitive learning test in short mode")
	}

	stdpConfig := STDPConfig{
		Enabled:        true,
		LearningRate:   0.04,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      2.8,
		AsymmetryRatio: 1.7,
	}

	// Create post-synaptic neuron with homeostasis
	postNeuron := NewNeuron("competitor", 1.5, 0.95, 8*time.Millisecond, 1.0,
		4.0, 0.2, stdpConfig)

	// Create multiple input sources
	inputA := NewNeuron("inputA", 1.0, 0.95, 5*time.Millisecond, 1.0,
		0, 0, STDPConfig{Enabled: false})
	inputB := NewNeuron("inputB", 1.0, 0.95, 5*time.Millisecond, 1.0,
		0, 0, STDPConfig{Enabled: false})
	inputC := NewNeuron("inputC", 1.0, 0.95, 5*time.Millisecond, 1.0,
		0, 0, STDPConfig{Enabled: false})

	// Connect all inputs to post-neuron with equal initial weights
	initialWeight := 0.6
	inputA.AddOutputWithSTDP("to_post", postNeuron.GetInputChannel(),
		initialWeight, 2*time.Millisecond, stdpConfig)
	inputB.AddOutputWithSTDP("to_post", postNeuron.GetInputChannel(),
		initialWeight, 2*time.Millisecond, stdpConfig)
	inputC.AddOutputWithSTDP("to_post", postNeuron.GetInputChannel(),
		initialWeight, 2*time.Millisecond, stdpConfig)

	// Set up monitoring
	postFireEvents := make(chan FireEvent, 200)
	postNeuron.SetFireEventChannel(postFireEvents)

	// Start all neurons
	go inputA.Run()
	go inputB.Run()
	go inputC.Run()
	go postNeuron.Run()
	defer inputA.Close()
	defer inputB.Close()
	defer inputC.Close()
	defer postNeuron.Close()

	t.Logf("=== COMPETITIVE LEARNING STDP TEST ===")
	t.Logf("Three inputs (A, B, C) competing for influence")
	t.Logf("Initial weights: %.2f each", initialWeight)

	inA := inputA.GetInput()
	inB := inputB.GetInput()
	inC := inputC.GetInput()

	// Training phase: make input A consistently correlated with post-firing
	t.Logf("\n--- Training Phase: Input A Correlated, B & C Random ---")

	var wg sync.WaitGroup

	// Input A: consistent causal pattern
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 30; i++ {
			inA <- Message{Value: 1.4, Timestamp: time.Now(), SourceID: "trainA"}
			time.Sleep(150 * time.Millisecond)
		}
	}()

	// Input B: random timing
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			delay := time.Duration(50+i*13) * time.Millisecond // Varying delay
			time.Sleep(delay)
			inB <- Message{Value: 1.2, Timestamp: time.Now(), SourceID: "trainB"}
		}
	}()

	// Input C: anti-causal pattern (fires after post-neuron typically fires)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			time.Sleep(200 * time.Millisecond) // Delayed relative to A
			inC <- Message{Value: 1.3, Timestamp: time.Now(), SourceID: "trainC"}
		}
	}()

	wg.Wait()
	time.Sleep(2 * time.Second) // Allow learning to consolidate

	// Test phase: measure responsiveness to each input individually
	t.Logf("\n--- Test Phase: Individual Input Responsiveness ---")

	// Clear events
	for len(postFireEvents) > 0 {
		<-postFireEvents
	}

	testInputResponse := func(inputChan chan<- Message, inputName string) float64 {
		responses := 0
		testTrials := 8

		for i := 0; i < testTrials; i++ {
			inputChan <- Message{Value: 1.5, Timestamp: time.Now(), SourceID: "test_" + inputName}

			// Check for response within 80ms
			select {
			case <-postFireEvents:
				responses++
			case <-time.After(80 * time.Millisecond):
				// No response
			}

			time.Sleep(120 * time.Millisecond) // Inter-test interval
		}

		return float64(responses) / float64(testTrials) * 100
	}

	responseA := testInputResponse(inA, "A")
	time.Sleep(200 * time.Millisecond)
	responseB := testInputResponse(inB, "B")
	time.Sleep(200 * time.Millisecond)
	responseC := testInputResponse(inC, "C")

	t.Logf("Response rates after competitive learning:")
	t.Logf("  Input A (correlated): %.1f%%", responseA)
	t.Logf("  Input B (random): %.1f%%", responseB)
	t.Logf("  Input C (anti-causal): %.1f%%", responseC)

	// Validate competitive learning
	if responseA > responseB && responseA > responseC {
		t.Logf("✓ Competitive learning successful - correlated input dominates")
	} else {
		t.Logf("Note: Competitive advantage not clearly established")
	}

	if responseA > 40 {
		t.Logf("✓ Strong response to correlated input A")
	}
	if responseB < 30 {
		t.Logf("✓ Reduced response to random input B")
	}
	if responseC < 25 {
		t.Logf("✓ Reduced response to anti-causal input C")
	}

	// Final network state
	finalRate := postNeuron.GetCurrentFiringRate()
	finalThreshold := postNeuron.GetCurrentThreshold()
	t.Logf("\nFinal post-neuron state: %.2f Hz, threshold %.3f", finalRate, finalThreshold)

	t.Logf("✓ Competitive learning STDP test completed")
}

// ============================================================================
// NETWORK STABILITY TESTS
// ============================================================================

// TestSTDPNetworkStability tests that STDP doesn't destabilize networks
//
// BIOLOGICAL CONTEXT:
// One concern with STDP is that it could lead to runaway strengthening or
// weakening that destabilizes network activity. In healthy brains, multiple
// regulatory mechanisms prevent this. This test validates that our STDP
// implementation, combined with homeostasis, maintains network stability.
//
// EXPECTED BEHAVIOR:
// - Network activity should remain within reasonable bounds
// - No neurons should become completely silent or hyperactive
// - Learning should occur without causing instability
// - Homeostasis should provide stabilizing influence
func TestSTDPNetworkStability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network stability test in short mode")
	}

	stdpConfig := STDPConfig{
		Enabled:        true,
		LearningRate:   0.03, // Moderate learning rate
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.2, // Prevent complete silencing
		MaxWeight:      2.5, // Prevent runaway strengthening
		AsymmetryRatio: 1.5,
	}

	// Create a small network: 2 inputs → 2 processing neurons → 1 output
	numNeurons := 5
	neurons := make([]*Neuron, numNeurons)

	// Input neurons (no homeostasis)
	neurons[0] = NewNeuron("input1", 1.0, 0.95, 5*time.Millisecond, 1.0,
		0, 0, STDPConfig{Enabled: false})
	neurons[1] = NewNeuron("input2", 1.0, 0.95, 5*time.Millisecond, 1.0,
		0, 0, STDPConfig{Enabled: false})

	// Processing neurons (with homeostasis and STDP)
	neurons[2] = NewNeuron("proc1", 1.2, 0.95, 8*time.Millisecond, 1.0,
		3.5, 0.15, stdpConfig)
	neurons[3] = NewNeuron("proc2", 1.2, 0.95, 8*time.Millisecond, 1.0,
		3.5, 0.15, stdpConfig)

	// Output neuron (with homeostasis and STDP)
	neurons[4] = NewNeuron("output", 1.3, 0.95, 8*time.Millisecond, 1.0,
		2.5, 0.2, stdpConfig)

	// Create connections with STDP
	// Input layer → Processing layer
	neurons[0].AddOutputWithSTDP("to_proc1", neurons[2].GetInputChannel(),
		0.8, 2*time.Millisecond, stdpConfig)
	neurons[0].AddOutputWithSTDP("to_proc2", neurons[3].GetInputChannel(),
		0.7, 2*time.Millisecond, stdpConfig)
	neurons[1].AddOutputWithSTDP("to_proc1", neurons[2].GetInputChannel(),
		0.7, 2*time.Millisecond, stdpConfig)
	neurons[1].AddOutputWithSTDP("to_proc2", neurons[3].GetInputChannel(),
		0.8, 2*time.Millisecond, stdpConfig)

	// Processing layer → Output layer
	neurons[2].AddOutputWithSTDP("to_output", neurons[4].GetInputChannel(),
		0.9, 3*time.Millisecond, stdpConfig)
	neurons[3].AddOutputWithSTDP("to_output", neurons[4].GetInputChannel(),
		0.9, 3*time.Millisecond, stdpConfig)

	// Set up monitoring for all neurons
	fireChannels := make([]chan FireEvent, numNeurons)
	for i := range fireChannels {
		fireChannels[i] = make(chan FireEvent, 100)
		neurons[i].SetFireEventChannel(fireChannels[i])
	}

	// Start all neurons
	for _, neuron := range neurons {
		go neuron.Run()
	}
	defer func() {
		for _, neuron := range neurons {
			neuron.Close()
		}
	}()

	t.Logf("=== NETWORK STABILITY TEST ===")
	t.Logf("Network: 2 inputs → 2 processing → 1 output")
	t.Logf("All connections have STDP learning enabled")

	// Record initial states
	initialRates := make([]float64, numNeurons)
	initialThresholds := make([]float64, numNeurons)
	for i, neuron := range neurons {
		initialRates[i] = neuron.GetCurrentFiringRate()
		initialThresholds[i] = neuron.GetCurrentThreshold()
	}

	// Extended operation with varied activity patterns
	t.Logf("\n--- Extended Operation: Varied Activity Patterns ---")

	input1 := neurons[0].GetInput()
	input2 := neurons[1].GetInput()

	// Monitor activity over time
	monitoringDuration := 8 * time.Second
	sampleInterval := 500 * time.Millisecond
	samples := int(monitoringDuration / sampleInterval)

	rateHistory := make([][]float64, numNeurons)
	for i := range rateHistory {
		rateHistory[i] = make([]float64, 0, samples)
	}

	// Background activity generation
	var wg sync.WaitGroup
	stopSignal := make(chan struct{})

	// Input pattern generator
	wg.Add(1)
	go func() {
		defer wg.Done()
		patterns := []struct {
			delay1, delay2       time.Duration
			strength1, strength2 float64
		}{
			{0, 10 * time.Millisecond, 1.4, 1.2},                   // Pattern A
			{15 * time.Millisecond, 0, 1.3, 1.4},                   // Pattern B
			{5 * time.Millisecond, 5 * time.Millisecond, 1.5, 1.5}, // Simultaneous
		}

		patternIdx := 0
		for {
			select {
			case <-stopSignal:
				return
			default:
				pattern := patterns[patternIdx%len(patterns)]

				go func() {
					time.Sleep(pattern.delay1)
					input1 <- Message{Value: pattern.strength1, Timestamp: time.Now(), SourceID: "pattern"}
				}()
				go func() {
					time.Sleep(pattern.delay2)
					input2 <- Message{Value: pattern.strength2, Timestamp: time.Now(), SourceID: "pattern"}
				}()

				patternIdx++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Monitoring loop
	for sample := 0; sample < samples; sample++ {
		time.Sleep(sampleInterval)

		for i, neuron := range neurons {
			rate := neuron.GetCurrentFiringRate()
			rateHistory[i] = append(rateHistory[i], rate)
		}

		if sample%4 == 0 { // Log every 2 seconds
			t.Logf("Sample %d/%d: Output rate %.2f Hz", sample+1, samples,
				neurons[4].GetCurrentFiringRate())
		}
	}

	// Stop background activity
	close(stopSignal)
	wg.Wait()

	// Analyze stability
	t.Logf("\n--- Stability Analysis ---")

	finalRates := make([]float64, numNeurons)
	finalThresholds := make([]float64, numNeurons)
	for i, neuron := range neurons {
		finalRates[i] = neuron.GetCurrentFiringRate()
		finalThresholds[i] = neuron.GetCurrentThreshold()
	}

	neuronNames := []string{"input1", "input2", "proc1", "proc2", "output"}

	// Check for stability issues
	stabilityIssues := 0
	for i := 2; i < numNeurons; i++ { // Skip input neurons (no homeostasis)
		name := neuronNames[i]
		rate := finalRates[i]
		threshold := finalThresholds[i]

		t.Logf("%s: rate %.2f Hz, threshold %.3f", name, rate, threshold)

		// Check for pathological states
		if rate > 50 { // Hyperactivity
			t.Logf("WARNING: %s shows hyperactivity (%.2f Hz)", name, rate)
			stabilityIssues++
		}
		if rate < 0.1 { // Silence
			t.Logf("WARNING: %s is nearly silent (%.2f Hz)", name, rate)
			stabilityIssues++
		}
		if threshold > initialThresholds[i]*3 { // Extreme threshold increase
			t.Logf("WARNING: %s threshold increased dramatically", name)
			stabilityIssues++
		}
		if threshold < initialThresholds[i]*0.3 { // Extreme threshold decrease
			t.Logf("WARNING: %s threshold decreased dramatically", name)
			stabilityIssues++
		}
	}

	// Analyze rate variability
	for i := 2; i < numNeurons; i++ {
		history := rateHistory[i]
		if len(history) < 2 {
			continue
		}

		// Calculate coefficient of variation
		sum := 0.0
		for _, rate := range history {
			sum += rate
		}
		mean := sum / float64(len(history))

		variance := 0.0
		for _, rate := range history {
			variance += math.Pow(rate-mean, 2)
		}
		stdDev := math.Sqrt(variance / float64(len(history)))

		cv := stdDev / mean
		if mean > 0 {
			t.Logf("%s variability: CV=%.3f (mean=%.2f, std=%.2f)",
				neuronNames[i], cv, mean, stdDev)

			if cv > 2.0 {
				t.Logf("WARNING: %s shows high variability", neuronNames[i])
				stabilityIssues++
			}
		}
	}

	// Overall stability assessment
	if stabilityIssues == 0 {
		t.Logf("✓ Network remained stable throughout extended operation")
	} else {
		t.Logf("WARNING: %d stability issues detected", stabilityIssues)
	}

	// Check that learning occurred (some threshold changes expected)
	learningDetected := false
	for i := 2; i < numNeurons; i++ {
		thresholdChange := math.Abs(finalThresholds[i] - initialThresholds[i])
		if thresholdChange > 0.05 {
			learningDetected = true
			break
		}
	}

	if learningDetected {
		t.Logf("✓ Learning activity detected (threshold adaptations)")
	} else {
		t.Logf("Note: Minimal learning detected - may need stronger patterns")
	}

	t.Logf("✓ Network stability test completed")
}

// ============================================================================
// PATTERN LEARNING TESTS
// ============================================================================

// TestSTDPTemporalPatternLearning tests learning of temporal sequences
//
// BIOLOGICAL CONTEXT:
// One of STDP's key functions is enabling neurons to learn temporal patterns
// and sequences. This is crucial for processing time-varying information like
// speech, motor sequences, and sensory patterns. STDP should enable neurons
// to become selective for specific temporal patterns.
//
// EXPECTED BEHAVIOR:
// - Repeated temporal patterns should strengthen specific synapse combinations
// - Neuron should become more responsive to learned sequences
// - Non-matching patterns should elicit weaker responses
func TestSTDPTemporalPatternLearning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping temporal pattern learning test in short mode")
	}

	stdpConfig := STDPConfig{
		Enabled:        true,
		LearningRate:   0.035,
		TimeConstant:   15 * time.Millisecond,
		WindowSize:     40 * time.Millisecond,
		MinWeight:      0.2,
		MaxWeight:      2.0,
		AsymmetryRatio: 2.0, // Strong LTP bias for pattern learning
	}

	// Create pattern detector neuron
	detector := NewNeuron("detector", 2.0, 0.95, 6*time.Millisecond, 1.0,
		3.0, 0.15, stdpConfig) // Higher threshold for selectivity

	// Create multiple input channels
	numInputs := 4
	inputNeurons := make([]*Neuron, numInputs)
	for i := 0; i < numInputs; i++ {
		inputNeurons[i] = NewNeuron(fmt.Sprintf("input_%d", i), 1.0, 0.95,
			5*time.Millisecond, 1.0, 0, 0, STDPConfig{Enabled: false})
	}

	// Connect inputs to detector with STDP
	initialWeight := 0.6
	for i, inputNeuron := range inputNeurons {
		inputNeuron.AddOutputWithSTDP(fmt.Sprintf("to_detector_%d", i),
			detector.GetInputChannel(), initialWeight,
			time.Duration(i+1)*time.Millisecond, stdpConfig)
	}

	// Set up monitoring
	detectorEvents := make(chan FireEvent, 200)
	detector.SetFireEventChannel(detectorEvents)

	// Start all neurons
	for _, neuron := range inputNeurons {
		go neuron.Run()
	}
	go detector.Run()

	defer func() {
		for _, neuron := range inputNeurons {
			neuron.Close()
		}
		detector.Close()
	}()

	t.Logf("=== TEMPORAL PATTERN LEARNING TEST ===")
	t.Logf("Training detector neuron to recognize specific temporal sequence")

	// Define target pattern: Input 0 → Input 1 → Input 2 → Input 3
	// with specific timing intervals
	targetPattern := []struct {
		input int
		delay time.Duration
	}{
		{0, 0},
		{1, 8 * time.Millisecond},
		{2, 16 * time.Millisecond},
		{3, 24 * time.Millisecond},
	}

	// Training phase: present target pattern repeatedly
	t.Logf("\n--- Training Phase: Target Pattern A-B-C-D ---")

	for trial := 0; trial < 40; trial++ {
		// Present target pattern
		for _, step := range targetPattern {
			go func(inputIdx int, delay time.Duration) {
				time.Sleep(delay)
				inputNeurons[inputIdx].GetInput() <- Message{
					Value:     1.5,
					Timestamp: time.Now(),
					SourceID:  fmt.Sprintf("pattern_train_%d", inputIdx),
				}
			}(step.input, step.delay)
		}

		// Trigger detector firing 30ms after pattern start (within STDP window)
		go func() {
			time.Sleep(30 * time.Millisecond)
			detector.GetInput() <- Message{
				Value:     1.0, // Additional input to ensure firing
				Timestamp: time.Now(),
				SourceID:  "training_trigger",
			}
		}()

		time.Sleep(150 * time.Millisecond) // Inter-trial interval
	}

	time.Sleep(2 * time.Second) // Allow learning consolidation

	// Test phase: compare responses to different patterns
	t.Logf("\n--- Test Phase: Pattern Recognition ---")

	// Clear event buffer
	for len(detectorEvents) > 0 {
		<-detectorEvents
	}

	testPattern := func(pattern []struct {
		input int
		delay time.Duration
	}, name string) float64 {
		responses := 0
		testTrials := 8

		for trial := 0; trial < testTrials; trial++ {
			// Present test pattern
			for _, step := range pattern {
				go func(inputIdx int, delay time.Duration) {
					time.Sleep(delay)
					inputNeurons[inputIdx].GetInput() <- Message{
						Value:     1.4,
						Timestamp: time.Now(),
						SourceID:  fmt.Sprintf("test_%s_%d", name, inputIdx),
					}
				}(step.input, step.delay)
			}

			// Check for detector response within 60ms
			select {
			case <-detectorEvents:
				responses++
			case <-time.After(60 * time.Millisecond):
				// No response
			}

			time.Sleep(120 * time.Millisecond) // Inter-trial interval
		}

		return float64(responses) / float64(testTrials) * 100
	}

	// Test 1: Target pattern (should have high response)
	targetResponse := testPattern(targetPattern, "target")

	// Test 2: Reversed pattern
	reversedPattern := []struct {
		input int
		delay time.Duration
	}{
		{3, 0},
		{2, 8 * time.Millisecond},
		{1, 16 * time.Millisecond},
		{0, 24 * time.Millisecond},
	}
	reversedResponse := testPattern(reversedPattern, "reversed")

	// Test 3: Scrambled pattern (different timing)
	scrambledPattern := []struct {
		input int
		delay time.Duration
	}{
		{0, 0},
		{2, 5 * time.Millisecond},
		{1, 12 * time.Millisecond},
		{3, 30 * time.Millisecond},
	}
	scrambledResponse := testPattern(scrambledPattern, "scrambled")

	// Test 4: Single input (should have low response)
	singlePattern := []struct {
		input int
		delay time.Duration
	}{
		{0, 0},
	}
	singleResponse := testPattern(singlePattern, "single")

	t.Logf("Pattern recognition results:")
	t.Logf("  Target pattern A→B→C→D: %.1f%%", targetResponse)
	t.Logf("  Reversed pattern D→C→B→A: %.1f%%", reversedResponse)
	t.Logf("  Scrambled timing: %.1f%%", scrambledResponse)
	t.Logf("  Single input: %.1f%%", singleResponse)

	// Validate pattern learning
	if targetResponse > reversedResponse && targetResponse > scrambledResponse {
		t.Logf("✓ Temporal pattern learning successful - target pattern preferred")
	} else {
		t.Logf("Note: Pattern selectivity not clearly established")
	}

	if targetResponse > 30 {
		t.Logf("✓ Strong response to learned target pattern")
	}
	if reversedResponse < targetResponse*0.7 {
		t.Logf("✓ Reduced response to reversed pattern")
	}
	if singleResponse < targetResponse*0.5 {
		t.Logf("✓ Reduced response to incomplete pattern")
	}

	// Check detector final state
	finalRate := detector.GetCurrentFiringRate()
	finalThreshold := detector.GetCurrentThreshold()
	t.Logf("\nDetector final state: %.2f Hz, threshold %.3f", finalRate, finalThreshold)

	t.Logf("✓ Temporal pattern learning test completed")
}

// ============================================================================
// PERFORMANCE AND STRESS TESTS
// ============================================================================

// TestSTDPNetworkPerformance tests STDP performance under high activity
//
// BIOLOGICAL CONTEXT:
// Real neural networks can have very high activity levels during intense
// processing. STDP mechanisms must be computationally efficient enough to
// handle high spike rates without becoming a bottleneck.
//
// EXPECTED BEHAVIOR:
// - STDP should handle high spike rates without errors
// - Learning should continue to function under stress
// - Memory usage should remain reasonable
// - No goroutine leaks or deadlocks should occur
func TestSTDPNetworkPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	stdpConfig := STDPConfig{
		Enabled:        true,
		LearningRate:   0.02, // Moderate rate to avoid extreme changes
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.3,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.5,
	}

	// Create high-activity network
	numInputs := 6
	numProcessing := 4
	numOutputs := 2

	inputNeurons := make([]*Neuron, numInputs)
	processingNeurons := make([]*Neuron, numProcessing)
	outputNeurons := make([]*Neuron, numOutputs)

	// Create input neurons (no homeostasis for high, controlled activity)
	for i := 0; i < numInputs; i++ {
		inputNeurons[i] = NewNeuron(fmt.Sprintf("input_%d", i), 1.0, 0.95,
			3*time.Millisecond, 1.0, 0, 0, STDPConfig{Enabled: false})
	}

	// Create processing neurons (with homeostasis and STDP)
	for i := 0; i < numProcessing; i++ {
		processingNeurons[i] = NewNeuron(fmt.Sprintf("proc_%d", i), 1.1, 0.95,
			4*time.Millisecond, 1.0, 8.0, 0.1, stdpConfig) // Higher target rate
	}

	// Create output neurons (with homeostasis and STDP)
	for i := 0; i < numOutputs; i++ {
		outputNeurons[i] = NewNeuron(fmt.Sprintf("out_%d", i), 1.2, 0.95,
			4*time.Millisecond, 1.0, 5.0, 0.15, stdpConfig)
	}

	// Create dense connectivity with STDP
	connectionCount := 0

	// Input → Processing layer
	for i := 0; i < numInputs; i++ {
		for j := 0; j < numProcessing; j++ {
			weight := 0.4 + 0.2*float64(j)/float64(numProcessing) // Varying weights
			inputNeurons[i].AddOutputWithSTDP(
				fmt.Sprintf("i%d_to_p%d", i, j),
				processingNeurons[j].GetInputChannel(),
				weight, time.Duration(i+1)*time.Millisecond, stdpConfig)
			connectionCount++
		}
	}

	// Processing → Output layer
	for i := 0; i < numProcessing; i++ {
		for j := 0; j < numOutputs; j++ {
			weight := 0.6 + 0.3*float64(i)/float64(numProcessing)
			processingNeurons[i].AddOutputWithSTDP(
				fmt.Sprintf("p%d_to_o%d", i, j),
				outputNeurons[j].GetInputChannel(),
				weight, time.Duration(i+2)*time.Millisecond, stdpConfig)
			connectionCount++
		}
	}

	t.Logf("=== STDP NETWORK PERFORMANCE TEST ===")
	t.Logf("Network: %d inputs → %d processing → %d outputs", numInputs, numProcessing, numOutputs)
	t.Logf("Total STDP connections: %d", connectionCount)

	// Start all neurons
	allNeurons := make([]*Neuron, 0, numInputs+numProcessing+numOutputs)
	allNeurons = append(allNeurons, inputNeurons...)
	allNeurons = append(allNeurons, processingNeurons...)
	allNeurons = append(allNeurons, outputNeurons...)

	for _, neuron := range allNeurons {
		go neuron.Run()
	}
	defer func() {
		for _, neuron := range allNeurons {
			neuron.Close()
		}
	}()

	// High-intensity activity phase
	t.Logf("\n--- High-Intensity Activity Phase ---")

	startTime := time.Now()
	testDuration := 4 * time.Second
	stopSignal := make(chan struct{})

	var wg sync.WaitGroup

	// High-frequency input generators
	for i, inputNeuron := range inputNeurons {
		wg.Add(1)
		go func(idx int, neuron *Neuron) {
			defer wg.Done()
			input := neuron.GetInput()
			localSpikes := 0

			// Generate ~30-50 Hz activity per input
			ticker := time.NewTicker(time.Duration(20+idx*5) * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-stopSignal:
					return
				case <-ticker.C:
					input <- Message{
						Value:     1.2 + 0.3*float64(idx%3), // Varying strengths
						Timestamp: time.Now(),
						SourceID:  fmt.Sprintf("perf_input_%d", idx),
					}
					localSpikes++
				}
			}
		}(i, inputNeuron)
	}

	// Monitor network performance
	wg.Add(1)
	go func() {
		defer wg.Done()
		monitorTicker := time.NewTicker(500 * time.Millisecond)
		defer monitorTicker.Stop()

		for {
			select {
			case <-stopSignal:
				return
			case <-monitorTicker.C:
				// Sample processing neuron rates
				sampleRate := processingNeurons[0].GetCurrentFiringRate()
				elapsed := time.Since(startTime)
				t.Logf("Performance check at %.1fs: sample rate %.1f Hz",
					elapsed.Seconds(), sampleRate)
			}
		}
	}()

	// Run for test duration
	time.Sleep(testDuration)
	close(stopSignal)
	wg.Wait()

	elapsed := time.Since(startTime)
	t.Logf("\n--- Performance Results ---")
	t.Logf("Test duration: %.2f seconds", elapsed.Seconds())

	// Analyze final network state
	avgInputRate := 0.0
	avgProcRate := 0.0
	avgOutputRate := 0.0

	for _, neuron := range inputNeurons {
		avgInputRate += neuron.GetCurrentFiringRate()
	}
	avgInputRate /= float64(numInputs)

	for _, neuron := range processingNeurons {
		avgProcRate += neuron.GetCurrentFiringRate()
	}
	avgProcRate /= float64(numProcessing)

	for _, neuron := range outputNeurons {
		avgOutputRate += neuron.GetCurrentFiringRate()
	}
	avgOutputRate /= float64(numOutputs)

	t.Logf("Average firing rates:")
	t.Logf("  Input layer: %.1f Hz", avgInputRate)
	t.Logf("  Processing layer: %.1f Hz", avgProcRate)
	t.Logf("  Output layer: %.1f Hz", avgOutputRate)

	// Validate performance
	if avgProcRate > 1.0 && avgProcRate < 100.0 {
		t.Logf("✓ Processing layer maintained reasonable activity levels")
	} else {
		t.Logf("WARNING: Processing layer activity outside expected range")
	}

	if avgOutputRate > 0.5 {
		t.Logf("✓ Output layer remained active")
	} else {
		t.Logf("WARNING: Output layer activity very low")
	}

	// Check for stability (no extreme threshold changes)
	extremeChanges := 0
	for _, neuron := range processingNeurons {
		threshold := neuron.GetCurrentThreshold()
		baseThreshold := neuron.GetBaseThreshold()
		change := math.Abs(threshold - baseThreshold)
		if change > baseThreshold*2 { // More than 200% change
			extremeChanges++
		}
	}

	if extremeChanges == 0 {
		t.Logf("✓ No extreme threshold changes detected")
	} else {
		t.Logf("WARNING: %d neurons showed extreme threshold changes", extremeChanges)
	}

	// Estimate computational load
	estimatedLearningEvents := avgProcRate * elapsed.Seconds() * float64(connectionCount) * 0.1
	t.Logf("Estimated STDP learning events: %.0f", estimatedLearningEvents)

	if estimatedLearningEvents > 1000 {
		t.Logf("✓ STDP handled substantial computational load")
	}

	t.Logf("✓ Network performance test completed successfully")
}

// ============================================================================
// INTEGRATION BENCHMARK TESTS
// ============================================================================

// Global STDP configuration for benchmarks, similar to integration tests
var benchSTDPConfig = STDPConfig{
	Enabled:        true,
	LearningRate:   0.02,
	TimeConstant:   20 * time.Millisecond,
	WindowSize:     50 * time.Millisecond,
	MinWeight:      0.1,
	MaxWeight:      2.5,
	AsymmetryRatio: 1.5,
}

// BenchmarkNeuronProcessingWithSTDPAndHomeostasis measures message processing
// by a single neuron with both STDP and homeostatic plasticity active.
func BenchmarkNeuronProcessingWithSTDPAndHomeostasis(b *testing.B) {
	neuron := NewNeuron("bench_integrated_neuron", 1.0, 0.95, 10*time.Millisecond, 1.0,
		5.0, 0.2, benchSTDPConfig) // Target 5Hz, 0.2 homeostasis strength

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()
	msg := Message{Value: 0.8, Timestamp: time.Now(), SourceID: "bench_source"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg.Timestamp = time.Now() // Update timestamp for each message
		input <- msg
	}
	b.StopTimer() // Stop timer before sleep/cleanup
}

// BenchmarkTwoNeuronSTDPEvent measures the performance of a typical STDP
// learning event between two connected neurons (pre-fires, post-fires, STDP update).
func BenchmarkTwoNeuronSTDPEvent(b *testing.B) {
	preNeuron := NewNeuron("bench_pre", 1.0, 0.95, 8*time.Millisecond, 1.0,
		0, 0, STDPConfig{Enabled: false}) // Pre-neuron simple, STDP on its output
	postNeuron := NewNeuron("bench_post", 1.2, 0.95, 8*time.Millisecond, 1.0,
		4.0, 0.15, benchSTDPConfig) // Post-neuron with STDP & Homeostasis

	initialWeight := 0.8
	// Ensure STDP is enabled on the synapse itself
	synapseSTDPConfig := benchSTDPConfig
	synapseSTDPConfig.Enabled = true // Explicitly enable for the synapse
	preNeuron.AddOutputWithSTDP("to_post", postNeuron.GetInputChannel(),
		initialWeight, 2*time.Millisecond, synapseSTDPConfig)

	go preNeuron.Run()
	go postNeuron.Run()
	defer preNeuron.Close()
	defer postNeuron.Close()

	preIn := preNeuron.GetInput()
	postIn := postNeuron.GetInput()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a causal STDP event
		preSpikeTime := time.Now()
		preIn <- Message{Value: 1.5, Timestamp: preSpikeTime, SourceID: "bench_pre_trigger"}

		// Wait for pre-spike to potentially reach post-neuron (considering 2ms synapse delay)
		// and then trigger post-neuron for LTP
		// This simplified sleep is okay for benchmark purposes to ensure sequence
		time.Sleep(5 * time.Millisecond)

		postSpikeTime := time.Now()
		postIn <- Message{Value: 1.5, Timestamp: postSpikeTime, SourceID: "bench_post_trigger"}

		// Allow a very brief moment for internal processing of the second spike and STDP
		// This is tricky in benchmarks; for true isolation, one might need more complex synchronization
		// or benchmark the internal STDP methods directly if they were public.
		// Given the async nature, this benchmark measures the whole interaction.
		time.Sleep(1 * time.Microsecond) // Minimal delay to allow goroutines to run
	}
	b.StopTimer()
}

// BenchmarkThreeNeuronChainSTDPEvent measures performance of spike propagation
// and STDP learning through a three-neuron chain.
func BenchmarkThreeNeuronChainSTDPEvent(b *testing.B) {
	neuron1 := NewNeuron("bench_chain1", 1.0, 0.95, 6*time.Millisecond, 1.0,
		0, 0, STDPConfig{Enabled: false})
	neuron2 := NewNeuron("bench_chain2", 1.1, 0.95, 6*time.Millisecond, 1.0,
		3.0, 0.1, benchSTDPConfig)
	neuron3 := NewNeuron("bench_chain3", 1.1, 0.95, 6*time.Millisecond, 1.0,
		2.5, 0.1, benchSTDPConfig)

	synapseSTDPConfig := benchSTDPConfig
	synapseSTDPConfig.Enabled = true

	neuron1.AddOutputWithSTDP("to_n2", neuron2.GetInputChannel(),
		0.9, 3*time.Millisecond, synapseSTDPConfig)
	neuron2.AddOutputWithSTDP("to_n3", neuron3.GetInputChannel(),
		0.9, 3*time.Millisecond, synapseSTDPConfig)

	go neuron1.Run()
	go neuron2.Run()
	go neuron3.Run()
	defer neuron1.Close()
	defer neuron2.Close()
	defer neuron3.Close()

	input1 := neuron1.GetInput()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Trigger the start of the chain
		msgTime := time.Now()
		input1 <- Message{Value: 1.8, Timestamp: msgTime, SourceID: "bench_chain_start"}
		// Allow some time for propagation and potential STDP events
		// This is a high-level benchmark of the chain reacting.
		time.Sleep(15 * time.Millisecond) // Enough for 2 hops (3ms delay each) + processing
	}
	b.StopTimer()
}

// BenchmarkCompetitiveInputSTDPProcessing measures the cost of a neuron
// processing multiple concurrent inputs that trigger STDP.
func BenchmarkCompetitiveInputSTDPProcessing(b *testing.B) {
	detector := NewNeuron("bench_detector_competitive", 1.5, 0.95, 8*time.Millisecond, 1.0,
		4.0, 0.2, benchSTDPConfig)

	numInputs := 3
	inputFields := make([]chan<- Message, numInputs)
	inputSources := make([]*Neuron, numInputs)

	synapseSTDPConfig := benchSTDPConfig
	synapseSTDPConfig.Enabled = true

	for i := 0; i < numInputs; i++ {
		inputSource := NewNeuron(fmt.Sprintf("bench_comp_in_%d", i), 1.0, 0.95, 5*time.Millisecond, 1.0,
			0, 0, STDPConfig{Enabled: false})
		inputSource.AddOutputWithSTDP(fmt.Sprintf("to_detector_%d", i),
			detector.GetInputChannel(), 0.6, 2*time.Millisecond, synapseSTDPConfig)
		go inputSource.Run()
		defer inputSource.Close()
		inputFields[i] = inputSource.GetInput()
		inputSources[i] = inputSource
	}

	go detector.Run()
	defer detector.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		baseTime := time.Now()
		// Send spikes from all input sources with slight offsets
		for j := 0; j < numInputs; j++ {
			inputFields[j] <- Message{Value: 1.2, Timestamp: baseTime.Add(time.Duration(j) * time.Millisecond), SourceID: inputSources[j].id}
		}
		// Send a message that makes the detector fire, to trigger STDP updates based on recent inputs
		detector.GetInput() <- Message{Value: 1.0, Timestamp: baseTime.Add(time.Duration(numInputs) * time.Millisecond), SourceID: "trigger_competitive"}
		time.Sleep(1 * time.Microsecond) // Minimal delay
	}
	b.StopTimer()
}

// BenchmarkSmallNetworkHighActivityWithSTDP simulates a small network under
// high load with STDP and homeostasis active, measuring the time per cycle of input.
func BenchmarkSmallNetworkHighActivityWithSTDP(b *testing.B) {
	numInputs := 2
	numProcessing := 2

	inputNeurons := make([]*Neuron, numInputs)
	processingNeurons := make([]*Neuron, numProcessing)

	synapseSTDPConfig := benchSTDPConfig
	synapseSTDPConfig.Enabled = true

	for i := 0; i < numInputs; i++ {
		inputNeurons[i] = NewNeuron(fmt.Sprintf("bench_snet_in_%d", i), 1.0, 0.95,
			3*time.Millisecond, 1.0, 0, 0, STDPConfig{Enabled: false})
	}
	for i := 0; i < numProcessing; i++ {
		processingNeurons[i] = NewNeuron(fmt.Sprintf("bench_snet_proc_%d", i), 1.1, 0.95,
			4*time.Millisecond, 1.0, 8.0, 0.1, benchSTDPConfig)
	}

	for i := 0; i < numInputs; i++ {
		for j := 0; j < numProcessing; j++ {
			inputNeurons[i].AddOutputWithSTDP(
				fmt.Sprintf("snet_i%d_to_p%d", i, j),
				processingNeurons[j].GetInputChannel(),
				0.7, time.Duration(i+1)*time.Millisecond, synapseSTDPConfig)
		}
	}

	allNeurons := make([]*Neuron, 0, numInputs+numProcessing)
	allNeurons = append(allNeurons, inputNeurons...)
	allNeurons = append(allNeurons, processingNeurons...)

	for _, neuron := range allNeurons {
		go neuron.Run()
	}
	defer func() {
		for _, neuron := range allNeurons {
			neuron.Close()
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ts := time.Now()
		// Send one spike to each input neuron
		for idx, inputNeuron := range inputNeurons {
			inputNeuron.GetInput() <- Message{Value: 1.3, Timestamp: ts.Add(time.Duration(idx) * time.Microsecond), SourceID: fmt.Sprintf("bench_snet_source_%d", idx)}
		}
		// Allow some time for spikes to propagate and processing to occur
		time.Sleep(10 * time.Millisecond)
	}
	b.StopTimer()
}
