package crypto

// 本文件测试全局加密器的"未初始化"语义与 Must* 辅助函数。
//
// 执行顺序依赖（重要）：
// 默认测试顺序（按文件名字典序、文件内按声明序）下，本文件先于
// hmac_test.go / payload_test.go 执行，且本文件自身不调用
// InitGlobalEncryptor，因此本文件中排在最前的"未初始化"用例执行时
// 全局加密器尚未被任何用例初始化（sync.Once 尚未触发）。
// 若顺序被 -shuffle 打乱或日后新增更早触发 InitGlobalEncryptor 的测试，
// 各未初始化态用例会通过 requireGlobalEncryptorUnitinitialized 探测到
// 已初始化并跳过（t.Skip），保证任何顺序下测试套件稳定。
//
// 未初始化语义是部署安全契约的一部分：GOWIND_CRYPTO_KEY 未配置时，
// 加解密能力整体降级为透明直通，且签名能力必须显式报错（见 hmac_test.go），
// 而不是静默产出可伪造的签名。

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// requireGlobalEncryptorUnitinitialized 探测全局加密器是否已被本进程内
// 更早的用例初始化：已初始化则跳过未初始化态断言（-shuffle 或用例
// 顺序变化时保持套件稳定），未初始化则继续执行断言。
func requireGlobalEncryptorUnitinitialized(t *testing.T) {
	t.Helper()
	probe, err := EncryptIfNeeded("probe")
	require.NoError(t, err)
	if IsEncrypted(probe) {
		t.Skip("全局加密器已被更早执行的用例初始化（顺序被打乱），跳过未初始化态断言")
	}
}

// TestGetGlobalEncryptor_BeforeInitReturnsNoOpInstance 全局加密器未初始化时
// 必须返回空 key 的直通实例，绝不返回 nil——调用方（如 EncryptIfNeeded）
// 依赖该实例判定"加密能力未启用"。
func TestGetGlobalEncryptor_BeforeInitReturnsNoOpInstance(t *testing.T) {
	requireGlobalEncryptorUnitinitialized(t)
	enc := GetGlobalEncryptor()
	require.NotNil(t, enc, "未初始化时也必须返回直通实例而非 nil")
	assert.Empty(t, enc.key, "直通实例的 key 必须为空（加密能力未启用）")
}

// TestEncryptIfNeeded_BeforeInitPassthrough 未初始化时加密请求必须
// 原样返回明文：此路径是任务配置等数据的写入路径，若抛错会让未启用
// 加密的部署整体不可用。
func TestEncryptIfNeeded_BeforeInitPassthrough(t *testing.T) {
	requireGlobalEncryptorUnitinitialized(t)
	out, err := EncryptIfNeeded("sensitive-plaintext")
	require.NoError(t, err)
	assert.Equal(t, "sensitive-plaintext", out,
		"未初始化时必须明文直通")
	assert.False(t, IsEncrypted(out), "直通结果不得带 enc: 前缀")
}

// TestDecryptIfNeeded_BeforeInitPassthrough 未初始化时解密请求原样返回，
// 包括带 enc: 前缀的输入（无 key 无法解密，直通是唯一无数据损坏的选择）。
func TestDecryptIfNeeded_BeforeInitPassthrough(t *testing.T) {
	requireGlobalEncryptorUnitinitialized(t)
	out, err := DecryptIfNeeded("plain-legacy-data")
	require.NoError(t, err)
	assert.Equal(t, "plain-legacy-data", out)

	out, err = DecryptIfNeeded("enc:U29tZU9sZERhdGE=")
	require.NoError(t, err)
	assert.Equal(t, "enc:U29tZU9sZERhdGE=", out,
		"未初始化时带前缀输入也必须直通")
}

// TestMustEncryptAndMustDecrypt_RoundTrip Must* 辅助函数在有效实例上的
// 成功路径：与 Encrypt/Decrypt 等价，密文往返无损（供内部测试与引导期
// 确定性配置使用）。
func TestMustEncryptAndMustDecrypt_RoundTrip(t *testing.T) {
	enc, err := NewEncryptor("must-roundtrip-key")
	require.NoError(t, err)

	plaintext := `{"host":"imap.example.com","password":"secret"}`
	ciphertext := enc.MustEncrypt(plaintext)
	assert.True(t, IsEncrypted(ciphertext), "MustEncrypt 产物必须带 enc: 前缀")
	assert.NotEqual(t, plaintext, ciphertext)
	assert.Equal(t, plaintext, enc.MustDecrypt(ciphertext), "往返必须无损")
}

// TestMustEncrypt_PanicsWhenCipherUnavailable 空 key 实例（等价未启用
// 加密却误走 Must 路径）上加密非空串必然失败，MustEncrypt 必须把错误
// 以 panic 暴露，而不是吞掉返回空串。
func TestMustEncrypt_PanicsWhenCipherUnavailable(t *testing.T) {
	noop := &Encryptor{}
	assert.Panics(t, func() {
		_ = noop.MustEncrypt("anything")
	}, "空 key 实例上 MustEncrypt 必须 panic")
}

// TestMustDecrypt_PanicsWhenCipherUnavailable 空 key 实例上对带前缀
// 密文调用 MustDecrypt：cipher 建立失败必须 panic 而非静默返回。
func TestMustDecrypt_PanicsWhenCipherUnavailable(t *testing.T) {
	noop := &Encryptor{}
	assert.Panics(t, func() {
		_ = noop.MustDecrypt("enc:YWJj") // 3 字节，短于 nonce，必失败
	}, "空 key 实例上 MustDecrypt 必须 panic")
}
