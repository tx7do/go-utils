package crypto

import (
	"fmt"
	"testing"

	"encoding/base64"

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
