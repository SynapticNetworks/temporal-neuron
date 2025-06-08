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

**Key Methods:**

#### Signal Transmission
```go
func (s *BasicSynapse) Transmit(signalValue float64)
```
Models biological synaptic transmission:
1. **Signal scaling** by synaptic weight (efficacy)
2. **Realistic delays** for axonal conduction + synaptic processing  
3. **Asynchronous delivery** via `time.AfterFunc`
4. **Message formatting** with timing metadata for STDP

**Biological Process**: Pre-synaptic action potential → axonal conduction → neurotransmitter release → post-synaptic potential

#### Spike-Timing Dependent Plasticity
```go
func (s *BasicSynapse) ApplyPlasticity(adjustment PlasticityAdjustment)
```
Implements Hebbian learning with precise timing:
- **LTP (strengthening)**: Pre-spike before post-spike (causal)
- **LTD (weakening)**: Pre-spike after post-spike (anti-causal)  
- **Exponential decay** with timing distance
- **Weight bounds** enforcement

**Biological Basis**: Models NMDA receptor-mediated calcium signaling and downstream molecular cascades.

#### Structural Plasticity  
```go
func (s *BasicSynapse) ShouldPrune() bool
```
"Use it or lose it" synaptic elimination:
- **Weight threshold**: Too weak to be effective?
- **Activity threshold**: Inactive for too long?
- **Two-factor protection**: Prevents premature pruning

**Biological Process**: Models microglial synaptic pruning observed in development and adult plasticity.

### STDP Learning Function

The core learning algorithm implementing biological timing rules:

```go
func calculateSTDPWeightChange(timeDifference time.Duration, config STDPConfig) float64
```

**Mathematical Model**:
- **Causal (Δt < 0)**: `learning_rate * exp(Δt/τ)` → LTP
- **Anti-causal (Δt > 0)**: `-learning_rate * asymmetry * exp(-Δt/τ)` → LTD  
- **Outside window**: `0` (no plasticity)

**Biological Correspondence**: 
- **Δt < 0**: NMDA receptor activation → Ca²⁺ influx → CaMKII → AMPA insertion → LTP
- **Δt > 0**: Weaker Ca²⁺ signal → phosphatases → AMPA removal → LTD

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
Biologically realistic structural plasticity timescales:

```go
pruningConfig := PruningConfig{
    WeightThreshold:     0.01,            // 1% of max strength
    InactivityThreshold: 5 * time.Minute, // Grace period for learning
}
```

## 📊 Performance Benchmarks

### Stress Test Results

Our comprehensive stress testing validates production readiness:

```
=== LAPTOP-FRIENDLY CONCURRENT TRANSMISSION TEST ===
System: 16 CPUs, 800 goroutines  
Test Duration: 2.57 seconds
Total Operations: 3,200,000
Success Rate: 99.38%
Operations/Second: 1,244,223
Average Latency: 676.05 μs
Max Latency: 600.68 ms
Max Concurrency: 2,064
Peak Memory: 5,924 MB
Memory Growth: 4 MB (1.25 bytes/operation)
```

### Key Performance Metrics
- ✅ **1.2M+ operations/second** sustained throughput
- ✅ **99.38% success rate** under extreme concurrent load  
- ✅ **Sub-millisecond average latency** (676μs)
- ✅ **2000+ concurrent operations** without blocking
- ✅ **Minimal memory growth** (1.25 bytes per operation)
- ✅ **Linear scaling** with CPU cores

### Benchmark Comparison
```bash
# Run performance benchmarks
go test -bench=BenchmarkTransmission -benchmem
go test -bench=BenchmarkPlasticity -benchmem  
go test -bench=BenchmarkConcurrent -benchmem
```

## 🧪 Testing

### Comprehensive Test Suite

```bash
# Quick development tests (30 seconds)
go test -short -v ./synapse

# Full biological validation tests  
go test -v ./synapse

# Stress testing (laptop-friendly)
go test -v ./synapse -run TestMassiveConcurrentTransmission

# Extended stability testing (custom duration)
go test -v ./synapse -run TestLongRunningStability -args -long-run=24h
```

### Test Categories

1. **Biological Realism Tests** (`synapse_biology_test.go`)
   - STDP timing windows and exponential decay
   - Activity-dependent pruning validation  
   - Synaptic transmission fidelity
   - Realistic parameter ranges

2. **Unit Tests** (`synapse_test.go`)
   - Synapse creation and initialization
   - Signal transmission mechanics
   - Weight management and bounds
   - Configuration validation

3. **Robustness Tests** (`synapse_robustness_test.go`)  
   - Massive concurrent access (2000+ goroutines)
   - High-frequency activity (1kHz+ sustained)
   - Resource exhaustion and recovery
   - Numerical stability with extreme values
   - Long-running stability (configurable duration)

## 🏗️ Architecture

### Modular Design

```
synapse/
├── synapse.go              # Core implementation
├── synapse_test.go         # Unit tests
├── synapse_biology_test.go # Biological validation
└── synapse_robustness_test.go # Stress tests
```

### Interface Hierarchy

```go
// Core synapse contract
SynapticProcessor interface {
    ID() string
    Transmit(signalValue float64) 
    ApplyPlasticity(adjustment PlasticityAdjustment)
    ShouldPrune() bool
    GetWeight() float64
    SetWeight(weight float64)
}

// Neuron communication contract  
SynapseCompatibleNeuron interface {
    ID() string
    Receive(msg SynapseMessage)
}
```

### Message System

```go
type SynapseMessage struct {
    Value     float64    // Signal strength
    Timestamp time.Time  // Precise timing for STDP
    SourceID  string     // Pre-synaptic neuron ID
    SynapseID string     // Transmitting synapse ID
}
```

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

### Example 2: Network with Feedback
```go
// Create network with excitatory and inhibitory connections
excitatory := NewBasicSynapse("E→I", excitatoryNeuron, inhibitoryNeuron, 
    CreateDefaultSTDPConfig(), CreateDefaultPruningConfig(), 1.0, 10*time.Millisecond)

inhibitory := NewBasicSynapse("I→E", inhibitoryNeuron, excitatoryNeuron,
    CreateDefaultSTDPConfig(), CreateDefaultPruningConfig(), -0.8, 15*time.Millisecond)

// Stimulate network
excitatory.Transmit(1.5) // Strong excitation
// Network dynamics emerge from timing and connectivity
```

### Example 3: Dynamic Synaptic Pruning
```go
// Create weak synapse that should be pruned
weakSynapse := NewBasicSynapse("weak", preNeuron, postNeuron,
    CreateDefaultSTDPConfig(), 
    PruningConfig{
        Enabled:             true,
        WeightThreshold:     0.1,
        InactivityThreshold: 1 * time.Minute,
    }, 
    0.05,  // Weak initial weight
    0)

// Check pruning status over time
time.Sleep(2 * time.Minute)
if weakSynapse.ShouldPrune() {
    fmt.Println("Synapse marked for elimination")
}
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

## 🌟 Advanced Features

### Thread-Safe Concurrent Operation
- All methods are thread-safe using `sync.RWMutex`
- Non-blocking signal transmission with `time.AfterFunc`
- Atomic operations for performance metrics

### Memory Efficiency
- Minimal per-synapse overhead (~1KB baseline)
- No goroutine per synapse (unlike some implementations)
- Efficient message passing without copying

### Monitoring and Debugging
```go
// Get detailed activity information
info := synapse.GetActivityInfo()
fmt.Printf("Last transmission: %v\n", info["lastTransmission"])
fmt.Printf("Last plasticity: %v\n", info["lastPlasticity"])
fmt.Printf("Weight: %.3f\n", info["weight"])

// Check if synapse is active
isActive := synapse.IsActive(1 * time.Minute)
```

## 🚀 Future Roadmap

### Enhanced Synapse Concepts

#### 🧵 **Goroutine-Based Enhanced Synapse**
*Research Direction: Individual Synaptic Processors*

While our current `BasicSynapse` runs efficiently in the pre-synaptic neuron's goroutine, we're exploring an advanced `EnhancedSynapse` where each synapse operates as its own independent goroutine. This could enable:

**🔬 Biological Advantages:**
- **Individual Synaptic Computation**: Each synapse processes its own local learning rules, homeostasis, and state evolution
- **Asynchronous Plasticity**: Independent timing for each synapse's learning updates
- **Complex Synaptic Dynamics**: Multi-timescale processes (protein synthesis, gene expression)
- **Synaptic Autonomy**: True independence modeling real synaptic behavior

**⚡ Technical Benefits:**
- **Parallel Plasticity**: Thousands of synapses learning simultaneously
- **Local Processing**: Each synapse maintains its own complex state and history
- **Scalable Architecture**: Perfect scaling with multi-core systems
- **Event-Driven**: Synapses only consume resources when active

**💡 Implementation Concept:**
```go
type EnhancedSynapse struct {
    // Core synapse state
    id     string
    weight float64
    
    // Independent processing
    inputChannel  chan SynapseMessage
    outputChannel chan SynapseMessage
    controlChannel chan SynapseControl
    
    // Advanced learning state
    eligibilityTrace float64
    metaplasticity   float64
    proteinLevels    map[string]float64
    
    // Autonomous processing
    ctx    context.Context
    cancel context.CancelFunc
}

func (s *EnhancedSynapse) Run() {
    ticker := time.NewTicker(1 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
        case msg := <-s.inputChannel:
            s.processSignal(msg)
        case <-ticker.C:
            s.autonomousUpdate() // Continuous synaptic maintenance
        case <-s.ctx.Done():
            return
        }
    }
}
```

**📊 Performance Considerations:**
- **Memory**: ~8KB per synapse (vs 1KB for BasicSynapse)
- **CPU**: Excellent parallelization but higher base overhead
- **Scalability**: Ideal for networks with complex synaptic dynamics
- **Use Cases**: Research simulations, detailed biological modeling

**🎯 When to Use Enhanced vs Basic:**
- **BasicSynapse**: Production networks, high throughput, simple learning
- **EnhancedSynapse**: Research applications, complex synaptic biology, detailed modeling

This represents the cutting edge of synaptic modeling - each synapse as a fully autonomous processor, just like in the biological brain.

---

### Other Planned Features
- [ ] **Additional Synapse Types**: Inhibitory, modulatory, static
- [ ] **Advanced Learning Rules**: Metaplasticity, homeostatic scaling
- [ ] **Connectome Integration**: Direct import of biological connectomes
- [ ] **Visualization Tools**: Real-time synapse activity monitoring
- [ ] **GPU Acceleration**: CUDA backend for massive networks
- [ ] **Serialization**: Save/load network states
- [ ] **Network Topologies**: Pre-built biological network patterns

## 🤝 Contributing

We welcome contributions! This package is part of the larger Temporal Neuron project building biologically realistic neural networks.

### Development Setup
```bash
git clone https://github.com/SynapticNetworks/temporal-neuron.git
cd temporal-neuron/synapse
go mod tidy
go test -v ./...
```

### Contributing Guidelines
- Follow biological realism principles
- Include comprehensive tests for new features
- Maintain high performance standards
- Document biological basis for implementations

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

This license encourages scientific advancement while ensuring sustainable project development.

**For commercial licensing inquiries**: [hannes.lehmann@sistemica.de]

## 🙏 Acknowledgments

- Inspired by the sophisticated dynamics of biological synapses
- Built on Go's excellent concurrency primitives  
- Informed by decades of neuroscience research
- Part of the Temporal Neuron project's mission to build brain-like AI

---

*Building the future of neural computation, one synapse at a time.*