# mapper

English | [简体中文](README.zh-CN.md)

A generics-based type mapper for mapping fields between DTOs and database entities, built on top of [jinzhu/copier](https://github.com/jinzhu/copier), with support for registering custom type converters (typical use case: protobuf int32 enums ↔ database string enums).

## Features

- `Mapper[DTO, ENTITY]` generic interface abstraction, handy for dependency injection and swapping implementations
- `CopierMapper`: the default copier-based implementation; fields with matching names are copied automatically
- `EnumTypeConverter`: bidirectional int32 ↔ string enum mapping
- Converters can be registered into `CopierMapper` so enum fields are converted automatically during copy

## Installation

```bash
go get github.com/tx7do/go-utils/mapper@latest
```

```go
import "github.com/tx7do/go-utils/mapper"
```

## Quick Start

### CopierMapper: field copying

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
	// name the variable m to avoid shadowing the mapper package
	m := mapper.NewCopierMapper[DtoType, EntityType]()

	// DTO -> Entity
	dto := &DtoType{Name: "Alice", Age: 25}
	entity := m.ToEntity(dto)
	fmt.Printf("%+v\n", entity) // &{Name:Alice Age:25}

	// nil in, nil out
	fmt.Println(m.ToEntity(nil) == nil) // true

	// Entity -> DTO
	entity = &EntityType{Name: "Bob", Age: 30}
	dtoResult := m.ToDTO(entity)
	fmt.Printf("%+v\n", dtoResult) // &{Name:Bob Age:30}

	// nil in, nil out
	fmt.Println(m.ToDTO(nil) == nil) // true
}
```

### EnumTypeConverter: enum mapping

Useful for converting between ent (string enums) and protobuf (int32 enums):

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

	// values missing from the maps return nil
	dtoInvalid := DtoType(3)
	fmt.Println(converter.ToEntity(&dtoInvalid) == nil) // true

	// Entity -> DTO
	entityTwo := EntityTypeTwo
	dtoResult := converter.ToDTO(&entityTwo)
	fmt.Println(*dtoResult) // 2

	// values missing from the maps return nil
	entityInvalid := EntityType("Three")
	fmt.Println(converter.ToDTO(&entityInvalid) == nil) // true
}
```

### Combined usage: converting enum fields automatically during copy

Register the converter pair produced by `EnumTypeConverter.NewConverterPair` into a `CopierMapper`, and same-named enum fields are converted between int32 and string automatically during copy:

```go
package main

import (
	"fmt"

	"github.com/tx7do/go-utils/mapper"
)

// DTO-side enum: int32 (e.g. protobuf enum)
type Status int32

// Entity-side enum: string (e.g. value stored in the database)
type StatusEntity string

const (
	StatusOff Status = 0
	StatusOn  Status = 1
)

const (
	StatusEntityOff StatusEntity = "OFF"
	StatusEntityOn  StatusEntity = "ON"
)

// Note: the fields must be pointer types (*Status / *StatusEntity),
// because NewConverterPair registers pointer-based SrcType/DstType and
// copier only applies a converter when the source/target field types
// (including pointerness) match exactly.
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

	// 1. build the enum converter and its copier converter pair
	converter := mapper.NewEnumTypeConverter[Status, StatusEntity](nameMap, valueMap)

	// 2. register it into CopierMapper: enum fields are converted during copy
	m := mapper.NewCopierMapper[TaskDTO, TaskEntity]()
	m.AppendConverters(converter.NewConverterPair())

	// 3. copy: *Status (int32) is converted to *StatusEntity (string) automatically
	st := StatusOn
	entity := m.ToEntity(&TaskDTO{Title: "write docs", Status: &st})
	fmt.Println(entity.Title, *entity.Status) // write docs ON

	// the reverse direction converts too: *StatusEntity (string) -> *Status (int32)
	dto := m.ToDTO(entity)
	fmt.Println(dto.Title, *dto.Status) // write docs 1
}
```

> You can also register single `copier.TypeConverter`s one by one with `AppendConverter`, or build a bidirectional converter pair for any type pair with `NewGenericTypeConverterPair`.

## API

### Mapper interface

```go
type Mapper[DTO any, ENTITY any] interface {
	// ToEntity converts a DTO to a Database Entity.
	ToEntity(*DTO) *ENTITY
	// ToDTO converts a Database Entity to a DTO.
	ToDTO(*ENTITY) *DTO
}
```

The core abstraction: business code can depend on this interface alone, regardless of the implementation.

### CopierMapper

```go
func NewCopierMapper[DTO any, ENTITY any]() *CopierMapper[DTO, ENTITY]
func (m *CopierMapper[DTO, ENTITY]) ToEntity(dto *DTO) *ENTITY
func (m *CopierMapper[DTO, ENTITY]) ToDTO(entity *ENTITY) *DTO
func (m *CopierMapper[DTO, ENTITY]) AppendConverter(converter copier.TypeConverter)
func (m *CopierMapper[DTO, ENTITY]) AppendConverters(converters []copier.TypeConverter)
```

- The copier-based implementation of the `Mapper` interface: copies by matching field names, applying registered converters during the copy;
- `AppendConverter` / `AppendConverters`: append custom type converters to the copier options.

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

- `nameMap`: int32 value → string name (DTO→ENTITY direction);
- `valueMap`: string name → int32 value (ENTITY→DTO direction);
- `NewConverterPair`: produces a bidirectional converter pair ready to register with copier.

### NewGenericTypeConverterPair

```go
func NewGenericTypeConverterPair[A interface{}, B interface{}](
	srcType A,
	dstType B,
	fromFn func(src A) B,
	toFn func(src B) A,
) []copier.TypeConverter
```

Builds a bidirectional copier converter pair for an arbitrary type pair `A ↔ B`. `EnumTypeConverter.NewConverterPair` is implemented on top of it (passing the `*ENTITY` / `*DTO` pointer types).

## Notes

- **nil safety**: both `ToEntity(nil)` and `ToDTO(nil)` return `nil` without panicking.
- **Unknown enums return nil**: `EnumTypeConverter`'s `ToEntity` / `ToDTO` return `nil` (not an error) when the value is missing from the maps.
- **copier errors panic**: `CopierMapper`'s `ToEntity` / `ToDTO` call `panic(err)` when the copier copy fails, so make sure your types are copyable.
- **Only pointer-typed fields trigger automatic conversion**: `NewConverterPair` registers pointer-based `*DTO` / `*ENTITY` src/dst types, and copier matches field types exactly — so enum fields in structs must be declared as pointers (see the combined usage example above).

## Acknowledgements

- [jinzhu/copier](https://github.com/jinzhu/copier) — the underlying copy implementation
