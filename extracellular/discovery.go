package extracellular

// =================================================================================
// DISCOVERY SERVICE
// =================================================================================

// DiscoveryService helps components find each other
type DiscoveryService struct {
	registry *ComponentRegistry
}

// NewDiscoveryService creates a discovery service
func NewDiscoveryService(registry *ComponentRegistry) *DiscoveryService {
	return &DiscoveryService{
		registry: registry,
	}
}

// FindNearby finds components within a spatial radius
func (ds *DiscoveryService) FindNearby(position Position3D, radius float64) []ComponentInfo {
	criteria := ComponentCriteria{
		Position: &position,
		Radius:   radius,
	}
	return ds.registry.Find(criteria)
}

// FindByType finds components of a specific type
func (ds *DiscoveryService) FindByType(componentType ComponentType) []ComponentInfo {
	criteria := ComponentCriteria{
		Type: &componentType,
	}
	return ds.registry.Find(criteria)
}
