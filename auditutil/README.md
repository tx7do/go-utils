# auditutil

HTTP 审计日志的信息采集与防篡改辅助。

## 采集

- `ClientRealIP` / `IPFromRemoteAddr` — 客户端 IP 的归一化提取
  （按转发头**如实记录**口径；防伪造的安全判定请用 `netutil` 的可信代理链版本）；
- `IsPrivateIP` — 内网/链路本地地址判定；
- `RequestID` — X-Request-ID / X-Correlation-ID / x-fc-request-id 的提取与回退生成；
- `ParseUsernameFromBytes` / `ExtractUsernameFromRequest` — 请求体（JSON 字段或表单键）
  中的用户名解析，附带 `StripLineBreaks` 的 CR/LF 剥离（防日志行注入）；
  Body 读取后恢复，不影响后续处理；
- `DetectPlatformFromUA` — UA 的平台类型启发式（仅作补充，优先客户端上报字段）。

## 防篡改

- `HashLog(msg)` — 审计日志的确定性 SHA256：在消息**副本**上清除
  `log_hash` / `signature` 字段（防自引用）后做 protobuf 序列化再哈希；
  可用变参覆盖排除集；
- `GenerateECDSAKeyPair` / `EncodeECDSADER` / `SignLogContent` —
  日志签名链：`sign(priv, tenant_id, user_id, created_at, log_hash)` 的
  ECDSA-P256 签名（DER，可被 `ecdsa.VerifyASN1` 校验）。

```go
import "github.com/tx7do/go-utils/auditutil"

ip := auditutil.ClientRealIP(req)
reqID := auditutil.RequestID(req)
log.LogHash = trans.Ptr(auditutil.HashLog(log))
log.Signature = auditutil.SignLogContent(priv, log.TenantId, log.UserId, log.CreatedAt, log.GetLogHash())
```
