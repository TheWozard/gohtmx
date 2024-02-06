package gohtmx

import (
	"io"
	"sort"
)

// Attrs creates a new Attributes. This should be used over creating them manually.
func Attrs() *Attributes {
	return &Attributes{Values: map[string][]string{}}
}

// Attributes defines a map of attributes for use in a Tag.
// Attributes are rendered in alphabetical order. Attributes are stored as Elements to allow for templating.
type Attributes struct {
	Values map[string][]string
}

// Get returns the first value of the attribute if it exists.
func (a *Attributes) Get(key string) (string, bool) {
	if a == nil || a.Values == nil {
		return "", false
	}
	values, ok := a.Values[key]
	if !ok || len(values) != 1 {
		return "", false
	}
	return values[0], true
}

// Ensure guarantees that the Attributes are not nil.
func (a *Attributes) Ensure() *Attributes {
	if a == nil || a.Values == nil {
		return Attrs()
	}
	return a
}

// Write writes the attributes to the passed io.Writer.
func (a *Attributes) Write(w io.Writer) error {
	if a.IsEmpty() {
		return nil
	}

	// Attributes are written in sorted order.
	keys := []string{}
	for key := range a.Values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var err error
	write := func(s string) {
		_, e := w.Write([]byte(s))
		if err == nil && e != nil {
			err = e
		}
	}
	for i, key := range keys {
		if i > 0 {
			write(" ")
		}
		write(key)
		data, ok := a.Values[key]
		if !ok || data == nil {
			continue
		}
		write(`="`)
		for i, element := range data {
			if i > 0 {
				write(" ")
			}
			write(element)
		}
		write(`"`)
	}
	return err
}

// Copy returns a shallow copy of the attributes.
func (a *Attributes) Copy() *Attributes {
	a.Ensure()
	n := make(map[string][]string, len(a.Values))
	for key, value := range a.Values {
		n[key] = value
	}
	return &Attributes{Values: n}
}

// IsEmpty returns true if there are no attributes.
func (a *Attributes) IsEmpty() bool {
	return a == nil || a.Values == nil || len(a.Values) == 0
}

func (a *Attributes) Delete(name string) *Attributes {
	a = a.Ensure()
	delete(a.Values, name)
	return a
}

// String adds a named value to the attributes if it is not empty.
func (a *Attributes) String(name string, value string) *Attributes {
	a = a.Ensure()
	if value != "" {
		a.Values[name] = append(a.Values[name], value)
	}
	return a
}

// Slice adds a named list of values to the attributes if it is not empty.
func (a *Attributes) Strings(name string, values ...string) *Attributes {
	a = a.Ensure()
	if len(values) > 0 {
		a.Values[name] = append(a.Values[name], values...)
	}
	return a
}

// Bool adds a named flag to the attributes if active is true.
func (a *Attributes) Bool(name string, active bool) *Attributes {
	a = a.Ensure()
	if active {
		a.Values[name] = nil
	}
	return a
}
