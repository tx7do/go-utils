package captcha_test

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tx7do/go-utils/captcha"
)

// ExampleNewCaptcha_OptionsPattern 使用 Options 模式创建验证码（推荐）
func ExampleNewCaptcha_OptionsPattern() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 使用 Options 模式 - 字符串验证码
	captchaInstance := captcha.NewCaptcha(rdb,
		captcha.WithDriverType(captcha.DriverString),
		captcha.WithExpire(10*time.Minute),
		captcha.WithKeyPrefix("myapp:captcha"),
		captcha.WithStringCount(6),
		captcha.WithStringSource("ABCDEFGHJKLMNPQRSTUVWXYZ23456789"),
	)

	id, _, answer, _ := captchaInstance.Generate()
	fmt.Printf("字符串验证码ID: %s, 长度: %d\n", id, len(answer))
	// Output:
}

// ExampleNewCaptcha_DigitCaptcha 数字验证码示例
func ExampleNewCaptcha_DigitCaptcha() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	captchaInstance := captcha.NewCaptcha(rdb,
		captcha.WithDriverType(captcha.DriverDigit),
		captcha.WithDigitCount(4),
		captcha.WithDigitHeight(80),
		captcha.WithDigitWidth(240),
	)

	id, _, answer, _ := captchaInstance.Generate()
	fmt.Printf("数字验证码ID: %s, 答案: %s\n", id, answer)
	// Output:
}

// ExampleNewCaptcha_MathCaptcha 算术验证码示例
func ExampleNewCaptcha_MathCaptcha() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	captchaInstance := captcha.NewCaptcha(rdb,
		captcha.WithDriverType(captcha.DriverMath),
	)

	id, question, answer, _ := captchaInstance.Generate()
	fmt.Printf("算术验证码ID: %s\n", id)
	fmt.Printf("问题: %s\n", question)
	fmt.Printf("答案: %s\n", answer)
	// Output:
}

// ExampleNewCaptcha_ChineseCaptcha 中文验证码示例
func ExampleNewCaptcha_ChineseCaptcha() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	captchaInstance := captcha.NewCaptcha(rdb,
		captcha.WithDriverType(captcha.DriverChinese),
		captcha.WithChineseCount(4),
		captcha.WithChineseLanguage("zh"),
	)

	id, _, answer, _ := captchaInstance.Generate()
	fmt.Printf("中文验证码ID: %s, 字符数: %d\n", id, len([]rune(answer)))
	// Output:
}

// ExampleVerifyWithoutDelete 验证但不删除示例
func ExampleVerifyWithoutDelete() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	captchaInstance := captcha.NewCaptcha(rdb,
		captcha.WithExpire(5*time.Minute),
		captcha.WithKeyPrefix("myapp:captcha"),
	)
	id, _, answer, _ := captchaInstance.Generate()

	ctx := context.Background()
	_ = captchaInstance.Save(ctx, id, answer)

	// 第一次验证（不删除）
	isValid1, _ := captchaInstance.VerifyWithoutDelete(ctx, id, answer)
	fmt.Printf("第一次验证: %v\n", isValid1)

	// 第二次验证（仍然有效）
	isValid2, _ := captchaInstance.VerifyWithoutDelete(ctx, id, answer)
	fmt.Printf("第二次验证: %v\n", isValid2)

	// Output:
	// 第一次验证: true
	// 第二次验证: true
}

// ExampleGetRemainingTime 获取剩余时间示例
func ExampleGetRemainingTime() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	captchaInstance := captcha.NewCaptcha(rdb,
		captcha.WithExpire(5*time.Minute),
		captcha.WithKeyPrefix("myapp:captcha"),
	)
	id, _, answer, _ := captchaInstance.Generate()

	ctx := context.Background()
	_ = captchaInstance.Save(ctx, id, answer)

	// 获取剩余时间
	ttl, err := captchaInstance.GetRemainingTime(ctx, id)
	if err != nil {
		panic(err)
	}

	fmt.Printf("验证码剩余时间: %v\n", ttl)
	// Output:
}
