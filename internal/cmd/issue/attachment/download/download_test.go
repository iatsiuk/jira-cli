package download

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeFilename(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal filename",
			input:    "screenshot.png",
			expected: "screenshot.png",
		},
		{
			name:     "path traversal attempt",
			input:    "../../../etc/passwd",
			expected: "passwd",
		},
		{
			name:     "backslashes in filename (sanitized)",
			input:    "..\\..\\etc\\passwd",
			expected: ".._.._etc_passwd",
		},
		{
			name:     "absolute path",
			input:    "/etc/passwd",
			expected: "passwd",
		},
		{
			name:     "filename with directory",
			input:    "some/dir/file.txt",
			expected: "file.txt",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "attachment",
		},
		{
			name:     "dot only",
			input:    ".",
			expected: "attachment",
		},
		{
			name:     "double dot",
			input:    "..",
			expected: "attachment",
		},
		{
			name:     "filename with spaces",
			input:    "my file.pdf",
			expected: "my file.pdf",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizeFilename(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
