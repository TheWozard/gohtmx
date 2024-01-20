package gohtmx

import (
	"io"
	"net/http"

	"github.com/TheWozard/gohtmx/core"
	"github.com/TheWozard/gohtmx/internal"
)

// Form defines a HTML form and the submit action.
type Form struct {
	ID      string
	Classes []string
	Attrs   Attributes

	Content Component

	// Submit defines what happens when the form is submitted. Returned Data will be passed to the success Component.
	// If an error occurs, the error Component will be rendered instead.
	// Submit is optional, as if omitted Success will be rendered with the request Data.
	Submit func(d Data) (Data, error)
	// UpdateParams defines Submit result values get set in the response.
	UpdateParams []string
	// UpdateForm adds an out of band update to the form on success or error.
	// This will update the form regardless of what element calls the form interaction endpoint.
	// This does require the form to load values from request data.
	UpdateForm bool
	// AutoSubmit defines when a form should be triggered on load automatically.
	// For simple Submit cases a TCondition in the Target Component is recommended as this will
	// load in the initial document and not require a second request.
	AutoSubmit func(r *http.Request) bool
	// Target defines CSS target to write either the Success or Error Component to.
	Target string
	// Error defines the Component to render on Submit error.
	Error Component
	// Success defines the Component to render on Submit success.
	Success Component
}

func (fr Form) Init(f *Framework, w io.Writer) error {
	// Validation/Setup
	f = f.AtPath(fr.ID)
	v := internal.NewValidate()
	v.Require(fr.ID != "", "ID required")
	v.Require(fr.Target != "", "Target required")
	v.Require(fr.Success != nil, "Success required")
	v.Require(fr.Error != nil, "Error required")
	if v.HasError() {
		return internal.ErrEnclosePath(v.Error(), "Form")
	}

	fr.Attrs = fr.Attrs.
		String("id", fr.ID).
		Slice("class", fr.Classes...).
		String("hx-post", f.Path()).
		String("hx-target", fr.Target).
		String("autocomplete", "off")
	if fr.UpdateForm {
		fr.Content = &MetaMono{Content: fr.Content}
		f.AddOutOfBand(Tag{
			Name:    "form",
			Attrs:   fr.Attrs.Copy().String("hx-swap-oob", "true").String("hx-target", "#"+fr.ID),
			Content: fr.Content,
		})
	}

	err := f.AddInteraction(TMultiComponent{
		Select: func(r *http.Request) (int, Data) {
			data := GetAllDataFromRequest(r)
			if fr.Submit == nil {
				return 0, data
			}
			data, err := fr.Submit(data)
			if err != nil {
				return 1, Data{"error": err}
			}
			return 0, data
		},
		Options: []Component{fr.Success, fr.Error},
	})
	if err != nil {
		return internal.ErrEnclosePath(err, "Form")
	}

	if fr.AutoSubmit != nil {
		fr.Attrs = fr.Attrs.If(fr.AutoSubmit, Attrs().String("hx-trigger", "load,submit"))
	}

	if len(fr.UpdateParams) > 0 {
		f.Use(UpdateParams(fr.UpdateParams...))
	}

	// Rendering
	return internal.ErrEnclosePath(Tag{Name: "form", Attrs: fr.Attrs, Content: fr.Content}.Init(f, w), "Form")
}

// InputText defines an input text field with additional features for label, and onChange.
type InputText struct {
	ID      string
	Classes []string
	Attrs   Attributes

	OnChange   func(r *http.Request) Data
	Additional Component

	Name        string
	Label       string
	Placeholder string
	Value       string
	AutoFocus   bool
	Disabled    bool
}

func (it InputText) Init(f *Framework, w io.Writer) error {
	content := Fragment{}
	f = f.AtPath(core.FirstNonEmptyString(it.ID, it.Name))
	if it.Label != "" {
		content = append(content, Tag{
			Name:    "label",
			Attrs:   Attrs().String("for", it.ID),
			Content: Raw(it.Label),
		})
	}
	inputAttr := Attrs()
	groupAttr := it.Attrs
	if it.OnChange != nil {
		inputAttr = inputAttr.
			String("hx-trigger", "keyup changed delay:400ms").
			String("hx-disabled-elt", "this").
			String("hx-post", f.Path()).
			Bool("autofocus", it.AutoFocus)
		groupAttr = groupAttr.
			String("hx-target", "this").
			String("hx-swap", "morph:outerHTML")
	}
	content = append(content, Tag{
		Name: "input",
		Attrs: inputAttr.
			String("type", "text").
			String("name", core.FirstNonEmptyString(it.Name, it.ID)).
			String("placeholder", core.FirstNonEmptyString(it.Placeholder, it.ID)).
			String("value", it.Value),
	})
	var result Component
	result = Tag{
		Name: "div",
		Attrs: groupAttr.
			String("id", it.ID).
			Slice("class", it.Classes...),
		Content: append(content, it.Additional),
	}

	if it.OnChange != nil {
		var err error
		result, err = f.Mono(result)
		if err != nil {
			return internal.ErrPrependPath(err, "InputText")
		}
		err = f.AddInteraction(TWith{
			Func:    it.OnChange,
			Content: result,
		})
		if err != nil {
			return internal.ErrPrependPath(err, "InputText")
		}
	}

	return internal.ErrPrependPath(result.Init(f, w), "InputText")
}

type InputSearch struct {
	ID      string
	Classes []string
	Attrs   Attributes

	Options     func(r *http.Request) []any
	PrePopulate bool
	Additional  Component
	Target      string

	Name        string
	Label       string
	Placeholder string
	Value       string
	AutoFocus   bool
	Disabled    bool
}

func (it InputSearch) Init(f *Framework, w io.Writer) error {
	addon := Fragment{
		Raw("{{if .options}}"),
		Div{
			Attrs:   Attrs().Slice("style", "position: absolute;", "top: 100%;", "left: 0;", "right: 0;"),
			Classes: []string{"menu"},
			Content: Fragment{
				Raw("{{range .options}}"),
				Button{Content: Raw("{{.}}"), Attr: Attrs().
					String("hx-get", f.Path()+"?search={{.}}").
					String("hx-target", it.Target).
					String("hx-swap", "innerHTML"),
				},
				Raw("{{end}}"),
			},
		},
		Raw("{{end}}"),
	}
	addon = append(addon, it.Additional)
	var input Component
	loadData := func(r *http.Request) Data {
		data := GetDataFromRequest(it.Name)(r)
		options := it.Options(r)
		return data.Merge(Data{"options": options})
	}
	input = InputText{
		ID:      it.ID,
		Classes: it.Classes,
		Attrs:   it.Attrs,

		OnChange:   loadData,
		Additional: addon,

		Name:        it.Name,
		Label:       it.Label,
		Placeholder: it.Placeholder,
		Value:       it.Value,
		AutoFocus:   it.AutoFocus,
		Disabled:    it.Disabled,
	}
	if it.PrePopulate {
		input = TWith{
			Func:    loadData,
			Content: input,
		}
	}
	return input.Init(f, w)
}

// InputHidden defines a hidden input element. This is useful for storing data in a form that isn't visible.
type InputHidden struct {
	ID    string
	Name  string
	Value string
}

func (ih InputHidden) Init(f *Framework, w io.Writer) error {
	return Input{
		ID:    ih.ID,
		Type:  "hidden",
		Name:  ih.Name,
		Value: ih.Value,
	}.Init(f, w)
}

// InputSubmit defines an input submit button.
type InputSubmit struct {
	ID      string
	Classes []string
	Attrs   Attributes

	Text string
}

func (is InputSubmit) Init(f *Framework, w io.Writer) error {
	return Input{
		ID:      is.ID,
		Type:    "submit",
		Classes: is.Classes,
		Attr:    is.Attrs,
		Value:   is.Text,
		Name:    "input",
	}.Init(f, w)
}
