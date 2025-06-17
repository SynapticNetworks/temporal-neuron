package synapse

// Helper function for max (since Go doesn't have built-in max for int)
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
