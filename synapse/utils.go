package synapse

import "math"

// Helper function for max (since Go doesn't have built-in max for int)
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// =================================================================================
// ROBUST INPUT VALIDATION HELPERS
// =================================================================================

// validateFloat64 checks if a float64 value is valid (not NaN or Inf) and returns a fallback if invalid
func validateFloat64(value, fallback float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fallback
	}
	return value
}

// clampFloat64 ensures a value is within specified bounds
func clampFloat64(value, min, max float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return (min + max) / 2.0 // Return midpoint for invalid values
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// validateCooperativity ensures cooperativity value is reasonable
func validateCooperativity(cooperativeInputs int) int {
	if cooperativeInputs < 0 {
		return 0
	}
	if cooperativeInputs > 1000 { // Reasonable biological upper limit
		return 1000
	}
	return cooperativeInputs
}
