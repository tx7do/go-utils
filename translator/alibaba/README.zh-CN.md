# alibaba

阿里云机器翻译（Alibaba Cloud Machine Translation）的 Go 翻译器封装，基于阿里云 alimt SDK 的通用翻译接口（TranslateGeneral）。

[English](README.md) | 简体中文

## 简介

本模块（`github.com/tx7do/go-utils/translator/alibaba`）是 [go-utils](https://github.com/tx7do/go-utils) translator 组件集的阿里云实现，是一个**独立的 Go 模块**（自带 `go.mod`）。

它提供的方法签名与 go-utils 的 `Translator` 接口一致：

```go
// Translate 翻译文本
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

内部调用阿里云通用版机器翻译 `TranslateGeneral` 接口，固定使用 `FormatType=text`、`Scene=general`。

## Features

- 基于官方 [alimt Go SDK](https://github.com/alibabacloud-go/alimt-20190107)（通用翻译接口）
- 阿里云标准 AccessKey 认证
- 区域可配置（`WithRegionID`），默认 `cn-hangzhou`
- 支持数十种语言的互译，常见语言代码见下方语言代码表（完整支持列表以[阿里云官方文档](https://help.aliyun.com/document/detail/209537.html)为准）
- 错误统一为普通 `error`，无自定义错误类型

## Installation

本模块是独立 Go 模块：

```bash
go get github.com/tx7do/go-utils/translator/alibaba
```

```go
import "github.com/tx7do/go-utils/translator/alibaba"
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/alibaba"
)

func main() {
	// 创建翻译器
	translator, err := alibaba.NewTranslator(
		"your_access_key_id",
		"your_access_key_secret",
	)
	if err != nil {
		log.Fatal(err)
	}

	// 翻译文本
	result, err := translator.Translate("Hello, world!", "en", "zh")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Translation:", result)
}
```

使用自定义区域：

```go
translator, err := alibaba.NewTranslator(
	"your_access_key_id",
	"your_access_key_secret",
	alibaba.WithRegionID("cn-shanghai"),
)
```

## Configuration / Options

### NewTranslator

```go
func NewTranslator(accessKeyID, accessKeySecret string, opts ...Option) (*Translator, error)
```

**参数：**

- `accessKeyID`：阿里云访问密钥 ID
- `accessKeySecret`：阿里云访问密钥 Secret
- `opts`：可选配置项

**返回值：** `*Translator` 实例和错误。客户端创建失败时返回错误。

### Translate

```go
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

**参数：**

- `source`：要翻译的文本
- `sourceLang`：源语言代码，可传 `auto` 由服务端自动检测
- `targetLang`：目标语言代码

**返回值：** 译文和错误。

### WithRegionID

```go
func WithRegionID(regionID string) Option
```

设置阿里云区域 ID。这是本模块**唯一**的配置项；不设置时默认 `cn-hangzhou`。请求端点按区域生成为 `mt.<regionID>.aliyuncs.com`。

源码注释中列出的支持区域示例：`cn-hangzhou`（默认）、`cn-shanghai`、`cn-beijing`、`ap-southeast-1` 等。可用区域以阿里云官方文档为准。

## 支持的语言代码

以下为常见语言代码参考，完整支持列表以[阿里云官方文档](https://help.aliyun.com/document/detail/209537.html)为准。

### 主流语言

| 语言 | 代码 | 语言 | 代码 |
|------|------|------|------|
| 中文（简体） | `zh` | 英语 | `en` |
| 中文（繁体） | `zh-TW` | 日语 | `ja` |
| 韩语 | `ko` | 俄语 | `ru` |
| 德语 | `de` | 法语 | `fr` |
| 西班牙语 | `es` | 意大利语 | `it` |
| 葡萄牙语 | `pt` | 荷兰语 | `nl` |
| 瑞典语 | `sv` | 挪威语 | `no` |
| 丹麦语 | `da` | 芬兰语 | `fi` |
| 波兰语 | `pl` | 捷克语 | `cs` |
| 罗马尼亚语 | `ro` | 匈牙利语 | `hu` |
| 保加利亚语 | `bg` | 克罗地亚语 | `hr` |
| 塞尔维亚语 | `sr` | 乌克兰语 | `uk` |
| 白俄罗斯语 | `be` | 希腊语 | `el` |

### 亚洲语言

| 语言 | 代码 | 语言 | 代码 |
|------|------|------|------|
| 泰语 | `th` | 越南语 | `vi` |
| 老挝语 | `lo` | 缅甸语 | `my` |
| 柬埔寨语 | `km` | 印尼语 | `id` |
| 马来语 | `ms` | 菲律宾语 | `tl` |
| 孟加拉语 | `bn` | 印地语 | `hi` |
| 乌尔都语 | `ur` | 旁遮普语 | `pa` |
| 古吉拉特语 | `gu` | 泰卢固语 | `te` |
| 泰米尔语 | `ta` | 马拉雅拉姆语 | `ml` |
| 卡纳达语 | `kn` | 马拉地语 | `mr` |

### 中东和非洲语言

| 语言 | 代码 | 语言 | 代码 |
|------|------|------|------|
| 阿拉伯语 | `ar` | 希伯来语 | `he` |
| 波斯语 | `fa` | 土耳其语 | `tr` |
| 库尔德语 | `ku` | 阿塞拜疆语 | `az` |
| 斯瓦希里语 | `sw` | 豪萨语 | `ha` |
| 约鲁巴语 | `yo` | 伊博语 | `ig` |

### 欧洲和其他语言

| 语言 | 代码 | 语言 | 代码 |
|------|------|------|------|
| 阿尔巴尼亚语 | `sq` | 巴斯克语 | `eu` |
| 加泰罗尼亚语 | `ca` | 加利西亚语 | `gl` |
| 冰岛语 | `is` | 爱尔兰语 | `ga` |
| 拉丁语 | `la` | 卢森堡语 | `lb` |
| 马耳他语 | `mt` | 威尔士语 | `cy` |
| 格鲁吉亚语 | `ka` | 亚美尼亚语 | `hy` |
| 爱沙尼亚语 | `et` | 拉脱维亚语 | `lv` |
| 立陶宛语 | `lt` | 马其顿语 | `mk` |
| 蒙古语 | `mn` | 尼泊尔语 | `ne` |

## Advanced Usage

### 错误处理

本模块**没有**专门的错误类型（如 `TranslateError`），所有错误都是 `fmt.Errorf` 包装的普通 `error`，直接按 `error` 判断即可：

```go
package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/tx7do/go-utils/translator/alibaba"
)

func main() {
	translator, err := alibaba.NewTranslator(
		os.Getenv("ALIBABA_ACCESS_KEY_ID"),
		os.Getenv("ALIBABA_ACCESS_KEY_SECRET"),
	)
	if err != nil {
		log.Fatal(err)
	}

	result, err := translator.Translate("Hello", "en", "zh")
	if err != nil {
		// 网络或 SDK 层错误包装了底层错误，可用 errors.Unwrap / errors.Is 继续解包
		if unwrapped := errors.Unwrap(err); unwrapped != nil {
			log.Printf("underlying error: %v", unwrapped)
		}
		log.Fatal(err)
	}

	fmt.Println(result)
}
```

错误消息格式（与源码一致）：

| 场景 | 错误消息格式 |
|------|--------------|
| SDK/网络错误 | `alibaba translate error: <底层错误>`（`%w` 包装，可解包） |
| 服务返回错误码 | `alibaba translate error: code=<数字>, message=<描述>` |
| 响应为空 | `alibaba translate: empty response` |
| 译文为空 | `alibaba translate: translation result is nil` |

具体的阿里云服务端错误码含义请查阅[阿里云 API 错误中心](https://api.aliyun.com/troubleshoot)。

### 结果缓存

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/translator/alibaba"
)

var cache = make(map[string]string)

func translateWithCache(translator *alibaba.Translator, text, sourceLang, targetLang string) string {
	key := fmt.Sprintf("%s_%s_%s", text, sourceLang, targetLang)
	if v, ok := cache[key]; ok {
		return v
	}

	result, _ := translator.Translate(text, sourceLang, targetLang)
	cache[key] = result
	return result
}

func main() { _ = translateWithCache }
```

### 批量翻译

```go
package main

import "github.com/tx7do/go-utils/translator/alibaba"

func translateBatch(translator *alibaba.Translator, items []string, sourceLang, targetLang string) []string {
	results := make([]string, len(items))
	for i, item := range items {
		results[i], _ = translator.Translate(item, sourceLang, targetLang)
	}
	return results
}

func main() { _ = translateBatch }
```

### 与统一接口配合

本模块不 import `go-utils/translator`，但方法签名与其 `Translator` 接口一致，可直接赋值使用：

```go
package main

import (
	"github.com/tx7do/go-utils/translator"
	"github.com/tx7do/go-utils/translator/alibaba"
)

func main() {
	tr, err := alibaba.NewTranslator("your_access_key_id", "your_access_key_secret")
	if err != nil {
		panic(err)
	}
	var iface translator.Translator = tr
	_, _ = iface.Translate("Hello", "en", "zh")
}
```

## 获取访问凭证

1. 访问 [阿里云控制台](https://console.aliyun.com/) 并登录
2. 创建 AccessKey：进入 "访问控制（RAM）" → 用户 → AccessKey 管理
3. 保存 AccessKey ID 和 AccessKey Secret
4. 开通[机器翻译服务](https://www.aliyun.com/product/ai/mtranslation)（提供免费额度，超出后按量付费，具体以[官方定价](https://www.aliyun.com/price/product)为准）

## Notes

- **不支持 context**：`Translate` 方法不接受 `context.Context`，模块也未暴露自定义 HTTP 客户端或超时配置。如需超时控制，只能在外部用 goroutine + `select` 自行封装，例如：

```go
package main

import (
	"context"
	"time"
)

type result struct {
	text string
	err  error
}

// translateWithTimeout 在外部为一次翻译调用加超时
func translateWithTimeout(t interface {
	Translate(source, sourceLang, targetLang string) (string, error)
}, text, from, to string, timeout time.Duration) (string, error) {
	ch := make(chan result, 1)
	go func() {
		text, err := t.Translate(text, from, to)
		ch <- result{text, err}
	}()

	select {
	case r := <-ch:
		return r.text, r.err
	case <-time.After(timeout):
		return "", context.DeadlineExceeded // 翻译 goroutine 仍在后台运行
	}
}

func main() { _ = translateWithTimeout }
```

> 注意：goroutine 中的请求无法被真正取消，仅能让调用方按时返回。

- 语言代码表为常见语言参考，完整列表以阿里云官方文档为准；不确定时优先传 `auto` 作为源语言
- 每次调用都会请求阿里云 API，请注意限流与配额；`Translator` 实例内部持有 SDK 客户端，可复用，不必每次重建
- 本模块固定使用通用翻译场景（`Scene=general`）与纯文本格式（`FormatType=text`）；如需电商、医疗等领域模型，需直接使用 alimt SDK

## 依赖项

与 `go.mod` 一致：

```
github.com/alibabacloud-go/alimt-20190107 v1.0.0
github.com/alibabacloud-go/darabonba-openapi/v2 v2.1.15
github.com/alibabacloud-go/tea v1.4.0
```

## Acknowledgements

- [阿里云机器翻译服务](https://www.aliyun.com/product/ai/mtranslation)
- [alibabacloud-go/alimt-20190107 Go SDK](https://github.com/alibabacloud-go/alimt-20190107)

## 相关链接

- [阿里云机器翻译文档](https://help.aliyun.com/product/2477301.html)
- [阿里云控制台](https://console.aliyun.com/)
- [项目主页](https://github.com/tx7do/go-utils)

## License

遵循项目 [MIT License](https://github.com/tx7do/go-utils/blob/main/LICENSE)。
