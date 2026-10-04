package password

import (
	"errors"
	"strconv"
	"unicode"
)

// ErrWeakPassword 复杂度不达标（调用方转成 4xx 响应）。
var ErrWeakPassword = errors.New("password does not meet complexity requirements: min length 8 and at least 3 of (lowercase, uppercase, digit, symbol)")

// ValidateComplexity 校验明文口令复杂度：长度达到 minLen 且至少包含
// 小写/大写/数字/符号 四类中的三类。minLen 由调用方从其配置体系读取
// （本库不管配置来源），传 0 或负值时按 1 处理（仅校验字符类别）。
func ValidateComplexity(plain string, minLen int) error {
	if minLen < 1 {
		minLen = 1
	}
	if len(plain) < minLen {
		return errors.New("password too short: minimum length is " + strconv.Itoa(minLen))
	}
	var hasLower, hasUpper, hasDigit, hasSymbol bool
	for _, r := range plain {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}
	classes := 0
	for _, b := range []bool{hasLower, hasUpper, hasDigit, hasSymbol} {
		if b {
			classes++
		}
	}
	if classes < 3 {
		return ErrWeakPassword
	}
	return nil
}
