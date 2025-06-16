package mapper

import (
	"github.com/jinzhu/copier"
)

type EnumTypeConverter[DTO ~int32, ENTITY ~string] struct {
	nameMap  map[int32]string
	valueMap map[string]int32
}

func NewEnumTypeConverter[DTO ~int32, ENTITY ~string](
	nameMap map[int32]string,
	valueMap map[string]int32,
) *EnumTypeConverter[DTO, ENTITY] {
	return &EnumTypeConverter[DTO, ENTITY]{
		valueMap: valueMap,
		nameMap:  nameMap,
	}
}

func (m *EnumTypeConverter[DTO, ENTITY]) ToEntity(dto *DTO) *ENTITY {
	if dto == nil {
		return nil
	}

	find, ok := m.nameMap[int32(*dto)]
	if !ok {
		return nil
	}

	entity := ENTITY(find)
	return &entity
}

func (m *EnumTypeConverter[DTO, ENTITY]) ToDTO(entity *ENTITY) *DTO {
	if entity == nil {
		return nil
	}

	find, ok := m.valueMap[string(*entity)]
	if !ok {
		return nil
	}

	dto := DTO(find)
	return &dto
}

func (m *EnumTypeConverter[DTO, ENTITY]) NewConverterPair() []copier.TypeConverter {
	srcType := ENTITY("")
	dstType := DTO(0)

	fromFn := m.ToDTO
	toFn := m.ToEntity

	return NewGenericTypeConverterPair(&srcType, &dstType, fromFn, toFn)
}

func NewGenericTypeConverterPair[A interface{}, B interface{}](
	srcType A,
	dstType B,
	fromFn func(src A) B,
	toFn func(src B) A,
) []copier.TypeConverter {
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
