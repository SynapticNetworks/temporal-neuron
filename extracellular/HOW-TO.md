
# 🧠 How-To: Building a Neural Circuit with the Factory-Enhanced Matrix

This guide demonstrates the new, streamlined workflow for creating and connecting neurons using the factory pattern in the `ExtracellularMatrix`. The new approach is:

- Simpler  
- Less error-prone  
- More biologically realistic  

The matrix itself now handles the complex process of *neural development*.

You define a configuration and ask the matrix to create it.  
A single call to `matrix.CreateNeuron()` or `matrix.CreateSynapse()` handles **all wiring and registration** automatically.

---

## 🧩 Step 1: Initialize the Extracellular Matrix

First, create an instance of the matrix. This acts as your complete biological environment.

```go
package main

import (
	"log"
	"time"
	"extracellular" // Assuming your package is in this path
)

func main() {
	config := extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  10 * time.Millisecond,
		MaxComponents:   1000,
	}

	matrix := extracellular.NewExtracellularMatrix(config)
	log.Println("Extracellular Matrix initialized.")
}
```

---

## 🧠 Step 2: Define and Create Neurons

Use `NeuronConfig` to define neurons. Then call `matrix.CreateNeuron()` to integrate them.

```go
// --- Define Neuron Configurations ---

sensoryNeuronConfig := extracellular.NeuronConfig{
	NeuronType:  "basic",
	Position:    extracellular.Position3D{X: 0, Y: 0, Z: 0},
	Receptors:   []extracellular.LigandType{extracellular.LigandGlutamate},
	SignalTypes: []extracellular.SignalType{extracellular.SignalFired},
}

processingNeuron1Config := extracellular.NeuronConfig{
	NeuronType:  "basic",
	Position:    extracellular.Position3D{X: 50, Y: 20, Z: 0},
	Receptors:   []extracellular.LigandType{extracellular.LigandGlutamate, extracellular.LigandGABA},
	SignalTypes: []extracellular.SignalType{extracellular.SignalFired},
}

processingNeuron2Config := extracellular.NeuronConfig{
	NeuronType:  "basic",
	Position:    extracellular.Position3D{X: 50, Y: -20, Z: 0},
	Receptors:   []extracellular.LigandType{extracellular.LigandGlutamate, extracellular.LigandGABA},
	SignalTypes: []extracellular.SignalType{extracellular.SignalFired},
}

// --- Create Neurons ---

sensoryNeuron, err := matrix.CreateNeuron(sensoryNeuronConfig)
// handle error...

procNeuron1, err := matrix.CreateNeuron(processingNeuron1Config)
// handle error...

procNeuron2, err := matrix.CreateNeuron(processingNeuron2Config)
// handle error...
```

---

## 🔗 Step 3: Define and Create Synapses to Connect Neurons

Each synapse connects a presynaptic to a postsynaptic neuron.

```go
synapse1Config := extracellular.SynapseConfig{
	SynapseType:    "excitatory_plastic",
	PresynapticID:  sensoryNeuron.ID(),
	PostsynapticID: procNeuron1.ID(),
	Position:       extracellular.Position3D{X: 25, Y: 10, Z: 0},
	InitialWeight:  0.8,
	Delay:          1 * time.Millisecond,
}

synapse2Config := extracellular.SynapseConfig{
	SynapseType:    "excitatory_plastic",
	PresynapticID:  sensoryNeuron.ID(),
	PostsynapticID: procNeuron2.ID(),
	Position:       extracellular.Position3D{X: 25, Y: -10, Z: 0},
	InitialWeight:  0.75,
	Delay:          1 * time.Millisecond,
}

// --- Create Synapses ---

synapse1, err := matrix.CreateSynapse(synapse1Config)
// handle error...

synapse2, err := matrix.CreateSynapse(synapse2Config)
// handle error...
```

---

## ⚡ Step 4: Activate the Network

Start the matrix to activate all internal processes:

```go
err = matrix.Start()
if err != nil {
	log.Fatalf("Failed to start the matrix: %v", err)
}
log.Println("Matrix and all neurons are now active and running.")

// Optional cleanup:
defer matrix.Stop()
```

---

## 🧪 Complete Example

```go
package main

import (
	"fmt"
	"log"
	"time"
	"extracellular" // Replace with your actual package path
)

func main() {
	// Setup
	config := extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  10 * time.Millisecond,
		MaxComponents:   1000,
	}
	matrix := extracellular.NewExtracellularMatrix(config)

	// Create Neurons
	sensoryNeuron, _ := matrix.CreateNeuron(extracellular.NeuronConfig{
		NeuronType: "basic",
		Position:   extracellular.Position3D{X: 0, Y: 0, Z: 0},
		Receptors:  []extracellular.LigandType{extracellular.LigandGlutamate},
		SignalTypes: []extracellular.SignalType{extracellular.SignalFired},
	})
	procNeuron1, _ := matrix.CreateNeuron(extracellular.NeuronConfig{
		NeuronType: "basic",
		Position:   extracellular.Position3D{X: 50, Y: 20, Z: 0},
		Receptors:  []extracellular.LigandType{extracellular.LigandGlutamate, extracellular.LigandGABA},
		SignalTypes: []extracellular.SignalType{extracellular.SignalFired},
	})
	procNeuron2, _ := matrix.CreateNeuron(extracellular.NeuronConfig{
		NeuronType: "basic",
		Position:   extracellular.Position3D{X: 50, Y: -20, Z: 0},
		Receptors:  []extracellular.LigandType{extracellular.LigandGlutamate, extracellular.LigandGABA},
		SignalTypes: []extracellular.SignalType{extracellular.SignalFired},
	})

	// Create Synapses
	matrix.CreateSynapse(extracellular.SynapseConfig{
		SynapseType:    "excitatory_plastic",
		PresynapticID:  sensoryNeuron.ID(),
		PostsynapticID: procNeuron1.ID(),
		Position:       extracellular.Position3D{X: 25, Y: 10, Z: 0},
		InitialWeight:  0.8,
		Delay:          1 * time.Millisecond,
	})
	matrix.CreateSynapse(extracellular.SynapseConfig{
		SynapseType:    "excitatory_plastic",
		PresynapticID:  sensoryNeuron.ID(),
		PostsynapticID: procNeuron2.ID(),
		Position:       extracellular.Position3D{X: 25, Y: -10, Z: 0},
		InitialWeight:  0.75,
		Delay:          1 * time.Millisecond,
	})

	// Start the network
	if err := matrix.Start(); err != nil {
		log.Fatalf("Start error: %v", err)
	}

	// Inspect
	fmt.Printf("Total neurons: %d\n", len(matrix.ListNeurons()))
	for _, n := range matrix.ListNeurons() {
		fmt.Printf(" - Neuron: %s\n", n.ID())
	}
	fmt.Printf("Total synapses: %d\n", len(matrix.ListSynapses()))
	for _, s := range matrix.ListSynapses() {
		fmt.Printf(" - Synapse: %s\n", s.ID())
	}
}
```
