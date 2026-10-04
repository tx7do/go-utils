package authorizer

import (
	"context"
	"fmt"

	authzEngine "github.com/tx7do/kratos-authz/engine"
	"github.com/tx7do/kratos-authz/engine/casbin"
	"github.com/tx7do/kratos-authz/engine/noop"
	"github.com/tx7do/kratos-authz/engine/opa"

	bLogger "github.com/tx7do/kratos-bootstrap/logger"
)

// EngineConfig 引擎选择配置。
type EngineConfig struct {
	// Type 引擎类型："noop" | "casbin" | "opa" | "zanzibar"；空串按 noop 处理。
	Type string
}

// Authorizer 权限管理器：按配置选择授权引擎，从 Provider 拉取
// 角色→(path,method) 策略并按引擎格式装载。
type Authorizer struct {
	log *bLogger.Helper

	engine   authzEngine.Engine
	provider Provider
}

// NewAuthorizer 创建权限管理器。
//
// cfg 为 nil 表示"未配置授权"，此时引擎保持 nil（与配置缺失语义一致，
// 调用方应在使用 Engine() 前自行判空）——不静默降级为放行的 noop。
func NewAuthorizer(
	ctx context.Context,
	log bLogger.Logger,
	cfg *EngineConfig,
	provider Provider,
) *Authorizer {
	if ctx == nil {
		ctx = context.Background()
	}
	if log == nil {
		log = bLogger.GetLogger()
	}

	a := &Authorizer{
		log:      bLogger.NewHelper(log.With("module", "authorizer")),
		provider: provider,
	}

	a.init(ctx, cfg)

	return a
}

func (a *Authorizer) init(ctx context.Context, cfg *EngineConfig) {
	if cfg == nil {
		a.log.Warn(ctx, "authorization config is nil")
		return
	}
	a.engine = a.newEngine(ctx, cfg)
}

func (a *Authorizer) Engine() authzEngine.Engine {
	return a.engine
}

// ResetPolicies 重置策略
func (a *Authorizer) ResetPolicies(ctx context.Context) error {
	if a.engine == nil {
		return fmt.Errorf("authorization engine is not initialized")
	}

	result, err := a.provider.ProvidePolicies(ctx)
	if err != nil {
		a.log.Errorf(ctx, "provide authorizer data error: %v", err)
		return err
	}

	var policies authzEngine.PolicyMap

	switch a.engine.Name() {
	case "casbin":
		if policies, err = a.generateCasbinPolicies(result); err != nil {
			a.log.Errorf(ctx, "generate casbin policies error: %v", err)
			return err
		}

	case "opa":
		if policies, err = a.generateOpaPolicies(result); err != nil {
			a.log.Errorf(ctx, "generate OPA policies error: %v", err)
			return err
		}

	case "noop":
		return nil

	default:
		err = fmt.Errorf("unknown engine name: %s", a.engine.Name())
		a.log.Warnf(ctx, "%s", err.Error())
		return err
	}

	if err = a.engine.SetPolicies(ctx, policies, nil); err != nil {
		a.log.Errorf(ctx, "set policies error: %v", err)
		return err
	}

	a.log.Infof(ctx, "reloaded policy rules [%d] successfully for engine: %s", len(policies), a.engine.Name())

	return nil
}

// generateCasbinPolicies 生成 Casbin 策略
func (a *Authorizer) generateCasbinPolicies(data PermissionDataMap) (authzEngine.PolicyMap, error) {
	var rules []casbin.PolicyRule

	for roleCode, aRules := range data {
		for _, api := range aRules {
			rules = append(rules, casbin.PolicyRule{
				PType: "p",
				V0:    roleCode,
				V1:    api.Path,
				V2:    api.Method,
				V3:    api.Domain,
			})
		}
	}

	policies := authzEngine.PolicyMap{
		"policies": rules,
		"projects": authzEngine.MakeProjects(),
	}

	return policies, nil
}

// generateOpaPolicies 生成 OPA 策略
func (a *Authorizer) generateOpaPolicies(data PermissionDataMap) (authzEngine.PolicyMap, error) {
	type OpaPolicyPath struct {
		Pattern string `json:"pattern"`
		Method  string `json:"method"`
	}

	policies := make(authzEngine.PolicyMap, len(data))

	for roleCode, aRule := range data {
		paths := make([]OpaPolicyPath, 0, len(aRule))

		for _, api := range aRule {
			paths = append(paths, OpaPolicyPath{
				Pattern: api.Path,
				Method:  api.Method,
			})
		}

		policies[roleCode] = paths
	}

	return policies, nil
}

// newEngine 创建权限引擎
func (a *Authorizer) newEngine(ctx context.Context, cfg *EngineConfig) authzEngine.Engine {
	switch cfg.Type {
	case "casbin":
		return a.newEngineCasbin(ctx)

	case "opa":
		return a.newEngineOPA(ctx)

	case "zanzibar":
		return a.newEngineZanzibar(ctx)

	default:
		fallthrough
	case "", "noop":
		return a.newEngineNoop(ctx)
	}
}

// newEngineZanzibar 创建 Zanzibar 引擎（未实现）
func (a *Authorizer) newEngineZanzibar(_ context.Context) authzEngine.Engine {
	return nil
}

// newEngineNoop 创建 Noop 引擎
func (a *Authorizer) newEngineNoop(ctx context.Context) authzEngine.Engine {
	state, err := noop.NewEngine(ctx)
	if err != nil {
		a.log.Errorf(ctx, "new noop engine error: %v", err)
		return nil
	}
	return state
}

// newEngineCasbin 创建 Casbin 引擎
func (a *Authorizer) newEngineCasbin(ctx context.Context) authzEngine.Engine {
	state, err := casbin.NewEngine(ctx)
	if err != nil {
		a.log.Errorf(ctx, "init casbin engine error: %v", err)
		return nil
	}
	return state
}

// newEngineOPA 创建 OPA 引擎
func (a *Authorizer) newEngineOPA(ctx context.Context) authzEngine.Engine {
	modelName := "rbac.rego"
	models := a.provider.ProvideModels("opa")
	var model []byte
	var ok bool
	if model, ok = models[modelName]; ok {
		a.log.Infof(ctx, "load custom OPA model: %s", modelName)
	} else {
		a.log.Errorf(ctx, "OPA model not found: %s", modelName)
		return nil
	}

	state, err := opa.NewEngine(ctx,
		opa.WithModulesFromString(map[string]string{
			modelName: string(model),
		}),
	)
	if err != nil {
		a.log.Errorf(ctx, "init opa engine error: %v", err)
		return nil
	}

	// 模型有效性探测：上游 WithModulesFromString 对无法解析的自定义模型吞错，
	// NewEngine 随即回退编译内置资产策略并正常返回引擎。此处显式再走一次解析：
	// 出错即证明引擎当前跑的是内置资产策略而非运营者部署的模型——该分支
	// 只记日志仍返回引擎会带错误语义静默上线，改为返回 denyAllEngine
	// 全量拒绝（fail-closed），并高声报错提示运营者修复模型。
	if err = state.InitModulesFromString(map[string]string{
		modelName: string(model),
	}); err != nil {
		a.log.Errorf(ctx,
			"custom OPA model [%s] failed to parse, engine rejected (all requests will be denied until fixed): %v",
			modelName, err)
		return denyAllEngine{}
	}

	return state
}
