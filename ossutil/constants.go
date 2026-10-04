package ossutil

import (
	"mime"
	"strings"
)

// 上传文件的安全约束常量。
// 这些值用于在文件上传与下载环节防御滥用（超大文件、可执行/HTML 等危险类型）。

const (
	// MaxUploadSize 单次直传文件最大字节数（50 MiB）。
	MaxUploadSize int64 = 50 << 20

	// MaxDownloadSize 从外部 URL 下载文件的最大字节数（50 MiB），用于 SSRF 路径防 DoS。
	MaxDownloadSize int64 = 50 << 20
)

// AllowedMimePrefixes 是允许上传的 MIME 类型前缀白名单。
// 通过嗅探真实文件内容得到 MIME 后，需命中其中之一才允许上传。
var AllowedMimePrefixes = []string{
	"image/", // png / jpg / gif / webp / bmp ...
	"video/", // mp4 / webm / quicktime ...
	"audio/", // mpeg / wav / ogg ...
	// 常见文档类型（精确值而非前缀），在 IsAllowedMimeType 内逐项精确匹配
}

// AllowedExactMimeTypes 是允许上传的精确 MIME 类型（非前缀匹配）。
var AllowedExactMimeTypes = map[string]struct{}{
	"text/plain":                   {},
	"text/markdown":                {},
	"application/json":             {},
	"application/pdf":              {},
	"application/zip":              {},
	"application/gzip":             {},
	"application/x-gzip":           {},
	"application/x-tar":            {},
	"application/x-7z-compressed":  {},
	"application/x-rar-compressed": {},
	// Office 文档
	"application/msword": {},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {},
	"application/vnd.ms-excel": {},
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         {},
	"application/vnd.ms-powerpoint":                                             {},
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": {},
}

// IsAllowedMimeType 判断某 MIME 类型是否在上传白名单内。
// 同时支持前缀（image/* 等）与精确匹配（pdf/zip 等）。
// 输入先经标准库 mime.ParseMediaType 归一：剥离参数（charset 等）与空白后
// 再比对——内容嗅探器对文本恒产 "text/plain; charset=utf-8" 带参形式，
// 按原串匹配会把白名单内的文本类型整体误拒；解析失败的畸形类型串
// （如 "image/*"）结构不合法，直接拒绝。
func IsAllowedMimeType(mimeType string) bool {
	if mimeType == "" {
		return false
	}
	mt, _, err := mime.ParseMediaType(mimeType)
	if err != nil {
		return false
	}
	mt = strings.ToLower(mt)

	if _, ok := AllowedExactMimeTypes[mt]; ok {
		return true
	}
	for _, prefix := range AllowedMimePrefixes {
		if len(mt) >= len(prefix) && mt[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// IsFileDirectorySafe 校验客户端传入的 FileDirectory 是否安全。
// 仅允许字母、数字、下划线、连字符与斜杠；拒绝 ..、绝对路径等穿越/注入手法。
func IsFileDirectorySafe(dir string) bool {
	if dir == "" {
		return true // 空目录允许（落到根命名空间）
	}
	// 整体字符白名单
	for _, r := range dir {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '_', r == '-', r == '/':
			// 合法字符
		default:
			return false
		}
	}
	// 拒绝路径穿越段
	if strings.Contains(dir, "..") {
		return false
	}
	// 拒绝绝对路径样式
	if strings.HasPrefix(dir, "/") {
		return false
	}
	return true
}
