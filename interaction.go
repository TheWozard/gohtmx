package gohtmx

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// SwapMethod defines the method of swapping content. See https://htmx.org/docs/#swapping
type SwapMethod string

// Show sets the SwapMethod to show the target at the given ScrollPosition when the content is swapped.
func (s SwapMethod) Show(target ScrollPosition) SwapMethod {
	return SwapMethod(string(s) + " show:" + string(target))
}

// Scroll sets the SwapMethod to scroll the target to the given ScrollPosition when the content is swapped.
func (s SwapMethod) Scroll(target ScrollPosition) SwapMethod {
	return SwapMethod(string(s) + " scroll:" + string(target))
}

// FocusScroll sets the SwapMethod to scroll to the focused element when the content is swapped.
func (s SwapMethod) FocusScroll(enabled bool) SwapMethod {
	if enabled {
		return SwapMethod(string(s) + " focus-scroll:true")
	}
	return SwapMethod(string(s) + " focus-scroll:false")
}

const (
	SwapInnerHTML   SwapMethod = "innerHTML"
	SwapOuterHTML   SwapMethod = "outerHTML"
	SwapAfterBegin  SwapMethod = "afterbegin"
	SwapBeforeBegin SwapMethod = "beforebegin"
	SwapAfterEnd    SwapMethod = "afterend"
	SwapBeforeEnd   SwapMethod = "beforeend"
	SwapDelete      SwapMethod = "delete"
	SwapNone        SwapMethod = "none"
)

// ScrollPosition is the location to Scroll or Show in the SwapMethod.
type ScrollPosition string

const (
	ScrollTop    ScrollPosition = "top"
	ScrollBottom ScrollPosition = "bottom"
)

// TriggerMethod defines the method of triggering an Interaction. See https://htmx.org/docs/#triggers
type TriggerMethod string

// Changed sets the TriggerMethod to trigger when the value of the target changes.
func (t TriggerMethod) Changed(delay time.Duration) TriggerMethod {
	return TriggerMethod(string(t) + " changed")
}

// Delay sets the TriggerMethod to trigger after the given delay. If a new event is triggered before the delay, the timer is reset.
func (t TriggerMethod) Delay(delay time.Duration) TriggerMethod {
	return TriggerMethod(string(t) + " delay:" + delay.String())
}

// Throttle sets the TriggerMethod to trigger at most once every delay. If a new event is triggered before the delay, the event is ignored.
func (t TriggerMethod) Throttle(delay time.Duration) TriggerMethod {
	return TriggerMethod(string(t) + " throttle:" + delay.String())
}

const (
	TriggerLoad     TriggerMethod = "load"
	TriggerRevealed TriggerMethod = "revealed"
	TriggerClick    TriggerMethod = "click"
	TriggerChange   TriggerMethod = "change"
	TriggerSubmit   TriggerMethod = "submit"
)

// -- Interaction --

// NewInteraction creates a new Interaction with the given name.
func NewInteraction(name string) *Interaction {
	return &Interaction{Name: name}
}

// Interaction defines a set of Swaps, Triggers, and Updates. Swaps define changes to the content of the page.
// Triggers define what can cause the Interaction to occur. Updates define changes to the backing data.
// All interactions are named to mount the interaction to the current page path.
type Interaction struct {
	Name string

	// TODO: Handler? is there value to multiple
	handlers []func(*http.Request)
	swaps    []*Swap
	triggers []*Trigger
}

// Trigger creates a new Trigger for the Interaction.
func (i *Interaction) Trigger() *Trigger {
	if i == nil {
		return nil
	}
	trigger := NewTrigger()
	i.AddTrigger(trigger)
	return trigger
}

// AddTrigger adds a Trigger to the Interaction.
func (i *Interaction) AddTrigger(t *Trigger) *Interaction {
	if i == nil || t == nil {
		return i
	}
	i.triggers = append(i.triggers, t)
	return i
}

// Swap creates a new Swap for the Interaction.
func (i *Interaction) Swap() *Swap {
	if i == nil {
		return nil
	}
	action := NewSwap()
	i.AddSwap(action)
	return action
}

// AddSwap adds a Swap to the Interaction.
func (i *Interaction) AddSwap(a *Swap) *Interaction {
	if i == nil || a == nil {
		return i
	}
	if len(i.swaps) > 0 {
		a.OutOfBounds()
	}
	i.swaps = append([]*Swap{a.addValidation(i.validate)}, i.swaps...)
	return i
}

// Handle adds a handler to the Interaction.
func (i *Interaction) Handle(f func(*http.Request)) *Interaction {
	if i == nil {
		return nil
	}
	i.handlers = append(i.handlers, f)
	return i
}

func (i *Interaction) validate(r *Reference) error {
	if i == nil {
		return nil
	}
	var swap *Swap
	page := r.Page
	if len(i.swaps) > 0 {
		last := i.swaps[len(i.swaps)-1]
		page = last.target.Page
		swap = last
	}
	page = page.AtPath(i.Name)
	contents := make(Fragment, len(i.swaps))
	for j, action := range i.swaps {
		contents[j] = action.target
	}
	page.Add(contents)
	page.Use(i.middleware)
	for _, trigger := range i.triggers {
		err := trigger.Update(page, swap)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Interaction) middleware(next http.Handler) http.Handler {
	if len(i.handlers) == 0 {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, h := range i.handlers {
			h(r)
		}
		next.ServeHTTP(w, r)
	})
}

// -- Swap --

// Creates a new Swap. This defines the application of new content to a target. This can occur either in or out of bounds.
func NewSwap() *Swap {
	return &Swap{}
}

// Swap defines the application of new content to a target. This can occur either in or out of bounds.
// If out of bounds, the contents will be updated to include the id of the contents.
type Swap struct {
	target      *Reference
	contents    *Reference
	validations []ReferenceValidator
	method      SwapMethod
	outOfBand   bool
}

// Method sets the swap method to replace the target with.
func (s *Swap) Method(m SwapMethod) *Swap {
	if s == nil {
		return nil
	}
	s.method = m
	return s
}

// OutOfBounds sets the Swap to be out of bounds.
func (s *Swap) OutOfBounds() *Swap {
	if s == nil {
		return nil
	}
	s.outOfBand = true
	if s.method == "" {
		s.method = SwapOuterHTML
	}
	return s
}

// Target sets the target of the Swap. The ID is used for targeting the element to swap.
// If the target is missing an ID, a new one is generated for the target. Target can only be set once.
func (s *Swap) Target(c Component) Component {
	if s == nil || c == nil {
		return nil
	}
	if s.target != nil {
		return RawError{Err: fmt.Errorf("target already set")}
	}
	s.target = &Reference{
		Target: c,
		// All validation is based on the target being a part of the tree.
		// If the target is not a part of the tree, then the swap will never be validated.
		Validation: s.validate,
	}
	return s.target
}

// Content sets the content of the Swap. Content can only be set once.
func (a *Swap) Content(c Component) Component {
	if a == nil || c == nil {
		return nil
	}
	if a.contents != nil {
		return RawError{Err: fmt.Errorf("content already set")}
	}
	a.contents = &Reference{
		Target: c,
	}
	return a.contents
}

// Shorthand for setting the content and target of the Swap to the same Component. Also sets the swap to OuterHTML.
func (s *Swap) Update(c Component) Component {
	s.Method(SwapOuterHTML)
	return s.Content(s.Target(c))
}

func (s *Swap) addValidation(v ReferenceValidator) *Swap {
	if s == nil || v == nil {
		return nil
	}
	s.validations = append(s.validations, v)
	return s
}

func (s *Swap) validate(r *Reference) error {
	if s == nil {
		return nil
	}
	if s.contents == nil {
		return fmt.Errorf("content not set")
	}
	if s.target == nil {
		return fmt.Errorf("target not set")
	}

	// Run all linked validations.
	for _, v := range s.validations {
		err := v(r)
		if err != nil {
			return err
		}
	}

	// If this swap is being validated then the target is expected to have been been initialized.
	// We will initialize the content relative to the target if it isn't mounted anywhere else.
	// This is a bit awkward to initialize during the validation stage, but is nessisary to ensure
	// the content is not mounted anywhere else in the tree.
	if s.contents.Initialized == nil {
		if s.target.Page == nil {
			return fmt.Errorf("target was never initialized")
		}
		_, err := s.contents.Init(s.target.Page)
		if err != nil {
			return err
		}
	}

	if !s.outOfBand {
		return nil
	}

	// Actually set the needed Attributes for out of bounds Swap.
	ca, err := s.contents.FindAttrs()
	if err != nil {
		return err
	}
	ca.String("hx-swap-oob", string(s.method))
	if s.method == SwapOuterHTML {
		tid, err := s.target.ID()
		if err != nil {
			return err
		}
		ca.String("id", tid)
	}
	return nil
}

func (s *Swap) triggerAttrs(a *Attributes) error {
	if s == nil {
		a.String("hx-swap", string(SwapNone))
	} else {
		id, err := s.target.ID()
		if err != nil {
			return err
		}
		a.String("hx-swap", string(s.method))
		a.String("hx-target", "#"+id)
	}
	return nil
}

// -- Trigger --

// NewTrigger creates a new Trigger.
func NewTrigger() *Trigger {
	return &Trigger{Values: url.Values{}}
}

// Trigger defines something that can cause an Interaction to occur. Each Trigger can contain a set of values to be sent with the request.
type Trigger struct {
	target *Reference
	method TriggerMethod
	Values url.Values
}

func (t *Trigger) Target(c Component) Component {
	if t == nil || c == nil {
		return nil
	}
	if t.target != nil {
		return RawError{Err: fmt.Errorf("target already set")}
	}
	t.target = &Reference{
		Target: c,
	}
	return t.target
}

func (t *Trigger) Method(m TriggerMethod) *Trigger {
	if t == nil {
		return nil
	}
	t.method = m
	return t
}

func (t *Trigger) Set(key, value string) *Trigger {
	if t == nil {
		return nil
	}
	if t.Values == nil {
		t.Values = url.Values{}
	}
	t.Values.Set(key, value)
	return t
}

func (t *Trigger) Update(p *Page, swap *Swap) error {
	a, err := t.target.FindAttrs()
	if err != nil {
		return err
	}
	a.String("hx-post", t.Path(p))
	a.String("hx-trigger", string(t.method))
	return swap.triggerAttrs(a)
}

func (t *Trigger) Path(p *Page) string {
	if t.Values != nil && len(t.Values) > 0 {
		return p.Path() + "?" + t.Values.Encode()
	}
	return p.Path()
}

func (t *Trigger) Init(p *Page) (Element, error) {
	if t == nil {
		return nil, nil
	}
	if t.target == nil {
		return nil, fmt.Errorf("target not set")
	}
	return t.target.Init(p)
}
