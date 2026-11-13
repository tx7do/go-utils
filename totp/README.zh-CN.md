# totp

RFC 6238 TOTP（基于时间的一次性密码）实现——纯 Go 标准库，零第三方依赖。

[English](README.md) | 简体中文

## 简介

本模块（`github.com/tx7do/go-utils/totp`）实现验证器 App 默认口径：**SHA1 / 6 位 / 30 秒步长**。secret 为 Base32（RFC 4648 无填充），`otpauth://` 绑定 URI 手工拼装——仅依赖标准库。

## 特性

- `GenerateSecret` —— 160-bit secret，Base32 无填充编码（32 字符）
- `ValidateSecret` —— 规范化小写/填充/内部空格并校验形态
- `Verify` —— ±1 步时钟偏移窗口（90 秒容忍），常数时间比较
- `OTPAuthURI` —— `otpauth://totp/` 绑定 URI，供扫码/手动录入
- 零依赖；经 RFC 6238 附录 B 官方参考向量验证

## 安装

```bash
go get github.com/tx7do/go-utils/totp
```

## 快速开始

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tx7do/go-utils/totp"
)

func main() {
	secret, err := totp.GenerateSecret()
	if err != nil {
		log.Fatal(err)
	}

	// 将该 URI 展示为二维码供验证器 App 扫码绑定
	uri := totp.OTPAuthURI("Example Corp", "admin@example.com", secret)
	fmt.Println(uri)

	// 校验 6 位验证码（±1 步窗口）
	ok := totp.Verify(secret, "287082", time.Now())
	fmt.Println("valid:", ok)
}
```

## API

| 函数 | 签名 | 说明 |
|------|------|------|
| `GenerateSecret` | `() (string, error)` | 新 160-bit secret，Base32 无填充 |
| `ValidateSecret` | `(secret string) (string, error)` | 规范化 + 校验；返回大写无填充形态 |
| `Verify` | `(secret, code string, at time.Time) bool` | 6 位码校验，±1 步偏移，常数时间 |
| `OTPAuthURI` | `(issuer, account, secret string) string` | `otpauth://totp/` 绑定 URI |

## 注意事项

1. 仅实现 SHA1 / 6 位 / 30 秒口径——Google Authenticator、Microsoft Authenticator、1Password 等主流验证器的默认配置。
2. `Verify` 接受 `at` 前后各一步（±30 秒）内的验证码；部署环境需要更大容忍度时调大 `verifyWindow`。
3. 比较为常数时间（`crypto/subtle`），抗时序攻击。

## 许可证

基于项目 [MIT 许可证](https://github.com/tx7do/go-utils/blob/main/LICENSE)发布。
