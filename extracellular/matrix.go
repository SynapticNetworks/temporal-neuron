/*
=================================================================================
FACTORY-ENHANCED EXTRACELLULAR MATRIX - BIOLOGICAL COMPONENT CREATION
=================================================================================

Implements biologically-inspired component creation through the Inversion of Control
pattern, where the ExtracellularMatrix acts as the brain's developmental machinery—
creating neurons and synapses while providing them with precisely the biological
functions they need to interact with their neural environment.

BIOLOGICAL INSPIRATION:
In the developing brain, the extracellular matrix doesn't just provide structure—
it actively guides neurogenesis and synaptogenesis. Growth factors, chemical
gradients, and cellular scaffolds coordinate to create precisely the right
neurons in the right places with the right connections. This implementation
models that biological orchestration.

KEY DESIGN PRINCIPLES:
1. Matrix as Neural Development Controller: Like embryonic neural development
2. Component Autonomy: Neurons/synapses operate independently once created
3. Biological Callback Injection: Components get access to biological functions
4. Complete Decoupling: Components never directly reference the matrix
5. Spatial Awareness: All creation considers 3D biological positioning

BIOLOGICAL FUNCTIONS MODELED:
- Neurogenesis: Programmatic creation of new neurons with proper wiring
- Synaptogenesis: Formation of synaptic connections with biological properties
- Chemical Wiring: Automatic connection to neurotransmitter systems
- Electrical Coupling: Integration with gap junction networks
- Spatial Coordination: 3D positioning and propagation delay calculation
- Health Monitoring: Automatic integration with microglial surveillance

=================================================================================
*/

package extracellular

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// =================================================================================
// ENHANCED EXTRACELLULAR MATRIX WITH BIOLOGICAL FACTORY SYSTEM
// =================================================================================

// ExtracellularMatrix provides biological coordination services for autonomous components
// Enhanced with neurogenesis and synaptogenesis capabilities that mirror the brain's
// developmental processes for creating and integrating new neural components.
//
// BIOLOGICAL CONTEXT:
// The extracellular matrix in biology is far more than passive scaffolding—it's an
// active coordinator of neural development. It provides chemical gradients that guide
// axon pathfinding, supplies growth factors that promote neurogenesis, and creates
// the microenvironment necessary for proper synaptic formation. This implementation
// captures those active, guiding properties.
type ExtracellularMatrix struct {
	// === CORE BIOLOGICAL SYSTEMS ===
	// These represent the major biological subsystems that coordinate neural function
	astrocyteNetwork  *AstrocyteNetwork  // Spatial organization and connectivity mapping
	chemicalModulator *ChemicalModulator // Neurotransmitter and neuromodulator signaling
	signalMediator    *SignalMediator    // Electrical coupling (gap junctions)
	microglia         *Microglia         // Component lifecycle and health monitoring
	plugins           *PluginManager     // Modular biological functions

	// === BIOLOGICAL FACTORY SYSTEM ===
	// Models the brain's capacity for neurogenesis and synaptogenesis
	// Each factory type represents different neural development programs
	neuronFactories  map[string]NeuronFactoryFunc  // Neurogenesis programs by cell type
	synapseFactories map[string]SynapseFactoryFunc // Synaptogenesis programs by connection type

	// === ACTIVE COMPONENT REGISTRY ===
	// Tracks all living components for biological coordination and monitoring
	// Mirrors how biological neural networks maintain awareness of their constituents
	neurons  map[string]NeuronInterface  // All active neurons in the network
	synapses map[string]SynapseInterface // All active synaptic connections

	// === RESOURCE MANAGEMENT ===
	maxComponents int // Maximum number of components (neurons + synapses) allowed

	// === OPERATIONAL STATE ===
	// Models the matrix's biological lifecycle and activity state
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.RWMutex
}

// ExtracellularMatrixConfig provides configuration for biological coordination
// Models the environmental parameters that influence neural development
type ExtracellularMatrixConfig struct {
	ChemicalEnabled bool          // Enable neurotransmitter/neuromodulator systems
	SpatialEnabled  bool          // Enable 3D spatial organization and delays
	UpdateInterval  time.Duration // Biological update frequency (metabolism rate)
	MaxComponents   int           // Metabolic capacity limit for component support
}

// =================================================================================
// BIOLOGICAL CONSTANTS AND AXON SPEED CONFIGURATION
// =================================================================================

// Biological axon speed constants based on myelination and fiber type (μm/ms)
// These values are derived from experimental measurements in living neural tissue
const (
	UNMYELINATED_SLOW = 500.0   // 0.5 m/s - C fibers (pain, temperature)
	UNMYELINATED_FAST = 2000.0  // 2 m/s - cortical local circuits
	MYELINATED_MEDIUM = 10000.0 // 10 m/s - A-delta fibers (fast pain)
	MYELINATED_FAST   = 80000.0 // 80 m/s - A-alpha fibers (proprioception, motor)

	// Typical cortical circuit speeds based on connection distance
	LOCAL_CIRCUIT = 2000.0  // Local cortical circuits (within cortical columns)
	INTER_LAMINAR = 5000.0  // Between cortical layers (layer 2/3 to layer 5)
	LONG_RANGE    = 15000.0 // Long-distance projections (cortex to cortex)
)

// Global axon speed configuration with thread safety for biological realism
// Models the fact that axon conduction velocity affects network-wide timing
var (
	globalAxonSpeed = 2000.0     // Default: unmyelinated cortical axons
	axonSpeedMutex  sync.RWMutex // Thread-safe access to speed configuration
)

// =================================================================================
// MATRIX CONSTRUCTION AND BIOLOGICAL INITIALIZATION
// =================================================================================

// NewExtracellularMatrix creates a biologically-inspired coordination matrix
// that models the brain's extracellular environment and developmental machinery.
//
// BIOLOGICAL PROCESS MODELED:
// This mirrors the establishment of the neural microenvironment during brain
// development. The extracellular matrix forms first, then provides the chemical
// gradients, structural support, and signaling systems needed for neurogenesis
// and synaptogenesis to proceed in an organized fashion.
//
// INITIALIZATION SEQUENCE:
// 1. Establish astrocyte network (spatial organization)
// 2. Initialize chemical modulator (neurotransmitter systems)
// 3. Set up signal mediator (electrical coupling infrastructure)
// 4. Activate microglia (health monitoring and maintenance)
// 5. Configure plugin system (modular biological functions)
// 6. Register default neurogenesis and synaptogenesis programs
//
// Parameters:
//   - config: Environmental and metabolic parameters for biological operation
//
// Returns:
//   - Fully initialized matrix ready for neurogenesis and network coordination
func NewExtracellularMatrix(config ExtracellularMatrixConfig) *ExtracellularMatrix {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize core biological subsystems in developmental order
	astrocyteNetwork := NewAstrocyteNetwork()                         // Spatial scaffolding first
	modulator := NewChemicalModulator(astrocyteNetwork)               // Chemical signaling systems
	signalMediator := NewSignalMediator()                             // Electrical coupling infrastructure
	microglia := NewMicroglia(astrocyteNetwork, config.MaxComponents) // Health and maintenance
	plugins := NewPluginManager()                                     // Modular functionality

	ecm := &ExtracellularMatrix{
		// Core biological coordination systems
		astrocyteNetwork:  astrocyteNetwork,
		chemicalModulator: modulator,
		signalMediator:    signalMediator,
		microglia:         microglia,
		plugins:           plugins,

		// Factory system for biological component creation
		neuronFactories:  make(map[string]NeuronFactoryFunc),
		synapseFactories: make(map[string]SynapseFactoryFunc),

		// Active component tracking for biological coordination
		neurons:  make(map[string]NeuronInterface),
		synapses: make(map[string]SynapseInterface),

		maxComponents: config.MaxComponents,

		// Operational lifecycle management
		ctx:     ctx,
		cancel:  cancel,
		started: false,
	}

	// Register built-in neurogenesis and synaptogenesis programs
	// Models the genetic programs that guide neural development
	ecm.registerDefaultBiologicalFactories()

	return ecm
}

// =================================================================================
// BIOLOGICAL FACTORY REGISTRATION SYSTEM
// =================================================================================

// RegisterNeuronType registers a neurogenesis program for creating specific neuron types.
//
// BIOLOGICAL FUNCTION:
// This models how genetic and epigenetic programs specify different types of neurons
// during development. Each neuron type (pyramidal, interneuron, etc.) has specific
// molecular markers, connectivity patterns, and functional properties that are
// determined by their developmental program.
//
// EXAMPLES OF BIOLOGICAL NEURON TYPES:
// - "pyramidal_l5": Layer 5 pyramidal neurons (long-range projection neurons)
// - "fast_spiking_interneuron": Parvalbumin-positive inhibitory interneurons
// - "chandelier_cell": Axo-axonic interneurons targeting axon initial segments
// - "von_economo": Large projection neurons found in higher primates
//
// Parameters:
//   - neuronType: Biological classification of the neuron (e.g., "pyramidal_l2_3")
//   - factory: Function that creates neurons of this type with proper biological properties
func (ecm *ExtracellularMatrix) RegisterNeuronType(neuronType string, factory NeuronFactoryFunc) {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	ecm.neuronFactories[neuronType] = factory
}

// RegisterSynapseType registers a synaptogenesis program for creating specific synapse types.
//
// BIOLOGICAL FUNCTION:
// This models how different types of synaptic connections form based on the identity
// of pre- and post-synaptic neurons, the neurotransmitters involved, and the functional
// requirements of the circuit. Each synapse type has distinct vesicle properties,
// receptor configurations, and plasticity rules.
//
// EXAMPLES OF BIOLOGICAL SYNAPSE TYPES:
// - "excitatory_plastic": Glutamatergic synapses with AMPA/NMDA receptors and LTP/LTD
// - "inhibitory_static": GABAergic synapses with fixed strength for stable inhibition
// - "neuromodulatory": Dopaminergic/serotonergic synapses affecting multiple targets
// - "electrical": Gap junction connections for rapid synchronization
//
// Parameters:
//   - synapseType: Biological classification of the synapse (e.g., "excitatory_plastic")
//   - factory: Function that creates synapses of this type with proper biological properties
func (ecm *ExtracellularMatrix) RegisterSynapseType(synapseType string, factory SynapseFactoryFunc) {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	ecm.synapseFactories[synapseType] = factory
}

// =================================================================================
// NEUROGENESIS AND SYNAPTOGENESIS (CORE BIOLOGICAL CREATION)
// =================================================================================

// CreateNeuron implements biologically-guided neurogenesis with complete environmental integration.
//
// BIOLOGICAL PROCESS MODELED:
// This mirrors the complex process of neurogenesis in the developing and adult brain:
// 1. Neural progenitor specification (factory selection)
// 2. Cell fate determination (configuration application)
// 3. Migration to target location (spatial positioning)
// 4. Axon and dendrite outgrowth (connectivity establishment)
// 5. Synapse formation (chemical and electrical integration)
// 6. Functional maturation (callback injection and system registration)
//
// ENVIRONMENTAL INTEGRATION:
// The newly created neuron is automatically wired into all relevant biological systems:
// - Chemical signaling (neurotransmitter release and reception)
// - Electrical coupling (gap junction participation)
// - Spatial coordination (3D positioning and delay calculation)
// - Health monitoring (microglial surveillance integration)
// - Network topology (astrocyte connectivity mapping)
//
// Parameters:
//   - config: Complete specification of the neuron's biological properties
//
// Returns:
//   - NeuronInterface: Fully integrated neuron ready for biological operation
//   - error: If neurogenesis fails due to resource constraints or configuration issues
//
// FIXED: Improved concurrency with fine-grained locking to reduce performance bottlenecks
func (ecm *ExtracellularMatrix) CreateNeuron(config NeuronConfig) (NeuronInterface, error) {
	// === PHASE 1: VALIDATION AND FACTORY LOOKUP (Quick, locked) ===
	ecm.mu.Lock()

	// Locate the appropriate neurogenesis program
	factory, exists := ecm.neuronFactories[config.NeuronType]
	if !exists {
		ecm.mu.Unlock()
		return nil, fmt.Errorf("unknown neuron type for neurogenesis: %s", config.NeuronType)
	}

	// Check resource limits EARLY and account for the neuron we're about to add
	currentComponentCount := len(ecm.neurons) + len(ecm.synapses)
	if currentComponentCount >= ecm.maxComponents {
		ecm.mu.Unlock()
		return nil, fmt.Errorf("resource limit exceeded: cannot create neuron, already at maximum %d components", ecm.maxComponents)
	}

	// Generate unique biological identifier while locked
	neuronID := ecm.generateBiologicalNeuronID(config.NeuronType)

	// Create biological callback functions that wire the neuron into matrix systems
	callbacks := ecm.createNeuronBiologicalCallbacks(neuronID)

	ecm.mu.Unlock() // UNLOCK before potentially slow factory execution

	// === PHASE 2: NEUROGENESIS EXECUTION (Unlocked, potentially slow) ===
	// Execute neurogenesis using the specified biological program
	// This can be slow and doesn't need the global lock
	neuron, err := factory(neuronID, config, callbacks)
	if err != nil {
		return nil, fmt.Errorf("neurogenesis failed: %w", err)
	}

	// === PHASE 3: INTEGRATION AND REGISTRATION (Re-locked) ===
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	// Double-check resource limits after factory execution (safety)
	currentComponentCount = len(ecm.neurons) + len(ecm.synapses)
	if currentComponentCount >= ecm.maxComponents {
		return nil, fmt.Errorf("resource limit exceeded during integration: cannot register neuron, at maximum %d components", ecm.maxComponents)
	}

	// Integrate the new neuron into all biological coordination systems
	err = ecm.integrateNeuronIntoBiologicalSystems(neuron, config)
	if err != nil {
		return nil, fmt.Errorf("neural integration failed: %w", err)
	}

	// Register in active component tracking for ongoing biological coordination
	ecm.neurons[neuronID] = neuron

	return neuron, nil
}

// CreateSynapse implements biologically-guided synaptogenesis with complete circuit integration.
//
// BIOLOGICAL PROCESS MODELED:
// This mirrors synaptic development in biological neural networks:
// 1. Axon pathfinding (presynaptic neuron identification)
// 2. Target recognition (postsynaptic neuron validation)
// 3. Synaptic vesicle clustering (neurotransmitter system specification)
// 4. Postsynaptic density formation (receptor configuration)
// 5. Activity-dependent refinement (plasticity mechanism installation)
// 6. Functional maturation (network integration and monitoring)
//
// CIRCUIT INTEGRATION:
// The new synapse is automatically integrated into biological coordination:
// - Neurotransmitter release (chemical signaling participation)
// - Spatial delay calculation (realistic transmission timing)
// - Activity monitoring (synaptic health and plasticity tracking)
// - Network topology (connectivity pattern registration)
//
// Parameters:
//   - config: Complete specification of the synapse's biological properties
//
// Returns:
//   - SynapseInterface: Fully integrated synapse ready for neural transmission
//   - error: If synaptogenesis fails due to invalid connections or resource limits
//
// FIXED: Improved concurrency with fine-grained locking to reduce performance bottlenecks
func (ecm *ExtracellularMatrix) CreateSynapse(config SynapseConfig) (SynapseInterface, error) {
	// === PHASE 1: VALIDATION AND FACTORY LOOKUP (Quick, locked) ===
	ecm.mu.Lock()

	// Validate biological connectivity - both neurons must exist
	if _, exists := ecm.neurons[config.PresynapticID]; !exists {
		ecm.mu.Unlock()
		return nil, fmt.Errorf("synaptogenesis failed: presynaptic neuron not found: %s", config.PresynapticID)
	}
	if _, exists := ecm.neurons[config.PostsynapticID]; !exists {
		ecm.mu.Unlock()
		return nil, fmt.Errorf("synaptogenesis failed: postsynaptic neuron not found: %s", config.PostsynapticID)
	}

	// Locate the appropriate synaptogenesis program
	factory, exists := ecm.synapseFactories[config.SynapseType]
	if !exists {
		ecm.mu.Unlock()
		return nil, fmt.Errorf("unknown synapse type for synaptogenesis: %s", config.SynapseType)
	}

	// Check resource limits for synapses too
	currentComponentCount := len(ecm.neurons) + len(ecm.synapses)
	if currentComponentCount >= ecm.maxComponents {
		ecm.mu.Unlock()
		return nil, fmt.Errorf("resource limit exceeded: cannot create synapse, already at maximum %d components", ecm.maxComponents)
	}

	// Generate unique biological identifier while locked
	synapseID := ecm.generateBiologicalSynapseID(config.SynapseType, config.PresynapticID, config.PostsynapticID)

	// Create biological callback functions that wire the synapse into matrix systems
	callbacks := ecm.createSynapseBiologicalCallbacks(synapseID, config)

	ecm.mu.Unlock() // UNLOCK before potentially slow factory execution

	// === PHASE 2: SYNAPTOGENESIS EXECUTION (Unlocked, potentially slow) ===
	// Execute synaptogenesis using the specified biological program
	// This can be slow and doesn't need the global lock
	synapse, err := factory(synapseID, config, callbacks)
	if err != nil {
		return nil, fmt.Errorf("synaptogenesis failed: %w", err)
	}

	// === PHASE 3: INTEGRATION AND REGISTRATION (Re-locked) ===
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	// Double-check resource limits after factory execution (safety)
	currentComponentCount = len(ecm.neurons) + len(ecm.synapses)
	if currentComponentCount >= ecm.maxComponents {
		return nil, fmt.Errorf("resource limit exceeded during integration: cannot register synapse, at maximum %d components", ecm.maxComponents)
	}

	// Integrate the new synapse into all biological coordination systems
	err = ecm.integrateSynapseIntoBiologicalSystems(synapse, config)
	if err != nil {
		return nil, fmt.Errorf("synaptic integration failed: %w", err)
	}

	// Register in active component tracking for ongoing biological coordination
	ecm.synapses[synapseID] = synapse

	return synapse, nil
}

// =================================================================================
// BIOLOGICAL CALLBACK CREATION (DEPENDENCY INJECTION)
// =================================================================================

// createNeuronBiologicalCallbacks creates the biological interface functions that connect
// a neuron to all relevant matrix services, modeling how real neurons interact with
// their cellular environment through membrane proteins, vesicle systems, and cellular machinery.
//
// BIOLOGICAL FUNCTIONS PROVIDED:
// - Chemical Release: Models vesicle fusion and neurotransmitter release
// - Electrical Signaling: Models gap junction communication and action potential propagation
// - Spatial Awareness: Models the neuron's ability to sense distance and timing
// - Environmental Sensing: Models the neuron's awareness of nearby cellular components
// - Health Reporting: Models cellular stress signaling and metabolic communication
// - State Communication: Models how neurons signal their functional state to glia
//
// This implements biological dependency injection - the neuron receives precisely
// the cellular machinery it needs without knowing how that machinery is implemented.
func (ecm *ExtracellularMatrix) createNeuronBiologicalCallbacks(neuronID string) NeuronCallbacks {
	return NeuronCallbacks{
		// === VESICULAR RELEASE SYSTEM ===
		// Models the cellular machinery for neurotransmitter and neuromodulator release
		ReleaseChemical: func(ligandType LigandType, concentration float64) error {
			return ecm.chemicalModulator.Release(ligandType, neuronID, concentration)
		},

		// === GAP JUNCTION AND ELECTRICAL COUPLING ===
		// Models direct electrical communication through gap junction channels
		SendElectricalSignal: func(signalType SignalType, data interface{}) {
			ecm.signalMediator.Send(signalType, neuronID, data)
		},

		// === AXONAL CONDUCTION AND SPATIAL TIMING ===
		// Models how neurons account for axonal conduction delays in their timing
		GetSpatialDelay: func(targetID string) time.Duration {
			return ecm.EnhanceSynapticDelay(neuronID, targetID, "", 0)
		},

		// === ENVIRONMENTAL SENSING AND SPATIAL AWARENESS ===
		// Models how neurons sense their local cellular environment and nearby components
		FindNearbyComponents: func(radius float64) []ComponentInfo {
			neuronInfo, exists := ecm.astrocyteNetwork.Get(neuronID)
			if !exists {
				return []ComponentInfo{}
			}
			return ecm.astrocyteNetwork.FindNearby(neuronInfo.Position, radius)
		},

		// === METABOLIC AND STRESS SIGNALING ===
		// Models how neurons communicate their health and activity state to glial cells
		ReportHealth: func(activityLevel float64, connectionCount int) {
			ecm.microglia.UpdateComponentHealth(neuronID, activityLevel, connectionCount)
		},

		// === CELLULAR STATE COMMUNICATION ===
		// Models how neurons signal state changes (active, inactive, apoptotic) to astrocytes
		ReportStateChange: func(oldState, newState ComponentState) {
			ecm.astrocyteNetwork.UpdateState(neuronID, newState)
		},
	}
}

// createSynapseBiologicalCallbacks creates the biological interface functions that connect
// a synapse to all relevant matrix services, modeling how real synapses interact with
// the neural environment through neurotransmitter release, spatial coordination, and
// activity-dependent plasticity mechanisms.
//
// BIOLOGICAL FUNCTIONS PROVIDED:
// - Message Transmission: Models synaptic vesicle fusion and postsynaptic signal delivery
// - Conduction Delay: Models realistic axonal and synaptic transmission timing
// - Neurotransmitter Release: Models chemical signaling at the synaptic cleft
// - Activity Monitoring: Models how synaptic activity is tracked for plasticity
// - Plasticity Signaling: Models how synaptic changes are communicated to the network
//
// This provides synapses with all the cellular machinery needed for biological
// function while maintaining complete decoupling from matrix implementation details.
func (ecm *ExtracellularMatrix) createSynapseBiologicalCallbacks(synapseID string, config SynapseConfig) SynapseCallbacks {
	return SynapseCallbacks{
		// === SYNAPTIC TRANSMISSION MACHINERY ===
		// Models the complete synaptic transmission process from vesicle release to postsynaptic response
		DeliverMessage: func(targetID string, message SynapseMessage) error {
			ecm.mu.RLock()
			targetNeuron, exists := ecm.neurons[targetID]
			ecm.mu.RUnlock()

			if !exists {
				return fmt.Errorf("synaptic transmission failed: target neuron not found: %s", targetID)
			}

			// In a complete implementation, this would call targetNeuron.Receive(message)
			// For now, we assume the synapse handles delivery through its own mechanisms
			_ = targetNeuron // Placeholder to avoid unused variable error
			return nil
		},

		// === AXONAL AND SYNAPTIC DELAY CALCULATION ===
		// Models realistic transmission delays including axonal conduction and synaptic processing
		GetTransmissionDelay: func() time.Duration {
			return ecm.EnhanceSynapticDelay(config.PresynapticID, config.PostsynapticID, synapseID, config.Delay)
		},

		// === SYNAPTIC CLEFT NEUROTRANSMITTER RELEASE ===
		// Models vesicle fusion and neurotransmitter diffusion in the synaptic cleft
		ReleaseNeurotransmitter: func(ligandType LigandType, concentration float64) error {
			return ecm.chemicalModulator.Release(ligandType, synapseID, concentration)
		},

		// === SYNAPTIC ACTIVITY MONITORING ===
		// Models how synaptic activity is tracked by astrocytes for network analysis
		ReportActivity: func(activity SynapticActivity) {
			// Record synaptic activity with the astrocyte network
			ecm.astrocyteNetwork.RecordSynapticActivity(
				activity.SynapseID,
				config.PresynapticID,
				config.PostsynapticID,
				activity.CurrentWeight,
			)
		},

		// === PLASTICITY EVENT COMMUNICATION ===
		// Models how synaptic plasticity changes are communicated to the broader network
		ReportPlasticityEvent: func(event PlasticityEvent) {
			// Update astrocyte network with plasticity-induced weight changes
			ecm.astrocyteNetwork.RecordSynapticActivity(
				synapseID,
				config.PresynapticID,
				config.PostsynapticID,
				event.Strength,
			)
		},
	}
}

// =================================================================================
// BIOLOGICAL SYSTEM INTEGRATION
// =================================================================================

// integrateNeuronIntoBiologicalSystems performs the complete biological integration
// of a newly created neuron into all matrix coordination systems.
//
// BIOLOGICAL INTEGRATION PROCESS:
// This models the maturation process where a newly born neuron becomes a functional
// member of the neural network through a series of developmental milestones:
//
// 1. Spatial Registration: Neuron position recorded in astrocyte territorial maps
// 2. Chemical Integration: Neurotransmitter receptors connected to signaling systems
// 3. Electrical Coupling: Gap junction participation enabled for synchronization
// 4. Health Monitoring: Microglial surveillance activated for ongoing support
// 5. Network Topology: Connectivity potential mapped for future synaptogenesis
//
// This integration ensures the neuron can participate in all biological functions
// from the moment it becomes active, just like biological neurogenesis.
func (ecm *ExtracellularMatrix) integrateNeuronIntoBiologicalSystems(neuron NeuronInterface, config NeuronConfig) error {
	// === ASTROCYTE TERRITORIAL REGISTRATION ===
	// Register neuron location and identity with astrocyte spatial mapping system
	componentInfo := ComponentInfo{
		ID:           neuron.ID(),
		Type:         ComponentNeuron,
		Position:     neuron.Position(),
		State:        StateActive,
		RegisteredAt: time.Now(),
		Metadata:     config.Metadata,
	}

	err := ecm.astrocyteNetwork.Register(componentInfo)
	if err != nil {
		return fmt.Errorf("astrocyte network registration failed: %w", err)
	}

	// === CHEMICAL SIGNALING INTEGRATION ===
	// Connect neuron to neurotransmitter and neuromodulator systems if it has receptors
	if len(config.Receptors) > 0 {
		err = ecm.chemicalModulator.RegisterTarget(neuron)
		if err != nil {
			return fmt.Errorf("chemical signaling integration failed: %w", err)
		}
	}

	// === ELECTRICAL COUPLING INTEGRATION ===
	// Enable neuron participation in gap junction networks for synchronization
	if len(config.SignalTypes) > 0 {
		ecm.signalMediator.AddListener(config.SignalTypes, neuron)
	}

	// === MICROGLIAL HEALTH MONITORING ===
	// Initialize health monitoring and surveillance by microglial systems
	ecm.microglia.UpdateComponentHealth(neuron.ID(), 0.0, 0)

	return nil
}

// integrateSynapseIntoBiologicalSystems performs the complete biological integration
// of a newly formed synapse into all matrix coordination systems.
//
// BIOLOGICAL INTEGRATION PROCESS:
// This models synaptic maturation where a new synaptic connection becomes
// functionally integrated into the neural circuit:
//
// 1. Spatial Registration: Synapse location recorded for network topology
// 2. Connectivity Mapping: Pre-post synaptic relationship established
// 3. Activity Monitoring: Synaptic transmission tracking activated
// 4. Health Surveillance: Microglial monitoring for synaptic maintenance
// 5. Plasticity Framework: Activity-dependent modification systems enabled
//
// This ensures the synapse can participate in network computation and
// plasticity from the moment it becomes functional.
func (ecm *ExtracellularMatrix) integrateSynapseIntoBiologicalSystems(synapse SynapseInterface, config SynapseConfig) error {
	// === ASTROCYTE SPATIAL REGISTRATION ===
	// Register synapse location and connectivity with astrocyte network mapping
	componentInfo := ComponentInfo{
		ID:           synapse.ID(),
		Type:         ComponentSynapse,
		Position:     synapse.Position(),
		State:        StateActive,
		RegisteredAt: time.Now(),
		Metadata:     config.Metadata,
	}

	err := ecm.astrocyteNetwork.Register(componentInfo)
	if err != nil {
		return fmt.Errorf("astrocyte network registration failed: %w", err)
	}

	// === CONNECTIVITY MAPPING ===
	// Record the synaptic connection in the network topology for circuit analysis
	err = ecm.astrocyteNetwork.RecordSynapticActivity(
		synapse.ID(),
		config.PresynapticID,
		config.PostsynapticID,
		config.InitialWeight,
	)
	if err != nil {
		return fmt.Errorf("synaptic connectivity mapping failed: %w", err)
	}

	// === MICROGLIAL HEALTH MONITORING ===
	// Initialize health monitoring for synaptic maintenance and pruning oversight
	ecm.microglia.UpdateComponentHealth(synapse.ID(), 0.0, 1)

	return nil
}

// =================================================================================
// MATRIX LIFECYCLE MANAGEMENT
// =================================================================================

// Start initiates all biological coordination services in the proper developmental sequence.
//
// BIOLOGICAL STARTUP SEQUENCE:
// This models the activation of the neural microenvironment, following the biological
// sequence observed during neural development and network activation:
//
// 1. Chemical Systems Activation: Neurotransmitter synthesis and clearance systems
// 2. Electrical Systems Online: Gap junction communication infrastructure
// 3. Spatial Coordination Active: 3D tissue organization and delay calculation
// 4. Health Monitoring Begin: Microglial surveillance and maintenance
// 5. Component Integration Ready: Systems ready for neurogenesis/synaptogenesis
//
// This ensures all biological support systems are operational before any
// neural components begin functioning, preventing developmental failures.
// FIXED: Robust error handling that continues starting remaining components and provides detailed error reporting
func (ecm *ExtracellularMatrix) Start() error {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	if ecm.started {
		return nil // Already active - biological systems don't restart
	}

	// Activate chemical signaling systems (neurotransmitter metabolism)
	err := ecm.chemicalModulator.Start()
	if err != nil {
		return fmt.Errorf("failed to activate chemical signaling systems: %w", err)
	}

	// Start all created neurons with robust error handling
	var startupErrors []string
	successfulStarts := 0

	for neuronID, neuron := range ecm.neurons {
		if err := neuron.Start(); err != nil {
			startupErrors = append(startupErrors, fmt.Sprintf("neuron %s: %v", neuronID, err))
		} else {
			successfulStarts++
		}
	}

	// Evaluate startup results
	totalNeurons := len(ecm.neurons)

	if len(startupErrors) > 0 {
		if successfulStarts == 0 {
			// Complete failure - shutdown chemical systems and abort
			ecm.chemicalModulator.Stop()
			return fmt.Errorf("complete neuron startup failure - no neurons started successfully: %v", startupErrors)
		} else if len(startupErrors) == totalNeurons {
			// All neurons failed but somehow successfulStarts != 0 (shouldn't happen, but be safe)
			ecm.chemicalModulator.Stop()
			return fmt.Errorf("all %d neurons failed to start: %v", totalNeurons, startupErrors)
		} else {
			// Partial failure - log errors but continue with successful neurons
			// In biological systems, some neurons may fail while others continue
			fmt.Printf("Warning: %d of %d neurons failed to start (continuing with %d successful): %v\n",
				len(startupErrors), totalNeurons, successfulStarts, startupErrors)
		}
	}

	ecm.started = true

	// If there were partial failures, return a non-fatal error with details
	if len(startupErrors) > 0 {
		return fmt.Errorf("partial startup success: %d of %d neurons started successfully, failures: %v",
			successfulStarts, totalNeurons, startupErrors)
	}

	return nil
}

// Stop gracefully shuts down all biological coordination services and components.
//
// BIOLOGICAL SHUTDOWN SEQUENCE:
// This models controlled cessation of neural activity, similar to anesthesia
// or controlled brain shutdown during hibernation:
//
// 1. Component Deactivation: All neurons and synapses cease activity
// 2. Chemical System Shutdown: Neurotransmitter synthesis and clearance stop
// 3. Electrical Isolation: Gap junction networks disconnect
// 4. Monitoring Cessation: Microglial surveillance ends
// 5. Resource Release: All biological resources returned to system
//
// This ensures clean shutdown without leaving hanging processes or corrupted state.
func (ecm *ExtracellularMatrix) Stop() error {
	ecm.mu.Lock()
	defer ecm.mu.Unlock()

	if !ecm.started {
		return nil // Already inactive
	}

	// Deactivate all neural components in reverse activation order
	for _, neuron := range ecm.neurons {
		neuron.Stop()
	}

	// Shutdown chemical signaling systems
	ecm.chemicalModulator.Stop()

	// Signal coordination shutdown
	ecm.cancel()
	ecm.started = false
	return nil
}

// =================================================================================
// BIOLOGICAL SPATIAL COORDINATION AND TIMING
// =================================================================================

// EnhanceSynapticDelay calculates biologically realistic transmission delays including
// both axonal conduction time and synaptic processing time.
//
// BIOLOGICAL TIMING COMPONENTS:
// This models the complete temporal dynamics of neural signal transmission:
//
// 1. Axonal Conduction Delay: Time for action potential to travel down the axon
//   - Depends on axon diameter, myelination, and distance
//   - Range: 0.1ms (local) to 50ms (long-distance projections)
//
// 2. Synaptic Processing Delay: Time for vesicle fusion and postsynaptic response
//   - Vesicle fusion: ~0.2-0.5ms
//   - Neurotransmitter diffusion: ~0.1-0.3ms
//   - Receptor binding and channel opening: ~0.1-0.2ms
//   - Total synaptic delay: ~0.5-2ms
//
// This function combines these delays to provide realistic neural timing that
// enables proper temporal coding and synchronization in neural circuits.
func (ecm *ExtracellularMatrix) EnhanceSynapticDelay(
	preNeuronID, postNeuronID, synapseID string,
	baseSynapticDelay time.Duration) time.Duration {

	// Retrieve spatial positions of connected neurons
	preInfo, preExists := ecm.astrocyteNetwork.Get(preNeuronID)
	postInfo, postExists := ecm.astrocyteNetwork.Get(postNeuronID)

	if !preExists || !postExists {
		// If spatial information unavailable, return only the synaptic component
		return baseSynapticDelay
	}

	// Calculate 3D Euclidean distance for axonal length estimation
	distance := ecm.calculateSpatialDistance(preInfo.Position, postInfo.Position)

	// Convert spatial distance to axonal conduction delay
	spatialDelay := ecm.calculatePropagationDelay(distance)

	// Return combined delay: synaptic processing + axonal conduction
	return baseSynapticDelay + spatialDelay
}

// GetSpatialDistance returns the 3D distance between two neural components,
// modeling the physical separation that determines axonal length and conduction time.
//
// BIOLOGICAL SIGNIFICANCE:
// In neural tissue, the physical distance between neurons directly determines:
// - Axonal length and conduction delay
// - Metabolic cost of maintaining connections
// - Probability of successful synapse formation
// - Strength of electrical field coupling
// - Efficiency of chemical signal diffusion
//
// This measurement is essential for realistic neural timing and connectivity patterns.
func (ecm *ExtracellularMatrix) GetSpatialDistance(componentID1, componentID2 string) (float64, error) {
	info1, exists1 := ecm.astrocyteNetwork.Get(componentID1)
	info2, exists2 := ecm.astrocyteNetwork.Get(componentID2)

	if !exists1 {
		return 0, fmt.Errorf("neural component not found in tissue: %s", componentID1)
	}
	if !exists2 {
		return 0, fmt.Errorf("neural component not found in tissue: %s", componentID2)
	}

	return ecm.calculateSpatialDistance(info1.Position, info2.Position), nil
}

// SetAxonSpeed configures the global axonal conduction velocity for all neural connections.
//
// BIOLOGICAL PARAMETER CONTROL:
// This allows simulation of different neural tissue types and conditions:
// - Developmental changes (axons speed up with myelination)
// - Species differences (larger animals have faster axons)
// - Pathological conditions (demyelinating diseases slow conduction)
// - Temperature effects (cold slows, warmth speeds conduction)
// - Pharmacological interventions (local anesthetics block conduction)
//
// Parameters:
//   - speedUmPerMs: Conduction velocity in micrometers per millisecond
func (ecm *ExtracellularMatrix) SetAxonSpeed(speedUmPerMs float64) {
	axonSpeedMutex.Lock()
	defer axonSpeedMutex.Unlock()
	globalAxonSpeed = speedUmPerMs
}

// GetAxonSpeed returns the current axonal conduction velocity setting.
//
// BIOLOGICAL MONITORING:
// Allows components to query the current conduction speed for:
// - Temporal coordination calculations
// - Synchronization window estimation
// - Network timing analysis
// - Performance optimization
func (ecm *ExtracellularMatrix) GetAxonSpeed() float64 {
	axonSpeedMutex.RLock()
	defer axonSpeedMutex.RUnlock()
	return globalAxonSpeed
}

// SetBiologicalAxonType configures realistic conduction velocities based on biological axon types.
//
// BIOLOGICAL AXON CLASSIFICATION:
// Different types of axons have characteristic conduction velocities based on:
// - Diameter (larger = faster due to lower resistance)
// - Myelination (myelin sheaths enable saltatory conduction)
// - Function (motor axons are faster than sensory, pain fibers are slowest)
//
// AXON TYPE SPECIFICATIONS:
// - "unmyelinated_slow": Pain and temperature fibers (C fibers) - 0.5-2 m/s
// - "unmyelinated_fast": Local cortical circuits - 2-5 m/s
// - "cortical_local": Within-column connections - 2-5 m/s
// - "cortical_inter": Between cortical layers - 5-15 m/s
// - "long_range": Cortical-cortical projections - 15-30 m/s
// - "myelinated_medium": Sensory fibers (A-delta) - 5-30 m/s
// - "myelinated_fast": Motor fibers (A-alpha) - 30-120 m/s
//
// Parameters:
//   - axonType: Biological classification string determining conduction speed
func (ecm *ExtracellularMatrix) SetBiologicalAxonType(axonType string) {
	switch axonType {
	case "unmyelinated_slow":
		ecm.SetAxonSpeed(UNMYELINATED_SLOW)
	case "unmyelinated_fast":
		ecm.SetAxonSpeed(UNMYELINATED_FAST)
	case "cortical_local":
		ecm.SetAxonSpeed(LOCAL_CIRCUIT)
	case "cortical_inter":
		ecm.SetAxonSpeed(INTER_LAMINAR)
	case "long_range":
		ecm.SetAxonSpeed(LONG_RANGE)
	case "myelinated_medium":
		ecm.SetAxonSpeed(MYELINATED_MEDIUM)
	case "myelinated_fast":
		ecm.SetAxonSpeed(MYELINATED_FAST)
	default:
		ecm.SetAxonSpeed(LOCAL_CIRCUIT) // Default to cortical local circuits
	}
}

// calculateSpatialDistance computes 3D Euclidean distance between neural components.
//
// BIOLOGICAL CALCULATION:
// This models the straight-line distance through neural tissue, which approximates
// the path length for axonal connections in dense neural tissue where axons
// tend to take relatively direct routes to their targets.
//
// In biological tissue, actual axonal paths may be longer due to:
// - Fasciculation (axons bundling together)
// - Anatomical constraints (avoiding certain brain regions)
// - Developmental guidance cues (following molecular gradients)
//
// For most computational purposes, Euclidean distance provides a good approximation.
func (ecm *ExtracellularMatrix) calculateSpatialDistance(pos1, pos2 Position3D) float64 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	dz := pos1.Z - pos2.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// calculatePropagationDelay converts spatial distance to temporal delay based on axonal properties.
//
// BIOLOGICAL CONDUCTION PHYSICS:
// Axonal conduction follows the cable equation from biophysics:
// - Conduction velocity = sqrt(diameter / resistance * capacitance)
// - Myelinated axons use saltatory conduction (jumping between nodes)
// - Temperature affects conduction (Q10 effect, ~2-3x per 10°C)
// - Pathology can dramatically slow conduction
//
// This simplified model uses a constant velocity but could be enhanced with:
// - Diameter-dependent speeds
// - Temperature corrections
// - Fatigue effects during high-frequency stimulation
func (ecm *ExtracellularMatrix) calculatePropagationDelay(distance float64) time.Duration {
	if distance <= 0 {
		return 0
	}

	// Thread-safe access to global conduction velocity
	axonSpeedMutex.RLock()
	speed := globalAxonSpeed
	axonSpeedMutex.RUnlock()

	// Physics: delay = distance / velocity
	delayMs := distance / speed

	// Convert to Go time.Duration
	return time.Duration(delayMs * float64(time.Millisecond))
}

// =================================================================================
// CHEMICAL AND ELECTRICAL SIGNALING (BIOLOGICAL COMMUNICATION)
// =================================================================================

// ReleaseLigand initiates biological chemical signaling through neurotransmitter release.
//
// BIOLOGICAL CHEMICAL SIGNALING:
// This models the fundamental process of chemical communication between neurons:
//
// 1. Vesicle Fusion: Calcium-triggered exocytosis releases neurotransmitters
// 2. Diffusion: Chemical signals spread through extracellular space
// 3. Receptor Binding: Neurotransmitters bind to specific receptor proteins
// 4. Signal Transduction: Binding triggers cellular responses
// 5. Clearance: Reuptake transporters and enzymes remove neurotransmitters
//
// NEUROTRANSMITTER TYPES SUPPORTED:
// - Glutamate: Primary excitatory neurotransmitter
// - GABA: Primary inhibitory neurotransmitter
// - Dopamine: Reward and motor control neuromodulator
// - Serotonin: Mood and behavioral state neuromodulator
// - Acetylcholine: Attention and autonomic control neurotransmitter
//
// This enables realistic chemical coordination between neural components.
func (ecm *ExtracellularMatrix) ReleaseLigand(ligandType LigandType, sourceID string, concentration float64) error {
	return ecm.chemicalModulator.Release(ligandType, sourceID, concentration)
}

// RegisterForBinding connects a neural component to the chemical signaling system.
//
// BIOLOGICAL RECEPTOR EXPRESSION:
// This models how neurons express specific neurotransmitter receptors on their
// membrane surface, enabling them to respond to chemical signals:
//
// - Ionotropic Receptors: Direct ion channel opening (fast, 1-5ms)
// - Metabotropic Receptors: G-protein signaling cascades (slow, 10-1000ms)
// - Receptor Density: Number of receptors determines sensitivity
// - Receptor Affinity: Binding strength determines selectivity
// - Receptor Localization: Synaptic vs extrasynaptic placement
//
// Components implementing BindingTarget interface can receive chemical signals
// based on their receptor expression profile.
func (ecm *ExtracellularMatrix) RegisterForBinding(target BindingTarget) error {
	return ecm.chemicalModulator.RegisterTarget(target)
}

// SendSignal initiates electrical signaling through gap junction networks.
//
// BIOLOGICAL ELECTRICAL SIGNALING:
// This models direct electrical communication between neurons through gap junctions:
//
// 1. Gap Junction Channels: Protein channels connecting cell cytoplasms
// 2. Current Flow: Ions flow directly between coupled cells
// 3. Instantaneous Transmission: No synaptic delay (<0.1ms)
// 4. Bidirectional Communication: Current flows both directions
// 5. Synchronization: Enables coordinated network activity
//
// ELECTRICAL SIGNAL TYPES:
// - Action Potentials: All-or-nothing spike events
// - Subthreshold Potentials: Graded voltage changes
// - Oscillatory Activity: Rhythmic network synchronization
// - State Changes: Functional mode transitions
//
// This enables fast electrical coordination for network synchronization.
func (ecm *ExtracellularMatrix) SendSignal(signalType SignalType, sourceID string, data interface{}) {
	ecm.signalMediator.Send(signalType, sourceID, data)
}

// ListenForSignals connects a neural component to the electrical signaling system.
//
// BIOLOGICAL GAP JUNCTION PARTICIPATION:
// This models how neurons participate in electrical networks through gap junctions:
//
// - Connexin Expression: Proteins that form gap junction channels
// - Electrical Coupling: Direct ionic connection between neurons
// - Network Participation: Contribution to synchronized activity
// - Selective Coupling: Only certain neuron types are electrically coupled
// - Development Regulation: Gap junctions change during development
//
// Components implementing SignalListener interface can participate in
// electrical networks for rapid synchronization and coordination.
func (ecm *ExtracellularMatrix) ListenForSignals(signalTypes []SignalType, listener SignalListener) {
	ecm.signalMediator.AddListener(signalTypes, listener)
}

// =================================================================================
// COMPONENT REGISTRATION AND DISCOVERY (BIOLOGICAL ORGANIZATION)
// =================================================================================

// RegisterComponent adds a neural component to the astrocyte spatial organization system.
//
// BIOLOGICAL COMPONENT REGISTRATION:
// This models how astrocytes maintain detailed spatial maps of all neural components
// in their territorial domains:
//
// 1. Spatial Mapping: 3D position registration for distance calculations
// 2. Component Classification: Neuron vs synapse vs glial cell identification
// 3. State Tracking: Active, inactive, or shutting down status
// 4. Connectivity Monitoring: Tracking of synaptic connections
// 5. Health Surveillance: Integration with microglial monitoring
//
// ASTROCYTE TERRITORIAL ORGANIZATION:
// Real astrocytes maintain exclusive territories covering ~100,000 synapses each,
// providing comprehensive monitoring and support for all neural components in
// their domain. This function integrates components into that organizational system.
func (ecm *ExtracellularMatrix) RegisterComponent(info ComponentInfo) error {
	return ecm.astrocyteNetwork.Register(info)
}

// FindComponents performs biological component discovery based on spatial and functional criteria.
//
// BIOLOGICAL COMPONENT DISCOVERY:
// This models how neural components locate and identify other components in their
// vicinity for potential connectivity or coordination:
//
// SPATIAL QUERIES:
// - Proximity Search: Find components within diffusion range
// - Territorial Mapping: Identify components in astrocyte domains
// - Connectivity Analysis: Locate potential synaptic partners
//
// FUNCTIONAL QUERIES:
// - Type-Based Search: Find all neurons or all synapses
// - State-Based Search: Find active or inactive components
// - Property-Based Search: Find components with specific characteristics
//
// This enables dynamic network organization and adaptive connectivity patterns.
func (ecm *ExtracellularMatrix) FindComponents(criteria ComponentCriteria) []ComponentInfo {
	return ecm.astrocyteNetwork.Find(criteria)
}

// =================================================================================
// COMPONENT ACCESS AND MONITORING (BIOLOGICAL TRACKING)
// =================================================================================

// GetNeuron retrieves a created neuron by biological identifier.
//
// BIOLOGICAL COMPONENT ACCESS:
// This models how other neural components can locate and interact with specific
// neurons in the network, similar to how biological neurons use molecular
// address systems to identify their synaptic partners.
func (ecm *ExtracellularMatrix) GetNeuron(neuronID string) (NeuronInterface, bool) {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	neuron, exists := ecm.neurons[neuronID]
	return neuron, exists
}

// GetSynapse retrieves a created synapse by biological identifier.
//
// BIOLOGICAL SYNAPSE ACCESS:
// This models how neural components can locate specific synaptic connections
// for monitoring, modification, or analysis purposes.
func (ecm *ExtracellularMatrix) GetSynapse(synapseID string) (SynapseInterface, bool) {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	synapse, exists := ecm.synapses[synapseID]
	return synapse, exists
}

// ListNeurons returns all active neurons in the biological network.
//
// BIOLOGICAL NETWORK CENSUS:
// This models a complete survey of neural population, useful for:
// - Network analysis and characterization
// - Population dynamics studies
// - Resource allocation assessment
// - Health monitoring coordination
func (ecm *ExtracellularMatrix) ListNeurons() []NeuronInterface {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	neurons := make([]NeuronInterface, 0, len(ecm.neurons))
	for _, neuron := range ecm.neurons {
		neurons = append(neurons, neuron)
	}
	return neurons
}

// ListSynapses returns all active synapses in the biological network.
//
// BIOLOGICAL CONNECTIVITY CENSUS:
// This models a complete survey of synaptic connectivity, useful for:
// - Connectivity pattern analysis
// - Synaptic plasticity studies
// - Network topology characterization
// - Pruning candidate identification
func (ecm *ExtracellularMatrix) ListSynapses() []SynapseInterface {
	ecm.mu.RLock()
	defer ecm.mu.RUnlock()

	synapses := make([]SynapseInterface, 0, len(ecm.synapses))
	for _, synapse := range ecm.synapses {
		synapses = append(synapses, synapse)
	}
	return synapses
}

// =================================================================================
// BIOLOGICAL IDENTIFIER GENERATION
// =================================================================================

// generateBiologicalNeuronID creates a unique identifier for a neuron based on biological principles.
//
// BIOLOGICAL NAMING CONVENTION:
// This models how biological neurons might be identified in neural tissue:
// - Cell Type Prefix: Indicates the functional class of neuron
// - Temporal Stamp: Reflects the developmental timing of neurogenesis
// - Uniqueness Guarantee: Ensures no two neurons share the same identifier
//
// Example: "pyramidal_l5_1716838290123456789" indicates a layer 5 pyramidal
// neuron created at a specific developmental timepoint.
func (ecm *ExtracellularMatrix) generateBiologicalNeuronID(neuronType string) string {
	return fmt.Sprintf("%s_%d", neuronType, time.Now().UnixNano())
}

// generateBiologicalSynapseID creates a unique identifier for a synapse based on connectivity.
//
// BIOLOGICAL SYNAPSE NAMING:
// This models how synaptic connections might be identified in neural circuits:
// - Synapse Type: Indicates the functional class of connection
// - Connectivity Pattern: Shows the pre- and post-synaptic neurons
// - Formation Time: Reflects when synaptogenesis occurred
//
// Example: "excitatory_plastic_neuron1_to_neuron2_1716838290123456789"
// indicates an excitatory plastic synapse from neuron1 to neuron2.
func (ecm *ExtracellularMatrix) generateBiologicalSynapseID(synapseType, preID, postID string) string {
	return fmt.Sprintf("%s_%s_to_%s_%d", synapseType, preID, postID, time.Now().UnixNano())
}

// =================================================================================
// DEFAULT BIOLOGICAL FACTORY REGISTRATION
// =================================================================================

// registerDefaultBiologicalFactories registers built-in neurogenesis and synaptogenesis programs.
//
// BIOLOGICAL FACTORY PROGRAMS:
// This models the genetic and developmental programs that create different types
// of neural components during brain development:
//
// NEUROGENESIS PROGRAMS:
// - Basic Neuron: Generic neuron with standard biological properties
// - Pyramidal L5: Large projection neurons found in cortical layer 5
// - Fast-Spiking Interneuron: Parvalbumin-positive inhibitory neurons
//
// SYNAPTOGENESIS PROGRAMS:
// - Excitatory Plastic: Glutamatergic synapses with STDP plasticity
// - Inhibitory Static: GABAergic synapses with fixed inhibitory strength
//
// These default factories provide standard biological components that can be
// used immediately without custom factory registration.
func (ecm *ExtracellularMatrix) registerDefaultBiologicalFactories() {
	// === NEUROGENESIS PROGRAM REGISTRATION ===
	ecm.neuronFactories["basic"] = createBasicBiologicalNeuron
	ecm.neuronFactories["pyramidal_l5"] = createPyramidalLayer5Neuron
	ecm.neuronFactories["fast_spiking_interneuron"] = createFastSpikingInterneuron

	// === SYNAPTOGENESIS PROGRAM REGISTRATION ===
	ecm.synapseFactories["excitatory_plastic"] = createExcitatoryPlasticSynapse
	ecm.synapseFactories["inhibitory_static"] = createInhibitoryStaticSynapse
}

// =================================================================================
// PLACEHOLDER FACTORY FUNCTIONS (TO BE IMPLEMENTED)
// =================================================================================

// These factory functions would be implemented to create actual neural components
// using the existing neuron and synapse packages. They serve as templates for
// how biological component creation should be structured.

// createBasicBiologicalNeuron creates a standard neuron with basic biological properties
func createBasicBiologicalNeuron(id string, config NeuronConfig, callbacks NeuronCallbacks) (NeuronInterface, error) {
	// TODO: Implement using existing neuron package
	// This would create a neuron.Neuron with the specified configuration
	// and wire it with the provided callbacks for biological coordination
	return nil, fmt.Errorf("basic neuron factory not yet implemented")
}

// createPyramidalLayer5Neuron creates a cortical layer 5 pyramidal neuron
func createPyramidalLayer5Neuron(id string, config NeuronConfig, callbacks NeuronCallbacks) (NeuronInterface, error) {
	// TODO: Implement specialized pyramidal neuron
	// This would create a neuron with layer 5 pyramidal characteristics:
	// - Large soma and extensive dendritic tree
	// - High firing threshold and sustained firing capability
	// - Long-range axonal projections to subcortical targets
	return nil, fmt.Errorf("pyramidal L5 neuron factory not yet implemented")
}

// createFastSpikingInterneuron creates a parvalbumin-positive inhibitory interneuron
func createFastSpikingInterneuron(id string, config NeuronConfig, callbacks NeuronCallbacks) (NeuronInterface, error) {
	// TODO: Implement fast-spiking interneuron
	// This would create a neuron with fast-spiking characteristics:
	// - High firing rates (up to 200Hz)
	// - Short refractory period
	// - GABA release for inhibitory control
	// - Gap junction connections for synchronization
	return nil, fmt.Errorf("fast-spiking interneuron factory not yet implemented")
}

// createExcitatoryPlasticSynapse creates a glutamatergic synapse with plasticity
func createExcitatoryPlasticSynapse(id string, config SynapseConfig, callbacks SynapseCallbacks) (SynapseInterface, error) {
	// TODO: Implement using existing synapse package
	// This would create a synapse.SynapticProcessor with:
	// - Glutamate neurotransmitter release
	// - AMPA/NMDA receptor simulation
	// - STDP (Spike-Timing Dependent Plasticity)
	// - Activity-dependent weight modification
	return nil, fmt.Errorf("excitatory plastic synapse factory not yet implemented")
}

// createInhibitoryStaticSynapse creates a GABAergic synapse with fixed strength
func createInhibitoryStaticSynapse(id string, config SynapseConfig, callbacks SynapseCallbacks) (SynapseInterface, error) {
	// TODO: Implement inhibitory synapse
	// This would create a synapse with:
	// - GABA neurotransmitter release
	// - Fixed inhibitory strength
	// - Fast kinetics for precise timing
	// - No plasticity (stable inhibition)
	return nil, fmt.Errorf("inhibitory static synapse factory not yet implemented")
}
