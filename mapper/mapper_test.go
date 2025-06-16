package mapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopierMapper(t *testing.T) {
	type DtoType struct {
		Name string
		Age  int
	}

	type EntityType struct {
		Name string
		Age  int
	}

	mapper := NewCopierMapper[DtoType, EntityType]()

	// 测试 ToEntity 方法
	dto := &DtoType{Name: "Alice", Age: 25}
	entity := mapper.ToEntity(dto)
	assert.NotNil(t, entity)
	assert.Equal(t, "Alice", entity.Name)
	assert.Equal(t, 25, entity.Age)

	// 测试 ToEntity 方法，传入 nil
	entityNil := mapper.ToEntity(nil)
	assert.Nil(t, entityNil)

	// 测试 ToDTO 方法
	entity = &EntityType{Name: "Bob", Age: 30}
	dtoResult := mapper.ToDTO(entity)
	assert.NotNil(t, dtoResult)
	assert.Equal(t, "Bob", dtoResult.Name)
	assert.Equal(t, 30, dtoResult.Age)

	// 测试 ToDTO 方法，传入 nil
	dtoNil := mapper.ToDTO(nil)
	assert.Nil(t, dtoNil)
}

func TestEnumTypeConverter(t *testing.T) {
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

	converter := NewEnumTypeConverter[DtoType, EntityType](nameMap, valueMap)

	// 测试 ToEntity 方法
	dto := DtoTypeOne
	entity := converter.ToEntity(&dto)
	assert.NotNil(t, entity)
	assert.Equal(t, "One", string(*entity))

	// 测试 ToEntity 方法，传入不存在的值
	dtoInvalid := DtoType(3)
	entityInvalid := converter.ToEntity(&dtoInvalid)
	assert.Nil(t, entityInvalid)

	// 测试 ToDTO 方法
	tmpEntityTwo := EntityTypeTwo
	entity = &tmpEntityTwo
	dtoResult := converter.ToDTO(entity)
	assert.NotNil(t, dtoResult)
	assert.Equal(t, DtoType(2), *dtoResult)

	// 测试 ToDTO 方法，传入不存在的值
	tmpEntityThree := EntityType("Three")
	entityInvalid = &tmpEntityThree
	dtoInvalidResult := converter.ToDTO(entityInvalid)
	assert.Nil(t, dtoInvalidResult)
}
