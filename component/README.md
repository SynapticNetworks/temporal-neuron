# Component Package

The **component package** provides the foundational architecture for building biologically-inspired neural networks. It defines the core interfaces, types, and base implementations that all neural components (neurons, synapses, glial cells) must implement to participate in coordinated neural computation.

## Table of Contents

- [Overview](#overview)
- [Architecture Philosophy](#architecture-philosophy)
- [Component Types](#component-types)
- [Core Interfaces](#core-interfaces)
- [Specialized Interfaces](#specialized-interfaces)
- [Base Implementation](#base-implementation)
- [Usage Examples](#usage-examples)
- [Integration with Matrix](#integration-with-matrix)
- [Testing](#testing)
- [API Reference](#api-reference)

## Overview

The component package serves as the **structural foundation** for the neural simulation system. It provides:

- **Common component architecture** - Shared lifecycle, metadata, and identification
- **Biological interfaces** - Chemical signaling, electrical coupling, spatial awareness
- **Thread-safe implementations** - Concurrent access protection for all operations
- **Extensible design** - Easy to add new component types and capabilities
- **Matrix integration** - Clean interfaces for extracellular matrix coordination

### What Components Are

Components are **autonomous biological entities** that can:
- Maintain their own state and lifecycle
- Participate in chemical signaling (neurotransmitter release/reception)
- Engage in electrical coupling (gap junction communication)
- Process neural signals (receive/transmit messages)
- Monitor their own health and activity
- Coordinate with the extracellular matrix environment

### What Components Are NOT

Components do **NOT** handle:
- **Spatial calculations** - Distance measurements are done by the matrix
- **Signal routing** - Message delivery is handled by the matrix
- **Network topology** - Connectivity patterns are managed by the matrix
- **Global coordination** - System-wide synchronization is matrix responsibility

## Architecture Philosophy

### Separation of Concerns

The component package follows strict separation of concerns:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   COMPONENTS    │    │     MATRIX      │    │    MESSAGES     │
│                 │    │                 │    │                 │
│ • Lifecycle     │    │ • Spatial Calc  │    │ • Signal Data   │
│ • State Mgmt    │    │ • Routing       │    │ • Chemical Info │
│ • Interfaces    │    │ • Coordination  │    │ • Timing Data   │
│ • Identity      │    │ • Networking    │    │ • Metadata      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Biological Inspiration

Each component represents a **biological cellular entity**:
- **Neurons** - Excitable cells that process and transmit information
- **Synapses** - Connections between neurons with plasticity and learning
- **Glial Cells** - Support cells (astrocytes, oligodendrocytes) that maintain tissue
- **Microglia** - Immune cells that monitor health and prune connections
- **Ependymal Cells** - Barrier cells that regulate chemical environments

### Interface-Driven Design

Components implement **specialized interfaces** based on their biological capabilities:
- Not all components can receive chemicals (only those with receptors)
- Not all components can transmit signals (only those with output mechanisms)
- Not all components have spatial awareness (some are purely computational)

## Component Types

### ComponentType Enum

```go
type ComponentType int

const (
    TypeNeuron        ComponentType = iota // Excitable neural cell
    TypeSynapse                            // Synaptic connection
    TypeGlialCell                          // Support cell (astrocyte, oligodendrocyte, etc.)
    TypeMicrogliaCell                      // Immune cell of the brain
    TypeEpendymalCell                      // CSF-brain barrier cell
)
```

### Biological Roles

| Type | Biological Function | Typical Interfaces |
|------|-------------------|------------------|
| **Neuron** | Information processing and transmission | `MessageReceiver`, `ChemicalReceiver`, `ElectricalReceiver` |
| **Synapse** | Connection and learning between neurons | `MessageTransmitter`, `MonitorableComponent` |
| **GlialCell** | Structural support and metabolic assistance | `ChemicalReleaser`, `SpatialComponent` |
| **MicrogliaCell** | Immune surveillance and synaptic pruning | `MonitorableComponent`, `ChemicalReceiver` |
| **EpendymalCell** | Chemical barrier and CSF regulation | `ChemicalReceiver`, `ChemicalReleaser` |

## Core Interfaces

### Component Interface

The foundational interface that **all components must implement**:

```go
type Component interface {
    // Core identification
    ID() string
    Type() ComponentType
    Position() Position3D
    State() ComponentState

    // Lifecycle management
    IsActive() bool
    Start() error
    Stop() error

    // Metadata and monitoring
    GetMetadata() map[string]interface{}
    UpdateMetadata(key string, value interface{})

    // State management
    SetState(state ComponentState)

    // Activity monitoring
    GetActivityLevel() float64
    GetLastActivity() time.Time
}
```

### Component States

```go
type ComponentState int

const (
    StateActive       ComponentState = iota // Normal operational state
    StateInactive                           // Temporarily disabled
    StateShuttingDown                       // Graceful shutdown in progress
    StateDeveloping                         // Growing/maturing (developmental)
    StateDying                              // Programmed cell death/apoptosis
    StateDamaged                            // Damaged but potentially recoverable
    StateMaintenance                        // Undergoing maintenance/repair
    StateHibernating                        // Low-activity conservation state
)
```

### Position3D

Components store their **3D spatial position** for matrix coordination:

```go
type Position3D struct {
    X, Y, Z float64
}
```

**Note**: Components store position but do **NOT** calculate distances. All spatial calculations are performed by the matrix.

## Specialized Interfaces

### Chemical Signaling

For components that participate in neurotransmitter communication:

```go
// Components that can receive chemical signals
type ChemicalReceiver interface {
    Component
    GetReceptors() []types.LigandType
    Bind(ligandType types.LigandType, sourceID string, concentration float64)
}

// Components that can release chemical signals
type ChemicalReleaser interface {
    Component
    GetReleasedLigands() []types.LigandType
}
```

### Electrical Signaling

For components that participate in gap junction networks:

```go
// Components that can receive electrical signals
type ElectricalReceiver interface {
    Component
    OnSignal(signalType types.SignalType, sourceID string, data interface{})
}

// Components that can send electrical signals
type ElectricalTransmitter interface {
    Component
    GetSignalTypes() []types.SignalType
}
```

### Message Processing

For components that handle neural signal transmission:

```go
// Components that can receive neural signals
type MessageReceiver interface {
    Component
    Receive(msg types.NeuralSignal)
}

// Components that can transmit neural signals
type MessageTransmitter interface {
    Component
    Transmit(signal float64) error
}
```

### Spatial Awareness

For components that have spatial properties:

```go
type SpatialComponent interface {
    Component
    SetPosition(position Position3D)
    GetRange() float64
}
```

### Health Monitoring

For components that provide health and performance metrics:

```go
type MonitorableComponent interface {
    Component
    GetHealthMetrics() HealthMetrics
}

type HealthMetrics struct {
    ActivityLevel   float64   `json:"activity_level"`
    ConnectionCount int       `json:"connection_count"`
    ProcessingLoad  float64   `json:"processing_load"`
    LastHealthCheck time.Time `json:"last_health_check"`
    HealthScore     float64   `json:"health_score"`
    Issues          []string  `json:"issues"`
}
```

## Base Implementation

### BaseComponent

The `BaseComponent` struct provides a **thread-safe implementation** of the core `Component` interface:

```go
type BaseComponent struct {
    id            string
    componentType ComponentType
    position      Position3D
    state         ComponentState
    metadata      map[string]interface{}
    lastActivity  time.Time
    isActive      bool
    mu            sync.RWMutex
}
```

### Key Features

- **Thread-safe** - All operations protected with read-write mutex
- **Metadata isolation** - Returns copies to prevent external modification
- **Activity tracking** - Automatic timestamps on state changes
- **Lifecycle management** - Proper start/stop state transitions

### Creating Components

```go
// Create a basic component
comp := NewBaseComponent("neuron-1", TypeNeuron, Position3D{X: 10, Y: 20, Z: 30})

// Create with specialized capabilities
spatialComp := NewSpatialComponent("spatial-1", TypeGlialCell, position, 50.0)
monitorableComp := NewMonitorableComponent("monitor-1", TypeMicrogliaCell, position)
```

## Usage Examples

### Basic Component Usage

```go
package main

import (
    "fmt"
    "github.com/SynapticNetworks/temporal-neuron/component"
)

func main() {
    // Create a neuron component
    neuron := component.NewBaseComponent(
        "neuron-001", 
        component.TypeNeuron, 
        component.Position3D{X: 0, Y: 0, Z: 0},
    )
    
    // Start the component
    err := neuron.Start()
    if err != nil {
        panic(err)
    }
    
    // Add metadata
    neuron.UpdateMetadata("type", "pyramidal")
    neuron.UpdateMetadata("layer", "L5")
    
    // Check status
    fmt.Printf("Component %s is active: %v\n", neuron.ID(), neuron.IsActive())
    fmt.Printf("Current state: %s\n", neuron.State())
    fmt.Printf("Activity level: %.2f\n", neuron.GetActivityLevel())
    
    // Stop the component
    neuron.Stop()
}
```

### Creating Custom Components

```go
// Custom neuron that implements multiple interfaces
type CustomNeuron struct {
    *component.BaseComponent
    receptors []types.LigandType
    synapses  map[string]SynapseConnection
}

// Implement ChemicalReceiver interface
func (cn *CustomNeuron) GetReceptors() []types.LigandType {
    return cn.receptors
}

func (cn *CustomNeuron) Bind(ligandType types.LigandType, sourceID string, concentration float64) {
    // Process chemical binding
    cn.UpdateMetadata("last_binding", time.Now())
    // ... custom binding logic
}

// Implement MessageReceiver interface
func (cn *CustomNeuron) Receive(msg types.NeuralSignal) {
    // Process incoming neural signal
    cn.UpdateMetadata("last_signal", msg.Timestamp)
    // ... custom signal processing
}
```

### Component Filtering and Management

```go
// Create a collection of components
components := []component.Component{
    component.NewBaseComponent("neuron1", component.TypeNeuron, component.Position3D{}),
    component.NewBaseComponent("synapse1", component.TypeSynapse, component.Position3D{}),
    component.NewBaseComponent("glia1", component.TypeGlialCell, component.Position3D{}),
}

// Filter by type
neurons := component.FilterComponentsByType(components, component.TypeNeuron)
fmt.Printf("Found %d neurons\n", len(neurons))

// Filter by state
activeComponents := component.FilterComponentsByState(components, component.StateActive)
fmt.Printf("Found %d active components\n", len(activeComponents))

// Create component info for registration
for _, comp := range components {
    info := component.CreateComponentInfo(comp)
    fmt.Printf("Component: %s, Type: %s, State: %s\n", 
        info.ID, info.Type, info.State)
}
```

## Integration with Matrix

### Matrix Coordination

Components are designed to work with the **ExtracellularMatrix** for coordination:

```go
// Components provide interfaces that the matrix can use
type NeuronInterface interface {
    component.Component
    component.ChemicalReceiver
    component.ElectricalReceiver
    component.MessageReceiver
    // ... additional neuron-specific methods
}

// Matrix creates components through factory functions
neuron, err := matrix.CreateNeuron(NeuronConfig{
    NeuronType: "pyramidal",
    Position:   component.Position3D{X: 10, Y: 20, Z: 30},
    // ... additional configuration
})
```

### Callback Injection

The matrix injects biological functions into components via callbacks:

```go
// Example of matrix callbacks provided to components
callbacks := NeuronCallbacks{
    ReleaseChemical: func(ligand types.LigandType, concentration float64) error {
        return matrix.chemicalSystem.Release(ligand, neuronID, concentration)
    },
    SendElectricalSignal: func(signal types.SignalType, data interface{}) {
        matrix.electricalSystem.Broadcast(signal, neuronID, data)
    },
    // ... other biological functions
}
```

## Testing

### Running Tests

```bash
# Run all component tests
go test ./component -v

# Run specific test categories
go test ./component -run TestBaseComponent -v
go test ./component -run TestInterface -v
go test ./component -run TestConcurrent -v

# Run benchmarks
go test ./component -bench=. -v
```

### Test Coverage

The component package includes comprehensive tests for:

- **Core functionality** - All interface methods and lifecycle operations
- **Thread safety** - Concurrent access to metadata and state
- **Interface compliance** - Mock implementations for all specialized interfaces
- **Type safety** - Proper enum handling and string representations
- **Edge cases** - Invalid states, boundary conditions, error handling

### Mock Components

The test suite includes mock implementations for testing interface compliance:

```go
// Mock chemical receiver for testing
type MockChemicalReceiver struct {
    *component.BaseComponent
    receptors []types.LigandType
    bindings  map[types.LigandType]float64
}

// Use in tests
receiver := NewMockChemicalReceiver("test-receiver")
receiver.Bind(types.LigandGlutamate, "source", 0.5)
```

## API Reference

### Core Types

| Type | Description |
|------|-------------|
| `Component` | Base interface all components implement |
| `ComponentType` | Enumeration of component types |
| `ComponentState` | Enumeration of component lifecycle states |
| `Position3D` | 3D spatial coordinates |
| `ComponentInfo` | Complete component registration information |
| `HealthMetrics` | Health and performance monitoring data |

### Specialized Interfaces

| Interface | Purpose |
|-----------|---------|
| `ChemicalReceiver` | Components that can receive neurotransmitters |
| `ChemicalReleaser` | Components that can release neurotransmitters |
| `ElectricalReceiver` | Components that can receive electrical signals |
| `ElectricalTransmitter` | Components that can send electrical signals |
| `MessageReceiver` | Components that can receive neural signals |
| `MessageTransmitter` | Components that can transmit neural signals |
| `SpatialComponent` | Components with spatial awareness |
| `MonitorableComponent` | Components that provide health metrics |

### Constructors

| Function | Returns | Description |
|----------|---------|-------------|
| `NewBaseComponent(id, type, pos)` | `*BaseComponent` | Creates basic component |
| `NewSpatialComponent(id, type, pos, range)` | `*DefaultSpatialComponent` | Creates spatial component |
| `NewMonitorableComponent(id, type, pos)` | `*DefaultMonitorableComponent` | Creates monitorable component |

### Utility Functions

| Function | Returns | Description |
|----------|---------|-------------|
| `CreateComponentInfo(comp)` | `ComponentInfo` | Creates registration info |
| `FilterComponentsByType(comps, type)` | `[]Component` | Filters by component type |
| `FilterComponentsByState(comps, state)` | `[]Component` | Filters by component state |

### Thread Safety

All component operations are **thread-safe**:
- Metadata operations use read-write mutex for optimal performance
- State changes are atomic and properly synchronized
- Position updates are protected against concurrent modification
- Activity tracking is safe for concurrent access

### Performance Characteristics

- **Metadata access**: O(1) for reads, O(1) for updates
- **State management**: O(1) for all operations
- **Type filtering**: O(n) linear scan
- **Interface compliance**: O(1) type assertion
- **Memory usage**: Minimal overhead, efficient metadata storage

---

## Contributing

When extending the component package:

1. **Follow interface patterns** - Use composition over inheritance
2. **Maintain thread safety** - Protect all shared state with appropriate locking
3. **Add comprehensive tests** - Include unit tests and thread safety tests
4. **Document biological inspiration** - Explain the biological basis for new features
5. **Preserve separation of concerns** - Keep spatial calculations in the matrix

## License

This component package is part of the temporal-neuron project and follows the same licensing terms.