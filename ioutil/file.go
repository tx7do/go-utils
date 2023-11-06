package ioutil

import "os"

// ReadFile 读取文件
func ReadFile(path string) []byte {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return content
}
