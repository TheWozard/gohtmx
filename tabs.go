package gohtmx

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Tabs Component for rendering a list of Tab Components.
type Tabs struct {
	// The unique ID of this Tabs Component used to for HTMX targeting.
	ID string
	// Tabs is a slice of Tab elements to be rendered.
	Tabs []Tab
	// DefaultRedirect if not empty will automatically redirect to the Tab of Value equal to Default Redirect
	// If the tab is missing DefaultRedirect will be ignored.
	DefaultRedirect string
	// Classes to be added to the wrapping div element.
	Classes []string
	// ActiveClasses to be added to the Active tab element.
	ActiveClasses []string
}

func (t Tabs) WriteTemplate(prefix string, w io.StringWriter) {
	Tag{
		"div",
		[]Attribute{
			{"id", t.ID},
		},
		t.content(prefix),
	}.WriteTemplate(prefix, w)
}

func (t Tabs) content(prefix string) Component {
	target := "#" + t.ID

	tags := Fragment{}
	conditions := TemplateConditionSet{}
	for _, tab := range t.Tabs {
		tags = append(tags, Tag{"li",
			[]Attribute{
				{"class", tab.IfCondition(prefix, strings.Join(t.ActiveClasses, " "))},
			},
			Tag{
				"a",
				tab.HTMXAttributes(prefix, target),
				tab.Tag,
			},
		})
		conditions = append(conditions, tab.AsCondition(prefix, target))
	}

	if t.DefaultRedirect != "" {
		for _, tab := range t.Tabs {
			if tab.Value == t.DefaultRedirect {
				conditions = append(conditions, TemplateCondition{
					Content: Tag{
						"div",
						append(tab.HTMXAttributes(prefix, target),
							Attribute{"hx-trigger", "load"},
						),
						nil,
					},
				})
			}
		}
	}

	return Fragment{
		Tag{
			"div",
			[]Attribute{
				{"class", strings.Join(t.Classes, " ")},
			},
			Tag{
				"ul",
				[]Attribute{},
				tags,
			},
		},
		conditions,
	}
}

func (t Tabs) LoadMux(prefix string, m *http.ServeMux) {
	tabTemplate := BuildTemplate(t.ID, prefix, t.content(prefix))
	// All the logic is in the template itself.
	m.Handle(prefix+"/", TemplateHandler{Template: tabTemplate})
	for _, tab := range t.Tabs {
		// We register each tab path "prefix/value" so if the content contains a tab as well
		// when it registers "prefix/value/" golang doesn't auto redirect our requests.
		// TODO: This is another indicator to get a better MUX.
		m.Handle(tab.Path(prefix), TemplateHandler{Template: tabTemplate})
		tab.LoadMux(prefix, m)
	}
}

// Tab represents a common handler for tab style components.
type Tab struct {
	// Value is the text value that corresponds to this tab. This is used in the Path to identify the tab.
	Value string
	// Tag represents the Component that should be rendered for this tabs tag.
	Tag Component
	// Content represents the Component to render for this tab.
	Content Component
}

// Path returns the path for this tab.
func (t Tab) Path(prefix string) string {
	return fmt.Sprintf("%s/%s", prefix, strings.ToLower(t.Value))
}

// Condition returns the template condition that would match this tab.
func (t Tab) Condition(prefix string) string {
	path := t.Path(prefix)
	return fmt.Sprintf(`or (eq .Path "%s") (hasPrefix .Path "%s/")`, path, path)
}

// IfCondition wraps the passed content in a template if condition.
// TODO: Hack in the best of case.
func (t Tab) IfCondition(prefix string, content string) string {
	return fmt.Sprintf(`{{if %s}}%s{{end}}`, t.Condition(prefix), content)
}

// HTMXAttributes the standard HTMX Attributes required to move to this tab.
func (t Tab) HTMXAttributes(prefix string, target string) []Attribute {
	return []Attribute{
		{"hx-get", t.Path(prefix)},
		{"hx-target", target},
		{"hx-push-url", "true"},
	}
}

// AsCondition converts the tab.Content to a TemplateCondition to make it conditionally render on init.
func (t Tab) AsCondition(prefix string, target string) TemplateCondition {
	return TemplateCondition{
		Condition: t.Condition(prefix),
		Content:   At{t.Path(prefix), t.Content},
	}
}

func (t Tab) LoadMux(prefix string, m *http.ServeMux) {
	t.Content.LoadMux(t.Path(prefix), m)
}
