package components

import (
	"net/http"

	"github.com/TheWozard/gohtmx"
)

type Counter struct {
	Attrs *gohtmx.Attributes
	Min   int
	Max   int
	Step  int
	Count int
}

func (c *Counter) Init(p *gohtmx.Page) (gohtmx.Element, error) {
	increment := gohtmx.NewInteraction("increment").Handle(func(r *http.Request) {
		if c.Step > 0 {
			c.Count += c.Step
		} else {
			c.Count++
		}
		if c.Count > c.Max {
			c.Count = c.Max
		}
	})
	decrement := gohtmx.NewInteraction("decrement").Handle(func(r *http.Request) {
		if c.Step > 0 {
			c.Count -= c.Step
		} else {
			c.Count--
		}
		if c.Count < c.Min {
			c.Count = c.Min
		}
	})

	return gohtmx.Div{
		Attrs: c.Attrs,
		Content: gohtmx.Fragment{
			increment.Trigger(gohtmx.Button{
				Content: gohtmx.Raw("+"),
			}),
			increment.Action().Update(decrement.Action().Update(gohtmx.Div{
				Content: gohtmx.TWith{
					Func: func(r *http.Request) any {
						return c
					},
					Content: gohtmx.Raw("{{.Count}}"),
				},
			})),
			decrement.Trigger(gohtmx.Button{
				Content: gohtmx.Raw("-"),
			}),
		},
	}.Init(p)
}
