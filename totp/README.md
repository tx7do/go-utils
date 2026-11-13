# totp

RFC 6238 TOTP (time-based one-time password) implementation — pure Go standard library, zero third-party dependencies.

English | [简体中文](README.zh-CN.md)

## Introduction

This module (`github.com/tx7do/go-utils/totp`) implements the verifier-app-default profile: **SHA1 / 6 digits / 30-second step**. Secrets are Base32 (RFC 4648, no padding); the `otpauth://` binding URI is assembled manually — nothing but the standard library.

## Features

- `GenerateSecret` — 160-bit secret, Base32 no-padding encoded (32 chars)
- `ValidateSecret` — normalizes lowercase / padding / inner spaces, validates shape
- `Verify` — ±1 step clock-skew window (90 s tolerance), constant-time comparison
- `OTPAuthURI` — `otpauth://totp/` URI for QR enrolment / manual entry
- Zero dependencies; verified against the RFC 6238 Appendix B reference vectors

## Installation

```bash
go get github.com/tx7do/go-utils/totp
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tx7do/go-utils/totp"
)

func main() {
	secret, err := totp.GenerateSecret()
	if err != nil {
		log.Fatal(err)
	}

	// Show this URI as a QR code for authenticator apps
	uri := totp.OTPAuthURI("Example Corp", "admin@example.com", secret)
	fmt.Println(uri)

	// Verify a 6-digit code (±1 step window)
	ok := totp.Verify(secret, "287082", time.Now())
	fmt.Println("valid:", ok)
}
```

## API

| Function | Signature | Description |
|----------|-----------|-------------|
| `GenerateSecret` | `() (string, error)` | New 160-bit secret, Base32 no-padding |
| `ValidateSecret` | `(secret string) (string, error)` | Normalize + validate; returns uppercase no-padding form |
| `Verify` | `(secret, code string, at time.Time) bool` | 6-digit code check, ±1 step skew, constant-time |
| `OTPAuthURI` | `(issuer, account, secret string) string` | `otpauth://totp/` binding URI |

## Notes

1. Only the SHA1 / 6-digit / 30 s profile is implemented — the default of Google Authenticator, Microsoft Authenticator, 1Password, etc.
2. `Verify` accepts codes from the previous/current/next step (±30 s around `at`); widen `verifyWindow` if your deployment needs more tolerance.
3. Comparison is constant-time (`crypto/subtle`) to resist timing attacks.

## License

Released under the project [MIT License](https://github.com/tx7do/go-utils/blob/main/LICENSE).
