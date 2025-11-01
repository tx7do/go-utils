package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
	"github.com/google/uuid"
)

var _ ent.Mixin = (*UuidId)(nil)

// UuidId defines an ID field as a UUID.
type UuidId struct{ mixin.Schema }

func (UuidId) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Comment("主键ID (UUID)").
			Default(uuid.New).
			Immutable(),
	}
}
