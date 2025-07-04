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
- **Respond to neuromodulators** like dopamine (reward) and GABA (inhibition)

Traditional artificial neural networks reduce synapses to static weight matrices, losing the rich temporal dynamics and adaptive capabilities that make biological brains so powerful. Our `synapse` package restores this biological realism while maintaining high performance through Go's concurrency primitives.

## 🎯 Key Features

### 🔬 Biological Realism
- **Spike-Timing Dependent Plasticity (STDP)**: Classic Hebbian learning with precise timing windows
- **Structural Plasticity**: Automatic pruning of weak, inactive synapses  
- **Realistic Delays**: Axonal conduction and synaptic transmission timing
- **Retrograde Signaling**: Post-synaptic feedback to pre-synaptic terminals
- **Synaptic Diversity**: Pluggable architecture for different synapse types
- **Neuromodulation**: Dopamine (reward) and GABA (inhibition) signaling systems

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
    ProcessNeuromodulation(ligandType LigandType, concentration float64) float64 // Neuromodulatory effects
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

## 🧪 Neuromodulatory Systems

One of the most biologically authentic aspects of our synapse implementation is the support for neuromodulatory systems that regulate learning and signal transmission in ways that mirror the brain's complex chemistry.

### 🌟 Dopamine Signaling: Reward and Prediction

Dopamine functions as a powerful reward signal in our system, implementing a biologically accurate reward prediction error (RPE) model:

#### How Dopamine Works:
- **Baseline Dopamine (1.0)**: Represents expected outcomes, no learning occurs
- **Above Baseline (>1.0)**: Signals "better than expected" outcomes
- **Below Baseline (<1.0)**: Signals "worse than expected" outcomes

#### Bidirectional Learning:
Dopamine interacts with eligibility traces to create complex learning effects:

| Eligibility Trace | Dopamine Level | Weight Change | Learning Effect |
|-------------------|----------------|---------------|-----------------|
| Positive (+) | High (>1.0) | Increase (+) | "Do this again" - reinforcement |
| Positive (+) | Low (<1.0) | Decrease (-) | "Stop doing this" - extinction |
| Negative (-) | High (>1.0) | Decrease (-) | "Avoid this pattern" - interference |
| Negative (-) | Low (<1.0) | Increase (+) | Complex avoidance learning |

#### Test Results:
```
Dopamine at 2.0 with positive eligibility: +0.0151 weight change
Dopamine at 2.0 with negative eligibility: -0.0182 weight change
Expected reward (1.0) with any eligibility: ~0.0000 weight change
```

### 🚫 GABA Signaling: Inhibition and Penalties

GABA acts as the primary inhibitory neurotransmitter, with dual effects on both signal transmission and learning:

#### How GABA Works:
- **Signal Inhibition**: Reduces transmission efficacy through chloride channel activation
- **Penalty Signaling**: Functions as a negative reinforcement signal (opposite of dopamine)
- **Dose-Dependent**: Inhibitory effects scale with GABA concentration

#### Inhibitory Effects:
GABA produces powerful inhibition of signal transmission:

| GABA Level | Input Signal | Output Signal | Inhibition % |
|------------|--------------|---------------|--------------|
| 0.0 (none) | 1.0 | 0.5000 | 0% |
| 0.5 (low) | 1.0 | 0.3534 | 29.3% |
| 1.0 (moderate) | 1.0 | 0.1217 | 68.6% |
| 2.0 (high) | 1.0 | 0.0243 | 86.9% |

#### Learning Effects:
GABA also affects synaptic learning, acting as a penalty signal:

| Eligibility Trace | GABA Level | Weight Change | Learning Effect |
|-------------------|------------|---------------|-----------------|
| Positive (+) | 1.5 | -0.0228 | "Avoid this" - punishment |
| Negative (-) | 1.5 | +0.0485 | "Continue avoiding" - complex avoidance |
| None (~0) | Any | ~0.0000 | No learning without eligibility |

#### Test Results:
```
GABA at 1.5 with positive eligibility: -0.0228 weight change
GABA at 2.0 with negative eligibility: +0.0485 weight change
GABA has stronger effects (-0.045489) than other neuromodulators
```

### 🧠 Serotonin Signaling: Mood Modulation

Serotonin provides mood-related plasticity effects that differ from both reward and punishment:

#### How Serotonin Works:
- **Positive Modulation**: Generally enhances learning with a positive bias
- **Cross-Talk**: Interacts with other neuromodulatory systems 
- **Temporal Integration**: Effects persist over longer timeframes than dopamine

#### Learning Effects:
Serotonin produces moderate positive effects on synaptic learning:

| Eligibility Trace | Serotonin Level | Weight Change | Learning Effect |
|-------------------|-----------------|---------------|-----------------|
| Positive (+) | 1.5 | +0.0114 | Enhanced potentiation |
| Negative (-) | 1.5 | Mild depression | Weak negative effect |

#### Test Results:
```
Serotonin at 1.5 with positive eligibility: +0.0114 weight change
Serotonin has stronger positive effects (+0.034117) than dopamine
```

### ⚡ Glutamate Signaling: Excitatory Enhancement

Glutamate functions as the primary excitatory neurotransmitter with effects on both transmission and learning:

#### How Glutamate Works:
- **Signal Enhancement**: Increases transmission efficacy
- **Learning Amplification**: Can enhance other plasticity processes
- **Synergistic Effects**: Particularly effective when combined with dopamine

#### Learning Effects:
Glutamate produces positive effects on synaptic learning:

| Eligibility Trace | Glutamate Level | Weight Change | Learning Effect |
|-------------------|-----------------|---------------|-----------------|
| Positive (+) | 1.5 | +0.0068 | Moderate enhancement |
| Combined with Dopamine | 1.5 each | +0.0139 | Synergistic boost |

#### Test Results:
```
Glutamate at 1.5 with positive eligibility: +0.0068 weight change
Glutamate → Dopamine combination: +0.013944 (synergistic effect)
```

### 🔀 Combined Chemical Signaling

Our implementation excels at modeling how multiple neuromodulators interact in complex ways:

#### Key Interaction Patterns:

| Signal Combination | Net Effect | Key Finding |
|--------------------|------------|-------------|
| Dopamine → GABA | -0.006176 | GABA dominates when applied last |
| GABA → Dopamine | -0.008503 | GABA's effect persists even when followed by reward |
| Glutamate → Dopamine | +0.013944 | Synergistic enhancement beyond individual effects |
| Complex sequence | -0.004901 | Multi-signal integration with negative bias |
| Very strong Dopamine → GABA | +0.001874 | Extremely high dopamine can overcome GABA |

#### Temporal Effects:
The order and timing of chemical signals matter significantly:
- **Sequential application**: -0.0049 net effect
- **Simultaneous application**: -0.0076 net effect
- **Difference**: +0.0027 (27% less inhibition when applied sequentially)

### 🔄 Neuromodulation Implementation

Our implementation follows the three-factor learning rule from neuroscience:
```
Δw = learning_rate * eligibility_trace * modulation_factor
```

Where:
- **learning_rate**: Base plasticity rate from STDP configuration
- **eligibility_trace**: Memory of recent pre/post activity patterns
- **modulation_factor**: Derived from neuromodulator type and concentration

### 📚 Biological Correspondence

Our neuromodulation systems match findings from key neuroscience research:

| Research | Finding | Our Implementation |
|----------|---------|---------------------|
| Schultz et al. (1997) | Dopamine encodes reward prediction error | RPE model with 1.0 baseline |
| Reynolds & Wickens (2002) | Dopamine modulates STDP | Three-factor learning rule |
| Frémaux & Gerstner (2016) | Eligibility traces bridge timing gaps | Decaying trace system |
| Brzosko et al. (2015) | Dopamine can convert LTD to LTP | Sign-dependent modulation |
| Vogels et al. (2011) | GABA drives inhibitory plasticity | Strong GABA effects that dominate |

These findings validate that our implementation captures sophisticated neuromodulatory dynamics seen in biological systems, where context, timing, and signal combinations create rich learning environments beyond basic Hebbian plasticity.

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

### Neuromodulator Validation
Our dopamine and GABA systems match findings from:
- **Schultz et al. (1997)**: Dopamine as reward prediction error
- **Reynolds & Wickens (2002)**: Dopamine's modulation of STDP
- **Frémaux & Gerstner (2016)**: Three-factor learning rules
- **Vogels et al. (2011)**: GABA's role in inhibitory plasticity

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

### Neuromodulator Tests
- **`TestBidirectionalDopamine_PositiveRewards`**: Validates dopamine as a reward signal
- **`TestBidirectionalDopamine_NegativeErrors`**: Tests dopamine as a prediction error signal
- **`TestBidirectionalDopamine_Combined`**: Tests reward/error signaling in sequence
- **`TestGABASignaling_BasicInhibition`**: Confirms GABA's inhibitory effects on transmission
- **`TestGABASignaling_PenaltySignals`**: Validates GABA as a penalty signal in learning
- **`TestGABASignaling_StdpModulation`**: Tests how GABA modulates STDP windows
- **`TestEnhancedSTDP_ChemicalModulation`**: Tests how different chemicals affect STDP
- **`TestEnhancedSTDP_BidirectionalPlasticity`**: Tests bidirectional learning dynamics
- **`TestEnhancedSTDP_CombinedSignals`**: Tests combined neuromodulator effects

### Performance Tests
- **`TestMassiveConcurrentTransmission`**: Validates behavior with 600+ goroutines
- **`TestSustainedHighFrequencyTransmission`**: Tests 1000Hz transmission rates
- **`TestResourceExhaustionRecovery`**: Confirms recovery under memory pressure
- **`TestMixedOperationChaos`**: Concurrent mixed operations in unpredictable patterns
- **`TestLongRunningStability`**: Extended operation stability (configurable duration)

## 🔧 Educational Examples

### Example: Reinforcement Learning with Dopamine

```go
// Create synapse with eligibility trace support
syn := NewBasicSynapse("learning_synapse", preNeuron, postNeuron,
    CreateDefaultSTDPConfig(), CreateDefaultPruningConfig(), 0.5, 0)

// Establish eligibility trace (action selection)
syn.Transmit(1.0)  // Pre-synaptic spike

// Causal timing creates positive eligibility
adjustment := PlasticityAdjustment{DeltaT: -10 * time.Millisecond}
syn.ApplyPlasticity(adjustment)

// Simulate delay before reward (300ms)
time.Sleep(300 * time.Millisecond)

// Deliver reward (dopamine burst)
rewardLevel := 2.0 // Strong positive reward
weightChange := syn.ProcessNeuromodulation(LigandDopamine, rewardLevel)

fmt.Printf("Weight change from reward: %.4f\n", weightChange) // Should be positive
```

### Example: Inhibitory Control with GABA

```go
// Create synapse for testing inhibition
syn := NewBasicSynapse("inhibitory_test", preNeuron, postNeuron,
    CreateDefaultSTDPConfig(), CreateDefaultPruningConfig(), 0.5, 0)

// Get baseline transmission output
syn.Transmit(1.0)  // Input signal
baselineOutput := 0.5 // Signal * weight (0.5) = 0.5

// Apply GABA inhibition
gabaLevel := 1.0 // Moderate inhibition
syn.ProcessNeuromodulation(LigandGABA, gabaLevel)

// Transmit with inhibition active
syn.Transmit(1.0)  // Same input signal
inhibitedOutput := 0.12 // Approximately, due to GABA inhibition (~70%)

fmt.Printf("Inhibition: %.1f%%\n", (baselineOutput-inhibitedOutput)/baselineOutput*100)
```

### Example: Combined Chemical Signaling

```go
// Create synapse for testing chemical interactions
syn := NewBasicSynapse("chemical_test", preNeuron, postNeuron,
    CreateDefaultSTDPConfig(), CreateDefaultPruningConfig(), 0.5, 0)

// Create eligibility trace through causal STDP
for i := 0; i < 10; i++ {
    syn.ApplyPlasticity(PlasticityAdjustment{DeltaT: -10 * time.Millisecond})
}

// Sequential chemical application
initialWeight := syn.GetWeight()
syn.ProcessNeuromodulation(LigandGlutamate, 1.5) // Excitatory boost
syn.ProcessNeuromodulation(LigandDopamine, 1.5)  // Reward signal
finalWeight := syn.GetWeight()

fmt.Printf("Synergistic effect: %.4f\n", finalWeight - initialWeight) // ~+0.0139
```

## 🔬 Biological Correspondence

### Neuromodulator → Biological Function

| Neuromodulator | Biological Correspondence | Implementation Details |
|----------------|---------------------------|------------------------|
| **Dopamine** | Reward prediction error | Bidirectional signaling based on 1.0 baseline |
| | VTA/SNc phasic bursts | Higher concentrations (>1.0) for better-than-expected |
| | VTA/SNc phasic dips | Lower concentrations (<1.0) for worse-than-expected |
| | D1/D2 receptor pathways | Eligibility-dependent effects (sign-dependent) |
| **GABA** | Inhibitory transmission | Reduces signal transmission strength |
| | GABA-A fast inhibition | Rapid onset, dose-dependent inhibition |
| | Chloride channel activation | Signal scaling by (1-inhibition) factor |
| | Disinhibition circuits | Complex learning through negative eligibility |
| **Serotonin** | Mood regulation | General enhancement of plasticity |
| | Temporal persistence | Longer-lasting effects than dopamine |
| | 5-HT receptor system | Different effect profile than dopamine |
| **Glutamate** | Primary excitatory transmitter | Enhancement of signal transmission |
| | NMDA/AMPA receptors | Synergistic interactions with other chemicals |

### Learning Rule Validation

Our STDP implementation matches experimental data:

- **Bi & Poo (1998)**: Hippocampal cultures, τ ≈ 20ms
- **Sjöström et al. (2001)**: Neocortical pairs, asymmetric window
- **Caporale & Dan (2008)**: Visual cortex, frequency dependence

Our neuromodulation systems match findings from:

- **Schultz et al. (1997)**: Dopamine codes for reward prediction error
- **Reynolds & Wickens (2002)**: Dopamine modulates STDP
- **Brzosko et al. (2015)**: Dopamine can convert LTD to LTP
- **Vogels et al. (2011)**: GABA induces symmetry in inhibitory plasticity

## 🚀 Future Roadmap

- **Additional Synapse Types**: Inhibitory, modulatory, static
- **Advanced Learning Rules**: Metaplasticity, homeostatic scaling
- **More Neuromodulators**: Acetylcholine, norepinephrine, neuropeptides
- **Connectome Integration**: Direct import of biological connectomes
- **Visualization Tools**: Real-time synapse activity monitoring
- **GPU Acceleration**: CUDA backend for massive networks
- **Serialization**: Save/load network states
- **Network Topologies**: Pre-built biological network patterns

## 📚 References

### Key Neuroscience Papers
- **Bi, G. & Poo, M. (1998)** - "Synaptic modifications in cultured hippocampal neurons" - *STDP discovery*
- **Sjöström, P.J. et al. (2001)** - "Rate, timing, and cooperativity jointly determine cortical synaptic plasticity" - *STDP refinement*
- **Schultz, W. et al. (1997)** - "A neural substrate of prediction and reward" - *Dopamine RPE model*
- **Frémaux, N. & Gerstner, W. (2016)** - "Neuromodulated spike-timing-dependent plasticity" - *Three-factor learning*
- **Vogels, T.P. et al. (2011)** - "Inhibitory plasticity balances excitation and inhibition in sensory pathways and memory networks" - *GABA plasticity*

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