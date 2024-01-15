package gohtmx_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/TheWozard/gohtmx/gohtmx"
	"github.com/TheWozard/gohtmx/gohtmx/core"
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
			expected:  "{{$varName := v2_test_TestTVariable_Init_func1_0 $r}}",
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
				Func: func(r *http.Request) core.TemplateData {
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
				Func: func(r *http.Request) core.TemplateData {
					return nil
				},
				Content: gohtmx.Raw("content"),
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  "{{with v2_test_TestTWith_Init_func2_0 $r}}content{{end}}",
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
