package neuron

import (
	"math"
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

/*
=================================================================================
REALISTIC ION CHANNEL IMPLEMENTATIONS - BIOPHYSICALLY ACCURATE CHANNEL MODELING
=================================================================================

OVERVIEW:
This file contains realistic implementations of voltage-gated and ligand-gated
ion channels that model specific biological channels found in neurons. Each
channel implementation follows biophysically accurate gating kinetics,
conductance properties, and voltage/ligand dependencies.

BIOLOGICAL CONTEXT:
Ion channels are the molecular basis of neural computation. Different channel
types (Nav1.6, Kv4.2, Cav1.2, GABA-A) have distinct biophysical properties
that determine their roles in synaptic integration, spike generation, and
plasticity. These implementations model experimentally-determined properties.

CHANNEL TYPES IMPLEMENTED:
1. VOLTAGE-GATED SODIUM CHANNELS (Nav1.6): Fast activation/inactivation
2. VOLTAGE-GATED POTASSIUM CHANNELS (Kv4.2): Delayed rectification
3. VOLTAGE-GATED CALCIUM CHANNELS (Cav1.2): High-threshold calcium influx
4. LIGAND-GATED CHLORIDE CHANNELS (GABA-A): Fast inhibitory transmission

=================================================================================
*/

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// validateGatingVariable ensures gating variables stay within [0, 1] range
// and handles NaN/Inf values that can occur with extreme conditions
func validateGatingVariable(value float64, name string) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0.0 // Safe default
	}
	if value < 0.0 {
		return 0.0
	}
	if value > 1.0 {
		return 1.0
	}
	return value
}

// ============================================================================
// REALISTIC VOLTAGE-GATED SODIUM CHANNEL (Nav1.6)
// ============================================================================

// RealisticNavChannel implements voltage-gated sodium channel Nav1.6
// Found in axon initial segments and nodes of Ranvier
type RealisticNavChannel struct {
	name           string
	conductance    float64 // Single channel conductance (pS)
	reversalPot    float64 // Sodium reversal potential (mV)
	isOpen         bool
	isInactivated  bool
	activationTime time.Time
	lastVoltage    float64

	// State variables (0-1)
	activationGate   float64 // m gate
	inactivationGate float64 // h gate
}

// NewRealisticNavChannel creates a biologically accurate Nav1.6 channel
func NewRealisticNavChannel(name string) *RealisticNavChannel {
	return &RealisticNavChannel{
		name:             name,
		conductance:      DENDRITE_CONDUCTANCE_SODIUM_DEFAULT, // 20 pS
		reversalPot:      DENDRITE_VOLTAGE_REVERSAL_SODIUM,    // +60 mV
		isOpen:           false,
		isInactivated:    false,
		activationGate:   0.0, // Closed at rest
		inactivationGate: 1.0, // Not inactivated at rest
		lastVoltage:      DENDRITE_VOLTAGE_RESTING_CORTICAL,
	}
}

func (n *RealisticNavChannel) ModulateCurrent(msg types.NeuralSignal, voltage, calcium float64) (*types.NeuralSignal, bool, float64) {
	// Update gating based on voltage
	n.updateGating(voltage, DENDRITE_TIME_CHANNEL_ACTIVATION)

	// Channel current calculation using proper Hodgkin-Huxley formulation
	var channelCurrent float64
	if n.activationGate > 0.01 && n.inactivationGate > 0.01 { // Small threshold to avoid numerical issues
		drivingForce := voltage - n.reversalPot
		// I = g_max * m^3 * h * (V - E_Na)
		gatingFactor := math.Pow(n.activationGate, 3) * n.inactivationGate
		// Convert from pS to pA: pS * mV = pA
		channelCurrent = n.conductance * gatingFactor * drivingForce
	}

	// Sodium channels don't typically block signals, just add current
	return &msg, true, channelCurrent
}

func (n *RealisticNavChannel) ShouldOpen(voltage, ligandConc, calcium float64, deltaTime time.Duration) (bool, time.Duration, float64) {
	// Update gating kinetics
	n.updateGating(voltage, deltaTime)

	// More realistic opening criteria
	activationProbability := n.activationGate
	inactivationAvailability := n.inactivationGate

	// Combined open probability
	openProbability := activationProbability * inactivationAvailability

	// Should open if combined probability is significant
	shouldOpen := openProbability > 0.1

	// Realistic open duration based on kinetics
	openDuration := DENDRITE_TIME_CHANNEL_ACTIVATION

	return shouldOpen, openDuration, openProbability
}

func (n *RealisticNavChannel) updateGating(voltage float64, deltaTime time.Duration) {
	dt := deltaTime.Seconds()

	// CORRECTED Hodgkin-Huxley voltage-dependent steady-state values
	// Activation (m_inf): sigmoid with V1/2 around -40 mV
	// Formula: m_inf = 1 / (1 + exp(-(V - V_half) / k))
	mInf := 1.0 / (1.0 + math.Exp(-(voltage-(-40.0))/5.0))

	// Inactivation (h_inf): sigmoid with V1/2 around -60 mV, NEGATIVE slope
	// Formula: h_inf = 1 / (1 + exp((V - V_half) / k))
	// Note the POSITIVE sign in the exponent for inactivation
	hInf := 1.0 / (1.0 + math.Exp((voltage-(-60.0))/5.0))

	// Time constants (voltage-dependent)
	tauM := 0.001 // 1 ms activation
	tauH := 0.010 // 10 ms inactivation

	// First-order kinetics: dx/dt = (x_inf - x) / tau
	newM := n.activationGate + dt*(mInf-n.activationGate)/tauM
	newH := n.inactivationGate + dt*(hInf-n.inactivationGate)/tauH

	// CRITICAL: Validate and bound gating variables to [0, 1]
	n.activationGate = validateGatingVariable(newM, "m")
	n.inactivationGate = validateGatingVariable(newH, "h")

	// Update channel state
	n.isOpen = n.activationGate > 0.5 && n.inactivationGate > 0.5
	n.isInactivated = n.inactivationGate < 0.1 // More realistic inactivation threshold
	n.lastVoltage = voltage
}

func (n *RealisticNavChannel) UpdateKinetics(feedback *ChannelFeedback, deltaTime time.Duration, voltage float64) {
	n.updateGating(voltage, deltaTime)

	// Activity-dependent modulation (use-dependent inactivation)
	if feedback != nil && feedback.ContributedToFiring {
		// Slight reduction in availability after contributing to spike
		n.inactivationGate = validateGatingVariable(n.inactivationGate*0.98, "h")
	}
}

func (n *RealisticNavChannel) GetConductance() float64       { return n.conductance }
func (n *RealisticNavChannel) GetReversalPotential() float64 { return n.reversalPot }
func (n *RealisticNavChannel) GetIonSelectivity() IonType    { return IonSodium }
func (n *RealisticNavChannel) Name() string                  { return n.name }
func (n *RealisticNavChannel) ChannelType() string           { return "nav1.6" }
func (n *RealisticNavChannel) Close()                        {}

func (n *RealisticNavChannel) GetState() ChannelState {
	// CORRECTED: Proper Hodgkin-Huxley conductance calculation
	// Conductance = maxConductance * m^3 * h
	effectiveConductance := n.conductance * math.Pow(n.activationGate, 3) * n.inactivationGate

	// Ensure non-negative conductance
	if effectiveConductance < 0 {
		effectiveConductance = 0
	}

	return ChannelState{
		IsOpen:               n.isOpen,
		Conductance:          effectiveConductance,
		EquilibriumPotential: n.reversalPot,
		MembraneVoltage:      n.lastVoltage,
	}
}

func (n *RealisticNavChannel) GetTrigger() ChannelTrigger {
	return ChannelTrigger{
		ActivationVoltage:        -40.0, // V1/2 for activation
		VoltageSlope:             5.0,   // Steepness factor
		InactivationVoltage:      -60.0, // V1/2 for inactivation
		ActivationTimeConstant:   DENDRITE_TIME_CHANNEL_ACTIVATION,
		InactivationTimeConstant: DENDRITE_TIME_CHANNEL_INACTIVATION,
	}
}

// ============================================================================
// REALISTIC VOLTAGE-GATED POTASSIUM CHANNEL (Kv4.2)
// ============================================================================

// RealisticKvChannel implements A-type potassium channel Kv4.2
// Important for dendritic integration and spike frequency adaptation
type RealisticKvChannel struct {
	name           string
	conductance    float64
	reversalPot    float64
	isOpen         bool
	activationGate float64 // n gate
	lastVoltage    float64
}

func NewRealisticKvChannel(name string) *RealisticKvChannel {
	return &RealisticKvChannel{
		name:           name,
		conductance:    DENDRITE_CONDUCTANCE_POTASSIUM_DEFAULT, // 10 pS
		reversalPot:    DENDRITE_VOLTAGE_REVERSAL_POTASSIUM,    // -90 mV
		isOpen:         false,
		activationGate: 0.0, // Closed at rest
		lastVoltage:    DENDRITE_VOLTAGE_RESTING_CORTICAL,
	}
}

func (k *RealisticKvChannel) ModulateCurrent(msg types.NeuralSignal, voltage, calcium float64) (*types.NeuralSignal, bool, float64) {
	k.updateGating(voltage, DENDRITE_TIME_CHANNEL_ACTIVATION)

	var channelCurrent float64
	if k.activationGate > 0.01 { // Small threshold to avoid numerical issues
		drivingForce := voltage - k.reversalPot
		// I = g * n^4 * (V - EK) - fourth power activation for delayed rectifier
		gatingFactor := math.Pow(k.activationGate, 4)
		// Convert from pS to pA: pS * mV = pA
		channelCurrent = k.conductance * gatingFactor * drivingForce
	}

	return &msg, true, channelCurrent
}

func (k *RealisticKvChannel) ShouldOpen(voltage, ligandConc, calcium float64, deltaTime time.Duration) (bool, time.Duration, float64) {
	k.updateGating(voltage, deltaTime)

	shouldOpen := k.activationGate > 0.1               // Lower threshold for K+ channels
	openDuration := DENDRITE_TIME_CHANNEL_DEACTIVATION // Slower than Na+
	openProbability := k.activationGate

	return shouldOpen, openDuration, openProbability
}

func (k *RealisticKvChannel) updateGating(voltage float64, deltaTime time.Duration) {
	dt := deltaTime.Seconds()

	// CORRECTED: K+ channel activation (n_inf): sigmoid with V1/2 around -30 mV
	// Formula: n_inf = 1 / (1 + exp(-(V - V_half) / k))
	nInf := 1.0 / (1.0 + math.Exp(-(voltage-(-30.0))/10.0))

	// Time constant (K+ channels are slower than Na+)
	tauN := 0.005 // 5 ms

	// Update activation gate with bounds checking
	newN := k.activationGate + dt*(nInf-k.activationGate)/tauN
	k.activationGate = validateGatingVariable(newN, "n")

	k.isOpen = k.activationGate > 0.1 // Lower threshold for K+ channels
	k.lastVoltage = voltage
}

func (k *RealisticKvChannel) UpdateKinetics(feedback *ChannelFeedback, deltaTime time.Duration, voltage float64) {
	k.updateGating(voltage, deltaTime)
}

func (k *RealisticKvChannel) GetConductance() float64       { return k.conductance }
func (k *RealisticKvChannel) GetReversalPotential() float64 { return k.reversalPot }
func (k *RealisticKvChannel) GetIonSelectivity() IonType    { return IonPotassium }
func (k *RealisticKvChannel) Name() string                  { return k.name }
func (k *RealisticKvChannel) ChannelType() string           { return "kv4.2" }
func (k *RealisticKvChannel) Close()                        {}

func (k *RealisticKvChannel) GetState() ChannelState {
	// CORRECTED: K+ channels use n^4 gating
	effectiveConductance := k.conductance * math.Pow(k.activationGate, 4)

	if effectiveConductance < 0 {
		effectiveConductance = 0
	}

	return ChannelState{
		IsOpen:               k.isOpen,
		Conductance:          effectiveConductance,
		EquilibriumPotential: k.reversalPot,
		MembraneVoltage:      k.lastVoltage,
	}
}

func (k *RealisticKvChannel) GetTrigger() ChannelTrigger {
	return ChannelTrigger{
		ActivationVoltage:      -30.0, // V1/2 for activation
		VoltageSlope:           10.0,  // Steepness factor
		ActivationTimeConstant: DENDRITE_TIME_CHANNEL_DEACTIVATION,
	}
}

// ============================================================================
// REALISTIC VOLTAGE-GATED CALCIUM CHANNEL (Cav1.2)
// ============================================================================

// RealisticCavChannel implements L-type calcium channel Cav1.2
// Critical for calcium signaling and plasticity
type RealisticCavChannel struct {
	name           string
	conductance    float64
	reversalPot    float64
	isOpen         bool
	activationGate float64
	lastVoltage    float64
	calciumInflux  float64 // Track calcium influx for feedback
}

func NewRealisticCavChannel(name string) *RealisticCavChannel {
	return &RealisticCavChannel{
		name:           name,
		conductance:    DENDRITE_CONDUCTANCE_CALCIUM_DEFAULT, // 5 pS
		reversalPot:    DENDRITE_VOLTAGE_REVERSAL_CALCIUM,    // +120 mV
		isOpen:         false,
		activationGate: 0.0,
		lastVoltage:    DENDRITE_VOLTAGE_RESTING_CORTICAL,
		calciumInflux:  0.0,
	}
}

func (c *RealisticCavChannel) ModulateCurrent(msg types.NeuralSignal, voltage, calcium float64) (*types.NeuralSignal, bool, float64) {
	c.updateGating(voltage, DENDRITE_TIME_CHANNEL_ACTIVATION)

	var channelCurrent float64
	if c.activationGate > 0.01 { // Small threshold to avoid numerical issues
		drivingForce := voltage - c.reversalPot
		// I = g * m^2 * (V - ECa) - square law for Ca2+ channels
		gatingFactor := math.Pow(c.activationGate, 2)
		// Convert from pS to pA: pS * mV = pA
		channelCurrent = c.conductance * gatingFactor * drivingForce

		// Track calcium influx (negative current = influx)
		if channelCurrent < 0 {
			c.calciumInflux += -channelCurrent * 0.01 // Convert to calcium units
		}
	}

	return &msg, true, channelCurrent
}

func (c *RealisticCavChannel) ShouldOpen(voltage, ligandConc, calcium float64, deltaTime time.Duration) (bool, time.Duration, float64) {
	c.updateGating(voltage, deltaTime)

	// Calcium-dependent inactivation
	calciumInactivation := 1.0 / (1.0 + calcium/DENDRITE_CALCIUM_BASELINE_INTRACELLULAR)

	shouldOpen := c.activationGate > 0.2 && calciumInactivation > 0.5 // Moderate threshold
	openDuration := DENDRITE_TIME_CHANNEL_ACTIVATION * 3              // Intermediate kinetics
	openProbability := c.activationGate * calciumInactivation

	return shouldOpen, openDuration, openProbability
}

func (c *RealisticCavChannel) updateGating(voltage float64, deltaTime time.Duration) {
	dt := deltaTime.Seconds()

	// CORRECTED: Ca2+ channel activation (m_inf): higher threshold around -20 mV
	// Formula: m_inf = 1 / (1 + exp(-(V - V_half) / k))
	mInf := 1.0 / (1.0 + math.Exp(-(voltage-(-20.0))/8.0))

	// Time constant (intermediate for Ca2+ channels)
	tauM := 0.003 // 3 ms

	// Update with bounds checking
	newM := c.activationGate + dt*(mInf-c.activationGate)/tauM
	c.activationGate = validateGatingVariable(newM, "m")

	c.isOpen = c.activationGate > 0.2 // Moderate threshold
	c.lastVoltage = voltage
}

func (c *RealisticCavChannel) UpdateKinetics(feedback *ChannelFeedback, deltaTime time.Duration, voltage float64) {
	c.updateGating(voltage, deltaTime)

	// Calcium-dependent facilitation
	if feedback != nil && feedback.CalciumInflux > 0 {
		// Slight increase in conductance with calcium influx
		c.conductance *= 1.001
	}
}

func (c *RealisticCavChannel) GetConductance() float64       { return c.conductance }
func (c *RealisticCavChannel) GetReversalPotential() float64 { return c.reversalPot }
func (c *RealisticCavChannel) GetIonSelectivity() IonType    { return IonCalcium }
func (c *RealisticCavChannel) Name() string                  { return c.name }
func (c *RealisticCavChannel) ChannelType() string           { return "cav1.2" }
func (c *RealisticCavChannel) Close()                        {}

func (c *RealisticCavChannel) GetState() ChannelState {
	// CORRECTED: Ca2+ channels use m^2 gating
	effectiveConductance := c.conductance * math.Pow(c.activationGate, 2)

	if effectiveConductance < 0 {
		effectiveConductance = 0
	}

	return ChannelState{
		IsOpen:               c.isOpen,
		Conductance:          effectiveConductance,
		EquilibriumPotential: c.reversalPot,
		MembraneVoltage:      c.lastVoltage,
		CalciumLevel:         c.calciumInflux,
	}
}

func (c *RealisticCavChannel) GetTrigger() ChannelTrigger {
	return ChannelTrigger{
		ActivationVoltage:      -20.0, // V1/2 for activation
		VoltageSlope:           8.0,   // Steepness factor
		CalciumThreshold:       DENDRITE_CALCIUM_BASELINE_INTRACELLULAR * 2,
		ActivationTimeConstant: DENDRITE_TIME_CHANNEL_ACTIVATION,
	}
}

// ============================================================================
// Realistic GABA-A Channel Implementation
// ============================================================================
// ============================================================================
// REALISTIC LIGAND-GATED CHLORIDE CHANNEL (GABA-A) - FIXED IMPLEMENTATION
// ============================================================================

// RealisticGabaAChannel implements GABA-A receptor channel
// Primary mediator of fast inhibition in the brain
type RealisticGabaAChannel struct {
	name             string
	conductance      float64
	reversalPot      float64
	isOpen           bool
	ligandGate       float64 // GABA binding gate (0-1)
	lastGABA         float64
	desensitized     bool
	inactivationGate float64 // Desensitization gate (0-1)
	mutex            sync.Mutex
}

func NewRealisticGabaAChannel(name string) *RealisticGabaAChannel {
	return &RealisticGabaAChannel{
		name:             name,
		conductance:      15.0,                               // GABA-A channels have higher conductance (pS)
		reversalPot:      DENDRITE_VOLTAGE_REVERSAL_CHLORIDE, // -70 mV
		isOpen:           false,
		ligandGate:       0.0, // Start with no GABA binding
		lastGABA:         0.0,
		desensitized:     false,
		inactivationGate: 1.0, // Start fully available (not desensitized)
	}
}

func (g *RealisticGabaAChannel) ModulateCurrent(msg types.NeuralSignal, voltage, calcium float64) (*types.NeuralSignal, bool, float64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Extract GABA concentration from neurotransmitter type
	var gabaConc float64
	if msg.NeurotransmitterType == types.LigandGABA {
		gabaConc = math.Abs(msg.Value) // Use absolute value for concentration
	}

	// Update gating based on GABA concentration
	g.updateGating(gabaConc, DENDRITE_TIME_CHANNEL_ACTIVATION)

	var channelCurrent float64
	if g.isOpen && !g.desensitized && g.ligandGate > 0.01 {
		drivingForce := voltage - g.reversalPot
		// I = g * (receptor occupancy) * (desensitization factor) * (V - ECl)
		// Convert from pS to pA: pS * mV = pA
		gatingFactor := g.ligandGate * g.inactivationGate
		channelCurrent = g.conductance * gatingFactor * drivingForce
	}

	return &msg, true, channelCurrent
}

func (g *RealisticGabaAChannel) ShouldOpen(voltage, ligandConc, calcium float64, deltaTime time.Duration) (bool, time.Duration, float64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Update gating with the provided ligand concentration
	g.updateGating(math.Abs(ligandConc), deltaTime)

	// Channel opens based on ligand binding and desensitization state
	shouldOpen := g.ligandGate > 0.1 && g.inactivationGate > 0.1
	openDuration := DENDRITE_TIME_CHANNEL_DEACTIVATION * 2 // GABA-A kinetics are slower

	// Open probability is the product of activation and availability
	openProbability := g.ligandGate * g.inactivationGate

	g.isOpen = shouldOpen
	return shouldOpen, openDuration, openProbability
}

func (g *RealisticGabaAChannel) updateGating(gabaConc float64, deltaTime time.Duration) {
	dt := deltaTime.Seconds()

	// Use absolute value to ensure positive concentration
	absGABA := math.Abs(gabaConc)

	// --- Activation Gate (ligand binding) ---
	// Hill equation for GABA binding with cooperativity
	kd := 5.0    // Dissociation constant (μM)
	nHill := 2.0 // Cooperativity factor

	var activationInf float64
	if absGABA > 0 {
		activationInf = math.Pow(absGABA, nHill) / (math.Pow(kd, nHill) + math.Pow(absGABA, nHill))
	} else {
		activationInf = 0.0
	}

	// Time constant for binding (fast)
	tauActivation := 0.002 // 2 ms

	// Update activation gate
	newActivation := g.ligandGate + dt*(activationInf-g.ligandGate)/tauActivation
	g.ligandGate = validateGatingVariable(newActivation, "ligand_activation")

	// --- Inactivation Gate (desensitization) ---
	// Desensitization depends on GABA concentration and time
	var inactivationInf float64
	if absGABA > 0 {
		// Higher GABA concentrations cause more desensitization
		desensitizationFactor := absGABA / (absGABA + 10.0)   // Saturation at high concentrations
		inactivationInf = 1.0 - (0.8 * desensitizationFactor) // Max 80% desensitization
	} else {
		inactivationInf = 1.0 // Full recovery without GABA
	}

	// Time constant for desensitization (slow)
	tauInactivation := 0.1 // 100 ms

	// Update inactivation gate
	newInactivation := g.inactivationGate + dt*(inactivationInf-g.inactivationGate)/tauInactivation
	g.inactivationGate = validateGatingVariable(newInactivation, "ligand_inactivation")

	// Update channel state
	g.isOpen = g.ligandGate > 0.1 && g.inactivationGate > 0.1
	g.desensitized = g.inactivationGate < 0.2
	g.lastGABA = gabaConc
}

// UpdateKinetics evolves the channel's internal gates over time
func (g *RealisticGabaAChannel) UpdateKinetics(feedback *ChannelFeedback, deltaTime time.Duration, voltage float64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Use the last known GABA concentration for kinetics evolution
	g.updateGating(g.lastGABA, deltaTime)

	// GABA-A channels can show plasticity
	if feedback != nil && feedback.ContributedToFiring {
		// Slight reduction in sensitivity after contributing to inhibition
		g.conductance *= 0.999
	}
}

// GetResponse returns the current open probability of the channel
func (g *RealisticGabaAChannel) GetResponse() float64 {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	return g.ligandGate * g.inactivationGate
}

func (g *RealisticGabaAChannel) GetConductance() float64 {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	// Calculate effective conductance directly without calling GetResponse()
	return g.conductance * g.ligandGate * g.inactivationGate
}

func (g *RealisticGabaAChannel) GetReversalPotential() float64 { return g.reversalPot }
func (g *RealisticGabaAChannel) GetIonSelectivity() IonType    { return IonChloride }
func (g *RealisticGabaAChannel) Name() string                  { return g.name }
func (g *RealisticGabaAChannel) ChannelType() string           { return "gabaa_realistic" }
func (g *RealisticGabaAChannel) Close()                        {}

func (g *RealisticGabaAChannel) GetState() ChannelState {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	effectiveConductance := g.conductance * g.ligandGate * g.inactivationGate
	if effectiveConductance < 0 {
		effectiveConductance = 0
	}

	return ChannelState{
		IsOpen:               g.isOpen,
		Conductance:          effectiveConductance,
		EquilibriumPotential: g.reversalPot,
		CalciumLevel:         0.0, // GABA-A channels don't conduct calcium
	}
}

func (g *RealisticGabaAChannel) GetTrigger() ChannelTrigger {
	return ChannelTrigger{
		LigandThreshold:          5.0, // GABA concentration threshold (μM)
		ActivationTimeConstant:   2 * time.Millisecond,
		DeactivationTimeConstant: 100 * time.Millisecond,
	}
}
