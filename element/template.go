package element

import "io"

type TBlock struct {
	Text       string
	IncludeEnd bool
	Element    Element
}

func (t TBlock) Validate() error {
	if t.Element == nil {
		return nil
	}
	return t.Element.Validate()
}

func (t TBlock) Render(w io.Writer) error {
	_, err := w.Write([]byte("{{" + t.Text + "}}"))
	if err != nil {
		return err
	}
	if t.Element != nil {
		err = t.Element.Render(w)
		if err != nil {
			return err
		}
	}
	if t.IncludeEnd {
		_, err = w.Write([]byte(`{{end}}`))
		if err != nil {
			return err
		}
	}
	return nil
}

func (t TBlock) GetTags() []*Tag {
	return t.Element.GetTags()
}
