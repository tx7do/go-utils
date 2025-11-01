package mixin

import (
	"entgo.io/contrib/entproto"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

var _ ent.Mixin = (*AutoIncrementId)(nil)

type AutoIncrementId struct{ mixin.Schema }

func (AutoIncrementId) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("id").
			Comment("id").
			StructTag(`json:"id,omitempty"`).
			Annotations(
				entproto.Field(1),
			).
			Positive(),
	}
}
