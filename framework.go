package gohtmx

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/TheWozard/gohtmx/core"
	"github.com/TheWozard/gohtmx/internal"
	"github.com/go-chi/chi/v5"
)

var (
	ErrInteractionExists = fmt.Errorf("interaction already exists")
	ErrNilComponent      = fmt.Errorf("component is required")
)

func NewDefaultFramework() *Framework {
	return &Framework{
		PathPrefix: "/",
		Index:      map[string]Interaction{},
		Template:   template.New("content"),
		Generator:  core.NewDefaultGenerator(),
	}
}

// Interaction defines a single location
type Interaction struct {
	Core       Component
	OutOfBand  []Component
	Middleware []Middleware
}

func (i Interaction) Component() Component {
	return Fragment(append(i.OutOfBand, i.Core))
}

func (i Interaction) Wrap(handler http.Handler) http.Handler {
	for _, middleware := range i.Middleware {
		handler = middleware(handler)
	}
	return handler
}

// Framework defines the process of loading interactive components of the page to be served.
// Framework also acts as an http.Handler to serve the loaded content.
type Framework struct {
	// PathPrefix defines the current prefix for a component to build requests from.
	PathPrefix string

	Index map[string]Interaction

	// The template to use for rendering components.
	Template *template.Template

	Generator core.Generator
}

func (f *Framework) Build() (http.Handler, error) {
	var err error
	var page http.Handler
	if index, ok := f.Index["/"]; ok {
		page, err = NewTemplateHandler(f, index.Component())
		if err != nil {
			return nil, err
		}
		page = index.Wrap(page)
	}
	htmx := chi.NewMux()
	var handler http.Handler
	for path, interaction := range f.Index {
		if path == "/" {
			continue
		}
		handler, err = NewTemplateHandler(f, interaction.Component())
		if err != nil {
			return nil, err
		}
		handler = interaction.Wrap(handler)
		htmx.Handle(path, handler)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// To create a SPA, we assume any non-HX-Request is a page request.
		if r.Header.Get("HX-Request") != "true" && page != nil {
			page.ServeHTTP(w, r)
		} else {
			htmx.ServeHTTP(w, r)
		}
	}), nil
}

// -- Features --

func (f *Framework) IsInteractive() bool {
	return f.Index != nil
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
		Index:      f.Index,
		Template:   f.Template,
	}
}

func (f *Framework) WithTemplate(t *template.Template) *Framework {
	return &Framework{
		PathPrefix: f.PathPrefix,
		Generator:  f.Generator,
		Index:      f.Index,
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
	interaction, ok := f.Index[f.Path()]
	if !ok {
		interaction = Interaction{}
	}
	interaction.Middleware = append(interaction.Middleware, middleware)
	f.Index[f.Path()] = interaction
}

func (f *Framework) AddInteraction(component Component) error {
	if f == nil || !f.IsInteractive() || component == nil {
		return nil
	}
	interaction, ok := f.Index[f.Path()]
	if !ok {
		interaction = Interaction{}
	}
	if interaction.Core != nil {
		return fmt.Errorf("interaction already exists at path %s: %w", f.Path(), ErrInteractionExists)
	}
	interaction.Core = MetaAtPath{Path: f.Path(), Content: component}
	f.Index[f.Path()] = interaction
	return nil
}

func (f *Framework) AddOutOfBand(component Component) {
	if f == nil || !f.IsInteractive() || component == nil {
		return
	}
	interaction, ok := f.Index[f.Path()]
	if !ok {
		interaction = Interaction{}
	}
	interaction.OutOfBand = append(interaction.OutOfBand, MetaAtPath{Path: f.Path(), Content: component})
	f.Index[f.Path()] = interaction
}

// -- Rendering --

func (f *Framework) Mono(component Component) (Component, error) {
	if component == nil {
		return nil, nil
	}
	data := bytes.NewBuffer(nil)
	err := component.Init(f, data)
	if err != nil {
		return nil, internal.ErrEnclosePath(err, "Mono")
	}
	return Raw(data.String()), nil
}

func (f *Framework) Render(component Component) (string, error) {
	if !f.CanTemplate() {
		return "", ErrCannotTemplate
	}
	if component == nil {
		return "", ErrNilComponent
	}
	data := bytes.NewBuffer(nil)
	err := component.Init(f, data)
	if err != nil {
		return "", err
	}
	name := f.Generator.NewGroupID("template")
	f.Template, err = f.Template.New(name).Parse(data.String())
	return name, err
}
