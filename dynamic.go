package gohtmx

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

// Dynamic defines a functional component that is rendered per request. The return value
// being template.HTML means the content of this component is not escaped and can lead to Cross-Site Scripting (XSS) attacks.
// These Components are rendered using a Slim version of the Framework. For more info see *Framework.Slim() documentation.
type Dynamic func(f *Framework, r *http.Request) template.HTML

func (d Dynamic) Init(f *Framework, w io.Writer) error {
	if !f.CanTemplate() {
		return ErrCannotTemplate
	}
	id := f.Generator.NewFunctionID(d)
	slim := f.Slim()
	f.Template = f.Template.Funcs(template.FuncMap{
		id: func(r *http.Request) template.HTML {
			return d(slim, r)
		},
	})
	return Raw(fmt.Sprintf("{{%s .request}}", id)).Init(f, w)
}

// Preview defines a component that can be used to view the resulting template of its contents.
// This is useful for debugging and testing, but could also be used to modify a template before parsing.
type Preview struct {
	View    func(c string) string
	Content Component
}

func (p Preview) Init(f *Framework, w io.Writer) error {
	buffer := bytes.NewBuffer(nil)
	contentErr := p.Content.Init(f, buffer)
	if p.View != nil {
		_, err := w.Write([]byte(p.View(buffer.String())))
		if err != nil {
			return fmt.Errorf("failed to write viewed preview: %w", err)
		}
	} else {
		_, err := w.Write(buffer.Bytes())
		if err != nil {
			return fmt.Errorf("failed to write preview: %w", err)
		}
	}
	return contentErr
}

type Repeated struct {
	Content Component
}

func (r Repeated) Init(f *Framework, w io.Writer) error {
	return r.Content.Init(f.NoMux(), w)
}
