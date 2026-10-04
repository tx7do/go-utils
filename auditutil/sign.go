package auditutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// 审计日志防篡改约定：日志体在计算哈希时排除的字段名（proto 字段名与
// json_name 两种拼写都命中）。哈希随日志落库，签名覆盖哈希，形成
// "签名(hash(body))" 的两级防篡改链。
var defaultExcludedFields = []string{"log_hash", "logHash", "signature"}

// GenerateECDSAKeyPair 生成 ECDSA 密钥对（secp256r1 曲线），用于审计日志签名。
func GenerateECDSAKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate ECDSA key failed: %w", err)
	}
	return privateKey, &privateKey.PublicKey, nil
}

// EncodeECDSADER 将 ECDSA 的 r、s 转为 DER 格式字节数组（标准签名字段，
// 可被 ecdsa.VerifyASN1 校验）。
func EncodeECDSADER(r, s *big.Int) ([]byte, error) {
	// DER 格式规则：0x30 + 总长度 + 0x02 + r长度 + r值 + 0x02 + s长度 + s值
	rBytes := r.Bytes()
	sBytes := s.Bytes()

	// 确保 r/s 是正整数（补 0 前缀）
	if len(rBytes) > 0 && rBytes[0]&0x80 != 0 {
		rBytes = append([]byte{0x00}, rBytes...)
	}
	if len(sBytes) > 0 && sBytes[0]&0x80 != 0 {
		sBytes = append([]byte{0x00}, sBytes...)
	}

	// 拼接 DER 字节
	der := make([]byte, 0)
	der = append(der, 0x30)                              // 序列标签
	der = append(der, byte(2+len(rBytes)+2+len(sBytes))) // 总长度
	der = append(der, 0x02)                              // 整数标签（r）
	der = append(der, byte(len(rBytes)))                 // r 长度
	der = append(der, rBytes...)
	der = append(der, 0x02)              // 整数标签（s）
	der = append(der, byte(len(sBytes))) // s 长度
	der = append(der, sBytes...)

	return der, nil
}

// HashLog 计算日志消息的 SHA256 哈希（十六进制小写字符串）。
// 规则：先在消息副本上清除 excludeFields 命中的字段（缺省为 log_hash 与
// signature），再对 Protobuf 确定性序列化结果哈希——哈希值不依赖日志自身
// 的哈希/签名字段，防止自引用。消息本身不被修改。
func HashLog(msg proto.Message, excludeFields ...string) string {
	if msg == nil || !msg.ProtoReflect().IsValid() {
		return ""
	}
	if len(excludeFields) == 0 {
		excludeFields = defaultExcludedFields
	}

	clone := proto.Clone(msg).ProtoReflect()
	fields := clone.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		name := string(fd.Name())
		jsonName := fd.JSONName()
		for _, want := range excludeFields {
			if name == want || jsonName == want {
				clone.Clear(fd)
				break
			}
		}
	}

	rawBytes, err := proto.Marshal(clone.Interface())
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(rawBytes)
	return hex.EncodeToString(hash[:])
}

// SignLogContent 生成审计日志的 ECDSA 数字签名（DER 格式）。
// 签名内容：tenant_id + user_id + created_at（原始时间戳）+ log_hash 的
// JSON 序列化的 SHA256。priv 为 nil 时返回错误——签名即凭证，调用方不应
// 降级为无签名落库。
func SignLogContent(priv *ecdsa.PrivateKey, tenantID, userID uint32, createdAt *timestamppb.Timestamp, logHash string) ([]byte, error) {
	if priv == nil {
		return nil, fmt.Errorf("ecdsa private key is nil: signing unavailable")
	}

	type signContent struct {
		TenantID uint32 `json:"tenant_id"`
		UserID   uint32 `json:"user_id"`
		Sec      int64  `json:"sec"`   // createdAt 秒数
		Nanos    int32  `json:"nanos"` // createdAt 纳秒数
		LogHash  string `json:"log_hash"`
	}
	sc := signContent{
		TenantID: tenantID,
		UserID:   userID,
		LogHash:  logHash,
	}
	if createdAt != nil {
		sc.Sec = createdAt.Seconds
		sc.Nanos = createdAt.Nanos
	}

	scBytes, err := json.Marshal(sc)
	if err != nil {
		return nil, fmt.Errorf("marshal sign content failed: %w", err)
	}

	scHash := sha256.Sum256(scBytes)

	r, s, err := ecdsa.Sign(rand.Reader, priv, scHash[:])
	if err != nil {
		return nil, fmt.Errorf("ECDSA sign failed: %w", err)
	}

	signBytes, err := EncodeECDSADER(r, s)
	if err != nil {
		return nil, fmt.Errorf("encode DER failed: %w", err)
	}

	return signBytes, nil
}
