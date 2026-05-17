# Captcha 验证码模块

一个高度可定制的验证码模块，支持多种驱动类型和丰富的配置选项。

## 特性

- ✅ 支持多种验证码类型：数字、字符串、算术、中文
- ✅ 完全可定制的外观和行为
- ✅ 基于 Redis 的存储和验证
- ✅ 自动过期管理
- ✅ 灵活的配置选项

## 安装

```bash
go get github.com/tx7do/go-utils/captcha
```

## 快速开始

### 使用 Options 模式（推荐）

```go
package main

import (
    "context"
    "time"
    
    "github.com/redis/go-redis/v9"
    "github.com/tx7do/go-utils/captcha"
)

func main() {
    // 创建 Redis 客户端
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // 使用 Options 模式创建验证码实例
    captchaInstance := captcha.NewCaptcha(rdb,
        captcha.WithDriverType(captcha.DriverString),
        captcha.WithExpire(10*time.Minute),
        captcha.WithKeyPrefix("myapp:captcha"),
        captcha.WithStringCount(6),
        captcha.WithStringSource("ABCDEFGHJKLMNPQRSTUVWXYZ23456789"),
    )
    
    // 生成验证码
    id, b64Image, answer, err := captchaInstance.Generate()
    if err != nil {
        panic(err)
    }
    
    // 保存验证码到 Redis
    ctx := context.Background()
    err = captchaInstance.Save(ctx, id, answer)
    if err != nil {
        panic(err)
    }
    
    // 验证用户输入
    isValid, err := captchaInstance.Verify(ctx, id, userInput)
    if err != nil {
        panic(err)
    }
    
    if isValid {
        println("验证成功")
    } else {
        println("验证失败")
    }
}
```

### 使用配置对象

```go
package main

import (
    "context"
    "time"
    
    "github.com/redis/go-redis/v9"
    "github.com/tx7do/go-utils/captcha"
)

func main() {
    // 创建 Redis 客户端
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // 创建自定义配置
    config := captcha.DefaultConfig()
    config.DriverType = captcha.DriverString
    config.Expire = 10 * time.Minute
    config.KeyPrefix = "myapp:captcha"
    
    // 自定义字符串验证码配置
    config.StringConfig = &captcha.StringConfig{
        Height:       100,
        Width:        300,
        CaptchaCount: 6,
        DotCount:     100,
        Source:       "ABCDEFGHJKLMNPQRSTUVWXYZ23456789",
    }
    
    // 创建验证码实例
    captchaInstance := captcha.NewCaptchaWithConfig(rdb, config)
    
    // ... 其他操作同上
}
```

## 支持的验证码类型

### 1. 数字验证码 (DriverDigit)

**使用 Options：**
```go
captchaInstance := captcha.NewCaptcha(rdb,
    captcha.WithDriverType(captcha.DriverDigit),
    captcha.WithDigitCount(4),
    captcha.WithDigitHeight(80),
    captcha.WithDigitWidth(240),
)
```

**使用配置对象：**
```go
config := captcha.DefaultConfig()
config.DriverType = captcha.DriverDigit
config.DigitConfig = &captcha.DigitConfig{
    Height:       80,
    Width:        240,
    CaptchaCount: 4,      // 4位数字
    MaxSkew:      0.7,
    DotCount:     80,
}
captchaInstance := captcha.NewCaptchaWithConfig(rdb, config)
```

### 2. 字符串验证码 (DriverString)

**使用 Options：**
```go
captchaInstance := captcha.NewCaptcha(rdb,
    captcha.WithDriverType(captcha.DriverString),
    captcha.WithStringCount(6),
    captcha.WithStringSource("ABCDEFGHJKLMNPQRSTUVWXYZ23456789"),
    captcha.WithStringHeight(100),
    captcha.WithStringWidth(300),
)
```

**使用配置对象：**
config := captcha.DefaultConfig()
config.DriverType = captcha.DriverString
config.StringConfig = &captcha.StringConfig{
    Height:       80,
    Width:        240,
    CaptchaCount: 6,      // 6个字符
    MaxSkew:      0.7,
    DotCount:     80,
    Source:       "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
}
```

### 3. 算术验证码 (DriverMath)

**使用 Options：**
```go
captchaInstance := captcha.NewCaptcha(rdb,
    captcha.WithDriverType(captcha.DriverMath),
    captcha.WithMathHeight(80),
    captcha.WithMathWidth(240),
)
// 生成的验证码类似: "3 + 5 = ?"，答案是 "8"
```

**使用配置对象：**
config := captcha.DefaultConfig()
config.DriverType = captcha.DriverMath
config.MathConfig = &captcha.MathConfig{
    Height:       80,
    Width:        240,
    CaptchaCount: 4,
    MaxSkew:      0.7,
    DotCount:     80,
}
// 生成的验证码类似: "3 + 5 = ?"，答案是 "8"
```

### 4. 中文验证码 (DriverChinese)

**使用 Options：**
```go
captchaInstance := captcha.NewCaptcha(rdb,
    captcha.WithDriverType(captcha.DriverChinese),
    captcha.WithChineseCount(4),
    captcha.WithChineseLanguage("zh"),
)
```

**使用配置对象：**
config := captcha.DefaultConfig()
config.DriverType = captcha.DriverChinese
config.ChineseConfig = &captcha.ChineseConfig{
    Height:       80,
    Width:        240,
    CaptchaCount: 4,      // 4个汉字
    MaxSkew:      0.7,
    DotCount:     80,
    Language:     "zh",   // 中文
}
```

## 配置选项详解

### Options 函数列表（推荐）

**通用选项：**
- `WithDriverType(driverType DriverType)` - 设置驱动类型
- `WithExpire(expire time.Duration)` - 设置过期时间
- `WithKeyPrefix(prefix string)` - 设置 Redis key 前缀

**数字验证码选项：**
- `WithDigitHeight(height int)` - 设置高度
- `WithDigitWidth(width int)` - 设置宽度
- `WithDigitCount(count int)` - 设置字符数量
- `WithDigitMaxSkew(skew float64)` - 设置最大倾斜度
- `WithDigitDotCount(count int)` - 设置干扰点数量
- `WithDigitConfig(config *DigitConfig)` - 直接设置完整配置

**字符串验证码选项：**
- `WithStringHeight(height int)` - 设置高度
- `WithStringWidth(width int)` - 设置宽度
- `WithStringCount(count int)` - 设置字符数量
- `WithStringSource(source string)` - 设置字符源
- `WithStringDotCount(count int)` - 设置干扰点数量
- `WithStringConfig(config *StringConfig)` - 直接设置完整配置

**算术验证码选项：**
- `WithMathHeight(height int)` - 设置高度
- `WithMathWidth(width int)` - 设置宽度
- `WithMathDotCount(count int)` - 设置干扰点数量
- `WithMathConfig(config *MathConfig)` - 直接设置完整配置

**中文验证码选项：**
- `WithChineseHeight(height int)` - 设置高度
- `WithChineseWidth(width int)` - 设置宽度
- `WithChineseCount(count int)` - 设置字符数量
- `WithChineseLanguage(language string)` - 设置语言 (zh/en)
- `WithChineseDotCount(count int)` - 设置干扰点数量
- `WithChineseConfig(config *ChineseConfig)` - 直接设置完整配置

### 配置对象结构

### 通用配置 (Config)

| 字段 | 类型 | 说明 |
|------|------|------|
| DriverType | DriverType | 验证码驱动类型 |
| Expire | time.Duration | 验证码过期时间 |
| KeyPrefix | string | Redis key 前缀 |
| DigitConfig | *DigitConfig | 数字验证码配置 |
| StringConfig | *StringConfig | 字符串验证码配置 |
| MathConfig | *MathConfig | 算术验证码配置 |
| ChineseConfig | *ChineseConfig | 中文验证码配置 |

### 各驱动配置项

所有驱动配置都包含以下字段：

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| Height | int | 图片高度 | 80 |
| Width | int | 图片宽度 | 240 |
| CaptchaCount | int | 验证码字符数量 | 4 |
| MaxSkew | float64 | 最大倾斜度 | 0.7 |
| DotCount | int | 干扰点数量 | 80 |
| BgColorR/G/B | uint8 | 背景色 RGB | 255, 255, 255 (白色) |
| FontColorR/G/B | uint8 | 字体色 RGB | 0, 0, 0 (黑色) |

字符串验证码额外字段：
- `Source`: 字符源字符串

中文验证码额外字段：
- `Language`: 语言类型 ("zh" 或 "en")

## API 参考

### 构造函数

- `NewCaptcha(rdb *redis.Client, expire time.Duration, keyPrefix string) *Captcha`
  - 创建验证码实例（兼容旧接口）
  
- `NewCaptchaWithConfig(rdb *redis.Client, config *Config) *Captcha`
  - 使用自定义配置创建验证码实例

### 核心方法

- `Generate() (string, string, string, error)`
  - 生成验证码
  - 返回: (id, base64图片, 答案, 错误)

- `Save(ctx context.Context, captchaID, answer string) error`
  - 将验证码答案存入 Redis

- `Verify(ctx context.Context, captchaID, userInput string) (bool, error)`
  - 验证用户输入，验证成功后自动删除验证码
  - 返回: (是否匹配, 错误)

### 扩展方法

- `VerifyAndDelete(ctx context.Context, captchaID, userInput string) (bool, error)`
  - 验证并删除验证码（与 Verify 相同）

- `VerifyWithoutDelete(ctx context.Context, captchaID, userInput string) (bool, error)`
  - 验证但不删除验证码（用于多次验证场景）

- `Delete(ctx context.Context, captchaID string) error`
  - 手动删除验证码

- `Exists(ctx context.Context, captchaID string) (bool, error)`
  - 检查验证码是否存在

- `GetRemainingTime(ctx context.Context, captchaID string) (time.Duration, error)`
  - 获取验证码剩余时间

- `GetConfig() *Config`
  - 获取当前配置

- `SetConfig(config *Config)`
  - 设置配置

## 最佳实践

### 1. 排除易混淆字符

```go
config.StringConfig.Source = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 排除 I, l, 1, O, 0 等
```

### 2. 调整难度

```go
// 更难的验证码
config.DigitConfig.MaxSkew = 1.0    // 更大的倾斜
config.DigitConfig.DotCount = 150   // 更多干扰点
config.DigitConfig.CaptchaCount = 6 // 更多位数

// 更简单的验证码
config.DigitConfig.MaxSkew = 0.3    // 更小的倾斜
config.DigitConfig.DotCount = 30    // 更少干扰点
config.DigitConfig.CaptchaCount = 3 // 更少位数
```

### 3. 自定义颜色

```go
// 深色主题
config.DigitConfig.BgColorR = 33
config.DigitConfig.BgColorG = 33
config.DigitConfig.BgColorB = 33
config.DigitConfig.FontColorR = 255
config.DigitConfig.FontColorG = 255
config.DigitConfig.FontColorB = 255
```

### 4. 动态切换验证码类型

```go
// 根据安全等级动态切换
if securityLevel == "high" {
    config.DriverType = captcha.DriverChinese
} else {
    config.DriverType = captcha.DriverDigit
}
cap.SetConfig(config)
```

## 注意事项

1. **Redis 依赖**: 确保 Redis 服务正常运行
2. **过期时间**: 合理设置过期时间，避免验证码长期有效
3. **一次性使用**: 默认验证后会自动删除，防止重放攻击
4. **字符源**: 字符串验证码建议使用排除易混淆字符的字符集
5. **性能考虑**: 高并发场景下注意 Redis 连接池配置

## 许可证

MIT License