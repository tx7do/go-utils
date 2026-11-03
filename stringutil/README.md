# stringutil

String utility package: type conversion, pseudo-random / cryptographically secure
random string generation, and JSON field value replacement.

English | [简体中文](README.zh-CN.md)

## Features

- **Type conversion**: between Go basic types and strings (pointer variants and
  default-value variants included)
- **Pseudo-random strings**: based on the PCG generator of `math/rand/v2`, suitable
  for general business scenarios
- **Crypto-random strings**: based on `crypto/rand`, suitable for passwords, tokens, keys
- **JSON field replacement**: replaces the values of given fields in a JSON string via regex
- **Reproducible randomness**: `RandomSeed` accepts an external `*rand.Rand`, so a
  fixed seed reproduces the output

## Installation

stringutil is a subpackage of the root module `github.com/tx7do/go-utils`
(no go.mod of its own); installing the root module is enough:

```bash
go get github.com/tx7do/go-utils
```

```go
import "github.com/tx7do/go-utils/stringutil"
```

> Version requirement: the code uses `math/rand/v2` (introduced in Go 1.22), so
> Go 1.22+ is required in practice; the root module's `go.mod` currently declares
> `go 1.25.0`.

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/stringutil"
)

func main() {
	// Type conversion.
	s := stringutil.IntToString(42)
	n, _ := stringutil.StringToInt("42")
	fmt.Println(s, n)

	// Pseudo-random strings.
	alpha, _ := stringutil.RandomAlphabetic(10)
	num, _ := stringutil.RandomNumeric(6)
	alphanum, _ := stringutil.RandomAlphaNumeric(16)
	fmt.Println(alpha, num, alphanum)

	// Cryptographically secure random string.
	safe, _ := stringutil.CryptoRandomAlphaNumeric(32)
	fmt.Println(safe)

	// JSON field replacement.
	jsonStr := `{"tenantId": "old_value", "name": "test"}`
	result := stringutil.ReplaceJSONField("tenantId|tenant_id", "new_value", jsonStr)
	fmt.Println(result)
}
```

## API Reference

### Type conversion (convert.go)

#### Basic types → string

| Function | Description |
|----------|-------------|
| `BoolToString(v bool)` | `true` → `"true"` |
| `IntToString(v int)` | Integer to decimal string |
| `Int8ToString` / `Int16ToString` / `Int32ToString` / `Int64ToString` | Signed integers of each size |
| `UintToString` / `Uint8ToString` / `Uint16ToString` / `Uint32ToString` / `Uint64ToString` | Unsigned integers of each size |
| `Float32ToString` / `Float64ToString` | Floats to string (`DoubleToString` is an alias of `Float64ToString`) |

#### String → basic types

All functions `TrimSpace` automatically and return an `error` on parse failure.

| Function | Description |
|----------|-------------|
| `StringToBool(s)` | Based on `strconv.ParseBool`; accepted values: `1`, `t`, `T`, `TRUE`, `true`, `True` → `true`; `0`, `f`, `F`, `FALSE`, `false`, `False` → `false` (anything else errors) |
| `StringToInt(s)` / `StringToIntBase(s, base)` | Decimal or a given base (base 2-36, 0 infers from prefix) |
| `StringToInt8` ~ `StringToInt64` | Signed integers of each size (`*Base` variants included) |
| `StringToUint` ~ `StringToUint64` | Unsigned integers of each size (`*Base` variants included) |
| `StringToFloat32` / `StringToFloat64` | Float parsing (`StringToDouble` is an alias of `StringToFloat64`) |

#### Default-value variants

Return `defaultValue` instead of `error` on parse failure:

```go
v := stringutil.StringToIntOrDefault("abc", 0)   // v == 0
v2 := stringutil.StringToIntOrDefault("123", 0)  // v2 == 123
```

Available functions: `StringToBoolOrDefault`, `StringToIntOrDefault`,
`StringToUintOrDefault`, `StringToDoubleOrDefault` (semantically the same as
`StringToFloat64OrDefault`), and so on, plus the corresponding `*BaseOrDefault` variants.

#### Pointer conversions

Conversions between `*T` and `*string`; a `nil` input yields `nil`:

```go
n := 42
sp := stringutil.IntPtrToStringPtr(&n) // *string

s := "100"
np := stringutil.StringPtrToIntPtr(&s) // *int, nil on parse failure
```

| Function family | Description |
|-----------------|-------------|
| `BoolPtrToStringPtr` ~ `DoublePtrToStringPtr` | `*T` → `*string` |
| `StringPtrToBoolPtr` ~ `StringPtrToDoublePtr` | `*string` → `*T` (`*Base` variants included) |

### Pseudo-random strings (randomstringutils.go)

Based on the PCG pseudo-random generator of `math/rand/v2`, suitable for general
business scenarios.

The package-level random source (seeded with the current Unix-nano timestamp; all
`Random*` functions use it):

```go
var RANDOM = rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))
```

#### Convenience functions

| Function | Charset | Example output |
|----------|---------|----------------|
| `RandomAlphabetic(count)` | a-z, A-Z | `KxTmBqRp` |
| `RandomNumeric(count)` | 0-9 | `83472615` |
| `RandomAlphaNumeric(count)` | a-z, A-Z, 0-9 | `aB3kR9xM` |
| `RandomAscii(count)` | ASCII 32-126 | `H_I;E` |
| `RandomNonAlphaNumeric(count)` | no filtering | may contain arbitrary Unicode characters (incl. surrogate ranges) |
| `RandomAlphaNumericCustom(count, letters, numbers)` | custom filtering | — |

#### General functions

```go
import "math/rand/v2"

// Using a predefined charset.
out, err := stringutil.Random(10, 0, 0, false, false, 'a', 'b', 'c', '1', '2', '3')

// Reproducible random string (fixed seed): import math/rand/v2 yourself.
random := rand.New(rand.NewPCG(42, 0))
out, err := stringutil.RandomSeed(10, 0, 0, true, true, nil, random)
```

Parameters of `Random` / `RandomSeed`: `count` is the output length; `start` / `end`
delimit the character range (end-exclusive, both ignored when both are 0 and `chars`
is nil); `letters` / `numbers` filter characters; `chars` is an optional custom
charset (when provided, `start`/`end` index into that slice). When
`start == 0 && end == 0` and no `chars` is given, the default range is `[0, MaxInt32)`
with no filtering, or `[' ', 'z'+1)` with any filtering enabled.

#### Error conditions

`RandomSeed` returns an `error` when:

- `count < 0`: a negative length was requested
- `chars` is non-`nil` but empty: empty charset
- `end <= start` (when `start`/`end` are explicitly provided)
- `end > len(chars)` (when `chars` is provided)

### Crypto-secure random strings (cryptorandomstringutils.go)

Based on the `crypto/rand` cryptographically secure generator, suitable for
passwords, tokens, keys and other security-sensitive scenarios.

The API mirrors the pseudo-random version; function names carry the `Crypto` prefix:

| Function | Charset |
|----------|---------|
| `CryptoRandomAlphabetic(count)` | a-z, A-Z |
| `CryptoRandomNumeric(count)` | 0-9 |
| `CryptoRandomAlphaNumeric(count)` | a-z, A-Z, 0-9 |
| `CryptoRandomAscii(count)` | ASCII 32-126 |
| `CryptoRandomNonAlphaNumeric(count)` | no filtering |
| `CryptoRandomAlphaNumericCustom(count, letters, numbers)` | custom filtering |
| `CryptoRandom(count, start, end, letters, numbers, chars...)` | general interface |

Error conditions are the same as the pseudo-random version (except for entropy
failures, see Notes).

### JSON field replacement (replace.go)

```go
func ReplaceJSONField(fieldNames, newValue, jsonStr string) string
```

Replaces the values of given fields in a JSON string using a regular expression.

- `fieldNames`: the field names to replace, separated by `|` (e.g. `"tenantId|tenant_id"`)
- `newValue`: the new value (it will be wrapped in quotes)
- Matching is case-insensitive

```go
jsonStr := `{"tenantId": "old", "Name": "test"}`
result := stringutil.ReplaceJSONField("tenantId|tenant_id", "new", jsonStr)
// result: `{"tenantId": "new", "Name": "test"}`
```

## Recommendations

- General business randomness: use the `Random*` functions (better performance)
- Passwords, tokens, keys and other security-sensitive scenarios: use the `CryptoRandom*` functions
- Reproducible random output: use `RandomSeed` with a fixed-seed `*rand.Rand`
- Fault-tolerant type conversion: prefer the `*OrDefault` functions

## Notes

- **The `CryptoRandom*` family panics if the entropy source fails**: internally
  `getCryptoRandomInt` calls `rand.Int(rand.Reader, ...)` and panics on error.
  Failure to read system entropy via `crypto/rand` is extremely rare, but it means
  the error is not returned as an `error` value.
- The package-level `RANDOM` is a globally shared `*rand.Rand`; concurrent calls to
  the `Random*` functions share the same source. `*rand.Rand` of `math/rand/v2` is
  **not safe for concurrent use**, so concurrent calls carry a data race risk.
  Under high concurrency, or when independent sequences are needed, use `RandomSeed`
  with your own (locked/isolated) random source instead.
- `RandomNonAlphaNumeric` / `CryptoRandomNonAlphaNumeric` apply no character
  filtering, so the output may contain arbitrary Unicode characters (including some
  surrogate ranges); verify charset safety before writing to logs or storage.
- `ReplaceJSONField` is regex-based and only fits JSON whose field values are plain
  strings; for nested structures or complex escapes, use a standard JSON library.
