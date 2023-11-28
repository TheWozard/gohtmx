package gohtmx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Stream defines a Component that when loaded by the UI will invoke the SSEEventGenerator and read events from it.
// It is required to include StreamTarget Components in the Content of the stream in order to be able to receive events.
type Stream struct {
	// The unique ID of this stream used during the mux to route requests.
	ID      string
	Classes []string
	Content Component
	// Called when the component is mounted.
	SSEEventGenerator SSEEventGenerator
}

func (s Stream) LoadTemplate(l *Location, w io.StringWriter) {
	Div{
		ID:      s.ID,
		Classes: s.Classes,
		Attr: []Attr{
			{Name: "hx-ext", Value: "sse"},
			{Name: "sse-connect", Value: l.Path(severSideEventPrefix, s.ID)},
		},
		Content: s.Content,
	}.LoadTemplate(l, w)
}

func (s Stream) LoadMux(l *Location, m *mux.Router) {
	m.Handle(l.Path(severSideEventPrefix, s.ID), SSEHandler{
		Location:       l,
		EventGenerator: s.SSEEventGenerator,
	})
	if s.Content != nil {
		s.Content.LoadMux(l, m)
	}
}

// StreamTarget is a Component that represents a target location to render an event from a Stream Component
// This Component must be wrapped by a Stream Component in order for events to be received.
type StreamTarget struct {
	// Event is the name of the events this component will receive.
	Events []string
	// Swap defines how events are added to the target. Defaults to SwapReplace
	Swap Swap
	ID   string
	// Classes adds classes to the wrapping div element.
	Classes []string
	// Content defines the Component this wraps.
	Content Component
}

func (st StreamTarget) events() []string {
	if len(st.Events) == 0 {
		return []string{"message"}
	}
	return st.Events
}

func (st StreamTarget) LoadTemplate(l *Location, w io.StringWriter) {
	Div{
		ID:      st.ID,
		Classes: st.Classes,
		Attr: []Attr{
			{Name: "sse-swap", Value: strings.Join(st.events(), ",")},
			{Name: "hx-swap", Value: string(st.Swap.OrDefault(SwapContent))},
		},
		Content: st.Content,
	}.LoadTemplate(l, w)
}

func (st StreamTarget) LoadMux(l *Location, m *mux.Router) {
	if st.Content != nil {
		st.Content.LoadMux(l, m)
	}
}

// SSEEvent and event to be sent to the front end.
type SSEEvent struct {
	// The Event name, this must match an expected event on the front end or it will be ignored.
	Event string
	// The Component to render for this event.
	Data Component
	// The prefix to render this component relative to. If blank will default to the prefix it is served from.
	Prefix string
}

func (se SSEEvent) event() string {
	if se.Event == "" {
		return "message"
	}
	return se.Event
}

func (se SSEEvent) string(l *Location) string {
	var builder strings.Builder
	_, _ = builder.WriteString(fmt.Sprintf("event: %s\n", se.event()))
	_, _ = builder.WriteString("data: ")
	if se.Data != nil {
		if se.Prefix != "" {
			se.Data.LoadTemplate(l, &builder)
		} else {
			se.Data.LoadTemplate(l, &builder)
		}
	} else {
		builder.WriteString("nil")
	}
	_, _ = builder.WriteString("\n\n")
	return builder.String()
}

// SSEEventGenerator defines a function for creating SSEEvents.
type SSEEventGenerator func(context.Context, chan SSEEvent)

// SSEHandler wraps an SSEEventGenerator into an http.Handler with responses being rendered from a given prefix.
type SSEHandler struct {
	Location       *Location
	EventGenerator SSEEventGenerator
}

func (sh SSEHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	stream := make(chan SSEEvent)
	defer close(stream)

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go sh.EventGenerator(ctx, stream)
	for event := range stream {
		_, err := fmt.Fprint(w, event.string(sh.Location))
		if err != nil {
			break
		}
		flusher.Flush()
	}
}
