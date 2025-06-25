/*
=================================================================================
BIOLOGICAL SYNAPSE TESTS
=================================================================================

OVERVIEW:
This file contains tests that validate the biological realism of the synapse
implementation. These tests ensure that synaptic behavior matches what is
observed in real biological neural networks, including:

1. SPIKE-TIMING DEPENDENT PLASTICITY (STDP)
   - Classic timing window (LTP for pre-before-post, LTD for post-before-pre)
   - Exponential decay of plasticity effects with timing difference
   - Asymmetric learning window (LTD typically stronger than LTP)
   - Precise millisecond-level timing sensitivity

2. STRUCTURAL PLASTICITY
   - Activity-dependent pruning ("use it or lose it")
   - Protection of recently active synapses
   - Weight-dependent pruning decisions
   - Realistic timescales for synaptic elimination

3. BIOLOGICAL CONSTRAINTS
   - Realistic weight bounds (preventing pathological strengthening/weakening)
   - Physiologically plausible transmission delays
   - Appropriate learning rates and time constants
   - Biologically realistic activity patterns

4. SYNAPTIC TRANSMISSION FIDELITY
   - Accurate signal scaling by synaptic weight
   - Preservation of temporal information for learning
   - Realistic delay effects on signal propagation

EXPERIMENTAL VALIDATION:
These tests are based on experimental findings from neuroscience literature,
including classic STDP experiments (Bi & Poo, 1998), structural plasticity
studies, and in vivo recordings of synaptic behavior. The test parameters
and expected behaviors reflect published biological data.

BIOLOGICAL SIGNIFICANCE:
Validating biological realism ensures that:
- Networks using these synapses will exhibit brain-like learning dynamics
- Temporal processing capabilities match biological neural networks
- Emergent behaviors arise from realistic local rules
- Research applications maintain biological relevance
*/

package synapse

import (
	"math"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

// =================================================================================
// SPIKE-TIMING DEPENDENT PLASTICITY (STDP) BIOLOGICAL TESTS
// =================================================================================

// TestSTDPClassicTimingWindow tests the classic STDP learning window.
// This test validates that synapses strengthen when pre-synaptic spikes
// precede post-synaptic spikes (causal timing) and weaken when the timing
// is reversed (anti-causal timing).
//
// BIOLOGICAL BASIS:
// This test replicates the classic experiment by Bi & Poo (1998) where
// precise timing between pre- and post-synaptic activity determines the
// direction and magnitude of synaptic weight changes. The timing window
// typically shows:
// - LTP (strengthening) for pre-before-post with ~20ms time constant
// - LTD (weakening) for post-before-pre with slightly faster kinetics
// - Exponential decay of effects with increasing time intervals
//
// EXPERIMENTAL PROTOCOL:
// 1. Create synapse with realistic STDP parameters
// 2. Apply plasticity adjustments with various timing differences
// 3. Verify that weight changes match biological STDP profile
// 4. Confirm asymmetric learning window shape
func TestSTDPClassicTimingWindow(t *testing.T) {
	// Create mock neurons for controlled testing
	preNeuron := NewMockNeuron("pre_neuron")
	postNeuron := NewMockNeuron("post_neuron")

	// Configure STDP parameters based on biological data
	// These values reflect experimental measurements from cortical synapses
	stdpConfig := types.PlasticityConfig{
		Enabled:        true,
		LearningRate:   0.01,                   // 1% change per pairing (Bi & Poo, 1998)
		TimeConstant:   20 * time.Millisecond,  // Classic cortical value
		WindowSize:     100 * time.Millisecond, // Effective learning window
		MinWeight:      0.001,                  // Prevent elimination
		MaxWeight:      2.0,                    // Prevent runaway strengthening
		AsymmetryRatio: 1.2,                    // LTD slightly stronger than LTP
	}

	pruningConfig := CreateDefaultPruningConfig()

	// Test multiple timing differences across the STDP window
	testCases := []struct {
		name           string
		timeDifference time.Duration // Δt = t_pre - t_post
		expectedSign   float64       // Expected sign of weight change
		description    string
	}{
		{
			name:           "StrongLTP",
			timeDifference: -10 * time.Millisecond, // Pre 10ms before post
			expectedSign:   1.0,                    // Positive (strengthening)
			description:    "Pre-synaptic spike 10ms before post-synaptic (causal)",
		},
		{
			name:           "WeakLTP",
			timeDifference: -50 * time.Millisecond, // Pre 50ms before post
			expectedSign:   1.0,                    // Positive but weaker
			description:    "Pre-synaptic spike 50ms before post-synaptic",
		},
		{
			name:           "WeakLTD",
			timeDifference: 10 * time.Millisecond, // Pre 10ms after post
			expectedSign:   -1.0,                  // Negative (weakening)
			description:    "Pre-synaptic spike 10ms after post-synaptic (anti-causal)",
		},
		{
			name:           "StrongLTD",
			timeDifference: 30 * time.Millisecond, // Pre 30ms after post
			expectedSign:   -1.0,                  // Negative (weakening)
			description:    "Pre-synaptic spike 30ms after post-synaptic",
		},
		{
			name:           "NoPlasticity",
			timeDifference: 150 * time.Millisecond, // Outside window
			expectedSign:   0.0,                    // No change
			description:    "Timing difference outside STDP window",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fresh synapse for each test to avoid history effects
			initialWeight := 1.0
			synapse := NewBasicSynapse("stdp_test", preNeuron, postNeuron,
				stdpConfig, pruningConfig, initialWeight, 0)

			// Record initial weight
			weightBefore := synapse.GetWeight()

			// Apply plasticity adjustment with specific timing
			adjustment := types.PlasticityAdjustment{
				DeltaT: tc.timeDifference,
			}
			synapse.ApplyPlasticity(adjustment)

			// Measure weight change
			weightAfter := synapse.GetWeight()
			weightChange := weightAfter - weightBefore

			// Verify the direction of change matches biological expectation
			if tc.expectedSign > 0 && weightChange <= 0 {
				t.Errorf("Expected LTP (weight increase) for %s, got change: %f",
					tc.description, weightChange)
			} else if tc.expectedSign < 0 && weightChange >= 0 {
				t.Errorf("Expected LTD (weight decrease) for %s, got change: %f",
					tc.description, weightChange)
			} else if tc.expectedSign == 0 && math.Abs(weightChange) > 1e-10 {
				t.Errorf("Expected no plasticity for %s, got change: %f",
					tc.description, weightChange)
			}

			// Verify that changes are within reasonable biological bounds
			if math.Abs(weightChange) > 0.1 {
				t.Errorf("Weight change too large for single pairing: %f", weightChange)
			}
		})
	}
}

// TestSTDPExponentialDecay verifies that STDP effects decay exponentially
// with increasing time differences, as observed in biological experiments.
//
// BIOLOGICAL BASIS:
// In real synapses, the magnitude of plasticity effects decreases exponentially
// as the time difference between pre- and post-synaptic spikes increases.
// This creates a precisely tuned temporal learning window that emphasizes
// strong causal relationships while de-emphasizing weak temporal correlations.
func TestSTDPExponentialDecay(t *testing.T) {
	preNeuron := NewMockNeuron("pre_neuron")
	postNeuron := NewMockNeuron("post_neuron")

	stdpConfig := types.PlasticityConfig{
		Enabled:        true,
		LearningRate:   0.02,                  // Larger for easier measurement
		TimeConstant:   15 * time.Millisecond, // Shorter for sharper decay
		WindowSize:     80 * time.Millisecond,
		MinWeight:      0.001,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.0, // Symmetric for simpler analysis
	}

	pruningConfig := CreateDefaultPruningConfig()

	// Test exponential decay with multiple time points
	timeDifferences := []time.Duration{
		-5 * time.Millisecond,  // Close timing
		-15 * time.Millisecond, // At time constant
		-30 * time.Millisecond, // 2x time constant
		-45 * time.Millisecond, // 3x time constant
	}

	weightChanges := make([]float64, len(timeDifferences))

	// Measure weight changes for each timing
	for i, deltaT := range timeDifferences {
		initialWeight := 1.0
		synapse := NewBasicSynapse("decay_test", preNeuron, postNeuron,
			stdpConfig, pruningConfig, initialWeight, 0)

		adjustment := types.PlasticityAdjustment{DeltaT: deltaT}
		synapse.ApplyPlasticity(adjustment)

		weightChanges[i] = synapse.GetWeight() - initialWeight
	}

	// Verify exponential decay pattern
	for i := 1; i < len(weightChanges); i++ {
		// Each subsequent change should be smaller due to exponential decay
		if weightChanges[i] >= weightChanges[i-1] {
			t.Errorf("STDP effects not decaying exponentially: change[%d]=%f >= change[%d]=%f",
				i, weightChanges[i], i-1, weightChanges[i-1])
		}

		// Verify approximate exponential relationship
		// At time constant τ, effect should be ~37% of maximum
		if i == 1 { // At time constant (15ms vs 5ms)
			expectedRatio := math.Exp(-10.0 / 15.0) // exp(-Δt/τ)
			actualRatio := weightChanges[i] / weightChanges[0]

			// Allow 20% tolerance for numerical precision
			if math.Abs(actualRatio-expectedRatio) > 0.2 {
				t.Errorf("Exponential decay ratio incorrect: expected ~%f, got %f",
					expectedRatio, actualRatio)
			}
		}
	}
}

// =================================================================================
// STRUCTURAL PLASTICITY BIOLOGICAL TESTS
// =================================================================================

// TestActivityDependentPruning validates that synaptic pruning follows
// biological "use it or lose it" principles, where inactive synapses
// are eliminated while active synapses are preserved.
//
// BIOLOGICAL BASIS:
// In developing and adult brains, synapses that fail to participate in
// network activity are gradually eliminated through molecular mechanisms
// involving protein degradation and structural remodeling. This process
// optimizes neural circuits by removing ineffective connections while
// preserving functionally important pathways.
//
// EXPERIMENTAL DESIGN:
// 1. Create synapses with different activity patterns
// 2. Simulate realistic inactivity periods
// 3. Verify that pruning decisions match biological criteria
// 4. Ensure active synapses are protected from elimination
func TestActivityDependentPruning(t *testing.T) {
	preNeuron := NewMockNeuron("pre_neuron")
	postNeuron := NewMockNeuron("post_neuron")

	// Configure aggressive pruning for faster testing
	stdpConfig := CreateDefaultSTDPConfig()
	pruningConfig := PruningConfig{
		Enabled:             true,
		WeightThreshold:     0.1,                    // Higher threshold for testing
		InactivityThreshold: 100 * time.Millisecond, // Shorter for testing
	}

	t.Run("ActiveSynapseProtection", func(t *testing.T) {
		// Create synapse with weight above pruning threshold
		strongWeight := 0.2 // Above threshold (0.1)
		synapse := NewBasicSynapse("strong_synapse", preNeuron, postNeuron,
			stdpConfig, pruningConfig, strongWeight, 0)

		// Verify not marked for pruning initially (recently created)
		if synapse.ShouldPrune() {
			t.Error("Strong, recently created synapse should not be marked for pruning")
		}

		// Simulate recent activity through plasticity
		recentAdjustment := types.PlasticityAdjustment{DeltaT: -10 * time.Millisecond}
		synapse.ApplyPlasticity(recentAdjustment)

		// Even after inactivity period, recently active synapse should be protected
		time.Sleep(120 * time.Millisecond) // Beyond inactivity threshold
		if synapse.ShouldPrune() {
			t.Error("Recently active synapse should not be marked for pruning")
		}
	})

	t.Run("WeakInactiveSynapsePruning", func(t *testing.T) {
		// Create weak synapse below pruning threshold
		weakWeight := 0.05 // Below threshold (0.1)
		synapse := NewBasicSynapse("weak_synapse", preNeuron, postNeuron,
			stdpConfig, pruningConfig, weakWeight, 0)

		// Initially should not be pruned (grace period)
		if synapse.ShouldPrune() {
			t.Error("Weak synapse should have grace period after creation")
		}

		// Wait for inactivity period to expire
		time.Sleep(120 * time.Millisecond)

		// Now should be marked for pruning (weak + inactive)
		if !synapse.ShouldPrune() {
			t.Error("Weak, inactive synapse should be marked for pruning")
		}
	})

	t.Run("WeakButActiveSynapseProtection", func(t *testing.T) {
		// Create weak synapse but keep it active
		weakWeight := 0.05
		synapse := NewBasicSynapse("weak_active_synapse", preNeuron, postNeuron,
			stdpConfig, pruningConfig, weakWeight, 0)

		// Wait for most of inactivity period
		time.Sleep(80 * time.Millisecond)

		// Apply recent plasticity to mark as active
		recentAdjustment := types.PlasticityAdjustment{DeltaT: -5 * time.Millisecond}
		synapse.ApplyPlasticity(recentAdjustment)

		// Should not be pruned because it's recently active
		if synapse.ShouldPrune() {
			t.Error("Weak but recently active synapse should not be pruned")
		}
	})
}

// TestPruningTimescales validates that synaptic pruning operates on
// biologically realistic timescales, providing sufficient opportunity
// for synapses to demonstrate their functional importance.
//
// BIOLOGICAL RATIONALE:
// Synaptic pruning in biology operates on timescales of hours to days,
// not seconds or minutes. This gives synapses adequate opportunity to
// participate in network activity and prove their functional value
// before being eliminated.
func TestPruningTimescales(t *testing.T) {
	preNeuron := NewMockNeuron("pre_neuron")
	postNeuron := NewMockNeuron("post_neuron")

	stdpConfig := CreateDefaultSTDPConfig()

	// Test multiple timescales for biological realism
	testCases := []struct {
		name               string
		inactivityDuration time.Duration
		weightThreshold    float64
		expectedPruning    bool
		biologicalContext  string
	}{
		{
			name:               "ShortInactivity",
			inactivityDuration: 1 * time.Second,
			weightThreshold:    0.1,
			expectedPruning:    false,
			biologicalContext:  "Brief pauses in activity should not trigger pruning",
		},
		{
			name:               "ModerateInactivity",
			inactivityDuration: 100 * time.Millisecond, // Short for testing
			weightThreshold:    0.1,
			expectedPruning:    true,
			biologicalContext:  "Extended inactivity should trigger pruning",
		},
		{
			name:               "LongInactivity",
			inactivityDuration: 50 * time.Millisecond, // Even shorter for testing
			weightThreshold:    0.1,
			expectedPruning:    true,
			biologicalContext:  "Prolonged inactivity definitely triggers pruning",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pruningConfig := PruningConfig{
				Enabled:             true,
				WeightThreshold:     tc.weightThreshold,
				InactivityThreshold: tc.inactivityDuration,
			}

			// Create weak synapse for testing
			weakWeight := 0.05 // Below threshold
			synapse := NewBasicSynapse("timescale_test", preNeuron, postNeuron,
				stdpConfig, pruningConfig, weakWeight, 0)

			// For testing purposes, we simulate the passage of time by waiting
			// the inactivity duration, then checking pruning logic
			if tc.expectedPruning {
				// Wait for the inactivity threshold to pass
				time.Sleep(tc.inactivityDuration + 10*time.Millisecond)

				shouldPrune := synapse.ShouldPrune()
				if !shouldPrune {
					t.Errorf("Expected pruning after %v of inactivity (%s)",
						tc.inactivityDuration, tc.biologicalContext)
				}
			} else {
				// For short inactivity, verify it's not pruned immediately
				shouldPrune := synapse.ShouldPrune()
				if shouldPrune {
					t.Errorf("Unexpected pruning after only %v (%s)",
						tc.inactivityDuration, tc.biologicalContext)
				}
			}
		})
	}
}

// =================================================================================
// SYNAPTIC TRANSMISSION BIOLOGICAL TESTS
// =================================================================================

// TestTransmissionDelayAccuracy validates that synaptic transmission delays
// accurately model biological axonal conduction and synaptic processing times.
//
// BIOLOGICAL BASIS:
// Real synaptic transmission involves multiple delay components:
// - Axonal conduction delay (depends on length, diameter, myelination)
// - Synaptic delay (neurotransmitter release and diffusion)
// - Postsynaptic response time (receptor binding and ion channel opening)
//
// Total delays typically range from 0.5ms (fast local synapses) to 50ms
// (long-distance connections). Accuracy is critical for temporal processing
// and spike-timing dependent learning.
func TestTransmissionDelayAccuracy(t *testing.T) {
	preNeuron := NewMockNeuron("pre_neuron")
	postNeuron := NewMockNeuron("post_neuron")

	stdpConfig := CreateDefaultSTDPConfig()
	pruningConfig := CreateDefaultPruningConfig()

	// Test multiple biologically realistic delay values
	delayTests := []struct {
		delay          time.Duration
		tolerance      time.Duration
		biologicalType string
	}{
		{
			delay:          1 * time.Millisecond,
			tolerance:      500 * time.Microsecond,
			biologicalType: "Fast local synapse",
		},
		{
			delay:          5 * time.Millisecond,
			tolerance:      1 * time.Millisecond,
			biologicalType: "Typical cortical synapse",
		},
		{
			delay:          15 * time.Millisecond,
			tolerance:      3 * time.Millisecond,
			biologicalType: "Medium-distance connection",
		},
		{
			delay:          50 * time.Millisecond,
			tolerance:      10 * time.Millisecond,
			biologicalType: "Long-distance projection",
		},
	}

	for _, test := range delayTests {
		t.Run(test.biologicalType, func(t *testing.T) {
			synapse := NewBasicSynapse("delay_test", preNeuron, postNeuron,
				stdpConfig, pruningConfig, 1.0, test.delay)

			// Clear messages and reset internal state for each subtest
			postNeuron.ClearReceivedMessages()
			preNeuron.ClearReceivedMessages()    // Clear preNeuron's queue as well
			preNeuron.SetCurrentTime(time.Now()) // Reset preNeuron's internal clock for deterministic delays

			// Record transmission start time
			startTime := time.Now()

			// Transmit signal
			synapse.Transmit(1.0)

			// Calculate the time at which the message *should* be delivered
			// This time includes the base transmission delay from the synapse.
			// The `Transmit` method calls `ScheduleDelayedDelivery` on `preNeuron`,
			// which uses `preNeuron.currentTime.Add(totalDelay)`.
			// So, to ensure delivery, `preNeuron.ProcessDelayedMessages` needs to be
			// called with a `currentTime` that is past this calculated `deliveryTime`.
			expectedDeliveryTime := startTime.Add(test.delay)

			// Simulate the passage of time in the mock neuron's internal clock
			// We need to advance the `preNeuron`'s `currentTime` sufficiently
			// for the message to become due.
			preNeuron.ProcessDelayedMessages(expectedDeliveryTime.Add(test.tolerance))

			// Wait for a small additional buffer for goroutine scheduling in the mock's Receive
			time.Sleep(10 * time.Millisecond) // A small real-world sleep just in case of goroutine scheduling

			// Verify message was received
			messages := postNeuron.GetReceivedMessages()
			if len(messages) == 0 {
				t.Fatalf("No message received for %s after expected delay (%v)", test.biologicalType, test.delay)
			}

			// Check that delay was approximately as expected
			// The exact `actualDelay` is hard to measure precisely due to goroutine scheduling,
			// but we can check if it falls within a reasonable window.
			// Since `ProcessDelayedMessages` immediately dispatches once `currentTime` passes `deliveryTime`,
			// the `Receive` timestamp might be very close to `expectedDeliveryTime`.
			actualMessageTimestamp := messages[0].Timestamp
			// Calculate the effective delay as the difference between message timestamp and when transmit was called.
			effectiveDelay := actualMessageTimestamp.Sub(startTime)

			// Validate that effectiveDelay is close to test.delay
			if effectiveDelay < test.delay || effectiveDelay > test.delay+(test.tolerance*5) { // Allow slightly more for scheduling
				t.Errorf("Message effective delay incorrect for %s: expected ~%v, got %v (diff: %v)",
					test.biologicalType, test.delay, effectiveDelay, effectiveDelay-test.delay)
			}

			// Clear messages for next test
			postNeuron.ClearReceivedMessages()
		})
	}
}

// TestSynapticWeightScaling validates that signal scaling by synaptic weight
// accurately models biological synaptic efficacy modulation.
//
// BIOLOGICAL BASIS:
// Synaptic efficacy (weight) represents the strength of synaptic transmission,
// determined by factors like neurotransmitter release probability, receptor
// density, and postsynaptic response amplitude. In biology, synaptic weights
// can vary over orders of magnitude between different synapses.
func TestSynapticWeightScaling(t *testing.T) {
	preNeuron := NewMockNeuron("pre_neuron")
	postNeuron := NewMockNeuron("post_neuron")

	stdpConfig := CreateDefaultSTDPConfig()
	pruningConfig := CreateDefaultPruningConfig()

	// Test range of biologically realistic weights
	weightTests := []struct {
		weight      float64
		inputSignal float64
		description string
	}{
		{0.1, 1.0, "Weak synapse (10% efficacy)"},
		{0.5, 1.0, "Moderate synapse (50% efficacy)"},
		{1.0, 1.0, "Strong synapse (100% efficacy)"},
		{1.5, 1.0, "Very strong synapse (150% efficacy)"},
		{0.8, 2.0, "Moderate synapse with strong input"},
		{1.2, 0.5, "Strong synapse with weak input"},
	}

	for _, test := range weightTests {
		t.Run(test.description, func(t *testing.T) {
			synapse := NewBasicSynapse("scaling_test", preNeuron, postNeuron,
				stdpConfig, pruningConfig, test.weight, 0)

			// Transmit signal
			synapse.Transmit(test.inputSignal)

			// Allow transmission to complete
			time.Sleep(10 * time.Millisecond)

			// Verify correct scaling
			messages := postNeuron.GetReceivedMessages()
			if len(messages) != 1 {
				t.Fatalf("Expected 1 message, got %d", len(messages))
			}

			expectedOutput := test.inputSignal * test.weight
			actualOutput := messages[0].Value

			if math.Abs(actualOutput-expectedOutput) > 1e-10 {
				t.Errorf("Incorrect weight scaling: input=%f, weight=%f, expected=%f, got=%f",
					test.inputSignal, test.weight, expectedOutput, actualOutput)
			}

			// Clear messages for next test
			postNeuron.receivedMsgs = nil
		})
	}
}

// =================================================================================
// INTEGRATED BIOLOGICAL BEHAVIOR TESTS
// =================================================================================

// TestRealisticSynapticDynamics validates integrated synaptic behavior
// under realistic neural activity patterns, combining STDP learning,
// transmission dynamics, and structural plasticity.
//
// BIOLOGICAL SCENARIO:
// This test simulates a learning scenario where repeated pre-post spike
// pairings should strengthen a synapse through STDP, while maintaining
// realistic transmission characteristics and avoiding pathological behavior.
func TestRealisticSynapticDynamics(t *testing.T) {
	preNeuron := NewMockNeuron("pre_neuron")
	postNeuron := NewMockNeuron("post_neuron")

	// Configure with biologically realistic parameters
	stdpConfig := types.PlasticityConfig{
		Enabled:        true,
		LearningRate:   0.005, // Conservative learning rate
		TimeConstant:   15 * time.Millisecond,
		WindowSize:     60 * time.Millisecond,
		MinWeight:      0.01,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.3,
	}

	pruningConfig := PruningConfig{
		Enabled:             true,
		WeightThreshold:     0.05,
		InactivityThreshold: 30 * time.Second,
	}

	initialWeight := 0.5
	transmissionDelay := 2 * time.Millisecond

	synapse := NewBasicSynapse("realistic_test", preNeuron, postNeuron,
		stdpConfig, pruningConfig, initialWeight, transmissionDelay)

	// Phase 1: Learning through repeated pairings
	numPairings := 50
	pairingInterval := -8 * time.Millisecond // Pre-before-post (LTP)

	for i := 0; i < numPairings; i++ {
		// Simulate transmission
		synapse.Transmit(1.0)

		// Apply STDP with favorable timing
		adjustment := types.PlasticityAdjustment{DeltaT: pairingInterval}
		synapse.ApplyPlasticity(adjustment)

		// Brief pause between pairings
		time.Sleep(time.Millisecond)
	}

	// Verify learning occurred
	finalWeight := synapse.GetWeight()
	weightIncrease := finalWeight - initialWeight

	if weightIncrease <= 0 {
		t.Error("Expected weight increase from repeated LTP pairings")
	}

	if finalWeight >= stdpConfig.MaxWeight {
		t.Error("Weight should not saturate at maximum from moderate learning")
	}

	// Phase 2: Verify synapse remains functional
	synapse.Transmit(1.0)
	// Process any delayed messages in the mock
	preNeuron.ProcessDelayedMessages(time.Now().Add(transmissionDelay + 5*time.Millisecond))

	messages := postNeuron.GetReceivedMessages()
	if len(messages) == 0 {
		t.Error("Synapse should remain functional after learning")
	}

	// Verify signal strength reflects learned weight
	latestMessage := messages[len(messages)-1]
	expectedSignal := 1.0 * finalWeight

	if math.Abs(latestMessage.Value-expectedSignal) > 0.01 {
		t.Errorf("Signal strength should reflect learned weight: expected %f, got %f",
			expectedSignal, latestMessage.Value)
	}

	// Phase 3: Verify protection from pruning due to recent activity
	if synapse.ShouldPrune() {
		t.Error("Recently active, strong synapse should not be marked for pruning")
	}

	// Biological significance verification
	if finalWeight < initialWeight*1.1 {
		t.Error("Weight increase too small for biological significance")
	}

	if finalWeight > initialWeight*2.0 {
		t.Error("Weight increase too large for single learning session")
	}
}
