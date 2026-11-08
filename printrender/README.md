# printrender

Server-side HTML → PDF rendering: prints a final HTML document with the local
Chrome/Edge browser (chromedp-driven, auto-detects the install path).

English | [简体中文](README.zh-CN.md)

## Design boundary

The typical division of labor for print/document pipelines: pagination, mm-based
layout and fonts are done by a **frontend rendering kernel**, which produces the
final HTML. This package only performs the last step — `PrintToPDF` — with paper
width/height and orientation supplied by the caller from template metadata. The
preview and the printed output are then guaranteed identical (same Chromium
engine); the backend never re-flows or understands the template. Feeding any
complete HTML document works just as well when no frontend is involved.

A `Renderer` interface is provided: the default implementation drives the local
browser in-process (chromedp); deployments preferring a sidecar render service
(Playwright + Chromium image) implement `Renderer` and swap it in.

## Installation

printrender is an independent Go module (`printrender/go.mod`):

```bash
go get github.com/tx7do/go-utils/printrender
```

```go
import "github.com/tx7do/go-utils/printrender"
```

## Quick Start

```go
pdf, err := printrender.RenderPDF(ctx, printrender.Options{
	HTML:          finalHTML, // complete <html> document
	PaperWidthMm:  210,       // A4 portrait
	PaperHeightMm: 297,
	Background:    true,
	Timeout:       30 * time.Second,
})
if err != nil {
	return err
}
// pdf is the PDF file content ([]byte)
```

Swap the renderer (e.g. to a sidecar service):

```go
printrender.SetRenderer(myRemoteRenderer{})
```

## API Reference

| Function | Description |
|----------|-------------|
| `RenderPDF(ctx context.Context, opts Options) ([]byte, error)` | Renders PDF via the default renderer (falls back to `ChromeRenderer`) |
| `SetRenderer(r Renderer)` | Replaces the default renderer process-wide |
| `ChromeRenderer` | Local Chrome/Edge renderer; browser instance is reused process-wide |

`Options`:

| Field | Description |
|-------|-------------|
| `HTML string` | The final document (already paginated), should be a complete `<html>` document |
| `PaperWidthMm / PaperHeightMm float64` | Paper size in millimetres (required) |
| `Landscape bool` | Landscape orientation |
| `Background bool` | Print background colors/images (PrintToPDF skips backgrounds by default) |
| `Timeout time.Duration` | Per-render timeout, default 60s |

## Browser discovery

`CHROME_PATH` env var → `chrome` on `PATH` → well-known Chrome install paths →
**Microsoft Edge fallback** (Windows Server machines often lack Chrome; Edge
speaks the same Chromium protocol).

On Linux running as root (common in containers), `--no-sandbox` is added
automatically since Chrome refuses to start without it; set
`PRINTRENDER_NO_SANDBOX=0` to disable this behavior.

## Notes

- The HTML is embedded into the page via `document.write` after JSON encoding
  (`json.Marshal` escapes `<`, `>`, `&`, U+2028/29), so arbitrary HTML is safe to
  inject into the evaluation expression.
- The allocator context is created once and reused across renders; each render
  creates a fresh browser tab and closes it afterwards.
- Printing has no network/extension needs; the launched browser disables
  extensions and first-run defaults to narrow the attack surface.
