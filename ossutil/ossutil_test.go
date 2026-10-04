package ossutil

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentTypeToBucketName(t *testing.T) {
	cases := map[string]string{
		"image/png":                                    "images",
		"video/mp4; charset=bogus":                     "videos",
		"audio/mpeg":                                   "audios",
		"text/plain":                                   "docs",
		"application/pdf":                              "docs",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": "docs",
		"application/json":                            "docs",
		"application/octet-stream":                    "files",
		"":                                            "files",
		"image/":                                      "files", // 畸形串不按主类型落图片桶
		"not-a-mime":                                  "files",
	}
	for ct, want := range cases {
		assert.Equal(t, want, ContentTypeToBucketName(ct), "ContentTypeToBucketName(%q)", ct)
	}
}

func TestFileExtensionToBucketName(t *testing.T) {
	assert.Equal(t, "images", FileExtensionToBucketName(".jpg"))
	assert.Equal(t, "images", FileExtensionToBucketName("jpg"))
	assert.Equal(t, "videos", FileExtensionToBucketName(".mp4"))
	assert.Equal(t, "docs", FileExtensionToBucketName(".pdf"))
	assert.Equal(t, "files", FileExtensionToBucketName(".zip"))
	assert.Equal(t, "files", FileExtensionToBucketName(""))
}

func TestContentTypeToFileExtension(t *testing.T) {
	assert.Equal(t, "jpg", ContentTypeToFileExtension("image/jpeg"))
	assert.Equal(t, "png", ContentTypeToFileExtension("image/png"))
	assert.Equal(t, "pdf", ContentTypeToFileExtension("application/pdf"))
	assert.Equal(t, "", ContentTypeToFileExtension(""))
}

func TestDetectFileType(t *testing.T) {
	png := []byte("\x89PNG\r\n\x1a\n" + strings.Repeat("\x00", 16))
	mt, ext := DetectFileType(png)
	assert.Equal(t, "image/png", mt)
	assert.Equal(t, ".png", ext)

	jpg := []byte{0xff, 0xd8, 0xff, 0xe0}
	mt, ext = DetectFileType(jpg)
	assert.Equal(t, "image/jpeg", mt)
	assert.Equal(t, ".jpg", ext)
}

func TestGenerateFileName(t *testing.T) {
	content := []byte("hello world")

	uuidName := GenerateFileName(content, ".png", GenerateFileNameTypeUUID)
	assert.NotEmpty(t, uuidName)
	assert.True(t, strings.HasSuffix(uuidName, ".png"))
	assert.NotEqual(t, uuidName, GenerateFileName(content, ".png", GenerateFileNameTypeUUID))

	sha1 := GenerateFileName(content, "", GenerateFileNameTypeContentSHA256)
	sha2 := GenerateFileName(content, "", GenerateFileNameTypeContentSHA256)
	assert.Equal(t, sha1, sha2, "内容哈希命名必须确定")

	// 未注入密钥：HMAC 策略退化为 UUID 命名（32 位十六进制、无扩展名）
	hmacFallback := GenerateFileName(content, "", GenerateFileNameTypeHMACContent)
	assert.Regexp(t, `^[0-9a-f]{32}$`, hmacFallback, "退化命名应为不带横线的 UUID")

	// 注入密钥后：确定性命名
	SetHMACSecret([]byte("0123456789abcdef0123456789abcdef"))
	h1 := GenerateFileName(content, "", GenerateFileNameTypeHMACContent)
	h2 := GenerateFileName(content, "", GenerateFileNameTypeHMACContent)
	assert.Equal(t, h1, h2, "同密钥同内容的 HMAC 命名必须确定")

	timeName := GenerateFileName(content, ".jpg", GenerateFileNameTypeTimeBase)
	assert.True(t, strings.HasSuffix(timeName, ".jpg"))
	assert.Contains(t, timeName, "_")
}

func TestGenerateObjectName(t *testing.T) {
	name := GenerateObjectName("images/2026/", []byte("x"), ".png", GenerateFileNameTypeUUID)
	require.True(t, strings.HasPrefix(name, "images/2026/"), "应保留目录前缀: %s", name)
	assert.False(t, strings.HasPrefix(name, "//"))
	pure := GenerateObjectName("/", []byte("x"), ".png", GenerateFileNameTypeUUID)
	assert.False(t, strings.Contains(pure, "//"), "目录仅含斜杠时不应产生双斜杠: %s", pure)
}

func TestEnsureObjectName(t *testing.T) {
	name := EnsureObjectName("docs", "report", "application/pdf", nil, GenerateFileNameTypeUUID)
	assert.True(t, strings.HasSuffix(name, ".pdf"), "应从 content-type 推断扩展名: %s", name)
}

func TestJoinObjectNameAndUrl(t *testing.T) {
	object, file := JoinObjectName("image/png", strPtr("imgs"), nil)
	assert.True(t, strings.HasPrefix(object, "imgs/"))
	assert.True(t, strings.HasSuffix(file, ".png"))

	url := JoinObjectUrl("http://minio:9000/", "/images/", "/a/b.png")
	assert.Equal(t, "http://minio:9000/images/a/b.png", url)

	assert.Equal(t, "http://host/a.png", ReplaceEndpointHost("http://minio:9000/a.png", "http://host", "http://minio:9000"))
}

func TestIsFileDirectorySafe(t *testing.T) {
	assert.True(t, IsFileDirectorySafe("images/2026"))
	assert.False(t, IsFileDirectorySafe("../etc"))
	assert.False(t, IsFileDirectorySafe("a/../../b"))
}

func TestIsAllowedMimeType(t *testing.T) {
	assert.True(t, IsAllowedMimeType("image/png"))
	assert.True(t, IsAllowedMimeType("application/pdf"))
	assert.False(t, IsAllowedMimeType("application/x-msdownload"))
	assert.False(t, IsAllowedMimeType("text/html"))
}

func TestEnsureFileExtension(t *testing.T) {
	assert.Equal(t, "jpg", EnsureFileExtension("photo.jpg", "", nil))
	assert.Equal(t, "png", EnsureFileExtension("photo", "image/png", nil))
	fallback := EnsureFileExtension("", "", nil)
	assert.NotEmpty(t, fallback)
	assert.NotContains(t, fallback, ".", "扩展名不应带前导点: %s", fallback)
}

func TestExtractFileExtension(t *testing.T) {
	assert.Equal(t, "png", ExtractFileExtension("a.b.png"))
	assert.Equal(t, "", ExtractFileExtension("noext"))
	assert.Equal(t, "", ExtractFileExtension(".hidden"))
}

func strPtr(s string) *string { return &s }
