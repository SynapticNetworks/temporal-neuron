package integration

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/extracellular"
	"github.com/SynapticNetworks/temporal-neuron/neuron"
	"github.com/SynapticNetworks/temporal-neuron/synapse"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// TestSystematicPatternRecognition tests pattern recognition with increasing complexity
func TestSystematicPatternRecognition(t *testing.T) {
	t.Log("=== SYSTEMATIC PATTERN RECOGNITION TEST ===")
	t.Log("Finding minimal requirements for 90%+ pattern memorization")

	// Test different pattern complexities
	complexities := []struct {
		name     string
		patterns []SimplePattern
		target   float64
	}{
		{
			name: "1-bit patterns",
			patterns: []SimplePattern{
				{name: "0", bits: []int{0}, expected: 0},
				{name: "1", bits: []int{1}, expected: 1},
			},
			target: 95.0,
		},
		{
			name: "2-bit patterns", 
			patterns: []SimplePattern{
				{name: "00", bits: []int{0, 0}, expected: 0},
				{name: "01", bits: []int{0, 1}, expected: 1},
				{name: "10", bits: []int{1, 0}, expected: 1},
				{name: "11", bits: []int{1, 1}, expected: 0},
			},
			target: 90.0,
		},
		{
			name: "3-bit patterns",
			patterns: []SimplePattern{
				{name: "000", bits: []int{0, 0, 0}, expected: 0},
				{name: "001", bits: []int{0, 0, 1}, expected: 1},
				{name: "010", bits: []int{0, 1, 0}, expected: 1},
				{name: "011", bits: []int{0, 1, 1}, expected: 0},
				{name: "100", bits: []int{1, 0, 0}, expected: 1},
				{name: "101", bits: []int{1, 0, 1}, expected: 0},
				{name: "110", bits: []int{1, 1, 0}, expected: 0},
				{name: "111", bits: []int{1, 1, 1}, expected: 1},
			},
			target: 85.0,
		},
		{
			name: "4-bit patterns",
			patterns: generate4BitPatterns(),
			target: 80.0,
		},
	}

	// Test different architectural configurations
	configs := []ArchConfig{
		{
			name:         "Basic",
			learningRate: 0.1,
			eligibility:  300 * time.Millisecond,
			memoryDecay:  3 * time.Millisecond,
		},
		{
			name:         "Enhanced",
			learningRate: 0.2,
			eligibility:  600 * time.Millisecond,
			memoryDecay:  6 * time.Millisecond,
		},
		{
			name:         "Strong",
			learningRate: 0.3,
			eligibility:  1200 * time.Millisecond,
			memoryDecay:  10 * time.Millisecond,
		},
		{
			name:         "Maximum",
			learningRate: 0.5,
			eligibility:  2000 * time.Millisecond,
			memoryDecay:  20 * time.Millisecond,
		},
	}

	results := make(map[string]map[string]float64)

	// Test each configuration against each complexity
	for _, config := range configs {
		results[config.name] = make(map[string]float64)
		
		t.Logf("\n" + strings.Repeat("=", 60))
		t.Logf("TESTING CONFIGURATION: %s", config.name)
		t.Logf("Learning Rate: %.1f, Eligibility: %v, Memory Decay: %v", 
			config.learningRate, config.eligibility, config.memoryDecay)
		t.Logf(strings.Repeat("=", 60))

		for _, complexity := range complexities {
			t.Logf("\n--- Testing %s ---", complexity.name)
			
			accuracy := testPatternRecognitionWithConfig(complexity.patterns, config, t)
			results[config.name][complexity.name] = accuracy
			
			status := "❌"
			if accuracy >= complexity.target {
				status = "✅"
			}
			
			t.Logf("Result: %.1f%% (target: %.1f%%) %s", accuracy, complexity.target, status)
		}
	}

	// Generate comprehensive results report
	generatePatternRecognitionReport(results, complexities, configs, t)
}

// SimplePattern represents a simple input-output pattern
type SimplePattern struct {
	name     string
	bits     []int
	expected int
}

// ArchConfig represents architectural configuration
type ArchConfig struct {
	name         string
	learningRate float64
	eligibility  time.Duration
	memoryDecay  time.Duration
}

// PatternRecognitionCircuit represents a simple pattern recognition circuit
type PatternRecognitionCircuit struct {
	inputNeurons  []component.NeuralComponent
	memoryNeurons []component.NeuralComponent
	hiddenNeurons []component.NeuralComponent
	outputNeurons []component.NeuralComponent
	allNeurons    []component.NeuralComponent
	synapses      []component.SynapticProcessor
}

// testPatternRecognitionWithConfig tests pattern recognition with specific configuration
func testPatternRecognitionWithConfig(patterns []SimplePattern, config ArchConfig, t *testing.T) float64 {
	// Create matrix
	matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  1 * time.Millisecond,
		MaxComponents:   200,
	})

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}
	defer matrix.Stop()

	// Register configurable neuron types
	registerConfigurableNeuronTypes(matrix, config, t)
	registerConfigurableSynapseTypes(matrix, config, t)

	// Build circuit based on pattern complexity
	circuit := buildPatternRecognitionCircuit(matrix, patterns, config, t)
	defer cleanupPatternCircuit(circuit)

	// Train the network
	trainingAccuracy := trainPatternRecognition(matrix, circuit, patterns, config, t)
	
	// Test recall
	recallAccuracy := testPatternRecall(circuit, patterns, t)

	t.Logf("  Training: %.1f%%, Recall: %.1f%%", trainingAccuracy, recallAccuracy)
	
	return recallAccuracy
}

// registerConfigurableNeuronTypes creates neurons with configurable parameters
func registerConfigurableNeuronTypes(matrix *extracellular.ExtracellularMatrix, config ArchConfig, t *testing.T) {
	// CONFIGURABLE NEURON: Adjustable learning parameters
	matrix.RegisterNeuronType("configurable_neuron", func(id string, neuronConfig types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, neuronConfig.Threshold, 0.95, config.memoryDecay, 1.5, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandDopamine, types.LigandGABA})
		n.SetCallbacks(callbacks)

		// Configure learning rate based on config
		n.EnableSTDPFeedback(5*time.Millisecond, config.learningRate)

		// Create dendritic mode with enhanced channels
		dendriticMode := neuron.NewTemporalSummationMode()
		dendriticMode.AddChannel(neuron.NewRealisticNavChannel("nav1.6_config"))
		dendriticMode.AddChannel(neuron.NewRealisticCavChannel("cav1.2_config"))
		dendriticMode.AddChannel(neuron.NewRealisticGabaAChannel("gabaa_config"))
		dendriticMode.AddChannel(neuron.NewRealisticKvChannel("kv4.2_config"))

		if err := n.SetDendriticMode(dendriticMode); err != nil {
			return nil, fmt.Errorf("failed to set dendritic mode: %v", err)
		}
		return n, nil
	})

	// MEMORY NEURON: Longer persistence based on config
	matrix.RegisterNeuronType("memory_neuron", func(id string, neuronConfig types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		memoryDecay := config.memoryDecay * 2 // Memory neurons persist longer
		n := neuron.NewNeuron(id, neuronConfig.Threshold, 0.90, memoryDecay, 1.2, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandDopamine})
		n.SetCallbacks(callbacks)

		// Enable learning for memory neurons too
		n.EnableSTDPFeedback(5*time.Millisecond, config.learningRate)

		// Enhanced memory channels
		dendriticMode := neuron.NewTemporalSummationMode()
		dendriticMode.AddChannel(neuron.NewRealisticNavChannel("nav1.6_memory"))
		dendriticMode.AddChannel(neuron.NewRealisticCavChannel("cav1.2_memory_1"))
		dendriticMode.AddChannel(neuron.NewRealisticCavChannel("cav1.2_memory_2"))
		dendriticMode.AddChannel(neuron.NewRealisticKvChannel("kv4.2_memory"))

		if err := n.SetDendriticMode(dendriticMode); err != nil {
			return nil, fmt.Errorf("failed to set dendritic mode: %v", err)
		}
		return n, nil
	})
}

// registerConfigurableSynapseTypes creates synapses with configurable eligibility traces
func registerConfigurableSynapseTypes(matrix *extracellular.ExtracellularMatrix, config ArchConfig, t *testing.T) {
	matrix.RegisterSynapseType("configurable_synapse", func(id string, synapseConfig types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, exists := matrix.GetNeuron(synapseConfig.PresynapticID)
		if !exists {
			return nil, fmt.Errorf("presynaptic neuron not found: %s", synapseConfig.PresynapticID)
		}
		postNeuron, exists := matrix.GetNeuron(synapseConfig.PostsynapticID)
		if !exists {
			return nil, fmt.Errorf("postsynaptic neuron not found: %s", synapseConfig.PostsynapticID)
		}

		syn := synapse.NewBasicSynapse(id, preNeuron, postNeuron,
			synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(),
			synapseConfig.InitialWeight, synapseConfig.Delay)

		// Configure eligibility trace based on config
		syn.SetEligibilityDecay(config.eligibility)

		return syn, nil
	})
}

// buildPatternRecognitionCircuit creates a circuit sized for the pattern complexity
func buildPatternRecognitionCircuit(matrix *extracellular.ExtracellularMatrix, patterns []SimplePattern, config ArchConfig, t *testing.T) *PatternRecognitionCircuit {
	circuit := &PatternRecognitionCircuit{}

	// Determine circuit size based on pattern complexity
	maxBits := 0
	for _, pattern := range patterns {
		if len(pattern.bits) > maxBits {
			maxBits = len(pattern.bits)
		}
	}

	numInputs := maxBits
	numMemory := maxBits  // One memory neuron per input
	numHidden := maxBits * 2 // More hidden neurons for complex patterns
	numOutputs := 2 // Binary output (0 or 1)

	t.Logf("Building circuit: %d inputs, %d memory, %d hidden, %d outputs", 
		numInputs, numMemory, numHidden, numOutputs)

	// Create input neurons
	for i := 0; i < numInputs; i++ {
		inputNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
			NeuronType: "configurable_neuron", 
			Threshold: 0.5,
		})
		if err != nil {
			t.Fatalf("Failed to create input neuron %d: %v", i, err)
		}
		circuit.inputNeurons = append(circuit.inputNeurons, inputNeuron)
	}

	// Create memory neurons
	for i := 0; i < numMemory; i++ {
		memoryNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
			NeuronType: "memory_neuron", 
			Threshold: 0.8,
		})
		if err != nil {
			t.Fatalf("Failed to create memory neuron %d: %v", i, err)
		}
		circuit.memoryNeurons = append(circuit.memoryNeurons, memoryNeuron)
	}

	// Create hidden neurons
	for i := 0; i < numHidden; i++ {
		hiddenNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
			NeuronType: "configurable_neuron", 
			Threshold: 1.0,
		})
		if err != nil {
			t.Fatalf("Failed to create hidden neuron %d: %v", i, err)
		}
		circuit.hiddenNeurons = append(circuit.hiddenNeurons, hiddenNeuron)
	}

	// Create output neurons
	for i := 0; i < numOutputs; i++ {
		outputNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
			NeuronType: "configurable_neuron", 
			Threshold: 1.2,
		})
		if err != nil {
			t.Fatalf("Failed to create output neuron %d: %v", i, err)
		}
		circuit.outputNeurons = append(circuit.outputNeurons, outputNeuron)
	}

	// Collect all neurons
	circuit.allNeurons = append(circuit.allNeurons, circuit.inputNeurons...)
	circuit.allNeurons = append(circuit.allNeurons, circuit.memoryNeurons...)
	circuit.allNeurons = append(circuit.allNeurons, circuit.hiddenNeurons...)
	circuit.allNeurons = append(circuit.allNeurons, circuit.outputNeurons...)

	// Start all neurons and register for chemical binding
	for _, neuron := range circuit.allNeurons {
		if err := neuron.Start(); err != nil {
			t.Fatalf("Failed to start neuron %s: %v", neuron.ID(), err)
		}
		if chemicalReceiver, ok := neuron.(component.ChemicalReceiver); ok {
			if err := matrix.RegisterForBinding(chemicalReceiver); err != nil {
				t.Fatalf("Failed to register %s for binding: %v", neuron.ID(), err)
			}
		}
	}

	// Wire the circuit
	wirePatternRecognitionCircuit(matrix, circuit, config, t)

	return circuit
}

// wirePatternRecognitionCircuit creates comprehensive connectivity
func wirePatternRecognitionCircuit(matrix *extracellular.ExtracellularMatrix, circuit *PatternRecognitionCircuit, config ArchConfig, t *testing.T) {
	connections := []struct {
		pre, post component.NeuralComponent
		weight    float64
		desc      string
	}{}

	// Input to memory (1:1 mapping)
	for i, input := range circuit.inputNeurons {
		if i < len(circuit.memoryNeurons) {
			connections = append(connections, struct {
				pre, post component.NeuralComponent
				weight    float64
				desc      string
			}{input, circuit.memoryNeurons[i], 0.8, fmt.Sprintf("input%d -> memory%d", i, i)})
		}
	}

	// Input to hidden (full connectivity)
	for i, input := range circuit.inputNeurons {
		for j, hidden := range circuit.hiddenNeurons {
			weight := 0.6 + float64((i+j)%3)*0.1 // Varied weights 0.6-0.8
			connections = append(connections, struct {
				pre, post component.NeuralComponent
				weight    float64
				desc      string
			}{input, hidden, weight, fmt.Sprintf("input%d -> hidden%d", i, j)})
		}
	}

	// Memory to hidden (full connectivity)
	for i, memory := range circuit.memoryNeurons {
		for j, hidden := range circuit.hiddenNeurons {
			weight := 0.5 + float64((i+j)%4)*0.1 // Varied weights 0.5-0.8
			connections = append(connections, struct {
				pre, post component.NeuralComponent
				weight    float64
				desc      string
			}{memory, hidden, weight, fmt.Sprintf("memory%d -> hidden%d", i, j)})
		}
	}

	// Hidden to output (full connectivity)
	for i, hidden := range circuit.hiddenNeurons {
		for j, output := range circuit.outputNeurons {
			weight := 0.7 + float64(j)*0.2 // Output0: 0.7, Output1: 0.9
			connections = append(connections, struct {
				pre, post component.NeuralComponent
				weight    float64
				desc      string
			}{hidden, output, weight, fmt.Sprintf("hidden%d -> output%d", i, j)})
		}
	}

	// Memory recurrence (for persistence)
	for i, memory := range circuit.memoryNeurons {
		connections = append(connections, struct {
			pre, post component.NeuralComponent
			weight    float64
			desc      string
		}{memory, memory, 0.3, fmt.Sprintf("memory%d recurrence", i)})
	}

	// Create synapses
	for _, conn := range connections {
		syn, err := matrix.CreateSynapse(types.SynapseConfig{
			SynapseType:    "configurable_synapse",
			PresynapticID:  conn.pre.ID(),
			PostsynapticID: conn.post.ID(),
			InitialWeight:  conn.weight,
			Delay:          1 * time.Millisecond,
		})
		if err != nil {
			t.Fatalf("Failed to create synapse %s: %v", conn.desc, err)
		}
		circuit.synapses = append(circuit.synapses, syn)
	}

	t.Logf("Created %d synapses", len(circuit.synapses))
}

// cleanupPatternCircuit stops all neurons
func cleanupPatternCircuit(circuit *PatternRecognitionCircuit) {
	for _, neuron := range circuit.allNeurons {
		neuron.Stop()
	}
}

// trainPatternRecognition trains the network on pattern recognition
func trainPatternRecognition(matrix *extracellular.ExtracellularMatrix, circuit *PatternRecognitionCircuit, patterns []SimplePattern, config ArchConfig, t *testing.T) float64 {
	correctTraining := 0
	totalTraining := 0
	epochs := 50 // More epochs for complex patterns

	for epoch := 0; epoch < epochs; epoch++ {
		for _, pattern := range patterns {
			totalTraining++

			// Present pattern
			presentPattern(circuit, pattern)
			time.Sleep(50 * time.Millisecond)

			// Get prediction
			predicted := getPatternOutput(circuit)
			correct := predicted == pattern.expected

			if correct {
				correctTraining++
			}

			// Apply supervised learning
			applyPatternSupervision(matrix, circuit, pattern.expected, predicted, config, t)
			time.Sleep(30 * time.Millisecond)

			// Reset between patterns
			time.Sleep(20 * time.Millisecond)
		}
	}

	return float64(correctTraining) / float64(totalTraining) * 100.0
}

// presentPattern presents a pattern to the network
func presentPattern(circuit *PatternRecognitionCircuit, pattern SimplePattern) {
	// Present each bit to corresponding input neuron
	for i, bit := range pattern.bits {
		if i < len(circuit.inputNeurons) && bit == 1 {
			circuit.inputNeurons[i].Receive(types.NeuralSignal{
				Value:     1.5,
				Timestamp: time.Now(),
				SourceID:  "pattern_input",
				TargetID:  circuit.inputNeurons[i].ID(),
			})
		}
	}
}

// getPatternOutput gets output from the network
func getPatternOutput(circuit *PatternRecognitionCircuit) int {
	if len(circuit.outputNeurons) >= 2 {
		output0 := circuit.outputNeurons[0].GetActivityLevel()
		output1 := circuit.outputNeurons[1].GetActivityLevel()
		
		if output1 > output0 {
			return 1
		}
	}
	return 0
}

// applyPatternSupervision applies supervised learning
func applyPatternSupervision(matrix *extracellular.ExtracellularMatrix, circuit *PatternRecognitionCircuit, expected int, predicted int, config ArchConfig, t *testing.T) {
	if expected != predicted {
		// Error correction
		for _, hidden := range circuit.hiddenNeurons {
			matrix.ReleaseLigand(types.LigandGABA, hidden.ID(), 0.8)
		}

		// Target signal
		if expected == 1 && len(circuit.outputNeurons) >= 2 {
			matrix.ReleaseLigand(types.LigandDopamine, circuit.outputNeurons[1].ID(), 1.5)
		}
	} else {
		// Positive reinforcement
		for _, hidden := range circuit.hiddenNeurons {
			matrix.ReleaseLigand(types.LigandDopamine, hidden.ID(), 1.2)
		}
	}
}

// testPatternRecall tests recall of trained patterns
func testPatternRecall(circuit *PatternRecognitionCircuit, patterns []SimplePattern, t *testing.T) float64 {
	correctRecall := 0
	totalRecall := len(patterns)

	for _, pattern := range patterns {
		// Present pattern
		presentPattern(circuit, pattern)
		time.Sleep(50 * time.Millisecond)

		// Get prediction
		predicted := getPatternOutput(circuit)
		correct := predicted == pattern.expected

		if correct {
			correctRecall++
		}

		// Reset between patterns
		time.Sleep(30 * time.Millisecond)
	}

	return float64(correctRecall) / float64(totalRecall) * 100.0
}

// generate4BitPatterns generates 4-bit XOR patterns
func generate4BitPatterns() []SimplePattern {
	patterns := []SimplePattern{}
	for i := 0; i < 16; i++ {
		bits := make([]int, 4)
		parity := 0
		for j := 0; j < 4; j++ {
			bits[j] = (i >> j) & 1
			parity ^= bits[j]
		}
		patterns = append(patterns, SimplePattern{
			name:     fmt.Sprintf("%04b", i),
			bits:     bits,
			expected: parity,
		})
	}
	return patterns
}

// generatePatternRecognitionReport generates comprehensive results
func generatePatternRecognitionReport(results map[string]map[string]float64, complexities []struct {
	name     string
	patterns []SimplePattern
	target   float64
}, configs []ArchConfig, t *testing.T) {
	t.Log("\n" + strings.Repeat("=", 80))
	t.Log("COMPREHENSIVE PATTERN RECOGNITION RESULTS")
	t.Log(strings.Repeat("=", 80))

	// Find best performing configuration for each complexity
	for _, complexity := range complexities {
		t.Logf("\n--- %s (Target: %.1f%%) ---", complexity.name, complexity.target)
		
		bestConfig := ""
		bestAccuracy := 0.0
		
		for _, config := range configs {
			accuracy := results[config.name][complexity.name]
			status := "❌"
			if accuracy >= complexity.target {
				status = "✅"
			}
			
			t.Logf("  %s: %.1f%% %s", config.name, accuracy, status)
			
			if accuracy > bestAccuracy {
				bestAccuracy = accuracy
				bestConfig = config.name
			}
		}
		
		if bestAccuracy >= complexity.target {
			t.Logf("  🎉 SUCCESS with %s configuration!", bestConfig)
		} else {
			t.Logf("  ❌ FAILED - Best was %s with %.1f%%", bestConfig, bestAccuracy)
		}
	}

	// Find optimal configuration
	t.Log("\n" + strings.Repeat("-", 40))
	t.Log("OPTIMAL CONFIGURATION ANALYSIS")
	t.Log(strings.Repeat("-", 40))

	configScores := make(map[string]int)
	for _, config := range configs {
		score := 0
		for _, complexity := range complexities {
			if results[config.name][complexity.name] >= complexity.target {
				score++
			}
		}
		configScores[config.name] = score
		t.Logf("%s: %d/%d targets achieved", config.name, score, len(complexities))
	}

	// Find best overall configuration
	bestOverallConfig := ""
	bestOverallScore := -1
	for config, score := range configScores {
		if score > bestOverallScore {
			bestOverallScore = score
			bestOverallConfig = config
		}
	}

	if bestOverallScore > 0 {
		t.Logf("\n🏆 OPTIMAL CONFIGURATION: %s", bestOverallConfig)
		for _, config := range configs {
			if config.name == bestOverallConfig {
				t.Logf("   Learning Rate: %.1f", config.learningRate)
				t.Logf("   Eligibility Trace: %v", config.eligibility)
				t.Logf("   Memory Decay: %v", config.memoryDecay)
			}
		}
	} else {
		t.Log("\n❌ NO CONFIGURATION ACHIEVED TARGETS")
		t.Log("Recommendations:")
		t.Log("- Increase learning rates further")
		t.Log("- Extend eligibility trace duration")
		t.Log("- Add more specialized neuron types")
		t.Log("- Increase network size")
	}

	t.Log("\n" + strings.Repeat("=", 80))
}