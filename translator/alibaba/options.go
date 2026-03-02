package alibaba

// Option 是阿里翻译的配置选项
type Option func(*Translator)

// WithRegionID 设置阿里云区域 ID
// 支持的区域：cn-hangzhou, cn-shanghai, cn-beijing, ap-southeast-1 等
func WithRegionID(regionID string) Option {
	return func(t *Translator) {
		t.regionID = regionID
	}
}
