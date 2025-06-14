/*
=================================================================================
CHEMICAL RELEASE RATE LIMITING
=================================================================================

Implements biological rate limiting for chemical releases to prevent
unrealistic release frequencies that violate biological constraints.
=================================================================================
*/

package extracellular

import (
	"time"
)

// ChemicalRateLimit manages release frequency limits
type ChemicalRateLimit struct {
	releases   []time.Time
	maxRate    float64 // releases per second
	windowSize time.Duration
}

// NewChemicalRateLimit creates a rate limiter
func NewChemicalRateLimit(maxReleasesPerSecond float64) *ChemicalRateLimit {
	return &ChemicalRateLimit{
		releases:   make([]time.Time, 0),
		maxRate:    maxReleasesPerSecond,
		windowSize: time.Second,
	}
}

// CanRelease checks if a release is allowed
func (crl *ChemicalRateLimit) CanRelease() bool {
	now := time.Now()

	// Remove old releases outside the window
	cutoff := now.Add(-crl.windowSize)
	validReleases := make([]time.Time, 0)
	for _, releaseTime := range crl.releases {
		if releaseTime.After(cutoff) {
			validReleases = append(validReleases, releaseTime)
		}
	}
	crl.releases = validReleases

	// Check if we can add another release
	if float64(len(crl.releases)) < crl.maxRate {
		crl.releases = append(crl.releases, now)
		return true
	}

	return false
}

// GetCurrentRate returns current release rate
func (crl *ChemicalRateLimit) GetCurrentRate() float64 {
	now := time.Now()
	cutoff := now.Add(-crl.windowSize)

	count := 0
	for _, releaseTime := range crl.releases {
		if releaseTime.After(cutoff) {
			count++
		}
	}

	return float64(count)
}
