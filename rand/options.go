package rand

const (
	DefaultSeedType SeedType = CryptoRandSeed
	DefaultRandType RandType = PCGRandType // 默认使用 PCG
)

// Option 定义配置函数类型
type Option func(*config)

// config 内部配置结构体，统一管理所有参数
type config struct {
	randType RandType

	seedType SeedType
	seed     uint64 // 固定种子（优先级最高）

	// Zipf 分布参数
	zipfS, zipfQ float64
	zipfV        uint64

	secureMode bool // 安全模式开关
}

// 默认配置
func defaultConfig() config {
	return config{
		randType:   DefaultRandType,
		seedType:   DefaultSeedType,
		seed:       0,
		zipfS:      1.1,
		zipfQ:      1.0,
		zipfV:      100,
		secureMode: false, // 默认关闭
	}
}

// WithRandType 设置随机数生成器类型
func WithRandType(t RandType) Option {
	return func(c *config) {
		c.randType = t
	}
}

// WithSeedType 设置种子类型
func WithSeedType(t SeedType) Option {
	return func(c *config) {
		c.seedType = t
	}
}

// WithFixedSeed 使用固定种子（可复现随机序列）
func WithFixedSeed(seed uint64) Option {
	return func(c *config) {
		c.seed = seed
	}
}

// WithZipfParams 设置 Zipf 分布参数
func WithZipfParams(s, q float64, v uint64) Option {
	return func(c *config) {
		c.zipfS = s
		c.zipfQ = q
		c.zipfV = v
	}
}

func WithSecureMode() Option {
	return func(c *config) {
		c.secureMode = true
	}
}
