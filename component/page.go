package component

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Document the baseline component of an HTML document/page.
type Document struct {
	// Header defines Component to be rendered in between the <head> tags
	Header Component
	// Body defines the Component to be rendered in between the <body> tags
	Body Component
}

func (d Document) WriteTemplate(prefix string, w io.StringWriter) {
	Fragment{
		Raw("<!DOCTYPE html>"),
		Tag{"html", []Attribute{}, Fragment{
			Tag{"head", []Attribute{}, d.Header},
			Tag{"body", []Attribute{}, d.Body},
		}},
	}.WriteTemplate(prefix, w)
}

func (d Document) LoadMux(prefix string, m *http.ServeMux) {
	d.Body.LoadMux(prefix, m)
}

// ServeComponent attaches the passed Component to the mux at the passed prefix.
// All sub-Components will be attached as well under the passed prefix.
func ServeComponent(prefix string, mux *http.ServeMux, c Component) error {
	prefix = strings.TrimRight(prefix, "/")
	body := BuildTemplate("body", prefix, c)

	internal := http.NewServeMux()
	c.LoadMux(prefix, internal)
	mux.HandleFunc(fmt.Sprintf("%s/", prefix), func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(TemplateDataFromRequestOnContext(r.Context(), r))
		// All non-HTMX/event-stream requests should serve as a SPA and loaded through the golang html/template.
		if r.Header.Get("HX-Request") != "" || r.Header.Get("Accept") == "text/event-stream" {
			internal.ServeHTTP(w, r)
		} else {
			_ = body.Execute(w, DataFromContext(r.Context()))
		}
	})
	return nil
}
