# Temporal Neuron

[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Temporal%20Neuron%20Research%20License-blue.svg)](./LICENSE.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/SynapticNetworks/temporal-neuron)](https://goreportcard.com/report/github.com/SynapticNetworks/temporal-neuron)
![Test Status](https://img.shields.io/badge/tests-passing-brightgreen)

## Overview

The temporal-neuron architecture supports sophisticated **retrograde feedback** mechanisms where post-synaptic neurons can influence the behavior of their pre-synaptic partners. This bidirectional communication enables advanced learning algorithms, homeostatic regulation, and network-wide coordination that goes far beyond traditional feedforward architectures.

⚠️ **Research Project - Active Development**: This is an active research project exploring biologically-inspired neural computation with autonomous neurons, dynamic connectivity, and real-time processing capabilities.

## 🧠 Revolutionary Approach

Traditional artificial neural networks suffer from fundamental limitations that make them unrealistic compared to biological brains:

### Traditional ANNs vs. Biological Reality
- **Batch Processing** → **Continuous Processing**: Always active, processing signals as they arrive
- **Complex Activation Functions** → **Simple Threshold Behavior**: Fire when electrical charge exceeds threshold  
- **Synchronous Operation** → **Asynchronous Operation**: Each neuron operates independently with its own timing
- **Static Architecture** → **Dynamic Connectivity**: Constantly growing and pruning connections
- **Dead Computation** → **Living Computation**: Networks maintain persistent activity and autonomous behavior

### Temporal Neuron Solution
We eliminate artificial constraints by creating neurons that truly live:
- ✅ **No Iterations/Batches**: Continuous real-time processing without artificial training epochs
- ✅ **No Activation Functions**: Simple threshold-based firing like real neurons
- ✅ **True Asynchronous Processing**: Each neuron operates independently on its own timeline
- ✅ **Biological Timing**: Refractory periods and membrane potential decay
- ✅ **Massive Scalability**: Go routines enable networks with 100K+ concurrent neurons
- ✅ **Living Networks**: Persistent activity and autonomous structural changes

## Biological Foundation

### Retrograde Signaling

In biological neural networks, communication is bidirectional. While primary signal flow is forward (pre-synaptic to post-synaptic), there are mechanisms for **backward signaling**:

- **Endocannabinoids**: Lipid-based molecules released by post-synaptic neurons that travel backward across synapses to modulate pre-synaptic neurotransmitter release
- **Nitric Oxide (NO)**: A gaseous messenger that diffuses from post-synaptic to pre-synaptic terminals, affecting plasticity and excitability
- **Brain-Derived Neurotrophic Factor (BDNF)**: Growth factors that provide long-term retrograde signaling for synaptic strengthening
- **Anti-Hebbian Plasticity**: Mechanisms where post-synaptic silence weakens pre-synaptic inputs ("use it or lose it")

### Biological Examples

**Visual System**: In the retina, horizontal cells provide retrograde feedback to photoreceptors, adjusting their sensitivity based on overall light levels.

**Motor Learning**: During skill acquisition, post-synaptic motor neurons in the spinal cord send retrograde signals to adjust the strength and timing of inputs from motor cortex.

**Homeostatic Scaling**: When post-synaptic neurons become too active or too quiet, they release retrograde factors that adjust the strength of all their inputs to maintain stable firing rates.

**Fear Conditioning**: In the amygdala, successful fear associations trigger retrograde signals that strengthen the synaptic pathways that led to the correct prediction, while failed predictions weaken them.

## Architecture Implementation

### Signal Flow Patterns

The temporal-neuron architecture supports multiple retrograde feedback patterns:

```
┌─────────────┐    Forward Signal    ┌─────────────┐
│             │ ────────────────────→ │             │
│ Pre-Neuron  │                      │ Post-Neuron │
│             │ ←──────────────────── │             │
└─────────────┘   Retrograde Signal  └─────────────┘
```

### Implementation Mechanisms

#### 1. Electrical Signaling (Primary Method)

The cleanest implementation uses the existing electrical signaling infrastructure. Post-synaptic neurons can send electrical signals back through the matrix to adjust pre-synaptic neuron properties:

- **Post-neuron** calculates timing relationships and effectiveness
- **Post-neuron** calls `SendElectricalSignal()` with adjustment parameters
- **Matrix** routes the signal to the appropriate pre-synaptic neurons
- **Pre-neuron** receives signal via `OnSignal()` and adjusts threshold, excitability, or firing patterns

This mechanism supports:
- **Spike-timing dependent plasticity**: Adjustments based on precise timing relationships
- **Activity-dependent scaling**: Global adjustments based on post-synaptic firing rates
- **Competitive learning**: Weakening of poorly-timed inputs

#### 2. Chemical Signaling (Advanced)

For more sophisticated retrograde feedback, the chemical signaling system can be used:

- **Post-neuron** calls `ReleaseChemical()` with retrograde ligands (endocannabinoids, nitric oxide)
- **Matrix** diffuses the chemical through the extracellular space
- **Pre-neurons** with appropriate receptors receive signals via `Bind()`
- **Pre-neurons** adjust release probability, excitability, or other properties

This enables:
- **Volume transmission**: Retrograde signals affecting multiple pre-synaptic partners
- **Neuromodulation**: Context-dependent adjustments based on network state
- **Homeostatic regulation**: Long-term stability mechanisms

#### 3. Synaptic Mediation

Synapses themselves can implement retrograde feedback by:

- Tracking post-synaptic response effectiveness
- Adjusting their own weights and properties
- Sending feedback signals to pre-synaptic neurons via callbacks
- Implementing sophisticated plasticity rules (STDP, BCM, homeostatic scaling)

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                NetworkGenome Manager                        │
│           (Planned: Serialization & Remote Control)         │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   Extracellular Matrix                      │
│                (Coordination Layer: ✅ Implemented)         │
│  • Neurogenesis & synaptogenesis   • Chemical signaling    │
│  • Component lifecycle             • Spatial organization   │
│  • Health monitoring              • Gap junction management │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│              Core Components (✅ Implemented)               │
│                                                             │
│  ┌─────────────────┐    ┌─────────────────┐    ┌──────────┐│
│  │ Temporal Neurons│    │ Synaptic       │    │Dendritic ││
│  │                 │    │ Processors      │    │Computing ││
│  │ • Autonomous    │←──→│                │←──→│          ││
│  │ • Concurrent    │    │ • STDP Learning │    │• Multi-  ││
│  │ • Homeostatic   │    │ • Self-pruning  │    │  mode    ││
│  │ • Real-time     │    │ • Plasticity    │    │• Biology ││
│  │ • Event-driven  │    │ • Thread-safe   │    │• Spatial ││
│  │ • Scaling       │    │ • Delays        │    │• Temporal││
│  └─────────────────┘    └─────────────────┘    └──────────┘│
│                                                             │
│  ┌─────────────────┐    ┌─────────────────┐    ┌──────────┐│
│  │ Chemical        │    │ Astrocyte       │    │Microglial││
│  │ Modulation      │    │ Networks        │    │Health    ││
│  │                 │    │                 │    │Monitoring││
│  │ • Multi-ligand  │    │ • Territorial   │    │• Activity││
│  │ • Diffusion     │    │ • Spatial query │    │• Patrol  ││
│  │ • Concentration │    │ • Connectivity  │    │• Cleanup ││
│  │ • Binding       │    │ • Thread-safe   │    │• Support ││
│  │ • Modulation    │    │ • Scalable      │    │• Lifecycle││
│  └─────────────────┘    └─────────────────┘    └──────────┘│
└─────────────────────────────────────────────────────────────┘
```

## Core Architecture Components

### Extracellular Matrix Coordination System

The central coordination layer manages all biological processes through specialized subsystems:

#### **Astrocyte Network**
- **Territorial organization** with realistic domain sizes (10-200μm)
- **Spatial query resolution** to micrometer precision
- **Synaptic connectivity tracking** and activity monitoring
- **Thread-safe concurrent access** for biological scalability
- **Species-specific variations** (human vs mouse astrocyte characteristics)

#### **Microglial Health Monitoring**
- **Real-time surveillance** of neural component health
- **Activity-dependent assessment** with configurable sensitivity
- **Lifecycle coordination** for neurogenesis and apoptosis
- **Territorial patrol systems** with biological timing (100ms-2s rates)
- **Cleanup coordination** for network maintenance

#### **Chemical Modulator System**
- **Multi-neurotransmitter support** (GABA, glutamate, dopamine, serotonin, calcium)
- **Spatial diffusion modeling** with realistic concentration gradients
- **Receptor binding kinetics** and competitive interactions
- **Activity-dependent release** with threshold-based triggering
- **Pharmacological simulation** capabilities

#### **Signal Mediator (Gap Junctions)**
- **Bidirectional electrical coupling** with configurable conductance
- **Sub-millisecond propagation** delays (<0.1ms)
- **Synchronization support** for network oscillations
- **Biological conductance ranges** (0.05-1.0 nS)

### Neural Components

#### **Advanced Neuron Implementation**
- **Multiple dendritic modes**: Passive, active, and biological temporal summation
- **Homeostatic plasticity** with configurable target firing rates
- **STDP feedback mechanisms** with biological timing windows
- **Synaptic scaling** for network stability
- **Custom behavior system** for research extensibility
- **Coincidence detection** with NMDA-like voltage dependence

#### **Sophisticated Synapse System**
- **STDP-enabled plasticity** with asymmetric learning windows
- **Spatial delay calculation** based on axonal conduction velocity
- **Weight adaptation** with biological constraints
- **Activity-dependent modulation** through chemical signaling

## 🎯 Key Features

### Biological Realism
- **Multi-Timescale Plasticity**: STDP (ms), homeostasis (sec-min), synaptic scaling (min-hours)
- **Leaky Integration**: Continuous membrane potential decay with biological time constants
- **Refractory Periods**: Neurons cannot fire immediately after firing
- **Threshold Firing**: Simple biological rule replaces complex activation functions
- **Synaptic Delays**: Realistic transmission delays based on distance and connection type
- **Activity-Dependent Adaptation**: Networks self-regulate and maintain stability

### Dynamic Network Architecture
- **Neurogenesis**: Create new neurons during runtime based on activity and need
- **Synaptogenesis**: Form new connections through biological growth rules
- **Structural Plasticity**: Automatic pruning of weak or inactive connections
- **Runtime Connectivity**: Add/remove connections while networks are actively processing
- **Spatial Organization**: 3D positioning with distance-based connection rules
- **Factory System**: Biological component creation with proper integration

### Neurochemical Modulation

#### **Neurotransmitter Effects**
- **GABA enhancement**: Realistic inhibitory effects reducing neural excitability
- **Glutamate modulation**: Adjustable excitatory signaling strength and reliability
- **Dose-dependent responses**: Chemical concentrations produce proportional effects

#### **Intoxication Modeling**
The system accurately simulates substance effects on neural circuits:
- **Motor coordination impairment**: Progressive degradation of fine motor control
- **Selective vulnerability**: Weak signals affected more than strong signals
- **BAC correlation**: Effects scale realistically with blood alcohol concentration (0.05%-0.25%)
- **Recovery patterns**: Chemical clearance enables neural function restoration

#### **Validation Testing**
- **Motor circuit testing**: Multi-neuron circuits demonstrate coordination degradation
- **Complex cortical simulation**: 8-neuron networks with strict failure conditions
- **Biological accuracy**: Matches known intoxication patterns in neural systems

### Spatial Processing

#### **3D Spatial Awareness**
- **Distance-based delays**: Realistic axonal conduction timing
- **Spatial organization**: Components positioned in 3D coordinate space
- **Propagation modeling**: Variable conduction velocities (0.5-120 m/s)
- **Biological scenarios**: Cortical circuit distance validation

#### **Delay Integration System**
- **Synaptic delays**: Base transmission delays (0.5-5ms)
- **Spatial delays**: Distance-dependent propagation timing
- **Combined timing**: Realistic total transmission delays
- **Validation testing**: Comprehensive timing accuracy verification

## Learning Algorithms Enabled

### Spike-Timing Dependent Plasticity (STDP)

Post-synaptic neurons track the timing relationship between their firing and incoming spikes. When they fire shortly after receiving an input (causality), they send positive retrograde feedback. When they fire before an input arrives (anti-causality), they send negative feedback.

### Homeostatic Plasticity

Post-synaptic neurons monitor their own firing rates. If they become too active or too quiet compared to their target rates, they send retrograde signals to scale all their inputs up or down proportionally, maintaining network stability.

### Predictive Coding

In hierarchical networks, higher-level neurons can send retrograde "prediction error" signals to lower levels, teaching them to better predict upcoming patterns and reducing overall network prediction error.

### Attention and Gating

Post-synaptic neurons can implement attention mechanisms by selectively sending positive retrograde feedback to inputs that are currently relevant, effectively gating information flow based on context.

## Functional Benefits

### Network Stability

Retrograde feedback provides multiple mechanisms for maintaining stable network dynamics:
- **Homeostatic scaling** prevents runaway excitation or silence
- **Competitive learning** ensures balanced representation
- **Activity regulation** maintains optimal firing rates

### Adaptive Learning

The bidirectional communication enables sophisticated learning:
- **Credit assignment**: Post-synaptic neurons can "teach" their inputs about effectiveness
- **Temporal learning**: Precise timing relationships can be learned and maintained
- **Context sensitivity**: Learning can be modulated based on network state

### Biological Realism

Retrograde feedback mechanisms closely mirror real neural network operation:
- **Developmental plasticity**: Activity-dependent refinement of connections
- **Experience-dependent plasticity**: Learning and memory formation
- **Homeostatic maintenance**: Long-term stability and health

## Factory System for Neurogenesis and Synaptogenesis

### Biological Component Creation

The matrix provides factory systems that mirror biological development:

#### **Neurogenesis** 
- **RegisterNeuronType()**: Register biological neuron types (pyramidal, interneuron, etc.)
- **CreateNeuron()**: Execute neurogenesis with complete biological integration
- **Automatic integration**: Spatial registration, chemical wiring, health monitoring
- **3D positioning**: Realistic spatial organization and distance-based connectivity

#### **Synaptogenesis**
- **RegisterSynapseType()**: Register synapse types (excitatory, inhibitory, modulatory)
- **CreateSynapse()**: Form synaptic connections with biological properties
- **Activity-dependent**: Formation based on neural activity and proximity
- **Realistic transmission**: Delays, weights, and plasticity appropriate to synapse type

### Usage Pattern

```go
// Initialize matrix environment
matrix := extracellular.NewExtracellularMatrix(config)
matrix.Start()

// Register neuron factory
matrix.RegisterNeuronType("cortical_pyramidal", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
    neuron := neuron.NewNeuron(id, config.Threshold, 0.95, 10*time.Millisecond, 1.0, 5.0, 0.2)
    neuron.SetReceptors(config.Receptors)
    neuron.SetCallbacks(callbacks)
    return neuron, nil
})

// Create neuron through biological process
neuronConfig := types.NeuronConfig{
    NeuronType: "cortical_pyramidal",
    Position:   types.Position3D{X: 10, Y: 10, Z: 5},
    Receptors:  []types.LigandType{types.LigandGABA, types.LigandGlutamate},
}

neuron, err := matrix.CreateNeuron(neuronConfig)
```

## Usage Patterns

### Basic Threshold Adjustment

Post-synaptic neurons can adjust the excitability of their inputs by sending threshold modification signals when connections are too strong or too weak.

### Release Probability Modulation

Retrograde signals can adjust how readily pre-synaptic neurons release neurotransmitter, providing fine-grained control over connection strength without changing synaptic weights.

### Temporal Coordination

Networks can self-organize their timing through retrograde feedback, with post-synaptic neurons teaching their inputs about optimal timing relationships.

### Competitive Learning

Multiple pre-synaptic neurons competing for the same post-synaptic target can be regulated through retrograde feedback, ensuring that the most effective inputs are strengthened while ineffective ones are weakened.

## Integration with Matrix Architecture

The component-based architecture makes retrograde feedback implementation clean and efficient:

- **No circular dependencies**: Feedback flows through the matrix coordination layer
- **Flexible routing**: Electrical and chemical signals can reach appropriate targets
- **Biological realism**: Multiple signaling modalities mirror real neural networks
- **Performance**: Direct callback mechanisms avoid routing bottlenecks

### Performance Characteristics
- **Sub-100μs detection times** for NMDA-like mechanisms
- **Concurrent processing** with thread-safe coordination
- **Memory-efficient operation** under realistic loads
- **Scalable architecture** supporting thousands of components

## 🚀 Quick Start

### Installation

```bash
go get github.com/SynapticNetworks/temporal-neuron
```

### Basic Living Neuron

```go
package main

import (
    "fmt"
    "time"
    "github.com/SynapticNetworks/temporal-neuron/neuron"
    "github.com/SynapticNetworks/temporal-neuron/types"
)

func main() {
    // Create a living neuron
    n := neuron.NewNeuron("cortical_neuron", 1.0, 0.95, 10*time.Millisecond, 1.0, 5.0, 0.2)
    
    // Start continuous processing - neuron is now ALIVE
    err := n.Start()
    if err != nil {
        panic(err)
    }
    defer n.Stop()
    
    // Send signals
    signal := types.NeuralSignal{
        Value:     1.5,
        Timestamp: time.Now(),
        SourceID:  "test_input",
        TargetID:  n.ID(),
    }
    
    n.Receive(signal)
    time.Sleep(50 * time.Millisecond)
    
    fmt.Printf("Neuron activity level: %.3f\n", n.GetActivityLevel())
}
```

### Matrix-Integrated Network

```go
// Initialize extracellular matrix
matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
    ChemicalEnabled: true,
    SpatialEnabled:  true,
    UpdateInterval:  10 * time.Millisecond,
    MaxComponents:   100,
})

err := matrix.Start()
if err != nil {
    panic(err)
}
defer matrix.Stop()

// Register neuron factory
matrix.RegisterNeuronType("cortical_neuron", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
    n := neuron.NewNeuron(id, config.Threshold, 0.95, 10*time.Millisecond, 1.0, 5.0, 0.2)
    n.SetReceptors(config.Receptors)
    n.SetCallbacks(callbacks)
    return n, nil
})

// Create integrated neuron
neuronConfig := types.NeuronConfig{
    NeuronType: "cortical_neuron",
    Threshold:  1.0,
    Position:   types.Position3D{X: 0, Y: 0, Z: 0},
    Receptors:  []types.LigandType{types.LigandGABA, types.LigandGlutamate},
}

neuron, err := matrix.CreateNeuron(neuronConfig)
if err != nil {
    panic(err)
}
```

## 📦 Core Packages

### `/neuron` - Temporal Neuron Implementation
**Status: ✅ Complete with comprehensive biological validation**

Biologically realistic neurons with homeostatic plasticity, STDP feedback, custom behavior systems, and multi-mode dendritic integration.

### `/synapse` - Synaptic Processors  
**Status: ✅ Complete with biological validation**

Intelligent synapses with STDP plasticity, structural plasticity, realistic delays, and thread-safe operations.

### `/extracellular` - Matrix Coordination System
**Status: ✅ Complete with factory system**

Central coordination layer with astrocyte networks, microglial monitoring, chemical modulation, gap junction management, and neurogenesis/synaptogenesis factories.

### `/component` - Core Interfaces
**Status: ✅ Complete**

Base interfaces and component abstractions for the neural architecture.

### `/types` - Type Definitions  
**Status: ✅ Complete**

Core types, configurations, and message structures.

## Research Applications

### Neural Network Studies
- Activity-dependent plasticity research
- Network synchronization and oscillation studies
- Developmental plasticity investigation
- Pathological pattern recognition and analysis

### Pharmacological Research
- Drug effect simulation on neural circuits
- Substance intoxication modeling with dose-response relationships
- Neurotransmitter interaction studies
- Therapeutic intervention testing

### Computational Neuroscience
- Biological realism validation against experimental data
- Scaling studies for large network simulation
- Performance optimization for real-time applications
- Cross-species comparison studies

### Clinical Applications
- Stroke recovery monitoring and prediction
- Neurodegenerative disease progression modeling
- Neural stimulation effectiveness assessment
- Rehabilitation progress tracking

## Validation and Testing

The system includes comprehensive test suites validating:

- **Biological accuracy**: Matches experimental neuroscience data
- **Performance characteristics**: Real-time simulation capabilities  
- **Integration robustness**: Seamless component interaction
- **Edge case handling**: Graceful failure and recovery
- **Concurrent operation**: Thread-safe multi-component processing

Test coverage includes intoxication modeling, spatial processing, chemical signaling, retrograde feedback mechanisms, and neurogenesis/synaptogenesis.

## Contributing

We welcome contributions from neuroscientists, AI researchers, and systems engineers.

### Development Setup
```bash
git clone https://github.com/SynapticNetworks/temporal-neuron.git
cd temporal-neuron
go mod tidy
go test ./...
```

### Guidelines
- Follow biological realism principles
- Include comprehensive tests for new features
- Maintain performance standards
- Document biological basis and validation
- Validate against published neuroscience research

## License

This project is licensed under the **Temporal Neuron Research License (TNRL) v1.0**

- ✅ Free for research, educational, and personal use
- ✅ Academic institutions and universities - unlimited use  
- ✅ Open source research projects - encouraged
- ⚠️ Commercial use requires permission

## Future Research Directions

This retrograde feedback capability transforms the temporal-neuron system from a simple feedforward network into a sophisticated, self-organizing neural architecture capable of advanced learning, adaptation, and homeostatic regulation—closely mirroring the computational power of biological neural networks.

Current development focuses on extending network-level behaviors and implementing additional biological mechanisms for complex neural circuit modeling.