# 🧬 Baby Steps: Temporal Neuron Biological Experiments

This experiment validates the core biological behaviors of temporal neurons before adding learning mechanisms.

## 🎯 Purpose

Prove that your temporal neuron implementation exhibits authentic biological behaviors:
- ✅ Leaky integration and temporal summation
- ✅ Refractory periods prevent rapid firing
- ✅ Synaptic transmission delays
- ✅ Excitatory and inhibitory signals
- ✅ Network signal propagation

## 🚀 Quick Start

```bash
# From the temporal-neuron root directory:
cd experiments/1-baby-steps

# Install dependencies
go mod tidy

# Run the experiment
go run main.go
```

## 🧪 Experiments Available

### 1. **Leaky Integration & Temporal Summation**
- **Test:** Multiple weak signals can sum to cause firing
- **Biology:** Models how real neurons integrate postsynaptic potentials over time
- **Expected:** Single weak signal fails, two quick signals succeed, two slow signals fail

### 2. **Refractory Period Demonstration**
- **Test:** Neurons cannot fire during recovery period
- **Biology:** Models Na+ channel inactivation after action potentials
- **Expected:** Strong signal fires, immediate second signal blocked, delayed third signal succeeds

### 3. **Synaptic Transmission Delays**
- **Test:** Signals take time to propagate between neurons
- **Biology:** Models axon conduction and synaptic delays
- **Expected:** Source fires immediately, target fires after programmed delay

### 4. **Excitation vs Inhibition**
- **Test:** Inhibitory signals can prevent firing
- **Biology:** Models GABA vs glutamate neurotransmitter effects
- **Expected:** Inhibition reduces excitation, strong excitation overcomes inhibition

### 5. **Network Signal Propagation**
- **Test:** Activity cascades through connected neurons
- **Biology:** Models how signals flow through neural circuits
- **Expected:** A→B→C cascade with cumulative delays

## 🎮 Controls

- **1-5**: Select experiment from menu
- **1-3**: Execute experiment steps (when in experiment)
- **n**: Next step (without executing)
- **r**: Restart current experiment
- **m**: Return to main menu
- **q**: Quit

## 📊 What Success Looks Like

Each experiment has **clear pass/fail criteria**:

✅ **Pass**: Neurons behave exactly as biological theory predicts  
❌ **Fail**: Unexpected behavior indicates implementation issues

## 🔬 Scientific Validation

This experiment suite provides **empirical evidence** that your temporal neurons are not just "artificial" but genuinely exhibit biological dynamics that differ fundamentally from traditional ANNs:

- **No batch processing** - Continuous real-time operation
- **No activation functions** - Simple threshold-based firing
- **No synchronous updates** - Asynchronous message passing
- **Temporal dynamics** - Time-dependent integration and delays
- **Biological constraints** - Refractory periods and leaky membranes

## 🎓 Learning Outcomes

After running these experiments, you'll have **quantitative proof** that:

1. Your neurons integrate signals over time (not instantaneously)
2. Biological timing constraints are properly enforced
3. Network dynamics emerge from simple local rules
4. The foundation is solid for adding learning mechanisms

## 🚪 Next Steps

Once all experiments pass consistently:
- **Phase 2**: Add calcium-based homeostasis
- **Phase 3**: Implement spike-timing dependent plasticity (STDP)
- **Phase 4**: Test emergent learning behaviors

This is your **biological validation suite** - proving the foundation works before building learning on top! 🧠✨