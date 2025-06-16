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

	type ModelType struct {
		Name string
		Age  int
	}

	mapper := NewCopierMapper[DtoType, ModelType]()

	// 测试 ToModel 方法
	dto := &DtoType{Name: "Alice", Age: 25}
	model := mapper.ToModel(dto)
	assert.NotNil(t, model)
	assert.Equal(t, "Alice", model.Name)
	assert.Equal(t, 25, model.Age)

	// 测试 ToModel 方法，传入 nil
	modelNil := mapper.ToModel(nil)
	assert.Nil(t, modelNil)

	// 测试 ToDto 方法
	model = &ModelType{Name: "Bob", Age: 30}
	dtoResult := mapper.ToDto(model)
	assert.NotNil(t, dtoResult)
	assert.Equal(t, "Bob", dtoResult.Name)
	assert.Equal(t, 30, dtoResult.Age)

	// 测试 ToDto 方法，传入 nil
	dtoNil := mapper.ToDto(nil)
	assert.Nil(t, dtoNil)
}

func TestEnumTypeConverter(t *testing.T) {
	type DtoType int32
	type ModelType string

	const (
		DtoTypeOne DtoType = 1
		DtoTypeTwo DtoType = 2
	)

	const (
		ModelTypeOne ModelType = "One"
		ModelTypeTwo ModelType = "Two"
	)

	nameMap := map[int32]string{
		1: "One",
		2: "Two",
	}
	valueMap := map[string]int32{
		"One": 1,
		"Two": 2,
	}

	converter := NewEnumTypeConverter[DtoType, ModelType](nameMap, valueMap)

	// 测试 ToModel 方法
	dto := DtoTypeOne
	model := converter.ToModel(&dto)
	assert.NotNil(t, model)
	assert.Equal(t, "One", string(*model))

	// 测试 ToModel 方法，传入不存在的值
	dtoInvalid := DtoType(3)
	modelInvalid := converter.ToModel(&dtoInvalid)
	assert.Nil(t, modelInvalid)

	// 测试 ToDto 方法
	tmpModelTwo := ModelTypeTwo
	model = &tmpModelTwo
	dtoResult := converter.ToDto(model)
	assert.NotNil(t, dtoResult)
	assert.Equal(t, DtoType(2), *dtoResult)

	// 测试 ToDto 方法，传入不存在的值
	tmpModelThree := ModelType("Three")
	modelInvalid = &tmpModelThree
	dtoInvalidResult := converter.ToDto(modelInvalid)
	assert.Nil(t, dtoInvalidResult)
}
