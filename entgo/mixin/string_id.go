package mixin

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// 确保 StringId 实现了 ent.Mixin 接口
var _ ent.Mixin = (*StringId)(nil)

type StringId struct {
	mixin.Schema
}

func (StringId) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Comment("id").
			MaxLen(25).
			NotEmpty().
			Match(regexp.MustCompile("^[0-9a-zA-Z_\\-]+$")).
			StructTag(`json:"id,omitempty"`),
	}
}
