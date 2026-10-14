package sms

import (
	"fmt"
	"strings"
)

// NormalizePhoneNumber 归一手机号为国内裸号与 E.164 两形。
//
// 接受两类输入：E.164（可带 +86）与中国大陆 11 位裸号（1 开头，自动补 +86）。
// 其他形态一律拒绝——短信目标错了没有"发到别家去"的软失败，宁可让这一次投递
// 失败在调用方的记录里说清楚。分隔符（空格/横线/括号）先剥掉，用户从通讯录粘出来的号常带。
//
// national（无 +86 前缀）与 e164（带 +）两形都给：阿里云国内短信收裸号，
// 腾讯云 PhoneNumberSet 只认 E.164。调用方可以用它做入口处的快反馈，
// Send 在投递前还会再跑一遍作最终裁决。
func NormalizePhoneNumber(raw string) (national, e164 string, err error) {
	cleaned := make([]byte, 0, len(raw))
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if c == ' ' || c == '-' || c == '(' || c == ')' {
			continue
		}
		cleaned = append(cleaned, c)
	}
	s := string(cleaned)
	if s == "" {
		return "", "", fmt.Errorf("phone number %q is empty", raw)
	}

	// 位数校验对去掉前导 + 的主体做：+ 是格式记号，不是数字。
	body := strings.TrimPrefix(s, "+")
	digitCount := 0
	for i := 0; i < len(body); i++ {
		if body[i] >= '0' && body[i] <= '9' {
			digitCount++
		}
	}
	if digitCount != len(body) || digitCount < 5 || digitCount > 20 {
		return "", "", fmt.Errorf("phone number %q is not a valid MSISDN (5-20 digits, optional +86 prefix)", raw)
	}

	switch {
	case strings.HasPrefix(s, "+86"):
		return s[3:], s, nil
	case s[0] == '+':
		return s[1:], s, nil
	case digitCount == 11 && s[0] == '1':
		return s, "+86" + s, nil
	default:
		return "", "", fmt.Errorf("phone number %q must be E.164 (with country code) or a CN mobile number", raw)
	}
}
