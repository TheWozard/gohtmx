package internal_test

import (
	"errors"
	"testing"

	"github.com/TheWozard/gohtmx/internal"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	testCases := []struct {
		desc  string
		check func(v *internal.Validate)
		err   error
	}{
		{
			desc: "positive",
			check: func(v *internal.Validate) {
				v.Require(true, "error")
			},
			err: nil,
		},
		{
			desc: "negative",
			check: func(v *internal.Validate) {
				v.Require(false, "error")
			},
			err: errors.Join(errors.New("error")),
		},
		{
			desc: "multiple",
			check: func(v *internal.Validate) {
				v.Require(false, "error1")
				v.Require(false, "error2")
			},
			err: errors.Join(errors.New("error1"), errors.New("error2")),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			v := internal.NewValidate()
			tC.check(v)
			assert.Equal(t, tC.err != nil, v.HasError())
			assert.Equal(t, tC.err, v.Error())
		})
	}
}
