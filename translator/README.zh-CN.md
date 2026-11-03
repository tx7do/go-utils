# translator

go-utils 的机器翻译模块：定义统一的 `Translator` 接口，并提供多个云服务商的翻译实现。

[English](README.md) | 简体中文

## 简介

`translator` 是 [go-utils](https://github.com/tx7do/go-utils) 中的翻译组件集合。

本模块（`github.com/tx7do/go-utils/translator`）只定义统一的翻译器接口，不含任何服务商实现，也没有第三方依赖；各服务商实现位于独立子模块中，按需单独安装。接口定义见 `translator.go`：

```go
// Translator 翻译器
type Translator interface {
	// Translate 翻译
	Translate(source, sourceLang, targetLang string) (string, error)
}
```

## Features

- **统一接口**：所有服务商实现同一个 `Translate(source, sourceLang, targetLang string) (string, error)` 方法，可互相替换
- **模块解耦**：接口模块零依赖；各服务商实现是独立的 Go 模块，按需引入，不会拖入不需要的依赖
- **多服务商**：阿里云翻译、百度翻译、Google 翻译、火山翻译四种实现可选

## Installation

接口模块（独立 Go 模块）：

```bash
go get github.com/tx7do/go-utils/translator
```

```go
import "github.com/tx7do/go-utils/translator"
```

> 本仓库是多模块仓库。`translator` 及其子模块 `alibaba`、`baidu`、`google`、`volc` 均为独立 Go 模块，各自携带独立的 `go.mod`，版本独立演进。

## 子模块对比

| 服务商 | 子模块 | 是否免费 | 认证方式 | Import 路径 | 服务商文档 |
|--------|--------|----------|----------|-------------|------------|
| 阿里云机器翻译 | [alibaba](alibaba/) | 有免费额度，超出后按量付费（以官方定价为准） | AccessKey ID + AccessKey Secret | `github.com/tx7do/go-utils/translator/alibaba` | [阿里云机器翻译](https://help.aliyun.com/product/2477301.html) |
| 百度翻译 | [baidu](baidu/) | 有免费额度（以官方定价为准） | AppID + SecretKey（MD5 签名） | `github.com/tx7do/go-utils/translator/baidu` | [百度翻译开放平台](https://fanyi-api.baidu.com/) |
| Google 翻译 | [google](google/) | V1 免费（非官方接口，无 SLA）；V2/V3 收费 | V1 无需认证；V2/V3 使用 API Key | `github.com/tx7do/go-utils/translator/google` | [Cloud Translation API](https://cloud.google.com/translate/docs) |
| 火山翻译 | [volc](volc/) | 有免费额度，超出后按量付费（以官方定价为准） | AccessKey + SecretKey（HMAC-SHA256 签名） | `github.com/tx7do/go-utils/translator/volc` | [火山翻译](https://www.volcengine.com/docs/4640/35107) |

## Quick Start

以 Google 翻译的 V1 免费接口为例（无需任何凭证）：

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/google"
)

func main() {
	// NewTranslator(opts ...Option) *Translator，默认使用 V1 版本
	tr := google.NewTranslator()

	// Translate(source, sourceLang, targetLang string) (string, error)
	// 源语言传 "auto" 表示自动检测
	result, err := tr.Translate("Hello, world!", "auto", "zh-CN")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Translation:", result)
}
```

安装对应的子模块：

```bash
go get github.com/tx7do/go-utils/translator/google
```

### 与统一接口配合

各子模块并不 import 本接口包，但其 `Translate` 方法签名与 `translator.Translator` 一致。利用 Go 的结构化类型，任何服务商翻译器都可以直接赋值给 `translator.Translator`：

```go
package main

import (
	"github.com/tx7do/go-utils/translator"
	"github.com/tx7do/go-utils/translator/volc"
)

// 面向接口编程：业务代码不依赖具体服务商
func translateText(tr translator.Translator, text, from, to string) (string, error) {
	return tr.Translate(text, from, to)
}

func main() {
	tr, err := volc.NewTranslator("your-access-key", "your-secret-key")
	if err != nil {
		panic(err)
	}
	_, _ = translateText(tr, "Hello", "en", "zh")
}
```

## Advanced Usage

- **切换服务商**：业务代码只依赖 `translator.Translator` 接口时，更换服务商只需替换构造函数（各服务商构造参数不同，详见各自 README）
- **多服务商容灾**：可组合多个实现，主服务商失败时降级到备用服务商
- **语言代码**：各服务商的语言代码体系不同，请查阅对应子模块 README 的语言代码表（Google 风格语言代码表见 [google/README.zh-CN.md](google/README.zh-CN.md)）

## Notes

- 各服务商的能力、计费与限流策略不同，生产使用前请阅读对应子模块 README 的「注意事项」
- 各子模块为独立 Go 模块，安装时必须使用完整的子模块路径（如 `github.com/tx7do/go-utils/translator/alibaba`），仅安装父模块不会带入任何服务商实现
- 本 README 不再提供通用语言代码表：原表为 Google 风格代码（如 `iw`、`zh-CN`），仅适用于 Google 服务商，已移至 [google](google/README.zh-CN.md)

## Acknowledgements

- 感谢各云服务商提供的机器翻译服务：[阿里云](https://www.aliyun.com/)、[百度](https://www.baidu.com/)、[Google Cloud](https://cloud.google.com/)、[火山引擎](https://www.volcengine.com/)
- 各子模块分别基于阿里云 Go SDK、百度翻译开放平台 API、Google Cloud Translation Go SDK、火山引擎翻译 API 实现

## License

本项目基于 [MIT License](https://github.com/tx7do/go-utils/blob/main/LICENSE) 开源。
