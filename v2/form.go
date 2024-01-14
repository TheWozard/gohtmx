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
		autoload = Condition{
			Condition: fr.CanAutoComplete,
			Content:   Repeated{Content: fr.Success},
		}
	}
	return Tag{
		Name: "form",
		Attrs: append(fr.Attr,
			Attr{Name: "id", Value: fr.ID},
			Attr{Name: "class", Value: strings.Join(fr.Classes, " ")},
			Attr{Name: "style", Value: strings.Join(fr.Style, ";")},
			Attr{Name: "hx-post", Value: f.Path()},
			Attr{Name: "hx-target", Value: fmt.Sprintf("#%s-results", fr.ID)},
		),
		Content: Fragment{
			fr.Content,
			Div{
				ID:      fmt.Sprintf("%s-results", fr.ID),
				Content: autoload,
			},
		},
	}.Init(f, w)
}

type InputText struct {
	ID      string
	Classes []string
	Style   []string
	Attr    []Attr

	Name        string
	Placeholder string
	Value       string
}

func (it InputText) Init(f *Framework, w io.Writer) error {
	return Tag{
		Name: "input",
		Attrs: append(it.Attr,
			Attr{Name: "id", Value: it.ID},
			Attr{Name: "class", Value: strings.Join(it.Classes, " ")},
			Attr{Name: "style", Value: strings.Join(it.Style, ";")},
			Attr{Name: "type", Value: "text"},
			Attr{Name: "name", Value: core.FirstNonEmptyString(it.Name, it.ID)},
			Attr{Name: "placeholder", Value: core.FirstNonEmptyString(it.Placeholder, it.ID)},
			Attr{Name: "value", Value: it.Value},
		),
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
