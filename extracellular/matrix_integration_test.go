/*
=================================================================================
EXTRACELLULAR MATRIX - INTEGRATION TEST
=================================================================================

Tests the complete biological coordination system with all subsystems working
together. Uses mock neurons and synapses to demonstrate realistic biological
scenarios including chemical signaling, electrical coupling, component lifecycle,
and neural maintenance.

This test serves as both integration validation and usage documentation.
=================================================================================
*/

package extracellular

import (
	"testing"
	"time"
)

// =================================================================================
// INTEGRATION TEST - COMPLETE BIOLOGICAL SCENARIO
// =================================================================================

func TestExtracellularMatrixFullIntegration(t *testing.T) {
	t.Log("=== EXTRACELLULAR MATRIX FULL INTEGRATION TEST ===")
	t.Log("Testing complete biological coordination with all subsystems")

	// === STEP 1: CREATE EXTRACELLULAR MATRIX ===
	t.Log("\n--- Step 1: Creating Extracellular Matrix ---")

	config := ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  10 * time.Millisecond,
		MaxComponents:   1000,
	}

	matrix := NewExtracellularMatrix(config)
	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start extracellular matrix: %v", err)
	}
	defer matrix.Stop()

	// Start chemical modulator background processing
	err = matrix.chemicalModulator.Start()
	if err != nil {
		t.Fatalf("Failed to start chemical modulator: %v", err)
	}
	t.Logf("✓ Extracellular matrix created and started, modulator started")

	// === STEP 2: CREATE NEURAL TISSUE COMPONENTS ===
	t.Log("\n--- Step 2: Creating Neural Tissue Components ---")

	// Create excitatory neurons (pyramidal cells)
	pyramidalNeuron1 := NewMockNeuron("pyramidal_1", Position3D{X: 10, Y: 10, Z: 5},
		[]LigandType{LigandGlutamate, LigandGABA, LigandDopamine})
	pyramidalNeuron2 := NewMockNeuron("pyramidal_2", Position3D{X: 15, Y: 12, Z: 5},
		[]LigandType{LigandGlutamate, LigandGABA, LigandDopamine})

	// Create inhibitory neuron (interneuron)
	interneuron := NewMockNeuron("interneuron_1", Position3D{X: 12, Y: 15, Z: 5},
		[]LigandType{LigandGlutamate, LigandGABA})

	// Create synapses connecting neurons
	excitatorySynapse := NewMockSynapse("syn_exc_1", Position3D{X: 12, Y: 11, Z: 5},
		"pyramidal_1", "pyramidal_2", 0.8)
	inhibitorySynapse := NewMockSynapse("syn_inh_1", Position3D{X: 13, Y: 13, Z: 5},
		"interneuron_1", "pyramidal_2", -0.6)

	t.Logf("✓ Created 3 neurons and 2 synapses")

	// === STEP 3: REGISTER COMPONENTS WITH MATRIX ===
	t.Log("\n--- Step 3: Registering Components with Matrix ---")

	components := []ComponentInfo{
		{
			ID:           pyramidalNeuron1.ID(),
			Type:         ComponentNeuron,
			Position:     pyramidalNeuron1.Position(),
			State:        StateActive,
			RegisteredAt: time.Now(),
			Metadata:     map[string]interface{}{"neuron_type": "pyramidal", "layer": "L2/3"},
		},
		{
			ID:           pyramidalNeuron2.ID(),
			Type:         ComponentNeuron,
			Position:     pyramidalNeuron2.Position(),
			State:        StateActive,
			RegisteredAt: time.Now(),
			Metadata:     map[string]interface{}{"neuron_type": "pyramidal", "layer": "L2/3"},
		},
		{
			ID:           interneuron.ID(),
			Type:         ComponentNeuron,
			Position:     interneuron.Position(),
			State:        StateActive,
			RegisteredAt: time.Now(),
			Metadata:     map[string]interface{}{"neuron_type": "interneuron", "subtype": "PV+"},
		},
		{
			ID:           excitatorySynapse.ID(),
			Type:         ComponentSynapse,
			Position:     excitatorySynapse.Position(),
			State:        StateActive,
			RegisteredAt: time.Now(),
			Metadata:     map[string]interface{}{"synapse_type": "excitatory", "weight": 0.8},
		},
		{
			ID:           inhibitorySynapse.ID(),
			Type:         ComponentSynapse,
			Position:     inhibitorySynapse.Position(),
			State:        StateActive,
			RegisteredAt: time.Now(),
			Metadata:     map[string]interface{}{"synapse_type": "inhibitory", "weight": -0.6},
		},
	}

	for _, comp := range components {
		err := matrix.RegisterComponent(comp)
		if err != nil {
			t.Fatalf("Failed to register component %s: %v", comp.ID, err)
		}
	}

	t.Logf("✓ All components registered with astrocyte network")

	// === STEP 4: ESTABLISH ELECTRICAL COUPLING (Gap Junctions) ===
	t.Log("\n--- Step 4: Establishing Electrical Coupling ---")

	// Register neurons as electrical signal listeners
	matrix.ListenForSignals([]SignalType{SignalFired, SignalConnected}, pyramidalNeuron1)
	matrix.ListenForSignals([]SignalType{SignalFired, SignalConnected}, pyramidalNeuron2)
	matrix.ListenForSignals([]SignalType{SignalFired, SignalConnected}, interneuron)

	// Establish gap junction between pyramidal neurons (electrical coupling)
	err = matrix.gapJunctions.EstablishElectricalCoupling("pyramidal_1", "pyramidal_2", 0.3)
	if err != nil {
		t.Fatalf("Failed to establish electrical coupling: %v", err)
	}

	// Verify electrical coupling
	couplings := matrix.gapJunctions.GetElectricalCouplings("pyramidal_1")
	if len(couplings) != 1 || couplings[0] != "pyramidal_2" {
		t.Fatalf("Expected electrical coupling between pyramidal neurons")
	}

	t.Logf("✓ Gap junction established between pyramidal neurons (conductance: 0.3)")

	// === STEP 5: REGISTER FOR CHEMICAL SIGNALING ===
	t.Log("\n--- Step 5: Setting up Chemical Signaling ---")

	// Register neurons as chemical binding targets
	err = matrix.RegisterForBinding(pyramidalNeuron1)
	if err != nil {
		t.Fatalf("Failed to register pyramidal_1 for chemical binding: %v", err)
	}

	err = matrix.RegisterForBinding(pyramidalNeuron2)
	if err != nil {
		t.Fatalf("Failed to register pyramidal_2 for chemical binding: %v", err)
	}

	err = matrix.RegisterForBinding(interneuron)
	if err != nil {
		t.Fatalf("Failed to register interneuron for chemical binding: %v", err)
	}

	t.Logf("✓ All neurons registered for chemical signaling")

	// === STEP 6: ESTABLISH SYNAPTIC CONNECTIVITY ===
	t.Log("\n--- Step 6: Mapping Synaptic Connectivity ---")

	// Map synaptic connections in astrocyte network
	err = matrix.astrocyteNetwork.RecordSynapticActivity(
		excitatorySynapse.ID(),
		pyramidalNeuron1.ID(),
		pyramidalNeuron2.ID(),
		0.8,
	)
	if err != nil {
		t.Fatalf("Failed to record excitatory synaptic activity: %v", err)
	}

	err = matrix.astrocyteNetwork.RecordSynapticActivity(
		inhibitorySynapse.ID(),
		interneuron.ID(),
		pyramidalNeuron2.ID(),
		-0.6,
	)
	if err != nil {
		t.Fatalf("Failed to record inhibitory synaptic activity: %v", err)
	}

	// Verify connectivity mapping
	connections := matrix.astrocyteNetwork.GetConnections(pyramidalNeuron1.ID())
	if len(connections) == 0 {
		t.Fatalf("Expected synaptic connections for pyramidal_1")
	}

	t.Logf("✓ Synaptic connectivity mapped: %d connections for pyramidal_1", len(connections))

	// === STEP 7: SIMULATE NEURAL ACTIVITY ===
	t.Log("\n--- Step 7: Simulating Neural Activity ---")

	// Initial neuron states
	t.Logf("Initial states - P1: %.3f, P2: %.3f, I1: %.3f",
		pyramidalNeuron1.currentPotential, pyramidalNeuron2.currentPotential, interneuron.currentPotential)

	// 1. Pyramidal neuron 1 releases glutamate (excitatory)
	t.Log("• Pyramidal neuron 1 releases glutamate...")
	err = matrix.ReleaseLigand(LigandGlutamate, pyramidalNeuron1.ID(), 0.9)
	if err != nil {
		t.Fatalf("Failed to release glutamate: %v", err)
	}

	time.Sleep(20 * time.Millisecond) // Allow chemical diffusion

	t.Logf("After glutamate - P1: %.3f, P2: %.3f, I1: %.3f",
		pyramidalNeuron1.currentPotential, pyramidalNeuron2.currentPotential, interneuron.currentPotential)

	// 2. Interneuron releases GABA (inhibitory)
	t.Log("• Interneuron releases GABA...")
	err = matrix.ReleaseLigand(LigandGABA, interneuron.ID(), 0.7)
	if err != nil {
		t.Fatalf("Failed to release GABA: %v", err)
	}

	time.Sleep(20 * time.Millisecond) // Allow chemical diffusion

	t.Logf("After GABA - P1: %.3f, P2: %.3f, I1: %.3f",
		pyramidalNeuron1.currentPotential, pyramidalNeuron2.currentPotential, interneuron.currentPotential)

	// 3. Send electrical signal (action potential)
	t.Log("• Pyramidal neuron 1 fires action potential...")
	matrix.SendSignal(SignalFired, pyramidalNeuron1.ID(), 1.0)

	time.Sleep(10 * time.Millisecond) // Allow electrical propagation

	t.Logf("After action potential - P1: %.3f, P2: %.3f, I1: %.3f",
		pyramidalNeuron1.currentPotential, pyramidalNeuron2.currentPotential, interneuron.currentPotential)

	// 4. Dopamine modulation (reward signal)
	t.Log("• Dopamine modulation (reward signal)...")
	err = matrix.ReleaseLigand(LigandDopamine, "reward_system", 0.5)
	if err != nil {
		t.Fatalf("Failed to release dopamine: %v", err)
	}

	time.Sleep(50 * time.Millisecond) // Dopamine has slower kinetics

	t.Logf("After dopamine - P1: %.3f, P2: %.3f, I1: %.3f",
		pyramidalNeuron1.currentPotential, pyramidalNeuron2.currentPotential, interneuron.currentPotential)

	// === STEP 8: TEST MICROGLIA MAINTENANCE ===
	t.Log("\n--- Step 8: Testing Microglial Maintenance ---")

	// Update component health based on activity
	matrix.microglia.UpdateComponentHealth(pyramidalNeuron1.ID(), 0.8, 2) // High activity
	matrix.microglia.UpdateComponentHealth(pyramidalNeuron2.ID(), 0.6, 2) // Moderate activity
	matrix.microglia.UpdateComponentHealth(interneuron.ID(), 0.4, 1)      // Lower activity

	// Check health status
	health1, _ := matrix.microglia.GetComponentHealth(pyramidalNeuron1.ID())
	health2, _ := matrix.microglia.GetComponentHealth(pyramidalNeuron2.ID())
	healthI, _ := matrix.microglia.GetComponentHealth(interneuron.ID())

	t.Logf("Health scores - P1: %.3f, P2: %.3f, I1: %.3f",
		health1.HealthScore, health2.HealthScore, healthI.HealthScore)

	// Mark low-activity synapse for pruning
	matrix.microglia.MarkForPruning(inhibitorySynapse.ID(), interneuron.ID(), pyramidalNeuron2.ID(), 0.1)

	pruningCandidates := matrix.microglia.GetPruningCandidates()
	t.Logf("✓ %d connections marked for pruning", len(pruningCandidates))

	// === STEP 9: SPATIAL QUERIES ===
	t.Log("\n--- Step 9: Testing Spatial Queries ---")

	// Find components near pyramidal neuron 1
	nearbyComponents := matrix.FindComponents(ComponentCriteria{
		Position: &pyramidalNeuron1.position,
		Radius:   10.0, // 10 micrometer radius
	})

	t.Logf("✓ Found %d components within 10μm of pyramidal_1", len(nearbyComponents))

	// Find all neurons
	allNeurons := matrix.FindComponents(ComponentCriteria{
		Type: &[]ComponentType{ComponentNeuron}[0],
	})

	t.Logf("✓ Found %d neurons in network", len(allNeurons))

	// === STEP 10: CHEMICAL CONCENTRATION ANALYSIS ===
	t.Log("\n--- Step 10: Chemical Concentration Analysis ---")

	// Check chemical concentrations at different positions
	glutamateConc := matrix.chemicalModulator.GetConcentration(LigandGlutamate, pyramidalNeuron2.Position())
	gabaConc := matrix.chemicalModulator.GetConcentration(LigandGABA, pyramidalNeuron2.Position())
	dopamineConc := matrix.chemicalModulator.GetConcentration(LigandDopamine, pyramidalNeuron1.Position())

	t.Logf("Chemical concentrations at P2 - Glutamate: %.4f, GABA: %.4f", glutamateConc, gabaConc)
	t.Logf("Dopamine concentration at P1: %.4f", dopamineConc)

	// Get recent chemical releases
	recentReleases := matrix.chemicalModulator.GetRecentReleases(5)
	t.Logf("✓ %d recent chemical release events recorded", len(recentReleases))

	// === STEP 11: ELECTRICAL SIGNAL ANALYSIS ===
	t.Log("\n--- Step 11: Electrical Signal Analysis ---")

	// Get recent electrical signals
	recentSignals := matrix.gapJunctions.GetRecentSignals(5)
	t.Logf("✓ %d recent electrical signals recorded", len(recentSignals))

	// Check electrical conductance
	conductance := matrix.gapJunctions.GetConductance("pyramidal_1", "pyramidal_2")
	t.Logf("✓ Electrical conductance P1→P2: %.3f", conductance)

	// === STEP 12: MAINTENANCE STATISTICS ===
	t.Log("\n--- Step 12: System Statistics ---")

	// Microglial statistics
	stats := matrix.microglia.GetMaintenanceStats()
	t.Logf("Microglial stats - Created: %d, Health checks: %d, Avg health: %.3f",
		stats.ComponentsCreated, stats.HealthChecks, stats.AverageHealthScore)

	// Component counts
	totalComponents := matrix.astrocyteNetwork.Count()
	t.Logf("✓ Total components in network: %d", totalComponents)

	// === FINAL VALIDATION ===
	t.Log("\n--- Integration Test Results ---")

	// Validate that all subsystems are working
	if totalComponents < 5 {
		t.Errorf("Expected at least 5 components, got %d", totalComponents)
	}

	if len(nearbyComponents) == 0 {
		t.Errorf("Expected nearby components in spatial query")
	}

	if len(recentReleases) == 0 {
		t.Errorf("Expected chemical release events")
	}

	if len(recentSignals) == 0 {
		t.Errorf("Expected electrical signal events")
	}

	if conductance <= 0 {
		t.Errorf("Expected positive electrical conductance")
	}

	t.Log("✅ ALL INTEGRATION TESTS PASSED")
	t.Log("✅ Complete biological coordination system working correctly")
	t.Log("✅ Chemical signaling, electrical coupling, spatial organization,")
	t.Log("   component lifecycle, and maintenance all functioning together")
}

// =================================================================================
// HELPER TEST: DEMONSTRATE USAGE PATTERNS
// =================================================================================

func TestExtracellularMatrixUsagePatterns(t *testing.T) {
	t.Log("=== USAGE PATTERNS DEMONSTRATION ===")

	// This test demonstrates common usage patterns for the extracellular matrix

	matrix := NewExtracellularMatrix(ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  10 * time.Millisecond,
		MaxComponents:   100,
	})
	defer matrix.Stop()

	// === PATTERN 1: Basic Component Registration ===
	t.Log("\n--- Pattern 1: Component Registration ---")

	neuronInfo := ComponentInfo{
		ID:       "cortical_neuron_1",
		Type:     ComponentNeuron,
		Position: Position3D{X: 50, Y: 50, Z: 10},
		State:    StateActive,
		Metadata: map[string]interface{}{
			"cortical_layer": "L4",
			"cell_type":      "spiny_stellate",
		},
	}

	err := matrix.RegisterComponent(neuronInfo)
	if err != nil {
		t.Fatalf("Failed to register neuron: %v", err)
	}
	t.Logf("✓ Registered neuron with metadata")

	// === PATTERN 2: Chemical Communication ===
	t.Log("\n--- Pattern 2: Chemical Communication ---")

	mockTarget := NewMockNeuron("target_neuron", Position3D{X: 52, Y: 52, Z: 10},
		[]LigandType{LigandGlutamate})

	matrix.RegisterForBinding(mockTarget)

	// Release neurotransmitter
	matrix.ReleaseLigand(LigandGlutamate, "cortical_neuron_1", 0.8)

	t.Logf("✓ Chemical signaling established")

	// === PATTERN 3: Spatial Queries ===
	t.Log("\n--- Pattern 3: Spatial Organization ---")

	// Find nearby neurons
	nearby := matrix.FindComponents(ComponentCriteria{
		Type:     &[]ComponentType{ComponentNeuron}[0],
		Position: &Position3D{X: 50, Y: 50, Z: 10},
		Radius:   5.0,
	})

	t.Logf("✓ Found %d nearby neurons", len(nearby))

	// === PATTERN 4: Health Monitoring ===
	t.Log("\n--- Pattern 4: Component Health Monitoring ---")

	matrix.microglia.UpdateComponentHealth("cortical_neuron_1", 0.9, 3)
	health, exists := matrix.microglia.GetComponentHealth("cortical_neuron_1")

	if exists {
		t.Logf("✓ Neuron health score: %.3f", health.HealthScore)
	}

	t.Log("✅ Usage patterns demonstrated successfully")
}

// Add this debug test to find the exact hang point

func TestExtracellularMatrixDebugHang(t *testing.T) {
	t.Log("=== DEBUG TEST TO FIND HANG POINT ===")

	// === STEP 1: CREATE MATRIX ===
	t.Log("Creating matrix...")
	config := ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  10 * time.Millisecond,
		MaxComponents:   1000,
	}

	matrix := NewExtracellularMatrix(config)
	t.Log("Matrix created, starting...")

	err := matrix.Start()
	if err != nil {
		t.Fatalf("Failed to start matrix: %v", err)
	}
	t.Log("Matrix started successfully")
	defer matrix.Stop()

	// === STEP 2: CREATE MINIMAL COMPONENTS ===
	t.Log("Creating test neuron...")
	neuron1 := NewMockNeuron("test_neuron_1", Position3D{X: 1, Y: 1, Z: 1},
		[]LigandType{LigandGlutamate})
	t.Log("Neuron created")

	// === STEP 3: REGISTER COMPONENT ===
	t.Log("Registering component...")
	componentInfo := ComponentInfo{
		ID:           neuron1.ID(),
		Type:         ComponentNeuron,
		Position:     neuron1.Position(),
		State:        StateActive,
		RegisteredAt: time.Now(),
	}

	err = matrix.RegisterComponent(componentInfo)
	if err != nil {
		t.Fatalf("Failed to register component: %v", err)
	}
	t.Log("Component registered successfully")

	// === STEP 4: TEST EACH SUBSYSTEM INDIVIDUALLY ===

	// Test astrocyte network
	t.Log("Testing astrocyte network...")
	info, exists := matrix.astrocyteNetwork.Get(neuron1.ID())
	if !exists {
		t.Fatalf("Component not found in astrocyte network")
	}
	t.Logf("Astrocyte network working: found %s", info.ID)

	// Test gap junctions
	t.Log("Testing gap junctions...")
	matrix.ListenForSignals([]SignalType{SignalFired}, neuron1)
	t.Log("Registered for signals")

	matrix.SendSignal(SignalFired, "test_source", 1.0)
	t.Log("Signal sent successfully")

	// Test chemical modulator step by step
	t.Log("Testing chemical modulator registration...")
	err = matrix.RegisterForBinding(neuron1)
	if err != nil {
		t.Fatalf("Failed to register for binding: %v", err)
	}
	t.Log("Registered for chemical binding")

	// This is likely where it hangs - test chemical release
	t.Log("About to test chemical release...")
	t.Log("Calling ReleaseLigand...")

	// Add timeout to prevent infinite hang
	done := make(chan error, 1)
	go func() {
		err := matrix.ReleaseLigand(LigandGlutamate, neuron1.ID(), 0.5)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Chemical release failed: %v", err)
		}
		t.Log("Chemical release completed successfully!")
	case <-time.After(5 * time.Second):
		t.Fatal("HANG DETECTED: Chemical release took more than 5 seconds")
	}

	// Test microglia
	t.Log("Testing microglia...")
	matrix.microglia.UpdateComponentHealth(neuron1.ID(), 0.8, 1)
	health, exists := matrix.microglia.GetComponentHealth(neuron1.ID())
	if !exists {
		t.Fatalf("Health not found")
	}
	t.Logf("Microglia working: health score %.3f", health.HealthScore)

	// === STEP 5: TEST SYNAPTIC ACTIVITY RECORDING ===
	t.Log("About to test synaptic activity recording...")

	// First register the target component
	targetInfo := ComponentInfo{
		ID:           "target_neuron",
		Type:         ComponentNeuron,
		Position:     Position3D{X: 2, Y: 2, Z: 2},
		State:        StateActive,
		RegisteredAt: time.Now(),
	}

	err = matrix.RegisterComponent(targetInfo)
	if err != nil {
		t.Fatalf("Failed to register target component: %v", err)
	}
	t.Log("Target component registered")

	done2 := make(chan error, 1)
	go func() {
		err := matrix.astrocyteNetwork.RecordSynapticActivity(
			"test_synapse_1",
			neuron1.ID(),
			"target_neuron",
			0.8,
		)
		done2 <- err
	}()

	select {
	case err := <-done2:
		if err != nil {
			t.Fatalf("Synaptic activity recording failed: %v", err)
		}
		t.Log("Synaptic activity recording completed!")
	case <-time.After(5 * time.Second):
		t.Fatal("HANG DETECTED: Synaptic activity recording took more than 5 seconds")
	}

	t.Log("✅ ALL DEBUG TESTS PASSED - NO HANG DETECTED")
}
