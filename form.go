package gohtmx

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Form struct {
	ID      string
	Classes []string
	Style   []string
	Attr    []Attr

	// Used to load Template when this component is rendered
	LoadTemplateData TemplateDataLoaderFunc
	LoadSubmitData   TemplateDataLoaderFunc

	// The interior contents of this element.
	Content Component
	Error   Component
	Success Component
}

func (f Form) LoadTemplate(l *Location, t *TemplateDataLoader, w io.StringWriter) {
	if f.LoadTemplateData != nil {
		t.AddNamedAtPath(f.ID, l.Path(), f.LoadTemplateData)
	}
	Tag{
		Name: "form",
		Attrs: removeEmptyAttributes(append(f.Attr,
			Attr{Name: "id", Value: f.ID},
			Attr{Name: "class", Value: strings.Join(f.Classes, " ")},
			Attr{Name: "style", Value: strings.Join(f.Style, ";")},
			Attr{Name: "hx-post", Value: l.Path(f.ID)},
			Attr{Name: "hx-target", Value: fmt.Sprintf("#%s-results", f.ID)},
		)),
		Content: Fragment{
			f.Content,
			Div{
				ID: fmt.Sprintf("%s-results", f.ID),
			},
		},
	}.LoadTemplate(l.AtData(f.ID), t, w)
}

func (f Form) LoadMux(l *Location, m *mux.Router) {
	errorHandler := l.BuildTemplateHandler(f.Error)
	successHandler := l.BuildTemplateHandler(f.Success)
	// Response
	m.HandleFunc(l.Path(f.ID), func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			errorHandler.ServeHTTPWithExtraData(w, r, TemplateData{"error": err.Error()})
			return
		}
		data, err := f.LoadSubmitData(r)
		if err != nil {
			errorHandler.ServeHTTPWithExtraData(w, r, TemplateData{"error": err.Error()})
			return
		}
		successHandler.ServeHTTPWithExtraData(w, r, TemplateData{
			f.ID: data,
		})
	})
	f.Error.LoadMux(l, m)
	f.Success.LoadMux(l, m)
	f.Content.LoadMux(l, m)
}

type InputText struct {
	Name        string
	Classes     []string
	Style       []string
	Disabled    bool
	Readonly    bool
	Placeholder string
	Value       string
}

func (it InputText) LoadTemplate(l *Location, t *TemplateDataLoader, w io.StringWriter) {
	Tag{
		Name: "input",
		Attrs: removeEmptyAttributes([]Attr{
			{Name: "name", Value: it.Name},
			{Name: "type", Value: "text"},
			{Name: "class", Value: strings.Join(it.Classes, " ")},
			{Name: "style", Value: strings.Join(it.Style, ";")},
			{Name: "placeholder", Value: it.Placeholder},
			{Name: "disabled", Enabled: it.Disabled},
			{Name: "readonly", Enabled: it.Readonly},
			{Name: "value", Value: orDefault(it.Value, fmt.Sprintf(`{{or %s ""}}`, l.Data(it.Name)))},
		}),
	}.LoadTemplate(l, t, w)
}

func (it InputText) LoadMux(l *Location, m *mux.Router) {
	// NoOp
}

type InputHidden struct {
	Name  string
	Value string
}

func (ih InputHidden) LoadTemplate(l *Location, t *TemplateDataLoader, w io.StringWriter) {
	Tag{
		Name: "input",
		Attrs: removeEmptyAttributes([]Attr{
			{Name: "name", Value: ih.Name},
			{Name: "type", Value: "hidden"},
			{Name: "value", Value: orDefault(ih.Value, fmt.Sprintf(`{{or %s ""}}`, l.Data(ih.Name)))},
		}),
	}.LoadTemplate(l, t, w)
}

func (ih InputHidden) LoadMux(l *Location, m *mux.Router) {
	// NoOp
}

type InputSubmit struct {
	ID       string
	Classes  []string
	Style    []string
	Disabled bool

	Text string
}

func (is InputSubmit) LoadTemplate(l *Location, t *TemplateDataLoader, w io.StringWriter) {
	Tag{
		Name: "input",
		Attrs: removeEmptyAttributes([]Attr{
			{Name: "id", Value: is.ID},
			{Name: "type", Value: "submit"},
			{Name: "class", Value: strings.Join(is.Classes, " ")},
			{Name: "style", Value: strings.Join(is.Style, ";")},
			{Name: "disabled", Enabled: is.Disabled},
			{Name: "value", Value: orDefault(is.Text, "Submit")},
		}),
	}.LoadTemplate(l, t, w)
}

func (is InputSubmit) LoadMux(l *Location, m *mux.Router) {
	// NoOp
}
