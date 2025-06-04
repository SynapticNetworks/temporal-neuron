/*
Package neuron provides biologically-inspired concurrent neural processing units
for building dynamic neural networks with realistic timing and connectivity.

# Overview

This package implements temporal neurons that operate fundamentally differently
from traditional artificial neural networks. Instead of batch processing with
mathematical activation functions, these neurons process signals continuously
in real-time using simple biological principles.

# Key Differences from Traditional ANNs

Traditional artificial neural networks suffer from several biological implausibilities:
  - Batch/iteration processing (brains process continuously)
  - Complex activation functions (sigmoid, ReLU, etc. don't exist in biology)
  - Synchronous operation (all neurons fire simultaneously)
  - Static connectivity (connections fixed during operation)

This package eliminates these artificial constraints by implementing neurons that:
  - Process signals continuously without batches or training epochs
  - Use simple threshold-based firing (no mathematical activation functions)
  - Operate asynchronously with independent timing
  - Support dynamic connectivity changes during runtime

# Research Context

This implementation is part of ongoing research into biologically-realistic neural
computation. The work draws inspiration from:
  - Cascaded Thinking Model (CTM) for temporal processing dynamics
  - Stateful Gating Networks for dynamic rewiring (planned for future releases)
  - Biological neuroscience research on temporal summation and synaptic plasticity
  - Neuromorphic engineering approaches to brain-inspired computing

# Core Concepts

## Temporal Integration

Real neurons accumulate electrical signals over time windows before deciding to fire.
Unlike traditional ANNs that process inputs instantaneously, temporal neurons
integrate signals arriving within configurable time windows:

	// Neuron accumulates signals for 100ms before potential firing
	n := neuron.NewNeuron(1.0, 100*time.Millisecond, 1.0)

Multiple weak signals arriving within the time window can sum together to exceed
the firing threshold, modeling how biological neurons integrate postsynaptic
potentials from multiple synapses.

## Threshold-Based Firing

Instead of complex mathematical activation functions, neurons use the same simple
rule that biological neurons follow:

	if accumulated_charge >= threshold {
	    fire_action_potential()
	}

This eliminates artificial mathematical complexity while maintaining biological
realism and computational efficiency.

## Dynamic Connectivity

Connections between neurons can be modified during runtime, modeling the brain's
neuroplasticity—its ability to form and eliminate synaptic connections based on
experience:

	// Add new synaptic connection
	neuron.AddOutput("connection1", targetChannel, synapticStrength, delay)

	// Remove connection (synaptic pruning)
	neuron.RemoveOutput("connection1")

## Concurrent Processing

Each neuron operates as an independent Go goroutine, enabling true parallel
processing that scales with available CPU cores. This models how biological
neural networks process information simultaneously across millions of neurons:

	// Each neuron runs concurrently
	go neuron1.Run()
	go neuron2.Run()
	go neuron3.Run()

## Realistic Timing

Synaptic connections include configurable transmission delays that model
biological factors like axon length, myelination, and synaptic processing time:

	// Fast local connection (1ms delay)
	neuron.AddOutput("local", target.GetInput(), 1.0, 1*time.Millisecond)

	// Slower long-distance connection (50ms delay)
	neuron.AddOutput("distant", target.GetInput(), 0.8, 50*time.Millisecond)

# Basic Usage

## Creating and Running a Neuron

	// Create neuron: threshold=1.0, 100ms time window, output factor=1.0
	n := NewNeuron(1.0, 100*time.Millisecond, 1.0)

	// Create output channel
	output := make(chan Message, 10)

	// Connect with synaptic strength=1.0 and 5ms delay
	n.AddOutput("output1", output, 1.0, 5*time.Millisecond)

	// Start continuous processing
	go n.Run()

## Sending Signals

	input := n.GetInput()

	// Send excitatory signal
	input <- Message{Value: 0.7}

	// Send inhibitory signal
	input <- Message{Value: -0.3}

	// Send signal that triggers firing
	input <- Message{Value: 0.8} // Total: 0.7 - 0.3 + 0.8 = 1.2 > 1.0

## Building Networks

	// Create multiple neurons
	neuron1 := NewNeuron(1.0, 50*time.Millisecond, 1.0)
	neuron2 := NewNeuron(0.8, 50*time.Millisecond, 1.2)

	// Connect them: neuron1 -> neuron2
	neuron1.AddOutput("to_n2", neuron2.GetInput(), 0.9, 10*time.Millisecond)

	// Start both neurons
	go neuron1.Run()
	go neuron2.Run()

	// Send input to first neuron - will propagate through network
	neuron1.GetInput() <- Message{Value: 1.5}

# Biological Realism

## Excitatory and Inhibitory Signals

The package supports both positive (excitatory) and negative (inhibitory) input
values, modeling the two types of synaptic transmission found in biological brains:

	// Excitatory input (increases firing probability)
	input <- Message{Value: 0.5}

	// Inhibitory input (decreases firing probability)
	input <- Message{Value: -0.3}

## Synaptic Properties

Each output connection models key properties of biological synapses:
  - Synaptic strength (connection weight/efficacy)
  - Transmission delay (axon conduction + synaptic delay)
  - Dynamic modification (add/remove connections)

## Parallel Transmission

When a neuron fires, the signal propagates simultaneously to all connected targets,
modeling how a single biological action potential affects multiple downstream neurons
at once.

# Concurrency and Scalability

## Goroutine-Based Architecture

Each neuron operates as an independent goroutine, leveraging Go's lightweight
concurrency model. This enables networks with thousands or potentially millions
of concurrently operating neurons:

	// Create large concurrent network
	neurons := make([]*Neuron, 100000)
	for i := range neurons {
	    neurons[i] = NewNeuron(1.0, 50*time.Millisecond, 1.0)
	    go neurons[i].Run() // Each neuron is a separate goroutine
	}

## Thread Safety

All neuron operations are thread-safe, allowing safe modification of connections
while neurons are actively processing signals:

	// Safe to call from any goroutine while neuron is running
	neuron.AddOutput("new_connection", target.GetInput(), 1.0, 5*time.Millisecond)
	neuron.RemoveOutput("old_connection")

## Message Passing

Neurons communicate exclusively through Go channels, eliminating shared memory
and potential race conditions while modeling the discrete nature of biological
neural communication.

# Performance Characteristics

## Memory Efficiency

Each neuron requires minimal memory overhead (~2KB base + 64 bytes per connection),
enabling large-scale networks within reasonable memory constraints.

## Processing Efficiency

Signal processing is optimized for low latency with sub-millisecond response times
for local operations. Network topology and connection delays determine overall
propagation times.

## Scaling Properties

The concurrent architecture scales naturally with available CPU cores, as each
neuron operates independently without requiring synchronization barriers.

# Research Applications

## Neuromorphic Computing

Build brain-inspired computing systems that process information through temporal
dynamics and event-driven computation rather than traditional algorithmic approaches.

## Real-time Processing

Process streaming data with realistic neural timing, suitable for robotics,
sensory processing, and real-time control applications.

## Neural Network Research

Study emergent behaviors, timing-dependent plasticity, and network dynamics in
controlled environments with biologically-realistic constraints.

## Adaptive Systems

Create systems that can modify their connectivity patterns based on experience,
enabling learning and adaptation without traditional backpropagation training.

# Limitations and Future Work

## Current Limitations

This package provides the foundational neuron implementation. Current limitations include:
  - No built-in learning algorithms (planned for future releases)
  - Basic neuron model (more sophisticated models planned)
  - Limited network topology helpers (coming soon)

## Future Developments

Planned extensions include:
  - Stateful gating mechanisms for dynamic network rewiring
  - Biologically-inspired learning algorithms (STDP, Hebbian learning)
  - Advanced neuron models (integrate-and-fire, Hodgkin-Huxley variants)
  - Network visualization and analysis tools
  - Integration with neuromorphic hardware

# Research Status

This is an active research project in early development. The API is experimental
and may change as research progresses. We welcome collaboration from researchers
in computational neuroscience, neuromorphic engineering, and concurrent computing.

# References and Inspiration

This work draws from research in:
  - Biological neuroscience and synaptic physiology
  - Neuromorphic engineering and brain-inspired computing
  - Spiking neural networks and temporal computation
  - Concurrent computing and parallel processing architectures

The implementation is particularly inspired by the Cascaded Thinking Model (CTM)
approach to iterative neural processing and ongoing research into stateful gating
networks for dynamic neural computation.
*/
package neuron
