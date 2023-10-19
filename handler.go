package gohtmx

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

// SingleFieldHandler acts as an HTTP handler and handles extracting a single form field name from the request.
type SingleFieldHandler struct {
	Name    string
	Handler func(value string, w http.ResponseWriter, r *http.Request)
}

func (sfh SingleFieldHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "error reading form data", http.StatusInternalServerError)
		return
	}
	value, ok := r.Form[sfh.Name]
	if !ok {
		http.Error(w, fmt.Sprintf("unable to locate form value %s", sfh.Name), http.StatusInternalServerError)
		return
	}
	if len(value) != 1 {
		http.Error(w, fmt.Sprintf("unexpected form value for %s: %s", sfh.Name, strings.Join(value, ",")), http.StatusInternalServerError)
		return
	}
	sfh.Handler(value[0], w, r)
}

// TemplateHandler wrapper for making a template into an http.Handler
type TemplateHandler struct {
	Template *template.Template
}

func (th TemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := th.Template.Execute(w, DataFromContext(r.Context()))
	if err != nil {
		http.Error(w, "error rendering template", http.StatusInternalServerError)
		return
	}
}
