# jwtutil

English | [简体中文](README.zh-CN.md)

`jwtutil` is a standalone Go module (`github.com/tx7do/go-utils/jwtutil`) providing JWT issuing, verification, and payload-inspection helpers built on [github.com/golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt/v5), plus utilities for extracting Bearer tokens from `http.Request` and generating cryptographically random refresh tokens and JWT IDs (jti).

## Features

- Issuing: `GenerateJWT` / `GenerateGenericJWT` (use any struct as the payload) / `GenerateJWTWithHeader` (custom headers such as `kid`) / `GenerateShortLivedJWT` (writes `exp` automatically) / `RefreshJWT` (renewal)
- Verification: `VerifyJWT` validates the signature with an HMAC key plus the standard claims (including `exp` expiry)
- Parse without verification: `ParseJWTPayload` / `GetJWTClaims` / `ParseJWTClaimsToStruct` (generically maps claims onto any struct) / `GetJWTHeader`
- Claim-level checks: `ValidateJWTAudience` / `ValidateJWTIssuer` / `ValidateJWTAlgorithm` / `IsJWTExpired` / `GetJWTIssuedAt`
- HTTP integration: `ExtractJWTFromRequest` pulls the token out of `Authorization: Bearer <token>`
- Secure randomness: `NewRefreshToken` (`crypto/rand` + Base64URL), `NewJWTId` (UUIDv7 hex jti)
- Shares the kratos JSON codec: struct↔claims conversion goes through `encoding.GetCodec("json")`

## Installation

This package is a standalone module (it has its own `go.mod`), not a subpackage of the root module `github.com/tx7do/go-utils`. Requires Go 1.23+:

```bash
go get github.com/tx7do/go-utils/jwtutil
```

```go
import "github.com/tx7do/go-utils/jwtutil"
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"
	"time"

	_ "github.com/go-kratos/kratos/v2/encoding/json" // registers the "json" codec (required by the generic converters, see below)

	"github.com/golang-jwt/jwt/v5"

	jwtutil "github.com/tx7do/go-utils/jwtutil"
)

func main() {
	secretKey := []byte("secret")

	// Issue: standard claims (exp / iat / iss / aud, ...) are the caller's responsibility
	token, err := jwtutil.GenerateJWT(jwt.MapClaims{
		"sub": "userId",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}, secretKey, jwt.SigningMethodHS256)
	if err != nil {
		log.Fatal(err)
	}

	// Verify: accepts HMAC algorithms only and validates standard claims such as exp
	claims, err := jwtutil.VerifyJWT(token, secretKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("sub:", claims["sub"])

	// Parse the payload without verifying the signature (numbers decode as float64)
	payload, err := jwtutil.ParseJWTPayload(token)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("exp:", payload["exp"]) // float64

	// Map claims straight onto a custom struct (also without verification)
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

### Generic issuing: use a struct as the payload

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

### Short-lived tokens / custom headers / renewal

```go
// Short-lived token: overwrites payload["exp"] with now+duration (mutates the given map)
short, err := jwtutil.GenerateShortLivedJWT(jwt.MapClaims{"sub": "userId"},
	secretKey, jwt.SigningMethodHS256, 15*time.Minute)

// Custom headers (merged on top of the default alg / typ)
token, err := jwtutil.GenerateJWTWithHeader(claims, secretKey, jwt.SigningMethodHS256,
	map[string]interface{}{"kid": "key-id-123"})

// Renewal: rewrites exp and re-signs (does NOT verify the old token's signature; always HS256)
refreshed, err := jwtutil.RefreshJWT(token, secretKey, time.Now().Add(2*time.Hour))
```

### Extracting the token from an HTTP request

```go
token, err := jwtutil.ExtractJWTFromRequest(req) // strict, case-sensitive "Bearer " prefix
```

## Configuration and Conventions

This module has no Options struct; its behavior conventions are:

- **Standard claims are the caller's responsibility**: except for `GenerateShortLivedJWT` (overwrites `exp`) and `RefreshJWT` (rewrites `exp`), the issuing functions never write `exp` / `iat` / `iss` / `aud` for you — put whichever standard claims you need into the claims yourself.
- **Secret key**: every issuing function requires a non-empty `secretKey []byte` (an empty key returns the `secret key cannot be empty` error). A `[]byte` key only works with the HMAC family (HS256/HS384/HS512); passing RS256/ES256 or other non-HMAC algorithms fails at signing time with a `key is of invalid type: ...` error.
- **Signing algorithm**: `signingMethod` is any `jwt.SigningMethod` (e.g. `jwt.SigningMethodHS256`).
- **JSON codec registration (required)**: `ToMapClaims` / `GenerateGenericJWT` / `ParseJWTClaimsToStruct` convert via kratos' `encoding.GetCodec("json")`. Kratos codecs are registered by package `init`; if your program does not (directly or indirectly) import any codec, these three functions panic because the codec is nil. Blank-import one at your entrypoint:

  ```go
  import _ "github.com/go-kratos/kratos/v2/encoding/json"
  ```

- **Number types**: after JSON decoding, every number in `MapClaims` is a `float64` (e.g. `exp: 1672728000` reads back as `float64(1672728000)`).

## API Reference

### Issuing

```go
func GenerateJWT(mapClaims jwt.MapClaims, secretKey []byte, signingMethod jwt.SigningMethod) (string, error)
func GenerateGenericJWT[T any](payload T, secretKey []byte, signingMethod jwt.SigningMethod) (string, error)
func GenerateJWTWithHeader(payload jwt.MapClaims, secretKey []byte, signingMethod jwt.SigningMethod, customHeader map[string]interface{}) (string, error)
func GenerateShortLivedJWT(payload jwt.MapClaims, secretKey []byte, signingMethod jwt.SigningMethod, duration time.Duration) (string, error)
func RefreshJWT(tokenString string, secretKey []byte, newExpiration time.Time) (string, error)
func ToMapClaims[T any](payload T) (jwt.MapClaims, error)
```

- `GenerateJWT`: signs with `jwt.NewWithClaims(signingMethod, mapClaims)` + `SignedString(secretKey)`.
- `GenerateGenericJWT`: converts any payload to `jwt.MapClaims` via `ToMapClaims` first, then signs as above.
- `GenerateJWTWithHeader`: merges `customHeader` entries into the token header (which already contains `alg` and `typ`).
- `GenerateShortLivedJWT`: sets `payload["exp"] = time.Now().Add(duration).Unix()` (overwrites any existing `exp` and mutates the passed map).
- `RefreshJWT`: extracts claims with `ParseJWTPayload` → sets `claims["exp"] = newExpiration.Unix()` → re-signs with `jwt.SigningMethodHS256` (hardcoded). **The old token's signature is not verified.**
- `ToMapClaims`: converts any payload to `jwt.MapClaims` through a JSON codec marshal/unmarshal round trip.

### Verification and Claim-Level Checks

```go
func VerifyJWT(tokenString string, secretKey []byte) (jwt.MapClaims, error)
func ValidateJWTAlgorithm(tokenString string, expectedAlgorithm string) (bool, error)
func ValidateJWTAudience(tokenString string, expectedAudience string) (bool, error)
func ValidateJWTIssuer(tokenString string, expectedIssuer string) (bool, error)
func IsJWTExpired(tokenString string) (bool, error)
```

- `VerifyJWT`: the only full-verification entry point. The key function enforces an HMAC signing method (otherwise `unexpected signing method: %v`), after which jwt/v5 validates the signature and the standard claims (expiry fails with `token has invalid claims: token is expired`). HS256/HS384/HS512 all work.
- `ValidateJWTAlgorithm`: compares the header `alg` against the expected value; missing or non-string reports `invalid or missing algorithm in token header`.
- `ValidateJWTAudience`: `aud` may be a string or an array; returns `true` if it contains the expected value. A missing `aud` yields `(false, nil)`.
- `ValidateJWTIssuer`: returns `true` on exact equality of `iss`. A missing `iss` yields `(false, nil)`.
- `IsJWTExpired`: compares `exp` against the current time (only strictly past expiry counts). **Panics when `exp` is missing** (see Notes).

### Payload and Header Parsing (none verify the signature)

```go
func ParseJWTPayload(tokenString string) (jwt.MapClaims, error)
func ParseJWTClaimsToStruct[T any](tokenString string) (*T, error)
func GetJWTClaims(tokenString string) (map[string]interface{}, error)
func GetJWTHeader(tokenString string) (map[string]interface{}, error)
func GetJWTIssuedAt(tokenString string) (*time.Time, error)
func ExtractJWTFromRequest(r *http.Request) (string, error)
```

- `ParseJWTPayload`: uses a `jwt.Parser` (`WithoutClaimsValidation`) + `ParseUnverified`; parses the format only, no signature check.
- `ParseJWTClaimsToStruct`: same as above, then unmarshals the claims into any type via the JSON codec.
- `GetJWTClaims`: an alias of `ParseJWTPayload` (with the return type widened to `map[string]interface{}`).
- `GetJWTHeader`: returns the full header (`alg`, `typ`, plus custom fields).
- `GetJWTIssuedAt`: reads `iat` and returns `*time.Time`. **Panics when `iat` is missing** (see Notes).
- `ExtractJWTFromRequest`: requires the `Authorization` header to start with the case-sensitive `"Bearer "` prefix; a missing header reports `authorization header is missing`, a wrong prefix (including an empty token) reports `invalid authorization header format`.

### Secure Random Tokens

```go
func NewRefreshToken() (string, error)
func NewJWTId() string
```

- `NewRefreshToken`: reads 32 bytes from `crypto/rand`, encodes them with unpadded Base64URL — always 43 characters from `[A-Za-z0-9_-]`.
- `NewJWTId`: hex encoding of a UUIDv7 — 32 lowercase hex characters without hyphens; falls back to UUIDv4 if the system clock is unreliable.

### Verification Behavior at a Glance

| Function | Verifies signature | Validates standard claims (exp etc.) |
|----------|:--:|:--:|
| `VerifyJWT` | Yes (HMAC only) | Yes (errors when expired) |
| `ParseJWTPayload` / `GetJWTClaims` / `ParseJWTClaimsToStruct` | No | No |
| `GetJWTHeader` / `ValidateJWTAlgorithm` | No | No |
| `ValidateJWTAudience` / `ValidateJWTIssuer` | No | No (reads that claim only) |
| `IsJWTExpired` / `GetJWTIssuedAt` | No | No (reads that claim only) |
| `RefreshJWT` | No | No (rewrites exp, then re-signs) |

## Notes

- **Trust without verification**: except for `VerifyJWT`, every parse/check function in this module skips signature verification (built on `ParseUnverified` / `WithoutClaimsValidation`). Before using `RefreshJWT`, `IsJWTExpired`, `ValidateJWTIssuer`, etc. in security decisions, run `VerifyJWT` yourself first.
- **Missing claims panic**: jwt/v5's `GetExpirationTime` / `GetIssuedAt` return `(nil, nil)` for missing fields, so `IsJWTExpired` panics on a token without `exp` and `GetJWTIssuedAt` panics on a token without `iat` (nil pointer dereference). Make sure the token carries the claim, or recover the panic.
- **`RefreshJWT` neither verifies the old token nor preserves the algorithm**: regardless of the original algorithm (HS384 included), the renewed token is always re-signed with HS256.
- **`[]byte` keys fit HMAC only**: `GenerateJWT` and friends return a `key is of invalid type: ...` error at signing time when given non-HMAC algorithms such as RS256/ES256.
- **Unregistered codec panics**: `ToMapClaims` / `GenerateGenericJWT` / `ParseJWTClaimsToStruct` depend on kratos' JSON codec registration (see Configuration and Conventions).
- `GenerateShortLivedJWT` overwrites the `exp` key of — and thus mutates — the claims map you pass in.
- Numeric claims are `float64` after the JSON round trip; `ParseJWTClaimsToStruct` restores concrete types when mapping onto structs with `json` tags (`exp float64` → `int64`).
- `ExtractJWTFromRequest` performs no validation beyond the prefix check; the extracted token still needs `VerifyJWT`.

## Acknowledgements

- [github.com/golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt/v5) — JWT parsing, signing, and verification
- [github.com/go-kratos/kratos/v2](https://github.com/go-kratos/kratos) — JSON codec registry (`encoding.GetCodec`)
- [github.com/google/uuid](https://github.com/google/uuid) — UUIDv7 (`NewJWTId`)

## License

MIT
