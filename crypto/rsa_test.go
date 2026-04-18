package crypto

import (
	"bytes"
	"testing"
)

func TestRSACipher_EncryptDecrypt(t *testing.T) {
	cipher, err := NewRSACipher(2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key pair: %v", err)
	}

	plain := []byte("hello, rsa cipher test!")
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

func TestRSACipher_NilKey(t *testing.T) {
	cipher := &RSACipher{}
	_, err := cipher.Encrypt([]byte("test"))
	if err == nil {
		t.Fatal("Encrypt should fail with nil public key")
	}
	_, err = cipher.Decrypt([]byte("test"))
	if err == nil {
		t.Fatal("Decrypt should fail with nil private key")
	}
}

func TestRSACipher_ExportKey(t *testing.T) {
	cipher, err := NewRSACipher(2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key pair: %v", err)
	}
	priv, err := cipher.ExportPrivateKey()
	if err != nil || len(priv) == 0 {
		t.Fatalf("ExportPrivateKey failed: %v", err)
	}
	pub, err := cipher.ExportPublicKey()
	if err != nil || len(pub) == 0 {
		t.Fatalf("ExportPublicKey failed: %v", err)
	}
}
