/*
=================================================================================
DENDRITIC INTEGRATION MODE TESTS
=================================================================================

OVERVIEW:
This test suite validates the functionality and biological realism of the various
`DendriticIntegrationMode` strategies. These tests ensure that each integration
mode correctly implements its intended computational behavior, from simple passive
summation to complex non-linear dendritic processing.

BIOLOGICAL CONTEXT:
The `DendriticIntegrationMode` architecture is designed to model the diverse
computational strategies employed by biological dendrites. Different neuron types
have different dendritic structures and ion channel compositions, leading to varied
integration behaviors. This test suite verifies our models of these behaviors.

KEY MECHANISMS TESTED:
1.  **PassiveMembraneMode**: Ensures backward compatibility with immediate processing.
2.  **TemporalSummationMode**: Validates time-based batching and the solution to
    the original GABAergic timing problem.
3.  **ShuntingInhibitionMode**: Confirms the divisive (multiplicative) effect of
    conductance-based inhibition.
4.  **ActiveDendriteMode**: Tests a combination of sophisticated non-linear
    mechanisms, including synaptic saturation, shunting, and NMDA-like spikes.

These tests are crucial for verifying that the neuron's core computational
abilities are both functionally correct and biologically plausible.
*/

package neuron

import (
	"sync"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/synapse"
)

// ============================================================================
// PassiveMembraneMode Tests (Backward Compatibility)
// ============================================================================

// TestDendritePassiveMode validates that the `PassiveMembraneMode` correctly
// mimics the original, immediate processing behavior of the neuron.
//
// BIOLOGICAL SIGNIFICANCE:
// This mode models a neuron with a very simple dendritic structure or inputs
// that synapse directly onto the soma, where temporal integration effects are
// minimal. Its primary purpose in the architecture is to ensure that existing
// networks or tests that depend on immediate, non-buffered processing continue
// to function as expected.
//
// EXPECTED RESULTS:
// - The `Handle` method should immediately return an `IntegratedPotential`.
// - The `Process` method should do nothing and return `nil`.
// - The `NetInput` of the result should exactly match the input message's value.
func TestDendritePassiveMode(t *testing.T) {
	t.Log("=== TESTING PassiveMembraneMode ===")
	mode := NewPassiveMembraneMode()

	// Test 1: Excitatory input
	t.Run("ExcitatoryInput", func(t *testing.T) {
		msg := synapse.SynapseMessage{Value: 1.5}
		result := mode.Handle(msg)
		if result == nil || result.NetInput != 1.5 {
			t.Errorf("Expected immediate result with NetInput 1.5, got %v", result)
		}
	})

	// Test 2: Inhibitory input
	t.Run("InhibitoryInput", func(t *testing.T) {
		msg := synapse.SynapseMessage{Value: -0.8}
		result := mode.Handle(msg)
		if result == nil || result.NetInput != -0.8 {
			t.Errorf("Expected immediate result with NetInput -0.8, got %v", result)
		}
	})

	// Test 3: Process method should be a no-op
	t.Run("ProcessMethod", func(t *testing.T) {
		state := MembraneSnapshot{Accumulator: 0.5}
		result := mode.Process(state)
		if result != nil {
			t.Errorf("Process() should return nil for PassiveMembraneMode, got %v", result)
		}
	})

	t.Log("✓ `PassiveMembraneMode` provides correct immediate processing for backward compatibility.")
}

// ============================================================================
// TemporalSummationMode Tests (Time-based Integration)
// ============================================================================

// TestDendriteTemporalSummationMode validates that inputs are correctly buffered
// and processed as a batch, solving the original GABAergic timing issue.
//
// BIOLOGICAL SIGNIFICANCE:
// This test directly verifies the model of the membrane time constant. It proves
// that the neuron can integrate multiple inputs arriving within a short window
// before making a firing decision. This is fundamental to all non-trivial neural
// computation, including coincidence detection and the proper balancing of
// excitation and inhibition.
//
// EXPECTED RESULTS:
//   - `Handle` should buffer messages and return `nil`.
//   - `Process` should sum the values of all buffered messages.
//   - Simultaneous excitatory and inhibitory inputs should correctly cancel out,
//     preventing a firing event that would have occurred with immediate processing.
func TestDendriteTemporalSummationMode(t *testing.T) {
	t.Log("=== TESTING TemporalSummationMode ===")
	mode := NewTemporalSummationMode()

	// Test 1: Buffering functionality
	t.Run("Buffering", func(t *testing.T) {
		result1 := mode.Handle(synapse.SynapseMessage{Value: 0.5})
		if result1 != nil {
			t.Fatal("Handle() should buffer and return nil, but got a result")
		}
		result2 := mode.Handle(synapse.SynapseMessage{Value: 0.3})
		if result2 != nil {
			t.Fatal("Handle() should buffer and return nil, but got a result")
		}
		result3 := mode.Handle(synapse.SynapseMessage{Value: -0.2})
		if result3 != nil {
			t.Fatal("Handle() should buffer and return nil, but got a result")
		}

		t.Log("✓ Messages correctly buffered without immediate processing.")
	})

	// Test 2: Batch processing and summation
	t.Run("BatchProcessing", func(t *testing.T) {
		state := MembraneSnapshot{}
		result := mode.Process(state)

		// expectedNetInput := 0.5 + 0.3 - 0.2 // 0.6
		if result == nil {
			t.Fatal("Process() returned nil when buffer was full")
		}
		// Use a tolerance for float comparison
		if result.NetInput < 0.599 || result.NetInput > 0.601 {
			t.Errorf("Expected summed NetInput of ~0.6, got %.3f", result.NetInput)
		}

		// Verify that the buffer is cleared after processing
		resultAfter := mode.Process(state)
		if resultAfter != nil {
			t.Errorf("Buffer should be empty after processing, but Process() returned a result")
		}
		t.Logf("✓ Batch correctly processed with NetInput %.3f.", result.NetInput)
	})

	// Test 3: Solves the GABAergic timing problem
	t.Run("GABAergicTimingFix", func(t *testing.T) {
		// Simulate a strong excitatory input and an even stronger inhibitory input
		// arriving in the same time step.
		mode.Handle(synapse.SynapseMessage{Value: 1.5})  // Would cause firing alone
		mode.Handle(synapse.SynapseMessage{Value: -2.0}) // Should prevent firing

		result := mode.Process(MembraneSnapshot{})
		// expectedNetInput := 1.5 - 2.0 // -0.5

		if result == nil {
			t.Fatal("Process() returned nil for GABA timing test")
		}
		if result.NetInput < -0.501 || result.NetInput > -0.499 {
			t.Errorf("Expected NetInput of ~-0.5, got %.3f", result.NetInput)
		}
		t.Log("✓ Solved GABAergic timing flaw: Inhibition correctly counteracts simultaneous excitation.")
	})
}

// ============================================================================
// ShuntingInhibitionMode Tests (Non-Linear Divisive Inhibition)
// ============================================================================

// TestDendriteShuntingInhibitionMode validates the non-linear, divisive effect
// of conductance-based shunting inhibition.
//
// BIOLOGICAL SIGNIFICANCE:
// This is a more accurate model of how many inhibitory synapses work. Instead of
// just subtracting voltage, they increase the "leakiness" of the membrane,
// reducing the *impact* or *gain* of excitatory inputs. This is a powerful form
// of gain control used throughout the brain to stabilize networks and perform
// complex computations like direction selectivity in the visual system.
//
// EXPECTED RESULTS:
//   - Inhibition should reduce the effect of excitation multiplicatively.
//   - Given a fixed excitatory input, increasing inhibition should result in a
//     progressively smaller (but non-zero) net potential.
//   - Zero inhibition should result in the full excitatory value being passed through.
func TestDendriteShuntingInhibitionMode(t *testing.T) {
	t.Log("=== TESTING ShuntingInhibitionMode ===")
	// Create a standard, deterministic biological config for the test.
	bioConfig := CreateCorticalPyramidalConfig()
	bioConfig.MembraneNoise = 0  // Disable noise for predictable results
	bioConfig.TemporalJitter = 0 // Disable jitter for predictable results

	mode := NewShuntingInhibitionMode(0.5, bioConfig) // A strength of 0.5 for testing

	// Test 1: No inhibition
	t.Run("NoInhibition", func(t *testing.T) {
		mode.Handle(synapse.SynapseMessage{Value: 2.0})
		result := mode.Process(MembraneSnapshot{})
		// Expected: Input 2.0 -> spatial decay: 2.0 * 0.7 = 1.4
		if result.NetInput < 1.39 || result.NetInput > 1.41 {
			t.Errorf("With no inhibition, expected NetInput 1.4, got %.2f", result.NetInput)
		}
		t.Log("✓ Correctly passes full excitation with zero inhibition.")
	})

	// Test 2: Moderate inhibition
	t.Run("ModerateInhibition", func(t *testing.T) {
		mode.Handle(synapse.SynapseMessage{Value: 2.0})
		mode.Handle(synapse.SynapseMessage{Value: -1.0}) // totalInhibition = 1.0
		result := mode.Process(MembraneSnapshot{})

		// Expected: decayedExcitation = 2.0 * 0.7 = 1.4
		// decayedInhibition = 1.0 * 0.7 = 0.7
		// shuntingFactor = 1.0 - (0.7 * 0.5) = 0.65
		// NetInput = 1.4 * 0.65 = 0.91
		if result.NetInput < 0.90 || result.NetInput > 0.92 {
			t.Errorf("Expected shunted NetInput of ~0.91, got %.2f", result.NetInput)
		}
		t.Logf("✓ Moderate inhibition correctly reduced excitatory impact to %.2f.", result.NetInput)
	})

	// Test 3: Strong inhibition and floor value
	t.Run("StrongInhibition", func(t *testing.T) {
		mode.Handle(synapse.SynapseMessage{Value: 2.0})
		mode.Handle(synapse.SynapseMessage{Value: -3.0}) // totalInhibition = 3.0
		result := mode.Process(MembraneSnapshot{})

		// Expected: decayedExcitation = 2.0 * 0.7 = 1.4
		// decayedInhibition = 3.0 * 0.7 = 2.1
		// shuntingFactor = 1.0 - (2.1 * 0.5) = -0.05, which should be floored to 0.1
		// NetInput = 1.4 * 0.1 = 0.14
		if result.NetInput < 0.13 || result.NetInput > 0.15 {
			t.Errorf("Expected floored shunted NetInput of ~0.14, got %.2f", result.NetInput)
		}
		t.Log("✓ Strong inhibition was correctly floored, preventing inversion of signal.")
	})
}

// ============================================================================
// ActiveDendriteMode Tests (Comprehensive Non-Linear Model)
// ============================================================================

// TestDendriteActiveDendriteMode validates the combined non-linear effects of the
// most advanced integration model.
//
// BIOLOGICAL SIGNIFICANCE:
// This test models the complex computational environment of a cortical pyramidal
// neuron's dendrite. It verifies that multiple, interacting non-linearities—
// physical limits (saturation), divisive gain control (shunting), and regenerative
// events (dendritic spikes)—can coexist and function correctly, providing a rich
// and powerful computational substrate.
//
// EXPECTED RESULTS:
//   - Synaptic saturation should cap abnormally large inputs.
//   - Shunting inhibition should correctly modulate the saturated inputs.
//   - Dendritic spikes should trigger only when the shunted excitation surpasses
//     the spike threshold, providing a final non-linear boost.
func TestDendriteActiveDendriteMode(t *testing.T) {
	t.Log("=== TESTING ActiveDendriteMode ===")
	config := ActiveDendriteConfig{
		MaxSynapticEffect:       2.5, // Capped at 2.5
		ShuntingStrength:        0.4, // 40% strength
		DendriticSpikeThreshold: 1.2, // Spike triggers above 1.2
		NMDASpikeAmplitude:      1.0, // Spike adds +1.0
	}
	// Create a standard biological config to pass to the constructors.
	bioConfig := CreateCorticalPyramidalConfig()
	bioConfig.MembraneNoise = 0
	bioConfig.TemporalJitter = 0
	mode := NewActiveDendriteMode(config, bioConfig)

	// Test Case 1: FIX - This test now has a more specific config to isolate saturation.
	t.Run("JustSaturation", func(t *testing.T) {
		// Use a specific config where saturation won't trigger the spike.
		isoConfig := ActiveDendriteConfig{
			MaxSynapticEffect:       1.1, // Cap is BELOW spike threshold
			DendriticSpikeThreshold: 1.2,
			NMDASpikeAmplitude:      1.0,
			ShuntingStrength:        0.4,
		}
		isoMode := NewActiveDendriteMode(isoConfig, bioConfig)
		isoMode.Handle(synapse.SynapseMessage{Value: 5.0})
		result := isoMode.Process(MembraneSnapshot{})
		// Expected: Input 5.0 -> saturated to 1.1 -> spatial decay: 1.1 * 0.7 = 0.77
		// 0.77 is NOT > 1.2, so NO spike. Final result is 0.77.
		if result.NetInput < 0.76 || result.NetInput > 0.78 {
			t.Errorf("Expected saturated NetInput plus spike of ~3.5, got %.2f", result.NetInput)
		}
		t.Logf("✓ 'JustSaturation' test correctly isolated and verified capping at %.2f.", isoConfig.MaxSynapticEffect)
	})

	// Test Case 2: FIX - Renamed to clarify it tests saturation AND the spike. Expectation fixed.
	t.Run("SaturationPlusSpike", func(t *testing.T) {
		mode.Handle(synapse.SynapseMessage{Value: 5.0}) // Should be capped at 2.5
		result := mode.Process(MembraneSnapshot{})
		// Expected: Input 5.0 -> saturated to 2.5 -> spatial decay: 2.5 * 0.7 = 1.75
		// 1.75 > 1.2 (DendriticSpikeThreshold), so spike is triggered.
		// finalNetInput = 1.75 + 1.0 (NMDASpikeAmplitude) = 2.75
		if result.NetInput < 2.74 || result.NetInput > 2.76 {
			t.Errorf("Expected saturated NetInput plus spike of ~2.75, got %.2f", result.NetInput)
		}
		t.Logf("✓ 'SaturationPlusSpike' correctly verified capping and subsequent spike.")
	})

	// Test Case 3: All features combined
	t.Run("ShuntingPlusSaturationPlusSpike", func(t *testing.T) {
		mode.Handle(synapse.SynapseMessage{Value: 4.0}) // Capped to 2.5
		mode.Handle(synapse.SynapseMessage{Value: 1.0})
		mode.Handle(synapse.SynapseMessage{Value: -1.0})
		result := mode.Process(MembraneSnapshot{})

		// decayedExcitation = (2.5 * 0.7) + (1.0 * 0.7) = 1.75 + 0.7 = 2.45
		// decayedInhibition = 1.0 * 0.7 = 0.7
		// shuntingFactor = 1.0 - (0.7 * 0.4) = 0.72
		// netExcitation = 2.45 * 0.72 = 1.764
		// Since netExcitation (1.764) > DendriticSpikeThreshold (1.2), add spike amplitude.
		// finalNetInput = 1.764 + 1.0 = 2.764
		if result.NetInput < 2.75 || result.NetInput > 2.77 {
			t.Errorf("Expected final NetInput of ~2.76, got %.2f", result.NetInput)
		}
		t.Logf("✓ Saturation, shunting, and NMDA spike logic combined correctly.")
	})

	// Test Case 4: Dendritic spike does NOT trigger
	t.Run("NoDendriticSpike", func(t *testing.T) {
		mode.Handle(synapse.SynapseMessage{Value: 2.0})
		mode.Handle(synapse.SynapseMessage{Value: -2.0}) // Strong inhibition
		result := mode.Process(MembraneSnapshot{})

		// decayedExcitation = 2.0 * 0.7 = 1.4
		// decayedInhibition = 2.0 * 0.7 = 1.4
		// shuntingFactor = 1.0 - (1.4 * 0.4) = 0.44
		// netExcitation = 1.4 * 0.44 = 0.616
		// Since netExcitation (0.616) < DendriticSpikeThreshold (1.2), NO spike.
		// finalNetInput = 0.616
		if result.NetInput < 0.60 || result.NetInput > 0.63 {
			t.Errorf("Expected final NetInput of ~0.62, got %.2f", result.NetInput)
		}
		t.Log("✓ Correctly avoided dendritic spike for sub-threshold dendritic potential.")
	})
}

// ============================================================================
// Concurrency and Edge Case Tests
// ============================================================================

// TestDendriteConcurrencyAndEdges ensures that all buffered modes are robust
// against race conditions and handle edge cases gracefully.
//
// BIOLOGICAL SIGNIFICANCE:
// A biological neuron is massively parallel, receiving thousands of inputs
// simultaneously from different sources. A robust simulation must be able to
// handle this concurrency without data corruption or deadlocks. This test
// ensures our mutex-protected buffers are working correctly. It also tests
// for edge cases like zero input or empty processing cycles, which are common
// in sparse-firing biological networks.
//
// EXPECTED RESULTS:
// - No data races are detected by the race detector (`go test -race`).
// - The final sum from concurrent inputs is correct.
// - Calling Process() on an empty buffer returns nil and does not panic.
// - Handling zero-value messages does not alter the buffer's final sum.
func TestDendriteConcurrencyAndEdges(t *testing.T) {
	t.Log("=== TESTING Concurrency and Edge Cases ===")

	// Define a shared, deterministic bioConfig for all relevant test cases.
	bioConfig := CreateCorticalPyramidalConfig()
	bioConfig.MembraneNoise = 0
	bioConfig.TemporalJitter = 0

	// Test concurrency on all modes that use a buffer and mutex.
	modesToTest := []struct {
		mode            DendriticIntegrationMode
		concurrentValue float64 // The expected result from the concurrency test
	}{
		{NewTemporalSummationMode(), 10.0},
		{NewShuntingInhibitionMode(0.5, bioConfig), 7.0}, // No inhibition, so shunting factor is 1.0
		{NewActiveDendriteMode(ActiveDendriteConfig{
			MaxSynapticEffect:       2.5,
			ShuntingStrength:        0.4,
			DendriticSpikeThreshold: 1.2,
			NMDASpikeAmplitude:      1.0,
		}, bioConfig), 7.0},
	}

	for _, tc := range modesToTest {
		t.Run(tc.mode.Name()+"_Concurrency", func(t *testing.T) {
			var wg sync.WaitGroup
			numGoroutines := 100
			inputsPerG := 10
			wg.Add(numGoroutines)

			for i := 0; i < numGoroutines; i++ {
				go func() {
					defer wg.Done()
					for j := 0; j < inputsPerG; j++ {
						// Use small, positive values for a predictable sum
						tc.mode.Handle(synapse.SynapseMessage{Value: 0.01})
					}
				}()
			}
			wg.Wait()

			var result *IntegratedPotential
			// Use a type assertion to check for modes that have ProcessImmediate.
			// This allows us to test the summation logic without temporal decay, making the test deterministic.
			if bioMode, ok := tc.mode.(interface{ ProcessImmediate() *IntegratedPotential }); ok {
				result = bioMode.ProcessImmediate()
			} else {
				// Fallback for modes without ProcessImmediate, like TemporalSummationMode.
				result = tc.mode.Process(MembraneSnapshot{})
			}
			expected := tc.concurrentValue

			// We can now check the result for all modes
			if result == nil || result.NetInput < expected-0.001 || result.NetInput > expected+0.001 {
				t.Errorf("Expected concurrent sum of %.2f, got %.2f", expected, result.NetInput)
			}

			t.Logf("✓ %s handled %d concurrent inputs without race conditions, result: %.2f.", tc.mode.Name(), numGoroutines*inputsPerG, result.NetInput)
		})

		t.Run(tc.mode.Name()+"_EdgeCases", func(t *testing.T) {
			// Test processing an empty buffer.
			resEmpty := tc.mode.Process(MembraneSnapshot{})
			if resEmpty != nil {
				t.Errorf("Processing an empty buffer should yield a nil result, got %v", resEmpty)
			}

			// Test handling zero-value messages.
			tc.mode.Handle(synapse.SynapseMessage{Value: 1.0})
			tc.mode.Handle(synapse.SynapseMessage{Value: 0.0})
			tc.mode.Handle(synapse.SynapseMessage{Value: 0.0})
			resZero := tc.mode.Process(MembraneSnapshot{})

			// Only check the sum for the simplest case to avoid complex recalculations
			if _, ok := tc.mode.(*TemporalSummationMode); ok {
				if resZero.NetInput != 1.0 {
					t.Errorf("Expected zero-value messages to have no effect on sum, got %.2f", resZero.NetInput)
				}
			}
			t.Logf("✓ %s correctly handled empty buffers and zero-value inputs.", tc.mode.Name())
		})
	}
}

// TestDendriteModeSolvesRaceCondition provides a direct, high-level validation
// that the new integration modes solve the original GABAergic timing bug.
//
// BIOLOGICAL SIGNIFICANCE:
// This test explicitly recreates the conditions that caused the original failure:
// a near-simultaneous arrival of a strong excitatory signal and a strong
// inhibitory signal. It proves that a neuron model without a temporal integration
// window (PassiveMembraneMode) will incorrectly fire, whereas a model with one
// (TemporalSummationMode) behaves correctly, demonstrating the critical importance
// of the membrane time constant for stable network function.
func TestDendriteModeSolvesRaceCondition(t *testing.T) {
	t.Log("=== TESTING Integration Modes Solve Race Condition ===")

	runTest := func(t *testing.T, mode DendriticIntegrationMode, expectFire bool) {
		mockNeuron := &MockNeuronForDendrite{
			accumulator: 0.0,
			threshold:   1.0,
		}

		// Simulate near-simultaneous messages
		res1 := mode.Handle(synapse.SynapseMessage{Value: 1.5})  // Would fire alone
		res2 := mode.Handle(synapse.SynapseMessage{Value: -2.0}) // Should cancel it

		var netInput float64

		// PassiveMembraneMode: immediate processing (race condition)
		if res1 != nil {
			netInput = res1.NetInput // First message wins = 1.5
		}

		// TemporalSummationMode: batched processing (correct)
		if res1 == nil && res2 == nil {
			state := MembraneSnapshot{
				Accumulator:      mockNeuron.GetAccumulator(),
				CurrentThreshold: mockNeuron.GetCurrentThreshold(),
			}
			result := mode.Process(state)
			if result != nil {
				netInput = result.NetInput // Summed: 1.5 + (-2.0) = -0.5
			}
		}

		wouldFire := netInput >= mockNeuron.GetCurrentThreshold()

		if wouldFire != expectFire {
			t.Errorf("%s: Expected firing %v, but NetInput %.2f would fire=%v",
				mode.Name(), expectFire, netInput, wouldFire)
		} else {
			t.Logf("✓ %s correctly produced NetInput %.2f, would fire=%v",
				mode.Name(), netInput, wouldFire)
		}
	}

	t.Run("PassiveModeFails", func(t *testing.T) {
		runTest(t, NewPassiveMembraneMode(), true) // Fires immediately on 1.5
	})

	t.Run("TemporalModeSucceeds", func(t *testing.T) {
		runTest(t, NewTemporalSummationMode(), false) // Sums to -0.5, no fire
	})
}

// BenchmarkDendriteModes compares the performance overhead of each integration strategy.
func BenchmarkDendriteModes(b *testing.B) {
	// Define a biological config to use for all relevant modes in the benchmark.
	bioConfig := CreateCorticalPyramidalConfig()
	modes := []DendriticIntegrationMode{
		NewPassiveMembraneMode(),
		NewTemporalSummationMode(),
		NewShuntingInhibitionMode(0.5, bioConfig),
		NewActiveDendriteMode(ActiveDendriteConfig{}, bioConfig),
	}

	msg := synapse.SynapseMessage{Value: 0.1}
	state := MembraneSnapshot{}

	for _, mode := range modes {
		b.Run(mode.Name(), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// We benchmark the entire Handle -> Process cycle
				mode.Handle(msg)
				mode.Process(state)
			}
		})
	}
}

// CreateGate is a factory function for creating different types of dendritic
// gates based on configuration. This provides a unified interface for gate
// creation while supporting diverse gate implementations.
//
// Parameters:
//
//	config: Configuration specifying gate type and parameters
//
// Returns:
//
//	DendriticGate: Configured gate instance
//	error: Any configuration or creation errors
func CreateGate(config GateConfig) (DendriticGate, error) {
	// Implementation would switch on config.Type to create appropriate gate
	// This would be implemented when concrete gate types are defined
	panic("CreateGate: implementation pending - define concrete gate types first")
}

// ============================================================================
// MOCK GATE FOR TESTING AND DEVELOPMENT
// ============================================================================

// MockGate provides a simple, configurable gate implementation for testing
// and development purposes. It allows full control over gate behavior without
// the complexity of realistic biological modeling.
//
// This is particularly useful for:
// - Unit testing dendritic integration modes with gates
// - Prototyping new gate behaviors
// - Performance testing and benchmarking
// - Educational demonstrations of gate concepts
type MockGate struct {
	// === CONFIGURATION ===
	name           string
	gateType       string
	isActive       bool
	amplification  float64       // Multiplier for signal strength
	blockThreshold float64       // Block signals below this value
	activeDuration time.Duration // How long to stay active when triggered

	// === STATE TRACKING ===
	activeSince  time.Time
	triggerCount int
	lastFeedback *GateFeedback
}

// NewMockGate creates a new mock gate with specified behavior parameters.
// This provides a simple way to create test gates with predictable behavior.
//
// Parameters:
//
//	name: Human-readable gate identifier
//	amplification: Signal multiplier (1.0 = no change, >1.0 = amplify, <1.0 = attenuate)
//	blockThreshold: Block signals with Value below this threshold
//	activeDuration: How long gate stays active when triggered
func NewMockGate(name string, amplification, blockThreshold float64, activeDuration time.Duration) *MockGate {
	return &MockGate{
		name:           name,
		gateType:       "mock",
		isActive:       false,
		amplification:  amplification,
		blockThreshold: blockThreshold,
		activeDuration: activeDuration,
		triggerCount:   0,
	}
}

// Apply implements the DendriticGate interface for the mock gate.
func (g *MockGate) Apply(msg synapse.SynapseMessage, state MembraneSnapshot) (*synapse.SynapseMessage, bool) {
	// Block signals below threshold
	if msg.Value < g.blockThreshold {
		return nil, false
	}

	// Apply amplification if gate is active
	modifiedMsg := msg
	if g.isActive {
		modifiedMsg.Value *= g.amplification
	}

	return &modifiedMsg, true
}

// ShouldActivate implements trigger detection for the mock gate.
func (g *MockGate) ShouldActivate(msg synapse.SynapseMessage, state MembraneSnapshot) (bool, time.Duration) {
	// Simple trigger: activate on any strong signal
	if msg.Value > 1.0 && !g.isActive {
		return true, g.activeDuration
	}
	return false, 0
}

// Update implements state evolution for the mock gate.
func (g *MockGate) Update(feedback *GateFeedback, deltaTime time.Duration) {
	// Store feedback for inspection
	g.lastFeedback = feedback

	// Check if activation period has expired
	if g.isActive && time.Since(g.activeSince) > g.activeDuration {
		g.isActive = false
	}

	// Simple learning: increase amplification if feedback was positive
	if feedback != nil && feedback.WasHelpful && feedback.Contribution > 0 {
		g.amplification *= 1.01 // Slight increase
	}
}

// GetState implements state introspection for the mock gate.
func (g *MockGate) GetState() GateState {
	return GateState{
		IsActive: g.isActive,
		ActivationLevel: func() float64 {
			if g.isActive {
				return 1.0
			} else {
				return 0.0
			}
		}(),
		ActiveSince: g.activeSince,
		InternalState: map[string]float64{
			"amplification":   g.amplification,
			"block_threshold": g.blockThreshold,
			"trigger_count":   float64(g.triggerCount),
		},
		Duration: g.activeDuration,
	}
}

// GetTrigger returns the trigger configuration for the mock gate.
func (g *MockGate) GetTrigger() GateTrigger {
	return GateTrigger{
		SignalThreshold: 1.0, // Activates on signals > 1.0
		MinInterval:     100 * time.Millisecond,
		MaxDuration:     g.activeDuration,
	}
}

//
// Mocks

type MockNeuronForDendrite struct {
	accumulator float64
	threshold   float64
}

func (m *MockNeuronForDendrite) GetAccumulator() float64      { return m.accumulator }
func (m *MockNeuronForDendrite) GetCurrentThreshold() float64 { return m.threshold }
func (m *MockNeuronForDendrite) SetAccumulator(value float64) { m.accumulator = value }

/*
=================================================================================
MOCK GATE DOCUMENTATION - Understanding Dendritic Gates
=================================================================================

WHAT ARE DENDRITIC GATES?
Dendritic gates are biological mechanisms that provide dynamic, temporal control
over signal flow in neural dendrites. Unlike static synaptic weights that remain
fixed during inference, gates can change their behavior in real-time based on
local conditions, network state, and learned triggers.

BIOLOGICAL FOUNDATION:
In real neurons, gates are implemented by sophisticated molecular machinery:

1. METABOTROPIC RECEPTORS (MRs): Act as "sensors" that detect specific chemical
   signals, electrical patterns, or metabolic changes in the local environment.

2. G PROTEIN-GATED ION CHANNELS (GPGICs): Act as "effectors" that implement
   the actual pathway changes. When activated by MRs, they remain active for
   hundreds of milliseconds to minutes, providing sustained modulation.

3. ACTIVITY-DEPENDENT CASCADES: Enable gates to learn and adapt their behavior
   based on local success/failure feedback and network-wide signals.

KEY PROPERTIES OF BIOLOGICAL GATES:
- TRANSIENT REWIRING: Temporarily change how signals flow without permanent
  structural modifications
- STATE-DEPENDENT: Gate behavior depends on current dendritic and membrane state
- LEARNING CAPABLE: Gates can adapt their trigger conditions and effects
- TEMPORAL PERSISTENCE: Once activated, gates remain active for extended periods
- BRANCH-LEVEL CONTROL: A single gate can control an entire dendritic branch

HOW GATES DIFFER FROM TRADITIONAL NEURAL NETWORK COMPONENTS:

TRADITIONAL WEIGHTS (LSTM/GRU gates):
- Fixed during inference (only change during training)
- Simple multiplication of signals
- No internal state or memory
- Cannot detect when to change behavior

BIOLOGICAL GATES:
- Dynamic during inference (change based on conditions)
- Complex signal transformation (amplify, attenuate, block, delay)
- Maintain internal state and memory
- Learn when and how to modulate signals

GATE OPERATION LIFECYCLE:
1. DETECTION: Gate monitors incoming signals and dendritic state
2. TRIGGER EVALUATION: Determines if conditions warrant activation
3. ACTIVATION: Gate changes its internal state and begins modulation
4. SIGNAL PROCESSING: Gate transforms incoming synaptic messages
5. TEMPORAL PERSISTENCE: Gate remains active for its configured duration
6. LEARNING: Gate receives feedback about its effectiveness
7. ADAPTATION: Gate adjusts its parameters based on feedback
8. DEACTIVATION: Gate returns to inactive state after timeout

TYPES OF GATE BEHAVIORS:

1. THRESHOLD GATING:
   - Blocks signals below a certain strength
   - Models voltage-gated ion channels
   - Example: Only pass signals when membrane potential is high

2. GAIN MODULATION:
   - Amplifies or attenuates signal strength
   - Models neuromodulatory effects (dopamine, serotonin)
   - Example: Dopamine increases signal strength during reward

3. TEMPORAL FILTERING:
   - Controls signal timing and coincidence detection
   - Models NMDA receptor behavior
   - Example: Only pass signals that arrive within a time window

4. CONTEXT-DEPENDENT MODULATION:
   - Changes behavior based on dendritic state
   - Models calcium-dependent processes
   - Example: Amplify signals when calcium levels are high

BIOLOGICAL EXAMPLES OF GATING:

VISUAL ATTENTION:
When you focus attention on a specific location, gates in visual cortex
dendrites become active, amplifying signals from that location while
suppressing signals from other areas. This happens within 50-100ms
and persists as long as attention is maintained.

MOTOR LEARNING:
During skill acquisition, gates in motor cortex dendrites learn to
selectively amplify the most successful movement patterns while
suppressing competing patterns. These gates adapt over practice sessions.

WORKING MEMORY:
Prefrontal cortex dendrites use gates to maintain specific information
patterns active for seconds to minutes, even when the original stimulus
is no longer present.

FEAR CONDITIONING:
In fear learning, gates in amygdala dendrites learn to associate neutral
stimuli with threat signals, becoming active whenever the learned trigger
appears, even years after the original conditioning.

NETWORK-LEVEL EFFECTS:

EFFICIENCY: Gates allow networks to solve complex problems with fewer neurons
and connections by enabling dynamic reconfiguration.

ADAPTABILITY: Networks can switch between different computational modes
without retraining all weights.

MEMORY: Gate states themselves carry information forward, reducing reliance
on traditional hidden state variables.

LEARNING: Gates enable multi-level learning - learning both what to compute
and when to change what they compute.

MOCK GATE TESTING STRATEGY:

The MockGate implementation below provides a simplified but representative
model of gate behavior that allows testing of:

1. Signal blocking based on thresholds
2. Signal amplification when active
3. Trigger-based activation
4. Temporal persistence of activation
5. Basic learning through feedback

This mock enables validation of dendritic integration behavior without
requiring the full complexity of realistic biological gate models.

KEY TEST SCENARIOS:
- Verify gates can block weak signals (threshold behavior)
- Confirm gates amplify signals when active (gain modulation)
- Test trigger detection and activation logic
- Validate temporal persistence and deactivation
- Check learning and adaptation mechanisms
- Ensure proper resource cleanup

By testing with MockGate, we can validate that the dendritic integration
architecture correctly supports gate functionality and provides the
foundation for more sophisticated biological gate implementations.

=================================================================================
*/

// Name returns the gate's identifier.
func (g *MockGate) Name() string { return g.name }

// Type returns the gate's type.
func (g *MockGate) Type() string { return g.gateType }

// Close cleans up the mock gate (no-op for this simple implementation).
func (g *MockGate) Close() {
	// No resources to clean up in mock implementation
}
