package synapse

import (
	"testing"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/types" // Assuming types package is imported correctly
)

// This file contains dedicated test cases to debug the STDP calculation issue.
// It will print intermediate values to help pinpoint where the calculation goes wrong.

// TestSTDPDebug_LTP_Calculation is a debug test case for Long-Term Potentiation (LTP).
// It aims to specifically show the values involved in the STDP calculation when LTP should occur.
// This test is expected to FAIL if the STDP calculation results in zero or negative change.
func TestSTDPDebug_LTP_Calculation(t *testing.T) {
	t.Log("=== DEBUG: STDP LTP Calculation (Expected to Fail if Bug Persists) ===")

	// 1. Setup: Create mock neurons
	preNeuron := NewMockNeuron("debug_pre_LTP")
	postNeuron := NewMockNeuron("debug_post_LTP")

	// 2. Configure STDP parameters: Use values that should clearly cause LTP
	stdpConfig := types.PlasticityConfig{
		Enabled:        true,
		LearningRate:   0.01,                  // 1% learning rate
		TimeConstant:   20 * time.Millisecond, // 20ms time constant
		WindowSize:     100 * time.Millisecond,
		MinWeight:      0.001,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.2,
	}

	// 3. Create synapse with a known initial weight
	initialWeight := 1.0
	synapse := NewBasicSynapse("debug_syn_LTP", preNeuron, postNeuron,
		stdpConfig, CreateDefaultPruningConfig(), initialWeight, 0)

	t.Logf("Synapse created: ID=%s, Initial Weight=%.6f", synapse.ID(), initialWeight)

	// 4. Define specific timing for LTP (pre-synaptic spike before post-synaptic)
	deltaT := -10 * time.Millisecond // Pre 10ms before post (strong LTP)
	adjustment := types.PlasticityAdjustment{DeltaT: deltaT}

	t.Logf("Applying plasticity with DeltaT = %v", deltaT)

	// --- Debugging output start ---
	// Print values before calling ApplyPlasticity
	t.Logf("DEBUG Input: LearningRate=%.4f, TimeConstant=%v", stdpConfig.LearningRate, stdpConfig.TimeConstant)

	// Capture initial weight
	weightBefore := synapse.GetWeight()

	// 5. Apply plasticity (this will call calculateSTDPWeightChange internally)
	synapse.ApplyPlasticity(adjustment)

	// Capture final weight
	weightAfter := synapse.GetWeight()
	weightChange := weightAfter - weightBefore

	t.Logf("Weight before ApplyPlasticity: %.10f", weightBefore)
	t.Logf("Weight after ApplyPlasticity:  %.10f", weightAfter)
	t.Logf("Calculated Weight Change (weightAfter - weightBefore): %.10f", weightChange)

	// --- Debugging output end ---

	// 6. Assertions: This is where the test should FAIL if STDP didn't work.
	// We expect a positive weight change, not zero.
	expectedMinChange := 0.0001 // A small positive value, based on typical STDP calculation

	if weightChange <= 0 {
		t.Errorf("FAIL: Expected a positive weight change (LTP), but got %.10f. STDP calculation bug!", weightChange)
		t.Log("This indicates that calculateSTDPWeightChange returned 0 or a negative value.")
		t.Log("Check intermediate calculations inside calculateSTDPWeightChange via debug logs.")
	} else if weightChange < expectedMinChange {
		t.Errorf("FAIL: Expected a significant positive weight change (LTP > %.4f), but got %.10f. STDP effect too small!", expectedMinChange, weightChange)
		t.Log("This might indicate that the learning rate or exponential decay are too weak.")
	} else {
		t.Logf("PASS: Successfully observed positive weight change (LTP): %.10f", weightChange)
	}

	t.Log("----------------------------------------------------------------------")
}

// TestSTDPDebug_LTD_Calculation is a debug test case for Long-Term Depression (LTD).
// It aims to specifically show the values involved in the STDP calculation when LTD should occur.
// This test is expected to FAIL if the STDP calculation results in zero or positive change.
func TestSTDPDebug_LTD_Calculation(t *testing.T) {
	t.Log("=== DEBUG: STDP LTD Calculation (Expected to Fail if Bug Persists) ===")

	// 1. Setup: Create mock neurons
	preNeuron := NewMockNeuron("debug_pre_LTD")
	postNeuron := NewMockNeuron("debug_post_LTD")

	// 2. Configure STDP parameters: Use values that should clearly cause LTD
	stdpConfig := types.PlasticityConfig{
		Enabled:        true,
		LearningRate:   0.01,
		TimeConstant:   20 * time.Millisecond,
		WindowSize:     100 * time.Millisecond,
		MinWeight:      0.001,
		MaxWeight:      2.0,
		AsymmetryRatio: 1.2, // LTD should be 1.2x stronger than LTP
	}

	// 3. Create synapse with a known initial weight
	initialWeight := 1.0
	synapse := NewBasicSynapse("debug_syn_LTD", preNeuron, postNeuron,
		stdpConfig, CreateDefaultPruningConfig(), initialWeight, 0)

	t.Logf("Synapse created: ID=%s, Initial Weight=%.6f", synapse.ID(), initialWeight)

	// 4. Define specific timing for LTD (post-synaptic spike before pre-synaptic)
	deltaT := 10 * time.Millisecond // Pre 10ms after post (strong LTD)
	adjustment := types.PlasticityAdjustment{DeltaT: deltaT}

	t.Logf("Applying plasticity with DeltaT = %v", deltaT)

	// --- Debugging output start ---
	// Print values before calling ApplyPlasticity
	t.Logf("DEBUG Input: LearningRate=%.4f, AsymmetryRatio=%.1f, TimeConstant=%v", stdpConfig.LearningRate, stdpConfig.AsymmetryRatio, stdpConfig.TimeConstant)

	// Capture initial weight
	weightBefore := synapse.GetWeight()

	// 5. Apply plasticity
	synapse.ApplyPlasticity(adjustment)

	// Capture final weight
	weightAfter := synapse.GetWeight()
	weightChange := weightAfter - weightBefore

	t.Logf("Weight before ApplyPlasticity: %.10f", weightBefore)
	t.Logf("Weight after ApplyPlasticity:  %.10f", weightAfter)
	t.Logf("Calculated Weight Change (weightAfter - weightBefore): %.10f", weightChange)

	// --- Debugging output end ---

	// 6. Assertions: This is where the test should FAIL if LTD didn't work.
	// We expect a negative weight change, not zero or positive.
	expectedMaxChange := -0.0001 // A small negative value, based on typical STDP calculation

	if weightChange >= 0 {
		t.Errorf("FAIL: Expected a negative weight change (LTD), but got %.10f. STDP calculation bug!", weightChange)
		t.Log("This indicates that calculateSTDPWeightChange returned 0 or a positive value.")
		t.Log("Check intermediate calculations inside calculateSTDPWeightChange via debug logs.")
	} else if weightChange > expectedMaxChange {
		t.Errorf("FAIL: Expected a significant negative weight change (LTD < %.4f), but got %.10f. STDP effect too small!", expectedMaxChange, weightChange)
		t.Log("This might indicate that the learning rate, asymmetry ratio, or exponential decay are too weak.")
	} else {
		t.Logf("PASS: Successfully observed negative weight change (LTD): %.10f", weightChange)
	}

	t.Log("----------------------------------------------------------------------")
}
