# copierutil

[English](README.md) | 简体中文

`copierutil` 基于 [`jinzhu/copier`](https://github.com/jinzhu/copier) 提供一组开箱即用的类型转换器（`copier.TypeConverter`），覆盖 `time.Time`、`*string`、`timestamppb.Timestamp`、`durationpb.Duration` 之间的常见互转，并提供工厂函数把任意 `A ↔ B` 类型对包装成双向转换器，供 `copier.CopyWithOption` 的 `Converters` 选项使用。

## 特性

- **内置 6 个单向转换器**：`time.Time` / `*string` / `timestamppb.Timestamp` 两两互转，直接取用；
- **4 个现成双向转换对**：`NewTimeStringConverterPair`、`NewTimeTimestamppbConverterPair`、`NewStringTimestamppbConverterPair`、`NewDurationpbNumberConverterPair`，每对同时注册正反两个方向；
- **通用工厂**：`NewTypeConverter` / `NewTypeConverterPair`（`any` 版本）与 `NewGenericTypeConverterPair` / `NewErrorHandlingGenericTypeConverterPair`（泛型版本，类型安全、无需手动断言）；
- **底层实现复用本仓库 [`timeutil`](../timeutil)**：格式化、解析与 nil 语义与 `timeutil` 保持一致（nil 输入 → nil 输出，不报错）。

## 安装

```bash
go get github.com/tx7do/go-utils/copierutil
```

```go
import "github.com/tx7do/go-utils/copierutil"
```

依赖会一并拉取：[`jinzhu/copier`](https://github.com/jinzhu/copier)（v0.4.0）、`google.golang.org/protobuf`（`timestamppb` / `durationpb`）以及本仓库的 `timeutil`、`trans`。

## 快速开始

### 结构体互拷 + 双向转换对

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
	fmt.Println(*model.CreatedAt) // ISO8601 字符串，如 "2026-11-03T12:34:56+0800"

	// 反向：字符串 → time.Time（字符串需为可解析格式，ISO8601 产物解析不回来，见"注意事项"）
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

### 只用单向转换器

```go
copier.CopyWithOption(&model, &entity, copier.Option{
	Converters: []copier.TypeConverter{
		copierutil.TimeToStringConverter, // *time.Time → *string（"2006-01-02 15:04:05"）
	},
})
```

### 自定义类型对

```go
import (
	"time"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/timeutil"
	"github.com/tx7do/go-utils/trans"
)

func customPair() []copier.TypeConverter {
	// 泛型工厂：A ↔ B，无需手写 any 断言
	converters := copierutil.NewGenericTypeConverterPair(
		&time.Time{},  // srcType（指针零值，供 copier 反射匹配）
		trans.Ptr(""), // dstType（即 *string）
		timeutil.TimeToTimeString,
		timeutil.StringTimeToTime,
	)

	// 转换函数可能返回错误时：
	converters = copierutil.NewErrorHandlingGenericTypeConverterPair(
		&time.Time{},
		trans.Ptr(""),
		func(src *time.Time) (*string, error) { return timeutil.TimeToTimeString(src), nil },
		func(src *string) (*time.Time, error) { return timeutil.StringTimeToTime(src), nil },
	)
	return converters
}
```

## 内置转换器一览

### 单向转换器（包级变量）

| 变量 | SrcType | DstType | 行为 |
| --- | --- | --- | --- |
| `TimeToStringConverter` | `*time.Time` | `*string` | `timeutil.TimeToTimeString`，输出 `"2006-01-02 15:04:05"`（默认时区） |
| `StringToTimeConverter` | `*string` | `*time.Time` | `timeutil.StringTimeToTime`，依次尝试 `"2006-01-02 15:04:05"`、`"2006-01-02"`、`"15:04:05"`、`"2006-01-02 15:04:05.999999-07"` 四种布局 |
| `TimeToTimestamppbConverter` | `*time.Time` | `*timestamppb.Timestamp` | `timeutil.TimeToTimestamppb`（`timestamppb.New`） |
| `TimestamppbToTimeConverter` | `*timestamppb.Timestamp` | `*time.Time` | `timeutil.TimestamppbToTime`（`Timestamp.AsTime`） |
| `StringToTimestamppbConverter` | `*string` | `*timestamppb.Timestamp` | `timeutil.StringToTimestamppb`（按上表布局解析后转 Timestamp） |
| `TimestamppbToStringConverter` | `*timestamppb.Timestamp` | `*string` | `timeutil.TimestamppbToString`，输出 `"2006-01-02 15:04:05"` |

### 双向转换对（工厂函数，每对返回 `[]copier.TypeConverter` 两条）

| 工厂函数 | 类型对 | 字符串格式 |
| --- | --- | --- |
| `NewTimeStringConverterPair()` | `*time.Time ↔ *string` | 正向 ISO8601（`"2006-01-02T15:04:05Z0700"`），反向按上表四种布局解析 |
| `NewTimeTimestamppbConverterPair()` | `*time.Time ↔ *timestamppb.Timestamp` | — |
| `NewStringTimestamppbConverterPair()` | `*string ↔ *timestamppb.Timestamp` | `"2006-01-02 15:04:05"` |
| `NewDurationpbNumberConverterPair[T](timePrecision time.Duration)` | `*durationpb.Duration ↔ *T` | — |

`NewTimeStringConverterPair` 的正向使用本包 `TimeToString`（即 `timeutil.TimeToString(tm, timeutil.ISO8601)`），格式与单向的 `TimeToStringConverter` **不同**。

`NewDurationpbNumberConverterPair` 的泛型参数 `T` 支持 `int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64`；数值按 `timePrecision` 粒度解释（如传 `time.Second` 则数值表示秒数）。换算实现（见 `timeutil/trans.go`）：durationpb → 数值取 `Seconds() / timePrecision.Seconds()`；数值 → durationpb 按 `time.Duration(数值) * timePrecision`。

### 工厂函数

- `func TimeToString(tm *time.Time) *string` —— 按 `timeutil.ISO8601` 格式化；
- `func NewTypeConverter(srcType, dstType any, fn func(src any) (any, error)) copier.TypeConverter`；
- `func NewTypeConverterPair(srcType, dstType any, fromFn, toFn func(src any) (any, error)) []copier.TypeConverter`；
- `func NewGenericTypeConverterPair[A any, B any](srcType A, dstType B, fromFn func(src A) B, toFn func(src B) A) []copier.TypeConverter`；
- `func NewErrorHandlingGenericTypeConverterPair[A any, B any](srcType A, dstType B, fromFn func(src A) (B, error), toFn func(src B) (A, error)) []copier.TypeConverter`。

所有转换对工厂返回长度为 2 的切片：第 1 条是 `srcType → dstType`，第 2 条是反方向 `dstType → srcType`。

## 注意事项

- **按指针类型注册**：各转换器的 `SrcType` / `DstType` 都是 `&time.Time{}`、`trans.Ptr("")`（即 `*string`）这类指针零值，copier 通过 `reflect.TypeOf` 匹配类型，`Fn` 收到的 `src` 也是指针——字段若为非指针的 `time.Time` / `string`，不会命中这些转换器；
- **nil 语义**：底层 `timeutil` 函数对 nil / 空串输入一律返回 nil 且不报错，因此转换器把 nil 字段原样映射为 nil；
- **两种字符串格式并存**：`TimeToStringConverter` / `TimestamppbToStringConverter` / `NewStringTimestamppbConverterPair` 输出 `"2006-01-02 15:04:05"`，而 `NewTimeStringConverterPair` 与包级 `TimeToString` 输出 ISO8601（含 `T` 分隔符与时区偏移，如 `"2026-11-03T12:34:56+0800"`）；
- **`NewTimeStringConverterPair` 两方向不互逆**：反向解析只认上表四种空格分隔布局，ISO8601 产物（含 `T`）无法被解析回来，会得到 nil——需要无损耗往返时请改用 `"2006-01-02 15:04:05"` 格式的转换器或自定义 `layout`；
- **durationpb 数值方向的精度**：整数 `T` 会截断小数秒；浮点 `T` 在数值 → durationpb 方向会先被 `time.Duration(n)` 截断为整纳秒，再乘以 `timePrecision`，小数部分同样被丢弃（如 `1.5` 秒 → `1s`）；
- **时区**：`"2006-01-02 15:04:05"` 的格式化与解析均落在 `timeutil.GetDefaultTimeLocation()` 的默认时区上，跨时区数据请自行确认语义。

## 致谢

- [jinzhu/copier](https://github.com/jinzhu/copier) —— 本模块所有转换器均为其 `TypeConverter` 机制的上游适配。
