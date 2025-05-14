package translator

// Translator 翻译器
type Translator interface {
	// Translate 翻译
	Translate(source, sourceLang, targetLang string) (string, error)
}
