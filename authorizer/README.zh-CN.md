# authorizer

[English](README.md) | 简体中文

`authorizer` 是授权引擎管理器：按配置选择 [`kratos-authz`](https://github.com/tx7do/kratos-authz) 的引擎（noop / casbin / opa），并从使用方注入的 `Provider` 拉取角色→(path, method) 策略数据、按引擎格式装载。

## 特性

- **引擎可插拔**：通过一个字符串配置在 noop / casbin / opa 引擎间切换，`Engine()` 返回统一的 `authzEngine.Engine` 接口，可直接交给 kratos-authz 中间件装配；
- **策略与模型解耦**：策略数据（角色→API 权限）与策略模型（如 Rego 脚本）都由调用方通过 `Provider` 接口提供，本包负责转译与装载；
- **fail-closed 安全语义**：OPA 自定义模型缺失或损坏时绝不静默放行——要么引擎为 nil，要么切换 `denyAllEngine` 全量拒绝；
- **无配置不降级**：`cfg` 传 nil 表示"未配置授权"，引擎保持 nil，不会静默降级为放行的 noop。

## 安装

```bash
go get github.com/tx7do/go-utils/authorizer
```

```go
import "github.com/tx7do/go-utils/authorizer"
```

## 快速开始

### 1. 实现 Provider

```go
package main

import (
	"context"

	"github.com/tx7do/go-utils/authorizer"
)

type myProvider struct{}

// ProvideModels 提供策略模型；OPA 引擎固定取键 "rbac.rego"
func (myProvider) ProvideModels(engineName string) authorizer.ModelDataMap {
	if engineName == "opa" {
		return authorizer.ModelDataMap{
			"rbac.rego": []byte("package rbac\n\ndefault allow = false\n\nallow = true\n"),
		}
	}
	return nil
}

// ProvidePolicies 提供角色 -> API 权限列表
func (myProvider) ProvidePolicies(ctx context.Context) (authorizer.PermissionDataMap, error) {
	return authorizer.PermissionDataMap{
		"admin": []authorizer.PermissionData{
			{Path: "/api/v1/users", Method: "GET", Domain: "*"},
		},
		"viewer": []authorizer.PermissionData{
			{Path: "/api/v1/reports", Method: "GET", Domain: "*"},
		},
	}, nil
}
```

### 2. 创建授权器并装载策略

```go
package main

import (
	"context"

	bLogger "github.com/tx7do/kratos-bootstrap/logger"

	"github.com/tx7do/go-utils/authorizer"
)

func main() {
	ctx := context.Background()

	authz := authorizer.NewAuthorizer(ctx, bLogger.NopLogger(),
		&authorizer.EngineConfig{Type: "casbin"}, myProvider{})

	// 把 Provider 数据转译并装载进引擎；角色/权限变更后再次调用即可
	if err := authz.ResetPolicies(ctx); err != nil {
		panic(err)
	}

	engine := authz.Engine() // authzEngine.Engine，交给 kratos-authz/middleware 装配
	_ = engine
}
```

## 配置（EngineConfig）

`EngineConfig` 只有一个字段：

- `Type string` —— 引擎类型：`"noop"` | `"casbin"` | `"opa"` | `"zanzibar"`；空串按 noop 处理。

| `cfg.Type` | 引擎行为 |
| --- | --- |
| `"casbin"` | Casbin 引擎，策略装载为 `PolicyRule{PType:"p", V0:角色, V1:Path, V2:Method, V3:Domain}` 表 |
| `"opa"` | OPA 引擎，模型取 Provider 的 `rbac.rego` 键，策略装载为 角色→`[{pattern, method}]` |
| `"zanzibar"` | 类型名被接受但引擎**未实现**，返回 nil |
| `""` / `"noop"` | Noop 引擎（`ResetPolicies` 直接返回 nil，无需装载） |
| `cfg` 为 nil | 引擎保持 nil——"未配置授权"，调用方使用 `Engine()` 前应自行判空 |

OPA 的两种失败模式（均为 fail-closed）：

- **模型缺失**（`ProvideModels` 未提供 `rbac.rego`）：引擎为 nil；
- **模型损坏**（无法解析）：上游 `opa.NewEngine` 会吞错并回退内置策略，本包显式再解析一次，失败即切换 `denyAllEngine`——拒绝一切判定与策略写入，错误信息随每次拒绝可观测，直到运营者修复模型。

## API 参考

### 数据结构

- `type EngineConfig struct{ Type string }` —— 引擎选择配置；
- `type PermissionData struct{ Path, Method, Domain string }` —— 单条 API 权限；
- `type PermissionDataArray []PermissionData`；
- `type PermissionDataMap map[string]PermissionDataArray` —— 角色编码 → API 权限列表；
- `type ModelDataMap map[string][]byte` —— 策略模型数据（OPA 使用键 `rbac.rego`）。

### Provider 接口

- `ProvideModels(engineName string) ModelDataMap` —— 提供策略模型，按引擎名区分；
- `ProvidePolicies(ctx context.Context) (PermissionDataMap, error)` —— 提供策略数据。

### Authorizer

- `func NewAuthorizer(ctx context.Context, log bLogger.Logger, cfg *EngineConfig, provider Provider) *Authorizer` —— 创建权限管理器；ctx 或 log 为 nil 时自动使用默认值；`cfg` 为 nil 不创建引擎；
- `func (a *Authorizer) Engine() authzEngine.Engine` —— 返回底层引擎（可能为 nil）；
- `func (a *Authorizer) ResetPolicies(ctx context.Context) error` —— 从 Provider 拉取策略并装载进引擎：
  - 引擎未初始化（cfg 为 nil）→ 返回错误；
  - casbin / opa → 按引擎格式转译后 `SetPolicies`；
  - noop → 直接返回 nil；
  - 其它未知引擎名 → 返回错误；
  - Provider 报错 → 原样返回该错误。

## 注意事项

- **引擎判空责任在使用方**：`cfg` 为 nil 或 `Type` 为 `"zanzibar"` 时 `Engine()` 返回 nil，装配中间件前请检查；
- **fail-closed 优先**：OPA 模型损坏时系统处于"全部拒绝"状态是有意为之；引擎在构造时确定，`ResetPolicies` 对 deny-all 引擎只会报"未知引擎名"，修复模型后需重建 Authorizer 才能恢复；
- **策略热更新**：Provider 的数据变化后调用 `ResetPolicies` 重新装载即可，无需重建 Authorizer；
- **依赖**：`kratos-authz`（引擎接口与 casbin/opa/noop 实现）、`kratos-bootstrap/logger`（日志）。

## 致谢

- [tx7do/kratos-authz](https://github.com/tx7do/kratos-authz)
- [casbin/casbin](https://github.com/casbin/casbin)
- [Open Policy Agent](https://www.openpolicyagent.org/)
