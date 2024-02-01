package gohtmx

import (
	"fmt"
	"io"
)

type Swap string

func (s Swap) Attrs(a *Attributes) {
	if s == "" {
		return
	}
	a.String("hx-swap", string(s))
}

const (
	SwapInnerHTML   Swap = "innerHTML"
	SwapOuterHTML   Swap = "outerHTML"
	SwapAfterBegin  Swap = "afterbegin"
	SwapBeforeBegin Swap = "beforebegin"
	SwapAfterEnd    Swap = "afterend"
	SwapBeforeEnd   Swap = "beforeend"
	SwapDelete      Swap = "delete"
	SwapNone        Swap = "none"
)

func NewInteraction(name string) Interaction {
	return Interaction{name: name}
}

type Interaction struct {
	name string

	target   *Reference
	contents *Reference
	trigger  []*Reference

	swap Swap
}

func (m *Interaction) Swap(s Swap) *Interaction {
	if m == nil {
		return nil
	}
	m.swap = s
	return m
}

// -- References --

func (m *Interaction) Trigger(c Component) Component {
	if m == nil || c == nil {
		return c
	}
	trigger := &Reference{Target: c}
	m.trigger = append(m.trigger, trigger)
	return trigger
}

func (m *Interaction) Target(c Component) Component {
	if m == nil || c == nil {
		return c
	}
	if m.target != nil {
		return nil
	}
	m.target = &Reference{
		Target: c,
		Validation: func(e Element) error {
			id, err := m.target.ID()
			if err != nil {
				return err
			}
			page := m.target.page.AtPath(m.name)
			fmt.Println(m.name, m.contents)
			page.Add(m.contents)
			for _, trigger := range m.trigger {
				a, err := trigger.FindAttrs()
				a.String("hx-target", "#"+id)
				a.String("hx-post", page.Path())
				m.swap.Attrs(a)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	return m.target
}

func (m *Interaction) Content(c Component) Component {
	if m == nil || c == nil {
		return c
	}
	if m.contents != nil {
		return nil
	}
	m.contents = &Reference{Target: c}
	return m.contents
}

// -- Shorthand Functions --

func (m *Interaction) Update(c Component) Component {
	m.Swap(SwapOuterHTML)
	return m.Target(m.Content(c))
}

// -- Wrapping Structures --

// Reference wraps a Component to allow for referencing during validation and rendering.
type Reference struct {
	Target      Component
	Validation  func(e Element) error
	Initialized Element

	page *Page
}

// -- Component --

func (m *Reference) Init(p *Page) (Element, error) {
	if m == nil || m.Target == nil {
		return nil, nil
	}
	if m.Initialized == nil {
		m.Initialized = p.Init(m.Target)
		m.page = p
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
		err := m.Validation(m.Initialized)
		if err != nil {
			return err
		}
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
		id = m.page.Generator.NewGroupID("gohtmx")
		// Delete ensure even if multiple IDs were set, only one is in the final output.
		a.Delete("id").String("id", id)
	}
	return id, nil
}
