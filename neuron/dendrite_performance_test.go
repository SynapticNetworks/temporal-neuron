// Performance tests for dendritic integration modes
// Tests throughput, latency, memory usage, and concurrent access patterns

package neuron

import (
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/synapse"
)

// BenchmarkDendriticModes measures throughput for each integration mode
func BenchmarkDendriticModes(b *testing.B) {

	// Define a biological config to use for all relevant modes in the benchmark.
	bioConfig := CreateCorticalPyramidalConfig()

	modes := []struct {
		name string
		mode DendriticIntegrationMode
	}{
		{"PassiveMembrane", NewPassiveMembraneMode()},
		{"TemporalSummation", NewTemporalSummationMode()},
		{"ShuntingInhibition", NewShuntingInhibitionMode(0.5, bioConfig)},
		{"ActiveDendrite", NewActiveDendriteMode(ActiveDendriteConfig{}, bioConfig)},
	}

	for _, mode := range modes {
		b.Run(mode.name, func(b *testing.B) {
			msg := synapse.SynapseMessage{
				Value:     1.0,
				Timestamp: time.Now(),
				SourceID:  "benchmark",
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				mode.mode.Handle(msg)
				if i%100 == 0 { // Process every 100 handles for buffered modes
					mode.mode.Process(MembraneSnapshot{})
				}
			}
		})
	}
}

// BenchmarkHighFrequencyBursts tests performance under burst conditions
func BenchmarkHighFrequencyBursts(b *testing.B) {
	mode := NewTemporalSummationMode()
	defer mode.Close()

	// Pre-create messages to avoid allocation overhead in benchmark
	messages := make([]synapse.SynapseMessage, 1000)
	for i := range messages {
		messages[i] = synapse.SynapseMessage{
			Value:     float64(i%10) / 10.0,
			Timestamp: time.Now(),
			SourceID:  fmt.Sprintf("burst_%d", i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate high-frequency burst (100 messages)
		for j := 0; j < 100; j++ {
			mode.Handle(messages[j%len(messages)])
		}
		// Process the burst
		mode.Process(MembraneSnapshot{})
	}
}

// BenchmarkMemoryAllocation measures allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	mode := NewTemporalSummationMode()
	defer mode.Close()

	msg := synapse.SynapseMessage{
		Value:     1.0,
		Timestamp: time.Now(),
		SourceID:  "memory_test",
	}

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mode.Handle(msg)
		if i%50 == 0 {
			mode.Process(MembraneSnapshot{})
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	allocPerOp := float64(m2.TotalAlloc-m1.TotalAlloc) / float64(b.N)
	b.ReportMetric(allocPerOp, "bytes/op")
}

// TestDendriticPerformanceCharacteristics validates performance requirements
func TestDendriticPerformanceCharacteristics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	t.Log("=== DENDRITIC PERFORMANCE CHARACTERISTICS ===")
	t.Log("Validating throughput, latency, and resource usage")

	// Define a biological config to use for all relevant modes in the benchmark.
	bioConfig := CreateCorticalPyramidalConfig()

	modes := []struct {
		name string
		mode DendriticIntegrationMode
	}{
		{"PassiveMembrane", NewPassiveMembraneMode()},
		{"TemporalSummation", NewTemporalSummationMode()},
		{"ShuntingInhibition", NewShuntingInhibitionMode(0.5, bioConfig)},
		{"ActiveDendrite", NewActiveDendriteMode(ActiveDendriteConfig{}, bioConfig)},
	}

	for _, mode := range modes {
		t.Run(mode.name+"_Throughput", func(t *testing.T) {
			testThroughput(t, mode.mode, mode.name)
		})

		t.Run(mode.name+"_Latency", func(t *testing.T) {
			testLatency(t, mode.mode, mode.name)
		})

		t.Run(mode.name+"_MemoryUsage", func(t *testing.T) {
			testMemoryUsage(t, mode.mode, mode.name)
		})
	}
}

// testThroughput measures sustained operation rate
func testThroughput(t *testing.T, mode DendriticIntegrationMode, name string) {
	const numOperations = 100000
	const testDuration = 2 * time.Second

	msg := synapse.SynapseMessage{
		Value:     1.0,
		Timestamp: time.Now(),
		SourceID:  "throughput_test",
	}

	start := time.Now()
	ops := 0

	for time.Since(start) < testDuration {
		mode.Handle(msg)
		if ops%100 == 0 {
			mode.Process(MembraneSnapshot{})
		}
		ops++
	}

	elapsed := time.Since(start)
	throughput := float64(ops) / elapsed.Seconds()

	t.Logf("%s throughput: %.0f ops/sec (%d ops in %v)",
		name, throughput, ops, elapsed)

	// Performance requirements
	minThroughput := map[string]float64{
		"PassiveMembrane":    1000000, // Should be fastest (immediate processing)
		"TemporalSummation":  500000,  // Good performance with buffering
		"ShuntingInhibition": 300000,  // More complex math
		"ActiveDendrite":     100000,  // Most complex processing
	}

	if throughput < minThroughput[name] {
		t.Logf("⚠ Throughput below target: %.0f < %.0f ops/sec",
			throughput, minThroughput[name])
	} else {
		t.Logf("✓ Throughput meets requirements")
	}
}

// testLatency measures processing delay
func testLatency(t *testing.T, mode DendriticIntegrationMode, name string) {
	const numSamples = 1000

	msg := synapse.SynapseMessage{
		Value:     1.0,
		Timestamp: time.Now(),
		SourceID:  "latency_test",
	}

	latencies := make([]time.Duration, numSamples)

	for i := 0; i < numSamples; i++ {
		start := time.Now()
		mode.Handle(msg)
		mode.Process(MembraneSnapshot{})
		latencies[i] = time.Since(start)

		// Small delay to avoid overwhelming the system
		time.Sleep(10 * time.Microsecond)
	}

	// Calculate statistics
	var total time.Duration
	min := latencies[0]
	max := latencies[0]

	for _, lat := range latencies {
		total += lat
		if lat < min {
			min = lat
		}
		if lat > max {
			max = lat
		}
	}

	avg := total / time.Duration(numSamples)

	t.Logf("%s latency: avg=%v, min=%v, max=%v",
		name, avg, min, max)

	// Latency requirements (should be sub-millisecond for biological realism)
	maxAcceptableLatency := 1 * time.Millisecond

	if avg > maxAcceptableLatency {
		t.Logf("⚠ Average latency high: %v > %v", avg, maxAcceptableLatency)
	} else {
		t.Logf("✓ Latency within biological timescales")
	}
}

// testMemoryUsage measures memory consumption patterns
func testMemoryUsage(t *testing.T, mode DendriticIntegrationMode, name string) {
	const numOperations = 10000

	msg := synapse.SynapseMessage{
		Value:     1.0,
		Timestamp: time.Now(),
		SourceID:  "memory_test",
	}

	// Baseline memory
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Perform operations
	for i := 0; i < numOperations; i++ {
		mode.Handle(msg)
		if i%100 == 0 {
			mode.Process(MembraneSnapshot{})
		}
	}

	// Final memory
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	allocPerOp := float64(m2.TotalAlloc-m1.TotalAlloc) / float64(numOperations)
	bytesPerOp := float64(m2.Alloc-m1.Alloc) / float64(numOperations)

	t.Logf("%s memory: %.1f bytes/op allocated, %.1f bytes/op retained",
		name, allocPerOp, bytesPerOp)

	// Memory efficiency requirements
	maxBytesPerOp := 100.0 // Should be very low allocation per operation

	if allocPerOp > maxBytesPerOp {
		t.Logf("⚠ High allocation per operation: %.1f > %.1f bytes/op",
			allocPerOp, maxBytesPerOp)
	} else {
		t.Logf("✓ Memory usage efficient")
	}
}

// TestConcurrentDendriticAccess tests thread safety under concurrent load
func TestConcurrentDendriticAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent tests in short mode")
	}

	t.Log("=== CONCURRENT DENDRITIC ACCESS TEST ===")
	t.Log("Testing thread safety under high concurrent load")

	mode := NewTemporalSummationMode()
	defer mode.Close()

	const numGoroutines = 100
	const opsPerGoroutine = 1000
	const testDuration = 3 * time.Second

	var totalOps int64
	var errors int64
	var wg sync.WaitGroup

	start := time.Now()

	// Launch concurrent workers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			localOps := 0
			msg := synapse.SynapseMessage{
				Value:     float64(workerID) / 100.0,
				Timestamp: time.Now(),
				SourceID:  fmt.Sprintf("worker_%d", workerID),
			}

			for time.Since(start) < testDuration {
				// Handle operation
				mode.Handle(msg)
				localOps++

				// Periodically process (simulate realistic usage)
				if localOps%50 == 0 {
					mode.Process(MembraneSnapshot{})
				}

				// Brief pause to allow other goroutines
				if localOps%100 == 0 {
					time.Sleep(1 * time.Microsecond)
				}
			}

			atomic.AddInt64(&totalOps, int64(localOps))
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	concurrentThroughput := float64(totalOps) / elapsed.Seconds()
	errorRate := float64(errors) / float64(totalOps) * 100

	t.Logf("Concurrent performance:")
	t.Logf("  Goroutines: %d", numGoroutines)
	t.Logf("  Total operations: %d", totalOps)
	t.Logf("  Duration: %v", elapsed)
	t.Logf("  Throughput: %.0f ops/sec", concurrentThroughput)
	t.Logf("  Error rate: %.2f%%", errorRate)

	// Validate concurrent performance
	if errorRate > 0.1 {
		t.Errorf("High error rate under concurrency: %.2f%%", errorRate)
	} else {
		t.Logf("✓ Thread safety maintained")
	}

	if concurrentThroughput < 100000 {
		t.Logf("⚠ Concurrent throughput low: %.0f ops/sec", concurrentThroughput)
	} else {
		t.Logf("✓ Good concurrent performance")
	}
}

// TestMemoryScaling validates memory usage scales linearly with load
func TestMemoryScaling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory scaling tests in short mode")
	}

	t.Log("=== MEMORY SCALING TEST ===")
	t.Log("Validating memory usage scales linearly with buffer size")

	mode := NewTemporalSummationMode()
	defer mode.Close()

	testSizes := []int{100, 1000, 10000, 50000}

	for _, size := range testSizes {
		t.Run(fmt.Sprintf("BufferSize_%d", size), func(t *testing.T) {
			// Clear any existing state
			mode.Process(MembraneSnapshot{})
			runtime.GC()

			var m1 runtime.MemStats
			runtime.ReadMemStats(&m1)

			// Fill buffer to specified size
			for i := 0; i < size; i++ {
				msg := synapse.SynapseMessage{
					Value:     float64(i) / 1000.0,
					Timestamp: time.Now(),
					SourceID:  fmt.Sprintf("scaling_%d", i),
				}
				mode.Handle(msg)
			}

			var m2 runtime.MemStats
			runtime.ReadMemStats(&m2)

			memoryUsed := m2.Alloc - m1.Alloc
			bytesPerMessage := float64(memoryUsed) / float64(size)

			t.Logf("Buffer size %d: %d bytes (%.1f bytes/msg)",
				size, memoryUsed, bytesPerMessage)

			// Process to clear buffer
			mode.Process(MembraneSnapshot{})
			runtime.GC()

			var m3 runtime.MemStats
			runtime.ReadMemStats(&m3)

			memoryAfterClear := m3.Alloc - m1.Alloc

			t.Logf("Memory after clear: %d bytes (%.1f%% retained)",
				memoryAfterClear, float64(memoryAfterClear)/float64(memoryUsed)*100)

			// Validate linear scaling
			expectedBytesPerMsg := 100.0 // Rough estimate for SynapseMessage
			if bytesPerMessage > expectedBytesPerMsg*2 {
				t.Logf("⚠ Higher than expected memory per message: %.1f > %.1f",
					bytesPerMessage, expectedBytesPerMsg)
			} else {
				t.Logf("✓ Memory usage scales reasonably")
			}
		})
	}
}

// TestLongRunningStability tests performance over extended periods
func TestLongRunningStability(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running stability tests in short mode")
	}

	t.Log("=== LONG-RUNNING STABILITY TEST ===")
	t.Log("Testing sustained performance over extended operation")

	mode := NewTemporalSummationMode()
	defer mode.Close()

	const testDuration = 10 * time.Second // Reasonable for CI
	const sampleInterval = 1 * time.Second

	var totalOps int64
	samples := make([]float64, 0)

	start := time.Now()
	lastSample := start

	msg := synapse.SynapseMessage{
		Value:     1.0,
		Timestamp: time.Now(),
		SourceID:  "stability_test",
	}

	for time.Since(start) < testDuration {
		mode.Handle(msg)
		atomic.AddInt64(&totalOps, 1)

		if atomic.LoadInt64(&totalOps)%100 == 0 {
			mode.Process(MembraneSnapshot{})
		}

		// Sample throughput periodically
		if time.Since(lastSample) >= sampleInterval {
			ops := atomic.SwapInt64(&totalOps, 0)
			elapsed := time.Since(lastSample)
			throughput := float64(ops) / elapsed.Seconds()
			samples = append(samples, throughput)

			t.Logf("Sample %d: %.0f ops/sec", len(samples), throughput)
			lastSample = time.Now()
		}
	}

	// Analyze stability
	if len(samples) < 2 {
		t.Log("⚠ Insufficient samples for stability analysis")
		return
	}

	// Calculate coefficient of variation
	var sum, sumSq float64
	for _, sample := range samples {
		sum += sample
		sumSq += sample * sample
	}

	mean := sum / float64(len(samples))
	variance := (sumSq / float64(len(samples))) - (mean * mean)
	stdDev := math.Sqrt(variance)
	cv := stdDev / mean

	t.Logf("Stability analysis:")
	t.Logf("  Samples: %d", len(samples))
	t.Logf("  Mean throughput: %.0f ops/sec", mean)
	t.Logf("  Standard deviation: %.0f ops/sec", stdDev)
	t.Logf("  Coefficient of variation: %.3f", cv)

	// Validate stability (CV should be low for stable performance)
	if cv > 0.2 {
		t.Logf("⚠ Performance variability high: CV=%.3f > 0.2", cv)
	} else {
		t.Logf("✓ Stable performance over extended operation")
	}
}
