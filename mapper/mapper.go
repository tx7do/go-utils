package mapper

import (
	"github.com/jinzhu/copier"
)

type CopierMapper[DTO any, MODEL any] struct {
	copierOption copier.Option
}

func NewCopierMapper[DTO any, MODEL any]() *CopierMapper[DTO, MODEL] {
	return &CopierMapper[DTO, MODEL]{
		copierOption: copier.Option{
			Converters: []copier.TypeConverter{},
		},
	}
}

func (m *CopierMapper[DTO, MODEL]) AppendConverter(converter copier.TypeConverter) {
	m.copierOption.Converters = append(m.copierOption.Converters, converter)
}

func (m *CopierMapper[DTO, MODEL]) AppendConverters(converters []copier.TypeConverter) {
	m.copierOption.Converters = append(m.copierOption.Converters, converters...)
}

func (m *CopierMapper[DTO, MODEL]) ToModel(dto *DTO) *MODEL {
	if dto == nil {
		return nil
	}

	var model MODEL
	if err := copier.CopyWithOption(&model, dto, m.copierOption); err != nil {
		panic(err) // Handle error appropriately in production code
	}

	return &model
}

func (m *CopierMapper[DTO, MODEL]) ToDto(model *MODEL) *DTO {
	if model == nil {
		return nil
	}

	var dto DTO
	if err := copier.CopyWithOption(&dto, model, m.copierOption); err != nil {
		panic(err) // Handle error appropriately in production code
	}

	return &dto
}
