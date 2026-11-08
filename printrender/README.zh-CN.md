# printrender

服务端 HTML → PDF 渲染：用本机 Chrome/Edge 浏览器（chromedp 驱动，自动探测安装
路径）把最终 HTML 文档打印成 PDF。

[English](README.md) | 简体中文

## 设计边界

打印/单据管线的典型分工：分页、毫米布局、字体由**前端渲染内核**完成后产出最终
HTML，本包只做最后一步 `PrintToPDF`，纸张宽高与方向由调用方按模板元数据传入。
预览与打印产物由此保证一致（同一 Chromium 引擎），后端不重排、不理解模板。
前端不参与打印的场景直接投喂任意完整 HTML 文档同样可用。

提供 `Renderer` 接口：默认实现以 chromedp 进程内驱动本机浏览器；部署侧如偏好
伴生渲染微服务（Playwright + Chromium 镜像），实现 `Renderer` 换入即可。

## 安装

printrender 是独立 Go 模块（`printrender/go.mod`）：

```bash
go get github.com/tx7do/go-utils/printrender
```

```go
import "github.com/tx7do/go-utils/printrender"
```

## 快速开始

```go
pdf, err := printrender.RenderPDF(ctx, printrender.Options{
	HTML:          finalHTML, // 完整 <html> 文档
	PaperWidthMm:  210,       // A4 纵向
	PaperHeightMm: 297,
	Background:    true,
	Timeout:       30 * time.Second,
})
if err != nil {
	return err
}
// pdf 即 PDF 文件内容（[]byte）
```

替换渲染器（如换成伴生服务）：

```go
printrender.SetRenderer(myRemoteRenderer{})
```

## API 参考

| 函数 | 说明 |
|------|------|
| `RenderPDF(ctx context.Context, opts Options) ([]byte, error)` | 用默认渲染器渲染 PDF（未设置时回落 `ChromeRenderer`） |
| `SetRenderer(r Renderer)` | 进程级替换默认渲染器 |
| `ChromeRenderer` | 本机 Chrome/Edge 渲染器；浏览器实例进程级复用 |

`Options`：

| 字段 | 说明 |
|------|------|
| `HTML string` | 最终文档（已完成分页），应为完整 `<html>` 文档 |
| `PaperWidthMm / PaperHeightMm float64` | 纸张尺寸（毫米），必填 |
| `Landscape bool` | 横向 |
| `Background bool` | 打印背景色/背景图（PrintToPDF 默认不打印背景） |
| `Timeout time.Duration` | 单次渲染超时，默认 60s |

## 浏览器探测

`CHROME_PATH` 环境变量 → `PATH` 上的 `chrome` → 常见 Chrome 安装路径 →
**Microsoft Edge 兜底**（Windows 服务器常无 Chrome；Edge 同 Chromium 协议）。

Linux 以 root 运行（容器常见）时自动追加 `--no-sandbox`（否则 Chrome 拒绝启动）；
设置 `PRINTRENDER_NO_SANDBOX=0` 可关闭该行为。

## 注意事项

- HTML 经 JSON 编码（`json.Marshal` 默认转义 `<`、`>`、`&`、U+2028/29）后经
  `document.write` 注入页面，任意 HTML 注入求值表达式都是安全的。
- allocator context 进程级创建复用；每次渲染新开标签页并在结束时关闭。
- 打印无网络/扩展需求；启动的浏览器禁用扩展与首启流程以收窄攻击面。
