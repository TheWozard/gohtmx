package gohtmx

import (
	"io"

	"github.com/gorilla/mux"
)

// At modifies the Prefix of the building template
type At struct {
	// Name is the action that is to be taken for this template.
	Location *Location
	// Content defines the Component this wraps.
	Content Component
}

func (a At) LoadTemplate(l *Location, t *TemplateDataLoader, w io.StringWriter) {
	a.Content.LoadTemplate(a.Location, t, w)
}

func (a At) LoadMux(l *Location, m *mux.Router) {
	a.Content.LoadMux(a.Location, m)
}
