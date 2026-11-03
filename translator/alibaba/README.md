# alibaba

Go translator wrapper for Alibaba Cloud Machine Translation, built on the generic translation API (TranslateGeneral) of the official Alibaba Cloud alimt SDK.

English | [简体中文](README.zh-CN.md)

## Introduction

This module (`github.com/tx7do/go-utils/translator/alibaba`) is the Alibaba Cloud implementation of the translator components in [go-utils](https://github.com/tx7do/go-utils). It is an **independent Go module** (with its own `go.mod`).

Its method signature matches the `Translator` interface of go-utils:

```go
// Translate translates text.
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

Internally it calls the Alibaba Cloud generic machine translation API `TranslateGeneral`, always using `FormatType=text` and `Scene=general`.

## Features

- Built on the official [alimt Go SDK](https://github.com/alibabacloud-go/alimt-20190107) (generic translation API)
- Standard Alibaba Cloud AccessKey authentication
- Configurable region (`WithRegionID`), defaulting to `cn-hangzhou`
- Translates between dozens of languages; common language codes are listed in the table below (the authoritative list is the [Alibaba Cloud documentation](https://help.aliyun.com/document/detail/209537.html))
- All errors are plain `error` values; there is no custom error type

## Installation

This module is an independent Go module:

```bash
go get github.com/tx7do/go-utils/translator/alibaba
```

```go
import "github.com/tx7do/go-utils/translator/alibaba"
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/alibaba"
)

func main() {
	// Create the translator
	translator, err := alibaba.NewTranslator(
		"your_access_key_id",
		"your_access_key_secret",
	)
	if err != nil {
		log.Fatal(err)
	}

	// Translate text
	result, err := translator.Translate("Hello, world!", "en", "zh")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Translation:", result)
}
```

With a custom region:

```go
translator, err := alibaba.NewTranslator(
	"your_access_key_id",
	"your_access_key_secret",
	alibaba.WithRegionID("cn-shanghai"),
)
```

## Configuration / Options

### NewTranslator

```go
func NewTranslator(accessKeyID, accessKeySecret string, opts ...Option) (*Translator, error)
```

**Parameters:**

- `accessKeyID`: Alibaba Cloud AccessKey ID
- `accessKeySecret`: Alibaba Cloud AccessKey Secret
- `opts`: optional configuration options

**Returns:** a `*Translator` instance and an error. An error is returned when the client fails to be created.

### Translate

```go
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

**Parameters:**

- `source`: the text to translate
- `sourceLang`: source language code; pass `auto` for server-side automatic detection
- `targetLang`: target language code

**Returns:** the translated text and an error.

### WithRegionID

```go
func WithRegionID(regionID string) Option
```

Sets the Alibaba Cloud region ID. This is the **only** configuration option of this module; when unset, the default is `cn-hangzhou`. The request endpoint is derived per region as `mt.<regionID>.aliyuncs.com`.

Regions listed in the source comments: `cn-hangzhou` (default), `cn-shanghai`, `cn-beijing`, `ap-southeast-1`, etc. Consult the Alibaba Cloud documentation for the available regions.

## Supported Language Codes

The following is a reference of common language codes; the authoritative list is the [Alibaba Cloud documentation](https://help.aliyun.com/document/detail/209537.html).

### Mainstream languages

| Language | Code | Language | Code |
|----------|------|----------|------|
| Chinese (Simplified) | `zh` | English | `en` |
| Chinese (Traditional) | `zh-TW` | Japanese | `ja` |
| Korean | `ko` | Russian | `ru` |
| German | `de` | French | `fr` |
| Spanish | `es` | Italian | `it` |
| Portuguese | `pt` | Dutch | `nl` |
| Swedish | `sv` | Norwegian | `no` |
| Danish | `da` | Finnish | `fi` |
| Polish | `pl` | Czech | `cs` |
| Romanian | `ro` | Hungarian | `hu` |
| Bulgarian | `bg` | Croatian | `hr` |
| Serbian | `sr` | Ukrainian | `uk` |
| Belarusian | `be` | Greek | `el` |

### Asian languages

| Language | Code | Language | Code |
|----------|------|----------|------|
| Thai | `th` | Vietnamese | `vi` |
| Lao | `lo` | Burmese | `my` |
| Khmer | `km` | Indonesian | `id` |
| Malay | `ms` | Filipino | `tl` |
| Bengali | `bn` | Hindi | `hi` |
| Urdu | `ur` | Punjabi | `pa` |
| Gujarati | `gu` | Telugu | `te` |
| Tamil | `ta` | Malayalam | `ml` |
| Kannada | `kn` | Marathi | `mr` |

### Middle East and Africa

| Language | Code | Language | Code |
|----------|------|----------|------|
| Arabic | `ar` | Hebrew | `he` |
| Persian | `fa` | Turkish | `tr` |
| Kurdish | `ku` | Azerbaijani | `az` |
| Swahili | `sw` | Hausa | `ha` |
| Yoruba | `yo` | Igbo | `ig` |

### European and other languages

| Language | Code | Language | Code |
|----------|------|----------|------|
| Albanian | `sq` | Basque | `eu` |
| Catalan | `ca` | Galician | `gl` |
| Icelandic | `is` | Irish | `ga` |
| Latin | `la` | Luxembourgish | `lb` |
| Maltese | `mt` | Welsh | `cy` |
| Georgian | `ka` | Armenian | `hy` |
| Estonian | `et` | Latvian | `lv` |
| Lithuanian | `lt` | Macedonian | `mk` |
| Mongolian | `mn` | Nepali | `ne` |

## Advanced Usage

### Error Handling

This module has **no** dedicated error type (such as `TranslateError`); every error is a plain `error` wrapped with `fmt.Errorf`. Simply check `err`:

```go
package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/tx7do/go-utils/translator/alibaba"
)

func main() {
	translator, err := alibaba.NewTranslator(
		os.Getenv("ALIBABA_ACCESS_KEY_ID"),
		os.Getenv("ALIBABA_ACCESS_KEY_SECRET"),
	)
	if err != nil {
		log.Fatal(err)
	}

	result, err := translator.Translate("Hello", "en", "zh")
	if err != nil {
		// Network/SDK errors wrap the underlying error; unwrap with
		// errors.Unwrap / errors.Is as needed
		if unwrapped := errors.Unwrap(err); unwrapped != nil {
			log.Printf("underlying error: %v", unwrapped)
		}
		log.Fatal(err)
	}

	fmt.Println(result)
}
```

Error message formats (matching the source code):

| Scenario | Error message format |
|----------|----------------------|
| SDK/network error | `alibaba translate error: <underlying error>` (wrapped with `%w`, can be unwrapped) |
| Service returned an error code | `alibaba translate error: code=<number>, message=<description>` |
| Empty response | `alibaba translate: empty response` |
| Nil translation result | `alibaba translate: translation result is nil` |

For the meaning of Alibaba Cloud server-side error codes, refer to the [Alibaba Cloud API Error Center](https://api.aliyun.com/troubleshoot).

### Caching Results

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/translator/alibaba"
)

var cache = make(map[string]string)

func translateWithCache(translator *alibaba.Translator, text, sourceLang, targetLang string) string {
	key := fmt.Sprintf("%s_%s_%s", text, sourceLang, targetLang)
	if v, ok := cache[key]; ok {
		return v
	}

	result, _ := translator.Translate(text, sourceLang, targetLang)
	cache[key] = result
	return result
}

func main() { _ = translateWithCache }
```

### Batch Translation

```go
package main

import "github.com/tx7do/go-utils/translator/alibaba"

func translateBatch(translator *alibaba.Translator, items []string, sourceLang, targetLang string) []string {
	results := make([]string, len(items))
	for i, item := range items {
		results[i], _ = translator.Translate(item, sourceLang, targetLang)
	}
	return results
}

func main() { _ = translateBatch }
```

### Using the Unified Interface

This module does not import `go-utils/translator`, but its method signature matches the `Translator` interface, so it can be assigned directly:

```go
package main

import (
	"github.com/tx7do/go-utils/translator"
	"github.com/tx7do/go-utils/translator/alibaba"
)

func main() {
	tr, err := alibaba.NewTranslator("your_access_key_id", "your_access_key_secret")
	if err != nil {
		panic(err)
	}
	var iface translator.Translator = tr
	_, _ = iface.Translate("Hello", "en", "zh")
}
```

## Obtaining Credentials

1. Visit the [Alibaba Cloud console](https://console.aliyun.com/) and sign in
2. Create an AccessKey: go to "Resource Access Management (RAM)" → Users → AccessKey management
3. Save the AccessKey ID and AccessKey Secret
4. Activate the [Machine Translation service](https://www.aliyun.com/product/ai/mtranslation) (a free tier is provided; usage beyond it is pay-per-use — see [official pricing](https://www.aliyun.com/price/product))

## Notes

- **No context support**: `Translate` does not accept a `context.Context`, and the module exposes no custom HTTP client or timeout configuration. To apply a timeout, wrap the call externally with a goroutine + `select`, for example:

```go
package main

import (
	"context"
	"time"
)

type result struct {
	text string
	err  error
}

// translateWithTimeout applies an external timeout to a single translation call
func translateWithTimeout(t interface {
	Translate(source, sourceLang, targetLang string) (string, error)
}, text, from, to string, timeout time.Duration) (string, error) {
	ch := make(chan result, 1)
	go func() {
		text, err := t.Translate(text, from, to)
		ch <- result{text, err}
	}()

	select {
	case r := <-ch:
		return r.text, r.err
	case <-time.After(timeout):
		return "", context.DeadlineExceeded // the translation goroutine keeps running in the background
	}
}

func main() { _ = translateWithTimeout }
```

> Note: the in-flight request cannot actually be cancelled; this only makes the caller return on time.

- The language code table is a reference of common languages; the authoritative list is the Alibaba Cloud documentation. When unsure, pass `auto` as the source language
- Every call hits the Alibaba Cloud API; watch your rate limits and quota. The `Translator` instance holds the SDK client internally and is safe to reuse — do not rebuild it per call
- This module always uses the general scene (`Scene=general`) and plain text format (`FormatType=text`); for domain-specific models (e-commerce, medical, etc.), use the alimt SDK directly

## Dependencies

Matching `go.mod`:

```
github.com/alibabacloud-go/alimt-20190107 v1.0.0
github.com/alibabacloud-go/darabonba-openapi/v2 v2.1.15
github.com/alibabacloud-go/tea v1.4.0
```

## Acknowledgements

- [Alibaba Cloud Machine Translation](https://www.aliyun.com/product/ai/mtranslation)
- [alibabacloud-go/alimt-20190107 Go SDK](https://github.com/alibabacloud-go/alimt-20190107)

## Related Links

- [Alibaba Cloud Machine Translation documentation](https://help.aliyun.com/product/2477301.html)
- [Alibaba Cloud console](https://console.aliyun.com/)
- [Project home](https://github.com/tx7do/go-utils)

## License

Released under the project [MIT License](https://github.com/tx7do/go-utils/blob/main/LICENSE).
