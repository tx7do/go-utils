package stringcase

import (
	"strings"
	"unicode"
)

const snakeCaseDelimiters rune = '_'

// ToSnakeCase 把字符转换为 蛇形命名法（snake_case）
func ToSnakeCase(input string) string {
	return SnakeCase(input)
}

// SnakeCase 把字符转换为 蛇形命名法（snake_case）
func SnakeCase(s string) string {
	return delimiterCase(s, snakeCaseDelimiters, false)
}

// UpperSnakeCase 把字符转换为 大写蛇形命名法（UPPER_SNAKE_CASE）
func UpperSnakeCase(s string) string {
	return delimiterCase(s, snakeCaseDelimiters, true)
}

// delimiterCase 使用指定的分隔符和大小写规则转换字符串
func delimiterCase(input string, delimiter rune, upperCase bool) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	words := Split(input)
	filteredWords := make([]string, 0, len(words))
	for _, word := range words {
		if w := strings.TrimSpace(word); w != "" {
			filteredWords = append(filteredWords, w)
		}
	}

	if len(filteredWords) == 0 {
		return ""
	}

	// 基于原始输入位置合并相邻词：仅在前词为字母且当前词为纯数字时合并
	trimmedInput := input
	merged := make([]string, 0, len(filteredWords))
	offset := 0
	prevEnd := -1

	isLettersOnly := func(s string) bool {
		if s == "" {
			return false
		}
		for _, r := range s {
			if !unicode.IsLetter(r) {
				return false
			}
		}
		return true
	}
	isDigitsOnly := func(s string) bool {
		if s == "" {
			return false
		}
		for _, r := range s {
			if !unicode.IsDigit(r) {
				return false
			}
		}
		return true
	}

	for _, w := range filteredWords {
		idx := strings.Index(trimmedInput[offset:], w)
		start := -1
		if idx != -1 {
			start = offset + idx
		}
		if start != -1 && prevEnd == start && len(merged) > 0 {
			// 仅当前词为纯数字且前词为纯字母时合并
			prev := merged[len(merged)-1]
			if isLettersOnly(prev) && isDigitsOnly(w) {
				merged[len(merged)-1] = prev + w
				prevEnd = start + len(w)
				offset = prevEnd
				continue
			}
		}
		// 否则作为新词追加
		merged = append(merged, w)
		if start != -1 {
			prevEnd = start + len(w)
			offset = prevEnd
		} else {
			// 找不到位置时，推进 offset 到末尾以避免无限循环
			offset = len(trimmedInput)
			prevEnd = offset
		}
	}

	// 大小写处理
	for i, word := range merged {
		runes := []rune(word)
		for j, r := range runes {
			if upperCase {
				runes[j] = unicode.ToUpper(r)
			} else {
				runes[j] = unicode.ToLower(r)
			}
		}
		merged[i] = string(runes)
	}

	return strings.Join(merged, string(delimiter))
}

// IsSnakeCase 判断字符串是否为 `snake_case`，规则：
// - 非空
// - 不能以下划线开头或结尾
// - 不能包含大写字母
// - 不允许连续下划线
// - 仅允许小写字母、数字和下划线
func IsSnakeCase(s string) bool {
	if s == "" {
		return false
	}
	runes := []rune(s)
	// 首尾不能是下划线
	if runes[0] == '_' || runes[len(runes)-1] == '_' {
		return false
	}
	prevUnderscore := false
	for _, r := range runes {
		switch {
		case r >= 'a' && r <= 'z':
			prevUnderscore = false
		case r >= '0' && r <= '9':
			prevUnderscore = false
		case r == snakeCaseDelimiters:
			if prevUnderscore {
				return false
			}
			prevUnderscore = true
		// 排除大写字母和其它非法字符
		default:
			return false
		}
	}
	return true
}
