package google

import (
	"context"

	translateV2 "cloud.google.com/go/translate"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

func (t *Translator) TranslateV2(source, _, targetLanguage string) (string, language.Tag, error) {
	ctx := context.Background()

	lang, err := language.Parse(targetLanguage)
	if err != nil {
		return "", language.English, err
	}

	if t.clientV2 == nil {
		client, err := translateV2.NewClient(ctx, option.WithAPIKey(t.options.apiKey))
		if err != nil {
			return "", language.English, err
		}
		t.clientV2 = client
	}

	resp, err := t.clientV2.Translate(ctx, []string{source}, lang, nil)
	if err != nil {
		return "", language.English, err
	}

	return resp[0].Text, resp[0].Source, nil
}
