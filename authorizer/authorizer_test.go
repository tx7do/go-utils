package authorizer

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bLogger "github.com/tx7do/kratos-bootstrap/logger"
)

// stubProvider 空模型/空策略桩。
type stubProvider struct{}

func (stubProvider) ProvideModels(string) ModelDataMap { return nil }

func (stubProvider) ProvidePolicies(context.Context) (PermissionDataMap, error) {
	return PermissionDataMap{}, nil
}

// policiesProvider 返回固定策略与模型的桩。
type policiesProvider struct{ stubProvider }

func (policiesProvider) ProvidePolicies(context.Context) (PermissionDataMap, error) {
	return PermissionDataMap{
		"admin": []PermissionData{{Path: "/api/v1/users", Method: "GET"}},
	}, nil
}

func (policiesProvider) ProvideModels(string) ModelDataMap {
	return ModelDataMap{"rbac.rego": []byte("package rbac\n\nallow = true\n")}
}

func TestNewAuthorizer_NilConfigKeepsEngineNil(t *testing.T) {
	a := NewAuthorizer(context.Background(), bLogger.NopLogger(), nil, stubProvider{})
	require.NotNil(t, a)
	assert.Nil(t, a.Engine(), "未配置授权时引擎必须保持 nil，不得静默降级为放行")
	assert.Error(t, a.ResetPolicies(context.Background()), "引擎未初始化时重载策略必须报错")
}

func TestNewAuthorizer_NoopEngine(t *testing.T) {
	a := NewAuthorizer(context.Background(), bLogger.NopLogger(), &EngineConfig{Type: "noop"}, stubProvider{})
	require.NotNil(t, a)
	require.NotNil(t, a.Engine())
	assert.Equal(t, "noop", a.Engine().Name())
	assert.NoError(t, a.ResetPolicies(context.Background()), "noop 引擎无需装载策略")
}

func TestNewAuthorizer_EmptyTypeDefaultsNoop(t *testing.T) {
	a := NewAuthorizer(context.Background(), bLogger.NopLogger(), &EngineConfig{}, stubProvider{})
	require.NotNil(t, a.Engine())
	assert.Equal(t, "noop", a.Engine().Name())
}

func TestNewAuthorizer_OPAValidModel(t *testing.T) {
	a := NewAuthorizer(context.Background(), bLogger.NopLogger(), &EngineConfig{Type: "opa"}, policiesProvider{})
	require.NotNil(t, a.Engine())
	assert.Equal(t, "opa", a.Engine().Name())
	require.NoError(t, a.ResetPolicies(context.Background()))
}

func TestNewAuthorizer_OPAMissingModelFailsClosed(t *testing.T) {
	// 模型缺失：newEngineOPA 返回 nil 引擎
	a := NewAuthorizer(context.Background(), bLogger.NopLogger(), &EngineConfig{Type: "opa"}, stubProvider{})
	assert.Nil(t, a.Engine(), "OPA 模型缺失时不得返回可用引擎")
}

func TestNewAuthorizer_OPABrokenModelDeniesAll(t *testing.T) {
	// 模型存在但语法损坏：必须落到 denyAllEngine（fail-closed），而非
	// 静默回退上游内置资产策略
	// 契约：损坏模型必须 fail-closed——要么引擎直接创建失败（nil），
	// 要么切换到 deny-all 兜底引擎（拒绝一切判定与策略写入），
	// 绝不允许带着上游内置资产策略静默放行。
	broken := brokenModelProvider{}
	a := NewAuthorizer(context.Background(), bLogger.NopLogger(), &EngineConfig{Type: "opa"}, broken)
	if eng := a.Engine(); eng != nil {
		assert.Equal(t, "deny-all", eng.Name(), "损坏模型必须切换到 deny-all 兜底引擎")

		ok, err := eng.IsAuthorized(context.Background(), "", "", "", "")
		assert.False(t, ok)
		assert.Error(t, err)

		assert.Error(t, eng.SetPolicies(context.Background(), nil, nil),
			"deny-all 引擎必须拒绝策略写入")
	} else {
		t.Log("上游 opa.NewEngine 对损坏模型直接报错（引擎为 nil），同为 fail-closed 语义")
	}
}

func TestNewAuthorizer_NilLoggerAndContext(t *testing.T) {
	assert.NotPanics(t, func() {
		a := NewAuthorizer(nil, nil, &EngineConfig{Type: "noop"}, stubProvider{})
		assert.NotNil(t, a.Engine())
	})
}

func TestResetPolicies_ProviderError(t *testing.T) {
	a := NewAuthorizer(context.Background(), bLogger.NopLogger(), &EngineConfig{Type: "casbin"}, errProvider{})
	require.NotNil(t, a.Engine())
	assert.Error(t, a.ResetPolicies(context.Background()))
}

type brokenModelProvider struct{}

func (brokenModelProvider) ProvideModels(string) ModelDataMap {
	return ModelDataMap{"opa": []byte("this is not a valid rego model !!!")}
}

func (brokenModelProvider) ProvidePolicies(context.Context) (PermissionDataMap, error) {
	return nil, nil
}

type errProvider struct{}

func (errProvider) ProvideModels(string) ModelDataMap { return nil }

func (errProvider) ProvidePolicies(context.Context) (PermissionDataMap, error) {
	return nil, errors.New("provider failure")
}
