package gohtmx

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"

	"github.com/TheWozard/gohtmx/element"
)

func NewPage() *Page {
	return &Page{
		PathPrefix: "/",
		Index:      map[string]Request{},
		Template:   template.New("content"),
		Generator:  NewDefaultGenerator(),
	}
}

// Page defines a single page application.
type Page struct {
	// PathPrefix defines the current prefix for a component to build requests from.
	PathPrefix string
	// Index collects all the interactive components of the page.
	Index map[string]Request
	// The template to use for rendering components.
	Template *template.Template
	// Generator provides generated content for initializing elements.
	Generator Generator
}

// Validate will validate all elements in the page. Each path is validated individually, and paths with errors are
// returned in a map. If no errors are found, nil is returned.
func (p *Page) Validate() map[string]error {
	errors := make(map[string]error, len(p.Index))
	for _, path := range p.paths() {
		e := p.Index[path].Validate()
		if e != nil {
			errors[path] = e
		}
	}
	if len(errors) > 0 {
		return errors
	}
	return nil
}

// Render will render all elements in the page to a map of byte slices. Each path is rendered individually, and all
// paths with data are returned in a map.
func (p *Page) Render() (map[string]string, error) {
	t := make(map[string]string, len(p.Index))
	errs := make([]error, 0, len(p.Index))
	for name, req := range p.Index {
		raw, e := req.Render()
		if e != nil {
			errs = append(errs, e)
		}
		if len(raw) > 0 {
			t[name] = string(raw)
		}
	}
	return t, errors.Join(errs...)
}

// Build creates a new http.Handler for the entire page.
func (p *Page) Build() (http.Handler, error) {
	paths := p.paths()
	htmx := http.NewServeMux()
	var page http.Handler
	for _, path := range paths {
		request := p.Index[path]
		err := request.Validate()
		if err != nil {
			return nil, fmt.Errorf("failed to validate request '%s': %w", path, err)
		}
		raw, err := request.Render()
		if err != nil {
			return nil, fmt.Errorf("failed to render request '%s': %w", path, err)
		}
		name := p.Generator.NewID("template")
		p.Template, err = p.Template.New(name).Parse(string(raw))
		if err != nil {
			return nil, fmt.Errorf("failed to parse template '%s': %w", path, err)
		}

		handler := request.Wrap(&TemplateHandler{
			Template: p.Template,
			Name:     name,
		})
		if path == "/" {
			page = handler
		} else {
			htmx.Handle(path, handler)
		}
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

func (p *Page) paths() []string {
	paths := make([]string, 0, len(p.Index))
	for path := range p.Index {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

// Init initializes a component and returns its created element. This handles the error without having to return it.
// This is a convenience method for Components to use to initialize other components/contents.
func (p *Page) Init(c Component) element.Element {
	if p != nil && c != nil {
		e, err := c.Init(p)
		if err != nil {
			// We add the error into the Element Tree. This will get picked up during the validation stage.
			return element.RawError{Err: err}
		}
		return e
	}
	return nil
}

// -- Location ---

// Path returns the full path of the page with any additional segments appended.
func (p *Page) Path(segments ...string) string {
	if len(segments) > 0 {
		return strings.TrimRight(p.PathPrefix, "/") + "/" + strings.TrimLeft(strings.Join(segments, "/"), "/")
	}
	return p.PathPrefix
}

// AtPath returns a new Page with the with the segments appended to the current path.
// Resources are shared between both new and old page, only the path is different.
func (p *Page) AtPath(segments ...string) *Page {
	return &Page{
		PathPrefix: p.Path(segments...),
		Generator:  p.Generator,
		Index:      p.Index,
		Template:   p.Template,
	}
}

// -- Interactions --

type Middleware func(http.Handler) http.Handler

// Use adds middleware to the request at this pages current path.
func (p *Page) Use(middleware ...Middleware) {
	if p == nil || middleware == nil {
		return
	}
	request := p.Index[p.Path()]
	request.Middleware = append(request.Middleware, middleware...)
	p.Index[p.Path()] = request
}

type Handle func(*http.Request)

func (p *Page) Handle(h Handle) {
	if p == nil || h == nil {
		return
	}
	p.Use(HandlerMiddleware(h).Middleware)
}

// Add adds a component to the request at this pages current path. This is when a Component is initialized through Init
// into elements.
func (p *Page) Add(component Component) {
	if p == nil || component == nil {
		return
	}
	request := p.Index[p.Path()]
	request.Elements = append(request.Elements, p.Init(component))
	p.Index[p.Path()] = request
}

type HandlerMiddleware func(*http.Request)

func (h HandlerMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h(r)
		next.ServeHTTP(w, r)
	})
}

// Request defines a single interactive endpoint.
type Request struct {
	Elements   element.Fragment
	Middleware []Middleware
}

func (r Request) Validate() error {
	return r.Elements.Validate()
}

func (r Request) Render() ([]byte, error) {
	data := bytes.NewBuffer(nil)
	err := element.Fragment{element.Raw("{{$r := .request}}"), r.Elements}.Render(data)
	return data.Bytes(), err
}

func (r Request) Wrap(handler http.Handler) http.Handler {
	for _, middleware := range r.Middleware {
		handler = middleware(handler)
	}
	return handler
}
