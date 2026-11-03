# auditutil

English | [简体中文](README.zh-CN.md)

`auditutil` provides information-collection and tamper-proofing helpers for HTTP audit logs: normalized extraction of request info (client IP / request ID), username parsing from request bodies with log-injection scrubbing, User-Agent platform detection, and deterministic hashing plus ECDSA signing of audit logs (a two-level "hash + signature" tamper-proofing chain).

## Features

- **Request info collection**: client real IP, private-address detection, request ID extraction with fallback generation;
- **Username parsing**: resolves the username from JSON / form bodies, strips CR/LF to block log-line injection, and restores the body after reading;
- **Platform detection**: heuristic platform classification from the User-Agent string;
- **Tamper proofing**: deterministic SHA256 hashing (excluding the `log_hash` / `signature` fields to prevent self-reference) plus ECDSA-P256 signatures (DER-encoded, verifiable with the standard library's `ecdsa.VerifyASN1`).

## Installation

```bash
go get github.com/tx7do/go-utils/auditutil
```

```go
import "github.com/tx7do/go-utils/auditutil"
```

## Quick Start

### 1. Collect request info

```go
package main

import (
	"net/http"

	"github.com/tx7do/go-utils/auditutil"
)

func collect(req *http.Request) (ip, reqID, username, platform string) {
	ip = auditutil.ClientRealIP(req)             // X-Forwarded-For -> X-Real-IP -> RemoteAddr
	reqID = auditutil.RequestID(req)             // X-Request-ID -> X-Correlation-ID -> x-fc-request-id -> fresh GUID
	username, _ = auditutil.ExtractUsernameFromRequest(req) // body is restored after reading
	platform = auditutil.DetectPlatformFromUA(req.UserAgent())
	return
}
```

### 2. Hash and sign a log entry

```go
package main

import (
	"crypto/ecdsa"

	"github.com/tx7do/go-utils/auditutil"
	"github.com/tx7do/go-utils/trans"
)

// priv comes from auditutil.GenerateECDSAKeyPair(), typically generated at
// service startup or loaded from configuration
var priv *ecdsa.PrivateKey

func seal(log *pb.AuditLog) error {
	// 1) Hash: clear log_hash/signature on a clone, then SHA256 over the
	//    deterministic protobuf serialization
	log.LogHash = trans.Ptr(auditutil.HashLog(log))

	// 2) Sign: sign(tenant_id, user_id, created_at, log_hash), DER format
	sig, err := auditutil.SignLogContent(priv, log.TenantId, log.UserId, log.CreatedAt, log.GetLogHash())
	if err != nil {
		return err
	}
	log.Signature = sig
	return nil
}
```

> Note that `SignLogContent` returns `([]byte, error)` — catch the error first, then assign the signature to the field. `trans.Ptr` takes a string pointer and requires importing `github.com/tx7do/go-utils/trans` separately.

### 3. Key generation and verification

Signatures are verified with the standard library's `ecdsa.VerifyASN1`. The signed content is the same JSON structure used inside the library (field names and order), SHA256-hashed before verification:

```go
priv, pub, err := auditutil.GenerateECDSAKeyPair() // ECDSA P-256 (secp256r1)
if err != nil {
	// ...
}

// Verification: rebuild the signed content -> SHA256 -> VerifyASN1
type signContent struct {
	TenantID uint32 `json:"tenant_id"`
	UserID   uint32 `json:"user_id"`
	Sec      int64  `json:"sec"`   // createdAt.Seconds
	Nanos    int32  `json:"nanos"` // createdAt.Nanos
	LogHash  string `json:"log_hash"`
}

func verify(pub *ecdsa.PublicKey, tenantID, userID uint32, createdAt *timestamppb.Timestamp, logHash string, sig []byte) bool {
	sc := signContent{TenantID: tenantID, UserID: userID, LogHash: logHash}
	if createdAt != nil {
		sc.Sec, sc.Nanos = createdAt.Seconds, createdAt.Nanos
	}
	b, err := json.Marshal(sc)
	if err != nil {
		return false
	}
	h := sha256.Sum256(b)
	return ecdsa.VerifyASN1(pub, h[:], sig)
}
```

Verification fails for any tampering with `log_hash` or the other signed inputs. `SignLogContent` returns an error when the private key is nil rather than an empty signature — the signature is the credential, so callers should never fall back to storing unsigned logs.

## Configuration / Options

The package has no configuration knobs. The tamper-proofing chain follows fixed conventions:

- **Hash exclusion set**: `HashLog` clears the `log_hash`, `logHash` and `signature` spellings on a clone by default; variadic arguments override the set: `HashLog(msg, "tenant_id")`;
- **Signed content**: the SHA256 of the JSON serialization of `tenant_id`, `user_id`, `created_at` (seconds + nanos) and `log_hash`; a nil `createdAt` contributes zeros.

## API Reference

### Request info collection (client_ip.go / request.go)

- `func ClientRealIP(request *http.Request) string` — returns the first valid IP from `X-Forwarded-For`, then `X-Real-IP`, falling back to `RemoteAddr`;
- `func IPFromRemoteAddr(hostAddress string) string` — extracts a valid IP from `host:port`; returns an empty string when parsing fails;
- `func IsPrivateIP(ipStr string) bool` — reports whether the address is private/link-local (10/8, 172.16/12, 192.168/16, 127/8, 169.254/16, ::1, fc00::/7); unparseable addresses return false;
- `func RequestID(request *http.Request) string` — checks `X-Request-ID` / `X-Correlation-ID` / `x-fc-request-id` in order, generating a fresh GUIDv4 when none is present;
- `func ExtractUsernameFromRequest(r *http.Request) (string, error)` — reads and restores the body, then parses the username without affecting later middleware;
- `func ParseUsernameFromBytes(body []byte) (string, error)` — first parses the JSON field `"username"` (regex), then the form key (via `url.ParseQuery`); errors when neither matches;
- `func StripLineBreaks(s string) string` — removes all CR/LF to block log-line injection (already applied to username parsing results).

### Platform detection (platform.go)

- `func DetectPlatformFromUA(ua string) string` — heuristic platform classification from the UA; the returned value is one of the constants below; an empty UA yields `PlatformOther`;
- Platform constants: `PlatformAndroidApp` (`"AndroidApp"`), `PlatformiOSApp` (`"iOSApp"`), `PlatformDesktopWindows` (`"DesktopWindows"`), `PlatformDesktopMac` (`"DesktopMac"`), `PlatformDesktopLinux` (`"DesktopLinux"`), `PlatformWeb` (`"Web"`), `PlatformOther` (`"Other"`).

### Tamper proofing (sign.go)

- `func GenerateECDSAKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error)` — generates a secp256r1 key pair;
- `func EncodeECDSADER(r, s *big.Int) ([]byte, error)` — encodes the ECDSA r and s values as a DER sequence;
- `func HashLog(msg proto.Message, excludeFields ...string) string` — computes the deterministic SHA256 (lowercase hex) of a log; excluded fields are cleared on a **clone** before deterministic protobuf serialization, leaving the original message untouched; returns an empty string for nil/invalid messages;
- `func SignLogContent(priv *ecdsa.PrivateKey, tenantID, userID uint32, createdAt *timestamppb.Timestamp, logHash string) ([]byte, error)` — ECDSA-P256 signs the signed content and returns DER-encoded bytes (starting with `0x30`); a nil `priv` yields an error.

## Notes

- **ClientRealIP records what the headers say**: it logs the forwarded headers as-is for auditing, so the outermost reverse proxy must overwrite (not append) `X-Forwarded-For` to prevent spoofing; for spoof-proof security decisions (trusted-proxy chain resolved right-to-left) use `netutil.ClientIPFromRequest` instead;
- **DetectPlatformFromUA is a supplementary signal**: prefer the client-reported `platform` / `os_name` fields when available;
- **Hash is self-reference free**: the `log_hash` and `signature` fields are excluded from hashing, matching both the proto field name and its `json_name` spelling;
- **Verifiable without the log body**: a verifier only needs `tenant_id`, `user_id`, `created_at` and `log_hash` to rebuild the signed content and check the signature with the standard library;
- **Dependencies**: `google.golang.org/protobuf` (proto.Message hashing), `github.com/tx7do/go-utils/id` (GUID generation).

## Acknowledgements

- [google.golang.org/protobuf](https://google.golang.org/protobuf)
- Go standard library `crypto/ecdsa` / `crypto/sha256`
