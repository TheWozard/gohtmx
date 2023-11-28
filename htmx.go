package gohtmx

const (
	// SwapContent swaps the Content of a Component with the new Content.
	SwapContent = Swap("innerHTML")
	// SwapComponent swaps a Component with the new Content.
	SwapComponent = Swap("outerHTML")
	// SwapAppend adds the new Component to the end of the current Content.
	SwapAppend = Swap("beforeend")
	// SwapPrepend adds the new Component before the current Content.
	SwapPrepend = Swap("afterbegin")
	// SwapNone does not swap any Content. Out of band swaps are still handled.
	SwapNone = Swap("none")

	SwapDelete = Swap("delete")
)

type Swap string

func (s Swap) OrDefault(d Swap) Swap {
	if s == "" {
		return d
	}
	return s
}
