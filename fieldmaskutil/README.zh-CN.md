# fieldmaskutil

[English](README.md) | 简体中文

Protobuf FieldMask 工具包：按字段路径或 `*fieldmaskpb.FieldMask` 对 proto 消息做保留（Filter）、清除（Prune）、覆盖（Overwrite），支持嵌套路径、列表/Map 字段，并提供路径校验与规范化能力，适合 PATCH 局部更新、响应字段裁剪等场景。

> **模块归属**：本包是根模块 `github.com/tx7do/go-utils` 的子包（**没有独立的 go.mod**），import 路径为 `github.com/tx7do/go-utils/fieldmaskutil`。安装时只能安装整个根模块——按子路径 `go get github.com/tx7do/go-utils/fieldmaskutil` 不受本仓库模块布局支持。

## Features

- **三种核心操作**：`Filter`（保留指定路径、清空其余）、`Prune`（清除指定路径、保留其余）、`Overwrite`（用 src 的值覆盖 dest 中指定路径）
- **FieldMask 直通版**：`FilterByFieldMask` / `PruneByFieldMask` / `OverwriteByFieldMask`，接收 `*fieldmaskpb.FieldMask`，自动规范化 + 校验，业务代码最常用
- **嵌套路径支持**：`a.b.c` 点分路径递归作用到嵌套消息，支持 repeated 字段与 map 字段
- **校验**：`Validate` / `ValidateFieldMask` 检查路径在消息上是否有效
- **规范化**：`NormalizePaths` / `NormalizeFieldMaskPaths` 统一转 snake_case，并把 `id_` / `_id` 归一为 `id`
- **辅助查询**：`PathsFromFieldNumbers`（字段编号 → 路径）、`NilValuePaths`（找出未设置的字段路径）

## Installation

本包属于根模块 `github.com/tx7do/go-utils`，安装整个模块：

```bash
go get github.com/tx7do/go-utils
```

```go
import "github.com/tx7do/go-utils/fieldmaskutil"
```

## Quick Start

完整示例（假设定义了 `message User { string name = 1; int32 age = 2; string email = 3; Profile profile = 4; }` 与 `message Profile { string bio = 1; }`）：

```go
package main

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/tx7do/go-utils/fieldmaskutil"
)

func main() {
	user := getUser() // *userpb.User，示意：来自请求上下文的待处理消息
	// {Name: "Alice", Age: 30, Email: "alice@example.com"}

	// Filter：保留 paths 列出的字段，清空其余（就地修改，无返回值）
	// {Name: "Alice", Age: 0, Email: "alice@example.com"}
	fieldmaskutil.Filter(user, []string{"name", "email"})

	// FieldMask 版（推荐业务代码使用）：先规范化 + 校验，再过滤
	var msg proto.Message = user
	fm := &fieldmaskpb.FieldMask{Paths: []string{"name", "email"}}
	if err := fieldmaskutil.FilterByFieldMask(&msg, fm); err != nil {
		panic(err)
	}
}
```

其它操作速览：

```go
// Prune：清除 paths 列出的字段，保留其余
// {Name: "Alice", Age: 30, Email: "alice@example.com"} 清除 age/email 后
// => {Name: "Alice", Age: 0, Email: ""}
fieldmaskutil.Prune(user, []string{"age", "email"})

// Overwrite：用 src 的值覆盖 dest 中 paths 列出的字段（其余保持不变）
fieldmaskutil.Overwrite(src, dest, []string{"name", "age"})

// Validate：检查所有路径在消息上是否有效（不存在则返回错误）
err := fieldmaskutil.Validate(user, []string{"name", "profile.bio"})

// PathsFromFieldNumbers：字段编号 → 字段路径（变参，不是切片）
paths := fieldmaskutil.PathsFromFieldNumbers(user, 1, 3) // ["name", "email"]

// NilValuePaths：返回未设置（零值）的字段路径
nilPaths := fieldmaskutil.NilValuePaths(user, []string{"name", "age", "email"}) // ["age"]

// 规范化：统一 snake_case，id_ / _id 归一为 id
paths = fieldmaskutil.NormalizePaths([]string{"userName", "_id"})
```

## API 参考

### 包级便捷函数（fieldmaskutil.go）

对消息**就地修改**，Filter / Prune / Overwrite 均无返回值：

| 函数 | 签名 | 说明 |
| --- | --- | --- |
| `Filter` | `func Filter(msg proto.Message, paths []string)` | 保留 paths 列出的字段，清空其余 |
| `Prune` | `func Prune(msg proto.Message, paths []string)` | 清除 paths 列出的字段，保留其余 |
| `Overwrite` | `func Overwrite(src, dest proto.Message, paths []string)` | 用 src 的值覆盖 dest 中 paths 列出的字段 |
| `Validate` | `func Validate(validationModel proto.Message, paths []string) error` | 检查所有路径是否有效，失败返回 `invalid mask: ...` |
| `PathsFromFieldNumbers` | `func PathsFromFieldNumbers(msg proto.Message, fieldNumbers ...int) []string` | **变参**；编号在消息中不存在时跳过该编号；不传编号返回 nil |
| `NilValuePaths` | `func NilValuePaths(msg proto.Message, paths []string) []string` | 返回未设置的字段路径；未知名跳过 |
| `NormalizePaths` | `func NormalizePaths(paths []string) []string` | 每个路径转 snake_case，`id_`/`_id` → `id`；**原地修改传入切片并返回它** |

### FieldMask 版（业务最常用，fieldmaskutil.go）

接收 `*fieldmaskpb.FieldMask` 与 `*proto.Message`（注意是指向接口的指针，调用时先声明 `var msg proto.Message = user` 再传 `&msg`）：

| 函数 | 签名 | 说明 |
| --- | --- | --- |
| `FilterByFieldMask` | `func FilterByFieldMask(msg *proto.Message, fm *fieldmaskpb.FieldMask) error` | 先 `fm.Normalize()` + `NormalizeFieldMaskPaths`（转 snake_case）再校验并过滤；msg 或 fm 为 nil 时静默返回 nil |
| `PruneByFieldMask` | `func PruneByFieldMask(msg *proto.Message, fm *fieldmaskpb.FieldMask) error` | `fm.Normalize()`（排序 + 去冗余前缀，**不转 snake_case**）后校验并清除 |
| `OverwriteByFieldMask` | `func OverwriteByFieldMask(msg *proto.Message, fm *fieldmaskpb.FieldMask) error` | `fm.Normalize()` 后校验，再执行 `Overwrite(msg, msg)`：mask 列出的字段已填充则保留，未填充/零值则清除 |
| `ValidateFieldMask` | `func ValidateFieldMask(msg proto.Message, fm *fieldmaskpb.FieldMask) error` | fm 为 nil 返回 nil；否则按 `fm.IsValid` 校验 |
| `NormalizeFieldMaskPaths` | `func NormalizeFieldMaskPaths(fm *fieldmaskpb.FieldMask)` | `fm.Normalize()` 后逐路径转 snake_case，`id_`/`_id` → `id`（原地修改 fm.Paths） |

### NestedMask（nested_mask.go）

把扁平路径列表展开为递归 map 形式的字段掩码；同一路径集合处理多条消息时，构造一次复用即可，比包级函数少一次构建：

| 成员 | 签名 | 说明 |
| --- | --- | --- |
| `NestedMask` | `type NestedMask map[string]NestedMask` | 递归掩码类型 |
| `NestedMaskFromPaths` | `func NestedMaskFromPaths(paths []string) NestedMask` | `["foo.bar", "foo.baz"]` → `{"foo": {"bar": nil, "baz": nil}}`；空段/非法输入静默忽略 |
| `(NestedMask).Filter` | `func (mask NestedMask) Filter(msg proto.Message)` | 掩码为空时保留全部字段 |
| `(NestedMask).Prune` | `func (mask NestedMask) Prune(msg proto.Message)` | 掩码为空时不清除任何字段 |
| `(NestedMask).Overwrite` | `func (mask NestedMask) Overwrite(src, dest proto.Message)` | 支持 scalar、message、repeated、map；父级消息为 nil 时自动初始化；src 字段为零值时清除 dest 对应字段 |
| `(NestedMask).Validate` | `func (mask NestedMask) Validate(validationModel proto.Message) error` | 校验所有路径，失败返回 `invalid mask: ...` |

## Notes / Caveats

- **路径必须先规范化**：Filter / Prune / Overwrite 及 `NestedMask` 方法假定路径已是合法的 snake_case proto 字段名，否则**可能 panic**（源码注释明确说明）。来自 protojson 的 mask 通常是 camelCase，请先过 `NormalizePaths` / `NormalizeFieldMaskPaths`，或直接用 `*ByFieldMask` 系列（其中 `FilterByFieldMask` 会自动转 snake_case）。
- `PruneByFieldMask` / `OverwriteByFieldMask` 只做 `fm.Normalize()`（protobuf 规范化：排序 + 去除冗余前缀路径），**不做 snake_case 转换**，camelCase 路径会因校验失败而返回 `invalid field mask` 错误。
- `id_` 与 `_id` 在 `NormalizePaths` / `NormalizeFieldMaskPaths` 中被归一为 `id`（兼容不同 ORM/DB 的主键命名）。
- `OverwriteByFieldMask` 等价于 `Overwrite(msg, msg, ...)`，效果是「mask 里的字段：有值保留、无值清除」，常用于按 mask 清洗响应。
- `NilValuePaths` 只查顶层字段（按名字查找），未知的字段名直接跳过。
- 依赖 `google.golang.org/protobuf`（根模块已引入）。
