package gohtmx

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
)

type TBlock struct {
	Text       string
	IncludeEnd bool
	Element    Element
	Component  Component
}

func (t TBlock) Init(p *Page) (Element, error) {
	if t.Component != nil && t.Element == nil {
		t.Element = p.Init(t.Component)
	}
	return t, nil
}

func (t TBlock) Validate() error {
	if t.Element == nil {
		return nil
	}
	return t.Element.Validate()
}

func (t TBlock) Render(w io.Writer) error {
	_, err := w.Write([]byte("{{" + t.Text + "}}"))
	if err != nil {
		return err
	}
	if t.Element != nil {
		err = t.Element.Render(w)
		if err != nil {
			return err
		}
	}
	if t.IncludeEnd {
		_, err = w.Write([]byte(`{{end}}`))
		if err != nil {
			return err
		}
	}
	return nil
}

func (t TBlock) FindAttrs() (*Attributes, error) {
	return t.Element.FindAttrs()
}

// TWith defines a template block to be executed with . being set to the result of the Func.
type TWith struct {
	Func    func(*http.Request) any
	Content Component
}

func (t TWith) Init(p *Page) (Element, error) {
	if t.Func == nil {
		return t.Content.Init(p)
	}
	id := p.Generator.NewID("func")
	p.Template = p.Template.Funcs(template.FuncMap{id: t.Func})
	return TBlock{
		Text:       fmt.Sprintf(`with %s $r`, id),
		IncludeEnd: true,
		Element:    p.Init(t.Content),
	}, nil
}
