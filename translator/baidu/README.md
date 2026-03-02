# 百度翻译 (Baidu Translator)

百度翻译是百度提供的在线翻译服务。

## 使用

### 基础使用

```go
package main

import (
	"fmt"
	"log"

	"github.com/chenmingyong/go-utils/translator/baidu"
)

func main() {
	// 创建翻译器，需要提供 AppID 和 SecretKey
	translator := baidu.NewTranslator("your_app_id", "your_secret_key")

	// 翻译文本
	result, err := translator.Translate("Hello, world!", "en", "zh")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Translation result:", result)
}
```

## 获取 AppID 和 SecretKey

1. 访问[百度翻译开放平台](https://fanyi-api.baidu.com/)
2. 登录或注册账号
3. 在"我的应用"中创建新应用
4. 获取 AppID 和 SecretKey

## 支持的语言

| 语言 | 代码 |
|------|------|
| 自动检测 | auto |
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
| 希腊语 | el |
| 瑞典语 | sv |
| 丹麦语 | da |
| 芬兰语 | fi |
| 荷兰语 | nl |
| 挪威语 | no |
| 捷克语 | cs |
| 罗马尼亚语 | ro |
| 匈牙利语 | hu |
| 保加利亚语 | bg |
| 斯洛文尼亚语 | sl |
| 斯洛伐克语 | sk |
| 克罗地亚语 | hr |
| 塞尔维亚语 | sr |
| 爱沙尼亚语 | et |
| 立陶宛语 | lt |
| 拉脱维亚语 | lv |
| 阿尔巴尼亚语 | sq |
| 马其顿语 | mk |
| 亚美尼亚语 | hy |
| 格鲁吉亚语 | ka |
| 白俄罗斯语 | be |
| 冰岛语 | is |
| 库尔德语 | ku |
| 吉尔吉斯语 | ky |
| 拉丁语 | la |
| 卢森堡语 | lb |
| 高棉语 | km |
| 梵语 | sa |
| 威尔士语 | cy |
| 尼泊尔语 | ne |
| 孟加拉语 | bn |
| 旁遮普语 | pa |
| 梵语 | sa |
| 乌尔都语 | ur |
| 马来语 | ms |
| 缅甸语 | my |
| 泰卢固语 | te |
| 泰米尔语 | ta |
| 卡纳达语 | kn |
| 马拉雅拉姆语 | ml |
| 古吉拉特语 | gu |
| 马拉地语 | mr |
| 奥利亚语 | or |
| 旁遮普语 | pa |
| 信德语 | sd |
| 僧伽罗语 | si |

更多语言支持请参考[官方文档](https://fanyi-api.baidu.com/api/trans/product/apidoc)。

## 接口

### NewTranslator

创建一个新的百度翻译器。

```go
func NewTranslator(appID, secretKey string) *Translator
```

**参数：**
- `appID`: 百度翻译的 AppID
- `secretKey`: 百度翻译的 SecretKey

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

## 错误处理

百度翻译 API 可能返回以下错误码：

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

## 注意事项

1. 需要配置正确的 AppID 和 SecretKey
2. 请确保账户有足够的翻译配额
3. 请遵守百度翻译的使用协议和限速要求
4. 签名使用 MD5 算法，格式为：`appID + query + salt + secretKey` 的 MD5 值

