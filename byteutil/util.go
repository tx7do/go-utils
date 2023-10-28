package byteutil

import (
	"bytes"
	"encoding/binary"
)

// IntToBytes 将int转换为[]byte
func IntToBytes(n int) []byte {
	data := int64(n)
	byteBuf := bytes.NewBuffer([]byte{})
	_ = binary.Write(byteBuf, binary.BigEndian, data)
	return byteBuf.Bytes()
}

// BytesToInt 将[]byte转换为int
func BytesToInt(bys []byte) int {
	byteBuf := bytes.NewBuffer(bys)
	var data int64
	_ = binary.Read(byteBuf, binary.BigEndian, &data)
	return int(data)
}

// ByteToLower lowers a byte
func ByteToLower(b byte) byte {
	if b <= '\u007F' {
		if 'A' <= b && b <= 'Z' {
			b += 'a' - 'A'
		}
		return b
	}
	return b
}

// ByteToUpper upper a byte
func ByteToUpper(b byte) byte {
	if b <= '\u007F' {
		if 'a' <= b && b <= 'z' {
			b -= 'a' - 'A'
		}
		return b
	}
	return b
}
