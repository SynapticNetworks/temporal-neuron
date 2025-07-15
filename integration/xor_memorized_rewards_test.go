package integration

import (
	"strings"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/extracellular"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// TestXORMemorizedPatternDetectionWithRewards tests memorization with proper supervised learning
func TestXORMemorizedPatternDetectionWithRewards(t *testing.T) {
	t.Log("=== XOR MEMORIZED PATTERN DETECTION WITH PROPER REWARDS ===")
	t.Log("Testing memorization with full supervised learning (like the 70% success case)")

	// Create matrix
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

	// Register neuron and synapse types
	registerRuleLearningNeuronTypes(matrix, t)
	registerRuleLearningSynapseTypes(matrix, t)

	// Build circuit
	circuit := buildRuleLearningCircuit(matrix, t)
	defer cleanupRuleLearningCircuit(circuit)

	// Define EXACT training patterns
	trainingSequences := []TemporalXORSequence{
		{name: "XOR_00", bits: []int{0, 0}, parities: []int{0, 0}},
		{name: "XOR_01", bits: []int{0, 1}, parities: []int{0, 1}},
		{name: "XOR_10", bits: []int{1, 0}, parities: []int{1, 1}},
		{name: "XOR_11", bits: []int{1, 1}, parities: []int{1, 0}},
	}

	// STEP 1: Train with PROPER supervised learning (like serial XOR)
	t.Log("\n--- STEP 1: Training with PROPER Supervised Learning ---")
	trainingAccuracy := trainWithProperSupervision(matrix, circuit, trainingSequences, 30, t)
	t.Logf("Training accuracy with proper supervision: %.1f%%", trainingAccuracy)

	// STEP 2: Test recall of exact patterns
	t.Log("\n--- STEP 2: Testing Recall After Proper Training ---")
	memoryRecall := testExactPatternRecall(circuit, trainingSequences, t)
	t.Logf("Exact pattern recall accuracy: %.1f%%", memoryRecall)

	// Analysis
	t.Log("\n" + strings.Repeat("=", 60))
	t.Log("MEMORIZED PATTERN DETECTION WITH REWARDS ANALYSIS")
	t.Log(strings.Repeat("=", 60))

	if memoryRecall >= 90.0 {
		t.Log("✅ EXCELLENT MEMORIZATION: Proper supervision enabled perfect recall")
	} else if memoryRecall >= 75.0 {
		t.Log("✅ GOOD MEMORIZATION: Proper supervision enabled reliable recall")
	} else if memoryRecall >= 50.0 {
		t.Log("⚠️  PARTIAL MEMORIZATION: Some improvement with proper supervision")
	} else {
		t.Log("❌ POOR MEMORIZATION: Even proper supervision failed")
	}

	if trainingAccuracy >= 70.0 {
		t.Log("✅ SUCCESSFUL TRAINING: Network learned during training phase")
	} else {
		t.Log("❌ FAILED TRAINING: Network failed to learn even with proper supervision")
	}

	t.Logf("\nCOMPARISON:")
	t.Logf("- Training accuracy: %.1f%%", trainingAccuracy)
	t.Logf("- Recall accuracy: %.1f%%", memoryRecall)

	improvementVsBasic := memoryRecall - 50.0 // Compare to 50% baseline
	if improvementVsBasic > 20.0 {
		t.Logf("🎉 MAJOR IMPROVEMENT: +%.1f%% vs basic training", improvementVsBasic)
	} else if improvementVsBasic > 10.0 {
		t.Logf("✅ GOOD IMPROVEMENT: +%.1f%% vs basic training", improvementVsBasic)
	} else if improvementVsBasic > 0 {
		t.Logf("⚠️  SLIGHT IMPROVEMENT: +%.1f%% vs basic training", improvementVsBasic)
	} else {
		t.Logf("❌ NO IMPROVEMENT: %.1f%% vs basic training", improvementVsBasic)
	}

	t.Log(strings.Repeat("=", 60))
}

// trainWithProperSupervision uses the same supervision method as the successful 70% serial XOR test
func trainWithProperSupervision(matrix *extracellular.ExtracellularMatrix, circuit *RuleLearningCircuit, sequences []TemporalXORSequence, epochs int, t *testing.T) float64 {
	correctTraining := 0
	totalTraining := 0

	for epoch := 0; epoch < epochs; epoch++ {
		if epoch%10 == 0 {
			t.Logf("Training epoch %d/%d", epoch, epochs)
		}

		for _, sequence := range sequences {
			for step, bit := range sequence.bits {
				totalTraining++
				expectedParity := sequence.parities[step]

				// Clear short-term state but preserve memory
				time.Sleep(20 * time.Millisecond)

				// Present current bit
				if bit == 1 {
					circuit.inputNeuron.Receive(types.NeuralSignal{
						Value:     1.5,
						Timestamp: time.Now(),
						SourceID:  "proper_training",
						TargetID:  circuit.inputNeuron.ID(),
					})
				}

				// Allow forward propagation
				time.Sleep(30 * time.Millisecond)

				// Get network prediction
				predicted := circuit.outputNeuron.GetActivityLevel()
				correct := (predicted > 0.5 && expectedParity == 1) || (predicted <= 0.5 && expectedParity == 0)

				if correct {
					correctTraining++
				}

				// Apply PROPER supervised learning (same as 70% success case)
				applyProperSupervisedLearning(matrix, circuit, expectedParity, predicted, t)

				// Allow error backpropagation and eligibility trace updates
				time.Sleep(25 * time.Millisecond)
			}

			// Reset memory between sequences
			time.Sleep(50 * time.Millisecond)
		}
	}

	return float64(correctTraining) / float64(totalTraining) * 100.0
}

// applyProperSupervisedLearning applies the same supervision as the successful serial XOR test
func applyProperSupervisedLearning(matrix *extracellular.ExtracellularMatrix, circuit *RuleLearningCircuit, expectedParity int, predicted float64, t *testing.T) {
	// Calculate specific error signal (same as serial XOR success)
	var error float64
	if expectedParity == 1 {
		error = 1.0 - predicted // Error magnitude for target=1
	} else {
		error = predicted - 0.0 // Error magnitude for target=0
	}

	errorMagnitude := absValue(error)

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

// absValue returns absolute value (helper function)
func absValue(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}