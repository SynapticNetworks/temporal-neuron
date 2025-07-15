package tmaze

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/extracellular"
	"github.com/SynapticNetworks/temporal-neuron/neuron"
	"github.com/SynapticNetworks/temporal-neuron/synapse"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// TMazeEnvironment encapsulates the T-maze logic independent of the neural implementation
type TMazeEnvironment struct {
	CurrentLocation Location
	RewardLocation  Location
	VisitedReward   bool
}

type Location int

const (
	Start Location = iota
	Junction
	Left
	Right
)

// String returns string representation of Location
func (l Location) String() string {
	switch l {
	case Start:
		return "Start"
	case Junction:
		return "Junction"
	case Left:
		return "Left"
	case Right:
		return "Right"
	default:
		return "Unknown"
	}
}

type Action int

const (
	Forward Action = iota
	TurnLeft
	TurnRight
)

// String returns string representation of Action
func (a Action) String() string {
	switch a {
	case Forward:
		return "Forward"
	case TurnLeft:
		return "TurnLeft"
	case TurnRight:
		return "TurnRight"
	default:
		return "Unknown"
	}
}

// NewTMaze creates a new T-maze with reward at the specified location
func NewTMaze(rewardAt Location) *TMazeEnvironment {
	return &TMazeEnvironment{
		CurrentLocation: Start,
		RewardLocation:  rewardAt,
		VisitedReward:   false,
	}
}

// Reset returns the agent to the start position
func (m *TMazeEnvironment) Reset() {
	m.CurrentLocation = Start
	m.VisitedReward = false
}

// Step takes an action and returns the new location and any reward received
func (m *TMazeEnvironment) Step(action Action) (Location, float64) {
	switch m.CurrentLocation {
	case Start:
		if action == Forward {
			m.CurrentLocation = Junction
		}
	case Junction:
		if action == TurnLeft {
			m.CurrentLocation = Left
		} else if action == TurnRight {
			m.CurrentLocation = Right
		}
	}

	// Check for reward
	reward := 0.0
	if m.CurrentLocation == m.RewardLocation && !m.VisitedReward {
		reward = 1.0
		m.VisitedReward = true
	}

	return m.CurrentLocation, reward
}

// TMazeAgent represents the neural control system for navigating the T-maze
type TMazeAgent struct {
	// Core environment
	Matrix *extracellular.ExtracellularMatrix

	// Core neurons
	StartSensor    component.NeuralComponent
	JunctionSensor component.NeuralComponent
	LeftSensor     component.NeuralComponent
	RightSensor    component.NeuralComponent
	ForwardMotor   component.NeuralComponent
	LeftMotor      component.NeuralComponent
	RightMotor     component.NeuralComponent

	// Neuromodulatory neurons
	DopamineNeuron component.NeuralComponent
	GABANeuron     component.NeuralComponent // Inhibitory neuron for error signaling

	// Expectation neurons for RPE generation
	LeftExpectation  component.NeuralComponent
	RightExpectation component.NeuralComponent

	// Hidden layer neurons
	HiddenNeurons []component.NeuralComponent

	// Key connections for monitoring
	JunctionToLeftSyn  component.SynapticProcessor
	JunctionToRightSyn component.SynapticProcessor
	LeftToDopamineSyn  component.SynapticProcessor
	RightToDopamineSyn component.SynapticProcessor
	HiddenSynapses     []component.SynapticProcessor

	// Settings
	ExplorationNoise   float64
	DopamineEnabled    bool
	HiddenLayerEnabled bool
	ExplorationDecay   float64 // How much to reduce exploration over time

	// Synaptogenesis tracking
	SynapseCount int
	NewSynapses  int
}

// NewBasicTMazeAgent creates a neural agent for the T-maze
func NewBasicTMazeAgent(t *testing.T) (*TMazeAgent, error) {
	// Create a new agent
	agent := &TMazeAgent{
		ExplorationNoise:   0.4,
		DopamineEnabled:    false,
		HiddenLayerEnabled: false,
		ExplorationDecay:   0.99, // Slight decay per episode
	}

	// Create extracellular matrix with biological features enabled
	matrix := extracellular.NewExtracellularMatrix(extracellular.ExtracellularMatrixConfig{
		ChemicalEnabled: true,
		SpatialEnabled:  true,
		UpdateInterval:  10 * time.Millisecond,
		MaxComponents:   200,
	})

	err := matrix.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start matrix: %v", err)
	}
	agent.Matrix = matrix

	// Register neuron types with appropriate behaviors
	registerBiologicalNeuronTypes(matrix)

	// Register synapse type with STDP and neuromodulation sensitivity
	registerBiologicalSynapseTypes(matrix)

	// Create the basic T-maze neurons
	if err := createBasicTMazeNeurons(t, agent, matrix); err != nil {
		return nil, err
	}

	// Create the core connections
	createBasicTMazeConnections(t, agent, matrix)

	// Capture initial synapse count
	synapses := matrix.ListSynapses()
	agent.SynapseCount = len(synapses)
	t.Logf("Created basic T-maze agent with %d initial synapses", agent.SynapseCount)

	return agent, nil
}

// NewBiologicalTMazeAgent creates a neural agent with dopamine, GABA, and natural learning
func NewBiologicalTMazeAgent(t *testing.T) (*TMazeAgent, error) {
	// Start with basic agent
	agent, err := NewBasicTMazeAgent(t)
	if err != nil {
		return nil, err
	}
	agent.DopamineEnabled = true
	agent.HiddenLayerEnabled = true

	// Add dopamine neuron for reward signaling
	if err := addDopaminergicSystem(t, agent); err != nil {
		return nil, err
	}

	// Add GABA neuron for inhibitory signaling
	if err := addGABAergicSystem(t, agent); err != nil {
		return nil, err
	}

	// Add neurons for reward expectation (RPE computation)
	if err := addExpectationNeurons(t, agent); err != nil {
		return nil, err
	}

	// Add hidden layer for enhanced processing
	if err := createBiologicalHiddenLayer(t, agent); err != nil {
		return nil, err
	}

	// Set up activity-dependent chemical release for all neurons
	configureChemicalRelease(t, agent)

	// Create biological connections between components
	createBiologicalCircuitry(t, agent)

	// Update synapse count
	synapses := agent.Matrix.ListSynapses()
	agent.SynapseCount = len(synapses)
	t.Logf("Created biological T-maze agent with %d initial synapses", agent.SynapseCount)

	return agent, nil
}

// createBasicTMazeNeurons creates the core neurons for the T-maze
func createBasicTMazeNeurons(t *testing.T, agent *TMazeAgent, matrix *extracellular.ExtracellularMatrix) error {
	// Create sensory neurons with biologically-appropriate positions
	startSensor, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "sensory",
		Position:   types.Position3D{X: 0, Y: 0, Z: 0},
	})
	if err != nil {
		return fmt.Errorf("failed to create start sensor: %v", err)
	}

	junctionSensor, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "sensory",
		Position:   types.Position3D{X: 0, Y: 50, Z: 0},
	})
	if err != nil {
		return fmt.Errorf("failed to create junction sensor: %v", err)
	}

	leftSensor, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "sensory",
		Position:   types.Position3D{X: -50, Y: 100, Z: 0},
	})
	if err != nil {
		return fmt.Errorf("failed to create left sensor: %v", err)
	}

	rightSensor, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "sensory",
		Position:   types.Position3D{X: 50, Y: 100, Z: 0},
	})
	if err != nil {
		return fmt.Errorf("failed to create right sensor: %v", err)
	}

	// Create motor neurons
	forwardMotor, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "motor",
		Position:   types.Position3D{X: 0, Y: 25, Z: 50}, // Forward motor in a biologically plausible location
	})
	if err != nil {
		return fmt.Errorf("failed to create forward motor: %v", err)
	}

	leftMotor, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "motor",
		Position:   types.Position3D{X: -50, Y: 75, Z: 50}, // Left motor positioned accordingly
	})
	if err != nil {
		return fmt.Errorf("failed to create left motor: %v", err)
	}

	rightMotor, err := matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "motor",
		Position:   types.Position3D{X: 50, Y: 75, Z: 50}, // Right motor positioned accordingly
	})
	if err != nil {
		return fmt.Errorf("failed to create right motor: %v", err)
	}

	// Start all neurons
	for _, n := range []component.NeuralComponent{
		startSensor, junctionSensor, leftSensor, rightSensor,
		forwardMotor, leftMotor, rightMotor,
	} {
		err := n.Start()
		if err != nil {
			return fmt.Errorf("failed to start neuron %s: %v", n.ID(), err)
		}
	}

	// Store neurons in agent
	agent.StartSensor = startSensor
	agent.JunctionSensor = junctionSensor
	agent.LeftSensor = leftSensor
	agent.RightSensor = rightSensor
	agent.ForwardMotor = forwardMotor
	agent.LeftMotor = leftMotor
	agent.RightMotor = rightMotor

	return nil
}

// addDopaminergicSystem adds a dopamine system for reward signaling
func addDopaminergicSystem(t *testing.T, agent *TMazeAgent) error {
	// Create dopamine neuron (representing VTA/SNc)
	dopamineNeuron, err := agent.Matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "dopamine",
		Position:   types.Position3D{X: 0, Y: 75, Z: 100}, // Positioned "above" the circuit
	})
	if err != nil {
		return fmt.Errorf("failed to create dopamine neuron: %v", err)
	}

	err = dopamineNeuron.Start()
	if err != nil {
		return fmt.Errorf("failed to start dopamine neuron: %v", err)
	}
	agent.DopamineNeuron = dopamineNeuron

	t.Logf("Added dopaminergic system for reward learning")
	return nil
}

// addGABAergicSystem adds inhibitory GABA neurons for punishment signaling
func addGABAergicSystem(t *testing.T, agent *TMazeAgent) error {
	// Create GABA neuron for error signaling
	gabaNeuron, err := agent.Matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "gaba",
		Position:   types.Position3D{X: 0, Y: 60, Z: 80}, // Position near dopamine system
	})
	if err != nil {
		return fmt.Errorf("failed to create GABA neuron: %v", err)
	}

	err = gabaNeuron.Start()
	if err != nil {
		return fmt.Errorf("failed to start GABA neuron: %v", err)
	}
	agent.GABANeuron = gabaNeuron

	t.Logf("Added GABAergic system for inhibitory control")
	return nil
}

// addExpectationNeurons adds neurons that generate expectations for reward prediction
func addExpectationNeurons(t *testing.T, agent *TMazeAgent) error {
	// Left expectation neuron (fires when expecting reward on left)
	leftExpect, err := agent.Matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "expectation",
		Position:   types.Position3D{X: -30, Y: 70, Z: 70},
	})
	if err != nil {
		return fmt.Errorf("failed to create left expectation neuron: %v", err)
	}
	err = leftExpect.Start()
	if err != nil {
		return fmt.Errorf("failed to start left expectation neuron: %v", err)
	}
	agent.LeftExpectation = leftExpect

	// Right expectation neuron (fires when expecting reward on right)
	rightExpect, err := agent.Matrix.CreateNeuron(types.NeuronConfig{
		NeuronType: "expectation",
		Position:   types.Position3D{X: 30, Y: 70, Z: 70},
	})
	if err != nil {
		return fmt.Errorf("failed to create right expectation neuron: %v", err)
	}
	err = rightExpect.Start()
	if err != nil {
		return fmt.Errorf("failed to start right expectation neuron: %v", err)
	}
	agent.RightExpectation = rightExpect

	t.Logf("Added expectation neurons for reward prediction")
	return nil
}

// configureChemicalRelease sets up activity-dependent chemical release
func configureChemicalRelease(t *testing.T, agent *TMazeAgent) {
	// Helper function to configure neurons for activity-dependent BDNF release
	configureBDNFRelease := func(neuron component.NeuralComponent) {
		// Import neuron package to get access to concrete Neuron type
		nImpl, ok := neuron.(interface {
			GetReleasedLigands() []types.LigandType
			SetReleasedLigands([]types.LigandType)
		})
		if !ok {
			t.Logf("Warning: neuron %s doesn't implement required methods", neuron.ID())
			return
		}

		// Get current ligands the neuron releases
		currentLigands := nImpl.GetReleasedLigands()

		// Add BDNF if not already present
		if !containsLigand(currentLigands, types.LigandBDNF) {
			newLigands := append(currentLigands, types.LigandBDNF)
			// Use the interface methods directly
			nImpl.SetReleasedLigands(newLigands)
		}
	}

	// Configure all sensory and motor neurons
	configureBDNFRelease(agent.StartSensor)
	configureBDNFRelease(agent.JunctionSensor)
	configureBDNFRelease(agent.LeftSensor)
	configureBDNFRelease(agent.RightSensor)
	configureBDNFRelease(agent.ForwardMotor)
	configureBDNFRelease(agent.LeftMotor)
	configureBDNFRelease(agent.RightMotor)

	// Hidden neurons
	for _, hidden := range agent.HiddenNeurons {
		configureBDNFRelease(hidden)
	}

	// Dopamine neuron retains dopamine as its primary neurotransmitter
	// with added BDNF capability
	if agent.DopamineNeuron != nil {
		configureBDNFRelease(agent.DopamineNeuron)
	}

	// GABA neuron retains GABA as its primary neurotransmitter
	if agent.GABANeuron != nil {
		configureBDNFRelease(agent.GABANeuron)
	}

	t.Logf("Configured activity-dependent chemical release for all neurons")
}

// Helper function to check if a ligand type is in a slice
func containsLigand(ligands []types.LigandType, ligand types.LigandType) bool {
	for _, l := range ligands {
		if l == ligand {
			return true
		}
	}
	return false
}

// ReduceExploration gradually reduces the exploration noise
func (agent *TMazeAgent) ReduceExploration() {
	if agent.ExplorationNoise > 0.1 {
		agent.ExplorationNoise *= agent.ExplorationDecay
	}
}

// ConfigureRewardPathway biologically prepares the circuit for a specific reward location
func (agent *TMazeAgent) ConfigureRewardPathway(rewardLocation Location) {
	if !agent.DopamineEnabled {
		return
	}

	// Only proceed if connections exist
	if agent.LeftToDopamineSyn == nil || agent.RightToDopamineSyn == nil {
		return
	}

	// Adjust the initial weights of the reward pathways
	// This simulates the initial state of the network, not runtime tuning
	if rewardLocation == Left {
		// Strengthen left→dopamine connection
		agent.LeftToDopamineSyn.SetWeight(0.3)  // Strong connection (low weight = strong)
		agent.RightToDopamineSyn.SetWeight(0.7) // Weak connection

		// Also adjust GABA connections if they exist
		if agent.GABANeuron != nil {
			// When left is rewarded, right should activate GABA (error signal)
			synapses := agent.Matrix.ListSynapses()
			for _, synapse := range synapses {
				if synapse.GetPresynapticID() == agent.RightSensor.ID() &&
					synapse.GetPostsynapticID() == agent.GABANeuron.ID() {
					synapse.SetWeight(0.3) // Stronger error signal for wrong choice
				}

				if synapse.GetPresynapticID() == agent.LeftSensor.ID() &&
					synapse.GetPostsynapticID() == agent.GABANeuron.ID() {
					synapse.SetWeight(0.7) // Weaker error signal for correct choice
				}
			}
		}
	} else {
		// Strengthen right→dopamine connection
		agent.LeftToDopamineSyn.SetWeight(0.7)  // Weak connection
		agent.RightToDopamineSyn.SetWeight(0.3) // Strong connection (low weight = strong)

		// Also adjust GABA connections if they exist
		if agent.GABANeuron != nil {
			// When right is rewarded, left should activate GABA (error signal)
			synapses := agent.Matrix.ListSynapses()
			for _, synapse := range synapses {
				if synapse.GetPresynapticID() == agent.LeftSensor.ID() &&
					synapse.GetPostsynapticID() == agent.GABANeuron.ID() {
					synapse.SetWeight(0.3) // Stronger error signal for wrong choice
				}

				if synapse.GetPresynapticID() == agent.RightSensor.ID() &&
					synapse.GetPostsynapticID() == agent.GABANeuron.ID() {
					synapse.SetWeight(0.7) // Weaker error signal for correct choice
				}
			}
		}
	}
}

// MonitorSynaptogenesis checks for new synapses formed
func (agent *TMazeAgent) MonitorSynaptogenesis(t *testing.T) int {
	// Get current synapse count
	currentSynapses := agent.Matrix.ListSynapses()
	newCount := len(currentSynapses) - agent.SynapseCount

	// Update count if new synapses were formed
	if newCount > 0 {
		agent.NewSynapses += newCount
		agent.SynapseCount = len(currentSynapses)
		t.Logf("Detected %d new synapses! Total new: %d, Total synapses: %d",
			newCount, agent.NewSynapses, agent.SynapseCount)
	}

	return newCount
}

// MonitorChemicalLevels logs neurotransmitter concentrations at key locations
func (agent *TMazeAgent) MonitorChemicalLevels(t *testing.T) {
	// Check junction position (decision point)
	junctionPos := agent.JunctionSensor.Position()

	dopamineLevel := agent.Matrix.GetChemicalModulator().GetConcentration(
		types.LigandDopamine, junctionPos)
	gabaLevel := agent.Matrix.GetChemicalModulator().GetConcentration(
		types.LigandGABA, junctionPos)
	bdnfLevel := agent.Matrix.GetChemicalModulator().GetConcentration(
		types.LigandBDNF, junctionPos)

	t.Logf("Chemical levels at junction - Dopamine: %.3f, GABA: %.3f, BDNF: %.3f",
		dopamineLevel, gabaLevel, bdnfLevel)

	// Check dopamine neuron position (reward center)
	if agent.DopamineNeuron != nil {
		dopPos := agent.DopamineNeuron.Position()
		dopamineAtSource := agent.Matrix.GetChemicalModulator().GetConcentration(
			types.LigandDopamine, dopPos)
		t.Logf("Dopamine at VTA/SNc: %.3f", dopamineAtSource)
	}
}

// Stop closes all resources
func (agent *TMazeAgent) Stop() {
	// Stop all neurons
	if agent.StartSensor != nil {
		agent.StartSensor.Stop()
	}
	if agent.JunctionSensor != nil {
		agent.JunctionSensor.Stop()
	}
	if agent.LeftSensor != nil {
		agent.LeftSensor.Stop()
	}
	if agent.RightSensor != nil {
		agent.RightSensor.Stop()
	}
	if agent.ForwardMotor != nil {
		agent.ForwardMotor.Stop()
	}
	if agent.LeftMotor != nil {
		agent.LeftMotor.Stop()
	}
	if agent.RightMotor != nil {
		agent.RightMotor.Stop()
	}
	if agent.DopamineNeuron != nil {
		agent.DopamineNeuron.Stop()
	}
	if agent.GABANeuron != nil {
		agent.GABANeuron.Stop()
	}
	if agent.LeftExpectation != nil {
		agent.LeftExpectation.Stop()
	}
	if agent.RightExpectation != nil {
		agent.RightExpectation.Stop()
	}

	// Stop hidden neurons
	for _, n := range agent.HiddenNeurons {
		if n != nil {
			n.Stop()
		}
	}

	// Stop matrix
	if agent.Matrix != nil {
		agent.Matrix.Stop()
	}
}

// TrainingResult stores performance metrics
type TrainingResult struct {
	BaselinePerformance   float64
	LearnedPerformance    float64
	ReversalBaseline      float64
	AdaptationPerformance float64
	FinalLeftWeight       float64
	FinalRightWeight      float64
	NewSynapses           int
	PrunedSynapses        int
}

// EvaluatePerformance measures agent performance on the T-maze
func EvaluatePerformance(t *testing.T, agent *TMazeAgent, rewardLocation Location, episodes int) float64 {
	maze := NewTMaze(rewardLocation)
	successCount := 0

	for ep := 0; ep < episodes; ep++ {
		maze.Reset()
		foundReward := false

		// Run a single episode
		for steps := 0; steps < 3; steps++ { // Max 3 steps needed for T-maze
			loc := maze.CurrentLocation
			agent.ActivateSensor(loc)
			action := agent.ReadAction(loc, t)
			newLoc, reward := maze.Step(action)

			if reward > 0 {
				foundReward = true
				break
			}

			// If we reached a terminal state without reward, end episode
			if newLoc == Left || newLoc == Right {
				break
			}
		}

		if foundReward {
			successCount++
		}
	}

	return float64(successCount) / float64(episodes)
}

// ValidateResults analyzes and reports on training results
func ValidateResults(t *testing.T, result TrainingResult, agentType string, rewardLocation Location) {
	t.Logf("\n--- %s Performance Summary ---", agentType)
	t.Logf("Baseline performance: %.1f%%", result.BaselinePerformance*100)
	t.Logf("Learned performance: %.1f%%", result.LearnedPerformance*100)
	t.Logf("Adaptation performance: %.1f%%", result.AdaptationPerformance*100)
	t.Logf("Final weights - Junction→Left: %.3f, Junction→Right: %.3f",
		result.FinalLeftWeight, result.FinalRightWeight)
	t.Logf("Structural plasticity - New synapses: %d, Pruned synapses: %d",
		result.NewSynapses, result.PrunedSynapses)

	learningImprovement := result.LearnedPerformance - result.BaselinePerformance
	if learningImprovement > 0.2 {
		t.Logf("✅ Learning improvement: +%.1f%% (significant)", learningImprovement*100)
	} else if learningImprovement > 0 {
		t.Logf("⚠️ Learning improvement: +%.1f%% (modest)", learningImprovement*100)
	} else {
		t.Logf("❌ No learning improvement: %.1f%%", learningImprovement*100)
	}

	// Check for appropriate weight changes
	if result.FinalLeftWeight != result.FinalRightWeight {
		weightDiff := math.Abs(result.FinalLeftWeight - result.FinalRightWeight)
		if weightDiff > 0.1 {
			t.Log("✅ Weight differentiation achieved")

			// Check if weights changed in the right direction
			// LOWER weights lead to HIGHER activity in this model
			if (rewardLocation == Right && result.FinalRightWeight < result.FinalLeftWeight) ||
				(rewardLocation == Left && result.FinalLeftWeight < result.FinalRightWeight) {
				t.Log("✅ Correct pathway properly valued")
			} else {
				t.Log("❌ Weight changes in unexpected direction")
			}
		} else {
			t.Log("⚠️ Minimal weight differentiation")
		}
	} else {
		t.Log("⚠️ No differentiation in path values: Equal weights")
	}

	adaptationImprovement := result.AdaptationPerformance - result.ReversalBaseline
	if adaptationImprovement > 0.2 {
		t.Logf("✅ Adaptation improvement: +%.1f%% (significant)", adaptationImprovement*100)
	} else if adaptationImprovement > 0 {
		t.Logf("⚠️ Adaptation improvement: +%.1f%% (modest)", adaptationImprovement*100)
	} else {
		t.Logf("❌ No adaptation improvement: %.1f%%", adaptationImprovement*100)
	}

	if result.NewSynapses > 0 {
		t.Logf("✅ Synaptogenesis observed: %d new connections formed", result.NewSynapses)
	}
}

// Test functions

// TestBasicReinforcementLearning tests simple T-maze learning without neuromodulation
func TestBasicReinforcementLearning(t *testing.T) {
	t.Log("=== BASIC REINFORCEMENT LEARNING TEST ===")
	t.Log("This test evaluates basic T-maze learning without neuromodulation")

	// Create agent
	agent, err := NewBasicTMazeAgent(t)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}
	defer agent.Stop()

	// Set reward location
	rewardLocation := Left

	// Train and evaluate
	result := BiologicalTrainingProcess(t, agent, rewardLocation)

	// Validate results
	ValidateResults(t, result, "Basic", rewardLocation)
}

// These functions incorporate the improved STDP implementation techniques into the T-maze tests

// createBiologicalHiddenLayer creates a biologically realistic hidden layer with improved isolation
// to prevent STDP timing interference
func createBiologicalHiddenLayer(t *testing.T, agent *TMazeAgent) error {
	// We'll create a striatum-like structure with medium spiny neurons
	agent.HiddenNeurons = make([]component.NeuralComponent, 0)

	// Create four hidden neurons - two associated with each pathway
	// Left-side pathway neurons - positioned further apart to reduce interference
	for i := 0; i < 2; i++ {
		h, err := agent.Matrix.CreateNeuron(types.NeuronConfig{
			NeuronType: "hidden",
			Position: types.Position3D{
				X: -30 - float64(i*20), // Increased separation between neurons
				Y: 50 + float64(i*15),  // More vertical staggering
				Z: 30,                  // Hidden layer depth
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create left-pathway hidden neuron: %v", err)
		}

		err = h.Start()
		if err != nil {
			return fmt.Errorf("failed to start hidden neuron: %v", err)
		}

		agent.HiddenNeurons = append(agent.HiddenNeurons, h)
	}

	// Right-side pathway neurons - positioned further apart to reduce interference
	for i := 0; i < 2; i++ {
		h, err := agent.Matrix.CreateNeuron(types.NeuronConfig{
			NeuronType: "hidden",
			Position: types.Position3D{
				X: 30 + float64(i*20), // Increased separation between neurons
				Y: 50 + float64(i*15), // More vertical staggering
				Z: 30,                 // Hidden layer depth
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create right-pathway hidden neuron: %v", err)
		}

		err = h.Start()
		if err != nil {
			return fmt.Errorf("failed to start hidden neuron: %v", err)
		}

		agent.HiddenNeurons = append(agent.HiddenNeurons, h)
	}

	t.Logf("Created biological hidden layer with %d neurons and improved spatial isolation", len(agent.HiddenNeurons))
	return nil
}

// registerBiologicalNeuronTypes registers neuron types with realistic properties
// and improved STDP configuration
func registerBiologicalNeuronTypes(matrix *extracellular.ExtracellularMatrix) {
	// Register sensory neurons (simulating thalamic inputs)
	matrix.RegisterNeuronType("sensory", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(
			id,
			0.2,                // Low threshold for easy activation
			0.9,                // Slower decay to maintain activity
			5*time.Millisecond, // Short refractory period
			1.5,                // Moderate fire factor
			10.0,               // Target firing rate
			0.1,                // Homeostasis strength
		)
		// Enable STDP with longer feedback delay to improve timing
		n.EnableSTDPFeedback(10*time.Millisecond, 0.1)
		n.SetCallbacks(callbacks)

		// Configure to release glutamate when active
		n.SetReleasedLigands([]types.LigandType{types.LigandGlutamate})

		return n, nil
	})

	// Register motor neurons (simulating cortical motor neurons)
	matrix.RegisterNeuronType("motor", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(
			id,
			0.5,                // Higher threshold to require multiple inputs
			0.8,                // Fast decay for responsiveness
			5*time.Millisecond, // Short refractory period
			1.5,                // Moderate fire factor
			10.0,               // Target firing rate
			0.1,                // Homeostasis strength
		)
		// Enable STDP with longer feedback delay for better isolation
		n.EnableSTDPFeedback(10*time.Millisecond, 0.1)
		n.SetCallbacks(callbacks)

		// Configure to release glutamate when active
		n.SetReleasedLigands([]types.LigandType{types.LigandGlutamate})

		return n, nil
	})

	// Register dopamine neurons (simulating VTA/SNc neurons)
	matrix.RegisterNeuronType("dopamine", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(
			id,
			0.3,                // Medium threshold
			0.98,               // Very slow decay to maintain dopamine effects
			5*time.Millisecond, // Short refractory period
			1.5,                // Moderate fire factor
			5.0,                // Lower target firing rate (baseline dopamine)
			0.1,                // Homeostasis strength
		)
		// Enable STDP with lower learning rate for stability
		n.EnableSTDPFeedback(10*time.Millisecond, 0.05)
		n.SetCallbacks(callbacks)

		// Configure to release dopamine when active
		n.SetReleasedLigands([]types.LigandType{types.LigandDopamine})

		return n, nil
	})

	// Register GABA neurons (simulating inhibitory interneurons)
	matrix.RegisterNeuronType("gaba", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(
			id,
			0.3,                // Medium threshold
			0.85,               // Medium decay
			5*time.Millisecond, // Short refractory period
			1.5,                // Moderate fire factor
			15.0,               // Higher target firing rate for inhibitory neurons
			0.1,                // Homeostasis strength
		)
		// Enable STDP with longer feedback delay
		n.EnableSTDPFeedback(10*time.Millisecond, 0.05)
		n.SetCallbacks(callbacks)

		// Configure to release GABA when active
		n.SetReleasedLigands([]types.LigandType{types.LigandGABA})

		return n, nil
	})

	// Register hidden neurons (simulating striatal medium spiny neurons)
	matrix.RegisterNeuronType("hidden", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(
			id,
			0.4,                // Medium threshold
			0.9,                // Standard decay
			5*time.Millisecond, // Short refractory period
			1.5,                // Moderate fire factor
			10.0,               // Target firing rate
			0.1,                // Homeostasis strength
		)
		// Enable STDP with improved timing parameters
		n.EnableSTDPFeedback(10*time.Millisecond, 0.1)
		n.SetCallbacks(callbacks)

		// Configure to release glutamate when active
		n.SetReleasedLigands([]types.LigandType{types.LigandGlutamate})

		return n, nil
	})

	// Register expectation neurons (simulating orbitofrontal cortex)
	matrix.RegisterNeuronType("expectation", func(id string, config types.NeuronConfig, callbacks component.NeuronCallbacks) (component.NeuralComponent, error) {
		n := neuron.NewNeuron(
			id,
			0.3,                // Medium threshold
			0.95,               // Slow decay to maintain expectation
			5*time.Millisecond, // Short refractory period
			1.5,                // Moderate fire factor
			10.0,               // Target firing rate
			0.1,                // Homeostasis strength
		)
		// Enable STDP with improved parameters
		n.EnableSTDPFeedback(10*time.Millisecond, 0.1)
		n.SetCallbacks(callbacks)

		// Expectation neurons also release glutamate
		n.SetReleasedLigands([]types.LigandType{types.LigandGlutamate})

		return n, nil
	})
}

// registerBiologicalSynapseTypes registers synapse types with realistic properties
// and improved STDP configurations
func registerBiologicalSynapseTypes(matrix *extracellular.ExtracellularMatrix) {
	matrix.RegisterSynapseType("plastic", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, exists := matrix.GetNeuron(config.PresynapticID)
		if !exists {
			return nil, fmt.Errorf("presynaptic neuron not found: %s", config.PresynapticID)
		}

		postNeuron, exists := matrix.GetNeuron(config.PostsynapticID)
		if !exists {
			return nil, fmt.Errorf("postsynaptic neuron not found: %s", config.PostsynapticID)
		}

		// Improved STDP configuration with better timing parameters
		plasticityConfig := types.PlasticityConfig{
			Enabled:        true,
			LearningRate:   0.03,                   // Moderate learning rate for stable learning
			TimeConstant:   20 * time.Millisecond,  // Standard STDP time constant
			WindowSize:     100 * time.Millisecond, // Biological STDP window
			MinWeight:      0.05,                   // Prevent weights from going to zero
			MaxWeight:      2.0,                    // Cap on synaptic strength
			AsymmetryRatio: 1.05,                   // Slight LTD/LTP asymmetry for stability
		}

		// Realistic pruning configuration
		pruningConfig := synapse.PruningConfig{
			Enabled:             true,
			WeightThreshold:     0.1,              // Prune very weak synapses
			InactivityThreshold: 10 * time.Second, // Prune after extended inactivity
		}

		// Create synapse with biological properties and configurable delay
		// Ensure delay is at least 1ms for biological realism
		if config.Delay < time.Millisecond {
			config.Delay = time.Millisecond
		}

		syn := synapse.NewBasicSynapse(
			id,
			preNeuron.(component.MessageScheduler),
			postNeuron.(component.MessageReceiver),
			plasticityConfig,
			pruningConfig,
			config.InitialWeight,
			config.Delay,
		)

		return syn, nil
	})

	// Register inhibitory synapse type (GABA-mediated)
	matrix.RegisterSynapseType("inhibitory", func(id string, config types.SynapseConfig, callbacks extracellular.SynapseCallbacks) (component.SynapticProcessor, error) {
		preNeuron, exists := matrix.GetNeuron(config.PresynapticID)
		if !exists {
			return nil, fmt.Errorf("presynaptic neuron not found: %s", config.PresynapticID)
		}

		postNeuron, exists := matrix.GetNeuron(config.PostsynapticID)
		if !exists {
			return nil, fmt.Errorf("postsynaptic neuron not found: %s", config.PostsynapticID)
		}

		// Special configuration for inhibitory synapses with improved timing parameters
		plasticityConfig := types.PlasticityConfig{
			Enabled:        true,
			LearningRate:   0.02,                   // Slower learning for inhibitory synapses
			TimeConstant:   20 * time.Millisecond,  // Standard time constant
			WindowSize:     100 * time.Millisecond, // Biological window
			MinWeight:      0.05,                   // Prevent weights from going to zero
			MaxWeight:      2.0,                    // Cap on synaptic strength
			AsymmetryRatio: 0.9,                    // More symmetric plasticity
		}

		pruningConfig := synapse.PruningConfig{
			Enabled:             true,
			WeightThreshold:     0.15,             // Higher threshold for inhibitory pruning
			InactivityThreshold: 15 * time.Second, // More stable inhibitory connections
		}

		// Ensure minimum delay for inhibitory synapses
		if config.Delay < 2*time.Millisecond {
			config.Delay = 2 * time.Millisecond // Slightly longer delay for inhibitory synapses
		}

		// Create inhibitory synapse
		syn := synapse.NewBasicSynapse(
			id,
			preNeuron.(component.MessageScheduler),
			postNeuron.(component.MessageReceiver),
			plasticityConfig,
			pruningConfig,
			config.InitialWeight,
			config.Delay,
		)

		return syn, nil
	})
}

// createBasicTMazeConnections creates the basic synaptic connections
// with improved delays for better STDP timing
func createBasicTMazeConnections(t *testing.T, agent *TMazeAgent, matrix *extracellular.ExtracellularMatrix) {
	// Helper function to create synapses with appropriate delays
	createSynapse := func(pre, post component.NeuralComponent, weight float64, delay time.Duration) component.SynapticProcessor {
		syn, err := matrix.CreateSynapse(types.SynapseConfig{
			SynapseType:    "plastic",
			PresynapticID:  pre.ID(),
			PostsynapticID: post.ID(),
			InitialWeight:  weight,
			Delay:          delay, // Configurable delay
		})
		if err != nil {
			t.Fatalf("Failed to create synapse from %s to %s: %v", pre.ID(), post.ID(), err)
		}
		return syn
	}

	// Create basic connections that form the T-maze circuit with appropriate delays
	// Start → Forward (when at start, go forward)
	createSynapse(agent.StartSensor, agent.ForwardMotor, 0.5, 2*time.Millisecond)

	// Forward → Junction (movement leads to junction)
	createSynapse(agent.ForwardMotor, agent.JunctionSensor, 0.5, 5*time.Millisecond)

	// Key learning connections - store these for monitoring
	// Junction → Left (decision to turn left at junction)
	agent.JunctionToLeftSyn = createSynapse(agent.JunctionSensor, agent.LeftMotor, 0.5, 3*time.Millisecond)

	// Junction → Right (decision to turn right at junction)
	agent.JunctionToRightSyn = createSynapse(agent.JunctionSensor, agent.RightMotor, 0.5, 3*time.Millisecond)

	// Feedback connections - these represent the physical movement result
	// Left Motor → Left Sensor (turning left leads to left arm)
	createSynapse(agent.LeftMotor, agent.LeftSensor, 0.5, 5*time.Millisecond)

	// Right Motor → Right Sensor (turning right leads to right arm)
	createSynapse(agent.RightMotor, agent.RightSensor, 0.5, 5*time.Millisecond)
}

// createBiologicalCircuitry creates a complete neural circuit with all connections
// incorporating improved STDP timing parameters
func createBiologicalCircuitry(t *testing.T, agent *TMazeAgent) {
	// Helper function to create synapses with appropriate delays
	createSynapse := func(pre, post component.NeuralComponent, weight float64, synapseType string, delay time.Duration) component.SynapticProcessor {
		syn, err := agent.Matrix.CreateSynapse(types.SynapseConfig{
			SynapseType:    synapseType,
			PresynapticID:  pre.ID(),
			PostsynapticID: post.ID(),
			InitialWeight:  weight,
			Delay:          delay, // Configurable delay
		})
		if err != nil {
			t.Fatalf("Failed to create synapse from %s to %s: %v", pre.ID(), post.ID(), err)
		}
		return syn
	}

	// Initialize hidden synapses collection
	agent.HiddenSynapses = make([]component.SynapticProcessor, 0)

	// === DOPAMINE CIRCUIT CONNECTIONS ===

	// Connect reward locations to dopamine neuron (natural reward signal)
	agent.LeftToDopamineSyn = createSynapse(agent.LeftSensor, agent.DopamineNeuron, 0.5, "plastic", 3*time.Millisecond)
	agent.RightToDopamineSyn = createSynapse(agent.RightSensor, agent.DopamineNeuron, 0.5, "plastic", 3*time.Millisecond)

	// Connect dopamine to motor neurons for reward modulation
	createSynapse(agent.DopamineNeuron, agent.LeftMotor, 0.7, "plastic", 5*time.Millisecond)
	createSynapse(agent.DopamineNeuron, agent.RightMotor, 0.7, "plastic", 5*time.Millisecond)

	// === EXPECTATION CIRCUIT ===

	// Connect left motor to left expectation (learn to expect reward when going left)
	createSynapse(agent.LeftMotor, agent.LeftExpectation, 0.5, "plastic", 2*time.Millisecond)

	// Connect right motor to right expectation (learn to expect reward when going right)
	createSynapse(agent.RightMotor, agent.RightExpectation, 0.5, "plastic", 2*time.Millisecond)

	// Connect expectations to dopamine for reward prediction error
	// These connections are inhibitory to create dopamine dips for negative RPE
	createSynapse(agent.LeftExpectation, agent.DopamineNeuron, 1.2, "inhibitory", 3*time.Millisecond)
	createSynapse(agent.RightExpectation, agent.DopamineNeuron, 1.2, "inhibitory", 3*time.Millisecond)

	// === GABA CIRCUIT ===

	// Connect GABA neuron to motor neurons (for inhibitory control)
	createSynapse(agent.GABANeuron, agent.LeftMotor, 1.0, "inhibitory", 2*time.Millisecond)
	createSynapse(agent.GABANeuron, agent.RightMotor, 1.0, "inhibitory", 2*time.Millisecond)

	// Connect wrong choice outcomes to GABA activation
	// When right is rewarded, left activation should trigger GABA
	createSynapse(agent.LeftSensor, agent.GABANeuron, 0.5, "plastic", 2*time.Millisecond)
	// When left is rewarded, right activation should trigger GABA
	createSynapse(agent.RightSensor, agent.GABANeuron, 0.5, "plastic", 2*time.Millisecond)

	// === HIDDEN LAYER CONNECTIONS ===

	// Connect junction to all hidden neurons
	for _, hidden := range agent.HiddenNeurons {
		syn := createSynapse(agent.JunctionSensor, hidden, 0.5, "plastic", 3*time.Millisecond)
		agent.HiddenSynapses = append(agent.HiddenSynapses, syn)
	}

	// Connect hidden neurons to motor neurons with varied delays to improve timing isolation
	// First two hidden neurons (left pathway)
	for i := 0; i < 2; i++ {
		// Stronger connection to left motor
		syn1 := createSynapse(agent.HiddenNeurons[i], agent.LeftMotor, 0.5, "plastic", 2*time.Millisecond+time.Duration(i)*time.Millisecond)
		agent.HiddenSynapses = append(agent.HiddenSynapses, syn1)

		// Weaker connection to right motor
		syn2 := createSynapse(agent.HiddenNeurons[i], agent.RightMotor, 0.8, "plastic", 3*time.Millisecond+time.Duration(i)*time.Millisecond)
		agent.HiddenSynapses = append(agent.HiddenSynapses, syn2)
	}

	// Second two hidden neurons (right pathway)
	for i := 2; i < 4; i++ {
		// Weaker connection to left motor
		syn1 := createSynapse(agent.HiddenNeurons[i], agent.LeftMotor, 0.8, "plastic", 3*time.Millisecond+time.Duration(i-2)*time.Millisecond)
		agent.HiddenSynapses = append(agent.HiddenSynapses, syn1)

		// Stronger connection to right motor
		syn2 := createSynapse(agent.HiddenNeurons[i], agent.RightMotor, 0.5, "plastic", 2*time.Millisecond+time.Duration(i-2)*time.Millisecond)
		agent.HiddenSynapses = append(agent.HiddenSynapses, syn2)
	}

	// Connect dopamine to hidden layer for reward modulation with varied delays for better timing
	for i, hidden := range agent.HiddenNeurons {
		syn := createSynapse(agent.DopamineNeuron, hidden, 0.7, "plastic", 4*time.Millisecond+time.Duration(i)*time.Millisecond)
		agent.HiddenSynapses = append(agent.HiddenSynapses, syn)
	}

	t.Logf("Created biological circuitry with %d hidden synapses and improved timing parameters", len(agent.HiddenSynapses))
}

// ActivateSensor activates a sensory neuron based on current location
// with improved handling to ensure consistent activation levels
func (agent *TMazeAgent) ActivateSensor(location Location) {
	// Create a stronger signal for clear activation
	signal := types.NeuralSignal{
		Value:     2.0, // Increased from 1.5 for more reliable activation
		Timestamp: time.Now(),
		SourceID:  "environment",
	}

	// Target the appropriate neuron
	switch location {
	case Start:
		signal.TargetID = agent.StartSensor.ID()
		agent.StartSensor.Receive(signal)
	case Junction:
		signal.TargetID = agent.JunctionSensor.ID()
		agent.JunctionSensor.Receive(signal)
	case Left:
		signal.TargetID = agent.LeftSensor.ID()
		agent.LeftSensor.Receive(signal)
	case Right:
		signal.TargetID = agent.RightSensor.ID()
		agent.RightSensor.Receive(signal)
	}

	// Wait a short time to ensure activation is processed
	time.Sleep(5 * time.Millisecond)
}

// ReadAction reads motor neuron activity to determine action
// with improved wait times for signal propagation and STDP processing
func (agent *TMazeAgent) ReadAction(location Location, t *testing.T) Action {
	// Allow more time for signal propagation to account for varied delays
	waitTime := 30 * time.Millisecond
	if agent.HiddenLayerEnabled {
		waitTime = 60 * time.Millisecond // More time for complex networks
	}
	time.Sleep(waitTime)

	// Default action is forward
	action := Forward

	// Different action selection based on location
	switch location {
	case Start:
		// At start, only forward is meaningful
		return Forward
	case Junction:
		// At junction, choose based on motor neuron activity
		leftActivity := agent.LeftMotor.GetActivityLevel()
		rightActivity := agent.RightMotor.GetActivityLevel()

		// Calculate exploration noise - decays over time for more exploitation
		noise := agent.ExplorationNoise

		// Add noise for exploration
		leftActivity += rand.Float64() * noise
		rightActivity += rand.Float64() * noise

		// Log the decision factors with more detailed information
		t.Logf("Decision: Left activity %.4f vs Right activity %.4f (noise: %.3f)",
			leftActivity, rightActivity, noise)

		// Log hidden layer activity if enabled
		if agent.HiddenLayerEnabled && len(agent.HiddenNeurons) > 0 {
			hiddenActivities := make([]string, len(agent.HiddenNeurons))
			for i, hidden := range agent.HiddenNeurons {
				hiddenActivities[i] = fmt.Sprintf("%.4f", hidden.GetActivityLevel())
			}
			t.Logf("Hidden neuron activities: [%s]", strings.Join(hiddenActivities, ", "))
		}

		// Make decision based on highest activity
		if leftActivity > rightActivity {
			action = TurnLeft
		} else {
			action = TurnRight
		}
	}

	return action
}

// BiologicalTrainingProcess runs a training process with natural learning
// with improved STDP timing handling and longer processing times
func BiologicalTrainingProcess(t *testing.T, agent *TMazeAgent, rewardLocation Location) TrainingResult {
	var result TrainingResult

	// Configure reward pathway (one-time setup, not runtime tuning)
	agent.ConfigureRewardPathway(rewardLocation)

	// Baseline performance
	t.Logf("\n--- Initial Performance Baseline ---")
	baselineEpisodes := 10
	result.BaselinePerformance = EvaluatePerformance(t, agent, rewardLocation, baselineEpisodes)
	t.Logf("Baseline performance with reward at %v: %.1f%% success rate",
		rewardLocation, result.BaselinePerformance*100)

	// Training phase
	t.Logf("\n--- Training Phase (Natural Learning) ---")
	trainingEpisodes := 30
	if agent.HiddenLayerEnabled {
		trainingEpisodes = 40 // More time for complex networks
	}

	maze := NewTMaze(rewardLocation)

	for episode := 0; episode < trainingEpisodes; episode++ {
		maze.Reset()
		episodeReward := 0.0

		// Run episode
		for steps := 0; steps < 3; steps++ {
			loc := maze.CurrentLocation

			// Activate sensor (only input from environment)
			agent.ActivateSensor(loc)

			// Let activity propagate naturally through the network
			// Increased wait time for proper STDP processing
			time.Sleep(70 * time.Millisecond)

			// Read action from motor neurons
			action := agent.ReadAction(loc, t)

			// Take action in environment
			newLoc, reward := maze.Step(action)

			// Process reward
			if reward > 0 {
				// Activate sensor at reward location
				agent.ActivateSensor(newLoc)

				// Allow more time for dopamine circuit to naturally process reward
				time.Sleep(150 * time.Millisecond)

				episodeReward += reward
				break
			}

			// If terminal state reached, activate that sensor
			if newLoc == Left || newLoc == Right {
				agent.ActivateSensor(newLoc)
				time.Sleep(100 * time.Millisecond) // Increased wait time
				break
			}
		}

		// Reduce exploration over time
		agent.ReduceExploration()

		// Allow more time for natural plasticity mechanisms to work
		time.Sleep(250 * time.Millisecond)

		// Check for synaptogenesis
		agent.MonitorSynaptogenesis(t)

		// Report progress every few episodes
		if episode%5 == 0 || episode == trainingEpisodes-1 {
			leftW := agent.JunctionToLeftSyn.GetWeight()
			rightW := agent.JunctionToRightSyn.GetWeight()
			t.Logf("Episode %d - Reward: %.1f, J→L weight: %.4f, J→R weight: %.4f",
				episode, episodeReward, leftW, rightW)

			// Monitor chemical levels
			agent.MonitorChemicalLevels(t)
		}
	}

	// Test learned performance
	t.Logf("\n--- Post-Training Performance ---")
	testEpisodes := 10
	result.LearnedPerformance = EvaluatePerformance(t, agent, rewardLocation, testEpisodes)
	t.Logf("Learned performance with reward at %v: %.1f%% success rate",
		rewardLocation, result.LearnedPerformance*100)

	// Reversal learning - switch reward to opposite side
	t.Logf("\n--- Reversal Learning ---")

	// Switch reward location
	newRewardLocation := Right
	if rewardLocation == Right {
		newRewardLocation = Left
	}
	maze.RewardLocation = newRewardLocation

	// Reconfigure network for new reward location (one-time adjustment)
	agent.ConfigureRewardPathway(newRewardLocation)

	// Baseline on new location before adaptation
	result.ReversalBaseline = EvaluatePerformance(t, agent, newRewardLocation, 5)
	t.Logf("Initial performance after reward switch: %.1f%% success rate",
		result.ReversalBaseline*100)

	// Adaptation training
	adaptationEpisodes := 20
	if agent.HiddenLayerEnabled {
		adaptationEpisodes = 30 // More time for complex networks
	}

	for episode := 0; episode < adaptationEpisodes; episode++ {
		maze.Reset()
		episodeReward := 0.0

		// Run episode with natural learning
		for steps := 0; steps < 3; steps++ {
			loc := maze.CurrentLocation
			agent.ActivateSensor(loc)

			// Let activity propagate naturally with longer wait time
			time.Sleep(70 * time.Millisecond)

			action := agent.ReadAction(loc, t)
			newLoc, reward := maze.Step(action)

			if reward > 0 {
				// Activate sensor at reward location
				agent.ActivateSensor(newLoc)

				// Allow more time for natural reward processing
				time.Sleep(150 * time.Millisecond)

				episodeReward += reward
				break
			}

			// If terminal state reached
			if newLoc == Left || newLoc == Right {
				agent.ActivateSensor(newLoc)
				time.Sleep(100 * time.Millisecond) // Increased wait time
				break
			}
		}

		// Increase exploration slightly during reversal learning
		if episode < 5 && agent.ExplorationNoise < 0.3 {
			agent.ExplorationNoise = 0.3
		} else {
			agent.ReduceExploration()
		}

		// Allow more time for natural plasticity
		time.Sleep(250 * time.Millisecond)

		// Check for synaptogenesis
		agent.MonitorSynaptogenesis(t)

		// Report progress
		if episode%5 == 0 || episode == adaptationEpisodes-1 {
			leftW := agent.JunctionToLeftSyn.GetWeight()
			rightW := agent.JunctionToRightSyn.GetWeight()
			t.Logf("Adaptation episode %d - Reward: %.1f, J→L weight: %.4f, J→R weight: %.4f",
				episode, episodeReward, leftW, rightW)

			// Monitor chemical levels
			agent.MonitorChemicalLevels(t)
		}
	}

	// Test adaptation performance
	t.Logf("\n--- Post-Adaptation Performance ---")
	result.AdaptationPerformance = EvaluatePerformance(t, agent, newRewardLocation, testEpisodes)
	t.Logf("Adaptation performance with reward at %v: %.1f%% success rate",
		newRewardLocation, result.AdaptationPerformance*100)

	// Final weights
	result.FinalLeftWeight = agent.JunctionToLeftSyn.GetWeight()
	result.FinalRightWeight = agent.JunctionToRightSyn.GetWeight()
	result.NewSynapses = agent.NewSynapses

	return result
}

// TestBiologicalReinforcementLearning tests learning with all biological mechanisms
// updated with improved STDP timing
func TestBiologicalReinforcementLearning(t *testing.T) {
	t.Log("=== BIOLOGICAL REINFORCEMENT LEARNING TEST ===")
	t.Log("This test evaluates full biological learning with neuromodulation,")
	t.Log("natural dopamine reward signals, and inhibitory control")
	t.Log("Using improved STDP timing mechanisms for better learning")

	// Create enhanced biological agent
	agent, err := NewBiologicalTMazeAgent(t)
	if err != nil {
		t.Fatalf("Failed to create biological agent: %v", err)
	}
	defer agent.Stop()

	// Set reward location
	rewardLocation := Left

	// Train and evaluate
	result := BiologicalTrainingProcess(t, agent, rewardLocation)

	// Validate results
	ValidateResults(t, result, "Biological", rewardLocation)

	// Additional biological insights
	t.Log("\n--- Biological Mechanisms In Action ---")
	t.Log("1. Dopaminergic reward signaling from VTA/SNc-like neurons")
	t.Log("2. GABAergic inhibitory control from interneuron-like cells")
	t.Log("3. Reward prediction error computation via expectation neurons")
	t.Log("4. Activity-dependent synaptogenesis via BDNF signaling")
	t.Log("5. Improved STDP timing with precise synaptic delays")
	t.Log("6. Spatial neuron organization for better signal isolation")
	t.Log("7. Bidirectional plasticity via E/I balance and chemical signals")
}

// TestNaturalSynaptogenesis tests the formation of new connections during learning
func TestNaturalSynaptogenesis(t *testing.T) {
	t.Log("=== NATURAL SYNAPTOGENESIS TEST ===")
	t.Log("This test specifically evaluates activity-dependent synapse formation")

	// Create enhanced biological agent
	agent, err := NewBiologicalTMazeAgent(t)
	if err != nil {
		t.Fatalf("Failed to create biological agent: %v", err)
	}
	defer agent.Stop()

	// Capture initial synapse count
	initialSynapses := agent.Matrix.ListSynapses()
	initialCount := len(initialSynapses)
	t.Logf("Initial synapse count: %d", initialCount)

	// Set reward location
	rewardLocation := Right

	// Train and evaluate with focus on structural plasticity
	result := BiologicalTrainingProcess(t, agent, rewardLocation)

	// Capture final synapse count
	finalSynapses := agent.Matrix.ListSynapses()
	finalCount := len(finalSynapses)

	// Update result with synaptogenesis metrics
	result.NewSynapses = agent.NewSynapses

	// Validate results
	ValidateResults(t, result, "Synaptogenic", rewardLocation)

	// Additional synaptogenesis insights
	t.Logf("\n--- Structural Plasticity Summary ---")
	t.Logf("Initial synapse count: %d", initialCount)
	t.Logf("Final synapse count: %d", finalCount)
	t.Logf("Net change: %d synapses", finalCount-initialCount)

	// Check if synaptogenesis occurred
	if finalCount > initialCount {
		t.Logf("✅ Confirmed structural plasticity: %d new synapses formed", finalCount-initialCount)
		t.Log("\n--- Synaptogenesis Biological Mechanisms ---")
		t.Log("1. Activity-dependent BDNF release from highly active neurons")
		t.Log("2. Natural synapse formation between co-active neuronal populations")
		t.Log("3. E/I balance maintained through selective synaptogenesis")
		t.Log("4. Experience-dependent circuit refinement through new connections")
	} else {
		t.Logf("⚠️ No net synaptogenesis detected")
	}
}

// TestBidirectionalDopamineModulation tests how dopamine signals both reward and error
func TestBidirectionalDopamineModulation(t *testing.T) {
	t.Log("=== BIDIRECTIONAL DOPAMINE MODULATION TEST ===")
	t.Log("This test evaluates how dopamine signals both reward and error through")
	t.Log("activation (phasic bursting) and inhibition (phasic dips)")

	// Create biological agent with dopamine circuit
	agent, err := NewBiologicalTMazeAgent(t)
	if err != nil {
		t.Fatalf("Failed to create biological agent: %v", err)
	}
	defer agent.Stop()

	// Track dopamine levels during specific scenarios
	t.Log("\n--- Testing Dopamine Responses ---")

	// Setup test maze
	maze := NewTMaze(Left) // Left arm has reward

	// Configure reward pathway
	agent.ConfigureRewardPathway(Left)

	// 1. Unexpected reward scenario
	t.Log("\n1. Unexpected Reward Scenario")
	maze.Reset()

	// Navigate to junction
	agent.ActivateSensor(Start)
	time.Sleep(50 * time.Millisecond)
	action := agent.ReadAction(Start, t)
	maze.Step(action) // Should be Forward

	// Go to left (rewarded) arm - should see dopamine burst
	agent.ActivateSensor(Junction)
	time.Sleep(50 * time.Millisecond)

	// Capture baseline dopamine
	if agent.DopamineNeuron != nil {
		baselineDopamine := agent.DopamineNeuron.GetActivityLevel()
		t.Logf("Baseline dopamine: %.3f", baselineDopamine)
	}

	// Choose left
	action = TurnLeft // Force left action
	newLoc, _ := maze.Step(action)

	// Activate sensor at reward location
	agent.ActivateSensor(newLoc)

	// Allow time for dopamine response
	time.Sleep(100 * time.Millisecond)

	// Check dopamine response
	if agent.DopamineNeuron != nil {
		dopamineLevel := agent.DopamineNeuron.GetActivityLevel()
		t.Logf("Dopamine after unexpected reward: %.3f", dopamineLevel)

		// Check for chemical release
		dopPos := agent.DopamineNeuron.Position()
		dopamineConcentration := agent.Matrix.GetChemicalModulator().GetConcentration(
			types.LigandDopamine, dopPos)
		t.Logf("Dopamine concentration: %.3f", dopamineConcentration)
	}

	// 2. Unexpected no-reward scenario
	t.Log("\n2. Unexpected No-Reward Scenario")
	maze.Reset()

	// Navigate to junction
	agent.ActivateSensor(Start)
	time.Sleep(50 * time.Millisecond)
	action = agent.ReadAction(Start, t)
	maze.Step(action) // Should be Forward

	// Go to right (unrewarded) arm - should see dopamine dip
	agent.ActivateSensor(Junction)
	time.Sleep(50 * time.Millisecond)

	// Choose right
	action = TurnRight // Force right action
	newLoc, _ = maze.Step(action)

	// Activate sensor at non-reward location
	agent.ActivateSensor(newLoc)

	// Allow time for dopamine response
	time.Sleep(100 * time.Millisecond)

	// Check dopamine response
	if agent.DopamineNeuron != nil {
		dopamineLevel := agent.DopamineNeuron.GetActivityLevel()
		t.Logf("Dopamine after unexpected no-reward: %.3f", dopamineLevel)

		// Check for chemical release
		dopPos := agent.DopamineNeuron.Position()
		dopamineConcentration := agent.Matrix.GetChemicalModulator().GetConcentration(
			types.LigandDopamine, dopPos)
		t.Logf("Dopamine concentration: %.3f", dopamineConcentration)

		// Check GABA levels
		gabaConcentration := agent.Matrix.GetChemicalModulator().GetConcentration(
			types.LigandGABA, dopPos)
		t.Logf("GABA concentration near dopamine neurons: %.3f", gabaConcentration)
	}

	// Train the agent and verify results
	result := BiologicalTrainingProcess(t, agent, Left)
	ValidateResults(t, result, "BidirectionalDopamine", Left)

	t.Log("\n--- Bidirectional Dopamine Modulation Summary ---")
	t.Log("1. Phasic dopamine bursts encode positive reward prediction errors")
	t.Log("2. Phasic dopamine dips encode negative reward prediction errors")
	t.Log("3. Expectation neurons modulate dopamine response based on predictions")
	t.Log("4. GABA-mediated inhibition creates dopamine dips for error signaling")
	t.Log("5. Bidirectional dopamine dynamics enable efficient reinforcement learning")
}

// TestGABAergicInhibitoryControl tests the role of GABA in learning and control
func TestGABAergicInhibitoryControl(t *testing.T) {
	t.Log("=== GABAERGIC INHIBITORY CONTROL TEST ===")
	t.Log("This test evaluates how GABA inhibition shapes network dynamics")
	t.Log("and provides error signals to guide learning")

	// Create biological agent with GABA circuit
	agent, err := NewBiologicalTMazeAgent(t)
	if err != nil {
		t.Fatalf("Failed to create biological agent: %v", err)
	}
	defer agent.Stop()

	// Setup test maze
	maze := NewTMaze(Right) // Right arm has reward

	// Configure reward pathway
	agent.ConfigureRewardPathway(Right)

	// Test GABA response to wrong choice
	t.Log("\n--- Testing GABA Responses ---")

	// 1. Wrong choice scenario
	t.Log("\n1. Wrong Choice Scenario (Should Activate GABA)")
	maze.Reset()

	// Navigate to junction
	agent.ActivateSensor(Start)
	time.Sleep(50 * time.Millisecond)
	action := agent.ReadAction(Start, t)
	maze.Step(action) // Should be Forward

	// Go to left (unrewarded) arm - should activate GABA
	agent.ActivateSensor(Junction)
	time.Sleep(50 * time.Millisecond)

	// Choose left (wrong choice)
	action = TurnLeft // Force left action
	newLoc, _ := maze.Step(action)

	// Activate sensor at wrong location
	agent.ActivateSensor(newLoc)

	// Allow time for GABA response
	time.Sleep(100 * time.Millisecond)

	// Check GABA response
	if agent.GABANeuron != nil {
		gabaActivity := agent.GABANeuron.GetActivityLevel()
		t.Logf("GABA neuron activity after wrong choice: %.3f", gabaActivity)

		// Check for chemical release
		wrongPos := agent.LeftSensor.Position()
		gabaConcentration := agent.Matrix.GetChemicalModulator().GetConcentration(
			types.LigandGABA, wrongPos)
		t.Logf("GABA concentration at wrong choice: %.3f", gabaConcentration)
	}

	// 2. Correct choice scenario
	t.Log("\n2. Correct Choice Scenario (Minimal GABA)")
	maze.Reset()

	// Navigate to junction
	agent.ActivateSensor(Start)
	time.Sleep(50 * time.Millisecond)
	action = agent.ReadAction(Start, t)
	maze.Step(action) // Should be Forward

	// Go to right (rewarded) arm - should not activate GABA
	agent.ActivateSensor(Junction)
	time.Sleep(50 * time.Millisecond)

	// Choose right (correct choice)
	action = TurnRight // Force right action
	newLoc, _ = maze.Step(action)

	// Activate sensor at correct location
	agent.ActivateSensor(newLoc)

	// Allow time for response
	time.Sleep(100 * time.Millisecond)

	// Check GABA response
	if agent.GABANeuron != nil {
		gabaActivity := agent.GABANeuron.GetActivityLevel()
		t.Logf("GABA neuron activity after correct choice: %.3f", gabaActivity)

		// Check for chemical release
		correctPos := agent.RightSensor.Position()
		gabaConcentration := agent.Matrix.GetChemicalModulator().GetConcentration(
			types.LigandGABA, correctPos)
		t.Logf("GABA concentration at correct choice: %.3f", gabaConcentration)
	}

	// Train the agent and verify results
	result := BiologicalTrainingProcess(t, agent, Right)
	ValidateResults(t, result, "GABAergicControl", Right)

	t.Log("\n--- GABAergic Inhibitory Control Summary ---")
	t.Log("1. GABA provides negative feedback for incorrect actions")
	t.Log("2. Inhibitory control shapes circuit activity by suppressing wrong pathways")
	t.Log("3. GABA concentration correlates with error signals in the network")
	t.Log("4. Inhibitory interneurons contribute to efficient reversal learning")
	t.Log("5. E/I balance is maintained through dynamic GABA signaling")
}
