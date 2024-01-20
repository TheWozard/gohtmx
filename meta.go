package gohtmx

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

// MetaUnsafeDynamic defines a functional component that is rendered per request. The return value
// being template.HTML means the content of this component is not escaped and can lead to Cross-Site Scripting (XSS) attacks.
// These Components are rendered using a Slim version of the Framework. For more info see *Framework.Slim() documentation.
type MetaUnsafeDynamic func(f *Framework, r *http.Request) template.HTML

func (m MetaUnsafeDynamic) Init(f *Framework, w io.Writer) error {
	if !f.CanTemplate() {
		return ErrCannotTemplate
	}
	id := f.Generator.NewFunctionID(m)
	slim := f.Slim()
	f.Template = f.Template.Funcs(template.FuncMap{
		id: func(r *http.Request) template.HTML {
			return m(slim, r)
		},
	})
	return Raw(fmt.Sprintf("{{%s .request}}", id)).Init(f, w)
}

// MetaView defines a component that can be used to view the resulting template of its contents.
// This is useful for debugging and testing, but could also be used to modify a template before parsing.
type MetaView struct {
	View    func(c string) string
	Content Component
}

func (m MetaView) Init(f *Framework, w io.Writer) error {
	buffer := bytes.NewBuffer(nil)
	contentErr := m.Content.Init(f, buffer)
	if m.View != nil {
		_, err := w.Write([]byte(m.View(buffer.String())))
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

// MetaDisableInteraction defines a component that disables interaction without raising interaction errors.
type MetaDisableInteraction struct {
	Content Component
}

func (m MetaDisableInteraction) Init(f *Framework, w io.Writer) error {
	return m.Content.Init(f.NoMux(), w)
}

// MetaAtPath defines a component that modifies the path of the framework for its Content.
type MetaAtPath struct {
	Path    string
	Content Component
}

func (m MetaAtPath) Init(f *Framework, w io.Writer) error {
	return m.Content.Init(f.AtPath(m.Path), w)
}

type UpdateWith struct {
	Paths   []string
	Content Component
}

func (u UpdateWith) Init(f *Framework, w io.Writer) error {
	content, err := f.Mono(u.Content)
	if err != nil {
		return AddPathToError(err, "UpdateWith")
	}
	for _, path := range u.Paths {
		f.AtPath(path).AddOutOfBand(content)
	}
	return AddPathToError(content.Init(f, w), "UpdateWith")
}
