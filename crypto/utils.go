package crypto

import "bytes"

// PKCS5Padding 填充明文
func PKCS5Padding(plaintext []byte, blockSize int) []byte {
	if blockSize <= 0 || blockSize > 255 {
		panic("blockSize must be in 1..255 for PKCS5Padding")
	}
	if len(plaintext) < 0 {
		panic("plaintext length invalid")
	}
	padding := blockSize - len(plaintext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, padText...)
}

// PKCS5UnPadding 去除填充数据
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	if length == 0 {
		return []byte{}
	}
	unpadding := int(origData[length-1])
	if unpadding == 0 || unpadding > length {
		return []byte{}
	}
	// 检查填充内容是否都等于 unpadding
	for i := length - unpadding; i < length; i++ {
		if int(origData[i]) != unpadding {
			return []byte{}
		}
	}
	return origData[:(length - unpadding)]
}
