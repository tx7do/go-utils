# 名字生成器

这是一个简单的名字生成器。它可以生成中文名、英文名和日文名。以及游戏的昵称，道具名等。

## 生成中文名

```go
package main

import "github.com/tx7do/go-utils/name_generator"

func main() {
	g := name_generator.New()

	result := g.GenerateChineseName(1, true, false)
	if result == "" {
		log.Errorf("result is empty, please check the dictionary data")
	} else {
		log.Logf("Generated single surname single name (female): %s", result)
	}

	result = g.GenerateChineseName(2, false, true)
	if result == "" {
		log.Errorf("result is empty, please check the dictionary data")
	} else {
		log.Logf("Generated compound surname double name (male): %s", result)
	}
}

```

输出效果：

```shell
Generated single surname single name (female): 候影
Generated compound surname double name (male): 宗政辰宁
```

## 生成英文名

```go
package main

import "github.com/tx7do/go-utils/name_generator"

func main() {
	g := name_generator.New()

	result := g.GenerateEnglishName(1, 0, 1, true)
	if result == "" {
		log.Errorf("result is empty, please check the dictionary data")
	} else {
		log.Logf("Generated female English name: %s", result)
	}

	result = g.GenerateEnglishName(2, 0, 1, false)
	if result == "" {
		log.Errorf("result is empty, please check the dictionary data")
	} else {
		log.Logf("Generated male English name: %s", result)
	}
}

```

输出效果：

```shell
Generated female English name: Magical Alexander
Generated male English name: Valentine Roderick Hayes
```

## 生成日文名

汉字版

```go
package main

import "github.com/tx7do/go-utils/name_generator"

func main() {
	g := name_generator.New()

	result := g.GenerateJapaneseNameCN()
	if result == "" {
		log.Errorf("result is empty, please check the dictionary data")
	} else {
		log.Logf("Generated Japanese name (CN): %s", result)
	}
}

```

输出效果：

```shell
Generated Japanese name (CN): 瀬尾和子
```

日文版

```go
package main

import "github.com/tx7do/go-utils/name_generator"

func main() {
	g := name_generator.New()

	result := g.GenerateJapaneseName()
	if result == "" {
		log.Errorf("result is empty, please check the dictionary data")
	} else {
		log.Logf("Generated Japanese name: %s", result)
	}
}

```

输出效果：

```shell
Generated Japanese name: 渋沢洋
```

## 生成游戏昵称

```go
package main

import "github.com/tx7do/go-utils/name_generator"

func main() {
	g := name_generator.New()

	dictTypes := name_generator.Scheme5

	result := g.Generate(dictTypes)

	if result == "" {
		log.Errorf("result is empty, please check the dictionary data")
	} else {
		log.Logf("generate`s nickname: %s", result)
	}
}

```

输出效果：

```shell
generate`s nickname: 谦逊之讲笑话呼保义
```

## 感谢

- [中文人名语料库（Chinese-Names-Corpus）](https://github.com/wainshine/Chinese-Names-Corpus)
