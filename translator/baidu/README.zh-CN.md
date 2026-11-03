# baidu

百度翻译开放平台通用翻译 API 的 Go 翻译器封装。

[English](README.md) | 简体中文

## 简介

本模块（`github.com/tx7do/go-utils/translator/baidu`）是 [go-utils](https://github.com/tx7do/go-utils) translator 组件集的百度翻译实现，是一个**独立的 Go 模块**（自带 `go.mod`），无第三方依赖。

它提供的方法签名与 go-utils 的 `Translator` 接口一致：

```go
// Translate 翻译文本
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

内部调用百度翻译通用翻译 API（`https://fanyi-api.baidu.com/api/trans/vip/translate`），自动生成随机 `salt` 并完成 MD5 签名。

## Features

- 基于百度翻译开放平台通用翻译 API
- AppID + SecretKey 认证，签名自动计算（MD5）
- 内置 30 秒 HTTP 超时
- 支持源语言 `auto` 自动检测
- 无自定义错误类型，错误为普通 `error`

## Installation

本模块是独立 Go 模块：

```bash
go get github.com/tx7do/go-utils/translator/baidu
```

```go
import "github.com/tx7do/go-utils/translator/baidu"
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/baidu"
)

func main() {
	// 创建翻译器，需要提供 AppID 和 SecretKey
	// 注意：NewTranslator 不返回 error
	translator := baidu.NewTranslator("your_app_id", "your_secret_key")

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
func NewTranslator(appID, secretKey string) *Translator
```

**参数：**

- `appID`：百度翻译的 AppID
- `secretKey`：百度翻译的 SecretKey

**返回值：** `*Translator` 实例。**不返回 error**——构造过程不发起网络请求。

本模块没有任何配置选项：HTTP 客户端固定 30 秒超时，不可配置。

### Translate

```go
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

**参数：**

- `source`：要翻译的文本
- `sourceLang`：源语言代码，可传 `auto` 自动检测
- `targetLang`：目标语言代码

**返回值：** 译文和错误。

## 获取 AppID 和 SecretKey

1. 访问[百度翻译开放平台](https://fanyi-api.baidu.com/)
2. 登录或注册账号
3. 在"管理控制台"中创建新应用
4. 获取 AppID 和 SecretKey

## 支持的语言

| 语言 | Code | 语言 | Code |
|------|------|------|------|
| 自动检测 | `auto` | 斯洛伐克语 | `sk` |
| 中文 | `zh` | 克罗地亚语 | `hr` |
| 英文 | `en` | 塞尔维亚语 | `sr` |
| 日语 | `ja` | 爱沙尼亚语 | `et` |
| 韩语 | `ko` | 立陶宛语 | `lt` |
| 俄语 | `ru` | 拉脱维亚语 | `lv` |
| 德语 | `de` | 阿尔巴尼亚语 | `sq` |
| 法语 | `fr` | 马其顿语 | `mk` |
| 西班牙语 | `es` | 亚美尼亚语 | `hy` |
| 葡萄牙语 | `pt` | 格鲁吉亚语 | `ka` |
| 意大利语 | `it` | 白俄罗斯语 | `be` |
| 泰语 | `th` | 冰岛语 | `is` |
| 越南语 | `vi` | 库尔德语 | `ku` |
| 印尼语 | `id` | 吉尔吉斯语 | `ky` |
| 缅甸语 | `my` | 拉丁语 | `la` |
| 柬埔寨语（高棉语） | `km` | 卢森堡语 | `lb` |
| 老挝语 | `lo` | 梵语 | `sa` |
| 阿拉伯语 | `ar` | 威尔士语 | `cy` |
| 土耳其语 | `tr` | 尼泊尔语 | `ne` |
| 波兰语 | `pl` | 孟加拉语 | `bn` |
| 乌克兰语 | `uk` | 旁遮普语 | `pa` |
| 希腊语 | `el` | 乌尔都语 | `ur` |
| 瑞典语 | `sv` | 马来语 | `ms` |
| 丹麦语 | `da` | 泰卢固语 | `te` |
| 芬兰语 | `fi` | 泰米尔语 | `ta` |
| 荷兰语 | `nl` | 卡纳达语 | `kn` |
| 挪威语 | `no` | 马拉雅拉姆语 | `ml` |
| 捷克语 | `cs` | 古吉拉特语 | `gu` |
| 罗马尼亚语 | `ro` | 马拉地语 | `mr` |
| 匈牙利语 | `hu` | 奥利亚语 | `or` |
| 保加利亚语 | `bg` | 信德语 | `sd` |
| 斯洛文尼亚语 | `sl` | 僧伽罗语 | `si` |

更多语言支持请参考[官方文档](https://fanyi-api.baidu.com/api/trans/product/apidoc)。

## Advanced Usage

### 错误处理

本模块没有专门的错误类型，所有错误都是 `fmt.Errorf` 生成的普通 `error`，直接判断 `err` 即可：

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/baidu"
)

func main() {
	translator := baidu.NewTranslator("your_app_id", "your_secret_key")

	result, err := translator.Translate("Hello", "en", "zh")
	if err != nil {
		// err.Error() 形如 "baidu translate error: 54001 - Invalid sign"
		// 可按需解析错误码（"-" 前的数字部分）
		log.Printf("translate failed: %v", err)
		return
	}

	fmt.Println(result)
}
```

错误消息格式（与源码一致）：

| 场景 | 错误消息格式 |
|------|--------------|
| API 返回错误码 | `baidu translate error: <error_code> - <error_msg>` |
| HTTP 状态码非 200 | `baidu translate error: status <状态码>, body: <响应体>` |
| 响应中没有翻译结果 | `baidu translate: invalid response format` |
| 网络请求失败 | 直接返回底层 `*url.Error` 等错误 |

### 常见错误码

以下为常见错误码，完整列表以[官方文档](https://fanyi-api.baidu.com/api/trans/product/apidoc)为准：

| 错误码 | 含义 |
|--------|------|
| 52000 | 成功 |
| 52001 | 请求超时 |
| 52002 | 系统错误 |
| 52003 | 未授权用户 |
| 54000 | 必需参数为空 |
| 54001 | 签名错误 |
| 54002 | 访问频率受限 |
| 54003 | 无效的请求 |
| 54004 | 无此字段 |
| 54005 | 无此翻译类型 |
| 58000 | 客户端IP非法 |
| 58001 | 译文语言方向不支持 |
| 90107 | 认证未通过 |

### 与统一接口配合

本模块不 import `go-utils/translator`，但方法签名与其 `Translator` 接口一致，可直接赋值使用：

```go
package main

import (
	"github.com/tx7do/go-utils/translator"
	"github.com/tx7do/go-utils/translator/baidu"
)

func main() {
	tr := baidu.NewTranslator("your_app_id", "your_secret_key")
	var iface translator.Translator = tr
	_, _ = iface.Translate("Hello", "en", "zh")
}
```

## Notes

1. **超长文本只返回首个分段**：百度 API 对超长文本会按行切分并返回多段译文（`trans_result` 数组），本模块只取第一段的译文（`trans_result[0].dst`）。翻译长文时请自行分段后再调用。
2. **HTTP 超时固定 30 秒且不可配置**：模块内部使用 `&http.Client{Timeout: 30 * time.Second}`，未提供自定义选项。
3. **签名算法**：`sign = MD5(appID + query + salt + secretKey)`，其中 `salt` 为每次请求随机生成的 `int64`，签名过程由模块自动完成，无需手工干预。
4. 单次请求的文本长度与频率限制见[官方文档](https://fanyi-api.baidu.com/api/trans/product/apidoc)，超长文本请分段，注意 QPS 限制。
5. 每次调用都会请求百度 API，请注意账户配额与计费；`Translator` 实例可复用。

## Acknowledgements

- [百度翻译开放平台](https://fanyi-api.baidu.com/)

## License

遵循项目 [MIT License](https://github.com/tx7do/go-utils/blob/main/LICENSE)。
