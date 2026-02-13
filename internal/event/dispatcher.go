package event

import (
	"log/slog"
	"sync"
)

const defaultBufferSize = 16

// Dispatcher handles fan-out event routing between plugins.
type Dispatcher struct {
	subscribers map[string][]chan Event
	mu          sync.RWMutex
	closed      bool
	logger      *slog.Logger
}

// New creates a new event dispatcher.
func New() *Dispatcher {
	return &Dispatcher{
		subscribers: make(map[string][]chan Event),
		logger:      slog.Default(),
	}
}

// NewWithLogger creates a dispatcher with custom logger.
func NewWithLogger(logger *slog.Logger) *Dispatcher {
	return &Dispatcher{
		subscribers: make(map[string][]chan Event),
		logger:      logger,
	}
}

// Subscribe creates a buffered channel for receiving events on a topic.
func (d *Dispatcher) Subscribe(topic string) <-chan Event {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		ch := make(chan Event)
		close(ch)
		return ch
	}

	ch := make(chan Event, defaultBufferSize)
	d.subscribers[topic] = append(d.subscribers[topic], ch)
	return ch
}

// Publish sends an event to all subscribers of a topic.
// Non-blocking: drops events if subscriber buffer is full.
func (d *Dispatcher) Publish(topic string, e Event) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed {
		return
	}

	subs, ok := d.subscribers[topic]
	if !ok {
		return
	}

	for _, ch := range subs {
		select {
		case ch <- e:
		default:
			// Buffer full, drop event (best-effort delivery)
			d.logger.Warn("event dropped", "topic", topic, "type", e.Type)
		}
	}
}

// PublishAll sends an event to all subscribers of all topics.
func (d *Dispatcher) PublishAll(e Event) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.closed {
		return
	}

	for topic, subs := range d.subscribers {
		for _, ch := range subs {
			select {
			case ch <- e:
			default:
				d.logger.Warn("event dropped", "topic", topic, "type", e.Type)
			}
		}
	}
}

// Close shuts down the dispatcher and all subscriber channels.
func (d *Dispatcher) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.closed {
		return
	}

	d.closed = true
	for _, subs := range d.subscribers {
		for _, ch := range subs {
			close(ch)
		}
	}
	d.subscribers = nil
}
