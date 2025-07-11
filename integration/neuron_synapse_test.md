# Spike-Timing-Dependent Plasticity (STDP) Usage Guide

This guide explains how to set up and use STDP-based learning in the Synaptic Networks temporal neuron system. STDP is a form of Hebbian learning where the timing relationship between pre-synaptic and post-synaptic spikes determines whether the connection strengthens or weakens.

## Core Principles

STDP follows these key biological principles:

1. **LTP (Long-Term Potentiation)**: When a pre-synaptic neuron fires shortly before a post-synaptic neuron (helping to cause it to fire), the connection strengthens
   - Pre-before-Post → Negative deltaT → Increased Weight

2. **LTD (Long-Term Depression)**: When a pre-synaptic neuron fires after a post-synaptic neuron (not contributing to its firing), the connection weakens
   - Post-before-Pre → Positive deltaT → Decreased Weight

## Setting Up STDP

### Step 1: Enable STDP on Post-Synaptic Neurons

STDP must be enabled on the post-synaptic neuron, which is responsible for sending feedback signals to incoming synapses:

```go
// Enable STDP on a neuron with specific parameters
neuron.EnableSTDPFeedback(
    10*time.Millisecond, // Feedback delay - how long after firing to apply STDP
    0.1,                 // Learning rate - how strongly to adjust weights (0.0-1.0)
)

// Verify STDP is enabled
if neuron.IsSTDPFeedbackEnabled() {
    log.Println("STDP is enabled on neuron", neuron.ID())
}

// Disable STDP when no longer needed
neuron.DisableSTDPFeedback()
```

### Step 2: Ensure Synapses Support Spike Timing History

Synapses must keep track of recent pre- and post-synaptic spikes to participate in STDP. The `BasicSynapse` implementation already includes this functionality.

### Step 3: Configure STDP Parameters

STDP parameters can be configured when creating synapses:

```go
// Create STDP configuration with custom parameters
stdpConfig := types.PlasticityConfig{
    Enabled:        true,           // Enable STDP for this synapse
    LearningRate:   0.1,            // Learning rate (0.0-1.0)
    TimeConstant:   20*time.Millisecond, // Controls width of learning window
    WindowSize:     100*time.Millisecond, // Maximum timing difference to consider
    MinWeight:      0.01,           // Minimum weight bound
    MaxWeight:      2.0,            // Maximum weight bound
    AsymmetryRatio: 1.2,            // LTD/LTP strength ratio (>1 means LTD is stronger)
}

// Or use the defaults
stdpConfig := synapse.CreateDefaultSTDPConfig()

// Create a synapse with this configuration
newSynapse := synapse.NewBasicSynapse(
    "my_synapse",
    preNeuron,
    postNeuron,
    stdpConfig,
    synapse.CreateDefaultPruningConfig(),
    0.5, // Initial weight
    0,   // Transmission delay
)
```

## How STDP Works in the System

1. **Pre-Synaptic Spike**: When a pre-synaptic neuron fires, the synapse records the spike time.

2. **Post-Synaptic Spike**: When a post-synaptic neuron fires, it schedules STDP feedback.

3. **STDP Feedback**: After a short delay, the post-synaptic neuron's `STDPSignalingSystem` sends feedback to all incoming synapses.

4. **Timing Analysis**: Each synapse analyzes the timing relationship between its most recent pre- and post-synaptic spikes.

5. **Weight Adjustment**: Based on the timing relationship (deltaT), the synapse adjusts its weight:
   - Negative deltaT (pre-before-post) → Weight increases (LTP)
   - Positive deltaT (post-before-pre) → Weight decreases (LTD)

## Example: Setting Up a Learning Network

Here's how to create a simple network with STDP learning:

```go
// Create matrix
matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
    ChemicalEnabled: true,
    SpatialEnabled:  true,
    UpdateInterval:  10*time.Millisecond,
    MaxComponents:   100,
})
matrix.Start()
defer matrix.Stop()

// Register neuron and synapse types
registerComponents(matrix)

// Create input and output neurons
inputNeuron, _ := matrix.CreateNeuron(types.NeuronConfig{
    NeuronType: "temporal_neuron",
    Threshold:  1.0,
})

outputNeuron, _ := matrix.CreateNeuron(types.NeuronConfig{
    NeuronType: "temporal_neuron",
    Threshold:  1.0,
})

// Enable STDP on the output neuron
if stdpNeuron, ok := outputNeuron.(interface {
    EnableSTDPFeedback(time.Duration, float64)
}); ok {
    stdpNeuron.EnableSTDPFeedback(10*time.Millisecond, 0.1)
}

// Create a synapse with STDP
synapse, _ := matrix.CreateSynapse(types.SynapseConfig{
    SynapseType:    "basic_synapse",
    PresynapticID:  inputNeuron.ID(),
    PostsynapticID: outputNeuron.ID(),
    InitialWeight:  0.5,
})

// Start all components
inputNeuron.Start()
outputNeuron.Start()

// Learning loop
for i := 0; i < 100; i++ {
    // 1. Fire the input neuron
    inputNeuron.Receive(types.NeuralSignal{
        Value:     1.5,
        Timestamp: time.Now(),
        SourceID:  "pattern_input",
        TargetID:  inputNeuron.ID(),
    })
    
    // 2. Wait for signal propagation
    time.Sleep(15*time.Millisecond)
    
    // 3. Fire the output neuron (for paired learning)
    outputNeuron.Receive(types.NeuralSignal{
        Value:     1.5,
        Timestamp: time.Now(),
        SourceID:  "training_signal",
        TargetID:  outputNeuron.ID(),
    })
    
    // 4. Allow time for STDP processing
    time.Sleep(50*time.Millisecond)
    
    // 5. Track weight changes (optional)
    if weightGetter, ok := synapse.(interface{ GetWeight() float64 }); ok {
        weight := weightGetter.GetWeight()
        fmt.Printf("Iteration %d: weight = %.4f\n", i, weight)
    }
}
```

## Advanced Usage

### Homeostasis and STDP

For stable learning, you may want to combine STDP with homeostatic mechanisms:

```go
// Enable both STDP and automatic homeostasis on a neuron
neuron.EnableSTDPFeedback(10*time.Millisecond, 0.1)
neuron.EnableAutoHomeostasis(1*time.Second)  // Check every second
```

### Controlling Pruning with STDP

STDP can naturally lead to pruning of unused connections:

```go
// Enable auto-pruning to remove weak connections
neuron.EnableAutoPruning(5*time.Second)
```

### Neuromodulation and STDP

The system supports neuromodulation effects on STDP:

```go
// Release dopamine to reinforce recent STDP changes
matrix.ReleaseChemical(position, types.LigandDopamine, 1.0)

// Release GABA to inhibit and potentially weaken connections
matrix.ReleaseChemical(position, types.LigandGABA, 1.0)
```

## Troubleshooting

### Common Issues

1. **STDP not affecting weights**: Ensure post-neuron has STDP enabled and check spike timing is within the WindowSize
2. **Weights not changing as expected**: Verify pre-spikes and post-spikes are being recorded with the expected timing relationships
3. **All weights decrease**: Check if post-spikes are consistently occurring before pre-spikes
4. **All weights increase**: Check if pre-spikes are consistently occurring before post-spikes

### Diagnostics

To diagnose STDP issues, you can use the following methods:

```go
// Check STDP status on a neuron
status := neuron.GetProcessingStatus()
stdpStatus := status["stdp_system"].(map[string]interface{})
fmt.Printf("STDP System Status: %+v\n", stdpStatus)

// Get spike history from a synapse
if spikeGetter, ok := synapse.(interface {
    GetPreSpikeTimes() []time.Time
    GetPostSpikeTimes() []time.Time
}); ok {
    preSpikes := spikeGetter.GetPreSpikeTimes()
    postSpikes := spikeGetter.GetPostSpikeTimes()
    
    // Check timing relationships
    for _, preTime := range preSpikes {
        for _, postTime := range postSpikes {
            deltaT := preTime.Sub(postTime)
            fmt.Printf("Spike timing: deltaT = %v\n", deltaT)
        }
    }
}
```

## Implementation Details

1. The `STDPSignalingSystem` (in `stdp_signaling.go`) is responsible for scheduling and delivering STDP feedback signals.
2. The feedback process automatically distinguishes between LTP and LTD based on spike timing.
3. The synapse's `ApplyPlasticity` method performs the actual weight adjustment based on deltaT.
4. The STDP curve is configured with time constants for both LTP and LTD phases.

## Best Practices

1. **Start with conservative learning rates** (0.01-0.1) to avoid instability.
2. **Combine STDP with homeostasis** for more stable learning.
3. **Set appropriate window sizes** based on your network's temporal dynamics.
4. **Use realistic timing** - biological STDP typically works in the range of 1-50ms.
5. **Include delays in synapses** to create realistic timing relationships.
6. **Monitor weight distributions** periodically to ensure learning is progressing.

---

By following these guidelines, you can effectively use STDP for unsupervised learning in your neural networks.