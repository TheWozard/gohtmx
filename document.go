package gohtmx

import (
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Document the baseline component of an HTML document.
type Document struct {
	// Header defines Component to be rendered in between the <head> tags
	Header Component
	// Body defines the Component to be rendered in between the <body> tags
	Body Component
}

func (d Document) LoadTemplate(l *Location, t *TemplateDataLoader, w io.StringWriter) {
	Fragment{
		Raw("<!DOCTYPE html>"),
		Tag{"html", []Attr{}, Fragment{
			Tag{"head", []Attr{}, d.Header},
			Tag{"body", []Attr{{Name: "id", Value: "body"}}, d.Body},
		}},
	}.LoadTemplate(l, t, w)
}

func (d Document) LoadMux(l *Location, m *mux.Router) {
	body := l.BuildTemplateHandler(d)

	internal := mux.NewRouter()
	if d.Body != nil {
		d.Body.LoadMux(l, internal)
	}
	m.PathPrefix(l.Path()).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(MergeDataToContext(r.Context(), TemplateDataFromRequest(r)))
		// All non-HTMX/event-stream requests should serve as a SPA and loaded through the golang html/template.
		if r.Header.Get("HX-Request") != "" || r.Header.Get("Accept") == "text/event-stream" {
			internal.ServeHTTP(w, r)
		} else {
			body.ServeHTTP(w, r)
		}
	})
}

func (d Document) Mount(path string, m *mux.Router) {
	d.LoadMux(d.NewLocation(path), m)
}

func (d Document) NewLocation(path string) *Location {
	return &Location{
		PathPrefix: strings.TrimRight(path, "/"),
		DataPrefix: "",
		TemplateBase: template.New("base").Funcs(template.FuncMap{
			"matchPath": func(match, prefix string) bool {
				return match == prefix || strings.HasPrefix(match, prefix+"/")
			},
		}),
	}
}
