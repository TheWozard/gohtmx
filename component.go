package gohtmx

import (
	"html/template"
	"io"
	"strings"

	"github.com/gorilla/mux"
)

const (
	actionPathPrefix     = "action"
	severSideEventPrefix = "sse"
)

// Component defines the requirements of a component in the UI. There are two paths that all components are mounted through.
// Components are not self rendering. Something must execute the returned template, often a Document.
type Component interface {
	Interaction
	// LoadTemplate defines what a component should be rendered as. Rendering of itself is done using
	// html/template and as such can include template information.
	// If custom data is needed to render this Component, use the DataLoader to make data available, not through LoadMux function.
	// ex: d.AddNamedAtPath("name", l.Path("path"), func() (any, bool) {})
	LoadTemplate(l *Location, d *TemplateDataLoader, w io.StringWriter)
}

type Interaction interface {
	// LoadMux defines how a component is interactive. To create interactivity add a handler to the passed mux
	// ex: m.HandleFunc(l.Path("path"), func(w http.ResponseWriter, r *http.Request) {})
	// Any contents/child Components should have their LoadMux called during this function.
	LoadMux(l *Location, m *mux.Router)
}

// Location defines information about the current rendering location. All Data and Mux Routing should be based off of the Location.
type Location struct {
	// PathPrefix defines the current prefix for a component to build requests from.
	PathPrefix string
	// DataPrefix defines the current prefix for relative data loading.
	DataPrefix string

	// Template defines the base template to be cloned for LoadTemplate.
	// This can be used to add custom functions or common templates to child components.
	TemplateBase *template.Template
}

func (l *Location) Sanitize() {
	l.PathPrefix = strings.TrimRight(l.PathPrefix, "/")
	l.PathPrefix = strings.TrimRight(l.PathPrefix, ".")
}

func (l *Location) Path(segments ...string) string {
	if len(segments) > 0 {
		return l.PathPrefix + "/" + strings.Join(segments, "/")
	}
	return l.PathPrefix
}

func (l *Location) Data(segments ...string) string {
	return l.DataPrefix + "." + strings.Join(segments, ".")
}

func (l *Location) AtPath(segments ...string) *Location {
	return &Location{
		PathPrefix:   l.Path(segments...),
		DataPrefix:   l.DataPrefix,
		TemplateBase: l.TemplateBase,
	}
}

func (l *Location) AtData(segments ...string) *Location {
	return &Location{
		PathPrefix:   l.PathPrefix,
		DataPrefix:   l.Data(segments...),
		TemplateBase: l.TemplateBase,
	}
}

func (l *Location) BuildString(c Component) (string, *TemplateDataLoader) {
	var builder strings.Builder
	var loader TemplateDataLoader
	c.LoadTemplate(l, &loader, &builder)
	return builder.String(), &loader
}

// BuildTemplate builds a given component template at this location to a template using the TemplateBase.
func (l *Location) BuildTemplate(c Component) (*template.Template, *TemplateDataLoader) {
	tmp, err := l.TemplateBase.Clone()
	if err != nil {
		panic(err)
	}
	raw, loader := l.BuildString(c)
	tmp, err = tmp.Parse(raw)
	if err != nil {
		panic(err)
	}
	return tmp, loader
}

// BuildTemplateHandler builds a given component template at this location to a ready to use handler.
func (l *Location) BuildTemplateHandler(c Component) *TemplateHandler {
	template, loader := l.BuildTemplate(c)
	return &TemplateHandler{
		Template: template, Loader: loader,
	}
}

// Fragment defines a slice of Components that can be used as a single Component.
type Fragment []Component

func (f Fragment) LoadTemplate(l *Location, d *TemplateDataLoader, w io.StringWriter) {
	for _, frag := range f {
		frag.LoadTemplate(l, d, w)
	}
}

func (f Fragment) LoadMux(l *Location, m *mux.Router) {
	for _, frag := range f {
		frag.LoadMux(l, m)
	}
}

func orDefault(value, def string) string {
	if value != "" {
		return value
	}
	return def
}
