package component

import (
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types"
)

// ============================================================================
// CHEMICAL SIGNALING INTERFACES
// ============================================================================

// ChemicalReceiver interface for components that can receive chemical signals
type ChemicalReceiver interface {
	Component
	GetReceptors() []types.LigandType
	Bind(ligandType types.LigandType, sourceID string, concentration float64)
}

// ChemicalReleaser interface for components that can release chemical signals
type ChemicalReleaser interface {
	Component
	GetReleasedLigands() []types.LigandType
}

// ============================================================================
// ELECTRICAL SIGNALING INTERFACES
// ============================================================================

// ElectricalReceiver interface for components that can receive electrical signals
type ElectricalReceiver interface {
	Component
	OnSignal(signalType types.SignalType, sourceID string, data interface{})
}

// ElectricalTransmitter interface for components that can send electrical signals
type ElectricalTransmitter interface {
	Component
	GetSignalTypes() []types.SignalType
}

// ============================================================================
// MESSAGE PROCESSING INTERFACES
// ============================================================================

// MessageReceiver interface for components that can receive neural signals
type MessageReceiver interface {
	Component
	Receive(msg types.NeuralSignal)
}

// MessageTransmitter interface for components that can transmit neural signals
type MessageTransmitter interface {
	Component
	Transmit(signal float64) error
}

// MessageScheduler defines the contract for components that can schedule
// messages for delayed delivery to a target MessageReceiver.
// This interface is implemented by components (like neurons) that manage
// their own outgoing message queues.
type MessageScheduler interface {
	MessageReceiver // A component capable of scheduling messages is also a MessageReceiver

	// ScheduleDelayedDelivery queues a message for delivery after the specified delay.
	// This method returns immediately (non-blocking) and the message will be
	// delivered to the target after the delay period.
	//
	// Parameters:
	//   msg: The neural signal to deliver
	//   target: The receiving component.MessageReceiver (e.g., the postsynaptic neuron)
	//   delay: Total transmission delay (synaptic + axonal)
	ScheduleDelayedDelivery(msg types.NeuralSignal, target MessageReceiver, delay time.Duration)
}

// SynapticProcessor defines the universal contract for any component that acts
// as a synapse. It is the key to the pluggable architecture, ensuring that the
// Neuron can work with any synapse type that fulfills these methods.
type SynapticProcessor interface {
	// ID returns the unique identifier for the synapse.
	ID() string

	// Transmit processes an outgoing signal from the pre-synaptic neuron.
	// IMPORTANT: This method completes synchronously - it does NOT block
	// the calling neuron while waiting for delays. Instead, it schedules
	// delayed delivery through the neuron's axonal queue system.
	//
	// Parameters:
	//   signalValue: The strength of the signal from the pre-synaptic neuron
	Transmit(signalValue float64)

	// ApplyPlasticity updates the synapse's internal state based on feedback
	ApplyPlasticity(adjustment types.PlasticityAdjustment) // Use types.PlasticityAdjustment
	// Note: The specific plasticity types (STDP, BCM, etc.) should be in types/events.go or types/configs.go

	// ShouldPrune evaluates if the synapse should be removed
	ShouldPrune() bool

	// GetWeight returns the current synaptic weight
	GetWeight() float64

	// SetWeight allows direct manipulation of synaptic strength
	SetWeight(weight float64)

	// Add other common methods here if needed, like:
	GetActivityInfo() types.ActivityInfo // For health monitoring etc.
	GetLastActivity() time.Time
	Type() types.ComponentType
	Position() types.Position3D
	IsActive() bool
	GetPresynapticID() string
	GetPostsynapticID() string
	GetDelay() time.Duration
	GetPlasticityConfig() types.PlasticityConfig // If a generic config is needed
	UpdateWeight(event types.PlasticityEvent)
}

// ============================================================================
// SPATIAL AWARENESS INTERFACES
// ============================================================================

// SpatialComponent interface for components with spatial awareness
type SpatialComponent interface {
	Component
	SetPosition(position types.Position3D)
	GetRange() float64
}

// ============================================================================
// HEALTH MONITORING INTERFACES
// ============================================================================

// MonitorableComponent interface for components that provide health metrics
type MonitorableComponent interface {
	Component
	GetHealthMetrics() HealthMetrics
}

// ============================================================================
// SPECIALIZED COMPONENT IMPLEMENTATIONS
// ============================================================================

// DefaultSpatialComponent provides basic spatial functionality
type DefaultSpatialComponent struct {
	*BaseComponent
	range_ float64
}

// NewSpatialComponent creates a component with spatial capabilities
func NewSpatialComponent(id string, componentType types.ComponentType, position types.Position3D, range_ float64) *DefaultSpatialComponent {
	return &DefaultSpatialComponent{
		BaseComponent: NewBaseComponent(id, componentType, position),
		range_:        range_,
	}
}

func (sc *DefaultSpatialComponent) GetRange() float64 {
	return sc.range_
}

// ============================================================================
// HEALTH MONITORING IMPLEMENTATION
// ============================================================================

// DefaultMonitorableComponent provides basic health monitoring
type DefaultMonitorableComponent struct {
	*BaseComponent
	healthMetrics HealthMetrics
	healthMutex   sync.RWMutex
}

// NewMonitorableComponent creates a component with health monitoring
func NewMonitorableComponent(id string, componentType types.ComponentType, position types.Position3D) *DefaultMonitorableComponent {
	return &DefaultMonitorableComponent{
		BaseComponent: NewBaseComponent(id, componentType, position),
		healthMetrics: HealthMetrics{
			ActivityLevel:   0.0,
			ConnectionCount: 0,
			ProcessingLoad:  0.0,
			LastHealthCheck: time.Now(),
			HealthScore:     1.0,
			Issues:          make([]string, 0),
		},
	}
}

func (mc *DefaultMonitorableComponent) GetHealthMetrics() HealthMetrics {
	mc.healthMutex.RLock()
	defer mc.healthMutex.RUnlock()

	// Return a copy to prevent external modification
	metrics := mc.healthMetrics
	metrics.Issues = make([]string, len(mc.healthMetrics.Issues))
	copy(metrics.Issues, mc.healthMetrics.Issues)

	return metrics
}

func (mc *DefaultMonitorableComponent) UpdateHealthMetrics(metrics HealthMetrics) {
	mc.healthMutex.Lock()
	defer mc.healthMutex.Unlock()
	mc.healthMetrics = metrics
	mc.healthMetrics.LastHealthCheck = time.Now()
	mc.UpdateMetadata("last_health_check", time.Now())
}
