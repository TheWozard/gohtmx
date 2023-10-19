package gohtmx

import (
	"io"
	"net/http"
)

type Raw string

func (r Raw) WriteTemplate(prefix string, w io.StringWriter) {
	_, _ = w.WriteString(string(r))
}

func (r Raw) LoadMux(_ string, _ *http.ServeMux) {
}

type Attribute struct {
	Name  string
	Value string
}

func (a Attribute) Write(w io.StringWriter) {
	_, _ = w.WriteString(` `)
	_, _ = w.WriteString(a.Name)
	if a.Value != "" {
		_, _ = w.WriteString(`="`)
		_, _ = w.WriteString(a.Value)
		_, _ = w.WriteString(`"`)
	}
}

type Tag struct {
	Name       string
	Attributes []Attribute
	Content    Component
}

func (t Tag) WriteTemplate(prefix string, w io.StringWriter) {
	_, _ = w.WriteString(`<`)
	_, _ = w.WriteString(t.Name)
	for _, attr := range t.Attributes {
		attr.Write(w)
	}
	_, _ = w.WriteString(`>`)
	if t.Content != nil {
		t.Content.WriteTemplate(prefix, w)
	}
	_, _ = w.WriteString(`</`)
	_, _ = w.WriteString(t.Name)
	_, _ = w.WriteString(`>`)
}

func (t Tag) LoadMux(prefix string, m *http.ServeMux) {
	t.Content.LoadMux(prefix, m)
}
