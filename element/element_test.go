package element_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/TheWozard/gohtmx/attributes"
	"github.com/TheWozard/gohtmx/element"
	"github.com/stretchr/testify/require"
)

func TestElements(t *testing.T) {
	testCases := []struct {
		desc        string
		element     element.Element
		validateErr error
		expected    string
		tags        []*element.Tag
	}{
		{
			desc:     "empty fragment",
			element:  element.Fragment{},
			expected: "",
			tags:     []*element.Tag{},
		},
		{
			desc:     "raw",
			element:  element.Raw(`some text`),
			expected: `some text`,
			tags:     []*element.Tag{},
		},
		{
			desc:        "raw error",
			element:     element.RawError{Err: errors.New(`some error`)},
			validateErr: errors.New(`some error`),
			expected:    `some error`,
			tags:        []*element.Tag{},
		},
		{
			desc: "fragment with raw",
			element: element.Fragment{
				element.Raw(`some text`),
			},
			expected: `some text`,
			tags:     []*element.Tag{},
		},
		{
			desc: "fragment with raw error",
			element: element.Fragment{
				element.RawError{Err: errors.New(`some error`)},
			},
			validateErr: errors.Join(errors.New(`some error`)),
			expected:    `some error`,
			tags:        []*element.Tag{},
		},
		{
			desc: "fragment with multiple raw errors",
			element: element.Fragment{
				element.RawError{Err: errors.New(`some error`)},
				element.RawError{Err: errors.New(`another error`)},
			},
			validateErr: errors.Join(errors.Join(errors.New(`some error`)), errors.New(`another error`)),
			expected:    `some erroranother error`,
			tags:        []*element.Tag{},
		},
		{
			desc:     "tag",
			element:  &element.Tag{Name: "div"},
			expected: `<div></div>`,
			tags:     []*element.Tag{{Name: "div"}},
		},
		{
			desc: "tag with attributes",
			element: &element.Tag{
				Name:       "div",
				Attributes: attributes.New().String("id", "test").Strings("class", "test"),
			},
			expected: `<div class="test" id="test"></div>`,
			tags:     []*element.Tag{{Name: "div", Attributes: attributes.New().String("id", "test").Strings("class", "test")}},
		},
		{
			desc: "tag with content",
			element: &element.Tag{
				Name:    "div",
				Content: element.Raw(`some text`),
			},
			expected: `<div>some text</div>`,
			tags:     []*element.Tag{{Name: "div", Content: element.Raw(`some text`)}},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			require.Equal(t, tC.validateErr, tC.element.Validate())
			data := bytes.NewBuffer(nil)
			require.NoError(t, tC.element.Render(data))
			require.Equal(t, tC.expected, data.String())
			require.Equal(t, tC.tags, tC.element.GetTags())
		})
	}
}
