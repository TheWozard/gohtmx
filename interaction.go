package gohtmx

import (
	"fmt"
	"net/http"
	"time"
)

type Swap string

func (s Swap) Default() Swap {
	if s == "" {
		return SwapOuterHTML
	}
	return s
}

func (s Swap) Show(target string) Swap {
	return Swap(string(s) + " show:" + target)
}

func (s Swap) Scroll(target string) Swap {
	return Swap(string(s) + " scroll:" + target)
}

func (s Swap) FocusScroll(enabled bool) Swap {
	if enabled {
		return Swap(string(s) + " focus-scroll:true")
	}
	return Swap(string(s) + " focus-scroll:false")
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

type Trigger string

func (t Trigger) Changed(delay time.Duration) Trigger {
	return Trigger(string(t) + " changed")
}

func (t Trigger) Delay(delay time.Duration) Trigger {
	return Trigger(string(t) + " delay:" + delay.String())
}

func (t Trigger) Throttle(delay time.Duration) Trigger {
	return Trigger(string(t) + " throttle:" + delay.String())
}

const (
	TriggerLoad     Trigger = "load"
	TriggerRevealed Trigger = "revealed"
	TriggerClick    Trigger = "click"
	TriggerChange   Trigger = "change"
	TriggerSubmit   Trigger = "submit"
)

// -- Interaction --

func NewInteraction(name string) *Interaction {
	return &Interaction{Name: name}
}

type Interaction struct {
	Name string

	Handlers []func(*http.Request)
	Actions  []*Action
	Triggers []*Reference
}

// -- References --

func (m *Interaction) Trigger(c Component) Component {
	if m == nil || c == nil {
		return c
	}
	trigger := &Reference{Target: c}
	m.Triggers = append(m.Triggers, trigger)
	return trigger
}

func (m *Interaction) Action() *Action {
	if m == nil {
		return nil
	}
	action := NewAction()
	m.AddAction(action)
	return action
}

func (m *Interaction) AddAction(a *Action) *Interaction {
	if m == nil || a == nil {
		return m
	}
	if len(m.Actions) > 0 {
		a.OOB()
	}
	m.Actions = append([]*Action{a.AddValidation(m)}, m.Actions...)
	return m
}

func (m *Interaction) Handle(f func(*http.Request)) *Interaction {
	if m == nil {
		return nil
	}
	m.Handlers = append(m.Handlers, f)
	return m
}

func (m *Interaction) Validate(r *Reference) error {
	if m == nil || len(m.Actions) == 0 {
		return nil
	}
	base := m.Actions[len(m.Actions)-1]
	if base.oob {
		// If the base action is oob, there is no non oob actions. So we cant directly swap anything.
		for _, trigger := range m.Triggers {
			a, err := trigger.FindAttrs()
			if err != nil {
				return err
			}
			a.String("hx-post", r.Page.Path())
			a.String("hx-swap", string(SwapNone))
		}
		return nil
	}
	// We have a valid base action.
	id, err := base.target.ID()
	if err != nil {
		return err
	}
	page := base.target.Page.AtPath(m.Name)
	contents := make(Fragment, len(m.Actions))
	for i, action := range m.Actions {
		contents[i] = action
	}
	page.Add(contents)
	page.Use(m.Middleware)
	for _, trigger := range m.Triggers {
		a, err := trigger.FindAttrs()
		if err != nil {
			return err
		}
		a.String("hx-target", "#"+id)
		a.String("hx-post", page.Path())
		a.String("hx-swap", string(base.swap))
	}
	return nil
}

func (m *Interaction) Middleware(next http.Handler) http.Handler {
	if len(m.Handlers) == 0 {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, h := range m.Handlers {
			h(r)
		}
		next.ServeHTTP(w, r)
	})
}

// -- Action --

// Creates a new Action. It is generally better to use the Interaction.Action method.
func NewAction() *Action {
	return &Action{}
}

// Action defines the application of new content to a target. This can occur either in or out of bounds.
// If out of bounds, the contents will be updated to include the needed Attributes.
type Action struct {
	target      *Reference
	contents    *Reference
	validations []ReferenceValidator
	swap        Swap
	oob         bool
}

// Swap sets the swap method for the Action.
func (a *Action) Swap(s Swap) *Action {
	if a == nil {
		return nil
	}
	a.swap = s
	return a
}

// OOB sets the Action to be out of bounds.
func (a *Action) OOB() *Action {
	if a == nil {
		return nil
	}
	a.oob = true
	return a
}

// Target sets the target of the Action. This is used to identify using the ID of the target.
// If the target is missing an ID, a new one is generated for the target.
// Target can only be set once.
func (a *Action) Target(c Component) Component {
	if a == nil || c == nil {
		return c
	}
	if a.target != nil {
		return RawError{Err: fmt.Errorf("target already set")}
	}
	a.target = &Reference{
		Target: c,
		// All action is based on the target being a part of the tree.
		Validation: a,
	}
	return a.target
}

// Content sets the content of the Action. This is the content that will be swapped with the target.
// Content can only be set once.
func (a *Action) Content(c Component) Component {
	if a == nil || c == nil {
		return c
	}
	if a.contents != nil {
		return RawError{Err: fmt.Errorf("content already set")}
	}
	a.contents = &Reference{
		Target: c,
	}
	return a.contents
}

// Shorthand for setting the content and target of the Action to the same Component. Set the swap to OuterHTML.
func (a *Action) Update(c Component) Component {
	a.Swap(SwapOuterHTML)
	return a.Content(a.Target(c))
}

// AddValidation adds a validation to the Action. This is run when the Action is validated.
func (a *Action) AddValidation(v ReferenceValidator) *Action {
	if a == nil || v == nil {
		return nil
	}
	a.validations = append(a.validations, v)
	return a
}

// Validate validates the Action. This includes setting the needed Attributes for out of bounds Actions.
func (a *Action) Validate(r *Reference) error {
	if a == nil {
		return nil
	}
	if a.contents == nil {
		return fmt.Errorf("content not set")
	}
	if a.target == nil {
		return fmt.Errorf("target not set")
	}

	// Run all linked validations.
	for _, v := range a.validations {
		err := v.Validate(r)
		if err != nil {
			return err
		}
	}

	// If this action is being validated then the target is expected to have been been initialized.
	// We will initialize the content relative to the target if it isn't mounted anywhere else.
	// This is a bit awkward to initialize during the validation stage, but is nessisary to ensure
	// the content is not mounted anywhere else in the tree.
	if a.contents.Initialized == nil {
		if a.target.Page == nil {
			return fmt.Errorf("target was never initialized")
		}
		_, err := a.contents.Init(a.target.Page)
		if err != nil {
			return err
		}
	}

	if !a.oob {
		return nil
	}

	// Actually set the needed Attributes for out of bounds Actions.
	ca, err := a.contents.FindAttrs()
	if err != nil {
		return err
	}
	a.swap = a.swap.Default()
	ca.String("hx-swap-oob", string(a.swap))
	if a.swap == SwapOuterHTML {
		tid, err := a.target.ID()
		if err != nil {
			return err
		}
		ca.String("id", tid)
	}
	return nil
}

// Should an Action be able to be used as a Component? Or should an Interaction directly reference the Contents.
func (a *Action) Init(p *Page) (Element, error) {
	if a == nil {
		return nil, nil
	}
	if a.contents == nil {
		return nil, fmt.Errorf("content not set")
	}
	if a.target == nil {
		return nil, fmt.Errorf("target not set")
	}
	return a.contents.Init(p)
}
