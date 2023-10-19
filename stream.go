package gohtmx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
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

func (s Stream) path(prefix string) string {
	return prefix + severSideEventPrefix + "/" + s.ID
}

func (s Stream) WriteTemplate(prefix string, w io.StringWriter) {
	Tag{
		"div",
		[]Attribute{
			{"id", s.ID},
			{"hx-ext", "sse"},
			{"sse-connect", s.path(prefix)},
			{"class", strings.Join(s.Classes, " ")},
		},
		s.Content,
	}.WriteTemplate(prefix, w)
}

func (s Stream) LoadMux(prefix string, mux *http.ServeMux) {
	mux.Handle(s.path(prefix), SSEHandler{
		Prefix:         prefix,
		EventGenerator: s.SSEEventGenerator,
	})
}

type SwapType string

const (
	// SwapReplace swaps the Content of a Component with the new data.
	SwapReplace = "innerHTML"
	// SwapAppend adds the new Component to the end of the current Content.
	SwapAppend = "beforeend"
	// SwapPrepend adds the new Component before the current Content.
	SwapPrepend = "afterbegin"
)

// StreamTarget is a Component that represents a target location to render an event from a Stream Component
// This Component must be wrapped by a Stream Component in order for events to be received.
type StreamTarget struct {
	// Event is the name of the events this component will receive.
	Events []string
	// Swap defines how events are added to the target. Defaults to SwapReplace
	Swap SwapType
	// Classes adds classes to the wrapping div element.
	Classes []string
	// Content defines the Component this wraps.
	Content Component
}

func (st StreamTarget) swap() string {
	if st.Swap != "" {
		return string(st.Swap)
	}
	return string(SwapReplace)
}

func (st StreamTarget) events() []string {
	if len(st.Events) == 0 {
		return []string{"message"}
	}
	return st.Events
}

func (st StreamTarget) WriteTemplate(prefix string, w io.StringWriter) {
	Tag{
		"div",
		[]Attribute{
			{"sse-swap", strings.Join(st.events(), ",")},
			{"hx-swap", st.swap()},
			{"class", strings.Join(st.Classes, " ")},
		},
		st.Content,
	}.WriteTemplate(prefix, w)
}

func (st StreamTarget) LoadMux(prefix string, m *http.ServeMux) {
	st.Content.LoadMux(prefix, m)
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

func (se SSEEvent) string(prefix string) string {
	var builder strings.Builder
	_, _ = builder.WriteString(fmt.Sprintf("event: %s\n", se.event()))
	_, _ = builder.WriteString("data: ")
	if se.Data != nil {
		if se.Prefix != "" {
			se.Data.WriteTemplate(se.Prefix, &builder)
		} else {
			se.Data.WriteTemplate(prefix, &builder)
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
	Prefix         string
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
		_, err := fmt.Fprint(w, event.string(sh.Prefix))
		if err != nil {
			break
		}
		flusher.Flush()
	}
}
