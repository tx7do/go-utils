package captcha

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/mojocn/base64Captcha"
	"github.com/redis/go-redis/v9"
	"github.com/wenlng/go-captcha/v2/base/option"
	"github.com/wenlng/go-captcha/v2/slide"
)

type Captcha struct {
	rdb          *redis.Client
	config       *Config
	slideCaptcha slide.Captcha // 滑动验证码实例（懒加载）
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

// initSlideCaptcha 初始化滑动验证码实例（懒加载）
func (c *Captcha) initSlideCaptcha() error {
	if c.slideCaptcha != nil {
		return nil
	}

	slideCfg := c.config.SlideConfig
	if slideCfg == nil {
		slideCfg = DefaultSlideConfig()
	}

	// 创建 builder
	builder := slide.NewBuilder(
		slide.WithImageSize(option.Size{Width: slideCfg.MasterWidth, Height: slideCfg.MasterHeight}),
		slide.WithRangeGraphSize(option.RangeVal{Min: slideCfg.TileWidth, Max: slideCfg.TileWidth}),
	)

	// 设置资源配置
	// 注意：这里使用默认的背景图片，实际项目中应该从资源文件加载
	bgImage, err := loadDefaultBackground()
	if err != nil {
		return fmt.Errorf("failed to load background: %w", err)
	}

	builder.SetResources(
		slide.WithBackgrounds([]image.Image{bgImage}),
	)

	c.slideCaptcha = builder.Make()
	return nil
}

// loadDefaultBackground 加载默认背景图片（创建一个简单的渐变色背景）
func loadDefaultBackground() (image.Image, error) {
	// 创建一个简单的纯色背景
	// 在实际项目中，应该从文件系统或嵌入资源中加载真实图片
	bg := image.NewRGBA(image.Rect(0, 0, 300, 220))
	for y := 0; y < 220; y++ {
		for x := 0; x < 300; x++ {
			// 渐变蓝色背景
			r := uint8(100 + x%50)
			g := uint8(150 + y%50)
			b := uint8(200)
			bg.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return bg, nil
}

// SlideCaptchaData 滑动验证码数据结构
type SlideCaptchaData struct {
	ID          string `json:"id"`           // 验证码ID
	MasterImage string `json:"master_image"` // 主图 base64
	TileImage   string `json:"tile_image"`   // 滑块图 base64
	XPosition   int    `json:"x_position"`   // 正确位置的 X 坐标
}

// Generate 生成验证码:返回 id, base64图片, 答案, err
// 对于滑动验证码，b64s 包含 JSON 格式的 SlideCaptchaData
func (c *Captcha) Generate() (id string, b64s string, answer string, err error) {
	// 如果是滑动验证码，使用特殊处理
	if c.config.DriverType == DriverSlide {
		return c.generateSlide()
	}

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

// generateSlide 生成滑动拼图验证码
func (c *Captcha) generateSlide() (id string, b64s string, answer string, err error) {
	// 初始化滑动验证码
	if err := c.initSlideCaptcha(); err != nil {
		return "", "", "", fmt.Errorf("failed to init slide captcha: %w", err)
	}

	// 生成验证码数据
	captData, err := c.slideCaptcha.Generate()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate slide captcha: %w", err)
	}

	// 获取验证数据（包含缺口位置）
	blockData := captData.GetData()
	if blockData == nil {
		return "", "", "", fmt.Errorf("slide captcha data is nil")
	}

	// 获取主图和滑块图的 base64
	masterBase64 := captData.GetMasterImage().ToBase64()
	tileBase64 := captData.GetTileImage().ToBase64()

	// 构造返回数据
	slideData := &SlideCaptchaData{
		ID:          fmt.Sprintf("%d_%d", blockData.X, blockData.Y),
		MasterImage: masterBase64,
		TileImage:   tileBase64,
		XPosition:   blockData.X, // 缺口的 X 坐标
	}

	// 将滑动验证码数据序列化为 JSON
	jsonData, err := json.Marshal(slideData)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to marshal slide data: %w", err)
	}

	// ID 用于 Redis 存储
	id = slideData.ID
	// b64s 包含完整的 JSON 数据
	b64s = string(jsonData)
	// answer 是缺口的 X 坐标（用于验证）
	answer = fmt.Sprintf("%d", blockData.X)

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
	var match bool
	if c.config.DriverType == DriverSlide {
		// 滑动验证码需要特殊处理，允许一定误差
		match = c.verifySlide(ans, userInput)
	} else {
		match = ans == userInput
	}

	if match {
		_ = c.rdb.Del(ctx, key).Err()
	}

	return match, nil
}

// verifySlide 验证滑动验证码（允许误差范围）
func (c *Captcha) verifySlide(expected, actual string) bool {
	var expectedX, actualX int
	if _, err := fmt.Sscanf(expected, "%d", &expectedX); err != nil {
		return false
	}
	if _, err := fmt.Sscanf(actual, "%d", &actualX); err != nil {
		return false
	}

	// 允许 ±5 像素的误差
	diff := expectedX - actualX
	if diff < 0 {
		diff = -diff
	}
	return diff <= 5
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
	if c.config.DriverType == DriverSlide {
		return c.verifySlide(ans, userInput), nil
	}
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
