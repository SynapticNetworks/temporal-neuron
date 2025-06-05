package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/neuron"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	neuron     *neuron.Neuron
	fireEvents chan neuron.FireEvent
	messages   []string
	step       int
	totalFires int

	// Step state
	stepRunning   bool
	stepCountdown int
	stepMessage   string

	// Neuron state
	neuronState string // "RESTING", "ACTIVE", "FIRING", "REFRACTORY"
	accumulator float64
	lastInput   float64
	lastFired   time.Time

	// Animation state
	inputActive  bool
	outputActive bool

	// Terminal size
	width  int
	height int
}

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33"))
	stepStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))

	// States
	restingStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	activeStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	firingStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	refractoryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("129"))
)

type fireEventMsg neuron.FireEvent
type tickMsg time.Time
type stepActionMsg struct {
	action string
	value  float64
}

func initialModel() Model {
	fireEvents := make(chan neuron.FireEvent, 100)

	// Working parameters
	n := neuron.NewNeuron("test", 1.0, 0.95, 1*time.Second, 1.0)
	n.SetFireEventChannel(fireEvents)
	go n.Run()

	return Model{
		neuron:      n,
		fireEvents:  fireEvents,
		messages:    []string{"Ready to test leaky integration"},
		step:        1,
		neuronState: "RESTING",
		width:       120,
		height:      40,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		listenForFires(m.fireEvents),
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func listenForFires(fireEvents <-chan neuron.FireEvent) tea.Cmd {
	return func() tea.Msg {
		select {
		case event := <-fireEvents:
			return fireEventMsg(event)
		default:
			return nil
		}
	}
}

func stepAction(action string, value float64, delay time.Duration) tea.Cmd {
	return func() tea.Msg {
		if delay > 0 {
			time.Sleep(delay)
		}
		return stepActionMsg{action: action, value: value}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.neuron.Close()
			return m, tea.Quit
		case "1":
			return m.executeStep1()
		case "2":
			return m.executeStep2()
		case "3":
			return m.executeStep3()
		case "t":
			// Test fire
			m.addMessage("TEST: Sending 1.5 signal")
			input := m.neuron.GetInput()
			input <- neuron.Message{Value: 1.5}
			m.lastInput = 1.5
			m.inputActive = true
			m.neuronState = "ACTIVE"
		case "r":
			return initialModel(), initialModel().Init()
		}

	case stepActionMsg:
		// Handle delayed step actions
		m.addMessage(fmt.Sprintf("→ %s: %.1f", msg.action, msg.value))
		input := m.neuron.GetInput()
		input <- neuron.Message{Value: msg.value}
		m.lastInput = msg.value
		m.inputActive = true
		if msg.value > 0.5 {
			m.neuronState = "ACTIVE"
		}

	case fireEventMsg:
		m.totalFires++
		m.neuronState = "FIRING"
		m.outputActive = true
		m.lastFired = time.Now()
		m.addMessage(fmt.Sprintf("🔥 FIRED! Value: %.2f", float64(msg.Value)))
		return m, listenForFires(m.fireEvents)

	case tickMsg:
		m.updateState()
		if m.stepRunning && m.stepCountdown > 0 {
			m.stepCountdown--
			if m.stepCountdown == 0 {
				m.stepMessage = "Step completed"
				m.stepRunning = false
			}
		}
		return m, tea.Batch(
			listenForFires(m.fireEvents),
			tickCmd(),
		)
	}

	return m, nil
}

func (m *Model) updateState() {
	now := time.Now()

	// Update neuron state based on recent activity
	if !m.lastFired.IsZero() && now.Sub(m.lastFired) < 1*time.Second {
		m.neuronState = "REFRACTORY"
	} else if m.inputActive {
		m.neuronState = "ACTIVE"
	} else {
		m.neuronState = "RESTING"
	}

	// Decay visual indicators
	if m.inputActive {
		m.inputActive = false // Turn off after one tick
	}
	if m.outputActive {
		m.outputActive = false // Turn off after one tick
	}

	// Simulate accumulator decay for display
	m.accumulator *= 0.95
	if m.accumulator < 0.01 {
		m.accumulator = 0
	}
}

func (m Model) executeStep1() (Model, tea.Cmd) {
	m.addMessage("STEP 1: Single weak signal")
	m.stepRunning = true
	m.stepCountdown = 10 // 2 seconds
	m.stepMessage = "Sending 0.5 signal..."

	input := m.neuron.GetInput()
	input <- neuron.Message{Value: 0.5}
	m.lastInput = 0.5
	m.inputActive = true
	m.neuronState = "ACTIVE"
	m.accumulator = 0.5

	m.step = 2
	return m, nil
}

func (m Model) executeStep2() (Model, tea.Cmd) {
	m.addMessage("STEP 2: Temporal summation")
	m.stepRunning = true
	m.stepCountdown = 25 // 5 seconds total
	m.stepMessage = "Sending first signal..."

	// Send first signal immediately
	input := m.neuron.GetInput()
	input <- neuron.Message{Value: 0.4}
	m.lastInput = 0.4
	m.inputActive = true
	m.neuronState = "ACTIVE"
	m.accumulator = 0.4

	m.step = 3

	// Schedule second and third signals
	return m, tea.Batch(
		stepAction("Second signal", 0.3, 2*time.Second),
		stepAction("Third signal", 0.4, 4*time.Second),
	)
}

func (m Model) executeStep3() (Model, tea.Cmd) {
	m.addMessage("STEP 3: Leaky integration")
	m.stepRunning = true
	m.stepCountdown = 50 // 10 seconds total
	m.stepMessage = "Sending first signal..."

	// Send first signal immediately
	input := m.neuron.GetInput()
	input <- neuron.Message{Value: 0.8}
	m.lastInput = 0.8
	m.inputActive = true
	m.neuronState = "ACTIVE"
	m.accumulator = 0.8

	m.step = 4

	// Schedule second signal after long delay
	return m, stepAction("Second signal (after decay)", 0.3, 8*time.Second)
}

func (m *Model) addMessage(msg string) {
	timestamp := time.Now().Format("15:04:05")
	fullMsg := fmt.Sprintf("[%s] %s", timestamp, msg)

	m.messages = append([]string{fullMsg}, m.messages...)

	if len(m.messages) > 15 {
		m.messages = m.messages[:15]
	}
}

func (m Model) renderLeftPanel() string {
	var content strings.Builder

	content.WriteString(titleStyle.Render("🧠 Leaky Integration Test") + "\n\n")

	// Current step with progress
	content.WriteString(stepStyle.Render("Current Step:") + "\n")
	if m.stepRunning {
		spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spinnerChar := spinner[m.stepCountdown%len(spinner)]
		content.WriteString(fmt.Sprintf("%s %s (%ds remaining)\n\n", spinnerChar, m.stepMessage, m.stepCountdown/5))
	} else {
		content.WriteString("Ready to run next step\n\n")
	}

	// Steps
	content.WriteString(stepStyle.Render("Experiment Steps:") + "\n")
	content.WriteString("1️⃣  Single weak signal (0.5)\n")
	content.WriteString("2️⃣  Three signals: 0.4, wait 2s, 0.3, wait 2s, 0.4\n")
	content.WriteString("3️⃣  Two signals: 0.8, wait 8s, 0.3\n\n")

	// Controls
	content.WriteString(warningStyle.Render("Controls:") + "\n")
	content.WriteString("• Press '1' - Run step 1\n")
	content.WriteString("• Press '2' - Run step 2\n")
	content.WriteString("• Press '3' - Run step 3\n")
	content.WriteString("• Press 't' - Test fire (1.5)\n")
	content.WriteString("• Press 'r' - Restart\n")
	content.WriteString("• Press 'q' - Quit\n\n")

	// Parameters
	content.WriteString(dimStyle.Render("Parameters:") + "\n")
	content.WriteString(dimStyle.Render("• Threshold: 1.0") + "\n")
	content.WriteString(dimStyle.Render("• Decay: 0.95") + "\n")
	content.WriteString(dimStyle.Render("• Refractory: 1 second") + "\n\n")

	content.WriteString(fmt.Sprintf("Total fires: %d", m.totalFires))

	return content.String()
}

func (m Model) renderRightPanel() string {
	var content strings.Builder

	content.WriteString(titleStyle.Render("🔬 Real-Time Visualization") + "\n\n")

	// Fixed neuron visualization table
	content.WriteString("┌────────────────────────────────────────────┐\n")
	content.WriteString("│               Neuron State                 │\n")
	content.WriteString("├────────────────────────────────────────────┤\n")

	// Get symbols
	inputSymbol := "○"
	if m.inputActive {
		inputSymbol = "📡"
	}

	var neuronSymbol string
	var neuronColor lipgloss.Style
	switch m.neuronState {
	case "FIRING":
		neuronSymbol = "🔥"
		neuronColor = firingStyle
	case "ACTIVE":
		neuronSymbol = "🟡"
		neuronColor = activeStyle
	case "REFRACTORY":
		neuronSymbol = "🟣"
		neuronColor = refractoryStyle
	default:
		neuronSymbol = "⚫"
		neuronColor = restingStyle
	}

	outputSymbol := "○"
	if m.outputActive {
		outputSymbol = "⚡"
	}

	// Fixed layout line
	visualLine := fmt.Sprintf("│     %s  ──→  %s  ──→  %s              │",
		inputSymbol, neuronColor.Render(neuronSymbol), outputSymbol)
	content.WriteString(visualLine + "\n")

	// State info
	content.WriteString("├────────────────────────────────────────────┤\n")
	stateInfo := fmt.Sprintf("│ State: %-10s  Last Input: %.1f      │", m.neuronState, m.lastInput)
	content.WriteString(stateInfo + "\n")
	accumInfo := fmt.Sprintf("│ Accumulator: %.2f / 1.0 (threshold)    │", m.accumulator)
	content.WriteString(accumInfo + "\n")
	content.WriteString("└────────────────────────────────────────────┘\n\n")

	// Legend
	content.WriteString("Legend:\n")
	content.WriteString("📡 Input   🔥 Firing   🟡 Active   🟣 Refractory   ⚫ Resting\n")
	content.WriteString("⚡ Output   ○ Inactive\n\n")

	// Activity log
	content.WriteString(warningStyle.Render("Activity Log:") + "\n")
	content.WriteString(strings.Repeat("─", 45) + "\n")

	for i, msg := range m.messages {
		if i >= 12 {
			break
		}

		if len(msg) > 42 {
			msg = msg[:39] + "..."
		}

		if strings.Contains(msg, "FIRED") {
			content.WriteString(successStyle.Render(msg) + "\n")
		} else if strings.Contains(msg, "STEP") {
			content.WriteString(stepStyle.Render(msg) + "\n")
		} else {
			content.WriteString(msg + "\n")
		}
	}

	return content.String()
}

func (m Model) View() string {
	leftWidth := m.width/2 - 3
	rightWidth := m.width/2 - 3
	panelHeight := m.height - 4

	leftBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2).
		Width(leftWidth).
		Height(panelHeight)

	rightBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(rightWidth).
		Height(panelHeight)

	leftPanel := leftBoxStyle.Render(m.renderLeftPanel())
	rightPanel := rightBoxStyle.Render(m.renderRightPanel())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
	}
}
