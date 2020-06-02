package oci

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHasPrefix(t *testing.T) {
	type testCase struct {
		path     string
		prefix   string
		expected bool
	}
	testCases := []testCase{
		{
			path:     "/foo/bar",
			prefix:   "/foo",
			expected: true,
		},
		{
			path:     "/foo/bar",
			prefix:   "/foo/",
			expected: true,
		},
		{
			path:     "/foo/bar",
			prefix:   "/",
			expected: true,
		},
		{
			path:     "/foo",
			prefix:   "/foo",
			expected: true,
		},
		{
			path:     "/foo/bar",
			prefix:   "/bar",
			expected: false,
		},
		{
			path:     "/foo/bar",
			prefix:   "foo",
			expected: false,
		},
		{
			path:     "/foobar",
			prefix:   "/foo",
			expected: false,
		},
	}
	for i, tc := range testCases {
		actual := hasPrefix(tc.path, tc.prefix)
		require.Equal(t, tc.expected, actual, "#%d: under(%q,%q)", i, tc.path, tc.prefix)
	}
}
