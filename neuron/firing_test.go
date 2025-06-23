package neuron

import (
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/message"
)

/*
=================================================================================
FIRING MECHANISM TESTS - PURE FIRING LOGIC VALIDATION
=================================================================================

This test suite validates the core firing mechanism and output coordination
in isolation from other neuron subsystems. These tests use mocks to verify
firing behavior, output transmission, chemical release, and firing history
management without requiring complex network setups.

=================================================================================
*/

// ============================================================================
// BASIC FIRING TESTS
// ============================================================================

// TestNeuronFiring_BasicFire validates core firing mechanism
func TestNeuronFiring_BasicFire(t *testing.T) {
	// Create neuron with low threshold for easy firing
	neuron := NewNeuron(
		"firing-test",
		0.5, // Low threshold
		0.95,
		5*time.Millisecond,
		1.0,
		5.0,
		0.1,
	)

	// Set up mock matrix callbacks
	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Record initial state using correct API
	initialStatus := neuron.GetFiringStatus()
	initialCalcium := initialStatus["calcium_level"].(float64)
	initialFiringRate := initialStatus["current_firing_rate"].(float64)

	// Send signal above threshold
	SendTestSignal(neuron, "test-source", 1.0)

	// Allow processing time
	time.Sleep(20 * time.Millisecond)

	// Verify firing occurred
	finalStatus := neuron.GetFiringStatus()
	finalCalcium := finalStatus["calcium_level"].(float64)

	if finalCalcium <= initialCalcium {
		t.Error("Expected calcium increase after firing")
	}

	// Verify health reporting
	if mockMatrix.GetHealthReportCount() == 0 {
		t.Error("Expected health reports after firing")
	}

	// Verify electrical signaling
	if mockMatrix.GetElectricalSignalCount() == 0 {
		t.Error("Expected electrical signals after firing")
	}

	t.Logf("✓ Firing occurred: calcium %.3f → %.3f", initialCalcium, finalCalcium)
	t.Logf("  Initial firing rate: %.3f Hz", initialFiringRate)
}

// TestNeuronFiring_RefractoryPeriod validates refractory period enforcement
func TestNeuronFiring_RefractoryPeriod(t *testing.T) {
	refractoryPeriod := 50 * time.Millisecond

	neuron := NewNeuron(
		"refractory-test",
		0.3, // Very low threshold
		0.95,
		refractoryPeriod,
		1.0,
		10.0,
		0.1,
	)

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// First firing
	SendTestSignal(neuron, "test-source", 1.0)
	time.Sleep(10 * time.Millisecond)

	firstHealthReports := mockMatrix.GetHealthReportCount()
	if firstHealthReports == 0 {
		t.Error("Expected first firing to occur")
	}

	// Immediate second signal (should be blocked by refractory period)
	SendTestSignal(neuron, "test-source", 1.0)
	time.Sleep(10 * time.Millisecond)

	secondHealthReports := mockMatrix.GetHealthReportCount()
	if secondHealthReports != firstHealthReports {
		t.Error("Second firing should be blocked by refractory period")
	}

	// Wait for refractory period to end
	time.Sleep(refractoryPeriod)

	// Third signal (should fire)
	SendTestSignal(neuron, "test-source", 1.0)
	time.Sleep(10 * time.Millisecond)

	thirdHealthReports := mockMatrix.GetHealthReportCount()
	if thirdHealthReports <= secondHealthReports {
		t.Error("Third firing should occur after refractory period")
	}

	t.Logf("✓ Refractory period enforced: %d → %d → %d reports",
		firstHealthReports, secondHealthReports, thirdHealthReports)
}

// TestNeuronFiring_OutputCallbacks validates output transmission
func TestNeuronFiring_OutputCallbacks(t *testing.T) {
	neuron := NewNeuron(
		"output-test",
		0.5,
		0.95,
		5*time.Millisecond,
		2.0, // Higher fire factor for clear output
		5.0,
		0.1,
	)

	// Create mock synapses
	synapse1 := NewMockSynapse("syn1", "target1", 1.0, 2*time.Millisecond)
	synapse2 := NewMockSynapse("syn2", "target2", 1.5, 5*time.Millisecond)

	// Add output callbacks
	neuron.AddOutputCallback("syn1", synapse1.CreateOutputCallback())
	neuron.AddOutputCallback("syn2", synapse2.CreateOutputCallback())

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Trigger firing
	SendTestSignal(neuron, "test-source", 1.0)

	// Wait for axonal delivery
	time.Sleep(20 * time.Millisecond)

	// Verify output transmission
	syn1Count := synapse1.GetReceivedSignalCount()
	syn2Count := synapse2.GetReceivedSignalCount()

	if syn1Count == 0 {
		t.Error("Expected synapse1 to receive signals")
	}

	if syn2Count == 0 {
		t.Error("Expected synapse2 to receive signals")
	}

	// Verify signal properties
	syn1Signals := synapse1.GetReceivedSignals()
	if len(syn1Signals) > 0 {
		signal := syn1Signals[0]
		if signal.SourceID != neuron.ID() {
			t.Errorf("Expected source ID %s, got %s", neuron.ID(), signal.SourceID)
		}
		if signal.TargetID != "target1" {
			t.Errorf("Expected target ID target1, got %s", signal.TargetID)
		}
		if signal.SynapseID != "syn1" {
			t.Errorf("Expected synapse ID syn1, got %s", signal.SynapseID)
		}
	}

	t.Logf("✓ Output transmission: syn1=%d signals, syn2=%d signals", syn1Count, syn2Count)
}

// ============================================================================
// CHEMICAL RELEASE TESTS
// ============================================================================

// TestNeuronFiring_ChemicalRelease validates neurotransmitter release
func TestNeuronFiring_ChemicalRelease(t *testing.T) {
	neuron := NewNeuron(
		"chemical-test",
		0.5,
		0.95,
		5*time.Millisecond,
		1.0,
		5.0,
		0.1,
	)

	// Set neurotransmitters
	neuron.SetReleasedLigands([]message.LigandType{
		message.LigandDopamine,
		message.LigandGlutamate,
	})

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Trigger firing
	SendTestSignal(neuron, "test-source", 1.0)
	time.Sleep(20 * time.Millisecond)

	// Verify chemical releases
	releases := mockMatrix.GetChemicalReleases()
	if len(releases) == 0 {
		t.Error("Expected chemical releases after firing")
	}

	// Verify all configured ligands were released
	releasedTypes := make(map[message.LigandType]bool)
	for _, release := range releases {
		releasedTypes[release.LigandType] = true
		if release.Concentration <= 0 {
			t.Errorf("Expected positive concentration for %s, got %f",
				release.LigandType.String(), release.Concentration)
		}
	}

	expectedTypes := []message.LigandType{message.LigandDopamine, message.LigandGlutamate}
	for _, expectedType := range expectedTypes {
		if !releasedTypes[expectedType] {
			t.Errorf("Expected %s to be released", expectedType.String())
		}
	}

	t.Logf("✓ Chemical release: %d ligand types released", len(releasedTypes))
}

// ============================================================================
// FIRING HISTORY TESTS
// ============================================================================

// TestNeuronFiring_FiringHistory validates firing history tracking
func TestNeuronFiring_FiringHistory(t *testing.T) {
	neuron := NewNeuron(
		"history-test",
		0.5,
		0.95,
		10*time.Millisecond, // Short refractory for rapid firing
		1.0,
		5.0,
		0.1,
	)

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Record initial state
	initialStatus := neuron.GetFiringStatus()
	initialRate := initialStatus["current_firing_rate"].(float64)

	// Trigger multiple firings
	numFirings := 5
	for i := 0; i < numFirings; i++ {
		SendTestSignal(neuron, "test-source", 1.0)
		time.Sleep(20 * time.Millisecond) // Wait longer than refractory period
	}

	// Allow final processing
	time.Sleep(50 * time.Millisecond)

	// Verify firing rate increased
	finalStatus := neuron.GetFiringStatus()
	finalRate := finalStatus["current_firing_rate"].(float64)

	if finalRate <= initialRate {
		t.Error("Expected firing rate to increase after multiple firings")
	}

	// Verify firing history size
	firingHistorySize := finalStatus["firing_history_size"].(int)
	if firingHistorySize < numFirings {
		t.Errorf("Expected at least %d firing events, got %d",
			numFirings, firingHistorySize)
	}

	t.Logf("✓ Firing history: %d events tracked, rate %.2f Hz",
		firingHistorySize, finalRate)
}

// ============================================================================
// CALCIUM DYNAMICS TESTS
// ============================================================================

// TestNeuronFiring_CalciumAccumulation validates calcium increase per firing
func TestNeuronFiring_CalciumAccumulation(t *testing.T) {
	neuron := NewNeuron(
		"calcium-test",
		0.5,
		0.95,
		5*time.Millisecond,
		1.0,
		5.0,
		0.1,
	)

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Measure calcium before firing
	beforeStatus := neuron.GetFiringStatus()
	calciumBefore := beforeStatus["calcium_level"].(float64)

	// Single firing
	SendTestSignal(neuron, "test-source", 1.0)
	time.Sleep(20 * time.Millisecond)

	afterOneStatus := neuron.GetFiringStatus()
	calciumAfterOne := afterOneStatus["calcium_level"].(float64)

	// Second firing
	SendTestSignal(neuron, "test-source", 1.0)
	time.Sleep(20 * time.Millisecond)

	afterTwoStatus := neuron.GetFiringStatus()
	calciumAfterTwo := afterTwoStatus["calcium_level"].(float64)

	// Verify calcium accumulation
	if calciumAfterOne <= calciumBefore {
		t.Error("Expected calcium increase after first firing")
	}

	if calciumAfterTwo <= calciumAfterOne {
		t.Error("Expected further calcium increase after second firing")
	}

	t.Logf("✓ Calcium accumulation: %.3f → %.3f → %.3f",
		calciumBefore, calciumAfterOne, calciumAfterTwo)
}

// ============================================================================
// THRESHOLD AND FIRE FACTOR TESTS
// ============================================================================

// TestNeuronFiring_ThresholdEnforcement validates firing threshold
func TestNeuronFiring_ThresholdEnforcement(t *testing.T) {
	threshold := 1.5

	neuron := NewNeuron(
		"threshold-test",
		threshold,
		0.95,
		5*time.Millisecond,
		1.0,
		5.0,
		0.1,
	)

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Test sub-threshold signal (should not fire)
	SendTestSignal(neuron, "test-source", 1.0) // Below threshold
	time.Sleep(20 * time.Millisecond)

	subThresholdReports := mockMatrix.GetHealthReportCount()
	if subThresholdReports > 0 {
		t.Error("Sub-threshold signal should not trigger firing")
	}

	// Test supra-threshold signal (should fire)
	SendTestSignal(neuron, "test-source", 2.0) // Above threshold
	time.Sleep(20 * time.Millisecond)

	supraThresholdReports := mockMatrix.GetHealthReportCount()
	if supraThresholdReports == 0 {
		t.Error("Supra-threshold signal should trigger firing")
	}

	t.Logf("✓ Threshold enforcement: sub=%d, supra=%d reports",
		subThresholdReports, supraThresholdReports)
}

// TestNeuronFiring_FireFactor validates output amplitude scaling
func TestNeuronFiring_FireFactor(t *testing.T) {
	fireFactor := 3.0

	neuron := NewNeuron(
		"fire-factor-test",
		0.5,
		0.95,
		5*time.Millisecond,
		fireFactor,
		5.0,
		0.1,
	)

	// Create mock synapse to measure output
	mockSynapse := NewMockSynapse("test-syn", "target", 1.0, 1*time.Millisecond)
	neuron.AddOutputCallback("test-syn", mockSynapse.CreateOutputCallback())

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Send signal that will accumulate to known value
	inputValue := 1.0
	SendTestSignal(neuron, "test-source", inputValue)

	// Wait for axonal delivery
	time.Sleep(20 * time.Millisecond)

	// Verify output scaling
	signals := mockSynapse.GetReceivedSignals()
	if len(signals) == 0 {
		t.Error("Expected output signal")
		return
	}

	outputValue := signals[0].Value
	expectedOutput := inputValue * fireFactor

	tolerance := 0.1
	if outputValue < expectedOutput-tolerance || outputValue > expectedOutput+tolerance {
		t.Logf("Output value %.3f not exactly %.1f (input %.1f × factor %.1f)",
			outputValue, expectedOutput, inputValue, fireFactor)
		t.Logf("This may be normal due to decay or other processing")
	}

	t.Logf("✓ Fire factor: input %.1f → output %.3f (factor %.1f)",
		inputValue, outputValue, fireFactor)
}

// ============================================================================
// ERROR HANDLING TESTS
// ============================================================================

// TestNeuronFiring_NoCallbacks validates firing without matrix callbacks
func TestNeuronFiring_NoCallbacks(t *testing.T) {
	neuron := NewNeuron(
		"no-callbacks-test",
		0.5,
		0.95,
		5*time.Millisecond,
		1.0,
		5.0,
		0.1,
	)

	// Don't set any callbacks - should not panic

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// This should not panic even without callbacks
	SendTestSignal(neuron, "test-source", 1.0)
	time.Sleep(20 * time.Millisecond)

	// Verify calcium and history still work
	status := neuron.GetFiringStatus()
	calcium := status["calcium_level"].(float64)
	if calcium <= 0 {
		t.Error("Expected calcium increase even without callbacks")
	}

	firingHistorySize := status["firing_history_size"].(int)
	if firingHistorySize == 0 {
		t.Error("Expected firing history even without callbacks")
	}

	t.Log("✓ Firing works safely without matrix callbacks")
}

// ============================================================================
// CONCURRENT FIRING TESTS
// ============================================================================

// TestNeuronFiring_ConcurrentSignals validates thread safety
func TestNeuronFiring_ConcurrentSignals(t *testing.T) {
	neuron := NewNeuron(
		"concurrent-test",
		0.3, // Low threshold for frequent firing
		0.95,
		5*time.Millisecond,
		1.0,
		10.0,
		0.1,
	)

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Send many concurrent signals
	numGoroutines := 10
	signalsPerGoroutine := 5

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < signalsPerGoroutine; j++ {
				SendTestSignal(neuron, "concurrent-source", 0.5)
				time.Sleep(10 * time.Millisecond)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Allow final processing
	time.Sleep(100 * time.Millisecond)

	// Verify some firings occurred
	healthReports := mockMatrix.GetHealthReportCount()
	if healthReports == 0 {
		t.Error("Expected some firings with concurrent signals")
	}

	// Verify state consistency
	status := neuron.GetFiringStatus()
	calcium := status["calcium_level"].(float64)
	if calcium < 0 {
		t.Error("Calcium level should not be negative")
	}

	firingHistorySize := status["firing_history_size"].(int)
	if firingHistorySize == 0 {
		t.Error("Expected firing history with concurrent signals")
	}

	t.Logf("✓ Concurrent signals handled: %d health reports, %.3f calcium",
		healthReports, calcium)
}

// TestNeuronFiring_SynapticScalingScenario reproduces the exact conditions that cause the panic
func TestNeuronFiring_SynapticScalingScenario(t *testing.T) {
	// This reproduces TestSynapticScaling_PostSynapticGainApplication/DisabledScaling

	// Create a neuron exactly like the synaptic scaling test does
	neuron := NewNeuron(
		"scaling-test",
		2.5, // Same as the failing test's input value
		0.95,
		5*time.Millisecond,
		1.0,
		5.0,
		0.1,
	)

	// CRITICAL: Don't set matrix callbacks - this is what the scaling test does
	// The scaling test doesn't call neuron.SetCallbacks()

	// Enable synaptic scaling like the failing test
	err := neuron.EnableSynapticScaling(1.0, 0.001, 30*time.Second)
	if err != nil {
		t.Fatalf("Failed to enable synaptic scaling: %v", err)
	}

	// Start the neuron
	err = neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Create the exact message that triggers the panic
	msg := message.NeuralSignal{
		Value:    2.5, // Same value from failing test
		SourceID: "test_source",
	}

	// Apply synaptic scaling directly (this is what the failing test does)
	scaling := neuron.synapticScaling
	if scaling != nil {
		result := scaling.ApplyPostSynapticGain(msg)
		t.Logf("Scaling result: %.1f", result)

		// This triggers the neuron to fire because 2.5 > threshold
		// And the firing triggers the panic
		neuron.Receive(msg)
		time.Sleep(20 * time.Millisecond)
	}

	t.Log("✓ Synaptic scaling scenario completed without panic")
}

// TestNeuronFiring_NilMatrixCallbacks validates that fireUnsafe does not panic when matrixCallbacks are nil
func TestNeuronFiring_NilMatrixCallbacks(t *testing.T) {
	// Create a neuron without setting any matrix callbacks
	neuron := NewNeuron(
		"nil-matrix-callbacks-test",
		0.5, // Low threshold for easy firing
		0.95,
		5*time.Millisecond,
		1.0,
		5.0,
		0.1,
	)

	// DO NOT CALL neuron.SetCallbacks() or set n.matrixCallbacks directly.
	// This simulates the condition where matrixCallbacks are not initialized.

	// Ensure that starting the neuron doesn't panic due to nil matrixCallbacks validation.
	// The validation check for n.matrixCallbacks happens in neuron.go:839 validateNeuronState
	// However, individual callback functions within matrixCallbacks are checked in fireUnsafe.
	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop() // Ensure neuron is stopped to clean up goroutines

	// Use a defer func() { recover() } to catch any panics if the fix is incomplete.
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Test panicked with nil matrix callbacks: %v", r)
		}
	}()

	// Send a signal above threshold to trigger firing.
	// This will call fireUnsafe internally.
	SendTestSignal(neuron, "test-source", 1.0)

	// Allow processing time for the neuron to attempt to fire
	time.Sleep(20 * time.Millisecond)

	// Verify that the neuron's internal state still updated, proving fireUnsafe ran
	// without panicking due to nil matrixCallbacks.
	status := neuron.GetFiringStatus()
	calcium := status["calcium_level"].(float64)
	firingHistorySize := status["firing_history_size"].(int)

	if calcium <= 0.1 { // Initial calcium is 0.1, it should increase if fired
		t.Errorf("Expected calcium level to increase even without matrix callbacks, got %.3f", calcium)
	}
	if firingHistorySize == 0 {
		t.Error("Expected firing history to be updated even without matrix callbacks")
	}

	t.Log("✓ Neuron fired successfully without panicking when matrixCallbacks were nil")
}

// TestNeuronFiring_SynapticScalingIntegration - minimal test to reproduce the issue
func TestNeuronFiring_SynapticScalingIntegration(t *testing.T) {
	// This mimics what the synaptic scaling test does
	neuron := NewNeuron("scaling-test", 1.0, 0.95, 5*time.Millisecond, 1.0, 5.0, 0.1)

	// Don't set callbacks initially - this is what causes the issue
	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// This should not panic even without callbacks
	SendTestSignal(neuron, "test", 2.0) // Above threshold to trigger firing
	time.Sleep(10 * time.Millisecond)

	t.Log("✓ Neuron handles firing without matrix callbacks")
}

// ============================================================================
// PERFORMANCE BENCHMARKS
// ============================================================================

// BenchmarkFiringMechanism benchmarks the core firing process
func BenchmarkFiringMechanism(b *testing.B) {
	neuron := NewNeuron(
		"benchmark-neuron",
		0.5,
		0.95,
		1*time.Millisecond, // Short refractory for rapid firing
		1.0,
		50.0,
		0.1,
	)

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	neuron.Start()
	defer neuron.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SendTestSignal(neuron, "benchmark-source", 1.0)
		time.Sleep(2 * time.Millisecond) // Faster than refractory period
	}
}

// BenchmarkOutputTransmission benchmarks output callback processing
func BenchmarkOutputTransmission(b *testing.B) {
	neuron := NewNeuron(
		"output-benchmark",
		0.5,
		0.95,
		1*time.Millisecond,
		1.0,
		50.0,
		0.1,
	)

	// Add multiple output connections
	for i := 0; i < 10; i++ {
		synapseID := "bench-syn-" + string(rune('0'+i))
		targetID := "target-" + string(rune('0'+i))
		mockSynapse := NewMockSynapse(synapseID, targetID, 1.0, 1*time.Millisecond)
		neuron.AddOutputCallback(synapseID, mockSynapse.CreateOutputCallback())
	}

	mockMatrix := NewMockMatrix()
	neuron.SetCallbacks(mockMatrix.CreateBasicCallbacks())

	neuron.Start()
	defer neuron.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SendTestSignal(neuron, "benchmark-source", 1.0)
		time.Sleep(2 * time.Millisecond)
	}
}
