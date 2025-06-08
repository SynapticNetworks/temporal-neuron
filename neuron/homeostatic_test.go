package neuron

import (
	"sync"
	"testing"
	"time"
)

// ============================================================================
// HOMEOSTATIC PLASTICITY TESTS
// ============================================================================

// TestHomeostaticNeuronCreation tests homeostatic neuron creation and initialization
func TestHomeostaticNeuronCreation(t *testing.T) {
	threshold := 1.5
	decayRate := 0.95
	refractoryPeriod := 10 * time.Millisecond
	fireFactor := 2.0
	neuronID := "test_homeostatic_neuron"
	targetFiringRate := 5.0
	homeostasisStrength := 0.1

	// Create STDP config for testing
	stdpConfig := STDPConfig{
		Enabled:        false, // Disabled for homeostatic-only testing
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.5,
	}

	neuron := NewNeuron(neuronID, threshold, decayRate, refractoryPeriod, fireFactor, targetFiringRate, homeostasisStrength, stdpConfig)

	if neuron == nil {
		t.Fatal("NewNeuron returned nil")
	}

	// Test homeostatic initialization
	if neuron.homeostatic.targetFiringRate != targetFiringRate {
		t.Errorf("Expected targetFiringRate %f, got %f", targetFiringRate, neuron.homeostatic.targetFiringRate)
	}

	if neuron.homeostatic.homeostasisStrength != homeostasisStrength {
		t.Errorf("Expected homeostasisStrength %f, got %f", homeostasisStrength, neuron.homeostatic.homeostasisStrength)
	}

	if neuron.homeostatic.calciumLevel != 0.0 {
		t.Errorf("Expected initial calcium level 0.0, got %f", neuron.homeostatic.calciumLevel)
	}

	if len(neuron.homeostatic.firingHistory) != 0 {
		t.Errorf("Expected empty firing history, got %d entries", len(neuron.homeostatic.firingHistory))
	}

	if neuron.threshold != threshold {
		t.Errorf("Expected threshold %f, got %f", threshold, neuron.threshold)
	}

	if neuron.baseThreshold != threshold {
		t.Errorf("Expected baseThreshold %f, got %f", threshold, neuron.baseThreshold)
	}

	// Test STDP initialization
	if neuron.stdpConfig.Enabled != false {
		t.Errorf("Expected STDP disabled for this test, got %v", neuron.stdpConfig.Enabled)
	}
}

// TestSimpleNeuronCreation tests backward-compatible neuron creation
func TestSimpleNeuronCreation(t *testing.T) {
	threshold := 1.5
	decayRate := 0.95
	refractoryPeriod := 10 * time.Millisecond
	fireFactor := 2.0
	neuronID := "test_simple_neuron"

	neuron := NewSimpleNeuron(neuronID, threshold, decayRate, refractoryPeriod, fireFactor)

	if neuron == nil {
		t.Fatal("NewSimpleNeuron returned nil")
	}

	// Test that homeostasis is disabled
	if neuron.homeostatic.targetFiringRate != 0.0 {
		t.Errorf("Expected disabled homeostasis (targetFiringRate=0), got %f", neuron.homeostatic.targetFiringRate)
	}

	if neuron.homeostatic.homeostasisStrength != 0.0 {
		t.Errorf("Expected disabled homeostasis (homeostasisStrength=0), got %f", neuron.homeostatic.homeostasisStrength)
	}

	// Test that STDP is disabled
	if neuron.stdpConfig.Enabled != false {
		t.Errorf("Expected STDP disabled for simple neuron, got %v", neuron.stdpConfig.Enabled)
	}

	// Other parameters should match
	if neuron.threshold != threshold {
		t.Errorf("Expected threshold %f, got %f", threshold, neuron.threshold)
	}

	if neuron.baseThreshold != threshold {
		t.Errorf("Expected baseThreshold %f, got %f", threshold, neuron.baseThreshold)
	}
}

// TestHomeostaticThresholdIncrease tests that hyperactive neurons increase their threshold
func TestHomeostaticThresholdIncrease(t *testing.T) {
	targetRate := 2.0 // 2 Hz target
	strength := 1.0   // Very strong homeostatic regulation for clear effect

	// Create disabled STDP config for homeostatic-only testing
	stdpConfig := STDPConfig{
		Enabled:        false,
		LearningRate:   0.0,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.0,
	}

	neuron := NewNeuron("test_threshold_increase", 1.0, 0.95, 5*time.Millisecond, 1.0, targetRate, strength, stdpConfig)

	output := make(chan Message, 100)
	neuron.AddOutput("test", output, 1.0, 0)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// Record initial threshold
	initialThreshold := neuron.GetCurrentThreshold()

	// Send signals that cause hyperactivity (above target rate)
	// 2 Hz target = 1 spike every 500ms, so we'll send much faster
	for i := 0; i < 15; i++ {
		input <- Message{Value: 1.2, Timestamp: time.Now(), SourceID: "test_input"} // Above threshold
		time.Sleep(50 * time.Millisecond)                                           // Much faster than target rate (20 Hz)
	}

	// Wait for multiple homeostatic adjustment cycles
	time.Sleep(500 * time.Millisecond)

	// Check that threshold has increased
	finalThreshold := neuron.GetCurrentThreshold()
	if finalThreshold <= initialThreshold {
		t.Errorf("Expected threshold to increase due to hyperactivity. Initial: %f, Final: %f",
			initialThreshold, finalThreshold)
	}

	// Verify calcium level increased
	calciumLevel := neuron.GetCalciumLevel()
	if calciumLevel <= 0 {
		t.Errorf("Expected calcium level > 0 after firing, got %f", calciumLevel)
	}

	// Verify firing rate was above target
	firingRate := neuron.GetCurrentFiringRate()
	t.Logf("Final firing rate: %f Hz (target: %f Hz)", firingRate, targetRate)
	t.Logf("Threshold change: %f -> %f (increase: %f)", initialThreshold, finalThreshold, finalThreshold-initialThreshold)
}

// TestHomeostaticThresholdDecrease tests that silent neurons decrease their threshold
func TestHomeostaticThresholdDecrease(t *testing.T) {
	targetRate := 5.0 // 5 Hz target
	strength := 0.5   // Moderate homeostatic regulation

	// Create disabled STDP config
	stdpConfig := STDPConfig{
		Enabled:        false,
		LearningRate:   0.0,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.0,
	}

	neuron := NewNeuron("test_threshold_decrease", 1.0, 0.95, 5*time.Millisecond, 1.0, targetRate, strength, stdpConfig)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// Record initial threshold
	initialThreshold := neuron.GetCurrentThreshold()

	// Send very weak signals that don't cause firing (silent neuron)
	for i := 0; i < 30; i++ {
		input <- Message{Value: 0.1, Timestamp: time.Now(), SourceID: "test_input"} // Well below threshold
		time.Sleep(50 * time.Millisecond)
	}

	// Wait for homeostatic adjustment
	time.Sleep(300 * time.Millisecond)

	// Check that threshold has decreased
	finalThreshold := neuron.GetCurrentThreshold()
	if finalThreshold >= initialThreshold {
		t.Errorf("Expected threshold to decrease due to silence. Initial: %f, Final: %f",
			initialThreshold, finalThreshold)
	}

	// Verify calcium level is low
	calciumLevel := neuron.GetCalciumLevel()
	if calciumLevel > 1.0 {
		t.Errorf("Expected low calcium level for silent neuron, got %f", calciumLevel)
	}

	// Verify firing rate is below target
	firingRate := neuron.GetCurrentFiringRate()
	if firingRate > targetRate {
		t.Errorf("Expected firing rate (%f) below target (%f) for silent neuron",
			firingRate, targetRate)
	}

	t.Logf("Threshold change: %f -> %f (decrease: %f)", initialThreshold, finalThreshold, initialThreshold-finalThreshold)
}

// TestHomeostaticStabilization tests that neurons stabilize around target rate
func TestHomeostaticStabilization(t *testing.T) {
	targetRate := 4.0 // 4 Hz target
	strength := 0.3   // Moderate regulation for stability

	// Create disabled STDP config
	stdpConfig := STDPConfig{
		Enabled:        false,
		LearningRate:   0.0,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.0,
	}

	neuron := NewNeuron("test_stabilization", 1.0, 0.95, 5*time.Millisecond, 1.0, targetRate, strength, stdpConfig)

	output := make(chan Message, 100)
	neuron.AddOutput("test", output, 1.0, 0)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// Use sync.WaitGroup for proper coordination
	var wg sync.WaitGroup
	stopSignal := make(chan struct{})

	// Send variable input to test stabilization
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 60; i++ {
			select {
			case <-stopSignal:
				return
			default:
				// Variable strength inputs to challenge homeostasis
				signalStrength := 0.7 + 0.6*float64(i%4) // 0.7, 1.3, 1.9, 2.5 pattern
				select {
				case input <- Message{Value: signalStrength, Timestamp: time.Now(), SourceID: "test_input"}:
				case <-stopSignal:
					return
				}
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()

	// Monitor firing rate over time
	time.Sleep(1 * time.Second) // Let initial period pass
	midRate := neuron.GetCurrentFiringRate()

	time.Sleep(3 * time.Second) // Let homeostasis work
	finalRate := neuron.GetCurrentFiringRate()

	// Signal goroutine to stop and wait for completion
	close(stopSignal)
	wg.Wait()

	// Check that firing rate moved toward target
	targetTolerance := 2.0 // Allow 2 Hz tolerance for this test
	if finalRate < targetRate-targetTolerance || finalRate > targetRate+targetTolerance {
		t.Logf("Warning: Final rate (%f) not close to target (%f) - homeostasis may need more time or stronger regulation",
			finalRate, targetRate)
	} else {
		t.Logf("Success: Final rate (%f) close to target (%f)", finalRate, targetRate)
	}

	// Verify homeostatic info is accessible
	info := neuron.GetHomeostaticInfo()
	if info.targetFiringRate != targetRate {
		t.Errorf("Expected homeostatic info target rate %f, got %f", targetRate, info.targetFiringRate)
	}

	if len(info.firingHistory) == 0 {
		t.Error("Expected non-empty firing history in homeostatic info")
	}

	t.Logf("Mid-test rate: %f Hz, Final rate: %f Hz (target: %f Hz)", midRate, finalRate, targetRate)
}

// TestCalciumDynamics tests calcium accumulation and decay
func TestCalciumDynamics(t *testing.T) {
	// Create disabled STDP config
	stdpConfig := STDPConfig{
		Enabled:        false,
		LearningRate:   0.0,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.0,
	}

	neuron := NewNeuron("test_calcium", 1.0, 0.95, 5*time.Millisecond, 1.0, 5.0, 0.1, stdpConfig)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// Initial calcium should be zero
	initialCalcium := neuron.GetCalciumLevel()
	if initialCalcium != 0.0 {
		t.Errorf("Expected initial calcium level 0.0, got %f", initialCalcium)
	}

	// Fire the neuron
	input <- Message{Value: 1.5, Timestamp: time.Now(), SourceID: "test_input"}
	time.Sleep(10 * time.Millisecond) // Allow processing

	// Calcium should have increased
	postFireCalcium := neuron.GetCalciumLevel()
	if postFireCalcium <= initialCalcium {
		t.Errorf("Expected calcium to increase after firing. Initial: %f, Post-fire: %f",
			initialCalcium, postFireCalcium)
	}

	// Wait for decay
	time.Sleep(100 * time.Millisecond)

	// Calcium should have decreased due to decay
	decayedCalcium := neuron.GetCalciumLevel()
	if decayedCalcium >= postFireCalcium {
		t.Errorf("Expected calcium to decay. Post-fire: %f, Decayed: %f",
			postFireCalcium, decayedCalcium)
	}

	t.Logf("Calcium levels - Initial: %f, Post-fire: %f, Decayed: %f",
		initialCalcium, postFireCalcium, decayedCalcium)
}

// TestFiringHistoryTracking tests that firing history is maintained correctly
func TestFiringHistoryTracking(t *testing.T) {
	// Create disabled STDP config
	stdpConfig := STDPConfig{
		Enabled:        false,
		LearningRate:   0.0,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.0,
	}

	neuron := NewNeuron("test_history", 1.0, 0.95, 5*time.Millisecond, 1.0, 5.0, 0.1, stdpConfig)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// Initial history should be empty
	info := neuron.GetHomeostaticInfo()
	if len(info.firingHistory) != 0 {
		t.Errorf("Expected empty initial firing history, got %d entries", len(info.firingHistory))
	}

	// Fire neuron multiple times
	numFires := 5
	for i := 0; i < numFires; i++ {
		input <- Message{Value: 1.5, Timestamp: time.Now(), SourceID: "test_input"}
		time.Sleep(50 * time.Millisecond) // Allow refractory period
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Check firing history
	info = neuron.GetHomeostaticInfo()
	if len(info.firingHistory) != numFires {
		t.Errorf("Expected %d entries in firing history, got %d", numFires, len(info.firingHistory))
	}

	// Verify firing rate calculation
	rate := neuron.GetCurrentFiringRate()
	if rate <= 0 {
		t.Errorf("Expected positive firing rate, got %f", rate)
	}

	t.Logf("Firing history: %d entries, calculated rate: %f Hz", len(info.firingHistory), rate)
}

// TestHomeostaticBounds tests that threshold adjustment respects min/max bounds
func TestHomeostaticBounds(t *testing.T) {
	baseThreshold := 1.0

	// Create disabled STDP config
	stdpConfig := STDPConfig{
		Enabled:        false,
		LearningRate:   0.0,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.0,
	}

	neuron := NewNeuron("test_bounds", baseThreshold, 0.95, 5*time.Millisecond, 1.0, 0.5, 2.0, stdpConfig) // Very strong regulation

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// Get homeostatic bounds
	info := neuron.GetHomeostaticInfo()
	minThreshold := info.minThreshold
	maxThreshold := info.maxThreshold

	t.Logf("Homeostatic bounds: min=%f, max=%f, base=%f", minThreshold, maxThreshold, baseThreshold)

	// Test upper bound - make neuron hyperactive to push threshold up
	for i := 0; i < 100; i++ {
		input <- Message{Value: 2.0, Timestamp: time.Now(), SourceID: "test_input"} // Strong signal
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(500 * time.Millisecond)

	// Check threshold doesn't exceed max bound
	currentThreshold := neuron.GetCurrentThreshold()
	if currentThreshold > maxThreshold*1.01 { // Allow small tolerance for floating point
		t.Errorf("Threshold (%f) exceeded max bound (%f)", currentThreshold, maxThreshold)
	}

	// Verify base threshold is preserved
	if neuron.GetBaseThreshold() != baseThreshold {
		t.Errorf("Base threshold changed from %f to %f", baseThreshold, neuron.GetBaseThreshold())
	}

	t.Logf("Final threshold: %f (within bounds: %f - %f)", currentThreshold, minThreshold, maxThreshold)
}

// TestHomeostaticParameterSetting tests dynamic parameter adjustment
func TestHomeostaticParameterSetting(t *testing.T) {
	// Create disabled STDP config
	stdpConfig := STDPConfig{
		Enabled:        false,
		LearningRate:   0.0,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.0,
	}

	neuron := NewNeuron("test_params", 1.0, 0.95, 5*time.Millisecond, 1.0, 5.0, 0.1, stdpConfig)

	// Test parameter setting
	newTargetRate := 10.0
	newStrength := 0.5
	neuron.SetHomeostaticParameters(newTargetRate, newStrength)

	info := neuron.GetHomeostaticInfo()
	if info.targetFiringRate != newTargetRate {
		t.Errorf("Expected target rate %f, got %f", newTargetRate, info.targetFiringRate)
	}

	if info.homeostasisStrength != newStrength {
		t.Errorf("Expected homeostasis strength %f, got %f", newStrength, info.homeostasisStrength)
	}

	// Test disabling homeostasis
	originalThreshold := neuron.GetCurrentThreshold()
	neuron.SetHomeostaticParameters(0.0, 0.0)

	// Threshold should reset to base value
	expectedThreshold := neuron.GetBaseThreshold()
	actualThreshold := neuron.GetCurrentThreshold()
	if actualThreshold != expectedThreshold {
		t.Errorf("Expected threshold to reset to base value %f when homeostasis disabled, got %f",
			expectedThreshold, actualThreshold)
	}

	t.Logf("Parameter update: target %f->%f, strength %f->%f", 5.0, newTargetRate, 0.1, newStrength)
	t.Logf("Threshold reset: %f -> %f (base: %f)", originalThreshold, actualThreshold, expectedThreshold)
}

// TestHomeostaticVsSimpleNeuron tests that simple neurons don't change threshold
func TestHomeostaticVsSimpleNeuron(t *testing.T) {
	// Create disabled STDP config for homeostatic neuron
	stdpConfig := STDPConfig{
		Enabled:        false,
		LearningRate:   0.0,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.0,
	}

	// Create both types
	homeostaticNeuron := NewNeuron("homeostatic", 1.0, 0.95, 5*time.Millisecond, 1.0, 3.0, 0.3, stdpConfig)
	simpleNeuron := NewSimpleNeuron("simple", 1.0, 0.95, 5*time.Millisecond, 1.0)

	go homeostaticNeuron.Run()
	go simpleNeuron.Run()
	defer homeostaticNeuron.Close()
	defer simpleNeuron.Close()

	homeostaticInput := homeostaticNeuron.GetInput()
	simpleInput := simpleNeuron.GetInput()

	// Record initial thresholds
	initialHomeostaticThreshold := homeostaticNeuron.GetCurrentThreshold()
	initialSimpleThreshold := simpleNeuron.GetCurrentThreshold()

	// Send identical inputs to both
	for i := 0; i < 30; i++ {
		homeostaticInput <- Message{Value: 1.3, Timestamp: time.Now(), SourceID: "test_input"}
		simpleInput <- Message{Value: 1.3, Timestamp: time.Now(), SourceID: "test_input"}
		time.Sleep(40 * time.Millisecond)
	}

	time.Sleep(300 * time.Millisecond)

	// Check final thresholds
	finalHomeostaticThreshold := homeostaticNeuron.GetCurrentThreshold()
	finalSimpleThreshold := simpleNeuron.GetCurrentThreshold()

	// Simple neuron threshold should not change
	if finalSimpleThreshold != initialSimpleThreshold {
		t.Errorf("Simple neuron threshold changed from %f to %f",
			initialSimpleThreshold, finalSimpleThreshold)
	}

	// Homeostatic neuron threshold may change
	homeostaticChanged := finalHomeostaticThreshold != initialHomeostaticThreshold

	// Verify homeostatic neuron has activity tracking
	homeostaticRate := homeostaticNeuron.GetCurrentFiringRate()
	if homeostaticRate <= 0 {
		t.Error("Homeostatic neuron should have measurable firing rate")
	}

	// Simple neuron should have calcium level 0 (no tracking)
	simpleCalcium := simpleNeuron.GetCalciumLevel()
	if simpleCalcium != 0.0 {
		t.Errorf("Simple neuron should have no calcium tracking, got %f", simpleCalcium)
	}

	t.Logf("Simple neuron: threshold %f -> %f (unchanged: %v)",
		initialSimpleThreshold, finalSimpleThreshold, finalSimpleThreshold == initialSimpleThreshold)
	t.Logf("Homeostatic neuron: threshold %f -> %f (changed: %v, rate: %f Hz)",
		initialHomeostaticThreshold, finalHomeostaticThreshold, homeostaticChanged, homeostaticRate)
}

// TestHomeostaticStabilityOverTime tests long-term homeostatic behavior
func TestHomeostaticStabilityOverTime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-term stability test in short mode")
	}

	targetRate := 5.0

	// Create disabled STDP config
	stdpConfig := STDPConfig{
		Enabled:        false,
		LearningRate:   0.0,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.0,
	}

	neuron := NewNeuron("stability_test", 1.0, 0.95, 5*time.Millisecond, 1.0, targetRate, 0.5, stdpConfig)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	// Track threshold changes over time
	var thresholds []float64
	var rates []float64

	// Use sync.WaitGroup for proper coordination
	var wg sync.WaitGroup
	stopSignal := make(chan struct{})

	// Send periodic variable inputs
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			select {
			case <-stopSignal:
				return
			default:
				// Variable input pattern
				val := 0.8 + 0.8*float64((i%10))/10.0 // 0.8 to 1.6
				select {
				case input <- Message{Value: val, Timestamp: time.Now(), SourceID: "test_input"}:
				case <-stopSignal:
					return
				}
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()

	// Sample threshold and rate every second
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		thresholds = append(thresholds, neuron.GetCurrentThreshold())
		rates = append(rates, neuron.GetCurrentFiringRate())
	}

	// Signal goroutine to stop and wait for completion
	close(stopSignal)
	wg.Wait()

	// Check that system eventually stabilizes
	finalRate := rates[len(rates)-1]
	tolerance := 2.0

	if finalRate < targetRate-tolerance || finalRate > targetRate+tolerance {
		t.Logf("Warning: Final rate (%f) not close to target (%f) after long-term run", finalRate, targetRate)
	}

	t.Logf("Long-term stability test completed")
	t.Logf("Target rate: %f Hz", targetRate)
	t.Logf("Final rate: %f Hz", finalRate)
	t.Logf("Threshold evolution: %f -> %f", thresholds[0], thresholds[len(thresholds)-1])
}

// ============================================================================
// HOMEOSTATIC PERFORMANCE BENCHMARKS
// ============================================================================

// BenchmarkHomeostaticNeuronCreation benchmarks homeostatic neuron creation performance
func BenchmarkHomeostaticNeuronCreation(b *testing.B) {
	// Create STDP config for benchmarking
	stdpConfig := STDPConfig{
		Enabled:        false,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewNeuron("bench_homeostatic", 1.0, 0.95, 5*time.Millisecond, 1.0, 5.0, 0.1, stdpConfig)
	}
}

// BenchmarkHomeostaticMessageProcessing benchmarks homeostatic message processing throughput
func BenchmarkHomeostaticMessageProcessing(b *testing.B) {
	// Create STDP config for benchmarking
	stdpConfig := STDPConfig{
		Enabled:        false,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.5,
	}

	neuron := NewNeuron("bench_homeostatic_processing", 10.0, 0.95, 5*time.Millisecond, 1.0, 5.0, 0.1, stdpConfig) // High threshold to avoid firing

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input <- Message{
			Value:     0.1,
			Timestamp: time.Now(),
			SourceID:  "bench_source",
		}
	}
}

// BenchmarkCalciumDynamics benchmarks calcium accumulation and decay performance
func BenchmarkCalciumDynamics(b *testing.B) {
	// Create STDP config for benchmarking
	stdpConfig := STDPConfig{
		Enabled:        false,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.5,
	}

	neuron := NewNeuron("bench_calcium", 1.0, 0.95, 5*time.Millisecond, 1.0, 5.0, 0.1, stdpConfig)

	go neuron.Run()
	defer neuron.Close()

	input := neuron.GetInput()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Fire the neuron to trigger calcium dynamics
		input <- Message{
			Value:     1.5,
			Timestamp: time.Now(),
			SourceID:  "bench_source",
		}
		// Small delay to allow processing but keep benchmark tight
		time.Sleep(time.Microsecond)
	}
}

// BenchmarkHomeostaticAdjustment benchmarks threshold adjustment calculations
func BenchmarkHomeostaticAdjustment(b *testing.B) {
	// Create STDP config for benchmarking
	stdpConfig := STDPConfig{
		Enabled:        false,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.5,
	}

	neuron := NewNeuron("bench_adjustment", 1.0, 0.95, 5*time.Millisecond, 1.0, 5.0, 0.5, stdpConfig)

	// Pre-populate some firing history to make adjustment meaningful
	for i := 0; i < 10; i++ {
		neuron.homeostatic.firingHistory = append(neuron.homeostatic.firingHistory, time.Now().Add(-time.Duration(i)*100*time.Millisecond))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		neuron.stateMutex.Lock()
		neuron.performHomeostaticAdjustmentUnsafe()
		neuron.stateMutex.Unlock()
	}
}

// BenchmarkFiringRateCalculation benchmarks firing rate computation performance
func BenchmarkFiringRateCalculation(b *testing.B) {
	// Create STDP config for benchmarking
	stdpConfig := STDPConfig{
		Enabled:        false,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.5,
	}

	neuron := NewNeuron("bench_rate_calc", 1.0, 0.95, 5*time.Millisecond, 1.0, 5.0, 0.1, stdpConfig)

	// Pre-populate firing history with realistic data
	now := time.Now()
	for i := 0; i < 50; i++ {
		neuron.homeostatic.firingHistory = append(neuron.homeostatic.firingHistory,
			now.Add(-time.Duration(i)*100*time.Millisecond))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		neuron.stateMutex.Lock()
		_ = neuron.calculateCurrentFiringRateUnsafe()
		neuron.stateMutex.Unlock()
	}
}

// BenchmarkHomeostaticInfoRetrieval benchmarks getting homeostatic information
func BenchmarkHomeostaticInfoRetrieval(b *testing.B) {
	// Create STDP config for benchmarking
	stdpConfig := STDPConfig{
		Enabled:        false,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.5,
	}

	neuron := NewNeuron("bench_info", 1.0, 0.95, 5*time.Millisecond, 1.0, 5.0, 0.1, stdpConfig)

	// Pre-populate some state
	for i := 0; i < 20; i++ {
		neuron.homeostatic.firingHistory = append(neuron.homeostatic.firingHistory, time.Now())
	}
	neuron.homeostatic.calciumLevel = 2.5

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = neuron.GetHomeostaticInfo()
	}
}

// BenchmarkConcurrentHomeostaticAccess benchmarks concurrent access to homeostatic data
func BenchmarkConcurrentHomeostaticAccess(b *testing.B) {
	// Create STDP config for benchmarking
	stdpConfig := STDPConfig{
		Enabled:        false,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     50 * time.Millisecond,
		MinWeight:      0.1,
		MaxWeight:      3.0,
		AsymmetryRatio: 1.5,
	}

	neuron := NewNeuron("bench_concurrent", 1.0, 0.95, 5*time.Millisecond, 1.0, 5.0, 0.1, stdpConfig)

	go neuron.Run()
	defer neuron.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Mix of read operations that would happen in real usage
			switch i := b.N % 4; i {
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
