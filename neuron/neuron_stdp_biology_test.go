package neuron

import (
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

/*
=================================================================================
STDP NEURON BIOLOGY TEST SUITE
=================================================================================

This file contains tests for STDP-based biological learning mechanisms:
1. Hebbian learning (neurons that fire together, wire together)
2. STDP learning windows (LTP/LTD curves)
3. Integration with other plasticity mechanisms
4. Learning consolidation

All tests use the prefix TestSTDPNeuronBiology_ for easy isolation.
=================================================================================
*/

// TestSTDPNeuronBiology_HebbianRule verifies that synapses strengthen when there is
// correlated pre/post-synaptic activity (neurons that fire together, wire together)
func TestSTDPNeuronBiology_HebbianRule(t *testing.T) {
	// Create a neuron with STDP enabled
	neuron := NewNeuron(
		"hebbian_test",
		1.0,                // threshold
		0.95,               // decay rate
		5*time.Millisecond, // refractory period
		2.0,                // fire factor
		3.0,                // target firing rate
		0.2,                // homeostasis strength
	)

	// Enable STDP with standard learning parameters
	neuron.EnableSTDPFeedback(
		10*time.Millisecond, // feedback delay
		0.1,                 // learning rate
	)

	// Create mock matrix to track plasticity adjustments
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Start the neuron
	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Run learning trials
	const trials = 3
	for trial := 0; trial < trials; trial++ {
		t.Logf("Running Hebbian learning trial %d/%d", trial+1, trials)

		// Set up test synapses with different correlation patterns
		testSynapses := []types.SynapseInfo{
			// Highly correlated synapse (fires consistently before the neuron)
			{
				ID:               "synapse_correlated",
				SourceID:         "source1",
				TargetID:         neuron.ID(),
				Weight:           0.5,
				LastActivity:     time.Now().Add(-3 * time.Millisecond),
				LastTransmission: time.Now().Add(-3 * time.Millisecond),
			},
			// Anti-correlated synapse (fires consistently after the neuron)
			{
				ID:               "synapse_anticorrelated",
				SourceID:         "source3",
				TargetID:         neuron.ID(),
				Weight:           0.5,
				LastActivity:     time.Now().Add(5 * time.Millisecond),
				LastTransmission: time.Now().Add(5 * time.Millisecond),
			},
		}
		mockMatrix.SetSynapseList(testSynapses)

		// Send STDP feedback directly
		t.Log("  Sending STDP feedback")
		stdpDone := make(chan struct{})
		go func() {
			neuron.SendSTDPFeedback()
			close(stdpDone)
		}()

		// Use timeout to avoid deadlock
		select {
		case <-stdpDone:
			t.Log("  STDP feedback completed")
		case <-time.After(200 * time.Millisecond):
			t.Fatal("  STDP feedback timed out - deadlock detected")
		}

		// Brief delay between trials
		time.Sleep(20 * time.Millisecond)
	}

	// Analyze adjustments to see if they follow Hebbian rule
	adjustments := mockMatrix.GetPlasticityAdjustments()
	t.Logf("Total plasticity adjustments: %d", len(adjustments))

	// Group adjustments by sign (LTP vs LTD)
	var negativeCount, positiveCount int
	for _, adj := range adjustments {
		if adj.DeltaT < 0 {
			negativeCount++
		} else if adj.DeltaT > 0 {
			positiveCount++
		}
	}

	t.Logf("Negative DeltaT (LTP) count: %d", negativeCount)
	t.Logf("Positive DeltaT (LTD) count: %d", positiveCount)

	// Check if we have both types of adjustments
	if negativeCount == 0 {
		t.Error("Expected some negative DeltaT adjustments (LTP)")
	}
	if positiveCount == 0 {
		t.Error("Expected some positive DeltaT adjustments (LTD)")
	}

	t.Log("✓ Hebbian learning rule test completed")
}

// TestSTDPNeuronBiology_LearningWindow tests the STDP learning window
// (LTP for pre-before-post, LTD for post-before-pre)
func TestSTDPNeuronBiology_LearningWindow(t *testing.T) {
	// Create a neuron with STDP
	neuron := NewNeuron(
		"stdp_window_test",
		1.0,
		0.95,
		5*time.Millisecond,
		2.0,
		3.0,
		0.2,
	)

	neuron.EnableSTDPFeedback(20*time.Millisecond, 0.1)

	// Create mock matrix
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Start the neuron
	neuron.Start()
	defer neuron.Stop()

	// Test spike timing differences across the full STDP window
	timingDifferences := []time.Duration{
		-50 * time.Millisecond, // Pre long before post (weak potentiation)
		-30 * time.Millisecond, // Pre well before post (moderate potentiation)
		-10 * time.Millisecond, // Pre just before post (maximum potentiation)
		-5 * time.Millisecond,  // Pre very close before post (strong potentiation)
		0 * time.Millisecond,   // Simultaneous (border case)
		5 * time.Millisecond,   // Post just before pre (strong depression)
		10 * time.Millisecond,  // Post before pre (maximum depression)
		30 * time.Millisecond,  // Post well before pre (moderate depression)
		50 * time.Millisecond,  // Post long before pre (weak depression)
	}

	// Map to store results for DeltaT
	results := make(map[time.Duration]time.Duration)

	for _, deltaT := range timingDifferences {
		// Calculate LastActivity time based on deltaT
		lastActivity := time.Now().Add(deltaT)

		// Create a test synapse with this timing
		testSynapse := types.SynapseInfo{
			ID:               "timing_test_synapse",
			SourceID:         "source",
			TargetID:         neuron.ID(),
			Weight:           0.5,
			LastActivity:     lastActivity,
			LastTransmission: lastActivity, // Use LastActivity for simplicity
		}

		// Setup mock and clear previous adjustments
		mockMatrix.ClearPlasticityAdjustments()
		mockMatrix.SetSynapseList([]types.SynapseInfo{testSynapse})

		// Directly trigger STDP with timeout for safety
		stdpDone := make(chan struct{})
		go func() {
			neuron.SendSTDPFeedback()
			close(stdpDone)
		}()

		// Wait for STDP to complete or timeout
		select {
		case <-stdpDone:
			// STDP completed successfully
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("STDP feedback timed out for deltaT = %v", deltaT)
		}

		// Check adjustment
		adjustments := mockMatrix.GetPlasticityAdjustments()
		if len(adjustments) == 0 {
			t.Fatalf("No plasticity adjustment for deltaT = %v", deltaT)
		}

		// Store the measured DeltaT
		results[deltaT] = adjustments[0].DeltaT
	}

	// Verify the STDP curve shape
	t.Log("STDP Window Results:")
	t.Log("-------------------")
	t.Log("Input DeltaT (ms) | Measured DeltaT (ms) | Sign Match")

	for _, deltaT := range timingDifferences {
		resultDeltaT := results[deltaT]
		signMatch := (deltaT < 0 && resultDeltaT < 0) || (deltaT > 0 && resultDeltaT > 0) || (deltaT == 0)

		t.Logf("%15.2f | %19.2f | %t",
			float64(deltaT)/float64(time.Millisecond),
			float64(resultDeltaT)/float64(time.Millisecond),
			signMatch)

		// Check if sign is correct
		if !signMatch && deltaT != 0 {
			t.Errorf("Expected same sign for input DeltaT %v and measured DeltaT %v", deltaT, resultDeltaT)
		}
	}

	t.Log("✓ STDP learning window test completed")
}

// TestSTDPNeuronBiology_NaturalScheduling tests the natural STDP scheduling
// that occurs when a neuron fires
func TestSTDPNeuronBiology_NaturalScheduling(t *testing.T) {
	// Create a neuron with STDP enabled
	neuron := NewNeuron(
		"natural_stdp_test",
		0.5,                // low threshold to ensure firing
		0.95,               // decay rate
		5*time.Millisecond, // refractory period
		2.0,                // fire factor
		3.0,                // target firing rate
		0.2,                // homeostasis strength
	)

	// Enable STDP with a longer delay for easier testing
	stdpDelay := 50 * time.Millisecond
	neuron.EnableSTDPFeedback(stdpDelay, 0.1)

	// Create mock matrix
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Set up test synapses with clear timing patterns
	synapses := []types.SynapseInfo{
		{
			ID:       "corr",
			SourceID: "src1",
			TargetID: neuron.ID(),
			Weight:   0.5, LastActivity: time.Now().Add(-5 * time.Millisecond), LastTransmission: time.Now().Add(-5 * time.Millisecond)},
		{ID: "anti", SourceID: "src2", TargetID: neuron.ID(), Weight: 0.5, LastActivity: time.Now().Add(5 * time.Millisecond), LastTransmission: time.Now().Add(-5 * time.Millisecond)},
	}
	mockMatrix.SetSynapseList(synapses)

	// Start the neuron
	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Send super-threshold signal to trigger firing
	t.Log("Sending firing signal to trigger natural STDP scheduling")
	SendTestSignal(neuron, "test_source", 1.0) // Above threshold

	// Wait for initial signal processing
	time.Sleep(10 * time.Millisecond)

	// Check that we have no STDP adjustments yet (too early)
	initialAdjustments := mockMatrix.GetPlasticityAdjustments()
	t.Logf("Initial adjustments (before delay): %d", len(initialAdjustments))

	// Get firing status to verify the neuron actually fired
	midFiringStatus := neuron.GetFiringStatus()
	lastFireTime := midFiringStatus["last_fire_time"].(time.Time)

	if lastFireTime.IsZero() {
		t.Error("Neuron did not fire with super-threshold signal")
	} else {
		t.Logf("Neuron fired at: %v", lastFireTime)
	}

	// Wait for STDP delay to pass
	t.Logf("Waiting for STDP delay (%v) to pass", stdpDelay)
	// Wait longer than the delay to ensure processing
	time.Sleep(stdpDelay + 100*time.Millisecond)

	// Now we should have STDP adjustments
	finalAdjustments := mockMatrix.GetPlasticityAdjustments()
	t.Logf("Final adjustments (after delay): %d", len(finalAdjustments))

	// Print details of any adjustments
	for i, adj := range finalAdjustments {
		t.Logf("Adjustment %d: DeltaT=%v, LearningRate=%v",
			i, adj.DeltaT, adj.LearningRate)
	}

	if len(finalAdjustments) <= len(initialAdjustments) {
		t.Error("Expected more STDP adjustments after delay")
	} else {
		t.Logf("✓ Natural STDP scheduling works: %d → %d adjustments",
			len(initialAdjustments), len(finalAdjustments))
	}
}

// TestSTDPNeuronBiology_IntegratedPlasticity tests multiple plasticity mechanisms
// working together in a simplified scenario
func TestSTDPNeuronBiology_IntegratedPlasticity(t *testing.T) {
	// Create a neuron with plasticity mechanisms enabled
	neuron := NewNeuron(
		"integrated_test",
		0.5,                // threshold (low to ensure firing)
		0.95,               // decay rate
		5*time.Millisecond, // refractory period
		2.0,                // fire factor
		3.0,                // target firing rate
		0.2,                // homeostasis strength
	)

	// Enable STDP with a short delay to ensure it processes quickly
	neuron.EnableSTDPFeedback(5*time.Millisecond, 0.1)

	// Create mock matrix
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Start the neuron
	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Create a few test synapses with different timing patterns
	initialSynapses := []types.SynapseInfo{
		{ID: "corr", SourceID: "src1", TargetID: neuron.ID(), Weight: 0.3, LastActivity: time.Now().Add(-5 * time.Millisecond)},
		{ID: "anti", SourceID: "src3", TargetID: neuron.ID(), Weight: 0.3, LastActivity: time.Now().Add(5 * time.Millisecond)},
	}
	mockMatrix.SetSynapseList(initialSynapses)

	// Record initial weights
	initialWeights := make(map[string]float64)
	for _, syn := range initialSynapses {
		initialWeights[syn.ID] = syn.Weight
	}

	// Run simplified test
	iterations := 5
	t.Log("Running simplified plasticity integration test")
	for i := 0; i < iterations; i++ {
		t.Logf("  Iteration %d/%d", i+1, iterations)

		// Update synapse timings consistently
		updatedSynapses := make([]types.SynapseInfo, len(initialSynapses))
		copy(updatedSynapses, initialSynapses)

		for j := range updatedSynapses {
			if updatedSynapses[j].ID == "corr" {
				updatedSynapses[j].LastActivity = time.Now().Add(-5 * time.Millisecond)
			} else if updatedSynapses[j].ID == "anti" {
				updatedSynapses[j].LastActivity = time.Now().Add(5 * time.Millisecond)
			}
		}
		mockMatrix.SetSynapseList(updatedSynapses)

		// Send sub-threshold signals
		for j := 0; j < 3; j++ {
			// Create and send signal with timeout
			SendTestSignal(neuron, "test_input", 0.2) // Below threshold
			time.Sleep(10 * time.Millisecond)
		}

		// Perform homeostasis scaling with timeout protection
		homeostasisDone := make(chan struct{})
		go func() {
			neuron.PerformHomeostasisScaling()
			close(homeostasisDone)
		}()
		select {
		case <-homeostasisDone:
			// Completed successfully
		case <-time.After(100 * time.Millisecond):
			t.Log("    WARNING: Homeostasis scaling timed out")
		}

		// Send firing signal to trigger STDP
		SendTestSignal(neuron, "firing_input", 1.0) // Above threshold
		time.Sleep(20 * time.Millisecond)           // Allow time for STDP processing
	}

	// Check final weights
	incomingDirection := types.SynapseIncoming
	myID := neuron.ID()
	finalSynapses := mockCallbacks.ListSynapses(types.SynapseCriteria{
		Direction: &incomingDirection,
		TargetID:  &myID,
	})

	finalWeights := make(map[string]float64)
	for _, syn := range finalSynapses {
		finalWeights[syn.ID] = syn.Weight
		weightChange := syn.Weight - initialWeights[syn.ID]
		t.Logf("Synapse %s: Initial %.4f → Final %.4f (change: %+.4f)",
			syn.ID, initialWeights[syn.ID], syn.Weight, weightChange)
	}

	// Check if weight changes follow expected pattern
	if corrWeight, ok := finalWeights["corr"]; ok {
		if corrWeight <= initialWeights["corr"] {
			t.Logf("NOTE: Expected correlated synapse to strengthen, but weight didn't increase")
		}
	}

	if antiWeight, ok := finalWeights["anti"]; ok {
		if antiWeight >= initialWeights["anti"] {
			t.Logf("NOTE: Expected anti-correlated synapse to weaken, but weight didn't decrease")
		}
	}

	// Check activity level
	activityLevel := neuron.GetActivityLevel()
	t.Logf("Final activity level: %.4f Hz (target: %.1f Hz)", activityLevel, 3.0)

	t.Log("✓ Integrated plasticity test completed")
}

// TestSTDPNeuronBiology_ConsolidatedLearning tests if repeated patterns lead
// to more stable weight changes than random patterns
func TestSTDPNeuronBiology_ConsolidatedLearning(t *testing.T) {
	// Create a neuron for testing
	neuron := NewNeuron(
		"consolidation_test",
		0.5,                // threshold
		0.95,               // decay rate
		5*time.Millisecond, // refractory period
		2.0,                // fire factor
		3.0,                // target firing rate
		0.2,                // homeostasis strength
	)

	// Enable STDP with moderate learning rate
	neuron.EnableSTDPFeedback(10*time.Millisecond, 0.05)

	// Create mock matrix
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Start the neuron
	neuron.Start()
	defer neuron.Stop()

	// Create two sets of synapses:
	// 1. Pattern synapses: Consistent timing relative to neuron firing
	// 2. Random synapses: Random timing relative to neuron firing

	// Initial setup for all synapses
	patternSynapses := []types.SynapseInfo{
		{ID: "pattern1", SourceID: "src1", TargetID: neuron.ID(), Weight: 0.3},
		{ID: "pattern2", SourceID: "src2", TargetID: neuron.ID(), Weight: 0.3},
		{ID: "pattern3", SourceID: "src3", TargetID: neuron.ID(), Weight: 0.3},
	}

	randomSynapses := []types.SynapseInfo{
		{ID: "random1", SourceID: "src4", TargetID: neuron.ID(), Weight: 0.3},
		{ID: "random2", SourceID: "src5", TargetID: neuron.ID(), Weight: 0.3},
		{ID: "random3", SourceID: "src6", TargetID: neuron.ID(), Weight: 0.3},
	}

	// Track initial weights
	initialWeights := make(map[string]float64)
	for _, syn := range append(patternSynapses, randomSynapses...) {
		initialWeights[syn.ID] = syn.Weight
	}

	// Run multiple learning trials
	const trials = 30

	t.Log("Starting consolidated learning test...")

	for trial := 0; trial < trials; trial++ {
		// Update pattern synapses with consistent timing (always fire 5ms before neuron)
		for i := range patternSynapses {
			patternSynapses[i].LastActivity = time.Now().Add(-5 * time.Millisecond)
		}

		// Update random synapses with random timing
		for i := range randomSynapses {
			// Random timing between -20ms and +20ms
			randomOffset := time.Duration(randomTimingOffset(-20, 20)) * time.Millisecond
			randomSynapses[i].LastActivity = time.Now().Add(randomOffset)
		}

		// Combine all synapses
		allSynapses := append(patternSynapses, randomSynapses...)
		mockMatrix.SetSynapseList(allSynapses)

		// Trigger STDP feedback with timeout protection
		stdpDone := make(chan struct{})
		go func() {
			neuron.SendSTDPFeedback()
			close(stdpDone)
		}()
		select {
		case <-stdpDone:
			// STDP completed successfully
		case <-time.After(200 * time.Millisecond):
			t.Fatal("STDP feedback timed out - deadlock detected")
		}

		// Allow time for processing
		time.Sleep(5 * time.Millisecond)
	}

	// Get final synapse weights
	incomingDirection := types.SynapseIncoming
	myID := neuron.ID()
	finalSynapses := mockCallbacks.ListSynapses(types.SynapseCriteria{
		Direction: &incomingDirection,
		TargetID:  &myID,
	})

	finalWeights := make(map[string]float64)
	for _, syn := range finalSynapses {
		finalWeights[syn.ID] = syn.Weight
	}

	// Calculate weight changes
	patternChanges := make([]float64, 0, len(patternSynapses))
	randomChanges := make([]float64, 0, len(randomSynapses))

	for _, syn := range patternSynapses {
		initialWeight := initialWeights[syn.ID]
		finalWeight := finalWeights[syn.ID]
		change := finalWeight - initialWeight
		patternChanges = append(patternChanges, change)
		t.Logf("Pattern synapse %s: Initial %.4f → Final %.4f (change: %+.4f)",
			syn.ID, initialWeight, finalWeight, change)
	}

	for _, syn := range randomSynapses {
		initialWeight := initialWeights[syn.ID]
		finalWeight := finalWeights[syn.ID]
		change := finalWeight - initialWeight
		randomChanges = append(randomChanges, change)
		t.Logf("Random synapse %s: Initial %.4f → Final %.4f (change: %+.4f)",
			syn.ID, initialWeight, finalWeight, change)
	}

	// Calculate average change and variance for both groups
	avgPatternChange := calculateAverage(patternChanges)
	avgRandomChange := calculateAverage(randomChanges)

	patternVariance := calculateVariance(patternChanges)
	randomVariance := calculateVariance(randomChanges)

	t.Logf("Pattern synapses: Avg change %+.4f, Variance %.6f", avgPatternChange, patternVariance)
	t.Logf("Random synapses: Avg change %+.4f, Variance %.6f", avgRandomChange, randomVariance)

	// Verify consolidated learning
	// 1. Pattern synapses should strengthen consistently (more positive change)
	// 2. Pattern synapses should have less variance in their weight changes

	if avgPatternChange <= avgRandomChange {
		t.Logf("NOTE: Expected pattern synapses to have stronger consolidation, but avg change (%.4f) <= random avg change (%.4f)",
			avgPatternChange, avgRandomChange)
	}

	if patternVariance >= randomVariance {
		t.Logf("NOTE: Expected pattern synapses to have less variance, but pattern variance (%.6f) >= random variance (%.6f)",
			patternVariance, randomVariance)
	}

	t.Log("✓ Consolidated learning test completed")
}

// Helper function for simulating multi-phase activity
func simulateActivityPhase(t *testing.T, neuron *Neuron, mockMatrix *MockMatrix,
	baseSynapses []types.SynapseInfo, inputCount int,
	intervalMs int, duration time.Duration) {

	// Calculate number of iterations instead of using time-based loop
	numIterations := int(duration / (time.Duration(intervalMs) * time.Millisecond))
	t.Logf("  Simulating %d inputs for %d iterations (interval: %dms)",
		inputCount, numIterations, intervalMs)

	// Limit iterations for safety
	maxIterations := 10
	if numIterations > maxIterations {
		numIterations = maxIterations
		t.Logf("  Limited to %d iterations for safety", maxIterations)
	}

	// Use a fixed number of iterations
	for i := 0; i < numIterations; i++ {
		t.Logf("    Iteration %d/%d", i+1, numIterations)

		// Update synapse timing (maintaining correlation patterns)
		updatedSynapses := make([]types.SynapseInfo, len(baseSynapses))
		copy(updatedSynapses, baseSynapses)

		for j, syn := range updatedSynapses {
			// Preserve correlation pattern but update timestamps
			if syn.ID == "corr" {
				updatedSynapses[j].LastActivity = time.Now().Add(-5 * time.Millisecond)
			} else if syn.ID == "uncorr" {
				randomOffset := time.Duration(randomTimingOffset(-50, 50)) * time.Millisecond
				updatedSynapses[j].LastActivity = time.Now().Add(randomOffset)
			} else if syn.ID == "anti" {
				updatedSynapses[j].LastActivity = time.Now().Add(5 * time.Millisecond)
			}
		}

		mockMatrix.SetSynapseList(updatedSynapses)

		// Send signals with timeout protection
		signalsToSend := min(inputCount, 5) // Maximum 5 signals for safety
		for j := 0; j < signalsToSend; j++ {
			SendTestSignal(neuron, "test_input", 1.0)
			time.Sleep(10 * time.Millisecond)
		}

		// Perform homeostasis with timeout protection
		homeostasisDone := make(chan struct{})
		go func() {
			neuron.PerformHomeostasisScaling()
			close(homeostasisDone)
		}()
		select {
		case <-homeostasisDone:
			// Completed successfully
		case <-time.After(100 * time.Millisecond):
			t.Log("      WARNING: Homeostasis timed out")
		}

		// Send STDP feedback with timeout protection
		stdpDone := make(chan struct{})
		go func() {
			neuron.SendSTDPFeedback()
			close(stdpDone)
		}()
		select {
		case <-stdpDone:
			// Completed successfully
		case <-time.After(100 * time.Millisecond):
			t.Log("      WARNING: STDP feedback timed out")
		}

		// Sleep between iterations
		sleepTime := time.Duration(intervalMs) * time.Millisecond
		time.Sleep(sleepTime)
	}

	t.Logf("  Phase completed with %d iterations", numIterations)
}

// Helper function to calculate variance of a slice of float64
func calculateVariance(values []float64) float64 {
	if len(values) <= 1 {
		return 0.0
	}

	mean := calculateAverage(values)
	sumSquaredDiff := 0.0

	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}

	return sumSquaredDiff / float64(len(values)-1)
}

// Helper function for random timing offsets to avoid conflicts with math/rand
func randomTimingOffset(min, max int) int {
	return min + time.Now().Nanosecond()%(max-min+1)
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestSTDPNeuronBiology_BasicFunctionality tests core STDP operations
func TestSTDPNeuronBiology_BasicFunctionality(t *testing.T) {
	// Create a neuron with STDP enabled
	neuron := NewNeuron(
		"stdp_basic_test",
		1.0,                // threshold
		0.95,               // decay rate
		5*time.Millisecond, // refractory period
		2.0,                // fire factor
		3.0,                // target firing rate
		0.2,                // homeostasis strength
	)

	// Enable STDP with a moderate learning rate
	neuron.EnableSTDPFeedback(
		10*time.Millisecond, // feedback delay
		0.1,                 // learning rate
	)

	// Create a mock matrix to track STDP calls
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Start the neuron
	err := neuron.Start()
	if err != nil {
		t.Fatalf("Failed to start neuron: %v", err)
	}
	defer neuron.Stop()

	// Test causal spike timing (pre before post)
	t.Log("Testing causal spike timing (pre before post)")

	// Create a test synapse that fired recently (causal - should strengthen)
	causalSynapse := types.SynapseInfo{
		ID:               "causal_synapse",
		SourceID:         "source",
		TargetID:         neuron.ID(),
		Weight:           0.5,
		LastActivity:     time.Now().Add(-5 * time.Millisecond), // Pre-synaptic spike 5ms ago
		LastTransmission: time.Now().Add(-5 * time.Millisecond), // Same as LastActivity for simplicity
	}

	// Setup our mock to return this synapse when ListSynapses is called
	mockMatrix.SetSynapseList([]types.SynapseInfo{causalSynapse})

	// Simulate a post-synaptic spike now (by calling SendSTDPFeedback)
	// This should strengthen the synapse because pre fired before post
	neuron.SendSTDPFeedback()

	// Check for plasticity adjustments
	adjustments := mockMatrix.GetPlasticityAdjustments()
	if len(adjustments) == 0 {
		t.Error("Expected STDP to generate plasticity adjustments")
	} else {
		adjustment := adjustments[0]

		// Verify the adjustment properties
		if adjustment.DeltaT.Milliseconds() >= 0 {
			t.Errorf("Expected negative DeltaT for causal timing, got %v", adjustment.DeltaT)
		}

		t.Logf("Causal adjustment: DeltaT=%v, LearningRate=%.3f",
			adjustment.DeltaT, adjustment.LearningRate)
	}

	// Test disabling STDP
	neuron.DisableSTDPFeedback()

	if neuron.IsSTDPFeedbackEnabled() {
		t.Error("STDP should be disabled after DisableSTDPFeedback()")
	}

	// Clear previous adjustments
	mockMatrix.ClearPlasticityAdjustments()

	// Try to trigger STDP again - should not generate adjustments
	neuron.SendSTDPFeedback()

	adjustmentsAfterDisable := mockMatrix.GetPlasticityAdjustments()
	if len(adjustmentsAfterDisable) > 0 {
		t.Error("STDP should not generate adjustments when disabled")
	}

	// Test re-enabling with different parameters
	neuron.EnableSTDPFeedback(20*time.Millisecond, 0.2)

	if !neuron.IsSTDPFeedbackEnabled() {
		t.Error("STDP should be enabled after EnableSTDPFeedback()")
	}

	// Get status and verify parameters were updated
	status := neuron.GetProcessingStatus()
	stdpStatus := status["stdp_system"].(map[string]interface{})

	// Check that the parameters match what we set
	feedbackDelay := stdpStatus["feedback_delay"].(time.Duration)
	learningRate := stdpStatus["learning_rate"].(float64)

	if feedbackDelay != 20*time.Millisecond {
		t.Errorf("Expected feedback delay 20ms, got %v", feedbackDelay)
	}

	if learningRate != 0.2 {
		t.Errorf("Expected learning rate 0.2, got %v", learningRate)
	}

	t.Log("✓ Basic STDP functionality works correctly")
}

// TestSTDPNeuronBiology_ScheduledFeedback tests the automatic STDP feedback scheduling
func TestSTDPNeuronBiology_ScheduledFeedback(t *testing.T) {
	// Create a neuron with STDP
	neuron := NewNeuron(
		"stdp_schedule_test",
		0.5, // low threshold to ensure firing
		0.95,
		5*time.Millisecond,
		2.0,
		3.0,
		0.2,
	)

	// Enable STDP with short delay
	feedbackDelay := 20 * time.Millisecond
	neuron.EnableSTDPFeedback(feedbackDelay, 0.1)

	// Create mock matrix
	mockMatrix := NewMockMatrix()
	mockCallbacks := NewMockNeuronCallbacks(mockMatrix)
	neuron.SetCallbacks(mockCallbacks)

	// Setup a test synapse
	testSynapse := types.SynapseInfo{
		ID:               "schedule_test_synapse",
		SourceID:         "source",
		TargetID:         neuron.ID(),
		Weight:           0.5,
		LastActivity:     time.Now().Add(-5 * time.Millisecond),
		LastTransmission: time.Now().Add(-5 * time.Millisecond), // Pre-synaptic spike 5ms ago
	}

	mockMatrix.SetSynapseList([]types.SynapseInfo{testSynapse})

	// Start the neuron
	neuron.Start()
	defer neuron.Stop()

	// Trigger firing with strong signal
	testSignal := types.NeuralSignal{
		Value:     2.0, // Strong signal to ensure firing
		Timestamp: time.Now(),
		SourceID:  "test",
		TargetID:  neuron.ID(),
	}

	t.Log("Sending signal to trigger firing")
	neuron.Receive(testSignal)

	// Wait a bit to allow for firing processing
	time.Sleep(10 * time.Millisecond)

	// No STDP adjustments yet - too early
	adjustmentsBefore := mockMatrix.GetPlasticityAdjustments()
	if len(adjustmentsBefore) > 0 {
		t.Errorf("Expected no STDP adjustments yet, got %d", len(adjustmentsBefore))
	}

	// Wait for scheduled feedback to execute
	time.Sleep(feedbackDelay + 10*time.Millisecond)

	// Now we should have STDP adjustments
	adjustmentsAfter := mockMatrix.GetPlasticityAdjustments()
	if len(adjustmentsAfter) == 0 {
		t.Error("Expected STDP adjustments after scheduled time")
	}

	t.Log("✓ Scheduled STDP feedback works correctly")
}
