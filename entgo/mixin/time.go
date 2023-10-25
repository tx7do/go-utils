package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ ent.Mixin = (*CreatedAt)(nil)

type CreatedAt struct{ mixin.Schema }

func (CreatedAt) Fields() []ent.Field {
	return []ent.Field{
		// 创建时间
		field.Time("created_at").
			Comment("创建时间").
			Immutable().
			Optional().
			Nillable(),
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ ent.Mixin = (*UpdatedAt)(nil)

type UpdatedAt struct{ mixin.Schema }

func (UpdatedAt) Fields() []ent.Field {
	return []ent.Field{
		// 更新时间
		field.Time("updated_at").
			Comment("更新时间").
			Optional().
			Nillable(),
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ ent.Mixin = (*DeletedAt)(nil)

type DeletedAt struct{ mixin.Schema }

func (DeletedAt) Fields() []ent.Field {
	return []ent.Field{
		// 删除时间
		field.Time("deleted_at").
			Comment("删除时间").
			Optional().
			Nillable(),
	}
}

var _ ent.Mixin = (*TimeAt)(nil)

type TimeAt struct{ mixin.Schema }

func (TimeAt) Fields() []ent.Field {
	var fields []ent.Field
	fields = append(fields, CreatedAt{}.Fields()...)
	fields = append(fields, UpdatedAt{}.Fields()...)
	fields = append(fields, DeletedAt{}.Fields()...)
	return fields
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ ent.Mixin = (*CreatedTime)(nil)

type CreatedTime struct{ mixin.Schema }

func (CreatedTime) Fields() []ent.Field {
	return []ent.Field{
		// 创建时间
		field.Time("create_time").
			Comment("创建时间").
			Immutable().
			Optional().
			Nillable(),
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ ent.Mixin = (*UpdatedTime)(nil)

type UpdatedTime struct{ mixin.Schema }

func (UpdatedTime) Fields() []ent.Field {
	return []ent.Field{
		// 更新时间
		field.Time("update_time").
			Comment("更新时间").
			Optional().
			Nillable(),
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ ent.Mixin = (*DeletedTime)(nil)

type DeletedTime struct{ mixin.Schema }

func (DeletedTime) Fields() []ent.Field {
	return []ent.Field{
		// 删除时间
		field.Time("delete_time").
			Comment("删除时间").
			Optional().
			Nillable(),
	}
}

var _ ent.Mixin = (*Time)(nil)

type Time struct{ mixin.Schema }

func (Time) Fields() []ent.Field {
	var fields []ent.Field
	fields = append(fields, CreatedTime{}.Fields()...)
	fields = append(fields, UpdatedTime{}.Fields()...)
	fields = append(fields, DeletedTime{}.Fields()...)
	return fields
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
