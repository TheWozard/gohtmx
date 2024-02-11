package gohtmx

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/TheWozard/gohtmx/element"
)

// TWith defines a template block to be executed with . being set to the result of the Func.
type TWith struct {
	Func    func(*http.Request) any
	Content Component
}

func (t TWith) Init(p *Page) (element.Element, error) {
	if t.Func == nil {
		return t.Content.Init(p)
	}
	id := p.Generator.NewID("func")
	p.Template = p.Template.Funcs(template.FuncMap{id: t.Func})
	return element.TBlock{
		Text:       fmt.Sprintf(`with %s $r`, id),
		IncludeEnd: true,
		Element:    p.Init(t.Content),
	}, nil
}
