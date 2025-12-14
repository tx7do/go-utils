package stringcase

import (
	"strings"
	"unicode"
)

// UpperCamelCase 把字符转换为 帕斯卡命名/大驼峰命名法（CamelCase）
func UpperCamelCase(input string) string {
	return camelCase(input, true)
}

// LowerCamelCase 把字符转换为 小驼峰命名法（lowerCamelCase）
func LowerCamelCase(input string) string {
	return camelCase(input, false)
}

// ToPascalCase 把字符转换为 帕斯卡命名/大驼峰命名法（CamelCase）
func ToPascalCase(input string) string {
	return camelCase(input, true)
}

// ToCamelCase 把字符转换为 小驼峰命名法（lowerCamelCase）
func ToCamelCase(input string) string {
	return camelCase(input, false)
}

// PascalCase 把字符转换为 帕斯卡命名/大驼峰命名法（CamelCase）
func PascalCase(input string) string {
	return camelCase(input, true)
}

// CamelCase 把字符转换为 小驼峰命名法（lowerCamelCase）
func CamelCase(input string) string {
	return camelCase(input, false)
}

// camelCase 是一个通用的驼峰命名转换函数。
// 参数 upper 决定了首字母是否大写：
//   - upper 为 true 时，生成 UpperCamelCase（帕斯卡命名法）
//   - upper 为 false 时，生成 lowerCamelCase（小驼峰命名法）
func camelCase(input string, upper bool) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return input
	}

	// 分割字符串
	words := Split(input)
	if len(words) == 0 {
		return ""
	}

	filteredWords := make([]string, 0, len(words))
	for _, word := range words {
		if strings.TrimSpace(word) != "" {
			filteredWords = append(filteredWords, word)
		}
	}
	words = filteredWords
	if len(words) == 0 {
		return ""
	}

	for i, word := range words {
		if word == "" {
			continue
		}

		runes := []rune(word)
		if len(runes) > 0 {
			if i == 0 && !upper {
				runes[0] = unicode.ToLower(runes[0]) // LowerCamelCase首单词首字母小写
			} else {
				runes[0] = unicode.ToUpper(runes[0]) // UpperCamelCase或后续单词首字母大写
			}
			for j := 1; j < len(runes); j++ {
				runes[j] = unicode.ToLower(runes[j]) // 其余字母统一小写
			}
			words[i] = string(runes)
		}
	}

	// 合并结果
	return strings.Join(words, "")
}
