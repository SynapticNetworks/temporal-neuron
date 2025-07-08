package neuron

import (
	"log"
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
		scheduledTime:       time.Time{}, // Zero time
		lastFeedback:        time.Time{}, // Zero time
		totalFeedbackEvents: 0,
	}
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
	feedbackCount := s.processSTDPFeedback(neuronID, callbacks, postSpikeTime)
	return feedbackCount > 0
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
	return s.processSTDPFeedback(neuronID, callbacks, postFiringTime)
}

// processSTDPFeedback is the core implementation
// Modified to accept explicit postSpikeTime
func (s *STDPSignalingSystem) processSTDPFeedback(neuronID string, callbacks component.NeuronCallbacks, postSpikeTime time.Time) int {
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

	// NOTE: No longer using time.Now() here
	feedbackCount := 0

	for _, synapse := range synapses {
		// Skip synapses with invalid LastActivity
		//if synapse.LastActivity.IsZero() {
		//	continue
		//}

		// Calculate timing difference
		//deltaT := synapse.LastActivity.Sub(postSpikeTime)

		// Skip synapses with invalid LastTransmission (instead of LastActivity)
		if synapse.LastTransmission.IsZero() {
			continue
		}

		// Calculate timing difference using LastTransmission
		deltaT := synapse.LastTransmission.Sub(postSpikeTime)

		// Add validation to ensure the sign is correct for biological interpretation
		if deltaT > 0 && synapse.LastTransmission.After(postSpikeTime) {
			// Correct: Post fired before Pre (LTD case)
		} else if deltaT < 0 && synapse.LastTransmission.Before(postSpikeTime) {
			// Correct: Pre fired before Post (LTP case)
		} else if deltaT != 0 {
			// Something is wrong with our timing calculation
			log.Printf("STDP Warning: deltaT=%v but LastTransmission=%v, postSpikeTime=%v",
				deltaT, synapse.LastTransmission, postSpikeTime)
		}

		// Add debugging
		log.Printf("STDP Debug: synapse=%s, lastActivity=%v, postSpikeTime=%v, deltaT=%v\n", synapse.ID, synapse.LastActivity, postSpikeTime, deltaT)

		// Skip synapses with zero deltaT
		if deltaT == 0 {
			continue
		}

		// If deltaT is very close to zero but not exactly zero, ensure it's at least 1 nanosecond
		if deltaT > -time.Nanosecond && deltaT < time.Nanosecond {
			if deltaT >= 0 {
				deltaT = time.Nanosecond // Ensure positive but non-zero
			} else {
				deltaT = -time.Nanosecond // Ensure negative but non-zero
			}
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
		err := callbacks.ApplyPlasticity(synapse.ID, adjustment)
		if err == nil {
			feedbackCount++
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
