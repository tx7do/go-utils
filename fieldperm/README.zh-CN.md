# fieldperm

[English](README.md) | 简体中文

字段级权限的通用消息裁剪原语：隐藏字段集来源于令牌 claim（登录期按角色并集聚合，条目格式 `"资源.字段"`，字段为 proto json_name），执行侧用它对 proto 消息做读写两向裁剪。全部基于 protoreflect 反射实现，新增受控资源无需改执行器。

## Features

- **读路径**（`ApplyReadMask` / `ApplyReadMaskList`）：就地清除命中字段。protojson 默认不输出未填充字段，清值即等于从 JSON 响应中移除该字段
- **写路径**（`StripWriteFields`）：清除请求载荷中命中字段的值，并同步剔除 field_mask 中的对应路径，防止 Update 的按 mask 过滤语义把已清空的值当作显式置零写库
- **双拼写命中**：隐藏集里放 proto 字段名（snake_case）或 json_name（camelCase）都能命中同一字段
- **泛型列表版**：`ApplyReadMaskList` 一行处理 List 响应的 items
- **纯反射实现**：不依赖具体消息类型，零代码生成

## Installation

本包为独立 Go 模块，单独安装：

```bash
go get github.com/tx7do/go-utils/fieldperm
```

```go
import "github.com/tx7do/go-utils/fieldperm"
```

## Quick Start

完整示例（`User` 消息含 `password_hash`，json_name 为 `passwordHash`）：

```go
package main

import (
	"github.com/tx7do/go-utils/fieldperm"
)

func main() {
	// 以下变量示意来自请求上下文：
	//   msg                —— 待裁剪的单个 *userpb.User 消息
	//   items              —— List 响应的 []*userpb.User
	//   req.Data           —— 写请求载荷 *userpb.User
	//   req.UpdateMask     —— 写请求的 *fieldmaskpb.FieldMask

	// 令牌 claim 中的隐藏字段条目，格式 "资源.字段"（字段建议用 json_name 拼写）
	entries := []string{"User.passwordHash", "User.salt"}

	// 取出目标资源的隐藏字段集
	fields := fieldperm.HiddenFieldsOf(entries, "User")

	// 读路径：就地清除 User 消息中的命中字段
	fieldperm.ApplyReadMask(msg, fields)

	// List 响应：对每个元素应用读裁剪
	fieldperm.ApplyReadMaskList(items, fields)

	// 写路径：清除载荷值 + 剔除 field_mask 中的对应路径
	fieldperm.StripWriteFields(req.Data, req.UpdateMask, fields)
}
```

也可以先解析再按资源取用：

```go
parsed := fieldperm.ParseTokenEntries(entries) // map[资源]字段集合
fields := parsed["User"]
```

## API 参考

| 函数 | 签名 | 说明 |
| --- | --- | --- |
| `ParseTokenEntries` | `func ParseTokenEntries(entries []string) map[string]map[string]struct{}` | 把 `"资源.字段"` 条目解析为 资源名 → 字段名集合；条目**按原样存储**（不做拼写转换）；首个 `.` 切分，无点、资源名或字段名为空的条目静默跳过 |
| `HiddenFieldsOf` | `func HiddenFieldsOf(entries []string, resource string) map[string]struct{}` | 等价于 `ParseTokenEntries(entries)[resource]`；无配置返回 nil |
| `IsEmpty` | `func IsEmpty(hidden map[string]struct{}) bool` | 隐藏集是否为空（空集 = 无需裁剪） |
| `ApplyReadMask` | `func ApplyReadMask(msg proto.Message, hidden map[string]struct{})` | 就地清除消息**顶层**命中字段；先按 proto 字段名匹配，再按 json_name 匹配（双拼写命中）；msg 为 nil 或反射消息无效时安静返回；空集 no-op |
| `ApplyReadMaskList` | `func ApplyReadMaskList[T proto.Message](items []T, hidden map[string]struct{})` | 泛型列表版，对每个元素应用读裁剪；空集整体 no-op |
| `StripWriteFields` | `func StripWriteFields(data proto.Message, fieldMask *fieldmaskpb.FieldMask, hidden map[string]struct{})` | 先对 data 应用读裁剪（双拼写清值），再从 `fieldMask.Paths` 中剔除命中项；路径按**顶层段**与隐藏集逐字匹配（`field.sub` 顶层段 `field` 命中即整条剔除，拼写须与条目一致）；fieldMask 为 nil 时仅清值；空集 no-op |

## Notes / Caveats

- **只裁剪顶层字段，不递归**：`ApplyReadMask` 仅遍历消息顶层字段，嵌套消息内部不受影响——本原语只管顶层业务字段。
- **双拼写命中的位置**：`ParseTokenEntries` 把条目按原样存入集合（不做拼写转换）；json_name（camelCase）与 proto 名（snake_case）的双拼写命中发生在 `ApplyReadMask` 里（先试 `fd.Name()`，不同再试 `fd.JSONName()`）。因此**清值**层面两种拼写任选其一即可命中。
- **mask 路径剔除按原样匹配**：`StripWriteFields` 剔除 `fieldMask.Paths` 时，用路径顶层段与隐藏集**逐字比对**——隐藏集里只有实际存入的拼写才会被剔除（测试用例佐证：隐藏集存 `createdBy` 时，mask 路径 `created_by` 不会被剔除，仅字段值被清）。为保证写路径完整防护，令牌条目建议同时收录两种拼写（如 `User.passwordHash` 与 `User.password_hash`）。
- protojson 序列化默认省略未填充字段，所以读路径「清值」即「从 JSON 响应中移除」；若你的序列化配置显式输出零值字段（`EmitUnpopulated`），清值后字段仍会出现——此时请勿开启该选项。
- `fieldMask.Paths` 被原地修改（`StripWriteFields` 直接复用底层切片），传入前无需拷贝但不要共享。

## Tests

在模块目录内运行：

```bash
cd fieldperm
go test ./...
```
