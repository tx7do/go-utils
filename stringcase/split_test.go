package stringcase

import (
	"testing"
)

func TestSplitSingle(t *testing.T) {
	input := "URL.DoParse"
	result := Split(input)
	t.Logf("Split(%q) = %q;", input, result)
	t.Log(input[:7])
}

func TestSplit(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"hello world", []string{"hello", "world"}},
		{"hello_world", []string{"hello", "world"}},
		{"hello-world", []string{"hello", "world"}},
		{"hello.world", []string{"hello", "world"}},
		{"helloWorld", []string{"hello", "World"}},
		{"HelloWorld", []string{"Hello", "World"}},
		{"HTTPStatusCode", []string{"HTTP", "Status", "Code"}},
		{"ParseURLDoParse", []string{"Parse", "URL", "Do", "Parse"}},
		{"ParseUrlDoParse", []string{"Parse", "Url", "Do", "Parse"}},
		{"ParseUrl.DoParse", []string{"Parse", "Url", "Do", "Parse"}},
		{"ParseURL.DoParse", []string{"Parse", "URL", "Do", "Parse"}},
		{"ParseURL", []string{"Parse", "URL"}},
		{"ParseURL.", []string{"Parse", "URL"}},
		{"parse_url.do_parse", []string{"parse", "url", "do", "parse"}},
		{"convert space", []string{"convert", "space"}},
		{"convert-dash", []string{"convert", "dash"}},
		{"skip___multiple_underscores", []string{"skip", "multiple", "underscores"}},
		{"skip   multiple spaces", []string{"skip", "multiple", "spaces"}},
		{"skip---multiple-dashes", []string{"skip", "multiple", "dashes"}},
		{"", []string{""}},
		{"a", []string{"a"}},
		{"Z", []string{"Z"}},
		{"special-characters_test", []string{"special", "characters", "test"}},
		{"numbers123test", []string{"numbers", "123", "test"}},
		{"hello world!", []string{"hello", "world"}},
		{"test@with#symbols", []string{"test", "with", "symbols"}},
		{"complexCase123!@#", []string{"complex", "Case", "123"}},

		{"snake_case_string", []string{"snake", "case", "string"}},
		{"kebab-case-string", []string{"kebab", "case", "string"}},
		{"PascalCaseString", []string{"Pascal", "Case", "String"}},
		{"camelCaseString", []string{"camel", "Case", "String"}},
		{"HTTPRequest", []string{"HTTP", "Request"}},
		{"user ID", []string{"user", "ID"}},
		{"UserId", []string{"User", "Id"}},
		{"userID", []string{"user", "ID"}},
		{"UserID", []string{"User", "ID"}},
		{"123NumberPrefix", []string{"123", "Number", "Prefix"}},
		{"__leading_underscores", []string{"leading", "underscores"}},
		{"trailing_underscores__", []string{"trailing", "underscores"}},
		{"multiple___underscores", []string{"multiple", "underscores"}},
		{" spaces around ", []string{"spaces", "around"}},
	}

	for _, test := range tests {
		result := Split(test.input)

		if !compareStringSlices(result, test.expected) {
			t.Errorf("Split(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}

func compareStringSlices(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}
