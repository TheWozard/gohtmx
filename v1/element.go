package gohtmx

import (
	"fmt"
	"io"
	"strings"
)

// Raw defines the most generic Component that directly renders the string as HTML
type Raw string

func (r Raw) Init(f *Framework, w io.Writer) error {
	_, err := w.Write([]byte(r))
	if err != nil {
		return fmt.Errorf("failed to write raw: %w", err)
	}
	return nil
}

// Attr defines an HTML tag attribute `<Name>="<Value>"“. If Value is empty, only `<Name>` is rendered.
type Attr struct {
	Name      string
	Value     string
	Stringer  fmt.Stringer
	Enabled   bool
	Off       bool
	True      bool
	Condition string
}

func (a Attr) Empty() bool {
	return a.Value == "" && !a.Enabled && !a.Off && !a.True && a.Stringer == nil
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
		return fmt.Sprintf(`%s="off"`, a.Name)
	}
	if a.True {
		return fmt.Sprintf(`%s="true"`, a.Name)
	}
	if a.Stringer != nil {
		return fmt.Sprintf(`%s="%s"`, a.Name, a.Stringer.String())
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

func (t Tag) Init(f *Framework, w io.Writer) error {
	_, err := w.Write([]byte(`<` + t.Name))
	if err != nil {
		return fmt.Errorf(`failed to write tag "%s" start: %w`, t.Name, err)
	}
	for i, attr := range t.Attrs {
		if !attr.Empty() {
			_, err = w.Write([]byte(` ` + attr.String()))
			if err != nil {
				return fmt.Errorf(`failed to write tag "%s" attr %d: %w`, t.Name, i, err)
			}
		}
	}
	_, err = w.Write([]byte(`>`))
	if err != nil {
		return fmt.Errorf(`failed to write tag "%s" start: %w`, t.Name, err)
	}
	if t.Content != nil {
		err = t.Content.Init(f, w)
		if err != nil {
			return fmt.Errorf(`failed to write tag "%s" content: %w`, t.Name, err)
		}
	}
	_, err = w.Write([]byte(`</` + t.Name + `>`))
	if err != nil {
		return fmt.Errorf(`failed to write tag "%s" end: %w`, t.Name, err)
	}
	return nil
}

// --- Shorthand Components ---

// Document the baseline component of an HTML document.
type Document struct {
	// Header defines Component to be rendered in between the <head> tags
	Header Component
	// Body defines the Component to be rendered in between the <body> tags
	Body Component
}

func (d Document) Init(f *Framework, w io.Writer) error {
	return Fragment{
		Raw("<!DOCTYPE html>"),
		Tag{Name: "html", Content: Fragment{
			Tag{Name: "head", Content: d.Header},
			Tag{Name: "body", Content: d.Body},
		}},
	}.Init(f, w)
}

// Div is a shorthand for a "div" Tag
type Div struct {
	ID      string
	Classes []string
	Style   []string
	Attr    []Attr

	Hidden bool

	Content Component
}

func (d Div) Init(f *Framework, w io.Writer) error {
	return Tag{
		Name: "div",
		Attrs: append(d.Attr,
			Attr{Name: "id", Value: d.ID},
			Attr{Name: "class", Value: strings.Join(d.Classes, " ")},
			Attr{Name: "style", Value: strings.Join(d.Style, ";")},
			Attr{Name: "hidden", Enabled: d.Hidden},
		),
		Content: d.Content,
	}.Init(f, w)
}

type Input struct {
	ID      string
	Name    string
	Value   string
	Classes []string
	Style   []string
	Attr    []Attr

	Hidden bool
}

func (i Input) Load(f *Framework, w io.Writer) error {
	return Tag{
		Name: "input",
		Attrs: append(i.Attr,
			Attr{Name: "id", Value: i.ID},
			Attr{Name: "name", Value: i.Name},
			Attr{Name: "type", Value: "text"},
			Attr{Name: "value", Value: i.Value},
			Attr{Name: "class", Value: strings.Join(i.Classes, " ")},
			Attr{Name: "style", Value: strings.Join(i.Style, ";")},
			Attr{Name: "hidden", Enabled: i.Hidden},
		),
	}.Init(f, w)
}
