package gohtmx

import (
	"fmt"
	"io"
)

type ReferenceValidator interface {
	Validate(r *Reference) error
}

// Reference wraps a Component to allow for referencing during validation and rendering.
type Reference struct {
	// Target is the Component that is being referenced.
	Target Component
	// Validation is an optional function to run when this reference is validated. This is monatomic, so will only ever be called once.
	Validation ReferenceValidator

	// Initialized is the initialized Element, if it has been initialized.
	Initialized Element
	// Page is the page that the Reference is rendered initialized at.
	// This is stored so any content generated during validation can be added relative to the same location the content is rendered.
	Page *Page
}

// -- Component --

func (m *Reference) Init(p *Page) (Element, error) {
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
	if m.Validation != nil {
		err := m.Validation.Validate(m)
		if err != nil {
			return err
		}
		// Validation is monatomic, so it is cleared after running.
		m.Validation = nil
	}
	return m.Initialized.Validate()
}

func (m *Reference) FindAttrs() (*Attributes, error) {
	if m == nil || m.Initialized == nil {
		return nil, fmt.Errorf(`cannot find attributes of uninitialized Target`)
	}
	return m.Initialized.FindAttrs()
}

// -- Convenience Methods --

// ID returns the ID of the initialized Target. If the Target is missing an ID, a new one is generated.
func (m *Reference) ID() (string, error) {
	if m == nil || m.Initialized == nil {
		return "", fmt.Errorf(`cannot get ID of uninitialized Reference`)
	}
	a, err := m.Initialized.FindAttrs()
	if err != nil {
		return "", err
	}
	id, ok := a.Get("id")
	if !ok {
		id = m.Page.Generator.NewGroupID("gohtmx")
		// Delete ensure even if multiple IDs were set, only one is in the final output.
		a.Delete("id").String("id", id)
	}
	return id, nil
}
