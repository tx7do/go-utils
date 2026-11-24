package sfs2x

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

// aesKeySize is the required AES key length (AES-128), matching the
// official SFS2X encryption scheme.
const aesKeySize = 16

// ivSize is the AES-CBC initialization-vector size, prefixed to every
// encrypted payload.
const ivSize = 16

// blockSize mirrors aes.BlockSize (PKCS#7 padding unit).
const blockSize = 16

// encryptPayload AES-128-CBC encrypts a frame payload (header bit 0x40):
// output = random IV (16B) || ciphertext, padded with PKCS#7. The key must
// be exactly 16 bytes.
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
