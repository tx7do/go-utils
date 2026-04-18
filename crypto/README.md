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

## 其它说明

- AES/SM4：对称加密，均实现 Cipher 接口
- RSA：非对称加密，Cipher 接口
- HMAC/SM3/SHA256：哈希算法，实现 Hasher 接口
- ECDSA/SM2：签名验签，Signer/Verifier 接口
- ECDH/SM2：密钥协商，实现 KeyExchanger 接口

所有算法均可通过接口组合和替换，便于扩展和测试。
