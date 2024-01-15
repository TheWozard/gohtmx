package core_test

import (
	"testing"

	"github.com/TheWozard/gohtmx/gohtmx/core"
	"github.com/stretchr/testify/assert"
)

func TestMerge(t *testing.T) {
	testCases := []struct {
		desc     string
		tData    core.TemplateData
		addition core.TemplateData
		expected core.TemplateData
	}{
		{
			desc:     "Empty TemplateData",
			tData:    core.TemplateData{},
			addition: core.TemplateData{"key1": "value1"},
			expected: core.TemplateData{"key1": "value1"},
		},
		{
			desc:     "Non-empty TemplateData, no overlap",
			tData:    core.TemplateData{"key1": "value1"},
			addition: core.TemplateData{"key2": "value2"},
			expected: core.TemplateData{"key1": "value1", "key2": "value2"},
		},
		{
			desc:     "Non-empty TemplateData, with overlap",
			tData:    core.TemplateData{"key1": "value1"},
			addition: core.TemplateData{"key1": "newvalue1", "key2": "value2"},
			expected: core.TemplateData{"key1": "newvalue1", "key2": "value2"},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			result := tC.tData.Merge(tC.addition)
			assert.Equal(t, tC.expected, result)
		})
	}
}
