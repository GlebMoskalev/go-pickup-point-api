package privacy

import "testing"

func TestMaskEmail(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"user@example.com", "us****@example.com"},
		{"u@example.com", "u****@example.com"},
		{"invalid", "invalid_email"},
		{"", "invalid_email"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := MaskEmail(tc.input)
			if result != tc.expected {
				t.Errorf("MaskEmail(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}
