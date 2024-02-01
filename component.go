package gohtmx

import "fmt"

// Component defines the high level abstraction of an HTML element.
// Components are decomposed into an Element through the Init call.
type Component interface {
	// Init converts the component into an Element based on the contextual data of the Framework.
	Init(f *Page) (Element, error)
}

// Fragment defines a slice of Components that can be used as a single Component.
type Fragment []Component

func (fr Fragment) Init(f *Page) (Element, error) {
	elements := make(Elements, len(fr))
	for i, fragment := range fr {
		if fragment != nil {
			elements[i] = f.Init(fragment)
		}
	}
	return elements, nil
}

// Document the baseline component of an HTML document.
type Document struct {
	// Header defines Component to be rendered in between the <head> tags.
	Header Component
	// Body defines the Component to be rendered in between the <body> tags.
	Body Component
}

func (d Document) Init(f *Page) (Element, error) {
	return Elements{
		Raw("<!DOCTYPE html>"),
		&Tag{Name: "html", Content: Elements{
			&Tag{Name: "head", Content: f.Init(d.Header)},
			&Tag{Name: "body", Content: f.Init(d.Body)},
		}},
	}, nil
}

// Div is a shorthand for a "div" Tag.
type Div struct {
	ID      string
	Classes []string
	Attrs   *Attributes
	Hidden  bool

	Content Component
}

func (d Div) Init(f *Page) (Element, error) {
	return &Tag{
		Name: "div",
		Attrs: d.Attrs.
			String("id", d.ID).
			Strings("class", d.Classes...).
			Bool("hidden", d.Hidden),
		Content: f.Init(d.Content),
	}, nil
}

// Button is a shorthand for a "button" Tag.
type Button struct {
	ID      string
	Classes []string
	Attr    *Attributes

	Content Component

	Hidden   bool
	Disabled bool
}

func (b Button) Init(f *Page) (Element, error) {
	return &Tag{
		Name: "button",
		Attrs: b.Attr.
			String("id", b.ID).
			Strings("class", b.Classes...).
			String("type", "button").
			Bool("hidden", b.Hidden).
			Bool("disabled", b.Disabled),
		Content: f.Init(b.Content),
	}, nil
}

// Input is a shorthand for an "input" Tag.
type Input struct {
	ID      string
	Classes []string
	Attr    *Attributes
	Hidden  bool

	Type     string
	Name     string
	Value    string
	Disabled bool
}

func (i Input) Init(f *Page) (Element, error) {
	return &Tag{
		Name: "input",
		Attrs: i.Attr.
			String("id", i.ID).
			Strings("class", i.Classes...).
			String("type", i.Type).
			String("name", i.Name).
			String("value", i.Value).
			Bool("hidden", i.Hidden).
			Bool("disabled", i.Disabled),
	}, nil
}

// Header is a shorthand for a "header" Tag.
type Header struct {
	ID      string
	Classes []string
	Attr    *Attributes
	Hidden  bool

	Content Component
}

func (h Header) Init(f *Page) (Element, error) {
	return &Tag{
		Name: "header",
		Attrs: h.Attr.
			String("id", h.ID).
			Strings("class", h.Classes...).
			Bool("hidden", h.Hidden),
		Content: f.Init(h.Content),
	}, nil
}

// Span is a shorthand for a "span" Tag.
type Span struct {
	ID      string
	Classes []string
	Attrs   *Attributes
	Hidden  bool

	Content Component
}

func (s Span) Init(f *Page) (Element, error) {
	return &Tag{
		Name: "span",
		Attrs: s.Attrs.
			String("id", s.ID).
			Strings("class", s.Classes...).
			Bool("hidden", s.Hidden),
		Content: f.Init(s.Content),
	}, nil
}

// P is a shorthand for a "p" Tag.
type P struct {
	ID      string
	Classes []string
	Attrs   *Attributes
	Hidden  bool

	Content Component
}

func (p P) Init(f *Page) (Element, error) {
	return &Tag{
		Name: "p",
		Attrs: p.Attrs.
			String("id", p.ID).
			Strings("class", p.Classes...).
			Bool("hidden", p.Hidden),
		Content: f.Init(p.Content),
	}, nil
}

// A is a shorthand for an "a" Tag.
type A struct {
	ID      string
	Classes []string
	Attrs   *Attributes
	Hidden  bool

	Href    string
	Content Component
}

func (a A) Init(f *Page) (Element, error) {
	return &Tag{
		Name: "a",
		Attrs: a.Attrs.
			String("id", a.ID).
			Strings("class", a.Classes...).
			String("href", a.Href).
			Bool("hidden", a.Hidden),
		Content: f.Init(a.Content),
	}, nil
}

// Img is a shorthand for an "img" Tag.
type Img struct {
	ID      string
	Classes []string
	Attrs   *Attributes
	Hidden  bool

	Src string
	Alt string
}

func (img Img) Init(f *Page) (Element, error) {
	return &Tag{
		Name: "img",
		Attrs: img.Attrs.
			String("id", img.ID).
			Strings("class", img.Classes...).
			String("src", img.Src).
			String("alt", img.Alt).
			Bool("hidden", img.Hidden),
	}, nil
}

// H is a shorthand for a "h*" Tag.
type H struct {
	ID      string
	Classes []string
	Attrs   *Attributes
	Hidden  bool
	Level   int

	Content Component
}

func (h H) Init(f *Page) (Element, error) {
	return &Tag{
		Name: fmt.Sprintf("h%d", h.Level),
		Attrs: h.Attrs.
			String("id", h.ID).
			Strings("class", h.Classes...).
			Bool("hidden", h.Hidden),
		Content: f.Init(h.Content),
	}, nil
}

// UL is a shorthand for a "ul" Tag.
type UL struct {
	ID      string
	Classes []string
	Attrs   *Attributes
	Hidden  bool

	Items []LI
}

func (ul UL) Init(f *Page) (Element, error) {
	items := make(Elements, len(ul.Items))
	for i, item := range ul.Items {
		items[i] = f.Init(item)
	}

	return &Tag{
		Name: "ul",
		Attrs: ul.Attrs.
			String("id", ul.ID).
			Strings("class", ul.Classes...).
			Bool("hidden", ul.Hidden),
		Content: items,
	}, nil
}

// OL is a shorthand for an "ol" Tag.
type OL struct {
	ID      string
	Classes []string
	Attrs   *Attributes
	Hidden  bool

	Items []LI
}

func (ol OL) Init(f *Page) (Element, error) {
	items := make(Elements, len(ol.Items))
	for i, item := range ol.Items {
		items[i] = f.Init(item)
	}

	return &Tag{
		Name: "ol",
		Attrs: ol.Attrs.
			String("id", ol.ID).
			Strings("class", ol.Classes...).
			Bool("hidden", ol.Hidden),
		Content: items,
	}, nil
}

// LI is a shorthand for a "li" Tag.
type LI struct {
	ID      string
	Classes []string
	Attrs   *Attributes
	Hidden  bool

	Content Component
}

func (li LI) Init(f *Page) (Element, error) {
	return &Tag{
		Name: "li",
		Attrs: li.Attrs.
			String("id", li.ID).
			Strings("class", li.Classes...).
			Bool("hidden", li.Hidden),
		Content: f.Init(li.Content),
	}, nil
}

// ComponentTag creates a html tag with the given name.
type ComponentTag struct {
	Name    string
	Attrs   *Attributes
	Content Component
}

func (c ComponentTag) Init(f *Page) (Element, error) {
	return &Tag{
		Name:    c.Name,
		Attrs:   c.Attrs,
		Content: f.Init(c.Content),
	}, nil
}
