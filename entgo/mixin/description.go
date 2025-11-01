package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// 确保 Description 实现了 ent.Mixin 接口
var _ ent.Mixin = (*Description)(nil)

type Description struct {
	mixin.Schema
}

func (Description) Fields() []ent.Field {
	return []ent.Field{
		field.String("description").
			Comment("描述").
			Optional().
			Nillable(),
	}
}
