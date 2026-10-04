# authorizer

授权引擎管理器：按配置选择 [`kratos-authz`](https://github.com/tx7do/kratos-authz)
的引擎（noop / casbin / opa），并从使用方注入的 `Provider` 拉取
角色→(path, method) 策略数据、按引擎格式装载。

- **`NewAuthorizer(ctx, logger, cfg, provider)`**：`cfg.Type` 选择引擎，
  缺省（空串）为 noop。`cfg` 传 nil 表示"未配置授权"——引擎保持 nil、
  `ResetPolicies` 报错，**不会**静默降级为放行的 noop 引擎；
- **`Provider`** 接口：`ProvideModels`（自定义策略模型，OPA 取键
  `rbac.rego`）、`ProvidePolicies`（角色→API 权限数据）；
- **`ResetPolicies`**：把 Provider 数据转译为 casbin 的 `PolicyRule` 表或
  OPA 的 role→[{pattern,method}] JSON 装载进引擎；角色/权限变更后调用；
- **fail-closed**：OPA 自定义模型解析失败时切换 `denyAllEngine`
  全量拒绝（错误随每次判定可观测），而非回退上游内置资产策略静默上线。

```go
import "github.com/tx7do/go-utils/authorizer"

authz := authorizer.NewAuthorizer(ctx, logger, &authorizer.EngineConfig{Type: "casbin"}, myProvider)
_ = authz.ResetPolicies(ctx)
engine := authz.Engine() // 交给 kratos-authz/middleware 装配
```
