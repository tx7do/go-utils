package slug

import (
	"github.com/gosimple/slug"
)

// Generate 生成短链接
func Generate(input string) string {
	slug.Lowercase = true
	return slug.MakeLang(input, "en")
}

// GenerateCaseSensitive 生成大小写敏感的短链接
func GenerateCaseSensitive(input string) string {
	slug.Lowercase = false
	return slug.MakeLang(input, "en")
}

// GenerateEnglish 生成英文短链接
func GenerateEnglish(input string) string {
	slug.Lowercase = true
	return slug.MakeLang(input, "en")
}

// GenerateGerman 生成德文短链接
func GenerateGerman(input string) string {
	slug.Lowercase = true
	return slug.MakeLang(input, "de")
}
