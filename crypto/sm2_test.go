package crypto

import (
	"bytes"
	"testing"
)

func TestSM2Cipher_EncryptDecrypt(t *testing.T) {
	cipher, err := NewSM2Cipher()
	if err != nil {
		t.Fatalf("Failed to generate SM2 key pair: %v", err)
	}

	plain := []byte("hello, sm2 cipher test!")
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

func TestSM2Cipher_EncryptDecryptAsn1(t *testing.T) {
	cipher, err := NewSM2Cipher()
	if err != nil {
		t.Fatalf("Failed to generate SM2 key pair: %v", err)
	}

	plain := []byte("hello, sm2 asn1 test!")
	crypted, err := cipher.EncryptAsn1(plain)
	if err != nil {
		t.Fatalf("EncryptAsn1 failed: %v", err)
	}
	if len(crypted) == 0 {
		t.Fatal("Encrypted ASN.1 result is empty")
	}

	decrypted, err := cipher.DecryptAsn1(crypted)
	if err != nil {
		t.Fatalf("DecryptAsn1 failed: %v", err)
	}
	if !bytes.Equal(plain, decrypted) {
		t.Fatalf("ASN.1 Decrypted text not match, got: %s, want: %s", decrypted, plain)
	}
}

func TestSM2Cipher_NilKey(t *testing.T) {
	cipher := &SM2Cipher{}
	_, err := cipher.Encrypt([]byte("test"))
	if err == nil {
		t.Fatal("Encrypt should fail with nil public key")
	}
	_, err = cipher.Decrypt([]byte("test"))
	if err == nil {
		t.Fatal("Decrypt should fail with nil private key")
	}
	_, err = cipher.EncryptAsn1([]byte("test"))
	if err == nil {
		t.Fatal("EncryptAsn1 should fail with nil public key")
	}
	_, err = cipher.DecryptAsn1([]byte("test"))
	if err == nil {
		t.Fatal("DecryptAsn1 should fail with nil private key")
	}
}

func TestSM2Cipher_SignVerify(t *testing.T) {
	cipher, err := NewSM2Cipher()
	if err != nil {
		t.Fatalf("Failed to generate SM2 key pair: %v", err)
	}
	data := []byte("hello, sm2 sign test!")
	sig, err := cipher.Sign(data)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}
	ok, err := cipher.Verify(data, sig)
	if err != nil || !ok {
		t.Fatalf("Verify failed, err: %v, ok: %v", err, ok)
	}
}
