package component

import (
	"io"
	"net/http"
)

type Page struct {
	Header Component
	Body   Component
}

func (p Page) WriteTemplate(w io.StringWriter) {
	Fragment{
		Raw("<!DOCTYPE html>"),
		Tag{"html", []Attribute{}, Fragment{
			Tag{"head", []Attribute{}, p.Header},
			Tag{"body", []Attribute{}, p.Body},
		}},
	}.WriteTemplate(w)
}

func (p Page) LoadMux(m *http.ServeMux) {
	p.Body.LoadMux(m)
}
