package gohtmx_test

import (
	"errors"
	"testing"

	"github.com/TheWozard/gohtmx"
	"github.com/stretchr/testify/require"
)

type PageNonAPITestCase struct {
	desc           string
	setup          func(p *gohtmx.Page)
	validationErrs map[string]error
	rendered       map[string]string
	renderErr      error
}

func (tC PageNonAPITestCase) Assert(t *testing.T) {
	t.Run(tC.desc, func(t *testing.T) {
		p := gohtmx.NewPage()
		tC.setup(p)
		require.Equal(t, tC.validationErrs, p.Validate())
		rendered, err := p.Render()
		require.Equal(t, tC.renderErr, err)
		require.Equal(t, tC.rendered, rendered)
	})
}

func TestPage(t *testing.T) {
	testCases := []PageNonAPITestCase{
		{
			desc: "single add",
			setup: func(p *gohtmx.Page) {
				p.Add(gohtmx.Raw("test"))
			},
			rendered: map[string]string{
				"/": `{{$r := .request}}test`,
			},
		},
		{
			desc: "multiple adds",
			setup: func(p *gohtmx.Page) {
				p.Add(gohtmx.Raw("test1"))
				p.Add(gohtmx.Raw("test2"))
			},
			rendered: map[string]string{
				"/": `{{$r := .request}}test1test2`,
			},
		},
		{
			desc: "single add with path",
			setup: func(p *gohtmx.Page) {
				p.AtPath("example").Add(gohtmx.Raw("test"))
			},
			rendered: map[string]string{
				"/example": `{{$r := .request}}test`,
			},
		},
		{
			desc: "error in validation",
			setup: func(p *gohtmx.Page) {
				p.Add(gohtmx.RawError{Err: errors.New("test error")})
			},
			validationErrs: map[string]error{
				"/": errors.Join(errors.New("test error")),
			},
			rendered: map[string]string{
				"/": `{{$r := .request}}test error`,
			},
		},
	}
	for _, tC := range testCases {
		tC.Assert(t)
	}
}
