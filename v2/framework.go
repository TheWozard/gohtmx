package gohtmx

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/TheWozard/gohtmx/v2/core"
	"github.com/go-chi/chi/v5"
)

var ErrCannotTemplate = fmt.Errorf("templating is not enabled")

func NewDefaultFramework() *Framework {
	return &Framework{
		PathPrefix: "/",
		DataPrefix: ".",
		Mux:        chi.NewMux(),
		Template:   template.New("content"),
		Generator:  core.NewDefaultGenerator(),
	}
}

// Framework defines the process of loading interactive components of the page to be served.
// Framework also acts as an http.Handler to serve the loaded content.
type Framework struct {
	// PathPrefix defines the current prefix for a component to build requests from.
	// TODO: this could maybe be integrated with the MUX layer to reduce the need for actual path prefixing.
	PathPrefix string
	// DataPrefix defines the current prefix for relative data loading.
	DataPrefix string

	// TODO: Interface
	Mux *chi.Mux

	// TODO: Interface?
	Template *template.Template

	Generator core.Generator
}

func (f *Framework) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.Mux.ServeHTTP(w, r)
}

// -- Features --

func (c *Framework) IsInteractive() bool {
	return c.Mux != nil
}

func (c *Framework) CanTemplate() bool {
	return c.Template != nil
}

// -- Location ---

func (c *Framework) Path(segments ...string) string {
	if len(segments) > 0 {
		return strings.TrimRight(c.PathPrefix, "/") + "/" + strings.TrimLeft(strings.Join(segments, "/"), "/")
	}
	return c.PathPrefix
}

func (c *Framework) Data(segments ...string) string {
	return c.DataPrefix + "." + strings.Join(segments, ".")
}

func (c *Framework) AtPath(segments ...string) *Framework {
	return &Framework{
		PathPrefix: c.Path(segments...),
		DataPrefix: c.DataPrefix,
		Mux:        c.Mux,
		Template:   c.Template,
	}
}

func (c *Framework) AtData(segments ...string) *Framework {
	return &Framework{
		PathPrefix: c.PathPrefix,
		DataPrefix: c.Data(segments...),
		Mux:        c.Mux,
		Template:   c.Template,
	}
}

func (c *Framework) WithTemplate(t *template.Template) *Framework {
	return &Framework{
		PathPrefix: c.PathPrefix,
		DataPrefix: c.DataPrefix,
		Mux:        c.Mux,
		Template:   t,
	}
}

func (c *Framework) Slim() *Framework {
	return &Framework{
		PathPrefix: c.PathPrefix,
		DataPrefix: c.DataPrefix,
	}
}

// -- Interactions --

type Middleware func(http.Handler) http.Handler

func (c *Framework) Use(middleware Middleware) {
	if c == nil || !c.IsInteractive() || middleware == nil {
		return
	}
	c.Mux.Use(middleware)
}

// AddInteraction adds an interaction at the passed path. This is a POST request at the relative of this context.
func (c *Framework) AddInteraction(handler http.Handler) {
	if c == nil || !c.IsInteractive() || handler == nil {
		return
	}
	c.Mux.Mount(c.Path(), handler)
}

// AddInteractionFunc adds an interaction at the passed path. This is a POST request at the relative of this context.
func (c *Framework) AddInteractionFunc(handler http.HandlerFunc) {
	c.AddInteraction(handler)
}

// AddComponentInteraction adds an interaction that specifically returns a fixed component.
func (c *Framework) AddComponentInteraction(component Component) error {
	if c == nil || !c.IsInteractive() || component == nil {
		return nil
	}
	buffer := bytes.NewBuffer(nil)
	// WithTemplate(nil) disables templating. This allows rendered components to note the lack of templating happening on this component.
	err := component.Init(c.WithTemplate(nil), buffer)
	if err != nil {
		return fmt.Errorf("failed to render component for component interaction: %w", err)
	}
	c.AddInteractionFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(buffer.Bytes())
	})
	return nil
}

func (c *Framework) AddTemplateInteraction(component Component) error {
	if c == nil || !c.IsInteractive() || !c.CanTemplate() || component == nil {
		return nil
	}
	handler, err := c.NewTemplateHandler(component)
	if err != nil {
		return fmt.Errorf("failed to create template handler: %w", err)
	}
	c.AddInteraction(handler)
	return nil
}

func (f *Framework) NewTemplateHandler(component Component) (*TemplateHandler, error) {
	return NewTemplateHandler(f, component)
}
