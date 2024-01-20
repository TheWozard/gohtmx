package internal_test

import (
	"errors"
	"testing"

	"github.com/TheWozard/gohtmx/internal"
	"github.com/stretchr/testify/assert"
)

var ErrBaseError = errors.New("base error")

func TestPathError_Unwrap(t *testing.T) {
	err := internal.ErrPrependPath(ErrBaseError, "path")
	assert.ErrorIs(t, err, ErrBaseError)
}

func TestPathError_Path(t *testing.T) {
	testCases := []struct {
		desc     string
		err      error
		expected error
	}{
		{
			desc:     "nil error unchanged",
			err:      internal.ErrEnclosePath(internal.ErrPrependPath(nil, "path"), "meta"),
			expected: nil,
		},
		{
			desc:     "single prepend",
			err:      internal.ErrPrependPath(ErrBaseError, "path"),
			expected: internal.PathError{Path: []string{"path"}, Err: ErrBaseError},
		},
		{
			desc:     "multi prepend",
			err:      internal.ErrPrependPath(internal.ErrPrependPath(ErrBaseError, "b"), "a"),
			expected: internal.PathError{Path: []string{"a", "b"}, Err: ErrBaseError},
		},
		{
			desc:     "single enclose",
			err:      internal.ErrEnclosePath(ErrBaseError, "meta"),
			expected: internal.PathError{Path: []string{"meta"}, Err: ErrBaseError},
		},
		{
			desc:     "multi enclose",
			err:      internal.ErrEnclosePath(internal.ErrEnclosePath(ErrBaseError, "b"), "a"),
			expected: internal.PathError{Path: []string{"a", "(b)"}, Err: ErrBaseError},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.expected, tC.err)
		})
	}
}

func TestPathError_Error(t *testing.T) {
	testCases := []struct {
		desc     string
		err      error
		expected string
	}{
		{
			desc:     "single prepend",
			err:      internal.ErrPrependPath(ErrBaseError, "path"),
			expected: "path base error",
		},
		{
			desc:     "multi prepend",
			err:      internal.ErrPrependPath(internal.ErrPrependPath(ErrBaseError, "b"), "a"),
			expected: "a.b base error",
		},
		{
			desc:     "single enclose",
			err:      internal.ErrEnclosePath(ErrBaseError, "meta"),
			expected: "meta base error",
		},
		{
			desc:     "multi enclose",
			err:      internal.ErrEnclosePath(internal.ErrEnclosePath(ErrBaseError, "b"), "a"),
			expected: "a(b) base error",
		},
		{
			desc:     "single prepend and enclose",
			err:      internal.ErrEnclosePath(internal.ErrPrependPath(ErrBaseError, "a", "b", "c"), "wrap"),
			expected: "wrap(a.b.c) base error",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.expected, tC.err.Error())
		})
	}
}
