# crypto

English | [简体中文](README.zh-CN.md)

A unified Go cryptography toolkit: AES, AES-GCM, SM2, SM4, RSA, ECDSA, ECDH, HMAC, SM3 and SHA implementations are organized behind the `Cipher` / `Signer` / `Verifier` / `Hasher` / `KeyExchanger` interfaces, plus a higher-level `Encryptor` and global-encryptor facade for encrypting sensitive configuration fields at rest.

## Features

- **Unified interfaces**: `Cipher` (encryption/decryption), `Signer` / `Verifier` (signing/verification), `Hasher` (hashing), `KeyExchanger` (key agreement) — algorithms can be swapped and combined seamlessly
- **Symmetric encryption**: AES-CBC (`AESCipher`), AES-GCM authenticated encryption (`AESGCMCipher`), SM4-ECB
- **Asymmetric encryption**: RSA-OAEP (SHA-256), SM2 (with interop-friendly ASN.1 format)
- **Signing / verification**: ECDSA (P-256), SM2
- **Key agreement**: ECDH (P-256)
- **Hashing / HMAC**: SM3, SHA-256, SHA-512, HMAC-SHA256 (hex digest + constant-time verification)
- **Config-at-rest facade**: `Encryptor` (AES-256-GCM, `enc:` prefix + base64), global singleton `InitGlobalEncryptor` / `EncryptIfNeeded` / `DecryptIfNeeded`, HMAC-SHA256 signing via `SignData` / `VerifyData`, and config payload wrapping via `EncryptPayload` / `DecryptPayload`

## Installation

This package is an independent Go module; install it on its own:

```bash
go get github.com/tx7do/go-utils/crypto
```

```go
import "github.com/tx7do/go-utils/crypto"
```

## Quick Start

Complete, compilable example:

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/crypto"
)

func main() {
	// AES (CBC mode): the constructor returns a single value
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

	// HMAC-SHA256: Sum returns a hex string, Verify is constant-time
	mac := crypto.NewHMAC([]byte("secret-key"))
	sig := mac.Sum([]byte("hello"))
	fmt.Println(mac.Verify([]byte("hello"), sig)) // true

	// SM3: Sum returns two values ([]byte, error)
	sm3 := crypto.NewSM3Hasher()
	digest, err := sm3.Sum([]byte("hello"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%x\n", digest)
}
```

Other algorithms at a glance (error handling omitted):

```go
// AES-GCM (authenticated encryption, ciphertext format: nonce|ciphertext)
gcm, _ := crypto.NewAESGCMCipher([]byte("1234567890abcdef"))

// RSA (OAEP + SHA-256, generates a 2048-bit key pair)
rsaCipher, _ := crypto.NewRSACipher(2048)

// SM2 (new key pair; supports encryption and signing)
sm2Cipher, _ := crypto.NewSM2Cipher()

// SM4 (ECB mode, key must be 16 bytes)
sm4Cipher, _ := crypto.NewSM4Cipher([]byte("1234567890abcdef"))

// ECDSA signing (P-256, signature format: base64(r)$base64(s))
ecdsaCipher, _ := crypto.NewECDSACipher()
sig, _ := ecdsaCipher.Sign([]byte("data"))
ok, _ := ecdsaCipher.Verify([]byte("data"), sig)

// ECDH key agreement (P-256)
ecdhCipher, _ := crypto.NewECDHCipher()
shared, _ := ecdhCipher.DeriveSharedSecret(peerPubBytes)
```

Configuration encryption at rest (the `Encryptor` facade):

```go
_ = crypto.InitGlobalEncryptor(os.Getenv("APP_CRYPTO_KEY"), true)
cipherText, _ := crypto.EncryptIfNeeded("smtp-password")
plainText, _ := crypto.DecryptIfNeeded(cipherText)
```

## API Reference

### Core interfaces (interface.go)

```go
// Cipher: encrypt / decrypt
Encrypt(plain []byte) ([]byte, error)
Decrypt(cipher []byte) ([]byte, error)
Name() string

// Hasher: hashing
Sum(data []byte) ([]byte, error)
Name() string

// KeyExchanger: key agreement
DeriveSharedSecret(peerPubBytes []byte) ([]byte, error)
PublicKeyBytes() []byte

// Signer: signing
Sign(data []byte) (string, error)
Name() string

// Verifier: signature verification
Verify(data []byte, signature string) (bool, error)
Name() string
```

Implementation map:

| Interface | Implementations |
| --- | --- |
| `Cipher` | `AESCipher`, `AESGCMCipher`, `SM2Cipher`, `SM4Cipher`, `RSACipher` |
| `Hasher` | `SHA256Hasher`, `SHA512Hasher`, `SM3Hasher` |
| `KeyExchanger` | `ECDHCipher` |
| `Signer` / `Verifier` | `ECDSACipher`, `SM2Cipher` |

### AES-CBC (aes.go)

| Member | Signature | Description |
| --- | --- | --- |
| `NewAESCipher` | `func NewAESCipher(key, iv []byte) *AESCipher` | Single-value constructor; empty key falls back to `DefaultAESKey` |
| `DefaultAESKey` | `var DefaultAESKey = []byte("f51d66a73d8a0927")` | Default 16-byte key |
| `GenerateAESKey` | `func GenerateAESKey(length int) ([]byte, error)` | Random key; length must be 16/24/32 bytes |
| `AesEncrypt` / `AesDecrypt` | `func AesEncrypt(plainText, key, iv []byte) ([]byte, error)` | Package-level functions; CBC + PKCS5Padding |

Behavior details: the key must be 16/24/32 bytes; a nil `iv` defaults to `key[:16]` (the block size), and a non-nil `iv` must be 16 bytes.

### AES-GCM (aes.go)

| Member | Signature | Description |
| --- | --- | --- |
| `NewAESGCMCipher` | `func NewAESGCMCipher(key []byte) (*AESGCMCipher, error)` | Key length must be 16/24/32 bytes |
| `AESGCMCipher.Encrypt` | `func (a *AESGCMCipher) Encrypt(plain []byte) ([]byte, error)` | Random nonce; output is `nonce\|ciphertext` |
| `AESGCMCipher.Decrypt` | `func (a *AESGCMCipher) Decrypt(ciphertext []byte) ([]byte, error)` | Errors if shorter than the nonce size |

### SM2 (sm2.go)

| Member | Signature | Description |
| --- | --- | --- |
| `NewSM2Cipher` | `func NewSM2Cipher() (*SM2Cipher, error)` | Generates a new SM2 key pair |
| `NewSM2CipherFromKey` | `func NewSM2CipherFromKey(priv *sm2.PrivateKey, pub *sm2.PublicKey) *SM2Cipher` | Initializes from existing keys |
| `Encrypt` / `Decrypt` | `([]byte, error)` | Public-key encrypt / private-key decrypt, C1C3C2 ordering |
| `EncryptAsn1` / `DecryptAsn1` | `([]byte, error)` | ASN.1 ciphertext, better cross-platform interoperability |
| `Sign` / `Verify` | `Sign(digest []byte) (string, error)`; `Verify(data []byte, signature string) (bool, error)` | ASN.1-encoded signature; a nil uid falls back to the default UID `1234567812345678` |
| `PublicKey` / `PrivateKey` | Return `*sm2.PublicKey` / `*sm2.PrivateKey` | Key export (never expose the private key) |

### SM4 (sm4.go)

| Member | Signature | Description |
| --- | --- | --- |
| `NewSM4Cipher` | `func NewSM4Cipher(key []byte) (*SM4Cipher, error)` | ECB mode; key must be 16 bytes |
| `GenerateSM4Key` | `func GenerateSM4Key() ([]byte, error)` | Generates a random 16-byte key |

### RSA (rsa.go)

| Member | Signature | Description |
| --- | --- | --- |
| `NewRSACipher` | `func NewRSACipher(bits int) (*RSACipher, error)` | Generates a new RSA key pair |
| `NewRSACipherFromKey` | `func NewRSACipherFromKey(priv *rsa.PrivateKey, pub *rsa.PublicKey) *RSACipher` | Initializes from existing keys |
| `Encrypt` / `Decrypt` | `([]byte, error)` | RSA-OAEP + SHA-256 |
| `ExportPrivateKey` | `func (r *RSACipher) ExportPrivateKey() (string, error)` | PKCS#1 PEM (`RSA PRIVATE KEY`) |
| `ExportPublicKey` | `func (r *RSACipher) ExportPublicKey() (string, error)` | PKIX PEM (`RSA PUBLIC KEY`) |
| `PublicKey` / `PrivateKey` | Return `*rsa.PublicKey` / `*rsa.PrivateKey` | Key accessors |

### ECDSA / ECDH (ecdsa_ecdh.go)

| Member | Signature | Description |
| --- | --- | --- |
| `NewECDSACipher` | `func NewECDSACipher() (*ECDSACipher, error)` | Generates a P-256 key pair |
| `NewECDSACipherFromKey` | `func NewECDSACipherFromKey(priv *ecdsa.PrivateKey, pub *ecdsa.PublicKey) *ECDSACipher` | Initializes from existing keys |
| `Sign` | `func (e *ECDSACipher) Sign(data []byte) (string, error)` | SHA-256 digest first; signature format `base64(r)$base64(s)` |
| `Verify` | `func (e *ECDSACipher) Verify(data []byte, signature string) (bool, error)` | Verifies the format above |
| `PublicKeyBytes` | `func (e *ECDSACipher) PublicKeyBytes() []byte` | Public key, X.509 PKIX/DER encoded |
| `ParseECDSAPublicKey` | `func ParseECDSAPublicKey(data []byte) (*ecdsa.PublicKey, error)` | Inverse of `PublicKeyBytes` |
| `NewECDHCipher` | `func NewECDHCipher() (*ECDHCipher, error)` | Generates a P-256 key pair (`crypto/ecdh`) |
| `ECDHCipher.PublicKeyBytes` | `func (e *ECDHCipher) PublicKeyBytes() []byte` | Local public key bytes |
| `ECDHCipher.DeriveSharedSecret` | `func (e *ECDHCipher) DeriveSharedSecret(peerPubBytes []byte) ([]byte, error)` | Derives the shared secret from the peer's public key bytes |

`ECDSACipher` / `ECDHCipher` also expose `PublicKey()` / `PrivateKey()` accessors (private keys are for secure storage only).

### Hashing (sha.go / sm3.go / hmac.go)

| Member | Signature | Description |
| --- | --- | --- |
| `SHA256Sum` | `func SHA256Sum(data []byte) []byte` | Package-level function, no error |
| `SHA512Sum` | `func SHA512Sum(data []byte) []byte` | Package-level function, no error |
| `SHA256Hasher` / `SHA512Hasher` | `Sum(data []byte) ([]byte, error)` | Implement the `Hasher` interface |
| `NewSM3Hasher` | `func NewSM3Hasher() *SM3Hasher` | Constructs an SM3 hasher |
| `SM3Hasher.Sum` | `func (h *SM3Hasher) Sum(data []byte) ([]byte, error)` | Note: two return values |
| `NewHMAC` | `func NewHMAC(key []byte) *HMAC` | HMAC-SHA256 |
| `HMAC.Sum` | `func (h *HMAC) Sum(data []byte) string` | Hex-encoded digest |
| `HMAC.Verify` | `func (h *HMAC) Verify(data []byte, expected string) bool` | Constant-time comparison |
| `HMAC.SetKey` | `func (h *HMAC) SetKey(key []byte)` | Replaces the key |

### Encryptor and config-at-rest (aes_gcm.go / encryptor.go / sign.go / payload.go)

`Encryptor`: AES-256-GCM authenticated encryption. Ciphertext carries the `enc:` prefix (`EncryptedPrefix`) and is base64 encoded; any non-empty string key is hashed with SHA-256 to derive the 32-byte key.

| Member | Signature | Description |
| --- | --- | --- |
| `NewEncryptor` | `func NewEncryptor(key string) (*Encryptor, error)` | Empty key returns `ErrInvalidKey` |
| `Encrypt` / `Decrypt` | `(string) (string, error)` | Empty input returns empty output; `Decrypt` passes through strings without the `enc:` prefix (backward compatible with legacy plaintext) |
| `IsEncrypted` | `func IsEncrypted(data string) bool` | Checks for the `enc:` prefix |
| `MustEncrypt` / `MustDecrypt` | `(string) string` | Panic on error; handy in tests |

Global singleton (encryptor.go, based on `sync.Once`):

| Member | Signature | Description |
| --- | --- | --- |
| `InitGlobalEncryptor` | `func InitGlobalEncryptor(key string, enabled bool) error` | Only the first call takes effect; `enabled=false` creates a no-op encryptor with an empty key |
| `GetGlobalEncryptor` | `func GetGlobalEncryptor() *Encryptor` | Returns a no-op encryptor when uninitialized |
| `EncryptIfNeeded` / `DecryptIfNeeded` | `(string) (string, error)` | Pass plaintext through when no key is configured |

Signing facade (sign.go):

| Member | Signature | Description |
| --- | --- | --- |
| `SignData` | `func SignData(data string) (string, error)` | HMAC-SHA256 signature (hex) bound to the global key; returns an error when the global encryptor is uninitialized instead of producing an empty signature |
| `VerifyData` | `func VerifyData(data, signature string) bool` | Constant-time verification; returns false when signing is unavailable |

Config payload wrapping (payload.go):

| Member | Signature | Description |
| --- | --- | --- |
| `EncryptPayload` | `func EncryptPayload(payload map[string]interface{}, reservedKeys ...string) (map[string]interface{}, error)` | Serializes and encrypts the whole config map into `{_encrypted_config, _is_encrypted}`; keys listed in `reservedKeys` are kept in plaintext at the top level (e.g. `task_id`/`task_type`); `_is_encrypted` is set to true only when encryption actually happened |
| `DecryptPayload` | `func DecryptPayload(payload map[string]interface{}) (map[string]interface{}, error)` | Recovers the original config; payloads not marked as encrypted are returned as-is |
| `HasEncryptedPayload` | `func HasEncryptedPayload(payload map[string]interface{}) bool` | Detects the wrapper |
| `MustEncryptPayload` / `MustDecryptPayload` | Panic on error | Handy in tests |

## Notes

- `InitGlobalEncryptor` is based on `sync.Once`: only the first call takes effect (including `enabled=false`); later calls are ignored.
- `NewEncryptor` accepts any non-empty string key and derives the 32-byte AES-256 key with SHA-256; the error message says "at least 32 bytes", but the implementation enforces no length requirement (other than non-empty).
- `AESCipher` (CBC, iv defaulting to the first 16 bytes of the key) and `SM4Cipher` (ECB) provide no authentication and are deterministic; prefer `AESGCMCipher` or `Encryptor` when integrity matters.
- `HMAC` does not implement the `Hasher` interface (its `Sum` returns a hex string rather than `([]byte, error)`).
- SM2 is only suitable for small data blocks (e.g. key exchange), not large data streams; for cross-platform interoperability use `EncryptAsn1` / `DecryptAsn1`.
- SM2 `Sign` / `Verify` use the gmsm default UID `1234567812345678` when uid is nil; both sides must agree on it.
- This module requires Go 1.25+; the Chinese national cryptography algorithms (SM2/SM3/SM4) are based on [tjfoc/gmsm](https://github.com/tjfoc/gmsm).

## Acknowledgements

- [github.com/tjfoc/gmsm](https://github.com/tjfoc/gmsm) — SM2/SM3/SM4 implementations
