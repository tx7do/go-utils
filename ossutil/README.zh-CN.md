# ossutil

[English](README.md) | 简体中文

对象存储的通用工具函数（与具体 SDK 解耦的部分）：桶映射、类型嗅探、对象命名、URL 拼接与上传安全约束。

## 功能特性

- **桶映射**（`constants.go` / `utils.go`）：`ContentTypeToBucketName` / `FileExtensionToBucketName` 把 MIME 类型或扩展名映射到分类桶（images/videos/audios/docs/files），畸形 content-type 一律落默认桶；
- **类型检测**：`DetectFileType` 按魔术头嗅探 MIME 与扩展名（带前导点），`EnsureFileExtension` / `ExtractFileExtension` / `ContentTypeToFileExtension` 的扩展名统一**不带前导点**；
- **对象命名**：`GenerateFileName` 按 UUID / 内容 SHA256 / HMAC 内容 / 时间戳四种策略生成文件名，`GenerateObjectName` / `EnsureObjectName` / `JoinObjectName` 拼目录与文件名；`GenerateFileNameTypeHMACContent` 策略依赖的密钥由 `SetHMACSecret` 在启动阶段注入——未注入时退化为 UUID 命名，不再有内置弱密钥；
- **URL 拼接**：`JoinObjectUrl` / `ReplaceEndpointHost`；
- **上传约束**（`constants.go`）：`MaxUploadSize` / `MaxDownloadSize`（各 50 MiB）、MIME 白名单 `IsAllowedMimeType`、路径穿越校验 `IsFileDirectorySafe`。

## 安装

独立模块（依赖同仓库的 `id` 包做 UUID 生成）：

```bash
go get github.com/tx7do/go-utils/ossutil@latest
```

```go
import "github.com/tx7do/go-utils/ossutil"
```

## 快速开始

一个完整可编译的示例：嗅探内容类型 → 选桶 → 按内容 SHA256 生成带扩展名的对象名 → 拼出访问 URL。

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/ossutil"
)

func main() {
	content := []byte("hello world")

	// 1. 按内容嗅探 MIME 与扩展名（ext 带前导点）
	mimeType, ext := ossutil.DetectFileType(content)

	// 2. 上传白名单校验（先于一切投递动作）
	if !ossutil.IsAllowedMimeType(mimeType) {
		fmt.Println("mime not allowed:", mimeType)
		return
	}

	// 3. 桶映射：image/* → images；畸形/未知类型 → files
	bucket := ossutil.ContentTypeToBucketName(mimeType)

	// 4. 生成对象名（dir + 文件名），扩展名从文件名/类型/内容依次兜底，最终兜底 "bin"
	objectName := ossutil.EnsureObjectName("avatar/2024", "photo", mimeType, content, ossutil.GenerateFileNameTypeUUID)

	// 5. 拼 URL
	url := ossutil.JoinObjectUrl("https://cdn.example.com", bucket, objectName)

	fmt.Println(mimeType, ext, bucket, objectName, url)
}
```

## 常量与白名单

**桶名常量**（`utils.go`）：

```go
const (
	BucketImages = "images"
	BucketVideos = "videos"
	BucketAudios = "audios"
	BucketDocs   = "docs"
	BucketFiles  = "files" // 默认桶
)
```

**命名策略常量**（`utils.go`）：

```go
const (
	GenerateFileNameTypeUUID          GenerateFileNameType = "uuid"
	GenerateFileNameTypeContentSHA256 GenerateFileNameType = "content_sha256"
	GenerateFileNameTypeHMACContent   GenerateFileNameType = "hmac_content"
	GenerateFileNameTypeTimeBase      GenerateFileNameType = "time_base"
)
```

**上传大小上限**（`constants.go`）：`MaxUploadSize` / `MaxDownloadSize` 均为 50 MiB。

**MIME 白名单**是两个可配置的包级变量（`constants.go`），调用方可按部署需求增删：

```go
var AllowedMimePrefixes = []string{"image/", "video/", "audio/"}     // 前缀匹配

var AllowedExactMimeTypes = map[string]struct{}{ /* text/plain、pdf、zip、Office 文档等精确匹配 */ }
```

`IsAllowedMimeType` 先经标准库 `mime.ParseMediaType` 归一（剥离 charset 等参数）再做前缀与精确两级匹配；解析失败的畸形类型串直接拒绝。

## 对象命名

四种策略的命名函数可直接调用（扩展名前导点自动剥离）：

```go
// UUID（去横线）：几乎无碰撞、不依赖内容、最快；不能内容去重。
func GeneraUUIDFileName(fileExt string) string

// 内容 SHA256 十六进制摘要：确定性、可去重校验；会泄露内容指纹，需读全文件。
func GeneraContentSHA265FileName(fileContent []byte, fileExt string) string

// 先 SHA256 再以 secret 做 HMAC：同样内容一致（可去重）且不可被离线匹配。
func GenerateHMACContentFileName(secret []byte, fileContent []byte, fileExt string) string

// UnixNano 时间戳 + 8 位十六进制随机后缀（下划线连接）：可读、按时间排序；
// 同秒高并发下碰撞概率上升，不适合内容去重。
func GeneraTimeBaseFileName(fileExt string) string
```

> 注意 `GeneraContentSHA265FileName` 的 `SHA265` 是源码实际拼写（历史命名，为兼容保留），功能是 SHA-256。

组合入口：

```go
// 按策略生成文件名；HMACContent 策略未注入密钥时退化为 UUID。
func GenerateFileName(fileContent []byte, fileExt string, typ GenerateFileNameType) string

// 目录 + 生成文件名（目录首尾斜杠清理）。
func GenerateObjectName(fileDirectory string, fileContent []byte, fileExt string, typ GenerateFileNameType) string

// 文件名扩展名兜底（文件名 → Content-Type → 内容嗅探 → "bin"）后再生成对象名；typ 为空按 UUID。
func EnsureObjectName(fileDirectory, sourceFileName, contentType string, fileContent []byte, typ GenerateFileNameType) string

// 按 Content-Type 补扩展名并拼接目录；fileName 传 nil 时生成 UUID 文件名。
func JoinObjectName(contentType string, filePath, fileName *string) (objectName string, fileName string)
```

HMAC 策略的密钥在启动阶段注入一次：

```go
ossutil.SetHMACSecret([]byte("your-strong-secret"))
```

## 类型检测

```go
// 按魔术头嗅探，返回 MIME 与扩展名（ext 带前导点，未知为空串）。
func DetectFileType(fileContent []byte) (mimeType, ext string)

// Content-Type → 不带点的扩展名（png/pdf/docx…），未知返回空串。
func ContentTypeToFileExtension(contentType string) string

// 文件名 → 不带点的小写扩展名（点在首位视作无扩展名）。
func ExtractFileExtension(fileName string) string

// 三级兜底取扩展名：文件名 → Content-Type → 内容嗅探 → "bin"。
func EnsureFileExtension(fileName, contentType string, fileContent []byte) string
```

## 桶映射与 URL

```go
// MIME → 桶：image/* → images、video/* → videos、audio/* → audios、
// text/* 与 pdf/json/Office 类 → docs，其余与畸形串 → files。
func ContentTypeToBucketName(contentType string) string

// 扩展名 → 桶（传入含 "/" 的串时复用 ContentTypeToBucketName）。
func FileExtensionToBucketName(ext string) string

// endpoint/bucket/object 三段拼接（各自清理多余斜杠）。
func JoinObjectUrl(endpoint, bucketName, objectName string) string

// 把 URL 中的 endpoint 段替换为 host（用于内网 endpoint 换外网域名等场景）。
func ReplaceEndpointHost(rawURL, host, endpoint string) string
```

## 上传安全约束

```go
// MIME 白名单校验（前缀 + 精确两级匹配，输入先归一）。
func IsAllowedMimeType(mimeType string) bool

// 目录名安全校验：仅字母/数字/下划线/连字符/斜杠；拒绝 ".." 与绝对路径。
// 空串合法（落根命名空间）。
func IsFileDirectorySafe(dir string) bool
```

## Notes / 注意事项

- 扩展名约定：`DetectFileType` 返回的 ext **带**前导点；`EnsureFileExtension` / `ExtractFileExtension` / `ContentTypeToFileExtension` 返回的**不带**——拼接时注意补点；
- `GenerateFileNameTypeHMACContent` 未注入密钥时退化为 UUID（连可去重性也一并失去），部署方必须在启动阶段调用 `SetHMACSecret`；
- 时间戳命名暴露时间信息且随机后缀只有 8 个十六进制字符，高并发去重场景请改用内容/HMAC 策略；
- 需要绑定具体 SDK（如 minio-go 的 `GetObjectOptions.SetRange`）的辅助函数由使用方仓库自行维护，本包刻意不依赖任何对象存储 SDK。
