package gohtmx

import (
	"fmt"
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
	// UpdateForm adds an out of band update to the form on success.
	// This will update the form regardless of what element calls the form interaction endpoint.
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
	attrs := fr.Attrs.
		Value("id", fr.ID).
		Values("class", fr.Classes...).
		Value("hx-post", f.Path()).
		Value("hx-target", fr.Target).
		Value("autocomplete", "off")
	successComponent, errorComponent := fr.Success, fr.Error
	if fr.UpdateForm {
		successComponent = Fragment{
			Tag{
				Name:    "form",
				Attrs:   attrs.Copy().Value("hx-swap-oob", "true").Value("hx-target", "#"+fr.ID),
				Content: MetaDisableInteraction{Content: fr.Content},
			},
			successComponent,
		}
		errorComponent = Fragment{
			Tag{
				Name:    "form",
				Attrs:   attrs.Copy().Value("hx-swap-oob", "true").Value("hx-target", "#"+fr.ID),
				Content: MetaDisableInteraction{Content: fr.Content},
			},
			errorComponent,
		}
	}
	successHandler := v.RequireTemplateHandler("Success", f, successComponent)
	errorHandler := v.RequireTemplateHandler("Error", f, errorComponent)
	if v.HasError() {
		return AddPathToError(v.Error(), "Form")
	}

	// Interaction
	f.AddInteractionFunc(func(w http.ResponseWriter, r *http.Request) {
		data := GetAllDataFromRequest(r)
		if fr.Submit != nil {
			var err error
			data, err = fr.Submit(data)
			if err != nil {
				errorHandler.ServeHTTPWithData(w, r, Data{"error": err})
				return
			}
		}
		if len(fr.UpdateParams) > 0 {
			data.Subset(fr.UpdateParams...).SetInResponse(w, r)
		}
		successHandler.ServeHTTPWithData(w, r, data)
	})

	if fr.AutoSubmit != nil {
		var err error
		attrs, err = attrs.Condition(f, fr.AutoSubmit, Attrs().Value("hx-trigger", "load,submit"))
		if err != nil {
			return AddPathToError(err)
		}
	}

	// Rendering
	return AddPathToError(Tag{Name: "form", Attrs: attrs, Content: fr.Content}.Init(f, w), "Form")
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
}

func (it InputText) Init(f *Framework, w io.Writer) error {
	c := f.AtPath(core.FirstNonEmptyString(it.ID, it.Name))
	content := Fragment{}
	// Label
	if it.Label != "" {
		content = append(content, Tag{
			Name:    "label",
			Attrs:   Attrs().Value("for", it.ID),
			Content: Raw(it.Label),
		})
	}
	// Input
	inputAttr := Attrs()
	groupAttr := it.Attrs
	if it.OnChange != nil {
		inputAttr = inputAttr.
			Value("hx-trigger", "keyup changed delay:500ms").
			Value("hx-post", c.Path())
		groupAttr = groupAttr.
			Value("hx-target", "this").
			Value("hx-swap", "morph:outerHTML")
		if f.IsInteractive() {
			handler, err := f.NoMux().NewTemplateHandler(it)
			if err != nil {
				return fmt.Errorf("failed to create validation handler: %w", err)
			}
			c.AddInteractionFunc(func(w http.ResponseWriter, r *http.Request) {
				data := it.OnChange(r)
				handler.ServeHTTPWithData(w, r, data)
			})
		}
	}
	content = append(content, Tag{
		Name: "input",
		Attrs: inputAttr.
			Value("type", "text").
			Value("name", core.FirstNonEmptyString(it.Name, it.ID)).
			Value("placeholder", core.FirstNonEmptyString(it.Placeholder, it.ID)).
			Value("value", it.Value),
	})

	content = append(content, it.Additional)

	return Tag{
		Name: "div",
		Attrs: groupAttr.
			Value("id", it.ID).
			Values("class", it.Classes...),
		Content: content,
	}.Init(f, w)
}

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
	}
	if it.PrePopulate {
		input = TWith{
			Func:    loadData,
			Content: input,
		}
	}
	return input.Init(f, w)
}
