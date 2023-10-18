package component

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

func (i Input) WriteTemplate(w io.StringWriter) {
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
	}.WriteTemplate(w)
}

func (i Input) LoadMux(_ *http.ServeMux) {
}

type FieldSet struct {
	// Label is the text to be displayed as a label. "" removes the label.
	Label string
	// Contents is the content of the FieldSet. This may be a Fragment.
	Contents Component
}

func (fs FieldSet) WriteTemplate(w io.StringWriter) {
	frag := Fragment{fs.Contents}
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
	}.WriteTemplate(w)
}

func (fs FieldSet) LoadMux(_ *http.ServeMux) {
}

type Button struct {
	Label   string
	Classes []string
	Action  http.Handler
}

func (b Button) Path() string {
	return fmt.Sprintf("%s/%s", actionPathPrefix, b.Label)
}

func (b Button) WriteTemplate(w io.StringWriter) {
	Tag{"button", []Attribute{
		{Name: "class", Value: strings.Join(b.Classes, " ")},
		{Name: "hx-get", Value: b.Path()},
	}, nil}.WriteTemplate(w)
}

func (b Button) LoadMux(m *http.ServeMux) {
	m.Handle(b.Path(), b.Action)

	var closeCache = BuildBytes(b)
	m.HandleFunc(fmt.Sprintf("%s/close", b.Path()), func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(closeCache)
	})
}
