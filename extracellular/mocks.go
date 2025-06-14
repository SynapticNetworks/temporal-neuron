package extracellular

// =================================================================================
// MOCK COMPONENTS FOR TESTING
// =================================================================================

// MockNeuron represents a simple neuron for testing
type MockNeuron struct {
	id               string
	position         Position3D
	receptors        []LigandType
	firingThreshold  float64
	currentPotential float64
	connections      []string
	isActive         bool
}

func NewMockNeuron(id string, pos Position3D, receptors []LigandType) *MockNeuron {
	return &MockNeuron{
		id:               id,
		position:         pos,
		receptors:        receptors,
		firingThreshold:  0.7,
		currentPotential: 0.0,
		connections:      make([]string, 0),
		isActive:         true,
	}
}

// Implement Component interface
func (mn *MockNeuron) ID() string                   { return mn.id }
func (mn *MockNeuron) Position() Position3D         { return mn.position }
func (mn *MockNeuron) ComponentType() ComponentType { return ComponentNeuron }

// Implement BindingTarget interface (for chemical signaling)
func (mn *MockNeuron) Bind(ligandType LigandType, sourceID string, concentration float64) {
	// Simple binding model: accumulate potential
	switch ligandType {
	case LigandGlutamate:
		mn.currentPotential += concentration * 0.8 // Increased from 0.5
	case LigandGABA:
		mn.currentPotential -= concentration * 0.5 // Increased from 0.3
	case LigandDopamine:
		mn.currentPotential += concentration * 0.4 // Increased from 0.2
	case LigandSerotonin:
		mn.currentPotential += concentration * 0.3
	case LigandAcetylcholine:
		mn.currentPotential += concentration * 0.6
	}

	// Lower threshold for better response
	firingThreshold := 0.3
	if mn.currentPotential > firingThreshold && mn.isActive {
		// Neuron fires - but don't reset to 0, just reduce
		mn.currentPotential *= 0.7 // Partial reset keeps some activation
	}
}

// Override OnSignal for better electrical responsiveness
func (mn *MockNeuron) OnSignal(signalType SignalType, sourceID string, data interface{}) {
	switch signalType {
	case SignalFired:
		// Stronger response to action potentials
		if value, ok := data.(float64); ok && mn.isActive {
			mn.currentPotential += value * 0.2 // Increased from 0.1
		}
	case SignalConnected:
		if connID, ok := data.(string); ok {
			mn.connections = append(mn.connections, connID)
		}
	}
}

func (mn *MockNeuron) GetReceptors() []LigandType { return mn.receptors }
func (mn *MockNeuron) GetPosition() Position3D    { return mn.position }

// Helper methods for testing
func (mn *MockNeuron) GetCurrentPotential() float64 {
	return mn.currentPotential
}

func (mn *MockNeuron) SetPotential(potential float64) {
	mn.currentPotential = potential
}

func (mn *MockNeuron) GetConnections() []string {
	return mn.connections
}

func (mn *MockNeuron) IsActive() bool {
	return mn.isActive
}

func (mn *MockNeuron) SetActive(active bool) {
	mn.isActive = active
}

// MockSynapse represents a simple synapse for testing
type MockSynapse struct {
	id           string
	position     Position3D
	presynaptic  string
	postsynaptic string
	weight       float64
	activity     float64
}

func NewMockSynapse(id string, pos Position3D, pre, post string, weight float64) *MockSynapse {
	return &MockSynapse{
		id:           id,
		position:     pos,
		presynaptic:  pre,
		postsynaptic: post,
		weight:       weight,
		activity:     0.0,
	}
}

// Implement Component interface
func (ms *MockSynapse) ID() string                   { return ms.id }
func (ms *MockSynapse) Position() Position3D         { return ms.position }
func (ms *MockSynapse) ComponentType() ComponentType { return ComponentSynapse }

// Helper methods for testing
func (ms *MockSynapse) GetWeight() float64 {
	return ms.weight
}

func (ms *MockSynapse) SetWeight(weight float64) {
	ms.weight = weight
}

func (ms *MockSynapse) GetActivity() float64 {
	return ms.activity
}

func (ms *MockSynapse) SetActivity(activity float64) {
	ms.activity = activity
}

func (ms *MockSynapse) GetPresynaptic() string {
	return ms.presynaptic
}

func (ms *MockSynapse) GetPostsynaptic() string {
	return ms.postsynaptic
}
