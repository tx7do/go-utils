package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// 确保 SwitchStatus 实现了 ent.Mixin 接口
var _ ent.Mixin = (*SwitchStatus)(nil)

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
			NamedValues(
				"Off", "OFF",
				"On", "ON",
			),
	}
}

// Indexes of the SwitchStatus mixin.
func (SwitchStatus) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
	}
}
