package component

import (
	"io"
	"net/http"
)

type Raw string

func (r Raw) WriteTemplate(w io.StringWriter) {
	_, _ = w.WriteString(string(r))
}

func (r Raw) LoadMux(_ *http.ServeMux) {
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

func (t Tag) WriteTemplate(w io.StringWriter) {
	_, _ = w.WriteString(`<`)
	_, _ = w.WriteString(t.Name)
	for _, attr := range t.Attributes {
		attr.Write(w)
	}
	_, _ = w.WriteString(`>`)
	if t.Content != nil {
		t.Content.WriteTemplate(w)
	}
	_, _ = w.WriteString(`</`)
	_, _ = w.WriteString(t.Name)
	_, _ = w.WriteString(`>`)
}

func (t Tag) LoadMux(m *http.ServeMux) {
	t.Content.LoadMux(m)
}

type TemplateDefinition struct {
	Name    string
	Content Component
}

func (td TemplateDefinition) WriteTemplate(w io.StringWriter) {
	_, _ = w.WriteString(`{{define "`)
	_, _ = w.WriteString(td.Name)
	_, _ = w.WriteString(`"}}`)
	if td.Content != nil {
		td.Content.WriteTemplate(w)
	}
	_, _ = w.WriteString(`{{end}}`)
}

func (td TemplateDefinition) LoadMux(m *http.ServeMux) {
	td.Content.LoadMux(m)
}
