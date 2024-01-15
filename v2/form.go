package gohtmx

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/TheWozard/gohtmx/v2/core"
)

// -- Value Handlers --

func RequestFloat(r *http.Request, key string, def float64) float64 {
	raw := RequestValue(r, key)
	if raw == "" {
		return def
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return def
	}
	return val
}

func RequestValue(r *http.Request, key string) string {
	if r.Method == http.MethodGet {
		return r.URL.Query().Get(key)

	}
	err := r.ParseForm()
	if err != nil {
		return ""
	}
	return r.Form.Get(key)
}

// -- Components --

type Form struct {
	ID      string
	Classes []string
	Style   []string
	Attr    []Attr

	// Action defines what happens when the form is submitted.
	// core.TemplateData is the data that will be passed to the success Component.
	// If an error occurs, the error Component will be rendered instead.
	Action          func(w http.ResponseWriter, r *http.Request) (core.TemplateData, error)
	CanAutoComplete func(r *http.Request) bool

	// The interior contents of this element.
	Content Component
	Error   Component
	Success Component
}

func (fr Form) Init(f *Framework, w io.Writer) error {
	f = f.AtPath(fr.ID)
	successHandler, err := f.NewTemplateHandler(fr.Success)
	if err != nil {
		return fmt.Errorf("failed to create success handler: %w", err)
	}
	errorHandler, err := f.NewTemplateHandler(fr.Error)
	if err != nil {
		return fmt.Errorf("failed to create success handler: %w", err)
	}
	f.AddInteractionFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			errorHandler.ServeHTTPWithExtraData(w, r, core.TemplateData{"error": err.Error()})
			return
		}
		var data core.TemplateData
		if fr.Action != nil {
			data, err = fr.Action(w, r)
			if err != nil {
				errorHandler.ServeHTTPWithExtraData(w, r, core.TemplateData{"error": err.Error()})
				return
			}
		}
		successHandler.ServeHTTPWithExtraData(w, r, data)
	})
	var autoload Component
	if fr.CanAutoComplete != nil {
		autoload = TCondition{
			Condition: fr.CanAutoComplete,
			Content:   Repeated{Content: fr.Success},
		}
	}
	return Fragment{
		Tag{
			Name: "form",
			Attrs: append(fr.Attr,
				Attr{Name: "id", Value: fr.ID},
				Attr{Name: "class", Value: strings.Join(fr.Classes, " ")},
				Attr{Name: "style", Value: strings.Join(fr.Style, ";")},
				Attr{Name: "hx-post", Value: f.Path()},
				Attr{Name: "hx-target", Value: fmt.Sprintf("#%s-results", fr.ID)},
				Attr{Name: "hx-trigger", Value: "submit"},
				Attr{Name: "hx-swap", Value: "innerHTML"},
				Attr{Name: "autocomplete", Value: "off"},
			),
			Content: fr.Content,
		},
		Div{
			ID:      fmt.Sprintf("%s-results", fr.ID),
			Content: autoload,
		},
	}.Init(f, w)
}

type InputText struct {
	ID      string
	Classes []string
	Style   []string
	Attr    []Attr

	Validate   func(r *http.Request) core.TemplateData
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
			Name: "label",
			Attrs: []Attr{
				{Name: "for", Value: it.ID},
			},
			Content: Raw(it.Label),
		})
	}
	// Input
	inputAttr := []Attr{}
	groupAttr := it.Attr
	if it.Validate != nil {
		inputAttr = append(inputAttr,
			Attr{Name: "hx-trigger", Value: "keyup changed delay:500ms"},
			Attr{Name: "hx-post", Value: c.Path()},
		)
		groupAttr = append(groupAttr,
			Attr{Name: "hx-target", Value: "this"},
			Attr{Name: "hx-swap", Value: "morph:outerHTML"},
		)
		if f.IsInteractive() {
			handler, err := f.NoMux().NewTemplateHandler(it)
			if err != nil {
				return fmt.Errorf("failed to create validation handler: %w", err)
			}
			c.AddInteractionFunc(func(w http.ResponseWriter, r *http.Request) {
				data := it.Validate(r)
				handler.ServeHTTPWithExtraData(w, r, data)
			})
		}
	}
	content = append(content, Tag{
		Name: "input",
		Attrs: append(inputAttr,
			Attr{Name: "type", Value: "text"},
			Attr{Name: "name", Value: core.FirstNonEmptyString(it.Name, it.ID)},
			Attr{Name: "placeholder", Value: core.FirstNonEmptyString(it.Placeholder, it.ID)},
			Attr{Name: "value", Value: it.Value},
		),
	})

	content = append(content, it.Additional)

	return Tag{
		Name: "div",
		Attrs: append(groupAttr,
			Attr{Name: "id", Value: it.ID},
			Attr{Name: "class", Value: strings.Join(it.Classes, " ")},
			Attr{Name: "style", Value: strings.Join(it.Style, ";")},
		),
		Content: content,
	}.Init(f, w)
}

type InputHidden struct {
	ID    string
	Name  string
	Value string
}

func (ih InputHidden) Init(f *Framework, w io.Writer) error {
	return Tag{
		Name: "input",
		Attrs: []Attr{
			{Name: "id", Value: ih.ID},
			{Name: "type", Value: "hidden"},
			{Name: "name", Value: core.FirstNonEmptyString(ih.Name, ih.ID)},
			{Name: "value", Value: ih.Value},
		},
	}.Init(f, w)
}

type InputSubmit struct {
	ID      string
	Classes []string
	Style   []string
	Attr    []Attr

	Text string
}

func (is InputSubmit) Init(f *Framework, w io.Writer) error {
	return Tag{
		Name: "input",
		Attrs: append(is.Attr,
			Attr{Name: "id", Value: is.ID},
			Attr{Name: "class", Value: strings.Join(is.Classes, " ")},
			Attr{Name: "style", Value: strings.Join(is.Style, ";")},
			Attr{Name: "type", Value: "submit"},
			Attr{Name: "value", Value: is.Text},
		),
	}.Init(f, w)
}

type InputSearch struct {
	ID      string
	Classes []string
	Style   []string
	Attr    []Attr

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
			Style:   []string{"position: absolute; top: 100%; left: 0; right: 0;"},
			Classes: []string{"menu"},
			Content: Fragment{
				Raw("{{range .options}}"),
				Button{Content: Raw("{{.}}"), Attr: []Attr{
					{Name: "hx-get", Value: f.Path() + "?search={{.}}"},
					{Name: "hx-target", Value: it.Target},
					{Name: "hx-swap", Value: "innerHTML"},
				}},
				Raw("{{end}}"),
			},
		},
		Raw("{{end}}"),
	}
	addon = append(addon, it.Additional)
	var input Component
	loadData := func(r *http.Request) core.TemplateData {
		data := LoadData(it.Name)(r)
		options := it.Options(r)
		return data.Merge(core.TemplateData{"options": options})
	}
	input = InputText{
		ID:      it.ID,
		Classes: it.Classes,
		Style:   it.Style,
		Attr:    it.Attr,

		Validate:   loadData,
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
