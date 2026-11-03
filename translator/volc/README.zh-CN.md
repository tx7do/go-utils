# volc

火山引擎翻译（Volcengine Translate）的 Go 翻译器封装，基于火山 TranslateText API（HMAC-SHA256 签名）。

[English](README.md) | 简体中文

## 简介

本模块（`github.com/tx7do/go-utils/translator/volc`）是 [go-utils](https://github.com/tx7do/go-utils) translator 组件集的火山引擎实现，是一个**独立的 Go 模块**（自带 `go.mod`），无第三方运行时依赖。

它提供单文本与批量翻译，方法签名与 go-utils 的 `Translator` 接口一致：

```go
// Translate 单文本翻译
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

内部调用火山引擎 `TranslateText` API（`Action=TranslateText&Version=2020-06-01`），自动完成 V4 风格的 HMAC-SHA256 请求签名，翻译场景固定为 `general`。

## Features

- 单文本翻译（`Translate`）与批量翻译（`TranslateBatch`）
- 火山引擎 HMAC-SHA256 签名自动处理（含请求体哈希、规范化请求、派生密钥）
- 区域可配置（`WithRegion`），默认 `cn-north-1`
- HTTP 客户端可自定义（`WithHTTPClient`）
- 无自定义错误类型，错误为普通 `error`

## Installation

本模块是独立 Go 模块：

```bash
go get github.com/tx7do/go-utils/translator/volc
```

```go
import "github.com/tx7do/go-utils/translator/volc"
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/volc"
)

func main() {
	// 创建翻译器，需要提供 AccessKey 和 SecretKey
	// NewTranslator 返回 (*Translator, error)，必须处理错误
	translator, err := volc.NewTranslator("your_access_key", "your_secret_key")
	if err != nil {
		log.Fatal(err)
	}

	// 翻译文本
	result, err := translator.Translate("Hello, world!", "en", "zh")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Translation result:", result)
}
```

## Configuration / Options

### NewTranslator

```go
func NewTranslator(accessKey, secretKey string, opts ...Option) (*Translator, error)
```

**参数：**

- `accessKey`：火山引擎 AccessKey
- `secretKey`：火山引擎 SecretKey（末尾换行符会被自动去除）
- `opts`：可选配置项

**返回值：** `*Translator` 实例和错误。`accessKey` 或 `secretKey` 为空时返回错误。

默认行为：区域 `cn-north-1`；HTTP 客户端 30 秒超时（可通过 `WithHTTPClient` 替换）。

### Translate

```go
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

**参数：**

- `source`：要翻译的文本（为空时返回错误 `待翻译文本不能为空`）
- `sourceLang`：源语言代码
- `targetLang`：目标语言代码

**返回值：** 译文和错误。

### TranslateBatch

```go
func (t *Translator) TranslateBatch(sources []string, sourceLang, targetLang string) ([]string, error)
```

批量翻译，一次请求翻译多条文本。

**参数：**

- `sources`：待翻译文本列表（传入空切片时直接返回空结果）
- `sourceLang`：源语言代码
- `targetLang`：目标语言代码

**返回值：** 译文列表和错误。若返回条数与输入不一致，返回错误（格式：`结果数量不匹配：输入%d条，返回%d条`）。

### Options

#### WithRegion

```go
func WithRegion(region string) Option
```

设置火山引擎区域。默认 `cn-north-1`（以实现为准，见下方 Notes）。区域用于签名凭证范围与请求上下文。

#### WithHTTPClient

```go
func WithHTTPClient(client *http.Client) Option
```

自定义 HTTP 客户端，可用于替换默认的 30 秒超时、配置代理或传输层。

## 获取 AccessKey 和 SecretKey

1. 访问[火山引擎控制台](https://console.volcengine.com/)
2. 在"密钥管理"（AccessKey 管理）中创建或查看 AccessKey（AK）和 SecretKey（SK）
3. 确保已开通[火山翻译服务](https://www.volcengine.com/product/translate)
4. 将凭证传入 `NewTranslator`

## 支持的语言

| 语言 | 代码 |
|------|------|
| 中文 | zh |
| 英文 | en |
| 日语 | ja |
| 韩语 | ko |
| 俄语 | ru |
| 德语 | de |
| 法语 | fr |
| 西班牙语 | es |
| 葡萄牙语 | pt |
| 意大利语 | it |
| 泰语 | th |
| 越南语 | vi |
| 印尼语 | id |
| 缅甸语 | my |
| 柬埔寨语 | km |
| 老挝语 | lo |
| 阿拉伯语 | ar |
| 土耳其语 | tr |
| 波兰语 | pl |
| 乌克兰语 | uk |

更多语言支持请参考[官方文档](https://www.volcengine.com/docs/4640/35107)。

## Advanced Usage

### 批量翻译

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/volc"
)

func main() {
	translator, err := volc.NewTranslator(
		"your_access_key",
		"your_secret_key",
	)
	if err != nil {
		log.Fatal(err)
	}

	results, err := translator.TranslateBatch(
		[]string{"火山翻译", "Go 客户端"},
		"zh", "en",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(results)
}
```

### 自定义区域与 HTTP 客户端

```go
package main

import (
	"net/http"
	"time"

	"github.com/tx7do/go-utils/translator/volc"
)

func main() {
	client := &http.Client{Timeout: 10 * time.Second}

	translator, err := volc.NewTranslator(
		"your_access_key",
		"your_secret_key",
		volc.WithRegion("cn-north-1"),
		volc.WithHTTPClient(client),
	)
	if err != nil {
		panic(err)
	}
	_ = translator
}
```

### 错误处理

本模块没有专门的错误类型，所有错误都是 `fmt.Errorf` 生成的普通 `error`，直接判断 `err` 即可：

```go
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/tx7do/go-utils/translator/volc"
)

func main() {
	translator, err := volc.NewTranslator("your_access_key", "your_secret_key")
	if err != nil {
		log.Fatal(err)
	}

	result, err := translator.Translate("Hello", "en", "zh")
	if err != nil {
		// err.Error() 形如 `API错误: InvalidAccessKey - The access key is invalid`
		// 可按需解析 "-" 前的错误 Code
		if strings.Contains(err.Error(), "API错误") {
			log.Printf("volc API error: %v", err)
		}
		return
	}

	fmt.Println(result)
}
```

错误消息格式（与源码一致，均为中文前缀）：

| 场景 | 错误消息格式 |
|------|--------------|
| API 返回错误 | `API错误: <Code> - <Message>` |
| HTTP 状态码非 200 | `HTTP错误: <状态码>, 响应: <响应体>` |
| 单文本调用失败 | `翻译失败: <底层错误>` |
| 批量调用失败 | `批量翻译失败: <底层错误>` |
| 待翻译文本为空 | `待翻译文本不能为空` |
| 翻译结果为空 | `翻译结果为空` |
| 结果数量不匹配 | `结果数量不匹配：输入%d条，返回%d条` |
| 网络请求失败 | `发送请求失败: <底层错误>` |
| 响应解析失败 | `解析响应失败: <底层错误>, 响应: <响应体>` |

### 与统一接口配合

本模块不 import `go-utils/translator`，但 `Translate` 方法签名与其 `Translator` 接口一致，可直接赋值使用：

```go
package main

import (
	"github.com/tx7do/go-utils/translator"
	"github.com/tx7do/go-utils/translator/volc"
)

func main() {
	tr, err := volc.NewTranslator("your_access_key", "your_secret_key")
	if err != nil {
		panic(err)
	}
	var iface translator.Translator = tr
	_, _ = iface.Translate("Hello", "en", "zh")
}
```

## Notes

1. **默认区域以实现为准**：默认区域是 `cn-north-1`（`translator.go` 中初始化）。注意 `options.go` 中 `WithRegion` 的注释写着 "cn-beijing（默认）"，与实现不一致——**请以实现为准**（测试也断言默认值为 `cn-north-1`）。
2. **调试日志无法关闭**：每次 API 调用都会通过 `fmt.Printf` 向标准输出打印请求 URL、`Authorization` 签名头、`X-Date` 和请求体。`Authorization` 头包含请求签名（由 AK 派生，不含明文 SK），但输出量大且可能不希望出现在生产日志中——当前版本没有任何开关可以关闭它，接入生产环境前请知晓这一点。
3. **导出的调试辅助方法**：
   - `String()`：返回脱敏的实例描述（区域 + 首尾截断的 AccessKey），实现了 `fmt.Stringer`；
   - `VerifySignature(testTimestamp, testDate, testBody string, expectedSignature string)`：用给定参数重新计算签名并与预期值比对，结果通过 `fmt.Printf` 打印，用于排查签名问题（测试中使用）。
4. **API 端点与场景固定**：端点固定为 `translate.volcengineapi.com`，`Action=TranslateText`、`Version=2020-06-01`、`Scene=general`；如需其他翻译场景或端点，需直接调用火山官方 SDK。
5. 每次调用都会请求火山 API，请注意账户配额与限流；`Translator` 实例可复用。

## Acknowledgements

- [火山引擎翻译服务](https://www.volcengine.com/product/translate)

## License

遵循项目 [MIT License](https://github.com/tx7do/go-utils/blob/main/LICENSE)。
