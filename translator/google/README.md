# Google 翻译 (Google Translator)

Google 翻译是谷歌提供的强大在线翻译服务，支持 100+ 种语言。本模块提供了三个版本的 Google 翻译 API 集成。

## 特性

- 📚 **三个版本支持**：V1、V2、V3 API
- 🌍 **100+ 语言支持**：覆盖全球大部分主流语言
- 🔄 **灵活的配置**：支持 Options 模式动态配置
- 🚀 **高性能**：优化的请求处理
- 📱 **接口统一**：实现标准 `Translator` 接口

## 安装

```bash
go get -u github.com/chenmingyong/go-utils/translator/google
```

## 快速开始

### 基础使用（V1 - 免费方案）

```go
package main

import (
	"fmt"
	"log"
	"github.com/chenmingyong/go-utils/translator/google"
)

func main() {
	// V1 不需要 API Key，可直接使用
	translator := google.NewTranslator()

	result, err := translator.Translate("Hello, world!", "en", "zh-CN")
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

result, err := translator.Translate("Hello", "en", "zh")
if err != nil {
log.Fatal(err)
}

fmt.Println("Translation:", result)
```

### 使用 V3 API（最新版本）

```go
translator := google.NewTranslator(
google.WithVersion("v3"),
google.WithApiKey("your_google_api_key"),
)

result, err := translator.Translate("Hello", "en", "zh")
if err != nil {
log.Fatal(err)
}

fmt.Println("Translation:", result)
```

## API 文档

### NewTranslator

创建一个新的 Google 翻译器。

```go
func NewTranslator(opts ...Option) *Translator
```

**参数：**

- `opts`: 可选的配置选项

**返回值：**

- `*Translator`: 翻译器实例

**示例：**

```go
translator := google.NewTranslator(
google.WithVersion("v2"),
google.WithApiKey("your_api_key"),
)
```

### Translate

翻译文本。

```go
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error)
```

**参数：**

- `source`: 要翻译的文本
- `sourceLang`: 源语言代码
- `targetLang`: 目标语言代码

**返回值：**

- `string`: 翻译结果
- `error`: 错误信息

**示例：**

```go
result, err := translator.Translate("Hello", "en", "zh-CN")
if err != nil {
log.Fatal(err)
}
fmt.Println(result) // 你好
```

### Options（配置选项）

#### WithVersion

设置 API 版本。

```go
func WithVersion(version string) Option
```

**支持的版本：**

- `"v1"`: 免费 API（默认），无需认证
- `"v2"`: 谷歌官方 V2 API，需要 API Key
- `"v3"`: 谷歌最新的 V3 API，需要 API Key

**示例：**

```go
google.WithVersion("v2")
```

#### WithApiKey

设置 Google API Key。

```go
func WithApiKey(key string) Option
```

**示例：**

```go
google.WithApiKey("AIzaSyDo...")
```

## 三个版本的对比

| 特性         | V1     | V2  | V3  |
|------------|--------|-----|-----|
| 是否免费       | ✅ 是    | ❌ 否 | ❌ 否 |
| 需要 API Key | ❌ 否    | ✅ 是 | ✅ 是 |
| 官方支持       | ⚠️ 半支持 | ✅ 是 | ✅ 是 |
| 速度         | 中      | 快   | 最快  |
| 准确度        | 良好     | 很好  | 最优  |
| 使用限制       | 有（非官方） | 有   | 有   |

### 版本选择建议

- **V1**：适合小流量项目、演示、学习
- **V2**：适合需要稳定、官方支持的项目
- **V3**：适合大规模、高性能需求的项目

## 支持的语言代码

### 主流语言

| 语言     | 代码      | 语言    | 代码   |
|--------|---------|-------|------|
| 中文（简体） | `zh-CN` | 英语    | `en` |
| 中文（繁体） | `zh-TW` | 日语    | `ja` |
| 韩语     | `ko`    | 俄语    | `ru` |
| 德语     | `de`    | 法语    | `fr` |
| 西班牙语   | `es`    | 意大利语  | `it` |
| 葡萄牙语   | `pt`    | 荷兰语   | `nl` |
| 瑞典语    | `sv`    | 挪威语   | `no` |
| 丹麦语    | `da`    | 芬兰语   | `fi` |
| 波兰语    | `pl`    | 捷克语   | `cs` |
| 罗马尼亚语  | `ro`    | 匈牙利语  | `hu` |
| 保加利亚语  | `bg`    | 克罗地亚语 | `hr` |
| 塞尔维亚语  | `sr`    | 乌克兰语  | `uk` |
| 俄语     | `ru`    | 白俄罗斯语 | `be` |

### 亚洲语言

| 语言     | 代码   | 语言    | 代码   |
|--------|------|-------|------|
| 泰语     | `th` | 越南语   | `vi` |
| 老挝语    | `lo` | 缅甸语   | `my` |
| 柬埔寨语   | `km` | 印尼语   | `id` |
| 马来语    | `ms` | 菲律宾语  | `tl` |
| 孟加拉语   | `bn` | 印地语   | `hi` |
| 乌尔都语   | `ur` | 巴基斯坦语 | `pa` |
| 古吉拉特语  | `gu` | 卡纳达语  | `kn` |
| 泰卢固语   | `te` | 泰米尔语  | `ta` |
| 马拉雅拉姆语 | `ml` | 马拉地语  | `mr` |

### 中东和非洲语言

| 语言    | 代码   | 语言    | 代码   |
|-------|------|-------|------|
| 阿拉伯语  | `ar` | 希伯来语  | `iw` |
| 波斯语   | `fa` | 土耳其语  | `tr` |
| 库尔德语  | `ku` | 阿塞拜疆语 | `az` |
| 斯瓦希里语 | `sw` | 豪萨语   | `ha` |
| 约鲁巴语  | `yo` | 伊博语   | `ig` |
| 科萨语   | `xh` | 祖鲁语   | `zu` |

### 欧洲语言

| 语言     | 代码   | 语言     | 代码   |
|--------|------|--------|------|
| 阿尔巴尼亚语 | `sq` | 巴斯克语   | `eu` |
| 加泰罗尼亚语 | `ca` | 科西嘉语   | `co` |
| 加利西亚语  | `gl` | 冰岛语    | `is` |
| 爱尔兰语   | `ga` | 拉丁语    | `la` |
| 卢森堡语   | `lb` | 马耳他语   | `mt` |
| 威尔士语   | `cy` | 苏格兰盖尔语 | `gd` |

### 其他语言

| 语言    | 代码    | 语言      | 代码    |
|-------|-------|---------|-------|
| 格鲁吉亚语 | `ka`  | 亚美尼亚语   | `hy`  |
| 阿姆哈拉语 | `am`  | 爱沙尼亚语   | `et`  |
| 拉脱维亚语 | `lv`  | 立陶宛语    | `lt`  |
| 马其顿语  | `mk`  | 蒙古语     | `mn`  |
| 尼泊尔语  | `ne`  | 僧伽罗语    | `si`  |
| 世界语   | `eo`  | 毛利语     | `mi`  |
| 夏威夷语  | `haw` | 海地克里奥尔语 | `ht`  |
| 萨摩亚语  | `sm`  | 卢旺达语    | `rw`  |
| 马尔加什语 | `mg`  | 哈萨克语    | `kk`  |
| 吉尔吉斯语 | `ky`  | 乌兹别克语   | `uz`  |
| 塔吉克语  | `tg`  | 土库曼语    | `tk`  |
| 维吾尔语  | `ug`  | 普什图语    | `ps`  |
| 信德语   | `sd`  | 奥利亚语    | `or`  |
| 鞑靼语   | `tt`  | 弗里西语    | `fy`  |
| 修纳语   | `sn`  | 齐切瓦语    | `ny`  |
| 索马里语  | `so`  | 印尼爪哇语   | `jw`  |
| 印尼巽他语 | `su`  | 宿务语     | `ceb` |
| 意第绪语  | `yi`  | 苗语      | `hmn` |

## 获取 API Key

### 获取 Google API Key

1. 访问 [Google Cloud Console](https://console.cloud.google.com/)
2. 创建新项目
3. 启用 Google Translate API
4. 创建 API Key
5. 复制 API Key 到配置中

### 费用

- **V1**: 免费（非官方 API）
- **V2**: 按使用量计费（约 ¥15.5/100万字符）
- **V3**: 按使用量计费（约 ¥15.5/100万字符）

## 错误处理

```go
result, err := translator.Translate("Hello", "en", "zh-CN")
if err != nil {
switch err.Error() {
case "error getting translate.googleapis.com":
log.Println("网络连接错误")
case "error 400 (Bad Request)":
log.Println("请求参数错误")
case "error unmarshalling data":
log.Println("响应解析错误")
default:
log.Println("其他错误:", err)
}
}
```

## 最佳实践

### 1. 错误处理

```go
result, err := translator.Translate(text, sourceLang, targetLang)
if err != nil {
// 记录错误
log.Printf("翻译失败: %v", err)
// 返回原文或缓存结果
return text, nil
}
```

### 2. 缓存翻译结果

```go
var cache = make(map[string]string)

func translateWithCache(text, sourceLang, targetLang string) string {
key := fmt.Sprintf("%s_%s_%s", text, sourceLang, targetLang)
if v, ok := cache[key]; ok {
return v
}

result, _ := translator.Translate(text, sourceLang, targetLang)
cache[key] = result
return result
}
```

### 3. 批量翻译

```go
func translateBatch(items []string, sourceLang, targetLang string) []string {
results := make([]string, len(items))
for i, item := range items {
result, _ := translator.Translate(item, sourceLang, targetLang)
results[i] = result
}
return results
}
```

### 4. 超时控制（V2/V3）

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// 在 V2/V3 中使用 context 控制超时
```

## 常见问题

### Q: V1 API 会被官方停用吗？

A: V1 是非官方 API，没有 SLA 保证，但目前仍在工作。建议重要项目使用 V2 或 V3。

### Q: 如何选择语言代码？

A:

- 对于中文，使用 `zh-CN`（简体）或 `zh-TW`（繁体）
- 对于英文，使用 `en`
- 其他语言参考上面的语言表格

### Q: V1 API 有速率限制吗？

A: 是的，建议不要超过 1000 请求/天。

### Q: 如何处理翻译错误？

A: 检查网络连接、语言代码、API Key 有效性等。

## 依赖项

```
cloud.google.com/go/translate v1.12.7
golang.org/x/text v0.34.0
google.golang.org/api v0.269.0
```

## License

遵循项目原有的 License。

## 相关链接

- [Google Translate API 文档](https://cloud.google.com/translate/docs)
- [Google Translate 语言代码](https://cloud.google.com/translate/docs/languages)
- [Google Cloud Console](https://console.cloud.google.com/)
- [项目主页](https://github.com/chenmingyong/go-utils)
