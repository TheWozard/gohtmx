package component

import (
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
	WriteTemplate(io.StringWriter)
	LoadMux(*http.ServeMux)
}

// Fragment defines a slice of Component that can be used as a single Component.
type Fragment []Component

func (f Fragment) WriteTemplate(w io.StringWriter) {
	for _, frag := range f {
		frag.WriteTemplate(w)
	}
}

func (f Fragment) LoadMux(m *http.ServeMux) {
	for _, frag := range f {
		frag.LoadMux(m)
	}
}

func BuildBytes(c Component) []byte {
	var close strings.Builder
	c.WriteTemplate(&close)
	return []byte(close.String())
}
