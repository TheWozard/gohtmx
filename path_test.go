package gohtmx_test

import (
	"bytes"
	"testing"

	"github.com/TheWozard/gohtmx"
	"github.com/stretchr/testify/assert"
)

func TestPath_Init(t *testing.T) {
	testCases := []struct {
		desc      string
		path      gohtmx.Path
		framework *gohtmx.Framework
		expected  string
		err       error
	}{
		{
			desc:      "Empty Path",
			path:      gohtmx.Path{},
			framework: gohtmx.NewDefaultFramework(),
			expected:  ``,
			err:       gohtmx.ErrMissingID,
		},
		{
			desc: "Path with Paths",
			path: gohtmx.Path{
				ID: "testID",
				Paths: map[string]gohtmx.Component{
					"test": gohtmx.Raw("content"),
				},
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  `<div id="testID">{{if gohtmx_Path_Init_func1_0 $r}}content{{end}}</div>`,
			err:       nil,
		},
		{
			desc: "Path with Default",
			path: gohtmx.Path{
				ID:          "testID",
				DefaultPath: "default",
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  `<div id="testID"></div>`,
			err:       nil,
		},
		{
			desc: "Path with DefaultPath and Paths",
			path: gohtmx.Path{
				ID:          "testID",
				DefaultPath: "test",
				Paths: map[string]gohtmx.Component{
					"test": gohtmx.Raw("content"),
				},
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  `<div id="testID">{{if gohtmx_Path_Init_func1_0 $r}}content{{else}}<div hx-get="/test" hx-target="#testID" hx-trigger="load"></div>{{end}}</div>`,
			err:       nil,
		},
		{
			desc: "Path with DefaultPath and DefaultComponent",
			path: gohtmx.Path{
				ID:               "testID",
				DefaultPath:      "test",
				DefaultComponent: gohtmx.Raw("default"),
				Paths: map[string]gohtmx.Component{
					"test": gohtmx.Raw("content"),
				},
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  `<div id="testID">{{if gohtmx_Path_Init_func1_0 $r}}content{{else}}<div hx-get="/test" hx-target="#testID" hx-trigger="load"></div>{{end}}</div>`,
			err:       nil,
		},
		{
			desc: "Path with DefaultComponent",
			path: gohtmx.Path{
				ID:               "testID",
				DefaultComponent: gohtmx.Raw("default"),
				Paths: map[string]gohtmx.Component{
					"test": gohtmx.Raw("content"),
				},
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  `<div id="testID">{{if gohtmx_Path_Init_func1_0 $r}}content{{else}}default{{end}}</div>`,
			err:       nil,
		},
		{
			desc: "Path with Attributes, Classes, and Style",
			path: gohtmx.Path{
				ID:      "testID",
				Attrs:   gohtmx.Attributes{}.Value("attr", "value"),
				Classes: []string{"class1", "class2"},
			},
			framework: gohtmx.NewDefaultFramework(),
			expected:  `<div attr="value" id="testID" class="class1 class2"></div>`,
			err:       nil,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// Byte Validation
			w := bytes.NewBuffer(nil)
			err := tC.path.Init(tC.framework, w)
			assert.ErrorIs(t, err, tC.err)
			assert.Equal(t, tC.expected, w.String())
			// Template Validation
			if tC.framework.CanTemplate() {
				_, err = tC.framework.Template.Parse("{{$r = .request}}" + w.String())
				assert.NoError(t, err, "failed to generate valid template")
			}
		})
	}
}
