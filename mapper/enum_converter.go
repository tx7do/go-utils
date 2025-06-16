package mapper

import (
	"github.com/jinzhu/copier"
)

type EnumTypeConverter[DTO ~int32, MODEL ~string] struct {
	nameMap  map[int32]string
	valueMap map[string]int32
}

func NewEnumTypeConverter[DTO ~int32, MODEL ~string](nameMap map[int32]string, valueMap map[string]int32) *EnumTypeConverter[DTO, MODEL] {
	return &EnumTypeConverter[DTO, MODEL]{
		valueMap: valueMap,
		nameMap:  nameMap,
	}
}

func (m *EnumTypeConverter[DTO, MODEL]) ToModel(dto *DTO) *MODEL {
	if dto == nil {
		return nil
	}

	find, ok := m.nameMap[int32(*dto)]
	if !ok {
		return nil
	}

	model := MODEL(find)
	return &model
}

func (m *EnumTypeConverter[DTO, MODEL]) ToDto(model *MODEL) *DTO {
	if model == nil {
		return nil
	}

	find, ok := m.valueMap[string(*model)]
	if !ok {
		return nil
	}

	dto := DTO(find)
	return &dto
}

func (m *EnumTypeConverter[DTO, MODEL]) NewConverterPair() []copier.TypeConverter {
	srcType := MODEL("")
	dstType := DTO(0)

	fromFn := m.ToDto
	toFn := m.ToModel

	return NewGenericTypeConverterPair(&srcType, &dstType, fromFn, toFn)
}

func NewGenericTypeConverterPair[A interface{}, B interface{}](srcType A, dstType B, fromFn func(src A) B, toFn func(src B) A) []copier.TypeConverter {
	return []copier.TypeConverter{
		{
			SrcType: srcType,
			DstType: dstType,
			Fn: func(src interface{}) (interface{}, error) {
				return fromFn(src.(A)), nil
			},
		},
		{
			SrcType: dstType,
			DstType: srcType,
			Fn: func(src interface{}) (interface{}, error) {
				return toFn(src.(B)), nil
			},
		},
	}
}
