# 加解密算法

## 主要接口

本包统一定义了如下接口，便于不同算法的无缝切换和组合：

- `Cipher`：对称/非对称加密接口
- `Signer`：签名接口（如SM2、ECDSA等）
- `Verifier`：验签接口
- `Hasher`：哈希接口（如SM3、SHA256等）
- `KeyExchanger`：密钥协商接口（如ECDH、SM2密钥交换）

接口定义见 `interface.go`，如：

```go
// Cipher
Encrypt(plain []byte) ([]byte, error)
Decrypt(cipher []byte) ([]byte, error)
Name() string

// Signer
Sign(data []byte) (string, error)
Name() string

// Verifier
Verify(data []byte, signature string) (bool, error)
Name() string
```

---

## SM2Cipher 用法示例

```go
cipher, _ := NewSM2Cipher()
plain := []byte("hello, sm2!")

// 加解密
crypted, _ := cipher.Encrypt(plain)
decrypted, _ := cipher.Decrypt(crypted)

// 签名/验签
sig, _ := cipher.Sign(plain)
ok, _ := cipher.Verify(plain, sig)
```

---

## AES 用法示例

```go
key := []byte("1234567890abcdef") // 16字节
cipher, _ := NewAESCipher(key, nil)
plain := []byte("hello, aes!")
crypted, _ := cipher.Encrypt(plain)
decrypted, _ := cipher.Decrypt(crypted)
```

---

## RSA 用法示例

```go
cipher, _ := NewRSACipher(2048)
plain := []byte("hello, rsa!")
crypted, _ := cipher.Encrypt(plain)
decrypted, _ := cipher.Decrypt(crypted)
```

---

## SM4 用法示例

```go
key := []byte("1234567890abcdef") // 16字节
cipher, _ := NewSM4Cipher(key)
plain := []byte("hello, sm4!")
crypted, _ := cipher.Encrypt(plain)
decrypted, _ := cipher.Decrypt(crypted)
```

---

## HMAC/SM3 用法示例

```go
h := NewHMAC([]byte("key"))
mac := h.Sum([]byte("hello"))

h2 := NewSM3Hasher()
hash := h2.Sum([]byte("hello"))
```


---

## 全局加密器与载荷封装（encryptor.go / sign.go / payload.go）

面向"配置类敏感字段落库加密"场景的更上层封装（与上表的原语接口相互独立）：

- **`Encryptor`**：AES-256-GCM 认证加密，密文带 `enc:` 前缀、base64 编码，
  密钥由任意长度口令经 SHA-256 派生；`IsEncrypted` 判定前缀；
- **全局单例**：`InitGlobalEncryptor(key, enabled)` / `GetGlobalEncryptor` /
  `EncryptIfNeeded` / `DecryptIfNeeded`——未配置密钥时透传明文
  （`DecryptIfNeeded` 对无前缀串原样返回，兼容历史明文行）；
- **`SignData` / `VerifyData`**：绑定全局密钥的 HMAC-SHA256 签名门面
  （用于签名 URL 等场景；未初始化时显式报错而非产出空签名）；
- **载荷封装**：`EncryptPayload(payload, reservedKeys...)` 把整份配置
  map 序列化加密为 `{_encrypted_config, _is_encrypted}` 封装，
  `reservedKeys`（如任务系统的 `task_id`/`task_type`）以明文保留供
  无需解密的路由/调度层读取；`DecryptPayload` /
  `HasEncryptedPayload` 处理封装的解出与识别，
  `Must*` 系列为测试便捷变体。

```go
_ = InitGlobalEncryptor(os.Getenv("APP_CRYPTO_KEY"), true)
cipherText, _ := EncryptIfNeeded("smtp-password")
plain, _ := DecryptIfNeeded(row.Config)
```

---

## 其它说明

- AES/SM4：对称加密，均实现 Cipher 接口
- RSA：非对称加密，Cipher 接口
- HMAC/SM3/SHA256：哈希算法，实现 Hasher 接口
- ECDSA/SM2：签名验签，Signer/Verifier 接口
- ECDH/SM2：密钥协商，实现 KeyExchanger 接口

所有算法均可通过接口组合和替换，便于扩展和测试。
