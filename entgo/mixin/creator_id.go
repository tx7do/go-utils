package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// 确保 CreatorId 实现了 ent.Mixin 接口
var _ ent.Mixin = (*CreatorId)(nil)

type CreatorId struct {
	mixin.Schema
}

func (CreatorId) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("creator_id").
			Comment("创建者用户ID").
			Immutable().
			Optional().
			Nillable(),
	}
}

// Indexes of the CreatorId mixin.
func (CreatorId) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("creator_id"),
	}
}
