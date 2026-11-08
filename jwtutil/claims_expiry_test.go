package jwtutil

// claims_expiry_test.go —— 过期/未生效判定的边界锁定。

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func futureExp(d time.Duration) jwt.MapClaims {
	return jwt.MapClaims{"exp": float64(time.Now().Add(d).Unix())}
}

func TestIsTokenExpired(t *testing.T) {
	assert.True(t, IsTokenExpired(nil), "nil claims 视为过期")
	assert.True(t, IsTokenExpired(jwt.MapClaims{"exp": float64(time.Now().Add(-2 * time.Hour).Unix())}),
		"明显过期必须判定过期")
	assert.False(t, IsTokenExpired(futureExp(2*time.Hour)), "未来 exp 未过期")
	assert.False(t, IsTokenExpired(jwt.MapClaims{}), "缺 exp 声明视为未过期（约定）")
	// 容差边界：过期 30s（< 60s leeway）不算过期；过期 2 分钟算过期
	assert.False(t, IsTokenExpired(futureExp(-30*time.Second)), "leeway 内不算过期")
	assert.True(t, IsTokenExpired(futureExp(-2*time.Minute)), "超出 leeway 判定过期")
}

func TestIsTokenExpiredWithLeeway(t *testing.T) {
	claims := futureExp(-30 * time.Second)
	assert.False(t, IsTokenExpiredWithLeeway(claims, time.Minute), "自定义大容差内不算过期")
	assert.True(t, IsTokenExpiredWithLeeway(claims, time.Second), "自定义小容差判定过期")
	assert.True(t, IsTokenExpiredWithLeeway(nil, time.Minute), "nil claims 恒过期")
}

func TestIsTokenNotValidYet(t *testing.T) {
	assert.True(t, IsTokenNotValidYet(nil), "nil claims 视为未生效")
	assert.False(t, IsTokenNotValidYet(jwt.MapClaims{"nbf": float64(time.Now().Add(-time.Hour).Unix())}),
		"早已生效的令牌")
	assert.False(t, IsTokenNotValidYet(jwt.MapClaims{}), "缺 nbf 声明视为已生效（约定）")
	assert.True(t, IsTokenNotValidYet(jwt.MapClaims{"nbf": float64(time.Now().Add(time.Hour).Unix())}),
		"一小时后才生效的令牌")
}
