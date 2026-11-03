# google

Go translator wrapper for Google Translate, integrating three versions of the API: V1 (free unofficial endpoint), V2, and V3 (official Cloud Translation API).

English | [简体中文](README.zh-CN.md)

## Introduction

This module (`github.com/tx7do/go-utils/translator/google`) is the Google Translate implementation of the translator components in [go-utils](https://github.com/tx7do/go-utils). It is an **independent Go module** (with its own `go.mod`).

It provides a unified `Translate` entry point that dispatches to V1 / V2 / V3 based on the configured version, and the versioned methods can also be called directly. Its method signature matches the `Translator` interface of go-utils:

```go
// Translate translates text.
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

## Features

- **Three API versions**: V1 free unofficial endpoint; V2 and V3 are official APIs requiring an API Key
- **Unified entry point**: `Translate` dispatches automatically according to `WithVersion`
- **Direct versioned calls**: `TranslateV1` / `TranslateV2` / `TranslateV3`
- **Automatic source-language detection**: V1 and V3 accept `auto`; V2 always detects automatically and returns the detected language
- **Matches the `Translator` interface shape**: `Translate(source, sourceLang, targetLang string) (string, error)`

## Installation

This module is an independent Go module:

```bash
go get github.com/tx7do/go-utils/translator/google
```

```go
import "github.com/tx7do/go-utils/translator/google"
```

## Quick Start

### Basic usage (V1, free, no credentials)

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/google"
)

func main() {
	// NewTranslator(opts ...Option) *Translator; it does not return an error.
	// Without options it defaults to V1.
	translator := google.NewTranslator()

	// "auto" as the source language enables automatic detection
	result, err := translator.Translate("Hello, world!", "auto", "zh-CN")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Translation:", result)
}
```

### Using the V2 API (official API, requires an API Key)

```go
translator := google.NewTranslator(
	google.WithVersion("v2"),
	google.WithApiKey("your_google_api_key"),
)

// Translate returns (string, error)
result, err := translator.Translate("Hello", "en", "zh-CN")
if err != nil {
	log.Fatal(err)
}
fmt.Println("Translation:", result)
```

### Using the V3 API (official API, requires an API Key)

```go
translator := google.NewTranslator(
	google.WithVersion("v3"),
	google.WithApiKey("your_google_api_key"),
)

result, err := translator.Translate("Hello", "en", "zh-CN")
if err != nil {
	log.Fatal(err)
}
fmt.Println("Translation:", result)
```

## Configuration / Options

### NewTranslator

```go
func NewTranslator(opts ...Option) *Translator
```

**Parameters:** `opts`: optional configuration options.

**Returns:** a `*Translator` instance; it does **not** return an error.

### Translate

```go
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

Dispatches according to the version set with `WithVersion`:

- `"v1"` or empty (default) → `TranslateV1`
- `"v2"` → `TranslateV2` (discards the detected language and returns only the text)
- `"v3"` → `TranslateV3`

### WithVersion

```go
func WithVersion(version string) Option
```

Sets the API version to use: `"v1"` (default), `"v2"`, `"v3"`.

### WithApiKey

```go
func WithApiKey(key string) Option
```

Sets the Google API Key, required for V2 / V3; V1 does not need one.

## Advanced Usage

### Calling a Specific Version Directly

Besides `Translate`, all three versioned methods can be called directly:

```go
func (t *Translator) TranslateV1(source, sourceLang, targetLang string) (string, error)

func (t *Translator) TranslateV2(source, sourceLang, targetLang string) (string, language.Tag, error)

func (t *Translator) TranslateV3(source, sourceLang, targetLang string) (string, error)
```

- **TranslateV1**: requests the free unofficial endpoint (`translate.googleapis.com/translate_a/single`); no authentication required.
- **TranslateV2**: returns a **three-value tuple** `(text, detected source language, error)` where the second value is a `golang.org/x/text/language.Tag`. The `sourceLang` parameter is **ignored** — V2 always detects the source language automatically. `targetLang` is parsed with `language.Parse` and must be a valid BCP-47 language tag (such as `zh-CN` or `es`); a parse failure returns an error.
- **TranslateV3**: calls the official Cloud Translation API v3 (`TranslateText`), always using the `text/plain` MIME type; `sourceLang` may be `auto`.

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/google"
)

func main() {
	translator := google.NewTranslator(google.WithApiKey("your_google_api_key"))

	// TranslateV2 returns (text, detected source language, error); "en" is ignored
	text, detected, err := translator.TranslateV2("Hello, world!", "en", "zh-CN")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Translation:", text)
	fmt.Println("Detected source language:", detected.String())
}
```

### Version Comparison

| Feature | V1 | V2 | V3 |
|---------|----|----|----|
| Free? | Free (unofficial endpoint, no SLA) | No (pay per use) | No (pay per use) |
| API Key required | No | Yes | Yes |
| Officially supported | No | Yes | Yes |
| Source language | Supports `auto` | `sourceLang` ignored, always auto-detected | Supports `auto` |
| Return value | Text | Text + detected source language (`TranslateV2`) | Text |
| MIME type | - | - | Fixed `text/plain` |

**Choosing a version**:

- **V1**: for low-traffic projects, prototyping, and learning; it is an unofficial endpoint that may change or be rate-limited at any time
- **V2**: for projects that need official support and the detected source language
- **V3**: the currently recommended official API version

### Error Handling

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/google"
)

func main() {
	translator := google.NewTranslator()

	result, err := translator.Translate("Hello", "en", "zh-CN")
	if err != nil {
		log.Printf("translate failed: %v", err)
		return
	}

	fmt.Println(result)
}
```

**V1** returns fixed error strings that can be matched with `err.Error()` (**V1 only**):

- `error getting translate.googleapis.com`
- `error reading response body`
- `error 400 (Bad Request)`
- `error unmarshalling data`
- `no translated data in response`

**V2 / V3** return errors from the official SDKs (including gRPC status context); handle them as plain `error` values and do **not** match on strings.

### Using the Unified Interface

This module does not import `go-utils/translator`, but its `Translate` method signature matches the `Translator` interface, so it can be assigned directly:

```go
package main

import (
	"github.com/tx7do/go-utils/translator"
	"github.com/tx7do/go-utils/translator/google"
)

func main() {
	tr := google.NewTranslator() // defaults to V1, free
	var iface translator.Translator = tr
	_, _ = iface.Translate("Hello", "en", "zh-CN")
}
```

## Supported Language Codes

The following are Google-style language codes (as used by Google Translate — e.g. Hebrew is `iw`, Simplified Chinese is `zh-CN`), 109 in total. The source language may be `auto`; V2 target languages must be parseable by `language.Parse`. The full list is in the [Google Translate language list](https://cloud.google.com/translate/docs/languages).

| Language | Code | Language | Code |
|----------|------|----------|------|
| Albanian | `sq` | Burmese | `my` |
| Arabic | `ar` | Hmong | `hmn` |
| Amharic | `am` | Xhosa | `xh` |
| Azerbaijani | `az` | Zulu | `zu` |
| Irish | `ga` | Nepali | `ne` |
| Estonian | `et` | Norwegian | `no` |
| Odia | `or` | Punjabi | `pa` |
| Basque | `eu` | Portuguese | `pt` |
| Belarusian | `be` | Pashto | `ps` |
| Bulgarian | `bg` | Chichewa | `ny` |
| Icelandic | `is` | Japanese | `ja` |
| Polish | `pl` | Swedish | `sv` |
| Bosnian | `bs` | Samoan | `sm` |
| Persian | `fa` | Serbian | `sr` |
| Afrikaans | `af` | Sesotho | `st` |
| Tatar | `tt` | Sinhala | `si` |
| Danish | `da` | Esperanto | `eo` |
| German | `de` | Slovak | `sk` |
| Russian | `ru` | Slovenian | `sl` |
| French | `fr` | Swahili | `sw` |
| Filipino | `tl` | Scottish Gaelic | `gd` |
| Finnish | `fi` | Cebuano | `ceb` |
| Frisian | `fy` | Somali | `so` |
| Khmer | `km` | Tajik | `tg` |
| Georgian | `ka` | Telugu | `te` |
| Gujarati | `gu` | Tamil | `ta` |
| Kazakh | `kk` | Thai | `th` |
| Haitian Creole | `ht` | Turkish | `tr` |
| Korean | `ko` | Turkmen | `tk` |
| Hausa | `ha` | Welsh | `cy` |
| Dutch | `nl` | Uyghur | `ug` |
| Kyrgyz | `ky` | Urdu | `ur` |
| Galician | `gl` | Ukrainian | `uk` |
| Catalan | `ca` | Uzbek | `uz` |
| Czech | `cs` | Spanish | `es` |
| Kannada | `kn` | Hebrew | `iw` |
| Corsican | `co` | Greek | `el` |
| Croatian | `hr` | Hawaiian | `haw` |
| Kurdish | `ku` | Sindhi | `sd` |
| Latin | `la` | Hungarian | `hu` |
| Latvian | `lv` | Shona | `sn` |
| Lao | `lo` | Armenian | `hy` |
| Lithuanian | `lt` | Igbo | `ig` |
| Luxembourgish | `lb` | Italian | `it` |
| Kinyarwanda | `rw` | Yiddish | `yi` |
| Romanian | `ro` | Hindi | `hi` |
| Malagasy | `mg` | Sundanese | `su` |
| Maltese | `mt` | Indonesian | `id` |
| Marathi | `mr` | Javanese | `jw` |
| Malayalam | `ml` | English | `en` |
| Malay | `ms` | Yoruba | `yo` |
| Macedonian | `mk` | Vietnamese | `vi` |
| Maori | `mi` | Chinese (Traditional) | `zh-TW` |
| Mongolian | `mn` | Chinese (Simplified) | `zh-CN` |
| Bengali | `bn` |  |  |

## Obtaining an API Key

1. Visit the [Google Cloud Console](https://console.cloud.google.com/)
2. Create (or select) a project
3. Enable the Cloud Translation API
4. Create an API Key and pass it to `WithApiKey`

### Pricing

- **V1**: free (unofficial endpoint, no SLA)
- **V2 / V3**: pay per use — see [Google Cloud pricing](https://cloud.google.com/translate/pricing)

## Notes

- **No context support**: `Translate` and all versioned methods accept no `context.Context`; V2 / V3 create their clients and call the API with a fixed `context.Background()` internally, so no timeout can be passed in from outside. To apply a timeout, wrap the call externally with a goroutine + `select`.
- **Clients are lazily created and cached**: the V2 / V3 clients are created on first use and stored in the `Translator` instance; reuse the instance instead of rebuilding it.
- **V1 is unofficial**: implemented against the web translate endpoint, with no official SLA; it may change, be rate-limited, or stop working at any time. Avoid it for critical workloads.
- **V2 ignores the source language**: regardless of what `sourceLang` is, V2 detects the language automatically and reports it via `TranslateV2`'s second return value.
- **V1 is rate-limited**: there is no official quota documentation for this endpoint; control your request rate and prepare a fallback for throttling.
- Advanced capabilities (very large texts, HTML translation via V3's `text/html` MIME) are not wrapped by this module; use the official SDK directly when needed.

## Dependencies

Matching `go.mod`:

```
cloud.google.com/go/translate v1.12.7
golang.org/x/text v0.34.0
google.golang.org/api v0.269.0
```

## Acknowledgements

- [Google Cloud Translation API](https://cloud.google.com/translate/docs)
- [cloud.google.com/go/translate Go SDK](https://github.com/googleapis/google-cloud-go)

## Related Links

- [Google Cloud Console](https://console.cloud.google.com/)
- [Google Translate language list](https://cloud.google.com/translate/docs/languages)
- [Project home](https://github.com/tx7do/go-utils)

## License

Released under the project [MIT License](https://github.com/tx7do/go-utils/blob/main/LICENSE).
