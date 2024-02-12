package gohtmx_test

import (
	"testing"

	"github.com/TheWozard/gohtmx"
)

func TestInteraction(t *testing.T) {
	testCases := []PageNonAPITestCase{
		{
			desc: "raw component",
			setup: func(p *gohtmx.Page) {
				interaction := gohtmx.NewInteraction("interaction")
				p.Add(gohtmx.Fragment{
					interaction,
					interaction.Swap().Update(gohtmx.Div{
						Content: gohtmx.Raw("test"),
					}),
					interaction.Trigger().Target(gohtmx.Button{
						Content: gohtmx.Raw("update"),
					}),
				})
			},
			rendered: map[string]string{
				"/": `{{$r := .request}}` +
					`<div id="gohtmx_0">test</div>` +
					`<button hx-post="/interaction" hx-swap="outerHTML" hx-target="#gohtmx_0" type="button">update</button>`,
				"/interaction": `{{$r := .request}}` +
					`<div id="gohtmx_0">test</div>`,
			},
		},
	}
	for _, tC := range testCases {
		tC.Assert(t)
	}
}
