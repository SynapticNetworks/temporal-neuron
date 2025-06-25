package neuron

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

/*
=================================================================================
DENDRITIC ION CHANNELS - BIOLOGICAL MEMBRANE CONDUCTANCE MECHANISMS
=================================================================================

BIOLOGICAL OVERVIEW:
Ion channels in biological dendrites represent the sophisticated protein complexes that
control ionic flow across the membrane, enabling dynamic modulation of signal
processing. These channels provide temporal and spatial control over dendritic
excitability, implementing the brain's remarkable ability to reconfigure neural
computation in real-time through changes in membrane conductance.

KEY BIOLOGICAL MECHANISMS MODELED:
1. VOLTAGE-GATED ION CHANNELS (VGICs): Respond to membrane potential changes
2. LIGAND-GATED ION CHANNELS (LGICs): Activated by neurotransmitter binding
3. CALCIUM-ACTIVATED ION CHANNELS: Modulated by intracellular Ca2+ levels
4. MECHANOSENSITIVE CHANNELS: Respond to membrane stretch and pressure
5. METABOTROPIC RECEPTOR CASCADES: G-protein coupled modulation of channels

CHANNEL STATES AND DYNAMICS:
- CLOSED: Channel pore is blocked, no ionic flow
- OPEN: Channel pore allows specific ions to flow down electrochemical gradients
- INACTIVATED: Channel temporarily non-responsive after opening
- DESENSITIZED: Reduced responsiveness after prolonged ligand exposure

BIOLOGICAL ION CHANNEL TYPES MODELED:
- Na+ channels: Fast depolarization, action potential initiation
- K+ channels: Repolarization, afterhyperpolarization, gain control
- Ca2+ channels: Intracellular signaling, plasticity triggers
- Cl- channels: Shunting inhibition, hyperpolarization
- Mixed cation channels: Depolarization, coincidence detection

ARCHITECTURAL INTEGRATION:
Ion channels integrate into the dendritic Handle() phase, operating on individual
synaptic messages before temporal summation. This biological ordering allows
channels to:
- Modulate membrane resistance and time constants
- Implement voltage-dependent signal amplification or attenuation
- Provide neurotransmitter-specific response modulation
- Enable branch-level computation and feature detection
- Control calcium influx for plasticity induction

=================================================================================
*/

// ============================================================================
// LOCAL CONSTANTS AND PARAMETERS
// ============================================================================

const (
	// === BUFFER AND PROCESSING LIMITS ===
	DEFAULT_INPUT_BUFFER_SIZE = DENDRITE_BUFFER_DEFAULT_CAPACITY // Default size for input message buffers

	// === NOISE AND RANDOM GENERATION ===
	NOISE_FREQUENCY_1 = 11.0 // Primary noise frequency component
	NOISE_FREQUENCY_2 = 17.0 // Secondary noise frequency component
	NOISE_DIVISOR     = 3.0  // Divisor for noise normalization
	NOISE_SCALE       = 2.0  // Scale factor for final noise amplitude

	// Linear congruential generator constants (standard values)
	LCG_MULTIPLIER = 1103515245
	LCG_INCREMENT  = 12345
	LCG_MODULUS    = 1 << 31

	// === TIME-BASED CALCULATIONS ===
	NANOSECONDS_TO_SECONDS  = 1e-9    // Conversion factor
	MICROSECONDS_CONVERSION = 1e-6    // For seed factor calculation
	MILLISECONDS_CONVERSION = 1e-5    // For secondary seed calculation
	SEED_MODULUS            = 1000000 // Modulus for seed factor
	SEED_SCALE_PRIMARY      = 1e-6    // Scale for primary seed component
	SEED_SCALE_SECONDARY    = 1e-5    // Scale for secondary seed component
	SEED_SCALE_TERTIARY     = 0.001   // Scale for tertiary seed component
)

// ============================================================================
// ION CHANNEL STATE AND FEEDBACK STRUCTURES
// ============================================================================

// ChannelState represents the current state of an ion channel.
// This models the biological reality that ion channels exist in multiple
// conformational states with complex kinetics and modulation mechanisms.
//
// BIOLOGICAL CORRESPONDENCE:
// - Channel protein conformation (closed, open, inactivated)
// - Phosphorylation and regulatory protein binding states
// - Local ionic concentrations and electrochemical gradients
// - Recent activity history and use-dependent modulation
type ChannelState struct {
	// === CHANNEL CONFORMATION ===
	IsOpen           bool      `json:"is_open"`           // Whether channel pore is open
	Conductance      float64   `json:"conductance"`       // Current conductance (0.0-1.0)
	OpenSince        time.Time `json:"open_since"`        // When channel opened
	InactivationTime time.Time `json:"inactivation_time"` // When channel will inactivate

	// === IONIC AND ELECTRICAL STATE ===
	MembraneVoltage      float64 `json:"membrane_voltage"`      // Local membrane potential (mV)
	CalciumLevel         float64 `json:"calcium_level"`         // Local [Ca2+] concentration
	EquilibriumPotential float64 `json:"equilibrium_potential"` // Reversal potential for this channel

	// === MODULATION STATE ===
	PhosphorylationLevel float64            `json:"phosphorylation_level"` // PKA/PKC phosphorylation
	RegulatoryState      map[string]float64 `json:"regulatory_state"`      // Auxiliary subunit states

	// === KINETIC PROPERTIES ===
	RecentOpenings  []time.Time   `json:"recent_openings"`  // History of channel openings
	AverageDuration time.Duration `json:"average_duration"` // Mean open time for this channel

	// === ADAPTATION AND PLASTICITY ===
	UseFrequency    float64 `json:"use_frequency"`    // How often channel has been used
	AdaptationLevel float64 `json:"adaptation_level"` // Long-term adaptation state
}

// ChannelFeedback provides information about the effectiveness of channel activity
// for experience-dependent modulation. This models the biological feedback
// mechanisms that allow ion channels to undergo activity-dependent regulation.
//
// BIOLOGICAL BASIS:
// - Activity-dependent channel trafficking (insertion/removal)
// - Phosphorylation-dependent modulation by kinases/phosphatases
// - Calcium-dependent plasticity of channel properties
// - Homeostatic scaling of channel density and kinetics
type ChannelFeedback struct {
	// === FUNCTIONAL EFFECTIVENESS ===
	ContributedToFiring bool    `json:"contributed_to_firing"` // Whether channel aided action potential
	CurrentContribution float64 `json:"current_contribution"`  // Magnitude of ionic current (pA)

	// === CELLULAR CONTEXT ===
	PostSynapticResponse bool    `json:"post_synaptic_response"` // Whether postsynaptic cell responded
	CalciumInflux        float64 `json:"calcium_influx"`         // Associated Ca2+ entry

	// === PLASTICITY SIGNALS ===
	PKAActivity    float64   `json:"pka_activity"`    // Protein kinase A activity level
	PKCActivity    float64   `json:"pkc_activity"`    // Protein kinase C activity level
	Timestamp      time.Time `json:"timestamp"`       // When feedback occurred
	PlasticityType string    `json:"plasticity_type"` // Type of plasticity event
}

// ChannelTrigger represents the biophysical conditions that cause ion channel
// state transitions. This models the actual molecular mechanisms that control
// channel gating in biological membranes.
//
// BIOLOGICAL GATING MECHANISMS:
// - Voltage sensing: S4 domain movement in response to electric field
// - Ligand binding: Conformational changes upon neurotransmitter binding
// - Calcium sensing: EF-hand or other Ca2+-binding domain activation
// - Mechanical stress: Membrane tension-induced conformational changes
type ChannelTrigger struct {
	// === VOLTAGE DEPENDENCE ===
	ActivationVoltage   float64 `json:"activation_voltage"`   // Voltage for 50% activation (mV)
	VoltageSlope        float64 `json:"voltage_slope"`        // Steepness of voltage dependence
	InactivationVoltage float64 `json:"inactivation_voltage"` // Voltage for inactivation

	// === LIGAND DEPENDENCE ===
	LigandType          types.LigandType `json:"ligand_type"`          // Required neurotransmitter
	LigandThreshold     float64          `json:"ligand_threshold"`     // Minimum concentration
	CooperativityN      int              `json:"cooperativity_n"`      // Hill coefficient
	DesensitizationRate float64          `json:"desensitization_rate"` // Rate of receptor desensitization

	// === CALCIUM DEPENDENCE ===
	CalciumThreshold float64 `json:"calcium_threshold"`  // [Ca2+] required for activation
	CalciumHillCoeff float64 `json:"calcium_hill_coeff"` // Cooperativity of Ca2+ binding

	// === KINETIC PROPERTIES ===
	ActivationTimeConstant   time.Duration `json:"activation_time_constant"`   // τ_activation
	DeactivationTimeConstant time.Duration `json:"deactivation_time_constant"` // τ_deactivation
	InactivationTimeConstant time.Duration `json:"inactivation_time_constant"` // τ_inactivation
	RecoveryTimeConstant     time.Duration `json:"recovery_time_constant"`     // τ_recovery
}

// ============================================================================
// CORE ION CHANNEL INTERFACE
// ============================================================================

// IonChannel defines the interface for biological ion channels that modulate
// dendritic signal processing through controlled ionic conductances.
// This interface captures the essential biophysical properties and behaviors
// of membrane ion channels.
//
// DESIGN PRINCIPLES:
// 1. BIOPHYSICAL REALISM: Methods correspond to actual channel biophysics
// 2. STATE-DEPENDENT GATING: Channels maintain kinetic state between calls
// 3. ACTIVITY-DEPENDENT MODULATION: Channels adapt based on usage patterns
// 4. IONIC SPECIFICITY: Different channels conduct different ion types
// 5. VOLTAGE AND LIGAND SENSITIVITY: Channels respond to multiple stimuli
//
// USAGE PATTERN:
// 1. Channel.ShouldOpen() checks if biophysical conditions favor opening
// 2. Channel.ModulateCurrent() transforms ionic current based on channel state
// 3. Channel.UpdateKinetics() evolves channel state and learns from feedback
// 4. Channel.GetState() provides introspection for monitoring and analysis
type IonChannel interface {
	// === CORE CHANNEL FUNCTION ===

	// ModulateCurrent processes an incoming synaptic current based on the
	// channel's current state and local membrane conditions. This is the
	// primary biophysical function that implements ionic current modulation.
	//
	// BIOLOGICAL PROCESS MODELED:
	// When synaptic current flows through a dendritic membrane segment containing
	// this ion channel, the channel's current state (open/closed) and conductance
	// properties modulate the effective current that reaches the soma.
	//
	// Parameters:
	//   msg: The incoming synaptic message containing current information
	//   voltage: Local membrane potential for voltage-dependent channels
	//   calcium: Local [Ca2+] for calcium-activated channels
	//
	// Returns:
	//   *types.NeuralSignal: The modified signal (or nil if blocked)
	//   bool: Whether the signal should continue processing
	//   float64: Additional ionic current generated by this channel (pA)
	ModulateCurrent(msg types.NeuralSignal, voltage, calcium float64) (*types.NeuralSignal, bool, float64)

	// === CHANNEL GATING ===

	// ShouldOpen determines if the channel should transition to the open state
	// based on current biophysical conditions. This models the stochastic
	// nature of channel gating and voltage/ligand dependence.
	//
	// BIOLOGICAL BASIS:
	// Ion channel opening is probabilistic and depends on membrane voltage,
	// ligand concentrations, and internal channel state. This method implements
	// the transition kinetics between closed, open, and inactivated states.
	//
	// Parameters:
	//   voltage: Current membrane potential (mV)
	//   ligandConc: Concentration of relevant ligand
	//   calcium: Intracellular calcium concentration
	//   deltaTime: Time step for kinetic calculations
	//
	// Returns:
	//   bool: Whether channel should open
	//   time.Duration: Expected open duration
	//   float64: Open probability (0.0-1.0)
	ShouldOpen(voltage, ligandConc, calcium float64, deltaTime time.Duration) (bool, time.Duration, float64)

	// UpdateKinetics evolves the channel's state based on time passage and
	// activity-dependent feedback. This implements the biological mechanisms
	// of channel modulation and plasticity.
	//
	// BIOLOGICAL PROCESSES:
	// - Voltage-dependent activation/inactivation kinetics
	// - Use-dependent facilitation or depression
	// - Phosphorylation-dependent modulation
	// - Activity-dependent trafficking and expression changes
	//
	// Parameters:
	//   feedback: Information about channel effectiveness for modulation
	//   deltaTime: Time elapsed since last update
	//   voltage: Current membrane voltage for voltage-dependent processes
	UpdateKinetics(feedback *ChannelFeedback, deltaTime time.Duration, voltage float64)

	// === BIOPHYSICAL PROPERTIES ===

	// GetConductance returns the current single-channel conductance.
	// This represents the ease with which ions flow through the open channel.
	GetConductance() float64 // picoSiemens (pS)

	// GetReversalPotential returns the equilibrium potential for this channel.
	// This determines the driving force and direction of ionic current.
	GetReversalPotential() float64 // millivolts (mV)

	// GetIonSelectivity returns the primary ion type conducted by this channel.
	GetIonSelectivity() IonType

	// === INTROSPECTION AND MONITORING ===

	// GetState returns the current state of the ion channel for monitoring
	// and analysis purposes.
	GetState() ChannelState

	// GetTrigger returns the current gating properties of the channel.
	GetTrigger() ChannelTrigger

	// === IDENTIFICATION AND LIFECYCLE ===

	// Name returns a human-readable identifier for this channel instance.
	Name() string

	// ChannelType returns the functional/molecular category of this channel
	// (e.g., "nav1.2", "kv4.2", "nmda", "ampa", "gabaa").
	ChannelType() string

	// Close releases any resources when the channel is no longer needed.
	Close()
}

// IonType represents the primary ion species conducted by a channel
type IonType int

const (
	IonSodium    IonType = iota // Na+ - fast depolarization
	IonPotassium                // K+ - repolarization, afterhyperpolarization
	IonCalcium                  // Ca2+ - signaling, plasticity
	IonChloride                 // Cl- - inhibition, shunting
	IonMixed                    // Mixed cation channels (Na+/K+/Ca2+)
)

func (it IonType) String() string {
	switch it {
	case IonSodium:
		return "Na+"
	case IonPotassium:
		return "K+"
	case IonCalcium:
		return "Ca2+"
	case IonChloride:
		return "Cl-"
	case IonMixed:
		return "Mixed"
	default:
		return "Unknown"
	}
}

// GetReversalPotential returns the typical reversal potential for this ion type
func (it IonType) GetReversalPotential() float64 {
	switch it {
	case IonSodium:
		return DENDRITE_VOLTAGE_REVERSAL_SODIUM
	case IonPotassium:
		return DENDRITE_VOLTAGE_REVERSAL_POTASSIUM
	case IonCalcium:
		return DENDRITE_VOLTAGE_REVERSAL_CALCIUM
	case IonChloride:
		return DENDRITE_VOLTAGE_REVERSAL_CHLORIDE
	case IonMixed:
		return DENDRITE_VOLTAGE_REVERSAL_MIXED
	default:
		return DENDRITE_VOLTAGE_RESTING_CORTICAL // Default to resting potential
	}
}

// InputProcessorFunc processes a single synaptic input value before integration.
// This is a generic, reusable function signature for value transformation.
type InputProcessorFunc func(value float64) float64

// ============================================================================
// ION CHANNEL FACTORY AND CONFIGURATION
// ============================================================================

// ChannelConfig provides the base configuration for creating different
// types of ion channels. Specific channel types extend this with their
// own biophysical parameters.
type ChannelConfig struct {
	// === BASIC PROPERTIES ===
	Name              string  `json:"name"`               // Human-readable identifier
	ChannelType       string  `json:"channel_type"`       // Molecular/functional type
	InitiallyOpen     bool    `json:"initially_open"`     // Whether channel starts open
	MaxConductance    float64 `json:"max_conductance"`    // Maximum conductance (pS)
	ReversalPotential float64 `json:"reversal_potential"` // Equilibrium potential (mV)

	// === ION SELECTIVITY ===
	PrimaryIon           IonType             `json:"primary_ion"`           // Primary ion conducted
	RelativePermeability map[IonType]float64 `json:"relative_permeability"` // Ion permeability ratios

	// === KINETIC PARAMETERS ===
	ActivationTimeConstant   time.Duration `json:"activation_time_constant"`
	DeactivationTimeConstant time.Duration `json:"deactivation_time_constant"`
	InactivationTimeConstant time.Duration `json:"inactivation_time_constant"`

	// === VOLTAGE DEPENDENCE ===
	HalfActivationVoltage float64 `json:"half_activation_voltage"` // V1/2 for activation
	ActivationSlope       float64 `json:"activation_slope"`        // Voltage sensitivity

	// === LIGAND DEPENDENCE ===
	RequiredLigand    types.LigandType `json:"required_ligand"`
	LigandSensitivity float64          `json:"ligand_sensitivity"`

	// === MODULATION PARAMETERS ===
	PlasticityEnabled bool    `json:"plasticity_enabled"` // Whether channel can be modulated
	ModulationRate    float64 `json:"modulation_rate"`    // Rate of activity-dependent changes

	// === TRIGGER CONFIGURATION ===
	Trigger ChannelTrigger `json:"trigger"` // Gating conditions
}

// ============================================================================
// Data Transfer Objects (DTOs) for Decoupled Communication
// ============================================================================

// MembraneSnapshot provides a read-only snapshot of the neuron's electrical state
// to the integration strategy. This allows dendritic computations to be context-aware
// without having direct access to the neuron's internal state.
//
// BIOLOGICAL CONTEXT:
// Dendritic integration is heavily influenced by the current state of the membrane.
// Voltage-gated channels, back-propagating action potentials, and local ionic
// concentrations all affect how incoming synaptic signals are processed.
// This snapshot provides the necessary biophysical context.
type MembraneSnapshot struct {
	// === ELECTRICAL STATE ===
	Accumulator      float64 `json:"accumulator"`       // Current membrane potential at soma (mV)
	CurrentThreshold float64 `json:"current_threshold"` // Dynamic firing threshold (mV)
	RestingPotential float64 `json:"resting_potential"` // Baseline membrane potential (mV)

	// === IONIC CONCENTRATIONS ===
	IntracellularCalcium   float64 `json:"intracellular_calcium"`   // [Ca2+]i concentration (μM)
	IntracellularSodium    float64 `json:"intracellular_sodium"`    // [Na+]i concentration (mM)
	IntracellularPotassium float64 `json:"intracellular_potassium"` // [K+]i concentration (mM)

	// === RECENT ACTIVITY ===
	LastSpikeTime        time.Time `json:"last_spike_time"`        // When neuron last fired
	RecentSpikeCount     int       `json:"recent_spike_count"`     // Spikes in last 100ms
	BackPropagatingSpike bool      `json:"back_propagating_spike"` // Whether bAP is present

	// === METABOLIC STATE ===
	ATPLevel        float64 `json:"atp_level"`        // Energy availability
	MetabolicStress float64 `json:"metabolic_stress"` // Cellular stress level
}

// IntegratedPotential is the result returned by dendritic integration modes.
// It represents the net effect of all synaptic inputs after complex dendritic
// computations have been performed.
//
// BIOLOGICAL CONTEXT:
// This represents the effective current that reaches the axon hillock after
// traveling through the dendritic tree and being processed by various ion
// channels and dendritic nonlinearities.
type IntegratedPotential struct {
	// === PRIMARY OUTPUT ===
	NetCurrent float64 `json:"net_current"` // Total current to be applied (pA)

	// === IONIC COMPONENTS ===
	SodiumCurrent    float64 `json:"sodium_current"`    // Na+ component (pA)
	PotassiumCurrent float64 `json:"potassium_current"` // K+ component (pA)
	CalciumCurrent   float64 `json:"calcium_current"`   // Ca2+ component (pA)
	ChlorideCurrent  float64 `json:"chloride_current"`  // Cl- component (pA)

	// === COMPUTATIONAL METADATA ===
	DendriticSpike         bool               `json:"dendritic_spike"`         // Whether dendritic spike occurred
	NonlinearAmplification float64            `json:"nonlinear_amplification"` // Amplification factor applied
	ChannelContributions   map[string]float64 `json:"channel_contributions"`   // Per-channel contributions
}

// ============================================================================
// Core Integration Interface
// ============================================================================

// DendriticIntegrationMode defines the strategy pattern interface for different
// methods of synaptic signal integration. This allows modeling the diverse
// computational capabilities of biological dendrites.
//
// PURPOSE (SOFTWARE ARCHITECTURE):
// Decouples the neuron's core functionality from specific dendritic algorithms,
// enabling modular and extensible dendritic computation models.
//
// PURPOSE (BIOLOGICAL MODELING):
// Represents the vast diversity of dendritic computation found across neuron types,
// from simple passive integration to complex active dendritic processing.
type DendriticIntegrationMode interface {
	// Handle processes incoming synaptic messages immediately or buffers them
	// for later integration. Models the immediate local effects of synaptic
	// activation on the dendritic membrane.
	Handle(msg types.NeuralSignal) *IntegratedPotential

	// Process performs time-based integration of buffered inputs, taking into
	// account membrane state and ion channel dynamics. Models the slower
	// integrative processes that occur over the membrane time constant.
	Process(state MembraneSnapshot) *IntegratedPotential

	// Name returns a human-readable identifier for this integration strategy.
	Name() string

	// SetCoincidenceDetector configures the coincidence detection mechanism
	// for this integration mode. Not all modes support coincidence detection.
	SetCoincidenceDetector(detector CoincidenceDetector)

	// Close releases any resources held by the integration mode.
	Close()
}

// ============================================================================
// Integration Strategy Implementations
// ============================================================================

// ----------------------------------------------------------------------------
// 1. PassiveMembraneMode (Immediate Processing)
// ----------------------------------------------------------------------------

// PassiveMembraneMode implements immediate processing without buffering.
// Models simple, passive dendrites that act as linear signal collectors.
//
// BIOLOGICAL CONTEXT:
// Represents neurons with minimal dendritic computation, where signals
// are passed directly to the soma without significant temporal integration
// or nonlinear processing.
type PassiveMembraneMode struct{}

// NewPassiveMembraneMode creates a new passive integration strategy.
func NewPassiveMembraneMode() *PassiveMembraneMode {
	return &PassiveMembraneMode{}
}

// Handle immediately converts the message to integrated potential.
func (m *PassiveMembraneMode) Handle(msg types.NeuralSignal) *IntegratedPotential {
	return &IntegratedPotential{
		NetCurrent: msg.Value,
		ChannelContributions: map[string]float64{
			"passive": msg.Value,
		},
	}
}

// Process does nothing in passive mode as all processing is immediate.
func (m *PassiveMembraneMode) Process(state MembraneSnapshot) *IntegratedPotential {
	return nil
}

// Name returns the identifier for this strategy.
func (m *PassiveMembraneMode) Name() string { return "PassiveMembrane" }

// SetCoincidenceDetector does nothing for passive mode (no coincidence detection)
func (m *PassiveMembraneMode) SetCoincidenceDetector(detector CoincidenceDetector) {
	// Passive mode doesn't use coincidence detection
	if detector != nil {
		detector.Close() // Clean up the detector since we won't use it
	}
}

// Close does nothing as there are no resources to release.
func (m *PassiveMembraneMode) Close() {}

// ----------------------------------------------------------------------------
// 2. TemporalSummationMode (Time-based Integration with Ion Channels)
// ----------------------------------------------------------------------------

// TemporalSummationMode implements buffered, time-based integration with
// ion channel modulation. Collects synaptic inputs within discrete time
// windows and processes them as batches.
//
// BIOLOGICAL CONTEXT:
// Models the membrane time constant and realistic temporal integration,
// solving race conditions between excitatory and inhibitory inputs.
// Includes ion channel-based signal modulation.
type TemporalSummationMode struct {
	buffer       []types.NeuralSignal
	mutex        sync.Mutex
	channelChain []IonChannel // Ion channel modulation chain
}

// NewTemporalSummationMode creates a new time-based integration strategy.
func NewTemporalSummationMode() *TemporalSummationMode {
	return &TemporalSummationMode{
		buffer:       make([]types.NeuralSignal, 0, DEFAULT_INPUT_BUFFER_SIZE),
		channelChain: make([]IonChannel, 0),
	}
}

// Handle processes the message through ion channels and buffers for integration.
func (m *TemporalSummationMode) Handle(msg types.NeuralSignal) *IntegratedPotential {
	// === ION CHANNEL PROCESSING CHAIN ===
	currentMsg := &msg
	totalChannelCurrent := 0.0
	channelContributions := make(map[string]float64)

	// Default membrane conditions (could be enhanced with actual state)
	voltage := DENDRITE_VOLTAGE_RESTING_CORTICAL
	calcium := DENDRITE_CALCIUM_BASELINE_INTRACELLULAR

	for _, channel := range m.channelChain {
		modifiedMsg, shouldContinue, channelCurrent := channel.ModulateCurrent(*currentMsg, voltage, calcium)
		if !shouldContinue {
			return nil // Signal blocked by this channel
		}
		currentMsg = modifiedMsg
		totalChannelCurrent += channelCurrent
		channelContributions[channel.Name()] = channelCurrent
	}

	// Buffer the processed message
	m.mutex.Lock()
	m.buffer = append(m.buffer, *currentMsg)
	m.mutex.Unlock()

	// Return immediate channel effects if significant
	if totalChannelCurrent != 0 {
		return &IntegratedPotential{
			NetCurrent:           totalChannelCurrent,
			ChannelContributions: channelContributions,
		}
	}

	return nil
}

// SetChannels configures the ion channel processing chain.
func (m *TemporalSummationMode) SetChannels(channels []IonChannel) {
	m.channelChain = channels
}

// AddChannel adds a single ion channel to the processing chain.
func (m *TemporalSummationMode) AddChannel(channel IonChannel) {
	m.channelChain = append(m.channelChain, channel)
}

// Process calculates the temporal sum of all buffered messages.
func (m *TemporalSummationMode) Process(state MembraneSnapshot) *IntegratedPotential {
	m.mutex.Lock()
	if len(m.buffer) == 0 {
		m.mutex.Unlock()
		return nil
	}

	// Copy and clear buffer
	currentBatch := make([]types.NeuralSignal, len(m.buffer))
	copy(currentBatch, m.buffer)
	m.buffer = m.buffer[:0]
	m.mutex.Unlock()

	// Simple linear summation
	var totalCurrent float64
	for _, msg := range currentBatch {
		totalCurrent += msg.Value
	}

	return &IntegratedPotential{
		NetCurrent: totalCurrent,
		ChannelContributions: map[string]float64{
			"temporal_sum": totalCurrent,
		},
	}
}

// Name returns the identifier for this strategy.
func (m *TemporalSummationMode) Name() string { return "TemporalSummation" }

// SetCoincidenceDetector does nothing for temporal summation mode (no coincidence detection)
func (m *TemporalSummationMode) SetCoincidenceDetector(detector CoincidenceDetector) {
	// Basic temporal summation doesn't use coincidence detection
	if detector != nil {
		detector.Close() // Clean up the detector since we won't use it
	}
}

// Close releases resources and closes all channels.
func (m *TemporalSummationMode) Close() {
	for _, channel := range m.channelChain {
		if channel != nil {
			channel.Close()
		}
	}
	m.channelChain = nil
}

// ----------------------------------------------------------------------------
// 3. BiologicalTemporalSummationMode (Realistic Membrane Dynamics)
// ----------------------------------------------------------------------------

// BiologicalTemporalSummationMode implements realistic dendritic integration
// with exponential decay, membrane time constants, and ion channel dynamics.
//
// BIOLOGICAL CONTEXT:
// Models the actual biophysics of dendritic integration including membrane
// time constants, exponential decay of PSPs, spatial heterogeneity, and
// realistic noise characteristics.
type BiologicalTemporalSummationMode struct {
	// === MEMBRANE BIOPHYSICS ===
	membraneTimeConstant time.Duration // τ = Rm × Cm (10-50ms typical)
	restingPotential     float64       // Baseline membrane potential (mV)

	// === BUFFERED INPUTS WITH PRECISE TIMING ===
	buffer          []TimestampedInput
	bufferMutex     sync.Mutex
	lastProcessTime time.Time

	// === DENDRITIC HETEROGENEITY ===
	branchTimeConstants map[string]time.Duration // Different τ per dendritic branch
	spatialDecayFactor  float64                  // Distance-dependent attenuation

	// === BIOLOGICAL NOISE AND VARIATION ===
	membraneNoise  float64       // Thermal and channel noise
	temporalJitter time.Duration // Realistic timing variability
	noiseSeed      int64         // Deterministic noise for reproducibility

	// === ION CHANNEL CHAIN ===
	channelChain []IonChannel // Ion channel processing pipeline
}

// TimestampedInput represents a synaptic input with precise temporal information.
type TimestampedInput struct {
	Message         types.NeuralSignal
	ArrivalTime     time.Time
	DecayFactor     float64            // Pre-computed spatial decay factor
	ChannelCurrents map[string]float64 // Currents from ion channels
}

// BiologicalConfig holds parameters for realistic dendritic integration.
type BiologicalConfig struct {
	MembraneTimeConstant time.Duration            // τ = Rm × Cm
	RestingPotential     float64                  // Baseline Vm (mV)
	BranchTimeConstants  map[string]time.Duration // Heterogeneous dendrites
	SpatialDecayFactor   float64                  // Distance attenuation
	MembraneNoise        float64                  // Biological noise level
	TemporalJitter       time.Duration            // Timing variability
}

// NewBiologicalTemporalSummationMode creates realistic dendritic integration.
func NewBiologicalTemporalSummationMode(config BiologicalConfig) *BiologicalTemporalSummationMode {
	return &BiologicalTemporalSummationMode{
		membraneTimeConstant: config.MembraneTimeConstant,
		restingPotential:     config.RestingPotential,
		buffer:               make([]TimestampedInput, 0, DEFAULT_INPUT_BUFFER_SIZE),
		lastProcessTime:      time.Now(),
		branchTimeConstants:  config.BranchTimeConstants,
		spatialDecayFactor:   config.SpatialDecayFactor,
		membraneNoise:        config.MembraneNoise,
		temporalJitter:       config.TemporalJitter,
		noiseSeed:            time.Now().UnixNano(),
		channelChain:         make([]IonChannel, 0),
	}
}

// Handle buffers input with timestamp for realistic temporal processing.
func (m *BiologicalTemporalSummationMode) Handle(msg types.NeuralSignal) *IntegratedPotential {
	now := time.Now()

	// === ION CHANNEL PROCESSING ===
	currentMsg := &msg
	channelCurrents := make(map[string]float64)
	totalChannelCurrent := 0.0

	// Use membrane state for channel calculations
	voltage := m.restingPotential
	calcium := DENDRITE_CALCIUM_BASELINE_INTRACELLULAR

	for _, channel := range m.channelChain {
		modifiedMsg, shouldContinue, channelCurrent := channel.ModulateCurrent(*currentMsg, voltage, calcium)
		if !shouldContinue {
			return nil // Signal blocked by this channel
		}
		currentMsg = modifiedMsg
		channelCurrents[channel.Name()] = channelCurrent
		totalChannelCurrent += channelCurrent
	}

	// Apply spatial decay based on source location
	spatialWeight := m.calculateSpatialWeight(currentMsg.SourceID)

	// Create timestamped input with biological modifications
	input := TimestampedInput{
		Message:         *currentMsg,
		ArrivalTime:     now,
		DecayFactor:     spatialWeight,
		ChannelCurrents: channelCurrents,
	}

	// Add temporal jitter for biological realism
	if m.temporalJitter > 0 {
		jitter := time.Duration(rand.NormFloat64() * float64(m.temporalJitter))
		input.ArrivalTime = input.ArrivalTime.Add(jitter)
	}

	m.bufferMutex.Lock()
	m.buffer = append(m.buffer, input)
	m.bufferMutex.Unlock()

	// Return immediate channel effects if significant
	if totalChannelCurrent != 0 {
		return &IntegratedPotential{
			NetCurrent:           totalChannelCurrent,
			ChannelContributions: channelCurrents,
		}
	}

	return nil // Will be processed during Process() call
}

// SetChannels configures the ion channel processing chain.
func (m *BiologicalTemporalSummationMode) SetChannels(channels []IonChannel) {
	m.channelChain = channels
}

// AddChannel adds a single ion channel to the processing chain.
func (m *BiologicalTemporalSummationMode) AddChannel(channel IonChannel) {
	m.channelChain = append(m.channelChain, channel)
}

// Process performs biologically realistic temporal integration with exponential decay.
func (m *BiologicalTemporalSummationMode) Process(state MembraneSnapshot) *IntegratedPotential {
	// Use helper method for consistent processing
	totalExcitation, totalInhibition, channelContributions := m.processDecayedComponents(time.Now(), nil)

	// Combine excitatory and inhibitory components
	netCurrent := totalExcitation - totalInhibition

	// Apply biological constraints using constants from dendrite_constants.go
	if netCurrent > DENDRITE_CURRENT_MAX_BIOLOGICAL {
		netCurrent = DENDRITE_CURRENT_MAX_BIOLOGICAL
	} else if netCurrent < DENDRITE_CURRENT_MIN_BIOLOGICAL {
		netCurrent = DENDRITE_CURRENT_MIN_BIOLOGICAL
	}

	// Only return significant changes
	if math.Abs(netCurrent) < DENDRITE_CURRENT_NOISE_FLOOR {
		return nil
	}

	return &IntegratedPotential{
		NetCurrent:           netCurrent,
		ChannelContributions: channelContributions,
	}
}

// processDecayedComponents is a reusable helper for exponential decay calculations.
func (m *BiologicalTemporalSummationMode) processDecayedComponents(now time.Time, processor InputProcessorFunc) (totalExcitation, totalInhibition float64, channelContributions map[string]float64) {
	m.bufferMutex.Lock()
	defer m.bufferMutex.Unlock()

	channelContributions = make(map[string]float64)

	if len(m.buffer) == 0 {
		m.lastProcessTime = now
		return 0.0, 0.0, channelContributions
	}

	// Find most recent input for temporal reference
	var mostRecentTime time.Time
	for _, input := range m.buffer {
		if input.ArrivalTime.After(mostRecentTime) {
			mostRecentTime = input.ArrivalTime
		}
	}

	processTime := mostRecentTime
	if processTime.IsZero() {
		processTime = now
	}

	var decayedExcitation, decayedInhibition float64

	// Process each input with exponential temporal decay
	for _, input := range m.buffer {
		// Calculate age of input for decay computation
		var ageOfInput time.Duration
		if !mostRecentTime.IsZero() && now.Sub(mostRecentTime) < 5*time.Millisecond {
			ageOfInput = processTime.Sub(input.ArrivalTime)
		} else {
			ageOfInput = now.Sub(input.ArrivalTime)
		}

		if ageOfInput >= 0 {
			// Start with original synaptic current
			value := input.Message.Value

			// Apply optional processing function (e.g., saturation)
			if processor != nil {
				value = processor(value)
			}

			// Calculate exponential temporal decay
			timeConstant := m.getEffectiveTimeConstant(input.Message.SourceID)
			tauInSeconds := timeConstant.Seconds()

			var temporalDecay float64
			if tauInSeconds > 0 {
				temporalDecay = math.Exp(-ageOfInput.Seconds() / tauInSeconds)
			} else {
				temporalDecay = 0.0
			}

			// Apply spatial and temporal decay to synaptic current
			effectiveInput := value * input.DecayFactor * temporalDecay

			// Add channel currents with same decay
			for channelName, channelCurrent := range input.ChannelCurrents {
				decayedChannelCurrent := channelCurrent * temporalDecay
				channelContributions[channelName] += decayedChannelCurrent
				effectiveInput += decayedChannelCurrent
			}

			// Add biological membrane noise
			if m.membraneNoise > 0 {
				noise := m.generateBiologicalNoise(input.ArrivalTime)
				effectiveInput += noise
			}

			// Separate excitatory and inhibitory components
			if effectiveInput >= 0 {
				decayedExcitation += effectiveInput
			} else {
				decayedInhibition += -effectiveInput // Store as positive value
			}
		}
	}

	// Clear processed inputs
	m.buffer = m.buffer[:0]
	m.lastProcessTime = now

	return decayedExcitation, decayedInhibition, channelContributions
}

// calculateSpatialWeight models distance-dependent signal attenuation.
func (m *BiologicalTemporalSummationMode) calculateSpatialWeight(sourceID string) float64 {
	baseWeight := 1.0

	// Apply distance-based decay if spatial information available
	if m.spatialDecayFactor > 0 {
		// Simplified spatial model - could be enhanced with actual 3D positions
		switch sourceID {
		case "distal":
			baseWeight = DENDRITE_FACTOR_WEIGHT_DISTAL
		case "proximal":
			baseWeight = DENDRITE_FACTOR_WEIGHT_PROXIMAL
		case "apical":
			baseWeight = DENDRITE_FACTOR_WEIGHT_APICAL
		case "basal":
			baseWeight = DENDRITE_FACTOR_WEIGHT_BASAL
		default:
			baseWeight = DENDRITE_FACTOR_WEIGHT_APICAL // Default moderate attenuation
		}
	}

	return baseWeight
}

// getEffectiveTimeConstant returns branch-specific membrane time constant.
func (m *BiologicalTemporalSummationMode) getEffectiveTimeConstant(sourceID string) time.Duration {
	// Use branch-specific time constant if available
	if branchTau, exists := m.branchTimeConstants[sourceID]; exists {
		return branchTau
	}

	// Use default membrane time constant
	return m.membraneTimeConstant
}

// generateBiologicalNoise creates realistic membrane noise.
func (m *BiologicalTemporalSummationMode) generateBiologicalNoise(arrivalTime time.Time) float64 {
	nanoTime := float64(arrivalTime.UnixNano())
	seedFactor := float64(m.noiseSeed % SEED_MODULUS)

	// Use multiple frequency components for realistic noise distribution
	noise1 := math.Sin(nanoTime*NANOSECONDS_TO_SECONDS*NOISE_FREQUENCY_1 + seedFactor*SEED_SCALE_PRIMARY)
	noise2 := math.Cos(nanoTime*NANOSECONDS_TO_SECONDS*NOISE_FREQUENCY_2 + seedFactor*SEED_SCALE_SECONDARY)
	noise3 := math.Sin(seedFactor * SEED_SCALE_TERTIARY)

	// Approximate normal distribution from sinusoids
	normalApprox := (noise1 + noise2 + noise3) / NOISE_DIVISOR
	noise := normalApprox * m.membraneNoise * NOISE_SCALE

	// Update noise seed for next calculation
	m.noiseSeed = (m.noiseSeed*LCG_MULTIPLIER + LCG_INCREMENT) % LCG_MODULUS

	return noise
}

// Name returns the identifier for this integration strategy.
func (m *BiologicalTemporalSummationMode) Name() string {
	return "BiologicalTemporalSummation"
}

// SetCoincidenceDetector does nothing for biological temporal summation mode (no coincidence detection)
func (m *BiologicalTemporalSummationMode) SetCoincidenceDetector(detector CoincidenceDetector) {
	// Biological temporal summation doesn't use coincidence detection
	if detector != nil {
		detector.Close() // Clean up the detector since we won't use it
	}
}

// Close releases resources and closes all ion channels.
func (m *BiologicalTemporalSummationMode) Close() {
	// Close all channels in the processing chain
	for _, channel := range m.channelChain {
		if channel != nil {
			channel.Close()
		}
	}

	// Clear buffer and channel chain
	m.bufferMutex.Lock()
	m.buffer = m.buffer[:0]
	m.bufferMutex.Unlock()

	m.channelChain = nil
}

// ----------------------------------------------------------------------------
// 4. ShuntingInhibitionMode (Divisive Inhibitory Effects)
// ----------------------------------------------------------------------------

// ShuntingInhibitionMode models the divisive effects of GABAergic inhibition.
// Builds upon BiologicalTemporalSummationMode but implements non-linear,
// multiplicative inhibition through chloride channel activation.
//
// BIOLOGICAL CONTEXT:
// GABA-A receptors create chloride-permeable channels that increase membrane
// conductance, effectively "shunting" excitatory currents. This creates
// divisive rather than subtractive inhibition, providing powerful gain control.
type ShuntingInhibitionMode struct {
	BiologicalTemporalSummationMode
	ShuntingStrength float64 // Strength of divisive inhibition (0.0-1.0)
}

// NewShuntingInhibitionMode creates inhibition mode with divisive effects.
func NewShuntingInhibitionMode(strength float64, config BiologicalConfig) *ShuntingInhibitionMode {
	if strength <= 0 {
		strength = DENDRITE_FACTOR_SHUNTING_DEFAULT // Default to moderate shunting using constant
	}
	return &ShuntingInhibitionMode{
		BiologicalTemporalSummationMode: *NewBiologicalTemporalSummationMode(config),
		ShuntingStrength:                strength,
	}
}

// Process implements divisive inhibition on the integrated signals.
func (m *ShuntingInhibitionMode) Process(state MembraneSnapshot) *IntegratedPotential {
	// Get decayed components using parent method
	totalExcitation, totalInhibition, channelContributions := m.processDecayedComponents(time.Now(), nil)

	// Early exit if no significant input
	if math.Abs(totalExcitation) < DENDRITE_CURRENT_NOISE_FLOOR && math.Abs(totalInhibition) < DENDRITE_CURRENT_NOISE_FLOOR {
		return nil
	}

	// Apply shunting inhibition (divisive effect)
	shuntingFactor := 1.0 - (totalInhibition * m.ShuntingStrength)
	if shuntingFactor < DENDRITE_FACTOR_SHUNTING_FLOOR {
		shuntingFactor = DENDRITE_FACTOR_SHUNTING_FLOOR // Prevent complete blocking using constant
	}

	netCurrent := totalExcitation * shuntingFactor

	return &IntegratedPotential{
		NetCurrent:             netCurrent,
		NonlinearAmplification: shuntingFactor,
		ChannelContributions:   channelContributions,
	}
}

// Name returns the identifier for this strategy.
func (m *ShuntingInhibitionMode) Name() string { return "ShuntingInhibition" }

// SetCoincidenceDetector does nothing for shunting inhibition mode (no coincidence detection)
func (m *ShuntingInhibitionMode) SetCoincidenceDetector(detector CoincidenceDetector) {
	// Shunting inhibition doesn't use coincidence detection
	if detector != nil {
		detector.Close() // Clean up the detector since we won't use it
	}
}

// ----------------------------------------------------------------------------
// 5. ActiveDendriteMode (Advanced Compartmental Integration with Type-Safe Coincidence Detection)
// ----------------------------------------------------------------------------

// ActiveDendriteMode provides comprehensive modeling of computationally active
// dendrites with multiple nonlinear mechanisms including synaptic saturation,
// shunting inhibition, and type-safe NMDA-like dendritic spikes.
//
// BIOLOGICAL CONTEXT:
// Models cortical pyramidal neuron dendrites that perform active computation
// through voltage-gated channels, NMDA receptors, and regenerative dendritic
// spikes. Includes realistic saturation and spike generation mechanisms with
// sophisticated coincidence detection.
type ActiveDendriteMode struct {
	BiologicalTemporalSummationMode
	Config              ActiveDendriteConfig // Configuration parameters
	coincidenceDetector CoincidenceDetector  // Type-safe coincidence detector
}

// ActiveDendriteConfig holds parameters for active dendritic computation.
// Uses type-safe coincidence detector configuration instead of string/map approach.
type ActiveDendriteConfig struct {
	MaxSynapticEffect       float64                   // Maximum effect of any single synapse (pA)
	ShuntingStrength        float64                   // Strength of divisive inhibition
	DendriticSpikeThreshold float64                   // Threshold for dendritic spike generation (pA)
	NMDASpikeAmplitude      float64                   // Additional current from dendritic spike (pA)
	VoltageThreshold        float64                   // Membrane voltage threshold for spike (mV)
	CoincidenceDetector     CoincidenceDetectorConfig // Type-safe detector configuration
}

// NewActiveDendriteMode creates advanced active dendritic integration.
func NewActiveDendriteMode(config ActiveDendriteConfig, bioConfig BiologicalConfig) *ActiveDendriteMode {
	// Provide sensible defaults using constants from dendrite_constants.go
	if config.MaxSynapticEffect <= 0 {
		config.MaxSynapticEffect = DENDRITE_CURRENT_SATURATION_DEFAULT
	}
	if config.ShuntingStrength <= 0 {
		config.ShuntingStrength = DENDRITE_FACTOR_SHUNTING_DEFAULT
	}
	if config.DendriticSpikeThreshold <= 0 {
		config.DendriticSpikeThreshold = DENDRITE_CURRENT_SPIKE_THRESHOLD_DEFAULT
	}
	if config.NMDASpikeAmplitude <= 0 {
		config.NMDASpikeAmplitude = DENDRITE_CURRENT_SPIKE_AMPLITUDE_DEFAULT
	}
	if config.VoltageThreshold <= 0 {
		config.VoltageThreshold = DENDRITE_VOLTAGE_SPIKE_THRESHOLD_DEFAULT
	}

	mode := &ActiveDendriteMode{
		BiologicalTemporalSummationMode: *NewBiologicalTemporalSummationMode(bioConfig),
		Config:                          config,
	}

	// Create coincidence detector if configured
	if config.CoincidenceDetector != nil {
		switch detectorConfig := config.CoincidenceDetector.(type) {
		case *NMDADetectorConfig:
			detector, err := CreateNMDACoincidenceDetector("active_dendrite_nmda", detectorConfig)
			if err != nil {
				fmt.Printf("Warning: Failed to create NMDA coincidence detector: %v\n", err)
			} else {
				mode.coincidenceDetector = detector
			}
		case *SimpleTemporalDetectorConfig:
			detector, err := CreateSimpleTemporalCoincidenceDetector("active_dendrite_simple", detectorConfig)
			if err != nil {
				fmt.Printf("Warning: Failed to create simple temporal coincidence detector: %v\n", err)
			} else {
				mode.coincidenceDetector = detector
			}
		default:
			fmt.Printf("Warning: Unknown coincidence detector config type: %T\n", detectorConfig)
		}
	}

	return mode
}

// Handle processes incoming synaptic messages through ion channels and buffers for integration.
// For ActiveDendriteMode, we need to preserve the original input information for coincidence detection.
func (m *ActiveDendriteMode) Handle(msg types.NeuralSignal) *IntegratedPotential {
	now := time.Now()

	// === ION CHANNEL PROCESSING CHAIN ===
	currentMsg := &msg
	channelCurrents := make(map[string]float64)
	totalChannelCurrent := 0.0

	// Use membrane state for channel calculations
	voltage := m.restingPotential
	calcium := DENDRITE_CALCIUM_BASELINE_INTRACELLULAR

	for _, channel := range m.channelChain {
		modifiedMsg, shouldContinue, channelCurrent := channel.ModulateCurrent(*currentMsg, voltage, calcium)
		if !shouldContinue {
			return nil // Signal blocked by this channel
		}
		currentMsg = modifiedMsg
		channelCurrents[channel.Name()] = channelCurrent
		totalChannelCurrent += channelCurrent
	}

	// Apply spatial decay based on source location
	spatialWeight := m.calculateSpatialWeight(currentMsg.SourceID)

	// Create timestamped input with biological modifications
	input := TimestampedInput{
		Message:         *currentMsg,
		ArrivalTime:     now,
		DecayFactor:     spatialWeight,
		ChannelCurrents: channelCurrents,
	}

	// Add temporal jitter for biological realism
	if m.temporalJitter > 0 {
		jitter := time.Duration(rand.NormFloat64() * float64(m.temporalJitter))
		input.ArrivalTime = input.ArrivalTime.Add(jitter)
	}

	// Buffer the input for both temporal summation and coincidence detection
	m.bufferMutex.Lock()
	m.buffer = append(m.buffer, input)
	m.bufferMutex.Unlock()

	// Return immediate channel effects if significant
	if totalChannelCurrent != 0 {
		return &IntegratedPotential{
			NetCurrent:           totalChannelCurrent,
			ChannelContributions: channelCurrents,
		}
	}

	return nil // Will be processed during Process() call
}

// FIXED: ActiveDendriteMode.Process() method
// The issue was that processDecayedComponents() cleared the buffer before
// coincidence detection could access it. This fix saves the buffer first.
func (m *ActiveDendriteMode) Process(state MembraneSnapshot) *IntegratedPotential {
	// Step 1: Define saturation logic using constants
	saturator := func(value float64) float64 {
		if value > m.Config.MaxSynapticEffect {
			return m.Config.MaxSynapticEffect
		}
		if value < -m.Config.MaxSynapticEffect {
			return -m.Config.MaxSynapticEffect
		}
		return value
	}

	// CRITICAL FIX: Save buffer contents BEFORE processing for coincidence detection
	var inputsForCoincidence []TimestampedInput
	if m.coincidenceDetector != nil {
		m.bufferMutex.Lock()
		inputsForCoincidence = make([]TimestampedInput, len(m.buffer))
		copy(inputsForCoincidence, m.buffer)
		m.bufferMutex.Unlock()
	}

	// Step 2: Process with saturation (this will clear the buffer)
	totalExcitation, totalInhibition, channelContributions := m.processDecayedComponents(time.Now(), saturator)

	// Early exit if no significant input using constant
	if math.Abs(totalExcitation) < DENDRITE_CURRENT_NOISE_FLOOR && math.Abs(totalInhibition) < DENDRITE_CURRENT_NOISE_FLOOR {
		return nil
	}

	// Step 3: Apply shunting inhibition
	shuntingFactor := 1.0 - (totalInhibition * m.Config.ShuntingStrength)
	if shuntingFactor < DENDRITE_FACTOR_SHUNTING_FLOOR {
		shuntingFactor = DENDRITE_FACTOR_SHUNTING_FLOOR // Use constant for floor
	}
	netExcitation := totalExcitation * shuntingFactor

	// Step 4: Model NMDA-like dendritic spikes with type-safe coincidence detection
	dendriticSpike := false
	additionalCalcium := 0.0
	coincidenceAmplification := 1.0

	if m.coincidenceDetector != nil {
		// FIXED: Use the saved inputs instead of the now-empty buffer
		coincidenceResult := m.coincidenceDetector.Detect(inputsForCoincidence, state)
		if coincidenceResult.CoincidenceDetected {
			// Apply coincidence detection results
			coincidenceAmplification = coincidenceResult.AmplificationFactor
			netExcitation *= coincidenceAmplification
			netExcitation += coincidenceResult.AdditionalCurrent
			additionalCalcium += coincidenceResult.AssociatedCalciumInflux
			dendriticSpike = true
		}
	} else {
		// Fallback to previous hardcoded NMDA-like spike logic if no detector is set
		if netExcitation >= m.Config.DendriticSpikeThreshold && state.Accumulator > m.Config.VoltageThreshold {
			netExcitation += m.Config.NMDASpikeAmplitude
			dendriticSpike = true
			// Use constant from coincidence_constants.go for calcium boost
			additionalCalcium += COINCIDENCE_AMPLIFICATION_CALCIUM_BOOST_DEFAULT
		}
	}

	return &IntegratedPotential{
		NetCurrent:             netExcitation,
		DendriticSpike:         dendriticSpike,
		NonlinearAmplification: shuntingFactor * coincidenceAmplification, // Include both shunting and coincidence amplification
		ChannelContributions:   channelContributions,
		CalciumCurrent:         additionalCalcium,
	}
}

// SetCoincidenceDetector configures a new coincidence detector for this active dendrite mode.
func (m *ActiveDendriteMode) SetCoincidenceDetector(detector CoincidenceDetector) {
	if m.coincidenceDetector != nil {
		m.coincidenceDetector.Close()
	}
	m.coincidenceDetector = detector
}

// Name returns the identifier for this strategy.
func (m *ActiveDendriteMode) Name() string { return "ActiveDendrite" }

// Close releases resources including the coincidence detector.
func (m *ActiveDendriteMode) Close() {
	// Close the parent biological mode
	m.BiologicalTemporalSummationMode.Close()

	// Close the coincidence detector
	if m.coincidenceDetector != nil {
		m.coincidenceDetector.Close()
		m.coincidenceDetector = nil
	}
}

// ============================================================================
// REALISTIC CONFIGURATION FACTORIES
// ============================================================================

// CreateCorticalPyramidalConfig returns realistic cortical neuron parameters using constants.
func CreateCorticalPyramidalConfig() BiologicalConfig {
	return BiologicalConfig{
		MembraneTimeConstant: DENDRITE_TIME_CONSTANT_CORTICAL,
		RestingPotential:     DENDRITE_VOLTAGE_RESTING_CORTICAL,
		SpatialDecayFactor:   DENDRITE_FACTOR_SPATIAL_DECAY_DEFAULT,
		MembraneNoise:        DENDRITE_NOISE_MEMBRANE_CORTICAL,
		TemporalJitter:       DENDRITE_TIME_JITTER_CORTICAL,
		BranchTimeConstants: map[string]time.Duration{
			"apical":   DENDRITE_TIME_CONSTANT_APICAL,
			"basal":    DENDRITE_TIME_CONSTANT_BASAL,
			"distal":   DENDRITE_TIME_CONSTANT_DISTAL,
			"proximal": DENDRITE_TIME_CONSTANT_PROXIMAL,
		},
	}
}

// CreateHippocampalConfig returns hippocampal CA1 pyramidal parameters using constants.
func CreateHippocampalConfig() BiologicalConfig {
	return BiologicalConfig{
		MembraneTimeConstant: DENDRITE_TIME_CONSTANT_HIPPOCAMPAL,
		RestingPotential:     DENDRITE_VOLTAGE_RESTING_HIPPOCAMPAL,
		SpatialDecayFactor:   DENDRITE_FACTOR_SPATIAL_DECAY_WEAK, // Less spatial decay
		MembraneNoise:        DENDRITE_NOISE_MEMBRANE_HIPPOCAMPAL,
		TemporalJitter:       DENDRITE_TIME_JITTER_HIPPOCAMPAL,
		BranchTimeConstants: map[string]time.Duration{
			"apical":   45 * time.Millisecond, // Longer for hippocampus
			"basal":    25 * time.Millisecond,
			"distal":   50 * time.Millisecond,
			"proximal": 15 * time.Millisecond,
		},
	}
}

// CreateInterneuronConfig returns fast-spiking interneuron parameters using constants.
func CreateInterneuronConfig() BiologicalConfig {
	return BiologicalConfig{
		MembraneTimeConstant: DENDRITE_TIME_CONSTANT_INTERNEURON,
		RestingPotential:     DENDRITE_VOLTAGE_RESTING_INTERNEURON,
		SpatialDecayFactor:   DENDRITE_FACTOR_SPATIAL_DECAY_STRONG, // Compact dendritic tree
		MembraneNoise:        DENDRITE_NOISE_MEMBRANE_INTERNEURON,
		TemporalJitter:       DENDRITE_TIME_JITTER_INTERNEURON,
		BranchTimeConstants: map[string]time.Duration{
			"dendrite": DENDRITE_TIME_CONSTANT_INTERNEURON, // Uniform fast dendrites
		},
	}
}

// CreateActiveDendriteConfig returns parameters for active dendritic computation with type-safe coincidence detection.
func CreateActiveDendriteConfig() ActiveDendriteConfig {
	// Create default NMDA coincidence detector config using constants
	defaultDetectorConfig := DefaultNMDADetectorConfig()

	return ActiveDendriteConfig{
		MaxSynapticEffect:       DENDRITE_CURRENT_SATURATION_DEFAULT,
		ShuntingStrength:        DENDRITE_FACTOR_SHUNTING_STRONG,       // Strong divisive inhibition
		DendriticSpikeThreshold: DENDRITE_CURRENT_SPIKE_THRESHOLD_HIGH, // Higher threshold for selectivity
		NMDASpikeAmplitude:      DENDRITE_CURRENT_SPIKE_AMPLITUDE_DEFAULT,
		VoltageThreshold:        DENDRITE_VOLTAGE_SPIKE_THRESHOLD_STRICT,
		CoincidenceDetector:     defaultDetectorConfig, // Type-safe detector config
	}
}
