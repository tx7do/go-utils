# fieldperm

English | [简体中文](README.zh-CN.md)

Generic field-level-permission message trimming primitives: the hidden-field set comes from a token claim (aggregated as a role union at login, entries in the form `"resource.field"` where field is the proto json_name), and the execution side uses it to trim proto messages in both read and write directions. Everything is implemented with protoreflect reflection — adding a new controlled resource requires no changes to the executor.

## Features

- **Read path** (`ApplyReadMask` / `ApplyReadMaskList`): clears matching fields in place. protojson omits unpopulated fields by default, so clearing a value equals removing the field from the JSON response
- **Write path** (`StripWriteFields`): clears matching field values in the request payload and removes the corresponding paths from the field_mask, preventing Update's mask-filtering semantics from writing the cleared values back as explicit zeros
- **Dual-spelling matching**: the hidden set may hold proto field names (snake_case) or json_names (camelCase) — either spelling hits the same field when clearing values
- **Generic list version**: `ApplyReadMaskList` trims every item of a List response in one line
- **Pure reflection**: no dependency on concrete message types, no code generation

## Installation

This package is an independent Go module; install it on its own:

```bash
go get github.com/tx7do/go-utils/fieldperm
```

```go
import "github.com/tx7do/go-utils/fieldperm"
```

## Quick Start

Complete example (a `User` message with `password_hash`, json_name `passwordHash`):

```go
package main

import (
	"github.com/tx7do/go-utils/fieldperm"
)

func main() {
	// The following variables are assumed to come from the request context:
	//   msg            — the *userpb.User message to trim
	//   items          — the []*userpb.User items of a List response
	//   req.Data       — write-request payload *userpb.User
	//   req.UpdateMask — write-request *fieldmaskpb.FieldMask

	// Hidden-field entries from the token claim, "resource.field" format
	// (json_name spelling recommended)
	entries := []string{"User.passwordHash", "User.salt"}

	// Hidden-field set for the target resource
	fields := fieldperm.HiddenFieldsOf(entries, "User")

	// Read path: clear matching fields of the User message in place
	fieldperm.ApplyReadMask(msg, fields)

	// List response: apply read trimming to every item
	fieldperm.ApplyReadMaskList(items, fields)

	// Write path: clear payload values + strip matching paths from the field mask
	fieldperm.StripWriteFields(req.Data, req.UpdateMask, fields)
}
```

Alternatively parse first and pick per resource:

```go
parsed := fieldperm.ParseTokenEntries(entries) // map[resource]field set
fields := parsed["User"]
```

## API Reference

| Function | Signature | Description |
| --- | --- | --- |
| `ParseTokenEntries` | `func ParseTokenEntries(entries []string) map[string]map[string]struct{}` | Parses `"resource.field"` entries into resource → field-set; entries are **stored verbatim** (no spelling conversion); split at the first `.`; entries without a dot or with an empty resource/field are silently skipped |
| `HiddenFieldsOf` | `func HiddenFieldsOf(entries []string, resource string) map[string]struct{}` | Equivalent to `ParseTokenEntries(entries)[resource]`; nil when not configured |
| `IsEmpty` | `func IsEmpty(hidden map[string]struct{}) bool` | Whether the hidden set is empty (empty = no trimming needed) |
| `ApplyReadMask` | `func ApplyReadMask(msg proto.Message, hidden map[string]struct{})` | Clears matching **top-level** fields in place; matches the proto field name first, then the json_name (dual-spelling hit); returns quietly for nil or invalid messages; empty set is a no-op |
| `ApplyReadMaskList` | `func ApplyReadMaskList[T proto.Message](items []T, hidden map[string]struct{})` | Generic list version; applies read trimming to every item; empty set is a no-op for the whole list |
| `StripWriteFields` | `func StripWriteFields(data proto.Message, fieldMask *fieldmaskpb.FieldMask, hidden map[string]struct{})` | Applies read trimming to data first (dual-spelling value clearing), then removes matching entries from `fieldMask.Paths`; paths are matched **verbatim by top-level segment** against the hidden set (`field.sub` is dropped entirely when its top segment `field` hits, spelling must match the entry); a nil fieldMask only clears values; empty set is a no-op |

## Notes / Caveats

- **Top-level only, no recursion**: `ApplyReadMask` walks only the message's top-level fields; nested message interiors are untouched — this primitive covers top-level business fields.
- **Where dual-spelling matching happens**: `ParseTokenEntries` stores entries verbatim; the json_name (camelCase) vs proto-name (snake_case) dual-spelling hit happens inside `ApplyReadMask` (tries `fd.Name()` first, then `fd.JSONName()` when different). So for **value clearing** either spelling in the config is enough.
- **Mask-path removal is verbatim**: when `StripWriteFields` strips `fieldMask.Paths`, the path's top segment is compared literally against the hidden set — only spellings actually present in the set are removed (covered by a test case: with `createdBy` in the hidden set, the mask path `created_by` is NOT removed, only the field value is cleared). For full write-path protection, include both spellings in the token entries (e.g. `User.passwordHash` and `User.password_hash`).
- protojson omits unpopulated fields by default, so on the read path "clear the value" equals "remove from the JSON response"; if your serialization explicitly emits zero-value fields (`EmitUnpopulated`), cleared fields still appear — avoid enabling that option.
- `fieldMask.Paths` is modified in place (`StripWriteFields` reuses the underlying slice); no need to copy before passing, but do not share it afterwards.

## Tests

Run tests from the module directory:

```bash
cd fieldperm
go test ./...
```
