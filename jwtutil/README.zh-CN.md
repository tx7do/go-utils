# jwtutil

[English](README.md) | 简体中文

`jwtutil` 是一个独立 Go 模块（`github.com/tx7do/go-utils/jwtutil`），基于 [github.com/golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt/v5) 提供 JWT 的签发、验证与负载检查工具函数，并附带从 `http.Request` 提取 Bearer Token、生成安全随机刷新令牌与 JWT ID（jti）的辅助能力。

## 特性

- 签发：`GenerateJWT` / `GenerateGenericJWT`（任意结构体直接作 payload）/ `GenerateJWTWithHeader`（自定义头部，如 `kid`）/ `GenerateShortLivedJWT`（自动写入 `exp`）/ `RefreshJWT`（续期）
- 验证：`VerifyJWT` 使用 HMAC 密钥完成签名与标准声明（含 `exp` 过期）校验
- 无验签解析：`ParseJWTPayload` / `GetJWTClaims` / `ParseJWTClaimsToStruct`（泛型映射到任意结构体）/ `GetJWTHeader`
- 声明级检查：`ValidateJWTAudience` / `ValidateJWTIssuer` / `ValidateJWTAlgorithm` / `IsJWTExpired` / `GetJWTIssuedAt`
- HTTP 集成：`ExtractJWTFromRequest` 从 `Authorization: Bearer <token>` 提取令牌
- 安全随机：`NewRefreshToken`（`crypto/rand` + Base64URL）、`NewJWTId`（UUIDv7 十六进制 jti）
- 与 kratos 生态共享 JSON 编解码器：结构体与 claims 的互转走 `encoding.GetCodec("json")`

## 安装

本包是独立模块（有自己的 `go.mod`），不是根模块 `github.com/tx7do/go-utils` 的子包，要求 Go 1.23+：

```bash
go get github.com/tx7do/go-utils/jwtutil
```

```go
import "github.com/tx7do/go-utils/jwtutil"
```

## 快速开始

```go
package main

import (
	"fmt"
	"log"
	"time"

	_ "github.com/go-kratos/kratos/v2/encoding/json" // 注册 "json" codec（泛型互转必需，见下文）

	"github.com/golang-jwt/jwt/v5"

	jwtutil "github.com/tx7do/go-utils/jwtutil"
)

func main() {
	secretKey := []byte("secret")

	// 签发：claims 中的标准字段（exp / iat / iss / aud 等）由调用方自行写入
	token, err := jwtutil.GenerateJWT(jwt.MapClaims{
		"sub": "userId",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}, secretKey, jwt.SigningMethodHS256)
	if err != nil {
		log.Fatal(err)
	}

	// 验证：仅接受 HMAC 算法，并校验 exp 等标准声明
	claims, err := jwtutil.VerifyJWT(token, secretKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("sub:", claims["sub"])

	// 不验签解析负载（数字解码后为 float64）
	payload, err := jwtutil.ParseJWTPayload(token)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("exp:", payload["exp"]) // float64

	// 直接映射到自定义结构体（同样不验签）
	type Payload struct {
		Sub string `json:"sub"`
		Exp int64  `json:"exp"`
	}
	p, err := jwtutil.ParseJWTClaimsToStruct[Payload](token)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(p.Sub, p.Exp)
}
```

### 泛型签发：结构体直接作 payload

```go
type Payload struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
}

token, err := jwtutil.GenerateGenericJWT(Payload{
	Sub: "userId",
	Exp: time.Now().Add(time.Hour).Unix(),
}, secretKey, jwt.SigningMethodHS256)
```

### 短期令牌 / 自定义头部 / 续期

```go
// 短期令牌：将 payload["exp"] 覆写为 now+duration（会改动传入的 map）
short, err := jwtutil.GenerateShortLivedJWT(jwt.MapClaims{"sub": "userId"},
	secretKey, jwt.SigningMethodHS256, 15*time.Minute)

// 自定义头部（合并进默认的 alg / typ 之外）
token, err := jwtutil.GenerateJWTWithHeader(claims, secretKey, jwt.SigningMethodHS256,
	map[string]interface{}{"kid": "key-id-123"})

// 续期：改写 exp 后重新签名（不验证旧令牌签名，且固定使用 HS256）
refreshed, err := jwtutil.RefreshJWT(token, secretKey, time.Now().Add(2*time.Hour))
```

### 从 HTTP 请求提取令牌

```go
token, err := jwtutil.ExtractJWTFromRequest(req) // 严格匹配大小写敏感的 "Bearer " 前缀
```

## 配置与约定

本模块没有 Options 结构体，行为约定如下：

- **标准声明由调用方负责**：除 `GenerateShortLivedJWT`（覆写 `exp`）和 `RefreshJWT`（改写 `exp`）外，签发函数不会自动写入 `exp` / `iat` / `iss` / `aud`，需要什么标准声明就自己在 claims 里放什么。
- **密钥**：所有签发函数都要求 `secretKey []byte` 非空（空密钥返回 `secret key cannot be empty` 错误）。`[]byte` 密钥只适用于 HMAC 家族（HS256/HS384/HS512）；传入 RS256/ES256 等算法会在签名时返回 `key is of invalid type: ...` 错误。
- **签名算法**：`signingMethod` 为任意的 `jwt.SigningMethod`（如 `jwt.SigningMethodHS256`）。
- **JSON 编解码器注册（必需）**：`ToMapClaims` / `GenerateGenericJWT` / `ParseJWTClaimsToStruct` 通过 kratos 的 `encoding.GetCodec("json")` 做互转。kratos 的 codec 靠包的 `init` 注册，若程序中没有（直接或间接）导入任何 codec，这三个函数会因 codec 为 nil 而 panic。需在入口处空白导入：

  ```go
  import _ "github.com/go-kratos/kratos/v2/encoding/json"
  ```

- **数字类型**：`MapClaims` 经 JSON 解码后数字一律为 `float64`（例如 `exp: 1672728000` 读出来是 `float64(1672728000)`）。

## API 参考

### 签发

```go
func GenerateJWT(mapClaims jwt.MapClaims, secretKey []byte, signingMethod jwt.SigningMethod) (string, error)
func GenerateGenericJWT[T any](payload T, secretKey []byte, signingMethod jwt.SigningMethod) (string, error)
func GenerateJWTWithHeader(payload jwt.MapClaims, secretKey []byte, signingMethod jwt.SigningMethod, customHeader map[string]interface{}) (string, error)
func GenerateShortLivedJWT(payload jwt.MapClaims, secretKey []byte, signingMethod jwt.SigningMethod, duration time.Duration) (string, error)
func RefreshJWT(tokenString string, secretKey []byte, newExpiration time.Time) (string, error)
func ToMapClaims[T any](payload T) (jwt.MapClaims, error)
```

- `GenerateJWT`：用 `jwt.NewWithClaims(signingMethod, mapClaims)` + `SignedString(secretKey)` 签发。
- `GenerateGenericJWT`：先经 `ToMapClaims` 把任意 payload 转 `jwt.MapClaims`，再同上。
- `GenerateJWTWithHeader`：把 `customHeader` 逐项合并进 token 头部（默认已有 `alg`、`typ`）。
- `GenerateShortLivedJWT`：`payload["exp"] = time.Now().Add(duration).Unix()`（覆写已有 `exp`，且会改动传入的 map）。
- `RefreshJWT`：`ParseJWTPayload` 取出 claims → `claims["exp"] = newExpiration.Unix()` → 固定用 `jwt.SigningMethodHS256` 重新签发。**不验证旧令牌签名**。
- `ToMapClaims`：经 JSON codec 序列化/反序列化把任意 payload 转 `jwt.MapClaims`。

### 验证与声明级检查

```go
func VerifyJWT(tokenString string, secretKey []byte) (jwt.MapClaims, error)
func ValidateJWTAlgorithm(tokenString string, expectedAlgorithm string) (bool, error)
func ValidateJWTAudience(tokenString string, expectedAudience string) (bool, error)
func ValidateJWTIssuer(tokenString string, expectedIssuer string) (bool, error)
func IsJWTExpired(tokenString string) (bool, error)
```

- `VerifyJWT`：唯一的完整验证入口。keyFunc 强制要求 HMAC 签名方法（否则报 `unexpected signing method: %v`），随后由 jwt/v5 完成签名与标准声明校验（过期返回 `token has invalid claims: token is expired`）。HS256/HS384/HS512 均可。
- `ValidateJWTAlgorithm`：读头部 `alg` 与期望值比较；缺失或非字符串报 `invalid or missing algorithm in token header`。
- `ValidateJWTAudience`：`aud` 支持 string 或数组，包含期望值即 `true`；`aud` 缺失返回 `(false, nil)`。
- `ValidateJWTIssuer`：`iss` 精确相等即 `true`；`iss` 缺失返回 `(false, nil)`。
- `IsJWTExpired`：`exp` 与当前时间比较（严格大于才判定过期）。**`exp` 缺失时会 panic**（见注意事项）。

### 负载与头部解析（均不验证签名）

```go
func ParseJWTPayload(tokenString string) (jwt.MapClaims, error)
func ParseJWTClaimsToStruct[T any](tokenString string) (*T, error)
func GetJWTClaims(tokenString string) (map[string]interface{}, error)
func GetJWTHeader(tokenString string) (map[string]interface{}, error)
func GetJWTIssuedAt(tokenString string) (*time.Time, error)
func ExtractJWTFromRequest(r *http.Request) (string, error)
```

- `ParseJWTPayload`：`jwt.Parser`（`WithoutClaimsValidation`）+ `ParseUnverified`，只解格式不验签。
- `ParseJWTClaimsToStruct`：同上，再经 JSON codec 反序列化为任意类型。
- `GetJWTClaims`：`ParseJWTPayload` 的别名（返回类型放宽为 `map[string]interface{}`）。
- `GetJWTHeader`：返回完整头部（`alg`、`typ` 及自定义字段）。
- `GetJWTIssuedAt`：读 `iat`，返回 `*time.Time`。**`iat` 缺失时会 panic**（见注意事项）。
- `ExtractJWTFromRequest`：要求 `Authorization` 头严格以大小写敏感的 `"Bearer "` 开头；缺失报 `authorization header is missing`，前缀不符（含空 token）报 `invalid authorization header format`。

### 安全随机令牌

```go
func NewRefreshToken() (string, error)
func NewJWTId() string
```

- `NewRefreshToken`：`crypto/rand` 读 32 字节，Base64URL（无填充）编码，输出固定 43 字符、仅含 `[A-Za-z0-9_-]`。
- `NewJWTId`：UUIDv7 的十六进制编码，32 位小写十六进制字符、无连字符；系统时钟不可靠时回退到 UUIDv4。

### 验签行为一览

| 函数 | 验证签名 | 校验标准声明（exp 等） |
|------|:--:|:--:|
| `VerifyJWT` | 是（仅 HMAC） | 是（过期报错） |
| `ParseJWTPayload` / `GetJWTClaims` / `ParseJWTClaimsToStruct` | 否 | 否 |
| `GetJWTHeader` / `ValidateJWTAlgorithm` | 否 | 否 |
| `ValidateJWTAudience` / `ValidateJWTIssuer` | 否 | 否（仅读取该声明） |
| `IsJWTExpired` / `GetJWTIssuedAt` | 否 | 否（仅读取该声明） |
| `RefreshJWT` | 否 | 否（改写 exp 后重签） |

## 注意事项

- **未验签即信任**：除 `VerifyJWT` 外，本模块所有解析/检查函数都不验证签名（基于 `ParseUnverified` / `WithoutClaimsValidation`）。`RefreshJWT`、`IsJWTExpired`、`ValidateJWTIssuer` 等用在安全决策前，请先自行 `VerifyJWT`。
- **缺失声明会 panic**：jwt/v5 的 `GetExpirationTime` / `GetIssuedAt` 对缺失字段返回 `(nil, nil)`，因此 `IsJWTExpired` 遇到没有 `exp` 的令牌、`GetJWTIssuedAt` 遇到没有 `iat` 的令牌会发生空指针 panic。调用前确保令牌带有对应声明，或先行 recover。
- **`RefreshJWT` 不验旧令牌且固定 HS256**：无论原令牌用什么算法（HS384 也一样），续期结果一律以 HS256 重签。
- **`[]byte` 密钥只适配 HMAC**：`GenerateJWT` 系列传 RS256/ES256 等非 HMAC 算法时，签名阶段会返回 `key is of invalid type: ...` 错误。
- **codec 未注册会 panic**：`ToMapClaims` / `GenerateGenericJWT` / `ParseJWTClaimsToStruct` 依赖 kratos JSON codec 的注册（见配置与约定）。
- `GenerateShortLivedJWT` 会覆写并修改传入的 claims map（`exp` 键）。
- 数字类声明经 JSON 往返后均为 `float64`；`ParseJWTClaimsToStruct` 映射到带 `json` 标签的结构体时由 codec 负责类型还原（`exp float64` → `int64`）。
- `ExtractJWTFromRequest` 不做除前缀检查以外的任何校验，取出的令牌仍需 `VerifyJWT`。

## 致谢

- [github.com/golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt/v5) — JWT 解析、签名与验证
- [github.com/go-kratos/kratos/v2](https://github.com/go-kratos/kratos) — JSON codec 注册表（`encoding.GetCodec`）
- [github.com/google/uuid](https://github.com/google/uuid) — UUIDv7（`NewJWTId`）

## 许可证

MIT
