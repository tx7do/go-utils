# mapper

[English](README.md) | 简体中文

基于泛型的类型映射器，用于在 DTO 与数据库实体（Entity）之间做字段映射，底层使用 [jinzhu/copier](https://github.com/jinzhu/copier) 完成拷贝，并支持注册自定义类型转换器（典型场景：protobuf 的 int32 枚举 ↔ 数据库的 string 枚举）。

## Features

- `Mapper[DTO, ENTITY]` 泛型接口抽象，便于依赖注入与替换实现
- `CopierMapper`：基于 copier 的默认实现，字段名匹配即自动拷贝
- 枚举转换器 `EnumTypeConverter`：int32 ↔ string 枚举双向映射
- 转换器可注册进 `CopierMapper`，拷贝时对同名字段自动应用类型转换

## Installation

```bash
go get github.com/tx7do/go-utils/mapper@latest
```

```go
import "github.com/tx7do/go-utils/mapper"
```

## Quick Start

### CopierMapper：字段拷贝

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/mapper"
)

type DtoType struct {
	Name string
	Age  int
}

type EntityType struct {
	Name string
	Age  int
}

func main() {
	// 变量命名为 m，避免遮蔽 mapper 包名
	m := mapper.NewCopierMapper[DtoType, EntityType]()

	// DTO -> Entity
	dto := &DtoType{Name: "Alice", Age: 25}
	entity := m.ToEntity(dto)
	fmt.Printf("%+v\n", entity) // &{Name:Alice Age:25}

	// 传 nil 返回 nil
	fmt.Println(m.ToEntity(nil) == nil) // true

	// Entity -> DTO
	entity = &EntityType{Name: "Bob", Age: 30}
	dtoResult := m.ToDTO(entity)
	fmt.Printf("%+v\n", dtoResult) // &{Name:Bob Age:30}

	// 传 nil 返回 nil
	fmt.Println(m.ToDTO(nil) == nil) // true
}
```

### EnumTypeConverter：枚举映射

适用于 ent（string 枚举）与 protobuf（int32 枚举）之间的双向转换：

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/mapper"
)

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

func main() {
	nameMap := map[int32]string{1: "One", 2: "Two"}
	valueMap := map[string]int32{"One": 1, "Two": 2}

	converter := mapper.NewEnumTypeConverter[DtoType, EntityType](nameMap, valueMap)

	// DTO -> Entity
	dto := DtoTypeOne
	entity := converter.ToEntity(&dto)
	fmt.Println(*entity) // One

	// 映射表中不存在的值返回 nil
	dtoInvalid := DtoType(3)
	fmt.Println(converter.ToEntity(&dtoInvalid) == nil) // true

	// Entity -> DTO
	entityTwo := EntityTypeTwo
	dtoResult := converter.ToDTO(&entityTwo)
	fmt.Println(*dtoResult) // 2

	// 映射表中不存在的值返回 nil
	entityInvalid := EntityType("Three")
	fmt.Println(converter.ToDTO(&entityInvalid) == nil) // true
}
```

### 组合用法：拷贝时自动转换枚举字段

把 `EnumTypeConverter` 生成的转换器对（`NewConverterPair`）注册进 `CopierMapper`，拷贝时即对同名字段自动做 int32 ↔ string 枚举转换：

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/mapper"
)

// DTO 侧枚举：int32（如 protobuf enum）
type Status int32

// Entity 侧枚举：string（如数据库存储值）
type StatusEntity string

const (
	StatusOff Status = 0
	StatusOn  Status = 1
)

const (
	StatusEntityOff StatusEntity = "OFF"
	StatusEntityOn  StatusEntity = "ON"
)

// 注意：字段必须是枚举的指针类型（*Status / *StatusEntity），
// 因为 NewConverterPair 注册的 SrcType/DstType 是指针类型，
// copier 只在源/目标字段类型（含指针）完全一致时才应用转换器。
type TaskDTO struct {
	Title  string
	Status *Status
}

type TaskEntity struct {
	Title  string
	Status *StatusEntity
}

func main() {
	nameMap := map[int32]string{int32(StatusOff): "OFF", int32(StatusOn): "ON"}
	valueMap := map[string]int32{"OFF": int32(StatusOff), "ON": int32(StatusOn)}

	// 1. 构造枚举转换器，生成 copier 转换器对
	converter := mapper.NewEnumTypeConverter[Status, StatusEntity](nameMap, valueMap)

	// 2. 注册进 CopierMapper：拷贝时自动转换枚举字段
	m := mapper.NewCopierMapper[TaskDTO, TaskEntity]()
	m.AppendConverters(converter.NewConverterPair())

	// 3. 拷贝：*Status(int32) 自动转换为 *StatusEntity(string)
	st := StatusOn
	entity := m.ToEntity(&TaskDTO{Title: "write docs", Status: &st})
	fmt.Println(entity.Title, *entity.Status) // write docs ON

	// 反向同样自动转换：*StatusEntity(string) -> *Status(int32)
	dto := m.ToDTO(entity)
	fmt.Println(dto.Title, *dto.Status) // write docs 1
}
```

> 也可用 `AppendConverter` 逐个注册单个 `copier.TypeConverter`，或用 `NewGenericTypeConverterPair` 为任意一对类型构造双向转换器。

## API

### Mapper 接口

```go
type Mapper[DTO any, ENTITY any] interface {
	// ToEntity converts a DTO to a Database Entity.
	ToEntity(*DTO) *ENTITY
	// ToDTO converts a Database Entity to a DTO.
	ToDTO(*ENTITY) *DTO
}
```

核心抽象：业务代码依赖此接口即可，不关心具体实现。

### CopierMapper

```go
func NewCopierMapper[DTO any, ENTITY any]() *CopierMapper[DTO, ENTITY]
func (m *CopierMapper[DTO, ENTITY]) ToEntity(dto *DTO) *ENTITY
func (m *CopierMapper[DTO, ENTITY]) ToDTO(entity *ENTITY) *DTO
func (m *CopierMapper[DTO, ENTITY]) AppendConverter(converter copier.TypeConverter)
func (m *CopierMapper[DTO, ENTITY]) AppendConverters(converters []copier.TypeConverter)
```

- `Mapper` 接口的 copier 实现：按字段名匹配拷贝，注册的转换器会在拷贝时自动应用；
- `AppendConverter` / `AppendConverters`：向 copier 选项追加自定义类型转换器。

### EnumTypeConverter

```go
func NewEnumTypeConverter[DTO ~int32, ENTITY ~string](
	nameMap map[int32]string,
	valueMap map[string]int32,
) *EnumTypeConverter[DTO, ENTITY]

func (m *EnumTypeConverter[DTO, ENTITY]) ToEntity(dto *DTO) *ENTITY
func (m *EnumTypeConverter[DTO, ENTITY]) ToDTO(entity *ENTITY) *DTO
func (m *EnumTypeConverter[DTO, ENTITY]) NewConverterPair() []copier.TypeConverter
```

- `nameMap`：int32 值 → string 名（DTO→ENTITY 方向）；
- `valueMap`：string 名 → int32 值（ENTITY→DTO 方向）；
- `NewConverterPair`：生成可直接注册给 copier 的双向转换器对。

### NewGenericTypeConverterPair

```go
func NewGenericTypeConverterPair[A interface{}, B interface{}](
	srcType A,
	dstType B,
	fromFn func(src A) B,
	toFn func(src B) A,
) []copier.TypeConverter
```

为任意类型对 `A ↔ B` 构造双向 copier 转换器。`EnumTypeConverter.NewConverterPair` 即基于它实现（传入 `*ENTITY` / `*DTO` 指针类型）。

## Notes

- **nil 安全**：`ToEntity(nil)` 与 `ToDTO(nil)` 均返回 `nil`，不会 panic。
- **枚举查不到返回 nil**：`EnumTypeConverter` 的 `ToEntity` / `ToDTO` 在映射表中查不到对应值时返回 `nil`（不是错误）。
- **copier 出错会 panic**：`CopierMapper` 的 `ToEntity` / `ToDTO` 在 copier 拷贝失败时直接 `panic(err)`，请确保类型定义可拷贝。
- **指针类型字段才能触发自动转换**：`NewConverterPair` 注册的源/目标类型是 `*DTO` / `*ENTITY` 指针类型，copier 按字段类型精确匹配，因此结构体中的枚举字段需声明为指针（见上文组合用法示例）。

## Acknowledgements

- [jinzhu/copier](https://github.com/jinzhu/copier) — 底层拷贝实现
