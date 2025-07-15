package integration

import (
	"fmt"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/extracellular"
	"github.com/SynapticNetworks/temporal-neuron/neuron"
	"github.com/SynapticNetworks/temporal-neuron/synapse"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// TestXORErrorDrivenLearning tests XOR learning using error-driven mechanisms
func TestXORErrorDrivenLearning(t *testing.T) {
	t.Log("=== XOR ERROR-DRIVEN LEARNING TEST ===")
	t.Log("Testing XOR learning using eligibility traces, GABA penalty signals, and error-driven dopamine")

	// Create matrix with full capabilities
	matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  1 * time.Millisecond,
		MaxComponents:   100,
	})

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}
	defer matrix.Stop()

	// Register error-driven neuron types
	registerErrorDrivenNeuronTypes(matrix, t)
	registerErrorDrivenSynapseTypes(matrix, t)

	// Build error-driven XOR circuit
	circuit := buildErrorDrivenXORCircuit(matrix, t)
	defer cleanupErrorDrivenCircuit(circuit)

	// Define XOR patterns for error-driven learning
	patterns := []XORPattern{
		{name: "00", bits: []int{0, 0}, expected: 0},
		{name: "01", bits: []int{0, 1}, expected: 1},
		{name: "10", bits: []int{1, 0}, expected: 1},
		{name: "11", bits: []int{1, 1}, expected: 0},
	}

	// Run error-driven XOR learning
	results := runErrorDrivenXORLearning(matrix, circuit, patterns, t)

	// Report results
	t.Logf("\n--- Error-Driven XOR Learning Results ---")
	t.Logf("Training Accuracy: %.1f%%", results.TrainingAccuracy)
	t.Logf("Test Accuracy: %.1f%% (%d/%d correct)", results.TestAccuracy, results.CorrectResponses, results.TotalTests)

	// Summary and analysis
	t.Log("\n=== ERROR-DRIVEN XOR LEARNING ANALYSIS ===")
	if results.TestAccuracy >= 75.0 {
		t.Log("✅ XOR learning successful! Error-driven mechanisms worked")
		if results.TestAccuracy >= 90.0 {
			t.Log("🎉 Excellent error-driven learning performance")
		}
	} else if results.TestAccuracy >= 50.0 {
		t.Log("⚠️  Partial XOR learning - some error-driven effect observed")
	} else {
		t.Log("❌ XOR learning failed - error-driven mechanisms insufficient")
	}

	t.Log("Learning mechanisms used:")
	t.Log("- Eligibility traces for temporal credit assignment")
	t.Log("- GABA penalty signals for incorrect responses")
	t.Log("- Error-driven dopamine (negative for errors, positive for rewards)")
	t.Log("- Back-propagating action potentials for feedback")
	t.Log("- STDP with feedback-driven learning rates")

	if results.TestAccuracy >= 75.0 {
		t.Log("\n🧠 Biological Significance:")
		t.Log("- Eligibility traces enabled delayed error correction")
		t.Log("- GABA penalty signals weakened incorrect pathways")
		t.Log("- Error-driven dopamine provided bidirectional learning")
		t.Log("- Temporal credit assignment solved non-linear XOR problem")
		t.Log("- Architecture demonstrates biological error correction")
	}
}

// ErrorDrivenXORCircuit represents the error-driven XOR learning circuit
type ErrorDrivenXORCircuit struct {
	// Input layer
	inputA, inputB component.NeuralComponent

	// Hidden layer with error-driven neurons
	hiddenNeuron1, hiddenNeuron2 component.NeuralComponent

	// Output layer with competitive neurons
	outputXOR0, outputXOR1 component.NeuralComponent

	// Error signaling neurons
	errorNeuron    component.NeuralComponent // GABA neuron for penalty signals
	rewardNeuron   component.NeuralComponent // Dopamine neuron for reward signals

	// All synapses for weight analysis
	synapses []component.SynapticProcessor

	allNeurons []component.NeuralComponent
}

// registerErrorDrivenNeuronTypes creates neurons optimized for error-driven learning
func registerErrorDrivenNeuronTypes(matrix *extracellular.ExtracellularMatrix, t *testing.T) {
	// ERROR-DRIVEN HIDDEN NEURON: Enhanced for eligibility traces and backpropagation
	matrix.RegisterNeuronType("error_driven_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.95, 3*time.Millisecond, 1.5, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandDopamine, types.LigandGABA})
		n.SetCallbacks(callbacks)

		// Enable STDP feedback for error-driven learning
		n.EnableSTDPFeedback(10*time.Millisecond, 0.1) // 10ms feedback delay, 0.1 learning rate

		// Create dendritic mode with backpropagation gating
		dendriticMode := neuron.NewTemporalSummationMode()
		// Add channels for enhanced backpropagation
		dendriticMode.AddChannel(neuron.NewRealisticNavChannel("nav1.6_error"))
		dendriticMode.AddChannel(neuron.NewRealisticCavChannel("cav1.2_error")) // For bAP integration
		dendriticMode.AddChannel(neuron.NewRealisticKvChannel("kv4.2_error"))

		if err := n.SetDendriticMode(dendriticMode); err != nil {
			return nil, fmt.Errorf("failed to set dendritic mode: %v", err)
		}
		return n, nil
	})

	// ERROR SIGNALING NEURON: GABA neuron for penalty signals
	matrix.RegisterNeuronType("error_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.95, 2*time.Millisecond, 1.8, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandGABA})
		n.SetCallbacks(callbacks)

		// Optimized for fast error signaling
		dendriticMode := neuron.NewTemporalSummationMode()
		dendriticMode.AddChannel(neuron.NewRealisticNavChannel("nav1.6_error_fast"))
		dendriticMode.AddChannel(neuron.NewRealisticGabaAChannel("gabaa_error_1"))
		dendriticMode.AddChannel(neuron.NewRealisticGabaAChannel("gabaa_error_2"))

		if err := n.SetDendriticMode(dendriticMode); err != nil {
			return nil, fmt.Errorf("failed to set dendritic mode: %v", err)
		}
		return n, nil
	})

	// REWARD SIGNALING NEURON: Dopamine neuron for reward/error signals
	matrix.RegisterNeuronType("reward_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.95, 3*time.Millisecond, 1.5, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandDopamine})
		n.SetCallbacks(callbacks)
		return n, nil
	})

	// STANDARD INPUT/OUTPUT NEURON
	matrix.RegisterNeuronType("standard_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.95, 3*time.Millisecond, 1.5, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate})
		n.SetCallbacks(callbacks)
		return n, nil
	})

	t.Log("✓ Registered error-driven neuron types: Error-driven, Error, Reward, Standard")
}

// registerErrorDrivenSynapseTypes creates synapses with eligibility traces
func registerErrorDrivenSynapseTypes(matrix *extracellular.ExtracellularMatrix, t *testing.T) {
	matrix.RegisterSynapseType("error_driven_synapse", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, exists := matrix.GetNeuron(config.PresynapticID)
		if !exists {
			return nil, fmt.Errorf("presynaptic neuron not found: %s", config.PresynapticID)
		}
		postNeuron, exists := matrix.GetNeuron(config.PostsynapticID)
		if !exists {
			return nil, fmt.Errorf("postsynaptic neuron not found: %s", config.PostsynapticID)
		}

		// Create synapse with enhanced eligibility trace decay (longer for error-driven learning)
		syn := synapse.NewBasicSynapse(id, preNeuron, postNeuron,
			synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(),
			config.InitialWeight, config.Delay)

		// Configure longer eligibility trace for error-driven learning
		syn.SetEligibilityDecay(800 * time.Millisecond) // 800ms decay for temporal credit assignment

		return syn, nil
	})
}

// buildErrorDrivenXORCircuit creates the error-driven XOR learning circuit
func buildErrorDrivenXORCircuit(matrix *extracellular.ExtracellularMatrix, t *testing.T) *ErrorDrivenXORCircuit {
	circuit := &ErrorDrivenXORCircuit{}
	var err error

	t.Log("\n--- Building Error-Driven XOR Circuit ---")

	// Create input neurons
	circuit.inputA, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "standard_neuron", Threshold: 0.5})
	if err != nil {
		t.Fatalf("Failed to create inputA: %v", err)
	}
	circuit.inputB, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "standard_neuron", Threshold: 0.5})
	if err != nil {
		t.Fatalf("Failed to create inputB: %v", err)
	}

	// Create hidden neurons with error-driven capabilities
	circuit.hiddenNeuron1, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "error_driven_neuron", Threshold: 1.2})
	if err != nil {
		t.Fatalf("Failed to create hiddenNeuron1: %v", err)
	}
	circuit.hiddenNeuron2, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "error_driven_neuron", Threshold: 1.2})
	if err != nil {
		t.Fatalf("Failed to create hiddenNeuron2: %v", err)
	}

	// Create output neurons
	circuit.outputXOR0, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "standard_neuron", Threshold: 1.0})
	if err != nil {
		t.Fatalf("Failed to create outputXOR0: %v", err)
	}
	circuit.outputXOR1, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "standard_neuron", Threshold: 1.0})
	if err != nil {
		t.Fatalf("Failed to create outputXOR1: %v", err)
	}

	// Create error signaling neurons
	circuit.errorNeuron, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "error_neuron", Threshold: 0.8})
	if err != nil {
		t.Fatalf("Failed to create errorNeuron: %v", err)
	}
	circuit.rewardNeuron, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "reward_neuron", Threshold: 0.8})
	if err != nil {
		t.Fatalf("Failed to create rewardNeuron: %v", err)
	}

	// Collect all neurons
	circuit.allNeurons = []component.NeuralComponent{
		circuit.inputA, circuit.inputB,
		circuit.hiddenNeuron1, circuit.hiddenNeuron2,
		circuit.outputXOR0, circuit.outputXOR1,
		circuit.errorNeuron, circuit.rewardNeuron,
	}

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

	// Wire the error-driven circuit
	wireErrorDrivenXORCircuit(matrix, circuit, t)

	t.Log("✓ Error-driven XOR circuit created with eligibility traces")
	return circuit
}

// wireErrorDrivenXORCircuit creates the error-driven connectivity pattern
func wireErrorDrivenXORCircuit(matrix *extracellular.ExtracellularMatrix, circuit *ErrorDrivenXORCircuit, t *testing.T) {
	connections := []struct {
		pre, post component.NeuralComponent
		weight    float64
		desc      string
	}{
		// Input to hidden layer (learning layer)
		{circuit.inputA, circuit.hiddenNeuron1, 0.8, "inputA -> hidden1"},
		{circuit.inputA, circuit.hiddenNeuron2, 0.6, "inputA -> hidden2"},
		{circuit.inputB, circuit.hiddenNeuron1, 0.6, "inputB -> hidden1"},
		{circuit.inputB, circuit.hiddenNeuron2, 0.8, "inputB -> hidden2"},

		// Hidden to output layer (decision layer) - differentiated connections for XOR
		{circuit.hiddenNeuron1, circuit.outputXOR0, 1.2, "hidden1 -> XOR0"}, // Stronger for XOR=0
		{circuit.hiddenNeuron1, circuit.outputXOR1, 0.4, "hidden1 -> XOR1"}, // Weaker for XOR=1
		{circuit.hiddenNeuron2, circuit.outputXOR0, 0.4, "hidden2 -> XOR0"}, // Weaker for XOR=0  
		{circuit.hiddenNeuron2, circuit.outputXOR1, 1.2, "hidden2 -> XOR1"}, // Stronger for XOR=1

		// Competitive inhibition between outputs
		{circuit.outputXOR0, circuit.outputXOR1, 0.6, "XOR0 -> XOR1 (inhibition)"},
		{circuit.outputXOR1, circuit.outputXOR0, 0.6, "XOR1 -> XOR0 (inhibition)"},

		// Output to error/reward neurons (supervision layer)
		{circuit.outputXOR0, circuit.errorNeuron, 0.5, "XOR0 -> error"},
		{circuit.outputXOR1, circuit.errorNeuron, 0.5, "XOR1 -> error"},
		{circuit.outputXOR0, circuit.rewardNeuron, 0.5, "XOR0 -> reward"},
		{circuit.outputXOR1, circuit.rewardNeuron, 0.5, "XOR1 -> reward"},
	}

	circuit.synapses = make([]component.SynapticProcessor, 0)
	for _, conn := range connections {
		syn, err := matrix.CreateSynapse(types.SynapseConfig{
			SynapseType:    "error_driven_synapse",
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

	t.Logf("✓ Created %d error-driven synapses with eligibility traces", len(circuit.synapses))
}

// cleanupErrorDrivenCircuit stops all neurons
func cleanupErrorDrivenCircuit(circuit *ErrorDrivenXORCircuit) {
	for _, neuron := range circuit.allNeurons {
		neuron.Stop()
	}
}

// runErrorDrivenXORLearning implements error-driven XOR learning with eligibility traces
func runErrorDrivenXORLearning(matrix *extracellular.ExtracellularMatrix, circuit *ErrorDrivenXORCircuit, patterns []XORPattern, t *testing.T) *XORLearningResults {
	t.Log("\n--- Error-Driven XOR Learning ---")

	results := &XORLearningResults{}
	trainingEpochs := 50 // More epochs for error-driven learning
	correctTraining := 0
	totalTraining := 0

	// Training phase with error-driven learning
	for epoch := 0; epoch < trainingEpochs; epoch++ {
		if epoch%10 == 0 {
			t.Logf("Training epoch %d/%d", epoch, trainingEpochs)
		}

		for _, pattern := range patterns {
			totalTraining++

			// Clear network state
			time.Sleep(30 * time.Millisecond)

			// Present input pattern
			presentErrorDrivenPattern(circuit, pattern)

			// Allow forward propagation
			time.Sleep(20 * time.Millisecond)

			// Get network prediction
			predicted := checkErrorDrivenOutput(circuit)
			correct := predicted == pattern.expected

			if correct {
				correctTraining++
			}

			// Apply error-driven learning
			applyErrorDrivenLearning(matrix, circuit, pattern, predicted, t)

			// Allow error signal propagation and eligibility trace updates
			time.Sleep(25 * time.Millisecond)
		}
	}

	results.TrainingAccuracy = float64(correctTraining) / float64(totalTraining) * 100.0
	t.Logf("Training completed. Accuracy: %.1f%% (%d/%d correct)", 
		results.TrainingAccuracy, correctTraining, totalTraining)

	// Test phase without error signals
	t.Log("\n--- Testing Error-Driven XOR Function ---")
	testPatterns := []XORPattern{
		{name: "00-test", bits: []int{0, 0}, expected: 0},
		{name: "01-test", bits: []int{0, 1}, expected: 1},
		{name: "10-test", bits: []int{1, 0}, expected: 1},
		{name: "11-test", bits: []int{1, 1}, expected: 0},
	}

	correctTest := 0
	for _, pattern := range testPatterns {
		t.Logf("\nTesting pattern %s (bits: %v, expected XOR: %d)", pattern.name, pattern.bits, pattern.expected)

		// Reset network state
		time.Sleep(50 * time.Millisecond)

		// Present test pattern
		presentErrorDrivenPattern(circuit, pattern)

		// Wait for response
		time.Sleep(25 * time.Millisecond)

		// Check output
		predicted := checkErrorDrivenOutput(circuit)
		correct := predicted == pattern.expected

		if correct {
			correctTest++
		}

		status := "❌"
		if correct {
			status = "✅"
		}

		output0 := circuit.outputXOR0.GetActivityLevel()
		output1 := circuit.outputXOR1.GetActivityLevel()
		t.Logf("  Output XOR0: %.3f, Output XOR1: %.3f, Predicted: %d, Expected: %d %s",
			output0, output1, predicted, pattern.expected, status)
	}

	results.TestAccuracy = float64(correctTest) / float64(len(testPatterns)) * 100.0
	results.CorrectResponses = correctTest
	results.TotalTests = len(testPatterns)

	// Analyze learned weights
	analyzeErrorDrivenWeights(circuit, t)

	return results
}

// presentErrorDrivenPattern presents input pattern to the circuit
func presentErrorDrivenPattern(circuit *ErrorDrivenXORCircuit, pattern XORPattern) {
	if pattern.bits[0] == 1 {
		circuit.inputA.Receive(types.NeuralSignal{
			Value:     1.5,
			Timestamp: time.Now(),
			SourceID:  "error_driven_input",
			TargetID:  circuit.inputA.ID(),
		})
	}
	if pattern.bits[1] == 1 {
		circuit.inputB.Receive(types.NeuralSignal{
			Value:     1.5,
			Timestamp: time.Now(),
			SourceID:  "error_driven_input",
			TargetID:  circuit.inputB.ID(),
		})
	}
}

// checkErrorDrivenOutput determines output from competitive neurons
func checkErrorDrivenOutput(circuit *ErrorDrivenXORCircuit) int {
	output0 := circuit.outputXOR0.GetActivityLevel()
	output1 := circuit.outputXOR1.GetActivityLevel()

	// Winner-takes-all decision
	if output1 > output0 {
		return 1
	}
	return 0
}

// applyErrorDrivenLearning applies error-driven learning with eligibility traces
func applyErrorDrivenLearning(matrix *extracellular.ExtracellularMatrix, circuit *ErrorDrivenXORCircuit, pattern XORPattern, predicted int, t *testing.T) {
	correct := predicted == pattern.expected

	if correct {
		// Reward signal: Positive dopamine reinforces correct pathways
		rewardCorrectResponse(matrix, circuit, pattern, t)
	} else {
		// Error signal: GABA penalty + negative dopamine weakens incorrect pathways
		penalizeIncorrectResponse(matrix, circuit, pattern, predicted, t)
	}
}

// rewardCorrectResponse releases positive dopamine for correct responses
func rewardCorrectResponse(matrix *extracellular.ExtracellularMatrix, circuit *ErrorDrivenXORCircuit, pattern XORPattern, t *testing.T) {
	// Activate reward neuron
	circuit.rewardNeuron.Receive(types.NeuralSignal{
		Value:     2.0, // Strong reward signal
		Timestamp: time.Now(),
		SourceID:  "reward_system",
		TargetID:  circuit.rewardNeuron.ID(),
	})

	// Release positive dopamine (reward signal > 1.0)
	matrix.ReleaseLigand(types.LigandDopamine, circuit.hiddenNeuron1.ID(), 1.3) // Positive reward
	matrix.ReleaseLigand(types.LigandDopamine, circuit.hiddenNeuron2.ID(), 1.3)

	// Target the correct output path with stronger reward
	if pattern.expected == 0 {
		matrix.ReleaseLigand(types.LigandDopamine, circuit.outputXOR0.ID(), 1.4)
	} else {
		matrix.ReleaseLigand(types.LigandDopamine, circuit.outputXOR1.ID(), 1.4)
	}
}

// penalizeIncorrectResponse releases GABA penalty and negative dopamine for errors
func penalizeIncorrectResponse(matrix *extracellular.ExtracellularMatrix, circuit *ErrorDrivenXORCircuit, pattern XORPattern, predicted int, t *testing.T) {
	// Activate error neuron
	circuit.errorNeuron.Receive(types.NeuralSignal{
		Value:     2.0, // Strong error signal
		Timestamp: time.Now(),
		SourceID:  "error_system",
		TargetID:  circuit.errorNeuron.ID(),
	})

	// Release GABA penalty signal to hidden neurons
	matrix.ReleaseLigand(types.LigandGABA, circuit.hiddenNeuron1.ID(), 1.2) // GABA penalty
	matrix.ReleaseLigand(types.LigandGABA, circuit.hiddenNeuron2.ID(), 1.2)

	// Release negative dopamine (error signal < 1.0) to weaken incorrect pathway
	matrix.ReleaseLigand(types.LigandDopamine, circuit.hiddenNeuron1.ID(), 0.7) // Negative error signal
	matrix.ReleaseLigand(types.LigandDopamine, circuit.hiddenNeuron2.ID(), 0.7)

	// Target the incorrect output path with stronger penalty
	if predicted == 0 {
		matrix.ReleaseLigand(types.LigandGABA, circuit.outputXOR0.ID(), 1.3)
	} else {
		matrix.ReleaseLigand(types.LigandGABA, circuit.outputXOR1.ID(), 1.3)
	}
}

// analyzeErrorDrivenWeights analyzes learned synaptic weights
func analyzeErrorDrivenWeights(circuit *ErrorDrivenXORCircuit, t *testing.T) {
	t.Log("\n--- Learned Synaptic Weights Analysis ---")
	
	for i, syn := range circuit.synapses {
		if weightGetter, ok := syn.(interface{ GetWeight() float64 }); ok {
			weight := weightGetter.GetWeight()
			
			// Get eligibility trace if available
			eligibility := 0.0
			if traceGetter, ok := syn.(interface{ GetEligibilityTrace() float64 }); ok {
				eligibility = traceGetter.GetEligibilityTrace()
			}
			
			t.Logf("  Synapse[%d]: Weight=%.4f, Eligibility=%.6f", i, weight, eligibility)
		}
	}
}