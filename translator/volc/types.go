package volc

// translateTextRequest 翻译请求
type translateTextRequest struct {
	SourceLanguage string   `json:"SourceLanguage"`
	TargetLanguage string   `json:"TargetLanguage"`
	TextList       []string `json:"TextList"`
	Scene          string   `json:"Scene"`
}

// translateTextResponse 翻译响应
type translateTextResponse struct {
	TranslationList []struct {
		Text         string `json:"Text"`
		DetectedLang string `json:"DetectedSourceLanguage,omitempty"`
	} `json:"TranslationList"`
}

// volcResponse 火山标准响应
type volcResponse struct {
	ResponseMetadata struct {
		RequestID string `json:"RequestId"`
		Action    string `json:"Action"`
		Version   string `json:"Version"`
		Service   string `json:"Service"`
		Region    string `json:"Region"`
		Error     *struct {
			Code    string `json:"Code"`
			Message string `json:"Message"`
		} `json:"Error,omitempty"`
	} `json:"ResponseMetadata"`
	TranslationResponse translateTextResponse `json:"TranslationResponse"`
}
