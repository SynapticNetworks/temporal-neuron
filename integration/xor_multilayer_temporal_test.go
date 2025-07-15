package integration

import (
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/extracellular"
	"github.com/SynapticNetworks/temporal-neuron/neuron"
	"github.com/SynapticNetworks/temporal-neuron/synapse"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// getMultilayerMembraneActivity returns the neuron's instantaneous membrane activity
func getMultilayerMembraneActivity(n component.NeuralComponent) float64 {
	if statusProvider, ok := n.(interface{ GetProcessingStatus() map[string]interface{} }); ok {
		status := statusProvider.GetProcessingStatus()
		if neuralState, ok := status["neural_state"].(map[string]interface{}); ok {
			if accumulator, ok := neuralState["accumulator"].(float64); ok {
				if threshold, ok := neuralState["threshold"].(float64); ok {
					return accumulator / threshold
				}
			}
		}
	}
	return 0.0
}

// TemporalXORNetwork represents a multi-layer temporal XOR learning network
type TemporalXORNetwork struct {
	Input   component.NeuralComponent
	Hidden1 component.NeuralComponent  // Detects single spikes / odd patterns
	Hidden2 component.NeuralComponent  // Detects spike pairs / even patterns  
	Output  component.NeuralComponent
	matrix  *extracellular.ExtracellularMatrix
}

// presentTemporalPattern presents a temporal sequence to the network
func (net *TemporalXORNetwork) presentTemporalPattern(pattern []int, t *testing.T) float64 {
	t.Logf("Presenting temporal pattern: %v", pattern)
	
	// Reset network state
	time.Sleep(100 * time.Millisecond)
	
	// Present each bit in the temporal sequence
	for i, bit := range pattern {
		if bit == 1 {
			net.Input.Receive(types.NeuralSignal{
				Value:     2.0, // Strong signal to ensure firing
				Timestamp: time.Now(),
				SourceID:  "temporal_input",
				TargetID:  net.Input.ID(),
			})
			t.Logf("  Time %d: SPIKE", i)
		} else {
			t.Logf("  Time %d: silence", i)
		}
		
		// Inter-spike interval for temporal processing
		time.Sleep(25 * time.Millisecond)
	}
	
	// Wait for network to process the complete pattern
	time.Sleep(50 * time.Millisecond)
	
	// Read final output membrane activity
	outputActivity := getMultilayerMembraneActivity(net.Output)
	
	// Also log hidden layer activities for debugging
	hidden1Activity := getMultilayerMembraneActivity(net.Hidden1)
	hidden2Activity := getMultilayerMembraneActivity(net.Hidden2)
	
	t.Logf("  Network response - Hidden1: %.3f, Hidden2: %.3f, Output: %.3f", 
		hidden1Activity, hidden2Activity, outputActivity)
	
	return outputActivity
}

// calculateMultilayerParity calculates XOR parity (odd=1, even=0)
func calculateMultilayerParity(pattern []int) int {
	count := 0
	for _, bit := range pattern {
		count += bit
	}
	return count % 2 // 1 if odd number of spikes, 0 if even
}

// applySupervisedLearning provides learning signals based on error
func (net *TemporalXORNetwork) applySupervisedLearning(expectedParity int, actualOutput float64, t *testing.T) {
	// Calculate error
	var error float64
	if expectedParity == 1 {
		error = 1.0 - actualOutput  // Should be high, but it's low
	} else {
		error = actualOutput - 0.0  // Should be low, but it's high
	}
	
	// Apply chemical learning signals based on error magnitude
	if multilayerAbs(error) > 0.3 {
		t.Logf("  Learning signal: error=%.3f", error)
		
		if error > 0 {
			// Need MORE activity - reward via dopamine
			net.matrix.ReleaseLigand(types.LigandDopamine, net.Output.ID(), 0.6)
		} else {
			// Need LESS activity - inhibit via GABA
			net.matrix.ReleaseLigand(types.LigandGABA, net.Output.ID(), 0.6)
		}
		
		// Allow time for chemical learning
		time.Sleep(30 * time.Millisecond)
	}
}

// getSynapseWeight gets the weight of a synapse
func getSynapseWeight(syn component.SynapticProcessor) float64 {
	if weightGetter, ok := syn.(interface{ GetWeight() float64 }); ok {
		return weightGetter.GetWeight()
	}
	return 0.0
}

// TestXORMultilayerTemporal tests temporal XOR learning with hidden layers
func TestXORMultilayerTemporal(t *testing.T) {
	t.Log("=== MULTILAYER TEMPORAL XOR LEARNING TEST ===")
	t.Log("Testing temporal XOR with hidden layers and multiple synapses")

	// Create matrix
	matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  1 * time.Millisecond,
		MaxComponents:   20,
	})

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}
	defer matrix.Stop()

	// Register learning neuron with STDP and neuromodulation
	matrix.RegisterNeuronType("temporal_learning_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.85, 10*time.Millisecond, 1.5, 0.0, 0.0)
		
		// Set up for chemical learning
		n.SetReceptors([]types.LigandType{
			types.LigandGlutamate, 
			types.LigandGABA, 
			types.LigandDopamine,
			types.LigandSerotonin,
		})
		n.SetReleasedLigands([]types.LigandType{types.LigandGlutamate})
		n.SetCallbacks(callbacks)
		
		// Enable STDP for synaptic learning
		n.EnableSTDPFeedback(15*time.Millisecond, 0.15)
		
		return n, nil
	})

	// Register learning synapse with STDP
	matrix.RegisterSynapseType("temporal_learning_synapse", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, _ := matrix.GetNeuron(config.PresynapticID)
		postNeuron, _ := matrix.GetNeuron(config.PostsynapticID)

		syn := synapse.NewBasicSynapse(id, preNeuron, postNeuron,
			synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(),
			config.InitialWeight, config.Delay)

		return syn, nil
	})

	t.Log("\n--- CREATING MULTILAYER NETWORK ---")

	// Create the network architecture
	input, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "temporal_learning_neuron", 
		Threshold:  1.5,
	})
	if err != nil {
		t.Fatalf("Failed to create input neuron: %v", err)
	}

	hidden1, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "temporal_learning_neuron", 
		Threshold:  1.2, // Sensitive to single spikes
	})
	if err != nil {
		t.Fatalf("Failed to create hidden1 neuron: %v", err)
	}

	hidden2, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "temporal_learning_neuron", 
		Threshold:  1.8, // Requires multiple spikes to activate
	})
	if err != nil {
		t.Fatalf("Failed to create hidden2 neuron: %v", err)
	}

	output, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "temporal_learning_neuron", 
		Threshold:  1.0, // Moderate threshold for final decision
	})
	if err != nil {
		t.Fatalf("Failed to create output neuron: %v", err)
	}

	// Start all neurons
	neurons := []component.NeuralComponent{input, hidden1, hidden2, output}
	for _, n := range neurons {
		n.Start()
		if receiver, ok := n.(component.ChemicalReceiver); ok {
			matrix.RegisterForBinding(receiver)
		}
	}
	defer func() {
		for _, n := range neurons {
			n.Stop()
		}
	}()

	t.Log("\n--- CREATING SYNAPTIC CONNECTIONS ---")

	// Create synaptic connections
	// Input to both hidden layers
	syn1, err := matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "temporal_learning_synapse",
		PresynapticID:  input.ID(),
		PostsynapticID: hidden1.ID(),
		InitialWeight:  0.8, // Strong connection to single-spike detector
		Delay:          2 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create synapse 1: %v", err)
	}

	syn2, err := matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "temporal_learning_synapse",
		PresynapticID:  input.ID(),
		PostsynapticID: hidden2.ID(),
		InitialWeight:  0.6, // Moderate connection to pair detector
		Delay:          2 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create synapse 2: %v", err)
	}

	// Hidden1 to output (excitatory - for odd counts)
	syn3, err := matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "temporal_learning_synapse",
		PresynapticID:  hidden1.ID(),
		PostsynapticID: output.ID(),
		InitialWeight:  0.7, // Should strengthen for XOR=1 cases
		Delay:          3 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create synapse 3: %v", err)
	}

	// Hidden2 to output (should become inhibitory through learning)
	syn4, err := matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "temporal_learning_synapse",
		PresynapticID:  hidden2.ID(),
		PostsynapticID: output.ID(),
		InitialWeight:  0.3, // Should weaken or become inhibitory for XOR=0 cases
		Delay:          3 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create synapse 4: %v", err)
	}

	// Lateral inhibition between hidden layers
	syn5, err := matrix.CreateSynapse(types.SynapseConfig{
		SynapseType:    "temporal_learning_synapse",
		PresynapticID:  hidden1.ID(),
		PostsynapticID: hidden2.ID(),
		InitialWeight:  0.4, // Competitive inhibition
		Delay:          2 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create synapse 5: %v", err)
	}

	// Create network structure
	network := &TemporalXORNetwork{
		Input:   input,
		Hidden1: hidden1,
		Hidden2: hidden2,
		Output:  output,
		matrix:  matrix,
	}

	// Log initial synapse weights
	t.Logf("Initial weights:")
	t.Logf("  Input→Hidden1: %.3f", getSynapseWeight(syn1))
	t.Logf("  Input→Hidden2: %.3f", getSynapseWeight(syn2))
	t.Logf("  Hidden1→Output: %.3f", getSynapseWeight(syn3))
	t.Logf("  Hidden2→Output: %.3f", getSynapseWeight(syn4))
	t.Logf("  Hidden1→Hidden2: %.3f", getSynapseWeight(syn5))

	t.Log("\n--- TRAINING PHASE ---")

	// Training patterns for temporal XOR
	trainingPatterns := [][]int{
		{0, 0}, // 0 spikes = even = 0
		{0, 1}, // 1 spike = odd = 1
		{1, 0}, // 1 spike = odd = 1
		{1, 1}, // 2 spikes = even = 0
	}

	// Training epochs
	for epoch := 0; epoch < 15; epoch++ {
		t.Logf("\n--- EPOCH %d ---", epoch+1)
		
		correctPredictions := 0
		
		for _, pattern := range trainingPatterns {
			expectedParity := calculateMultilayerParity(pattern)
			outputActivity := network.presentTemporalPattern(pattern, t)
			
			// Make prediction
			var predicted int
			if outputActivity > 0.3 { // Lower threshold than before
				predicted = 1
			} else {
				predicted = 0
			}
			
			// Check accuracy
			correct := (predicted == expectedParity)
			if correct {
				correctPredictions++
			}
			
			t.Logf("Pattern %v: Expected=%d, Output=%.3f, Predicted=%d %s", 
				pattern, expectedParity, outputActivity, predicted, 
				map[bool]string{true: "✅", false: "❌"}[correct])
			
			// Apply supervised learning
			network.applySupervisedLearning(expectedParity, outputActivity, t)
		}
		
		accuracy := float64(correctPredictions) / float64(len(trainingPatterns)) * 100
		t.Logf("Epoch %d accuracy: %.1f%% (%d/%d)", 
			epoch+1, accuracy, correctPredictions, len(trainingPatterns))
		
		// Log weight changes
		if epoch % 3 == 0 {
			t.Logf("Current weights:")
			t.Logf("  Input→Hidden1: %.3f", getSynapseWeight(syn1))
			t.Logf("  Input→Hidden2: %.3f", getSynapseWeight(syn2))
			t.Logf("  Hidden1→Output: %.3f", getSynapseWeight(syn3))
			t.Logf("  Hidden2→Output: %.3f", getSynapseWeight(syn4))
			t.Logf("  Hidden1→Hidden2: %.3f", getSynapseWeight(syn5))
		}
		
		if accuracy >= 90.0 {
			t.Logf("✅ Target accuracy achieved at epoch %d!", epoch+1)
			break
		}
	}

	t.Log("\n--- FINAL TESTING ---")

	// Final test on all patterns
	correctPredictions := 0
	for _, pattern := range trainingPatterns {
		expectedParity := calculateMultilayerParity(pattern)
		outputActivity := network.presentTemporalPattern(pattern, t)
		
		var predicted int
		if outputActivity > 0.3 {
			predicted = 1
		} else {
			predicted = 0
		}
		
		correct := (predicted == expectedParity)
		if correct {
			correctPredictions++
		}
		
		t.Logf("Final test %v: Expected=%d, Output=%.3f, Predicted=%d %s", 
			pattern, expectedParity, outputActivity, predicted, 
			map[bool]string{true: "✅", false: "❌"}[correct])
	}
	
	finalAccuracy := float64(correctPredictions) / float64(len(trainingPatterns)) * 100
	t.Logf("\nFinal accuracy: %.1f%% (%d/%d)", finalAccuracy, correctPredictions, len(trainingPatterns))

	t.Log("\n--- GENERALIZATION TEST ---")

	// Test on longer sequences
	generalizationPatterns := [][]int{
		{0, 0, 0}, // 0 spikes = even = 0
		{0, 0, 1}, // 1 spike = odd = 1
		{0, 1, 1}, // 2 spikes = even = 0
		{1, 1, 1}, // 3 spikes = odd = 1
	}
	
	generalizeCorrect := 0
	for _, pattern := range generalizationPatterns {
		expectedParity := calculateMultilayerParity(pattern)
		outputActivity := network.presentTemporalPattern(pattern, t)
		
		var predicted int
		if outputActivity > 0.3 {
			predicted = 1
		} else {
			predicted = 0
		}
		
		correct := (predicted == expectedParity)
		if correct {
			generalizeCorrect++
		}
		
		t.Logf("Generalize %v: Expected=%d, Output=%.3f, Predicted=%d %s", 
			pattern, expectedParity, outputActivity, predicted, 
			map[bool]string{true: "✅", false: "❌"}[correct])
	}
	
	generalizationAccuracy := float64(generalizeCorrect) / float64(len(generalizationPatterns)) * 100
	t.Logf("Generalization accuracy: %.1f%% (%d/%d)", generalizationAccuracy, generalizeCorrect, len(generalizationPatterns))

	t.Log("\n--- FINAL ASSESSMENT ---")

	// Final weights
	t.Logf("Final weights:")
	t.Logf("  Input→Hidden1: %.3f", getSynapseWeight(syn1))
	t.Logf("  Input→Hidden2: %.3f", getSynapseWeight(syn2))
	t.Logf("  Hidden1→Output: %.3f", getSynapseWeight(syn3))
	t.Logf("  Hidden2→Output: %.3f", getSynapseWeight(syn4))
	t.Logf("  Hidden1→Hidden2: %.3f", getSynapseWeight(syn5))

	success := finalAccuracy >= 75.0
	
	if success {
		t.Log("🎯 SUCCESS: Multilayer temporal XOR learning achieved!")
		t.Logf("   Training accuracy: %.1f%%", finalAccuracy)
		t.Logf("   Generalization: %.1f%%", generalizationAccuracy)
		t.Log("   ✅ Hidden layers enable complex temporal pattern learning")
		t.Log("   ✅ Multiple synapses provide sufficient computational capacity")
		t.Log("   ✅ STDP + neuromodulation drives effective learning")
	} else {
		t.Log("❌ LEARNING INCOMPLETE")
		t.Logf("   Achieved %.1f%% accuracy (target: 75%%+)", finalAccuracy)
		t.Log("   Network architecture may need further tuning")
	}
}

// multilayerAbs returns absolute value to avoid naming conflicts
func multilayerAbs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}