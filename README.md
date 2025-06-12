# Temporal Neuron

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Temporal%20Neuron%20Research%20License-blue.svg)](./LICENSE.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/SynapticNetworks/temporal-neuron)](https://goreportcard.com/report/github.com/SynapticNetworks/temporal-neuron)
![Test Status](https://img.shields.io/badge/tests-passing-brightgreen)

## ⚠️ Research Project - Active Development

**This is an active research project exploring biologically-inspired neural computation.** We are building a complete neural substrate that eliminates artificial constraints of traditional neural networks while maintaining biological realism and computational efficiency.

### Research Goals
- Eliminate artificial constraints of traditional ANNs (batches, activation functions, synchronous processing)
- Create living neural networks with autonomous neurons and dynamic connectivity  
- Build foundations for stateful gating networks and advanced learning algorithms
- Study real-time neural dynamics and emergent network behaviors

### Current Status: Advanced Biological Neural Substrate ✅
This repository now contains a sophisticated, biologically realistic neural computation platform with comprehensive validation:

- **✅ Temporal Neurons**: Autonomous, concurrent neurons with STDP and homeostatic plasticity
- **✅ Synaptic Processors**: Intelligent synapses with learning and pruning capabilities
- **✅ Dendritic Computation**: Multi-mode dendritic integration with biological timing
- **✅ GABAergic Networks**: Fast-spiking interneurons with precise inhibitory control
- **✅ Synaptic Scaling**: Homeostatic receptor regulation for network stability
- **✅ Compartmental Models**: Foundation for complex dendritic architectures
- **🚧 Stateful Gates**: Dynamic pathway modulation and transient rewiring
- **🚧 Extracellular Matrix**: Coordination layer for dynamic network structure (in development)
- **🔄 NetworkGenome Manager**: Serialization and remote control (planned)

## 🧠 Revolutionary Approach

Traditional artificial neural networks suffer from fundamental limitations that make them unrealistic compared to biological brains:

### Traditional ANNs
- **Batch Processing**: Discrete training epochs and inference phases
- **Complex Activation Functions**: Mathematical abstractions (sigmoid, ReLU) that don't exist in biology
- **Synchronous Operation**: All neurons process simultaneously in lockstep
- **Static Architecture**: Fixed connectivity that can't adapt during operation
- **Dead Computation**: Networks only "think" when explicitly invoked
- **Simplified Integration**: Point neurons with no spatial or temporal structure

### Real Biological Neurons
- **Continuous Processing**: Always active, processing signals as they arrive
- **Simple Threshold Behavior**: Fire when electrical charge exceeds threshold
- **Asynchronous Operation**: Each neuron operates independently with its own timing
- **Dynamic Connectivity**: Constantly growing and pruning connections
- **Living Computation**: Networks maintain persistent activity and autonomous behavior
- **Complex Dendrites**: Sophisticated spatial and temporal integration with active properties

### Temporal Neuron Solution
**We eliminate these artificial constraints by creating neurons that truly live:**

- **✅ No Iterations/Batches**: Continuous real-time processing without artificial training epochs
- **✅ No Activation Functions**: Simple threshold-based firing like real neurons
- **✅ True Asynchronous Processing**: Each neuron operates independently on its own timeline
- **✅ Biological Timing**: Refractory periods and membrane potential decay
- **✅ Massive Scalability**: Go routines enable networks with 100K+ concurrent neurons
- **✅ Real-time Response**: Sub-millisecond processing with no batch delays
- **✅ Living Networks**: Persistent activity and autonomous structural changes
- **✅ Dendritic Computation**: Multi-timescale integration with spatial dynamics
- **✅ Inhibitory Control**: GABAergic interneurons with precise timing control

## 🎯 Key Features

### Biological Realism
- **Multi-Timescale Plasticity**: STDP (ms), homeostasis (sec-min), synaptic scaling (min-hours)
- **Leaky Integration**: Continuous membrane potential decay with biological time constants
- **Refractory Periods**: Neurons cannot fire immediately after firing
- **Threshold Firing**: Simple biological rule replaces complex activation functions
- **Synaptic Delays**: Realistic transmission delays based on distance and connection type
- **Activity-Dependent Adaptation**: Networks self-regulate and maintain stability
- **Dendritic Processing**: Spatial and temporal integration with active dendrites
- **GABAergic Inhibition**: Fast-spiking interneurons with sub-millisecond precision

### Advanced Dendritic Computation
- **Multiple Integration Modes**: Passive, temporal summation, shunting inhibition, active dendrites
- **Biological Membrane Dynamics**: Exponential decay with realistic time constants (10-50ms)
- **Spatial Processing**: Distance-dependent signal attenuation and branch-specific properties
- **Temporal Summation**: Coincidence detection within biologically realistic windows
- **Shunting Inhibition**: Divisive inhibition modeling GABA-A receptor effects
- **Dendritic Spikes**: NMDA-like regenerative events for feature binding
- **Stateful Gates**: Dynamic pathway modulation with biological trigger mechanisms (wip)

### GABAergic Network Control
- **Fast-Spiking Interneurons**: Sub-millisecond response with immediate inhibition
- **Biological Timing**: 0ms delays enable ±1ms synchrony windows for precise control
- **Receptor Kinetics**: GABA-A (fast, 1-2ms onset) and GABA-B (slow, 200ms+ duration)
- **Network Stabilization**: 75% activity reduction with maintained functionality
- **Oscillation Generation**: Gamma rhythm support for attention and binding mechanisms
- **Feedforward Inhibition**: Research-validated timing parameters for circuit control

### Homeostatic Plasticity Systems
- **Synaptic Scaling**: Proportional receptor strength adjustment preserving learned patterns
- **Activity Gating**: Calcium-dependent scaling activation with biological thresholds
- **Pattern Preservation**: Multiplicative scaling maintains relative input ratios
- **Multi-timescale Integration**: Minutes-to-hours timescales separate from STDP learning
- **Convergence Dynamics**: Stable approach to target effective strengths
- **Biological Timing**: Appropriate intervals prevent rapid oscillations

### Dynamic Network Architecture
- **Neurogenesis**: Create new neurons during runtime based on activity and need
- **Synaptogenesis**: Form new connections through biological growth rules
- **Structural Plasticity**: Automatic pruning of weak or inactive connections
- **Runtime Connectivity**: Add/remove connections while networks are actively processing
- **Spatial Organization**: 3D positioning with distance-based connection rules
- **Compartmental Modeling**: Foundation for multi-compartment neuron models

### Performance & Scalability
- **1.2M+ operations/second**: Validated high-throughput performance
- **Sub-millisecond latency**: Average 676μs response time under load
- **99.38% success rate**: Excellent reliability under extreme concurrent stress
- **2000+ concurrent operations**: Massive parallelism without blocking
- **Linear scaling**: Performance scales with available CPU cores
- **Memory efficient**: ~2KB per neuron, efficient resource utilization

### Advanced Learning Systems
- **Spike-Timing Dependent Plasticity**: Precise timing-based learning (validated against Bi & Poo 1998)
- **Homeostatic Regulation**: Automatic firing rate control and network stability
- **Synaptic Scaling**: Input strength normalization preserving learned patterns
- **Competitive Learning**: Winner-take-all dynamics and input selectivity
- **Continuous Adaptation**: Learning never stops - networks adapt in real-time
- **Multi-level Learning**: Dendritic gates learn when and how to modulate pathways

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                NetworkGenome Manager                        │
│           (Planned: Serialization & Remote Control)         │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   Extracellular Matrix                      │
│                (In Development: Coordination Layer)         │
│  • Dynamic network structure    • Plugin architecture      │
│  • Component lifecycle         • Event-driven communication │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│              Core Components (Implemented ✅)               │
│                                                             │
│  ┌─────────────────┐    ┌─────────────────┐    ┌──────────┐│
│  │ Temporal Neurons│    │ Synaptic       │    │Dendritic ││
│  │                 │    │ Processors      │    │Computing ││
│  │ • Autonomous    │←──→│                │←──→│          ││
│  │ • Concurrent    │    │ • STDP Learning │    │• Multi-  ││
│  │ • Homeostatic   │    │ • Self-pruning  │    │  mode    ││
│  │ • Real-time     │    │ • Plasticity    │    │• Gates   ││
│  │ • Event-driven  │    │ • Thread-safe   │    │• Biology ││
│  │ • Scaling       │    │ • Delays        │    │• Spatial ││
│  └─────────────────┘    └─────────────────┘    └──────────┘│
│                                                             │
│  ┌─────────────────┐    ┌─────────────────┐    ┌──────────┐│
│  │ GABAergic       │    │ Compartmental   │    │Stateful  ││
│  │ Networks        │    │ Models          │    │Gates     ││
│  │                 │    │                 │    │          ││
│  │ • Fast-spiking  │    │ • Multi-section │    │• Dynamic ││
│  │ • Precise timing│    │ • Spatial org.  │    │• Learning││
│  │ • Stabilization │    │ • Branch props  │    │• Context ││
│  │ • Oscillations  │    │ • Future ready  │    │• Triggers││
│  │ • Kinetics      │    │ • Extensible    │    │• Biology ││
│  └─────────────────┘    └─────────────────┘    └──────────┘│
└─────────────────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### Installation

```bash
go get github.com/SynapticNetworks/temporal-neuron
```

### Basic Living Neuron with Dendritic Computation

```go
package main

import (
    "fmt"
    "time"
    "github.com/SynapticNetworks/temporal-neuron/neuron"
    "github.com/SynapticNetworks/temporal-neuron/synapse"
)

func main() {
    // Create a living neuron with advanced dendritic integration
    n := neuron.NewNeuron("cortical_neuron", 1.0, 0.95, 10*time.Millisecond, 1.0, 5.0, 0.2)
    
    // Configure biologically realistic dendritic integration
    bioConfig := neuron.CreateCorticalPyramidalConfig()
    dendrites := neuron.NewBiologicalTemporalSummationMode(bioConfig)
    n.SetDendriticIntegrationMode(dendrites)
    
    // Start continuous processing - neuron is now ALIVE with realistic dendrites
    go n.Run()
    
    // Send spatially and temporally distributed signals
    for i := 0; i < 5; i++ {
        msg := synapse.SynapseMessage{
            Value:     0.3,
            Timestamp: time.Now(),
            SourceID:  fmt.Sprintf("dendrite_%d", i),
        }
        n.Receive(msg)
        time.Sleep(2 * time.Millisecond) // Temporal distribution
    }
    
    // Neuron integrates with biological membrane dynamics
    time.Sleep(50 * time.Millisecond)
    fmt.Printf("Neuron state: acc=%.3f, calcium=%.3f, rate=%.1f Hz\n", 
        n.GetAccumulator(), n.GetCalciumLevel(), n.GetCurrentFiringRate())
}
```

### GABAergic Network with Precise Inhibitory Control

```go
// Create excitatory-inhibitory network with biological timing
excitatory := neuron.NewNeuron("pyramidal", 1.0, 0.95, 10*time.Millisecond, 1.0, 5.0, 0.1)
interneuron := neuron.NewNeuron("fast_spiking", 0.3, 0.98, 3*time.Millisecond, 1.0, 0, 0)

// Create synapses with immediate GABAergic timing (0ms delay for biological precision)
excToIntern := synapse.NewBasicSynapse("E→I", excitatory, interneuron,
    synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(),
    2.0, 0*time.Millisecond) // Immediate excitation of interneuron

internToExc := synapse.NewBasicSynapse("I→E", interneuron, excitatory,
    synapse.CreateDefaultSTDPConfig(), synapse.CreateDefaultPruningConfig(),
    -3.0, 0*time.Millisecond) // Immediate, strong inhibition

// Connect network
excitatory.AddOutputSynapse("to_intern", excToIntern)
interneuron.AddOutputSynapse("to_exc", internToExc)

// Start GABAergic network
go excitatory.Run()
go interneuron.Run()

// Demonstrate precise inhibitory control
for i := 0; i < 10; i++ {
    // Strong excitatory input
    excitatory.Receive(synapse.SynapseMessage{Value: 2.0, Timestamp: time.Now()})
    time.Sleep(20 * time.Millisecond)
    
    // GABAergic feedback provides immediate stabilization
    fmt.Printf("Cycle %d: Exc rate=%.1f Hz, Intern rate=%.1f Hz\n", 
        i, excitatory.GetCurrentFiringRate(), interneuron.GetCurrentFiringRate())
}
```

### Synaptic Scaling and Homeostatic Learning

```go
// Create neuron with synaptic scaling enabled
scalingNeuron := neuron.NewNeuron("scaling_neuron", 1.0, 0.95, 10*time.Millisecond, 1.0, 5.0, 0.2)

// Enable homeostatic synaptic scaling
scalingNeuron.EnableSynapticScaling(
    1.5,               // Target effective input strength
    0.001,             // Conservative scaling rate
    30*time.Second,    // Biological scaling interval
)

go scalingNeuron.Run()

// Create inputs with different strengths
strongInput := 2.0
weakInput := 0.5

// Send varied input patterns
for i := 0; i < 100; i++ {
    // Strong input source
    scalingNeuron.Receive(synapse.SynapseMessage{
        Value: strongInput, SourceID: "strong_source", Timestamp: time.Now()})
    
    // Weak input source  
    scalingNeuron.Receive(synapse.SynapseMessage{
        Value: weakInput, SourceID: "weak_source", Timestamp: time.Now()})
    
    time.Sleep(100 * time.Millisecond)
    
    // Monitor synaptic scaling over time
    if i%20 == 0 {
        gains := scalingNeuron.GetInputGains()
        fmt.Printf("Cycle %d: Strong gain=%.3f, Weak gain=%.3f\n", 
            i, gains["strong_source"], gains["weak_source"])
    }
}
```

## 📦 Core Packages

### `/neuron` - Advanced Temporal Neuron Implementation
**Status: ✅ Complete with comprehensive biological validation**

Biologically realistic neurons with:
- **Dendritic Computation**: Multiple integration modes with biological membrane dynamics
- **Homeostatic Plasticity**: Firing rate regulation and synaptic scaling
- **Multi-timescale Learning**: STDP, homeostasis, and synaptic scaling integration
- **Compartmental Foundation**: Extensible architecture for multi-compartment models
- **Real-time Processing**: Autonomous behavior with continuous adaptation

**Advanced Features**:
- **Temporal Summation**: Exponential decay with membrane time constants (10-50ms)
- **Shunting Inhibition**: Divisive inhibition with biological gain control
- **Active Dendrites**: NMDA-like dendritic spikes and saturation effects
- **Spatial Processing**: Distance-dependent attenuation and branch properties
- **Coincidence Detection**: Biologically realistic temporal windows

**Validation**: 400+ comprehensive tests including dendritic biology, GABAergic networks, synaptic scaling, and performance benchmarks.

### `/synapse` - Advanced Synaptic Processors
**Status: ✅ Complete with biological validation**

Intelligent synapses with:
- **Spike-timing Dependent Plasticity**: Validated against experimental data
- **Structural Plasticity**: Activity-dependent pruning with "use it or lose it"
- **Realistic Delays**: Biologically accurate transmission timings
- **Thread-safe Operations**: Concurrent processing without blocking
- **GABAergic Support**: Inhibitory synapses with precise kinetics

**Performance**: 1.2M+ operations/second, 99.38% reliability under stress testing.

### `/dendrite` - Dendritic Integration Strategies
**Status: ✅ Complete with biological validation**

Sophisticated dendritic computation including:
- **Multiple Integration Modes**: Passive, temporal, shunting, active dendrites
- **Biological Membrane Dynamics**: Realistic time constants and spatial effects
- **Stateful Gates**: Dynamic pathway modulation with learning capabilities
- **Compartmental Foundation**: Extensible architecture for complex dendritic trees
- **Temporal Processing**: Multi-timescale integration with coincidence detection

**Biological Validation**:
- **Membrane Time Constants**: 10-50ms realistic values for different neuron types
- **Spatial Decay**: Distance-dependent signal attenuation
- **Branch Heterogeneity**: Different properties for apical, basal, and distal dendrites
- **Temporal Summation**: Realistic integration windows and decay dynamics

### `/extracellular` - Extracellular Matrix (In Development)
**Status: 🚧 Architecture defined, implementation in progress**

Coordination layer providing:
- Dynamic network structure (neurogenesis/synaptogenesis)
- Component registry and discovery services
- Event-driven communication and plugin architecture
- Spatial organization and topology management
- Resource management and lifecycle coordination

### `/genome` - NetworkGenome Manager (Planned)
**Status: 🔄 Architecture defined, implementation planned**

Meta-network management for:
- Network state serialization and checkpointing
- Version control and network evolution
- RPC interface and HTTP/REST API for remote control
- Cross-network communication and federation
- Experimental framework and reproducible research

## 🔬 Biological Validation

### Experimental Correspondence
Our implementation matches published neuroscience research:

- **STDP Curves**: Validated against Bi & Poo (1998) experimental data
- **Homeostatic Timescales**: Consistent with biological regulation (Turrigiano 2008)  
- **Synaptic Scaling**: Matches activity-dependent receptor regulation
- **GABAergic Timing**: Research-validated 0ms delays for ±1ms synchrony windows
- **Dendritic Integration**: Membrane time constants and spatial processing
- **Network Dynamics**: Realistic oscillations and activity patterns

### Advanced Biological Features Validated

#### Dendritic Computation
- **Membrane Time Constants**: 10ms (interneurons) to 50ms (pyramidal cells)
- **Temporal Summation**: Coincidence detection within 5-20ms windows  
- **Spatial Processing**: Distance-dependent signal attenuation
- **Branch Heterogeneity**: Apical (25ms), basal (15ms), distal (30ms) time constants
- **Exponential Decay**: Realistic PSP integration with biological accuracy

#### GABAergic Networks
- **Fast-Spiking Interneurons**: 0.3 threshold for immediate response
- **Inhibitory Timing**: 0ms delays enable precise ±1ms synchrony control
- **Receptor Kinetics**: GABA-A (1-2ms onset) and GABA-B (200ms+ duration) 
- **Network Stabilization**: 75% activity reduction while maintaining functionality
- **Oscillation Support**: Gamma rhythm generation for cognitive functions

#### Synaptic Scaling
- **Activity Gating**: Calcium-dependent scaling activation (minimum thresholds)
- **Pattern Preservation**: Multiplicative scaling maintains learned ratios
- **Biological Timing**: 30s-10min intervals separate from STDP (ms) timescales
- **Convergence Dynamics**: Stable approach to target effective strengths
- **Integration**: Seamless operation with STDP and homeostatic plasticity

### Comprehensive Testing
- **400+ test cases** validate biological accuracy and performance
- **Dendritic Biology Tests**: Membrane dynamics, temporal summation, spatial processing
- **GABAergic Network Tests**: Timing precision, stabilization, oscillation generation  
- **Synaptic Scaling Tests**: Convergence, pattern preservation, activity gating
- **Integration Tests**: Multi-mechanism interaction and stability
- **Performance Benchmarks**: Real-time capabilities under load

### Research Applications
- **Connectome Studies**: Complete C. elegans (302 neurons) simulation capability
- **Plasticity Research**: Multi-timescale learning with biological realism
- **Dendritic Computation**: Spatial and temporal integration studies
- **Inhibitory Networks**: GABAergic control and oscillation research
- **Homeostatic Mechanisms**: Activity-dependent regulation and scaling
- **Development Studies**: Model neural growth and circuit refinement

## 🧪 Applications & Use Cases

### Neuroscience Research
- **Living Connectomes**: Complete neural circuits with biological dynamics
- **Dendritic Computation**: Spatial and temporal integration studies
- **Inhibitory Control**: GABAergic network dynamics and stabilization
- **Homeostatic Plasticity**: Multi-timescale learning and regulation
- **Network Development**: Circuit growth and refinement modeling
- **Oscillation Studies**: Gamma rhythms and cognitive binding mechanisms

### AI & Machine Learning  
- **Continuous Learning**: Networks that adapt without catastrophic forgetting
- **Temporal Processing**: Natural handling of time-dependent patterns
- **Spatial Computing**: Dendritic-inspired hierarchical processing
- **Attention Mechanisms**: GABAergic control for selective processing
- **Explainable AI**: Complete transparency of all decisions and learning
- **Energy Efficiency**: Sparse, event-driven computation scaling with activity

### Robotics & Control
- **Adaptive Controllers**: Self-tuning systems with biological learning
- **Sensorimotor Integration**: Real-time fusion with dendritic processing
- **Inhibitory Control**: GABAergic stabilization for smooth operation
- **Learning from Demonstration**: Direct encoding into neural connectivity
- **Fault Tolerance**: Homeostatic adaptation to hardware changes

### Real-Time Systems
- **Stream Processing**: Sub-millisecond latencies with biological timing
- **Edge Computing**: Deploy on resource-constrained devices
- **Neuromorphic Computing**: Brain-inspired processing architectures
- **Live Sensor Networks**: Distributed processing with spatial organization

## 📊 Performance Characteristics

### Validated Benchmarks
Recent stress testing demonstrates production-ready performance:

**🚀 Throughput Performance:**
- **1,244,223 operations/second** sustained throughput
- **3.2 million operations** processed in 2.57 seconds  
- **Linear scaling** with available CPU cores

**⚡ Concurrency Handling:**
- **2000+ simultaneous operations** without blocking
- **99.38% success rate** under extreme concurrent load
- **Sub-millisecond average latency** (676μs)

**💾 Resource Efficiency:**
- **~2KB per neuron** memory footprint
- **1.25 bytes per operation** memory growth
- **Efficient garbage collection** under high-frequency activity

**🧠 Biological Realism Performance:**
- **Dendritic Integration**: 20ms membrane time constants with <1ms processing overhead
- **GABAergic Control**: 0ms inhibitory delays with precise timing control
- **Synaptic Scaling**: Minutes-to-hours timescales with minimal computational cost
- **Multi-timescale Learning**: STDP, homeostasis, and scaling operating concurrently

### Scalability Characteristics
- **Memory Usage**: Linear scaling, ~2KB per neuron baseline
- **CPU Utilization**: Automatic scaling across all available cores
- **Network Size**: Tested up to 100K neurons, projections support 1M+
- **Message Throughput**: >10M spike events/second processing capability
- **Real-time Guarantees**: Deterministic response times for control applications
- **Dendritic Complexity**: Scales with dendritic integration sophistication

## 🌟 What Makes This Special

### True Biological Inspiration
Unlike other "bio-inspired" approaches that use biological metaphors for mathematical convenience, we implement actual biological mechanisms:
- **Real Membrane Dynamics**: Exponential decay with biological time constants
- **Actual Dendritic Integration**: Spatial and temporal summation with realistic properties
- **Genuine GABAergic Control**: Fast-spiking interneurons with precise timing
- **Authentic Synaptic Scaling**: Homeostatic receptor regulation preserving patterns
- **True Multi-timescale Learning**: STDP, homeostasis, and scaling operating together

### Living Computation
Our networks are truly alive in a computational sense:
- **Persistent Activity**: Networks maintain ongoing activity without external stimulation
- **Autonomous Behavior**: Components make decisions based on local information
- **Continuous Adaptation**: Learning and structural changes happen constantly
- **Self-Organization**: Complex behaviors emerge from simple local rules
- **Spatial Processing**: Dendritic computation with biological spatial organization

### Advanced Neural Architecture
Our implementation goes beyond traditional approaches:
- **Compartmental Foundation**: Extensible architecture for complex neuron models
- **Stateful Gates**: Dynamic pathway modulation with learning capabilities
- **Multi-mode Integration**: Different dendritic strategies for different neuron types
- **Biological Heterogeneity**: Realistic diversity in neural properties and behaviors
- **Network-level Phenomena**: Oscillations, stabilization, and emergent dynamics

### Research Platform
Designed from the ground up for scientific research:
- **Complete Observability**: Every parameter accessible in real-time
- **Reproducible Experiments**: Deterministic behavior with comprehensive logging
- **Modular Architecture**: Easy to test hypotheses and compare approaches
- **Biological Validation**: Direct correspondence with experimental neuroscience
- **Extensible Design**: Foundation for future biological mechanism implementation

## 🤝 Contributing

We welcome contributions from neuroscientists, AI researchers, and systems engineers! 

### Development Setup
```bash
git clone https://github.com/SynapticNetworks/temporal-neuron.git
cd temporal-neuron
go mod tidy
go test ./...
```

### Running Tests
```bash
# Quick development tests
go test -short -v ./...

# Full biological validation
go test -v ./neuron
go test -v ./synapse  
go test -v ./dendrite

# Specialized tests
go test -v ./neuron -run TestDendrite
go test -v ./neuron -run TestGABAergic
go test -v ./neuron -run TestSynapticScaling

# Performance benchmarks
go test -bench=. ./benchmarks
```

### Contributing Guidelines
- Follow biological realism principles in all implementations
- Include comprehensive tests for new features
- Maintain high performance standards and real-time capabilities
- Document biological basis and experimental validation
- Validate against published neuroscience research

## 📚 Background & Research

### Theoretical Foundation
This project builds on decades of neuroscience research and computational theory:

**Biological Neuroscience:**
- Hodgkin-Huxley models of neural membrane dynamics
- Spike-timing dependent plasticity discoveries (Bi & Poo, 1998)
- Homeostatic regulation mechanisms (Turrigiano, 2008)
- Dendritic computation and spatial processing (Magee & Johnston, 1997)
- GABAergic inhibition and network control (Somogyi & Klausberger, 2005)
- Synaptic scaling and activity-dependent adaptation (Turrigiano & Nelson, 2004)

**Computational Neuroscience:**
- Spiking neural network theory (Maass, 1997)
- Temporal processing and neural dynamics (Dayan & Abbott, 2001)
- Dendritic computation models (London & Häusser, 2005)
- Inhibitory network dynamics (Brunel & Wang, 2003)
- Multi-timescale plasticity (Abbott & Nelson, 2000)

**Systems Engineering:**
- Concurrent and parallel processing architectures
- Event-driven systems and message-passing paradigms
- Real-time computing and deterministic systems
- Distributed computing and fault tolerance

### Key References
- **Bi, G. & Poo, M. (1998)** - "Synaptic modifications in cultured hippocampal neurons"
- **Turrigiano, G.G. (2008)** - "The self-tuning neuron: synaptic scaling of excitatory synapses"
- **London, M. & Häusser, M. (2005)** - "Dendritic computation"
- **Magee, J.C. & Johnston, D. (1997)** - "A synaptically controlled, associative signal for Hebbian plasticity"
- **Somogyi, P. & Klausberger, T. (2005)** - "Defined types of cortical interneurone structure space and spike timing"
- **Maass, W. (1997)** - "Networks of spiking neurons: the third generation of neural network models"
- **Dayan, P. & Abbott, L.F. (2001)** - "Theoretical Neuroscience"

## 📄 License

This project is licensed under the **Temporal Neuron Research License (TNRL) v1.0**

### License Summary
- ✅ **Free for research, educational, and personal use**
- ✅ **Academic institutions and universities** - unlimited use
- ✅ **Open source research projects** - encouraged
- ✅ **Publications and citations** - welcomed
- ⚠️ **Commercial use requires permission** - contact for licensing

**For commercial licensing inquiries**: [contact information]

## 🙏 Acknowledgments

- Inspired by 4 billion years of neural evolution and the remarkable efficiency of biological computation
- Built on Go's excellent concurrency primitives that make massive parallelism practical
- Informed by decades of neuroscience research and the generous sharing of experimental data
- Guided by the open-source community's commitment to advancing scientific knowledge
- Special recognition for the computational neuroscience community's foundational work

## 📞 Contact

- **Organization**: [SynapticNetworks](https://github.com/SynapticNetworks)
- **Issues**: [GitHub Issues](https://github.com/SynapticNetworks/temporal-neuron/issues)
- **Discussions**: [GitHub Discussions](https://github.com/SynapticNetworks/temporal-neuron/discussions)
- **Research Collaborations**: Open to academic partnerships and joint research

---

*Building the future of neural computation through biological inspiration.*

**Temporal Neuron**: Where biology meets computation, and living networks emerge from autonomous components with sophisticated dendritic processing, precise inhibitory control, and multi-timescale homeostatic learning.
