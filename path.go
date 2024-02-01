package gohtmx

// import (
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"strings"

// 	"github.com/TheWozard/gohtmx/internal"
// )

// func IsRequestAtPath(path string) func(r *http.Request) bool {
// 	return func(r *http.Request) bool {
// 		return strings.HasPrefix(r.URL.Path, path) &&
// 			(len(r.URL.Path) == len(path) || r.URL.Path[len(path)] == '/')
// 	}
// }

// type Path struct {
// 	ID      string
// 	Classes []string
// 	Attrs   Attributes

// 	// Paths defines the possible paths and their components.
// 	Paths map[string]Component
// 	// DefaultPath defines the path to automatically switch to when no path matches. Only valid if also exists in Paths.
// 	DefaultPath string
// 	// DefaultComponent defines the component to load when no path matches. Will be ignored if DefaultPath is set and valid.
// 	DefaultComponent Component
// }

// func (p Path) Init(f *Page, w io.Writer) error {
// 	v := internal.NewValidate()
// 	v.Require(p.ID != "", "ID required")
// 	if p.DefaultPath != "" {
// 		_, ok := p.Paths[p.DefaultPath]
// 		v.Require(ok, "DefaultPath must be a valid path")
// 	}
// 	if v.HasError() {
// 		return fmt.Errorf("Path failed to validate: %w", v.Error())
// 	}

// 	conditions := TConditions{}
// 	for key, content := range p.Paths {
// 		f := f.AtPath(key)
// 		path := f.Path()
// 		mono, err := f.Mono(content)
// 		if err != nil {
// 			return internal.ErrEnclosePath(err, "Path")
// 		}
// 		// Initial load.
// 		conditions = append(conditions, TCondition{
// 			Condition: IsRequestAtPath(path),
// 			Content:   mono,
// 		})

// 		// Interactions for other components to use. Initial load will mount any interactions.
// 		f.Use(func(h http.Handler) http.Handler {
// 			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 				w.Header().Add("HX-Replace-Url", path)
// 				h.ServeHTTP(w, r)
// 			})
// 		})
// 		err = f.AddInteraction(mono)
// 		if err != nil {
// 			return internal.ErrEnclosePath(err, "Path")
// 		}
// 	}
// 	if p.DefaultPath != "" && p.Paths[p.DefaultPath] != nil {
// 		// Default loading options - because we cant set headers on load we do the next best with an onload callback.
// 		conditions = append(conditions, TCondition{
// 			Content: Div{Attrs: Attributes{}.
// 				String("hx-get", f.Path(p.DefaultPath)).
// 				String("hx-target", "#"+p.ID).
// 				String("hx-trigger", "load"),
// 			},
// 		})
// 	} else if p.DefaultComponent != nil {
// 		conditions = append(conditions, TCondition{
// 			Content: p.DefaultComponent,
// 		})
// 	}
// 	return internal.ErrEnclosePath(Tag{
// 		Name: "div",
// 		Attrs: p.Attrs.
// 			String("id", p.ID).
// 			String("class", strings.Join(p.Classes, " ")),
// 		Content: conditions,
// 	}.Init(f, w), "Path")
// }
