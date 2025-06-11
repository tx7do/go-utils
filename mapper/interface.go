package mapper

// Mapper defines the interface for converting between DTOs and models.
type Mapper[DTO any, MODEL any] interface {
	// ToModel converts a DTO to a MODEL.
	ToModel(*DTO) *MODEL

	// ToDto converts a MODEL to a DTO.
	ToDto(*MODEL) *DTO
}
