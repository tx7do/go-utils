# doctext

[English](README.md) | 简体中文

从常见文档格式中抽取纯文本，供知识库 RAG 入库、全文搜索、预览等场景使用。按文件扩展名路由到对应抽取器，其余扩展名直接拒绝——宁可报「不支持」也不要把二进制乱码灌进向量库。

## Features

- **支持格式**：纯文本族 `.txt` / `.md` / `.markdown` / `.csv` / `.log` / `.json` / `.xml` / `.yml` / `.yaml` / `.html` / `.htm`，以及 `.docx`（zip + XML 标准库解析 `word/document.xml`）和 `.pdf`（[ledongthuc/pdf](https://github.com/ledongthuc/pdf) 纯 Go 抽取文本层）
- **大小上限**：默认 10 MiB（`MaxFileSize = 10 << 20`），超限报错
- **输出清洗**：强制有效 UTF-8（非法字节替换为 U+FFFD），`\r\n` 统一为 `\n`
- **docx 处理**：段落 `</w:p>` 与 `<w:br/>` 转换行、`<w:tab/>` 转制表符、剥除全部 XML 标签、反转义 HTML 实体
- **PDF 处理**：逐页抽取纯文本（非 OCR），单页失败不阻断整本；无文本层的扫描件/图片型 PDF 报错
- `SupportedExtensions()` 返回支持扩展名清单，供管理页提示用

## Installation

本包为独立 Go 模块，单独安装：

```bash
go get github.com/tx7do/go-utils/doctext
```

```go
import "github.com/tx7do/go-utils/doctext"
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tx7do/go-utils/doctext"
)

func main() {
	data, err := os.ReadFile("manual.docx")
	if err != nil {
		log.Fatal(err)
	}

	text, err := doctext.Extract("manual.docx", data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(text)

	// 管理页展示支持的格式
	fmt.Println(doctext.SupportedExtensions())
}
```

## API 参考

| 成员 | 签名 | 说明 |
| --- | --- | --- |
| `Extract` | `func Extract(fileName string, data []byte) (string, error)` | 按扩展名（大小写不敏感）抽取纯文本；先检查空内容与大小上限，再按扩展名分发 |
| `SupportedExtensions` | `func SupportedExtensions() []string` | 返回全部 13 个支持扩展名 |
| `MaxFileSize` | `const MaxFileSize = 10 << 20` | 10 MiB 上限 |

### 错误场景

| 场景 | 错误消息 |
| --- | --- |
| 内容为空 | `file is empty` |
| 超过 10 MiB | `file too large: %d bytes (max %d)` |
| 不支持的扩展名 | `unsupported file type %q (supported: %s)` |
| 纯文本族抽出的内容全部为空白 | `no text content extracted` |
| docx 缺少 `word/document.xml` | `docx extract: word/document.xml not found in docx` |
| docx 剥标签后为空 | `docx extract: no text content extracted from docx` |
| PDF 无法打开 | `pdf extract: open pdf: ...` |
| PDF 无文本层（扫描件/图片型） | `pdf extract: no text layer found (scanned/image pdf is not supported)` |

## Notes / Caveats

- 扩展名匹配大小写不敏感（`MANUAL.DOCX` 与 `manual.docx` 等价）。
- 纯文本族不识别编码，按原始字节处理；非法 UTF-8 字节会被替换为 U+FFFD，不报错。
- PDF 抽取**不做 OCR**：扫描件没有文本层，会得到「无文本层」错误而非图片内容。
- 空检查先于扩展名检查：空文件无论扩展名是什么都报 `file is empty`。
- docx 只解析 `word/document.xml` 主文档，不提取页眉、页脚、表格外的嵌入对象。

## Acknowledgements

- [github.com/ledongthuc/pdf](https://github.com/ledongthuc/pdf) — 纯 Go PDF 文本抽取
