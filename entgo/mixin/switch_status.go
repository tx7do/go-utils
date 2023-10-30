package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type SwitchStatus struct {
	mixin.Schema
}

func (SwitchStatus) Fields() []ent.Field {
	return []ent.Field{
		/**
		在PostgreSQL下，还需要为此创建一个Type，否则无法使用。

		DROP TYPE IF EXISTS switch_status CASCADE;
		CREATE TYPE switch_status AS ENUM (
		    'OFF',
		    'ON'
		    );
		*/
		field.Enum("status").
			Comment("状态").
			Optional().
			Nillable().
			//SchemaType(map[string]string{
			//	dialect.MySQL:    "switch_status",
			//	dialect.Postgres: "switch_status",
			//}).
			Default("ON").
			Values(
				"OFF",
				"ON",
			),
	}
}
