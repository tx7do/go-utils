# name_generator

[English](README.md) | 简体中文

一个简单的随机名字生成器：可生成中文名、英文名、日文名，以及游戏昵称、道具名等。内置 17 个词库文件（通过 Go `embed` 打包），开箱即用，也支持加载自定义词库。

## Features

- 中文名（单姓/复姓、男/女名）、英文名（名/中间名/姓）、日文名（汉字版/日文版）
- 按词库组合生成游戏昵称、道具名（Scheme1-Scheme6 等预置组合）
- 词库可扩展：`LoadDict` / `LoadDictFromFile` 加载自定义词典
- 内置敏感词词库（`sensitive.txt`），可用于业务侧自建过滤

## Installation

```bash
go get github.com/tx7do/go-utils/name_generator@latest
```

```go
import "github.com/tx7do/go-utils/name_generator"
```

## Quick Start

### 生成中文名

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

输出效果（随机）：

```shell
Generated single surname single name (female): 候影
Generated compound surname double name (male): 宗政辰宁
```

### 生成英文名

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

输出效果（随机）：

```shell
Generated female English name: Magical Alexander
Generated male English name: Valentine Roderick Hayes
```

### 生成日文名

汉字版：

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

输出效果（随机）：

```shell
Generated Japanese name (CN): 瀬尾和子
```

日文版：

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

输出效果（随机）：

```shell
Generated Japanese name: 渋沢洋
```

### 生成游戏昵称

使用预置词库组合（Scheme）随机拼接：

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

输出效果（随机）：

```shell
generate`s nickname: 谦逊之讲笑话呼保义
```

## API

### Generator

```go
func New() *Generator
```

创建生成器并加载内置词库（英语、日语词库为按需懒加载，见下文词库清单）。

### 词库管理

```go
func (g *Generator) LoadDict(dictType DictionaryType, textData []byte) error
func (g *Generator) LoadDictFromFile(dictType DictionaryType, filePath string) error
func (g *Generator) ExistDict(dictType DictionaryType) bool
func (g *Generator) DictCount() int
func (g *Generator) DictItemCount(dictType DictionaryType) int
```

- `LoadDict` / `LoadDictFromFile`：自定义词典入口。逐行读取文本，跳过空行；同一种 `dictType` 重复加载会返回错误；
- `ExistDict`：判断某词库是否已加载；
- `DictCount`：已加载词库的数量；
- `DictItemCount`：某词库的词条数量。

### 生成方法

```go
func (g *Generator) Generate(dictTypes CombinedDictionaryType) string
func (g *Generator) GenerateParts(dictTypes CombinedDictionaryType) []string
func (g *Generator) GenerateChineseName(firstNameCount int, isFemale, isCompoundSurname bool) string
func (g *Generator) GenerateEnglishName(firstNameCount, middleNameCount, lastNameCount int, isFemale bool) string
func (g *Generator) GenerateJapaneseNameCN() string
func (g *Generator) GenerateJapaneseName() string
```

- `Generate`：按词库组合各随机取一词，无分隔符拼接返回；`GenerateParts` 返回片段本身，便于自行组装；
- `GenerateChineseName`：`firstNameCount` 取 1 或 2（其余返回空串）；`isCompoundSurname` 决定用单姓还是复姓词库；
- `GenerateEnglishName`：以空格分隔；`firstNameCount` 取 1-2、`lastNameCount` ≥ 1，否则返回空串；
- `GenerateJapaneseNameCN`：从日文人名汉字语料库中整名随机；
- `GenerateJapaneseName`：日文姓 + 名拼接。

### 类型

```go
type DictionaryType string       // 词库类型标识
type Dictionary []string         // 单个词库（词条切片）
type DictionaryMap map[DictionaryType]Dictionary
type CombinedDictionaryType []DictionaryType // 词库组合
```

### 词库类型常量（DictionaryType）

| 常量 | 值 | 说明 |
| -------- | -------- | -------- |
| `DictionaryTypeAdjective` | `adjective` | 形容词 |
| `DictionaryTypeGoods` | `goods` | 道具/物品名 |
| `DictionaryTypeName` | `name` | 名称 |
| `DictionaryTypePrefix` | `prefix` | 前缀 |
| `DictionaryTypeRole` | `role` | 角色 |
| `DictionaryTypeVerb` | `verb` | 动词 |
| `DictionaryTypeSensitive` | `sensitive` | 敏感词 |
| `DictionaryTypeSingleSurnames` | `single_surnames` | 中国单姓 |
| `DictionaryTypeCompoundSurnames` | `compound_surnames` | 中国复姓 |
| `DictionaryTypeChineseFirstNameFemale` | `chinese_first_name_female` | 中文女名 |
| `DictionaryTypeChineseFirstNameMale` | `chinese_first_name_male` | 中文男名 |
| `DictionaryTypeEnglishFirstNameFemale` | `english_first_name_female` | 英文女名 |
| `DictionaryTypeEnglishFirstNameMale` | `english_first_name_male` | 英文男名 |
| `DictionaryTypeEnglishLastName` | `english_last_name` | 英文姓 |
| `DictionaryTypeJapaneseName` | `japanese_name` | 日文人名（汉字语料） |
| `DictionaryTypeJapaneseSurnames` | `japanese_surnames` | 日文姓 |
| `DictionaryTypeJapaneseLastName` | `japanese_last_name` | 日文名 |

### 预置词库组合（Scheme）

| 变量 | 词库组合 |
| -------- | -------- |
| `Scheme1` | prefix + name + verb |
| `Scheme2` | prefix + role + verb |
| `Scheme3` | prefix + name + adjective |
| `Scheme4` | prefix + verb + role |
| `Scheme5` | prefix + verb + name |
| `Scheme6` | name + prefix + goods |
| `SchemeChineseNameFemale` | single_surnames + chinese_first_name_female |
| `SchemeChineseNameMale` | single_surnames + chinese_first_name_male |

### 内置词库清单（assets/）

共 17 个词库文件：

| 文件 | 词库类型 | `New()` 时加载 |
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
| `english_first_name_female.txt` | `DictionaryTypeEnglishFirstNameFemale` | 懒加载 |
| `english_first_name_male.txt` | `DictionaryTypeEnglishFirstNameMale` | 懒加载 |
| `english_last_name.txt` | `DictionaryTypeEnglishLastName` | 懒加载 |
| `japanese_names_corpus.txt` | `DictionaryTypeJapaneseName` | 懒加载 |
| `japanese_surnames.txt` | `DictionaryTypeJapaneseSurnames` | 懒加载 |
| `japanese_last_name.txt` | `DictionaryTypeJapaneseLastName` | 懒加载 |

## Notes

- **懒加载**：英语与日语词库不在 `New()` 时加载，而在调用对应生成方法（`GenerateEnglishName` / `GenerateJapaneseName*`）时自动补载；中文姓名词库在生成时也会兜底补载。
- **生成结果可能为空**：词库未加载或组合中的词库为空时，`Generate` 会跳过空词、可能返回空字符串，调用方应判空。
- **参数校验**：`GenerateChineseName` 的 `firstNameCount` 超出 1-2、`GenerateEnglishName` 的 `firstNameCount` 超出 1-2 或 `lastNameCount` < 1 时返回空字符串。
- **敏感词词库**：`sensitive.txt` 仅作为内置数据（可用 `DictItemCount(DictionaryTypeSensitive)` 查看词条数），包内未提供敏感词校验函数，需业务自行实现过滤。
- **随机性**：基于 `math/rand/v2` 全局随机源，结果不可复现。

## Acknowledgements

- [中文人名语料库（Chinese-Names-Corpus）](https://github.com/wainshine/Chinese-Names-Corpus)
