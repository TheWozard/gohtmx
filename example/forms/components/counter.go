package components

import (
	"net/http"
	"strconv"

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
		step := c.Step
		if amount := r.URL.Query().Get("amount"); amount != "" {
			step, _ = strconv.Atoi(amount)
		}
		if c.Step > 0 {
			c.Count += step
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
			increment.Trigger().Target(gohtmx.Button{
				Content: gohtmx.Raw("+"),
			}),
			increment.Trigger().Set("amount", "1").Target(gohtmx.Button{
				Content: gohtmx.Raw("+1"),
			}),
			increment.Trigger().Set("amount", "2").Target(gohtmx.Button{
				Content: gohtmx.Raw("+2"),
			}),
			increment.Trigger().Set("amount", "3").Target(gohtmx.Button{
				Content: gohtmx.Raw("+3"),
			}),
			increment.Swap().Update(decrement.Swap().Update(gohtmx.Div{
				Content: gohtmx.TWith{
					Func: func(r *http.Request) any {
						return c
					},
					Content: gohtmx.Raw("{{.Count}}"),
				},
			})),
			decrement.Trigger().Target(gohtmx.Button{
				Content: gohtmx.Raw("-"),
			}),
		},
	}.Init(p)
}
