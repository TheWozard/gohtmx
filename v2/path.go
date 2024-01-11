package gohtmx

import (
	"io"
	"net/http"
	"strings"
)

type Path struct {
	ID      string
	Classes []string
	Style   []string
	Attr    []Attr

	Default string
	Paths   map[string]Component
}

func (p Path) Init(f *Framework, w io.Writer) error {
	conditions := make(Conditions, len(p.Paths))
	for key, content := range p.Paths {
		path := f.Path(key)
		conditions = append(conditions, Condition{
			Condition: func(r *http.Request) bool {
				return strings.HasPrefix(r.URL.Path, path) &&
					(len(r.URL.Path) == len(path) || r.URL.Path[len(path)] == '/')
			},
			Content: content,
			Vars:    []string{"$r"},
		})
		f.AtPath(key).AddComponentInteraction(content, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("HX-PUSH-URL", path)
		})
	}
	if p.Default != "" && p.Paths[p.Default] != nil {
		conditions = append(conditions, Condition{
			Content: Div{Attr: []Attr{
				{Name: "hx-get", Value: f.Path(p.Default)},
				{Name: "hx-trigger", Value: "load"},
				{Name: "hx-target", Value: "#" + p.ID},
			}},
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
