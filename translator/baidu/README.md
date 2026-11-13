# baidu

Go translator wrapper for the Baidu Translate Open Platform general translation API.

English | [简体中文](README.zh-CN.md)

## Introduction

This module (`github.com/tx7do/go-utils/translator/baidu`) is the Baidu Translate implementation of the translator components in [go-utils](https://github.com/tx7do/go-utils). It is an **independent Go module** (with its own `go.mod`) and has no third-party runtime dependencies (testify is used in tests only).

Its method signature matches the `Translator` interface of go-utils:

```go
// Translate translates text.
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

Internally it calls the Baidu general translation API (`https://fanyi-api.baidu.com/api/trans/vip/translate`), generating a random `salt` and computing the MD5 signature automatically.

## Features

- Built on the Baidu Translate Open Platform general translation API
- AppID + SecretKey authentication with the signature computed automatically (MD5)
- **BCP47 language normalization**: pass `zh-CN` / `en-US` / `ar` … — `NormalizeLang` maps them to Baidu language codes (`zh` / `en` / `ara` …); unknown codes fall back to the lowercased language part without the region suffix
- **Multi-segment results joined**: the Baidu API splits text by line and returns multiple segments; this module joins them back with `\n`
- Built-in 30-second request timeout, 10 MiB response size cap
- Empty input short-circuits to `("", nil)`; missing credentials fail fast with a clear error
- Configurable endpoint (`WithEndpoint`) and HTTP client (`WithHTTPClient`) — the endpoint option is handy for injecting a mock server in tests
- No custom error type; errors are plain `error` values

## Installation

This module is an independent Go module:

```bash
go get github.com/tx7do/go-utils/translator/baidu
```

```go
import "github.com/tx7do/go-utils/translator/baidu"
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/baidu"
)

func main() {
	// Create the translator; an AppID and SecretKey are required.
	// Note: NewTranslator does not return an error.
	translator := baidu.NewTranslator("your_app_id", "your_secret_key")

	// Translate text; language tags may be BCP47 ("en-US") or Baidu codes ("en")
	result, err := translator.Translate("Hello, world!", "en-US", "zh-CN")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Translation result:", result)
}
```

## Configuration / Options

### NewTranslator

```go
func NewTranslator(appID, secretKey string, opts ...Option) *Translator
```

**Parameters:**

- `appID`: the Baidu Translate AppID
- `secretKey`: the Baidu Translate SecretKey
- `opts`: optional configuration (see below)

**Returns:** a `*Translator` instance. It does **not** return an error — construction performs no network requests, and credential presence is checked per `Translate` call instead.

### Options

| Option | Signature | Description |
|--------|-----------|-------------|
| `WithEndpoint` | `WithEndpoint(endpoint string)` | Override the API endpoint; inject a mock server in tests |
| `WithHTTPClient` | `WithHTTPClient(c *http.Client)` | Replace the HTTP client (custom proxy / timeout policy). The default is a 30-second-timeout client |

### NormalizeLang

```go
func NormalizeLang(lang string) string
```

Maps a BCP47 language tag to a Baidu language code (`zh-CN` → `zh`, `zh-TW` → `cht`, `ar` → `ara`, …). Unlisted tags are passed through as the lowercased language part without the region suffix (`xxx-YY` → `xxx`). `Translate` applies it to both language arguments automatically; call it yourself only if you need the mapping elsewhere (e.g. for display).

### Translate

```go
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

**Parameters:**

- `source`: the text to translate; whitespace-only input returns `("", nil)` without hitting the API
- `sourceLang`: source language; `auto`, empty string, or any BCP47 tag not resolvable to a concrete language triggers automatic detection
- `targetLang`: target language

**Returns:** the translated text and an error.

## Obtaining the AppID and SecretKey

1. Visit the [Baidu Translate Open Platform](https://fanyi-api.baidu.com/)
2. Sign in or register an account
3. Create a new application in the management console
4. Obtain the AppID and SecretKey

## Supported Languages

Inputs accept BCP47 tags (normalized automatically); the table below lists the resulting Baidu codes.

| Language | Code | Language | Code |
|----------|------|----------|------|
| Auto-detect | `auto` | Slovak | `sk` |
| Chinese | `zh` | Croatian | `hr` |
| English | `en` | Serbian | `sr` |
| Japanese | `ja` | Estonian | `et` |
| Korean | `ko` | Lithuanian | `lt` |
| Russian | `ru` | Latvian | `lv` |
| German | `de` | Albanian | `sq` |
| French | `fr` | Macedonian | `mk` |
| Spanish | `es` | Armenian | `hy` |
| Portuguese | `pt` | Georgian | `ka` |
| Italian | `it` | Belarusian | `be` |
| Thai | `th` | Icelandic | `is` |
| Vietnamese | `vi` | Kurdish | `ku` |
| Indonesian | `id` | Kyrgyz | `ky` |
| Burmese | `my` | Latin | `la` |
| Khmer | `km` | Luxembourgish | `lb` |
| Lao | `lo` | Sanskrit | `sa` |
| Arabic | `ar` | Welsh | `cy` |
| Turkish | `tr` | Nepali | `ne` |
| Polish | `pl` | Bengali | `bn` |
| Ukrainian | `uk` | Punjabi | `pa` |
| Greek | `el` | Urdu | `ur` |
| Swedish | `sv` | Malay | `ms` |
| Danish | `da` | Telugu | `te` |
| Finnish | `fi` | Tamil | `ta` |
| Dutch | `nl` | Kannada | `kn` |
| Norwegian | `no` | Malayalam | `ml` |
| Czech | `cs` | Gujarati | `gu` |
| Romanian | `ro` | Marathi | `mr` |
| Hungarian | `hu` | Odia | `or` |
| Bulgarian | `bg` | Sindhi | `sd` |
| Slovenian | `sl` | Sinhala | `si` |

Traditional Chinese (`zh-TW` / `zh-HK`) maps to `cht`. See the [official documentation](https://fanyi-api.baidu.com/api/trans/product/apidoc) for more supported languages.

## Advanced Usage

### Error Handling

This module has no dedicated error type; every error is a plain `error` created with `fmt.Errorf`. Simply check `err`:

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/baidu"
)

func main() {
	translator := baidu.NewTranslator("your_app_id", "your_secret_key")

	result, err := translator.Translate("Hello", "en", "zh")
	if err != nil {
		// err.Error() looks like "baidu translator: api error 54001: Invalid sign";
		// parse the numeric code after "api error" if needed
		log.Printf("translate failed: %v", err)
		return
	}

	fmt.Println(result)
}
```

Error message formats (matching the source code):

| Scenario | Error message format |
|----------|----------------------|
| Missing credentials | `baidu translator: appid/secret is not configured` |
| API returned an error code | `baidu translator: api error <error_code>: <error_msg>` |
| HTTP status is not 200 | `baidu translator: unexpected status <status>: <response body>` |
| Response decoding failed | `baidu translator: decode response: <underlying error>` |
| Network request failed | `baidu translator: request: <underlying error>` |

An API response with an empty `trans_result` array returns `("", nil)` — no error.

### Common Error Codes

The following are common error codes; the full list is in the [official documentation](https://fanyi-api.baidu.com/api/trans/product/apidoc):

| Code | Meaning |
|------|---------|
| 52000 | Success |
| 52001 | Request timed out |
| 52002 | System error |
| 52003 | Unauthorized user |
| 54000 | A required parameter is empty |
| 54001 | Invalid signature |
| 54002 | Rate limit exceeded |
| 54003 | Invalid request |
| 54004 | Unknown field |
| 54005 | Unknown translation type |
| 58000 | Invalid client IP |
| 58001 | Unsupported target language direction |
| 90107 | Authentication failed |

### Using the Unified Interface

This module does not import `go-utils/translator`, but its method signature matches the `Translator` interface, so it can be assigned directly:

```go
package main

import (
	"github.com/tx7do/go-utils/translator"
	"github.com/tx7do/go-utils/translator/baidu"
)

func main() {
	tr := baidu.NewTranslator("your_app_id", "your_secret_key")
	var iface translator.Translator = tr
	_, _ = iface.Translate("Hello", "en", "zh")
}
```

### Testing with a Mock Endpoint

```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = json.NewEncoder(w).Encode(map[string]any{
		"trans_result": []map[string]string{{"src": "hello", "dst": "你好"}},
	})
}))

tr := baidu.NewTranslator("appid", "secret", baidu.WithEndpoint(srv.URL))
```

## Notes

1. **Signature algorithm**: `sign = MD5(appID + query + salt + secretKey)`, where `salt` is a random `uint64` per request. The signature is computed automatically; no manual work is needed.
2. **Long texts**: the Baidu API splits text by line and returns multiple segments; this module joins them with `\n` and returns the whole translation. Mind the per-request length and QPS limits in the [official documentation](https://fanyi-api.baidu.com/api/trans/product/apidoc).
3. **Timeouts**: the default HTTP client and the per-request context are both capped at 30 seconds; replace the client via `WithHTTPClient` if you need a different policy.
4. The response body is read up to 10 MiB; larger responses fail with a decode error rather than exhausting memory.
5. Every call hits the Baidu API; watch your account quota and billing. The `Translator` instance is reusable.

## Acknowledgements

- [Baidu Translate Open Platform](https://fanyi-api.baidu.com/)

## License

Released under the project [MIT License](https://github.com/tx7do/go-utils/blob/main/LICENSE).
