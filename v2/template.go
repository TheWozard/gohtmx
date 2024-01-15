package gohtmx

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/TheWozard/gohtmx/v2/core"
)

var ErrCannotTemplate = fmt.Errorf("templating is not enabled")
var ErrInvalidVariableName = fmt.Errorf("invalid variable name")
var ErrMissingAction = fmt.Errorf("missing action")
var ErrMissingFunction = fmt.Errorf("missing function")

// TAction defines a template action.
type TAction string

func (t TAction) Init(f *Framework, w io.Writer) error {
	if !f.CanTemplate() {
		return ErrCannotTemplate
	}
	if t == "" {
		return ErrMissingAction
	}
	err := Raw("{{"+t+"}}").Init(f, w)
	if err != nil {
		return fmt.Errorf("failed to write template action: %w", err)
	}
	return nil
}

// TBlock defines a template block that requires one end statement.
type TBlock struct {
	// The action of this block.
	Action string
	// The interior contents of this template block.
	Content Component
}

func (t TBlock) Init(f *Framework, w io.Writer) error {
	if !f.CanTemplate() {
		return ErrCannotTemplate
	}
	if t.Action == "" {
		return ErrMissingAction
	}
	if t.Content == nil {
		return ErrMissingContent
	}
	err := Raw("{{"+t.Action+"}}").Init(f, w)
	if err != nil {
		return fmt.Errorf("failed to write template block prefix: %w", err)
	}
	err = t.Content.Init(f, w)
	if err != nil {
		return fmt.Errorf("failed to write template block content: %w", err)
	}
	err = Raw("{{end}}").Init(f, w)
	if err != nil {
		return fmt.Errorf("failed to write template block suffix: %w", err)
	}
	return nil
}

// TBlocks defines a list of template blocks that require one end statement between them.
type TBlocks []TBlock

func (t TBlocks) Init(f *Framework, w io.Writer) error {
	if !f.CanTemplate() {
		return ErrCannotTemplate
	}
	if len(t) == 0 {
		return nil
	}
	for i, b := range t {
		if b.Action == "" {
			return fmt.Errorf("missing action for template block[%d]: %w", i, ErrMissingAction)
		}
		if b.Content == nil {
			return fmt.Errorf("missing content for template block[%d]: %w", i, ErrMissingContent)
		}
		err := Raw("{{"+b.Action+"}}").Init(f, w)
		if err != nil {
			return fmt.Errorf("failed to write template blocks prefix[%d]: %w", i, err)
		}
		err = b.Content.Init(f, w)
		if err != nil {
			return fmt.Errorf("failed to write template blocks content[%d]: %w", i, err)
		}
	}
	err := Raw("{{end}}").Init(f, w)
	if err != nil {
		return fmt.Errorf("failed to write template blocks suffix: %w", err)
	}
	return nil
}

// TVariable defines a template variable.
type TVariable struct {
	// The name of this variable. Example: "$data".
	Name string
	// The function that will be called to get the value of this variable.
	Func func(*http.Request) any
}

func (t TVariable) Init(f *Framework, w io.Writer) error {
	if !f.CanTemplate() {
		return ErrCannotTemplate
	}
	if t.Name == "" || !strings.HasPrefix(t.Name, "$") {
		return ErrInvalidVariableName
	}
	if t.Func == nil {
		return ErrMissingFunction
	}
	id := f.Generator.NewFunctionID(t.Func)
	f.Template = f.Template.Funcs(template.FuncMap{id: t.Func})
	err := TAction(fmt.Sprintf(`%s := %s $r`, t.Name, id)).Init(f, w)
	if err != nil {
		return fmt.Errorf("failed to write template variable: %w", err)
	}
	return nil
}

// TWith defines a template block to be executed with . being set to the result of the Func.
type TWith struct {
	Func    func(*http.Request) core.TemplateData
	Content Component
}

func (t TWith) Init(f *Framework, w io.Writer) error {
	if !f.CanTemplate() {
		return ErrCannotTemplate
	}
	if t.Func == nil {
		return ErrMissingFunction
	}
	id := f.Generator.NewFunctionID(t.Func)
	f.Template = f.Template.Funcs(template.FuncMap{id: t.Func})
	err := TBlock{
		Action:  fmt.Sprintf(`with %s $r`, id),
		Content: t.Content,
	}.Init(f, w)
	if err != nil {
		return fmt.Errorf("failed to write template with: %w", err)
	}
	return nil
}

// TCondition defines an conditional template block based on Condition.
type TCondition struct {
	Condition func(r *http.Request) bool
	Content   Component
}

func (t TCondition) Init(f *Framework, w io.Writer) error {
	if !f.CanTemplate() {
		return ErrCannotTemplate
	}
	// Special case when no condition
	if t.Condition == nil {
		err := t.Content.Init(f, w)
		if err != nil {
			return fmt.Errorf("failed to write template condition independent content: %w", err)
		}
		return nil
	}
	// Standard case
	id := f.Generator.NewFunctionID(t.Condition)
	f.Template = f.Template.Funcs(template.FuncMap{id: t.Condition})
	err := TBlock{
		Action:  "if" + id + " $r",
		Content: t.Content,
	}.Init(f, w)
	if err != nil {
		return fmt.Errorf("failed to write template condition: %w", err)
	}
	return nil
}

// TConditions defines a list of conditional template blocks that form an if-elseif-else series.
// Any conditions with nil condition will be treated as an else.
type TConditions []TCondition

func (t TConditions) Init(f *Framework, w io.Writer) error {
	if !f.CanTemplate() {
		return ErrCannotTemplate
	}
	elses := Fragment{}
	conditions := TBlocks{}
	for _, tc := range t {
		if tc.Condition == nil {
			elses = append(elses, tc.Content)
			continue
		}
		prefix := "if "
		if len(conditions) > 0 {
			prefix = "else if"
		}
		id := f.Generator.NewFunctionID(tc.Condition)
		f.Template = f.Template.Funcs(template.FuncMap{id: tc.Condition})
		conditions = append(conditions, TBlock{
			Action: prefix + id + " $r",
		})
	}

	if len(elses) > 0 {
		// Special case for only elses
		if len(conditions) == 0 {
			err := elses.Init(f, w)
			if err != nil {
				return fmt.Errorf("failed to write conditions independent else content: %w", err)
			}
			return nil
		}
		conditions = append(conditions, TBlock{
			Action:  "else",
			Content: elses,
		})
	}

	err := conditions.Init(f, w)
	if err != nil {
		return fmt.Errorf("failed to write conditions: %w", err)
	}
	return nil
}
