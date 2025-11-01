package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// 确保 SortOrder 实现了 ent.Mixin 接口
var _ ent.Mixin = (*SortOrder)(nil)

type SortOrder struct {
	mixin.Schema
}

func (SortOrder) Fields() []ent.Field {
	return []ent.Field{
		field.Int32("sort_order").
			Comment("排序顺序，值越小越靠前").
			Optional().
			Nillable().
			Default(0),
	}
}
