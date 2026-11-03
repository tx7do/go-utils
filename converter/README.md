# converter

English | [简体中文](README.zh-CN.md)

`converter` provides the `ApiPermissionConverter` permission-code converter: it maps kratos-style operation IDs, or an HTTP method plus a REST path, to `resource:action` style permission codes (e.g. `user:view`, `user:delete`).

## Features

- **Two entry points**: operation IDs (`ConvertCodeByOperationID`) and HTTP method + path (`ConvertCodeByPath`);
- **Resource name normalization**: singularization (via [`inflection`](https://github.com/jinzhu/inflection)) plus kebab-case (via [`stringcase`](../stringcase));
- **Action normalization**: common actions are unified into the four categories `view` / `create` / `edit` / `delete`; other actions pass through lower-cased;
- **Zero configuration**: `NewApiPermissionConverter()` is ready on construction — stateless and safe for concurrent use.

## Installation

```bash
go get github.com/tx7do/go-utils/converter
```

```go
import "github.com/tx7do/go-utils/converter"
```

## Quick Start

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/converter"
)

func main() {
	conv := converter.NewApiPermissionConverter()

	// By operation ID
	fmt.Println(conv.ConvertCodeByOperationID("UserService_List"))      // "user:view"
	fmt.Println(conv.ConvertCodeByOperationID("UserService_ListUsers")) // "user:users:view"
	fmt.Println(conv.ConvertCodeByOperationID("GetStatus"))             // "" (no "_", no match)

	// By HTTP method + path
	fmt.Println(conv.ConvertCodeByPath("GET", "/api/v1/users/{id}")) // "user:view"
	fmt.Println(conv.ConvertCodeByPath("DELETE", "/v1/users/{id}"))  // "user:delete"
}
```

## Conversion Rules

### ConvertCodeByOperationID

Input has the shape `<Service>_<Method><Name>` (regex: `^(?:([A-Za-z0-9]+)Service|([A-Za-z0-9]+))_([A-Za-z]+)([A-Za-z0-9_]*)$`). Where each capture group goes:

1. **Service segment** (capture group 1 when the `Service` suffix is present; otherwise group 2) → **singularized** → kebab-case → the first segment (resource) of the permission code;
2. **Method segment** (capture group 3, letters only) → kebab-cased, then action-normalized (see below) → the last segment (action) of the permission code;
3. **Name segment** (capture group 4, optional) → kebab-cased (**not** singularized) → when non-empty it is appended to the resource, producing a three-part `resource:name:action` code; when group 4 is empty, the part after the first camel-case word is split out of the Method segment as the Name (e.g. `ListTaskTypeName` → action=`List`, name=`TaskTypeName`).

Unmatched input or an empty string returns `""`.

Action normalization mapping (matched after kebab-casing):

| Method segment value (case-insensitive) | Normalized to |
| --- | --- |
| `list` / `get` / `retrieve` / `query` / `exist` | `view` |
| `create` / `add` / `new` | `create` |
| `update` / `edit` / `modify` / `change` | `edit` |
| `delete` / `remove` / `del` | `delete` |
| anything else | passed through lower-cased (**no forced bucket**; the switch has no default clause) |

Verified examples:

| Input | Output | Notes |
| --- | --- | --- |
| `UserService_List` | `user:view` | No Name segment, two-part code |
| `UserService_ListUsers` | `user:users:view` | Name segment `Users` is only kebab-cased, not singularized |
| `TaskService_ListTaskTypeName` | `task:task-type-name:view` | |
| `ApiService_GetWalkRouteData` | `api:walk-route-data:view` | |
| `GetStatus` | `""` | No `_`, no match |

### ConvertCodeByPath

The path goes through the following stages:

1. **Strip the version prefix** (`stripVersionPrefix`): removes a leading `api` segment and the first version segment (`v1`, `v2`, `v1.9`, ... — only when it sits within the first two segments), then filters out empty segments and any leftover `api` segments;
2. **Remove path parameters** (`removePathParams`): drops `{id}`-shaped parameter segments;
3. **Collapse to a single resource segment** (`pathToResource`): takes the **first path segment** of the processed result, singularizes each sub-token split by non-alphanumeric characters (`inflection.Singular`) and rejoins them with the original separators; if the segment contains `:`, only the part before the first `:` is kept. No kebab-casing is applied — the original casing is preserved;
4. **Action mapping** (`methodToAction`):
   - a path ending in `/list` → always `view` (takes priority over the method mapping);
   - `GET` → `view`, `POST` → `create`, `PUT` / `PATCH` → `edit`, `DELETE` → `delete` (the method is upper-cased before matching);
   - any other method → the lower-cased method name passes through (e.g. `OPTIONS` → `options`).

The result is `resource:action`. **Edge case**: an empty path, or one that becomes empty after preprocessing (e.g. `/v2`, `/{id}`), yields an empty resource and returns `:action` (e.g. `GET` + `""` → `":view"`).

Verified examples:

| Method | Path | Output |
| --- | --- | --- |
| `GET` | `/v1/users` | `user:view` |
| `GET` | `/v1/users/{id}` | `user:view` |
| `POST` | `/v1/users` | `user:create` |
| `PUT` | `/v1/users/{id}` | `user:edit` |
| `DELETE` | `/v1/users/{id}` | `user:delete` |
| `GET` | `/api/v1/admin/settings` | `admin:view` |
| `GET` | `/v1/user-groups` | `user-group:view` |
| `GET` | `/admin/v1/tasks:type-names` | `task:view` |
| `POST` | `/v1/users/list` | `view` (the `/list` suffix rule overrides POST→create) |
| `GET` | `""` | `:view` |

## API Reference

- `type ApiPermissionConverter struct{}` — a stateless converter;
- `func NewApiPermissionConverter() *ApiPermissionConverter`;
- `func (c *ApiPermissionConverter) ConvertCodeByOperationID(operationID string) string` — converts by the `<Service>_<Method><Name>` rules;
- `func (c *ApiPermissionConverter) ConvertCodeByPath(method, path string) string` — converts by the method + path rules.

(The following methods are implementation details exposed on the exported struct, useful for understanding the exact rules: `methodToAction`, `pathToResource`, `splitOperationID`, `splitCamelAfterFirstWord`, `stripVersionPrefix`, `removePathParams`, `singularizeSegments`.)

## Notes

- **Action normalization is not exhaustive**: actions not in the mapping pass through lower-cased — align the action codes your permission system accepts with the gateway/middleware conventions;
- **The Name segment is not singularized**: only kebab-case is applied to the operation ID's Name segment, so plural forms (e.g. `users` from `ListUsers`) survive into the permission code;
- **Casing**: operation-ID resources and actions always come out lower-cased via kebab-case, while the path route (`ConvertCodeByPath`) preserves the original casing of the resource segment and only singularizes it — when mixing both entry points, normalize path casing first;
- **Dependencies**: [`jinzhu/inflection`](https://github.com/jinzhu/inflection) (singularization), this repository's [`stringcase`](../stringcase) (kebab-case).

## Acknowledgements

- [jinzhu/inflection](https://github.com/jinzhu/inflection)
