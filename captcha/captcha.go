package captcha

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
)

type Captcha struct {
	rdb    *redis.Client
	config *Config
}

// NewCaptcha 创建验证码实例（使用 Options 模式）
// 示例:
//
//	cap := captcha.NewCaptcha(rdb,
//	    captcha.WithDriverType(captcha.DriverString),
//	    captcha.WithExpire(10*time.Minute),
//	    captcha.WithStringCount(6),
//	)
func NewCaptcha(rdb *redis.Client, opts ...Option) *Captcha {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	return &Captcha{
		rdb:    rdb,
		config: config,
	}
}

// NewCaptchaWithConfig 使用自定义配置创建实例
func NewCaptchaWithConfig(rdb *redis.Client, config *Config) *Captcha {
	if config == nil {
		config = DefaultConfig()
	}
	return &Captcha{
		rdb:    rdb,
		config: config,
	}
}

// Generate 生成验证码:返回 id, base64图片, 答案, err
func (c *Captcha) Generate() (id string, b64s string, answer string, err error) {
	var driver base64Captcha.Driver

	switch c.config.DriverType {
	case DriverDigit:
		digitCfg := c.config.DigitConfig
		if digitCfg == nil {
			digitCfg = DefaultDigitConfig()
		}
		driver = base64Captcha.NewDriverDigit(
			digitCfg.Height,
			digitCfg.Width,
			digitCfg.CaptchaCount,
			digitCfg.MaxSkew,
			digitCfg.DotCount,
		)
	case DriverString:
		stringCfg := c.config.StringConfig
		if stringCfg == nil {
			stringCfg = DefaultStringConfig()
		}
		driver = base64Captcha.NewDriverString(
			stringCfg.Height,
			stringCfg.Width,
			stringCfg.DotCount,
			0, // showLineOptions: 0=不显示干扰线
			stringCfg.CaptchaCount,
			stringCfg.Source,
			nil, // bgColor
			nil, // fontsStorage
			nil, // fonts
		)
	case DriverMath:
		mathCfg := c.config.MathConfig
		if mathCfg == nil {
			mathCfg = DefaultMathConfig()
		}
		driver = base64Captcha.NewDriverMath(
			mathCfg.Height,
			mathCfg.Width,
			mathCfg.DotCount,
			0,   // showLineOptions
			nil, // bgColor
			nil, // fontsStorage
			nil, // fonts
		)
	case DriverChinese:
		chineseCfg := c.config.ChineseConfig
		if chineseCfg == nil {
			chineseCfg = DefaultChineseConfig()
		}
		driver = base64Captcha.NewDriverChinese(
			chineseCfg.Height,
			chineseCfg.Width,
			chineseCfg.DotCount,
			0, // showLineOptions
			chineseCfg.CaptchaCount,
			chineseCfg.Language,
			nil, // bgColor
			nil, // fontsStorage
			nil, // fonts
		)
	default:
		// 默认使用数字验证码
		digitCfg := c.config.DigitConfig
		if digitCfg == nil {
			digitCfg = DefaultDigitConfig()
		}
		driver = base64Captcha.NewDriverDigit(
			digitCfg.Height,
			digitCfg.Width,
			digitCfg.CaptchaCount,
			digitCfg.MaxSkew,
			digitCfg.DotCount,
		)
	}

	id, question, answer := driver.GenerateIdQuestionAnswer()

	// 绘制验证码图片
	item, err := driver.DrawCaptcha(question)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to draw captcha: %w", err)
	}

	// 转换为 base64
	b64s = item.EncodeB64string()

	return id, b64s, answer, nil
}

// Save 将验证码答案存入 Redis
func (c *Captcha) Save(ctx context.Context, captchaID, answer string) error {
	key := fmt.Sprintf("%s:%s", c.config.KeyPrefix, captchaID)
	return c.rdb.Set(ctx, key, answer, c.config.Expire).Err()
}

// Verify 从Redis读取并校验验证码
// 验证成功自动删除
func (c *Captcha) Verify(ctx context.Context, captchaID, userInput string) (bool, error) {
	key := fmt.Sprintf("%s:%s", c.config.KeyPrefix, captchaID)

	// 获取答案
	ans, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil // 过期/不存在
		}
		return false, err
	}

	// 校验成功就删除
	match := ans == userInput
	if match {
		_ = c.rdb.Del(ctx, key).Err()
	}

	return match, nil
}

// GetConfig 获取当前配置
func (c *Captcha) GetConfig() *Config {
	return c.config
}

// SetConfig 设置配置
func (c *Captcha) SetConfig(config *Config) {
	if config != nil {
		c.config = config
	}
}

// VerifyAndDelete 验证并删除验证码（与Verify相同，但名称更明确）
func (c *Captcha) VerifyAndDelete(ctx context.Context, captchaID, userInput string) (bool, error) {
	return c.Verify(ctx, captchaID, userInput)
}

// VerifyWithoutDelete 验证但不删除验证码（用于多次验证场景）
func (c *Captcha) VerifyWithoutDelete(ctx context.Context, captchaID, userInput string) (bool, error) {
	key := fmt.Sprintf("%s:%s", c.config.KeyPrefix, captchaID)

	// 获取答案
	ans, err := c.rdb.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil // 过期/不存在
		}
		return false, err
	}

	// 只校验，不删除
	return ans == userInput, nil
}

// Delete 手动删除验证码
func (c *Captcha) Delete(ctx context.Context, captchaID string) error {
	key := fmt.Sprintf("%s:%s", c.config.KeyPrefix, captchaID)
	return c.rdb.Del(ctx, key).Err()
}

// Exists 检查验证码是否存在
func (c *Captcha) Exists(ctx context.Context, captchaID string) (bool, error) {
	key := fmt.Sprintf("%s:%s", c.config.KeyPrefix, captchaID)
	result, err := c.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// GetRemainingTime 获取验证码剩余时间
func (c *Captcha) GetRemainingTime(ctx context.Context, captchaID string) (time.Duration, error) {
	key := fmt.Sprintf("%s:%s", c.config.KeyPrefix, captchaID)
	return c.rdb.TTL(ctx, key).Result()
}
