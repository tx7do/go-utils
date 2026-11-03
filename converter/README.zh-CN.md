# converter

[English](README.md) | 简体中文

`converter` 提供权限码转换器 `ApiPermissionConverter`：把 kratos 风格的 operation ID 或 HTTP 方法 + REST 路径映射为 `resource:action` 风格的权限码（如 `user:view`、`user:delete`）。

## 特性

- **两种入口**：operation ID（`ConvertCodeByOperationID`）与 HTTP 方法 + 路径（`ConvertCodeByPath`）；
- **资源名归一**：单数化（基于 [`inflection`](https://github.com/jinzhu/inflection)）+ kebab-case（基于 [`stringcase`](../stringcase)）；
- **动作归一**：常见动作统一为 `view` / `create` / `edit` / `delete` 四类，其余动作小写透传；
- **零配置**：`NewApiPermissionConverter()` 即建即用，无状态、并发安全。

## 安装

```bash
go get github.com/tx7do/go-utils/converter
```

```go
import "github.com/tx7do/go-utils/converter"
```

## 快速开始

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/converter"
)

func main() {
	conv := converter.NewApiPermissionConverter()

	// operation ID 方式
	fmt.Println(conv.ConvertCodeByOperationID("UserService_List"))     // "user:view"
	fmt.Println(conv.ConvertCodeByOperationID("UserService_ListUsers")) // "user:users:view"
	fmt.Println(conv.ConvertCodeByOperationID("GetStatus"))            // ""（不含 "_"，不匹配）

	// HTTP 方法 + 路径方式
	fmt.Println(conv.ConvertCodeByPath("GET", "/api/v1/users/{id}")) // "user:view"
	fmt.Println(conv.ConvertCodeByPath("DELETE", "/v1/users/{id}"))  // "user:delete"
}
```

## 转换规则

### ConvertCodeByOperationID

输入形如 `<Service>_<Method><Name>`（正则：`^(?:([A-Za-z0-9]+)Service|([A-Za-z0-9]+))_([A-Za-z]+)([A-Za-z0-9_]*)$`）。各捕获组的去向：

1. **Service 段**（捕获组 1，带 `Service` 后缀时；否则组 2）→ **单数化** → kebab-case → 权限码第一段 resource；
2. **Method 段**（捕获组 3，纯字母）→ kebab-case 后做动作归一（见下）→ 权限码最后一段 action；
3. **Name 段**（捕获组 4，可选）→ kebab-case（**不单数化**）→ 非空时拼进 resource，形成 `resource:name:action` 三段码；组 4 为空时，从 Method 段按驼峰边界拆出第一个词之后的部分充当 Name（如 `ListTaskTypeName` → action=`List`、name=`TaskTypeName`）。

不匹配正则或空输入返回空串 `""`。

动作归一映射（kebab 化之后匹配）：

| Method 段取值（不区分大小写） | 归一为 |
| --- | --- |
| `list` / `get` / `retrieve` / `query` / `exist` | `view` |
| `create` / `add` / `new` | `create` |
| `update` / `edit` / `modify` / `change` | `edit` |
| `delete` / `remove` / `del` | `delete` |
| 其余取值 | 原样小写透传（**不强制归类**，switch 无 default） |

实证样例：

| 输入 | 输出 | 说明 |
| --- | --- | --- |
| `UserService_List` | `user:view` | 无 Name 段，两段码 |
| `UserService_ListUsers` | `user:users:view` | Name 段 `Users` 只做 kebab，不单数化 |
| `TaskService_ListTaskTypeName` | `task:task-type-name:view` | |
| `ApiService_GetWalkRouteData` | `api:walk-route-data:view` | |
| `GetStatus` | `""` | 不含 `_`，不匹配 |

### ConvertCodeByPath

对路径依次执行：

1. **剥离版本前缀**（`stripVersionPrefix`）：去掉开头的 `api` 段与第一个版本段（`v1`、`v2`、`v1.9` 等，仅当其位于前两段内），再过滤空段与残余 `api` 段；
2. **剔除路径参数**（`removePathParams`）：删除 `{id}` 形态的参数段；
3. **折叠为单段资源**（`pathToResource`）：取处理后的**第一个路径段**，按非字母数字字符分段各自单数化（`inflection.Singular`）后保留原分隔符拼回；若该段含 `:`，只取 `:` 前的部分。不做 kebab 转换，大小写保持原样；
4. **动作映射**（`methodToAction`）：
   - 路径以 `/list` 结尾 → 一律 `view`（优先于方法映射）；
   - `GET` → `view`，`POST` → `create`，`PUT` / `PATCH` → `edit`，`DELETE` → `delete`（方法名先转大写再匹配）；
   - 其余方法 → 小写方法名透传（如 `OPTIONS` → `options`）。

最终返回 `resource:action`。**边界**：路径为空、或预处理后为空（如 `/v2`、`/{id}`）时 resource 为空串，返回 `:action`（如 `GET` + `""` → `":view"`）。

实证样例：

| Method | Path | 输出 |
| --- | --- | --- |
| `GET` | `/v1/users` | `user:view` |
| `GET` | `/v1/users/{id}` | `user:view` |
| `POST` | `/v1/users` | `user:create` |
| `PUT` | `/v1/users/{id}` | `user:edit` |
| `DELETE` | `/v1/users/{id}` | `user:delete` |
| `GET` | `/api/v1/admin/settings` | `admin:view` |
| `GET` | `/v1/user-groups` | `user-group:view` |
| `GET` | `/admin/v1/tasks:type-names` | `task:view` |
| `POST` | `/v1/users/list` | `view`（`/list` 后缀特判，覆盖 POST→create） |
| `GET` | `""` | `:view` |

## API 参考

- `type ApiPermissionConverter struct{}` —— 无状态转换器；
- `func NewApiPermissionConverter() *ApiPermissionConverter`；
- `func (c *ApiPermissionConverter) ConvertCodeByOperationID(operationID string) string` —— 按 `<Service>_<Method><Name>` 规则转换；
- `func (c *ApiPermissionConverter) ConvertCodeByPath(method, path string) string` —— 按方法 + 路径规则转换。

（以下方法为导出结构上的内部实现细节，供深入了解规则时参考：`methodToAction`、`pathToResource`、`splitOperationID`、`splitCamelAfterFirstWord`、`stripVersionPrefix`、`removePathParams`、`singularizeSegments`。）

## 注意事项

- **动作归一并非穷举**：未命中映射表的动作原样小写透传，权限系统接入了哪些动作码需与网关/中间件的约定对齐；
- **Name 段不单数化**：operation ID 的 Name 段只做 kebab-case，复数形式（如 `ListUsers` 的 `users`）会保留在权限码中；
- **大小写**：operation ID 路径的 resource/action 全部经 kebab-case 变为小写；而路径方式（`ConvertCodeByPath`）的 resource 段保留原大小写，仅单数化——若两套入口混用，建议先统一路径大小写；
- **依赖**：[`jinzhu/inflection`](https://github.com/jinzhu/inflection)（单数化）、本仓库 [`stringcase`](../stringcase)（kebab-case）。

## 致谢

- [jinzhu/inflection](https://github.com/jinzhu/inflection)
