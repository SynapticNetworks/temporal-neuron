# Message Package

The **message package** defines the fundamental unit of neural communication in the temporal-neuron simulation system. It provides pure signal content types and structures that enable direct neuron-to-neuron communication through synaptic transmission, without any architectural or routing concerns.

## Table of Contents

- [Overview](#overview)
- [Architecture Philosophy](#architecture-philosophy)
- [Core Message Type](#core-message-type)
- [Chemical Signaling Types](#chemical-signaling-types)
- [Electrical Signaling Types](#electrical-signaling-types)
- [Direct Synaptic Communication](#direct-synaptic-communication)
- [Signal Flow Examples](#signal-flow-examples)
- [Integration Patterns](#integration-patterns)
- [Usage Guidelines](#usage-guidelines)
- [API Reference](#api-reference)

## Overview

The message package serves as the **signal content foundation** for neural communication. It defines:

- **Pure signal data** - Value, timing, and biological properties
- **Chemical messenger types** - Neurotransmitters and signaling molecules
- **Electrical signal types** - Gap junction and coordination signals
- **Signal quality metrics** - Reliability, noise, and transmission success
- **Biological metadata** - Calcium levels, vesicle release, receptor information

### What Messages Contain

Messages contain **only signal content and biological properties**:
- Signal strength and timing information
- Chemical properties (neurotransmitter type, vesicle data)
- Transmission characteristics (delays, reliability, noise)
- Source and target identification for routing
- Biological state data (calcium levels, receptor types)

### What Messages Do NOT Contain

Messages do **NOT** contain:
- **Architectural information** - Component types, states, lifecycle data
- **Learning configuration** - Plasticity types, learning parameters
- **Spatial calculations** - Distance measurements, range detection
- **Routing logic** - Path determination, connection management
- **System coordination** - Network topology, global state

## Architecture Philosophy

### Pure Signal Content

The message package follows the principle of **pure signal content**:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│    MESSAGES     │    │   COMPONENTS    │    │     MATRIX      │
│                 │    │                 │    │                 │
│ • Signal Data   │    │ • Architecture  │    │ • Coordination  │
│ • Timing Info   │    │ • Interfaces    │    │ • Routing       │
│ • Chemical Props│    │ • Lifecycle     │    │ • Spatial Calc  │
│ • Biological    │    │ • State Mgmt    │    │ • Networking    │
│   State         │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Biological Inspiration

Messages model **biological neural signals**:
- **Action potentials** - Electrical spikes with timing and amplitude
- **Neurotransmitter release** - Chemical signals with concentration and type
- **Synaptic transmission** - Signal processing with delays and reliability
- **Signal quality** - Noise, attenuation, and transmission success
- **Biological context** - Calcium levels, vesicle state, receptor binding

### Direct Communication

Messages enable **direct neuron-to-synapse-to-neuron communication**:

```
Neuron A → Synapse → Neuron B
    ↓         ↓         ↓
  Fires    Processes  Receives
  Signal   & Learns   Signal
```

No routing through matrix - signals flow directly between neural components.

## Core Message Type

### NeuralSignal Structure

The `NeuralSignal` is the fundamental unit of neural communication:

```go
type NeuralSignal struct {
    // === CORE SIGNAL PROPERTIES ===
    Value         float64 `json:"value"`          // Final signal strength
    OriginalValue float64 `json:"original_value"` // Pre-processing strength

    // === TIMING INFORMATION ===
    Timestamp     time.Time     `json:"timestamp"`       // Signal initiation time
    SynapticDelay time.Duration `json:"synaptic_delay"`  // Synaptic processing delay
    SpatialDelay  time.Duration `json:"spatial_delay"`   // Axonal conduction delay
    TotalDelay    time.Duration `json:"total_delay"`     // Combined delay

    // === ROUTING INFORMATION ===
    SourceID  string `json:"source_id"`  // Originating component
    TargetID  string `json:"target_id"`  // Destination component
    SynapseID string `json:"synapse_id"` // Processing synapse (if applicable)

    // === CHEMICAL SIGNAL CONTENT ===
    NeurotransmitterType LigandType `json:"neurotransmitter_type"` // Chemical messenger
    VesicleReleased      bool       `json:"vesicle_released"`      // Vesicle consumption
    CalciumLevel         float64    `json:"calcium_level"`         // Presynaptic calcium

    // === SIGNAL QUALITY ===
    TransmissionSuccess bool    `json:"transmission_success"` // Success indicator
    FailureReason       string  `json:"failure_reason"`       // Error description
    NoiseLevel          float64 `json:"noise_level"`          // Background noise

    // === METADATA ===
    Metadata map[string]interface{} `json:"metadata"` // Additional data
}
```

### Key Fields Explained

| Field | Purpose | Biological Basis |
|-------|---------|------------------|
| `Value` | Final signal strength reaching target | Postsynaptic potential amplitude |
| `OriginalValue` | Signal before synaptic processing | Presynaptic action potential strength |
| `Timestamp` | When signal was initiated | Action potential timing |
| `SynapticDelay` | Synapse processing time | Vesicle fusion + diffusion time |
| `SpatialDelay` | Axon conduction time | Distance/conduction velocity |
| `NeurotransmitterType` | Chemical messenger | Glutamate, GABA, dopamine, etc. |
| `VesicleReleased` | Whether vesicle was consumed | Vesicle pool dynamics |
| `CalciumLevel` | Presynaptic calcium concentration | Calcium-dependent release |

## Chemical Signaling Types

### LigandType Enum

Chemical messengers used in neural communication:

```go
type LigandType int

const (
    LigandNone            LigandType = iota // No chemical signal
    LigandGlutamate                         // Primary excitatory
    LigandGABA                              // Primary inhibitory
    LigandDopamine                          // Reward/motor control
    LigandSerotonin                         // Mood/behavioral state
    LigandAcetylcholine                     // Attention/autonomic
    LigandNorepinephrine                    // Attention/arousal
    LigandHistamine                         // Arousal/inflammation
    LigandGlycine                           // Inhibitory (spinal cord)
    LigandAdenosine                         // Sleep/neuroprotection
    LigandNitricOxide                       // Retrograde signaling
    LigandEndocannabinoid                   // Retrograde messenger
    LigandNeuropeptideY                     // Feeding/anxiety
    LigandSubstanceP                        // Pain/inflammation
    LigandVasopressin                       // Social behavior
    LigandOxytocin                          // Social bonding
)
```

### Chemical Properties

Each ligand type has biological properties:

```go
// Get typical effect: +1 (excitatory), -1 (inhibitory), 0 (modulatory)
effect := LigandGlutamate.GetPolarityEffect() // Returns 1.0
gaba := LigandGABA.GetPolarityEffect()        // Returns -1.0
dopa := LigandDopamine.GetPolarityEffect()    // Returns 0.0 (modulatory)
```

### Major Neurotransmitter Classes

| Class | Examples | Function | Typical Effect |
|-------|----------|----------|----------------|
| **Fast Excitatory** | Glutamate | Rapid excitation | +1.0 |
| **Fast Inhibitory** | GABA, Glycine | Rapid inhibition | -1.0 |
| **Modulatory** | Dopamine, Serotonin | Behavioral state | 0.0 |
| **Autonomic** | Acetylcholine, Norepinephrine | System control | Variable |
| **Neuropeptides** | Substance P, Oxytocin | Long-term effects | 0.0 |
| **Retrograde** | Nitric Oxide, Endocannabinoids | Feedback signaling | 0.0 |

## Electrical Signaling Types

### SignalType Enum

Electrical and coordination signals for network communication:

```go
type SignalType int

const (
    SignalNone               SignalType = iota // No electrical signal
    SignalFired                                // Action potential occurred
    SignalConnected                            // New connection established
    SignalDisconnected                         // Connection removed
    SignalThresholdChanged                     // Firing threshold adjusted
    SignalSynchronization                      // Network sync pulse
    SignalCalciumWave                          // Glial calcium wave
    SignalPlasticityEvent                      // Learning occurred
    SignalChemicalGradient                     // Chemical gradient change
    SignalStructuralChange                     // Physical modification
    SignalMetabolicState                       // Energy state change
    SignalHealthWarning                        // Component health issue
    SignalNetworkOscillation                   // Network rhythm
)
```

### Signal Categories

| Category | Signals | Purpose |
|----------|---------|---------|
| **Action Potentials** | `SignalFired` | Direct electrical communication |
| **Connectivity** | `SignalConnected`, `SignalDisconnected` | Network topology changes |
| **Plasticity** | `SignalThresholdChanged`, `SignalPlasticityEvent` | Learning and adaptation |
| **Coordination** | `SignalSynchronization`, `SignalNetworkOscillation` | Network-wide timing |
| **Glial Activity** | `SignalCalciumWave`, `SignalChemicalGradient` | Support cell coordination |
| **Health Monitoring** | `SignalHealthWarning`, `SignalMetabolicState` | System maintenance |

## Direct Synaptic Communication

### Signal Flow Architecture

The message package enables **direct neural communication** without matrix routing:

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Neuron A  │────▶│   Synapse   │────▶│   Neuron B  │
│             │     │             │     │             │
│ • Fires     │     │ • Processes │     │ • Receives  │
│ • Creates   │     │ • Learns    │     │ • Integrates│
│   Signal    │     │ • Transmits │     │ • Responds  │
└─────────────┘     └─────────────┘     └─────────────┘
```

### Communication Steps

1. **Signal Generation**: Neuron A fires and creates `NeuralSignal`
2. **Synaptic Processing**: Synapse receives signal, applies weight/plasticity
3. **Direct Delivery**: Synapse sends processed signal directly to Neuron B
4. **Signal Integration**: Neuron B receives and integrates the signal

### No Matrix Routing

Signals flow **directly between components**:
- ✅ **Direct paths**: Neuron → Synapse → Neuron
- ✅ **No bottlenecks**: No central routing or queuing
- ✅ **Realistic timing**: Biological delays preserved
- ✅ **Parallel processing**: Multiple signals simultaneously
- ❌ **No matrix routing**: Matrix provides coordination, not routing

## Signal Flow Examples

### Basic Excitatory Transmission

```go
// Neuron A fires and creates signal
signal := message.NeuralSignal{
    Value:                1.0,                    // Strong signal
    OriginalValue:        1.0,                    // Before processing
    Timestamp:            time.Now(),             // When fired
    SourceID:             "neuron-A",             // Source neuron
    TargetID:             "neuron-B",             // Target neuron
    SynapseID:            "synapse-A-B",          // Processing synapse
    NeurotransmitterType: message.LigandGlutamate, // Excitatory
    VesicleReleased:      true,                   // Vesicle consumed
    CalciumLevel:         2.5,                    // High calcium
    SynapticDelay:        2 * time.Millisecond,   // Synapse processing
    SpatialDelay:         5 * time.Millisecond,   // Axon conduction
    TotalDelay:           7 * time.Millisecond,   // Combined delay
    TransmissionSuccess:  true,                   // Successful transmission
    NoiseLevel:           0.1,                    // Low noise
}

// Synapse processes the signal
processedSignal := synapse.Process(signal)

// Direct delivery to target neuron
targetNeuron.Receive(processedSignal)
```

### Inhibitory Feedback

```go
// Inhibitory interneuron fires
inhibSignal := message.NeuralSignal{
    Value:                -0.8,                  // Inhibitory signal
    OriginalValue:        1.0,                   // Original spike
    NeurotransmitterType: message.LigandGABA,    // Inhibitory neurotransmitter
    VesicleReleased:      true,                  // GABA vesicle released
    // ... other fields
}

// Fast inhibitory synapse (minimal delay)
inhibSignal.SynapticDelay = 1 * time.Millisecond
inhibSignal.SpatialDelay = 1 * time.Millisecond

// Direct inhibitory effect
pyramidalNeuron.Receive(inhibSignal)
```

### Neuromodulatory Signaling

```go
// Dopaminergic neuron releases neuromodulator
modulatorSignal := message.NeuralSignal{
    Value:                0.5,                      // Moderate signal
    NeurotransmitterType: message.LigandDopamine,   // Modulatory
    VesicleReleased:      true,                     // Dopamine released
    SynapticDelay:        10 * time.Millisecond,    // Slower processing
    // ... neuromodulatory effects
}

// Affects multiple targets through volume transmission
for _, target := range modulatoryTargets {
    target.Receive(modulatorSignal)
}
```

### Signal Quality Variations

```go
// High-quality transmission
reliableSignal := message.NeuralSignal{
    Value:               1.0,
    TransmissionSuccess: true,
    FailureReason:       "",
    NoiseLevel:          0.05,  // Very low noise
}

// Degraded transmission
noisySignal := message.NeuralSignal{
    Value:               0.7,    // Attenuated
    TransmissionSuccess: true,
    FailureReason:       "",
    NoiseLevel:          0.3,    // High noise
}

// Failed transmission
failedSignal := message.NeuralSignal{
    Value:               0.0,
    TransmissionSuccess: false,
    FailureReason:       "vesicle depletion",
    NoiseLevel:          0.2,
}
```

## Integration Patterns

### Neuron Signal Generation

```go
// In neuron firing logic
func (n *Neuron) Fire() {
    outputValue := n.accumulator * n.fireFactor
    
    for synapseID, synapse := range n.outputSynapses {
        signal := message.NeuralSignal{
            Value:                outputValue,
            OriginalValue:        outputValue,
            Timestamp:            time.Now(),
            SourceID:             n.ID(),
            TargetID:             synapse.GetTargetID(),
            SynapseID:            synapseID,
            NeurotransmitterType: n.getPrimaryNeurotransmitter(),
            VesicleReleased:      true,
            CalciumLevel:         n.getCalciumLevel(),
        }
        
        // Direct call to synapse
        synapse.Transmit(signal)
    }
}
```

### Synaptic Processing

```go
// In synapse transmission logic
func (s *Synapse) Transmit(signal message.NeuralSignal) error {
    // Apply synaptic weight
    processedValue := signal.Value * s.weight
    
    // Create processed signal
    processed := signal
    processed.Value = processedValue
    processed.SynapticDelay = s.delay
    processed.TotalDelay = signal.SpatialDelay + s.delay
    
    // Apply biological delays and deliver directly
    time.AfterFunc(processed.TotalDelay, func() {
        s.targetNeuron.Receive(processed)
    })
    
    return nil
}
```

### Neural Signal Reception

```go
// In neuron signal reception
func (n *Neuron) Receive(signal message.NeuralSignal) {
    // Queue signal for processing
    select {
    case n.inputBuffer <- signal:
        // Successfully queued
    default:
        // Buffer full - signal lost (biologically realistic)
    }
}

func (n *Neuron) processSignal(signal message.NeuralSignal) {
    // Integrate signal into membrane potential
    n.accumulator += signal.Value
    
    // Record neurotransmitter effect
    if signal.VesicleReleased {
        n.applyNeurotransmitterEffect(signal.NeurotransmitterType, signal.Value)
    }
    
    // Check for firing threshold
    if n.accumulator >= n.threshold {
        n.Fire()
    }
}
```

## Usage Guidelines

### When to Use Messages

Use `NeuralSignal` for:
- ✅ **Synaptic transmission** - Neuron-to-neuron communication
- ✅ **Signal processing** - Carrying signal data through network
- ✅ **Learning algorithms** - Plasticity calculations need signal timing
- ✅ **Activity monitoring** - Tracking neural communication patterns
- ✅ **Biological simulation** - Realistic neurotransmitter dynamics

### When NOT to Use Messages

Don't use `NeuralSignal` for:
- ❌ **Component management** - Use component interfaces instead
- ❌ **Configuration data** - Use separate config structures
- ❌ **State synchronization** - Use direct method calls
- ❌ **Bulk data transfer** - Use appropriate data structures
- ❌ **Non-neural communication** - Use other communication patterns

### Best Practices

1. **Keep signals focused** - Only include signal-relevant data
2. **Use appropriate types** - Choose correct `LigandType` and `SignalType`
3. **Set realistic delays** - Use biologically plausible timing
4. **Handle failures gracefully** - Check `TransmissionSuccess` flag
5. **Preserve signal integrity** - Don't modify signals in transit
6. **Document signal flow** - Clearly document signal pathways

### Performance Considerations

- **Signal creation is lightweight** - Simple struct allocation
- **Timing is critical** - Use precise timestamps and delays
- **Memory efficient** - No large embedded objects
- **Concurrent safe** - Signals are immutable once created
- **Garbage collection friendly** - No circular references

## API Reference

### Core Types

| Type | Description |
|------|-------------|
| `NeuralSignal` | Primary neural communication message |
| `LigandType` | Chemical neurotransmitter types |
| `SignalType` | Electrical and coordination signal types |

### NeuralSignal Fields

| Field | Type | Description |
|-------|------|-------------|
| `Value` | `float64` | Final signal strength |
| `OriginalValue` | `float64` | Pre-processing signal strength |
| `Timestamp` | `time.Time` | Signal initiation time |
| `SynapticDelay` | `time.Duration` | Synaptic processing delay |
| `SpatialDelay` | `time.Duration` | Axonal conduction delay |
| `TotalDelay` | `time.Duration` | Combined transmission delay |
| `SourceID` | `string` | Originating component ID |
| `TargetID` | `string` | Destination component ID |
| `SynapseID` | `string` | Processing synapse ID |
| `NeurotransmitterType` | `LigandType` | Chemical messenger type |
| `VesicleReleased` | `bool` | Whether vesicle was consumed |
| `CalciumLevel` | `float64` | Presynaptic calcium level |
| `TransmissionSuccess` | `bool` | Success indicator |
| `FailureReason` | `string` | Error description |
| `NoiseLevel` | `float64` | Background noise level |
| `Metadata` | `map[string]interface{}` | Additional signal data |

### LigandType Methods

| Method | Returns | Description |
|--------|---------|-------------|
| `String()` | `string` | Human-readable ligand name |
| `GetPolarityEffect()` | `float64` | Typical effect (+1, -1, or 0) |

### SignalType Methods

| Method | Returns | Description |
|--------|---------|-------------|
| `String()` | `string` | Human-readable signal name |

### Usage Patterns

```go
// Create basic neural signal
signal := message.NeuralSignal{
    Value:     1.0,
    SourceID:  "neuron-1",
    TargetID:  "neuron-2",
    Timestamp: time.Now(),
}

// Check neurotransmitter properties
if signal.NeurotransmitterType.GetPolarityEffect() > 0 {
    // Excitatory signal
}

// Handle transmission failures
if !signal.TransmissionSuccess {
    log.Printf("Signal failed: %s", signal.FailureReason)
}

// Access signal timing
totalTime := signal.SynapticDelay + signal.SpatialDelay
```

---

## Contributing

When extending the message package:

1. **Maintain signal purity** - Only add signal-relevant fields
2. **Preserve biological accuracy** - Base new types on neuroscience
3. **Keep it simple** - Messages should be lightweight and focused
4. **Document biology** - Explain the biological basis for new fields
5. **Maintain compatibility** - Don't break existing signal flows

## License

This message package is part of the temporal-neuron project and follows the same licensing terms.