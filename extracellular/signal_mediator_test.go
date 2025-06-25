package extracellular

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// =================================================================================
// TEST INFRASTRUCTURE: Mocks and Helpers
// =================================================================================

// mockSignalListener implements SignalListener for testing
type mockSignalListener struct {
	id             string
	receivedCount  int32
	receivedData   []interface{}
	receivedFrom   []string
	receivedSignal []SignalType
	mu             sync.Mutex
	lastSignalTime time.Time
}

func newMockListener(id string) *mockSignalListener {
	return &mockSignalListener{
		id:             id,
		receivedData:   make([]interface{}, 0),
		receivedFrom:   make([]string, 0),
		receivedSignal: make([]SignalType, 0),
	}
}

// ID implements SignalListener interface
func (m *mockSignalListener) ID() string {
	return m.id
}

// OnSignal implements SignalListener interface
func (m *mockSignalListener) OnSignal(signalType SignalType, sourceID string, data interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	atomic.AddInt32(&m.receivedCount, 1)
	m.receivedData = append(m.receivedData, data)
	m.receivedFrom = append(m.receivedFrom, sourceID)
	m.receivedSignal = append(m.receivedSignal, signalType)
	m.lastSignalTime = time.Now()
}

// Reset clears received signal data for new test cases
func (m *mockSignalListener) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	atomic.StoreInt32(&m.receivedCount, 0)
	m.receivedData = make([]interface{}, 0)
	m.receivedFrom = make([]string, 0)
	m.receivedSignal = make([]SignalType, 0)
}

// GetReceivedCount safely returns the number of received signals
func (m *mockSignalListener) GetReceivedCount() int {
	return int(atomic.LoadInt32(&m.receivedCount))
}

// GetLastReceivedData returns the most recent signal data
func (m *mockSignalListener) GetLastReceivedData() interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.receivedData) == 0 {
		return nil
	}
	return m.receivedData[len(m.receivedData)-1]
}

// GetLastReceivedFrom returns the source of the most recent signal
func (m *mockSignalListener) GetLastReceivedFrom() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.receivedFrom) == 0 {
		return ""
	}
	return m.receivedFrom[len(m.receivedFrom)-1]
}

// GetLastSignalType returns the type of the most recent signal
func (m *mockSignalListener) GetLastSignalType() SignalType {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.receivedSignal) == 0 {
		return SignalFired // Default value
	}
	return m.receivedSignal[len(m.receivedSignal)-1]
}

// GetAllReceivedData returns all received data
func (m *mockSignalListener) GetAllReceivedData() []interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]interface{}, len(m.receivedData))
	copy(result, m.receivedData)
	return result
}

// =================================================================================
// BASIC FUNCTIONALITY TESTS
// =================================================================================

func TestSignalMediatorCreation(t *testing.T) {
	t.Log("=== TESTING: SignalMediator Creation ===")

	mediator := NewSignalMediator()
	if mediator == nil {
		t.Fatal("FAIL: NewSignalMediator returned nil")
	}

	// Test initial state
	if count := mediator.GetSignalCount(); count != 0 {
		t.Errorf("FAIL: New mediator should have 0 signals, got %d", count)
	}

	history := mediator.GetRecentSignals(10)
	if len(history) != 0 {
		t.Errorf("FAIL: New mediator should have empty history, got %d events", len(history))
	}

	t.Log("✓ SignalMediator created successfully with correct initial state")
}

func TestSignalMediatorBasicListenerManagement(t *testing.T) {
	t.Log("=== TESTING: Basic Listener Registration and Removal ===")

	mediator := NewSignalMediator()
	listener1 := newMockListener("listener1")
	listener2 := newMockListener("listener2")

	// Add listeners for different signal types
	mediator.AddListener([]SignalType{SignalFired, SignalConnected}, listener1)
	mediator.AddListener([]SignalType{SignalFired}, listener2)

	// Test signal delivery
	t.Log("--- Sending SignalFired ---")
	mediator.Send(SignalFired, "source_neuron", "test_data")

	// Allow signal processing
	time.Sleep(10 * time.Millisecond)

	if listener1.GetReceivedCount() != 1 {
		t.Errorf("FAIL: Listener1 should have received 1 signal, got %d", listener1.GetReceivedCount())
	}
	if listener2.GetReceivedCount() != 1 {
		t.Errorf("FAIL: Listener2 should have received 1 signal, got %d", listener2.GetReceivedCount())
	}
	t.Log("✓ Both listeners received SignalFired correctly")

	// Test selective signal delivery
	t.Log("--- Sending SignalConnected ---")
	mediator.Send(SignalConnected, "source_neuron", "connection_data")
	time.Sleep(10 * time.Millisecond)

	if listener1.GetReceivedCount() != 2 {
		t.Errorf("FAIL: Listener1 should have received 2 signals total, got %d", listener1.GetReceivedCount())
	}
	if listener2.GetReceivedCount() != 1 {
		t.Errorf("FAIL: Listener2 should still have 1 signal, got %d", listener2.GetReceivedCount())
	}
	t.Log("✓ Selective signal delivery working correctly")

	// Test listener removal
	t.Log("--- Removing Listener1 from SignalFired ---")
	mediator.RemoveListener([]SignalType{SignalFired}, listener1)
	mediator.Send(SignalFired, "source_neuron", "third_signal")
	time.Sleep(10 * time.Millisecond)

	if listener1.GetReceivedCount() != 2 {
		t.Errorf("FAIL: Listener1 should not have received third signal, count is %d", listener1.GetReceivedCount())
	}
	if listener2.GetReceivedCount() != 2 {
		t.Errorf("FAIL: Listener2 should have received third signal, count is %d", listener2.GetReceivedCount())
	}
	t.Log("✓ Listener removal working correctly")
}

func TestSignalMediatorSelfSignalPrevention(t *testing.T) {
	t.Log("=== TESTING: Self-Signal Prevention ===")

	mediator := NewSignalMediator()
	listener := newMockListener("self_test_component")

	// Register listener
	mediator.AddListener([]SignalType{SignalFired}, listener)

	// Component sends signal with its own ID as source
	mediator.Send(SignalFired, "self_test_component", "self_signal")
	time.Sleep(10 * time.Millisecond)

	if listener.GetReceivedCount() != 0 {
		t.Errorf("FAIL: Component should not receive its own signal, but received %d", listener.GetReceivedCount())
	}

	// But it should receive signals from other sources
	mediator.Send(SignalFired, "other_component", "external_signal")
	time.Sleep(10 * time.Millisecond)

	if listener.GetReceivedCount() != 1 {
		t.Errorf("FAIL: Component should receive external signals, got %d", listener.GetReceivedCount())
	}

	t.Log("✓ Self-signal prevention working correctly")
}

// =================================================================================
// ELECTRICAL COUPLING TESTS (Gap Junctions)
// =================================================================================

func TestSignalMediatorElectricalCouplingBasics(t *testing.T) {
	t.Log("=== TESTING: Electrical Coupling (Gap Junctions) ===")

	mediator := NewSignalMediator()
	neuronA, neuronB := "neuronA", "neuronB"

	// Test establishing coupling
	t.Log("--- Establishing coupling between A and B ---")
	err := mediator.EstablishElectricalCoupling(neuronA, neuronB, 0.8)
	if err != nil {
		t.Fatalf("FAIL: Failed to establish coupling: %v", err)
	}

	// Verify bidirectional conductance
	conductanceAB := mediator.GetConductance(neuronA, neuronB)
	conductanceBA := mediator.GetConductance(neuronB, neuronA)

	if conductanceAB != 0.8 || conductanceBA != 0.8 {
		t.Errorf("FAIL: Expected bidirectional conductance of 0.8, got A->B: %.2f, B->A: %.2f",
			conductanceAB, conductanceBA)
	}
	t.Logf("✓ Bidirectional conductance of %.2f established", conductanceAB)

	// Verify coupling lists
	couplingsA := mediator.GetElectricalCouplings(neuronA)
	couplingsB := mediator.GetElectricalCouplings(neuronB)

	if len(couplingsA) != 1 || couplingsA[0] != neuronB {
		t.Errorf("FAIL: NeuronA should be coupled to NeuronB. Got: %v", couplingsA)
	}
	if len(couplingsB) != 1 || couplingsB[0] != neuronA {
		t.Errorf("FAIL: NeuronB should be coupled to NeuronA. Got: %v", couplingsB)
	}
	t.Log("✓ Bidirectional coupling lists correct")

	// Test coupling removal
	t.Log("--- Removing coupling ---")
	err = mediator.RemoveElectricalCoupling(neuronA, neuronB)
	if err != nil {
		t.Fatalf("FAIL: Failed to remove coupling: %v", err)
	}

	// Verify removal
	if mediator.GetConductance(neuronA, neuronB) != 0.0 {
		t.Error("FAIL: Conductance should be 0.0 after removal")
	}
	if len(mediator.GetElectricalCouplings(neuronA)) != 0 {
		t.Error("FAIL: Coupling list should be empty after removal")
	}
	t.Log("✓ Coupling removal successful")
}

func TestSignalMediatorElectricalCouplingEdgeCases(t *testing.T) {
	t.Log("=== TESTING: Electrical Coupling Edge Cases ===")

	mediator := NewSignalMediator()

	// Test invalid conductance values
	t.Log("--- Testing invalid conductance values ---")
	mediator.EstablishElectricalCoupling("n1", "n2", -1.0) // Negative
	if c := mediator.GetConductance("n1", "n2"); c != 0.5 {
		t.Errorf("FAIL: Negative conductance should default to 0.5, got %.2f", c)
	}

	mediator.RemoveElectricalCoupling("n1", "n2")
	mediator.EstablishElectricalCoupling("n1", "n2", 2.0) // Too high
	if c := mediator.GetConductance("n1", "n2"); c != 0.5 {
		t.Errorf("FAIL: Excessive conductance should default to 0.5, got %.2f", c)
	}
	t.Log("✓ Invalid conductance values handled correctly")

	// Test self-coupling
	t.Log("--- Testing self-coupling ---")
	mediator.EstablishElectricalCoupling("self", "self", 1.0)
	selfCouplings := mediator.GetElectricalCouplings("self")
	if len(selfCouplings) != 1 || selfCouplings[0] != "self" {
		t.Errorf("FAIL: Self-coupling should work. Got: %v", selfCouplings)
	}
	t.Log("✓ Self-coupling works correctly")

	// Test non-existent components
	t.Log("--- Testing non-existent components ---")
	couplings := mediator.GetElectricalCouplings("non_existent")
	if len(couplings) != 0 {
		t.Errorf("FAIL: Non-existent component should have empty coupling list, got: %v", couplings)
	}
	t.Log("✓ Non-existent component queries handled safely")
}

// =================================================================================
// SIGNAL HISTORY AND MONITORING TESTS
// =================================================================================

func TestSignalMediatorSignalHistoryManagement(t *testing.T) {
	t.Log("=== TESTING: Signal History Management ===")

	mediator := NewSignalMediator()
	mediator.maxHistory = 3 // Set small limit for testing

	// Send multiple signals
	for i := 0; i < 5; i++ {
		mediator.Send(SignalFired, fmt.Sprintf("source_%d", i), i)
	}

	// Test history size limit
	history := mediator.GetRecentSignals(10)
	if len(history) != 3 {
		t.Errorf("FAIL: History should be limited to 3, got %d", len(history))
	}
	t.Log("✓ History size limit enforced correctly")

	// Verify history contains most recent signals
	if data, ok := history[2].Data.(int); !ok || data != 4 {
		t.Errorf("FAIL: Most recent signal should have data 4, got %v", history[2].Data)
	}
	if data, ok := history[0].Data.(int); !ok || data != 2 {
		t.Errorf("FAIL: Oldest signal in history should have data 2, got %v", history[0].Data)
	}
	t.Log("✓ History contains correct recent signals")

	// Test signal count
	if count := mediator.GetSignalCount(); count != 3 {
		t.Errorf("FAIL: Signal count should be 3, got %d", count)
	}

	// Test history clearing
	mediator.ClearSignalHistory()
	if len(mediator.GetRecentSignals(10)) != 0 {
		t.Error("FAIL: History should be empty after clearing")
	}
	if mediator.GetSignalCount() != 0 {
		t.Error("FAIL: Signal count should be 0 after clearing")
	}
	t.Log("✓ History clearing works correctly")
}

func TestSignalMediatorSignalEventStructure(t *testing.T) {
	t.Log("=== TESTING: Signal Event Data Structure ===")

	mediator := NewSignalMediator()

	// Establish some electrical couplings
	mediator.EstablishElectricalCoupling("neuron1", "neuron2", 0.7)
	mediator.EstablishElectricalCoupling("neuron1", "neuron3", 0.5)

	// Send a signal
	testData := "test_signal_data"
	mediator.Send(SignalFired, "neuron1", testData)

	// Get the recorded event
	history := mediator.GetRecentSignals(1)
	if len(history) != 1 {
		t.Fatalf("FAIL: Should have recorded 1 event, got %d", len(history))
	}

	event := history[0]

	// Verify event structure
	if event.SignalType != SignalFired {
		t.Errorf("FAIL: Event should have SignalFired type, got %v", event.SignalType)
	}
	if event.SourceID != "neuron1" {
		t.Errorf("FAIL: Event should have source 'neuron1', got %s", event.SourceID)
	}
	if event.Data != testData {
		t.Errorf("FAIL: Event should have correct data, got %v", event.Data)
	}
	if len(event.TargetIDs) != 2 {
		t.Errorf("FAIL: Event should have 2 target IDs, got %d: %v", len(event.TargetIDs), event.TargetIDs)
	}
	if event.Timestamp.IsZero() {
		t.Error("FAIL: Event should have valid timestamp")
	}

	t.Log("✓ Signal event structure is correct")
}

// =================================================================================
// PERFORMANCE AND CONCURRENCY TESTS
// =================================================================================

func TestSignalMediatorConcurrentStressTest(t *testing.T) {
	t.Log("=== TESTING: Concurrent Stress Test ===")

	mediator := NewSignalMediator()
	numGoroutines := 50
	operationsPerGoroutine := 200
	var wg sync.WaitGroup
	var errorCount int32

	wg.Add(numGoroutines)

	// Concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(gID int) {
			defer wg.Done()
			listener := newMockListener(fmt.Sprintf("listener-%d", gID))

			for j := 0; j < operationsPerGoroutine; j++ {
				opType := (gID + j) % 5
				switch opType {
				case 0: // Add Listener
					mediator.AddListener([]SignalType{SignalFired}, listener)
				case 1: // Send Signal
					mediator.Send(SignalFired, fmt.Sprintf("source-%d", gID), gID*1000+j)
				case 2: // Establish Coupling
					err := mediator.EstablishElectricalCoupling(
						fmt.Sprintf("neuron-%d", gID),
						fmt.Sprintf("neuron-%d", j%numGoroutines),
						rand.Float64())
					if err != nil {
						atomic.AddInt32(&errorCount, 1)
					}
				case 3: // Remove Listener
					mediator.RemoveListener([]SignalType{SignalFired}, listener)
				case 4: // Query Operations
					mediator.GetElectricalCouplings(fmt.Sprintf("neuron-%d", gID))
					mediator.GetRecentSignals(5)
				}
			}
		}(i)
	}

	wg.Wait()

	if errorCount > 0 {
		t.Errorf("FAIL: Encountered %d errors during concurrent stress test", errorCount)
	} else {
		t.Log("✓ Concurrent stress test completed successfully")
	}

	// Verify final state consistency
	finalSignalCount := mediator.GetSignalCount()
	t.Logf("✓ Final signal count: %d", finalSignalCount)
}

func TestSignalMediatorMemoryEfficiency(t *testing.T) {
	t.Log("=== TESTING: Memory Efficiency ===")

	mediator := NewSignalMediator()
	listener := newMockListener("memory_test")
	mediator.AddListener([]SignalType{SignalFired}, listener)

	// Send many signals to test memory management
	numSignals := 2000
	for i := 0; i < numSignals; i++ {
		mediator.Send(SignalFired, "source", i)
	}

	// Verify history is limited
	history := mediator.GetRecentSignals(mediator.maxHistory + 100)
	if len(history) > mediator.maxHistory {
		t.Errorf("FAIL: History size exceeded maximum: got %d, max %d",
			len(history), mediator.maxHistory)
	}
	t.Logf("✓ History properly limited to %d events", len(history))

	// Test history clearing for memory management
	mediator.ClearSignalHistory()
	if mediator.GetSignalCount() != 0 {
		t.Error("FAIL: Signal count should be 0 after clearing")
	}
	t.Log("✓ Memory management through history clearing works")
}

// =================================================================================
// BIOLOGICAL REALISM TESTS
// =================================================================================

func TestSignalMediatorNetworkSynchronization(t *testing.T) {
	t.Log("=== TESTING: Biological Network Synchronization ===")

	mediator := NewSignalMediator()

	// Create a network of interneurons
	interneuronIDs := []string{"inh1", "inh2", "inh3", "inh4", "inh5"}
	listeners := make(map[string]*mockSignalListener)

	// Register all interneurons as listeners
	for _, id := range interneuronIDs {
		listeners[id] = newMockListener(id)
		mediator.AddListener([]SignalType{SignalFired}, listeners[id])
	}

	// Create electrical coupling between all interneurons (gap junction network)
	for i := 0; i < len(interneuronIDs); i++ {
		for j := i + 1; j < len(interneuronIDs); j++ {
			conductance := 0.8 + rand.Float64()*0.2 // High conductance for synchronization
			mediator.EstablishElectricalCoupling(interneuronIDs[i], interneuronIDs[j], conductance)
		}
	}
	t.Logf("✓ Created fully coupled network of %d interneurons", len(interneuronIDs))

	// Simulate initial firing from one interneuron
	t.Log("--- Simulating network synchronization cascade ---")
	initialSpike := "gamma_oscillation_start"
	mediator.Send(SignalFired, "inh1", initialSpike)

	// Allow signal propagation
	time.Sleep(20 * time.Millisecond)

	// Verify that all other interneurons received the synchronization signal
	synchronizedCount := 0
	for _, id := range interneuronIDs {
		if id != "inh1" && listeners[id].GetReceivedCount() > 0 {
			synchronizedCount++
			lastData := listeners[id].GetLastReceivedData()
			if lastData != initialSpike {
				t.Errorf("FAIL: Neuron %s received wrong sync data: %v", id, lastData)
			}
		}
	}

	expectedSynchronized := len(interneuronIDs) - 1 // All except the sender
	if synchronizedCount != expectedSynchronized {
		t.Errorf("FAIL: Expected %d synchronized neurons, got %d",
			expectedSynchronized, synchronizedCount)
	} else {
		t.Logf("✓ %d interneurons successfully synchronized", synchronizedCount)
	}

	// Test rapid synchronization cascade
	t.Log("--- Testing rapid synchronization responses ---")
	for _, id := range interneuronIDs {
		listeners[id].Reset()
	}

	// Multiple rapid spikes to test gamma-like oscillations
	for i := 0; i < 5; i++ {
		mediator.Send(SignalFired, "inh1", fmt.Sprintf("gamma_pulse_%d", i))
		time.Sleep(2 * time.Millisecond) // ~500Hz gamma frequency
	}

	time.Sleep(20 * time.Millisecond)

	// Verify all neurons received multiple synchronization pulses
	for _, id := range interneuronIDs {
		if id != "inh1" {
			receivedCount := listeners[id].GetReceivedCount()
			if receivedCount != 5 {
				t.Errorf("FAIL: Neuron %s should have received 5 gamma pulses, got %d",
					id, receivedCount)
			}
		}
	}
	t.Log("✓ Rapid gamma-like synchronization successfully demonstrated")
}

func TestSignalMediatorElectricalSynapseVsChemicalSynapse(t *testing.T) {
	t.Log("=== TESTING: Electrical vs Chemical Synapse Characteristics ===")

	mediator := NewSignalMediator()

	// Test electrical synapse characteristics: bidirectional, fast, no delay
	neuronA := newMockListener("electricalA")
	neuronB := newMockListener("electricalB")

	mediator.AddListener([]SignalType{SignalFired}, neuronA)
	mediator.AddListener([]SignalType{SignalFired}, neuronB)

	// Establish electrical coupling
	mediator.EstablishElectricalCoupling("electricalA", "electricalB", 1.0)

	// Test bidirectional communication
	t.Log("--- Testing bidirectional electrical coupling ---")

	// A signals to B
	startTime := time.Now()
	mediator.Send(SignalFired, "electricalA", "A_to_B")
	time.Sleep(5 * time.Millisecond)

	if neuronB.GetReceivedCount() != 1 {
		t.Error("FAIL: Electrical coupling A->B failed")
	}

	neuronA.Reset()
	neuronB.Reset()

	// B signals to A (reverse direction)
	mediator.Send(SignalFired, "electricalB", "B_to_A")
	time.Sleep(5 * time.Millisecond)

	if neuronA.GetReceivedCount() != 1 {
		t.Error("FAIL: Electrical coupling B->A failed")
	}

	electricalDelay := time.Since(startTime)
	t.Logf("✓ Bidirectional electrical coupling confirmed (delay: %v)", electricalDelay)

	// Test speed characteristics
	t.Log("--- Testing electrical synapse speed ---")
	signals := 100
	startTime = time.Now()

	for i := 0; i < signals; i++ {
		mediator.Send(SignalFired, "electricalA", i)
	}

	electricalTime := time.Since(startTime)
	t.Logf("✓ Processed %d electrical signals in %v (%.2f μs per signal)",
		signals, electricalTime, float64(electricalTime.Nanoseconds())/float64(signals)/1000.0)
}

func TestSignalMediatorGapJunctionConductanceEffects(t *testing.T) {
	t.Log("=== TESTING: Gap Junction Conductance Effects ===")

	mediator := NewSignalMediator()

	// Test different conductance levels
	conductanceLevels := []float64{0.1, 0.5, 0.9, 1.0}

	for _, conductance := range conductanceLevels {
		neuronPair := fmt.Sprintf("test_pair_%.1f", conductance)
		neuronA := neuronPair + "_A"
		neuronB := neuronPair + "_B"

		// Establish coupling with specific conductance
		mediator.EstablishElectricalCoupling(neuronA, neuronB, conductance)

		// Verify conductance was set correctly
		actualConductance := mediator.GetConductance(neuronA, neuronB)
		if actualConductance != conductance {
			t.Errorf("FAIL: Expected conductance %.2f, got %.2f", conductance, actualConductance)
		}

		// Verify bidirectional symmetry
		reverseConductance := mediator.GetConductance(neuronB, neuronA)
		if reverseConductance != conductance {
			t.Errorf("FAIL: Conductance not symmetric: A->B=%.2f, B->A=%.2f",
				actualConductance, reverseConductance)
		}
	}

	t.Log("✓ Gap junction conductance levels set and verified correctly")
}

// =================================================================================
// INTEGRATION TESTS
// =================================================================================

func TestSignalMediatorRealWorldScenario(t *testing.T) {
	t.Log("=== TESTING: Real-World Biological Scenario ===")

	mediator := NewSignalMediator()

	// Scenario: Motor cortex network with pyramidal neurons and interneurons
	pyramidalNeurons := []string{"pyr1", "pyr2", "pyr3"}
	interneurons := []string{"inh1", "inh2"}

	listeners := make(map[string]*mockSignalListener)

	// Register all neurons
	for _, id := range pyramidalNeurons {
		listeners[id] = newMockListener(id)
		mediator.AddListener([]SignalType{SignalFired, SignalConnected}, listeners[id])
	}
	for _, id := range interneurons {
		listeners[id] = newMockListener(id)
		mediator.AddListener([]SignalType{SignalFired}, listeners[id])
	}

	// Establish biological connectivity
	// 1. Interneurons are electrically coupled (gap junctions)
	mediator.EstablishElectricalCoupling("inh1", "inh2", 0.8)

	// 2. Pyramidal neurons have some electrical coupling (weaker)
	mediator.EstablishElectricalCoupling("pyr1", "pyr2", 0.3)

	t.Log("✓ Motor cortex network topology established")

	// Simulate motor command sequence
	t.Log("--- Simulating motor command execution ---")

	// 1. Primary motor neuron fires
	mediator.Send(SignalFired, "pyr1", "motor_command_start")
	time.Sleep(5 * time.Millisecond)

	// 2. This should activate connected pyramidal neuron
	if listeners["pyr2"].GetReceivedCount() == 0 {
		t.Error("FAIL: Connected pyramidal neuron should respond")
	}

	// 3. Interneuron network provides inhibitory control
	mediator.Send(SignalFired, "inh1", "inhibitory_control")
	time.Sleep(5 * time.Millisecond)

	// 4. Both interneurons should be synchronized
	if listeners["inh2"].GetReceivedCount() == 0 {
		t.Error("FAIL: Coupled interneuron should be synchronized")
	}

	// 5. Send connection events
	mediator.Send(SignalConnected, "pyr1", "new_synapse_formed")
	time.Sleep(5 * time.Millisecond)

	// Verify pyramidal neurons received connection signals
	for _, id := range pyramidalNeurons {
		if id != "pyr1" && listeners[id].GetReceivedCount() == 0 {
			t.Errorf("FAIL: Pyramidal neuron %s should receive connection signals", id)
		}
	}

	t.Log("✓ Motor cortex simulation completed successfully")
}

func TestSignalMediatorBiologicalConstraintsValidation(t *testing.T) {
	t.Log("=== TESTING: Biological Constraints Validation ===")

	mediator := NewSignalMediator()

	// Test 1: Realistic gap junction conductance ranges
	t.Log("--- Testing realistic conductance ranges ---")

	// Typical biological gap junction conductances: 0.1-1.0 nS normalized
	validConductances := []float64{0.1, 0.2, 0.5, 0.8, 1.0}
	for _, conductance := range validConductances {
		err := mediator.EstablishElectricalCoupling("neuron1", "neuron2", conductance)
		if err != nil {
			t.Errorf("FAIL: Valid conductance %.2f rejected", conductance)
		}
		mediator.RemoveElectricalCoupling("neuron1", "neuron2")
	}
	t.Log("✓ Valid conductance ranges accepted")

	// Test 2: Signal frequency limits (biological neurons max ~1000Hz)
	t.Log("--- Testing signal frequency limits ---")

	listener := newMockListener("frequency_test")
	mediator.AddListener([]SignalType{SignalFired}, listener)

	// Send signals at high frequency
	highFreqCount := 100
	startTime := time.Now()
	for i := 0; i < highFreqCount; i++ {
		mediator.Send(SignalFired, "high_freq_source", i)
	}
	duration := time.Since(startTime)

	frequency := float64(highFreqCount) / duration.Seconds()
	t.Logf("✓ Processed signals at %.0f Hz", frequency)

	// Test 3: Network size scalability
	t.Log("--- Testing network scalability ---")

	networkSize := 50
	for i := 0; i < networkSize; i++ {
		neuronID := fmt.Sprintf("neuron_%d", i)
		listener := newMockListener(neuronID)
		mediator.AddListener([]SignalType{SignalFired}, listener)

		// Create some electrical couplings
		if i > 0 {
			prevNeuron := fmt.Sprintf("neuron_%d", i-1)
			mediator.EstablishElectricalCoupling(neuronID, prevNeuron, 0.5)
		}
	}

	// Send signal through network
	mediator.Send(SignalFired, "network_stimulus", "propagation_test")
	time.Sleep(50 * time.Millisecond)

	t.Logf("✓ Network with %d components created and tested", networkSize)
}

// =================================================================================
// ERROR HANDLING AND EDGE CASE TESTS
// =================================================================================

func TestSignalMediatorErrorHandling(t *testing.T) {
	t.Log("=== TESTING: Error Handling and Edge Cases ===")

	mediator := NewSignalMediator()

	// Test 1: Operations on empty mediator
	t.Log("--- Testing operations on empty mediator ---")

	// Should not panic
	mediator.Send(SignalFired, "ghost_source", nil)
	couplings := mediator.GetElectricalCouplings("non_existent")
	if len(couplings) != 0 {
		t.Error("FAIL: Non-existent component should have empty couplings")
	}

	conductance := mediator.GetConductance("none1", "none2")
	if conductance != 0.0 {
		t.Error("FAIL: Non-existent coupling should have 0 conductance")
	}
	t.Log("✓ Empty mediator operations are safe")

	// Test 2: Duplicate listener registration
	t.Log("--- Testing duplicate listener registration ---")

	listener := newMockListener("duplicate_test")
	mediator.AddListener([]SignalType{SignalFired}, listener)
	mediator.AddListener([]SignalType{SignalFired}, listener) // Duplicate

	mediator.Send(SignalFired, "test_source", "duplicate_signal")
	time.Sleep(10 * time.Millisecond)

	// Should only receive once, not twice
	if listener.GetReceivedCount() != 1 {
		t.Errorf("FAIL: Duplicate registration should be prevented, got %d signals",
			listener.GetReceivedCount())
	}
	t.Log("✓ Duplicate listener registration handled correctly")

	// Test 3: Removing non-existent listener
	t.Log("--- Testing removal of non-existent listener ---")

	nonExistentListener := newMockListener("non_existent")
	mediator.RemoveListener([]SignalType{SignalFired}, nonExistentListener) // Should not panic
	t.Log("✓ Non-existent listener removal is safe")

	// Test 4: Multiple coupling establishment/removal
	t.Log("--- Testing multiple coupling operations ---")

	mediator.EstablishElectricalCoupling("multi1", "multi2", 0.5)
	mediator.EstablishElectricalCoupling("multi1", "multi2", 0.8) // Override

	finalConductance := mediator.GetConductance("multi1", "multi2")
	if finalConductance != 0.8 {
		t.Errorf("FAIL: Coupling override failed, got %.2f", finalConductance)
	}

	// Multiple removals should be safe
	mediator.RemoveElectricalCoupling("multi1", "multi2")
	mediator.RemoveElectricalCoupling("multi1", "multi2") // Should not panic
	t.Log("✓ Multiple coupling operations handled correctly")
}

func TestSignalMediatorResourceCleanup(t *testing.T) {
	t.Log("=== TESTING: Resource Cleanup ===")

	mediator := NewSignalMediator()

	// Create many temporary connections
	numConnections := 100
	for i := 0; i < numConnections; i++ {
		neuronA := fmt.Sprintf("temp_neuron_A_%d", i)
		neuronB := fmt.Sprintf("temp_neuron_B_%d", i)
		mediator.EstablishElectricalCoupling(neuronA, neuronB, 0.5)
	}

	// Verify connections exist
	totalCouplings := 0
	for i := 0; i < numConnections; i++ {
		neuronA := fmt.Sprintf("temp_neuron_A_%d", i)
		couplings := mediator.GetElectricalCouplings(neuronA)
		totalCouplings += len(couplings)
	}

	if totalCouplings != numConnections {
		t.Errorf("FAIL: Expected %d couplings, got %d", numConnections, totalCouplings)
	}

	// Remove all connections
	for i := 0; i < numConnections; i++ {
		neuronA := fmt.Sprintf("temp_neuron_A_%d", i)
		neuronB := fmt.Sprintf("temp_neuron_B_%d", i)
		mediator.RemoveElectricalCoupling(neuronA, neuronB)
	}

	// Verify cleanup
	totalAfterCleanup := 0
	for i := 0; i < numConnections; i++ {
		neuronA := fmt.Sprintf("temp_neuron_A_%d", i)
		couplings := mediator.GetElectricalCouplings(neuronA)
		totalAfterCleanup += len(couplings)
	}

	if totalAfterCleanup != 0 {
		t.Errorf("FAIL: Expected 0 couplings after cleanup, got %d", totalAfterCleanup)
	}

	t.Log("✓ Resource cleanup successful")
}

// =================================================================================
// BENCHMARKS AND PERFORMANCE TESTS
// =================================================================================

func BenchmarkSignalSending(b *testing.B) {
	mediator := NewSignalMediator()
	listener := newMockListener("bench_listener")
	mediator.AddListener([]SignalType{SignalFired}, listener)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mediator.Send(SignalFired, "bench_source", i)
	}
}

func BenchmarkElectricalCouplingOperations(b *testing.B) {
	mediator := NewSignalMediator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		neuronA := fmt.Sprintf("neuron_%d", i%100)
		neuronB := fmt.Sprintf("neuron_%d", (i+1)%100)

		if i%2 == 0 {
			mediator.EstablishElectricalCoupling(neuronA, neuronB, 0.5)
		} else {
			mediator.RemoveElectricalCoupling(neuronA, neuronB)
		}
	}
}

func BenchmarkConcurrentSignaling(b *testing.B) {
	mediator := NewSignalMediator()
	numListeners := 10

	// Set up listeners
	for i := 0; i < numListeners; i++ {
		listener := newMockListener(fmt.Sprintf("concurrent_listener_%d", i))
		mediator.AddListener([]SignalType{SignalFired}, listener)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		sourceID := fmt.Sprintf("concurrent_source_%d", rand.Int())
		for pb.Next() {
			mediator.Send(SignalFired, sourceID, rand.Int())
		}
	})
}
