package element

import (
	"errors"
	"fmt"
	"io"

	"github.com/TheWozard/gohtmx/attributes"
)

// Element defines the low level abstraction of an HTML element.
// Elements write directly to an io.Writer and are used to build a html/template.
type Element interface {
	Render(w io.Writer) error
	Validate() error
	GetTags() []*Tag
}

// Fragment defines a slice of Fragment that can be used as a single Element.
type Fragment []Element

func (f Fragment) Render(w io.Writer) error {
	for _, element := range f {
		if element != nil {
			err := element.Render(w)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (f Fragment) Validate() error {
	errs := make([]error, 0, len(f))
	for _, element := range f {
		if element != nil {
			errs = append(errs, element.Validate())
		}
	}
	return errors.Join(errs...)
}

func (f Fragment) GetTags() []*Tag {
	children := make([]*Tag, 0, len(f))
	for _, element := range f {
		if element != nil {
			children = append(children, element.GetTags()...)
		}
	}
	return children
}

// Raw defines the most simple Element that defines purely string data.
type Raw string

func (r Raw) Render(w io.Writer) error {
	_, err := w.Write([]byte(r))
	return err
}

func (r Raw) Validate() error {
	return nil
}

func (r Raw) GetTags() []*Tag {
	return []*Tag{}
}

type RawError struct {
	Err error
}

func (r RawError) Render(w io.Writer) error {
	_, err := w.Write([]byte(r.Err.Error()))
	return err
}

func (r RawError) Validate() error {
	return r.Err
}

func (r RawError) GetTags() []*Tag {
	return []*Tag{}
}

// Tag defines the lowest level HTML generic tag element.
type Tag struct {
	// Name of this tag.
	Name string
	// Attr defines the list of Attributes for this tag.
	Attributes *attributes.Attributes
	// Content defines the contents this wraps.
	Content Element
}

func (t *Tag) Render(w io.Writer) error {
	_, err := w.Write([]byte(`<` + t.Name))
	if err != nil {
		return ErrPrependPath(fmt.Errorf(`failed to write start tag start: %w`, err), t.Name)
	}
	if !t.Attributes.IsEmpty() {
		_, err = w.Write([]byte(` `))
		if err != nil {
			return ErrPrependPath(fmt.Errorf(`failed to write start tag attribute separator: %w`, err), t.Name)
		}
		err = t.Attributes.Write(w)
		if err != nil {
			return ErrPrependPath(err, t.Name)
		}
	}
	_, err = w.Write([]byte(`>`))
	if err != nil {
		return ErrPrependPath(fmt.Errorf(`failed to write start tag end: %w`, err), t.Name)
	}
	if t.Content != nil {
		err = t.Content.Render(w)
		if err != nil {
			return ErrPrependPath(err, t.Name)
		}
	}
	_, err = w.Write([]byte(`</` + t.Name + `>`))
	if err != nil {
		return ErrPrependPath(fmt.Errorf(`failed to write "%s" end tag: %w`, t.Name, err), t.Name)
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

func (t *Tag) GetTags() []*Tag {
	return []*Tag{t}
}
