# crypto

[English](README.md) | 简体中文

统一的 Go 加解密工具包：以 `Cipher` / `Signer` / `Verifier` / `Hasher` / `KeyExchanger` 接口组织 AES、AES-GCM、SM2、SM4、RSA、ECDSA、ECDH、HMAC、SM3、SHA 等算法实现，并提供面向「配置类敏感字段落库加密」的上层 `Encryptor` 与全局加密器门面。

## Features

- **统一接口**：`Cipher`（加解密）、`Signer` / `Verifier`（签名验签）、`Hasher`（哈希）、`KeyExchanger`（密钥协商），算法可无缝切换与组合
- **对称加密**：AES-CBC（`AESCipher`）、AES-GCM 认证加密（`AESGCMCipher`）、SM4-ECB
- **非对称加密**：RSA-OAEP（SHA-256）、SM2（支持通用 ASN.1 格式）
- **签名验签**：ECDSA（P-256）、SM2
- **密钥协商**：ECDH（P-256）
- **哈希 / HMAC**：SM3、SHA-256、SHA-512、HMAC-SHA256（hex 摘要 + 恒定时间校验）
- **落库加密门面**：`Encryptor`（AES-256-GCM，`enc:` 前缀 + base64）、全局单例 `InitGlobalEncryptor` / `EncryptIfNeeded` / `DecryptIfNeeded`、HMAC-SHA256 签名 `SignData` / `VerifyData`、配置载荷封装 `EncryptPayload` / `DecryptPayload`

## Installation

本包为独立 Go 模块，单独安装：

```bash
go get github.com/tx7do/go-utils/crypto
```

```go
import "github.com/tx7do/go-utils/crypto"
```

## Quick Start

完整可编译示例：

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/crypto"
)

func main() {
	// AES（CBC 模式）：构造函数只有单个返回值
	aes := crypto.NewAESCipher([]byte("1234567890abcdef"), nil)
	crypted, err := aes.Encrypt([]byte("hello, aes!"))
	if err != nil {
		log.Fatal(err)
	}
	plain, err := aes.Decrypt(crypted)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(plain)) // hello, aes!

	// HMAC-SHA256：Sum 返回 hex 字符串，Verify 恒定时间比较
	mac := crypto.NewHMAC([]byte("secret-key"))
	sig := mac.Sum([]byte("hello"))
	fmt.Println(mac.Verify([]byte("hello"), sig)) // true

	// SM3：Sum 返回 ([]byte, error) 双值
	sm3 := crypto.NewSM3Hasher()
	digest, err := sm3.Sum([]byte("hello"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%x\n", digest)
}
```

其它算法速览（错误处理从略）：

```go
// AES-GCM（认证加密，密文格式 nonce|ciphertext）
gcm, _ := crypto.NewAESGCMCipher([]byte("1234567890abcdef"))

// RSA（OAEP + SHA-256，生成 2048 位密钥对）
rsaCipher, _ := crypto.NewRSACipher(2048)

// SM2（生成新密钥对，支持加解密 + 签名验签）
sm2Cipher, _ := crypto.NewSM2Cipher()

// SM4（ECB 模式，密钥必须 16 字节）
sm4Cipher, _ := crypto.NewSM4Cipher([]byte("1234567890abcdef"))

// ECDSA 签名（P-256，签名格式 base64(r)$base64(s)）
ecdsaCipher, _ := crypto.NewECDSACipher()
sig, _ := ecdsaCipher.Sign([]byte("data"))
ok, _ := ecdsaCipher.Verify([]byte("data"), sig)

// ECDH 密钥协商（P-256）
ecdhCipher, _ := crypto.NewECDHCipher()
shared, _ := ecdhCipher.DeriveSharedSecret(peerPubBytes)
```

配置落库加密（`Encryptor` 门面）：

```go
_ = crypto.InitGlobalEncryptor(os.Getenv("APP_CRYPTO_KEY"), true)
cipherText, _ := crypto.EncryptIfNeeded("smtp-password")
plainText, _ := crypto.DecryptIfNeeded(cipherText)
```

## API 参考

### 核心接口（interface.go）

```go
// Cipher：加密 / 解密
Encrypt(plain []byte) ([]byte, error)
Decrypt(cipher []byte) ([]byte, error)
Name() string

// Hasher：哈希
Sum(data []byte) ([]byte, error)
Name() string

// KeyExchanger：密钥协商
DeriveSharedSecret(peerPubBytes []byte) ([]byte, error)
PublicKeyBytes() []byte

// Signer：签名
Sign(data []byte) (string, error)
Name() string

// Verifier：验签
Verify(data []byte, signature string) (bool, error)
Name() string
```

实现对照：

| 接口 | 实现 |
| --- | --- |
| `Cipher` | `AESCipher`、`AESGCMCipher`、`SM2Cipher`、`SM4Cipher`、`RSACipher` |
| `Hasher` | `SHA256Hasher`、`SHA512Hasher`、`SM3Hasher` |
| `KeyExchanger` | `ECDHCipher` |
| `Signer` / `Verifier` | `ECDSACipher`、`SM2Cipher` |

### AES-CBC（aes.go）

| 成员 | 签名 | 说明 |
| --- | --- | --- |
| `NewAESCipher` | `func NewAESCipher(key, iv []byte) *AESCipher` | 单返回值构造；key 为空时回退 `DefaultAESKey` |
| `DefaultAESKey` | `var DefaultAESKey = []byte("f51d66a73d8a0927")` | 默认 16 字节密钥 |
| `GenerateAESKey` | `func GenerateAESKey(length int) ([]byte, error)` | 随机密钥，长度必须 16/24/32 字节 |
| `AesEncrypt` / `AesDecrypt` | `func AesEncrypt(plainText, key, iv []byte) ([]byte, error)` | 包级函数，CBC + PKCS5Padding |

行为细节：key 长度必须为 16/24/32 字节；`iv` 为 nil 时取 `key[:16]`（即块大小），非 nil 时必须为 16 字节。

### AES-GCM（aes.go）

| 成员 | 签名 | 说明 |
| --- | --- | --- |
| `NewAESGCMCipher` | `func NewAESGCMCipher(key []byte) (*AESGCMCipher, error)` | key 长度必须 16/24/32 字节 |
| `AESGCMCipher.Encrypt` | `func (a *AESGCMCipher) Encrypt(plain []byte) ([]byte, error)` | 随机 nonce，输出 `nonce\|ciphertext` |
| `AESGCMCipher.Decrypt` | `func (a *AESGCMCipher) Decrypt(ciphertext []byte) ([]byte, error)` | 长度不足 nonce 时报错 |

### SM2（sm2.go）

| 成员 | 签名 | 说明 |
| --- | --- | --- |
| `NewSM2Cipher` | `func NewSM2Cipher() (*SM2Cipher, error)` | 生成新 SM2 密钥对 |
| `NewSM2CipherFromKey` | `func NewSM2CipherFromKey(priv *sm2.PrivateKey, pub *sm2.PublicKey) *SM2Cipher` | 用已有密钥初始化 |
| `Encrypt` / `Decrypt` | `([]byte, error)` | 公钥加密 / 私钥解密，C1C3C2 排列 |
| `EncryptAsn1` / `DecryptAsn1` | `([]byte, error)` | ASN.1 格式密文，跨平台兼容性更好 |
| `Sign` / `Verify` | `Sign(digest []byte) (string, error)`；`Verify(data []byte, signature string) (bool, error)` | ASN.1 编码签名；uid 为 nil 时使用默认 UID `1234567812345678` |
| `PublicKey` / `PrivateKey` | 返回 `*sm2.PublicKey` / `*sm2.PrivateKey` | 密钥导出（私钥勿外泄） |

### SM4（sm4.go）

| 成员 | 签名 | 说明 |
| --- | --- | --- |
| `NewSM4Cipher` | `func NewSM4Cipher(key []byte) (*SM4Cipher, error)` | ECB 模式，密钥必须 16 字节 |
| `GenerateSM4Key` | `func GenerateSM4Key() ([]byte, error)` | 生成随机 16 字节密钥 |

### RSA（rsa.go）

| 成员 | 签名 | 说明 |
| --- | --- | --- |
| `NewRSACipher` | `func NewRSACipher(bits int) (*RSACipher, error)` | 生成新 RSA 密钥对 |
| `NewRSACipherFromKey` | `func NewRSACipherFromKey(priv *rsa.PrivateKey, pub *rsa.PublicKey) *RSACipher` | 用已有密钥初始化 |
| `Encrypt` / `Decrypt` | `([]byte, error)` | RSA-OAEP + SHA-256 |
| `ExportPrivateKey` | `func (r *RSACipher) ExportPrivateKey() (string, error)` | PKCS#1 PEM（`RSA PRIVATE KEY`） |
| `ExportPublicKey` | `func (r *RSACipher) ExportPublicKey() (string, error)` | PKIX PEM（`RSA PUBLIC KEY`） |
| `PublicKey` / `PrivateKey` | 返回 `*rsa.PublicKey` / `*rsa.PrivateKey` | 密钥访问 |

### ECDSA / ECDH（ecdsa_ecdh.go）

| 成员 | 签名 | 说明 |
| --- | --- | --- |
| `NewECDSACipher` | `func NewECDSACipher() (*ECDSACipher, error)` | 生成 P-256 密钥对 |
| `NewECDSACipherFromKey` | `func NewECDSACipherFromKey(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) *ECDSACipher` | 用已有密钥初始化 |
| `Sign` | `func (e *ECDSACipher) Sign(data []byte) (string, error)` | 先做 SHA-256 摘要，签名格式 `base64(r)$base64(s)` |
| `Verify` | `func (e *ECDSACipher) Verify(data []byte, signature string) (bool, error)` | 校验上述格式签名 |
| `PublicKeyBytes` | `func (e *ECDSACipher) PublicKeyBytes() []byte` | X.509 PKIX/DER 编码公钥 |
| `ParseECDSAPublicKey` | `func ParseECDSAPublicKey(data []byte) (*ecdsa.PublicKey, error)` | `PublicKeyBytes` 的逆操作 |
| `NewECDHCipher` | `func NewECDHCipher() (*ECDHCipher, error)` | 生成 P-256 密钥对（`crypto/ecdh`） |
| `ECDHCipher.PublicKeyBytes` | `func (e *ECDHCipher) PublicKeyBytes() []byte` | 本地公钥字节 |
| `ECDHCipher.DeriveSharedSecret` | `func (e *ECDHCipher) DeriveSharedSecret(peerPubBytes []byte) ([]byte, error)` | 由对端公钥字节计算共享密钥 |

`ECDSACipher` / `ECDHCipher` 均提供 `PublicKey()` / `PrivateKey()` 访问器（私钥仅用于安全存储）。

### 哈希（sha.go / sm3.go / hmac.go）

| 成员 | 签名 | 说明 |
| --- | --- | --- |
| `SHA256Sum` | `func SHA256Sum(data []byte) []byte` | 包级函数，无 error |
| `SHA512Sum` | `func SHA512Sum(data []byte) []byte` | 包级函数，无 error |
| `SHA256Hasher` / `SHA512Hasher` | `Sum(data []byte) ([]byte, error)` | 实现 `Hasher` 接口 |
| `NewSM3Hasher` | `func NewSM3Hasher() *SM3Hasher` | 构造 SM3 哈希器 |
| `SM3Hasher.Sum` | `func (h *SM3Hasher) Sum(data []byte) ([]byte, error)` | 注意：双返回值 |
| `NewHMAC` | `func NewHMAC(key []byte) *HMAC` | HMAC-SHA256 |
| `HMAC.Sum` | `func (h *HMAC) Sum(data []byte) string` | hex 编码摘要 |
| `HMAC.Verify` | `func (h *HMAC) Verify(data []byte, expected string) bool` | 恒定时间比较 |
| `HMAC.SetKey` | `func (h *HMAC) SetKey(key []byte)` | 变更密钥 |

### Encryptor 与落库加密（aes_gcm.go / encryptor.go / sign.go / payload.go）

`Encryptor`：AES-256-GCM 认证加密。密文带 `enc:` 前缀（`EncryptedPrefix`）并 base64 编码；任意非空字符串密钥经 SHA-256 派生为 32 字节密钥。

| 成员 | 签名 | 说明 |
| --- | --- | --- |
| `NewEncryptor` | `func NewEncryptor(key string) (*Encryptor, error)` | key 为空返回 `ErrInvalidKey` |
| `Encrypt` / `Decrypt` | `(string) (string, error)` | 空串入参原样返回空串；`Decrypt` 对无 `enc:` 前缀的串原样返回（兼容历史明文） |
| `IsEncrypted` | `func IsEncrypted(data string) bool` | 判定是否带 `enc:` 前缀 |
| `MustEncrypt` / `MustDecrypt` | `(string) string` | 出错即 panic，测试便捷变体 |

全局单例（`encryptor.go`，基于 `sync.Once`）：

| 成员 | 签名 | 说明 |
| --- | --- | --- |
| `InitGlobalEncryptor` | `func InitGlobalEncryptor(key string, enabled bool) error` | 仅第一次调用生效；`enabled=false` 时创建空密钥的 no-op 加密器 |
| `GetGlobalEncryptor` | `func GetGlobalEncryptor() *Encryptor` | 未初始化时返回 no-op 加密器 |
| `EncryptIfNeeded` / `DecryptIfNeeded` | `(string) (string, error)` | 未配置密钥时透传明文 |

签名门面（`sign.go`）：

| 成员 | 签名 | 说明 |
| --- | --- | --- |
| `SignData` | `func SignData(data string) (string, error)` | 绑定全局密钥的 HMAC-SHA256 签名（hex）；未初始化全局加密器时返回错误，不会降级产出空签名 |
| `VerifyData` | `func VerifyData(data, signature string) bool` | 恒定时间校验；签名能力不可用时返回 false |

配置载荷封装（`payload.go`）：

| 成员 | 签名 | 说明 |
| --- | --- | --- |
| `EncryptPayload` | `func EncryptPayload(payload map[string]interface{}, reservedKeys ...string) (map[string]interface{}, error)` | 整份配置序列化加密为 `{_encrypted_config, _is_encrypted}`；`reservedKeys` 指定的键以明文保留在外层（如 `task_id`/`task_type`）；仅当真正完成加密时才置 `_is_encrypted=true` |
| `DecryptPayload` | `func DecryptPayload(payload map[string]interface{}) (map[string]interface{}, error)` | 解出原配置；未标记加密时原样返回 |
| `HasEncryptedPayload` | `func HasEncryptedPayload(payload map[string]interface{}) bool` | 识别封装 |
| `MustEncryptPayload` / `MustDecryptPayload` | 出错即 panic | 测试便捷变体 |

## Notes

- `InitGlobalEncryptor` 基于 `sync.Once`，仅第一次调用生效（含 `enabled=false` 的情况），后续调用直接忽略。
- `NewEncryptor` 接受任意非空字符串密钥，内部经 SHA-256 派生 32 字节 AES-256 密钥；错误消息文案为 "at least 32 bytes"，但实现并无长度硬性要求（空串除外）。
- `AESCipher`（CBC，iv 缺省取 key 前 16 字节）与 `SM4Cipher`（ECB）均不带认证、加密结果确定，需要完整性保护时优先使用 `AESGCMCipher` 或 `Encryptor`。
- `HMAC` 不实现 `Hasher` 接口（其 `Sum` 返回 hex 字符串而非 `([]byte, error)`）。
- SM2 仅适合小数据块加密（如密钥交换），不适合大数据流；跨平台互通建议使用 `EncryptAsn1` / `DecryptAsn1`。
- SM2 的 `Sign` / `Verify` 在 uid 传 nil 时使用 gmsm 默认 UID `1234567812345678`，两端需保持一致。
- 本模块要求 Go 1.25+；国密算法基于 [tjfoc/gmsm](https://github.com/tjfoc/gmsm)。

## Acknowledgements

- [github.com/tjfoc/gmsm](https://github.com/tjfoc/gmsm) — SM2/SM3/SM4 国密算法实现
