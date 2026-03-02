package alibaba

import (
	"fmt"

	alimt "github.com/alibabacloud-go/alimt-20190107/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

type Translator struct {
	client   *alimt.Client
	regionID string
}

// NewTranslator 创建阿里翻译器
// accessKeyID: 阿里云访问密钥 ID
// accessKeySecret: 阿里云访问密钥密码
func NewTranslator(accessKeyID, accessKeySecret string, opts ...Option) (*Translator, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(accessKeyID),
		AccessKeySecret: tea.String(accessKeySecret),
	}

	// 默认使用杭州区域
	config.Endpoint = tea.String("mt.cn-hangzhou.aliyuncs.com")

	translator := &Translator{
		regionID: "cn-hangzhou",
	}

	// 应用选项
	for _, opt := range opts {
		opt(translator)
	}

	// 根据区域设置端点
	if translator.regionID != "" {
		config.Endpoint = tea.String(fmt.Sprintf("mt.%s.aliyuncs.com", translator.regionID))
	}

	client, err := alimt.NewClient(config)
	if err != nil {
		return nil, err
	}

	translator.client = client
	return translator, nil
}

// Translate 翻译文本
func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error) {
	request := &alimt.TranslateGeneralRequest{
		SourceLanguage: tea.String(sourceLang),
		TargetLanguage: tea.String(targetLang),
		SourceText:     tea.String(source),
		FormatType:     tea.String("text"),
		Scene:          tea.String("general"),
	}

	response, err := t.client.TranslateGeneral(request)
	if err != nil {
		return "", fmt.Errorf("alibaba translate error: %w", err)
	}

	if response.Body.Code != nil && *response.Body.Code != 200 {
		return "", fmt.Errorf("alibaba translate error: code=%d, message=%s", *response.Body.Code, tea.StringValue(response.Body.Message))
	}

	if response.Body == nil || response.Body.Data == nil {
		return "", fmt.Errorf("alibaba translate: empty response")
	}

	if response.Body.Data.Translated == nil {
		return "", fmt.Errorf("alibaba translate: translation result is nil")
	}

	return tea.StringValue(response.Body.Data.Translated), nil
}
