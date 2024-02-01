package gohtmx_test

import (
	"bytes"
	"testing"

	"github.com/TheWozard/gohtmx"
	"github.com/stretchr/testify/assert"
)

func TestAttributes_Get(t *testing.T) {
	testCases := []struct {
		desc     string
		attrs    *gohtmx.Attributes
		key      string
		expected string
		ok       bool
	}{
		{
			desc:     "nil attributes",
			attrs:    nil,
			key:      "key",
			expected: "",
			ok:       false,
		},
		{
			desc:     "empty attributes",
			attrs:    gohtmx.Attrs(),
			key:      "key",
			expected: "",
			ok:       false,
		},
		{
			desc:     "non-existing key",
			attrs:    gohtmx.Attrs().String("key", "value"),
			key:      "nonExistingKey",
			expected: "",
			ok:       false,
		},
		{
			desc:     "existing key with single value",
			attrs:    gohtmx.Attrs().String("key", "value"),
			key:      "key",
			expected: "value",
			ok:       true,
		},
		{
			desc:     "existing key with multiple values",
			attrs:    gohtmx.Attrs().Strings("key", "value1", "value2"),
			key:      "key",
			expected: "",
			ok:       false,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			value, ok := tC.attrs.Get(tC.key)
			assert.Equal(t, tC.ok, ok)
			assert.Equal(t, tC.expected, value)
		})
	}
}

func TestAttributes_Render(t *testing.T) {
	testCases := []struct {
		desc     string
		attrs    *gohtmx.Attributes
		expected string
	}{
		{
			desc:     "empty attributes",
			attrs:    gohtmx.Attrs(),
			expected: ``,
		},
		{
			desc:     "string attribute",
			attrs:    gohtmx.Attrs().String("key", "value"),
			expected: `key="value"`,
		},
		{
			desc:     "slice attribute",
			attrs:    gohtmx.Attrs().Strings("key", "a", "b", "c"),
			expected: `key="a b c"`,
		},
		{
			desc:     "bool attribute",
			attrs:    gohtmx.Attrs().Bool("key", true),
			expected: `key`,
		},
		{
			desc:     "multiple attributes",
			attrs:    gohtmx.Attrs().String("keyA", "A").String("keyB", "B").Strings("keyC", "C", "D").Bool("keyD", true),
			expected: `keyA="A" keyB="B" keyC="C D" keyD`,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			w := bytes.NewBuffer(nil)
			assert.Nil(t, tC.attrs.Render(w))
			assert.Equal(t, tC.expected, w.String())
		})
	}
}
