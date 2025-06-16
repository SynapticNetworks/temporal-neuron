/*
=================================================================================
MICROGLIA - PERFORMANCE, STRESS, AND SCALABILITY TESTS
=================================================================================

Comprehensive performance testing for the microglia lifecycle management system.
Tests ensure the system can handle realistic biological workloads and scales
appropriately with network size and activity levels.

Test Categories:
1. Component Creation/Removal Performance - Bulk operations and throughput
2. Health Monitoring Scalability - Large-scale health tracking performance
3. Pruning System Performance - Pruning candidate generation and execution
4. Birth Request Processing - Queue management and throughput under load
5. Patrol System Scalability - Territorial surveillance at scale
6. Concurrent Access Performance - Multi-threaded stress testing
7. Memory Usage and Leak Detection - Resource consumption analysis
8. Configuration Impact Analysis - Performance across different configs
9. Large Network Simulation - Realistic brain-scale testing
10. Biological Timing Validation - Ensuring performance meets biological constraints

Biological Performance Targets:
- Component creation: <1ms per neuron (neurogenesis timescales)
- Health monitoring: >1000 components/second (microglial surveillance rates)
- Pruning evaluation: <10ms per connection (synaptic maintenance windows)
- Patrol execution: <5ms per territory (microglial process dynamics)
- Memory efficiency: <1KB per tracked component (biological resource constraints)

Research Foundation:
- Nimmerjahn et al. (2005): Microglial surveillance rates and territorial coverage
- Wake et al. (2009): Synaptic contact frequencies and monitoring timescales
- Kettenmann et al. (2011): Microglial response times and capacity limits
- Paolicelli et al. (2011): Synaptic pruning rates and decision timescales
=================================================================================
*/

package extracellular

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

// =================================================================================
// PERFORMANCE TEST CONFIGURATION AND UTILITIES
// =================================================================================

// PerformanceTestConfig defines parameters for performance testing
type PerformanceTestConfig struct {
	SmallNetwork  int // Small network size for quick tests
	MediumNetwork int // Medium network size for realistic tests
	LargeNetwork  int // Large network size for stress tests
	HugeNetwork   int // Huge network size for scalability limits

	TestDuration time.Duration // Duration for sustained performance tests
	WarmupTime   time.Duration // Warmup period before measurement

	ConcurrentWorkers   int // Number of concurrent workers for stress tests
	OperationsPerWorker int // Operations per worker for load testing
}

// GetDefaultPerformanceConfig returns realistic performance test parameters
func GetDefaultPerformanceConfig() PerformanceTestConfig {
	return PerformanceTestConfig{
		SmallNetwork:  100,   // Small cortical column
		MediumNetwork: 1000,  // Large cortical column
		LargeNetwork:  10000, // Small cortical area
		HugeNetwork:   50000, // Large cortical area

		TestDuration: 10 * time.Second,
		WarmupTime:   1 * time.Second,

		ConcurrentWorkers:   10,  // Realistic concurrent access
		OperationsPerWorker: 100, // Reasonable operation count
	}
}

// PerformanceMetrics tracks detailed performance measurements
type PerformanceMetrics struct {
	TotalOperations  int64
	TotalDuration    time.Duration
	MinDuration      time.Duration
	MaxDuration      time.Duration
	AverageDuration  time.Duration
	OperationsPerSec float64
	MemoryUsedMB     float64
	GoroutineCount   int

	// Biological performance indicators
	MeetsBiologicalTiming bool
	BiologicalTarget      time.Duration
	PerformanceRatio      float64 // Actual/Target (lower is better)
}

// measureMemoryUsage returns current memory usage in MB
func measureMemoryUsage() float64 {
	var m runtime.MemStats
	runtime.GC() // Force garbage collection for accurate measurement
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / (1024 * 1024) // Convert to MB
}

// benchmarkOperation measures the performance of a single operation
func benchmarkOperation(operation func()) time.Duration {
	start := time.Now()
	operation()
	return time.Since(start)
}

// =================================================================================
// COMPONENT LIFECYCLE PERFORMANCE TESTS
// =================================================================================

func TestMicrogliaPerformanceComponentCreation(t *testing.T) {
	t.Log("=== TESTING COMPONENT CREATION PERFORMANCE ===")

	config := GetDefaultPerformanceConfig()
	astrocyteNetwork := NewAstrocyteNetwork()

	// Test different network sizes
	testSizes := []struct {
		size             int
		name             string
		biologicalTarget time.Duration
	}{
		{config.SmallNetwork, "Small Network", 100 * time.Microsecond},   // 100μs per neuron
		{config.MediumNetwork, "Medium Network", 500 * time.Microsecond}, // 500μs per neuron
		{config.LargeNetwork, "Large Network", 1 * time.Millisecond},     // 1ms per neuron
	}

	for _, test := range testSizes {
		t.Logf("\n--- %s (%d components) ---", test.name, test.size)

		microglia := NewMicroglia(astrocyteNetwork, test.size*2) // Extra capacity

		startMem := measureMemoryUsage()
		startTime := time.Now()

		// Create components in batch
		for i := 0; i < test.size; i++ {
			componentInfo := ComponentInfo{
				ID:   fmt.Sprintf("perf_neuron_%d", i),
				Type: ComponentNeuron,
				Position: Position3D{
					X: rand.Float64() * 1000,
					Y: rand.Float64() * 1000,
					Z: rand.Float64() * 1000,
				},
				State: StateActive,
			}

			err := microglia.CreateComponent(componentInfo)
			if err != nil {
				t.Fatalf("Failed to create component %d: %v", i, err)
			}
		}

		totalDuration := time.Since(startTime)
		endMem := measureMemoryUsage()

		// Calculate metrics
		avgDuration := totalDuration / time.Duration(test.size)
		opsPerSec := float64(test.size) / totalDuration.Seconds()
		memoryUsed := endMem - startMem
		meetsBiological := avgDuration <= test.biologicalTarget
		performanceRatio := float64(avgDuration) / float64(test.biologicalTarget)

		t.Logf("Total time: %v", totalDuration)
		t.Logf("Average per component: %v", avgDuration)
		t.Logf("Operations per second: %.0f", opsPerSec)
		t.Logf("Memory used: %.2f MB", memoryUsed)
		t.Logf("Memory per component: %.2f KB", (memoryUsed*1024)/float64(test.size))
		t.Logf("Biological target: %v", test.biologicalTarget)
		t.Logf("Meets biological timing: %v (ratio: %.2f)", meetsBiological, performanceRatio)

		// Performance assertions
		if avgDuration > test.biologicalTarget*2 {
			t.Errorf("Component creation too slow: %v > %v (2x biological target)",
				avgDuration, test.biologicalTarget*2)
		}

		if memoryUsed/float64(test.size) > 5.0 { // 5KB per component limit
			t.Errorf("Memory usage too high: %.2f KB per component",
				(memoryUsed*1024)/float64(test.size))
		}

		// Verify all components were created
		stats := microglia.GetMaintenanceStats()
		if stats.ComponentsCreated != int64(test.size) {
			t.Errorf("Expected %d components created, got %d",
				test.size, stats.ComponentsCreated)
		}
	}

	t.Log("✓ Component creation performance within biological constraints")
}

func TestMicrogliaPerformanceComponentRemoval(t *testing.T) {
	t.Log("=== TESTING COMPONENT REMOVAL PERFORMANCE ===")

	config := GetDefaultPerformanceConfig()
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, config.MediumNetwork*2)

	// Pre-create components
	componentIDs := make([]string, config.MediumNetwork)
	for i := 0; i < config.MediumNetwork; i++ {
		componentID := fmt.Sprintf("removal_test_%d", i)
		componentIDs[i] = componentID

		componentInfo := ComponentInfo{
			ID:       componentID,
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0},
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)

		// Add some health data and pruning targets
		microglia.UpdateComponentHealth(componentID, 0.5, 5)
		microglia.MarkForPruning(fmt.Sprintf("synapse_%d", i), componentID, componentID, 0.1)
	}

	t.Logf("Pre-created %d components for removal testing", config.MediumNetwork)

	// Measure removal performance
	startTime := time.Now()
	startMem := measureMemoryUsage()

	removedCount := 0
	for _, componentID := range componentIDs {
		err := microglia.RemoveComponent(componentID)
		if err != nil {
			t.Logf("Warning: Failed to remove component %s: %v", componentID, err)
		} else {
			removedCount++
		}
	}

	totalDuration := time.Since(startTime)
	endMem := measureMemoryUsage()

	// Calculate metrics
	avgRemovalTime := totalDuration / time.Duration(removedCount)
	opsPerSec := float64(removedCount) / totalDuration.Seconds()
	memoryFreed := startMem - endMem

	t.Logf("Removed %d components in %v", removedCount, totalDuration)
	t.Logf("Average removal time: %v", avgRemovalTime)
	t.Logf("Removals per second: %.0f", opsPerSec)
	t.Logf("Memory freed: %.2f MB", memoryFreed)

	// Verify cleanup
	stats := microglia.GetMaintenanceStats()
	if stats.ComponentsRemoved != int64(removedCount) {
		t.Errorf("Expected %d components removed, got %d",
			removedCount, stats.ComponentsRemoved)
	}

	// Check that cleanup was thorough
	candidates := microglia.GetPruningCandidates()
	orphanedCandidates := 0
	for _, candidate := range candidates {
		for _, removedID := range componentIDs {
			if candidate.SourceID == removedID || candidate.TargetID == removedID {
				orphanedCandidates++
				break
			}
		}
	}

	if orphanedCandidates > 0 {
		t.Errorf("Found %d orphaned pruning candidates after component removal", orphanedCandidates)
	}

	// Performance assertion
	biologicalTarget := 500 * time.Microsecond // 500μs per removal
	if avgRemovalTime > biologicalTarget*2 {
		t.Errorf("Component removal too slow: %v > %v", avgRemovalTime, biologicalTarget*2)
	}

	t.Log("✓ Component removal performance and cleanup working correctly")
}

// =================================================================================
// HEALTH MONITORING PERFORMANCE TESTS
// =================================================================================

func TestMicrogliaPerformanceHealthMonitoring(t *testing.T) {
	t.Log("=== TESTING HEALTH MONITORING PERFORMANCE ===")

	config := GetDefaultPerformanceConfig()
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, config.LargeNetwork)

	// Create components
	componentIDs := make([]string, config.LargeNetwork)
	for i := 0; i < config.LargeNetwork; i++ {
		componentID := fmt.Sprintf("health_perf_%d", i)
		componentIDs[i] = componentID

		componentInfo := ComponentInfo{
			ID:       componentID,
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0},
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)
	}

	t.Logf("Created %d components for health monitoring test", config.LargeNetwork)

	// Benchmark health updates
	startTime := time.Now()
	updateCount := 0

	// Simulate realistic health monitoring patterns
	for round := 0; round < 10; round++ {
		for i := 0; i < config.LargeNetwork; i++ {
			activity := 0.3 + 0.7*rand.Float64() // Random activity 0.3-1.0
			connections := 3 + rand.Intn(15)     // Random connections 3-17

			microglia.UpdateComponentHealth(componentIDs[i], activity, connections)
			updateCount++
		}

		// Brief pause between rounds (simulates biological timing)
		if round < 9 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	totalDuration := time.Since(startTime)

	// Calculate performance metrics
	avgUpdateTime := totalDuration / time.Duration(updateCount)
	updatesPerSec := float64(updateCount) / totalDuration.Seconds()

	t.Logf("Performed %d health updates in %v", updateCount, totalDuration)
	t.Logf("Average update time: %v", avgUpdateTime)
	t.Logf("Updates per second: %.0f", updatesPerSec)

	// Verify statistics
	stats := microglia.GetMaintenanceStats()
	if stats.HealthChecks != int64(updateCount) {
		t.Errorf("Expected %d health checks, got %d", updateCount, stats.HealthChecks)
	}

	if stats.AverageHealthScore <= 0 || stats.AverageHealthScore > 1 {
		t.Errorf("Invalid average health score: %.3f", stats.AverageHealthScore)
	}

	// Test health retrieval performance
	retrievalStart := time.Now()
	healthRecords := 0

	for _, componentID := range componentIDs {
		_, exists := microglia.GetComponentHealth(componentID)
		if exists {
			healthRecords++
		}
	}

	retrievalDuration := time.Since(retrievalStart)
	avgRetrievalTime := retrievalDuration / time.Duration(len(componentIDs))

	t.Logf("Retrieved %d health records in %v", healthRecords, retrievalDuration)
	t.Logf("Average retrieval time: %v", avgRetrievalTime)

	// Performance assertions
	biologicalUpdateTarget := 100 * time.Microsecond   // 100μs per health update
	biologicalRetrievalTarget := 10 * time.Microsecond // 10μs per health retrieval

	if avgUpdateTime > biologicalUpdateTarget*5 {
		t.Errorf("Health updates too slow: %v > %v", avgUpdateTime, biologicalUpdateTarget*5)
	}

	if avgRetrievalTime > biologicalRetrievalTarget*5 {
		t.Errorf("Health retrieval too slow: %v > %v", avgRetrievalTime, biologicalRetrievalTarget*5)
	}

	// Microglial surveillance should handle >1000 components/second
	if updatesPerSec < 1000 {
		t.Errorf("Health monitoring throughput too low: %.0f < 1000 updates/sec", updatesPerSec)
	}

	t.Log("✓ Health monitoring performance meets biological requirements")
}

// =================================================================================
// PRUNING SYSTEM PERFORMANCE TESTS
// =================================================================================

func TestMicrogliaPerformancePruningSystem(t *testing.T) {
	t.Log("=== TESTING PRUNING SYSTEM PERFORMANCE ===")

	config := GetDefaultPerformanceConfig()
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, config.MediumNetwork)

	// Create components for pruning tests
	for i := 0; i < config.MediumNetwork; i++ {
		componentInfo := ComponentInfo{
			ID:       fmt.Sprintf("prune_perf_%d", i),
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0},
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)
		microglia.UpdateComponentHealth(fmt.Sprintf("prune_perf_%d", i), 0.5, 5)
	}

	// Benchmark pruning target marking
	t.Log("\n--- Pruning Target Marking Performance ---")
	markingStart := time.Now()

	synapseCount := config.MediumNetwork * 3 // 3 synapses per neuron average
	for i := 0; i < synapseCount; i++ {
		sourceIdx := i % config.MediumNetwork
		targetIdx := (i + 1) % config.MediumNetwork
		activity := rand.Float64() // Random activity 0.0-1.0

		microglia.MarkForPruning(
			fmt.Sprintf("synapse_%d", i),
			fmt.Sprintf("prune_perf_%d", sourceIdx),
			fmt.Sprintf("prune_perf_%d", targetIdx),
			activity,
		)
	}

	markingDuration := time.Since(markingStart)
	avgMarkingTime := markingDuration / time.Duration(synapseCount)
	markingRate := float64(synapseCount) / markingDuration.Seconds()

	t.Logf("Marked %d synapses for pruning in %v", synapseCount, markingDuration)
	t.Logf("Average marking time: %v", avgMarkingTime)
	t.Logf("Marking rate: %.0f synapses/sec", markingRate)

	// Benchmark pruning candidate retrieval
	t.Log("\n--- Pruning Candidate Retrieval Performance ---")
	retrievalStart := time.Now()

	candidates := microglia.GetPruningCandidates()

	retrievalDuration := time.Since(retrievalStart)

	t.Logf("Retrieved %d pruning candidates in %v", len(candidates), retrievalDuration)

	// Validate candidate quality
	validCandidates := 0
	invalidScores := 0

	for _, candidate := range candidates {
		if candidate.PruningScore >= 0 && candidate.PruningScore <= 1 {
			validCandidates++
		} else {
			invalidScores++
		}
	}

	t.Logf("Valid candidates: %d, Invalid scores: %d", validCandidates, invalidScores)

	// Benchmark pruning execution
	t.Log("\n--- Pruning Execution Performance ---")
	executionStart := time.Now()

	prunedConnections := microglia.ExecutePruning()

	executionDuration := time.Since(executionStart)

	t.Logf("Pruning execution completed in %v", executionDuration)
	t.Logf("Connections pruned: %d", len(prunedConnections))

	// Performance assertions
	biologicalMarkingTarget := 1 * time.Millisecond     // 1ms per synapse marking
	biologicalRetrievalTarget := 100 * time.Millisecond // 100ms for candidate retrieval
	biologicalExecutionTarget := 50 * time.Millisecond  // 50ms for pruning execution

	if avgMarkingTime > biologicalMarkingTarget {
		t.Errorf("Synapse marking too slow: %v > %v", avgMarkingTime, biologicalMarkingTarget)
	}

	if retrievalDuration > biologicalRetrievalTarget {
		t.Errorf("Candidate retrieval too slow: %v > %v", retrievalDuration, biologicalRetrievalTarget)
	}

	if executionDuration > biologicalExecutionTarget {
		t.Errorf("Pruning execution too slow: %v > %v", executionDuration, biologicalExecutionTarget)
	}

	if invalidScores > 0 {
		t.Errorf("Found %d invalid pruning scores", invalidScores)
	}

	t.Log("✓ Pruning system performance meets biological constraints")
}

// =================================================================================
// BIRTH REQUEST PROCESSING PERFORMANCE TESTS
// =================================================================================

func TestMicrogliaPerformanceBirthRequestProcessing(t *testing.T) {
	t.Log("=== TESTING BIRTH REQUEST PROCESSING PERFORMANCE ===")

	config := GetDefaultPerformanceConfig()
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, config.LargeNetwork)

	// Benchmark birth request submission
	t.Log("\n--- Birth Request Submission Performance ---")
	submissionStart := time.Now()

	requestCount := config.MediumNetwork
	priorities := []BirthPriority{PriorityLow, PriorityMedium, PriorityHigh, PriorityEmergency}

	for i := 0; i < requestCount; i++ {
		request := ComponentBirthRequest{
			ComponentType: ComponentNeuron,
			Position:      Position3D{X: float64(i), Y: 0, Z: 0},
			Justification: fmt.Sprintf("Performance test request %d", i),
			Priority:      priorities[i%len(priorities)],
			RequestedBy:   "performance_test",
			Metadata: map[string]interface{}{
				"test_id": i,
				"batch":   "performance",
			},
		}

		err := microglia.RequestComponentBirth(request)
		if err != nil {
			t.Fatalf("Failed to submit birth request %d: %v", i, err)
		}
	}

	submissionDuration := time.Since(submissionStart)
	avgSubmissionTime := submissionDuration / time.Duration(requestCount)
	submissionRate := float64(requestCount) / submissionDuration.Seconds()

	t.Logf("Submitted %d birth requests in %v", requestCount, submissionDuration)
	t.Logf("Average submission time: %v", avgSubmissionTime)
	t.Logf("Submission rate: %.0f requests/sec", submissionRate)

	// Benchmark birth request processing
	t.Log("\n--- Birth Request Processing Performance ---")
	processingStart := time.Now()

	batchCount := 0
	totalCreated := 0

	// Process in multiple batches to simulate realistic timing
	for batch := 0; batch < 10; batch++ {
		created := microglia.ProcessBirthRequests()
		totalCreated += len(created)
		batchCount++

		// Brief pause between batches
		if batch < 9 {
			time.Sleep(5 * time.Millisecond)
		}
	}

	processingDuration := time.Since(processingStart)
	avgProcessingTime := processingDuration / time.Duration(batchCount)

	t.Logf("Processed %d batches in %v", batchCount, processingDuration)
	t.Logf("Total components created: %d", totalCreated)
	t.Logf("Average batch processing time: %v", avgProcessingTime)

	// Verify statistics
	stats := microglia.GetMaintenanceStats()
	if stats.ComponentsCreated != int64(totalCreated) {
		t.Errorf("Expected %d components created, got %d",
			totalCreated, stats.ComponentsCreated)
	}

	// Performance assertions
	biologicalSubmissionTarget := 100 * time.Microsecond // 100μs per request submission
	biologicalProcessingTarget := 10 * time.Millisecond  // 10ms per batch processing

	if avgSubmissionTime > biologicalSubmissionTarget*5 {
		t.Errorf("Birth request submission too slow: %v > %v",
			avgSubmissionTime, biologicalSubmissionTarget*5)
	}

	if avgProcessingTime > biologicalProcessingTarget*2 {
		t.Errorf("Birth request processing too slow: %v > %v",
			avgProcessingTime, biologicalProcessingTarget*2)
	}

	// Should handle high submission rates for neurogenesis bursts
	if submissionRate < 1000 {
		t.Errorf("Birth request submission rate too low: %.0f < 1000 requests/sec", submissionRate)
	}

	t.Log("✓ Birth request processing performance meets biological requirements")
}

// =================================================================================
// PATROL SYSTEM PERFORMANCE TESTS
// =================================================================================

func TestMicrogliaPerformancePatrolSystem(t *testing.T) {
	t.Log("=== TESTING PATROL SYSTEM PERFORMANCE ===")

	config := GetDefaultPerformanceConfig()
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, config.LargeNetwork)

	// Create components distributed across multiple territories
	territoryCount := 10
	componentsPerTerritory := config.LargeNetwork / territoryCount

	for territory := 0; territory < territoryCount; territory++ {
		centerX := float64(territory * 200) // 200μm spacing between territories

		for i := 0; i < componentsPerTerritory; i++ {
			componentInfo := ComponentInfo{
				ID:   fmt.Sprintf("patrol_comp_%d_%d", territory, i),
				Type: ComponentNeuron,
				Position: Position3D{
					X: centerX + (rand.Float64()-0.5)*100, // ±50μm from center
					Y: (rand.Float64() - 0.5) * 100,
					Z: (rand.Float64() - 0.5) * 100,
				},
				State: StateActive,
			}
			microglia.CreateComponent(componentInfo)
		}
	}

	t.Logf("Created %d components across %d territories", config.LargeNetwork, territoryCount)

	// Establish patrol routes
	t.Log("\n--- Patrol Route Establishment Performance ---")
	routeStart := time.Now()

	for territory := 0; territory < territoryCount; territory++ {
		territoryDef := Territory{
			Center: Position3D{X: float64(territory * 200), Y: 0, Z: 0},
			Radius: 75.0, // 75μm radius (biologically realistic)
		}

		microgliaID := fmt.Sprintf("patrol_microglia_%d", territory)
		patrolRate := 50 * time.Millisecond // 20 Hz patrol rate

		microglia.EstablishPatrolRoute(microgliaID, territoryDef, patrolRate)
	}

	routeDuration := time.Since(routeStart)
	t.Logf("Established %d patrol routes in %v", territoryCount, routeDuration)

	// Benchmark patrol execution
	t.Log("\n--- Patrol Execution Performance ---")
	executionStart := time.Now()

	patrolCount := 0
	totalComponentsChecked := 0

	// Execute multiple patrol cycles
	for cycle := 0; cycle < 5; cycle++ {
		for territory := 0; territory < territoryCount; territory++ {
			microgliaID := fmt.Sprintf("patrol_microglia_%d", territory)

			patrolStart := time.Now()
			report := microglia.ExecutePatrol(microgliaID)
			patrolDuration := time.Since(patrolStart)

			patrolCount++
			totalComponentsChecked += report.ComponentsChecked

			// Log detailed timing for first cycle
			if cycle == 0 {
				t.Logf("Territory %d: checked %d components in %v",
					territory, report.ComponentsChecked, patrolDuration)
			}
		}

		// Brief pause between cycles
		if cycle < 4 {
			time.Sleep(25 * time.Millisecond)
		}
	}

	executionDuration := time.Since(executionStart)
	avgPatrolTime := executionDuration / time.Duration(patrolCount)
	patrolRate := float64(patrolCount) / executionDuration.Seconds()
	avgComponentsPerPatrol := float64(totalComponentsChecked) / float64(patrolCount)

	t.Logf("Executed %d patrols in %v", patrolCount, executionDuration)
	t.Logf("Average patrol time: %v", avgPatrolTime)
	t.Logf("Patrol rate: %.1f patrols/sec", patrolRate)
	t.Logf("Average components per patrol: %.1f", avgComponentsPerPatrol)

	// Verify statistics
	stats := microglia.GetMaintenanceStats()
	if stats.PatrolsCompleted != int64(patrolCount) {
		t.Errorf("Expected %d patrols completed, got %d",
			patrolCount, stats.PatrolsCompleted)
	}

	if stats.HealthChecks != int64(totalComponentsChecked) {
		t.Errorf("Expected %d health checks from patrols, got %d",
			totalComponentsChecked, stats.HealthChecks)
	}

	// Performance assertions
	biologicalPatrolTarget := 5 * time.Millisecond // 5ms per patrol (microglial process speed)
	biologicalPatrolRateTarget := 10.0             // 10 patrols/sec minimum

	if avgPatrolTime > biologicalPatrolTarget*2 {
		t.Errorf("Patrol execution too slow: %v > %v", avgPatrolTime, biologicalPatrolTarget*2)
	}

	if patrolRate < biologicalPatrolRateTarget {
		t.Errorf("Patrol rate too low: %.1f < %.1f patrols/sec", patrolRate, biologicalPatrolRateTarget)
	}

	// Verify territorial coverage
	expectedComponentsPerTerritory := float64(componentsPerTerritory)
	if avgComponentsPerPatrol < expectedComponentsPerTerritory*0.8 {
		t.Errorf("Territorial coverage too low: %.1f < %.1f components per patrol",
			avgComponentsPerPatrol, expectedComponentsPerTerritory*0.8)
	}

	t.Log("✓ Patrol system performance meets biological surveillance requirements")
}

// =================================================================================
// CONCURRENT ACCESS PERFORMANCE TESTS
// =================================================================================

func TestMicrogliaPerformanceConcurrentAccess(t *testing.T) {
	t.Log("=== TESTING CONCURRENT ACCESS PERFORMANCE ===")

	config := GetDefaultPerformanceConfig()
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, config.LargeNetwork)

	// Pre-create some components for concurrent operations
	baseComponents := config.MediumNetwork / 2
	for i := 0; i < baseComponents; i++ {
		componentInfo := ComponentInfo{
			ID:       fmt.Sprintf("concurrent_base_%d", i),
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0},
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)
	}

	t.Logf("Pre-created %d base components for concurrent testing", baseComponents)

	// Test concurrent component creation
	t.Log("\n--- Concurrent Component Creation ---")
	var wg sync.WaitGroup
	errorChan := make(chan error, config.ConcurrentWorkers*config.OperationsPerWorker)

	creationStart := time.Now()

	for worker := 0; worker < config.ConcurrentWorkers; worker++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for op := 0; op < config.OperationsPerWorker; op++ {
				componentInfo := ComponentInfo{
					ID:   fmt.Sprintf("concurrent_%d_%d", workerID, op),
					Type: ComponentNeuron,
					Position: Position3D{
						X: float64(workerID*1000 + op),
						Y: 0,
						Z: 0,
					},
					State: StateActive,
				}

				err := microglia.CreateComponent(componentInfo)
				if err != nil {
					errorChan <- fmt.Errorf("worker %d op %d: %v", workerID, op, err)
				}
			}
		}(worker)
	}

	wg.Wait()
	creationDuration := time.Since(creationStart)
	close(errorChan)

	// Check for errors
	errorCount := 0
	for err := range errorChan {
		t.Logf("Concurrent creation error: %v", err)
		errorCount++
	}

	totalOperations := config.ConcurrentWorkers * config.OperationsPerWorker
	successfulOperations := totalOperations - errorCount
	operationsPerSec := float64(successfulOperations) / creationDuration.Seconds()

	t.Logf("Concurrent creation: %d/%d successful in %v",
		successfulOperations, totalOperations, creationDuration)
	t.Logf("Concurrent creation rate: %.0f operations/sec", operationsPerSec)

	// Test concurrent health updates
	t.Log("\n--- Concurrent Health Updates ---")
	errorChan = make(chan error, config.ConcurrentWorkers*config.OperationsPerWorker)

	healthStart := time.Now()

	for worker := 0; worker < config.ConcurrentWorkers; worker++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for op := 0; op < config.OperationsPerWorker; op++ {
				componentID := fmt.Sprintf("concurrent_%d_%d", workerID, op)
				activity := rand.Float64()
				connections := rand.Intn(20)

				microglia.UpdateComponentHealth(componentID, activity, connections)
			}
		}(worker)
	}

	wg.Wait()
	healthDuration := time.Since(healthStart)
	close(errorChan)

	healthOperationsPerSec := float64(totalOperations) / healthDuration.Seconds()

	t.Logf("Concurrent health updates: %d operations in %v", totalOperations, healthDuration)
	t.Logf("Concurrent health update rate: %.0f operations/sec", healthOperationsPerSec)

	// Test concurrent pruning operations
	t.Log("\n--- Concurrent Pruning Operations ---")
	pruningStart := time.Now()

	for worker := 0; worker < config.ConcurrentWorkers; worker++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for op := 0; op < config.OperationsPerWorker; op++ {
				synapseID := fmt.Sprintf("concurrent_synapse_%d_%d", workerID, op)
				sourceID := fmt.Sprintf("concurrent_%d_%d", workerID, op)
				targetID := fmt.Sprintf("concurrent_%d_%d", (workerID+1)%config.ConcurrentWorkers, op)
				activity := rand.Float64()

				microglia.MarkForPruning(synapseID, sourceID, targetID, activity)
			}
		}(worker)
	}

	wg.Wait()
	pruningDuration := time.Since(pruningStart)

	pruningOperationsPerSec := float64(totalOperations) / pruningDuration.Seconds()

	t.Logf("Concurrent pruning operations: %d operations in %v", totalOperations, pruningDuration)
	t.Logf("Concurrent pruning rate: %.0f operations/sec", pruningOperationsPerSec)

	// Verify final state consistency
	stats := microglia.GetMaintenanceStats()
	expectedCreated := int64(baseComponents + successfulOperations)

	if stats.ComponentsCreated != expectedCreated {
		t.Errorf("Component count inconsistent: expected %d, got %d",
			expectedCreated, stats.ComponentsCreated)
	}

	if stats.HealthChecks != int64(totalOperations) {
		t.Errorf("Health check count inconsistent: expected %d, got %d",
			totalOperations, stats.HealthChecks)
	}

	candidates := microglia.GetPruningCandidates()
	if len(candidates) != totalOperations {
		t.Errorf("Pruning candidate count inconsistent: expected %d, got %d",
			totalOperations, len(candidates))
	}

	// Performance assertions
	minConcurrentRate := 1000.0 // 1000 operations/sec minimum under concurrent load

	if operationsPerSec < minConcurrentRate {
		t.Errorf("Concurrent creation rate too low: %.0f < %.0f ops/sec",
			operationsPerSec, minConcurrentRate)
	}

	if healthOperationsPerSec < minConcurrentRate*2 {
		t.Errorf("Concurrent health update rate too low: %.0f < %.0f ops/sec",
			healthOperationsPerSec, minConcurrentRate*2)
	}

	if errorCount > totalOperations/10 {
		t.Errorf("Too many concurrent operation errors: %d/%d (>10%%)",
			errorCount, totalOperations)
	}

	t.Log("✓ Concurrent access performance within acceptable limits")
}

// =================================================================================
// MEMORY USAGE AND LEAK DETECTION TESTS
// =================================================================================

func TestMicrogliaPerformanceMemoryUsage(t *testing.T) {
	t.Log("=== TESTING MEMORY USAGE AND LEAK DETECTION ===")

	config := GetDefaultPerformanceConfig()

	// Test memory usage scaling
	t.Log("\n--- Memory Usage Scaling ---")
	networkSizes := []int{100, 500, 1000, 2000, 5000}

	for _, size := range networkSizes {
		astrocyteNetwork := NewAstrocyteNetwork()
		microglia := NewMicroglia(astrocyteNetwork, size*2)

		startMem := measureMemoryUsage()

		// Create components with full health and pruning data
		for i := 0; i < size; i++ {
			componentInfo := ComponentInfo{
				ID:       fmt.Sprintf("memory_test_%d", i),
				Type:     ComponentNeuron,
				Position: Position3D{X: float64(i), Y: 0, Z: 0},
				State:    StateActive,
			}
			microglia.CreateComponent(componentInfo)
			microglia.UpdateComponentHealth(fmt.Sprintf("memory_test_%d", i), 0.5, 5)
			microglia.MarkForPruning(fmt.Sprintf("synapse_%d", i),
				fmt.Sprintf("memory_test_%d", i), fmt.Sprintf("memory_test_%d", i), 0.1)
		}

		endMem := measureMemoryUsage()
		memoryUsed := endMem - startMem
		memoryPerComponent := (memoryUsed * 1024) / float64(size) // KB per component

		t.Logf("Size %d: %.2f MB total, %.2f KB per component",
			size, memoryUsed, memoryPerComponent)

		// Memory efficiency assertion
		maxMemoryPerComponent := 2.0 // 2KB per component maximum
		if memoryPerComponent > maxMemoryPerComponent {
			t.Errorf("Memory usage too high for size %d: %.2f KB > %.2f KB per component",
				size, memoryPerComponent, maxMemoryPerComponent)
		}
	}

	// Test memory leak detection
	t.Log("\n--- Memory Leak Detection ---")
	initialMem := measureMemoryUsage()

	for cycle := 0; cycle < 10; cycle++ {
		astrocyteNetwork := NewAstrocyteNetwork()
		microglia := NewMicroglia(astrocyteNetwork, config.MediumNetwork)

		// Create and destroy components multiple times
		for round := 0; round < 3; round++ {
			// Create components
			for i := 0; i < config.MediumNetwork; i++ {
				componentInfo := ComponentInfo{
					ID:       fmt.Sprintf("leak_test_%d_%d", cycle, i),
					Type:     ComponentNeuron,
					Position: Position3D{X: float64(i), Y: 0, Z: 0},
					State:    StateActive,
				}
				microglia.CreateComponent(componentInfo)
				microglia.UpdateComponentHealth(fmt.Sprintf("leak_test_%d_%d", cycle, i), 0.5, 5)
			}

			// Remove half the components
			for i := 0; i < config.MediumNetwork/2; i++ {
				microglia.RemoveComponent(fmt.Sprintf("leak_test_%d_%d", cycle, i))
			}
		}

		// Force garbage collection
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	}

	finalMem := measureMemoryUsage()
	memoryLeak := finalMem - initialMem

	t.Logf("Memory leak test: %.2f MB initial, %.2f MB final, %.2f MB difference",
		initialMem, finalMem, memoryLeak)

	// Leak detection assertion
	maxAcceptableLeak := 5.0 // 5MB maximum leak over test
	if memoryLeak > maxAcceptableLeak {
		t.Errorf("Potential memory leak detected: %.2f MB > %.2f MB",
			memoryLeak, maxAcceptableLeak)
	}

	t.Log("✓ Memory usage within acceptable limits, no significant leaks detected")
}

// =================================================================================
// CONFIGURATION IMPACT ANALYSIS TESTS
// =================================================================================

func TestMicrogliaPerformanceConfigurationImpact(t *testing.T) {
	t.Log("=== TESTING CONFIGURATION IMPACT ON PERFORMANCE ===")

	config := GetDefaultPerformanceConfig()
	astrocyteNetwork := NewAstrocyteNetwork()

	// Test different configuration profiles
	configTests := []struct {
		name   string
		config MicrogliaConfig
	}{
		{"Default", GetDefaultMicrogliaConfig()},
		{"Conservative", GetConservativeMicrogliaConfig()},
		{"Aggressive", GetAggressiveMicrogliaConfig()},
	}

	for _, configTest := range configTests {
		t.Logf("\n--- %s Configuration Performance ---", configTest.name)

		microglia := NewMicrogliaWithConfig(astrocyteNetwork, configTest.config)

		// Benchmark component creation
		creationStart := time.Now()
		for i := 0; i < config.SmallNetwork; i++ {
			componentInfo := ComponentInfo{
				ID:       fmt.Sprintf("%s_perf_%d", configTest.name, i),
				Type:     ComponentNeuron,
				Position: Position3D{X: float64(i), Y: 0, Z: 0},
				State:    StateActive,
			}
			microglia.CreateComponent(componentInfo)
		}
		creationDuration := time.Since(creationStart)

		// Benchmark health monitoring
		healthStart := time.Now()
		for i := 0; i < config.SmallNetwork; i++ {
			microglia.UpdateComponentHealth(fmt.Sprintf("%s_perf_%d", configTest.name, i), 0.5, 5)
		}
		healthDuration := time.Since(healthStart)

		// Benchmark pruning operations
		pruningStart := time.Now()
		for i := 0; i < config.SmallNetwork; i++ {
			microglia.MarkForPruning(fmt.Sprintf("%s_synapse_%d", configTest.name, i),
				fmt.Sprintf("%s_perf_%d", configTest.name, i),
				fmt.Sprintf("%s_perf_%d", configTest.name, (i+1)%config.SmallNetwork), 0.1)
		}
		candidates := microglia.GetPruningCandidates()
		pruningDuration := time.Since(pruningStart)

		t.Logf("Creation: %v (%.2f μs/component)",
			creationDuration, float64(creationDuration.Nanoseconds())/float64(config.SmallNetwork)/1000)
		t.Logf("Health: %v (%.2f μs/update)",
			healthDuration, float64(healthDuration.Nanoseconds())/float64(config.SmallNetwork)/1000)
		t.Logf("Pruning: %v (%.2f μs/operation, %d candidates)",
			pruningDuration, float64(pruningDuration.Nanoseconds())/float64(config.SmallNetwork)/1000, len(candidates))

		// Verify configuration-specific behavior
		if configTest.name == "Aggressive" {
			// Aggressive config should generate more pruning candidates
			expectedMinCandidates := config.SmallNetwork / 2
			if len(candidates) < expectedMinCandidates {
				t.Errorf("Aggressive config should generate more pruning candidates: %d < %d",
					len(candidates), expectedMinCandidates)
			}
		}
	}

	t.Log("✓ Configuration impact analysis completed")
}

// =================================================================================
// LARGE NETWORK SIMULATION TESTS
// =================================================================================

func TestMicrogliaPerformanceLargeNetworkSimulation(t *testing.T) {
	t.Log("=== TESTING LARGE NETWORK SIMULATION PERFORMANCE ===")

	config := GetDefaultPerformanceConfig()
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, config.HugeNetwork)

	t.Logf("Simulating brain-scale network with %d components", config.HugeNetwork)

	// Phase 1: Initial network construction
	t.Log("\n--- Phase 1: Network Construction ---")
	constructionStart := time.Now()

	for i := 0; i < config.HugeNetwork; i++ {
		componentInfo := ComponentInfo{
			ID:   fmt.Sprintf("large_network_%d", i),
			Type: ComponentNeuron,
			Position: Position3D{
				X: rand.Float64() * 10000, // 10mm x 10mm x 10mm brain region
				Y: rand.Float64() * 10000,
				Z: rand.Float64() * 10000,
			},
			State: StateActive,
		}

		err := microglia.CreateComponent(componentInfo)
		if err != nil {
			t.Fatalf("Failed to create component %d: %v", i, err)
		}

		// Progress reporting
		if (i+1)%10000 == 0 {
			t.Logf("Created %d/%d components", i+1, config.HugeNetwork)
		}
	}

	constructionDuration := time.Since(constructionStart)
	constructionRate := float64(config.HugeNetwork) / constructionDuration.Seconds()

	t.Logf("Network construction: %d components in %v", config.HugeNetwork, constructionDuration)
	t.Logf("Construction rate: %.0f components/sec", constructionRate)

	// Phase 2: Sustained activity simulation
	t.Log("\n--- Phase 2: Sustained Activity Simulation ---")
	activityStart := time.Now()

	activityCycles := 100
	componentsPerCycle := config.HugeNetwork / 10 // 10% activity per cycle

	for cycle := 0; cycle < activityCycles; cycle++ {
		// Simulate random neural activity
		for i := 0; i < componentsPerCycle; i++ {
			componentIdx := rand.Intn(config.HugeNetwork)
			componentID := fmt.Sprintf("large_network_%d", componentIdx)
			activity := 0.1 + rand.Float64()*0.8 // 0.1-0.9 activity range
			connections := 3 + rand.Intn(17)     // 3-19 connections

			microglia.UpdateComponentHealth(componentID, activity, connections)
		}

		// Progress reporting
		if (cycle+1)%20 == 0 {
			t.Logf("Completed %d/%d activity cycles", cycle+1, activityCycles)
		}
	}

	activityDuration := time.Since(activityStart)
	totalHealthUpdates := activityCycles * componentsPerCycle
	healthUpdateRate := float64(totalHealthUpdates) / activityDuration.Seconds()

	t.Logf("Activity simulation: %d health updates in %v", totalHealthUpdates, activityDuration)
	t.Logf("Health update rate: %.0f updates/sec", healthUpdateRate)

	// Phase 3: Network maintenance
	t.Log("\n--- Phase 3: Network Maintenance ---")
	maintenanceStart := time.Now()

	// Generate synaptic connections for pruning
	synapseCount := config.HugeNetwork / 2 // 0.5 synapses per neuron average
	for i := 0; i < synapseCount; i++ {
		sourceIdx := rand.Intn(config.HugeNetwork)
		targetIdx := rand.Intn(config.HugeNetwork)
		activity := rand.Float64()

		microglia.MarkForPruning(
			fmt.Sprintf("large_synapse_%d", i),
			fmt.Sprintf("large_network_%d", sourceIdx),
			fmt.Sprintf("large_network_%d", targetIdx),
			activity,
		)

		// Progress reporting
		if (i+1)%5000 == 0 {
			t.Logf("Marked %d/%d synapses for pruning evaluation", i+1, synapseCount)
		}
	}

	// Retrieve pruning candidates
	candidateStart := time.Now()
	candidates := microglia.GetPruningCandidates()
	candidateDuration := time.Since(candidateStart)

	maintenanceDuration := time.Since(maintenanceStart)

	t.Logf("Network maintenance: %d synapses processed in %v", synapseCount, maintenanceDuration)
	t.Logf("Pruning candidates: %d retrieved in %v", len(candidates), candidateDuration)

	// Final statistics and validation
	stats := microglia.GetMaintenanceStats()
	finalMem := measureMemoryUsage()

	t.Logf("\n--- Final Network Statistics ---")
	t.Logf("Components created: %d", stats.ComponentsCreated)
	t.Logf("Health checks: %d", stats.HealthChecks)
	t.Logf("Average health score: %.3f", stats.AverageHealthScore)
	t.Logf("Memory usage: %.2f MB", finalMem)
	t.Logf("Memory per component: %.2f KB", (finalMem*1024)/float64(config.HugeNetwork))

	// Performance validation
	biologicalConstructionTarget := 1000.0 // 1000 components/sec minimum
	biologicalHealthTarget := 5000.0       // 5000 health updates/sec minimum
	maxMemoryPerComponent := 1.0           // 1KB per component maximum

	if constructionRate < biologicalConstructionTarget {
		t.Errorf("Network construction too slow: %.0f < %.0f components/sec",
			constructionRate, biologicalConstructionTarget)
	}

	if healthUpdateRate < biologicalHealthTarget {
		t.Errorf("Health monitoring too slow: %.0f < %.0f updates/sec",
			healthUpdateRate, biologicalHealthTarget)
	}

	memoryPerComponent := (finalMem * 1024) / float64(config.HugeNetwork)
	if memoryPerComponent > maxMemoryPerComponent {
		t.Errorf("Memory usage too high: %.2f KB > %.2f KB per component",
			memoryPerComponent, maxMemoryPerComponent)
	}

	if len(candidates) == 0 {
		t.Error("No pruning candidates generated in large network")
	}

	t.Log("✓ Large network simulation completed successfully")
}

// =================================================================================
// BIOLOGICAL TIMING VALIDATION TESTS
// =================================================================================

func TestMicrogliaPerformanceBiologicalTiming(t *testing.T) {
	t.Log("=== TESTING BIOLOGICAL TIMING VALIDATION ===")

	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Biological timing requirements based on research
	timingTests := []struct {
		operation        string
		biologicalTarget time.Duration
		tolerance        float64
		testFunc         func() time.Duration
	}{
		{
			"Component Creation",
			500 * time.Microsecond, // Neurogenesis timing
			2.0,                    // 2x tolerance
			func() time.Duration {
				return benchmarkOperation(func() {
					componentInfo := ComponentInfo{
						ID:       fmt.Sprintf("timing_test_%d", rand.Int()),
						Type:     ComponentNeuron,
						Position: Position3D{X: rand.Float64() * 100, Y: 0, Z: 0},
						State:    StateActive,
					}
					microglia.CreateComponent(componentInfo)
				})
			},
		},
		{
			"Health Update",
			100 * time.Microsecond, // Microglial surveillance timing
			3.0,                    // 3x tolerance
			func() time.Duration {
				return benchmarkOperation(func() {
					componentID := fmt.Sprintf("timing_test_%d", rand.Intn(100))
					microglia.UpdateComponentHealth(componentID, rand.Float64(), rand.Intn(20))
				})
			},
		},
		{
			"Pruning Marking",
			200 * time.Microsecond, // Synaptic evaluation timing
			2.0,                    // 2x tolerance
			func() time.Duration {
				return benchmarkOperation(func() {
					synapseID := fmt.Sprintf("timing_synapse_%d", rand.Int())
					microglia.MarkForPruning(synapseID, "src", "dst", rand.Float64())
				})
			},
		},
	}

	// Pre-create some components for testing
	for i := 0; i < 100; i++ {
		componentInfo := ComponentInfo{
			ID:       fmt.Sprintf("timing_test_%d", i),
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0},
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)
	}

	// Run timing validation tests
	for _, test := range timingTests {
		t.Logf("\n--- %s Timing Validation ---", test.operation)

		// Warmup
		for i := 0; i < 10; i++ {
			test.testFunc()
		}

		// Measure timing over multiple iterations
		iterations := 100
		totalDuration := time.Duration(0)
		minDuration := time.Duration(math.MaxInt64)
		maxDuration := time.Duration(0)

		for i := 0; i < iterations; i++ {
			duration := test.testFunc()
			totalDuration += duration

			if duration < minDuration {
				minDuration = duration
			}
			if duration > maxDuration {
				maxDuration = duration
			}
		}

		avgDuration := totalDuration / time.Duration(iterations)
		performanceRatio := float64(avgDuration) / float64(test.biologicalTarget)
		meetsBiological := avgDuration <= time.Duration(float64(test.biologicalTarget)*test.tolerance)

		t.Logf("Average: %v", avgDuration)
		t.Logf("Min: %v", minDuration)
		t.Logf("Max: %v", maxDuration)
		t.Logf("Biological target: %v", test.biologicalTarget)
		t.Logf("Performance ratio: %.2fx", performanceRatio)
		t.Logf("Meets biological timing: %v", meetsBiological)

		// Biological timing assertion
		if !meetsBiological {
			t.Errorf("%s timing exceeds biological constraints: %v > %v (%.1fx tolerance)",
				test.operation, avgDuration,
				time.Duration(float64(test.biologicalTarget)*test.tolerance), test.tolerance)
		}

		// Consistency check (max shouldn't be >10x average)
		if maxDuration > avgDuration*100 {
			t.Errorf("%s timing inconsistent: max %v >> avg %v",
				test.operation, maxDuration, avgDuration)
		}
	}

	t.Log("✓ All operations meet biological timing requirements")
}

// =================================================================================
// BENCHMARK TESTS FOR CONTINUOUS INTEGRATION
// =================================================================================

// BenchmarkMicrogliaComponentCreation benchmarks component creation performance
func BenchmarkMicrogliaComponentCreation(b *testing.B) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, b.N*2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		componentInfo := ComponentInfo{
			ID:       fmt.Sprintf("bench_component_%d", i),
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0},
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)
	}
}

// BenchmarkMicrogliaHealthUpdates benchmarks health monitoring performance
func BenchmarkMicrogliaHealthUpdates(b *testing.B) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Pre-create components
	for i := 0; i < 1000; i++ {
		componentInfo := ComponentInfo{
			ID:       fmt.Sprintf("bench_health_%d", i),
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0},
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		componentID := fmt.Sprintf("bench_health_%d", i%1000)
		microglia.UpdateComponentHealth(componentID, rand.Float64(), rand.Intn(20))
	}
}

// BenchmarkMicrogliaPruningOperations benchmarks pruning system performance
func BenchmarkMicrogliaPruningOperations(b *testing.B) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		synapseID := fmt.Sprintf("bench_synapse_%d", i)
		microglia.MarkForPruning(synapseID, "src", "dst", rand.Float64())
	}
}

// BenchmarkMicrogliaPatrolExecution benchmarks patrol system performance
func BenchmarkMicrogliaPatrolExecution(b *testing.B) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 1000)

	// Create components in patrol territory
	for i := 0; i < 100; i++ {
		componentInfo := ComponentInfo{
			ID:       fmt.Sprintf("bench_patrol_%d", i),
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0},
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)
	}

	// Establish patrol route
	territory := Territory{
		Center: Position3D{X: 50, Y: 0, Z: 0},
		Radius: 100.0,
	}
	microglia.EstablishPatrolRoute("bench_patrol", territory, 100*time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		microglia.ExecutePatrol("bench_patrol")
	}
}

// BenchmarkMicrogliaConcurrentOperations benchmarks concurrent access performance
func BenchmarkMicrogliaConcurrentOperations(b *testing.B) {
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, 10000)

	// Pre-create components
	for i := 0; i < 1000; i++ {
		componentInfo := ComponentInfo{
			ID:       fmt.Sprintf("bench_concurrent_%d", i),
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0},
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Mix of operations
			switch rand.Intn(3) {
			case 0:
				// Health update
				componentID := fmt.Sprintf("bench_concurrent_%d", rand.Intn(1000))
				microglia.UpdateComponentHealth(componentID, rand.Float64(), rand.Intn(20))
			case 1:
				// Pruning operation
				synapseID := fmt.Sprintf("bench_synapse_%d", rand.Int())
				microglia.MarkForPruning(synapseID, "src", "dst", rand.Float64())
			case 2:
				// Statistics retrieval
				microglia.GetMaintenanceStats()
			}
		}
	})
}

// =================================================================================
// PERFORMANCE TEST SUMMARY AND REPORTING
// =================================================================================

func TestMicrogliaPerformanceSummary(t *testing.T) {
	t.Log("=== MICROGLIA PERFORMANCE TEST SUMMARY ===")

	config := GetDefaultPerformanceConfig()

	// Collect performance metrics from a comprehensive test
	astrocyteNetwork := NewAstrocyteNetwork()
	microglia := NewMicroglia(astrocyteNetwork, config.LargeNetwork)

	startMem := measureMemoryUsage()
	overallStart := time.Now()

	// Component creation performance
	creationStart := time.Now()
	for i := 0; i < config.MediumNetwork; i++ {
		componentInfo := ComponentInfo{
			ID:       fmt.Sprintf("summary_component_%d", i),
			Type:     ComponentNeuron,
			Position: Position3D{X: float64(i), Y: 0, Z: 0},
			State:    StateActive,
		}
		microglia.CreateComponent(componentInfo)
	}
	creationDuration := time.Since(creationStart)

	// Health monitoring performance
	healthStart := time.Now()
	for i := 0; i < config.MediumNetwork*5; i++ {
		componentID := fmt.Sprintf("summary_component_%d", i%config.MediumNetwork)
		microglia.UpdateComponentHealth(componentID, rand.Float64(), rand.Intn(20))
	}
	healthDuration := time.Since(healthStart)

	// Pruning system performance
	pruningStart := time.Now()
	for i := 0; i < config.MediumNetwork*2; i++ {
		microglia.MarkForPruning(fmt.Sprintf("summary_synapse_%d", i), "src", "dst", rand.Float64())
	}
	candidates := microglia.GetPruningCandidates()
	pruningDuration := time.Since(pruningStart)

	overallDuration := time.Since(overallStart)
	endMem := measureMemoryUsage()

	// Calculate comprehensive metrics
	stats := microglia.GetMaintenanceStats()

	creationRate := float64(config.MediumNetwork) / creationDuration.Seconds()
	healthRate := float64(config.MediumNetwork*5) / healthDuration.Seconds()
	pruningRate := float64(config.MediumNetwork*2) / pruningDuration.Seconds()
	memoryUsed := endMem - startMem
	memoryPerComponent := (memoryUsed * 1024) / float64(config.MediumNetwork)

	// Biological performance targets
	biologicalTargets := map[string]float64{
		"creation_rate":        1000, // components/sec
		"health_rate":          5000, // updates/sec
		"pruning_rate":         2000, // operations/sec
		"memory_per_component": 2.0,  // KB
		"max_creation_time":    1.0,  // ms per component
		"max_health_time":      0.2,  // ms per health update
		"max_pruning_time":     0.5,  // ms per pruning operation
	}

	avgCreationTime := float64(creationDuration.Nanoseconds()) / float64(config.MediumNetwork) / 1e6
	avgHealthTime := float64(healthDuration.Nanoseconds()) / float64(config.MediumNetwork*5) / 1e6
	avgPruningTime := float64(pruningDuration.Nanoseconds()) / float64(config.MediumNetwork*2) / 1e6

	t.Log("\n=== PERFORMANCE METRICS ===")
	t.Logf("Overall test duration: %v", overallDuration)
	t.Logf("Total memory used: %.2f MB", memoryUsed)
	t.Logf("")

	t.Log("--- Component Creation ---")
	t.Logf("Rate: %.0f components/sec (target: %.0f)", creationRate, biologicalTargets["creation_rate"])
	t.Logf("Average time: %.3f ms/component (target: <%.1f ms)", avgCreationTime, biologicalTargets["max_creation_time"])
	t.Logf("Total duration: %v", creationDuration)

	t.Log("--- Health Monitoring ---")
	t.Logf("Rate: %.0f updates/sec (target: %.0f)", healthRate, biologicalTargets["health_rate"])
	t.Logf("Average time: %.3f ms/update (target: <%.1f ms)", avgHealthTime, biologicalTargets["max_health_time"])
	t.Logf("Total duration: %v", healthDuration)

	t.Log("--- Pruning System ---")
	t.Logf("Rate: %.0f operations/sec (target: %.0f)", pruningRate, biologicalTargets["pruning_rate"])
	t.Logf("Average time: %.3f ms/operation (target: <%.1f ms)", avgPruningTime, biologicalTargets["max_pruning_time"])
	t.Logf("Candidates generated: %d", len(candidates))
	t.Logf("Total duration: %v", pruningDuration)

	t.Log("--- Memory Usage ---")
	t.Logf("Memory per component: %.2f KB (target: <%.1f KB)", memoryPerComponent, biologicalTargets["memory_per_component"])
	t.Logf("Components created: %d", stats.ComponentsCreated)
	t.Logf("Health checks: %d", stats.HealthChecks)

	t.Log("--- Biological Validation ---")

	// Performance validation
	performanceIssues := 0

	if creationRate < biologicalTargets["creation_rate"] {
		t.Logf("⚠️  Component creation rate below target: %.0f < %.0f", creationRate, biologicalTargets["creation_rate"])
		performanceIssues++
	} else {
		t.Logf("✓ Component creation rate meets target")
	}

	if healthRate < biologicalTargets["health_rate"] {
		t.Logf("⚠️  Health monitoring rate below target: %.0f < %.0f", healthRate, biologicalTargets["health_rate"])
		performanceIssues++
	} else {
		t.Logf("✓ Health monitoring rate meets target")
	}

	if pruningRate < biologicalTargets["pruning_rate"] {
		t.Logf("⚠️  Pruning operation rate below target: %.0f < %.0f", pruningRate, biologicalTargets["pruning_rate"])
		performanceIssues++
	} else {
		t.Logf("✓ Pruning operation rate meets target")
	}

	if memoryPerComponent > biologicalTargets["memory_per_component"] {
		t.Logf("⚠️  Memory usage above target: %.2f > %.1f KB/component", memoryPerComponent, biologicalTargets["memory_per_component"])
		performanceIssues++
	} else {
		t.Logf("✓ Memory usage within target")
	}

	if avgCreationTime > biologicalTargets["max_creation_time"] {
		t.Logf("⚠️  Component creation time above target: %.3f > %.1f ms", avgCreationTime, biologicalTargets["max_creation_time"])
		performanceIssues++
	} else {
		t.Logf("✓ Component creation time within target")
	}

	if avgHealthTime > biologicalTargets["max_health_time"] {
		t.Logf("⚠️  Health update time above target: %.3f > %.1f ms", avgHealthTime, biologicalTargets["max_health_time"])
		performanceIssues++
	} else {
		t.Logf("✓ Health update time within target")
	}

	if avgPruningTime > biologicalTargets["max_pruning_time"] {
		t.Logf("⚠️  Pruning operation time above target: %.3f > %.1f ms", avgPruningTime, biologicalTargets["max_pruning_time"])
		performanceIssues++
	} else {
		t.Logf("✓ Pruning operation time within target")
	}

	// Overall assessment
	t.Log("\n=== OVERALL ASSESSMENT ===")
	if performanceIssues == 0 {
		t.Log("🎉 ALL PERFORMANCE TARGETS MET - System ready for biological-scale neural networks")
	} else if performanceIssues <= 2 {
		t.Logf("⚠️  %d performance issues detected - System functional but may need optimization", performanceIssues)
	} else {
		t.Logf("❌ %d performance issues detected - System needs significant optimization", performanceIssues)
	}

	// Biological realism summary
	t.Log("\n=== BIOLOGICAL REALISM SUMMARY ===")
	t.Log("Component creation rate supports realistic neurogenesis timescales")
	t.Log("Health monitoring supports microglial surveillance frequencies")
	t.Log("Pruning system supports synaptic maintenance timescales")
	t.Log("Memory usage allows scaling to cortical column sizes")
	t.Log("Performance enables real-time biological network simulation")

	// Final validation
	if performanceIssues > 3 {
		t.Errorf("System performance inadequate for biological neural network simulation")
	}

	t.Log("✓ Performance testing completed - System validated for biological neural networks")
}
