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

// TestSerialXORSupervisedLearning tests temporal XOR (parity) learning with supervised error backpropagation
func TestSerialXORSupervisedLearning(t *testing.T) {
	t.Log("=== SERIAL XOR SUPERVISED LEARNING TEST ===")
	t.Log("Testing temporal parity function with GABA error backpropagation and eligibility traces")

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

	// Register supervised learning neuron types
	registerSupervisedNeuronTypes(matrix, t)
	registerSupervisedSynapseTypes(matrix, t)

	// Build serial XOR circuit with temporal memory
	circuit := buildSerialXORCircuit(matrix, t)
	defer cleanupSerialCircuit(circuit)

	// Define temporal XOR sequences for supervised learning
	sequences := []TemporalXORSequence{
		{name: "seq1", bits: []int{1, 0, 1}, parities: []int{1, 1, 0}},     // Running XOR: 1, 1^0=1, 1^0^1=0
		{name: "seq2", bits: []int{0, 1, 1, 0}, parities: []int{0, 1, 0, 0}}, // Running XOR: 0, 0^1=1, 1^1=0, 0^0=0
		{name: "seq3", bits: []int{1, 1, 0, 1}, parities: []int{1, 0, 0, 1}}, // Running XOR: 1, 1^1=0, 0^0=0, 0^1=1
		{name: "seq4", bits: []int{0, 0, 1, 1, 1}, parities: []int{0, 0, 1, 0, 1}}, // 5-bit sequence
	}

	// Run supervised serial XOR learning
	results := runSupervisedSerialXORLearning(matrix, circuit, sequences, t)

	// Report results
	t.Logf("\n--- Supervised Serial XOR Learning Results ---")
	t.Logf("Training Accuracy: %.1f%%", results.TrainingAccuracy)
	t.Logf("Test Accuracy: %.1f%% (%d/%d correct)", results.TestAccuracy, results.CorrectResponses, results.TotalTests)

	// Summary and analysis
	t.Log("\n=== SUPERVISED SERIAL XOR LEARNING ANALYSIS ===")
	if results.TestAccuracy >= 75.0 {
		t.Log("✅ Serial XOR learning successful! Supervised error backpropagation worked")
		if results.TestAccuracy >= 90.0 {
			t.Log("🎉 Excellent temporal parity learning performance")
		}
	} else if results.TestAccuracy >= 50.0 {
		t.Log("⚠️  Partial XOR learning - some supervised effect observed")
	} else {
		t.Log("❌ Serial XOR learning failed - supervised backpropagation insufficient")
	}

	t.Log("Supervised learning mechanisms used:")
	t.Log("- GABA error backpropagation with specific error signals (target - predicted)")
	t.Log("- Eligibility traces for temporal credit assignment")
	t.Log("- Error-driven dopamine modulation based on prediction accuracy")
	t.Log("- Temporal memory neurons for sequence state maintenance")
	t.Log("- Known target supervision from XOR truth table")

	if results.TestAccuracy >= 75.0 {
		t.Log("\n🧠 Biological Significance:")
		t.Log("- GABA carried specific error information for supervised learning")
		t.Log("- Eligibility traces enabled temporal credit assignment over sequences")
		t.Log("- Error backpropagation adjusted weights toward correct parity computation")
		t.Log("- Temporal memory enabled sequential bit processing")
		t.Log("- Demonstrates biological supervised learning with neurotransmitter signaling")
	}
}

// TemporalXORSequence represents a sequence of bits with their running parity
type TemporalXORSequence struct {
	name     string
	bits     []int // Input bit sequence
	parities []int // Expected running XOR/parity at each step
}

// SerialXORCircuit represents the temporal XOR learning circuit
type SerialXORCircuit struct {
	// Input layer
	inputNeuron component.NeuralComponent // Single input for serial bits

	// Memory layer - maintains temporal state
	memoryNeuron1 component.NeuralComponent // Previous bit memory
	memoryNeuron2 component.NeuralComponent // Running parity memory

	// Hidden layer - processes current input + memory
	hiddenNeuron1 component.NeuralComponent // Temporal integration
	hiddenNeuron2 component.NeuralComponent // Parity computation

	// Output layer
	outputNeuron component.NeuralComponent // Current parity output

	// Supervision layer
	errorNeuron    component.NeuralComponent // GABA error backpropagation
	teacherNeuron  component.NeuralComponent // Target signal neuron

	// All synapses for weight analysis
	synapses []component.SynapticProcessor

	allNeurons []component.NeuralComponent
}

// SupervisedXORResults stores results from supervised learning
type SupervisedXORResults struct {
	TrainingAccuracy float64
	TestAccuracy     float64
	CorrectResponses int
	TotalTests       int
	WeightChanges    []float64 // Track weight evolution
}

// registerSupervisedNeuronTypes creates neurons optimized for supervised learning
func registerSupervisedNeuronTypes(matrix *extracellular.ExtracellularMatrix, t *testing.T) {
	// SUPERVISED LEARNING NEURON: Enhanced for error backpropagation
	matrix.RegisterNeuronType("supervised_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.95, 3*time.Millisecond, 1.5, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandDopamine, types.LigandGABA})
		n.SetCallbacks(callbacks)

		// Enable STDP feedback with higher learning rate for supervised learning
		n.EnableSTDPFeedback(5*time.Millisecond, 0.2) // 5ms feedback delay, 0.2 learning rate

		// Create dendritic mode with enhanced error sensitivity
		dendriticMode := neuron.NewTemporalSummationMode()
		dendriticMode.AddChannel(neuron.NewRealisticNavChannel("nav1.6_supervised"))
		dendriticMode.AddChannel(neuron.NewRealisticCavChannel("cav1.2_supervised")) // For temporal memory
		dendriticMode.AddChannel(neuron.NewRealisticGabaAChannel("gabaa_supervised")) // For error signals
		dendriticMode.AddChannel(neuron.NewRealisticKvChannel("kv4.2_supervised"))

		if err := n.SetDendriticMode(dendriticMode); err != nil {
			return nil, fmt.Errorf("failed to set dendritic mode: %v", err)
		}
		return n, nil
	})

	// MEMORY NEURON: Enhanced for temporal state maintenance
	matrix.RegisterNeuronType("memory_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.90, 10*time.Millisecond, 1.2, 0.0, 0.0) // Longer decay for memory
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandDopamine})
		n.SetCallbacks(callbacks)

		// Optimized for persistent activity
		dendriticMode := neuron.NewTemporalSummationMode()
		dendriticMode.AddChannel(neuron.NewRealisticNavChannel("nav1.6_memory"))
		dendriticMode.AddChannel(neuron.NewRealisticCavChannel("cav1.2_memory_1")) // High calcium for persistence
		dendriticMode.AddChannel(neuron.NewRealisticCavChannel("cav1.2_memory_2"))
		dendriticMode.AddChannel(neuron.NewRealisticKvChannel("kv4.2_memory"))

		if err := n.SetDendriticMode(dendriticMode); err != nil {
			return nil, fmt.Errorf("failed to set dendritic mode: %v", err)
		}
		return n, nil
	})

	// ERROR BACKPROPAGATION NEURON: Specialized for GABA error signaling
	matrix.RegisterNeuronType("error_backprop_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.95, 2*time.Millisecond, 1.8, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandGABA})
		n.SetCallbacks(callbacks)

		// Optimized for error signal propagation
		dendriticMode := neuron.NewTemporalSummationMode()
		dendriticMode.AddChannel(neuron.NewRealisticNavChannel("nav1.6_error_fast"))
		dendriticMode.AddChannel(neuron.NewRealisticGabaAChannel("gabaa_error_1"))
		dendriticMode.AddChannel(neuron.NewRealisticGabaAChannel("gabaa_error_2"))
		dendriticMode.AddChannel(neuron.NewRealisticGabaAChannel("gabaa_error_3")) // High GABA sensitivity

		if err := n.SetDendriticMode(dendriticMode); err != nil {
			return nil, fmt.Errorf("failed to set dendritic mode: %v", err)
		}
		return n, nil
	})

	// STANDARD INPUT/OUTPUT NEURON
	matrix.RegisterNeuronType("standard_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.95, 3*time.Millisecond, 1.5, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate})
		n.SetCallbacks(callbacks)
		return n, nil
	})

	t.Log("✓ Registered supervised learning neuron types: Supervised, Memory, Error-Backprop, Standard")
}

// registerSupervisedSynapseTypes creates synapses with enhanced eligibility traces for supervision
func registerSupervisedSynapseTypes(matrix *extracellular.ExtracellularMatrix, t *testing.T) {
	matrix.RegisterSynapseType("supervised_synapse", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, exists := matrix.GetNeuron(config.PresynapticID)
		if !exists {
			return nil, fmt.Errorf("presynaptic neuron not found: %s", config.PresynapticID)
		}
		postNeuron, exists := matrix.GetNeuron(config.PostsynapticID)
		if !exists {
			return nil, fmt.Errorf("postsynaptic neuron not found: %s", config.PostsynapticID)
		}

		// Create synapse with very long eligibility trace for temporal sequences
		syn := synapse.NewBasicSynapse(id, preNeuron, postNeuron,
			synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(),
			config.InitialWeight, config.Delay)

		// Configure long eligibility trace for temporal credit assignment in sequences
		syn.SetEligibilityDecay(1200 * time.Millisecond) // 1.2s decay for long sequences

		return syn, nil
	})
}

// buildSerialXORCircuit creates the temporal XOR circuit with memory
func buildSerialXORCircuit(matrix *extracellular.ExtracellularMatrix, t *testing.T) *SerialXORCircuit {
	circuit := &SerialXORCircuit{}
	var err error

	t.Log("\n--- Building Serial XOR Circuit with Temporal Memory ---")

	// Create single input neuron for serial bit stream
	circuit.inputNeuron, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "standard_neuron", Threshold: 0.5})
	if err != nil {
		t.Fatalf("Failed to create inputNeuron: %v", err)
	}

	// Create memory neurons for temporal state
	circuit.memoryNeuron1, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "memory_neuron", Threshold: 0.8})
	if err != nil {
		t.Fatalf("Failed to create memoryNeuron1: %v", err)
	}
	circuit.memoryNeuron2, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "memory_neuron", Threshold: 0.8})
	if err != nil {
		t.Fatalf("Failed to create memoryNeuron2: %v", err)
	}

	// Create hidden neurons for temporal processing
	circuit.hiddenNeuron1, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "supervised_neuron", Threshold: 1.0})
	if err != nil {
		t.Fatalf("Failed to create hiddenNeuron1: %v", err)
	}
	circuit.hiddenNeuron2, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "supervised_neuron", Threshold: 1.0})
	if err != nil {
		t.Fatalf("Failed to create hiddenNeuron2: %v", err)
	}

	// Create output neuron
	circuit.outputNeuron, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "supervised_neuron", Threshold: 0.9})
	if err != nil {
		t.Fatalf("Failed to create outputNeuron: %v", err)
	}

	// Create supervision neurons
	circuit.errorNeuron, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "error_backprop_neuron", Threshold: 0.6})
	if err != nil {
		t.Fatalf("Failed to create errorNeuron: %v", err)
	}
	circuit.teacherNeuron, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "standard_neuron", Threshold: 0.5})
	if err != nil {
		t.Fatalf("Failed to create teacherNeuron: %v", err)
	}

	// Collect all neurons
	circuit.allNeurons = []component.NeuralComponent{
		circuit.inputNeuron,
		circuit.memoryNeuron1, circuit.memoryNeuron2,
		circuit.hiddenNeuron1, circuit.hiddenNeuron2,
		circuit.outputNeuron,
		circuit.errorNeuron, circuit.teacherNeuron,
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

	// Wire the temporal circuit
	wireSerialXORCircuit(matrix, circuit, t)

	t.Log("✓ Serial XOR circuit created with temporal memory and supervised learning")
	return circuit
}

// wireSerialXORCircuit creates the temporal connectivity pattern
func wireSerialXORCircuit(matrix *extracellular.ExtracellularMatrix, circuit *SerialXORCircuit, t *testing.T) {
	connections := []struct {
		pre, post component.NeuralComponent
		weight    float64
		desc      string
	}{
		// Input to memory layer (temporal state maintenance)
		{circuit.inputNeuron, circuit.memoryNeuron1, 0.9, "input -> memory1 (bit memory)"},
		{circuit.inputNeuron, circuit.memoryNeuron2, 0.7, "input -> memory2 (parity memory)"},

		// Memory recurrence (temporal persistence)
		{circuit.memoryNeuron1, circuit.memoryNeuron1, 0.3, "memory1 -> memory1 (recurrence)"},
		{circuit.memoryNeuron2, circuit.memoryNeuron2, 0.4, "memory2 -> memory2 (recurrence)"},

		// Input + Memory to hidden layer (temporal integration)
		{circuit.inputNeuron, circuit.hiddenNeuron1, 0.8, "input -> hidden1"},
		{circuit.inputNeuron, circuit.hiddenNeuron2, 0.6, "input -> hidden2"},
		{circuit.memoryNeuron1, circuit.hiddenNeuron1, 0.7, "memory1 -> hidden1"},
		{circuit.memoryNeuron1, circuit.hiddenNeuron2, 0.5, "memory1 -> hidden2"},
		{circuit.memoryNeuron2, circuit.hiddenNeuron1, 0.5, "memory2 -> hidden1"},
		{circuit.memoryNeuron2, circuit.hiddenNeuron2, 0.8, "memory2 -> hidden2"},

		// Hidden to output (parity computation)
		{circuit.hiddenNeuron1, circuit.outputNeuron, 0.9, "hidden1 -> output"},
		{circuit.hiddenNeuron2, circuit.outputNeuron, 0.7, "hidden2 -> output"},

		// Output to memory update (feedback for next timestep)
		{circuit.outputNeuron, circuit.memoryNeuron2, 0.6, "output -> memory2 (parity update)"},

		// Supervision layer (error backpropagation)
		{circuit.teacherNeuron, circuit.errorNeuron, 1.0, "teacher -> error"},
		{circuit.outputNeuron, circuit.errorNeuron, 0.8, "output -> error"},
	}

	circuit.synapses = make([]component.SynapticProcessor, 0)
	for _, conn := range connections {
		syn, err := matrix.CreateSynapse(types.SynapseConfig{
			SynapseType:    "supervised_synapse",
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

	t.Logf("✓ Created %d temporal synapses with long eligibility traces", len(circuit.synapses))
}

// cleanupSerialCircuit stops all neurons
func cleanupSerialCircuit(circuit *SerialXORCircuit) {
	for _, neuron := range circuit.allNeurons {
		neuron.Stop()
	}
}

// runSupervisedSerialXORLearning implements supervised temporal XOR learning
func runSupervisedSerialXORLearning(matrix *extracellular.ExtracellularMatrix, circuit *SerialXORCircuit, sequences []TemporalXORSequence, t *testing.T) *SupervisedXORResults {
	t.Log("\n--- Supervised Serial XOR Learning ---")

	results := &SupervisedXORResults{}
	trainingEpochs := 30
	correctTraining := 0
	totalTraining := 0

	// Training phase with supervised error backpropagation
	for epoch := 0; epoch < trainingEpochs; epoch++ {
		if epoch%10 == 0 {
			t.Logf("Training epoch %d/%d", epoch, trainingEpochs)
		}

		for _, sequence := range sequences {
			// Process each bit in the sequence
			for step, bit := range sequence.bits {
				totalTraining++
				expectedParity := sequence.parities[step]

				// Clear short-term state but preserve memory
				time.Sleep(20 * time.Millisecond)

				// Present current bit
				presentSerialBit(circuit, bit)

				// Allow forward propagation through temporal network
				time.Sleep(30 * time.Millisecond)

				// Get network prediction
				predicted := getSerialOutput(circuit)
				correct := (predicted > 0.5 && expectedParity == 1) || (predicted <= 0.5 && expectedParity == 0)

				if correct {
					correctTraining++
				}

				// Apply supervised error backpropagation
				applySupervisedLearning(matrix, circuit, expectedParity, predicted, t)

				// Allow error backpropagation and eligibility trace updates
				time.Sleep(25 * time.Millisecond)

				t.Logf("Epoch %d, Seq %s, Step %d: bit=%d, expected=%d, predicted=%.3f, correct=%v", 
					epoch, sequence.name, step, bit, expectedParity, predicted, correct)
			}

			// Reset memory between sequences
			resetTemporalMemory(circuit)
			time.Sleep(50 * time.Millisecond)
		}
	}

	results.TrainingAccuracy = float64(correctTraining) / float64(totalTraining) * 100.0
	t.Logf("Training completed. Accuracy: %.1f%% (%d/%d correct)", 
		results.TrainingAccuracy, correctTraining, totalTraining)

	// Test phase on new sequences
	t.Log("\n--- Testing Supervised Serial XOR Function ---")
	testSequences := []TemporalXORSequence{
		{name: "test1", bits: []int{1, 0, 0}, parities: []int{1, 1, 1}},     // 1, 1^0=1, 1^0=1
		{name: "test2", bits: []int{0, 1, 0, 1}, parities: []int{0, 1, 1, 0}}, // 0, 0^1=1, 1^0=1, 1^1=0
		{name: "test3", bits: []int{1, 1, 1}, parities: []int{1, 0, 1}},     // 1, 1^1=0, 0^1=1
	}

	correctTest := 0
	totalTest := 0
	for _, sequence := range testSequences {
		t.Logf("\nTesting sequence %s: %v", sequence.name, sequence.bits)
		
		for step, bit := range sequence.bits {
			totalTest++
			expectedParity := sequence.parities[step]

			// Reset and present bit
			time.Sleep(30 * time.Millisecond)
			presentSerialBit(circuit, bit)
			time.Sleep(30 * time.Millisecond)

			// Get prediction
			predicted := getSerialOutput(circuit)
			correct := (predicted > 0.5 && expectedParity == 1) || (predicted <= 0.5 && expectedParity == 0)

			if correct {
				correctTest++
			}

			status := "❌"
			if correct {
				status = "✅"
			}

			t.Logf("  Step %d: bit=%d, expected_parity=%d, predicted=%.3f %s", 
				step, bit, expectedParity, predicted, status)
		}

		resetTemporalMemory(circuit)
		time.Sleep(50 * time.Millisecond)
	}

	results.TestAccuracy = float64(correctTest) / float64(totalTest) * 100.0
	results.CorrectResponses = correctTest
	results.TotalTests = totalTest

	// Analyze learned weights
	analyzeSupervisedWeights(circuit, t)

	return results
}

// presentSerialBit presents a single bit to the temporal network
func presentSerialBit(circuit *SerialXORCircuit, bit int) {
	if bit == 1 {
		circuit.inputNeuron.Receive(types.NeuralSignal{
			Value:     1.5,
			Timestamp: time.Now(),
			SourceID:  "serial_input",
			TargetID:  circuit.inputNeuron.ID(),
		})
	}
	// Note: bit=0 means no input signal (temporal absence)
}

// getSerialOutput gets the current parity output from the network
func getSerialOutput(circuit *SerialXORCircuit) float64 {
	return circuit.outputNeuron.GetActivityLevel()
}

// resetTemporalMemory clears the temporal memory between sequences
func resetTemporalMemory(circuit *SerialXORCircuit) {
	// Allow memory neurons to decay naturally
	// In a real implementation, could send inhibitory signals to reset state
}

// applySupervisedLearning applies supervised error backpropagation
func applySupervisedLearning(matrix *extracellular.ExtracellularMatrix, circuit *SerialXORCircuit, expectedParity int, predicted float64, t *testing.T) {
	// Calculate specific error signal
	var error float64
	if expectedParity == 1 {
		error = 1.0 - predicted // Error magnitude for target=1
	} else {
		error = predicted - 0.0 // Error magnitude for target=0
	}

	errorMagnitude := abs(error)

	if errorMagnitude > 0.1 { // Only apply supervision if error is significant
		// Present target signal to teacher neuron
		teacherSignal := float64(expectedParity) * 1.5
		if teacherSignal > 0 {
			circuit.teacherNeuron.Receive(types.NeuralSignal{
				Value:     teacherSignal,
				Timestamp: time.Now(),
				SourceID:  "teacher",
				TargetID:  circuit.teacherNeuron.ID(),
			})
		}

		// Release GABA error signal proportional to error magnitude
		gabaConcentration := 0.5 + errorMagnitude // 0.5-1.5 range
		matrix.ReleaseLigand(types.LigandGABA, circuit.hiddenNeuron1.ID(), gabaConcentration)
		matrix.ReleaseLigand(types.LigandGABA, circuit.hiddenNeuron2.ID(), gabaConcentration)
		matrix.ReleaseLigand(types.LigandGABA, circuit.outputNeuron.ID(), gabaConcentration)

		// Release dopamine based on error direction
		if error > 0 { // Need to increase output
			dopamineConcentration := 1.0 + errorMagnitude // Positive dopamine
			matrix.ReleaseLigand(types.LigandDopamine, circuit.hiddenNeuron1.ID(), dopamineConcentration)
			matrix.ReleaseLigand(types.LigandDopamine, circuit.hiddenNeuron2.ID(), dopamineConcentration)
		} else { // Need to decrease output
			dopamineConcentration := 1.0 - errorMagnitude // Negative dopamine
			if dopamineConcentration < 0.1 {
				dopamineConcentration = 0.1 // Minimum concentration
			}
			matrix.ReleaseLigand(types.LigandDopamine, circuit.hiddenNeuron1.ID(), dopamineConcentration)
			matrix.ReleaseLigand(types.LigandDopamine, circuit.hiddenNeuron2.ID(), dopamineConcentration)
		}
	}
}

// abs returns absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// analyzeSupervisedWeights analyzes learned synaptic weights
func analyzeSupervisedWeights(circuit *SerialXORCircuit, t *testing.T) {
	t.Log("\n--- Supervised Learning Weight Analysis ---")
	
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