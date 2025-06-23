package stringcase

import (
	"testing"
)

func TestUpperCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "HelloWorld"},
		{"hello_world", "HelloWorld"},
		{"hello-world", "HelloWorld"},
		{"hello.world", "HelloWorld"},
		{"helloWorld", "HelloWorld"},
		{"HelloWorld", "HelloWorld"},
		{"HTTPStatusCode", "HttpStatusCode"},
		{"ParseURL.DoParse", "ParseUrlDoParse"},
		{"ParseUrl.DoParse", "ParseUrlDoParse"},
		{"parse_url.do_parse", "ParseUrlDoParse"},
		{"convert space", "ConvertSpace"},
		{"convert-dash", "ConvertDash"},
		{"skip___multiple_underscores", "SkipMultipleUnderscores"},
		{"skip   multiple spaces", "SkipMultipleSpaces"},
		{"skip---multiple-dashes", "SkipMultipleDashes"},
		{"", ""},
		{"a", "A"},
		{"Z", "Z"},
		{"special-characters_test", "SpecialCharactersTest"},
		{"numbers123test", "Numbers123Test"},
		{"hello world!", "HelloWorld"},
		{"test@with#symbols", "TestWithSymbols"},
		{"complexCase123!@#", "ComplexCase123"},

		{"snake_case_string", "SnakeCaseString"},
		{"kebab-case-string", "KebabCaseString"},
		{"PascalCaseString", "PascalCaseString"},
		{"camelCaseString", "CamelCaseString"},
		{"HTTPRequest", "HttpRequest"},
		{"user ID", "UserId"},
		{"UserId", "UserId"},
		{"userID", "UserId"},
		{"UserID", "UserId"},
		{"123NumberPrefix", "123NumberPrefix"},
		{"__leading_underscores", "LeadingUnderscores"},
		{"trailing_underscores__", "TrailingUnderscores"},
		{"multiple___underscores", "MultipleUnderscores"},
		{" spaces around ", "SpacesAround"},
	}

	for _, test := range tests {
		result := UpperCamelCase(test.input)
		if result != test.expected {
			t.Errorf("UpperCamelCase(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}

func TestLowerCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "helloWorld"},
		{"hello_world", "helloWorld"},
		{"hello-world", "helloWorld"},
		{"hello.world", "helloWorld"},
		{"helloWorld", "helloWorld"},
		{"HelloWorld", "helloWorld"},
		{"HTTPStatusCode", "httpStatusCode"},
		{"ParseURL.DoParse", "parseUrlDoParse"},
		{"ParseUrl.DoParse", "parseUrlDoParse"},
		{"parse_url.do_parse", "parseUrlDoParse"},
		{"convert space", "convertSpace"},
		{"convert-dash", "convertDash"},
		{"skip___multiple_underscores", "skipMultipleUnderscores"},
		{"skip   multiple spaces", "skipMultipleSpaces"},
		{"skip---multiple-dashes", "skipMultipleDashes"},
		{"", ""},
		{"a", "a"},
		{"Z", "z"},
		{"special-characters_test", "specialCharactersTest"},
		{"numbers123test", "numbers123Test"},
		{"hello world!", "helloWorld"},
		{"test@with#symbols", "testWithSymbols"},
		{"complexCase123!@#", "complexCase123"},

		{"snake_case_string", "snakeCaseString"},
		{"kebab-case-string", "kebabCaseString"},
		{"PascalCaseString", "pascalCaseString"},
		{"camelCaseString", "camelCaseString"},
		{"HTTPRequest", "httpRequest"},
		{"user ID", "userId"},
		{"UserId", "userId"},
		{"userID", "userId"},
		{"UserID", "userId"},
		{"123NumberPrefix", "123NumberPrefix"},
		{"__leading_underscores", "leadingUnderscores"},
		{"trailing_underscores__", "trailingUnderscores"},
		{"multiple___underscores", "multipleUnderscores"},
		{" spaces around ", "spacesAround"},
	}

	for _, test := range tests {
		result := LowerCamelCase(test.input)
		if result != test.expected {
			t.Errorf("LowerCamelCase(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}
