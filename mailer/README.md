# mailer

English | [简体中文](README.zh-CN.md)

Lightweight SMTP mail sending built on the standard library `net/smtp`: one plain-text email, one config struct, one function call.

## Features

- **Two TLS modes**: STARTTLS (587/25, upgraded automatically when the server supports it) and implicit SSL/TLS (465, TLS from the first byte);
- **SMTP PLAIN auth**: covers the auth-code flows of common mailbox providers; auth is skipped when `Username` is empty (anonymous relay, useful on intranets);
- **`ctx` reaches all the way to the TCP dial**: when the SMTP server is unreachable the call is not stuck until the OS-level connect timeout (20s+ on Windows) — synchronous paths such as login verification codes need interruptible timeouts;
- **Pre-dial TLS mode check**: `IsSupportedTlsMode` separates "unknown config value" (a configuration problem) from "SMTP server complained" (a delivery problem) so they are not conflated into one failure;
- **Sender display name**: when `FromName` is non-empty the From header becomes `Name <address>`, with non-ASCII names encoded per RFC 2047 (built via `net/mail.Address`, avoiding hand-rolled quoting pitfalls).

## Installation

Standalone module with no other go-utils dependencies:

```bash
go get github.com/tx7do/go-utils/mailer@latest
```

```go
import "github.com/tx7do/go-utils/mailer"
```

## Quick Start

```go
package main

import (
	"context"
	"log"

	"github.com/tx7do/go-utils/mailer"
)

func main() {
	err := mailer.SendMail(context.Background(),
		mailer.SmtpConfig{
			Host:     "smtp.example.com",
			Port:     465,
			Username: "no-reply@example.com",
			Password: "auth-code",           // provider auth code
			From:     "no-reply@example.com",
			FromName: "Example App",         // optional: sender display name
			TlsMode:  "SSL",
		},
		[]string{"to@example.com"},          // multiple recipients supported
		"Your login code",
		"Your verification code is 123456, valid for 5 minutes.",
	)
	if err != nil {
		log.Fatal(err)
	}
}
```

## Configuration

All `SmtpConfig` fields, matched word-for-word against the definition in `smtp.go`:

| Field | Type | Description |
|---|---|---|
| `Host` | `string` | SMTP server address, required |
| `Port` | `uint32` | Port, required (465 for SSL; 587/25 for START_TLS or NONE) |
| `Username` | `string` | Username; when empty authentication is skipped |
| `Password` | `string` | Password / auth code |
| `From` | `string` | Sender address; falls back to `Username` when empty |
| `FromName` | `string` | Optional sender display name; when non-empty the From header becomes `Name <address>` with non-ASCII names encoded per RFC 2047 |
| `TlsMode` | `string` | TLS mode: `NONE` / `START_TLS` / `SSL`, case-insensitive; empty string is treated as `NONE` |

Behavior of each `TlsMode` value:

- **`SSL`**: dial with TLS from the start (implicit TLS, typically 465);
- **`START_TLS` / `NONE`**: connect in plaintext first, upgrade via STARTTLS if the server advertises the extension; if the server does not support STARTTLS, delivery continues in plaintext (common with intranet debug SMTP servers);
- **Any other value**: `SendMail` fails with `unsupported tls mode` before dialing. `IsSupportedTlsMode` lets you pre-check the same config value before delivery.

## API Reference

```go
// Send one plain-text email via SMTP; to may contain multiple recipients.
func SendMail(ctx context.Context, cfg SmtpConfig, to []string, subject, body string) error

// Report whether mode is a TLS mode SendMail recognizes
// (NONE / START_TLS / SSL, case-insensitive, empty string is valid).
func IsSupportedTlsMode(mode string) bool
```

The body is always `text/plain; charset=UTF-8`; the `Subject` is not additionally encoded (some clients may render non-ASCII subjects poorly — encode with RFC 2047 yourself if needed).

## Notes / Caveats

- **Error wrapping format**: errors returned by `SendMail` are wrapped per stage; match the prefix or the message to locate the failing step:
  - `connect smtp server failed: ...` — TCP dial / STARTTLS upgrade failure;
  - `smtp auth failed: ...` — authentication rejected;
  - `smtp MAIL failed: ...` / `smtp RCPT failed for <addr>: ...` / `smtp DATA failed: ...` — server rejected the session step;
  - `unsupported tls mode: <value>` — config value outside the enum;
  - `smtp host/port is not configured` / `recipient is empty` — argument validation.
- **From fallback**: when `From` is empty, `Username` is used as the sender address;
- **Missing STARTTLS is not an error**: in `START_TLS`/`NONE` mode delivery continues in plaintext when the server does not support STARTTLS — use `SSL` mode if you require encryption;
- **One call per message**: every `SendMail` dials, sends and `Quit`s independently with no connection reuse; put queuing and throttling in your own layer for high-volume sending.
