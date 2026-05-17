package captcha

import (
	"time"
)

// DriverType 验证码驱动类型
type DriverType string

const (
	DriverDigit   DriverType = "digit"   // 数字验证码
	DriverString  DriverType = "string"  // 字符串验证码
	DriverMath    DriverType = "math"    // 算术验证码
	DriverChinese DriverType = "chinese" // 中文验证码
)

// DigitConfig 数字验证码配置
type DigitConfig struct {
	Height       int     `json:"height"`        // 图片高度
	Width        int     `json:"width"`         // 图片宽度
	CaptchaCount int     `json:"captcha_count"` // 验证码字符数量
	MaxSkew      float64 `json:"max_skew"`      // 最大倾斜度
	DotCount     int     `json:"dot_count"`     // 干扰点数量
	BgColorR     uint8   `json:"bg_color_r"`    // 背景色R
	BgColorG     uint8   `json:"bg_color_g"`    // 背景色G
	BgColorB     uint8   `json:"bg_color_b"`    // 背景色B
	FontColorR   uint8   `json:"font_color_r"`  // 字体色R
	FontColorG   uint8   `json:"font_color_g"`  // 字体色G
	FontColorB   uint8   `json:"font_color_b"`  // 字体色B
	CaptchaLen   int     `json:"captcha_len"`   // 验证码长度（兼容字段）
}

// StringConfig 字符串验证码配置
type StringConfig struct {
	Height       int     `json:"height"`        // 图片高度
	Width        int     `json:"width"`         // 图片宽度
	CaptchaCount int     `json:"captcha_count"` // 验证码字符数量
	MaxSkew      float64 `json:"max_skew"`      // 最大倾斜度
	DotCount     int     `json:"dot_count"`     // 干扰点数量
	BgColorR     uint8   `json:"bg_color_r"`    // 背景色R
	BgColorG     uint8   `json:"bg_color_g"`    // 背景色G
	BgColorB     uint8   `json:"bg_color_b"`    // 背景色B
	FontColorR   uint8   `json:"font_color_r"`  // 字体色R
	FontColorG   uint8   `json:"font_color_g"`  // 字体色G
	FontColorB   uint8   `json:"font_color_b"`  // 字体色B
	Source       string  `json:"source"`        // 字符源
	CaptchaLen   int     `json:"captcha_len"`   // 验证码长度（兼容字段）
}

// MathConfig 算术验证码配置
type MathConfig struct {
	Height       int     `json:"height"`        // 图片高度
	Width        int     `json:"width"`         // 图片宽度
	CaptchaCount int     `json:"captcha_count"` // 验证码字符数量
	MaxSkew      float64 `json:"max_skew"`      // 最大倾斜度
	DotCount     int     `json:"dot_count"`     // 干扰点数量
	BgColorR     uint8   `json:"bg_color_r"`    // 背景色R
	BgColorG     uint8   `json:"bg_color_g"`    // 背景色G
	BgColorB     uint8   `json:"bg_color_b"`    // 背景色B
	FontColorR   uint8   `json:"font_color_r"`  // 字体色R
	FontColorG   uint8   `json:"font_color_g"`  // 字体色G
	FontColorB   uint8   `json:"font_color_b"`  // 字体色B
	CaptchaLen   int     `json:"captcha_len"`   // 验证码长度（兼容字段）
}

// ChineseConfig 中文验证码配置
type ChineseConfig struct {
	Height       int     `json:"height"`        // 图片高度
	Width        int     `json:"width"`         // 图片宽度
	CaptchaCount int     `json:"captcha_count"` // 验证码字符数量
	MaxSkew      float64 `json:"max_skew"`      // 最大倾斜度
	DotCount     int     `json:"dot_count"`     // 干扰点数量
	BgColorR     uint8   `json:"bg_color_r"`    // 背景色R
	BgColorG     uint8   `json:"bg_color_g"`    // 背景色G
	BgColorB     uint8   `json:"bg_color_b"`    // 背景色B
	FontColorR   uint8   `json:"font_color_r"`  // 字体色R
	FontColorG   uint8   `json:"font_color_g"`  // 字体色G
	FontColorB   uint8   `json:"font_color_b"`  // 字体色B
	Language     string  `json:"language"`      // 语言类型 zh, en
	CaptchaLen   int     `json:"captcha_len"`   // 验证码长度（兼容字段）
}

// Config 验证码总配置
type Config struct {
	DriverType    DriverType     `json:"driver_type"`    // 驱动类型
	Expire        time.Duration  `json:"expire"`         // 过期时间
	KeyPrefix     string         `json:"key_prefix"`     // Redis key前缀
	DigitConfig   *DigitConfig   `json:"digit_config"`   // 数字配置
	StringConfig  *StringConfig  `json:"string_config"`  // 字符串配置
	MathConfig    *MathConfig    `json:"math_config"`    // 算术配置
	ChineseConfig *ChineseConfig `json:"chinese_config"` // 中文配置
}

// DefaultDigitConfig 默认数字配置
func DefaultDigitConfig() *DigitConfig {
	return &DigitConfig{
		Height:       80,
		Width:        240,
		CaptchaCount: 4,
		MaxSkew:      0.7,
		DotCount:     80,
		BgColorR:     255,
		BgColorG:     255,
		BgColorB:     255,
		FontColorR:   0,
		FontColorG:   0,
		FontColorB:   0,
		CaptchaLen:   4,
	}
}

// DefaultStringConfig 默认字符串配置
func DefaultStringConfig() *StringConfig {
	return &StringConfig{
		Height:       80,
		Width:        240,
		CaptchaCount: 4,
		MaxSkew:      0.7,
		DotCount:     80,
		BgColorR:     255,
		BgColorG:     255,
		BgColorB:     255,
		FontColorR:   0,
		FontColorG:   0,
		FontColorB:   0,
		Source:       "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		CaptchaLen:   4,
	}
}

// DefaultMathConfig 默认算术配置
func DefaultMathConfig() *MathConfig {
	return &MathConfig{
		Height:       80,
		Width:        240,
		CaptchaCount: 4,
		MaxSkew:      0.7,
		DotCount:     80,
		BgColorR:     255,
		BgColorG:     255,
		BgColorB:     255,
		FontColorR:   0,
		FontColorG:   0,
		FontColorB:   0,
		CaptchaLen:   4,
	}
}

// DefaultChineseConfig 默认中文配置
func DefaultChineseConfig() *ChineseConfig {
	return &ChineseConfig{
		Height:       80,
		Width:        240,
		CaptchaCount: 4,
		MaxSkew:      0.7,
		DotCount:     80,
		BgColorR:     255,
		BgColorG:     255,
		BgColorB:     255,
		FontColorR:   0,
		FontColorG:   0,
		FontColorB:   0,
		Language:     "zh",
		CaptchaLen:   4,
	}
}

// DefaultConfig 默认总配置
func DefaultConfig() *Config {
	return &Config{
		DriverType:    DriverDigit,
		Expire:        5 * time.Minute,
		KeyPrefix:     "captcha",
		DigitConfig:   DefaultDigitConfig(),
		StringConfig:  DefaultStringConfig(),
		MathConfig:    DefaultMathConfig(),
		ChineseConfig: DefaultChineseConfig(),
	}
}

// Option 配置选项函数类型
type Option func(*Config)

// WithDriverType 设置驱动类型
func WithDriverType(driverType DriverType) Option {
	return func(c *Config) {
		c.DriverType = driverType
	}
}

// WithExpire 设置过期时间
func WithExpire(expire time.Duration) Option {
	return func(c *Config) {
		c.Expire = expire
	}
}

// WithKeyPrefix 设置 Redis key 前缀
func WithKeyPrefix(prefix string) Option {
	return func(c *Config) {
		c.KeyPrefix = prefix
	}
}

// WithDigitConfig 设置数字验证码配置
func WithDigitConfig(config *DigitConfig) Option {
	return func(c *Config) {
		c.DigitConfig = config
	}
}

// WithStringConfig 设置字符串验证码配置
func WithStringConfig(config *StringConfig) Option {
	return func(c *Config) {
		c.StringConfig = config
	}
}

// WithMathConfig 设置算术验证码配置
func WithMathConfig(config *MathConfig) Option {
	return func(c *Config) {
		c.MathConfig = config
	}
}

// WithChineseConfig 设置中文验证码配置
func WithChineseConfig(config *ChineseConfig) Option {
	return func(c *Config) {
		c.ChineseConfig = config
	}
}

// WithDigitHeight 设置数字验证码高度
func WithDigitHeight(height int) Option {
	return func(c *Config) {
		if c.DigitConfig == nil {
			c.DigitConfig = DefaultDigitConfig()
		}
		c.DigitConfig.Height = height
	}
}

// WithDigitWidth 设置数字验证码宽度
func WithDigitWidth(width int) Option {
	return func(c *Config) {
		if c.DigitConfig == nil {
			c.DigitConfig = DefaultDigitConfig()
		}
		c.DigitConfig.Width = width
	}
}

// WithDigitCount 设置数字验证码字符数量
func WithDigitCount(count int) Option {
	return func(c *Config) {
		if c.DigitConfig == nil {
			c.DigitConfig = DefaultDigitConfig()
		}
		c.DigitConfig.CaptchaCount = count
	}
}

// WithDigitMaxSkew 设置数字验证码最大倾斜度
func WithDigitMaxSkew(skew float64) Option {
	return func(c *Config) {
		if c.DigitConfig == nil {
			c.DigitConfig = DefaultDigitConfig()
		}
		c.DigitConfig.MaxSkew = skew
	}
}

// WithDigitDotCount 设置数字验证码干扰点数量
func WithDigitDotCount(count int) Option {
	return func(c *Config) {
		if c.DigitConfig == nil {
			c.DigitConfig = DefaultDigitConfig()
		}
		c.DigitConfig.DotCount = count
	}
}

// WithStringHeight 设置字符串验证码高度
func WithStringHeight(height int) Option {
	return func(c *Config) {
		if c.StringConfig == nil {
			c.StringConfig = DefaultStringConfig()
		}
		c.StringConfig.Height = height
	}
}

// WithStringWidth 设置字符串验证码宽度
func WithStringWidth(width int) Option {
	return func(c *Config) {
		if c.StringConfig == nil {
			c.StringConfig = DefaultStringConfig()
		}
		c.StringConfig.Width = width
	}
}

// WithStringCount 设置字符串验证码字符数量
func WithStringCount(count int) Option {
	return func(c *Config) {
		if c.StringConfig == nil {
			c.StringConfig = DefaultStringConfig()
		}
		c.StringConfig.CaptchaCount = count
	}
}

// WithStringSource 设置字符串验证码字符源
func WithStringSource(source string) Option {
	return func(c *Config) {
		if c.StringConfig == nil {
			c.StringConfig = DefaultStringConfig()
		}
		c.StringConfig.Source = source
	}
}

// WithStringDotCount 设置字符串验证码干扰点数量
func WithStringDotCount(count int) Option {
	return func(c *Config) {
		if c.StringConfig == nil {
			c.StringConfig = DefaultStringConfig()
		}
		c.StringConfig.DotCount = count
	}
}

// WithMathHeight 设置算术验证码高度
func WithMathHeight(height int) Option {
	return func(c *Config) {
		if c.MathConfig == nil {
			c.MathConfig = DefaultMathConfig()
		}
		c.MathConfig.Height = height
	}
}

// WithMathWidth 设置算术验证码宽度
func WithMathWidth(width int) Option {
	return func(c *Config) {
		if c.MathConfig == nil {
			c.MathConfig = DefaultMathConfig()
		}
		c.MathConfig.Width = width
	}
}

// WithMathDotCount 设置算术验证码干扰点数量
func WithMathDotCount(count int) Option {
	return func(c *Config) {
		if c.MathConfig == nil {
			c.MathConfig = DefaultMathConfig()
		}
		c.MathConfig.DotCount = count
	}
}

// WithChineseHeight 设置中文验证码高度
func WithChineseHeight(height int) Option {
	return func(c *Config) {
		if c.ChineseConfig == nil {
			c.ChineseConfig = DefaultChineseConfig()
		}
		c.ChineseConfig.Height = height
	}
}

// WithChineseWidth 设置中文验证码宽度
func WithChineseWidth(width int) Option {
	return func(c *Config) {
		if c.ChineseConfig == nil {
			c.ChineseConfig = DefaultChineseConfig()
		}
		c.ChineseConfig.Width = width
	}
}

// WithChineseCount 设置中文验证码字符数量
func WithChineseCount(count int) Option {
	return func(c *Config) {
		if c.ChineseConfig == nil {
			c.ChineseConfig = DefaultChineseConfig()
		}
		c.ChineseConfig.CaptchaCount = count
	}
}

// WithChineseLanguage 设置中文验证码语言
func WithChineseLanguage(language string) Option {
	return func(c *Config) {
		if c.ChineseConfig == nil {
			c.ChineseConfig = DefaultChineseConfig()
		}
		c.ChineseConfig.Language = language
	}
}

// WithChineseDotCount 设置中文验证码干扰点数量
func WithChineseDotCount(count int) Option {
	return func(c *Config) {
		if c.ChineseConfig == nil {
			c.ChineseConfig = DefaultChineseConfig()
		}
		c.ChineseConfig.DotCount = count
	}
}
