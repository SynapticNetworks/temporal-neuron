// ============================================================================
// temporal-neuron/examples/xor_problem/biological_circuit/main.go
// ============================================================================

// Biologically Realistic XOR Circuit
// Demonstrates key differences from traditional ANNs:
// - Real inhibitory neurons (GABAergic-like)
// - Dale's Principle (neurons are E or I, not both)
// - Realistic synaptic delays and weights
// - Emergent XOR from biological connectivity patterns
// - No artificial activation functions or batch processing

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/neuron"
	"github.com/SynapticNetworks/temporal-neuron/synapse"
)

func main() {
	fmt.Println("🧬 Biologically Realistic XOR Circuit")
	fmt.Println("=====================================")
	fmt.Println()
	fmt.Println("🔬 Key Biological Features:")
	fmt.Println("  • Real inhibitory interneurons (GABAergic-like)")
	fmt.Println("  • Dale's Principle: E neurons vs I neurons")
	fmt.Println("  • Realistic synaptic delays and weights")
	fmt.Println("  • Emergent XOR from biological connectivity")
	fmt.Println("  • No activation functions or batch processing")
	fmt.Println()

	// Create the biological XOR circuit
	circuit := NewBiologicalXORCircuit()
	defer circuit.Shutdown()

	// Start the living neural network
	err := circuit.Start()
	if err != nil {
		log.Fatalf("Failed to start biological circuit: %v", err)
	}

	fmt.Println("🚀 Biological neural network is now ALIVE!")
	fmt.Println()

	// Print network architecture
	circuit.PrintArchitecture()

	// Test XOR function with biological dynamics
	testCases := []struct {
		name     string
		inputA   float64
		inputB   float64
		expected int
	}{
		{"0 XOR 0", 0.0, 0.0, 0},
		{"0 XOR 1", 0.0, 1.0, 1},
		{"1 XOR 0", 1.0, 0.0, 1},
		{"1 XOR 1", 1.0, 1.0, 0},
	}

	fmt.Println("🧪 Testing Biological XOR Computation:")
	fmt.Println("--------------------------------------")

	allCorrect := true
	for i, test := range testCases {
		fmt.Printf("\n--- Test %d: %s ---\n", i+1, test.name)

		result, dynamics, err := circuit.ComputeBiologicalXOR(test.inputA, test.inputB)
		if err != nil {
			fmt.Printf("❌ Test %d failed: %v\n", i+1, err)
			allCorrect = false
			continue
		}

		status := "✅"
		if result != test.expected {
			status = "❌"
			allCorrect = false
		}

		fmt.Printf("Expected: %d, Got: %d %s\n", test.expected, result, status)
		fmt.Printf("🧬 Biological Dynamics:\n")
		fmt.Printf("  • E1 fired: %t, E2 fired: %t\n", dynamics.E1Fired, dynamics.E2Fired)
		fmt.Printf("  • Inhibition active: %t\n", dynamics.InhibitionActive)
		fmt.Printf("  • Output firing: %t\n", dynamics.OutputFired)
		fmt.Printf("  • Processing time: %v\n", dynamics.ProcessingTime)
		fmt.Printf("  • Peak inhibition: %.2f\n", dynamics.PeakInhibition)

		// Allow network to return to resting state
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	if allCorrect {
		fmt.Println("🎉 SUCCESS: Biological XOR circuit working perfectly!")
		fmt.Println("✨ Non-linear computation emerged from biological principles!")
	} else {
		fmt.Println("❌ FAILED: Some biological computations were incorrect")
	}

	// Show final statistics
	fmt.Println()
	circuit.PrintBiologicalStats()
}

// ============================================================================
// BIOLOGICALLY REALISTIC XOR CIRCUIT
// ============================================================================

type BiologicalXORCircuit struct {
	// Excitatory neurons (glutamatergic-like)
	excitatory1  *neuron.Neuron // Input processing neuron 1
	excitatory2  *neuron.Neuron // Input processing neuron 2
	outputNeuron *neuron.Neuron // Final output neuron

	// Inhibitory interneuron (GABAergic-like)
	inhibitory *neuron.Neuron

	// Synaptic connections with biological properties
	synapses map[string]*synapse.BasicSynapse

	// Event monitoring
	fireEvents map[string]chan neuron.FireEvent

	// Network state
	isRunning           bool
	computeCount        int
	successCount        int
	totalProcessingTime time.Duration
}

type BiologicalDynamics struct {
	E1Fired          bool
	E2Fired          bool
	InhibitionActive bool
	OutputFired      bool
	ProcessingTime   time.Duration
	PeakInhibition   float64
}

func NewBiologicalXORCircuit() *BiologicalXORCircuit {
	// Create excitatory neurons (glutamatergic-like)
	// These represent cortical pyramidal neurons
	excitatory1 := neuron.NewNeuron(
		"E1_pyramidal",     // ID
		0.8,                // Lower threshold - easier to fire
		0.95,               // Realistic membrane time constant (~20ms)
		5*time.Millisecond, // Cortical refractory period
		1.0,                // Standard spike amplitude
		0, 0,               // No homeostasis for simplicity
	)

	excitatory2 := neuron.NewNeuron(
		"E2_pyramidal",
		0.8, // Lower threshold - easier to fire
		0.95,
		5*time.Millisecond,
		1.0,
		0, 0,
	)

	outputNeuron := neuron.NewNeuron(
		"Output_pyramidal",
		0.8, // Lower threshold - should fire with single input
		0.96,
		6*time.Millisecond,
		1.0,
		0, 0,
	)

	// Create inhibitory interneuron (GABAergic-like)
	// These represent fast-spiking interneurons
	inhibitory := neuron.NewNeuron(
		"Inhibitory_interneuron",
		1.0,                // Higher threshold - needs both inputs
		0.92,               // Faster dynamics
		2*time.Millisecond, // Shorter refractory (interneurons are fast)
		1.0,
		0, 0,
	)

	// Create fire event channels for monitoring
	fireEvents := make(map[string]chan neuron.FireEvent)
	fireEvents["E1"] = make(chan neuron.FireEvent, 10)
	fireEvents["E2"] = make(chan neuron.FireEvent, 10)
	fireEvents["Inhibitory"] = make(chan neuron.FireEvent, 10)
	fireEvents["Output"] = make(chan neuron.FireEvent, 10)

	excitatory1.SetFireEventChannel(fireEvents["E1"])
	excitatory2.SetFireEventChannel(fireEvents["E2"])
	inhibitory.SetFireEventChannel(fireEvents["Inhibitory"])
	outputNeuron.SetFireEventChannel(fireEvents["Output"])

	return &BiologicalXORCircuit{
		excitatory1:  excitatory1,
		excitatory2:  excitatory2,
		outputNeuron: outputNeuron,
		inhibitory:   inhibitory,
		synapses:     make(map[string]*synapse.BasicSynapse),
		fireEvents:   fireEvents,
		isRunning:    false,
	}
}

func (circuit *BiologicalXORCircuit) Start() error {
	if circuit.isRunning {
		return fmt.Errorf("biological circuit already running")
	}

	// Create biologically realistic synaptic connections
	circuit.createBiologicalConnections()

	// Start all neurons (they become autonomous living entities)
	go circuit.excitatory1.Run()
	go circuit.excitatory2.Run()
	go circuit.inhibitory.Run()
	go circuit.outputNeuron.Run()

	// Allow biological startup time
	time.Sleep(20 * time.Millisecond)

	circuit.isRunning = true
	return nil
}

func (circuit *BiologicalXORCircuit) createBiologicalConnections() {
	// Standard STDP and pruning configurations
	stdpConfig := synapse.CreateDefaultSTDPConfig()
	pruningConfig := synapse.CreateDefaultPruningConfig()

	// E1 → Output (excitatory connection)
	// Models glutamatergic synapses from pyramidal to pyramidal
	circuit.synapses["E1→Output"] = synapse.NewBasicSynapse(
		"E1_to_Output",
		circuit.excitatory1, circuit.outputNeuron,
		stdpConfig, pruningConfig,
		+0.9,               // Positive weight (excitatory)
		4*time.Millisecond, // Slower - allows inhibition to arrive first
	)
	circuit.excitatory1.AddOutputSynapse("to_output", circuit.synapses["E1→Output"])

	// E2 → Output (excitatory connection)
	circuit.synapses["E2→Output"] = synapse.NewBasicSynapse(
		"E2_to_Output",
		circuit.excitatory2, circuit.outputNeuron,
		stdpConfig, pruningConfig,
		+0.9,               // Positive weight (excitatory)
		4*time.Millisecond, // Slower - allows inhibition to arrive first
	)
	circuit.excitatory2.AddOutputSynapse("to_output", circuit.synapses["E2→Output"])

	// E1 → Inhibitory (excitatory to interneuron)
	// Models pyramidal → interneuron connections
	circuit.synapses["E1→Inh"] = synapse.NewBasicSynapse(
		"E1_to_Inhibitory",
		circuit.excitatory1, circuit.inhibitory,
		stdpConfig, pruningConfig,
		+0.7,               // Moderate excitatory weight
		1*time.Millisecond, // Fast local connection
	)
	circuit.excitatory1.AddOutputSynapse("to_inhibitory", circuit.synapses["E1→Inh"])

	// E2 → Inhibitory (excitatory to interneuron)
	circuit.synapses["E2→Inh"] = synapse.NewBasicSynapse(
		"E2_to_Inhibitory",
		circuit.excitatory2, circuit.inhibitory,
		stdpConfig, pruningConfig,
		+0.7,
		1*time.Millisecond,
	)
	circuit.excitatory2.AddOutputSynapse("to_inhibitory", circuit.synapses["E2→Inh"])

	// Inhibitory → Output (GABAergic inhibition)
	// This is the KEY biological difference: real inhibitory synapses!
	inhibStdpConfig := stdpConfig
	inhibStdpConfig.MinWeight = -2.0 // Allow negative weights
	inhibStdpConfig.MaxWeight = 0.0  // Keep it inhibitory

	circuit.synapses["Inh→Output"] = synapse.NewBasicSynapse(
		"Inhibitory_to_Output",
		circuit.inhibitory, circuit.outputNeuron,
		inhibStdpConfig, pruningConfig,
		-1.2,               // NEGATIVE weight (inhibitory/GABAergic)
		2*time.Millisecond, // Slightly delayed inhibition
	)
	circuit.inhibitory.AddOutputSynapse("to_output", circuit.synapses["Inh→Output"])
}

func (circuit *BiologicalXORCircuit) ComputeBiologicalXOR(inputA, inputB float64) (int, BiologicalDynamics, error) {
	if !circuit.isRunning {
		return 0, BiologicalDynamics{}, fmt.Errorf("biological circuit not running")
	}

	startTime := time.Now()
	circuit.computeCount++

	fmt.Printf("🧬 Biological computation starting...\n")

	// Clear previous activity
	circuit.drainAllEvents()

	// Send biological inputs with realistic timing
	inputTime := time.Now()
	fmt.Printf("📡 Sending biological inputs:\n")

	if inputA > 0.5 {
		circuit.excitatory1.Receive(synapse.SynapseMessage{
			Value:     inputA,
			Timestamp: inputTime,
			SourceID:  "sensory_input_A",
			SynapseID: "sensory_A",
		})
		fmt.Printf("  E1 ← A (%.1f) [glutamate-like]\n", inputA)
	}

	if inputB > 0.5 {
		circuit.excitatory2.Receive(synapse.SynapseMessage{
			Value:     inputB,
			Timestamp: inputTime,
			SourceID:  "sensory_input_B",
			SynapseID: "sensory_B",
		})
		fmt.Printf("  E2 ← B (%.1f) [glutamate-like]\n", inputB)
	}

	// Allow biological processing time
	// In real cortex: ~5-15ms for local circuit computation
	fmt.Printf("⏳ Allowing biological dynamics to unfold...\n")
	time.Sleep(25 * time.Millisecond)

	// Debug: Check neuron states after processing
	fmt.Printf("🔍 Post-processing neuron states:\n")
	e1State := circuit.excitatory1.GetNeuronState()
	e2State := circuit.excitatory2.GetNeuronState()
	inhState := circuit.inhibitory.GetNeuronState()
	outState := circuit.outputNeuron.GetNeuronState()

	fmt.Printf("  E1: acc=%.3f, thresh=%.1f, calcium=%.3f\n",
		e1State["accumulator"], e1State["threshold"], e1State["calciumLevel"])
	fmt.Printf("  E2: acc=%.3f, thresh=%.1f, calcium=%.3f\n",
		e2State["accumulator"], e2State["threshold"], e2State["calciumLevel"])
	fmt.Printf("  Inh: acc=%.3f, thresh=%.1f, calcium=%.3f\n",
		inhState["accumulator"], inhState["threshold"], inhState["calciumLevel"])
	fmt.Printf("  Out: acc=%.3f, thresh=%.1f, calcium=%.3f\n",
		outState["accumulator"], outState["threshold"], outState["calciumLevel"])

	// Analyze biological dynamics
	dynamics := circuit.analyzeBiologicalDynamics(startTime)

	// XOR logic emerges from biological circuit:
	// Output fires when: (E1 OR E2) AND NOT (strong inhibition)
	// Inhibition is strong when: both E1 AND E2 fire → inhibitory fires
	result := 0
	if dynamics.OutputFired {
		result = 1
	}

	// Validate against expected XOR
	expected := 0
	if (inputA > 0.5) != (inputB > 0.5) { // XOR logic
		expected = 1
	}

	if result == expected {
		circuit.successCount++
	}

	circuit.totalProcessingTime += dynamics.ProcessingTime

	fmt.Printf("🧮 Biological XOR logic:\n")
	fmt.Printf("  E1 fired: %t, E2 fired: %t\n", dynamics.E1Fired, dynamics.E2Fired)
	fmt.Printf("  Both fired → Inhibition: %t\n", dynamics.InhibitionActive)
	fmt.Printf("  Output neuron fires: %t → XOR = %d\n", dynamics.OutputFired, result)

	return result, dynamics, nil
}

func (circuit *BiologicalXORCircuit) analyzeBiologicalDynamics(startTime time.Time) BiologicalDynamics {
	dynamics := BiologicalDynamics{
		ProcessingTime: time.Since(startTime),
	}

	// Check which neurons fired
	dynamics.E1Fired = circuit.checkFiringEvents("E1")
	dynamics.E2Fired = circuit.checkFiringEvents("E2")
	dynamics.InhibitionActive = circuit.checkFiringEvents("Inhibitory")
	dynamics.OutputFired = circuit.checkFiringEvents("Output")

	// Measure peak inhibition strength (biological detail)
	if dynamics.InhibitionActive {
		// When inhibitory neuron fires, it creates strong GABA-like inhibition
		dynamics.PeakInhibition = 1.2 // Matches inhibitory synapse weight
	}

	return dynamics
}

func (circuit *BiologicalXORCircuit) checkFiringEvents(neuronType string) bool {
	events := circuit.fireEvents[neuronType]
	fired := false
	eventCount := 0

	for len(events) > 0 {
		event := <-events
		fired = true
		eventCount++
		fmt.Printf("  🔥 %s fired (event %d): %.2f at %v\n",
			neuronType, eventCount, event.Value, event.Timestamp.Format("15:04:05.000"))
	}

	return fired
}

func (circuit *BiologicalXORCircuit) drainAllEvents() {
	for neuronType, events := range circuit.fireEvents {
		count := 0
		for len(events) > 0 {
			<-events
			count++
		}
		if count > 0 {
			fmt.Printf("  🧹 Cleared %d old %s events\n", count, neuronType)
		}
	}
}

func (circuit *BiologicalXORCircuit) PrintArchitecture() {
	fmt.Println("🏗️  Biological Network Architecture:")
	fmt.Println("    ")
	fmt.Println("    Sensory Inputs")
	fmt.Println("         │    │")
	fmt.Println("         ▼    ▼")
	fmt.Println("       E1 ──── E2    (Excitatory/Pyramidal neurons)")
	fmt.Println("        │ \\  / │")
	fmt.Println("        │  \\/  │     (Glutamatergic synapses)")
	fmt.Println("        │  /\\  │")
	fmt.Println("        ▼ /  \\ ▼")
	fmt.Println("       Output  Inhibitory")
	fmt.Println("          ▲      │")
	fmt.Println("          │      │     (GABAergic synapse)")
	fmt.Println("          └──────┘")
	fmt.Println()
	fmt.Println("🧬 Biological Features:")
	fmt.Println("  • Dale's Principle: E neurons (glutamate) vs I neurons (GABA)")
	fmt.Println("  • Realistic delays: 1-3ms (local cortical)")
	fmt.Println("  • Inhibitory feedback: Strong GABA-like suppression")
	fmt.Println("  • Emergent XOR: No explicit logic gates!")
	fmt.Println()

	// Print synaptic weights and delays
	fmt.Println("⚡ Synaptic Properties:")
	for name, syn := range circuit.synapses {
		weight := syn.GetWeight()
		delay := syn.GetDelay()
		synapseType := "Excitatory (glutamate-like)"
		if weight < 0 {
			synapseType = "Inhibitory (GABA-like)"
		}
		fmt.Printf("  %s: weight=%.2f, delay=%v [%s]\n",
			name, weight, delay, synapseType)
	}
	fmt.Println()
}

func (circuit *BiologicalXORCircuit) PrintBiologicalStats() {
	avgTime := time.Duration(0)
	successRate := 0.0

	if circuit.computeCount > 0 {
		avgTime = circuit.totalProcessingTime / time.Duration(circuit.computeCount)
		successRate = float64(circuit.successCount) / float64(circuit.computeCount) * 100
	}

	fmt.Println("📊 Biological Network Statistics:")
	fmt.Printf("  Total biological computations: %d\n", circuit.computeCount)
	fmt.Printf("  Successful XOR operations: %d\n", circuit.successCount)
	fmt.Printf("  Biological success rate: %.1f%%\n", successRate)
	fmt.Printf("  Average processing time: %v\n", avgTime)
	fmt.Printf("  Total computation time: %v\n", circuit.totalProcessingTime)
	fmt.Println()
	fmt.Println("🧬 Key Biological Insights:")
	fmt.Println("  • No activation functions - just biological thresholds")
	fmt.Println("  • No backpropagation - connectivity determines function")
	fmt.Println("  • Real inhibition creates the XOR behavior")
	fmt.Println("  • Emergent computation from biological principles")
}

func (circuit *BiologicalXORCircuit) Shutdown() {
	if !circuit.isRunning {
		return
	}

	fmt.Println("\n🛑 Shutting down biological neural network...")

	neurons := map[string]*neuron.Neuron{
		"E1":         circuit.excitatory1,
		"E2":         circuit.excitatory2,
		"Inhibitory": circuit.inhibitory,
		"Output":     circuit.outputNeuron,
	}

	for name, n := range neurons {
		if n != nil {
			n.Close()
			fmt.Printf("  ✅ %s neuron stopped\n", name)
		}
	}

	circuit.isRunning = false
	fmt.Println("✅ Biological circuit shutdown complete")
}
