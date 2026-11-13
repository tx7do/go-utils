// Package baidu 提供百度通用翻译 API 的 Translator 实现
// （https://fanyi-api.baidu.com/ ，标准版有免费额度，大陆网络可达）。
//
// 语言标签走 BCP47（zh-CN/en-US/ar），由 NormalizeLang 归一为百度语种码
// （zh/en/ara）；未收录的语种去掉地区后缀后小写透传。
package baidu

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const DefaultEndpoint = "https://fanyi-api.baidu.com/api/trans/vip/translate"

// bcp47ToBaidu 将 BCP47 语言标签映射为百度语种码。
var bcp47ToBaidu = map[string]string{
	"auto":  "auto",
	"zh":    "zh",
	"zh-cn": "zh",
	"zh-tw": "cht",
	"zh-hk": "cht",
	"en":    "en",
	"en-us": "en",
	"en-gb": "en",
	"ar":    "ara",
	"ja":    "jp",
	"ja-jp": "jp",
	"ko":    "kor",
	"ko-kr": "kor",
	"fr":    "fra",
	"fr-fr": "fra",
	"es":    "spa",
	"es-es": "spa",
	"de":    "de",
	"de-de": "de",
	"ru":    "ru",
	"ru-ru": "ru",
	"pt":    "pt",
	"pt-br": "pt",
	"it":    "it",
	"it-it": "it",
	"th":    "th",
	"th-th": "th",
	"vi":    "vie",
	"vi-vn": "vie",
}

// NormalizeLang 将 BCP47 语言标签归一为百度语种码。
// 未收录的语种去掉地区后缀后小写透传（xxx-YY → xxx）。
func NormalizeLang(lang string) string {
	l := strings.ToLower(strings.TrimSpace(lang))
	if v, ok := bcp47ToBaidu[l]; ok {
		return v
	}
	if i := strings.IndexByte(l, '-'); i > 0 {
		if v, ok := bcp47ToBaidu[l[:i]]; ok {
			return v
		}
		return l[:i]
	}
	return l
}

type Translator struct {
	appID      string
	secretKey  string
	endpoint   string
	httpClient *http.Client
}

// Option 是百度翻译的配置选项
type Option func(*Translator)

// WithEndpoint 覆盖请求端点（测试注入 mock 服务用）
func WithEndpoint(endpoint string) Option {
	return func(t *Translator) { t.endpoint = endpoint }
}

// WithHTTPClient 覆盖 HTTP 客户端（代理/超时策略自定义用）
func WithHTTPClient(c *http.Client) Option {
	return func(t *Translator) { t.httpClient = c }
}

func NewTranslator(appID, secretKey string, opts ...Option) *Translator {
	t := &Translator{
		appID:      appID,
		secretKey:  secretKey,
		endpoint:   DefaultEndpoint,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, o := range opts {
		o(t)
	}
	return t
}

type baiduResponse struct {
	TransResult []struct {
		Src string `json:"src"`
		Dst string `json:"dst"`
	} `json:"trans_result"`
	ErrorCode string `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

// Translate 翻译文本。sourceLang 传 "auto" 或空串时由百度自动检测源语言。
// 多段结果（换行分隔的输入）按换行合并返回。
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error) {
	if t.appID == "" || t.secretKey == "" {
		return "", errors.New("baidu translator: appid/secret is not configured")
	}
	if strings.TrimSpace(source) == "" {
		return "", nil
	}

	from := NormalizeLang(sourceLang)
	if from == "" {
		from = "auto"
	}
	to := NormalizeLang(targetLang)

	salt := strconv.FormatUint(rand.Uint64(), 10)

	data := url.Values{}
	data.Set("q", source)
	data.Set("from", from)
	data.Set("to", to)
	data.Set("appid", t.appID)
	data.Set("salt", salt)
	data.Set("sign", t.sign(source, salt))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("baidu translator: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("baidu translator: request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return "", fmt.Errorf("baidu translator: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("baidu translator: unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var out baiduResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", fmt.Errorf("baidu translator: decode response: %w", err)
	}
	if out.ErrorCode != "" {
		return "", fmt.Errorf("baidu translator: api error %s: %s", out.ErrorCode, out.ErrorMsg)
	}

	var sb strings.Builder
	for i, r := range out.TransResult {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(r.Dst)
	}
	return sb.String(), nil
}

// sign 生成百度签名：MD5(appid + q + salt + 密钥)
//
// MD5 是百度通用翻译 API 协议强制的签名算法（服务端只认这一种，无法升级为
// SHA-256 等更强哈希）；此处用途是请求鉴权 MAC，并非口令/敏感数据的存储哈希。
func (t *Translator) sign(q, salt string) string {
	sum := md5.Sum([]byte(t.appID + q + salt + t.secretKey)) // codeql[go/weak-sensitive-data-hashing] Justification: MD5 is mandated by the Baidu Translate API signing protocol
	return hex.EncodeToString(sum[:])
}
