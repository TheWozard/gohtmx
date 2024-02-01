package gohtmx

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

type TemplateHandler struct {
	Template *template.Template
	Name     string
}

func (t TemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.ServeHTTPWithData(w, r, nil)
}

func (t TemplateHandler) ServeHTTPWithData(w http.ResponseWriter, r *http.Request, data Data) {
	raw, err := t.ExecuteWith(r, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(raw)
}

func (t TemplateHandler) ExecuteWith(r *http.Request, data Data) ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	err := t.Template.ExecuteTemplate(buffer, t.Name, data.Merge(Data{"request": r}))
	if err != nil {
		return nil, fmt.Errorf(`failed to render template %s: %w`, t.Name, err)
	}
	return buffer.Bytes(), err
}
