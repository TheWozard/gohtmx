package gohtmx

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"github.com/TheWozard/gohtmx/v2/core"
)

func NewTemplateHandler(f *Framework, component Component) (*TemplateHandler, error) {
	if !f.CanTemplate() {
		return nil, fmt.Errorf("failed to render component: %w", ErrCannotTemplate)
	}
	if component == nil {
		return nil, fmt.Errorf("failed to render component: %w", ErrNilComponent)
	}
	data := bytes.NewBuffer(nil)
	err := Fragment{
		Raw("{{$r := .request}}"),
		component,
	}.Init(f, data)
	if err != nil {
		return nil, fmt.Errorf("failed to render component: %w", err)
	}
	name := f.Generator.NewGroupID("template")
	f.Template, err = f.Template.New(name).Parse(data.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	return &TemplateHandler{
		Template: f.Template,
		Name:     name,
	}, nil
}

type TemplateHandler struct {
	Template *template.Template
	Name     string
}

func (t TemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	buffer := bytes.NewBuffer(nil)
	err := t.Template.ExecuteTemplate(buffer, t.Name, core.DataFromContext(r.Context()).Merge(core.TemplateData{"request": r}))
	if err != nil {
		http.Error(w, fmt.Sprintf(`error rendering template interaction: , %s`, err.Error()), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(buffer.Bytes())
}

func (t TemplateHandler) ServeHTTPWithExtraData(w http.ResponseWriter, r *http.Request, data core.TemplateData) {
	t.ServeHTTP(w, r.WithContext(core.DataFromContext(r.Context()).Merge(data).Context(r.Context())))
}
