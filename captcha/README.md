# captcha

English | [简体中文](README.zh-CN.md)

`captcha` is a standalone Go module (`github.com/tx7do/go-utils/captcha`) providing highly customizable CAPTCHAs with multiple driver types and rich configuration, backed by Redis for storage and verification.

## Features

- Seven CAPTCHA types: digit, string, math, Chinese, slide-jigsaw, click-text, rotate-image
- Two configuration styles: functional Options and a `Config` object
- Redis-backed storage, expiration, and one-shot verification
- Behavioral CAPTCHAs (slide/click/rotate) with built-in pixel/degree tolerance

## Installation

This package is a standalone module (it has its own `go.mod`), not a subpackage of the root module `github.com/tx7do/go-utils`:

```bash
go get github.com/tx7do/go-utils/captcha
```

```go
import "github.com/tx7do/go-utils/captcha"
```

## Quick Start

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

	// Create a captcha instance with functional options
	captchaInstance := captcha.NewCaptcha(rdb,
		captcha.WithDriverType(captcha.DriverString),
		captcha.WithExpire(10*time.Minute),
		captcha.WithKeyPrefix("myapp:captcha"),
		captcha.WithStringCount(6),
		captcha.WithStringSource("ABCDEFGHJKLMNPQRSTUVWXYZ23456789"),
	)

	// Generate: returns (id, base64 image, answer, err)
	id, b64Image, answer, err := captchaInstance.Generate()
	if err != nil {
		panic(err)
	}

	// Save the answer in Redis (key prefix + id)
	ctx := context.Background()
	if err := captchaInstance.Save(ctx, id, answer); err != nil {
		panic(err)
	}

	// Verify the user input
	userInput := "AB3dEf" // characters submitted by the user (from a front-end form)
	isValid, err := captchaInstance.Verify(ctx, id, userInput)
	if err != nil {
		panic(err)
	}
	if isValid {
		fmt.Println("verified")
	} else {
		fmt.Println("verification failed")
	}
}
```

### Using a Config Object

```go
config := captcha.DefaultConfig() // returns *Config with every field defaulted
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

## CAPTCHA Types

> The per-type examples below are fragments; they reuse the imports from Quick Start (plus `encoding/json`) and the `rdb`/`captchaInstance`/`ctx` variables.

### 1. Digit (DriverDigit)

```go
captchaInstance := captcha.NewCaptcha(rdb,
	captcha.WithDriverType(captcha.DriverDigit),
	captcha.WithDigitCount(4),
	captcha.WithDigitHeight(80),
	captcha.WithDigitWidth(240),
)
```

### 2. String (DriverString)

```go
captchaInstance := captcha.NewCaptcha(rdb,
	captcha.WithDriverType(captcha.DriverString),
	captcha.WithStringCount(6),
	captcha.WithStringSource("ABCDEFGHJKLMNPQRSTUVWXYZ23456789"),
	captcha.WithStringHeight(100),
	captcha.WithStringWidth(300),
)
```

### 3. Math (DriverMath)

```go
captchaInstance := captcha.NewCaptcha(rdb,
	captcha.WithDriverType(captcha.DriverMath),
	captcha.WithMathHeight(80),
	captcha.WithMathWidth(240),
)
// Generates something like "3+5=?" with answer "8"
```

### 4. Chinese (DriverChinese)

```go
captchaInstance := captcha.NewCaptcha(rdb,
	captcha.WithDriverType(captcha.DriverChinese),
	captcha.WithChineseCount(4),
	captcha.WithChineseLanguage("zh"),
)
```

### 5. Slide Jigsaw (DriverSlide)

The user drags a tile to the gap position.

```go
captchaInstance := captcha.NewCaptcha(rdb,
	captcha.WithDriverType(captcha.DriverSlide),
	captcha.WithSlideMasterSize(300, 220), // master image size
	captcha.WithSlideTileSize(60, 60),     // tile width (height is a reserved field)
)

id, jsonData, answer, err := captchaInstance.Generate()
if err != nil {
	panic(err)
}

// jsonData is the JSON encoding of SlideCaptchaData
var slideData captcha.SlideCaptchaData
_ = json.Unmarshal([]byte(jsonData), &slideData)
// slideData.MasterImage - master image base64 (with the gap)
// slideData.TileImage   - tile image base64
// slideData.XPosition   - gap X coordinate (see the security note below)
// slideData.ID          - built from the gap coordinates as "X_Y" (see the security note below)

// Save the answer (the answer is the gap X coordinate as a string)
ctx := context.Background()
_ = captchaInstance.Save(ctx, id, answer)

// Verify: userInput is the user's X coordinate as a string; tolerance is ±5 px
userX := 123 // X coordinate from the front end
isValid, err := captchaInstance.Verify(ctx, id, fmt.Sprintf("%d", userX))
```

### 6. Click Text (DriverClick)

The user clicks the specified characters in order.

```go
captchaInstance := captcha.NewCaptcha(rdb,
	captcha.WithDriverType(captcha.DriverClick),
	captcha.WithClickMasterSize(300, 220), // master image size
	captcha.WithClickCaptchaCount(6),      // show 6 characters on the master image
	captcha.WithClickVerifyCount(3),       // 3 characters must be clicked in order
	captcha.WithClickChars("这的是随了机文我你他字"),
)

id, jsonData, answer, err := captchaInstance.Generate()
if err != nil {
	panic(err)
}

// jsonData is the JSON encoding of ClickCaptchaData
var clickData captcha.ClickCaptchaData
_ = json.Unmarshal([]byte(jsonData), &clickData)
// clickData.MasterImage - master image base64 (randomly placed characters)
// clickData.ThumbImage  - thumb image base64 (shows the characters to click)
// clickData.Dots        - correct answer coordinates (map[int]*Dot)

// Save the answer
ctx := context.Background()
_ = captchaInstance.Save(ctx, id, answer)

// Verify: userInput is a JSON array of coordinates in click order; tolerance is ±10 px
userClicksJSON := `[{"x":100,"y":50},{"x":150,"y":80},{"x":200,"y":120}]`
isValid, err := captchaInstance.Verify(ctx, id, userClicksJSON)
```

### 7. Rotate Image (DriverRotate)

The user rotates the image to the correct angle.

```go
captchaInstance := captcha.NewCaptcha(rdb,
	captcha.WithDriverType(captcha.DriverRotate),
)

id, jsonData, answer, err := captchaInstance.Generate()
if err != nil {
	panic(err)
}

// jsonData is the JSON encoding of RotateCaptchaData
var rotateData captcha.RotateCaptchaData
_ = json.Unmarshal([]byte(jsonData), &rotateData)
// rotateData.MasterImage - master image base64 (to be rotated)
// rotateData.ThumbImage  - thumb image base64 (target orientation)
// rotateData.Angle       - correct angle

// Save the answer (the answer is the angle as a string)
ctx := context.Background()
_ = captchaInstance.Save(ctx, id, answer)

// Verify: userInput is the angle as a string; tolerance is ±5 degrees (wrapped over 360)
userAngle := "90"
isValid, err := captchaInstance.Verify(ctx, id, userAngle)
```

> Note: the rotate image dimensions are determined by the upstream `go-captcha` default resources. `WithRotateMasterSize` / `WithRotateThumbSize` / `RotateConfig` are currently **not** read (see "Reserved fields" below).

## Configuration

### Functional Options

**Common:**

| Option | Description |
|--------|-------------|
| `WithDriverType(driverType DriverType)` | Driver type (digit/string/math/chinese/slide/click/rotate) |
| `WithExpire(expire time.Duration)` | Redis storage TTL |
| `WithKeyPrefix(prefix string)` | Redis key prefix |

**Digit:** `WithDigitHeight` / `WithDigitWidth` / `WithDigitCount` / `WithDigitMaxSkew` / `WithDigitDotCount` / `WithDigitConfig(*DigitConfig)`

**String:** `WithStringHeight` / `WithStringWidth` / `WithStringCount` / `WithStringSource` / `WithStringDotCount` / `WithStringConfig(*StringConfig)`

**Math:** `WithMathHeight` / `WithMathWidth` / `WithMathDotCount` / `WithMathConfig(*MathConfig)`

**Chinese:** `WithChineseHeight` / `WithChineseWidth` / `WithChineseCount` / `WithChineseLanguage` / `WithChineseDotCount` / `WithChineseConfig(*ChineseConfig)`

**Slide:** `WithSlideMasterSize(width, height)` / `WithSlideTileSize(width, height)` / `WithSlideTileRadius` / `WithSlideJigsawRadius` / `WithSlideShadow(offsetX, offsetY, blur)` / `WithSlideConfig(*SlideConfig)`

**Click:** `WithClickMasterSize(width, height)` / `WithClickThumbSize(width, height)` / `WithClickCaptchaCount` / `WithClickVerifyCount` / `WithClickChars` / `WithClickLanguage` / `WithClickShadow(display, color, offsetX, offsetY)` / `WithClickConfig(*ClickConfig)`

**Rotate:** `WithRotateMasterSize(width, height)` / `WithRotateThumbSize(width, height)` / `WithRotateConfig(*RotateConfig)`

### Default Config Factories

`DefaultConfig()` and per-driver defaults are available as factory functions (they return pointers you can mutate and pass to `NewCaptchaWithConfig`):

```go
func DefaultConfig() *Config           // DriverDigit, Expire=5m, KeyPrefix="captcha", all sub-configs
func DefaultDigitConfig() *DigitConfig
func DefaultStringConfig() *StringConfig
func DefaultMathConfig() *MathConfig
func DefaultChineseConfig() *ChineseConfig
func DefaultSlideConfig() *SlideConfig
func DefaultClickConfig() *ClickConfig
func DefaultRotateConfig() *RotateConfig
```

### Config Structures

#### Common (Config)

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| DriverType | DriverType | Driver type | `DriverDigit` |
| Expire | time.Duration | Storage TTL | 5 minutes |
| KeyPrefix | string | Redis key prefix | `captcha` |
| DigitConfig / StringConfig / MathConfig / ChineseConfig / SlideConfig / ClickConfig / RotateConfig | *XxxConfig | Per-driver configs | their own defaults |

#### Flat CAPTCHAs (Digit/String/Math/Chinese) Common Fields

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| Height | int | Image height | 80 |
| Width | int | Image width | 240 |
| CaptchaCount | int | Character count | 4 |
| MaxSkew | float64 | Maximum skew | 0.7 |
| DotCount | int | Noise dot count | 80 |
| BgColorR/G/B | uint8 | Background RGB | 255, 255, 255 |
| FontColorR/G/B | uint8 | Font RGB | 0, 0, 0 |

Extra fields:

- `StringConfig.Source`: character source (default `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`)
- `ChineseConfig.Language`: language (default `zh`)

> Fields actually consumed per driver (determined by the upstream `base64Captcha` driver constructors):
>
> - Digit: `Height/Width/CaptchaCount/MaxSkew/DotCount`
> - String: `Height/Width/CaptchaCount/DotCount/Source`
> - Math: `Height/Width/DotCount` (`CaptchaCount/MaxSkew` do not affect the generated arithmetic question)
> - Chinese: `Height/Width/CaptchaCount/DotCount/Language`
>
> `BgColor*/FontColor*` for all flat drivers, and `MaxSkew` for String/Math/Chinese, are not passed to the driver constructors, so setting them has no effect.

#### Slide (SlideConfig)

| Field | Type | Description | Default | Currently effective |
|-------|------|-------------|---------|---------------------|
| MasterWidth | int | Master image width | 300 | Yes |
| MasterHeight | int | Master image height | 220 | Yes |
| TileWidth | int | Tile width | 60 | Yes |
| TileHeight | int | Tile height | 60 | **Reserved, not effective yet** |
| TileRadius | int | Tile corner radius | 5 | **Reserved, not effective yet** |
| JigsawRadius | int | Jigsaw gap radius | 10 | **Reserved, not effective yet** |
| ShadowOffsetX | int | Shadow X offset | 5 | **Reserved, not effective yet** |
| ShadowOffsetY | int | Shadow Y offset | 5 | **Reserved, not effective yet** |
| ShadowBlur | int | Shadow blur | 10 | **Reserved, not effective yet** |

#### Click (ClickConfig)

| Field | Type | Description | Default | Currently effective |
|-------|------|-------------|---------|---------------------|
| MasterWidth | int | Master image width | 300 | Yes |
| MasterHeight | int | Master image height | 220 | Yes |
| ThumbWidth | int | Thumb image width | 150 | **Reserved, not effective yet** |
| ThumbHeight | int | Thumb image height | 40 | **Reserved, not effective yet** |
| CaptchaCount | int | Characters on the master image | 6 | Yes |
| VerifyCount | int | Characters to click in order | 3 | Yes |
| DisplayShadow | bool | Show shadow | true | Yes |
| ShadowColor | string | Shadow color | `#000000` | Yes |
| ShadowOffsetX | int | Shadow X offset | 2 | Yes |
| ShadowOffsetY | int | Shadow Y offset | 2 | Yes |
| Chars | string | Character set | 这的是随了机文我你他字在有不么中 | Yes |
| Language | string | Language | zh | **Reserved, not effective yet** |

#### Rotate (RotateConfig)

| Field | Type | Description | Default | Currently effective |
|-------|------|-------------|---------|---------------------|
| MasterWidth | int | Master image width | 300 | **Reserved, not effective yet** |
| MasterHeight | int | Master image height | 300 | **Reserved, not effective yet** |
| ThumbWidth | int | Thumb image width | 150 | **Reserved, not effective yet** |
| ThumbHeight | int | Thumb image height | 150 | **Reserved, not effective yet** |

> **About reserved fields**: the slide builder reads only `MasterWidth/MasterHeight/TileWidth`; the click builder reads only `master size/CaptchaCount/VerifyCount/the three shadow fields/Chars`; the rotate CAPTCHA is always built with a plain `rotate.NewBuilder()` and reads no `RotateConfig` fields. These option functions can be called without error, but they do not affect generation.

## API Reference

### Constructors

```go
func NewCaptcha(rdb *redis.Client, opts ...Option) *Captcha
func NewCaptchaWithConfig(rdb *redis.Client, config *Config) *Captcha  // nil config falls back to DefaultConfig()
```

### Data Structures

```go
type SlideCaptchaData struct {
	ID          string `json:"id"`           // built from the gap coordinates: "X_Y"
	MasterImage string `json:"master_image"` // master image base64
	TileImage   string `json:"tile_image"`   // tile image base64
	XPosition   int    `json:"x_position"`   // gap X coordinate (the correct answer)
}

type ClickCaptchaData struct {
	ID          string       `json:"id"`           // "click_<UnixNano>"
	MasterImage string       `json:"master_image"` // master image base64
	ThumbImage  string       `json:"thumb_image"`  // thumb image base64
	Dots        map[int]*Dot `json:"dots"`         // click dots (coordinates/size/angle)
}

type RotateCaptchaData struct {
	ID          string `json:"id"`           // "rotate_<UnixNano>"
	MasterImage string `json:"master_image"` // master image base64
	ThumbImage  string `json:"thumb_image"`  // thumb image base64
	Angle       int    `json:"angle"`        // correct angle
}
```

### Core Methods

```go
func (c *Captcha) Generate() (id, b64s, answer string, err error)
func (c *Captcha) Save(ctx context.Context, captchaID, answer string) error
func (c *Captcha) Verify(ctx context.Context, captchaID, userInput string) (bool, error)
```

- `Generate`: digit/string/math/chinese return a base64 image; for slide/click/rotate `b64s` is the JSON string of the corresponding `XxxCaptchaData`
- `Save`: writes to Redis with key `<KeyPrefix>:<captchaID>` and TTL `Expire`
- `Verify`: a successful verification deletes the captcha; a missing/expired key returns `(false, nil)`
  - Slide: X coordinate within ±5 px
  - Click: JSON coordinate array compared point by point in click order, ±10 px per axis, count must match
  - Rotate: angle within ±5 degrees (circular difference over 360)
  - Others: exact string match

### Extended Methods

```go
func (c *Captcha) VerifyAndDelete(ctx context.Context, captchaID, userInput string) (bool, error) // same as Verify
func (c *Captcha) VerifyWithoutDelete(ctx context.Context, captchaID, userInput string) (bool, error)
func (c *Captcha) Delete(ctx context.Context, captchaID string) error
func (c *Captcha) Exists(ctx context.Context, captchaID string) (bool, error)
func (c *Captcha) GetRemainingTime(ctx context.Context, captchaID string) (time.Duration, error)
func (c *Captcha) GetConfig() *Config
func (c *Captcha) SetConfig(config *Config) // ignored when config is nil
```

## Best Practices

### Exclude Confusable Characters

```go
config.StringConfig.Source = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // exclude I, l, 1, O, 0, etc.
```

### Adjust Difficulty

```go
// Harder
config.DigitConfig.MaxSkew = 1.0    // more skew
config.DigitConfig.DotCount = 150   // more noise dots
config.DigitConfig.CaptchaCount = 6 // more digits

// Easier
config.DigitConfig.MaxSkew = 0.3
config.DigitConfig.DotCount = 30
config.DigitConfig.CaptchaCount = 3
```

## Notes

1. **Redis dependency**: make sure Redis is reachable; tune the connection pool for high concurrency
2. **One-shot verification**: `Verify` deletes the captcha on success to prevent replay; use `VerifyWithoutDelete` when multiple checks are needed
3. **Slide IDs carry the answer**: `SlideCaptchaData.ID` is built from the gap coordinates (`"X_Y"`), and the JSON also contains `XPosition` — both are sent to the client, effectively leaking the answer. The current implementation suits internal/demo use; for production, strip `XPosition` and the coordinate-based ID before responding, or generate your own opaque ID with a server-side mapping
4. **Built-in background**: slide/click CAPTCHAs use a built-in 300x220 gradient background; the public API does not yet support replacing background images or fonts. Keep the default 300x220 master size for best results
5. **Reserved fields**: see the "not effective yet" marks in the config tables above; tuning them has no effect
6. **TTL**: the default is 5 minutes; avoid keeping captchas valid for too long
7. **Character source**: exclude confusable characters for string captchas

## License

MIT
