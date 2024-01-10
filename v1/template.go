package gohtmx

import (
	"fmt"
	"html/template"
	"io"
	"strings"
)

func NewVariable(name string, get any, vars ...string) TemplateVariable {
	return TemplateVariable{Name: name, Func: get, Vars: vars}
}

type TemplateVariable struct {
	Name string
	Func any
	Vars []string
}

func (tv TemplateVariable) Init(f *Framework, w io.Writer) error {
	if f.CanTemplate() {
		id := f.Generator.NewFunctionID(tv.Func)
		f.Template = f.Template.Funcs(template.FuncMap{
			id: tv.Func,
		})
		return Raw(fmt.Sprintf("{{%s := %s}}", tv.Name, strings.Join(append([]string{id}, tv.Vars...), " "))).Init(f, w)
	}
	return ErrCannotTemplate
}

type Condition struct {
	Condition any
	Vars      []string
	Content   Component
}

func (tc Condition) Init(f *Framework, w io.Writer) error {
	if f.CanTemplate() {
		id := f.Generator.NewFunctionID(tc.Condition)
		f.Template = f.Template.Funcs(template.FuncMap{
			id: tc.Condition,
		})
		err := Raw(fmt.Sprintf("{{if %s}}", strings.Join(append([]string{id}, tc.Vars...), " "))).Init(f, w)
		if err != nil {
			return fmt.Errorf("failed to write condition prefix: %w", err)
		}
		err = tc.Content.Init(f, w)
		if err != nil {
			return fmt.Errorf("failed to write condition content: %w", err)
		}
		err = Raw("{{end}}").Init(f, w)
		if err != nil {
			return fmt.Errorf("failed to write condition suffix: %w", err)
		}
		return nil
	}
	return ErrCannotTemplate
}
