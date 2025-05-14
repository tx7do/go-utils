package google

import (
	"context"
	"fmt"

	translateV3 "cloud.google.com/go/translate/apiv3"
	"cloud.google.com/go/translate/apiv3/translatepb"
	"google.golang.org/api/option"
)

func (t *Translator) TranslateV3(source, sourceLang, targetLang string) (string, error) {
	ctx := context.Background()

	if t.clientV3 == nil {
		client, err := translateV3.NewTranslationClient(ctx, option.WithAPIKey(t.options.apiKey))
		if err != nil {
			return "", fmt.Errorf("NewTranslationClient: %w", err)
		}
		t.clientV3 = client
	}

	req := &translatepb.TranslateTextRequest{
		SourceLanguageCode: sourceLang,
		TargetLanguageCode: targetLang,
		MimeType:           "text/plain", // Mime types: "text/plain", "text/html"
		Contents:           []string{source},
	}

	resp, err := t.clientV3.TranslateText(ctx, req)
	if err != nil {
		return "", fmt.Errorf("TranslateText: %w", err)
	}

	if len(resp.GetTranslations()) != 1 {
		return "", fmt.Errorf("TranslateText: %w", err)
	}

	return resp.GetTranslations()[0].GetTranslatedText(), nil
}
