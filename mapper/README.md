# 类型映射器

类型映射器的作用是将一个数据结构的字段映射到另一个数据结构的字段，通常用于对象之间的数据转换。它可以简化不同类型或结构之间的数据传递，减少手动赋值的代码量，提高开发效率。例如，在处理数据库实体与业务模型或 API 请求/响应模型之间的转换时，类型映射器非常有用。

## 基于Copier的类型映射器

```go
package main

import (
	"github.com/tx7do/go-utils/mapper"
)

func main() {
	type DtoType struct {
		Name string
		Age  int
	}

	type ModelType struct {
		Name string
		Age  int
	}

	mapper := mapper.NewCopierMapper[DtoType, ModelType]()

	// 测试 ToModel 方法
	dto := &DtoType{Name: "Alice", Age: 25}
	model := mapper.ToModel(dto)

	// 测试 ToModel 方法，传入 nil
	modelNil := mapper.ToModel(nil)

	// 测试 ToDto 方法
	model = &ModelType{Name: "Bob", Age: 30}
	dtoResult := mapper.ToDto(model)

	// 测试 ToDto 方法，传入 nil
	dtoNil := mapper.ToDto(nil)
}

```

## ent 与 protobuf 的枚举类型映射器

```go
package main

import "github.com/tx7do/go-utils/mapper"

func main() {
	type DtoType string
	type ModelType int32

	const (
		DtoTypeOne DtoType = "One"
		DtoTypeTwo DtoType = "Two"
	)

	const (
		ModelTypeOne ModelType = 1
		ModelTypeTwo ModelType = 2
	)

	nameMap := map[int32]string{
		1: "One",
		2: "Two",
	}
	valueMap := map[string]int32{
		"One": 1,
		"Two": 2,
	}

	converter := mapper.NewEnumTypeConverter[DtoType, ModelType](nameMap, valueMap)

	// 测试 ToModel 方法
	dto := DtoTypeOne
	model := converter.ToModel(&dto)

	// 测试 ToModel 方法，传入不存在的值
	dtoInvalid := DtoType("Three")
	modelInvalid := converter.ToModel(&dtoInvalid)

	// 测试 ToDto 方法
	tmpModelTwo := ModelTypeTwo
	model = &tmpModelTwo
	dtoResult := converter.ToDto(model)

	// 测试 ToDto 方法，传入不存在的值
	tmpModelThree := ModelType(3)
	modelInvalid = &tmpModelThree
	dtoInvalidResult := converter.ToDto(modelInvalid)
}

```
