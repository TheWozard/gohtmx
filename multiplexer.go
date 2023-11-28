package gohtmx

import (
	"context"
	"sync"
)

// Multiplexer allows for subscription to existing channel with all subscriptions receiving all events.
// Subscriptions will be automatically dropped once context is closed or the channel panics.
type Multiplexer struct {
	sync.Mutex
	subscriptions []Subscription
	// TODO: extend to support last N messages.
	lastEvent *SSEEvent
}

func (m *Multiplexer) Subscriptions() []Subscription {
	return m.subscriptions
}

// Subscribe attaches a channel into the multiplexer.
func (m *Multiplexer) Subscribe(ctx context.Context, c chan SSEEvent) context.Context {
	m.Lock()
	defer m.Unlock()
	ctx, cancel := context.WithCancel(ctx)
	subscription := Subscription{Ctx: ctx, Cancel: cancel, C: c}
	m.subscriptions = append(m.subscriptions, subscription)
	if m.lastEvent != nil {
		subscription.SafeWrite(*m.lastEvent)
	}
	return ctx
}

// Send sends the event to all current subscribers and will drop any subscription that is canceled.
// Automatically cancels any subscription that panics on a send.
func (m *Multiplexer) Send(event SSEEvent) {
	m.Lock()
	defer m.Unlock()
	m.lastEvent = &event
	cleanup := false
	// Attempt to send events to all channels.
	for _, subscription := range m.subscriptions {
		if subscription.IsReady() {
			subscription.SafeWrite(event)
		} else {
			cleanup = true
		}
	}

	// An invalid subscription has been found so we should clean up connections.
	if cleanup {
		results := make([]Subscription, len(m.subscriptions))
		offset := 0
		for i, subscription := range m.subscriptions {
			if subscription.IsReady() {
				results[i-offset] = subscription
			} else {
				offset += 1
			}
		}
		m.subscriptions = results[:len(m.subscriptions)-offset]
	}
}

// Start opens a new channel and starts forwarding the events to all subscriptions.
func (m *Multiplexer) Start() chan SSEEvent {
	events := make(chan SSEEvent)
	go func() {
		for event := range events {
			m.Send(event)
		}
	}()
	return events
}

type Subscription struct {
	Ctx    context.Context
	Cancel context.CancelFunc
	C      chan SSEEvent
}

func (s Subscription) IsReady() bool {
	return s.Ctx.Err() == nil
}

func (s Subscription) SafeWrite(e SSEEvent) {
	defer func() {
		if r := recover(); r != nil {
			s.Cancel()
		}
	}()
	s.C <- e
}
