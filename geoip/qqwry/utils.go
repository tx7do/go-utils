package qqwry

import (
	"bytes"
	"io"
	"regexp"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var regSpiltAddress = regexp.MustCompile(`.+?(省|市|自治区|自治州|盟|县|区|管委会|街道|镇|乡)`)

func gb18030Decode(src []byte) string {
	in := bytes.NewReader(src)
	out := transform.NewReader(in, simplifiedchinese.GB18030.NewDecoder())
	d, _ := io.ReadAll(out)
	return string(d)
}

// getMiddleOffset 取得begin和end之间的偏移量，用于二分搜索
func getMiddleOffset(start uint32, end uint32) uint32 {
	return start + (((end-start)/ipRecordLength)>>1)*ipRecordLength
}

func byte3ToUInt32(data []byte) uint32 {
	i := uint32(data[0]) & 0xff
	i |= (uint32(data[1]) << 8) & 0xff00
	i |= (uint32(data[2]) << 16) & 0xff0000
	return i
}

func SpiltAddress(addr string) []string {
	return regSpiltAddress.FindAllString(addr, -1)
}
