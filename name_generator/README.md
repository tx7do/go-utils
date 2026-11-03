# name_generator

English | [简体中文](README.zh-CN.md)

A simple random name generator: it generates Chinese, English and Japanese names, as well as game nicknames, item names and more. 17 dictionary files are built in via Go `embed`, so it works out of the box; custom dictionaries are supported too.

## Features

- Chinese names (single/compound surname, female/male), English names (first/middle/last), Japanese names (kanji and kana editions)
- Game nicknames and item names generated from dictionary combinations (presets Scheme1-Scheme6 and more)
- Extensible dictionaries: `LoadDict` / `LoadDictFromFile` for custom word lists
- Built-in sensitive-words dictionary (`sensitive.txt`) for building your own filtering

## Installation

```bash
go get github.com/tx7do/go-utils/name_generator@latest
```

```go
import "github.com/tx7do/go-utils/name_generator"
```

## Quick Start

### Chinese names

```go
package main

import (
	"log"

	"github.com/tx7do/go-utils/name_generator"
)

func main() {
	g := name_generator.New()

	result := g.GenerateChineseName(1, true, false)
	if result == "" {
		log.Fatalf("result is empty, please check the dictionary data")
	} else {
		log.Printf("Generated single surname single name (female): %s", result)
	}

	result = g.GenerateChineseName(2, false, true)
	if result == "" {
		log.Fatalf("result is empty, please check the dictionary data")
	} else {
		log.Printf("Generated compound surname double name (male): %s", result)
	}
}
```

Sample output (random):

```shell
Generated single surname single name (female): 候影
Generated compound surname double name (male): 宗政辰宁
```

### English names

```go
package main

import (
	"log"

	"github.com/tx7do/go-utils/name_generator"
)

func main() {
	g := name_generator.New()

	result := g.GenerateEnglishName(1, 0, 1, true)
	if result == "" {
		log.Fatalf("result is empty, please check the dictionary data")
	} else {
		log.Printf("Generated female English name: %s", result)
	}

	result = g.GenerateEnglishName(2, 0, 1, false)
	if result == "" {
		log.Fatalf("result is empty, please check the dictionary data")
	} else {
		log.Printf("Generated male English name: %s", result)
	}
}
```

Sample output (random):

```shell
Generated female English name: Magical Alexander
Generated male English name: Valentine Roderick Hayes
```

### Japanese names

Kanji edition:

```go
package main

import (
	"log"

	"github.com/tx7do/go-utils/name_generator"
)

func main() {
	g := name_generator.New()

	result := g.GenerateJapaneseNameCN()
	if result == "" {
		log.Fatalf("result is empty, please check the dictionary data")
	} else {
		log.Printf("Generated Japanese name (CN): %s", result)
	}
}
```

Sample output (random):

```shell
Generated Japanese name (CN): 瀬尾和子
```

Kana edition:

```go
package main

import (
	"log"

	"github.com/tx7do/go-utils/name_generator"
)

func main() {
	g := name_generator.New()

	result := g.GenerateJapaneseName()
	if result == "" {
		log.Fatalf("result is empty, please check the dictionary data")
	} else {
		log.Printf("Generated Japanese name: %s", result)
	}
}
```

Sample output (random):

```shell
Generated Japanese name: 渋沢洋
```

### Game nicknames

Randomly assembled from a preset scheme (dictionary combination):

```go
package main

import (
	"log"

	"github.com/tx7do/go-utils/name_generator"
)

func main() {
	g := name_generator.New()

	dictTypes := name_generator.Scheme5

	result := g.Generate(dictTypes)
	if result == "" {
		log.Fatalf("result is empty, please check the dictionary data")
	} else {
		log.Printf("generate`s nickname: %s", result)
	}
}
```

Sample output (random):

```shell
generate`s nickname: 谦逊之讲笑话呼保义
```

## API

### Generator

```go
func New() *Generator
```

Creates a generator and loads the built-in dictionaries (English and Japanese dictionaries are lazy-loaded on demand — see the dictionary table below).

### Dictionary management

```go
func (g *Generator) LoadDict(dictType DictionaryType, textData []byte) error
func (g *Generator) LoadDictFromFile(dictType DictionaryType, filePath string) error
func (g *Generator) ExistDict(dictType DictionaryType) bool
func (g *Generator) DictCount() int
func (g *Generator) DictItemCount(dictType DictionaryType) int
```

- `LoadDict` / `LoadDictFromFile`: entry points for custom dictionaries. Text is read line by line, skipping empty lines; loading the same `dictType` twice returns an error;
- `ExistDict`: whether a dictionary has been loaded;
- `DictCount`: number of loaded dictionaries;
- `DictItemCount`: number of entries in one dictionary.

### Generation methods

```go
func (g *Generator) Generate(dictTypes CombinedDictionaryType) string
func (g *Generator) GenerateParts(dictTypes CombinedDictionaryType) []string
func (g *Generator) GenerateChineseName(firstNameCount int, isFemale, isCompoundSurname bool) string
func (g *Generator) GenerateEnglishName(firstNameCount, middleNameCount, lastNameCount int, isFemale bool) string
func (g *Generator) GenerateJapaneseNameCN() string
func (g *Generator) GenerateJapaneseName() string
```

- `Generate`: picks one random word per dictionary in the combination and joins them with no separator; `GenerateParts` returns the parts themselves so you can assemble them your own way;
- `GenerateChineseName`: `firstNameCount` must be 1 or 2 (anything else returns an empty string); `isCompoundSurname` selects the compound-surname or single-surname dictionary;
- `GenerateEnglishName`: joined with spaces; `firstNameCount` must be 1-2 and `lastNameCount` >= 1, otherwise an empty string is returned;
- `GenerateJapaneseNameCN`: picks a whole name from the kanji Japanese name corpus;
- `GenerateJapaneseName`: joins a Japanese surname and given name.

### Types

```go
type DictionaryType string       // dictionary type identifier
type Dictionary []string         // a single dictionary (slice of words)
type DictionaryMap map[DictionaryType]Dictionary
type CombinedDictionaryType []DictionaryType // a dictionary combination
```

### Dictionary type constants (DictionaryType)

| Constant | Value | Description |
| -------- | -------- | -------- |
| `DictionaryTypeAdjective` | `adjective` | Adjectives |
| `DictionaryTypeGoods` | `goods` | Item/goods names |
| `DictionaryTypeName` | `name` | Names |
| `DictionaryTypePrefix` | `prefix` | Prefixes |
| `DictionaryTypeRole` | `role` | Roles |
| `DictionaryTypeVerb` | `verb` | Verbs |
| `DictionaryTypeSensitive` | `sensitive` | Sensitive words |
| `DictionaryTypeSingleSurnames` | `single_surnames` | Chinese single-character surnames |
| `DictionaryTypeCompoundSurnames` | `compound_surnames` | Chinese compound surnames |
| `DictionaryTypeChineseFirstNameFemale` | `chinese_first_name_female` | Chinese female given names |
| `DictionaryTypeChineseFirstNameMale` | `chinese_first_name_male` | Chinese male given names |
| `DictionaryTypeEnglishFirstNameFemale` | `english_first_name_female` | English female first names |
| `DictionaryTypeEnglishFirstNameMale` | `english_first_name_male` | English male first names |
| `DictionaryTypeEnglishLastName` | `english_last_name` | English last names |
| `DictionaryTypeJapaneseName` | `japanese_name` | Japanese names (kanji corpus) |
| `DictionaryTypeJapaneseSurnames` | `japanese_surnames` | Japanese surnames |
| `DictionaryTypeJapaneseLastName` | `japanese_last_name` | Japanese given names |

### Preset schemes (Scheme)

| Variable | Dictionary combination |
| -------- | -------- |
| `Scheme1` | prefix + name + verb |
| `Scheme2` | prefix + role + verb |
| `Scheme3` | prefix + name + adjective |
| `Scheme4` | prefix + verb + role |
| `Scheme5` | prefix + verb + name |
| `Scheme6` | name + prefix + goods |
| `SchemeChineseNameFemale` | single_surnames + chinese_first_name_female |
| `SchemeChineseNameMale` | single_surnames + chinese_first_name_male |

### Built-in dictionary list (assets/)

17 dictionary files in total:

| File | Dictionary type | Loaded by `New()` |
| -------- | -------- | :--: |
| `adjective.txt` | `DictionaryTypeAdjective` | ✓ |
| `goods.txt` | `DictionaryTypeGoods` | ✓ |
| `name.txt` | `DictionaryTypeName` | ✓ |
| `prefix.txt` | `DictionaryTypePrefix` | ✓ |
| `role.txt` | `DictionaryTypeRole` | ✓ |
| `verb.txt` | `DictionaryTypeVerb` | ✓ |
| `sensitive.txt` | `DictionaryTypeSensitive` | ✓ |
| `chinese_single_surnames.txt` | `DictionaryTypeSingleSurnames` | ✓ |
| `chinese_compound_surnames.txt` | `DictionaryTypeCompoundSurnames` | ✓ |
| `chinese_first_name_female.txt` | `DictionaryTypeChineseFirstNameFemale` | ✓ |
| `chinese_first_name_male.txt` | `DictionaryTypeChineseFirstNameMale` | ✓ |
| `english_first_name_female.txt` | `DictionaryTypeEnglishFirstNameFemale` | lazy |
| `english_first_name_male.txt` | `DictionaryTypeEnglishFirstNameMale` | lazy |
| `english_last_name.txt` | `DictionaryTypeEnglishLastName` | lazy |
| `japanese_names_corpus.txt` | `DictionaryTypeJapaneseName` | lazy |
| `japanese_surnames.txt` | `DictionaryTypeJapaneseSurnames` | lazy |
| `japanese_last_name.txt` | `DictionaryTypeJapaneseLastName` | lazy |

## Notes

- **Lazy loading**: English and Japanese dictionaries are not loaded by `New()`; they are loaded automatically when the corresponding methods (`GenerateEnglishName` / `GenerateJapaneseName*`) are called. Chinese name dictionaries are also loaded on demand as a fallback during generation.
- **Results can be empty**: when a dictionary is missing or empty, `Generate` skips empty words and may return an empty string — always check for emptiness.
- **Parameter validation**: `GenerateChineseName` returns an empty string when `firstNameCount` is outside 1-2; `GenerateEnglishName` returns an empty string when `firstNameCount` is outside 1-2 or `lastNameCount` < 1.
- **Sensitive words**: `sensitive.txt` is bundled as data only (check its size with `DictItemCount(DictionaryTypeSensitive)`); the package provides no sensitive-word checking function, so implement filtering in your own code.
- **Randomness**: uses the global `math/rand/v2` source; results are not reproducible.

## Acknowledgements

- [Chinese-Names-Corpus](https://github.com/wainshine/Chinese-Names-Corpus)
