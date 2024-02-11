package gohtmx

import "github.com/TheWozard/gohtmx/element"

// MetaScope defines a Component that modifies the current path of the Page.
type MetaScope struct {
	Path    string
	Content Component
}

func (s MetaScope) Init(p *Page) (element.Element, error) {
	return s.Content.Init(p.AtPath(s.Path))
}
