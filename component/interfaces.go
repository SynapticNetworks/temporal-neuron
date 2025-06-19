package component

import (
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/message"
)

// ============================================================================
// CHEMICAL SIGNALING INTERFACES
// ============================================================================

// ChemicalReceiver interface for components that can receive chemical signals
type ChemicalReceiver interface {
	Component
	GetReceptors() []message.LigandType
	Bind(ligandType message.LigandType, sourceID string, concentration float64)
}

// ChemicalReleaser interface for components that can release chemical signals
type ChemicalReleaser interface {
	Component
	GetReleasedLigands() []message.LigandType
}

// ============================================================================
// ELECTRICAL SIGNALING INTERFACES
// ============================================================================

// ElectricalReceiver interface for components that can receive electrical signals
type ElectricalReceiver interface {
	Component
	OnSignal(signalType message.SignalType, sourceID string, data interface{})
}

// ElectricalTransmitter interface for components that can send electrical signals
type ElectricalTransmitter interface {
	Component
	GetSignalTypes() []message.SignalType
}

// ============================================================================
// MESSAGE PROCESSING INTERFACES
// ============================================================================

// MessageReceiver interface for components that can receive neural signals
type MessageReceiver interface {
	Component
	Receive(msg message.NeuralSignal)
}

// MessageTransmitter interface for components that can transmit neural signals
type MessageTransmitter interface {
	Component
	Transmit(signal float64) error
}

// ============================================================================
// SPATIAL AWARENESS INTERFACES
// ============================================================================

// SpatialComponent interface for components with spatial awareness
type SpatialComponent interface {
	Component
	SetPosition(position Position3D)
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
func NewSpatialComponent(id string, componentType ComponentType, position Position3D, range_ float64) *DefaultSpatialComponent {
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
func NewMonitorableComponent(id string, componentType ComponentType, position Position3D) *DefaultMonitorableComponent {
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
