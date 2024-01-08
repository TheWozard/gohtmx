package gohtmx

import (
	"fmt"
	"io"
	"strings"

	"github.com/gorilla/mux"
)

// Raw defines the most generic Component that directly renders the string as HTML
type Raw string

func (r Raw) LoadTemplate(l *Location, t *TemplateDataLoader, w io.StringWriter) {
	_, _ = w.WriteString(string(r))
}

func (r Raw) LoadMux(l *Location, m *mux.Router) {
}

// Attr defines an HTML tag attribute `<Name>="<Value>"“. If Value is empty, only `<Name>` is rendered.
type Attr struct {
	Name      string
	Value     string
	Enabled   bool
	Off       bool
	Condition string
}

func (a Attr) Empty() bool {
	return a.Value == "" && !a.Enabled && !a.Off
}

func (a Attr) String() string {
	if a.Condition != "" {
		return fmt.Sprintf(`{{if %s}}%s{{end}}`, a.Condition, a.core())
	}
	return a.core()
}

func (a Attr) core() string {
	if a.Value != "" {
		return fmt.Sprintf(`%s="%s"`, a.Name, a.Value)
	}
	if a.Off {
		return fmt.Sprintf(`%s="%s"`, a.Name, "off")
	}
	return a.Name
}

// Tag defines the lowest level HTML generic tag element.
type Tag struct {
	// Name of this tag.
	Name string
	// Attr defines the list of Attributes for this tag.
	Attrs []Attr
	// Content defines the contents this wraps.
	Content Component
}

func (t Tag) LoadTemplate(l *Location, d *TemplateDataLoader, w io.StringWriter) {
	_, _ = w.WriteString(`<`)
	_, _ = w.WriteString(t.Name)
	for _, attr := range t.Attrs {
		if !attr.Empty() {
			_, _ = w.WriteString(` `)
			_, _ = w.WriteString(attr.String())
		}
	}
	if t.Content != nil {
		_, _ = w.WriteString(`>`)
		t.Content.LoadTemplate(l, d, w)
		_, _ = w.WriteString(`</`)
		_, _ = w.WriteString(t.Name)
		_, _ = w.WriteString(`>`)
	} else {
		_, _ = w.WriteString(`/>`)
	}
}

func (t Tag) LoadMux(l *Location, m *mux.Router) {
	t.Content.LoadMux(l, m)
}

// --- Shorthand Components ---

// Div is a shorthand for a "div" Tag
type Div struct {
	ID      string
	Classes []string
	Style   []string
	Attr    []Attr

	Content Component
}

func (d Div) LoadTemplate(l *Location, t *TemplateDataLoader, w io.StringWriter) {
	Tag{
		Name: "div",
		Attrs: append(d.Attr,
			Attr{Name: "id", Value: d.ID},
			Attr{Name: "class", Value: strings.Join(d.Classes, " ")},
			Attr{Name: "style", Value: strings.Join(d.Style, ";")},
		),
		Content: d.Content,
	}.LoadTemplate(l, t, w)
}

func (d Div) LoadMux(l *Location, m *mux.Router) {
	if d.Content != nil {
		d.Content.LoadMux(l, m)
	}
}

// Label is a shorthand for a "label" Tag
type Label struct {
	ID      string
	Classes []string
	Style   []string
	Attr    []Attr

	Text string
}

func (l Label) LoadTemplate(c *Location, t *TemplateDataLoader, w io.StringWriter) {
	Tag{
		Name: "div",
		Attrs: append(l.Attr,
			Attr{Name: "id", Value: l.ID},
			Attr{Name: "class", Value: strings.Join(l.Classes, " ")},
			Attr{Name: "style", Value: strings.Join(l.Style, ";")},
		),
		Content: Raw(l.Text),
	}.LoadTemplate(c, t, w)
}

func (l Label) LoadMux(c *Location, m *mux.Router) {
	// NoOp
}

// Span is a shorthand for a "span" Tag
type Span struct {
	ID      string
	Classes []string
	Style   []string
	Attr    []Attr

	Content Component
}

func (s Span) LoadTemplate(l *Location, t *TemplateDataLoader, w io.StringWriter) {
	Tag{
		Name: "span",
		Attrs: append(s.Attr,
			Attr{Name: "id", Value: s.ID},
			Attr{Name: "class", Value: strings.Join(s.Classes, " ")},
			Attr{Name: "style", Value: strings.Join(s.Style, ";")},
		),
		Content: s.Content,
	}.LoadTemplate(l, t, w)
}

func (s Span) LoadMux(l *Location, m *mux.Router) {
	s.Content.LoadMux(l, m)
}

// Nav is a shorthand for a "div" Tag
type Nav struct {
	ID      string
	Classes []string
	Style   []string

	Content Component
}

func (n Nav) LoadTemplate(l *Location, t *TemplateDataLoader, w io.StringWriter) {
	Tag{
		Name: "nav",
		Attrs: []Attr{
			{Name: "id", Value: n.ID},
			{Name: "class", Value: strings.Join(n.Classes, " ")},
			{Name: "style", Value: strings.Join(n.Style, ";")},
			{Name: "role", Value: "navigation"},
		},
		Content: n.Content,
	}.LoadTemplate(l, t, w)
}

func (n Nav) LoadMux(l *Location, m *mux.Router) {
	n.Content.LoadMux(l, m)
}

type Button struct {
	ID       string
	Classes  []string
	Style    []string
	Disabled bool
	Attr     []Attr

	OnClick func() Component

	Content Component
}

// func (b Button) LoadTemplate(l *Location, t *TemplateDataLoader, w io.StringWriter) {
// 	attr := b.Attr
// 	if b.OnClick != nil {
// 		attr = append(attr,
// 			Attr{Name: "hx-get", Value: l.Path(b.ID)},
// 			Attr{Name: "hx-swap", Value: "none"},
// 		)
// 	}
// 	Tag{
// 		Name: "button",
// 		Attrs: removeEmptyAttributes(append(attr,
// 			Attr{Name: "id", Value: b.ID},
// 			Attr{Name: "class", Value: strings.Join(b.Classes, " ")},
// 			Attr{Name: "style", Value: strings.Join(b.Style, ";")},
// 			Attr{Name: "disabled", Enabled: b.Disabled},
// 		)),
// 		Content: b.Content,
// 	}.LoadTemplate(l, t, w)
// }

// func (b Button) LoadMux(l *Location, m *mux.Router) {
// 	if b.OnClick != nil {
// 		m.Handle(l.Path(b.ID), ActionHandler{Location: l, Action: b.OnClick})
// 	}
// 	b.Content.LoadMux(l, m)
// }

type A struct {
	ID      string
	Classes []string
	Style   []string
	Attr    []Attr
	Href    string

	Content Component
}

func (a A) LoadTemplate(l *Location, t *TemplateDataLoader, w io.StringWriter) {
	Tag{
		Name: "input",
		Attrs: append(a.Attr,
			Attr{Name: "id", Value: a.ID},
			Attr{Name: "class", Value: strings.Join(a.Classes, " ")},
			Attr{Name: "style", Value: strings.Join(a.Style, ";")},
			Attr{Name: "href", Value: a.Href},
		),
		Content: a.Content,
	}.LoadTemplate(l, t, w)
}

func (a A) LoadMux(l *Location, m *mux.Router) {
	a.Content.LoadMux(l, m)
}

type UL struct {
	ID      string
	Classes []string
	Style   []string

	Contents Fragment
}

func (u UL) LoadTemplate(l *Location, t *TemplateDataLoader, w io.StringWriter) {
	wrapped := make(Fragment, len(u.Contents))
	for i, content := range u.Contents {
		if _, ok := content.(LI); !ok {
			wrapped[i] = LI{Content: content}
		}
	}
	Tag{
		Name: "ul",
		Attrs: []Attr{
			{Name: "id", Value: u.ID},
			{Name: "class", Value: strings.Join(u.Classes, " ")},
			{Name: "style", Value: strings.Join(u.Style, ";")},
		},
		Content: wrapped,
	}.LoadTemplate(l, t, w)
}

func (u UL) LoadMux(l *Location, m *mux.Router) {
	u.Contents.LoadMux(l, m)
}

type LI struct {
	ID      string
	Classes []string
	Style   []string

	Content Component
}

func (li LI) LoadTemplate(l *Location, t *TemplateDataLoader, w io.StringWriter) {
	Tag{
		Name: "li",
		Attrs: []Attr{
			{Name: "id", Value: li.ID},
			{Name: "class", Value: strings.Join(li.Classes, " ")},
			{Name: "style", Value: strings.Join(li.Style, ";")},
		},
		Content: li.Content,
	}.LoadTemplate(l, t, w)
}

func (li LI) LoadMux(l *Location, m *mux.Router) {
	li.Content.LoadMux(l, m)
}

type FieldSet struct {
	ID      string
	Classes []string
	Style   []string

	Content Component
}

func (fs FieldSet) LoadTemplate(l *Location, t *TemplateDataLoader, w io.StringWriter) {
	Tag{
		Name: "fieldset",
		Attrs: []Attr{
			{Name: "id", Value: fs.ID},
			{Name: "class", Value: strings.Join(fs.Classes, " ")},
			{Name: "style", Value: strings.Join(fs.Style, ";")},
		},
		Content: fs.Content,
	}.LoadTemplate(l, t, w)
}

func (fs FieldSet) LoadMux(l *Location, m *mux.Router) {
	fs.Content.LoadMux(l, m)
}

type Legend struct {
	ID      string
	Classes []string
	Style   []string

	Content Component
}

func (le Legend) LoadTemplate(l *Location, t *TemplateDataLoader, w io.StringWriter) {
	Tag{
		Name: "fieldset",
		Attrs: []Attr{
			{Name: "id", Value: le.ID},
			{Name: "class", Value: strings.Join(le.Classes, " ")},
			{Name: "style", Value: strings.Join(le.Style, ";")},
		},
		Content: le.Content,
	}.LoadTemplate(l, t, w)
}

func (fs Legend) LoadMux(l *Location, m *mux.Router) {
	fs.Content.LoadMux(l, m)
}
