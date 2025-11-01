package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// 确保 TenantID 实现了 ent.Mixin 接口
var _ ent.Mixin = (*TenantID)(nil)

type TenantID struct{ mixin.Schema }

func (TenantID) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("tenant_id").
			Comment("租户ID").
			Positive().
			Immutable().
			Nillable().
			Optional(),
	}
}

// Indexes of the TenantID.
func (TenantID) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id"),
	}
}
