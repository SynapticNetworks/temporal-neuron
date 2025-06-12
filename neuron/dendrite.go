/*
=================================================================================
DENDRITIC GATES - BIOLOGICAL TRANSIENT REWIRING MECHANISMS
=================================================================================

BIOLOGICAL OVERVIEW:
Gates in biological dendrites represent the sophisticated molecular machinery that
enables dynamic pathway modulation and transient rewiring. Unlike static synaptic
weights, gates provide temporal control over signal flow through dendritic
compartments, implementing the brain's remarkable ability to reconfigure itself
in real-time without permanent structural changes.

KEY BIOLOGICAL MECHANISMS MODELED:
1. METABOTROPIC RECEPTORS (MRs): Detect specific chemical signals and trigger
   intracellular cascades that modulate dendritic processing
2. G PROTEIN-GATED ION CHANNELS (GPGICs): Implement the actual pathway changes,
   remaining active for hundreds of milliseconds to minutes
3. ACTIVITY-DEPENDENT PLASTICITY: Gates learn when and how to modulate based
   on local activity patterns and network feedback

STATEFUL vs STATELESS DISTINCTION:
- STATELESS: Traditional gates (LSTM/GRU) that simply multiply signals by learned weights
- STATEFUL: Biological gates that maintain internal state, detect triggers, and
  dynamically switch between configurations during inference

This implementation supports STATEFUL gating where:
- Gate states change during inference based on biological triggers
- Memory resides in the gate states themselves, not separate hidden variables
- Gates learn both how to behave AND when to change behavior
- Supports multi-level learning and dynamic network reconfiguration

ARCHITECTURAL INTEGRATION:
Gates integrate into the dendritic Handle() phase, operating on individual
synaptic messages before temporal summation. This biological ordering allows
gates to:
- Block or amplify specific inputs based on dendritic state
- Implement compartment-specific processing rules
- Provide context-dependent signal transformation
- Enable branch-level computation and feature detection

=================================================================================
*/

package neuron

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/synapse"
)

// ============================================================================
// GATE STATE AND FEEDBACK STRUCTURES
// ============================================================================

// GateState represents the current internal state of a dendritic gate.
// This models the biological reality that gates are not simple switches but
// complex molecular systems with multiple internal parameters that evolve
// over time based on local conditions and network feedback.
//
// BIOLOGICAL CORRESPONDENCE:
// - G protein activation levels and signaling cascade states
// - Ion channel phosphorylation and modulation states
// - Calcium concentration and metabolic activity levels
// - Recent activity history and adaptation mechanisms
type GateState struct {
	// === ACTIVATION STATE ===
	IsActive        bool      `json:"is_active"`        // Whether gate is currently modulating signals
	ActivationLevel float64   `json:"activation_level"` // Strength of current activation (0.0-1.0)
	ActiveSince     time.Time `json:"active_since"`     // When current activation began

	// === INTERNAL DYNAMICS ===
	InternalState map[string]float64 `json:"internal_state"` // Gate-specific internal parameters
	// Examples: "calcium_level", "g_protein_activity", "phosphorylation_level"

	// === LEARNING AND ADAPTATION ===
	RecentTriggers  []time.Time `json:"recent_triggers"`  // History of recent activations
	AdaptationLevel float64     `json:"adaptation_level"` // How much gate has adapted

	// === TEMPORAL PROPERTIES ===
	Duration      time.Duration `json:"duration"`       // How long gate remains active
	RefractoryEnd time.Time     `json:"refractory_end"` // When gate can activate again
}

// GateFeedback provides information about the effectiveness of gate actions
// for learning and adaptation. This models the biological feedback mechanisms
// that allow gates to learn when and how to modulate dendritic processing.
//
// BIOLOGICAL BASIS:
// - Retrograde signaling from post-synaptic activity
// - Local calcium dynamics and metabolic feedback
// - Network-wide neuromodulatory signals
// - Homeostatic pressure and activity regulation
type GateFeedback struct {
	// === EFFECTIVENESS METRICS ===
	WasHelpful   bool    `json:"was_helpful"`  // Whether gate action contributed to desired outcome
	Contribution float64 `json:"contribution"` // Quantified contribution (-1.0 to +1.0)

	// === CONTEXT INFORMATION ===
	PostSynapticFired bool    `json:"post_synaptic_fired"` // Whether post-synaptic neuron fired
	NetworkActivity   float64 `json:"network_activity"`    // Overall network activity level

	// === LEARNING SIGNALS ===
	RewardSignal float64   `json:"reward_signal"` // Dopaminergic or other reward input
	ErrorSignal  float64   `json:"error_signal"`  // Error-based learning signal
	Timestamp    time.Time `json:"timestamp"`     // When feedback occurred
}

// GateTrigger represents the conditions that can activate or deactivate a gate.
// This models the biological detection mechanisms that determine when dendritic
// gates should change their state.
//
// BIOLOGICAL MECHANISMS:
// - Chemical signal detection (neurotransmitters, neuromodulators)
// - Electrical activity patterns (coincidence detection, bursting)
// - Metabolic state changes (energy availability, stress signals)
// - Network synchronization and oscillatory patterns
type GateTrigger struct {
	// === TRIGGER CONDITIONS ===
	SignalThreshold  float64       `json:"signal_threshold"`  // Minimum signal strength to trigger
	ActivityWindow   time.Duration `json:"activity_window"`   // Time window for activity detection
	CoincidenceCount int           `json:"coincidence_count"` // Number of coincident inputs required

	// === CONTEXT SENSITIVITY ===
	MembraneThreshold float64 `json:"membrane_threshold"` // Required membrane potential
	CalciumThreshold  float64 `json:"calcium_threshold"`  // Required calcium concentration

	// === TEMPORAL PROPERTIES ===
	MinInterval time.Duration `json:"min_interval"` // Minimum time between activations
	MaxDuration time.Duration `json:"max_duration"` // Maximum activation duration
}

// ============================================================================
// CORE DENDRITIC GATE INTERFACE
// ============================================================================

// DendriticGate defines the interface for biological gates that modulate
// dendritic signal processing. This interface captures the essential
// capabilities of biological gating mechanisms while providing flexibility
// for different types of gate implementations.
//
// DESIGN PRINCIPLES:
// 1. BIOLOGICAL REALISM: Methods correspond to actual biological processes
// 2. STATEFUL OPERATION: Gates maintain internal state between invocations
// 3. LEARNING CAPABILITY: Gates can adapt their behavior based on feedback
// 4. TEMPORAL AWARENESS: Gates understand timing and duration constraints
// 5. CONTEXT SENSITIVITY: Gates respond to dendritic and membrane state
//
// USAGE PATTERN:
// 1. Gate.ShouldActivate() checks if current conditions warrant activation
// 2. Gate.Apply() transforms or blocks the synaptic message
// 3. Gate.Update() evolves internal state and learns from feedback
// 4. Gate.GetState() provides introspection for monitoring and debugging
type DendriticGate interface {
	// === CORE GATING FUNCTIONALITY ===

	// Apply modulates an incoming synaptic message based on the gate's current
	// state and the dendritic context. This is the primary computational
	// function that implements the gate's effect on signal processing.
	//
	// BIOLOGICAL PROCESS MODELED:
	// When a synaptic signal arrives at a dendritic branch containing an
	// active gate, the gate's molecular machinery (GPGICs, ion channels,
	// signaling cascades) modulates the signal strength, timing, or even
	// blocks it entirely based on current conditions.
	//
	// Parameters:
	//   msg: The incoming synaptic message to be processed
	//   state: Current membrane and dendritic state for context
	//
	// Returns:
	//   *synapse.SynapseMessage: The modified message (or nil if blocked)
	//   bool: Whether the message should continue to be processed
	//
	// Examples:
	//   - Amplitude modulation: multiply msg.Value by gate activation level
	//   - Temporal filtering: block messages outside specific timing windows
	//   - Threshold gating: only pass messages above certain strengths
	//   - Context-dependent: modify based on membrane potential or calcium
	Apply(msg synapse.SynapseMessage, state MembraneSnapshot) (*synapse.SynapseMessage, bool)

	// === STATE MANAGEMENT ===

	// ShouldActivate determines if the gate should change its activation state
	// based on current conditions. This models the biological trigger detection
	// mechanisms that determine when gates open or close.
	//
	// BIOLOGICAL BASIS:
	// Metabotropic receptors continuously monitor for specific chemical signals,
	// electrical activity patterns, or metabolic changes that indicate the
	// need for pathway modulation. This method captures that detection process.
	//
	// Parameters:
	//   msg: The current synaptic message (trigger signal)
	//   state: Current dendritic and membrane state
	//
	// Returns:
	//   bool: Whether gate should change activation state
	//   time.Duration: How long the new state should last (if activating)
	ShouldActivate(msg synapse.SynapseMessage, state MembraneSnapshot) (bool, time.Duration)

	// Update evolves the gate's internal state based on time passage and
	// feedback about its effectiveness. This implements the biological
	// learning and adaptation mechanisms that allow gates to improve their
	// performance over time.
	//
	// BIOLOGICAL PROCESSES:
	// - Calcium-dependent plasticity and protein synthesis changes
	// - Activity-dependent gene expression and channel modulation
	// - Homeostatic adjustment of trigger thresholds and sensitivity
	// - Long-term potentiation/depression of gating mechanisms
	//
	// Parameters:
	//   feedback: Information about gate effectiveness for learning
	//   deltaTime: Time elapsed since last update for temporal evolution
	Update(feedback *GateFeedback, deltaTime time.Duration)

	// === INTROSPECTION AND MONITORING ===

	// GetState returns the current internal state of the gate for monitoring,
	// debugging, and analysis purposes. This provides full transparency into
	// the gate's operation without affecting its behavior.
	GetState() GateState

	// GetTrigger returns the current trigger configuration, allowing external
	// systems to understand what conditions will activate this gate.
	GetTrigger() GateTrigger

	// === IDENTIFICATION AND LIFECYCLE ===

	// Name returns a human-readable identifier for this gate instance,
	// useful for logging, debugging, and network analysis.
	Name() string

	// Type returns the functional category of this gate (e.g., "threshold",
	// "gain_modulation", "temporal_filter") for classification and analysis.
	Type() string

	// Close releases any resources held by the gate when it's no longer needed.
	// This ensures clean shutdown and prevents resource leaks.
	Close()
}

// NEW: An InputProcessor can modify a single synaptic input value before it is decayed.
// This is a generic, reusable function signature.
type InputProcessorFunc func(value float64) float64

// ============================================================================
// GATE FACTORY AND CONFIGURATION
// ============================================================================

// GateConfig provides the base configuration structure for creating different
// types of dendritic gates. Specific gate types extend this with their own
// parameters while maintaining a common configuration interface.
type GateConfig struct {
	// === BASIC PROPERTIES ===
	Name            string `json:"name"`             // Human-readable identifier
	Type            string `json:"type"`             // Gate functional category
	InitiallyActive bool   `json:"initially_active"` // Whether gate starts active

	// === TEMPORAL CONFIGURATION ===
	DefaultDuration  time.Duration `json:"default_duration"`  // Default activation duration
	RefractoryPeriod time.Duration `json:"refractory_period"` // Minimum time between activations

	// === LEARNING PARAMETERS ===
	LearningEnabled bool    `json:"learning_enabled"` // Whether gate can adapt
	LearningRate    float64 `json:"learning_rate"`    // Rate of adaptation

	// === TRIGGER CONFIGURATION ===
	Trigger GateTrigger `json:"trigger"` // Conditions for activation
}

// ============================================================================
// Data Transfer Objects (DTOs) for Decoupled Communication
// ============================================================================

// MembraneSnapshot provides a read-only snapshot of the neuron's electrical state
// to the integration strategy. This allows the strategy to make context-aware
// computational decisions without having direct access to the neuron's internal
// state, promoting a clean, decoupled architecture.
//
// BIOLOGICAL CONTEXT:
// A neuron's decision to fire an action potential is not solely based on immediate
// inputs. It is influenced by its current state, which is the result of recent
// history. For example, a neuron that has recently fired may be in a state of
// hyperpolarization (lower membrane potential), making it harder to fire again.
// More advanced biological models also include back-propagating action potentials,
// where a spike from the axon can travel backward into the dendrites, momentarily
// changing their electrical state and affecting how they integrate subsequent inputs.
// This snapshot provides the necessary context for the dendrite to perform these
// more complex, state-dependent computations.
type MembraneSnapshot struct {
	// Accumulator represents the current membrane potential (Vm) at the axon
	// hillock—the neuron's trigger zone. It is the summed and integrated result
	// of all postsynaptic potentials that have propagated from the dendrites
	// to the cell body. This is the value that is ultimately compared against
	// the firing threshold.
	Accumulator float64

	// CurrentThreshold represents the neuron's excitability at this moment.
	// In biology, the firing threshold is not static; it can be dynamically
	// adjusted by homeostatic plasticity mechanisms in response to the neuron's
	// recent activity history. Providing this value to the strategy allows it
	// to model state-dependent effects, such as inputs having a greater or lesser
	// impact depending on how close the neuron is to firing.
	CurrentThreshold float64
}

// IntegratedPotential is the result returned by a `DendriticIntegrationMode`.
// It represents the net effect of all synaptic inputs processed within a given
// time step, after all complex dendritic computations (like summation, shunting,
// and non-linear spikes) have been performed.
//
// BIOLOGICAL CONTEXT:
// This value is analogous to the effective current that arrives at the neuron's
// soma (cell body) after traveling through the dendritic tree. It is not a
// simple sum of inputs; it is the final, computed result of the dendrite's
// information processing. The neuron's core machinery takes this value as a
// command: "Change your membrane potential by this much," and then applies it
// to its accumulator, which determines the final firing decision.
type IntegratedPotential struct {
	// NetInput is the single floating-point value representing the total
	// change to be applied to the neuron's accumulator for this time step.
	// A positive value is excitatory, pushing the neuron closer to its firing
	// threshold, while a negative value is inhibitory.
	NetInput float64
}

// ============================================================================
// Core Integration Interface
// ============================================================================

// DendriticIntegrationMode defines the core interface for the Strategy Pattern,
// allowing for different, pluggable methods of synaptic signal integration. This
// interface is central to modeling the diverse computational capabilities of
// biological dendrites.
//
// PURPOSE (SOFTWARE ARCHITECTURE):
// This interface decouples the `Neuron`'s core state and lifecycle management
// from the specific algorithms used to process incoming signals. This makes the
// system highly modular and extensible. We can create new, complex integration
// behaviors (e.g., modeling specific ion channels, different neurotransmitter
// effects) by simply creating a new struct that implements this interface,
// without ever needing to modify the `Neuron`'s code.
//
// PURPOSE (BIOLOGICAL MODELING):
// This interface allows us to represent the vast diversity of dendritic computation
// found in the brain. The dendrites of a cortical pyramidal neuron, which perform
// complex coincidence detection, integrate inputs very differently from the
// compact dendritic structure of a fast-spiking interneuron. This pattern allows
// us to create different modes (`PassiveMembraneMode`, `TemporalSummationMode`, etc.)
// to accurately model these distinct neuronal types.
type DendriticIntegrationMode interface {
	// Handle is called by the neuron for every incoming synaptic message. It models
	// the immediate, local effect of a synaptic event on a dendritic branch.
	// For simple, passive models, it may process the signal and return a result
	// immediately. For more complex, time-dependent models, it will typically
	// buffer the message for later processing and return nil.
	Handle(msg synapse.SynapseMessage) *IntegratedPotential

	// Process is called periodically by the neuron's internal clock (its `Run`
	// loop). It models the slower, integrative processes that occur over the
	// neuron's membrane time constant. This method takes the current state of the
	// neuron's membrane as context, performs its calculations on any buffered
	// messages, and returns the final, integrated potential for that time step.
	Process(state MembraneSnapshot) *IntegratedPotential

	// Name returns a human-readable name for the strategy, useful for logging
	// and debugging.
	Name() string

	// Close allows the strategy to release any resources it might hold, such as
	// tickers or background goroutines, when the neuron is shut down.
	Close()
}

// ============================================================================
// Integration Strategy Implementations
// ============================================================================

// ----------------------------------------------------------------------------
// 1. PassiveMembraneMode (Backward Compatibility)
// ----------------------------------------------------------------------------

// PassiveMembraneMode implements the `DendriticIntegrationMode` interface by
// processing each synaptic input immediately, without buffering. It models a
// simple, passive dendrite that acts as a basic wire for signal collection.
//
// BIOLOGICAL CONTEXT:
// This mode is computationally equivalent to the original neuron implementation.
// It is useful for modeling neurons with very simple dendritic structures or for
// simulating direct inputs to the soma, where temporal integration effects are
// minimal. It serves as the default mode to ensure that existing networks and
// tests that rely on immediate processing continue to function without modification.
type PassiveMembraneMode struct{}

// NewPassiveMembraneMode creates a new instance of the immediate processing strategy.
func NewPassiveMembraneMode() *PassiveMembraneMode {
	return &PassiveMembraneMode{}
}

// Handle immediately packages the incoming message's value into a result for
// the neuron to apply.
func (m *PassiveMembraneMode) Handle(msg synapse.SynapseMessage) *IntegratedPotential {
	return &IntegratedPotential{NetInput: msg.Value}
}

// Process does nothing in this mode, as all processing is handled immediately.
func (m *PassiveMembraneMode) Process(state MembraneSnapshot) *IntegratedPotential {
	return nil
}

// Name returns the identifier for this strategy.
func (m *PassiveMembraneMode) Name() string { return "PassiveMembrane" }

// Close does nothing as there are no background resources to release.
func (m *PassiveMembraneMode) Close() {}

// ----------------------------------------------------------------------------
// 2. TemporalSummationMode (Biologically Realistic Time-based Integration)
// ----------------------------------------------------------------------------

// TemporalSummationMode implements buffered, time-based integration. It collects
// all incoming signals within a discrete time window (defined by the neuron's
// processing ticker) and processes them together as a single batch.
//
// BIOLOGICAL CONTEXT:
// This is the most fundamental improvement over the original model. It directly
// simulates the **membrane time constant**, a key biophysical property that allows
// a neuron to integrate inputs over a brief period (typically 5-20ms). By
// buffering inputs, this mode solves the temporal race condition where an
// excitatory signal could cause a firing event before a simultaneous inhibitory
// signal was processed. This allows for realistic summation of excitatory (EPSPs)
// and inhibitory (IPSPs) postsynaptic potentials.
type TemporalSummationMode struct {
	buffer    []synapse.SynapseMessage
	mutex     sync.Mutex
	gateChain []DendriticGate // Middleware chain
}

// NewTemporalSummationMode creates a new instance of the time-based batching strategy.
func NewTemporalSummationMode() *TemporalSummationMode {
	return &TemporalSummationMode{
		// Pre-allocate a reasonably sized buffer to reduce memory allocations.
		buffer: make([]synapse.SynapseMessage, 0, 100),
	}
}

// Handle adds the message to an internal buffer for processing at the end of the
// current time step. It returns nil to indicate that no immediate action is needed.
func (m *TemporalSummationMode) Handle(msg synapse.SynapseMessage) *IntegratedPotential {
	// === GATE MIDDLEWARE CHAIN ===
	currentMsg := &msg
	for _, gate := range m.gateChain {
		modifiedMsg, shouldContinue := gate.Apply(*currentMsg, MembraneSnapshot{})
		if !shouldContinue {
			return nil // Blocked by this gate in the chain
		}
		currentMsg = modifiedMsg
	}

	m.mutex.Lock()
	m.buffer = append(m.buffer, msg)
	m.mutex.Unlock()
	return nil
}

// Setter for gate chain
func (m *TemporalSummationMode) SetGates(gates []DendriticGate) {
	m.gateChain = gates
}

// Add single gate to chain
func (m *TemporalSummationMode) AddGate(gate DendriticGate) {
	m.gateChain = append(m.gateChain, gate)
}

// Process calculates the simple linear sum of all buffered messages.
func (m *TemporalSummationMode) Process(state MembraneSnapshot) *IntegratedPotential {
	m.mutex.Lock()
	// Early exit if the buffer is empty to avoid unnecessary work.
	if len(m.buffer) == 0 {
		m.mutex.Unlock()
		return nil
	}

	// Copy messages to a local variable to minimize the time the mutex is held.
	currentBatch := make([]synapse.SynapseMessage, len(m.buffer))
	copy(currentBatch, m.buffer)
	m.buffer = m.buffer[:0] // Clear the shared buffer.
	m.mutex.Unlock()

	var totalInput float64
	for _, msg := range currentBatch {
		totalInput += msg.Value
	}

	return &IntegratedPotential{NetInput: totalInput}
}

// Name returns the identifier for this strategy.
func (m *TemporalSummationMode) Name() string { return "TemporalSummation" }

// Close does nothing as there are no background resources to release.
func (m *TemporalSummationMode) Close() {
	// Close all gates in the chain
	for _, gate := range m.gateChain {
		if gate != nil {
			gate.Close()
		}
	}
	// Optional: clear the chain
	m.gateChain = nil
}

// ----------------------------------------------------------------------------
// 3. ShuntingInhibitionMode (Advanced Non-Linear Integration)
// ----------------------------------------------------------------------------

// ShuntingInhibitionMode models the powerful, divisive effect of GABAergic
// inhibition. It builds upon TemporalSummationMode but overrides the processing
// logic to implement a non-linear, multiplicative form of inhibition.
//
// BIOLOGICAL CONTEXT:
// While some inhibition is subtractive, a primary mechanism, especially from
// GABA-A receptors located on the soma and proximal dendrites, is **shunting**.
// When these channels open, they dramatically increase the membrane's conductance
// (i.e., they make it "leakier") to chloride ions. This doesn't just lower the
// voltage; it effectively creates a short-circuit that shunts away the current
// from nearby excitatory synapses. The effect is divisive: it reduces the *impact*
// of excitatory signals, rather than just subtracting from their value. This is
// a highly effective and efficient way to control a neuron's output gain.
type ShuntingInhibitionMode struct {
	BiologicalTemporalSummationMode // Embed the basic temporal summation for its buffering logic.
	ShuntingStrength                float64
}

// NewShuntingInhibitionMode creates a mode that models divisive inhibition.
// strength: A factor (e.g., 0.1-1.0) determining how powerful the shunting effect is.
func NewShuntingInhibitionMode(strength float64, config BiologicalConfig) *ShuntingInhibitionMode {
	if strength <= 0 {
		strength = 0.5 // Default to a reasonable strength.
	}
	return &ShuntingInhibitionMode{
		BiologicalTemporalSummationMode: *NewBiologicalTemporalSummationMode(config),
		ShuntingStrength:                strength,
	}
}

func (m *ShuntingInhibitionMode) SetGates(gates []DendriticGate) {
	m.gateChain = gates
}

func (m *ShuntingInhibitionMode) AddGate(gate DendriticGate) {
	m.gateChain = append(m.gateChain, gate)
}

// Process overrides the embedded Process method to implement shunting logic.
// AFTER (Corrected): Shunting mode's Process method using the complete helper.
func (m *ShuntingInhibitionMode) Process(state MembraneSnapshot) *IntegratedPotential {
	// Step 1: Call the reusable helper.
	// This mode has no special per-input logic, so it passes `nil` for the processor.
	totalExcitation, totalInhibition := m.processDecayedComponents(time.Now(), nil)

	// Early exit if no significant input after decay.
	if math.Abs(totalExcitation) < 0.001 && math.Abs(totalInhibition) < 0.001 {
		return nil
	}

	// Step 2: Apply the shunting logic to the decayed, summed inputs.
	shuntingFactor := 1.0 - (totalInhibition * m.ShuntingStrength)
	if shuntingFactor < 0.1 {
		shuntingFactor = 0.1
	}

	netInput := totalExcitation * shuntingFactor

	return &IntegratedPotential{NetInput: netInput}
}

// Name returns the identifier for this strategy.
func (m *ShuntingInhibitionMode) Name() string { return "ShuntingInhibition" }

// ----------------------------------------------------------------------------
// 4. ActiveDendriteMode (Advanced Compartmental Integration)
// ----------------------------------------------------------------------------

// ActiveDendriteMode provides a comprehensive model of a computationally active
// dendrite, incorporating multiple non-linear mechanisms. It demonstrates the
// full power of the strategy pattern by combining temporal summation with synaptic
// saturation, shunting inhibition, and NMDA-like dendritic spikes.
//
// BIOLOGICAL CONTEXT:
// This mode models a cortical pyramidal neuron's dendrite, which is not a passive
// receiver but an active computational unit. It integrates the following key features:
//   - **Synaptic Saturation:** Models the physical limit of a synapse's influence,
//     preventing any single input from being unrealistically powerful.
//   - **Shunting Inhibition:** Models the divisive, gain-control effects of GABA.
//   - **NMDA-like Dendritic Spikes:** Models the all-or-nothing regenerative
//     potentials triggered by coincident excitatory input. This is a crucial
//     mechanism for "feature binding" and Hebbian learning ("cells that fire
//     together, wire together").
type ActiveDendriteMode struct {
	BiologicalTemporalSummationMode                      // Embed buffering logic.
	Config                          ActiveDendriteConfig // Holds all specific parameters for this mode

}

// ActiveDendriteConfig holds the parameters for the ActiveDendriteMode.
type ActiveDendriteConfig struct {
	MaxSynapticEffect       float64
	ShuntingStrength        float64
	DendriticSpikeThreshold float64
	NMDASpikeAmplitude      float64
}

// NewActiveDendriteMode creates a new instance of the advanced integration mode.
func NewActiveDendriteMode(config ActiveDendriteConfig, bioConfig BiologicalConfig) *ActiveDendriteMode {
	// Provide sensible defaults if values are zero.
	if config.MaxSynapticEffect <= 0 {
		config.MaxSynapticEffect = 2.0
	}
	if config.ShuntingStrength <= 0 {
		config.ShuntingStrength = 0.5
	}
	if config.DendriticSpikeThreshold <= 0 {
		config.DendriticSpikeThreshold = 1.5
	}
	if config.NMDASpikeAmplitude <= 0 {
		config.NMDASpikeAmplitude = 1.0
	}
	return &ActiveDendriteMode{
		BiologicalTemporalSummationMode: *NewBiologicalTemporalSummationMode(bioConfig),
		Config:                          config, // Assign the config struct directly
	}
}

// Setter for gate chain
func (m *ActiveDendriteMode) SetGates(gates []DendriticGate) {
	m.gateChain = gates
}

// Add single gate to chain
func (m *ActiveDendriteMode) AddGate(gate DendriticGate) {
	m.gateChain = append(m.gateChain, gate)
}

// Process implements the full, non-linear integration logic.
// AFTER: The refactored Process method for ActiveDendriteMode.
// AFTER (Corrected): The final, clean Process method.
func (m *ActiveDendriteMode) Process(state MembraneSnapshot) *IntegratedPotential {
	// Step 1: Define the saturation logic specific to this mode as a function.
	saturator := func(value float64) float64 {
		if value > m.Config.MaxSynapticEffect {
			return m.Config.MaxSynapticEffect
		}
		if value < -m.Config.MaxSynapticEffect {
			return -m.Config.MaxSynapticEffect
		}
		return value
	}

	// Step 2: Call the reusable helper, passing in the mode-specific saturation function.
	totalExcitation, totalInhibition := m.processDecayedComponents(time.Now(), saturator)

	// Early exit if no significant input after decay and saturation.
	if math.Abs(totalExcitation) < 0.001 && math.Abs(totalInhibition) < 0.001 {
		return nil
	}

	// Step 3: Apply Shunting Inhibition to the decayed, saturated, and summed inputs.
	shuntingFactor := 1.0 - (totalInhibition * m.Config.ShuntingStrength)
	if shuntingFactor < 0.1 {
		shuntingFactor = 0.1
	}
	netExcitation := totalExcitation * shuntingFactor

	// Step 4: Model NMDA-like Dendritic Spikes on the shunted potential.
	if netExcitation > m.Config.DendriticSpikeThreshold {
		netExcitation += m.Config.NMDASpikeAmplitude
	}

	return &IntegratedPotential{NetInput: netExcitation}
}

// Name returns the identifier for this strategy.
func (m *ActiveDendriteMode) Name() string { return "ActiveDendrite" }

/*
BIOLOGICALLY REALISTIC TEMPORAL SUMMATION WITH EXPONENTIAL DECAY
=================================================================

FIXED IMPLEMENTATION: Proper exponential temporal decay
This corrects the accumulation issue and implements true biological membrane dynamics

BIOLOGICAL REALITY:
- Membrane time constant τ = Rm × Cm (typically 10-50ms)
- PSPs decay exponentially: V(t) = V₀ × e^(-t/τ)
- Each input contributes according to its age since arrival
- Continuous membrane potential evolution between Process() calls

MATHEMATICAL MODEL:
V(t) = V₀ × e^(-t/τ) where:
- V₀ = initial PSP amplitude
- t = time since PSP arrival
- τ = membrane time constant
- e = Euler's number (2.718...)

EXPECTED TEST RESULTS:
- Delay 0ms: integration ≈ 1.000 (no decay)
- Delay 5ms: integration ≈ 0.779 (τ/4 decay)
- Delay 10ms: integration ≈ 0.607 (τ/2 decay)
- Delay 20ms: integration ≈ 0.368 (1/e decay at τ)
- Delay 40ms: integration ≈ 0.135 (2τ decay)
- Delay 100ms: integration ≈ 0.007 (5τ decay)
*/

// BiologicalTemporalSummationMode implements realistic dendritic integration
// with exponential decay, membrane time constants, and temporal dynamics
type BiologicalTemporalSummationMode struct {
	// === TEMPORAL DECAY PARAMETERS ===
	membraneTimeConstant time.Duration // τ = Rm × Cm (biological: 10-50ms)
	leakConductance      float64       // Membrane leak (biological: 0.95-0.99)

	// === BUFFERED INPUTS WITH TIMESTAMPS ===
	buffer          []TimestampedInput
	bufferMutex     sync.Mutex
	lastProcessTime time.Time

	// === DENDRITIC HETEROGENEITY ===
	branchTimeConstants map[string]time.Duration // Different τ per branch
	spatialDecayFactor  float64                  // Distance-dependent attenuation

	// === BIOLOGICAL NOISE AND VARIATION ===
	membraneNoise  float64       // Thermal and channel noise
	temporalJitter time.Duration // Realistic timing variability
	noiseSeed      int64         // Deterministic noise seed for reproducible trials

	// === GATE MIDDLEWARE CHAIN ===
	gateChain []DendriticGate // Middleware chain for signal processing
}

// TimestampedInput represents a synaptic input with precise timing
type TimestampedInput struct {
	Message     synapse.SynapseMessage
	ArrivalTime time.Time
	DecayFactor float64 // Pre-computed spatial factor
}

// NewBiologicalTemporalSummationMode creates realistic dendritic integration
func NewBiologicalTemporalSummationMode(config BiologicalConfig) *BiologicalTemporalSummationMode {
	return &BiologicalTemporalSummationMode{
		membraneTimeConstant: config.MembraneTimeConstant,
		leakConductance:      config.LeakConductance,
		buffer:               make([]TimestampedInput, 0, 100),
		lastProcessTime:      time.Now(),
		branchTimeConstants:  config.BranchTimeConstants,
		spatialDecayFactor:   config.SpatialDecayFactor,
		membraneNoise:        config.MembraneNoise,
		temporalJitter:       config.TemporalJitter,
		noiseSeed:            time.Now().UnixNano(), // Initialize with unique seed
		gateChain:            make([]DendriticGate, 0),
	}
}

type BiologicalConfig struct {
	MembraneTimeConstant time.Duration            // τ = Rm × Cm (10-50ms typical)
	LeakConductance      float64                  // 0.95-0.99 per ms (not used in corrected version)
	BranchTimeConstants  map[string]time.Duration // Heterogeneous dendrites
	SpatialDecayFactor   float64                  // Distance attenuation
	MembraneNoise        float64                  // Biological noise level
	TemporalJitter       time.Duration            // Timing variability
}

// Handle buffers input with timestamp for realistic temporal processing
func (m *BiologicalTemporalSummationMode) Handle(msg synapse.SynapseMessage) *IntegratedPotential {
	// === GATE MIDDLEWARE CHAIN ===
	currentMsg := &msg
	for _, gate := range m.gateChain {
		modifiedMsg, shouldContinue := gate.Apply(*currentMsg, MembraneSnapshot{})
		if !shouldContinue {
			return nil // Blocked by this gate in the chain
		}
		currentMsg = modifiedMsg
	}

	now := time.Now()

	// Apply spatial decay based on source branch
	spatialWeight := m.calculateSpatialWeight(currentMsg.SourceID)

	// Create timestamped input with biological modifications
	input := TimestampedInput{
		Message:     *currentMsg,
		ArrivalTime: now,
		DecayFactor: spatialWeight,
	}

	// Add temporal jitter for biological realism
	if m.temporalJitter > 0 {
		jitter := time.Duration(rand.NormFloat64() * float64(m.temporalJitter))
		input.ArrivalTime = input.ArrivalTime.Add(jitter)
	}

	m.bufferMutex.Lock()
	m.buffer = append(m.buffer, input)
	m.bufferMutex.Unlock()

	return nil // Process during Process() call
}

// Process performs biologically realistic temporal integration with exponential decay
// AFTER (Corrected): The public Process method for the base mode.
func (m *BiologicalTemporalSummationMode) Process(state MembraneSnapshot) *IntegratedPotential {
	// Step 1: Call the reusable helper.
	// This base mode has no special per-input logic (like saturation), so it passes `nil` for the processor.
	totalExcitation, totalInhibition := m.processDecayedComponents(time.Now(), nil)

	// Step 2: Combine the results. Inhibition is subtractive.
	netInput := totalExcitation - totalInhibition

	// Step 3: Apply the final biological constraints from the original method.
	if netInput > 100.0 { // Biological maximum
		netInput = 100.0
	} else if netInput < -100.0 { // Biological minimum
		netInput = -100.0
	}

	// Step 4: Only return significant changes to avoid floating-point noise.
	if math.Abs(netInput) < 0.001 {
		return nil
	}

	return &IntegratedPotential{NetInput: netInput}
}

// A new, reusable helper method on BiologicalTemporalSummationMode
// This logic is extracted from the original BiologicalTemporalSummationMode.Process method.
// AFTER (Corrected): The complete, reusable helper with all timing and noise logic.
// This logic is a direct and complete extraction from BiologicalTemporalSummationMode.Process.
func (m *BiologicalTemporalSummationMode) processDecayedComponents(now time.Time, processor InputProcessorFunc) (totalExcitation, totalInhibition float64) {
	m.bufferMutex.Lock()
	defer m.bufferMutex.Unlock()

	if len(m.buffer) == 0 {
		m.lastProcessTime = now
		return 0.0, 0.0
	}

	// --- The following is the complete logic from the original Process method, ---
	// --- refactored to use the processor function correctly. ---

	var mostRecentTime time.Time
	for _, input := range m.buffer {
		if input.ArrivalTime.After(mostRecentTime) {
			mostRecentTime = input.ArrivalTime
		}
	}
	//

	processTime := mostRecentTime
	if processTime.IsZero() {
		processTime = now
	}
	//

	var decayedExcitation, decayedInhibition float64

	for _, input := range m.buffer {
		var ageOfInput time.Duration

		if !mostRecentTime.IsZero() && now.Sub(mostRecentTime) < 5*time.Millisecond {
			ageOfInput = processTime.Sub(input.ArrivalTime)
		} else {
			ageOfInput = now.Sub(input.ArrivalTime)
		}
		//

		if ageOfInput >= 0 {
			// Start with the original value from the message.
			value := input.Message.Value

			// If a processor function was provided, use it to transform the value.
			if processor != nil {
				value = processor(value)
			}

			// --- All subsequent calculations now correctly use the processed `value` ---

			timeConstant := m.getEffectiveTimeConstant(input.Message.SourceID) //
			tauInSeconds := timeConstant.Seconds()

			var temporalDecay float64
			if tauInSeconds > 0 {
				temporalDecay = math.Exp(-ageOfInput.Seconds() / tauInSeconds)
			} else {
				temporalDecay = 0.0
			}
			//

			// CRITICAL FIX: Use the processed `value` here, not `input.Message.Value`.
			effectiveInput := value * input.DecayFactor * temporalDecay

			if m.membraneNoise > 0 {
				nanoTime := float64(input.ArrivalTime.UnixNano())
				seedFactor := float64(m.noiseSeed % 1000000)
				inputIndex := float64(len(m.buffer))
				noise1 := math.Sin(nanoTime*1e-9*11.0 + seedFactor*1e-6)
				noise2 := math.Cos(nanoTime*1e-9*17.0 + seedFactor*1e-5)
				noise3 := math.Sin(seedFactor*0.001 + inputIndex*0.1)
				normalApprox := (noise1 + noise2 + noise3) / 3.0
				noise := normalApprox * m.membraneNoise * 2.0
				effectiveInput += noise
				m.noiseSeed = (m.noiseSeed*1103515245 + 12345) % (1 << 31)
			}
			//

			if effectiveInput >= 0 {
				decayedExcitation += effectiveInput
			} else {
				decayedInhibition += -effectiveInput // Store as a positive value
			}
		}
	}

	m.buffer = m.buffer[:0]
	m.lastProcessTime = now
	//

	return decayedExcitation, decayedInhibition
}

// calculateSpatialWeight models distance-dependent signal attenuation
// λ = sqrt(Rm/Ri) - biological space constant
func (m *BiologicalTemporalSummationMode) calculateSpatialWeight(sourceID string) float64 {
	// Simplified spatial model - could be enhanced with actual positions
	baseWeight := 1.0

	// Apply distance-based decay if spatial information available
	if m.spatialDecayFactor > 0 {
		// Model: different branches have different effective distances
		if sourceID == "distal" {
			baseWeight = baseWeight * 0.5 // 50% attenuation for distal inputs
		} else if sourceID == "proximal" {
			baseWeight *= 1.0 // No attenuation for proximal
		} else {
			baseWeight *= 0.7 // Moderate attenuation for mid-dendrite
		}
	}

	return baseWeight
}

// getEffectiveTimeConstant returns branch-specific time constant
func (m *BiologicalTemporalSummationMode) getEffectiveTimeConstant(sourceID string) time.Duration {
	// Use branch-specific time constant if available
	if branchTau, exists := m.branchTimeConstants[sourceID]; exists {
		return branchTau
	}

	// Use default membrane time constant
	return m.membraneTimeConstant
}

// SetGates configures the gate middleware chain
func (m *BiologicalTemporalSummationMode) SetGates(gates []DendriticGate) {
	m.gateChain = gates
}

// AddGate adds a single gate to the middleware chain
func (m *BiologicalTemporalSummationMode) AddGate(gate DendriticGate) {
	m.gateChain = append(m.gateChain, gate)
}

// Name returns identifier for this integration mode
func (m *BiologicalTemporalSummationMode) Name() string {
	return "BiologicalTemporalSummation"
}

// Close releases resources and closes all gates
func (m *BiologicalTemporalSummationMode) Close() {
	// Close all gates in the chain
	for _, gate := range m.gateChain {
		if gate != nil {
			gate.Close()
		}
	}

	// Clear buffer and gate chain
	m.bufferMutex.Lock()
	m.buffer = m.buffer[:0]
	m.bufferMutex.Unlock()

	m.gateChain = nil
}

// ProcessImmediate processes all buffered inputs immediately without temporal decay
// This is useful for testing scenarios where immediate processing is expected
func (m *BiologicalTemporalSummationMode) ProcessImmediate() *IntegratedPotential {
	m.bufferMutex.Lock()
	defer m.bufferMutex.Unlock()

	// Early exit if no inputs to process
	if len(m.buffer) == 0 {
		return nil
	}

	var totalInput float64 = 0.0

	for _, input := range m.buffer {
		// Apply spatial decay but no temporal decay (immediate processing)
		effectiveInput := input.Message.Value * input.DecayFactor

		// Add biological membrane noise (deterministic for testing)
		if m.membraneNoise > 0 {
			nanoTime := float64(input.ArrivalTime.UnixNano())
			seedFactor := float64(m.noiseSeed % 1000000)

			// Use multiple frequencies for better distribution
			noise1 := math.Sin(nanoTime*1e-9*11.0 + seedFactor*1e-6)
			noise2 := math.Cos(nanoTime*1e-9*17.0 + seedFactor*1e-5)
			noise3 := math.Sin(seedFactor * 0.001)

			// Create stronger, more proportional noise
			normalApprox := (noise1 + noise2 + noise3) / 3.0
			noise := normalApprox * m.membraneNoise * 2.0
			effectiveInput += noise
		}

		totalInput += effectiveInput
	}

	// Clear the buffer after processing
	m.buffer = m.buffer[:0]

	// Apply biological constraints
	if totalInput > 100.0 {
		totalInput = 100.0
	} else if totalInput < -100.0 {
		totalInput = -100.0
	}

	if math.Abs(totalInput) < 0.001 {
		return nil
	}

	return &IntegratedPotential{NetInput: totalInput}
}

// === FACTORY FUNCTIONS FOR REALISTIC CONFIGURATIONS ===

// CreateCorticalPyramidalConfig returns realistic cortical neuron parameters
func CreateCorticalPyramidalConfig() BiologicalConfig {
	return BiologicalConfig{
		MembraneTimeConstant: 20 * time.Millisecond,                          // Typical cortical τ
		LeakConductance:      0.97,                                           // 3% leak per ms (legacy, not used)
		SpatialDecayFactor:   0.1,                                            // 10% per distance unit
		MembraneNoise:        0.01,                                           // 1% noise level
		TemporalJitter:       time.Duration(0.5 * float64(time.Millisecond)), // 0.5ms timing variability
		BranchTimeConstants: map[string]time.Duration{
			"apical":   25 * time.Millisecond, // Longer for apical dendrites
			"basal":    15 * time.Millisecond, // Shorter for basal dendrites
			"distal":   30 * time.Millisecond, // Longest for distal branches
			"proximal": 10 * time.Millisecond, // Shortest for proximal
		},
	}
}

// CreateHippocampalConfig returns hippocampal CA1 pyramidal parameters
func CreateHippocampalConfig() BiologicalConfig {
	return BiologicalConfig{
		MembraneTimeConstant: 35 * time.Millisecond,                          // Longer τ for hippocampus
		LeakConductance:      0.98,                                           // Less leaky than cortex (legacy)
		SpatialDecayFactor:   0.05,                                           // Less spatial decay
		MembraneNoise:        0.005,                                          // Lower noise
		TemporalJitter:       time.Duration(0.3 * float64(time.Millisecond)), // Tighter timing

		BranchTimeConstants: map[string]time.Duration{
			"apical":   45 * time.Millisecond,
			"basal":    25 * time.Millisecond,
			"distal":   50 * time.Millisecond,
			"proximal": 15 * time.Millisecond,
		},
	}
}

// CreateInterneuronConfig returns fast-spiking interneuron parameters
func CreateInterneuronConfig() BiologicalConfig {
	return BiologicalConfig{
		MembraneTimeConstant: 8 * time.Millisecond,   // Fast τ for interneurons
		LeakConductance:      0.93,                   // More leaky membrane (legacy)
		SpatialDecayFactor:   0.2,                    // Compact dendritic tree
		MembraneNoise:        0.02,                   // Higher noise
		TemporalJitter:       1.0 * time.Millisecond, // More variable timing
		BranchTimeConstants: map[string]time.Duration{
			"dendrite": 8 * time.Millisecond, // Uniform fast dendrites
		},
	}
}
