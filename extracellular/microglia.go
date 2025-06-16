/*
=================================================================================
MICROGLIA - BIOLOGICAL LIFECYCLE MANAGEMENT AND NEURAL MAINTENANCE
=================================================================================

Models microglial cells that serve as the brain's resident immune system and
maintenance crew. Microglia continuously patrol neural tissue, monitoring
component health, coordinating birth/death, and maintaining network integrity.

BIOLOGICAL FUNCTIONS:
- Component lifecycle coordination (neurogenesis, apoptosis)
- Synaptic pruning and structural optimization
- Neural health surveillance and damage detection
- Cleanup of dead components and debris
- Activity-dependent structural plasticity
- Resource allocation and metabolic coordination

---------------------------------------------------------------------------------
NEUROGENESIS & SYNAPTOGENESIS: A COLLABORATIVE PROCESS
---------------------------------------------------------------------------------
The creation of new neurons (neurogenesis) and synapses (synaptogenesis) is
handled as a collaborative process involving both the ExtracellularMatrix (via
Microglia) and other network components (like Neurons or higher-level plugins).
The matrix acts as the gatekeeper and executor, while other components act as
the requesters. A neuron cannot create another neuron directly, but it can
signal the *need* for one.

ROLE OF THE EXTRACELLULAR MATRIX (VIA MICROGLIA):
The Microglia system is the **sole authority for creating and destroying
components**. It acts as the biological resource manager and maintenance crew.

- Gatekeeper: It evaluates all requests for new components against the
  network's overall resource constraints (e.g., `MaxComponents`).
- Executor: If a request is approved, Microglia is responsible for actually
  creating the component, registering it with the AstrocyteNetwork, and
  initializing its health monitoring.
- Prioritization: It manages a queue of `ComponentBirthRequest` objects,
  processing them based on biological priority (e.g., an `PriorityEmergency`
  request can bypass normal resource limits).

ROLE OF NEURONS (AND OTHER COMPONENTS):
Individual neurons (or other logical units like plugins) are the **initiators**
of neurogenesis. They sense their local conditions and signal to the matrix
when a new component is needed.

- Requester: A component detects a condition that justifies creating a new
  neuron or synapse. The `TestMicrogliaBiologicalNeurogenesis` test gives
  perfect examples of these conditions:
    - "Critical damage response" (`PriorityEmergency`)
    - "High activity region overloaded" (`PriorityHigh`)
    - "Learning-induced demand" (`PriorityMedium`)
- Local Knowledge: A neuron knows its own firing rate and can determine if
  it's "overloaded." It can't, however, know the global state of the network
  or its resource limits. Therefore, it can only *request* and not *command*
  creation.

STEP-BY-STEP WORKFLOW:

Neurogenesis (Creating a New Neuron):
1. Initiation (Neuron/Plugin): An existing component (e.g., a controller
   plugin) detects that a cluster of neurons is consistently overloaded.
2. Request (Neuron/Plugin): The controller creates a `ComponentBirthRequest`,
   setting `ComponentType: ComponentNeuron`, specifying a position, and
   assigning a `Priority` and `Justification`.
3. Submission (to Matrix): The controller calls
   `microglia.RequestComponentBirth(request)` to submit the request.
4. Evaluation (Matrix/Microglia): The Microglia system processes the request,
   checking resource limits or priority bypasses.
5. Execution (Matrix/Microglia): If approved, Microglia creates the new
   `ComponentInfo`, registers it, and initializes its health stats. A new,
   unconnected neuron now exists.

Synaptogenesis (Creating a New Synapse to Connect Neurons):
This process is more complex and highlights the collaboration.
1. Initiation (Neuron/Plugin): A neuron (Neuron A) needs to connect to
   another (Neuron B), driven by a learning algorithm or growth controller.
2. Request Synapse Component (to Matrix): The logic submits a
   `ComponentBirthRequest` for a `ComponentType: ComponentSynapse`.
3. Synapse Creation (Matrix/Microglia): Microglia approves and creates the
   synapse component. It now exists but is not connected.
4. Forming the Connection (Neuron): The controlling logic instantiates a
   `synapse.SynapticProcessor` and configures it to target Neuron B.
5. Neuron A's Role: Calls its own `neuron.AddOutputSynapse()` method, adding
   the new synapse processor to its outputs.
6. Neuron B's Role: Upon receiving the first signal, Neuron B's
   `applyPostSynapticGainUnsafe` method automatically registers Neuron A
   and sets a default input gain, integrating the new connection.

BIOLOGICAL ANALOGY:
This model is analogous to real biology:
- Neurons (Requesters): Active neurons release chemical **growth factors** (like
  BDNF). This is a local signal indicating a need for more connections or
  support, equivalent to a `ComponentBirthRequest`.
- Microglia (Gatekeepers): Microglia and other glial cells respond to these
  growth factors, but their ability to support new growth is limited by
  **local metabolic resources** (oxygen, glucose), equivalent to checking
  `MaxComponents`. They facilitate creation only if the environment can support it.

Extends basic lifecycle management with sophisticated biological monitoring
and adaptive maintenance based on neural activity patterns.
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
// CONFIGURATION STRUCTURES
// =================================================================================

// MicrogliaConfig defines all configurable parameters for microglial behavior
type MicrogliaConfig struct {
	HealthThresholds HealthScoringConfig `json:"health_thresholds"`
	PruningSettings  PruningConfig       `json:"pruning_settings"`
	PatrolSettings   PatrolConfig        `json:"patrol_settings"`
	ResourceLimits   ResourceConfig      `json:"resource_limits"`
}

// HealthScoringConfig controls health assessment and issue detection
type HealthScoringConfig struct {
	// Activity level thresholds (0.0-1.0)
	CriticalActivityThreshold float64 `json:"critical_activity_threshold"` // Below this = critically low
	VeryLowActivityThreshold  float64 `json:"very_low_activity_threshold"` // Below this = very low
	LowActivityThreshold      float64 `json:"low_activity_threshold"`      // Below this = low
	ModerateActivityThreshold float64 `json:"moderate_activity_threshold"` // Below this = moderate low

	// Connection count thresholds
	IsolatedConnectionThreshold int `json:"isolated_connection_threshold"` // 0 connections
	PoorConnectionThreshold     int `json:"poor_connection_threshold"`     // Few connections
	MinHealthyConnections       int `json:"min_healthy_connections"`       // Minimum for good health

	// Health score multipliers (0.0-1.0, applied to base score of 1.0)
	CriticalActivityPenalty  float64 `json:"critical_activity_penalty"`  // Multiplier for critical activity
	LowActivityPenalty       float64 `json:"low_activity_penalty"`       // Multiplier for low activity
	ModerateActivityPenalty  float64 `json:"moderate_activity_penalty"`  // Multiplier for moderate activity
	PoorConnectionPenalty    float64 `json:"poor_connection_penalty"`    // Multiplier for poor connections
	FewConnectionPenalty     float64 `json:"few_connection_penalty"`     // Multiplier for few connections
	StalenessHourPenalty     float64 `json:"staleness_hour_penalty"`     // Multiplier for 1+ hour staleness
	StalenessModeratePenalty float64 `json:"staleness_moderate_penalty"` // Multiplier for 30+ min staleness
	ConsistencyPenalty       float64 `json:"consistency_penalty"`        // Multiplier for consistently low activity

	// Time-based thresholds
	StalenessHourThreshold     time.Duration `json:"staleness_hour_threshold"`     // Time for severe staleness
	StalenessModerateThreshold time.Duration `json:"staleness_moderate_threshold"` // Time for moderate staleness
	InactiveHourThreshold      time.Duration `json:"inactive_hour_threshold"`      // Time for inactive classification
	StaleHourThreshold         time.Duration `json:"stale_hour_threshold"`         // Time for stale classification

	// Pattern detection
	MinPatrolsForConsistency float64 `json:"min_patrols_for_consistency"` // Minimum patrols to detect patterns
	ConsistencyActivityLimit float64 `json:"consistency_activity_limit"`  // Activity level for consistency check
}

// PruningConfig controls synaptic pruning behavior
type PruningConfig struct {
	AgeThreshold     time.Duration `json:"age_threshold"`      // Minimum age before pruning
	ScoreThreshold   float64       `json:"score_threshold"`    // Minimum score for pruning (higher = more likely)
	ActivityWeight   float64       `json:"activity_weight"`    // Weight of activity in pruning score
	RedundancyWeight float64       `json:"redundancy_weight"`  // Weight of redundancy in pruning score
	MetabolicWeight  float64       `json:"metabolic_weight"`   // Weight of metabolic cost in pruning score
	BasePruningScore float64       `json:"base_pruning_score"` // Base score before adjustments
	MaxPruningScore  float64       `json:"max_pruning_score"`  // Maximum possible pruning score
}

// PatrolConfig controls microglial surveillance patterns
type PatrolConfig struct {
	DefaultPatrolRate    time.Duration `json:"default_patrol_rate"`    // Default time between patrols
	DefaultTerritorySize float64       `json:"default_territory_size"` // Default patrol radius
	MaxPatrolHistory     int           `json:"max_patrol_history"`     // Maximum patrol events to track
	HealthUpdateRate     float64       `json:"health_update_rate"`     // Default activity level during patrol
}

// ResourceConfig controls component creation and resource management
type ResourceConfig struct {
	MaxComponents       int           `json:"max_components"`        // Maximum components allowed
	HighPriorityBypass  bool          `json:"high_priority_bypass"`  // Allow high priority to exceed limits
	ResourceCheckWindow time.Duration `json:"resource_check_window"` // Time window for resource calculations
	DefaultComponentTTL time.Duration `json:"default_component_ttl"` // Default component time-to-live
}

// =================================================================================
// PRESET CONFIGURATIONS
// =================================================================================

// GetDefaultMicrogliaConfig returns biologically realistic default configuration
func GetDefaultMicrogliaConfig() MicrogliaConfig {
	return MicrogliaConfig{
		HealthThresholds: HealthScoringConfig{
			// Activity thresholds based on typical neural firing rates
			CriticalActivityThreshold: 0.02, // 2% - critically low
			VeryLowActivityThreshold:  0.05, // 5% - very low
			LowActivityThreshold:      0.15, // 15% - low but detectable
			ModerateActivityThreshold: 0.30, // 30% - moderate

			// Connection thresholds based on typical neural connectivity
			IsolatedConnectionThreshold: 0, // Completely isolated
			PoorConnectionThreshold:     3, // Poorly connected
			MinHealthyConnections:       5, // Minimum for healthy function

			// Biologically realistic penalties - not too harsh for functional neurons
			CriticalActivityPenalty:  0.4,  // 60% penalty for critical activity
			LowActivityPenalty:       0.6,  // 40% penalty for low activity
			ModerateActivityPenalty:  0.75, // 25% penalty for moderate activity (reasonable for 30% firing)
			PoorConnectionPenalty:    0.7,  // 30% penalty for poor connections
			FewConnectionPenalty:     0.85, // 15% penalty for few connections
			StalenessHourPenalty:     0.5,  // 50% penalty for staleness
			StalenessModeratePenalty: 0.8,  // 20% penalty for moderate staleness
			ConsistencyPenalty:       0.6,  // 40% penalty for consistent inactivity

			// Time thresholds based on biological timescales
			StalenessHourThreshold:     1 * time.Hour,    // 1 hour for severe staleness
			StalenessModerateThreshold: 30 * time.Minute, // 30 minutes for moderate staleness
			InactiveHourThreshold:      2 * time.Hour,    // 2 hours for inactive
			StaleHourThreshold:         6 * time.Hour,    // 6 hours for stale

			// Pattern detection parameters
			MinPatrolsForConsistency: 5,    // Need 5+ patrols to detect patterns
			ConsistencyActivityLimit: 0.15, // Below 15% is consistently low
		},

		PruningSettings: PruningConfig{
			AgeThreshold:     24 * time.Hour, // 24 hours minimum age (biological realism)
			ScoreThreshold:   0.8,            // High threshold for actual pruning
			ActivityWeight:   0.7,            // Primary factor in pruning decisions
			RedundancyWeight: 0.2,            // Secondary factor
			MetabolicWeight:  0.1,            // Tertiary factor
			BasePruningScore: 0.0,            // Start from zero
			MaxPruningScore:  1.0,            // Maximum possible score
		},

		PatrolSettings: PatrolConfig{
			DefaultPatrolRate:    100 * time.Millisecond, // 10 Hz patrol rate
			DefaultTerritorySize: 50.0,                   // 50μm radius territory
			MaxPatrolHistory:     1000,                   // Track last 1000 patrols
			HealthUpdateRate:     0.5,                    // Moderate activity during patrol
		},

		ResourceLimits: ResourceConfig{
			MaxComponents:       1000,            // Default component limit
			HighPriorityBypass:  true,            // Allow emergency overrides
			ResourceCheckWindow: 1 * time.Minute, // Check resources every minute
			DefaultComponentTTL: 24 * time.Hour,  // 24 hour default lifetime
		},
	}
}

// GetConservativeMicrogliaConfig returns configuration with conservative thresholds
func GetConservativeMicrogliaConfig() MicrogliaConfig {
	config := GetDefaultMicrogliaConfig()

	// More lenient health scoring
	config.HealthThresholds.CriticalActivityThreshold = 0.01
	config.HealthThresholds.VeryLowActivityThreshold = 0.03
	config.HealthThresholds.LowActivityThreshold = 0.10
	config.HealthThresholds.CriticalActivityPenalty = 0.7 // Less severe penalties
	config.HealthThresholds.LowActivityPenalty = 0.8

	// Slower pruning
	config.PruningSettings.AgeThreshold = 72 * time.Hour // 3 days
	config.PruningSettings.ScoreThreshold = 0.9          // Very high threshold

	// Less frequent patrols
	config.PatrolSettings.DefaultPatrolRate = 200 * time.Millisecond

	return config
}

// GetAggressiveMicrogliaConfig returns configuration with aggressive maintenance
func GetAggressiveMicrogliaConfig() MicrogliaConfig {
	config := GetDefaultMicrogliaConfig()

	// Stricter health scoring
	config.HealthThresholds.CriticalActivityThreshold = 0.05
	config.HealthThresholds.VeryLowActivityThreshold = 0.10
	config.HealthThresholds.LowActivityThreshold = 0.25
	config.HealthThresholds.CriticalActivityPenalty = 0.3 // More severe penalties
	config.HealthThresholds.LowActivityPenalty = 0.5

	// FIXED: More aggressive pruning configuration
	config.PruningSettings.AgeThreshold = 6 * time.Hour // 6 hours
	config.PruningSettings.ScoreThreshold = 0.6         // Lower threshold (easier to prune)

	// FIXED: Increased weight factors for more aggressive pruning
	config.PruningSettings.ActivityWeight = 0.8   // Higher emphasis on activity (vs 0.7 default)
	config.PruningSettings.RedundancyWeight = 0.3 // Higher emphasis on redundancy (vs 0.2 default)
	config.PruningSettings.MetabolicWeight = 0.2  // Higher emphasis on metabolic cost (vs 0.1 default)

	// More frequent patrols
	config.PatrolSettings.DefaultPatrolRate = 50 * time.Millisecond

	return config
}

// =================================================================================
// CORE MICROGLIA STRUCTURE (Updated)
// =================================================================================

// Microglia handles component lifecycle management and neural maintenance
type Microglia struct {
	// === CONFIGURATION ===
	config MicrogliaConfig // All configurable parameters

	// === PRE-COMPUTED VALUES ===
	// TODO: Add central logging for config validation warnings
	// When extreme/unrealistic values are detected, log warnings but allow execution
	precomputed precomputedValues // Derived values for performance

	// === COMPONENT INTEGRATION ===
	astrocyteNetwork *AstrocyteNetwork // Component tracking and connectivity

	// === LIFECYCLE TRACKING ===
	birthRequests  []ComponentBirthRequest // Pending component creation requests
	deathQueue     []ComponentDeathRequest // Components scheduled for removal
	pruningTargets map[string]PruningInfo  // Synapses/connections marked for pruning

	// === HEALTH MONITORING ===
	healthStatus map[string]ComponentHealth // Component health tracking
	patrolRoutes []PatrolRoute              // Microglial patrol patterns
	lastPatrol   map[string]time.Time       // Last patrol time per region

	// === MAINTENANCE STATISTICS ===
	maintenanceStats MicroglialStats

	// === STATE MANAGEMENT ===
	isActive bool
	mu       sync.RWMutex
}

// precomputedValues stores derived configuration values for performance
type precomputedValues struct {
	// Health scoring lookup values
	activityRanges []activityRange

	// Pruning calculation values
	maxActivityScore float64

	// Time comparisons (pre-converted to nanoseconds for faster comparison)
	stalenessHourNanos     int64
	stalenessModerateNanos int64
	inactiveHourNanos      int64
	staleHourNanos         int64
}

type activityRange struct {
	threshold float64
	penalty   float64
	issueName string
}

// =================================================================================
// EXISTING TYPES (Unchanged)
// =================================================================================

// ComponentBirthRequest represents a request for new component creation
type ComponentBirthRequest struct {
	RequestID     string                 `json:"request_id"`
	ComponentType ComponentType          `json:"component_type"`
	Position      Position3D             `json:"position"`
	Justification string                 `json:"justification"` // Why this component is needed
	Priority      BirthPriority          `json:"priority"`
	RequestedAt   time.Time              `json:"requested_at"`
	RequestedBy   string                 `json:"requested_by"` // Source component ID
	Metadata      map[string]interface{} `json:"metadata"`
}

// ComponentDeathRequest represents a request for component removal
type ComponentDeathRequest struct {
	ComponentID string        `json:"component_id"`
	Reason      DeathReason   `json:"reason"`
	Severity    int           `json:"severity"` // 1-10 urgency scale
	RequestedAt time.Time     `json:"requested_at"`
	RequestedBy string        `json:"requested_by"` // Source of death request
	GracePeriod time.Duration `json:"grace_period"` // Time before execution
}

// PruningInfo tracks synapses/connections marked for removal
type PruningInfo struct {
	ConnectionID  string    `json:"connection_id"`
	SourceID      string    `json:"source_id"`
	TargetID      string    `json:"target_id"`
	ActivityLevel float64   `json:"activity_level"` // Recent activity (0.0-1.0)
	LastUsed      time.Time `json:"last_used"`
	MarkedAt      time.Time `json:"marked_at"`
	PruningScore  float64   `json:"pruning_score"` // Higher = more likely to prune
}

// ComponentHealth tracks the health status of network components
type ComponentHealth struct {
	ComponentID     string    `json:"component_id"`
	HealthScore     float64   `json:"health_score"`   // 0.0-1.0 (1.0 = perfect health)
	ActivityLevel   float64   `json:"activity_level"` // Recent activity rate
	ConnectionCount int       `json:"connection_count"`
	LastSeen        time.Time `json:"last_seen"`
	Issues          []string  `json:"issues"`       // Health problems detected
	PatrolCount     int64     `json:"patrol_count"` // Times checked by microglia
}

// PatrolRoute defines microglial surveillance patterns
type PatrolRoute struct {
	MicrogliaID       string        `json:"microglia_id"`
	Territory         Territory     `json:"territory"`   // Spatial patrol area
	PatrolRate        time.Duration `json:"patrol_rate"` // How often to patrol
	LastPatrol        time.Time     `json:"last_patrol"`
	ComponentsChecked int64         `json:"components_checked"`
}

// PatrolReport summarizes patrol findings
type PatrolReport struct {
	MicrogliaID       string    `json:"microglia_id"`
	PatrolTime        time.Time `json:"patrol_time"`
	ComponentsChecked int       `json:"components_checked"`
	IssuesFound       []string  `json:"issues_found"`
	HealthProblems    int       `json:"health_problems"`
	PruningCandidates int       `json:"pruning_candidates"`
}

// MicroglialStats tracks maintenance activity statistics
type MicroglialStats struct {
	ComponentsCreated  int64     `json:"components_created"`
	ComponentsRemoved  int64     `json:"components_removed"`
	ConnectionsPruned  int64     `json:"connections_pruned"`
	HealthChecks       int64     `json:"health_checks"`
	PatrolsCompleted   int64     `json:"patrols_completed"`
	LastResetTime      time.Time `json:"last_reset_time"`
	AverageHealthScore float64   `json:"average_health_score"`
}

// BirthPriority defines urgency levels for component creation
type BirthPriority int

const (
	PriorityLow       BirthPriority = iota // Normal developmental needs
	PriorityMedium                         // Activity-dependent demand
	PriorityHigh                           // Network bottleneck detected
	PriorityEmergency                      // Critical network failure
)

// DeathReason defines why a component is being removed
type DeathReason int

const (
	DeathInactivity   DeathReason = iota // Low activity/usage
	DeathDamage                          // Component damage detected
	DeathRedundancy                      // Unnecessary duplication
	DeathOptimization                    // Network optimization
	DeathResource                        // Resource constraints
	DeathRequest                         // External removal request
)

// =================================================================================
// CONSTRUCTOR AND CONFIGURATION
// =================================================================================

// NewMicroglia creates a microglia system with default configuration
func NewMicroglia(astrocyteNetwork *AstrocyteNetwork, maxComponents int) *Microglia {
	config := GetDefaultMicrogliaConfig()

	// Override max components if specified
	if maxComponents > 0 {
		config.ResourceLimits.MaxComponents = maxComponents
	}

	return NewMicrogliaWithConfig(astrocyteNetwork, config)
}

// NewMicrogliaWithConfig creates a microglia system with custom configuration
func NewMicrogliaWithConfig(astrocyteNetwork *AstrocyteNetwork, config MicrogliaConfig) *Microglia {
	mg := &Microglia{
		config:           config,
		astrocyteNetwork: astrocyteNetwork,
		birthRequests:    make([]ComponentBirthRequest, 0),
		deathQueue:       make([]ComponentDeathRequest, 0),
		pruningTargets:   make(map[string]PruningInfo),
		healthStatus:     make(map[string]ComponentHealth),
		patrolRoutes:     make([]PatrolRoute, 0),
		lastPatrol:       make(map[string]time.Time),
		maintenanceStats: MicroglialStats{LastResetTime: time.Now()},
		isActive:         false,
	}

	// Pre-compute derived values for performance
	mg.precomputed = mg.computeDerivedValues()

	return mg
}

// computeDerivedValues pre-computes configuration-derived values for performance
func (mg *Microglia) computeDerivedValues() precomputedValues {
	cfg := mg.config.HealthThresholds

	// Build activity ranges for efficient health scoring
	ranges := []activityRange{
		{cfg.CriticalActivityThreshold, cfg.CriticalActivityPenalty, "critically_low_activity"},
		{cfg.VeryLowActivityThreshold, cfg.LowActivityPenalty, "very_low_activity"},
		{cfg.LowActivityThreshold, cfg.LowActivityPenalty, "low_activity"},
		{cfg.ModerateActivityThreshold, cfg.ModerateActivityPenalty, "moderate_low_activity"},
	}

	// Pre-convert time durations to nanoseconds for faster comparison
	return precomputedValues{
		activityRanges:         ranges,
		maxActivityScore:       mg.config.PruningSettings.ActivityWeight,
		stalenessHourNanos:     cfg.StalenessHourThreshold.Nanoseconds(),
		stalenessModerateNanos: cfg.StalenessModerateThreshold.Nanoseconds(),
		inactiveHourNanos:      cfg.InactiveHourThreshold.Nanoseconds(),
		staleHourNanos:         cfg.StaleHourThreshold.Nanoseconds(),
	}
}

// UpdateConfig updates the microglia configuration and recomputes derived values
func (mg *Microglia) UpdateConfig(config MicrogliaConfig) {
	mg.mu.Lock()
	defer mg.mu.Unlock()

	mg.config = config
	mg.precomputed = mg.computeDerivedValues()
}

// GetConfig returns a copy of the current configuration
func (mg *Microglia) GetConfig() MicrogliaConfig {
	mg.mu.RLock()
	defer mg.mu.RUnlock()

	return mg.config // Return copy
}

// =================================================================================
// COMPONENT LIFECYCLE MANAGEMENT (Updated with configuration)
// =================================================================================

// CreateComponent handles component creation with biological validation
func (mg *Microglia) CreateComponent(info ComponentInfo) error {
	mg.mu.Lock()
	defer mg.mu.Unlock()

	return mg.createComponentUnsafe(info)
}

// RemoveComponent handles component removal with cleanup
func (mg *Microglia) RemoveComponent(id string) error {
	mg.mu.Lock()
	defer mg.mu.Unlock()

	// Remove from astrocyte network (handles connection cleanup)
	err := mg.astrocyteNetwork.Unregister(id)
	if err != nil {
		return err
	}

	// Clean up health monitoring
	delete(mg.healthStatus, id)

	// Remove from pruning targets
	for connID, pruning := range mg.pruningTargets {
		if pruning.SourceID == id || pruning.TargetID == id {
			delete(mg.pruningTargets, connID)
		}
	}

	// Update statistics
	mg.maintenanceStats.ComponentsRemoved++

	return nil
}

// =================================================================================
// BIRTH REQUEST MANAGEMENT (Updated with configuration)
// =================================================================================

// RequestComponentBirth submits a request for new component creation
func (mg *Microglia) RequestComponentBirth(request ComponentBirthRequest) error {
	mg.mu.Lock()
	defer mg.mu.Unlock()

	// Set request timestamp
	request.RequestedAt = time.Now()
	request.RequestID = fmt.Sprintf("birth_%d_%d", time.Now().UnixNano(), request.ComponentType)

	mg.birthRequests = append(mg.birthRequests, request)
	return nil
}

// ProcessBirthRequests evaluates and processes pending birth requests
// move func here

// createComponentUnsafe handles component creation assuming a lock is already held
func (mg *Microglia) createComponentUnsafe(info ComponentInfo) error {
	// Check resource constraints before creating component
	totalComponents := mg.astrocyteNetwork.Count()
	if totalComponents >= mg.config.ResourceLimits.MaxComponents {
		return fmt.Errorf("resource limit exceeded: cannot create component, already at maximum capacity (%d/%d)",
			totalComponents, mg.config.ResourceLimits.MaxComponents)
	}

	// Register with astrocyte network
	err := mg.astrocyteNetwork.Register(info)
	if err != nil {
		return err
	}

	// Initialize health monitoring
	mg.healthStatus[info.ID] = ComponentHealth{
		ComponentID:     info.ID,
		HealthScore:     1.0, // Perfect health initially
		ActivityLevel:   0.0, // No activity yet
		ConnectionCount: 0,
		LastSeen:        time.Now(),
		Issues:          make([]string, 0),
		PatrolCount:     0,
	}

	// Update statistics
	mg.maintenanceStats.ComponentsCreated++

	return nil
}

// FIXED: Add new method for emergency component creation that bypasses limits
func (mg *Microglia) createEmergencyComponentUnsafe(info ComponentInfo) error {
	// Emergency creation bypasses resource limits entirely

	// Register with astrocyte network
	err := mg.astrocyteNetwork.Register(info)
	if err != nil {
		return err
	}

	// Initialize health monitoring
	mg.healthStatus[info.ID] = ComponentHealth{
		ComponentID:     info.ID,
		HealthScore:     1.0, // Perfect health initially
		ActivityLevel:   0.0, // No activity yet
		ConnectionCount: 0,
		LastSeen:        time.Now(),
		Issues:          make([]string, 0),
		PatrolCount:     0,
	}

	// Update statistics
	mg.maintenanceStats.ComponentsCreated++

	return nil
}

// evaluateBirthNeed determines if a birth request should be fulfilled
func (mg *Microglia) evaluateBirthNeed(request ComponentBirthRequest) bool {
	// Always approve high priority requests if bypass is enabled
	if request.Priority >= PriorityHigh && mg.config.ResourceLimits.HighPriorityBypass {
		return true
	}

	// For lower priority, check resource availability
	totalComponents := mg.astrocyteNetwork.Count()
	if totalComponents < mg.config.ResourceLimits.MaxComponents {
		return true
	}

	return false
}

// FIXED: Update ProcessBirthRequests to use emergency creation for high priority
func (mg *Microglia) ProcessBirthRequests() []ComponentInfo {
	mg.mu.Lock()
	defer mg.mu.Unlock()

	createdComponents := make([]ComponentInfo, 0)

	// Process requests by priority
	for i := len(mg.birthRequests) - 1; i >= 0; i-- {
		request := mg.birthRequests[i]

		// Evaluate if birth is justified
		if mg.evaluateBirthNeed(request) {
			// Create the component
			newComponent := ComponentInfo{
				ID:           fmt.Sprintf("%d_%d", request.ComponentType, time.Now().UnixNano()),
				Type:         request.ComponentType,
				Position:     request.Position,
				State:        StateActive,
				Metadata:     request.Metadata,
				RegisteredAt: time.Now(),
			}

			var err error
			// FIXED: Use emergency creation for high priority requests
			if request.Priority >= PriorityHigh && mg.config.ResourceLimits.HighPriorityBypass {
				err = mg.createEmergencyComponentUnsafe(newComponent)
			} else {
				err = mg.createComponentUnsafe(newComponent)
			}

			if err == nil {
				createdComponents = append(createdComponents, newComponent)
			}

			// Remove processed request
			mg.birthRequests = append(mg.birthRequests[:i], mg.birthRequests[i+1:]...)
		}
	}

	return createdComponents
}

// =================================================================================
// SYNAPTIC PRUNING (Updated with configuration)
// =================================================================================

// MarkForPruning marks a connection for potential removal
func (mg *Microglia) MarkForPruning(connectionID, sourceID, targetID string, activityLevel float64) {
	mg.mu.Lock()
	defer mg.mu.Unlock()

	mg.markForPruningUnsafe(connectionID, sourceID, targetID, activityLevel)
}

// PROPER SOLUTION: Redesign the lock architecture
// The issue is that we're holding mg.mu when calling calculatePruningScore
// but the helper functions need to access different data structures

// METHOD: Pre-calculate scores without holding locks

// markForPruningUnsafe marks a connection for potential removal (assumes lock held)
func (mg *Microglia) markForPruningUnsafe(connectionID, sourceID, targetID string, activityLevel float64) {
	// FIXED: Calculate pruning score BEFORE acquiring locks
	// Release the lock temporarily to calculate the score
	mg.mu.Unlock()
	pruningScore := mg.calculatePruningScore(activityLevel, sourceID, targetID)
	mg.mu.Lock()

	pruningInfo := PruningInfo{
		ConnectionID:  connectionID,
		SourceID:      sourceID,
		TargetID:      targetID,
		ActivityLevel: activityLevel,
		LastUsed:      time.Now(),
		MarkedAt:      time.Now(),
		PruningScore:  pruningScore,
	}

	mg.pruningTargets[connectionID] = pruningInfo
}

// ExecutePruning removes connections marked for pruning
func (mg *Microglia) ExecutePruning() []string {
	mg.mu.Lock()
	defer mg.mu.Unlock()

	return mg.executePruningUnsafe()
}

// executePruningUnsafe removes connections marked for pruning (assumes lock held)
func (mg *Microglia) executePruningUnsafe() []string {
	prunedConnections := make([]string, 0)

	for connID, pruning := range mg.pruningTargets {
		// Check if connection should be pruned
		if mg.shouldPruneUnsafe(pruning) {
			// Remove the connection (implementation depends on connection type)
			// For now, just remove from tracking
			delete(mg.pruningTargets, connID)
			prunedConnections = append(prunedConnections, connID)
			mg.maintenanceStats.ConnectionsPruned++
		}
	}

	return prunedConnections
}

// =================================================================================
// HEALTH MONITORING (Updated with configuration)
// =================================================================================

// UpdateComponentHealth updates health status based on activity
func (mg *Microglia) UpdateComponentHealth(componentID string, activityLevel float64, connectionCount int) {
	mg.mu.Lock()
	defer mg.mu.Unlock()

	mg.updateComponentHealthUnsafe(componentID, activityLevel, connectionCount)
}

// updateComponentHealthUnsafe updates health status (assumes lock held)
func (mg *Microglia) updateComponentHealthUnsafe(componentID string, activityLevel float64, connectionCount int) {
	health, exists := mg.healthStatus[componentID]
	if !exists {
		health = ComponentHealth{
			ComponentID: componentID,
			Issues:      make([]string, 0),
		}
	}

	// Update health metrics
	health.ActivityLevel = activityLevel
	health.ConnectionCount = connectionCount
	health.LastSeen = time.Now()
	health.PatrolCount++

	// Calculate health score using configuration
	health.HealthScore = mg.calculateHealthScoreUnsafe(health)

	// Check for health issues using configuration
	health.Issues = mg.detectHealthIssuesUnsafe(health)

	mg.healthStatus[componentID] = health
	mg.maintenanceStats.HealthChecks++
}

// GetComponentHealth returns current health status
func (mg *Microglia) GetComponentHealth(componentID string) (ComponentHealth, bool) {
	mg.mu.RLock()
	defer mg.mu.RUnlock()

	health, exists := mg.healthStatus[componentID]
	return health, exists
}

// =================================================================================
// PATROL AND SURVEILLANCE (Updated with configuration)
// =================================================================================

// EstablishPatrolRoute creates a microglial surveillance pattern
func (mg *Microglia) EstablishPatrolRoute(microgliaID string, territory Territory, patrolRate time.Duration) {
	mg.mu.Lock()
	defer mg.mu.Unlock()

	route := PatrolRoute{
		MicrogliaID:       microgliaID,
		Territory:         territory,
		PatrolRate:        patrolRate,
		LastPatrol:        time.Now(),
		ComponentsChecked: 0,
	}

	mg.patrolRoutes = append(mg.patrolRoutes, route)
}

// ExecutePatrol performs surveillance of assigned territory
func (mg *Microglia) ExecutePatrol(microgliaID string) PatrolReport {
	mg.mu.Lock()
	defer mg.mu.Unlock()

	return mg.executePatrolUnsafe(microgliaID)
}

// executePatrolUnsafe performs surveillance (assumes lock held)
func (mg *Microglia) executePatrolUnsafe(microgliaID string) PatrolReport {
	report := PatrolReport{
		MicrogliaID:       microgliaID,
		PatrolTime:        time.Now(),
		ComponentsChecked: 0,
		IssuesFound:       make([]string, 0),
	}

	// Find patrol route for this microglia
	for i, route := range mg.patrolRoutes {
		if route.MicrogliaID == microgliaID {
			// FIXED: Release lock before calling astrocyte network to prevent deadlock
			mg.mu.Unlock()
			components := mg.astrocyteNetwork.FindNearby(route.Territory.Center, route.Territory.Radius)
			mg.mu.Lock()

			for _, comp := range components {
				// FIXED: Get connections without holding microglia lock
				mg.mu.Unlock()
				connections := mg.astrocyteNetwork.GetConnections(comp.ID)
				mg.mu.Lock()

				// Update health based on component status using configured rate
				mg.updateComponentHealthUnsafe(comp.ID, mg.config.PatrolSettings.HealthUpdateRate, len(connections))
				report.ComponentsChecked++
			}

			// Update patrol route
			mg.patrolRoutes[i].LastPatrol = time.Now()
			mg.patrolRoutes[i].ComponentsChecked += int64(len(components))
			mg.maintenanceStats.PatrolsCompleted++
			break
		}
	}

	return report
}

// =================================================================================
// STATISTICS AND MONITORING (Unchanged)
// =================================================================================

// GetMaintenanceStats returns microglial activity statistics
func (mg *Microglia) GetMaintenanceStats() MicroglialStats {
	mg.mu.RLock()
	defer mg.mu.RUnlock()

	// Calculate average health score
	totalHealth := 0.0
	healthCount := 0
	for _, health := range mg.healthStatus {
		totalHealth += health.HealthScore
		healthCount++
	}

	stats := mg.maintenanceStats
	if healthCount > 0 {
		stats.AverageHealthScore = totalHealth / float64(healthCount)
	}

	return stats
}

// GetPruningCandidates returns connections marked for pruning
func (mg *Microglia) GetPruningCandidates() []PruningInfo {
	mg.mu.RLock()
	defer mg.mu.RUnlock()

	candidates := make([]PruningInfo, 0, len(mg.pruningTargets))
	for _, pruning := range mg.pruningTargets {
		candidates = append(candidates, pruning)
	}

	return candidates
}

// =================================================================================
// INTERNAL UTILITY FUNCTIONS (Updated with configuration)
// =================================================================================

// calculateRedundancyScore determines if this connection is redundant
// Higher redundancy = higher pruning score (more likely to prune)
// This function manages its own locking and does not assume any locks are held
// Fix calculateRedundancyScore to handle empty IDs
func (mg *Microglia) calculateRedundancyScore(sourceID, targetID string) float64 {
	// FIXED: Handle empty or invalid IDs
	if sourceID == "" || targetID == "" {
		return 0.5 // Default moderate redundancy for invalid IDs
	}

	// Get component info from astrocyte network
	sourceInfo, sourceExists := mg.astrocyteNetwork.Get(sourceID)
	if !sourceExists {
		return 0.5 // Default moderate redundancy if component doesn't exist
	}

	targetInfo, targetExists := mg.astrocyteNetwork.Get(targetID)
	if !targetExists {
		return 0.5 // Default moderate redundancy if component doesn't exist
	}

	// Calculate redundancy based on multiple factors
	redundancyScore := 0.0

	// Factor 1: Connection count redundancy
	sourceConnections := mg.astrocyteNetwork.GetConnections(sourceID)
	targetConnections := mg.astrocyteNetwork.GetConnections(targetID)

	sourceConnectionFactor := math.Min(float64(len(sourceConnections))/10.0, 1.0)
	targetConnectionFactor := math.Min(float64(len(targetConnections))/10.0, 1.0)
	redundancyScore += (sourceConnectionFactor + targetConnectionFactor) * 0.3

	// Factor 2: Spatial redundancy
	spatialRedundancy := mg.calculateSpatialRedundancy(sourceInfo.Position, targetInfo.Position)
	redundancyScore += spatialRedundancy * 0.4

	// Factor 3: Functional redundancy (with proper locking)
	mg.mu.RLock()
	sourceHealth, sourceExists := mg.healthStatus[sourceID]
	targetHealth, targetExists := mg.healthStatus[targetID]
	mg.mu.RUnlock()

	var functionalRedundancy float64
	if sourceExists && targetExists {
		healthDifference := math.Abs(sourceHealth.HealthScore - targetHealth.HealthScore)
		if healthDifference < 0.2 {
			functionalRedundancy = 0.7
		} else if healthDifference < 0.5 {
			functionalRedundancy = 0.4
		} else {
			functionalRedundancy = 0.1
		}
	} else {
		functionalRedundancy = 0.5 // Default moderate redundancy
	}
	redundancyScore += functionalRedundancy * 0.3

	// FIXED: Ensure score is valid and bounded
	if math.IsNaN(redundancyScore) || math.IsInf(redundancyScore, 0) {
		redundancyScore = 0.5 // Default moderate redundancy
	}

	// Ensure score is between 0.0 and 1.0
	if redundancyScore < 0.0 {
		redundancyScore = 0.0
	}
	if redundancyScore > 1.0 {
		redundancyScore = 1.0
	}

	return redundancyScore
}

// Fix calculateMetabolicCost to handle empty IDs and invalid activity
func (mg *Microglia) calculateMetabolicCost(sourceID, targetID string, activityLevel float64) float64 {
	// FIXED: Handle empty or invalid IDs
	if sourceID == "" || targetID == "" {
		return 0.5 // Default moderate cost for invalid IDs
	}

	// Get component positions
	sourceInfo, sourceExists := mg.astrocyteNetwork.Get(sourceID)
	targetInfo, targetExists := mg.astrocyteNetwork.Get(targetID)

	if !sourceExists || !targetExists {
		return 0.5 // Default moderate cost if components don't exist
	}

	metabolicCost := 0.0

	// Factor 1: Distance cost
	dx := sourceInfo.Position.X - targetInfo.Position.X
	dy := sourceInfo.Position.Y - targetInfo.Position.Y
	dz := sourceInfo.Position.Z - targetInfo.Position.Z
	distance := math.Sqrt(dx*dx + dy*dy + dz*dz)

	var distanceCost float64
	if math.IsNaN(distance) || math.IsInf(distance, 0) {
		distanceCost = 1.0 // Maximum cost for invalid distances
	} else {
		distanceCost = math.Min(distance/100.0, 1.0)
	}
	metabolicCost += distanceCost * 0.4

	// Factor 2: Activity cost - FIXED to handle invalid activity
	var activityCost float64
	if math.IsNaN(activityLevel) || math.IsInf(activityLevel, 0) {
		activityCost = 0.8 // High cost for invalid activity
	} else if activityLevel < 0 {
		activityCost = 0.8 // Treat negative as very low activity
	} else if activityLevel > 1 {
		activityCost = 0.6 // Treat >1 as very high activity
	} else if activityLevel < 0.1 {
		activityCost = 0.8 // Low activity = maintenance cost without benefit
	} else if activityLevel > 0.8 {
		activityCost = 0.6 // High activity = high energy consumption
	} else {
		activityCost = 0.2 // Medium activity = optimal efficiency
	}
	metabolicCost += activityCost * 0.4

	// Factor 3: Health cost (with proper locking)
	mg.mu.RLock()
	sourceHealth, sourceExists := mg.healthStatus[sourceID]
	targetHealth, targetExists := mg.healthStatus[targetID]
	mg.mu.RUnlock()

	var healthCost float64
	if sourceExists && targetExists {
		avgHealth := (sourceHealth.HealthScore + targetHealth.HealthScore) / 2.0
		healthCost = 1.0 - avgHealth // Lower health = higher cost
	} else {
		healthCost = 0.5 // Default moderate cost
	}
	metabolicCost += healthCost * 0.2

	// FIXED: Ensure cost is valid and bounded
	if math.IsNaN(metabolicCost) || math.IsInf(metabolicCost, 0) {
		metabolicCost = 0.5 // Default moderate cost
	}

	// Ensure cost is between 0.0 and 1.0
	if metabolicCost < 0.0 {
		metabolicCost = 0.0
	}
	if metabolicCost > 1.0 {
		metabolicCost = 1.0
	}

	return metabolicCost
}

// calculateSpatialRedundancy assesses redundancy based on spatial proximity
// Pure calculation function - no locks needed
func (mg *Microglia) calculateSpatialRedundancy(sourcePos, targetPos Position3D) float64 {
	// Calculate distance between components
	dx := sourcePos.X - targetPos.X
	dy := sourcePos.Y - targetPos.Y
	dz := sourcePos.Z - targetPos.Z
	distance := math.Sqrt(dx*dx + dy*dy + dz*dz)

	// Handle edge cases gracefully
	if math.IsNaN(distance) || math.IsInf(distance, 0) {
		return 0.5 // Default moderate redundancy for invalid distances
	}

	// Shorter connections are more likely to be redundant
	if distance < 10.0 {
		return 0.8 // High redundancy for very close connections
	} else if distance < 50.0 {
		return 0.5 // Medium redundancy for medium distance
	} else {
		return 0.1 // Low redundancy for long-range connections
	}
}

// calculateFunctionalRedundancy assesses redundancy based on component health/function
func (mg *Microglia) calculateFunctionalRedundancy(sourceID, targetID string) float64 {
	sourceHealth, sourceExists := mg.healthStatus[sourceID]
	targetHealth, targetExists := mg.healthStatus[targetID]

	if !sourceExists || !targetExists {
		return 0.5 // Default moderate redundancy if health unknown
	}

	// If both components are in similar health states, connection might be redundant
	healthDifference := math.Abs(sourceHealth.HealthScore - targetHealth.HealthScore)

	// Similar health = higher redundancy (both healthy or both unhealthy)
	if healthDifference < 0.2 {
		return 0.7 // High functional redundancy
	} else if healthDifference < 0.5 {
		return 0.4 // Medium functional redundancy
	} else {
		return 0.1 // Low functional redundancy (complementary function)
	}
}

// calculatePruningScore determines how likely a connection is to be pruned
// This version does NOT assume any locks are held
func (mg *Microglia) calculatePruningScore(activityLevel float64, sourceID, targetID string) float64 {
	// Read configuration without holding the main lock
	mg.mu.RLock()
	cfg := mg.config.PruningSettings
	mg.mu.RUnlock()

	// Start with base score
	score := cfg.BasePruningScore

	// FIXED: Handle invalid activity level gracefully
	normalizedActivity := activityLevel
	if math.IsNaN(activityLevel) || math.IsInf(activityLevel, 0) {
		normalizedActivity = 0.0 // Treat invalid activity as zero (high pruning score)
	} else if activityLevel < 0 {
		normalizedActivity = 0.0 // Clamp negative values
	} else if activityLevel > 1 {
		normalizedActivity = 1.0 // Clamp values > 1
	}

	// Activity component (higher activity = lower pruning score)
	activityScore := (1.0 - normalizedActivity) * cfg.ActivityWeight
	score += activityScore

	// FIXED: Handle invalid IDs gracefully
	var redundancyScore, metabolicScore float64
	if sourceID == "" || targetID == "" {
		// For invalid IDs, use moderate default scores
		redundancyScore = 0.5 * cfg.RedundancyWeight // Moderate redundancy
		metabolicScore = 0.5 * cfg.MetabolicWeight   // Moderate metabolic cost
	} else {
		// Calculate normally for valid IDs
		redundancyScore = mg.calculateRedundancyScore(sourceID, targetID) * cfg.RedundancyWeight
		metabolicScore = mg.calculateMetabolicCost(sourceID, targetID, normalizedActivity) * cfg.MetabolicWeight
	}

	score += redundancyScore
	score += metabolicScore

	// FIXED: Ensure score is always valid and within bounds
	if math.IsNaN(score) || math.IsInf(score, 0) {
		score = 0.5 // Default moderate pruning score for invalid calculations
	}

	// Clamp to valid range
	if score < 0.0 {
		score = 0.0
	}
	if score > cfg.MaxPruningScore {
		score = cfg.MaxPruningScore
	}

	return score
}

// shouldPrune determines if a connection should be removed
func (mg *Microglia) shouldPrune(pruning PruningInfo) bool {
	mg.mu.RLock()
	defer mg.mu.RUnlock()
	return mg.shouldPruneUnsafe(pruning)
}

// shouldPruneUnsafe determines if connection should be removed (assumes lock held)
func (mg *Microglia) shouldPruneUnsafe(pruning PruningInfo) bool {
	cfg := mg.config.PruningSettings

	// Check age requirement
	age := time.Since(pruning.MarkedAt)
	if age < cfg.AgeThreshold {
		return false
	}

	// Check pruning score threshold
	return pruning.PruningScore > cfg.ScoreThreshold
}

// calculateHealthScore computes overall component health using configuration
func (mg *Microglia) calculateHealthScore(health ComponentHealth) float64 {
	mg.mu.RLock()
	defer mg.mu.RUnlock()
	return mg.calculateHealthScoreUnsafe(health)
}

// calculateHealthScoreUnsafe computes health score (assumes lock held)
func (mg *Microglia) calculateHealthScoreUnsafe(health ComponentHealth) float64 {
	cfg := mg.config.HealthThresholds
	score := 1.0

	// FIXED: Apply activity-based penalties with correct biological thresholds
	// Use <= for boundary conditions to ensure 0.30 gets moderate penalty
	if health.ActivityLevel < cfg.CriticalActivityThreshold {
		score *= cfg.CriticalActivityPenalty
	} else if health.ActivityLevel < cfg.VeryLowActivityThreshold {
		score *= cfg.LowActivityPenalty
	} else if health.ActivityLevel < cfg.LowActivityThreshold {
		score *= cfg.LowActivityPenalty
	} else if health.ActivityLevel <= cfg.ModerateActivityThreshold {
		// FIXED: Use <= so that exactly 0.30 gets the moderate penalty
		score *= cfg.ModerateActivityPenalty
	} else if health.ActivityLevel < 0.70 {
		// Penalty for activity levels 0.30-0.70 (good but not excellent)
		score *= 0.9 // 10% penalty for good but not excellent activity
	}
	// Activity >= 0.70 gets no penalty (excellent range)

	// Apply connection-based penalties
	if health.ConnectionCount == cfg.IsolatedConnectionThreshold {
		score *= cfg.PoorConnectionPenalty
	} else if health.ConnectionCount < cfg.PoorConnectionThreshold {
		score *= cfg.PoorConnectionPenalty
	} else if health.ConnectionCount < cfg.MinHealthyConnections {
		score *= cfg.FewConnectionPenalty
	}

	// Apply time-based penalties using pre-computed nanosecond values
	timeSinceLastSeen := time.Since(health.LastSeen)
	timeNanos := timeSinceLastSeen.Nanoseconds()

	if timeNanos > mg.precomputed.stalenessHourNanos {
		score *= cfg.StalenessHourPenalty
	} else if timeNanos > mg.precomputed.stalenessModerateNanos {
		score *= cfg.StalenessModeratePenalty
	}

	// Apply consistency penalty
	if health.PatrolCount > int64(cfg.MinPatrolsForConsistency) {
		avgActivityPerCheck := health.ActivityLevel / float64(health.PatrolCount)
		if avgActivityPerCheck < cfg.ConsistencyActivityLimit {
			score *= cfg.ConsistencyPenalty
		}
	}

	// Ensure score stays in valid range
	if score < 0.0 {
		score = 0.0
	}
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// detectHealthIssues identifies potential component problems using configuration
func (mg *Microglia) detectHealthIssues(health ComponentHealth) []string {
	mg.mu.RLock()
	defer mg.mu.RUnlock()
	return mg.detectHealthIssuesUnsafe(health)
}

// detectHealthIssuesUnsafe identifies component problems (assumes lock held)
func (mg *Microglia) detectHealthIssuesUnsafe(health ComponentHealth) []string {
	cfg := mg.config.HealthThresholds
	issues := make([]string, 0)

	// Check activity levels against configured thresholds
	if health.ActivityLevel < cfg.CriticalActivityThreshold {
		issues = append(issues, "critically_low_activity")
		issues = append(issues, "very_low_activity") // Also add parent category
	} else if health.ActivityLevel < cfg.VeryLowActivityThreshold {
		issues = append(issues, "very_low_activity")
	} else if health.ActivityLevel < cfg.LowActivityThreshold {
		issues = append(issues, "low_activity")
		issues = append(issues, "very_low_activity") // Add for comprehensive detection
	} else if health.ActivityLevel < cfg.ModerateActivityThreshold {
		issues = append(issues, "moderate_low_activity")
	}

	// Check connection levels
	if health.ConnectionCount == cfg.IsolatedConnectionThreshold {
		issues = append(issues, "isolated_component")
	} else if health.ConnectionCount < cfg.PoorConnectionThreshold {
		issues = append(issues, "poorly_connected")
	}

	// Check staleness using pre-computed nanosecond values for performance
	timeSinceLastSeen := time.Since(health.LastSeen)
	timeNanos := timeSinceLastSeen.Nanoseconds()

	if timeNanos > mg.precomputed.staleHourNanos {
		issues = append(issues, "stale_component")
	} else if timeNanos > mg.precomputed.inactiveHourNanos {
		issues = append(issues, "inactive_component")
	}

	// Check for persistent inactivity patterns
	if health.PatrolCount > int64(cfg.MinPatrolsForConsistency) &&
		health.ActivityLevel < cfg.ConsistencyActivityLimit {
		issues = append(issues, "persistently_inactive")
	}

	// Ensure activity issues are always detected for low activity (compatibility)
	if health.ActivityLevel <= 0.1 {
		hasActivityIssue := false
		for _, issue := range issues {
			if issue == "very_low_activity" || issue == "critically_low_activity" || issue == "low_activity" {
				hasActivityIssue = true
				break
			}
		}
		if !hasActivityIssue {
			issues = append(issues, "very_low_activity")
		}
	}

	return issues
}
