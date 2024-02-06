package gohtmx_test

import (
	"errors"
	"testing"

	"github.com/TheWozard/gohtmx"
	"github.com/stretchr/testify/require"
)

var ErrBaseError = errors.New("base error")

func TestPathError_Unwrap(t *testing.T) {
	err := gohtmx.ErrPrependPath(ErrBaseError, "path")
	require.ErrorIs(t, err, ErrBaseError)
}

func TestPathError_Path(t *testing.T) {
	testCases := []struct {
		desc     string
		err      error
		expected error
	}{
		{
			desc:     "single prepend",
			err:      gohtmx.ErrPrependPath(ErrBaseError, "path"),
			expected: gohtmx.PathError{Path: []string{"path"}, Err: ErrBaseError},
		},
		{
			desc:     "multi prepend",
			err:      gohtmx.ErrPrependPath(gohtmx.ErrPrependPath(ErrBaseError, "b"), "a"),
			expected: gohtmx.PathError{Path: []string{"a", "b"}, Err: ErrBaseError},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			require.Equal(t, tC.expected, tC.err)
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
			err:      gohtmx.ErrPrependPath(ErrBaseError, "path"),
			expected: "path base error",
		},
		{
			desc:     "multi prepend",
			err:      gohtmx.ErrPrependPath(gohtmx.ErrPrependPath(ErrBaseError, "b"), "a"),
			expected: "a.b base error",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			require.Equal(t, tC.expected, tC.err.Error())
		})
	}
}
