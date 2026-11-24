package sfs2x

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// aesKeySize is the required AES key length (AES-128), matching the
// official SFS2X encryption scheme.
const aesKeySize = 16

// ivSize is the AES-CBC initialization-vector size.
const ivSize = 16

// blockSize mirrors aes.BlockSize (PKCS#7 padding unit).
const blockSize = 16

// CryptoKey is the official SFS2X encryption key pair: a 16-byte AES key
// and a 16-byte IV, both delivered by the server over HTTPS
// (/BlueBox/CryptoManager servlet, response = Base64(key || iv)) and reused
// for every encrypted frame.
type CryptoKey struct {
	Key []byte
	IV  []byte
}

// ParseCryptoKeyResponse parses the /BlueBox/CryptoManager servlet response:
// 32 raw bytes — key (16) followed by IV (16).
func ParseCryptoKeyResponse(data []byte) (*CryptoKey, error) {
	if len(data) != aesKeySize+ivSize {
		return nil, fmt.Errorf("crypto key response must be %d bytes (key||iv), got %d", aesKeySize+ivSize, len(data))
	}
	key := make([]byte, aesKeySize)
	iv := make([]byte, ivSize)
	copy(key, data[:aesKeySize])
	copy(iv, data[aesKeySize:])
	return &CryptoKey{Key: key, IV: iv}, nil
}

// ParseCryptoKeyResponseBase64 parses the servlet response as a Base64
// string (the servlet emits Base64 text over HTTPS).
func ParseCryptoKeyResponseBase64(s string) (*CryptoKey, error) {
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("crypto key response base64: %w", err)
	}
	return ParseCryptoKeyResponse(raw)
}

// Validate checks the key/IV lengths.
func (ck *CryptoKey) Validate() error {
	if ck == nil {
		return fmt.Errorf("nil crypto key")
	}
	if len(ck.Key) != aesKeySize {
		return fmt.Errorf("encryption key must be %d bytes, got %d", aesKeySize, len(ck.Key))
	}
	if len(ck.IV) != ivSize {
		return fmt.Errorf("encryption iv must be %d bytes, got %d", ivSize, len(ck.IV))
	}
	return nil
}

// encryptPayloadFixedIV AES-128-CBC encrypts with the official fixed
// key+IV pair: output = ciphertext only (PKCS#7 padded), no IV prefix —
// the peer decrypts with the same pair.
func encryptPayloadFixedIV(payload, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes: %w", err)
	}
	padded := pkcs7Pad(payload, blockSize)
	out := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, padded)
	return out, nil
}

// decryptPayloadFixedIV AES-128-CBC decrypts an official fixed-IV payload
// (ciphertext only, PKCS#7 unpadded).
func decryptPayloadFixedIV(payload, key, iv []byte) ([]byte, error) {
	if len(payload) == 0 || len(payload)%blockSize != 0 {
		return nil, fmt.Errorf("encrypted payload malformed: len %d", len(payload))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes: %w", err)
	}
	plain := make([]byte, len(payload))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, payload)
	return pkcs7Unpad(plain, blockSize)
}

// encryptPayload AES-128-CBC encrypts a frame payload with a fresh random
// IV, prefixing it (output = IV || ciphertext) so the frame is
// self-describing. The key must be exactly 16 bytes.
func encryptPayload(payload, key []byte) ([]byte, error) {
	if len(key) != aesKeySize {
		return nil, fmt.Errorf("encryption key must be %d bytes, got %d", aesKeySize, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes: %w", err)
	}

	iv := make([]byte, ivSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, fmt.Errorf("generate iv: %w", err)
	}

	padded := pkcs7Pad(payload, blockSize)
	out := make([]byte, ivSize+len(padded))
	copy(out, iv)

	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out[ivSize:], padded)
	return out, nil
}

// decryptPayload AES-128-CBC decrypts a frame payload: input = IV (16B) ||
// ciphertext, PKCS#7 unpadded. The key must be exactly 16 bytes.
func decryptPayload(payload, key []byte) ([]byte, error) {
	if len(key) != aesKeySize {
		return nil, fmt.Errorf("encryption key must be %d bytes, got %d", aesKeySize, len(key))
	}
	if len(payload) < ivSize+blockSize || len(payload)%blockSize != 0 {
		return nil, fmt.Errorf("encrypted payload malformed: len %d", len(payload))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes: %w", err)
	}

	iv := payload[:ivSize]
	ct := payload[ivSize:]

	plain := make([]byte, len(ct))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plain, ct)

	unpadded, err := pkcs7Unpad(plain, blockSize)
	if err != nil {
		return nil, err
	}
	return unpadded, nil
}

// pkcs7Pad pads data to a multiple of blockSize (RFC 5652 / PKCS#7):
// every pad byte equals the pad length; a full block is added when the
// input is already aligned.
func pkcs7Pad(data []byte, blockSize int) []byte {
	pad := blockSize - len(data)%blockSize
	out := make([]byte, len(data)+pad)
	copy(out, data)
	for i := len(data); i < len(out); i++ {
		out[i] = byte(pad)
	}
	return out
}

// pkcs7Unpad removes and validates PKCS#7 padding. A malformed pad is an
// error (never a panic) — with CBC this doubles as an integrity check for
// wrong-key garbage.
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, fmt.Errorf("invalid padded length %d", len(data))
	}
	pad := int(data[len(data)-1])
	if pad == 0 || pad > blockSize || pad > len(data) {
		return nil, fmt.Errorf("invalid pkcs7 padding")
	}
	for _, b := range data[len(data)-pad:] {
		if int(b) != pad {
			return nil, fmt.Errorf("invalid pkcs7 padding")
		}
	}
	return data[:len(data)-pad], nil
}
