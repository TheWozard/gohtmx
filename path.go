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
	Attrs   Attributes

	// Paths defines the possible paths and their components.
	Paths map[string]Component
	// DefaultPath defines the path to automatically switch to when no path matches. Only valid if also exists in Paths.
	DefaultPath string
	// DefaultComponent defines the component to load when no path matches. Will be ignored if DefaultPath is set and valid.
	DefaultComponent Component
}

func (p Path) Init(f *Framework, w io.Writer) error {
	v := NewValidate()
	v.RequireID(p.ID)
	_, ok := p.Paths[p.DefaultPath]
	v.Require(ok, "DefaultPath must be a valid path")
	if v.HasError() {
		return fmt.Errorf("Path failed to validate: %w", v.Error())
	}

	conditions := TConditions{}
	for key, content := range p.Paths {
		f := f.AtPath(key)
		path := f.Path()
		// Initial load.
		conditions = append(conditions, TCondition{
			Condition: func(r *http.Request) bool {
				return strings.HasPrefix(r.URL.Path, path) &&
					(len(r.URL.Path) == len(path) || r.URL.Path[len(path)] == '/')
			},
			Content: MetaAtPath{Path: key, Content: content},
		})

		// Interactions for other components to use. Initial load will mount any interactions.
		err := f.AddTemplateInteraction(MetaDisableInteraction{Content: content}, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("HX-Replace-Url", path)
		})
		if err != nil {
			return fmt.Errorf("failed to add path interaction at %s: %w", path, err)
		}
	}
	if p.DefaultPath != "" && p.Paths[p.DefaultPath] != nil {
		// Default loading options - because we cant set headers on load we do the next best with an onload callback.
		conditions = append(conditions, TCondition{
			Content: Div{Attrs: Attributes{}.
				Value("hx-get", f.Path(p.DefaultPath)).
				Value("hx-target", "#"+p.ID).
				Value("hx-trigger", "load"),
			},
		})
	} else if p.DefaultComponent != nil {
		conditions = append(conditions, TCondition{
			Content: p.DefaultComponent,
		})
	}
	return Tag{
		Name: "div",
		Attrs: p.Attrs.
			Value("id", p.ID).
			Value("class", strings.Join(p.Classes, " ")),
		Content: conditions,
	}.Init(f, w)
}
