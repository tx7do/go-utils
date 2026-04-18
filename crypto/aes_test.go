package crypto

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecryptAES(t *testing.T) {
	//key的长度必须是16、24或者32字节，分别用于选择AES-128, AES-192, or AES-256
	aesKey, _ := GenerateAESKey(16)
	aesKey = DefaultAESKey

	plainText := []byte("cloud123456")
	encryptText, err := AesEncrypt(plainText, aesKey, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	pass64 := base64.StdEncoding.EncodeToString(encryptText)
	fmt.Printf("加密后:%v\n", pass64)

	bytesPass, err := base64.StdEncoding.DecodeString(pass64)
	if err != nil {
		fmt.Println(err)
		return
	}

	decryptText, err := AesDecrypt(bytesPass, aesKey, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("解密后:%s\n", decryptText)
	assert.Equal(t, plainText, decryptText)
}

func TestGenerateAESKey_ValidLengths(t *testing.T) {
	lengths := []int{16, 24, 32}
	for _, length := range lengths {
		key, err := GenerateAESKey(length)
		assert.NoError(t, err)
		assert.Equal(t, length, len(key))
		t.Logf("%d : %x\n", length, string(key))
	}
}

func TestPKCS5UnPadding_InvalidCases(t *testing.T) {
	cases := [][]byte{
		{},           // 空
		{1, 2, 3, 0}, // unpadding=0
		{1, 2, 3, 5}, // unpadding>len
		{1, 2, 3, 2}, // 填充内容不一致
		{4, 4, 4, 4}, // 全填充，合法，返回空
		{1, 2, 3, 1}, // 合法，返回{1,2,3}
	}
	expects := [][]byte{
		{}, {}, {}, {}, {}, {1, 2, 3},
	}
	for i, c := range cases {
		out := PKCS5UnPadding(c)
		assert.Equal(t, expects[i], out, "case %d failed", i)
	}
}

func TestPKCS5Padding_InvalidCases(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("should panic for invalid blockSize")
		}
	}()
	_ = PKCS5Padding([]byte("abc"), 0)
}

func TestPKCS5Padding_Normal(t *testing.T) {
	b := []byte("abc")
	padded := PKCS5Padding(b, 8)
	assert.Equal(t, []byte{'a', 'b', 'c', 5, 5, 5, 5, 5}, padded)
}

func TestAesDecrypt_InvalidCases(t *testing.T) {
	key := []byte("1234567890abcdef")
	iv := []byte("1234567890abcdef")
	_, err := AesDecrypt(nil, key, iv)
	assert.Error(t, err)
	_, err = AesDecrypt([]byte{}, key, iv)
	assert.Error(t, err)
	_, err = AesDecrypt([]byte("1234"), []byte("short"), iv)
	assert.Error(t, err)
	_, err = AesDecrypt([]byte("1234"), key, []byte("shortiv"))
	assert.Error(t, err)
	_, err = AesDecrypt([]byte("1234"), key, iv)
	assert.Error(t, err) // 密文长度不是blockSize倍数
}

func TestAESCipher_EncryptDecrypt(t *testing.T) {
	key := make([]byte, 16)
	iv := make([]byte, 16)
	_, _ = rand.Read(key)
	_, _ = rand.Read(iv)

	cipher := NewAESCipher(key, iv)
	plain := []byte("hello, aes cipher test!")

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

func TestAESCipher_EmptyPlaintext(t *testing.T) {
	key := make([]byte, 16)
	iv := make([]byte, 16)
	_, _ = rand.Read(key)
	_, _ = rand.Read(iv)

	cipher := NewAESCipher(key, iv)
	_, err := cipher.Encrypt([]byte{})
	if err == nil {
		t.Fatal("Encrypt should fail on empty plaintext")
	}
}

func TestAESCipher_InvalidKeyIV(t *testing.T) {
	key := make([]byte, 8) // invalid key
	iv := make([]byte, 16)
	cipher := NewAESCipher(key, iv)
	_, err := cipher.Encrypt([]byte("test"))
	if err == nil {
		t.Fatal("Encrypt should fail on invalid key length")
	}
}
