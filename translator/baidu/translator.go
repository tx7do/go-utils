package baidu

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	baiduEndpoint = "https://fanyi-api.baidu.com/api/trans/vip/translate"
)

type Translator struct {
	appID     string
	secretKey string
	client    *http.Client
}

func NewTranslator(appID, secretKey string) *Translator {
	return &Translator{
		appID:     appID,
		secretKey: secretKey,
		client:    &http.Client{Timeout: 30 * time.Second},
	}
}

type baiduResponse struct {
	TransResult []struct {
		Src string `json:"src"`
		Dst string `json:"dst"`
	} `json:"trans_result"`
	ErrorCode string `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error) {
	salt := rand.Int63()
	sign := t.generateSign(source, salt)

	// 构建请求参数
	data := url.Values{}
	data.Set("q", source)
	data.Set("from", sourceLang)
	data.Set("to", targetLang)
	data.Set("appid", t.appID)
	data.Set("salt", strconv.FormatInt(salt, 10))
	data.Set("sign", sign)

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", baiduEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	resp, err := t.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("baidu translate error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var baiduResp baiduResponse
	if err := json.Unmarshal(respBody, &baiduResp); err != nil {
		return "", err
	}

	// 检查错误码
	if baiduResp.ErrorCode != "" {
		return "", fmt.Errorf("baidu translate error: %s - %s", baiduResp.ErrorCode, baiduResp.ErrorMsg)
	}

	if len(baiduResp.TransResult) > 0 {
		return baiduResp.TransResult[0].Dst, nil
	}

	return "", fmt.Errorf("baidu translate: invalid response format")
}

func (t *Translator) generateSign(query string, salt int64) string {
	str := fmt.Sprintf("%s%s%d%s", t.appID, query, salt, t.secretKey)
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}
