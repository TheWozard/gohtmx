package gohtmx

import (
	"fmt"
	"io"

	"github.com/TheWozard/gohtmx/attributes"
	"github.com/TheWozard/gohtmx/element"
)

type ValidationFunc func(r *Reference) error
type RenderFunc func(r *Reference, w io.Writer) (io.Writer, error)

// Reference wraps a Component to allow for referencing during validation and rendering.
type Reference struct {
	// Target is the Component that is being referenced.
	Target Component
	// Initialized is the initialized Element, if it has been initialized.
	Initialized element.Element
	// Page is the page that the Reference is rendered initialized at.
	// This is stored so any content generated during validation can be added relative to the same location the content
	// is rendered.
	Page *Page
}

// -- Component --

func (m *Reference) Init(p *Page) (element.Element, error) {
	if m == nil || m.Target == nil {
		return nil, nil
	}
	if m.Initialized == nil {
		m.Initialized = p.Init(m.Target)
		m.Page = p
	}
	return m, nil
}

// -- Element --

func (m *Reference) Render(w io.Writer) error {
	if m == nil || m.Initialized == nil {
		return fmt.Errorf(`cannot render uninitialized Target`)
	}
	return m.Initialized.Render(w)
}

func (m *Reference) Validate() error {
	if m == nil || m.Initialized == nil {
		return fmt.Errorf(`cannot validate uninitialized Target`)
	}
	return m.Initialized.Validate()
}

func (m *Reference) GetTags() []*element.Tag {
	if m == nil || m.Initialized == nil {
		return []*element.Tag{}
	}
	return m.Initialized.GetTags()
}

// -- Convenience Methods --

// ID returns the ID of the initialized Target. If the Target is missing an ID, a new one is generated.
func (m *Reference) ID() (string, error) {
	if m == nil || m.Initialized == nil {
		return "", fmt.Errorf(`cannot get ID of uninitialized Reference`)
	}
	a, err := m.FindAttrs()
	if err != nil {
		return "", err
	}
	id, ok := a.Get("id")
	if !ok {
		id = m.Page.Generator.NewID("gohtmx")
		// Delete ensure even if multiple IDs were set, only one is in the final output.
		a.Delete("id").String("id", id)
	}
	return id, nil
}

func (m *Reference) FindAttrs() (*attributes.Attributes, error) {
	if m == nil || m.Initialized == nil {
		return nil, fmt.Errorf(`cannot find attributes of uninitialized Reference`)
	}
	tags := m.Initialized.GetTags()
	if len(tags) == 0 {
		return nil, fmt.Errorf(`failed to get attributes, missing tags in initialized Reference`)
	}
	if len(tags) > 1 {
		return nil, fmt.Errorf(`failed to get attributes, multiple tags in initialized Reference`)
	}
	return tags[0].Attrs, nil
}
