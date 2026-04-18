package crypto

import (
	"bytes"
	"testing"
)

func TestSM4Cipher_EncryptDecrypt(t *testing.T) {
	key, err := GenerateSM4Key()
	if err != nil {
		t.Fatalf("Failed to generate SM4 key: %v", err)
	}
	cipher, err := NewSM4Cipher(key)
	if err != nil {
		t.Fatalf("Failed to create SM4Cipher: %v", err)
	}

	plain := []byte("hello, sm4 cipher test!")
	crypted, err := cipher.Encrypt(plain)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if len(crypted) == 0 {
		t.Fatal("Encrypted result is empty")
	}

	decrypted, err := cipher.Decrypt(crypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if !bytes.Equal(plain, decrypted) {
		t.Fatalf("Decrypted text not match, got: %s, want: %s", decrypted, plain)
	}
}

func TestSM4Cipher_InvalidKey(t *testing.T) {
	key := make([]byte, 8) // invalid key
	_, err := NewSM4Cipher(key)
	if err == nil {
		t.Fatal("Should fail with invalid key length")
	}
}

func TestSM4Cipher_EmptyPlaintext(t *testing.T) {
	key, _ := GenerateSM4Key()
	cipher, _ := NewSM4Cipher(key)
	crypted, err := cipher.Encrypt([]byte{})
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	decrypted, err := cipher.Decrypt(crypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if len(decrypted) != 0 {
		t.Fatalf("Decrypted empty plaintext should be empty, got: %v", decrypted)
	}
}
