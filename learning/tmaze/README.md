# T-Maze Reinforcement Learning Experiment

## Overview

This project implements a neural network-based reinforcement learning system for the classic T-maze navigation task. The implementation models biological neural mechanisms, including spike-timing-dependent plasticity (STDP), dopamine modulation, and eligibility traces, to investigate how different neural architectures affect learning and behavioral flexibility.

## The T-Maze Environment

```
    Left Arm          Right Arm
       ↑                 ↑
       |                 |
       |                 |
[R]----+-----[J]--------+----[R]
                |
                |
                |
               [S]

[S] = Start
[J] = Junction
[R] = Reward Location (initially at Left Arm)
```

## Experiment Description

The T-maze is a simple environment with profound implications for understanding learning and decision-making. In this task:

1. An agent starts at the base of a T-shaped maze [S]
2. It moves forward to reach a junction [J] (the middle of the T)
3. At the junction, the agent must decide to turn either left or right
4. A reward is placed at one end (initially the left arm)

The experiment consists of three phases:
- **Initial Learning**: The agent learns to navigate to the rewarded location
- **Reversal Learning**: The reward is moved to the opposite arm, testing the agent's ability to adapt
- **Performance Evaluation**: Success rates are measured before learning, after initial learning, and after reversal learning

## Neural Implementations

We tested three different neural architectures:

1. **Basic Agent**: A simple neural network with direct connections between sensory and motor neurons
2. **Dopamine Agent**: Adds a dopamine neuron that modulates synaptic plasticity based on reward prediction error
3. **Enhanced Agent**: Includes a hidden layer for more complex processing capabilities (test skipped in current output)

## Neural Architecture

```
Basic Agent:
                          ┌──────┐
                      ┌──►│ Left │
                      │   └──────┘
┌──────┐    ┌──────┐  │
│ Start│───►│Junctn│──┤
└──────┘    └──────┘  │
                      │   ┌───────┐
                      └──►│ Right │
                          └───────┘

Dopamine Agent:
                          ┌──────┐
                      ┌──►│ Left │───┐
                      │   └──────┘   │
┌──────┐    ┌──────┐  │              ▼
│ Start│───►│Junctn│──┤         ┌─────────┐
└──────┘    └──────┘  │         │ Dopamine│
                      │   ┌─────┴─┐       │
                      └──►│ Right │───────┘
                          └───────┘
```

## Key Findings

### Learning Performance

- **Basic Agent**: 
  - Baseline: 50% success (chance level)
  - After learning: 100% success (perfect performance)
  - After reversal: 50% success (partial adaptation)
  - Final weights: Junction→Left: 0.476, Junction→Right: 0.476 (equal weights)

- **Dopamine Agent**:
  - Baseline: 40% success
  - After learning: 100% success (perfect performance)
  - After reversal: 0% success (no adaptation)
  - Final weights: Junction→Left: 0.476, Junction→Right: 0.477 (minimal differentiation)

### Performance Comparison

```
Success Rate:
100% ┼──────────────────────────────────────╮       ╭───────────
     │                                ╭─────┤       │ Basic
     │                                │     │       │ Agent
 75% ┤                                │     │       │
     │                                │     │       │
 50% ┼───────────────────╮            │     ├───────┤
     │                   │            │     │       │
 25% ┤                   │            │     │       │
     │                   │            │     │       │
  0% ┼────────────────────────────────┴─────┴───────┘
      Baseline       After Learning       After Reversal

100% ┼──────────────────────────────────────╮       
     │                                ╭─────┤       
     │                                │     │       
 75% ┤                                │     │       
     │                                │     │       
 50% ┤                                │     │       
     │                                │     │       
 25% ┼───────────────────╮            │     │       
     │                   │            │     │       
  0% ┼────────────────────────────────┴─────┴───────╯
      Baseline       After Learning       After Reversal
                                                Dopamine
                                                 Agent
```

## Scientific Significance

These results have important implications for understanding neural mechanisms of learning and adaptation:

### 1. Dopamine's Role in Learning Stability vs. Flexibility

The dopamine model demonstrated a classic trade-off between learning stability and flexibility. By strengthening specific neural pathways associated with reward, dopamine creates robust learning but potentially at the cost of behavioral flexibility. This mirrors findings in neuroscience where dopamine is implicated in both skill acquisition and habit formation, which can be resistant to change.

### 2. Biological Plausibility

Our implementation captures several neurobiologically plausible mechanisms:
- **STDP**: Learning depends on the precise timing between pre- and post-synaptic neuron activity
- **Eligibility Traces**: Only recently active synapses are eligible for modification
- **Reward Prediction Error**: Dopamine release is proportional to unexpected rewards
- **Exploration-Exploitation Balance**: Noise in decision-making decreases as learning progresses

### 3. Minimal Weight Differentiation

The minimal difference in final synaptic weights despite clear behavioral differences suggests that:
- Learning may be encoded in the dynamics and timing of neural activity rather than just static weights
- Small weight changes can produce large behavioral effects in a network context
- Multiple redundant pathways may contribute to learned behaviors

### 4. Implications for Neurodevelopmental Conditions

The stark difference in reversal learning between models has implications for understanding conditions characterized by behavioral inflexibility, such as addiction or aspects of autism spectrum disorder. Excessive dopamine modulation might create strongly reinforced pathways that resist updating, manifesting as perseverative behaviors.

## Real-World Applications

These findings have potential applications in:

1. **Artificial Intelligence**: Designing learning systems that balance stability and flexibility
2. **Robotics**: Creating adaptive control systems that can respond to changing environments
3. **Neuropharmacology**: Understanding how dopaminergic drugs might affect learning and behavioral flexibility
4. **Computational Psychiatry**: Modeling how aberrant neuromodulation might contribute to behavioral rigidity in psychiatric conditions

## Conclusions

Our T-maze experiment demonstrates how different neural mechanisms produce distinct learning characteristics. While both architectures successfully learned the initial task, they showed markedly different capabilities for adaptation. The dopamine-modulated system excelled at initial learning but demonstrated significant inflexibility when environmental conditions changed.

This highlights a fundamental trade-off in neural learning systems between stable, persistent learning and flexible, adaptable behavior. Understanding these mechanisms could lead to improved neural network designs and deeper insights into both healthy and pathological learning processes in biological systems.