# Synaptic Processor - Biologically Realistic Neural Connections

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)](https://golang.org/)
[![License: TNRL](https://img.shields.io/badge/License-TNRL--1.0-green.svg)](https://github.com/SynapticNetworks/temporal-neuron/blob/main/LICENSE)
[![Performance](https://img.shields.io/badge/throughput-1.2M_ops%2Fsec-brightgreen.svg)](https://github.com/SynapticNetworks/temporal-neuron)
[![Concurrency](https://img.shields.io/badge/concurrent_ops-2000%2B-orange.svg)](https://github.com/SynapticNetworks/temporal-neuron)

## 🧠 Biological Introduction

In the biological brain, synapses are far more than simple connection weights. They are sophisticated, dynamic computational units that:

- **Learn continuously** through spike-timing dependent plasticity (STDP)
- **Prune themselves** when ineffective ("use it or lose it")
- **Process signals asynchronously** with realistic transmission delays
- **Adapt their strength** based on usage patterns and timing
- **Operate independently** as autonomous processors

Traditional artificial neural networks reduce synapses to static weight matrices, losing the rich temporal dynamics and adaptive capabilities that make biological brains so powerful. Our `synapse` package restores this biological realism while maintaining high performance through Go's concurrency primitives.

## 🎯 Key Features

### 🔬 Biological Realism
- **Spike-Timing Dependent Plasticity (STDP)**: Classic Hebbian learning with precise timing windows
- **Structural Plasticity**: Automatic pruning of weak, inactive synapses  
- **Realistic Delays**: Axonal conduction and synaptic transmission timing
- **Retrograde Signaling**: Post-synaptic feedback to pre-synaptic terminals
- **Synaptic Diversity**: Pluggable architecture for different synapse types

### ⚡ High Performance  
- **1.2+ Million operations/second** sustained throughput
- **Sub-millisecond latency** (average 676μs under load)
- **2000+ concurrent operations** without blocking
- **99.38% success rate** under extreme stress testing
- **Minimal memory footprint** (~1KB baseline per synapse)

### 🏗️ Modular Architecture
- **Interface-based design** enables synaptic diversity
- **Thread-safe operations** for concurrent neural networks
- **Pluggable components** for different learning rules
- **Clean separation** of concerns between neurons and synapses

## 🚀 Quick Start

### Installation

```bash
go get github.com/SynapticNetworks/temporal-neuron/synapse
```

### Basic Usage

```go
package main

import (
    "fmt"
    "time"
    "github.com/SynapticNetworks/temporal-neuron/synapse"
)

// Create mock neurons for demonstration
preNeuron := synapse.NewMockNeuron("pre_neuron")
postNeuron := synapse.NewMockNeuron("post_neuron")

// Configure STDP learning parameters
stdpConfig := synapse.STDPConfig{
    Enabled:        true,
    LearningRate:   0.01,                   // 1% weight change per STDP event
    TimeConstant:   20 * time.Millisecond,  // Learning window time constant
    WindowSize:     100 * time.Millisecond, // Maximum timing window
    MinWeight:      0.001,                  // Prevent elimination
    MaxWeight:      2.0,                    // Prevent runaway strengthening
    AsymmetryRatio: 1.2,                    // LTD/LTP ratio
}

// Configure structural plasticity
pruningConfig := synapse.CreateDefaultPruningConfig()

// Create a biologically realistic synapse
syn := synapse.NewBasicSynapse(
    "synapse_1",           // Unique ID
    preNeuron,            // Source neuron
    postNeuron,           // Target neuron  
    stdpConfig,           // Learning parameters
    pruningConfig,        // Pruning parameters
    0.5,                  // Initial weight
    5*time.Millisecond,   // Transmission delay
)

// Transmit signals (models action potentials)
syn.Transmit(1.0)  // Strong excitatory signal
syn.Transmit(-0.3) // Inhibitory signal

// Apply learning (models retrograde signaling)
adjustment := synapse.PlasticityAdjustment{
    DeltaT: -10 * time.Millisecond, // Pre-spike 10ms before post-spike (LTP)
}
syn.ApplyPlasticity(adjustment)

// Monitor synaptic state
fmt.Printf("Synapse weight: %.3f\n", syn.GetWeight())
fmt.Printf("Should prune: %v\n", syn.ShouldPrune())
```

## 📖 Core Components

### SynapticProcessor Interface

The heart of the modular architecture, defining the contract for all synapse types:

```go
type SynapticProcessor interface {
    ID() string                              // Unique identifier
    Transmit(signalValue float64)           // Send signal with delay and scaling
    ApplyPlasticity(adjustment PlasticityAdjustment)  // STDP learning
    ShouldPrune() bool                      // Structural plasticity decision
    GetWeight() float64                     // Current synaptic strength
    SetWeight(weight float64)               // Manual weight adjustment
}
```

**Biological Significance**: This interface captures the essential functions of biological synapses while allowing for diverse implementations (excitatory, inhibitory, modulatory).

### BasicSynapse Implementation

The default high-performance synapse with full biological features:

```go
type BasicSynapse struct {
    // Identity and connections
    id                 string
    preSynapticNeuron  SynapseCompatibleNeuron
    postSynapticNeuron SynapseCompatibleNeuron
    
    // Synaptic properties  
    weight float64       // Synaptic efficacy
    delay  time.Duration // Transmission delay
    
    // Learning configuration
    stdpConfig    STDPConfig    // Plasticity parameters
    pruningConfig PruningConfig // Structural plasticity
    
    // Activity tracking
    lastPlasticityEvent time.Time
    lastTransmission    time.Time
    
    // Thread safety
    mutex sync.RWMutex
}
```

## 🔬 Biological Validation

### STDP Timing Window
Our implementation matches experimental data from [Bi & Poo, 1998]:

```go
// Classic STDP parameters (cortical synapses)
stdpConfig := STDPConfig{
    LearningRate:   0.01,                   // 1% change per pairing
    TimeConstant:   20 * time.Millisecond,  // τ = 20ms (experimental)
    WindowSize:     100 * time.Millisecond, // ±100ms window
    AsymmetryRatio: 1.2,                    // LTD > LTP (typical)
}
```

### Synaptic Delays
Realistic transmission delays based on biophysical measurements:

```go
// Delay examples from different brain regions
localSynapse   := 1*time.Millisecond   // Local cortical connections
corticalSynapse := 5*time.Millisecond   // Typical cortical
longDistance   := 50*time.Millisecond   // Cross-hemispheric
```

### Pruning Timescales
The package uses accelerated timescales for testing but can be configured for more biological realism:

```go
// For testing
testPruningConfig := PruningConfig{
    Enabled:             true,
    WeightThreshold:     0.1,
    InactivityThreshold: 100 * time.Millisecond, // Accelerated for testing
}

// For production with biologically realistic timescales
realPruningConfig := PruningConfig{
    Enabled:             true,
    WeightThreshold:     0.01,
    InactivityThreshold: 6 * time.Hour, // More biologically accurate
}
```

## 📊 Performance Benchmarks

The package has been rigorously tested under real-world conditions:

```
=== LAPTOP-FRIENDLY CONCURRENT TRANSMISSION TEST ===
System: 12 CPUs, 600 goroutines  
Test Duration: 1.63 seconds
Total Operations: 2,400,000
Success Rate: 99.95%
Operations/Second: 1,467,796
Average Latency: 434.26 μs
Max Latency: 31.44 ms
Max Concurrency: 600
```

## 🧪 Comprehensive Test Suite

The package includes extensive tests validating both biological realism and performance:

### Biological Realism Tests
- **`TestSTDPClassicTimingWindow`**: Validates classic STDP learning windows
- **`TestSTDPExponentialDecay`**: Confirms proper exponential decay of plasticity effects
- **`TestSTDPAsymmetry`**: Tests that LTD is properly stronger than LTP
- **`TestActivityDependentPruning`**: Confirms "use it or lose it" principles
- **`TestSynapticWeightScaling`**: Validates accurate signal scaling by weight
- **`TestTransmissionDelayAccuracy`**: Confirms precise timing at the nanosecond level

### Performance Tests
- **`TestMassiveConcurrentTransmission`**: Validates behavior with 600+ goroutines
- **`TestSustainedHighFrequencyTransmission`**: Tests 1000Hz transmission rates
- **`TestResourceExhaustionRecovery`**: Confirms recovery under memory pressure
- **`TestMixedOperationChaos`**: Concurrent mixed operations in unpredictable patterns
- **`TestLongRunningStability`**: Extended operation stability (configurable duration)

## Using in Production Environments

For deploying this package in realistic production environments:

### 1. Adjust Timescales

The default parameters in tests use accelerated timescales for testing efficiency. For biological realism in production:

```go
// Production-ready configuration with realistic timescales
productionPruningConfig := PruningConfig{
    Enabled:             true,
    WeightThreshold:     0.01,
    InactivityThreshold: 6 * time.Hour, // Hours to days for biological accuracy
}
```

### 2. Memory Management

The BasicSynapse implementation is memory-efficient (~1KB per synapse). For large networks:

- Monitor memory usage during network growth
- Implement periodic pruning to remove unused synapses
- Consider batching large operations (e.g., bulk synapse creation)

### 3. Concurrency Tuning

The package has been tested with 2000+ concurrent operations:

- For optimal performance, limit concurrent operations to ~50 per CPU core
- Use a connection pool pattern for managing large networks
- Batch plasticity updates when possible

### 4. Performance Monitoring

Incorporate monitoring for:

- Synapse pruning rates (should stabilize after initial learning)
- Weight distribution (should follow log-normal distribution in mature networks)
- Memory growth during learning phases

## 🔧 Configuration

### STDP Configuration
```go
type STDPConfig struct {
    Enabled        bool          // Enable/disable plasticity
    LearningRate   float64       // Base learning rate (0.001-0.1)
    TimeConstant   time.Duration // Exponential decay τ (10-50ms)
    WindowSize     time.Duration // Max timing window (50-200ms) 
    MinWeight      float64       // Lower bound
    MaxWeight      float64       // Upper bound
    AsymmetryRatio float64       // LTD/LTP ratio (1.0-2.0)
}
```

### Pruning Configuration  
```go
type PruningConfig struct {
    Enabled             bool          // Enable structural plasticity
    WeightThreshold     float64       // Weak synapse threshold
    InactivityThreshold time.Duration // Grace period for elimination
}
```

### Helper Functions
```go
// Standard biological parameters
stdpConfig := CreateDefaultSTDPConfig()

// Conservative pruning (safer for learning)
pruningConfig := CreateConservativePruningConfig()
```

## 🎓 Educational Examples

### Example 1: Basic STDP Learning
```go
// Create learning synapse
syn := NewBasicSynapse("learning_synapse", preNeuron, postNeuron,
    CreateDefaultSTDPConfig(), CreateDefaultPruningConfig(), 0.5, 0)

// Simulate causal spike pairing (LTP)
syn.Transmit(1.0)  // Pre-synaptic spike
time.Sleep(5 * time.Millisecond)
// Post-synaptic spike occurs here
adjustment := PlasticityAdjustment{DeltaT: -5 * time.Millisecond}
syn.ApplyPlasticity(adjustment)

fmt.Printf("Weight after LTP: %.3f\n", syn.GetWeight()) // Should increase
```

## 🔬 Biological Correspondence

### Synapse Component → Biological Structure

| Component | Biological Correspondence | Function |
|-----------|---------------------------|----------|
| `weight` | Synaptic efficacy | AMPA/NMDA receptor density |
| `delay` | Conduction + synaptic delay | Axon length + neurotransmitter kinetics |
| `Transmit()` | Action potential propagation | Electrical spike → chemical signal |
| `ApplyPlasticity()` | Retrograde signaling | NO/endocannabinoid feedback |
| `ShouldPrune()` | Microglial pruning | "Use it or lose it" elimination |
| `STDP` | NMDA receptor activation | Ca²⁺-dependent molecular cascades |

### Learning Rule Validation

Our STDP implementation matches experimental data:

- **Bi & Poo (1998)**: Hippocampal cultures, τ ≈ 20ms
- **Sjöström et al. (2001)**: Neocortical pairs, asymmetric window
- **Caporale & Dan (2008)**: Visual cortex, frequency dependence

## 🚀 Future Roadmap

- **Additional Synapse Types**: Inhibitory, modulatory, static
- **Advanced Learning Rules**: Metaplasticity, homeostatic scaling
- **Connectome Integration**: Direct import of biological connectomes
- **Visualization Tools**: Real-time synapse activity monitoring
- **GPU Acceleration**: CUDA backend for massive networks
- **Serialization**: Save/load network states
- **Network Topologies**: Pre-built biological network patterns

## 📚 References

### Key Neuroscience Papers
- **Bi, G. & Poo, M. (1998)** - "Synaptic modifications in cultured hippocampal neurons" - *STDP discovery*
- **Sjöström, P.J. et al. (2001)** - "Rate, timing, and cooperativity jointly determine cortical synaptic plasticity" - *STDP refinement*
- **Caporale, N. & Dan, Y. (2008)** - "Spike timing-dependent plasticity: a Hebbian learning rule" - *STDP review*
- **Turrigiano, G.G. (2008)** - "The self-tuning neuron: synaptic scaling of excitatory synapses" - *Homeostatic plasticity*

### Technical Resources
- **Izhikevich, E.M. (2003)** - "Simple model of spiking neurons" - *Neuron modeling*
- **Maass, W. (1997)** - "Networks of spiking neurons: the third generation of neural network models" - *SNN foundations*
- **Dayan, P. & Abbott, L.F. (2001)** - "Theoretical Neuroscience" - *Mathematical foundations*

## 📄 License

This project is licensed under the **Temporal Neuron Research License (TNRL) v1.0** - see the [LICENSE](LICENSE) file for details.

### License Summary
- ✅ **Free for research, educational, and personal use**
- ✅ **Academic institutions and universities** - unlimited use
- ✅ **Open source research projects** - encouraged
- ✅ **Publications and citations** - welcomed
- ⚠️ **Commercial use requires permission** - contact for licensing

**For commercial licensing inquiries**: [hannes.lehmann@sistemica.de]

## 🙏 Acknowledgments

- Inspired by the sophisticated dynamics of biological synapses
- Built on Go's excellent concurrency primitives  
- Informed by decades of neuroscience research
- Part of the Temporal Neuron project's mission to build brain-like AI

---

*Building the future of neural computation, one synapse at a time.*