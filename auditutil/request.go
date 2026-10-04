package auditutil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/tx7do/go-utils/id"
)

// RequestID 获取请求 ID：依次检查 X-Request-ID / X-Correlation-ID /
// x-fc-request-id（函数计算），都没有时生成新的 GUID。
func RequestID(request *http.Request) string {
	if request == nil {
		return ""
	}

	for _, key := range []string{headerXRequestID, headerXCorrelationID, headerXFcRequestID} {
		if v := request.Header.Get(key); v != "" {
			return v
		}
	}

	return id.NewGUIDv4(false)
}

var reUsername = regexp.MustCompile(`"username"\s*:\s*"([^"]+)"`)

// ParseUsernameFromBytes 从请求体中解析用户名（先按 JSON 字段、再按表单键）。
// 返回前剥离 CR/LF，防止含换行的用户名注入文本日志行（伪造条目/行内注入）。
func ParseUsernameFromBytes(body []byte) (string, error) {
	if m := reUsername.FindSubmatch(body); m != nil {
		return StripLineBreaks(string(m[1])), nil
	}
	if values, err := url.ParseQuery(string(body)); err == nil {
		if u := values.Get("username"); u != "" {
			return StripLineBreaks(u), nil
		}
	}
	return "", fmt.Errorf("username not found")
}

// StripLineBreaks 移除所有 CR/LF，阻断日志行注入。
func StripLineBreaks(s string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(s)
}

// ExtractUsernameFromRequest 从 HTTP 请求体中提取用户名。
// 会读取并恢复 Body，避免影响后续中间件与 handler 的读取。
func ExtractUsernameFromRequest(r *http.Request) (username string, err error) {
	if r == nil || r.Body == nil {
		return "", fmt.Errorf("nil request")
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	_ = r.Body.Close()
	// 恢复 Body，避免影响后续处理
	r.Body = io.NopCloser(bytes.NewReader(body))

	if username, err = ParseUsernameFromBytes(body); err == nil {
		return username, nil
	}

	if values, err := url.ParseQuery(string(body)); err == nil {
		return values.Get("username"), nil
	}

	return "", err
}
