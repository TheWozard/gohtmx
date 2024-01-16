package gohtmx_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/TheWozard/gohtmx"
	"github.com/stretchr/testify/assert"
)

func TestTAction_Init(t *testing.T) {
	testCases := []struct {
		desc      string
		action    gohtmx.TAction
		framework *gohtmx.Framework
		expected  string
		err       error
	}{
		{
			desc:      "Empty Action",
			action:    "",
			framework: gohtmx.NewDefaultFramework(),
			expected:  "",
			err:       gohtmx.ErrMissingAction,
		},
		{
			desc:      "Action with Content",
			action:    "content",
			framework: gohtmx.NewDefaultFramework(),
			expected:  "{{content}}",
			err:       nil,
		},
		{
			desc:      "Cannot Template",
			action:    "content",
			framework: &gohtmx.Framework{},
			expected:  "",
			err:       gohtmx.ErrCannotTemplate,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// Byte Validation
			w := bytes.NewBuffer(nil)
			err := tC.action.Init(tC.framework, w)
			assert.Equal(t, tC.err, err)
			assert.Equal(t, tC.expected, w.String())
		})
	}
}

func TestTBlock_Init(t *testing.T) {
	testCases := []struct {
		desc      string
		block     gohtmx.TBlock
		framework *gohtmx.Framework
		expected  string
		err       error
	}{
		{
			desc:      "Empty Block",
			block:     gohtmx.TBlock{},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "",
			err:       gohtmx.ErrMissingAction,
		},
		{
			desc: "Block with Content",
			block: gohtmx.TBlock{
				Action:  "action",
				Content: gohtmx.Raw("content"),
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "{{action}}content{{end}}",
			err:       nil,
		},
		{
			desc: "Cannot Template",
			block: gohtmx.TBlock{
				Action:  "action",
				Content: gohtmx.Raw("content"),
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
			err := tC.block.Init(tC.framework, w)
			assert.Equal(t, tC.err, err)
			assert.Equal(t, tC.expected, w.String())
		})
	}
}

func TestTBlocks_Init(t *testing.T) {
	testCases := []struct {
		desc      string
		blocks    gohtmx.TBlocks
		framework *gohtmx.Framework
		expected  string
		err       error
	}{
		{
			desc:      "Empty Blocks",
			blocks:    gohtmx.TBlocks{},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "",
			err:       nil,
		},
		{
			desc: "Blocks with Content",
			blocks: gohtmx.TBlocks{
				{
					Action:  "action1",
					Content: gohtmx.Raw("content1"),
				},
				{
					Action:  "action2",
					Content: gohtmx.Raw("content2"),
				},
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "{{action1}}content1{{action2}}content2{{end}}",
			err:       nil,
		},
		{
			desc: "Cannot Template",
			blocks: gohtmx.TBlocks{
				{
					Action:  "action",
					Content: gohtmx.Raw("content"),
				},
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
			err := tC.blocks.Init(tC.framework, w)
			assert.Equal(t, tC.err, err)
			assert.Equal(t, tC.expected, w.String())
		})
	}
}

func TestTVariable_Init(t *testing.T) {
	testCases := []struct {
		desc      string
		variable  gohtmx.TVariable
		framework *gohtmx.Framework
		expected  string
		err       error
	}{
		{
			desc: "Empty Variable",
			variable: gohtmx.TVariable{
				Name: "",
				Func: nil,
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "",
			err:       gohtmx.ErrInvalidVariableName,
		},
		{
			desc: "Empty Func",
			variable: gohtmx.TVariable{
				Name: "$varName",
				Func: nil,
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "",
			err:       gohtmx.ErrMissingFunction,
		},
		{
			desc: "Variable with Content",
			variable: gohtmx.TVariable{
				Name: "$varName",
				Func: func(r *http.Request) any {
					return nil
				},
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "{{$varName := gohtmx_test_TestTVariable_Init_func1_0 $r}}",
			err:       nil,
		},
		{
			desc: "Cannot Template",
			variable: gohtmx.TVariable{
				Name: "varName",
				Func: func(r *http.Request) any {
					return nil
				},
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
			err := tC.variable.Init(tC.framework, w)
			assert.Equal(t, tC.err, err)
			assert.Equal(t, tC.expected, w.String())
			// Template Validation
			if tC.framework.CanTemplate() {
				_, err = tC.framework.Template.Parse("{{$r := nil}}" + w.String())
				assert.NoError(t, err, "failed to generate valid template")
			}
		})
	}
}

func TestTWith_Init(t *testing.T) {
	testCases := []struct {
		desc      string
		with      gohtmx.TWith
		framework *gohtmx.Framework
		expected  string
		err       error
	}{
		{
			desc: "Cannot Template",
			with: gohtmx.TWith{
				Func: func(r *http.Request) gohtmx.Data {
					return nil
				},
				Content: gohtmx.Raw("content"),
			},
			framework: &gohtmx.Framework{},
			expected:  "",
			err:       gohtmx.ErrCannotTemplate,
		},
		{
			desc: "Missing Function",
			with: gohtmx.TWith{
				Func:    nil,
				Content: gohtmx.Raw("content"),
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "",
			err:       gohtmx.ErrMissingFunction,
		},
		{
			desc: "With Content",
			with: gohtmx.TWith{
				Func: func(r *http.Request) gohtmx.Data {
					return nil
				},
				Content: gohtmx.Raw("content"),
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "{{with gohtmx_test_TestTWith_Init_func2_0 $r}}content{{end}}",
			err:       nil,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// Byte Validation
			w := bytes.NewBuffer(nil)
			err := tC.with.Init(tC.framework, w)
			assert.Equal(t, tC.err, err)
			assert.Equal(t, tC.expected, w.String())
			// Template Validation
			if tC.framework.CanTemplate() {
				_, err = tC.framework.Template.Parse("{{$r := nil}}" + w.String())
				assert.NoError(t, err, "failed to generate valid template")
			}
		})
	}
}

func TestTCondition_Init(t *testing.T) {
	testCases := []struct {
		desc      string
		condition gohtmx.TCondition
		framework *gohtmx.Framework
		expected  string
		err       error
	}{
		{
			desc: "Cannot Template",
			condition: gohtmx.TCondition{
				Condition: func(r *http.Request) bool {
					return true
				},
				Content: gohtmx.Raw("content"),
			},
			framework: &gohtmx.Framework{},
			expected:  "",
			err:       gohtmx.ErrCannotTemplate,
		},
		{
			desc: "No Condition",
			condition: gohtmx.TCondition{
				Condition: nil,
				Content:   gohtmx.Raw("content"),
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "content",
			err:       nil,
		},
		{
			desc: "With Condition",
			condition: gohtmx.TCondition{
				Condition: func(r *http.Request) bool {
					return true
				},
				Content: gohtmx.Raw("content"),
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "{{if gohtmx_test_TestTCondition_Init_func2_0 $r}}content{{end}}",
			err:       nil,
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
				_, err = tC.framework.Template.Parse("{{$r := nil}}" + w.String())
				assert.NoError(t, err, "failed to generate valid template")
			}
		})
	}
}

func TestTConditions_Init(t *testing.T) {
	testCases := []struct {
		desc       string
		conditions gohtmx.TConditions
		framework  *gohtmx.Framework
		expected   string
		err        error
	}{
		{
			desc: "Cannot Template",
			conditions: gohtmx.TConditions{
				{
					Condition: func(r *http.Request) bool {
						return true
					},
					Content: gohtmx.Raw("content"),
				},
			},
			framework: &gohtmx.Framework{},
			expected:  "",
			err:       gohtmx.ErrCannotTemplate,
		},
		{
			desc: "No Conditions",
			conditions: gohtmx.TConditions{
				{
					Condition: nil,
					Content:   gohtmx.Raw("content"),
				},
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "content",
			err:       nil,
		},
		{
			desc: "With Conditions",
			conditions: gohtmx.TConditions{
				{
					Condition: func(r *http.Request) bool {
						return true
					},
					Content: gohtmx.Raw("content1"),
				},
				{
					Condition: nil,
					Content:   gohtmx.Raw("content2"),
				},
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "{{if gohtmx_test_TestTConditions_Init_func2_0 $r}}content1{{else}}content2{{end}}",
			err:       nil,
		},
		{
			desc: "Many Conditions",
			conditions: gohtmx.TConditions{
				{
					Condition: func(r *http.Request) bool {
						return true
					},
					Content: gohtmx.Raw("content1"),
				},
				{
					Condition: func(r *http.Request) bool {
						return true
					},
					Content: gohtmx.Raw("content2"),
				},
				{
					Condition: func(r *http.Request) bool {
						return true
					},
					Content: gohtmx.Raw("content3"),
				},
				{
					Condition: func(r *http.Request) bool {
						return true
					},
					Content: gohtmx.Raw("content4"),
				},
				{
					Condition: nil,
					Content:   gohtmx.Raw("content5"),
				},
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "{{if gohtmx_test_TestTConditions_Init_func3_0 $r}}content1{{else if gohtmx_test_TestTConditions_Init_func4_0 $r}}content2{{else if gohtmx_test_TestTConditions_Init_func5_0 $r}}content3{{else if gohtmx_test_TestTConditions_Init_func6_0 $r}}content4{{else}}content5{{end}}",
			err:       nil,
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
				_, err = tC.framework.Template.Parse("{{$r := nil}}" + w.String())
				assert.NoError(t, err, "failed to generate valid template")
			}
		})
	}
}
