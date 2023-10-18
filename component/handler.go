package component

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

func LoadComponent(mux *http.ServeMux, c Component) error {
	var builder strings.Builder
	c.WriteTemplate(&builder)
	body, err := template.New("body").Parse(builder.String())
	if err != nil {
		return err
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_ = body.Execute(w, DataFromContext(r.Context()))
	})
	c.LoadMux(mux)
	return nil
}

type SingleFieldHandler struct {
	Name    string
	Handler func(value string, w http.ResponseWriter, r *http.Request)
}

func (sfh SingleFieldHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "error reading form data", http.StatusInternalServerError)
		return
	}
	value, ok := r.Form[sfh.Name]
	if !ok {
		http.Error(w, fmt.Sprintf("unable to locate form value %s", sfh.Name), http.StatusInternalServerError)
		return
	}
	if len(value) != 1 {
		http.Error(w, fmt.Sprintf("unexpected form value for %s: %s", sfh.Name, strings.Join(value, ",")), http.StatusInternalServerError)
		return
	}
	sfh.Handler(value[0], w, r)
}

type SSEEvent struct {
	Event  string
	Data   string
	Append bool
}

func (se SSEEvent) EventOrDefault() string {
	if se.Event == "" {
		return "message"
	}
	return se.Event
}

func (se SSEEvent) String() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("event: %s\n", se.EventOrDefault()))
	builder.WriteString(fmt.Sprintf("data: %v", se.Data))
	builder.WriteString("\n\n")
	return builder.String()
}

type SSEEventGenerator func(context.Context, chan SSEEvent)

type SSEHandler struct {
	EventConnector SSEEventGenerator
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

	go sh.EventConnector(ctx, stream)
	for event := range stream {
		_, err := fmt.Fprint(w, event.String())
		if err != nil {
			break
		}
		flusher.Flush()
	}
}
