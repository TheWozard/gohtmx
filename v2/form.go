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

	// Used to load Template when this component is rendered
	LoadTemplateData func(r *http.Request) (core.TemplateData, error)

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
		data, err := fr.LoadTemplateData(r)
		if err != nil {
			errorHandler.ServeHTTPWithExtraData(w, r, core.TemplateData{"error": err.Error()})
			return
		}
		successHandler.ServeHTTPWithExtraData(w, r, data)
	})

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
				ID: fmt.Sprintf("%s-results", fr.ID),
			},
		},
	}.Init(f, w)
}
