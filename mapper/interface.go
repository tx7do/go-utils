package mapper

// Mapper defines the interface for converting between Data Transfer Objects (DTOs) and Database Entities.
type Mapper[DTO any, ENTITY any] interface {
	// ToEntity converts a DTO to a Database Entity.
	ToEntity(*DTO) *ENTITY

	// ToDTO converts a Database Entity to a DTO.
	ToDTO(*ENTITY) *DTO
}
