# ossutil

English | [简体中文](README.zh-CN.md)

Generic helpers for object storage (the SDK-agnostic parts): bucket mapping, content sniffing, object naming, URL joining, and upload safety constraints.

## Features

- **Bucket mapping** (`constants.go` / `utils.go`): `ContentTypeToBucketName` / `FileExtensionToBucketName` map MIME types or extensions onto classification buckets (images/videos/audios/docs/files); malformed content types always fall to the default bucket;
- **Type detection**: `DetectFileType` sniffs MIME and extension (leading dot included) from magic bytes; extensions returned by `EnsureFileExtension` / `ExtractFileExtension` / `ContentTypeToFileExtension` uniformly **omit the leading dot**;
- **Object naming**: `GenerateFileName` builds file names under four strategies (UUID / content SHA256 / HMAC content / timestamp); `GenerateObjectName` / `EnsureObjectName` / `JoinObjectName` combine directories and file names. The secret for the `GenerateFileNameTypeHMACContent` strategy is injected at startup via `SetHMACSecret` — without it the strategy degrades to UUID naming, with no built-in weak key;
- **URL joining**: `JoinObjectUrl` / `ReplaceEndpointHost`;
- **Upload constraints** (`constants.go`): `MaxUploadSize` / `MaxDownloadSize` (50 MiB each), the MIME whitelist `IsAllowedMimeType`, and the path-traversal check `IsFileDirectorySafe`.

## Installation

Standalone module (depends on the sibling `id` package for UUID generation):

```bash
go get github.com/tx7do/go-utils/ossutil@latest
```

```go
import "github.com/tx7do/go-utils/ossutil"
```

## Quick Start

A complete, compilable example: sniff the content type → pick the bucket → generate a content-SHA256 object name with extension → join the access URL.

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/ossutil"
)

func main() {
	content := []byte("hello world")

	// 1. Sniff MIME and extension from content (ext carries the leading dot)
	mimeType, ext := ossutil.DetectFileType(content)

	// 2. Upload whitelist check (before any delivery step)
	if !ossutil.IsAllowedMimeType(mimeType) {
		fmt.Println("mime not allowed:", mimeType)
		return
	}

	// 3. Bucket mapping: image/* → images; malformed/unknown types → files
	bucket := ossutil.ContentTypeToBucketName(mimeType)

	// 4. Object name (dir + file name); the extension falls back through
	//    file name / content type / sniffing, ending at "bin"
	objectName := ossutil.EnsureObjectName("avatar/2024", "photo", mimeType, content, ossutil.GenerateFileNameTypeUUID)

	// 5. Join the URL
	url := ossutil.JoinObjectUrl("https://cdn.example.com", bucket, objectName)

	fmt.Println(mimeType, ext, bucket, objectName, url)
}
```

## Constants & Whitelist

**Bucket name constants** (`utils.go`):

```go
const (
	BucketImages = "images"
	BucketVideos = "videos"
	BucketAudios = "audios"
	BucketDocs   = "docs"
	BucketFiles  = "files" // default bucket
)
```

**Naming strategy constants** (`utils.go`):

```go
const (
	GenerateFileNameTypeUUID          GenerateFileNameType = "uuid"
	GenerateFileNameTypeContentSHA256 GenerateFileNameType = "content_sha256"
	GenerateFileNameTypeHMACContent   GenerateFileNameType = "hmac_content"
	GenerateFileNameTypeTimeBase      GenerateFileNameType = "time_base"
)
```

**Upload size caps** (`constants.go`): `MaxUploadSize` / `MaxDownloadSize` are both 50 MiB.

**The MIME whitelist** is two package-level variables (`constants.go`) that callers may tune per deployment:

```go
var AllowedMimePrefixes = []string{"image/", "video/", "audio/"}     // prefix match

var AllowedExactMimeTypes = map[string]struct{}{ /* exact matches: text/plain, pdf, zip, Office docs… */ }
```

`IsAllowedMimeType` normalizes its input with the standard library's `mime.ParseMediaType` (stripping charset and other parameters) before a two-level prefix/exact match; malformed type strings that fail parsing are rejected outright.

## Object Naming

The four strategy functions can be called directly (leading dots on extensions are stripped):

```go
// UUID (dashes removed): near-zero collisions, content-independent, fastest; no content dedup.
func GeneraUUIDFileName(fileExt string) string

// Hex SHA-256 digest of the content: deterministic, dedup/verifiable; leaks a content
// fingerprint and requires reading the whole file.
func GeneraContentSHA265FileName(fileContent []byte, fileExt string) string

// SHA-256 first, then HMAC with a secret: identical content stays identical (dedup)
// yet cannot be matched offline.
func GenerateHMACContentFileName(secret []byte, fileContent []byte, fileExt string) string

// UnixNano timestamp + 8 hex random suffix (joined with an underscore): readable,
// time-sortable; collision odds rise under high same-second concurrency, and it
// cannot dedup by content.
func GeneraTimeBaseFileName(fileExt string) string
```

> Note: `GeneraContentSHA265FileName` keeps the historical source-code spelling `SHA265` for compatibility; it computes SHA-256.

Combining entry points:

```go
// Generate a file name per strategy; the HMACContent strategy degrades to UUID when no secret is set.
func GenerateFileName(fileContent []byte, fileExt string, typ GenerateFileNameType) string

// Directory + generated file name (slashes trimmed at both ends of the directory).
func GenerateObjectName(fileDirectory string, fileContent []byte, fileExt string, typ GenerateFileNameType) string

// Resolve the extension with fallbacks (file name → content type → sniffing → "bin"), then build the object name; empty typ means UUID.
func EnsureObjectName(fileDirectory, sourceFileName, contentType string, fileContent []byte, typ GenerateFileNameType) string

// Append an extension derived from Content-Type and join the directory; a nil fileName generates a UUID file name.
func JoinObjectName(contentType string, filePath, fileName *string) (objectName string, fileName string)
```

Inject the HMAC strategy's secret once at startup:

```go
ossutil.SetHMACSecret([]byte("your-strong-secret"))
```

## Type Detection

```go
// Sniff by magic bytes; returns the MIME and extension (ext carries the leading dot, empty when unknown).
func DetectFileType(fileContent []byte) (mimeType, ext string)

// Content-Type → dot-less extension (png/pdf/docx…), empty string when unknown.
func ContentTypeToFileExtension(contentType string) string

// File name → lower-cased dot-less extension (a leading dot does not count as one).
func ExtractFileExtension(fileName string) string

// Three-level fallback for extensions: file name → Content-Type → sniffing → "bin".
func EnsureFileExtension(fileName, contentType string, fileContent []byte) string
```

## Bucket Mapping & URLs

```go
// MIME → bucket: image/* → images, video/* → videos, audio/* → audios,
// text/* and pdf/json/Office-like → docs; everything else and malformed strings → files.
func ContentTypeToBucketName(contentType string) string

// Extension → bucket (strings containing "/" are re-dispatched through ContentTypeToBucketName).
func FileExtensionToBucketName(ext string) string

// Join endpoint/bucket/object (each segment's stray slashes trimmed).
func JoinObjectUrl(endpoint, bucketName, objectName string) string

// Replace the endpoint segment in a URL with host (e.g. intranet endpoint swapped for a public domain).
func ReplaceEndpointHost(rawURL, host, endpoint string) string
```

## Upload Safety Constraints

```go
// MIME whitelist check (two-level prefix + exact match, input normalized first).
func IsAllowedMimeType(mimeType string) bool

// Directory name safety: only letters/digits/underscore/hyphen/slash; rejects ".." and absolute paths.
// The empty string is allowed (root namespace).
func IsFileDirectorySafe(dir string) bool
```

## Notes / Caveats

- Extension convention: the ext returned by `DetectFileType` **includes** the leading dot; those returned by `EnsureFileExtension` / `ExtractFileExtension` / `ContentTypeToFileExtension` **omit** it — add the dot yourself when joining;
- `GenerateFileNameTypeHMACContent` degrades to UUID when no secret is injected (losing dedup too), so deployments must call `SetHMACSecret` at startup;
- Timestamp naming leaks time information and its random suffix is only 8 hex characters — prefer the content/HMAC strategies for high-concurrency dedup scenarios;
- Helpers that bind a concrete SDK (such as minio-go's `GetObjectOptions.SetRange`) are maintained in the consuming repositories; this package deliberately depends on no object-storage SDK.
