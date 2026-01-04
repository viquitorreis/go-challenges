package challenges

import "testing"

func TestReverseString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"single character", "a", "a"},
		{"simple word", "hello", "olleh"},
		{"with spaces", "hello world", "dlrow olleh"},
		{"with unicode", "hello 世界", "界世 olleh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReverseString(tt.input)
			if result != tt.expected {
				t.Errorf("ReverseString(%q) = %q; expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty string", "", true},
		{"single character", "a", true},
		{"palindrome odd length", "racecar", true},
		{"palindrome even length", "abba", true},
		{"not palindrome", "hello", false},
		{"palindrome with unicode", "上中上", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPalindrome(tt.input)
			if result != tt.expected {
				t.Errorf("IsPalindrome(%q) = %v; expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
