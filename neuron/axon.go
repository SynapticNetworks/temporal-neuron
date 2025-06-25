package neuron

import (
	"sort"
	"time"

	"github.com/SynapticNetworks/temporal-neuron/component"
	"github.com/SynapticNetworks/temporal-neuron/types"
)

// delayedMessage represents a neural signal awaiting axonal delivery.
// Models action potential propagation down the axon with timing delays
// before reaching target synapses.
type delayedMessage struct {
	message      types.NeuralSignal        // The neural signal to deliver
	target       component.MessageReceiver // Target post-synaptic neuron
	deliveryTime time.Time                 // When the message should be delivered
}

// ScheduleDelayedDelivery queues a message for delivery after total propagation delay.
// This function is a helper to be called by the Neuron's method, encapsulating
// the logic of adding to the delivery queue.
//
// IMPORTANT: This function only adds messages to a channel. It DOES NOT spawn new goroutines.
// The actual message dispatching is managed by the Neuron's main processing loop.
//
// Parameters:
//
//	deliveryQueue: The channel to which the delayed message should be added.
//	msg: The neural signal to deliver (includes timing and source info).
//	target: The post-synaptic neuron to receive the types.
//	delay: Total delay including synaptic and spatial components.
func ScheduleDelayedDelivery(deliveryQueue chan<- delayedMessage, msg types.NeuralSignal, target component.MessageReceiver, delay time.Duration) {
	delayedMsg := delayedMessage{
		message:      msg,
		target:       target,
		deliveryTime: time.Now().Add(delay),
	}

	// Attempt to queue for axonal delivery (non-blocking).
	select {
	case deliveryQueue <- delayedMsg:
		// Successfully queued.
	default:
		// Queue full - immediate delivery fallback.
		// This models graceful degradation under extreme network load.
		target.Receive(msg) // Direct receive if queue is full.
	}
}

// ProcessAxonDeliveries manages the dispatching of delayed messages from the axon's queue.
// This function is intended to be called periodically from the Neuron's main Run() loop.
// It checks for new messages added to the queue and dispatches messages whose delivery
// time has arrived.
//
// Parameters:
//
//	pending: The slice holding messages currently awaiting delivery (managed by the caller).
//	newDeliveries: The channel from which new delayed messages are received.
//	now: The current simulation time.
//
// Returns:
//
//	The updated slice of pending deliveries after processing.
func ProcessAxonDeliveries(pending []delayedMessage, newDeliveries <-chan delayedMessage, now time.Time) []delayedMessage {
	// Drain any new messages that have arrived since the last check.
	// This loop uses a `select` with `default` to ensure non-blocking reads.
	// `drainLoop` label is used to break out of the outer `for` loop.
drainLoop:
	for {
		select {
		case msg, ok := <-newDeliveries:
			if !ok {
				// Channel was closed, no more messages will arrive.
				// Set newDeliveries to nil to prevent further select cases from matching it.
				newDeliveries = nil
				break drainLoop // Exit the labeled `for` loop
			}
			pending = append(pending, msg)
		default:
			// No more messages immediately available in the channel.
			break drainLoop // Exit the labeled `for` loop
		}
	}

	// Sort pending deliveries by delivery time for efficient processing.
	// This allows delivering ready messages sequentially and breaking early.
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].deliveryTime.Before(pending[j].deliveryTime)
	})

	remaining := pending[:0] // Reuse slice capacity efficiently

	// Deliver all messages whose delivery time has arrived
	for _, msg := range pending {
		if now.After(msg.deliveryTime) || now.Equal(msg.deliveryTime) {
			// Delivery time reached - transmit to target neuron
			msg.target.Receive(msg.message)
		} else {
			// Not yet time - keep in pending list for future delivery
			remaining = append(remaining, msg)
		}
	}

	// Return the updated pending list (removes delivered messages)
	return remaining
}
