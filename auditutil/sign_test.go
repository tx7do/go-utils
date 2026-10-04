package auditutil

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testpb "github.com/tx7do/go-utils/auditutil/testpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func newTestLog() *testpb.TestLog {
	return &testpb.TestLog{
		TenantId:  7,
		UserId:    42,
		Action:    "login",
		CreatedAt: timestamppb.New(time.Unix(1735689600, 0)), // 2025-01-01 00:00:00 UTC
	}
}

func TestHashLog_ExcludesHashAndSignatureFields(t *testing.T) {
	log := newTestLog()
	base := HashLog(log)

	// 填上 log_hash / signature 后哈希不变——这两字段被排除
	log.LogHash = strPtrOf("deadbeef")
	log.Signature = []byte{1, 2, 3}
	assert.Equal(t, base, HashLog(log), "log_hash/signature 不得影响哈希值")

	// 其他字段变化必须改变哈希
	log.Action = "logout"
	assert.NotEqual(t, base, HashLog(log))
}

func TestHashLog_NilAndInvalid(t *testing.T) {
	assert.Equal(t, "", HashLog(nil))
	var typedNil *testpb.TestLog
	assert.Equal(t, "", HashLog(typedNil))
}

func TestHashLog_OriginalMessageUnmodified(t *testing.T) {
	log := newTestLog()
	log.LogHash = strPtrOf("keepme")
	_ = HashLog(log)
	assert.Equal(t, "keepme", log.GetLogHash(), "哈希不得修改原消息")
}

func TestHashLog_CustomExcludeFields(t *testing.T) {
	log := newTestLog()
	base := HashLog(log)

	assert.Equal(t, base, HashLog(log, "log_hash", "logHash", "signature"), "显式默认排除集应等价")
	assert.NotEqual(t, base, HashLog(log, "tenant_id"), "排除 tenant_id 后哈希必须不同")
}

func TestSignLogContent(t *testing.T) {
	priv, pub, err := GenerateECDSAKeyPair()
	require.NoError(t, err)
	require.NotNil(t, pub)

	logHash := HashLog(newTestLog())

	sig, err := SignLogContent(priv, 7, 42, newTestLog().CreatedAt, logHash)
	require.NoError(t, err)
	assert.NotEmpty(t, sig)
	assert.Equal(t, 0x30, int(sig[0]), "签名应为 DER 序列（0x30 开头）")

	// 签名必须可通过 DER 校验（与签名内容逐字节一致的重算路径）
	scJSON := signContentJSON(7, 42, 1735689600, 0, logHash)
	ok := verifyECDSA(pub, scJSON, sig)
	assert.True(t, ok, "签名必须能被标准 ecdsa 验签")

	// 内容篡改后验签必须失败
	ok = verifyECDSA(pub, signContentJSON(7, 42, 1735689600, 0, "tampered"), sig)
	assert.False(t, ok, "篡改 log_hash 后验签必须失败")
}

func TestSignLogContent_NilKeyErrors(t *testing.T) {
	_, err := SignLogContent(nil, 1, 1, nil, "hash")
	assert.Error(t, err, "私钥缺失必须报错而非返回空签名")
}

// ---- 测试辅助：与 SignLogContent 相同的签名内容序列化 + 标准库验签 ----

type signContent struct {
	TenantID uint32 `json:"tenant_id"`
	UserID   uint32 `json:"user_id"`
	Sec      int64  `json:"sec"`
	Nanos    int32  `json:"nanos"`
	LogHash  string `json:"log_hash"`
}

func signContentJSON(tenantID, userID uint32, sec int64, nanos int32, logHash string) []byte {
	b, _ := json.Marshal(signContent{TenantID: tenantID, UserID: userID, Sec: sec, Nanos: nanos, LogHash: logHash})
	return b
}

func verifyECDSA(pub *ecdsa.PublicKey, content, derSig []byte) bool {
	h := sha256.Sum256(content)
	return ecdsa.VerifyASN1(pub, h[:], derSig)
}

func strPtrOf(s string) *string { return &s }
