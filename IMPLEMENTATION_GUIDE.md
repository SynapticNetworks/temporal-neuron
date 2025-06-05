# Complete Implementation Guide: Biological Neural Computation

> **Detailed Implementation Roadmap**  
> **File Location:** `docs/COMPLETE_IMPLEMENTATION_GUIDE.md`  
> **Companion to:** `IMPLEMENTATION_PLAN.md` (high-level overview)

This document provides the **complete, detailed implementation guide** with specific experiments, unit tests, visualizations, and code for each phase of biological neural computation development.

---

## 🔬 Phase 1: Biological Basis Validation
*"Proving your current neurons exhibit genuine biological behaviors"*

> **🎯 Goal:** Validate foundational biological behaviors before adding learning mechanisms.

### **📁 Experiment Structure**

```
experiments/phase-1-basis/
├── README.md                     ← Phase overview and quick start
├── main.go                       ← Interactive experiment runner
├── experiments/
│   ├── 1-leaky-integration/
│   │   ├── experiment.go
│   │   ├── README.md
│   │   └── expected_output.txt
│   ├── 2-refractory-period/
│   ├── 3-synaptic-delays/
│   ├── 4-excitation-inhibition/
│   └── 5-network-propagation/
├── common/
│   ├── visualization.go          ← ASCII visualizations
│   ├── metrics.go               ← Success criteria validation
│   └── test_networks.go         ← Standard test configurations
└── results/                     ← Output logs and data
```

### **🧪 Experiment 1.1: Leaky Integration & Temporal Summation**

**Biological Principle:** Weak signals arriving close together sum up (temporal summation), but signals separated by time don't sum because membrane potential decays.

**Unit Test Implementation:**

```go
// File: experiments/phase-1-basis/experiments/1-leaky-integration/experiment.go
package leaky_integration

import (
    "testing"
    "time"
    "github.com/SynapticNetworks/temporal-neuron/neuron"
)

func TestLeakyIntegrationSummation(t *testing.T) {
    n := neuron.NewNeuron("test_leaky", 1.0, 0.95, 10*time.Millisecond, 1.0)
    output := make(chan neuron.Message, 10)
    n.AddOutput("test", output, 1.0, 0)
    
    go n.Run()
    defer n.Close()
    
    input := n.GetInput()
    
    // Test 1: Single weak signal should NOT fire
    input <- neuron.Message{Value: 0.6}
    select {
    case <-output:
        t.Error("Single weak signal should not fire")
    case <-time.After(20 * time.Millisecond):
        // Expected: no firing
    }
    
    // Test 2: Two quick weak signals should fire (summation)
    input <- neuron.Message{Value: 0.6}
    time.Sleep(2 * time.Millisecond) // Quick succession
    input <- neuron.Message{Value: 0.6} // Total: ~1.2 > 1.0
    
    select {
    case <-output:
        // Expected: neuron fires from temporal summation
    case <-time.After(20 * time.Millisecond):
        t.Error("Two quick signals should sum and fire")
    }
    
    // Test 3: Two slow weak signals should NOT fire (decay)
    time.Sleep(100 * time.Millisecond) // Reset
    input <- neuron.Message{Value: 0.6}
    time.Sleep(50 * time.Millisecond) // Long delay for decay
    input <- neuron.Message{Value: 0.6} // First signal decayed
    
    select {
    case <-output:
        t.Error("Slow signals should not fire due to decay")
    case <-time.After(20 * time.Millisecond):
        // Expected: no firing due to decay
    }
}

func TestLeakyIntegrationDecayRate(t *testing.T) {
    slowDecay := neuron.NewNeuron("slow", 1.0, 0.99, 5*time.Millisecond, 1.0)
    fastDecay := neuron.NewNeuron("fast", 1.0, 0.90, 5*time.Millisecond, 1.0)
    
    // Test that decay rate affects temporal summation window
    // Implementation details...
}

func TestContinuousMembraneDecay(t *testing.T) {
    n := neuron.NewNeuron("decay_test", 2.0, 0.9, 5*time.Millisecond, 1.0)
    
    go n.Run()
    defer n.Close()
    
    // Send sub-threshold signal
    n.GetInput() <- neuron.Message{Value: 1.0} // Below threshold of 2.0
    
    // Wait for significant decay
    time.Sleep(50 * time.Millisecond)
    
    // Add small signal - should not fire if decay worked
    n.GetInput() <- neuron.Message{Value: 0.8} // Total would be 1.8 without decay
    
    output := make(chan neuron.Message, 10)
    n.AddOutput("test", output, 1.0, 0)
    
    select {
    case <-output:
        t.Error("Signal should have decayed below firing threshold")
    case <-time.After(20 * time.Millisecond):
        // Expected: no firing due to decay
    }
}
```

**Interactive Visualization:**

```go
// File: experiments/phase-1-basis/common/visualization.go
package common

import (
    "fmt"
    "strings"
    "time"
)

type MembraneVisualizer struct {
    threshold    float64
    membrane     float64
    maxWidth     int
    timeStep     int
}

func NewMembraneVisualizer(threshold float64) *MembraneVisualizer {
    return &MembraneVisualizer{
        threshold: threshold,
        maxWidth:  50,
    }
}

func (mv *MembraneVisualizer) Update(membrane float64, signal float64) {
    mv.membrane = membrane
    mv.timeStep++
    
    // Create visual bar representation
    barLength := int((membrane / (mv.threshold * 1.5)) * float64(mv.maxWidth))
    if barLength > mv.maxWidth {
        barLength = mv.maxWidth
    }
    
    bar := strings.Repeat("█", barLength)
    threshold_mark := int((mv.threshold / (mv.threshold * 1.5)) * float64(mv.maxWidth))
    
    // Insert threshold marker
    display := fmt.Sprintf("%-50s", bar)
    if threshold_mark < len(display) {
        runes := []rune(display)
        runes[threshold_mark] = '|'
        display = string(runes)
    }
    
    status := "charging"
    if membrane >= mv.threshold {
        status = "🔥 FIRE!"
    } else if signal == 0 {
        status = "decaying..."
    }
    
    fmt.Printf("T=%03d [%s] Signal:%.1f Membrane:%.2f → %s\n", 
        mv.timeStep, display, signal, membrane, status)
}

func RunLeakyIntegrationDemo() {
    fmt.Println("🧠 Leaky Integration Demo")
    fmt.Println("========================")
    fmt.Println("Threshold: 1.0, Decay Rate: 0.95")
    fmt.Println()
    
    viz := NewMembraneVisualizer(1.0)
    
    fmt.Println("Step 1: Single weak signal (0.6)")
    viz.Update(0.6, 0.6)
    viz.Update(0.57, 0.0) // Decay
    viz.Update(0.54, 0.0) // Decay
    fmt.Println()
    
    fmt.Println("Step 2: Two quick signals (0.6 + 0.6)")  
    viz.Update(0.6, 0.6)
    viz.Update(1.2, 0.6)  // Should fire!
    fmt.Println()
    
    fmt.Println("Step 3: Two slow signals (0.6 ... wait ... 0.6)")
    viz.Update(0.6, 0.6)
    viz.Update(0.3, 0.0)  // Significant decay
    viz.Update(0.9, 0.6)  // Second signal, but total < threshold
    fmt.Println()
}
```

**Expected Output Visualization:**

```
🧠 Leaky Integration Demo
========================
Threshold: 1.0, Decay Rate: 0.95

Step 1: Single weak signal (0.6)
T=001 [████████████████████         |                         ] Signal:0.6 Membrane:0.60 → charging
T=002 [███████████████████          |                         ] Signal:0.0 Membrane:0.57 → decaying...
T=003 [██████████████████           |                         ] Signal:0.0 Membrane:0.54 → decaying...

Step 2: Two quick signals (0.6 + 0.6)
T=004 [████████████████████         |                         ] Signal:0.6 Membrane:0.60 → charging
T=005 [████████████████████████████████████████|            ] Signal:0.6 Membrane:1.20 → 🔥 FIRE!

Step 3: Two slow signals (0.6 ... wait ... 0.6)
T=006 [████████████████████         |                         ] Signal:0.6 Membrane:0.60 → charging
T=007 [██████████                   |                         ] Signal:0.0 Membrane:0.30 → decayed!
T=008 [██████████████████████████████|                       ] Signal:0.6 Membrane:0.90 → charging (no fire)
```

**Success Criteria:**
- [ ] Single weak signal (0.6) does NOT fire
- [ ] Two quick weak signals (0.6 + 0.6) DO fire  
- [ ] Two slow weak signals (0.6, wait, 0.6) do NOT fire
- [ ] Membrane potential continuously decays without input
- [ ] Decay rate affects summation window duration

### **🧪 Experiment 1.2: Refractory Period Enforcement**

**Biological Principle:** After firing, neurons enter a recovery period where they cannot fire again, regardless of input strength.

**Unit Test Implementation:**

```go
// File: experiments/phase-1-basis/experiments/2-refractory-period/experiment.go
package refractory_period

func TestRefractoryPeriodPreventsRapidFiring(t *testing.T) {
    refractoryPeriod := 20 * time.Millisecond
    n := neuron.NewNeuron("refractory_test", 1.0, 0.95, refractoryPeriod, 1.0)
    output := make(chan neuron.Message, 10)
    n.AddOutput("test", output, 1.0, 0)
    
    go n.Run()
    defer n.Close()
    
    input := n.GetInput()
    
    // Fire first time
    input <- neuron.Message{Value: 1.5}
    select {
    case <-output:
        // Expected first firing
    case <-time.After(10 * time.Millisecond):
        t.Fatal("First firing should occur")
    }
    
    // Immediately try to fire again (should be blocked)
    input <- neuron.Message{Value: 2.0} // Strong signal
    select {
    case <-output:
        t.Error("Neuron fired during refractory period")
    case <-time.After(10 * time.Millisecond):
        // Expected: blocked by refractory period
    }
    
    // Wait for refractory period to end
    time.Sleep(refractoryPeriod + 5*time.Millisecond)
    input <- neuron.Message{Value: 1.5}
    select {
    case <-output:
        // Expected: can fire again
    case <-time.After(10 * time.Millisecond):
        t.Error("Neuron should fire after refractory period")
    }
}

func TestRefractoryPeriodDuration(t *testing.T) {
    shortRefractory := 10 * time.Millisecond
    longRefractory := 50 * time.Millisecond
    
    // Test that different refractory periods behave correctly
    // Implementation details...
}

func TestRefractoryPeriodIgnoresInputStrength(t *testing.T) {
    n := neuron.NewNeuron("strength_test", 1.0, 0.95, 20*time.Millisecond, 1.0)
    
    // Test that even very strong inputs are blocked during refractory period
    // Implementation details...
}
```

**Interactive Visualization:**

```go
func RunRefractoryPeriodDemo() {
    fmt.Println("🧠 Refractory Period Demo")
    fmt.Println("=========================")
    fmt.Println("Threshold: 1.0, Refractory: 20ms")
    fmt.Println()
    
    states := []string{
        "READY", "REFRACTORY", "REFRACTORY", "REFRACTORY", "READY"
    }
    
    signals := []float64{1.5, 2.0, 2.5, 1.8, 1.5}
    times := []int{0, 5, 10, 15, 25}
    
    for i, state := range states {
        refractoryBar := ""
        if state == "REFRACTORY" {
            remaining := 20 - times[i]
            refractoryBar = fmt.Sprintf("[%s%s]", 
                strings.Repeat("R", remaining/2), 
                strings.Repeat("-", 10-remaining/2))
        } else {
            refractoryBar = "[READY    ]"
        }
        
        fired := ""
        if state == "READY" && signals[i] >= 1.0 {
            fired = "🔥 FIRE!"
        } else if state == "REFRACTORY" {
            fired = "❌ BLOCKED"
        }
        
        fmt.Printf("T=%02dms %s Signal:%.1f → %s\n", 
            times[i], refractoryBar, signals[i], fired)
    }
}
```

**Expected Output:**

```
🧠 Refractory Period Demo
=========================
Threshold: 1.0, Refractory: 20ms

T=00ms [READY    ] Signal:1.5 → 🔥 FIRE!
T=05ms [RRRRRRRRR-] Signal:2.0 → ❌ BLOCKED
T=10ms [RRRRR----] Signal:2.5 → ❌ BLOCKED  
T=15ms [RRR------] Signal:1.8 → ❌ BLOCKED
T=25ms [READY    ] Signal:1.5 → 🔥 FIRE!
```

**Success Criteria:**
- [ ] First strong signal fires normally
- [ ] Immediate second signal is blocked (regardless of strength)
- [ ] Signals during refractory period are all blocked
- [ ] After refractory period, neuron can fire again
- [ ] Refractory duration matches specified parameter

### **🧪 Experiment 1.3: Synaptic Delays and Transmission**

**Biological Principle:** Signals don't travel instantaneously - axon length and synaptic processing create realistic transmission delays.

**Unit Test Implementation:**

```go
// File: experiments/phase-1-basis/experiments/3-synaptic-delays/experiment.go
package synaptic_delays

func TestSynapticDelayTiming(t *testing.T) {
    source := neuron.NewNeuron("source", 1.0, 0.95, 5*time.Millisecond, 2.0)
    target := make(chan neuron.Message, 10)
    
    delay := 15 * time.Millisecond
    factor := 0.5
    source.AddOutput("delayed", target, factor, delay)
    
    go source.Run()
    defer source.Close()
    
    // Fire source neuron
    startTime := time.Now()
    source.GetInput() <- neuron.Message{Value: 1.5}
    
    // Wait for delayed signal
    select {
    case msg := <-target:
        elapsed := time.Since(startTime)
        
        // Check delay timing (±5ms tolerance)
        if elapsed < delay-5*time.Millisecond || elapsed > delay+5*time.Millisecond {
            t.Errorf("Delay incorrect: expected %v±5ms, got %v", delay, elapsed)
        }
        
        // Check signal strength: input(1.5) * fireFactor(2.0) * synapticFactor(0.5) = 1.5
        expected := 1.5 * 2.0 * 0.5
        if msg.Value != expected {
            t.Errorf("Signal strength incorrect: expected %f, got %f", expected, msg.Value)
        }
        
    case <-time.After(delay + 50*time.Millisecond):
        t.Error("Signal never arrived")
    }
}

func TestMultipleSynapticDelays(t *testing.T) {
    source := neuron.NewNeuron("source", 1.0, 0.95, 5*time.Millisecond, 1.0)
    
    fast := make(chan neuron.Message, 10)
    medium := make(chan neuron.Message, 10)
    slow := make(chan neuron.Message, 10)
    
    source.AddOutput("fast", fast, 1.0, 5*time.Millisecond)
    source.AddOutput("medium", medium, 1.0, 15*time.Millisecond)
    source.AddOutput("slow", slow, 1.0, 30*time.Millisecond)
    
    go source.Run()
    defer source.Close()
    
    // Fire once, expect three arrivals at different times
    startTime := time.Now()
    source.GetInput() <- neuron.Message{Value: 1.5}
    
    // Check arrival order and timing
    // Implementation details...
}

func TestSynapticFactorModulation(t *testing.T) {
    source := neuron.NewNeuron("source", 1.0, 0.95, 5*time.Millisecond, 1.0)
    
    // Test different synaptic strengths
    weak := make(chan neuron.Message, 10)
    strong := make(chan neuron.Message, 10)
    
    source.AddOutput("weak", weak, 0.2, 10*time.Millisecond)
    source.AddOutput("strong", strong, 1.8, 10*time.Millisecond)
    
    // Verify signal strengths are correctly modulated
    // Implementation details...
}
```

**Interactive Visualization:**

```go
func RunSynapticDelayDemo() {
    fmt.Println("🧠 Synaptic Transmission Demo")
    fmt.Println("=============================")
    fmt.Println()
    
    fmt.Println("Source Neuron → [Delays] → Target Neurons")
    fmt.Println("Input: 1.5, Fire Factor: 2.0")
    fmt.Println()
    
    delays := []struct{
        name string
        delay int
        factor float64
    }{
        {"Fast  ", 5, 1.0},
        {"Medium", 15, 0.8}, 
        {"Slow  ", 30, 0.6},
    }
    
    fmt.Println("T=0ms:  Source fires! Signal strength: 1.5 * 2.0 = 3.0")
    fmt.Println()
    
    for _, d := range delays {
        expectedStrength := 3.0 * d.factor
        
        // Show transmission progress
        for t := 0; t <= d.delay; t += 5 {
            progress := float64(t) / float64(d.delay)
            progressBar := strings.Repeat("█", int(progress*10))
            remaining := strings.Repeat("░", 10-int(progress*10))
            
            if t == d.delay {
                fmt.Printf("T=%02dms: %s [%s%s] → Arrived! Strength: %.1f ✅\n", 
                    t, d.name, progressBar, remaining, expectedStrength)
            } else if t % 10 == 0 {
                fmt.Printf("T=%02dms: %s [%s%s] → traveling...\n", 
                    t, d.name, progressBar, remaining)
            }
        }
        fmt.Println()
    }
}
```

**Expected Output:**

```
🧠 Synaptic Transmission Demo
=============================

Source Neuron → [Delays] → Target Neurons
Input: 1.5, Fire Factor: 2.0

T=0ms:  Source fires! Signal strength: 1.5 * 2.0 = 3.0

T=00ms: Fast   [░░░░░░░░░░] → traveling...
T=05ms: Fast   [██████████] → Arrived! Strength: 3.0 ✅

T=00ms: Medium [░░░░░░░░░░] → traveling...
T=10ms: Medium [██████░░░░] → traveling...
T=15ms: Medium [██████████] → Arrived! Strength: 2.4 ✅

T=00ms: Slow   [░░░░░░░░░░] → traveling...
T=10ms: Slow   [███░░░░░░░] → traveling...
T=20ms: Slow   [██████░░░░] → traveling...
T=30ms: Slow   [██████████] → Arrived! Strength: 1.8 ✅
```

**Success Criteria:**
- [ ] Signals arrive with correct delay timing (±5ms tolerance)
- [ ] Signal strengths correctly modified by synaptic factors
- [ ] Multiple outputs fire in parallel with independent delays
- [ ] Transmission delays don't interfere with each other
- [ ] No signals lost during transmission

### **🧪 Experiment 1.4: Excitatory/Inhibitory Balance**

**Biological Principle:** Neurons receive both excitatory (+) and inhibitory (-) inputs. The balance determines firing probability.

**Unit Test Implementation:**

```go
// File: experiments/phase-1-basis/experiments/4-excitation-inhibition/experiment.go
package excitation_inhibition

func TestExcitatoryInhibitoryBalance(t *testing.T) {
    n := neuron.NewNeuron("balance_test", 1.0, 0.98, 5*time.Millisecond, 1.0)
    output := make(chan neuron.Message, 10)
    n.AddOutput("test", output, 1.0, 0)
    
    go n.Run()
    defer n.Close()
    
    input := n.GetInput()
    
    // Test 1: Excitation builds up
    input <- neuron.Message{Value: 0.8}  // Below threshold
    input <- neuron.Message{Value: 0.5}  // Total: 1.3 > threshold
    
    // But then inhibition intervenes
    input <- neuron.Message{Value: -0.5} // Total back to 0.8 < threshold
    
    select {
    case <-output:
        t.Error("Neuron should not fire due to inhibition")
    case <-time.After(30 * time.Millisecond):
        // Expected: inhibition prevented firing
    }
    
    // Test 2: Excitation overcomes inhibition
    input <- neuron.Message{Value: 0.4} // Total: 1.2 > threshold
    
    select {
    case <-output:
        // Expected: excitation overcomes inhibition
    case <-time.After(20 * time.Millisecond):
        t.Error("Neuron should fire when excitation overcomes inhibition")
    }
}

func TestPureInhibition(t *testing.T) {
    n := neuron.NewNeuron("inhibit_test", 1.0, 0.98, 5*time.Millisecond, 1.0)
    
    // Test that pure inhibitory input cannot cause firing
    n.GetInput() <- neuron.Message{Value: -2.0} // Strong inhibition
    
    // Should never fire from negative input
    // Implementation details...
}

func TestInhibitionStrength(t *testing.T) {
    n := neuron.NewNeuron("strength_test", 1.0, 0.98, 5*time.Millisecond, 1.0)
    
    // Test different levels of inhibition
    // Implementation details...
}
```

**Interactive Visualization:**

```go
func RunExcitationInhibitionDemo() {
    fmt.Println("🧠 Excitation/Inhibition Balance Demo")
    fmt.Println("====================================")
    fmt.Println("Threshold: 1.0")
    fmt.Println()
    
    membrane := 0.0
    threshold := 1.0
    
    signals := []struct{
        value float64
        desc string
    }{
        {0.8, "Excitatory (+0.8)"},
        {0.5, "Excitatory (+0.5)"},
        {-0.5, "Inhibitory (-0.5)"},
        {0.4, "Excitatory (+0.4)"},
    }
    
    for i, sig := range signals {
        membrane += sig.value
        
        // Visual representation of excitation vs inhibition
        excitation := 0.0
        inhibition := 0.0
        if sig.value > 0 {
            excitation = sig.value
        } else {
            inhibition = -sig.value
        }
        
        excBar := strings.Repeat("█", int(excitation*10))
        inhBar := strings.Repeat("▓", int(inhibition*10))
        
        status := "building..."
        if membrane >= threshold {
            status = "🔥 FIRE!"
        } else if membrane < 0 {
            status = "suppressed"
            membrane = 0 // Can't go below 0
        }
        
        fmt.Printf("Step %d: %s\n", i+1, sig.desc)
        fmt.Printf("  Excitation: [%s]\n", fmt.Sprintf("%-10s", excBar))
        fmt.Printf("  Inhibition: [%s]\n", fmt.Sprintf("%-10s", inhBar))
        fmt.Printf("  Membrane: %.1f/%.1f → %s\n", membrane, threshold, status)
        fmt.Println()
        
        if membrane >= threshold {
            membrane = 0 // Reset after firing
        }
    }
}
```

**Expected Output:**

```
🧠 Excitation/Inhibition Balance Demo
====================================
Threshold: 1.0

Step 1: Excitatory (+0.8)
  Excitation: [████████  ]
  Inhibition: [          ]
  Membrane: 0.8/1.0 → building...

Step 2: Excitatory (+0.5)  
  Excitation: [█████     ]
  Inhibition: [          ]
  Membrane: 1.3/1.0 → would fire, but...

Step 3: Inhibitory (-0.5)
  Excitation: [          ]
  Inhibition: [█████     ]
  Membrane: 0.8/1.0 → building...

Step 4: Excitatory (+0.4)
  Excitation: [████      ]
  Inhibition: [          ] 
  Membrane: 1.2/1.0 → 🔥 FIRE!
```

**Success Criteria:**
- [ ] Positive signals increase membrane potential
- [ ] Negative signals decrease membrane potential
- [ ] Inhibition can prevent firing even when excitation is strong
- [ ] Strong excitation can overcome moderate inhibition
- [ ] Membrane potential cannot go below zero
- [ ] Balance determines final firing decision

### **🧪 Experiment 1.5: Network Signal Propagation**

**Biological Principle:** Activity cascades through connected neurons with cumulative delays and synaptic modifications.

**Unit Test Implementation:**

```go
// File: experiments/phase-1-basis/experiments/5-network-propagation/experiment.go
package network_propagation

func TestLinearChainPropagation(t *testing.T) {
    // Create chain: A → B → C → D
    neuronA := neuron.NewNeuron("A", 1.0, 0.95, 5*time.Millisecond, 1.0)
    neuronB := neuron.NewNeuron("B", 1.0, 0.95, 5*time.Millisecond, 1.0)
    neuronC := neuron.NewNeuron("C", 1.0, 0.95, 5*time.Millisecond, 1.0)
    neuronD := neuron.NewNeuron("D", 1.0, 0.95, 5*time.Millisecond, 1.0)
    
    // Connect with 10ms delays
    neuronA.AddOutput("to_B", neuronB.GetInputChannel(), 1.2, 10*time.Millisecond)
    neuronB.AddOutput("to_C", neuronC.GetInputChannel(), 1.2, 10*time.Millisecond)
    neuronC.AddOutput("to_D", neuronD.GetInputChannel(), 1.2, 10*time.Millisecond)
    
    // Output monitoring
    outputD := make(chan neuron.Message, 10)
    neuronD.AddOutput("monitor", outputD, 1.0, 0)
    
    // Start all neurons
    go neuronA.Run()
    go neuronB.Run()
    go neuronC.Run()
    go neuronD.Run()
    
    defer func() {
        neuronA.Close()
        neuronB.Close()
        neuronC.Close()
        neuronD.Close()
    }()
    
    // Fire A and measure total propagation time
    startTime := time.Now()
    neuronA.GetInput() <- neuron.Message{Value: 1.5}
    
    select {
    case <-outputD:
        elapsed := time.Since(startTime)
        expectedDelay := 30 * time.Millisecond // 3 hops × 10ms each
        
        if elapsed < expectedDelay-5*time.Millisecond || elapsed > expectedDelay+10*time.Millisecond {
            t.Errorf("Chain propagation timing incorrect: expected ~%v, got %v", expectedDelay, elapsed)
        }
        
    case <-time.After(50 * time.Millisecond):
        t.Error("Signal failed to propagate through chain")
    }
}

func TestParallelBranchPropagation(t *testing.T) {
    // Create branching network: A → [B, C, D]
    source := neuron.NewNeuron("source", 1.0, 0.95, 5*time.Millisecond, 1.0)
    targetB := neuron.NewNeuron("B", 1.0, 0.95, 5*time.Millisecond, 1.0)
    targetC := neuron.NewNeuron("C", 1.0, 0.95, 5*time.Millisecond, 1.0)
    targetD := neuron.NewNeuron("D", 1.0, 0.95, 5*time.Millisecond, 1.0)
    
    // Connect with different delays and strengths
    source.AddOutput("to_B", targetB.GetInputChannel(), 1.2, 5*time.Millisecond)
    source.AddOutput("to_C", targetC.GetInputChannel(), 1.1, 15*time.Millisecond)
    source.AddOutput("to_D", targetD.GetInputChannel(), 1.3, 25*time.Millisecond)
    
    // Monitor all outputs
    outputB := make(chan neuron.Message, 10)
    outputC := make(chan neuron.Message, 10)
    outputD := make(chan neuron.Message, 10)
    
    targetB.AddOutput("monitor", outputB, 1.0, 0)
    targetC.AddOutput("monitor", outputC, 1.0, 0)
    targetD.AddOutput("monitor", outputD, 1.0, 0)
    
    // Start all neurons
    neurons := []*neuron.Neuron{source, targetB, targetC, targetD}
    for _, n := range neurons {
        go n.Run()
        defer n.Close()
    }
    
    // Fire source and verify all targets receive signals
    startTime := time.Now()
    source.GetInput() <- neuron.Message{Value: 1.5}
    
    received := 0
    for received < 3 {
        select {
        case <-outputB:
            elapsed := time.Since(startTime)
            if elapsed < 5*time.Millisecond-2*time.Millisecond || elapsed > 5*time.Millisecond+5*time.Millisecond {
                t.Errorf("Branch B timing incorrect: expected ~5ms, got %v", elapsed)
            }
            received++
        case <-outputC:
            elapsed := time.Since(startTime)
            if elapsed < 15*time.Millisecond-2*time.Millisecond || elapsed > 15*time.Millisecond+5*time.Millisecond {
                t.Errorf("Branch C timing incorrect: expected ~15ms, got %v", elapsed)
            }
            received++
        case <-outputD:
            elapsed := time.Since(startTime)
            if elapsed < 25*time.Millisecond-2*time.Millisecond || elapsed > 25*time.Millisecond+5*time.Millisecond {
                t.Errorf("Branch D timing incorrect: expected ~25ms, got %v", elapsed)
            }
            received++
        case <-time.After(40 * time.Millisecond):
            t.Errorf("Only received %d/3 signals", received)
            return
        }
    }
}

func TestConvergentNetworkPropagation(t *testing.T) {
    // Create convergent network: [A, B, C] → D
    sourceA := neuron.NewNeuron("A", 1.0, 0.95, 5*time.Millisecond, 1.0)
    sourceB := neuron.NewNeuron("B", 1.0, 0.95, 5*time.Millisecond, 1.0)
    sourceC := neuron.NewNeuron("C", 1.0, 0.95, 5*time.Millisecond, 1.0)
    target := neuron.NewNeuron("target", 2.0, 0.95, 5*time.Millisecond, 1.0) // Higher threshold
    
    // All sources connect to target with weak individual connections
    sourceA.AddOutput("to_target", target.GetInputChannel(), 0.8, 10*time.Millisecond)
    sourceB.AddOutput("to_target", target.GetInputChannel(), 0.7, 12*time.Millisecond)
    sourceC.AddOutput("to_target", target.GetInputChannel(), 0.9, 8*time.Millisecond)
    
    // Monitor target output
    targetOutput := make(chan neuron.Message, 10)
    target.AddOutput("monitor", targetOutput, 1.0, 0)
    
    // Start all neurons
    neurons := []*neuron.Neuron{sourceA, sourceB, sourceC, target}
    for _, n := range neurons {
        go n.Run()
        defer n.Close()
    }
    
    // Test 1: Single input should not cause target to fire
    sourceA.GetInput() <- neuron.Message{Value: 1.5}
    
    select {
    case <-targetOutput:
        t.Error("Single weak input should not fire target")
    case <-time.After(30 * time.Millisecond):
        // Expected: no firing from single input
    }
    
    // Test 2: Multiple inputs should cause firing through convergence
    sourceA.GetInput() <- neuron.Message{Value: 1.5}
    sourceB.GetInput() <- neuron.Message{Value: 1.5}
    sourceC.GetInput() <- neuron.Message{Value: 1.5}
    
    select {
    case <-targetOutput:
        // Expected: convergent inputs sum to fire target
    case <-time.After(30 * time.Millisecond):
        t.Error("Convergent inputs should fire target")
    }
}
```

**Interactive Visualization:**

```go
func RunNetworkPropagationDemo() {
    fmt.Println("🧠 Network Signal Propagation Demo")
    fmt.Println("==================================")
    fmt.Println()
    
    // Linear Chain Demo
    fmt.Println("Linear Chain: A → B → C → D")
    fmt.Println("Connection delays: 10ms each")
    fmt.Println()
    
    chain := []string{"A", "B", "C", "D"}
    delays := []int{0, 10, 20, 30}
    
    fmt.Println("T=0ms:  A fires! 🔥")
    fmt.Println()
    
    for i, neuron := range chain {
        if i == 0 {
            continue // A already fired
        }
        
        fmt.Printf("T=%dms: Signal reaches %s", delays[i], neuron)
        
        // Show propagation path
        path := ""
        for j := 0; j <= i; j++ {
            if j == i {
                path += chain[j] + "🔥"
            } else {
                path += chain[j] + "→"
            }
        }
        
        fmt.Printf(" [%s]\n", path)
        
        if i < len(chain)-1 {
            fmt.Printf("        %s fires, signal continues...\n", neuron)
        } else {
            fmt.Printf("        Final destination reached! ✅\n")
        }
        fmt.Println()
    }
    
    // Branching Network Demo
    fmt.Println("Parallel Branches: A → [B(5ms), C(15ms), D(25ms)]")
    fmt.Println()
    
    branches := []struct{
        name string
        delay int
    }{
        {"B", 5},
        {"C", 15}, 
        {"D", 25},
    }
    
    fmt.Println("T=0ms:  A fires! Signal splits three ways 🔥")
    fmt.Println()
    
    for _, branch := range branches {
        fmt.Printf("T=%dms: Branch %s fires! 🔥\n", branch.delay, branch.name)
    }
    
    fmt.Println()
    fmt.Println("All branches fire in parallel with different timing! ✅")
}

func RunConvergentNetworkDemo() {
    fmt.Println("🧠 Convergent Network Demo")
    fmt.Println("==========================")
    fmt.Println("Network: [A, B, C] → Target (threshold: 2.0)")
    fmt.Println("Individual connections: 0.8, 0.7, 0.9")
    fmt.Println()
    
    fmt.Println("Test 1: Single input")
    fmt.Printf("A fires → 0.8 reaches target → No fire (0.8 < 2.0) ❌\n")
    fmt.Println()
    
    fmt.Println("Test 2: Convergent inputs")  
    fmt.Printf("A fires → 0.8 to target\n")
    fmt.Printf("B fires → 0.7 to target  \n")
    fmt.Printf("C fires → 0.9 to target\n")
    fmt.Printf("Total: 0.8 + 0.7 + 0.9 = 2.4 > 2.0 → Target fires! 🔥✅\n")
    fmt.Println()
    
    fmt.Println("Convergent summation enables target firing! 🎯")
}
```

**Expected Output:**

```
🧠 Network Signal Propagation Demo
==================================

Linear Chain: A → B → C → D
Connection delays: 10ms each

T=0ms:  A fires! 🔥

T=10ms: Signal reaches B [A→B🔥]
        B fires, signal continues...

T=20ms: Signal reaches C [A→B→C🔥]
        C fires, signal continues...

T=30ms: Signal reaches D [A→B→C→D🔥]
        Final destination reached! ✅

Parallel Branches: A → [B(5ms), C(15ms), D(25ms)]

T=0ms:  A fires! Signal splits three ways 🔥

T=5ms:  Branch B fires! 🔥
T=15ms: Branch C fires! 🔥
T=25ms: Branch D fires! 🔥

All branches fire in parallel with different timing! ✅
```

**Success Criteria:**
- [ ] Linear chains propagate with cumulative delays
- [ ] Parallel branches fire independently with correct timing
- [ ] Convergent networks sum inputs correctly
- [ ] Signal strength preserved through propagation
- [ ] Complex topologies work as expected

---

## 🏗️ Phase 2: Homeostatic Plasticity
*"Teaching neurons to self-regulate their activity"*

> **🎯 Goal:** Add biological homeostasis so neurons automatically maintain stable activity levels.

### **📁 Experiment Structure**

```
experiments/phase-2-homeostasis/
├── README.md
├── main.go
├── experiments/
│   ├── 1-activity-tracking/
│   ├── 2-threshold-adaptation/
│   ├── 3-hyperactive-regulation/
│   ├── 4-silent-activation/
│   └── 5-network-stabilization/
├── common/
│   ├── homeostatic_neuron.go    ← Extended neuron with homeostasis
│   ├── activity_metrics.go      ← Activity measurement tools
│   └── stability_analysis.go    ← Network stability metrics
└── results/
```

### **Implementation: Homeostatic Neuron Extension**

```go
// File: experiments/phase-2-homeostasis/common/homeostatic_neuron.go
package common

import (
    "sync"
    "time"
    "github.com/SynapticNetworks/temporal-neuron/neuron"
)

type HomeostaticNeuron struct {
    *neuron.Neuron // Embed base neuron
    
    // Homeostatic regulation
    baselineThreshold  float64      // Original threshold value
    activityWindow     []time.Time  // Recent firing times (circular buffer)
    targetFireRate     float64      // Desired fires per second (e.g., 5.0)
    adaptationRate     float64      // How fast to adapt (e.g., 0.001)
    lastAdaptation     time.Time    // When we last updated threshold
    
    // Calcium-based activity tracking (biological model)
    calciumLevel       float64      // Current calcium concentration
    calciumDecayRate   float64      // How fast calcium clears (0.995)
    calciumInflux      float64      // Calcium added per spike (1.0)
    
    // Thread safety
    homeoMutex         sync.Mutex   // Protects homeostatic state
}

func NewHomeostaticNeuron(id string, threshold float64, decayRate float64, 
                         refractoryPeriod time.Duration, fireFactor float64,
                         targetFireRate float64) *HomeostaticNeuron {
    
    baseNeuron := neuron.NewNeuron(id, threshold, decayRate, refractoryPeriod, fireFactor)
    
    return &HomeostaticNeuron{
        Neuron:            baseNeuron,
        baselineThreshold: threshold,
        targetFireRate:    targetFireRate,
        adaptationRate:    0.001,
        calciumDecayRate:  0.995,
        calciumInflux:     1.0,
        activityWindow:    make([]time.Time, 0, 100),
    }
}

// Override the firing method to include homeostatic tracking
func (hn *HomeostaticNeuron) recordFiring(fireTime time.Time) {
    hn.homeoMutex.Lock()
    defer hn.homeoMutex.Unlock()
    
    // Add to activity window
    hn.activityWindow = append(hn.activityWindow, fireTime)
    
    // Keep only last second of activity
    cutoff := fireTime.Add(-1 * time.Second)
    for len(hn.activityWindow) > 0 && hn.activityWindow[0].Before(cutoff) {
        hn.activityWindow = hn.activityWindow[1:]
    }
    
    // Calcium influx with each spike (biological model)
    hn.calciumLevel += hn.calciumInflux
}

func (hn *HomeostaticNeuron) getCurrentFireRate() float64 {
    hn.homeoMutex.Lock()
    defer hn.homeoMutex.Unlock()
    
    if len(hn.activityWindow) < 2 {
        return 0.0
    }
    
    timeSpan := hn.activityWindow[len(hn.activityWindow)-1].Sub(hn.activityWindow[0])
    if timeSpan == 0 {
        return 0.0
    }
    
    return float64(len(hn.activityWindow)) / timeSpan.Seconds()
}

// Call this periodically (e.g., every 100ms) to update homeostasis
func (hn *HomeostaticNeuron) updateHomeostasis() {
    if time.Since(hn.lastAdaptation) < 100*time.Millisecond {
        return
    }
    
    hn.homeoMutex.Lock()
    defer hn.homeoMutex.Unlock()
    
    // Update calcium decay
    hn.calciumLevel *= hn.calciumDecayRate
    
    // Calculate current firing rate
    currentRate := hn.getCurrentFireRateUnsafe()
    rateError := currentRate - hn.targetFireRate
    
    // Adjust threshold based on activity (calcium-modulated)
    calciumFactor := 1.0 + (hn.calciumLevel * 0.001) // More calcium = higher threshold
    thresholdChange := rateError * hn.adaptationRate * calciumFactor
    
    // Apply threshold adaptation
    newThreshold := hn.baselineThreshold + thresholdChange
    
    // Keep threshold in reasonable bounds
    if newThreshold < 0.1 { newThreshold = 0.1 }
    if newThreshold > 5.0 { newThreshold = 5.0 }
    
    // Update neuron threshold (requires access to base neuron internals)
    hn.setThreshold(newThreshold)
    
    hn.lastAdaptation = time.Now()
}

func (hn *HomeostaticNeuron) getCurrentFireRateUnsafe() float64 {
    if len(hn.activityWindow) < 2 {
        return 0.0
    }
    
    timeSpan := hn.activityWindow[len(hn.activityWindow)-1].Sub(hn.activityWindow[0])
    if timeSpan == 0 {
        return 0.0
    }
    
    return float64(len(hn.activityWindow)) / timeSpan.Seconds()
}

// Get current homeostatic state for monitoring
func (hn *HomeostaticNeuron) GetHomeostaticState() HomeostaticState {
    hn.homeoMutex.Lock()
    defer hn.homeoMutex.Unlock()
    
    return HomeostaticState{
        CurrentThreshold:  hn.getThreshold(),
        BaselineThreshold: hn.baselineThreshold,
        CurrentFireRate:   hn.getCurrentFireRateUnsafe(),
        TargetFireRate:    hn.targetFireRate,
        CalciumLevel:      hn.calciumLevel,
        RecentFires:       len(hn.activityWindow),
    }
}

type HomeostaticState struct {
    CurrentThreshold  float64
    BaselineThreshold float64
    CurrentFireRate   float64
    TargetFireRate    float64
    CalciumLevel      float64
    RecentFires       int
}
```

### **🧪 Experiment 2.1: Activity Tracking and Measurement**

**What to Test:**
- Firing rate calculation accuracy
- Activity window management  
- Calcium level tracking
- Real-time activity monitoring

**Unit Test Implementation:**

```go
// File: experiments/phase-2-homeostasis/experiments/1-activity-tracking/experiment.go
package activity_tracking

func TestFiringRateCalculation(t *testing.T) {
    hn := common.NewHomeostaticNeuron("rate_test", 1.0, 0.95, 10*time.Millisecond, 1.0, 5.0)
    
    go hn.Run()
    defer hn.Close()
    
    // Fire at known rate (10 fires in 1 second = 10 Hz)
    input := hn.GetInput()
    startTime := time.Now()
    
    for i := 0; i < 10; i++ {
        input <- neuron.Message{Value: 1.5}
        time.Sleep(100 * time.Millisecond) // 10 Hz rate
    }
    
    elapsed := time.Since(startTime)
    rate := hn.getCurrentFireRate()
    expectedRate := 10.0 / elapsed.Seconds()
    
    if math.Abs(rate-expectedRate) > 1.0 { // ±1 Hz tolerance
        t.Errorf("Firing rate calculation incorrect: expected ~%.1f Hz, got %.1f Hz", expectedRate, rate)
    }
}

func TestActivityWindowMaintenance(t *testing.T) {
    hn := common.NewHomeostaticNeuron("window_test", 1.0, 0.95, 5*time.Millisecond, 1.0, 5.0)
    
    // Test that activity window correctly maintains only recent activity
    // Implementation details...
}

func TestCalciumTracking(t *testing.T) {
    hn := common.NewHomeostaticNeuron("calcium_test", 1.0, 0.95, 5*time.Millisecond, 1.0, 5.0)
    
    initialCalcium := hn.GetHomeostaticState().CalciumLevel
    
    // Fire multiple times and check calcium accumulation
    input := hn.GetInput()
    for i := 0; i < 5; i++ {
        input <- neuron.Message{Value: 1.5}
        time.Sleep(10 * time.Millisecond)
    }
    
    finalCalcium := hn.GetHomeostaticState().CalciumLevel
    
    if finalCalcium <= initialCalcium {
        t.Error("Calcium should accumulate with firing activity")
    }
    
    // Wait for decay
    time.Sleep(200 * time.Millisecond)
    hn.updateHomeostasis() // Force update
    
    decayedCalcium := hn.GetHomeostaticState().CalciumLevel
    
    if decayedCalcium >= finalCalcium {
        t.Error("Calcium should decay over time without activity")
    }
}
```

**Interactive Visualization:**

```go
func RunActivityTrackingDemo() {
    fmt.Println("🧠 Activity Tracking Demo")
    fmt.Println("=========================")
    
    hn := common.NewHomeostaticNeuron("demo", 1.0, 0.95, 10*time.Millisecond, 1.0, 5.0)
    go hn.Run()
    defer hn.Close()
    
    input := hn.GetInput()
    
    fmt.Println("Firing neuron at different rates and tracking activity...")
    fmt.Println()
    
    // Phase 1: Slow firing
    fmt.Println("Phase 1: Slow firing (2 Hz)")
    for i := 0; i < 10; i++ {
        input <- neuron.Message{Value: 1.5}
        time.Sleep(500 * time.Millisecond)
        
        if i%3 == 0 {
            state := hn.GetHomeostaticState()
            fmt.Printf("  Fires: %d, Rate: %.1f Hz, Calcium: %.2f\n", 
                state.RecentFires, state.CurrentFireRate, state.CalciumLevel)
        }
    }
    
    fmt.Println()
    
    // Phase 2: Fast firing
    fmt.Println("Phase 2: Fast firing (10 Hz)")
    for i := 0; i < 20; i++ {
        input <- neuron.Message{Value: 1.5}
        time.Sleep(100 * time.Millisecond)
        
        if i%5 == 0 {
            state := hn.GetHomeostaticState()
            fmt.Printf("  Fires: %d, Rate: %.1f Hz, Calcium: %.2f\n", 
                state.RecentFires, state.CurrentFireRate, state.CalciumLevel)
        }
    }
    
    fmt.Println()
    
    // Phase 3: Silence and decay
    fmt.Println("Phase 3: Silence (calcium decay)")
    for i := 0; i < 10; i++ {
        time.Sleep(200 * time.Millisecond)
        hn.updateHomeostasis()
        
        state := hn.GetHomeostaticState()
        fmt.Printf("  Time: %ds, Rate: %.1f Hz, Calcium: %.2f\n", 
            i*200/1000, state.CurrentFireRate, state.CalciumLevel)
    }
    
    fmt.Println("\n✅ Activity tracking working correctly!")
}
```

### **🧪 Experiment 2.2: Threshold Adaptation Mechanism**

**What to Test:**
- Automatic threshold adjustment based on activity
- Bidirectional adaptation (up and down)
- Adaptation rate and stability
- Calcium-modulated adaptation

**Unit Test Implementation:**

```go
// File: experiments/phase-2-homeostasis/experiments/2-threshold-adaptation/experiment.go
package threshold_adaptation

func TestThresholdAdaptationUp(t *testing.T) {
    hn := common.NewHomeostaticNeuron("adapt_up", 1.0, 0.95, 10*time.Millisecond, 1.0, 5.0)
    
    go hn.Run()
    defer hn.Close()
    
    initialState := hn.GetHomeostaticState()
    initialThreshold := initialState.CurrentThreshold
    
    // Overstimulate neuron (target: 5Hz, we'll do 15Hz)
    input := hn.GetInput()
    for i := 0; i < 50; i++ {
        input <- neuron.Message{Value: 1.5}
        time.Sleep(66 * time.Millisecond) // ~15 Hz
        
        if i%10 == 0 {
            hn.updateHomeostasis()
        }
    }
    
    // Allow final adaptation
    time.Sleep(200 * time.Millisecond)
    hn.updateHomeostasis()
    
    finalState := hn.GetHomeostaticState()
    
    if finalState.CurrentThreshold <= initialThreshold {
        t.Error("Hyperactive neuron should have raised its threshold")
    }
    
    if finalState.CurrentFireRate > finalState.TargetFireRate*1.5 {
        t.Error("Adaptation should have reduced firing rate toward target")
    }
}

func TestThresholdAdaptationDown(t *testing.T) {
    // Start with high threshold to create "silent" neuron
    hn := common.NewHomeostaticNeuron("adapt_down", 2.0, 0.95, 10*time.Millisecond, 1.0, 5.0)
    
    go hn.Run()
    defer hn.Close()
    
    initialState := hn.GetHomeostaticState()
    initialThreshold := initialState.CurrentThreshold
    
    // Let it sit in silence (no input)
    for i := 0; i < 20; i++ {
        time.Sleep(100 * time.Millisecond)
        hn.updateHomeostasis()
    }
    
    finalState := hn.GetHomeostaticState()
    
    if finalState.CurrentThreshold >= initialThreshold {
        t.Error("Silent neuron should have lowered its threshold")
    }
    
    // Now test if it's more responsive
    input := hn.GetInput()
    input <- neuron.Message{Value: 1.5} // Would not fire with original threshold
    
    // Check if it fires more easily now
    output := make(chan neuron.Message, 10)
    hn.AddOutput("test", output, 1.0, 0)
    
    select {
    case <-output:
        // Good: now fires with previously sub-threshold input
    case <-time.After(20 * time.Millisecond):
        t.Error("Adapted neuron should be more responsive to weak inputs")
    }
}

func TestAdaptationStability(t *testing.T) {
    hn := common.NewHomeostaticNeuron("stability", 1.0, 0.95, 10*time.Millisecond, 1.0, 5.0)
    
    // Test that adaptation converges and doesn't oscillate wildly
    // Implementation details...
}
```

**Interactive Visualization:**

```go
func RunThresholdAdaptationDemo() {
    fmt.Println("🧠 Threshold Adaptation Demo")
    fmt.Println("============================")
    
    // Demo 1: Hyperactive neuron self-regulation
    fmt.Println("Demo 1: Hyperactive Neuron Self-Regulation")
    fmt.Println("Target Rate: 5.0 Hz, Stimulation: 15 Hz")
    fmt.Println()
    
    hn := common.NewHomeostaticNeuron("hyperactive", 1.0, 0.95, 10*time.Millisecond, 1.0, 5.0)
    go hn.Run()
    defer hn.Close()
    
    input := hn.GetInput()
    
    for second := 0; second < 10; second++ {
        // High frequency stimulation
        for i := 0; i < 15; i++ {
            input <- neuron.Message{Value: 1.5}
            time.Sleep(66 * time.Millisecond) // ~15 Hz
        }
        
        hn.updateHomeostasis()
        state := hn.GetHomeostaticState()
        
        // Visual threshold representation
        thresholdBar := strings.Repeat("█", int(state.CurrentThreshold*10))
        baselineBar := strings.Repeat("░", int(state.BaselineThreshold*10))
        
        fmt.Printf("T=%02ds: Threshold:[%s] Rate:%.1f/%.1f Hz Calcium:%.1f\n",
            second, fmt.Sprintf("%-15s", thresholdBar), 
            state.CurrentFireRate, state.TargetFireRate, state.CalciumLevel)
        
        if second == 0 {
            fmt.Printf("       Baseline: [%s]\n", fmt.Sprintf("%-15s", baselineBar))
            fmt.Println()
        }
    }
    
    fmt.Println("\n✅ Neuron successfully self-regulated!")
    fmt.Println()
    
    // Demo 2: Silent neuron activation
    fmt.Println("Demo 2: Silent Neuron Activation")
    fmt.Println("High threshold (2.0), no stimulation → should lower threshold")
    fmt.Println()
    
    silentNeuron := common.NewHomeostaticNeuron("silent", 2.0, 0.95, 10*time.Millisecond, 1.0, 5.0)
    go silentNeuron.Run()
    defer silentNeuron.Close()
    
    for second := 0; second < 8; second++ {
        time.Sleep(1 * time.Second)
        silentNeuron.updateHomeostasis()
        
        state := silentNeuron.GetHomeostaticState()
        
        thresholdBar := strings.Repeat("█", int(state.CurrentThreshold*5))
        
        fmt.Printf("T=%02ds: Threshold:[%s] (%.2f) Rate:%.1f Hz\n",
            second, fmt.Sprintf("%-10s", thresholdBar), 
            state.CurrentThreshold, state.CurrentFireRate)
    }
    
    fmt.Println("\n✅ Silent neuron increased sensitivity!")
}
```

**Expected Output:**

```
🧠 Threshold Adaptation Demo
============================
Demo 1: Hyperactive Neuron Self-Regulation
Target Rate: 5.0 Hz, Stimulation: 15 Hz

T=00s: Threshold:[██████████    ] Rate:15.2/5.0 Hz Calcium:15.3
       Baseline: [██████████    ]

T=01s: Threshold:[███████████   ] Rate:14.8/5.0 Hz Calcium:28.9
T=02s: Threshold:[████████████  ] Rate:13.1/5.0 Hz Calcium:35.2
T=03s: Threshold:[█████