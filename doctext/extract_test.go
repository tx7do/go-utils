package doctext

import (
	"archive/zip"
	"bytes"
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

// ---- 纯文本族 ----

func TestExtractPlainTextFamily(t *testing.T) {
	for _, ext := range []string{".txt", ".md", ".markdown", ".csv", ".log", ".json", ".xml", ".yml", ".yaml", ".html", ".htm"} {
		text, err := Extract("doc"+ext, []byte("hello 世界"))
		require.NoError(t, err, ext)
		require.Equal(t, "hello 世界", text, ext)
	}
}

func TestExtractNormalizesCRLFAndSanitizesUTF8(t *testing.T) {
	text, err := Extract("a.txt", []byte("l1\r\nl2"))
	require.NoError(t, err)
	require.Equal(t, "l1\nl2", text)

	// 无效 UTF-8 字节替换为 U+FFFD，有效部分保留
	text, err = Extract("b.txt", []byte("ok\xff\xfe"))
	require.NoError(t, err)
	require.Contains(t, text, "ok")
	require.Contains(t, text, "\uFFFD")
}

func TestExtractRejectsEmptyBlankOversize(t *testing.T) {
	_, err := Extract("a.txt", nil)
	require.ErrorContains(t, err, "file is empty")

	_, err = Extract("a.txt", []byte(" \t\n \r\n "))
	require.ErrorContains(t, err, "no text content extracted")

	_, err = Extract("a.pdf", bytes.Repeat([]byte("a"), MaxFileSize+1))
	require.ErrorContains(t, err, "file too large")
}

func TestExtractRejectsUnsupportedExtension(t *testing.T) {
	// 扩展名大小写不敏感（extract 内部小写化）：.EXE 仍落在拒绝分支
	for _, name := range []string{"a.exe", "b.doc", "noext", "c.pdf.txtx", "d.EXE"} {
		_, err := Extract(name, []byte("hello"))
		require.ErrorContains(t, err, "unsupported file type", name)
	}
}

func TestSupportedExtensions(t *testing.T) {
	require.True(t, slices.Equal([]string{
		".txt", ".md", ".markdown", ".csv", ".log", ".json", ".xml", ".yml", ".yaml", ".html", ".htm", ".docx", ".pdf",
	}, SupportedExtensions()))
}

// ---- docx ----

func buildDocx(t *testing.T, members map[string]string) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range members {
		w, err := zw.Create(name)
		require.NoError(t, err)
		_, err = w.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())
	return buf.Bytes()
}

func TestExtractDocx(t *testing.T) {
	// 覆盖：非 document 成员被跳过、段落边界与 br/tab 转可读边界、
	// XML 标签全剥、实体反转义。组合后的精确输出被钉死。
	docXML := "<w:document><w:body>" +
		"<w:p>first</w:p><w:p>second</w:p>" +
		"<w:p>l1<w:br/>l2</w:p>" +
		"<w:p>x<w:tab/>y</w:p>" +
		"<w:p>a &amp; b</w:p>" +
		"</w:body></w:document>"
	data := buildDocx(t, map[string]string{
		"[Content_Types].xml": "<?xml version=\"1.0\"?><Types/>",
		"word/document.xml":   docXML,
	})
	text, err := Extract("a.docx", data)
	require.NoError(t, err)
	require.Equal(t, "first\nsecond\nl1\nl2\nx\ty\na & b", text)
}

func TestExtractDocxMissingDocument(t *testing.T) {
	data := buildDocx(t, map[string]string{"word/settings.xml": "<settings/>"})
	_, err := Extract("a.docx", data)
	require.ErrorContains(t, err, "word/document.xml not found in docx")
}

func TestExtractDocxCorruptZip(t *testing.T) {
	_, err := Extract("a.docx", []byte("this is not a zip at all"))
	require.ErrorContains(t, err, "docx extract")
	require.ErrorContains(t, err, "open zip")
}

func TestExtractDocxNoText(t *testing.T) {
	data := buildDocx(t, map[string]string{"word/document.xml": "<w:p></w:p>"})
	_, err := Extract("a.docx", data)
	require.ErrorContains(t, err, "no text content extracted from docx")
}

// ---- pdf ----

// buildPdf 构造一个最小合法的单页 PDF，contentStream 为页面内容流。
// 传文本绘制算子得到带文本层的页面；传路径算子得到无文本层页面
// （扫描件等价物）。xref 偏移按实际字节位置计算，保证结构合法。
func buildPdf(contentStream string) []byte {
	var b bytes.Buffer
	offsets := make([]int, 6)
	b.WriteString("%PDF-1.4\n")
	obj := func(n int, body string) {
		offsets[n] = b.Len()
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", n, body)
	}
	obj(1, "<< /Type /Catalog /Pages 2 0 R >>")
	obj(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	obj(3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>")
	obj(4, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(contentStream), contentStream))
	obj(5, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	xref := b.Len()
	b.WriteString("xref\n0 6\n0000000000 65535 f \n")
	for i := 1; i <= 5; i++ {
		fmt.Fprintf(&b, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&b, "trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xref)
	return b.Bytes()
}

func TestExtractPdfTextLayer(t *testing.T) {
	stream := "BT\n/F1 12 Tf\n72 720 Td\n(HELLO PDF) Tj\nET"
	text, err := Extract("a.pdf", buildPdf(stream))
	require.NoError(t, err)
	require.Contains(t, text, "HELLO PDF")
}

func TestExtractPdfNoTextLayer(t *testing.T) {
	// 只有路径绘制算子：无文本层（扫描件等价物）必须报错而非返回空串成功
	stream := "1 1 m\n10 10 l\nS"
	_, err := Extract("a.pdf", buildPdf(stream))
	require.ErrorContains(t, err, "no text layer")
}

func TestExtractPdfCorrupt(t *testing.T) {
	_, err := Extract("a.pdf", []byte("definitely not a pdf"))
	require.ErrorContains(t, err, "pdf extract")
}
