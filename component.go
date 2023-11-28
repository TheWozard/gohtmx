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

// Component defines the requirements of a component of the UI
type Component interface {
	// LoadTemplate defines what a component should be rendered as. Rendering itself is done using
	// html/template and as such can include template information.
	LoadTemplate(l *Location, w io.StringWriter)
	// LoadMux defines how a component is interactive. Often LoadTemplate is called from LoadMux so
	// a component can create a template of itself.
	LoadMux(l *Location, m *mux.Router)
}

// Location defines information about the current rendering location.
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
	return l.PathPrefix + "/" + strings.Join(segments, "/")
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

// BuildString builds a given component template at this location as a string.
func (l *Location) BuildString(c Component) string {
	var builder strings.Builder
	c.LoadTemplate(l, &builder)
	return builder.String()
}

// BuildString builds a given component template at this location as bytes.
func (l *Location) BuildBytes(c Component) []byte {
	return []byte(l.BuildString(c))
}

// BuildTemplate builds a given component template at this location to a template using the TemplateBase.
func (l *Location) BuildTemplate(c Component) *template.Template {
	tmp, err := l.TemplateBase.Clone()
	if err != nil {
		panic(err)
	}
	tmp, err = tmp.Parse(l.BuildString(c))
	if err != nil {
		panic(err)
	}
	return tmp
}

// Fragment defines a slice of Components that can be used as a single Component.
type Fragment []Component

func (f Fragment) LoadTemplate(l *Location, w io.StringWriter) {
	for _, frag := range f {
		frag.LoadTemplate(l, w)
	}
}

func (f Fragment) LoadMux(l *Location, m *mux.Router) {
	for _, frag := range f {
		frag.LoadMux(l, m)
	}
}
