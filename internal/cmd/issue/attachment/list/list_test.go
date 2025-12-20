package list

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatSize(t *testing.T) {
	cases := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "bytes",
			bytes:    500,
			expected: "500 B",
		},
		{
			name:     "kilobytes",
			bytes:    1024,
			expected: "1.0 KB",
		},
		{
			name:     "kilobytes with decimal",
			bytes:    1536,
			expected: "1.5 KB",
		},
		{
			name:     "megabytes",
			bytes:    1048576,
			expected: "1.0 MB",
		},
		{
			name:     "megabytes with decimal",
			bytes:    5242880,
			expected: "5.0 MB",
		},
		{
			name:     "gigabytes",
			bytes:    1073741824,
			expected: "1.0 GB",
		},
		{
			name:     "zero bytes",
			bytes:    0,
			expected: "0 B",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatSize(tc.bytes)
			assert.Equal(t, tc.expected, result)
		})
	}
}
