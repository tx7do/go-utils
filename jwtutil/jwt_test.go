package jwtutil

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestParseJWTPayload(t *testing.T) {
	var secretKey = []byte("secret")

	// 测试有效的 JWT 负载解析
	token, err := GenerateJWT(jwt.MapClaims{
		"sub": "userId",
		"exp": 1672728000,
	}, secretKey, jwt.SigningMethodHS256)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	result, err := ParseJWTPayload(token)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "userId", result["sub"])
	assert.Equal(t, float64(1672728000), result["exp"]) // 注意：JSON 解码后数字为 float64

	// 测试无效的 JWT
	invalidToken := "invalid.token.string"
	result, err = ParseJWTPayload(invalidToken)
	assert.Error(t, err)
	assert.Nil(t, result)

	// 测试空字符串
	result, err = ParseJWTPayload("")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestParseJWTClaimsToStruct(t *testing.T) {
	// 测试有效的 JWT 负载解析
	type Payload struct {
		Sub string `json:"sub"`
		Exp int64  `json:"exp"`
	}

	var secretKey = []byte("secret")

	// 验证签名以生成有效的 token
	token, err := GenerateJWT(jwt.MapClaims{
		"sub": "userId",
		"exp": 1672728000,
	}, secretKey, jwt.SigningMethodHS256)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	result, err := ParseJWTClaimsToStruct[Payload](token)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "userId", result.Sub)
	assert.Equal(t, int64(1672728000), result.Exp)

	// 测试无效的 JWT
	invalidToken := "invalid.token.string"
	result, err = ParseJWTClaimsToStruct[Payload](invalidToken)
	assert.Error(t, err)
	assert.Nil(t, result)

	// 测试空字符串
	result, err = ParseJWTClaimsToStruct[Payload]("")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestVerifyJWT(t *testing.T) {
	// 测试有效的 JWT 验证
	secretKey := []byte("secret")

	token, err := GenerateJWT(jwt.MapClaims{
		"sub": "userId",
	}, secretKey, jwt.SigningMethodHS256)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	result, err := VerifyJWT(token, secretKey)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "userId", result["sub"])

	// 测试无效的签名
	invalidSecretKey := []byte("invalid_secret")
	result, err = VerifyJWT(token, invalidSecretKey)
	assert.Error(t, err)
	assert.Nil(t, result)

	// 测试无效的 JWT 格式
	invalidToken := "invalid.token.string"
	result, err = VerifyJWT(invalidToken, secretKey)
	assert.Error(t, err)
	assert.Nil(t, result)

	// 测试空字符串
	result, err = VerifyJWT("", secretKey)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGenerateGenericJWT(t *testing.T) {
	// 定义测试用的 payload 和密钥
	type Payload struct {
		Sub string `json:"sub"`
		Exp int64  `json:"exp"`
	}
	secretKey := []byte("secret")
	payload := Payload{
		Sub: "userId",
		Exp: 1672728000,
	}

	// 测试生成有效的 JWT
	token, err := GenerateGenericJWT(payload, secretKey, jwt.SigningMethodHS256)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 验证生成的 JWT 是否正确
	parsedPayload, err := ParseJWTClaimsToStruct[Payload](token)
	assert.NoError(t, err)
	assert.NotNil(t, parsedPayload)
	assert.Equal(t, payload.Sub, parsedPayload.Sub)
	assert.Equal(t, payload.Exp, parsedPayload.Exp)

	// 测试使用空密钥生成 JWT
	emptySecretKey := []byte("")
	token, err = GenerateGenericJWT(payload, emptySecretKey, jwt.SigningMethodHS256)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestGenerateJWT(t *testing.T) {
	// 定义测试用的 payload 和密钥
	payload := jwt.MapClaims{
		"sub": "userId",
		"exp": 1672728000,
	}
	secretKey := []byte("secret")

	// 测试生成有效的 JWT
	token, err := GenerateJWT(payload, secretKey, jwt.SigningMethodHS256)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 验证生成的 JWT 是否正确
	parsedPayload, err := ParseJWTPayload(token)
	assert.NoError(t, err)
	assert.NotNil(t, parsedPayload)
	assert.Equal(t, payload["sub"], parsedPayload["sub"])
	assert.Equal(t, float64(payload["exp"].(int)), parsedPayload["exp"].(float64)) // 注意：JSON 解码后数字为 float64

	// 测试使用空密钥生成 JWT
	emptySecretKey := []byte("")
	token, err = GenerateJWT(payload, emptySecretKey, jwt.SigningMethodHS256)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestRefreshJWT(t *testing.T) {
	secretKey := []byte("secret")

	// 创建一个初始的 JWT
	originalToken, err := GenerateJWT(jwt.MapClaims{
		"sub": "userId",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}, secretKey, jwt.SigningMethodHS256)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// 刷新 JWT，设置新的过期时间
	newExpiration := time.Now().Add(2 * time.Hour)
	refreshedToken, err := RefreshJWT(originalToken, secretKey, newExpiration)
	assert.NoError(t, err)
	assert.NotEmpty(t, refreshedToken)

	// 验证刷新后的 JWT 是否包含新的过期时间
	claims, err := ParseJWTPayload(refreshedToken)
	assert.NoError(t, err)
	assert.NotNil(t, claims)

	exp, ok := claims["exp"].(float64)
	assert.True(t, ok)
	assert.Equal(t, newExpiration.Unix(), int64(exp))

	// 测试无效的 JWT
	invalidToken := "invalid.token.string"
	_, err = RefreshJWT(invalidToken, secretKey, newExpiration)
	assert.Error(t, err)

	// 测试空字符串
	_, err = RefreshJWT("", secretKey, newExpiration)
	assert.Error(t, err)
}

func TestGenerateJWTWithHeader(t *testing.T) {
	// 定义测试用的 payload、密钥和自定义头部
	payload := jwt.MapClaims{
		"sub": "userId",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}
	secretKey := []byte("secret")
	customHeader := map[string]interface{}{
		"kid": "key-id-123",
	}

	// 测试生成有效的 JWT
	token, err := GenerateJWTWithHeader(payload, secretKey, jwt.SigningMethodHS256, customHeader)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 验证生成的 JWT 是否包含自定义头部
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	parsedToken, _, err := parser.ParseUnverified(token, jwt.MapClaims{})
	assert.NoError(t, err)
	assert.NotNil(t, parsedToken)

	// 检查头部是否包含自定义字段
	assert.Equal(t, "key-id-123", parsedToken.Header["kid"])

	// 测试使用空密钥生成 JWT
	emptySecretKey := []byte("")
	token, err = GenerateJWTWithHeader(payload, emptySecretKey, jwt.SigningMethodHS256, customHeader)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestExtractJWTFromRequest(t *testing.T) {
	// 测试有效的 Authorization Header
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid.jwt.token")
	token, err := ExtractJWTFromRequest(req)
	assert.NoError(t, err)
	assert.Equal(t, "valid.jwt.token", token)

	// 测试缺少 Authorization Header
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	token, err = ExtractJWTFromRequest(req)
	assert.Error(t, err)
	assert.Equal(t, "authorization header is missing", err.Error())
	assert.Empty(t, token)

	// 测试无效的 Authorization Header 格式
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	token, err = ExtractJWTFromRequest(req)
	assert.Error(t, err)
	assert.Equal(t, "invalid authorization header format", err.Error())
	assert.Empty(t, token)

	// 测试空的 Bearer Token
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer ")
	token, err = ExtractJWTFromRequest(req)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestGenerateShortLivedJWT(t *testing.T) {
	secretKey := []byte("secret")
	payload := jwt.MapClaims{
		"sub": "userId",
	}

	// 测试生成短期有效的 JWT
	duration := 1 * time.Hour
	token, err := GenerateShortLivedJWT(payload, secretKey, jwt.SigningMethodHS256, duration)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 验证生成的 JWT 是否包含正确的过期时间
	claims, err := ParseJWTPayload(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)

	exp, ok := claims["exp"].(float64)
	assert.True(t, ok)
	expectedExp := time.Now().Add(duration).Unix()
	assert.InDelta(t, expectedExp, int64(exp), 5) // 允许 5 秒的时间误差

	// 测试空密钥
	emptySecretKey := []byte("")
	token, err = GenerateShortLivedJWT(payload, emptySecretKey, jwt.SigningMethodHS256, duration)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestValidateJWTAudience(t *testing.T) {
	secretKey := []byte("secret")

	// 创建一个包含受众的 JWT
	token, err := GenerateJWT(jwt.MapClaims{
		"aud": []string{"audience1", "audience2"},
	}, secretKey, jwt.SigningMethodHS256)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 测试受众验证成功
	valid, err := ValidateJWTAudience(token, "audience1")
	assert.NoError(t, err)
	assert.True(t, valid)

	// 测试受众验证失败
	valid, err = ValidateJWTAudience(token, "audience3")
	assert.NoError(t, err)
	assert.False(t, valid)

	// 测试无效的 JWT
	invalidToken := "invalid.token.string"
	valid, err = ValidateJWTAudience(invalidToken, "audience1")
	assert.Error(t, err)
	assert.False(t, valid)

	// 测试空字符串
	valid, err = ValidateJWTAudience("", "audience1")
	assert.Error(t, err)
	assert.False(t, valid)
}

func TestValidateJWTAlgorithm(t *testing.T) {
	secretKey := []byte("secret")

	// 测试有效的 JWT 算法验证
	token, err := GenerateJWT(jwt.MapClaims{
		"sub": "userId",
	}, secretKey, jwt.SigningMethodHS256)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	valid, err := ValidateJWTAlgorithm(token, "HS256")
	assert.NoError(t, err)
	assert.True(t, valid)

	// 测试无效的算法
	valid, err = ValidateJWTAlgorithm(token, "RS256")
	assert.NoError(t, err)
	assert.False(t, valid)

	// 测试无效的 JWT
	invalidToken := "invalid.token.string"
	valid, err = ValidateJWTAlgorithm(invalidToken, "HS256")
	assert.Error(t, err)
	assert.False(t, valid)

	// 测试空字符串
	valid, err = ValidateJWTAlgorithm("", "HS256")
	assert.Error(t, err)
	assert.False(t, valid)
}

func TestIsJWTExpired(t *testing.T) {
	secretKey := []byte("secret")

	// 测试未过期的 JWT
	notExpiredToken, err := GenerateJWT(jwt.MapClaims{
		"sub": "userId",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}, secretKey, jwt.SigningMethodHS256)
	assert.NoError(t, err)
	assert.NotEmpty(t, notExpiredToken)

	isExpired, err := IsJWTExpired(notExpiredToken)
	assert.NoError(t, err)
	assert.False(t, isExpired)

	// 测试已过期的 JWT
	expiredToken, err := GenerateJWT(jwt.MapClaims{
		"sub": "userId",
		"exp": time.Now().Add(-1 * time.Hour).Unix(),
	}, secretKey, jwt.SigningMethodHS256)
	assert.NoError(t, err)
	assert.NotEmpty(t, expiredToken)

	isExpired, err = IsJWTExpired(expiredToken)
	assert.NoError(t, err)
	assert.True(t, isExpired)

	// 测试无效的 JWT
	invalidToken := "invalid.token.string"
	isExpired, err = IsJWTExpired(invalidToken)
	assert.Error(t, err)
	assert.False(t, isExpired)

	// 测试空字符串
	isExpired, err = IsJWTExpired("")
	assert.Error(t, err)
	assert.False(t, isExpired)
}

func TestGetJWTHeader(t *testing.T) {
	secretKey := []byte("secret")

	// 测试有效的 JWT 头部解析
	token, err := GenerateJWT(jwt.MapClaims{
		"sub": "userId",
	}, secretKey, jwt.SigningMethodHS256)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	header, err := GetJWTHeader(token)
	assert.NoError(t, err)
	assert.NotNil(t, header)
	assert.Equal(t, "JWT", header["typ"])
	assert.Equal(t, "HS256", header["alg"])

	// 测试无效的 JWT
	invalidToken := "invalid.token.string"
	header, err = GetJWTHeader(invalidToken)
	assert.Error(t, err)
	assert.Nil(t, header)

	// 测试空字符串
	header, err = GetJWTHeader("")
	assert.Error(t, err)
	assert.Nil(t, header)
}

func TestValidateJWTIssuer(t *testing.T) {
	secretKey := []byte("secret")

	// 测试有效的发行者验证
	token, err := GenerateJWT(jwt.MapClaims{
		"iss": "trusted-issuer",
	}, secretKey, jwt.SigningMethodHS256)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	valid, err := ValidateJWTIssuer(token, "trusted-issuer")
	assert.NoError(t, err)
	assert.True(t, valid)

	// 测试无效的发行者
	valid, err = ValidateJWTIssuer(token, "untrusted-issuer")
	assert.NoError(t, err)
	assert.False(t, valid)

	// 测试无效的 JWT
	invalidToken := "invalid.token.string"
	valid, err = ValidateJWTIssuer(invalidToken, "trusted-issuer")
	assert.Error(t, err)
	assert.False(t, valid)

	// 测试空字符串
	valid, err = ValidateJWTIssuer("", "trusted-issuer")
	assert.Error(t, err)
	assert.False(t, valid)
}

func TestGetJWTIssuedAt(t *testing.T) {
	secretKey := []byte("secret")

	// 测试有效的 JWT 签发时间
	issuedAt := time.Now().Add(-1 * time.Hour).Unix()
	token, err := GenerateJWT(jwt.MapClaims{
		"sub": "userId",
		"iat": issuedAt,
	}, secretKey, jwt.SigningMethodHS256)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	iat, err := GetJWTIssuedAt(token)
	assert.NoError(t, err)
	assert.NotNil(t, iat)
	assert.Equal(t, issuedAt, iat.Unix())

	// 测试无效的 JWT
	invalidToken := "invalid.token.string"
	iat, err = GetJWTIssuedAt(invalidToken)
	assert.Error(t, err)
	assert.Nil(t, iat)

	// 测试空字符串
	iat, err = GetJWTIssuedAt("")
	assert.Error(t, err)
	assert.Nil(t, iat)
}

func TestGetJWTClaims(t *testing.T) {
	secretKey := []byte("secret")

	// 测试有效的 JWT
	token, err := GenerateJWT(jwt.MapClaims{
		"sub": "userId",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}, secretKey, jwt.SigningMethodHS256)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := GetJWTClaims(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "userId", claims["sub"])

	// 测试无效的 JWT
	invalidToken := "invalid.token.string"
	claims, err = GetJWTClaims(invalidToken)
	assert.Error(t, err)
	assert.Nil(t, claims)

	// 测试空字符串
	claims, err = GetJWTClaims("")
	assert.Error(t, err)
	assert.Nil(t, claims)
}
