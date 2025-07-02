package neuron

import (
	"math"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

/*
=================================================================================
ION CHANNEL KINETICS TESTS - DETAILED BIOPHYSICAL MODELING VALIDATION
=================================================================================

OVERVIEW:
This comprehensive test suite validates realistic ion channel implementations
that model specific biological voltage-gated and ligand-gated channels found
in neurons. Tests focus on biophysically accurate gating kinetics, conductance
properties, and voltage/ligand dependencies using experimentally-determined
constants for biological realism.

BIOLOGICAL CONTEXT:
Ion channels are the fundamental molecular basis of neural computation. Different
channel types have evolved distinct biophysical properties that determine their
specialized roles in neural function:

- Nav1.6 (Sodium): Rapid spike initiation and propagation
- Kv4.2 (Potassium): Spike repolarization and frequency adaptation
- Cav1.2 (Calcium): Calcium signaling and synaptic plasticity
- GABA-A (Chloride): Fast inhibitory neurotransmission

EXPERIMENTAL BASIS:
All channel parameters are based on patch-clamp electrophysiology data from
mammalian neurons. Voltage dependence, time constants, and conductance values
match experimentally observed ranges from the literature.

KEY VALIDATION AREAS:
1. VOLTAGE DEPENDENCE: Sigmoid activation/inactivation curves with correct V1/2
2. KINETIC TIMING: Realistic time constants for biological processes
3. GATING COOPERATIVITY: Proper exponents (m³h, n⁴, m², Hill equation)
4. ACTIVITY DEPENDENCE: Use-dependent modulation and plasticity
5. MULTI-CHANNEL INTEGRATION: Realistic interactions during action potentials

=================================================================================
*/

// ============================================================================
// SODIUM CHANNEL (Nav1.6) KINETICS TESTS
// ============================================================================

// TestChannelKinetics_NavActivationCurve validates voltage-dependent activation
//
// BIOLOGICAL BASIS:
// Nav1.6 channels are responsible for the rapid upstroke of action potentials.
// They exhibit fast, voltage-dependent activation with a sigmoidal relationship
// between membrane voltage and open probability.
//
// EXPECTED BEHAVIOR:
// - V1/2 (half-activation voltage): ~-40 mV (typical for Nav1.6)
// - Slope factor: ~5 mV (steep voltage sensitivity)
// - Range: Near 0% at -80mV, near 100% at +10mV
// - Monotonic increase across voltage range
//
// CLINICAL RELEVANCE:
// Mutations affecting Nav1.6 activation cause epilepsy and movement disorders.
func TestChannelKinetics_NavActivationCurve(t *testing.T) {
	t.Log("=== TESTING Nav1.6 Activation Curve ===")
	t.Log("Biological target: V1/2 = -40mV, slope = 5mV")

	var channel *RealisticNavChannel

	// Test activation curve across physiological voltage range
	voltages := []float64{-80, -70, -60, -50, -40, -30, -20, -10, 0, 10}
	activationProbs := make([]float64, len(voltages))

	for i, voltage := range voltages {
		// Reset channel to fresh state for each voltage test
		channel = NewRealisticNavChannel("nav_kinetics_test")

		// Apply voltage for sufficient time to reach steady state (100ms >> τ_m = 1ms)
		for j := 0; j < 100; j++ {
			channel.ShouldOpen(voltage, 0, 0, 1*time.Millisecond)
		}

		// Measure steady-state activation gate (m_∞)
		activationProbs[i] = channel.activationGate
		t.Logf("Voltage: %.1f mV, Activation (m_∞): %.3f", voltage, activationProbs[i])
	}

	// VALIDATION 1: Low voltage behavior (hyperpolarized)
	// At -80mV, should be nearly closed (< 10% activation)
	if activationProbs[0] > 0.1 {
		t.Errorf("Activation too high at -80 mV: %.3f (expected < 0.1)", activationProbs[0])
	}

	// VALIDATION 2: High voltage behavior (depolarized)
	// At +10mV, should be nearly fully activated (> 95% activation)
	if activationProbs[len(activationProbs)-1] < 0.95 {
		t.Errorf("Activation too low at +10 mV: %.3f (expected > 0.95)", activationProbs[len(activationProbs)-1])
	}

	// VALIDATION 3: Monotonic increase (fundamental property of activation)
	for i := 1; i < len(activationProbs); i++ {
		if activationProbs[i] < activationProbs[i-1]-0.001 {
			t.Errorf("Activation curve not monotonic at %.1f mV (%.3f < %.3f)",
				voltages[i], activationProbs[i], activationProbs[i-1])
		}
	}

	// VALIDATION 4: Half-activation voltage (V1/2)
	// Should occur around -40 mV for Nav1.6 channels
	v50Index := -1
	for i, prob := range activationProbs {
		if prob >= 0.5 {
			v50Index = i
			break
		}
	}

	if v50Index >= 0 {
		v50 := voltages[v50Index]
		if math.Abs(v50-(-40.0)) > 5.0 {
			t.Errorf("V1/2 activation (%.1f mV) differs from expected -40 mV", v50)
		}
		t.Logf("✓ V1/2 activation: ~%.1f mV (target: -40mV)", v50)
	} else {
		t.Error("Could not determine V1/2 for activation")
	}

	t.Log("✓ Nav1.6 activation curve matches experimental data")
}

// TestChannelKinetics_NavInactivationCurve validates voltage-dependent inactivation
//
// BIOLOGICAL BASIS:
// Nav1.6 inactivation prevents sustained sodium influx and enables repolarization.
// Inactivation is essential for action potential termination and refractoriness.
//
// EXPECTED BEHAVIOR:
// - V1/2 (half-inactivation voltage): ~-60 mV
// - Opposite slope to activation (decreases with depolarization)
// - Range: Near 100% at -100mV, near 0% at -20mV
// - Critical for spike shape and neuronal excitability
func TestChannelKinetics_NavInactivationCurve(t *testing.T) {
	t.Log("=== TESTING Nav1.6 Inactivation Curve ===")
	t.Log("Biological target: V1/2 = -60mV, availability decreases with depolarization")

	// Test inactivation curve (h_∞) across extended voltage range
	voltages := []float64{-100, -90, -80, -70, -60, -50, -40, -30, -20}
	inactivationProbs := make([]float64, len(voltages))

	for i, voltage := range voltages {
		channel := NewRealisticNavChannel("nav_inact_test")

		// Apply voltage for long time to reach steady state (200ms >> τ_h = 10ms)
		for j := 0; j < 200; j++ {
			channel.ShouldOpen(voltage, 0, 0, 1*time.Millisecond)
		}

		// Measure steady-state inactivation gate (h_∞)
		inactivationProbs[i] = channel.inactivationGate
		t.Logf("Voltage: %.1f mV, h_∞: %.3f", voltage, inactivationProbs[i])
	}

	// VALIDATION 1: Hyperpolarized behavior
	// At -100mV, should be fully available (> 95% non-inactivated)
	if inactivationProbs[0] < 0.95 {
		t.Errorf("Inactivation availability too low at -100 mV: %.3f (expected > 0.95)", inactivationProbs[0])
	}

	// VALIDATION 2: Depolarized behavior
	// At -20mV, should be heavily inactivated (< 10% available)
	if inactivationProbs[len(inactivationProbs)-1] > 0.1 {
		t.Errorf("Inactivation availability too high at -20 mV: %.3f (expected < 0.1)", inactivationProbs[len(inactivationProbs)-1])
	}

	// VALIDATION 3: Monotonic decrease (fundamental property of inactivation)
	for i := 1; i < len(inactivationProbs); i++ {
		if inactivationProbs[i] > inactivationProbs[i-1]+0.001 {
			t.Errorf("Inactivation curve not monotonically decreasing at %.1f mV", voltages[i])
		}
	}

	t.Log("✓ Nav1.6 inactivation curve matches experimental data")
	t.Log("✓ Proper voltage dependence for spike termination and refractoriness")
}

// TestChannelKinetics_NavActivationTimeConstant validates activation kinetics
//
// BIOLOGICAL BASIS:
// Fast Nav1.6 activation (τ ~1ms) enables rapid action potential upstroke.
// This speed is critical for high-frequency firing and signal propagation.
//
// EXPECTED BEHAVIOR:
// - Time constant (τ_m): ~1-3 ms at physiological voltages
// - Exponential approach to steady state
// - Fast enough for action potential initiation
//
// CLINICAL RELEVANCE:
// Slowed activation kinetics can cause conduction defects and arrhythmias.
func TestChannelKinetics_NavActivationTimeConstant(t *testing.T) {
	t.Log("=== TESTING Nav1.6 Activation Time Constant ===")
	t.Log("Biological target: τ_m = 1-3ms for rapid spike initiation")

	channel := NewRealisticNavChannel("nav_tau_test")

	// Voltage step protocol: rest → depolarization
	restVoltage := DENDRITE_VOLTAGE_RESTING_CORTICAL // -70mV
	testVoltage := -30.0                             // Strong depolarization

	// Equilibrate at rest (ensure starting from steady state)
	for i := 0; i < 50; i++ {
		channel.ShouldOpen(restVoltage, 0, 0, 1*time.Millisecond)
	}

	// Record activation over time after voltage step
	var timePoints []time.Duration
	for i := 0; i <= 10; i++ {
		timePoints = append(timePoints, time.Duration(i)*time.Millisecond)
	}

	activations := make([]float64, len(timePoints))
	activations[0] = channel.activationGate

	// Apply voltage step and measure time course
	for i := 1; i < len(timePoints); i++ {
		channel.ShouldOpen(testVoltage, 0, 0, 1*time.Millisecond)
		activations[i] = channel.activationGate
		t.Logf("Time: %v, Activation: %.3f", timePoints[i], activations[i])
	}

	// VALIDATION 1: Direction of change
	// Activation should increase with depolarization
	if activations[1] < activations[0] {
		t.Error("Activation should increase over time during depolarization")
	}

	// VALIDATION 2: Speed of activation
	// Should reach significant activation within ~10ms (>> τ_m)
	if activations[len(activations)-1] < 0.8 {
		t.Error("Activation too slow - should reach high levels in ~10ms")
	}

	// VALIDATION 3: Time constant estimation
	// Time to reach ~63% of maximum change
	maxActivation := activations[len(activations)-1]
	target := maxActivation * 0.63

	tauIndex := -1
	for i, activation := range activations {
		if activation >= target {
			tauIndex = i
			break
		}
	}

	if tauIndex >= 0 {
		estimatedTau := timePoints[tauIndex]
		if estimatedTau > 5*time.Millisecond {
			t.Errorf("Activation τ too slow: %v (should be ~1-3ms)", estimatedTau)
		}
		t.Logf("✓ Estimated activation τ: ~%v (target: 1-3ms)", estimatedTau)
	}

	t.Log("✓ Nav1.6 activation kinetics enable rapid spike initiation")
}

// TestChannelKinetics_NavInactivationDynamics validates inactivation kinetics
//
// BIOLOGICAL BASIS:
// Nav1.6 inactivation (τ ~10ms) terminates action potentials and creates
// refractory periods essential for directional propagation and firing patterns.
//
// EXPECTED BEHAVIOR:
// - Slower than activation (τ_h > τ_m)
// - Progressive reduction in availability during sustained depolarization
// - Creates absolute and relative refractory periods
//
// PATHOPHYSIOLOGY:
// Impaired inactivation causes persistent sodium currents linked to epilepsy.
func TestChannelKinetics_NavInactivationDynamics(t *testing.T) {
	t.Log("=== TESTING Nav1.6 Inactivation Dynamics ===")
	t.Log("Biological target: τ_h = 5-15ms, slower than activation")

	testVoltage := 0.0                               // Strong depolarization
	restVoltage := DENDRITE_VOLTAGE_RESTING_CORTICAL // -70mV

	// 1. Establish baseline inactivation at rest
	initialChannel := NewRealisticNavChannel("nav_inact_initial")
	for i := 0; i < 200; i++ { // Equilibrate at rest
		initialChannel.ShouldOpen(restVoltage, 0, 0, 1*time.Millisecond)
	}
	initialH := initialChannel.inactivationGate
	t.Logf("Initial h-gate at rest (%.1f mV): %.3f", restVoltage, initialH)

	// 2. Test inactivation development over time
	timePoints := []time.Duration{5, 10, 20, 50, 100} // ms
	inactivations := make([]float64, len(timePoints))

	for i, duration := range timePoints {
		// Fresh channel for each duration test
		testChannel := NewRealisticNavChannel("nav_inact_dynamics_test")

		// Pre-condition at rest (same as baseline)
		for j := 0; j < 200; j++ {
			testChannel.ShouldOpen(restVoltage, 0, 0, 1*time.Millisecond)
		}

		// Apply depolarization for specified duration
		for step := time.Duration(0); step < duration; step += 1 * time.Millisecond {
			testChannel.ShouldOpen(testVoltage, 0, 0, 1*time.Millisecond)
		}
		inactivations[i] = testChannel.inactivationGate
		t.Logf("Time at %.1f mV: %v, h-gate availability: %.3f",
			testVoltage, duration*time.Millisecond, inactivations[i])
	}

	// VALIDATION 1: Baseline availability
	// Should start with high availability at rest
	if initialH < 0.8 {
		t.Errorf("Initial h-gate availability should be high at rest, got %.3f", initialH)
	}

	// VALIDATION 2: Inactivation development
	// Should decrease from initial state during depolarization
	if inactivations[0] > initialH {
		t.Errorf("Inactivation should develop during depolarization: %.3f → %.3f",
			initialH, inactivations[0])
	}

	// VALIDATION 3: Progressive inactivation
	// Should continue to decrease over time (monotonic)
	for i := 1; i < len(inactivations); i++ {
		if inactivations[i] > inactivations[i-1]+0.001 {
			t.Errorf("Inactivation should progress over time: %.3f → %.3f",
				inactivations[i-1], inactivations[i])
		}
	}

	t.Log("✓ Nav1.6 inactivation dynamics create proper refractory periods")
	t.Log("✓ Progressive inactivation enables spike termination")
}

// ============================================================================
// CALCIUM CHANNEL (Cav1.2) KINETICS TESTS
// ============================================================================

// TestChannelKinetics_CaVoltageThreshold validates Ca2+ channel high threshold
//
// BIOLOGICAL BASIS:
// Cav1.2 (L-type) channels have high voltage thresholds (~-20mV) that prevent
// calcium influx at rest but allow robust calcium signaling during action
// potentials. This creates voltage-dependent calcium signaling essential for
// synaptic plasticity and gene expression.
//
// EXPECTED BEHAVIOR:
// - Higher threshold than Nav/Kv channels (V1/2 ~-20mV vs -40mV)
// - Minimal activation below -30mV
// - Strong activation above -10mV
// - Critical for LTP/LTD and calcium-dependent processes
//
// CLINICAL RELEVANCE:
// L-type channel dysfunction affects memory formation and cardiac function.
func TestChannelKinetics_CaVoltageThreshold(t *testing.T) {
	t.Log("=== TESTING Cav1.2 High Voltage Threshold ===")
	t.Log("Biological target: V1/2 = -20mV, higher threshold than Nav/Kv channels")

	// Test activation across voltage range spanning threshold
	voltages := []float64{-60, -50, -40, -30, -20, -10, 0, 10}
	caActivations := make([]float64, len(voltages))

	for i, voltage := range voltages {
		testChannel := NewRealisticCavChannel("cav_test")

		// Apply voltage to steady state (50ms >> τ_m = 3ms)
		for j := 0; j < 50; j++ {
			testChannel.ShouldOpen(voltage, 0, 0, 1*time.Millisecond)
		}

		_, _, prob := testChannel.ShouldOpen(voltage, 0, 0, 1*time.Millisecond)
		caActivations[i] = prob
		t.Logf("Ca²⁺ Voltage: %.1f mV, Activation: %.3f", voltage, prob)
	}

	// VALIDATION 1: Low voltage behavior (subthreshold)
	// Little activation below -30mV (< 20%)
	lowVoltageActivation := caActivations[2] // -40mV
	if lowVoltageActivation > 0.2 {
		t.Errorf("Ca²⁺ activation too high at -40 mV: %.3f (expected < 0.2)", lowVoltageActivation)
	}

	// VALIDATION 2: High voltage behavior (suprathreshold)
	// Good activation above -10mV (> 50%)
	highVoltageActivation := caActivations[len(caActivations)-2] // 0mV
	if highVoltageActivation < 0.5 {
		t.Errorf("Ca²⁺ activation too low at 0 mV: %.3f (expected > 0.5)", highVoltageActivation)
	}

	// VALIDATION 3: Threshold determination
	// Find voltage where activation reaches 10%
	thresholdVoltage := -60.0
	for i, activation := range caActivations {
		if activation >= 0.1 {
			thresholdVoltage = voltages[i]
			break
		}
	}

	if thresholdVoltage > -25.0 {
		t.Errorf("Ca²⁺ threshold too high: %.1f mV (should be around -30 to -20 mV)", thresholdVoltage)
	}

	t.Logf("✓ Ca²⁺ activation threshold: ~%.1f mV (target: -30 to -20mV)", thresholdVoltage)
	t.Log("✓ High threshold prevents calcium influx at rest")
	t.Log("✓ Enables voltage-dependent calcium signaling during spikes")
}

// ============================================================================
// GABA-A CHANNEL KINETICS TESTS
// ============================================================================

// TestChannelKinetics_GabaADesensitizationKinetics validates GABA-A desensitization timing
//
// BIOLOGICAL BASIS:
// GABA-A receptor desensitization limits inhibitory responses during sustained
// GABA exposure. This process involves conformational changes that reduce
// receptor sensitivity while ligand remains bound.
//
// EXPECTED BEHAVIOR:
// - Fast activation (τ ~2ms) upon GABA binding
// - Slow desensitization (τ ~100ms) during sustained exposure
// - Peak response occurs early, then declines
// - Recovery requires GABA removal
//
// FUNCTIONAL SIGNIFICANCE:
// Desensitization prevents prolonged inhibition and enables dynamic inhibitory
// control. Impaired desensitization can cause excessive inhibition.
//
// PHARMACOLOGY:
// Benzodiazepines slow desensitization, enhancing inhibition therapeutically.
func TestChannelKinetics_GabaADesensitizationKinetics(t *testing.T) {
	t.Log("=== TESTING GABA-A Desensitization Kinetics ===")
	t.Log("Biological targets: fast activation (τ ~2ms), slow desensitization (τ ~100ms)")

	// High GABA concentration to drive desensitization
	highGABA := 30.0                             // μM (saturating concentration)
	voltage := DENDRITE_VOLTAGE_RESTING_CORTICAL // -70mV

	// Track response over time during sustained GABA exposure
	timePoints := []int{1, 10, 25, 50, 100, 200} // Number of 1ms updates
	responses := make([]float64, len(timePoints))

	for i, numUpdates := range timePoints {
		testChannel := NewRealisticGabaAChannel("gaba_test")

		// Apply GABA for specified duration
		for j := 0; j < numUpdates; j++ {
			testChannel.ShouldOpen(voltage, highGABA, 0, 1*time.Millisecond)
		}

		_, _, prob := testChannel.ShouldOpen(voltage, highGABA, 0, 1*time.Millisecond)
		responses[i] = prob
		t.Logf("Updates: %d (%.0fms), Response: %.3f", numUpdates, float64(numUpdates), prob)
	}

	// ANALYSIS 1: Find peak response timing
	peakResponse := 0.0
	peakIndex := 0
	for i, response := range responses {
		if response > peakResponse {
			peakResponse = response
			peakIndex = i
		}
	}

	// VALIDATION 1: Peak response strength
	// Should achieve robust activation (> 40%)
	if peakResponse < 0.4 {
		t.Errorf("Peak GABA response should be strong, got %.3f (expected > 0.4)", peakResponse)
	}

	// VALIDATION 2: Peak timing
	// Should occur early (within first 25ms) due to fast activation
	if peakIndex > 2 { // Within first 3 time points (1, 10, 25 updates)
		t.Errorf("GABA peak too late, occurred at time point %d (expected ≤ 2)", peakIndex)
	}

	// VALIDATION 3: Desensitization progression
	// Response should decrease after peaking (100ms < 25ms response)
	if responses[4] >= responses[2] { // 100 vs 25 updates
		t.Errorf("GABA response should desensitize over time: 25ms=%.3f, 100ms=%.3f",
			responses[2], responses[4])
	}

	// VALIDATION 4: Significant desensitization
	// Final response should be substantially reduced from peak
	finalResponse := responses[len(responses)-1]
	reductionThreshold := peakResponse * 0.6 // 60% of peak
	if finalResponse > reductionThreshold {
		t.Errorf("Insufficient GABA desensitization: final=%.3f vs threshold=%.3f",
			finalResponse, reductionThreshold)
	}

	// ANALYSIS 2: Half-desensitization time
	halfResponse := peakResponse * 0.5
	halfTime := -1
	for i, response := range responses {
		if response <= halfResponse {
			halfTime = timePoints[i]
			break
		}
	}

	if halfTime > 0 {
		t.Logf("✓ Half-desensitization time: ~%dms", halfTime)
	}

	// Summary metrics
	reductionPercent := (peakResponse - finalResponse) / peakResponse * 100
	t.Logf("✓ GABA-A desensitization: peak=%.3f (at %dms), final=%.3f, reduction=%.1f%%",
		peakResponse, timePoints[peakIndex], finalResponse, reductionPercent)

	t.Log("✓ Fast activation enables rapid inhibitory responses")
	t.Log("✓ Slow desensitization prevents excessive inhibition")
	t.Log("✓ Kinetics match experimental GABA-A receptor data")
}

// ============================================================================
// POTASSIUM CHANNEL (Kv4.2) KINETICS TESTS
// ============================================================================

// TestChannelKinetics_KvDelayedRectifier validates K+ channel delayed activation
//
// BIOLOGICAL BASIS:
// Kv4.2 channels exhibit delayed activation that is slower than Nav channels.
// This delay enables action potential generation by preventing premature
// repolarization, while still providing repolarization once activated.
//
// EXPECTED BEHAVIOR:
// - Slower activation than Nav channels (τ ~5ms vs 1ms)
// - Progressive increase in activation over time
// - Fourth-power gating kinetics (n⁴)
// - Critical for action potential shape and frequency adaptation
//
// CLINICAL RELEVANCE:
// Kv4.2 mutations affect cardiac repolarization and neuronal excitability.
func TestChannelKinetics_KvDelayedRectifier(t *testing.T) {
	t.Log("=== TESTING Kv4.2 Delayed Rectifier Kinetics ===")
	t.Log("Biological target: slower activation than Nav (τ ~5ms vs 1ms)")

	channel := NewRealisticKvChannel("kv_delayed_test")

	// Test voltage that activates both Nav and Kv channels
	testVoltage := -20.0 // Above threshold for both channel types

	// Test K+ channel activation timing (20ms = 4x activation time constant)
	kvActivations := make([]float64, 20)
	for i := 0; i < len(kvActivations); i++ {
		channel.ShouldOpen(testVoltage, 0, 0, 1*time.Millisecond)
		_, _, prob := channel.ShouldOpen(testVoltage, 0, 0, 1*time.Millisecond)
		kvActivations[i] = prob
	}

	// Test Na+ channel activation timing for comparison
	navChannel := NewRealisticNavChannel("nav_compare")
	navActivations := make([]float64, 20)
	for i := 0; i < len(navActivations); i++ {
		navChannel.ShouldOpen(testVoltage, 0, 0, 1*time.Millisecond)
		_, _, prob := navChannel.ShouldOpen(testVoltage, 0, 0, 1*time.Millisecond)
		navActivations[i] = prob
	}

	// VALIDATION 1: Relative speed comparison
	// K+ should activate slower than Na+ initially (delayed rectifier property)
	if kvActivations[1] > navActivations[1] {
		t.Errorf("K+ should activate slower than Na+ initially. At 2ms: Kv=%.3f, Nav=%.3f",
			kvActivations[1], navActivations[1])
	}

	// VALIDATION 2: Continued activation
	// K+ channels should show continued increase over time (delayed activation)
	kvLate := kvActivations[len(kvActivations)-1]
	kvEarly := kvActivations[5]
	if (kvLate - kvEarly) < 0.01 {
		t.Error("K+ channels should show delayed activation (continued increase over time)")
	}

	// Comparison metrics
	navPeak, _ := findPeak(navActivations)
	t.Logf("✓ Delayed rectification confirmed: Kv_late=%.3f, Nav_peak=%.3f", kvLate, navPeak)
	t.Logf("✓ Early activation ratio (Kv/Nav at 2ms): %.3f", kvActivations[1]/navActivations[1])

	t.Log("✓ Delayed activation prevents premature spike termination")
	t.Log("✓ Enables proper action potential generation and repolarization")
}

// ============================================================================
// ACTIVITY-DEPENDENT MODULATION TESTS
// ============================================================================

// TestChannelKinetics_UseDependentModulation validates activity-dependent changes
//
// BIOLOGICAL BASIS:
// Ion channels exhibit use-dependent changes that provide feedback regulation:
// - Nav channels: Use-dependent inactivation reduces availability with repeated firing
// - Cav channels: Calcium-dependent facilitation enhances calcium influx
//
// FUNCTIONAL SIGNIFICANCE:
// - Use-dependent inactivation: Prevents excessive firing, creates spike frequency adaptation
// - Calcium facilitation: Enhances synaptic strength with increased activity
//
// PATHOPHYSIOLOGY:
// Impaired use-dependent modulation contributes to epilepsy and synaptic dysfunction.
func TestChannelKinetics_UseDependentModulation(t *testing.T) {
	t.Log("=== TESTING Use-Dependent Channel Modulation ===")
	t.Log("Biological targets: Nav inactivation (adaptation), Ca facilitation (enhancement)")

	// TEST 1: Sodium channel use-dependent inactivation
	t.Log("\n--- Testing Nav1.6 Use-Dependent Inactivation ---")

	channel := NewRealisticNavChannel("nav_use_dependent")
	channel.updateGating(-30.0, 1*time.Millisecond) // Pre-activate channel
	initialState := channel.GetState()
	initialConductance := initialState.Conductance

	// Simulate repetitive spiking activity (10 spikes)
	for i := 0; i < 10; i++ {
		feedback := &ChannelFeedback{
			ContributedToFiring: true,
			CurrentContribution: 100.0, // pA
			Timestamp:           time.Now(),
		}
		channel.UpdateKinetics(feedback, 1*time.Millisecond, -30.0)
	}

	finalState := channel.GetState()
	finalConductance := finalState.Conductance

	// VALIDATION 1: Use-dependent reduction
	// Repeated activity should reduce channel availability
	if finalConductance >= initialConductance {
		t.Errorf("Expected use-dependent reduction: initial=%.3f, final=%.3f",
			initialConductance, finalConductance)
	}

	reductionPercent := (initialConductance - finalConductance) / initialConductance * 100

	// VALIDATION 2: Magnitude of reduction
	// Should show significant but not complete inactivation (> 5% reduction)
	if reductionPercent < 5.0 {
		t.Errorf("Use-dependent inactivation too weak: %.1f%% reduction (expected > 5%%)",
			reductionPercent)
	}

	t.Logf("✓ Use-dependent inactivation: %.1f%% reduction after 10 spikes", reductionPercent)
	t.Log("✓ Provides spike frequency adaptation and prevents excessive firing")

	// TEST 2: Calcium channel facilitation
	t.Log("\n--- Testing Cav1.2 Calcium-Dependent Facilitation ---")

	caChannel := NewRealisticCavChannel("cav_facilitation")
	initialCaConductance := caChannel.GetConductance()

	// Simulate repeated calcium influx events
	for i := 0; i < 5; i++ {
		feedback := &ChannelFeedback{
			CalciumInflux:       50.0, // Significant calcium influx (pA equivalent)
			ContributedToFiring: true,
			Timestamp:           time.Now(),
		}
		caChannel.UpdateKinetics(feedback, 1*time.Millisecond, -10.0)
	}

	finalCaConductance := caChannel.GetConductance()

	// VALIDATION 3: Facilitation
	// Repeated calcium influx should enhance channel conductance
	if finalCaConductance <= initialCaConductance {
		t.Error("Expected Ca²⁺ channel facilitation with repeated calcium influx")
	}

	facilitationPercent := (finalCaConductance - initialCaConductance) / initialCaConductance * 100

	t.Logf("✓ Ca²⁺ channel facilitation: %.2f%% increase", facilitationPercent)
	t.Log("✓ Enhances calcium signaling with increased activity")
	t.Log("✓ Contributes to synaptic plasticity and activity-dependent adaptation")
}

// ============================================================================
// MULTI-CHANNEL INTEGRATION TESTS
// ============================================================================

// TestChannelKinetics_MultiChannelTiming validates realistic channel timing interactions
//
// BIOLOGICAL BASIS:
// During action potentials, different channel types activate and inactivate
// with distinct timing that shapes the spike waveform:
// - Nav: Fast activation → spike initiation
// - Kv: Delayed activation → repolarization
// - Cav: High threshold → calcium influx during spike peak
//
// EXPECTED BEHAVIOR:
// - Nav channels peak early (fast activation)
// - Kv channels peak later (delayed activation)
// - Cav channels require higher voltages (high threshold)
// - Proper sequence enables normal action potential shape
//
// CLINICAL RELEVANCE:
// Altered channel timing contributes to epilepsy, cardiac arrhythmias, and
// neurodevelopmental disorders.
func TestChannelKinetics_MultiChannelTiming(t *testing.T) {
	t.Log("=== TESTING Multi-Channel Timing Interactions ===")
	t.Log("Biological target: Nav early, Kv delayed, Cav high-threshold")

	// Create ensemble of channels (typical dendritic complement)
	navChannel := NewRealisticNavChannel("nav_timing")
	kvChannel := NewRealisticKvChannel("kv_timing")
	cavChannel := NewRealisticCavChannel("cav_timing")

	// Simulate action potential-like voltage trajectory
	// Represents typical somatic action potential waveform
	voltageTrajectory := []float64{
		-70, -60, -50, -40, -30, -20, -10, 0, 10, // Depolarization phase
		0, -10, -20, -30, -40, -50, -60, -70, // Repolarization phase
	}

	navResponses := make([]float64, len(voltageTrajectory))
	kvResponses := make([]float64, len(voltageTrajectory))
	cavResponses := make([]float64, len(voltageTrajectory))

	// Test each channel's response to the voltage trajectory
	for i, voltage := range voltageTrajectory {
		_, _, navProb := navChannel.ShouldOpen(voltage, 0, 0, 1*time.Millisecond)
		_, _, kvProb := kvChannel.ShouldOpen(voltage, 0, 0, 1*time.Millisecond)
		_, _, cavProb := cavChannel.ShouldOpen(voltage, 0, 0, 1*time.Millisecond)

		navResponses[i] = navProb
		kvResponses[i] = kvProb
		cavResponses[i] = cavProb

		t.Logf("V=%.0f mV: Nav=%.3f, Kv=%.3f, Cav=%.3f", voltage, navProb, kvProb, cavProb)
	}

	// ANALYSIS: Find peak responses and timing
	navPeak, navPeakIndex := findPeak(navResponses)
	kvPeak, kvPeakIndex := findPeak(kvResponses)
	cavPeak, cavPeakIndex := findPeak(cavResponses)

	// VALIDATION 1: Sodium channel timing
	// Should peak early during depolarization phase (index ≤ 8)
	if navPeakIndex > 8 {
		t.Errorf("Na⁺ peak too late: index %d (expected ≤ 8)", navPeakIndex)
	}

	// VALIDATION 2: Potassium channel timing
	// Should peak later than sodium (delayed rectifier property)
	if kvPeakIndex <= navPeakIndex {
		t.Errorf("K⁺ should peak after Na⁺: Nav=%d, Kv=%d", navPeakIndex, kvPeakIndex)
	}

	// VALIDATION 3: Calcium channel voltage requirement
	// Should require higher voltages (peak voltage > -10mV)
	peakVoltage := voltageTrajectory[cavPeakIndex]
	if peakVoltage < -10.0 {
		t.Errorf("Ca²⁺ peak at too low voltage: %.1f mV (expected > -10mV)", peakVoltage)
	}

	// Summary of timing relationships
	t.Logf("✓ Channel timing: Nav peak=%.3f@%d, Kv peak=%.3f@%d, Cav peak=%.3f@%d",
		navPeak, navPeakIndex, kvPeak, kvPeakIndex, cavPeak, cavPeakIndex)

	t.Log("✓ Nav channels provide fast spike initiation")
	t.Log("✓ Kv channels provide delayed repolarization")
	t.Log("✓ Cav channels provide high-threshold calcium influx")
	t.Log("✓ Timing sequence enables proper action potential shape")
}

// ============================================================================
// STATE CONSISTENCY AND VALIDATION TESTS
// ============================================================================

// TestChannelKinetics_StateTransitionValidation validates proper state machine behavior
//
// BIOLOGICAL BASIS:
// Ion channels must maintain thermodynamic consistency and physical constraints
// across all voltage and time conditions. State variables must remain bounded
// and conductance must reflect actual channel availability.
//
// VALIDATION TARGETS:
// - Conductance ≥ 0 (physical constraint)
// - Conductance ≤ maximum (no amplification)
// - State consistency across voltage changes
// - Proper bounds on all gating variables
//
// ENGINEERING IMPORTANCE:
// Numerical stability and biological realism require robust state management.
func TestChannelKinetics_StateTransitionValidation(t *testing.T) {
	t.Log("=== TESTING Channel State Transition Validation ===")
	t.Log("Validating thermodynamic consistency and numerical stability")

	channel := NewRealisticNavChannel("nav_state_test")

	// Test state consistency during complex voltage protocol
	voltages := []float64{-70, -40, 0, -40, -70} // Rest → activate → inactivate → recover → rest

	for i, voltage := range voltages {
		// Apply voltage for several time steps (allow kinetics to evolve)
		for j := 0; j < 5; j++ {
			channel.ShouldOpen(voltage, 0, 0, 1*time.Millisecond)
		}

		state := channel.GetState()
		trigger := channel.GetTrigger()

		// VALIDATION 1: Physical constraints
		// Conductance must be non-negative (thermodynamic requirement)
		if state.Conductance < 0 {
			t.Errorf("Invalid conductance at step %d: %.3f (must be ≥ 0)", i, state.Conductance)
		}

		// VALIDATION 2: Maximum conductance constraint
		// Effective conductance cannot exceed maximum (no amplification)
		if state.Conductance > channel.GetConductance()*1.01 { // Allow for float tolerance
			t.Errorf("Conductance exceeds maximum at step %d: %.3f > %.3f",
				i, state.Conductance, channel.GetConductance())
		}

		// VALIDATION 3: Parameter consistency
		// Channel properties should remain constant across states
		if trigger.ActivationVoltage != -40.0 {
			t.Errorf("Activation voltage changed: %.1f (should be constant)", trigger.ActivationVoltage)
		}

		t.Logf("Step %d (%.0f mV): Open=%v, Conductance=%.3f pS",
			i, voltage, state.IsOpen, state.Conductance)
	}

	t.Log("✓ State transitions maintain thermodynamic consistency")
	t.Log("✓ All physical constraints satisfied")
	t.Log("✓ Numerical stability confirmed across voltage range")
}

// ============================================================================
// DETAILED INTERFACE VALIDATION TESTS
// ============================================================================

// TestIonChannel_GatingDetails covers finer points of the channel gating interface
//
// BIOLOGICAL BASIS:
// Ion channel interfaces must accurately represent biophysical properties
// including open durations, current modulation, and trigger characteristics.
// These details affect integration with dendritic computation models.
//
// VALIDATION AREAS:
// - Open duration reporting (affects temporal integration)
// - Current modulation during inactivation (affects signal processing)
// - Trigger parameter accuracy (affects channel recruitment)
func TestIonChannel_GatingDetails(t *testing.T) {
	t.Log("=== TESTING Gating Details (Duration, Blocking, Trigger) ===")
	t.Log("Validating interface accuracy for dendritic integration")

	channel := NewRealisticNavChannel("nav_gating_details")

	// Test 1: Open duration reporting
	t.Run("OpenDuration", func(t *testing.T) {
		voltage := -30.0 // Voltage that reliably opens channel
		_, duration, prob := channel.ShouldOpen(voltage, 0, 0, 1*time.Millisecond)

		// VALIDATION: Duration should be positive when open probability is significant
		if prob > 0.1 && duration <= 0 {
			t.Errorf("Expected positive open duration when prob=%.3f, got %v", prob, duration)
		}

		// VALIDATION: Duration should match expected time constant
		if duration != DENDRITE_TIME_CHANNEL_ACTIVATION {
			t.Errorf("Expected open duration = %v, got %v",
				DENDRITE_TIME_CHANNEL_ACTIVATION, duration)
		}

		t.Logf("✓ Open duration correctly reported: %v", duration)
	})

	// Test 2: Signal modulation during inactivation
	t.Run("SignalBlocking", func(t *testing.T) {
		// Force deep inactivation with prolonged strong depolarization
		voltage := 20.0 // Very strong depolarization
		for i := 0; i < 100; i++ {
			channel.ShouldOpen(voltage, 0, 0, 1*time.Millisecond)
		}

		state := channel.GetState()
		// VALIDATION: Channel should be heavily inactivated
		if state.Conductance > channel.GetConductance()*0.1 {
			t.Fatalf("Channel not properly inactivated: conductance=%.3f pS", state.Conductance)
		}

		// Test current modulation through inactivated channel
		msg := types.NeuralSignal{Value: 100.0} // Large test signal
		_, shouldContinue, channelCurrent := channel.ModulateCurrent(msg, voltage, 0)

		// VALIDATION: Inactivated channel should produce minimal current
		if channelCurrent > 0.1 { // Allow tiny residual current
			t.Errorf("Inactivated channel produced significant current: %.3f pA", channelCurrent)
		}

		if !shouldContinue {
			t.Logf("✓ Inactivated channel blocked signal transmission")
		} else {
			t.Logf("✓ Inactivated channel passed minimal current: %.3f pA", channelCurrent)
		}
	})

	// Test 3: Trigger parameter validation
	t.Run("TriggerValidation", func(t *testing.T) {
		trigger := channel.GetTrigger()

		// VALIDATION: Activation voltage should match experimental data
		if trigger.ActivationVoltage != -40.0 {
			t.Errorf("Expected Nav activation voltage = -40.0mV, got %.1f", trigger.ActivationVoltage)
		}

		// VALIDATION: Inactivation voltage should match experimental data
		if trigger.InactivationVoltage != -60.0 {
			t.Errorf("Expected Nav inactivation voltage = -60.0mV, got %.1f", trigger.InactivationVoltage)
		}

		// VALIDATION: Time constant should be consistent
		if trigger.ActivationTimeConstant != DENDRITE_TIME_CHANNEL_ACTIVATION {
			t.Errorf("Mismatched activation time constant in trigger")
		}

		t.Logf("✓ Trigger parameters: V_act=%.1fmV, V_inact=%.1fmV",
			trigger.ActivationVoltage, trigger.InactivationVoltage)
	})

	t.Log("✓ All interface details validated for accurate dendritic integration")
}

// TestIonChannel_StateIntrospection validates the completeness of state reporting
//
// BIOLOGICAL BASIS:
// Accurate state introspection is essential for monitoring channel behavior
// and debugging biophysical models. State reporting must reflect actual
// channel conditions across all operating regimes.
//
// VALIDATION TARGETS:
// - Resting state accuracy (baseline conditions)
// - Open state detection (active periods)
// - Inactivated state recognition (refractory periods)
// - Voltage tracking (membrane potential coupling)
func TestIonChannel_StateIntrospection(t *testing.T) {
	t.Log("=== TESTING Complete State Introspection ===")
	t.Log("Validating state reporting across all operating regimes")

	channel := NewRealisticNavChannel("nav_introspection")

	// STATE 1: Resting conditions
	t.Log("\n--- Testing Resting State ---")
	restVoltage := -70.0
	channel.updateGating(restVoltage, 100*time.Millisecond) // Equilibrate
	state1 := channel.GetState()

	// VALIDATION: Near-zero conductance at rest
	if state1.Conductance > 0.001 {
		t.Errorf("Channel should have near-zero conductance at rest, got %.3f", state1.Conductance)
	}

	// VALIDATION: Voltage tracking accuracy
	if state1.MembraneVoltage != restVoltage {
		t.Errorf("Expected voltage %.1f, reported %.1f", restVoltage, state1.MembraneVoltage)
	}

	t.Logf("✓ Resting state: Open=%v, Vm=%.1fmV, g=%.3fpS",
		state1.IsOpen, state1.MembraneVoltage, state1.Conductance)

	// STATE 2: Open/active conditions
	t.Log("\n--- Testing Open State ---")
	openVoltage := -30.0
	channel.updateGating(openVoltage, 2*time.Millisecond) // Brief activation
	state2 := channel.GetState()

	// VALIDATION: Conductance should increase with opening
	if state2.Conductance <= state1.Conductance {
		t.Error("Conductance should increase upon channel opening")
	}

	t.Logf("✓ Open state: Open=%v, Vm=%.1fmV, g=%.3fpS",
		state2.IsOpen, state2.MembraneVoltage, state2.Conductance)

	// STATE 3: Inactivated conditions
	t.Log("\n--- Testing Inactivated State ---")
	inactivatedVoltage := 0.0
	channel.updateGating(inactivatedVoltage, 100*time.Millisecond) // Long inactivation
	state3 := channel.GetState()

	// VALIDATION: Conductance should decrease with inactivation
	if state3.Conductance >= state2.Conductance {
		t.Error("Conductance should decrease upon channel inactivation")
	}

	t.Logf("✓ Inactivated state: Open=%v, Vm=%.1fmV, g=%.3fpS",
		state3.IsOpen, state3.MembraneVoltage, state3.Conductance)

	t.Log("✓ State introspection provides complete channel monitoring")
	t.Log("✓ All operating regimes accurately reported")
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// findPeak locates the maximum value and its index in a slice
// Used for analyzing channel response timing during voltage protocols
func findPeak(values []float64) (float64, int) {
	maxVal := -1.0
	maxIndex := 0

	for i, val := range values {
		if val > maxVal {
			maxVal = val
			maxIndex = i
		}
	}
	return maxVal, maxIndex
}

// ============================================================================
// FUTURE EXTENSION PLACEHOLDERS
// ============================================================================

// TestIonChannel_AdvancedModulation_Placeholders provides stubs for future tests
//
// BIOLOGICAL BASIS:
// Advanced modulation mechanisms represent important areas for future
// development including phosphorylation, additional channel subtypes,
// and calcium-activated channels.
//
// DEVELOPMENT ROADMAP:
// These placeholder tests outline implementation targets for enhanced
// biophysical accuracy and expanded channel repertoire.
func TestIonChannel_AdvancedModulation_Placeholders(t *testing.T) {
	t.Log("=== TESTING Advanced Modulation Mechanisms (Placeholders) ===")
	t.Log("Outlining future development targets for enhanced biophysical accuracy")

	// FUTURE TEST 1: Phosphorylation-dependent modulation
	t.Run("PhosphorylationDependentModulation", func(t *testing.T) {
		t.Skip("FUTURE: Phosphorylation modulation (PKA/PKC pathways)")
		// IMPLEMENTATION TARGET:
		// - Add phosphorylation state tracking to channels
		// - Implement PKA/PKC-dependent conductance changes
		// - Validate against experimental phosphorylation data
		// - Test activity-dependent phosphorylation dynamics
	})

	// FUTURE TEST 2: Calcium-activated potassium channels
	t.Run("CalciumActivatedKChannel", func(t *testing.T) {
		t.Skip("FUTURE: SK/BK calcium-activated K⁺ channels")
		// IMPLEMENTATION TARGET:
		// - Create SK (small conductance) and BK (big conductance) channels
		// - Implement calcium-dependent gating kinetics
		// - Add calcium buffering and diffusion effects
		// - Validate against experimental calcium sensitivity data
	})

	// FUTURE TEST 3: Additional calcium channel subtypes
	t.Run("CalciumChannelSubTypes", func(t *testing.T) {
		t.Skip("FUTURE: N-type, P/Q-type, T-type calcium channels")
		// IMPLEMENTATION TARGET:
		// - Implement Cav2.1 (P/Q-type), Cav2.2 (N-type), Cav3.x (T-type)
		// - Model distinct voltage dependencies and kinetics
		// - Add calcium-dependent inactivation variations
		// - Validate presynaptic vs postsynaptic channel distributions
	})

	// FUTURE TEST 4: Metabotropic modulation
	t.Run("MetabotropicModulation", func(t *testing.T) {
		t.Skip("FUTURE: G-protein coupled receptor modulation")
		// IMPLEMENTATION TARGET:
		// - Implement mGluR, muscarinic, adrenergic modulation
		// - Add second messenger pathway effects (cAMP, IP3, DAG)
		// - Model slow timescale modulation (seconds to minutes)
		// - Validate against experimental neuromodulation data
	})

	t.Log("✓ Future development roadmap established")
	t.Log("✓ Placeholders provide implementation guidance")
	t.Log("✓ Advanced features will enhance biological accuracy")
}
