package crypto

import (
	"github.com/tjfoc/gmsm/sm3"
)

// SM3Hasher 提供SM3哈希算法的简单封装
// 用于数据摘要、签名等场景

type SM3Hasher struct{}

// NewSM3Hasher 构造SM3哈希器
func NewSM3Hasher() *SM3Hasher {
	return &SM3Hasher{}
}

// Sum 计算数据的SM3哈希值
func (h *SM3Hasher) Sum(data []byte) []byte {
	hasher := sm3.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}

// Name 返回哈希算法的名称
func (h *SM3Hasher) Name() string {
	return "SM3"
}
