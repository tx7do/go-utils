# captcha

[English](README.md) | 简体中文

`captcha` 是一个独立 Go 模块（`github.com/tx7do/go-utils/captcha`），提供高度可定制的验证码能力，支持多种驱动类型与丰富的配置选项，基于 Redis 存储与校验。

## 功能特性

- 七种验证码类型：数字、字符串、算术、中文、滑动拼图、点击文字、旋转图片
- 两种配置方式：Options 函数模式与 Config 配置对象模式
- 基于 Redis 的存储、过期与一次性校验
- 行为式验证码（滑动/点击/旋转）自带像素/角度容差

## 安装

本包是独立模块（有自己的 `go.mod`），不是根模块 `github.com/tx7do/go-utils` 的子包：

```bash
go get github.com/tx7do/go-utils/captcha
```

```go
import "github.com/tx7do/go-utils/captcha"
```

## 快速开始

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tx7do/go-utils/captcha"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Options 模式创建验证码实例
	captchaInstance := captcha.NewCaptcha(rdb,
		captcha.WithDriverType(captcha.DriverString),
		captcha.WithExpire(10*time.Minute),
		captcha.WithKeyPrefix("myapp:captcha"),
		captcha.WithStringCount(6),
		captcha.WithStringSource("ABCDEFGHJKLMNPQRSTUVWXYZ23456789"),
	)

	// 生成验证码：返回 (id, base64图片, 答案, err)
	id, b64Image, answer, err := captchaInstance.Generate()
	if err != nil {
		panic(err)
	}

	// 将答案保存到 Redis（key 前缀 + id）
	ctx := context.Background()
	if err := captchaInstance.Save(ctx, id, answer); err != nil {
		panic(err)
	}

	// 校验用户输入
	userInput := "AB3dEf" // 用户提交的验证码字符（来自前端表单）
	isValid, err := captchaInstance.Verify(ctx, id, userInput)
	if err != nil {
		panic(err)
	}
	if isValid {
		fmt.Println("验证成功")
	} else {
		fmt.Println("验证失败")
	}
}
```

### 使用配置对象

```go
config := captcha.DefaultConfig() // 返回 *Config，全部字段带默认值
config.DriverType = captcha.DriverString
config.Expire = 10 * time.Minute
config.KeyPrefix = "myapp:captcha"

config.StringConfig = &captcha.StringConfig{
	Height:       100,
	Width:        300,
	CaptchaCount: 6,
	DotCount:     100,
	Source:       "ABCDEFGHJKLMNPQRSTUVWXYZ23456789",
}

captchaInstance := captcha.NewCaptchaWithConfig(rdb, config)
```

## 各类型验证码

> 以下各类型示例为片段，默认沿用"快速开始"中的 import（另需 `encoding/json`）与 `rdb`/`captchaInstance`/`ctx` 等变量。

### 1. 数字验证码（DriverDigit）

```go
captchaInstance := captcha.NewCaptcha(rdb,
	captcha.WithDriverType(captcha.DriverDigit),
	captcha.WithDigitCount(4),
	captcha.WithDigitHeight(80),
	captcha.WithDigitWidth(240),
)
```

### 2. 字符串验证码（DriverString）

```go
captchaInstance := captcha.NewCaptcha(rdb,
	captcha.WithDriverType(captcha.DriverString),
	captcha.WithStringCount(6),
	captcha.WithStringSource("ABCDEFGHJKLMNPQRSTUVWXYZ23456789"),
	captcha.WithStringHeight(100),
	captcha.WithStringWidth(300),
)
```

### 3. 算术验证码（DriverMath）

```go
captchaInstance := captcha.NewCaptcha(rdb,
	captcha.WithDriverType(captcha.DriverMath),
	captcha.WithMathHeight(80),
	captcha.WithMathWidth(240),
)
// 生成的验证码类似 "3+5=?"，答案为 "8"
```

### 4. 中文验证码（DriverChinese）

```go
captchaInstance := captcha.NewCaptcha(rdb,
	captcha.WithDriverType(captcha.DriverChinese),
	captcha.WithChineseCount(4),
	captcha.WithChineseLanguage("zh"),
)
```

### 5. 滑动拼图验证码（DriverSlide）

用户需要把滑块拖动到缺口位置。

```go
captchaInstance := captcha.NewCaptcha(rdb,
	captcha.WithDriverType(captcha.DriverSlide),
	captcha.WithSlideMasterSize(300, 220), // 主图尺寸
	captcha.WithSlideTileSize(60, 60),     // 滑块宽度（高度为预留字段）
)

id, jsonData, answer, err := captchaInstance.Generate()
if err != nil {
	panic(err)
}

// jsonData 是 SlideCaptchaData 的 JSON 字符串
var slideData captcha.SlideCaptchaData
_ = json.Unmarshal([]byte(jsonData), &slideData)
// slideData.MasterImage - 主图 base64（带缺口）
// slideData.TileImage   - 滑块图 base64
// slideData.XPosition   - 缺口 X 坐标（见下方安全提示）
// slideData.ID          - 由缺口坐标拼成的 "X_Y" 字符串（见下方安全提示）

// 保存答案（answer 即缺口 X 坐标字符串）
ctx := context.Background()
_ = captchaInstance.Save(ctx, id, answer)

// 校验：userInput 为用户滑动的 X 坐标字符串，允许 ±5 像素误差
userX := 123 // 前端传来的 X 坐标
isValid, err := captchaInstance.Verify(ctx, id, fmt.Sprintf("%d", userX))
```

### 6. 点击文字验证码（DriverClick）

用户需要按顺序点击主图中指定的字符。

```go
captchaInstance := captcha.NewCaptcha(rdb,
	captcha.WithDriverType(captcha.DriverClick),
	captcha.WithClickMasterSize(300, 220), // 主图尺寸
	captcha.WithClickCaptchaCount(6),      // 主图显示 6 个字符
	captcha.WithClickVerifyCount(3),       // 需要按顺序点击 3 个字符
	captcha.WithClickChars("这的是随了机文我你他字"),
)

id, jsonData, answer, err := captchaInstance.Generate()
if err != nil {
	panic(err)
}

// jsonData 是 ClickCaptchaData 的 JSON 字符串
var clickData captcha.ClickCaptchaData
_ = json.Unmarshal([]byte(jsonData), &clickData)
// clickData.MasterImage - 主图 base64（随机分布的字符）
// clickData.ThumbImage  - 缩略图 base64（提示要点击的字符）
// clickData.Dots        - 正确答案坐标（map[int]*Dot）

// 保存答案
ctx := context.Background()
_ = captchaInstance.Save(ctx, id, answer)

// 校验：userInput 为 JSON 坐标数组，按点击顺序排列，允许 ±10 像素误差
userClicksJSON := `[{"x":100,"y":50},{"x":150,"y":80},{"x":200,"y":120}]`
isValid, err := captchaInstance.Verify(ctx, id, userClicksJSON)
```

### 7. 旋转图片验证码（DriverRotate）

用户需要把图片旋转到正确角度。

```go
captchaInstance := captcha.NewCaptcha(rdb,
	captcha.WithDriverType(captcha.DriverRotate),
)

id, jsonData, answer, err := captchaInstance.Generate()
if err != nil {
	panic(err)
}

// jsonData 是 RotateCaptchaData 的 JSON 字符串
var rotateData captcha.RotateCaptchaData
_ = json.Unmarshal([]byte(jsonData), &rotateData)
// rotateData.MasterImage - 主图 base64（需要旋转）
// rotateData.ThumbImage  - 缩略图 base64（目标方向）
// rotateData.Angle       - 正确角度

// 保存答案（answer 即正确角度字符串）
ctx := context.Background()
_ = captchaInstance.Save(ctx, id, answer)

// 校验：userInput 为角度字符串，允许 ±5 度误差（按 360 度环形计算）
userAngle := "90"
isValid, err := captchaInstance.Verify(ctx, id, userAngle)
```

> 注意：旋转验证码的图片尺寸由上游 `go-captcha` 的默认资源决定，`WithRotateMasterSize` / `WithRotateThumbSize` / `RotateConfig` 当前**不会**被读取（见下文"预留字段"）。

## 配置选项

### Options 函数

**通用：**

| 选项 | 说明 |
|------|------|
| `WithDriverType(driverType DriverType)` | 驱动类型（digit/string/math/chinese/slide/click/rotate） |
| `WithExpire(expire time.Duration)` | Redis 存储过期时间 |
| `WithKeyPrefix(prefix string)` | Redis key 前缀 |

**数字：** `WithDigitHeight` / `WithDigitWidth` / `WithDigitCount` / `WithDigitMaxSkew` / `WithDigitDotCount` / `WithDigitConfig(*DigitConfig)`

**字符串：** `WithStringHeight` / `WithStringWidth` / `WithStringCount` / `WithStringSource` / `WithStringDotCount` / `WithStringConfig(*StringConfig)`

**算术：** `WithMathHeight` / `WithMathWidth` / `WithMathDotCount` / `WithMathConfig(*MathConfig)`

**中文：** `WithChineseHeight` / `WithChineseWidth` / `WithChineseCount` / `WithChineseLanguage` / `WithChineseDotCount` / `WithChineseConfig(*ChineseConfig)`

**滑动拼图：** `WithSlideMasterSize(width, height)` / `WithSlideTileSize(width, height)` / `WithSlideTileRadius` / `WithSlideJigsawRadius` / `WithSlideShadow(offsetX, offsetY, blur)` / `WithSlideConfig(*SlideConfig)`

**点击文字：** `WithClickMasterSize(width, height)` / `WithClickThumbSize(width, height)` / `WithClickCaptchaCount` / `WithClickVerifyCount` / `WithClickChars` / `WithClickLanguage` / `WithClickShadow(display, color, offsetX, offsetY)` / `WithClickConfig(*ClickConfig)`

**旋转：** `WithRotateMasterSize(width, height)` / `WithRotateThumbSize(width, height)` / `WithRotateConfig(*RotateConfig)`

### 默认配置工厂函数

`DefaultConfig()` 及各驱动的默认配置可通过工厂函数获取（返回指针，可直接修改字段后传入 `NewCaptchaWithConfig`）：

```go
func DefaultConfig() *Config           // DriverDigit, Expire=5min, KeyPrefix="captcha"，含全部子配置
func DefaultDigitConfig() *DigitConfig
func DefaultStringConfig() *StringConfig
func DefaultMathConfig() *MathConfig
func DefaultChineseConfig() *ChineseConfig
func DefaultSlideConfig() *SlideConfig
func DefaultClickConfig() *ClickConfig
func DefaultRotateConfig() *RotateConfig
```

### 配置结构

#### 通用配置（Config）

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| DriverType | DriverType | 驱动类型 | `DriverDigit` |
| Expire | time.Duration | 过期时间 | 5 分钟 |
| KeyPrefix | string | Redis key 前缀 | `captcha` |
| DigitConfig / StringConfig / MathConfig / ChineseConfig / SlideConfig / ClickConfig / RotateConfig | *XxxConfig | 各驱动配置 | 各自默认值 |

#### 平面验证码（Digit/String/Math/Chinese）公共字段

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| Height | int | 图片高度 | 80 |
| Width | int | 图片宽度 | 240 |
| CaptchaCount | int | 字符数量 | 4 |
| MaxSkew | float64 | 最大倾斜度 | 0.7 |
| DotCount | int | 干扰点数量 | 80 |
| BgColorR/G/B | uint8 | 背景色 RGB | 255, 255, 255 |
| FontColorR/G/B | uint8 | 字体色 RGB | 0, 0, 0 |

额外字段：

- `StringConfig.Source`：字符源（默认 `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`）
- `ChineseConfig.Language`：语言（默认 `zh`）

> 各驱动的实际生效字段（由上游 `base64Captcha` 的驱动构造参数决定）：
>
> - Digit：`Height/Width/CaptchaCount/MaxSkew/DotCount`
> - String：`Height/Width/CaptchaCount/DotCount/Source`
> - Math：`Height/Width/DotCount`（`CaptchaCount/MaxSkew` 不参与算术题生成）
> - Chinese：`Height/Width/CaptchaCount/DotCount/Language`
>
> 所有平面驱动的 `BgColor*/FontColor*` 与 String/Math/Chinese 的 `MaxSkew` 在构造驱动时均未传入，设置后无效果。

#### 滑动拼图（SlideConfig）

| 字段 | 类型 | 说明 | 默认值 | 当前是否生效 |
|------|------|------|--------|--------------|
| MasterWidth | int | 主图宽度 | 300 | 是 |
| MasterHeight | int | 主图高度 | 220 | 是 |
| TileWidth | int | 滑块宽度 | 60 | 是 |
| TileHeight | int | 滑块高度 | 60 | **预留字段，当前未生效** |
| TileRadius | int | 滑块圆角半径 | 5 | **预留字段，当前未生效** |
| JigsawRadius | int | 拼图缺口圆角半径 | 10 | **预留字段，当前未生效** |
| ShadowOffsetX | int | 阴影 X 偏移 | 5 | **预留字段，当前未生效** |
| ShadowOffsetY | int | 阴影 Y 偏移 | 5 | **预留字段，当前未生效** |
| ShadowBlur | int | 阴影模糊度 | 10 | **预留字段，当前未生效** |

#### 点击文字（ClickConfig）

| 字段 | 类型 | 说明 | 默认值 | 当前是否生效 |
|------|------|------|--------|--------------|
| MasterWidth | int | 主图宽度 | 300 | 是 |
| MasterHeight | int | 主图高度 | 220 | 是 |
| ThumbWidth | int | 缩略图宽度 | 150 | **预留字段，当前未生效** |
| ThumbHeight | int | 缩略图高度 | 40 | **预留字段，当前未生效** |
| CaptchaCount | int | 主图字符数量 | 6 | 是 |
| VerifyCount | int | 需按顺序点击的字符数量 | 3 | 是 |
| DisplayShadow | bool | 是否显示阴影 | true | 是 |
| ShadowColor | string | 阴影颜色 | `#000000` | 是 |
| ShadowOffsetX | int | 阴影 X 偏移 | 2 | 是 |
| ShadowOffsetY | int | 阴影 Y 偏移 | 2 | 是 |
| Chars | string | 字符集 | 这的是随了机文我你他字在有不么中 | 是 |
| Language | string | 语言类型 | zh | **预留字段，当前未生效** |

#### 旋转（RotateConfig）

| 字段 | 类型 | 说明 | 默认值 | 当前是否生效 |
|------|------|------|--------|--------------|
| MasterWidth | int | 主图宽度 | 300 | **预留字段，当前未生效** |
| MasterHeight | int | 主图高度 | 300 | **预留字段，当前未生效** |
| ThumbWidth | int | 缩略图宽度 | 150 | **预留字段，当前未生效** |
| ThumbHeight | int | 缩略图高度 | 150 | **预留字段，当前未生效** |

> **预留字段说明**：滑动验证码构建时只读取 `MasterWidth/MasterHeight/TileWidth`；点击验证码只读取 `Master 尺寸/CaptchaCount/VerifyCount/阴影三项/Chars`；旋转验证码固定使用 `rotate.NewBuilder()` 默认构建，不读取任何 `RotateConfig` 字段。这些选项函数可以调用、不会报错，但对生成结果没有影响。

## API 参考

### 构造函数

```go
func NewCaptcha(rdb *redis.Client, opts ...Option) *Captcha
func NewCaptchaWithConfig(rdb *redis.Client, config *Config) *Captcha  // config 为 nil 时使用 DefaultConfig()
```

### 数据结构

```go
type SlideCaptchaData struct {
	ID          string `json:"id"`           // 由缺口坐标拼成："X_Y"
	MasterImage string `json:"master_image"` // 主图 base64
	TileImage   string `json:"tile_image"`   // 滑块图 base64
	XPosition   int    `json:"x_position"`   // 缺口 X 坐标（正确答案）
}

type ClickCaptchaData struct {
	ID          string       `json:"id"`           // "click_<UnixNano>"
	MasterImage string       `json:"master_image"` // 主图 base64
	ThumbImage  string       `json:"thumb_image"`  // 缩略图 base64
	Dots        map[int]*Dot `json:"dots"`         // 点击点（坐标/宽高/角度）
}

type RotateCaptchaData struct {
	ID          string `json:"id"`           // "rotate_<UnixNano>"
	MasterImage string `json:"master_image"` // 主图 base64
	ThumbImage  string `json:"thumb_image"`  // 缩略图 base64
	Angle       int    `json:"angle"`        // 正确角度
}
```

### 核心方法

```go
func (c *Captcha) Generate() (id, b64s, answer string, err error)
func (c *Captcha) Save(ctx context.Context, captchaID, answer string) error
func (c *Captcha) Verify(ctx context.Context, captchaID, userInput string) (bool, error)
```

- `Generate`：digit/string/math/chinese 返回 base64 图片；slide/click/rotate 的 `b64s` 为对应 `XxxCaptchaData` 的 JSON 字符串
- `Save`：以 `<KeyPrefix>:<captchaID>` 为 key、`Expire` 为 TTL 写入 Redis
- `Verify`：校验通过后自动删除验证码；key 不存在/已过期返回 `(false, nil)`
  - 滑动：X 坐标 ±5 像素
  - 点击：JSON 坐标数组按点击顺序逐点比较，每点 X/Y ±10 像素，数量必须一致
  - 旋转：角度 ±5 度（按 360 度环形差值计算）
  - 其余类型：字符串精确匹配

### 扩展方法

```go
func (c *Captcha) VerifyAndDelete(ctx context.Context, captchaID, userInput string) (bool, error) // 等价于 Verify
func (c *Captcha) VerifyWithoutDelete(ctx context.Context, captchaID, userInput string) (bool, error)
func (c *Captcha) Delete(ctx context.Context, captchaID string) error
func (c *Captcha) Exists(ctx context.Context, captchaID string) (bool, error)
func (c *Captcha) GetRemainingTime(ctx context.Context, captchaID string) (time.Duration, error)
func (c *Captcha) GetConfig() *Config
func (c *Captcha) SetConfig(config *Config) // config 为 nil 时忽略
```

## 最佳实践

### 排除易混淆字符

```go
config.StringConfig.Source = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 排除 I、l、1、O、0 等
```

### 调整难度

```go
// 更难
config.DigitConfig.MaxSkew = 1.0    // 更大倾斜
config.DigitConfig.DotCount = 150   // 更多干扰点
config.DigitConfig.CaptchaCount = 6 // 更多位数

// 更简单
config.DigitConfig.MaxSkew = 0.3
config.DigitConfig.DotCount = 30
config.DigitConfig.CaptchaCount = 3
```

## 注意事项

1. **Redis 依赖**：使用前确保 Redis 可用；高并发场景注意连接池配置
2. **一次性使用**：`Verify` 成功后自动删除，防止重放攻击；需要多次校验用 `VerifyWithoutDelete`
3. **滑动验证码的 ID 携带答案**：`SlideCaptchaData.ID` 由缺口坐标拼成（`"X_Y"`），且 JSON 中包含 `XPosition`，两者都会下发给前端，等于把答案暴露给了客户端。当前实现仅适用于内部/演示场景；生产环境应在下发前剥离 `XPosition` 与坐标型 ID，或自行生成不透明 ID 存储映射
4. **内置背景图**：滑动/点击验证码使用内置的 300x220 渐变色背景，公开 API 暂不支持替换背景图与字体资源；主图尺寸保持默认 300x220 效果最佳
5. **预留字段**：见上文各配置表中的"当前未生效"标注，避免无效调参
6. **合理设置过期时间**：默认 5 分钟，避免验证码长期有效
7. **字符源**：字符串验证码建议排除易混淆字符

## 许可证

MIT
