package netutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateURL 合法 http/https URL 应通过校验。
func TestValidateURL_Valid(t *testing.T) {
	cases := []string{
		"http://example.com/path",
		"https://example.com:443/path?q=1",
		"https://8.8.8.8/",
		"http://example.com",
	}
	for _, u := range cases {
		t.Run(u, func(t *testing.T) {
			parsed, err := ValidateURL(u)
			assert.NoError(t, err)
			assert.NotNil(t, parsed)
		})
	}
}

// TestValidateURL_InvalidScheme 非 http/https scheme 应被拒。
func TestValidateURL_InvalidScheme(t *testing.T) {
	cases := []string{
		"file:///etc/passwd",
		"gopher://localhost",
		"ftp://example.com",
		"dict://localhost:6379",
		"javascript:alert(1)",
	}
	for _, u := range cases {
		t.Run(u, func(t *testing.T) {
			_, err := ValidateURL(u)
			assert.Error(t, err, "scheme %q 应被拒", u)
		})
	}
}

// TestValidateURL_EmptyURL 空 URL 应报错。
func TestValidateURL_EmptyURL(t *testing.T) {
	_, err := ValidateURL("")
	assert.Error(t, err)
}

// TestValidateURL_MissingHost 缺少 host 应报错。
func TestValidateURL_MissingHost(t *testing.T) {
	cases := []string{
		"http://",
		"https:///path",
	}
	for _, u := range cases {
		t.Run(u, func(t *testing.T) {
			_, err := ValidateURL(u)
			assert.Error(t, err, "缺少 host 的 URL %q 应被拒", u)
		})
	}
}

// TestValidateURL_UserinfoRejected 携带 userinfo 的 URL 应被拒（防 SSRF 混淆）。
func TestValidateURL_UserinfoRejected(t *testing.T) {
	cases := []string{
		"http://user:pass@example.com/",
		"http://admin@example.com/",
		"https://user%40host:pass@evil.com/",
	}
	for _, u := range cases {
		t.Run(u, func(t *testing.T) {
			_, err := ValidateURL(u)
			assert.Error(t, err, "携带 userinfo 的 URL %q 应被拒", u)
		})
	}
}

// TestValidateURL_Malformed 解析失败的 URL 应报错。
func TestValidateURL_Malformed(t *testing.T) {
	cases := []string{
		"http://example.com:badport",
		"://missing-scheme",
	}
	for _, u := range cases {
		t.Run(u, func(t *testing.T) {
			_, err := ValidateURL(u)
			assert.Error(t, err)
		})
	}
}
