# google

Google 翻译的 Go 翻译器封装，集成三个版本的 Google 翻译 API：V1（免费非官方接口）、V2 / V3（官方 Cloud Translation API）。

[English](README.md) | 简体中文

## 简介

本模块（`github.com/tx7do/go-utils/translator/google`）是 [go-utils](https://github.com/tx7do/go-utils) translator 组件集的 Google 翻译实现，是一个**独立的 Go 模块**（自带 `go.mod`）。

它提供统一的 `Translate` 入口，按配置的版本分发到 V1 / V2 / V3，也可直接调用指定版本的方法。方法签名与 go-utils 的 `Translator` 接口一致：

```go
// Translate 翻译文本
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

## Features

- **三个版本可选**：V1 免费非官方接口；V2、V3 为官方 API，需要 API Key
- **统一入口**：`Translate` 根据 `WithVersion` 配置自动分发
- **可直调指定版本**：`TranslateV1` / `TranslateV2` / `TranslateV3`
- **源语言自动检测**：V1、V3 支持传 `auto`；V2 始终自动检测并返回检测到的源语言
- **实现 `Translator` 接口形状**：`Translate(source, sourceLang, targetLang string) (string, error)`

## Installation

本模块是独立 Go 模块：

```bash
go get github.com/tx7do/go-utils/translator/google
```

```go
import "github.com/tx7do/go-utils/translator/google"
```

## Quick Start

### 基础使用（V1，免费，无需凭证）

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/google"
)

func main() {
	// NewTranslator(opts ...Option) *Translator，不返回 error
	// 不传选项时默认使用 V1
	translator := google.NewTranslator()

	// 源语言传 "auto" 表示自动检测
	result, err := translator.Translate("Hello, world!", "auto", "zh-CN")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Translation:", result)
}
```

### 使用 V2 API（官方 API，需要 API Key）

```go
translator := google.NewTranslator(
	google.WithVersion("v2"),
	google.WithApiKey("your_google_api_key"),
)

// Translate 返回 (string, error)
result, err := translator.Translate("Hello", "en", "zh-CN")
if err != nil {
	log.Fatal(err)
}
fmt.Println("Translation:", result)
```

### 使用 V3 API（官方 API，需要 API Key）

```go
translator := google.NewTranslator(
	google.WithVersion("v3"),
	google.WithApiKey("your_google_api_key"),
)

result, err := translator.Translate("Hello", "en", "zh-CN")
if err != nil {
	log.Fatal(err)
}
fmt.Println("Translation:", result)
```

## Configuration / Options

### NewTranslator

```go
func NewTranslator(opts ...Option) *Translator
```

**参数：** `opts`：可选配置项。

**返回值：** `*Translator` 实例，**不返回 error**。

### Translate

```go
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

按 `WithVersion` 设置的版本分发：

- `"v1"` 或空值（默认）→ `TranslateV1`
- `"v2"` → `TranslateV2`（丢弃检测语言，只返回译文）
- `"v3"` → `TranslateV3`

### WithVersion

```go
func WithVersion(version string) Option
```

设置使用的 API 版本：`"v1"`（默认）、`"v2"`、`"v3"`。

### WithApiKey

```go
func WithApiKey(key string) Option
```

设置 Google API Key，V2 / V3 必需；V1 不需要。

## Advanced Usage

### 直接调用指定版本

`Translate` 之外，三个版本的方法都可以直接调用：

```go
func (t *Translator) TranslateV1(source, sourceLang, targetLang string) (string, error)

func (t *Translator) TranslateV2(source, sourceLang, targetLang string) (string, language.Tag, error)

func (t *Translator) TranslateV3(source, sourceLang, targetLang string) (string, error)
```

- **TranslateV1**：请求免费的非官方接口（`translate.googleapis.com/translate_a/single`），无需认证。
- **TranslateV2**：返回**三元组** `(译文, 检测到的源语言, error)`，第二个返回值是 `golang.org/x/text/language.Tag`。`sourceLang` 参数会被**忽略**——V2 始终自动检测源语言。`targetLang` 会经 `language.Parse` 解析，须是合法的 BCP-47 语言标签（如 `zh-CN`、`es`），解析失败会返回错误。
- **TranslateV3**：调用官方 Cloud Translation API v3（`TranslateText`），固定使用 `text/plain` MIME 类型；`sourceLang` 可传 `auto`。

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/google"
)

func main() {
	translator := google.NewTranslator(google.WithApiKey("your_google_api_key"))

	// TranslateV2 返回 (译文, 检测到的源语言, error)；"en" 会被忽略
	text, detected, err := translator.TranslateV2("Hello, world!", "en", "zh-CN")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Translation:", text)
	fmt.Println("Detected source language:", detected.String())
}
```

### 版本对比

| 特性 | V1 | V2 | V3 |
|------|----|----|----|
| 是否免费 | 免费（非官方接口，无 SLA） | 否（按量计费） | 否（按量计费） |
| 需要 API Key | 否 | 是 | 是 |
| 官方支持 | 否 | 是 | 是 |
| 源语言 | 支持 `auto` | 忽略 `sourceLang`，始终自动检测 | 支持 `auto` |
| 返回值 | 译文 | 译文 + 检测到的源语言（`TranslateV2`） | 译文 |
| MIME 类型 | - | - | 固定 `text/plain` |

**版本选择建议**：

- **V1**：适合小流量、原型验证、学习演示；注意它是非官方接口，可能随时变更或被限制
- **V2**：适合需要官方支持、且要拿到检测出的源语言的项目
- **V3**：Google 当前主推的官方 API 版本

### 错误处理

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/google"
)

func main() {
	translator := google.NewTranslator()

	result, err := translator.Translate("Hello", "en", "zh-CN")
	if err != nil {
		log.Printf("translate failed: %v", err)
		return
	}

	fmt.Println(result)
}
```

**V1** 返回固定的错误文本，可用 `err.Error()` 匹配（**仅适用于 V1**）：

- `error getting translate.googleapis.com`
- `error reading response body`
- `error 400 (Bad Request)`
- `error unmarshalling data`
- `no translated data in response`

**V2 / V3** 返回官方 SDK 的错误（含 gRPC 状态等上下文），请直接按 `error` 处理，**不要**用字符串匹配。

### 与统一接口配合

本模块不 import `go-utils/translator`，但 `Translate` 方法签名与其 `Translator` 接口一致，可直接赋值使用：

```go
package main

import (
	"github.com/tx7do/go-utils/translator"
	"github.com/tx7do/go-utils/translator/google"
)

func main() {
	tr := google.NewTranslator() // 默认 V1，免费
	var iface translator.Translator = tr
	_, _ = iface.Translate("Hello", "en", "zh-CN")
}
```

## 支持的语言代码

以下为 Google 风格语言代码（与 Google 翻译一致，如希伯来语为 `iw`、中文简体为 `zh-CN`），共 109 个。源语言可用 `auto`；V2 的目标语言须能被 `language.Parse` 解析。完整列表见 [Google 翻译语言列表](https://cloud.google.com/translate/docs/languages)。

| 语言 | Code | 语言 | Code |
|------|------|------|------|
| 阿尔巴尼亚语 | `sq` | 缅甸语 | `my` |
| 阿拉伯语 | `ar` | 苗语 | `hmn` |
| 阿姆哈拉语 | `am` | 南非科萨语 | `xh` |
| 阿塞拜疆语 | `az` | 南非祖鲁语 | `zu` |
| 爱尔兰语 | `ga` | 尼泊尔语 | `ne` |
| 爱沙尼亚语 | `et` | 挪威语 | `no` |
| 奥利亚语 | `or` | 旁遮普语 | `pa` |
| 巴斯克语 | `eu` | 葡萄牙语 | `pt` |
| 白俄罗斯语 | `be` | 普什图语 | `ps` |
| 保加利亚语 | `bg` | 齐切瓦语 | `ny` |
| 冰岛语 | `is` | 日语 | `ja` |
| 波兰语 | `pl` | 瑞典语 | `sv` |
| 波斯尼亚语 | `bs` | 萨摩亚语 | `sm` |
| 波斯语 | `fa` | 塞尔维亚语 | `sr` |
| 布尔语(南非荷兰语) | `af` | 塞索托语 | `st` |
| 鞑靼语 | `tt` | 僧伽罗语 | `si` |
| 丹麦语 | `da` | 世界语 | `eo` |
| 德语 | `de` | 斯洛伐克语 | `sk` |
| 俄语 | `ru` | 斯洛文尼亚语 | `sl` |
| 法语 | `fr` | 斯瓦希里语 | `sw` |
| 菲律宾语 | `tl` | 苏格兰盖尔语 | `gd` |
| 芬兰语 | `fi` | 宿务语 | `ceb` |
| 弗里西语 | `fy` | 索马里语 | `so` |
| 高棉语 | `km` | 塔吉克语 | `tg` |
| 格鲁吉亚语 | `ka` | 泰卢固语 | `te` |
| 古吉拉特语 | `gu` | 泰米尔语 | `ta` |
| 哈萨克语 | `kk` | 泰语 | `th` |
| 海地克里奥尔语 | `ht` | 土耳其语 | `tr` |
| 韩语 | `ko` | 土库曼语 | `tk` |
| 豪萨语 | `ha` | 威尔士语 | `cy` |
| 荷兰语 | `nl` | 维吾尔语 | `ug` |
| 吉尔吉斯语 | `ky` | 乌尔都语 | `ur` |
| 加利西亚语 | `gl` | 乌克兰语 | `uk` |
| 加泰罗尼亚语 | `ca` | 乌兹别克语 | `uz` |
| 捷克语 | `cs` | 西班牙语 | `es` |
| 卡纳达语 | `kn` | 希伯来语 | `iw` |
| 科西嘉语 | `co` | 希腊语 | `el` |
| 克罗地亚语 | `hr` | 夏威夷语 | `haw` |
| 库尔德语 | `ku` | 信德语 | `sd` |
| 拉丁语 | `la` | 匈牙利语 | `hu` |
| 拉脱维亚语 | `lv` | 修纳语 | `sn` |
| 老挝语 | `lo` | 亚美尼亚语 | `hy` |
| 立陶宛语 | `lt` | 伊博语 | `ig` |
| 卢森堡语 | `lb` | 意大利语 | `it` |
| 卢旺达语 | `rw` | 意第绪语 | `yi` |
| 罗马尼亚语 | `ro` | 印地语 | `hi` |
| 马尔加什语 | `mg` | 印尼巽他语 | `su` |
| 马耳他语 | `mt` | 印尼语 | `id` |
| 马拉地语 | `mr` | 印尼爪哇语 | `jw` |
| 马拉雅拉姆语 | `ml` | 英语 | `en` |
| 马来语 | `ms` | 约鲁巴语 | `yo` |
| 马其顿语 | `mk` | 越南语 | `vi` |
| 毛利语 | `mi` | 中文（繁体） | `zh-TW` |
| 蒙古语 | `mn` | 中文（简体） | `zh-CN` |
| 孟加拉语 | `bn` |  |  |

## 获取 API Key

1. 访问 [Google Cloud Console](https://console.cloud.google.com/)
2. 创建（或选择）项目
3. 启用 Cloud Translation API
4. 创建 API Key 并复制到 `WithApiKey`

### 费用

- **V1**：免费（非官方接口，无 SLA）
- **V2 / V3**：按使用量计费，价格以 [Google Cloud 官方定价](https://cloud.google.com/translate/pricing)为准

## Notes

- **不支持 context**：`Translate` 及各版本方法均不接受 `context.Context`；V2 / V3 内部固定使用 `context.Background()` 创建客户端和调用接口，外部无法传入超时。如需超时控制，须在外部用 goroutine + `select` 自行封装。
- **客户端懒加载并缓存**：V2 / V3 的客户端在首次调用时创建并保存在 `Translator` 实例内，请复用实例，避免重复创建。
- **V1 为非官方接口**：通过网页翻译端点实现，无官方 SLA，可能随时变更、限流或失效，不建议用于关键业务。
- **V2 忽略源语言**：无论 `sourceLang` 传什么，V2 都会自动检测，并通过 `TranslateV2` 的第二个返回值告诉你检测结果。
- **V1 存在速率限制**：该接口无官方配额文档，请自行控制请求频率，并对被限流做好降级准备。
- 超长/大文本、HTML 翻译（V3 的 `text/html` MIME）等高级能力本模块未封装，需要时请直接使用官方 SDK。

## 依赖项

与 `go.mod` 一致：

```
cloud.google.com/go/translate v1.12.7
golang.org/x/text v0.34.0
google.golang.org/api v0.269.0
```

## Acknowledgements

- [Google Cloud Translation API](https://cloud.google.com/translate/docs)
- [cloud.google.com/go/translate Go SDK](https://github.com/googleapis/google-cloud-go)

## 相关链接

- [Google Cloud Console](https://console.cloud.google.com/)
- [Google 翻译语言列表](https://cloud.google.com/translate/docs/languages)
- [项目主页](https://github.com/tx7do/go-utils)

## License

遵循项目 [MIT License](https://github.com/tx7do/go-utils/blob/main/LICENSE)。
