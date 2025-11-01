package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/mixin"
)

type SoftDelete struct {
	mixin.Schema
}

func (SoftDelete) Fields() []ent.Field {
	var fields []ent.Field
	fields = append(fields, DeletedAt{}.Fields()...)
	fields = append(fields, DeletedBy{}.Fields()...)
	return fields
}
