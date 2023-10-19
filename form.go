package gohtmx

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type InputType string

const (
	InputTypeText InputType = "text"
)

type Input struct {
	// Label is the text to be displayed as a label. "" removes the label.
	Label string
	// Placeholder is the text displayed in the input.
	Placeholder string
	// Path is the json path that this field represents.
	Path string
	// Type is the type of this input. Default: "text".
	Type InputType
	// Classes gives access to add html classes to the input tag.
	Classes []string
}

func (i Input) WriteTemplate(prefix string, w io.StringWriter) {
	typ := i.Type
	if i.Type == "" {
		typ = InputTypeText
	}

	var frag Fragment
	if i.Label != "" {
		frag = append(frag, Tag{
			Name: "label",
			Attributes: []Attribute{
				{Name: "for", Value: i.Path},
			},
			Content: Raw(i.Label),
		})
	}

	frag = append(frag, Tag{
		"input",
		[]Attribute{
			{Name: "type", Value: string(typ)},
			{Name: "class", Value: strings.Join(append(i.Classes, "form-control"), " ")},
			{Name: "name", Value: i.Path},
			{Name: "placeholder", Value: i.Placeholder},
			{Name: "id", Value: i.Path},
			{Name: "value", Value: "{{." + i.Path + "}}"},
		},
		nil,
	})

	Tag{
		"div",
		[]Attribute{
			{Name: "class", Value: "form-group"},
		},
		frag,
	}.WriteTemplate(prefix, w)
}

func (i Input) LoadMux(_ string, _ *http.ServeMux) {
}

type FieldSet struct {
	// Label is the text to be displayed as a label. "" removes the label.
	Label string
	// Content is the content of the FieldSet. This may be a Fragment.
	Content Component
}

func (fs FieldSet) WriteTemplate(prefix string, w io.StringWriter) {
	frag := Fragment{fs.Content}
	if fs.Label != "" {
		frag = append(Fragment{Tag{
			"legend",
			[]Attribute{},
			Raw(fs.Label),
		}}, frag...)
	}
	Tag{
		"fieldset",
		[]Attribute{},
		frag,
	}.WriteTemplate(prefix, w)
}

func (fs FieldSet) LoadMux(_ string, _ *http.ServeMux) {
}

type Button struct {
	Label   string
	Classes []string
	Action  http.Handler
}

func (b Button) path(prefix string) string {
	return fmt.Sprintf("%s/%s/%s", prefix, actionPathPrefix, b.Label)
}

func (b Button) WriteTemplate(prefix string, w io.StringWriter) {
	Tag{"button", []Attribute{
		{Name: "class", Value: strings.Join(b.Classes, " ")},
		{Name: "hx-get", Value: b.path(prefix)},
	}, nil}.WriteTemplate(prefix, w)
}

func (b Button) LoadMux(prefix string, m *http.ServeMux) {
	m.Handle(b.path(prefix), b.Action)

	var closeCache = BuildBytes(prefix, b)
	m.HandleFunc(fmt.Sprintf("%s/close", b.path(prefix)), func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(closeCache)
	})
}
