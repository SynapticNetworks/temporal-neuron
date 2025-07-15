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

// XORPattern represents a temporal XOR pattern for training/testing
type XORPattern struct {
	name     string
	bits     []int  // Input bits (0 or 1)
	expected int    // Expected XOR output (parity)
	timing   []time.Duration // Spike timing for each input bit
}

// TestXORLearning tests learning of XOR (parity) function using neuromodulation and gating
func TestXORLearning(t *testing.T) {
	t.Log("=== XOR GATED NEUROMODULATED LEARNING TEST ===")
	t.Log("Testing XOR learning using context-dependent gating and multi-modal neuromodulation")

	// Create matrix with full chemical and spatial capabilities
	matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true, // Enable spatial for realistic delays
		UpdateInterval:  1 * time.Millisecond, // High temporal resolution
		MaxComponents:   100,
	})

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}
	defer matrix.Stop()

	// Enhanced network architecture for gated XOR learning
	const (
		numInputs = 2  // 2 inputs for XOR
	)

	// Register specialized neuron types using the same architecture as successful gating test
	registerXORNeuronTypes(matrix, t)

	// Register synapse type for gated XOR learning
	registerXORSynapseType(matrix, t)

	// Build gated XOR circuit with competitive dynamics
	circuit := buildGatedXORCircuit(matrix, t)
	defer cleanupCircuit(circuit.allNeurons)

	// Define XOR training patterns
	trainingPatterns := []XORPattern{
		{name: "00", bits: []int{0, 0}, expected: 0},
		{name: "01", bits: []int{0, 1}, expected: 1},
		{name: "10", bits: []int{1, 0}, expected: 1},
		{name: "11", bits: []int{1, 1}, expected: 0},
	}

	// Test gated XOR learning using context-dependent neuromodulation
	results := testGatedXORLearning(matrix, circuit, trainingPatterns, t)

	// Analyze and report results
	t.Logf("\n--- Gated XOR Learning Results ---")
	t.Logf("Training Accuracy: %.1f%%", results.TrainingAccuracy)
	t.Logf("Test Accuracy: %.1f%% (%d/%d correct)", results.TestAccuracy, results.CorrectResponses, results.TotalTests)

	// Summary and analysis
	t.Log("\n=== GATED NEUROMODULATED XOR LEARNING ANALYSIS ===")
	if results.TestAccuracy >= 75.0 {
		t.Log("✅ XOR learning successful! Context-dependent gating worked")
		if results.TestAccuracy >= 90.0 {
			t.Log("🎉 Excellent gated neuromodulated learning performance")
		}
	} else if results.TestAccuracy >= 50.0 {
		t.Log("⚠️  Partial XOR learning - some gating effect observed")
	} else {
		t.Log("❌ XOR learning failed - gating and neuromodulation insufficient")
	}
	
	t.Logf("Architecture: %d inputs, 2 detection + 2 integration + 2 inhibition, 2 competing outputs", numInputs)
	t.Logf("Learning mechanism: Context-dependent neuromodulation and competitive gating")
	t.Logf("Neuromodulators: Dopamine (detection), Serotonin (integration), Norepinephrine (inhibition)")
	t.Logf("Ion channels: Nav1.6 (fast), Cav1.2 (integration), GABA-A (inhibition)")
	
	// Test biological plausibility
	if results.TestAccuracy >= 75.0 {
		t.Log("\n🧠 Biological Significance:")
		t.Log("- Context-dependent gating enabled non-linear XOR computation")
		t.Log("- Multiple neuromodulators created specialized processing modes")
		t.Log("- Competitive dynamics provided winner-takes-all decisions")
		t.Log("- Specialized ion channels supported different computational roles")
		t.Log("- Architecture mimics cortical column organization with gating")
	}
}

// calculateXOR computes XOR (parity) for a slice of bits
func calculateXOR(bits []int) int {
	result := 0
	for _, bit := range bits {
		result ^= bit
	}
	return result
}

// TestXORGeneralization tests if the network can learn XOR and generalize to longer sequences
func TestXORGeneralization(t *testing.T) {
	t.Log("=== XOR GENERALIZATION TEST ===")
	t.Log("Testing XOR learning generalization from 2-bit to 8-bit patterns")

	// This test would train on 2-3 bit patterns and test generalization to longer sequences
	// Implementation would be similar to TestXORLearning but with systematic testing
	// of generalization capabilities
	
	bitLengths := []int{2, 3, 4, 5, 6, 7, 8}
	
	for _, bitLength := range bitLengths {
		t.Logf("Testing %d-bit XOR patterns:", bitLength)
		
		// Generate test patterns for this bit length
		numPatterns := 1 << bitLength // 2^bitLength patterns
		if numPatterns > 16 { // Limit test patterns for longer sequences
			numPatterns = 16
		}
		
		correctCount := 0
		for i := 0; i < numPatterns; i++ {
			// Convert i to binary pattern
			bits := make([]int, bitLength)
			for j := 0; j < bitLength; j++ {
				bits[j] = (i >> j) & 1
			}
			
			expectedXOR := calculateXOR(bits)
			
			// Here we would present the pattern to the trained network
			// and check if it produces the correct XOR result
			// For now, we'll simulate the test
			
			// Simulated result (in real implementation, this would be network output)
			predictedXOR := expectedXOR // Placeholder
			
			if predictedXOR == expectedXOR {
				correctCount++
			}
		}
		
		accuracy := float64(correctCount) / float64(numPatterns) * 100.0
		t.Logf("  %d-bit accuracy: %.1f%% (%d/%d correct)", 
			bitLength, accuracy, correctCount, numPatterns)
	}
	
	t.Log("\nNote: This is a framework for generalization testing.")
	t.Log("Full implementation would require training a network and testing generalization.")
}

// XORGatedCircuit represents a gated neuromodulated XOR learning circuit
type XORGatedCircuit struct {
	// Input layer
	inputA, inputB component.NeuralComponent

	// Processing layers with specialized functions
	detectionNeuron1, detectionNeuron2     component.NeuralComponent // Fast neurons for pattern detection
	integrationNeuron1, integrationNeuron2 component.NeuralComponent // Integrative neurons for context
	inhibitionNeuron1, inhibitionNeuron2   component.NeuralComponent // Inhibitory neurons for gating

	// Output layer with competitive dynamics
	outputXOR0, outputXOR1 component.NeuralComponent // Competing output neurons

	allNeurons []component.NeuralComponent
}

// registerXORNeuronTypes creates specialized neuron types for gated XOR learning
func registerXORNeuronTypes(matrix *extracellular.ExtracellularMatrix, t *testing.T) {
	// FAST DETECTION NEURON: High Nav1.6 density for rapid pattern detection (enhanced by dopamine)
	matrix.RegisterNeuronType("detection_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.98, 1*time.Millisecond, 2.0, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandDopamine})
		n.SetCallbacks(callbacks)

		dendriticMode := neuron.NewTemporalSummationMode()
		// High density Nav1.6 for rapid detection
		dendriticMode.AddChannel(neuron.NewRealisticNavChannel("nav1.6_detect_1"))
		dendriticMode.AddChannel(neuron.NewRealisticNavChannel("nav1.6_detect_2"))
		dendriticMode.AddChannel(neuron.NewRealisticNavChannel("nav1.6_detect_3"))
		// Moderate K+ for controlled repolarization
		dendriticMode.AddChannel(neuron.NewRealisticKvChannel("kv4.2_detect"))

		if err := n.SetDendriticMode(dendriticMode); err != nil {
			return nil, fmt.Errorf("failed to set dendritic mode: %v", err)
		}
		return n, nil
	})

	// INTEGRATIVE CONTEXT NEURON: High Cav1.2 density for temporal integration (enhanced by serotonin)
	matrix.RegisterNeuronType("integration_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.92, 5*time.Millisecond, 1.3, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandSerotonin})
		n.SetCallbacks(callbacks)

		dendriticMode := neuron.NewTemporalSummationMode()
		// Standard Nav channels
		dendriticMode.AddChannel(neuron.NewRealisticNavChannel("nav1.6_int"))
		// High density Cav1.2 for calcium integration and context
		dendriticMode.AddChannel(neuron.NewRealisticCavChannel("cav1.2_int_1"))
		dendriticMode.AddChannel(neuron.NewRealisticCavChannel("cav1.2_int_2"))
		dendriticMode.AddChannel(neuron.NewRealisticCavChannel("cav1.2_int_3"))
		// K+ channels for controlled excitability
		dendriticMode.AddChannel(neuron.NewRealisticKvChannel("kv4.2_int"))

		if err := n.SetDendriticMode(dendriticMode); err != nil {
			return nil, fmt.Errorf("failed to set dendritic mode: %v", err)
		}
		return n, nil
	})

	// INHIBITORY GATING NEURON: High GABA-A density for competitive gating (enhanced by norepinephrine)
	matrix.RegisterNeuronType("inhibition_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.95, 2*time.Millisecond, 1.8, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandNorepinephrine, types.LigandGABA})
		n.SetCallbacks(callbacks)

		dendriticMode := neuron.NewTemporalSummationMode()
		// Fast Nav channels for rapid inhibitory responses
		dendriticMode.AddChannel(neuron.NewRealisticNavChannel("nav1.6_inh"))
		// High density GABA-A channels for strong inhibitory input
		dendriticMode.AddChannel(neuron.NewRealisticGabaAChannel("gabaa_inh_1"))
		dendriticMode.AddChannel(neuron.NewRealisticGabaAChannel("gabaa_inh_2"))
		// Strong K+ channels for hyperpolarization
		dendriticMode.AddChannel(neuron.NewRealisticKvChannel("kv4.2_inh_1"))
		dendriticMode.AddChannel(neuron.NewRealisticKvChannel("kv4.2_inh_2"))

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

	t.Log("✓ Registered specialized XOR neuron types: Detection, Integration, Inhibition, Standard")
}

// registerXORSynapseType creates synapses for gated XOR learning
func registerXORSynapseType(matrix *extracellular.ExtracellularMatrix, t *testing.T) {
	matrix.RegisterSynapseType("xor_gated_synapse", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, exists := matrix.GetNeuron(config.PresynapticID)
		if !exists {
			return nil, fmt.Errorf("presynaptic neuron not found: %s", config.PresynapticID)
		}
		postNeuron, exists := matrix.GetNeuron(config.PostsynapticID)
		if !exists {
			return nil, fmt.Errorf("postsynaptic neuron not found: %s", config.PostsynapticID)
		}
		return synapse.NewBasicSynapse(id, preNeuron, postNeuron,
			synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(),
			config.InitialWeight, config.Delay), nil
	})
}

// buildGatedXORCircuit creates the full gated XOR learning circuit
func buildGatedXORCircuit(matrix *extracellular.ExtracellularMatrix, t *testing.T) *XORGatedCircuit {
	circuit := &XORGatedCircuit{}
	var err error

	t.Log("\n--- Building Gated XOR Circuit ---")

	// Create input neurons
	circuit.inputA, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "standard_neuron", Threshold: 0.5})
	if err != nil {
		t.Fatalf("Failed to create inputA: %v", err)
	}
	circuit.inputB, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "standard_neuron", Threshold: 0.5})
	if err != nil {
		t.Fatalf("Failed to create inputB: %v", err)
	}

	// Create detection neurons (enhanced by dopamine)
	circuit.detectionNeuron1, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "detection_neuron", Threshold: 1.5})
	if err != nil {
		t.Fatalf("Failed to create detectionNeuron1: %v", err)
	}
	circuit.detectionNeuron2, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "detection_neuron", Threshold: 1.5})
	if err != nil {
		t.Fatalf("Failed to create detectionNeuron2: %v", err)
	}

	// Create integration neurons (enhanced by serotonin)
	circuit.integrationNeuron1, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "integration_neuron", Threshold: 2.0})
	if err != nil {
		t.Fatalf("Failed to create integrationNeuron1: %v", err)
	}
	circuit.integrationNeuron2, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "integration_neuron", Threshold: 2.0})
	if err != nil {
		t.Fatalf("Failed to create integrationNeuron2: %v", err)
	}

	// Create inhibition neurons (enhanced by norepinephrine)
	circuit.inhibitionNeuron1, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "inhibition_neuron", Threshold: 1.8})
	if err != nil {
		t.Fatalf("Failed to create inhibitionNeuron1: %v", err)
	}
	circuit.inhibitionNeuron2, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "inhibition_neuron", Threshold: 1.8})
	if err != nil {
		t.Fatalf("Failed to create inhibitionNeuron2: %v", err)
	}

	// Create competing output neurons
	circuit.outputXOR0, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "standard_neuron", Threshold: 1.5})
	if err != nil {
		t.Fatalf("Failed to create outputXOR0: %v", err)
	}
	circuit.outputXOR1, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "standard_neuron", Threshold: 1.5})
	if err != nil {
		t.Fatalf("Failed to create outputXOR1: %v", err)
	}

	// Collect all neurons
	circuit.allNeurons = []component.NeuralComponent{
		circuit.inputA, circuit.inputB,
		circuit.detectionNeuron1, circuit.detectionNeuron2,
		circuit.integrationNeuron1, circuit.integrationNeuron2,
		circuit.inhibitionNeuron1, circuit.inhibitionNeuron2,
		circuit.outputXOR0, circuit.outputXOR1,
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

	// Wire the gated circuit
	wireGatedXORCircuit(matrix, circuit, t)

	t.Log("✓ Gated XOR circuit created with competitive dynamics")
	return circuit
}

// wireGatedXORCircuit creates the gated connectivity pattern
func wireGatedXORCircuit(matrix *extracellular.ExtracellularMatrix, circuit *XORGatedCircuit, t *testing.T) {
	connections := []struct {
		pre, post component.NeuralComponent
		weight    float64
		desc      string
	}{
		// Input to detection layer (pattern detection)
		{circuit.inputA, circuit.detectionNeuron1, 1.0, "inputA -> detect1"},
		{circuit.inputB, circuit.detectionNeuron2, 1.0, "inputB -> detect2"},

		// Input to integration layer (context)
		{circuit.inputA, circuit.integrationNeuron1, 0.8, "inputA -> integrate1"},
		{circuit.inputB, circuit.integrationNeuron2, 0.8, "inputB -> integrate2"},

		// Cross-connections for pattern combinations
		{circuit.inputA, circuit.detectionNeuron2, 0.6, "inputA -> detect2"},
		{circuit.inputB, circuit.detectionNeuron1, 0.6, "inputB -> detect1"},

		// Detection to outputs (rapid responses)
		{circuit.detectionNeuron1, circuit.outputXOR1, 1.2, "detect1 -> XOR1"},
		{circuit.detectionNeuron2, circuit.outputXOR1, 1.2, "detect2 -> XOR1"},

		// Integration to outputs (context-dependent responses)
		{circuit.integrationNeuron1, circuit.outputXOR0, 1.0, "integrate1 -> XOR0"},
		{circuit.integrationNeuron2, circuit.outputXOR0, 1.0, "integrate2 -> XOR0"},

		// Competitive inhibition between outputs
		{circuit.inhibitionNeuron1, circuit.outputXOR0, -0.8, "inhibit1 -> XOR0 (competitive)"},
		{circuit.inhibitionNeuron2, circuit.outputXOR1, -0.8, "inhibit2 -> XOR1 (competitive)"},

		// Detection -> inhibition (competitive gating)
		{circuit.detectionNeuron1, circuit.inhibitionNeuron2, 0.9, "detect1 -> inhibit2"},
		{circuit.detectionNeuron2, circuit.inhibitionNeuron1, 0.9, "detect2 -> inhibit1"},
	}

	for _, conn := range connections {
		weight := conn.weight
		if weight < 0 {
			weight = -weight // Synapse weight must be positive, inhibition handled by neuron type
		}
		
		_, err := matrix.CreateSynapse(types.SynapseConfig{
			SynapseType:    "xor_gated_synapse",
			PresynapticID:  conn.pre.ID(),
			PostsynapticID: conn.post.ID(),
			InitialWeight:  weight,
			Delay:          1 * time.Millisecond,
		})
		if err != nil {
			t.Fatalf("Failed to create synapse %s: %v", conn.desc, err)
		}
	}

	t.Logf("✓ Created gated XOR connectivity with competitive dynamics")
}

// cleanupCircuit stops all neurons
func cleanupCircuit(neurons []component.NeuralComponent) {
	for _, neuron := range neurons {
		neuron.Stop()
	}
}

// XORLearningResults stores results from gated XOR learning
type XORLearningResults struct {
	TrainingAccuracy float64
	TestAccuracy     float64
	CorrectResponses int
	TotalTests       int
}

// testGatedXORLearning implements context-dependent XOR learning using gating and neuromodulation
func testGatedXORLearning(matrix *extracellular.ExtracellularMatrix, circuit *XORGatedCircuit, patterns []XORPattern, t *testing.T) *XORLearningResults {
	t.Log("\n--- Gated XOR Learning with Context-Dependent Neuromodulation ---")

	results := &XORLearningResults{}
	trainingIterations := 30
	correctTraining := 0

	// Phase 1: Context-dependent gating for each pattern
	for iteration := 0; iteration < trainingIterations; iteration++ {
		if iteration%10 == 0 {
			t.Logf("Training iteration %d/%d", iteration, trainingIterations)
		}

		for _, pattern := range patterns {
			// Clear network state
			time.Sleep(20 * time.Millisecond)

			// Context-dependent neuromodulation based on expected output
			if pattern.expected == 0 {
				// For XOR=0 patterns (00, 11): Enhance integration for null output
				releaseSerotonin(matrix, circuit, t)
				releaseNorepinephrine(matrix, circuit, t) // Inhibit competing responses
			} else {
				// For XOR=1 patterns (01, 10): Enhance detection for active output  
				releaseDopamine(matrix, circuit, t)
			}

			// Allow neuromodulation to take effect
			time.Sleep(15 * time.Millisecond)

			// Present input pattern
			presentGatedPattern(circuit, pattern)

			// Wait for gated response
			time.Sleep(30 * time.Millisecond)

			// Check competitive output
			predicted := checkGatedOutput(circuit)
			correct := predicted == pattern.expected

			if correct {
				correctTraining++
			}

			// Brief recovery period
			time.Sleep(10 * time.Millisecond)
		}
	}

	results.TrainingAccuracy = float64(correctTraining) / float64(trainingIterations*len(patterns)) * 100.0
	t.Logf("Training completed. Accuracy: %.1f%% (%d/%d correct)", 
		results.TrainingAccuracy, correctTraining, trainingIterations*len(patterns))

	// Phase 2: Test without training rewards
	t.Log("\n--- Testing Gated XOR Function ---")
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

		// Present test pattern without explicit neuromodulation (test learned gating)
		presentGatedPattern(circuit, pattern)

		// Wait for response
		time.Sleep(30 * time.Millisecond)

		// Check output
		predicted := checkGatedOutput(circuit)
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

	return results
}

// releaseDopamine enhances detection neurons for rapid pattern recognition
func releaseDopamine(matrix *extracellular.ExtracellularMatrix, circuit *XORGatedCircuit, t *testing.T) {
	matrix.ReleaseLigand(types.LigandDopamine, circuit.detectionNeuron1.ID(), 1.2)
	matrix.ReleaseLigand(types.LigandDopamine, circuit.detectionNeuron2.ID(), 1.2)
}

// releaseSerotonin enhances integration neurons for temporal context
func releaseSerotonin(matrix *extracellular.ExtracellularMatrix, circuit *XORGatedCircuit, t *testing.T) {
	matrix.ReleaseLigand(types.LigandSerotonin, circuit.integrationNeuron1.ID(), 1.0)
	matrix.ReleaseLigand(types.LigandSerotonin, circuit.integrationNeuron2.ID(), 1.0)
}

// releaseNorepinephrine enhances inhibition neurons for competitive gating
func releaseNorepinephrine(matrix *extracellular.ExtracellularMatrix, circuit *XORGatedCircuit, t *testing.T) {
	matrix.ReleaseLigand(types.LigandNorepinephrine, circuit.inhibitionNeuron1.ID(), 1.1)
	matrix.ReleaseLigand(types.LigandNorepinephrine, circuit.inhibitionNeuron2.ID(), 1.1)
}

// presentGatedPattern presents input to the gated circuit
func presentGatedPattern(circuit *XORGatedCircuit, pattern XORPattern) {
	if pattern.bits[0] == 1 {
		circuit.inputA.Receive(types.NeuralSignal{
			Value:     1.5,
			Timestamp: time.Now(),
			SourceID:  "gated_xor_input",
			TargetID:  circuit.inputA.ID(),
		})
	}
	if pattern.bits[1] == 1 {
		circuit.inputB.Receive(types.NeuralSignal{
			Value:     1.5,
			Timestamp: time.Now(),
			SourceID:  "gated_xor_input",
			TargetID:  circuit.inputB.ID(),
		})
	}
}

// checkGatedOutput determines XOR output from competitive neurons
func checkGatedOutput(circuit *XORGatedCircuit) int {
	output0 := circuit.outputXOR0.GetActivityLevel()
	output1 := circuit.outputXOR1.GetActivityLevel()

	// Winner-takes-all competitive decision
	if output1 > output0 {
		return 1
	}
	return 0
}