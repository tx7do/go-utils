package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// SignData 用全局加密器密钥对 data 计算 HMAC-SHA256 签名（hex 编码）。
// 用于签名 URL 场景（签名即凭证，需配合有效期由调用方自行拼接校验）。
// 全局加密器未初始化时返回错误——此时签名能力不可用，调用方应拒绝
// 生成签名 URL 而非降级为无签名。
func SignData(data string) (string, error) {
	enc := GetGlobalEncryptor()
	if enc == nil || len(enc.key) == 0 {
		return "", fmt.Errorf("global encryptor is not initialized: signing unavailable")
	}
	mac := hmac.New(sha256.New, enc.key)
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// VerifyData 校验 SignData 生成的签名（恒定时间比较）。
func VerifyData(data, signature string) bool {
	expected, err := SignData(data)
	if err != nil {
		return false
	}
	return hmac.Equal([]byte(expected), []byte(signature))
}
