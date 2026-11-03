# password

[English](README.md) | 简体中文

`password` 是一个独立 Go 模块（`github.com/tx7do/go-utils/password`），提供密码哈希、消息认证与基础非对称加密能力，覆盖 Bcrypt、PBKDF2、Argon2id、加盐 SHA、HMAC、RSA、ECDSA/ECDH 等算法，并附带统一的 `Crypto` 接口与口令复杂度校验。

## 功能特性

- 统一的 `Crypto` 接口（`Encrypt` / `Verify`），算法可热切换
- 密码哈希：Bcrypt、PBKDF2（SHA-256/SHA-512）、Argon2id、加盐 SHA-256/SHA-512——全部内置随机盐
- 消息认证：HMAC-SHA256
- 非对称：RSA（OAEP-SHA256 加解密 + PEM 导出）、ECDSA（P-256 签名/验签）、ECDH（P-256 密钥交换）
- 口令复杂度校验：`ValidateComplexity`

## 安装

本包是独立模块（有自己的 `go.mod`），不是根模块 `github.com/tx7do/go-utils` 的子包：

```bash
go get github.com/tx7do/go-utils/password
```

```go
import "github.com/tx7do/go-utils/password"
```

## 快速开始

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/password"
)

func main() {
	// 按算法名创建（bcrypt / pbkdf2 / argon2）
	crypto, err := password.CreateCrypto("bcrypt")
	if err != nil {
		log.Fatal(err)
	}

	hash, err := crypto.Encrypt("s3cret!")
	if err != nil {
		log.Fatal(err)
	}

	ok, err := crypto.Verify("s3cret!", hash)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("verify:", ok)

	// 口令复杂度：长度 ≥ 10 且至少包含小写/大写/数字/符号中的三类
	if err := password.ValidateComplexity("S3cret!Passw0rd", 10); err != nil {
		log.Fatal(err) // password.ErrWeakPassword 或长度不足错误
	}
}
```

## 各算法最小用法

### Bcrypt（密码存储，首选之一）

```go
c := password.NewBCryptCrypto()          // 默认 cost = bcrypt.DefaultCost(10)
c := password.NewBCryptCrypto(12)        // 自定义 cost（合法范围 bcrypt.MinCost~MaxCost）
hash, _ := c.Encrypt("password")
ok, _ := c.Verify("password", hash)
```

### PBKDF2（密码存储，兼容性好）

```go
c := password.NewPBKDF2Crypto()      // PBKDF2-HMAC-SHA256：310000 次迭代、32 字节密钥
c := password.NewPBKDF2WithSHA512()  // PBKDF2-HMAC-SHA512：600000 次迭代、64 字节密钥
hash, _ := c.Encrypt("password")     // 格式 pbkdf2:<hash>:<iterations>:<salt>:<key>
ok, _ := c.Verify("password", hash)
```

### Argon2id（密码存储，内存困难）

```go
c := password.NewArgon2Crypto()      // Memory=64MB, Iterations=3, Parallelism=2, Salt=16B, Key=32B
c.Memory = 128 * 1024                // 参数可按机器规格调整
hash, _ := c.Encrypt("password")     // 标准 $argon2id$v=...$m=...,t=...,p=...$salt$hash 格式
ok, _ := c.Verify("password", hash)
```

### 加盐 SHA-256/SHA-512（快速哈希）

```go
c := password.NewSHA256Crypto()      // 内置 16 字节随机盐，无需手动加盐
c := password.NewSHA512Crypto()
hash, _ := c.Encrypt("password")     // 格式 sha256:$salt$hash
ok, _ := c.Verify("password", hash)
```

> SHA 家族是快速哈希，不适合作为面向暴力破解的口令存储首选；优先使用 Bcrypt/PBKDF2/Argon2。

### HMAC-SHA256（消息认证）

```go
c := password.NewHMACCrypto("shared-secret")
mac, _ := c.Encrypt("payload")       // HMAC-SHA256 十六进制串
ok, _ := c.Verify("payload", mac)
```

### RSA（非对称加解密 + 签名素材）

```go
c, _ := password.NewRSACrypto(2048)  // 生成 2048 位密钥对
cipher, _ := c.Encrypt("data")       // RSA-OAEP(SHA-256) 公钥加密，返回 base64
plain, _ := c.Decrypt(cipher)        // 私钥解密
privPEM, _ := c.ExportPrivateKey()   // PKCS#1 PEM
pubPEM, _ := c.ExportPublicKey()     // PKIX PEM
```

### ECDSA（P-256 签名/验签）

```go
c, _ := password.NewECDSACrypto()
sig, _ := c.Encrypt("message")       // 注意：这里是对消息 SHA-256 后做 ECDSA 签名，不是加密
ok, _ := c.Verify("message", sig)    // 验证签名
```

### ECDH（P-256 密钥交换）

```go
a, _ := password.NewECDHCrypto()     // A 方
b, _ := password.NewECDHCrypto()     // B 方

pubA, _ := a.Encrypt("dummy")        // 参数仅作非空校验，返回 "ecdh$<base64 公钥>"
pubB, _ := b.Encrypt("dummy")

secretA, _ := a.DeriveSharedSecret(pubB) // 双方计算结果一致
secretB, _ := b.DeriveSharedSecret(pubA)
```

## API 参考

### Crypto 接口与工厂

```go
type Crypto interface {
	// Encrypt 加密/签名，返回包含算法标识与盐值的字符串
	Encrypt(plainPassword string) (encrypted string, err error)
	// Verify 验证是否匹配
	Verify(plainPassword, encrypted string) (bool, error)
}

func CreateCrypto(algorithm string) (Crypto, error)
// 支持 "bcrypt" / "pbkdf2" / "argon2"（大小写不敏感），其他值返回错误
```

上述所有算法类型（`BCryptCrypto`、`PBKDF2Crypto`、`Argon2Crypto`、`SHACrypto`、`HMACCrypto`、`RSACrypto`、`ECDSACrypto`、`ECDHCrypto`）均实现 `Crypto` 接口。

### 口令策略

```go
func ValidateComplexity(plain string, minLen int) error
var ErrWeakPassword = errors.New("password does not meet complexity requirements: ...")
```

规则：长度达到 `minLen`（0 或负值按 1 处理）且至少包含小写/大写/数字/符号四类中的三类。长度不足返回专门的长度错误，类别不足返回 `ErrWeakPassword`。

### 算法一览

| 算法 | 构造函数 | 内置随机盐 | 典型用途 |
|------|----------|-----------|----------|
| Bcrypt | `NewBCryptCrypto(cost ...int)` | 是（算法内建） | 密码存储 |
| Argon2id | `NewArgon2Crypto()`（参数字段可调：Memory/Iterations/Parallelism/SaltLength/KeyLength） | 是 | 密码存储（高安全） |
| PBKDF2 | `NewPBKDF2Crypto()`（SHA-256, 310000 次）、`NewPBKDF2WithSHA512()`（SHA-512, 600000 次） | 是（16 字节） | 密码存储（兼容性好） |
| SHA-256/512 | `NewSHA256Crypto()`、`NewSHA512Crypto()` | 是（16 字节） | 快速校验场景 |
| HMAC-SHA256 | `NewHMACCrypto(secretKey string)` | 否（对称密钥） | 消息认证 |
| RSA | `NewRSACrypto(keySize int)` + `Decrypt`/`ExportPrivateKey`/`ExportPublicKey` | OAEP 内部随机 | 数据加解密 |
| ECDSA (P-256) | `NewECDSACrypto()`（`Encrypt` 实为签名） | 签名随机化 | 签名/验签 |
| ECDH (P-256) | `NewECDHCrypto()` + `DeriveSharedSecret(publicKey string) ([]byte, error)` | — | 密钥交换 |

> 本包不包含 AES 或其他对称分组加密算法。

## 注意事项

- **语义差异**：多数实现的 `Encrypt` 是"哈希/签名"，结果不可逆；只有 `RSACrypto` 的 `Encrypt/Decrypt` 是真正的可逆加解密；`ECDHCrypto.Encrypt` 仅返回本方公钥
- **ECDSA/ECDH 不是口令哈希**：`ECDSACrypto` 同一明文每次签名不同（ECDSA 签名含随机数），`Verify` 做的是验签而非哈希比对
- `HMACCrypto` 需要妥善保管 `secretKey`；泄露后可伪造任意 MAC
- RSA `Encrypt` 使用 OAEP-SHA256，明文长度受密钥长度限制（2048 位约 190 字节）
- 哈希结果自带算法标识与参数（PBKDF2/Argon2/SHA），`Verify` 按存储串中的参数重算并常数时间比较，升级参数后旧哈希仍可验证
- 密码存储场景建议优先 Argon2id 或 Bcrypt；SHA 系列仅用于对性能敏感且可接受快速哈希风险的场景

## 许可证

MIT
