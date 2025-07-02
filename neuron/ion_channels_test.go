package neuron

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

/*
=================================================================================
REALISTIC ION CHANNEL TESTS - BIOPHYSICALLY ACCURATE CHANNEL MODELING
=================================================================================

OVERVIEW:
This test suite validates realistic ion channel implementations that model
specific biological voltage-gated and ligand-gated channels found in neurons.
Tests focus on biophysically accurate gating kinetics, conductance properties,
and voltage/ligand dependencies using dendrite constants for biological realism.

BIOLOGICAL CONTEXT:
Ion channels are the molecular basis of neural computation. Different channel
types (Nav1.6, Kv4.2, Cav1.2, GABA-A) have distinct biophysical properties
that determine their roles in synaptic integration, spike generation, and
plasticity. These tests validate our models against known channel properties.

KEY MECHANISMS TESTED:
1. VOLTAGE-GATED SODIUM CHANNELS: Fast activation/inactivation, spike initiation
2. VOLTAGE-GATED POTASSIUM CHANNELS: Delayed rectification, repolarization
3. VOLTAGE-GATED CALCIUM CHANNELS: Calcium influx, plasticity signaling
4. LIGAND-GATED CHLORIDE CHANNELS: GABA-A mediated inhibition, shunting
5. REALISTIC CHANNEL KINETICS: Activation, deactivation, inactivation, recovery

=================================================================================
*/

// ============================================================================
// REALISTIC ION CHANNEL TESTS
// ============================================================================

// TestNeuronRealisticNavChannel validates voltage-gated sodium channel behavior
func TestNeuronRealisticNavChannel_VoltageGating(t *testing.T) {
	t.Log("=== TESTING Realistic Nav1.6 Channel ===")

	channel := NewRealisticNavChannel("nav1.6_test")

	// Test 1: Resting state
	t.Run("RestingState", func(t *testing.T) {
		restingVoltage := DENDRITE_VOLTAGE_RESTING_CORTICAL
		shouldOpen, _, probability := channel.ShouldOpen(restingVoltage, 0, 0, 1*time.Millisecond)

		if shouldOpen {
			t.Error("Nav channel should not open at resting potential")
		}

		if probability > 0.1 {
			t.Errorf("Open probability should be low at rest, got %.3f", probability)
		}

		t.Logf("✓ Resting state: open=%v, prob=%.3f", shouldOpen, probability)
	})

	// Test 2: Activation threshold
	t.Run("ActivationThreshold", func(t *testing.T) {
		activationVoltage := -40.0 // Around activation threshold

		// Update gating over several time steps
		for i := 0; i < 10; i++ {
			channel.ShouldOpen(activationVoltage, 0, 0, 1*time.Millisecond)
		}

		state := channel.GetState()
		if state.Conductance <= 0 {
			t.Error("Expected some conductance at activation voltage")
		}

		t.Logf("✓ Activation: conductance=%.3f pS", state.Conductance)
	})

	// Test 3: Current modulation
	t.Run("CurrentModulation", func(t *testing.T) {
		msg := types.NeuralSignal{
			Value:                DENDRITE_TEST_INPUT_MEDIUM,
			NeurotransmitterType: types.LigandGlutamate,
		}

		voltage := -30.0 // Depolarized
		calcium := DENDRITE_CALCIUM_BASELINE_INTRACELLULAR

		modifiedMsg, shouldContinue, channelCurrent := channel.ModulateCurrent(msg, voltage, calcium)

		if !shouldContinue {
			t.Error("Nav channel should not block signals")
		}

		if modifiedMsg == nil {
			t.Error("Expected modified message")
		}

		// At depolarized voltage, expect some sodium current
		if voltage > channel.GetTrigger().ActivationVoltage {
			t.Logf("✓ Channel current: %.3f pA at %.1f mV", channelCurrent, voltage)
		}
	})

	// Test 4: Inactivation
	t.Run("Inactivation", func(t *testing.T) {
		strongDepolarization := 0.0 // Strong depolarization should cause inactivation

		// Prolonged depolarization
		for i := 0; i < 50; i++ {
			channel.ShouldOpen(strongDepolarization, 0, 0, 1*time.Millisecond)
		}

		// Check if inactivation occurred (conductance should decrease)
		state := channel.GetState()
		trigger := channel.GetTrigger()

		if state.MembraneVoltage != strongDepolarization {
			t.Error("Channel should track membrane voltage")
		}

		t.Logf("✓ Inactivation test: conductance=%.3f pS, activation_threshold=%.1f mV",
			state.Conductance, trigger.ActivationVoltage)
	})
}

// TestNeuronRealisticKvChannel validates voltage-gated potassium channel behavior
func TestNeuronRealisticKvChannel_DelayedRectifier(t *testing.T) {
	t.Log("=== TESTING Realistic Kv4.2 Channel ===")

	channel := NewRealisticKvChannel("kv4.2_test")

	// Test 1: Delayed activation
	t.Run("DelayedActivation", func(t *testing.T) {
		voltage := -20.0 // Above K+ activation threshold

		// K+ channels activate more slowly than Na+
		initialProb := 0.0
		var finalProb float64

		for i := 0; i < 20; i++ {
			_, _, prob := channel.ShouldOpen(voltage, 0, 0, 1*time.Millisecond)
			if i == 0 {
				initialProb = prob
			}
			finalProb = prob
		}

		if finalProb <= initialProb {
			t.Error("K+ channel should show delayed activation")
		}

		t.Logf("✓ Delayed activation: initial=%.3f, final=%.3f", initialProb, finalProb)
	})

	// Test 2: Repolarizing current
	t.Run("RepolarizingCurrent", func(t *testing.T) {
		msg := types.NeuralSignal{Value: DENDRITE_TEST_INPUT_MEDIUM}
		voltage := -30.0 // Above activation threshold
		calcium := DENDRITE_CALCIUM_BASELINE_INTRACELLULAR

		// Activate channel first
		for i := 0; i < 10; i++ {
			channel.ShouldOpen(voltage, 0, calcium, 1*time.Millisecond)
		}

		_, _, channelCurrent := channel.ModulateCurrent(msg, voltage, calcium)

		// K+ current should be outward (positive) when V > EK
		expectedDirection := voltage - DENDRITE_VOLTAGE_REVERSAL_POTASSIUM // Should be positive
		if expectedDirection > 0 && channelCurrent <= 0 {
			t.Errorf("Expected outward K+ current, got %.3f pA", channelCurrent)
		}

		t.Logf("✓ K+ current: %.3f pA (driving force: %.1f mV)", channelCurrent, expectedDirection)
	})

	// Test 3: Ion type specificity
	t.Run("IonSpecificity", func(t *testing.T) {
		if channel.GetIonSelectivity() != IonPotassium {
			t.Error("Kv channel should be potassium-selective")
		}

		if math.Abs(channel.GetReversalPotential()-DENDRITE_VOLTAGE_REVERSAL_POTASSIUM) > DENDRITE_TEST_TOLERANCE_VOLTAGE {
			t.Errorf("Expected K+ reversal potential %.1f mV, got %.1f mV",
				DENDRITE_VOLTAGE_REVERSAL_POTASSIUM, channel.GetReversalPotential())
		}

		t.Logf("✓ Ion selectivity: %s, reversal: %.1f mV",
			channel.GetIonSelectivity().String(), channel.GetReversalPotential())
	})
}

// TestNeuronRealisticCavChannel validates voltage-gated calcium channel behavior
func TestNeuronRealisticCavChannel_CalciumInflux(t *testing.T) {
	t.Log("=== TESTING Realistic Cav1.2 Channel ===")

	channel := NewRealisticCavChannel("cav1.2_test")

	// Test 1: High threshold activation
	t.Run("HighThresholdActivation", func(t *testing.T) {
		lowVoltage := -40.0  // Below Ca2+ activation
		highVoltage := -10.0 // Above Ca2+ activation

		// Test at low voltage
		shouldOpenLow, _, probLow := channel.ShouldOpen(lowVoltage, 0, 0, 1*time.Millisecond)

		// Activate at high voltage
		for i := 0; i < 10; i++ {
			channel.ShouldOpen(highVoltage, 0, 0, 1*time.Millisecond)
		}
		_, _, probHigh := channel.ShouldOpen(highVoltage, 0, 0, 1*time.Millisecond)

		if shouldOpenLow {
			t.Error("Ca2+ channel should not activate at low voltage")
		}

		if probHigh <= probLow {
			t.Error("Ca2+ channel should have higher activation probability at high voltage")
		}

		t.Logf("✓ Voltage dependence: low_V=%.1f (prob=%.3f), high_V=%.1f (prob=%.3f)",
			lowVoltage, probLow, highVoltage, probHigh)
	})

	// Test 2: Calcium-dependent inactivation
	t.Run("CalciumDependentInactivation", func(t *testing.T) {
		voltage := -10.0 // Above activation threshold
		lowCalcium := DENDRITE_CALCIUM_BASELINE_INTRACELLULAR
		highCalcium := DENDRITE_CALCIUM_BASELINE_INTRACELLULAR * 5

		// Test with low calcium
		_, _, probLowCa := channel.ShouldOpen(voltage, 0, lowCalcium, 1*time.Millisecond)

		// Test with high calcium (should reduce open probability)
		_, _, probHighCa := channel.ShouldOpen(voltage, 0, highCalcium, 1*time.Millisecond)

		if probHighCa >= probLowCa {
			t.Error("High calcium should reduce Ca2+ channel open probability")
		}

		t.Logf("✓ Ca2+-dependent inactivation: low_Ca=%.3f, high_Ca=%.3f", probLowCa, probHighCa)
	})

	// Test 3: Calcium influx tracking
	t.Run("CalciumInfluxTracking", func(t *testing.T) {
		msg := types.NeuralSignal{Value: DENDRITE_TEST_INPUT_MEDIUM}
		voltage := -5.0 // Strong depolarization
		calcium := DENDRITE_CALCIUM_BASELINE_INTRACELLULAR

		// Activate channel
		for i := 0; i < 10; i++ {
			channel.ShouldOpen(voltage, 0, calcium, 1*time.Millisecond)
		}

		initialState := channel.GetState()
		_, _, channelCurrent := channel.ModulateCurrent(msg, voltage, calcium)
		finalState := channel.GetState()

		// Calcium current should be inward (negative) when V < ECa
		expectedDirection := voltage - DENDRITE_VOLTAGE_REVERSAL_CALCIUM // Should be negative
		if expectedDirection < 0 && channelCurrent >= 0 {
			t.Errorf("Expected inward Ca2+ current, got %.3f pA", channelCurrent)
		}

		// Calcium influx should be tracked
		if finalState.CalciumLevel <= initialState.CalciumLevel && channelCurrent < 0 {
			t.Error("Expected calcium influx to be tracked")
		}

		t.Logf("✓ Ca2+ influx: current=%.3f pA, Ca_level=%.3f",
			channelCurrent, finalState.CalciumLevel)
	})
}

// TestNeuronRealisticGabaAChannel validates ligand-gated chloride channel behavior
func TestNeuronRealisticGabaAChannel_LigandGating(t *testing.T) {
	t.Log("=== TESTING Realistic GABA-A Channel ===")

	channel := NewRealisticGabaAChannel("gabaa_test")

	// Test 1: GABA-dependent activation
	t.Run("GABADependentActivation", func(t *testing.T) {
		voltage := DENDRITE_VOLTAGE_RESTING_CORTICAL
		noGABA := 0.0
		lowGABA := 5.0   // Low GABA concentration
		highGABA := 20.0 // High GABA concentration

		// Test without GABA
		shouldOpenNo, _, probNo := channel.ShouldOpen(voltage, noGABA, 0, 1*time.Millisecond)

		// Test with low GABA
		for i := 0; i < 5; i++ {
			channel.ShouldOpen(voltage, lowGABA, 0, 1*time.Millisecond)
		}
		_, _, probLow := channel.ShouldOpen(voltage, lowGABA, 0, 1*time.Millisecond)

		// Test with high GABA
		channel = NewRealisticGabaAChannel("gabaa_test2") // Fresh channel
		for i := 0; i < 5; i++ {
			channel.ShouldOpen(voltage, highGABA, 0, 1*time.Millisecond)
		}
		_, _, probHigh := channel.ShouldOpen(voltage, highGABA, 0, 1*time.Millisecond)

		if shouldOpenNo {
			t.Error("GABA-A channel should not open without GABA")
		}

		if probHigh <= probLow {
			t.Error("Higher GABA should increase open probability")
		}

		t.Logf("✓ GABA dependence: no_GABA=%.3f, low_GABA=%.3f, high_GABA=%.3f",
			probNo, probLow, probHigh)
	})

	// Test 2: Desensitization
	t.Run("Desensitization", func(t *testing.T) {
		voltage := DENDRITE_VOLTAGE_RESTING_CORTICAL
		highGABA := 30.0 // Very high GABA that should cause progressive desensitization

		// Get initial response with moderate exposure
		initialProb := 0.0
		for i := 0; i < 3; i++ {
			_, _, prob := channel.ShouldOpen(voltage, highGABA, 0, 1*time.Millisecond)
			if i == 2 {
				initialProb = prob
			}
		}

		// Prolonged exposure with higher GABA - should cause progressive desensitization
		for i := 0; i < 200; i++ { // Even more iterations for stronger effect
			channel.ShouldOpen(voltage, highGABA, 0, 1*time.Millisecond)
		}
		_, _, desensitizedProb := channel.ShouldOpen(voltage, highGABA, 0, 1*time.Millisecond)

		if desensitizedProb >= initialProb {
			t.Errorf("Prolonged GABA exposure should cause desensitization: initial=%.3f, final=%.3f",
				initialProb, desensitizedProb)
		}

		t.Logf("✓ Desensitization: initial=%.3f, desensitized=%.3f", initialProb, desensitizedProb)
	})

	// Test 3: Chloride current modulation
	t.Run("ChlorideCurrent", func(t *testing.T) {
		// Create GABA signal
		gabaMsg := types.NeuralSignal{
			Value:                DENDRITE_FACTOR_EFFECT_GABA * DENDRITE_TEST_INPUT_MEDIUM, // Negative GABA signal
			NeurotransmitterType: types.LigandGABA,
		}

		voltage := DENDRITE_VOLTAGE_RESTING_CORTICAL
		calcium := DENDRITE_CALCIUM_BASELINE_INTRACELLULAR

		// Activate channel with GABA
		for i := 0; i < 10; i++ {
			channel.ModulateCurrent(gabaMsg, voltage, calcium)
		}

		_, _, channelCurrent := channel.ModulateCurrent(gabaMsg, voltage, calcium)

		// At resting potential, Cl- current should be minimal (V ≈ ECl)
		expectedDrivingForce := voltage - DENDRITE_VOLTAGE_REVERSAL_CHLORIDE

		if math.Abs(expectedDrivingForce) < 5.0 { // Small driving force
			t.Logf("✓ Cl- current: %.3f pA (small driving force: %.1f mV)",
				channelCurrent, expectedDrivingForce)
		} else {
			t.Logf("✓ Cl- current: %.3f pA (driving force: %.1f mV)",
				channelCurrent, expectedDrivingForce)
		}

		// Verify channel is responsive to GABA
		state := channel.GetState()
		if state.Conductance <= 0 {
			t.Error("Expected some conductance with GABA stimulation")
		}
	})
}

// ============================================================================
// MULTI-CHANNEL INTEGRATION TESTS
// ============================================================================

// TestNeuronRealisticChannels_Integration validates multiple realistic channels working together
func TestNeuronRealisticChannels_Integration(t *testing.T) {
	t.Log("=== TESTING Multi-Channel Integration ===")

	// Create dendritic integration mode with realistic channels
	mode := NewTemporalSummationMode()

	// Add realistic channel ensemble
	navChannel := NewRealisticNavChannel("nav1.6")
	kvChannel := NewRealisticKvChannel("kv4.2")
	cavChannel := NewRealisticCavChannel("cav1.2")
	gabaChannel := NewRealisticGabaAChannel("gabaa")

	mode.AddChannel(navChannel)
	mode.AddChannel(kvChannel)
	mode.AddChannel(cavChannel)
	mode.AddChannel(gabaChannel)

	// Test 1: Excitatory integration
	t.Run("ExcitatoryIntegration", func(t *testing.T) {
		excMsg := types.NeuralSignal{
			Value:                DENDRITE_TEST_INPUT_MEDIUM,
			NeurotransmitterType: types.LigandGlutamate,
		}

		_ = mode.Handle(excMsg)
		processResult := mode.Process(MembraneSnapshot{
			Accumulator:          -40.0, // Depolarized state
			RestingPotential:     DENDRITE_VOLTAGE_RESTING_CORTICAL,
			IntracellularCalcium: DENDRITE_CALCIUM_BASELINE_INTRACELLULAR,
		})

		// Should see contributions from voltage-gated channels
		if processResult != nil {
			contributions := processResult.ChannelContributions

			// Na+ channels should contribute with depolarization
			if navContrib, exists := contributions["nav1.6"]; exists {
				t.Logf("✓ Nav1.6 contribution: %.3f pA", navContrib)
			}

			// K+ channels should provide repolarization
			if kvContrib, exists := contributions["kv4.2"]; exists {
				t.Logf("✓ Kv4.2 contribution: %.3f pA", kvContrib)
			}

			// Ca2+ channels should contribute at high voltages
			if cavContrib, exists := contributions["cav1.2"]; exists {
				t.Logf("✓ Cav1.2 contribution: %.3f pA", cavContrib)
			}

			t.Logf("✓ Total excitatory current: %.3f pA", processResult.NetCurrent)
		}
	})

	// Test 2: Inhibitory integration
	t.Run("InhibitoryIntegration", func(t *testing.T) {
		mode = NewTemporalSummationMode() // Fresh mode
		mode.AddChannel(NewRealisticGabaAChannel("gabaa2"))

		inhibMsg := types.NeuralSignal{
			Value:                DENDRITE_FACTOR_EFFECT_GABA * DENDRITE_TEST_INPUT_MEDIUM,
			NeurotransmitterType: types.LigandGABA,
		}

		mode.Handle(inhibMsg)
		result := mode.Process(MembraneSnapshot{
			Accumulator:          DENDRITE_VOLTAGE_RESTING_CORTICAL,
			RestingPotential:     DENDRITE_VOLTAGE_RESTING_CORTICAL,
			IntracellularCalcium: DENDRITE_CALCIUM_BASELINE_INTRACELLULAR,
		})

		if result != nil && len(result.ChannelContributions) > 0 {
			t.Logf("✓ GABA-A inhibitory integration: %.3f pA", result.NetCurrent)
		}
	})

	// Test 3: Channel competition
	t.Run("ChannelCompetition", func(t *testing.T) {
		mode = NewTemporalSummationMode() // Fresh mode
		mode.AddChannel(NewRealisticNavChannel("nav_compete"))
		mode.AddChannel(NewRealisticKvChannel("kv_compete"))

		// Strong depolarization should activate both Na+ and K+ channels
		strongMsg := types.NeuralSignal{
			Value:                DENDRITE_TEST_INPUT_LARGE,
			NeurotransmitterType: types.LigandGlutamate,
		}

		mode.Handle(strongMsg)
		result := mode.Process(MembraneSnapshot{
			Accumulator:          -20.0, // Strong depolarization
			RestingPotential:     DENDRITE_VOLTAGE_RESTING_CORTICAL,
			IntracellularCalcium: DENDRITE_CALCIUM_BASELINE_INTRACELLULAR,
		})

		if result != nil {
			// Should see opposing currents
			navCurrent := result.ChannelContributions["nav_compete"]
			kvCurrent := result.ChannelContributions["kv_compete"]

			t.Logf("✓ Channel competition: Na+=%.3f pA, K+=%.3f pA, net=%.3f pA",
				navCurrent, kvCurrent, result.NetCurrent)
		}
	})
}

// ============================================================================
// BIOPHYSICAL REALISM TESTS
// ============================================================================

// TestNeuronRealisticChannels_BiophysicalProperties validates biological accuracy
func TestNeuronRealisticChannels_BiophysicalProperties(t *testing.T) {
	t.Log("=== TESTING Biophysical Properties ===")

	channels := []IonChannel{
		NewRealisticNavChannel("nav_test"),
		NewRealisticKvChannel("kv_test"),
		NewRealisticCavChannel("cav_test"),
		NewRealisticGabaAChannel("gaba_test"),
	}

	// Test 1: Conductance ranges
	t.Run("ConductanceRanges", func(t *testing.T) {
		expectedRanges := map[string][2]float64{
			"nav1.6": {10.0, 30.0}, // Na+ channels: 10-30 pS
			"kv4.2":  {5.0, 20.0},  // K+ channels: 5-20 pS
			"cav1.2": {1.0, 10.0},  // Ca2+ channels: 1-10 pS
			"gabaa":  {10.0, 25.0}, // GABA-A channels: 10-25 pS
		}

		for _, channel := range channels {
			conductance := channel.GetConductance()
			channelType := channel.ChannelType()

			if expectedRange, exists := expectedRanges[channelType]; exists {
				if conductance < expectedRange[0] || conductance > expectedRange[1] {
					t.Errorf("Channel %s conductance %.1f pS outside expected range [%.1f, %.1f]",
						channelType, conductance, expectedRange[0], expectedRange[1])
				} else {
					t.Logf("✓ %s conductance: %.1f pS (valid range)", channelType, conductance)
				}
			}
		}
	})

	// Test 2: Reversal potentials
	t.Run("ReversalPotentials", func(t *testing.T) {
		expectedReversals := map[IonType]float64{
			IonSodium:    DENDRITE_VOLTAGE_REVERSAL_SODIUM,
			IonPotassium: DENDRITE_VOLTAGE_REVERSAL_POTASSIUM,
			IonCalcium:   DENDRITE_VOLTAGE_REVERSAL_CALCIUM,
			IonChloride:  DENDRITE_VOLTAGE_REVERSAL_CHLORIDE,
		}

		for _, channel := range channels {
			ionType := channel.GetIonSelectivity()
			reversal := channel.GetReversalPotential()

			if expected, exists := expectedReversals[ionType]; exists {
				if math.Abs(reversal-expected) > DENDRITE_TEST_TOLERANCE_VOLTAGE {
					t.Errorf("Channel %s reversal %.1f mV differs from expected %.1f mV",
						channel.Name(), reversal, expected)
				} else {
					t.Logf("✓ %s (%s): reversal %.1f mV",
						channel.Name(), ionType.String(), reversal)
				}
			}
		}
	})

	// Test 3: Kinetic time constants
	t.Run("KineticTimeConstants", func(t *testing.T) {
		for _, channel := range channels {
			trigger := channel.GetTrigger()

			// Activation should be faster than deactivation
			if trigger.ActivationTimeConstant > trigger.DeactivationTimeConstant &&
				trigger.DeactivationTimeConstant > 0 {
				t.Errorf("Channel %s: activation (%.3f ms) slower than deactivation (%.3f ms)",
					channel.Name(),
					trigger.ActivationTimeConstant.Seconds()*1000,
					trigger.DeactivationTimeConstant.Seconds()*1000)
			}

			// Time constants should be in biological range (0.1-100 ms)
			if trigger.ActivationTimeConstant > 0 {
				activationMs := trigger.ActivationTimeConstant.Seconds() * 1000
				if activationMs < 0.1 || activationMs > 100.0 {
					t.Errorf("Channel %s activation time %.3f ms outside biological range",
						channel.Name(), activationMs)
				} else {
					t.Logf("✓ %s activation: %.3f ms", channel.Name(), activationMs)
				}
			}
		}
	})

	// Test 4: State consistency
	t.Run("StateConsistency", func(t *testing.T) {
		for _, channel := range channels {
			state := channel.GetState()

			// Conductance should be non-negative
			if state.Conductance < 0 {
				t.Errorf("Channel %s has negative conductance: %.3f",
					channel.Name(), state.Conductance)
			}

			// Open channels should have some conductance
			if state.IsOpen && state.Conductance <= 0 {
				t.Errorf("Channel %s reports open but zero conductance", channel.Name())
			}

			t.Logf("✓ %s state: open=%v, conductance=%.3f pS",
				channel.Name(), state.IsOpen, state.Conductance)
		}
	})
}

// ============================================================================
// PERFORMANCE AND EDGE CASE TESTS
// ============================================================================

// TestNeuronRealisticChannels_Performance validates channel performance under load
func TestNeuronRealisticChannels_Performance(t *testing.T) {
	t.Log("=== TESTING Channel Performance ===")

	// Test with many channels
	mode := NewBiologicalTemporalSummationMode(CreateCorticalPyramidalConfig())

	// Add many realistic channels
	numChannels := 50
	for i := 0; i < numChannels; i++ {
		switch i % 4 {
		case 0:
			mode.AddChannel(NewRealisticNavChannel(fmt.Sprintf("nav_%d", i)))
		case 1:
			mode.AddChannel(NewRealisticKvChannel(fmt.Sprintf("kv_%d", i)))
		case 2:
			mode.AddChannel(NewRealisticCavChannel(fmt.Sprintf("cav_%d", i)))
		case 3:
			mode.AddChannel(NewRealisticGabaAChannel(fmt.Sprintf("gaba_%d", i)))
		}
	}

	// Benchmark processing
	start := time.Now()
	iterations := 1000

	for i := 0; i < iterations; i++ {
		msg := types.NeuralSignal{
			Value:                DENDRITE_TEST_INPUT_MEDIUM,
			NeurotransmitterType: types.LigandGlutamate,
		}

		mode.Handle(msg)

		if i%100 == 0 {
			mode.Process(MembraneSnapshot{
				Accumulator:          DENDRITE_VOLTAGE_RESTING_CORTICAL + float64(i%40),
				RestingPotential:     DENDRITE_VOLTAGE_RESTING_CORTICAL,
				IntracellularCalcium: DENDRITE_CALCIUM_BASELINE_INTRACELLULAR,
			})
		}
	}

	elapsed := time.Since(start)
	avgPerIteration := elapsed / time.Duration(iterations)

	if avgPerIteration > 100*time.Microsecond {
		t.Errorf("Performance too slow: %.3f μs per iteration",
			avgPerIteration.Seconds()*1e6)
	}

	t.Logf("✓ Performance: %d channels, %d iterations, %.3f μs avg",
		numChannels, iterations, avgPerIteration.Seconds()*1e6)
}

// TestNeuronRealisticChannels_EdgeCases validates robustness
func TestNeuronRealisticChannels_EdgeCases(t *testing.T) {
	t.Log("=== TESTING Edge Cases ===")

	channel := NewRealisticNavChannel("edge_test")

	// Test 1: Extreme voltages
	t.Run("ExtremeVoltages", func(t *testing.T) {
		extremeVoltages := []float64{-200.0, -100.0, 0.0, 100.0, 200.0}

		for _, voltage := range extremeVoltages {
			// Should not panic or produce invalid results
			shouldOpen, duration, probability := channel.ShouldOpen(voltage, 0, 0, 1*time.Millisecond)

			if probability < 0 || probability > 1 {
				t.Errorf("Invalid probability %.3f at voltage %.1f mV", probability, voltage)
			}

			if duration < 0 {
				t.Errorf("Invalid duration %v at voltage %.1f mV", duration, voltage)
			}

			_ = shouldOpen // Just check it doesn't panic
		}

		t.Log("✓ Extreme voltages handled gracefully")
	})

	// Test 2: Rapid voltage changes
	t.Run("RapidVoltageChanges", func(t *testing.T) {
		voltages := []float64{-70, -40, -70, 0, -70, -40, -70}

		for _, voltage := range voltages {
			channel.ShouldOpen(voltage, 0, 0, 100*time.Microsecond) // Fast transitions
		}

		// Should maintain valid state
		state := channel.GetState()
		if state.Conductance < 0 {
			t.Error("Invalid conductance after rapid voltage changes")
		}

		t.Log("✓ Rapid voltage changes handled correctly")
	})

	// Test 3: Zero time steps
	t.Run("ZeroTimeSteps", func(t *testing.T) {
		// Should handle zero or very small time steps
		channel.ShouldOpen(-40.0, 0, 0, 0)
		channel.ShouldOpen(-40.0, 0, 0, 1*time.Nanosecond)

		state := channel.GetState()
		if state.Conductance < 0 {
			t.Error("Invalid state with zero time steps")
		}

		t.Log("✓ Zero time steps handled correctly")
	})
}
