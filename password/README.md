# password

English | [简体中文](README.zh-CN.md)

`password` is a standalone Go module (`github.com/tx7do/go-utils/password`) providing password hashing, message authentication, and basic asymmetric cryptography across Bcrypt, PBKDF2, Argon2id, salted SHA, HMAC, RSA, and ECDSA/ECDH, with a unified `Crypto` interface and password complexity validation.

## Features

- Unified `Crypto` interface (`Encrypt` / `Verify`) for hot-swappable algorithms
- Password hashing: Bcrypt, PBKDF2 (SHA-256/SHA-512), Argon2id, salted SHA-256/SHA-512 — all with built-in random salts
- Message authentication: HMAC-SHA256
- Asymmetric: RSA (OAEP-SHA256 encryption/decryption + PEM export), ECDSA (P-256 signing/verification), ECDH (P-256 key exchange)
- Password complexity validation: `ValidateComplexity`

## Installation

This package is a standalone module (it has its own `go.mod`), not a subpackage of the root module `github.com/tx7do/go-utils`:

```bash
go get github.com/tx7do/go-utils/password
```

```go
import "github.com/tx7do/go-utils/password"
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/tx7do/go-utils/password"
)

func main() {
	// Create by algorithm name (bcrypt / pbkdf2 / argon2)
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

	// Complexity: length >= 10 and at least 3 of (lowercase, uppercase, digit, symbol)
	if err := password.ValidateComplexity("S3cret!Passw0rd", 10); err != nil {
		log.Fatal(err) // password.ErrWeakPassword or a too-short error
	}
}
```

## Minimal Usage per Algorithm

### Bcrypt (password storage, a top pick)

```go
c := password.NewBCryptCrypto()          // default cost = bcrypt.DefaultCost (10)
c := password.NewBCryptCrypto(12)        // custom cost (valid range bcrypt.MinCost..MaxCost)
hash, _ := c.Encrypt("password")
ok, _ := c.Verify("password", hash)
```

### PBKDF2 (password storage, best compatibility)

```go
c := password.NewPBKDF2Crypto()      // PBKDF2-HMAC-SHA256: 310000 iterations, 32-byte key
c := password.NewPBKDF2WithSHA512()  // PBKDF2-HMAC-SHA512: 600000 iterations, 64-byte key
hash, _ := c.Encrypt("password")     // format pbkdf2:<hash>:<iterations>:<salt>:<key>
ok, _ := c.Verify("password", hash)
```

### Argon2id (password storage, memory-hard)

```go
c := password.NewArgon2Crypto()      // Memory=64MB, Iterations=3, Parallelism=2, Salt=16B, Key=32B
c.Memory = 128 * 1024                // tune parameters to your hardware
hash, _ := c.Encrypt("password")     // standard $argon2id$v=...$m=...,t=...,p=...$salt$hash format
ok, _ := c.Verify("password", hash)
```

### Salted SHA-256/SHA-512 (fast hashing)

```go
c := password.NewSHA256Crypto()      // built-in 16-byte random salt; no manual salting needed
c := password.NewSHA512Crypto()
hash, _ := c.Encrypt("password")     // format sha256:$salt$hash
ok, _ := c.Verify("password", hash)
```

> The SHA family is fast hashing and a poor first choice against brute-force password cracking; prefer Bcrypt/PBKDF2/Argon2.

### HMAC-SHA256 (message authentication)

```go
c := password.NewHMACCrypto("shared-secret")
mac, _ := c.Encrypt("payload")       // HMAC-SHA256 hex string
ok, _ := c.Verify("payload", mac)
```

### RSA (asymmetric encryption/decryption)

```go
c, _ := password.NewRSACrypto(2048)  // generate a 2048-bit key pair
cipher, _ := c.Encrypt("data")       // RSA-OAEP(SHA-256) public-key encryption, returns base64
plain, _ := c.Decrypt(cipher)        // private-key decryption
privPEM, _ := c.ExportPrivateKey()   // PKCS#1 PEM
pubPEM, _ := c.ExportPublicKey()     // PKIX PEM
```

### ECDSA (P-256 signing/verification)

```go
c, _ := password.NewECDSACrypto()
sig, _ := c.Encrypt("message")       // note: this signs the SHA-256 digest of the message, it does not encrypt
ok, _ := c.Verify("message", sig)    // verifies the signature
```

### ECDH (P-256 key exchange)

```go
a, _ := password.NewECDHCrypto()     // party A
b, _ := password.NewECDHCrypto()     // party B

pubA, _ := a.Encrypt("dummy")        // the argument is only checked for non-emptiness; returns "ecdh$<base64 public key>"
pubB, _ := b.Encrypt("dummy")

secretA, _ := a.DeriveSharedSecret(pubB) // both sides derive the same secret
secretB, _ := b.DeriveSharedSecret(pubA)
```

## API Reference

### Crypto Interface and Factory

```go
type Crypto interface {
	// Encrypt hashes/signs the input, returning a string that embeds the
	// algorithm identifier and salt
	Encrypt(plainPassword string) (encrypted string, err error)
	// Verify reports whether the input matches
	Verify(plainPassword, encrypted string) (bool, error)
}

func CreateCrypto(algorithm string) (Crypto, error)
// supports "bcrypt" / "pbkdf2" / "argon2" (case-insensitive); any other value returns an error
```

All algorithm types (`BCryptCrypto`, `PBKDF2Crypto`, `Argon2Crypto`, `SHACrypto`, `HMACCrypto`, `RSACrypto`, `ECDSACrypto`, `ECDHCrypto`) implement the `Crypto` interface.

### Password Policy

```go
func ValidateComplexity(plain string, minLen int) error
var ErrWeakPassword = errors.New("password does not meet complexity requirements: ...")
```

Rule: length must reach `minLen` (0 or negative treated as 1) and the password must contain at least 3 of the 4 classes (lowercase, uppercase, digit, symbol). Too-short input returns a dedicated length error; too-few-classes input returns `ErrWeakPassword`.

### Algorithm Overview

| Algorithm | Constructor | Built-in random salt | Typical use |
|-----------|-------------|----------------------|-------------|
| Bcrypt | `NewBCryptCrypto(cost ...int)` | Yes (built into the algorithm) | Password storage |
| Argon2id | `NewArgon2Crypto()` (tunable fields: Memory/Iterations/Parallelism/SaltLength/KeyLength) | Yes | Password storage (high security) |
| PBKDF2 | `NewPBKDF2Crypto()` (SHA-256, 310000 iters), `NewPBKDF2WithSHA512()` (SHA-512, 600000 iters) | Yes (16 bytes) | Password storage (compatibility) |
| SHA-256/512 | `NewSHA256Crypto()`, `NewSHA512Crypto()` | Yes (16 bytes) | Fast checksum scenarios |
| HMAC-SHA256 | `NewHMACCrypto(secretKey string)` | No (symmetric key) | Message authentication |
| RSA | `NewRSACrypto(keySize int)` + `Decrypt`/`ExportPrivateKey`/`ExportPublicKey` | OAEP internal randomness | Data encryption/decryption |
| ECDSA (P-256) | `NewECDSACrypto()` (`Encrypt` actually signs) | Randomized signatures | Sign/verify |
| ECDH (P-256) | `NewECDHCrypto()` + `DeriveSharedSecret(publicKey string) ([]byte, error)` | — | Key exchange |

> This package contains no AES or other symmetric block ciphers.

## Notes

- **Semantics differ**: for most implementations `Encrypt` is a hash/signature and is irreversible; only `RSACrypto` provides true reversible encryption/decryption; `ECDHCrypto.Encrypt` merely returns the local public key
- **ECDSA/ECDH are not password hashes**: the same plaintext yields a different `ECDSACrypto` signature each time (ECDSA signatures are randomized); `Verify` performs signature verification, not hash comparison
- Guard the `HMACCrypto` `secretKey`; once leaked, arbitrary MACs can be forged
- RSA `Encrypt` uses OAEP-SHA256, so plaintext length is bounded by the key size (~190 bytes for 2048-bit keys)
- Hash outputs embed their algorithm identifier and parameters (PBKDF2/Argon2/SHA); `Verify` recomputes using the stored parameters and compares in constant time, so old hashes remain verifiable after parameter upgrades
- For password storage prefer Argon2id or Bcrypt; use the SHA family only in performance-sensitive scenarios that accept fast-hash risks

## License

MIT
