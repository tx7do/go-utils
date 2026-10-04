# doctext

从常见文档格式中抽取纯文本，供知识库 RAG 入库、全文搜索、预览等场景使用。

支持格式：纯文本族（txt/md/csv/log/json/xml/yaml/html）、docx（zip+XML 标准库解析）、
pdf（ledongthuc/pdf 纯 Go 抽取，扫描件/图片型 PDF 无文本层会返回空并报错）。
其余扩展名直接拒绝——宁可报"不支持"也不要把二进制乱码灌进向量库。

```go
import "github.com/tx7do/go-utils/doctext"

text, err := doctext.Extract("manual.docx", data)
exts := doctext.SupportedExtensions() // 管理页提示用
```

默认 10MB 大小上限（`MaxFileSize`），抽取内容强制有效 UTF-8 并规范化换行。
