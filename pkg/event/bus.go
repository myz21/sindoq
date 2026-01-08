package event

import (
	"sync"
)

// Bus implements Emitter with fan-out to subscribers.
type Bus struct {
	mu          sync.RWMutex
	subscribers map[EventType][]subscription
	allSubs     []subscription
	nextID      int
}

type subscription struct {
	id      int
	handler EventHandler
}

// NewBus creates a new event bus.
func NewBus() *Bus {
	return &Bus{
		subscribers: make(map[EventType][]subscription),
	}
}

// Emit sends an event to all relevant subscribers.
func (b *Bus) Emit(event *Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Type-specific subscribers
	for _, sub := range b.subscribers[event.Type] {
		go sub.handler(event)
	}

	// All-events subscribers
	for _, sub := range b.allSubs {
		go sub.handler(event)
	}
}

// EmitSync sends an event synchronously (blocks until all handlers complete).
func (b *Bus) EmitSync(event *Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Type-specific subscribers
	for _, sub := range b.subscribers[event.Type] {
		sub.handler(event)
	}

	// All-events subscribers
	for _, sub := range b.allSubs {
		sub.handler(event)
	}
}

// Subscribe registers a handler for a specific event type.
// Returns an unsubscribe function.
func (b *Bus) Subscribe(eventType EventType, handler EventHandler) func() {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := b.nextID
	b.nextID++

	sub := subscription{id: id, handler: handler}
	b.subscribers[eventType] = append(b.subscribers[eventType], sub)

	return func() {
		b.unsubscribe(eventType, id)
	}
}

// SubscribeMultiple registers a handler for multiple event types.
// Returns an unsubscribe function that removes all subscriptions.
func (b *Bus) SubscribeMultiple(eventTypes []EventType, handler EventHandler) func() {
	unsubscribers := make([]func(), 0, len(eventTypes))
	for _, et := range eventTypes {
		unsubscribers = append(unsubscribers, b.Subscribe(et, handler))
	}

	return func() {
		for _, unsub := range unsubscribers {
			unsub()
		}
	}
}

// SubscribeAll registers a handler for all events.
// Returns an unsubscribe function.
func (b *Bus) SubscribeAll(handler EventHandler) func() {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := b.nextID
	b.nextID++

	sub := subscription{id: id, handler: handler}
	b.allSubs = append(b.allSubs, sub)

	return func() {
		b.unsubscribeAll(id)
	}
}

func (b *Bus) unsubscribe(eventType EventType, id int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	subs := b.subscribers[eventType]
	for i, sub := range subs {
		if sub.id == id {
			b.subscribers[eventType] = append(subs[:i], subs[i+1:]...)
			return
		}
	}
}

func (b *Bus) unsubscribeAll(id int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i, sub := range b.allSubs {
		if sub.id == id {
			b.allSubs = append(b.allSubs[:i], b.allSubs[i+1:]...)
			return
		}
	}
}

// SubscriberCount returns the total number of subscribers.
func (b *Bus) SubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	count := len(b.allSubs)
	for _, subs := range b.subscribers {
		count += len(subs)
	}
	return count
}

// Clear removes all subscribers.
func (b *Bus) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers = make(map[EventType][]subscription)
	b.allSubs = nil
}

// ensure Bus implements Emitter
var _ Emitter = (*Bus)(nil)
