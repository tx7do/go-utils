# fieldmaskutil

English | [简体中文](README.zh-CN.md)

Protobuf FieldMask utilities: keep (Filter), clear (Prune) or overwrite (Overwrite) proto message fields by path list or `*fieldmaskpb.FieldMask`, with nested-path, repeated-field and map support, plus path validation and normalization — ideal for PATCH-style partial updates and response field trimming.

> **Module ownership**: this package is a subpackage of the root module `github.com/tx7do/go-utils` (**it has no go.mod of its own**); the import path is `github.com/tx7do/go-utils/fieldmaskutil`. Install the whole root module — installing by sub-path (`go get github.com/tx7do/go-utils/fieldmaskutil`) is not supported by this repository's module layout.

## Features

- **Three core operations**: `Filter` (keep the listed paths, clear everything else), `Prune` (clear the listed paths, keep everything else), `Overwrite` (overwrite the listed paths in dest with values from src)
- **FieldMask variants**: `FilterByFieldMask` / `PruneByFieldMask` / `OverwriteByFieldMask` take a `*fieldmaskpb.FieldMask` and run normalize + validate automatically; these are the ones business code uses most
- **Nested paths**: dotted paths like `a.b.c` recurse into nested messages; repeated fields and map fields are supported
- **Validation**: `Validate` / `ValidateFieldMask` check that all paths are valid for a message
- **Normalization**: `NormalizePaths` / `NormalizeFieldMaskPaths` convert paths to snake_case and map `id_` / `_id` to `id`
- **Helpers**: `PathsFromFieldNumbers` (field numbers → paths), `NilValuePaths` (find unset field paths)

## Installation

This package belongs to the root module `github.com/tx7do/go-utils`; install the whole module:

```bash
go get github.com/tx7do/go-utils
```

```go
import "github.com/tx7do/go-utils/fieldmaskutil"
```

## Quick Start

Complete example (assuming `message User { string name = 1; int32 age = 2; string email = 3; Profile profile = 4; }` and `message Profile { string bio = 1; }`):

```go
package main

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/tx7do/go-utils/fieldmaskutil"
)

func main() {
	user := getUser() // *userpb.User, illustrative: the message to process from the request context
	// {Name: "Alice", Age: 30, Email: "alice@example.com"}

	// Filter: keep the fields listed in paths, clear the rest (in place, no return value)
	// {Name: "Alice", Age: 0, Email: "alice@example.com"}
	fieldmaskutil.Filter(user, []string{"name", "email"})

	// FieldMask variant (recommended in business code): normalize + validate, then filter
	var msg proto.Message = user
	fm := &fieldmaskpb.FieldMask{Paths: []string{"name", "email"}}
	if err := fieldmaskutil.FilterByFieldMask(&msg, fm); err != nil {
		panic(err)
	}
}
```

Other operations at a glance:

```go
// Prune: clear the fields listed in paths, keep the rest
// {Name: "Alice", Age: 30, Email: "alice@example.com"} with age/email cleared
// => {Name: "Alice", Age: 0, Email: ""}
fieldmaskutil.Prune(user, []string{"age", "email"})

// Overwrite: overwrite the fields listed in paths in dest with values from src (rest untouched)
fieldmaskutil.Overwrite(src, dest, []string{"name", "age"})

// Validate: check that all paths are valid for the message (error if any is not)
err := fieldmaskutil.Validate(user, []string{"name", "profile.bio"})

// PathsFromFieldNumbers: field numbers -> field paths (variadic, not a slice)
paths := fieldmaskutil.PathsFromFieldNumbers(user, 1, 3) // ["name", "email"]

// NilValuePaths: return paths of fields that are not set (zero value)
nilPaths := fieldmaskutil.NilValuePaths(user, []string{"name", "age", "email"}) // ["age"]

// Normalize: convert to snake_case, id_ / _id -> id
paths = fieldmaskutil.NormalizePaths([]string{"userName", "_id"})
```

## API Reference

### Package-level helpers (fieldmaskutil.go)

Messages are modified **in place**; Filter / Prune / Overwrite have no return value:

| Function | Signature | Description |
| --- | --- | --- |
| `Filter` | `func Filter(msg proto.Message, paths []string)` | Keeps the fields listed in paths, clears all the rest |
| `Prune` | `func Prune(msg proto.Message, paths []string)` | Clears the fields listed in paths, keeps all the rest |
| `Overwrite` | `func Overwrite(src, dest proto.Message, paths []string)` | Overwrites the fields listed in paths in dest using values from src |
| `Validate` | `func Validate(validationModel proto.Message, paths []string) error` | Checks that all paths are valid; on failure returns `invalid mask: ...` |
| `PathsFromFieldNumbers` | `func PathsFromFieldNumbers(msg proto.Message, fieldNumbers ...int) []string` | **Variadic**; numbers not present in the message are skipped; no numbers returns nil |
| `NilValuePaths` | `func NilValuePaths(msg proto.Message, paths []string) []string` | Returns paths of unset fields; unknown names are skipped |
| `NormalizePaths` | `func NormalizePaths(paths []string) []string` | Converts each path to snake_case, `id_`/`_id` → `id`; **mutates the given slice in place and returns it** |

### FieldMask variants (most used in business code, fieldmaskutil.go)

These take a `*fieldmaskpb.FieldMask` and a `*proto.Message` (note: a pointer to the interface — declare `var msg proto.Message = user` first, then pass `&msg`):

| Function | Signature | Description |
| --- | --- | --- |
| `FilterByFieldMask` | `func FilterByFieldMask(msg *proto.Message, fm *fieldmaskpb.FieldMask) error` | Runs `fm.Normalize()` + `NormalizeFieldMaskPaths` (snake_case) first, then validates and filters; a nil msg or fm returns nil silently |
| `PruneByFieldMask` | `func PruneByFieldMask(msg *proto.Message, fm *fieldmaskpb.FieldMask) error` | Runs `fm.Normalize()` (sorting + removing redundant prefix paths, **no snake_case conversion**), then validates and prunes |
| `OverwriteByFieldMask` | `func OverwriteByFieldMask(msg *proto.Message, fm *fieldmaskpb.FieldMask) error` | Runs `fm.Normalize()`, validates, then executes `Overwrite(msg, msg)`: listed fields that are populated stay, unset/zero-valued ones are cleared |
| `ValidateFieldMask` | `func ValidateFieldMask(msg proto.Message, fm *fieldmaskpb.FieldMask) error` | nil fm returns nil; otherwise validates via `fm.IsValid` |
| `NormalizeFieldMaskPaths` | `func NormalizeFieldMaskPaths(fm *fieldmaskpb.FieldMask)` | Runs `fm.Normalize()` then converts each path to snake_case, `id_`/`_id` → `id` (mutates fm.Paths in place) |

### NestedMask (nested_mask.go)

Expands a flat path list into a recursive-map field mask; when the same set of paths is applied to many messages, build it once and reuse it — cheaper than the package-level functions:

| Member | Signature | Description |
| --- | --- | --- |
| `NestedMask` | `type NestedMask map[string]NestedMask` | Recursive mask type |
| `NestedMaskFromPaths` | `func NestedMaskFromPaths(paths []string) NestedMask` | `["foo.bar", "foo.baz"]` → `{"foo": {"bar": nil, "baz": nil}}`; empty/invalid segments are silently ignored |
| `(NestedMask).Filter` | `func (mask NestedMask) Filter(msg proto.Message)` | An empty mask keeps all fields |
| `(NestedMask).Prune` | `func (mask NestedMask) Prune(msg proto.Message)` | An empty mask clears nothing |
| `(NestedMask).Overwrite` | `func (mask NestedMask) Overwrite(src, dest proto.Message)` | Supports scalars, messages, repeated fields and maps; a nil parent message is initialized before overwriting; a zero-valued src field clears the dest field |
| `(NestedMask).Validate` | `func (mask NestedMask) Validate(validationModel proto.Message) error` | Validates all paths; on failure returns `invalid mask: ...` |

## Notes / Caveats

- **Paths must be normalized first**: Filter / Prune / Overwrite and the `NestedMask` methods assume paths are valid snake_case proto field names, otherwise they **may panic** (explicitly documented in the source). Masks coming from protojson are usually camelCase — run them through `NormalizePaths` / `NormalizeFieldMaskPaths` first, or use the `*ByFieldMask` family (where `FilterByFieldMask` converts to snake_case automatically).
- `PruneByFieldMask` / `OverwriteByFieldMask` only apply `fm.Normalize()` (protobuf canonicalization: sorting + removing redundant prefix paths) and do **not** convert to snake_case; camelCase paths fail validation with an `invalid field mask` error.
- `id_` and `_id` are normalized to `id` in `NormalizePaths` / `NormalizeFieldMaskPaths` (accommodating different ORM/DB primary-key spellings).
- `OverwriteByFieldMask` is equivalent to `Overwrite(msg, msg, ...)`: for the masked fields, populated ones stay and unset ones are cleared — handy for cleaning responses by mask.
- `NilValuePaths` only inspects top-level fields (looked up by name); unknown field names are skipped.
- Depends on `google.golang.org/protobuf` (already required by the root module).
