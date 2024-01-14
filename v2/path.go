package gohtmx

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Path struct {
	ID      string
	Classes []string
	Style   []string
	Attr    []Attr

	// Paths defines the possible paths and their components.
	Paths map[string]Component
	// DefaultPath defines the path to automatically switch to when no path matches.
	DefaultPath string
	// DefaultComponent defines the component to load when no path matches.
	DefaultComponent Component
}

func (p Path) Init(f *Framework, w io.Writer) error {
	if p.ID == "" {
		return fmt.Errorf("path component requires an id")
	}

	conditions := make(Conditions, len(p.Paths))
	for key, content := range p.Paths {
		path := f.Path(key)
		// Initial Load
		conditions = append(conditions, Condition{
			Condition: func(r *http.Request) bool {
				return strings.HasPrefix(r.URL.Path, path) &&
					(len(r.URL.Path) == len(path) || r.URL.Path[len(path)] == '/')
			},
			Content: content,
		})

		// Interactions for other components to use.
		err := f.AtPath(key).AddTemplateInteraction(content, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("HX-Replace-Url", path)
		})
		if err != nil {
			return fmt.Errorf("failed to add path interaction at %s: %w", path, err)
		}
	}
	// Default loading options.
	if p.DefaultPath != "" && p.Paths[p.DefaultPath] != nil {
		conditions = append(conditions, Condition{
			Content: Div{Attr: []Attr{
				{Name: "hx-get", Value: f.Path(p.DefaultPath)},
				{Name: "hx-target", Value: "#" + p.ID},
				{Name: "hx-trigger", Value: "load"},
			}},
		})
	} else if p.DefaultComponent != nil {
		conditions = append(conditions, Condition{
			Content: p.DefaultComponent,
		})
	}
	return Tag{
		Name: "div",
		Attrs: append(p.Attr,
			Attr{Name: "id", Value: p.ID},
			Attr{Name: "class", Value: strings.Join(p.Classes, " ")},
			Attr{Name: "style", Value: strings.Join(p.Style, ";")},
		),
		Content: conditions,
	}.Init(f, w)
}
