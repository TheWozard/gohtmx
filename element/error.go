package element

import (
	"errors"
	"strings"
)

// ErrPrependPath either adds the path to an existing PathError or creates a new one.
func ErrPrependPath(err error, path ...string) error {
	if err == nil {
		return nil
	}
	var pe PathError
	if errors.As(err, &pe) {
		return pe.Prepend(path...)
	}
	return PathError{Path: path, Err: err}
}

// PathError is a custom error for attaching a path to an error.
type PathError struct {
	Path []string
	Err  error
}

func (p PathError) Error() string {
	return p.String() + " " + p.Err.Error()
}

func (p PathError) Unwrap() error {
	return p.Err
}

func (p PathError) String() string {
	var builder strings.Builder
	for i, segment := range p.Path {
		if i > 0 && !strings.HasPrefix(segment, "(") && !strings.HasSuffix(segment, ")") {
			builder.WriteString(".")
		}
		builder.WriteString(segment)
	}
	return builder.String()
}

func (p PathError) Prepend(path ...string) PathError {
	return PathError{
		Path: append(path, p.Path...),
		Err:  p.Err,
	}
}
