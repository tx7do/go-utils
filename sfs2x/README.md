# sfs2x — SFS2X Binary Protocol Codec

`sfs2x` implements the SFS2X binary wire protocol (the object serialization
and message framing used by SmartFoxServer 2X compatible services): typed
key/value objects, nested objects and arrays, and length-prefixed frames.
It is a pure standard-library implementation — no external dependencies.

English | [简体中文](README.zh-CN.md)

## Wire Format

```
frame  = [header 1B][len][payload]
         header: bit5 (0x20) = zlib-compressed payload
                 bit3 (0x08) = length is uint32, otherwise uint16
object = [0x12][u16 count][entries...]
entry  = [u16 keyLen][key utf8][1B type][value]
```

Nested `SFS_OBJECT` / `SFS_ARRAY` values carry no separate marker — their
type byte doubles as the marker; the decoder seeks back one byte on entry.

## Type Codes

| Code | Type | Code | Type |
|---|---|---|---|
| 0 | NULL | 9 | BOOL_ARRAY |
| 1 | BOOL | 10 | BYTE_ARRAY |
| 2 | BYTE | 11 | SHORT_ARRAY |
| 3 | SHORT | 12 | INT_ARRAY |
| 4 | INT | 13 | LONG_ARRAY |
| 5 | LONG | 14 | FLOAT_ARRAY |
| 6 | FLOAT | 15 | DOUBLE_ARRAY |
| 7 | DOUBLE | 16 | UTF_STRING_ARRAY |
| 8 | UTF_STRING | 17 | SFS_ARRAY |
| | | 18 | SFS_OBJECT |

All integers are big-endian.

## Installation

```bash
go get github.com/tx7do/go-utils/sfs2x
```

```go
import "github.com/tx7do/go-utils/sfs2x"
```

## Quick Start

```go
o := sfs2x.New()
o.Put("action", sfs2x.TUTFString, "login")
o.Put("retries", sfs2x.TInt, int32(3))
o.Put("ratio", sfs2x.TDouble, 0.75)

inner := sfs2x.New()
inner.Put("token", sfs2x.TUTFString, "abc")
o.Put("params", sfs2x.TSFSObject, inner)

// Object → bytes (with 0x12 marker)
data := o.Encode()

// bytes → Object
decoded, err := sfs2x.Decode(data)

// Framed transport (signature + uint16 length + payload)
frame := sfs2x.EncodeFrame(o)
msg, err := sfs2x.DecodeFrame(frame)
```

## API

| Function | Description |
|---|---|
| `New() *Object` | Create an empty object. |
| `Put(key string, typ Type, val interface{})` | Write a typed value. |
| `Get(key string) interface{}` | Read the raw value (nil when absent). |
| `GetType(key string) Type` | Read the entry's type code. |
| `GetKeys() []string` | Keys in insertion order. |
| `GetByte/GetShort/GetInt/GetLong(key)` | Typed readers (panic on wrong type — pair with `GetType` when untrusted). |
| `GetUTFString(key) string` | UTF_STRING reader; empty string when absent/other type. |
| `GetSFSObject(key) *Object` | Nested object reader; nil when absent/other type. |
| `GetSFSArray(key) []interface{}` | Nested array reader; nil when absent/other type. |
| `Encode() []byte` | Serialize with the 0x12 marker. |
| `Decode(data []byte) (*Object, error)` | Parse from binary (marker required). |
| `EncodeFrame(o *Object) []byte` | Wrap into a frame (signature + uint16 length). |
| `DecodeFrame(data []byte) (*Object, error)` | Parse a frame; accepts uint16 and uint32 length forms. |
| `EncodeFrameWithOptions(o *Object, opts FrameOptions) ([]byte, error)` | Frame with optional zlib compression (threshold) and/or AES-128-CBC encryption (compress → encrypt). |
| `DecodeFrameWithOptions(data []byte, opts FrameOptions) (*Object, error)` | Parse a frame with optional decryption/decompression (decrypt → decompress). |

## Notes

- **Deterministic encoding**: object keys are sorted before serialization,
  so identical content always produces byte-identical output regardless of
  insertion order (`GetKeys` still reports insertion order).
- **Frame forms**: `DecodeFrame` accepts both length forms (header bit
  `0x08`). `EncodeFrame` emits the short form and automatically switches to
  BIG_SIZE (uint32) for payloads over 65535 bytes, like the official client.
- **Compression**: `DecodeFrame` decompresses `0x20` frames transparently;
  `EncodeFrameWithOptions` zlib-compresses payloads over a configurable
  threshold (official default: 1024).
- **Encryption, both official modes**: `EncodeFrameWithOptions` applies
  AES-128-CBC with PKCS#7 (header `0x40`, 16-byte key):
  - **official fixed key+IV mode** (`EncryptionKey` + `EncryptionIV`): the
    server delivers the pair over HTTPS (/BlueBox/CryptoManager, response =
    Base64(key||iv), parse with `ParseCryptoKeyResponseBase64`); the payload
    is ciphertext only, deterministic per input — exactly what real SFS2X
    peers speak;
  - **self-describing mode** (`EncryptionKey` only): a fresh random IV is
    prefixed to the payload, so frames stand alone without shared state.
- **Message envelope**: `Message{Controller, Action, Params}` with the
  standard `{c: BYTE, a: SHORT, p}` wire shape, system-controller action
  codes, and handshake request/response builders (`NewHandshakeRequest`,
  `ParseHandshakeResponse` — session token `tk` drives the crypto-key
  servlet call).
- **CLASS (19) recognition**: the wire form of a CLASS value is an ordinary
  object carrying `$C` (class name) + `$F` (field values); `AsClass`
  recognizes the convention (full custom-type serde stays with the app —
  it needs the peer's field registry).
- **Ordered encoding**: `EncodeOrdered` / `FrameOptions.Ordered` emit keys
  in insertion order — byte-for-byte compatible with the official client
  encoder; plain `Encode` stays deterministic (sorted).
- **Bounds-safe decoding**: truncated or malformed input produces errors,
  never panics. Encrypted frames without a key are rejected with an
  explicit error. Decompressed output is capped at 64 MiB to resist
  decompression bombs.
- **UTF_STRING limit**: the wire format carries string lengths in a uint16,
  so values longer than 65535 bytes cannot be encoded (silently truncated by
  the length field — same ceiling as the official client). Use BYTE_ARRAY
  (uint32 length) for larger payloads.
- **Type safety**: the typed readers (`GetInt` etc.) use unchecked type
  assertions and panic on mismatch. For untrusted input, check `GetType`
  first or use `Get` with a safe cast.

## Protocol Fidelity

Verified against the official SFS2X Java client API behavior (via a faithful
port) and interoperability captures from a production peer:

| Behavior | This package | Official client API |
|---|---|---|
| Type codes 0–18 | ✓ identical | — |
| Typed arrays (BOOL/SHORT/INT/LONG/FLOAT/DOUBLE/UTF_STRING) | ✓ encode + decode | identical |
| BYTE_ARRAY | 4-byte length + raw bytes (official shape) | uint32 `writeInt` — ZewSFS wrongly uses u16 count + signed bytes; their Float/Double are native-endian (LE) instead of big-endian. Golden fixtures patch both. |
| Other arrays / strings length field | uint16 | uint16 |
| Nested object/array marker | type byte doubles as marker (seek-back) | identical |
| Frame header flags | `0x80` BINARY, `0x08` BIG_SIZE (decode) | adds `0x40` ENCRYPTED, `0x10` BLUEBOX |
| Compression (header `0x20`) | ✓ zlib, threshold-configurable (default 1024) | zlib, threshold 1024 |
| Encryption (header `0x40`) interop | roundtrip only (offline RSA handshake not implemented) | session key via server RSA token handshake |
| Encryption (header `0x40`) | ✓ AES-128-CBC PKCS#7: official fixed key+IV pair **or** self-describing random-IV prefix | AES-128-CBC with server-issued fixed key+IV pair |
| TEXT (20) type | ✓ encode + decode (4-byte length) | `writeInt` length |
| CLASS (19) type | ✓ `$C`/`$F` convention recognized (`AsClass`); custom serde stays app-side | ordinary object with `$C`/`$F` keys |
| Encryption key delivery | `CryptoKey` parsed from servlet response (HTTPS call is app-side) | `/BlueBox/CryptoManager` servlet |
| Key ordering on encode | sorted (default) **or** insertion order (`EncodeOrdered`/`FrameOptions.Ordered`) | insertion order |
| **Object key ordering on encode** | **sorted** (deterministic output) | **insertion order** (OrderedHashMap) |

The one deliberate divergence is key ordering: this package sorts keys so
that identical content yields byte-identical encodings (reproducible tests,
hashable payloads, stable logging). Every compliant parser treats an
SFSObject as a key/value map, so decode-side behavior is unaffected. If you
need byte-for-byte fidelity with an official-client encoder, sort your
insertion order to match theirs instead — or swap the `sort.Strings` call in
`entriesBin` for insertion-order emission.
