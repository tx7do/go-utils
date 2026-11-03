# doctext

English | [简体中文](README.zh-CN.md)

Extracts plain text from common document formats for knowledge-base RAG ingestion, full-text search, previewing and similar use cases. Extractors are routed by file extension; unsupported extensions are rejected outright — better an explicit "unsupported" error than feeding binary garbage into your vector store.

## Features

- **Supported formats**: plain-text family `.txt` / `.md` / `.markdown` / `.csv` / `.log` / `.json` / `.xml` / `.yml` / `.yaml` / `.html` / `.htm`, plus `.docx` (zip + XML parsed with the standard library, `word/document.xml`) and `.pdf` (pure-Go text-layer extraction via [ledongthuc/pdf](https://github.com/ledongthuc/pdf))
- **Size cap**: 10 MiB by default (`MaxFileSize = 10 << 20`); larger files are rejected with an error
- **Output sanitization**: output is forced to valid UTF-8 (invalid bytes replaced with U+FFFD) and `\r\n` is normalized to `\n`
- **docx handling**: paragraphs `</w:p>` and `<w:br/>` become newlines, `<w:tab/>` becomes a tab, all XML tags are stripped, HTML entities are unescaped
- **PDF handling**: per-page plain-text extraction (no OCR); a failed page does not abort the whole document; scanned/image-only PDFs without a text layer return an error
- `SupportedExtensions()` returns the list of supported extensions, e.g. for admin UI hints

## Installation

This package is an independent Go module; install it on its own:

```bash
go get github.com/tx7do/go-utils/doctext
```

```go
import "github.com/tx7do/go-utils/doctext"
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tx7do/go-utils/doctext"
)

func main() {
	data, err := os.ReadFile("manual.docx")
	if err != nil {
		log.Fatal(err)
	}

	text, err := doctext.Extract("manual.docx", data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(text)

	// Show supported formats in an admin page
	fmt.Println(doctext.SupportedExtensions())
}
```

## API Reference

| Member | Signature | Description |
| --- | --- | --- |
| `Extract` | `func Extract(fileName string, data []byte) (string, error)` | Extracts plain text by extension (case-insensitive); checks empty content and the size cap first, then dispatches by extension |
| `SupportedExtensions` | `func SupportedExtensions() []string` | Returns all 13 supported extensions |
| `MaxFileSize` | `const MaxFileSize = 10 << 20` | 10 MiB limit |

### Error scenarios

| Scenario | Error message |
| --- | --- |
| Empty content | `file is empty` |
| Larger than 10 MiB | `file too large: %d bytes (max %d)` |
| Unsupported extension | `unsupported file type %q (supported: %s)` |
| Plain-text content is all whitespace | `no text content extracted` |
| docx missing `word/document.xml` | `docx extract: word/document.xml not found in docx` |
| docx empty after tag stripping | `docx extract: no text content extracted from docx` |
| PDF cannot be opened | `pdf extract: open pdf: ...` |
| PDF without a text layer (scanned/image-only) | `pdf extract: no text layer found (scanned/image pdf is not supported)` |

## Notes / Caveats

- Extension matching is case-insensitive (`MANUAL.DOCX` equals `manual.docx`).
- The plain-text family does not detect encodings; bytes are processed as-is, and invalid UTF-8 is replaced with U+FFFD rather than causing an error.
- PDF extraction is **not OCR**: a scanned document has no text layer and yields the "no text layer" error instead of image content.
- The empty/size checks run before extension matching: an empty file reports `file is empty` regardless of extension.
- Only the main `word/document.xml` is parsed from docx; headers, footers and embedded objects are not extracted.

## Acknowledgements

- [github.com/ledongthuc/pdf](https://github.com/ledongthuc/pdf) — pure-Go PDF text extraction
