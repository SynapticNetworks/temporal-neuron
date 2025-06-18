/*
=================================================================================
SYNAPSE FACTORY - CENTRALIZED SYNAPSE CREATION
=================================================================================

This file implements the factory pattern for creating different types of
biologically realistic synapses. It provides a single entry point for constructing
synapses from a SynapseConfig, decoupling the creation logic from the client code
(e.g., the matrix or network builder).

DESIGN PRINCIPLES:
1. CENTRALIZED CREATION: A single `CreateSynapse` function serves as the entry point.
2. EXTENSIBILITY: New synapse types can be registered at runtime, allowing for
   the system to be extended without modifying the core factory code.
3. DECOUPLING: The client only needs to know about `SynapseConfig` and the
   `SynapticProcessor` interface, not the concrete implementations.
4. CONFIGURATION-DRIVEN: The behavior and components of the created synapse are
   entirely determined by the provided `SynapseConfig`.

WORKFLOW:
1. The client (e.g., Matrix) creates a `SynapseConfig` using a preset from `config.go`
   or by defining a custom configuration.
2. The client calls `CreateSynapse(config)`.
3. The factory looks up the `SynapseType` in its registry to find the appropriate
   constructor function.
4. The constructor function is called, which assembles a new `Synapse` with
   all its necessary sub-components (VesicleDynamics, ActivityMonitor, etc.).
5. The fully constructed synapse, conforming to the `SynapticProcessor` interface,
   is returned to the client.
=================================================================================
*/

package synapse

import (
	"fmt"
	"sync"
	"time"
)

var (
	// factoryRegistry holds the mapping from synapse type strings to their
	// respective factory functions.
	factoryRegistry = make(map[string]SynapseFactory)
	// registryMutex protects the factoryRegistry from concurrent access during
	// registration, ensuring thread safety.
	registryMutex = &sync.RWMutex{}
)

// RegisterSynapseType adds a new synapse constructor to the factory registry.
// This function is the primary mechanism for extending the synapse system with new types.
// It is thread-safe.
func RegisterSynapseType(synapseType string, factory SynapseFactory) error {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	if _, exists := factoryRegistry[synapseType]; exists {
		return fmt.Errorf("synapse type '%s' is already registered", synapseType)
	}

	factoryRegistry[synapseType] = factory
	return nil
}

// CreateSynapse is the main entry point for creating any type of synapse.
// It looks up the synapse type from the config in the registry and uses the
// corresponding factory to construct the synapse.
//
// THIS FUNCTION SIGNATURE IS NOW CORRECTED to accept all necessary parameters.
func CreateSynapse(id string, config SynapseConfig, callbacks SynapseCallbacks) (SynapticProcessor, error) {
	// First, perform a comprehensive validation of the configuration.
	validationResult := ValidateConfig(config)
	if !validationResult.IsValid {
		return nil, fmt.Errorf("invalid synapse configuration for ID '%s': %v", config.SynapseID, validationResult.Errors)
	}

	registryMutex.RLock()
	factory, exists := factoryRegistry[config.SynapseType]
	registryMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unrecognized synapse type: '%s'. Ensure it has been registered", config.SynapseType)
	}

	// It now correctly passes all required arguments to the factory function.
	return factory(id, config, callbacks)
}

// init function automatically registers all the standard, predefined synapse types
// when the package is loaded.
// The type error is now resolved because newStandardSynapse matches SynapseFactory.
func init() {
	// Note: We ignore the error here because we know these types are not duplicates.
	_ = RegisterSynapseType("excitatory_glutamatergic", newStandardSynapse)     //
	_ = RegisterSynapseType("inhibitory_gabaergic", newStandardSynapse)         //
	_ = RegisterSynapseType("neuromodulatory_dopaminergic", newStandardSynapse) //
	_ = RegisterSynapseType("developmental_plastic", newStandardSynapse)
	_ = RegisterSynapseType("aged_reduced_plasticity", newStandardSynapse) //
}

// newStandardSynapse is the corrected generic factory function for creating the main synapse implementation.
// It now correctly matches the SynapseFactory signature from types.go.
func newStandardSynapse(id string, config SynapseConfig, callbacks SynapseCallbacks) (SynapticProcessor, error) {
	// 1. Create the core sub-components.
	activityMonitor := NewSynapticActivityMonitor(id)
	plasticityCalculator := NewPlasticityCalculator(config.STDPConfig)

	// 2. Create the vesicle dynamics system if it's enabled in the config.
	var vesicleSystem VesicleSystem
	if config.VesicleConfig.Enabled {
		vd := NewVesicleDynamics(config.VesicleConfig.MaxReleaseRate)

		// Further configure the vesicle dynamics from the VesicleConfig struct.
		vd.baseLinesReleaseProbability = config.VesicleConfig.BaselineProbability
		vd.readyPoolSize = config.VesicleConfig.ReadyPoolSize
		vd.recyclingPoolSize = config.VesicleConfig.RecyclingPoolSize
		vd.reservePoolSize = config.VesicleConfig.ReservePoolSize
		vd.fastRecyclingRate = float64(vd.readyPoolSize) / config.VesicleConfig.FastRecyclingTime.Seconds()
		vd.slowRecyclingRate = float64(vd.recyclingPoolSize) / config.VesicleConfig.SlowRecyclingTime.Seconds()
		vd.refillTime = config.VesicleConfig.RefillTime

		vesicleSystem = vd
	}

	// 3. Assemble the main synapse struct.
	synapse := &Synapse{
		id:                   id, // Use the ID passed into the factory
		config:               config,
		weight:               config.InitialWeight,
		delay:                config.BaseSynapticDelay,
		activityMonitor:      activityMonitor,
		plasticityCalculator: plasticityCalculator,
		vesicleSystem:        vesicleSystem,
		callbacks:            callbacks, // Use the callbacks passed into the factory
		state:                StateActive,
		lastTransmission:     time.Now(),
		lastPlasticityEvent:  time.Now(),
	}

	// Record the initial weight in the monitor's history.
	synapse.activityMonitor.RecordPlasticity(PlasticityEvent{
		SynapseID:    synapse.id,
		EventType:    PlasticityHebbian,
		Timestamp:    time.Now(),
		WeightBefore: 0,
		WeightAfter:  synapse.weight,
		WeightChange: synapse.weight,
		Context:      map[string]interface{}{"reason": "initial_weight_setting"},
	})

	return synapse, nil
}

// GetRegisteredTypes returns a slice of all registered synapse type names.
func GetRegisteredTypes() []string {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	types := make([]string, 0, len(factoryRegistry))
	for t := range factoryRegistry {
		types = append(types, t)
	}
	return types
}
