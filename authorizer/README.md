# authorizer

English | [简体中文](README.zh-CN.md)

`authorizer` is an authorization engine manager: it picks an engine from [`kratos-authz`](https://github.com/tx7do/kratos-authz) (noop / casbin / opa) according to configuration, pulls role→(path, method) policy data from a caller-supplied `Provider`, and loads it in the engine's format.

## Features

- **Pluggable engines**: switch between noop / casbin / opa with a single string field; `Engine()` returns the unified `authzEngine.Engine` interface, ready to wire into kratos-authz middleware;
- **Decoupled policies and models**: both the policy data (role→API permissions) and the policy model (e.g. a Rego script) are supplied by the caller through the `Provider` interface; this package handles translation and loading;
- **Fail-closed semantics**: a missing or broken OPA custom model never degrades into silent allow — the engine is either nil or swapped for a `denyAllEngine` that rejects everything;
- **No config, no fallback**: passing `cfg` as nil means "authorization not configured"; the engine stays nil instead of silently degrading to a permissive noop.

## Installation

```bash
go get github.com/tx7do/go-utils/authorizer
```

```go
import "github.com/tx7do/go-utils/authorizer"
```

## Quick Start

### 1. Implement a Provider

```go
package main

import (
	"context"

	"github.com/tx7do/go-utils/authorizer"
)

type myProvider struct{}

// ProvideModels supplies policy models; the OPA engine reads the "rbac.rego" key
func (myProvider) ProvideModels(engineName string) authorizer.ModelDataMap {
	if engineName == "opa" {
		return authorizer.ModelDataMap{
			"rbac.rego": []byte("package rbac\n\ndefault allow = false\n\nallow = true\n"),
		}
	}
	return nil
}

// ProvidePolicies supplies role -> API permission list
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

### 2. Create the authorizer and load policies

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

	// Translate Provider data and load it into the engine; call again after
	// roles or permissions change
	if err := authz.ResetPolicies(ctx); err != nil {
		panic(err)
	}

	engine := authz.Engine() // authzEngine.Engine, hand it to kratos-authz/middleware
	_ = engine
}
```

## Configuration (EngineConfig)

`EngineConfig` has a single field:

- `Type string` — engine type: `"noop"` | `"casbin"` | `"opa"` | `"zanzibar"`; an empty string is treated as noop.

| `cfg.Type` | Engine behavior |
| --- | --- |
| `"casbin"` | Casbin engine; policies load as a `PolicyRule{PType:"p", V0:role, V1:Path, V2:Method, V3:Domain}` table |
| `"opa"` | OPA engine; the model is read from the Provider's `rbac.rego` key, and policies load as role→`[{pattern, method}]` |
| `"zanzibar"` | The type name is accepted but the engine is **not implemented**; returns nil |
| `""` / `"noop"` | Noop engine (`ResetPolicies` returns nil immediately, nothing to load) |
| `cfg` is nil | The engine stays nil — "authorization not configured"; callers should nil-check before using `Engine()` |

Two failure modes for OPA (both fail-closed):

- **Missing model** (`ProvideModels` does not supply `rbac.rego`): the engine is nil;
- **Broken model** (fails to parse): upstream `opa.NewEngine` swallows the error and falls back to built-in policies, so this package parses the model once more explicitly; on failure it swaps in `denyAllEngine`, which rejects every check and every policy write with an observable error on each denial until the model is fixed.

## API Reference

### Data structures

- `type EngineConfig struct{ Type string }` — engine selection config;
- `type PermissionData struct{ Path, Method, Domain string }` — a single API permission;
- `type PermissionDataArray []PermissionData`;
- `type PermissionDataMap map[string]PermissionDataArray` — role code → API permission list;
- `type ModelDataMap map[string][]byte` — policy model data (OPA uses the key `rbac.rego`).

### Provider interface

- `ProvideModels(engineName string) ModelDataMap` — supplies policy models, keyed by engine name;
- `ProvidePolicies(ctx context.Context) (PermissionDataMap, error)` — supplies policy data.

### Authorizer

- `func NewAuthorizer(ctx context.Context, log bLogger.Logger, cfg *EngineConfig, provider Provider) *Authorizer` — creates the manager; nil ctx or log falls back to defaults; a nil `cfg` creates no engine;
- `func (a *Authorizer) Engine() authzEngine.Engine` — returns the underlying engine (may be nil);
- `func (a *Authorizer) ResetPolicies(ctx context.Context) error` — pulls policies from the Provider and loads them into the engine:
  - engine not initialized (nil cfg) → returns an error;
  - casbin / opa → translates to the engine format and calls `SetPolicies`;
  - noop → returns nil immediately;
  - any other engine name → returns an error;
  - Provider failure → returns the error as-is.

## Notes

- **Nil-checking the engine is the caller's job**: with a nil `cfg` or `Type: "zanzibar"`, `Engine()` returns nil — check before wiring middleware;
- **Fail-closed by design**: a broken OPA model intentionally leaves the system denying everything; the engine is chosen at construction time and `ResetPolicies` only reports "unknown engine name" on a deny-all engine, so rebuilding the Authorizer is required to recover after fixing the model;
- **Hot policy reload**: after the Provider's data changes, call `ResetPolicies` to reload — no need to rebuild the Authorizer;
- **Dependencies**: `kratos-authz` (engine interface plus casbin/opa/noop implementations), `kratos-bootstrap/logger` (logging).

## Acknowledgements

- [tx7do/kratos-authz](https://github.com/tx7do/kratos-authz)
- [casbin/casbin](https://github.com/casbin/casbin)
- [Open Policy Agent](https://www.openpolicyagent.org/)
