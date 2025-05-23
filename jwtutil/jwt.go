package jwtutil

import (
	"fmt"
	"net/http"
	"time"

	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
)

// ParseJWTPayload 使用 github.com/golang-jwt/jwt/v5 从 JWT 中解析出 payload
func ParseJWTPayload(tokenString string) (jwt.MapClaims, error) {
	// 不验证签名，仅解析
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ParseJWTClaimsToStruct 解析 JWT 的负载部分，不验证签名，仅解析。
func ParseJWTClaimsToStruct[T any](tokenString string) (*T, error) {
	// 不验证签名，仅解析
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// 将 claims 转换为目标类型
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal claims: %v", err)
	}

	var ret T
	err = json.Unmarshal(claimsBytes, &ret)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %v", err)
	}

	return &ret, nil
}

// VerifyJWT 验证 JWT 的签名
func VerifyJWT(tokenString string, secretKey []byte) (jwt.MapClaims, error) {
	// 解析并验证 JWT
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 确保使用的是 HMAC 签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	// 验证 token 是否有效
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func GetJWTClaims(tokenString string) (map[string]interface{}, error) {
	claims, err := ParseJWTPayload(tokenString)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

// GenerateJWT 生成 JWT
func GenerateJWT(mapClaims jwt.MapClaims, secretKey []byte, signingMethod jwt.SigningMethod) (string, error) {
	// 检查密钥是否为空，密钥不能为空，否则会有安全隐患。
	if len(secretKey) == 0 {
		return "", fmt.Errorf("secret key cannot be empty")
	}

	// 创建一个新的 JWT Token
	token := jwt.NewWithClaims(signingMethod, mapClaims)

	// 使用 HMAC 签名方法签名 token
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// GenerateGenericJWT 生成 JWT
func GenerateGenericJWT[T any](payload T, secretKey []byte, signingMethod jwt.SigningMethod) (string, error) {
	// 检查密钥是否为空，密钥不能为空，否则会有安全隐患。
	if len(secretKey) == 0 {
		return "", fmt.Errorf("secret key cannot be empty")
	}

	// 将泛型 payload 转换为 jwt.MapClaims
	claimsMap, err := ToMapClaims(payload)
	if err != nil {
		return "", fmt.Errorf("failed to convert payload to claims: %v", err)
	}

	// 创建一个新的 JWT Token
	token := jwt.NewWithClaims(signingMethod, claimsMap)

	// 使用 HMAC 签名方法签名 token
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ToMapClaims 将泛型 payload 转换为 jwt.MapClaims
func ToMapClaims[T any](payload T) (jwt.MapClaims, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	var claims jwt.MapClaims
	err = json.Unmarshal(payloadBytes, &claims)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

// RefreshJWT 刷新JWT
func RefreshJWT(tokenString string, secretKey []byte, newExpiration time.Time) (string, error) {
	// 解析 JWT 的 payload
	claims, err := ParseJWTPayload(tokenString)
	if err != nil {
		return "", err
	}

	// 更新过期时间
	claims["exp"] = newExpiration.Unix()

	// 生成新的 JWT
	newToken, err := GenerateJWT(claims, secretKey, jwt.SigningMethodHS256)
	if err != nil {
		return "", err
	}

	return newToken, nil
}

// GenerateJWTWithHeader 生成带自定义头部的JWT
func GenerateJWTWithHeader(payload jwt.MapClaims, secretKey []byte, signingMethod jwt.SigningMethod, customHeader map[string]interface{}) (string, error) {
	// 检查密钥是否为空
	if len(secretKey) == 0 {
		return "", fmt.Errorf("secret key cannot be empty")
	}

	// 创建一个新的 JWT Token
	token := jwt.NewWithClaims(signingMethod, payload)

	// 添加自定义头部
	for key, value := range customHeader {
		token.Header[key] = value
	}

	// 使用密钥签名生成 JWT
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ExtractJWTFromRequest 从请求中提取JWT
func ExtractJWTFromRequest(r *http.Request) (string, error) {
	// 从 Authorization Header 中提取 JWT
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header is missing")
	}

	// 检查是否以 "Bearer " 开头
	const bearerPrefix = "Bearer "
	if len(authHeader) <= len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", fmt.Errorf("invalid authorization header format")
	}

	// 提取 JWT 部分
	token := authHeader[len(bearerPrefix):]
	return token, nil
}

// GenerateShortLivedJWT 生成短期有效的JWT
func GenerateShortLivedJWT(payload jwt.MapClaims, secretKey []byte, signingMethod jwt.SigningMethod, duration time.Duration) (string, error) {
	// 检查密钥是否为空
	if len(secretKey) == 0 {
		return "", fmt.Errorf("secret key cannot be empty")
	}

	// 设置过期时间
	expirationTime := time.Now().Add(duration).Unix()
	payload["exp"] = expirationTime

	// 创建一个新的 JWT Token
	token := jwt.NewWithClaims(signingMethod, payload)

	// 使用密钥签名生成 JWT
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ValidateJWTAudience 验证JWT受众
func ValidateJWTAudience(tokenString string, expectedAudience string) (bool, error) {
	// 解析 JWT 的 payload
	claims, err := ParseJWTPayload(tokenString)
	if err != nil {
		return false, err
	}

	// 获取 `aud` 字段
	audience, err := claims.GetAudience()
	if err != nil {
		return false, err
	}

	// 验证 `aud` 是否包含预期值
	for _, aud := range audience {
		if aud == expectedAudience {
			return true, nil
		}
	}

	return false, nil
}

// ValidateJWTAlgorithm 验证JWT算法
func ValidateJWTAlgorithm(tokenString string, expectedAlgorithm string) (bool, error) {
	// 使用 jwt.Parser 解析 JWT，不验证签名
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return false, err
	}

	// 获取 `alg` 字段
	alg, ok := token.Header["alg"].(string)
	if !ok {
		return false, fmt.Errorf("invalid or missing algorithm in token header")
	}

	// 验证算法是否符合预期
	if alg != expectedAlgorithm {
		return false, nil
	}

	return true, nil
}

// IsJWTExpired 检查JWT是否过期
func IsJWTExpired(tokenString string) (bool, error) {
	// 解析 JWT 的 payload
	claims, err := ParseJWTPayload(tokenString)
	if err != nil {
		return false, err
	}

	// 获取 `exp` 字段
	exp, err := claims.GetExpirationTime()
	if err != nil {
		return false, err
	}

	// 检查当前时间是否超过 `exp`
	if time.Now().After(exp.Time) {
		return true, nil
	}

	return false, nil
}

// GetJWTHeader 获取JWT头部
func GetJWTHeader(tokenString string) (map[string]interface{}, error) {
	// 使用 jwt.Parser 解析 JWT，不验证签名
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}

	return token.Header, nil
}

// ValidateJWTIssuer 验证JWT发行者
func ValidateJWTIssuer(tokenString string, expectedIssuer string) (bool, error) {
	// 解析 JWT 的 payload
	claims, err := ParseJWTPayload(tokenString)
	if err != nil {
		return false, err
	}

	// 获取 `iss` 字段
	issuer, err := claims.GetIssuer()
	if err != nil {
		return false, err
	}

	// 验证 `iss` 是否符合预期
	if issuer != expectedIssuer {
		return false, nil
	}

	return true, nil
}

// GetJWTIssuedAt 获取JWT签发时间
func GetJWTIssuedAt(tokenString string) (*time.Time, error) {
	claims, err := ParseJWTPayload(tokenString)
	if err != nil {
		return nil, err
	}

	iat, err := claims.GetIssuedAt()
	if err != nil {
		return nil, err
	}

	return &iat.Time, nil
}
