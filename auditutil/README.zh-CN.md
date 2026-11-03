# auditutil

[English](README.md) | 简体中文

`auditutil` 提供 HTTP 审计日志的信息采集与防篡改辅助：客户端 IP / 请求 ID 等请求信息的归一化提取、请求体用户名解析与日志行注入清洗、User-Agent 平台识别，以及审计日志的确定性哈希与 ECDSA 签名（构成"hash + 签名"的两级防篡改链）。

## 特性

- **请求信息采集**：客户端真实 IP、内网地址判定、请求 ID 提取与回退生成；
- **用户名解析**：从 JSON / 表单请求体解析用户名，剥离 CR/LF 防日志行注入，Body 读取后自动恢复；
- **平台识别**：基于 User-Agent 的平台类型启发式判断；
- **防篡改**：确定性 SHA256 哈希（排除 `log_hash` / `signature` 字段防自引用）+ ECDSA-P256 签名（DER 编码，可用标准库 `ecdsa.VerifyASN1` 验签）。

## 安装

```bash
go get github.com/tx7do/go-utils/auditutil
```

```go
import "github.com/tx7do/go-utils/auditutil"
```

## 快速开始

### 1. 请求信息采集

```go
package main

import (
	"net/http"

	"github.com/tx7do/go-utils/auditutil"
)

func collect(req *http.Request) (ip, reqID, username, platform string) {
	ip = auditutil.ClientRealIP(req)             // X-Forwarded-For -> X-Real-IP -> RemoteAddr
	reqID = auditutil.RequestID(req)             // X-Request-ID -> X-Correlation-ID -> x-fc-request-id -> 新 GUID
	username, _ = auditutil.ExtractUsernameFromRequest(req) // Body 读取后自动恢复
	platform = auditutil.DetectPlatformFromUA(req.UserAgent())
	return
}
```

### 2. 日志哈希与签名

```go
package main

import (
	"crypto/ecdsa"

	"github.com/tx7do/go-utils/auditutil"
	"github.com/tx7do/go-utils/trans"
)

// priv 来自 auditutil.GenerateECDSAKeyPair()，通常在服务启动时生成或从配置加载
var priv *ecdsa.PrivateKey

func seal(log *pb.AuditLog) error {
	// 1) 计算哈希：在消息副本上清除 log_hash/signature 后做 protobuf 确定性序列化再 SHA256
	log.LogHash = trans.Ptr(auditutil.HashLog(log))

	// 2) 签名：sign(tenant_id, user_id, created_at, log_hash)，DER 格式
	sig, err := auditutil.SignLogContent(priv, log.TenantId, log.UserId, log.CreatedAt, log.GetLogHash())
	if err != nil {
		return err
	}
	log.Signature = sig
	return nil
}
```

> 注意 `SignLogContent` 返回 `([]byte, error)` 两个值，需先接住 error 再赋给字段；`trans.Ptr` 用于取字符串指针，需要单独导入 `github.com/tx7do/go-utils/trans`。

### 3. 密钥生成与验签

签名可用标准库 `ecdsa.VerifyASN1` 校验。签名内容是与库内一致的 JSON 结构（字段顺序、键名），先 SHA256 再验签：

```go
priv, pub, err := auditutil.GenerateECDSAKeyPair() // ECDSA P-256 (secp256r1)
if err != nil {
	// ...
}

// 验签：重算签名内容 -> SHA256 -> VerifyASN1
type signContent struct {
	TenantID uint32 `json:"tenant_id"`
	UserID   uint32 `json:"user_id"`
	Sec      int64  `json:"sec"`   // createdAt.Seconds
	Nanos    int32  `json:"nanos"` // createdAt.Nanos
	LogHash  string `json:"log_hash"`
}

func verify(pub *ecdsa.PublicKey, tenantID, userID uint32, createdAt *timestamppb.Timestamp, logHash string, sig []byte) bool {
	sc := signContent{TenantID: tenantID, UserID: userID, LogHash: logHash}
	if createdAt != nil {
		sc.Sec, sc.Nanos = createdAt.Seconds, createdAt.Nanos
	}
	b, err := json.Marshal(sc)
	if err != nil {
		return false
	}
	h := sha256.Sum256(b)
	return ecdsa.VerifyASN1(pub, h[:], sig)
}
```

篡改 `log_hash` 或任何签名输入后验签必然失败；`SignLogContent` 在私钥为 nil 时返回错误而非空签名——签名即凭证，调用方不应降级为无签名落库。

## 配置 / 选项

本包无可配置项。防篡改链的行为由约定固定：

- **哈希排除集**：`HashLog` 默认在副本上清除 `log_hash`、`logHash`、`signature` 三种拼写字段；可用变参覆盖：`HashLog(msg, "tenant_id")`；
- **签名内容**：`tenant_id`、`user_id`、`created_at`（秒 + 纳秒）、`log_hash` 的 JSON 序列化的 SHA256；`createdAt` 为 nil 时按 0 处理。

## API 参考

### 请求信息采集（client_ip.go / request.go）

- `func ClientRealIP(request *http.Request) string` —— 依次取 `X-Forwarded-For` 中第一个合法 IP、`X-Real-IP`，都取不到时回退 `RemoteAddr`；
- `func IPFromRemoteAddr(hostAddress string) string` —— 从 `host:port` 提取合法 IP，解析失败返回空串；
- `func IsPrivateIP(ipStr string) bool` —— 判断是否内网/链路本地地址（10/8、172.16/12、192.168/16、127/8、169.254/16、::1、fc00::/7）；无法解析返回 false；
- `func RequestID(request *http.Request) string` —— 依次检查 `X-Request-ID` / `X-Correlation-ID` / `x-fc-request-id`，都没有时生成新的 GUIDv4；
- `func ExtractUsernameFromRequest(r *http.Request) (string, error)` —— 读取并恢复 Body 后解析用户名，不影响后续中间件；
- `func ParseUsernameFromBytes(body []byte) (string, error)` —— 先按 JSON 字段 `"username"`（正则）解析，再按表单键（`url.ParseQuery`）解析；两级都未命中报错；
- `func StripLineBreaks(s string) string` —— 移除所有 CR/LF，阻断日志行注入（用户名解析结果已自动应用）。

### 平台识别（platform.go）

- `func DetectPlatformFromUA(ua string) string` —— 按 UA 启发式判断平台，返回值域为下列常量；空 UA 返回 `PlatformOther`；
- 平台常量：`PlatformAndroidApp`（`"AndroidApp"`）、`PlatformiOSApp`（`"iOSApp"`）、`PlatformDesktopWindows`（`"DesktopWindows"`）、`PlatformDesktopMac`（`"DesktopMac"`）、`PlatformDesktopLinux`（`"DesktopLinux"`）、`PlatformWeb`（`"Web"`）、`PlatformOther`（`"Other"`）。

### 防篡改（sign.go）

- `func GenerateECDSAKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error)` —— 生成 secp256r1 密钥对；
- `func EncodeECDSADER(r, s *big.Int) ([]byte, error)` —— 将 ECDSA 的 r、s 编码为 DER 序列；
- `func HashLog(msg proto.Message, excludeFields ...string) string` —— 计算日志的确定性 SHA256（十六进制小写）；在**副本**上清除排除字段后做 protobuf 确定性序列化再哈希，原消息不被修改；`msg` 为 nil 或无效时返回空串；
- `func SignLogContent(priv *ecdsa.PrivateKey, tenantID, userID uint32, createdAt *timestamppb.Timestamp, logHash string) ([]byte, error)` —— 对签名内容做 ECDSA-P256 签名，返回 DER 格式字节（`0x30` 开头）；`priv` 为 nil 返回错误。

## 注意事项

- **ClientRealIP 是"如实记录"口径**：按转发头原样记录审计信息，最外层反向代理必须覆写（而非追加）`X-Forwarded-For` 以防伪造；需要防伪造的安全判定（可信代理链从右向左解析）请使用 `netutil.ClientIPFromRequest`；
- **DetectPlatformFromUA 仅作补充判断**：优先使用客户端上报的 `platform` / `os_name` 字段；
- **哈希防自引用**：`log_hash` 与 `signature` 字段被排除在哈希之外，字段名与 `json_name` 两种拼写都命中；
- **可重放验证**：验签方只需知道 `tenant_id`、`user_id`、`created_at`、`log_hash` 即可重算签名内容并用标准库验签，无需访问日志原文；
- **依赖**：`google.golang.org/protobuf`（proto.Message 哈希）、`github.com/tx7do/go-utils/id`（GUID 生成）。

## 致谢

- [google.golang.org/protobuf](https://google.golang.org/protobuf)
- Go 标准库 `crypto/ecdsa` / `crypto/sha256`
