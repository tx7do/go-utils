# slug

URL slug（短链接/别名）生成工具：基于
[github.com/gosimple/slug](https://github.com/gosimple/slug) v1.15.0 的薄封装，
把任意语言的文本（中文、俄文、德文、英文等）转换为适合 URL 的 slug。

[English](README.md) | 简体中文

## 特性

- 任意语言文本转 URL slug（内部经 unicode 音译处理）
- 默认强制小写，也可保留大小写
- 提供英文语料与德文语料（`&` → `und`）两种语言变体
- 透传上游的 `slug.CustomSub` 自定义替换能力

## 安装

slug 是独立 Go 模块（`slug/go.mod`），依赖 `github.com/gosimple/slug v1.15.0`：

```bash
go get github.com/tx7do/go-utils/slug
```

```go
import "github.com/tx7do/go-utils/slug"
```

## 快速开始

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/slug"
)

func main() {
	fmt.Println(slug.Generate("Hellö Wörld хелло ворлд")) // hello-world-khello-vorld
	fmt.Println(slug.Generate("影師"))                     // ying-shi
	fmt.Println(slug.Generate("天下太平"))                   // tian-xia-tai-ping

	// 英文语料：& 转为 and
	fmt.Println(slug.GenerateEnglish("This & that")) // this-and-that

	// 保留大小写
	fmt.Println(slug.GenerateCaseSensitive("This & That")) // This-and-That

	// 德文语料：& 转为 und
	fmt.Println(slug.GenerateGerman("Diese & Dass")) // diese-und-dass
}
```

## API 参考

| 函数 | 签名 | 说明 |
|------|------|------|
| `Generate` | `func(input string) string` | 生成 slug：强制小写，使用英文语料（`slug.MakeLang(input, "en")`） |
| `GenerateCaseSensitive` | `func(input string) string` | 生成 slug：保留大小写，使用英文语料 |
| `GenerateEnglish` | `func(input string) string` | 与 `Generate` 等价的语义别名（强制小写 + 英文语料） |
| `GenerateGerman` | `func(input string) string` | 生成 slug：强制小写，使用德文语料（`slug.MakeLang(input, "de")`，`&` 转为 `und`） |

所有函数均转调上游 `slug.MakeLang`；转换规则（unicode 音译、非法字符剔除、空白转
`-` 等）由上游库决定。

### 自定义替换

上游暴露了包级变量 `slug.CustomSub`（`map[string]string`），可在调用前设置自定义
替换规则（先于音译执行）：

```go
slug.CustomSub = map[string]string{
	"water": "sand",
}
slug.Generate("water is hot") // sand-is-hot
```

## 注意事项

- **并发场景谨慎**：上游的 `slug.Lowercase` 与 `slug.CustomSub` 是包级全局变量，
  本包的每个函数在每次调用时都会改写 `slug.Lowercase`（如 `Generate` 设为 `true`、
  `GenerateCaseSensitive` 设为 `false`）。多 goroutine 并发混用不同函数时，
  大小写行为可能互相覆盖；`slug.CustomSub` 一旦设置即为全局生效。需要严格隔离时，
  建议直接使用上游库并自行加锁，或按语种固定使用单一函数。
- `Generate` 与 `GenerateEnglish` 行为完全一致，按语义偏好选用即可。
- 音译结果取决于上游 [gosimple/unidecode](https://github.com/gosimple/unidecode)
  的转写表：俄文 `хелло` 转写为 `khello`、繁体 `影師` 转写为 `ying shi` 等，均以
  单元测试（`slug_test.go`）断言为准。
