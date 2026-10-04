package crypto

import (
	"encoding/json"
	"fmt"
)

const (
	// EncryptedConfigKey is the key used to store encrypted configuration in task payload
	EncryptedConfigKey = "_encrypted_config"
	// IsEncryptedKey indicates if the payload contains encrypted data
	IsEncryptedKey = "_is_encrypted"
)

// EncryptPayload encrypts the entire payload and returns a map with encrypted data
// This is used to store encrypted configuration in Redis/Asynq.
// reservedKeys 列出的键以明文保留在外层，供无需解密的路由/调度层读取
// （如任务系统的 "task_id"/"task_type"）。
//
// 安全契约：仅当真正完成加密时才置 IsEncryptedKey=true。若全局加密器未配置
// （EncryptIfNeeded 返回明文原值），则不标记为已加密，避免调用方误判明文为密文
// 而在解密侧产生"解密失败"或静默数据错乱。
func EncryptPayload(payload map[string]interface{}, reservedKeys ...string) (map[string]interface{}, error) {
	// Marshal the payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Encrypt the JSON
	plaintext := string(jsonData)
	encrypted, err := EncryptIfNeeded(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt payload: %w", err)
	}

	// 仅当密文与明文不同（确实加密）时标记 IsEncryptedKey，
	// 否则保持明文存储，与 DecryptPayload 的"未加密即原样返回"逻辑一致。
	result := map[string]interface{}{
		EncryptedConfigKey: encrypted,
		IsEncryptedKey:     encrypted != plaintext,
	}

	// 保留调用方声明的非敏感路由字段（如任务系统的 task_id/task_type），
	// 使调度/路由层无需解密即可读取。
	for _, key := range reservedKeys {
		if v, ok := payload[key]; ok {
			result[key] = v
		}
	}

	return result, nil
}

// DecryptPayload decrypts the payload if it contains encrypted configuration
// Returns the decrypted payload map
func DecryptPayload(payload map[string]interface{}) (map[string]interface{}, error) {
	// Check if payload is encrypted
	isEncrypted, ok := payload[IsEncryptedKey].(bool)
	if !ok || !isEncrypted {
		// Not encrypted, return as-is for backward compatibility
		return payload, nil
	}

	// Get encrypted config
	encryptedConfig, ok := payload[EncryptedConfigKey].(string)
	if !ok {
		return nil, fmt.Errorf("encrypted config not found or invalid type")
	}

	// Decrypt the config
	decrypted, err := DecryptIfNeeded(encryptedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt payload: %w", err)
	}

	// Unmarshal back to map
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(decrypted), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal decrypted payload: %w", err)
	}

	return result, nil
}

// MustEncryptPayload encrypts payload and panics on error (useful for testing)
func MustEncryptPayload(payload map[string]interface{}, reservedKeys ...string) map[string]interface{} {
	encrypted, err := EncryptPayload(payload, reservedKeys...)
	if err != nil {
		panic(err)
	}
	return encrypted
}

// MustDecryptPayload decrypts payload and panics on error (useful for testing)
func MustDecryptPayload(payload map[string]interface{}) map[string]interface{} {
	decrypted, err := DecryptPayload(payload)
	if err != nil {
		panic(err)
	}
	return decrypted
}

// HasEncryptedPayload checks if the payload contains encrypted configuration
func HasEncryptedPayload(payload map[string]interface{}) bool {
	isEncrypted, ok := payload[IsEncryptedKey].(bool)
	return ok && isEncrypted
}
