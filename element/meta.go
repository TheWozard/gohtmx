package element

import (
	"io"
)

// OnValidate is a function that is called when the element is validated.
type OnValidate func() error

func (v OnValidate) Render(_ io.Writer) error {
	return nil
}

func (v OnValidate) Validate() error {
	return v()
}

func (v OnValidate) GetTags() []*Tag {
	return []*Tag{}
}
