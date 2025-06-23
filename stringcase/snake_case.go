package stringcase

import (
	"strings"
)

// ToSnakeCase 把字符转换为 蛇形命名法（snake_case）
func ToSnakeCase(input string) string {
	return SnakeCase(input)
}

func SnakeCase(s string) string {
	return delimiterCase(s, '_', false)
}

func UpperSnakeCase(s string) string {
	return delimiterCase(s, '_', true)
}

func delimiterCase(input string, delimiter rune, upperCase bool) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return input
	}

	// 使用 Split 分割字符串
	words := Split(input)
	filteredWords := make([]string, 0, len(words))
	for _, word := range words {
		if strings.TrimSpace(word) != "" {
			filteredWords = append(filteredWords, word)
		}
	}

	adjustCase := toLower
	if upperCase {
		adjustCase = toUpper
	}

	for i, word := range filteredWords {
		runes := []rune(word)
		for j := 0; j < len(runes); j++ {
			runes[j] = adjustCase(runes[j])
		}
		filteredWords[i] = string(runes)
	}

	// 使用分隔符连接结果
	return strings.Join(filteredWords, string(delimiter))
}
