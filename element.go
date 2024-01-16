package gohtmx

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
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

func Attrs(attrs ...string) Attributes {
	return Attributes(attrs)
}

// Attributes defines a slice of HTML attributes.
type Attributes []string

func (a Attributes) Copy() Attributes {
	return append(Attributes{}, a...)
}

func (a Attributes) IsEmpty() bool {
	return len(a) == 0
}

// Value adds a named value to the attributes if it is not empty.
func (a Attributes) Value(name, value string) Attributes {
	if value != "" {
		a = append(a, fmt.Sprintf(`%s="%s"`, name, value))
	}
	return a
}

// Values adds a named value to the attributes if the values are not empty.
func (a Attributes) Values(name string, values ...string) Attributes {
	return a.Value(name, strings.Join(values, " "))
}

// Flag adds a named flag to the attributes if active is true.
func (a Attributes) Flag(name string, active bool) Attributes {
	if active {
		a = append(a, name)
	}
	return a
}

// Condition adds the given attributes if the condition is true.
func (a Attributes) Condition(f *Framework, condition func(*http.Request) bool, attrs Attributes) (Attributes, error) {
	if len(attrs) > 0 {
		temp := bytes.NewBuffer(nil)
		err := TCondition{
			Condition: condition,
			Content:   Raw(attrs.String()),
		}.Init(f.NoMux(), temp)
		if err != nil {
			return nil, fmt.Errorf("failed to write attributes condition: %w", err)
		}
		a = append(a, temp.String())
	}
	return a, nil
}

// String returns the string representation of this Attributes.
func (a Attributes) String() string {
	return strings.Join(a, ` `)
}

// Tag defines the lowest level HTML generic tag element.
type Tag struct {
	// Name of this tag.
	Name string
	// Attr defines the list of Attributes for this tag.
	Attrs Attributes
	// Content defines the contents this wraps.
	Content Component
}

func (t Tag) Init(f *Framework, w io.Writer) error {
	_, err := w.Write([]byte(`<` + t.Name))
	if err != nil {
		return AddPathToError(fmt.Errorf(`failed to write start tag start: %w`, err), t.Name)
	}
	if !t.Attrs.IsEmpty() {
		_, err = w.Write([]byte(` ` + t.Attrs.String()))
		if err != nil {
			return AddPathToError(fmt.Errorf(`failed to write start tag attributes: %w`, err), t.Name)
		}
	}
	_, err = w.Write([]byte(`>`))
	if err != nil {
		return AddPathToError(fmt.Errorf(`failed to write start tag end: %w`, err), t.Name)
	}
	if t.Content != nil {
		err = t.Content.Init(f, w)
		if err != nil {
			return AddPathToError(fmt.Errorf(`in %s: %w`, t.Name, err), t.Name)
		}
	}
	_, err = w.Write([]byte(`</` + t.Name + `>`))
	if err != nil {
		return AddPathToError(fmt.Errorf(`failed to write "%s" end tag: %w`, t.Name, err), t.Name)
	}
	return nil
}

// --- Shorthand Components ---

// Document the baseline component of an HTML document.
type Document struct {
	// Header defines Component to be rendered in between the <head> tags.
	Header Component
	// Body defines the Component to be rendered in between the <body> tags.
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

// Div is a shorthand for a "div" Tag.
type Div struct {
	ID      string
	Classes []string
	Attrs   Attributes
	Hidden  bool

	Content Component
}

func (d Div) Init(f *Framework, w io.Writer) error {
	return Tag{
		Name: "div",
		Attrs: d.Attrs.
			Value("id", d.ID).
			Values("class", d.Classes...).
			Flag("hidden", d.Hidden),
		Content: d.Content,
	}.Init(f, w)
}

// Button is a shorthand for a "button" Tag.
type Button struct {
	ID      string
	Classes []string
	Attr    Attributes
	Hidden  bool

	// TODO: HTMX attributes as first class citizens.

	Content Component
}

func (b Button) Init(f *Framework, w io.Writer) error {
	return Tag{
		Name: "button",
		Attrs: b.Attr.
			Value("id", b.ID).
			Values("class", b.Classes...).
			Flag("hidden", b.Hidden),
		Content: b.Content,
	}.Init(f, w)
}

// Input is a shorthand for an "input" Tag.
type Input struct {
	ID      string
	Classes []string
	Attr    Attributes
	Hidden  bool

	Type  string
	Name  string
	Value string
}

func (i Input) Init(f *Framework, w io.Writer) error {
	return Tag{
		Name: "input",
		Attrs: i.Attr.
			Value("id", i.ID).
			Values("class", i.Classes...).
			Flag("hidden", i.Hidden).
			Value("type", i.Type).
			Value("name", i.Name).
			Value("value", i.Value),
	}.Init(f, w)
}
