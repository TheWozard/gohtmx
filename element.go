package gohtmx

import (
	"errors"
	"fmt"
	"io"

	"github.com/TheWozard/gohtmx/internal"
)

// Element defines the low level abstraction of an HTML element.
// Elements write directly to an io.Writer and are used to build a html/template.
type Element interface {
	Render(w io.Writer) error
	Validate() error
	FindAttrs() (*Attributes, error)
}

// Elements defines a slice of Elements that can be used as a single Element.
type Elements []Element

func (e Elements) Render(w io.Writer) error {
	for _, element := range e {
		if element != nil {
			err := element.Render(w)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (e Elements) Validate() error {
	var err error
	for _, element := range e {
		if element != nil {
			err = errors.Join(err, element.Validate())
		}
	}
	return err
}

func (e Elements) FindAttrs() (*Attributes, error) {
	return nil, fmt.Errorf(`cannot find attributes in Elements, use Tag instead`)
}

// Raw defines the most simple Element that defines purely string data
type Raw string

func (r Raw) Init(f *Page) (Element, error) {
	return r, nil
}

func (r Raw) Render(w io.Writer) error {
	_, err := w.Write([]byte(r))
	return err
}

func (r Raw) Validate() error {
	return nil
}

func (r Raw) FindAttrs() (*Attributes, error) {
	return nil, fmt.Errorf(`cannot find attributes in Raw, use Tag instead`)
}

type RawError struct {
	Err error
}

func (r RawError) Init(f *Page) (Element, error) {
	return r, nil
}

func (r RawError) Render(w io.Writer) error {
	_, err := w.Write([]byte(r.Err.Error()))
	return err
}

func (r RawError) Validate() error {
	return r.Err
}

func (r RawError) FindAttrs() (*Attributes, error) {
	return nil, fmt.Errorf(`cannot find attributes in RawError, use Tag instead`)
}

// Tag defines the lowest level HTML generic tag element.
type Tag struct {
	// Name of this tag.
	Name string
	// Attr defines the list of Attributes for this tag.
	Attrs *Attributes
	// Content defines the contents this wraps.
	Content Element
}

func (t *Tag) Render(w io.Writer) error {
	_, err := w.Write([]byte(`<` + t.Name))
	if err != nil {
		return internal.ErrPrependPath(fmt.Errorf(`failed to write start tag start: %w`, err), t.Name)
	}
	if !t.Attrs.IsEmpty() {
		_, err = w.Write([]byte(` `))
		if err != nil {
			return internal.ErrPrependPath(fmt.Errorf(`failed to write start tag attribute separator: %w`, err), t.Name)
		}
		err = t.Attrs.Render(w)
		if err != nil {
			return internal.ErrPrependPath(err, t.Name)
		}
	}
	_, err = w.Write([]byte(`>`))
	if err != nil {
		return internal.ErrPrependPath(fmt.Errorf(`failed to write start tag end: %w`, err), t.Name)
	}
	if t.Content != nil {
		err = t.Content.Render(w)
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

func (t *Tag) Validate() error {
	if t.Name == "" {
		return fmt.Errorf(`missing tag name`)
	}
	if t.Content != nil {
		return t.Content.Validate()
	}
	return nil
}

func (t *Tag) FindAttrs() (*Attributes, error) {
	t.Attrs = t.Attrs.Ensure()
	return t.Attrs, nil
}
