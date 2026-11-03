# copierutil

English | [简体中文](README.zh-CN.md)

`copierutil` builds on [`jinzhu/copier`](https://github.com/jinzhu/copier) to provide a set of ready-made type converters (`copier.TypeConverter`) covering the common conversions between `time.Time`, `*string`, `timestamppb.Timestamp`, and `durationpb.Duration`, plus factory functions that wrap any `A ↔ B` type pair into bidirectional converters for use with the `Converters` option of `copier.CopyWithOption`.

## Features

- **6 built-in one-way converters**: pairwise conversions between `time.Time`, `*string`, and `timestamppb.Timestamp`, ready to use as-is;
- **4 ready-made converter pairs**: `NewTimeStringConverterPair`, `NewTimeTimestamppbConverterPair`, `NewStringTimestamppbConverterPair`, `NewDurationpbNumberConverterPair` — each registers both directions at once;
- **Generic factories**: `NewTypeConverter` / `NewTypeConverterPair` (`any`-based) and `NewGenericTypeConverterPair` / `NewErrorHandlingGenericTypeConverterPair` (generic, type-safe, no manual assertions);
- **Backed by this repository's [`timeutil`](../timeutil)**: formatting, parsing, and nil semantics stay consistent with `timeutil` (nil in → nil out, no error).

## Installation

```bash
go get github.com/tx7do/go-utils/copierutil
```

```go
import "github.com/tx7do/go-utils/copierutil"
```

Dependencies pulled in along the way: [`jinzhu/copier`](https://github.com/jinzhu/copier) (v0.4.0), `google.golang.org/protobuf` (`timestamppb` / `durationpb`), and this repository's `timeutil` and `trans`.

## Quick Start

### Struct-to-struct copying with a converter pair

```go
package main

import (
	"fmt"
	"time"

	"github.com/jinzhu/copier"

	"github.com/tx7do/go-utils/copierutil"
)

type Entity struct {
	CreatedAt *time.Time
}

type Model struct {
	CreatedAt *string
}

func main() {
	now := time.Now()
	entity := Entity{CreatedAt: &now}

	var model Model
	if err := copier.CopyWithOption(&model, &entity, copier.Option{
		Converters: copierutil.NewTimeStringConverterPair(),
	}); err != nil {
		panic(err)
	}
	fmt.Println(*model.CreatedAt) // ISO8601 string, e.g. "2026-11-03T12:34:56+0800"

	// Reverse: string → time.Time (the string must be in a parseable layout;
	// ISO8601 output cannot be parsed back — see "Notes")
	s := "2026-11-03 12:34:56"
	model2 := Model{CreatedAt: &s}
	var entity2 Entity
	if err := copier.CopyWithOption(&entity2, &model2, copier.Option{
		Converters: copierutil.NewTimeStringConverterPair(),
	}); err != nil {
		panic(err)
	}
	fmt.Println(entity2.CreatedAt.Format(time.DateOnly)) // "2026-11-03"
}
```

### Using a one-way converter only

```go
copier.CopyWithOption(&model, &entity, copier.Option{
	Converters: []copier.TypeConverter{
		copierutil.TimeToStringConverter, // *time.Time → *string ("2006-01-02 15:04:05")
	},
})
```

### Custom type pairs

```go
import (
	"time"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/timeutil"
	"github.com/tx7do/go-utils/trans"
)

func customPair() []copier.TypeConverter {
	// Generic factory: A ↔ B, no manual any assertions needed
	converters := copierutil.NewGenericTypeConverterPair(
		&time.Time{},  // srcType (zero pointer value for copier's reflection-based matching)
		trans.Ptr(""), // dstType (i.e. *string)
		timeutil.TimeToTimeString,
		timeutil.StringTimeToTime,
	)

	// When the conversion functions can return errors:
	converters = copierutil.NewErrorHandlingGenericTypeConverterPair(
		&time.Time{},
		trans.Ptr(""),
		func(src *time.Time) (*string, error) { return timeutil.TimeToTimeString(src), nil },
		func(src *string) (*time.Time, error) { return timeutil.StringTimeToTime(src), nil },
	)
	return converters
}
```

## Built-in Converters

### One-way converters (package-level variables)

| Variable | SrcType | DstType | Behavior |
| --- | --- | --- | --- |
| `TimeToStringConverter` | `*time.Time` | `*string` | `timeutil.TimeToTimeString`, outputs `"2006-01-02 15:04:05"` (default time zone) |
| `StringToTimeConverter` | `*string` | `*time.Time` | `timeutil.StringTimeToTime`, tries the layouts `"2006-01-02 15:04:05"`, `"2006-01-02"`, `"15:04:05"`, `"2006-01-02 15:04:05.999999-07"` in order |
| `TimeToTimestamppbConverter` | `*time.Time` | `*timestamppb.Timestamp` | `timeutil.TimeToTimestamppb` (`timestamppb.New`) |
| `TimestamppbToTimeConverter` | `*timestamppb.Timestamp` | `*time.Time` | `timeutil.TimestamppbToTime` (`Timestamp.AsTime`) |
| `StringToTimestamppbConverter` | `*string` | `*timestamppb.Timestamp` | `timeutil.StringToTimestamppb` (parses per the layouts above, then converts to a Timestamp) |
| `TimestamppbToStringConverter` | `*timestamppb.Timestamp` | `*string` | `timeutil.TimestamppbToString`, outputs `"2006-01-02 15:04:05"` |

### Converter pairs (factory functions; each returns a `[]copier.TypeConverter` with two entries)

| Factory | Type pair | String format |
| --- | --- | --- |
| `NewTimeStringConverterPair()` | `*time.Time ↔ *string` | Forward: ISO8601 (`"2006-01-02T15:04:05Z0700"`); reverse: parses the four layouts listed above |
| `NewTimeTimestamppbConverterPair()` | `*time.Time ↔ *timestamppb.Timestamp` | — |
| `NewStringTimestamppbConverterPair()` | `*string ↔ *timestamppb.Timestamp` | `"2006-01-02 15:04:05"` |
| `NewDurationpbNumberConverterPair[T](timePrecision time.Duration)` | `*durationpb.Duration ↔ *T` | — |

`NewTimeStringConverterPair`'s forward direction uses this package's `TimeToString` (i.e. `timeutil.TimeToString(tm, timeutil.ISO8601)`); its format **differs** from the one-way `TimeToStringConverter`.

`NewDurationpbNumberConverterPair`'s type parameter `T` supports `int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64`; the number is interpreted at `timePrecision` granularity (e.g. pass `time.Second` and the number represents seconds). The conversions (see `timeutil/trans.go`): durationpb → number takes `Seconds() / timePrecision.Seconds()`; number → durationpb computes `time.Duration(number) * timePrecision`.

### Factory functions

- `func TimeToString(tm *time.Time) *string` — formats with `timeutil.ISO8601`;
- `func NewTypeConverter(srcType, dstType any, fn func(src any) (any, error)) copier.TypeConverter`;
- `func NewTypeConverterPair(srcType, dstType any, fromFn, toFn func(src any) (any, error)) []copier.TypeConverter`;
- `func NewGenericTypeConverterPair[A any, B any](srcType A, dstType B, fromFn func(src A) B, toFn func(src B) A) []copier.TypeConverter`;
- `func NewErrorHandlingGenericTypeConverterPair[A any, B any](srcType A, dstType B, fromFn func(src A) (B, error), toFn func(src B) (A, error)) []copier.TypeConverter`.

Every pair factory returns a slice of length 2: entry 1 is `srcType → dstType`, entry 2 is the reverse `dstType → srcType`.

## Notes

- **Registered as pointer types**: every converter's `SrcType` / `DstType` is a zero pointer value such as `&time.Time{}` or `trans.Ptr("")` (i.e. `*string`); copier matches types via `reflect.TypeOf`, and `Fn` receives `src` as a pointer too — a non-pointer `time.Time` / `string` field will not hit these converters;
- **nil semantics**: the underlying `timeutil` functions return nil for nil / empty-string inputs without an error, so converters map nil fields to nil as-is;
- **Two string formats coexist**: `TimeToStringConverter` / `TimestamppbToStringConverter` / `NewStringTimestamppbConverterPair` output `"2006-01-02 15:04:05"`, while `NewTimeStringConverterPair` and the package-level `TimeToString` output ISO8601 (with the `T` separator and a time-zone offset, e.g. `"2026-11-03T12:34:56+0800"`);
- **`NewTimeStringConverterPair`'s two directions are not inverses**: the reverse direction only parses the four space-separated layouts listed above; ISO8601 output (with the `T`) cannot be parsed back and yields nil — for lossless round trips use the `"2006-01-02 15:04:05"`-format converters or a custom `layout`;
- **Durationpb number precision**: integer `T` truncates fractional seconds; for float `T` the number → durationpb direction first truncates the value to whole nanoseconds via `time.Duration(n)` before multiplying by `timePrecision`, so the fraction is dropped there too (e.g. `1.5` seconds → `1s`);
- **Time zone**: formatting and parsing of `"2006-01-02 15:04:05"` both land in `timeutil.GetDefaultTimeLocation()`'s default time zone — confirm the semantics yourself for cross-time-zone data.

## Acknowledgements

- [jinzhu/copier](https://github.com/jinzhu/copier) — every converter in this module is an upstream adaptation of its `TypeConverter` mechanism.
