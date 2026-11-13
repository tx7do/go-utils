# baidu

百度翻译开放平台通用翻译 API 的 Go 翻译器封装。

[English](README.md) | 简体中文

## 简介

本模块（`github.com/tx7do/go-utils/translator/baidu`）是 [go-utils](https://github.com/tx7do/go-utils) 翻译器组件的百度翻译实现。它是一个**独立的 Go 模块**（自带 `go.mod`），无第三方运行时依赖（testify 仅测试使用）。

它提供的方法签名与 go-utils 的 `Translator` 接口一致：

```go
// Translate 翻译文本
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

内部调用百度通用翻译 API（`https://fanyi-api.baidu.com/api/trans/vip/translate`），自动生成随机 `salt` 并计算 MD5 签名。

## 特性

- 基于百度翻译开放平台通用翻译 API
- AppID + SecretKey 鉴权，签名（MD5）自动计算
- **BCP47 语种归一化**：可直接传 `zh-CN` / `en-US` / `ar` ……，`NormalizeLang` 自动映射为百度语种码（`zh` / `en` / `ara` ……）；未收录的语种去掉地区后缀后小写透传
- **多段结果合并**：百度 API 按行拆分文本返回多段结果，本模块用 `\n` 合并后整体返回
- 内置 30 秒请求超时、10 MiB 响应体积上限
- 空输入直接短路返回 `("", nil)`；凭据缺失快速失败并给出明确错误
- 端点可配置（`WithEndpoint`）、HTTP 客户端可替换（`WithHTTPClient`）——端点选项便于测试时注入 mock 服务
- 无自定义错误类型，错误均为普通 `error`

## 安装

本模块是独立的 Go 模块：

```bash
go get github.com/tx7do/go-utils/translator/baidu
```

```go
import "github.com/tx7do/go-utils/translator/baidu"
```

## 快速开始

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/translator/baidu"
)

func main() {
	// 创建翻译器；AppID 与 SecretKey 必填。
	// 注意：NewTranslator 不返回 error。
	translator := baidu.NewTranslator("your_app_id", "your_secret_key")

	// 翻译文本；语言标签可用 BCP47（"en-US"）或百度语种码（"en"）
	result, err := translator.Translate("Hello, world!", "en-US", "zh-CN")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("翻译结果:", result)
}
```

## 配置 / 选项

### NewTranslator

```go
func NewTranslator(appID, secretKey string, opts ...Option) *Translator
```

**参数：**

- `appID`：百度翻译 AppID
- `secretKey`：百度翻译 SecretKey
- `opts`：可选配置（见下表）

**返回：** `*Translator` 实例。**不返回 error**——构造不发任何网络请求，凭据是否配置延迟到每次 `Translate` 调用时检查。

### 选项

| 选项 | 签名 | 说明 |
|------|------|------|
| `WithEndpoint` | `WithEndpoint(endpoint string)` | 覆盖 API 端点；测试注入 mock 服务用 |
| `WithHTTPClient` | `WithHTTPClient(c *http.Client)` | 替换 HTTP 客户端（自定义代理/超时策略）；默认为 30 秒超时的客户端 |

### NormalizeLang

```go
func NormalizeLang(lang string) string
```

将 BCP47 语言标签映射为百度语种码（`zh-CN` → `zh`、`zh-TW` → `cht`、`ar` → `ara`……）。未收录的标签去掉地区后缀后小写透传（`xxx-YY` → `xxx`）。`Translate` 会自动对两个语言参数做归一化；仅当别处也需要该映射（如用于展示）时才需自行调用。

### Translate

```go
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

**参数：**

- `source`：待翻译文本；纯空白输入不请求 API，直接返回 `("", nil)`
- `sourceLang`：源语言；传 `auto`、空串或无法解析出具体语种的 BCP47 标签时由百度自动检测
- `targetLang`：目标语言

**返回：** 译文与 error。

## 获取 AppID 和 SecretKey

1. 访问[百度翻译开放平台](https://fanyi-api.baidu.com/)
2. 登录或注册账号
3. 在管理控制台创建应用
4. 获取 AppID 和 SecretKey

## 支持的语种

输入可直接使用 BCP47 标签（自动归一化）；下表列出对应的百度语种码。

| 语种 | 语种码 | 语种 | 语种码 |
|------|--------|------|--------|
| 自动检测 | `auto` | 斯洛伐克语 | `sk` |
| 中文 | `zh` | 克罗地亚语 | `hr` |
| 英语 | `en` | 塞尔维亚语 | `sr` |
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
| 高棉语 | `km` | 卢森堡语 | `lb` |
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
| 匈牙利语 | `hu` | 奥里亚语 | `or` |
| 保加利亚语 | `bg` | 信德语 | `sd` |
| 斯洛文尼亚语 | `sl` | 僧伽罗语 | `si` |

繁体中文（`zh-TW` / `zh-HK`）映射为 `cht`。更多语种见[官方文档](https://fanyi-api.baidu.com/api/trans/product/apidoc)。

## 高级用法

### 错误处理

本模块无自定义错误类型，所有错误均为 `fmt.Errorf` 创建的普通 `error`，直接判 `err` 即可：

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
		// err.Error() 形如 "baidu translator: api error 54001: Invalid sign"；
		// 需要时解析 "api error" 之后的数字码
		log.Printf("翻译失败: %v", err)
		return
	}

	fmt.Println(result)
}
```

错误信息格式（与源码一致）：

| 场景 | 错误信息格式 |
|------|--------------|
| 凭据未配置 | `baidu translator: appid/secret is not configured` |
| API 返回错误码 | `baidu translator: api error <error_code>: <error_msg>` |
| HTTP 状态码非 200 | `baidu translator: unexpected status <状态码>: <响应体>` |
| 响应解码失败 | `baidu translator: decode response: <底层错误>` |
| 网络请求失败 | `baidu translator: request: <底层错误>` |

API 响应的 `trans_result` 为空数组时返回 `("", nil)`，不算错误。

### 常见错误码

以下为常见错误码，完整列表见[官方文档](https://fanyi-api.baidu.com/api/trans/product/apidoc)：

| 错误码 | 含义 |
|--------|------|
| 52000 | 成功 |
| 52001 | 请求超时 |
| 52002 | 系统错误 |
| 52003 | 未授权用户 |
| 54000 | 必填参数为空 |
| 54001 | 签名错误 |
| 54002 | 请求频率超限 |
| 54003 | 请求非法 |
| 54004 | 未知字段 |
| 54005 | 未知翻译类型 |
| 58000 | 客户端 IP 非法 |
| 58001 | 目标语言方向不受支持 |
| 90107 | 认证失败 |

### 使用统一接口

本模块不 import `go-utils/translator`，但方法签名与 `Translator` 接口一致，可直接赋值：

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

### 测试注入 Mock 端点

```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = json.NewEncoder(w).Encode(map[string]any{
		"trans_result": []map[string]string{{"src": "hello", "dst": "你好"}},
	})
}))

tr := baidu.NewTranslator("appid", "secret", baidu.WithEndpoint(srv.URL))
```

## 注意事项

1. **签名算法**：`sign = MD5(appID + query + salt + secretKey)`，其中 `salt` 为每次请求随机生成的 `uint64`。签名自动计算，无需手工处理。
2. **长文本**：百度 API 按行拆分文本返回多段结果，本模块用 `\n` 合并整体返回。单次请求长度与 QPS 限制见[官方文档](https://fanyi-api.baidu.com/api/trans/product/apidoc)。
3. **超时**：默认 HTTP 客户端与单次请求上下文均以 30 秒为上限；需要不同策略时经 `WithHTTPClient` 替换客户端。
4. 响应体最多读取 10 MiB，超长响应以解码错误失败，不会耗尽内存。
5. 每次调用都会请求百度 API，注意账号额度与计费；`Translator` 实例可复用。

## 鸣谢

- [百度翻译开放平台](https://fanyi-api.baidu.com/)

## 许可证

基于项目 [MIT 许可证](https://github.com/tx7do/go-utils/blob/main/LICENSE)发布。
