package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateComplexity(t *testing.T) {
	t.Run("valid passwords", func(t *testing.T) {
		for _, pw := range []string{
			"Abcd@1234",   // 四类齐全
			"abcdefgh1A",  // 小写+数字+大写
			"Abcdefg1",    // 长度刚好 8
			"abcd@123456", // 小写+符号+数字
		} {
			assert.NoError(t, ValidateComplexity(pw, 8), "密码 %q 应通过校验", pw)
		}
	})

	t.Run("too short", func(t *testing.T) {
		err := ValidateComplexity("Ab1@", 8)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "too short")
	})

	t.Run("insufficient character classes", func(t *testing.T) {
		// 只有小写+数字两类
		err := ValidateComplexity("abcdefg12345", 8)
		require.ErrorIs(t, err, ErrWeakPassword)
		// 长度达标但单一类别
		err = ValidateComplexity("abcdefghijklmnopqrstuvwxyz", 8)
		require.ErrorIs(t, err, ErrWeakPassword)
	})

	t.Run("symbol and unicode classes", func(t *testing.T) {
		// 符号计为一类
		assert.NoError(t, ValidateComplexity("Abcdefg!", 8))
		// unicode 符号计为一类
		assert.NoError(t, ValidateComplexity("Abcdefg€", 8))
	})

	t.Run("non-positive minLen clamps to 1", func(t *testing.T) {
		// 仅类别校验，不再要求长度
		assert.ErrorIs(t, ValidateComplexity("a1", 0), ErrWeakPassword) // 两类不足
		assert.NoError(t, ValidateComplexity("Ab1", 0))                 // 三类
	})
}
