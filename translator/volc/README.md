# volc

Go translator wrapper for Volcengine Translate, built on the Volcengine TranslateText API (HMAC-SHA256 signing).

English | [简体中文](README.zh-CN.md)

## Introduction

This module (`github.com/tx7do/go-utils/translator/volc`) is the Volcengine implementation of the translator components in [go-utils](https://github.com/tx7do/go-utils). It is an **independent Go module** (with its own `go.mod`) with no third-party runtime dependencies.

It provides single-text and batch translation, and its method signature matches the `Translator` interface of go-utils:

```go
// Translate translates a single text.
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

Internally it calls the Volcengine `TranslateText` API (`Action=TranslateText&Version=2020-06-01`), handling the V4-style HMAC-SHA256 request signing automatically; the translation scene is always `general`.

## Features

- Single-text translation (`Translate`) and batch translation (`TranslateBatch`)
- Automatic Volcengine HMAC-SHA256 request signing (body hash, canonical request, derived signing key)
- Configurable region (`WithRegion`), defaulting to `cn-north-1`
- Custom HTTP client (`WithHTTPClient`)
- No custom error type; errors are plain `error` values

## Installation

This module is an independent Go module:

```bash
go get github.com/tx7do/go-utils/translator/volc
```

```go
import "github.com/tx7do/go-utils/translator/volc"
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/volc"
)

func main() {
	// Create the translator; an AccessKey and SecretKey are required.
	// NewTranslator returns (*Translator, error) — the error must be handled.
	translator, err := volc.NewTranslator("your_access_key", "your_secret_key")
	if err != nil {
		log.Fatal(err)
	}

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
func NewTranslator(accessKey, secretKey string, opts ...Option) (*Translator, error)
```

**Parameters:**

- `accessKey`: the Volcengine AccessKey
- `secretKey`: the Volcengine SecretKey (trailing line breaks are stripped automatically)
- `opts`: optional configuration options

**Returns:** a `*Translator` instance and an error. An error is returned when `accessKey` or `secretKey` is empty.

Default behavior: region `cn-north-1`; HTTP client with a 30-second timeout (replaceable via `WithHTTPClient`).

### Translate

```go
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

**Parameters:**

- `source`: the text to translate (empty input returns the error `待翻译文本不能为空`)
- `sourceLang`: source language code
- `targetLang`: target language code

**Returns:** the translated text and an error.

### TranslateBatch

```go
func (t *Translator) TranslateBatch(sources []string, sourceLang, targetLang string) ([]string, error)
```

Batch translation: translates multiple texts in one request.

**Parameters:**

- `sources`: texts to translate (an empty slice returns an empty result immediately)
- `sourceLang`: source language code
- `targetLang`: target language code

**Returns:** the list of translations and an error. If the number of returned items differs from the input, an error is returned (format: `结果数量不匹配：输入%d条，返回%d条`).

### Options

#### WithRegion

```go
func WithRegion(region string) Option
```

Sets the Volcengine region. The default is `cn-north-1` (per the implementation — see Notes below). The region is part of the signing credential scope and request context.

#### WithHTTPClient

```go
func WithHTTPClient(client *http.Client) Option
```

Sets a custom HTTP client, useful for replacing the default 30-second timeout or configuring proxies and transport settings.

## Obtaining the AccessKey and SecretKey

1. Visit the [Volcengine console](https://console.volcengine.com/)
2. Create or view an AccessKey (AK) and SecretKey (SK) under key management
3. Make sure the [Volcengine Translate service](https://www.volcengine.com/product/translate) is activated
4. Pass the credentials to `NewTranslator`

## Supported Languages

| Language | Code |
|----------|------|
| Chinese | zh |
| English | en |
| Japanese | ja |
| Korean | ko |
| Russian | ru |
| German | de |
| French | fr |
| Spanish | es |
| Portuguese | pt |
| Italian | it |
| Thai | th |
| Vietnamese | vi |
| Indonesian | id |
| Burmese | my |
| Khmer | km |
| Lao | lo |
| Arabic | ar |
| Turkish | tr |
| Polish | pl |
| Ukrainian | uk |

See the [official documentation](https://www.volcengine.com/docs/4640/35107) for more supported languages.

## Advanced Usage

### Batch Translation

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/volc"
)

func main() {
	translator, err := volc.NewTranslator(
		"your_access_key",
		"your_secret_key",
	)
	if err != nil {
		log.Fatal(err)
	}

	results, err := translator.TranslateBatch(
		[]string{"火山翻译", "Go 客户端"},
		"zh", "en",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(results)
}
```

### Custom Region and HTTP Client

```go
package main

import (
	"net/http"
	"time"

	"github.com/tx7do/go-utils/translator/volc"
)

func main() {
	client := &http.Client{Timeout: 10 * time.Second}

	translator, err := volc.NewTranslator(
		"your_access_key",
		"your_secret_key",
		volc.WithRegion("cn-north-1"),
		volc.WithHTTPClient(client),
	)
	if err != nil {
		panic(err)
	}
	_ = translator
}
```

### Error Handling

This module has no dedicated error type; every error is a plain `error` created with `fmt.Errorf`. Simply check `err`:

```go
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/tx7do/go-utils/translator/volc"
)

func main() {
	translator, err := volc.NewTranslator("your_access_key", "your_secret_key")
	if err != nil {
		log.Fatal(err)
	}

	result, err := translator.Translate("Hello", "en", "zh")
	if err != nil {
		// err.Error() looks like `API错误: InvalidAccessKey - The access key is invalid`;
		// parse the error Code before the "-" if needed
		if strings.Contains(err.Error(), "API错误") {
			log.Printf("volc API error: %v", err)
		}
		return
	}

	fmt.Println(result)
}
```

Error message formats (matching the source code; the prefixes are Chinese):

| Scenario | Error message format |
|----------|----------------------|
| API returned an error | `API错误: <Code> - <Message>` |
| HTTP status is not 200 | `HTTP错误: <status code>, 响应: <response body>` |
| Single-text call failed | `翻译失败: <underlying error>` |
| Batch call failed | `批量翻译失败: <underlying error>` |
| Source text is empty | `待翻译文本不能为空` |
| Translation result is empty | `翻译结果为空` |
| Result count mismatch | `结果数量不匹配：输入%d条，返回%d条` |
| Network request failed | `发送请求失败: <underlying error>` |
| Response parsing failed | `解析响应失败: <underlying error>, 响应: <response body>` |

### Using the Unified Interface

This module does not import `go-utils/translator`, but its `Translate` method signature matches the `Translator` interface, so it can be assigned directly:

```go
package main

import (
	"github.com/tx7do/go-utils/translator"
	"github.com/tx7do/go-utils/translator/volc"
)

func main() {
	tr, err := volc.NewTranslator("your_access_key", "your_secret_key")
	if err != nil {
		panic(err)
	}
	var iface translator.Translator = tr
	_, _ = iface.Translate("Hello", "en", "zh")
}
```

## Notes

1. **The default region follows the implementation**: the default region is `cn-north-1` (initialized in `translator.go`). Note that the `WithRegion` comment in `options.go` says "cn-beijing (default)", which contradicts the implementation — **the implementation is authoritative** (the tests also assert `cn-north-1` as the default).
2. **Debug logging cannot be turned off**: every API call prints the request URL, the `Authorization` signature header, `X-Date`, and the request body to standard output via `fmt.Printf`. The `Authorization` header contains the request signature (derived from the AK; it does not contain the plaintext SK), but the output is verbose and may not be welcome in production logs — the current version has no switch to disable it. Be aware of this before deploying to production.
3. **Exported debug helpers**:
   - `String()`: returns a masked instance description (region + first/last-truncated AccessKey); implements `fmt.Stringer`;
   - `VerifySignature(testTimestamp, testDate, testBody string, expectedSignature string)`: recalculates the signature from the given parameters and compares it against the expected value, printing the outcome via `fmt.Printf`; useful for debugging signature issues (used in the tests).
4. **Fixed API endpoint and scene**: the endpoint is always `translate.volcengineapi.com`, with `Action=TranslateText`, `Version=2020-06-01`, and `Scene=general`; for other scenes or endpoints, use the official Volcengine SDK directly.
5. Every call hits the Volcengine API; watch your quota and rate limits. The `Translator` instance is reusable.

## Acknowledgements

- [Volcengine Translate](https://www.volcengine.com/product/translate)

## License

Released under the project [MIT License](https://github.com/tx7do/go-utils/blob/main/LICENSE).
