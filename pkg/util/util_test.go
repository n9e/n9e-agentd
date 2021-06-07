package util

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSanitizeMetric(t *testing.T) {
	cases := [][2]string{
		{"abc", "abc"},
		{"a?b", "a_b"},
		{"a.b", "a_b"},
		{"a#b", "a_b"},
		{"a?#b", "a__b"},
	}

	for _, c := range cases {
		got := SanitizeMetric(c[0])
		require.Equal(t, c[1], got)
	}
}

func TestSanitizeTags(t *testing.T) {
	cases := [][2][]string{
		{{"a?c", "a?c:d"}, {"a_c", "a_c:d"}},
	}
	for _, c := range cases {
		got := SanitizeTags(c[0])
		require.Equal(t, c[1], got)
	}
}
