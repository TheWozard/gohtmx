package gohtmx

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"github.com/TheWozard/gohtmx/v1/core"
)

func NewTemplateHandler(f *Framework, component Component) (*TemplateHandler, error) {
	data := bytes.NewBuffer(nil)
	err := Fragment{
		Raw("{{$r := .request}}"),
		component,
	}.Init(f, data)
	if err != nil {
		return nil, fmt.Errorf("failed to render component: %w", err)
	}
	temp, err := f.Template.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone template: %w", err)
	}
	temp, err = temp.Parse(data.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	return &TemplateHandler{
		Template: temp,
	}, nil
}

type TemplateHandler struct {
	Template *template.Template
}

func (t TemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	buffer := bytes.NewBuffer(nil)
	err := t.Template.Execute(buffer, core.DataFromContext(r.Context()).Merge(core.TemplateData{"request": r}))
	if err != nil {
		http.Error(w, fmt.Sprintf(`error rendering template interaction: , %s`, err.Error()), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(buffer.Bytes())
}

func (t TemplateHandler) ServeHTTPWithExtraData(w http.ResponseWriter, r *http.Request, data core.TemplateData) {
	t.ServeHTTP(w, r.WithContext(core.DataFromContext(r.Context()).Merge(data).Context(r.Context())))
}
