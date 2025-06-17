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

This implementation uses efficient spatial indexing to scale to tens of thousands
of components while maintaining biological accuracy and thread safety.

PERFORMANCE CHARACTERISTICS:
- Component registration: O(1) average case
- Spatial queries: O(k) where k = components in relevant grid cells
- Memory usage: Linear with components + sparse grid overhead
- Concurrency: Lock-free reads, minimal write contention
=================================================================================
*/

package extracellular

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// =================================================================================
// SPATIAL INDEXING STRUCTURES
// =================================================================================

// GridKey represents coordinates in the 3D spatial grid
// Each grid cell contains components within a cubic volume of space
type GridKey struct {
	X, Y, Z int64
}

// GridCell represents a single cell in the 3D spatial grid
// BIOLOGICAL ANALOGY: Similar to how astrocytes divide brain tissue into
// non-overlapping domains, grid cells partition 3D space for efficient lookup
type GridCell struct {
	components map[string]ComponentInfo // Components in this cell
	mu         sync.RWMutex             // Fine-grained locking per cell
}

// SpatialGrid provides efficient 3D spatial indexing for biological components
// BIOLOGICAL INSPIRATION: Models how astrocytes organize their spatial awareness
// through territorial domains that allow rapid identification of nearby components
type SpatialGrid struct {
	cellSize float64               // Size of each grid cell in micrometers
	cells    map[GridKey]*GridCell // Sparse grid storage
	mu       sync.RWMutex          // Protects grid structure modifications
}

// =================================================================================
// CORE ASTROCYTE NETWORK STRUCTURES
// =================================================================================

// Territory represents an astrocyte's spatial monitoring domain
// BIOLOGICAL FUNCTION: Real astrocytes establish territorial domains spanning
// 50-100μm radius, monitoring thousands of synapses within their territory
type Territory struct {
	AstrocyteID  string     `json:"astrocyte_id"`  // Owner astrocyte identifier
	Center       Position3D `json:"center"`        // Territorial center point
	Radius       float64    `json:"radius"`        // Territorial radius in μm
	MonitoredIDs []string   `json:"monitored_ids"` // Components under surveillance
	LastActivity time.Time  `json:"last_activity"` // Most recent territorial activity
}

// SynapticInfo tracks detailed synaptic connections and activity
// BIOLOGICAL FUNCTION: Astrocytes monitor synaptic transmission strength,
// frequency, and plasticity changes to coordinate neural network function
type SynapticInfo struct {
	PresynapticID  string    `json:"presynaptic_id"`  // Source neuron identifier
	PostsynapticID string    `json:"postsynaptic_id"` // Target neuron identifier
	SynapseID      string    `json:"synapse_id"`      // Synapse identifier
	Strength       float64   `json:"strength"`        // Current synaptic strength
	LastActivity   time.Time `json:"last_activity"`   // Last transmission time
	ActivityCount  int64     `json:"activity_count"`  // Total activity events
}

// AstrocyteNetwork tracks network components and connectivity like biological astrocytes
// BIOLOGICAL MODEL: Represents the distributed astrocyte network that provides
// spatial organization, component tracking, and connectivity mapping for neural circuits
type AstrocyteNetwork struct {
	// === COMPONENT REGISTRY ===
	// Central registry of all neural components (neurons, synapses, gates)
	// BIOLOGICAL FUNCTION: Like astrocyte memory of all components in their domain
	components map[string]ComponentInfo

	// === SPATIAL INDEXING ===
	// Efficient spatial lookup system for proximity-based queries
	// BIOLOGICAL FUNCTION: Models astrocyte spatial awareness and territorial organization
	spatialGrid *SpatialGrid

	// === CONNECTIVITY MAPPING ===
	// Network topology tracking and synaptic relationship monitoring
	// BIOLOGICAL FUNCTION: Astrocytes maintain detailed maps of neural connectivity
	connections map[string][]string     // Component -> connected components
	synapticMap map[string]SynapticInfo // Detailed synaptic information

	// === TERRITORIAL MANAGEMENT ===
	// Astrocyte territorial domains and monitoring responsibilities
	// BIOLOGICAL FUNCTION: Models how astrocytes divide brain tissue into domains
	territories map[string]Territory

	// === CONCURRENCY CONTROL ===
	// Thread-safe access coordination for concurrent neural operations
	mu sync.RWMutex
}

// =================================================================================
// CONSTRUCTOR AND INITIALIZATION
// =================================================================================

// NewAstrocyteNetwork creates a biological component tracking network
// BIOLOGICAL INITIALIZATION: Sets up the distributed astrocyte network with
// spatial indexing optimized for typical neural component densities and sizes
func NewAstrocyteNetwork() *AstrocyteNetwork {
	return &AstrocyteNetwork{
		components:  make(map[string]ComponentInfo),
		spatialGrid: newSpatialGrid(50.0), // 50μm grid cells (typical astrocyte domain size)
		connections: make(map[string][]string),
		synapticMap: make(map[string]SynapticInfo),
		territories: make(map[string]Territory),
	}
}

// newSpatialGrid creates an efficient 3D spatial indexing system
// BIOLOGICAL OPTIMIZATION: Grid cell size chosen to match typical astrocyte
// territorial domains (~50μm) for optimal spatial query performance
func newSpatialGrid(cellSize float64) *SpatialGrid {
	return &SpatialGrid{
		cellSize: cellSize,
		cells:    make(map[GridKey]*GridCell),
	}
}

// =================================================================================
// COMPONENT REGISTRATION AND LIFECYCLE MANAGEMENT
// =================================================================================

// Register adds a neural component to the astrocyte network monitoring system
// BIOLOGICAL FUNCTION: Models how astrocytes detect and begin monitoring new
// neural components that appear in their territorial domains
//
// PROCESS MODELED:
// 1. Component identity validation and timestamping
// 2. Addition to central component registry
// 3. Spatial indexing for proximity-based queries
// 4. Connection list initialization for connectivity tracking
//
// THREAD SAFETY: Fully thread-safe for concurrent component registration
func (an *AstrocyteNetwork) Register(info ComponentInfo) error {
	if info.ID == "" {
		return fmt.Errorf("component ID cannot be empty")
	}

	// Set registration timestamp if not provided
	if info.RegisteredAt.IsZero() {
		info.RegisteredAt = time.Now()
	}

	an.mu.Lock()
	defer an.mu.Unlock()

	// Add to central component registry
	an.components[info.ID] = info

	// Add to spatial indexing system for efficient proximity queries
	an.spatialGrid.addComponent(info)

	// Initialize empty connection list for connectivity tracking
	if an.connections[info.ID] == nil {
		an.connections[info.ID] = make([]string, 0)
	}

	return nil
}

// Unregister removes a component from astrocyte network monitoring
// BIOLOGICAL FUNCTION: Models component death, migration, or silencing where
// astrocytes cease monitoring and clean up all associated connectivity data
//
// CLEANUP PROCESS:
// 1. Remove from central registry and spatial indexing
// 2. Clean up all connection mappings (incoming and outgoing)
// 3. Remove synaptic information involving this component
// 4. Maintain network integrity after component removal
//
// BIOLOGICAL ANALOGY: Like microglial cleanup after cell death
func (an *AstrocyteNetwork) Unregister(id string) error {
	an.mu.Lock()
	defer an.mu.Unlock()

	// Get component info before removal for spatial cleanup
	if info, exists := an.components[id]; exists {
		an.spatialGrid.removeComponent(info)
	}

	// Remove from central registry
	delete(an.components, id)

	// Clean up connectivity mappings
	delete(an.connections, id)

	// Remove from other components' connection lists
	for componentID, connList := range an.connections {
		an.connections[componentID] = an.removeFromSlice(connList, id)
	}

	// Clean up synaptic information
	for synapseID, synInfo := range an.synapticMap {
		if synInfo.PresynapticID == id || synInfo.PostsynapticID == id {
			delete(an.synapticMap, synapseID)
		}
	}

	return nil
}

// Get retrieves component information by identifier
// BIOLOGICAL FUNCTION: Models astrocyte recall of specific component information
// from their territorial monitoring records
func (an *AstrocyteNetwork) Get(id string) (ComponentInfo, bool) {
	an.mu.RLock()
	defer an.mu.RUnlock()

	info, exists := an.components[id]
	return info, exists
}

// List returns all registered components under astrocyte monitoring
// BIOLOGICAL FUNCTION: Provides complete inventory of all neural components
// within the astrocyte network's collective territorial domains
func (an *AstrocyteNetwork) List() []ComponentInfo {
	an.mu.RLock()
	defer an.mu.RUnlock()

	results := make([]ComponentInfo, 0, len(an.components))
	for _, info := range an.components {
		results = append(results, info)
	}

	return results
}

// Count returns the total number of components under astrocyte monitoring
// BIOLOGICAL FUNCTION: Provides network size metrics for territorial
// load assessment and resource allocation decisions
func (an *AstrocyteNetwork) Count() int {
	an.mu.RLock()
	defer an.mu.RUnlock()

	return len(an.components)
}

// UpdateState modifies a component's functional state
// BIOLOGICAL FUNCTION: Models how astrocytes track component state changes
// (active, inactive, shutting down) for network health monitoring
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
// SPATIAL DISCOVERY AND PROXIMITY QUERIES
// =================================================================================

// FindNearby efficiently locates components within spatial radius
// BIOLOGICAL FUNCTION: Models astrocyte spatial awareness - ability to rapidly
// identify all neural components within a specified distance from any point
//
// OPTIMIZATION: Uses spatial grid indexing to achieve O(k) performance where
// k = components in relevant grid cells, rather than O(N) linear scan
//
// BIOLOGICAL RELEVANCE: Essential for modeling diffusion, local connectivity,
// territorial overlap assessment, and proximity-based neural interactions
func (an *AstrocyteNetwork) FindNearby(position Position3D, radius float64) []ComponentInfo {
	criteria := ComponentCriteria{
		Position: &position,
		Radius:   radius,
	}
	return an.Find(criteria)
}

// FindByType locates all components of a specific biological type
// BIOLOGICAL FUNCTION: Allows astrocytes to selectively monitor specific
// component types (neurons vs synapses vs gates) for specialized functions
func (an *AstrocyteNetwork) FindByType(componentType ComponentType) []ComponentInfo {
	criteria := ComponentCriteria{
		Type: &componentType,
	}
	return an.Find(criteria)
}

// Find performs sophisticated component discovery with multiple criteria
// BIOLOGICAL FUNCTION: Models complex astrocyte queries combining spatial,
// functional, and state-based criteria for targeted component identification
//
// QUERY CAPABILITIES:
// - Type filtering (neurons, synapses, gates)
// - State filtering (active, inactive, shutting down)
// - Spatial filtering (proximity-based with radius)
// - Combined criteria for complex biological queries
func (an *AstrocyteNetwork) Find(criteria ComponentCriteria) []ComponentInfo {
	an.mu.RLock()
	defer an.mu.RUnlock()

	// Use spatial indexing for proximity queries when applicable
	if criteria.Position != nil && criteria.Radius > 0 {
		return an.spatialGrid.findNearby(*criteria.Position, criteria.Radius, criteria)
	}

	// Fall back to full scan for non-spatial queries
	var results []ComponentInfo
	for _, info := range an.components {
		if an.matches(info, criteria) {
			results = append(results, info)
		}
	}

	return results
}

// =================================================================================
// CONNECTIVITY MAPPING AND SYNAPTIC TRACKING
// =================================================================================

// MapConnection records a directed connection between neural components
// BIOLOGICAL FUNCTION: Models how astrocytes track neural connectivity patterns
// by observing synaptic formation and axonal pathfinding
//
// CONNECTION TRACKING: Maintains directed graph of neural connectivity for
// network topology analysis and pathway discovery
func (an *AstrocyteNetwork) MapConnection(fromID, toID string) error {
	an.mu.Lock()
	defer an.mu.Unlock()

	// Verify both components exist in the network
	if _, exists := an.components[fromID]; !exists {
		return fmt.Errorf("source component %s not found", fromID)
	}
	if _, exists := an.components[toID]; !exists {
		return fmt.Errorf("target component %s not found", toID)
	}

	// Initialize connection list if needed
	if an.connections[fromID] == nil {
		an.connections[fromID] = make([]string, 0)
	}

	// Avoid duplicate connections
	for _, connID := range an.connections[fromID] {
		if connID == toID {
			return nil // Connection already exists
		}
	}

	// Add new connection
	an.connections[fromID] = append(an.connections[fromID], toID)
	return nil
}

// RecordSynapticActivity tracks detailed synaptic transmission events
// BIOLOGICAL FUNCTION: Models how astrocytes monitor synaptic activity to
// assess connection strength, plasticity, and network dynamics
//
// SYNAPTIC MONITORING: Records transmission strength, timing, and frequency
// for synaptic plasticity analysis and network optimization
func (an *AstrocyteNetwork) RecordSynapticActivity(synapseID, preID, postID string, strength float64) error {
	an.mu.Lock()
	defer an.mu.Unlock()

	// Verify components exist
	if _, exists := an.components[preID]; !exists {
		return fmt.Errorf("presynaptic component %s not found", preID)
	}
	if _, exists := an.components[postID]; !exists {
		return fmt.Errorf("postsynaptic component %s not found", postID)
	}

	// Update or create synaptic information
	if existing, exists := an.synapticMap[synapseID]; exists {
		// Update existing synaptic record
		existing.Strength = strength
		existing.LastActivity = time.Now()
		existing.ActivityCount++
		an.synapticMap[synapseID] = existing
	} else {
		// Create new synaptic record
		an.synapticMap[synapseID] = SynapticInfo{
			PresynapticID:  preID,
			PostsynapticID: postID,
			SynapseID:      synapseID,
			Strength:       strength,
			LastActivity:   time.Now(),
			ActivityCount:  1,
		}
	}

	// Ensure basic connectivity mapping exists
	if an.connections[preID] == nil {
		an.connections[preID] = make([]string, 0)
	}

	// Add connection if not already present
	connectionExists := false
	for _, connID := range an.connections[preID] {
		if connID == postID {
			connectionExists = true
			break
		}
	}

	if !connectionExists {
		an.connections[preID] = append(an.connections[preID], postID)
	}

	return nil
}

// GetConnections retrieves all components connected to the specified component
// BIOLOGICAL FUNCTION: Provides astrocyte view of component's connectivity
// pattern for network topology analysis and pathway tracing
func (an *AstrocyteNetwork) GetConnections(componentID string) []string {
	an.mu.RLock()
	defer an.mu.RUnlock()

	connections := an.connections[componentID]
	if connections == nil {
		return []string{}
	}

	// Return defensive copy to prevent concurrent modification
	result := make([]string, len(connections))
	copy(result, connections)
	return result
}

// GetSynapticInfo retrieves detailed synaptic connection information
// BIOLOGICAL FUNCTION: Provides astrocyte monitoring data for specific
// synaptic connections including strength, activity, and timing information
func (an *AstrocyteNetwork) GetSynapticInfo(synapseID string) (SynapticInfo, bool) {
	an.mu.RLock()
	defer an.mu.RUnlock()

	info, exists := an.synapticMap[synapseID]
	return info, exists
}

// =================================================================================
// TERRITORIAL MANAGEMENT
// =================================================================================

// EstablishTerritory creates an astrocyte territorial monitoring domain
// BIOLOGICAL FUNCTION: Models how astrocytes establish spatial domains for
// neural component monitoring, typically 50-100μm radius spherical regions
//
// TERRITORIAL ORGANIZATION: Each astrocyte monitors a specific brain region,
// providing localized oversight of neural components and their interactions
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

// GetTerritory retrieves astrocyte territorial information
// BIOLOGICAL FUNCTION: Provides access to astrocyte domain specifications
// including spatial boundaries and monitored component lists
func (an *AstrocyteNetwork) GetTerritory(astrocyteID string) (Territory, bool) {
	an.mu.RLock()
	defer an.mu.RUnlock()

	territory, exists := an.territories[astrocyteID]
	return territory, exists
}

// ValidateAstrocyteLoad checks and adjusts territorial monitoring load
// BIOLOGICAL FUNCTION: Models astrocyte territorial adjustment when monitoring
// capacity is exceeded - biological astrocytes can adjust territory size to
// maintain effective monitoring of neural components
//
// LOAD BALANCING: When too many neurons exist in a territory, astrocytes
// reduce territorial radius to maintain manageable monitoring load
func (an *AstrocyteNetwork) ValidateAstrocyteLoad(astrocyteID string, maxNeurons int) error {
	// Get territory information with minimal lock contention
	an.mu.RLock()
	territory, exists := an.territories[astrocyteID]
	an.mu.RUnlock()

	if !exists {
		return fmt.Errorf("astrocyte %s not found", astrocyteID)
	}

	// Count neurons in territory using spatial indexing
	neuronsInTerritory := an.FindNearby(territory.Center, territory.Radius)
	neuronCount := 0
	for _, comp := range neuronsInTerritory {
		if comp.Type == ComponentNeuron {
			neuronCount++
		}
	}

	// Adjust territory size if overloaded
	if neuronCount > maxNeurons {
		an.mu.Lock()
		defer an.mu.Unlock()

		// Re-verify territory exists (could have been deleted)
		territory, exists = an.territories[astrocyteID]
		if !exists {
			return fmt.Errorf("astrocyte %s not found", astrocyteID)
		}

		originalRadius := territory.Radius

		// Calculate new radius using biological scaling principle
		// Territory area scales with monitoring capacity: newRadius = oldRadius * sqrt(targetLoad/currentLoad)
		ratio := float64(maxNeurons) / float64(neuronCount)
		territory.Radius = originalRadius * math.Sqrt(ratio)
		an.territories[astrocyteID] = territory

		return fmt.Errorf("astrocyte %s territory adjusted: radius %.1f→%.1f to manage %d→%d neurons",
			astrocyteID, originalRadius, territory.Radius, neuronCount, maxNeurons)
	}

	return nil
}

// =================================================================================
// SPATIAL INDEXING IMPLEMENTATION
// =================================================================================

// addComponent adds a component to the spatial indexing grid
// BIOLOGICAL FUNCTION: Models astrocyte spatial memory - ability to quickly
// recall which components exist in each region of brain tissue
func (sg *SpatialGrid) addComponent(info ComponentInfo) {
	key := sg.positionToGridKey(info.Position)

	sg.mu.Lock()
	defer sg.mu.Unlock()

	// Create grid cell if it doesn't exist
	if sg.cells[key] == nil {
		sg.cells[key] = &GridCell{
			components: make(map[string]ComponentInfo),
		}
	}

	// Add component to appropriate grid cell
	sg.cells[key].mu.Lock()
	sg.cells[key].components[info.ID] = info
	sg.cells[key].mu.Unlock()
}

// removeComponent removes a component from the spatial indexing grid
func (sg *SpatialGrid) removeComponent(info ComponentInfo) {
	key := sg.positionToGridKey(info.Position)

	sg.mu.RLock()
	cell := sg.cells[key]
	sg.mu.RUnlock()

	if cell != nil {
		cell.mu.Lock()
		delete(cell.components, info.ID)
		cell.mu.Unlock()
	}
}

// findNearby performs efficient spatial query using grid indexing
// OPTIMIZATION: Only checks grid cells that intersect with query radius
func (sg *SpatialGrid) findNearby(position Position3D, radius float64, criteria ComponentCriteria) []ComponentInfo {
	// Calculate grid cells that could contain relevant components
	cellsToCheck := sg.getCellsInRadius(position, radius)

	var results []ComponentInfo

	sg.mu.RLock()
	defer sg.mu.RUnlock()

	// Check each relevant grid cell
	for _, key := range cellsToCheck {
		if cell := sg.cells[key]; cell != nil {
			cell.mu.RLock()
			for _, info := range cell.components {
				if sg.matches(info, criteria, position, radius) {
					results = append(results, info)
				}
			}
			cell.mu.RUnlock()
		}
	}

	return results
}

// positionToGridKey converts 3D position to grid coordinates
// Handles extreme coordinates gracefully by clamping to reasonable bounds
func (sg *SpatialGrid) positionToGridKey(pos Position3D) GridKey {
	// Handle special floating point values
	if math.IsInf(pos.X, 0) || math.IsNaN(pos.X) {
		pos.X = 0
	}
	if math.IsInf(pos.Y, 0) || math.IsNaN(pos.Y) {
		pos.Y = 0
	}
	if math.IsInf(pos.Z, 0) || math.IsNaN(pos.Z) {
		pos.Z = 0
	}

	// Clamp extremely large coordinates to prevent integer overflow
	const maxCoord = 1e12 // Reasonable maximum for biological simulations
	if math.Abs(pos.X) > maxCoord {
		pos.X = math.Copysign(maxCoord, pos.X)
	}
	if math.Abs(pos.Y) > maxCoord {
		pos.Y = math.Copysign(maxCoord, pos.Y)
	}
	if math.Abs(pos.Z) > maxCoord {
		pos.Z = math.Copysign(maxCoord, pos.Z)
	}

	return GridKey{
		X: int64(math.Floor(pos.X / sg.cellSize)),
		Y: int64(math.Floor(pos.Y / sg.cellSize)),
		Z: int64(math.Floor(pos.Z / sg.cellSize)),
	}
}

// getCellsInRadius determines which grid cells intersect with query sphere
// Optimized to handle extreme coordinates and large radii efficiently
func (sg *SpatialGrid) getCellsInRadius(center Position3D, radius float64) []GridKey {
	// Handle extreme radii
	if radius <= 0 {
		return []GridKey{sg.positionToGridKey(center)}
	}

	// Limit maximum search area to prevent excessive iteration
	const maxCellsPerDimension = 100 // Prevents exponential explosion
	cellRadius := int64(math.Ceil(radius / sg.cellSize))
	if cellRadius > maxCellsPerDimension {
		cellRadius = maxCellsPerDimension
	}

	centerKey := sg.positionToGridKey(center)

	var keys []GridKey
	// Pre-allocate with reasonable estimate
	keys = make([]GridKey, 0, (2*cellRadius+1)*(2*cellRadius+1)*(2*cellRadius+1))

	for x := centerKey.X - cellRadius; x <= centerKey.X+cellRadius; x++ {
		for y := centerKey.Y - cellRadius; y <= centerKey.Y+cellRadius; y++ {
			for z := centerKey.Z - cellRadius; z <= centerKey.Z+cellRadius; z++ {
				keys = append(keys, GridKey{X: x, Y: y, Z: z})
			}
		}
	}

	return keys
}

// =================================================================================
// UTILITY FUNCTIONS
// =================================================================================

// matches checks if component satisfies search criteria with spatial optimization
func (sg *SpatialGrid) matches(info ComponentInfo, criteria ComponentCriteria, queryPos Position3D, queryRadius float64) bool {
	// Type filter
	if criteria.Type != nil && info.Type != *criteria.Type {
		return false
	}

	// State filter
	if criteria.State != nil && info.State != *criteria.State {
		return false
	}

	// Spatial filter with precise distance calculation
	distance := sg.calculateDistance(info.Position, queryPos)

	// Handle special floating point values
	if math.IsInf(distance, 0) || math.IsNaN(distance) {
		return false // Infinite or NaN distances don't match
	}

	const epsilon = 1e-9 // Floating point tolerance
	return distance <= queryRadius+epsilon
}

// matches checks if component satisfies general search criteria (non-spatial)
func (an *AstrocyteNetwork) matches(info ComponentInfo, criteria ComponentCriteria) bool {
	// Type filter
	if criteria.Type != nil && info.Type != *criteria.Type {
		return false
	}

	// State filter
	if criteria.State != nil && info.State != *criteria.State {
		return false
	}

	// Spatial filter
	if criteria.Position != nil {
		if criteria.Radius == 0.0 {
			// Exact position match
			return info.Position.X == criteria.Position.X &&
				info.Position.Y == criteria.Position.Y &&
				info.Position.Z == criteria.Position.Z
		}

		if criteria.Radius > 0.0 {
			distance := an.Distance(info.Position, *criteria.Position)
			const epsilon = 1e-9
			return distance <= criteria.Radius+epsilon
		}
	}

	return true
}

// calculateDistance computes 3D Euclidean distance between positions
func (sg *SpatialGrid) calculateDistance(pos1, pos2 Position3D) float64 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	dz := pos1.Z - pos2.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// Distance calculates 3D Euclidean distance between positions
// BIOLOGICAL FUNCTION: Models astrocyte spatial measurement capabilities
// for territorial organization and proximity assessment
func (an *AstrocyteNetwork) Distance(pos1, pos2 Position3D) float64 {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	dz := pos1.Z - pos2.Z
	return math.Sqrt(dx*dx + dy*dy + dz*dz)
}

// removeFromSlice efficiently removes an element from a string slice
func (an *AstrocyteNetwork) removeFromSlice(slice []string, element string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item != element {
			result = append(result, item)
		}
	}
	return result
}
