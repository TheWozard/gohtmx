package gohtmx

import "strings"

// AddPathToError either adds to an existing PathError or creates a new one.
func AddPathToError(err error, path ...string) error {
	if err == nil {
		return nil
	}
	if pe, ok := err.(PathError); ok {
		return pe.Prepend(path...)
	}
	return PathError{
		path: path,
		err:  err,
	}
}

// PathError is a custom error for attaching a path to an error.
type PathError struct {
	path []string
	err  error
}

func (p PathError) Error() string {
	return strings.Join(p.path, ".") + p.err.Error()
}

func (p PathError) Unwrap() error {
	return p.err
}

func (p PathError) Prepend(path ...string) error {
	return PathError{
		path: append(path, p.path...),
		err:  p.err,
	}
}
