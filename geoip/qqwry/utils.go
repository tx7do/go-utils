package qqwry

import (
	"bytes"
	"io"
	"net"
	"regexp"
	"strings"

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

// IsPrivateIP 判断 IP 是否为内网地址
func IsPrivateIP(ipStr string) bool {
	// 处理 IPv6 或无效 IP（qqwry 主要支持 IPv4，可直接返回 true 视为内网）
	ip := net.ParseIP(ipStr)
	if ip == nil || strings.Contains(ipStr, ":") {
		return true
	}

	// 定义内网网段（IPv4）
	privateCIDRs := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",    // 本地回环地址
		"169.254.0.0/16", // 链路本地地址
	}

	for _, cidr := range privateCIDRs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}
