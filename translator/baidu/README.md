# baidu

Go translator wrapper for the Baidu Translate Open Platform general translation API.

English | [简体中文](README.zh-CN.md)

## Introduction

This module (`github.com/tx7do/go-utils/translator/baidu`) is the Baidu Translate implementation of the translator components in [go-utils](https://github.com/tx7do/go-utils). It is an **independent Go module** (with its own `go.mod`) and has no third-party dependencies.

Its method signature matches the `Translator` interface of go-utils:

```go
// Translate translates text.
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

Internally it calls the Baidu general translation API (`https://fanyi-api.baidu.com/api/trans/vip/translate`), generating a random `salt` and computing the MD5 signature automatically.

## Features

- Built on the Baidu Translate Open Platform general translation API
- AppID + SecretKey authentication with the signature computed automatically (MD5)
- Built-in 30-second HTTP timeout
- `auto` source language for automatic detection
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

	// Translate text
	result, err := translator.Translate("Hello, world!", "en", "zh")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Translation result:", result)
}
```

## Configuration / Options

### NewTranslator

```go
func NewTranslator(appID, secretKey string) *Translator
```

**Parameters:**

- `appID`: the Baidu Translate AppID
- `secretKey`: the Baidu Translate SecretKey

**Returns:** a `*Translator` instance. It does **not** return an error — construction performs no network requests.

This module has no configuration options: the HTTP client uses a fixed 30-second timeout that cannot be changed.

### Translate

```go
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

**Parameters:**

- `source`: the text to translate
- `sourceLang`: source language code; pass `auto` for automatic detection
- `targetLang`: target language code

**Returns:** the translated text and an error.

## Obtaining the AppID and SecretKey

1. Visit the [Baidu Translate Open Platform](https://fanyi-api.baidu.com/)
2. Sign in or register an account
3. Create a new application in the management console
4. Obtain the AppID and SecretKey

## Supported Languages

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

See the [official documentation](https://fanyi-api.baidu.com/api/trans/product/apidoc) for more supported languages.

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
		// err.Error() looks like "baidu translate error: 54001 - Invalid sign";
		// parse the numeric code before the "-" if needed
		log.Printf("translate failed: %v", err)
		return
	}

	fmt.Println(result)
}
```

Error message formats (matching the source code):

| Scenario | Error message format |
|----------|----------------------|
| API returned an error code | `baidu translate error: <error_code> - <error_msg>` |
| HTTP status is not 200 | `baidu translate error: status <status code>, body: <response body>` |
| No translation result in the response | `baidu translate: invalid response format` |
| Network request failed | the underlying error (e.g. `*url.Error`) is returned as-is |

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

## Notes

1. **Long texts return only the first segment**: the Baidu API splits long text by line and returns multiple segments in the `trans_result` array; this module returns only the first segment's translation (`trans_result[0].dst`). Split long texts yourself before calling.
2. **HTTP timeout is fixed at 30 seconds and is not configurable**: the module uses `&http.Client{Timeout: 30 * time.Second}` internally and offers no customization option.
3. **Signature algorithm**: `sign = MD5(appID + query + salt + secretKey)`, where `salt` is a randomly generated `int64` per request. The signature is computed automatically; no manual work is needed.
4. Per-request text length and rate limits are documented in the [official documentation](https://fanyi-api.baidu.com/api/trans/product/apidoc); split long texts and mind the QPS limits.
5. Every call hits the Baidu API; watch your account quota and billing. The `Translator` instance is reusable.

## Acknowledgements

- [Baidu Translate Open Platform](https://fanyi-api.baidu.com/)

## License

Released under the project [MIT License](https://github.com/tx7do/go-utils/blob/main/LICENSE).
