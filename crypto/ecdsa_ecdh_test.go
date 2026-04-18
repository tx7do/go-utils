package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestECDSACipher_SignVerify(t *testing.T) {
	cipher, err := NewECDSACipher()
	assert.NoError(t, err)
	data := []byte("ecdsa sign test data")

	sig, err := cipher.Sign(data)
	assert.NoError(t, err)
	ok, err := cipher.Verify(data, sig)
	assert.NoError(t, err)
	assert.True(t, ok, "ECDSA signature should verify")

	// tampered data
	ok, err = cipher.Verify([]byte("tampered"), sig)
	assert.NoError(t, err)
	assert.False(t, ok, "ECDSA signature should not verify for tampered data")
}

func TestECDSACipher_SignVerify_Fail(t *testing.T) {
	cipher, _ := NewECDSACipher()
	data := []byte("ecdsa fail test")
	sig, _ := cipher.Sign(data)
	ok, err := cipher.Verify([]byte("other data"), sig)
	if err != nil {
		t.Fatalf("Verify error: %v", err)
	}
	if ok {
		t.Fatalf("Verify should fail for wrong data")
	}
}

func TestECDHCipher_KeyExchange(t *testing.T) {
	alice, err := NewECDHCipher()
	assert.NoError(t, err)
	bob, err := NewECDHCipher()
	assert.NoError(t, err)

	alicePub := alice.PublicKeyBytes()
	bobPub := bob.PublicKeyBytes()

	aliceSecret, err := alice.DeriveSharedSecret(bobPub)
	assert.NoError(t, err)
	bobSecret, err := bob.DeriveSharedSecret(alicePub)
	assert.NoError(t, err)

	assert.Equal(t, aliceSecret, bobSecret, "ECDH shared secrets should match")
}

func TestECDHCipher_InvalidPeerKey(t *testing.T) {
	alice, _ := NewECDHCipher()
	_, err := alice.DeriveSharedSecret([]byte("invalid"))
	if err == nil {
		t.Fatalf("Should fail for invalid peer public key")
	}
}
