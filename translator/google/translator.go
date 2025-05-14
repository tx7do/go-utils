package google

import (
	translateV2 "cloud.google.com/go/translate"
	translateV3 "cloud.google.com/go/translate/apiv3"
)

type Translator struct {
	options *options

	clientV2 *translateV2.Client
	clientV3 *translateV3.TranslationClient
}

func NewTranslator(opts ...Option) *Translator {
	op := options{}
	for _, o := range opts {
		o(&op)
	}

	return &Translator{
		options: &op,
	}
}

func (t *Translator) Translate(source, sourceLang, targetLang string) (string, error) {
	switch t.options.version {
	default:
		fallthrough
	case "v1":
		return t.TranslateV1(source, sourceLang, targetLang)

	case "v2":
		text, _, err := t.TranslateV2(source, sourceLang, targetLang)
		return text, err

	case "v3":
		return t.TranslateV3(source, sourceLang, targetLang)
	}
}
