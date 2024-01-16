package gohtmx

import (
	"errors"
	"fmt"
)

var ErrNilComponent = fmt.Errorf("component cannot be nil")
var ErrMissingContent = fmt.Errorf("missing component")
var ErrMissingID = fmt.Errorf("missing id")
var ErrMissingTarget = fmt.Errorf("missing target")

func NewValidate() *Validate {
	return &Validate{}
}

type Validate struct {
	errs []error
}

func (v *Validate) HasError() bool {
	return len(v.errs) > 0
}

func (v *Validate) Error() error {
	return errors.Join(v.errs...)
}

func (v *Validate) Require(check bool, err string) bool {
	if !check {
		v.errs = append(v.errs, errors.New(err))
		return false
	}
	return true
}

func (v *Validate) RequireID(id string) bool {
	if id == "" {
		v.errs = append(v.errs, ErrMissingID)
		return false
	}
	return true
}

func (v *Validate) RequireTarget(target string) bool {
	if target == "" {
		v.errs = append(v.errs, ErrMissingTarget)
		return false
	}
	return true
}

func (v *Validate) RequireComponent(name string, c Component) bool {
	if c == nil {
		v.errs = append(v.errs, fmt.Errorf("failure with component %s: %w", name, ErrNilComponent))
		return false
	}
	return true
}

func (v *Validate) RequireFunction(name string, f any) bool {
	if f == nil {
		v.errs = append(v.errs, fmt.Errorf("failure with function %s: %w", name, ErrMissingFunction))
		return false
	}
	return true
}

func (v *Validate) RequireTemplateHandler(name string, f *Framework, c Component) *TemplateHandler {
	if !v.RequireComponent(name, c) {
		return nil
	}
	handler, err := NewTemplateHandler(f, c)
	if err != nil {
		v.errs = append(v.errs, fmt.Errorf("failure with component %s: %w", name, err))
		return nil
	}
	return handler
}
