package google

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"
)

func TestTranslateV1(t *testing.T) {
	translator := NewTranslator(
		WithVersion("v1"),
	)
	assert.NotNil(t, translator)

	const text string = `Hello, World!`
	// you can use "auto" for source language
	// so, translator will detect language
	result, err := translator.TranslateV1(text, "en", "es")
	assert.Nil(t, err)
	fmt.Println(result)
	// Output: "Hola, Mundo!"

	result, err = translator.TranslateV1(text, "en", "zh-CN")
	assert.Nil(t, err)
	fmt.Println(result)

	result, err = translator.TranslateV1(text, "en", "lo")
	assert.Nil(t, err)
	fmt.Println(result)

	result, err = translator.TranslateV1(text, "en", "my")
	assert.Nil(t, err)
	fmt.Println(result)
}

func TestTranslateV2(t *testing.T) {
	translator := NewTranslator(
		WithVersion("v2"),
		WithApiKey("apikey"),
	)
	assert.NotNil(t, translator)

	const text string = `Hello, World!`

	result, _, err := translator.TranslateV2(text, "en", "es")
	assert.Nil(t, err)
	fmt.Println(result)
	// Output: "Hola, Mundo!"

	result, _, err = translator.TranslateV2(text, "en", "zh-CN")
	assert.Nil(t, err)
	fmt.Println(result)

	result, _, err = translator.TranslateV2(text, "en", "lo")
	assert.Nil(t, err)
	fmt.Println(result)

	result, _, err = translator.TranslateV2(text, "en", "my")
	assert.Nil(t, err)
	fmt.Println(result)
}

func TestTranslateV3(t *testing.T) {
	translator := NewTranslator(
		WithVersion("v3"),
		WithApiKey("apikey"),
	)
	assert.NotNil(t, translator)

	const text string = `Hello, World!`
	// you can use "auto" for source language
	// so, translator will detect language
	result, err := translator.TranslateV3(text, "en", "es")
	assert.Nil(t, err)
	fmt.Println(result)
	// Output: "Hola, Mundo!"

	result, err = translator.TranslateV3(text, "en", "zh-CN")
	assert.Nil(t, err)
	fmt.Println(result)

	result, err = translator.TranslateV3(text, "en", "lo")
	assert.Nil(t, err)
	fmt.Println(result)

	result, err = translator.TranslateV3(text, "en", "my")
	assert.Nil(t, err)
	fmt.Println(result)
}

func TestLanguageParse(t *testing.T) {
	lang, _ := language.Parse("en")
	assert.Equal(t, lang, language.English)
	fmt.Println(lang)
	fmt.Println(language.English.String())
}
