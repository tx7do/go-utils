# 阿里翻译 (Alibaba Translator)

阿里翻译是阿里云提供的机器翻译服务，支持 200+ 种语言对组合，具有高准确度和低延迟。

## 特性

- 🌍 **200+ 语言支持**：涵盖全球主流语言
- ⚡ **高性能低延迟**：优化的请求处理
- 🔐 **企业级安全**：支持阿里云完整的安全机制
- 🎯 **准确度高**：采用神经网络机器翻译 (NMT)
- 📱 **接口统一**：实现标准 `Translator` 接口
- 🔧 **灵活配置**：支持自定义区域设置

## 安装

```bash
go get -u github.com/chenmingyong/go-utils/translator/alibaba
```

## 快速开始

### 基础使用

```go
package main

import (
	"fmt"
	"log"
	"github.com/chenmingyong/go-utils/translator/alibaba"
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

	fmt.Println("Translation:", result) // 你好，世界！
}
```

### 使用自定义区域

```go
translator, err := alibaba.NewTranslator(
	"your_access_key_id",
	"your_access_key_secret",
	alibaba.WithRegionID("cn-shanghai"),
)
if err != nil {
	log.Fatal(err)
}

result, err := translator.Translate("Hello", "en", "zh")
if err != nil {
	log.Fatal(err)
}

fmt.Println("Translation:", result)
```

## API 文档

### NewTranslator

创建一个新的阿里翻译器。

```go
func NewTranslator(accessKeyID, accessKeySecret string, opts ...Option) (*Translator, error)
```

**参数：**
- `accessKeyID`: 阿里云访问密钥 ID
- `accessKeySecret`: 阿里云访问密钥密码
- `opts`: 可选的配置选项

**返回值：**
- `*Translator`: 翻译器实例
- `error`: 错误信息

**示例：**
```go
translator, err := alibaba.NewTranslator(
	"LTAI5tXXXXXXXXXXXXXX",
	"xXXXXXXXXXXXXXXXXXXXXXXXXX",
)
if err != nil {
	log.Fatal(err)
}
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
result, err := translator.Translate("Hello", "en", "zh")
if err != nil {
	log.Fatal(err)
}
fmt.Println(result) // 你好
```

### Options（配置选项）

#### WithRegionID

设置阿里云区域 ID。

```go
func WithRegionID(regionID string) Option
```

**支持的区域：**
- `cn-hangzhou`: 杭州（默认）
- `cn-shanghai`: 上海
- `cn-beijing`: 北京
- `cn-zhangjiakou`: 张家口
- `cn-huhehaote`: 呼和浩特
- `cn-wulanchabu`: 乌兰察布
- `cn-chengdu`: 成都
- `cn-hongkong`: 香港
- `ap-southeast-1`: 新加坡
- `ap-southeast-2`: 悉尼
- `ap-northeast-1`: 日本
- `ap-south-1`: 孟买
- `us-west-1`: 美国西部
- `us-east-1`: 美国东部
- `eu-west-1`: 欧洲

**示例：**
```go
alibaba.WithRegionID("cn-shanghai")
```

## 支持的语言代码

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
| 乌尔都语 | `ur` | 巴基斯坦语 | `pa` |
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

## 获取访问凭证

### 获取 AccessKey ID 和 Secret

1. 访问 [阿里云控制台](https://console.aliyun.com/)
2. 登录你的阿里云账号
3. 进入 "访问控制 (IAM)" 服务
4. 创建 AccessKey：用户 → 我的信息 → AccessKey 管理
5. 创建新的 AccessKey
6. 保存 AccessKey ID 和 Secret

### 启用机器翻译服务

1. 访问 [阿里云机器翻译服务](https://www.aliyun.com/product/ai/mtranslation)
2. 开通服务
3. 选择合适的套餐（免费/按量付费）

### 费用

- **免费额度**：每月 100 万字符
- **按量付费**：¥50/百万字符

## 错误处理

```go
result, err := translator.Translate("Hello", "en", "zh")
if err != nil {
	switch err.(type) {
	case *alibaba.TranslateError:
		log.Println("翻译 API 错误:", err)
	default:
		log.Println("其他错误:", err)
	}
}
```

## 常见错误码

| 错误码 | 含义 | 解决方案 |
|--------|------|--------|
| `InvalidParameter.RegionId` | 无效的区域 ID | 检查区域 ID 是否正确 |
| `InvalidParameter.SourceLanguage` | 不支持的源语言 | 检查语言代码 |
| `InvalidParameter.TargetLanguage` | 不支持的目标语言 | 检查语言代码 |
| `Forbidden.Quota` | 配额已用尽 | 升级套餐或等待刷新 |
| `Unauthorized` | 认证失败 | 检查 AccessKey ID 和 Secret |

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

### 4. 超时控制

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// 使用带超时的 context
```

## 常见问题

### Q: 如何获取 AccessKey ID 和 Secret？

A: 登录阿里云控制台，进入 IAM 服务，在 AccessKey 管理中创建新的 AccessKey。

### Q: 支持哪些语言？

A: 支持 200+ 种语言，包括所有主流语言。详见上面的语言表。

### Q: 翻译是否有字符限制？

A: 单次请求无字符限制，但超过 5000 字符会按比例收费。

### Q: 如何处理翻译错误？

A: 检查网络连接、凭证有效性、语言代码、配额使用情况等。

### Q: 翻译准确度如何？

A: 使用神经网络机器翻译 (NMT) 技术，准确度高于统计模型。

## 性能建议

1. **连接复用**：重用翻译器实例
2. **批量处理**：一次翻译多个文本减少开销
3. **缓存策略**：缓存常用翻译结果
4. **异步翻译**：使用 goroutine 并发翻译
5. **监控日志**：记录翻译失败和性能数据

## 依赖项

```
github.com/alibabacloud-go/alimt-20190107 v1.0.2
github.com/alibabacloud-go/darabonba-openapi/v2 v2.1.15
github.com/alibabacloud-go/tea v1.4.0
```

## License

遵循项目原有的 License。

## 相关链接

- [阿里云机器翻译官方文档](https://help.aliyun.com/product/2477301.html)
- [阿里云控制台](https://console.aliyun.com/)
- [阿里云 API 参考](https://next.api.aliyun.com/document/Mt/2018-10-12/Translate)
- [项目主页](https://github.com/chenmingyong/go-utils)

