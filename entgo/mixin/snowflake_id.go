package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"

	"github.com/tx7do/go-utils/id"
)

type SnowflackId struct {
	mixin.Schema
}

func (SnowflackId) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("id").
			Comment("id").
			DefaultFunc(id.GenerateSonyflakeID).
			Positive().
			Immutable().
			StructTag(`json:"id,omitempty"`).
			SchemaType(map[string]string{
				dialect.MySQL:    "bigint",
				dialect.Postgres: "bigint",
			}),
	}
}

// Indexes of the SnowflackId.
func (SnowflackId) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id"),
	}
}
