package core_test

import (
	"fmt"
	"html/template"
	"testing"

	"github.com/TheWozard/gohtmx/gohtmx/core"
	"github.com/stretchr/testify/assert"
)

func NamedFunc() string { return "" }

type StructFunc struct{}

func (t StructFunc) Func() string { return "" }

func TestDefaultNewFunctionID(t *testing.T) {
	iter := core.NewDefaultGenerator()

	testCases := []struct {
		desc     string
		function any
		output   string
	}{
		{
			desc:     "inline",
			function: func() string { return "" },
			output:   "core_test_TestDefaultNewFunctionID_func1_0",
		},
		{
			desc:     "multiple_inline",
			function: func() string { return "" },
			output:   "core_test_TestDefaultNewFunctionID_func2_0",
		},
		{
			desc:     "named",
			function: NamedFunc,
			output:   "core_test_NamedFunc_0",
		},
		{
			desc:     "repeat",
			function: NamedFunc,
			output:   "core_test_NamedFunc_1",
		},
		{
			desc:     "struct",
			function: StructFunc{}.Func,
			output:   "core_test_StructFunc_Func_fm_0",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// Test Expectations
			id := iter.NewFunctionID(tC.function)
			assert.Equal(t, tC.output, id)

			// Test it can be used as a identifier in a template for a function.
			_, err := template.New("test").Funcs(template.FuncMap{
				id: tC.function,
			}).Parse(fmt.Sprintf("{{%s}}", id))
			assert.Nil(t, err)
		})
	}
}
