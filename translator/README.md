# translator

Machine translation package for go-utils: a unified `Translator` interface plus pluggable cloud-provider implementations.

English | [简体中文](README.zh-CN.md)

## Introduction

`translator` is the translation component collection of [go-utils](https://github.com/tx7do/go-utils).

This module (`github.com/tx7do/go-utils/translator`) only defines the unified translator interface. It contains no provider implementations and has no third-party dependencies; each provider lives in its own sub-module and is installed separately. The interface is defined in `translator.go`:

```go
// Translator is a translator.
type Translator interface {
	// Translate translates text.
	Translate(source, sourceLang, targetLang string) (string, error)
}
```

## Features

- **Unified interface**: every provider implements the same `Translate(source, sourceLang, targetLang string) (string, error)` method, so implementations are interchangeable
- **Decoupled modules**: the interface module has zero dependencies; each provider is an independent Go module, so you only pull in what you use
- **Multiple providers**: Alibaba Cloud Translate, Baidu Translate, Google Translate, and Volcengine Translate

## Installation

The interface module (an independent Go module):

```bash
go get github.com/tx7do/go-utils/translator
```

```go
import "github.com/tx7do/go-utils/translator"
```

> This repository is a multi-module repository. `translator` and its sub-modules `alibaba`, `baidu`, `google`, and `volc` are all independent Go modules, each with its own `go.mod` and independently versioned.

## Sub-module Comparison

| Provider | Sub-module | Free? | Authentication | Import path | Provider docs |
|----------|------------|-------|----------------|-------------|---------------|
| Alibaba Cloud Machine Translation | [alibaba](alibaba/) | Free tier, pay-per-use beyond it (see official pricing) | AccessKey ID + AccessKey Secret | `github.com/tx7do/go-utils/translator/alibaba` | [Alibaba Cloud MT](https://help.aliyun.com/product/2477301.html) |
| Baidu Translate | [baidu](baidu/) | Free tier (see official pricing) | AppID + SecretKey (MD5 signature) | `github.com/tx7do/go-utils/translator/baidu` | [Baidu Translate Open Platform](https://fanyi-api.baidu.com/) |
| Google Translate | [google](google/) | V1 free (unofficial endpoint, no SLA); V2/V3 paid | None for V1; API Key for V2/V3 | `github.com/tx7do/go-utils/translator/google` | [Cloud Translation API](https://cloud.google.com/translate/docs) |
| Volcengine Translate | [volc](volc/) | Free tier, pay-per-use beyond it (see official pricing) | AccessKey + SecretKey (HMAC-SHA256 signature) | `github.com/tx7do/go-utils/translator/volc` | [Volcengine Translate](https://www.volcengine.com/docs/4640/35107) |

## Quick Start

Example using Google Translate's free V1 endpoint (no credentials required):

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/google"
)

func main() {
	// NewTranslator(opts ...Option) *Translator, defaults to the V1 version
	tr := google.NewTranslator()

	// Translate(source, sourceLang, targetLang string) (string, error)
	// "auto" as the source language enables automatic detection
	result, err := tr.Translate("Hello, world!", "auto", "zh-CN")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Translation:", result)
}
```

Install the corresponding sub-module:

```bash
go get github.com/tx7do/go-utils/translator/google
```

### Using the Unified Interface

The sub-modules do not import this interface package, but their `Translate` method signature matches `translator.Translator`. Thanks to Go's structural typing, any provider's translator can be assigned directly to a `translator.Translator`:

```go
package main

import (
	"github.com/tx7do/go-utils/translator"
	"github.com/tx7do/go-utils/translator/volc"
)

// Program against the interface: business code does not depend on a specific provider
func translateText(tr translator.Translator, text, from, to string) (string, error) {
	return tr.Translate(text, from, to)
}

func main() {
	tr, err := volc.NewTranslator("your-access-key", "your-secret-key")
	if err != nil {
		panic(err)
	}
	_, _ = translateText(tr, "Hello", "en", "zh")
}
```

## Advanced Usage

- **Switching providers**: when business code only depends on `translator.Translator`, switching providers is just a matter of replacing the constructor (constructor parameters differ per provider; see each sub-module's README)
- **Multi-provider fallback**: combine multiple implementations and fall back to a secondary provider when the primary one fails
- **Language codes**: each provider uses its own language code system; see the language code table in the corresponding sub-module's README (the Google-style code table is in [google/README.md](google/README.md))

## Notes

- Providers differ in capabilities, billing, and rate limiting; read the "Notes" section of the corresponding sub-module's README before production use
- Each sub-module is an independent Go module: you must install the full sub-module path (e.g. `github.com/tx7do/go-utils/translator/alibaba`); installing only the parent module brings in no provider implementation
- This README no longer ships a generic language code table: the old table used Google-style codes (such as `iw` and `zh-CN`) and only applied to the Google provider; it has moved to [google](google/README.md)

## Acknowledgements

- Thanks to the cloud providers offering machine translation services: [Alibaba Cloud](https://www.aliyun.com/), [Baidu](https://www.baidu.com/), [Google Cloud](https://cloud.google.com/), and [Volcengine](https://www.volcengine.com/)
- The sub-modules are built on the Alibaba Cloud Go SDK, the Baidu Translate Open Platform API, the Google Cloud Translation Go SDK, and the Volcengine Translate API respectively

## License

This project is licensed under the [MIT License](https://github.com/tx7do/go-utils/blob/main/LICENSE).
