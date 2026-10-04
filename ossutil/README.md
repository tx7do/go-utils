# ossutil

对象存储的通用工具函数（与具体 SDK 解耦的部分）：

- **桶映射**（`constants.go` / `utils.go`）：`ContentTypeToBucketName` /
  `FileExtensionToBucketName` 把 MIME 类型或扩展名映射到分类桶
  （images/videos/audios/docs/files），畸形 content-type 一律落默认桶；
- **类型检测**：`DetectFileType` 按魔术头嗅探 MIME 与扩展名（带前导点），
  `EnsureFileExtension` / `ExtractFileExtension` / `ContentTypeToFileExtension`
  的扩展名统一**不带前导点**；
- **对象命名**：`GenerateFileName` 按 UUID / 内容 SHA256 / HMAC 内容 /
  时间戳四种策略生成文件名，`GenerateObjectName` / `EnsureObjectName` /
  `JoinObjectName` 拼目录与文件名。`GenerateFileNameTypeHMACContent`
  策略依赖的密钥由 `SetHMACSecret` 在启动阶段注入——未注入时退化为
  UUID 命名，不再有内置弱密钥；
- **URL 拼接**：`JoinObjectUrl` / `ReplaceEndpointHost`；
- **上传约束**（`constants.go`）：`MaxUploadSize` / `MaxDownloadSize`、
  MIME 白名单 `IsAllowedMimeType`、路径穿越校验 `IsFileDirectorySafe`。

```go
import "github.com/tx7do/go-utils/ossutil"

if !ossutil.IsAllowedMimeType(realMime) { ... }
bucket := ossutil.ContentTypeToBucketName(mime)
name := ossutil.EnsureObjectName(dir, fileName, mime, content, ossutil.GenerateFileNameTypeUUID)
```

需要绑定具体 SDK（如 minio-go 的 `GetObjectOptions.SetRange`）的辅助
函数由使用方仓库自行维护。
