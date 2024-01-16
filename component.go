package gohtmx

import (
	"fmt"
	"io"
)

// Component defines the interface for initializing a component. Components are initialized only once.
// A Component can use the Framework to add interactive functionality, and will write its initial data out to the io.Writer.
// Commonly the writer is not written to directly and instead use other Components or a Tag struct to continue the Init.
type Component interface {
	Init(f *Framework, w io.Writer) error
}

// ComponentFunc defines a Component as a function.
type ComponentFunc func(f *Framework, w io.Writer) error

func (c ComponentFunc) Init(f *Framework, w io.Writer) error {
	return c(f, w)
}

// Fragment defines a slice of Components that can be used as a single Component.
type Fragment []Component

func (fr Fragment) Init(f *Framework, w io.Writer) error {
	for i, fragment := range fr {
		if fragment != nil {
			err := fragment.Init(f, w)
			if err != nil {
				return AddPathToError(err, fmt.Sprintf("Fragment[%d]", i))
			}
		}
	}
	return nil
}
