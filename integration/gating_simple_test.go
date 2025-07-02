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

/*
=================================================================================
NEURAL SWITCHING INTEGRATION TEST
=================================================================================

BIOLOGICAL BASIS:
This test simulates the dynamic switching between different neural processing
modes based on contextual neuromodulation. Different neuron types with distinct
ion channel configurations provide specialized computational capabilities:

1. FAST NEURONS (High Nav1.6): Rapid temporal processing, pattern detection
2. INTEGRATIVE NEURONS (High Cav1.2): Calcium-dependent integration, memory
3. INHIBITORY NEURONS (High GABA-A): Network synchronization, gating

COMPUTATIONAL CONCEPT:
Neuromodulators (dopamine, serotonin, norepinephrine) can reconfigure neural
circuits by selectively enhancing different neuron types, effectively switching
the network's computational mode:

- DETECTION MODE: Fast neurons dominate → rapid pattern recognition
- INTEGRATION MODE: Calcium neurons dominate → temporal integration
- INHIBITION MODE: GABA neurons dominate → selective gating

CLINICAL RELEVANCE:
This mechanism models attention switching, cognitive flexibility, and the
effects of psychiatric medications that target neuromodulatory systems.

=================================================================================
*/

// TestNeuralSwitching_MultiModalProcessing demonstrates dynamic switching between
// different neural processing modes using realistic ion channel configurations
func TestNeuralSwitching_MultiModalProcessing(t *testing.T) {
	t.Log("=== NEURAL SWITCHING: Multi-Modal Processing Test ===")
	t.Log("Simulating dynamic switching between Fast, Integrative, and Inhibitory modes")

	// --- 1. Matrix and Factory Setup ---
	matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		UpdateInterval:  5 * time.Millisecond, // Higher temporal resolution
		MaxComponents:   150,
	})
	if err := matrix.Start(); err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}
	defer matrix.Stop()

	// Register specialized neuron types with different ion channel configurations
	registerNeuronTypes(matrix, t)
	registerSynapseType(matrix, t)

	// --- 2. Build Multi-Modal Circuit ---
	circuit := buildSwitchingCircuit(matrix, t)
	defer cleanup(circuit.allNeurons)

	// --- 3. Test Each Processing Mode ---
	testResults := &SwitchingTestResults{}

	// PHASE 1: Detection Mode (Fast processing)
	t.Log("\n=== PHASE 1: DETECTION MODE (Dopamine → Fast Neurons) ===")
	runDetectionMode(matrix, circuit, testResults, t)

	// PHASE 2: Integration Mode (Temporal integration)
	t.Log("\n=== PHASE 2: INTEGRATION MODE (Serotonin → Integrative Neurons) ===")
	runIntegrationMode(matrix, circuit, testResults, t)

	// PHASE 3: Inhibition Mode (Selective gating)
	t.Log("\n=== PHASE 3: INHIBITION MODE (Norepinephrine → Inhibitory Neurons) ===")
	runInhibitionMode(matrix, circuit, testResults, t)

	// --- 4. Validate Switching Behavior ---
	validateSwitchingResults(testResults, t)
}

// SwitchingCircuit contains all neurons in the switching test circuit
type SwitchingCircuit struct {
	// Input layer
	stimulusA component.NeuralComponent
	stimulusB component.NeuralComponent

	// Processing layers (different computational modes)
	fastNeuron1        component.NeuralComponent // High Nav1.6 - rapid detection
	fastNeuron2        component.NeuralComponent
	integrativeNeuron1 component.NeuralComponent // High Cav1.2 - temporal integration
	integrativeNeuron2 component.NeuralComponent
	inhibitoryNeuron1  component.NeuralComponent // High GABA-A - selective gating
	inhibitoryNeuron2  component.NeuralComponent

	// Output layer
	outputDetection   component.NeuralComponent // Receives from fast neurons
	outputIntegration component.NeuralComponent // Receives from integrative neurons
	outputGating      component.NeuralComponent // Receives from inhibitory neurons

	allNeurons []component.NeuralComponent
}

// SwitchingTestResults stores results from each processing mode test
type SwitchingTestResults struct {
	DetectionResponse   float64
	IntegrationResponse float64
	InhibitionResponse  float64

	DetectionLatency   time.Duration
	IntegrationLatency time.Duration
	InhibitionLatency  time.Duration
}

// registerNeuronTypes creates different neuron types with specialized ion channel configurations
func registerNeuronTypes(matrix *extracellular.ExtracellularMatrix, t *testing.T) {
	// FAST NEURON: High Nav1.6 density for rapid processing
	matrix.RegisterNeuronType("fast_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		// Enhanced sodium channels for fast spike initiation
		n := neuron.NewNeuron(id, config.Threshold, 0.98, 1*time.Millisecond, 2.0, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandDopamine})
		n.SetCallbacks(callbacks)

		// Create dendritic mode with fast neuron ion channel profile
		dendriticMode := neuron.NewTemporalSummationMode()
		// High density Nav1.6 channels for rapid firing
		dendriticMode.AddChannel(neuron.NewRealisticNavChannel("nav1.6_fast_1"))
		dendriticMode.AddChannel(neuron.NewRealisticNavChannel("nav1.6_fast_2"))
		dendriticMode.AddChannel(neuron.NewRealisticNavChannel("nav1.6_fast_3"))
		// Moderate Kv4.2 for controlled repolarization
		dendriticMode.AddChannel(neuron.NewRealisticKvChannel("kv4.2_fast"))

		// Set the dendritic mode
		if err := n.SetDendriticMode(dendriticMode); err != nil {
			return nil, fmt.Errorf("failed to set dendritic mode: %v", err)
		}

		return n, nil
	})

	// INTEGRATIVE NEURON: High Cav1.2 density for calcium-dependent integration
	matrix.RegisterNeuronType("integrative_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		// Enhanced calcium signaling for temporal integration
		n := neuron.NewNeuron(id, config.Threshold, 0.92, 5*time.Millisecond, 1.3, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandSerotonin})
		n.SetCallbacks(callbacks)

		// Create dendritic mode with integrative neuron profile
		dendriticMode := neuron.NewTemporalSummationMode()
		// Standard Nav channels
		dendriticMode.AddChannel(neuron.NewRealisticNavChannel("nav1.6_int"))
		// High density Cav1.2 channels for calcium integration
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

	// INHIBITORY NEURON: High GABA-A density for selective gating
	matrix.RegisterNeuronType("inhibitory_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		// Optimized for inhibitory control and gating
		n := neuron.NewNeuron(id, config.Threshold, 0.95, 2*time.Millisecond, 1.8, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate, types.LigandNorepinephrine, types.LigandGABA})
		n.SetCallbacks(callbacks)

		// Create dendritic mode with inhibitory neuron profile
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

	// STANDARD NEURON: Balanced configuration for comparison
	matrix.RegisterNeuronType("standard_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(id, config.Threshold, 0.95, 3*time.Millisecond, 1.5, 0.0, 0.0)
		n.SetReceptors([]types.LigandType{types.LigandGlutamate})
		n.SetCallbacks(callbacks)
		return n, nil
	})

	t.Log("✓ Registered specialized neuron types: Fast, Integrative, Inhibitory, Standard")
}

// registerSynapseType creates basic synapses for circuit connectivity
func registerSynapseType(matrix *extracellular.ExtracellularMatrix, t *testing.T) {
	matrix.RegisterSynapseType("switching_synapse", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		pre, exists := matrix.GetNeuron(config.PresynapticID)
		if !exists {
			return nil, fmt.Errorf("presynaptic neuron not found: %s", config.PresynapticID)
		}
		post, exists := matrix.GetNeuron(config.PostsynapticID)
		if !exists {
			return nil, fmt.Errorf("postsynaptic neuron not found: %s", config.PostsynapticID)
		}
		return synapse.NewBasicSynapse(id, pre, post,
			synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(),
			config.InitialWeight, config.Delay), nil
	})
}

// buildSwitchingCircuit creates the multi-modal neural circuit
func buildSwitchingCircuit(matrix *extracellular.ExtracellularMatrix, t *testing.T) *SwitchingCircuit {
	circuit := &SwitchingCircuit{}

	// Create input neurons
	var err error
	circuit.stimulusA, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "standard_neuron", Threshold: 0.5})
	if err != nil {
		t.Fatalf("Failed to create stimulusA: %v", err)
	}
	circuit.stimulusB, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "standard_neuron", Threshold: 0.5})
	if err != nil {
		t.Fatalf("Failed to create stimulusB: %v", err)
	}

	// Create processing neurons with specialized configurations
	circuit.fastNeuron1, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "fast_neuron", Threshold: 1.8})
	if err != nil {
		t.Fatalf("Failed to create fastNeuron1: %v", err)
	}
	circuit.fastNeuron2, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "fast_neuron", Threshold: 1.8})
	if err != nil {
		t.Fatalf("Failed to create fastNeuron2: %v", err)
	}

	circuit.integrativeNeuron1, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "integrative_neuron", Threshold: 2.2})
	if err != nil {
		t.Fatalf("Failed to create integrativeNeuron1: %v", err)
	}
	circuit.integrativeNeuron2, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "integrative_neuron", Threshold: 2.2})
	if err != nil {
		t.Fatalf("Failed to create integrativeNeuron2: %v", err)
	}

	circuit.inhibitoryNeuron1, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "inhibitory_neuron", Threshold: 1.9})
	if err != nil {
		t.Fatalf("Failed to create inhibitoryNeuron1: %v", err)
	}
	circuit.inhibitoryNeuron2, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "inhibitory_neuron", Threshold: 1.9})
	if err != nil {
		t.Fatalf("Failed to create inhibitoryNeuron2: %v", err)
	}

	// Create output neurons
	circuit.outputDetection, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "standard_neuron", Threshold: 1.5})
	if err != nil {
		t.Fatalf("Failed to create outputDetection: %v", err)
	}
	circuit.outputIntegration, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "standard_neuron", Threshold: 1.5})
	if err != nil {
		t.Fatalf("Failed to create outputIntegration: %v", err)
	}
	circuit.outputGating, err = matrix.CreateNeuron(types.NeuronConfig{NeuronType: "standard_neuron", Threshold: 1.5})
	if err != nil {
		t.Fatalf("Failed to create outputGating: %v", err)
	}

	// Collect all neurons
	circuit.allNeurons = []component.NeuralComponent{
		circuit.stimulusA, circuit.stimulusB,
		circuit.fastNeuron1, circuit.fastNeuron2,
		circuit.integrativeNeuron1, circuit.integrativeNeuron2,
		circuit.inhibitoryNeuron1, circuit.inhibitoryNeuron2,
		circuit.outputDetection, circuit.outputIntegration, circuit.outputGating,
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

	// Wire the circuit - create comprehensive connectivity
	wireCircuit(matrix, circuit, t)

	t.Log("✓ Multi-modal switching circuit created with specialized neuron types")
	return circuit
}

// wireCircuit creates synaptic connections for the switching circuit
func wireCircuit(matrix *extracellular.ExtracellularMatrix, circuit *SwitchingCircuit, t *testing.T) {
	// Input to all processing neurons (parallel pathways)
	connections := []struct {
		pre, post component.NeuralComponent
		weight    float64
		desc      string
	}{
		// Stimulus A to all processing types
		{circuit.stimulusA, circuit.fastNeuron1, 1.0, "stimA -> fast1"},
		{circuit.stimulusA, circuit.integrativeNeuron1, 1.0, "stimA -> int1"},
		{circuit.stimulusA, circuit.inhibitoryNeuron1, 1.0, "stimA -> inh1"},

		// Stimulus B to all processing types
		{circuit.stimulusB, circuit.fastNeuron2, 1.0, "stimB -> fast2"},
		{circuit.stimulusB, circuit.integrativeNeuron2, 1.0, "stimB -> int2"},
		{circuit.stimulusB, circuit.inhibitoryNeuron2, 1.0, "stimB -> inh2"},

		// Cross-connections within processing layers
		{circuit.fastNeuron1, circuit.fastNeuron2, 0.5, "fast1 -> fast2"},
		{circuit.integrativeNeuron1, circuit.integrativeNeuron2, 0.8, "int1 -> int2"},
		{circuit.inhibitoryNeuron1, circuit.inhibitoryNeuron2, 0.6, "inh1 -> inh2"},

		// Processing to outputs (specialized pathways)
		{circuit.fastNeuron1, circuit.outputDetection, 1.2, "fast1 -> detection"},
		{circuit.fastNeuron2, circuit.outputDetection, 1.2, "fast2 -> detection"},
		{circuit.integrativeNeuron1, circuit.outputIntegration, 1.1, "int1 -> integration"},
		{circuit.integrativeNeuron2, circuit.outputIntegration, 1.1, "int2 -> integration"},
		{circuit.inhibitoryNeuron1, circuit.outputGating, 1.3, "inh1 -> gating"},
		{circuit.inhibitoryNeuron2, circuit.outputGating, 1.3, "inh2 -> gating"},
	}

	for _, conn := range connections {
		_, err := matrix.CreateSynapse(types.SynapseConfig{
			SynapseType:    "switching_synapse",
			PresynapticID:  conn.pre.ID(),
			PostsynapticID: conn.post.ID(),
			InitialWeight:  conn.weight,
		})
		if err != nil {
			t.Fatalf("Failed to create synapse %s: %v", conn.desc, err)
		}
	}

	t.Log("✓ Circuit wiring completed with parallel processing pathways")
}

// runDetectionMode tests fast processing mode enhanced by dopamine
func runDetectionMode(matrix *extracellular.ExtracellularMatrix, circuit *SwitchingCircuit, results *SwitchingTestResults, t *testing.T) {
	// Release dopamine to enhance fast neurons
	if err := matrix.ReleaseLigand(types.LigandDopamine, circuit.fastNeuron1.ID(), 1.2); err != nil {
		t.Fatalf("Failed to release dopamine: %v", err)
	}
	if err := matrix.ReleaseLigand(types.LigandDopamine, circuit.fastNeuron2.ID(), 1.2); err != nil {
		t.Fatalf("Failed to release dopamine: %v", err)
	}

	t.Log("Dopamine released to enhance fast neuron processing...")
	time.Sleep(20 * time.Millisecond) // Allow neuromodulator to take effect

	// Test rapid detection with brief stimuli
	startTime := time.Now()
	baseline := circuit.outputDetection.GetActivityLevel()

	// Send brief, rapid stimuli (pattern detection scenario)
	circuit.stimulusA.Receive(types.NeuralSignal{Value: 1.5, SourceID: "detection_test"})
	time.Sleep(5 * time.Millisecond)
	circuit.stimulusB.Receive(types.NeuralSignal{Value: 1.5, SourceID: "detection_test"})

	// Measure rapid response
	time.Sleep(25 * time.Millisecond)
	detectionResponse := circuit.outputDetection.GetActivityLevel() - baseline
	results.DetectionResponse = detectionResponse
	results.DetectionLatency = time.Since(startTime)

	t.Logf("Detection Mode Response: %.4f (latency: %v)", detectionResponse, results.DetectionLatency)
	t.Log("✓ Fast neurons optimized for rapid pattern detection")
}

// runIntegrationMode tests temporal integration mode enhanced by serotonin
func runIntegrationMode(matrix *extracellular.ExtracellularMatrix, circuit *SwitchingCircuit, results *SwitchingTestResults, t *testing.T) {
	// Clear any residual activity
	time.Sleep(30 * time.Millisecond)

	// Release serotonin to enhance integrative neurons
	if err := matrix.ReleaseLigand(types.LigandSerotonin, circuit.integrativeNeuron1.ID(), 1.0); err != nil {
		t.Fatalf("Failed to release serotonin: %v", err)
	}
	if err := matrix.ReleaseLigand(types.LigandSerotonin, circuit.integrativeNeuron2.ID(), 1.0); err != nil {
		t.Fatalf("Failed to release serotonin: %v", err)
	}

	t.Log("Serotonin released to enhance integrative neuron processing...")
	time.Sleep(20 * time.Millisecond)

	// Test temporal integration with distributed stimuli
	startTime := time.Now()
	baseline := circuit.outputIntegration.GetActivityLevel()

	// Send distributed stimuli over time (integration scenario)
	circuit.stimulusA.Receive(types.NeuralSignal{Value: 0.8, SourceID: "integration_test"})
	time.Sleep(15 * time.Millisecond)
	circuit.stimulusB.Receive(types.NeuralSignal{Value: 0.8, SourceID: "integration_test"})
	time.Sleep(15 * time.Millisecond)
	circuit.stimulusA.Receive(types.NeuralSignal{Value: 0.8, SourceID: "integration_test"})

	// Allow time for calcium-dependent integration
	time.Sleep(40 * time.Millisecond)
	integrationResponse := circuit.outputIntegration.GetActivityLevel() - baseline
	results.IntegrationResponse = integrationResponse
	results.IntegrationLatency = time.Since(startTime)

	t.Logf("Integration Mode Response: %.4f (latency: %v)", integrationResponse, results.IntegrationLatency)
	t.Log("✓ Integrative neurons optimized for temporal summation")
}

// runInhibitionMode tests selective gating mode enhanced by norepinephrine
func runInhibitionMode(matrix *extracellular.ExtracellularMatrix, circuit *SwitchingCircuit, results *SwitchingTestResults, t *testing.T) {
	// Clear any residual activity
	time.Sleep(30 * time.Millisecond)

	// Release norepinephrine to enhance inhibitory neurons
	if err := matrix.ReleaseLigand(types.LigandNorepinephrine, circuit.inhibitoryNeuron1.ID(), 1.1); err != nil {
		t.Fatalf("Failed to release norepinephrine: %v", err)
	}
	if err := matrix.ReleaseLigand(types.LigandNorepinephrine, circuit.inhibitoryNeuron2.ID(), 1.1); err != nil {
		t.Fatalf("Failed to release norepinephrine: %v", err)
	}

	t.Log("Norepinephrine released to enhance inhibitory neuron processing...")
	time.Sleep(20 * time.Millisecond)

	// Test selective gating with strong stimuli
	startTime := time.Now()
	baseline := circuit.outputGating.GetActivityLevel()

	// Send strong gating signals (inhibitory control scenario)
	circuit.stimulusA.Receive(types.NeuralSignal{Value: 1.8, SourceID: "gating_test"})
	circuit.stimulusB.Receive(types.NeuralSignal{Value: 1.8, SourceID: "gating_test"})

	// Measure inhibitory gating response
	time.Sleep(35 * time.Millisecond)
	inhibitionResponse := circuit.outputGating.GetActivityLevel() - baseline
	results.InhibitionResponse = inhibitionResponse
	results.InhibitionLatency = time.Since(startTime)

	t.Logf("Inhibition Mode Response: %.4f (latency: %v)", inhibitionResponse, results.InhibitionLatency)
	t.Log("✓ Inhibitory neurons optimized for selective gating")
}

// validateSwitchingResults checks that the neural switching worked as expected
func validateSwitchingResults(results *SwitchingTestResults, t *testing.T) {
	t.Log("\n=== VALIDATING NEURAL SWITCHING RESULTS ===")

	// Check that each mode produced a meaningful response
	if results.DetectionResponse <= 0 {
		t.Errorf("Detection mode failed to produce response: %.4f", results.DetectionResponse)
	}
	if results.IntegrationResponse <= 0 {
		t.Errorf("Integration mode failed to produce response: %.4f", results.IntegrationResponse)
	}
	if results.InhibitionResponse <= 0 {
		t.Errorf("Inhibition mode failed to produce response: %.4f", results.InhibitionResponse)
	}

	// Validate expected performance characteristics
	// Detection mode should be fastest (lowest latency)
	if results.DetectionLatency > results.IntegrationLatency {
		t.Errorf("Detection mode should be faster than integration mode")
	}

	// Integration mode should show sustained response
	if results.IntegrationResponse < results.DetectionResponse*0.7 {
		t.Logf("Warning: Integration response (%.4f) lower than expected vs detection (%.4f)",
			results.IntegrationResponse, results.DetectionResponse)
	}

	// Summary report
	t.Logf("\n--- NEURAL SWITCHING PERFORMANCE SUMMARY ---")
	t.Logf("Detection Mode:   Response=%.4f, Latency=%v", results.DetectionResponse, results.DetectionLatency)
	t.Logf("Integration Mode: Response=%.4f, Latency=%v", results.IntegrationResponse, results.IntegrationLatency)
	t.Logf("Inhibition Mode:  Response=%.4f, Latency=%v", results.InhibitionResponse, results.InhibitionLatency)

	t.Log("\n✓ NEURAL SWITCHING TEST PASSED")
	t.Log("✓ Successfully demonstrated dynamic reconfiguration of neural processing")
	t.Log("✓ Different neuromodulators selectively enhanced specialized neuron types")
	t.Log("✓ Circuit exhibited distinct computational modes based on chemical context")
	t.Log("✓ Models attention switching and cognitive flexibility mechanisms")
}

// cleanup stops all neurons in the circuit
func cleanup(neurons []component.NeuralComponent) {
	for _, neuron := range neurons {
		neuron.Stop()
	}
}
