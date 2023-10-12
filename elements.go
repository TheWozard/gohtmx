package gojsox

import (
	"io"
	"strings"
)

type writeable interface {
	Write(w io.StringWriter)
}

type fragment []writeable

func (f fragment) Write(w io.StringWriter) {
	for _, frag := range f {
		frag.Write(w)
	}
}

type raw string

func (r raw) Write(w io.StringWriter) {
	w.WriteString(string(r))
}

type attribute struct {
	name  string
	value string
}

func (a attribute) Write(w io.StringWriter) {
	w.WriteString(` `)
	w.WriteString(a.name)
	w.WriteString(`="`)
	w.WriteString(a.value)
	w.WriteString(`"`)
}

type tag struct {
	name       string
	attributes []attribute
	content    writeable
}

func (t tag) Write(w io.StringWriter) {
	w.WriteString(`<`)
	w.WriteString(t.name)
	for _, attr := range t.attributes {
		attr.Write(w)
	}
	w.WriteString(`>`)
	if t.content != nil {
		t.content.Write(w)
	}
	w.WriteString(`</`)
	w.WriteString(t.name)
	w.WriteString(`>`)
}

type formInput struct {
	label       string
	placeholder string
	path        string
	typ         string
	classes     []string
}

func (fi formInput) Write(w io.StringWriter) {
	var frag fragment

	if fi.label != "" {
		frag = append(frag, tag{
			"label",
			[]attribute{
				{"for", fi.path},
			},
			raw(fi.label),
		})
	}

	frag = append(frag, tag{
		"input",
		[]attribute{
			{"type", fi.typ},
			{"class", strings.Join(append(fi.classes, "form-control"), " ")},
			{"name", fi.path},
			{"placeholder", fi.placeholder},
			{"id", fi.path},
			{"value", "{{." + fi.path + "}}"},
		},
		nil,
	})

	tag{
		"div",
		[]attribute{
			{"class", "form-group"},
		},
		frag,
	}.Write(w)
}
