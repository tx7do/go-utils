// Copyright 2022 The Ip2Region Authors. All rights reserved.
// Use of this source code is governed by a Apache2.0-style
// license that can be found in the LICENSE file.

// ---
// @Author Lion <chenxin619315@gmail.com>
// @Date   2022/06/16

package xdb

import (
	"bytes"
	"fmt"
	"net"
)

func ParseIP(ip string) ([]byte, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, fmt.Errorf("invalid ip address: %s", ip)
	}

	v4 := parsedIP.To4()
	if v4 != nil {
		return v4, nil
	}

	v6 := parsedIP.To16()
	if v6 != nil {
		return v6, nil
	}

	return nil, fmt.Errorf("invalid ip address: %s", ip)
}

func IP2String(ip []byte) string {
	return net.IP(ip[:]).String()
}

// IPCompare compares two IP addresses
// Returns: -1 if ip1 < ip2, 0 if ip1 == ip2, 1 if ip1 > ip2
func IPCompare(ip1, ip2 []byte) int {
	// for i := 0; i < len(ip1); i++ {
	// 	if ip1[i] < ip2[i] {
	// 		return -1
	// 	}
	// 	if ip1[i] > ip2[i] {
	// 		return 1
	// 	}
	// }
	// return 0
	return bytes.Compare(ip1, ip2)
}

func IPAddOne(ip []byte) []byte {
	var r = make([]byte, len(ip))
	copy(r, ip)
	for i := len(ip) - 1; i >= 0; i-- {
		r[i]++
		if r[i] != 0 { // No overflow
			break
		}
	}

	return r
}

func IPSubOne(ip []byte) []byte {
	var r = make([]byte, len(ip))
	copy(r, ip)
	for i := len(ip) - 1; i >= 0; i-- {
		if r[i] != 0 { // No borrow needed
			r[i]--
			break
		}
		r[i] = 0xFF // borrow from the next byte
	}

	return r
}

// LoadHeaderFromBuff wrap the header info from the content buffer
func LoadHeaderFromBuff(cBuff []byte) (*Header, error) {
	return NewHeader(cBuff[0:HeaderInfoLength])
}

func LoadVectorIndexFromBuff(cBuff []byte) ([]byte, error) {
	needed := HeaderInfoLength + VectorIndexRows*VectorIndexCols*VectorIndexSize
	if len(cBuff) < needed {
		return nil, fmt.Errorf("buffer too small: need %d bytes, got %d", needed, len(cBuff))
	}

	offset := HeaderInfoLength
	total := VectorIndexRows * VectorIndexCols * VectorIndexSize
	buff := make([]byte, total)
	copy(buff, cBuff[offset:offset+total])

	return buff, nil
}
