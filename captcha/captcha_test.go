package captcha

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRedis 创建测试用的 Redis 客户端
func setupTestRedis(t *testing.T) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // 使用数据库1，避免影响其他数据
	})

	// 测试连接
	ctx := context.Background()
	err := rdb.Ping(ctx).Err()
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}

	// 清理测试数据
	_ = rdb.FlushDB(ctx).Err()

	return rdb
}

// teardownTestRedis 清理测试数据
func teardownTestRedis(rdb *redis.Client) {
	ctx := context.Background()
	_ = rdb.FlushDB(ctx).Err()
	_ = rdb.Close()
}

func TestNewCaptcha(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	t.Run("默认配置", func(t *testing.T) {
		captchaInstance := NewCaptcha(rdb)
		assert.NotNil(t, captchaInstance)
		assert.Equal(t, DriverDigit, captchaInstance.config.DriverType)
		assert.Equal(t, 5*time.Minute, captchaInstance.config.Expire)
		assert.Equal(t, "captcha", captchaInstance.config.KeyPrefix)
	})

	t.Run("自定义选项", func(t *testing.T) {
		captchaInstance := NewCaptcha(rdb,
			WithDriverType(DriverString),
			WithExpire(10*time.Minute),
			WithKeyPrefix("test:captcha"),
		)
		assert.NotNil(t, captchaInstance)
		assert.Equal(t, DriverString, captchaInstance.config.DriverType)
		assert.Equal(t, 10*time.Minute, captchaInstance.config.Expire)
		assert.Equal(t, "test:captcha", captchaInstance.config.KeyPrefix)
	})
}

func TestNewCaptchaWithConfig(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	t.Run("正常配置", func(t *testing.T) {
		config := DefaultConfig()
		config.DriverType = DriverMath
		config.Expire = 15 * time.Minute

		captchaInstance := NewCaptchaWithConfig(rdb, config)
		assert.NotNil(t, captchaInstance)
		assert.Equal(t, DriverMath, captchaInstance.config.DriverType)
		assert.Equal(t, 15*time.Minute, captchaInstance.config.Expire)
	})

	t.Run("nil配置", func(t *testing.T) {
		captchaInstance := NewCaptchaWithConfig(rdb, nil)
		assert.NotNil(t, captchaInstance)
		assert.Equal(t, DriverDigit, captchaInstance.config.DriverType)
	})
}

func TestGenerate_DigitCaptcha(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithDriverType(DriverDigit),
		WithDigitCount(4),
	)

	id, b64Image, answer, err := captchaInstance.Generate()
	require.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.NotEmpty(t, b64Image)
	assert.NotEmpty(t, answer)
	assert.Len(t, answer, 4)
}

func TestGenerate_StringCaptcha(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithDriverType(DriverString),
		WithStringCount(6),
		WithStringSource("ABCDEF"),
	)

	id, b64Image, answer, err := captchaInstance.Generate()
	require.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.NotEmpty(t, b64Image)
	assert.NotEmpty(t, answer)
	assert.Len(t, answer, 6)
}

func TestGenerate_MathCaptcha(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithDriverType(DriverMath),
	)

	id, question, answer, err := captchaInstance.Generate()
	require.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.NotEmpty(t, question)
	assert.NotEmpty(t, answer)
	// 算术验证码的问题应该包含运算符
	assert.Contains(t, question, "+")
}

func TestGenerate_ChineseCaptcha(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithDriverType(DriverChinese),
		WithChineseCount(4),
		WithChineseLanguage("zh"),
	)

	id, b64Image, answer, err := captchaInstance.Generate()
	require.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.NotEmpty(t, b64Image)
	assert.NotEmpty(t, answer)
	// 中文字符数应该是4
	assert.Equal(t, 4, len([]rune(answer)))
}

func TestGenerate_SlideCaptcha(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithDriverType(DriverSlide),
		WithSlideMasterSize(300, 220),
		WithSlideTileSize(60, 60),
	)

	id, jsonData, answer, err := captchaInstance.Generate()
	require.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.NotEmpty(t, jsonData)
	assert.NotEmpty(t, answer)

	// 解析 JSON 数据
	var slideData SlideCaptchaData
	err = json.Unmarshal([]byte(jsonData), &slideData)
	require.NoError(t, err)

	assert.Equal(t, id, slideData.ID)
	assert.NotEmpty(t, slideData.MasterImage)
	assert.NotEmpty(t, slideData.TileImage)
	assert.Greater(t, slideData.XPosition, 0)

	// answer 应该是 X 坐标的字符串形式
	assert.Equal(t, answer, fmt.Sprintf("%d", slideData.XPosition))
}

func TestSaveAndVerify(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithExpire(5*time.Minute),
		WithKeyPrefix("test"),
	)

	ctx := context.Background()

	// 生成验证码
	id, _, answer, err := captchaInstance.Generate()
	require.NoError(t, err)

	// 保存验证码
	err = captchaInstance.Save(ctx, id, answer)
	require.NoError(t, err)

	// 验证正确答案
	valid, err := captchaInstance.Verify(ctx, id, answer)
	require.NoError(t, err)
	assert.True(t, valid)

	// 验证后应该被删除，再次验证应该失败
	valid, err = captchaInstance.Verify(ctx, id, answer)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestSaveAndVerify_SlideCaptcha(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithDriverType(DriverSlide),
		WithExpire(5*time.Minute),
		WithKeyPrefix("test:slide"),
	)

	ctx := context.Background()

	// 生成滑动验证码
	id, jsonData, answer, err := captchaInstance.Generate()
	require.NoError(t, err)

	// 解析获取 X 位置
	var slideData SlideCaptchaData
	err = json.Unmarshal([]byte(jsonData), &slideData)
	require.NoError(t, err)

	// 保存验证码
	err = captchaInstance.Save(ctx, id, answer)
	require.NoError(t, err)

	// 验证正确答案（精确匹配）
	valid, err := captchaInstance.Verify(ctx, id, answer)
	require.NoError(t, err)
	assert.True(t, valid)

	// 验证后应该被删除，再次验证应该失败
	valid, err = captchaInstance.Verify(ctx, id, answer)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestVerify_SlideCaptcha_WithTolerance(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithDriverType(DriverSlide),
		WithExpire(5*time.Minute),
		WithKeyPrefix("test:slide"),
	)

	ctx := context.Background()

	// 生成滑动验证码
	id, jsonData, answer, err := captchaInstance.Generate()
	require.NoError(t, err)

	// 解析获取 X 位置
	var slideData SlideCaptchaData
	err = json.Unmarshal([]byte(jsonData), &slideData)
	require.NoError(t, err)

	// 保存验证码
	err = captchaInstance.Save(ctx, id, answer)
	require.NoError(t, err)

	// 模拟用户滑动，允许 ±5 像素误差
	var expectedX int
	fmt.Sscanf(answer, "%d", &expectedX)

	// 测试在误差范围内的值
	closeValue := fmt.Sprintf("%d", expectedX+3) // +3 像素
	valid, err := captchaInstance.Verify(ctx, id, closeValue)
	require.NoError(t, err)
	assert.True(t, valid, "应该接受在误差范围内的值")
}

func TestGenerate_ClickCaptcha(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithDriverType(DriverClick),
		WithClickMasterSize(300, 220),
		WithClickCaptchaCount(6),
		WithClickVerifyCount(3),
	)

	id, jsonData, answer, err := captchaInstance.Generate()
	require.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.NotEmpty(t, jsonData)
	assert.NotEmpty(t, answer)

	// 解析 JSON 数据
	var clickData ClickCaptchaData
	err = json.Unmarshal([]byte(jsonData), &clickData)
	require.NoError(t, err)

	assert.Equal(t, id, clickData.ID)
	assert.NotEmpty(t, clickData.MasterImage)
	assert.NotEmpty(t, clickData.ThumbImage)
	assert.NotEmpty(t, clickData.Dots)
	assert.Greater(t, len(clickData.Dots), 0)
}

func TestSaveAndVerify_ClickCaptcha(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithDriverType(DriverClick),
		WithExpire(5*time.Minute),
		WithKeyPrefix("test:click"),
	)

	ctx := context.Background()

	// 生成点击验证码
	id, jsonData, answer, err := captchaInstance.Generate()
	require.NoError(t, err)

	// 解析获取点击点数据
	var clickData ClickCaptchaData
	err = json.Unmarshal([]byte(jsonData), &clickData)
	require.NoError(t, err)

	// 保存验证码
	err = captchaInstance.Save(ctx, id, answer)
	require.NoError(t, err)

	// 构造用户点击数据（模拟正确的点击）
	// 注意：这里简化处理，实际应该从前端获取用户点击的坐标
	userClicks := make([]map[string]interface{}, 0)
	for _, dot := range clickData.Dots {
		userClicks = append(userClicks, map[string]interface{}{
			"x": float64(dot.X),
			"y": float64(dot.Y),
		})
	}
	userClicksJSON, _ := json.Marshal(userClicks)

	// 验证正确答案
	valid, err := captchaInstance.Verify(ctx, id, string(userClicksJSON))
	require.NoError(t, err)
	assert.True(t, valid)

	// 验证后应该被删除，再次验证应该失败
	valid, err = captchaInstance.Verify(ctx, id, string(userClicksJSON))
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestVerify_ClickCaptcha_WithTolerance(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithDriverType(DriverClick),
		WithExpire(5*time.Minute),
		WithKeyPrefix("test:click"),
	)

	ctx := context.Background()

	// 生成点击验证码
	id, jsonData, answer, err := captchaInstance.Generate()
	require.NoError(t, err)

	// 解析获取点击点数据
	var clickData ClickCaptchaData
	err = json.Unmarshal([]byte(jsonData), &clickData)
	require.NoError(t, err)

	// 保存验证码
	err = captchaInstance.Save(ctx, id, answer)
	require.NoError(t, err)

	// 构造用户点击数据（带 ±8 像素误差，在允许范围内）
	userClicks := make([]map[string]interface{}, 0)
	for _, dot := range clickData.Dots {
		userClicks = append(userClicks, map[string]interface{}{
			"x": float64(dot.X + 8), // +8 像素误差
			"y": float64(dot.Y + 8),
		})
	}
	userClicksJSON, _ := json.Marshal(userClicks)

	// 验证应该成功（在 ±10 像素误差范围内）
	valid, err := captchaInstance.Verify(ctx, id, string(userClicksJSON))
	require.NoError(t, err)
	assert.True(t, valid, "应该接受在误差范围内的点击")
}

func TestVerify_WrongAnswer(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithExpire(5*time.Minute),
		WithKeyPrefix("test"),
	)

	ctx := context.Background()

	// 生成并保存验证码
	id, _, answer, err := captchaInstance.Generate()
	require.NoError(t, err)
	err = captchaInstance.Save(ctx, id, answer)
	require.NoError(t, err)

	// 验证错误答案
	valid, err := captchaInstance.Verify(ctx, id, "wrong")
	require.NoError(t, err)
	assert.False(t, valid)

	// 验证码应该仍然存在（因为验证失败不会删除）
	exists, err := captchaInstance.Exists(ctx, id)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestVerifyWithoutDelete(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithExpire(5*time.Minute),
		WithKeyPrefix("test"),
	)

	ctx := context.Background()

	// 生成并保存验证码
	id, _, answer, err := captchaInstance.Generate()
	require.NoError(t, err)
	err = captchaInstance.Save(ctx, id, answer)
	require.NoError(t, err)

	// 第一次验证（不删除）
	valid1, err := captchaInstance.VerifyWithoutDelete(ctx, id, answer)
	require.NoError(t, err)
	assert.True(t, valid1)

	// 第二次验证（仍然有效）
	valid2, err := captchaInstance.VerifyWithoutDelete(ctx, id, answer)
	require.NoError(t, err)
	assert.True(t, valid2)

	// 验证码应该仍然存在
	exists, err := captchaInstance.Exists(ctx, id)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestVerifyExpired(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithExpire(1*time.Second), // 1秒过期
		WithKeyPrefix("test"),
	)

	ctx := context.Background()

	// 生成并保存验证码
	id, _, answer, err := captchaInstance.Generate()
	require.NoError(t, err)
	err = captchaInstance.Save(ctx, id, answer)
	require.NoError(t, err)

	// 等待过期
	time.Sleep(2 * time.Second)

	// 验证应该失败
	valid, err := captchaInstance.Verify(ctx, id, answer)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestDelete(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithExpire(5*time.Minute),
		WithKeyPrefix("test"),
	)

	ctx := context.Background()

	// 生成并保存验证码
	id, _, answer, err := captchaInstance.Generate()
	require.NoError(t, err)
	err = captchaInstance.Save(ctx, id, answer)
	require.NoError(t, err)

	// 确认存在
	exists, err := captchaInstance.Exists(ctx, id)
	require.NoError(t, err)
	assert.True(t, exists)

	// 删除验证码
	err = captchaInstance.Delete(ctx, id)
	require.NoError(t, err)

	// 确认已删除
	exists, err = captchaInstance.Exists(ctx, id)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestExists(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithExpire(5*time.Minute),
		WithKeyPrefix("test"),
	)

	ctx := context.Background()

	// 生成并保存验证码
	id, _, answer, err := captchaInstance.Generate()
	require.NoError(t, err)
	err = captchaInstance.Save(ctx, id, answer)
	require.NoError(t, err)

	// 检查存在
	exists, err := captchaInstance.Exists(ctx, id)
	require.NoError(t, err)
	assert.True(t, exists)

	// 检查不存在的ID
	exists, err = captchaInstance.Exists(ctx, "nonexistent")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestGetRemainingTime(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	expireTime := 10 * time.Second
	captchaInstance := NewCaptcha(rdb,
		WithExpire(expireTime),
		WithKeyPrefix("test"),
	)

	ctx := context.Background()

	// 生成并保存验证码
	id, _, answer, err := captchaInstance.Generate()
	require.NoError(t, err)
	err = captchaInstance.Save(ctx, id, answer)
	require.NoError(t, err)

	// 获取剩余时间
	ttl, err := captchaInstance.GetRemainingTime(ctx, id)
	require.NoError(t, err)
	assert.True(t, ttl > 0*time.Second)
	assert.True(t, ttl <= expireTime)

	// 等待一段时间后再次检查
	time.Sleep(2 * time.Second)
	ttl2, err := captchaInstance.GetRemainingTime(ctx, id)
	require.NoError(t, err)
	assert.True(t, ttl2 < ttl)
}

func TestGetConfig(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	config := DefaultConfig()
	config.DriverType = DriverString
	config.Expire = 20 * time.Minute

	captchaInstance := NewCaptchaWithConfig(rdb, config)

	retrievedConfig := captchaInstance.GetConfig()
	assert.Equal(t, DriverString, retrievedConfig.DriverType)
	assert.Equal(t, 20*time.Minute, retrievedConfig.Expire)
}

func TestSetConfig(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb)

	// 初始配置
	assert.Equal(t, DriverDigit, captchaInstance.GetConfig().DriverType)

	// 设置新配置
	newConfig := DefaultConfig()
	newConfig.DriverType = DriverMath
	captchaInstance.SetConfig(newConfig)

	assert.Equal(t, DriverMath, captchaInstance.GetConfig().DriverType)
}

func TestSetConfig_Nil(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb)
	originalConfig := captchaInstance.GetConfig()

	// 设置nil应该不影响现有配置
	captchaInstance.SetConfig(nil)
	assert.Equal(t, originalConfig, captchaInstance.GetConfig())
}

func TestOptions_DigitConfig(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithDriverType(DriverDigit),
		WithDigitHeight(100),
		WithDigitWidth(300),
		WithDigitCount(6),
		WithDigitMaxSkew(0.8),
		WithDigitDotCount(100),
	)

	cfg := captchaInstance.GetConfig()
	assert.Equal(t, 100, cfg.DigitConfig.Height)
	assert.Equal(t, 300, cfg.DigitConfig.Width)
	assert.Equal(t, 6, cfg.DigitConfig.CaptchaCount)
	assert.Equal(t, 0.8, cfg.DigitConfig.MaxSkew)
	assert.Equal(t, 100, cfg.DigitConfig.DotCount)
}

func TestOptions_StringConfig(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithDriverType(DriverString),
		WithStringHeight(120),
		WithStringWidth(350),
		WithStringCount(8),
		WithStringSource("XYZ"),
		WithStringDotCount(120),
	)

	cfg := captchaInstance.GetConfig()
	assert.Equal(t, 120, cfg.StringConfig.Height)
	assert.Equal(t, 350, cfg.StringConfig.Width)
	assert.Equal(t, 8, cfg.StringConfig.CaptchaCount)
	assert.Equal(t, "XYZ", cfg.StringConfig.Source)
	assert.Equal(t, 120, cfg.StringConfig.DotCount)
}

func TestOptions_MathConfig(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithDriverType(DriverMath),
		WithMathHeight(90),
		WithMathWidth(280),
		WithMathDotCount(90),
	)

	cfg := captchaInstance.GetConfig()
	assert.Equal(t, 90, cfg.MathConfig.Height)
	assert.Equal(t, 280, cfg.MathConfig.Width)
	assert.Equal(t, 90, cfg.MathConfig.DotCount)
}

func TestOptions_ChineseConfig(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithDriverType(DriverChinese),
		WithChineseHeight(110),
		WithChineseWidth(320),
		WithChineseCount(5),
		WithChineseLanguage("en"),
		WithChineseDotCount(110),
	)

	cfg := captchaInstance.GetConfig()
	assert.Equal(t, 110, cfg.ChineseConfig.Height)
	assert.Equal(t, 320, cfg.ChineseConfig.Width)
	assert.Equal(t, 5, cfg.ChineseConfig.CaptchaCount)
	assert.Equal(t, "en", cfg.ChineseConfig.Language)
	assert.Equal(t, 110, cfg.ChineseConfig.DotCount)
}

func TestOptions_SlideConfig(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithDriverType(DriverSlide),
		WithSlideMasterSize(350, 250),
		WithSlideTileSize(70, 70),
		WithSlideTileRadius(8),
		WithSlideJigsawRadius(12),
		WithSlideShadow(6, 6, 12),
	)

	cfg := captchaInstance.GetConfig()
	assert.Equal(t, 350, cfg.SlideConfig.MasterWidth)
	assert.Equal(t, 250, cfg.SlideConfig.MasterHeight)
	assert.Equal(t, 70, cfg.SlideConfig.TileWidth)
	assert.Equal(t, 70, cfg.SlideConfig.TileHeight)
	assert.Equal(t, 8, cfg.SlideConfig.TileRadius)
	assert.Equal(t, 12, cfg.SlideConfig.JigsawRadius)
	assert.Equal(t, 6, cfg.SlideConfig.ShadowOffsetX)
	assert.Equal(t, 6, cfg.SlideConfig.ShadowOffsetY)
	assert.Equal(t, 12, cfg.SlideConfig.ShadowBlur)
}

func TestOptions_ClickConfig(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb,
		WithDriverType(DriverClick),
		WithClickMasterSize(350, 250),
		WithClickThumbSize(180, 50),
		WithClickCaptchaCount(8),
		WithClickVerifyCount(4),
		WithClickChars("ABCDEFGHIJK"),
		WithClickLanguage("en"),
		WithClickShadow(true, "#FF0000", 3, 3),
	)

	cfg := captchaInstance.GetConfig()
	assert.Equal(t, 350, cfg.ClickConfig.MasterWidth)
	assert.Equal(t, 250, cfg.ClickConfig.MasterHeight)
	assert.Equal(t, 180, cfg.ClickConfig.ThumbWidth)
	assert.Equal(t, 50, cfg.ClickConfig.ThumbHeight)
	assert.Equal(t, 8, cfg.ClickConfig.CaptchaCount)
	assert.Equal(t, 4, cfg.ClickConfig.VerifyCount)
	assert.Equal(t, "ABCDEFGHIJK", cfg.ClickConfig.Chars)
	assert.Equal(t, "en", cfg.ClickConfig.Language)
	assert.True(t, cfg.ClickConfig.DisplayShadow)
	assert.Equal(t, "#FF0000", cfg.ClickConfig.ShadowColor)
	assert.Equal(t, 3, cfg.ClickConfig.ShadowOffsetX)
	assert.Equal(t, 3, cfg.ClickConfig.ShadowOffsetY)
}

func TestDefaultConfigs(t *testing.T) {
	t.Run("DefaultDigitConfig", func(t *testing.T) {
		cfg := DefaultDigitConfig()
		assert.Equal(t, 80, cfg.Height)
		assert.Equal(t, 240, cfg.Width)
		assert.Equal(t, 4, cfg.CaptchaCount)
	})

	t.Run("DefaultStringConfig", func(t *testing.T) {
		cfg := DefaultStringConfig()
		assert.Equal(t, 80, cfg.Height)
		assert.Equal(t, 240, cfg.Width)
		assert.NotEmpty(t, cfg.Source)
	})

	t.Run("DefaultMathConfig", func(t *testing.T) {
		cfg := DefaultMathConfig()
		assert.Equal(t, 80, cfg.Height)
		assert.Equal(t, 240, cfg.Width)
	})

	t.Run("DefaultChineseConfig", func(t *testing.T) {
		cfg := DefaultChineseConfig()
		assert.Equal(t, 80, cfg.Height)
		assert.Equal(t, 240, cfg.Width)
		assert.Equal(t, "zh", cfg.Language)
	})

	t.Run("DefaultSlideConfig", func(t *testing.T) {
		cfg := DefaultSlideConfig()
		assert.Equal(t, 300, cfg.MasterWidth)
		assert.Equal(t, 220, cfg.MasterHeight)
		assert.Equal(t, 60, cfg.TileWidth)
		assert.Equal(t, 60, cfg.TileHeight)
	})

	t.Run("DefaultClickConfig", func(t *testing.T) {
		cfg := DefaultClickConfig()
		assert.Equal(t, 300, cfg.MasterWidth)
		assert.Equal(t, 220, cfg.MasterHeight)
		assert.Equal(t, 150, cfg.ThumbWidth)
		assert.Equal(t, 40, cfg.ThumbHeight)
		assert.Equal(t, 6, cfg.CaptchaCount)
		assert.Equal(t, 3, cfg.VerifyCount)
		assert.NotEmpty(t, cfg.Chars)
	})

	t.Run("DefaultConfig", func(t *testing.T) {
		cfg := DefaultConfig()
		assert.Equal(t, DriverDigit, cfg.DriverType)
		assert.Equal(t, 5*time.Minute, cfg.Expire)
		assert.Equal(t, "captcha", cfg.KeyPrefix)
		assert.NotNil(t, cfg.DigitConfig)
		assert.NotNil(t, cfg.StringConfig)
		assert.NotNil(t, cfg.MathConfig)
		assert.NotNil(t, cfg.ChineseConfig)
		assert.NotNil(t, cfg.SlideConfig)
		assert.NotNil(t, cfg.ClickConfig)
	})
}

func TestMultipleCaptchas(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	ctx := context.Background()

	// 创建多个验证码实例
	captchaInstance1 := NewCaptcha(rdb, WithKeyPrefix("app1"))
	captchaInstance2 := NewCaptcha(rdb, WithKeyPrefix("app2"))

	// 生成并保存
	id1, _, ans1, _ := captchaInstance1.Generate()
	id2, _, ans2, _ := captchaInstance2.Generate()

	_ = captchaInstance1.Save(ctx, id1, ans1)
	_ = captchaInstance2.Save(ctx, id2, ans2)

	// 验证各自的验证码
	valid1, _ := captchaInstance1.Verify(ctx, id1, ans1)
	valid2, _ := captchaInstance2.Verify(ctx, id2, ans2)

	assert.True(t, valid1)
	assert.True(t, valid2)
}

func TestConcurrentAccess(t *testing.T) {
	rdb := setupTestRedis(t)
	defer teardownTestRedis(rdb)

	captchaInstance := NewCaptcha(rdb, WithKeyPrefix("concurrent"))
	ctx := context.Background()

	// 并发生成和验证
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			id, _, ans, err := captchaInstance.Generate()
			if err != nil {
				t.Error(err)
				done <- false
				return
			}

			_ = captchaInstance.Save(ctx, id, ans)
			valid, err := captchaInstance.Verify(ctx, id, ans)
			if err != nil {
				t.Error(err)
				done <- false
				return
			}

			if !valid {
				t.Error("Verification failed")
				done <- false
				return
			}

			done <- true
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		result := <-done
		assert.True(t, result)
	}
}
