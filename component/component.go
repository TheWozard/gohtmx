package component

import (
	"html/template"
	"io"
	"net/http"
	"strings"
)

const (
	actionPathPrefix     = "/action"
	severSideEventPrefix = "/sse"
)

// Component defines the requirements of a component of the UI.
type Component interface {
	WriteTemplate(prefix string, w io.StringWriter)
	LoadMux(prefix string, m *http.ServeMux)
}

// Fragment defines a slice of Component that can be used as a single Component.
type Fragment []Component

func (f Fragment) WriteTemplate(prefix string, w io.StringWriter) {
	for _, frag := range f {
		frag.WriteTemplate(prefix, w)
	}
}

func (f Fragment) LoadMux(prefix string, m *http.ServeMux) {
	for _, frag := range f {
		frag.LoadMux(prefix, m)
	}
}

// BuildBytes convenience function for converting a Component to bytes at a given prefix.
func BuildBytes(prefix string, c Component) []byte {
	var builder strings.Builder
	c.WriteTemplate(prefix, &builder)
	return []byte(builder.String())
}

func BuildTemplate(name, prefix string, c Component) *template.Template {
	var builder strings.Builder
	c.WriteTemplate(prefix, &builder)
	tmp, err := template.New(name).Funcs(template.FuncMap{
		"hasPrefix": strings.HasPrefix,
	}).Parse(builder.String())
	if err != nil {
		panic(err)
	}
	return tmp
}
