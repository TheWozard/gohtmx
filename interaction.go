package gohtmx

import (
	"fmt"
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
		Validation: func(r *Reference) error {
			id, err := m.target.ID()
			if err != nil {
				return err
			}
			page := m.target.Page.AtPath(m.name)
			fmt.Println(page.Path(), m.name, m.contents)
			page.Add(m.contents)
			for _, trigger := range m.trigger {
				a, err := trigger.FindAttrs()
				if err != nil {
					return err
				}
				a.String("hx-target", "#"+id)
				a.String("hx-post", page.Path())
				m.swap.Attrs(a)
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
