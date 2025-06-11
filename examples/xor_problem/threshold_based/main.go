// ============================================================================
// temporal-neuron/examples/xor_problem/inhibitory_circuit/main.go
// ============================================================================

// Simple 2-Neuron XOR Circuit with Detailed Debug Output
//

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
	fmt.Println("🧠 2-Neuron XOR Circuit")
	fmt.Println("=======================")
	fmt.Println()

	// Create the XOR circuit
	circuit := NewSimpleXORCircuit()
	defer circuit.Shutdown()

	// Start the circuit
	err := circuit.Start()
	if err != nil {
		log.Fatalf("Failed to start XOR circuit: %v", err)
	}

	fmt.Println("🚀 XOR Circuit started successfully!")
	fmt.Println()
	fmt.Println("Network Architecture:")
	fmt.Println("  OR Neuron:  threshold=0.8 (fires for A OR B)")
	fmt.Println("  AND Neuron: threshold=1.8 (fires for A AND B)")
	fmt.Println("  XOR Logic:  (A OR B) AND NOT (A AND B)")
	fmt.Println()

	// Test all XOR truth table entries
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

	fmt.Println("Testing XOR Truth Table:")
	fmt.Println("------------------------")

	allCorrect := true
	for i, test := range testCases {
		fmt.Printf("\n--- Test %d: %s ---\n", i+1, test.name)
		fmt.Printf("Inputs: A=%.1f, B=%.1f\n", test.inputA, test.inputB)

		result, duration, err := circuit.ComputeXOR(test.inputA, test.inputB)
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
		fmt.Printf("Computation time: %v\n", duration)

		// Brief pause between computations
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	if allCorrect {
		fmt.Println("🎉 SUCCESS: All XOR computations correct!")
		fmt.Println("✨ Non-linear problem solved with only 2 neurons!")
	} else {
		fmt.Println("❌ FAILED: Some XOR computations were incorrect")
	}

	fmt.Println()
	fmt.Println("Final Statistics:")
	stats := circuit.GetStatistics()
	for key, value := range stats {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

// ============================================================================
// SIMPLIFIED XOR CIRCUIT (NO GLIAL COORDINATION)
// ============================================================================

type SimpleXORCircuit struct {
	// Core neurons
	orNeuron  *neuron.Neuron
	andNeuron *neuron.Neuron

	// Output monitoring
	orOutput  chan neuron.FireEvent
	andOutput chan neuron.FireEvent

	// State tracking
	isRunning    bool
	computeCount int
	successCount int
	totalTime    time.Duration
}

func NewSimpleXORCircuit() *SimpleXORCircuit {
	// Create OR neuron (low threshold - fires easily)
	orNeuron := neuron.NewNeuron(
		"OR_gate",           // ID
		0.8,                 // Low threshold
		0.98,                // Slow decay
		10*time.Millisecond, // Refractory period
		1.0,                 // Fire factor
		0, 0,                // No homeostasis
	)

	// Create AND neuron (high threshold - needs both inputs)
	andNeuron := neuron.NewNeuron(
		"AND_gate",          // ID
		1.8,                 // High threshold
		0.98,                // Slow decay
		10*time.Millisecond, // Refractory period
		1.0,                 // Fire factor
		0, 0,                // No homeostasis
	)

	// Create output channels
	orOutput := make(chan neuron.FireEvent, 10)
	andOutput := make(chan neuron.FireEvent, 10)

	orNeuron.SetFireEventChannel(orOutput)
	andNeuron.SetFireEventChannel(andOutput)

	return &SimpleXORCircuit{
		orNeuron:  orNeuron,
		andNeuron: andNeuron,
		orOutput:  orOutput,
		andOutput: andOutput,
		isRunning: false,
	}
}

func (xor *SimpleXORCircuit) Start() error {
	if xor.isRunning {
		return fmt.Errorf("circuit already running")
	}

	// Start neurons
	go xor.orNeuron.Run()
	go xor.andNeuron.Run()

	// Allow startup time
	time.Sleep(20 * time.Millisecond)

	xor.isRunning = true
	return nil
}

func (xor *SimpleXORCircuit) ComputeXOR(inputA, inputB float64) (int, time.Duration, error) {
	if !xor.isRunning {
		return 0, 0, fmt.Errorf("circuit not running")
	}

	startTime := time.Now()
	xor.computeCount++

	fmt.Printf("  🔄 Starting computation...\n")

	// Clear previous outputs
	xor.drainOutputs()
	fmt.Printf("  🧹 Cleared previous outputs\n")

	// Show initial neuron states
	orState := xor.orNeuron.GetNeuronState()
	andState := xor.andNeuron.GetNeuronState()
	fmt.Printf("  📊 Initial states:\n")
	fmt.Printf("     OR:  acc=%.3f, thresh=%.1f\n", orState["accumulator"], orState["threshold"])
	fmt.Printf("     AND: acc=%.3f, thresh=%.1f\n", andState["accumulator"], andState["threshold"])

	// Create and send input messages
	inputTime := time.Now()
	fmt.Printf("  📨 Sending inputs at %v\n", inputTime.Format("15:04:05.000"))

	// Send to OR neuron
	if inputA > 0 {
		xor.orNeuron.Receive(synapse.SynapseMessage{
			Value:     inputA,
			Timestamp: inputTime,
			SourceID:  "external_A",
		})
		fmt.Printf("     OR ← A (%.1f)\n", inputA)
	}

	if inputB > 0 {
		xor.orNeuron.Receive(synapse.SynapseMessage{
			Value:     inputB,
			Timestamp: inputTime,
			SourceID:  "external_B",
		})
		fmt.Printf("     OR ← B (%.1f)\n", inputB)
	}

	// Send to AND neuron
	if inputA > 0 {
		xor.andNeuron.Receive(synapse.SynapseMessage{
			Value:     inputA,
			Timestamp: inputTime,
			SourceID:  "external_A",
		})
		fmt.Printf("     AND ← A (%.1f)\n", inputA)
	}

	if inputB > 0 {
		xor.andNeuron.Receive(synapse.SynapseMessage{
			Value:     inputB,
			Timestamp: inputTime,
			SourceID:  "external_B",
		})
		fmt.Printf("     AND ← B (%.1f)\n", inputB)
	}

	// Wait for processing
	fmt.Printf("  ⏱️  Waiting for neural processing...\n")
	time.Sleep(30 * time.Millisecond)

	// Check final states
	orStateFinal := xor.orNeuron.GetNeuronState()
	andStateFinal := xor.andNeuron.GetNeuronState()
	fmt.Printf("  📊 Final states:\n")
	fmt.Printf("     OR:  acc=%.3f, thresh=%.1f\n", orStateFinal["accumulator"], orStateFinal["threshold"])
	fmt.Printf("     AND: acc=%.3f, thresh=%.1f\n", andStateFinal["accumulator"], andStateFinal["threshold"])

	// Check for firing events
	orFired := xor.checkFiring(xor.orOutput, "OR")
	andFired := xor.checkFiring(xor.andOutput, "AND")

	// XOR logic: (A OR B) AND NOT (A AND B)
	result := 0
	if orFired && !andFired {
		result = 1
	}

	// Count success based on expected result, not just when result = 1
	expected := 0
	if (inputA > 0.5) != (inputB > 0.5) { // XOR logic for expected result
		expected = 1
	}

	if result == expected {
		xor.successCount++
	}

	fmt.Printf("  🧮 Logic evaluation:\n")
	fmt.Printf("     OR fired:  %t\n", orFired)
	fmt.Printf("     AND fired: %t\n", andFired)
	fmt.Printf("     XOR = (OR AND NOT AND) = (%t AND %t) = %t → %d\n",
		orFired, !andFired, orFired && !andFired, result)

	duration := time.Since(startTime)
	xor.totalTime += duration

	return result, duration, nil
}

func (xor *SimpleXORCircuit) drainOutputs() {
	orCount := 0
	andCount := 0

	for len(xor.orOutput) > 0 {
		<-xor.orOutput
		orCount++
	}
	for len(xor.andOutput) > 0 {
		<-xor.andOutput
		andCount++
	}

	if orCount > 0 || andCount > 0 {
		fmt.Printf("     Drained %d OR events, %d AND events\n", orCount, andCount)
	}
}

func (xor *SimpleXORCircuit) checkFiring(output chan neuron.FireEvent, name string) bool {
	eventCount := 0
	fired := false

	for len(output) > 0 {
		event := <-output
		eventCount++
		fired = true
		fmt.Printf("     %s fired! Event %d: value=%.2f at %v\n",
			name, eventCount, event.Value, event.Timestamp.Format("15:04:05.000"))
	}

	if !fired {
		fmt.Printf("     %s did not fire\n", name)
	}

	return fired
}

func (xor *SimpleXORCircuit) GetStatistics() map[string]interface{} {
	avgTime := time.Duration(0)
	successRate := 0.0

	if xor.computeCount > 0 {
		avgTime = xor.totalTime / time.Duration(xor.computeCount)
		successRate = float64(xor.successCount) / float64(xor.computeCount) * 100
	}

	return map[string]interface{}{
		"Total Computations": xor.computeCount,
		"Successful XORs":    xor.successCount,
		"Success Rate":       fmt.Sprintf("%.1f%%", successRate),
		"Average Time":       avgTime,
		"Total Time":         xor.totalTime,
		"OR Neuron Status":   "Running",
		"AND Neuron Status":  "Running",
	}
}

func (xor *SimpleXORCircuit) Shutdown() {
	if !xor.isRunning {
		return
	}

	fmt.Println("\n🛑 Shutting down XOR circuit...")

	if xor.orNeuron != nil {
		xor.orNeuron.Close()
		fmt.Println("  ✅ OR neuron stopped")
	}
	if xor.andNeuron != nil {
		xor.andNeuron.Close()
		fmt.Println("  ✅ AND neuron stopped")
	}

	xor.isRunning = false
	fmt.Println("✅ Circuit shutdown complete")
}
