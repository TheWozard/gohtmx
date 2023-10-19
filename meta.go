package gohtmx

import (
	"io"
	"net/http"
)

// At modifies the Prefix of the building template
type At struct {
	// Name is the action that is to be taken for this template.
	Prefix string
	// Content defines the Component this wraps.
	Content Component
}

func (a At) WriteTemplate(prefix string, w io.StringWriter) {
	a.Content.WriteTemplate(a.Prefix, w)
}

func (a At) LoadMux(prefix string, m *http.ServeMux) {
	a.Content.LoadMux(a.Prefix, m)
}
