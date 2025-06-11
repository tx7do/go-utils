package name_generator

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/tx7do/go-utils/name_generator/assets"
)

type Generator struct {
	dictionaries DictionaryMap
}

func New() *Generator {
	g := &Generator{
		dictionaries: make(DictionaryMap),
	}

	g.init()

	return g
}

func (g *Generator) init() {
	g.loadAllDict()
}

func (g *Generator) loadAllDict() {
	_ = g.LoadDict(DictionaryTypeAdjective, assets.Adjective)
	_ = g.LoadDict(DictionaryTypeGoods, assets.Goods)
	_ = g.LoadDict(DictionaryTypeName, assets.Name)
	_ = g.LoadDict(DictionaryTypePrefix, assets.Prefix)
	_ = g.LoadDict(DictionaryTypeRole, assets.Role)
	_ = g.LoadDict(DictionaryTypeVerb, assets.Verb)

	_ = g.LoadDict(DictionaryTypeSensitive, assets.Sensitive)

	_ = g.LoadDict(DictionaryTypeSingleSurnames, assets.ChineseSingleSurnames)
	_ = g.LoadDict(DictionaryTypeCompoundSurnames, assets.ChineseCompoundSurnames)
	_ = g.LoadDict(DictionaryTypeChineseFirstNameFemale, assets.ChineseFirstNameFemale)
	_ = g.LoadDict(DictionaryTypeChineseFirstNameMale, assets.ChineseFirstNameMale)

	//_ = g.LoadDict(DictionaryTypeEnglishFirstNameFemale, assets.EnglishFirstNameFemale)
	//_ = g.LoadDict(DictionaryTypeEnglishFirstNameMale, assets.EnglishFirstNameMale)
	//_ = g.LoadDict(DictionaryTypeEnglishLastName, assets.EnglishLastName)
}

func (g *Generator) LoadDict(dictType DictionaryType, textData []byte) error {
	if g.dictionaries == nil {
		g.dictionaries = make(DictionaryMap)
	}

	if _, ok := g.dictionaries[dictType]; ok {
		return errors.New("dictionary already exists for type: " + string(dictType))
	}

	var dict Dictionary

	reader := bytes.NewReader(textData)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word == "" {
			continue // Skip empty lines
		}
		dict = append(dict, word)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	g.dictionaries[dictType] = dict

	return nil
}

func (g *Generator) LoadDictFromFile(dictType DictionaryType, filePath string) error {
	if g.dictionaries == nil {
		g.dictionaries = make(DictionaryMap)
	}

	if _, ok := g.dictionaries[dictType]; ok {
		return errors.New("dictionary already exists for type: " + string(dictType))
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			fmt.Println("Error closing file:", err)
		}
	}(file)

	var dict Dictionary

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word == "" {
			continue // Skip empty lines
		}
		dict = append(dict, strings.TrimSpace(word))
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	g.dictionaries[dictType] = dict

	return nil
}

func (g *Generator) ExistDict(dictType DictionaryType) bool {
	if g.dictionaries == nil {
		return false
	}

	_, exists := g.dictionaries[dictType]
	return exists
}

func (g *Generator) DictCount() int {
	if g.dictionaries == nil {
		return 0
	}

	return len(g.dictionaries)
}

func (g *Generator) DictItemCount(dictType DictionaryType) int {
	if g.dictionaries == nil {
		return 0
	}

	dict, exists := g.dictionaries[dictType]
	if !exists {
		return 0
	}

	return len(dict)
}

func (g *Generator) randomWordFromDict(dictType DictionaryType) string {
	dict, exists := g.dictionaries[dictType]
	if !exists {
		return ""
	}

	if len(dict) == 0 {
		return ""
	}

	randomIndex := rand.Intn(len(dict))
	return dict[randomIndex]
}

func (g *Generator) Generate(dictTypes CombinedDictionaryType) string {
	if len(dictTypes) == 0 {
		return ""
	}

	parts := g.GenerateParts(dictTypes)
	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "")
}

func (g *Generator) GenerateParts(dictTypes CombinedDictionaryType) []string {
	if len(dictTypes) == 0 {
		return nil
	}

	var parts []string
	for _, dictType := range dictTypes {
		word := g.randomWordFromDict(dictType)
		if word != "" {
			parts = append(parts, word)
		}
	}

	return parts
}

func (g *Generator) GenerateChineseName(firstNameCount int, isFemale, isCompoundSurname bool) string {
	if !g.ExistDict(DictionaryTypeSingleSurnames) {
		_ = g.LoadDict(DictionaryTypeSingleSurnames, assets.ChineseSingleSurnames)
	}
	if !g.ExistDict(DictionaryTypeCompoundSurnames) {
		_ = g.LoadDict(DictionaryTypeCompoundSurnames, assets.ChineseCompoundSurnames)
	}
	if !g.ExistDict(DictionaryTypeChineseFirstNameFemale) {
		_ = g.LoadDict(DictionaryTypeChineseFirstNameFemale, assets.ChineseFirstNameFemale)
	}
	if !g.ExistDict(DictionaryTypeChineseFirstNameMale) {
		_ = g.LoadDict(DictionaryTypeChineseFirstNameMale, assets.ChineseFirstNameMale)
	}

	if firstNameCount < 1 || firstNameCount > 2 {
		return ""
	}

	dictTypes := make(CombinedDictionaryType, 0)

	if isCompoundSurname {
		dictTypes = append(dictTypes, DictionaryTypeCompoundSurnames)
	} else {
		dictTypes = append(dictTypes, DictionaryTypeSingleSurnames)
	}

	for i := 0; i < firstNameCount; i++ {
		if isFemale {
			dictTypes = append(dictTypes, DictionaryTypeChineseFirstNameFemale)
		} else {
			dictTypes = append(dictTypes, DictionaryTypeChineseFirstNameMale)
		}
	}

	parts := g.GenerateParts(dictTypes)
	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "")
}

func (g *Generator) GenerateEnglishName(firstNameCount, middleNameCount, lastNameCount int, isFemale bool) string {
	if !g.ExistDict(DictionaryTypeEnglishFirstNameFemale) {
		_ = g.LoadDict(DictionaryTypeEnglishFirstNameFemale, assets.EnglishFirstNameFemale)
	}
	if !g.ExistDict(DictionaryTypeEnglishFirstNameMale) {
		_ = g.LoadDict(DictionaryTypeEnglishFirstNameMale, assets.EnglishFirstNameMale)
	}
	if !g.ExistDict(DictionaryTypeEnglishLastName) {
		_ = g.LoadDict(DictionaryTypeEnglishLastName, assets.EnglishLastName)
	}

	if firstNameCount < 1 || firstNameCount > 2 ||
		lastNameCount < 1 {
		return ""
	}

	dictTypes := make(CombinedDictionaryType, 0)

	for i := 0; i < firstNameCount; i++ {
		if isFemale {
			dictTypes = append(dictTypes, DictionaryTypeEnglishFirstNameFemale)
		} else {
			dictTypes = append(dictTypes, DictionaryTypeEnglishFirstNameMale)
		}
	}

	for i := 0; i < middleNameCount; i++ {
		if isFemale {
			dictTypes = append(dictTypes, DictionaryTypeEnglishFirstNameFemale)
		} else {
			dictTypes = append(dictTypes, DictionaryTypeEnglishFirstNameMale)
		}
	}

	for i := 0; i < lastNameCount; i++ {
		dictTypes = append(dictTypes, DictionaryTypeEnglishLastName)
	}

	parts := g.GenerateParts(dictTypes)
	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, " ")
}

func (g *Generator) GenerateJapaneseNameCN() string {
	if !g.ExistDict(DictionaryTypeJapaneseName) {
		_ = g.LoadDict(DictionaryTypeJapaneseName, assets.JapaneseNamesCorpus)
	}

	dictTypes := CombinedDictionaryType{
		DictionaryTypeJapaneseName,
	}

	parts := g.GenerateParts(dictTypes)
	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "")
}

func (g *Generator) GenerateJapaneseName() string {
	if !g.ExistDict(DictionaryTypeJapaneseSurnames) {
		_ = g.LoadDict(DictionaryTypeJapaneseSurnames, assets.JapaneseSurnames)
	}
	if !g.ExistDict(DictionaryTypeJapaneseLastName) {
		_ = g.LoadDict(DictionaryTypeJapaneseLastName, assets.JapaneseLastName)
	}

	dictTypes := CombinedDictionaryType{
		DictionaryTypeJapaneseSurnames,
		DictionaryTypeJapaneseLastName,
	}

	parts := g.GenerateParts(dictTypes)
	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, "")
}
