package neuron

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// STDPSignalingSystem manages the scheduling and delivery of STDP feedback signals
// This system is responsible only for the timing and delivery of signals,
// not the actual STDP weight adjustment logic which is handled by synapses
type STDPSignalingSystem struct {
	// Configuration
	enabled       bool
	feedbackDelay time.Duration
	learningRate  float64

	windowSize time.Duration // How far back in time to consider spikes

	// Configuration for STDP curve
	ltpTimeConstant time.Duration
	ltdTimeConstant time.Duration
	ltpMaxChange    float64
	ltdMaxChange    float64

	//store the post-spike time
	postSpikeTime time.Time

	// State
	scheduledTime time.Time
	lastFeedback  time.Time

	// Statistics
	totalFeedbackEvents int

	// Thread safety
	mutex sync.Mutex
}

// NewSTDPSignalingSystem creates a new STDP signaling system
func NewSTDPSignalingSystem(enabled bool, feedbackDelay time.Duration, learningRate float64) *STDPSignalingSystem {
	return &STDPSignalingSystem{
		enabled:             enabled,
		feedbackDelay:       feedbackDelay,
		learningRate:        learningRate,
		windowSize:          100 * time.Millisecond, // Default window size
		scheduledTime:       time.Time{},
		lastFeedback:        time.Time{},
		totalFeedbackEvents: 0,
	}
}

// SetWindowSize sets how far back in time to consider spikes for STDP
func (s *STDPSignalingSystem) SetWindowSize(windowSize time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.windowSize = windowSize
}

// IsEnabled returns whether STDP signaling is currently enabled
func (s *STDPSignalingSystem) IsEnabled() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.enabled
}

// Enable turns on STDP signaling with the specified parameters
func (s *STDPSignalingSystem) Enable(feedbackDelay time.Duration, learningRate float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.enabled = true
	s.feedbackDelay = feedbackDelay
	s.learningRate = learningRate
}

// Disable turns off STDP signaling
func (s *STDPSignalingSystem) Disable() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.enabled = false
	s.scheduledTime = time.Time{} // Clear any scheduled events
}

// GetParameters returns the current STDP parameters
func (s *STDPSignalingSystem) GetParameters() (time.Duration, float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.feedbackDelay, s.learningRate
}

// ScheduleFeedback schedules an STDP feedback event for future delivery
// Returns true if scheduling was successful, false otherwise
func (s *STDPSignalingSystem) ScheduleFeedback(fireTime time.Time) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Don't schedule if disabled
	if !s.enabled {
		return false
	}

	// Calculate execution time based on firing time and feedback delay
	executeTime := fireTime.Add(s.feedbackDelay)

	// Only schedule if there's no pending feedback or this one is earlier
	if s.scheduledTime.IsZero() || executeTime.Before(s.scheduledTime) {
		s.scheduledTime = executeTime
		s.postSpikeTime = fireTime // Store the actual post-spike time
		return true
	}

	return false
}

// CheckAndDeliverFeedback checks if it's time to deliver scheduled feedback
// This method is designed to be called regularly, e.g., from a neuron's main loop
// Returns true if feedback was delivered, false otherwise
func (s *STDPSignalingSystem) CheckAndDeliverFeedback(neuronID string, callbacks component.NeuronCallbacks) bool {
	// First, check if we should execute without holding the lock for long
	s.mutex.Lock()

	// Quick exit conditions
	if !s.enabled || callbacks == nil || s.scheduledTime.IsZero() || !time.Now().After(s.scheduledTime) {
		s.mutex.Unlock()
		return false
	}
	// Get the stored post-spike time
	postSpikeTime := s.postSpikeTime

	// Reset scheduled time while holding the lock
	s.scheduledTime = time.Time{}
	s.postSpikeTime = time.Time{}

	s.mutex.Unlock()

	// Call the shared implementation with current time
	// In CheckAndDeliverFeedback method
	feedbackCount := s.processSTDPFeedbackWithSpikeHistory(neuronID, callbacks, postSpikeTime)
	return feedbackCount > 0
}

// This is a replacement implementation for the processSTDPFeedbackWithSpikeHistory method
// in the STDPSignalingSystem struct located in stdp_signaling.go
// Improved implementation for processSTDPFeedbackWithSpikeHistory
// This version prioritizes finding the proper LTD (post before pre) timing when available
// Corrected implementation for processSTDPFeedbackWithSpikeHistory
// This version properly uses bestPreSpike and bestPostSpike for deltaT calculation
// Replace the existing processSTDPFeedbackWithSpikeHistory function in stdp_signaling.go
// with this implementation that properly handles LTD

func (s *STDPSignalingSystem) processSTDPFeedbackWithSpikeHistory(neuronID string, callbacks component.NeuronCallbacks, postSpikeTime time.Time) int {
	s.mutex.Lock()
	// Quick exit conditions
	if !s.enabled || callbacks == nil {
		s.mutex.Unlock()
		return 0
	}
	learningRate := s.learningRate
	s.mutex.Unlock()

	// Get all incoming synapses to this neuron
	incomingDirection := types.SynapseIncoming
	targetID := neuronID
	synapses := callbacks.ListSynapses(types.SynapseCriteria{
		TargetID:  &targetID,
		Direction: &incomingDirection,
	})

	feedbackCount := 0

	//fmt.Printf("STDP Feedback: Processing for neuron=%s, postSpikeTime=%v with %d incoming synapses\n",neuronID, postSpikeTime, len(synapses))

	// Process each synapse using spike history
	for _, synapseInfo := range synapses {
		// Get the actual synapse object
		synObj, err := callbacks.GetSynapse(synapseInfo.ID)
		if err != nil {
			continue
		}

		// Check if synapse implements spike history interface
		if spikesGetter, ok := synObj.(interface {
			GetPreSpikeTimes() []time.Time
			GetPostSpikeTimes() []time.Time
		}); ok {
			// Get spike histories
			preSpikes := spikesGetter.GetPreSpikeTimes()
			postSpikes := spikesGetter.GetPostSpikeTimes()

			//fmt.Printf("STDP History: Synapse=%s has %d pre-spikes and %d post-spikes\n",synapseInfo.ID, len(preSpikes), len(postSpikes))

			// If no spikes, try fallback to LastActivity/LastTransmission
			if len(preSpikes) == 0 || len(postSpikes) == 0 {
				// Fallback to old method
				if synapseInfo.LastTransmission.IsZero() {
					continue
				}

				// Calculate deltaT using old method
				deltaT := synapseInfo.LastTransmission.Sub(postSpikeTime)

				// Skip synapses with zero deltaT
				if deltaT == 0 {
					continue
				}

				// Create plasticity adjustment
				adjustment := types.PlasticityAdjustment{
					DeltaT:       deltaT,
					LearningRate: learningRate,
					PostSynaptic: true,
					PreSynaptic:  true,
					Timestamp:    postSpikeTime,
					EventType:    types.PlasticitySTDP,
				}

				// Apply plasticity
				err := callbacks.ApplyPlasticity(synapseInfo.ID, adjustment)
				if err == nil {
					feedbackCount++
				}
			} else {
				// CRITICAL FIX: Instead of just finding closest pre-spike,
				// we need to search for both LTP and LTD spike patterns:

				// 1. Look for LTD pattern (post-before-pre)
				var ltdPreSpike, ltdPostSpike time.Time
				var ltdDeltaT time.Duration
				var foundLTD bool

				// For each post-spike, look for a pre-spike that comes after it
				for _, postSpike := range postSpikes {
					for _, preSpike := range preSpikes {
						if preSpike.After(postSpike) {
							// This is a potential LTD pair - post spike happens before pre spike
							deltaT := preSpike.Sub(postSpike) // POSITIVE for LTD

							// Keep the pair with the smallest positive delta (closest timing)
							if !foundLTD || deltaT < ltdDeltaT {
								ltdPostSpike = postSpike
								ltdPreSpike = preSpike
								ltdDeltaT = deltaT
								foundLTD = true
							}
						}
					}
				}

				// 2. Look for LTP pattern (pre-before-post)
				var ltpPreSpike, ltpPostSpike time.Time
				var ltpDeltaT time.Duration
				var foundLTP bool

				// For each pre-spike, look for a post-spike that comes after it
				for _, preSpike := range preSpikes {
					for _, postSpike := range postSpikes {
						if postSpike.After(preSpike) {
							// This is a potential LTP pair - pre spike happens before post spike
							deltaT := preSpike.Sub(postSpike) // NEGATIVE for LTP

							// We need the absolute value for finding the closest pair
							absDeltaT := deltaT
							if absDeltaT < 0 {
								absDeltaT = -absDeltaT
							}

							// Keep the pair with the smallest absolute delta (closest timing)
							if !foundLTP || absDeltaT < -ltpDeltaT {
								ltpPreSpike = preSpike
								ltpPostSpike = postSpike
								ltpDeltaT = deltaT // This will be negative
								foundLTP = true
							}
						}
					}
				}

				// 3. Decide which pattern to use based on recency and clear ordering
				var finalDeltaT time.Duration
				var finalTimestamp time.Time
				var applyLTD bool
				var spikePairDesc string

				if foundLTD && foundLTP {
					// Both patterns found - prefer LTD if it involves more recent spikes
					// or if the timing is more clear (larger absolute deltaT)
					ltdRecent := ltdPostSpike.After(time.Now().Add(-200*time.Millisecond)) &&
						ltdPreSpike.After(time.Now().Add(-200*time.Millisecond))
					ltpRecent := ltpPreSpike.After(time.Now().Add(-200*time.Millisecond)) &&
						ltpPostSpike.After(time.Now().Add(-200*time.Millisecond))

					// If both are recent, prefer the one with clearer timing
					if ltdRecent && ltpRecent {
						// Compare absolute values of deltaT
						if ltdDeltaT > -ltpDeltaT {
							applyLTD = true
							spikePairDesc = "LTD (clearer timing)"
						} else {
							applyLTD = false
							spikePairDesc = "LTP (clearer timing)"
						}
					} else if ltdRecent {
						applyLTD = true
						spikePairDesc = "LTD (more recent)"
					} else if ltpRecent {
						applyLTD = false
						spikePairDesc = "LTP (more recent)"
					} else {
						// Neither is very recent - prefer LTD by default
						applyLTD = true
						spikePairDesc = "LTD (default preference)"
					}
				} else if foundLTD {
					applyLTD = true
					spikePairDesc = "LTD (only pattern found)"
				} else if foundLTP {
					applyLTD = false
					spikePairDesc = "LTP (only pattern found)"
				} else {
					// Neither pattern found - fall back to closest spike
					fmt.Printf("STDP Warning: No clear LTP or LTD pattern found for synapse %s\n", synapseInfo.ID)
					continue
				}

				// Set the correct parameters based on our choice
				if applyLTD {
					finalDeltaT = ltdDeltaT // Positive for LTD
					finalTimestamp = ltdPostSpike
					//fmt.Printf("STDP Debug: Found LTD pair - post=%v, pre=%v, deltaT=%v (%s)\n",ltdPostSpike, ltdPreSpike, ltdDeltaT, spikePairDesc)
				} else {
					finalDeltaT = ltpDeltaT // Negative for LTP
					finalTimestamp = ltpPostSpike
					//fmt.Printf("STDP Debug: Found LTP pair - pre=%v, post=%v, deltaT=%v (%s)\n",ltpPreSpike, ltpPostSpike, ltpDeltaT, spikePairDesc)
				}

				_ = spikePairDesc

				// Apply the chosen STDP adjustment
				adjustment := types.PlasticityAdjustment{
					DeltaT:       finalDeltaT,
					LearningRate: learningRate,
					PostSynaptic: true,
					PreSynaptic:  true,
					Timestamp:    finalTimestamp,
					EventType:    types.PlasticitySTDP,
				}

				// Apply plasticity
				err := callbacks.ApplyPlasticity(synapseInfo.ID, adjustment)
				if err == nil {
					if applyLTD {
						//fmt.Printf("STDP LTD Applied: Synapse %s adjusted with deltaT=%v (positive = LTD)\n",synapseInfo.ID, finalDeltaT)
					} else {
						//fmt.Printf("STDP LTP Applied: Synapse %s adjusted with deltaT=%v (negative = LTP)\n",synapseInfo.ID, finalDeltaT)
					}
					feedbackCount++
				} else {
					fmt.Printf("STDP Error: Failed to apply plasticity: %v\n", err)
				}
			}
		} else {
			// Synapse doesn't support spike history
			// Fall back to the original method
			if synapseInfo.LastTransmission.IsZero() {
				continue
			}

			// Calculate timing difference using LastTransmission
			deltaT := synapseInfo.LastTransmission.Sub(postSpikeTime)

			// Skip synapses with zero deltaT
			if deltaT == 0 {
				continue
			}

			// Create plasticity adjustment
			adjustment := types.PlasticityAdjustment{
				DeltaT:       deltaT,
				LearningRate: learningRate,
				PostSynaptic: true,
				PreSynaptic:  true,
				Timestamp:    postSpikeTime,
				EventType:    types.PlasticitySTDP,
			}

			// Apply plasticity
			err := callbacks.ApplyPlasticity(synapseInfo.ID, adjustment)
			if err == nil {
				feedbackCount++
			}
		}
	}

	// Update statistics
	if feedbackCount > 0 {
		s.mutex.Lock()
		s.lastFeedback = postSpikeTime
		s.totalFeedbackEvents++
		s.mutex.Unlock()
	}

	return feedbackCount
}

// DeliverFeedbackNow forces immediate STDP feedback delivery
// This is useful for testing and direct control
// Returns the number of synapses that received feedback
// Fixed version of DeliverFeedbackNow to correct deltaT sign issue
// DeliverFeedbackNow forces immediate STDP feedback delivery with current neuron firing time
// This is useful for testing and direct control
// Returns the number of synapses that received feedback
// Update the DeliverFeedbackNow method to accept an explicit fire time
func (s *STDPSignalingSystem) DeliverFeedbackNow(neuronID string, callbacks component.NeuronCallbacks, postFiringTime time.Time) int {
	// Use provided time or current time as fallback
	if postFiringTime.IsZero() {
		postFiringTime = time.Now()
	}
	return s.processSTDPFeedbackWithSpikeHistory(neuronID, callbacks, postFiringTime)
}

// ProcessSTDP applies STDP to a synapse based on recent spike history
func (s *STDPSignalingSystem) ProcessSTDP(synapse component.SynapticProcessor,
	preSpikes []time.Time,
	postSpikes []time.Time) float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.enabled || len(preSpikes) == 0 || len(postSpikes) == 0 {
		return 0.0
	}

	// Calculate nearest-neighbor STDP
	var totalWeightChange float64

	// For each post-synaptic spike...
	for _, postTime := range postSpikes {
		// Find nearest pre-synaptic spike
		var nearestPreTime time.Time
		var nearestDeltaT time.Duration
		var foundNearestSpike bool

		for _, preTime := range preSpikes {
			deltaT := postTime.Sub(preTime) // Positive: post after pre (LTP)
			// Negative: pre after post (LTD)

			// Skip spikes outside the STDP window
			if abs(deltaT) > s.windowSize {
				continue
			}

			// Update nearest spike if this is first or closer than previous nearest
			if !foundNearestSpike || abs(deltaT) < abs(nearestDeltaT) {
				nearestPreTime = preTime
				nearestDeltaT = deltaT
				foundNearestSpike = true
			}
		}

		_ = nearestPreTime // Use this to avoid unused variable warning

		// If we found a spike in the STDP window, calculate weight change
		if foundNearestSpike {
			// Calculate weight change based on STDP rule
			var weightDelta float64

			if nearestDeltaT >= 0 {
				// LTP: post after pre
				// Note the sign convention: deltaTMs is positive for LTP
				deltaTMs := float64(nearestDeltaT) / float64(time.Millisecond)
				weightDelta = s.ltpMaxChange * math.Exp(-deltaTMs/
					float64(s.ltpTimeConstant/time.Millisecond))
			} else {
				// LTD: pre after post
				// For LTD, deltaTMs is negative, so we take absolute value for the exp
				deltaTMs := float64(-nearestDeltaT) / float64(time.Millisecond)
				weightDelta = -s.ltdMaxChange * math.Exp(-deltaTMs/
					float64(s.ltdTimeConstant/time.Millisecond))
			}

			totalWeightChange += weightDelta * s.learningRate
		}
	}

	// Apply the total weight change
	if totalWeightChange != 0 {
		currentWeight := synapse.GetWeight()
		newWeight := currentWeight + totalWeightChange

		// Apply bounds (synapse should handle this internally too)
		synapse.SetWeight(newWeight)

		return totalWeightChange
	}

	return 0.0
}

// GetStatus returns the current status of the STDP signaling system
func (s *STDPSignalingSystem) GetStatus() map[string]interface{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return map[string]interface{}{
		"enabled":               s.enabled,
		"feedback_delay":        s.feedbackDelay,
		"learning_rate":         s.learningRate,
		"has_scheduled":         !s.scheduledTime.IsZero(),
		"scheduled_time":        s.scheduledTime,
		"last_feedback":         s.lastFeedback,
		"total_feedback_events": s.totalFeedbackEvents,
	}
}

// Helper function for absolute duration
func abs(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
