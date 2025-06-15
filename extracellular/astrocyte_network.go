/*
=================================================================================
ASTROCYTE NETWORK - BIOLOGICAL COMPONENT TRACKING AND CONNECTIVITY
=================================================================================

Models the astrocyte network that maintains detailed maps of neural connectivity.
Astrocytes monitor synaptic activity, guide growth, and coordinate information
flow between neural components. They are the "living registry" of the brain.

BIOLOGICAL FUNCTIONS:
- Track all neural components in their territorial domains
- Maintain connectivity maps between neurons and synapses
- Guide growth and connection formation through spatial awareness
- Coordinate activity patterns across neural regions
- Provide discovery services for nearby components

Combines the functions of ComponentRegistry and DiscoveryService into a
single biologically-inspired astrocyte network coordination system.
=================================================================================
*/

package extracellular

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// AstrocyteNetwork tracks network components and connectivity like biological astrocytes
type AstrocyteNetwork struct {
	// === COMPONENT TRACKING ===
	components map[string]ComponentInfo // All tracked neural components

	// === CONNECTIVITY MAPPING ===
	connections map[string][]string     // Component ID -> Connected Component IDs
	synapticMap map[string]SynapticInfo // Detailed synaptic connectivity

	// === SPATIAL ORGANIZATION ===
	territories map[string]Territory // Astrocyte territorial domains

	// === CONCURRENCY CONTROL ===
	mu sync.RWMutex
}

// Territory represents an astrocyte's spatial monitoring domain
type Territory struct {
	AstrocyteID  string     `json:"astrocyte_id"`
	Center       Position3D `json:"center"`
	Radius       float64    `json:"radius"`
	MonitoredIDs []string   `json:"monitored_ids"`
	LastActivity time.Time  `json:"last_activity"`
}

// SynapticInfo tracks detailed synaptic connections
type SynapticInfo struct {
	PresynapticID  string    `json:"presynaptic_id"`
	PostsynapticID string    `json:"postsynaptic_id"`
	SynapseID      string    `json:"synapse_id"`
	Strength       float64   `json:"strength"`
	LastActivity   time.Time `json:"last_activity"`
	ActivityCount  int64     `json:"activity_count"`
}

// NewAstrocyteNetwork creates a biological component tracking network
func NewAstrocyteNetwork() *AstrocyteNetwork {
	return &AstrocyteNetwork{
		components:  make(map[string]ComponentInfo),
		connections: make(map[string][]string),
		synapticMap: make(map[string]SynapticInfo),
		territories: make(map[string]Territory),
	}
}

// =================================================================================
// COMPONENT REGISTRATION AND TRACKING (was ComponentRegistry functions)
// =================================================================================

// Register adds a component to the astrocyte network
func (an *AstrocyteNetwork) Register(info ComponentInfo) error {
	an.mu.Lock()
	defer an.mu.Unlock()

	if info.ID == "" {
		return fmt.Errorf("component ID cannot be empty")
	}

	// Set registration time if not provided
	if info.RegisteredAt.IsZero() {
		info.RegisteredAt = time.Now()
	}

	an.components[info.ID] = info

	// Initialize empty connection list
	if an.connections[info.ID] == nil {
		an.connections[info.ID] = make([]string, 0)
	}

	return nil
}

// Unregister removes a component from the astrocyte network
func (an *AstrocyteNetwork) Unregister(id string) error {
	an.mu.Lock()
	defer an.mu.Unlock()

	// Remove component
	delete(an.components, id)

	// Remove connections (cleanup like biological microglia)
	delete(an.connections, id)

	// Remove from other components' connection lists
	for componentID, connList := range an.connections {
		an.connections[componentID] = an.removeFromSlice(connList, id)
	}

	// Remove from synaptic map
	for synapseID, synInfo := range an.synapticMap {
		if synInfo.PresynapticID == id || synInfo.PostsynapticID == id {
			delete(an.synapticMap, synapseID)
		}
	}

	return nil
}

// Get retrieves a component by ID
func (an *AstrocyteNetwork) Get(id string) (ComponentInfo, bool) {
	an.mu.RLock()
	defer an.mu.RUnlock()

	info, exists := an.components[id]
	return info, exists
}

// List returns all registered components
func (an *AstrocyteNetwork) List() []ComponentInfo {
	an.mu.RLock()
	defer an.mu.RUnlock()

	results := make([]ComponentInfo, 0, len(an.components))
	for _, info := range an.components {
		results = append(results, info)
	}

	return results
}

// Count returns the number of registered components
func (an *AstrocyteNetwork) Count() int {
	an.mu.RLock()
	defer an.mu.RUnlock()

	return len(an.components)
}

// UpdateState updates a component's state
func (an *AstrocyteNetwork) UpdateState(id string, state ComponentState) error {
	an.mu.Lock()
	defer an.mu.Unlock()

	if info, exists := an.components[id]; exists {
		info.State = state
		an.components[id] = info
		return nil
	}

	return fmt.Errorf("component %s not found", id)
}

// =================================================================================
// COMPONENT DISCOVERY (was DiscoveryService functions)
// =================================================================================

// Find searches for components matching criteria
func (an *AstrocyteNetwork) Find(criteria ComponentCriteria) []ComponentInfo {
	an.mu.RLock()
	defer an.mu.RUnlock()

	var results []ComponentInfo

	for _, info := range an.components {
		if an.matches(info, criteria) {
			results = append(results, info)
		}
	}

	return results
}

// FindNearby finds components within a spatial radius (astrocyte spatial awareness)
func (an *AstrocyteNetwork) FindNearby(position Position3D, radius float64) []ComponentInfo {
	criteria := ComponentCriteria{
		Position: &position,
		Radius:   radius,
	}
	return an.Find(criteria)
}

// FindByType finds components of a specific type
func (an *AstrocyteNetwork) FindByType(componentType ComponentType) []ComponentInfo {
	criteria := ComponentCriteria{
		Type: &componentType,
	}
	return an.Find(criteria)
}

// =================================================================================
// CONNECTIVITY MAPPING (NEW - Biological astrocyte function)
// =================================================================================

// MapConnection records a connection between components (astrocyte connectivity tracking)
func (an *AstrocyteNetwork) MapConnection(fromID, toID string) error {
	an.mu.Lock()
	defer an.mu.Unlock()

	// Ensure both components exist
	if _, exists := an.components[fromID]; !exists {
		return fmt.Errorf("source component %s not found", fromID)
	}
	if _, exists := an.components[toID]; !exists {
		return fmt.Errorf("target component %s not found", toID)
	}

	// Add connection
	if an.connections[fromID] == nil {
		an.connections[fromID] = make([]string, 0)
	}

	// Avoid duplicates
	for _, connID := range an.connections[fromID] {
		if connID == toID {
			return nil // Connection already exists
		}
	}

	an.connections[fromID] = append(an.connections[fromID], toID)
	return nil
}

// RecordSynapticActivity tracks synaptic connections with detailed info
func (an *AstrocyteNetwork) RecordSynapticActivity(synapseID, preID, postID string, strength float64) error {
	an.mu.Lock()
	defer an.mu.Unlock()

	synInfo := SynapticInfo{
		PresynapticID:  preID,
		PostsynapticID: postID,
		SynapseID:      synapseID,
		Strength:       strength,
		LastActivity:   time.Now(),
		ActivityCount:  1,
	}

	// Update existing or create new
	if existing, exists := an.synapticMap[synapseID]; exists {
		existing.Strength = strength
		existing.LastActivity = time.Now()
		existing.ActivityCount++
		an.synapticMap[synapseID] = existing
	} else {
		an.synapticMap[synapseID] = synInfo
	}

	// Also map the basic connection - INLINE instead of calling MapConnection
	// Ensure both components exist
	if _, exists := an.components[preID]; !exists {
		return fmt.Errorf("source component %s not found", preID)
	}
	if _, exists := an.components[postID]; !exists {
		return fmt.Errorf("target component %s not found", postID)
	}

	// Add connection inline (avoid mutex deadlock)
	if an.connections[preID] == nil {
		an.connections[preID] = make([]string, 0)
	}

	// Avoid duplicates
	for _, connID := range an.connections[preID] {
		if connID == postID {
			return nil // Connection already exists
		}
	}

	an.connections[preID] = append(an.connections[preID], postID)
	return nil
}

// GetConnections returns all components connected to the given component
func (an *AstrocyteNetwork) GetConnections(componentID string) []string {
	an.mu.RLock()
	defer an.mu.RUnlock()

	connections := an.connections[componentID]
	if connections == nil {
		return []string{}
	}

	// Return copy to avoid concurrent modification
	result := make([]string, len(connections))
	copy(result, connections)
	return result
}

// GetSynapticInfo returns detailed synaptic information
func (an *AstrocyteNetwork) GetSynapticInfo(synapseID string) (SynapticInfo, bool) {
	an.mu.RLock()
	defer an.mu.RUnlock()

	info, exists := an.synapticMap[synapseID]
	return info, exists
}

// =================================================================================
// TERRITORIAL MANAGEMENT (NEW - Biological astrocyte territories)
// =================================================================================

// EstablishTerritory creates an astrocyte territorial domain
func (an *AstrocyteNetwork) EstablishTerritory(astrocyteID string, center Position3D, radius float64) error {
	an.mu.Lock()
	defer an.mu.Unlock()

	territory := Territory{
		AstrocyteID:  astrocyteID,
		Center:       center,
		Radius:       radius,
		MonitoredIDs: make([]string, 0),
		LastActivity: time.Now(),
	}

	an.territories[astrocyteID] = territory
	return nil
}

// GetTerritory returns astrocyte territorial information
func (an *AstrocyteNetwork) GetTerritory(astrocyteID string) (Territory, bool) {
	an.mu.RLock()
	defer an.mu.RUnlock()

	territory, exists := an.territories[astrocyteID]
	return territory, exists
}

// =================================================================================
// UTILITY FUNCTIONS
// =================================================================================

// FIXED matches function - corrects spatial filtering bugs with proper floating point handling
func (an *AstrocyteNetwork) matches(info ComponentInfo, criteria ComponentCriteria) bool {
	// Check type filter
	if criteria.Type != nil && info.Type != *criteria.Type {
		return false
	}

	// Check state filter
	if criteria.State != nil && info.State != *criteria.State {
		return false
	}

	// Check spatial filter
	if criteria.Position != nil {
		// FIXED: Handle zero radius correctly
		if criteria.Radius == 0.0 {
			// Zero radius should only match components at exactly the same position
			return info.Position.X == criteria.Position.X &&
				info.Position.Y == criteria.Position.Y &&
				info.Position.Z == criteria.Position.Z
		}

		if criteria.Radius > 0.0 {
			// FIXED: Calculate actual distance and compare to radius with floating point tolerance
			distance := an.Distance(info.Position, *criteria.Position)

			// FIXED: Use small epsilon for floating point comparison to handle precision issues
			const epsilon = 1e-9
			return distance <= criteria.Radius+epsilon
		}
		// Note: negative radius is ignored (matches all components)
	}

	return true
}

// FIXED calculateDistance function - now returns actual distance for spatial queries
func (an *AstrocyteNetwork) calculateDistance(pos1, pos2 Position3D) float64 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	dz := pos1.Z - pos2.Z
	return dx*dx + dy*dy + dz*dz // Squared distance for performance
}

// Distance calculates actual 3D distance between positions
func (an *AstrocyteNetwork) Distance(pos1, pos2 Position3D) float64 {
	return math.Sqrt(an.calculateDistance(pos1, pos2))
}

// removeFromSlice removes an element from a string slice
func (an *AstrocyteNetwork) removeFromSlice(slice []string, element string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item != element {
			result = append(result, item)
		}
	}
	return result
}

// ADD this function to the end of astrocyte_network.go

// ValidateAstrocyteLoad checks and adjusts territory load
// FIXED: Resolves deadlock by releasing lock before spatial query
func (an *AstrocyteNetwork) ValidateAstrocyteLoad(astrocyteID string, maxNeurons int) error {
	// STEP 1: Get territory info with minimal lock time
	an.mu.RLock()
	territory, exists := an.territories[astrocyteID]
	an.mu.RUnlock() // ✅ RELEASE LOCK BEFORE SPATIAL QUERY

	if !exists {
		return fmt.Errorf("astrocyte %s not found", astrocyteID)
	}

	// STEP 2: Count neurons in territory (without holding lock)
	neuronsInTerritory := an.FindNearby(territory.Center, territory.Radius)
	neuronCount := 0
	for _, comp := range neuronsInTerritory {
		if comp.Type == ComponentNeuron {
			neuronCount++
		}
	}

	// STEP 3: If overloaded, adjust territory (reacquire lock for modification)
	if neuronCount > maxNeurons {
		an.mu.Lock() // ✅ ACQUIRE LOCK ONLY FOR MODIFICATION
		defer an.mu.Unlock()

		// Re-check territory still exists (could have been deleted)
		territory, exists = an.territories[astrocyteID]
		if !exists {
			return fmt.Errorf("astrocyte %s not found", astrocyteID)
		}

		// FIXED: Store original radius before modification
		originalRadius := territory.Radius

		// Calculate new radius to achieve target load
		// Target area = current area * (target neurons / current neurons)
		// Since area ∝ radius², new radius = current radius * sqrt(ratio)
		ratio := float64(maxNeurons) / float64(neuronCount)
		targetRadius := territory.Radius * math.Sqrt(ratio)

		// Update territory with reduced radius
		territory.Radius = targetRadius
		an.territories[astrocyteID] = territory

		return fmt.Errorf("astrocyte %s territory adjusted: radius %.1f→%.1f to manage %d→%d neurons",
			astrocyteID, originalRadius, targetRadius, neuronCount, maxNeurons)
	}

	return nil
}
