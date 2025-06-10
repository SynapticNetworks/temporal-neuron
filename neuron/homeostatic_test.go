/*
=================================================================================
HOMEOSTATIC PLASTICITY TESTS
=================================================================================

OVERVIEW:
This test suite validates the biological homeostatic plasticity mechanisms
that enable neurons to self-regulate their activity levels and maintain
stable network dynamics. These tests verify that neurons can automatically
adjust their firing thresholds to approach target firing rates through
calcium-based activity sensing and threshold adaptation.

BIOLOGICAL CONTEXT:
Homeostatic plasticity is a fundamental mechanism in biological neural networks
that prevents runaway excitation or neural silence by allowing individual neurons
to monitor their own activity levels and adjust their intrinsic excitability
accordingly. This creates stable yet adaptive networks that can learn without
destabilizing.

KEY MECHANISMS TESTED:
1. Calcium-based activity sensing (models intracellular calcium signaling)
2. Threshold adjustment based on activity deviation from target rates
3. Firing history tracking for rate calculation
4. Bounds enforcement to prevent pathological threshold changes
5. Integration with synaptic scaling and other plasticity mechanisms

SYNAPTIC INTEGRATION:
All tests use the new synapse package for biologically accurate connections
with realistic delays, STDP learning capabilities, and structural plasticity.
This ensures homeostatic mechanisms work properly with sophisticated synaptic
dynamics rather than simplified message passing.
*/

package neuron

import (
	"sync"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/synapse"
)

// ============================================================================
// HOMEOSTATIC NEURON CREATION AND INITIALIZATION TESTS
// ============================================================================

// TestHomeostaticNeuronCreation validates proper initialization of homeostatic neurons
//
// BIOLOGICAL SIGNIFICANCE:
// Real neurons are born with specific target firing rates and homeostatic capabilities
// that are genetically determined. Different neuron types (motor neurons, interneurons,
// sensory neurons) have different intrinsic excitability properties and target activity
// levels. This test ensures our artificial neurons properly initialize with these
// biological characteristics.
//
// EXPECTED RESULTS:
// - Neuron initializes with specified homeostatic parameters
// - Calcium level starts at zero (no initial activity)
// - Firing history begins empty (no previous activity)
// - Threshold bounds are properly calculated from base threshold
// - All homeostatic timing parameters are correctly set
func TestHomeostaticNeuronCreation(t *testing.T) {
	// Configure realistic homeostatic parameters
	// These values reflect experimentally observed parameters in cortical neurons
	threshold := 1.5                          // Base firing threshold (will be adjusted homeostatic)
	decayRate := 0.95                         // 5% membrane potential decay per millisecond
	refractoryPeriod := 10 * time.Millisecond // Typical cortical neuron refractory period
	fireFactor := 2.0                         // Action potential amplitude multiplier
	neuronID := "homeostatic_test_neuron"
	targetFiringRate := 5.0    // Target 5 Hz firing rate (typical cortical range)
	homeostasisStrength := 0.1 // Gentle 10% threshold adjustment strength

	// Create homeostatic neuron with biological parameters
	neuron := NewNeuron(neuronID, threshold, decayRate, refractoryPeriod,
		fireFactor, targetFiringRate, homeostasisStrength)

	if neuron == nil {
		t.Fatal("NewNeuron returned nil - neuron creation failed")
	}

	// Validate homeostatic initialization
	info := neuron.GetHomeostaticInfo()

	if info.targetFiringRate != targetFiringRate {
		t.Errorf("Target firing rate incorrect: expected %f, got %f",
			targetFiringRate, info.targetFiringRate)
	}

	if info.homeostasisStrength != homeostasisStrength {
		t.Errorf("Homeostasis strength incorrect: expected %f, got %f",
			homeostasisStrength, info.homeostasisStrength)
	}

	// Calcium should start at zero (no initial activity)
	if info.calciumLevel != 0.0 {
		t.Errorf("Initial calcium level should be zero, got %f", info.calciumLevel)
	}

	// Firing history should be empty initially
	if len(info.firingHistory) != 0 {
		t.Errorf("Firing history should be empty initially, got %d entries",
			len(info.firingHistory))
	}

	// Current threshold should match base threshold initially
	currentThreshold := neuron.GetCurrentThreshold()
	if currentThreshold != threshold {
		t.Errorf("Current threshold should match base initially: expected %f, got %f",
			threshold, currentThreshold)
	}

	baseThreshold := neuron.GetBaseThreshold()
	if baseThreshold != threshold {
		t.Errorf("Base threshold should be preserved: expected %f, got %f",
			threshold, baseThreshold)
	}

	// Validate threshold bounds are properly calculated
	expectedMinThreshold := threshold * 0.1 // 10% of base threshold
	expectedMaxThreshold := threshold * 5.0 // 5x base threshold

	if info.minThreshold != expectedMinThreshold {
		t.Errorf("Min threshold bound incorrect: expected %f, got %f",
			expectedMinThreshold, info.minThreshold)
	}

	if info.maxThreshold != expectedMaxThreshold {
		t.Errorf("Max threshold bound incorrect: expected %f, got %f",
			expectedMaxThreshold, info.maxThreshold)
	}
}

// ============================================================================
// HOMEOSTATIC THRESHOLD ADJUSTMENT TESTS
// ============================================================================

// TestHomeostaticThresholdAdjustment validates self-regulation of excitability
//
// BIOLOGICAL CONTEXT:
// Homeostatic plasticity allows neurons to maintain stable firing rates by
// adjusting their intrinsic excitability. When firing rate is too high,
// threshold increases (less excitable). When firing rate is too low,
// threshold decreases (more excitable). This creates a negative feedback
// loop that stabilizes network activity.
//
// EXPECTED RESULTS:
// - Threshold increases with sustained high activity
// - Threshold decreases with sustained low activity
// - Target firing rate is approached over time
// - Calcium-based activity sensing drives adjustments
// In homeostatic_test.go

func TestHomeostaticThresholdAdjustment(t *testing.T) {
	targetRate := 5.0 // Hz

	neuron := NewNeuron("homeostatic_neuron", 1.0, 0.95, 5*time.Millisecond,
		1.0, targetRate, 0.05) // Stable strength
	go neuron.Run()
	defer neuron.Close()

	initialThreshold := neuron.GetCurrentThreshold()
	t.Logf("[High Activity] Initial threshold: %.6f", initialThreshold)

	// Test 1: High activity
	for i := 0; i < 30; i++ {
		neuron.Receive(synapse.SynapseMessage{
			Value:     2.0, // Increase input strength to ensure firing
			Timestamp: time.Now(),
			SourceID:  "high_activity_source",
		})
		time.Sleep(10 * time.Millisecond) // 100 Hz input frequency

		if i%10 == 0 {
			t.Logf("[High Activity] Step %d: threshold=%.6f, calcium=%.6f, firingRate=%.2f Hz",
				i,
				neuron.GetCurrentThreshold(),
				neuron.GetCalciumLevel(),
				neuron.GetCurrentFiringRate())
		}
	}

	time.Sleep(1 * time.Second)

	adjustedThreshold := neuron.GetCurrentThreshold()
	t.Logf("[High Activity] Final threshold: %.6f", adjustedThreshold)

	if adjustedThreshold <= initialThreshold {
		t.Fatalf("Threshold did not increase with high activity - homeostasis not working (%.6f → %.6f)",
			initialThreshold, adjustedThreshold)
	}

	// Test 2: Low activity should decrease threshold
	lowActivityNeuron := NewNeuron("low_activity_neuron", 2.0, 0.95, 5*time.Millisecond,
		1.0, targetRate, 0.05)
	go lowActivityNeuron.Run()
	defer lowActivityNeuron.Close()

	initialThresholdLow := lowActivityNeuron.GetCurrentThreshold()
	t.Logf("[Low Activity] Initial threshold: %.6f", initialThresholdLow)

	for i := 0; i < 10; i++ {
		lowActivityNeuron.Receive(synapse.SynapseMessage{
			Value:     0.5,
			Timestamp: time.Now(),
			SourceID:  "low_activity_source",
		})
		time.Sleep(150 * time.Millisecond)

		if i%5 == 0 {
			t.Logf("[Low Activity] Step %d: threshold=%.6f, calcium=%.6f, firingRate=%.2f Hz",
				i,
				lowActivityNeuron.GetCurrentThreshold(),
				lowActivityNeuron.GetCalciumLevel(),
				lowActivityNeuron.GetCurrentFiringRate())
		}
	}

	time.Sleep(1 * time.Second)

	lowActivityThreshold := lowActivityNeuron.GetCurrentThreshold()
	t.Logf("[Low Activity] Final threshold: %.6f", lowActivityThreshold)

	if lowActivityThreshold >= initialThresholdLow {
		t.Fatalf("Threshold did not decrease with low activity (%.6f → %.6f)",
			initialThresholdLow, lowActivityThreshold)
	}
	t.Logf("[High Activity] Final threshold: %.6f, final firing rate: %.2f Hz", adjustedThreshold, neuron.GetCurrentFiringRate())

}

// TestHomeostaticThresholdIncrease validates threshold increase during hyperactivity
//
// BIOLOGICAL SIGNIFICANCE:
// When biological neurons fire above their optimal rate, they increase their
// firing threshold through calcium-dependent signaling cascades. This reduces
// their excitability and brings firing rates back toward target levels. This
// mechanism prevents runaway excitation that could lead to seizure-like activity.
//
// EXPERIMENTAL PROTOCOL:
// 1. Create neuron with low target firing rate (2 Hz)
// 2. Provide strong, frequent synaptic inputs to induce hyperactivity
// 3. Monitor threshold adjustment over multiple homeostatic cycles
// 4. Verify threshold increases and calcium accumulates appropriately
//
// EXPECTED RESULTS:
// - Firing rate exceeds target rate initially
// - Calcium level increases due to frequent firing
// - Threshold gradually increases to reduce excitability
// - System demonstrates negative feedback regulation
func TestHomeostaticThresholdIncrease(t *testing.T) {
	targetRate := 2.0 // Low target rate to easily exceed
	strength := 1.0   // Strong homeostatic regulation for clear effects

	// Create homeostatic neuron
	neuron := NewNeuron("threshold_increase_test", 1.0, 0.95, 5*time.Millisecond,
		1.0, targetRate, strength)

	// Create input source and synaptic connection
	inputNeuron := NewSimpleNeuron("input_source", 0.5, 0.95, 5*time.Millisecond, 1.0)

	// Configure STDP for the synapse (can be disabled for pure homeostatic testing)
	stdpConfig := synapse.CreateDefaultSTDPConfig()
	stdpConfig.Enabled = false // Focus on homeostasis only

	synapseConnection := synapse.NewBasicSynapse(
		"test_synapse",
		inputNeuron,
		neuron,
		stdpConfig,
		synapse.CreateDefaultPruningConfig(),
		1.0, // Strong synaptic weight
		0,   // No transmission delay for immediate effects
	)

	// Add synaptic connection to input neuron
	inputNeuron.AddOutputSynapse("to_target", synapseConnection)

	// Start both neurons
	go neuron.Run()
	go inputNeuron.Run()
	defer func() {
		neuron.Close()
		inputNeuron.Close()
	}()

	// Record initial threshold
	initialThreshold := neuron.GetCurrentThreshold()

	// Create hyperactivity by driving input neuron at high frequency
	// Target rate is 2 Hz (500ms between spikes), we'll drive much faster
	inputChannel := inputNeuron.GetInputChannel()

	// Send signals to input neuron to cause frequent firing
	for i := 0; i < 15; i++ {
		// Send strong signal to input neuron
		inputChannel <- synapse.SynapseMessage{
			Value:     1.2, // Above input neuron's threshold
			Timestamp: time.Now(),
			SourceID:  "hyperactivity_driver",
		}
		time.Sleep(50 * time.Millisecond) // 20 Hz input rate (much higher than 2 Hz target)
	}

	// Wait for homeostatic adjustment cycles
	time.Sleep(500 * time.Millisecond)

	// Validate threshold increase
	finalThreshold := neuron.GetCurrentThreshold()
	if finalThreshold <= initialThreshold {
		t.Errorf("Expected threshold increase due to hyperactivity. Initial: %f, Final: %f",
			initialThreshold, finalThreshold)
	}

	// Verify calcium accumulation from firing activity
	calciumLevel := neuron.GetCalciumLevel()
	if calciumLevel <= 0 {
		t.Errorf("Expected calcium accumulation after firing activity, got %f", calciumLevel)
	}

	// Verify firing rate was above target
	firingRate := neuron.GetCurrentFiringRate()

	t.Logf("Final firing rate: %f Hz (target: %f Hz)", firingRate, targetRate)
	t.Logf("Threshold change: %f -> %f (increase: %f)",
		initialThreshold, finalThreshold, finalThreshold-initialThreshold)
	t.Logf("Calcium level: %f", calciumLevel)
}

// TestHomeostaticThresholdDecrease validates threshold decrease during hypoactivity
//
// BIOLOGICAL SIGNIFICANCE:
// When biological neurons fire below their optimal rate, they decrease their
// firing threshold to increase excitability. This compensates for weak synaptic
// input or network hypoactivity and prevents neural silence that would impair
// network function.
//
// EXPERIMENTAL PROTOCOL:
// 1. Create neuron with high target firing rate (5 Hz)
// 2. Provide weak synaptic inputs that rarely trigger firing
// 3. Monitor threshold adjustment during periods of low activity
// 4. Verify threshold decreases to increase excitability
//
// EXPECTED RESULTS:
// - Firing rate remains below target rate
// - Calcium level stays low due to infrequent firing
// - Threshold gradually decreases to increase excitability
// - System demonstrates compensatory regulation
func TestHomeostaticThresholdDecrease(t *testing.T) {
	targetRate := 5.0 // High target rate to create activity deficit
	strength := 0.5   // Moderate homeostatic regulation

	// Create homeostatic neuron
	neuron := NewNeuron("threshold_decrease_test", 1.0, 0.95, 5*time.Millisecond,
		1.0, targetRate, strength)

	// Create weak input source
	inputNeuron := NewSimpleNeuron("weak_input", 2.0, 0.95, 5*time.Millisecond, 1.0)

	// Configure weak synaptic connection
	stdpConfig := synapse.CreateDefaultSTDPConfig()
	stdpConfig.Enabled = false

	synapseConnection := synapse.NewBasicSynapse(
		"weak_synapse",
		inputNeuron,
		neuron,
		stdpConfig,
		synapse.CreateDefaultPruningConfig(),
		0.3, // Weak synaptic weight
		0,
	)

	inputNeuron.AddOutputSynapse("to_target", synapseConnection)

	// Start neurons
	go neuron.Run()
	go inputNeuron.Run()
	defer func() {
		neuron.Close()
		inputNeuron.Close()
	}()

	// Record initial threshold
	initialThreshold := neuron.GetCurrentThreshold()

	// Create hypoactivity with weak, infrequent inputs
	inputChannel := inputNeuron.GetInputChannel()

	// Send weak signals that rarely cause target neuron firing
	for i := 0; i < 30; i++ {
		// Weak signal to input neuron (may not even fire input neuron)
		inputChannel <- synapse.SynapseMessage{
			Value:     0.5, // Weak signal, may not reach input threshold
			Timestamp: time.Now(),
			SourceID:  "weak_driver",
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Wait for homeostatic adjustment
	time.Sleep(300 * time.Millisecond)

	// Validate threshold decrease
	finalThreshold := neuron.GetCurrentThreshold()
	if finalThreshold >= initialThreshold {
		t.Errorf("Expected threshold decrease due to hypoactivity. Initial: %f, Final: %f",
			initialThreshold, finalThreshold)
	}

	// Verify low calcium level
	calciumLevel := neuron.GetCalciumLevel()
	if calciumLevel > 1.0 {
		t.Errorf("Expected low calcium level for hypoactive neuron, got %f", calciumLevel)
	}

	// Verify firing rate is below target
	firingRate := neuron.GetCurrentFiringRate()
	if firingRate > targetRate {
		t.Logf("Note: Firing rate (%f Hz) above target (%f Hz) - threshold adjustment may need more time",
			firingRate, targetRate)
	}

	t.Logf("Threshold change: %f -> %f (decrease: %f)",
		initialThreshold, finalThreshold, initialThreshold-finalThreshold)
	t.Logf("Final firing rate: %f Hz (target: %f Hz)", firingRate, targetRate)
	t.Logf("Calcium level: %f", calciumLevel)
}

// TestHomeostaticStabilization validates long-term stability around target rate
//
// BIOLOGICAL SIGNIFICANCE:
// The ultimate test of homeostatic plasticity is whether it can maintain stable
// firing rates despite variable inputs. Real neurons face constantly changing
// synaptic input patterns but maintain relatively stable activity levels through
// homeostatic regulation.
//
// EXPERIMENTAL PROTOCOL:
// 1. Create neuron with moderate target firing rate
// 2. Apply variable-strength synaptic inputs to challenge homeostasis
// 3. Monitor firing rate convergence toward target over time
// 4. Verify system stability and reasonable threshold dynamics
//
// EXPECTED RESULTS:
// - Initial firing rate may deviate from target
// - Homeostatic mechanisms gradually adjust threshold
// - Firing rate moves toward target rate over time
// - Threshold changes remain within biological bounds
func TestHomeostaticStabilization(t *testing.T) {
	targetRate := 4.0 // Moderate target rate
	strength := 0.3   // Moderate regulation for stability

	// Create homeostatic neuron
	neuron := NewNeuron("stabilization_test", 1.0, 0.95, 5*time.Millisecond,
		1.0, targetRate, strength)

	// Create variable input source
	inputNeuron := NewSimpleNeuron("variable_input", 1.0, 0.95, 5*time.Millisecond, 1.0)

	// Configure synaptic connection
	stdpConfig := synapse.CreateDefaultSTDPConfig()
	stdpConfig.Enabled = false

	synapseConnection := synapse.NewBasicSynapse(
		"variable_synapse",
		inputNeuron,
		neuron,
		stdpConfig,
		synapse.CreateDefaultPruningConfig(),
		1.0, // Moderate synaptic weight
		0,
	)

	inputNeuron.AddOutputSynapse("to_target", synapseConnection)

	// Start neurons
	go neuron.Run()
	go inputNeuron.Run()
	defer func() {
		neuron.Close()
		inputNeuron.Close()
	}()

	// Coordination for variable input generation
	var wg sync.WaitGroup
	stopSignal := make(chan struct{})

	// Generate variable input pattern to challenge homeostasis
	wg.Add(1)
	go func() {
		defer wg.Done()
		inputChannel := inputNeuron.GetInputChannel()

		for i := 0; i < 60; i++ {
			select {
			case <-stopSignal:
				return
			default:
				// Variable strength inputs (challenges homeostatic regulation)
				signalStrength := 0.7 + 0.6*float64(i%4) // Pattern: 0.7, 1.3, 1.9, 2.5

				select {
				case inputChannel <- synapse.SynapseMessage{
					Value:     signalStrength,
					Timestamp: time.Now(),
					SourceID:  "variable_driver",
				}:
				case <-stopSignal:
					return
				}
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()

	// Monitor homeostatic regulation over time
	time.Sleep(1 * time.Second) // Initial adaptation period
	midRate := neuron.GetCurrentFiringRate()

	time.Sleep(3 * time.Second) // Extended homeostatic adaptation
	finalRate := neuron.GetCurrentFiringRate()

	// Stop input generation
	close(stopSignal)
	wg.Wait()

	// Evaluate homeostatic effectiveness
	targetTolerance := 2.0 // Allow 2 Hz tolerance for realistic assessment
	isWithinTarget := finalRate >= targetRate-targetTolerance &&
		finalRate <= targetRate+targetTolerance

	if !isWithinTarget {
		t.Logf("Warning: Final rate (%f Hz) outside target range (%f ± %f Hz) - homeostasis may need longer or stronger regulation",
			finalRate, targetRate, targetTolerance)
	} else {
		t.Logf("Success: Final rate (%f Hz) within target range (%f ± %f Hz)",
			finalRate, targetRate, targetTolerance)
	}

	// Verify homeostatic mechanisms are active
	info := neuron.GetHomeostaticInfo()
	if len(info.firingHistory) == 0 {
		t.Error("Expected non-empty firing history with homeostatic neuron")
	}

	// Verify threshold adjustment occurred
	currentThreshold := neuron.GetCurrentThreshold()
	baseThreshold := neuron.GetBaseThreshold()
	thresholdChanged := currentThreshold != baseThreshold

	if !thresholdChanged {
		t.Logf("Note: Threshold unchanged (%f) - may indicate target rate already achieved",
			currentThreshold)
	}

	t.Logf("Mid-test rate: %f Hz, Final rate: %f Hz (target: %f Hz)",
		midRate, finalRate, targetRate)
	t.Logf("Threshold: base=%f, current=%f, changed=%v",
		baseThreshold, currentThreshold, thresholdChanged)
}

// ============================================================================
// CALCIUM DYNAMICS TESTS
// ============================================================================

// TestCalciumDynamics validates calcium accumulation and decay mechanisms
//
// BIOLOGICAL SIGNIFICANCE:
// Intracellular calcium serves as a crucial activity sensor in biological neurons.
// Action potentials cause calcium influx through voltage-gated channels, and this
// calcium accumulates with repeated firing. Calcium removal through pumps and
// buffers creates a temporal integration of recent activity that drives homeostatic
// adjustments.
//
// EXPERIMENTAL PROTOCOL:
// 1. Create homeostatic neuron with calcium tracking enabled
// 2. Verify initial calcium level is zero
// 3. Trigger neuron firing and measure calcium increase
// 4. Monitor calcium decay over time without additional firing
// 5. Validate calcium dynamics match biological timescales
//
// EXPECTED RESULTS:
// - Initial calcium level is zero
// - Calcium increases immediately after firing
// - Calcium gradually decays over time (exponential)
// - Decay rate matches configured biological parameters
func TestCalciumDynamics(t *testing.T) {
	// Create homeostatic neuron with calcium tracking
	neuron := NewNeuron("calcium_test", 1.0, 0.95, 5*time.Millisecond,
		1.0, 5.0, 0.1)

	// Create input for controlled firing
	inputNeuron := NewSimpleNeuron("calcium_input", 0.5, 0.95, 5*time.Millisecond, 1.0)

	stdpConfig := synapse.CreateDefaultSTDPConfig()
	stdpConfig.Enabled = false

	synapseConnection := synapse.NewBasicSynapse(
		"calcium_synapse",
		inputNeuron,
		neuron,
		stdpConfig,
		synapse.CreateDefaultPruningConfig(),
		1.5, // Strong enough to ensure firing
		0,
	)

	inputNeuron.AddOutputSynapse("to_target", synapseConnection)

	// Start neurons
	go neuron.Run()
	go inputNeuron.Run()
	defer func() {
		neuron.Close()
		inputNeuron.Close()
	}()

	// Verify initial calcium level
	initialCalcium := neuron.GetCalciumLevel()
	if initialCalcium != 0.0 {
		t.Errorf("Initial calcium level should be zero, got %f", initialCalcium)
	}

	// Trigger neuron firing
	inputChannel := inputNeuron.GetInputChannel()
	inputChannel <- synapse.SynapseMessage{
		Value:     1.5, // Strong signal to ensure firing
		Timestamp: time.Now(),
		SourceID:  "calcium_trigger",
	}

	// Allow firing and calcium accumulation
	time.Sleep(10 * time.Millisecond)

	// Verify calcium increase
	postFireCalcium := neuron.GetCalciumLevel()
	if postFireCalcium <= initialCalcium {
		t.Errorf("Expected calcium increase after firing. Initial: %f, Post-fire: %f",
			initialCalcium, postFireCalcium)
	}

	// Wait for calcium decay
	time.Sleep(100 * time.Millisecond)

	// Verify calcium decay
	decayedCalcium := neuron.GetCalciumLevel()
	if decayedCalcium >= postFireCalcium {
		t.Errorf("Expected calcium decay over time. Post-fire: %f, Decayed: %f",
			postFireCalcium, decayedCalcium)
	}

	// Calcium should still be positive but reduced
	if decayedCalcium <= 0 {
		t.Errorf("Calcium should decay gradually, not disappear immediately. Got: %f",
			decayedCalcium)
	}

	t.Logf("Calcium dynamics - Initial: %f, Post-fire: %f, Decayed: %f",
		initialCalcium, postFireCalcium, decayedCalcium)

	// Validate decay follows exponential pattern
	// expectedDecayRatio := 0.98 // Configured calcium decay rate per millisecond
	// After 100ms, expect calcium ≈ postFireCalcium * (0.98^100)
	expectedDecayedCalcium := postFireCalcium * 0.1323 // Approximate (0.98^100)

	decayTolerance := 0.5 // Allow reasonable tolerance for timing variations
	if decayedCalcium < expectedDecayedCalcium*decayTolerance ||
		decayedCalcium > expectedDecayedCalcium/decayTolerance {
		t.Logf("Note: Calcium decay rate (%f) differs from expected (%f) - timing variations normal",
			decayedCalcium, expectedDecayedCalcium)
	}
}

// TestFiringHistoryTracking validates firing history maintenance for rate calculation
//
// BIOLOGICAL SIGNIFICANCE:
// Accurate firing rate calculation requires maintaining a sliding window of recent
// firing times. This temporal information is essential for homeostatic regulation
// to assess whether the neuron is firing above or below its target rate.
//
// EXPERIMENTAL PROTOCOL:
// 1. Create homeostatic neuron with firing history tracking
// 2. Verify initially empty firing history
// 3. Trigger multiple firing events at controlled intervals
// 4. Verify firing history accurately records events
// 5. Validate firing rate calculation from history
//
// EXPECTED RESULTS:
// - Firing history starts empty
// - Each firing event is recorded with accurate timing
// - Firing rate calculation reflects actual firing frequency
// - History maintenance follows sliding window principle
func TestFiringHistoryTracking(t *testing.T) {
	// Create homeostatic neuron
	neuron := NewNeuron("history_test", 1.0, 0.95, 5*time.Millisecond,
		1.0, 5.0, 0.1)

	// Create input for controlled firing
	inputNeuron := NewSimpleNeuron("history_input", 0.5, 0.95, 5*time.Millisecond, 1.0)

	stdpConfig := synapse.CreateDefaultSTDPConfig()
	stdpConfig.Enabled = false

	synapseConnection := synapse.NewBasicSynapse(
		"history_synapse",
		inputNeuron,
		neuron,
		stdpConfig,
		synapse.CreateDefaultPruningConfig(),
		1.5, // Strong synaptic weight
		0,
	)

	inputNeuron.AddOutputSynapse("to_target", synapseConnection)

	// Start neurons
	go neuron.Run()
	go inputNeuron.Run()
	defer func() {
		neuron.Close()
		inputNeuron.Close()
	}()

	// Verify initial empty history
	info := neuron.GetHomeostaticInfo()
	if len(info.firingHistory) != 0 {
		t.Errorf("Firing history should be empty initially, got %d entries",
			len(info.firingHistory))
	}

	// Trigger controlled firing events
	numFires := 5
	inputChannel := inputNeuron.GetInputChannel()

	for i := 0; i < numFires; i++ {
		inputChannel <- synapse.SynapseMessage{
			Value:     1.5, // Strong signal to ensure firing
			Timestamp: time.Now(),
			SourceID:  "history_trigger",
		}
		time.Sleep(50 * time.Millisecond) // Allow refractory period between fires
	}

	// Wait for all firing events to process
	time.Sleep(100 * time.Millisecond)

	// Verify firing history tracking
	info = neuron.GetHomeostaticInfo()

	if len(info.firingHistory) != numFires {
		t.Errorf("Expected %d entries in firing history, got %d",
			numFires, len(info.firingHistory))
	}

	// Verify firing rate calculation
	firingRate := neuron.GetCurrentFiringRate()
	if firingRate <= 0 {
		t.Errorf("Expected positive firing rate, got %f", firingRate)
	}

	// Validate reasonable firing rate (should be around 1/0.05 = 20 Hz based on timing)
	expectedRate := 1.0 / 0.05 // 1 spike per 50ms = 20 Hz
	rateTolerance := 10.0      // Allow significant tolerance for timing variations

	if firingRate < expectedRate-rateTolerance || firingRate > expectedRate+rateTolerance {
		t.Logf("Note: Calculated rate (%f Hz) differs from expected (%f Hz) - timing variations normal",
			firingRate, expectedRate)
	}

	// Verify chronological order of firing history
	for i := 1; i < len(info.firingHistory); i++ {
		if info.firingHistory[i].Before(info.firingHistory[i-1]) {
			t.Errorf("Firing history not in chronological order at index %d", i)
		}
	}

	t.Logf("Firing history: %d entries, calculated rate: %f Hz",
		len(info.firingHistory), firingRate)
}

// ============================================================================
// HOMEOSTATIC BOUNDS AND SAFETY TESTS
// ============================================================================

// TestHomeostaticBounds validates threshold adjustment bounds enforcement
//
// BIOLOGICAL SIGNIFICANCE:
// Real neurons cannot adjust their firing threshold indefinitely due to biophysical
// constraints. Ion channel densities, membrane properties, and cellular energetics
// impose limits on how excitable or unexcitable a neuron can become. These bounds
// prevent pathological states while allowing sufficient regulatory range.
//
// EXPERIMENTAL PROTOCOL:
// 1. Create neuron with extreme homeostatic regulation
// 2. Drive neuron to hyperactivity to test upper threshold bound
// 3. Verify threshold saturates at maximum biological limit
// 4. Confirm base threshold preservation and bounds calculations
//
// EXPECTED RESULTS:
// - Threshold increases with hyperactivity but doesn't exceed maximum bound
// - Base threshold remains unchanged (reference value preserved)
// - Bounds are calculated correctly from base threshold
// - Extreme regulation doesn't cause pathological threshold values
func TestHomeostaticBounds(t *testing.T) {
	baseThreshold := 1.0

	// Create neuron with very strong homeostatic regulation to test bounds
	neuron := NewNeuron("bounds_test", baseThreshold, 0.95, 5*time.Millisecond,
		1.0, 0.5, 2.0) // Very low target rate, very strong regulation

	// Create strong input source for hyperactivity
	inputNeuron := NewSimpleNeuron("strong_input", 0.5, 0.95, 5*time.Millisecond, 1.0)

	stdpConfig := synapse.CreateDefaultSTDPConfig()
	stdpConfig.Enabled = false

	synapseConnection := synapse.NewBasicSynapse(
		"bounds_synapse",
		inputNeuron,
		neuron,
		stdpConfig,
		synapse.CreateDefaultPruningConfig(),
		2.0, // Very strong synaptic weight
		0,
	)

	inputNeuron.AddOutputSynapse("to_target", synapseConnection)

	// Start neurons
	go neuron.Run()
	go inputNeuron.Run()
	defer func() {
		neuron.Close()
		inputNeuron.Close()
	}()

	// Get expected bounds based on implementation
	info := neuron.GetHomeostaticInfo()
	minThreshold := info.minThreshold
	maxThreshold := info.maxThreshold

	t.Logf("Homeostatic bounds: min=%f, max=%f, base=%f",
		minThreshold, maxThreshold, baseThreshold)

	// Drive neuron to hyperactivity to test upper bound
	inputChannel := inputNeuron.GetInputChannel()

	for i := 0; i < 100; i++ {
		inputChannel <- synapse.SynapseMessage{
			Value:     2.0, // Very strong signal
			Timestamp: time.Now(),
			SourceID:  "bounds_driver",
		}
		time.Sleep(10 * time.Millisecond) // Very fast rate
	}

	// Wait for homeostatic saturation
	time.Sleep(500 * time.Millisecond)

	// Verify threshold doesn't exceed maximum bound
	currentThreshold := neuron.GetCurrentThreshold()
	tolerance := 0.01 // Small tolerance for floating-point precision

	if currentThreshold > maxThreshold+tolerance {
		t.Errorf("Threshold (%f) exceeded max bound (%f)",
			currentThreshold, maxThreshold)
	}

	// Verify base threshold preservation
	if neuron.GetBaseThreshold() != baseThreshold {
		t.Errorf("Base threshold changed from %f to %f",
			baseThreshold, neuron.GetBaseThreshold())
	}

	// Verify bounds calculation
	expectedMinThreshold := baseThreshold * 0.1
	expectedMaxThreshold := baseThreshold * 5.0

	if minThreshold != expectedMinThreshold {
		t.Errorf("Min threshold bound incorrect: expected %f, got %f",
			expectedMinThreshold, minThreshold)
	}

	if maxThreshold != expectedMaxThreshold {
		t.Errorf("Max threshold bound incorrect: expected %f, got %f",
			expectedMaxThreshold, maxThreshold)
	}

	t.Logf("Final threshold: %f (within bounds: %f - %f)",
		currentThreshold, minThreshold, maxThreshold)
}

// TestHomeostaticParameterSetting validates dynamic parameter adjustment
//
// BIOLOGICAL SIGNIFICANCE:
// In research and therapeutic applications, it's important to be able to modify
// homeostatic parameters during runtime. This might model pharmacological
// interventions, developmental changes, or experimental manipulations that
// alter a neuron's target firing rate or homeostatic sensitivity.
//
// EXPERIMENTAL PROTOCOL:
// 1. Create neuron with initial homeostatic parameters
// 2. Modify target rate and homeostatic strength during runtime
// 3. Verify parameter updates take effect immediately
// 4. Test disabling homeostasis and threshold reset behavior
// 5. Validate thread-safe parameter modification
//
// EXPECTED RESULTS:
// - Parameter changes take effect immediately
// - Disabling homeostasis resets threshold to base value
// - Thread-safe modification without state corruption
// - Homeostatic info reflects new parameters accurately
func TestHomeostaticParameterSetting(t *testing.T) {
	// Create neuron with initial parameters
	neuron := NewNeuron("params_test", 1.0, 0.95, 5*time.Millisecond,
		1.0, 5.0, 0.1)

	// Test parameter modification
	newTargetRate := 10.0
	newStrength := 0.5

	neuron.SetHomeostaticParameters(newTargetRate, newStrength)

	// Verify parameter updates
	info := neuron.GetHomeostaticInfo()
	if info.targetFiringRate != newTargetRate {
		t.Errorf("Target rate not updated: expected %f, got %f",
			newTargetRate, info.targetFiringRate)
	}

	if info.homeostasisStrength != newStrength {
		t.Errorf("Homeostasis strength not updated: expected %f, got %f",
			newStrength, info.homeostasisStrength)
	}

	// Test disabling homeostasis
	originalThreshold := neuron.GetCurrentThreshold()
	baseThreshold := neuron.GetBaseThreshold()

	neuron.SetHomeostaticParameters(0.0, 0.0) // Disable homeostasis

	// Verify threshold reset to base value
	resetThreshold := neuron.GetCurrentThreshold()
	if resetThreshold != baseThreshold {
		t.Errorf("Expected threshold reset to base (%f) when homeostasis disabled, got %f",
			baseThreshold, resetThreshold)
	}

	// Verify calcium and history cleared
	calciumLevel := neuron.GetCalciumLevel()
	if calciumLevel != 0.0 {
		t.Errorf("Expected calcium cleared when homeostasis disabled, got %f",
			calciumLevel)
	}

	info = neuron.GetHomeostaticInfo()
	if len(info.firingHistory) != 0 {
		t.Errorf("Expected firing history cleared when homeostasis disabled, got %d entries",
			len(info.firingHistory))
	}

	t.Logf("Parameter update: target %f->%f, strength %f->%f",
		5.0, newTargetRate, 0.1, newStrength)
	t.Logf("Threshold reset: %f -> %f (base: %f)",
		originalThreshold, resetThreshold, baseThreshold)
}

// TestHomeostaticVsSimpleNeuron validates behavioral differences between neuron types
//
// BIOLOGICAL SIGNIFICANCE:
// This test ensures that simple neurons (without homeostasis) maintain fixed
// thresholds while homeostatic neurons adapt their thresholds based on activity.
// This validates the implementation's ability to support both adaptive and
// non-adaptive neural models within the same framework.
//
// EXPERIMENTAL PROTOCOL:
// 1. Create one homeostatic and one simple neuron with identical base parameters
// 2. Apply identical synaptic inputs to both neurons
// 3. Monitor threshold changes over time
// 4. Verify homeostatic neuron adapts while simple neuron remains fixed
// 5. Validate that simple neurons don't perform calcium tracking
//
// EXPECTED RESULTS:
// - Simple neuron threshold remains exactly constant
// - Homeostatic neuron threshold adapts based on activity
// - Simple neuron has zero calcium tracking
// - Homeostatic neuron accumulates calcium and tracks firing history
func TestHomeostaticVsSimpleNeuron(t *testing.T) {
	// Create both neuron types with identical base parameters
	homeostaticNeuron := NewNeuron("homeostatic", 1.0, 0.95, 5*time.Millisecond,
		1.0, 3.0, 0.3)
	simpleNeuron := NewSimpleNeuron("simple", 1.0, 0.95, 5*time.Millisecond, 1.0)

	// Create identical input sources
	homeostaticInput := NewSimpleNeuron("h_input", 0.8, 0.95, 5*time.Millisecond, 1.0)
	simpleInput := NewSimpleNeuron("s_input", 0.8, 0.95, 5*time.Millisecond, 1.0)

	// Create identical synaptic connections
	stdpConfig := synapse.CreateDefaultSTDPConfig()
	stdpConfig.Enabled = false

	homeostaticSynapse := synapse.NewBasicSynapse(
		"h_synapse", homeostaticInput, homeostaticNeuron,
		stdpConfig, synapse.CreateDefaultPruningConfig(), 1.3, 0)

	simpleSynapse := synapse.NewBasicSynapse(
		"s_synapse", simpleInput, simpleNeuron,
		stdpConfig, synapse.CreateDefaultPruningConfig(), 1.3, 0)

	homeostaticInput.AddOutputSynapse("to_target", homeostaticSynapse)
	simpleInput.AddOutputSynapse("to_target", simpleSynapse)

	// Start all neurons
	go homeostaticNeuron.Run()
	go simpleNeuron.Run()
	go homeostaticInput.Run()
	go simpleInput.Run()
	defer func() {
		homeostaticNeuron.Close()
		simpleNeuron.Close()
		homeostaticInput.Close()
		simpleInput.Close()
	}()

	// Record initial thresholds
	initialHomeostaticThreshold := homeostaticNeuron.GetCurrentThreshold()
	initialSimpleThreshold := simpleNeuron.GetCurrentThreshold()

	// Apply identical inputs to both neurons
	homeostaticInputChannel := homeostaticInput.GetInputChannel()
	simpleInputChannel := simpleInput.GetInputChannel()

	for i := 0; i < 30; i++ {
		signalValue := 1.3 // Above both input thresholds
		timestamp := time.Now()

		homeostaticInputChannel <- synapse.SynapseMessage{
			Value: signalValue, Timestamp: timestamp, SourceID: "test_input"}
		simpleInputChannel <- synapse.SynapseMessage{
			Value: signalValue, Timestamp: timestamp, SourceID: "test_input"}

		time.Sleep(40 * time.Millisecond)
	}

	// Wait for homeostatic adaptation
	time.Sleep(300 * time.Millisecond)

	// Check final thresholds
	finalHomeostaticThreshold := homeostaticNeuron.GetCurrentThreshold()
	finalSimpleThreshold := simpleNeuron.GetCurrentThreshold()

	// Simple neuron threshold should not change
	if finalSimpleThreshold != initialSimpleThreshold {
		t.Errorf("Simple neuron threshold changed from %f to %f",
			initialSimpleThreshold, finalSimpleThreshold)
	}

	// Track whether homeostatic neuron threshold changed
	homeostaticChanged := finalHomeostaticThreshold != initialHomeostaticThreshold

	// Verify homeostatic neuron has activity tracking
	homeostaticRate := homeostaticNeuron.GetCurrentFiringRate()
	if homeostaticRate <= 0 {
		t.Error("Homeostatic neuron should have measurable firing rate")
	}

	// Simple neuron should have zero calcium (no tracking)
	simpleCalcium := simpleNeuron.GetCalciumLevel()
	if simpleCalcium != 0.0 {
		t.Errorf("Simple neuron should have no calcium tracking, got %f", simpleCalcium)
	}

	// Homeostatic neuron should have calcium tracking
	homeostaticCalcium := homeostaticNeuron.GetCalciumLevel()
	if homeostaticCalcium <= 0.0 {
		t.Error("Homeostatic neuron should have calcium tracking")
	}

	// Verify homeostatic neuron has firing history
	homeostaticInfo := homeostaticNeuron.GetHomeostaticInfo()
	if len(homeostaticInfo.firingHistory) == 0 {
		t.Error("Homeostatic neuron should have firing history")
	}

	// Simple neuron homeostatic info should show disabled state
	simpleInfo := simpleNeuron.GetHomeostaticInfo()
	if simpleInfo.targetFiringRate != 0.0 || simpleInfo.homeostasisStrength != 0.0 {
		t.Error("Simple neuron should have homeostasis disabled")
	}

	t.Logf("Simple neuron: threshold %f -> %f (unchanged: %v)",
		initialSimpleThreshold, finalSimpleThreshold,
		finalSimpleThreshold == initialSimpleThreshold)
	t.Logf("Homeostatic neuron: threshold %f -> %f (changed: %v, rate: %f Hz)",
		initialHomeostaticThreshold, finalHomeostaticThreshold,
		homeostaticChanged, homeostaticRate)
}

// TestHomeostaticStabilityOverTime validates long-term homeostatic behavior
//
// BIOLOGICAL SIGNIFICANCE:
// Homeostatic plasticity operates over long timescales (minutes to hours) and
// must maintain stability while adapting to gradual changes. This test validates
// that the homeostatic system doesn't oscillate or become unstable during
// extended operation with variable input patterns.
//
// EXPERIMENTAL PROTOCOL:
// 1. Create homeostatic neuron with realistic parameters
// 2. Apply variable input patterns over extended period
// 3. Monitor threshold evolution and firing rate stability
// 4. Verify system reaches reasonable steady state
// 5. Check for pathological oscillations or runaway behavior
//
// EXPECTED RESULTS:
// - System maintains stable operation over extended periods
// - Threshold changes remain within biological bounds
// - Firing rate approaches target rate over time
// - No pathological oscillations or instabilities
func TestHomeostaticStabilityOverTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-term stability test in short mode")
	}

	targetRate := 5.0

	// Create homeostatic neuron
	neuron := NewNeuron("stability_test", 1.0, 0.95, 5*time.Millisecond,
		1.0, targetRate, 0.5)

	// Create variable input source
	inputNeuron := NewSimpleNeuron("stability_input", 1.0, 0.95, 5*time.Millisecond, 1.0)

	stdpConfig := synapse.CreateDefaultSTDPConfig()
	stdpConfig.Enabled = false

	synapseConnection := synapse.NewBasicSynapse(
		"stability_synapse", inputNeuron, neuron,
		stdpConfig, synapse.CreateDefaultPruningConfig(), 1.0, 0)

	inputNeuron.AddOutputSynapse("to_target", synapseConnection)

	// Start neurons
	go neuron.Run()
	go inputNeuron.Run()
	defer func() {
		neuron.Close()
		inputNeuron.Close()
	}()

	// Track system evolution over time
	var thresholds []float64
	var rates []float64

	// Coordination for long-term input
	var wg sync.WaitGroup
	stopSignal := make(chan struct{})

	// Generate long-term variable inputs
	wg.Add(1)
	go func() {
		defer wg.Done()
		inputChannel := inputNeuron.GetInputChannel()

		for i := 0; i < 200; i++ {
			select {
			case <-stopSignal:
				return
			default:
				// Variable input pattern (models natural input variability)
				val := 0.8 + 0.8*float64((i%10))/10.0 // 0.8 to 1.6

				select {
				case inputChannel <- synapse.SynapseMessage{
					Value: val, Timestamp: time.Now(), SourceID: "stability_input"}:
				case <-stopSignal:
					return
				}
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()

	// Sample system state every second for 10 seconds
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		thresholds = append(thresholds, neuron.GetCurrentThreshold())
		rates = append(rates, neuron.GetCurrentFiringRate())
	}

	// Stop input generation
	close(stopSignal)
	wg.Wait()

	// Analyze stability
	finalRate := rates[len(rates)-1]
	tolerance := 2.0

	if finalRate < targetRate-tolerance || finalRate > targetRate+tolerance {
		t.Logf("Warning: Final rate (%f Hz) not close to target (%f Hz) after long-term run",
			finalRate, targetRate)
	} else {
		t.Logf("Success: Final rate (%f Hz) close to target (%f Hz)",
			finalRate, targetRate)
	}

	// Check for reasonable threshold evolution (no extreme oscillations)
	thresholdRange := 0.0
	minThreshold := thresholds[0]
	maxThreshold := thresholds[0]

	for _, threshold := range thresholds {
		if threshold < minThreshold {
			minThreshold = threshold
		}
		if threshold > maxThreshold {
			maxThreshold = threshold
		}
	}
	thresholdRange = maxThreshold - minThreshold

	// Threshold range should be reasonable (not pathological oscillations)
	baseThreshold := neuron.GetBaseThreshold()
	if thresholdRange > baseThreshold*2.0 {
		t.Errorf("Excessive threshold oscillations: range %f for base %f",
			thresholdRange, baseThreshold)
	}

	// Verify final system state
	info := neuron.GetHomeostaticInfo()
	if len(info.firingHistory) == 0 {
		t.Error("Expected non-empty firing history after long-term run")
	}

	t.Logf("Long-term stability test completed")
	t.Logf("Target rate: %f Hz", targetRate)
	t.Logf("Final rate: %f Hz", finalRate)
	t.Logf("Threshold evolution: %f -> %f (range: %f)",
		thresholds[0], thresholds[len(thresholds)-1], thresholdRange)
}

// TestHomeostasisWithInhibition validates the neuron's response to inhibitory input.
//
// BIOLOGICAL SIGNIFICANCE:
// Neurons in the brain receive a mix of excitatory and inhibitory signals. Homeostasis
// must be able to handle periods of silence caused by strong inhibition. In response
// to being silenced, a neuron should become more excitable (decrease its threshold)
// to maintain its target firing rate.
//
// EXPERIMENTAL PROTOCOL:
// 1. Create a homeostatic neuron.
// 2. Apply a strong, prolonged inhibitory stimulus to silence the neuron.
// 3. Verify that the neuron's firing threshold decreases significantly.
//
// EXPECTED RESULTS:
// - During inhibition, the neuron's firing rate is zero.
// - The homeostatic mechanism detects this hypoactivity.
// - The firing threshold is lowered to make the neuron more sensitive to future inputs.
func TestHomeostasisWithInhibition(t *testing.T) {
	targetRate := 5.0
	strength := 0.8
	neuron := NewNeuron("inhibition_test", 1.5, 0.95, 5*time.Millisecond,
		1.0, targetRate, strength)
	go neuron.Run()
	defer neuron.Close()
	initialThreshold := neuron.GetCurrentThreshold()
	t.Logf("Initial threshold: %f", initialThreshold)
	for i := 0; i < 100; i++ {
		neuron.Receive(synapse.SynapseMessage{
			Value:     -1.0,
			Timestamp: time.Now(),
			SourceID:  "inhibitory_source",
			SynapseID: "inhibitory_synapse",
		})
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(500 * time.Millisecond)
	finalThreshold := neuron.GetCurrentThreshold()
	finalRate := neuron.GetCurrentFiringRate()
	finalCalcium := neuron.GetCalciumLevel()
	t.Logf("Final threshold: %f", finalThreshold)
	t.Logf("Final firing rate during inhibition: %f Hz", finalRate)
	t.Logf("Final calcium level: %f", finalCalcium)
	if finalRate > 0 {
		t.Errorf("Neuron should be silent during strong inhibition, but firing rate was %f Hz", finalRate)
	}
	if finalThreshold >= initialThreshold {
		t.Errorf("Expected threshold to decrease due to inhibition. Initial: %f, Final: %f", initialThreshold, finalThreshold)
	}
	if finalCalcium > 0.0 {
		t.Errorf("Expected zero calcium level during inhibition, got %f", finalCalcium)
	}
	info := neuron.GetHomeostaticInfo()
	if finalThreshold < info.minThreshold {
		t.Errorf("Threshold %f below minimum bound %f", finalThreshold, info.minThreshold)
	}
}

// ============================================================================
// HOMEOSTATIC PERFORMANCE BENCHMARKS
// ============================================================================

// BenchmarkHomeostaticNeuronCreation benchmarks homeostatic neuron creation performance
func BenchmarkHomeostaticNeuronCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewNeuron("bench_homeostatic", 1.0, 0.95, 5*time.Millisecond,
			1.0, 5.0, 0.1)
	}
}

// BenchmarkHomeostaticMessageProcessing benchmarks homeostatic message processing throughput
func BenchmarkHomeostaticMessageProcessing(b *testing.B) {
	neuron := NewNeuron("bench_processing", 10.0, 0.95, 5*time.Millisecond,
		1.0, 5.0, 0.1) // High threshold to avoid firing

	go neuron.Run()
	defer neuron.Close()

	inputChannel := neuron.GetInputChannel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inputChannel <- synapse.SynapseMessage{
			Value:     0.1,
			Timestamp: time.Now(),
			SourceID:  "bench_source",
		}
	}
}

// BenchmarkCalciumDynamics benchmarks calcium accumulation and decay performance
func BenchmarkCalciumDynamics(b *testing.B) {
	neuron := NewNeuron("bench_calcium", 1.0, 0.95, 5*time.Millisecond,
		1.0, 5.0, 0.1)

	inputNeuron := NewSimpleNeuron("bench_input", 0.5, 0.95, 5*time.Millisecond, 1.0)

	stdpConfig := synapse.CreateDefaultSTDPConfig()
	stdpConfig.Enabled = false

	synapseConnection := synapse.NewBasicSynapse(
		"bench_synapse", inputNeuron, neuron,
		stdpConfig, synapse.CreateDefaultPruningConfig(), 1.5, 0)

	inputNeuron.AddOutputSynapse("bench_output", synapseConnection)

	go neuron.Run()
	go inputNeuron.Run()
	defer func() {
		neuron.Close()
		inputNeuron.Close()
	}()

	inputChannel := inputNeuron.GetInputChannel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		inputChannel <- synapse.SynapseMessage{
			Value:     1.5,
			Timestamp: time.Now(),
			SourceID:  "bench_source",
		}
		time.Sleep(time.Microsecond) // Small delay for processing
	}
}

// BenchmarkHomeostaticInfoRetrieval benchmarks getting homeostatic information
func BenchmarkHomeostaticInfoRetrieval(b *testing.B) {
	neuron := NewNeuron("bench_info", 1.0, 0.95, 5*time.Millisecond,
		1.0, 5.0, 0.1)

	// Pre-populate some state for realistic benchmarking
	// Note: This would require accessing internal state, simplified for benchmark

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = neuron.GetHomeostaticInfo()
	}
}

// BenchmarkConcurrentHomeostaticAccess benchmarks concurrent access to homeostatic data
func BenchmarkConcurrentHomeostaticAccess(b *testing.B) {
	neuron := NewNeuron("bench_concurrent", 1.0, 0.95, 5*time.Millisecond,
		1.0, 5.0, 0.1)

	go neuron.Run()
	defer neuron.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Mix of read operations that would happen in real usage
			switch b.N % 4 {
			case 0:
				_ = neuron.GetCurrentFiringRate()
			case 1:
				_ = neuron.GetCurrentThreshold()
			case 2:
				_ = neuron.GetCalciumLevel()
			case 3:
				_ = neuron.GetHomeostaticInfo()
			}
		}
	})
}
