package components

import (
	"fmt"

	"github.com/TheWozard/gohtmx"
)

type Counter struct {
	Attrs *gohtmx.Attributes
	Count int
}

func (c *Counter) Component() gohtmx.Component {
	increment := gohtmx.NewInteraction("increment")
	decrement := gohtmx.NewInteraction("decrement")

	return gohtmx.Div{
		Attrs: c.Attrs,
		Content: gohtmx.Fragment{
			increment.Trigger(gohtmx.Button{
				Content: gohtmx.Raw("+"),
			}),
			increment.Update(decrement.Update(gohtmx.Div{
				// TODO: We need With in order to load the value per request.
				Content: gohtmx.Raw(fmt.Sprintf("%d", c.Count)),
			})),
			decrement.Trigger(gohtmx.Button{
				Content: gohtmx.Raw("-"),
			}),
		},
	}
}
