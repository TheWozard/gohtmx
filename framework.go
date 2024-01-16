package gohtmx

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/TheWozard/gohtmx/core"
	"github.com/go-chi/chi/v5"
)

func NewDefaultFramework() *Framework {
	return &Framework{
		PathPrefix: "/",
		Mux:        chi.NewMux(),
		Template:   template.New("content"),
		Generator:  core.NewDefaultGenerator(),
	}
}

// Framework defines the process of loading interactive components of the page to be served.
// Framework also acts as an http.Handler to serve the loaded content.
type Framework struct {
	// PathPrefix defines the current prefix for a component to build requests from.
	PathPrefix string

	// TODO: Interface
	Mux  *chi.Mux
	Page http.Handler

	// The template to use for rendering components.
	Template *template.Template

	Generator core.Generator
}

func (f *Framework) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// To create a SPA, we assume any non-HX-Request is a page request.
	if r.Header.Get("HX-Request") != "true" && f.Page != nil {
		f.Page.ServeHTTP(w, r)
	} else {
		f.Mux.ServeHTTP(w, r)
	}
}

// -- Features --

func (f *Framework) IsInteractive() bool {
	return f.Mux != nil
}

func (f *Framework) CanTemplate() bool {
	return f.Template != nil
}

// -- Location ---

func (f *Framework) Path(segments ...string) string {
	if len(segments) > 0 {
		return strings.TrimRight(f.PathPrefix, "/") + "/" + strings.TrimLeft(strings.Join(segments, "/"), "/")
	}
	return f.PathPrefix
}

func (f *Framework) AtPath(segments ...string) *Framework {
	return &Framework{
		PathPrefix: f.Path(segments...),
		Generator:  f.Generator,
		Mux:        f.Mux,
		Template:   f.Template,
	}
}

func (f *Framework) WithTemplate(t *template.Template) *Framework {
	return &Framework{
		PathPrefix: f.PathPrefix,
		Generator:  f.Generator,
		Mux:        f.Mux,
		Template:   t,
	}
}

func (f *Framework) NoMux() *Framework {
	return &Framework{
		PathPrefix: f.PathPrefix,
		Generator:  f.Generator,
		Template:   f.Template,
	}
}

func (f *Framework) Slim() *Framework {
	return &Framework{
		PathPrefix: f.PathPrefix,
		Generator:  f.Generator,
	}
}

// -- Interactions --

type Middleware func(http.Handler) http.Handler

func (f *Framework) Use(middleware Middleware) {
	if f == nil || !f.IsInteractive() || middleware == nil {
		return
	}
	f.Mux.Use(middleware)
}

// AddInteraction adds an interaction at the passed path. This is a POST request at the relative of this context.
func (f *Framework) AddInteraction(handler http.Handler) {
	if f == nil || !f.IsInteractive() || handler == nil {
		return
	}
	path := f.Path()
	f.Mux.Mount(path, handler)
	if path == "/" {
		f.Page = handler
	}
}

// AddInteractionFunc adds an interaction at the passed path. This is a POST request at the relative of this context.
func (f *Framework) AddInteractionFunc(handler http.HandlerFunc) {
	f.AddInteraction(handler)
}

// AddComponentInteraction adds an interaction that specifically returns a fixed component.
func (f *Framework) AddComponentInteraction(component Component, handlers ...http.HandlerFunc) error {
	if f == nil || !f.IsInteractive() || component == nil {
		return nil
	}
	buffer := bytes.NewBuffer(nil)
	// WithTemplate(nil) disables templating. This allows rendered components to note the lack of templating happening on this component.
	err := component.Init(f.WithTemplate(nil), buffer)
	if err != nil {
		return fmt.Errorf("failed to render component for component interaction: %w", err)
	}
	f.AddInteractionFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, handler := range handlers {
			handler(w, r)
		}
		_, _ = w.Write(buffer.Bytes())
	})
	return nil
}

func (f *Framework) AddTemplateInteraction(component Component, handlers ...http.HandlerFunc) error {
	if f == nil || !f.IsInteractive() || !f.CanTemplate() || component == nil {
		return nil
	}
	handler, err := f.NewTemplateHandler(component)
	if err != nil {
		return fmt.Errorf("failed to create template handler: %w", err)
	}
	f.AddInteractionFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, handler := range handlers {
			handler(w, r)
		}
		handler.ServeHTTP(w, r)
	})
	return nil
}

func (f *Framework) NewTemplateHandler(component Component) (*TemplateHandler, error) {
	return NewTemplateHandler(f, component)
}
