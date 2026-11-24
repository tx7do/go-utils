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

## Notes

- **Deterministic encoding**: object keys are sorted before serialization,
  so identical content always produces byte-identical output regardless of
  insertion order (`GetKeys` still reports insertion order).
- **Frame forms**: `EncodeFrame` always emits the short (uint16) length
  form. `DecodeFrame` accepts both length forms (header bit `0x08`).
- **Compression**: frames with header bit `0x20` set carry a zlib-compressed
  payload; this package does not decompress them — handle that variant at
  the transport layer if a peer emits it.
- **Type safety**: the typed readers (`GetInt` etc.) use unchecked type
  assertions and panic on mismatch. For untrusted input, check `GetType`
  first or use `Get` with a safe cast.
