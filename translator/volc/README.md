# 火山翻译 (Volc Translator)

火山翻译 (Volc) 是字节跳动旗下的翻译服务。

## 使用

### 基础使用

```go
package main

import (
	"fmt"
	"log"

	"github.com/chenmingyong/go-utils/translator/volc"
)

func main() {
	// 创建翻译器，需要提供 AccessKey 和 SecretKey
	translator := volc.NewTranslator("your_access_key", "your_secret_key")

	// 翻译文本
	result, err := translator.Translate("Hello, world!", "en", "zh")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Translation result:", result)
}
```

## 获取 AccessKey 和 SecretKey

1. 访问[火山翻译控制台](https://console.volcengineapi.com/)
2. 创建项目并获取 AccessKey 和 SecretKey
3. 将凭证信息传入 NewTranslator 函数

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

## 接口

### NewTranslator

创建一个新的火山翻译器。

```go
func NewTranslator(accessKey, secretKey string) *Translator
```

**参数：**
- `accessKey`: 火山翻译的 AccessKey
- `secretKey`: 火山翻译的 SecretKey

**返回值：**
- `*Translator`: 翻译器实例

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

## 注意事项

1. 需要配置正确的 AccessKey 和 SecretKey
2. 请确保账户有足够的翻译配额
3. 请遵守火山翻译的使用协议和限速要求

