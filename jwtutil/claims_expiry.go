// claims_expiry.go —— 已验签 claims 的过期/未生效判定。
//
// 语义约定（与下游仓历史实现一致，勿单独漂移）：
//   - nil claims 一律视为过期/未生效（fail-closed）；
//   - 缺 exp/nbf 声明视为未过期/未未生效——判定函数只负责"有声明时校验"，
//     强制 exp 的策略由签发侧与验证侧另行约定；
//   - leeway 默认 60s，容忍服务器间时钟偏移。
package jwtutil

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// defaultTokenLeeway 令牌时间容差，防止因时间不同步导致的验证失败。
const defaultTokenLeeway = 60 * time.Second

// IsTokenExpired 检查令牌是否过期（缺 exp 声明视为未过期）。
func IsTokenExpired(claims jwt.MapClaims) bool {
	return IsTokenExpiredWithLeeway(claims, defaultTokenLeeway)
}

// IsTokenExpiredWithLeeway 以自定义容差检查令牌是否过期。
// exp 缺失视为未过期；claims 为 nil（nil map）视为过期。
func IsTokenExpiredWithLeeway(claims jwt.MapClaims, leeway time.Duration) bool {
	if claims == nil {
		return true
	}
	exp, _ := claims.GetExpirationTime()
	if exp == nil {
		return false
	}
	now := time.Now().UTC()
	return now.After(exp.Time.UTC().Add(leeway))
}

// IsTokenNotValidYet 检查令牌是否未生效（缺 nbf 声明视为已生效）。
func IsTokenNotValidYet(claims jwt.MapClaims) bool {
	if claims == nil {
		return true
	}
	nbf, _ := claims.GetNotBefore()
	if nbf == nil {
		return false
	}
	now := time.Now().UTC()
	return now.Add(defaultTokenLeeway).Before(nbf.UTC())
}
