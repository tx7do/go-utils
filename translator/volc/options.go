package volc

import "net/http"

// Option 是火山翻译的配置选项
type Option func(*Translator)

// WithRegion 设置火山引擎区域
// 支持的区域：cn-beijing（默认）, cn-shanghai, ap-singapore 等
func WithRegion(region string) Option {
	return func(t *Translator) {
		t.region = region
	}
}

// WithHTTPClient 自定义HTTP客户端（扩展配置）
func WithHTTPClient(client *http.Client) Option {
	return func(t *Translator) {
		t.client = client
	}
}
