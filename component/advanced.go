package component

import (
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Collapsible struct {
	// Label is the text to be displayed as a label. "" removes the label.
	Label string
	// Contents is the content of the FieldSet. This may be a Fragment.
	Contents Component
	// Open is the initial state of the collapsed field when the page is loaded.
	Open bool
	// TODO:
	UseHidden bool
}

func (c Collapsible) Path(open bool) string {
	return actionPathPrefix + "/" + c.Label + "/" + strconv.FormatBool(open)
}

func (c Collapsible) WriteTemplate(w io.StringWriter) {
	c.writeTemplate(w, c.Open)
}

func (c Collapsible) writeTemplate(w io.StringWriter, open bool) {
	frag := Fragment{Tag{
		Name: "button",
		Attributes: []Attribute{
			{Name: "hx-post", Value: c.Path(!open)},
			{Name: "hx-target", Value: "#" + c.Label},
		},
		Content: Raw(c.Label),
	}}
	if open {
		frag = append(frag, c.Contents)
	}
	Tag{
		"div",
		[]Attribute{
			{Name: "id", Value: c.Label},
		},
		frag,
	}.WriteTemplate(w)
}

func (c Collapsible) LoadMux(mux *http.ServeMux) {
	// Cache the templates
	var open strings.Builder
	var closed strings.Builder
	c.writeTemplate(&open, true)
	c.writeTemplate(&closed, false)
	openCache := open.String()
	closedCache := closed.String()

	mux.HandleFunc(c.Path(true), func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(openCache))
	})
	mux.HandleFunc(c.Path(false), func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(closedCache))
	})
}

type Selection struct {
	Label   string
	Options []SelectionOption
}

type SelectionOption struct {
	Name    string
	Content Component
}

func (s Selection) Path() string {
	return actionPathPrefix + "/" + s.Label
}

func (s Selection) WriteTemplate(w io.StringWriter) {
	s.writeTemplate(w, s.Options[0])
}

func (s Selection) writeTemplate(w io.StringWriter, option SelectionOption) {
	options := Fragment{}
	for _, o := range s.Options {
		tag := Tag{
			Name: "option", Attributes: []Attribute{
				{Name: "value", Value: o.Name},
			}, Content: Raw(o.Name),
		}
		if o.Name == option.Name {
			tag.Attributes = append(tag.Attributes, Attribute{Name: "selected"})
		}
		options = append(options, tag)
	}
	frag := Fragment{Tag{
		Name: "select",
		Attributes: []Attribute{
			{Name: "hx-put", Value: s.Path()},
			{Name: "hx-target", Value: "#" + s.Label},
			{Name: "hx-trigger", Value: "change"},
			{Name: "name", Value: s.Label},
		},
		Content: options,
	}, option.Content}
	Tag{
		"div",
		[]Attribute{
			{Name: "id", Value: s.Label},
		},
		frag,
	}.WriteTemplate(w)
}

func (s Selection) LoadMux(mux *http.ServeMux) {
	templates := map[string]*template.Template{}
	for _, option := range s.Options {
		var body strings.Builder
		s.writeTemplate(&body, option)
		templates[option.Name], _ = template.New("resp").Parse(body.String())
	}

	mux.HandleFunc(s.Path(), func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		value := r.Form[s.Label][0]
		if temp, ok := templates[value]; ok {
			_ = temp.Execute(w, nil)
		}
	})
}

type Stream struct {
	Name              string
	Events            []string
	SSEEventGenerator SSEEventGenerator
}

func (s Stream) Path() string {
	return severSideEventPrefix + "/" + s.Name
}

func (s Stream) WriteTemplate(w io.StringWriter) {
	frag := Fragment{}
	for _, event := range s.Events {
		frag = append(frag, Tag{"div", []Attribute{{"sse-swap", event}, {"hx-swap", "afterbegin"}}, nil})
	}
	Tag{
		Name: "div",
		Attributes: []Attribute{
			{Name: "hx-ext", Value: "sse"},
			{Name: "sse-connect", Value: s.Path()},
		},
		Content: frag,
	}.WriteTemplate(w)
}

func (s Stream) LoadMux(mux *http.ServeMux) {
	mux.Handle(s.Path(), SSEHandler{
		EventConnector: s.SSEEventGenerator,
	})
}
