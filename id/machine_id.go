package id

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/tx7do/go-utils/id/machineid"
)

// FormatOption 定义格式化选项
type FormatOption struct {
	UpperCase  bool // true: 大写，false: 小写
	WithHyphen bool // true: 带横线，false: 不带横线
}

// FormatMachineID 获取格式化后的 machineId
func FormatMachineID(opt FormatOption) (string, error) {
	id, err := getRawMachineID()
	if err != nil {
		return "", err
	}
	// 只保留16进制字符
	cleaned := make([]byte, 0, 32)
	for i := 0; i < len(id) && len(cleaned) < 32; i++ {
		c := id[i]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
			if !opt.UpperCase && c >= 'A' && c <= 'F' {
				c += 32 // 转小写
			} else if opt.UpperCase && c >= 'a' && c <= 'f' {
				c -= 32 // 转大写
			}
			cleaned = append(cleaned, c)
		}
	}
	result := string(cleaned)
	if len(result) != 32 {
		hash := sha256.Sum256([]byte(id))
		result = hex.EncodeToString(hash[:])[:32]
		if opt.UpperCase {
			// 转大写
			b := []byte(result)
			for i := range b {
				if b[i] >= 'a' && b[i] <= 'f' {
					b[i] -= 32
				}
			}
			result = string(b)
		}
	}
	if opt.WithHyphen {
		// 插入标准GUID横线 8-4-4-4-12
		if len(result) == 32 {
			result = result[:8] + "-" + result[8:12] + "-" + result[12:16] + "-" + result[16:20] + "-" + result[20:]
		}
	}
	return result, nil
}

// getRawMachineID 获取原始 machineid
func getRawMachineID() (string, error) {
	var err error
	cacheOnce.Do(func() {
		machineIDCache, err = unifyMachineIDInternal(machineid.ID)
	})
	return machineIDCache, err
}

var (
	machineIDCache string
	cacheOnce      sync.Once
)

// UnifyMachineID 兼容旧接口，等价于 FormatMachineID(小写无横线)
func UnifyMachineID() (string, error) {
	return FormatMachineID(FormatOption{UpperCase: false, WithHyphen: false})
}

// unifyMachineIDInternal 内部实现，便于单元测试（支持依赖注入）
func unifyMachineIDInternal(idFetcher func() (string, error)) (string, error) {
	raw, err := idFetcher()
	if err != nil {
		return "", fmt.Errorf("获取machineId失败: %w", err)
	}

	// 方案A: 直接清洗（适用于标准GUID格式）
	cleaned := make([]byte, 0, 32)
	for i := 0; i < len(raw) && len(cleaned) < 32; i++ {
		c := raw[i]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
			if c >= 'A' && c <= 'F' {
				c += 32 // 转小写，比 strings.ToLower 更高效
			}
			cleaned = append(cleaned, c)
		}
	}

	result := string(cleaned)
	if len(result) != 32 {
		// 方案B: 降级策略 - 如果长度不符，用SHA256哈希确保32位输出
		// 避免直接失败，提高兼容性
		hash := sha256.Sum256([]byte(raw))
		result = hex.EncodeToString(hash[:])[:32]
	}
	return result, nil
}

// formatMachineIDWithFetcher 便于测试注入 mockFetcher
func formatMachineIDWithFetcher(opt FormatOption, fetcher func() (string, error)) (string, error) {
	id, err := fetcher()
	if err != nil {
		return "", err
	}
	cleaned := make([]byte, 0, 32)
	for i := 0; i < len(id) && len(cleaned) < 32; i++ {
		c := id[i]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') {
			if !opt.UpperCase && c >= 'A' && c <= 'F' {
				c += 32 // 转小写
			} else if opt.UpperCase && c >= 'a' && c <= 'f' {
				c -= 32 // 转大写
			}
			cleaned = append(cleaned, c)
		}
	}
	result := string(cleaned)
	if len(result) != 32 {
		hash := sha256.Sum256([]byte(id))
		result = hex.EncodeToString(hash[:])[:32]
		if opt.UpperCase {
			b := []byte(result)
			for i := range b {
				if b[i] >= 'a' && b[i] <= 'f' {
					b[i] -= 32
				}
			}
			result = string(b)
		}
	}
	if opt.WithHyphen {
		if len(result) == 32 {
			result = result[:8] + "-" + result[8:12] + "-" + result[12:16] + "-" + result[16:20] + "-" + result[20:]
		}
	}
	return result, nil
}
