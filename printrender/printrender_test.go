package printrender

// 纯逻辑单测：入参校验、毫米→英寸换算、超时缺省、SetRenderer 替换语义。
// chromedp 真实打印路径需要浏览器，不在单测覆盖范围（部署侧联调验证）。

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateOptions(t *testing.T) {
	valid := Options{HTML: "<html></html>", PaperWidthMm: 210, PaperHeightMm: 297}
	require.NoError(t, validateOptions(valid))

	assert.Error(t, validateOptions(Options{PaperWidthMm: 210, PaperHeightMm: 297}),
		"空 HTML 必须拒绝")
	assert.Error(t, validateOptions(Options{HTML: "<html></html>"}),
		"缺纸张尺寸必须拒绝")
	assert.Error(t, validateOptions(Options{HTML: "<html></html>", PaperWidthMm: -1, PaperHeightMm: 297}),
		"负纸张尺寸必须拒绝")
	assert.Error(t, validateOptions(Options{HTML: "<html></html>", PaperWidthMm: 210, PaperHeightMm: 0}),
		"零纸张尺寸必须拒绝")
}

func TestPaperSizeInInches(t *testing.T) {
	w, h := paperSizeInInches(Options{PaperWidthMm: 210, PaperHeightMm: 297})
	assert.InDelta(t, 8.2677, w, 0.001, "A4 宽 210mm ≈ 8.268in")
	assert.InDelta(t, 11.6929, h, 0.001, "A4 高 297mm ≈ 11.693in")

	w, h = paperSizeInInches(Options{PaperWidthMm: 80, PaperHeightMm: 100})
	assert.InDelta(t, 3.1496, w, 0.001, "小票宽 80mm ≈ 3.150in")
	assert.InDelta(t, 3.937, h, 0.001, "小票高 100mm ≈ 3.937in")
}

func TestRenderTimeout(t *testing.T) {
	assert.Equal(t, 60*time.Second, renderTimeout(Options{}), "零值缺省 60s")
	assert.Equal(t, 60*time.Second, renderTimeout(Options{Timeout: -1}), "负值缺省 60s")
	assert.Equal(t, 5*time.Second, renderTimeout(Options{Timeout: 5 * time.Second}),
		"显式超时优先")
}

// stubRenderer 记录调用并返回固定结果,验证 SetRenderer 的替换语义。
type stubRenderer struct {
	called bool
}

func (s *stubRenderer) RenderPDF(_ context.Context, opts Options) ([]byte, error) {
	s.called = true
	return []byte(opts.HTML), nil
}

func TestSetRendererOverride(t *testing.T) {
	t.Cleanup(func() { SetRenderer(nil) })

	stub := &stubRenderer{}
	SetRenderer(stub)
	pdf, err := RenderPDF(t.Context(), Options{HTML: "hi", PaperWidthMm: 210, PaperHeightMm: 297})
	require.NoError(t, err)
	assert.True(t, stub.called, "替换后的渲染器必须被使用")
	assert.Equal(t, []byte("hi"), pdf)

	SetRenderer(nil)
	// 回落默认 ChromeRenderer:本机有浏览器则真实渲染(非空 PDF),
	// 没有则报 findBrowser 错误;两种结果都证明 stub 已被卸载
	// (不再返回 stub 的"原样回显 HTML")。
	pdf, err = RenderPDF(t.Context(), Options{HTML: "<html></html>", PaperWidthMm: 210, PaperHeightMm: 297})
	if err == nil {
		assert.NotEmpty(t, pdf, "默认渲染器应产出真实 PDF 而非回显 HTML")
		assert.NotEqual(t, []byte("<html></html>"), pdf)
	} else {
		assert.Contains(t, err.Error(), "printrender", "错误应来自渲染器自身")
	}
}
