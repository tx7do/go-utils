// Package doctext 从常见文档格式中抽取纯文本，供知识库 RAG 入库使用。
//
// 支持格式：纯文本族（txt/md/csv/log/json/xml/yaml/html）、docx（zip+XML 标准库解析）、
// pdf（ledongthuc/pdf 纯 Go 抽取，扫描件/图片型 PDF 无文本层会返回空并报错）。
// 其余扩展名直接拒绝——宁可报"不支持"也不要把二进制乱码灌进向量库。
package doctext

import (
	"archive/zip"
	"bytes"
	"fmt"
	"html"
	"io"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

// MaxFileSize 上传文件大小上限（10MB）。
const MaxFileSize = 10 << 20

// SupportedExtensions 返回支持的扩展名清单（管理页提示用）。
func SupportedExtensions() []string {
	return []string{".txt", ".md", ".markdown", ".csv", ".log", ".json", ".xml", ".yml", ".yaml", ".html", ".htm", ".docx", ".pdf"}
}

// Extract 按扩展名抽取纯文本。
func Extract(fileName string, data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("file is empty")
	}
	if len(data) > MaxFileSize {
		return "", fmt.Errorf("file too large: %d bytes (max %d)", len(data), MaxFileSize)
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".txt", ".md", ".markdown", ".csv", ".log", ".json", ".xml", ".yml", ".yaml", ".html", ".htm":
		text := sanitizeText(string(data))
		if strings.TrimSpace(text) == "" {
			return "", fmt.Errorf("no text content extracted")
		}
		return text, nil
	case ".docx":
		text, err := extractDocx(data)
		if err != nil {
			return "", fmt.Errorf("docx extract: %w", err)
		}
		return text, nil
	case ".pdf":
		text, err := extractPdf(data)
		if err != nil {
			return "", fmt.Errorf("pdf extract: %w", err)
		}
		return text, nil
	default:
		return "", fmt.Errorf("unsupported file type %q (supported: %s)", ext, strings.Join(SupportedExtensions(), " "))
	}
}

// sanitizeText 清洗纯文本：强制有效 UTF-8、规范化换行。
func sanitizeText(s string) string {
	s = strings.ToValidUTF8(s, string([]byte{0xEF, 0xBF, 0xBD})) // U+FFFD
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return s
}

// extractDocx 解 docx（zip 容器）：word/document.xml 里把段落边界换成换行、剥全部 XML 标签、反转义实体。
func extractDocx(data []byte) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("open zip: %w", err)
	}

	const docXML = "word/document.xml"
	var xmlContent []byte
	for _, file := range reader.File {
		if file.Name != docXML {
			continue
		}
		f, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("open %s: %w", docXML, err)
		}
		xmlContent, err = io.ReadAll(f)
		_ = f.Close()
		if err != nil {
			return "", fmt.Errorf("read %s: %w", docXML, err)
		}
		break
	}
	if xmlContent == nil {
		return "", fmt.Errorf("%s not found in docx", docXML)
	}

	// 段落/换行/制表转成可读边界，再剥标签
	s := string(xmlContent)
	s = strings.ReplaceAll(s, "</w:p>", "\n")
	s = strings.ReplaceAll(s, "<w:br/>", "\n")
	s = strings.ReplaceAll(s, "<w:tab/>", "\t")

	var sb strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '<' {
			if j := strings.IndexByte(s[i:], '>'); j >= 0 {
				i += j + 1
				continue
			}
		}
		sb.WriteByte(s[i])
		i++
	}
	text := sanitizeText(html.UnescapeString(sb.String()))
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("no text content extracted from docx")
	}
	return text, nil
}

// extractPdf 抽取 PDF 文本层（非 OCR：扫描件无文本层会得到空内容）。
func extractPdf(data []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("open pdf: %w", err)
	}

	var sb strings.Builder
	pages := reader.NumPage()
	for i := 1; i <= pages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue // 单页抽取失败不阻断整本
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}
	text := sanitizeText(strings.TrimSpace(sb.String()))
	if strings.TrimSpace(strings.ReplaceAll(text, "\n", "")) == "" {
		return "", fmt.Errorf("no text layer found (scanned/image pdf is not supported)")
	}
	return text, nil
}
