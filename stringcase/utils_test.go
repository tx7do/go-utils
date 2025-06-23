package stringcase

import (
	"reflect"
	"testing"
)

func TestSplitByNonAlphanumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"hello-world", []string{"hello", "world"}},
		{"hello_world", []string{"hello", "world"}},
		{"hello.world", []string{"hello", "world"}},
		{"hello world", []string{"hello", "world"}},
		{"hello123world", []string{"hello123world"}},
		{"hello123 world", []string{"hello123", "world"}},
		{"hello-world_123", []string{"hello", "world", "123"}},
		{"!hello@world#", []string{"hello", "world"}},
	}

	for _, test := range tests {
		result := SplitByNonAlphanumeric(test.input)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("SplitByNonAlphanumeric(%q) = %v; expected %v", test.input, result, test.expected)
		}
	}
}

func TestSplitAndKeepDelimiters(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"hello-world", []string{"hello", "-", "world"}},
		{"hello_world", []string{"hello", "_", "world"}},
		{"hello.world", []string{"hello", ".", "world"}},
		{"hello world", []string{"hello", " ", "world"}},
		{"hello123world", []string{"hello123world"}},
		{"hello123 world", []string{"hello123", " ", "world"}},
		{"hello-world_123", []string{"hello", "-", "world", "_", "123"}},
		{"!hello@world#", []string{"!", "hello", "@", "world", "#"}},
	}

	for _, test := range tests {
		result := SplitAndKeepDelimiters(test.input)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("SplitAndKeepDelimiters(%q) = %v; expected %v", test.input, result, test.expected)
		}
	}
}
