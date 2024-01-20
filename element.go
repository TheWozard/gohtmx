package gohtmx

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/TheWozard/gohtmx/internal"
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
	components := make(Fragment, len(attrs))
	for i, attr := range attrs {
		components[i] = Raw(attr)
	}
	return Attributes(components)
}

// Attributes defines a slice of HTML attributes. Functions hung off of this type are chainable, and designed
// to handle empty values. Empty values will skip adding the attribute making it safe to call without checking for empty values.
type Attributes Fragment

func (a Attributes) Init(f *Framework, w io.Writer) error {
	for i, attr := range a {
		if i > 0 {
			_, err := w.Write([]byte(" "))
			if err != nil {
				return internal.ErrEnclosePath(fmt.Errorf("failed to write attribute separator: %w", err), "Attributes")
			}
		}
		err := attr.Init(f, w)
		if err != nil {
			return internal.ErrPrependPath(err, "Attributes")
		}
	}
	return nil
}

func (a Attributes) Copy() Attributes {
	return append(Attributes{}, a...)
}

func (a Attributes) IsEmpty() bool {
	return len(a) == 0
}

// String adds a named value to the attributes if it is not empty.
func (a Attributes) String(name, value string) Attributes {
	if value != "" {
		a = append(a, Raw(fmt.Sprintf(`%s="%s"`, name, value)))
	}
	return a
}

// Slice adds a named value to the attributes if the values are not empty.
func (a Attributes) Slice(name string, values ...string) Attributes {
	return a.String(name, strings.Join(values, " "))
}

// Bool adds a named flag to the attributes if active is true.
func (a Attributes) Bool(name string, active bool) Attributes {
	if active {
		a = append(a, Raw(name))
	}
	return a
}

// If adds the given attributes if the condition is true.
func (a Attributes) If(condition func(*http.Request) bool, positive Attributes) Attributes {
	a = append(a, TCondition{
		Condition: condition,
		Content:   positive,
	})
	return a
}

// IfElse adds the positive attributes if the condition is true, or the negative attributes if the condition returns false.
func (a Attributes) IfElse(condition func(*http.Request) bool, positive, negative Attributes) Attributes {
	a = append(a, TConditions{
		{
			Condition: condition,
			Content:   positive,
		},
		{
			Content: negative,
		},
	})
	return a
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
		return internal.ErrPrependPath(fmt.Errorf(`failed to write start tag start: %w`, err), t.Name)
	}
	if !t.Attrs.IsEmpty() {
		_, err = w.Write([]byte(` `))
		if err != nil {
			return internal.ErrPrependPath(fmt.Errorf(`failed to write start tag attribute separator: %w`, err), t.Name)
		}
		err = t.Attrs.Init(f, w)
		if err != nil {
			return internal.ErrPrependPath(err, t.Name)
		}
	}
	_, err = w.Write([]byte(`>`))
	if err != nil {
		return internal.ErrPrependPath(fmt.Errorf(`failed to write start tag end: %w`, err), t.Name)
	}
	if t.Content != nil {
		err = t.Content.Init(f, w)
		if err != nil {
			return internal.ErrPrependPath(err, t.Name)
		}
	}
	_, err = w.Write([]byte(`</` + t.Name + `>`))
	if err != nil {
		return internal.ErrPrependPath(fmt.Errorf(`failed to write "%s" end tag: %w`, t.Name, err), t.Name)
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
	return internal.ErrPrependPath(Fragment{
		Raw("<!DOCTYPE html>"),
		Tag{Name: "html", Content: Fragment{
			Tag{Name: "head", Content: d.Header},
			Tag{Name: "body", Content: d.Body},
		}},
	}.Init(f, w), "Document")
}

// Div is a shorthand for a "div" Tag.
type Div struct {
	ID      string
	Classes []string
	Attrs   Attributes
	Hidden  bool

	// Update with will cause this div to be rendered out of band when other interactions are done.
	UpdateWith []string

	Content Component
}

func (d Div) Init(f *Framework, w io.Writer) error {
	content := d.Content
	id := d.ID
	// If we are going to update this component, we have to have an id.
	if id == "" && len(d.UpdateWith) > 0 {
		id = f.Generator.NewGroupID("gohtmx-id")
	}
	attrs := d.Attrs.
		String("id", id).
		Slice("class", d.Classes...).
		Bool("hidden", d.Hidden)

	if len(d.UpdateWith) > 0 {
		var err error
		content, err = f.Mono(content)
		if err != nil {
			return internal.ErrPrependPath(err, "div")
		}
		for _, path := range d.UpdateWith {
			f.AtPath(path).AddOutOfBand(Tag{
				Name:    "div",
				Attrs:   attrs.Copy().String("hx-swap-oob", "true").String("hx-target", "#"+id),
				Content: d.Content,
			})
		}
	}

	return Tag{
		Name:    "div",
		Attrs:   attrs,
		Content: content,
	}.Init(f, w)
}

// Button is a shorthand for a "button" Tag.
type Button struct {
	ID      string
	Classes []string
	Attr    Attributes
	Hidden  bool

	Content Component

	Disabled bool
}

func (b Button) Init(f *Framework, w io.Writer) error {
	return Tag{
		Name: "button",
		Attrs: b.Attr.
			String("id", b.ID).
			Slice("class", b.Classes...).
			String("type", "button").
			Bool("hidden", b.Hidden).
			Bool("disabled", b.Disabled),
		Content: b.Content,
	}.Init(f, w)
}

// Input is a shorthand for an "input" Tag.
type Input struct {
	ID      string
	Classes []string
	Attr    Attributes
	Hidden  bool

	Type     string
	Name     string
	Value    string
	Disabled bool
}

func (i Input) Init(f *Framework, w io.Writer) error {
	return Tag{
		Name: "input",
		Attrs: i.Attr.
			String("id", i.ID).
			Slice("class", i.Classes...).
			String("type", i.Type).
			String("name", i.Name).
			String("value", i.Value).
			Bool("hidden", i.Hidden).
			Bool("disabled", i.Disabled),
	}.Init(f, w)
}
