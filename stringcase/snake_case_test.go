package stringcase

import (
	"testing"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"snake_case", "snake_case"},
		{"CamelCase", "camel_case"},
		{"lowerCamelCase", "lower_camel_case"},
		{"F", "f"},
		{"Foo", "foo"},
		{"FooB", "foo_b"},
		{"FooID", "foo_id"},
		{" FooBar\t", "foo_bar"},
		{"HTTPStatusCode", "http_status_code"},
		{"ParseURL.DoParse", "parse_url_do_parse"},
		{"Convert Space", "convert_space"},
		{"Convert-dash", "convert_dash"},
		{"Skip___MultipleUnderscores", "skip_multiple_underscores"},
		{"Skip   MultipleSpaces", "skip_multiple_spaces"},
		{"Skip---MultipleDashes", "skip_multiple_dashes"},
		{"Hello World", "hello_world"},
		{"Multiple Words Example", "multiple_words_example"},
		{"", ""},
		{"A", "a"},
		{"z", "z"},
		{"Special-Characters_Test", "special_characters_test"},
		{"Numbers123Test", "numbers_123_test"},
		{"Hello World!", "hello_world"},
		{"Test@With#Symbols", "test_with_symbols"},
		{"ComplexCase123!@#", "complex_case_123"},
	}

	for _, test := range tests {
		result := ToSnakeCase(test.input)
		if result != test.expected {
			t.Errorf("ToSnakeCase(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}

func TestUpperSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"snake_case", "SNAKE_CASE"},
		{"CamelCase", "CAMEL_CASE"},
		{"lowerCamelCase", "LOWER_CAMEL_CASE"},
		{"F", "F"},
		{"Foo", "FOO"},
		{"FooB", "FOO_B"},
		{"FooID", "FOO_ID"},
		{" FooBar\t", "FOO_BAR"},
		{"HTTPStatusCode", "HTTP_STATUS_CODE"},
		{"ParseURL.DoParse", "PARSE_URL_DO_PARSE"},
		{"Convert Space", "CONVERT_SPACE"},
		{"Convert-dash", "CONVERT_DASH"},
		{"Skip___MultipleUnderscores", "SKIP_MULTIPLE_UNDERSCORES"},
		{"Skip   MultipleSpaces", "SKIP_MULTIPLE_SPACES"},
		{"Skip---MultipleDashes", "SKIP_MULTIPLE_DASHES"},
		{"Hello World", "HELLO_WORLD"},
		{"Multiple Words Example", "MULTIPLE_WORDS_EXAMPLE"},
		{"", ""},
		{"A", "A"},
		{"z", "Z"},
		{"Special-Characters_Test", "SPECIAL_CHARACTERS_TEST"},
		{"Numbers123Test", "NUMBERS_123_TEST"},
		{"Hello World!", "HELLO_WORLD"},
		{"Test@With#Symbols", "TEST_WITH_SYMBOLS"},
		{"ComplexCase123!@#", "COMPLEX_CASE_123"},
	}

	for _, test := range tests {
		result := UpperSnakeCase(test.input)
		if result != test.expected {
			t.Errorf("UpperSnakeCase(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}
