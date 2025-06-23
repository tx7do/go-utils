package stringcase

import "testing"

func TestKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HelloWorld", "hello-world"},
		{"helloWorld", "hello-world"},
		{"Hello World", "hello-world"},
		{"hello world!", "hello-world"},
		{"Numbers123Test", "numbers-123-test"},
		{"", ""},
		{"_", ""},
		{"__Hello__World__", "hello-world"},
	}

	for _, test := range tests {
		result := KebabCase(test.input)
		if result != test.expected {
			t.Errorf("KebabCase(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}

func TestUpperKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HelloWorld", "HELLO-WORLD"},
		{"helloWorld", "HELLO-WORLD"},
		{"Hello World", "HELLO-WORLD"},
		{"hello world!", "HELLO-WORLD"},
		{"Numbers123Test", "NUMBERS-123-TEST"},
		{"", ""},
		{"_", ""},
		{"__Hello__World__", "HELLO-WORLD"},
	}

	for _, test := range tests {
		result := UpperKebabCase(test.input)
		if result != test.expected {
			t.Errorf("UpperKebabCase(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}
