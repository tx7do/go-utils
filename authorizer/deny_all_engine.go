package authorizer

import (
	"context"
	"errors"

	authzEngine "github.com/tx7do/kratos-authz/engine"
)

// errDenyAllEngine 兜底引擎的统一拒绝错误。
var errDenyAllEngine = errors.New("authorization engine unavailable: custom policy model failed to load")

// denyAllEngine 拒绝一切判定与策略写入的兜底引擎。
//
// 上游 opa.NewEngine 对无法解析的自定义模型吞掉错误并回退编译内置资产策略、
// 仍返回非 nil 引擎——运营者部署了损坏的自定义模型时会带着一套并非本意的
// 默认策略静默上线。该情形下改用本引擎：全部请求拒绝（fail-closed），
// 错误信息随每次拒绝可观测。
type denyAllEngine struct{}

func (denyAllEngine) Name() string { return "deny-all" }

func (denyAllEngine) IsAuthorized(
	_ context.Context,
	_ authzEngine.Subject,
	_ authzEngine.Action,
	_ authzEngine.Resource,
	_ authzEngine.Project,
) (bool, error) {
	return false, errDenyAllEngine
}

func (denyAllEngine) ProjectsAuthorized(
	_ context.Context,
	_ authzEngine.Subjects,
	_ authzEngine.Action,
	_ authzEngine.Resource,
	_ authzEngine.Projects,
) (authzEngine.Projects, error) {
	return nil, errDenyAllEngine
}

func (denyAllEngine) FilterAuthorizedPairs(
	_ context.Context,
	_ authzEngine.Subjects,
	_ authzEngine.Pairs,
) (authzEngine.Pairs, error) {
	return nil, errDenyAllEngine
}

func (denyAllEngine) FilterAuthorizedProjects(
	_ context.Context,
	_ authzEngine.Subjects,
) (authzEngine.Projects, error) {
	return nil, errDenyAllEngine
}

func (denyAllEngine) SetPolicies(
	_ context.Context,
	_ authzEngine.PolicyMap,
	_ authzEngine.RoleMap,
) error {
	return errDenyAllEngine
}
