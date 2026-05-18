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
	"github.com/wenlng/go-captcha/v2/click"
	"github.com/wenlng/go-captcha/v2/rotate"
	"github.com/wenlng/go-captcha/v2/slide"
)

type Captcha struct {
	rdb           *redis.Client
	config        *Config
	slideCaptcha  slide.Captcha  // 滑动验证码实例（懒加载）
	clickCaptcha  click.Captcha  // 点击验证码实例（懒加载）
	rotateCaptcha rotate.Captcha // 旋转验证码实例（懒加载）
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

// ClickCaptchaData 点击验证码数据结构
type ClickCaptchaData struct {
	ID          string       `json:"id"`           // 验证码ID
	MasterImage string       `json:"master_image"` // 主图 base64
	ThumbImage  string       `json:"thumb_image"`  // 缩略图 base64
	Dots        map[int]*Dot `json:"dots"`         // 点击点数据
}

// Dot 点击点数据
type Dot struct {
	X      int    `json:"x"`      // X 坐标
	Y      int    `json:"y"`      // Y 坐标
	Char   string `json:"char"`   // 字符
	Width  int    `json:"width"`  // 宽度
	Height int    `json:"height"` // 高度
	Angle  int    `json:"angle"`  // 角度
}

// RotateCaptchaData 旋转验证码数据结构
type RotateCaptchaData struct {
	ID          string `json:"id"`           // 验证码ID
	MasterImage string `json:"master_image"` // 主图 base64（需要旋转的图片）
	ThumbImage  string `json:"thumb_image"`  // 缩略图 base64（提示目标方向）
	Angle       int    `json:"angle"`        // 正确角度
}

// Generate 生成验证码:返回 id, base64图片, 答案, err
// 对于滑动验证码，b64s 包含 JSON 格式的 SlideCaptchaData
// 对于点击验证码，b64s 包含 JSON 格式的 ClickCaptchaData
// 对于旋转验证码，b64s 包含 JSON 格式的 RotateCaptchaData
func (c *Captcha) Generate() (id string, b64s string, answer string, err error) {
	// 如果是滑动验证码，使用特殊处理
	if c.config.DriverType == DriverSlide {
		return c.generateSlide()
	}

	// 如果是点击验证码，使用特殊处理
	if c.config.DriverType == DriverClick {
		return c.generateClick()
	}

	// 如果是旋转验证码，使用特殊处理
	if c.config.DriverType == DriverRotate {
		return c.generateRotate()
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

// initClickCaptcha 初始化点击验证码实例（懒加载）
func (c *Captcha) initClickCaptcha() error {
	if c.clickCaptcha != nil {
		return nil
	}

	clickCfg := c.config.ClickConfig
	if clickCfg == nil {
		clickCfg = DefaultClickConfig()
	}

	// 创建 builder
	builder := click.NewBuilder(
		click.WithImageSize(option.Size{Width: clickCfg.MasterWidth, Height: clickCfg.MasterHeight}),
		click.WithRangeLen(option.RangeVal{Min: clickCfg.CaptchaCount, Max: clickCfg.CaptchaCount}),
		click.WithRangeVerifyLen(option.RangeVal{Min: clickCfg.VerifyCount, Max: clickCfg.VerifyCount}),
		click.WithDisplayShadow(clickCfg.DisplayShadow),
		click.WithShadowColor(clickCfg.ShadowColor),
		click.WithShadowPoint(option.Point{X: clickCfg.ShadowOffsetX, Y: clickCfg.ShadowOffsetY}),
	)

	// 设置字符集
	chars := clickCfg.Chars
	if chars == "" {
		chars = "这的是随了机文我你他字在有不么中"
	}

	// 将字符串转换为字符数组
	charArr := make([]string, 0, len([]rune(chars)))
	for _, ch := range chars {
		charArr = append(charArr, string(ch))
	}

	builder.SetResources(
		click.WithChars(charArr),
	)

	c.clickCaptcha = builder.Make()
	return nil
}

// generateClick 生成点击文字验证码
func (c *Captcha) generateClick() (id string, b64s string, answer string, err error) {
	// 初始化点击验证码
	if err := c.initClickCaptcha(); err != nil {
		return "", "", "", fmt.Errorf("failed to init click captcha: %w", err)
	}

	// 生成验证码数据
	captData, err := c.clickCaptcha.Generate()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate click captcha: %w", err)
	}

	// 获取验证数据（包含点击点位置）
	dotData := captData.GetData()
	if dotData == nil {
		return "", "", "", fmt.Errorf("click captcha data is nil")
	}

	// 获取主图和缩略图的 base64
	masterBase64 := captData.GetMasterImage().ToBase64()
	thumbBase64 := captData.GetThumbImage().ToBase64()

	// 构造返回数据
	dots := make(map[int]*Dot)
	for idx, dot := range dotData {
		dots[idx] = &Dot{
			X:      dot.X,
			Y:      dot.Y,
			Char:   "",
			Width:  dot.Width,
			Height: dot.Height,
			Angle:  dot.Angle,
		}
	}

	clickData := &ClickCaptchaData{
		ID:          fmt.Sprintf("click_%d", time.Now().UnixNano()),
		MasterImage: masterBase64,
		ThumbImage:  thumbBase64,
		Dots:        dots,
	}

	// 将点击验证码数据序列化为 JSON
	jsonData, err := json.Marshal(clickData)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to marshal click data: %w", err)
	}

	// ID 用于 Redis 存储
	id = clickData.ID
	// b64s 包含完整的 JSON 数据
	b64s = string(jsonData)
	// answer 是点击点的坐标信息（JSON 格式）
	answerBytes, _ := json.Marshal(dotData)
	answer = string(answerBytes)

	return id, b64s, answer, nil
}

// initRotateCaptcha 初始化旋转验证码实例（懒加载）
func (c *Captcha) initRotateCaptcha() error {
	if c.rotateCaptcha != nil {
		return nil
	}

	// 创建 builder - 使用默认配置，不设置自定义资源
	// rotate 模块会使用内置的默认图片资源
	builder := rotate.NewBuilder()

	c.rotateCaptcha = builder.Make()
	return nil
}

// generateRotate 生成旋转验证码
func (c *Captcha) generateRotate() (id string, b64s string, answer string, err error) {
	// 初始化旋转验证码
	if err := c.initRotateCaptcha(); err != nil {
		return "", "", "", fmt.Errorf("failed to init rotate captcha: %w", err)
	}

	// 生成验证码数据
	captData, err := c.rotateCaptcha.Generate()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate rotate captcha: %w", err)
	}

	// 获取验证数据（包含正确角度）
	angleData := captData.GetData()
	if angleData == nil {
		return "", "", "", fmt.Errorf("rotate captcha data is nil")
	}

	// 获取主图和缩略图的 base64
	masterBase64 := captData.GetMasterImage().ToBase64()
	thumbBase64 := captData.GetThumbImage().ToBase64()

	// 构造返回数据
	rotateData := &RotateCaptchaData{
		ID:          fmt.Sprintf("rotate_%d", time.Now().UnixNano()),
		MasterImage: masterBase64,
		ThumbImage:  thumbBase64,
		Angle:       angleData.Angle,
	}

	// 将旋转验证码数据序列化为 JSON
	jsonData, err := json.Marshal(rotateData)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to marshal rotate data: %w", err)
	}

	// ID 用于 Redis 存储
	id = rotateData.ID
	// b64s 包含完整的 JSON 数据
	b64s = string(jsonData)
	// answer 是正确的角度值
	answer = fmt.Sprintf("%d", angleData.Angle)

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
	} else if c.config.DriverType == DriverClick {
		// 点击验证码需要特殊处理，比较坐标
		match = c.verifyClick(ans, userInput)
	} else if c.config.DriverType == DriverRotate {
		// 旋转验证码需要特殊处理，比较角度
		match = c.verifyRotate(ans, userInput)
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

// verifyClick 验证点击验证码（比较坐标）
func (c *Captcha) verifyClick(expected, actual string) bool {
	// 解析期望的点击数据
	var expectedDots []*click.Dot
	if err := json.Unmarshal([]byte(expected), &expectedDots); err != nil {
		return false
	}

	// 解析用户点击的数据
	var actualDots []map[string]interface{}
	if err := json.Unmarshal([]byte(actual), &actualDots); err != nil {
		return false
	}

	// 检查数量是否一致
	if len(expectedDots) != len(actualDots) {
		return false
	}

	// 验证每个点击点（允许 ±10 像素误差）
	for i, expectedDot := range expectedDots {
		if i >= len(actualDots) {
			return false
		}

		actualX, ok1 := actualDots[i]["x"].(float64)
		actualY, ok2 := actualDots[i]["y"].(float64)
		if !ok1 || !ok2 {
			return false
		}

		diffX := float64(expectedDot.X) - actualX
		diffY := float64(expectedDot.Y) - actualY
		if diffX < 0 {
			diffX = -diffX
		}
		if diffY < 0 {
			diffY = -diffY
		}

		// 允许 ±10 像素误差
		if diffX > 10 || diffY > 10 {
			return false
		}
	}

	return true
}

// verifyRotate 验证旋转验证码（比较角度）
func (c *Captcha) verifyRotate(expected, actual string) bool {
	var expectedAngle, actualAngle int
	if _, err := fmt.Sscanf(expected, "%d", &expectedAngle); err != nil {
		return false
	}
	if _, err := fmt.Sscanf(actual, "%d", &actualAngle); err != nil {
		return false
	}

	// 计算角度差（考虑360度循环）
	diff := expectedAngle - actualAngle
	if diff < 0 {
		diff = -diff
	}
	// 处理360度循环：如果差值大于180，说明应该从另一侧计算
	if diff > 180 {
		diff = 360 - diff
	}

	// 允许 ±5 度的误差
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
	} else if c.config.DriverType == DriverClick {
		return c.verifyClick(ans, userInput), nil
	} else if c.config.DriverType == DriverRotate {
		return c.verifyRotate(ans, userInput), nil
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
