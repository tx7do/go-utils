package stringutil

import (
	"fmt"
	"regexp"
)

// ReplaceJSONField 使用正则表达式替换 JSON 字符串中指定字段的值
// fieldNames: 要替换的多个字段名，使用竖线(|)分隔（例如："tenantId|tenant_id"）
// newValue: 新的值（字符串形式，将被包装在引号中）
// jsonStr: 原始 JSON 字符串
func ReplaceJSONField(fieldNames, newValue, jsonStr string) string {
	// 构建正则表达式模式
	// 匹配模式: ("field1"|"field2"|...): "任意值"
	pattern := fmt.Sprintf(`(?i)("(%s)")\s*:\s*"([^"]*)"`, fieldNames)
	re := regexp.MustCompile(pattern)

	// 替换匹配到的内容
	return re.ReplaceAllString(jsonStr, fmt.Sprintf(`${1}: "%s"`, newValue))
}
