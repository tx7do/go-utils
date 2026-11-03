# slug

URL slug generator: a thin wrapper around
[github.com/gosimple/slug](https://github.com/gosimple/slug) v1.15.0 that turns
text in any language (Chinese, Russian, German, English, ...) into URL-friendly slugs.

English | [简体中文](README.zh-CN.md)

## Features

- Converts text in any language to a URL slug (transliterated internally via unicode)
- Lowercases by default, with a case-preserving variant
- English and German language corpora (`&` → `and` / `und`)
- Passes through the upstream `slug.CustomSub` custom-replacement capability

## Installation

slug is an independent Go module (`slug/go.mod`) depending on
`github.com/gosimple/slug v1.15.0`:

```bash
go get github.com/tx7do/go-utils/slug
```

```go
import "github.com/tx7do/go-utils/slug"
```

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/slug"
)

func main() {
	fmt.Println(slug.Generate("Hellö Wörld хелло ворлд")) // hello-world-khello-vorld
	fmt.Println(slug.Generate("影師"))                     // ying-shi
	fmt.Println(slug.Generate("天下太平"))                   // tian-xia-tai-ping

	// English corpus: & becomes "and".
	fmt.Println(slug.GenerateEnglish("This & that")) // this-and-that

	// Case preserved.
	fmt.Println(slug.GenerateCaseSensitive("This & That")) // This-and-That

	// German corpus: & becomes "und".
	fmt.Println(slug.GenerateGerman("Diese & Dass")) // diese-und-dass
}
```

## API Reference

| Function | Signature | Description |
|----------|-----------|-------------|
| `Generate` | `func(input string) string` | Generates a slug: forces lowercase, uses the English corpus (`slug.MakeLang(input, "en")`) |
| `GenerateCaseSensitive` | `func(input string) string` | Generates a slug: preserves case, uses the English corpus |
| `GenerateEnglish` | `func(input string) string` | Semantic alias equivalent to `Generate` (forced lowercase + English corpus) |
| `GenerateGerman` | `func(input string) string` | Generates a slug: forces lowercase, uses the German corpus (`slug.MakeLang(input, "de")`, `&` becomes `und`) |

Every function delegates to the upstream `slug.MakeLang`; the conversion rules
(unicode transliteration, invalid-character removal, whitespace-to-`-`, etc.) are
defined by the upstream library.

### Custom replacements

The upstream library exposes the package-level variable `slug.CustomSub`
(`map[string]string`); set it before calling to apply custom replacements
(running before transliteration):

```go
slug.CustomSub = map[string]string{
	"water": "sand",
}
slug.Generate("water is hot") // sand-is-hot
```

## Notes

- **Be careful under concurrency**: the upstream `slug.Lowercase` and `slug.CustomSub`
  are package-level globals, and every function in this package rewrites
  `slug.Lowercase` on every call (`Generate` sets it to `true`,
  `GenerateCaseSensitive` to `false`). When goroutines call different functions
  concurrently, the case behavior may override each other; `slug.CustomSub`, once
  set, applies globally. For strict isolation, use the upstream library directly
  with your own locking, or stick to a single function for a given corpus.
- `Generate` and `GenerateEnglish` behave identically; choose by semantics.
- Transliteration results follow the conversion tables of the upstream
  [gosimple/unidecode](https://github.com/gosimple/unidecode): Russian `хелло`
  becomes `khello`, traditional Chinese `影師` becomes `ying shi`, etc.; the unit
  tests (`slug_test.go`) are the source of truth.
