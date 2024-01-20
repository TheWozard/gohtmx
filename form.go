package gohtmx

import (
	"io"
	"net/http"

	"github.com/TheWozard/gohtmx/core"
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
	v := NewValidate()
	v.RequireID(fr.ID)
	v.RequireTarget(fr.Target)
	v.RequireComponent("Success", fr.Success)
	v.RequireComponent("Error", fr.Error)
	if v.HasError() {
		return AddMetaPathToError(v.Error(), "Form")
	}

	content := fr.Content
	attrs := fr.Attrs.
		Value("id", fr.ID).
		Values("class", fr.Classes...).
		Value("hx-post", f.Path()).
		Value("hx-target", fr.Target).
		Value("autocomplete", "off")
	if fr.UpdateForm {
		var err error
		content, err = f.Mono(content)
		if err != nil {
			return AddMetaPathToError(err, "Form")
		}
		f.AddOutOfBand(Tag{
			Name:    "form",
			Attrs:   attrs.Copy().Value("hx-swap-oob", "true").Value("hx-target", "#"+fr.ID),
			Content: content,
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
		return AddMetaPathToError(err, "Form")
	}

	if fr.AutoSubmit != nil {
		attrs = attrs.Condition(fr.AutoSubmit, Attrs().Value("hx-trigger", "load,submit"))
	}

	if len(fr.UpdateParams) > 0 {
		f.Use(UpdateParams(fr.UpdateParams...))
	}

	// Rendering
	return AddMetaPathToError(Tag{Name: "form", Attrs: attrs, Content: content}.Init(f, w), "Form")
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
			Attrs:   Attrs().Value("for", it.ID),
			Content: Raw(it.Label),
		})
	}
	inputAttr := Attrs()
	groupAttr := it.Attrs
	if it.OnChange != nil {
		inputAttr = inputAttr.
			Value("hx-trigger", "keyup changed delay:400ms").
			Value("hx-disabled-elt", "this").
			Value("hx-post", f.Path()).
			Flag("autofocus", it.AutoFocus)
		groupAttr = groupAttr.
			Value("hx-target", "this").
			Value("hx-swap", "morph:outerHTML")
	}
	content = append(content, Tag{
		Name: "input",
		Attrs: inputAttr.
			Value("type", "text").
			Value("name", core.FirstNonEmptyString(it.Name, it.ID)).
			Value("placeholder", core.FirstNonEmptyString(it.Placeholder, it.ID)).
			Value("value", it.Value),
	})
	var result Component
	result = Tag{
		Name: "div",
		Attrs: groupAttr.
			Value("id", it.ID).
			Values("class", it.Classes...),
		Content: append(content, it.Additional),
	}

	if it.OnChange != nil {
		var err error
		result, err = f.Mono(result)
		if err != nil {
			return AddPathToError(err, "InputText")
		}
		err = f.AddInteraction(TWith{
			Func:    it.OnChange,
			Content: result,
		})
		if err != nil {
			return AddPathToError(err, "InputText")
		}
	}

	return AddPathToError(result.Init(f, w), "InputText")
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
			Attrs:   Attrs().Values("style", "position: absolute;", "top: 100%;", "left: 0;", "right: 0;"),
			Classes: []string{"menu"},
			Content: Fragment{
				Raw("{{range .options}}"),
				Button{Content: Raw("{{.}}"), Attr: Attrs().
					Value("hx-get", f.Path()+"?search={{.}}").
					Value("hx-target", it.Target).
					Value("hx-swap", "innerHTML"),
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
