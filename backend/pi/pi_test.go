package pi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersion_String(t *testing.T) {
	type Case struct {
		version  Version
		expected string
	}

	testCases := []Case{
		{
			version:  Version1,
			expected: "v1",
		},
		{
			version:  Version2,
			expected: "v2",
		},
		{
			version:  VersionUnknown,
			expected: "unknown",
		},
		{
			version:  Version(99),
			expected: "unknown",
		},
	}

	for _, tc := range testCases {
		got := tc.version.String()
		require.Equal(t, tc.expected, got)
	}
}

func TestDetectVersion(t *testing.T) {
	type Case struct {
		name             string
		data             string
		expectedVersion  Version
		expectedHasError bool
	}

	testCases := []Case{
		{
			name:             "v1 explicit",
			data:             `{"interface_version": 1}`,
			expectedVersion:  Version1,
			expectedHasError: false,
		},
		{
			name:             "v1 implicit (0)",
			data:             `{"interface_version": 0}`,
			expectedVersion:  Version1,
			expectedHasError: false,
		},
		{
			name:             "v1 missing field",
			data:             `{"name": "test"}`,
			expectedVersion:  Version1,
			expectedHasError: false,
		},
		{
			name:             "v2",
			data:             `{"interface_version": 2}`,
			expectedVersion:  Version2,
			expectedHasError: false,
		},
		{
			name:             "unknown version",
			data:             `{"interface_version": 99}`,
			expectedVersion:  VersionUnknown,
			expectedHasError: true,
		},
		{
			name:             "invalid json",
			data:             `{invalid}`,
			expectedVersion:  VersionUnknown,
			expectedHasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DetectVersion([]byte(tc.data))
			require.Equal(t, tc.expectedVersion, got)
			require.Equal(t, tc.expectedHasError, err != nil)
		})
	}
}

func TestIsI18nString(t *testing.T) {
	type Case struct {
		input    string
		expected bool
	}

	testCases := []Case{
		{
			input:    "$key",
			expected: true,
		},
		{
			input:    "$",
			expected: false,
		},
		{
			input:    "key",
			expected: false,
		},
		{
			input:    "",
			expected: false,
		},
		{
			input:    " $key",
			expected: false,
		},
	}

	for _, tc := range testCases {
		got := IsI18nString(tc.input)
		require.Equal(t, tc.expected, got)
	}
}

func TestGetI18nKey(t *testing.T) {
	type Case struct {
		input    string
		expected string
	}

	testCases := []Case{
		{
			input:    "$key",
			expected: "key",
		},
		{
			input:    "$",
			expected: "$",
		},
		{
			input:    "key",
			expected: "key",
		},
		{
			input:    "",
			expected: "",
		},
		{
			input:    "$hello_world",
			expected: "hello_world",
		},
	}

	for _, tc := range testCases {
		got := GetI18nKey(tc.input)
		require.Equal(t, tc.expected, got)
	}
}
