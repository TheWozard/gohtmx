package internal

import (
	"errors"
)

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
