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

	type EntityType struct {
		Name string
		Age  int
	}

	mapper := mapper.NewCopierMapper[DtoType, EntityType]()

	// 测试 ToEntity 方法
	dto := &DtoType{Name: "Alice", Age: 25}
	entity := mapper.ToEntity(dto)

	// 测试 ToEntity 方法，传入 nil
	entityNil := mapper.ToEntity(nil)

	// 测试 ToDTO 方法
	entity = &EntityType{Name: "Bob", Age: 30}
	dtoResult := mapper.ToDTO(entity)

	// 测试 ToDTO 方法，传入 nil
	dtoNil := mapper.ToDTO(nil)
}

```

## ent 与 protobuf 的枚举类型映射器

```go
package main

import "github.com/tx7do/go-utils/mapper"

func main() {
	type DtoType int32
	type EntityType string

	const (
		DtoTypeOne DtoType = 1
		DtoTypeTwo DtoType = 2
	)

	const (
		EntityTypeOne EntityType = "One"
		EntityTypeTwo EntityType = "Two"
	)

	nameMap := map[int32]string{
		1: "One",
		2: "Two",
	}
	valueMap := map[string]int32{
		"One": 1,
		"Two": 2,
	}

	converter := mapper.NewEnumTypeConverter[DtoType, EntityType](nameMap, valueMap)

	// 测试 ToEntity 方法
	dto := DtoTypeOne
	entity := converter.ToEntity(&dto)

	// 测试 ToEntity 方法，传入不存在的值
	dtoInvalid := DtoType(3)
	entityInvalid := converter.ToEntity(&dtoInvalid)

	// 测试 ToDTO 方法
	tmpEntityTwo := EntityTypeTwo
	entity = &tmpEntityTwo
	dtoResult := converter.ToDTO(entity)

	// 测试 ToDTO 方法，传入不存在的值
	tmpEntityThree := EntityType("Three")
	entityInvalid = &tmpEntityThree
	dtoInvalidResult := converter.ToDTO(entityInvalid)
}

```
