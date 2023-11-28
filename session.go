package gohtmx

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

type Session struct {
	Content Component
}

func (s Session) LoadTemplate(l *Location, w io.StringWriter) {
	s.Content.LoadTemplate(l, w)
}

func (s Session) LoadMux(l *Location, m *mux.Router) {
	m.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m.PathPrefix(l.Path()).Match(r, nil)
			h.ServeHTTP(w, r)
		})
	})
}
