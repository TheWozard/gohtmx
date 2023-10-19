package gohtmx

import (
	"fmt"
	"io"
	"net/http"
)

// TemplateDefinition Component implementation of the golang template/html for {{define "<Name>"}}<Component>{{End}}
type TemplateDefinition struct {
	// Name is the name this template defines.
	Name string
	// Content defines the Component this wraps.
	Content Component
}

func (td TemplateDefinition) WriteTemplate(prefix string, w io.StringWriter) {
	_, _ = w.WriteString(fmt.Sprintf(`{{define "%s"}}`, td.Name))
	if td.Content != nil {
		td.Content.WriteTemplate(prefix, w)
	}
	_, _ = w.WriteString(`{{end}}`)
}

func (td TemplateDefinition) LoadMux(prefix string, m *http.ServeMux) {
	td.Content.LoadMux(prefix, m)
}

// TemplateBlock Component implementation of the golang template/html for {{block "<Name>" <Path>}}<Default>{{end}}
type TemplateBlock struct {
	// Name is the name of the template to be loaded.
	Name string
	// Path defines the path of data is used by the template.
	Path string
	// Default defines the Component that is used if the named template is missing.
	Default Component
}

func (tb TemplateBlock) path() string {
	if tb.Path == "" {
		return "."
	}
	return tb.Path
}

func (tb TemplateBlock) WriteTemplate(prefix string, w io.StringWriter) {
	_, _ = w.WriteString(fmt.Sprintf(`{{template "%s" %s}}`, tb.Name, tb.path()))
	if tb.Default != nil {
		tb.Default.WriteTemplate(prefix, w)
	}
	_, _ = w.WriteString(`{{end}}`)
}

func (tb TemplateBlock) LoadMux(prefix string, m *http.ServeMux) {
	tb.Default.LoadMux(prefix, m)
}

// TemplateCondition represents a single if condition in a TemplateConditionSet.
type TemplateCondition struct {
	// Condition represents the condition the template must match to pass to be rendered.
	// If Empty the Condition is ignored and the Content is rendered without template wrapper.
	// If Empty in a TemplateConditionSet will be treated as an Else clause.
	Condition string
	// Content to render if the Condition passes.
	Content Component
}

func (tc TemplateCondition) WriteTemplate(prefix string, w io.StringWriter) {
	if tc.Condition == "" {
		tc.Content.WriteTemplate(prefix, w)
	} else {
		_, _ = w.WriteString(fmt.Sprintf(`{{if %s}}`, tc.Condition))
		tc.Content.WriteTemplate(prefix, w)
		_, _ = w.WriteString(`{{end}}`)
	}
}

func (tc TemplateCondition) LoadMux(prefix string, m *http.ServeMux) {
	tc.Content.LoadMux(prefix, m)
}

// TemplateConditionSet is a slice that represents an If/IfElse/Else template.
// TemplateCondition.Conditions are evaluated in order. Any empty TemplateCondition.Condition are grouped under a single else.
type TemplateConditionSet []TemplateCondition

func (tcs TemplateConditionSet) WriteTemplate(prefix string, w io.StringWriter) {
	if len(tcs) == 0 {
		return
	}

	elses := Fragment{}
	first := true
	for _, condition := range tcs {
		if condition.Condition == "" {
			elses = append(elses, condition.Content)
			continue
		} else if first {
			_, _ = w.WriteString(fmt.Sprintf(`{{if %s}}`, condition.Condition))
			first = false
		} else {
			_, _ = w.WriteString(fmt.Sprintf(`{{else if %s}}`, condition.Condition))
		}
		condition.Content.WriteTemplate(prefix, w)
	}
	if first {
		// No other conditions so no need to add template info.
		elses.WriteTemplate(prefix, w)
		return
	}
	_, _ = w.WriteString(`{{else}}`)
	elses.WriteTemplate(prefix, w)
	_, _ = w.WriteString(`{{end}}`)
}

func (tcs TemplateConditionSet) LoadMux(prefix string, m *http.ServeMux) {
	for _, condition := range tcs {
		condition.Content.LoadMux(prefix, m)
	}
}
