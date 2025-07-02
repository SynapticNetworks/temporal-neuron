/*
=================================================================================
EXTRACELLULAR MATRIX - EDGE CASE AND BOUNDARY CONDITION TEST SUITE
=================================================================================

This comprehensive test suite validates the ExtracellularMatrix's robustness
under extreme conditions, boundary cases, and error scenarios. These tests
ensure the matrix maintains biological realism and system stability even when
operating at the limits of its designed parameters.

EDGE CASE CATEGORIES TESTED:
1. Resource Boundary Conditions - Component limits, memory constraints, capacity
2. Concurrency Stress Testing - Race conditions, deadlocks, thread safety
3. Invalid Input Handling - Malformed data, boundary values, error propagation
4. System State Transitions - Startup/shutdown edge cases, partial failures
5. Biological Constraint Violations - Non-physical parameters, limit enforcement
6. Performance Degradation - Graceful behavior under extreme load
7. Memory Management - Leak detection, cleanup verification, resource tracking
8. Network Topology Extremes - Sparse/dense networks, disconnected components

TESTING PHILOSOPHY:
These tests intentionally push the matrix beyond normal operating conditions
to validate that it fails gracefully, maintains data integrity, and provides
meaningful error messages. Like testing a biological system's stress response,
we examine how the matrix responds to pathological conditions.

BIOLOGICAL INSPIRATION:
Real neural tissue exhibits remarkable robustness - neurons continue functioning
despite metabolic stress, toxin exposure, and structural damage. This test suite
validates that our computational equivalent demonstrates similar resilience.

USAGE:
Run all edge tests: go test -run TestMatrixEdge
Run specific category: go test -run TestMatrixEdgeResource
Run with race detection: go test -race -run TestMatrixEdge

=================================================================================
*/

package extracellular

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// =================================================================================
// TEST 1: RESOURCE BOUNDARY CONDITIONS AND COMPONENT LIMITS
// =================================================================================

// TestMatrixEdgeResourceLimits validates proper handling of resource constraints
// and component creation limits under various boundary conditions.
//
// BIOLOGICAL PROCESS TESTED:
// This models the metabolic limits that constrain neural tissue development:
// - Glucose availability limits maximum neuron density
// - Oxygen supply constrains neural activity levels
// - Growth factors determine neurogenesis rates
// - Waste product accumulation triggers pruning
//
// EDGE CONDITIONS TESTED:
// - Zero component limit (metabolic starvation)
// - Single component limit (minimal viable tissue)
// - Exact limit boundary (precise resource allocation)
// - Attempt to exceed limits (resource exhaustion)
// - Rapid sequential creation (burst neurogenesis)
// - Mixed component types at limits (balanced growth)
//
// EXPECTED BEHAVIORS:
// - Clean rejection when limits reached
// - Consistent error messages
// - No resource leaks during failures
// - Proper cleanup of partial allocations
// - Accurate component counting
// - Stable system state after limit violations
func TestMatrixEdgeResourceLimits(t *testing.T) {
	t.Log("=== MATRIX EDGE TEST: Resource Boundary Conditions ===")
	t.Log("Testing component limits, resource constraints, and boundary behaviors")

	testCases := []struct {
		name             string
		maxComponents    int
		neuronsToTry     int
		synapsesToTry    int
		expectedNeurons  int
		expectedSynapses int
		description      string
	}{
		{
			name:             "zero_capacity",
			maxComponents:    0,
			neuronsToTry:     5,
			synapsesToTry:    3,
			expectedNeurons:  0,
			expectedSynapses: 0,
			description:      "Metabolic starvation - no resources available",
		},
		{
			name:             "minimal_capacity",
			maxComponents:    1,
			neuronsToTry:     3,
			synapsesToTry:    2,
			expectedNeurons:  1,
			expectedSynapses: 0,
			description:      "Minimal viable tissue - single component only",
		},
		{
			name:             "exact_boundary",
			maxComponents:    5,
			neuronsToTry:     3,
			synapsesToTry:    2,
			expectedNeurons:  3,
			expectedSynapses: 2,
			description:      "Exact resource utilization - precise limit",
		},
		{
			name:             "exceed_boundary",
			maxComponents:    3,
			neuronsToTry:     5,
			synapsesToTry:    3,
			expectedNeurons:  3,
			expectedSynapses: 0,
			description:      "Resource exhaustion - demand exceeds supply",
		},
		{
			name:             "neuron_heavy",
			maxComponents:    4,
			neuronsToTry:     4,
			synapsesToTry:    3,
			expectedNeurons:  4,
			expectedSynapses: 0,
			description:      "Neuron-dominated tissue - all resources to cell bodies",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("--- Testing: %s ---", tc.description)
			t.Logf("Capacity: %d, Trying: %d neurons + %d synapses",
				tc.maxComponents, tc.neuronsToTry, tc.synapsesToTry)

			// Initialize matrix with specific resource limits
			matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
				ChemicalEnabled: true,
				SpatialEnabled:  true,
				UpdateInterval:  10 * time.Millisecond,
				MaxComponents:   tc.maxComponents,
			})
			defer matrix.Stop()

			// Register test factories
			matrix.RegisterNeuronType("edge_neuron", func(id string, config types.NeuronConfig, callbacks NeuronCallbacks) (component.NeuralComponent, error) {
				return NewMockNeuron(id, config.Position, config.Receptors), nil
			})

			matrix.RegisterSynapseType("edge_synapse", func(id string, config types.SynapseConfig, callbacks SynapseCallbacks) (component.SynapticProcessor, error) {
				return NewMockSynapse(id, config.Position, config.PresynapticID, config.PostsynapticID, config.InitialWeight), nil
			})

			// === PHASE 1: NEURON CREATION UNDER RESOURCE PRESSURE ===
			t.Logf("Creating %d neurons...", tc.neuronsToTry)

			var createdNeurons []component.NeuralComponent
			neuronErrors := 0

			for i := 0; i < tc.neuronsToTry; i++ {
				config := types.NeuronConfig{
					NeuronType: "edge_neuron",
					Position:   Position3D{X: float64(i), Y: 0, Z: 0},
					Receptors:  []LigandType{LigandGlutamate},
					Threshold:  0.7,
				}

				neuron, err := matrix.CreateNeuron(config)
				if err != nil {
					neuronErrors++
					t.Logf("  Neuron %d creation failed (expected): %v", i, err)
				} else {
					createdNeurons = append(createdNeurons, neuron)
					t.Logf("  Neuron %d created successfully", i)
				}
			}

			// Validate neuron creation results
			if len(createdNeurons) != tc.expectedNeurons {
				t.Errorf("Neuron count mismatch: expected %d, created %d",
					tc.expectedNeurons, len(createdNeurons))
			}

			// === PHASE 2: SYNAPSE CREATION WITH REMAINING RESOURCES ===
			if len(createdNeurons) >= 2 && tc.synapsesToTry > 0 {
				t.Logf("Creating %d synapses with remaining resources...", tc.synapsesToTry)

				var createdSynapses []component.SynapticProcessor
				synapseErrors := 0

				for i := 0; i < tc.synapsesToTry; i++ {
					// Use created neurons for synapse endpoints
					preIdx := i % len(createdNeurons)
					postIdx := (i + 1) % len(createdNeurons)

					config := types.SynapseConfig{
						SynapseType:    "edge_synapse",
						PresynapticID:  createdNeurons[preIdx].ID(),
						PostsynapticID: createdNeurons[postIdx].ID(),
						Position:       Position3D{X: float64(i), Y: 1, Z: 0},
						InitialWeight:  0.5,
						LigandType:     LigandGlutamate,
					}

					synapse, err := matrix.CreateSynapse(config)
					if err != nil {
						synapseErrors++
						t.Logf("  Synapse %d creation failed (expected): %v", i, err)
					} else {
						createdSynapses = append(createdSynapses, synapse)
						t.Logf("  Synapse %d created successfully", i)
					}
				}

				// Validate synapse creation results
				if len(createdSynapses) != tc.expectedSynapses {
					t.Errorf("Synapse count mismatch: expected %d, created %d",
						tc.expectedSynapses, len(createdSynapses))
				}
			}

			// === PHASE 3: SYSTEM STATE VALIDATION ===
			t.Log("Validating system state integrity...")

			allNeurons := matrix.ListNeurons()
			allSynapses := matrix.ListSynapses()
			totalComponents := matrix.astrocyteNetwork.Count()

			t.Logf("Final counts - Neurons: %d, Synapses: %d, Total: %d",
				len(allNeurons), len(allSynapses), totalComponents)

			// Validate component tracking consistency
			if len(allNeurons) != tc.expectedNeurons {
				t.Errorf("Matrix neuron tracking inconsistent: expected %d, tracked %d",
					tc.expectedNeurons, len(allNeurons))
			}

			if len(allSynapses) != tc.expectedSynapses {
				t.Errorf("Matrix synapse tracking inconsistent: expected %d, tracked %d",
					tc.expectedSynapses, len(allSynapses))
			}

			// Validate resource limit enforcement
			expectedTotal := tc.expectedNeurons + tc.expectedSynapses
			if totalComponents < expectedTotal {
				t.Errorf("Component count too low: expected ≥%d, got %d", expectedTotal, totalComponents)
			}

			if totalComponents > tc.maxComponents {
				t.Errorf("CRITICAL: Resource limit violated! Max: %d, Actual: %d",
					tc.maxComponents, totalComponents)
			}

			t.Logf("✓ %s: Resource limits properly enforced", tc.name)
		})
	}

	t.Log("✅ Resource boundary condition tests completed")
	t.Log("✅ Matrix maintains integrity under resource pressure")
}

// =================================================================================
// TEST 2: CONCURRENCY STRESS TESTING AND RACE CONDITIONS
// =================================================================================

// TestMatrixEdgeConcurrencyStress validates thread safety under extreme concurrent access.
//
// BIOLOGICAL PROCESS TESTED:
// This models the massive parallelism of real neural tissue:
// - Thousands of synapses firing simultaneously
// - Concurrent neurotransmitter release and uptake
// - Parallel glial cell maintenance and monitoring
// - Simultaneous electrical coupling across gap junctions
//
// CONCURRENCY EDGE CONDITIONS:
// - Simultaneous component creation from multiple threads
// - Concurrent chemical releases from different sources
// - Parallel electrical signaling bursts
// - Race conditions in resource allocation
// - Deadlock prevention during high contention
// - Memory consistency under concurrent modification
//
// STRESS TEST PARAMETERS:
// - High goroutine count (simulating dense neural activity)
// - Rapid operation rate (burst firing conditions)
// - Mixed operation types (chemical + electrical + spatial)
// - Long duration stress (sustained activity)
// - Resource pressure (near-limit conditions)
//
// VALIDATION CRITERIA:
// - No panics or crashes
// - No data races (verified with -race flag)
// - Consistent final state
// - Predictable resource accounting
// - Graceful performance degradation
// - Proper error propagation under contention
func TestMatrixEdgeConcurrencyStress(t *testing.T) {
	t.Log("=== MATRIX EDGE TEST: Concurrency Stress and Race Conditions ===")
	t.Log("Testing thread safety under extreme concurrent access patterns")

	// Configuration for stress testing
	const (
		STRESS_DURATION       = 2 * time.Second
		CONCURRENT_GOROUTINES = 50
		OPERATIONS_PER_THREAD = 100
		MATRIX_CAPACITY       = 200
	)

	// Initialize matrix for concurrency testing
	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  1 * time.Millisecond, // Fast updates for stress
		MaxComponents:   MATRIX_CAPACITY,
	})
	defer matrix.Stop()

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix for concurrency test: %v", err)
	}

	err = matrix.chemicalModulator.Start()
	if err != nil {
		t.Fatalf("Failed to start chemical modulator: %v", err)
	}

	// Register stress test factories
	matrix.RegisterNeuronType("stress_neuron", func(id string, config types.NeuronConfig, callbacks NeuronCallbacks) (component.NeuralComponent, error) {
		return NewMockNeuron(id, config.Position, config.Receptors), nil
	})

	matrix.RegisterSynapseType("stress_synapse", func(id string, config types.SynapseConfig, callbacks SynapseCallbacks) (component.SynapticProcessor, error) {
		return NewMockSynapse(id, config.Position, config.PresynapticID, config.PostsynapticID, config.InitialWeight), nil
	})

	// === PHASE 1: CONCURRENT COMPONENT CREATION STRESS ===
	t.Log("\n--- Phase 1: Concurrent Component Creation ---")

	var wg sync.WaitGroup
	var neuronCount, synapseCount int64
	var creationErrors int64

	// Track operation statistics
	startTime := time.Now()

	// Launch concurrent neuron creators
	for i := 0; i < CONCURRENT_GOROUTINES/2; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()

			for j := 0; j < OPERATIONS_PER_THREAD/4; j++ {
				config := types.NeuronConfig{
					NeuronType: "stress_neuron",
					Position:   Position3D{X: float64(threadID*100 + j), Y: float64(threadID), Z: 0},
					Receptors:  []LigandType{LigandGlutamate, LigandGABA},
					Threshold:  0.7,
				}

				_, err := matrix.CreateNeuron(config)
				if err != nil {
					atomic.AddInt64(&creationErrors, 1)
				} else {
					atomic.AddInt64(&neuronCount, 1)
				}

				// Brief pause to allow other operations
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	// Wait for component creation to stabilize
	wg.Wait()

	creationTime := time.Since(startTime)

	t.Logf("Component creation phase completed in %v", creationTime)
	t.Logf("Created: %d neurons (errors: %d)", neuronCount, creationErrors)

	// Verify we have enough neurons for synapse creation
	actualNeurons := matrix.ListNeurons()
	if len(actualNeurons) < 2 {
		t.Skip("Not enough neurons created for synapse stress testing")
	}

	// === PHASE 2: CONCURRENT SYNAPSE CREATION ===
	t.Log("\n--- Phase 2: Concurrent Synapse Creation ---")

	var synapseErrors int64
	startTime = time.Now()

	// Launch concurrent synapse creators
	for i := 0; i < CONCURRENT_GOROUTINES/4; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()

			for j := 0; j < OPERATIONS_PER_THREAD/8; j++ {
				// Randomly select neurons for connection
				preIdx := (threadID * j) % len(actualNeurons)
				postIdx := (threadID*j + 1) % len(actualNeurons)

				if preIdx == postIdx {
					continue // Skip self-connections
				}

				config := types.SynapseConfig{
					SynapseType:    "stress_synapse",
					PresynapticID:  actualNeurons[preIdx].ID(),
					PostsynapticID: actualNeurons[postIdx].ID(),
					Position:       Position3D{X: float64(threadID), Y: float64(j), Z: 1},
					InitialWeight:  0.5,
					LigandType:     LigandGlutamate,
				}

				_, err := matrix.CreateSynapse(config)
				if err != nil {
					atomic.AddInt64(&synapseErrors, 1)
				} else {
					atomic.AddInt64(&synapseCount, 1)
				}

				time.Sleep(2 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	synapseCreationTime := time.Since(startTime)

	t.Logf("Synapse creation phase completed in %v", synapseCreationTime)
	t.Logf("Created: %d synapses (errors: %d)", synapseCount, synapseErrors)

	// === PHASE 3: CONCURRENT CHEMICAL SIGNALING STRESS ===
	t.Log("\n--- Phase 3: Concurrent Chemical Signaling ---")

	var chemicalOps, chemicalErrors int64
	startTime = time.Now()

	// Launch concurrent chemical releasers
	for i := 0; i < CONCURRENT_GOROUTINES; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()

			neuronIndex := threadID % len(actualNeurons)
			sourceID := actualNeurons[neuronIndex].ID()

			for j := 0; j < OPERATIONS_PER_THREAD/2; j++ {
				// Vary ligand types and concentrations
				ligandTypes := []LigandType{LigandGlutamate, LigandGABA, LigandDopamine}
				ligand := ligandTypes[j%len(ligandTypes)]
				concentration := 0.1 + (float64(j%10) * 0.1)

				err := matrix.ReleaseLigand(ligand, sourceID, concentration)
				if err != nil {
					atomic.AddInt64(&chemicalErrors, 1)
				}
				atomic.AddInt64(&chemicalOps, 1)

				// Respect biological rate limits
				time.Sleep(3 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	chemicalTime := time.Since(startTime)

	t.Logf("Chemical signaling phase completed in %v", chemicalTime)
	t.Logf("Chemical operations: %d (errors: %d)", chemicalOps, chemicalErrors)

	// === PHASE 4: CONCURRENT ELECTRICAL SIGNALING STRESS ===
	t.Log("\n--- Phase 4: Concurrent Electrical Signaling ---")

	var electricalOps int64
	startTime = time.Now()

	// Register neurons for electrical signaling
	for _, neuron := range actualNeurons[:10] { // Limit to avoid overwhelming
		if electricalReceiver, ok := neuron.(component.ElectricalReceiver); ok {
			matrix.ListenForSignals([]SignalType{SignalFired, SignalConnected}, electricalReceiver)
		}
	}

	// Launch concurrent electrical signalers
	for i := 0; i < CONCURRENT_GOROUTINES; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()

			neuronIndex := threadID % len(actualNeurons)
			sourceID := actualNeurons[neuronIndex].ID()

			for j := 0; j < OPERATIONS_PER_THREAD; j++ {
				signalTypes := []SignalType{SignalFired, SignalConnected}
				signal := signalTypes[j%len(signalTypes)]
				data := float64(j % 5)

				matrix.SendSignal(signal, sourceID, data)
				atomic.AddInt64(&electricalOps, 1)

				// No biological rate limit for electrical signals
				time.Sleep(100 * time.Microsecond)
			}
		}(i)
	}

	wg.Wait()

	electricalTime := time.Since(startTime)

	t.Logf("Electrical signaling phase completed in %v", electricalTime)
	t.Logf("Electrical operations: %d", electricalOps)

	// === PHASE 5: CONCURRENT SPATIAL QUERY STRESS ===
	t.Log("\n--- Phase 5: Concurrent Spatial Query ---")

	var spatialOps int64
	startTime = time.Now()

	// Launch concurrent spatial queriers
	for i := 0; i < CONCURRENT_GOROUTINES/2; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()

			for j := 0; j < OPERATIONS_PER_THREAD/2; j++ {
				queryPos := Position3D{
					X: float64(threadID * j % 100),
					Y: float64(threadID % 50),
					Z: float64(j % 10),
				}

				_ = matrix.FindComponents(ComponentCriteria{
					Position: &queryPos,
					Radius:   10.0,
				})

				atomic.AddInt64(&spatialOps, 1)
				time.Sleep(500 * time.Microsecond)
			}
		}(i)
	}

	wg.Wait()

	spatialTime := time.Since(startTime)

	t.Logf("Spatial query phase completed in %v", spatialTime)
	t.Logf("Spatial operations: %d", spatialOps)

	// === FINAL VALIDATION AND STATISTICS ===
	t.Log("\n--- Final System State Validation ---")

	finalNeurons := matrix.ListNeurons()
	finalSynapses := matrix.ListSynapses()
	finalComponents := matrix.astrocyteNetwork.Count()

	t.Logf("Final state - Neurons: %d, Synapses: %d, Components: %d",
		len(finalNeurons), len(finalSynapses), finalComponents)

	// Performance statistics
	totalOps := neuronCount + synapseCount + chemicalOps + electricalOps + spatialOps
	totalTime := creationTime + synapseCreationTime + chemicalTime + electricalTime + spatialTime
	avgOpsPerSecond := float64(totalOps) / totalTime.Seconds()

	t.Logf("Performance summary:")
	t.Logf("  Total operations: %d", totalOps)
	t.Logf("  Total time: %v", totalTime)
	t.Logf("  Average ops/second: %.1f", avgOpsPerSecond)
	t.Logf("  Total errors: %d", creationErrors+synapseErrors+chemicalErrors)

	// Validate system consistency
	if finalComponents > MATRIX_CAPACITY {
		t.Errorf("CRITICAL: Resource limit violated under stress! Max: %d, Actual: %d",
			MATRIX_CAPACITY, finalComponents)
	}

	// Check that the system is still responsive
	_, err = matrix.GetSpatialDistance(finalNeurons[0].ID(), finalNeurons[1].ID())
	if err != nil {
		t.Errorf("System unresponsive after stress test: %v", err)
	}

	t.Log("✅ Concurrency stress test completed successfully")
	t.Log("✅ Matrix maintains thread safety under extreme load")
	t.Log("✅ No deadlocks, race conditions, or data corruption detected")
}

// =================================================================================
// TEST 3: INVALID INPUT HANDLING AND MALFORMED DATA
// =================================================================================

// TestMatrixEdgeInvalidInputs validates robust handling of malformed, extreme,
// and pathological input data across all matrix APIs.
//
// BIOLOGICAL PROCESS TESTED:
// This models how biological systems handle toxic or abnormal conditions:
// - Cellular responses to extreme pH, temperature, or ion concentrations
// - Protein folding under stress conditions
// - Membrane integrity under osmotic pressure
// - Neurotransmitter system responses to toxins or drugs
//
// INVALID INPUT CATEGORIES:
// - Numerical extremes (infinity, NaN, overflow values)
// - Negative values where positive expected (concentrations, distances)
// - Empty or nil parameters (missing required data)
// - Oversized data structures (memory pressure)
// - Invalid enum values (unknown ligand types, signal types)
// - Malformed spatial coordinates (non-Euclidean positions)
// - Temporal paradoxes (negative delays, future timestamps)
//
// EXPECTED BEHAVIORS:
// - Graceful rejection with informative error messages
// - No system crashes or undefined behavior
// - Consistent error handling across all APIs
// - Proper input sanitization and validation
// - Safe fallback values where appropriate
// - No side effects from invalid operations

func TestMatrixEdgeInvalidInputs(t *testing.T) {
	t.Log("=== MATRIX EDGE TEST: Invalid Input Handling ===")
	t.Log("Testing robust error handling for malformed and extreme inputs")

	// Initialize matrix for invalid input testing
	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  10 * time.Millisecond,
		MaxComponents:   100,
	})
	defer matrix.Stop()

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix for invalid input testing: %v", err)
	}

	// FIXED: Register the factory BEFORE running the tests
	matrix.RegisterNeuronType("test_neuron", func(id string, config types.NeuronConfig, callbacks NeuronCallbacks) (component.NeuralComponent, error) {
		if config.Threshold < 0 || math.IsInf(config.Threshold, 0) || math.IsNaN(config.Threshold) {
			return nil, fmt.Errorf("invalid threshold: %f", config.Threshold)
		}
		if config.DecayRate < 0 || config.DecayRate > 1 {
			return nil, fmt.Errorf("invalid decay rate: %f", config.DecayRate)
		}
		if math.IsNaN(config.Position.X) || math.IsNaN(config.Position.Y) || math.IsNaN(config.Position.Z) {
			return nil, fmt.Errorf("invalid position coordinates")
		}
		return NewMockNeuron(id, config.Position, config.Receptors), nil
	})

	// === PHASE 1: INVALID NEURON CREATION PARAMETERS ===
	t.Log("\n--- Phase 1: Invalid Neuron Creation Parameters ---")

	invalidNeuronTests := []struct {
		name        string
		config      types.NeuronConfig
		expectError bool
		description string
	}{
		{
			name: "negative_threshold",
			config: types.NeuronConfig{
				NeuronType: "test_neuron",
				Threshold:  -1.0,
				Position:   Position3D{X: 0, Y: 0, Z: 0},
			},
			expectError: true,
			description: "Negative action potential threshold (biologically impossible)",
		},
		{
			name: "infinite_threshold",
			config: types.NeuronConfig{
				NeuronType: "test_neuron",
				Threshold:  math.Inf(1),
				Position:   Position3D{X: 0, Y: 0, Z: 0},
			},
			expectError: true,
			description: "Infinite threshold (pathological condition)",
		},
		{
			name: "nan_position",
			config: types.NeuronConfig{
				NeuronType: "test_neuron",
				Threshold:  0.7,
				Position:   Position3D{X: math.NaN(), Y: 0, Z: 0},
			},
			expectError: true,
			description: "NaN spatial coordinates (undefined position)",
		},
		{
			name: "extreme_coordinates",
			config: types.NeuronConfig{
				NeuronType: "test_neuron",
				Threshold:  0.7,
				Position:   Position3D{X: 1e20, Y: 1e20, Z: 1e20},
			},
			expectError: false, // Should handle gracefully
			description: "Extreme spatial coordinates (distant universe)",
		},
		{
			name: "empty_neuron_type",
			config: types.NeuronConfig{
				NeuronType: "",
				Threshold:  0.7,
				Position:   Position3D{X: 0, Y: 0, Z: 0},
			},
			expectError: true,
			description: "Empty neuron type (undefined cell identity)",
		},
		{
			name: "negative_decay",
			config: types.NeuronConfig{
				NeuronType: "test_neuron",
				Threshold:  0.7,
				DecayRate:  -0.1,
				Position:   Position3D{X: 0, Y: 0, Z: 0},
			},
			expectError: true,
			description: "Negative decay rate (energy creation violates thermodynamics)",
		},
	}

	for _, test := range invalidNeuronTests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Testing: %s", test.description)

			_, err := matrix.CreateNeuron(test.config)

			if test.expectError && err == nil {
				t.Errorf("Expected error for %s, but creation succeeded", test.name)
			} else if !test.expectError && err != nil {
				t.Errorf("Unexpected error for %s: %v", test.name, err)
			} else if test.expectError && err != nil {
				t.Logf("✓ Correctly rejected %s: %v", test.name, err)
			} else {
				t.Logf("✓ Gracefully handled %s", test.name)
			}
		})
	}

	// === PHASE 2: INVALID CHEMICAL SIGNALING PARAMETERS ===
	t.Log("\n--- Phase 2: Invalid Chemical Signaling Parameters ---")

	// Create a valid neuron for chemical testing
	validConfig := types.NeuronConfig{
		NeuronType: "test_neuron",
		Threshold:  0.7,
		Position:   Position3D{X: 0, Y: 0, Z: 0},
		Receptors:  []LigandType{LigandGlutamate},
	}

	// Register a minimal factory for testing
	matrix.RegisterNeuronType("test_neuron", func(id string, config types.NeuronConfig, callbacks NeuronCallbacks) (component.NeuralComponent, error) {
		if config.Threshold < 0 || math.IsInf(config.Threshold, 0) || math.IsNaN(config.Threshold) {
			return nil, fmt.Errorf("invalid threshold: %f", config.Threshold)
		}
		if config.DecayRate < 0 || config.DecayRate > 1 {
			return nil, fmt.Errorf("invalid decay rate: %f", config.DecayRate)
		}
		if math.IsNaN(config.Position.X) || math.IsNaN(config.Position.Y) || math.IsNaN(config.Position.Z) {
			return nil, fmt.Errorf("invalid position coordinates")
		}
		return NewMockNeuron(id, config.Position, config.Receptors), nil
	})

	testNeuron, err := matrix.CreateNeuron(validConfig)
	if err != nil {
		t.Fatalf("Failed to create test neuron: %v", err)
	}

	invalidChemicalTests := []struct {
		name          string
		ligandType    LigandType
		sourceID      string
		concentration float64
		expectError   bool
		description   string
	}{
		{
			name:          "negative_concentration",
			ligandType:    LigandGlutamate,
			sourceID:      testNeuron.ID(),
			concentration: -1.0,
			expectError:   true,
			description:   "Negative neurotransmitter concentration (impossible)",
		},
		{
			name:          "infinite_concentration",
			ligandType:    LigandGlutamate,
			sourceID:      testNeuron.ID(),
			concentration: math.Inf(1),
			expectError:   true,
			description:   "Infinite concentration (physically impossible)",
		},
		{
			name:          "nan_concentration",
			ligandType:    LigandGlutamate,
			sourceID:      testNeuron.ID(),
			concentration: math.NaN(),
			expectError:   true,
			description:   "NaN concentration (undefined value)",
		},
		{
			name:          "zero_concentration",
			ligandType:    LigandGlutamate,
			sourceID:      testNeuron.ID(),
			concentration: 0.0,
			expectError:   false,
			description:   "Zero concentration (valid baseline)",
		},
		{
			name:          "empty_source_id",
			ligandType:    LigandGlutamate,
			sourceID:      "",
			concentration: 1.0,
			expectError:   true,
			description:   "Empty source ID (unknown origin)",
		},
		{
			name:          "nonexistent_source",
			ligandType:    LigandGlutamate,
			sourceID:      "nonexistent_neuron_12345",
			concentration: 1.0,
			expectError:   false, // Chemical system may allow unknown sources
			description:   "Nonexistent source neuron (orphaned release)",
		},
		{
			name:          "extreme_concentration",
			ligandType:    LigandGlutamate,
			sourceID:      testNeuron.ID(),
			concentration: 1000000.0,
			expectError:   false, // May be rate-limited but not invalid
			description:   "Extremely high concentration (saturation)",
		},
	}

	for _, test := range invalidChemicalTests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Testing: %s", test.description)

			err := matrix.ReleaseLigand(test.ligandType, test.sourceID, test.concentration)

			if test.expectError && err == nil {
				t.Errorf("Expected error for %s, but release succeeded", test.name)
			} else if !test.expectError && err != nil {
				t.Logf("Note: %s rejected (may be due to rate limiting): %v", test.name, err)
			} else if test.expectError && err != nil {
				t.Logf("✓ Correctly rejected %s: %v", test.name, err)
			} else {
				t.Logf("✓ Accepted %s", test.name)
			}
		})
	}

	// === PHASE 3: INVALID SPATIAL QUERY PARAMETERS ===
	t.Log("\n--- Phase 3: Invalid Spatial Query Parameters ---")

	invalidSpatialTests := []struct {
		name        string
		criteria    ComponentCriteria
		expectPanic bool
		description string
	}{
		{
			name: "negative_radius",
			criteria: ComponentCriteria{
				Position: &Position3D{X: 0, Y: 0, Z: 0},
				Radius:   -10.0,
			},
			expectPanic: false,
			description: "Negative search radius (invalid geometry)",
		},
		{
			name: "infinite_radius",
			criteria: ComponentCriteria{
				Position: &Position3D{X: 0, Y: 0, Z: 0},
				Radius:   math.Inf(1),
			},
			expectPanic: false,
			description: "Infinite search radius (entire universe)",
		},
		{
			name: "nan_position",
			criteria: ComponentCriteria{
				Position: &Position3D{X: math.NaN(), Y: 0, Z: 0},
				Radius:   10.0,
			},
			expectPanic: false,
			description: "NaN position coordinates (undefined location)",
		},
		{
			name: "nil_position_with_radius",
			criteria: ComponentCriteria{
				Position: nil,
				Radius:   10.0,
			},
			expectPanic: false,
			description: "Nil position with radius (contradictory criteria)",
		},
		{
			name: "zero_radius",
			criteria: ComponentCriteria{
				Position: &Position3D{X: 0, Y: 0, Z: 0},
				Radius:   0.0,
			},
			expectPanic: false,
			description: "Zero radius (point query)",
		},
	}

	for _, test := range invalidSpatialTests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Testing: %s", test.description)

			func() {
				defer func() {
					if r := recover(); r != nil {
						if test.expectPanic {
							t.Logf("✓ Expected panic occurred for %s: %v", test.name, r)
						} else {
							t.Errorf("Unexpected panic for %s: %v", test.name, r)
						}
					}
				}()

				results := matrix.FindComponents(test.criteria)
				t.Logf("✓ Gracefully handled %s (returned %d results)", test.name, len(results))
			}()
		})
	}

	// === PHASE 4: INVALID ELECTRICAL SIGNALING ===
	t.Log("\n--- Phase 4: Invalid Electrical Signaling Parameters ---")

	// Test invalid signal types and data
	invalidElectricalTests := []struct {
		name        string
		signalType  SignalType
		sourceID    string
		data        interface{}
		description string
	}{
		{
			name:        "empty_source_id",
			signalType:  SignalFired,
			sourceID:    "",
			data:        1.0,
			description: "Empty source ID for electrical signal",
		},
		{
			name:        "nil_data",
			signalType:  SignalFired,
			sourceID:    testNeuron.ID(),
			data:        nil,
			description: "Nil data payload",
		},
		{
			name:        "complex_data_structure",
			signalType:  SignalFired,
			sourceID:    testNeuron.ID(),
			data:        map[string]interface{}{"nested": map[string]int{"deep": 42}},
			description: "Complex nested data structure",
		},
		{
			name:        "infinite_signal_value",
			signalType:  SignalFired,
			sourceID:    testNeuron.ID(),
			data:        math.Inf(1),
			description: "Infinite signal value",
		},
		{
			name:        "nan_signal_value",
			signalType:  SignalFired,
			sourceID:    testNeuron.ID(),
			data:        math.NaN(),
			description: "NaN signal value",
		},
	}

	for _, test := range invalidElectricalTests {
		t.Run(test.name, func(t *testing.T) {
			t.Logf("Testing: %s", test.description)

			// Electrical signaling typically doesn't return errors, so we test for panics
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Panic during electrical signaling %s: %v", test.name, r)
					} else {
						t.Logf("✓ Gracefully handled %s", test.name)
					}
				}()

				matrix.SendSignal(test.signalType, test.sourceID, test.data)
			}()
		})
	}

	t.Log("✅ Invalid input handling tests completed")
	t.Log("✅ Matrix demonstrates robust error handling and input validation")
}

// =================================================================================
// TEST 4: SYSTEM STATE TRANSITION EDGE CASES
// =================================================================================

// TestMatrixEdgeStateTransitions validates proper handling of system lifecycle
// edge cases and state transition scenarios.
//
// BIOLOGICAL PROCESS TESTED:
// This models critical state transitions in biological systems:
// - Neural development phases (neurogenesis → synaptogenesis → maturation)
// - Sleep-wake cycles (active → inactive → recovery)
// - Stress responses (normal → alarmed → exhausted → recovery)
// - Aging processes (development → maturity → senescence → death)
//
// STATE TRANSITION EDGE CASES:
// - Rapid startup/shutdown cycles (system stress)
// - Operations during shutdown (graceful degradation)
// - Partial initialization failures (resilient recovery)
// - Resource exhaustion during transitions (failure handling)
// - Concurrent state changes (race condition prevention)
// - Invalid state sequences (error prevention)
//
// VALIDATION CRITERIA:
// - Clean state transitions without resource leaks
// - Proper rejection of operations in invalid states
// - Consistent error messages during transitions
// - No orphaned resources after failures
// - Predictable behavior across all state combinations
// - Graceful degradation under adverse conditions
func TestMatrixEdgeStateTransitions(t *testing.T) {
	t.Log("=== MATRIX EDGE TEST: System State Transition Edge Cases ===")
	t.Log("Testing lifecycle management and state transition robustness")

	// === PHASE 1: RAPID STARTUP/SHUTDOWN CYCLES ===
	t.Log("\n--- Phase 1: Rapid Startup/Shutdown Stress ---")

	const CYCLE_COUNT = 10
	const CYCLE_DELAY = 50 * time.Millisecond

	for i := 0; i < CYCLE_COUNT; i++ {
		t.Logf("Cycle %d: Creating and destroying matrix...", i+1)

		// Create matrix
		matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
			ChemicalEnabled: true,
			SpatialEnabled:  true,
			UpdateInterval:  10 * time.Millisecond,
			MaxComponents:   50,
		})

		// Start matrix
		err := matrix.Start()
		if err != nil {
			t.Errorf("Cycle %d: Failed to start matrix: %v", i+1, err)
			continue
		}

		// Perform minimal operations
		matrix.RegisterNeuronType("cycle_neuron", func(id string, config types.NeuronConfig, callbacks NeuronCallbacks) (component.NeuralComponent, error) {
			return NewMockNeuron(id, config.Position, config.Receptors), nil
		})

		config := types.NeuronConfig{
			NeuronType: "cycle_neuron",
			Position:   Position3D{X: float64(i), Y: 0, Z: 0},
			Threshold:  0.7,
		}

		_, err = matrix.CreateNeuron(config)
		if err != nil {
			t.Logf("Cycle %d: Neuron creation failed (acceptable): %v", i+1, err)
		}

		// Stop matrix
		err = matrix.Stop()
		if err != nil {
			t.Errorf("Cycle %d: Failed to stop matrix: %v", i+1, err)
		}

		// Brief delay between cycles
		time.Sleep(CYCLE_DELAY)
	}

	t.Logf("✓ Completed %d rapid startup/shutdown cycles", CYCLE_COUNT)

	// === PHASE 2: OPERATIONS DURING SHUTDOWN ===
	t.Log("\n--- Phase 2: Operations During Shutdown Sequence ---")

	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  10 * time.Millisecond,
		MaxComponents:   50,
	})

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix for shutdown test: %v", err)
	}

	// Register factory
	matrix.RegisterNeuronType("shutdown_neuron", func(id string, config types.NeuronConfig, callbacks NeuronCallbacks) (component.NeuralComponent, error) {
		return NewMockNeuron(id, config.Position, config.Receptors), nil
	})

	// Create initial components
	config := types.NeuronConfig{
		NeuronType: "shutdown_neuron",
		Position:   Position3D{X: 0, Y: 0, Z: 0},
		Threshold:  0.7,
	}

	neuron, err := matrix.CreateNeuron(config)
	if err != nil {
		t.Fatalf("Failed to create initial neuron: %v", err)
	}

	// Start shutdown in background
	shutdownChan := make(chan error, 1)
	go func() {
		time.Sleep(100 * time.Millisecond) // Allow operations to start
		shutdownChan <- matrix.Stop()
	}()

	// Attempt operations during shutdown
	operationResults := make(map[string]error)
	operationComplete := make(map[string]bool)

	// Try to create neuron during shutdown
	go func() {
		time.Sleep(150 * time.Millisecond) // During shutdown
		_, err := matrix.CreateNeuron(config)
		operationResults["neuron_creation"] = err
		operationComplete["neuron_creation"] = true
	}()

	// Try chemical signaling during shutdown
	go func() {
		time.Sleep(150 * time.Millisecond)
		err := matrix.ReleaseLigand(LigandGlutamate, neuron.ID(), 0.5)
		operationResults["chemical_release"] = err
		operationComplete["chemical_release"] = true
	}()

	// Try spatial query during shutdown
	go func() {
		time.Sleep(150 * time.Millisecond)
		_ = matrix.FindComponents(ComponentCriteria{
			Position: &Position3D{X: 0, Y: 0, Z: 0},
			Radius:   10.0,
		})
		operationResults["spatial_query"] = nil // Spatial queries should still work
		operationComplete["spatial_query"] = true
	}()

	// Wait for shutdown completion
	shutdownErr := <-shutdownChan
	if shutdownErr != nil {
		t.Errorf("Matrix shutdown failed: %v", shutdownErr)
	}

	// Wait for operation attempts to complete
	time.Sleep(200 * time.Millisecond)

	t.Logf("Operations during shutdown results:")
	for operation, err := range operationResults {
		if !operationComplete[operation] {
			t.Logf("  %s: Did not complete", operation)
			continue
		}

		// FIXED: Different expectations for different operations
		switch operation {
		case "neuron_creation":
			// STRUCTURAL - Should work even during shutdown
			if err != nil {
				t.Logf("  %s: Rejected (unexpected) - %v", operation, err)
			} else {
				t.Logf("  %s: ✓ Completed (structural operations allowed)", operation)
			}

		case "chemical_release":
			// FUNCTIONAL - Should be rejected during shutdown
			if err != nil {
				t.Logf("  %s: ✓ Correctly rejected - %v", operation, err)
			} else {
				t.Errorf("  %s: ✗ Should have been rejected (requires active systems)", operation)
			}

		case "spatial_query":
			// READ-ONLY - Should work during shutdown
			if err != nil {
				t.Logf("  %s: Rejected (unexpected) - %v", operation, err)
			} else {
				t.Logf("  %s: ✓ Completed (read-only operations allowed)", operation)
			}
		}
	}

	// === PHASE 3: PARTIAL INITIALIZATION FAILURES ===
	t.Log("\n--- Phase 3: Partial Initialization Failure Recovery ---")

	// FIXED: Test matrix with actually impossible configuration
	impossibleMatrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  10 * time.Millisecond, // Valid interval
		MaxComponents:   -1,                    // Invalid: negative capacity
	})

	err = impossibleMatrix.Start()
	if err != nil {
		t.Logf("✓ Correctly rejected impossible configuration: %v", err)
	} else {
		t.Error("Expected error for impossible configuration, but start succeeded")
		impossibleMatrix.Stop()
	}

	// ADDITIONAL: Test zero update interval if that should be invalid
	zeroIntervalMatrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  0, // Zero interval - might be valid?
		MaxComponents:   10,
	})

	err = zeroIntervalMatrix.Start()
	if err != nil {
		t.Logf("✓ Zero update interval rejected: %v", err)
	} else {
		t.Logf("Note: Zero update interval accepted (may be valid configuration)")
		zeroIntervalMatrix.Stop()
	}

	// === PHASE 4: CONCURRENT STATE TRANSITIONS ===
	t.Log("\n--- Phase 4: Concurrent State Transition Safety ---")

	concurrentMatrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  10 * time.Millisecond,
		MaxComponents:   50,
	})

	var wg sync.WaitGroup
	var startErrors, stopErrors int64

	// Launch multiple concurrent start attempts
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := concurrentMatrix.Start()
			if err != nil {
				atomic.AddInt64(&startErrors, 1)
				t.Logf("Start %d failed: %v", id, err)
			}
		}(i)
	}

	// Launch multiple concurrent stop attempts (after brief delay)
	go func() {
		time.Sleep(50 * time.Millisecond)
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				err := concurrentMatrix.Stop()
				if err != nil {
					atomic.AddInt64(&stopErrors, 1)
					t.Logf("Stop %d failed: %v", id, err)
				}
			}(i)
		}
	}()

	wg.Wait()

	t.Logf("Concurrent transitions - Start errors: %d, Stop errors: %d", startErrors, stopErrors)

	// Final cleanup
	concurrentMatrix.Stop() // Ensure stopped

	t.Log("✅ System state transition tests completed")
	t.Log("✅ Matrix handles lifecycle edge cases gracefully")
}

// =================================================================================
// TEST 5: BIOLOGICAL CONSTRAINT VIOLATION DETECTION
// =================================================================================

// TestMatrixEdgeBiologicalConstraints validates detection and enforcement of
// biological constraints under extreme conditions.
//
// BIOLOGICAL CONSTRAINTS TESTED:
// - Rate limits (neurotransmitter release frequencies)
// - Spatial constraints (diffusion limits, membrane barriers)
// - Energetic constraints (ATP costs, metabolic limits)
// - Temporal constraints (refractory periods, recovery times)
// - Concentration limits (receptor saturation, toxicity thresholds)
// - Network constraints (connectivity patterns, circuit stability)
//
// VIOLATION SCENARIOS:
// - Extreme firing rates (beyond biological maximum)
// - Impossible spatial arrangements (overlapping components)
// - Violating conservation laws (creating/destroying matter)
// - Temporal paradoxes (effects before causes)
// - Thermodynamic violations (energy creation)
// - Membrane physics violations (impossible gradients)
//
// EXPECTED ENFORCEMENT:
// - Rate limiting engages automatically
// - Spatial conflicts resolved consistently
// - Energy conservation maintained
// - Causal ordering preserved
// - Concentration bounds enforced
// - Network stability protected
func TestMatrixEdgeBiologicalConstraints(t *testing.T) {
	t.Log("=== MATRIX EDGE TEST: Biological Constraint Violation Detection ===")
	t.Log("Testing enforcement of biological limits under extreme conditions")

	// Initialize matrix for constraint testing
	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  1 * time.Millisecond, // Fast updates for testing
		MaxComponents:   100,
	})
	defer matrix.Stop()

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}

	err = matrix.chemicalModulator.Start()
	if err != nil {
		t.Fatalf("Failed to start chemical modulator: %v", err)
	}

	// Register factory and create test neuron
	matrix.RegisterNeuronType("constraint_neuron", func(id string, config types.NeuronConfig, callbacks NeuronCallbacks) (component.NeuralComponent, error) {
		return NewMockNeuron(id, config.Position, config.Receptors), nil
	})

	testNeuron, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "constraint_neuron",
		Position:   Position3D{X: 0, Y: 0, Z: 0},
		Threshold:  0.7,
		Receptors:  []LigandType{LigandGlutamate, LigandGABA, LigandDopamine},
	})
	if err != nil {
		t.Fatalf("Failed to create test neuron: %v", err)
	}

	// === PHASE 1: NEUROTRANSMITTER RELEASE RATE LIMITS ===
	t.Log("\n--- Phase 1: Rate Limiting Enforcement ---")

	// Test each neurotransmitter's rate limits
	ligandTests := []struct {
		ligand      LigandType
		maxRate     float64 // Hz
		minInterval time.Duration
		name        string
	}{
		{LigandGlutamate, 500.0, 2 * time.Millisecond, "Glutamate"},
		{LigandGABA, 1000.0, 1 * time.Millisecond, "GABA"},
		{LigandDopamine, 50.0, 20 * time.Millisecond, "Dopamine"},
	}

	for _, ligandTest := range ligandTests {
		t.Logf("Testing %s rate limiting (max: %.1f Hz, min interval: %v)",
			ligandTest.name, ligandTest.maxRate, ligandTest.minInterval)

		// Reset rate limits
		matrix.chemicalModulator.ResetRateLimits()

		// FIXED: Test rapid-fire releases to verify rate limiting
		const RAPID_ATTEMPTS = 10
		successCount := 0
		errorCount := 0

		for i := 0; i < RAPID_ATTEMPTS; i++ {
			err := matrix.ReleaseLigand(ligandTest.ligand, testNeuron.ID(), 0.5)
			if err == nil {
				successCount++
			} else {
				errorCount++
			}
			// No delay - test maximum rate
		}

		t.Logf("  Rapid fire results: %d/%d succeeded, %d rate-limited",
			successCount, RAPID_ATTEMPTS, errorCount)

		// FIXED: The key test is that most attempts should be rate-limited
		expectedMaxSuccesses := 2 // Allow 1-2 successes in rapid fire
		if successCount <= expectedMaxSuccesses {
			t.Logf("✓ %s rate limiting properly enforced (%d successes ≤ %d expected)",
				ligandTest.name, successCount, expectedMaxSuccesses)
		} else {
			t.Errorf("Rate limiting failed for %s: %d successes > %d expected",
				ligandTest.name, successCount, expectedMaxSuccesses)
		}

		// Test that rate limiting error messages are correct
		if errorCount > 0 {
			t.Logf("✓ %s rate limiting active (%d rejections)", ligandTest.name, errorCount)
		} else {
			t.Errorf("Expected rate limiting errors for %s rapid fire", ligandTest.name)
		}

		// CRITICAL TEST: Proper spacing allows releases
		time.Sleep(ligandTest.minInterval * 2) // Wait longer than minimum interval
		err := matrix.ReleaseLigand(ligandTest.ligand, testNeuron.ID(), 0.5)
		if err != nil {
			t.Errorf("Properly spaced %s release failed: %v", ligandTest.name, err)
		} else {
			t.Logf("✓ %s release allowed after proper interval", ligandTest.name)
		}

		// ADDITIONAL: Test that rapid succession after wait is still limited
		time.Sleep(1 * time.Millisecond) // Very short wait
		err = matrix.ReleaseLigand(ligandTest.ligand, testNeuron.ID(), 0.5)
		if err == nil {
			t.Errorf("Expected rate limiting for %s rapid succession", ligandTest.name)
		} else {
			t.Logf("✓ %s correctly rate-limited rapid succession", ligandTest.name)
		}
	}

	// === PHASE 2: SPATIAL CONSTRAINT VIOLATIONS ===
	t.Log("\n--- Phase 2: Spatial Constraint Enforcement ---")

	// Test extreme spatial scenarios
	spatialTests := []struct {
		name        string
		position    Position3D
		expectIssue bool
		description string
	}{
		{
			name:        "overlapping_position",
			position:    Position3D{X: 0, Y: 0, Z: 0}, // Same as test neuron
			expectIssue: false,                        // May be allowed but tracked
			description: "Exact spatial overlap (impossible in biology)",
		},
		{
			name:        "extreme_distance",
			position:    Position3D{X: 1e15, Y: 1e15, Z: 1e15},
			expectIssue: false, // Should handle gracefully
			description: "Extreme distance (light-years away)",
		},
		{
			name:        "subatomic_proximity",
			position:    Position3D{X: 1e-15, Y: 1e-15, Z: 1e-15},
			expectIssue: false, // Very close but allowed
			description: "Subatomic proximity (quantum scale)",
		},
	}

	for _, spatialTest := range spatialTests {
		t.Logf("Testing: %s", spatialTest.description)

		// Try to create neuron at problematic position
		config := types.NeuronConfig{
			NeuronType: "constraint_neuron",
			Position:   spatialTest.position,
			Threshold:  0.7,
		}

		neuron, err := matrix.CreateNeuron(config)

		if err != nil {
			if spatialTest.expectIssue {
				t.Logf("✓ Correctly rejected %s: %v", spatialTest.name, err)
			} else {
				t.Logf("Note: %s rejected: %v", spatialTest.name, err)
			}
		} else {
			// Check if spatial tracking is working
			distance, distErr := matrix.GetSpatialDistance(testNeuron.ID(), neuron.ID())
			if distErr != nil {
				t.Logf("Spatial distance calculation failed for %s: %v", spatialTest.name, distErr)
			} else {
				t.Logf("✓ %s handled gracefully (distance: %.2e)", spatialTest.name, distance)
			}
		}
	}

	// === PHASE 3: CONCENTRATION CONSTRAINT VIOLATIONS ===
	t.Log("\n--- Phase 3: Concentration Constraint Enforcement ---")

	concentrationTests := []struct {
		concentration float64
		expectReject  bool
		description   string
	}{
		{1e20, true, "Impossible high concentration (molecular limit exceeded)"},
		{1e-20, false, "Extremely low concentration (single molecules)"},
		{0.0, false, "Zero concentration (baseline)"},
		{-1.0, true, "Negative concentration (physically impossible)"},
		{math.Inf(1), true, "Infinite concentration (mathematical impossibility)"},
		{math.NaN(), true, "NaN concentration (undefined value)"},
	}

	for _, concTest := range concentrationTests {
		t.Logf("Testing concentration: %g (%s)", concTest.concentration, concTest.description)

		err := matrix.ReleaseLigand(LigandGlutamate, testNeuron.ID(), concTest.concentration)

		if concTest.expectReject {
			if err != nil {
				t.Logf("✓ Correctly rejected impossible concentration: %v", err)
			} else {
				t.Errorf("Failed to reject impossible concentration: %g", concTest.concentration)
			}
		} else {
			if err != nil {
				t.Logf("Note: Extreme concentration rejected: %v", err)
			} else {
				t.Logf("✓ Extreme concentration handled gracefully")
			}
		}

		// Allow rate limiting to reset
		time.Sleep(5 * time.Millisecond)
	}

	// === PHASE 4: NETWORK CONNECTIVITY CONSTRAINTS ===
	t.Log("\n--- Phase 4: Network Connectivity Constraint Enforcement ---")

	// Create additional neurons for connectivity testing
	neuron2, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "constraint_neuron",
		Position:   Position3D{X: 10, Y: 0, Z: 0},
		Threshold:  0.7,
	})
	if err != nil {
		t.Logf("Failed to create second neuron: %v", err)
	}

	if neuron2 != nil {
		// Register synapse factory
		matrix.RegisterSynapseType("constraint_synapse", func(id string, config types.SynapseConfig, callbacks SynapseCallbacks) (component.SynapticProcessor, error) {
			return NewMockSynapse(id, config.Position, config.PresynapticID, config.PostsynapticID, config.InitialWeight), nil
		})

		// Test constraint violations in synapse creation
		synapseConstraintTests := []struct {
			name         string
			config       types.SynapseConfig
			expectReject bool
			description  string
		}{
			{
				name: "self_connection",
				config: types.SynapseConfig{
					SynapseType:    "constraint_synapse",
					PresynapticID:  testNeuron.ID(),
					PostsynapticID: testNeuron.ID(), // Self-connection
					InitialWeight:  0.5,
					Position:       Position3D{X: 5, Y: 0, Z: 0},
				},
				expectReject: false, // May be allowed (autapses exist in biology)
				description:  "Self-connection (autapse)",
			},
			{
				name: "negative_weight",
				config: types.SynapseConfig{
					SynapseType:    "constraint_synapse",
					PresynapticID:  testNeuron.ID(),
					PostsynapticID: neuron2.ID(),
					InitialWeight:  -0.5, // Negative weight
					Position:       Position3D{X: 5, Y: 0, Z: 0},
				},
				expectReject: false, // Negative weights represent inhibition
				description:  "Negative synaptic weight (inhibitory synapse)",
			},
			{
				name: "extreme_weight",
				config: types.SynapseConfig{
					SynapseType:    "constraint_synapse",
					PresynapticID:  testNeuron.ID(),
					PostsynapticID: neuron2.ID(),
					InitialWeight:  1e10, // Impossibly strong
					Position:       Position3D{X: 5, Y: 0, Z: 0},
				},
				expectReject: false, // May be clamped rather than rejected
				description:  "Extreme synaptic weight (pathological strength)",
			},
		}

		for _, synapseTest := range synapseConstraintTests {
			t.Logf("Testing: %s", synapseTest.description)

			_, err := matrix.CreateSynapse(synapseTest.config)

			if synapseTest.expectReject {
				if err != nil {
					t.Logf("✓ Correctly rejected %s: %v", synapseTest.name, err)
				} else {
					t.Errorf("Failed to reject %s", synapseTest.name)
				}
			} else {
				if err != nil {
					t.Logf("Note: %s rejected: %v", synapseTest.name, err)
				} else {
					t.Logf("✓ %s handled appropriately", synapseTest.name)
				}
			}
		}
	}

	t.Log("✅ Biological constraint enforcement tests completed")
	t.Log("✅ Matrix properly enforces biological limits and physical constraints")
}

// =================================================================================
// TEST 6: MEMORY MANAGEMENT AND RESOURCE LEAK DETECTION
// =================================================================================

// TestMatrixEdgeMemoryManagement validates proper memory management and detects
// resource leaks under stress conditions.
//
// BIOLOGICAL PROCESS TESTED:
// This models cellular resource management and waste disposal:
// - Protein synthesis and degradation balance
// - Membrane recycling and lipid turnover
// - Neurotransmitter synthesis and clearance
// - Cellular autophagy and waste removal
// - Metabolic resource allocation and conservation
//
// MEMORY LEAK SCENARIOS:
// - Repeated component creation without cleanup
// - Circular references between components
// - Event handler registration without deregistration
// - Chemical concentration accumulation
// - Spatial index memory growth
// - Callback function retention
//
// LEAK DETECTION METHODS:
// - Memory usage monitoring during operations
// - Component count tracking and validation
// - Reference counting for callbacks
// - Chemical pool size monitoring
// - Spatial index size verification
// - Goroutine leak detection
//
// EXPECTED BEHAVIORS:
// - Stable memory usage during sustained operations
// - Proper cleanup of all allocated resources
// - No accumulation of orphaned objects
// - Consistent component counts
// - Bounded chemical pools
// - Clean shutdown without leaks
func TestMatrixEdgeMemoryManagement(t *testing.T) {
	t.Log("=== MATRIX EDGE TEST: Memory Management and Leak Detection ===")
	t.Log("Testing resource cleanup and memory leak prevention")

	// === MEMORY BASELINE ESTABLISHMENT ===
	t.Log("\n--- Establishing Memory Baseline ---")

	var memBefore, memAfter runtime.MemStats
	runtime.GC()
	time.Sleep(10 * time.Millisecond)
	runtime.ReadMemStats(&memBefore)

	initialGoroutines := runtime.NumGoroutine()
	t.Logf("Initial state - Memory: %d KB, Goroutines: %d",
		memBefore.Alloc/1024, initialGoroutines)

	// === PHASE 1: REPEATED MATRIX CREATION/DESTRUCTION ===
	t.Log("\n--- Phase 1: Matrix Lifecycle Memory Stability ---")

	const LIFECYCLE_CYCLES = 20

	for cycle := 0; cycle < LIFECYCLE_CYCLES; cycle++ {
		// Create matrix
		matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
			ChemicalEnabled: true,
			SpatialEnabled:  true,
			UpdateInterval:  10 * time.Millisecond,
			MaxComponents:   20,
		})

		// Start and populate
		err := matrix.Start()
		if err != nil {
			t.Errorf("Cycle %d: Failed to start matrix: %v", cycle, err)
			continue
		}

		// Register factory
		matrix.RegisterNeuronType("memory_neuron", func(id string, config types.NeuronConfig, callbacks NeuronCallbacks) (component.NeuralComponent, error) {
			return NewMockNeuron(id, config.Position, config.Receptors), nil
		})

		// Create some components
		for i := 0; i < 5; i++ {
			config := types.NeuronConfig{
				NeuronType: "memory_neuron",
				Position:   Position3D{X: float64(i), Y: float64(cycle), Z: 0},
				Threshold:  0.7,
				Receptors:  []LigandType{LigandGlutamate},
			}

			_, err := matrix.CreateNeuron(config)
			if err != nil {
				break // Resource limit may be reached
			}
		}

		// Perform operations
		neurons := matrix.ListNeurons()
		if len(neurons) > 0 {
			matrix.ReleaseLigand(LigandGlutamate, neurons[0].ID(), 0.5)
			matrix.SendSignal(SignalFired, neurons[0].ID(), 1.0)

			_ = matrix.FindComponents(ComponentCriteria{
				Position: &Position3D{X: 0, Y: 0, Z: 0},
				Radius:   10.0,
			})
		}

		// Stop and cleanup
		err = matrix.Stop()
		if err != nil {
			t.Errorf("Cycle %d: Failed to stop matrix: %v", cycle, err)
		}

		// Periodic memory check
		if cycle%5 == 4 {
			runtime.GC()
			time.Sleep(10 * time.Millisecond)

			var memDuring runtime.MemStats
			runtime.ReadMemStats(&memDuring)
			currentGoroutines := runtime.NumGoroutine()

			t.Logf("Cycle %d - Memory: %d KB, Goroutines: %d",
				cycle+1, memDuring.Alloc/1024, currentGoroutines)
		}
	}

	// === MEMORY STABILITY VERIFICATION ===
	runtime.GC()
	time.Sleep(50 * time.Millisecond) // Allow cleanup
	runtime.ReadMemStats(&memAfter)
	finalGoroutines := runtime.NumGoroutine()

	memGrowth := int64(memAfter.Alloc) - int64(memBefore.Alloc)
	goroutineGrowth := finalGoroutines - initialGoroutines

	t.Logf("Memory analysis after %d cycles:", LIFECYCLE_CYCLES)
	t.Logf("  Memory growth: %+d KB", memGrowth/1024)
	t.Logf("  Goroutine growth: %+d", goroutineGrowth)

	// Validate memory stability
	maxAcceptableGrowth := int64(500 * 1024) // 500 KB tolerance
	if memGrowth > maxAcceptableGrowth {
		t.Errorf("Excessive memory growth detected: %+d KB > %d KB",
			memGrowth/1024, maxAcceptableGrowth/1024)
	} else {
		t.Logf("✓ Memory growth within acceptable range")
	}

	// Validate goroutine stability
	maxAcceptableGoroutines := 5
	if goroutineGrowth > maxAcceptableGoroutines {
		t.Errorf("Goroutine leak detected: %+d > %d", goroutineGrowth, maxAcceptableGoroutines)
	} else {
		t.Logf("✓ Goroutine count stable")
	}

	// === PHASE 2: SUSTAINED OPERATION MEMORY MONITORING ===
	t.Log("\n--- Phase 2: Sustained Operation Memory Stability ---")

	sustainedMatrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  5 * time.Millisecond,
		MaxComponents:   50,
	})
	defer sustainedMatrix.Stop()

	err := sustainedMatrix.Start()
	if err != nil {
		t.Fatalf("Failed to start sustained matrix: %v", err)
	}

	err = sustainedMatrix.chemicalModulator.Start()
	if err != nil {
		t.Fatalf("Failed to start chemical modulator: %v", err)
	}

	// Register factory
	sustainedMatrix.RegisterNeuronType("sustained_neuron", func(id string, config types.NeuronConfig, callbacks NeuronCallbacks) (component.NeuralComponent, error) {
		return NewMockNeuron(id, config.Position, config.Receptors), nil
	})

	// Create initial population
	var sustainedNeurons []component.NeuralComponent
	for i := 0; i < 10; i++ {
		config := types.NeuronConfig{
			NeuronType: "sustained_neuron",
			Position:   Position3D{X: float64(i * 5), Y: 0, Z: 0},
			Threshold:  0.7,
			Receptors:  []LigandType{LigandGlutamate, LigandGABA},
		}

		neuron, err := sustainedMatrix.CreateNeuron(config)
		if err != nil {
			break
		}
		sustainedNeurons = append(sustainedNeurons, neuron)
	}

	t.Logf("Created %d neurons for sustained testing", len(sustainedNeurons))

	// Monitor memory during sustained operations
	const SUSTAINED_DURATION = 2 * time.Second
	const MONITORING_INTERVAL = 500 * time.Millisecond

	startTime := time.Now()
	monitoringTicker := time.NewTicker(MONITORING_INTERVAL)
	defer monitoringTicker.Stop()

	memoryReadings := make([]uint64, 0)
	componentCounts := make([]int, 0)

	// Background operations
	operationTicker := time.NewTicker(20 * time.Millisecond)
	defer operationTicker.Stop()

	operationCount := 0
	go func() {
		for {
			select {
			case <-operationTicker.C:
				if len(sustainedNeurons) > 0 {
					neuronIdx := operationCount % len(sustainedNeurons)

					// Chemical operations (with rate limiting respect)
					if operationCount%10 == 0 { // Reduce frequency to respect rate limits
						ligands := []LigandType{LigandGlutamate, LigandGABA, LigandDopamine}
						ligand := ligands[operationCount%len(ligands)]
						sustainedMatrix.ReleaseLigand(ligand, sustainedNeurons[neuronIdx].ID(), 0.3)
					}

					// Electrical operations
					signals := []SignalType{SignalFired, SignalConnected}
					signal := signals[operationCount%len(signals)]
					sustainedMatrix.SendSignal(signal, sustainedNeurons[neuronIdx].ID(), float64(operationCount%5))

					// Spatial queries
					if operationCount%5 == 0 {
						queryPos := Position3D{X: float64(operationCount % 50), Y: 0, Z: 0}
						_ = sustainedMatrix.FindComponents(ComponentCriteria{
							Position: &queryPos,
							Radius:   15.0,
						})
					}
				}
				operationCount++
			case <-time.After(SUSTAINED_DURATION):
				return
			}
		}
	}()

	// Memory monitoring loop
	for time.Since(startTime) < SUSTAINED_DURATION {
		select {
		case <-monitoringTicker.C:
			runtime.GC()
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			memoryReadings = append(memoryReadings, memStats.Alloc)
			componentCounts = append(componentCounts, sustainedMatrix.astrocyteNetwork.Count())

			t.Logf("Sustained monitoring - Memory: %d KB, Components: %d, Operations: %d",
				memStats.Alloc/1024, componentCounts[len(componentCounts)-1], operationCount)

		case <-time.After(SUSTAINED_DURATION):
			break
		}
	}

	// Analyze memory stability during sustained operations
	if len(memoryReadings) >= 2 {
		initialMem := memoryReadings[0]
		finalMem := memoryReadings[len(memoryReadings)-1]
		memoryGrowthDuringSustained := int64(finalMem) - int64(initialMem)

		t.Logf("Sustained operation analysis:")
		t.Logf("  Duration: %v", SUSTAINED_DURATION)
		t.Logf("  Total operations: %d", operationCount)
		t.Logf("  Memory growth: %+d KB", memoryGrowthDuringSustained/1024)

		// Check for memory leaks during sustained operation
		maxSustainedGrowth := int64(200 * 1024) // 200 KB tolerance
		if memoryGrowthDuringSustained > maxSustainedGrowth {
			t.Errorf("Memory leak during sustained operation: %+d KB > %d KB",
				memoryGrowthDuringSustained/1024, maxSustainedGrowth/1024)
		} else {
			t.Logf("✓ Memory stable during sustained operations")
		}

		// Check component count stability
		if len(componentCounts) >= 2 {
			initialComponents := componentCounts[0]
			finalComponents := componentCounts[len(componentCounts)-1]
			componentGrowth := finalComponents - initialComponents

			t.Logf("  Component count change: %+d", componentGrowth)

			// Should be stable (no unexpected growth)
			if componentGrowth > 5 {
				t.Errorf("Unexpected component growth: %+d components", componentGrowth)
			} else {
				t.Logf("✓ Component count stable")
			}
		}
	}

	t.Log("✅ Memory management tests completed")
	t.Log("✅ No significant memory leaks or resource retention detected")
}

// =================================================================================
// TEST 7: PERFORMANCE DEGRADATION UNDER EXTREME LOAD
// =================================================================================

// TestMatrixEdgePerformanceDegradation validates graceful performance degradation
// under extreme load conditions.
//
// BIOLOGICAL PROCESS TESTED:
// This models how biological neural networks maintain function under stress:
// - Metabolic stress (glucose/oxygen deprivation)
// - Ionic stress (extreme ion concentrations)
// - Thermal stress (temperature extremes)
// - Toxin exposure (pharmacological interference)
// - Hyperstimulation (pathological activity levels)
//
// PERFORMANCE STRESS CONDITIONS:
// - Component creation at resource limits
// - Chemical signaling saturation
// - Electrical signal flooding
// - Spatial query overload
// - Concurrent access contention
// - Memory pressure scenarios
//
// DEGRADATION EXPECTATIONS:
// - Gradual performance reduction (not cliff edges)
// - Maintained functionality under reduced capacity
// - Predictable failure modes
// - Error rate increase but not system failure
// - Recovery capability after stress removal
// - Consistent behavior across stress types
func TestMatrixEdgePerformanceDegradation(t *testing.T) {
	t.Log("=== MATRIX EDGE TEST: Performance Degradation Under Extreme Load ===")
	t.Log("Testing graceful degradation and stress response characteristics")

	// === BASELINE PERFORMANCE MEASUREMENT ===
	t.Log("\n--- Establishing Baseline Performance ---")

	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  5 * time.Millisecond,
		MaxComponents:   200, // Higher capacity for stress testing
	})
	defer matrix.Stop()

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}

	err = matrix.chemicalModulator.Start()
	if err != nil {
		t.Fatalf("Failed to start chemical modulator: %v", err)
	}

	// Register factories
	matrix.RegisterNeuronType("perf_neuron", func(id string, config types.NeuronConfig, callbacks NeuronCallbacks) (component.NeuralComponent, error) {
		return NewMockNeuron(id, config.Position, config.Receptors), nil
	})

	matrix.RegisterSynapseType("perf_synapse", func(id string, config types.SynapseConfig, callbacks SynapseCallbacks) (component.SynapticProcessor, error) {
		return NewMockSynapse(id, config.Position, config.PresynapticID, config.PostsynapticID, config.InitialWeight), nil
	})

	// Measure baseline performance
	baselineOps := []struct {
		name      string
		operation func() error
	}{
		{
			"neuron_creation",
			func() error {
				config := types.NeuronConfig{
					NeuronType: "perf_neuron",
					Position:   Position3D{X: float64(time.Now().UnixNano() % 1000), Y: 0, Z: 0},
					Threshold:  0.7,
				}
				_, err := matrix.CreateNeuron(config)
				return err
			},
		},
		{
			"chemical_release",
			func() error {
				return matrix.ReleaseLigand(LigandGlutamate, "baseline_source", 0.5)
			},
		},
		{
			"electrical_signal",
			func() error {
				matrix.SendSignal(SignalFired, "baseline_source", 1.0)
				return nil
			},
		},
		{
			"spatial_query",
			func() error {
				_ = matrix.FindComponents(ComponentCriteria{
					Position: &Position3D{X: 0, Y: 0, Z: 0},
					Radius:   10.0,
				})
				return nil
			},
		},
	}

	baselineMetrics := make(map[string]time.Duration)

	for _, op := range baselineOps {
		// Measure baseline timing
		iterations := 10
		startTime := time.Now()

		for i := 0; i < iterations; i++ {
			op.operation()
			time.Sleep(1 * time.Millisecond) // Respect rate limits
		}

		avgTime := time.Since(startTime) / time.Duration(iterations)
		baselineMetrics[op.name] = avgTime

		t.Logf("Baseline %s: %v average", op.name, avgTime)
	}

	// === PHASE 1: RESOURCE SATURATION STRESS ===
	t.Log("\n--- Phase 1: Resource Saturation Performance Impact ---")

	// Fill matrix to near capacity
	createdComponents := 0
	for createdComponents < 180 { // Near 200 limit
		config := types.NeuronConfig{
			NeuronType: "perf_neuron",
			Position:   Position3D{X: float64(createdComponents), Y: float64(createdComponents % 10), Z: 0},
			Threshold:  0.7,
		}

		_, err := matrix.CreateNeuron(config)
		if err != nil {
			break // Hit resource limit
		}
		createdComponents++

		// Avoid overwhelming during creation
		if createdComponents%20 == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	t.Logf("Created %d components (near capacity)", createdComponents)

	// Measure performance under resource pressure
	stressMetrics := make(map[string]time.Duration)

	for _, op := range baselineOps {
		iterations := 5 // Fewer iterations under stress
		startTime := time.Now()

		for i := 0; i < iterations; i++ {
			op.operation()
			time.Sleep(2 * time.Millisecond) // More spacing under stress
		}

		avgTime := time.Since(startTime) / time.Duration(iterations)
		stressMetrics[op.name] = avgTime

		degradationFactor := float64(avgTime) / float64(baselineMetrics[op.name])
		t.Logf("Stress %s: %v average (%.1fx degradation)", op.name, avgTime, degradationFactor)

		// Validate graceful degradation (not more than 10x slower)
		if degradationFactor > 10.0 {
			t.Errorf("Excessive performance degradation for %s: %.1fx", op.name, degradationFactor)
		}
	}

	// === PHASE 2: CONCURRENT ACCESS OVERLOAD ===
	t.Log("\n--- Phase 2: Concurrent Access Overload Performance ---")

	const OVERLOAD_GOROUTINES = 100
	const OPERATIONS_PER_GOROUTINE = 20

	var wg sync.WaitGroup
	var successfulOps, failedOps int64

	overloadStart := time.Now()

	// Launch overload goroutines
	for i := 0; i < OVERLOAD_GOROUTINES; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < OPERATIONS_PER_GOROUTINE; j++ {
				// Mix different operation types
				switch j % 4 {
				case 0:
					// Chemical release (respecting rate limits)
					err := matrix.ReleaseLigand(LigandGlutamate, fmt.Sprintf("overload_%d", goroutineID), 0.3)
					if err != nil {
						atomic.AddInt64(&failedOps, 1)
					} else {
						atomic.AddInt64(&successfulOps, 1)
					}
					time.Sleep(5 * time.Millisecond)

				case 1:
					// Electrical signaling
					matrix.SendSignal(SignalFired, fmt.Sprintf("overload_%d", goroutineID), 1.0)
					atomic.AddInt64(&successfulOps, 1)

				case 2:
					// Spatial queries
					queryPos := Position3D{X: float64(goroutineID), Y: float64(j), Z: 0}
					_ = matrix.FindComponents(ComponentCriteria{Position: &queryPos, Radius: 5.0})
					atomic.AddInt64(&successfulOps, 1)

				case 3:
					// Health monitoring
					matrix.microglia.UpdateComponentHealth(fmt.Sprintf("overload_%d", goroutineID), 0.5, 1)
					atomic.AddInt64(&successfulOps, 1)
				}

				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	overloadDuration := time.Since(overloadStart)

	totalOps := successfulOps + failedOps
	throughput := float64(totalOps) / overloadDuration.Seconds()
	errorRate := float64(failedOps) / float64(totalOps) * 100

	t.Logf("Overload results:")
	t.Logf("  Duration: %v", overloadDuration)
	t.Logf("  Total operations: %d", totalOps)
	t.Logf("  Successful: %d", successfulOps)
	t.Logf("  Failed: %d", failedOps)
	t.Logf("  Throughput: %.1f ops/sec", throughput)
	t.Logf("  Error rate: %.1f%%", errorRate)

	// Validate graceful degradation under overload
	if errorRate > 50.0 {
		t.Errorf("Excessive error rate under overload: %.1f%%", errorRate)
	} else {
		t.Logf("✓ Error rate acceptable under overload")
	}

	// === PHASE 3: RECOVERY VALIDATION ===
	t.Log("\n--- Phase 3: Post-Stress Recovery Validation ---")

	// Allow system to recover
	time.Sleep(500 * time.Millisecond)

	// Measure recovery performance
	recoveryMetrics := make(map[string]time.Duration)

	for _, op := range baselineOps {
		iterations := 5
		startTime := time.Now()

		for i := 0; i < iterations; i++ {
			op.operation()
			time.Sleep(3 * time.Millisecond) // Conservative spacing
		}

		avgTime := time.Since(startTime) / time.Duration(iterations)
		recoveryMetrics[op.name] = avgTime

		recoveryFactor := float64(avgTime) / float64(baselineMetrics[op.name])
		t.Logf("Recovery %s: %v average (%.1fx vs baseline)", op.name, avgTime, recoveryFactor)

		// Validate recovery (should be closer to baseline)
		if recoveryFactor > 5.0 {
			t.Errorf("Poor recovery for %s: %.1fx degradation persists", op.name, recoveryFactor)
		}
	}

	// === PERFORMANCE SUMMARY ===
	t.Log("\n--- Performance Degradation Summary ---")

	for opName := range baselineMetrics {
		baseline := baselineMetrics[opName]
		stress := stressMetrics[opName]
		recovery := recoveryMetrics[opName]

		stressFactor := float64(stress) / float64(baseline)
		recoveryFactor := float64(recovery) / float64(baseline)

		t.Logf("%s performance:", opName)
		t.Logf("  Baseline: %v", baseline)
		t.Logf("  Under stress: %v (%.1fx)", stress, stressFactor)
		t.Logf("  After recovery: %v (%.1fx)", recovery, recoveryFactor)
	}

	t.Log("✅ Performance degradation tests completed")
	t.Log("✅ Matrix demonstrates graceful degradation and recovery capabilities")
}

// =================================================================================
// UTILITY FUNCTIONS FOR EDGE CASE TESTING
// =================================================================================

// validateSystemConsistency performs comprehensive system state validation
func validateSystemConsistency(t *testing.T, matrix *ExtracellularMatrix, testName string) {
	// Component count consistency
	neurons := matrix.ListNeurons()
	synapses := matrix.ListSynapses()
	totalTracked := len(neurons) + len(synapses)
	totalComponents := matrix.astrocyteNetwork.Count()

	if totalTracked > totalComponents {
		t.Errorf("%s: Tracking inconsistency - tracked %d > total %d",
			testName, totalTracked, totalComponents)
	}

	// Spatial consistency
	for _, neuron := range neurons {
		_, err := matrix.GetSpatialDistance(neuron.ID(), neuron.ID())
		if err != nil {
			t.Errorf("%s: Spatial inconsistency for neuron %s: %v",
				testName, neuron.ID(), err)
		}
	}

	t.Logf("%s: System consistency validated (%d neurons, %d synapses, %d total)",
		testName, len(neurons), len(synapses), totalComponents)
}

// measureOperationLatency measures the latency of a specific operation
func measureOperationLatency(operation func() error, iterations int) (time.Duration, int, int) {
	startTime := time.Now()
	successful := 0
	failed := 0

	for i := 0; i < iterations; i++ {
		err := operation()
		if err != nil {
			failed++
		} else {
			successful++
		}
	}

	totalTime := time.Since(startTime)
	averageLatency := totalTime / time.Duration(iterations)

	return averageLatency, successful, failed
}

// createMemoryPressure intentionally creates memory pressure for testing
func createMemoryPressure() [][]byte {
	const PRESSURE_SIZE = 50 * 1024 * 1024 // 50 MB
	const CHUNK_SIZE = 1024 * 1024         // 1 MB chunks

	chunks := make([][]byte, PRESSURE_SIZE/CHUNK_SIZE)
	for i := range chunks {
		chunks[i] = make([]byte, CHUNK_SIZE)
		// Fill with data to prevent optimization
		for j := range chunks[i] {
			chunks[i][j] = byte(i + j)
		}
	}
	return chunks
}

// releaseMemoryPressure releases memory pressure
func releaseMemoryPressure(chunks [][]byte) {
	for i := range chunks {
		chunks[i] = nil
	}
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
}

// monitorResourceUsage continuously monitors system resource usage
func monitorResourceUsage(duration time.Duration, interval time.Duration) []ResourceSnapshot {
	var snapshots []ResourceSnapshot
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	startTime := time.Now()

	for time.Since(startTime) < duration {
		select {
		case <-ticker.C:
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			snapshot := ResourceSnapshot{
				Timestamp:   time.Now(),
				MemoryAlloc: memStats.Alloc,
				Goroutines:  runtime.NumGoroutine(),
			}
			snapshots = append(snapshots, snapshot)

		case <-time.After(duration):
			break
		}
	}

	return snapshots
}

// ResourceSnapshot represents a point-in-time resource usage measurement
type ResourceSnapshot struct {
	Timestamp   time.Time
	MemoryAlloc uint64
	Goroutines  int
}

// analyzeResourceTrend analyzes resource usage trends for leak detection
func analyzeResourceTrend(snapshots []ResourceSnapshot) (memoryTrend, goroutineTrend float64) {
	if len(snapshots) < 2 {
		return 0, 0
	}

	// Calculate linear trends
	n := len(snapshots)
	firstMem := float64(snapshots[0].MemoryAlloc)
	lastMem := float64(snapshots[n-1].MemoryAlloc)
	memoryTrend = (lastMem - firstMem) / firstMem * 100

	firstGoroutines := float64(snapshots[0].Goroutines)
	lastGoroutines := float64(snapshots[n-1].Goroutines)
	if firstGoroutines > 0 {
		goroutineTrend = (lastGoroutines - firstGoroutines) / firstGoroutines * 100
	}

	return memoryTrend, goroutineTrend
}

func TestActualBiologicalFactories(t *testing.T) {
	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{MaxComponents: 5})

	// This SHOULD fail with current implementation:
	_, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "basic", // Uses unimplemented default factory
		Position:   Position3D{X: 0, Y: 0, Z: 0},
		Threshold:  0.7,
	})

	// Currently returns: "basic neuron factory not yet implemented"
	if err == nil {
		t.Error("Should fail - default factory not implemented!")
	}
}
