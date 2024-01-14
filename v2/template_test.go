package gohtmx_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/TheWozard/gohtmx/v2"
	"github.com/stretchr/testify/assert"
)

func condition(r *http.Request) bool { return true }

func TestCondition_Init(t *testing.T) {
	testCases := []struct {
		desc      string
		condition gohtmx.Condition
		framework *gohtmx.Framework
		expected  string
		err       error
	}{
		{
			desc:      "Empty Condition",
			condition: gohtmx.Condition{},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "",
			err:       nil,
		},
		{
			desc: "Condition with Content",
			condition: gohtmx.Condition{
				Condition: condition,
				Content:   gohtmx.Raw("content"),
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "{{if v2_test_condition_0}}content{{end}}",
			err:       nil,
		},
		{
			desc: "Content without Condition",
			condition: gohtmx.Condition{
				Content: gohtmx.Raw("content"),
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "content",
			err:       nil,
		},
		{
			desc: "Cannot Template",
			condition: gohtmx.Condition{
				Condition: condition,
				Content:   gohtmx.Raw("content"),
			},
			framework: &gohtmx.Framework{},
			expected:  "",
			err:       gohtmx.ErrCannotTemplate,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// Byte Validation
			w := bytes.NewBuffer(nil)
			err := tC.condition.Init(tC.framework, w)
			assert.Equal(t, tC.err, err)
			assert.Equal(t, tC.expected, w.String())
			// Template Validation
			if tC.framework.CanTemplate() {
				_, err = tC.framework.Template.Parse(w.String())
				assert.NoError(t, err, "failed to generate valid template")
			}
		})
	}
}

func TestConditions_Init(t *testing.T) {

	testCases := []struct {
		desc       string
		conditions gohtmx.Conditions
		framework  *gohtmx.Framework
		expected   string
		err        error
	}{
		{
			desc:       "Empty Conditions",
			conditions: gohtmx.Conditions{},
			framework:  gohtmx.NewDefaultFramework(),
			expected:   "",
			err:        nil,
		},
		{
			desc: "Single Condition",
			conditions: gohtmx.Conditions{
				{Condition: condition, Content: gohtmx.Raw("content")},
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "{{if v2_test_condition_0}}content{{end}}",
			err:       nil,
		},
		{
			desc: "Multiple Conditions",
			conditions: gohtmx.Conditions{
				{Condition: condition, Content: gohtmx.Raw("content1")},
				{Condition: condition, Content: gohtmx.Raw("content2")},
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "{{if v2_test_condition_0}}content1{{else if v2_test_condition_1}}content2{{end}}",
			err:       nil,
		},
		{
			desc: "Else",
			conditions: gohtmx.Conditions{
				{Content: gohtmx.Raw("content1")},
				{Content: gohtmx.Raw("content2")},
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "content1content2",
			err:       nil,
		},
		{
			desc: "Multiple Conditions With Else",
			conditions: gohtmx.Conditions{
				{Condition: condition, Content: gohtmx.Raw("content1")},
				{Condition: condition, Content: gohtmx.Raw("content2")},
				{Content: gohtmx.Raw("content3")},
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "{{if v2_test_condition_0}}content1{{else if v2_test_condition_1}}content2{{else}}content3{{end}}",
			err:       nil,
		},
		{
			desc: "Cannot Template",
			conditions: gohtmx.Conditions{
				{Condition: condition, Content: gohtmx.Raw("content")},
			},
			framework: &gohtmx.Framework{},
			expected:  "",
			err:       gohtmx.ErrCannotTemplate,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// Byte Validation
			w := bytes.NewBuffer(nil)
			err := tC.conditions.Init(tC.framework, w)
			assert.Equal(t, tC.err, err)
			assert.Equal(t, tC.expected, w.String())
			// Template Validation
			if tC.framework.CanTemplate() {
				_, err = tC.framework.Template.Parse(w.String())
				assert.NoError(t, err, "failed to generate valid template")
			}
		})
	}
}
