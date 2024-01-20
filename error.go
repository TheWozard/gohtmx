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

func AddMetaPathToError(err error, meta string) error {
	if err == nil {
		return nil
	}
	if pe, ok := err.(PathError); ok {
		return pe.Enclose(meta)
	}
	return PathError{
		path: []string{meta},
		err:  err,
	}
}

func AddOrUpgradePathInError(err error, path string) error {
	if err == nil {
		return nil
	}
	if pe, ok := err.(PathError); ok && len(pe.path) > 0 {
		return pe.Upgrade(path)
	}
	return PathError{
		path: []string{path},
		err:  err,
	}
}

// PathError is a custom error for attaching a path to an error.
type PathError struct {
	path []string
	err  error
}

func (p PathError) Error() string {
	return p.Path() + " " + p.err.Error()
}

func (p PathError) Unwrap() error {
	return p.err
}

func (p PathError) Path() string {
	var builder strings.Builder
	for i, segment := range p.path {
		if i > 0 && !strings.HasPrefix(segment, "(") && !strings.HasSuffix(segment, ")") {
			builder.WriteString(".")
		}
		builder.WriteString(segment)
	}
	return builder.String()
}

func (p PathError) Prepend(path ...string) error {
	return PathError{
		path: append(path, p.path...),
		err:  p.err,
	}
}

func (p PathError) Enclose(path string) error {
	return PathError{
		path: []string{path, "(" + p.Path() + ")"},
		err:  p.err,
	}
}

func (p PathError) Upgrade(path string) error {
	return PathError{
		path: append([]string{path}, p.path[1:]...),
		err:  p.err,
	}
}
