package gohtmx

import (
	"fmt"
	"io"

	"github.com/gorilla/mux"
)

// TabSelector Component for triggering the switching of a TabTarget.
type TabSelector struct {
	// The unique ID of this Tabs Component used to for HTMX targeting.
	ID string
	// The tab to be rendered to the TabTarget when this TabSelector is clicked.
	Tab Tab
	// Classes to be added to the wrapping div element.
	Classes []string

	Content Component
}

func (ts TabSelector) LoadTemplate(l *Location, w io.StringWriter) {
	A{
		Classes: ts.Classes,
		Attr: []Attr{
			{Name: "hx-get", Value: ts.Tab.Path(l)},
			{Name: "hx-target", Value: "#" + ts.ID},
			{Name: "hx-push-url", Value: "true"},
		},
		Content: ts.Content,
	}.LoadTemplate(l, w)
}

func (ts TabSelector) LoadMux(l *Location, m *mux.Router) {
	m.Handle(l.Path(ts.Tab.Value), TemplateHandler{Template: l.BuildTemplate(ts.Tab.Content)})
	ts.Content.LoadMux(l, m)
}

// TabTarget Component for target location to render tabs. Includes features to pre-rendering tabs.
// Will not call LoadMux on any tab contents. That is done by the TabSelector
type TabTarget struct {
	ID      string
	Classes []string

	Tabs         []Tab
	AutoRedirect string
}

func (tt TabTarget) LoadTemplate(l *Location, w io.StringWriter) {
	contents := make(TemplateConditionSet, len(tt.Tabs))
	for i, tab := range tt.Tabs {
		contents[i] = tab.AsCondition(l)
		if tab.Value == tt.AutoRedirect {
			contents = append(contents, TemplateCondition{
				Content: Div{
					Attr: []Attr{
						{Name: "hx-get", Value: tab.Path(l)},
						{Name: "hx-target", Value: "#" + tt.ID},
						{Name: "hx-push-url", Value: "true"},
						{Name: "hx-trigger", Value: "load"},
					},
				},
			})
		}
	}
	Div{
		ID:      tt.ID,
		Classes: tt.Classes,
		Content: contents,
	}.LoadTemplate(l, w)
}

func (tt TabTarget) LoadMux(l *Location, m *mux.Router) {

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
func (t Tab) Path(l *Location) string {
	return l.Path(t.Value)
}

// Condition returns the template condition that would match this tab.
func (t Tab) Condition(l *Location) string {
	path := t.Path(l)
	return fmt.Sprintf(`matchPath .path "%s/"`, path)
}

// AsCondition converts the tab.Content to a TemplateCondition to make it conditionally render.
func (t Tab) AsCondition(l *Location) TemplateCondition {
	return TemplateCondition{
		Condition: t.Condition(l),
		Content:   At{Location: l, Content: t.Content},
	}
}

// LoadMux attaches all of the Tabs context under the tabs Path.
func (t Tab) LoadMux(l *Location, m *mux.Router) {
	t.Content.LoadMux(l.AtPath(t.Value), m)
}
